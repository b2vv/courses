package domain

import "time"

type OrderStatus string

const (
	OrderPending OrderStatus = "pending"
	OrderPaid    OrderStatus = "paid"
	OrderFailed  OrderStatus = "failed"
)

type Order struct {
	idempotencyKey string
	id             string
	userID         string
	courseID       string
	amount         Money
	status         OrderStatus
	createdAt      time.Time
	paidAt         *time.Time
}

func NewOrder(id, userID, courseID string, amount Money, idempotencyKey string, createdAt time.Time) (Order, error) {
	if id == "" || userID == "" || courseID == "" || idempotencyKey == "" || createdAt.IsZero() ||
		amount.MinorUnits() <= 0 || amount.Currency() != CurrencyUAH {
		return Order{}, ErrOrderInvalid
	}
	return Order{
		id:             id,
		userID:         userID,
		courseID:       courseID,
		amount:         amount,
		status:         OrderPending,
		idempotencyKey: idempotencyKey,
		createdAt:      createdAt,
	}, nil
}

func (order Order) ID() string { return order.id }

func (order Order) UserID() string { return order.userID }

func (order Order) CourseID() string { return order.courseID }

func (order Order) Amount() Money { return order.amount }

func (order Order) Status() OrderStatus { return order.status }

func (order Order) IdempotencyKey() string { return order.idempotencyKey }

func (order Order) CreatedAt() time.Time { return order.createdAt }

func (order Order) PaidAt() *time.Time {
	if order.paidAt == nil {
		return nil
	}
	paidAt := *order.paidAt
	return &paidAt
}

func (order Order) MarkPaid(paidAt time.Time) (Order, error) {
	if order.status != OrderPending || paidAt.IsZero() || paidAt.Before(order.createdAt) {
		return Order{}, ErrOrderTransitionInvalid
	}
	order.status = OrderPaid
	order.paidAt = &paidAt
	return order, nil
}

func (order Order) MarkFailed() (Order, error) {
	if order.status != OrderPending {
		return Order{}, ErrOrderTransitionInvalid
	}
	order.status = OrderFailed
	order.paidAt = nil
	return order, nil
}
