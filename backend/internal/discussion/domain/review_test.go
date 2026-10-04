package domain

import (
	"errors"
	"strings"
	"testing"
)

func TestCanReview(t *testing.T) {
	for _, test := range []struct {
		name  string
		facts ReviewFacts
		actor Actor
		want  error
	}{
		{name: "owner with grant and progress", facts: ReviewFacts{OwnerID: "owner", Status: Published, Paid: true, Entitled: true, HasProgress: true}, actor: Actor{UserID: "owner"}, want: ErrReviewOwnerForbidden},
		{name: "admin paid without grant", facts: ReviewFacts{Status: Published, Paid: true, HasProgress: true}, actor: Actor{UserID: "admin", Administrator: true}, want: ErrReviewPurchaseRequired},
		{name: "admin free with progress", facts: ReviewFacts{Status: Published, HasProgress: true}, actor: Actor{UserID: "admin", Administrator: true}},
		{name: "admin paid with grant", facts: ReviewFacts{Status: Published, Paid: true, Entitled: true, HasProgress: true}, actor: Actor{UserID: "admin", Administrator: true}},
		{name: "preview only watcher", facts: ReviewFacts{Status: Published, Paid: true, HasProgress: true}, actor: Actor{UserID: "watcher"}, want: ErrReviewPurchaseRequired},
		{name: "archived with grant", facts: ReviewFacts{Status: Archived, Paid: true, Entitled: true, HasProgress: true}, actor: Actor{UserID: "student"}},
		{name: "archived free without grant", facts: ReviewFacts{Status: Archived, HasProgress: true}, actor: Actor{UserID: "student"}, want: ErrReviewNotAvailable},
		{name: "draft", facts: ReviewFacts{Status: Draft, HasProgress: true}, actor: Actor{UserID: "student"}, want: ErrReviewNotAvailable},
		{name: "free no progress", facts: ReviewFacts{Status: Published}, actor: Actor{UserID: "student"}, want: ErrReviewNotStarted},
		{name: "anonymous", facts: ReviewFacts{Status: Published, HasProgress: true}, want: ErrForbidden},
	} {
		t.Run(test.name, func(t *testing.T) {
			if err := CanReview(test.facts, test.actor); !errors.Is(err, test.want) {
				t.Fatalf("CanReview() error = %v, want %v", err, test.want)
			}
		})
	}
}

func TestNewReviewContent(t *testing.T) {
	valid := strings.Repeat("a", 20)
	for _, test := range []struct {
		name  string
		stars int
		body  string
		want  error
	}{
		{name: "nineteen runes", stars: 3, body: strings.Repeat("a", 19), want: ErrReviewInvalid},
		{name: "twenty runes", stars: 3, body: valid},
		{name: "two thousand runes", stars: 3, body: strings.Repeat("a", 2000)},
		{name: "two thousand and one runes", stars: 3, body: strings.Repeat("a", 2001), want: ErrReviewInvalid},
		{name: "spaces do not make twenty", stars: 3, body: " " + strings.Repeat("a", 18) + "  ", want: ErrReviewInvalid},
		{name: "zero stars", stars: 0, body: valid, want: ErrReviewInvalid},
		{name: "one star", stars: 1, body: valid},
		{name: "five stars", stars: 5, body: valid},
		{name: "six stars", stars: 6, body: valid, want: ErrReviewInvalid},
		{name: "invalid utf8", stars: 3, body: string([]byte{0xff}) + strings.Repeat("a", 20), want: ErrReviewInvalid},
	} {
		t.Run(test.name, func(t *testing.T) {
			content, err := NewReviewContent(test.stars, test.body)
			if !errors.Is(err, test.want) {
				t.Fatalf("NewReviewContent() error = %v, want %v", err, test.want)
			}
			if test.want == nil && content.Body != strings.TrimSpace(test.body) {
				t.Fatalf("body = %q, want trimmed %q", content.Body, strings.TrimSpace(test.body))
			}
		})
	}
}

func TestReviewProgressPercent(t *testing.T) {
	for _, test := range []struct {
		name      string
		completed int
		total     int
		want      int
		wantErr   error
	}{
		{name: "zero of four", total: 4},
		{name: "one of three", completed: 1, total: 3, want: 33},
		{name: "two of three", completed: 2, total: 3, want: 67},
		{name: "half rounds up", completed: 1, total: 8, want: 13},
		{name: "all complete", completed: 4, total: 4, want: 100},
		{name: "zero total", want: 0},
		{name: "negative completed", completed: -1, wantErr: ErrReviewInvalid},
		{name: "completed exceeds total", completed: 2, total: 1, wantErr: ErrReviewInvalid},
	} {
		t.Run(test.name, func(t *testing.T) {
			progress, err := NewReviewProgress(test.completed, test.total)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("NewReviewProgress() error = %v, want %v", err, test.wantErr)
			}
			if test.wantErr == nil && progress.Percent() != test.want {
				t.Fatalf("Percent() = %d, want %d", progress.Percent(), test.want)
			}
		})
	}
}
