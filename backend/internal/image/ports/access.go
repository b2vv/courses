package ports

import (
	"context"

	"courses/internal/image/domain"
)

// AccessReader loads the per-viewer course access facts for the lesson that owns an image.
type AccessReader interface {
	LoadImageAccess(ctx context.Context, userID, imageID string) (domain.ImageAccess, error)
}

// AuthoringReader resolves the instructor who owns the lesson, for upload authorization.
type AuthoringReader interface {
	LoadLessonOwner(ctx context.Context, lessonID string) (ownerID string, found bool, err error)
}
