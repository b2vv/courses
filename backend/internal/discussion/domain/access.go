package domain

import "courses/internal/contracts/access"

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

func (course Course) CanBeViewedBy(actor Actor) bool {
	return access.CanView(access.Subject{UserID: actor.UserID, Administrator: actor.Administrator}, access.Course{OwnerID: course.OwnerID, Published: course.Status == Published})
}

func (course Course) CanBeModeratedBy(actor Actor) bool {
	return actor.UserID != "" && (actor.Administrator || actor.UserID == course.OwnerID)
}

func IsModeration(actor Actor, comment Comment) bool {
	return actor.UserID != comment.AuthorID
}
