package ports

import (
	"context"
	"time"

	"courses/internal/identity/domain"
)

type AdministrationQuery struct {
	Limit  int
	Offset int
	Search string
	Status domain.AccountStatus
}

type StatusPolicy func(domain.User, int) error

type DeletePolicy func(domain.User) error

type AdministrationRepository interface {
	List(context.Context, AdministrationQuery) ([]domain.User, int64, error)
	SetStatus(context.Context, string, domain.AccountStatus, time.Time, StatusPolicy) error
	Delete(context.Context, string, time.Time, DeletePolicy) error
	SetPassword(context.Context, string, string, time.Time) error
}
