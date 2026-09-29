package application

import (
	"context"

	"courses/internal/playback/domain"
	"courses/internal/playback/ports"
)

type Authorizer struct {
	ownership ports.OwnershipReader
	access    ports.VideoAccessReader
}

func NewAuthorizer(ownership ports.OwnershipReader, access ports.VideoAccessReader) *Authorizer {
	return &Authorizer{ownership: ownership, access: access}
}

func (authorizer *Authorizer) Authorize(ctx context.Context, actor domain.Actor, playbackID, videoID string) error {
	owner, err := authorizer.ownership.Load(ctx, actor.UserID)
	if err != nil {
		return err
	}
	if owner.PlaybackID != playbackID || owner.VideoID != videoID {
		return domain.ErrForbidden
	}
	course, err := authorizer.access.LoadVideoCourse(ctx, actor.UserID, videoID)
	if err != nil {
		return err
	}
	if !course.CanBeViewedBy(actor) {
		if course.PurchaseRequiredBy(actor) {
			return domain.ErrPurchaseRequired
		}
		return domain.ErrForbidden
	}
	return nil
}
