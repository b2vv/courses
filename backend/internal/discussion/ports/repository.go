package ports

import (
	"context"
	"time"

	"courses/internal/discussion/domain"
)

type LessonFacts struct {
	Course       domain.Course
	AssetID      string
	Generation   int32
	DurationSecs *float64
}

type CommentRecord struct {
	Comment     domain.Comment
	DisplayName *string
	Course      domain.Course
	CourseID    string
}

type Thread struct {
	Root    CommentRecord
	Replies []CommentRecord
}

type CommentPage struct {
	Threads    []Thread
	Comments   []CommentRecord
	NextCursor string
}

type CourseFacts struct {
	Course   domain.Course
	CourseID string
}

type QueueItem struct {
	CommentRecord
	OpenReports int
}

type QueuePage struct {
	Items      []QueueItem
	NextCursor string
}

type ModerationLogEntry struct {
	CourseID  string
	CommentID string
	ActorID   string
	ActorRole string
	Action    string
	Reason    *string
	CreatedAt time.Time
}

type ModerationRepository interface {
	CourseFacts(context.Context, string) (CourseFacts, error)
	ListByCourse(context.Context, string, string, int) (CommentPage, error)
	InsertReport(context.Context, domain.Report, time.Time) error
	DismissReports(context.Context, string, string, time.Time) error
	ListQueue(context.Context, string, string, int) (QueuePage, error)
	WriteModerationLog(context.Context, ModerationLogEntry) error
}

type Repository interface {
	ModerationRepository
	LessonFacts(context.Context, string) (LessonFacts, error)
	List(context.Context, string, string, int) (CommentPage, error)
	Get(context.Context, string) (CommentRecord, error)
	InsertRoot(context.Context, domain.Comment) (CommentRecord, error)
	InsertReply(context.Context, domain.Comment) (CommentRecord, error)
	UpdateBody(context.Context, string, string, string, time.Time) (CommentRecord, error)
	Remove(context.Context, string, string, time.Time) error
}
