package domain

import (
	"strings"
	"time"
	"unicode/utf8"
)

const (
	ErrInvalidRequest          Error = "invalid_request"
	ErrForbidden               Error = "forbidden"
	ErrCourseForbidden         Error = "course_forbidden"
	ErrSlugTaken               Error = "slug_taken"
	ErrSlugImmutable           Error = "slug_immutable"
	ErrInvalidOwner            Error = "invalid_owner"
	ErrInvalidStatusTransition Error = "invalid_status_transition"
	ErrCourseNotDeletable      Error = "course_not_deletable"
	ErrCourseModified          Error = "course_modified"
	ErrLessonLimit             Error = "lesson_limit"
)

// The domain owns these ceilings; the contract derives from them and does not restate them as
// a second truth (backlog B-3). MaximumModules and MaximumLessonDescription map onto a single
// JSON Schema keyword each, so openapi.yaml can enforce them. MaximumLessons cannot: it counts
// lessons cumulatively across every module, and JSON Schema has no way to bound a sum over
// sibling arrays. A schema that declared it per module would read as a course ceiling of
// 40*120 and enforce nothing, so the per-module maxItems stays at MaximumLessons - one module
// may legitimately hold every lesson - the course-wide rule is documented in the schema, and a
// request that breaks it is refused by name through ErrLessonLimit rather than as a generic
// invalid request.
const (
	MaximumModules           = 40
	MaximumLessons           = 120
	MaximumLessonDescription = 1000
)

type ActorRole string

const (
	ActorStudent    ActorRole = "student"
	ActorInstructor ActorRole = "instructor"
	ActorAdmin      ActorRole = "admin"
)

type Actor struct {
	UserID string
	Role   ActorRole
}

type CourseDraft struct {
	Slug         string
	Title        string
	Summary      string
	Description  string
	InstructorID string
}

type Lesson struct {
	ID          string
	Title       string
	Description string
	Position    int
}

type Module struct {
	ID       string
	Title    string
	Position int
	Lessons  []Lesson
}

type CourseTree struct {
	ID           string
	Slug         string
	Title        string
	Summary      string
	Description  string
	Status       Status
	InstructorID string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	PublishedAt  *time.Time
	Modules      []Module
}

type AuthoringCourse struct {
	ID           string
	Slug         string
	Title        string
	Status       Status
	InstructorID string
	UpdatedAt    time.Time
}

type AuthoringPage struct {
	Items  []AuthoringCourse
	Limit  int
	Offset int
	Total  int64
}

func (role ActorRole) Valid() bool {
	return role == ActorStudent || role == ActorInstructor || role == ActorAdmin
}

func (actor Actor) CanCreate() error {
	if !validID(actor.UserID) || actor.Role != ActorInstructor && actor.Role != ActorAdmin {
		return ErrForbidden
	}
	return nil
}

func (actor Actor) CanManage(instructorID string) error {
	if err := actor.CanCreate(); err != nil {
		return err
	}
	if actor.Role == ActorInstructor && actor.UserID != instructorID {
		return ErrCourseForbidden
	}
	return nil
}

func (actor Actor) CanAssignOwner(instructorID string) error {
	if err := actor.CanCreate(); err != nil {
		return err
	}
	if actor.Role == ActorInstructor {
		if instructorID == "" || instructorID == actor.UserID {
			return nil
		}
		return ErrCourseForbidden
	}
	if !validID(instructorID) {
		return ErrInvalidOwner
	}
	return nil
}

func (actor Actor) CanTransition(instructorID string, from, to Status) error {
	if err := actor.CanManage(instructorID); err != nil {
		return err
	}
	return ValidateTransition(from, to)
}

func ValidateCourseDraft(draft CourseDraft) error {
	if ValidateSlug(draft.Slug) != nil || !validText(draft.Title, 200) || !validText(draft.Summary, 500) || !validText(draft.Description, 20000) {
		return ErrInvalidRequest
	}
	if !validID(draft.InstructorID) {
		return ErrInvalidOwner
	}
	return nil
}

func ValidateCourseTree(tree CourseTree) error {
	if err := ValidateCourseTreeEdit(tree); err != nil {
		return err
	}
	if _, err := NewCourse(Course{
		ID: tree.ID, Slug: tree.Slug, Title: tree.Title, Summary: tree.Summary,
		Description: tree.Description, Status: tree.Status, CreatedAt: tree.CreatedAt,
		UpdatedAt: tree.UpdatedAt, PublishedAt: tree.PublishedAt,
	}); err != nil {
		return ErrInvalidRequest
	}
	return nil
}

func ValidateCourseTreeEdit(tree CourseTree) error {
	if err := ValidateID(tree.ID); err != nil {
		return err
	}
	if err := ValidateCourseDraft(CourseDraft{
		Slug: tree.Slug, Title: tree.Title, Summary: tree.Summary,
		Description: tree.Description, InstructorID: tree.InstructorID,
	}); err != nil {
		return err
	}
	return validateModules(tree.Modules)
}

func validateModules(modules []Module) error {
	if len(modules) > MaximumModules {
		return ErrInvalidRequest
	}
	seenIDs := make(map[string]struct{}, len(modules))
	lessonCount := 0
	for modulePosition, module := range modules {
		if !validNode(module.ID, module.Title, module.Position, modulePosition, seenIDs) {
			return ErrInvalidRequest
		}
		lessonCount += len(module.Lessons)
		if lessonCount > MaximumLessons {
			return ErrLessonLimit
		}
		for lessonPosition, lesson := range module.Lessons {
			if !validNode(lesson.ID, lesson.Title, lesson.Position, lessonPosition, seenIDs) || !validOptionalText(lesson.Description, MaximumLessonDescription) {
				return ErrInvalidRequest
			}
		}
	}
	return nil
}

func ValidateID(id string) error {
	if !validID(id) {
		return ErrInvalidRequest
	}
	return nil
}

func ValidateStatusFilter(status Status) error {
	if status != "" && status != Draft && status != Published && status != Archived {
		return ErrInvalidRequest
	}
	return nil
}

func ValidateTransition(from, to Status) error {
	if from == Draft && to == Published || from == Published && to == Archived || from == Archived && to == Published {
		return nil
	}
	return ErrInvalidStatusTransition
}

func ValidateSlugChange(currentSlug string, publishedAt *time.Time, proposedSlug string) error {
	if currentSlug != proposedSlug && publishedAt != nil {
		return ErrSlugImmutable
	}
	return nil
}

func ValidateDelete(status Status) error {
	if status != Draft {
		return ErrCourseNotDeletable
	}
	return nil
}

func validNode(id, title string, position, expectedPosition int, seenIDs map[string]struct{}) bool {
	if !validID(id) || !validText(title, 200) || position != expectedPosition {
		return false
	}
	if _, duplicate := seenIDs[id]; duplicate {
		return false
	}
	seenIDs[id] = struct{}{}
	return true
}

func validOptionalText(value string, maximum int) bool {
	return utf8.ValidString(value) && !strings.ContainsRune(value, 0) && utf8.RuneCountInString(value) <= maximum
}
