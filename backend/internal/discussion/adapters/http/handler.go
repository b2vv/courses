package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"time"

	"courses/internal/discussion/application"
	"courses/internal/discussion/domain"
	"courses/platform/httpserver"
)

const maximumCommentBody = 16 * 1024

type Handler struct{ service *application.Service }

func NewHandler(service *application.Service) *Handler { return &Handler{service: service} }

func (handler *Handler) Register(mux *http.ServeMux, authenticate, sameOrigin func(http.Handler) http.Handler) {
	mux.Handle("GET /api/v1/learn/lessons/{lessonId}/comments", authenticate(http.HandlerFunc(handler.list)))
	mux.Handle("POST /api/v1/learn/lessons/{lessonId}/comments", sameOrigin(authenticate(http.HandlerFunc(handler.publish))))
	mux.Handle("PATCH /api/v1/learn/comments/{commentId}", sameOrigin(authenticate(http.HandlerFunc(handler.update))))
	mux.Handle("DELETE /api/v1/learn/comments/{commentId}", sameOrigin(authenticate(http.HandlerFunc(handler.remove))))
}

type authorDTO struct {
	DisplayName *string `json:"displayName"`
}

type commentDTO struct {
	ID         string     `json:"id"`
	Author     authorDTO  `json:"author"`
	Body       *string    `json:"body"`
	PositionMs *int64     `json:"positionMs"`
	CreatedAt  time.Time  `json:"createdAt"`
	EditedAt   *time.Time `json:"editedAt"`
	Removed    bool       `json:"removed"`
	CanEdit    bool       `json:"canEdit"`
	CanRemove  bool       `json:"canRemove"`
}

type threadDTO struct {
	Root    commentDTO   `json:"root"`
	Replies []commentDTO `json:"replies"`
}

type pageDTO struct {
	Threads    []threadDTO `json:"threads"`
	NextCursor string      `json:"nextCursor"`
}

type publishRequest struct {
	Body       string  `json:"body"`
	PositionMs *int64  `json:"positionMs"`
	ParentID   *string `json:"parentId"`
}

type updateRequest struct {
	Body string `json:"body"`
}

func (handler *Handler) list(w http.ResponseWriter, r *http.Request) {
	noStore(w)
	query, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil || hasUnexpectedQuery(query, "cursor") || len(query["cursor"]) > 1 {
		writeError(w, r, http.StatusBadRequest, domain.ErrInvalidCursor)
		return
	}
	actor, ok := actorFromRequest(w, r)
	if !ok {
		return
	}
	page, err := handler.service.List(r.Context(), actor, r.PathValue("lessonId"), query.Get("cursor"))
	if err != nil {
		writeListError(w, r, err)
		return
	}
	threads := make([]threadDTO, len(page.Threads))
	for i, thread := range page.Threads {
		threads[i] = threadDTO{Root: mapComment(thread.Root), Replies: make([]commentDTO, len(thread.Replies))}
		for j, reply := range thread.Replies {
			threads[i].Replies[j] = mapComment(reply)
		}
	}
	httpserver.WriteJSON(w, http.StatusOK, pageDTO{Threads: threads, NextCursor: page.NextCursor})
}

func (handler *Handler) publish(w http.ResponseWriter, r *http.Request) {
	noStore(w)
	actor, ok := actorFromRequest(w, r)
	if !ok {
		return
	}
	input, err := decodeJSON[publishRequest](w, r)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, err)
		return
	}
	comment, err := handler.service.Publish(r.Context(), actor, application.PublishRequest{
		LessonID: r.PathValue("lessonId"), Body: input.Body, PositionMs: input.PositionMs, ParentID: input.ParentID,
	})
	if err != nil {
		writePublishError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusCreated, mapComment(comment))
}

func (handler *Handler) update(w http.ResponseWriter, r *http.Request) {
	noStore(w)
	actor, ok := actorFromRequest(w, r)
	if !ok {
		return
	}
	input, err := decodeJSON[updateRequest](w, r)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, err)
		return
	}
	comment, err := handler.service.Update(r.Context(), actor, r.PathValue("commentId"), input.Body)
	if err != nil {
		writeMutationError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, mapComment(comment))
}

func (handler *Handler) remove(w http.ResponseWriter, r *http.Request) {
	noStore(w)
	actor, ok := actorFromRequest(w, r)
	if !ok {
		return
	}
	if err := handler.service.Remove(r.Context(), actor, r.PathValue("commentId")); err != nil {
		writeMutationError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func actorFromRequest(w http.ResponseWriter, r *http.Request) (domain.Actor, bool) {
	transportActor, ok := httpserver.ActorFromContext(r.Context())
	if !ok {
		httpserver.WriteError(w, r, http.StatusUnauthorized, "authentication_required", "Authentication is required")
		return domain.Actor{}, false
	}
	return domain.Actor{UserID: transportActor.UserID, Administrator: transportActor.Role == "admin"}, true
}

func decodeJSON[T any](w http.ResponseWriter, r *http.Request) (T, error) {
	var result T
	if r.Header.Get("Content-Type") != "application/json" {
		return result, domain.ErrInvalidRequest
	}
	r.Body = http.MaxBytesReader(w, r.Body, maximumCommentBody)
	body, err := io.ReadAll(r.Body)
	if err != nil || validateUniqueJSONObject(body) != nil {
		return result, domain.ErrInvalidRequest
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&result); err != nil {
		return result, domain.ErrInvalidRequest
	}
	if decoder.Decode(&struct{}{}) != io.EOF {
		return result, domain.ErrInvalidRequest
	}
	return result, nil
}

func validateUniqueJSONObject(body []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(body))
	token, err := decoder.Token()
	if err != nil || token != json.Delim('{') {
		return domain.ErrInvalidRequest
	}
	seen := map[string]struct{}{}
	for decoder.More() {
		key, err := decoder.Token()
		name, ok := key.(string)
		if err != nil || !ok {
			return domain.ErrInvalidRequest
		}
		if _, exists := seen[name]; exists {
			return domain.ErrInvalidRequest
		}
		seen[name] = struct{}{}
		var value json.RawMessage
		if err := decoder.Decode(&value); err != nil {
			return domain.ErrInvalidRequest
		}
	}
	closing, err := decoder.Token()
	if err != nil || closing != json.Delim('}') {
		return domain.ErrInvalidRequest
	}
	return nil
}

func mapComment(comment application.Comment) commentDTO {
	return commentDTO{ID: comment.ID, Author: authorDTO{DisplayName: comment.Author}, Body: comment.Body,
		PositionMs: comment.PositionMs, CreatedAt: comment.CreatedAt, EditedAt: comment.EditedAt,
		Removed: comment.Removed, CanEdit: comment.CanEdit, CanRemove: comment.CanRemove}
}

func noStore(w http.ResponseWriter) { w.Header().Set("Cache-Control", "no-store") }
func hasUnexpectedQuery(query url.Values, allowed ...string) bool {
	for key := range query {
		found := false
		for _, candidate := range allowed {
			if key == candidate {
				found = true
				break
			}
		}
		if !found {
			return true
		}
	}
	return false
}

func writeListError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, domain.ErrNotFound) {
		writeError(w, r, http.StatusNotFound, errors.New("lesson_not_found"))
		return
	}
	writeError(w, r, statusFor(err), err)
}

func writePublishError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, domain.ErrNotFound) {
		writeError(w, r, http.StatusNotFound, errors.New("lesson_not_found"))
		return
	}
	writeError(w, r, statusFor(err), err)
}

func writeMutationError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, domain.ErrNotFound) {
		writeError(w, r, http.StatusNotFound, errors.New("comment_not_found"))
		return
	}
	writeError(w, r, statusFor(err), err)
}

func statusFor(err error) int {
	switch {
	case errors.Is(err, domain.ErrForbidden), errors.Is(err, application.ErrCommentForbidden):
		return http.StatusForbidden
	case errors.Is(err, application.ErrCommentRemoved), errors.Is(err, domain.ErrThreadFull):
		return http.StatusConflict
	case errors.Is(err, domain.ErrDependencyUnavailable):
		return http.StatusServiceUnavailable
	case errors.Is(err, domain.ErrBodyInvalid), errors.Is(err, domain.ErrPositionInvalid), errors.Is(err, domain.ErrParentInvalid), errors.Is(err, domain.ErrInvalidCursor), errors.Is(err, domain.ErrInvalidRequest):
		return http.StatusBadRequest
	default:
		return http.StatusBadRequest
	}
}

func writeError(w http.ResponseWriter, r *http.Request, status int, err error) {
	code := err.Error()
	message := "Request is not valid"
	switch code {
	case "course_forbidden":
		message = "Course access is forbidden"
	case "comment_forbidden":
		message = "Comment access is forbidden"
	case "comment_removed":
		message = "The comment has been removed"
	case "comment_not_found":
		message = "Comment was not found"
	case "lesson_not_found":
		message = "Lesson was not found"
	case "comment_thread_full":
		message = "This comment thread is full"
	case "comment_cursor_invalid":
		message = "The comment cursor is invalid"
	case "comment_body_invalid":
		message = "The comment body is invalid"
	case "comment_position_invalid":
		message = "The comment position is invalid"
	case "comment_parent_invalid":
		message = "The comment parent is invalid"
	case "origin_forbidden":
		status = http.StatusForbidden
		message = "Origin is forbidden"
	case "dependency_unavailable":
		message = "A dependency is unavailable"
	}
	httpserver.WriteError(w, r, status, code, message)
}
