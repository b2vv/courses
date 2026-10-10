package ports

import (
	"context"
	"time"

	"courses/internal/adminstats/domain"
)

// StatsRepository reads admin aggregates. The adapter owns the cross-table SQL;
// the window passed to Timeseries is already validated and normalised to UTC.
type StatsRepository interface {
	Snapshot(ctx context.Context) (domain.Snapshot, error)
	Timeseries(ctx context.Context, metric domain.Metric, bucket domain.Bucket, from, to time.Time) ([]domain.TimeseriesPoint, error)
}
