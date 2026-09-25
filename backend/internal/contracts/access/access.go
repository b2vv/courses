// Package access contains domain-safe access predicates shared across modules.
// Modules map their own facts into this contract; see ADR 013.
package access

// Subject contains the neutral facts about the viewer used by CanView.
type Subject struct {
	UserID        string
	Administrator bool
}

// Course contains the neutral facts about the course used by CanView.
type Course struct {
	OwnerID   string
	Published bool
}

// CanView reports whether subject may view course.
func CanView(subject Subject, course Course) bool {
	return subject.Administrator || subject.UserID == course.OwnerID || course.Published
}
