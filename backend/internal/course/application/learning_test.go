package application_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"courses/internal/course/application"
	"courses/internal/course/domain"
)

const (
	learningViewerID = "40000000-0000-4000-8000-000000000001"
	learningOwnerID  = "40000000-0000-4000-8000-000000000002"
)

type learningRepository struct {
	calls  int
	userID string
	slug   string
	course domain.LearningCourse
	err    error
}

func (repository *learningRepository) GetLearningCourse(_ context.Context, userID, slug string) (domain.LearningCourse, error) {
	repository.calls++
	repository.userID = userID
	repository.slug = slug
	return repository.course, repository.err
}

func TestLearningCourseValidatesBeforeLookupAndAppliesVisibility(t *testing.T) {
	actor := domain.Actor{UserID: learningViewerID, Role: domain.ActorStudent}

	invalid := &learningRepository{}
	if _, err := application.NewLearning(invalid).GetCourse(context.Background(), actor, "Invalid"); !errors.Is(err, domain.ErrInvalidSlug) {
		t.Fatalf("invalid slug error=%v", err)
	}
	if invalid.calls != 0 {
		t.Fatalf("invalid slug reached repository %d times", invalid.calls)
	}

	deniedCourse := learningCourse(nil)
	deniedCourse.Status = domain.Archived
	denied := &learningRepository{course: deniedCourse}
	if _, err := application.NewLearning(denied).GetCourse(context.Background(), actor, "learning-course"); !errors.Is(err, domain.ErrCourseForbidden) {
		t.Fatalf("denied error=%v", err)
	}
	if denied.calls != 1 {
		t.Fatalf("denied lookup calls=%d", denied.calls)
	}
	if denied.userID != learningViewerID || denied.slug != "learning-course" {
		t.Fatalf("lookup user=%q slug=%q", denied.userID, denied.slug)
	}

	owner := domain.Actor{UserID: learningOwnerID, Role: domain.ActorInstructor}
	allowed := &learningRepository{course: deniedCourse}
	if _, err := application.NewLearning(allowed).GetCourse(context.Background(), owner, "learning-course"); err != nil {
		t.Fatalf("R34 owner denied: %v", err)
	}
	if allowed.userID != learningOwnerID {
		t.Fatalf("owner lookup user=%q", allowed.userID)
	}
}

// Owner decision 2026-10-01: a published paid course shows its whole outline to a viewer
// without access; preview lessons open, the rest are locked for that viewer.
func TestLearningCourseShowsTheOutlineAndLocksLessonsPerViewer(t *testing.T) {
	lessons := []domain.LearningLesson{
		{ID: "70000000-0000-4000-8000-000000000001", Title: "First"},
		{ID: "70000000-0000-4000-8000-000000000002", Title: "Preview", IsPreview: true},
		{ID: "70000000-0000-4000-8000-000000000003", Title: "Third"},
	}
	locks := func(course domain.LearningCourse) []bool {
		var result []bool
		for _, lesson := range course.Modules[0].Lessons {
			result = append(result, lesson.Locked)
		}
		return result
	}
	for _, test := range []struct {
		name     string
		actor    domain.Actor
		entitled bool
		want     []bool
	}{
		{name: "student without access", actor: domain.Actor{UserID: learningViewerID, Role: domain.ActorStudent}, want: []bool{true, false, true}},
		{name: "entitled student", actor: domain.Actor{UserID: learningViewerID, Role: domain.ActorStudent}, entitled: true, want: []bool{false, false, false}},
		{name: "owner", actor: domain.Actor{UserID: learningOwnerID, Role: domain.ActorInstructor}, want: []bool{false, false, false}},
		{name: "admin", actor: domain.Actor{UserID: learningViewerID, Role: domain.ActorAdmin}, want: []bool{false, false, false}},
	} {
		t.Run(test.name, func(t *testing.T) {
			paid := learningCourse(append([]domain.LearningLesson(nil), lessons...))
			paid.Paid = true
			paid.Entitled = test.entitled
			course, err := application.NewLearning(&learningRepository{course: paid}).GetCourse(context.Background(), test.actor, "learning-course")
			if err != nil {
				t.Fatalf("outline refused: %v", err)
			}
			if got := locks(course); fmt.Sprint(got) != fmt.Sprint(test.want) {
				t.Fatalf("locked=%v want=%v", got, test.want)
			}
		})
	}

	draft := learningCourse(append([]domain.LearningLesson(nil), lessons...))
	draft.Paid = true
	draft.Status = domain.Draft
	draft.PublishedAt = nil
	if _, err := application.NewLearning(&learningRepository{course: draft}).GetCourse(context.Background(), domain.Actor{UserID: learningViewerID, Role: domain.ActorStudent}, "learning-course"); !errors.Is(err, domain.ErrCourseForbidden) {
		t.Fatalf("draft error=%v, want course forbidden", err)
	}
}

func learningCourse(lessons []domain.LearningLesson) domain.LearningCourse {
	now := time.Date(2026, 9, 8, 12, 0, 0, 0, time.UTC)
	return domain.LearningCourse{
		Course: domain.Course{
			ID: "50000000-0000-4000-8000-000000000001", Slug: "learning-course",
			Title: "Learning course", Summary: "Learning summary", Description: "Learning description",
			Status: domain.Published, CreatedAt: now, UpdatedAt: now, PublishedAt: &now,
		},
		InstructorID: learningOwnerID,
		Modules: []domain.LearningModule{{
			ID: "60000000-0000-4000-8000-000000000001", Title: "Module", Lessons: lessons,
		}},
	}
}
