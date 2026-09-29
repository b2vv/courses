package http_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
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
	order     domain.Order
	grant     domain.Grant
	courseErr error
	getErr    error
	grantErr  error
	active    []domain.Grant
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
func (store *fakeStore) ActiveGrants(context.Context, string, string) ([]domain.Grant, error) {
	return store.active, nil
}
func (store *fakeStore) GrantWithAudit(_ context.Context, command ports.GrantCommand) (domain.Grant, error) {
	if store.grantErr != nil {
		return domain.Grant{}, store.grantErr
	}
	store.grant = command.Grant
	return command.Grant, nil
}
func (*fakeStore) GrandfatherOnPriceSet(context.Context, string, time.Time, domain.Money, time.Time) error {
	return nil
}
func (*fakeStore) Apply(context.Context, domain.PaymentEvent, func() string) error { return nil }

func TestPurchaseAndGrantHTTPMatchOpenAPI(t *testing.T) {
	now := time.Date(2026, 9, 30, 12, 0, 0, 0, time.UTC)
	store := &fakeStore{}
	ids := []string{testOrder, testGrant, "40000000-0000-4000-8000-000000000002"}
	service := application.New(store, store, store, func() time.Time { return now }, func() string {
		id := ids[0]
		ids = ids[1:]
		return id
	})
	handler := ordershttp.NewHandler(service)
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
