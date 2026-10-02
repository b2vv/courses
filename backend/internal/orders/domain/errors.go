package domain

type Error string

func (e Error) Error() string { return string(e) }

const (
	ErrMoneyInvalid           Error = "money_invalid"
	ErrOrderInvalid           Error = "order_invalid"
	ErrOrderTransitionInvalid Error = "order_status_transition_invalid"
	ErrGrantInvalid           Error = "grant_invalid"
	ErrCourseAlreadyOwned     Error = "course_already_owned"
	ErrPaymentEventRejected   Error = "payment_event_rejected"
	ErrPaymentEventUnmatched  Error = "payment_event_unmatched"
	ErrCourseNotPurchasable   Error = "course_not_purchasable"
	ErrOrderNotFound          Error = "order_not_found"
	ErrCheckoutAlreadyOpened  Error = "checkout_already_opened"
	ErrGrantForbidden         Error = "grant_forbidden"
	ErrDependencyUnavailable  Error = "dependency_unavailable"
	ErrCourseModified         Error = "course_modified"
)
