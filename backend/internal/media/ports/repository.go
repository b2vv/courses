package ports

import (
	"context"
	"time"

	"courses/internal/media/domain"
)

type AssetRepository interface {
	Mint(context.Context, domain.Actor, string, int64, string, time.Time, time.Time) (domain.VideoAsset, error)
	BeginUpload(context.Context, string, time.Time) (domain.VideoAsset, error)
	Advance(context.Context, string, int64, int64, time.Time) (domain.VideoAsset, error)
	Complete(context.Context, string, time.Time) (domain.VideoAsset, error)
	Claim(context.Context, string, domain.ClaimToken, time.Time) (domain.VideoAsset, error)
	Publish(context.Context, string, domain.ClaimToken, []domain.Rendition, time.Time) (domain.VideoAsset, *domain.VideoAsset, error)
	Requeue(context.Context, string, domain.ClaimToken, time.Time) error
	Fail(context.Context, string, domain.State, domain.ClaimToken, domain.FailureReason, time.Time) error
	Cancel(context.Context, string, time.Time) error
	ListCourseVideos(context.Context, domain.Actor, string) ([]domain.VideoAsset, error)
	Retry(context.Context, domain.Actor, string, time.Duration, time.Time) (domain.VideoAsset, error)
	SweepCandidates(context.Context, time.Time, int) ([]domain.VideoAsset, error)
	Recover(context.Context, time.Time, int) ([]domain.VideoAsset, error)
}
