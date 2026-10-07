package application

import (
	"context"
	"time"

	"courses/internal/transcript/domain"
	"courses/internal/transcript/ports"
)

type Reading struct{ repository ports.TranscriptRepository }

func NewReading(repository ports.TranscriptRepository) *Reading {
	return &Reading{repository: repository}
}

// Transcript is the read model handed to transport: the text plus the derived staleness, so no
// caller recomputes the comparison and no two callers can disagree about it.
type Transcript struct {
	Found    bool
	Stale    bool
	LessonID string
	Language domain.Language

	// UpdatedAt is the token an author must send back when saving. Empty when no transcript.
	UpdatedAt string
	Segments  []domain.Segment

	// Documents are the library entries the segments reference, each carried once.
	Documents []domain.DocumentLink
}

func (reading *Reading) Get(ctx context.Context, actor domain.Actor, lessonID, rawLanguage string) (Transcript, error) {
	if err := domain.ValidateID(lessonID); err != nil {
		return Transcript{}, err
	}
	if err := domain.ValidateID(actor.UserID); err != nil {
		return Transcript{}, domain.ErrForbidden
	}

	var language domain.Language
	if rawLanguage != "" {
		normalised, err := domain.NewLanguage(rawLanguage)
		if err != nil {
			return Transcript{}, err
		}
		language = normalised
	}

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	found, err := reading.repository.GetLessonTranscript(ctx, actor.UserID, lessonID, language)
	if err != nil {
		return Transcript{}, err
	}

	// The access rule runs even when there is no transcript. Otherwise the absence of text would
	// answer differently from its presence, and "no transcript" would quietly confirm that a
	// draft course exists — a transcript must not become a second way into a course.
	if !actor.EmailVerified {
		return Transcript{}, domain.ErrEmailNotVerified
	}
	if !found.Course.CanConsumeBy(actor) {
		if found.Course.PurchaseRequiredBy(actor) {
			return Transcript{}, domain.ErrPurchaseRequired
		}
		return Transcript{}, domain.ErrForbidden
	}
	if !found.Found {
		return Transcript{Found: false}, nil
	}
	return Transcript{
		Found:     true,
		Stale:     found.Transcript.Stale(found.ReadyGeneration),
		LessonID:  found.Transcript.LessonID,
		Language:  found.Transcript.Language,
		UpdatedAt: found.UpdatedAt,
		Segments:  found.Transcript.Segments,
		Documents: found.Transcript.Documents,
	}, nil
}
