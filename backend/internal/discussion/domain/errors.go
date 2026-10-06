package domain

import "errors"

var (
	ErrNotFound                  = errors.New("not_found")
	ErrDependencyUnavailable     = errors.New("dependency_unavailable")
	ErrForbidden                 = errors.New("course_forbidden")
	ErrPurchaseRequired          = errors.New("course_purchase_required")
	ErrEnrollmentRequired        = errors.New("enrollment_required")
	ErrEnrollmentBlocked         = errors.New("enrollment_blocked")
	ErrInvalidRequest            = errors.New("invalid_request")
	ErrReportOwnComment          = errors.New("report_own_comment")
	ErrReviewOwnerForbidden      = errors.New("review_owner_forbidden")
	ErrReviewNotAvailable        = errors.New("review_not_available")
	ErrReviewPurchaseRequired    = errors.New("review_purchase_required")
	ErrReviewNotStarted          = errors.New("review_not_started")
	ErrReviewInvalid             = errors.New("review_invalid")
	ErrReviewAlreadyExists       = errors.New("review_already_exists")
	ErrReviewRemoved             = errors.New("review_removed")
	ErrReportOwnReview           = errors.New("report_own_review")
	ErrReviewModerationForbidden = errors.New("review_moderation_forbidden")
)
