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
// the input validation and, on delete, the master regeneration from the remaining
// Ready set.
type AudioTracks struct {
	repository ports.AssetRepository
	transcoder ports.Transcoder
	now        Clock
}

func NewAudioTracks(repository ports.AssetRepository, transcoder ports.Transcoder, now Clock) *AudioTracks {
	return &AudioTracks{repository: repository, transcoder: transcoder, now: now}
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
	return mediaOperationError(ctx, service.repository.SetDefaultAudioTrack(callCtx, actor, assetID, trackID, service.timestamp()))
}

func (service *AudioTracks) Rename(ctx context.Context, actor domain.Actor, assetID, trackID, language, label string) (domain.AudioTrack, error) {
	if err := domain.ValidateAudioTrackMetadata(language, label); err != nil {
		return domain.AudioTrack{}, err
	}
	callCtx, cancel := context.WithTimeout(ctx, mediaOperationTimeout)
	defer cancel()
	track, err := service.repository.UpdateAudioTrackMetadata(callCtx, actor, assetID, trackID, language, label, service.timestamp())
	return track, mediaOperationError(ctx, err)
}

func (service *AudioTracks) Delete(ctx context.Context, actor domain.Actor, assetID, trackID string) error {
	callCtx, cancel := context.WithTimeout(ctx, mediaOperationTimeout)
	defer cancel()
	return mediaOperationError(ctx, service.repository.DeleteAudioTrack(callCtx, actor, assetID, trackID, service.timestamp(), func(ready []domain.AudioTrack) error {
		return service.transcoder.RegenerateMaster(callCtx, assetID, audioRenditions(ready))
	}))
}

func (service *AudioTracks) timestamp() time.Time {
	return service.now().UTC().Truncate(mediaStoreResolution)
}
