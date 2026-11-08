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

	LoadAudioTrack(ctx context.Context, assetID, trackID string) (domain.AudioTrack, error)
	// PromoteAudioTrack marks the uploaded track Ready, elects a default when the
	// asset has none yet, lifts the asset out of AwaitingAudio once a track is Ready
	// (T11), and regenerates the master from the resulting Ready set — all under the
	// per-asset lock, in one transaction (D4/D7). rewrite does the atomic file write.
	// When the asset becomes Ready it joins its lesson's Ready set, so the same
	// per-lesson eviction Publish runs applies: the returned asset is the evicted
	// oldest generation (nil when nothing was evicted), whose bytes the caller purges.
	PromoteAudioTrack(ctx context.Context, assetID, trackID string, now time.Time, rewrite func(ready []domain.AudioTrack) error) (evicted *domain.VideoAsset, err error)
	// FailAudioTrack marks the track Failed and, when the AwaitingAudio asset is
	// left with no Ready track and nothing still in flight, fails the asset (T12).
	FailAudioTrack(ctx context.Context, assetID, trackID string, now time.Time) error

	// The authoring surface (S5-L1): each verifies the actor may edit the asset's
	// lesson's course before touching a track (ErrCourseForbidden otherwise), a
	// missing or evicted asset being ErrVideoNotFound.
	ListAudioTracks(ctx context.Context, actor domain.Actor, assetID string) ([]domain.AudioTrack, error)
	// SetDefaultAudioTrack makes trackID the sole default; the target must be Ready.
	SetDefaultAudioTrack(ctx context.Context, actor domain.Actor, assetID, trackID string, now time.Time) error
	// UpdateAudioTrackMetadata renames a track; a language clash with another
	// non-failed track of the asset is ErrDuplicateLanguage (D6).
	UpdateAudioTrackMetadata(ctx context.Context, actor domain.Actor, assetID, trackID, language, label string, now time.Time) (domain.AudioTrack, error)
	// DeleteAudioTrack removes a track, re-electing a default and regenerating the
	// master from the remaining Ready set (rewrite) under the lock; it refuses to
	// drop the last Ready track of a Ready asset.
	DeleteAudioTrack(ctx context.Context, actor domain.Actor, assetID, trackID string, now time.Time, rewrite func(ready []domain.AudioTrack) error) error
	Requeue(ctx context.Context, id string, token domain.ClaimToken, now time.Time) error
	Fail(ctx context.Context, id string, from domain.State, token domain.ClaimToken,
		reason domain.FailureReason, now time.Time) error
	Retry(ctx context.Context, actor domain.Actor, id string, claimAge time.Duration, now time.Time) (domain.VideoAsset, error)

	SweepCandidates(ctx context.Context, now time.Time, limit int) ([]domain.VideoAsset, error)
	DeleteSwept(ctx context.Context, id string) error
	Recover(ctx context.Context, now time.Time, limit int) ([]domain.VideoAsset, error)
}
