package ports

import (
	"context"

	"courses/internal/playback/domain"
)

type OwnershipReader interface {
	Load(context.Context, string) (domain.Ownership, error)
}

type VideoAccessReader interface {
	LoadVideoCourse(context.Context, string, string) (domain.VideoCourse, error)
}
