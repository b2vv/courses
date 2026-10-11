package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"courses/internal/contracts/rating"
	"courses/internal/course/adapters/postgres/generated"
	"courses/internal/course/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct{ pool *pgxpool.Pool }

func NewRepository(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

func (r *Repository) ListPublished(ctx context.Context, input domain.Query) (domain.Page, error) {
	query, err := domain.NewQuery(input.Limit, input.Offset, input.Search, string(input.Sort))
	if err != nil {
		return domain.Page{}, err
	}
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return domain.Page{}, persistenceError(err)
	}
	defer rollback(tx)
	queries := generated.New(tx)
	// query.Search goes straight into websearch_to_tsquery (full-text match, not an
	// ILIKE pattern), so no wildcard escaping is needed here.
	rows, err := queries.ListPublishedCourses(ctx, generated.ListPublishedCoursesParams{SearchPattern: query.Search, SortBy: string(query.Sort), PageLimit: int32(query.Limit), PageOffset: int32(query.Offset)})
	if err != nil {
		return domain.Page{}, persistenceError(err)
	}
	total, err := queries.CountPublishedCourses(ctx, query.Search)
	if err != nil {
		return domain.Page{}, persistenceError(err)
	}
	items := make([]domain.Course, 0, len(rows))
	for _, row := range rows {
		course, err := mapCourse(row.ID, row.Slug, row.Title, row.Summary, row.Description, row.Status, row.CreatedAt, row.UpdatedAt, row.PublishedAt, row.PriceMinor, row.Currency, row.ReviewCount, row.ReviewStarsSum, row.LessonCount)
		if err != nil {
			return domain.Page{}, err
		}
		items = append(items, course)
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.Page{}, persistenceError(err)
	}
	return domain.Page{Items: items, Limit: query.Limit, Offset: query.Offset, Total: total}, nil
}

func (r *Repository) GetPublished(ctx context.Context, slug string) (domain.CourseDetail, error) {
	if err := domain.ValidateSlug(slug); err != nil {
		return domain.CourseDetail{}, err
	}
	rows, err := generated.New(r.pool).GetPublishedCourse(ctx, slug)
	if err != nil {
		return domain.CourseDetail{}, persistenceError(err)
	}
	if len(rows) == 0 {
		return domain.CourseDetail{}, domain.ErrCourseNotFound
	}
	row := rows[0]
	course, err := mapCourse(row.ID, row.Slug, row.Title, row.Summary, row.Description, row.Status, row.CreatedAt, row.UpdatedAt, row.PublishedAt, row.PriceMinor, row.Currency, row.ReviewCount, row.ReviewStarsSum, row.LessonCount)
	if err != nil {
		return domain.CourseDetail{}, err
	}
	program := make([]domain.ProgramModule, 0)
	moduleID := ""
	for _, row := range rows {
		if row.ModuleID == "" {
			continue
		}
		if row.ModuleID != moduleID {
			program = append(program, domain.ProgramModule{Title: row.ModuleTitle, Lessons: []domain.ProgramLesson{}})
			moduleID = row.ModuleID
		}
		if row.LessonPresent {
			program[len(program)-1].Lessons = append(program[len(program)-1].Lessons, domain.ProgramLesson{Title: row.LessonTitle})
		}
	}
	return domain.CourseDetail{Course: course, Program: program}, nil
}

func persistenceError(err error) error {
	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) {
		// Connection, resource, shutdown, rollback and lock failures are transient.
		for _, class := range []string{"08", "53", "57", "58", "40", "55"} {
			if strings.HasPrefix(postgresError.Code, class) {
				return domain.ErrDependencyUnavailable
			}
		}
		return fmt.Errorf("course persistence failure: %w", err)
	}
	return domain.ErrDependencyUnavailable
}

func mapCourse(id pgtype.UUID, slug, title, summary, description, status string, createdAt, updatedAt, storedPublishedAt pgtype.Timestamptz, priceMinor pgtype.Int8, currency pgtype.Text, reviewCount, reviewStarsSum, lessonCount int64) (domain.Course, error) {
	var publishedAt *time.Time
	if storedPublishedAt.Valid {
		value := storedPublishedAt.Time
		publishedAt = &value
	}
	var price *domain.Price
	if priceMinor.Valid && currency.Valid {
		price = &domain.Price{Minor: priceMinor.Int64, Currency: currency.String}
	} else if priceMinor.Valid != currency.Valid {
		return domain.Course{}, domain.ErrDependencyUnavailable
	}
	if reviewCount < 0 || reviewStarsSum < 0 || lessonCount < 0 || reviewCount > int64(^uint(0)>>1) || reviewStarsSum > int64(^uint(0)>>1) || lessonCount > int64(^uint(0)>>1) {
		return domain.Course{}, domain.ErrDependencyUnavailable
	}
	count, sum := int(reviewCount), int(reviewStarsSum)
	var ratingSummary *domain.Rating
	if average, visible := rating.Summary(count, sum); visible {
		ratingSummary = &domain.Rating{Average: average, Count: count}
	}
	return domain.NewCourse(domain.Course{ID: id.String(), Slug: slug, Title: title, Summary: summary, Description: description, Status: domain.Status(status), CreatedAt: createdAt.Time, UpdatedAt: updatedAt.Time, PublishedAt: publishedAt, Price: price, Paid: price != nil, Rating: ratingSummary, LessonCount: int(lessonCount)})
}

func rollback(tx pgx.Tx) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	// Commit and context cancellation can already have closed the transaction.
	// pgx closes a broken connection itself if rollback fails.
	_ = tx.Rollback(ctx)
}
