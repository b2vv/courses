//go:build measure

// Run: DOCKER_CONFIG=/tmp/claude-501/dockercfg go test -tags measure ./internal/course/adapters/postgres -run TestCatalogAggregateMeasurement -v
package postgres_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"

	containers "courses/internal/testsupport/containers"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	postgrescontainer "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestCatalogAggregateMeasurement(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	container, err := containers.Start(ctx, func(ctx context.Context) (*postgrescontainer.PostgresContainer, error) {
		return postgrescontainer.Run(ctx, "postgres:17.5", postgrescontainer.WithDatabase("courses_measure"), postgrescontainer.WithUsername("courses"), postgrescontainer.WithPassword("ephemeral-test-only"), testcontainers.WithWaitStrategy(wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(3*time.Minute)))
	})
	if err != nil {
		t.Fatalf("PostgreSQL measurement requires Docker: %v", err)
	}
	t.Cleanup(func() { _ = container.Terminate(context.Background()) })
	connection, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("pgx", connection)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	provider, err := goose.NewProvider(goose.DialectPostgres, db, os.DirFS(filepath.Join("..", "..", "..", "..", "migrations")), goose.WithDisableGlobalRegistry(true))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := provider.Up(ctx); err != nil {
		t.Fatal(err)
	}
	seedCatalogMeasurement(t, ctx, db)
	// Steady state, not post-bulk-load: without VACUUM the visibility map is empty (index-only scans
	// fetch the heap) and without ANALYZE the planner guesses row counts. Autovacuum does both in a
	// running database; the measurement must not charge the aggregate for their absence.
	if _, err := db.ExecContext(ctx, `VACUUM ANALYZE`); err != nil {
		t.Fatal(err)
	}
	// The production texts of ListPublishedCourses (backend/sql/queries/courses.sql) with the catalog's
	// default arguments: empty search, sort "recent", limit 20, offset 0. Selecting the aggregate columns
	// matters: an unreferenced LEFT JOIN LATERAL can be removed by the planner, which would measure nothing.
	withAggregate := `SELECT courses.id, courses.slug, courses.title, courses.summary, courses.description, courses.status, courses.created_at, courses.updated_at, courses.published_at,
  courses.price_minor, courses.currency,
  COALESCE(reviews.review_count, 0)::bigint AS review_count,
  COALESCE(reviews.review_stars_sum, 0)::bigint AS review_stars_sum,
  COALESCE(lessons.lesson_count, 0)::bigint AS lesson_count
FROM courses
LEFT JOIN LATERAL (
  SELECT count(*) AS review_count, COALESCE(sum(stars), 0)::bigint AS review_stars_sum
  FROM discussion_reviews
  WHERE course_id = courses.id AND removed_at IS NULL
) AS reviews ON true
LEFT JOIN LATERAL (
  SELECT count(*) AS lesson_count
  FROM course_lessons
  JOIN course_modules ON course_modules.id = course_lessons.module_id
  WHERE course_modules.course_id = courses.id
) AS lessons ON true
WHERE status = 'published'
  AND (title ILIKE '%' || ''::text || '%' ESCAPE '!' OR summary ILIKE '%' || ''::text || '%' ESCAPE '!')
ORDER BY
  CASE WHEN 'recent'::text = 'title' THEN title COLLATE "C" END ASC,
  CASE WHEN 'recent'::text = 'recent' THEN published_at END DESC,
  courses.id ASC
LIMIT 20 OFFSET 0`
	withoutAggregate := `SELECT id, slug, title, summary, description, status, created_at, updated_at, published_at,
  price_minor, currency
FROM courses
WHERE status = 'published'
  AND (title ILIKE '%' || ''::text || '%' ESCAPE '!' OR summary ILIKE '%' || ''::text || '%' ESCAPE '!')
ORDER BY
  CASE WHEN 'recent'::text = 'title' THEN title COLLATE "C" END ASC,
  CASE WHEN 'recent'::text = 'recent' THEN published_at END DESC,
  id ASC
LIMIT 20 OFFSET 0`
	withTimes, withPlan := explainRuns(t, ctx, db, withAggregate)
	withoutTimes, withoutPlan := explainRuns(t, ctx, db, withoutAggregate)
	withMedian, withoutMedian := median(withTimes), median(withoutTimes)
	t.Logf("R6 catalog aggregate median=%s baseline median=%s ratio=%.2fx\nWITH:\n%s\nWITHOUT:\n%s", withMedian, withoutMedian, float64(withMedian)/float64(withoutMedian), withPlan, withoutPlan)
	// R6 budget (owner, 2026-10-03, docs/ratings/assumptions.md): the catalog page with the aggregate
	// must stay within 5 ms; the baseline and ratio are printed for context only.
	if budget := 5 * time.Millisecond; withMedian > budget {
		t.Fatalf("R6 budget exceeded: catalog page median %s > %s — next step is the materialized view (spec R6)", withMedian, budget)
	}
}

func seedCatalogMeasurement(t *testing.T, ctx context.Context, db *sql.DB) {
	t.Helper()
	_, err := db.ExecContext(ctx, `SELECT setseed(0.42); INSERT INTO users(id,email,display_name,role,password_hash) SELECT ('40000000-0000-4000-8000-'||lpad(n::text,12,'0'))::uuid,'measure-'||n||'@courseforge.local','Measure','student','$argon2id$v=19$m=65536,t=3,p=2$AAAAAAAAAAAAAAAAAAAAAA$AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA' FROM generate_series(1,20000)n; INSERT INTO courses(id,instructor_id,slug,title,summary,description,status,created_at,updated_at,published_at) SELECT ('50000000-0000-4000-8000-'||lpad(n::text,12,'0'))::uuid,'40000000-0000-4000-8000-000000000001','measure-'||n,'Measure course','Measure summary','Measure description','published',now(),now(),now() FROM generate_series(1,100)n; INSERT INTO course_modules(id,course_id,title,position) SELECT ('60000000-0000-4000-8000-'||lpad(n::text,12,'0'))::uuid,('50000000-0000-4000-8000-'||lpad(n::text,12,'0'))::uuid,'Module',0 FROM generate_series(1,100)n; INSERT INTO course_lessons(id,module_id,title,description,position) SELECT gen_random_uuid(),('60000000-0000-4000-8000-'||lpad(n::text,12,'0'))::uuid,'Lesson','Measure lesson',p FROM generate_series(1,100)n CROSS JOIN generate_series(1,3)p; INSERT INTO discussion_reviews(id,course_id,author_id,stars,body,completed_lessons,total_lessons,created_at) SELECT gen_random_uuid(),('50000000-0000-4000-8000-'||lpad((CASE WHEN n <= 2000 THEN 1 ELSE 2+floor(98*power(random(),3))::int END)::text,12,'0'))::uuid,('40000000-0000-4000-8000-'||lpad(n::text,12,'0'))::uuid,1+floor(random()*5)::int,'Measurement review body that is long enough.',0,3,now() FROM generate_series(1,20000)n`)
	if err != nil {
		t.Fatal(err)
	}
}

func explainRuns(t *testing.T, ctx context.Context, db *sql.DB, query string) ([]time.Duration, string) {
	t.Helper()
	if _, err := db.ExecContext(ctx, "EXPLAIN (ANALYZE, BUFFERS, FORMAT TEXT) "+query); err != nil {
		t.Fatal(err)
	}
	times := make([]time.Duration, 0, 5)
	plan := ""
	for range 5 {
		rows, err := db.QueryContext(ctx, "EXPLAIN (ANALYZE, BUFFERS, FORMAT TEXT) "+query)
		if err != nil {
			t.Fatal(err)
		}
		var line string
		lines := ""
		for rows.Next() {
			if err := rows.Scan(&line); err != nil {
				t.Fatal(err)
			}
			lines += line + "\n"
			if _, err := fmt.Sscanf(line, "Execution Time: %f ms", new(float64)); err == nil {
				var milliseconds float64
				_, _ = fmt.Sscanf(line, "Execution Time: %f ms", &milliseconds)
				times = append(times, time.Duration(milliseconds*float64(time.Millisecond)))
			}
		}
		if err := rows.Err(); err != nil {
			t.Fatal(err)
		}
		_ = rows.Close()
		plan = lines
	}
	return times, plan
}
func median(values []time.Duration) time.Duration {
	sort.Slice(values, func(i, j int) bool { return values[i] < values[j] })
	return values[len(values)/2]
}
