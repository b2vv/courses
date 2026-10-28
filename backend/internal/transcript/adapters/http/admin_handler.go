package http

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"courses/internal/transcript/application"
	"courses/internal/transcript/domain"
	"courses/platform/httpserver"
)

// The transcript carries its OWN body limit. The admin ceiling of 512 KiB was sized for a course
// tree; 5000 segments of text are a different class of payload, and reusing the neighbour's number
// would quietly halve the ceiling the product promises. nginx must allow more than this, strictly:
// if the two were equal, every over-sized body would be cut by nginx with an HTML 413 and the named
// refusal below would be unreachable.
const maximumTranscriptBody = 4 * 1024 * 1024

type documentsResult struct {
	Documents []documentDTO `json:"documents"`
}

type saveResult struct {
	UpdatedAt string `json:"updatedAt"`
}

type AdminHandler struct {
	authoring *application.Authoring
	library   *application.Library
}

func NewAdminHandler(authoring *application.Authoring, library *application.Library) *AdminHandler {
	return &AdminHandler{authoring: authoring, library: library}
}

func (handler *AdminHandler) RegisterAdmin(
	mux *http.ServeMux,
	authenticate func(http.Handler) http.Handler,
	sameOrigin func(http.Handler) http.Handler,
) {
	mux.Handle("PUT /api/v1/admin/lessons/{lessonId}/transcript",
		authenticate(sameOrigin(http.HandlerFunc(handler.save))))
	mux.Handle("GET /api/v1/admin/courses/{courseId}/documents",
		authenticate(http.HandlerFunc(handler.listDocuments)))
	mux.Handle("POST /api/v1/admin/courses/{courseId}/documents",
		authenticate(sameOrigin(http.HandlerFunc(handler.addDocument))))
	mux.Handle("DELETE /api/v1/admin/courses/{courseId}/documents/{documentId}",
		authenticate(sameOrigin(http.HandlerFunc(handler.removeDocument))))
}

const maximumDocumentBody = 8 * 1024

func (handler *AdminHandler) adminActor(w http.ResponseWriter, r *http.Request) (domain.Actor, bool) {
	transportActor, ok := httpserver.ActorFromContext(r.Context())
	if !ok {
		httpserver.WriteError(w, r, http.StatusUnauthorized, "authentication_required", "Authentication is required")
		return domain.Actor{}, false
	}
	return domain.Actor{
		UserID: transportActor.UserID, Administrator: transportActor.Role == "admin",
	}, true
}

func (handler *AdminHandler) listDocuments(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	actor, ok := handler.adminActor(w, r)
	if !ok {
		return
	}
	documents, err := handler.library.List(r.Context(), actor, r.PathValue("courseId"))
	if err != nil {
		writeTranscriptError(w, r, err)
		return
	}
	mapped := make([]documentDTO, 0, len(documents))
	for _, link := range documents {
		mapped = append(mapped, documentDTO{ID: link.ID, Label: link.Label, URL: link.URL})
	}
	httpserver.WriteJSON(w, http.StatusOK, documentsResult{Documents: mapped})
}

func (handler *AdminHandler) addDocument(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	actor, ok := handler.adminActor(w, r)
	if !ok {
		return
	}
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, maximumDocumentBody))
	decoder.DisallowUnknownFields()
	var request documentDTO
	if err := decoder.Decode(&request); err != nil {
		writeTranscriptError(w, r, domain.ErrInvalidRequest)
		return
	}
	if err := decoder.Decode(new(json.RawMessage)); !errors.Is(err, io.EOF) {
		writeTranscriptError(w, r, domain.ErrInvalidRequest)
		return
	}
	err := handler.library.Add(r.Context(), actor, r.PathValue("courseId"), domain.DocumentLink{
		ID: request.ID, Label: request.Label, URL: request.URL,
	})
	if err != nil {
		writeTranscriptError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (handler *AdminHandler) removeDocument(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	actor, ok := handler.adminActor(w, r)
	if !ok {
		return
	}
	err := handler.library.Remove(r.Context(), actor, r.PathValue("courseId"), r.PathValue("documentId"))
	if err != nil {
		writeTranscriptError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type saveSegmentRequest struct {
	ID          string   `json:"id"`
	StartMs     int64    `json:"startMs"`
	EndMs       int64    `json:"endMs"`
	Text        string   `json:"text"`
	Speaker     *string  `json:"speaker"`
	Confidence  *float64 `json:"confidence"`
	DocumentIDs []string `json:"documentIds"`
}

type saveTranscriptRequest struct {
	Language          string               `json:"language"`
	ExpectedUpdatedAt *string              `json:"expectedUpdatedAt"`
	Segments          []saveSegmentRequest `json:"segments"`
}

func (handler *AdminHandler) save(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")

	body := http.MaxBytesReader(w, r.Body, maximumTranscriptBody)
	decoder := json.NewDecoder(body)
	decoder.DisallowUnknownFields()
	var request saveTranscriptRequest
	if err := decoder.Decode(&request); err != nil {
		writeTranscriptError(w, r, domain.ErrInvalidRequest)
		return
	}
	// Trailing content is refused rather than ignored: a body with two documents in it means the
	// caller and the server disagree about what was sent.
	if err := decoder.Decode(new(json.RawMessage)); !errors.Is(err, io.EOF) {
		writeTranscriptError(w, r, domain.ErrInvalidRequest)
		return
	}

	transportActor, ok := httpserver.ActorFromContext(r.Context())
	if !ok {
		httpserver.WriteError(w, r, http.StatusUnauthorized, "authentication_required", "Authentication is required")
		return
	}

	segments := make([]domain.Segment, 0, len(request.Segments))
	for _, segment := range request.Segments {
		mapped := domain.Segment{
			ID: segment.ID, StartMs: segment.StartMs, EndMs: segment.EndMs,
			Text: segment.Text, Confidence: segment.Confidence, DocumentIDs: segment.DocumentIDs,
		}
		if segment.Speaker != nil {
			mapped.Speaker = *segment.Speaker
		}
		segments = append(segments, mapped)
	}

	expected := ""
	if request.ExpectedUpdatedAt != nil {
		expected = *request.ExpectedUpdatedAt
	}
	updatedAt, err := handler.authoring.Save(r.Context(), domain.Actor{
		UserID: transportActor.UserID, Administrator: transportActor.Role == "admin",
	}, application.SaveRequest{
		LessonID:          r.PathValue("lessonId"),
		Language:          request.Language,
		ExpectedUpdatedAt: expected,
		Segments:          segments,
	})
	if err != nil {
		writeTranscriptError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, saveResult{UpdatedAt: updatedAt})
}
