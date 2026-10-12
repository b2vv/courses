package ports

import (
	"context"
	"time"

	"courses/internal/course/domain"
)

type CourseRepository interface {
	ListPublished(context.Context, domain.Query) (domain.Page, error)
	GetPublished(context.Context, string) (domain.CourseDetail, error)
}

type LearningRepository interface {
	GetLearningCourse(context.Context, string, string) (domain.LearningCourse, error)
}

type AuthoringRepository interface {
	List(context.Context, domain.Actor, domain.Query, domain.Status) (domain.AuthoringPage, error)
	Load(context.Context, string) (domain.CourseTree, error)
	Create(context.Context, domain.Actor, domain.CourseDraft, time.Time) (domain.CourseTree, error)
	Save(context.Context, domain.Actor, domain.CourseTree, time.Time, time.Time) (domain.CourseTree, error)
	Transition(context.Context, domain.Actor, string, domain.Status, domain.Status, time.Time, time.Time) (domain.CourseTree, error)
	Delete(context.Context, domain.Actor, string, time.Time) error
}

type Grandfatherer interface {
	GrandfatherOnPriceSet(context.Context, string, time.Time, int64, string, time.Time) error
}

// LessonDocumentRepository keeps draft and published revisions separate.
// SaveDraft and PublishDocument return domain.ErrDocumentConflict when the
// expected updated_at does not match; an absent draft uses a zero time token.
type LessonDocumentRepository interface {
	SaveDraft(ctx context.Context, lessonID, locale string, body domain.LessonDocument, expectedUpdatedAt time.Time) (time.Time, error)
	// PublishDocument atomically copies draft_body to published_body under the lock.
	PublishDocument(ctx context.Context, lessonID, locale string, expectedUpdatedAt time.Time) error
	LoadDraft(ctx context.Context, lessonID, locale string) (body domain.LessonDocument, updatedAt time.Time, found bool, err error)
	LoadPublished(ctx context.Context, lessonID, locale string) (body domain.LessonDocument, found bool, err error)
}

// LessonDocumentAccessReader loads the full course and lesson access facts,
// including entitlement and membership for this user, and the lesson's preview.
type LessonDocumentAccessReader interface {
	LoadLessonAccess(ctx context.Context, userID, lessonID string) (domain.LessonDocumentAccess, error)
}
