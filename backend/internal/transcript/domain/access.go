package domain

import "courses/internal/contracts/access"

// Status and Actor are this module's own shapes, not imports: a module's domain may not import
// another module's domain (AGENTS.md), so this module maps its facts into the shared contract.
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
	OwnerID  string
	Status   Status
	Paid     bool
	Entitled bool
	Preview  bool
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
	return access.CanView(access.Subject{UserID: actor.UserID, Administrator: actor.Administrator}, access.Course{OwnerID: course.OwnerID, Published: course.Status == Published, Paid: course.Paid, Entitled: course.Entitled, Preview: course.Preview})
}
