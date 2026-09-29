package domain

import "time"

type PaymentEventType string

const PaymentSucceeded PaymentEventType = "paid"

type PaymentEvent struct {
	Provider   string
	EventID    string
	OrderID    string
	Type       PaymentEventType
	Amount     Money
	ReceivedAt time.Time
}

type PaymentEventAction string

const (
	PaymentEventApply  PaymentEventAction = "apply"
	PaymentEventNoOp   PaymentEventAction = "no_op"
	PaymentEventRetry  PaymentEventAction = "retry_unmatched"
	PaymentEventReject PaymentEventAction = "reject"
)

type PaymentEventDecision struct {
	Action        PaymentEventAction
	Order         Order
	IssueGrant    bool
	GrantSource   GrantSource
	GrantOrderID  string
	GrantUserID   string
	GrantCourseID string
}

// DecidePaymentEvent is persistence-free. The application owns the inbox transaction and tells
// the domain whether this provider event was already applied.
func DecidePaymentEvent(event PaymentEvent, order *Order, alreadyApplied bool) (PaymentEventDecision, error) {
	if event.Provider == "" || event.EventID == "" || event.OrderID == "" || event.ReceivedAt.IsZero() ||
		event.Type != PaymentSucceeded || event.Amount.MinorUnits() <= 0 || event.Amount.Currency() != CurrencyUAH {
		return rejectedPaymentEvent()
	}
	if alreadyApplied {
		return PaymentEventDecision{Action: PaymentEventNoOp}, nil
	}
	if order == nil {
		return PaymentEventDecision{Action: PaymentEventRetry}, nil
	}
	if order.ID() != event.OrderID || order.Status() != OrderPending || !order.Amount().Equal(event.Amount) {
		return rejectedPaymentEvent()
	}
	paid, err := order.MarkPaid(event.ReceivedAt)
	if err != nil {
		return rejectedPaymentEvent()
	}
	return PaymentEventDecision{
		Action:        PaymentEventApply,
		Order:         paid,
		IssueGrant:    true,
		GrantSource:   GrantPurchase,
		GrantOrderID:  order.ID(),
		GrantUserID:   order.UserID(),
		GrantCourseID: order.CourseID(),
	}, nil
}

func rejectedPaymentEvent() (PaymentEventDecision, error) {
	return PaymentEventDecision{Action: PaymentEventReject}, ErrPaymentEventRejected
}
