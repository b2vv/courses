package domain

type Error string

func (e Error) Error() string { return string(e) }

const (
	ErrMoneyInvalid           Error = "money_invalid"
	ErrOrderInvalid           Error = "order_invalid"
	ErrOrderTransitionInvalid Error = "order_status_transition_invalid"
	ErrGrantInvalid           Error = "grant_invalid"
	ErrGrantTransitionInvalid Error = "grant_transition_invalid"
	ErrMembershipConflict     Error = "membership_conflict"
	ErrCourseAlreadyOwned     Error = "course_already_owned"
	ErrPaymentEventRejected   Error = "payment_event_rejected"
	ErrPaymentEventUnmatched  Error = "payment_event_unmatched"
	ErrCourseNotPurchasable   Error = "course_not_purchasable"
	ErrCourseNotEnrollable    Error = "course_not_enrollable"
	ErrEnrollmentForbidden    Error = "enrollment_forbidden"
	ErrOrderNotFound          Error = "order_not_found"
	ErrCheckoutAlreadyOpened  Error = "checkout_already_opened"
	ErrGrantForbidden         Error = "grant_forbidden"
	ErrGrantSelfForbidden     Error = "grant_self_forbidden"
	ErrAdminNotLearner        Error = "admin_not_learner"
	ErrDependencyUnavailable  Error = "dependency_unavailable"
	ErrCourseModified         Error = "course_modified"
	ErrEmailNotVerified       Error = "email_not_verified"
	ErrInvalidCursor          Error = "roster_cursor_invalid"
	ErrLearnerCourseForbidden Error = "learner_course_forbidden"
	ErrCommentNotFound        Error = "comment_not_found"
	ErrReviewNotFound         Error = "review_not_found"
)
