package postgres

import (
	"context"
	"errors"
	"time"

	"courses/internal/transcript/domain"
	"courses/internal/transcript/ports"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const getLessonCourse = `SELECT courses.instructor_id::text, courses.status
FROM course_lessons AS lessons
JOIN course_modules AS modules ON modules.id = lessons.module_id
JOIN courses ON courses.id = modules.course_id
WHERE lessons.id = $1::text::uuid`

func (repository *Repository) LessonCourse(ctx context.Context, lessonID string) (domain.Course, error) {
	var ownerID, status string
	err := repository.pool.QueryRow(ctx, getLessonCourse, lessonID).Scan(&ownerID, &status)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Course{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Course{}, persistenceError(err)
	}
	return domain.Course{OwnerID: ownerID, Status: domain.Status(status)}, nil
}

// SaveLessonTranscript replaces the whole transcript in ONE transaction. The old segments are
// deleted and the new ones inserted inside it, so there is never a window where the lesson shows
// half a transcript: a partially written transcript looks written and is not.
func (repository *Repository) SaveLessonTranscript(ctx context.Context, save ports.SaveTranscript) (string, error) {
	transaction, err := repository.pool.Begin(ctx)
	if err != nil {
		return "", persistenceError(err)
	}
	defer func() { _ = transaction.Rollback(ctx) }()

	// The generation is stamped from the ready asset AT WRITE TIME. Staleness is then derived on
	// every read by comparing it with the generation the player would pick, so no second write
	// exists that could fall out of step.
	var generation int32
	if err := transaction.QueryRow(ctx, readyGenerationForLesson, save.LessonID).Scan(&generation); err != nil {
		return "", persistenceError(err)
	}

	var updatedAt time.Time
	switch save.ExpectedUpdatedAt {
	case "":
		// No token means the author believes nothing is there. The unique key is what makes that
		// belief checkable: a second writer racing into the same empty lesson loses here rather
		// than appending to the first one's work.
		// The first transcript of a lesson is its primary (B-21): a no-language read returns it,
		// and a later language added here arrives non-primary so it cannot displace that read. The
		// one-primary index is what makes this true even under a race, not this EXISTS alone.
		err = transaction.QueryRow(ctx, `INSERT INTO transcripts (lesson_id, language, authored_against_generation, is_primary)
			VALUES ($1::text::uuid, $2, $3,
			  NOT EXISTS (SELECT 1 FROM transcripts WHERE lesson_id = $1::text::uuid AND is_primary))
			RETURNING updated_at`,
			save.LessonID, string(save.Language), generation).Scan(&updatedAt)
		var conflict *pgconn.PgError
		if errors.As(err, &conflict) && conflict.Code == "23505" {
			return "", domain.ErrTranscriptExists
		}
	default:
		err = transaction.QueryRow(ctx, `UPDATE transcripts
			SET authored_against_generation = $3, updated_at = now()
			WHERE lesson_id = $1::text::uuid AND language = $2 AND updated_at = $4
			RETURNING updated_at`,
			save.LessonID, string(save.Language), generation, save.ExpectedUpdatedAt).Scan(&updatedAt)
		if errors.Is(err, pgx.ErrNoRows) {
			return "", domain.ErrTranscriptModified
		}
	}
	if err != nil {
		return "", persistenceError(err)
	}

	if _, err := transaction.Exec(ctx,
		`DELETE FROM transcript_segments WHERE lesson_id = $1::text::uuid AND language = $2`,
		save.LessonID, string(save.Language)); err != nil {
		return "", persistenceError(err)
	}
	for ordinal, segment := range save.Segments {
		var speaker *string
		if segment.Speaker != "" {
			value := segment.Speaker
			speaker = &value
		}
		if _, err := transaction.Exec(ctx, `INSERT INTO transcript_segments
			(id, lesson_id, language, ordinal, start_ms, end_ms, text, speaker, confidence)
			VALUES ($1::text::uuid, $2::text::uuid, $3, $4, $5, $6, $7, $8, $9)`,
			segment.ID, save.LessonID, string(save.Language), ordinal,
			segment.StartMs, segment.EndMs, segment.Text, speaker, segment.Confidence); err != nil {
			return "", persistenceError(err)
		}
		for position, documentID := range segment.DocumentIDs {
			if _, err := transaction.Exec(ctx, `INSERT INTO segment_documents (segment_id, document_id, position)
				VALUES ($1::text::uuid, $2::text::uuid, $3)`, segment.ID, documentID, position); err != nil {
				var violation *pgconn.PgError
				if errors.As(err, &violation) && violation.Code == "23503" {
					// A reference to a document that is not in this course's library is a named
					// refusal, not a 500: the editor can tell the author which link is unknown.
					return "", domain.ErrInvalidRequest
				}
				return "", persistenceError(err)
			}
		}
	}
	if err := transaction.Commit(ctx); err != nil {
		return "", persistenceError(err)
	}
	return updatedAt.UTC().Format(time.RFC3339Nano), nil
}
