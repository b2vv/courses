package domain

// Status and Actor are this module's own shapes, not imports: a module's domain may not import
// another module's domain (AGENTS.md), so the access rule is copied rather than called. The copy
// is structural, not sloppy — and because it is a copy, it is held to the others by the truth
// table in platform/architecture/shared_rules_test.go.
//
// The other copies of this same rule: course/domain/learning.go LearningCourse.CanBeViewedBy and
// playback/domain/authorization.go VideoCourse.CanBeViewedBy. Change one, change all three, or
// the pin fails and names the disagreement.
type Status string

const (
	Draft     Status = "draft"
	Published Status = "published"
	Archived  Status = "archived"
)

type Actor struct {
	UserID        string
	Administrator bool
}

type Course struct {
	OwnerID string
	Status  Status
}

// CanBeEditedBy answers who may write one. Deliberately NOT the same rule as reading: a published
// course is readable by everyone signed in, and if writing shared that rule every student could
// rewrite the lecture. Ownership-scoped mutations are the model Phase 4a established — admin or
// the course owner, and no super_admin above them.
func (course Course) CanBeEditedBy(actor Actor) bool {
	return actor.Administrator || (actor.UserID != "" && actor.UserID == course.OwnerID)
}

// CanBeViewedBy answers who may read a transcript, and it must answer exactly as the other two
// copies do: a transcript is not a second way into a course (A6).
func (course Course) CanBeViewedBy(actor Actor) bool {
	return actor.Administrator || actor.UserID == course.OwnerID || course.Status == Published
}
