package domain

// Level is the severity of a captured browser event. The set is closed: an
// unknown level is a malformed event, not a silently-accepted one.
type Level string

const (
	LevelError Level = "error"
	LevelWarn  Level = "warn"
	LevelInfo  Level = "info"
)

// unknownSentinel fills an empty route or release so the earliest boot error —
// captured before the app knows its route, or on a build without a release id —
// is still emitted rather than dropped on an empty required field.
const unknownSentinel = "unknown"

// Field size caps. The HTTP body cap bounds the whole request; these bound each
// field so one oversized value cannot blow up a single log line. Oversized free
// text is truncated (the event is still worth keeping), never rejected.
const (
	maxMessageLength    = 2000
	maxRouteLength      = 512
	maxReleaseLength    = 128
	maxUserAgentLength  = 256
	maxErrorNameLength  = 256
	maxErrorStackLength = 8000
)

// Input is the raw, untrusted event a browser posted. Every field is a string
// so a non-Error rejection (`reject("x")`, `reject(undefined)`) still produces
// a well-formed input on the client side.
type Input struct {
	Level      string
	Message    string
	Route      string
	Release    string
	UserAgent  string
	ErrorName  string
	ErrorStack string
}

// ClientEvent is a validated, redacted, size-bounded browser event ready to
// emit as a structured log line. Constructing one is the only way to get one,
// so an emitted event is always safe.
type ClientEvent struct {
	Level      Level
	Message    string
	Route      string
	Release    string
	UserAgent  string
	ErrorName  string
	ErrorStack string
}

// NewClientEvent validates the level, applies the §17 value-scrub to every
// free-text field, truncates each field to its cap, and fills the route/release
// sentinels. It returns ErrInvalidEvent only for a structurally-malformed
// event (an unknown level); everything else is coerced into a safe line rather
// than dropped.
func NewClientEvent(in Input) (ClientEvent, error) {
	level, err := normalizeLevel(in.Level)
	if err != nil {
		return ClientEvent{}, err
	}
	return ClientEvent{
		Level:      level,
		Message:    truncate(scrubSecrets(in.Message), maxMessageLength),
		Route:      sentinel(truncate(scrubSecrets(in.Route), maxRouteLength)),
		Release:    sentinel(truncate(scrubSecrets(in.Release), maxReleaseLength)),
		UserAgent:  truncate(scrubSecrets(in.UserAgent), maxUserAgentLength),
		ErrorName:  truncate(scrubSecrets(in.ErrorName), maxErrorNameLength),
		ErrorStack: truncate(scrubSecrets(in.ErrorStack), maxErrorStackLength),
	}, nil
}

// normalizeLevel defaults an empty level to error (the common client capture)
// and rejects any non-empty value outside the closed set.
func normalizeLevel(raw string) (Level, error) {
	switch Level(raw) {
	case "":
		return LevelError, nil
	case LevelError, LevelWarn, LevelInfo:
		return Level(raw), nil
	default:
		return "", ErrInvalidEvent
	}
}

func sentinel(value string) string {
	if value == "" {
		return unknownSentinel
	}
	return value
}

// truncate cuts a string to at most max runes, preserving valid UTF-8.
func truncate(value string, max int) string {
	if len(value) <= max {
		return value
	}
	runes := []rune(value)
	if len(runes) <= max {
		return value
	}
	return string(runes[:max])
}
