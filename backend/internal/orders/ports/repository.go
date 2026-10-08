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

// GrantRepository reads and writes membership records and their audit rows.
type GrantRepository interface {
	Membership(context.Context, string, string) (*domain.Grant, error)
	// IsAdministrator reports whether the user is a live administrator, so the service can keep
	// an administrator from being made a learner (G8). The grants adapter already reads users.
	IsAdministrator(context.Context, string) (bool, error)
	MyMemberships(context.Context, string) ([]MembershipCard, error)
	Roster(context.Context, string, string, int) (RosterPage, error)
	CreateMembership(context.Context, domain.Grant, string, string, *string) (domain.Grant, error)
	UpdateMembership(context.Context, domain.Grant, domain.GrantState, string, string, *string) (domain.Grant, error)
	GrandfatherOnPriceSet(context.Context, string, time.Time, domain.Money, time.Time) error
}

type MembershipCard struct {
	CourseID        string
	Slug            string
	Title           string
	State           domain.GrantState
	ProgressPercent int
	EnrolledAt      time.Time
}

type RosterEntry struct {
	UserID          string
	DisplayName     string
	State           domain.GrantState
	EnrolledAt      time.Time
	ProgressPercent int
}

type RosterPage struct {
	Entries    []RosterEntry
	NextCursor string
}

// PaymentEventRepository owns the inbox/order/grant/audit transaction.
type PaymentEventRepository interface {
	Apply(context.Context, domain.PaymentEvent, func() string) error
}
