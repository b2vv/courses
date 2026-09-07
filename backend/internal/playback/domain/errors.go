package domain

type Error string

func (e Error) Error() string { return string(e) }

const (
	ErrForbidden             Error = "forbidden"
	ErrDependencyUnavailable Error = "dependency_unavailable"
)
