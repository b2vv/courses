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

func TestDecidePaymentEventFailsPendingOrderWithoutGrantAndPaidCannotReopen(t *testing.T) {
	createdAt := time.Date(2026, 9, 30, 9, 0, 0, 0, time.UTC)
	order, err := NewOrder("order-1", "user-1", "course-1", mustMoney(t, "50000"), "checkout-1", createdAt)
	if err != nil {
		t.Fatal(err)
	}
	failedEvent := PaymentEvent{Provider: "provider", EventID: "failed-1", OrderID: "order-1", Type: PaymentFailed, Amount: mustMoney(t, "50000"), ReceivedAt: createdAt.Add(time.Minute)}
	decision, err := DecidePaymentEvent(failedEvent, &order, false)
	if err != nil || decision.Action != PaymentEventApply || decision.Order.Status() != OrderFailed || decision.IssueGrant {
		t.Fatalf("decision=%+v error=%v", decision, err)
	}
	replay, err := DecidePaymentEvent(failedEvent, &order, true)
	if err != nil || replay.Action != PaymentEventNoOp {
		t.Fatalf("replay=%+v error=%v", replay, err)
	}
	paid, err := order.MarkPaid(createdAt.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := DecidePaymentEvent(failedEvent, &paid, false); err != ErrPaymentEventRejected {
		t.Fatalf("failed event reopened paid order: %v", err)
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
		{name: "unsupported event is terminal reject", event: func() PaymentEvent { event := baseEvent; event.Type = PaymentEventType("chargeback"); return event }(), order: func(t *testing.T) *Order { order := newPending(t); return &order }, wantAction: PaymentEventReject, wantErr: ErrPaymentEventRejected},
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

func TestDecidePaymentEventRefund(t *testing.T) {
	createdAt := time.Date(2026, 9, 30, 9, 0, 0, 0, time.UTC)
	newPaid := func(t *testing.T) Order {
		t.Helper()
		order, err := NewOrder("order-1", "user-1", "course-1", mustMoney(t, "50000"), "checkout-1", createdAt)
		if err != nil {
			t.Fatalf("NewOrder: %v", err)
		}
		paid, err := order.MarkPaid(createdAt.Add(time.Minute))
		if err != nil {
			t.Fatalf("MarkPaid: %v", err)
		}
		return paid
	}
	newPending := func(t *testing.T) Order {
		t.Helper()
		order, err := NewOrder("order-1", "user-1", "course-1", mustMoney(t, "50000"), "checkout-1", createdAt)
		if err != nil {
			t.Fatalf("NewOrder: %v", err)
		}
		return order
	}
	refund := func(amount string) PaymentEvent {
		return PaymentEvent{Provider: "provider", EventID: "refund-1", OrderID: "order-1", Type: PaymentRefunded, Amount: mustMoney(t, amount), ReceivedAt: createdAt.Add(2 * time.Minute)}
	}

	t.Run("partial refund on a paid order revokes, amount not matched", func(t *testing.T) {
		order := newPaid(t)
		decision, err := DecidePaymentEvent(refund("10000"), &order, false)
		if err != nil {
			t.Fatalf("DecidePaymentEvent: %v", err)
		}
		if decision.Action != PaymentEventRevoke || decision.GrantOrderID != "order-1" || decision.IssueGrant {
			t.Fatalf("decision=%+v want revoke, order-1, no grant", decision)
		}
	})

	t.Run("void (zero amount) on a paid order revokes", func(t *testing.T) {
		order := newPaid(t)
		decision, err := DecidePaymentEvent(refund("0"), &order, false)
		if err != nil || decision.Action != PaymentEventRevoke {
			t.Fatalf("decision=%+v error=%v want revoke", decision, err)
		}
	})

	t.Run("refund on a non-paid order is a benign ignore, not reject", func(t *testing.T) {
		for _, order := range []Order{newPending(t), func() Order { o := newPending(t); f, _ := o.MarkFailed(); return f }()} {
			decision, err := DecidePaymentEvent(refund("50000"), &order, false)
			if err != nil || decision.Action != PaymentEventIgnore {
				t.Fatalf("status=%s decision=%+v error=%v want ignore", order.Status(), decision, err)
			}
		}
	})

	t.Run("duplicate refund is a no-op before order lookup", func(t *testing.T) {
		order := newPaid(t)
		decision, err := DecidePaymentEvent(refund("50000"), &order, true)
		if err != nil || decision.Action != PaymentEventNoOp {
			t.Fatalf("decision=%+v error=%v want no-op", decision, err)
		}
	})

	t.Run("refund for a different order is a terminal reject", func(t *testing.T) {
		order := newPaid(t)
		event := refund("50000")
		event.OrderID = "order-2"
		if _, err := DecidePaymentEvent(event, &order, false); err != ErrPaymentEventRejected {
			t.Fatalf("mismatched refund error=%v want rejected", err)
		}
	})
}
