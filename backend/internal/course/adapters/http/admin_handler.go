package http

import (
	"errors"
	"net/http"
	"net/url"
	"time"

	"courses/internal/course/application"
	"courses/internal/course/domain"
	"courses/platform/httpserver"
)

type adminCourseDTO struct {
	ID           string    `json:"id"`
	Slug         string    `json:"slug"`
	Title        string    `json:"title"`
	Status       string    `json:"status"`
	InstructorID string    `json:"instructorId"`
	UpdatedAt    time.Time `json:"updatedAt"`
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
	CreatedAt    time.Time        `json:"createdAt"`
	UpdatedAt    time.Time        `json:"updatedAt"`
	PublishedAt  *time.Time       `json:"publishedAt"`
	Modules      []adminModuleDTO `json:"modules"`
}

func (h *Handler) RegisterAdmin(
	mux *http.ServeMux,
	authenticate func(http.Handler) http.Handler,
	_ func(http.Handler) http.Handler,
) {
	mux.Handle("GET /api/v1/admin/courses", authenticate(http.HandlerFunc(h.listAdminCourses)))
	mux.Handle("GET /api/v1/admin/courses/{id}", authenticate(http.HandlerFunc(h.getAdminCourse)))
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
			InstructorID: course.InstructorID, UpdatedAt: course.UpdatedAt,
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
		CreatedAt: tree.CreatedAt, UpdatedAt: tree.UpdatedAt, PublishedAt: tree.PublishedAt, Modules: modules,
	}
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
	case errors.Is(err, domain.ErrDependencyUnavailable):
		httpserver.WriteError(w, r, http.StatusServiceUnavailable, "dependency_unavailable", "Course authoring is temporarily unavailable")
	default:
		httpserver.WriteError(w, r, http.StatusInternalServerError, "internal_error", "Unexpected server error")
	}
}
