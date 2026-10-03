package domain

import (
	"strings"
	"unicode/utf8"
)

const (
	MinimumReviewRunes = 20
	MaximumReviewRunes = 2000
)

type ReviewFacts struct {
	OwnerID     string
	Status      Status
	Paid        bool
	Entitled    bool
	HasProgress bool
}

func CanReview(facts ReviewFacts, actor Actor) error {
	if actor.UserID == "" {
		return ErrForbidden
	}
	if actor.UserID == facts.OwnerID {
		return ErrReviewOwnerForbidden
	}
	if facts.Status != Published && !facts.Entitled {
		return ErrReviewNotAvailable
	}
	if facts.Paid && !facts.Entitled {
		return ErrReviewPurchaseRequired
	}
	if !facts.HasProgress {
		return ErrReviewNotStarted
	}
	return nil
}

type ReviewContent struct {
	Stars int
	Body  string
}

func NewReviewContent(stars int, body string) (ReviewContent, error) {
	body = strings.TrimSpace(body)
	if stars < 1 || stars > 5 || !utf8.ValidString(body) {
		return ReviewContent{}, ErrReviewInvalid
	}
	runes := utf8.RuneCountInString(body)
	if runes < MinimumReviewRunes || runes > MaximumReviewRunes {
		return ReviewContent{}, ErrReviewInvalid
	}
	return ReviewContent{Stars: stars, Body: body}, nil
}

type ReviewProgress struct {
	Completed int
	Total     int
}

func NewReviewProgress(completed, total int) (ReviewProgress, error) {
	if completed < 0 || total < completed {
		return ReviewProgress{}, ErrReviewInvalid
	}
	return ReviewProgress{Completed: completed, Total: total}, nil
}

func (progress ReviewProgress) Percent() int {
	if progress.Total == 0 {
		return 0
	}
	return (progress.Completed*100 + progress.Total/2) / progress.Total
}
