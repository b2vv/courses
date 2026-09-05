package application

import (
	"context"
	"courses/internal/course/domain"
	"courses/internal/course/ports"
	"time"
)

type Catalog struct{ repository ports.CourseRepository }

func NewCatalog(repository ports.CourseRepository) *Catalog { return &Catalog{repository: repository} }

func (c *Catalog) ListPublishedCourses(ctx context.Context, input domain.Query) (domain.Page, error) {
	query, err := domain.NewQuery(input.Limit, input.Offset, input.Search, string(input.Sort))
	if err != nil {
		return domain.Page{}, err
	}
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	return c.repository.ListPublished(ctx, query)
}

func (c *Catalog) GetPublishedCourse(ctx context.Context, slug string) (domain.CourseDetail, error) {
	if err := domain.ValidateSlug(slug); err != nil {
		return domain.CourseDetail{}, err
	}
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	return c.repository.GetPublished(ctx, slug)
}
