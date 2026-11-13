package application

import (
	"context"
	"errors"
	"io"
	"os"
	"syscall"
	"time"

	"courses/internal/media/domain"
	"courses/internal/media/ports"
)

type AudioUpload struct {
	repository ports.AssetRepository
	storage    ports.MediaStorage
	queue      ports.TranscodingQueue
	now        Clock
}

func NewAudioUpload(repository ports.AssetRepository, storage ports.MediaStorage, queue ports.TranscodingQueue, now Clock) *AudioUpload {
	return &AudioUpload{repository: repository, storage: storage, queue: queue, now: now}
}

func (upload *AudioUpload) HeadUpload(ctx context.Context, trackID string) (domain.AudioTrack, error) {
	callCtx, cancel := context.WithTimeout(ctx, mediaOperationTimeout)
	defer cancel()
	track, err := upload.repository.LoadAudioTrackUpload(callCtx, trackID)
	if err != nil {
		return domain.AudioTrack{}, mediaOperationError(ctx, err)
	}
	now := upload.timestamp()
	if !track.ExpiresAt.After(now) || (track.State != domain.Created && track.State != domain.Uploading) {
		return domain.AudioTrack{}, domain.ErrVideoNotFound
	}
	return track, nil
}

func (upload *AudioUpload) AppendChunk(ctx context.Context, trackID string, offset int64, body io.Reader) (domain.AudioTrack, error) {
	callCtx, cancel := context.WithTimeout(ctx, uploadChunkTimeout)
	defer cancel()

	track, err := upload.repository.AppendAudioChunk(callCtx, trackID, offset, upload.timestamp(), func(locked domain.AudioTrack) (int64, error) {
		scratch := audioUploadScratchKey(locked.VideoAssetID, locked.ID)
		if locked.ReceivedBytes == 0 {
			complete, err := upload.completedSource(callCtx, locked)
			if err != nil {
				return 0, err
			}
			if complete {
				return locked.DeclaredBytes, nil
			}
		}
		if locked.State == domain.Created {
			free, err := upload.storage.Free(callCtx)
			if err != nil {
				return 0, err
			}
			if free < 3*locked.DeclaredBytes {
				return 0, syscall.ENOSPC
			}
		}
		if locked.ReceivedBytes == 0 {
			if err := upload.storage.Truncate(callCtx, scratch, 0); err != nil {
				return 0, err
			}
		} else {
			size, err := upload.storage.Size(callCtx, scratch)
			if errors.Is(err, os.ErrNotExist) {
				complete, sourceErr := upload.completedSource(callCtx, locked)
				if sourceErr != nil {
					return 0, sourceErr
				}
				if complete {
					return locked.DeclaredBytes, nil
				}
				return 0, err
			}
			if err != nil {
				return 0, err
			}
			if size < locked.ReceivedBytes {
				return 0, io.ErrUnexpectedEOF
			}
			if size > locked.ReceivedBytes {
				if err := upload.storage.Truncate(callCtx, scratch, locked.ReceivedBytes); err != nil {
					return 0, err
				}
			}
		}
		written, err := upload.storage.Append(callCtx, scratch, locked.ReceivedBytes, body)
		if err != nil {
			return 0, upload.rollbackChunk(callCtx, scratch, locked.ReceivedBytes, err)
		}
		newReceived := locked.ReceivedBytes + written
		if newReceived > locked.DeclaredBytes {
			return 0, upload.rollbackChunk(callCtx, scratch, locked.ReceivedBytes, domain.ErrInvalidRequest)
		}
		return newReceived, nil
	}, func(locked domain.AudioTrack) error {
		complete, err := upload.completedSource(callCtx, locked)
		if err != nil {
			return err
		}
		if complete {
			return nil
		}
		return upload.storage.Move(callCtx,
			audioUploadScratchKey(locked.VideoAssetID, locked.ID),
			audioSourceKey(locked.VideoAssetID, locked.ID))
	})
	if err != nil {
		return domain.AudioTrack{}, mediaOperationError(ctx, err)
	}
	if track.State == domain.Queued {
		if err := upload.queue.EnqueueAudioTrack(callCtx, track.VideoAssetID, track.ID, 1); err != nil {
			return domain.AudioTrack{}, mediaOperationError(ctx, err)
		}
	}
	return track, nil
}

func (upload *AudioUpload) CancelUpload(ctx context.Context, trackID string) error {
	callCtx, cancel := context.WithTimeout(ctx, mediaOperationTimeout)
	defer cancel()
	return mediaOperationError(ctx, upload.repository.CancelAudioUpload(callCtx, trackID, upload.timestamp()))
}

func (upload *AudioUpload) rollbackChunk(ctx context.Context, scratch domain.StorageKey, offset int64, cause error) error {
	if err := upload.storage.Truncate(ctx, scratch, offset); err != nil {
		return err
	}
	return cause
}

func (upload *AudioUpload) completedSource(ctx context.Context, track domain.AudioTrack) (bool, error) {
	size, err := upload.storage.Size(ctx, audioSourceKey(track.VideoAssetID, track.ID))
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if size != track.DeclaredBytes {
		return false, io.ErrUnexpectedEOF
	}
	return true, nil
}

func (upload *AudioUpload) timestamp() time.Time {
	return upload.now().UTC().Truncate(mediaStoreResolution)
}

func audioUploadScratchKey(assetID, trackID string) domain.StorageKey {
	return domain.StorageKey{Kind: domain.StorageTemporary, ID: assetID, Leaf: domain.AudioUploadLeafPrefix + trackID}
}
