//go:build integration

package postgres_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
	"time"

	"courses/internal/course/adapters/postgres"
	"courses/internal/course/domain"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	postgrescontainer "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestRepositoryAgainstPostgreSQL(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	container, err := postgrescontainer.Run(ctx, "postgres:17.5", postgrescontainer.WithDatabase("courses_test"), postgrescontainer.WithUsername("courses"), postgrescontainer.WithPassword("ephemeral-test-only"), testcontainers.WithWaitStrategy(wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(time.Minute)))
	if err != nil {
		t.Fatalf("PostgreSQL integration requires Docker: %v", err)
	}
	t.Cleanup(func() {
		if err := container.Terminate(context.Background()); err != nil {
			t.Errorf("terminate test container: %v", err)
		}
	})
	connection, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("pgx", connection)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	directory := filepath.Join("..", "..", "..", "..", "migrations")
	provider, err := goose.NewProvider(goose.DialectPostgres, db, os.DirFS(directory), goose.WithDisableGlobalRegistry(true))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := provider.Up(ctx); err != nil {
		t.Fatal(err)
	}
	pool, err := pgxpool.New(ctx, connection)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	repository := postgres.NewRepository(pool)
	query, _ := domain.NewQuery(20, 0, "", "recent")
	page, err := repository.ListPublished(ctx, query)
	if err != nil {
		t.Fatal(err)
	}
	if page.Items == nil || len(page.Items) != 0 || page.Total != 0 {
		t.Fatalf("empty page contract: %+v", page)
	}

	t.Run("authoring migration backfill failure and rollback", func(t *testing.T) {
		if _, err := provider.DownTo(ctx, 0); err != nil {
			t.Fatal(err)
		}
		if _, err := provider.UpTo(ctx, 2); err != nil {
			t.Fatal(err)
		}
		identitySeed := mustRead(t, filepath.Join("..", "..", "..", "..", "sql", "seed-identities.sql"))
		if _, err := pool.Exec(ctx, string(identitySeed)); err != nil {
			t.Fatal(err)
		}
		if _, err := pool.Exec(ctx, "UPDATE users SET disabled_at=updated_at WHERE id='40000000-0000-4000-8000-000000000003'"); err != nil {
			t.Fatal(err)
		}
		if _, err := pool.Exec(ctx, `INSERT INTO courses (id,slug,title,summary,description,status)
		SELECT ('50000000-0000-4000-8000-'||lpad(n::text,12,'0'))::uuid,'migration-'||n,'Migration fixture','Fixture summary','Fixture description','draft' FROM generate_series(1,5) AS n`); err != nil {
			t.Fatal(err)
		}
		if _, err := provider.Up(ctx); err != nil {
			t.Fatal(err)
		}
		var courseCount, ownerCount int
		var ownerID string
		var ownersEnabled bool
		if err := db.QueryRowContext(ctx, `SELECT count(*), count(DISTINCT courses.instructor_id), min(courses.instructor_id::text), bool_and(users.disabled_at IS NULL)
		FROM courses JOIN users ON users.id=courses.instructor_id`).Scan(&courseCount, &ownerCount, &ownerID, &ownersEnabled); err != nil {
			t.Fatal(err)
		}
		if courseCount != 5 || ownerCount != 1 || ownerID != "40000000-0000-4000-8000-000000000002" || !ownersEnabled {
			t.Fatalf("authoring backfill: courses=%d owners=%d owner=%s enabled=%t", courseCount, ownerCount, ownerID, ownersEnabled)
		}
		var nullable string
		if err := db.QueryRowContext(ctx, "SELECT is_nullable FROM information_schema.columns WHERE table_schema='public' AND table_name='courses' AND column_name='instructor_id'").Scan(&nullable); err != nil {
			t.Fatal(err)
		}
		if nullable != "NO" {
			t.Fatalf("instructor_id nullable: %s", nullable)
		}

		if _, err := provider.DownTo(ctx, 2); err != nil {
			t.Fatal(err)
		}
		var lessonsGone, modulesGone, ownerGone bool
		if err := db.QueryRowContext(ctx, `SELECT to_regclass('course_lessons') IS NULL, to_regclass('course_modules') IS NULL,
		NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema='public' AND table_name='courses' AND column_name='instructor_id')`).Scan(&lessonsGone, &modulesGone, &ownerGone); err != nil {
			t.Fatal(err)
		}
		if !lessonsGone || !modulesGone || !ownerGone {
			t.Fatalf("authoring down migration: lessons gone=%t modules gone=%t owner gone=%t", lessonsGone, modulesGone, ownerGone)
		}
		if _, err := repository.ListPublished(ctx, query); err != nil {
			t.Fatal(err)
		}

		if _, err := pool.Exec(ctx, "DELETE FROM courses"); err != nil {
			t.Fatal(err)
		}
		if _, err := pool.Exec(ctx, "DELETE FROM users"); err != nil {
			t.Fatal(err)
		}
		if _, err := pool.Exec(ctx, "INSERT INTO courses (id,slug,title,summary,description,status) VALUES ('50000000-0000-4000-8000-000000000999','ownerless','Ownerless','Fixture summary','Fixture description','draft')"); err != nil {
			t.Fatal(err)
		}
		if _, err := provider.Up(ctx); err == nil {
			t.Fatal("expected authoring migration to reject courses without users")
		} else {
			var pgError *pgconn.PgError
			if !errors.As(err, &pgError) || pgError.Code != "23502" {
				t.Fatalf("expected not-null violation, got %v", err)
			}
		}
		if err := db.QueryRowContext(ctx, "SELECT count(*) FROM courses").Scan(&courseCount); err != nil {
			t.Fatal(err)
		}
		if courseCount != 1 {
			t.Fatalf("failed migration changed course count: %d", courseCount)
		}
		if err := db.QueryRowContext(ctx, `SELECT to_regclass('course_lessons') IS NULL, to_regclass('course_modules') IS NULL,
		NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema='public' AND table_name='courses' AND column_name='instructor_id')`).Scan(&lessonsGone, &modulesGone, &ownerGone); err != nil {
			t.Fatal(err)
		}
		if !lessonsGone || !modulesGone || !ownerGone {
			t.Fatalf("failed migration changed schema: lessons gone=%t modules gone=%t owner gone=%t", lessonsGone, modulesGone, ownerGone)
		}

		if _, err := pool.Exec(ctx, "DELETE FROM courses"); err != nil {
			t.Fatal(err)
		}
		if _, err := pool.Exec(ctx, string(identitySeed)); err != nil {
			t.Fatal(err)
		}
		if _, err := provider.Up(ctx); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("failed migration rolls back its complete transaction", func(t *testing.T) {
		initial, err := os.ReadFile(filepath.Join(directory, "00001_courses.sql"))
		if err != nil {
			t.Fatal(err)
		}
		files := fstest.MapFS{
			"00001_courses.sql":   &fstest.MapFile{Data: initial},
			"00002_identity.sql":  &fstest.MapFile{Data: mustRead(t, filepath.Join(directory, "00002_identity.sql"))},
			"00003_authoring.sql": &fstest.MapFile{Data: mustRead(t, filepath.Join(directory, "00003_authoring.sql"))},
			"00004_failure.sql":   &fstest.MapFile{Data: []byte("-- +goose Up\nCREATE TABLE migration_canary (id integer);\nSELECT 1/0;\n-- +goose Down\nDROP TABLE migration_canary;\n")},
		}
		failure, err := goose.NewProvider(goose.DialectPostgres, db, files, goose.WithDisableGlobalRegistry(true))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := failure.Up(ctx); err == nil {
			t.Fatal("expected failed migration")
		}
		var exists bool
		if err := db.QueryRowContext(ctx, "SELECT to_regclass('migration_canary') IS NOT NULL").Scan(&exists); err != nil {
			t.Fatal(err)
		}
		if exists {
			t.Fatal("failed migration committed partial schema")
		}
		if _, err := repository.ListPublished(ctx, query); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("concurrent repeated seed and hidden courses", func(t *testing.T) {
		identitySeed, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "sql", "seed-identities.sql"))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := pool.Exec(ctx, string(identitySeed)); err != nil {
			t.Fatal(err)
		}
		seed, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "sql", "seed.sql"))
		if err != nil {
			t.Fatal(err)
		}
		results := make(chan error, 2)
		for range 2 {
			go func() { _, err := pool.Exec(ctx, string(seed)); results <- err }()
		}
		for range 2 {
			if err := <-results; err != nil {
				t.Fatal(err)
			}
		}
		page, err := repository.ListPublished(ctx, query)
		if err != nil {
			t.Fatal(err)
		}
		if page.Total != 3 || len(page.Items) != 3 || page.Items[0].Slug != "go-backend-foundations" {
			t.Fatalf("seed/visibility/order: %+v", page)
		}
		if _, err := pool.Exec(ctx, `INSERT INTO course_modules (id,course_id,title,position) VALUES
		('60000000-0000-4000-8000-000000000002','10000000-0000-4000-8000-000000000001','PostgreSQL',1),
		('60000000-0000-4000-8000-000000000001','10000000-0000-4000-8000-000000000001','HTTP API',0),
		('60000000-0000-4000-8000-000000000003','10000000-0000-4000-8000-000000000004','Draft secret',0),
		('60000000-0000-4000-8000-000000000004','10000000-0000-4000-8000-000000000005','Archived secret',0);
		INSERT INTO course_lessons (id,module_id,title,description,position) VALUES
		('70000000-0000-4000-8000-000000000002','60000000-0000-4000-8000-000000000001','Contracts','Must stay private',1),
		('70000000-0000-4000-8000-000000000001','60000000-0000-4000-8000-000000000001','Handlers','Must stay private',0),
		('70000000-0000-4000-8000-000000000003','60000000-0000-4000-8000-000000000002','Transactions','Must stay private',0),
		('70000000-0000-4000-8000-000000000004','60000000-0000-4000-8000-000000000003','Draft lesson','Must stay private',0),
		('70000000-0000-4000-8000-000000000005','60000000-0000-4000-8000-000000000004','Archived lesson','Must stay private',0)`); err != nil {
			t.Fatal(err)
		}
		for _, slug := range []string{"draft-course", "archived-course", "missing-course"} {
			if _, err := repository.GetPublished(ctx, slug); !errors.Is(err, domain.ErrCourseNotFound) {
				t.Fatalf("%s: %v", slug, err)
			}
		}
		course, err := repository.GetPublished(ctx, "go-backend-foundations")
		if err != nil {
			t.Fatal(err)
		}
		if course.Title != "Go: фундамент надійного бекенду" || course.Status != domain.Published || course.PublishedAt == nil {
			t.Fatalf("wrong mapped course: %+v", course)
		}
		if len(course.Program) != 2 || course.Program[0].Title != "HTTP API" || course.Program[1].Title != "PostgreSQL" ||
			len(course.Program[0].Lessons) != 2 || course.Program[0].Lessons[0].Title != "Handlers" ||
			course.Program[0].Lessons[1].Title != "Contracts" || len(course.Program[1].Lessons) != 1 ||
			course.Program[1].Lessons[0].Title != "Transactions" {
			t.Fatalf("wrong public program order: %+v", course.Program)
		}
	})
	t.Run("database constraints reject invalid entities", func(t *testing.T) {
		for _, statement := range []string{
			"INSERT INTO courses (id,slug,title,summary,description,status) VALUES ('90000000-0000-4000-8000-000000000001','missing-owner','Missing owner','Fixture summary','Fixture description','draft')",
			"UPDATE courses SET slug='invalid--slug' WHERE slug='draft-course'",
			"UPDATE courses SET status='unknown' WHERE slug='draft-course'",
			"UPDATE courses SET status='published' WHERE slug='draft-course'",
			"UPDATE courses SET title='  ' WHERE slug='draft-course'",
			"UPDATE courses SET slug='go-backend-foundations' WHERE slug='draft-course'",
			"UPDATE courses SET updated_at='2020-01-01T00:00:00Z' WHERE slug='draft-course'",
		} {
			if _, err := pool.Exec(ctx, statement); err == nil {
				t.Fatalf("constraint did not reject %s", statement)
			}
		}
	})
	t.Run("one max beyond end and stable ties", func(t *testing.T) {
		_, err := pool.Exec(ctx, `INSERT INTO courses (id,instructor_id,slug,title,summary,description,status,created_at,updated_at,published_at)
		SELECT ('20000000-0000-4000-8000-'||lpad(n::text,12,'0'))::uuid,'40000000-0000-4000-8000-000000000002','bulk-'||n,'Tied title','Bulk fixture','Fixture description','published','2026-09-01T00:00:00Z','2026-09-01T00:00:00Z','2026-09-01T00:00:00Z' FROM generate_series(1,51) AS n`)
		if err != nil {
			t.Fatal(err)
		}
		for _, sort := range []string{"recent", "title"} {
			for _, test := range []struct{ limit, offset, want int }{{1, 0, 1}, {50, 0, 50}, {50, 50, 1}, {20, 10000, 0}} {
				q, _ := domain.NewQuery(test.limit, test.offset, "Tied title", sort)
				page, err := repository.ListPublished(ctx, q)
				if err != nil {
					t.Fatal(err)
				}
				if len(page.Items) != test.want || page.Total != 51 {
					t.Fatalf("%s limit=%d offset=%d: %+v", sort, test.limit, test.offset, page)
				}
				for i, course := range page.Items {
					if course.Slug != fmt.Sprintf("bulk-%d", test.offset+i+1) {
						t.Fatalf("unstable tie: %s", course.Slug)
					}
				}
			}
		}
	})
	t.Run("literal wildcard search", func(t *testing.T) {
		for index, title := range []string{"100% reliable", "snake_case", `path\to!`} {
			_, err := pool.Exec(ctx, `INSERT INTO courses(id,instructor_id,slug,title,summary,description,status,created_at,updated_at,published_at) VALUES($1,'40000000-0000-4000-8000-000000000002',$2,$3,'Literal fixture','Fixture description','published','2026-09-01T00:00:00Z','2026-09-01T00:00:00Z','2026-09-01T00:00:00Z')`, fmt.Sprintf("30000000-0000-4000-8000-%012d", index+1), fmt.Sprintf("literal-%d", index), title)
			if err != nil {
				t.Fatal(err)
			}
		}
		for _, search := range []string{"%", "_", "!", `\`, "RELIABLE"} {
			q, _ := domain.NewQuery(20, 0, search, "recent")
			page, err := repository.ListPublished(ctx, q)
			if err != nil {
				t.Fatal(err)
			}
			if page.Total != 1 || len(page.Items) != 1 {
				t.Fatalf("%q matched nonliteral rows: %+v", search, page)
			}
		}
	})
	t.Run("query cancellation releases the connection", func(t *testing.T) {
		lock, err := pool.Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		defer lock.Rollback(context.Background())
		if _, err := lock.Exec(ctx, "LOCK TABLE courses IN ACCESS EXCLUSIVE MODE"); err != nil {
			t.Fatal(err)
		}
		request, cancel := context.WithTimeout(ctx, 50*time.Millisecond)
		defer cancel()
		started := time.Now()
		if _, err := repository.ListPublished(request, query); !errors.Is(err, domain.ErrDependencyUnavailable) {
			t.Fatalf("got %v", err)
		}
		if time.Since(started) > time.Second {
			t.Fatal("query ignored cancellation")
		}
		if err := lock.Rollback(ctx); err != nil {
			t.Fatal(err)
		}
		if _, err := repository.ListPublished(ctx, query); err != nil {
			t.Fatalf("connection failed after cancellation: %v", err)
		}
	})
}

func mustRead(t *testing.T, path string) []byte {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return content
}
