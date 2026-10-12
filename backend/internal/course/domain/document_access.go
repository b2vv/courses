package domain

import "courses/internal/contracts/access"

const (
	ErrLessonNotFound   Error = "lesson_not_found"
	ErrEmailNotVerified Error = "email_not_verified"
	ErrDocumentConflict Error = "document_conflict"
)

type LessonDocumentAccess struct {
	OwnerID    string
	Status     Status
	Paid       bool
	Entitled   bool
	Membership access.Membership
	Preview    bool
}

func (document LessonDocumentAccess) CanBeEditedBy(actor Actor) bool {
	return actor.CanManage(document.OwnerID) == nil
}

func (document LessonDocumentAccess) CanConsumeBy(actor Actor) bool {
	return access.CanConsume(access.Subject{UserID: actor.UserID, Administrator: actor.Role == ActorAdmin}, document.courseAccess())
}

func (document LessonDocumentAccess) PurchaseRequiredBy(actor Actor) bool {
	return access.PurchaseRequired(access.Subject{UserID: actor.UserID, Administrator: actor.Role == ActorAdmin}, document.courseAccess())
}

func (document LessonDocumentAccess) courseAccess() access.Course {
	return access.Course{OwnerID: document.OwnerID, Published: document.Status == Published, Paid: document.Paid, Entitled: document.Entitled, Membership: document.Membership, Preview: document.Preview}
}
