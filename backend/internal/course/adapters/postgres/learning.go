package postgres

import (
	"context"
	"time"

	"courses/internal/contracts/access"
	"courses/internal/contracts/rating"
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
  courses.price_minor,
  courses.currency,
  grants.state,
  COALESCE(reviews.review_count, 0)::bigint,
  COALESCE(reviews.review_stars_sum, 0)::bigint,
  COALESCE(lessons.is_preview, false),
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
LEFT JOIN course_grants AS grants
  ON grants.user_id = $1::text::uuid AND grants.course_id = courses.id
LEFT JOIN LATERAL (
  SELECT count(*) AS review_count, COALESCE(sum(stars), 0)::bigint AS review_stars_sum
  FROM discussion_reviews
  WHERE course_id = courses.id AND removed_at IS NULL
) AS reviews ON true
LEFT JOIN course_modules AS modules ON modules.course_id = courses.id
LEFT JOIN course_lessons AS lessons ON lessons.module_id = modules.id
LEFT JOIN LATERAL (
  -- This is the sorting form, unlike transcript, playback and discussion, which need only the
  -- playable ready asset. It reports the best state for the learning UI during an upload.
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
	priceMinor                                                  pgtype.Int8
	currency                                                    pgtype.Text
	membership                                                  pgtype.Text
	preview                                                     bool
	reviewCount, reviewStarsSum                                 int64
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
			&courseRow.priceMinor, &courseRow.currency, &courseRow.membership, &courseRow.reviewCount, &courseRow.reviewStarsSum, &courseRow.preview,
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
			result = domain.LearningCourse{
				Course: course, InstructorID: courseRow.instructorID, Modules: []domain.LearningModule{},
				Paid: course.Paid, Entitled: courseRow.membership.Valid && courseRow.membership.String == "active", Membership: membership(courseRow.membership),
			}
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
				IsPreview: courseRow.preview,
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

func membership(state pgtype.Text) access.Membership {
	if !state.Valid {
		return access.MembershipNone
	}
	return access.Membership(state.String)
}

func mapLearningCourseRow(row learningCourseRow) (domain.Course, error) {
	var publishedAt *time.Time
	if row.publishedAt.Valid {
		value := row.publishedAt.Time
		publishedAt = &value
	}
	// The learning page asks a locked lesson's buyer for this amount, so it is read here
	// rather than only its presence.
	var price *domain.Price
	if row.priceMinor.Valid && row.currency.Valid {
		price = &domain.Price{Minor: row.priceMinor.Int64, Currency: row.currency.String}
	} else if row.priceMinor.Valid != row.currency.Valid {
		return domain.Course{}, domain.ErrDependencyUnavailable
	}
	if row.reviewCount < 0 || row.reviewStarsSum < 0 || row.reviewCount > int64(^uint(0)>>1) || row.reviewStarsSum > int64(^uint(0)>>1) {
		return domain.Course{}, domain.ErrDependencyUnavailable
	}
	var summary *domain.Rating
	if average, visible := rating.Summary(int(row.reviewCount), int(row.reviewStarsSum)); visible {
		summary = &domain.Rating{Average: average, Count: int(row.reviewCount)}
	}
	return domain.NewCourse(domain.Course{
		ID: row.id, Slug: row.slug, Title: row.title, Summary: row.summary, Description: row.description,
		Status: domain.Status(row.status), CreatedAt: row.createdAt.Time, UpdatedAt: row.updatedAt.Time, PublishedAt: publishedAt,
		Price: price, Paid: price != nil, Rating: summary,
	})
}
