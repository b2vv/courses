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

type DraftHandler struct{ drafting *application.Drafting }

func NewDraftHandler(drafting *application.Drafting) *DraftHandler {
	return &DraftHandler{drafting: drafting}
}

func (handler *DraftHandler) RegisterAdmin(
	mux *http.ServeMux,
	authenticate func(http.Handler) http.Handler,
	sameOrigin func(http.Handler) http.Handler,
) {
	mux.Handle("GET /api/v1/admin/lessons/{lessonId}/draft",
		authenticate(http.HandlerFunc(handler.getDraft)))
	mux.Handle("POST /api/v1/admin/lessons/{lessonId}/draft/regenerate",
		authenticate(sameOrigin(http.HandlerFunc(handler.regenerate))))
	mux.Handle("POST /api/v1/admin/lessons/{lessonId}/draft/publish",
		authenticate(sameOrigin(http.HandlerFunc(handler.publish))))
}

type draftSegmentDTO struct {
	ID         string   `json:"id"`
	StartMs    int64    `json:"startMs"`
	EndMs      int64    `json:"endMs"`
	Text       string   `json:"text"`
	Speaker    *string  `json:"speaker,omitempty"`
	Confidence *float64 `json:"confidence,omitempty"`
}

type draftDTO struct {
	LessonID   string            `json:"lessonId"`
	State      *string           `json:"state"`
	Language   *string           `json:"language"`
	Failure    *string           `json:"failure"`
	Stale      bool              `json:"stale"`
	UpdatedAt  *string           `json:"updatedAt"`
	AssetID    *string           `json:"assetId"`
	Generation *int32            `json:"generation"`
	Segments   []draftSegmentDTO `json:"segments"`
}

func (handler *DraftHandler) getDraft(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")

	transportActor, ok := httpserver.ActorFromContext(r.Context())
	if !ok {
		httpserver.WriteError(w, r, http.StatusUnauthorized, "authentication_required", "Authentication is required")
		return
	}
	actor := domain.Actor{UserID: transportActor.UserID, Administrator: transportActor.Role == "admin", EmailVerified: transportActor.EmailVerified}

	draft, err := handler.drafting.GetDraft(r.Context(), actor, r.PathValue("lessonId"))
	if err != nil {
		writeTranscriptError(w, r, err)
		return
	}
	if !draft.Found {
		httpserver.WriteJSON(w, http.StatusOK, draftDTO{
			LessonID: r.PathValue("lessonId"),
			Segments: []draftSegmentDTO{},
		})
		return
	}

	segments := make([]draftSegmentDTO, 0, len(draft.Segments))
	for _, s := range draft.Segments {
		mapped := draftSegmentDTO{
			ID: s.ID, StartMs: s.StartMs, EndMs: s.EndMs,
			Text: s.Text, Confidence: s.Confidence,
		}
		if s.Speaker != "" {
			v := s.Speaker
			mapped.Speaker = &v
		}
		segments = append(segments, mapped)
	}

	state := string(draft.State)
	dto := draftDTO{
		LessonID:   r.PathValue("lessonId"),
		State:      &state,
		Stale:      draft.Stale,
		Segments:   segments,
		AssetID:    &draft.AssetID,
		Generation: &draft.Generation,
	}
	if draft.Language != "" {
		lang := string(draft.Language)
		dto.Language = &lang
	}
	if draft.Failure != "" {
		dto.Failure = &draft.Failure
	}
	if draft.UpdatedAt != "" {
		dto.UpdatedAt = &draft.UpdatedAt
	}
	httpserver.WriteJSON(w, http.StatusOK, dto)
}

type publishRequest struct {
	ExpectedUpdatedAt string `json:"expectedUpdatedAt"`
}

func (handler *DraftHandler) publish(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")

	transportActor, ok := httpserver.ActorFromContext(r.Context())
	if !ok {
		httpserver.WriteError(w, r, http.StatusUnauthorized, "authentication_required", "Authentication is required")
		return
	}
	actor := domain.Actor{UserID: transportActor.UserID, Administrator: transportActor.Role == "admin", EmailVerified: transportActor.EmailVerified}

	var request publishRequest
	if r.ContentLength > 0 {
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1024))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&request); err != nil {
			writeTranscriptError(w, r, domain.ErrInvalidRequest)
			return
		}
		if err := decoder.Decode(new(json.RawMessage)); !errors.Is(err, io.EOF) {
			writeTranscriptError(w, r, domain.ErrInvalidRequest)
			return
		}
	}

	token, err := handler.drafting.Publish(r.Context(), actor, application.PublishRequest{
		LessonID:          r.PathValue("lessonId"),
		ExpectedUpdatedAt: request.ExpectedUpdatedAt,
	})
	if err != nil {
		writeTranscriptError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, struct {
		UpdatedAt string `json:"updatedAt"`
	}{UpdatedAt: token})
}

type regenerateRequest struct {
	Language *string `json:"language"`
}

func (handler *DraftHandler) regenerate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")

	transportActor, ok := httpserver.ActorFromContext(r.Context())
	if !ok {
		httpserver.WriteError(w, r, http.StatusUnauthorized, "authentication_required", "Authentication is required")
		return
	}
	actor := domain.Actor{UserID: transportActor.UserID, Administrator: transportActor.Role == "admin", EmailVerified: transportActor.EmailVerified}

	var language domain.Language
	if r.ContentLength > 0 {
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1024))
		decoder.DisallowUnknownFields()
		var request regenerateRequest
		if err := decoder.Decode(&request); err != nil {
			writeTranscriptError(w, r, domain.ErrInvalidRequest)
			return
		}
		if err := decoder.Decode(new(json.RawMessage)); !errors.Is(err, io.EOF) {
			writeTranscriptError(w, r, domain.ErrInvalidRequest)
			return
		}
		if request.Language != nil && *request.Language != "" {
			parsed, err := domain.NewLanguage(*request.Language)
			if err != nil {
				writeTranscriptError(w, r, err)
				return
			}
			language = parsed
		}
	}

	err := handler.drafting.Regenerate(r.Context(), actor, r.PathValue("lessonId"), language)
	if err != nil {
		writeTranscriptError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}
