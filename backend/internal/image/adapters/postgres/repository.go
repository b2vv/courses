package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"courses/internal/image/domain"
	"courses/internal/image/ports"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct{ pool *pgxpool.Pool }

func NewRepository(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

var _ ports.Repository = (*Repository)(nil)

func (r *Repository) Insert(ctx context.Context, image domain.Image) error {
	_, err := r.pool.Exec(ctx, `
INSERT INTO lesson_images (id, lesson_id, content_type, width, height, byte_size, created_by, created_at)
VALUES ($1::text::uuid, $2::text::uuid, $3, $4, $5, $6, $7::text::uuid, $8)`,
		image.ID, image.LessonID, image.ContentType, image.Width, image.Height, image.ByteSize, image.CreatedBy, image.CreatedAt)
	if err != nil {
		return persistenceError(err)
	}
	return nil
}

func (r *Repository) FindByID(ctx context.Context, id string) (domain.Image, error) {
	var image domain.Image
	var createdBy pgtype.Text
	err := r.pool.QueryRow(ctx, `
SELECT id::text, lesson_id::text, content_type, width, height, byte_size, created_by::text, created_at
FROM lesson_images WHERE id = $1::text::uuid`, id).Scan(
		&image.ID, &image.LessonID, &image.ContentType, &image.Width, &image.Height, &image.ByteSize, &createdBy, &image.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Image{}, domain.ErrImageNotFound
	}
	if err != nil {
		return domain.Image{}, persistenceError(err)
	}
	if createdBy.Valid {
		image.CreatedBy = createdBy.String
	}
	return image, nil
}

func persistenceError(err error) error {
	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) {
		// Connection, resource, shutdown, rollback and lock failures are transient.
		for _, class := range []string{"08", "53", "57", "58", "40", "55"} {
			if strings.HasPrefix(postgresError.Code, class) {
				return fmt.Errorf("image persistence unavailable: %w", err)
			}
		}
		return fmt.Errorf("image persistence failure: %w", err)
	}
	return fmt.Errorf("image persistence unavailable: %w", err)
}
