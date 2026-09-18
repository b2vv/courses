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

type AdminHandler struct{ authoring *application.Authoring }

func NewAdminHandler(authoring *application.Authoring) *AdminHandler {
	return &AdminHandler{authoring: authoring}
}

func (handler *AdminHandler) RegisterAdmin(
	mux *http.ServeMux,
	authenticate func(http.Handler) http.Handler,
	sameOrigin func(http.Handler) http.Handler,
) {
	mux.Handle("PUT /api/v1/admin/lessons/{lessonId}/transcript",
		authenticate(sameOrigin(http.HandlerFunc(handler.save))))
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
	httpserver.WriteJSON(w, http.StatusOK, struct {
		UpdatedAt string `json:"updatedAt"`
	}{UpdatedAt: updatedAt})
}
