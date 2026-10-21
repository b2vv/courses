package http

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestHandler() (*Handler, *bytes.Buffer) {
	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, &slog.HandlerOptions{Level: slog.LevelInfo}))
	return NewHandler(logger), &logs
}

func post(t *testing.T, handler *Handler, contentType, body string) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/client-logs", strings.NewReader(body))
	if contentType != "" {
		request.Header.Set("Content-Type", contentType)
	}
	recorder := httptest.NewRecorder()
	mux := http.NewServeMux()
	handler.Register(mux)
	mux.ServeHTTP(recorder, request)
	return recorder
}

func TestIngestValidEventReturns204AndEmits(t *testing.T) {
	handler, logs := newTestHandler()
	body := `{"level":"error","message":"render boundary caught","route":"/learn/x","release":"a1b2c3d","userAgent":"Chrome","errorName":"TypeError","errorStack":"at render"}`
	recorder := post(t, handler, "application/json", body)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("want 204, got %d (%s)", recorder.Code, recorder.Body.String())
	}
	if recorder.Body.Len() != 0 {
		t.Fatalf("204 must have an empty body, got %q", recorder.Body.String())
	}
	line := logs.String()
	for _, want := range []string{`"source":"client"`, `"route":"/learn/x"`, `"release":"a1b2c3d"`, `"level":"ERROR"`} {
		if !strings.Contains(line, want) {
			t.Fatalf("emitted line missing %s: %s", want, line)
		}
	}
}

func TestIngestUnknownLevelReturns400(t *testing.T) {
	handler, logs := newTestHandler()
	recorder := post(t, handler, "application/json", `{"level":"fatal","message":"x"}`)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), `"invalid_request"`) {
		t.Fatalf("want invalid_request code, got %s", recorder.Body.String())
	}
	if logs.Len() != 0 {
		t.Fatalf("invalid event must not emit a log line, got %s", logs.String())
	}
}

func TestIngestInvalidBodyWithSecretDoesNotEcho(t *testing.T) {
	handler, _ := newTestHandler()
	// Malformed JSON (unterminated) carrying a secret: must 400 and never echo it.
	recorder := post(t, handler, "application/json", `{"level":"error","message":"password=hunter2`)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", recorder.Code)
	}
	if strings.Contains(recorder.Body.String(), "hunter2") {
		t.Fatalf("400 envelope echoed the secret: %s", recorder.Body.String())
	}
}

func TestIngestUnknownFieldReturns400(t *testing.T) {
	handler, _ := newTestHandler()
	recorder := post(t, handler, "application/json", `{"level":"error","message":"x","injected":"y"}`)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("want 400 for unknown field, got %d", recorder.Code)
	}
}

func TestIngestDuplicateKeyReturns400(t *testing.T) {
	handler, _ := newTestHandler()
	recorder := post(t, handler, "application/json", `{"level":"error","message":"a","message":"b"}`)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("want 400 for duplicate key, got %d", recorder.Code)
	}
}

func TestIngestWrongContentTypeReturns400(t *testing.T) {
	handler, _ := newTestHandler()
	recorder := post(t, handler, "text/plain", `{"level":"error","message":"x"}`)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("want 400 for wrong content-type, got %d", recorder.Code)
	}
}

func TestIngestOversizedBodyReturns400(t *testing.T) {
	handler, _ := newTestHandler()
	huge := strings.Repeat("a", maximumClientLogBody+1024)
	body := `{"level":"error","message":"` + huge + `"}`
	recorder := post(t, handler, "application/json", body)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("want 400 for oversized body, got %d", recorder.Code)
	}
}

func TestIngestScrubsSecretEndToEnd(t *testing.T) {
	handler, logs := newTestHandler()
	body := `{"level":"error","message":"login failed password=hunter2","errorStack":"Bearer sk_live_123"}`
	recorder := post(t, handler, "application/json", body)
	if recorder.Code != http.StatusNoContent {
		t.Fatalf("want 204, got %d (%s)", recorder.Code, recorder.Body.String())
	}
	line := logs.String()
	for _, secret := range []string{"hunter2", "sk_live_123"} {
		if strings.Contains(line, secret) {
			t.Fatalf("secret %q reached the log line: %s", secret, line)
		}
	}
}
