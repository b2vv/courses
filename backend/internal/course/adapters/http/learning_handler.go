package http

import (
	"errors"
	"net/http"
	"net/url"

	"courses/internal/course/application"
	"courses/internal/course/domain"
	"courses/platform/httpserver"
)

type LearningHandler struct{ learning *application.Learning }

func NewLearningHandler(learning *application.Learning) *LearningHandler {
	return &LearningHandler{learning: learning}
}

func (handler *LearningHandler) Register(mux *http.ServeMux, authenticate func(http.Handler) http.Handler) {
	mux.Handle("GET /api/v1/learn/courses/{slug}", authenticate(http.HandlerFunc(handler.getCourse)))
}

type learningLessonDTO struct {
	ID         string  `json:"id"`
	Title      string  `json:"title"`
	VideoState *string `json:"videoState"`
	Position   float64 `json:"position"`
	Completed  bool    `json:"completed"`
	ClientSeq  int64   `json:"clientSeq"`
	IsPreview  bool    `json:"isPreview"`
	Locked     bool    `json:"locked"`
}

type learningModuleDTO struct {
	ID      string              `json:"id"`
	Title   string              `json:"title"`
	Lessons []learningLessonDTO `json:"lessons"`
}

type learningCourseDTO struct {
	courseDTO
	Modules          []learningModuleDTO `json:"modules"`
	ContinueLessonID *string             `json:"continueLessonId"`
	CanPurchase      bool                `json:"canPurchase"`
}

func (handler *LearningHandler) getCourse(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	query, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil || len(query) != 0 {
		writeError(w, r, domain.ErrInvalidQuery)
		return
	}
	actor, ok := learningActor(w, r)
	if !ok {
		return
	}
	course, err := handler.learning.GetCourse(r.Context(), actor, r.PathValue("slug"))
	if err != nil {
		writeLearningError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, mapLearningCourse(course, actor))
}

func learningActor(w http.ResponseWriter, r *http.Request) (domain.Actor, bool) {
	transportActor, ok := httpserver.ActorFromContext(r.Context())
	if !ok {
		httpserver.WriteError(w, r, http.StatusUnauthorized, "authentication_required", "Authentication is required")
		return domain.Actor{}, false
	}
	actor := domain.Actor{UserID: transportActor.UserID, Role: domain.ActorRole(transportActor.Role), EmailVerified: transportActor.EmailVerified}
	if !actor.Role.Valid() {
		writeLearningError(w, r, domain.ErrForbidden)
		return domain.Actor{}, false
	}
	return actor, true
}

func mapLearningCourse(course domain.LearningCourse, actor domain.Actor) learningCourseDTO {
	base := mapCourse(course.Course)
	base.Paid = course.Paid
	modules := make([]learningModuleDTO, 0, len(course.Modules))
	for _, module := range course.Modules {
		lessons := make([]learningLessonDTO, 0, len(module.Lessons))
		for _, lesson := range module.Lessons {
			var videoState *string
			if lesson.VideoState != domain.VideoAbsent {
				value := string(lesson.VideoState)
				videoState = &value
			}
			lessons = append(lessons, learningLessonDTO{
				ID: lesson.ID, Title: lesson.Title, VideoState: videoState,
				Position: lesson.Position, Completed: lesson.Completed, ClientSeq: lesson.ClientSequence,
				IsPreview: lesson.IsPreview, Locked: lesson.Locked,
			})
		}
		modules = append(modules, learningModuleDTO{ID: module.ID, Title: module.Title, Lessons: lessons})
	}
	var continueLessonID *string
	if course.ContinueLessonID != "" {
		value := course.ContinueLessonID
		continueLessonID = &value
	}
	return learningCourseDTO{
		courseDTO: base, Modules: modules, ContinueLessonID: continueLessonID, CanPurchase: course.CanPurchaseBy(actor),
	}
}

func writeLearningError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, domain.ErrPurchaseRequired) {
		httpserver.WriteError(w, r, http.StatusForbidden, "course_purchase_required", "Purchase is required to access this course")
		return
	}
	if errors.Is(err, domain.ErrForbidden) || errors.Is(err, domain.ErrCourseForbidden) {
		httpserver.WriteError(w, r, http.StatusForbidden, "course_forbidden", "Course is not available to this account")
		return
	}
	writeError(w, r, err)
}
