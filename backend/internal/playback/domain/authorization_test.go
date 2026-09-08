package domain_test

import (
	"testing"

	"courses/internal/playback/domain"
	"courses/internal/testsupport/coursevisibility"
)

const (
	visibilityOwnerID  = "40000000-0000-4000-8000-000000000002"
	visibilityViewerID = "40000000-0000-4000-8000-000000000001"
)

func TestO4PlaybackVisibilityUsesSharedCourseCases(t *testing.T) {
	for _, test := range coursevisibility.Cases() {
		t.Run(test.Name, func(t *testing.T) {
			actor := domain.Actor{UserID: visibilityViewerID, Administrator: test.Actor == coursevisibility.Admin}
			if test.Actor == coursevisibility.Owner {
				actor.UserID = visibilityOwnerID
			}
			course := domain.VideoCourse{OwnerID: visibilityOwnerID, Status: domain.CourseStatus(test.Status)}
			if got := course.CanBeViewedBy(actor); got != test.Want {
				t.Fatalf("playback=%t want=%t", got, test.Want)
			}
		})
	}
}
