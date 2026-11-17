package application

import (
	"context"
	"errors"
	"time"

	"courses/internal/media/domain"
	"courses/internal/media/ports"
)

// AudioTracks is the authoring service for a lesson's audio tracks (S5-L1): list,
// rename, set-default and delete. Ownership is enforced inside the repository (the
// same course-instructor join the video authoring paths use); this service adds
// input validation and master regeneration for manifest metadata changes, because the
// default flag, the label and the language are all rendered into the master's
// EXT-X-MEDIA lines (D7). The rewrite runs inside the repository's transaction,
// under the asset lock promote also takes, so the API and the media worker never
// write one master concurrently.
type AudioTracks struct {
	repository          ports.AssetRepository
	master              ports.MasterWriter
	now                 Clock
	multilingualDefault bool
}

func NewAudioTracks(repository ports.AssetRepository, master ports.MasterWriter, multilingualDefault bool, now Clock) *AudioTracks {
	return &AudioTracks{repository: repository, master: master, now: now, multilingualDefault: multilingualDefault}
}

func (service *AudioTracks) MintAudioTicket(ctx context.Context, actor domain.Actor, assetID string, declaredBytes int64, originalName, language, label string, multilingual *bool) (domain.AudioTrack, error) {
	if err := domain.ValidateAudioTicket(declaredBytes, originalName); err != nil {
		return domain.AudioTrack{}, err
	}
	if err := domain.ValidateAudioTrackMetadata(language, label); err != nil {
		return domain.AudioTrack{}, err
	}
	now := service.timestamp()
	callCtx, cancel := context.WithTimeout(ctx, mediaOperationTimeout)
	defer cancel()
	mode := service.multilingualDefault
	if multilingual != nil {
		mode = *multilingual
	}
	track, err := service.repository.MintAudioTrack(callCtx, actor, assetID, declaredBytes, originalName, language, label, mode, now.Add(ticketLifetime), now)
	return track, mediaOperationError(ctx, err)
}

func (service *AudioTracks) List(ctx context.Context, actor domain.Actor, assetID string) ([]domain.AudioTrack, error) {
	callCtx, cancel := context.WithTimeout(ctx, mediaOperationTimeout)
	defer cancel()
	tracks, err := service.repository.ListAudioTracks(callCtx, actor, assetID)
	return tracks, mediaOperationError(ctx, err)
}

func (service *AudioTracks) SetDefault(ctx context.Context, actor domain.Actor, assetID, trackID string) error {
	callCtx, cancel := context.WithTimeout(ctx, mediaOperationTimeout)
	defer cancel()
	return mediaOperationError(ctx, service.repository.SetDefaultAudioTrack(callCtx, actor, assetID, trackID, service.timestamp(), service.rewrite(callCtx, assetID)))
}

func (service *AudioTracks) Rename(ctx context.Context, actor domain.Actor, assetID, trackID, language, label string, multilingual *bool) (domain.AudioTrack, error) {
	if err := domain.ValidateAudioTrackMetadata(language, label); err != nil {
		return domain.AudioTrack{}, err
	}
	callCtx, cancel := context.WithTimeout(ctx, mediaOperationTimeout)
	defer cancel()
	track, err := service.repository.UpdateAudioTrackMetadata(callCtx, actor, assetID, trackID, language, label, multilingual, service.timestamp(), service.rewrite(callCtx, assetID))
	return track, mediaOperationError(ctx, err)
}

func (service *AudioTracks) Delete(ctx context.Context, actor domain.Actor, assetID, trackID string) error {
	callCtx, cancel := context.WithTimeout(ctx, mediaOperationTimeout)
	defer cancel()
	return mediaOperationError(ctx, service.repository.DeleteAudioTrack(callCtx, actor, assetID, trackID, service.timestamp(), service.rewrite(callCtx, assetID)))
}

// MasterRegeneration is the outcome of RegenerateAllMasters: how many masters were rewritten,
// which assets disappeared (evicted or deleted) between the listing and their rewrite, and which
// failed.
type MasterRegeneration struct {
	Regenerated int
	Skipped     []string
	Failed      []string
}

// RegenerateAllMasters rewrites the master of every Ready asset from the database (MS1 M-D4):
// a migration can change track metadata the manifest renders (00030 relabels legacy `und` tracks)
// but cannot write files. Each rewrite goes through the same serialised path a rename takes, so
// it is safe against concurrent authoring and can be repeated. One asset's failure does not stop
// the others; the error is only for the listing itself.
func (service *AudioTracks) RegenerateAllMasters(ctx context.Context) (MasterRegeneration, error) {
	ids, err := service.repository.ListReadyAssetIDs(ctx)
	if err != nil {
		return MasterRegeneration{}, err
	}
	result := MasterRegeneration{Skipped: []string{}, Failed: []string{}}
	for _, id := range ids {
		callCtx, cancel := context.WithTimeout(ctx, mediaOperationTimeout)
		err := service.repository.RegenerateMaster(callCtx, id, service.rewrite(callCtx, id))
		cancel()
		switch {
		case err == nil:
			result.Regenerated++
		case errors.Is(err, domain.ErrVideoNotFound):
			result.Skipped = append(result.Skipped, id)
		default:
			result.Failed = append(result.Failed, id)
		}
	}
	return result, nil
}

// rewrite regenerates the master from the Ready set the repository hands it. An
// empty set means the asset is not Ready (D4 forbids a Ready asset without a Ready
// track), so there is no audio section to publish and the master is left as is;
// RegenerateMaster would refuse an empty set anyway.
func (service *AudioTracks) rewrite(ctx context.Context, assetID string) func([]domain.AudioTrack) error {
	return func(ready []domain.AudioTrack) error {
		if len(ready) == 0 {
			return nil
		}
		return service.master.RegenerateMaster(ctx, assetID, audioRenditions(ready))
	}
}

func (service *AudioTracks) timestamp() time.Time {
	return service.now().UTC().Truncate(mediaStoreResolution)
}

func (service *AudioTracks) MultilingualDefault() bool { return service.multilingualDefault }
