package postgres

import (
	"context"
	"errors"
	"time"

	"courses/internal/playback/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// discussion/adapters/postgres/repository.go has its own copy of this filtered selector;
// course's learning query intentionally uses a sorting selector because it reports UI state.
const loadLessonDuration = `SELECT (
    SELECT assets.duration_seconds
    FROM video_assets AS assets
    WHERE assets.lesson_id=lessons.id
      AND assets.state='ready' AND assets.evicted_at IS NULL
    ORDER BY assets.generation DESC
    LIMIT 1
)
FROM course_lessons AS lessons
WHERE lessons.id=$1::text::uuid`

const ensureProgress = `INSERT INTO lesson_progress(
    user_id,lesson_id,position_seconds,completed_at,client_seq,updated_at
) VALUES($1::text::uuid,$2::text::uuid,0,NULL,-1,$3)
ON CONFLICT (user_id,lesson_id) DO NOTHING`

const loadProgress = `SELECT position_seconds,completed_at,client_seq,updated_at
FROM lesson_progress
WHERE user_id=$1::text::uuid AND lesson_id=$2::text::uuid
FOR UPDATE`

const storeProgress = `UPDATE lesson_progress
SET position_seconds=$3,completed_at=$4,client_seq=$5,updated_at=$6
WHERE user_id=$1::text::uuid AND lesson_id=$2::text::uuid`

type progressTransaction interface {
	QueryRow(context.Context, string, ...any) pgx.Row
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	Commit(context.Context) error
	Rollback(context.Context) error
}

type ProgressWriter struct {
	begin func(context.Context) (progressTransaction, error)
}

func NewProgressWriter(pool *pgxpool.Pool) *ProgressWriter {
	return newProgressWriter(func(ctx context.Context) (progressTransaction, error) {
		return pool.Begin(ctx)
	})
}

func newProgressWriter(begin func(context.Context) (progressTransaction, error)) *ProgressWriter {
	return &ProgressWriter{begin: begin}
}

func (writer *ProgressWriter) Save(
	ctx context.Context,
	userID, lessonID string,
	update domain.ProgressUpdate,
	now time.Time,
) error {
	tx, err := writer.begin(ctx)
	if err != nil {
		return domain.ErrDependencyUnavailable
	}
	defer rollbackProgress(tx)
	now = now.UTC().Truncate(time.Microsecond)

	durationSeconds, err := progressDuration(ctx, tx, lessonID)
	if err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, ensureProgress, userID, lessonID, now); err != nil {
		return domain.ErrDependencyUnavailable
	}
	saved, err := loadStoredProgress(ctx, tx, userID, lessonID)
	if err != nil {
		return err
	}
	next, accepted := domain.ApplyProgress(saved, update, durationSeconds, now)
	if !accepted {
		return nil
	}
	tag, err := tx.Exec(
		ctx,
		storeProgress,
		userID,
		lessonID,
		next.PositionSeconds,
		next.CompletedAt,
		next.ClientSequence,
		next.UpdatedAt,
	)
	if err != nil || tag.RowsAffected() != 1 {
		return domain.ErrDependencyUnavailable
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.ErrDependencyUnavailable
	}
	return nil
}

func progressDuration(ctx context.Context, tx progressTransaction, lessonID string) (*float64, error) {
	var stored pgtype.Numeric
	if err := tx.QueryRow(ctx, loadLessonDuration, lessonID).Scan(&stored); errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrLessonNotFound
	} else if err != nil {
		return nil, domain.ErrDependencyUnavailable
	}
	value, err := stored.Float64Value()
	if err != nil {
		return nil, domain.ErrDependencyUnavailable
	}
	if !value.Valid {
		return nil, nil
	}
	return &value.Float64, nil
}

func loadStoredProgress(ctx context.Context, tx progressTransaction, userID, lessonID string) (domain.Progress, error) {
	var (
		progress    domain.Progress
		completedAt pgtype.Timestamptz
	)
	if err := tx.QueryRow(ctx, loadProgress, userID, lessonID).Scan(
		&progress.PositionSeconds,
		&completedAt,
		&progress.ClientSequence,
		&progress.UpdatedAt,
	); err != nil {
		return domain.Progress{}, domain.ErrDependencyUnavailable
	}
	progress.Present = true
	if completedAt.Valid {
		completed := completedAt.Time
		progress.CompletedAt = &completed
	}
	return progress, nil
}

func rollbackProgress(tx progressTransaction) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_ = tx.Rollback(ctx)
}
