package http

import (
	"courses/internal/media/domain"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestErrorBoundaries(t *testing.T) {
	writers := map[string]func(http.ResponseWriter, *http.Request, error){
		"writeMediaError": writeMediaError,
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
				assertResponse(t, writer, err, http.StatusInternalServerError, "internal_error")
			}
		})
	}
	cases := []struct {
		writer string
		err    error
		status int
		code   string
	}{
		{"writeMediaError", domain.ErrInvalidRequest, http.StatusBadRequest, "invalid_request"},
		{"writeMediaError", domain.ErrUnsupportedMedia, http.StatusBadRequest, "unsupported_media"},
		{"writeMediaError", domain.ErrUploadTooLarge, http.StatusBadRequest, "upload_too_large"},
		{"writeMediaError", domain.ErrAuthenticationRequired, http.StatusUnauthorized, "authentication_required"},
		{"writeMediaError", domain.ErrForbidden, http.StatusForbidden, "forbidden"},
		{"writeMediaError", domain.ErrCourseForbidden, http.StatusForbidden, "course_forbidden"},
		{"writeMediaError", domain.ErrOriginForbidden, http.StatusForbidden, "origin_forbidden"},
		{"writeMediaError", domain.ErrLessonNotFound, http.StatusNotFound, "lesson_not_found"},
		{"writeMediaError", domain.ErrVideoNotFound, http.StatusNotFound, "video_not_found"},
		{"writeMediaError", domain.ErrUploadInProgress, http.StatusConflict, "upload_in_progress"},
		{"writeMediaError", domain.ErrVideoNotRetryable, http.StatusConflict, "video_not_retryable"},
		{"writeMediaError", domain.ErrDependencyUnavailable, http.StatusServiceUnavailable, "dependency_unavailable"},
	}
	for _, tc := range cases {
		t.Run(tc.writer+"/"+tc.code, func(t *testing.T) {
			assertResponse(t, writers[tc.writer], tc.err, tc.status, tc.code)
			assertResponse(t, writers[tc.writer], fmt.Errorf("private infrastructure context: %w", tc.err), tc.status, tc.code)
		})
	}
}
