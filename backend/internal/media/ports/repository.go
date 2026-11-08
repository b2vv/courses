package ports

import (
	"context"
	"time"

	"courses/internal/media/domain"
)

// ChunkPolicy runs while AppendChunk holds the asset row lock. ErrUnsupportedSource
// is its single terminal error: the repository commits the created asset as failed.
// Every other error rolls the transaction back.
type ChunkPolicy func(domain.VideoAsset) (int64, error)

type CompletePolicy func(domain.VideoAsset) error

type PublishPolicy func(domain.VideoAsset) error

type AssetRepository interface {
	Load(ctx context.Context, id string) (domain.VideoAsset, error)
	ListCourseVideos(ctx context.Context, actor domain.Actor, courseID string) ([]domain.VideoAsset, error)

	Mint(ctx context.Context, actor domain.Actor, lessonID string,
		declaredBytes int64, originalName string, expiresAt, now time.Time) (domain.VideoAsset, error)
	AppendChunk(ctx context.Context, id string, offset int64, now time.Time,
		chunkPolicy ChunkPolicy, completePolicy CompletePolicy) (domain.VideoAsset, error)
	Cancel(ctx context.Context, id string, now time.Time) error

	Claim(ctx context.Context, id string, token domain.ClaimToken, now time.Time) (domain.VideoAsset, bool, error)
	// Publish finishes a successful video transcode. embedded non-nil materialises
	// that embedded audio track and takes the asset to Ready in the same
	// transaction (D4: Ready always has a Ready track); embedded nil means the
	// source carried no audio, so the asset goes to AwaitingAudio and no eviction
	// runs. The returned asset is empty when the claim no longer holds.
	Publish(ctx context.Context, id string, token domain.ClaimToken, renditions []domain.Rendition,
		durationSeconds *float64, now time.Time, embedded *domain.AudioTrack, policy PublishPolicy) (published domain.VideoAsset, evicted *domain.VideoAsset, err error)
	// RegenerateMaster serialises a master rewrite per asset (D7): it locks the
	// asset and its tracks, then calls rewrite with the current Ready track set in
	// deterministic order while holding the lock, so concurrent audio transcodes
	// cannot interleave two writers. rewrite performs the atomic file write.
	RegenerateMaster(ctx context.Context, assetID string, rewrite func(ready []domain.AudioTrack) error) error
	Requeue(ctx context.Context, id string, token domain.ClaimToken, now time.Time) error
	Fail(ctx context.Context, id string, from domain.State, token domain.ClaimToken,
		reason domain.FailureReason, now time.Time) error
	Retry(ctx context.Context, actor domain.Actor, id string, claimAge time.Duration, now time.Time) (domain.VideoAsset, error)

	SweepCandidates(ctx context.Context, now time.Time, limit int) ([]domain.VideoAsset, error)
	DeleteSwept(ctx context.Context, id string) error
	Recover(ctx context.Context, now time.Time, limit int) ([]domain.VideoAsset, error)
}
