package http

import (
	"courses/internal/course/domain"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestErrorBoundaries(t *testing.T) {
	writers := map[string]func(http.ResponseWriter, *http.Request, error){
		"writeError":         writeError,
		"writeAdminError":    writeAdminError,
		"writeDocumentError": writeDocumentError,
		"writeLearningError": writeLearningError,
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
		{"writeError", domain.ErrInvalidQuery, http.StatusBadRequest, "invalid_query"},
		{"writeError", domain.ErrInvalidSlug, http.StatusBadRequest, "invalid_slug"},
		{"writeError", domain.ErrCourseNotFound, http.StatusNotFound, "course_not_found"},
		{"writeError", domain.ErrDependencyUnavailable, http.StatusServiceUnavailable, "dependency_unavailable"},
		{"writeDocumentError", domain.ErrInvalidLessonDocument, http.StatusBadRequest, "invalid_lesson_document"},
		{"writeDocumentError", domain.ErrInvalidRequest, http.StatusBadRequest, "invalid_request"},
		{"writeDocumentError", domain.ErrInvalidQuery, http.StatusBadRequest, "invalid_query"},
		{"writeDocumentError", domain.ErrEmailNotVerified, http.StatusForbidden, "email_not_verified"},
		{"writeDocumentError", domain.ErrPurchaseRequired, http.StatusForbidden, "course_purchase_required"},
		{"writeDocumentError", domain.ErrForbidden, http.StatusForbidden, "forbidden"},
		{"writeDocumentError", domain.ErrCourseForbidden, http.StatusForbidden, "course_forbidden"},
		{"writeDocumentError", domain.ErrLessonNotFound, http.StatusNotFound, "lesson_not_found"},
		{"writeDocumentError", domain.ErrDocumentConflict, http.StatusConflict, "document_conflict"},
		{"writeDocumentError", domain.ErrDependencyUnavailable, http.StatusServiceUnavailable, "dependency_unavailable"},
		{"writeAdminError", domain.ErrInvalidRequest, http.StatusBadRequest, "invalid_request"},
		{"writeAdminError", domain.ErrForbidden, http.StatusForbidden, "forbidden"},
		{"writeAdminError", domain.ErrCourseForbidden, http.StatusForbidden, "course_forbidden"},
		{"writeAdminError", domain.ErrCourseNotFound, http.StatusNotFound, "course_not_found"},
		{"writeAdminError", domain.ErrInvalidOwner, http.StatusBadRequest, "invalid_owner"},
		{"writeAdminError", domain.ErrSlugTaken, http.StatusConflict, "slug_taken"},
		{"writeAdminError", domain.ErrSlugImmutable, http.StatusConflict, "slug_immutable"},
		{"writeAdminError", domain.ErrCourseModified, http.StatusConflict, "course_modified"},
		{"writeAdminError", domain.ErrInvalidStatusTransition, http.StatusConflict, "invalid_status_transition"},
		{"writeAdminError", domain.ErrCourseNotDeletable, http.StatusConflict, "course_not_deletable"},
		{"writeAdminError", domain.ErrLessonLimit, http.StatusBadRequest, "lesson_limit"},
		{"writeAdminError", domain.ErrDependencyUnavailable, http.StatusServiceUnavailable, "dependency_unavailable"},
		{"writeLearningError", domain.ErrPurchaseRequired, http.StatusForbidden, "course_purchase_required"},
		{"writeLearningError", domain.ErrForbidden, http.StatusForbidden, "course_forbidden"},
		{"writeLearningError", domain.ErrCourseForbidden, http.StatusForbidden, "course_forbidden"},
	}
	for _, tc := range cases {
		t.Run(tc.writer+"/"+tc.code, func(t *testing.T) {
			assertResponse(t, writers[tc.writer], tc.err, tc.status, tc.code)
			assertResponse(t, writers[tc.writer], fmt.Errorf("private infrastructure context: %w", tc.err), tc.status, tc.code)
		})
	}
}
