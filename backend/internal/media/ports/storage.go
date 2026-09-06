package ports

import (
	"context"
	"io"

	"courses/internal/media/domain"
)

type MediaStorage interface {
	Free(ctx context.Context) (int64, error)
	Size(ctx context.Context, key domain.StorageKey) (int64, error)
	Head(ctx context.Context, key domain.StorageKey, n int) ([]byte, error)
	Append(ctx context.Context, key domain.StorageKey, offset int64, body io.Reader) (int64, error)
	Truncate(ctx context.Context, key domain.StorageKey, size int64) error
	Move(ctx context.Context, from, to domain.StorageKey) error
	Publish(ctx context.Context, work, destination domain.StorageKey) error
	Purge(ctx context.Context, key domain.StorageKey) error
}
