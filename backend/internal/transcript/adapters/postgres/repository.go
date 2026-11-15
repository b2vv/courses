package postgres

import (
	"context"
	"errors"
	"time"

	"courses/internal/contracts/access"
	"courses/internal/transcript/domain"
	"courses/internal/transcript/ports"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// The generation the player would pick. This is the FILTERED form of the rule, matching
// playback/adapters/postgres/playback.go and progress.go; discussion/adapters/postgres/repository.go
// uses the same filtered form, while course/adapters/postgres/learning.go
// carries a third copy that SORTS ready first instead, because it reports the video state rather
// than only the playable asset. Copying the sorting form here would make a transcript read stale
// from the moment an upload begins and stay stale after a failed transcode.
const readyGenerationForLesson = `SELECT COALESCE((
  SELECT assets.generation
  FROM video_assets AS assets
  WHERE assets.lesson_id = $1::text::uuid AND assets.state = 'ready' AND assets.evicted_at IS NULL
  ORDER BY assets.generation DESC
  LIMIT 1
), 0)`

// The course facts the access rule needs are read even when no transcript exists, so that a
// missing transcript and a forbidden course cannot be told apart by the caller.
const getLessonAccessCourse = `SELECT courses.instructor_id::text, courses.status,
  courses.price_minor IS NOT NULL,
  grants.state,
  lessons.is_preview
FROM course_lessons AS lessons
JOIN course_modules AS modules ON modules.id = lessons.module_id
JOIN courses ON courses.id = modules.course_id
LEFT JOIN course_grants AS grants
  ON grants.user_id = $1::text::uuid AND grants.course_id = courses.id
WHERE lessons.id = $2::text::uuid`

// An empty language selects the only transcript, or the lesson's primary one when several exist
// (B-21). is_primary is the first-written language unless an author later changes it, so adding a
// second language cannot silently displace the canonical read the way the old alphabetical order
// did. The trailing language keeps the choice deterministic if no primary is marked, rather than
// leaving it to whatever the planner returns first.
const getTranscript = `SELECT language, authored_against_generation, updated_at
FROM transcripts
WHERE lesson_id = $1::text::uuid AND ($2 = '' OR language = $2)
ORDER BY is_primary DESC, language
LIMIT 1`

const getSegments = `SELECT
  segments.id::text,
  segments.start_ms,
  segments.end_ms,
  segments.text,
  COALESCE(segments.speaker, ''),
  segments.confidence,
  COALESCE(documents.id::text, ''),
  COALESCE(documents.label, ''),
  COALESCE(documents.url, '')
FROM transcript_segments AS segments
LEFT JOIN segment_documents AS links ON links.segment_id = segments.id
LEFT JOIN course_documents AS documents ON documents.id = links.document_id
WHERE segments.lesson_id = $1::text::uuid AND segments.language = $2
ORDER BY segments.ordinal, segments.id, links.position`

type Repository struct{ pool *pgxpool.Pool }

func NewRepository(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

func (repository *Repository) GetLessonTranscript(ctx context.Context, userID, lessonID string, language domain.Language) (ports.LessonTranscript, error) {
	var ownerID, status string
	var paid, preview bool
	var membership pgtype.Text
	err := repository.pool.QueryRow(ctx, getLessonAccessCourse, userID, lessonID).Scan(&ownerID, &status, &paid, &membership, &preview)
	if errors.Is(err, pgx.ErrNoRows) {
		return ports.LessonTranscript{}, domain.ErrNotFound
	}
	if err != nil {
		return ports.LessonTranscript{}, persistenceError(err)
	}
	result := ports.LessonTranscript{Course: domain.Course{
		OwnerID: ownerID, Status: domain.Status(status), Paid: paid, Entitled: membership.Valid && membership.String == "active", Membership: membershipState(membership), Preview: preview,
	}}

	var storedLanguage string
	var generation int32
	var updatedAt time.Time
	err = repository.pool.QueryRow(ctx, getTranscript, lessonID, string(language)).
		Scan(&storedLanguage, &generation, &updatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return result, nil
	}
	if err != nil {
		return ports.LessonTranscript{}, persistenceError(err)
	}

	if err := repository.pool.QueryRow(ctx, readyGenerationForLesson, lessonID).Scan(&result.ReadyGeneration); err != nil {
		return ports.LessonTranscript{}, persistenceError(err)
	}

	segments, library, err := repository.readSegments(ctx, lessonID, storedLanguage)
	if err != nil {
		return ports.LessonTranscript{}, err
	}
	transcript, err := domain.NewTranscript(domain.Transcript{
		LessonID: lessonID, Language: domain.Language(storedLanguage),
		Segments: segments, Documents: library, AuthoredAgainstGeneration: generation,
	})
	if err != nil {
		return ports.LessonTranscript{}, err
	}
	result.Found = true
	result.Transcript = transcript
	result.UpdatedAt = updatedAt.UTC().Format(time.RFC3339Nano)
	return result, nil
}

func membershipState(state pgtype.Text) access.Membership {
	if !state.Valid {
		return access.MembershipNone
	}
	return access.Membership(state.String)
}

func (repository *Repository) readSegments(ctx context.Context, lessonID, language string) ([]domain.Segment, []domain.DocumentLink, error) {
	rows, err := repository.pool.Query(ctx, getSegments, lessonID, language)
	if err != nil {
		return nil, nil, persistenceError(err)
	}
	defer rows.Close()

	segments := make([]domain.Segment, 0)
	index := make(map[string]int)
	documents := make(map[string]domain.DocumentLink)
	order := make([]string, 0)
	for rows.Next() {
		var id, text, speaker, documentID, label, url string
		var startMs, endMs int64
		var confidence pgtype.Float8
		if err := rows.Scan(&id, &startMs, &endMs, &text, &speaker, &confidence, &documentID, &label, &url); err != nil {
			return nil, nil, persistenceError(err)
		}
		position, seen := index[id]
		if !seen {
			segment := domain.Segment{ID: id, StartMs: startMs, EndMs: endMs, Text: text, Speaker: speaker}
			if confidence.Valid {
				value := confidence.Float64
				segment.Confidence = &value
			}
			segments = append(segments, segment)
			position = len(segments) - 1
			index[id] = position
		}
		// A LEFT JOIN repeats the segment once per document, so a segment with no documents
		// arrives with empty columns rather than not arriving at all.
		if documentID != "" {
			segments[position].DocumentIDs = append(segments[position].DocumentIDs, documentID)
			// The library entry itself travels once, no matter how many segments reference it.
			if _, carried := documents[documentID]; !carried {
				documents[documentID] = domain.DocumentLink{ID: documentID, Label: label, URL: url}
				order = append(order, documentID)
			}
		}
	}
	if err := rows.Err(); err != nil {
		return nil, nil, persistenceError(err)
	}
	library := make([]domain.DocumentLink, 0, len(order))
	for _, documentID := range order {
		library = append(library, documents[documentID])
	}
	return segments, library, nil
}

func persistenceError(err error) error {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	return domain.ErrDependencyUnavailable
}
