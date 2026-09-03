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

	t.Run("migration up down up", func(t *testing.T) {
		if _, err := provider.DownTo(ctx, 0); err != nil {
			t.Fatal(err)
		}
		if _, err := provider.Up(ctx); err != nil {
			t.Fatal(err)
		}
		if _, err := repository.ListPublished(ctx, query); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("failed migration rolls back its complete transaction", func(t *testing.T) {
		initial, err := os.ReadFile(filepath.Join(directory, "00001_courses.sql"))
		if err != nil {
			t.Fatal(err)
		}
		files := fstest.MapFS{
			"00001_courses.sql": &fstest.MapFile{Data: initial},
			"00002_failure.sql": &fstest.MapFile{Data: []byte("-- +goose Up\nCREATE TABLE migration_canary (id integer);\nSELECT 1/0;\n-- +goose Down\nDROP TABLE migration_canary;\n")},
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
	})
	t.Run("database constraints reject invalid entities", func(t *testing.T) {
		for _, statement := range []string{
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
		_, err := pool.Exec(ctx, `INSERT INTO courses (id,slug,title,summary,description,status,created_at,updated_at,published_at)
		SELECT ('20000000-0000-4000-8000-'||lpad(n::text,12,'0'))::uuid,'bulk-'||n,'Tied title','Bulk fixture','Fixture description','published','2026-09-01T00:00:00Z','2026-09-01T00:00:00Z','2026-09-01T00:00:00Z' FROM generate_series(1,51) AS n`)
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
			_, err := pool.Exec(ctx, `INSERT INTO courses(id,slug,title,summary,description,status,created_at,updated_at,published_at) VALUES($1,$2,$3,'Literal fixture','Fixture description','published','2026-09-01T00:00:00Z','2026-09-01T00:00:00Z','2026-09-01T00:00:00Z')`, fmt.Sprintf("30000000-0000-4000-8000-%012d", index+1), fmt.Sprintf("literal-%d", index), title)
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
