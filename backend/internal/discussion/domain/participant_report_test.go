package domain

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestNewParticipantReport(t *testing.T) {
	now := time.Date(2026, 10, 7, 12, 0, 0, 0, time.UTC)
	for _, tc := range []struct {
		name, reporter, target string
		kind                   ReportKind
		reason                 string
		want                   error
	}{
		{"happy", "reporter", "target", ReportOnLecturer, "  clear reason  ", nil},
		{"self", "same", "same", ReportOnLecturer, "reason", ErrReportSelf},
		{"bad kind", "reporter", "target", "other", "reason", ErrReportInvalid},
		{"empty reason", "reporter", "target", ReportOnStudent, "  ", ErrReportInvalid},
		{"too long", "reporter", "target", ReportOnStudent, strings.Repeat("a", 501), ErrReportInvalid},
		{"invalid utf8", "reporter", "target", ReportOnStudent, string([]byte{0xff}), ErrReportInvalid},
	} {
		t.Run(tc.name, func(t *testing.T) {
			report, err := NewParticipantReport("", "course", tc.reporter, tc.target, tc.kind, tc.reason, now)
			if !errors.Is(err, tc.want) {
				t.Fatalf("error=%v want=%v", err, tc.want)
			}
			if tc.want == nil && (report.Reason != "clear reason" || report.State != ReportOpen || !report.CreatedAt.Equal(now)) {
				t.Fatalf("report=%+v", report)
			}
		})
	}
}
