package postgres

import (
	"context"
	"errors"

	"courses/internal/playback/domain"
	"github.com/jackc/pgx/v5"
)

const loadVideoCourse = `SELECT courses.instructor_id::text, courses.status
FROM video_assets AS assets
JOIN course_lessons AS lessons ON lessons.id=assets.lesson_id
JOIN course_modules AS modules ON modules.id=lessons.module_id
JOIN courses ON courses.id=modules.course_id
WHERE assets.id=$1::text::uuid AND assets.evicted_at IS NULL`

type QueryRower interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}

type Counter interface{ Inc() }

type VideoAccessReader struct {
	database QueryRower
	counter  Counter
}

func NewVideoAccessReader(database QueryRower, counter Counter) *VideoAccessReader {
	return &VideoAccessReader{database: database, counter: counter}
}

func (reader *VideoAccessReader) LoadVideoCourse(ctx context.Context, videoID string) (domain.VideoCourse, error) {
	reader.counter.Inc()
	var ownerID, status string
	err := reader.database.QueryRow(ctx, loadVideoCourse, videoID).Scan(&ownerID, &status)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.VideoCourse{}, domain.ErrForbidden
	}
	if err != nil {
		return domain.VideoCourse{}, domain.ErrDependencyUnavailable
	}
	courseStatus := domain.CourseStatus(status)
	if ownerID == "" || courseStatus != domain.Draft && courseStatus != domain.Published && courseStatus != domain.Archived {
		return domain.VideoCourse{}, domain.ErrDependencyUnavailable
	}
	return domain.VideoCourse{OwnerID: ownerID, Status: courseStatus}, nil
}
