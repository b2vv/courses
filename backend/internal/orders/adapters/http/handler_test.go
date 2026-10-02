package http_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	ordershttp "courses/internal/orders/adapters/http"
	"courses/internal/orders/application"
	"courses/internal/orders/domain"
	"courses/internal/orders/ports"
	"courses/platform/httpserver"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/legacy"
)

const (
	testUser   = "10000000-0000-4000-8000-000000000001"
	testAdmin  = "10000000-0000-4000-8000-000000000002"
	testCourse = "20000000-0000-4000-8000-000000000001"
	testOrder  = "30000000-0000-4000-8000-000000000001"
	testGrant  = "40000000-0000-4000-8000-000000000001"
)

type fakeStore struct {
	order       domain.Order
	grant       domain.Grant
	courseErr   error
	getErr      error
	grantErr    error
	active      []domain.Grant
	applyCalls  int
	applyErr    error
	grantCalls  int
	checkoutErr error
}

func (store *fakeStore) CourseForPurchase(context.Context, string) (ports.CourseFacts, error) {
	if store.courseErr != nil {
		return ports.CourseFacts{}, store.courseErr
	}
	money, _ := domain.NewMoney("50000", "UAH")
	return ports.CourseFacts{Published: true, Price: &money}, nil
}
func (store *fakeStore) FindByIdempotencyKey(context.Context, string, string) (domain.Order, error) {
	if store.order.ID() == "" {
		return domain.Order{}, domain.ErrOrderNotFound
	}
	return store.order, nil
}
func (store *fakeStore) Create(_ context.Context, order domain.Order) (domain.Order, error) {
	store.order = order
	return order, nil
}
func (store *fakeStore) Get(context.Context, string, string) (domain.Order, error) {
	if store.getErr != nil {
		return domain.Order{}, store.getErr
	}
	if store.order.ID() == "" {
		return domain.Order{}, domain.ErrOrderNotFound
	}
	return store.order, nil
}
func (store *fakeStore) MarkCheckoutOpened(ctx context.Context, userID, orderID string) (domain.Order, error) {
	if store.checkoutErr != nil {
		return domain.Order{}, store.checkoutErr
	}
	if store.getErr != nil {
		return domain.Order{}, store.getErr
	}
	if store.order.ID() == "" || store.order.Status() != domain.OrderPending {
		return domain.Order{}, domain.ErrOrderNotFound
	}
	return store.order, nil
}
func (store *fakeStore) ActiveGrants(context.Context, string, string) ([]domain.Grant, error) {
	return store.active, nil
}
func (store *fakeStore) GrantWithAudit(_ context.Context, command ports.GrantCommand) (domain.Grant, error) {
	store.grantCalls++
	if store.grantErr != nil {
		return domain.Grant{}, store.grantErr
	}
	store.grant = command.Grant
	return command.Grant, nil
}
func (*fakeStore) GrandfatherOnPriceSet(context.Context, string, time.Time, domain.Money, time.Time) error {
	return nil
}
func (store *fakeStore) Apply(context.Context, domain.PaymentEvent, func() string) error {
	store.applyCalls++
	return store.applyErr
}
func TestPurchaseAndGrantHTTPMatchOpenAPI(t *testing.T) {
	now := time.Date(2026, 9, 30, 12, 0, 0, 0, time.UTC)
	store := &fakeStore{}
	ids := []string{testOrder, testGrant, "40000000-0000-4000-8000-000000000002"}
	service := application.New(store, store, store, func() time.Time { return now }, func() string {
		id := ids[0]
		ids = ids[1:]
		return id
	}, nil)
	handler := ordershttp.NewHandler(service, nil, "")
	authenticate := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", "no-store")
			role := r.Header.Get("X-Test-Role")
			userID := testUser
			if role == "admin" {
				userID = testAdmin
			}
			next.ServeHTTP(w, r.WithContext(httpserver.WithActor(r.Context(), httpserver.Actor{UserID: userID, Role: role})))
		})
	}
	identity := func(next http.Handler) http.Handler { return next }
	administrator := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("X-Test-Role") != "admin" {
				httpserver.WriteError(w, r, http.StatusForbidden, "grant_forbidden", "Administrator permission is required")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
	server := httpserver.NewHandler(httpserver.Options{Service: "test", Register: func(mux *http.ServeMux) {
		handler.Register(mux, authenticate, identity)
		handler.RegisterAdmin(mux, authenticate, identity, administrator)
	}})

	start := httptest.NewRequest(http.MethodPost, "/api/v1/learn/courses/"+testCourse+"/purchase", nil)
	start.Header.Set("Idempotency-Key", "purchase-1")
	start.Header.Set("X-Test-Role", "student")
	response := httptest.NewRecorder()
	server.ServeHTTP(response, start)
	if response.Code != http.StatusCreated || store.order.ID() != testOrder {
		t.Fatalf("start status=%d body=%s order=%s", response.Code, response.Body.String(), store.order.ID())
	}
	assertContract(t, start, response)

	replay := httptest.NewRequest(http.MethodPost, "/api/v1/learn/courses/"+testCourse+"/purchase", nil)
	replay.Header.Set("Idempotency-Key", "purchase-1")
	replay.Header.Set("X-Test-Role", "student")
	response = httptest.NewRecorder()
	server.ServeHTTP(response, replay)
	if response.Code != http.StatusCreated || store.order.ID() != testOrder {
		t.Fatalf("replay status=%d body=%s order=%s", response.Code, response.Body.String(), store.order.ID())
	}
	assertContract(t, replay, response)

	adminID := testAdmin
	owned, err := domain.NewGrant(testGrant, testUser, testCourse, domain.GrantAdmin, nil, domain.GrantActive, now, &adminID)
	if err != nil {
		t.Fatal(err)
	}
	store.active = []domain.Grant{owned}
	conflict := httptest.NewRequest(http.MethodPost, "/api/v1/learn/courses/"+testCourse+"/purchase", nil)
	conflict.Header.Set("Idempotency-Key", "purchase-owned")
	conflict.Header.Set("X-Test-Role", "student")
	response = httptest.NewRecorder()
	server.ServeHTTP(response, conflict)
	if response.Code != http.StatusConflict {
		t.Fatalf("owned status=%d body=%s", response.Code, response.Body.String())
	}
	assertContract(t, conflict, response)
	store.active = nil

	status := httptest.NewRequest(http.MethodGet, "/api/v1/learn/orders/"+testOrder, nil)
	status.Header.Set("X-Test-Role", "student")
	response = httptest.NewRecorder()
	server.ServeHTTP(response, status)
	if response.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	assertContract(t, status, response)

	grantBody := []byte(`{"userId":"` + testUser + `","reason":"support resolution"}`)
	for _, test := range []struct {
		name string
		role string
		want int
	}{
		{name: "student forbidden", role: "student", want: http.StatusForbidden},
		{name: "admin grant", role: "admin", want: http.StatusCreated},
	} {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/api/v1/admin/courses/"+testCourse+"/grants", bytes.NewReader(grantBody))
			request.Header.Set("Content-Type", "application/json")
			request.Header.Set("X-Test-Role", test.role)
			response := httptest.NewRecorder()
			server.ServeHTTP(response, request)
			if response.Code != test.want {
				t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
			}
			request.Body = io.NopCloser(bytes.NewReader(grantBody))
			assertContract(t, request, response)
		})
	}

	t.Run("Pay22 dependency failures are explicit", func(t *testing.T) {
		store.courseErr = domain.ErrDependencyUnavailable
		request := httptest.NewRequest(http.MethodPost, "/api/v1/learn/courses/"+testCourse+"/purchase", nil)
		request.Header.Set("Idempotency-Key", "purchase-2")
		request.Header.Set("X-Test-Role", "student")
		response := httptest.NewRecorder()
		server.ServeHTTP(response, request)
		if response.Code != http.StatusServiceUnavailable {
			t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
		}
		assertContract(t, request, response)
		store.courseErr = nil

		store.getErr = domain.ErrDependencyUnavailable
		request = httptest.NewRequest(http.MethodGet, "/api/v1/learn/orders/"+testOrder, nil)
		request.Header.Set("X-Test-Role", "student")
		response = httptest.NewRecorder()
		server.ServeHTTP(response, request)
		if response.Code != http.StatusServiceUnavailable {
			t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
		}
		assertContract(t, request, response)
		store.getErr = nil

		store.grantErr = domain.ErrDependencyUnavailable
		request = httptest.NewRequest(http.MethodPost, "/api/v1/admin/courses/"+testCourse+"/grants", bytes.NewReader(grantBody))
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("X-Test-Role", "admin")
		response = httptest.NewRecorder()
		server.ServeHTTP(response, request)
		if response.Code != http.StatusServiceUnavailable {
			t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
		}
		request.Body = io.NopCloser(bytes.NewReader(grantBody))
		assertContract(t, request, response)
	})

	t.Run("duplicate JSON keys are rejected", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPost, "/api/v1/admin/courses/"+testCourse+"/grants", strings.NewReader(`{"userId":"`+testUser+`","userId":"`+testUser+`","reason":"support resolution"}`))
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("X-Test-Role", "admin")
		response := httptest.NewRecorder()
		server.ServeHTTP(response, request)
		if response.Code != http.StatusBadRequest || !strings.Contains(response.Body.String(), "invalid_request") {
			t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
		}
	})
}

func TestCheckoutAlreadyOpenedMatchesOpenAPI(t *testing.T) {
	now := time.Date(2026, 9, 30, 12, 0, 0, 0, time.UTC)
	store := &fakeStore{}
	money, err := domain.NewMoney("50000", "UAH")
	if err != nil {
		t.Fatal(err)
	}
	order, err := domain.NewOrder(testOrder, testUser, testCourse, money, "purchase-1", now)
	if err != nil {
		t.Fatal(err)
	}
	store.order = order
	store.checkoutErr = domain.ErrCheckoutAlreadyOpened
	service := application.New(store, store, store, func() time.Time { return now }, func() string { return testGrant }, &fakePaymentProvider{})
	handler := ordershttp.NewHandler(service, nil, "https://courses.test")
	server := httpserver.NewHandler(httpserver.Options{Service: "test", Register: func(mux *http.ServeMux) {
		handler.Register(mux, func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				next.ServeHTTP(w, r.WithContext(httpserver.WithActor(r.Context(), httpserver.Actor{UserID: testUser, Role: "student"})))
			})
		}, func(next http.Handler) http.Handler { return next })
	}})
	request := httptest.NewRequest(http.MethodPost, "/api/v1/learn/orders/"+testOrder+"/checkout", nil)
	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)
	if response.Code != http.StatusConflict || !strings.Contains(response.Body.String(), `"code":"checkout_already_opened"`) {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	assertContract(t, request, response)
}

func assertContract(t *testing.T, request *http.Request, response *httptest.ResponseRecorder) {
	t.Helper()
	document, err := openapi3.NewLoader().LoadFromFile(filepath.Join("..", "..", "..", "..", "openapi", "openapi.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	router, err := legacy.NewRouter(document)
	if err != nil {
		t.Fatal(err)
	}
	route, params, err := router.FindRoute(request)
	if err != nil {
		t.Fatal(err)
	}
	input := &openapi3filter.RequestValidationInput{Request: request, PathParams: params, Route: route}
	if err := openapi3filter.ValidateRequest(context.Background(), input); err != nil {
		t.Fatalf("request contract: %v", err)
	}
	body := response.Body.Bytes()
	if response.Code == http.StatusNoContent {
		body = nil
	}
	if err := openapi3filter.ValidateResponse(context.Background(), &openapi3filter.ResponseValidationInput{RequestValidationInput: input, Status: response.Code, Header: response.Header(), Body: io.NopCloser(bytes.NewReader(body))}); err != nil {
		t.Fatalf("response contract: %v body=%s", err, response.Body.String())
	}
}

type fakeCallbackVerifier struct {
	callback ports.Callback
	err      error
	ackCalls int
}

func (verifier *fakeCallbackVerifier) VerifyCallback([]byte, time.Time) (ports.Callback, error) {
	return verifier.callback, verifier.err
}

func (verifier *fakeCallbackVerifier) Acknowledge(orderReference string, now time.Time) []byte {
	verifier.ackCalls++
	return []byte(`{"orderReference":"` + orderReference + `","status":"accept","time":` + fmt.Sprint(now.Unix()) + `,"signature":"0123456789abcdef0123456789abcdef"}`)
}

type fakePaymentProvider struct {
	lastURLs ports.CheckoutURLs
}

func (provider *fakePaymentProvider) Checkout(_ context.Context, _ domain.Order, _ ports.Buyer, urls ports.CheckoutURLs) (ports.CheckoutForm, error) {
	provider.lastURLs = urls
	return ports.CheckoutForm{Action: "https://secure.wayforpay.com/pay", Fields: []ports.CheckoutField{{Name: "merchantAccount", Value: "test"}}}, nil
}
func (*fakePaymentProvider) OrderStatus(context.Context, string) (ports.PaymentStatus, error) {
	return ports.PaymentStatus{}, nil
}

func callbackServer(t *testing.T, store *fakeStore, verifier *fakeCallbackVerifier, provider ports.PaymentProvider) http.Handler {
	t.Helper()
	service := application.New(store, store, store, func() time.Time { return time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC) }, func() string { return "40000000-0000-4000-8000-000000000099" }, provider)
	handler := ordershttp.NewHandler(service, verifier, "")
	authenticate := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r.WithContext(httpserver.WithActor(r.Context(), httpserver.Actor{UserID: testUser, Role: "student"})))
		})
	}
	return httpserver.NewHandler(httpserver.Options{Service: "test", Register: func(mux *http.ServeMux) {
		handler.Register(mux, authenticate, func(next http.Handler) http.Handler { return next })
	}})
}

func makeOrder(t *testing.T, status domain.OrderStatus) domain.Order {
	t.Helper()
	money, err := domain.NewMoney("50000", "UAH")
	if err != nil {
		t.Fatal(err)
	}
	created := time.Date(2026, 9, 30, 12, 0, 0, 0, time.UTC)
	var paidAt *time.Time
	if status == domain.OrderPaid {
		value := created.Add(time.Minute)
		paidAt = &value
	}
	order, err := domain.RestoreOrder(testOrder, testUser, testCourse, money, status, "purchase-1", created, paidAt)
	if err != nil {
		t.Fatal(err)
	}
	return order
}

func TestWayForPayCallbackHTTPContract(t *testing.T) {
	amount, err := domain.NewMoney("50000", "UAH")
	if err != nil {
		t.Fatal(err)
	}
	event := domain.PaymentEvent{Provider: "wayforpay", EventID: "order:event", OrderID: testOrder, Type: domain.PaymentSucceeded, Amount: amount, ReceivedAt: time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC)}
	for _, test := range []struct {
		name       string
		verifyErr  error
		event      *domain.PaymentEvent
		applyErr   error
		wantStatus int
		wantApply  int
		wantAck    int
	}{
		{name: "bad signature", verifyErr: errors.New("bad signature"), wantStatus: http.StatusUnauthorized},
		{name: "approved", event: &event, wantStatus: http.StatusOK, wantApply: 1, wantAck: 1},
		{name: "unmatched retries", event: func() *domain.PaymentEvent {
			copy := event
			copy.OrderID = "30000000-0000-4000-8000-000000000099"
			return &copy
		}(), applyErr: domain.ErrPaymentEventUnmatched, wantStatus: http.StatusServiceUnavailable, wantApply: 1},
		{name: "ignored status acknowledged", wantStatus: http.StatusOK, wantAck: 1},
	} {
		t.Run(test.name, func(t *testing.T) {
			store := &fakeStore{order: makeOrder(t, domain.OrderPending), applyErr: test.applyErr}
			verifier := &fakeCallbackVerifier{callback: ports.Callback{Event: test.event, OrderReference: testOrder}}
			if test.verifyErr != nil {
				verifier.err = test.verifyErr
			}
			server := callbackServer(t, store, verifier, nil)
			request := httptest.NewRequest(http.MethodPost, "/api/v1/payments/wayforpay/callback", bytes.NewReader([]byte(`{"signed":true}`)))
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()
			server.ServeHTTP(response, request)
			if response.Code != test.wantStatus {
				t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
			}
			if store.applyCalls != test.wantApply {
				t.Fatalf("apply calls=%d want=%d", store.applyCalls, test.wantApply)
			}
			if verifier.ackCalls != test.wantAck {
				t.Fatalf("ack calls=%d want=%d", verifier.ackCalls, test.wantAck)
			}
			request.Body = io.NopCloser(bytes.NewReader([]byte(`{"signed":true}`)))
			assertContract(t, request, response)
		})
	}
	t.Run("replay is acknowledged without a second grant", func(t *testing.T) {
		store := &fakeStore{order: makeOrder(t, domain.OrderPending)}
		verifier := &fakeCallbackVerifier{callback: ports.Callback{Event: &event, OrderReference: testOrder}}
		server := callbackServer(t, store, verifier, nil)
		for range 2 {
			request := httptest.NewRequest(http.MethodPost, "/api/v1/payments/wayforpay/callback", bytes.NewReader([]byte(`{"signed":true}`)))
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()
			server.ServeHTTP(response, request)
			if response.Code != http.StatusOK {
				t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
			}
		}
		if verifier.ackCalls != 2 || store.grantCalls != 0 {
			t.Fatalf("ack=%d grant calls=%d", verifier.ackCalls, store.grantCalls)
		}
	})
}

func TestCheckoutHTTPContract(t *testing.T) {
	for _, test := range []struct {
		name     string
		store    *fakeStore
		provider ports.PaymentProvider
		want     int
		code     string
	}{
		{name: "another user order", store: &fakeStore{getErr: domain.ErrOrderNotFound}, provider: &fakePaymentProvider{}, want: http.StatusNotFound},
		{name: "non-pending order", store: &fakeStore{order: makeOrder(t, domain.OrderPaid)}, provider: &fakePaymentProvider{}, want: http.StatusNotFound},
		{name: "provider unavailable", store: &fakeStore{order: makeOrder(t, domain.OrderPending)}, want: http.StatusServiceUnavailable, code: "dependency_unavailable"},
	} {
		t.Run(test.name, func(t *testing.T) {
			server := callbackServer(t, test.store, nil, test.provider)
			request := httptest.NewRequest(http.MethodPost, "/api/v1/learn/orders/"+testOrder+"/checkout", nil)
			response := httptest.NewRecorder()
			server.ServeHTTP(response, request)
			if response.Code != test.want {
				t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
			}
			if test.code != "" && !strings.Contains(response.Body.String(), test.code) {
				t.Fatalf("body=%s", response.Body.String())
			}
			assertContract(t, request, response)
		})
	}
}

func TestCheckoutReturnURLCarriesOnlyAValidCourseSlug(t *testing.T) {
	provider := &fakePaymentProvider{}
	server := callbackServer(t, &fakeStore{order: makeOrder(t, domain.OrderPending)}, nil, provider)
	for _, test := range []struct {
		name string
		body string
		want string
	}{
		{name: "valid", body: `{"courseSlug":"go-basics"}`, want: "/orders/" + testOrder + "?course=go-basics"},
		{name: "invalid", body: `{"courseSlug":"../private"}`, want: "/orders/" + testOrder},
		{name: "absent", want: "/orders/" + testOrder},
	} {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/api/v1/learn/orders/"+testOrder+"/checkout", strings.NewReader(test.body))
			if test.body != "" {
				request.Header.Set("Content-Type", "application/json")
			}
			response := httptest.NewRecorder()
			server.ServeHTTP(response, request)
			if response.Code != http.StatusOK {
				t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
			}
			if !strings.HasSuffix(provider.lastURLs.ReturnURL, test.want) {
				t.Fatalf("returnURL=%s want suffix=%s", provider.lastURLs.ReturnURL, test.want)
			}
		})
	}
	t.Run("duplicate courseSlug is rejected", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPost, "/api/v1/learn/orders/"+testOrder+"/checkout", strings.NewReader(`{"courseSlug":"go-basics","courseSlug":"other-course"}`))
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()
		server.ServeHTTP(response, request)
		if response.Code != http.StatusBadRequest || !strings.Contains(response.Body.String(), "invalid_request") {
			t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
		}
	})
}
