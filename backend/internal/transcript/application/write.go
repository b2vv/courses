package application

import (
	"context"
	"time"

	"courses/internal/transcript/domain"
	"courses/internal/transcript/ports"
)

type Authoring struct{ repository ports.TranscriptRepository }

func NewAuthoring(repository ports.TranscriptRepository) *Authoring {
	return &Authoring{repository: repository}
}

type SaveRequest struct {
	LessonID          string
	Language          string
	ExpectedUpdatedAt string
	Segments          []domain.Segment
}

// Save replaces the whole transcript of one lesson in one language.
func (authoring *Authoring) Save(ctx context.Context, actor domain.Actor, request SaveRequest) (string, error) {
	if err := domain.ValidateID(request.LessonID); err != nil {
		return "", err
	}
	if err := domain.ValidateID(actor.UserID); err != nil {
		return "", domain.ErrForbidden
	}
	language, err := domain.NewLanguage(request.Language)
	if err != nil {
		return "", err
	}

	// The whole transcript is validated BEFORE anything is written. Every ceiling and every field
	// rule is the domain's, so the repository never sees a segment the domain would refuse and
	// the database never has to be the one that says no.
	transcript, err := domain.NewTranscript(domain.Transcript{
		LessonID: request.LessonID, Language: language, Segments: request.Segments,
	})
	if err != nil {
		return "", err
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Authorization reads the course directly rather than through the transcript, so that writing
	// the first transcript of a lesson is refused exactly as writing the tenth is.
	course, err := authoring.repository.LessonCourse(ctx, request.LessonID)
	if err != nil {
		return "", err
	}
	if !course.CanBeEditedBy(actor) {
		return "", domain.ErrForbidden
	}

	return authoring.repository.SaveLessonTranscript(ctx, ports.SaveTranscript{
		LessonID: transcript.LessonID, Language: transcript.Language,
		Segments: transcript.Segments, ExpectedUpdatedAt: request.ExpectedUpdatedAt,
	})
}
