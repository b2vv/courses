package postgres

import (
	"context"
	"errors"

	"courses/internal/transcript/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func (repository *Repository) CourseOfID(ctx context.Context, courseID string) (domain.Course, error) {
	var ownerID, status string
	err := repository.pool.QueryRow(ctx,
		`SELECT instructor_id::text, status FROM courses WHERE id = $1::text::uuid`, courseID).
		Scan(&ownerID, &status)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Course{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Course{}, persistenceError(err)
	}
	return domain.Course{OwnerID: ownerID, Status: domain.Status(status)}, nil
}

func (repository *Repository) CourseDocuments(ctx context.Context, courseID string) ([]domain.DocumentLink, error) {
	rows, err := repository.pool.Query(ctx,
		`SELECT id::text, label, url FROM course_documents WHERE course_id = $1::text::uuid ORDER BY label, id`,
		courseID)
	if err != nil {
		return nil, persistenceError(err)
	}
	defer rows.Close()
	documents := make([]domain.DocumentLink, 0)
	for rows.Next() {
		var link domain.DocumentLink
		if err := rows.Scan(&link.ID, &link.Label, &link.URL); err != nil {
			return nil, persistenceError(err)
		}
		documents = append(documents, link)
	}
	if err := rows.Err(); err != nil {
		return nil, persistenceError(err)
	}
	return documents, nil
}

func (repository *Repository) AddCourseDocument(ctx context.Context, courseID string, link domain.DocumentLink) error {
	transaction, err := repository.pool.Begin(ctx)
	if err != nil {
		return persistenceError(err)
	}
	defer func() { _ = transaction.Rollback(ctx) }()

	// The COURSE row is locked, not the count: PostgreSQL refuses FOR UPDATE together with an
	// aggregate, and locking the parent is what actually serialises two concurrent additions —
	// locking the existing document rows would not stop a second writer inserting a new one.
	// Without this, two additions both see room at 199 and the library ends at 201.
	var locked string
	if err := transaction.QueryRow(ctx,
		`SELECT id::text FROM courses WHERE id = $1::text::uuid FOR UPDATE`, courseID).Scan(&locked); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		return persistenceError(err)
	}
	var count int
	if err := transaction.QueryRow(ctx,
		`SELECT count(*) FROM course_documents WHERE course_id = $1::text::uuid`, courseID).
		Scan(&count); err != nil {
		return persistenceError(err)
	}
	if count >= domain.MaximumCourseDocuments {
		return domain.ErrDocumentLimit
	}
	_, err = transaction.Exec(ctx,
		`INSERT INTO course_documents (id, course_id, label, url) VALUES ($1::text::uuid, $2::text::uuid, $3, $4)`,
		link.ID, courseID, link.Label, link.URL)
	var conflict *pgconn.PgError
	if errors.As(err, &conflict) && conflict.Code == "23505" {
		return domain.ErrDocumentExists
	}
	if err != nil {
		return persistenceError(err)
	}
	if err := transaction.Commit(ctx); err != nil {
		return persistenceError(err)
	}
	return nil
}

func (repository *Repository) RemoveCourseDocument(ctx context.Context, courseID, documentID string) error {
	var referenced bool
	if err := repository.pool.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM segment_documents WHERE document_id = $1::text::uuid)`, documentID).
		Scan(&referenced); err != nil {
		return persistenceError(err)
	}
	if referenced {
		return domain.ErrDocumentReferenced
	}
	tag, err := repository.pool.Exec(ctx,
		`DELETE FROM course_documents WHERE id = $1::text::uuid AND course_id = $2::text::uuid`,
		documentID, courseID)
	if err != nil {
		return persistenceError(err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
