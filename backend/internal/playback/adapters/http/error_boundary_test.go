package http

import (
	"courses/internal/playback/domain"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestErrorBoundaries(t *testing.T) {
	writers := map[string]func(http.ResponseWriter, *http.Request, error){
		"writePlayError":     writePlayError,
		"writePlaybackError": writePlaybackError,
	}
	assertResponse := func(t *testing.T, writer func(http.ResponseWriter, *http.Request, error), err error, status int, code string) {
		t.Helper()
		response := httptest.NewRecorder()
		writer(response, httptest.NewRequest(http.MethodGet, "/", nil), err)
		var body struct {
			Error struct {
				Code string `json:"code"`
			} `json:"error"`
		}
		if decodeErr := json.Unmarshal(response.Body.Bytes(), &body); decodeErr != nil {
			t.Fatal(decodeErr)
		}
		if response.Code != status || body.Error.Code != code {
			t.Fatalf("got %d/%q, want %d/%q", response.Code, body.Error.Code, status, code)
		}
	}
	for name, writer := range writers {
		t.Run(name+"/unknown", func(t *testing.T) {
			for _, err := range []error{errors.New("boom database=private user=private"), fmt.Errorf("query: %w", errors.New("boom")), errors.New("invalid_request"), errors.New("course_not_found"), errors.New("origin_forbidden")} {
				assertResponse(t, writer, err, http.StatusServiceUnavailable, "dependency_unavailable")
			}
		})
	}
	cases := []struct {
		writer string
		err    error
		status int
		code   string
	}{
		{"writePlaybackError", errInvalidRequest, http.StatusBadRequest, "invalid_request"},
		{"writePlaybackError", domain.ErrInvalidProgress, http.StatusBadRequest, "invalid_request"},
		{"writePlaybackError", domain.ErrForbidden, http.StatusForbidden, "playback_forbidden"},
		{"writePlaybackError", domain.ErrPurchaseRequired, http.StatusForbidden, "course_purchase_required"},
		{"writePlaybackError", domain.ErrEmailNotVerified, http.StatusForbidden, "email_not_verified"},
		{"writePlaybackError", domain.ErrLessonNotFound, http.StatusNotFound, "lesson_not_found"},
		{"writePlaybackError", domain.ErrVideoNotReady, http.StatusNotFound, "video_not_ready"},
		{"writePlaybackError", domain.ErrPlaybackTakenOver, http.StatusConflict, "playback_taken_over"},
		{"writePlaybackError", domain.ErrPlaybackSessionExpired, http.StatusGone, "playback_session_expired"},
		{"writePlaybackError", domain.ErrDependencyUnavailable, http.StatusServiceUnavailable, "dependency_unavailable"},
		{"writePlayError", domain.ErrPlaybackTakenOver, http.StatusInternalServerError, "internal_error"},
		{"writePlayError", domain.ErrPlaybackSessionExpired, http.StatusInternalServerError, "internal_error"},
	}
	for _, tc := range cases {
		t.Run(tc.writer+"/"+tc.code, func(t *testing.T) {
			assertResponse(t, writers[tc.writer], tc.err, tc.status, tc.code)
			assertResponse(t, writers[tc.writer], fmt.Errorf("private infrastructure context: %w", tc.err), tc.status, tc.code)
		})
	}
}
