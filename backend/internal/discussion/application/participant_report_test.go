package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"courses/internal/contracts/access"
	"courses/internal/discussion/domain"
	"courses/internal/discussion/ports"
)

type participantReportFake struct {
	reviewFake
	ownerID     string
	membership  access.Membership
	courseFound bool
	insertErr   error
	inserted    domain.ParticipantReport
	queued      ports.ParticipantReportQueuePage
	resolved    string
}

func (f *participantReportFake) ReportSubject(context.Context, string, string) (string, access.Membership, bool, error) {
	return f.ownerID, f.membership, f.courseFound, nil
}
func (f *participantReportFake) InsertParticipantReport(_ context.Context, report domain.ParticipantReport) error {
	f.inserted = report
	return f.insertErr
}
func (f *participantReportFake) ParticipantReportQueue(context.Context, string, int) (ports.ParticipantReportQueuePage, error) {
	return f.queued, nil
}
func (f *participantReportFake) ResolveParticipantReport(_ context.Context, id string) error {
	f.resolved = id
	return nil
}

func newParticipantReportService() (*Service, *participantReportFake) {
	f := &participantReportFake{ownerID: "owner", membership: access.MembershipActive, courseFound: true}
	return New(f, func() time.Time { return time.Date(2026, 10, 7, 0, 0, 0, 0, time.UTC) }), f
}

func TestReportLecturer(t *testing.T) {
	for _, tc := range []struct {
		name       string
		verified   bool
		membership access.Membership
		actor      string
		insert     error
		want       error
	}{
		{"active", true, access.MembershipActive, "student", nil, nil}, {"blocked", true, access.MembershipBlocked, "student", nil, nil},
		{"none", true, access.MembershipNone, "student", nil, domain.ErrReportForbidden}, {"owner", true, access.MembershipActive, "owner", nil, domain.ErrReportSelf},
		{"unverified", false, access.MembershipActive, "student", nil, ErrEmailNotVerified}, {"duplicate", true, access.MembershipActive, "student", domain.ErrReportDuplicate, domain.ErrReportDuplicate},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s, f := newParticipantReportService()
			f.membership = tc.membership
			f.insertErr = tc.insert
			report, err := s.ReportLecturer(context.Background(), Actor{UserID: tc.actor, EmailVerified: tc.verified}, "course", "reason")
			if !errors.Is(err, tc.want) {
				t.Fatalf("err=%v want=%v", err, tc.want)
			}
			if tc.want == nil && (report.Kind != domain.ReportOnLecturer || f.inserted.TargetID != "owner") {
				t.Fatalf("report=%+v inserted=%+v", report, f.inserted)
			}
		})
	}
}

func TestReportStudent(t *testing.T) {
	for _, tc := range []struct {
		name, actor string
		admin       bool
		member      access.Membership
		want        error
	}{
		{"owner", "owner", false, access.MembershipActive, nil}, {"admin", "admin", true, access.MembershipBlocked, nil}, {"stranger", "other", false, access.MembershipActive, domain.ErrReportForbidden}, {"nonmember", "owner", false, access.MembershipNone, domain.ErrReportForbidden},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s, f := newParticipantReportService()
			f.membership = tc.member
			report, err := s.ReportStudent(context.Background(), Actor{UserID: tc.actor, Administrator: tc.admin, EmailVerified: true}, "course", "student", "reason")
			if !errors.Is(err, tc.want) {
				t.Fatalf("err=%v want=%v", err, tc.want)
			}
			if tc.want == nil && report.Kind != domain.ReportOnStudent {
				t.Fatalf("report=%+v", report)
			}
		})
	}
}

func TestReportQueueAndResolveAdminOnly(t *testing.T) {
	s, f := newParticipantReportService()
	if _, err := s.ReportQueue(context.Background(), Actor{}, ""); !errors.Is(err, domain.ErrReportForbidden) {
		t.Fatalf("queue err=%v", err)
	}
	if err := s.ResolveReport(context.Background(), Actor{}, "report"); !errors.Is(err, domain.ErrReportForbidden) {
		t.Fatalf("resolve err=%v", err)
	}
	f.queued = ports.ParticipantReportQueuePage{Items: []ports.ParticipantReportQueueItem{{Report: domain.ParticipantReport{ID: "report"}, ReporterName: stringPointer("Reporter"), TargetName: stringPointer("Target"), CourseTitle: "Course"}}}
	page, err := s.ReportQueue(context.Background(), Actor{Administrator: true}, "")
	if err != nil || len(page.Items) != 1 || page.Items[0].ReporterName == nil || *page.Items[0].ReporterName != "Reporter" {
		t.Fatalf("page=%+v err=%v", page, err)
	}
	if err := s.ResolveReport(context.Background(), Actor{Administrator: true}, "report"); err != nil || f.resolved != "report" {
		t.Fatalf("err=%v resolved=%q", err, f.resolved)
	}
}

func stringPointer(value string) *string { return &value }
