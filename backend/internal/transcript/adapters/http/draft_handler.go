package http

import (
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
	mux.Handle("GET /api/v1/admin/lessons/{lessonId}/drafts",
		authenticate(http.HandlerFunc(handler.listDrafts)))
	mux.Handle("POST /api/v1/admin/lessons/{lessonId}/drafts/{trackId}/regenerate",
		authenticate(sameOrigin(http.HandlerFunc(handler.regenerate))))
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
	TrackID           string            `json:"trackId"`
	TrackLanguage     string            `json:"trackLanguage"`
	TrackMultilingual bool              `json:"trackMultilingual"`
	State             domain.DraftState `json:"state"`
	Language          *string           `json:"language"`
	Failure           *string           `json:"failure"`
	Stale             bool              `json:"stale"`
	UpdatedAt         string            `json:"updatedAt"`
	AssetID           string            `json:"assetId"`
	Generation        int32             `json:"generation"`
	Segments          []draftSegmentDTO `json:"segments"`
}
type draftsDTO struct {
	Drafts []draftDTO `json:"drafts"`
}

func (handler *DraftHandler) listDrafts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	transportActor, ok := httpserver.ActorFromContext(r.Context())
	if !ok {
		httpserver.WriteError(w, r, http.StatusUnauthorized, "authentication_required", "Authentication is required")
		return
	}
	actor := domain.Actor{UserID: transportActor.UserID, Administrator: transportActor.Role == "admin", EmailVerified: transportActor.EmailVerified}
	drafts, err := handler.drafting.ListLessonDrafts(r.Context(), actor, r.PathValue("lessonId"))
	if err != nil {
		writeTranscriptError(w, r, err)
		return
	}
	result := draftsDTO{Drafts: make([]draftDTO, 0, len(drafts))}
	for _, item := range drafts {
		draft := item.Draft
		dto := draftDTO{TrackID: draft.TrackID, TrackLanguage: string(draft.TrackLanguage), TrackMultilingual: draft.TrackMultilingual,
			State: draft.State, Stale: draft.Stale(item.TrackLanguage, item.TrackMultilingual), UpdatedAt: item.UpdatedAt,
			AssetID: draft.AssetID, Generation: draft.Generation, Segments: make([]draftSegmentDTO, 0, len(draft.Segments))}
		if draft.Language != "" {
			lang := string(draft.Language)
			dto.Language = &lang
		}
		if draft.Failure != "" {
			dto.Failure = &draft.Failure
		}
		for _, segment := range draft.Segments {
			mapped := draftSegmentDTO{ID: segment.ID, StartMs: segment.StartMs, EndMs: segment.EndMs, Text: segment.Text, Confidence: segment.Confidence}
			if segment.Speaker != "" {
				speaker := segment.Speaker
				mapped.Speaker = &speaker
			}
			dto.Segments = append(dto.Segments, mapped)
		}
		result.Drafts = append(result.Drafts, dto)
	}
	httpserver.WriteJSON(w, http.StatusOK, result)
}

func (handler *DraftHandler) regenerate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")

	transportActor, ok := httpserver.ActorFromContext(r.Context())
	if !ok {
		httpserver.WriteError(w, r, http.StatusUnauthorized, "authentication_required", "Authentication is required")
		return
	}
	actor := domain.Actor{UserID: transportActor.UserID, Administrator: transportActor.Role == "admin", EmailVerified: transportActor.EmailVerified}

	// Reject every byte, including whitespace and chunked bodies with unknown ContentLength.
	if body, err := io.ReadAll(io.LimitReader(r.Body, 1)); err != nil || len(body) != 0 {
		writeTranscriptError(w, r, domain.ErrInvalidRequest)
		return
	}

	err := handler.drafting.Regenerate(r.Context(), actor, r.PathValue("lessonId"), r.PathValue("trackId"))
	if err != nil {
		writeTranscriptError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}
