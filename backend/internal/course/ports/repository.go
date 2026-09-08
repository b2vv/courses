package ports

import (
	"context"
	"time"

	"courses/internal/course/domain"
)

type CourseRepository interface {
	ListPublished(context.Context, domain.Query) (domain.Page, error)
	GetPublished(context.Context, string) (domain.CourseDetail, error)
}

type LearningRepository interface {
	GetLearningCourse(context.Context, string, string) (domain.LearningCourse, error)
}

type AuthoringRepository interface {
	List(context.Context, domain.Actor, domain.Query, domain.Status) (domain.AuthoringPage, error)
	Load(context.Context, string) (domain.CourseTree, error)
	Create(context.Context, domain.Actor, domain.CourseDraft, time.Time) (domain.CourseTree, error)
	Save(context.Context, domain.Actor, domain.CourseTree, time.Time, time.Time) (domain.CourseTree, error)
	Transition(context.Context, domain.Actor, string, domain.Status, domain.Status, time.Time, time.Time) (domain.CourseTree, error)
	Delete(context.Context, domain.Actor, string, time.Time) error
}
