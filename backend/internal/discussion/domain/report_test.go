package domain

import (
	"errors"
	"testing"
)

func TestNewReportRejectsReportsByCommentAuthor(t *testing.T) {
	comment := Comment{ID: "comment", AuthorID: "author"}
	for _, test := range []struct {
		name       string
		reporterID string
		wantErr    error
	}{
		{name: "own comment", reporterID: "author", wantErr: ErrReportOwnComment},
		{name: "other comment", reporterID: "other"},
	} {
		t.Run(test.name, func(t *testing.T) {
			report, err := NewReport(Report{CommentID: comment.ID, ReporterID: test.reporterID}, comment)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("error=%v want=%v", err, test.wantErr)
			}
			if test.wantErr == nil && report.ReporterID != test.reporterID {
				t.Fatalf("reporter=%q want=%q", report.ReporterID, test.reporterID)
			}
		})
	}
}

func TestCourseCanBeModeratedByOwnerOrAdministrator(t *testing.T) {
	course := Course{OwnerID: "owner"}
	for _, test := range []struct {
		name  string
		actor Actor
		want  bool
	}{
		{name: "owner", actor: Actor{UserID: "owner"}, want: true},
		{name: "administrator", actor: Actor{UserID: "admin", Administrator: true}, want: true},
		{name: "stranger", actor: Actor{UserID: "stranger"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := course.CanBeModeratedBy(test.actor); got != test.want {
				t.Fatalf("allowed=%v want=%v", got, test.want)
			}
		})
	}
}

func TestIsModerationDistinguishesAuthorRemoval(t *testing.T) {
	comment := Comment{AuthorID: "author"}
	for _, test := range []struct {
		name  string
		actor Actor
		want  bool
	}{
		{name: "author removal", actor: Actor{UserID: "author"}},
		{name: "other actor removal", actor: Actor{UserID: "moderator"}, want: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := IsModeration(test.actor, comment); got != test.want {
				t.Fatalf("moderation=%v want=%v", got, test.want)
			}
		})
	}
}
