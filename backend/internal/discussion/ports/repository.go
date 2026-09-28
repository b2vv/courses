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
}

type Thread struct {
	Root    CommentRecord
	Replies []CommentRecord
}

type CommentPage struct {
	Threads    []Thread
	NextCursor string
}

type Repository interface {
	LessonFacts(context.Context, string) (LessonFacts, error)
	List(context.Context, string, string, int) (CommentPage, error)
	Get(context.Context, string) (CommentRecord, error)
	InsertRoot(context.Context, domain.Comment) (CommentRecord, error)
	InsertReply(context.Context, domain.Comment) (CommentRecord, error)
	UpdateBody(context.Context, string, string, string, time.Time) (CommentRecord, error)
	Remove(context.Context, string, time.Time) error
}
