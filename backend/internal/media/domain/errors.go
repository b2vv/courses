package domain

type Error string

func (e Error) Error() string { return string(e) }

const (
	ErrLessonNotFound         Error = "lesson_not_found"
	ErrUploadInProgress       Error = "upload_in_progress"
	ErrUnsupportedMedia       Error = "unsupported_media"
	ErrUploadTooLarge         Error = "upload_too_large"
	ErrVideoNotFound          Error = "video_not_found"
	ErrVideoNotRetryable      Error = "video_not_retryable"
	ErrDuplicateLanguage      Error = "duplicate_language"
	ErrAuthenticationRequired Error = "authentication_required"
	ErrForbidden              Error = "forbidden"
	ErrCourseForbidden        Error = "course_forbidden"
	ErrOriginForbidden        Error = "origin_forbidden"
	ErrInvalidRequest         Error = "invalid_request"
	ErrDependencyUnavailable  Error = "dependency_unavailable"
	ErrInternal               Error = "internal_error"

	ErrInvalidStateTransition Error = "invalid_state_transition"
	ErrUnsupportedSource      Error = "unsupported_source"
)
