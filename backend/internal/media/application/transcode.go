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
	if err := transcode.transcoder.Render(callCtx, source, work, renditions); err != nil {
		return transcode.transcoderFailure(callCtx, job, token, err, domain.RenderFailed)
	}
	if err := transcode.transcoder.Thumbnail(callCtx, source, workThumbnail, probe); err != nil {
		return transcode.transcoderFailure(callCtx, job, token, err, domain.RenderFailed)
	}

	var policyErr error
	published, evicted, err := transcode.repository.Publish(callCtx, asset.ID, token, renditions, probe.UsableDurationSeconds(), transcode.timestamp(), func(domain.VideoAsset) error {
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
	return nil
}

func (transcode *Transcode) Recover(ctx context.Context) error {
	callCtx, cancel := context.WithTimeout(ctx, mediaOperationTimeout)
	defer cancel()
	assets, err := transcode.repository.Recover(callCtx, transcode.timestamp(), sweepLimit)
	if err != nil {
		return mediaOperationError(ctx, err)
	}
	for _, asset := range assets {
		if err := transcode.queue.Enqueue(callCtx, asset.ID, asset.Attempt); err != nil {
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

func thumbnailKey(id string) domain.StorageKey {
	return domain.StorageKey{Kind: domain.StorageThumbnail, ID: id}
}
