package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"courses/internal/media/adapters/postgres/generated"
	"courses/internal/media/domain"
	"courses/internal/media/ports"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

const mediaAssetColumns = `
assets.id::text,
COALESCE(assets.lesson_id::text, ''),
assets.generation,
assets.state,
COALESCE(replace(assets.claim_token::text, '-', ''), ''),
assets.declared_bytes,
assets.received_bytes,
assets.original_name,
assets.attempt,
COALESCE(assets.failure_reason, ''),
assets.renditions,
assets.expires_at,
assets.evicted_at,
assets.created_at,
assets.updated_at`

const lockMediaAsset = `SELECT ` + mediaAssetColumns + `
FROM video_assets AS assets
WHERE assets.id=$1::text::uuid
FOR UPDATE OF assets`

const lockOwnedMediaAsset = `SELECT ` + mediaAssetColumns + `, courses.instructor_id::text
FROM video_assets AS assets
JOIN course_lessons AS lessons ON lessons.id=assets.lesson_id
JOIN course_modules AS modules ON modules.id=lessons.module_id
JOIN courses ON courses.id=modules.course_id
WHERE assets.id=$1::text::uuid AND assets.evicted_at IS NULL
FOR UPDATE OF assets`

type Repository struct{ pool *pgxpool.Pool }

func NewRepository(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

func (repository *Repository) Load(ctx context.Context, id string) (domain.VideoAsset, error) {
	row, err := generated.New(repository.pool).LoadMediaAsset(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.VideoAsset{}, domain.ErrVideoNotFound
	}
	if err != nil {
		return domain.VideoAsset{}, persistenceError(err)
	}
	return mapAsset(
		row.ID, row.LessonID, row.Generation, row.State, row.ClaimToken,
		row.DeclaredBytes, row.ReceivedBytes, row.OriginalName, row.Attempt,
		row.FailureReason, row.Renditions, row.ExpiresAt, row.EvictedAt,
		row.CreatedAt, row.UpdatedAt,
	)
}

func (repository *Repository) ListCourseVideos(ctx context.Context, actor domain.Actor, courseID string) ([]domain.VideoAsset, error) {
	rows, err := generated.New(repository.pool).ListMediaCourseVideos(ctx, courseID)
	if err != nil {
		return nil, persistenceError(err)
	}
	if len(rows) == 0 {
		return nil, domain.ErrCourseForbidden
	}
	if err := actor.CanEditLesson(rows[0].InstructorID); err != nil {
		return nil, err
	}
	assets := make([]domain.VideoAsset, 0, len(rows))
	for _, row := range rows {
		if !row.AssetPresent {
			continue
		}
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

func (repository *Repository) Mint(
	ctx context.Context,
	actor domain.Actor,
	lessonID string,
	declaredBytes int64,
	originalName string,
	expiresAt, now time.Time,
) (domain.VideoAsset, error) {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return domain.VideoAsset{}, persistenceError(err)
	}
	defer rollback(tx)
	now = storeTime(now)
	expiresAt = storeTime(expiresAt)

	var ownerID string
	err = tx.QueryRow(ctx, `SELECT courses.instructor_id::text
		FROM course_lessons AS lessons
		JOIN course_modules AS modules ON modules.id=lessons.module_id
		JOIN courses ON courses.id=modules.course_id
		WHERE lessons.id=$1::text::uuid
		FOR SHARE OF courses`, lessonID).Scan(&ownerID)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.VideoAsset{}, domain.ErrLessonNotFound
	}
	if err != nil {
		return domain.VideoAsset{}, persistenceError(err)
	}
	if err := actor.CanEditLesson(ownerID); err != nil {
		return domain.VideoAsset{}, err
	}

	if _, err := tx.Exec(ctx, `UPDATE video_assets
		SET evicted_at=$2, updated_at=$2
		WHERE lesson_id=$1::text::uuid AND evicted_at IS NULL
		  AND state IN ('created','uploading') AND expires_at <= $2`, lessonID, now); err != nil {
		return domain.VideoAsset{}, persistenceError(err)
	}
	var generation int32
	if err := tx.QueryRow(ctx, `SELECT COALESCE(MAX(generation),0)+1 FROM video_assets WHERE lesson_id=$1::text::uuid`, lessonID).Scan(&generation); err != nil {
		return domain.VideoAsset{}, persistenceError(err)
	}
	asset, err := scanAsset(tx.QueryRow(ctx, `INSERT INTO video_assets AS assets(
		id,lesson_id,generation,state,declared_bytes,original_name,expires_at,created_at,updated_at)
		VALUES(gen_random_uuid(),$1::text::uuid,$2,'created',$3,$4,$5,$6,$6)
		RETURNING `+mediaAssetColumns,
		lessonID, generation, declaredBytes, originalName, expiresAt, now))
	if err != nil {
		if uploadInProgress(err) {
			return domain.VideoAsset{}, domain.ErrUploadInProgress
		}
		return domain.VideoAsset{}, persistenceError(err)
	}
	if _, err := domain.NewAsset(asset); err != nil {
		return domain.VideoAsset{}, err
	}
	if err := commit(ctx, tx); err != nil {
		return domain.VideoAsset{}, err
	}
	return asset, nil
}

func (repository *Repository) AppendChunk(ctx context.Context, id string, offset int64, now time.Time, policy ports.ChunkPolicy) (domain.VideoAsset, error) {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return domain.VideoAsset{}, persistenceError(err)
	}
	defer rollback(tx)
	now = storeTime(now)
	locked, err := scanAsset(tx.QueryRow(ctx, lockMediaAsset, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.VideoAsset{}, domain.ErrVideoNotFound
	}
	if err != nil {
		return domain.VideoAsset{}, persistenceError(err)
	}
	if !activeUpload(locked, now) {
		return domain.VideoAsset{}, domain.ErrVideoNotFound
	}
	if locked.ReceivedBytes != offset {
		return domain.VideoAsset{}, domain.ErrInvalidStateTransition
	}
	newReceived, err := policy(locked)
	if err != nil {
		return domain.VideoAsset{}, err
	}
	if newReceived < offset || newReceived > locked.DeclaredBytes {
		return domain.VideoAsset{}, domain.ErrInvalidRequest
	}
	updated, err := scanAsset(tx.QueryRow(ctx, `UPDATE video_assets AS assets
		SET received_bytes=$3, state='uploading', updated_at=$4
		WHERE id=$1::text::uuid AND received_bytes=$2
		  AND state IN ('created','uploading') AND expires_at>$4 AND evicted_at IS NULL
		RETURNING `+mediaAssetColumns, id, offset, newReceived, now))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.VideoAsset{}, domain.ErrInvalidStateTransition
	}
	if err != nil {
		return domain.VideoAsset{}, persistenceError(err)
	}
	if err := commit(ctx, tx); err != nil {
		return domain.VideoAsset{}, err
	}
	return updated, nil
}

func (repository *Repository) Complete(ctx context.Context, id string, now time.Time, policy ports.CompletePolicy) (domain.VideoAsset, error) {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return domain.VideoAsset{}, persistenceError(err)
	}
	defer rollback(tx)
	now = storeTime(now)
	locked, err := scanAsset(tx.QueryRow(ctx, lockMediaAsset, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.VideoAsset{}, domain.ErrVideoNotFound
	}
	if err != nil {
		return domain.VideoAsset{}, persistenceError(err)
	}
	if !activeUpload(locked, now) {
		return domain.VideoAsset{}, domain.ErrVideoNotFound
	}
	if locked.State != domain.Uploading || locked.ReceivedBytes != locked.DeclaredBytes {
		return domain.VideoAsset{}, domain.ErrInvalidStateTransition
	}
	if err := policy(locked); err != nil {
		return domain.VideoAsset{}, err
	}
	updated, err := scanAsset(tx.QueryRow(ctx, `UPDATE video_assets AS assets
		SET state='queued', updated_at=$2
		WHERE id=$1::text::uuid AND state='uploading'
		  AND received_bytes=declared_bytes AND expires_at>$2 AND evicted_at IS NULL
		RETURNING `+mediaAssetColumns, id, now))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.VideoAsset{}, domain.ErrInvalidStateTransition
	}
	if err != nil {
		return domain.VideoAsset{}, persistenceError(err)
	}
	if err := commit(ctx, tx); err != nil {
		return domain.VideoAsset{}, err
	}
	return updated, nil
}

func (repository *Repository) Claim(ctx context.Context, id string, token domain.ClaimToken, now time.Time) (domain.VideoAsset, bool, error) {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return domain.VideoAsset{}, false, persistenceError(err)
	}
	defer rollback(tx)
	now = storeTime(now)
	asset, err := scanAsset(tx.QueryRow(ctx, `UPDATE video_assets AS assets
		SET state='processing', claim_token=$2::text::uuid, updated_at=$3
		WHERE id=$1::text::uuid AND state='queued' AND evicted_at IS NULL
		RETURNING `+mediaAssetColumns, id, token, now))
	claimed := true
	if errors.Is(err, pgx.ErrNoRows) {
		claimed = false
		asset, err = scanAsset(tx.QueryRow(ctx, `SELECT `+mediaAssetColumns+`
			FROM video_assets AS assets WHERE id=$1::text::uuid AND evicted_at IS NULL`, id))
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.VideoAsset{}, false, domain.ErrVideoNotFound
	}
	if err != nil {
		return domain.VideoAsset{}, false, persistenceError(err)
	}
	if err := commit(ctx, tx); err != nil {
		return domain.VideoAsset{}, false, err
	}
	return asset, claimed, nil
}

// Requeue returns a claimed asset to the queue, releasing its claim. The guard is the
// claim token, so an attempt whose claim was already reclaimed changes nothing.
func (repository *Repository) Requeue(ctx context.Context, id string, token domain.ClaimToken, now time.Time) error {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return persistenceError(err)
	}
	defer rollback(tx)
	now = storeTime(now)
	if _, err := scanAsset(tx.QueryRow(ctx, lockMediaAsset, id)); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrVideoNotFound
		}
		return persistenceError(err)
	}
	result, err := tx.Exec(ctx, `UPDATE video_assets
		SET state='queued', claim_token=NULL, updated_at=$3
		WHERE id=$1::text::uuid AND state='processing' AND claim_token=$2::text::uuid`,
		id, string(token), now)
	if err != nil {
		return persistenceError(err)
	}
	if result.RowsAffected() != 1 {
		return domain.ErrInvalidStateTransition
	}
	return commit(ctx, tx)
}

// Fail records a terminal failure for a claimed asset. The from state and the claim token
// are both guards, so a reclaimed or already-advanced asset is never overwritten.
func (repository *Repository) Fail(ctx context.Context, id string, from domain.State,
	token domain.ClaimToken, reason domain.FailureReason, now time.Time) error {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return persistenceError(err)
	}
	defer rollback(tx)
	now = storeTime(now)
	if _, err := scanAsset(tx.QueryRow(ctx, lockMediaAsset, id)); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrVideoNotFound
		}
		return persistenceError(err)
	}
	result, err := tx.Exec(ctx, `UPDATE video_assets
		SET state='failed', failure_reason=$4, claim_token=NULL, updated_at=$5
		WHERE id=$1::text::uuid AND state=$2
		  AND (claim_token=$3::text::uuid OR ($3='' AND claim_token IS NULL))`,
		id, string(from), string(token), string(reason), now)
	if err != nil {
		return persistenceError(err)
	}
	if result.RowsAffected() != 1 {
		return domain.ErrInvalidStateTransition
	}
	return commit(ctx, tx)
}

func (repository *Repository) Cancel(ctx context.Context, id string, now time.Time) error {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return persistenceError(err)
	}
	defer rollback(tx)
	now = storeTime(now)
	locked, err := scanAsset(tx.QueryRow(ctx, lockMediaAsset, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrVideoNotFound
	}
	if err != nil {
		return persistenceError(err)
	}
	if !activeUpload(locked, now) {
		return domain.ErrVideoNotFound
	}
	result, err := tx.Exec(ctx, `UPDATE video_assets
		SET evicted_at=$2, updated_at=$2
		WHERE id=$1::text::uuid AND state IN ('created','uploading')
		  AND expires_at>$2 AND evicted_at IS NULL`, id, now)
	if err != nil {
		return persistenceError(err)
	}
	if result.RowsAffected() != 1 {
		return domain.ErrInvalidStateTransition
	}
	return commit(ctx, tx)
}

func (repository *Repository) Retry(ctx context.Context, actor domain.Actor, id string, claimAge time.Duration, now time.Time) (domain.VideoAsset, error) {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return domain.VideoAsset{}, persistenceError(err)
	}
	defer rollback(tx)
	now = storeTime(now)
	locked, ownerID, err := scanOwnedAsset(tx.QueryRow(ctx, lockOwnedMediaAsset, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.VideoAsset{}, domain.ErrVideoNotFound
	}
	if err != nil {
		return domain.VideoAsset{}, persistenceError(err)
	}
	if err := actor.CanEditLesson(ownerID); err != nil {
		return domain.VideoAsset{}, err
	}
	cutoff := storeTime(now.Add(-claimAge))
	retryable := domain.ValidateTransition(domain.T9, locked.State, domain.Queued) == nil &&
		locked.ReceivedBytes == locked.DeclaredBytes &&
		(locked.State != domain.Processing || locked.UpdatedAt.Before(cutoff))
	if !retryable {
		return domain.VideoAsset{}, domain.ErrVideoNotRetryable
	}
	updated, err := scanAsset(tx.QueryRow(ctx, `UPDATE video_assets AS assets
		SET state='queued', claim_token=NULL, attempt=attempt+1, failure_reason=NULL, updated_at=$3
		WHERE id=$1::text::uuid AND evicted_at IS NULL AND received_bytes=declared_bytes
		  AND (state IN ('failed','queued') OR (state='processing' AND updated_at<$2))
		RETURNING `+mediaAssetColumns, id, cutoff, now))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.VideoAsset{}, domain.ErrVideoNotRetryable
	}
	if err != nil {
		return domain.VideoAsset{}, persistenceError(err)
	}
	if err := commit(ctx, tx); err != nil {
		return domain.VideoAsset{}, err
	}
	return updated, nil
}

func activeUpload(asset domain.VideoAsset, now time.Time) bool {
	return asset.EvictedAt == nil && asset.ExpiresAt.After(now) && (asset.State == domain.Created || asset.State == domain.Uploading)
}

type storedAsset struct {
	id, lessonID, state, claimToken, originalName, failureReason string
	generation, attempt                                          int32
	declaredBytes, receivedBytes                                 int64
	renditions                                                   []string
	expiresAt, evictedAt, createdAt, updatedAt                   pgtype.Timestamptz
}

func (stored *storedAsset) scanTargets() []any {
	return []any{
		&stored.id, &stored.lessonID, &stored.generation, &stored.state, &stored.claimToken,
		&stored.declaredBytes, &stored.receivedBytes, &stored.originalName, &stored.attempt,
		&stored.failureReason, &stored.renditions, &stored.expiresAt, &stored.evictedAt,
		&stored.createdAt, &stored.updatedAt,
	}
}

func (stored storedAsset) mapDomain() (domain.VideoAsset, error) {
	return mapAsset(
		stored.id, stored.lessonID, stored.generation, stored.state, stored.claimToken,
		stored.declaredBytes, stored.receivedBytes, stored.originalName, stored.attempt,
		stored.failureReason, stored.renditions, stored.expiresAt, stored.evictedAt,
		stored.createdAt, stored.updatedAt,
	)
}

func scanAsset(row pgx.Row) (domain.VideoAsset, error) {
	var stored storedAsset
	if err := row.Scan(stored.scanTargets()...); err != nil {
		return domain.VideoAsset{}, err
	}
	return stored.mapDomain()
}

func scanOwnedAsset(row pgx.Row) (domain.VideoAsset, string, error) {
	var stored storedAsset
	var ownerID string
	targets := append(stored.scanTargets(), &ownerID)
	if err := row.Scan(targets...); err != nil {
		return domain.VideoAsset{}, "", err
	}
	asset, err := stored.mapDomain()
	return asset, ownerID, err
}

func mapAsset(
	id, lessonID string,
	generation int32,
	state, claimToken string,
	declaredBytes, receivedBytes int64,
	originalName string,
	attempt int32,
	failureReason string,
	storedRenditions []string,
	expiresAt, evictedAt, createdAt, updatedAt pgtype.Timestamptz,
) (domain.VideoAsset, error) {
	if !expiresAt.Valid || !createdAt.Valid || !updatedAt.Valid || generation < 1 || attempt < 0 ||
		receivedBytes < 0 || receivedBytes > declaredBytes || updatedAt.Time.Before(createdAt.Time) {
		return domain.VideoAsset{}, domain.ErrDependencyUnavailable
	}
	if err := domain.ValidateTicket(declaredBytes, originalName); err != nil {
		return domain.VideoAsset{}, domain.ErrDependencyUnavailable
	}
	assetState := domain.State(state)
	switch assetState {
	case domain.Created, domain.Uploading, domain.Queued, domain.Processing, domain.Ready, domain.Failed:
	default:
		return domain.VideoAsset{}, domain.ErrDependencyUnavailable
	}
	reason := domain.FailureReason(failureReason)
	if reason != "" {
		switch reason {
		case domain.UnsupportedSource, domain.ProbeFailed, domain.RenderFailed, domain.PublishFailed, domain.StorageUnavailable:
		default:
			return domain.VideoAsset{}, domain.ErrDependencyUnavailable
		}
	}
	if assetState == domain.Failed && reason == "" || claimToken != "" && assetState != domain.Processing {
		return domain.VideoAsset{}, domain.ErrDependencyUnavailable
	}
	renditions := make([]domain.Rendition, 0, len(storedRenditions))
	for _, stored := range storedRenditions {
		rendition := domain.Rendition(stored)
		switch rendition {
		case domain.Rendition360p, domain.Rendition720p, domain.Rendition1080p:
			renditions = append(renditions, rendition)
		default:
			return domain.VideoAsset{}, domain.ErrDependencyUnavailable
		}
	}
	if assetState == domain.Ready && len(renditions) == 0 {
		return domain.VideoAsset{}, domain.ErrDependencyUnavailable
	}
	var evicted *time.Time
	if evictedAt.Valid {
		value := evictedAt.Time
		evicted = &value
	}
	return domain.VideoAsset{
		ID: id, LessonID: lessonID, Generation: domain.Generation(generation), State: assetState,
		ClaimToken: domain.ClaimToken(claimToken), DeclaredBytes: declaredBytes, ReceivedBytes: receivedBytes,
		OriginalName: originalName, Attempt: int(attempt), FailureReason: reason, Renditions: renditions,
		ExpiresAt: expiresAt.Time, EvictedAt: evicted, CreatedAt: createdAt.Time, UpdatedAt: updatedAt.Time,
	}, nil
}

func uploadInProgress(err error) bool {
	var postgresError *pgconn.PgError
	return errors.As(err, &postgresError) && postgresError.Code == "23505" &&
		(postgresError.ConstraintName == "video_assets_pending_per_lesson" || postgresError.ConstraintName == "video_assets_generation")
}

func persistenceError(err error) error {
	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) {
		for _, class := range []string{"08", "53", "57", "58", "40", "55"} {
			if strings.HasPrefix(postgresError.Code, class) {
				return domain.ErrDependencyUnavailable
			}
		}
		return fmt.Errorf("media persistence failure: %w", err)
	}
	return domain.ErrDependencyUnavailable
}

func storeTime(value time.Time) time.Time { return value.UTC().Truncate(time.Microsecond) }

func commit(ctx context.Context, tx pgx.Tx) error {
	if err := tx.Commit(ctx); err != nil {
		return persistenceError(err)
	}
	return nil
}

func rollback(tx pgx.Tx) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_ = tx.Rollback(ctx)
}
