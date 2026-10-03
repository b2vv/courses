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
)

const courseFactsQuery = `SELECT id::text, instructor_id::text, status
FROM courses WHERE id=$1::text::uuid`

const courseCommentsQuery = `SELECT c.id::text, c.lesson_id::text, c.author_id::text, c.parent_id::text,
  c.body, c.position_ms, c.created_at, c.edited_at, c.removed_at, c.removed_by::text, u.display_name,
  courses.id::text, courses.instructor_id::text, courses.status, lessons.title
FROM discussion_comments AS c
JOIN course_lessons AS lessons ON lessons.id = c.lesson_id
JOIN course_modules AS modules ON modules.id = lessons.module_id
JOIN courses ON courses.id = modules.course_id
LEFT JOIN users AS u ON u.id = c.author_id AND u.deleted_at IS NULL
WHERE courses.id = $1::text::uuid
  AND ($2::timestamptz IS NULL OR (c.created_at, c.id) < ($2, $3::text::uuid))
ORDER BY c.created_at DESC, c.id DESC
LIMIT $4`

const queueQuery = `SELECT c.id::text, c.lesson_id::text, c.author_id::text, c.parent_id::text,
  c.body, c.position_ms, c.created_at, c.edited_at, c.removed_at, c.removed_by::text, u.display_name,
  courses.id::text, courses.instructor_id::text, courses.status, lessons.title, open_reports.open_count
FROM discussion_comments AS c
JOIN course_lessons AS lessons ON lessons.id = c.lesson_id
JOIN course_modules AS modules ON modules.id = lessons.module_id
JOIN courses ON courses.id = modules.course_id
JOIN (
  SELECT comment_id, count(*) AS open_count, max(created_at) AS latest_report
  FROM discussion_comment_reports
  WHERE dismissed_at IS NULL
  GROUP BY comment_id
) AS open_reports ON open_reports.comment_id = c.id
LEFT JOIN users AS u ON u.id = c.author_id AND u.deleted_at IS NULL
WHERE courses.id = $1::text::uuid AND c.removed_at IS NULL
  AND ($2::int IS NULL OR (open_reports.open_count, open_reports.latest_report, c.id) < ($2::int, $3::timestamptz, $4::text::uuid))
ORDER BY open_reports.open_count DESC, open_reports.latest_report DESC, c.id DESC
LIMIT $5`

func (repository *Repository) CourseFacts(ctx context.Context, courseID string) (ports.CourseFacts, error) {
	var facts ports.CourseFacts
	var status string
	err := repository.pool.QueryRow(ctx, courseFactsQuery, courseID).Scan(&facts.CourseID, &facts.Course.OwnerID, &status)
	if errors.Is(err, pgx.ErrNoRows) {
		return ports.CourseFacts{}, domain.ErrNotFound
	}
	if err != nil {
		return ports.CourseFacts{}, persistenceError(err)
	}
	facts.Course.Status = domain.Status(status)
	return facts, nil
}

func (repository *Repository) ListByCourse(ctx context.Context, courseID, cursor string, limit int) (ports.CommentPage, error) {
	limit = moderationLimit(limit)
	stamp, id, err := decodeCursor(cursor)
	if err != nil {
		return ports.CommentPage{}, err
	}
	rows, err := repository.pool.Query(ctx, courseCommentsQuery, courseID, stamp, id, limit+1)
	if err != nil {
		return ports.CommentPage{}, persistenceError(err)
	}
	defer rows.Close()
	comments := make([]ports.ModerationCommentRecord, 0, limit+1)
	for rows.Next() {
		record, err := scanModerationRecord(rows, nil)
		if err != nil {
			return ports.CommentPage{}, err
		}
		comments = append(comments, record)
	}
	if err := rows.Err(); err != nil {
		return ports.CommentPage{}, persistenceError(err)
	}
	page := ports.CommentPage{}
	if len(comments) > limit {
		last := comments[limit-1].Comment
		page.NextCursor = encodeCursor(last.CreatedAt, last.ID)
		comments = comments[:limit]
	}
	page.Comments = comments
	return page, nil
}

func (repository *Repository) InsertReport(ctx context.Context, report domain.Report, now time.Time) error {
	_, err := repository.pool.Exec(ctx, `INSERT INTO discussion_comment_reports(id, comment_id, reporter_id, created_at)
VALUES(COALESCE(NULLIF($1::text, '')::uuid, gen_random_uuid()), $2::text::uuid, $3::text::uuid, $4)
ON CONFLICT (comment_id, reporter_id) DO NOTHING`, report.ID, report.CommentID, report.ReporterID, now)
	if err != nil {
		return persistenceError(err)
	}
	return nil
}

func (repository *Repository) DismissReports(ctx context.Context, commentID, dismissedBy string, now time.Time) error {
	_, err := repository.pool.Exec(ctx, `UPDATE discussion_comment_reports
SET dismissed_at=$3, dismissed_by=$2::text::uuid
WHERE comment_id=$1::text::uuid AND dismissed_at IS NULL`, commentID, dismissedBy, now)
	if err != nil {
		return persistenceError(err)
	}
	return nil
}

func (repository *Repository) WriteModerationLog(ctx context.Context, entry ports.ModerationLogEntry) error {
	_, err := repository.pool.Exec(ctx, `INSERT INTO discussion_moderation_log(course_id, comment_id, review_id, actor_id, actor_role, action, reason, created_at)
VALUES($1::text::uuid, NULLIF($2,'')::uuid, NULLIF($3,'')::uuid, $4::text::uuid, $5, $6, $7, $8)`, entry.CourseID, entry.CommentID, entry.ReviewID, entry.ActorID, entry.ActorRole, entry.Action, entry.Reason, entry.CreatedAt)
	if err != nil {
		return persistenceError(err)
	}
	return nil
}

func (repository *Repository) ListQueue(ctx context.Context, courseID, cursor string, limit int) (ports.QueuePage, error) {
	limit = moderationLimit(limit)
	var count *int
	var stamp *time.Time
	var id *string
	if cursor != "" {
		decoded, err := decodeQueueCursor(cursor)
		if err != nil {
			return ports.QueuePage{}, err
		}
		countValue := decoded.count
		count = &countValue
		stamp, id = decoded.stamp, decoded.id
	}
	rows, err := repository.pool.Query(ctx, queueQuery, courseID, count, stamp, id, limit+1)
	if err != nil {
		return ports.QueuePage{}, persistenceError(err)
	}
	defer rows.Close()
	items := make([]ports.QueueItem, 0, limit+1)
	for rows.Next() {
		var openReports int
		record, err := scanModerationRecord(rows, &openReports)
		if err != nil {
			return ports.QueuePage{}, err
		}
		items = append(items, ports.QueueItem{
			ModerationCommentRecord: record,
			OpenReports:             openReports,
		})
	}
	if err := rows.Err(); err != nil {
		return ports.QueuePage{}, persistenceError(err)
	}
	page := ports.QueuePage{}
	if len(items) > limit {
		last := items[limit-1]
		page.NextCursor = encodeQueueCursor(last.OpenReports, last.Comment.CreatedAt, last.Comment.ID)
		items = items[:limit]
	}
	page.Items = items
	return page, nil
}

type queueCursor struct {
	count int
	stamp *time.Time
	id    *string
}

func decodeQueueCursor(value string) (queueCursor, error) {
	decodedBytes, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return queueCursor{}, domain.ErrInvalidCursor
	}
	parts := strings.Split(string(decodedBytes), "|")
	if len(parts) != 3 || parts[2] == "" {
		return queueCursor{}, domain.ErrInvalidCursor
	}
	count, err := strconv.Atoi(parts[0])
	if err != nil {
		return queueCursor{}, domain.ErrInvalidCursor
	}
	stamp, err := time.Parse(time.RFC3339Nano, parts[1])
	if err != nil {
		return queueCursor{}, domain.ErrInvalidCursor
	}
	return queueCursor{count: count, stamp: &stamp, id: &parts[2]}, nil
}

func encodeQueueCursor(count int, stamp time.Time, id string) string {
	value := strconv.Itoa(count) + "|" + stamp.UTC().Format(time.RFC3339Nano) + "|" + id
	return base64.RawURLEncoding.EncodeToString([]byte(value))
}

func moderationLimit(limit int) int {
	if limit <= 0 || limit > 50 {
		return 50
	}
	return limit
}
