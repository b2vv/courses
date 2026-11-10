package application

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"syscall"
	"time"

	"courses/internal/media/domain"
	"courses/internal/media/ports"
)

const (
	transcodeTimeout = 30 * time.Minute
	sweepLimit       = 50
)

// TranscodeOutcome is the job's business result. An error returned beside it
// is reserved for dependency faults, which are always retryable.
type TranscodeOutcome struct {
	Failed    bool
	Reason    domain.FailureReason
	Permanent bool
}

type Transcode struct {
	repository ports.AssetRepository
	storage    ports.MediaStorage
	transcoder ports.Transcoder
	queue      ports.TranscodingQueue
	now        Clock
}

func NewTranscode(repository ports.AssetRepository, storage ports.MediaStorage, transcoder ports.Transcoder, queue ports.TranscodingQueue, now Clock) *Transcode {
	return &Transcode{repository: repository, storage: storage, transcoder: transcoder, queue: queue, now: now}
}

func (transcode *Transcode) Process(ctx context.Context, job ports.Job) (TranscodeOutcome, error) {
	callCtx, cancel := context.WithTimeout(ctx, transcodeTimeout)
	defer cancel()
	if job.TrackID != "" {
		return transcode.processAudioTrack(callCtx, job)
	}
	now := transcode.timestamp()
	if err := transcode.sweep(callCtx, now); err != nil {
		return TranscodeOutcome{}, mediaOperationError(ctx, err)
	}

	asset, err := transcode.repository.Load(callCtx, job.AssetID)
	if errors.Is(err, domain.ErrVideoNotFound) {
		return TranscodeOutcome{}, nil
	}
	if err != nil {
		return TranscodeOutcome{}, mediaOperationError(ctx, err)
	}
	if disposition := dispositionOf(asset); disposition != processContinue {
		if disposition == processInvalid {
			return TranscodeOutcome{}, domain.ErrDependencyUnavailable
		}
		return TranscodeOutcome{}, nil
	}

	token, err := freshClaimToken()
	if err != nil {
		return TranscodeOutcome{}, err
	}
	claimedAsset, claimed, err := transcode.repository.Claim(callCtx, job.AssetID, token, now)
	if errors.Is(err, domain.ErrVideoNotFound) {
		return TranscodeOutcome{}, nil
	}
	if err != nil {
		return TranscodeOutcome{}, mediaOperationError(ctx, err)
	}
	if !claimed {
		switch dispositionOf(claimedAsset) {
		case processDone:
			return TranscodeOutcome{}, nil
		case processContinue:
			if claimedAsset.State == domain.Processing {
				return TranscodeOutcome{}, fmt.Errorf("media asset is already processing")
			}
		}
		return TranscodeOutcome{}, domain.ErrDependencyUnavailable
	}

	asset = claimedAsset
	free, err := transcode.storage.Free(callCtx)
	if err != nil {
		return transcode.dependencyFailure(callCtx, job, token, err)
	}
	if free < 3*asset.DeclaredBytes {
		return transcode.businessFailure(callCtx, job, token, domain.StorageUnavailable)
	}

	source := sourceKey(asset.ID)
	work := renderWorkKey(asset.ID, token)
	workThumbnail := renderThumbnailKey(asset.ID, token)
	probe, err := transcode.transcoder.Probe(callCtx, source)
	if err != nil {
		return transcode.transcoderFailure(callCtx, job, token, err, domain.ProbeFailed)
	}
	if err := domain.AcceptSource(probe); err != nil {
		return transcode.businessFailure(callCtx, job, token, domain.UnsupportedSource)
	}
	renditions := domain.SelectLadder(probe)
	if err := transcode.storage.Purge(callCtx, work); err != nil {
		return transcode.storageFailure(callCtx, job, token, err, domain.StorageUnavailable)
	}
	if err := transcode.storage.Purge(callCtx, workThumbnail); err != nil {
		return transcode.storageFailure(callCtx, job, token, err, domain.StorageUnavailable)
	}
	if err := transcode.transcoder.Render(callCtx, source, work, renditions, domain.SourceHasAudio(probe)); err != nil {
		return transcode.transcoderFailure(callCtx, job, token, err, domain.RenderFailed)
	}
	if err := transcode.transcoder.Thumbnail(callCtx, source, workThumbnail, probe); err != nil {
		return transcode.transcoderFailure(callCtx, job, token, err, domain.RenderFailed)
	}

	var embedded *domain.AudioTrack
	if domain.SourceHasAudio(probe) {
		track, err := embeddedTrack(asset.ID, transcode.timestamp())
		if err != nil {
			return TranscodeOutcome{}, err
		}
		embedded = &track
	}
	var policyErr error
	published, evicted, err := transcode.repository.Publish(callCtx, asset.ID, token, renditions, probe.UsableDurationSeconds(), transcode.timestamp(), embedded, func(domain.VideoAsset) error {
		if err := transcode.storage.Publish(callCtx, work, hlsKey(asset.ID)); err != nil {
			policyErr = err
			return err
		}
		if err := transcode.storage.Move(callCtx, workThumbnail, thumbnailKey(asset.ID)); err != nil {
			policyErr = err
			return err
		}
		return nil
	})
	if err != nil {
		if policyErr != nil {
			return transcode.storageFailure(callCtx, job, token, policyErr, domain.PublishFailed)
		}
		return transcode.dependencyFailure(callCtx, job, token, err)
	}
	if published.ID == "" {
		if err := transcode.cleanup(callCtx, asset.ID, token); err != nil {
			return TranscodeOutcome{}, err
		}
		return TranscodeOutcome{}, nil
	}
	if err := transcode.cleanup(callCtx, asset.ID, token); err != nil {
		return TranscodeOutcome{}, err
	}
	if evicted != nil {
		if err := transcode.purgeAsset(callCtx, evicted.ID); err != nil {
			return TranscodeOutcome{}, err
		}
		if err := transcode.repository.DeleteSwept(callCtx, evicted.ID); err != nil {
			return TranscodeOutcome{}, err
		}
	}
	return TranscodeOutcome{}, nil
}

// processAudioTrack transcodes one uploaded audio track and publishes it into the
// asset's HLS tree, then promotes the track (and, for a video-only asset, the
// asset itself) and regenerates the master under the per-asset lock (D1/D4/D7).
// The lifecycle is lean (owner decision 2026-11-08): no per-track claim token —
// idempotency rests on the asynq task id and the Ready guard; a transient fault
// is retried by asynq; only an unsupported source fails the track permanently.
func (transcode *Transcode) processAudioTrack(ctx context.Context, job ports.Job) (TranscodeOutcome, error) {
	track, err := transcode.repository.LoadAudioTrack(ctx, job.AssetID, job.TrackID)
	if errors.Is(err, domain.ErrVideoNotFound) {
		return TranscodeOutcome{}, nil
	}
	if err != nil {
		return TranscodeOutcome{}, mediaOperationError(ctx, err)
	}
	// An embedded track never transcodes; a Ready track is already done (idempotent
	// retry); any state other than Queued/Processing is not ready to run.
	if track.Kind != domain.UploadedTrack || track.State == domain.Ready ||
		(track.State != domain.Queued && track.State != domain.Processing) {
		return TranscodeOutcome{}, nil
	}

	token, err := freshClaimToken()
	if err != nil {
		return TranscodeOutcome{}, err
	}
	source := audioSourceKey(job.AssetID, job.TrackID)
	work := renderWorkKey(job.AssetID, token)

	probe, err := transcode.transcoder.Probe(ctx, source)
	if err != nil {
		return transcode.audioTrackTransientFailure(ctx, job, err, domain.ProbeFailed)
	}
	if err := domain.AcceptAudioTrackSource(probe); err != nil {
		if failErr := transcode.repository.FailAudioTrack(ctx, job.AssetID, job.TrackID, transcode.timestamp()); failErr != nil {
			return TranscodeOutcome{}, failErr
		}
		return TranscodeOutcome{Failed: true, Reason: domain.UnsupportedSource, Permanent: true}, nil
	}
	if err := transcode.storage.Purge(ctx, work); err != nil {
		return transcode.audioTrackTransientFailure(ctx, job, err, domain.StorageUnavailable)
	}
	if err := transcode.transcoder.RenderAudioTrack(ctx, source, work); err != nil {
		_ = transcode.storage.Purge(ctx, work)
		return transcode.audioTrackTransientFailure(ctx, job, err, domain.RenderFailed)
	}
	if err := transcode.storage.Publish(ctx, work, audioHLSKey(job.AssetID, job.TrackID)); err != nil {
		_ = transcode.storage.Purge(ctx, work)
		return transcode.audioTrackTransientFailure(ctx, job, err, domain.PublishFailed)
	}
	evicted, err := transcode.repository.PromoteAudioTrack(ctx, job.AssetID, job.TrackID, transcode.timestamp(), func(ready []domain.AudioTrack) error {
		return transcode.transcoder.RegenerateMaster(ctx, job.AssetID, audioRenditions(ready))
	})
	if err != nil {
		return transcode.audioTrackTransientFailure(ctx, job, err, domain.RenderFailed)
	}
	if err := transcode.storage.Purge(ctx, work); err != nil {
		return TranscodeOutcome{}, err
	}
	// Promoting a video-only asset to Ready can evict an older generation of the
	// same lesson (D4), exactly as the video publish path does; reclaim its bytes.
	if evicted != nil {
		if err := transcode.purgeAsset(ctx, evicted.ID); err != nil {
			return TranscodeOutcome{}, err
		}
		if err := transcode.repository.DeleteSwept(ctx, evicted.ID); err != nil {
			return TranscodeOutcome{}, err
		}
	}
	return TranscodeOutcome{}, nil
}

// audioTrackTransientFailure retries a transient audio-transcode fault while asynq
// attempts remain, but once they are exhausted it fails the track terminally (T12)
// instead of returning a retryable error. An exhausted asynq task never runs again,
// so a transient return would strand the track Queued and the asset in AwaitingAudio
// forever — the gap the video path avoids with its RetryCount>=MaxRetry fail in
// businessFailure. There is no track-level Recover sweep to fall back on, so
// exhaustion is always terminal here. finalizationContext detaches cancellation so
// the fail still commits when the job context is already done.
func (transcode *Transcode) audioTrackTransientFailure(ctx context.Context, job ports.Job, cause error, reason domain.FailureReason) (TranscodeOutcome, error) {
	if job.RetryCount < job.MaxRetry {
		return TranscodeOutcome{}, mediaOperationError(ctx, cause)
	}
	finalCtx, cancel := transcode.finalizationContext(ctx)
	defer cancel()
	if err := transcode.repository.FailAudioTrack(finalCtx, job.AssetID, job.TrackID, transcode.timestamp()); err != nil {
		return TranscodeOutcome{}, err
	}
	return TranscodeOutcome{Failed: true, Reason: reason}, nil
}

func audioRenditions(tracks []domain.AudioTrack) []ports.AudioRendition {
	renditions := make([]ports.AudioRendition, 0, len(tracks))
	for _, track := range tracks {
		renditions = append(renditions, ports.AudioRendition{
			TrackID:  track.ID,
			Name:     track.Label,
			Language: track.Language,
			Default:  track.IsDefault,
			Embedded: track.Kind == domain.EmbeddedTrack,
		})
	}
	return renditions
}

func (transcode *Transcode) Sweep(ctx context.Context) error {
	callCtx, cancel := context.WithTimeout(ctx, mediaOperationTimeout)
	defer cancel()
	return mediaOperationError(ctx, transcode.sweep(callCtx, transcode.timestamp()))
}

func (transcode *Transcode) sweep(ctx context.Context, now time.Time) error {
	assets, err := transcode.repository.SweepCandidates(ctx, now, sweepLimit)
	if err != nil {
		return err
	}
	for _, asset := range assets {
		if err := transcode.purgeAsset(ctx, asset.ID); err != nil {
			return err
		}
		if err := transcode.repository.DeleteSwept(ctx, asset.ID); err != nil {
			return err
		}
	}
	tracks, err := transcode.repository.SweepAudioTrackCandidates(ctx, now, sweepLimit)
	if err != nil {
		return err
	}
	for _, track := range tracks {
		track := track
		err := transcode.repository.DeleteSweptAudioTrackLocked(ctx, track.ID, now, func() error {
			if err := transcode.storage.Purge(ctx, audioUploadScratchKey(track.VideoAssetID, track.ID)); err != nil {
				return err
			}
			return transcode.storage.Purge(ctx, audioSourceKey(track.VideoAssetID, track.ID))
		})
		if errors.Is(err, domain.ErrInvalidStateTransition) {
			continue
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (transcode *Transcode) Recover(ctx context.Context) error {
	callCtx, cancel := context.WithTimeout(ctx, mediaOperationTimeout)
	defer cancel()
	now := transcode.timestamp()
	assets, err := transcode.repository.Recover(callCtx, now, sweepLimit)
	if err != nil {
		return mediaOperationError(ctx, err)
	}
	for _, asset := range assets {
		if err := transcode.queue.Enqueue(callCtx, asset.ID, asset.Attempt); err != nil {
			return mediaOperationError(ctx, err)
		}
	}
	tracks, err := transcode.repository.RecoverQueuedAudioTracks(callCtx, now, sweepLimit)
	if err != nil {
		return mediaOperationError(ctx, err)
	}
	for _, track := range tracks {
		if err := transcode.queue.EnqueueAudioTrack(callCtx, track.VideoAssetID, track.ID, 1); err != nil {
			return mediaOperationError(ctx, err)
		}
	}
	return nil
}

func (transcode *Transcode) businessFailure(ctx context.Context, job ports.Job, token domain.ClaimToken, reason domain.FailureReason) (TranscodeOutcome, error) {
	finalCtx, cancel := transcode.finalizationContext(ctx)
	defer cancel()
	if err := transcode.cleanup(finalCtx, job.AssetID, token); err != nil {
		return TranscodeOutcome{}, err
	}
	if reason.Permanent() || job.RetryCount >= job.MaxRetry {
		if err := transcode.repository.Fail(finalCtx, job.AssetID, domain.Processing, token, reason, transcode.timestamp()); err != nil {
			return TranscodeOutcome{}, err
		}
	} else if err := transcode.repository.Requeue(finalCtx, job.AssetID, token, transcode.timestamp()); err != nil {
		return TranscodeOutcome{}, err
	}
	return TranscodeOutcome{Failed: true, Reason: reason, Permanent: reason.Permanent()}, nil
}

func (transcode *Transcode) dependencyFailure(ctx context.Context, job ports.Job, token domain.ClaimToken, dependencyErr error) (TranscodeOutcome, error) {
	finalCtx, cancel := transcode.finalizationContext(ctx)
	defer cancel()
	if err := transcode.cleanup(finalCtx, job.AssetID, token); err != nil {
		return TranscodeOutcome{}, dependencyErr
	}
	if err := transcode.repository.Requeue(finalCtx, job.AssetID, token, transcode.timestamp()); err != nil {
		return TranscodeOutcome{}, dependencyErr
	}
	return TranscodeOutcome{}, dependencyErr
}

func (transcode *Transcode) transcoderFailure(ctx context.Context, job ports.Job, token domain.ClaimToken, err error, fallback domain.FailureReason) (TranscodeOutcome, error) {
	if ctxErr := ctx.Err(); ctxErr != nil {
		return transcode.cancelled(ctx, job, token, ctxErr)
	}
	if errors.Is(err, domain.ErrDependencyUnavailable) {
		return transcode.dependencyFailure(ctx, job, token, err)
	}
	if diskUnavailable(err) {
		fallback = domain.StorageUnavailable
	}
	return transcode.businessFailure(ctx, job, token, fallback)
}

func (transcode *Transcode) storageFailure(ctx context.Context, job ports.Job, token domain.ClaimToken, err error, fallback domain.FailureReason) (TranscodeOutcome, error) {
	if ctxErr := ctx.Err(); ctxErr != nil {
		return transcode.cancelled(ctx, job, token, ctxErr)
	}
	if errors.Is(err, domain.ErrDependencyUnavailable) {
		return transcode.dependencyFailure(ctx, job, token, err)
	}
	if diskUnavailable(err) {
		fallback = domain.StorageUnavailable
	}
	return transcode.businessFailure(ctx, job, token, fallback)
}

func (transcode *Transcode) cancelled(ctx context.Context, job ports.Job, token domain.ClaimToken, cause error) (TranscodeOutcome, error) {
	finalCtx, cancel := transcode.finalizationContext(ctx)
	defer cancel()
	if err := transcode.cleanup(finalCtx, job.AssetID, token); err != nil {
		return TranscodeOutcome{}, err
	}
	if errors.Is(cause, context.DeadlineExceeded) && job.RetryCount >= job.MaxRetry {
		if err := transcode.repository.Fail(finalCtx, job.AssetID, domain.Processing, token, domain.RenderFailed, transcode.timestamp()); err != nil {
			return TranscodeOutcome{}, err
		}
		return TranscodeOutcome{Failed: true, Reason: domain.RenderFailed}, nil
	}
	if err := transcode.repository.Requeue(finalCtx, job.AssetID, token, transcode.timestamp()); err != nil {
		return TranscodeOutcome{}, err
	}
	return TranscodeOutcome{}, cause
}

func (transcode *Transcode) finalizationContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.WithoutCancel(ctx), mediaOperationTimeout)
}

func (transcode *Transcode) cleanup(ctx context.Context, id string, token domain.ClaimToken) error {
	if err := transcode.storage.Purge(ctx, renderWorkKey(id, token)); err != nil {
		return err
	}
	return transcode.storage.Purge(ctx, renderThumbnailKey(id, token))
}

func (transcode *Transcode) purgeAsset(ctx context.Context, id string) error {
	for _, key := range []domain.StorageKey{
		{Kind: domain.StorageSource, ID: id},
		{Kind: domain.StorageHLS, ID: id},
		{Kind: domain.StorageThumbnail, ID: id},
		{Kind: domain.StorageTemporary, ID: id},
	} {
		if err := transcode.storage.Purge(ctx, key); err != nil {
			return err
		}
	}
	return nil
}

func (transcode *Transcode) timestamp() time.Time {
	return transcode.now().UTC().Truncate(mediaStoreResolution)
}

type processDisposition uint8

const (
	processInvalid processDisposition = iota
	processDone
	processContinue
)

func dispositionOf(asset domain.VideoAsset) processDisposition {
	if asset.EvictedAt != nil || asset.LessonID == "" {
		return processDone
	}
	switch asset.State {
	case domain.Ready, domain.Failed, domain.Created, domain.Uploading:
		return processDone
	case domain.Queued, domain.Processing:
		return processContinue
	default:
		return processInvalid
	}
}

// embeddedTrackLabel and embeddedTrackLanguage name the derived embedded track
// until an author renames it. "und" is the ISO 639-2 undetermined-language tag.
const (
	embeddedTrackLabel    = "Original"
	embeddedTrackLanguage = "und"
)

// embeddedTrack builds the Ready embedded audio track materialised when a video
// source carries its own audio (D3/D4). Its source leaf is the shared video
// source ("original"), so the id can be freshly minted here without affecting any
// storage path.
func embeddedTrack(assetID string, now time.Time) (domain.AudioTrack, error) {
	id, err := freshUUID()
	if err != nil {
		return domain.AudioTrack{}, err
	}
	return domain.NewEmbeddedTrack(domain.AudioTrack{
		ID:           id,
		VideoAssetID: assetID,
		Kind:         domain.EmbeddedTrack,
		Language:     embeddedTrackLanguage,
		Label:        embeddedTrackLabel,
		IsDefault:    true,
		State:        domain.Ready,
		SourceLeaf:   domain.EmbeddedAudioLeaf,
		CreatedAt:    now,
		UpdatedAt:    now,
	})
}

func freshUUID() (string, error) {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", err
	}
	value[6] = (value[6] & 0x0f) | 0x40 // version 4
	value[8] = (value[8] & 0x3f) | 0x80 // variant 10
	return fmt.Sprintf("%x-%x-%x-%x-%x", value[0:4], value[4:6], value[6:8], value[8:10], value[10:16]), nil
}

func freshClaimToken() (domain.ClaimToken, error) {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "", err
	}
	return domain.ClaimToken(hex.EncodeToString(bytes[:])), nil
}

func diskUnavailable(err error) bool {
	return errors.Is(err, syscall.ENOSPC) || errors.Is(err, syscall.EROFS)
}

func renderWorkKey(id string, token domain.ClaimToken) domain.StorageKey {
	return domain.StorageKey{Kind: domain.StorageTemporary, ID: id, Leaf: string(token)}
}

func renderThumbnailKey(id string, token domain.ClaimToken) domain.StorageKey {
	return domain.StorageKey{Kind: domain.StorageTemporary, ID: id, Leaf: string(token) + ".jpg"}
}

func hlsKey(id string) domain.StorageKey {
	return domain.StorageKey{Kind: domain.StorageHLS, ID: id}
}

func audioSourceKey(assetID, trackID string) domain.StorageKey {
	return domain.StorageKey{Kind: domain.StorageSource, ID: assetID, Leaf: domain.AudioLeafPrefix + trackID}
}

func audioHLSKey(assetID, trackID string) domain.StorageKey {
	return domain.StorageKey{Kind: domain.StorageHLS, ID: assetID, Leaf: domain.AudioLeafPrefix + trackID}
}

func thumbnailKey(id string) domain.StorageKey {
	return domain.StorageKey{Kind: domain.StorageThumbnail, ID: id}
}
