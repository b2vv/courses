package application_test

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"courses/internal/media/application"
	"courses/internal/media/domain"
	"courses/internal/media/ports"
)

const audioTrackID = "50000000-0000-4000-8000-0000000000a1"

func uploadedTrack(state domain.State) domain.AudioTrack {
	now := time.Date(2026, 9, 6, 12, 0, 0, 0, time.UTC)
	return domain.AudioTrack{
		ID:            audioTrackID,
		VideoAssetID:  mediaAssetID,
		Kind:          domain.UploadedTrack,
		Language:      "uk",
		Label:         "Українська",
		State:         state,
		SourceLeaf:    "audio-" + audioTrackID,
		DeclaredBytes: 12,
		OriginalName:  "track.aac",
		ExpiresAt:     now.Add(24 * time.Hour),
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

func TestAudioHeadUploadGatesOnStateAndExpiry(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 9, 6, 12, 0, 0, 0, time.UTC)
	clock := func() time.Time { return now }

	cases := []struct {
		name  string
		track domain.AudioTrack
		want  error
	}{
		{name: "created is resumable", track: uploadedTrack(domain.Created)},
		{name: "uploading is resumable", track: uploadedTrack(domain.Uploading)},
		{name: "queued is not resumable", track: uploadedTrack(domain.Queued), want: domain.ErrVideoNotFound},
		{name: "ready is not resumable", track: uploadedTrack(domain.Ready), want: domain.ErrVideoNotFound},
		{name: "expired is not resumable", track: func() domain.AudioTrack {
			track := uploadedTrack(domain.Uploading)
			track.ExpiresAt = now.Add(-time.Minute)
			return track
		}(), want: domain.ErrVideoNotFound},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			repository := &fakeRepository{loadAudioTrackUpload: func(ctx context.Context, trackID string) (domain.AudioTrack, error) {
				assertMediaDeadline(t, ctx, 3*time.Second)
				if trackID != audioTrackID {
					t.Fatalf("trackID=%q", trackID)
				}
				return testCase.track, nil
			}}
			upload := application.NewAudioUpload(repository, defaultStorage(), &fakeQueue{}, clock)
			got, err := upload.HeadUpload(context.Background(), audioTrackID)
			if !errors.Is(err, testCase.want) {
				t.Fatalf("err=%v want=%v", err, testCase.want)
			}
			if testCase.want == nil && got.ID != audioTrackID {
				t.Fatalf("track=%+v", got)
			}
		})
	}
}

func TestAudioAppendChunkPastDeclaredLengthRollsStorageBack(t *testing.T) {
	t.Parallel()
	track := uploadedTrack(domain.Uploading)
	track.ReceivedBytes = track.DeclaredBytes
	truncated := make([]int64, 0)
	storage := defaultStorage()
	storage.truncate = func(_ context.Context, _ domain.StorageKey, size int64) error {
		truncated = append(truncated, size)
		return nil
	}
	repository := &fakeRepository{appendAudioChunk: func(_ context.Context, _ string, _ int64, _ time.Time, chunkPolicy ports.AudioChunkPolicy, _ ports.AudioCompletePolicy) (domain.AudioTrack, error) {
		_, err := chunkPolicy(track)
		return domain.AudioTrack{}, err
	}}
	upload := application.NewAudioUpload(repository, storage, &fakeQueue{}, time.Now)
	_, err := upload.AppendChunk(context.Background(), track.ID, track.ReceivedBytes, bytes.NewBufferString("x"))
	if !errors.Is(err, domain.ErrInvalidRequest) {
		t.Fatalf("err=%v want ErrInvalidRequest", err)
	}
	if len(truncated) == 0 || truncated[len(truncated)-1] != track.ReceivedBytes {
		t.Fatalf("expected rollback truncate to %d, got %v", track.ReceivedBytes, truncated)
	}
}

func TestAudioAppendChunkEnqueuesOnCompletion(t *testing.T) {
	t.Parallel()
	queued := uploadedTrack(domain.Queued)
	queued.ReceivedBytes = queued.DeclaredBytes
	queue := &fakeQueue{}
	repository := &fakeRepository{appendAudioChunk: func(ctx context.Context, _ string, _ int64, _ time.Time, _ ports.AudioChunkPolicy, _ ports.AudioCompletePolicy) (domain.AudioTrack, error) {
		return queued, nil
	}}
	upload := application.NewAudioUpload(repository, defaultStorage(), queue, time.Now)
	got, err := upload.AppendChunk(context.Background(), audioTrackID, queued.DeclaredBytes, bytes.NewBufferString(""))
	if err != nil {
		t.Fatal(err)
	}
	if got.State != domain.Queued {
		t.Fatalf("state=%v", got.State)
	}
	want := mediaAssetID + ":" + audioTrackID
	if len(queue.calls) != 1 || queue.calls[0] != want {
		t.Fatalf("enqueue calls=%v want [%s]", queue.calls, want)
	}
}

func TestAudioAppendChunkDoesNotEnqueueWhileUploading(t *testing.T) {
	t.Parallel()
	queue := &fakeQueue{}
	repository := &fakeRepository{appendAudioChunk: func(context.Context, string, int64, time.Time, ports.AudioChunkPolicy, ports.AudioCompletePolicy) (domain.AudioTrack, error) {
		return uploadedTrack(domain.Uploading), nil
	}}
	upload := application.NewAudioUpload(repository, defaultStorage(), queue, time.Now)
	if _, err := upload.AppendChunk(context.Background(), audioTrackID, 4, bytes.NewBufferString("abcd")); err != nil {
		t.Fatal(err)
	}
	if len(queue.calls) != 0 {
		t.Fatalf("unexpected enqueue while still uploading: %v", queue.calls)
	}
}

func TestAudioAppendChunkPropagatesEnqueueFailure(t *testing.T) {
	t.Parallel()
	queued := uploadedTrack(domain.Queued)
	queued.ReceivedBytes = queued.DeclaredBytes
	sentinel := errors.New("redis down")
	queue := &fakeQueue{enqueueAudio: func(context.Context, string, string, int) error { return sentinel }}
	repository := &fakeRepository{appendAudioChunk: func(context.Context, string, int64, time.Time, ports.AudioChunkPolicy, ports.AudioCompletePolicy) (domain.AudioTrack, error) {
		return queued, nil
	}}
	upload := application.NewAudioUpload(repository, defaultStorage(), queue, time.Now)
	if _, err := upload.AppendChunk(context.Background(), audioTrackID, queued.DeclaredBytes, bytes.NewBufferString("")); !errors.Is(err, sentinel) {
		t.Fatalf("err=%v want sentinel", err)
	}
}

func TestAudioCancelUploadDelegatesWithinDeadline(t *testing.T) {
	t.Parallel()
	called := false
	repository := &fakeRepository{cancelAudioUpload: func(ctx context.Context, trackID string, _ time.Time) error {
		assertMediaDeadline(t, ctx, 3*time.Second)
		if trackID != audioTrackID {
			t.Fatalf("trackID=%q", trackID)
		}
		called = true
		return nil
	}}
	upload := application.NewAudioUpload(repository, defaultStorage(), &fakeQueue{}, time.Now)
	if err := upload.CancelUpload(context.Background(), audioTrackID); err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Fatal("CancelAudioUpload not invoked")
	}
}
