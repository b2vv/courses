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
	ErrCannotTargetSelf       Error = "cannot_target_self"
	ErrAdminImmutable         Error = "admin_immutable"
	ErrLastAdmin              Error = "last_admin"
	ErrUserNotFound           Error = "user_not_found"
)
