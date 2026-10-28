package http

import (
	"courses/internal/transcript/domain"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestErrorBoundaries(t *testing.T) {
	writers := map[string]func(http.ResponseWriter, *http.Request, error){
		"writeTranscriptError": writeTranscriptError,
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
				assertResponse(t, writer, err, http.StatusBadRequest, "invalid_request")
			}
		})
	}
	cases := []struct {
		writer string
		err    error
		status int
		code   string
	}{
		{"writeTranscriptError", domain.ErrPurchaseRequired, http.StatusForbidden, "course_purchase_required"},
		{"writeTranscriptError", domain.ErrEmailNotVerified, http.StatusForbidden, "email_not_verified"},
		{"writeTranscriptError", domain.ErrNotFound, http.StatusNotFound, "lesson_not_found"},
		{"writeTranscriptError", domain.ErrTranscriptModified, http.StatusConflict, "transcript_modified"},
		{"writeTranscriptError", domain.ErrTranscriptExists, http.StatusConflict, "transcript_exists"},
		{"writeTranscriptError", domain.ErrAlreadyRunning, http.StatusConflict, "already_running"},
		{"writeTranscriptError", domain.ErrSegmentLimit, http.StatusBadRequest, "transcript_segment_limit"},
		{"writeTranscriptError", domain.ErrTextLimit, http.StatusBadRequest, "transcript_text_limit"},
		{"writeTranscriptError", domain.ErrDocumentLimit, http.StatusBadRequest, "transcript_document_limit"},
		{"writeTranscriptError", domain.ErrDocumentReferenced, http.StatusConflict, "transcript_document_referenced"},
		{"writeTranscriptError", domain.ErrDocumentExists, http.StatusConflict, "transcript_document_exists"},
		{"writeTranscriptError", domain.ErrDependencyUnavailable, http.StatusServiceUnavailable, "dependency_unavailable"},
		{"writeTranscriptError", domain.ErrInvalidRequest, http.StatusBadRequest, "invalid_request"},
	}
	for _, tc := range cases {
		t.Run(tc.writer+"/"+tc.code, func(t *testing.T) {
			assertResponse(t, writers[tc.writer], tc.err, tc.status, tc.code)
			assertResponse(t, writers[tc.writer], fmt.Errorf("private infrastructure context: %w", tc.err), tc.status, tc.code)
		})
	}
}
