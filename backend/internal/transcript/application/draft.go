package application

import (
	"cmp"
	"context"
	"log/slog"
	"slices"
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

// ListLessonDrafts returns current drafts in the editor's tab order (MS-D6).
func (d *Drafting) ListLessonDrafts(ctx context.Context, actor domain.Actor, lessonID string) ([]ports.LessonDraft, error) {
	if err := domain.ValidateID(lessonID); err != nil {
		return nil, err
	}
	if err := domain.ValidateID(actor.UserID); err != nil {
		return nil, domain.ErrForbidden
	}
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	course, err := d.courses.LessonCourse(ctx, lessonID)
	if err != nil {
		return nil, err
	}
	if !course.CanBeEditedBy(actor) {
		return nil, domain.ErrForbidden
	}
	drafts, err := d.drafts.ListLessonDrafts(ctx, lessonID)
	if err != nil {
		return nil, err
	}
	slices.SortFunc(drafts, func(a, b ports.LessonDraft) int {
		rank := func(d ports.LessonDraft) int {
			if d.IsDefault {
				return 0
			}
			if d.IsEmbedded {
				return 1
			}
			return 2
		}
		if c := cmp.Compare(rank(a), rank(b)); c != 0 {
			return c
		}
		if c := cmp.Compare(a.TrackLanguage, b.TrackLanguage); c != 0 {
			return c
		}
		return cmp.Compare(a.Draft.TrackID, b.Draft.TrackID)
	})
	return drafts, nil
}

func (d *Drafting) Regenerate(ctx context.Context, actor domain.Actor, lessonID, trackID string) error {
	if err := domain.ValidateID(lessonID); err != nil {
		return err
	}
	if err := domain.ValidateID(trackID); err != nil {
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

	asset, err := d.drafts.LessonTrack(ctx, lessonID, trackID)
	if err != nil {
		return err
	}
	if asset == nil {
		return domain.ErrTrackNotFound
	}
	if asset.DurationSeconds == nil {
		return domain.ErrDependencyUnavailable
	}

	drafts, err := d.drafts.ListLessonDrafts(ctx, lessonID)
	if err != nil {
		return err
	}
	var marker string
	for _, draft := range drafts {
		if draft.Draft.TrackID != asset.TrackID {
			continue
		}
		if draft.Draft.State == domain.DraftRunning {
			return domain.ErrAlreadyRunning
		}
		marker = draft.UpdatedAt
		break
	}

	timeout := TaskTimeout(*asset.DurationSeconds)
	return d.queue.Enqueue(ctx, ports.RecognitionJob{
		LessonID:   lessonID,
		AssetID:    asset.AssetID,
		Generation: asset.Generation,
		TrackID:    asset.TrackID,
		TaskMarker: marker,
		// The author asked: this job may replace the track's existing draft.
		Requested: true,
	}, timeout)
}
