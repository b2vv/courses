package ports

import (
	"context"
	"time"

	"courses/internal/orders/domain"
)

type CourseFacts struct {
	Published bool
	Price     *domain.Money
	Title     string
}

type EnrollmentFacts struct {
	Status       string
	Free         bool
	InstructorID string
}

type CourseReader interface {
	CourseForPurchase(context.Context, string) (CourseFacts, error)
	CourseForEnrollment(context.Context, string) (EnrollmentFacts, error)
}

type OrderRepository interface {
	FindByIdempotencyKey(context.Context, string, string) (domain.Order, error)
	Create(context.Context, domain.Order) (domain.Order, error)
	Get(context.Context, string, string) (domain.Order, error)
	MarkCheckoutOpened(context.Context, string, string) (domain.Order, error)
}

// GrantRepository owns the transaction which writes a grant and its audit row.
type GrantRepository interface {
	Membership(context.Context, string, string) (*domain.Grant, error)
	CreateMembership(context.Context, domain.Grant, string, string, *string) (domain.Grant, error)
	UpdateMembership(context.Context, domain.Grant, domain.GrantState, string, string, *string) (domain.Grant, error)
	GrandfatherOnPriceSet(context.Context, string, time.Time, domain.Money, time.Time) error
}

// PaymentEventRepository owns the inbox/order/grant/audit transaction.
type PaymentEventRepository interface {
	Apply(context.Context, domain.PaymentEvent, func() string) error
}
