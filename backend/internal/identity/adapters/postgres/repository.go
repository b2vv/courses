package postgres

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io"
	"time"

	"courses/internal/identity/adapters/postgres/generated"
	"courses/internal/identity/domain"
	"courses/internal/identity/ports"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

const rotationAttempts = 3

type Passwords interface {
	Verify(string, string) (bool, error)
	VerifyDummy(string) error
	Hash(string) (string, error)
}

type Repository struct {
	pool      *pgxpool.Pool
	passwords Passwords
	random    io.Reader
}

var _ ports.UserWriter = (*Repository)(nil)
var _ ports.TokenStore = (*Repository)(nil)

func NewRepository(pool *pgxpool.Pool, passwords Passwords) *Repository {
	return newRepository(pool, passwords, rand.Reader)
}

func newRepository(pool *pgxpool.Pool, passwords Passwords, random io.Reader) *Repository {
	return &Repository{pool: pool, passwords: passwords, random: random}
}

func (repository *Repository) VerifyCredentials(ctx context.Context, email, password string) (domain.User, error) {
	row, err := generated.New(repository.pool).GetIdentityCredential(ctx, email)
	if errors.Is(err, pgx.ErrNoRows) {
		if err := repository.passwords.VerifyDummy(password); err != nil {
			return domain.User{}, domain.ErrDependencyUnavailable
		}
		return domain.User{}, domain.ErrInvalidCredentials
	}
	if err != nil {
		return domain.User{}, persistenceError(err)
	}
	matched, err := repository.passwords.Verify(row.PasswordHash, password)
	if err != nil {
		return domain.User{}, domain.ErrDependencyUnavailable
	}
	user, err := mapCredential(row)
	if err != nil {
		return domain.User{}, domain.ErrDependencyUnavailable
	}
	if !matched || !user.Active() {
		return domain.User{}, domain.ErrInvalidCredentials
	}
	return user, nil
}

func (repository *Repository) Rotate(ctx context.Context, userID, presentedToken string, createdAt, expiresAt time.Time) (ports.IssuedSession, error) {
	for range rotationAttempts {
		token, digest, sessionID, err := repository.newSessionMaterial()
		if err != nil {
			return ports.IssuedSession{}, domain.ErrDependencyUnavailable
		}
		tx, err := repository.pool.Begin(ctx)
		if err != nil {
			return ports.IssuedSession{}, persistenceError(err)
		}
		queries := generated.New(tx)
		if _, err := queries.LockActiveIdentityUser(ctx, userID); errors.Is(err, pgx.ErrNoRows) {
			rollback(tx)
			return ports.IssuedSession{}, domain.ErrInvalidCredentials
		} else if err != nil {
			rollback(tx)
			return ports.IssuedSession{}, persistenceError(err)
		}
		if raw, valid := decodeToken(presentedToken); valid {
			presentedDigest := sha256.Sum256(raw)
			if _, err := queries.DeleteIdentitySession(ctx, presentedDigest[:]); err != nil {
				rollback(tx)
				return ports.IssuedSession{}, persistenceError(err)
			}
		}
		row, err := queries.CreateIdentitySession(ctx, generated.CreateIdentitySessionParams{
			SessionID:   sessionID,
			UserID:      userID,
			TokenDigest: digest[:],
			CreatedAt:   timestamp(createdAt),
			ExpiresAt:   timestamp(expiresAt),
		})
		if err != nil {
			rollback(tx)
			if uniqueCollision(err) {
				continue
			}
			return ports.IssuedSession{}, persistenceError(err)
		}
		session, err := mapSession(row)
		if err != nil {
			rollback(tx)
			return ports.IssuedSession{}, domain.ErrDependencyUnavailable
		}
		if err := tx.Commit(ctx); err != nil {
			return ports.IssuedSession{}, persistenceError(err)
		}
		return ports.IssuedSession{Session: session, Token: token}, nil
	}
	return ports.IssuedSession{}, domain.ErrDependencyUnavailable
}

func (repository *Repository) Resolve(ctx context.Context, token string, now time.Time) (domain.User, error) {
	raw, valid := decodeToken(token)
	if !valid {
		return domain.User{}, domain.ErrAuthenticationRequired
	}
	digest := sha256.Sum256(raw)
	queries := generated.New(repository.pool)
	deleted, err := queries.DeleteExpiredIdentitySession(ctx, generated.DeleteExpiredIdentitySessionParams{TokenDigest: digest[:], Now: timestamp(now)})
	if err != nil {
		return domain.User{}, persistenceError(err)
	}
	if deleted > 0 {
		return domain.User{}, domain.ErrAuthenticationRequired
	}
	row, err := queries.ResolveIdentitySession(ctx, generated.ResolveIdentitySessionParams{TokenDigest: digest[:], Now: timestamp(now)})
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.User{}, domain.ErrAuthenticationRequired
	}
	if err != nil {
		return domain.User{}, persistenceError(err)
	}
	user, err := mapResolved(row)
	if err != nil {
		return domain.User{}, domain.ErrDependencyUnavailable
	}
	return user, nil
}

func (repository *Repository) Revoke(ctx context.Context, token string) error {
	raw, valid := decodeToken(token)
	if !valid {
		return nil
	}
	digest := sha256.Sum256(raw)
	if _, err := generated.New(repository.pool).DeleteIdentitySession(ctx, digest[:]); err != nil {
		return persistenceError(err)
	}
	return nil
}

func (repository *Repository) DeleteUserSessionsExceptCurrent(ctx context.Context, userID string, currentDigest []byte) error {
	if _, err := generated.New(repository.pool).DeleteUserSessionsExceptCurrent(ctx, generated.DeleteUserSessionsExceptCurrentParams{UserID: userID, TokenDigest: currentDigest}); err != nil {
		return persistenceError(err)
	}
	return nil
}

func (repository *Repository) CreateUser(ctx context.Context, user domain.User, passwordHash string) (domain.User, error) {
	row, err := generated.New(repository.pool).CreateUser(ctx, generated.CreateUserParams{
		UserID: user.ID, Email: user.Email, DisplayName: user.DisplayName, Role: string(user.Role), PasswordHash: passwordHash,
		EmailVerified: user.EmailVerified, CreatedAt: timestamp(user.CreatedAt), UpdatedAt: timestamp(user.UpdatedAt),
	})
	if err != nil {
		if emailTaken(err) {
			return domain.User{}, domain.ErrEmailTaken
		}
		return domain.User{}, persistenceError(err)
	}
	created, err := mapUser(row.ID, row.Email, row.DisplayName, row.Role, row.EmailVerified, row.CreatedAt, row.UpdatedAt, row.DisabledAt)
	if err != nil {
		return domain.User{}, domain.ErrDependencyUnavailable
	}
	return created, nil
}

func (repository *Repository) SetEmailVerified(ctx context.Context, userID string) error {
	rows, err := generated.New(repository.pool).SetUserEmailVerified(ctx, userID)
	if err != nil {
		return persistenceError(err)
	}
	if rows != 1 {
		return domain.ErrUserNotFound
	}
	return nil
}

func (repository *Repository) UpdatePassword(ctx context.Context, userID, passwordHash string) error {
	rows, err := generated.New(repository.pool).UpdateUserPassword(ctx, generated.UpdateUserPasswordParams{UserID: userID, PasswordHash: passwordHash})
	if err != nil {
		return persistenceError(err)
	}
	if rows != 1 {
		return domain.ErrUserNotFound
	}
	return nil
}

func (repository *Repository) UpdateEmail(ctx context.Context, userID, email string) error {
	rows, err := generated.New(repository.pool).UpdateUserEmail(ctx, generated.UpdateUserEmailParams{UserID: userID, Email: email})
	if err != nil {
		if emailTaken(err) {
			return domain.ErrEmailTaken
		}
		return persistenceError(err)
	}
	if rows != 1 {
		return domain.ErrUserNotFound
	}
	return nil
}

func (repository *Repository) Issue(ctx context.Context, token domain.AccountToken, digest []byte) error {
	newEmail := pgtype.Text{}
	if token.NewEmail != "" {
		newEmail = pgtype.Text{String: token.NewEmail, Valid: true}
	}
	if err := generated.New(repository.pool).InsertAccountToken(ctx, generated.InsertAccountTokenParams{
		TokenID: token.ID, UserID: token.UserID, Purpose: string(token.Purpose), TokenDigest: digest, NewEmail: newEmail,
		ExpiresAt: timestamp(token.ExpiresAt), CreatedAt: timestamp(token.CreatedAt),
	}); err != nil {
		return persistenceError(err)
	}
	return nil
}

func (repository *Repository) Consume(ctx context.Context, purpose domain.AccountTokenPurpose, digest []byte) (string, string, error) {
	row, err := generated.New(repository.pool).ConsumeAccountToken(ctx, generated.ConsumeAccountTokenParams{Purpose: string(purpose), TokenDigest: digest})
	if errors.Is(err, pgx.ErrNoRows) {
		return "", "", domain.ErrTokenUsed
	}
	if err != nil {
		return "", "", persistenceError(err)
	}
	newEmail := ""
	if row.NewEmail.Valid {
		newEmail = row.NewEmail.String
	}
	return row.UserID, newEmail, nil
}

func (repository *Repository) InvalidateByPurpose(ctx context.Context, userID string, purpose domain.AccountTokenPurpose) error {
	if _, err := generated.New(repository.pool).InvalidateAccountTokensByPurpose(ctx, generated.InvalidateAccountTokensByPurposeParams{UserID: userID, Purpose: string(purpose)}); err != nil {
		return persistenceError(err)
	}
	return nil
}

func (repository *Repository) ListInstructors(ctx context.Context) ([]domain.User, error) {
	rows, err := generated.New(repository.pool).ListInstructors(ctx)
	if err != nil {
		return nil, persistenceError(err)
	}
	users := make([]domain.User, 0, len(rows))
	for _, row := range rows {
		user, err := mapUser(row.ID, row.Email, row.DisplayName, row.Role, false, row.CreatedAt, row.UpdatedAt, row.DisabledAt)
		if err != nil {
			return nil, domain.ErrDependencyUnavailable
		}
		users = append(users, user)
	}
	return users, nil
}

func (repository *Repository) newSessionMaterial() (string, [32]byte, string, error) {
	raw := make([]byte, 32)
	if _, err := io.ReadFull(repository.random, raw); err != nil {
		return "", [32]byte{}, "", err
	}
	uuidBytes := make([]byte, 16)
	if _, err := io.ReadFull(repository.random, uuidBytes); err != nil {
		return "", [32]byte{}, "", err
	}
	uuidBytes[6] = uuidBytes[6]&0x0f | 0x40
	uuidBytes[8] = uuidBytes[8]&0x3f | 0x80
	token := base64.RawURLEncoding.EncodeToString(raw)
	digest := sha256.Sum256(raw)
	return token, digest, formatUUID(uuidBytes), nil
}

func formatUUID(value []byte) string {
	encoded := hex.EncodeToString(value)
	return encoded[0:8] + "-" + encoded[8:12] + "-" + encoded[12:16] + "-" + encoded[16:20] + "-" + encoded[20:32]
}

func decodeToken(token string) ([]byte, bool) {
	if len(token) != 43 {
		return nil, false
	}
	raw, err := base64.RawURLEncoding.DecodeString(token)
	return raw, err == nil && len(raw) == 32
}

func timestamp(value time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: value, Valid: true}
}

func mapCredential(row generated.GetIdentityCredentialRow) (domain.User, error) {
	return mapUser(row.ID, row.Email, row.DisplayName, row.Role, false, row.CreatedAt, row.UpdatedAt, row.DisabledAt)
}

func mapResolved(row generated.ResolveIdentitySessionRow) (domain.User, error) {
	return mapUser(row.ID, row.Email, row.DisplayName, row.Role, row.EmailVerified, row.CreatedAt, row.UpdatedAt, row.DisabledAt)
}

func mapUser(id, email, displayName, role string, emailVerified bool, createdAt, updatedAt, disabledAt pgtype.Timestamptz) (domain.User, error) {
	var disabled *time.Time
	if disabledAt.Valid {
		value := disabledAt.Time
		disabled = &value
	}
	if !createdAt.Valid || !updatedAt.Valid {
		return domain.User{}, domain.ErrInvalidUser
	}
	return domain.NewUser(domain.User{ID: id, Email: email, DisplayName: displayName, Role: domain.Role(role), EmailVerified: emailVerified, CreatedAt: createdAt.Time, UpdatedAt: updatedAt.Time, DisabledAt: disabled})
}

func mapSession(row generated.CreateIdentitySessionRow) (domain.Session, error) {
	if !row.CreatedAt.Valid || !row.ExpiresAt.Valid {
		return domain.Session{}, domain.ErrInvalidSession
	}
	return domain.NewSession(domain.Session{ID: row.ID, UserID: row.UserID, CreatedAt: row.CreatedAt.Time, ExpiresAt: row.ExpiresAt.Time})
}

func uniqueCollision(err error) bool {
	var postgresError *pgconn.PgError
	return errors.As(err, &postgresError) && postgresError.Code == "23505" && (postgresError.ConstraintName == "user_sessions_token_digest_key" || postgresError.ConstraintName == "user_sessions_pkey")
}

func emailTaken(err error) bool {
	var postgresError *pgconn.PgError
	return errors.As(err, &postgresError) && postgresError.Code == "23505"
}

func persistenceError(_ error) error {
	return domain.ErrDependencyUnavailable
}

func rollback(tx pgx.Tx) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_ = tx.Rollback(ctx)
}
