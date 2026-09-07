package http

import (
	"context"
	"errors"
	"net/http"

	"courses/internal/playback/domain"
	"courses/platform/httpserver"
)

type authorizationService interface {
	Authorize(context.Context, domain.Actor, string, string) error
}

type Handler struct{ authorization authorizationService }

func NewHandler(authorization authorizationService) *Handler {
	return &Handler{authorization: authorization}
}

func (handler *Handler) Register(mux *http.ServeMux, authenticate func(http.Handler) http.Handler) {
	endpoint := authenticate(http.HandlerFunc(handler.authorize))
	mux.Handle("GET /internal/playback/authorize", normalizeAuthenticationDenial(endpoint))
}

func (handler *Handler) authorize(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	transportActor, ok := httpserver.ActorFromContext(r.Context())
	if !ok {
		writeAuthorizationError(w, r, http.StatusForbidden)
		return
	}
	actor := domain.Actor{UserID: transportActor.UserID, Administrator: transportActor.Role == "admin"}
	err := handler.authorization.Authorize(r.Context(), actor, r.Header.Get("X-Playback-Id"), r.Header.Get("X-Video-Id"))
	if err == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if errors.Is(err, domain.ErrForbidden) {
		writeAuthorizationError(w, r, http.StatusForbidden)
		return
	}
	writeAuthorizationError(w, r, http.StatusServiceUnavailable)
}

func writeAuthorizationError(w http.ResponseWriter, r *http.Request, status int) {
	if status == http.StatusForbidden {
		httpserver.WriteError(w, r, status, "playback_forbidden", "Playback is not authorized")
		return
	}
	httpserver.WriteError(w, r, status, "dependency_unavailable", "Playback authorization is temporarily unavailable")
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
