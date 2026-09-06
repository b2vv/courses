package ports

import (
	"context"
	"io"

	"courses/internal/media/domain"
)

type MediaStorage interface {
	Free(context.Context) (int64, error)
	Size(context.Context, domain.StorageKey) (int64, error)
	Append(context.Context, domain.StorageKey, int64, io.Reader) (int64, error)
	Truncate(context.Context, domain.StorageKey, int64) error
	Move(context.Context, domain.StorageKey, domain.StorageKey) error
	Publish(context.Context, domain.StorageKey, domain.StorageKey) error
	Purge(context.Context, domain.StorageKey) error
}
