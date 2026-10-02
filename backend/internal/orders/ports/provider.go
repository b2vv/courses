package ports

import (
	"context"
	"time"

	"courses/internal/orders/domain"
)

type Buyer struct {
	ID    string
	Email string
}

type CheckoutURLs struct {
	ReturnURL   string
	ServiceURL  string
	ProductName string
}

type CheckoutField struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type CheckoutForm struct {
	Action string          `json:"action"`
	Fields []CheckoutField `json:"fields"`
}

type PaymentStatus struct {
	Event domain.PaymentEvent
}

type PaymentProvider interface {
	Checkout(context.Context, domain.Order, Buyer, CheckoutURLs) (CheckoutForm, error)
	OrderStatus(context.Context, string) (PaymentStatus, error)
}

type Callback struct {
	Event          *domain.PaymentEvent
	OrderReference string
	ReceivedAt     time.Time
}

type CallbackVerifier interface {
	VerifyCallback([]byte, time.Time) (Callback, error)
	Acknowledge(string, time.Time) []byte
}
