package http_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	adminstatshttp "courses/internal/adminstats/adapters/http"
	"courses/internal/adminstats/application"
	"courses/internal/adminstats/domain"
	"courses/platform/httpserver"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/legacy"
)

const testInstructorID = "10000000-0000-4000-8000-000000000009"

type fakeRepository struct {
	snapshot  domain.Snapshot
	points    []domain.TimeseriesPoint
	snapErr   error
	tsErr     error
	lastFrom  time.Time
	lastTo    time.Time
	lastMetr  domain.Metric
	lastBuck  domain.Bucket
	tsQueried bool
}

func (repository *fakeRepository) Snapshot(context.Context) (domain.Snapshot, error) {
	return repository.snapshot, repository.snapErr
}

func (repository *fakeRepository) Timeseries(_ context.Context, metric domain.Metric, bucket domain.Bucket, from, to time.Time) ([]domain.TimeseriesPoint, error) {
	repository.tsQueried = true
	repository.lastMetr, repository.lastBuck, repository.lastFrom, repository.lastTo = metric, bucket, from, to
	return repository.points, repository.tsErr
}

func sampleSnapshot() domain.Snapshot {
	avg := 123.5
	return domain.Snapshot{
		Users:   domain.UsersStat{Active: 5, Blocked: 1, Deleted: 2, ByRole: domain.RoleStat{Student: 3, Instructor: 2, Admin: 1}, Verified: 4, TwoFactor: 1},
		Courses: domain.CoursesStat{Draft: 1, Published: 2, Archived: 1},
		Grants:  domain.GrantsStat{Active: 3, Blocked: 1, Revoked: 1},
		Orders:  domain.OrdersStat{Pending: 1, Paid: 2, Failed: 1, RevenueMinorUAH: 12000},
		Reports: domain.ReportsStat{Open: 1, Resolved: 2},
		Media:   domain.MediaStat{InProgress: 1, Ready: 2, Failed: 1, TotalDurationSeconds: 247, AvgDurationSeconds: &avg, WithDurationCount: 2},
		Instructors: []domain.InstructorStat{
			{InstructorID: testInstructorID, DisplayName: "Ada", PublishedCourses: 2, ActiveEnrollments: 3, RevenueMinorUAH: 12000},
		},
	}
}

// newServer wires the real httpserver around the adminstats handler with stub
// authenticate/administrator middlewares driven by X-Test-Role: an empty role
// is a guest (401), a non-admin role is 403 grant_forbidden, admin passes.
func newServer(t *testing.T, repository *fakeRepository, singleInstructor bool) http.Handler {
	t.Helper()
	fixedNow := func() time.Time { return time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC) }
	handler := adminstatshttp.NewHandler(application.New(repository, singleInstructor, fixedNow))
	authenticate := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Mirror the real identity Authenticate, which sets the no-store
			// header first so every downstream response (including a 403 from
			// the administrator gate) carries it.
			w.Header().Set("Cache-Control", "no-store")
			role := r.Header.Get("X-Test-Role")
			if role == "" {
				httpserver.WriteError(w, r, http.StatusUnauthorized, "authentication_required", "Authentication is required")
				return
			}
			next.ServeHTTP(w, r.WithContext(httpserver.WithActor(r.Context(), httpserver.Actor{UserID: testInstructorID, Role: role, EmailVerified: true})))
		})
	}
	administrator := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("X-Test-Role") != "admin" {
				httpserver.WriteError(w, r, http.StatusForbidden, "grant_forbidden", "Administrator permission is required")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
	return httpserver.NewHandler(httpserver.Options{Service: "test", Register: func(mux *http.ServeMux) {
		handler.RegisterAdmin(mux, authenticate, administrator)
	}})
}

func request(t *testing.T, server http.Handler, target, role string) (*http.Request, *httptest.ResponseRecorder) {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, target, nil)
	if role != "" {
		req.Header.Set("X-Test-Role", role)
	}
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)
	return req, rec
}

func TestSnapshotEndpoint(t *testing.T) {
	repository := &fakeRepository{snapshot: sampleSnapshot()}
	server := newServer(t, repository, true)

	req, rec := request(t, server, "/api/v1/admin/stats", "admin")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", rec.Code, rec.Body.String())
	}
	if rec.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("Cache-Control = %q, want no-store", rec.Header().Get("Cache-Control"))
	}
	// The platform flag the service stamped must reach the payload.
	if !strings.Contains(rec.Body.String(), `"singleInstructor":true`) {
		t.Fatalf("payload missing singleInstructor flag: %s", rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"revenueMinorUAH":12000`) {
		t.Fatalf("payload missing revenue: %s", rec.Body.String())
	}
	assertContract(t, req, rec)
}

func TestSnapshotEmptyLeaderboardIsArray(t *testing.T) {
	repository := &fakeRepository{snapshot: domain.Snapshot{}} // no instructors, nil avg
	server := newServer(t, repository, false)

	req, rec := request(t, server, "/api/v1/admin/stats", "admin")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", rec.Code, rec.Body.String())
	}
	// Required array field renders as [] (not null) on an empty leaderboard,
	// and the nullable average renders as null (not 0).
	if !strings.Contains(rec.Body.String(), `"instructors":[]`) {
		t.Fatalf("empty leaderboard not []: %s", rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"avgDurationSeconds":null`) {
		t.Fatalf("nil average not null: %s", rec.Body.String())
	}
	assertContract(t, req, rec)
}

func TestSnapshotAuthorization(t *testing.T) {
	cases := []struct {
		name string
		role string
		want int
	}{
		{"guest unauthenticated", "", http.StatusUnauthorized},
		{"student forbidden", "student", http.StatusForbidden},
		{"instructor forbidden", "instructor", http.StatusForbidden},
		{"admin ok", "admin", http.StatusOK},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			repository := &fakeRepository{snapshot: sampleSnapshot()}
			server := newServer(t, repository, true)
			req, rec := request(t, server, "/api/v1/admin/stats", test.role)
			if rec.Code != test.want {
				t.Fatalf("status = %d, want %d: %s", rec.Code, test.want, rec.Body.String())
			}
			if test.want == http.StatusForbidden && !strings.Contains(rec.Body.String(), `"code":"grant_forbidden"`) {
				t.Fatalf("forbidden body missing grant_forbidden: %s", rec.Body.String())
			}
			assertContract(t, req, rec)
		})
	}
}

func TestTimeseriesEndpoint(t *testing.T) {
	repository := &fakeRepository{points: []domain.TimeseriesPoint{
		{BucketStart: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC), Value: 4},
		{BucketStart: time.Date(2026, 3, 2, 0, 0, 0, 0, time.UTC), Value: 7},
	}}
	server := newServer(t, repository, true)

	req, rec := request(t, server, "/api/v1/admin/stats/timeseries?metric=signups&bucket=day", "admin")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", rec.Code, rec.Body.String())
	}
	if rec.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("Cache-Control = %q, want no-store", rec.Header().Get("Cache-Control"))
	}
	if !repository.tsQueried || repository.lastMetr != domain.MetricSignups || repository.lastBuck != domain.BucketDay {
		t.Fatalf("service did not pass parsed metric/bucket to repo: %+v", repository)
	}
	if !strings.Contains(rec.Body.String(), `"bucketStart":"2026-03-01T00:00:00Z"`) {
		t.Fatalf("payload missing formatted bucket: %s", rec.Body.String())
	}
	assertContract(t, req, rec)
}

func TestTimeseriesEmptyIsArray(t *testing.T) {
	repository := &fakeRepository{points: nil}
	server := newServer(t, repository, true)
	req, rec := request(t, server, "/api/v1/admin/stats/timeseries?metric=orders&bucket=week", "admin")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", rec.Code, rec.Body.String())
	}
	if strings.TrimSpace(rec.Body.String()) != "[]" {
		t.Fatalf("empty series not []: %q", rec.Body.String())
	}
	assertContract(t, req, rec)
}

func TestTimeseriesRangeErrors(t *testing.T) {
	cases := []struct {
		name   string
		query  string
		code   string
		assert bool // whether the request itself is contract-valid
	}{
		// Valid enum metric/bucket; the window is what is wrong, so the
		// request conforms to the contract and the 400 response does too.
		{"inverted range", "metric=signups&bucket=day&from=2026-05-01T00:00:00Z&to=2026-01-01T00:00:00Z", "invalid_range", true},
		// from=banana is not a date-time, so the request is out-of-contract;
		// the handler must still reject it (production does not validate
		// requests against OpenAPI), but assertContract is skipped.
		{"unparseable from", "metric=signups&bucket=day&from=banana", "invalid_range", false},
		{"too wide", "metric=signups&bucket=day&from=2020-01-01T00:00:00Z&to=2026-01-01T00:00:00Z", "range_too_wide", true},
		// Deliberately out-of-contract request values: production does not
		// OpenAPI-validate requests, so the handler must still reject them with
		// a business code. assertContract is skipped (the request violates the
		// request-side enum on purpose).
		{"unknown metric", "metric=bogus&bucket=day", "unknown_metric", false},
		{"unknown bucket", "metric=signups&bucket=fortnight", "unknown_bucket", false},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			repository := &fakeRepository{}
			server := newServer(t, repository, true)
			req, rec := request(t, server, "/api/v1/admin/stats/timeseries?"+test.query, "admin")
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400: %s", rec.Code, rec.Body.String())
			}
			if !strings.Contains(rec.Body.String(), `"code":"`+test.code+`"`) {
				t.Fatalf("body missing code %q: %s", test.code, rec.Body.String())
			}
			if repository.tsQueried {
				t.Fatalf("repo must not be queried on a rejected range (%s)", test.name)
			}
			if test.assert {
				assertContract(t, req, rec)
			}
		})
	}
}

func TestTimeseriesAuthorization(t *testing.T) {
	repository := &fakeRepository{}
	server := newServer(t, repository, true)
	_, rec := request(t, server, "/api/v1/admin/stats/timeseries?metric=signups&bucket=day", "student")
	if rec.Code != http.StatusForbidden || !strings.Contains(rec.Body.String(), `"code":"grant_forbidden"`) {
		t.Fatalf("student timeseries = %d: %s", rec.Code, rec.Body.String())
	}
}

func assertContract(t *testing.T, req *http.Request, rec *httptest.ResponseRecorder) {
	t.Helper()
	document, err := openapi3.NewLoader().LoadFromFile(filepath.Join("..", "..", "..", "..", "openapi", "openapi.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	router, err := legacy.NewRouter(document)
	if err != nil {
		t.Fatal(err)
	}
	route, params, err := router.FindRoute(req)
	if err != nil {
		t.Fatal(err)
	}
	input := &openapi3filter.RequestValidationInput{Request: req, PathParams: params, Route: route}
	if err := openapi3filter.ValidateRequest(context.Background(), input); err != nil {
		t.Fatalf("request contract: %v", err)
	}
	if err := openapi3filter.ValidateResponse(context.Background(), &openapi3filter.ResponseValidationInput{
		RequestValidationInput: input, Status: rec.Code, Header: rec.Header(), Body: io.NopCloser(bytes.NewReader(rec.Body.Bytes())),
	}); err != nil {
		t.Fatalf("response contract: %v body=%s", err, rec.Body.String())
	}
}
