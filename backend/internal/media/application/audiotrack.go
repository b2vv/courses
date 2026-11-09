package application

import (
	"context"
	"time"

	"courses/internal/media/domain"
	"courses/internal/media/ports"
)

// AudioTracks is the authoring service for a lesson's audio tracks (S5-L1): list,
// rename, set-default and delete. Ownership is enforced inside the repository (the
// same course-instructor join the video authoring paths use); this service adds
// input validation and the master regeneration every mutation needs, because the
// default flag, the label and the language are all rendered into the master's
// EXT-X-MEDIA lines (D7). The rewrite runs inside the repository's transaction,
// under the asset lock promote also takes, so the API and the media worker never
// write one master concurrently.
type AudioTracks struct {
	repository ports.AssetRepository
	master     ports.MasterWriter
	now        Clock
}

func NewAudioTracks(repository ports.AssetRepository, master ports.MasterWriter, now Clock) *AudioTracks {
	return &AudioTracks{repository: repository, master: master, now: now}
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

func (service *AudioTracks) Rename(ctx context.Context, actor domain.Actor, assetID, trackID, language, label string) (domain.AudioTrack, error) {
	if err := domain.ValidateAudioTrackMetadata(language, label); err != nil {
		return domain.AudioTrack{}, err
	}
	callCtx, cancel := context.WithTimeout(ctx, mediaOperationTimeout)
	defer cancel()
	track, err := service.repository.UpdateAudioTrackMetadata(callCtx, actor, assetID, trackID, language, label, service.timestamp(), service.rewrite(callCtx, assetID))
	return track, mediaOperationError(ctx, err)
}

func (service *AudioTracks) Delete(ctx context.Context, actor domain.Actor, assetID, trackID string) error {
	callCtx, cancel := context.WithTimeout(ctx, mediaOperationTimeout)
	defer cancel()
	return mediaOperationError(ctx, service.repository.DeleteAudioTrack(callCtx, actor, assetID, trackID, service.timestamp(), service.rewrite(callCtx, assetID)))
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
