package application_test

import (
	"context"
	"errors"
	"sync"
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
	service := application.New(store, store, store, func() time.Time { return fixedNow }, func() string { return orderID }, nil)

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
	paidExisting, err := existing.MarkPaid(fixedNow.Add(time.Second))
	if err != nil {
		t.Fatal(err)
	}
	failedExisting, err := existing.MarkFailed()
	if err != nil {
		t.Fatal(err)
	}
	otherCourseOrder, err := domain.NewOrder(orderID, userID, "50000000-0000-4000-8000-000000000002", *money(t, "50000"), "purchase-key", fixedNow)
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name       string
		configure  func(*fakeStore)
		want       error
		wantOrder  string
		wantCreate int
	}{
		{name: "same key pending returns existing order", configure: func(store *fakeStore) { store.idempotentOrder = &existing }, wantOrder: orderID},
		// Paid without an active grant is not a production state (paid and the grant
		// are written in one transaction); it stands in for a grant revoked since.
		{name: "same key paid without active grant returns paid order", configure: func(store *fakeStore) { store.idempotentOrder = &paidExisting }, wantOrder: orderID},
		// The production replay of a paid key: the owner check answers first (Pay12
		// accepts already_owned here rather than the order itself).
		{name: "same key after a completed purchase is already owned", configure: func(store *fakeStore) {
			store.idempotentOrder = &paidExisting
			store.activeGrant = true
		}, want: domain.ErrCourseAlreadyOwned},
		{name: "same key failed returns failed order", configure: func(store *fakeStore) { store.idempotentOrder = &failedExisting }, wantOrder: orderID},
		// The same key raced for another course: the insert hands back the other
		// course's order, which must never be answered as this course's purchase.
		{name: "same key raced for another course is refused", configure: func(store *fakeStore) { store.createReturns = &otherCourseOrder }, want: domain.ErrOrderInvalid, wantCreate: 1},
		{name: "active grant blocks purchase", configure: func(store *fakeStore) { store.activeGrant = true }, want: domain.ErrCourseAlreadyOwned},
		{name: "free course", configure: func(store *fakeStore) { store.course.Price = nil }, want: domain.ErrCourseNotPurchasable},
		{name: "unpublished course", configure: func(store *fakeStore) { store.course.Published = false }, want: domain.ErrCourseNotPurchasable},
	} {
		t.Run(test.name, func(t *testing.T) {
			store := newFakeStore(t)
			test.configure(store)
			service := application.New(store, store, store, func() time.Time { return fixedNow }, func() string { return orderID }, nil)
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
	service := application.New(store, store, store, func() time.Time { return fixedNow }, func() string { return orderID }, nil)

	got, err := service.PurchaseStatus(context.Background(), application.Actor{UserID: userID}, orderID)
	if err != nil {
		t.Fatalf("PurchaseStatus: %v", err)
	}
	if got.ID() != orderID || store.getUserID != userID {
		t.Fatalf("order=%q queried user=%q", got.ID(), store.getUserID)
	}
}

func TestPurchaseStatusChecksProviderAndAppliesApprovedOnce(t *testing.T) {
	store := newFakeStore(t)
	order := pendingOrder(t)
	store.order = &order
	provider := &fakePaymentProvider{status: ports.PaymentStatus{Event: domain.PaymentEvent{Provider: "provider", EventID: "approved-1", OrderID: orderID, Type: domain.PaymentSucceeded, Amount: *money(t, "50000"), ReceivedAt: fixedNow}}}
	service := application.New(store, store, store, func() time.Time { return fixedNow }, func() string { return "70000000-0000-4000-8000-000000000001" }, provider)
	if _, err := service.PurchaseStatus(context.Background(), application.Actor{UserID: userID}, orderID); err != nil {
		t.Fatal(err)
	}
	if provider.calls != 1 || store.appliedEvent.EventID != "approved-1" {
		t.Fatalf("provider calls=%d event=%q", provider.calls, store.appliedEvent.EventID)
	}
}

func TestPurchaseStatusIgnoresProviderFailure(t *testing.T) {
	store := newFakeStore(t)
	order := pendingOrder(t)
	store.order = &order
	provider := &fakePaymentProvider{err: errors.New("provider unavailable")}
	service := application.New(store, store, store, func() time.Time { return fixedNow }, func() string { return orderID }, provider)
	got, err := service.PurchaseStatus(context.Background(), application.Actor{UserID: userID}, orderID)
	if err != nil || got.ID() != orderID || got.Status() != domain.OrderPending {
		t.Fatalf("order=%+v error=%v", got, err)
	}
}

func TestCheckoutClaimsAnOrderOnce(t *testing.T) {
	store := newFakeStore(t)
	order := pendingOrder(t)
	store.order = &order
	provider := &fakePaymentProvider{}
	service := application.New(store, store, store, func() time.Time { return fixedNow }, func() string { return orderID }, provider)

	if _, err := service.Checkout(context.Background(), application.Actor{UserID: userID}, orderID, ports.Buyer{}, ports.CheckoutURLs{}); err != nil {
		t.Fatalf("first checkout: %v", err)
	}
	if _, err := service.Checkout(context.Background(), application.Actor{UserID: userID}, orderID, ports.Buyer{}, ports.CheckoutURLs{}); !errors.Is(err, domain.ErrCheckoutAlreadyOpened) {
		t.Fatalf("second checkout error=%v want=%v", err, domain.ErrCheckoutAlreadyOpened)
	}
	if provider.checkoutCalls != 1 {
		t.Fatalf("provider checkout calls=%d want=1", provider.checkoutCalls)
	}
}

func TestPurchaseStatusThrottleIsSafeUnderConcurrentPolling(t *testing.T) {
	store := newFakeStore(t)
	order := pendingOrder(t)
	store.order = &order
	concurrentStore := &concurrentStatusStore{fakeStore: store}
	provider := &concurrentPaymentProvider{}
	service := application.New(concurrentStore, concurrentStore, concurrentStore, func() time.Time { return fixedNow }, func() string { return orderID }, provider)

	var wait sync.WaitGroup
	for range 32 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			got, err := service.PurchaseStatus(context.Background(), application.Actor{UserID: userID}, orderID)
			if err != nil || got.Status() != domain.OrderPending {
				t.Errorf("order=%q error=%v", got.Status(), err)
			}
		}()
	}
	wait.Wait()
}

func TestApplyPaymentEventDelegatesTheAtomicOperation(t *testing.T) {
	store := newFakeStore(t)
	service := application.New(store, store, store, func() time.Time { return fixedNow }, func() string { return "70000000-0000-4000-8000-000000000001" }, nil)
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
		targetID  string
		want      error
		wantCalls int
	}{
		{name: "administrator grants another user", admin: true, targetID: userID, wantCalls: 1},
		{name: "administrator cannot grant themselves", admin: true, targetID: adminID, want: domain.ErrGrantSelfForbidden},
		{name: "student is forbidden", targetID: userID, want: domain.ErrGrantForbidden},
	} {
		t.Run(test.name, func(t *testing.T) {
			store := newFakeStore(t)
			service := application.New(store, store, store, func() time.Time { return fixedNow }, func() string { return "70000000-0000-4000-8000-000000000001" }, nil)
			grant, err := service.GrantCourse(context.Background(), application.Actor{UserID: adminID, Administrator: test.admin}, test.targetID, courseID, "support decision")
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
	service := application.New(store, store, store, func() time.Time { return fixedNow }, func() string { return orderID }, nil)

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
	createReturns       *domain.Order
	activeGrant         bool
	createCalls         int
	checkoutCalls       int
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

type fakePaymentProvider struct {
	status        ports.PaymentStatus
	err           error
	calls         int
	checkoutCalls int
}

type concurrentStatusStore struct {
	*fakeStore
	mu sync.Mutex
}

func (store *concurrentStatusStore) Get(_ context.Context, _ string, _ string) (domain.Order, error) {
	store.mu.Lock()
	defer store.mu.Unlock()
	if store.order == nil {
		return domain.Order{}, domain.ErrOrderNotFound
	}
	return *store.order, nil
}

type concurrentPaymentProvider struct {
	mu sync.Mutex
}

func (provider *concurrentPaymentProvider) Checkout(context.Context, domain.Order, ports.Buyer, ports.CheckoutURLs) (ports.CheckoutForm, error) {
	return ports.CheckoutForm{}, nil
}

func (provider *concurrentPaymentProvider) OrderStatus(context.Context, string) (ports.PaymentStatus, error) {
	provider.mu.Lock()
	defer provider.mu.Unlock()
	return ports.PaymentStatus{}, nil
}

func (provider *fakePaymentProvider) Checkout(context.Context, domain.Order, ports.Buyer, ports.CheckoutURLs) (ports.CheckoutForm, error) {
	provider.checkoutCalls++
	return ports.CheckoutForm{}, nil
}
func (provider *fakePaymentProvider) OrderStatus(context.Context, string) (ports.PaymentStatus, error) {
	provider.calls++
	return provider.status, provider.err
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
	// A unique (user, key) conflict makes the repository hand back the row that won
	// the race; createReturns models that row.
	if store.createReturns != nil {
		return *store.createReturns, nil
	}
	return order, nil
}

func (store *fakeStore) Get(_ context.Context, userID, _ string) (domain.Order, error) {
	store.getUserID = userID
	if store.order == nil {
		return domain.Order{}, domain.ErrOrderNotFound
	}
	return *store.order, nil
}

func (store *fakeStore) MarkCheckoutOpened(_ context.Context, userID, orderID string) (domain.Order, error) {
	store.getUserID = userID
	if store.order == nil || store.order.ID() != orderID {
		return domain.Order{}, domain.ErrOrderNotFound
	}
	if store.checkoutCalls > 0 {
		return domain.Order{}, domain.ErrCheckoutAlreadyOpened
	}
	store.checkoutCalls++
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
