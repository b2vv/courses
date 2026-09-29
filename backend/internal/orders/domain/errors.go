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
)
