package postgres

import (
	"context"
	"time"

	"courses/internal/adminstats/domain"
)

// Timeseries buckets one metric over the [from, to) window. The window is
// already validated and normalised to UTC by the caller; the adapter trusts it
// and truncates each row's timestamp in UTC so buckets do not drift with the
// server's local zone or DST (N3). An empty window yields a non-nil empty slice
// (rendered as [] rather than null), so the chart shows an empty state instead
// of crashing (scenario 6).
func (repository *Repository) Timeseries(ctx context.Context, metric domain.Metric, bucket domain.Bucket, from, to time.Time) ([]domain.TimeseriesPoint, error) {
	query, err := timeseriesQuery(metric)
	if err != nil {
		return nil, err
	}
	rows, err := repository.pool.Query(ctx, query, string(bucket), from, to)
	if err != nil {
		return nil, persistenceError(err)
	}
	defer rows.Close()

	points := make([]domain.TimeseriesPoint, 0)
	for rows.Next() {
		var point domain.TimeseriesPoint
		if err := rows.Scan(&point.BucketStart, &point.Value); err != nil {
			return nil, persistenceError(err)
		}
		points = append(points, point)
	}
	if err := rows.Err(); err != nil {
		return nil, persistenceError(err)
	}
	return points, nil
}

// timeseriesQuery returns the bucketed SQL for a metric. Each query truncates
// its own timestamp column ($1 = bucket unit, 'day'|'week') and filters the
// half-open window ($2 = from inclusive, $3 = to exclusive). Identifiers are
// fixed per metric (never caller input); only the bucket unit and the window
// are parameters. revenue sums paid order amounts by paid_at, not created_at
// (scenario 4), as double precision for a uniform float value.
func timeseriesQuery(metric domain.Metric) (string, error) {
	switch metric {
	case domain.MetricSignups:
		return bucketedQuery("users", "created_at", "COUNT(*)::double precision", ""), nil
	case domain.MetricEnrollments:
		return bucketedQuery("course_grants", "granted_at", "COUNT(*)::double precision", ""), nil
	case domain.MetricOrders:
		return bucketedQuery("orders", "created_at", "COUNT(*)::double precision", ""), nil
	case domain.MetricRevenue:
		return bucketedQuery("orders", "paid_at", "COALESCE(SUM(amount_minor), 0)::double precision", " AND status = 'paid'"), nil
	default:
		return "", domain.ErrUnknownMetric
	}
}

func bucketedQuery(table, timeColumn, valueExpr, extraWhere string) string {
	return `
SELECT date_trunc($1, ` + timeColumn + ` AT TIME ZONE 'UTC') AS bucket_start,
       ` + valueExpr + ` AS value
FROM ` + table + `
WHERE ` + timeColumn + ` >= $2 AND ` + timeColumn + ` < $3` + extraWhere + `
GROUP BY bucket_start
ORDER BY bucket_start ASC`
}
