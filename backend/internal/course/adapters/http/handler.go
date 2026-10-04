package http

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"courses/internal/course/application"
	"courses/internal/course/domain"
	"courses/platform/httpserver"
)

type Handler struct {
	catalog   *application.Catalog
	authoring *application.Authoring
}

func NewHandler(catalog *application.Catalog, authoring *application.Authoring) *Handler {
	return &Handler{catalog: catalog, authoring: authoring}
}
func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/courses", h.list)
	mux.HandleFunc("GET /api/v1/courses/{slug}", h.get)
}

type courseDTO struct {
	ID          string     `json:"id"`
	Slug        string     `json:"slug"`
	Title       string     `json:"title"`
	Summary     string     `json:"summary"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
	PublishedAt *time.Time `json:"publishedAt"`
	Price       *priceDTO  `json:"price"`
	Paid        bool       `json:"paid"`
	Rating      *ratingDTO `json:"rating"`
	LessonCount int        `json:"lessonCount"`
}
type ratingDTO struct {
	Average float64 `json:"average"`
	Count   int     `json:"count"`
}
type priceDTO struct {
	Minor    int64  `json:"minor"`
	Currency string `json:"currency"`
}
type programLessonDTO struct {
	Title string `json:"title"`
}
type programModuleDTO struct {
	Title   string             `json:"title"`
	Lessons []programLessonDTO `json:"lessons"`
}
type courseDetailDTO struct {
	courseDTO
	Program []programModuleDTO `json:"program"`
}
type paginationDTO struct {
	Limit  int   `json:"limit"`
	Offset int   `json:"offset"`
	Total  int64 `json:"total"`
}
type pageDTO struct {
	Items      []courseDTO   `json:"items"`
	Pagination paginationDTO `json:"pagination"`
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	query, err := parseQuery(r.URL.RawQuery)
	if err != nil {
		writeError(w, r, err)
		return
	}
	page, err := h.catalog.ListPublishedCourses(r.Context(), query)
	if err != nil {
		writeError(w, r, err)
		return
	}
	items := make([]courseDTO, 0, len(page.Items))
	for _, course := range page.Items {
		items = append(items, mapCourse(course))
	}
	httpserver.WriteJSON(w, http.StatusOK, pageDTO{Items: items, Pagination: paginationDTO{Limit: page.Limit, Offset: page.Offset, Total: page.Total}})
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	query, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil || len(query) != 0 {
		writeError(w, r, domain.ErrInvalidQuery)
		return
	}
	course, err := h.catalog.GetPublishedCourse(r.Context(), r.PathValue("slug"))
	if err != nil {
		writeError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, mapCourseDetail(course))
}

func parseQuery(raw string) (domain.Query, error) {
	values, err := url.ParseQuery(raw)
	if err != nil {
		return domain.Query{}, domain.ErrInvalidQuery
	}
	for key, value := range values {
		if len(value) != 1 {
			return domain.Query{}, domain.ErrInvalidQuery
		}
		switch key {
		case "limit", "offset", "q", "sort":
		default:
			return domain.Query{}, domain.ErrInvalidQuery
		}
	}
	limit, offset := 20, 0
	if value, exists := values["limit"]; exists {
		limit, err = decimal(value[0])
		if err != nil {
			return domain.Query{}, err
		}
	}
	if value, exists := values["offset"]; exists {
		offset, err = decimal(value[0])
		if err != nil {
			return domain.Query{}, err
		}
	}
	sort := "recent"
	if value, exists := values["sort"]; exists {
		sort = value[0]
	}
	return domain.Query{Limit: limit, Offset: offset, Search: values.Get("q"), Sort: domain.Sort(sort)}, nil
}

func decimal(value string) (int, error) {
	if value == "" {
		return 0, domain.ErrInvalidQuery
	}
	for _, char := range value {
		if char < '0' || char > '9' {
			return 0, domain.ErrInvalidQuery
		}
	}
	number, err := strconv.Atoi(value)
	if err != nil {
		return 0, domain.ErrInvalidQuery
	}
	return number, nil
}

func mapCourse(course domain.Course) courseDTO {
	var price *priceDTO
	if course.Price != nil {
		price = &priceDTO{Minor: course.Price.Minor, Currency: course.Price.Currency}
	}
	var summary *ratingDTO
	if course.Rating != nil {
		summary = &ratingDTO{Average: course.Rating.Average, Count: course.Rating.Count}
	}
	return courseDTO{ID: course.ID, Slug: course.Slug, Title: course.Title, Summary: course.Summary, Description: course.Description, Status: string(course.Status), CreatedAt: course.CreatedAt, UpdatedAt: course.UpdatedAt, PublishedAt: course.PublishedAt, Price: price, Paid: course.Paid, Rating: summary, LessonCount: course.LessonCount}
}

func mapCourseDetail(course domain.CourseDetail) courseDetailDTO {
	program := make([]programModuleDTO, 0, len(course.Program))
	for _, module := range course.Program {
		lessons := make([]programLessonDTO, 0, len(module.Lessons))
		for _, lesson := range module.Lessons {
			lessons = append(lessons, programLessonDTO{Title: lesson.Title})
		}
		program = append(program, programModuleDTO{Title: module.Title, Lessons: lessons})
	}
	return courseDetailDTO{courseDTO: mapCourse(course.Course), Program: program}
}

func writeError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidQuery):
		httpserver.WriteError(w, r, http.StatusBadRequest, "invalid_query", "Query parameters are invalid")
	case errors.Is(err, domain.ErrInvalidSlug):
		httpserver.WriteError(w, r, http.StatusBadRequest, "invalid_slug", "Course slug is invalid")
	case errors.Is(err, domain.ErrCourseNotFound):
		httpserver.WriteError(w, r, http.StatusNotFound, "course_not_found", "Course was not found")
	case errors.Is(err, domain.ErrDependencyUnavailable):
		httpserver.WriteError(w, r, http.StatusServiceUnavailable, "dependency_unavailable", "Course catalog is temporarily unavailable")
	default:
		httpserver.WriteError(w, r, http.StatusInternalServerError, "internal_error", "Unexpected server error")
	}
}
