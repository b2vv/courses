package application

import (
	"context"
	"time"

	"courses/internal/identity/domain"
	"courses/internal/identity/ports"
)

type Administration struct {
	repository ports.AdministrationRepository
	now        Clock
}

type ListUsersInput struct {
	Limit  int
	Offset int
	Search string
	Status domain.AccountStatus
}

type UserPage struct {
	Items  []domain.User
	Limit  int
	Offset int
	Total  int64
}

func NewAdministration(repository ports.AdministrationRepository, now Clock) *Administration {
	return &Administration{repository: repository, now: now}
}

func (administration *Administration) ListUsers(ctx context.Context, input ListUsersInput) (UserPage, error) {
	callCtx, cancel := context.WithTimeout(ctx, operationTimeout)
	defer cancel()
	users, total, err := administration.repository.List(callCtx, ports.AdministrationQuery{
		Limit: input.Limit, Offset: input.Offset, Search: input.Search, Status: input.Status,
	})
	if err != nil {
		return UserPage{}, administrationOperationError(ctx, err)
	}
	if total < 0 {
		return UserPage{}, domain.ErrDependencyUnavailable
	}
	for _, user := range users {
		if _, err := domain.NewUser(user); err != nil {
			return UserPage{}, domain.ErrDependencyUnavailable
		}
	}
	return UserPage{Items: users, Limit: input.Limit, Offset: input.Offset, Total: total}, nil
}

func (administration *Administration) SetStatus(ctx context.Context, actor domain.Actor, targetID string, status domain.AccountStatus) error {
	if status != domain.AccountActive && status != domain.AccountBlocked {
		return domain.ErrInvalidUser
	}
	if status == domain.AccountBlocked && actor.UserID == targetID {
		return domain.CanBlock(actor, domain.User{ID: targetID}, 0)
	}

	callCtx, cancel := context.WithTimeout(ctx, operationTimeout)
	defer cancel()
	err := administration.repository.SetStatus(callCtx, targetID, status, administration.timestamp(), func(target domain.User, otherActiveAdmins int) error {
		if err := validStoredUser(target); err != nil {
			return err
		}
		if status == domain.AccountBlocked {
			return domain.CanBlock(actor, target, otherActiveAdmins)
		}
		return nil
	})
	return administrationOperationError(ctx, err)
}

func (administration *Administration) Delete(ctx context.Context, actor domain.Actor, targetID string) error {
	if actor.UserID == targetID {
		return domain.CanDelete(actor, domain.User{ID: targetID})
	}

	callCtx, cancel := context.WithTimeout(ctx, operationTimeout)
	defer cancel()
	err := administration.repository.Delete(callCtx, targetID, administration.timestamp(), func(target domain.User) error {
		if err := validStoredUser(target); err != nil {
			return err
		}
		return domain.CanDelete(actor, target)
	})
	return administrationOperationError(ctx, err)
}

func (administration *Administration) ResetPassword(ctx context.Context, targetID, password string) error {
	if !domain.ValidPassword(password) {
		return domain.ErrInvalidUser
	}

	callCtx, cancel := context.WithTimeout(ctx, operationTimeout)
	defer cancel()
	return administrationOperationError(ctx, administration.repository.SetPassword(callCtx, targetID, password, administration.timestamp()))
}

func (administration *Administration) timestamp() time.Time {
	return administration.now().UTC().Truncate(storeResolution)
}

func administrationOperationError(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}
	if ctxErr := ctx.Err(); ctxErr != nil {
		return ctxErr
	}
	return err
}

func validStoredUser(user domain.User) error {
	if _, err := domain.NewUser(user); err != nil {
		return domain.ErrDependencyUnavailable
	}
	return nil
}
