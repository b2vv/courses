package application

import (
	"bytes"
	"context"
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
	if _, err := u.storage.Put(ctx, image.ID, bytes.NewReader(data)); err != nil {
		return domain.Image{}, err
	}
	// Insert failure orphans stored bytes in this tiny window after a successful write; v1 has no image sweeper.
	if err := u.repository.Insert(ctx, image); err != nil {
		return domain.Image{}, err
	}
	return image, nil
}
