package domain_test

import (
	"errors"
	"testing"

	"courses/internal/adminstats/domain"
)

func TestParseMetric(t *testing.T) {
	for _, m := range []domain.Metric{domain.MetricSignups, domain.MetricEnrollments, domain.MetricOrders, domain.MetricRevenue} {
		got, err := domain.ParseMetric(string(m))
		if err != nil || got != m {
			t.Fatalf("ParseMetric(%q) = %q, %v; want %q, nil", m, got, err, m)
		}
	}
	if _, err := domain.ParseMetric("bogus"); !errors.Is(err, domain.ErrUnknownMetric) {
		t.Fatalf("unknown metric: got %v, want ErrUnknownMetric", err)
	}
	if _, err := domain.ParseMetric(""); !errors.Is(err, domain.ErrUnknownMetric) {
		t.Fatalf("empty metric: got %v, want ErrUnknownMetric", err)
	}
}

func TestParseBucket(t *testing.T) {
	for _, b := range []domain.Bucket{domain.BucketDay, domain.BucketWeek} {
		got, err := domain.ParseBucket(string(b))
		if err != nil || got != b {
			t.Fatalf("ParseBucket(%q) = %q, %v; want %q, nil", b, got, err, b)
		}
	}
	if _, err := domain.ParseBucket("month"); !errors.Is(err, domain.ErrUnknownBucket) {
		t.Fatalf("unknown bucket: got %v, want ErrUnknownBucket", err)
	}
}

func TestBucketDuration(t *testing.T) {
	if domain.BucketDay.Duration() != 24*60*60*1e9 {
		t.Fatalf("day duration wrong: %v", domain.BucketDay.Duration())
	}
	if domain.BucketWeek.Duration() != 7*domain.BucketDay.Duration() {
		t.Fatalf("week duration wrong: %v", domain.BucketWeek.Duration())
	}
}
