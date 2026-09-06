package tus

import (
	"context"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"syscall"

	"courses/internal/media/domain"
	"courses/platform/httpserver"
)

const (
	version    = "1.0.0"
	chunkBytes = int64(8 * 1024 * 1024)
)

type uploadService interface {
	HeadUpload(context.Context, string) (domain.VideoAsset, error)
	AppendChunk(context.Context, string, int64, io.Reader) (domain.VideoAsset, error)
	CancelUpload(context.Context, string) error
}

type Server struct{ upload uploadService }

func NewServer(upload uploadService) *Server { return &Server{upload: upload} }

func (server *Server) Register(mux *http.ServeMux) {
	mux.Handle("/uploads/{videoId}", server)
}

func (server *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tusHeaders(w)
	if r.Header.Get("Tus-Resumable") != version {
		httpserver.WriteError(w, r, http.StatusPreconditionFailed, "precondition_failed", "Tus-Resumable must be 1.0.0")
		return
	}
	id := r.PathValue("videoId")
	if !validID(id) {
		httpserver.WriteError(w, r, http.StatusNotFound, "video_not_found", "Upload was not found")
		return
	}
	switch r.Method {
	case http.MethodHead:
		server.head(w, r, id)
	case http.MethodPatch:
		server.patch(w, r, id)
	case http.MethodDelete:
		server.delete(w, r, id)
	default:
		w.Header().Set("Allow", "HEAD, PATCH, DELETE")
		httpserver.WriteError(w, r, http.StatusMethodNotAllowed, "method_not_allowed", "Method is not allowed")
	}
}

func (server *Server) head(w http.ResponseWriter, r *http.Request, id string) {
	asset, err := server.upload.HeadUpload(r.Context(), id)
	if err != nil {
		writeUploadError(w, r, err)
		return
	}
	w.Header().Set("Upload-Offset", strconv.FormatInt(asset.ReceivedBytes, 10))
	w.Header().Set("Upload-Length", strconv.FormatInt(asset.DeclaredBytes, 10))
	w.Header().Set("Upload-Expires", asset.ExpiresAt.UTC().Format(http.TimeFormat))
	w.WriteHeader(http.StatusOK)
}

func (server *Server) patch(w http.ResponseWriter, r *http.Request, id string) {
	if r.Header.Get("Content-Type") != "application/offset+octet-stream" {
		httpserver.WriteError(w, r, http.StatusUnsupportedMediaType, "unsupported_media_type", "Content-Type must be application/offset+octet-stream")
		return
	}
	offset, ok := parseOffset(r.Header.Get("Upload-Offset"))
	if !ok {
		httpserver.WriteError(w, r, http.StatusBadRequest, "invalid_request", "Upload-Offset is invalid")
		return
	}
	body := http.MaxBytesReader(w, r.Body, chunkBytes)
	asset, err := server.upload.AppendChunk(r.Context(), id, offset, body)
	if err != nil {
		writeUploadError(w, r, err)
		return
	}
	w.Header().Set("Upload-Offset", strconv.FormatInt(asset.ReceivedBytes, 10))
	w.Header().Set("Upload-Expires", asset.ExpiresAt.UTC().Format(http.TimeFormat))
	w.WriteHeader(http.StatusNoContent)
}

func (server *Server) delete(w http.ResponseWriter, r *http.Request, id string) {
	if err := server.upload.CancelUpload(r.Context(), id); err != nil {
		writeUploadError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func tusHeaders(w http.ResponseWriter) {
	w.Header().Set("Tus-Resumable", version)
	w.Header().Set("Cache-Control", "no-store")
}

func writeUploadError(w http.ResponseWriter, r *http.Request, err error) {
	var tooLarge *http.MaxBytesError
	switch {
	case errors.Is(err, domain.ErrVideoNotFound):
		httpserver.WriteError(w, r, http.StatusNotFound, "video_not_found", "Upload was not found")
	case errors.Is(err, domain.ErrInvalidStateTransition):
		httpserver.WriteError(w, r, http.StatusConflict, "upload_offset_conflict", "Upload offset conflicts with the stored offset")
	case errors.Is(err, domain.ErrInvalidRequest), errors.As(err, &tooLarge):
		httpserver.WriteError(w, r, http.StatusRequestEntityTooLarge, "request_too_large", "Upload chunk exceeds the allowed size")
	case errors.Is(err, domain.ErrUnsupportedMedia):
		httpserver.WriteError(w, r, http.StatusBadRequest, "unsupported_media", "Upload media is unsupported")
	case errors.Is(err, syscall.ENOSPC):
		httpserver.WriteError(w, r, http.StatusInsufficientStorage, "insufficient_storage", "Media storage has insufficient space")
	default:
		httpserver.WriteError(w, r, http.StatusServiceUnavailable, "dependency_unavailable", "Upload service is temporarily unavailable")
	}
}

func parseOffset(value string) (int64, bool) {
	if value == "" {
		return 0, false
	}
	for index := range value {
		if value[index] < '0' || value[index] > '9' {
			return 0, false
		}
	}
	offset, err := strconv.ParseInt(value, 10, 64)
	return offset, err == nil
}

func validID(value string) bool {
	if len(value) != 36 || value[8] != '-' || value[13] != '-' || value[18] != '-' || value[23] != '-' || strings.ToLower(value) != value || value == "00000000-0000-0000-0000-000000000000" {
		return false
	}
	decoded, err := hex.DecodeString(strings.ReplaceAll(value, "-", ""))
	return err == nil && len(decoded) == 16
}
