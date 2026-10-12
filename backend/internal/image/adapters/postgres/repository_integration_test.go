//go:build integration

package postgres_test

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	imagepostgres "courses/internal/image/adapters/postgres"
	"courses/internal/image/domain"
	containers "courses/internal/testsupport/containers"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	postgrescontainer "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestImageRepositoryAgainstPostgreSQL(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	container, err := containers.Start(ctx, func(ctx context.Context) (*postgrescontainer.PostgresContainer, error) {
		return postgrescontainer.Run(ctx, "postgres:17.5",
			postgrescontainer.WithDatabase("image_test"),
			postgrescontainer.WithUsername("courses"),
			postgrescontainer.WithPassword("ephemeral-test-only"),
			testcontainers.WithWaitStrategy(wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(3*time.Minute)),
		)
	})
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
	database, err := sql.Open("pgx", connection)
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	migrations := filepath.Join("..", "..", "..", "..", "migrations")
	provider, err := goose.NewProvider(goose.DialectPostgres, database, os.DirFS(migrations), goose.WithDisableGlobalRegistry(true))
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
	repository := imagepostgres.NewRepository(pool)
	// This suite owns an isolated database; container termination removes all fixtures.
	var ownerID, courseID, moduleID, lessonID string
	if err := pool.QueryRow(ctx, `INSERT INTO users(id,email,display_name,role,password_hash)
VALUES(gen_random_uuid(),'instructor@image.test','Image user','instructor','$argon2id$v=19$m=65536,t=3,p=2$AAAAAAAAAAAAAAAAAAAAAA$AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA') RETURNING id::text`).Scan(&ownerID); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx, `INSERT INTO courses(id,instructor_id,slug,title,summary,description,status,published_at,price_minor,currency)
VALUES(gen_random_uuid(),$1,'image-course','Image course','Summary','Description','published',now(),100,'UAH') RETURNING id::text`, ownerID).Scan(&courseID); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx, `INSERT INTO course_modules(id,course_id,title,position) VALUES(gen_random_uuid(),$1,'Module',0) RETURNING id::text`, courseID).Scan(&moduleID); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx, `INSERT INTO course_lessons(id,module_id,title,position) VALUES(gen_random_uuid(),$1,'Lesson',0) RETURNING id::text`, moduleID).Scan(&lessonID); err != nil {
		t.Fatal(err)
	}
	image, err := domain.NewImage("12345678-1234-4234-8234-123456789abc", lessonID, domain.ContentTypeJPEG, ownerID, 1920, 1080, 2048, time.Now().UTC().Truncate(time.Microsecond))
	if err != nil {
		t.Fatal(err)
	}
	if err := repository.Insert(ctx, image); err != nil {
		t.Fatal(err)
	}
	t.Run("round trip", func(t *testing.T) {
		got, err := repository.FindByID(ctx, image.ID)
		if err != nil {
			t.Fatal(err)
		}
		if got.ID != image.ID || got.LessonID != image.LessonID || got.ContentType != image.ContentType || got.Width != image.Width || got.Height != image.Height || got.ByteSize != image.ByteSize || got.CreatedBy != image.CreatedBy || !got.CreatedAt.UTC().Equal(image.CreatedAt.UTC()) {
			t.Fatalf("FindByID = %+v; want %+v", got, image)
		}
	})
	t.Run("missing image", func(t *testing.T) {
		if _, err := repository.FindByID(ctx, "ffffffff-ffff-4fff-8fff-ffffffffffff"); !errors.Is(err, domain.ErrImageNotFound) {
			t.Fatalf("missing image: %v", err)
		}
	})
	t.Run("duplicate image", func(t *testing.T) {
		err := repository.Insert(ctx, image)
		var postgresError *pgconn.PgError
		if !errors.As(err, &postgresError) || postgresError.Code != "23505" {
			t.Fatalf("duplicate image: %v; want unique violation", err)
		}
	})
	t.Run("missing lesson", func(t *testing.T) {
		missing := image
		missing.ID = "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"
		missing.LessonID = "ffffffff-ffff-4fff-8fff-ffffffffffff"
		err := repository.Insert(ctx, missing)
		var postgresError *pgconn.PgError
		if !errors.As(err, &postgresError) || postgresError.Code != "23503" {
			t.Fatalf("missing lesson: %v; want foreign key violation", err)
		}
	})
	t.Run("nullable creator", func(t *testing.T) {
		if _, err := pool.Exec(ctx, `UPDATE lesson_images SET created_by = NULL WHERE id = $1::text::uuid`, image.ID); err != nil {
			t.Fatal(err)
		}
		got, err := repository.FindByID(ctx, image.ID)
		if err != nil || got.CreatedBy != "" {
			t.Fatalf("nullable creator = %q, %v; want empty string, nil", got.CreatedBy, err)
		}
	})
}
