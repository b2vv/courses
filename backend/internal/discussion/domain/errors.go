package domain

import "errors"

var (
	ErrNotFound              = errors.New("not_found")
	ErrDependencyUnavailable = errors.New("dependency_unavailable")
)
