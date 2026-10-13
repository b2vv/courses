package http

import (
	"bytes"
	"context"
	"errors"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"

	"courses/internal/image/domain"
	"courses/platform/httpserver"
	_ "golang.org/x/image/webp"
)

// LB10 tunes nginx client_max_body_size above the 5 MiB image limit by the envelope, or genuine 5-MiB images are rejected at the edge.
const maxImageUploadBody = domain.MaximumImageBytes + (512 << 10)

type uploadService interface {
	Upload(ctx context.Context, actor domain.Actor, lessonID, contentType string, width, height int, data []byte) (domain.Image, error)
}

type authorizeService interface {
	Authorize(ctx context.Context, actor domain.Actor, imageID string) error
}

type Handler struct {
	uploader      uploadService
	authorization authorizeService
}

func NewHandler(upload uploadService, authorize authorizeService) *Handler {
	return &Handler{uploader: upload, authorization: authorize}
}

func (handler *Handler) RegisterAdmin(mux *http.ServeMux, authenticate, requireSameOrigin func(http.Handler) http.Handler) {
	mux.Handle("POST /api/v1/admin/lessons/{id}/images", requireSameOrigin(authenticate(http.HandlerFunc(handler.upload))))
}

func (handler *Handler) RegisterInternal(mux *http.ServeMux, authenticate func(http.Handler) http.Handler) {
	mux.Handle("GET /internal/image/authorize", normalizeAuthenticationDenial(authenticate(http.HandlerFunc(handler.authorize))))
}

type uploadResponse struct {
	ID          string `json:"id"`
	ContentType string `json:"contentType"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
}

func (handler *Handler) upload(w http.ResponseWriter, r *http.Request) {
	transportActor, ok := httpserver.ActorFromContext(r.Context())
	if !ok {
		writeUploadError(w, r, domain.ErrForbidden)
		return
	}
	actor := domain.Actor{UserID: transportActor.UserID, Administrator: transportActor.Role == "admin", EmailVerified: transportActor.EmailVerified}
	lessonID := r.PathValue("id")
	r.Body = http.MaxBytesReader(w, r.Body, maxImageUploadBody)
	file, _, err := r.FormFile("file")
	if err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			writeUploadError(w, r, domain.ErrImageTooLarge)
		} else {
			httpserver.WriteError(w, r, http.StatusBadRequest, "invalid_request", "expected a multipart form field named file")
		}
		return
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			writeUploadError(w, r, domain.ErrImageTooLarge)
		} else {
			httpserver.WriteError(w, r, http.StatusBadRequest, "invalid_request", "Image file could not be read")
		}
		return
	}
	head := data
	if len(head) > domain.SniffBytes {
		head = data[:domain.SniffBytes]
	}
	contentType, err := domain.DetectContentType(head)
	if err != nil {
		writeUploadError(w, r, err)
		return
	}
	config, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		writeUploadError(w, r, domain.ErrInvalidImage)
		return
	}
	img, err := handler.uploader.Upload(r.Context(), actor, lessonID, contentType, config.Width, config.Height, data)
	if err != nil {
		writeUploadError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusCreated, uploadResponse{ID: img.ID, ContentType: img.ContentType, Width: img.Width, Height: img.Height})
}

func writeUploadError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, domain.ErrLessonNotFound):
		httpserver.WriteError(w, r, http.StatusNotFound, "lesson_not_found", "Lesson was not found")
	case errors.Is(err, domain.ErrForbidden):
		httpserver.WriteError(w, r, http.StatusForbidden, "forbidden", "Permission is required")
	case errors.Is(err, domain.ErrUnsupportedImage):
		httpserver.WriteError(w, r, http.StatusUnsupportedMediaType, "unsupported_image", "Image format is not supported")
	case errors.Is(err, domain.ErrImageTooLarge):
		httpserver.WriteError(w, r, http.StatusRequestEntityTooLarge, "image_too_large", "Image is too large")
	case errors.Is(err, domain.ErrImageTooManyPixels):
		httpserver.WriteError(w, r, http.StatusBadRequest, "image_too_many_pixels", "Image has too many pixels")
	case errors.Is(err, domain.ErrInvalidImage):
		httpserver.WriteError(w, r, http.StatusBadRequest, "invalid_image", "Image is invalid")
	default:
		httpserver.WriteError(w, r, http.StatusServiceUnavailable, "dependency_unavailable", "Image upload is temporarily unavailable")
	}
}

func (handler *Handler) authorize(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	transportActor, ok := httpserver.ActorFromContext(r.Context())
	if !ok {
		httpserver.WriteError(w, r, http.StatusForbidden, "forbidden", "Image access is not authorized")
		return
	}
	actor := domain.Actor{UserID: transportActor.UserID, Administrator: transportActor.Role == "admin", EmailVerified: transportActor.EmailVerified}
	err := handler.authorization.Authorize(r.Context(), actor, r.Header.Get("X-Image-Id"))
	switch {
	case err == nil:
		w.WriteHeader(http.StatusNoContent)
	case errors.Is(err, domain.ErrEmailNotVerified):
		httpserver.WriteError(w, r, http.StatusForbidden, "email_not_verified", "Email verification is required")
	case errors.Is(err, domain.ErrForbidden), errors.Is(err, domain.ErrImageNotFound):
		// This endpoint only answers nginx auth_request, which reads 2xx as allow and 401/403 as
		// deny; any other status becomes a 500 it turns into the serve location's 503. A missing
		// image is a denial, and collapsing it into 403 (rather than 404) also keeps a denied
		// request from revealing whether the image exists.
		httpserver.WriteError(w, r, http.StatusForbidden, "forbidden", "Image access is not authorized")
	default:
		httpserver.WriteError(w, r, http.StatusServiceUnavailable, "dependency_unavailable", "Image authorization is temporarily unavailable")
	}
}

func normalizeAuthenticationDenial(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(&authenticationWriter{ResponseWriter: w}, r)
	})
}

type authenticationWriter struct{ http.ResponseWriter }

func (writer *authenticationWriter) WriteHeader(status int) {
	if status == http.StatusUnauthorized {
		status = http.StatusForbidden
	}
	writer.ResponseWriter.WriteHeader(status)
}
