package postgres

import (
	"context"
	"time"

	"courses/internal/course/domain"
	"github.com/jackc/pgx/v5/pgtype"
)

const getLearningCourse = `SELECT
  courses.id::text,
  courses.slug,
  courses.title,
  courses.summary,
  courses.description,
  courses.status,
  courses.created_at,
  courses.updated_at,
  courses.published_at,
  courses.instructor_id::text,
  COALESCE(modules.id::text, ''),
  COALESCE(modules.title, ''),
  COALESCE(lessons.id::text, ''),
  COALESCE(lessons.title, ''),
  COALESCE(
    CASE
      WHEN current_video.state IN ('created', 'uploading', 'queued', 'processing') THEN 'processing'
      ELSE current_video.state
    END,
    ''
  ),
  current_video.duration_seconds,
  COALESCE(progress.position_seconds, 0)::double precision,
  (progress.completed_at IS NOT NULL),
  COALESCE(progress.client_seq, 0)
FROM courses
LEFT JOIN course_modules AS modules ON modules.course_id = courses.id
LEFT JOIN course_lessons AS lessons ON lessons.module_id = modules.id
LEFT JOIN LATERAL (
  SELECT assets.state, assets.duration_seconds
  FROM video_assets AS assets
  WHERE assets.lesson_id = lessons.id AND assets.evicted_at IS NULL
  ORDER BY CASE WHEN assets.state = 'ready' THEN 0 ELSE 1 END, assets.generation DESC, assets.id
  LIMIT 1
) AS current_video ON true
LEFT JOIN lesson_progress AS progress
  ON progress.user_id = $1::text::uuid AND progress.lesson_id = lessons.id
WHERE courses.slug = $2
ORDER BY modules.position, modules.id, lessons.position, lessons.id`

type learningCourseRow struct {
	id, slug, title, summary, description, status, instructorID string
	createdAt, updatedAt, publishedAt                           pgtype.Timestamptz
}

func (repository *Repository) GetLearningCourse(ctx context.Context, userID, slug string) (domain.LearningCourse, error) {
	rows, err := repository.pool.Query(ctx, getLearningCourse, userID, slug)
	if err != nil {
		return domain.LearningCourse{}, persistenceError(err)
	}
	defer rows.Close()

	var result domain.LearningCourse
	moduleID := ""
	found := false
	for rows.Next() {
		var (
			courseRow                                                      learningCourseRow
			storedModuleID, moduleTitle, lessonID, lessonTitle, videoState string
			duration                                                       pgtype.Numeric
			position                                                       float64
			completed                                                      bool
			clientSequence                                                 int64
		)
		if err := rows.Scan(
			&courseRow.id, &courseRow.slug, &courseRow.title, &courseRow.summary, &courseRow.description, &courseRow.status,
			&courseRow.createdAt, &courseRow.updatedAt, &courseRow.publishedAt, &courseRow.instructorID,
			&storedModuleID, &moduleTitle, &lessonID, &lessonTitle, &videoState,
			&duration, &position, &completed, &clientSequence,
		); err != nil {
			return domain.LearningCourse{}, persistenceError(err)
		}
		if !found {
			course, err := mapLearningCourseRow(courseRow)
			if err != nil {
				return domain.LearningCourse{}, err
			}
			result = domain.LearningCourse{Course: course, InstructorID: courseRow.instructorID, Modules: []domain.LearningModule{}}
			found = true
		}
		if storedModuleID == "" {
			continue
		}
		if moduleID != storedModuleID {
			result.Modules = append(result.Modules, domain.LearningModule{ID: storedModuleID, Title: moduleTitle, Lessons: []domain.LearningLesson{}})
			moduleID = storedModuleID
		}
		if lessonID != "" {
			durationValue, err := duration.Float64Value()
			if err != nil {
				return domain.LearningCourse{}, persistenceError(err)
			}
			var durationSeconds *float64
			if durationValue.Valid {
				durationSeconds = &durationValue.Float64
			}
			lesson := domain.NewLearningLesson(domain.LearningLesson{
				ID: lessonID, Title: lessonTitle, VideoState: domain.LearningVideoState(videoState),
				Position: position, Completed: completed, ClientSequence: clientSequence,
			}, durationSeconds)
			result.Modules[len(result.Modules)-1].Lessons = append(result.Modules[len(result.Modules)-1].Lessons, lesson)
		}
	}
	if err := rows.Err(); err != nil {
		return domain.LearningCourse{}, persistenceError(err)
	}
	if !found {
		return domain.LearningCourse{}, domain.ErrCourseNotFound
	}
	return domain.NewLearningCourse(result)
}

func mapLearningCourseRow(row learningCourseRow) (domain.Course, error) {
	var publishedAt *time.Time
	if row.publishedAt.Valid {
		value := row.publishedAt.Time
		publishedAt = &value
	}
	return domain.NewCourse(domain.Course{
		ID: row.id, Slug: row.slug, Title: row.title, Summary: row.summary, Description: row.description,
		Status: domain.Status(row.status), CreatedAt: row.createdAt.Time, UpdatedAt: row.updatedAt.Time, PublishedAt: publishedAt,
	})
}
