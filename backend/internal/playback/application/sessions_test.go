package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"courses/internal/playback/application"
	"courses/internal/playback/domain"
)

type ownershipWriter struct {
	taken         domain.Ownership
	takeUserID    string
	takeTTL       time.Duration
	takeErr       error
	renewal       domain.OwnershipRenewal
	renewUserID   string
	renewPlayback string
	renewTTL      time.Duration
	renewErr      error
}

func (writer *ownershipWriter) Take(_ context.Context, userID string, ownership domain.Ownership, ttl time.Duration) error {
	writer.takeUserID, writer.taken, writer.takeTTL = userID, ownership, ttl
	return writer.takeErr
}

func (writer *ownershipWriter) Renew(_ context.Context, userID, playbackID string, ttl time.Duration) (domain.OwnershipRenewal, error) {
	writer.renewUserID, writer.renewPlayback, writer.renewTTL = userID, playbackID, ttl
	return writer.renewal, writer.renewErr
}

type progressWriter struct {
	saved    bool
	userID   string
	lessonID string
	update   domain.ProgressUpdate
	now      time.Time
	err      error
}

func (writer *progressWriter) Save(_ context.Context, userID, lessonID string, update domain.ProgressUpdate, now time.Time) error {
	writer.saved, writer.userID, writer.lessonID, writer.update, writer.now = true, userID, lessonID, update, now
	return writer.err
}

func TestTakeUnconditionallyStoresTheNewOwnership(t *testing.T) {
	t.Parallel()
	ownerships := &ownershipWriter{}
	progress := &progressWriter{}
	service := application.NewSessions(ownerships, progress, 3*time.Minute, time.Now)
	want := domain.Ownership{PlaybackID: playbackID, VideoID: videoID, LessonID: lessonID}

	if err := service.Take(context.Background(), userID, want); err != nil {
		t.Fatal(err)
	}
	if ownerships.takeUserID != userID || ownerships.taken != want || ownerships.takeTTL != 3*time.Minute {
		t.Fatalf("user=%q ownership=%+v ttl=%v", ownerships.takeUserID, ownerships.taken, ownerships.takeTTL)
	}
}

func TestHeartbeatRenewsBeforeSavingProgress(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, time.September, 8, 13, 0, 0, 0, time.UTC)
	ownerships := &ownershipWriter{renewal: domain.OwnershipRenewed}
	progress := &progressWriter{}
	service := application.NewSessions(ownerships, progress, 3*time.Minute, func() time.Time { return now })
	update, err := domain.NewProgressUpdate(42, 7, "pause")
	if err != nil {
		t.Fatal(err)
	}

	if err := service.Heartbeat(context.Background(), userID, playbackID, lessonID, update); err != nil {
		t.Fatal(err)
	}
	if ownerships.renewUserID != userID || ownerships.renewPlayback != playbackID || ownerships.renewTTL != 3*time.Minute {
		t.Fatalf("renew user=%q playback=%q ttl=%v", ownerships.renewUserID, ownerships.renewPlayback, ownerships.renewTTL)
	}
	if !progress.saved || progress.userID != userID || progress.lessonID != lessonID || progress.update.ClientSequence() != 7 || !progress.now.Equal(now) {
		t.Fatalf("progress=%+v", progress)
	}
}

func TestHeartbeatRefusalsNeverWriteProgress(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		outcome domain.OwnershipRenewal
		err     error
		want    error
	}{
		{name: "ownership missing", outcome: domain.OwnershipMissing, want: domain.ErrPlaybackSessionExpired},
		{name: "ownership taken over", outcome: domain.OwnershipTakenOver, want: domain.ErrPlaybackTakenOver},
		{name: "redis unavailable", err: domain.ErrDependencyUnavailable, want: domain.ErrDependencyUnavailable},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			ownerships := &ownershipWriter{renewal: test.outcome, renewErr: test.err}
			progress := &progressWriter{}
			service := application.NewSessions(ownerships, progress, 3*time.Minute, time.Now)
			update, err := domain.NewProgressUpdate(42, 7, "tick")
			if err != nil {
				t.Fatal(err)
			}
			err = service.Heartbeat(context.Background(), userID, playbackID, lessonID, update)
			if !errors.Is(err, test.want) || progress.saved {
				t.Fatalf("err=%v saved=%t, want %v and no save", err, progress.saved, test.want)
			}
		})
	}
}

func TestProgressDependencyFailureIsPreserved(t *testing.T) {
	t.Parallel()
	service := application.NewSessions(
		&ownershipWriter{renewal: domain.OwnershipRenewed},
		&progressWriter{err: domain.ErrDependencyUnavailable},
		3*time.Minute,
		time.Now,
	)
	update, err := domain.NewProgressUpdate(42, 7, "tick")
	if err != nil {
		t.Fatal(err)
	}
	if err := service.Heartbeat(context.Background(), userID, playbackID, lessonID, update); !errors.Is(err, domain.ErrDependencyUnavailable) {
		t.Fatalf("got %v, want dependency unavailable", err)
	}
}
