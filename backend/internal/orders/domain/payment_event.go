package domain

import "time"

type PaymentEventType string

const (
	PaymentSucceeded PaymentEventType = "paid"
	PaymentFailed    PaymentEventType = "failed"
	// PaymentRefunded carries both WayForPay Refunded and Voided — for us a single outcome:
	// money returned, access must go. The adapter maps both statuses onto this type.
	PaymentRefunded PaymentEventType = "refunded"
)

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
	// PaymentEventRevoke: a refund on a paid order — revoke that order's grant (GrantOrderID).
	PaymentEventRevoke PaymentEventAction = "revoke"
	// PaymentEventIgnore: a refund on a non-paid order — benign no-op (warn+ACK), not a reject,
	// so it does not flag refund_required the way a rejected paid event does.
	PaymentEventIgnore PaymentEventAction = "ignore"
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
	// Envelope validation common to every type. Amount is checked per-type below: a settle
	// (paid/failed) needs a positive amount matching the order; a refund (G9 all-or-nothing)
	// never inspects the amount — a partial Refunded or a zero-amount Voided still revokes.
	knownType := event.Type == PaymentSucceeded || event.Type == PaymentFailed || event.Type == PaymentRefunded
	if event.Provider == "" || event.EventID == "" || event.OrderID == "" || event.ReceivedAt.IsZero() ||
		!knownType || event.Amount.Currency() != CurrencyUAH {
		return rejectedPaymentEvent()
	}
	if alreadyApplied {
		return PaymentEventDecision{Action: PaymentEventNoOp}, nil
	}
	if order == nil {
		return PaymentEventDecision{Action: PaymentEventRetry}, nil
	}
	if order.ID() != event.OrderID {
		return rejectedPaymentEvent()
	}
	// Refund branch BEFORE the settle amount-guard (G9): a signed Refunded/Voided revokes the
	// order's grant regardless of amount; on any non-paid order it is a benign ignore (G11),
	// not a reject (which would flag refund_required for a paid event).
	if event.Type == PaymentRefunded {
		if order.Status() != OrderPaid {
			return PaymentEventDecision{Action: PaymentEventIgnore}, nil
		}
		return PaymentEventDecision{Action: PaymentEventRevoke, GrantOrderID: order.ID()}, nil
	}
	if order.Status() != OrderPending || event.Amount.MinorUnits() <= 0 || !order.Amount().Equal(event.Amount) {
		return rejectedPaymentEvent()
	}
	if event.Type == PaymentFailed {
		failed, err := order.MarkFailed()
		if err != nil {
			return rejectedPaymentEvent()
		}
		return PaymentEventDecision{Action: PaymentEventApply, Order: failed}, nil
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
