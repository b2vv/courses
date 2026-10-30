//go:build integration

package postgres

import (
	"bytes"
	"context"
	cryptorand "crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	passwords "courses/internal/identity/adapters/argon2id"
	identityhttp "courses/internal/identity/adapters/http"
	"courses/internal/identity/application"
	"courses/internal/identity/domain"
	containers "courses/internal/testsupport/containers"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	postgrescontainer "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"golang.org/x/crypto/argon2"
)

func TestIdentitySchemaAgainstPostgreSQL(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	container, err := containers.Start(ctx, func(ctx context.Context) (*postgrescontainer.PostgresContainer, error) {
		return postgrescontainer.Run(ctx, "postgres:17.5", postgrescontainer.WithDatabase("identity_test"), postgrescontainer.WithUsername("courses"), postgrescontainer.WithPassword("ephemeral-test-only"), testcontainers.WithWaitStrategy(wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(3*time.Minute)))
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

	t.Run("identity seed is explicit, authoritative, repeatable, and concurrent", func(t *testing.T) {
		backendRoot := filepath.Join("..", "..", "..", "..")
		seed, err := os.ReadFile(filepath.Join(backendRoot, "sql", "seed-identities.sql"))
		if err != nil {
			t.Fatal(err)
		}
		courseSeed, err := os.ReadFile(filepath.Join(backendRoot, "sql", "seed.sql"))
		if err != nil {
			t.Fatal(err)
		}
		normalizedCourseSeed := strings.ToLower(strings.Join(strings.Fields(string(courseSeed)), " "))
		if strings.Contains(normalizedCourseSeed, "insert into users") {
			t.Fatal("the course seed must not insert identity fixtures")
		}
		if !strings.Contains(strings.ToUpper(string(seed)), "LOCK TABLE USERS IN SHARE ROW EXCLUSIVE MODE") {
			t.Fatal("the identity seed must serialize concurrent fixture reconciliation")
		}

		repository := NewRepository(pool, passwords.NewVerifier(), testKeys())
		config, err := identityhttp.NewConfig("test", "http://localhost:8080", false)
		if err != nil {
			t.Fatal(err)
		}
		identity := application.NewIdentity(repository, repository, func() time.Time {
			return time.Date(2026, 9, 5, 12, 0, 0, 0, time.UTC)
		})
		mux := http.NewServeMux()
		identityhttp.NewHandler(identity, repository, nil, config).Register(mux)

		type fixture struct {
			id, email, displayName, password string
			role                             domain.Role
		}
		fixtures := []fixture{
			{"40000000-0000-4000-8000-000000000001", "student@courseforge.local", "Студентка Курсу", "courseforge-student-demo", domain.Student},
			{"40000000-0000-4000-8000-000000000002", "instructor@courseforge.local", "Викладач Курсу", "courseforge-instructor-demo", domain.Instructor},
			{"40000000-0000-4000-8000-000000000003", "admin@courseforge.local", "Адміністратор Курсу", "courseforge-admin-demo", domain.Admin},
			{"40000000-0000-4000-8000-0000000000d1", "block-me@courseforge.local", "Тест блокування", "courseforge-student-demo", domain.Student},
			{"40000000-0000-4000-8000-0000000000d2", "delete-me@courseforge.local", "Тест видалення", "courseforge-student-demo", domain.Student},
			{"40000000-0000-4000-8000-0000000000d3", "reset-me@courseforge.local", "Тест скидання пароля", "courseforge-student-demo", domain.Student},
			{"40000000-0000-4000-8000-0000000000d4", "outsider@courseforge.local", "Сторонній Глядач", "courseforge-student-demo", domain.Student},
		}
		defer func() {
			if _, err := pool.Exec(context.Background(), `DELETE FROM users`); err != nil {
				t.Errorf("clean identity seed: %v", err)
			}
		}()

		login := func(email, password string) *httptest.ResponseRecorder {
			body := fmt.Sprintf(`{"email":%q,"password":%q}`, email, password)
			request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(body))
			request.Header.Set("Content-Type", "application/json")
			request.Header.Set("Origin", "http://localhost:8080")
			response := httptest.NewRecorder()
			mux.ServeHTTP(response, request)
			return response
		}
		snapshot := func() string {
			var value string
			err := pool.QueryRow(ctx, `SELECT string_agg(id::text || '|' || email || '|' || display_name || '|' || role || '|' || password_hash || '|' || coalesce(disabled_at::text, '') || '|' || coalesce(deleted_at::text, ''), E'\n' ORDER BY id) FROM users`).Scan(&value)
			if err != nil {
				t.Fatal(err)
			}
			return value
		}
		assertFixtureSet := func() {
			var count int
			if err := pool.QueryRow(ctx, `SELECT count(*) FROM users`).Scan(&count); err != nil || count != len(fixtures) {
				t.Fatalf("fixture count=%d err=%v", count, err)
			}
			for _, expected := range fixtures {
				var email, displayName string
				var role domain.Role
				var disabled, deleted bool
				err := pool.QueryRow(ctx, `SELECT email, display_name, role, disabled_at IS NOT NULL, deleted_at IS NOT NULL FROM users WHERE id=$1`, expected.id).Scan(&email, &displayName, &role, &disabled, &deleted)
				if err != nil || email != expected.email || displayName != expected.displayName || role != expected.role || disabled || deleted {
					t.Fatalf("fixture %s: email=%q displayName=%q role=%q disabled=%v deleted=%v err=%v", expected.id, email, displayName, role, disabled, deleted, err)
				}
			}
		}

		if _, err := pool.Exec(ctx, string(seed)); err != nil {
			t.Fatal(err)
		}
		assertFixtureSet()
		for _, expected := range fixtures[:3] {
			response := login(expected.email, expected.password)
			body := response.Body.String()
			if response.Code != http.StatusOK || len(response.Result().Cookies()) == 0 || !strings.Contains(body, `"role":"`+string(expected.role)+`"`) || strings.Contains(body, "password") || strings.Contains(body, "token") {
				t.Fatalf("login %s: status=%d cookies=%d body=%s", expected.email, response.Code, len(response.Result().Cookies()), body)
			}
		}

		first := snapshot()
		if _, err := pool.Exec(ctx, string(seed)); err != nil {
			t.Fatal(err)
		}
		if second := snapshot(); second != first {
			t.Fatalf("repeat seed changed fixture fields:\nfirst:\n%s\nsecond:\n%s", first, second)
		}

		var studentHash string
		if err := pool.QueryRow(ctx, `SELECT password_hash FROM users WHERE id=$1`, fixtures[0].id).Scan(&studentHash); err != nil {
			t.Fatal(err)
		}
		updatedPassword := "courseforge-student-updated"
		updatedHash := encodePassword(updatedPassword, []byte("student-new-salt"))
		updatedPasswordSeed := strings.Replace(string(seed), studentHash, updatedHash, 1)
		if updatedPasswordSeed == string(seed) {
			t.Fatal("student password hash was not found in the seed")
		}
		if _, err := pool.Exec(ctx, updatedPasswordSeed); err != nil {
			t.Fatal(err)
		}
		if response := login(fixtures[0].email, fixtures[0].password); response.Code != http.StatusUnauthorized {
			t.Fatalf("old password status=%d body=%s", response.Code, response.Body.String())
		}
		if response := login(fixtures[0].email, updatedPassword); response.Code != http.StatusOK {
			t.Fatalf("updated password status=%d body=%s", response.Code, response.Body.String())
		}

		changedEmail := "student-updated@courseforge.local"
		changedAddressSeed := strings.Replace(string(seed), fixtures[0].email, changedEmail, 1)
		if changedAddressSeed == string(seed) {
			t.Fatal("student email was not found in the seed")
		}
		if _, err := pool.Exec(ctx, changedAddressSeed); err != nil {
			t.Fatal(err)
		}
		var storedEmail string
		if err := pool.QueryRow(ctx, `SELECT email FROM users WHERE id=$1`, fixtures[0].id).Scan(&storedEmail); err != nil || storedEmail != changedEmail {
			t.Fatalf("changed email=%q err=%v", storedEmail, err)
		}
		var count int
		if err := pool.QueryRow(ctx, `SELECT count(*) FROM users`).Scan(&count); err != nil || count != len(fixtures) {
			t.Fatalf("changed-address fixture count=%d err=%v", count, err)
		}

		if _, err := pool.Exec(ctx, string(seed)); err != nil {
			t.Fatal(err)
		}
		if _, err := pool.Exec(ctx, `UPDATE users SET updated_at=now(), disabled_at=now(), deleted_at=now() WHERE id=$1`, fixtures[0].id); err != nil {
			t.Fatal(err)
		}
		if _, err := pool.Exec(ctx, string(seed)); err != nil {
			t.Fatal(err)
		}
		assertFixtureSet()
		if response := login(fixtures[0].email, fixtures[0].password); response.Code != http.StatusOK {
			t.Fatalf("re-enabled fixture status=%d body=%s", response.Code, response.Body.String())
		}

		if _, err := pool.Exec(ctx, `DELETE FROM users`); err != nil {
			t.Fatal(err)
		}
		ready := make(chan struct{}, 2)
		start := make(chan struct{})
		errors := make(chan error, 2)
		var group sync.WaitGroup
		for range 2 {
			group.Add(1)
			go func() {
				defer group.Done()
				ready <- struct{}{}
				<-start
				_, err := pool.Exec(ctx, string(seed))
				errors <- err
			}()
		}
		for range 2 {
			<-ready
		}
		close(start)
		group.Wait()
		close(errors)
		for err := range errors {
			if err != nil {
				t.Error(err)
			}
		}
		assertFixtureSet()
	})

	insertUser := `INSERT INTO users(id,email,display_name,role,password_hash,created_at,updated_at)
		VALUES($1,$2,$3,$4,'$argon2id$v=19$m=65536,t=3,p=2$AAAAAAAAAAAAAAAAAAAAAA$AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA','2026-09-03T12:00:00Z','2026-09-03T12:00:00Z')`
	if _, err := database.ExecContext(ctx, insertUser, "40000000-0000-4000-8000-000000000001", "student@courseforge.local", "Студентка Курсу", "student"); err != nil {
		t.Fatal(err)
	}
	if _, err := database.ExecContext(ctx, insertUser, "40000000-0000-4000-8000-000000000002", "instructor@courseforge.local", "Викладач Курсу", "instructor"); err != nil {
		t.Fatal(err)
	}
	collisionRaw := bytes.Repeat([]byte{0x11}, 32)
	collisionToken := base64.RawURLEncoding.EncodeToString(collisionRaw)
	collisionDigest := sha256.Sum256(collisionRaw)
	previousRaw := bytes.Repeat([]byte{0x22}, 32)
	previousToken := base64.RawURLEncoding.EncodeToString(previousRaw)
	previousDigest := sha256.Sum256(previousRaw)
	if _, err := database.ExecContext(ctx, `INSERT INTO user_sessions(id,user_id,token_digest,created_at,expires_at) VALUES
		('50000000-0000-4000-8000-000000000001','40000000-0000-4000-8000-000000000001',$1,'2026-09-03T12:00:00Z','2026-09-04T00:00:00Z'),
		('50000000-0000-4000-8000-000000000002','40000000-0000-4000-8000-000000000001',$2,'2026-09-03T12:00:00Z','2026-09-04T00:00:00Z')`, collisionDigest[:], previousDigest[:]); err != nil {
		t.Fatal(err)
	}
	t.Run("identity administration schema has soft-delete constraints and partial email uniqueness", func(t *testing.T) {
		var column, validConstraint, impliesDisabled, activeEmailIndex, oldEmailConstraint bool
		err := database.QueryRowContext(ctx, `SELECT
		EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_schema='public' AND table_name='users' AND column_name='deleted_at'
				AND data_type='timestamp with time zone' AND is_nullable='YES'
		),
		EXISTS (
			SELECT 1 FROM pg_constraint
			WHERE conrelid='users'::regclass AND conname='users_deleted_at_valid' AND contype='c'
		),
		EXISTS (
			SELECT 1 FROM pg_constraint
			WHERE conrelid='users'::regclass AND conname='users_deleted_implies_disabled' AND contype='c'
		),
		EXISTS (
			SELECT 1
			FROM pg_index
			WHERE indexrelid='users_email_active_key'::regclass
				AND indisunique
				AND pg_get_expr(indpred, indrelid) LIKE '%deleted_at IS NULL%'
		),
		EXISTS (
			SELECT 1 FROM pg_constraint
			WHERE conrelid='users'::regclass AND conname='users_email_key'
		)`).Scan(&column, &validConstraint, &impliesDisabled, &activeEmailIndex, &oldEmailConstraint)
		if err != nil {
			t.Fatal(err)
		}
		if !column || !validConstraint || !impliesDisabled || !activeEmailIndex || oldEmailConstraint {
			t.Fatalf("identity administration schema: column=%t valid=%t implies-disabled=%t active-index=%t old-constraint=%t", column, validConstraint, impliesDisabled, activeEmailIndex, oldEmailConstraint)
		}
	})
	t.Run("credential verification hides unknown and disabled users", func(t *testing.T) {
		passwords := &passwordVerifier{matched: true}
		repository := NewRepository(pool, passwords, testKeys())
		user, err := repository.VerifyCredentials(ctx, "student@courseforge.local", "demo-password")
		if err != nil || user.Email != "student@courseforge.local" || passwords.verifyCalls != 1 || passwords.dummyCalls != 0 {
			t.Fatalf("user=%+v err=%v verifier=%+v", user, err, passwords)
		}
		if _, err := repository.VerifyCredentials(ctx, "missing@courseforge.local", "demo-password"); !errors.Is(err, domain.ErrInvalidCredentials) || passwords.dummyCalls != 1 {
			t.Fatalf("unknown err=%v verifier=%+v", err, passwords)
		}
		passwords.matched = false
		if _, err := repository.VerifyCredentials(ctx, "student@courseforge.local", "wrong"); !errors.Is(err, domain.ErrInvalidCredentials) {
			t.Fatalf("wrong password: %v", err)
		}
		passwords.matched = true
		if _, err := pool.Exec(ctx, `UPDATE users SET updated_at='2026-09-03T12:00:01Z', disabled_at='2026-09-03T12:00:01Z' WHERE email='student@courseforge.local'`); err != nil {
			t.Fatal(err)
		}
		before := passwords.verifyCalls
		if _, err := repository.VerifyCredentials(ctx, "student@courseforge.local", "demo-password"); !errors.Is(err, domain.ErrInvalidCredentials) || passwords.verifyCalls != before+1 {
			t.Fatalf("disabled err=%v verifier=%+v", err, passwords)
		}
		if _, err := pool.Exec(ctx, `UPDATE users SET updated_at='2026-09-03T12:00:00Z', disabled_at=NULL WHERE email='student@courseforge.local'`); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("PostgreSQL credential adapter uses the real Argon2id verifier", func(t *testing.T) {
		hash := encodePassword("courseforge-student-demo", []byte("student-seed-slt"))
		if _, err := pool.Exec(ctx, `UPDATE users SET password_hash=$1 WHERE email='student@courseforge.local'`, hash); err != nil {
			t.Fatal(err)
		}
		repository := NewRepository(pool, passwords.NewVerifier(), testKeys())
		if user, err := repository.VerifyCredentials(ctx, "student@courseforge.local", "courseforge-student-demo"); err != nil || user.Role != domain.Student {
			t.Fatalf("user=%+v err=%v", user, err)
		}
		for _, email := range []string{"student@courseforge.local", "unknown@courseforge.local"} {
			if _, err := repository.VerifyCredentials(ctx, email, "wrong-password"); !errors.Is(err, domain.ErrInvalidCredentials) {
				t.Fatalf("%s: %v", email, err)
			}
		}
	})

	t.Run("login selects the active user after email reuse", func(t *testing.T) {
		const (
			deletedID = "40000000-0000-4000-8000-000000000090"
			activeID  = "40000000-0000-4000-8000-000000000091"
			email     = "reused@courseforge.local"
			password  = "courseforge-reused-active"
		)
		hash := encodePassword(password, []byte("reused-user-salt"))
		if _, err := pool.Exec(ctx, `INSERT INTO users(id,email,display_name,role,password_hash,created_at,updated_at,disabled_at,deleted_at) VALUES
		($1,$3,'Deleted reuse fixture','student',$4,'2026-09-03T12:00:00Z','2026-09-03T12:00:00Z','2026-09-03T12:00:00Z','2026-09-03T12:00:00Z'),
		($2,$3,'Active reuse fixture','student',$4,'2026-09-03T12:00:00Z','2026-09-03T12:00:00Z',NULL,NULL)`, deletedID, activeID, email, hash); err != nil {
			t.Fatal(err)
		}
		defer func() {
			if _, err := pool.Exec(context.Background(), `DELETE FROM users WHERE id IN ($1,$2)`, deletedID, activeID); err != nil {
				t.Errorf("clean email-reuse pair before migration down: %v", err)
			}
		}()

		repository := NewRepository(pool, passwords.NewVerifier(), testKeys())
		identity := application.NewIdentity(repository, repository, func() time.Time {
			return time.Date(2026, 9, 3, 13, 0, 0, 0, time.UTC)
		})
		config, err := identityhttp.NewConfig("test", "http://localhost:8080", false)
		if err != nil {
			t.Fatal(err)
		}
		mux := http.NewServeMux()
		identityhttp.NewHandler(identity, repository, nil, config).Register(mux)
		request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(fmt.Sprintf(`{"email":%q,"password":%q}`, email, password)))
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Origin", "http://localhost:8080")
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, request)
		if response.Code != http.StatusOK || len(response.Result().Cookies()) == 0 || !strings.Contains(response.Body.String(), activeID) {
			t.Fatalf("reused-email login: status=%d cookies=%d body=%s", response.Code, len(response.Result().Cookies()), response.Body.String())
		}
	})

	t.Run("rotation is atomic and retries a digest collision", func(t *testing.T) {
		freshRaw := bytes.Repeat([]byte{0x33}, 32)
		freshToken := base64.RawURLEncoding.EncodeToString(freshRaw)
		random := bytes.NewReader(bytes.Join([][]byte{collisionRaw, bytes.Repeat([]byte{0x44}, 16), freshRaw, bytes.Repeat([]byte{0x55}, 16)}, nil))
		repository := newRepository(pool, &passwordVerifier{matched: true}, random, testKeys())
		now := time.Date(2026, 9, 3, 13, 0, 0, 0, time.UTC)
		issued, err := repository.Rotate(ctx, "40000000-0000-4000-8000-000000000001", previousToken, now, now.Add(12*time.Hour))
		if err != nil {
			t.Fatal(err)
		}
		if issued.Token != freshToken || issued.Session.UserID != "40000000-0000-4000-8000-000000000001" {
			t.Fatalf("issued=%+v", issued)
		}
		if _, err := repository.Resolve(ctx, previousToken, now); !errors.Is(err, domain.ErrAuthenticationRequired) {
			t.Fatalf("previous token: %v", err)
		}
		if _, err := repository.Resolve(ctx, collisionToken, now); err != nil {
			t.Fatalf("collision row was lost: %v", err)
		}
		if user, err := repository.Resolve(ctx, freshToken, now); err != nil || user.Email != "student@courseforge.local" {
			t.Fatalf("fresh token user=%+v err=%v", user, err)
		}
		freshDigest := sha256.Sum256(freshRaw)
		var storedDigest []byte
		if err := pool.QueryRow(ctx, `SELECT token_digest FROM user_sessions WHERE id=$1::text::uuid`, issued.Session.ID).Scan(&storedDigest); err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(storedDigest, freshDigest[:]) || bytes.Equal(storedDigest, []byte(freshToken)) {
			t.Fatal("persistence did not store only the SHA-256 token digest")
		}
		if err := repository.Revoke(ctx, freshToken); err != nil {
			t.Fatal(err)
		}
		if err := repository.Revoke(ctx, freshToken); err != nil {
			t.Fatalf("repeat revoke: %v", err)
		}
	})

	t.Run("collision exhaustion retains the presented session", func(t *testing.T) {
		presentedRaw := bytes.Repeat([]byte{0x77}, 32)
		presentedToken := base64.RawURLEncoding.EncodeToString(presentedRaw)
		presentedDigest := sha256.Sum256(presentedRaw)
		now := time.Date(2026, 9, 3, 13, 30, 0, 0, time.UTC)
		if _, err := pool.Exec(ctx, `INSERT INTO user_sessions(id,user_id,token_digest,created_at,expires_at) VALUES('50000000-0000-4000-8000-000000000007','40000000-0000-4000-8000-000000000001',$1,$2,$3)`, presentedDigest[:], now.Add(-time.Hour), now.Add(time.Hour)); err != nil {
			t.Fatal(err)
		}
		material := make([][]byte, 0, rotationAttempts*2)
		for attempt := range rotationAttempts {
			material = append(material, collisionRaw, bytes.Repeat([]byte{byte(0x80 + attempt)}, 16))
		}
		repository := newRepository(pool, &passwordVerifier{matched: true}, bytes.NewReader(bytes.Join(material, nil)), testKeys())
		if _, err := repository.Rotate(ctx, "40000000-0000-4000-8000-000000000001", presentedToken, now, now.Add(12*time.Hour)); !errors.Is(err, domain.ErrDependencyUnavailable) {
			t.Fatalf("rotation exhaustion: %v", err)
		}
		if _, err := repository.Resolve(ctx, presentedToken, now); err != nil {
			t.Fatalf("presented session was partially revoked: %v", err)
		}
	})

	t.Run("cross-account browser session is rotated", func(t *testing.T) {
		otherRaw := bytes.Repeat([]byte{0x88}, 32)
		otherToken := base64.RawURLEncoding.EncodeToString(otherRaw)
		otherDigest := sha256.Sum256(otherRaw)
		now := time.Date(2026, 9, 3, 13, 45, 0, 0, time.UTC)
		if _, err := pool.Exec(ctx, `INSERT INTO user_sessions(id,user_id,token_digest,created_at,expires_at) VALUES('50000000-0000-4000-8000-000000000008','40000000-0000-4000-8000-000000000002',$1,$2,$3)`, otherDigest[:], now.Add(-time.Hour), now.Add(time.Hour)); err != nil {
			t.Fatal(err)
		}
		repository := NewRepository(pool, &passwordVerifier{matched: true}, testKeys())
		issued, err := repository.Rotate(ctx, "40000000-0000-4000-8000-000000000001", otherToken, now, now.Add(12*time.Hour))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := repository.Resolve(ctx, otherToken, now); !errors.Is(err, domain.ErrAuthenticationRequired) {
			t.Fatalf("cross-account token remains active: %v", err)
		}
		if user, err := repository.Resolve(ctx, issued.Token, now); err != nil || user.ID != "40000000-0000-4000-8000-000000000001" {
			t.Fatalf("new session user=%+v err=%v", user, err)
		}
	})

	t.Run("concurrent rotations issue distinct sessions", func(t *testing.T) {
		repository := NewRepository(pool, &passwordVerifier{matched: true}, testKeys())
		now := time.Date(2026, 9, 3, 14, 0, 0, 0, time.UTC)
		const count = 6
		results := make(chan string, count)
		errors := make(chan error, count)
		var group sync.WaitGroup
		for range count {
			group.Add(1)
			go func() {
				defer group.Done()
				issued, err := repository.Rotate(ctx, "40000000-0000-4000-8000-000000000001", "", now, now.Add(12*time.Hour))
				if err != nil {
					errors <- err
					return
				}
				results <- issued.Token
			}()
		}
		group.Wait()
		close(results)
		close(errors)
		for err := range errors {
			t.Error(err)
		}
		seen := make(map[string]struct{}, count)
		for token := range results {
			seen[token] = struct{}{}
		}
		if len(seen) != count {
			t.Fatalf("got %d distinct tokens", len(seen))
		}
	})

	t.Run("expiry boundary cleans the matching row", func(t *testing.T) {
		raw := bytes.Repeat([]byte{0x66}, 32)
		token := base64.RawURLEncoding.EncodeToString(raw)
		digest := sha256.Sum256(raw)
		now := time.Date(2026, 9, 3, 15, 0, 0, 0, time.UTC)
		if _, err := pool.Exec(ctx, `INSERT INTO user_sessions(id,user_id,token_digest,created_at,expires_at) VALUES('50000000-0000-4000-8000-000000000006','40000000-0000-4000-8000-000000000001',$1,$2,$3)`, digest[:], now.Add(-time.Hour), now); err != nil {
			t.Fatal(err)
		}
		repository := NewRepository(pool, &passwordVerifier{matched: true}, testKeys())
		if _, err := repository.Resolve(ctx, token, now); !errors.Is(err, domain.ErrAuthenticationRequired) {
			t.Fatalf("resolve at expiry: %v", err)
		}
		var count int
		if err := pool.QueryRow(ctx, `SELECT count(*) FROM user_sessions WHERE token_digest=$1`, digest[:]).Scan(&count); err != nil || count != 0 {
			t.Fatalf("expired count=%d err=%v", count, err)
		}
	})

	t.Run("cancelled rotation rolls back and the repository recovers", func(t *testing.T) {
		lock, err := pool.Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		defer lock.Rollback(context.Background())
		if _, err := lock.Exec(ctx, `SELECT id FROM users WHERE id='40000000-0000-4000-8000-000000000001' FOR UPDATE`); err != nil {
			t.Fatal(err)
		}
		repository := NewRepository(pool, &passwordVerifier{matched: true}, testKeys())
		now := time.Date(2026, 9, 3, 16, 0, 0, 0, time.UTC)
		request, cancel := context.WithTimeout(ctx, 50*time.Millisecond)
		defer cancel()
		started := time.Now()
		if _, err := repository.Rotate(request, "40000000-0000-4000-8000-000000000001", collisionToken, now, now.Add(12*time.Hour)); !errors.Is(err, domain.ErrDependencyUnavailable) {
			t.Fatalf("cancelled rotation: %v", err)
		}
		if time.Since(started) > time.Second {
			t.Fatal("rotation ignored context cancellation")
		}
		if err := lock.Rollback(ctx); err != nil {
			t.Fatal(err)
		}
		if _, err := repository.Resolve(ctx, collisionToken, now); err != nil {
			t.Fatalf("presented token changed during cancelled rotation: %v", err)
		}
		if _, err := repository.Rotate(ctx, "40000000-0000-4000-8000-000000000001", "", now, now.Add(12*time.Hour)); err != nil {
			t.Fatalf("repository did not recover: %v", err)
		}
	})

	t.Run("session timestamps round trip at microsecond resolution", func(t *testing.T) {
		repository := newRepository(pool, &passwordVerifier{matched: true}, cryptorand.Reader, testKeys())
		wall := time.Now().UTC()
		if wall.Nanosecond()%int(time.Microsecond) == 0 {
			wall = wall.Add(321 * time.Nanosecond)
		}
		stored := wall.Truncate(time.Microsecond)
		issued, err := repository.Rotate(ctx, "40000000-0000-4000-8000-000000000001", "", stored, stored.Add(12*time.Hour))
		if err != nil {
			t.Fatal(err)
		}
		if !issued.Session.CreatedAt.Equal(stored) || !issued.Session.ExpiresAt.Equal(stored.Add(12*time.Hour)) {
			t.Fatalf("microsecond timestamps did not survive the round trip: %+v", issued.Session)
		}
		if _, err := repository.Rotate(ctx, "40000000-0000-4000-8000-000000000001", issued.Token, wall, wall.Add(12*time.Hour)); err != nil {
			t.Fatal(err)
		}
		var remainder int64
		if err := database.QueryRowContext(ctx, `SELECT (EXTRACT(microseconds FROM created_at)::bigint) FROM user_sessions ORDER BY created_at DESC LIMIT 1`).Scan(&remainder); err != nil {
			t.Fatal(err)
		}
		if time.Duration(wall.Nanosecond()) == time.Duration(remainder)*time.Microsecond {
			t.Fatalf("PostgreSQL preserved sub-microsecond precision; the application clock need not be normalized")
		}
	})

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
			`UPDATE users SET disabled_at='2026-09-03T12:00:00Z', deleted_at='2026-09-03T11:59:59Z' WHERE email='student@courseforge.local'`,
			`UPDATE users SET updated_at='2026-09-03T12:00:01Z', disabled_at='2026-09-03T12:00:01Z', deleted_at='2026-09-03T12:00:02Z' WHERE email='student@courseforge.local'`,
			`UPDATE users SET deleted_at='2026-09-03T12:00:00Z' WHERE email='student@courseforge.local'`,
			`UPDATE user_sessions SET token_digest=decode('ab','hex')`,
			`UPDATE user_sessions SET expires_at=created_at`,
			`UPDATE user_sessions SET token_digest=(SELECT min(token_digest) FROM user_sessions)`,
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

type passwordVerifier struct {
	matched     bool
	err         error
	verifyCalls int
	dummyCalls  int
}

func (verifier *passwordVerifier) Hash(string) (string, error) {
	return "", verifier.err
}

func (verifier *passwordVerifier) Verify(_ string, _ string) (bool, error) {
	verifier.verifyCalls++
	return verifier.matched, verifier.err
}

func (verifier *passwordVerifier) VerifyDummy(_ string) error {
	verifier.dummyCalls++
	return verifier.err
}

func encodePassword(password string, salt []byte) string {
	hash := argon2.IDKey([]byte(password), salt, 3, 64*1024, 2, 32)
	return fmt.Sprintf("$argon2id$v=19$m=65536,t=3,p=2$%s$%s", base64.RawStdEncoding.EncodeToString(salt), base64.RawStdEncoding.EncodeToString(hash))
}
