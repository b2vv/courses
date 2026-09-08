package application

import (
	"context"
	"time"

	"courses/internal/playback/domain"
	"courses/internal/playback/ports"
)

type Sessions struct {
	ownership ports.OwnershipWriter
	progress  ports.ProgressWriter
	ttl       time.Duration
	now       func() time.Time
}

func NewSessions(ownership ports.OwnershipWriter, progress ports.ProgressWriter, ttl time.Duration, now func() time.Time) *Sessions {
	return &Sessions{ownership: ownership, progress: progress, ttl: ttl, now: now}
}

func (sessions *Sessions) Take(ctx context.Context, userID string, ownership domain.Ownership) error {
	return sessions.ownership.Take(ctx, userID, ownership, sessions.ttl)
}

func (sessions *Sessions) Heartbeat(
	ctx context.Context,
	userID, playbackID, lessonID string,
	update domain.ProgressUpdate,
) error {
	renewal, err := sessions.ownership.Renew(ctx, userID, playbackID, sessions.ttl)
	if err != nil {
		return err
	}
	switch renewal {
	case domain.OwnershipMissing:
		return domain.ErrPlaybackSessionExpired
	case domain.OwnershipTakenOver:
		return domain.ErrPlaybackTakenOver
	case domain.OwnershipRenewed:
		return sessions.progress.Save(ctx, userID, lessonID, update, sessions.now())
	default:
		return domain.ErrDependencyUnavailable
	}
}
