package ports

import (
	"context"
	"time"

	"courses/internal/orders/domain"
)

type CourseFacts struct {
	Published bool
	Price     *domain.Money
}

type CourseReader interface {
	CourseForPurchase(context.Context, string) (CourseFacts, error)
}

type OrderRepository interface {
	FindByIdempotencyKey(context.Context, string, string) (domain.Order, error)
	Create(context.Context, domain.Order) (domain.Order, error)
	Get(context.Context, string, string) (domain.Order, error)
}

type GrantCommand struct {
	Grant  domain.Grant
	Reason string
}

// GrantRepository owns the transaction which writes a grant and its audit row.
type GrantRepository interface {
	ActiveGrants(context.Context, string, string) ([]domain.Grant, error)
	GrantWithAudit(context.Context, GrantCommand) (domain.Grant, error)
	GrandfatherOnPriceSet(context.Context, string, time.Time, domain.Money, time.Time) error
}

// PaymentEventRepository owns the inbox/order/grant/audit transaction.
type PaymentEventRepository interface {
	Apply(context.Context, domain.PaymentEvent, func() string) error
}
