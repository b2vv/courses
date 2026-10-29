package postgres

import (
	"context"
	"encoding/base64"
	"errors"
	"strings"
	"time"

	"courses/internal/contracts/access"
	"courses/internal/discussion/domain"
	"courses/internal/discussion/ports"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func (repository *Repository) ReportSubject(ctx context.Context, courseID, userID string) (string, access.Membership, bool, error) {
	var ownerID string
	var membership pgtype.Text
	err := repository.pool.QueryRow(ctx, `SELECT c.instructor_id::text, g.state
FROM courses AS c LEFT JOIN course_grants AS g ON g.course_id=c.id AND g.user_id=$2::text::uuid
WHERE c.id=$1::text::uuid`, courseID, userID).Scan(&ownerID, &membership)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", access.MembershipNone, false, nil
	}
	if err != nil {
		return "", access.MembershipNone, false, persistenceError(err)
	}
	if !membership.Valid {
		return ownerID, access.MembershipNone, true, nil
	}
	return ownerID, access.Membership(membership.String), true, nil
}

func (repository *Repository) InsertParticipantReport(ctx context.Context, report domain.ParticipantReport) (string, error) {
	var id string
	err := repository.pool.QueryRow(ctx, `INSERT INTO participant_reports(id,course_id,reporter_id,target_id,kind,reason,state,created_at)
VALUES(COALESCE(NULLIF($1::text,'')::uuid,gen_random_uuid()),$2::text::uuid,$3::text::uuid,$4::text::uuid,$5,$6,'open',$7)
ON CONFLICT (course_id,reporter_id,target_id) WHERE state='open' DO NOTHING
RETURNING id::text`, report.ID, report.CourseID, report.ReporterID, report.TargetID, report.Kind, report.Reason, report.CreatedAt).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", domain.ErrReportDuplicate
	}
	if err != nil {
		return "", persistenceError(err)
	}
	return id, nil
}

func (repository *Repository) ParticipantReportQueue(ctx context.Context, cursor, stateFilter string, limit int) (ports.ParticipantReportQueuePage, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	stamp, id, err := decodeParticipantReportCursor(cursor)
	if err != nil {
		return ports.ParticipantReportQueuePage{}, err
	}
	// An empty stateFilter selects every state; "open"/"resolved" narrows to one. The
	// comparison is parameterised, never interpolated, so an unexpected value simply
	// matches nothing rather than reaching the query text.
	rows, err := repository.pool.Query(ctx, `SELECT r.id::text,r.course_id::text,r.reporter_id::text,r.target_id::text,r.kind,r.reason,r.state,r.created_at,
 reporter.display_name,target.display_name,c.title
FROM participant_reports AS r JOIN courses AS c ON c.id=r.course_id
LEFT JOIN users AS reporter ON reporter.id=r.reporter_id AND reporter.deleted_at IS NULL
LEFT JOIN users AS target ON target.id=r.target_id AND target.deleted_at IS NULL
WHERE ($4::text='' OR r.state=$4) AND ($1::timestamptz IS NULL OR (r.created_at,r.id)<($1,$2::text::uuid))
ORDER BY r.created_at DESC,r.id DESC LIMIT $3`, stamp, id, limit+1, stateFilter)
	if err != nil {
		return ports.ParticipantReportQueuePage{}, persistenceError(err)
	}
	defer rows.Close()
	items := make([]ports.ParticipantReportQueueItem, 0, limit+1)
	for rows.Next() {
		item, scanErr := scanParticipantReport(rows)
		if scanErr != nil {
			return ports.ParticipantReportQueuePage{}, scanErr
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return ports.ParticipantReportQueuePage{}, persistenceError(err)
	}
	page := ports.ParticipantReportQueuePage{Items: items}
	if len(items) > limit {
		last := items[limit-1].Report
		page.NextCursor = encodeParticipantReportCursor(last.CreatedAt, last.ID)
		page.Items = items[:limit]
	}
	return page, nil
}

func (repository *Repository) ResolveParticipantReport(ctx context.Context, reportID string) error {
	_, err := repository.pool.Exec(ctx, `UPDATE participant_reports SET state='resolved' WHERE id=$1::text::uuid AND state='open'`, reportID)
	if err != nil {
		return persistenceError(err)
	}
	return nil
}

func scanParticipantReport(row interface{ Scan(...any) error }) (ports.ParticipantReportQueueItem, error) {
	var item ports.ParticipantReportQueueItem
	var reporter, target pgtype.Text
	err := row.Scan(&item.Report.ID, &item.Report.CourseID, &item.Report.ReporterID, &item.Report.TargetID, &item.Report.Kind, &item.Report.Reason, &item.Report.State, &item.Report.CreatedAt, &reporter, &target, &item.CourseTitle)
	if err != nil {
		return ports.ParticipantReportQueueItem{}, persistenceError(err)
	}
	if reporter.Valid {
		item.ReporterName = &reporter.String
	}
	if target.Valid {
		item.TargetName = &target.String
	}
	return item, nil
}

func decodeParticipantReportCursor(value string) (*time.Time, *string, error) {
	if value == "" {
		return nil, nil, nil
	}
	decoded, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return nil, nil, domain.ErrInvalidCursor
	}
	parts := strings.Split(string(decoded), "|")
	if len(parts) != 2 || !cursorUUIDValid(parts[1]) {
		return nil, nil, domain.ErrInvalidCursor
	}
	stamp, err := time.Parse(time.RFC3339Nano, parts[0])
	if err != nil {
		return nil, nil, domain.ErrInvalidCursor
	}
	return &stamp, &parts[1], nil
}
func encodeParticipantReportCursor(stamp time.Time, id string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(stamp.UTC().Format(time.RFC3339Nano) + "|" + id))
}
