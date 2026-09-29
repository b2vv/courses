package domain

import (
	"errors"
	"testing"
	"time"
)

func TestNewOrderSnapshotsMoneyAsPending(t *testing.T) {
	amount := mustMoney(t, "50000")
	createdAt := time.Date(2026, 9, 30, 9, 0, 0, 0, time.UTC)
	order, err := NewOrder("order-1", "user-1", "course-1", amount, "checkout-1", createdAt)
	if err != nil {
		t.Fatalf("NewOrder: %v", err)
	}
	if order.Status() != OrderPending {
		t.Fatalf("status=%q want=%q", order.Status(), OrderPending)
	}
	if !order.Amount().Equal(amount) {
		t.Fatalf("amount=%v %v want snapshot=%v %v", order.Amount().MinorUnits(), order.Amount().Currency(), amount.MinorUnits(), amount.Currency())
	}
	if order.PaidAt() != nil {
		t.Fatalf("paid at=%v want=nil", order.PaidAt())
	}
}

func TestOrderAllowsOnlyPendingToTerminalTransitionsAndKeepsSnapshot(t *testing.T) {
	amount := mustMoney(t, "50000")
	createdAt := time.Date(2026, 9, 30, 9, 0, 0, 0, time.UTC)
	paidAt := createdAt.Add(time.Minute)

	for _, test := range []struct {
		name       string
		transition func(Order) (Order, error)
		wantStatus OrderStatus
		wantPaidAt bool
	}{
		{name: "pending to paid", transition: func(order Order) (Order, error) { return order.MarkPaid(paidAt) }, wantStatus: OrderPaid, wantPaidAt: true},
		{name: "pending to failed", transition: func(order Order) (Order, error) { return order.MarkFailed() }, wantStatus: OrderFailed},
	} {
		t.Run(test.name, func(t *testing.T) {
			pending, err := NewOrder("order-1", "user-1", "course-1", amount, "checkout-1", createdAt)
			if err != nil {
				t.Fatalf("NewOrder: %v", err)
			}
			got, err := test.transition(pending)
			if err != nil {
				t.Fatalf("transition: %v", err)
			}
			if got.Status() != test.wantStatus || !got.Amount().Equal(amount) {
				t.Fatalf("status=%q amount=%v want status=%q unchanged amount", got.Status(), got.Amount().MinorUnits(), test.wantStatus)
			}
			if (got.PaidAt() != nil) != test.wantPaidAt {
				t.Fatalf("paid at=%v want present=%v", got.PaidAt(), test.wantPaidAt)
			}
		})
	}
}

func TestOrderRejectsInvalidCreationAndTerminalTransitions(t *testing.T) {
	createdAt := time.Date(2026, 9, 30, 9, 0, 0, 0, time.UTC)
	zero := mustMoney(t, "0")
	if _, err := NewOrder("order-1", "user-1", "course-1", zero, "checkout-1", createdAt); !errors.Is(err, ErrOrderInvalid) {
		t.Fatalf("zero order error=%v want=%v", err, ErrOrderInvalid)
	}

	pending, err := NewOrder("order-1", "user-1", "course-1", mustMoney(t, "50000"), "checkout-1", createdAt)
	if err != nil {
		t.Fatalf("NewOrder: %v", err)
	}
	paid, err := pending.MarkPaid(createdAt.Add(time.Minute))
	if err != nil {
		t.Fatalf("MarkPaid: %v", err)
	}
	for _, test := range []struct {
		name       string
		transition func() error
	}{
		{name: "paid to paid", transition: func() error { _, err := paid.MarkPaid(createdAt.Add(2 * time.Minute)); return err }},
		{name: "paid to failed", transition: func() error { _, err := paid.MarkFailed(); return err }},
	} {
		t.Run(test.name, func(t *testing.T) {
			if err := test.transition(); !errors.Is(err, ErrOrderTransitionInvalid) {
				t.Fatalf("error=%v want=%v", err, ErrOrderTransitionInvalid)
			}
		})
	}
}

func TestRestoreOrderValidatesPersistedStatusAndPaidTimestamp(t *testing.T) {
	createdAt := time.Date(2026, 9, 30, 9, 0, 0, 0, time.UTC)
	paidAt := createdAt.Add(time.Minute)
	for _, test := range []struct {
		name   string
		status OrderStatus
		paidAt *time.Time
		want   error
	}{
		{name: "pending", status: OrderPending},
		{name: "paid", status: OrderPaid, paidAt: &paidAt},
		{name: "failed", status: OrderFailed},
		{name: "paid without timestamp", status: OrderPaid, want: ErrOrderInvalid},
		{name: "pending with timestamp", status: OrderPending, paidAt: &paidAt, want: ErrOrderInvalid},
		{name: "unknown status", status: OrderStatus("refunded"), want: ErrOrderInvalid},
	} {
		t.Run(test.name, func(t *testing.T) {
			order, err := RestoreOrder("order-1", "user-1", "course-1", mustMoney(t, "50000"), test.status, "key", createdAt, test.paidAt)
			if !errors.Is(err, test.want) {
				t.Fatalf("error=%v want=%v", err, test.want)
			}
			if test.want == nil && order.Status() != test.status {
				t.Fatalf("status=%q want=%q", order.Status(), test.status)
			}
		})
	}
}

func mustMoney(t *testing.T, minor string) Money {
	t.Helper()
	money, err := NewMoney(minor, "UAH")
	if err != nil {
		t.Fatalf("NewMoney: %v", err)
	}
	return money
}
