package postgres

import (
	"context"
	"errors"
	"time"

	"courses/internal/course/adapters/postgres/generated"
	"courses/internal/course/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

const upsertAuthoringModule = `
INSERT INTO course_modules (id, course_id, title, position, created_at, updated_at)
VALUES ($1::text::uuid, $2::text::uuid, $3, $4, $5, $5)
ON CONFLICT (id) DO UPDATE
SET title = EXCLUDED.title, position = EXCLUDED.position, updated_at = EXCLUDED.updated_at
WHERE course_modules.course_id = $2::text::uuid`

const upsertAuthoringLesson = `
INSERT INTO course_lessons (id, module_id, title, description, position, is_preview, created_at, updated_at)
VALUES ($1::text::uuid, $3::text::uuid, $4, $5, $6, $7, $8, $8)
ON CONFLICT (id) DO UPDATE
SET module_id = EXCLUDED.module_id, title = EXCLUDED.title,
    description = EXCLUDED.description, position = EXCLUDED.position, is_preview = EXCLUDED.is_preview,
    updated_at = EXCLUDED.updated_at
WHERE EXISTS (
  SELECT 1
  FROM course_modules AS current_module
  WHERE current_module.id = course_lessons.module_id
    AND current_module.course_id = $2::text::uuid
)`

const deleteMissingAuthoringLessons = `
DELETE FROM course_lessons AS lesson
WHERE EXISTS (
  SELECT 1 FROM course_modules AS module
  WHERE module.id = lesson.module_id AND module.course_id = $1::text::uuid
)
AND NOT (lesson.id::text = ANY($2::text[]))`

const deleteMissingAuthoringModules = `
DELETE FROM course_modules
WHERE course_id = $1::text::uuid
  AND NOT (id::text = ANY($2::text[]))`

const updateAuthoringCourse = `
UPDATE courses
SET slug = $4, title = $5, summary = $6, description = $7,
    instructor_id = $8::text::uuid, price_minor = $9, currency = $10, updated_at = $3
WHERE id = $1::text::uuid AND updated_at = $2`

const transitionAuthoringCourse = `
UPDATE courses
SET status = $3,
    published_at = CASE
      WHEN $3 = 'published' THEN COALESCE(published_at, $5)
      ELSE published_at
    END,
    updated_at = $5
WHERE id = $1::text::uuid
  AND status = $2
  AND updated_at = $4
  AND ($6 OR instructor_id = $7::text::uuid)`

const deleteAuthoringCourse = `
DELETE FROM courses
WHERE id = $1::text::uuid
  AND status = 'draft'
  AND updated_at = $2
  AND ($3 OR instructor_id = $4::text::uuid)`

func (r *Repository) List(ctx context.Context, actor domain.Actor, query domain.Query, status domain.Status) (domain.AuthoringPage, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return domain.AuthoringPage{}, persistenceError(err)
	}
	defer rollback(tx)

	queries := generated.New(tx)
	params := generated.ListAuthoringCoursesParams{
		IsAdmin:    actor.Role == domain.ActorAdmin,
		ActorID:    actor.UserID,
		Status:     string(status),
		PageLimit:  int32(query.Limit),
		PageOffset: int32(query.Offset),
	}
	rows, err := queries.ListAuthoringCourses(ctx, params)
	if err != nil {
		return domain.AuthoringPage{}, persistenceError(err)
	}
	total, err := queries.CountAuthoringCourses(ctx, generated.CountAuthoringCoursesParams{
		IsAdmin: actor.Role == domain.ActorAdmin,
		ActorID: actor.UserID,
		Status:  string(status),
	})
	if err != nil {
		return domain.AuthoringPage{}, persistenceError(err)
	}

	items := make([]domain.AuthoringCourse, 0, len(rows))
	for _, row := range rows {
		if !row.UpdatedAt.Valid {
			return domain.AuthoringPage{}, domain.ErrDependencyUnavailable
		}
		items = append(items, domain.AuthoringCourse{
			ID:           row.ID,
			Slug:         row.Slug,
			Title:        row.Title,
			Status:       domain.Status(row.Status),
			InstructorID: row.InstructorID,
			UpdatedAt:    row.UpdatedAt.Time,
		})
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.AuthoringPage{}, persistenceError(err)
	}
	return domain.AuthoringPage{Items: items, Limit: query.Limit, Offset: query.Offset, Total: total}, nil
}

func (r *Repository) Load(ctx context.Context, id string) (domain.CourseTree, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return domain.CourseTree{}, persistenceError(err)
	}
	defer rollback(tx)

	tree, err := loadAuthoringTree(ctx, generated.New(tx), id)
	if err != nil {
		return domain.CourseTree{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.CourseTree{}, persistenceError(err)
	}
	return tree, nil
}

func (r *Repository) Create(ctx context.Context, actor domain.Actor, draft domain.CourseDraft, now time.Time) (domain.CourseTree, error) {
	if err := actor.CanAssignOwner(draft.InstructorID); err != nil {
		return domain.CourseTree{}, err
	}
	if actor.Role == domain.ActorInstructor && draft.InstructorID == "" {
		draft.InstructorID = actor.UserID
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.CourseTree{}, persistenceError(err)
	}
	defer rollback(tx)

	queries := generated.New(tx)
	if _, err := queries.LockEligibleAuthoringOwner(ctx, draft.InstructorID); errors.Is(err, pgx.ErrNoRows) {
		return domain.CourseTree{}, domain.ErrInvalidOwner
	} else if err != nil {
		return domain.CourseTree{}, persistenceError(err)
	}

	createdAt := pgtype.Timestamptz{Time: now.UTC().Truncate(time.Microsecond), Valid: true}
	var priceMinor pgtype.Int8
	var currency pgtype.Text
	if draft.Price != nil {
		priceMinor = pgtype.Int8{Int64: draft.Price.Minor, Valid: true}
		currency = pgtype.Text{String: draft.Price.Currency, Valid: true}
	}
	id, err := queries.CreateAuthoringCourse(ctx, generated.CreateAuthoringCourseParams{
		OwnerID:     draft.InstructorID,
		Slug:        draft.Slug,
		Title:       draft.Title,
		Summary:     draft.Summary,
		Description: draft.Description,
		CreatedAt:   createdAt,
		PriceMinor:  priceMinor,
		Currency:    currency,
	})
	if err != nil {
		if slugTaken(err) {
			return domain.CourseTree{}, domain.ErrSlugTaken
		}
		return domain.CourseTree{}, persistenceError(err)
	}
	tree, err := loadAuthoringTree(ctx, queries, id)
	if err != nil {
		return domain.CourseTree{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.CourseTree{}, persistenceError(err)
	}
	return tree, nil
}

func (r *Repository) Save(ctx context.Context, actor domain.Actor, tree domain.CourseTree, expectedUpdatedAt, updatedAt time.Time) (domain.CourseTree, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.CourseTree{}, persistenceError(err)
	}
	defer rollback(tx)

	queries := generated.New(tx)
	current, err := lockAuthoringCourse(ctx, queries, tree.ID)
	if err != nil {
		return domain.CourseTree{}, err
	}
	if err := actor.CanManage(current.InstructorID); err != nil {
		return domain.CourseTree{}, err
	}
	if err := actor.CanAssignOwner(tree.InstructorID); err != nil {
		return domain.CourseTree{}, err
	}
	if err := domain.ValidateSlugChange(current.Slug, current.PublishedAt, tree.Slug); err != nil {
		return domain.CourseTree{}, err
	}
	if _, err := queries.LockEligibleAuthoringOwner(ctx, tree.InstructorID); errors.Is(err, pgx.ErrNoRows) {
		return domain.CourseTree{}, domain.ErrInvalidOwner
	} else if err != nil {
		return domain.CourseTree{}, persistenceError(err)
	}

	writeTime := nextAuthoringTimestamp(updatedAt, current.UpdatedAt)
	moduleIDs := make([]string, 0, len(tree.Modules))
	lessonIDs := make([]string, 0)
	for _, module := range tree.Modules {
		result, err := tx.Exec(ctx, upsertAuthoringModule, module.ID, tree.ID, module.Title, module.Position, writeTime)
		if err != nil {
			return domain.CourseTree{}, persistenceError(err)
		}
		if result.RowsAffected() != 1 {
			return domain.CourseTree{}, domain.ErrInvalidRequest
		}
		moduleIDs = append(moduleIDs, module.ID)
	}
	for _, module := range tree.Modules {
		for _, lesson := range module.Lessons {
			result, err := tx.Exec(ctx, upsertAuthoringLesson, lesson.ID, tree.ID, module.ID, lesson.Title, lesson.Description, lesson.Position, lesson.IsPreview, writeTime)
			if err != nil {
				return domain.CourseTree{}, persistenceError(err)
			}
			if result.RowsAffected() != 1 {
				return domain.CourseTree{}, domain.ErrInvalidRequest
			}
			lessonIDs = append(lessonIDs, lesson.ID)
		}
	}
	if _, err := tx.Exec(ctx, deleteMissingAuthoringLessons, tree.ID, lessonIDs); err != nil {
		return domain.CourseTree{}, persistenceError(err)
	}
	if _, err := tx.Exec(ctx, deleteMissingAuthoringModules, tree.ID, moduleIDs); err != nil {
		return domain.CourseTree{}, persistenceError(err)
	}

	var priceMinor any
	var currency any
	if tree.Price != nil {
		priceMinor = tree.Price.Minor
		currency = tree.Price.Currency
	}
	result, err := tx.Exec(ctx, updateAuthoringCourse,
		tree.ID, expectedUpdatedAt, writeTime, tree.Slug, tree.Title,
		tree.Summary, tree.Description, tree.InstructorID, priceMinor, currency,
	)
	if err != nil {
		if slugTaken(err) {
			return domain.CourseTree{}, domain.ErrSlugTaken
		}
		return domain.CourseTree{}, persistenceError(err)
	}
	if result.RowsAffected() != 1 {
		if _, err := queries.GetAuthoringCourse(ctx, tree.ID); errors.Is(err, pgx.ErrNoRows) {
			return domain.CourseTree{}, domain.ErrCourseNotFound
		} else if err != nil {
			return domain.CourseTree{}, persistenceError(err)
		}
		return domain.CourseTree{}, domain.ErrCourseModified
	}

	saved, err := loadAuthoringTree(ctx, queries, tree.ID)
	if err != nil {
		return domain.CourseTree{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.CourseTree{}, persistenceError(err)
	}
	return saved, nil
}

func (r *Repository) Transition(ctx context.Context, actor domain.Actor, id string, from, to domain.Status, expectedUpdatedAt, updatedAt time.Time) (domain.CourseTree, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.CourseTree{}, persistenceError(err)
	}
	defer rollback(tx)

	writeTime := nextAuthoringTimestamp(updatedAt, expectedUpdatedAt)
	result, err := tx.Exec(ctx, transitionAuthoringCourse,
		id, string(from), string(to), expectedUpdatedAt, writeTime,
		actor.Role == domain.ActorAdmin, actor.UserID,
	)
	if err != nil {
		return domain.CourseTree{}, persistenceError(err)
	}
	queries := generated.New(tx)
	if result.RowsAffected() != 1 {
		current, err := lockAuthoringCourse(ctx, queries, id)
		if err != nil {
			return domain.CourseTree{}, err
		}
		if err := actor.CanManage(current.InstructorID); err != nil {
			return domain.CourseTree{}, err
		}
		if current.Status != from {
			return domain.CourseTree{}, domain.ErrInvalidStatusTransition
		}
		if !current.UpdatedAt.Equal(expectedUpdatedAt) {
			return domain.CourseTree{}, domain.ErrCourseModified
		}
		return domain.CourseTree{}, domain.ErrDependencyUnavailable
	}

	transitioned, err := loadAuthoringTree(ctx, queries, id)
	if err != nil {
		return domain.CourseTree{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.CourseTree{}, persistenceError(err)
	}
	return transitioned, nil
}

func (r *Repository) Delete(ctx context.Context, actor domain.Actor, id string, expectedUpdatedAt time.Time) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return persistenceError(err)
	}
	defer rollback(tx)

	result, err := tx.Exec(ctx, deleteAuthoringCourse, id, expectedUpdatedAt, actor.Role == domain.ActorAdmin, actor.UserID)
	if err != nil {
		return persistenceError(err)
	}
	if result.RowsAffected() != 1 {
		current, err := lockAuthoringCourse(ctx, generated.New(tx), id)
		if err != nil {
			return err
		}
		if err := actor.CanManage(current.InstructorID); err != nil {
			return err
		}
		if current.Status != domain.Draft {
			return domain.ErrCourseNotDeletable
		}
		if !current.UpdatedAt.Equal(expectedUpdatedAt) {
			return domain.ErrCourseModified
		}
		return domain.ErrDependencyUnavailable
	}
	if err := tx.Commit(ctx); err != nil {
		return persistenceError(err)
	}
	return nil
}

func loadAuthoringTree(ctx context.Context, queries *generated.Queries, id string) (domain.CourseTree, error) {
	row, err := queries.GetAuthoringCourse(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.CourseTree{}, domain.ErrCourseNotFound
	}
	if err != nil {
		return domain.CourseTree{}, persistenceError(err)
	}
	tree, err := mapAuthoringTree(
		row.ID, row.Slug, row.Title, row.Summary, row.Description, row.Status,
		row.InstructorID, row.CreatedAt, row.UpdatedAt, row.PublishedAt,
		row.PriceMinor, row.Currency,
	)
	if err != nil {
		return domain.CourseTree{}, err
	}

	modules, err := queries.ListAuthoringModules(ctx, id)
	if err != nil {
		return domain.CourseTree{}, persistenceError(err)
	}
	tree.Modules = make([]domain.Module, 0, len(modules))
	moduleIndex := make(map[string]int, len(modules))
	for _, module := range modules {
		moduleIndex[module.ID] = len(tree.Modules)
		tree.Modules = append(tree.Modules, domain.Module{
			ID: module.ID, Title: module.Title, Position: int(module.Position), Lessons: []domain.Lesson{},
		})
	}
	lessons, err := queries.ListAuthoringLessons(ctx, id)
	if err != nil {
		return domain.CourseTree{}, persistenceError(err)
	}
	for _, lesson := range lessons {
		index, exists := moduleIndex[lesson.ModuleID]
		if !exists {
			return domain.CourseTree{}, domain.ErrDependencyUnavailable
		}
		tree.Modules[index].Lessons = append(tree.Modules[index].Lessons, domain.Lesson{
			ID: lesson.ID, Title: lesson.Title, Description: lesson.Description, Position: int(lesson.Position), IsPreview: lesson.IsPreview,
		})
	}
	return tree, nil
}

func lockAuthoringCourse(ctx context.Context, queries *generated.Queries, id string) (domain.CourseTree, error) {
	row, err := queries.LockAuthoringCourse(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.CourseTree{}, domain.ErrCourseNotFound
	}
	if err != nil {
		return domain.CourseTree{}, persistenceError(err)
	}
	return mapAuthoringTree(
		row.ID, row.Slug, row.Title, row.Summary, row.Description, row.Status,
		row.InstructorID, row.CreatedAt, row.UpdatedAt, row.PublishedAt,
		row.PriceMinor, row.Currency,
	)
}

func mapAuthoringTree(
	id, slug, title, summary, description, status, instructorID string,
	createdAt, updatedAt, publishedAt pgtype.Timestamptz,
	priceMinor pgtype.Int8, currency pgtype.Text,
) (domain.CourseTree, error) {
	if !createdAt.Valid || !updatedAt.Valid {
		return domain.CourseTree{}, domain.ErrDependencyUnavailable
	}
	var published *time.Time
	if publishedAt.Valid {
		value := publishedAt.Time
		published = &value
	}
	var price *domain.Price
	if priceMinor.Valid && currency.Valid {
		price = &domain.Price{Minor: priceMinor.Int64, Currency: currency.String}
	} else if priceMinor.Valid != currency.Valid {
		return domain.CourseTree{}, domain.ErrDependencyUnavailable
	}
	return domain.CourseTree{
		ID: id, Slug: slug, Title: title, Summary: summary, Description: description,
		Status: domain.Status(status), InstructorID: instructorID,
		CreatedAt: createdAt.Time, UpdatedAt: updatedAt.Time, PublishedAt: published, Price: price,
	}, nil
}

func slugTaken(err error) bool {
	var postgresError *pgconn.PgError
	return errors.As(err, &postgresError) && postgresError.Code == "23505" && postgresError.ConstraintName == "courses_slug_key"
}

func nextAuthoringTimestamp(clock, after time.Time) time.Time {
	timestamp := clock.UTC().Truncate(time.Microsecond)
	after = after.UTC().Truncate(time.Microsecond)
	if !timestamp.After(after) {
		return after.Add(time.Microsecond)
	}
	return timestamp
}
