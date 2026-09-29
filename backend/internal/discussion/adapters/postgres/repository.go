package postgres

import (
	"context"
	"encoding/base64"
	"errors"
	"strings"
	"time"

	"courses/internal/discussion/domain"
	"courses/internal/discussion/ports"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// This is the filtered current-video selector, shared in shape with transcript and playback;
// course uses its deliberately different sorting selector to report UI state. discussion must
// not copy that sorting form: a processing upload is not the playable asset.
const lessonFactsQuery = `SELECT courses.instructor_id::text, courses.status,
  COALESCE(current_asset.id::text, ''), COALESCE(current_asset.generation, 0), current_asset.duration_seconds
FROM course_lessons AS lessons
JOIN course_modules AS modules ON modules.id = lessons.module_id
JOIN courses ON courses.id = modules.course_id
LEFT JOIN LATERAL (
  SELECT assets.id, assets.generation, assets.duration_seconds
  FROM video_assets AS assets
  WHERE assets.lesson_id = lessons.id AND assets.state = 'ready' AND assets.evicted_at IS NULL
  ORDER BY assets.generation DESC
  LIMIT 1
) AS current_asset ON true
WHERE lessons.id = $1::text::uuid`

const rootsQuery = `SELECT c.id::text, c.lesson_id::text, c.author_id::text, c.parent_id::text,
  c.body, c.position_ms, c.created_at, c.edited_at, c.removed_at, c.removed_by::text, u.display_name,
  courses.id::text, courses.instructor_id::text, courses.status
FROM discussion_comments AS c
JOIN course_lessons AS lessons ON lessons.id = c.lesson_id
JOIN course_modules AS modules ON modules.id = lessons.module_id
JOIN courses ON courses.id = modules.course_id
LEFT JOIN users AS u ON u.id = c.author_id AND u.deleted_at IS NULL
WHERE c.lesson_id = $1::text::uuid AND c.parent_id IS NULL
  AND ($2::timestamptz IS NULL OR (c.created_at, c.id) < ($2, $3::text::uuid))
  AND (c.removed_at IS NULL OR EXISTS (
    SELECT 1 FROM discussion_comments AS replies
    WHERE replies.parent_id = c.id AND replies.removed_at IS NULL
  ))
ORDER BY c.created_at DESC, c.id DESC
LIMIT $4`

const repliesQuery = `SELECT c.id::text, c.lesson_id::text, c.author_id::text, c.parent_id::text,
  c.body, c.position_ms, c.created_at, c.edited_at, c.removed_at, c.removed_by::text, u.display_name,
  courses.id::text, courses.instructor_id::text, courses.status
FROM discussion_comments AS c
JOIN course_lessons AS lessons ON lessons.id = c.lesson_id
JOIN course_modules AS modules ON modules.id = lessons.module_id
JOIN courses ON courses.id = modules.course_id
LEFT JOIN users AS u ON u.id = c.author_id AND u.deleted_at IS NULL
WHERE c.parent_id = ANY($1::uuid[])
ORDER BY c.parent_id, c.created_at, c.id`

const commentQuery = `SELECT c.id::text, c.lesson_id::text, c.author_id::text, c.parent_id::text,
  c.body, c.position_ms, c.created_at, c.edited_at, c.removed_at, c.removed_by::text, u.display_name,
  courses.id::text, courses.instructor_id::text, courses.status
FROM discussion_comments AS c
JOIN course_lessons AS lessons ON lessons.id = c.lesson_id
JOIN course_modules AS modules ON modules.id = lessons.module_id
JOIN courses ON courses.id = modules.course_id
LEFT JOIN users AS u ON u.id = c.author_id AND u.deleted_at IS NULL
WHERE c.id = $1::text::uuid`

type Repository struct{ pool *pgxpool.Pool }

func NewRepository(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

func (repository *Repository) LessonFacts(ctx context.Context, lessonID string) (ports.LessonFacts, error) {
	var result ports.LessonFacts
	var status string
	var duration pgtype.Numeric
	err := repository.pool.QueryRow(ctx, lessonFactsQuery, lessonID).Scan(&result.Course.OwnerID, &status, &result.AssetID, &result.Generation, &duration)
	if errors.Is(err, pgx.ErrNoRows) {
		return ports.LessonFacts{}, domain.ErrNotFound
	}
	if err != nil {
		return ports.LessonFacts{}, persistenceError(err)
	}
	result.Course.Status = domain.Status(status)
	if value, err := duration.Float64Value(); err == nil && value.Valid {
		result.DurationSecs = &value.Float64
	}
	return result, nil
}

func (repository *Repository) List(ctx context.Context, lessonID, cursor string, limit int) (ports.CommentPage, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 50 {
		limit = 50
	}
	stamp, id, err := decodeCursor(cursor)
	if err != nil {
		return ports.CommentPage{}, err
	}
	rows, err := repository.pool.Query(ctx, rootsQuery, lessonID, stamp, id, limit+1)
	if err != nil {
		return ports.CommentPage{}, persistenceError(err)
	}
	defer rows.Close()
	roots := make([]ports.CommentRecord, 0, limit+1)
	for rows.Next() {
		record, err := scanRecord(rows)
		if err != nil {
			return ports.CommentPage{}, err
		}
		roots = append(roots, record)
	}
	if err := rows.Err(); err != nil {
		return ports.CommentPage{}, persistenceError(err)
	}
	page := ports.CommentPage{}
	if len(roots) > limit {
		last := roots[limit-1].Comment
		page.NextCursor = encodeCursor(last.CreatedAt, last.ID)
		roots = roots[:limit]
	}
	if len(roots) == 0 {
		return page, nil
	}
	ids := make([]string, len(roots))
	for i := range roots {
		ids[i] = roots[i].Comment.ID
	}
	replyRows, err := repository.pool.Query(ctx, repliesQuery, ids)
	if err != nil {
		return ports.CommentPage{}, persistenceError(err)
	}
	defer replyRows.Close()
	replies := make(map[string][]ports.CommentRecord)
	for replyRows.Next() {
		record, err := scanRecord(replyRows)
		if err != nil {
			return ports.CommentPage{}, err
		}
		if record.Comment.ParentID != nil {
			replies[*record.Comment.ParentID] = append(replies[*record.Comment.ParentID], record)
		}
	}
	if err := replyRows.Err(); err != nil {
		return ports.CommentPage{}, persistenceError(err)
	}
	page.Threads = make([]ports.Thread, len(roots))
	for i, root := range roots {
		page.Threads[i] = ports.Thread{Root: root, Replies: replies[root.Comment.ID]}
	}
	return page, nil
}

func (repository *Repository) Get(ctx context.Context, id string) (ports.CommentRecord, error) {
	return scanRecord(repository.pool.QueryRow(ctx, commentQuery, id))
}

func (repository *Repository) InsertRoot(ctx context.Context, comment domain.Comment) (ports.CommentRecord, error) {
	return repository.insert(ctx, comment)
}

func (repository *Repository) InsertReply(ctx context.Context, comment domain.Comment) (ports.CommentRecord, error) {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return ports.CommentRecord{}, domain.ErrDependencyUnavailable
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var parentLesson, parentID, body string
	var removedAt pgtype.Timestamptz
	err = tx.QueryRow(ctx, `SELECT lesson_id::text, id::text, body, removed_at FROM discussion_comments WHERE id=$1::text::uuid AND parent_id IS NULL FOR UPDATE`, *comment.ParentID).Scan(&parentLesson, &parentID, &body, &removedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return ports.CommentRecord{}, domain.ErrParentInvalid
	}
	if err != nil {
		return ports.CommentRecord{}, persistenceError(err)
	}
	if removedAt.Valid || parentLesson != comment.LessonID {
		return ports.CommentRecord{}, domain.ErrParentInvalid
	}
	var count int
	if err := tx.QueryRow(ctx, `SELECT count(*) FROM discussion_comments WHERE parent_id=$1::text::uuid AND removed_at IS NULL`, parentID).Scan(&count); err != nil {
		return ports.CommentRecord{}, persistenceError(err)
	}
	if err := domain.ValidateReplyCount(count); err != nil {
		return ports.CommentRecord{}, err
	}
	result, err := insertOn(tx, ctx, comment)
	if err != nil {
		return ports.CommentRecord{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return ports.CommentRecord{}, persistenceError(err)
	}
	return result, nil
}

func (repository *Repository) insert(ctx context.Context, comment domain.Comment) (ports.CommentRecord, error) {
	return insertOn(repository.pool, ctx, comment)
}

type queryer interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}

func insertOn(q queryer, ctx context.Context, comment domain.Comment) (ports.CommentRecord, error) {
	var id string
	var createdAt time.Time
	err := q.QueryRow(ctx, `INSERT INTO discussion_comments(lesson_id,author_id,parent_id,body,position_ms) VALUES($1::text::uuid,$2::text::uuid,$3::text::uuid,$4,$5) RETURNING id::text, created_at`, comment.LessonID, comment.AuthorID, comment.ParentID, comment.Body, comment.PositionMs).Scan(&id, &createdAt)
	if err != nil {
		return ports.CommentRecord{}, persistenceError(err)
	}
	comment.ID = id
	comment.CreatedAt = createdAt
	return ports.CommentRecord{Comment: comment}, nil
}

func (repository *Repository) UpdateBody(ctx context.Context, id, authorID, body string, now time.Time) (ports.CommentRecord, error) {
	var result ports.CommentRecord
	err := repository.pool.QueryRow(ctx, `UPDATE discussion_comments SET body=$3, edited_at=$4 WHERE id=$1::text::uuid AND author_id=$2::text::uuid AND removed_at IS NULL RETURNING id::text`, id, authorID, body, now).Scan(&result.Comment.ID)
	if errors.Is(err, pgx.ErrNoRows) {
		return ports.CommentRecord{}, domain.ErrNotFound
	}
	if err != nil {
		return ports.CommentRecord{}, persistenceError(err)
	}
	return repository.Get(ctx, id)
}

func (repository *Repository) Remove(ctx context.Context, id, actorID string, now time.Time) error {
	_, err := repository.pool.Exec(ctx, `UPDATE discussion_comments SET body=NULL, removed_at=COALESCE(removed_at,$2), removed_by=COALESCE(removed_by,$3::text::uuid) WHERE id=$1::text::uuid`, id, now, actorID)
	if err != nil {
		return persistenceError(err)
	}
	return nil
}

func scanRecord(row interface{ Scan(...any) error }) (ports.CommentRecord, error) {
	return scanRecordFields(row, recordScanTail{})
}

type recordScanTail struct {
	lessonTitle *string
	openReports *int
}

func scanModerationRecord(row interface{ Scan(...any) error }, openReports *int) (ports.ModerationCommentRecord, error) {
	var lessonTitle string
	record, err := scanRecordFields(row, recordScanTail{lessonTitle: &lessonTitle, openReports: openReports})
	if err != nil {
		return ports.ModerationCommentRecord{}, err
	}
	return ports.ModerationCommentRecord{CommentRecord: record, LessonTitle: lessonTitle}, nil
}

func scanRecordFields(row interface{ Scan(...any) error }, tail recordScanTail) (ports.CommentRecord, error) {
	var c ports.CommentRecord
	var status string
	var parentID, body pgtype.Text
	var position pgtype.Int8
	var edited, removed pgtype.Timestamptz
	var removedBy, display pgtype.Text
	values := []any{&c.Comment.ID, &c.Comment.LessonID, &c.Comment.AuthorID, &parentID, &body, &position, &c.Comment.CreatedAt, &edited, &removed, &removedBy, &display, &c.CourseID, &c.Course.OwnerID, &status}
	if tail.lessonTitle != nil {
		values = append(values, tail.lessonTitle)
	}
	if tail.openReports != nil {
		values = append(values, tail.openReports)
	}
	err := row.Scan(values...)
	if errors.Is(err, pgx.ErrNoRows) {
		return ports.CommentRecord{}, domain.ErrNotFound
	}
	if err != nil {
		return ports.CommentRecord{}, persistenceError(err)
	}
	c.Comment.ParentID = textPointer(parentID)
	if body.Valid {
		c.Comment.Body = body.String
	}
	if position.Valid {
		c.Comment.PositionMs = &position.Int64
	}
	if edited.Valid {
		c.Comment.EditedAt = &edited.Time
	}
	if removed.Valid {
		c.Comment.RemovedAt = &removed.Time
	}
	if removedBy.Valid {
		c.Comment.RemovedBy = &removedBy.String
	}
	if display.Valid {
		c.DisplayName = &display.String
	}
	c.Course.Status = domain.Status(status)
	return c, nil
}

func textPointer(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	return &value.String
}

func encodeCursor(at time.Time, id string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(at.UTC().Format(time.RFC3339Nano) + "|" + id))
}

func decodeCursor(value string) (*time.Time, *string, error) {
	if value == "" {
		return nil, nil, nil
	}
	decoded, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return nil, nil, domain.ErrInvalidCursor
	}
	parts := strings.Split(string(decoded), "|")
	if len(parts) != 2 {
		return nil, nil, domain.ErrInvalidCursor
	}
	at, err := time.Parse(time.RFC3339Nano, parts[0])
	if err != nil || parts[1] == "" {
		return nil, nil, domain.ErrInvalidCursor
	}
	return &at, &parts[1], nil
}

func persistenceError(err error) error {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	return domain.ErrDependencyUnavailable
}
