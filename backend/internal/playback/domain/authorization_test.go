package domain_test

import (
	"testing"

	"courses/internal/contracts/access"
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

func TestCanConsumeByMirrorsMembershipContentRule(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name   string
		actor  domain.Actor
		course domain.VideoCourse
		want   bool
	}{
		{name: "active member consumes", actor: domain.Actor{UserID: visibilityViewerID}, course: domain.VideoCourse{Status: domain.Draft, Membership: access.MembershipActive}, want: true},
		{name: "blocked member consumes", actor: domain.Actor{UserID: visibilityViewerID}, course: domain.VideoCourse{Status: domain.Draft, Membership: access.MembershipBlocked}, want: true},
		{name: "revoked member cannot consume", actor: domain.Actor{UserID: visibilityViewerID}, course: domain.VideoCourse{Status: domain.Published, Membership: access.MembershipRevoked}},
		{name: "published preview consumes", actor: domain.Actor{UserID: visibilityViewerID}, course: domain.VideoCourse{Status: domain.Published, Preview: true}, want: true},
		{name: "draft preview cannot consume", actor: domain.Actor{UserID: visibilityViewerID}, course: domain.VideoCourse{Status: domain.Draft, Preview: true}},
		{name: "owner consumes", actor: domain.Actor{UserID: visibilityOwnerID}, course: domain.VideoCourse{OwnerID: visibilityOwnerID, Status: domain.Draft}, want: true},
		{name: "administrator consumes", actor: domain.Actor{UserID: visibilityViewerID, Administrator: true}, course: domain.VideoCourse{Status: domain.Draft}, want: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := test.course.CanConsumeBy(test.actor); got != test.want {
				t.Fatalf("CanConsumeBy() = %v, want %v", got, test.want)
			}
		})
	}
}
