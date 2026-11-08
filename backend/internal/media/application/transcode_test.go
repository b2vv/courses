package application_test

import (
	"context"
	"errors"
	"syscall"
	"testing"
	"time"

	"courses/internal/media/application"
	"courses/internal/media/domain"
	"courses/internal/media/ports"
)

type fakeTranscoder struct {
	calls            []string
	probe            func(context.Context, domain.StorageKey) (domain.SourceProbe, error)
	render           func(context.Context, domain.StorageKey, domain.StorageKey, []domain.Rendition, bool) error
	renderAudio      func(context.Context, domain.StorageKey, domain.StorageKey) error
	regenerateMaster func(context.Context, string, []ports.AudioRendition) error
	thumbnail        func(context.Context, domain.StorageKey, domain.StorageKey) error
}

func (transcoder *fakeTranscoder) Probe(ctx context.Context, source domain.StorageKey) (domain.SourceProbe, error) {
	transcoder.calls = append(transcoder.calls, "probe")
	return transcoder.probe(ctx, source)
}

func (transcoder *fakeTranscoder) Render(ctx context.Context, source, work domain.StorageKey, renditions []domain.Rendition, hasAudio bool) error {
	transcoder.calls = append(transcoder.calls, "render")
	return transcoder.render(ctx, source, work, renditions, hasAudio)
}

func (transcoder *fakeTranscoder) RenderAudioTrack(ctx context.Context, source, work domain.StorageKey) error {
	transcoder.calls = append(transcoder.calls, "render_audio")
	if transcoder.renderAudio != nil {
		return transcoder.renderAudio(ctx, source, work)
	}
	return nil
}

func (transcoder *fakeTranscoder) RegenerateMaster(ctx context.Context, assetID string, audios []ports.AudioRendition) error {
	transcoder.calls = append(transcoder.calls, "regenerate_master")
	if transcoder.regenerateMaster != nil {
		return transcoder.regenerateMaster(ctx, assetID, audios)
	}
	return nil
}

func (transcoder *fakeTranscoder) Thumbnail(ctx context.Context, source, destination domain.StorageKey, probe domain.SourceProbe) error {
	transcoder.calls = append(transcoder.calls, "thumbnail")
	return transcoder.thumbnail(ctx, source, destination)
}

func TestProcessClaimsRendersAndPublishes(t *testing.T) {
	t.Parallel()
	repository := processingRepository(domain.Queued)
	storage := defaultStorage()
	transcoder := defaultTranscoder()
	transcoder.render = func(_ context.Context, source, work domain.StorageKey, renditions []domain.Rendition, _ bool) error {
		if source.Kind != domain.StorageSource || source.ID != mediaAssetID || source.Leaf != "original" {
			t.Fatalf("source=%+v", source)
		}
		if work.Kind != domain.StorageTemporary || work.ID != mediaAssetID || work.Leaf == "" {
			t.Fatalf("work=%+v", work)
		}
		if len(renditions) != 2 || renditions[0] != domain.Rendition360p || renditions[1] != domain.Rendition720p {
			t.Fatalf("renditions=%v", renditions)
		}
		return nil
	}

	outcome, err := newTranscode(repository, storage, transcoder, nil).Process(context.Background(), ports.Job{AssetID: mediaAssetID, MaxRetry: 3})
	if err != nil || outcome != (application.TranscodeOutcome{}) {
		t.Fatalf("outcome=%+v err=%v", outcome, err)
	}
	if repository.asset.State != domain.Ready || repository.asset.ClaimToken != "" {
		t.Fatalf("asset=%+v", repository.asset)
	}
	if repository.publishedDuration == nil || *repository.publishedDuration != 125.5 {
		t.Fatalf("published duration=%v", repository.publishedDuration)
	}
	assertStrings(t, transcoder.calls, []string{"probe", "render", "thumbnail"})
	if countCall(repository.calls, "claim") != 1 || countCall(repository.calls, "publish") != 1 || countCall(repository.calls, "requeue") != 0 || countCall(repository.calls, "fail") != 0 {
		t.Fatalf("repository calls=%v", repository.calls)
	}
	if countString(storage.calls, "publish") != 1 || countString(storage.calls, "move") != 1 {
		t.Fatalf("storage calls=%v", storage.calls)
	}
}

func TestProcessEmbedsAudioTrackWhenSourceHasAudio(t *testing.T) {
	t.Parallel()
	repository := processingRepository(domain.Queued)
	transcoder := defaultTranscoder()
	var renderedAudio bool
	transcoder.render = func(_ context.Context, _, _ domain.StorageKey, _ []domain.Rendition, hasAudio bool) error {
		renderedAudio = hasAudio
		return nil
	}

	outcome, err := newTranscode(repository, defaultStorage(), transcoder, nil).Process(context.Background(), ports.Job{AssetID: mediaAssetID, MaxRetry: 3})
	if err != nil || outcome != (application.TranscodeOutcome{}) {
		t.Fatalf("outcome=%+v err=%v", outcome, err)
	}
	if !renderedAudio {
		t.Fatal("render was told the source has no audio")
	}
	if repository.asset.State != domain.Ready {
		t.Fatalf("asset state=%s want ready", repository.asset.State)
	}
	embedded := repository.publishedEmbedded
	if embedded == nil {
		t.Fatal("no embedded track materialised for an audio-carrying source")
	}
	if embedded.Kind != domain.EmbeddedTrack || embedded.State != domain.Ready || !embedded.IsDefault ||
		embedded.SourceLeaf != domain.EmbeddedAudioLeaf || embedded.VideoAssetID != mediaAssetID {
		t.Fatalf("embedded=%+v", embedded)
	}
	if _, err := domain.NewEmbeddedTrack(*embedded); err != nil {
		t.Fatalf("embedded track fails domain validation: %v", err)
	}
}

func TestProcessVideoOnlySourceAwaitsAudio(t *testing.T) {
	t.Parallel()
	repository := processingRepository(domain.Queued)
	transcoder := defaultTranscoder()
	transcoder.probe = func(context.Context, domain.StorageKey) (domain.SourceProbe, error) {
		return domain.SourceProbe{Width: 1280, Height: 720, VideoCodec: domain.CodecH264, DurationSeconds: 125.5}, nil
	}
	var renderedAudio bool
	transcoder.render = func(_ context.Context, _, _ domain.StorageKey, _ []domain.Rendition, hasAudio bool) error {
		renderedAudio = hasAudio
		return nil
	}

	outcome, err := newTranscode(repository, defaultStorage(), transcoder, nil).Process(context.Background(), ports.Job{AssetID: mediaAssetID, MaxRetry: 3})
	if err != nil || outcome != (application.TranscodeOutcome{}) {
		t.Fatalf("outcome=%+v err=%v", outcome, err)
	}
	if renderedAudio {
		t.Fatal("render mapped audio for a video-only source")
	}
	if repository.asset.State != domain.AwaitingAudio {
		t.Fatalf("asset state=%s want awaiting_audio", repository.asset.State)
	}
	if repository.publishedEmbedded != nil {
		t.Fatalf("video-only source materialised an embedded track: %+v", repository.publishedEmbedded)
	}
}

func TestProcessAudioTrackPublishesPromotesAndRegenerates(t *testing.T) {
	t.Parallel()
	const trackID = "99999999-9999-4999-8999-999999999999"
	repository := &fakeRepository{}
	repository.loadAudioTrack = func(_ context.Context, assetID, id string) (domain.AudioTrack, error) {
		if assetID != mediaAssetID || id != trackID {
			t.Fatalf("load audio track asset=%s track=%s", assetID, id)
		}
		return domain.AudioTrack{ID: trackID, VideoAssetID: mediaAssetID, Kind: domain.UploadedTrack,
			Language: "en", Label: "English", State: domain.Queued, SourceLeaf: domain.AudioLeafPrefix + trackID}, nil
	}
	var rewroteWithReady []domain.AudioTrack
	repository.promoteAudioTrack = func(_ context.Context, _, _ string, _ time.Time, rewrite func([]domain.AudioTrack) error) error {
		rewroteWithReady = []domain.AudioTrack{{ID: trackID, Kind: domain.UploadedTrack, State: domain.Ready, IsDefault: true, Label: "English", Language: "en"}}
		return rewrite(rewroteWithReady)
	}
	transcoder := defaultTranscoder()
	var regeneratedFor string
	transcoder.regenerateMaster = func(_ context.Context, assetID string, audios []ports.AudioRendition) error {
		regeneratedFor = assetID
		if len(audios) != 1 || audios[0].TrackID != trackID || !audios[0].Default || audios[0].Embedded {
			t.Fatalf("regenerate audios=%+v", audios)
		}
		return nil
	}
	var audioSource domain.StorageKey
	transcoder.renderAudio = func(_ context.Context, source, _ domain.StorageKey) error {
		audioSource = source
		return nil
	}

	outcome, err := newTranscode(repository, defaultStorage(), transcoder, nil).Process(
		context.Background(), ports.Job{AssetID: mediaAssetID, TrackID: trackID, MaxRetry: 3})
	if err != nil || outcome != (application.TranscodeOutcome{}) {
		t.Fatalf("outcome=%+v err=%v", outcome, err)
	}
	if audioSource.Kind != domain.StorageSource || audioSource.Leaf != domain.AudioLeafPrefix+trackID {
		t.Fatalf("audio source key=%+v", audioSource)
	}
	if regeneratedFor != mediaAssetID || rewroteWithReady == nil {
		t.Fatalf("regenerated=%q rewrote=%v", regeneratedFor, rewroteWithReady)
	}
	assertStrings(t, transcoder.calls, []string{"probe", "render_audio", "regenerate_master"})
	if countCall(repository.calls, "promote_audio_track") != 1 || countCall(repository.calls, "fail_audio_track") != 0 {
		t.Fatalf("repository calls=%v", repository.calls)
	}
}

func TestProcessAudioTrackFailsPermanentlyOnUnsupportedSource(t *testing.T) {
	t.Parallel()
	const trackID = "99999999-9999-4999-8999-999999999999"
	repository := &fakeRepository{}
	repository.loadAudioTrack = func(context.Context, string, string) (domain.AudioTrack, error) {
		return domain.AudioTrack{ID: trackID, VideoAssetID: mediaAssetID, Kind: domain.UploadedTrack,
			Language: "en", Label: "English", State: domain.Queued, SourceLeaf: domain.AudioLeafPrefix + trackID}, nil
	}
	failed := false
	repository.failAudioTrack = func(_ context.Context, _, id string, _ time.Time) error {
		failed = id == trackID
		return nil
	}
	transcoder := defaultTranscoder()
	transcoder.probe = func(context.Context, domain.StorageKey) (domain.SourceProbe, error) {
		return domain.SourceProbe{DurationSeconds: 3}, nil // no audio codec
	}

	outcome, err := newTranscode(repository, defaultStorage(), transcoder, nil).Process(
		context.Background(), ports.Job{AssetID: mediaAssetID, TrackID: trackID, MaxRetry: 3})
	if err != nil {
		t.Fatal(err)
	}
	if !outcome.Failed || !outcome.Permanent || outcome.Reason != domain.UnsupportedSource {
		t.Fatalf("outcome=%+v", outcome)
	}
	if !failed {
		t.Fatal("track was not failed")
	}
	if countCall(repository.calls, "promote_audio_track") != 0 {
		t.Fatalf("promote should not run on unsupported source: %v", repository.calls)
	}
}

func TestProcessAudioTrackIsIdempotentOnAReadyTrack(t *testing.T) {
	t.Parallel()
	const trackID = "99999999-9999-4999-8999-999999999999"
	repository := &fakeRepository{}
	repository.loadAudioTrack = func(context.Context, string, string) (domain.AudioTrack, error) {
		return domain.AudioTrack{ID: trackID, VideoAssetID: mediaAssetID, Kind: domain.UploadedTrack,
			Language: "en", Label: "English", State: domain.Ready, IsDefault: true, SourceLeaf: domain.AudioLeafPrefix + trackID}, nil
	}
	transcoder := defaultTranscoder()

	outcome, err := newTranscode(repository, defaultStorage(), transcoder, nil).Process(
		context.Background(), ports.Job{AssetID: mediaAssetID, TrackID: trackID, MaxRetry: 3})
	if err != nil || outcome != (application.TranscodeOutcome{}) {
		t.Fatalf("outcome=%+v err=%v", outcome, err)
	}
	if len(transcoder.calls) != 0 || countCall(repository.calls, "promote_audio_track") != 0 {
		t.Fatalf("a ready track must not re-transcode: transcoder=%v repo=%v", transcoder.calls, repository.calls)
	}
}

func TestProcessFinalAndNonFinalTransientFailures(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name       string
		retryCount int
		wantState  domain.State
		wantCall   string
	}{
		{name: "non-final requeues", retryCount: 2, wantState: domain.Queued, wantCall: "requeue"},
		{name: "final fails", retryCount: 3, wantState: domain.Failed, wantCall: "fail"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			repository := processingRepository(domain.Queued)
			transcoder := defaultTranscoder()
			transcoder.render = func(context.Context, domain.StorageKey, domain.StorageKey, []domain.Rendition, bool) error {
				return errors.New("ffmpeg stderr with /private/media/source")
			}
			outcome, err := newTranscode(repository, defaultStorage(), transcoder, nil).Process(context.Background(), ports.Job{
				AssetID: mediaAssetID, RetryCount: test.retryCount, MaxRetry: 3,
			})
			if err != nil || !outcome.Failed || outcome.Permanent || outcome.Reason != domain.RenderFailed {
				t.Fatalf("outcome=%+v err=%v", outcome, err)
			}
			if repository.asset.State != test.wantState || repository.asset.FailureReason != map[bool]domain.FailureReason{true: domain.RenderFailed}[test.wantState == domain.Failed] {
				t.Fatalf("asset=%+v", repository.asset)
			}
			if countCall(repository.calls, test.wantCall) != 1 {
				t.Fatalf("calls=%v", repository.calls)
			}
			if string(outcome.Reason) == "ffmpeg stderr with /private/media/source" {
				t.Fatal("raw transcoder detail escaped")
			}
		})
	}
}

func TestMalformedSourceFailsPermanentlyWithoutRendering(t *testing.T) {
	t.Parallel()
	repository := processingRepository(domain.Queued)
	transcoder := defaultTranscoder()
	transcoder.probe = func(context.Context, domain.StorageKey) (domain.SourceProbe, error) {
		return domain.SourceProbe{Width: 1280, Height: 720, VideoCodec: "vp9", AudioCodec: domain.CodecAAC}, nil
	}
	outcome, err := newTranscode(repository, defaultStorage(), transcoder, nil).Process(context.Background(), ports.Job{AssetID: mediaAssetID, MaxRetry: 3})
	if err != nil || !outcome.Failed || !outcome.Permanent || outcome.Reason != domain.UnsupportedSource {
		t.Fatalf("outcome=%+v err=%v", outcome, err)
	}
	if repository.asset.State != domain.Failed || repository.asset.FailureReason != domain.UnsupportedSource {
		t.Fatalf("asset=%+v", repository.asset)
	}
	assertStrings(t, transcoder.calls, []string{"probe"})
}

func TestProcessNoOpsExhaustiveFinishedAndAbandonedStates(t *testing.T) {
	t.Parallel()
	evictedAt := time.Now()
	for _, test := range []struct {
		name  string
		asset domain.VideoAsset
		err   error
	}{
		{name: "ready", asset: mediaAsset(domain.Ready)},
		{name: "failed", asset: mediaAsset(domain.Failed)},
		{name: "evicted", asset: func() domain.VideoAsset {
			asset := mediaAsset(domain.Queued)
			asset.EvictedAt = &evictedAt
			return asset
		}()},
		{name: "orphan", asset: func() domain.VideoAsset { asset := mediaAsset(domain.Queued); asset.LessonID = ""; return asset }()},
		{name: "created", asset: mediaAsset(domain.Created)},
		{name: "uploading", asset: mediaAsset(domain.Uploading)},
		{name: "absent", err: domain.ErrVideoNotFound},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			repository := processingRepository(domain.Queued)
			repository.asset = test.asset
			repository.load = func(context.Context, string) (domain.VideoAsset, error) { return repository.asset, test.err }
			storage := defaultStorage()
			transcoder := defaultTranscoder()
			outcome, err := newTranscode(repository, storage, transcoder, nil).Process(context.Background(), ports.Job{AssetID: mediaAssetID, MaxRetry: 3})
			if err != nil || outcome != (application.TranscodeOutcome{}) {
				t.Fatalf("outcome=%+v err=%v", outcome, err)
			}
			if countCall(repository.calls, "claim") != 0 || countCall(repository.calls, "publish") != 0 || len(transcoder.calls) != 0 || countString(storage.calls, "publish") != 0 {
				t.Fatalf("repository=%v transcoder=%v storage=%v", repository.calls, transcoder.calls, storage.calls)
			}
		})
	}
}

func TestUnclaimableProcessingDeliveryRemainsRetryable(t *testing.T) {
	t.Parallel()
	repository := processingRepository(domain.Processing)
	repository.claim = func(context.Context, string, domain.ClaimToken, time.Time) (domain.VideoAsset, bool, error) {
		return repository.asset, false, nil
	}
	transcoder := defaultTranscoder()
	_, err := newTranscode(repository, defaultStorage(), transcoder, nil).Process(context.Background(), ports.Job{AssetID: mediaAssetID, MaxRetry: 3})
	if err == nil || errors.Is(err, domain.ErrDependencyUnavailable) {
		t.Fatalf("got %v", err)
	}
	if len(transcoder.calls) != 0 || countCall(repository.calls, "publish") != 0 {
		t.Fatalf("repository=%v transcoder=%v", repository.calls, transcoder.calls)
	}
}

func TestCancellationCausesAreHandledSeparately(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name       string
		ctx        func() (context.Context, context.CancelFunc)
		retryCount int
		wantState  domain.State
		wantErr    error
		wantReason domain.FailureReason
	}{
		{name: "plain cancellation requeues", ctx: func() (context.Context, context.CancelFunc) {
			return context.WithCancel(context.Background())
		}, retryCount: 3, wantState: domain.Queued, wantErr: context.Canceled},
		{name: "final deadline fails", ctx: func() (context.Context, context.CancelFunc) {
			ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
			return ctx, cancel
		}, retryCount: 3, wantState: domain.Failed, wantReason: domain.RenderFailed},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			repository := processingRepository(domain.Queued)
			started := make(chan struct{})
			transcoder := defaultTranscoder()
			transcoder.render = func(ctx context.Context, _ domain.StorageKey, _ domain.StorageKey, _ []domain.Rendition, _ bool) error {
				close(started)
				<-ctx.Done()
				return ctx.Err()
			}
			ctx, cancel := test.ctx()
			result := make(chan struct {
				outcome application.TranscodeOutcome
				err     error
			}, 1)
			go func() {
				outcome, err := newTranscode(repository, defaultStorage(), transcoder, nil).Process(ctx, ports.Job{AssetID: mediaAssetID, RetryCount: test.retryCount, MaxRetry: 3})
				result <- struct {
					outcome application.TranscodeOutcome
					err     error
				}{outcome: outcome, err: err}
			}()
			<-started
			cancel()
			got := <-result
			if repository.asset.State != test.wantState || repository.asset.ClaimToken != "" || repository.asset.FailureReason != test.wantReason {
				t.Fatalf("asset=%+v", repository.asset)
			}
			if test.wantErr != nil {
				if !errors.Is(got.err, test.wantErr) || got.outcome != (application.TranscodeOutcome{}) {
					t.Fatalf("outcome=%+v err=%v", got.outcome, got.err)
				}
			} else if got.err != nil || !got.outcome.Failed || got.outcome.Permanent || got.outcome.Reason != test.wantReason {
				t.Fatalf("outcome=%+v err=%v", got.outcome, got.err)
			}
		})
	}
}

func TestDiskExhaustionIsRetryableAndCleansWorkingKeys(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name      string
		configure func(*fakeStorage, *fakeTranscoder)
	}{
		{name: "gate", configure: func(storage *fakeStorage, _ *fakeTranscoder) {
			storage.free = func(context.Context) (int64, error) { return 1, nil }
		}},
		{name: "render", configure: func(_ *fakeStorage, transcoder *fakeTranscoder) {
			transcoder.render = func(context.Context, domain.StorageKey, domain.StorageKey, []domain.Rendition, bool) error {
				return syscall.ENOSPC
			}
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			repository := processingRepository(domain.Queued)
			storage := defaultStorage()
			transcoder := defaultTranscoder()
			test.configure(storage, transcoder)
			outcome, err := newTranscode(repository, storage, transcoder, nil).Process(context.Background(), ports.Job{AssetID: mediaAssetID, RetryCount: 1, MaxRetry: 3})
			if err != nil || !outcome.Failed || outcome.Permanent || outcome.Reason != domain.StorageUnavailable {
				t.Fatalf("outcome=%+v err=%v", outcome, err)
			}
			if repository.asset.State != domain.Queued || repository.asset.ClaimToken != "" || countCall(repository.calls, "requeue") != 1 || countString(storage.calls, "purge") < 2 {
				t.Fatalf("asset=%+v repository=%v storage=%v", repository.asset, repository.calls, storage.calls)
			}
			if _, claimed, claimErr := repository.Claim(context.Background(), mediaAssetID, "later", time.Now()); claimErr != nil || !claimed {
				t.Fatalf("later claim=%t err=%v", claimed, claimErr)
			}
		})
	}
}

func TestPublishFailureIsRequeuedWithClosedReason(t *testing.T) {
	t.Parallel()
	repository := processingRepository(domain.Queued)
	storage := defaultStorage()
	storage.publish = func(context.Context, domain.StorageKey, domain.StorageKey) error {
		return errors.New("rename failed at /media")
	}
	outcome, err := newTranscode(repository, storage, defaultTranscoder(), nil).Process(context.Background(), ports.Job{AssetID: mediaAssetID, RetryCount: 1, MaxRetry: 3})
	if err != nil || !outcome.Failed || outcome.Reason != domain.PublishFailed || outcome.Permanent {
		t.Fatalf("outcome=%+v err=%v", outcome, err)
	}
	if repository.asset.State != domain.Queued || countCall(repository.calls, "requeue") != 1 {
		t.Fatalf("asset=%+v calls=%v", repository.asset, repository.calls)
	}
}

func TestDependencyFailurePassesThroughAndDoesNotBecomeBusinessFailure(t *testing.T) {
	t.Parallel()
	want := domain.ErrDependencyUnavailable
	repository := processingRepository(domain.Queued)
	repository.load = func(context.Context, string) (domain.VideoAsset, error) { return domain.VideoAsset{}, want }
	transcoder := defaultTranscoder()
	outcome, err := newTranscode(repository, defaultStorage(), transcoder, nil).Process(context.Background(), ports.Job{AssetID: mediaAssetID, MaxRetry: 3})
	if err != want || outcome != (application.TranscodeOutcome{}) || err.Error() != "dependency_unavailable" {
		t.Fatalf("outcome=%+v err=%v", outcome, err)
	}
	if countCall(repository.calls, "claim") != 0 || len(transcoder.calls) != 0 {
		t.Fatalf("repository=%v transcoder=%v", repository.calls, transcoder.calls)
	}
}

func TestLostPublishClaimTouchesNoPublishedBytes(t *testing.T) {
	t.Parallel()
	repository := processingRepository(domain.Queued)
	repository.publish = func(context.Context, string, domain.ClaimToken, []domain.Rendition, *float64, time.Time, *domain.AudioTrack, ports.PublishPolicy) (domain.VideoAsset, *domain.VideoAsset, error) {
		return domain.VideoAsset{}, nil, nil
	}
	storage := defaultStorage()
	outcome, err := newTranscode(repository, storage, defaultTranscoder(), nil).Process(context.Background(), ports.Job{AssetID: mediaAssetID, MaxRetry: 3})
	if err != nil || outcome != (application.TranscodeOutcome{}) {
		t.Fatalf("outcome=%+v err=%v", outcome, err)
	}
	if countString(storage.calls, "publish") != 0 || countString(storage.calls, "move") != 0 {
		t.Fatalf("storage calls=%v", storage.calls)
	}
}

func TestSweepPurgesEveryKeyBeforeDeletingTheRow(t *testing.T) {
	t.Parallel()
	var purgeCalls []string
	repository := processingRepository(domain.Queued)
	repository.sweepCandidates = func(ctx context.Context, _ time.Time, limit int) ([]domain.VideoAsset, error) {
		assertMediaDeadline(t, ctx, 3*time.Second)
		if limit != 50 {
			t.Fatalf("limit=%d", limit)
		}
		return []domain.VideoAsset{mediaAsset(domain.Failed)}, nil
	}
	repository.deleteSwept = func(context.Context, string) error {
		if len(purgeCalls) != 4 {
			t.Fatalf("delete before purge: %v", purgeCalls)
		}
		return nil
	}
	storage := defaultStorage()
	storage.purge = func(_ context.Context, key domain.StorageKey) error {
		purgeCalls = append(purgeCalls, string(key.Kind))
		return nil
	}
	if err := newTranscode(repository, storage, defaultTranscoder(), nil).Sweep(context.Background()); err != nil {
		t.Fatal(err)
	}
	assertStrings(t, purgeCalls, []string{"source", "hls", "thumbnails", "temporary"})
	if countCall(repository.calls, "delete_swept") != 1 {
		t.Fatalf("calls=%v", repository.calls)
	}
}

func TestSweepLeavesTheRowWhenPurgeFails(t *testing.T) {
	t.Parallel()
	repository := processingRepository(domain.Queued)
	repository.sweepCandidates = func(context.Context, time.Time, int) ([]domain.VideoAsset, error) {
		return []domain.VideoAsset{mediaAsset(domain.Failed)}, nil
	}
	storage := defaultStorage()
	storage.purge = func(_ context.Context, key domain.StorageKey) error {
		if key.Kind == domain.StorageHLS {
			return domain.ErrDependencyUnavailable
		}
		return nil
	}
	if err := newTranscode(repository, storage, defaultTranscoder(), nil).Sweep(context.Background()); !errors.Is(err, domain.ErrDependencyUnavailable) {
		t.Fatalf("got %v", err)
	}
	if countCall(repository.calls, "delete_swept") != 0 {
		t.Fatalf("row deleted: %v", repository.calls)
	}
}

func TestRecoverEnqueuesIncrementedAttemptsAndDuplicateDeliveryIsInert(t *testing.T) {
	t.Parallel()
	for _, scenario := range []string{"archived old task", "pending old task"} {
		t.Run(scenario, func(t *testing.T) {
			t.Parallel()
			repository := processingRepository(domain.Queued)
			repository.asset.Attempt = 0
			repository.recover = func(ctx context.Context, _ time.Time, limit int) ([]domain.VideoAsset, error) {
				assertMediaDeadline(t, ctx, 3*time.Second)
				if limit != 50 {
					t.Fatalf("limit=%d", limit)
				}
				repository.asset.Attempt++
				return []domain.VideoAsset{repository.asset}, nil
			}
			queue := &fakeQueue{enqueue: func(_ context.Context, id string, attempt int) error {
				if id != mediaAssetID || attempt != 1 {
					t.Fatalf("id=%q attempt=%d", id, attempt)
				}
				return nil
			}}
			transcoder := defaultTranscoder()
			service := newTranscode(repository, defaultStorage(), transcoder, queue)
			if err := service.Recover(context.Background()); err != nil {
				t.Fatal(err)
			}
			if len(queue.calls) != 1 || repository.asset.Attempt != 1 {
				t.Fatalf("queue=%v asset=%+v", queue.calls, repository.asset)
			}

			if scenario == "pending old task" {
				// The older pending task finishes first; the recovery delivery then sees ready.
				repository.asset.State = domain.Ready
				if outcome, err := service.Process(context.Background(), ports.Job{AssetID: mediaAssetID, Attempt: 1, MaxRetry: 3}); err != nil || outcome != (application.TranscodeOutcome{}) {
					t.Fatalf("outcome=%+v err=%v", outcome, err)
				}
				if len(transcoder.calls) != 0 || countCall(repository.calls, "claim") != 0 {
					t.Fatalf("repository=%v transcoder=%v", repository.calls, transcoder.calls)
				}
			}
		})
	}
}

func TestRecoverFreesProcessingRowsWithoutAddingASecondTask(t *testing.T) {
	t.Parallel()
	repository := processingRepository(domain.Processing)
	repository.recover = func(context.Context, time.Time, int) ([]domain.VideoAsset, error) {
		repository.asset.State = domain.Queued
		repository.asset.ClaimToken = ""
		return nil, nil
	}
	queue := &fakeQueue{enqueue: func(context.Context, string, int) error { return nil }}
	if err := newTranscode(repository, defaultStorage(), defaultTranscoder(), queue).Recover(context.Background()); err != nil {
		t.Fatal(err)
	}
	if repository.asset.State != domain.Queued || len(queue.calls) != 0 {
		t.Fatalf("asset=%+v queue=%v", repository.asset, queue.calls)
	}
}

func processingRepository(state domain.State) *fakeRepository {
	repository := &fakeRepository{asset: mediaAsset(state)}
	repository.sweepCandidates = func(context.Context, time.Time, int) ([]domain.VideoAsset, error) { return nil, nil }
	repository.load = func(context.Context, string) (domain.VideoAsset, error) { return repository.asset, nil }
	repository.claim = func(_ context.Context, _ string, token domain.ClaimToken, _ time.Time) (domain.VideoAsset, bool, error) {
		if repository.asset.State != domain.Queued {
			return repository.asset, false, nil
		}
		repository.asset.State = domain.Processing
		repository.asset.ClaimToken = token
		return repository.asset, true, nil
	}
	repository.publish = func(_ context.Context, _ string, token domain.ClaimToken, renditions []domain.Rendition, durationSeconds *float64, _ time.Time, embedded *domain.AudioTrack, policy ports.PublishPolicy) (domain.VideoAsset, *domain.VideoAsset, error) {
		if repository.asset.State != domain.Processing || repository.asset.ClaimToken != token {
			return domain.VideoAsset{}, nil, nil
		}
		if err := policy(repository.asset); err != nil {
			return domain.VideoAsset{}, nil, err
		}
		if embedded != nil {
			repository.asset.State = domain.Ready
		} else {
			repository.asset.State = domain.AwaitingAudio
		}
		repository.asset.ClaimToken = ""
		repository.asset.Renditions = renditions
		repository.publishedDuration = durationSeconds
		return repository.asset, nil, nil
	}
	repository.requeue = func(_ context.Context, _ string, token domain.ClaimToken, _ time.Time) error {
		if repository.asset.State == domain.Processing && repository.asset.ClaimToken == token {
			repository.asset.State = domain.Queued
			repository.asset.ClaimToken = ""
		}
		return nil
	}
	repository.fail = func(_ context.Context, _ string, from domain.State, token domain.ClaimToken, reason domain.FailureReason, _ time.Time) error {
		if repository.asset.State == from && repository.asset.ClaimToken == token {
			repository.asset.State = domain.Failed
			repository.asset.ClaimToken = ""
			repository.asset.FailureReason = reason
		}
		return nil
	}
	repository.deleteSwept = func(context.Context, string) error { return nil }
	repository.recover = func(context.Context, time.Time, int) ([]domain.VideoAsset, error) { return nil, nil }
	return repository
}

func defaultTranscoder() *fakeTranscoder {
	return &fakeTranscoder{
		probe: func(context.Context, domain.StorageKey) (domain.SourceProbe, error) {
			return domain.SourceProbe{Width: 1280, Height: 720, VideoCodec: domain.CodecH264, AudioCodec: domain.CodecAAC, DurationSeconds: 125.5}, nil
		},
		render: func(context.Context, domain.StorageKey, domain.StorageKey, []domain.Rendition, bool) error {
			return nil
		},
		thumbnail: func(context.Context, domain.StorageKey, domain.StorageKey) error { return nil },
	}
}

func newTranscode(repository *fakeRepository, storage *fakeStorage, transcoder *fakeTranscoder, queue *fakeQueue) *application.Transcode {
	return application.NewTranscode(repository, storage, transcoder, queue, func() time.Time {
		return time.Date(2026, 9, 6, 12, 0, 0, 123456789, time.UTC)
	})
}

func countCall(calls []string, want string) int { return countString(calls, want) }

func countString(values []string, want string) int {
	count := 0
	for _, value := range values {
		if value == want {
			count++
		}
	}
	return count
}
