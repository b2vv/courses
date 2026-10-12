package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"courses/internal/course/domain"
	"courses/internal/course/ports"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

var (
	_ ports.LessonDocumentRepository   = (*Repository)(nil)
	_ ports.LessonDocumentAccessReader = (*Repository)(nil)
)

func (r *Repository) SaveDraft(ctx context.Context, lessonID, locale string, body domain.LessonDocument, expectedUpdatedAt time.Time) (time.Time, error) {
	draftJSON, err := json.Marshal(body)
	if err != nil {
		return time.Time{}, err
	}
	publishedJSON, err := json.Marshal(domain.LessonDocument{SchemaVersion: domain.CurrentLessonDocumentSchemaVersion, Locale: locale, Sections: []domain.Section{}})
	if err != nil {
		return time.Time{}, err
	}
	var updatedAt time.Time
	// A conflicting row whose guard fails is locked but not returned by PostgreSQL.
	err = r.pool.QueryRow(ctx, `
INSERT INTO lesson_documents (lesson_id, locale, draft_body, published_body, updated_at)
VALUES ($1::text::uuid, $2, $3, $4, now())
ON CONFLICT (lesson_id, locale) DO UPDATE
SET draft_body = EXCLUDED.draft_body, updated_at = now()
WHERE lesson_documents.updated_at = $5
RETURNING updated_at`, lessonID, locale, draftJSON, publishedJSON, expectedUpdatedAt).Scan(&updatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return time.Time{}, domain.ErrDocumentConflict
	}
	if err != nil {
		return time.Time{}, persistenceError(err)
	}
	return updatedAt, nil
}

func (r *Repository) PublishDocument(ctx context.Context, lessonID, locale string, expectedUpdatedAt time.Time) error {
	var updatedAt time.Time
	err := r.pool.QueryRow(ctx, `
UPDATE lesson_documents SET published_body = draft_body, updated_at = now()
WHERE lesson_id = $1::text::uuid AND locale = $2 AND updated_at = $3
RETURNING updated_at`, lessonID, locale, expectedUpdatedAt).Scan(&updatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrDocumentConflict
	}
	if err != nil {
		return persistenceError(err)
	}
	return nil
}

func (r *Repository) LoadDraft(ctx context.Context, lessonID, locale string) (body domain.LessonDocument, updatedAt time.Time, found bool, err error) {
	var data []byte
	err = r.pool.QueryRow(ctx, `SELECT draft_body, updated_at FROM lesson_documents WHERE lesson_id = $1::text::uuid AND locale = $2`, lessonID, locale).Scan(&data, &updatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.LessonDocument{}, time.Time{}, false, nil
	}
	if err != nil {
		return domain.LessonDocument{}, time.Time{}, false, persistenceError(err)
	}
	if err := json.Unmarshal(data, &body); err != nil {
		return domain.LessonDocument{}, time.Time{}, false, domain.ErrDependencyUnavailable
	}
	return body, updatedAt, true, nil
}

func (r *Repository) LoadPublished(ctx context.Context, lessonID, locale string) (body domain.LessonDocument, found bool, err error) {
	var data []byte
	err = r.pool.QueryRow(ctx, `SELECT published_body FROM lesson_documents WHERE lesson_id = $1::text::uuid AND locale = $2`, lessonID, locale).Scan(&data)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.LessonDocument{}, false, nil
	}
	if err != nil {
		return domain.LessonDocument{}, false, persistenceError(err)
	}
	if err := json.Unmarshal(data, &body); err != nil {
		return domain.LessonDocument{}, false, domain.ErrDependencyUnavailable
	}
	return body, true, nil
}

func (r *Repository) LoadLessonAccess(ctx context.Context, userID, lessonID string) (domain.LessonDocumentAccess, error) {
	var result domain.LessonDocumentAccess
	var status string
	var priceMinor pgtype.Int8
	var currency, state pgtype.Text
	err := r.pool.QueryRow(ctx, `
SELECT courses.instructor_id::text, courses.status, courses.price_minor, courses.currency,
       grants.state, lessons.is_preview
FROM course_lessons AS lessons
JOIN course_modules AS modules ON modules.id = lessons.module_id
JOIN courses ON courses.id = modules.course_id
LEFT JOIN course_grants AS grants ON grants.user_id = $1::text::uuid AND grants.course_id = courses.id
WHERE lessons.id = $2::text::uuid`, userID, lessonID).Scan(&result.OwnerID, &status, &priceMinor, &currency, &state, &result.Preview)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.LessonDocumentAccess{}, domain.ErrLessonNotFound
	}
	if err != nil {
		return domain.LessonDocumentAccess{}, persistenceError(err)
	}
	if priceMinor.Valid != currency.Valid {
		return domain.LessonDocumentAccess{}, domain.ErrDependencyUnavailable
	}
	result.Status = domain.Status(status)
	result.Paid = priceMinor.Valid && currency.Valid
	result.Entitled = state.Valid && state.String == "active"
	result.Membership = membership(state)
	return result, nil
}
