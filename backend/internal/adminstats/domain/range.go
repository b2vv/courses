package domain

import "time"

// MaxBuckets caps a timeseries window so a day bucket spans at most ~1 year and
// a week bucket at most ~7 years, keeping the aggregate query cheap.
const MaxBuckets = 366

// DefaultWindow is the span used when a bound is omitted.
const DefaultWindow = 30 * 24 * time.Hour

// ResolveRange validates and normalises a timeseries window to UTC.
//
// Each bound defaults independently: an empty `to` becomes now; an empty `from`
// becomes `to` minus the default window. A non-empty bound that does not parse
// as RFC3339 is an invalid_range error, checked before any semantic rule. The
// bucket count floor((to-from)/bucket)+1 must not exceed MaxBuckets.
func ResolveRange(fromRaw, toRaw string, bucket Bucket, now time.Time) (time.Time, time.Time, error) {
	to := now.UTC()
	if toRaw != "" {
		parsed, err := time.Parse(time.RFC3339, toRaw)
		if err != nil {
			return time.Time{}, time.Time{}, ErrInvalidRange
		}
		to = parsed.UTC()
	}

	from := to.Add(-DefaultWindow)
	if fromRaw != "" {
		parsed, err := time.Parse(time.RFC3339, fromRaw)
		if err != nil {
			return time.Time{}, time.Time{}, ErrInvalidRange
		}
		from = parsed.UTC()
	}

	if from.After(to) {
		return time.Time{}, time.Time{}, ErrInvalidRange
	}

	if BucketCount(from, to, bucket) > MaxBuckets {
		return time.Time{}, time.Time{}, ErrRangeTooWide
	}

	return from, to, nil
}

// BucketCount is floor((to-from)/bucket)+1, computed over fixed UTC durations.
func BucketCount(from, to time.Time, bucket Bucket) int {
	return int(to.Sub(from)/bucket.Duration()) + 1
}
