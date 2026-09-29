package postgres

import (
	"context"
	"errors"

	"courses/internal/playback/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

const loadLessonPlayback = `SELECT courses.instructor_id::text, courses.status,
courses.price_minor IS NOT NULL,
grants.id IS NOT NULL,
lessons.is_preview,
(
    SELECT assets.id::text
    FROM video_assets AS assets
    WHERE assets.lesson_id=lessons.id
      AND assets.state='ready' AND assets.evicted_at IS NULL
    ORDER BY assets.generation DESC
    LIMIT 1
),
COALESCE(progress.client_seq,0)
FROM course_lessons AS lessons
JOIN course_modules AS modules ON modules.id=lessons.module_id
JOIN courses ON courses.id=modules.course_id
LEFT JOIN course_grants AS grants
  ON grants.user_id=$1::text::uuid AND grants.course_id=courses.id AND grants.state='active'
LEFT JOIN lesson_progress AS progress
  ON progress.user_id=$1::text::uuid AND progress.lesson_id=lessons.id
WHERE lessons.id=$2::text::uuid`

type LessonPlaybackReader struct{ database QueryRower }

func NewLessonPlaybackReader(database QueryRower) *LessonPlaybackReader {
	return &LessonPlaybackReader{database: database}
}

func (reader *LessonPlaybackReader) LoadLessonPlayback(ctx context.Context, userID, lessonID string) (domain.LessonPlayback, error) {
	var (
		ownerID, status         string
		paid, entitled, preview bool
		videoID                 pgtype.Text
		clientSequence          int64
	)
	err := reader.database.QueryRow(ctx, loadLessonPlayback, userID, lessonID).Scan(
		&ownerID,
		&status,
		&paid,
		&entitled,
		&preview,
		&videoID,
		&clientSequence,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.LessonPlayback{}, domain.ErrLessonNotFound
	}
	if err != nil {
		return domain.LessonPlayback{}, domain.ErrDependencyUnavailable
	}
	if !videoID.Valid {
		return domain.LessonPlayback{}, domain.ErrVideoNotReady
	}
	courseStatus := domain.CourseStatus(status)
	if ownerID == "" || videoID.String == "" || clientSequence < 0 || clientSequence > domain.MaximumClientSequence ||
		courseStatus != domain.Draft && courseStatus != domain.Published && courseStatus != domain.Archived {
		return domain.LessonPlayback{}, domain.ErrDependencyUnavailable
	}
	return domain.LessonPlayback{
		VideoID:        videoID.String,
		ClientSequence: clientSequence,
		Course:         domain.VideoCourse{OwnerID: ownerID, Status: courseStatus, Paid: paid, Entitled: entitled, Preview: preview},
	}, nil
}
