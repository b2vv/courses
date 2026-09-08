package application

import (
	"context"
	"time"

	"courses/internal/playback/domain"
	"courses/internal/playback/ports"
)

type Sessions struct {
	playback  ports.LessonPlaybackReader
	ownership ports.OwnershipWriter
	progress  ports.ProgressWriter
	ttl       time.Duration
	now       func() time.Time
}

func NewSessions(playback ports.LessonPlaybackReader, ownership ports.OwnershipWriter, progress ports.ProgressWriter, ttl time.Duration, now func() time.Time) *Sessions {
	return &Sessions{playback: playback, ownership: ownership, progress: progress, ttl: ttl, now: now}
}

func (sessions *Sessions) Take(ctx context.Context, userID string, ownership domain.Ownership) error {
	return sessions.ownership.Take(ctx, userID, ownership, sessions.ttl)
}

func (sessions *Sessions) Play(ctx context.Context, actor domain.Actor, lessonID, playbackID string) (domain.LessonPlayback, error) {
	playback, err := sessions.loadViewablePlayback(ctx, actor, lessonID)
	if err != nil {
		return domain.LessonPlayback{}, err
	}
	err = sessions.Take(ctx, actor.UserID, domain.Ownership{
		PlaybackID: playbackID,
		VideoID:    playback.VideoID,
		LessonID:   lessonID,
	})
	if err != nil {
		return domain.LessonPlayback{}, err
	}
	return playback, nil
}

func (sessions *Sessions) Heartbeat(
	ctx context.Context,
	actor domain.Actor,
	playbackID, lessonID string,
	update domain.ProgressUpdate,
) error {
	_, err := sessions.loadViewablePlayback(ctx, actor, lessonID)
	if err != nil {
		return err
	}
	renewal, err := sessions.ownership.Renew(ctx, actor.UserID, playbackID, sessions.ttl)
	if err != nil {
		return err
	}
	switch renewal {
	case domain.OwnershipMissing:
		return domain.ErrPlaybackSessionExpired
	case domain.OwnershipTakenOver:
		return domain.ErrPlaybackTakenOver
	case domain.OwnershipRenewed:
		return sessions.progress.Save(ctx, actor.UserID, lessonID, update, sessions.now())
	default:
		return domain.ErrDependencyUnavailable
	}
}

func (sessions *Sessions) loadViewablePlayback(ctx context.Context, actor domain.Actor, lessonID string) (domain.LessonPlayback, error) {
	playback, err := sessions.playback.LoadLessonPlayback(ctx, actor.UserID, lessonID)
	if err != nil {
		return domain.LessonPlayback{}, err
	}
	if !playback.Course.CanBeViewedBy(actor) {
		return domain.LessonPlayback{}, domain.ErrForbidden
	}
	return playback, nil
}
