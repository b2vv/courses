package postgres

import (
	"context"
	"encoding/base64"
	"errors"
	"strconv"
	"strings"
	"time"

	"courses/internal/discussion/domain"
	"courses/internal/discussion/ports"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

const reviewFactsQuery = `SELECT courses.instructor_id::text, courses.status,
  courses.price_minor IS NOT NULL,
  g.state,
  EXISTS (SELECT 1 FROM course_modules m JOIN course_lessons l ON l.module_id=m.id JOIN lesson_progress p ON p.lesson_id=l.id WHERE m.course_id=courses.id AND p.user_id=$2::text::uuid),
  (SELECT count(*) FROM course_modules m JOIN course_lessons l ON l.module_id=m.id JOIN lesson_progress p ON p.lesson_id=l.id WHERE m.course_id=courses.id AND p.user_id=$2::text::uuid AND p.completed_at IS NOT NULL),
  (SELECT count(*) FROM course_modules m JOIN course_lessons l ON l.module_id=m.id WHERE m.course_id=courses.id),
  EXISTS (SELECT 1 FROM discussion_reviews r WHERE r.course_id=courses.id AND r.author_id=$2::text::uuid),
  COALESCE((SELECT r.removed_at IS NOT NULL FROM discussion_reviews r WHERE r.course_id=courses.id AND r.author_id=$2::text::uuid), false)
FROM courses
LEFT JOIN course_grants g ON g.course_id=courses.id AND g.user_id=$2::text::uuid
WHERE courses.id=$1::text::uuid`

const reviewSelect = `SELECT r.id::text, r.course_id::text, r.author_id::text, r.stars, r.body,
 r.completed_lessons, r.total_lessons, r.created_at, r.edited_at, r.removed_at, r.removed_by::text, u.display_name
FROM discussion_reviews r LEFT JOIN users u ON u.id=r.author_id AND u.deleted_at IS NULL`

func (repository *Repository) PublishedReviewCourse(ctx context.Context, slug string) (string, error) {
	var id string
	err := repository.pool.QueryRow(ctx, `SELECT id::text FROM courses WHERE slug=$1 AND status='published'`, slug).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", domain.ErrNotFound
	}
	if err != nil {
		return "", persistenceError(err)
	}
	return id, nil
}

func (repository *Repository) GetReview(ctx context.Context, reviewID string) (ports.ReviewRecord, error) {
	return repository.getReview(ctx, ` WHERE r.id=$1::text::uuid`, reviewID)
}

func (repository *Repository) ReviewFacts(ctx context.Context, courseID, userID string) (ports.ReviewFactsRecord, error) {
	var result ports.ReviewFactsRecord
	var status string
	var membership pgtype.Text
	err := repository.pool.QueryRow(ctx, reviewFactsQuery, courseID, userID).Scan(&result.Facts.OwnerID, &status, &result.Facts.Paid, &membership, &result.Facts.HasProgress, &result.CompletedLessons, &result.TotalLessons, &result.Exists, &result.Removed)
	if errors.Is(err, pgx.ErrNoRows) {
		return ports.ReviewFactsRecord{}, domain.ErrNotFound
	}
	if err != nil {
		return ports.ReviewFactsRecord{}, persistenceError(err)
	}
	result.Facts.Status = domain.Status(status)
	result.Facts.Entitled = membership.Valid && membership.String == "active"
	result.Facts.Membership = membershipState(membership)
	return result, nil
}

func (repository *Repository) CreateReview(ctx context.Context, review ports.ReviewRecord) (ports.ReviewRecord, error) {
	row := repository.pool.QueryRow(ctx, `INSERT INTO discussion_reviews(id,course_id,author_id,stars,body,completed_lessons,total_lessons,created_at)
VALUES(COALESCE(NULLIF($1,'')::uuid, gen_random_uuid()),$2::text::uuid,$3::text::uuid,$4,$5,$6,$7,$8)
ON CONFLICT(course_id,author_id) DO NOTHING RETURNING id::text, course_id::text, author_id::text, stars, body, completed_lessons, total_lessons, created_at, edited_at, removed_at, removed_by::text, NULL::text`, review.ID, review.CourseID, review.AuthorID, review.Stars, review.Body, review.CompletedLessons, review.TotalLessons, review.CreatedAt)
	created, err := scanReview(row)
	if err == nil {
		return created, nil
	}
	if !errors.Is(err, domain.ErrNotFound) {
		return ports.ReviewRecord{}, err
	}
	existing, err := repository.getReview(ctx, ` WHERE r.course_id=$1::text::uuid AND r.author_id=$2::text::uuid`, review.CourseID, review.AuthorID)
	if err != nil {
		return ports.ReviewRecord{}, err
	}
	if existing.RemovedAt != nil {
		return ports.ReviewRecord{}, domain.ErrReviewRemoved
	}
	return ports.ReviewRecord{}, domain.ErrReviewAlreadyExists
}

func (repository *Repository) UpdateReview(ctx context.Context, courseID, authorID string, content domain.ReviewContent, progress domain.ReviewProgress, editedAt time.Time) (ports.ReviewRecord, error) {
	record, err := scanReview(repository.pool.QueryRow(ctx, `UPDATE discussion_reviews SET stars=$3,body=$4,completed_lessons=$5,total_lessons=$6,edited_at=$7
WHERE course_id=$1::text::uuid AND author_id=$2::text::uuid AND removed_at IS NULL
RETURNING id::text,course_id::text,author_id::text,stars,body,completed_lessons,total_lessons,created_at,edited_at,removed_at,removed_by::text,NULL::text`, courseID, authorID, content.Stars, content.Body, progress.Completed, progress.Total, editedAt))
	if err == nil || !errors.Is(err, domain.ErrNotFound) {
		return record, err
	}
	return ports.ReviewRecord{}, repository.reviewAbsentError(ctx, courseID, authorID)
}

func (repository *Repository) DeleteReview(ctx context.Context, courseID, authorID string) error {
	command, err := repository.pool.Exec(ctx, `DELETE FROM discussion_reviews WHERE course_id=$1::text::uuid AND author_id=$2::text::uuid AND removed_at IS NULL`, courseID, authorID)
	if err != nil {
		return persistenceError(err)
	}
	if command.RowsAffected() != 0 {
		return nil
	}
	return repository.reviewAbsentError(ctx, courseID, authorID)
}

func (repository *Repository) GetOwnReview(ctx context.Context, courseID, authorID string) (*ports.ReviewRecord, error) {
	record, err := repository.getReview(ctx, ` WHERE r.course_id=$1::text::uuid AND r.author_id=$2::text::uuid`, courseID, authorID)
	if errors.Is(err, domain.ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func (repository *Repository) ListReviews(ctx context.Context, courseID, cursor string, limit int) (ports.ReviewPage, error) {
	limit = moderationLimit(limit)
	at, id, err := decodeCursor(cursor)
	if err != nil {
		return ports.ReviewPage{}, err
	}
	rows, err := repository.pool.Query(ctx, reviewSelect+` WHERE r.course_id=$1::text::uuid AND r.removed_at IS NULL AND ($2::timestamptz IS NULL OR (r.created_at,r.id)<($2,$3::text::uuid)) ORDER BY r.created_at DESC,r.id DESC LIMIT $4`, courseID, at, id, limit+1)
	if err != nil {
		return ports.ReviewPage{}, persistenceError(err)
	}
	defer rows.Close()
	items := make([]ports.ReviewRecord, 0, limit+1)
	for rows.Next() {
		item, err := scanReview(rows)
		if err != nil {
			return ports.ReviewPage{}, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return ports.ReviewPage{}, persistenceError(err)
	}
	page := ports.ReviewPage{Items: items}
	if len(items) > limit {
		last := items[limit-1]
		page.NextCursor = encodeCursor(last.CreatedAt, last.ID)
		page.Items = items[:limit]
	}
	return page, nil
}

func (repository *Repository) ReviewSummary(ctx context.Context, courseID string) (int, int, error) {
	var count, sum int
	err := repository.pool.QueryRow(ctx, `SELECT count(*),COALESCE(sum(stars),0) FROM discussion_reviews WHERE course_id=$1::text::uuid AND removed_at IS NULL`, courseID).Scan(&count, &sum)
	if err != nil {
		return 0, 0, persistenceError(err)
	}
	return count, sum, nil
}

func (repository *Repository) InsertReviewReport(ctx context.Context, reviewID, reporterID string, at time.Time) error {
	var author string
	err := repository.pool.QueryRow(ctx, `SELECT author_id::text FROM discussion_reviews WHERE id=$1::text::uuid`, reviewID).Scan(&author)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}
	if err != nil {
		return persistenceError(err)
	}
	if author == reporterID {
		return domain.ErrReportOwnReview
	}
	_, err = repository.pool.Exec(ctx, `INSERT INTO discussion_review_reports(review_id,reporter_id,created_at) VALUES($1::text::uuid,$2::text::uuid,$3) ON CONFLICT(review_id,reporter_id) DO NOTHING`, reviewID, reporterID, at)
	if err != nil {
		return persistenceError(err)
	}
	return nil
}
func (repository *Repository) DismissReviewReports(ctx context.Context, reviewID, actorID string, at time.Time) error {
	_, err := repository.pool.Exec(ctx, `UPDATE discussion_review_reports SET dismissed_at=$3,dismissed_by=$2::text::uuid WHERE review_id=$1::text::uuid AND dismissed_at IS NULL`, reviewID, actorID, at)
	if err != nil {
		return persistenceError(err)
	}
	return nil
}

func (repository *Repository) RemoveReview(ctx context.Context, reviewID, actorID, actorRole string, reason *string, at time.Time) error {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return domain.ErrDependencyUnavailable
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var courseID string
	err = tx.QueryRow(ctx, `UPDATE discussion_reviews SET removed_at=$2,removed_by=$3::text::uuid WHERE id=$1::text::uuid AND removed_at IS NULL RETURNING course_id::text`, reviewID, at, actorID).Scan(&courseID)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}
	if err != nil {
		return persistenceError(err)
	}
	_, err = tx.Exec(ctx, `INSERT INTO discussion_moderation_log(course_id,comment_id,review_id,actor_id,actor_role,action,reason,created_at) VALUES($1::text::uuid,NULL,$2::text::uuid,$3::text::uuid,$4,'remove',$5,$6)`, courseID, reviewID, actorID, actorRole, reason, at)
	if err != nil {
		return persistenceError(err)
	}
	if err = tx.Commit(ctx); err != nil {
		return persistenceError(err)
	}
	return nil
}

func (repository *Repository) reviewAbsentError(ctx context.Context, courseID, authorID string) error {
	record, err := repository.getReview(ctx, ` WHERE r.course_id=$1::text::uuid AND r.author_id=$2::text::uuid`, courseID, authorID)
	if errors.Is(err, domain.ErrNotFound) {
		return domain.ErrNotFound
	}
	if err != nil {
		return err
	}
	if record.RemovedAt != nil {
		return domain.ErrReviewRemoved
	}
	return domain.ErrNotFound
}
func (repository *Repository) getReview(ctx context.Context, where string, args ...any) (ports.ReviewRecord, error) {
	return scanReview(repository.pool.QueryRow(ctx, reviewSelect+where, args...))
}

func scanReview(row interface{ Scan(...any) error }) (ports.ReviewRecord, error) {
	var r ports.ReviewRecord
	var edited, removed pgtype.Timestamptz
	var removedBy, display pgtype.Text
	err := row.Scan(&r.ID, &r.CourseID, &r.AuthorID, &r.Stars, &r.Body, &r.CompletedLessons, &r.TotalLessons, &r.CreatedAt, &edited, &removed, &removedBy, &display)
	if errors.Is(err, pgx.ErrNoRows) {
		return ports.ReviewRecord{}, domain.ErrNotFound
	}
	if err != nil {
		return ports.ReviewRecord{}, persistenceError(err)
	}
	if edited.Valid {
		r.EditedAt = &edited.Time
	}
	if removed.Valid {
		r.RemovedAt = &removed.Time
	}
	if removedBy.Valid {
		r.RemovedBy = &removedBy.String
	}
	if display.Valid {
		r.DisplayName = &display.String
	}
	return r, nil
}

// queue cursor also carries kind so equally-ranked review and comment rows page deterministically.
type moderationQueueCursor struct {
	count    int
	stamp    time.Time
	kind, id string
}

func encodeModerationQueueCursor(count int, stamp time.Time, kind, id string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(strconv.Itoa(count) + "|" + stamp.UTC().Format(time.RFC3339Nano) + "|" + kind + "|" + id))
}
func decodeModerationQueueCursor(value string) (moderationQueueCursor, error) {
	b, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return moderationQueueCursor{}, domain.ErrInvalidCursor
	}
	p := strings.Split(string(b), "|")
	if len(p) != 4 || p[2] == "" || !cursorUUIDValid(p[3]) {
		return moderationQueueCursor{}, domain.ErrInvalidCursor
	}
	c, err := strconv.Atoi(p[0])
	if err != nil {
		return moderationQueueCursor{}, domain.ErrInvalidCursor
	}
	at, err := time.Parse(time.RFC3339Nano, p[1])
	if err != nil {
		return moderationQueueCursor{}, domain.ErrInvalidCursor
	}
	return moderationQueueCursor{c, at, p[2], p[3]}, nil
}

const moderationQueueQuery = `WITH queued AS (
 SELECT 'comment'::text kind,c.id,courses.id course_id,c.author_id,u.display_name,NULL::smallint stars,c.body,c.created_at,
   reports.open_count,reports.latest_report,c.lesson_id,lessons.title lesson_title
 FROM discussion_comments c JOIN course_lessons lessons ON lessons.id=c.lesson_id JOIN course_modules modules ON modules.id=lessons.module_id JOIN courses ON courses.id=modules.course_id
 JOIN (SELECT comment_id,count(*) open_count,max(created_at) latest_report FROM discussion_comment_reports WHERE dismissed_at IS NULL GROUP BY comment_id) reports ON reports.comment_id=c.id
 LEFT JOIN users u ON u.id=c.author_id AND u.deleted_at IS NULL WHERE courses.id=$1::text::uuid AND c.removed_at IS NULL
 UNION ALL
 SELECT 'review',r.id,r.course_id,r.author_id,u.display_name,r.stars,r.body,r.created_at,
   reports.open_count,reports.latest_report,NULL::uuid,NULL::text
 FROM discussion_reviews r JOIN (SELECT review_id,count(*) open_count,max(created_at) latest_report FROM discussion_review_reports WHERE dismissed_at IS NULL GROUP BY review_id) reports ON reports.review_id=r.id
 LEFT JOIN users u ON u.id=r.author_id AND u.deleted_at IS NULL WHERE r.course_id=$1::text::uuid AND r.removed_at IS NULL
 ) SELECT kind,id::text,course_id::text,author_id::text,display_name,stars,body,created_at,open_count,latest_report,lesson_id::text,lesson_title,
   EXISTS(SELECT 1 FROM course_grants g WHERE g.user_id=queued.author_id AND g.course_id=queued.course_id AND g.state='blocked') author_blocked
 FROM queued
 WHERE ($2::int IS NULL OR (open_count,latest_report,kind,id) < ($2::int,$3::timestamptz,$4::text,$5::text::uuid))
 ORDER BY open_count DESC,latest_report DESC,kind DESC,id DESC LIMIT $6`

func (repository *Repository) ListModerationQueue(ctx context.Context, courseID, cursor string, limit int) (ports.ModerationQueuePage, error) {
	limit = moderationLimit(limit)
	var count *int
	var stamp *time.Time
	var kind, id *string
	if cursor != "" {
		decoded, err := decodeModerationQueueCursor(cursor)
		if err != nil {
			return ports.ModerationQueuePage{}, err
		}
		count = &decoded.count
		stamp = &decoded.stamp
		kind = &decoded.kind
		id = &decoded.id
	}
	rows, err := repository.pool.Query(ctx, moderationQueueQuery, courseID, count, stamp, kind, id, limit+1)
	if err != nil {
		return ports.ModerationQueuePage{}, persistenceError(err)
	}
	defer rows.Close()
	items := make([]ports.ModerationQueueItem, 0, limit+1)
	for rows.Next() {
		var item ports.ModerationQueueItem
		var display, lessonID, lessonTitle pgtype.Text
		var stars pgtype.Int2
		if err := rows.Scan(&item.Kind, &item.ID, &item.CourseID, &item.AuthorID, &display, &stars, &item.Body, &item.CreatedAt, &item.OpenReports, &item.LatestReport, &lessonID, &lessonTitle, &item.AuthorBlocked); err != nil {
			return ports.ModerationQueuePage{}, persistenceError(err)
		}
		if display.Valid {
			item.DisplayName = &display.String
		}
		if stars.Valid {
			value := int(stars.Int16)
			item.Stars = &value
		}
		if lessonID.Valid {
			item.LessonID = &lessonID.String
		}
		if lessonTitle.Valid {
			item.LessonTitle = &lessonTitle.String
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return ports.ModerationQueuePage{}, persistenceError(err)
	}
	page := ports.ModerationQueuePage{Items: items}
	if len(items) > limit {
		last := items[limit-1]
		page.NextCursor = encodeModerationQueueCursor(last.OpenReports, last.LatestReport, last.Kind, last.ID)
		page.Items = items[:limit]
	}
	return page, nil
}
