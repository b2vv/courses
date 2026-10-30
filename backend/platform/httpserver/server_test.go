package httpserver_test

import (
	"bytes"
	"context"
	"courses/platform/observability"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"courses/platform/httpserver"
)

func TestHealthAndReadiness(t *testing.T) {
	handler := httpserver.NewHandler(httpserver.Options{
		Service: "api",
		Ready:   func(context.Context) error { return nil },
	})
	for _, path := range []string{"/healthz", "/readyz"} {
		t.Run(path, func(t *testing.T) {
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
			if response.Code != http.StatusOK {
				t.Fatalf("got %d: %s", response.Code, response.Body.String())
			}
			if response.Header().Get("X-Request-ID") == "" {
				t.Fatal("missing request ID")
			}
			if got := response.Header().Get("Content-Security-Policy"); got != "default-src 'none'; frame-ancestors 'none'" {
				t.Fatalf("Content-Security-Policy=%q", got)
			}
		})
	}
}

func TestRequestObservabilityExcludesSecrets(t *testing.T) {
	var logs bytes.Buffer
	metrics := observability.NewMetrics()
	tracing, err := observability.NewTracing(context.Background(), "test")
	if err != nil {
		t.Fatal(err)
	}
	defer tracing.Shutdown(context.Background())
	handler := httpserver.NewHandler(httpserver.Options{
		Logger: slog.New(slog.NewJSONHandler(&logs, nil)), Metrics: metrics, Tracer: tracing.Tracer("http"),
		Ready: func(context.Context) error { return nil },
	})
	r := httptest.NewRequest("GET", "/unknown-secret?token=query-secret", nil)
	r.Header.Set("Cookie", "session=cookie-secret")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	for _, secret := range []string{"unknown-secret", "query-secret", "cookie-secret"} {
		if strings.Contains(logs.String(), secret) {
			t.Fatalf("secret in logs: %s", logs.String())
		}
	}
	var entry map[string]interface{}
	if err := json.Unmarshal(logs.Bytes(), &entry); err != nil {
		t.Fatal(err)
	}
	if entry["requestId"] != w.Header().Get("X-Request-ID") || entry["traceId"] == "" || entry["status"] != float64(404) {
		t.Fatalf("missing correlation: %+v", entry)
	}
	metricsResponse := httptest.NewRecorder()
	metrics.Handler().ServeHTTP(metricsResponse, httptest.NewRequest("GET", "/metrics", nil))
	if !strings.Contains(metricsResponse.Body.String(), `courses_http_requests_total{method="GET",route="unmatched",status="404"} 1`) {
		t.Fatal("request counter absent")
	}
}

func TestMetricsTokenGuard(t *testing.T) {
	// Token unset: /metrics stays open, as local and test runs expect (Prometheus
	// scrapes the service port directly with no credentials).
	open := httpserver.NewHandler(httpserver.Options{Service: "api", Metrics: observability.NewMetrics()})
	openResponse := httptest.NewRecorder()
	open.ServeHTTP(openResponse, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if openResponse.Code != http.StatusOK || openResponse.Body.Len() == 0 {
		t.Fatalf("token unset: got %d, %d bytes", openResponse.Code, openResponse.Body.Len())
	}

	// Token set: the endpoint demands a matching bearer and hides behind 404 otherwise.
	guarded := httpserver.NewHandler(httpserver.Options{Service: "api", Metrics: observability.NewMetrics(), MetricsToken: "scrape-secret"})
	for _, tc := range []struct {
		name, authorization string
		status              int
	}{
		{"no header", "", http.StatusNotFound},
		{"wrong token", "Bearer wrong", http.StatusNotFound},
		{"missing bearer scheme", "scrape-secret", http.StatusNotFound},
		{"correct token", "Bearer scrape-secret", http.StatusOK},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/metrics", nil)
			if tc.authorization != "" {
				r.Header.Set("Authorization", tc.authorization)
			}
			w := httptest.NewRecorder()
			guarded.ServeHTTP(w, r)
			if w.Code != tc.status {
				t.Fatalf("got %d: %s", w.Code, w.Body.String())
			}
			if tc.status == http.StatusNotFound {
				var body httpserver.ErrorResponse
				if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
					t.Fatalf("not JSON envelope: %v", err)
				}
				if body.Error.Code != "not_found" {
					t.Fatalf("code=%q", body.Error.Code)
				}
			} else if w.Body.Len() == 0 {
				t.Fatal("metrics body empty on authorized scrape")
			}
		})
	}

	// The guard is scoped to /metrics; health probes stay open with a token set.
	healthz := httptest.NewRecorder()
	guarded.ServeHTTP(healthz, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if healthz.Code != http.StatusOK {
		t.Fatalf("healthz guarded: %d", healthz.Code)
	}
}

func TestPanicLogsValueAndStackWithoutSecrets(t *testing.T) {
	var logs bytes.Buffer
	handler := httpserver.NewHandler(httpserver.Options{
		Service: "api",
		Logger:  slog.New(slog.NewJSONHandler(&logs, nil)),
		Register: func(mux *http.ServeMux) {
			mux.HandleFunc("GET /boom", func(http.ResponseWriter, *http.Request) {
				panic("boom-sentinel-7f3a")
			})
		},
	})
	r := httptest.NewRequest(http.MethodGet, "/boom?token=query-secret", nil)
	r.Header.Set("Cookie", "session=cookie-secret")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status=%d: %s", w.Code, w.Body.String())
	}
	var body httpserver.ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("not JSON envelope: %v", err)
	}
	if body.Error.Code != "internal_error" {
		t.Fatalf("code=%q", body.Error.Code)
	}

	var panicEntry map[string]any
	for _, line := range strings.Split(strings.TrimSpace(logs.String()), "\n") {
		var entry map[string]any
		if json.Unmarshal([]byte(line), &entry) == nil && entry["msg"] == "request panicked" {
			panicEntry = entry
			break
		}
	}
	if panicEntry == nil {
		t.Fatalf("no panic log entry: %s", logs.String())
	}
	if panicEntry["panic"] != "boom-sentinel-7f3a" {
		t.Fatalf("panic value not logged: %v", panicEntry["panic"])
	}
	stack, _ := panicEntry["stack"].(string)
	if !strings.Contains(stack, "goroutine") || !strings.Contains(stack, "server_test.go") {
		t.Fatalf("stack did not reach the panic site: %q", stack)
	}
	for _, secret := range []string{"query-secret", "cookie-secret"} {
		if strings.Contains(logs.String(), secret) {
			t.Fatalf("secret leaked into panic logs: %s", logs.String())
		}
	}
}

func TestSafeErrorsAndCorrelation(t *testing.T) {
	handler := httpserver.NewHandler(httpserver.Options{
		Service: "api", Ready: func(context.Context) error { return errors.New("password=private") },
	})
	for _, tc := range []struct {
		name, method, path, code string
		status                   int
	}{
		{"dependency", "GET", "/readyz", "dependency_unavailable", 503},
		{"unknown route", "GET", "/secret?token=private", "not_found", 404},
		{"wrong method", "POST", "/healthz", "method_not_allowed", 405},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest(tc.method, tc.path, nil)
			r.Header.Set("X-Request-ID", "external-untrusted-value")
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, r)
			if w.Code != tc.status {
				t.Fatalf("got %d: %s", w.Code, w.Body.String())
			}
			var body httpserver.ErrorResponse
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if body.Error.Code != tc.code || body.Error.RequestID == "" || body.Error.RequestID != w.Header().Get("X-Request-ID") {
				t.Fatalf("invalid envelope: %+v", body)
			}
			if body.Error.RequestID == "external-untrusted-value" {
				t.Fatal("untrusted ID accepted")
			}
		})
	}
}
