// Package access contains domain-safe access predicates shared across modules.
// Modules map their own facts into this contract; see ADR 013.
package access

// Subject contains the neutral facts about the viewer used by CanView.
type Subject struct {
	UserID        string
	Administrator bool
	EmailVerified bool
}

// Membership is a participant's state in a course.
type Membership string

const (
	MembershipNone    Membership = ""
	MembershipActive  Membership = "active"
	MembershipBlocked Membership = "blocked"
	MembershipRevoked Membership = "revoked"
)

// Course contains the neutral facts about the course used by CanView.
type Course struct {
	OwnerID    string
	Published  bool
	Paid       bool
	Entitled   bool
	Preview    bool
	Membership Membership
}

// CanView reports whether subject may view course.
func CanView(subject Subject, course Course) bool {
	return accessExempt(subject, course) || course.Published && (!course.Paid || course.Preview)
}

// CanStream reports whether subject may consume lesson content.
func CanStream(subject Subject, course Course) bool {
	return CanView(subject, course) && subject.EmailVerified
}

// CanConsume reports whether subject may consume lesson content. A preview lesson is open only
// while the course is published (mirroring CanView); membership (active or blocked) grants content
// regardless of published state, which is the archived-member case and matches the old
// entitlement-via-accessExempt path.
func CanConsume(subject Subject, course Course) bool {
	return subject.Administrator || (subject.UserID != "" && subject.UserID == course.OwnerID) || (course.Published && course.Preview) || course.Membership == MembershipActive || course.Membership == MembershipBlocked
}

// CanCommunicate reports whether subject may communicate in a course.
func CanCommunicate(subject Subject, course Course) bool {
	return subject.Administrator || (subject.UserID != "" && subject.UserID == course.OwnerID) || course.Membership == MembershipActive
}

// PurchaseRequired reports whether payment is the reason subject may not view course.
func PurchaseRequired(subject Subject, course Course) bool {
	return course.Published && course.Paid && !course.Preview && !accessExempt(subject, course)
}

func accessExempt(subject Subject, course Course) bool {
	return subject.Administrator || (subject.UserID != "" && subject.UserID == course.OwnerID) || course.Entitled
}
