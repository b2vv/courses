package http

import (
	"errors"
	"net/http"
	"net/url"
	"time"

	"courses/internal/course/application"
	"courses/internal/course/domain"
	"courses/platform/httpserver"
)

const maximumDocumentWriteBody = 2 << 20

type DocumentHandler struct{ documents *application.Documents }

func NewDocumentHandler(documents *application.Documents) *DocumentHandler {
	return &DocumentHandler{documents: documents}
}

func (h *DocumentHandler) Register(mux *http.ServeMux, authenticate func(http.Handler) http.Handler) {
	mux.Handle("GET /api/v1/learn/lessons/{lessonId}/document", authenticate(http.HandlerFunc(h.getPublished)))
}

func (h *DocumentHandler) RegisterAdmin(mux *http.ServeMux, authenticate, requireSameOrigin func(http.Handler) http.Handler) {
	mux.Handle("GET /api/v1/admin/lessons/{lessonId}/document", authenticate(http.HandlerFunc(h.getDraft)))
	mux.Handle("PUT /api/v1/admin/lessons/{lessonId}/document", requireSameOrigin(authenticate(http.HandlerFunc(h.saveDraft))))
	mux.Handle("POST /api/v1/admin/lessons/{lessonId}/document/publish", requireSameOrigin(authenticate(http.HandlerFunc(h.publish))))
}

type lessonDocumentDraftResponse struct {
	Document  domain.LessonDocument `json:"document"`
	UpdatedAt *string               `json:"updatedAt"`
}

type saveLessonDocumentRequest struct {
	ExpectedUpdatedAt *string               `json:"expectedUpdatedAt,omitempty"`
	Document          domain.LessonDocument `json:"document"`
}

type lessonDocumentUpdatedResponse struct {
	UpdatedAt string `json:"updatedAt"`
}

func (h *DocumentHandler) getDraft(w http.ResponseWriter, r *http.Request) {
	adminHeaders(w)
	actor, ok := adminActor(w, r)
	if !ok {
		return
	}
	locale, err := documentLocale(r.URL.RawQuery)
	if err != nil {
		writeDocumentError(w, r, err)
		return
	}
	body, updatedAt, err := h.documents.GetDraft(r.Context(), actor, r.PathValue("lessonId"), locale)
	if err != nil {
		writeDocumentError(w, r, err)
		return
	}
	var token *string
	if !updatedAt.IsZero() {
		value := formatAdminTime(updatedAt)
		token = &value
	}
	httpserver.WriteJSON(w, http.StatusOK, lessonDocumentDraftResponse{Document: body, UpdatedAt: token})
}

func (h *DocumentHandler) saveDraft(w http.ResponseWriter, r *http.Request) {
	adminHeaders(w)
	actor, ok := adminActor(w, r)
	if !ok {
		return
	}
	if r.URL.RawQuery != "" {
		writeDocumentError(w, r, domain.ErrInvalidQuery)
		return
	}
	request, err := decodeLimitedAdminJSON[saveLessonDocumentRequest](w, r, maximumDocumentWriteBody)
	if err != nil {
		writeDocumentError(w, r, err)
		return
	}
	var expected time.Time
	if request.ExpectedUpdatedAt != nil {
		expected, err = parseAdminTime(*request.ExpectedUpdatedAt)
		if err != nil {
			writeDocumentError(w, r, domain.ErrInvalidRequest)
			return
		}
	}
	updatedAt, err := h.documents.SaveDraft(r.Context(), actor, r.PathValue("lessonId"), request.Document.Locale, request.Document, expected)
	if err != nil {
		writeDocumentError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, lessonDocumentUpdatedResponse{UpdatedAt: formatAdminTime(updatedAt)})
}

func (h *DocumentHandler) publish(w http.ResponseWriter, r *http.Request) {
	adminHeaders(w)
	actor, ok := adminActor(w, r)
	if !ok {
		return
	}
	locale, err := documentLocale(r.URL.RawQuery)
	if err != nil {
		writeDocumentError(w, r, err)
		return
	}
	request, err := decodeLimitedAdminJSON[lessonDocumentUpdatedRequest](w, r, 1024)
	if err != nil {
		writeDocumentError(w, r, err)
		return
	}
	expected, err := parseAdminTime(request.ExpectedUpdatedAt)
	if err != nil {
		writeDocumentError(w, r, domain.ErrInvalidRequest)
		return
	}
	if err := h.documents.PublishDocument(r.Context(), actor, r.PathValue("lessonId"), locale, expected); err != nil {
		writeDocumentError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

type lessonDocumentUpdatedRequest struct {
	ExpectedUpdatedAt string `json:"expectedUpdatedAt"`
}

func (h *DocumentHandler) getPublished(w http.ResponseWriter, r *http.Request) {
	adminHeaders(w)
	actor, ok := learningActor(w, r)
	if !ok {
		return
	}
	locale, err := documentLocale(r.URL.RawQuery)
	if err != nil {
		writeDocumentError(w, r, err)
		return
	}
	body, err := h.documents.GetPublishedDocument(r.Context(), actor, r.PathValue("lessonId"), locale)
	if err != nil {
		writeDocumentError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, body)
}

func documentLocale(raw string) (string, error) {
	query, err := url.ParseQuery(raw)
	if err != nil {
		return "", domain.ErrInvalidQuery
	}
	locale := "uk"
	for key, values := range query {
		if key != "locale" || len(values) != 1 || (values[0] != "uk" && values[0] != "en") {
			return "", domain.ErrInvalidQuery
		}
		locale = values[0]
	}
	return locale, nil
}

func writeDocumentError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidLessonDocument):
		httpserver.WriteError(w, r, http.StatusBadRequest, "invalid_lesson_document", "Lesson document is invalid")
	case errors.Is(err, domain.ErrInvalidRequest):
		httpserver.WriteError(w, r, http.StatusBadRequest, "invalid_request", "Request is invalid")
	case errors.Is(err, domain.ErrInvalidQuery):
		httpserver.WriteError(w, r, http.StatusBadRequest, "invalid_query", "Query is invalid")
	case errors.Is(err, domain.ErrEmailNotVerified):
		httpserver.WriteError(w, r, http.StatusForbidden, "email_not_verified", "Email verification is required")
	case errors.Is(err, domain.ErrPurchaseRequired):
		httpserver.WriteError(w, r, http.StatusForbidden, "course_purchase_required", "Course purchase is required")
	case errors.Is(err, domain.ErrForbidden):
		httpserver.WriteError(w, r, http.StatusForbidden, "forbidden", "Permission is required")
	case errors.Is(err, domain.ErrCourseForbidden):
		httpserver.WriteError(w, r, http.StatusForbidden, "course_forbidden", "Course access is forbidden")
	case errors.Is(err, domain.ErrLessonNotFound):
		httpserver.WriteError(w, r, http.StatusNotFound, "lesson_not_found", "Lesson was not found")
	case errors.Is(err, domain.ErrDocumentConflict):
		httpserver.WriteError(w, r, http.StatusConflict, "document_conflict", "Document changed since it was loaded")
	case errors.Is(err, domain.ErrDependencyUnavailable):
		httpserver.WriteError(w, r, http.StatusServiceUnavailable, "dependency_unavailable", "A dependency is unavailable")
	default:
		httpserver.WriteError(w, r, http.StatusInternalServerError, "internal_error", "Unexpected server error")
	}
}
