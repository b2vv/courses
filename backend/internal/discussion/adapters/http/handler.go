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
	mux.Handle("GET /api/v1/courses/{slug}/reviews", http.HandlerFunc(handler.publicReviews))
	mux.Handle("GET /api/v1/learn/courses/{courseId}/reviews", authenticate(http.HandlerFunc(handler.learningReviews)))
	mux.Handle("GET /api/v1/learn/courses/{courseId}/review", authenticate(http.HandlerFunc(handler.ownReview)))
	mux.Handle("PUT /api/v1/learn/courses/{courseId}/review", sameOrigin(authenticate(http.HandlerFunc(handler.putReview))))
	mux.Handle("DELETE /api/v1/learn/courses/{courseId}/review", sameOrigin(authenticate(http.HandlerFunc(handler.deleteReview))))
	mux.Handle("POST /api/v1/learn/reviews/{reviewId}/reports", sameOrigin(authenticate(http.HandlerFunc(handler.reportReview))))
	mux.Handle("GET /api/v1/learn/lessons/{lessonId}/comments", authenticate(http.HandlerFunc(handler.list)))
	mux.Handle("POST /api/v1/learn/lessons/{lessonId}/comments", sameOrigin(authenticate(http.HandlerFunc(handler.publish))))
	mux.Handle("PATCH /api/v1/learn/comments/{commentId}", sameOrigin(authenticate(http.HandlerFunc(handler.update))))
	mux.Handle("DELETE /api/v1/learn/comments/{commentId}", sameOrigin(authenticate(http.HandlerFunc(handler.remove))))
}

func (handler *Handler) RegisterTeach(mux *http.ServeMux, authenticate, sameOrigin func(http.Handler) http.Handler) {
	mux.Handle("GET /api/v1/teach/courses/{courseId}/comments", authenticate(http.HandlerFunc(handler.listCourseComments)))
	mux.Handle("GET /api/v1/teach/courses/{courseId}/moderation-queue", authenticate(http.HandlerFunc(handler.queue)))
	mux.Handle("POST /api/v1/teach/comments/{commentId}/dismiss-report", sameOrigin(authenticate(http.HandlerFunc(handler.dismissReports))))
	mux.Handle("POST /api/v1/learn/comments/{commentId}/reports", sameOrigin(authenticate(http.HandlerFunc(handler.reportComment))))
	mux.Handle("POST /api/v1/teach/reviews/{reviewId}/remove", sameOrigin(authenticate(http.HandlerFunc(handler.removeReview))))
	mux.Handle("POST /api/v1/teach/reviews/{reviewId}/dismiss-report", sameOrigin(authenticate(http.HandlerFunc(handler.dismissReviewReports))))
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

type dismissReportsRequest struct {
	Reason string `json:"reason"`
}

type moderationCommentDTO struct {
	ID          string     `json:"id"`
	LessonID    string     `json:"lessonId"`
	LessonTitle string     `json:"lessonTitle"`
	ParentID    *string    `json:"parentId"`
	Author      authorDTO  `json:"author"`
	Body        *string    `json:"body"`
	PositionMs  *int64     `json:"positionMs"`
	CreatedAt   time.Time  `json:"createdAt"`
	EditedAt    *time.Time `json:"editedAt"`
	Removed     bool       `json:"removed"`
}

type courseCommentsDTO struct {
	Comments   []moderationCommentDTO `json:"comments"`
	NextCursor string                 `json:"nextCursor"`
}

type queueItemDTO struct {
	Kind        string                `json:"kind"`
	Comment     *moderationCommentDTO `json:"comment"`
	Review      *reviewDTO            `json:"review"`
	OpenReports int                   `json:"openReports"`
}

type reviewDTO struct {
	ID                string     `json:"id"`
	AuthorDisplayName *string    `json:"authorDisplayName"`
	Stars             int        `json:"stars"`
	Body              string     `json:"body"`
	CompletedLessons  int        `json:"completedLessons"`
	TotalLessons      int        `json:"totalLessons"`
	CreatedAt         time.Time  `json:"createdAt"`
	EditedAt          *time.Time `json:"editedAt"`
}
type reviewRatingDTO struct {
	Average float64 `json:"average"`
	Count   int     `json:"count"`
}
type publicReviewsDTO struct {
	Rating     *reviewRatingDTO `json:"rating"`
	Items      []reviewDTO      `json:"items"`
	NextCursor *string          `json:"nextCursor"`
}
type ownReviewDTO struct {
	Review      *reviewDTO `json:"review"`
	Eligibility string     `json:"eligibility"`
}
type putReviewRequest struct {
	Stars int    `json:"stars"`
	Body  string `json:"body"`
}

type queueDTO struct {
	Items      []queueItemDTO `json:"items"`
	NextCursor string         `json:"nextCursor"`
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

func (handler *Handler) reportComment(w http.ResponseWriter, r *http.Request) {
	noStore(w)
	actor, ok := actorFromRequest(w, r)
	if !ok {
		return
	}
	if err := handler.service.ReportComment(r.Context(), actor, r.PathValue("commentId")); err != nil {
		writeReportError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (handler *Handler) listCourseComments(w http.ResponseWriter, r *http.Request) {
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
	page, err := handler.service.ListCourseComments(r.Context(), actor, r.PathValue("courseId"), query.Get("cursor"))
	if err != nil {
		writeTeachListError(w, r, err)
		return
	}
	items := make([]moderationCommentDTO, len(page.Comments))
	for i, comment := range page.Comments {
		items[i] = mapModerationComment(comment)
	}
	httpserver.WriteJSON(w, http.StatusOK, courseCommentsDTO{Comments: items, NextCursor: page.NextCursor})
}

func (handler *Handler) queue(w http.ResponseWriter, r *http.Request) {
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
	page, err := handler.service.Queue(r.Context(), actor, r.PathValue("courseId"), query.Get("cursor"))
	if err != nil {
		writeTeachListError(w, r, err)
		return
	}
	items := make([]queueItemDTO, len(page.Items))
	for i, item := range page.Items {
		items[i] = queueItemDTO{Kind: item.Kind, OpenReports: item.OpenReports}
		if item.Comment != nil {
			comment := mapModerationComment(*item.Comment)
			items[i].Comment = &comment
		}
		if item.Review != nil {
			review := mapReview(*item.Review)
			items[i].Review = &review
		}
	}
	httpserver.WriteJSON(w, http.StatusOK, queueDTO{Items: items, NextCursor: page.NextCursor})
}

func (handler *Handler) publicReviews(w http.ResponseWriter, r *http.Request) {
	noStore(w)
	query, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil || hasUnexpectedQuery(query, "cursor") || len(query["cursor"]) > 1 {
		writeError(w, r, http.StatusBadRequest, domain.ErrInvalidCursor)
		return
	}
	page, err := handler.service.PublicReviews(r.Context(), r.PathValue("slug"), query.Get("cursor"))
	if err != nil {
		writeReviewError(w, r, err, "course_not_found")
		return
	}
	items := make([]reviewDTO, len(page.Items))
	for i, review := range page.Items {
		items[i] = mapReview(review)
	}
	result := publicReviewsDTO{Items: items, NextCursor: nextCursorOrNull(page.NextCursor)}
	if page.Rating != nil {
		result.Rating = &reviewRatingDTO{Average: page.Rating.Average, Count: page.Rating.Count}
	}
	httpserver.WriteJSON(w, http.StatusOK, result)
}
func (handler *Handler) learningReviews(w http.ResponseWriter, r *http.Request) {
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
	page, err := handler.service.LearningReviews(r.Context(), actor, r.PathValue("courseId"), query.Get("cursor"))
	if err != nil {
		writeReviewError(w, r, err, "course_not_found")
		return
	}
	items := make([]reviewDTO, len(page.Items))
	for i, review := range page.Items {
		items[i] = mapReview(review)
	}
	result := publicReviewsDTO{Items: items, NextCursor: nextCursorOrNull(page.NextCursor)}
	if page.Rating != nil {
		result.Rating = &reviewRatingDTO{Average: page.Rating.Average, Count: page.Rating.Count}
	}
	httpserver.WriteJSON(w, http.StatusOK, result)
}
func (handler *Handler) ownReview(w http.ResponseWriter, r *http.Request) {
	noStore(w)
	actor, ok := actorFromRequest(w, r)
	if !ok {
		return
	}
	result, err := handler.service.OwnReview(r.Context(), actor, r.PathValue("courseId"))
	if err != nil {
		writeReviewError(w, r, err, "course_not_found")
		return
	}
	dto := ownReviewDTO{Eligibility: result.Eligibility}
	if result.Review != nil {
		review := mapReview(*result.Review)
		dto.Review = &review
	}
	httpserver.WriteJSON(w, http.StatusOK, dto)
}
func (handler *Handler) putReview(w http.ResponseWriter, r *http.Request) {
	noStore(w)
	actor, ok := actorFromRequest(w, r)
	if !ok {
		return
	}
	input, err := decodeJSON[putReviewRequest](w, r)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, err)
		return
	}
	review, created, err := handler.service.PutReview(r.Context(), actor, r.PathValue("courseId"), input.Stars, input.Body)
	if err != nil {
		writeReviewError(w, r, err, "course_not_found")
		return
	}
	status := http.StatusOK
	if created {
		status = http.StatusCreated
	}
	httpserver.WriteJSON(w, status, mapReview(review))
}
func (handler *Handler) deleteReview(w http.ResponseWriter, r *http.Request) {
	noStore(w)
	actor, ok := actorFromRequest(w, r)
	if !ok {
		return
	}
	if err := handler.service.DeleteReview(r.Context(), actor, r.PathValue("courseId")); err != nil {
		writeReviewError(w, r, err, "review_not_found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
func (handler *Handler) reportReview(w http.ResponseWriter, r *http.Request) {
	noStore(w)
	actor, ok := actorFromRequest(w, r)
	if !ok {
		return
	}
	if err := handler.service.ReportReview(r.Context(), actor, r.PathValue("reviewId")); err != nil {
		writeReviewError(w, r, err, "review_not_found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
func (handler *Handler) removeReview(w http.ResponseWriter, r *http.Request) {
	noStore(w)
	actor, ok := actorFromRequest(w, r)
	if !ok {
		return
	}
	input, err := decodeJSON[dismissReportsRequest](w, r)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, err)
		return
	}
	if err = handler.service.RemoveReview(r.Context(), actor, r.PathValue("reviewId"), input.Reason); err != nil {
		writeReviewError(w, r, err, "review_not_found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
func (handler *Handler) dismissReviewReports(w http.ResponseWriter, r *http.Request) {
	noStore(w)
	actor, ok := actorFromRequest(w, r)
	if !ok {
		return
	}
	input, err := decodeJSON[dismissReportsRequest](w, r)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, err)
		return
	}
	if err = handler.service.DismissReviewReports(r.Context(), actor, r.PathValue("reviewId"), input.Reason); err != nil {
		writeReviewError(w, r, err, "review_not_found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (handler *Handler) dismissReports(w http.ResponseWriter, r *http.Request) {
	noStore(w)
	actor, ok := actorFromRequest(w, r)
	if !ok {
		return
	}
	input, err := decodeJSON[dismissReportsRequest](w, r)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, err)
		return
	}
	if err := handler.service.DismissReports(r.Context(), actor, r.PathValue("commentId"), input.Reason); err != nil {
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
	return domain.Actor{UserID: transportActor.UserID, Administrator: transportActor.Role == "admin", EmailVerified: transportActor.EmailVerified}, true
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

func mapModerationComment(comment application.ModerationComment) moderationCommentDTO {
	return moderationCommentDTO{ID: comment.ID, LessonID: comment.LessonID, LessonTitle: comment.LessonTitle, ParentID: comment.ParentID,
		Author: authorDTO{DisplayName: comment.Author}, Body: comment.Body, PositionMs: comment.PositionMs,
		CreatedAt: comment.CreatedAt, EditedAt: comment.EditedAt, Removed: comment.Removed}
}
func mapReview(review application.Review) reviewDTO {
	return reviewDTO{ID: review.ID, AuthorDisplayName: review.AuthorDisplayName, Stars: review.Stars, Body: review.Body, CompletedLessons: review.CompletedLessons, TotalLessons: review.TotalLessons, CreatedAt: review.CreatedAt, EditedAt: review.EditedAt}
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

func writeReportError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, domain.ErrNotFound) {
		writeError(w, r, http.StatusNotFound, errors.New("comment_not_found"))
		return
	}
	writeError(w, r, statusFor(err), err)
}

func writeTeachListError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, domain.ErrNotFound) {
		writeError(w, r, http.StatusNotFound, errors.New("course_not_found"))
		return
	}
	writeError(w, r, statusFor(err), err)
}
func writeReviewError(w http.ResponseWriter, r *http.Request, err error, notFound string) {
	if errors.Is(err, domain.ErrNotFound) {
		writeError(w, r, http.StatusNotFound, errors.New(notFound))
		return
	}
	writeError(w, r, statusFor(err), err)
}

func statusFor(err error) int {
	switch {
	case errors.Is(err, domain.ErrForbidden), errors.Is(err, domain.ErrPurchaseRequired), errors.Is(err, domain.ErrEnrollmentRequired), errors.Is(err, domain.ErrEnrollmentBlocked), errors.Is(err, domain.ErrReportOwnComment), errors.Is(err, application.ErrCommentForbidden), errors.Is(err, application.ErrEmailNotVerified), errors.Is(err, domain.ErrReviewOwnerForbidden), errors.Is(err, domain.ErrReviewNotAvailable), errors.Is(err, domain.ErrReviewPurchaseRequired), errors.Is(err, domain.ErrReviewNotStarted), errors.Is(err, domain.ErrReviewModerationForbidden), errors.Is(err, domain.ErrReportOwnReview):
		return http.StatusForbidden
	case errors.Is(err, application.ErrCommentRemoved), errors.Is(err, domain.ErrThreadFull), errors.Is(err, domain.ErrReviewAlreadyExists), errors.Is(err, domain.ErrReviewRemoved):
		return http.StatusConflict
	case errors.Is(err, domain.ErrDependencyUnavailable):
		return http.StatusServiceUnavailable
	case errors.Is(err, domain.ErrBodyInvalid), errors.Is(err, domain.ErrPositionInvalid), errors.Is(err, domain.ErrParentInvalid), errors.Is(err, domain.ErrInvalidCursor), errors.Is(err, domain.ErrInvalidRequest), errors.Is(err, domain.ErrReviewInvalid):
		return http.StatusBadRequest
	default:
		return http.StatusBadRequest
	}
}

func writeError(w http.ResponseWriter, r *http.Request, status int, err error) {
	code := err.Error()
	message := "Request is not valid"
	switch code {
	case "course_purchase_required":
		message = "Purchase is required to access this course"
	case "course_forbidden":
		message = "Course access is forbidden"
	case "enrollment_required":
		message = "Enrollment is required to communicate in this course"
	case "enrollment_blocked":
		message = "Communication is blocked for this enrollment"
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

// nextCursorOrNull: the review lists say "no next page" with null (backlog B-28); the repository's keyset
// encoding uses "" for that.
func nextCursorOrNull(cursor string) *string {
	if cursor == "" {
		return nil
	}
	return &cursor
}
