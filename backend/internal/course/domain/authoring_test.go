package domain_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"courses/internal/course/domain"
)

const (
	adminID      = "40000000-0000-4000-8000-000000000003"
	instructorID = "40000000-0000-4000-8000-000000000002"
	otherOwnerID = "40000000-0000-4000-8000-000000000004"
)

func TestActorPolicy(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name string
		act  func() error
		want error
	}{
		{name: "admin creates", act: func() error { return (domain.Actor{UserID: adminID, Role: domain.ActorAdmin}).CanCreate() }},
		{name: "instructor creates", act: func() error { return (domain.Actor{UserID: instructorID, Role: domain.ActorInstructor}).CanCreate() }},
		{name: "student refused", act: func() error { return (domain.Actor{UserID: instructorID, Role: domain.ActorStudent}).CanCreate() }, want: domain.ErrForbidden},
		{name: "unknown role refused", act: func() error { return (domain.Actor{UserID: instructorID, Role: "editor"}).CanCreate() }, want: domain.ErrForbidden},
		{name: "malformed actor refused", act: func() error { return (domain.Actor{UserID: "bad", Role: domain.ActorAdmin}).CanCreate() }, want: domain.ErrForbidden},
		{name: "admin manages any", act: func() error { return (domain.Actor{UserID: adminID, Role: domain.ActorAdmin}).CanManage(otherOwnerID) }},
		{name: "instructor manages own", act: func() error {
			return (domain.Actor{UserID: instructorID, Role: domain.ActorInstructor}).CanManage(instructorID)
		}},
		{name: "instructor cannot manage foreign", act: func() error {
			return (domain.Actor{UserID: instructorID, Role: domain.ActorInstructor}).CanManage(otherOwnerID)
		}, want: domain.ErrCourseForbidden},
		{name: "admin requires valid owner", act: func() error { return (domain.Actor{UserID: adminID, Role: domain.ActorAdmin}).CanAssignOwner("bad") }, want: domain.ErrInvalidOwner},
		{name: "instructor may omit owner on create", act: func() error {
			return (domain.Actor{UserID: instructorID, Role: domain.ActorInstructor}).CanAssignOwner("")
		}},
		{name: "instructor may echo self", act: func() error {
			return (domain.Actor{UserID: instructorID, Role: domain.ActorInstructor}).CanAssignOwner(instructorID)
		}},
		{name: "instructor cannot reassign", act: func() error {
			return (domain.Actor{UserID: instructorID, Role: domain.ActorInstructor}).CanAssignOwner(otherOwnerID)
		}, want: domain.ErrCourseForbidden},
		{name: "instructor publishes own", act: func() error {
			return (domain.Actor{UserID: instructorID, Role: domain.ActorInstructor}).CanTransition(instructorID, domain.Draft, domain.Published)
		}},
		{name: "instructor archives own", act: func() error {
			return (domain.Actor{UserID: instructorID, Role: domain.ActorInstructor}).CanTransition(instructorID, domain.Published, domain.Archived)
		}},
		{name: "instructor restores own", act: func() error {
			return (domain.Actor{UserID: instructorID, Role: domain.ActorInstructor}).CanTransition(instructorID, domain.Archived, domain.Published)
		}},
		{name: "instructor cannot transition foreign", act: func() error {
			return (domain.Actor{UserID: instructorID, Role: domain.ActorInstructor}).CanTransition(otherOwnerID, domain.Draft, domain.Published)
		}, want: domain.ErrCourseForbidden},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if err := test.act(); !errors.Is(err, test.want) {
				t.Fatalf("got %v, want %v", err, test.want)
			}
		})
	}
}

func TestStatusPolicy(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		from domain.Status
		to   domain.Status
		want error
	}{
		{from: domain.Draft, to: domain.Published},
		{from: domain.Published, to: domain.Archived},
		{from: domain.Archived, to: domain.Published},
		{from: domain.Draft, to: domain.Archived, want: domain.ErrInvalidStatusTransition},
		{from: domain.Published, to: domain.Draft, want: domain.ErrInvalidStatusTransition},
		{from: domain.Archived, to: domain.Draft, want: domain.ErrInvalidStatusTransition},
		{from: domain.Draft, to: domain.Draft, want: domain.ErrInvalidStatusTransition},
	} {
		name := string(test.from) + "_to_" + string(test.to)
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if err := domain.ValidateTransition(test.from, test.to); !errors.Is(err, test.want) {
				t.Fatalf("got %v, want %v", err, test.want)
			}
		})
	}

	if err := domain.ValidateDelete(domain.Draft); err != nil {
		t.Fatalf("draft delete: %v", err)
	}
	for _, status := range []domain.Status{domain.Published, domain.Archived} {
		if err := domain.ValidateDelete(status); !errors.Is(err, domain.ErrCourseNotDeletable) {
			t.Fatalf("delete %s: %v", status, err)
		}
	}
}

func TestSlugChangesOnlyBeforeFirstPublication(t *testing.T) {
	t.Parallel()

	publishedAt := time.Date(2026, 9, 5, 12, 0, 0, 0, time.UTC)
	for _, test := range []struct {
		name        string
		publishedAt *time.Time
		proposed    string
		want        error
	}{
		{name: "draft rename", proposed: "renamed-course"},
		{name: "published unchanged", publishedAt: &publishedAt, proposed: "course"},
		{name: "published rename", publishedAt: &publishedAt, proposed: "renamed-course", want: domain.ErrSlugImmutable},
		{name: "archived rename", publishedAt: &publishedAt, proposed: "renamed-course", want: domain.ErrSlugImmutable},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if err := domain.ValidateSlugChange("course", test.publishedAt, test.proposed); !errors.Is(err, test.want) {
				t.Fatalf("got %v, want %v", err, test.want)
			}
		})
	}
}

func TestCourseDraftFieldBounds(t *testing.T) {
	t.Parallel()

	valid := domain.CourseDraft{
		Slug: "course", Title: strings.Repeat("т", 200), Summary: strings.Repeat("с", 500),
		Description: strings.Repeat("о", 20000), InstructorID: instructorID,
	}
	if err := domain.ValidateCourseDraft(valid); err != nil {
		t.Fatalf("bounds rejected: %v", err)
	}

	for _, test := range []struct {
		name   string
		change func(*domain.CourseDraft)
		want   error
	}{
		{name: "blank title", change: func(draft *domain.CourseDraft) { draft.Title = " " }, want: domain.ErrInvalidRequest},
		{name: "summary over 500", change: func(draft *domain.CourseDraft) { draft.Summary += "с" }, want: domain.ErrInvalidRequest},
		{name: "description over 20000", change: func(draft *domain.CourseDraft) { draft.Description += "о" }, want: domain.ErrInvalidRequest},
		{name: "invalid slug", change: func(draft *domain.CourseDraft) { draft.Slug = "Course" }, want: domain.ErrInvalidRequest},
		{name: "invalid owner", change: func(draft *domain.CourseDraft) { draft.InstructorID = "bad" }, want: domain.ErrInvalidOwner},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			draft := valid
			test.change(&draft)
			if err := domain.ValidateCourseDraft(draft); !errors.Is(err, test.want) {
				t.Fatalf("got %v, want %v", err, test.want)
			}
		})
	}
}

func TestCourseTreeBoundsAndIdentity(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name   string
		change func(*domain.CourseTree)
	}{
		{name: "41 modules", change: func(tree *domain.CourseTree) {
			for len(tree.Modules) < domain.MaximumModules+1 {
				position := len(tree.Modules)
				tree.Modules = append(tree.Modules, domain.Module{ID: nodeID(1000 + position), Title: "Module", Position: position})
			}
		}},
		{name: "121 lessons", change: func(tree *domain.CourseTree) {
			tree.Modules[0].Lessons = nil
			for position := range domain.MaximumLessons + 1 {
				tree.Modules[0].Lessons = append(tree.Modules[0].Lessons, domain.Lesson{ID: nodeID(2000 + position), Title: "Lesson", Position: position})
			}
		}},
		{name: "lesson description over 1000", change: func(tree *domain.CourseTree) {
			tree.Modules[0].Lessons[0].Description = strings.Repeat("ї", domain.MaximumLessonDescription+1)
		}},
		{name: "module title over 200", change: func(tree *domain.CourseTree) { tree.Modules[0].Title = strings.Repeat("m", 201) }},
		{name: "blank lesson title", change: func(tree *domain.CourseTree) { tree.Modules[0].Lessons[0].Title = " " }},
		{name: "gapped module position", change: func(tree *domain.CourseTree) { tree.Modules[0].Position = 1 }},
		{name: "gapped lesson position", change: func(tree *domain.CourseTree) { tree.Modules[0].Lessons[0].Position = 1 }},
		{name: "duplicate module ids", change: func(tree *domain.CourseTree) {
			tree.Modules = append(tree.Modules, domain.Module{ID: tree.Modules[0].ID, Title: "Second", Position: 1})
		}},
		{name: "duplicate lesson ids across modules", change: func(tree *domain.CourseTree) {
			tree.Modules = append(tree.Modules, domain.Module{ID: nodeID(3), Title: "Second", Position: 1, Lessons: []domain.Lesson{{ID: tree.Modules[0].Lessons[0].ID, Title: "Duplicate", Position: 0}}})
		}},
		{name: "module and lesson share id", change: func(tree *domain.CourseTree) { tree.Modules[0].Lessons[0].ID = tree.Modules[0].ID }},
		{name: "nil uuid", change: func(tree *domain.CourseTree) { tree.Modules[0].ID = "00000000-0000-0000-0000-000000000000" }},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			tree := validAuthoringTree()
			test.change(&tree)
			if err := domain.ValidateCourseTree(tree); !errors.Is(err, domain.ErrInvalidRequest) {
				t.Fatalf("got %v", err)
			}
		})
	}
}

func TestCourseTreeAcceptsBoundsAndEmptyLessonDescription(t *testing.T) {
	t.Parallel()

	tree := validAuthoringTree()
	tree.Modules = nil
	for modulePosition := range domain.MaximumModules {
		module := domain.Module{ID: nodeID(1000 + modulePosition), Title: strings.Repeat("м", 200), Position: modulePosition}
		for lessonPosition := range 3 {
			module.Lessons = append(module.Lessons, domain.Lesson{
				ID:          nodeID(2000 + modulePosition*3 + lessonPosition),
				Title:       strings.Repeat("у", 200),
				Description: strings.Repeat("ї", domain.MaximumLessonDescription),
				Position:    lessonPosition,
			})
		}
		tree.Modules = append(tree.Modules, module)
	}
	tree.Modules[0].Lessons[0].Description = ""
	if err := domain.ValidateCourseTree(tree); err != nil {
		t.Fatal(err)
	}
}

func validAuthoringTree() domain.CourseTree {
	now := time.Date(2026, 9, 5, 12, 0, 0, 0, time.UTC)
	return domain.CourseTree{
		ID:           "10000000-0000-4000-8000-000000000001",
		Slug:         "course",
		Title:        "Course",
		Summary:      "Summary",
		Description:  "Description",
		Status:       domain.Draft,
		InstructorID: instructorID,
		CreatedAt:    now,
		UpdatedAt:    now,
		Modules: []domain.Module{{
			ID:       nodeID(1),
			Title:    "Module",
			Position: 0,
			Lessons:  []domain.Lesson{{ID: nodeID(2), Title: "Lesson", Description: "", Position: 0}},
		}},
	}
}

func nodeID(number int) string {
	return fmt.Sprintf("20000000-0000-4000-8000-%012d", number)
}
