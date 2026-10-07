package application

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"courses/internal/orders/domain"
	"courses/internal/orders/ports"
)

type Actor struct {
	UserID        string
	Administrator bool
	EmailVerified bool
}

type Service struct {
	courses  ports.CourseReader
	orders   ports.OrderRepository
	grants   ports.GrantRepository
	events   ports.PaymentEventRepository
	provider ports.PaymentProvider
	statusAt map[string]time.Time
	statusMu sync.Mutex
	now      func() time.Time
	newID    func() string
}

func New(courses ports.CourseReader, orders ports.OrderRepository, store interface {
	ports.GrantRepository
	ports.PaymentEventRepository
}, now func() time.Time, newID func() string, provider ports.PaymentProvider) *Service {
	if now == nil {
		now = time.Now
	}
	return &Service{courses: courses, orders: orders, grants: store, events: store, provider: provider, now: now, newID: newID, statusAt: make(map[string]time.Time)}
}

func (service *Service) StartPurchase(ctx context.Context, actor Actor, courseID, idempotencyKey string) (domain.Order, error) {
	if !actor.EmailVerified {
		return domain.Order{}, domain.ErrEmailNotVerified
	}
	facts, err := service.courses.CourseForPurchase(ctx, courseID)
	if err != nil {
		return domain.Order{}, err
	}
	if !facts.Published || facts.Price == nil {
		return domain.Order{}, domain.ErrCourseNotPurchasable
	}
	membership, err := service.grants.Membership(ctx, actor.UserID, courseID)
	if err != nil {
		return domain.Order{}, err
	}
	if err := domain.CanStartOrder(membership); err != nil {
		return domain.Order{}, err
	}
	if existing, err := service.orders.FindByIdempotencyKey(ctx, actor.UserID, idempotencyKey); err == nil {
		if existing.CourseID() != courseID {
			return domain.Order{}, domain.ErrOrderInvalid
		}
		return existing, nil
	} else if !errors.Is(err, domain.ErrOrderNotFound) {
		return domain.Order{}, err
	}
	order, err := domain.NewOrder(service.newID(), actor.UserID, courseID, *facts.Price, idempotencyKey, service.now())
	if err != nil {
		return domain.Order{}, err
	}
	created, err := service.orders.Create(ctx, order)
	if err != nil {
		return domain.Order{}, err
	}
	if created.CourseID() != courseID {
		return domain.Order{}, domain.ErrOrderInvalid
	}
	return created, nil
}

func (service *Service) PurchaseStatus(ctx context.Context, actor Actor, orderID string) (domain.Order, error) {
	order, err := service.orders.Get(ctx, actor.UserID, orderID)
	if err != nil || order.Status() != domain.OrderPending || service.provider == nil {
		return order, err
	}
	now := service.now()
	service.statusMu.Lock()
	for trackedOrderID, last := range service.statusAt {
		if now.Sub(last) >= 3*time.Second {
			delete(service.statusAt, trackedOrderID)
		}
	}
	last := service.statusAt[orderID]
	if !last.IsZero() && now.Sub(last) < 3*time.Second {
		service.statusMu.Unlock()
		return order, nil
	}
	service.statusAt[orderID] = now
	service.statusMu.Unlock()
	status, statusErr := service.provider.OrderStatus(ctx, order.ID())
	if statusErr == nil && status.Event.OrderID != "" {
		if applyErr := service.ApplyPaymentEvent(ctx, status.Event); applyErr == nil {
			return service.orders.Get(ctx, actor.UserID, orderID)
		}
	}
	return order, nil
}

func (service *Service) Checkout(ctx context.Context, actor Actor, orderID string, buyer ports.Buyer, urls ports.CheckoutURLs) (ports.CheckoutForm, error) {
	if !actor.EmailVerified {
		return ports.CheckoutForm{}, domain.ErrEmailNotVerified
	}
	if service.provider == nil {
		return ports.CheckoutForm{}, domain.ErrDependencyUnavailable
	}
	order, err := service.orders.MarkCheckoutOpened(ctx, actor.UserID, orderID)
	if err != nil {
		return ports.CheckoutForm{}, err
	}
	if urls.ProductName == "" {
		if facts, factsErr := service.courses.CourseForPurchase(ctx, order.CourseID()); factsErr == nil {
			urls.ProductName = facts.Title
		}
	}
	if buyer.ID == "" {
		buyer.ID = actor.UserID
	}
	return service.provider.Checkout(ctx, order, buyer, urls)
}

func (service *Service) ApplyPaymentEvent(ctx context.Context, event domain.PaymentEvent) error {
	return service.events.Apply(ctx, event, service.newID)
}

func (service *Service) GrantCourse(ctx context.Context, actor Actor, userID, courseID, reason string) (domain.Grant, error) {
	if !actor.Administrator {
		return domain.Grant{}, domain.ErrGrantForbidden
	}
	if userID == actor.UserID {
		return domain.Grant{}, domain.ErrGrantSelfForbidden
	}
	reason = strings.TrimSpace(reason)
	if !utf8.ValidString(reason) || utf8.RuneCountInString(reason) < 1 || utf8.RuneCountInString(reason) > 500 {
		return domain.Grant{}, domain.ErrGrantInvalid
	}
	membership, err := service.grants.Membership(ctx, userID, courseID)
	if err != nil {
		return domain.Grant{}, err
	}
	if membership == nil {
		grant, newErr := domain.NewGrant(service.newID(), userID, courseID, domain.GrantAdmin, nil, domain.GrantActive, service.now(), &actor.UserID)
		if newErr != nil {
			return domain.Grant{}, newErr
		}
		return service.grants.CreateMembership(ctx, grant, "grant", reason, &actor.UserID)
	}
	if membership.State() != domain.GrantRevoked {
		return domain.Grant{}, domain.ErrCourseAlreadyOwned
	}
	grant, err := membership.Reinstate(domain.GrantAdmin, &actor.UserID)
	if err != nil {
		return domain.Grant{}, err
	}
	return service.grants.UpdateMembership(ctx, grant, domain.GrantRevoked, "grant", reason, &actor.UserID)
}

func (service *Service) Enroll(ctx context.Context, actor Actor, courseID string) (domain.Grant, error) {
	if !actor.EmailVerified {
		return domain.Grant{}, domain.ErrEmailNotVerified
	}
	facts, err := service.courses.CourseForEnrollment(ctx, courseID)
	if err != nil {
		return domain.Grant{}, err
	}
	if facts.Status != "published" || !facts.Free {
		return domain.Grant{}, domain.ErrCourseNotEnrollable
	}
	membership, err := service.grants.Membership(ctx, actor.UserID, courseID)
	if err != nil {
		return domain.Grant{}, err
	}
	if membership == nil {
		grant, newErr := domain.SelfEnroll(service.newID(), actor.UserID, courseID, service.now())
		if newErr != nil {
			return domain.Grant{}, newErr
		}
		created, createErr := service.grants.CreateMembership(ctx, grant, "enroll", "", &actor.UserID)
		if errors.Is(createErr, domain.ErrCourseAlreadyOwned) {
			// A concurrent enroll won the insert; enrolling an existing member is a no-op,
			// so re-read the row the other transaction created and return it.
			if existing, readErr := service.grants.Membership(ctx, actor.UserID, courseID); readErr == nil && existing != nil {
				return *existing, nil
			}
		}
		return created, createErr
	}
	if membership.State() != domain.GrantRevoked {
		return *membership, nil
	}
	grant, err := membership.Reinstate(domain.GrantSelf, nil)
	if err != nil {
		return domain.Grant{}, err
	}
	return service.grants.UpdateMembership(ctx, grant, domain.GrantRevoked, "enroll", "", &actor.UserID)
}

func (service *Service) Block(ctx context.Context, actor Actor, userID, courseID string) (domain.Grant, error) {
	return service.transitionMembership(ctx, actor, userID, courseID, "block", false, domain.Grant.Block)
}

func (service *Service) Unblock(ctx context.Context, actor Actor, userID, courseID string) (domain.Grant, error) {
	return service.transitionMembership(ctx, actor, userID, courseID, "unblock", false, domain.Grant.Unblock)
}

func (service *Service) Revoke(ctx context.Context, actor Actor, userID, courseID string) (domain.Grant, error) {
	return service.transitionMembership(ctx, actor, userID, courseID, "revoke", true, domain.Grant.Revoke)
}

func (service *Service) MembershipState(ctx context.Context, actor Actor, userID, courseID string) (domain.GrantState, bool, error) {
	facts, err := service.courses.CourseForEnrollment(ctx, courseID)
	if err != nil {
		return "", false, err
	}
	if !actor.Administrator && actor.UserID != userID && actor.UserID != facts.InstructorID {
		return "", false, domain.ErrEnrollmentForbidden
	}
	membership, err := service.grants.Membership(ctx, userID, courseID)
	if err != nil {
		return "", false, err
	}
	if membership == nil {
		return "", false, nil
	}
	return membership.State(), true, nil
}

func (service *Service) MyMemberships(ctx context.Context, actor Actor) ([]ports.MembershipCard, error) {
	return service.grants.MyMemberships(ctx, actor.UserID)
}

func (service *Service) Roster(ctx context.Context, actor Actor, courseID, cursor string, limit int) (ports.RosterPage, error) {
	facts, err := service.courses.CourseForEnrollment(ctx, courseID)
	if err != nil {
		return ports.RosterPage{}, err
	}
	if !actor.Administrator && actor.UserID != facts.InstructorID {
		return ports.RosterPage{}, domain.ErrEnrollmentForbidden
	}
	if limit <= 0 {
		limit = 20
	} else if limit > 100 {
		limit = 100
	}
	return service.grants.Roster(ctx, courseID, cursor, limit)
}

func (service *Service) transitionMembership(ctx context.Context, actor Actor, userID, courseID, action string, administratorOnly bool, transition func(domain.Grant) (domain.Grant, error)) (domain.Grant, error) {
	facts, err := service.courses.CourseForEnrollment(ctx, courseID)
	if err != nil {
		return domain.Grant{}, err
	}
	if !actor.Administrator && (administratorOnly || actor.UserID != facts.InstructorID) {
		return domain.Grant{}, domain.ErrEnrollmentForbidden
	}
	membership, err := service.grants.Membership(ctx, userID, courseID)
	if err != nil {
		return domain.Grant{}, err
	}
	if membership == nil {
		if action == "revoke" {
			return domain.Grant{}, domain.ErrGrantInvalid
		}
		return domain.Grant{}, domain.ErrGrantTransitionInvalid
	}
	if (action == "block" && membership.State() == domain.GrantBlocked) ||
		(action == "unblock" && membership.State() == domain.GrantActive) ||
		(action == "revoke" && membership.State() == domain.GrantRevoked) {
		return *membership, nil
	}
	grant, err := transition(*membership)
	if err != nil {
		return domain.Grant{}, err
	}
	return service.grants.UpdateMembership(ctx, grant, membership.State(), action, "", &actor.UserID)
}

func (service *Service) GrandfatherOnPriceSet(ctx context.Context, courseID string, expectedUpdatedAt time.Time, priceMinor int64, currency string, updatedAt time.Time) error {
	price, err := domain.NewMoney(strconv.FormatInt(priceMinor, 10), currency)
	if err != nil || price.MinorUnits() <= 0 || expectedUpdatedAt.IsZero() || updatedAt.IsZero() {
		return domain.ErrMoneyInvalid
	}
	return service.grants.GrandfatherOnPriceSet(ctx, courseID, expectedUpdatedAt, price, updatedAt)
}
