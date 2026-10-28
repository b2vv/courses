package http

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"unicode/utf8"

	"courses/internal/playback/domain"
	"courses/platform/httpserver"
)

const maximumPlaybackBody = 8 * 1024

var errInvalidRequest = errors.New("invalid playback request")

type authorizationService interface {
	Authorize(context.Context, domain.Actor, string, string) error
}

type sessionService interface {
	Play(context.Context, domain.Actor, string, string) (domain.LessonPlayback, error)
	Heartbeat(context.Context, domain.Actor, string, string, domain.ProgressUpdate) error
}

type Handler struct {
	authorization authorizationService
	sessions      sessionService
	config        Config
	observeAV1    func(string)
}

func NewHandler(authorization authorizationService, sessions sessionService, config Config, observeAV1 func(string)) *Handler {
	return &Handler{authorization: authorization, sessions: sessions, config: config, observeAV1: observeAV1}
}

func (handler *Handler) Register(mux *http.ServeMux, authenticate, requireSameOrigin func(http.Handler) http.Handler) {
	endpoint := authenticate(http.HandlerFunc(handler.authorize))
	mux.Handle("GET /internal/playback/authorize", normalizeAuthenticationDenial(endpoint))
	mux.Handle("POST /api/v1/playback/sessions", requireSameOrigin(authenticate(http.HandlerFunc(handler.play))))
	mux.Handle("POST /api/v1/playback/heartbeat", requireSameOrigin(authenticate(http.HandlerFunc(handler.heartbeat))))
}

type av1Support string

const (
	av1Supported   av1Support = "supported"
	av1Unsupported av1Support = "unsupported"
	av1ProbeFailed av1Support = "probe_failed"
)

// av1SupportValues is the ordered set used by validation and the HTTP contract.
var av1SupportValues = []av1Support{av1Supported, av1Unsupported, av1ProbeFailed}

type playRequest struct {
	LessonID   string      `json:"lessonId"`
	AV1Support *av1Support `json:"av1Support,omitempty"`
}

type heartbeatRequest struct {
	PlaybackID      string                `json:"playbackId"`
	LessonID        string                `json:"lessonId"`
	PositionSeconds float64               `json:"positionSeconds"`
	ClientSequence  int64                 `json:"clientSeq"`
	Reason          domain.ProgressReason `json:"reason"`
}

type playResponse struct {
	PlaybackID          string `json:"playbackId"`
	MediaPath           string `json:"mediaPath"`
	ClientSequence      int64  `json:"clientSeq"`
	HeartbeatSeconds    int    `json:"heartbeatSeconds"`
	OwnershipTTLSeconds int    `json:"ownershipTtlSeconds"`
}

func (handler *Handler) play(w http.ResponseWriter, r *http.Request) {
	playbackHeaders(w)
	if !emptyPlaybackQuery(r.URL.RawQuery) {
		writePlayError(w, r, errInvalidRequest)
		return
	}
	actor, ok := playbackActor(r)
	if !ok {
		writePlayError(w, r, domain.ErrForbidden)
		return
	}
	request, err := decodePlay(w, r)
	if err != nil || !validUUID(request.LessonID) {
		writePlayError(w, r, errInvalidRequest)
		return
	}
	if request.AV1Support != nil {
		handler.observeAV1(string(*request.AV1Support))
	}
	playbackID, err := freshPlaybackID()
	if err != nil {
		writePlayError(w, r, err)
		return
	}
	playback, err := handler.sessions.Play(r.Context(), actor, request.LessonID, playbackID)
	if err != nil {
		writePlayError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusCreated, playResponse{
		PlaybackID:          playbackID,
		MediaPath:           "/media/p/" + playbackID + "/hls/" + playback.VideoID + "/master.m3u8",
		ClientSequence:      playback.ClientSequence,
		HeartbeatSeconds:    handler.config.HeartbeatSeconds(),
		OwnershipTTLSeconds: handler.config.OwnershipTTLSeconds(),
	})
}

func (handler *Handler) heartbeat(w http.ResponseWriter, r *http.Request) {
	playbackHeaders(w)
	if !emptyPlaybackQuery(r.URL.RawQuery) {
		writePlaybackError(w, r, errInvalidRequest)
		return
	}
	actor, ok := playbackActor(r)
	if !ok {
		writePlaybackError(w, r, domain.ErrForbidden)
		return
	}
	request, err := decodeHeartbeat(w, r)
	if err != nil || !validPlaybackID(request.PlaybackID) || !validUUID(request.LessonID) {
		writePlaybackError(w, r, errInvalidRequest)
		return
	}
	update, err := domain.NewProgressUpdate(request.PositionSeconds, request.ClientSequence, string(request.Reason))
	if err != nil {
		writePlaybackError(w, r, err)
		return
	}
	if err := handler.sessions.Heartbeat(r.Context(), actor, request.PlaybackID, request.LessonID, update); err != nil {
		writePlaybackError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (handler *Handler) authorize(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	transportActor, ok := httpserver.ActorFromContext(r.Context())
	if !ok {
		writeAuthorizationError(w, r, http.StatusForbidden)
		return
	}
	actor := domain.Actor{UserID: transportActor.UserID, Administrator: transportActor.Role == "admin", EmailVerified: transportActor.EmailVerified}
	err := handler.authorization.Authorize(r.Context(), actor, r.Header.Get("X-Playback-Id"), r.Header.Get("X-Video-Id"))
	if err == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if errors.Is(err, domain.ErrForbidden) {
		writeAuthorizationError(w, r, http.StatusForbidden)
		return
	}
	if errors.Is(err, domain.ErrPurchaseRequired) {
		httpserver.WriteError(w, r, http.StatusForbidden, "course_purchase_required", "Purchase is required to access this course")
		return
	}
	if errors.Is(err, domain.ErrEmailNotVerified) {
		httpserver.WriteError(w, r, http.StatusForbidden, "email_not_verified", "Email verification is required")
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

func decodePlay(w http.ResponseWriter, r *http.Request) (playRequest, error) {
	var result playRequest
	err := decodePlaybackObject(w, r, map[string]func(*json.Decoder) error{
		"lessonId": func(decoder *json.Decoder) error { return decoder.Decode(&result.LessonID) },
	}, map[string]func(*json.Decoder) error{
		"av1Support": func(decoder *json.Decoder) error {
			var value *av1Support
			if err := decoder.Decode(&value); err != nil || value == nil {
				return errInvalidRequest
			}
			result.AV1Support = value
			return nil
		},
	})
	if err != nil || result.LessonID == "" || !validAV1Support(result.AV1Support) {
		return playRequest{}, errInvalidRequest
	}
	return result, nil
}

func decodeHeartbeat(w http.ResponseWriter, r *http.Request) (heartbeatRequest, error) {
	var (
		result   heartbeatRequest
		position *float64
		sequence *int64
	)
	err := decodePlaybackObject(w, r, map[string]func(*json.Decoder) error{
		"playbackId":      func(decoder *json.Decoder) error { return decoder.Decode(&result.PlaybackID) },
		"lessonId":        func(decoder *json.Decoder) error { return decoder.Decode(&result.LessonID) },
		"positionSeconds": func(decoder *json.Decoder) error { return decoder.Decode(&position) },
		"clientSeq":       func(decoder *json.Decoder) error { return decoder.Decode(&sequence) },
		"reason":          func(decoder *json.Decoder) error { return decoder.Decode(&result.Reason) },
	}, nil)
	if err != nil || result.PlaybackID == "" || result.LessonID == "" || position == nil || sequence == nil || result.Reason == "" {
		return heartbeatRequest{}, errInvalidRequest
	}
	result.PositionSeconds, result.ClientSequence = *position, *sequence
	return result, nil
}

func validAV1Support(value *av1Support) bool {
	return value == nil || slices.Contains(av1SupportValues, *value)
}

func decodePlaybackObject(w http.ResponseWriter, r *http.Request, required, optional map[string]func(*json.Decoder) error) error {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		return errInvalidRequest
	}
	r.Body = http.MaxBytesReader(w, r.Body, maximumPlaybackBody)
	body, err := io.ReadAll(r.Body)
	if err != nil || !utf8.Valid(body) {
		return errInvalidRequest
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	opening, err := decoder.Token()
	if err != nil || opening != json.Delim('{') {
		return errInvalidRequest
	}
	seen := make(map[string]struct{}, len(required)+len(optional))
	for decoder.More() {
		token, err := decoder.Token()
		key, ok := token.(string)
		decode, known := required[key]
		if !known {
			decode, known = optional[key]
		}
		if err != nil || !ok || !known {
			return errInvalidRequest
		}
		if _, duplicate := seen[key]; duplicate {
			return errInvalidRequest
		}
		seen[key] = struct{}{}
		if err := decode(decoder); err != nil {
			return errInvalidRequest
		}
	}
	closing, err := decoder.Token()
	if err != nil || closing != json.Delim('}') || decoder.Decode(&struct{}{}) != io.EOF {
		return errInvalidRequest
	}
	for key := range required {
		if _, ok := seen[key]; !ok {
			return errInvalidRequest
		}
	}
	return nil
}

func playbackActor(r *http.Request) (domain.Actor, bool) {
	actor, ok := httpserver.ActorFromContext(r.Context())
	if !ok || actor.UserID == "" {
		return domain.Actor{}, false
	}
	return domain.Actor{UserID: actor.UserID, Administrator: actor.Role == "admin", EmailVerified: actor.EmailVerified}, true
}

func freshPlaybackID() (string, error) {
	value := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, value); err != nil {
		return "", err
	}
	return hex.EncodeToString(value), nil
}

func validPlaybackID(value string) bool {
	if len(value) != 32 || strings.ToLower(value) != value {
		return false
	}
	decoded, err := hex.DecodeString(value)
	return err == nil && len(decoded) == 16
}

func validUUID(value string) bool {
	if len(value) != 36 || value[8] != '-' || value[13] != '-' || value[18] != '-' || value[23] != '-' || strings.ToLower(value) != value || value == "00000000-0000-0000-0000-000000000000" {
		return false
	}
	decoded, err := hex.DecodeString(strings.ReplaceAll(value, "-", ""))
	return err == nil && len(decoded) == 16
}

func emptyPlaybackQuery(raw string) bool {
	query, err := url.ParseQuery(raw)
	return err == nil && len(query) == 0
}

func playbackHeaders(w http.ResponseWriter) { w.Header().Set("Cache-Control", "no-store") }

func writePlayError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, domain.ErrPlaybackTakenOver) || errors.Is(err, domain.ErrPlaybackSessionExpired) {
		httpserver.WriteError(w, r, http.StatusInternalServerError, "internal_error", "Unexpected server error")
		return
	}
	writePlaybackError(w, r, err)
}

func writePlaybackError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, errInvalidRequest), errors.Is(err, domain.ErrInvalidProgress):
		httpserver.WriteError(w, r, http.StatusBadRequest, "invalid_request", "Request is invalid")
	case errors.Is(err, domain.ErrForbidden):
		httpserver.WriteError(w, r, http.StatusForbidden, "playback_forbidden", "Playback is not authorized")
	case errors.Is(err, domain.ErrPurchaseRequired):
		httpserver.WriteError(w, r, http.StatusForbidden, "course_purchase_required", "Purchase is required to access this course")
	case errors.Is(err, domain.ErrEmailNotVerified):
		httpserver.WriteError(w, r, http.StatusForbidden, "email_not_verified", "Email verification is required")
	case errors.Is(err, domain.ErrLessonNotFound):
		httpserver.WriteError(w, r, http.StatusNotFound, "lesson_not_found", "Lesson was not found")
	case errors.Is(err, domain.ErrVideoNotReady):
		httpserver.WriteError(w, r, http.StatusNotFound, "video_not_ready", "Lesson video is not ready")
	case errors.Is(err, domain.ErrPlaybackTakenOver):
		httpserver.WriteError(w, r, http.StatusConflict, "playback_taken_over", "Playback continued on another device")
	case errors.Is(err, domain.ErrPlaybackSessionExpired):
		httpserver.WriteError(w, r, http.StatusGone, "playback_session_expired", "Playback session expired")
	case errors.Is(err, domain.ErrDependencyUnavailable):
		httpserver.WriteError(w, r, http.StatusServiceUnavailable, "dependency_unavailable", "Playback is temporarily unavailable")
	default:
		httpserver.WriteError(w, r, http.StatusInternalServerError, "internal_error", "Unexpected server error")
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
