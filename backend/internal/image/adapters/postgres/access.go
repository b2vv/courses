package postgres

import (
	"context"
	"errors"
	"fmt"

	"courses/internal/contracts/access"
	"courses/internal/image/domain"
	"courses/internal/image/ports"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

var (
	_ ports.AccessReader    = (*Repository)(nil)
	_ ports.AuthoringReader = (*Repository)(nil)
)

func (r *Repository) LoadImageAccess(ctx context.Context, userID, imageID string) (domain.ImageAccess, error) {
	var result domain.ImageAccess
	var status string
	var priceMinor pgtype.Int8
	var currency, state pgtype.Text
	err := r.pool.QueryRow(ctx, `
SELECT courses.instructor_id::text, courses.status, courses.price_minor, courses.currency,
       grants.state, lessons.is_preview
FROM lesson_images AS images
JOIN course_lessons AS lessons ON lessons.id = images.lesson_id
JOIN course_modules AS modules ON modules.id = lessons.module_id
JOIN courses ON courses.id = modules.course_id
LEFT JOIN course_grants AS grants ON grants.user_id = $1::text::uuid AND grants.course_id = courses.id
WHERE images.id = $2::text::uuid`, userID, imageID).Scan(&result.OwnerID, &status, &priceMinor, &currency, &state, &result.Preview)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ImageAccess{}, domain.ErrImageNotFound
	}
	if err != nil {
		return domain.ImageAccess{}, persistenceError(err)
	}
	if priceMinor.Valid != currency.Valid {
		return domain.ImageAccess{}, fmt.Errorf("image course facts inconsistent: price/currency mismatch")
	}
	result.Published = status == "published"
	result.Paid = priceMinor.Valid && currency.Valid
	result.Entitled = state.Valid && state.String == "active"
	result.Membership = membership(state)
	return result, nil
}

func (r *Repository) LoadLessonOwner(ctx context.Context, lessonID string) (string, bool, error) {
	var ownerID string
	err := r.pool.QueryRow(ctx, `
SELECT courses.instructor_id::text
FROM course_lessons AS lessons
JOIN course_modules AS modules ON modules.id = lessons.module_id
JOIN courses ON courses.id = modules.course_id
WHERE lessons.id = $1::text::uuid`, lessonID).Scan(&ownerID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, persistenceError(err)
	}
	return ownerID, true, nil
}

func membership(state pgtype.Text) access.Membership {
	if !state.Valid {
		return access.MembershipNone
	}
	return access.Membership(state.String)
}
