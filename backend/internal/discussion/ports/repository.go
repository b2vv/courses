package ports

import (
	"context"
	"time"

	"courses/internal/contracts/access"
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

type ModerationCommentRecord struct {
	CommentRecord
	LessonTitle string
}

type Thread struct {
	Root    CommentRecord
	Replies []CommentRecord
}

type CommentPage struct {
	Threads    []Thread
	Comments   []ModerationCommentRecord
	NextCursor string
}

type CourseFacts struct {
	Course   domain.Course
	CourseID string
}

type QueueItem struct {
	ModerationCommentRecord
	OpenReports int
}

type QueuePage struct {
	Items      []QueueItem
	NextCursor string
}

type ModerationLogEntry struct {
	CourseID  string
	CommentID string
	ReviewID  string
	ActorID   string
	ActorRole string
	Action    string
	Reason    *string
	CreatedAt time.Time
}

type ReviewRecord struct {
	ID               string
	CourseID         string
	AuthorID         string
	Stars            int
	Body             string
	CompletedLessons int
	TotalLessons     int
	CreatedAt        time.Time
	EditedAt         *time.Time
	RemovedAt        *time.Time
	RemovedBy        *string
	DisplayName      *string
}

type ReviewFactsRecord struct {
	Facts            domain.ReviewFacts
	CompletedLessons int
	TotalLessons     int
	Exists           bool
	Removed          bool
}

type ReviewPage struct {
	Items      []ReviewRecord
	NextCursor string
}

type ModerationQueueItem struct {
	Kind          string
	ID            string
	CourseID      string
	AuthorID      string
	DisplayName   *string
	Stars         *int
	Body          string
	CreatedAt     time.Time
	OpenReports   int
	LatestReport  time.Time
	LessonID      *string
	LessonTitle   *string
	AuthorBlocked bool
}

type ModerationQueuePage struct {
	Items      []ModerationQueueItem
	NextCursor string
}

type ParticipantReportQueueItem struct {
	Report       domain.ParticipantReport
	ReporterName *string
	TargetName   *string
	CourseTitle  string
}

type ParticipantReportQueuePage struct {
	Items      []ParticipantReportQueueItem
	NextCursor string
}

type ParticipantReportRepository interface {
	ReportSubject(context.Context, string, string) (ownerID string, membership access.Membership, courseFound bool, err error)
	InsertParticipantReport(context.Context, domain.ParticipantReport) (string, error)
	ParticipantReportQueue(ctx context.Context, cursor, stateFilter string, limit int) (ParticipantReportQueuePage, error)
	ResolveParticipantReport(context.Context, string) error
}

type ReviewRepository interface {
	PublishedReviewCourse(context.Context, string) (string, error)
	ReviewFacts(context.Context, string, string) (ReviewFactsRecord, error)
	GetReview(context.Context, string) (ReviewRecord, error)
	CreateReview(context.Context, ReviewRecord) (ReviewRecord, error)
	UpdateReview(context.Context, string, string, domain.ReviewContent, domain.ReviewProgress, time.Time) (ReviewRecord, error)
	DeleteReview(context.Context, string, string) error
	GetOwnReview(context.Context, string, string) (*ReviewRecord, error)
	ListReviews(context.Context, string, string, int) (ReviewPage, error)
	ReviewSummary(context.Context, string) (int, int, error)
	InsertReviewReport(context.Context, string, string, time.Time) error
	DismissReviewReports(context.Context, string, string, time.Time) error
	RemoveReview(context.Context, string, string, string, *string, time.Time) error
	ListModerationQueue(context.Context, string, string, int) (ModerationQueuePage, error)
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
	ReviewRepository
	ParticipantReportRepository
	LessonFacts(context.Context, string, string) (LessonFacts, error)
	List(context.Context, string, string, int) (CommentPage, error)
	Get(context.Context, string, string) (CommentRecord, error)
	InsertRoot(context.Context, domain.Comment) (CommentRecord, error)
	InsertReply(context.Context, domain.Comment) (CommentRecord, error)
	UpdateBody(context.Context, string, string, string, time.Time) (CommentRecord, error)
	Remove(context.Context, string, string, time.Time) error
}
