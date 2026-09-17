package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"net/url"
	"time"

	"courses/internal/course/application"
	"courses/internal/course/domain"
	"courses/platform/httpserver"
)

const (
	maximumAdminWriteBody = 512 * 1024
	microsecondTimeLayout = "2006-01-02T15:04:05.000000Z07:00"
)

type adminCourseDTO struct {
	ID           string `json:"id"`
	Slug         string `json:"slug"`
	Title        string `json:"title"`
	Status       string `json:"status"`
	InstructorID string `json:"instructorId"`
	UpdatedAt    string `json:"updatedAt"`
}

type adminCoursePageDTO struct {
	Items      []adminCourseDTO `json:"items"`
	Pagination paginationDTO    `json:"pagination"`
}

type adminLessonDTO struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type adminModuleDTO struct {
	ID      string           `json:"id"`
	Title   string           `json:"title"`
	Lessons []adminLessonDTO `json:"lessons"`
}

type adminCourseTreeDTO struct {
	ID           string           `json:"id"`
	Slug         string           `json:"slug"`
	Title        string           `json:"title"`
	Summary      string           `json:"summary"`
	Description  string           `json:"description"`
	Status       string           `json:"status"`
	InstructorID string           `json:"instructorId"`
	CreatedAt    string           `json:"createdAt"`
	UpdatedAt    string           `json:"updatedAt"`
	PublishedAt  *string          `json:"publishedAt"`
	Modules      []adminModuleDTO `json:"modules"`
}

type adminCourseWriteRequest struct {
	Slug         *string `json:"slug"`
	Title        *string `json:"title"`
	Summary      *string `json:"summary"`
	Description  *string `json:"description"`
	InstructorID *string `json:"instructorId"`
}

type createAdminCourseRequest adminCourseWriteRequest

type saveAdminLessonRequest struct {
	ID          *string `json:"id"`
	Title       *string `json:"title"`
	Description *string `json:"description"`
}

type saveAdminModuleRequest struct {
	ID      *string                   `json:"id"`
	Title   *string                   `json:"title"`
	Lessons *[]saveAdminLessonRequest `json:"lessons"`
}

type saveAdminCourseRequest struct {
	ExpectedUpdatedAt *string                   `json:"expectedUpdatedAt"`
	Course            *adminCourseWriteRequest  `json:"course"`
	Modules           *[]saveAdminModuleRequest `json:"modules"`
}

type transitionAdminCourseRequest struct {
	Status            *string `json:"status"`
	ExpectedUpdatedAt *string `json:"expectedUpdatedAt"`
}

type deleteAdminCourseRequest struct {
	ExpectedUpdatedAt *string `json:"expectedUpdatedAt"`
}

func (h *Handler) RegisterAdmin(
	mux *http.ServeMux,
	authenticate func(http.Handler) http.Handler,
	requireSameOrigin func(http.Handler) http.Handler,
) {
	mux.Handle("GET /api/v1/admin/courses", authenticate(http.HandlerFunc(h.listAdminCourses)))
	mux.Handle("GET /api/v1/admin/courses/{id}", authenticate(http.HandlerFunc(h.getAdminCourse)))
	mux.Handle("POST /api/v1/admin/courses", requireSameOrigin(authenticate(http.HandlerFunc(h.createAdminCourse))))
	mux.Handle("PUT /api/v1/admin/courses/{id}", requireSameOrigin(authenticate(http.HandlerFunc(h.saveAdminCourse))))
	mux.Handle("POST /api/v1/admin/courses/{id}/status", requireSameOrigin(authenticate(http.HandlerFunc(h.transitionAdminCourse))))
	mux.Handle("DELETE /api/v1/admin/courses/{id}", requireSameOrigin(authenticate(http.HandlerFunc(h.deleteAdminCourse))))
}

func (h *Handler) listAdminCourses(w http.ResponseWriter, r *http.Request) {
	adminHeaders(w)
	actor, ok := adminActor(w, r)
	if !ok {
		return
	}
	input, err := parseAdminCourseQuery(r.URL.RawQuery)
	if err != nil {
		writeAdminError(w, r, err)
		return
	}
	page, err := h.authoring.List(r.Context(), actor, input)
	if err != nil {
		writeAdminError(w, r, err)
		return
	}
	items := make([]adminCourseDTO, 0, len(page.Items))
	for _, course := range page.Items {
		items = append(items, adminCourseDTO{
			ID: course.ID, Slug: course.Slug, Title: course.Title, Status: string(course.Status),
			InstructorID: course.InstructorID, UpdatedAt: formatAdminTime(course.UpdatedAt),
		})
	}
	httpserver.WriteJSON(w, http.StatusOK, adminCoursePageDTO{
		Items:      items,
		Pagination: paginationDTO{Limit: page.Limit, Offset: page.Offset, Total: page.Total},
	})
}

func (h *Handler) getAdminCourse(w http.ResponseWriter, r *http.Request) {
	adminHeaders(w)
	actor, ok := adminActor(w, r)
	if !ok {
		return
	}
	if query, err := url.ParseQuery(r.URL.RawQuery); err != nil || len(query) != 0 {
		writeAdminError(w, r, domain.ErrInvalidRequest)
		return
	}
	tree, err := h.authoring.Load(r.Context(), actor, r.PathValue("id"))
	if err != nil {
		writeAdminError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, mapAdminCourseTree(tree))
}

func (h *Handler) createAdminCourse(w http.ResponseWriter, r *http.Request) {
	adminHeaders(w)
	actor, ok := adminActor(w, r)
	if !ok {
		return
	}
	if !emptyAdminQuery(r.URL.RawQuery) {
		writeAdminError(w, r, domain.ErrInvalidRequest)
		return
	}
	request, err := decodeAdminJSON[createAdminCourseRequest](w, r)
	if err != nil {
		writeAdminError(w, r, err)
		return
	}
	draft, err := adminCourseWriteRequest(request).draft()
	if err != nil {
		writeAdminError(w, r, err)
		return
	}
	tree, err := h.authoring.Create(r.Context(), actor, draft)
	if err != nil {
		writeAdminError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusCreated, mapAdminCourseTree(tree))
}

func (h *Handler) saveAdminCourse(w http.ResponseWriter, r *http.Request) {
	adminHeaders(w)
	actor, ok := adminActor(w, r)
	if !ok {
		return
	}
	if !emptyAdminQuery(r.URL.RawQuery) {
		writeAdminError(w, r, domain.ErrInvalidRequest)
		return
	}
	request, err := decodeAdminJSON[saveAdminCourseRequest](w, r)
	if err != nil {
		writeAdminError(w, r, err)
		return
	}
	tree, expectedUpdatedAt, err := request.tree(r.PathValue("id"))
	if err != nil {
		writeAdminError(w, r, err)
		return
	}
	saved, err := h.authoring.Save(r.Context(), actor, tree, expectedUpdatedAt)
	if err != nil {
		writeAdminError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, mapAdminCourseTree(saved))
}

func (h *Handler) transitionAdminCourse(w http.ResponseWriter, r *http.Request) {
	adminHeaders(w)
	actor, ok := adminActor(w, r)
	if !ok {
		return
	}
	if !emptyAdminQuery(r.URL.RawQuery) {
		writeAdminError(w, r, domain.ErrInvalidRequest)
		return
	}
	request, err := decodeAdminJSON[transitionAdminCourseRequest](w, r)
	if err != nil || request.Status == nil || request.ExpectedUpdatedAt == nil {
		writeAdminError(w, r, domain.ErrInvalidRequest)
		return
	}
	expectedUpdatedAt, err := parseAdminTime(*request.ExpectedUpdatedAt)
	if err != nil {
		writeAdminError(w, r, domain.ErrInvalidRequest)
		return
	}
	tree, err := h.authoring.Transition(r.Context(), actor, r.PathValue("id"), domain.Status(*request.Status), expectedUpdatedAt)
	if err != nil {
		writeAdminError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, mapAdminCourseTree(tree))
}

func (h *Handler) deleteAdminCourse(w http.ResponseWriter, r *http.Request) {
	adminHeaders(w)
	actor, ok := adminActor(w, r)
	if !ok {
		return
	}
	if !emptyAdminQuery(r.URL.RawQuery) {
		writeAdminError(w, r, domain.ErrInvalidRequest)
		return
	}
	request, err := decodeAdminJSON[deleteAdminCourseRequest](w, r)
	if err != nil || request.ExpectedUpdatedAt == nil {
		writeAdminError(w, r, domain.ErrInvalidRequest)
		return
	}
	expectedUpdatedAt, err := parseAdminTime(*request.ExpectedUpdatedAt)
	if err != nil {
		writeAdminError(w, r, domain.ErrInvalidRequest)
		return
	}
	if err := h.authoring.Delete(r.Context(), actor, r.PathValue("id"), expectedUpdatedAt); err != nil {
		writeAdminError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func parseAdminCourseQuery(raw string) (application.AuthoringListInput, error) {
	values, err := url.ParseQuery(raw)
	if err != nil {
		return application.AuthoringListInput{}, domain.ErrInvalidRequest
	}
	for key, value := range values {
		if len(value) != 1 {
			return application.AuthoringListInput{}, domain.ErrInvalidRequest
		}
		switch key {
		case "limit", "offset", "status":
		default:
			return application.AuthoringListInput{}, domain.ErrInvalidRequest
		}
	}
	input := application.AuthoringListInput{Limit: 20}
	if value, exists := values["limit"]; exists {
		input.Limit, err = decimal(value[0])
		if err != nil {
			return application.AuthoringListInput{}, domain.ErrInvalidRequest
		}
	}
	if value, exists := values["offset"]; exists {
		input.Offset, err = decimal(value[0])
		if err != nil {
			return application.AuthoringListInput{}, domain.ErrInvalidRequest
		}
	}
	if value, exists := values["status"]; exists {
		input.Status = domain.Status(value[0])
		if err := domain.ValidateStatusFilter(input.Status); err != nil || input.Status == "" {
			return application.AuthoringListInput{}, domain.ErrInvalidRequest
		}
	}
	return input, nil
}

func adminActor(w http.ResponseWriter, r *http.Request) (domain.Actor, bool) {
	transportActor, ok := httpserver.ActorFromContext(r.Context())
	if !ok {
		httpserver.WriteError(w, r, http.StatusUnauthorized, "authentication_required", "Authentication is required")
		return domain.Actor{}, false
	}
	actor := domain.Actor{UserID: transportActor.UserID}
	switch transportActor.Role {
	case string(domain.ActorStudent):
		actor.Role = domain.ActorStudent
	case string(domain.ActorInstructor):
		actor.Role = domain.ActorInstructor
	case string(domain.ActorAdmin):
		actor.Role = domain.ActorAdmin
	default:
		writeAdminError(w, r, domain.ErrForbidden)
		return domain.Actor{}, false
	}
	if err := actor.CanCreate(); err != nil {
		writeAdminError(w, r, err)
		return domain.Actor{}, false
	}
	return actor, true
}

func mapAdminCourseTree(tree domain.CourseTree) adminCourseTreeDTO {
	modules := make([]adminModuleDTO, 0, len(tree.Modules))
	for _, module := range tree.Modules {
		lessons := make([]adminLessonDTO, 0, len(module.Lessons))
		for _, lesson := range module.Lessons {
			lessons = append(lessons, adminLessonDTO{ID: lesson.ID, Title: lesson.Title, Description: lesson.Description})
		}
		modules = append(modules, adminModuleDTO{ID: module.ID, Title: module.Title, Lessons: lessons})
	}
	return adminCourseTreeDTO{
		ID: tree.ID, Slug: tree.Slug, Title: tree.Title, Summary: tree.Summary,
		Description: tree.Description, Status: string(tree.Status), InstructorID: tree.InstructorID,
		CreatedAt: formatAdminTime(tree.CreatedAt), UpdatedAt: formatAdminTime(tree.UpdatedAt),
		PublishedAt: formatOptionalAdminTime(tree.PublishedAt), Modules: modules,
	}
}

func (request adminCourseWriteRequest) draft() (domain.CourseDraft, error) {
	if request.Slug == nil || request.Title == nil || request.Summary == nil || request.Description == nil {
		return domain.CourseDraft{}, domain.ErrInvalidRequest
	}
	instructorID := ""
	if request.InstructorID != nil {
		instructorID = *request.InstructorID
	}
	return domain.CourseDraft{
		Slug: *request.Slug, Title: *request.Title, Summary: *request.Summary,
		Description: *request.Description, InstructorID: instructorID,
	}, nil
}

func (request saveAdminCourseRequest) tree(id string) (domain.CourseTree, time.Time, error) {
	if request.ExpectedUpdatedAt == nil || request.Course == nil || request.Modules == nil {
		return domain.CourseTree{}, time.Time{}, domain.ErrInvalidRequest
	}
	expectedUpdatedAt, err := parseAdminTime(*request.ExpectedUpdatedAt)
	if err != nil {
		return domain.CourseTree{}, time.Time{}, domain.ErrInvalidRequest
	}
	draft, err := request.Course.draft()
	if err != nil {
		return domain.CourseTree{}, time.Time{}, err
	}
	modules := make([]domain.Module, 0, len(*request.Modules))
	for modulePosition, input := range *request.Modules {
		if input.ID == nil || input.Title == nil || input.Lessons == nil {
			return domain.CourseTree{}, time.Time{}, domain.ErrInvalidRequest
		}
		lessons := make([]domain.Lesson, 0, len(*input.Lessons))
		for lessonPosition, lesson := range *input.Lessons {
			if lesson.ID == nil || lesson.Title == nil || lesson.Description == nil {
				return domain.CourseTree{}, time.Time{}, domain.ErrInvalidRequest
			}
			lessons = append(lessons, domain.Lesson{
				ID: *lesson.ID, Title: *lesson.Title, Description: *lesson.Description, Position: lessonPosition,
			})
		}
		modules = append(modules, domain.Module{ID: *input.ID, Title: *input.Title, Position: modulePosition, Lessons: lessons})
	}
	return domain.CourseTree{
		ID: id, Slug: draft.Slug, Title: draft.Title, Summary: draft.Summary,
		Description: draft.Description, InstructorID: draft.InstructorID, Modules: modules,
	}, expectedUpdatedAt, nil
}

func decodeAdminJSON[T any](w http.ResponseWriter, r *http.Request) (T, error) {
	var result T
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		return result, domain.ErrInvalidRequest
	}
	r.Body = http.MaxBytesReader(w, r.Body, maximumAdminWriteBody)
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
	opening, err := decoder.Token()
	if err != nil || opening != json.Delim('{') {
		return domain.ErrInvalidRequest
	}
	if err := validateUniqueObject(decoder); err != nil {
		return err
	}
	if decoder.Decode(&struct{}{}) != io.EOF {
		return domain.ErrInvalidRequest
	}
	return nil
}

func validateUniqueObject(decoder *json.Decoder) error {
	seen := make(map[string]struct{})
	for decoder.More() {
		token, err := decoder.Token()
		key, ok := token.(string)
		if err != nil || !ok {
			return domain.ErrInvalidRequest
		}
		if _, duplicate := seen[key]; duplicate {
			return domain.ErrInvalidRequest
		}
		seen[key] = struct{}{}
		if err := validateUniqueValue(decoder); err != nil {
			return err
		}
	}
	closing, err := decoder.Token()
	if err != nil || closing != json.Delim('}') {
		return domain.ErrInvalidRequest
	}
	return nil
}

func validateUniqueValue(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return domain.ErrInvalidRequest
	}
	delimiter, ok := token.(json.Delim)
	if !ok {
		return nil
	}
	switch delimiter {
	case '{':
		return validateUniqueObject(decoder)
	case '[':
		for decoder.More() {
			if err := validateUniqueValue(decoder); err != nil {
				return err
			}
		}
		closing, err := decoder.Token()
		if err != nil || closing != json.Delim(']') {
			return domain.ErrInvalidRequest
		}
		return nil
	default:
		return domain.ErrInvalidRequest
	}
}

func emptyAdminQuery(raw string) bool {
	query, err := url.ParseQuery(raw)
	return err == nil && len(query) == 0
}

func formatAdminTime(value time.Time) string {
	return value.UTC().Truncate(time.Microsecond).Format(microsecondTimeLayout)
}

func parseAdminTime(value string) (time.Time, error) {
	validZone := len(value) == 27 && value[26] == 'Z' ||
		len(value) == 32 && (value[26] == '+' || value[26] == '-') && value[29] == ':'
	if !validZone || value[19] != '.' {
		return time.Time{}, domain.ErrInvalidRequest
	}
	for _, digit := range value[20:26] {
		if digit < '0' || digit > '9' {
			return time.Time{}, domain.ErrInvalidRequest
		}
	}
	parsed, err := time.Parse(microsecondTimeLayout, value)
	if err != nil {
		return time.Time{}, domain.ErrInvalidRequest
	}
	return parsed, nil
}

func formatOptionalAdminTime(value *time.Time) *string {
	if value == nil {
		return nil
	}
	formatted := formatAdminTime(*value)
	return &formatted
}

func adminHeaders(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-store")
}

func writeAdminError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidRequest):
		httpserver.WriteError(w, r, http.StatusBadRequest, "invalid_request", "Request is invalid")
	case errors.Is(err, domain.ErrForbidden):
		httpserver.WriteError(w, r, http.StatusForbidden, "forbidden", "Permission is required")
	case errors.Is(err, domain.ErrCourseForbidden):
		httpserver.WriteError(w, r, http.StatusForbidden, "course_forbidden", "Course access is forbidden")
	case errors.Is(err, domain.ErrCourseNotFound):
		httpserver.WriteError(w, r, http.StatusNotFound, "course_not_found", "Course was not found")
	case errors.Is(err, domain.ErrInvalidOwner):
		httpserver.WriteError(w, r, http.StatusBadRequest, "invalid_owner", "Course owner is invalid")
	case errors.Is(err, domain.ErrSlugTaken):
		httpserver.WriteError(w, r, http.StatusConflict, "slug_taken", "Course slug is already in use")
	case errors.Is(err, domain.ErrSlugImmutable):
		httpserver.WriteError(w, r, http.StatusConflict, "slug_immutable", "Published course slug cannot be changed")
	case errors.Is(err, domain.ErrCourseModified):
		httpserver.WriteError(w, r, http.StatusConflict, "course_modified", "Course changed since it was loaded")
	case errors.Is(err, domain.ErrInvalidStatusTransition):
		httpserver.WriteError(w, r, http.StatusConflict, "invalid_status_transition", "Course status transition is invalid")
	case errors.Is(err, domain.ErrCourseNotDeletable):
		httpserver.WriteError(w, r, http.StatusConflict, "course_not_deletable", "Only draft courses can be deleted")
	case errors.Is(err, domain.ErrLessonLimit):
		httpserver.WriteError(w, r, http.StatusBadRequest, "lesson_limit", "A course accepts at most 120 lessons across all of its modules")
	case errors.Is(err, domain.ErrDependencyUnavailable):
		httpserver.WriteError(w, r, http.StatusServiceUnavailable, "dependency_unavailable", "Course authoring is temporarily unavailable")
	default:
		httpserver.WriteError(w, r, http.StatusInternalServerError, "internal_error", "Unexpected server error")
	}
}
