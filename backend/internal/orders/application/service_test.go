package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"courses/internal/orders/application"
	"courses/internal/orders/domain"
	"courses/internal/orders/ports"
)

const (
	userID   = "40000000-0000-4000-8000-000000000001"
	adminID  = "40000000-0000-4000-8000-000000000002"
	courseID = "50000000-0000-4000-8000-000000000001"
	orderID  = "60000000-0000-4000-8000-000000000001"
)

var fixedNow = time.Date(2026, 9, 30, 9, 0, 0, 0, time.UTC)

func TestStartPurchaseSnapshotsPublishedPaidCoursePrice(t *testing.T) {
	store := newFakeStore(t)
	store.course = ports.CourseFacts{Published: true, Price: money(t, "50000")}
	service := application.New(store, store, store, func() time.Time { return fixedNow }, func() string { return orderID })

	order, err := service.StartPurchase(context.Background(), application.Actor{UserID: userID}, courseID, "purchase-key")
	if err != nil {
		t.Fatalf("StartPurchase: %v", err)
	}
	if order.ID() != orderID || order.UserID() != userID || order.CourseID() != courseID || order.Status() != domain.OrderPending || order.Amount().MinorUnits() != 50000 {
		t.Fatalf("unexpected order: id=%q user=%q course=%q status=%q amount=%d", order.ID(), order.UserID(), order.CourseID(), order.Status(), order.Amount().MinorUnits())
	}
}

func TestStartPurchaseIsIdempotentAndRejectsOwnedOrUnpurchasableCourses(t *testing.T) {
	existing := pendingOrder(t)
	for _, test := range []struct {
		name       string
		configure  func(*fakeStore)
		want       error
		wantOrder  string
		wantCreate int
	}{
		{name: "same key returns existing order", configure: func(store *fakeStore) { store.idempotentOrder = &existing }, wantOrder: orderID},
		{name: "active grant blocks purchase", configure: func(store *fakeStore) { store.activeGrant = true }, want: domain.ErrCourseAlreadyOwned},
		{name: "free course", configure: func(store *fakeStore) { store.course.Price = nil }, want: domain.ErrCourseNotPurchasable},
		{name: "unpublished course", configure: func(store *fakeStore) { store.course.Published = false }, want: domain.ErrCourseNotPurchasable},
	} {
		t.Run(test.name, func(t *testing.T) {
			store := newFakeStore(t)
			test.configure(store)
			service := application.New(store, store, store, func() time.Time { return fixedNow }, func() string { return orderID })
			order, err := service.StartPurchase(context.Background(), application.Actor{UserID: userID}, courseID, "purchase-key")
			if !errors.Is(err, test.want) {
				t.Fatalf("error=%v want=%v", err, test.want)
			}
			if test.wantOrder != "" && order.ID() != test.wantOrder {
				t.Fatalf("order id=%q want=%q", order.ID(), test.wantOrder)
			}
			if store.createCalls != test.wantCreate {
				t.Fatalf("create calls=%d want=%d", store.createCalls, test.wantCreate)
			}
		})
	}
}

func TestPurchaseStatusIsScopedToActor(t *testing.T) {
	store := newFakeStore(t)
	existing := pendingOrder(t)
	store.order = &existing
	service := application.New(store, store, store, func() time.Time { return fixedNow }, func() string { return orderID })

	got, err := service.PurchaseStatus(context.Background(), application.Actor{UserID: userID}, orderID)
	if err != nil {
		t.Fatalf("PurchaseStatus: %v", err)
	}
	if got.ID() != orderID || store.getUserID != userID {
		t.Fatalf("order=%q queried user=%q", got.ID(), store.getUserID)
	}
}

func TestApplyPaymentEventDelegatesTheAtomicOperation(t *testing.T) {
	store := newFakeStore(t)
	service := application.New(store, store, store, func() time.Time { return fixedNow }, func() string { return "70000000-0000-4000-8000-000000000001" })
	event := domain.PaymentEvent{Provider: "provider", EventID: "event-1", OrderID: orderID, Type: domain.PaymentSucceeded, Amount: *money(t, "50000"), ReceivedAt: fixedNow}

	if err := service.ApplyPaymentEvent(context.Background(), event); err != nil {
		t.Fatalf("ApplyPaymentEvent: %v", err)
	}
	if store.appliedEvent.EventID != event.EventID {
		t.Fatalf("event=%q want=%q", store.appliedEvent.EventID, event.EventID)
	}
}

func TestGrantCourseRequiresAdministratorAndWritesAudit(t *testing.T) {
	for _, test := range []struct {
		name      string
		admin     bool
		want      error
		wantCalls int
	}{
		{name: "administrator grants", admin: true, wantCalls: 1},
		{name: "student is forbidden", want: domain.ErrGrantForbidden},
	} {
		t.Run(test.name, func(t *testing.T) {
			store := newFakeStore(t)
			service := application.New(store, store, store, func() time.Time { return fixedNow }, func() string { return "70000000-0000-4000-8000-000000000001" })
			grant, err := service.GrantCourse(context.Background(), application.Actor{UserID: adminID, Administrator: test.admin}, userID, courseID, "support decision")
			if !errors.Is(err, test.want) {
				t.Fatalf("error=%v want=%v", err, test.want)
			}
			if store.grantCalls != test.wantCalls {
				t.Fatalf("grant calls=%d want=%d", store.grantCalls, test.wantCalls)
			}
			if test.want == nil && (grant.Source() != domain.GrantAdmin || grant.GrantedBy() == nil || *grant.GrantedBy() != adminID || store.grantCommand.Reason != "support decision") {
				t.Fatalf("unexpected grant or audit command: %#v %#v", grant, store.grantCommand)
			}
		})
	}
}

func TestGrandfatherOnPriceSetUsesOneRepositoryOperation(t *testing.T) {
	store := newFakeStore(t)
	service := application.New(store, store, store, func() time.Time { return fixedNow }, func() string { return orderID })

	updatedAt := fixedNow.Add(time.Microsecond)
	err := service.GrandfatherOnPriceSet(context.Background(), courseID, fixedNow, 50000, "UAH", updatedAt)
	if err != nil {
		t.Fatalf("GrandfatherOnPriceSet: %v", err)
	}
	if store.grandfatherCourse != courseID || !store.grandfatherExpected.Equal(fixedNow) || !store.grandfatherUpdated.Equal(updatedAt) || store.grandfatherPrice.MinorUnits() != 50000 {
		t.Fatalf("course=%q expected=%v updated=%v price=%d", store.grandfatherCourse, store.grandfatherExpected, store.grandfatherUpdated, store.grandfatherPrice.MinorUnits())
	}
}

func newFakeStore(t *testing.T) *fakeStore {
	t.Helper()
	return &fakeStore{course: ports.CourseFacts{Published: true, Price: money(t, "50000")}}
}

type fakeStore struct {
	course              ports.CourseFacts
	idempotentOrder     *domain.Order
	activeGrant         bool
	createCalls         int
	order               *domain.Order
	getUserID           string
	appliedEvent        domain.PaymentEvent
	applyErr            error
	grantCalls          int
	grantCommand        ports.GrantCommand
	grandfatherCourse   string
	grandfatherExpected time.Time
	grandfatherUpdated  time.Time
	grandfatherPrice    domain.Money
}

func (store *fakeStore) CourseForPurchase(context.Context, string) (ports.CourseFacts, error) {
	return store.course, nil
}

func (store *fakeStore) FindByIdempotencyKey(context.Context, string, string) (domain.Order, error) {
	if store.idempotentOrder == nil {
		return domain.Order{}, domain.ErrOrderNotFound
	}
	return *store.idempotentOrder, nil
}

func (store *fakeStore) Create(_ context.Context, order domain.Order) (domain.Order, error) {
	store.createCalls++
	return order, nil
}

func (store *fakeStore) Get(_ context.Context, userID, _ string) (domain.Order, error) {
	store.getUserID = userID
	if store.order == nil {
		return domain.Order{}, domain.ErrOrderNotFound
	}
	return *store.order, nil
}

func (store *fakeStore) ActiveGrants(_ context.Context, userID, courseID string) ([]domain.Grant, error) {
	if !store.activeGrant {
		return nil, nil
	}
	grant, err := domain.NewGrant("grant-1", userID, courseID, domain.GrantAdmin, nil, domain.GrantActive, fixedNow, nil)
	if err != nil {
		return nil, err
	}
	return []domain.Grant{grant}, nil
}

func (store *fakeStore) GrantWithAudit(_ context.Context, command ports.GrantCommand) (domain.Grant, error) {
	store.grantCalls++
	store.grantCommand = command
	return command.Grant, nil
}

func (store *fakeStore) GrandfatherOnPriceSet(_ context.Context, courseID string, expected time.Time, price domain.Money, updated time.Time) error {
	store.grandfatherCourse = courseID
	store.grandfatherExpected = expected
	store.grandfatherPrice = price
	store.grandfatherUpdated = updated
	return nil
}

func (store *fakeStore) Apply(_ context.Context, event domain.PaymentEvent, _ func() string) error {
	store.appliedEvent = event
	return store.applyErr
}

func money(t *testing.T, amount string) *domain.Money {
	t.Helper()
	value, err := domain.NewMoney(amount, "UAH")
	if err != nil {
		t.Fatal(err)
	}
	return &value
}

func pendingOrder(t *testing.T) domain.Order {
	t.Helper()
	order, err := domain.NewOrder(orderID, userID, courseID, *money(t, "50000"), "purchase-key", fixedNow)
	if err != nil {
		t.Fatal(err)
	}
	return order
}
