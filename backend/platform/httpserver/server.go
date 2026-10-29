package httpserver

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"courses/platform/observability"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

type requestIDKey struct{}

type Options struct {
	Service  string
	Ready    func(context.Context) error
	Register func(*http.ServeMux)
	Logger   *slog.Logger
	Metrics  *observability.Metrics
	// MetricsToken, when non-empty, requires `Authorization: Bearer <token>` on
	// GET /metrics; an absent or wrong token is answered 404 (as if the endpoint
	// did not exist). Empty leaves /metrics unauthenticated, which is the default
	// for local and test runs where Prometheus scrapes the service port directly.
	MetricsToken string
	Tracer       trace.Tracer
}

func RequestID(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey{}).(string)
	return id
}

type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}
type ErrorDetail struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"requestId"`
}

const contentSecurityPolicy = "default-src 'none'; frame-ancestors 'none'"

func WriteJSON(w http.ResponseWriter, status int, value interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func WriteError(w http.ResponseWriter, r *http.Request, status int, code, message string) {
	WriteJSON(w, status, ErrorResponse{Error: ErrorDetail{Code: code, Message: message, RequestID: RequestID(r.Context())}})
}

func NewHandler(options Options) http.Handler {
	if options.Logger == nil {
		options.Logger = slog.New(slog.NewJSONHandler(io.Discard, nil))
	}
	if options.Tracer == nil {
		options.Tracer = noop.NewTracerProvider().Tracer("http")
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		WriteJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": options.Service})
	})
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if options.Ready == nil || options.Ready(ctx) != nil {
			WriteError(w, r, http.StatusServiceUnavailable, "dependency_unavailable", "Service is not ready")
			return
		}
		WriteJSON(w, http.StatusOK, map[string]string{"status": "ready", "service": options.Service})
	})
	if options.Metrics != nil {
		metricsHandler := options.Metrics.Handler()
		if options.MetricsToken != "" {
			metricsHandler = requireMetricsToken(options.MetricsToken, metricsHandler)
		}
		mux.Handle("GET /metrics", metricsHandler)
	}
	if options.Register != nil {
		options.Register(mux)
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy", contentSecurityPolicy)
		start := time.Now()
		bytes := make([]byte, 16)
		if _, err := rand.Read(bytes); err != nil {
			panic("request ID entropy unavailable")
		}
		id := hex.EncodeToString(bytes)
		w.Header().Set("X-Request-ID", id)
		r = r.WithContext(context.WithValue(r.Context(), requestIDKey{}, id))
		_, pattern := mux.Handler(r)
		route := pattern
		if route == "" {
			route = "unmatched"
		}
		method := r.Method
		switch method {
		case "GET", "HEAD", "POST", "PUT", "PATCH", "DELETE", "OPTIONS":
		default:
			method = "OTHER"
		}
		ctx := propagation.TraceContext{}.Extract(r.Context(), propagation.HeaderCarrier(r.Header))
		ctx, span := options.Tracer.Start(ctx, method+" "+route, trace.WithSpanKind(trace.SpanKindServer), trace.WithAttributes(attribute.String("http.request.method", method), attribute.String("http.route", route)))
		defer span.End()
		r = r.WithContext(ctx)
		response := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		defer func() {
			if recover() != nil {
				if !response.written {
					WriteError(response, r, http.StatusInternalServerError, "internal_error", "Unexpected server error")
				}
				span.SetStatus(codes.Error, "request panicked")
				options.Logger.Error("request panicked", "requestId", id, "traceId", span.SpanContext().TraceID().String())
			}
			span.SetAttributes(attribute.Int("http.response.status_code", response.status))
			if response.status >= 500 {
				span.SetStatus(codes.Error, "request failed")
			}
			if options.Metrics != nil {
				options.Metrics.Requests.WithLabelValues(method, route, strconv.Itoa(response.status)).Inc()
				options.Metrics.Duration.WithLabelValues(method, route).Observe(time.Since(start).Seconds())
			}
			options.Logger.Info("http request", "method", method, "route", route, "status", response.status, "durationMs", time.Since(start).Milliseconds(), "requestId", id, "traceId", span.SpanContext().TraceID().String())
		}()
		if pattern == "" {
			mux.ServeHTTP(&routingErrorWriter{ResponseWriter: response, request: r}, r)
			return
		}
		mux.ServeHTTP(response, r)
	})
}

// requireMetricsToken gates next behind a bearer token. A missing or mismatched
// token is answered 404 so the endpoint's existence is not revealed to an
// unauthenticated caller — the same posture nginx takes for /metrics from
// outside. Both sides are hashed to a fixed 32 bytes before the constant-time
// compare so neither the length of the expected token nor its content leaks
// through timing (ConstantTimeCompare returns early on a length mismatch).
func requireMetricsToken(token string, next http.Handler) http.Handler {
	want := sha256.Sum256([]byte("Bearer " + token))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got := sha256.Sum256([]byte(r.Header.Get("Authorization")))
		if subtle.ConstantTimeCompare(got[:], want[:]) != 1 {
			WriteError(w, r, http.StatusNotFound, "not_found", "Resource was not found")
			return
		}
		next.ServeHTTP(w, r)
	})
}

type responseWriter struct {
	http.ResponseWriter
	status  int
	written bool
}

func (w *responseWriter) WriteHeader(status int) {
	if w.written {
		return
	}
	w.status, w.written = status, true
	w.ResponseWriter.WriteHeader(status)
}
func (w *responseWriter) Write(value []byte) (int, error) {
	if !w.written {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(value)
}
func (w *responseWriter) Unwrap() http.ResponseWriter { return w.ResponseWriter }

type routingErrorWriter struct {
	http.ResponseWriter
	request  *http.Request
	replaced bool
}

func (w *routingErrorWriter) WriteHeader(status int) {
	switch status {
	case http.StatusNotFound:
		w.replaced = true
		WriteError(w.ResponseWriter, w.request, status, "not_found", "Resource was not found")
	case http.StatusMethodNotAllowed:
		w.replaced = true
		WriteError(w.ResponseWriter, w.request, status, "method_not_allowed", "Method is not allowed")
	default:
		w.ResponseWriter.WriteHeader(status)
	}
}

func (w *routingErrorWriter) Write(value []byte) (int, error) {
	if w.replaced {
		return len(value), nil
	}
	return w.ResponseWriter.Write(value)
}
