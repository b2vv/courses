package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"courses/internal/course/application"
	"courses/internal/course/domain"
)

type repository struct {
	query domain.Query
	list  func(context.Context, domain.Query) (domain.Page, error)
	get   func(context.Context, string) (domain.CourseDetail, error)
}

func (r *repository) ListPublished(ctx context.Context, query domain.Query) (domain.Page, error) {
	r.query = query
	return r.list(ctx, query)
}
func (r *repository) GetPublished(ctx context.Context, slug string) (domain.CourseDetail, error) {
	return r.get(ctx, slug)
}

func TestListPublishedCoursesNormalizesQueryAndRetainsPagination(t *testing.T) {
	repo := &repository{list: func(_ context.Context, query domain.Query) (domain.Page, error) {
		return domain.Page{Items: []domain.Course{}, Limit: query.Limit, Offset: query.Offset, Total: 3}, nil
	}}
	useCases := application.NewCatalog(repo)
	page, err := useCases.ListPublishedCourses(context.Background(), domain.Query{Limit: 1, Offset: 10, Search: " Go ", Sort: domain.SortTitle})
	if err != nil {
		t.Fatal(err)
	}
	if page.Total != 3 || page.Offset != 10 || page.Limit != 1 || len(page.Items) != 0 || repo.query.Search != "Go" {
		t.Fatalf("wrong page/query: %+v %+v", page, repo.query)
	}
	if _, err := useCases.ListPublishedCourses(context.Background(), domain.Query{Limit: 0}); !errors.Is(err, domain.ErrInvalidQuery) {
		t.Fatalf("got %v", err)
	}
}

func TestGetPublishedCourseValidatesSlugAndPropagatesExplicitNotFound(t *testing.T) {
	useCases := application.NewCatalog(&repository{get: func(context.Context, string) (domain.CourseDetail, error) {
		return domain.CourseDetail{}, domain.ErrCourseNotFound
	}})
	if _, err := useCases.GetPublishedCourse(context.Background(), "invalid--slug"); !errors.Is(err, domain.ErrInvalidSlug) {
		t.Fatalf("got %v", err)
	}
	if _, err := useCases.GetPublishedCourse(context.Background(), "missing-course"); !errors.Is(err, domain.ErrCourseNotFound) {
		t.Fatalf("got %v", err)
	}
}

func TestReadCancellationReachesRepository(t *testing.T) {
	repo := &repository{list: func(ctx context.Context, _ domain.Query) (domain.Page, error) {
		<-ctx.Done()
		return domain.Page{}, domain.ErrDependencyUnavailable
	}}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	_, err := application.NewCatalog(repo).ListPublishedCourses(ctx, domain.Query{Limit: 20, Sort: domain.SortRecent})
	if !errors.Is(err, domain.ErrDependencyUnavailable) {
		t.Fatalf("got %v", err)
	}
}

func TestSlowDependencyIsBoundedWithoutCallerCancellation(t *testing.T) {
	repo := &repository{get: func(ctx context.Context, _ string) (domain.CourseDetail, error) {
		<-ctx.Done()
		return domain.CourseDetail{}, domain.ErrDependencyUnavailable
	}}
	// Outer deadline only protects a regressed test from hanging indefinitely.
	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
	defer cancel()
	started := time.Now()
	_, err := application.NewCatalog(repo).GetPublishedCourse(ctx, "slow-course")
	if !errors.Is(err, domain.ErrDependencyUnavailable) {
		t.Fatalf("got %v", err)
	}
	if time.Since(started) > 4500*time.Millisecond {
		t.Fatal("catalog did not apply its bounded read budget")
	}
}
