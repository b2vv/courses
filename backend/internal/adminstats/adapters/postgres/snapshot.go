// Package postgres is the admin-stats read adapter: raw read-only SQL over the
// shared pool, aggregating the six platform domains into a domain.Snapshot. It
// follows the orders/discussion precedent of cross-table reads in one module's
// own adapter rather than importing another module (AGENTS.md).
package postgres

import (
	"context"
	"errors"

	"courses/internal/adminstats/domain"
	"courses/internal/adminstats/ports"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository reads admin aggregates from the shared Postgres pool.
type Repository struct{ pool *pgxpool.Pool }

// Repository implements the stats read port.
var _ ports.StatsRepository = (*Repository)(nil)

// NewRepository builds the read adapter over the shared pool.
func NewRepository(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

// Snapshot aggregates the six platform domains in one point-in-time read.
// Every status/role/state count uses COUNT(*) FILTER per key so a key with zero
// rows reports 0 rather than a missing field (N1): an empty platform yields a
// fully-populated snapshot of zeros. Platform flags are set by the service, not
// here; Instructors is filled by the leaderboard query (instructors.go).
func (repository *Repository) Snapshot(ctx context.Context) (domain.Snapshot, error) {
	var snapshot domain.Snapshot
	if err := repository.scanUsers(ctx, &snapshot.Users); err != nil {
		return domain.Snapshot{}, err
	}
	if err := repository.scanCourses(ctx, &snapshot.Courses); err != nil {
		return domain.Snapshot{}, err
	}
	if err := repository.scanGrants(ctx, &snapshot.Grants); err != nil {
		return domain.Snapshot{}, err
	}
	if err := repository.scanOrders(ctx, &snapshot.Orders); err != nil {
		return domain.Snapshot{}, err
	}
	if err := repository.scanReports(ctx, &snapshot.Reports); err != nil {
		return domain.Snapshot{}, err
	}
	if err := repository.scanMedia(ctx, &snapshot.Media); err != nil {
		return domain.Snapshot{}, err
	}
	instructors, err := repository.loadInstructors(ctx)
	if err != nil {
		return domain.Snapshot{}, err
	}
	snapshot.Instructors = instructors
	return snapshot, nil
}

// scanUsers counts accounts by lifecycle, role and security attributes. Status
// derives from deleted_at/disabled_at (identity); role counts exclude deleted
// accounts so "total accounts" = active+blocked downstream.
func (repository *Repository) scanUsers(ctx context.Context, into *domain.UsersStat) error {
	err := repository.pool.QueryRow(ctx, `
SELECT
  COUNT(*) FILTER (WHERE deleted_at IS NULL AND disabled_at IS NULL) AS active,
  COUNT(*) FILTER (WHERE deleted_at IS NULL AND disabled_at IS NOT NULL) AS blocked,
  COUNT(*) FILTER (WHERE deleted_at IS NOT NULL) AS deleted,
  COUNT(*) FILTER (WHERE deleted_at IS NULL AND role = 'student') AS students,
  COUNT(*) FILTER (WHERE deleted_at IS NULL AND role = 'instructor') AS instructors,
  COUNT(*) FILTER (WHERE deleted_at IS NULL AND role = 'admin') AS admins,
  COUNT(*) FILTER (WHERE deleted_at IS NULL AND email_verified) AS verified,
  COUNT(*) FILTER (WHERE deleted_at IS NULL AND totp_enabled) AS two_factor
FROM users`).Scan(
		&into.Active, &into.Blocked, &into.Deleted,
		&into.ByRole.Student, &into.ByRole.Instructor, &into.ByRole.Admin,
		&into.Verified, &into.TwoFactor,
	)
	return persistenceError(err)
}

// scanCourses counts courses by lifecycle status.
func (repository *Repository) scanCourses(ctx context.Context, into *domain.CoursesStat) error {
	err := repository.pool.QueryRow(ctx, `
SELECT
  COUNT(*) FILTER (WHERE status = 'draft') AS draft,
  COUNT(*) FILTER (WHERE status = 'published') AS published,
  COUNT(*) FILTER (WHERE status = 'archived') AS archived
FROM courses`).Scan(&into.Draft, &into.Published, &into.Archived)
	return persistenceError(err)
}

// scanGrants counts course memberships by state (post-00015 three-state).
func (repository *Repository) scanGrants(ctx context.Context, into *domain.GrantsStat) error {
	err := repository.pool.QueryRow(ctx, `
SELECT
  COUNT(*) FILTER (WHERE state = 'active') AS active,
  COUNT(*) FILTER (WHERE state = 'blocked') AS blocked,
  COUNT(*) FILTER (WHERE state = 'revoked') AS revoked
FROM course_grants`).Scan(&into.Active, &into.Blocked, &into.Revoked)
	return persistenceError(err)
}

// scanOrders counts orders by status and sums paid revenue. SUM over bigint is
// numeric in Postgres, so the paid-filtered sum is cast back to bigint.
func (repository *Repository) scanOrders(ctx context.Context, into *domain.OrdersStat) error {
	err := repository.pool.QueryRow(ctx, `
SELECT
  COUNT(*) FILTER (WHERE status = 'pending') AS pending,
  COUNT(*) FILTER (WHERE status = 'paid') AS paid,
  COUNT(*) FILTER (WHERE status = 'failed') AS failed,
  COALESCE(SUM(amount_minor) FILTER (WHERE status = 'paid'), 0)::bigint AS revenue
FROM orders`).Scan(&into.Pending, &into.Paid, &into.Failed, &into.RevenueMinorUAH)
	return persistenceError(err)
}

// scanReports counts participant reports by moderation state.
func (repository *Repository) scanReports(ctx context.Context, into *domain.ReportsStat) error {
	err := repository.pool.QueryRow(ctx, `
SELECT
  COUNT(*) FILTER (WHERE state = 'open') AS open,
  COUNT(*) FILTER (WHERE state = 'resolved') AS resolved
FROM participant_reports`).Scan(&into.Open, &into.Resolved)
	return persistenceError(err)
}

// scanMedia maps the six-state video_assets lifecycle onto three coarse buckets
// and aggregates duration honestly: the average is over assets that carry a
// duration only (AVG over zero such rows is NULL, left nil rather than a
// misleading 0), with WithDurationCount exposing the coverage.
func (repository *Repository) scanMedia(ctx context.Context, into *domain.MediaStat) error {
	err := repository.pool.QueryRow(ctx, `
SELECT
  COUNT(*) FILTER (WHERE state IN ('created','uploading','queued','processing')) AS in_progress,
  COUNT(*) FILTER (WHERE state = 'ready') AS ready,
  COUNT(*) FILTER (WHERE state = 'failed') AS failed,
  COALESCE(SUM(duration_seconds), 0) AS total_duration,
  AVG(duration_seconds) FILTER (WHERE duration_seconds IS NOT NULL) AS avg_duration,
  COUNT(*) FILTER (WHERE duration_seconds IS NOT NULL) AS with_duration
FROM video_assets`).Scan(
		&into.InProgress, &into.Ready, &into.Failed,
		&into.TotalDurationSeconds, &into.AvgDurationSeconds, &into.WithDurationCount,
	)
	return persistenceError(err)
}

// persistenceError collapses a read failure to the domain's dependency error,
// leaving context cancellation untouched (orders/discussion precedent).
func persistenceError(err error) error {
	if err == nil || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	return domain.ErrDependencyUnavailable
}
