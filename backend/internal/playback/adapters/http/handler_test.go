package http_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	playbackhttp "courses/internal/playback/adapters/http"
	"courses/internal/playback/domain"
	"courses/platform/httpserver"
)

const (
	userID     = "40000000-0000-4000-8000-000000000001"
	playbackID = "11111111111111111111111111111111"
	videoID    = "50000000-0000-4000-8000-000000000001"
)

type authorizer struct {
	calls      int
	actor      domain.Actor
	playbackID string
	videoID    string
	err        error
}

func (service *authorizer) Authorize(_ context.Context, actor domain.Actor, playbackID, videoID string) error {
	service.calls++
	service.actor, service.playbackID, service.videoID = actor, playbackID, videoID
	return service.err
}

func rejectedAuthentication(status int) func(http.Handler) http.Handler {
	return func(http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(status)
		})
	}
}

func authenticated(actor httpserver.Actor) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r.WithContext(httpserver.WithActor(r.Context(), actor)))
		})
	}
}

func TestAuthorizeReturnsOnlyForbiddenOrUnavailableForFailures(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		authenticate func(http.Handler) http.Handler
		failure      error
		wantStatus   int
		wantCalls    int
		wantCode     string
	}{
		{name: "identity denial is normalized", authenticate: rejectedAuthentication(http.StatusUnauthorized), wantStatus: http.StatusForbidden},
		{name: "identity dependency stays unavailable", authenticate: rejectedAuthentication(http.StatusServiceUnavailable), wantStatus: http.StatusServiceUnavailable},
		{name: "ownership or access denial", authenticate: authenticated(httpserver.Actor{UserID: userID, Role: "student"}), failure: domain.ErrForbidden, wantStatus: http.StatusForbidden, wantCalls: 1, wantCode: "playback_forbidden"},
		{name: "dependency failure", authenticate: authenticated(httpserver.Actor{UserID: userID, Role: "student"}), failure: domain.ErrDependencyUnavailable, wantStatus: http.StatusServiceUnavailable, wantCalls: 1, wantCode: "dependency_unavailable"},
		{name: "unknown verification failure", authenticate: authenticated(httpserver.Actor{UserID: userID, Role: "student"}), failure: errors.New("unexpected"), wantStatus: http.StatusServiceUnavailable, wantCalls: 1, wantCode: "dependency_unavailable"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			service := &authorizer{err: test.failure}
			adapter := playbackhttp.NewHandler(service)
			handler := httpserver.NewHandler(httpserver.Options{Service: "api", Register: func(mux *http.ServeMux) {
				adapter.Register(mux, test.authenticate)
			}})
			request := httptest.NewRequest(http.MethodGet, "/internal/playback/authorize", nil)
			request.Header.Set("X-Playback-Id", playbackID)
			request.Header.Set("X-Video-Id", videoID)
			response := httptest.NewRecorder()

			handler.ServeHTTP(response, request)

			if response.Code != test.wantStatus || service.calls != test.wantCalls {
				t.Fatalf("status=%d calls=%d, want %d and %d", response.Code, service.calls, test.wantStatus, test.wantCalls)
			}
			if test.wantCode != "" && !strings.Contains(response.Body.String(), `"code":"`+test.wantCode+`"`) {
				t.Fatalf("body=%q, want code %q", response.Body.String(), test.wantCode)
			}
		})
	}
}

func TestAuthorizeReturnsNoContentForAnAllowedMediaRequest(t *testing.T) {
	t.Parallel()
	service := &authorizer{}
	adapter := playbackhttp.NewHandler(service)
	handler := httpserver.NewHandler(httpserver.Options{Service: "api", Register: func(mux *http.ServeMux) {
		adapter.Register(mux, authenticated(httpserver.Actor{UserID: "40000000-0000-4000-8000-000000000001", Role: "student"}))
	}})
	request := httptest.NewRequest(http.MethodGet, "/internal/playback/authorize", nil)
	request.Header.Set("X-Playback-Id", "11111111111111111111111111111111")
	request.Header.Set("X-Video-Id", "50000000-0000-4000-8000-000000000001")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent || response.Body.Len() != 0 {
		t.Fatalf("status=%d body=%q", response.Code, response.Body.String())
	}
	if service.actor != (domain.Actor{UserID: "40000000-0000-4000-8000-000000000001"}) || service.playbackID != "11111111111111111111111111111111" || service.videoID != "50000000-0000-4000-8000-000000000001" {
		t.Fatalf("actor=%+v playback=%q video=%q", service.actor, service.playbackID, service.videoID)
	}
}
