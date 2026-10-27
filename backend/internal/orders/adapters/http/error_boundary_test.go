package http

import (
	"courses/internal/orders/domain"
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
		{"writeError", domain.ErrEmailNotVerified, http.StatusForbidden, "email_not_verified"},
		{"writeError", domain.ErrOrderInvalid, http.StatusBadRequest, "invalid_request"},
		{"writeError", domain.ErrGrantInvalid, http.StatusBadRequest, "invalid_request"},
		{"writeError", domain.ErrMoneyInvalid, http.StatusBadRequest, "invalid_request"},
		{"writeError", domain.ErrCourseNotPurchasable, http.StatusBadRequest, "invalid_request"},
		{"writeError", domain.ErrCourseAlreadyOwned, http.StatusConflict, "course_already_owned"},
		{"writeError", domain.ErrMembershipConflict, http.StatusConflict, "membership_conflict"},
		{"writeError", domain.ErrGrantTransitionInvalid, http.StatusConflict, "grant_transition_invalid"},
		{"writeError", domain.ErrCourseNotEnrollable, http.StatusConflict, "course_not_enrollable"},
		{"writeError", domain.ErrEnrollmentForbidden, http.StatusForbidden, "enrollment_forbidden"},
		{"writeError", domain.ErrLearnerCourseForbidden, http.StatusForbidden, "learner_course_forbidden"},
		{"writeError", domain.ErrInvalidCursor, http.StatusBadRequest, "invalid_request"},
		{"writeError", domain.ErrInvalidQuery, http.StatusBadRequest, "invalid_request"},
		{"writeError", domain.ErrOrderNotFound, http.StatusNotFound, "order_not_found"},
		{"writeError", domain.ErrCommentNotFound, http.StatusNotFound, "comment_not_found"},
		{"writeError", domain.ErrReviewNotFound, http.StatusNotFound, "review_not_found"},
		{"writeError", domain.ErrCheckoutAlreadyOpened, http.StatusConflict, "checkout_already_opened"},
		{"writeError", domain.ErrGrantForbidden, http.StatusForbidden, "grant_forbidden"},
		{"writeError", domain.ErrGrantSelfForbidden, http.StatusForbidden, "grant_self_forbidden"},
		{"writeError", domain.ErrAdminNotLearner, http.StatusForbidden, "admin_not_learner"},
		{"writeError", domain.ErrPaymentEventRejected, http.StatusConflict, "payment_event_rejected"},
		{"writeError", domain.ErrPaymentEventUnmatched, http.StatusConflict, "payment_event_rejected"},
		{"writeError", domain.ErrCourseModified, http.StatusConflict, "course_modified"},
		{"writeError", domain.ErrDependencyUnavailable, http.StatusServiceUnavailable, "dependency_unavailable"},
	}
	for _, tc := range cases {
		t.Run(tc.writer+"/"+tc.code, func(t *testing.T) {
			assertResponse(t, writers[tc.writer], tc.err, tc.status, tc.code)
			assertResponse(t, writers[tc.writer], fmt.Errorf("private infrastructure context: %w", tc.err), tc.status, tc.code)
		})
	}
}
