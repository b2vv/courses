package ports

import (
	"context"
	"time"

	"courses/internal/playback/domain"
)

type LessonPlaybackReader interface {
	LoadLessonPlayback(context.Context, string, string) (domain.LessonPlayback, error)
}

type OwnershipWriter interface {
	Take(context.Context, string, domain.Ownership, time.Duration) error
	Renew(context.Context, string, string, time.Duration) (domain.OwnershipRenewal, error)
}

type ProgressWriter interface {
	Save(context.Context, string, string, domain.ProgressUpdate, time.Time) error
}
