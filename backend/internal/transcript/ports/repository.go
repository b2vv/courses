package ports

import (
	"context"

	"courses/internal/transcript/domain"
)

// LessonTranscript is what one read returns: the text, the course facts the access rule needs,
// and the generation the player would pick. Staleness is not stored, so the repository hands
// back the two generations and the domain compares them.
type LessonTranscript struct {
	Found      bool
	Course     domain.Course
	Transcript domain.Transcript

	// ReadyGeneration is the generation of the asset the player would pick, and zero when the
	// lesson has no ready asset. Only a ready asset counts: a lesson whose first video is still
	// uploading, or whose transcode failed, must not read as stale.
	ReadyGeneration int32
}

// SaveTranscript is the whole transcript of one lesson in one language: the write replaces, it
// does not merge. A partial write would need a merge rule nobody asked for, and the author edits
// a document, not a set of rows.
type SaveTranscript struct {
	LessonID string
	Language domain.Language
	Segments []domain.Segment

	// ExpectedUpdatedAt is the optimistic-lock token, empty when the author believes no
	// transcript exists yet. Two tabs, or an author against an import, otherwise overwrite each
	// other silently — the same expectation course authoring already sets with course_modified.
	ExpectedUpdatedAt string
}

type TranscriptRepository interface {
	// GetLessonTranscript returns the transcript of one lesson. An absent transcript is not an
	// error: a lesson simply may not have one yet, and the caller renders that as a state.
	// Language may be empty, in which case the repository picks the only one, or the first by
	// normalised code when several exist.
	GetLessonTranscript(ctx context.Context, lessonID string, language domain.Language) (LessonTranscript, error)

	// LessonCourse returns the course facts the write rule needs without reading the transcript,
	// so authorization does not depend on whether any text exists yet.
	LessonCourse(ctx context.Context, lessonID string) (domain.Course, error)

	// CourseDocuments lists the library of one course, ordered deterministically.
	CourseDocuments(ctx context.Context, courseID string) ([]domain.DocumentLink, error)

	// AddCourseDocument refuses with ErrDocumentLimit when the library is full. The ceiling is
	// checked inside the same transaction as the insert, or two concurrent additions both see
	// room and both take it.
	AddCourseDocument(ctx context.Context, courseID string, link domain.DocumentLink) error

	// RemoveCourseDocument refuses with ErrDocumentReferenced while any segment still points at
	// it. Deleting a course still cascades: this rule governs the author's action, not the
	// cascade, or saving a course tree would start failing on a flow that already works.
	RemoveCourseDocument(ctx context.Context, courseID, documentID string) error

	// CourseOfID returns the course facts for authorisation of library operations.
	CourseOfID(ctx context.Context, courseID string) (domain.Course, error)

	// SaveLessonTranscript replaces the transcript in one transaction and stamps the generation
	// of the ready asset it was written against. It returns the new token.
	SaveLessonTranscript(ctx context.Context, save SaveTranscript) (string, error)
}
