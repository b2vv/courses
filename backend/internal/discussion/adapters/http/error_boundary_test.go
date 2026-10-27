package http

import (
	"courses/internal/discussion/application"
	"courses/internal/discussion/domain"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestErrorBoundaries(t *testing.T) {
	writers := map[string]func(http.ResponseWriter, *http.Request, error){
		"writeListError":              writeListError,
		"writePublishError":           writePublishError,
		"writeMutationError":          writeMutationError,
		"writeReportError":            writeReportError,
		"writeTeachListError":         writeTeachListError,
		"writeParticipantReportError": writeParticipantReportError,
		"writeReviewError": func(w http.ResponseWriter, r *http.Request, err error) {
			writeReviewError(w, r, err, "course_not_found")
		},
		"writeError": func(w http.ResponseWriter, r *http.Request, err error) { writeError(w, r, http.StatusBadRequest, err) },
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
		{"writeListError", domain.ErrBodyInvalid, http.StatusBadRequest, "comment_body_invalid"},
		{"writeListError", domain.ErrPositionInvalid, http.StatusBadRequest, "comment_position_invalid"},
		{"writeListError", domain.ErrParentInvalid, http.StatusBadRequest, "comment_parent_invalid"},
		{"writeListError", domain.ErrThreadFull, http.StatusConflict, "comment_thread_full"},
		{"writeListError", domain.ErrInvalidCursor, http.StatusBadRequest, "comment_cursor_invalid"},
		{"writeListError", domain.ErrDependencyUnavailable, http.StatusServiceUnavailable, "dependency_unavailable"},
		{"writeListError", domain.ErrForbidden, http.StatusForbidden, "course_forbidden"},
		{"writeListError", domain.ErrPurchaseRequired, http.StatusForbidden, "course_purchase_required"},
		{"writeListError", domain.ErrEnrollmentRequired, http.StatusForbidden, "enrollment_required"},
		{"writeListError", domain.ErrEnrollmentBlocked, http.StatusForbidden, "enrollment_blocked"},
		{"writeListError", domain.ErrInvalidRequest, http.StatusBadRequest, "invalid_request"},
		{"writeListError", domain.ErrReportOwnComment, http.StatusForbidden, "report_own_comment"},
		{"writeListError", domain.ErrReviewOwnerForbidden, http.StatusForbidden, "review_owner_forbidden"},
		{"writeListError", domain.ErrReviewNotAvailable, http.StatusForbidden, "review_not_available"},
		{"writeListError", domain.ErrReviewPurchaseRequired, http.StatusForbidden, "review_purchase_required"},
		{"writeListError", domain.ErrReviewNotStarted, http.StatusForbidden, "review_not_started"},
		{"writeListError", domain.ErrReviewInvalid, http.StatusBadRequest, "review_invalid"},
		{"writeListError", domain.ErrReviewAlreadyExists, http.StatusConflict, "review_already_exists"},
		{"writeListError", domain.ErrReviewRemoved, http.StatusConflict, "review_removed"},
		{"writeListError", domain.ErrReportOwnReview, http.StatusForbidden, "report_own_review"},
		{"writeListError", domain.ErrReviewModerationForbidden, http.StatusForbidden, "review_moderation_forbidden"},
		{"writeListError", domain.ErrReportSelf, http.StatusForbidden, "report_own_person"},
		{"writeListError", domain.ErrReportInvalid, http.StatusBadRequest, "report_invalid"},
		{"writeListError", domain.ErrReportForbidden, http.StatusForbidden, "report_forbidden"},
		{"writeListError", domain.ErrReportDuplicate, http.StatusConflict, "report_already_open"},
		{"writeListError", application.ErrCommentForbidden, http.StatusForbidden, "comment_forbidden"},
		{"writeListError", application.ErrCommentRemoved, http.StatusConflict, "comment_removed"},
		{"writeListError", application.ErrEmailNotVerified, http.StatusForbidden, "email_not_verified"},
		{"writeListError", domain.ErrNotFound, http.StatusNotFound, "lesson_not_found"},
		{"writePublishError", domain.ErrNotFound, http.StatusNotFound, "lesson_not_found"},
		{"writeMutationError", domain.ErrNotFound, http.StatusNotFound, "comment_not_found"},
		{"writeReportError", domain.ErrNotFound, http.StatusNotFound, "comment_not_found"},
		{"writeTeachListError", domain.ErrNotFound, http.StatusNotFound, "course_not_found"},
		{"writeParticipantReportError", domain.ErrNotFound, http.StatusNotFound, "course_not_found"},
		{"writeReviewError", domain.ErrNotFound, http.StatusNotFound, "course_not_found"},
	}
	for _, tc := range cases {
		t.Run(tc.writer+"/"+tc.code, func(t *testing.T) {
			assertResponse(t, writers[tc.writer], tc.err, tc.status, tc.code)
			assertResponse(t, writers[tc.writer], fmt.Errorf("private infrastructure context: %w", tc.err), tc.status, tc.code)
		})
	}
}

func TestStaticTransportErrorMappings(t *testing.T) {
	for _, tc := range []struct {
		code   string
		status int
	}{{"origin_forbidden", http.StatusForbidden}, {"course_not_found", http.StatusNotFound}, {"review_not_found", http.StatusNotFound}} {
		response := httptest.NewRecorder()
		writeErrorCode(response, httptest.NewRequest(http.MethodGet, "/", nil), tc.status, tc.code)
		var body struct {
			Error struct {
				Code string `json:"code"`
			} `json:"error"`
		}
		if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		if response.Code != tc.status || body.Error.Code != tc.code {
			t.Fatalf("got %d/%q, want %d/%q", response.Code, body.Error.Code, tc.status, tc.code)
		}
	}
}
