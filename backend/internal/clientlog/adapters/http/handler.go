package http

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"courses/internal/clientlog/domain"
	"courses/platform/httpserver"
	"courses/platform/jsonbody"
)

// maximumClientLogBody bounds the ingest request. One event with an 8 KB stack
// plus the other fields fits inside this cap; anything larger is rejected
// before decoding so a hostile client cannot force a large read.
const maximumClientLogBody = 16 * 1024

// Handler ingests browser-captured events on a single public endpoint and emits
// each one as a structured log line through the platform logger. It holds no
// state: the per-session storm ceiling lives in the client adapter and the
// transport rate limit in the nginx zone.
type Handler struct {
	logger *slog.Logger
}

// NewHandler builds the client-log ingest handler around the platform logger.
func NewHandler(logger *slog.Logger) *Handler {
	return &Handler{logger: logger}
}

// Register wires the public, unauthenticated ingest endpoint. It is
// unauthenticated by design: a pre-auth client error (a failed login render, a
// boot crash) must still report. The nginx clientlog zone throttles it.
func (handler *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/client-logs", handler.ingest)
}

type clientLogRequest struct {
	Level      domain.Level `json:"level,omitempty"`
	Message    string       `json:"message"`
	Route      string       `json:"route,omitempty"`
	Release    string       `json:"release,omitempty"`
	UserAgent  string       `json:"userAgent,omitempty"`
	ErrorName  string       `json:"errorName,omitempty"`
	ErrorStack string       `json:"errorStack,omitempty"`
}

func (handler *Handler) ingest(w http.ResponseWriter, r *http.Request) {
	request, err := decode(w, r)
	if err != nil {
		// Never echo the offending body back: a fixed envelope code only.
		httpserver.WriteError(w, r, http.StatusBadRequest, string(domain.ErrInvalidEvent), "The client log event is invalid")
		return
	}
	event, err := domain.NewClientEvent(domain.Input{
		Level:      string(request.Level),
		Message:    request.Message,
		Route:      request.Route,
		Release:    request.Release,
		UserAgent:  request.UserAgent,
		ErrorName:  request.ErrorName,
		ErrorStack: request.ErrorStack,
	})
	if err != nil {
		httpserver.WriteError(w, r, http.StatusBadRequest, string(domain.ErrInvalidEvent), "The client log event is invalid")
		return
	}
	handler.emit(r.Context(), event)
	w.WriteHeader(http.StatusNoContent)
}

// emit writes the redacted event as one structured line tagged source=client so
// promtail can split it from the api/worker/ssr lines that share stdout.
func (handler *Handler) emit(ctx context.Context, event domain.ClientEvent) {
	attrs := []any{
		"source", "client",
		"route", event.Route,
		"release", event.Release,
		"ua", event.UserAgent,
		"message", event.Message,
		"requestId", httpserver.RequestID(ctx),
	}
	if event.ErrorName != "" {
		attrs = append(attrs, "errorName", event.ErrorName)
	}
	if event.ErrorStack != "" {
		attrs = append(attrs, "errorStack", event.ErrorStack)
	}
	switch event.Level {
	case domain.LevelWarn:
		handler.logger.WarnContext(ctx, "client_event", attrs...)
	case domain.LevelInfo:
		handler.logger.InfoContext(ctx, "client_event", attrs...)
	default:
		handler.logger.ErrorContext(ctx, "client_event", attrs...)
	}
}

// decode enforces the content type, the body cap, unique keys and no unknown
// fields, mirroring the other modules' ingest paths. Any failure collapses to a
// single error so the caller answers one fixed 400 without leaking the reason.
func decode(w http.ResponseWriter, r *http.Request) (clientLogRequest, error) {
	var request clientLogRequest
	if r.Header.Get("Content-Type") != "application/json" {
		return request, domain.ErrInvalidEvent
	}
	r.Body = http.MaxBytesReader(w, r.Body, maximumClientLogBody)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return request, domain.ErrInvalidEvent
	}
	if jsonbody.ValidateUniqueKeys(body) != nil {
		return request, domain.ErrInvalidEvent
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		return request, domain.ErrInvalidEvent
	}
	if decoder.Decode(&struct{}{}) != io.EOF {
		return request, domain.ErrInvalidEvent
	}
	return request, nil
}
