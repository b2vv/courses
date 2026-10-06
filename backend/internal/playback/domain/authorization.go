package domain

import "courses/internal/contracts/access"

type Actor struct {
	UserID        string
	Administrator bool
	EmailVerified bool
}

type Ownership struct {
	PlaybackID string
	VideoID    string
	LessonID   string
}

type CourseStatus string

const (
	Draft     CourseStatus = "draft"
	Published CourseStatus = "published"
	Archived  CourseStatus = "archived"
)

type VideoCourse struct {
	OwnerID  string
	Status   CourseStatus
	Paid     bool
	Entitled bool
	Preview  bool
}

type LessonPlayback struct {
	VideoID        string
	ClientSequence int64
	Course         VideoCourse
}

func (course VideoCourse) CanBeViewedBy(actor Actor) bool {
	return access.CanView(access.Subject{UserID: actor.UserID, Administrator: actor.Administrator}, access.Course{OwnerID: course.OwnerID, Published: course.Status == Published, Paid: course.Paid, Entitled: course.Entitled, Preview: course.Preview})
}

func (course VideoCourse) CanStreamBy(actor Actor) bool {
	return access.CanStream(access.Subject{UserID: actor.UserID, Administrator: actor.Administrator, EmailVerified: actor.EmailVerified}, access.Course{OwnerID: course.OwnerID, Published: course.Status == Published, Paid: course.Paid, Entitled: course.Entitled, Preview: course.Preview})
}

func (course VideoCourse) PurchaseRequiredBy(actor Actor) bool {
	return access.PurchaseRequired(access.Subject{UserID: actor.UserID, Administrator: actor.Administrator}, access.Course{OwnerID: course.OwnerID, Published: course.Status == Published, Paid: course.Paid, Entitled: course.Entitled, Preview: course.Preview})
}
