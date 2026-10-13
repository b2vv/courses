package domain

import "courses/internal/contracts/access"

type Actor struct {
	UserID        string
	Administrator bool
	EmailVerified bool
}

func (actor Actor) CanEditLesson(ownerID string) error {
	if actor.Administrator || (actor.UserID != "" && actor.UserID == ownerID) {
		return nil
	}
	return ErrForbidden
}

// ImageAccess carries the per-viewer course facts used to gate serving an image.
type ImageAccess struct {
	OwnerID    string
	Published  bool
	Paid       bool
	Entitled   bool
	Membership access.Membership
	Preview    bool
}

func (image ImageAccess) CanConsumeBy(actor Actor) bool {
	return access.CanConsume(access.Subject{UserID: actor.UserID, Administrator: actor.Administrator}, image.courseAccess())
}

func (image ImageAccess) courseAccess() access.Course {
	return access.Course{OwnerID: image.OwnerID, Published: image.Published, Paid: image.Paid, Entitled: image.Entitled, Membership: image.Membership, Preview: image.Preview}
}
