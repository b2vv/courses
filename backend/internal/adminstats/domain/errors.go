package domain

// Error is a stable business-error code the adapters map to an HTTP status.
type Error string

func (e Error) Error() string { return string(e) }

const (
	ErrInvalidRange  Error = "invalid_range"
	ErrRangeTooWide  Error = "range_too_wide"
	ErrUnknownMetric Error = "unknown_metric"
	ErrUnknownBucket Error = "unknown_bucket"
	// ErrDependencyUnavailable is a read-path failure the adapter cannot
	// attribute to the caller's input (the database is unreachable or errored).
	ErrDependencyUnavailable Error = "dependency_unavailable"
)
