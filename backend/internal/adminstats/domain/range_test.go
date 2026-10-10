package domain_test

import (
	"errors"
	"testing"
	"time"

	"courses/internal/adminstats/domain"
)

func rfc(t time.Time) string { return t.UTC().Format(time.RFC3339) }

func TestResolveRangeCapBoundary(t *testing.T) {
	now := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	// span 365 days = 366 day-buckets → passes (== MaxBuckets).
	to := base.Add(365 * 24 * time.Hour)
	if _, _, err := domain.ResolveRange(rfc(base), rfc(to), domain.BucketDay, now); err != nil {
		t.Fatalf("366 buckets should pass, got %v", err)
	}

	// span 366 days = 367 day-buckets → range_too_wide.
	to = base.Add(366 * 24 * time.Hour)
	if _, _, err := domain.ResolveRange(rfc(base), rfc(to), domain.BucketDay, now); !errors.Is(err, domain.ErrRangeTooWide) {
		t.Fatalf("367 buckets should be range_too_wide, got %v", err)
	}
}

func TestResolveRangeFromAfterTo(t *testing.T) {
	now := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	from := "2026-05-01T00:00:00Z"
	to := "2026-04-01T00:00:00Z"
	if _, _, err := domain.ResolveRange(from, to, domain.BucketDay, now); !errors.Is(err, domain.ErrInvalidRange) {
		t.Fatalf("from>to should be invalid_range, got %v", err)
	}
}

func TestResolveRangeParseFailure(t *testing.T) {
	now := time.Now()
	if _, _, err := domain.ResolveRange("banana", "", domain.BucketDay, now); !errors.Is(err, domain.ErrInvalidRange) {
		t.Fatalf("unparseable from should be invalid_range, got %v", err)
	}
	if _, _, err := domain.ResolveRange("", "not-a-date", domain.BucketDay, now); !errors.Is(err, domain.ErrInvalidRange) {
		t.Fatalf("unparseable to should be invalid_range, got %v", err)
	}
}

func TestResolveRangeIndependentDefaults(t *testing.T) {
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)

	// only `from` → `to` defaults to now.
	from := time.Date(2026, 5, 20, 0, 0, 0, 0, time.UTC)
	gotFrom, gotTo, err := domain.ResolveRange(rfc(from), "", domain.BucketDay, now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !gotTo.Equal(now) {
		t.Fatalf("to should default to now, got %v", gotTo)
	}
	if !gotFrom.Equal(from) {
		t.Fatalf("from should be preserved, got %v", gotFrom)
	}

	// only `to` → `from` defaults to to-30d.
	to := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)
	gotFrom, gotTo, err = domain.ResolveRange("", rfc(to), domain.BucketDay, now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !gotTo.Equal(to) {
		t.Fatalf("to should be preserved, got %v", gotTo)
	}
	if want := to.Add(-domain.DefaultWindow); !gotFrom.Equal(want) {
		t.Fatalf("from should default to to-30d (%v), got %v", want, gotFrom)
	}

	// both empty → [now-30d, now].
	gotFrom, gotTo, err = domain.ResolveRange("", "", domain.BucketDay, now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !gotTo.Equal(now) || !gotFrom.Equal(now.Add(-domain.DefaultWindow)) {
		t.Fatalf("both empty → [now-30d, now], got [%v, %v]", gotFrom, gotTo)
	}
}
