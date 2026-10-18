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
	order               domain.Order
	grant               domain.Grant
	courseErr           error
	getErr              error
	grantErr            error
	membership          *domain.Grant
	recipientIsAdmin    bool
	recipientIsAdminErr error
	applyCalls          int
	applyErr            error
	grantCalls          int
	checkoutErr         error
	enrollment          ports.EnrollmentFacts
	roster              ports.RosterPage
	rosterErr           error
	memberships         []ports.MembershipCard
	membershipsErr      error
	learnerPage         ports.LearnerPage
	learnerErr          error
	learnerFilter       ports.LearnerFilter
	authorUserID        string
	authorCourseID      string
	authorErr           error
	authorKind          string
	refundQueue         []ports.RefundQueueItem
	refundQueueErr      error
	refundQueueState    string
	orderByIDErr        error
	resolveErr          error
	resolvedOrderID     string
	resolvedActorID     string
	settledOutcome      string
}

func (store *fakeStore) CourseForPurchase(context.Context, string) (ports.CourseFacts, error) {
	if store.courseErr != nil {
		return ports.CourseFacts{}, store.courseErr
	}
	money, _ := domain.NewMoney("50000", "UAH")
	return ports.CourseFacts{Published: true, Price: &money}, nil
}
func (store *fakeStore) CourseForEnrollment(context.Context, string) (ports.EnrollmentFacts, error) {
	if store.enrollment.Status != "" {
		return store.enrollment, nil
	}
	return ports.EnrollmentFacts{Status: "published", Free: true, InstructorID: testAdmin}, nil
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
func (store *fakeStore) OrderByID(context.Context, string) (domain.Order, error) {
	if store.orderByIDErr != nil {
		return domain.Order{}, store.orderByIDErr
	}
	if store.order.ID() == "" {
		return domain.Order{}, domain.ErrOrderNotFound
	}
	return store.order, nil
}
func (store *fakeStore) RefundQueue(_ context.Context, state string) ([]ports.RefundQueueItem, error) {
	store.refundQueueState = state
	return store.refundQueue, store.refundQueueErr
}
func (store *fakeStore) SettleRefund(_ context.Context, orderID, actorID, outcome string) error {
	store.resolvedOrderID = orderID
	store.resolvedActorID = actorID
	store.settledOutcome = outcome
	return store.resolveErr
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
func (store *fakeStore) Membership(context.Context, string, string) (*domain.Grant, error) {
	return store.membership, nil
}
func (store *fakeStore) IsAdministrator(context.Context, string) (bool, error) {
	return store.recipientIsAdmin, store.recipientIsAdminErr
}
func (store *fakeStore) Learners(_ context.Context, filter ports.LearnerFilter, cursor string, limit int) (ports.LearnerPage, error) {
	store.learnerFilter = filter
	return store.learnerPage, store.learnerErr
}
func (store *fakeStore) AuthorOf(_ context.Context, kind, id string) (string, string, error) {
	store.authorKind = kind
	return store.authorUserID, store.authorCourseID, store.authorErr
}
func (store *fakeStore) MyMemberships(context.Context, string) ([]ports.MembershipCard, error) {
	return store.memberships, store.membershipsErr
}
func (store *fakeStore) Roster(context.Context, string, string, int) (ports.RosterPage, error) {
	return store.roster, store.rosterErr
}
func (store *fakeStore) CreateMembership(_ context.Context, grant domain.Grant, _ string, _ string, _ *string) (domain.Grant, error) {
	store.grantCalls++
	if store.grantErr != nil {
		return domain.Grant{}, store.grantErr
	}
	store.grant = grant
	return grant, nil
}
func (store *fakeStore) UpdateMembership(_ context.Context, grant domain.Grant, _ domain.GrantState, _ string, _ string, _ *string) (domain.Grant, error) {
	store.grantCalls++
	if store.grantErr != nil {
		return domain.Grant{}, store.grantErr
	}
	store.grant = grant
	return grant, nil
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
			if role == "admin" || role == "instructor" {
				userID = testAdmin
			}
			next.ServeHTTP(w, r.WithContext(httpserver.WithActor(r.Context(), httpserver.Actor{UserID: userID, Role: role, EmailVerified: true})))
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
	store.membership = &owned
	conflict := httptest.NewRequest(http.MethodPost, "/api/v1/learn/courses/"+testCourse+"/purchase", nil)
	conflict.Header.Set("Idempotency-Key", "purchase-owned")
	conflict.Header.Set("X-Test-Role", "student")
	response = httptest.NewRecorder()
	server.ServeHTTP(response, conflict)
	if response.Code != http.StatusConflict {
		t.Fatalf("owned status=%d body=%s", response.Code, response.Body.String())
	}
	assertContract(t, conflict, response)
	store.membership = nil

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

	t.Run("refund queue is administrator-only and lists flagged orders", func(t *testing.T) {
		charged := int64(50000)
		received := now.Add(time.Minute)
		activeState := domain.GrantActive
		store.refundQueue = []ports.RefundQueueItem{{OrderID: testOrder, CourseID: testCourse, UserID: testUser, OrderAmountMinor: 50000, ChargedAmountMinor: &charged, ReceivedAt: &received, Reason: "grant_conflict", MembershipState: &activeState}}
		defer func() { store.refundQueue = nil }()
		forbidden := httptest.NewRequest(http.MethodGet, "/api/v1/admin/refunds", nil)
		forbidden.Header.Set("X-Test-Role", "student")
		response := httptest.NewRecorder()
		server.ServeHTTP(response, forbidden)
		if response.Code != http.StatusForbidden {
			t.Fatalf("student refund queue status=%d body=%s", response.Code, response.Body.String())
		}
		assertContract(t, forbidden, response)

		request := httptest.NewRequest(http.MethodGet, "/api/v1/admin/refunds", nil)
		request.Header.Set("X-Test-Role", "admin")
		response = httptest.NewRecorder()
		server.ServeHTTP(response, request)
		if response.Code != http.StatusOK ||
			!strings.Contains(response.Body.String(), `"reason":"grant_conflict"`) ||
			!strings.Contains(response.Body.String(), `"membershipState":"active"`) {
			t.Fatalf("admin refund queue status=%d body=%s", response.Code, response.Body.String())
		}
		assertContract(t, request, response)
	})

	t.Run("refund queue rejects an unknown state", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/api/v1/admin/refunds?state=bogus", nil)
		request.Header.Set("X-Test-Role", "admin")
		response := httptest.NewRecorder()
		server.ServeHTTP(response, request)
		if response.Code != http.StatusBadRequest {
			t.Fatalf("bogus state status=%d body=%s", response.Code, response.Body.String())
		}
		// No assertContract: the request intentionally violates the state enum, so it cannot
		// conform to the spec — only the 400 response matters here.
	})

	t.Run("refund queue threads the state filter and surfaces the outcome", func(t *testing.T) {
		store.refundQueue = []ports.RefundQueueItem{{OrderID: testOrder, CourseID: testCourse, UserID: testUser, OrderAmountMinor: 50000, Reason: "grant_conflict", Outcome: "declined"}}
		defer func() { store.refundQueue = nil }()
		request := httptest.NewRequest(http.MethodGet, "/api/v1/admin/refunds?state=declined", nil)
		request.Header.Set("X-Test-Role", "admin")
		response := httptest.NewRecorder()
		server.ServeHTTP(response, request)
		if response.Code != http.StatusOK || store.refundQueueState != "declined" ||
			!strings.Contains(response.Body.String(), `"outcome":"declined"`) {
			t.Fatalf("declined queue status=%d state=%q body=%s", response.Code, store.refundQueueState, response.Body.String())
		}
		assertContract(t, request, response)
	})

	for _, settle := range []struct {
		path    string
		outcome string
	}{{"refund", "refunded"}, {"decline", "declined"}} {
		t.Run("refund "+settle.path+" is administrator-only and idempotent-ack", func(t *testing.T) {
			store.resolvedOrderID, store.resolvedActorID, store.settledOutcome = "", "", ""
			forbidden := httptest.NewRequest(http.MethodPost, "/api/v1/admin/refunds/"+testOrder+"/"+settle.path, nil)
			forbidden.Header.Set("X-Test-Role", "student")
			response := httptest.NewRecorder()
			server.ServeHTTP(response, forbidden)
			if response.Code != http.StatusForbidden {
				t.Fatalf("student %s status=%d body=%s", settle.path, response.Code, response.Body.String())
			}
			assertContract(t, forbidden, response)

			request := httptest.NewRequest(http.MethodPost, "/api/v1/admin/refunds/"+testOrder+"/"+settle.path, nil)
			request.Header.Set("X-Test-Role", "admin")
			response = httptest.NewRecorder()
			server.ServeHTTP(response, request)
			if response.Code != http.StatusNoContent || store.resolvedOrderID != testOrder || store.resolvedActorID != testAdmin || store.settledOutcome != settle.outcome {
				t.Fatalf("admin %s status=%d order=%q actor=%q outcome=%q", settle.path, response.Code, store.resolvedOrderID, store.resolvedActorID, store.settledOutcome)
			}
			assertContract(t, request, response)
		})
	}

	t.Run("refund revoke reuses the revoke transition and maps not-found", func(t *testing.T) {
		adminID := testAdmin
		owned, err := domain.NewGrant(testGrant, testUser, testCourse, domain.GrantPurchase, &[]string{testOrder}[0], domain.GrantActive, now, &adminID)
		if err != nil {
			t.Fatal(err)
		}
		store.membership = &owned
		defer func() { store.membership = nil }()
		request := httptest.NewRequest(http.MethodPost, "/api/v1/admin/refunds/"+testOrder+"/revoke", nil)
		request.Header.Set("X-Test-Role", "admin")
		response := httptest.NewRecorder()
		server.ServeHTTP(response, request)
		if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"state":"revoked"`) {
			t.Fatalf("admin refund revoke status=%d body=%s", response.Code, response.Body.String())
		}
		assertContract(t, request, response)

		forbidden := httptest.NewRequest(http.MethodPost, "/api/v1/admin/refunds/"+testOrder+"/revoke", nil)
		forbidden.Header.Set("X-Test-Role", "student")
		response = httptest.NewRecorder()
		server.ServeHTTP(response, forbidden)
		if response.Code != http.StatusForbidden {
			t.Fatalf("student refund revoke status=%d body=%s", response.Code, response.Body.String())
		}
		assertContract(t, forbidden, response)

		store.orderByIDErr = domain.ErrOrderNotFound
		defer func() { store.orderByIDErr = nil }()
		missing := httptest.NewRequest(http.MethodPost, "/api/v1/admin/refunds/"+testOrder+"/revoke", nil)
		missing.Header.Set("X-Test-Role", "admin")
		response = httptest.NewRecorder()
		server.ServeHTTP(response, missing)
		if response.Code != http.StatusNotFound || !strings.Contains(response.Body.String(), `"code":"order_not_found"`) {
			t.Fatalf("missing order refund revoke status=%d body=%s", response.Code, response.Body.String())
		}
		assertContract(t, missing, response)
	})

	t.Run("admin cannot grant themselves", func(t *testing.T) {
		body := []byte(`{"userId":"` + testAdmin + `","reason":"support resolution"}`)
		request := httptest.NewRequest(http.MethodPost, "/api/v1/admin/courses/"+testCourse+"/grants", bytes.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("X-Test-Role", "admin")
		response := httptest.NewRecorder()
		server.ServeHTTP(response, request)
		if response.Code != http.StatusForbidden || !strings.Contains(response.Body.String(), `"code":"grant_self_forbidden"`) {
			t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
		}
		request.Body = io.NopCloser(bytes.NewReader(body))
		assertContract(t, request, response)
	})

	t.Run("Lo18 an administrator recipient cannot be granted a course", func(t *testing.T) {
		store.recipientIsAdmin = true
		defer func() { store.recipientIsAdmin = false }()
		body := []byte(`{"userId":"` + testUser + `","reason":"support resolution"}`)
		request := httptest.NewRequest(http.MethodPost, "/api/v1/admin/courses/"+testCourse+"/grants", bytes.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("X-Test-Role", "admin")
		response := httptest.NewRecorder()
		server.ServeHTTP(response, request)
		if response.Code != http.StatusForbidden || !strings.Contains(response.Body.String(), `"code":"admin_not_learner"`) {
			t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
		}
		request.Body = io.NopCloser(bytes.NewReader(body))
		assertContract(t, request, response)
	})

	t.Run("Lo18 an administrator cannot start a purchase", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPost, "/api/v1/learn/courses/"+testCourse+"/purchase", nil)
		request.Header.Set("Idempotency-Key", "purchase-admin")
		request.Header.Set("X-Test-Role", "admin")
		response := httptest.NewRecorder()
		server.ServeHTTP(response, request)
		if response.Code != http.StatusForbidden || !strings.Contains(response.Body.String(), `"code":"admin_not_learner"`) {
			t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
		}
		assertContract(t, request, response)
	})

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
				next.ServeHTTP(w, r.WithContext(httpserver.WithActor(r.Context(), httpserver.Actor{UserID: testUser, Role: "student", EmailVerified: true})))
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

func TestEnrollmentHTTPMatchesOpenAPI(t *testing.T) {
	now := time.Date(2026, 10, 7, 12, 0, 0, 0, time.UTC)
	store := &fakeStore{roster: ports.RosterPage{Entries: []ports.RosterEntry{{UserID: testUser, DisplayName: "Student", State: domain.GrantActive, EnrolledAt: now, ProgressPercent: 35}}}}
	service := application.New(store, store, store, func() time.Time { return now }, func() string { return testGrant }, nil)
	handler := ordershttp.NewHandler(service, nil, "")
	authenticate := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role := r.Header.Get("X-Test-Role")
			userID := testUser
			if role == "admin" || role == "instructor" {
				userID = testAdmin
			}
			next.ServeHTTP(w, r.WithContext(httpserver.WithActor(r.Context(), httpserver.Actor{UserID: userID, Role: role, EmailVerified: r.Header.Get("X-Unverified") == ""})))
		})
	}
	server := httpserver.NewHandler(httpserver.Options{Service: "test", Register: func(mux *http.ServeMux) {
		handler.Register(mux, authenticate, func(next http.Handler) http.Handler { return next })
		handler.RegisterTeach(mux, authenticate, func(next http.Handler) http.Handler { return next })
		handler.RegisterAdmin(mux, authenticate, func(next http.Handler) http.Handler { return next }, func(next http.Handler) http.Handler { return next })
	}})
	request := func(method, path, role string) (*http.Request, *httptest.ResponseRecorder) {
		r := httptest.NewRequest(method, path, nil)
		r.Header.Set("X-Test-Role", role)
		response := httptest.NewRecorder()
		server.ServeHTTP(response, r)
		return r, response
	}

	path := "/api/v1/learn/courses/" + testCourse
	r, response := request(http.MethodPost, path+"/enroll", "student")
	if response.Code != http.StatusCreated || !strings.Contains(response.Body.String(), `"source":"self"`) {
		t.Fatalf("enroll status=%d body=%s", response.Code, response.Body.String())
	}
	assertContract(t, r, response)

	// Lo18 (G8): an administrator is not a learner and cannot self-enrol.
	r, response = request(http.MethodPost, path+"/enroll", "admin")
	if response.Code != http.StatusForbidden || !strings.Contains(response.Body.String(), `"code":"admin_not_learner"`) {
		t.Fatalf("admin enroll status=%d body=%s", response.Code, response.Body.String())
	}
	assertContract(t, r, response)

	store.membership = &store.grant
	r, response = request(http.MethodGet, path+"/membership", "student")
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"member":true`) {
		t.Fatalf("membership status=%d body=%s", response.Code, response.Body.String())
	}
	assertContract(t, r, response)

	store.membership = nil
	r, response = request(http.MethodGet, path+"/membership", "student")
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"state":null`) {
		t.Fatalf("none membership status=%d body=%s", response.Code, response.Body.String())
	}
	assertContract(t, r, response)

	store.enrollment = ports.EnrollmentFacts{Status: "published", Free: false, InstructorID: testAdmin}
	r, response = request(http.MethodPost, path+"/enroll", "student")
	if response.Code != http.StatusConflict || !strings.Contains(response.Body.String(), "course_not_enrollable") {
		t.Fatalf("paid enroll status=%d body=%s", response.Code, response.Body.String())
	}
	assertContract(t, r, response)
	store.enrollment = ports.EnrollmentFacts{}

	r, response = request(http.MethodPost, path+"/enroll", "student")
	r.Header.Set("X-Unverified", "true")
	response = httptest.NewRecorder()
	server.ServeHTTP(response, r)
	if response.Code != http.StatusForbidden || !strings.Contains(response.Body.String(), "email_not_verified") {
		t.Fatalf("unverified enroll status=%d body=%s", response.Code, response.Body.String())
	}
	assertContract(t, r, response)

	store.membership = nil
	store.grantErr = domain.ErrMembershipConflict
	r, response = request(http.MethodPost, path+"/enroll", "student")
	if response.Code != http.StatusConflict || !strings.Contains(response.Body.String(), "membership_conflict") {
		t.Fatalf("conflict enroll status=%d body=%s", response.Code, response.Body.String())
	}
	assertContract(t, r, response)
	store.grantErr = nil

	store.membership = nil
	participant := "/participants/" + testUser
	r, response = request(http.MethodPost, "/api/v1/teach/courses/"+testCourse+participant+"/block", "student")
	if response.Code != http.StatusForbidden || !strings.Contains(response.Body.String(), "enrollment_forbidden") {
		t.Fatalf("stranger block status=%d body=%s", response.Code, response.Body.String())
	}
	assertContract(t, r, response)

	// Blocking a non-member is an invalid transition, not a server error: 409 grant_transition_invalid.
	r, response = request(http.MethodPost, "/api/v1/teach/courses/"+testCourse+participant+"/block", "instructor")
	if response.Code != http.StatusConflict || !strings.Contains(response.Body.String(), `"code":"grant_transition_invalid"`) {
		t.Fatalf("non-member block status=%d body=%s", response.Code, response.Body.String())
	}
	assertContract(t, r, response)

	active, err := domain.SelfEnroll(testGrant, testUser, testCourse, now)
	if err != nil {
		t.Fatal(err)
	}
	store.membership = &active
	r, response = request(http.MethodPost, "/api/v1/teach/courses/"+testCourse+participant+"/block", "instructor")
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"state":"blocked"`) {
		t.Fatalf("block status=%d body=%s", response.Code, response.Body.String())
	}
	assertContract(t, r, response)
	store.membership = &store.grant
	r, response = request(http.MethodPost, "/api/v1/teach/courses/"+testCourse+participant+"/unblock", "instructor")
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"state":"active"`) {
		t.Fatalf("unblock status=%d body=%s", response.Code, response.Body.String())
	}
	assertContract(t, r, response)
	store.membership = &store.grant
	store.memberships = []ports.MembershipCard{{CourseID: testCourse, Slug: "course", Title: "Course", State: domain.GrantActive, EnrolledAt: now, ProgressPercent: 35}}
	r, response = request(http.MethodGet, "/api/v1/learn/memberships", "student")
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"items":[{"courseId":"`+testCourse+`"`) || !strings.Contains(response.Body.String(), `"progressPercent":35`) {
		t.Fatalf("memberships status=%d body=%s", response.Code, response.Body.String())
	}
	assertContract(t, r, response)

	store.memberships = nil
	r, response = request(http.MethodGet, "/api/v1/learn/memberships", "student")
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"items":[]`) {
		t.Fatalf("empty memberships status=%d body=%s", response.Code, response.Body.String())
	}
	assertContract(t, r, response)

	r, response = request(http.MethodPost, "/api/v1/admin/courses/"+testCourse+participant+"/revoke", "student")
	if response.Code != http.StatusForbidden || !strings.Contains(response.Body.String(), "enrollment_forbidden") {
		t.Fatalf("revoke status=%d body=%s", response.Code, response.Body.String())
	}
	assertContract(t, r, response)
	r, response = request(http.MethodPost, "/api/v1/admin/courses/"+testCourse+participant+"/revoke", "admin")
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"state":"revoked"`) {
		t.Fatalf("revoke status=%d body=%s", response.Code, response.Body.String())
	}
	assertContract(t, r, response)

	r, response = request(http.MethodGet, "/api/v1/teach/courses/"+testCourse+"/roster?limit=20", "instructor")
	if response.Code != http.StatusOK || strings.Contains(response.Body.String(), "email") {
		t.Fatalf("roster status=%d body=%s", response.Code, response.Body.String())
	}
	assertContract(t, r, response)
	store.rosterErr = domain.ErrInvalidCursor
	r, response = request(http.MethodGet, "/api/v1/teach/courses/"+testCourse+"/roster?cursor=invalid", "instructor")
	if response.Code != http.StatusBadRequest {
		t.Fatalf("cursor status=%d body=%s", response.Code, response.Body.String())
	}
	assertContract(t, r, response)

	// Lo07/Lo02: the learner overview returns the page shape with no email, and a row carries its
	// courses and last activity. assertContract validates the body against LearnerPage.
	lastActivity := now
	nextCursor := "next"
	store.learnerPage = ports.LearnerPage{
		Entries: []ports.LearnerRow{{
			UserID:         testUser,
			DisplayName:    "Ada",
			LastActivityAt: &lastActivity,
			Courses:        []ports.LearnerCourse{{CourseID: testCourse, Slug: "course", Title: "Course", State: domain.GrantActive, ProgressPercent: 35, EnrolledAt: now}},
		}},
		NextCursor: &nextCursor,
	}
	for _, path := range []string{"/api/v1/teach/learners", "/api/v1/admin/learners"} {
		role := "instructor"
		if strings.Contains(path, "admin") {
			role = "admin"
		}
		r, response = request(http.MethodGet, path+"?limit=20", role)
		if response.Code != http.StatusOK || strings.Contains(response.Body.String(), "email") || !strings.Contains(response.Body.String(), `"displayName":"Ada"`) || !strings.Contains(response.Body.String(), `"lastActivityAt":"`) {
			t.Fatalf("%s status=%d body=%s", path, response.Code, response.Body.String())
		}
		assertContract(t, r, response)
	}

	// Lo05: invalid query parameters are rejected before the service is reached. (The response is
	// not contract-checked here: the spec validator rejects these as malformed requests, which is
	// exactly what the handler is also doing.)
	for _, bad := range []string{"state=paused", "limit=0", "limit=200", "q=", "unknown=1"} {
		_, response = request(http.MethodGet, "/api/v1/teach/learners?"+bad, "instructor")
		if response.Code != http.StatusBadRequest {
			t.Fatalf("bad query %q status=%d body=%s", bad, response.Code, response.Body.String())
		}
	}

	// Lo05: filtering by a course the instructor does not teach is refused by name, not revealed.
	store.enrollment = ports.EnrollmentFacts{Status: "published", Free: true, InstructorID: testUser}
	r, response = request(http.MethodGet, "/api/v1/teach/learners?courseId="+testCourse, "instructor")
	if response.Code != http.StatusForbidden || !strings.Contains(response.Body.String(), `"code":"learner_course_forbidden"`) {
		t.Fatalf("foreign course status=%d body=%s", response.Code, response.Body.String())
	}
	assertContract(t, r, response)
	store.enrollment = ports.EnrollmentFacts{}

	// Lo08: a student cannot read the administrator overview.
	r, response = request(http.MethodGet, "/api/v1/admin/learners", "student")
	if response.Code != http.StatusForbidden {
		t.Fatalf("student admin overview status=%d body=%s", response.Code, response.Body.String())
	}
	assertContract(t, r, response)

	// Lo12/Lo13/Lo15: block/unblock an item's author from the queue returns the membership grant;
	// an unknown comment or review is a named 404.
	store.authorUserID, store.authorCourseID = testUser, testCourse
	blocked, _ := domain.SelfEnroll(testGrant, testUser, testCourse, now)
	store.membership = &blocked
	r, response = request(http.MethodPost, "/api/v1/teach/comments/"+testGrant+"/block-author", "admin")
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"state":"blocked"`) {
		t.Fatalf("block comment author status=%d body=%s", response.Code, response.Body.String())
	}
	if store.authorKind != "comment" {
		t.Fatalf("AuthorOf kind=%q", store.authorKind)
	}
	assertContract(t, r, response)

	member, _ := domain.SelfEnroll(testGrant, testUser, testCourse, now)
	blockedMember, _ := member.Block()
	store.membership = &blockedMember
	r, response = request(http.MethodPost, "/api/v1/teach/reviews/"+testGrant+"/unblock-author", "admin")
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"state":"active"`) {
		t.Fatalf("unblock review author status=%d body=%s", response.Code, response.Body.String())
	}
	assertContract(t, r, response)

	store.authorErr = domain.ErrCommentNotFound
	r, response = request(http.MethodPost, "/api/v1/teach/comments/"+testGrant+"/block-author", "admin")
	if response.Code != http.StatusNotFound || !strings.Contains(response.Body.String(), `"code":"comment_not_found"`) {
		t.Fatalf("unknown comment status=%d body=%s", response.Code, response.Body.String())
	}
	assertContract(t, r, response)

	store.authorErr = domain.ErrReviewNotFound
	r, response = request(http.MethodPost, "/api/v1/teach/reviews/"+testGrant+"/block-author", "admin")
	if response.Code != http.StatusNotFound || !strings.Contains(response.Body.String(), `"code":"review_not_found"`) {
		t.Fatalf("unknown review status=%d body=%s", response.Code, response.Body.String())
	}
	assertContract(t, r, response)
	store.authorErr = nil
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
			next.ServeHTTP(w, r.WithContext(httpserver.WithActor(r.Context(), httpserver.Actor{UserID: testUser, Role: "student", EmailVerified: true})))
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
