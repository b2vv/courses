package ports

import (
	"context"
	"courses/internal/course/domain"
)

type CourseRepository interface {
	ListPublished(context.Context, domain.Query) (domain.Page, error)
	GetPublished(context.Context, string) (domain.Course, error)
}
