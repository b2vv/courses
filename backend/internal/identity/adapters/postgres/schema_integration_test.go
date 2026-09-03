//go:build integration

package postgres_test

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	postgrescontainer "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestIdentitySchemaAgainstPostgreSQL(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	container, err := postgrescontainer.Run(ctx, "postgres:17.5", postgrescontainer.WithDatabase("identity_test"), postgrescontainer.WithUsername("courses"), postgrescontainer.WithPassword("ephemeral-test-only"), testcontainers.WithWaitStrategy(wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(time.Minute)))
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

	insertUser := `INSERT INTO users(id,email,display_name,role,password_hash,created_at,updated_at)
		VALUES($1,$2,$3,$4,'$argon2id$v=19$m=65536,t=3,p=2$AAAAAAAAAAAAAAAAAAAAAA$AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA','2026-09-03T12:00:00Z','2026-09-03T12:00:00Z')`
	if _, err := database.ExecContext(ctx, insertUser, "40000000-0000-4000-8000-000000000001", "student@courseforge.local", "Студентка Курсу", "student"); err != nil {
		t.Fatal(err)
	}
	if _, err := database.ExecContext(ctx, `INSERT INTO user_sessions(id,user_id,token_digest,created_at,expires_at)
		VALUES('50000000-0000-4000-8000-000000000001','40000000-0000-4000-8000-000000000001',decode(repeat('ab',32),'hex'),'2026-09-03T12:00:00Z','2026-09-04T00:00:00Z')`); err != nil {
		t.Fatal(err)
	}

	t.Run("constraints reject invalid persistence records", func(t *testing.T) {
		statements := []string{
			`UPDATE users SET id='00000000-0000-0000-0000-000000000000' WHERE email='student@courseforge.local'`,
			`UPDATE users SET email=' Student@CourseForge.Local ' WHERE email='student@courseforge.local'`,
			`UPDATE users SET email='missing-at.example' WHERE email='student@courseforge.local'`,
			`UPDATE users SET display_name='  ' WHERE email='student@courseforge.local'`,
			`UPDATE users SET role='owner' WHERE email='student@courseforge.local'`,
			`UPDATE users SET password_hash='plaintext' WHERE email='student@courseforge.local'`,
			`UPDATE users SET updated_at='2026-09-03T11:59:59Z' WHERE email='student@courseforge.local'`,
			`UPDATE users SET disabled_at='2026-09-03T12:00:01Z' WHERE email='student@courseforge.local'`,
			`UPDATE user_sessions SET token_digest=decode('ab','hex')`,
			`UPDATE user_sessions SET expires_at=created_at`,
			`INSERT INTO user_sessions(id,user_id,token_digest,created_at,expires_at) VALUES('50000000-0000-4000-8000-000000000002','40000000-0000-4000-8000-000000000001',decode(repeat('ab',32),'hex'),'2026-09-03T12:00:00Z','2026-09-04T00:00:00Z')`,
		}
		for _, statement := range statements {
			if _, err := database.ExecContext(ctx, statement); err == nil {
				t.Fatalf("constraint did not reject %s", statement)
			}
		}
	})

	t.Run("user deletion cascades sessions", func(t *testing.T) {
		if _, err := database.ExecContext(ctx, `DELETE FROM users WHERE id='40000000-0000-4000-8000-000000000001'`); err != nil {
			t.Fatal(err)
		}
		var count int
		if err := database.QueryRowContext(ctx, `SELECT count(*) FROM user_sessions`).Scan(&count); err != nil {
			t.Fatal(err)
		}
		if count != 0 {
			t.Fatalf("got %d sessions after user deletion", count)
		}
	})

	t.Run("migration down and up is reversible", func(t *testing.T) {
		if _, err := provider.DownTo(ctx, 0); err != nil {
			t.Fatal(err)
		}
		for _, table := range []string{"users", "user_sessions"} {
			var exists bool
			if err := database.QueryRowContext(ctx, `SELECT to_regclass($1) IS NOT NULL`, table).Scan(&exists); err != nil {
				t.Fatal(err)
			}
			if exists {
				t.Fatalf("%s still exists after down migration", table)
			}
		}
		if _, err := provider.Up(ctx); err != nil {
			t.Fatal(err)
		}
	})
}
