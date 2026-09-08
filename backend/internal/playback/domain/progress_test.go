package domain_test

import (
	"errors"
	"math"
	"testing"
	"time"

	"courses/internal/playback/domain"
)

func TestProgressUpdateRejectsInvalidSequencePositionAndReason(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		position float64
		sequence int64
		reason   string
	}{
		{name: "negative sequence", position: 1, sequence: -1, reason: "tick"},
		{name: "sequence at upper bound", position: 1, sequence: 1 << 31, reason: "tick"},
		{name: "huge sequence", position: 1, sequence: 1 << 62, reason: "tick"},
		{name: "negative position", position: -0.1, sequence: 0, reason: "tick"},
		{name: "non finite position", position: math.Inf(1), sequence: 0, reason: "tick"},
		{name: "unknown reason", position: 1, sequence: 0, reason: "close"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			_, err := domain.NewProgressUpdate(test.position, test.sequence, test.reason)
			if !errors.Is(err, domain.ErrInvalidProgress) {
				t.Fatalf("got %v, want invalid progress", err)
			}
		})
	}
}

func TestProgressUpdateAcceptsClosedReasonsAndSequenceBounds(t *testing.T) {
	t.Parallel()
	for _, reason := range []string{"tick", "pause", "seek", "hide", "nav", "ended"} {
		for _, sequence := range []int64{0, 1<<31 - 1} {
			update, err := domain.NewProgressUpdate(12.5, sequence, reason)
			if err != nil {
				t.Fatalf("reason=%q sequence=%d: %v", reason, sequence, err)
			}
			if update.PositionSeconds() != 12.5 || update.ClientSequence() != sequence || string(update.Reason()) != reason {
				t.Fatalf("update=%+v", update)
			}
		}
	}
}

func TestApplyProgressUsesServerCompletionAndMonotonicityRules(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, time.September, 8, 12, 0, 0, 0, time.UTC)
	duration100 := 100.0
	duration90 := 90.0
	tests := []struct {
		name          string
		saved         domain.Progress
		position      float64
		sequence      int64
		reason        string
		duration      *float64
		wantAccepted  bool
		wantPosition  float64
		wantSequence  int64
		wantCompleted bool
	}{
		{name: "exactly 95 percent completes", saved: domain.Progress{Present: true, PositionSeconds: 90, ClientSequence: 1}, position: 95, sequence: 2, reason: "tick", duration: &duration100, wantAccepted: true, wantPosition: 95, wantSequence: 2, wantCompleted: true},
		{name: "below 95 percent stays incomplete", saved: domain.Progress{Present: true, PositionSeconds: 90, ClientSequence: 1}, position: 94.9, sequence: 2, reason: "tick", duration: &duration100, wantAccepted: true, wantPosition: 94.9, wantSequence: 2},
		{name: "same position after shorter replacement does not complete", saved: domain.Progress{Present: true, PositionSeconds: 94, ClientSequence: 1}, position: 94, sequence: 2, reason: "seek", duration: &duration90, wantAccepted: true, wantPosition: 90, wantSequence: 2},
		{name: "playing to end completes", saved: domain.Progress{Present: true, PositionSeconds: 90, ClientSequence: 1}, position: 100, sequence: 2, reason: "ended", duration: &duration100, wantAccepted: true, wantPosition: 100, wantSequence: 2, wantCompleted: true},
		{name: "unknown duration stores without clamp or completion", saved: domain.Progress{Present: true, PositionSeconds: 2, ClientSequence: 1}, position: 120, sequence: 2, reason: "tick", duration: nil, wantAccepted: true, wantPosition: 120, wantSequence: 2},
		{name: "position beyond duration is clamped", saved: domain.Progress{Present: true, PositionSeconds: 80, ClientSequence: 1}, position: 120, sequence: 2, reason: "tick", duration: &duration90, wantAccepted: true, wantPosition: 90, wantSequence: 2, wantCompleted: true},
		{name: "stale sequence is ignored", saved: domain.Progress{Present: true, PositionSeconds: 80, ClientSequence: 9}, position: 90, sequence: 4, reason: "ended", duration: &duration100, wantPosition: 80, wantSequence: 9},
		{name: "periodic regression is ignored", saved: domain.Progress{Present: true, PositionSeconds: 80, ClientSequence: 9}, position: 20, sequence: 10, reason: "tick", duration: &duration100, wantPosition: 80, wantSequence: 9},
		{name: "periodic position is compared after clamp", saved: domain.Progress{Present: true, PositionSeconds: 94, ClientSequence: 9}, position: 100, sequence: 10, reason: "tick", duration: &duration90, wantPosition: 94, wantSequence: 9},
		{name: "event regression records deliberate seek", saved: domain.Progress{Present: true, PositionSeconds: 80, ClientSequence: 9}, position: 20, sequence: 10, reason: "seek", duration: &duration100, wantAccepted: true, wantPosition: 20, wantSequence: 10},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			update, err := domain.NewProgressUpdate(test.position, test.sequence, test.reason)
			if err != nil {
				t.Fatal(err)
			}
			got, accepted := domain.ApplyProgress(test.saved, update, test.duration, now)
			if accepted != test.wantAccepted || got.PositionSeconds != test.wantPosition || got.ClientSequence != test.wantSequence {
				t.Fatalf("accepted=%t progress=%+v", accepted, got)
			}
			if (got.CompletedAt != nil) != test.wantCompleted {
				t.Fatalf("completedAt=%v, want completed=%t", got.CompletedAt, test.wantCompleted)
			}
			if accepted && !got.UpdatedAt.Equal(now) {
				t.Fatalf("updatedAt=%v, want %v", got.UpdatedAt, now)
			}
		})
	}
}

func TestCompletedProgressNeverLosesItsCompletionMark(t *testing.T) {
	t.Parallel()
	completedAt := time.Date(2026, time.September, 7, 9, 0, 0, 0, time.UTC)
	update, err := domain.NewProgressUpdate(10, 4, "nav")
	if err != nil {
		t.Fatal(err)
	}
	got, accepted := domain.ApplyProgress(domain.Progress{Present: true, PositionSeconds: 100, ClientSequence: 3, CompletedAt: &completedAt}, update, nil, completedAt.Add(time.Hour))
	if accepted || got.PositionSeconds != 100 || got.ClientSequence != 3 || got.CompletedAt == nil || !got.CompletedAt.Equal(completedAt) {
		t.Fatalf("accepted=%t progress=%+v", accepted, got)
	}
}

func TestFirstClientSequenceZeroIsAccepted(t *testing.T) {
	t.Parallel()
	update, err := domain.NewProgressUpdate(1, 0, "tick")
	if err != nil {
		t.Fatal(err)
	}
	got, accepted := domain.ApplyProgress(domain.Progress{}, update, nil, time.Now())
	if !accepted || !got.Present || got.PositionSeconds != 1 || got.ClientSequence != 0 {
		t.Fatalf("accepted=%t progress=%+v", accepted, got)
	}
}
