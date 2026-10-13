package domain

type Error string

func (e Error) Error() string { return string(e) }

const (
	ErrUnsupportedImage   Error = "unsupported_image"
	ErrImageTooLarge      Error = "image_too_large"
	ErrImageTooManyPixels Error = "image_too_many_pixels"
	ErrInvalidImage       Error = "invalid_image"
	ErrImageNotFound      Error = "image_not_found"
	ErrForbidden          Error = "forbidden"
	ErrLessonNotFound     Error = "lesson_not_found"
	ErrEmailNotVerified   Error = "email_not_verified"
)
