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

type playbackReader struct {
	playback domain.LessonPlayback
	err      error
	calls    int
	userID   string
	lessonID string
}

func (reader *playbackReader) LoadLessonPlayback(_ context.Context, userID, lessonID string) (domain.LessonPlayback, error) {
	reader.calls++
	reader.userID, reader.lessonID = userID, lessonID
	return reader.playback, reader.err
}

func viewablePlaybackReader() *playbackReader {
	return &playbackReader{playback: domain.LessonPlayback{
		VideoID: videoID,
		Course:  domain.VideoCourse{OwnerID: "owner", Status: domain.Published},
	}}
}

func (writer *progressWriter) Save(_ context.Context, userID, lessonID string, update domain.ProgressUpdate, now time.Time) error {
	writer.saved, writer.userID, writer.lessonID, writer.update, writer.now = true, userID, lessonID, update, now
	return writer.err
}

func TestTakeUnconditionallyStoresTheNewOwnership(t *testing.T) {
	t.Parallel()
	ownerships := &ownershipWriter{}
	progress := &progressWriter{}
	service := application.NewSessions(nil, ownerships, progress, 3*time.Minute, time.Now)
	want := domain.Ownership{PlaybackID: playbackID, VideoID: videoID, LessonID: lessonID}

	if err := service.Take(context.Background(), userID, want); err != nil {
		t.Fatal(err)
	}
	if ownerships.takeUserID != userID || ownerships.taken != want || ownerships.takeTTL != 3*time.Minute {
		t.Fatalf("user=%q ownership=%+v ttl=%v", ownerships.takeUserID, ownerships.taken, ownerships.takeTTL)
	}
}

func TestPlayChecksLessonAndPolicyBeforeTakingOwnership(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		playback domain.LessonPlayback
		loadErr  error
		actor    domain.Actor
		wantErr  error
		wantTake bool
	}{
		{name: "missing lesson", loadErr: domain.ErrLessonNotFound, actor: domain.Actor{UserID: userID}, wantErr: domain.ErrLessonNotFound},
		{name: "video not ready", loadErr: domain.ErrVideoNotReady, actor: domain.Actor{UserID: userID}, wantErr: domain.ErrVideoNotReady},
		{name: "draft denied to student", playback: domain.LessonPlayback{VideoID: videoID, Course: domain.VideoCourse{OwnerID: "owner", Status: domain.Draft}}, actor: domain.Actor{UserID: userID}, wantErr: domain.ErrForbidden},
		{name: "paid course requires purchase", playback: domain.LessonPlayback{VideoID: videoID, Course: domain.VideoCourse{OwnerID: "owner", Status: domain.Published, Paid: true}}, actor: domain.Actor{UserID: userID}, wantErr: domain.ErrPurchaseRequired},
		{name: "published allowed", playback: domain.LessonPlayback{VideoID: videoID, ClientSequence: 12, Course: domain.VideoCourse{OwnerID: "owner", Status: domain.Published}}, actor: domain.Actor{UserID: userID}, wantTake: true},
		{name: "draft allowed to owner", playback: domain.LessonPlayback{VideoID: videoID, Course: domain.VideoCourse{OwnerID: userID, Status: domain.Draft}}, actor: domain.Actor{UserID: userID}, wantTake: true},
		{name: "archived allowed to admin", playback: domain.LessonPlayback{VideoID: videoID, Course: domain.VideoCourse{OwnerID: "owner", Status: domain.Archived}}, actor: domain.Actor{UserID: userID, Administrator: true}, wantTake: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			reader := &playbackReader{playback: test.playback, err: test.loadErr}
			ownerships := &ownershipWriter{}
			service := application.NewSessions(reader, ownerships, &progressWriter{}, 3*time.Minute, time.Now)

			got, err := service.Play(context.Background(), test.actor, lessonID, playbackID)

			if !errors.Is(err, test.wantErr) {
				t.Fatalf("err=%v, want %v", err, test.wantErr)
			}
			if reader.calls != 1 || reader.userID != test.actor.UserID || reader.lessonID != lessonID {
				t.Fatalf("lookup calls=%d user=%q lesson=%q", reader.calls, reader.userID, reader.lessonID)
			}
			if (ownerships.takeUserID != "") != test.wantTake {
				t.Fatalf("take user=%q, wantTake=%t", ownerships.takeUserID, test.wantTake)
			}
			if test.wantTake {
				wantOwnership := domain.Ownership{PlaybackID: playbackID, VideoID: test.playback.VideoID, LessonID: lessonID}
				if ownerships.taken != wantOwnership || got != test.playback {
					t.Fatalf("ownership=%+v playback=%+v", ownerships.taken, got)
				}
			}
		})
	}
}

func TestHeartbeatRenewsBeforeSavingProgress(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, time.September, 8, 13, 0, 0, 0, time.UTC)
	ownerships := &ownershipWriter{renewal: domain.OwnershipRenewed}
	progress := &progressWriter{}
	service := application.NewSessions(viewablePlaybackReader(), ownerships, progress, 3*time.Minute, func() time.Time { return now })
	update, err := domain.NewProgressUpdate(42, 7, "pause")
	if err != nil {
		t.Fatal(err)
	}

	if err := service.Heartbeat(context.Background(), domain.Actor{UserID: userID}, playbackID, lessonID, update); err != nil {
		t.Fatal(err)
	}
	if ownerships.renewUserID != userID || ownerships.renewPlayback != playbackID || ownerships.renewTTL != 3*time.Minute {
		t.Fatalf("renew user=%q playback=%q ttl=%v", ownerships.renewUserID, ownerships.renewPlayback, ownerships.renewTTL)
	}
	if !progress.saved || progress.userID != userID || progress.lessonID != lessonID || progress.update.ClientSequence() != 7 || !progress.now.Equal(now) {
		t.Fatalf("progress=%+v", progress)
	}
}

func TestR17HeartbeatChecksCourseVisibilityBeforeRenewalAndProgress(t *testing.T) {
	t.Parallel()
	reader := &playbackReader{playback: domain.LessonPlayback{
		VideoID: videoID,
		Course:  domain.VideoCourse{OwnerID: "owner", Status: domain.Archived},
	}}
	ownerships := &ownershipWriter{renewal: domain.OwnershipRenewed}
	progress := &progressWriter{}
	service := application.NewSessions(reader, ownerships, progress, 3*time.Minute, time.Now)
	update, err := domain.NewProgressUpdate(42, 7, "tick")
	if err != nil {
		t.Fatal(err)
	}

	err = service.Heartbeat(context.Background(), domain.Actor{UserID: userID}, playbackID, lessonID, update)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("error=%v", err)
	}
	if reader.calls != 1 || ownerships.renewUserID != "" || progress.saved {
		t.Fatalf("reader calls=%d renew user=%q progress saved=%t", reader.calls, ownerships.renewUserID, progress.saved)
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
			service := application.NewSessions(viewablePlaybackReader(), ownerships, progress, 3*time.Minute, time.Now)
			update, err := domain.NewProgressUpdate(42, 7, "tick")
			if err != nil {
				t.Fatal(err)
			}
			err = service.Heartbeat(context.Background(), domain.Actor{UserID: userID}, playbackID, lessonID, update)
			if !errors.Is(err, test.want) || progress.saved {
				t.Fatalf("err=%v saved=%t, want %v and no save", err, progress.saved, test.want)
			}
		})
	}
}

func TestProgressDependencyFailureIsPreserved(t *testing.T) {
	t.Parallel()
	service := application.NewSessions(
		viewablePlaybackReader(),
		&ownershipWriter{renewal: domain.OwnershipRenewed},
		&progressWriter{err: domain.ErrDependencyUnavailable},
		3*time.Minute,
		time.Now,
	)
	update, err := domain.NewProgressUpdate(42, 7, "tick")
	if err != nil {
		t.Fatal(err)
	}
	if err := service.Heartbeat(context.Background(), domain.Actor{UserID: userID}, playbackID, lessonID, update); !errors.Is(err, domain.ErrDependencyUnavailable) {
		t.Fatalf("got %v, want dependency unavailable", err)
	}
}
