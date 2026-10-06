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
	grant, err := domain.NewGrant(service.newID(), userID, courseID, domain.GrantAdmin, nil, domain.GrantActive, service.now(), &actor.UserID)
	if err != nil {
		return domain.Grant{}, err
	}
	return service.grants.GrantWithAudit(ctx, ports.GrantCommand{Grant: grant, Reason: reason})
}

func (service *Service) GrandfatherOnPriceSet(ctx context.Context, courseID string, expectedUpdatedAt time.Time, priceMinor int64, currency string, updatedAt time.Time) error {
	price, err := domain.NewMoney(strconv.FormatInt(priceMinor, 10), currency)
	if err != nil || price.MinorUnits() <= 0 || expectedUpdatedAt.IsZero() || updatedAt.IsZero() {
		return domain.ErrMoneyInvalid
	}
	return service.grants.GrandfatherOnPriceSet(ctx, courseID, expectedUpdatedAt, price, updatedAt)
}
