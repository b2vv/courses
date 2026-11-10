package tus

import (
	"context"
	"io"
	"net/http"
	"strconv"

	"courses/internal/media/domain"
	"courses/platform/httpserver"
)

type audioUploadService interface {
	HeadUpload(context.Context, string) (domain.AudioTrack, error)
	AppendChunk(context.Context, string, int64, io.Reader) (domain.AudioTrack, error)
	CancelUpload(context.Context, string) error
}

type AudioServer struct{ upload audioUploadService }

func NewAudioServer(upload audioUploadService) *AudioServer { return &AudioServer{upload: upload} }

func (server *AudioServer) Register(mux *http.ServeMux) {
	mux.Handle("/uploads/audio/{trackId}", server)
}

func (server *AudioServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tusHeaders(w)
	if r.Header.Get("Tus-Resumable") != version {
		httpserver.WriteError(w, r, http.StatusPreconditionFailed, "precondition_failed", "Tus-Resumable must be 1.0.0")
		return
	}
	id := r.PathValue("trackId")
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

func (server *AudioServer) head(w http.ResponseWriter, r *http.Request, id string) {
	track, err := server.upload.HeadUpload(r.Context(), id)
	if err != nil {
		writeUploadError(w, r, err)
		return
	}
	w.Header().Set("Upload-Offset", strconv.FormatInt(track.ReceivedBytes, 10))
	w.Header().Set("Upload-Length", strconv.FormatInt(track.DeclaredBytes, 10))
	w.Header().Set("Upload-Expires", track.ExpiresAt.UTC().Format(http.TimeFormat))
	w.WriteHeader(http.StatusOK)
}

func (server *AudioServer) patch(w http.ResponseWriter, r *http.Request, id string) {
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
	track, err := server.upload.AppendChunk(r.Context(), id, offset, body)
	if err != nil {
		writeUploadError(w, r, err)
		return
	}
	w.Header().Set("Upload-Offset", strconv.FormatInt(track.ReceivedBytes, 10))
	w.Header().Set("Upload-Expires", track.ExpiresAt.UTC().Format(http.TimeFormat))
	w.WriteHeader(http.StatusNoContent)
}

func (server *AudioServer) delete(w http.ResponseWriter, r *http.Request, id string) {
	if err := server.upload.CancelUpload(r.Context(), id); err != nil {
		writeUploadError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
