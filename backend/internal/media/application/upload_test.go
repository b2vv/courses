package application_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"syscall"
	"testing"
	"time"

	"courses/internal/media/application"
	"courses/internal/media/domain"
	"courses/internal/media/ports"
)

const (
	mediaAssetID  = "50000000-0000-4000-8000-000000000001"
	mediaLessonID = "50000000-0000-4000-8000-000000000002"
	mediaCourseID = "50000000-0000-4000-8000-000000000003"
)

type fakeRepository struct {
	calls             []string
	asset             domain.VideoAsset
	mint              func(context.Context, domain.Actor, string, int64, string, time.Time, time.Time) (domain.VideoAsset, error)
	appendChunk       func(context.Context, string, int64, time.Time, ports.ChunkPolicy, ports.CompletePolicy) (domain.VideoAsset, error)
	cancel            func(context.Context, string, time.Time) error
	list              func(context.Context, domain.Actor, string) ([]domain.VideoAsset, error)
	load              func(context.Context, string) (domain.VideoAsset, error)
	claim             func(context.Context, string, domain.ClaimToken, time.Time) (domain.VideoAsset, bool, error)
	publish           func(context.Context, string, domain.ClaimToken, []domain.Rendition, *float64, time.Time, *domain.AudioTrack, ports.PublishPolicy) (domain.VideoAsset, *domain.VideoAsset, error)
	publishedEmbedded *domain.AudioTrack
	regenerateMaster  func(context.Context, string, func([]domain.AudioTrack) error) error
	loadAudioTrack    func(context.Context, string, string) (domain.AudioTrack, error)
	promoteAudioTrack func(context.Context, string, string, time.Time, func([]domain.AudioTrack) error) error
	failAudioTrack    func(context.Context, string, string, time.Time) error
	publishedDuration *float64
	requeue           func(context.Context, string, domain.ClaimToken, time.Time) error
	fail              func(context.Context, string, domain.State, domain.ClaimToken, domain.FailureReason, time.Time) error
	retry             func(context.Context, domain.Actor, string, time.Duration, time.Time) (domain.VideoAsset, error)
	sweepCandidates   func(context.Context, time.Time, int) ([]domain.VideoAsset, error)
	deleteSwept       func(context.Context, string) error
	recover           func(context.Context, time.Time, int) ([]domain.VideoAsset, error)
}

func (repository *fakeRepository) Mint(ctx context.Context, actor domain.Actor, lessonID string, declaredBytes int64, originalName string, expiresAt, now time.Time) (domain.VideoAsset, error) {
	repository.calls = append(repository.calls, "mint")
	return repository.mint(ctx, actor, lessonID, declaredBytes, originalName, expiresAt, now)
}

func (repository *fakeRepository) AppendChunk(ctx context.Context, id string, offset int64, now time.Time, chunkPolicy ports.ChunkPolicy, completePolicy ports.CompletePolicy) (domain.VideoAsset, error) {
	repository.calls = append(repository.calls, "append")
	return repository.appendChunk(ctx, id, offset, now, chunkPolicy, completePolicy)
}

func (repository *fakeRepository) Cancel(ctx context.Context, id string, now time.Time) error {
	repository.calls = append(repository.calls, "cancel")
	return repository.cancel(ctx, id, now)
}

func (repository *fakeRepository) ListCourseVideos(ctx context.Context, actor domain.Actor, courseID string) ([]domain.VideoAsset, error) {
	repository.calls = append(repository.calls, "list")
	return repository.list(ctx, actor, courseID)
}

func (repository *fakeRepository) Load(ctx context.Context, id string) (domain.VideoAsset, error) {
	repository.calls = append(repository.calls, "load")
	return repository.load(ctx, id)
}

func (repository *fakeRepository) Claim(ctx context.Context, id string, token domain.ClaimToken, now time.Time) (domain.VideoAsset, bool, error) {
	repository.calls = append(repository.calls, "claim")
	return repository.claim(ctx, id, token, now)
}

func (repository *fakeRepository) Publish(ctx context.Context, id string, token domain.ClaimToken, renditions []domain.Rendition, durationSeconds *float64, now time.Time, embedded *domain.AudioTrack, policy ports.PublishPolicy) (domain.VideoAsset, *domain.VideoAsset, error) {
	repository.calls = append(repository.calls, "publish")
	repository.publishedEmbedded = embedded
	return repository.publish(ctx, id, token, renditions, durationSeconds, now, embedded, policy)
}

func (repository *fakeRepository) RegenerateMaster(ctx context.Context, assetID string, rewrite func([]domain.AudioTrack) error) error {
	repository.calls = append(repository.calls, "regenerate_master")
	if repository.regenerateMaster != nil {
		return repository.regenerateMaster(ctx, assetID, rewrite)
	}
	return nil
}

func (repository *fakeRepository) LoadAudioTrack(ctx context.Context, assetID, trackID string) (domain.AudioTrack, error) {
	repository.calls = append(repository.calls, "load_audio_track")
	return repository.loadAudioTrack(ctx, assetID, trackID)
}

func (repository *fakeRepository) PromoteAudioTrack(ctx context.Context, assetID, trackID string, now time.Time, rewrite func([]domain.AudioTrack) error) error {
	repository.calls = append(repository.calls, "promote_audio_track")
	return repository.promoteAudioTrack(ctx, assetID, trackID, now, rewrite)
}

func (repository *fakeRepository) FailAudioTrack(ctx context.Context, assetID, trackID string, now time.Time) error {
	repository.calls = append(repository.calls, "fail_audio_track")
	return repository.failAudioTrack(ctx, assetID, trackID, now)
}

func (repository *fakeRepository) Requeue(ctx context.Context, id string, token domain.ClaimToken, now time.Time) error {
	repository.calls = append(repository.calls, "requeue")
	return repository.requeue(ctx, id, token, now)
}

func (repository *fakeRepository) Fail(ctx context.Context, id string, from domain.State, token domain.ClaimToken, reason domain.FailureReason, now time.Time) error {
	repository.calls = append(repository.calls, "fail")
	return repository.fail(ctx, id, from, token, reason, now)
}

func (repository *fakeRepository) Retry(ctx context.Context, actor domain.Actor, id string, claimAge time.Duration, now time.Time) (domain.VideoAsset, error) {
	repository.calls = append(repository.calls, "retry")
	return repository.retry(ctx, actor, id, claimAge, now)
}

func (repository *fakeRepository) SweepCandidates(ctx context.Context, now time.Time, limit int) ([]domain.VideoAsset, error) {
	repository.calls = append(repository.calls, "sweep")
	return repository.sweepCandidates(ctx, now, limit)
}

func (repository *fakeRepository) DeleteSwept(ctx context.Context, id string) error {
	repository.calls = append(repository.calls, "delete_swept")
	return repository.deleteSwept(ctx, id)
}

func (repository *fakeRepository) Recover(ctx context.Context, now time.Time, limit int) ([]domain.VideoAsset, error) {
	repository.calls = append(repository.calls, "recover")
	return repository.recover(ctx, now, limit)
}

type fakeStorage struct {
	calls    []string
	free     func(context.Context) (int64, error)
	size     func(context.Context, domain.StorageKey) (int64, error)
	head     func(context.Context, domain.StorageKey, int) ([]byte, error)
	append   func(context.Context, domain.StorageKey, int64, io.Reader) (int64, error)
	truncate func(context.Context, domain.StorageKey, int64) error
	move     func(context.Context, domain.StorageKey, domain.StorageKey) error
	publish  func(context.Context, domain.StorageKey, domain.StorageKey) error
	purge    func(context.Context, domain.StorageKey) error
}

func (storage *fakeStorage) Free(ctx context.Context) (int64, error) {
	storage.calls = append(storage.calls, "free")
	return storage.free(ctx)
}

func (storage *fakeStorage) Size(ctx context.Context, key domain.StorageKey) (int64, error) {
	storage.calls = append(storage.calls, "size")
	return storage.size(ctx, key)
}

func (storage *fakeStorage) Head(ctx context.Context, key domain.StorageKey, n int) ([]byte, error) {
	storage.calls = append(storage.calls, "head")
	return storage.head(ctx, key, n)
}

func (storage *fakeStorage) Append(ctx context.Context, key domain.StorageKey, offset int64, body io.Reader) (int64, error) {
	storage.calls = append(storage.calls, "append")
	return storage.append(ctx, key, offset, body)
}

func (storage *fakeStorage) Truncate(ctx context.Context, key domain.StorageKey, size int64) error {
	storage.calls = append(storage.calls, "truncate")
	return storage.truncate(ctx, key, size)
}

func (storage *fakeStorage) Move(ctx context.Context, from, to domain.StorageKey) error {
	storage.calls = append(storage.calls, "move")
	return storage.move(ctx, from, to)
}

func (storage *fakeStorage) Publish(ctx context.Context, work, destination domain.StorageKey) error {
	storage.calls = append(storage.calls, "publish")
	return storage.publish(ctx, work, destination)
}

func (storage *fakeStorage) Purge(ctx context.Context, key domain.StorageKey) error {
	storage.calls = append(storage.calls, "purge")
	return storage.purge(ctx, key)
}

type fakeQueue struct {
	calls   []string
	enqueue func(context.Context, string, int) error
}

func (queue *fakeQueue) Enqueue(ctx context.Context, id string, attempt int) error {
	queue.calls = append(queue.calls, id)
	return queue.enqueue(ctx, id, attempt)
}

func (queue *fakeQueue) EnqueueAudioTrack(ctx context.Context, assetID, trackID string, attempt int) error {
	queue.calls = append(queue.calls, assetID+":"+trackID)
	return nil
}

func TestMintTicketValidatesBeforeRepositoryAndBoundsTheCall(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 9, 6, 12, 0, 0, 123456789, time.FixedZone("test", 2*60*60))
	wantNow := now.UTC().Truncate(time.Microsecond)
	actor := domain.Actor{UserID: "user", Administrator: true}
	repository := &fakeRepository{mint: func(ctx context.Context, gotActor domain.Actor, lessonID string, declared int64, name string, expiresAt, at time.Time) (domain.VideoAsset, error) {
		assertMediaDeadline(t, ctx, 3*time.Second)
		if gotActor != actor || lessonID != mediaLessonID || declared != 12 || name != "lesson.mp4" || !at.Equal(wantNow) || !expiresAt.Equal(wantNow.Add(24*time.Hour)) {
			t.Fatalf("mint input actor=%+v lesson=%q bytes=%d name=%q expires=%v now=%v", gotActor, lessonID, declared, name, expiresAt, at)
		}
		return mediaAsset(domain.Created), nil
	}}
	tickets := application.NewTickets(repository, nil, func() time.Time { return now })
	if _, err := tickets.MintTicket(context.Background(), actor, mediaLessonID, 12, "lesson.mp4"); err != nil {
		t.Fatal(err)
	}
	if len(repository.calls) != 1 || repository.calls[0] != "mint" {
		t.Fatalf("calls=%v", repository.calls)
	}

	repository.calls = nil
	if _, err := tickets.MintTicket(context.Background(), actor, mediaLessonID, 0, "lesson.mp4"); !errors.Is(err, domain.ErrInvalidRequest) {
		t.Fatalf("zero bytes: %v", err)
	}
	if _, err := tickets.MintTicket(context.Background(), actor, mediaLessonID, domain.MaximumDeclaredBytes+1, "lesson.mp4"); !errors.Is(err, domain.ErrUploadTooLarge) {
		t.Fatalf("oversized: %v", err)
	}
	if len(repository.calls) != 0 {
		t.Fatalf("repository called for invalid ticket: %v", repository.calls)
	}
}

func TestAppendChunkPastDeclaredLengthRollsStorageBack(t *testing.T) {
	t.Parallel()
	asset := mediaAsset(domain.Uploading)
	asset.ReceivedBytes = asset.DeclaredBytes
	storage := defaultStorage()
	repository := &fakeRepository{appendChunk: func(_ context.Context, _ string, _ int64, _ time.Time, chunkPolicy ports.ChunkPolicy, _ ports.CompletePolicy) (domain.VideoAsset, error) {
		_, err := chunkPolicy(asset)
		return domain.VideoAsset{}, err
	}}

	_, err := application.NewUpload(repository, storage, nil, time.Now).AppendChunk(context.Background(), asset.ID, asset.ReceivedBytes, bytes.NewBufferString("x"))
	if !errors.Is(err, domain.ErrInvalidRequest) {
		t.Fatalf("got %v, want invalid request", err)
	}
	assertStrings(t, storage.calls, []string{"size", "append", "truncate"})
}

func TestAppendChunkRejectsAScratchFileShorterThanTheColumn(t *testing.T) {
	t.Parallel()
	asset := mediaAsset(domain.Uploading)
	asset.ReceivedBytes = 12
	asset.DeclaredBytes = 24
	storage := defaultStorage()
	storage.size = func(context.Context, domain.StorageKey) (int64, error) { return 11, nil }
	repository := &fakeRepository{appendChunk: func(_ context.Context, _ string, _ int64, _ time.Time, chunkPolicy ports.ChunkPolicy, _ ports.CompletePolicy) (domain.VideoAsset, error) {
		_, err := chunkPolicy(asset)
		return domain.VideoAsset{}, err
	}}

	_, err := application.NewUpload(repository, storage, nil, time.Now).AppendChunk(context.Background(), asset.ID, asset.ReceivedBytes, bytes.NewBufferString("x"))
	if !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("got %v, want unexpected EOF", err)
	}
	assertStrings(t, storage.calls, []string{"size"})
}

func TestAppendChunkSurfacesRollbackFailure(t *testing.T) {
	t.Parallel()
	asset := mediaAsset(domain.Uploading)
	asset.ReceivedBytes = 12
	asset.DeclaredBytes = 24
	appendFailure := errors.New("append failed")
	rollbackFailure := errors.New("rollback failed")
	storage := defaultStorage()
	storage.append = func(context.Context, domain.StorageKey, int64, io.Reader) (int64, error) { return 1, appendFailure }
	storage.truncate = func(context.Context, domain.StorageKey, int64) error { return rollbackFailure }
	repository := &fakeRepository{appendChunk: func(_ context.Context, _ string, _ int64, _ time.Time, chunkPolicy ports.ChunkPolicy, _ ports.CompletePolicy) (domain.VideoAsset, error) {
		_, err := chunkPolicy(asset)
		return domain.VideoAsset{}, err
	}}

	_, err := application.NewUpload(repository, storage, nil, time.Now).AppendChunk(context.Background(), asset.ID, asset.ReceivedBytes, bytes.NewBufferString("x"))
	if !errors.Is(err, rollbackFailure) || errors.Is(err, appendFailure) {
		t.Fatalf("got %v, want rollback failure only", err)
	}
	assertStrings(t, storage.calls, []string{"size", "append", "truncate"})
}

func TestHeadUploadExposesOnlyActiveTickets(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 9, 6, 12, 0, 0, 0, time.UTC)
	for _, test := range []struct {
		name  string
		asset domain.VideoAsset
		want  error
	}{
		{name: "created", asset: mediaAsset(domain.Created)},
		{name: "uploading", asset: mediaAsset(domain.Uploading)},
		{name: "expired", asset: func() domain.VideoAsset {
			asset := mediaAsset(domain.Uploading)
			asset.ExpiresAt = now
			return asset
		}(), want: domain.ErrVideoNotFound},
		{name: "completed", asset: mediaAsset(domain.Queued), want: domain.ErrVideoNotFound},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			repository := &fakeRepository{load: func(ctx context.Context, id string) (domain.VideoAsset, error) {
				assertMediaDeadline(t, ctx, 3*time.Second)
				if id != mediaAssetID {
					t.Fatalf("id=%q", id)
				}
				return test.asset, nil
			}}
			got, err := application.NewUpload(repository, nil, nil, func() time.Time { return now }).HeadUpload(context.Background(), mediaAssetID)
			if !errors.Is(err, test.want) || test.want == nil && got.ID != test.asset.ID {
				t.Fatalf("asset=%+v err=%v", got, err)
			}
		})
	}
}

func TestAppendChunkRunsRepairWriteAndFirstChunkPolicies(t *testing.T) {
	t.Parallel()
	asset := mediaAsset(domain.Created)
	asset.DeclaredBytes = 24
	storage := defaultStorage()
	storage.append = func(_ context.Context, key domain.StorageKey, offset int64, body io.Reader) (int64, error) {
		if key.Kind != domain.StorageTemporary || key.ID != asset.ID || key.Leaf != "upload" || offset != 0 {
			t.Fatalf("append key=%+v offset=%d", key, offset)
		}
		data, err := io.ReadAll(body)
		return int64(len(data)), err
	}
	repository := &fakeRepository{appendChunk: func(ctx context.Context, _ string, offset int64, _ time.Time, chunkPolicy ports.ChunkPolicy, _ ports.CompletePolicy) (domain.VideoAsset, error) {
		assertMediaDeadlineBetween(t, ctx, 119*time.Second, 120*time.Second)
		newOffset, err := chunkPolicy(asset)
		if err != nil {
			return domain.VideoAsset{}, err
		}
		asset.ReceivedBytes = newOffset
		asset.State = domain.Uploading
		return asset, nil
	}}

	got, err := application.NewUpload(repository, storage, nil, time.Now).AppendChunk(context.Background(), asset.ID, 0, bytes.NewBufferString("0123456789ab"))
	if err != nil {
		t.Fatal(err)
	}
	if got.State != domain.Uploading || got.ReceivedBytes != 12 {
		t.Fatalf("asset=%+v", got)
	}
	assertStrings(t, storage.calls, []string{"size", "free", "truncate", "append", "head"})
}

func TestCompletedSingleChunkSourceRepairsBeforeTheDiskGate(t *testing.T) {
	t.Parallel()
	asset := mediaAsset(domain.Created)
	storage := defaultStorage()
	storage.size = func(_ context.Context, key domain.StorageKey) (int64, error) {
		if key.Kind != domain.StorageSource {
			t.Fatalf("Size key=%+v", key)
		}
		return asset.DeclaredBytes, nil
	}
	storage.free = func(context.Context) (int64, error) {
		t.Fatal("Free called for an already-complete source")
		return 0, nil
	}
	repository := &fakeRepository{appendChunk: func(_ context.Context, _ string, _ int64, _ time.Time, chunkPolicy ports.ChunkPolicy, completePolicy ports.CompletePolicy) (domain.VideoAsset, error) {
		newOffset, err := chunkPolicy(asset)
		if err != nil {
			return domain.VideoAsset{}, err
		}
		asset.ReceivedBytes = newOffset
		asset.State = domain.Uploading
		if err := completePolicy(asset); err != nil {
			return domain.VideoAsset{}, err
		}
		asset.State = domain.Queued
		return asset, nil
	}}
	queue := &fakeQueue{enqueue: func(context.Context, string, int) error { return nil }}

	got, err := application.NewUpload(repository, storage, queue, time.Now).AppendChunk(context.Background(), asset.ID, 0, bytes.NewBufferString("not-read"))
	if err != nil || got.State != domain.Queued || got.ReceivedBytes != got.DeclaredBytes {
		t.Fatalf("asset=%+v err=%v", got, err)
	}
	assertStrings(t, storage.calls, []string{"size", "size"})
}

func TestLaterChunkSkipsTheFirstChunkGateAndSniff(t *testing.T) {
	t.Parallel()
	asset := mediaAsset(domain.Uploading)
	asset.DeclaredBytes = 36
	asset.ReceivedBytes = 12
	storage := defaultStorage()
	repository := &fakeRepository{appendChunk: func(_ context.Context, _ string, _ int64, _ time.Time, chunkPolicy ports.ChunkPolicy, _ ports.CompletePolicy) (domain.VideoAsset, error) {
		newOffset, err := chunkPolicy(asset)
		if err != nil {
			return domain.VideoAsset{}, err
		}
		asset.ReceivedBytes = newOffset
		return asset, nil
	}}
	got, err := application.NewUpload(repository, storage, nil, time.Now).AppendChunk(context.Background(), asset.ID, 12, bytes.NewBufferString("0123456789ab"))
	if err != nil {
		t.Fatal(err)
	}
	if got.ReceivedBytes != 24 || got.State != domain.Uploading {
		t.Fatalf("asset=%+v", got)
	}
	assertStrings(t, storage.calls, []string{"size", "append"})
}

func TestAppendChunkDiskAndContainerRefusalsDoNotCommitUploading(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name       string
		configure  func(*fakeStorage)
		want       error
		wantState  domain.State
		wantReason domain.FailureReason
	}{
		{name: "disk full", configure: func(storage *fakeStorage) {
			storage.free = func(context.Context) (int64, error) { return 2, nil }
		}, want: syscall.ENOSPC, wantState: domain.Created},
		{name: "unsupported container", configure: func(storage *fakeStorage) {
			storage.head = func(context.Context, domain.StorageKey, int) ([]byte, error) { return []byte("plain text!!"), nil }
		}, want: domain.ErrUnsupportedMedia, wantState: domain.Failed, wantReason: domain.UnsupportedSource},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			asset := mediaAsset(domain.Created)
			storage := defaultStorage()
			test.configure(storage)
			repository := &fakeRepository{}
			repository.appendChunk = func(_ context.Context, _ string, _ int64, _ time.Time, chunkPolicy ports.ChunkPolicy, _ ports.CompletePolicy) (domain.VideoAsset, error) {
				newOffset, err := chunkPolicy(asset)
				if err != nil {
					if errors.Is(err, domain.ErrUnsupportedSource) {
						asset.State = domain.Failed
						asset.FailureReason = domain.UnsupportedSource
					}
					return domain.VideoAsset{}, err
				}
				asset.State = domain.Uploading
				asset.ReceivedBytes = newOffset
				return asset, nil
			}
			repository.fail = func(context.Context, string, domain.State, domain.ClaimToken, domain.FailureReason, time.Time) error {
				t.Fatal("unsupported first chunk must be failed inside AppendChunk")
				return nil
			}

			_, err := application.NewUpload(repository, storage, nil, time.Now).AppendChunk(context.Background(), asset.ID, 0, bytes.NewBufferString("0123456789ab"))
			if !errors.Is(err, test.want) {
				t.Fatalf("got %v, want %v", err, test.want)
			}
			if asset.State != test.wantState || asset.FailureReason != test.wantReason || asset.ReceivedBytes != 0 {
				t.Fatalf("asset=%+v", asset)
			}
		})
	}
}

func TestCompleteCommitsBeforeEnqueueAndRetryRepairsAnEnqueueFailure(t *testing.T) {
	t.Parallel()
	asset := mediaAsset(domain.Created)
	storage := defaultStorage()
	repository := &fakeRepository{}
	repository.appendChunk = func(_ context.Context, _ string, _ int64, _ time.Time, chunkPolicy ports.ChunkPolicy, completePolicy ports.CompletePolicy) (domain.VideoAsset, error) {
		newOffset, err := chunkPolicy(asset)
		if err != nil {
			return domain.VideoAsset{}, err
		}
		asset.ReceivedBytes = newOffset
		asset.State = domain.Uploading
		if err := completePolicy(asset); err != nil {
			return domain.VideoAsset{}, err
		}
		asset.State = domain.Queued
		return asset, nil
	}
	repository.retry = func(_ context.Context, _ domain.Actor, _ string, claimAge time.Duration, _ time.Time) (domain.VideoAsset, error) {
		if asset.State != domain.Queued || claimAge != 40*time.Minute {
			t.Fatalf("retry observed state=%s claimAge=%v", asset.State, claimAge)
		}
		asset.Attempt++
		return asset, nil
	}
	enqueueFailure := domain.ErrDependencyUnavailable
	queue := &fakeQueue{enqueue: func(_ context.Context, _ string, attempt int) error {
		if asset.State != domain.Queued {
			t.Fatalf("enqueue ran before commit: %s", asset.State)
		}
		if attempt == 0 {
			return enqueueFailure
		}
		return nil
	}}
	upload := application.NewUpload(repository, storage, queue, time.Now)
	tickets := application.NewTickets(repository, queue, time.Now)

	if _, err := upload.AppendChunk(context.Background(), asset.ID, 0, bytes.NewBufferString("\x00\x00\x00\x18ftypisom")); !errors.Is(err, enqueueFailure) {
		t.Fatalf("complete: %v", err)
	}
	if asset.State != domain.Queued {
		t.Fatalf("state=%s", asset.State)
	}
	if got, err := tickets.Retry(context.Background(), domain.Actor{Administrator: true}, asset.ID); err != nil || got.Attempt != 1 {
		t.Fatalf("retry asset=%+v err=%v", got, err)
	}
	if len(queue.calls) != 2 {
		t.Fatalf("enqueue calls=%v", queue.calls)
	}
}

func TestDuplicateEnqueueIsSuccessAtTheQueuePortBoundary(t *testing.T) {
	t.Parallel()
	asset := mediaAsset(domain.Created)
	repository := &fakeRepository{appendChunk: func(_ context.Context, _ string, _ int64, _ time.Time, chunkPolicy ports.ChunkPolicy, completePolicy ports.CompletePolicy) (domain.VideoAsset, error) {
		newOffset, err := chunkPolicy(asset)
		if err != nil {
			return domain.VideoAsset{}, err
		}
		asset.ReceivedBytes = newOffset
		asset.State = domain.Uploading
		if err := completePolicy(asset); err != nil {
			return domain.VideoAsset{}, err
		}
		asset.State = domain.Queued
		return asset, nil
	}}
	repository.retry = func(context.Context, domain.Actor, string, time.Duration, time.Time) (domain.VideoAsset, error) {
		asset.Attempt++
		return asset, nil
	}
	queue := &fakeQueue{enqueue: func(context.Context, string, int) error {
		// The queue adapter's contract maps an existing task id to success.
		return nil
	}}
	if _, err := application.NewUpload(repository, defaultStorage(), queue, time.Now).AppendChunk(context.Background(), asset.ID, 0, bytes.NewBufferString("\x00\x00\x00\x18ftypisom")); err != nil {
		t.Fatal(err)
	}
	if _, err := application.NewTickets(repository, queue, time.Now).Retry(context.Background(), domain.Actor{Administrator: true}, asset.ID); err != nil {
		t.Fatal(err)
	}
	if asset.State != domain.Queued || len(queue.calls) != 2 {
		t.Fatalf("state=%s calls=%v", asset.State, queue.calls)
	}
}

func TestCancelListAndDependencyFailuresPassThrough(t *testing.T) {
	t.Parallel()
	want := domain.ErrDependencyUnavailable
	repository := &fakeRepository{
		cancel: func(ctx context.Context, _ string, _ time.Time) error {
			assertMediaDeadline(t, ctx, 3*time.Second)
			return nil
		},
		list: func(ctx context.Context, _ domain.Actor, courseID string) ([]domain.VideoAsset, error) {
			assertMediaDeadline(t, ctx, 3*time.Second)
			if courseID != mediaCourseID {
				t.Fatalf("course=%q", courseID)
			}
			return nil, want
		},
		mint: func(context.Context, domain.Actor, string, int64, string, time.Time, time.Time) (domain.VideoAsset, error) {
			return domain.VideoAsset{}, want
		},
	}
	upload := application.NewUpload(repository, nil, nil, time.Now)
	tickets := application.NewTickets(repository, nil, time.Now)
	if err := upload.CancelUpload(context.Background(), mediaAssetID); err != nil {
		t.Fatal(err)
	}
	if _, err := tickets.ListCourseVideos(context.Background(), domain.Actor{}, mediaCourseID); err != want || err.Error() != "dependency_unavailable" {
		t.Fatalf("list detail escaped: %v", err)
	}
	if _, err := tickets.MintTicket(context.Background(), domain.Actor{}, mediaLessonID, 12, "lesson.mp4"); err != want || err.Error() != "dependency_unavailable" {
		t.Fatalf("mint detail escaped: %v", err)
	}
}

func TestUploadReportsCallerCancellation(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	repository := &fakeRepository{cancel: func(callCtx context.Context, _ string, _ time.Time) error {
		cancel()
		<-callCtx.Done()
		return domain.ErrDependencyUnavailable
	}}
	err := application.NewUpload(repository, nil, nil, time.Now).CancelUpload(ctx, mediaAssetID)
	if !errors.Is(err, context.Canceled) || errors.Is(err, domain.ErrDependencyUnavailable) {
		t.Fatalf("got %v", err)
	}
}

func defaultStorage() *fakeStorage {
	return &fakeStorage{
		free: func(context.Context) (int64, error) { return 1 << 40, nil },
		size: func(_ context.Context, key domain.StorageKey) (int64, error) {
			if key.Kind == domain.StorageSource {
				return 0, os.ErrNotExist
			}
			return 12, nil
		},
		head: func(context.Context, domain.StorageKey, int) ([]byte, error) {
			return []byte("\x00\x00\x00\x18ftypisom"), nil
		},
		append: func(_ context.Context, _ domain.StorageKey, _ int64, body io.Reader) (int64, error) {
			data, err := io.ReadAll(body)
			return int64(len(data)), err
		},
		truncate: func(context.Context, domain.StorageKey, int64) error { return nil },
		move:     func(context.Context, domain.StorageKey, domain.StorageKey) error { return nil },
		publish:  func(context.Context, domain.StorageKey, domain.StorageKey) error { return nil },
		purge:    func(context.Context, domain.StorageKey) error { return nil },
	}
}

func mediaAsset(state domain.State) domain.VideoAsset {
	now := time.Date(2026, 9, 6, 12, 0, 0, 0, time.UTC)
	return domain.VideoAsset{
		ID:            mediaAssetID,
		LessonID:      mediaLessonID,
		Generation:    1,
		State:         state,
		DeclaredBytes: 12,
		OriginalName:  "lesson.mp4",
		ExpiresAt:     now.Add(24 * time.Hour),
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

func assertMediaDeadline(t *testing.T, ctx context.Context, maximum time.Duration) {
	t.Helper()
	deadline, ok := ctx.Deadline()
	remaining := time.Until(deadline)
	if !ok || remaining <= 0 || remaining > maximum {
		t.Fatalf("deadline ok=%v remaining=%v", ok, remaining)
	}
}

func assertMediaDeadlineBetween(t *testing.T, ctx context.Context, minimum, maximum time.Duration) {
	t.Helper()
	deadline, ok := ctx.Deadline()
	remaining := time.Until(deadline)
	if !ok || remaining < minimum || remaining > maximum {
		t.Fatalf("deadline ok=%v remaining=%v, want between %v and %v", ok, remaining, minimum, maximum)
	}
}

func assertStrings(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for index := range got {
		if got[index] != want[index] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}
