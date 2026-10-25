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
	// PendingForCourse returns the single unpaid pending order a user holds on a course, or
	// ErrOrderNotFound when there is none. One pending order per (user, course) is the double-charge
	// guard (SA05/M2): StartPurchase reuses this order instead of opening a second that could also
	// pay. A partial unique index enforces the invariant under concurrency.
	PendingForCourse(context.Context, string, string) (domain.Order, error)
	Create(context.Context, domain.Order) (domain.Order, error)
	Get(context.Context, string, string) (domain.Order, error)
	MarkCheckoutOpened(context.Context, string, string) (domain.Order, error)
	// OrderByID reads an order by id alone, for an administrator acting on the refund queue who
	// does not know the buyer up front. ErrOrderNotFound when it does not exist.
	OrderByID(context.Context, string) (domain.Order, error)
	// RefundQueue lists orders flagged refund_required, each with its charged amount and received
	// time read from the paid payment_events row (nil when none exists). state narrows the list:
	// "pending" (not yet settled), "refunded", "declined", or "" for all flagged orders.
	RefundQueue(ctx context.Context, state string) ([]RefundQueueItem, error)
	// SettleRefund records the administrator's decision on a refund-queue row (outcome "refunded"
	// or "declined") and stamps who/when. Idempotent: a no-op when already settled or not flagged.
	SettleRefund(ctx context.Context, orderID, actorID, outcome string) error
}

// RefundQueueItem is one order needing attention: money arrived without the expected grant, or a
// WayForPay refund needs follow-up. Charged amount and received time come from the paid event, so
// they are nil for a row whose settle was rejected before any paid event (Reason "" is historical).
type RefundQueueItem struct {
	OrderID            string
	CourseID           string
	UserID             string
	OrderAmountMinor   int64
	ChargedAmountMinor *int64
	ReceivedAt         *time.Time
	Reason             string
	// MembershipState is the current (user, course) grant state, or nil when no grant row exists.
	// The admin UI offers revoke only for a live grant (active/blocked) — A14.
	MembershipState *domain.GrantState
	// Outcome is the administrator's recorded decision: "refunded", "declined", or "" while the
	// order is still pending in the queue.
	Outcome string
}

// GrantRepository reads and writes membership records and their audit rows.
type GrantRepository interface {
	Membership(context.Context, string, string) (*domain.Grant, error)
	// IsAdministrator reports whether the user is a live administrator, so the service can keep
	// an administrator from being made a learner (G8). The grants adapter already reads users.
	IsAdministrator(context.Context, string) (bool, error)
	MyMemberships(context.Context, string) ([]MembershipCard, error)
	Roster(context.Context, string, string, int) (RosterPage, error)
	Learners(context.Context, LearnerFilter, string, int) (LearnerPage, error)
	// AuthorOf resolves a moderation-queue item (kind "comment" or "review") to its author and the
	// course it belongs to, so the block-from-queue path can reuse the membership transition. An
	// unknown item returns ErrCommentNotFound or ErrReviewNotFound. The grants adapter reads the
	// discussion tables directly (a SQL read, no module import — ADR 013), as the roster already
	// reads lesson_progress.
	AuthorOf(ctx context.Context, kind, id string) (userID, courseID string, err error)
	CreateMembership(context.Context, domain.Grant, string, string, *string) (domain.Grant, error)
	UpdateMembership(context.Context, domain.Grant, domain.GrantState, string, string, *string) (domain.Grant, error)
	GrandfatherOnPriceSet(context.Context, string, time.Time, domain.Money, time.Time) error
}

// LearnerFilter selects which learners and which of their memberships the overview returns.
// Administrator sees every course; otherwise only courses ViewerID instructs. CourseID, State and
// Query are optional narrowings (empty means no narrowing).
type LearnerFilter struct {
	ViewerID      string
	Administrator bool
	CourseID      string
	State         string
	Query         string
}

type LearnerCourse struct {
	CourseID        string
	Slug            string
	Title           string
	State           domain.GrantState
	ProgressPercent int
	EnrolledAt      time.Time
}

type LearnerRow struct {
	UserID         string
	DisplayName    string
	LastActivityAt *time.Time
	Courses        []LearnerCourse
}

type LearnerPage struct {
	Entries []LearnerRow
	// nil on the last page — the overview's cursor is null, not "" (the roster still uses the
	// string sentinel; the learner page carries the absence natively, LR2 review nit).
	NextCursor *string
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
