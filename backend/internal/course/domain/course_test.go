package domain_test

import (
	"errors"
	"testing"
	"time"

	"courses/internal/course/domain"
)

func TestCourseRejectsInvalidPublicationState(t *testing.T) {
	now := time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC)
	valid := domain.Course{ID: "10000000-0000-4000-8000-000000000001", Slug: "go-foundations", Title: "Go foundations", Summary: "Learn Go", Description: "Build a Go application.", Status: domain.Published, CreatedAt: now, UpdatedAt: now, PublishedAt: &now}
	if _, err := domain.NewCourse(valid); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name   string
		change func(*domain.Course)
	}{
		{"published without date", func(c *domain.Course) { c.PublishedAt = nil }},
		{"draft with publication date", func(c *domain.Course) { c.Status = domain.Draft }},
		{"unknown status", func(c *domain.Course) { c.Status = "in-progress" }},
		{"invalid ID", func(c *domain.Course) { c.ID = "not-an-id" }},
		{"invalid slug", func(c *domain.Course) { c.Slug = "Go--foundations" }},
		{"empty title", func(c *domain.Course) { c.Title = " " }},
		{"missing timestamp", func(c *domain.Course) { c.CreatedAt = time.Time{} }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			course := valid
			tc.change(&course)
			if _, err := domain.NewCourse(course); !errors.Is(err, domain.ErrInvalidCourse) {
				t.Fatalf("expected invalid course, got %v", err)
			}
		})
	}
}
