package domain

import "time"

// Metric is a supported timeseries series.
type Metric string

const (
	MetricSignups     Metric = "signups"
	MetricEnrollments Metric = "enrollments"
	MetricOrders      Metric = "orders"
	MetricRevenue     Metric = "revenue"
)

// ParseMetric validates a raw metric name.
func ParseMetric(raw string) (Metric, error) {
	switch Metric(raw) {
	case MetricSignups, MetricEnrollments, MetricOrders, MetricRevenue:
		return Metric(raw), nil
	default:
		return "", ErrUnknownMetric
	}
}

// Bucket is a timeseries bucket width.
type Bucket string

const (
	BucketDay  Bucket = "day"
	BucketWeek Bucket = "week"
)

// ParseBucket validates a raw bucket name.
func ParseBucket(raw string) (Bucket, error) {
	switch Bucket(raw) {
	case BucketDay, BucketWeek:
		return Bucket(raw), nil
	default:
		return "", ErrUnknownBucket
	}
}

// Duration is the fixed width of a bucket, in UTC (no DST drift).
func (b Bucket) Duration() time.Duration {
	if b == BucketWeek {
		return 7 * 24 * time.Hour
	}
	return 24 * time.Hour
}

// TimeseriesPoint is one bucket of a series.
type TimeseriesPoint struct {
	BucketStart time.Time
	Value       float64
}
