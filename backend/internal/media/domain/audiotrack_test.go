package domain_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"courses/internal/media/domain"
)

const (
	trackA = "60000000-0000-4000-8000-00000000000a"
	trackB = "60000000-0000-4000-8000-00000000000b"
	trackC = "60000000-0000-4000-8000-00000000000c"
)

func validUploadedTrack() domain.AudioTrack {
	now := time.Date(2026, 11, 7, 12, 0, 0, 0, time.UTC)
	return domain.AudioTrack{
		ID:            trackA,
		VideoAssetID:  assetID,
		Kind:          domain.UploadedTrack,
		Language:      "uk",
		Label:         "Українська",
		State:         domain.Created,
		SourceLeaf:    domain.AudioLeafPrefix + trackA,
		DeclaredBytes: 1024 * 1024,
		OriginalName:  "dubbing.aac",
		ExpiresAt:     now.Add(24 * time.Hour),
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

func validEmbeddedTrack() domain.AudioTrack {
	now := time.Date(2026, 11, 7, 12, 0, 0, 0, time.UTC)
	return domain.AudioTrack{
		ID:           trackB,
		VideoAssetID: assetID,
		Kind:         domain.EmbeddedTrack,
		Language:     "en",
		Label:        "Original",
		State:        domain.Ready,
		SourceLeaf:   domain.EmbeddedAudioLeaf,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

func TestNewUploadedTrack(t *testing.T) {
	t.Parallel()
	if _, err := domain.NewUploadedTrack(validUploadedTrack()); err != nil {
		t.Fatalf("valid uploaded track refused: %v", err)
	}

	for _, test := range []struct {
		name   string
		change func(*domain.AudioTrack)
	}{
		{name: "malformed id", change: func(tr *domain.AudioTrack) { tr.ID = "bad" }},
		{name: "malformed asset id", change: func(tr *domain.AudioTrack) { tr.VideoAssetID = "bad" }},
		{name: "wrong kind", change: func(tr *domain.AudioTrack) { tr.Kind = domain.EmbeddedTrack }},
		{name: "not created", change: func(tr *domain.AudioTrack) { tr.State = domain.Ready }},
		{name: "default before ready", change: func(tr *domain.AudioTrack) { tr.IsDefault = true }},
		{name: "leaf not audio-id", change: func(tr *domain.AudioTrack) { tr.SourceLeaf = domain.EmbeddedAudioLeaf }},
		{name: "leaf of other id", change: func(tr *domain.AudioTrack) { tr.SourceLeaf = domain.AudioLeafPrefix + trackB }},
		{name: "empty language", change: func(tr *domain.AudioTrack) { tr.Language = "" }},
		{name: "language too long", change: func(tr *domain.AudioTrack) { tr.Language = strings.Repeat("a", domain.MaximumLanguageLength+1) }},
		{name: "language with space", change: func(tr *domain.AudioTrack) { tr.Language = "uk ua" }},
		{name: "empty label", change: func(tr *domain.AudioTrack) { tr.Label = "" }},
		{name: "label too long", change: func(tr *domain.AudioTrack) { tr.Label = strings.Repeat("ї", domain.MaximumLabelLength+1) }},
		{name: "label with double quote", change: func(tr *domain.AudioTrack) { tr.Label = `Ukr "dub"` }},
		{name: "label with newline", change: func(tr *domain.AudioTrack) { tr.Label = "Ukr\ndub" }},
		{name: "zero declared bytes", change: func(tr *domain.AudioTrack) { tr.DeclaredBytes = 0 }},
		{name: "non-zero received", change: func(tr *domain.AudioTrack) { tr.ReceivedBytes = 1 }},
		{name: "empty original name", change: func(tr *domain.AudioTrack) { tr.OriginalName = "" }},
		{name: "zero expires", change: func(tr *domain.AudioTrack) { tr.ExpiresAt = time.Time{} }},
		{name: "expires before created", change: func(tr *domain.AudioTrack) { tr.ExpiresAt = tr.CreatedAt.Add(-time.Second) }},
		{name: "zero created", change: func(tr *domain.AudioTrack) { tr.CreatedAt = time.Time{} }},
		{name: "update after create", change: func(tr *domain.AudioTrack) { tr.UpdatedAt = tr.UpdatedAt.Add(time.Microsecond) }},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			track := validUploadedTrack()
			test.change(&track)
			if _, err := domain.NewUploadedTrack(track); !errors.Is(err, domain.ErrInvalidRequest) {
				t.Fatalf("got %v, want %v", err, domain.ErrInvalidRequest)
			}
		})
	}
}

func TestNewUploadedTrackTooLarge(t *testing.T) {
	t.Parallel()
	track := validUploadedTrack()
	track.DeclaredBytes = domain.MaximumAudioDeclaredBytes + 1
	if _, err := domain.NewUploadedTrack(track); !errors.Is(err, domain.ErrUploadTooLarge) {
		t.Fatalf("got %v, want %v", err, domain.ErrUploadTooLarge)
	}
}

func TestValidateAudioTicket(t *testing.T) {
	t.Parallel()
	if err := domain.ValidateAudioTicket(1024, "track.aac"); err != nil {
		t.Fatalf("valid ticket refused: %v", err)
	}
	if err := domain.ValidateAudioTicket(domain.MaximumAudioDeclaredBytes, "track.aac"); err != nil {
		t.Fatalf("max-size ticket refused: %v", err)
	}
	if err := domain.ValidateAudioTicket(domain.MaximumAudioDeclaredBytes+1, "track.aac"); !errors.Is(err, domain.ErrUploadTooLarge) {
		t.Fatalf("got %v, want %v", err, domain.ErrUploadTooLarge)
	}
	if err := domain.ValidateAudioTicket(0, "track.aac"); !errors.Is(err, domain.ErrInvalidRequest) {
		t.Fatalf("got %v, want %v", err, domain.ErrInvalidRequest)
	}
	if err := domain.ValidateAudioTicket(1024, ""); !errors.Is(err, domain.ErrInvalidRequest) {
		t.Fatalf("got %v, want %v", err, domain.ErrInvalidRequest)
	}
}

func TestNewEmbeddedTrack(t *testing.T) {
	t.Parallel()
	if _, err := domain.NewEmbeddedTrack(validEmbeddedTrack()); err != nil {
		t.Fatalf("valid embedded track refused: %v", err)
	}
	if _, err := domain.NewEmbeddedTrack(func() domain.AudioTrack { tr := validEmbeddedTrack(); tr.IsDefault = true; return tr }()); err != nil {
		t.Fatalf("default embedded track refused: %v", err)
	}

	for _, test := range []struct {
		name   string
		change func(*domain.AudioTrack)
	}{
		{name: "malformed id", change: func(tr *domain.AudioTrack) { tr.ID = "bad" }},
		{name: "wrong kind", change: func(tr *domain.AudioTrack) { tr.Kind = domain.UploadedTrack }},
		{name: "not ready", change: func(tr *domain.AudioTrack) { tr.State = domain.Created }},
		{name: "leaf not original", change: func(tr *domain.AudioTrack) { tr.SourceLeaf = domain.AudioLeafPrefix + trackB }},
		{name: "empty language", change: func(tr *domain.AudioTrack) { tr.Language = "" }},
		{name: "zero created", change: func(tr *domain.AudioTrack) { tr.CreatedAt = time.Time{} }},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			track := validEmbeddedTrack()
			test.change(&track)
			if _, err := domain.NewEmbeddedTrack(track); !errors.Is(err, domain.ErrInvalidRequest) {
				t.Fatalf("got %v, want %v", err, domain.ErrInvalidRequest)
			}
		})
	}
}

func TestValidateAudioTrackState(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name  string
		kind  domain.AudioTrackKind
		state domain.State
		ok    bool
	}{
		{name: "embedded ready", kind: domain.EmbeddedTrack, state: domain.Ready, ok: true},
		{name: "embedded created", kind: domain.EmbeddedTrack, state: domain.Created},
		{name: "embedded failed", kind: domain.EmbeddedTrack, state: domain.Failed},
		{name: "uploaded created", kind: domain.UploadedTrack, state: domain.Created, ok: true},
		{name: "uploaded uploading", kind: domain.UploadedTrack, state: domain.Uploading, ok: true},
		{name: "uploaded queued", kind: domain.UploadedTrack, state: domain.Queued, ok: true},
		{name: "uploaded processing", kind: domain.UploadedTrack, state: domain.Processing, ok: true},
		{name: "uploaded ready", kind: domain.UploadedTrack, state: domain.Ready, ok: true},
		{name: "uploaded failed", kind: domain.UploadedTrack, state: domain.Failed, ok: true},
		{name: "uploaded awaiting audio", kind: domain.UploadedTrack, state: domain.AwaitingAudio},
		{name: "uploaded unknown", kind: domain.UploadedTrack, state: domain.State("bogus")},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			err := domain.ValidateAudioTrackState(test.kind, test.state)
			if test.ok && err != nil {
				t.Fatalf("got %v, want nil", err)
			}
			if !test.ok && !errors.Is(err, domain.ErrInvalidStateTransition) {
				t.Fatalf("got %v, want %v", err, domain.ErrInvalidStateTransition)
			}
		})
	}
}

func TestAcceptAudioTrackSource(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name  string
		probe domain.SourceProbe
		want  error
	}{
		{name: "aac", probe: domain.SourceProbe{AudioCodec: domain.CodecAAC}},
		{name: "mp3", probe: domain.SourceProbe{AudioCodec: domain.CodecMP3}},
		{name: "no audio stream", probe: domain.SourceProbe{}, want: domain.ErrUnsupportedSource},
		{name: "unsupported codec", probe: domain.SourceProbe{AudioCodec: "opus"}, want: domain.ErrUnsupportedSource},
		{name: "video only behind container", probe: domain.SourceProbe{Width: 1280, Height: 720, VideoCodec: domain.CodecH264}, want: domain.ErrUnsupportedSource},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if err := domain.AcceptAudioTrackSource(test.probe); !errors.Is(err, test.want) {
				t.Fatalf("got %v, want %v", err, test.want)
			}
		})
	}
}

func TestAssetMayBeReady(t *testing.T) {
	t.Parallel()

	if domain.AssetMayBeReady(nil) {
		t.Fatal("empty set must not be Ready-eligible")
	}
	queued := []domain.AudioTrack{{State: domain.Processing}, {State: domain.Failed}}
	if domain.AssetMayBeReady(queued) {
		t.Fatal("no Ready track must not be Ready-eligible")
	}
	if !domain.AssetMayBeReady([]domain.AudioTrack{{State: domain.Failed}, {State: domain.Ready}}) {
		t.Fatal("a Ready track must be Ready-eligible")
	}
}

func TestElectDefault(t *testing.T) {
	t.Parallel()

	early := time.Date(2026, 11, 7, 10, 0, 0, 0, time.UTC)
	late := early.Add(time.Hour)

	if got := domain.ElectDefault([]domain.AudioTrack{{ID: trackA, State: domain.Processing}}); got != "" {
		t.Fatalf("no Ready track: got %q, want empty", got)
	}
	// Earliest CreatedAt wins.
	tracks := []domain.AudioTrack{
		{ID: trackB, State: domain.Ready, CreatedAt: late},
		{ID: trackA, State: domain.Ready, CreatedAt: early},
		{ID: trackC, State: domain.Processing, CreatedAt: early},
	}
	if got := domain.ElectDefault(tracks); got != trackA {
		t.Fatalf("got %q, want %q", got, trackA)
	}
	// Equal CreatedAt falls back to ID order.
	tie := []domain.AudioTrack{
		{ID: trackC, State: domain.Ready, CreatedAt: early},
		{ID: trackA, State: domain.Ready, CreatedAt: early},
	}
	if got := domain.ElectDefault(tie); got != trackA {
		t.Fatalf("tie-break got %q, want %q", got, trackA)
	}
}

func TestValidateDefaultInvariant(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name   string
		tracks []domain.AudioTrack
		state  domain.State
		want   error
	}{
		{name: "pre-ready none default", tracks: []domain.AudioTrack{{State: domain.Processing}}, state: domain.Processing},
		{name: "pre-ready one default", tracks: []domain.AudioTrack{{State: domain.Ready, IsDefault: true}}, state: domain.Processing},
		{name: "pre-ready two defaults", tracks: []domain.AudioTrack{{IsDefault: true}, {IsDefault: true}}, state: domain.Processing, want: domain.ErrInvalidRequest},
		{name: "ready exactly one ready default", tracks: []domain.AudioTrack{{State: domain.Ready, IsDefault: true}, {State: domain.Failed}}, state: domain.Ready},
		{name: "ready no default", tracks: []domain.AudioTrack{{State: domain.Ready}}, state: domain.Ready, want: domain.ErrInvalidRequest},
		{name: "ready default not ready", tracks: []domain.AudioTrack{{State: domain.Processing, IsDefault: true}, {State: domain.Ready}}, state: domain.Ready, want: domain.ErrInvalidRequest},
		{name: "ready two defaults", tracks: []domain.AudioTrack{{State: domain.Ready, IsDefault: true}, {State: domain.Ready, IsDefault: true}}, state: domain.Ready, want: domain.ErrInvalidRequest},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if err := domain.ValidateDefaultInvariant(test.tracks, test.state); !errors.Is(err, test.want) {
				t.Fatalf("got %v, want %v", err, test.want)
			}
		})
	}
}

func TestApplyDefault(t *testing.T) {
	t.Parallel()

	tracks := []domain.AudioTrack{
		{ID: trackA, State: domain.Ready, IsDefault: true},
		{ID: trackB, State: domain.Ready},
	}
	swapped, err := domain.ApplyDefault(tracks, trackB)
	if err != nil {
		t.Fatalf("swap refused: %v", err)
	}
	for _, track := range swapped {
		if (track.ID == trackB) != track.IsDefault {
			t.Fatalf("track %q default = %t, want %t", track.ID, track.IsDefault, track.ID == trackB)
		}
	}
	if tracks[0].IsDefault != true || tracks[1].IsDefault != false {
		t.Fatal("ApplyDefault mutated its input")
	}
	if _, err := domain.ApplyDefault(tracks, trackC); !errors.Is(err, domain.ErrInvalidRequest) {
		t.Fatalf("missing target: got %v, want %v", err, domain.ErrInvalidRequest)
	}
	notReady := []domain.AudioTrack{{ID: trackA, State: domain.Processing}}
	if _, err := domain.ApplyDefault(notReady, trackA); !errors.Is(err, domain.ErrInvalidRequest) {
		t.Fatalf("non-ready target: got %v, want %v", err, domain.ErrInvalidRequest)
	}
}

func TestValidateTrackDeletion(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name    string
		tracks  []domain.AudioTrack
		trackID string
		state   domain.State
		want    error
	}{
		{name: "missing track", tracks: []domain.AudioTrack{{ID: trackA, State: domain.Ready}}, trackID: trackC, state: domain.Ready, want: domain.ErrVideoNotFound},
		{name: "ready asset last ready track", tracks: []domain.AudioTrack{{ID: trackA, State: domain.Ready}, {ID: trackB, State: domain.Failed}}, trackID: trackA, state: domain.Ready, want: domain.ErrInvalidRequest},
		{name: "ready asset keeps a ready track", tracks: []domain.AudioTrack{{ID: trackA, State: domain.Ready}, {ID: trackB, State: domain.Ready}}, trackID: trackA, state: domain.Ready},
		{name: "not ready asset may drop last ready", tracks: []domain.AudioTrack{{ID: trackA, State: domain.Ready}}, trackID: trackA, state: domain.AwaitingAudio},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if err := domain.ValidateTrackDeletion(test.tracks, test.trackID, test.state); !errors.Is(err, test.want) {
				t.Fatalf("got %v, want %v", err, test.want)
			}
		})
	}
}
