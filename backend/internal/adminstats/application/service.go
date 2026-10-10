// Package application is the admin-stats boundary: it resolves the timeseries
// window and metric through the domain, stamps the platform flag onto the
// snapshot, and delegates reads to the repository. The admin-only gate lives in
// the RequireAdministrator transport middleware (the single enforcement point),
// so the service trusts the callers the composition root wires behind it.
package application

import (
	"context"
	"time"

	"courses/internal/adminstats/domain"
	"courses/internal/adminstats/ports"
)

// Service reads admin aggregates and validates timeseries input.
type Service struct {
	repository       ports.StatsRepository
	singleInstructor bool
	now              func() time.Time
}

// New builds the service over a stats repository, the single-instructor flag
// and a clock (injected so range resolution is deterministic in tests).
func New(repository ports.StatsRepository, singleInstructor bool, now func() time.Time) *Service {
	return &Service{repository: repository, singleInstructor: singleInstructor, now: now}
}

// Snapshot reads the platform aggregate and stamps the platform flag the UI
// uses to hide the instructor dimension.
func (service *Service) Snapshot(ctx context.Context) (domain.Snapshot, error) {
	snapshot, err := service.repository.Snapshot(ctx)
	if err != nil {
		return domain.Snapshot{}, err
	}
	snapshot.Platform.SingleInstructor = service.singleInstructor
	return snapshot, nil
}

// Timeseries validates the metric, bucket and window as explicit business
// errors (never a silent empty result or a 500) before the bucketed read. The
// metric and bucket are checked before the window so an unknown series is
// rejected as such rather than masked by a range error.
func (service *Service) Timeseries(ctx context.Context, metricRaw, bucketRaw, fromRaw, toRaw string) ([]domain.TimeseriesPoint, error) {
	metric, err := domain.ParseMetric(metricRaw)
	if err != nil {
		return nil, err
	}
	bucket, err := domain.ParseBucket(bucketRaw)
	if err != nil {
		return nil, err
	}
	from, to, err := domain.ResolveRange(fromRaw, toRaw, bucket, service.now())
	if err != nil {
		return nil, err
	}
	return service.repository.Timeseries(ctx, metric, bucket, from, to)
}
