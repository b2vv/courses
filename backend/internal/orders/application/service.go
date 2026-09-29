package application

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"courses/internal/orders/domain"
	"courses/internal/orders/ports"
)

type Actor struct {
	UserID        string
	Administrator bool
}

type Service struct {
	courses ports.CourseReader
	orders  ports.OrderRepository
	grants  ports.GrantRepository
	events  ports.PaymentEventRepository
	now     func() time.Time
	newID   func() string
}

func New(courses ports.CourseReader, orders ports.OrderRepository, store interface {
	ports.GrantRepository
	ports.PaymentEventRepository
}, now func() time.Time, newID func() string) *Service {
	if now == nil {
		now = time.Now
	}
	return &Service{courses: courses, orders: orders, grants: store, events: store, now: now, newID: newID}
}

func (service *Service) StartPurchase(ctx context.Context, actor Actor, courseID, idempotencyKey string) (domain.Order, error) {
	facts, err := service.courses.CourseForPurchase(ctx, courseID)
	if err != nil {
		return domain.Order{}, err
	}
	if !facts.Published || facts.Price == nil {
		return domain.Order{}, domain.ErrCourseNotPurchasable
	}
	grants, err := service.grants.ActiveGrants(ctx, actor.UserID, courseID)
	if err != nil {
		return domain.Order{}, err
	}
	if err := domain.CanStartOrder(actor.UserID, courseID, grants); err != nil {
		return domain.Order{}, err
	}
	if existing, err := service.orders.FindByIdempotencyKey(ctx, actor.UserID, idempotencyKey); err == nil {
		if existing.CourseID() != courseID || existing.Status() != domain.OrderPending {
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
	return service.orders.Create(ctx, order)
}

func (service *Service) PurchaseStatus(ctx context.Context, actor Actor, orderID string) (domain.Order, error) {
	return service.orders.Get(ctx, actor.UserID, orderID)
}

func (service *Service) ApplyPaymentEvent(ctx context.Context, event domain.PaymentEvent) error {
	return service.events.Apply(ctx, event, service.newID)
}

func (service *Service) GrantCourse(ctx context.Context, actor Actor, userID, courseID, reason string) (domain.Grant, error) {
	if !actor.Administrator {
		return domain.Grant{}, domain.ErrGrantForbidden
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
