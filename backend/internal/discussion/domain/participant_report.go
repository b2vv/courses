package domain

import (
	"strings"
	"time"
	"unicode/utf8"

	"courses/internal/contracts/access"
)

type ReportKind string
type ReportState string

const (
	ReportOnLecturer ReportKind  = "on_lecturer"
	ReportOnStudent  ReportKind  = "on_student"
	ReportOpen       ReportState = "open"
	ReportResolved   ReportState = "resolved"
)

type ParticipantReport struct {
	ID         string
	CourseID   string
	ReporterID string
	TargetID   string
	Kind       ReportKind
	Reason     string
	State      ReportState
	CreatedAt  time.Time
}

func NewParticipantReport(id, courseID, reporterID, targetID string, kind ReportKind, reason string, now time.Time) (ParticipantReport, error) {
	if reporterID == targetID {
		return ParticipantReport{}, ErrReportSelf
	}
	if courseID == "" || reporterID == "" || targetID == "" || (kind != ReportOnLecturer && kind != ReportOnStudent) {
		return ParticipantReport{}, ErrReportInvalid
	}
	reason = strings.TrimSpace(reason)
	if !utf8.ValidString(reason) || utf8.RuneCountInString(reason) < 1 || utf8.RuneCountInString(reason) > 500 {
		return ParticipantReport{}, ErrReportInvalid
	}
	return ParticipantReport{ID: id, CourseID: courseID, ReporterID: reporterID, TargetID: targetID, Kind: kind, Reason: reason, State: ReportOpen, CreatedAt: now}, nil
}

func IsParticipant(membership access.Membership) bool {
	return membership == access.MembershipActive || membership == access.MembershipBlocked
}
