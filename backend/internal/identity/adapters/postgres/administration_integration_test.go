//go:build integration

package postgres

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	passwords "courses/internal/identity/adapters/argon2id"
	"courses/internal/identity/domain"
	"courses/internal/identity/ports"
	containers "courses/internal/testsupport/containers"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	postgrescontainer "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	administrationAdminOneID = "41000000-0000-4000-8000-000000000001"
	administrationAdminTwoID = "41000000-0000-4000-8000-000000000002"
	administrationStudentID  = "41000000-0000-4000-8000-000000000003"
	administrationTeacherID  = "41000000-0000-4000-8000-000000000004"
	administrationReuserID   = "41000000-0000-4000-8000-000000000005"
	administrationPassword   = "courseforge-original-password"
)

type administrationHarness struct {
	ctx          context.Context
	pool         *pgxpool.Pool
	repository   *Repository
	passwords    *passwords.Verifier
	passwordHash string
}

func TestAdministrationRepositoryAgainstPostgreSQL(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	container, err := containers.Start(ctx, func(ctx context.Context) (*postgrescontainer.PostgresContainer, error) {
		return postgrescontainer.Run(ctx, "postgres:17.5",
			postgrescontainer.WithDatabase("administration_test"),
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
	passwordVerifier := passwords.NewVerifier()
	passwordHash, err := passwordVerifier.Hash(administrationPassword)
	if err != nil {
		t.Fatal(err)
	}
	harness := &administrationHarness{
		ctx: ctx, pool: pool, repository: NewRepository(pool, passwordVerifier),
		passwords: passwordVerifier, passwordHash: passwordHash,
	}

	t.Run("P03 P04 B11 B12 block revokes sessions and restore reenables login", harness.testBlockRestore)
	t.Run("P05 R12 delete disables the row frees email and is terminal", harness.testDeleteAndReuse)
	t.Run("P06 B05 reset stores Argon2id and revokes sessions", harness.testResetPassword)
	t.Run("directory maps escaped ordered paged status rows", harness.testList)
	t.Run("B03 a non-last admin can be blocked", harness.testBlockNonLastAdmin)
	t.Run("R09 the last active admin is refused sequentially", harness.testLastAdmin)
	t.Run("R10 concurrent admin blocks leave one active admin", harness.testConcurrentAdminBlocks)
	t.Run("R16 a hashing failure performs no write", harness.testHasherFailure)
	t.Run("R18 cancellation rolls back block and session revoke", harness.testCancelledBlock)
	t.Run("R22 login races serialize with block and delete", harness.testLoginRaces)
	t.Run("R22 concurrent delete has one not-found loser", harness.testConcurrentDelete)
	t.Run("A03b account writers, tokens, session preservation, and verified resolve", harness.testAccountPersistence)
}

func (h *administrationHarness) testAccountPersistence(t *testing.T) {
	h.reset(t)
	now := time.Now().UTC().Truncate(time.Microsecond)
	newUser := func(id, email string) domain.User {
		user, err := domain.NewUser(domain.User{ID: id, Email: email, DisplayName: "Account User", Role: domain.Student, CreatedAt: now, UpdatedAt: now})
		if err != nil {
			t.Fatal(err)
		}
		return user
	}
	user := newUser("42000000-0000-4000-8000-000000000001", "account@example.test")
	created, err := h.repository.CreateUser(h.ctx, user, h.passwordHash)
	if err != nil || created.ID != user.ID || created.EmailVerified {
		t.Fatalf("create user=%+v err=%v", created, err)
	}
	if _, err := h.repository.CreateUser(h.ctx, newUser("42000000-0000-4000-8000-000000000002", user.Email), h.passwordHash); !errors.Is(err, domain.ErrEmailTaken) {
		t.Fatalf("duplicate email: %v", err)
	}

	start := make(chan struct{})
	results := make(chan error, 2)
	for _, id := range []string{"42000000-0000-4000-8000-000000000003", "42000000-0000-4000-8000-000000000004"} {
		candidate := newUser(id, "race@example.test")
		go func() {
			<-start
			_, err := h.repository.CreateUser(h.ctx, candidate, h.passwordHash)
			results <- err
		}()
	}
	close(start)
	successes, taken := 0, 0
	for range 2 {
		err := <-results
		if err == nil {
			successes++
		} else if errors.Is(err, domain.ErrEmailTaken) {
			taken++
		} else {
			t.Fatalf("racing create: %v", err)
		}
	}
	if successes != 1 || taken != 1 {
		t.Fatalf("racing create successes=%d email_taken=%d", successes, taken)
	}

	issued := h.issueSession(t, user.ID, now)
	if err := h.repository.SetEmailVerified(h.ctx, user.ID); err != nil {
		t.Fatal(err)
	}
	if resolved, err := h.repository.Resolve(h.ctx, issued.Token, now); err != nil || !resolved.EmailVerified {
		t.Fatalf("verified resolve=%+v err=%v", resolved, err)
	}

	digest := sha256.Sum256([]byte("email-change-token"))
	token, err := domain.NewAccountToken(domain.AccountToken{ID: "43000000-0000-4000-8000-000000000001", UserID: user.ID, Purpose: domain.EmailChangeToken, NewEmail: "new@example.test", CreatedAt: now, ExpiresAt: now.Add(time.Hour)})
	if err != nil {
		t.Fatal(err)
	}
	if err := h.repository.Issue(h.ctx, token, digest[:]); err != nil {
		t.Fatal(err)
	}
	if userID, newEmail, err := h.repository.Consume(h.ctx, token.Purpose, digest[:]); err != nil || userID != user.ID || newEmail != token.NewEmail {
		t.Fatalf("consume user=%q email=%q err=%v", userID, newEmail, err)
	}
	if _, _, err := h.repository.Consume(h.ctx, token.Purpose, digest[:]); !errors.Is(err, domain.ErrTokenUsed) {
		t.Fatalf("replayed consume: %v", err)
	}

	expiredDigest := sha256.Sum256([]byte("expired-token"))
	expired, err := domain.NewAccountToken(domain.AccountToken{ID: "43000000-0000-4000-8000-000000000002", UserID: user.ID, Purpose: domain.PasswordResetToken, CreatedAt: now.Add(-2 * time.Hour), ExpiresAt: now.Add(-time.Hour)})
	if err != nil {
		t.Fatalf("issue expired token: %v", err)
	}
	if err := h.repository.Issue(h.ctx, expired, expiredDigest[:]); err != nil {
		t.Fatalf("issue expired token: %v", err)
	}
	if _, _, err := h.repository.Consume(h.ctx, expired.Purpose, expiredDigest[:]); !errors.Is(err, domain.ErrTokenUsed) {
		t.Fatalf("expired consume: %v", err)
	}

	invalidatedDigest := sha256.Sum256([]byte("invalidated-token"))
	invalidated, err := domain.NewAccountToken(domain.AccountToken{ID: "43000000-0000-4000-8000-000000000003", UserID: user.ID, Purpose: domain.EmailVerificationToken, CreatedAt: now, ExpiresAt: now.Add(time.Hour)})
	if err != nil {
		t.Fatalf("issue invalidated token: %v", err)
	}
	if err := h.repository.Issue(h.ctx, invalidated, invalidatedDigest[:]); err != nil {
		t.Fatalf("issue invalidated token: %v", err)
	}
	if err := h.repository.InvalidateByPurpose(h.ctx, user.ID, invalidated.Purpose); err != nil {
		t.Fatal(err)
	}
	if _, _, err := h.repository.Consume(h.ctx, invalidated.Purpose, invalidatedDigest[:]); !errors.Is(err, domain.ErrTokenUsed) {
		t.Fatalf("invalidated consume: %v", err)
	}

	current := h.issueSession(t, user.ID, now)
	h.issueSession(t, user.ID, now)
	raw, valid := decodeToken(current.Token)
	if !valid {
		t.Fatal("issued invalid session token")
	}
	currentDigest := sha256.Sum256(raw)
	if err := h.repository.DeleteUserSessionsExceptCurrent(h.ctx, user.ID, currentDigest[:]); err != nil {
		t.Fatal(err)
	}
	if count := h.sessionCount(t, user.ID); count != 1 {
		t.Fatalf("remaining sessions=%d", count)
	}
	if _, err := h.repository.Resolve(h.ctx, current.Token, now); err != nil {
		t.Fatalf("current session removed: %v", err)
	}
}

func (h *administrationHarness) testBlockRestore(t *testing.T) {
	h.reset(t)
	now := time.Now().UTC().Truncate(time.Microsecond)
	issued := h.issueSession(t, administrationStudentID, now)
	if err := h.repository.SetStatus(h.ctx, administrationStudentID, domain.AccountBlocked, now, allowStatus); err != nil {
		t.Fatal(err)
	}
	if _, err := h.repository.Resolve(h.ctx, issued.Token, now); !errors.Is(err, domain.ErrAuthenticationRequired) {
		t.Fatalf("blocked session resolve: %v", err)
	}
	if _, err := h.repository.VerifyCredentials(h.ctx, "student@courseforge.local", administrationPassword); !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Fatalf("blocked login: %v", err)
	}
	var disabledAt, updatedAt time.Time
	if err := h.pool.QueryRow(h.ctx, `SELECT disabled_at, updated_at FROM users WHERE id=$1`, administrationStudentID).Scan(&disabledAt, &updatedAt); err != nil {
		t.Fatal(err)
	}
	if !disabledAt.Equal(now) || !updatedAt.Equal(now) {
		t.Fatalf("block timestamps disabled=%s updated=%s want=%s", disabledAt, updatedAt, now)
	}
	if count := h.sessionCount(t, administrationStudentID); count != 0 {
		t.Fatalf("blocked sessions=%d", count)
	}
	if _, err := h.pool.Exec(h.ctx, `INSERT INTO user_sessions(id,user_id,token_digest,created_at,expires_at)
		VALUES('51000000-0000-4000-8000-000000000001',$1,decode(repeat('ab',32),'hex'),$2,$3)`, administrationStudentID, now, now.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	if err := h.repository.SetStatus(h.ctx, administrationStudentID, domain.AccountBlocked, now.Add(time.Second), allowStatus); err != nil {
		t.Fatal(err)
	}
	var repeatedDisabledAt, repeatedUpdatedAt time.Time
	if err := h.pool.QueryRow(h.ctx, `SELECT disabled_at, updated_at FROM users WHERE id=$1`, administrationStudentID).Scan(&repeatedDisabledAt, &repeatedUpdatedAt); err != nil {
		t.Fatal(err)
	}
	if !repeatedDisabledAt.Equal(disabledAt) || !repeatedUpdatedAt.Equal(updatedAt) {
		t.Fatalf("same-state block moved timestamps disabled=%s updated=%s", repeatedDisabledAt, repeatedUpdatedAt)
	}
	if count := h.sessionCount(t, administrationStudentID); count != 1 {
		t.Fatalf("same-state block revoked sessions=%d", count)
	}

	restoredAt := now.Add(2 * time.Second)
	if err := h.repository.SetStatus(h.ctx, administrationStudentID, domain.AccountActive, restoredAt, allowStatus); err != nil {
		t.Fatal(err)
	}
	if _, err := h.repository.VerifyCredentials(h.ctx, "student@courseforge.local", administrationPassword); err != nil {
		t.Fatalf("restored login: %v", err)
	}
	restored := h.issueSession(t, administrationStudentID, restoredAt)
	if user, err := h.repository.Resolve(h.ctx, restored.Token, restoredAt); err != nil || user.ID != administrationStudentID {
		t.Fatalf("restored session user=%+v err=%v", user, err)
	}
}

func (h *administrationHarness) testDeleteAndReuse(t *testing.T) {
	h.reset(t)
	now := time.Now().UTC().Truncate(time.Microsecond)
	h.issueSession(t, administrationStudentID, now)
	if err := h.repository.Delete(h.ctx, administrationStudentID, now, allowDelete); err != nil {
		t.Fatal(err)
	}
	var disabledAt, deletedAt time.Time
	if err := h.pool.QueryRow(h.ctx, `SELECT disabled_at, deleted_at FROM users WHERE id=$1`, administrationStudentID).Scan(&disabledAt, &deletedAt); err != nil {
		t.Fatal(err)
	}
	if !disabledAt.Equal(now) || !deletedAt.Equal(now) || h.sessionCount(t, administrationStudentID) != 0 {
		t.Fatalf("deleted disabled=%s deleted=%s sessions=%d", disabledAt, deletedAt, h.sessionCount(t, administrationStudentID))
	}

	newPassword := "courseforge-reused-password"
	newHash, err := h.passwords.Hash(newPassword)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := h.pool.Exec(h.ctx, `INSERT INTO users(id,email,display_name,role,password_hash,created_at,updated_at) VALUES($1,'student@courseforge.local','Reused Student','student',$2,$3,$3)`, administrationReuserID, newHash, now); err != nil {
		t.Fatal(err)
	}
	if _, err := h.repository.VerifyCredentials(h.ctx, "student@courseforge.local", administrationPassword); !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Fatalf("deleted user's old password: %v", err)
	}
	if user, err := h.repository.VerifyCredentials(h.ctx, "student@courseforge.local", newPassword); err != nil || user.ID != administrationReuserID {
		t.Fatalf("reused email user=%+v err=%v", user, err)
	}
	if err := h.repository.SetStatus(h.ctx, administrationStudentID, domain.AccountActive, now.Add(time.Second), allowStatus); !errors.Is(err, domain.ErrUserNotFound) {
		t.Fatalf("restore deleted: %v", err)
	}
	if err := h.repository.Delete(h.ctx, administrationStudentID, now.Add(time.Second), allowDelete); !errors.Is(err, domain.ErrUserNotFound) {
		t.Fatalf("delete deleted: %v", err)
	}
	if err := h.repository.SetPassword(h.ctx, administrationStudentID, "replacement", now.Add(time.Second)); !errors.Is(err, domain.ErrUserNotFound) {
		t.Fatalf("reset deleted: %v", err)
	}
}

func (h *administrationHarness) testResetPassword(t *testing.T) {
	h.reset(t)
	now := time.Now().UTC().Truncate(time.Microsecond)
	issued := h.issueSession(t, administrationStudentID, now)
	const newPassword = "courseforge-reset-password"
	if err := h.repository.SetPassword(h.ctx, administrationStudentID, newPassword, now); err != nil {
		t.Fatal(err)
	}
	var hash string
	if err := h.pool.QueryRow(h.ctx, `SELECT password_hash FROM users WHERE id=$1`, administrationStudentID).Scan(&hash); err != nil {
		t.Fatal(err)
	}
	matched, err := h.passwords.Verify(hash, newPassword)
	if err != nil || !matched || !strings.HasPrefix(hash, "$argon2id$v=19$m=65536,t=3,p=2$") {
		t.Fatalf("stored hash=%q matched=%v err=%v", hash, matched, err)
	}
	if _, err := h.repository.VerifyCredentials(h.ctx, "student@courseforge.local", administrationPassword); !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Fatalf("old password: %v", err)
	}
	if _, err := h.repository.VerifyCredentials(h.ctx, "student@courseforge.local", newPassword); err != nil {
		t.Fatalf("new password: %v", err)
	}
	if _, err := h.repository.Resolve(h.ctx, issued.Token, now); !errors.Is(err, domain.ErrAuthenticationRequired) {
		t.Fatalf("pre-reset session: %v", err)
	}
	blockedAt := now.Add(time.Second)
	if err := h.repository.SetStatus(h.ctx, administrationStudentID, domain.AccountBlocked, blockedAt, allowStatus); err != nil {
		t.Fatal(err)
	}
	const blockedPassword = "courseforge-blocked-reset"
	if err := h.repository.SetPassword(h.ctx, administrationStudentID, blockedPassword, blockedAt.Add(time.Second)); err != nil {
		t.Fatalf("reset blocked user: %v", err)
	}
	var blockedHash string
	var blocked bool
	if err := h.pool.QueryRow(h.ctx, `SELECT password_hash, disabled_at IS NOT NULL FROM users WHERE id=$1`, administrationStudentID).Scan(&blockedHash, &blocked); err != nil {
		t.Fatal(err)
	}
	matched, err = h.passwords.Verify(blockedHash, blockedPassword)
	if err != nil || !matched || !blocked {
		t.Fatalf("blocked reset hash=%q matched=%v blocked=%v err=%v", blockedHash, matched, blocked, err)
	}
}

func (h *administrationHarness) testList(t *testing.T) {
	h.reset(t)
	now := time.Now().UTC().Truncate(time.Microsecond)
	if _, err := h.pool.Exec(h.ctx, `UPDATE users SET email='literal%_!@courseforge.local', display_name='Literal %_! User' WHERE id=$1`, administrationStudentID); err != nil {
		t.Fatal(err)
	}
	if err := h.repository.SetStatus(h.ctx, administrationStudentID, domain.AccountBlocked, now, allowStatus); err != nil {
		t.Fatal(err)
	}
	if err := h.repository.Delete(h.ctx, administrationTeacherID, now, allowDelete); err != nil {
		t.Fatal(err)
	}
	items, total, err := h.repository.List(h.ctx, ports.AdministrationQuery{Limit: 10, Search: "%_!", Status: domain.AccountBlocked})
	if err != nil {
		t.Fatal(err)
	}
	if total != 1 || len(items) != 1 || items[0].ID != administrationStudentID ||
		items[0].Email != "literal%_!@courseforge.local" || items[0].DisplayName != "Literal %_! User" ||
		items[0].Role != domain.Student || items[0].AccountStatus() != domain.AccountBlocked {
		t.Fatalf("escaped list total=%d items=%+v", total, items)
	}
	items, total, err = h.repository.List(h.ctx, ports.AdministrationQuery{Limit: 2})
	if err != nil {
		t.Fatal(err)
	}
	if total != 3 || len(items) != 2 || items[0].Email != "admin-one@courseforge.local" || items[1].Email != "admin-two@courseforge.local" {
		t.Fatalf("default ordered page total=%d items=%+v", total, items)
	}
	items, total, err = h.repository.List(h.ctx, ports.AdministrationQuery{Limit: 2, Offset: 10})
	if err != nil || total != 3 || len(items) != 0 {
		t.Fatalf("past-end page total=%d items=%+v err=%v", total, items, err)
	}
	items, total, err = h.repository.List(h.ctx, ports.AdministrationQuery{Limit: 10, Status: domain.AccountDeleted})
	if err != nil || total != 1 || len(items) != 1 || items[0].ID != administrationTeacherID ||
		items[0].DisabledAt == nil || items[0].DeletedAt == nil || items[0].AccountStatus() != domain.AccountDeleted {
		t.Fatalf("deleted list total=%d items=%+v err=%v", total, items, err)
	}

	// Give the three non-deleted accounts deterministic names and ages so email,
	// name and recency produce three distinct orders. Emails (COLLATE "C"):
	// admin-one < admin-two < literal%_! (student). Names put admin-two first;
	// recency puts the student first.
	if _, err := h.pool.Exec(h.ctx, `UPDATE users SET display_name=CASE id
		WHEN $1 THEN 'Bohdan' WHEN $2 THEN 'Andrii' WHEN $3 THEN 'Zoriana' END,
		created_at=CASE id WHEN $1 THEN $4::timestamptz WHEN $2 THEN $5::timestamptz WHEN $3 THEN $6::timestamptz END
		WHERE id IN ($1,$2,$3)`,
		administrationAdminOneID, administrationAdminTwoID, administrationStudentID,
		now.Add(-3*time.Hour), now.Add(-2*time.Hour), now.Add(-1*time.Hour)); err != nil {
		t.Fatal(err)
	}
	recent, _, err := h.repository.List(h.ctx, ports.AdministrationQuery{Limit: 10, Sort: domain.SortByRecent})
	if err != nil || len(recent) != 3 || recent[0].ID != administrationStudentID || recent[2].ID != administrationAdminOneID {
		t.Fatalf("recent sort items=%+v err=%v", recent, err)
	}
	byName, _, err := h.repository.List(h.ctx, ports.AdministrationQuery{Limit: 10, Sort: domain.SortByName})
	if err != nil || len(byName) != 3 || byName[0].ID != administrationAdminTwoID || byName[2].ID != administrationStudentID {
		t.Fatalf("name sort items=%+v err=%v", byName, err)
	}
	admins, total, err := h.repository.List(h.ctx, ports.AdministrationQuery{Limit: 10, Role: domain.Admin})
	if err != nil || total != 2 || len(admins) != 2 || admins[0].Role != domain.Admin || admins[1].Role != domain.Admin {
		t.Fatalf("role filter total=%d items=%+v err=%v", total, admins, err)
	}
}

func (h *administrationHarness) testBlockNonLastAdmin(t *testing.T) {
	h.reset(t)
	actor := domain.Actor{UserID: administrationAdminOneID}
	err := h.repository.SetStatus(h.ctx, administrationAdminTwoID, domain.AccountBlocked, time.Now().UTC().Truncate(time.Microsecond), func(target domain.User, otherActiveAdmins int) error {
		return domain.CanBlock(actor, target, otherActiveAdmins)
	})
	if err != nil {
		t.Fatal(err)
	}
	if active := h.activeAdminCount(t); active != 1 {
		t.Fatalf("active admins=%d", active)
	}
}

func (h *administrationHarness) testLastAdmin(t *testing.T) {
	h.reset(t)
	now := time.Now().UTC().Truncate(time.Microsecond)
	if _, err := h.pool.Exec(h.ctx, `UPDATE users SET disabled_at=$2, updated_at=$2 WHERE id=$1`, administrationAdminTwoID, now); err != nil {
		t.Fatal(err)
	}
	actor := domain.Actor{UserID: administrationStudentID}
	err := h.repository.SetStatus(h.ctx, administrationAdminOneID, domain.AccountBlocked, now.Add(time.Second), func(target domain.User, otherActiveAdmins int) error {
		return domain.CanBlock(actor, target, otherActiveAdmins)
	})
	if !errors.Is(err, domain.ErrLastAdmin) || h.activeAdminCount(t) != 1 {
		t.Fatalf("last admin err=%v active=%d", err, h.activeAdminCount(t))
	}
}

func (h *administrationHarness) testConcurrentAdminBlocks(t *testing.T) {
	h.reset(t)
	start := make(chan struct{})
	errorsChannel := make(chan error, 2)
	for _, pair := range [][2]string{{administrationAdminOneID, administrationAdminTwoID}, {administrationAdminTwoID, administrationAdminOneID}} {
		actorID, targetID := pair[0], pair[1]
		go func() {
			<-start
			errorsChannel <- h.repository.SetStatus(h.ctx, targetID, domain.AccountBlocked, time.Now().UTC().Truncate(time.Microsecond), func(target domain.User, otherActiveAdmins int) error {
				return domain.CanBlock(domain.Actor{UserID: actorID}, target, otherActiveAdmins)
			})
		}()
	}
	close(start)
	var successes, refusals int
	for range 2 {
		err := <-errorsChannel
		switch {
		case err == nil:
			successes++
		case errors.Is(err, domain.ErrLastAdmin):
			refusals++
		default:
			t.Fatalf("concurrent block: %v", err)
		}
	}
	if successes != 1 || refusals != 1 || h.activeAdminCount(t) != 1 {
		t.Fatalf("successes=%d refusals=%d active=%d", successes, refusals, h.activeAdminCount(t))
	}
}

func (h *administrationHarness) testHasherFailure(t *testing.T) {
	h.reset(t)
	var before string
	if err := h.pool.QueryRow(h.ctx, `SELECT password_hash FROM users WHERE id=$1`, administrationStudentID).Scan(&before); err != nil {
		t.Fatal(err)
	}
	repository := NewRepository(h.pool, failingPasswords{})
	if err := repository.SetPassword(h.ctx, administrationStudentID, "new password", time.Now().UTC().Truncate(time.Microsecond)); !errors.Is(err, domain.ErrDependencyUnavailable) {
		t.Fatalf("hash failure: %v", err)
	}
	var after string
	if err := h.pool.QueryRow(h.ctx, `SELECT password_hash FROM users WHERE id=$1`, administrationStudentID).Scan(&after); err != nil {
		t.Fatal(err)
	}
	if after != before {
		t.Fatal("hash failure changed the credential")
	}
}

func (h *administrationHarness) testCancelledBlock(t *testing.T) {
	h.reset(t)
	now := time.Now().UTC().Truncate(time.Microsecond)
	h.issueSession(t, administrationStudentID, now)
	lock, err := h.pool.Begin(h.ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer lock.Rollback(context.Background())
	if _, err := lock.Exec(h.ctx, `SELECT id FROM users WHERE id=$1 FOR UPDATE`, administrationStudentID); err != nil {
		t.Fatal(err)
	}
	request, cancel := context.WithTimeout(h.ctx, 50*time.Millisecond)
	defer cancel()
	if err := h.repository.SetStatus(request, administrationStudentID, domain.AccountBlocked, now.Add(time.Second), allowStatus); err == nil {
		t.Fatal("cancelled block succeeded")
	}
	if err := lock.Rollback(h.ctx); err != nil {
		t.Fatal(err)
	}
	var disabled bool
	if err := h.pool.QueryRow(h.ctx, `SELECT disabled_at IS NOT NULL FROM users WHERE id=$1`, administrationStudentID).Scan(&disabled); err != nil {
		t.Fatal(err)
	}
	if disabled || h.sessionCount(t, administrationStudentID) != 1 {
		t.Fatalf("cancelled block disabled=%v sessions=%d", disabled, h.sessionCount(t, administrationStudentID))
	}
}

func (h *administrationHarness) testLoginRaces(t *testing.T) {
	for _, mutation := range []string{"block", "delete"} {
		t.Run(mutation, func(t *testing.T) {
			h.reset(t)
			now := time.Now().UTC().Truncate(time.Microsecond)
			lock, err := h.pool.Begin(h.ctx)
			if err != nil {
				t.Fatal(err)
			}
			defer lock.Rollback(context.Background())
			if _, err := lock.Exec(h.ctx, `SELECT id FROM users WHERE id=$1 FOR UPDATE`, administrationStudentID); err != nil {
				t.Fatal(err)
			}
			login := make(chan error, 1)
			go func() {
				_, err := h.repository.Rotate(h.ctx, administrationStudentID, "", now, now.Add(time.Hour))
				login <- err
			}()
			h.waitForBlockedQuery(t, "-- name: LockActiveIdentityUser")
			mutated := make(chan error, 1)
			go func() {
				if mutation == "block" {
					mutated <- h.repository.SetStatus(h.ctx, administrationStudentID, domain.AccountBlocked, now.Add(time.Second), allowStatus)
					return
				}
				mutated <- h.repository.Delete(h.ctx, administrationStudentID, now.Add(time.Second), allowDelete)
			}()
			h.waitForBlockedQuery(t, "-- name: LockUserForStatus")
			if err := lock.Commit(h.ctx); err != nil {
				t.Fatal(err)
			}
			if err := <-login; err != nil && !errors.Is(err, domain.ErrInvalidCredentials) {
				t.Fatalf("login result: %v", err)
			}
			if err := <-mutated; err != nil {
				t.Fatalf("%s result: %v", mutation, err)
			}
			if count := h.sessionCount(t, administrationStudentID); count != 0 {
				t.Fatalf("%s race left sessions=%d", mutation, count)
			}
		})
	}
}

func (h *administrationHarness) testConcurrentDelete(t *testing.T) {
	h.reset(t)
	start := make(chan struct{})
	results := make(chan error, 2)
	for range 2 {
		go func() {
			<-start
			results <- h.repository.Delete(h.ctx, administrationStudentID, time.Now().UTC().Truncate(time.Microsecond), allowDelete)
		}()
	}
	close(start)
	var successes, missing int
	for range 2 {
		err := <-results
		switch {
		case err == nil:
			successes++
		case errors.Is(err, domain.ErrUserNotFound):
			missing++
		default:
			t.Fatalf("concurrent delete: %v", err)
		}
	}
	if successes != 1 || missing != 1 {
		t.Fatalf("successes=%d missing=%d", successes, missing)
	}
}

func (h *administrationHarness) reset(t *testing.T) {
	t.Helper()
	if _, err := h.pool.Exec(h.ctx, `TRUNCATE users CASCADE`); err != nil {
		t.Fatal(err)
	}
	createdAt := time.Now().UTC().Add(-time.Minute).Truncate(time.Microsecond)
	if _, err := h.pool.Exec(h.ctx, `INSERT INTO users(id,email,display_name,role,password_hash,created_at,updated_at) VALUES
		($1,'admin-one@courseforge.local','Admin One','admin',$5,$6,$6),
		($2,'admin-two@courseforge.local','Admin Two','admin',$5,$6,$6),
		($3,'student@courseforge.local','Student User','student',$5,$6,$6),
		($4,'teacher@courseforge.local','Teacher User','instructor',$5,$6,$6)`,
		administrationAdminOneID, administrationAdminTwoID, administrationStudentID, administrationTeacherID, h.passwordHash, createdAt); err != nil {
		t.Fatal(err)
	}
}

func (h *administrationHarness) issueSession(t *testing.T, userID string, now time.Time) ports.IssuedSession {
	t.Helper()
	issued, err := h.repository.Rotate(h.ctx, userID, "", now, now.Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	return issued
}

func (h *administrationHarness) sessionCount(t *testing.T, userID string) int {
	t.Helper()
	var count int
	if err := h.pool.QueryRow(h.ctx, `SELECT count(*) FROM user_sessions WHERE user_id=$1`, userID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	return count
}

func (h *administrationHarness) activeAdminCount(t *testing.T) int {
	t.Helper()
	var count int
	if err := h.pool.QueryRow(h.ctx, `SELECT count(*) FROM users WHERE role='admin' AND disabled_at IS NULL AND deleted_at IS NULL`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	return count
}

func (h *administrationHarness) waitForBlockedQuery(t *testing.T, queryName string) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		var blocked bool
		err := h.pool.QueryRow(h.ctx, `SELECT EXISTS (
			SELECT 1 FROM pg_stat_activity
			WHERE pid <> pg_backend_pid() AND datname=current_database()
				AND state='active' AND wait_event_type='Lock' AND position($1 in query) > 0
		)`, queryName).Scan(&blocked)
		if err != nil {
			t.Fatal(err)
		}
		if blocked {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("query %q did not block", queryName)
}

func allowStatus(domain.User, int) error { return nil }

func allowDelete(domain.User) error { return nil }

type failingPasswords struct{}

func (failingPasswords) Verify(string, string) (bool, error) { return false, errors.New("unused") }

func (failingPasswords) VerifyDummy(string) error { return errors.New("unused") }

func (failingPasswords) Hash(string) (string, error) { return "", errors.New("hash failed") }
