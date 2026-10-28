package http

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"

	"courses/internal/media/domain"
	"courses/platform/httpserver"
)

const (
	maximumMediaWriteBody = 8 * 1024
	mediaChunkBytes       = 8 * 1024 * 1024
	mediaTimeLayout       = "2006-01-02T15:04:05.000000Z07:00"
)

type uploadService interface {
	MintTicket(context.Context, domain.Actor, string, int64, string) (domain.VideoAsset, error)
	ListCourseVideos(context.Context, domain.Actor, string) ([]domain.VideoAsset, error)
	Retry(context.Context, domain.Actor, string) (domain.VideoAsset, error)
}

type Handler struct{ upload uploadService }

type mintVideoRequest struct {
	DeclaredBytes int64  `json:"declaredBytes"`
	OriginalName  string `json:"originalName"`
}

type uploadTicketDTO struct {
	VideoID    string `json:"videoId"`
	UploadURL  string `json:"uploadUrl"`
	ChunkBytes int64  `json:"chunkBytes"`
	MaxBytes   int64  `json:"maxBytes"`
	ExpiresAt  string `json:"expiresAt"`
}

type videoDTO struct {
	LessonID      string               `json:"lessonId"`
	VideoID       string               `json:"videoId"`
	State         domain.State         `json:"state"`
	Generation    int32                `json:"generation"`
	OriginalName  string               `json:"originalName"`
	ReceivedBytes int64                `json:"receivedBytes"`
	DeclaredBytes int64                `json:"declaredBytes"`
	UpdatedAt     string               `json:"updatedAt"`
	FailureReason domain.FailureReason `json:"failureReason,omitempty"`
}

type courseVideosDTO struct {
	Items []videoDTO `json:"items"`
}

func NewHandler(upload uploadService) *Handler { return &Handler{upload: upload} }

func (handler *Handler) RegisterAdmin(
	mux *http.ServeMux,
	authenticate func(http.Handler) http.Handler,
	requireSameOrigin func(http.Handler) http.Handler,
) {
	mux.Handle("POST /api/v1/admin/lessons/{lessonId}/video", requireSameOrigin(authenticate(http.HandlerFunc(handler.mintTicket))))
	mux.Handle("GET /api/v1/admin/courses/{courseId}/videos", authenticate(http.HandlerFunc(handler.listCourseVideos)))
	mux.Handle("POST /api/v1/admin/videos/{videoId}/retry", requireSameOrigin(authenticate(http.HandlerFunc(handler.retry))))
}

func (handler *Handler) mintTicket(w http.ResponseWriter, r *http.Request) {
	mediaHeaders(w)
	actor, ok := mediaActor(w, r)
	if !ok {
		return
	}
	lessonID := r.PathValue("lessonId")
	if !validMediaID(lessonID) || !emptyMediaQuery(r.URL.RawQuery) {
		writeMediaError(w, r, domain.ErrInvalidRequest)
		return
	}
	request, err := decodeMintVideo(w, r)
	if err != nil {
		writeMediaError(w, r, err)
		return
	}
	asset, err := handler.upload.MintTicket(r.Context(), actor, lessonID, request.DeclaredBytes, request.OriginalName)
	if err != nil {
		writeMediaError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusCreated, uploadTicketDTO{
		VideoID: asset.ID, UploadURL: "/uploads/" + asset.ID, ChunkBytes: mediaChunkBytes,
		MaxBytes: domain.MaximumDeclaredBytes, ExpiresAt: formatMediaTime(asset.ExpiresAt),
	})
}

func (handler *Handler) listCourseVideos(w http.ResponseWriter, r *http.Request) {
	mediaHeaders(w)
	actor, ok := mediaActor(w, r)
	if !ok {
		return
	}
	courseID := r.PathValue("courseId")
	if !validMediaID(courseID) || !emptyMediaQuery(r.URL.RawQuery) {
		writeMediaError(w, r, domain.ErrInvalidRequest)
		return
	}
	assets, err := handler.upload.ListCourseVideos(r.Context(), actor, courseID)
	if err != nil {
		writeMediaError(w, r, err)
		return
	}
	items := make([]videoDTO, 0, len(assets))
	for _, asset := range assets {
		items = append(items, mapVideo(asset))
	}
	httpserver.WriteJSON(w, http.StatusOK, courseVideosDTO{Items: items})
}

func (handler *Handler) retry(w http.ResponseWriter, r *http.Request) {
	mediaHeaders(w)
	actor, ok := mediaActor(w, r)
	if !ok {
		return
	}
	videoID := r.PathValue("videoId")
	if !validMediaID(videoID) || !emptyMediaQuery(r.URL.RawQuery) {
		writeMediaError(w, r, domain.ErrInvalidRequest)
		return
	}
	asset, err := handler.upload.Retry(r.Context(), actor, videoID)
	if err != nil {
		writeMediaError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, mapVideo(asset))
}

func mediaActor(w http.ResponseWriter, r *http.Request) (domain.Actor, bool) {
	actor, ok := httpserver.ActorFromContext(r.Context())
	if !ok {
		writeMediaError(w, r, domain.ErrAuthenticationRequired)
		return domain.Actor{}, false
	}
	switch actor.Role {
	case "admin":
		return domain.Actor{UserID: actor.UserID, Administrator: true}, true
	case "instructor":
		return domain.Actor{UserID: actor.UserID}, true
	default:
		writeMediaError(w, r, domain.ErrForbidden)
		return domain.Actor{}, false
	}
}

func decodeMintVideo(w http.ResponseWriter, r *http.Request) (mintVideoRequest, error) {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		return mintVideoRequest{}, domain.ErrInvalidRequest
	}
	r.Body = http.MaxBytesReader(w, r.Body, maximumMediaWriteBody)
	body, err := io.ReadAll(r.Body)
	if err != nil || !utf8.Valid(body) {
		return mintVideoRequest{}, domain.ErrInvalidRequest
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	opening, err := decoder.Token()
	if err != nil || opening != json.Delim('{') {
		return mintVideoRequest{}, domain.ErrInvalidRequest
	}
	var declaredBytes *int64
	var originalName *string
	seen := make(map[string]struct{}, 2)
	for decoder.More() {
		token, err := decoder.Token()
		key, ok := token.(string)
		if err != nil || !ok {
			return mintVideoRequest{}, domain.ErrInvalidRequest
		}
		if _, duplicate := seen[key]; duplicate {
			return mintVideoRequest{}, domain.ErrInvalidRequest
		}
		seen[key] = struct{}{}
		switch key {
		case "declaredBytes":
			err = decoder.Decode(&declaredBytes)
		case "originalName":
			err = decoder.Decode(&originalName)
		default:
			return mintVideoRequest{}, domain.ErrInvalidRequest
		}
		if err != nil {
			return mintVideoRequest{}, domain.ErrInvalidRequest
		}
	}
	closing, err := decoder.Token()
	if err != nil || closing != json.Delim('}') || declaredBytes == nil || originalName == nil || decoder.Decode(&struct{}{}) != io.EOF {
		return mintVideoRequest{}, domain.ErrInvalidRequest
	}
	return mintVideoRequest{DeclaredBytes: *declaredBytes, OriginalName: *originalName}, nil
}

func mapVideo(asset domain.VideoAsset) videoDTO {
	return videoDTO{
		LessonID: asset.LessonID, VideoID: asset.ID, State: asset.State,
		Generation: int32(asset.Generation), OriginalName: asset.OriginalName,
		ReceivedBytes: asset.ReceivedBytes, DeclaredBytes: asset.DeclaredBytes,
		UpdatedAt: formatMediaTime(asset.UpdatedAt), FailureReason: asset.FailureReason,
	}
}

func validMediaID(value string) bool {
	if len(value) != 36 || value[8] != '-' || value[13] != '-' || value[18] != '-' || value[23] != '-' || strings.ToLower(value) != value || value == "00000000-0000-0000-0000-000000000000" {
		return false
	}
	decoded, err := hex.DecodeString(strings.ReplaceAll(value, "-", ""))
	return err == nil && len(decoded) == 16
}

func emptyMediaQuery(raw string) bool {
	query, err := url.ParseQuery(raw)
	return err == nil && len(query) == 0
}

func formatMediaTime(value time.Time) string {
	return value.UTC().Truncate(time.Microsecond).Format(mediaTimeLayout)
}

func mediaHeaders(w http.ResponseWriter) { w.Header().Set("Cache-Control", "no-store") }

func writeMediaError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidRequest):
		httpserver.WriteError(w, r, http.StatusBadRequest, "invalid_request", "Request is invalid")
	case errors.Is(err, domain.ErrUnsupportedMedia):
		httpserver.WriteError(w, r, http.StatusBadRequest, "unsupported_media", "Upload media is unsupported")
	case errors.Is(err, domain.ErrUploadTooLarge):
		httpserver.WriteError(w, r, http.StatusBadRequest, "upload_too_large", "Upload exceeds the maximum size")
	case errors.Is(err, domain.ErrAuthenticationRequired):
		httpserver.WriteError(w, r, http.StatusUnauthorized, "authentication_required", "Authentication is required")
	case errors.Is(err, domain.ErrForbidden):
		httpserver.WriteError(w, r, http.StatusForbidden, "forbidden", "Permission is required")
	case errors.Is(err, domain.ErrCourseForbidden):
		httpserver.WriteError(w, r, http.StatusForbidden, "course_forbidden", "Course access is forbidden")
	case errors.Is(err, domain.ErrOriginForbidden):
		httpserver.WriteError(w, r, http.StatusForbidden, "origin_forbidden", "Request origin is not allowed")
	case errors.Is(err, domain.ErrLessonNotFound):
		httpserver.WriteError(w, r, http.StatusNotFound, "lesson_not_found", "Lesson was not found")
	case errors.Is(err, domain.ErrVideoNotFound):
		httpserver.WriteError(w, r, http.StatusNotFound, "video_not_found", "Video was not found")
	case errors.Is(err, domain.ErrUploadInProgress):
		httpserver.WriteError(w, r, http.StatusConflict, "upload_in_progress", "A video upload is already in progress")
	case errors.Is(err, domain.ErrVideoNotRetryable):
		httpserver.WriteError(w, r, http.StatusConflict, "video_not_retryable", "Video cannot be retried")
	case errors.Is(err, domain.ErrDependencyUnavailable):
		httpserver.WriteError(w, r, http.StatusServiceUnavailable, "dependency_unavailable", "Media service is temporarily unavailable")
	default:
		httpserver.WriteError(w, r, http.StatusInternalServerError, "internal_error", "Unexpected server error")
	}
}
