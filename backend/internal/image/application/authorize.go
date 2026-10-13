package application

import (
	"context"

	"courses/internal/image/domain"
	"courses/internal/image/ports"
)

type Authorizer struct {
	access ports.AccessReader
}

func NewAuthorizer(access ports.AccessReader) *Authorizer {
	return &Authorizer{access: access}
}

func (a *Authorizer) Authorize(ctx context.Context, actor domain.Actor, imageID string) error {
	course, err := a.access.LoadImageAccess(ctx, actor.UserID, imageID)
	if err != nil {
		return err
	}
	if !actor.EmailVerified {
		return domain.ErrEmailNotVerified
	}
	if !course.CanConsumeBy(actor) {
		return domain.ErrForbidden
	}
	return nil
}
