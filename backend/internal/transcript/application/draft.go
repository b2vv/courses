package application

import (
	"context"
	"log/slog"
	"time"

	"courses/internal/transcript/domain"
	"courses/internal/transcript/ports"
)

type CourseReader interface {
	LessonCourse(ctx context.Context, lessonID string) (domain.Course, error)
}

type Drafting struct {
	drafts  ports.DraftRepository
	courses CourseReader
	queue   ports.RecognitionQueue
	now     Clock
	logger  *slog.Logger
}

func NewDrafting(
	drafts ports.DraftRepository,
	courses CourseReader,
	queue ports.RecognitionQueue,
	now Clock,
	logger *slog.Logger,
) *Drafting {
	return &Drafting{drafts: drafts, courses: courses, queue: queue, now: now, logger: logger}
}

type DraftStatus struct {
	Found      bool
	State      domain.DraftState
	Language   domain.Language
	Segments   []domain.Segment
	Failure    string
	Stale      bool
	UpdatedAt  string
	AssetID    string
	Generation int32
}

func (d *Drafting) GetDraft(ctx context.Context, actor domain.Actor, lessonID string) (DraftStatus, error) {
	if err := domain.ValidateID(lessonID); err != nil {
		return DraftStatus{}, err
	}
	if err := domain.ValidateID(actor.UserID); err != nil {
		return DraftStatus{}, domain.ErrForbidden
	}

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	course, err := d.courses.LessonCourse(ctx, lessonID)
	if err != nil {
		return DraftStatus{}, err
	}
	if !course.CanBeEditedBy(actor) {
		return DraftStatus{}, domain.ErrForbidden
	}

	lesson, err := d.drafts.GetLessonDraft(ctx, lessonID)
	if err != nil {
		return DraftStatus{}, err
	}
	if !lesson.Found {
		return DraftStatus{}, nil
	}
	return DraftStatus{
		Found:      true,
		State:      lesson.Draft.State,
		Language:   lesson.Draft.Language,
		Segments:   lesson.Draft.Segments,
		Failure:    lesson.Draft.Failure,
		Stale:      lesson.Draft.Stale(lesson.ReadyGeneration),
		UpdatedAt:  lesson.UpdatedAt,
		AssetID:    lesson.Draft.AssetID,
		Generation: lesson.Draft.Generation,
	}, nil
}

func (d *Drafting) Regenerate(ctx context.Context, actor domain.Actor, lessonID string, language domain.Language) error {
	if err := domain.ValidateID(lessonID); err != nil {
		return err
	}
	if err := domain.ValidateID(actor.UserID); err != nil {
		return domain.ErrForbidden
	}

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	course, err := d.courses.LessonCourse(ctx, lessonID)
	if err != nil {
		return err
	}
	if !course.CanBeEditedBy(actor) {
		return domain.ErrForbidden
	}

	asset, err := d.drafts.LessonAsset(ctx, lessonID)
	if err != nil {
		return err
	}
	if asset == nil {
		return domain.ErrNotFound
	}
	if asset.DurationSeconds == nil {
		return domain.ErrDependencyUnavailable
	}

	timeout := TaskTimeout(*asset.DurationSeconds)
	return d.queue.Enqueue(ctx, ports.RecognitionJob{
		LessonID:   lessonID,
		AssetID:    asset.AssetID,
		Generation: asset.Generation,
		Language:   language,
	}, timeout)
}
