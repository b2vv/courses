package domain

type Actor struct {
	UserID        string
	Administrator bool
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
	OwnerID string
	Status  CourseStatus
}

type LessonPlayback struct {
	VideoID        string
	ClientSequence int64
	Course         VideoCourse
}

func (course VideoCourse) CanBeViewedBy(actor Actor) bool {
	return actor.Administrator || actor.UserID == course.OwnerID || course.Status == Published
}
