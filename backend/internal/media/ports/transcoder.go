package ports

import (
	"context"

	"courses/internal/media/domain"
)

type Transcoder interface {
	Probe(context.Context, domain.StorageKey) (domain.SourceProbe, error)
	Render(context.Context, domain.StorageKey, domain.StorageKey, []domain.Rendition) error
	Thumbnail(context.Context, domain.StorageKey, domain.StorageKey) error
}
