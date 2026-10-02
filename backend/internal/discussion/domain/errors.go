package domain

import "errors"

var (
	ErrNotFound              = errors.New("not_found")
	ErrDependencyUnavailable = errors.New("dependency_unavailable")
	ErrForbidden             = errors.New("course_forbidden")
	ErrPurchaseRequired      = errors.New("course_purchase_required")
	ErrInvalidRequest        = errors.New("invalid_request")
	ErrReportOwnComment      = errors.New("report_own_comment")
)
