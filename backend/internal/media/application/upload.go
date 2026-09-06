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

const (
	mediaOperationTimeout = 3 * time.Second
	uploadChunkTimeout    = 120 * time.Second
	ticketLifetime        = 24 * time.Hour
	retryClaimAge         = 40 * time.Minute
	mediaStoreResolution  = time.Microsecond
)

type Clock func() time.Time

type Tickets struct {
	repository ports.AssetRepository
	queue      ports.TranscodingQueue
	now        Clock
}

func NewTickets(repository ports.AssetRepository, queue ports.TranscodingQueue, now Clock) *Tickets {
	return &Tickets{repository: repository, queue: queue, now: now}
}

type Upload struct {
	repository ports.AssetRepository
	storage    ports.MediaStorage
	queue      ports.TranscodingQueue
	now        Clock
}

func NewUpload(repository ports.AssetRepository, storage ports.MediaStorage, queue ports.TranscodingQueue, now Clock) *Upload {
	return &Upload{repository: repository, storage: storage, queue: queue, now: now}
}

func (tickets *Tickets) MintTicket(ctx context.Context, actor domain.Actor, lessonID string, declaredBytes int64, originalName string) (domain.VideoAsset, error) {
	if err := domain.ValidateTicket(declaredBytes, originalName); err != nil {
		return domain.VideoAsset{}, err
	}
	now := tickets.timestamp()
	callCtx, cancel := context.WithTimeout(ctx, mediaOperationTimeout)
	defer cancel()
	asset, err := tickets.repository.Mint(callCtx, actor, lessonID, declaredBytes, originalName, now.Add(ticketLifetime), now)
	return asset, mediaOperationError(ctx, err)
}

func (upload *Upload) AppendChunk(ctx context.Context, id string, offset int64, body io.Reader) (domain.VideoAsset, error) {
	callCtx, cancel := context.WithTimeout(ctx, uploadChunkTimeout)
	defer cancel()

	asset, err := upload.repository.AppendChunk(callCtx, id, offset, upload.timestamp(), func(locked domain.VideoAsset) (int64, error) {
		scratch := uploadScratchKey(locked.ID)
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
		if locked.State == domain.Created {
			head, err := upload.storage.Head(callCtx, scratch, domain.ContainerSniffBytes)
			if err != nil {
				return 0, upload.rollbackChunk(callCtx, scratch, locked.ReceivedBytes, err)
			}
			if err := domain.AcceptContainer(head); err != nil {
				return 0, upload.rollbackChunk(callCtx, scratch, locked.ReceivedBytes, err)
			}
		}
		return newReceived, nil
	}, func(locked domain.VideoAsset) error {
		complete, err := upload.completedSource(callCtx, locked)
		if err != nil {
			return err
		}
		if complete {
			return nil
		}
		return upload.storage.Move(callCtx, uploadScratchKey(locked.ID), sourceKey(locked.ID))
	})
	if errors.Is(err, domain.ErrUnsupportedSource) {
		return domain.VideoAsset{}, domain.ErrUnsupportedMedia
	}
	if err != nil {
		return domain.VideoAsset{}, mediaOperationError(ctx, err)
	}
	if asset.State == domain.Queued {
		if err := upload.queue.Enqueue(callCtx, asset.ID, asset.Attempt); err != nil {
			return domain.VideoAsset{}, mediaOperationError(ctx, err)
		}
	}
	return asset, nil
}

func (upload *Upload) rollbackChunk(ctx context.Context, scratch domain.StorageKey, offset int64, cause error) error {
	if err := upload.storage.Truncate(ctx, scratch, offset); err != nil {
		return err
	}
	return cause
}

func (upload *Upload) completedSource(ctx context.Context, asset domain.VideoAsset) (bool, error) {
	size, err := upload.storage.Size(ctx, sourceKey(asset.ID))
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if size != asset.DeclaredBytes {
		return false, io.ErrUnexpectedEOF
	}
	return true, nil
}

func (upload *Upload) HeadUpload(ctx context.Context, id string) (domain.VideoAsset, error) {
	callCtx, cancel := context.WithTimeout(ctx, mediaOperationTimeout)
	defer cancel()
	asset, err := upload.repository.Load(callCtx, id)
	if err != nil {
		return domain.VideoAsset{}, mediaOperationError(ctx, err)
	}
	now := upload.timestamp()
	if asset.EvictedAt != nil || !asset.ExpiresAt.After(now) || asset.State != domain.Created && asset.State != domain.Uploading {
		return domain.VideoAsset{}, domain.ErrVideoNotFound
	}
	return asset, nil
}

func (upload *Upload) CancelUpload(ctx context.Context, id string) error {
	callCtx, cancel := context.WithTimeout(ctx, mediaOperationTimeout)
	defer cancel()
	return mediaOperationError(ctx, upload.repository.Cancel(callCtx, id, upload.timestamp()))
}

func (tickets *Tickets) ListCourseVideos(ctx context.Context, actor domain.Actor, courseID string) ([]domain.VideoAsset, error) {
	callCtx, cancel := context.WithTimeout(ctx, mediaOperationTimeout)
	defer cancel()
	assets, err := tickets.repository.ListCourseVideos(callCtx, actor, courseID)
	return assets, mediaOperationError(ctx, err)
}

func (tickets *Tickets) Retry(ctx context.Context, actor domain.Actor, id string) (domain.VideoAsset, error) {
	callCtx, cancel := context.WithTimeout(ctx, mediaOperationTimeout)
	defer cancel()
	asset, err := tickets.repository.Retry(callCtx, actor, id, retryClaimAge, tickets.timestamp())
	if err != nil {
		return domain.VideoAsset{}, mediaOperationError(ctx, err)
	}
	if err := tickets.queue.Enqueue(callCtx, asset.ID, asset.Attempt); err != nil {
		return domain.VideoAsset{}, mediaOperationError(ctx, err)
	}
	return asset, nil
}

func (tickets *Tickets) timestamp() time.Time {
	return tickets.now().UTC().Truncate(mediaStoreResolution)
}

func (upload *Upload) timestamp() time.Time {
	return upload.now().UTC().Truncate(mediaStoreResolution)
}

func mediaOperationError(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}
	if ctxErr := ctx.Err(); ctxErr != nil {
		return ctxErr
	}
	return err
}

func sourceKey(id string) domain.StorageKey {
	return domain.StorageKey{Kind: domain.StorageSource, ID: id, Leaf: "original"}
}

func uploadScratchKey(id string) domain.StorageKey {
	return domain.StorageKey{Kind: domain.StorageTemporary, ID: id, Leaf: "upload"}
}
