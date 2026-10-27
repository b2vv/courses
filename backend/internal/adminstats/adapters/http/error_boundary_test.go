package http

import (
	"courses/internal/adminstats/domain"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestErrorBoundaries(t *testing.T) {
	writers := map[string]func(http.ResponseWriter, *http.Request, error){
		"writeError": writeError,
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
		{"writeError", domain.ErrInvalidRange, http.StatusBadRequest, string(domain.ErrInvalidRange)},
		{"writeError", domain.ErrRangeTooWide, http.StatusBadRequest, string(domain.ErrRangeTooWide)},
		{"writeError", domain.ErrUnknownMetric, http.StatusBadRequest, string(domain.ErrUnknownMetric)},
		{"writeError", domain.ErrUnknownBucket, http.StatusBadRequest, string(domain.ErrUnknownBucket)},
		{"writeError", domain.ErrDependencyUnavailable, http.StatusServiceUnavailable, string(domain.ErrDependencyUnavailable)},
	}
	for _, tc := range cases {
		t.Run(tc.writer+"/"+tc.code, func(t *testing.T) {
			assertResponse(t, writers[tc.writer], tc.err, tc.status, tc.code)
			assertResponse(t, writers[tc.writer], fmt.Errorf("private infrastructure context: %w", tc.err), tc.status, tc.code)
		})
	}
}
