package domain

import (
	"slices"
	"time"
)

// Metric is a supported timeseries series.
type Metric string

const (
	MetricSignups     Metric = "signups"
	MetricEnrollments Metric = "enrollments"
	MetricOrders      Metric = "orders"
	MetricRevenue     Metric = "revenue"
)

// Metrics is the ordered set accepted by ParseMetric and the HTTP contract.
var Metrics = []Metric{MetricSignups, MetricEnrollments, MetricOrders, MetricRevenue}

// ParseMetric validates a raw metric name.
func ParseMetric(raw string) (Metric, error) {
	if slices.Contains(Metrics, Metric(raw)) {
		return Metric(raw), nil
	}
	return "", ErrUnknownMetric
}

// Bucket is a timeseries bucket width.
type Bucket string

const (
	BucketDay  Bucket = "day"
	BucketWeek Bucket = "week"
)

// Buckets is the ordered set accepted by ParseBucket and the HTTP contract.
var Buckets = []Bucket{BucketDay, BucketWeek}

// ParseBucket validates a raw bucket name.
func ParseBucket(raw string) (Bucket, error) {
	if slices.Contains(Buckets, Bucket(raw)) {
		return Bucket(raw), nil
	}
	return "", ErrUnknownBucket
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
