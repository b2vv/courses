package domain

import (
	"errors"
	"testing"
	"time"
)

func TestDecidePaymentEventAppliesMatchingPendingOrder(t *testing.T) {
	createdAt := time.Date(2026, 9, 30, 9, 0, 0, 0, time.UTC)
	order, err := NewOrder("order-1", "user-1", "course-1", mustMoney(t, "50000"), "checkout-1", createdAt)
	if err != nil {
		t.Fatalf("NewOrder: %v", err)
	}
	event := PaymentEvent{
		Provider: "provider", EventID: "event-1", OrderID: "order-1",
		Type: PaymentSucceeded, Amount: mustMoney(t, "50000"), ReceivedAt: createdAt.Add(time.Minute),
	}

	decision, err := DecidePaymentEvent(event, &order, false)
	if err != nil {
		t.Fatalf("DecidePaymentEvent: %v", err)
	}
	if decision.Action != PaymentEventApply || decision.Order.Status() != OrderPaid || !decision.IssueGrant {
		t.Fatalf("decision=%+v want apply, paid, grant", decision)
	}
	if !decision.Order.Amount().Equal(mustMoney(t, "50000")) {
		t.Fatal("order snapshot changed while applying payment")
	}
	if decision.GrantSource != GrantPurchase || decision.GrantOrderID != "order-1" ||
		decision.GrantUserID != "user-1" || decision.GrantCourseID != "course-1" {
		t.Fatalf("grant decision=%+v", decision)
	}
}

func TestDecidePaymentEventClassifiesRetryRejectAndDuplicate(t *testing.T) {
	createdAt := time.Date(2026, 9, 30, 9, 0, 0, 0, time.UTC)
	newPending := func(t *testing.T) Order {
		t.Helper()
		order, err := NewOrder("order-1", "user-1", "course-1", mustMoney(t, "50000"), "checkout-1", createdAt)
		if err != nil {
			t.Fatalf("NewOrder: %v", err)
		}
		return order
	}
	baseEvent := PaymentEvent{
		Provider: "provider", EventID: "event-1", OrderID: "order-1",
		Type: PaymentSucceeded, Amount: mustMoney(t, "50000"), ReceivedAt: createdAt.Add(time.Minute),
	}
	for _, test := range []struct {
		name           string
		event          PaymentEvent
		order          func(*testing.T) *Order
		alreadyApplied bool
		wantAction     PaymentEventAction
		wantErr        error
	}{
		{name: "unknown order retries", event: baseEvent, wantAction: PaymentEventRetry},
		{name: "duplicate is no-op before lookup", event: baseEvent, alreadyApplied: true, wantAction: PaymentEventNoOp},
		{name: "amount mismatch is terminal reject", event: func() PaymentEvent { event := baseEvent; event.Amount = mustMoney(t, "70000"); return event }(), order: func(t *testing.T) *Order { order := newPending(t); return &order }, wantAction: PaymentEventReject, wantErr: ErrPaymentEventRejected},
		{name: "unsupported event is terminal reject", event: func() PaymentEvent { event := baseEvent; event.Type = PaymentEventType("refund"); return event }(), order: func(t *testing.T) *Order { order := newPending(t); return &order }, wantAction: PaymentEventReject, wantErr: ErrPaymentEventRejected},
		{name: "wrong order is terminal reject", event: func() PaymentEvent { event := baseEvent; event.OrderID = "order-2"; return event }(), order: func(t *testing.T) *Order { order := newPending(t); return &order }, wantAction: PaymentEventReject, wantErr: ErrPaymentEventRejected},
		{name: "terminal order is reject", event: baseEvent, order: func(t *testing.T) *Order {
			order := newPending(t)
			paid, err := order.MarkPaid(createdAt.Add(30 * time.Second))
			if err != nil {
				t.Fatalf("MarkPaid: %v", err)
			}
			return &paid
		}, wantAction: PaymentEventReject, wantErr: ErrPaymentEventRejected},
	} {
		t.Run(test.name, func(t *testing.T) {
			var order *Order
			if test.order != nil {
				order = test.order(t)
			}
			decision, err := DecidePaymentEvent(test.event, order, test.alreadyApplied)
			if decision.Action != test.wantAction || !errors.Is(err, test.wantErr) {
				t.Fatalf("action=%q error=%v want action=%q error=%v", decision.Action, err, test.wantAction, test.wantErr)
			}
			if decision.IssueGrant {
				t.Fatal("non-apply decision issued a grant")
			}
		})
	}
}
