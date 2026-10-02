package http

import (
	"errors"
	"net/http"
	"net/url"

	"courses/internal/transcript/application"
	"courses/internal/transcript/domain"
	"courses/platform/httpserver"
)

type Handler struct{ reading *application.Reading }

func NewHandler(reading *application.Reading) *Handler { return &Handler{reading: reading} }

func (handler *Handler) Register(mux *http.ServeMux, authenticate func(http.Handler) http.Handler) {
	mux.Handle("GET /api/v1/learn/lessons/{lessonId}/transcript", authenticate(http.HandlerFunc(handler.get)))
}

type documentDTO struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	URL   string `json:"url"`
}

type segmentDTO struct {
	ID          string   `json:"id"`
	StartMs     int64    `json:"startMs"`
	EndMs       int64    `json:"endMs"`
	Text        string   `json:"text"`
	Speaker     *string  `json:"speaker,omitempty"`
	Confidence  *float64 `json:"confidence,omitempty"`
	DocumentIDs []string `json:"documentIds"`
}

type transcriptDTO struct {
	LessonID  string        `json:"lessonId"`
	Language  *string       `json:"language"`
	UpdatedAt *string       `json:"updatedAt"`
	Stale     bool          `json:"stale"`
	Segments  []segmentDTO  `json:"segments"`
	Documents []documentDTO `json:"documents"`
}

func (handler *Handler) get(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")

	// The read takes no paging parameters by design: search runs on the client over the whole
	// lesson, so a partial payload would make a miss indistinguishable from an absent word.
	// Unknown parameters are refused rather than ignored, so that a caller who thinks it is
	// paging finds out instead of silently receiving everything.
	query, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		writeTranscriptError(w, r, domain.ErrInvalidRequest)
		return
	}
	language := query.Get("language")
	query.Del("language")
	if len(query) != 0 {
		writeTranscriptError(w, r, domain.ErrInvalidRequest)
		return
	}

	transportActor, ok := httpserver.ActorFromContext(r.Context())
	if !ok {
		httpserver.WriteError(w, r, http.StatusUnauthorized, "authentication_required", "Authentication is required")
		return
	}
	actor := domain.Actor{UserID: transportActor.UserID, Administrator: transportActor.Role == "admin"}

	transcript, err := handler.reading.Get(r.Context(), actor, r.PathValue("lessonId"), language)
	if err != nil {
		writeTranscriptError(w, r, err)
		return
	}
	if !transcript.Found {
		// Absent is a state of the lesson, not a missing resource: the lesson exists and the
		// caller may see it. 404 would be answered identically for a lesson that does not exist,
		// which is the one thing the read must not leak.
		httpserver.WriteJSON(w, http.StatusOK, transcriptDTO{
			LessonID: r.PathValue("lessonId"), Language: nil, UpdatedAt: nil, Stale: false,
			Segments: []segmentDTO{}, Documents: []documentDTO{},
		})
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, mapTranscript(transcript))
}

func mapTranscript(transcript application.Transcript) transcriptDTO {
	segments := make([]segmentDTO, 0, len(transcript.Segments))
	for _, segment := range transcript.Segments {
		mapped := segmentDTO{
			ID: segment.ID, StartMs: segment.StartMs, EndMs: segment.EndMs,
			Text: segment.Text, Confidence: segment.Confidence,
			DocumentIDs: segment.DocumentIDs,
		}
		if mapped.DocumentIDs == nil {
			mapped.DocumentIDs = []string{}
		}
		if segment.Speaker != "" {
			value := segment.Speaker
			mapped.Speaker = &value
		}
		segments = append(segments, mapped)
	}
	documents := make([]documentDTO, 0, len(transcript.Documents))
	for _, link := range transcript.Documents {
		documents = append(documents, documentDTO{ID: link.ID, Label: link.Label, URL: link.URL})
	}
	language := string(transcript.Language)
	updatedAt := transcript.UpdatedAt
	return transcriptDTO{
		LessonID: transcript.LessonID, Language: &language, UpdatedAt: &updatedAt,
		Stale: transcript.Stale, Segments: segments, Documents: documents,
	}
}

func writeTranscriptError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, domain.ErrPurchaseRequired):
		httpserver.WriteError(w, r, http.StatusForbidden, "course_purchase_required", "Purchase is required to access this course")
	case errors.Is(err, domain.ErrForbidden):
		// The same answer whether the transcript exists or not, so that the refusal does not tell
		// the caller which of the two it was.
		httpserver.WriteError(w, r, http.StatusForbidden, "course_forbidden", "Course is not available to this account")
	case errors.Is(err, domain.ErrNotFound):
		httpserver.WriteError(w, r, http.StatusNotFound, "lesson_not_found", "Lesson was not found")
	case errors.Is(err, domain.ErrTranscriptModified):
		httpserver.WriteError(w, r, http.StatusConflict, "transcript_modified", "The transcript changed since it was read")
	case errors.Is(err, domain.ErrTranscriptExists):
		httpserver.WriteError(w, r, http.StatusConflict, "transcript_exists", "This lesson already has a transcript in that language")
	case errors.Is(err, domain.ErrAlreadyRunning):
		httpserver.WriteError(w, r, http.StatusConflict, "already_running", "Recognition is already running for this asset")
	case errors.Is(err, domain.ErrSegmentLimit):
		httpserver.WriteError(w, r, http.StatusBadRequest, "transcript_segment_limit", "A transcript accepts at most 5000 segments")
	case errors.Is(err, domain.ErrTextLimit):
		httpserver.WriteError(w, r, http.StatusBadRequest, "transcript_text_limit", "A segment accepts at most 1000 characters")
	case errors.Is(err, domain.ErrDocumentLimit):
		httpserver.WriteError(w, r, http.StatusBadRequest, "transcript_document_limit", "A segment accepts at most 3 documents")
	case errors.Is(err, domain.ErrDocumentReferenced):
		httpserver.WriteError(w, r, http.StatusConflict, "transcript_document_referenced", "The document is still referenced by a transcript line")
	case errors.Is(err, domain.ErrDocumentExists):
		httpserver.WriteError(w, r, http.StatusConflict, "transcript_document_exists", "A document with that id already exists")
	case errors.Is(err, domain.ErrDependencyUnavailable):
		httpserver.WriteError(w, r, http.StatusServiceUnavailable, "dependency_unavailable", "A dependency is unavailable")
	default:
		httpserver.WriteError(w, r, http.StatusBadRequest, "invalid_request", "Request is not valid")
	}
}
