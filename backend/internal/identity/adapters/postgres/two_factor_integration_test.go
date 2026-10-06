//go:build integration

package postgres

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	passwords "courses/internal/identity/adapters/argon2id"
	containers "courses/internal/testsupport/containers"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	postgrescontainer "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

const twoFactorUserID = "44000000-0000-4000-8000-000000000001"

func TestTwoFactorRepositoryAgainstPostgreSQL(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	container, err := containers.Start(ctx, func(ctx context.Context) (*postgrescontainer.PostgresContainer, error) {
		return postgrescontainer.Run(ctx, "postgres:17.5",
			postgrescontainer.WithDatabase("two_factor_test"),
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

	repository := NewRepository(pool, passwords.NewVerifier())
	if _, err := pool.Exec(ctx, `INSERT INTO users(id,email,display_name,role,password_hash,created_at,updated_at)
		VALUES($1, 'two-factor@example.test', 'Two Factor User', 'student', '$argon2id$v=19$m=65536,t=3,p=2$c3R1ZGVudC1kZW1vLXNsdA$5S1MhXX3wgDMnIKuygFRfUNVvTtRktPECxV3ZUBg0is', now(), now())`, twoFactorUserID); err != nil {
		t.Fatal(err)
	}

	t.Run("TOTP state follows enrollment lifecycle", func(t *testing.T) {
		state, err := repository.GetTwoFactor(ctx, twoFactorUserID)
		if err != nil || state.Secret != "" || state.Enabled {
			t.Fatalf("default state=%+v err=%v", state, err)
		}
		if err := repository.SetTotpSecret(ctx, twoFactorUserID, "JBSWY3DPEHPK3PXP"); err != nil {
			t.Fatal(err)
		}
		state, err = repository.GetTwoFactor(ctx, twoFactorUserID)
		if err != nil || state.Secret != "JBSWY3DPEHPK3PXP" || state.Enabled {
			t.Fatalf("pending state=%+v err=%v", state, err)
		}
		if err := repository.EnableTotp(ctx, twoFactorUserID); err != nil {
			t.Fatal(err)
		}
		state, err = repository.GetTwoFactor(ctx, twoFactorUserID)
		if err != nil || !state.Enabled {
			t.Fatalf("enabled state=%+v err=%v", state, err)
		}
		if err := repository.DisableTotp(ctx, twoFactorUserID); err != nil {
			t.Fatal(err)
		}
		state, err = repository.GetTwoFactor(ctx, twoFactorUserID)
		if err != nil || state.Secret != "" || state.Enabled {
			t.Fatalf("disabled state=%+v err=%v", state, err)
		}
	})

	t.Run("recovery codes are consumed once and can be deleted", func(t *testing.T) {
		digests := make([][32]byte, 10)
		for index := range digests {
			digests[index] = sha256.Sum256([]byte{byte(index)})
		}
		if err := repository.InsertCodes(ctx, twoFactorUserID, digests); err != nil {
			t.Fatal(err)
		}
		consumed, err := repository.ConsumeCode(ctx, twoFactorUserID, digests[0])
		if err != nil || !consumed {
			t.Fatalf("first consume consumed=%t err=%v", consumed, err)
		}
		consumed, err = repository.ConsumeCode(ctx, twoFactorUserID, digests[0])
		if err != nil || consumed {
			t.Fatalf("replayed consume consumed=%t err=%v", consumed, err)
		}
		unknown := sha256.Sum256([]byte("unknown"))
		consumed, err = repository.ConsumeCode(ctx, twoFactorUserID, unknown)
		if err != nil || consumed {
			t.Fatalf("unknown consume consumed=%t err=%v", consumed, err)
		}
		if err := repository.DeleteCodes(ctx, twoFactorUserID); err != nil {
			t.Fatal(err)
		}
		var remaining int
		if err := pool.QueryRow(ctx, `SELECT count(*) FROM recovery_codes WHERE user_id=$1`, twoFactorUserID).Scan(&remaining); err != nil {
			t.Fatal(err)
		}
		if remaining != 0 {
			t.Fatalf("remaining recovery codes=%d", remaining)
		}
	})
}
