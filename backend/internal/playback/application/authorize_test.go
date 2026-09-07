package application_test

import (
	"context"
	"errors"
	"testing"

	"courses/internal/playback/application"
	"courses/internal/playback/domain"
)

const (
	userID     = "40000000-0000-4000-8000-000000000001"
	playbackID = "11111111111111111111111111111111"
	videoID    = "50000000-0000-4000-8000-000000000001"
	lessonID   = "50000000-0000-4000-8000-000000000002"
)

type ownershipReader struct {
	calls int
	owner domain.Ownership
	err   error
}

func (reader *ownershipReader) Load(_ context.Context, gotUserID string) (domain.Ownership, error) {
	reader.calls++
	if gotUserID != userID {
		return domain.Ownership{}, errors.New("unexpected user")
	}
	return reader.owner, reader.err
}

type accessReader struct {
	calls  int
	access domain.VideoCourse
	err    error
}

func (reader *accessReader) LoadVideoCourse(_ context.Context, gotVideoID string) (domain.VideoCourse, error) {
	reader.calls++
	if gotVideoID != videoID {
		return domain.VideoCourse{}, errors.New("unexpected video")
	}
	return reader.access, reader.err
}

func TestRandomPlaybackIDStopsAfterOwnershipRead(t *testing.T) {
	t.Parallel()
	ownership := &ownershipReader{owner: domain.Ownership{PlaybackID: playbackID, VideoID: videoID, LessonID: lessonID}}
	access := &accessReader{}
	authorizer := application.NewAuthorizer(ownership, access)

	err := authorizer.Authorize(context.Background(), domain.Actor{UserID: userID}, "22222222222222222222222222222222", videoID)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("got %v, want forbidden", err)
	}
	if ownership.calls != 1 || access.calls != 0 {
		t.Fatalf("ownership calls=%d access calls=%d, want 1 and 0", ownership.calls, access.calls)
	}
}

func TestOwnerOfMatchingPlaybackCanViewPublishedCourse(t *testing.T) {
	t.Parallel()
	ownership := &ownershipReader{owner: domain.Ownership{PlaybackID: playbackID, VideoID: videoID, LessonID: lessonID}}
	access := &accessReader{access: domain.VideoCourse{OwnerID: "40000000-0000-4000-8000-000000000002", Status: domain.Published}}
	authorizer := application.NewAuthorizer(ownership, access)

	err := authorizer.Authorize(context.Background(), domain.Actor{UserID: userID}, playbackID, videoID)
	if err != nil {
		t.Fatalf("authorize published course: %v", err)
	}
	if ownership.calls != 1 || access.calls != 1 {
		t.Fatalf("ownership calls=%d access calls=%d, want 1 and 1", ownership.calls, access.calls)
	}
}

func TestCourseVisibilityPolicyRunsOnlyAfterMatchingOwnership(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		actor  domain.Actor
		course domain.VideoCourse
		want   error
	}{
		{name: "student cannot view archived course", actor: domain.Actor{UserID: userID}, course: domain.VideoCourse{OwnerID: "owner", Status: domain.Archived}, want: domain.ErrForbidden},
		{name: "owner can view archived course", actor: domain.Actor{UserID: userID}, course: domain.VideoCourse{OwnerID: userID, Status: domain.Archived}},
		{name: "administrator can view draft course", actor: domain.Actor{UserID: userID, Administrator: true}, course: domain.VideoCourse{OwnerID: "owner", Status: domain.Draft}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			ownership := &ownershipReader{owner: domain.Ownership{PlaybackID: playbackID, VideoID: videoID, LessonID: lessonID}}
			access := &accessReader{access: test.course}
			err := application.NewAuthorizer(ownership, access).Authorize(context.Background(), test.actor, playbackID, videoID)
			if !errors.Is(err, test.want) {
				t.Fatalf("got %v, want %v", err, test.want)
			}
			if ownership.calls != 1 || access.calls != 1 {
				t.Fatalf("ownership calls=%d access calls=%d", ownership.calls, access.calls)
			}
		})
	}
}

func TestDependencyFailuresStayUnavailableAndStopTheSequence(t *testing.T) {
	t.Parallel()
	t.Run("ownership unavailable", func(t *testing.T) {
		ownership := &ownershipReader{err: domain.ErrDependencyUnavailable}
		access := &accessReader{}
		err := application.NewAuthorizer(ownership, access).Authorize(context.Background(), domain.Actor{UserID: userID}, playbackID, videoID)
		if !errors.Is(err, domain.ErrDependencyUnavailable) || ownership.calls != 1 || access.calls != 0 {
			t.Fatalf("err=%v ownership=%d access=%d", err, ownership.calls, access.calls)
		}
	})
	t.Run("access unavailable", func(t *testing.T) {
		ownership := &ownershipReader{owner: domain.Ownership{PlaybackID: playbackID, VideoID: videoID, LessonID: lessonID}}
		access := &accessReader{err: domain.ErrDependencyUnavailable}
		err := application.NewAuthorizer(ownership, access).Authorize(context.Background(), domain.Actor{UserID: userID}, playbackID, videoID)
		if !errors.Is(err, domain.ErrDependencyUnavailable) || ownership.calls != 1 || access.calls != 1 {
			t.Fatalf("err=%v ownership=%d access=%d", err, ownership.calls, access.calls)
		}
	})
}
