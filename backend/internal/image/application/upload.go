package application

import (
	"bytes"
	"context"
	"errors"
	"time"

	"courses/internal/image/domain"
	"courses/internal/image/ports"
)

type Uploader struct {
	repository ports.Repository
	storage    ports.Storage
	authoring  ports.AuthoringReader
	newID      func() string
	now        func() time.Time
}

func NewUploader(repository ports.Repository, storage ports.Storage, authoring ports.AuthoringReader, newID func() string, now func() time.Time) *Uploader {
	return &Uploader{repository: repository, storage: storage, authoring: authoring, newID: newID, now: now}
}

func (u *Uploader) Upload(ctx context.Context, actor domain.Actor, lessonID, contentType string, width, height int, data []byte) (domain.Image, error) {
	ownerID, found, err := u.authoring.LoadLessonOwner(ctx, lessonID)
	if err != nil {
		return domain.Image{}, err
	}
	if !found {
		return domain.Image{}, domain.ErrLessonNotFound
	}
	if err := actor.CanEditLesson(ownerID); err != nil {
		return domain.Image{}, err
	}
	image, err := domain.NewImage(u.newID(), lessonID, contentType, actor.UserID, width, height, int64(len(data)), u.now())
	if err != nil {
		return domain.Image{}, err
	}
	extension, err := domain.ExtensionFor(image.ContentType)
	if err != nil {
		return domain.Image{}, err
	}
	if _, err := u.storage.Put(ctx, image.ID, extension, bytes.NewReader(data)); err != nil {
		return domain.Image{}, err
	}
	if err := u.repository.Insert(ctx, image); err != nil {
		// The bytes were written but no row references them (the id was just minted here and the
		// insert failed). Remove them so a failed upload does not orphan bytes (LB09a). Join the
		// delete outcome with the insert error, mirroring the storage adapter's own compensation
		// idiom: a delete failure then surfaces without masking the insert error, and errors.Is
		// still sees the insert error so the handler maps it unchanged. A crash between Put and
		// Insert, or a delete that itself fails, still orphans — the strictly smaller residual.
		return domain.Image{}, errors.Join(err, u.storage.Delete(ctx, image.ID, extension))
	}
	return image, nil
}
