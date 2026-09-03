package domain

type Error string

func (e Error) Error() string { return string(e) }

const (
	ErrInvalidUser            Error = "invalid_user"
	ErrInvalidEmail           Error = "invalid_email"
	ErrInvalidSession         Error = "invalid_session"
	ErrInvalidCredentials     Error = "invalid_credentials"
	ErrAuthenticationRequired Error = "authentication_required"
	ErrForbidden              Error = "forbidden"
	ErrDependencyUnavailable  Error = "dependency_unavailable"
)
