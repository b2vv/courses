package domain

// Error is a stable business-error code the adapters map to an HTTP status.
type Error string

func (e Error) Error() string { return string(e) }

const (
	// ErrInvalidEvent is any malformed client event the caller sent: an
	// unknown level, a message past the size cap, or a field that cannot be
	// coerced into a safe log line. The adapter answers 400 with the standard
	// envelope and never echoes the offending value back.
	ErrInvalidEvent Error = "invalid_request"
)
