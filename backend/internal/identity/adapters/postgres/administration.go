package postgres

import (
	"context"
	"errors"
	"strings"
	"time"

	"courses/internal/identity/adapters/postgres/generated"
	"courses/internal/identity/domain"
	"courses/internal/identity/ports"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func (repository *Repository) List(ctx context.Context, query ports.AdministrationQuery) ([]domain.User, int64, error) {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return nil, 0, persistenceError(err)
	}
	defer rollback(tx)

	queries := generated.New(tx)
	search := strings.NewReplacer("!", "!!", "%", "!%", "_", "!_").Replace(query.Search)
	params := generated.ListAdminUsersParams{
		Status: string(query.Status), SearchPattern: search,
		PageLimit: int32(query.Limit), PageOffset: int32(query.Offset),
	}
	rows, err := queries.ListAdminUsers(ctx, params)
	if err != nil {
		return nil, 0, persistenceError(err)
	}
	total, err := queries.CountAdminUsers(ctx, generated.CountAdminUsersParams{Status: string(query.Status), SearchPattern: search})
	if err != nil {
		return nil, 0, persistenceError(err)
	}
	users := make([]domain.User, 0, len(rows))
	for _, row := range rows {
		user, err := mapAdministrationUser(row.ID, row.Email, row.DisplayName, row.Role, row.CreatedAt, row.UpdatedAt, row.DisabledAt, row.DeletedAt)
		if err != nil {
			return nil, 0, domain.ErrDependencyUnavailable
		}
		users = append(users, user)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, 0, persistenceError(err)
	}
	return users, total, nil
}

func (repository *Repository) SetStatus(ctx context.Context, userID string, status domain.AccountStatus, now time.Time, policy ports.StatusPolicy) error {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return persistenceError(err)
	}
	defer rollback(tx)

	queries := generated.New(tx)
	target, err := getAdministrationUser(ctx, queries, userID)
	if err != nil {
		return err
	}
	otherActiveAdmins := 0
	var disabledAt pgtype.Timestamptz
	if status == domain.AccountBlocked && target.Role == domain.Admin {
		admins, err := queries.LockAdmins(ctx)
		if err != nil {
			return persistenceError(err)
		}
		found := false
		for _, admin := range admins {
			if admin.ID == userID {
				disabledAt = admin.DisabledAt
				found = true
			}
		}
		if !found {
			return domain.ErrUserNotFound
		}
		if !disabledAt.Valid {
			for _, admin := range admins {
				if admin.ID != userID && !admin.DisabledAt.Valid {
					otherActiveAdmins++
				}
			}
		}
	} else {
		locked, err := queries.LockUserForStatus(ctx, userID)
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrUserNotFound
		}
		if err != nil {
			return persistenceError(err)
		}
		disabledAt = locked.DisabledAt
	}
	target, err = getAdministrationUser(ctx, queries, userID)
	if err != nil {
		return err
	}
	if err := policy(target, otherActiveAdmins); err != nil {
		return err
	}
	wantsBlocked := status == domain.AccountBlocked
	if wantsBlocked == disabledAt.Valid {
		return commit(ctx, tx)
	}
	writeTime := timestamp(now)
	if !wantsBlocked {
		writeTime = pgtype.Timestamptz{}
	}
	rows, err := queries.SetUserDisabledAt(ctx, generated.SetUserDisabledAtParams{
		DisabledAt: writeTime, UpdatedAt: timestamp(now), UserID: userID,
	})
	if err != nil {
		return persistenceError(err)
	}
	if rows != 1 {
		return domain.ErrUserNotFound
	}
	if wantsBlocked {
		if _, err := queries.DeleteIdentitySessionsForUser(ctx, userID); err != nil {
			return persistenceError(err)
		}
	}
	return commit(ctx, tx)
}

func (repository *Repository) Delete(ctx context.Context, userID string, now time.Time, policy ports.DeletePolicy) error {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return persistenceError(err)
	}
	defer rollback(tx)

	queries := generated.New(tx)
	if _, err := queries.LockUserForStatus(ctx, userID); errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrUserNotFound
	} else if err != nil {
		return persistenceError(err)
	}
	target, err := getAdministrationUser(ctx, queries, userID)
	if err != nil {
		return err
	}
	if err := policy(target); err != nil {
		return err
	}
	rows, err := queries.SoftDeleteUser(ctx, generated.SoftDeleteUserParams{DeletedAt: timestamp(now), UserID: userID})
	if err != nil {
		return persistenceError(err)
	}
	if rows != 1 {
		return domain.ErrUserNotFound
	}
	if _, err := queries.DeleteIdentitySessionsForUser(ctx, userID); err != nil {
		return persistenceError(err)
	}
	return commit(ctx, tx)
}

func (repository *Repository) SetPassword(ctx context.Context, userID, password string, now time.Time) error {
	hash, err := repository.passwords.Hash(password)
	if err != nil {
		return domain.ErrDependencyUnavailable
	}
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return persistenceError(err)
	}
	defer rollback(tx)

	queries := generated.New(tx)
	rows, err := queries.SetUserPasswordHash(ctx, generated.SetUserPasswordHashParams{
		PasswordHash: hash, UpdatedAt: timestamp(now), UserID: userID,
	})
	if err != nil {
		return persistenceError(err)
	}
	if rows != 1 {
		return domain.ErrUserNotFound
	}
	if _, err := queries.DeleteIdentitySessionsForUser(ctx, userID); err != nil {
		return persistenceError(err)
	}
	return commit(ctx, tx)
}

func getAdministrationUser(ctx context.Context, queries *generated.Queries, userID string) (domain.User, error) {
	row, err := queries.GetAdministrationUser(ctx, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.User{}, domain.ErrUserNotFound
	}
	if err != nil {
		return domain.User{}, persistenceError(err)
	}
	user, err := mapAdministrationUser(row.ID, row.Email, row.DisplayName, row.Role, row.CreatedAt, row.UpdatedAt, row.DisabledAt, row.DeletedAt)
	if err != nil {
		return domain.User{}, domain.ErrDependencyUnavailable
	}
	return user, nil
}

func mapAdministrationUser(
	id, email, displayName, role string,
	createdAt, updatedAt, disabledAt, deletedAt pgtype.Timestamptz,
) (domain.User, error) {
	if !createdAt.Valid || !updatedAt.Valid {
		return domain.User{}, domain.ErrInvalidUser
	}
	var disabled, deleted *time.Time
	if disabledAt.Valid {
		value := disabledAt.Time
		disabled = &value
	}
	if deletedAt.Valid {
		value := deletedAt.Time
		deleted = &value
	}
	return domain.NewUser(domain.User{
		ID: id, Email: email, DisplayName: displayName, Role: domain.Role(role),
		CreatedAt: createdAt.Time, UpdatedAt: updatedAt.Time,
		DisabledAt: disabled, DeletedAt: deleted,
	})
}

func commit(ctx context.Context, tx pgx.Tx) error {
	if err := tx.Commit(ctx); err != nil {
		return persistenceError(err)
	}
	return nil
}
