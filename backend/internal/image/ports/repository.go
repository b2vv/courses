package ports

import (
	"context"

	"courses/internal/image/domain"
)

// Repository persists image metadata and resolves an image to its lesson for authorization.
type Repository interface {
	Insert(ctx context.Context, image domain.Image) error
	FindByID(ctx context.Context, id string) (domain.Image, error)
}
