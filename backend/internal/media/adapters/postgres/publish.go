package postgres

import (
	"context"
	"errors"
	"time"

	"courses/internal/media/adapters/postgres/generated"
	"courses/internal/media/domain"
	"courses/internal/media/ports"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func (repository *Repository) Publish(
	ctx context.Context,
	id string,
	token domain.ClaimToken,
	renditions []domain.Rendition,
	now time.Time,
	policy ports.PublishPolicy,
) (domain.VideoAsset, *domain.VideoAsset, error) {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return domain.VideoAsset{}, nil, persistenceError(err)
	}
	defer rollback(tx)
	now = storeTime(now)

	locked, err := scanAsset(tx.QueryRow(ctx, lockMediaAsset, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.VideoAsset{}, nil, nil
	}
	if err != nil {
		return domain.VideoAsset{}, nil, persistenceError(err)
	}
	if locked.State != domain.Processing || locked.ClaimToken != token || locked.EvictedAt != nil || locked.LessonID == "" {
		return domain.VideoAsset{}, nil, nil
	}

	rows, err := tx.Query(ctx, `SELECT id::text
		FROM video_assets
		WHERE lesson_id=$1::text::uuid AND evicted_at IS NULL
		ORDER BY generation
		FOR UPDATE`, locked.LessonID)
	if err != nil {
		return domain.VideoAsset{}, nil, persistenceError(err)
	}
	for rows.Next() {
		var siblingID string
		if err := rows.Scan(&siblingID); err != nil {
			rows.Close()
			return domain.VideoAsset{}, nil, persistenceError(err)
		}
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return domain.VideoAsset{}, nil, persistenceError(err)
	}

	if err := policy(locked); err != nil {
		return domain.VideoAsset{}, nil, err
	}
	storedRenditions := make([]string, len(renditions))
	for index, rendition := range renditions {
		storedRenditions[index] = string(rendition)
	}
	published, err := scanAsset(tx.QueryRow(ctx, `UPDATE video_assets AS assets
		SET state='ready', claim_token=NULL, failure_reason=NULL, renditions=$3, updated_at=$4
		WHERE id=$1::text::uuid AND state='processing'
		  AND claim_token=$2::text::uuid AND evicted_at IS NULL
		RETURNING `+mediaAssetColumns, id, token, storedRenditions, now))
	if err != nil {
		return domain.VideoAsset{}, nil, persistenceError(err)
	}

	var readyCount int
	if err := tx.QueryRow(ctx, `SELECT count(*) FROM video_assets
		WHERE lesson_id=$1::text::uuid AND state='ready' AND evicted_at IS NULL`, locked.LessonID).Scan(&readyCount); err != nil {
		return domain.VideoAsset{}, nil, persistenceError(err)
	}
	var evicted *domain.VideoAsset
	if readyCount > 2 {
		victim, err := scanAsset(tx.QueryRow(ctx, `UPDATE video_assets AS assets
			SET evicted_at=$2, updated_at=$2
			WHERE id=(
				SELECT id FROM video_assets
				WHERE lesson_id=$1::text::uuid AND state='ready' AND evicted_at IS NULL
				ORDER BY generation
				LIMIT 1
			)
			RETURNING `+mediaAssetColumns, locked.LessonID, now))
		if err != nil {
			return domain.VideoAsset{}, nil, persistenceError(err)
		}
		evicted = &victim
	}
	if err := commit(ctx, tx); err != nil {
		return domain.VideoAsset{}, nil, err
	}
	return published, evicted, nil
}

func (repository *Repository) SweepCandidates(ctx context.Context, now time.Time, limit int) ([]domain.VideoAsset, error) {
	if limit <= 0 {
		return nil, nil
	}
	rows, err := generated.New(repository.pool).SweepMediaCandidates(ctx, generated.SweepMediaCandidatesParams{
		Now:     databaseTime(now),
		MaxRows: int32(limit),
	})
	if err != nil {
		return nil, persistenceError(err)
	}
	assets := make([]domain.VideoAsset, 0, len(rows))
	for _, row := range rows {
		asset, err := mapAsset(
			row.ID, row.LessonID, row.Generation, row.State, row.ClaimToken,
			row.DeclaredBytes, row.ReceivedBytes, row.OriginalName, row.Attempt,
			row.FailureReason, row.Renditions, row.ExpiresAt, row.EvictedAt,
			row.CreatedAt, row.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		assets = append(assets, asset)
	}
	return assets, nil
}

func (repository *Repository) DeleteSwept(ctx context.Context, id string) error {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return persistenceError(err)
	}
	defer rollback(tx)
	if err := generated.New(tx).DeleteSweptMediaAsset(ctx, id); err != nil {
		return persistenceError(err)
	}
	return commit(ctx, tx)
}

func (repository *Repository) Recover(ctx context.Context, now time.Time, limit int) ([]domain.VideoAsset, error) {
	if limit <= 0 {
		return nil, nil
	}
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return nil, persistenceError(err)
	}
	defer rollback(tx)
	now = storeTime(now)
	queries := generated.New(tx)
	processingCount, err := queries.RecoverProcessingMediaAssets(ctx, generated.RecoverProcessingMediaAssetsParams{
		MaxRows: int32(limit),
		Now:     databaseTime(now),
	})
	if err != nil {
		return nil, persistenceError(err)
	}
	assets := make([]domain.VideoAsset, 0)
	remaining := limit - int(processingCount)
	if remaining > 0 {
		rows, err := queries.RecoverQueuedMediaAssets(ctx, generated.RecoverQueuedMediaAssetsParams{
			Now:     databaseTime(now),
			MaxRows: int32(remaining),
		})
		if err != nil {
			return nil, persistenceError(err)
		}
		assets = make([]domain.VideoAsset, 0, len(rows))
		for _, row := range rows {
			asset, err := mapAsset(
				row.ID, row.LessonID, row.Generation, row.State, row.ClaimToken,
				row.DeclaredBytes, row.ReceivedBytes, row.OriginalName, row.Attempt,
				row.FailureReason, row.Renditions, row.ExpiresAt, row.EvictedAt,
				row.CreatedAt, row.UpdatedAt,
			)
			if err != nil {
				return nil, err
			}
			assets = append(assets, asset)
		}
	}
	if err := commit(ctx, tx); err != nil {
		return nil, err
	}
	return assets, nil
}

func databaseTime(value time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: storeTime(value), Valid: true}
}
