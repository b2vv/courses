package domain

type Error string

func (e Error) Error() string { return string(e) }

const (
	ErrForbidden              Error = "forbidden"
	ErrDependencyUnavailable  Error = "dependency_unavailable"
	ErrInvalidProgress        Error = "invalid_progress"
	ErrLessonNotFound         Error = "lesson_not_found"
	ErrPlaybackTakenOver      Error = "playback_taken_over"
	ErrPlaybackSessionExpired Error = "playback_session_expired"
)
