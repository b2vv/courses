package domain_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"courses/internal/media/domain"
)

const (
	assetID  = "50000000-0000-4000-8000-000000000001"
	lessonID = "50000000-0000-4000-8000-000000000002"
)

func TestTransitionMatrix(t *testing.T) {
	t.Parallel()

	legal := map[domain.Transition]map[[2]domain.State]bool{
		domain.T1: {{"", domain.Created}: true},
		domain.T2: {{domain.Created, domain.Uploading}: true},
		domain.T3: {{domain.Uploading, domain.Queued}: true},
		domain.T4: {{domain.Queued, domain.Processing}: true},
		domain.T5: {{domain.Processing, domain.Ready}: true},
		domain.T6: {{domain.Processing, domain.Queued}: true},
		domain.T7: {{domain.Processing, domain.Queued}: true},
		domain.T8: {
			{domain.Created, domain.Failed}:    true,
			{domain.Uploading, domain.Failed}:  true,
			{domain.Processing, domain.Failed}: true,
		},
		domain.T9: {
			{domain.Failed, domain.Queued}:     true,
			{domain.Queued, domain.Queued}:     true,
			{domain.Processing, domain.Queued}: true,
		},
	}
	states := []domain.State{"", domain.Created, domain.Uploading, domain.Queued, domain.Processing, domain.Ready, domain.Failed}

	for transition, permitted := range legal {
		transition := transition
		permitted := permitted
		t.Run(string(transition), func(t *testing.T) {
			t.Parallel()
			for _, from := range states {
				for _, to := range states {
					want := permitted[[2]domain.State{from, to}]
					err := domain.ValidateTransition(transition, from, to)
					if want && err != nil {
						t.Errorf("%q -> %q refused: %v", from, to, err)
					}
					if !want && !errors.Is(err, domain.ErrInvalidStateTransition) {
						t.Errorf("%q -> %q: got %v, want %v", from, to, err, domain.ErrInvalidStateTransition)
					}
				}
			}
		})
	}
}

func TestTransitionWriters(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		transition domain.Transition
		want       []domain.TransitionWriter
	}{
		{transition: domain.T1, want: []domain.TransitionWriter{domain.API}},
		{transition: domain.T2, want: []domain.TransitionWriter{domain.UploadService}},
		{transition: domain.T3, want: []domain.TransitionWriter{domain.UploadService}},
		{transition: domain.T4, want: []domain.TransitionWriter{domain.Worker}},
		{transition: domain.T5, want: []domain.TransitionWriter{domain.Worker}},
		{transition: domain.T6, want: []domain.TransitionWriter{domain.Worker}},
		{transition: domain.T7, want: []domain.TransitionWriter{domain.Worker}},
		{transition: domain.T8, want: []domain.TransitionWriter{domain.Worker, domain.UploadService}},
		{transition: domain.T9, want: []domain.TransitionWriter{domain.API}},
	} {
		if got := test.transition.Writers(); !sameWriters(got, test.want) {
			t.Errorf("%s writers = %v, want %v", test.transition, got, test.want)
		}
	}
}

func TestNewAssetTicketBounds(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name   string
		change func(*domain.VideoAsset)
		want   error
	}{
		{name: "minimum size"},
		{name: "maximum size", change: func(asset *domain.VideoAsset) { asset.DeclaredBytes = domain.MaximumDeclaredBytes }},
		{name: "unicode name at character limit", change: func(asset *domain.VideoAsset) {
			asset.OriginalName = strings.Repeat("ї", domain.MaximumOriginalNameLength)
		}},
		{name: "zero size", change: func(asset *domain.VideoAsset) { asset.DeclaredBytes = 0 }, want: domain.ErrInvalidRequest},
		{name: "over maximum size", change: func(asset *domain.VideoAsset) { asset.DeclaredBytes = domain.MaximumDeclaredBytes + 1 }, want: domain.ErrUploadTooLarge},
		{name: "empty name", change: func(asset *domain.VideoAsset) { asset.OriginalName = "" }, want: domain.ErrInvalidRequest},
		{name: "name over character limit", change: func(asset *domain.VideoAsset) {
			asset.OriginalName = strings.Repeat("ї", domain.MaximumOriginalNameLength+1)
		}, want: domain.ErrInvalidRequest},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			asset := validNewAsset()
			if test.change != nil {
				test.change(&asset)
			}
			err := domain.ValidateTicket(asset.DeclaredBytes, asset.OriginalName)
			if !errors.Is(err, test.want) {
				t.Fatalf("got %v, want %v", err, test.want)
			}
		})
	}
}

func TestNewAssetInvariants(t *testing.T) {
	t.Parallel()
	if _, err := domain.NewAsset(validNewAsset()); err != nil {
		t.Fatalf("valid asset refused: %v", err)
	}

	for _, test := range []struct {
		name   string
		change func(*domain.VideoAsset)
	}{
		{name: "malformed asset id", change: func(asset *domain.VideoAsset) { asset.ID = "bad" }},
		{name: "malformed lesson id", change: func(asset *domain.VideoAsset) { asset.LessonID = "bad" }},
		{name: "generation below one", change: func(asset *domain.VideoAsset) { asset.Generation = 0 }},
		{name: "wrong initial state", change: func(asset *domain.VideoAsset) { asset.State = domain.Uploading }},
		{name: "claim already present", change: func(asset *domain.VideoAsset) { asset.ClaimToken = domain.ClaimToken(assetID) }},
		{name: "bytes already received", change: func(asset *domain.VideoAsset) { asset.ReceivedBytes = 1 }},
		{name: "attempt already made", change: func(asset *domain.VideoAsset) { asset.Attempt = 1 }},
		{name: "failure already recorded", change: func(asset *domain.VideoAsset) { asset.FailureReason = domain.UnsupportedSource }},
		{name: "renditions already recorded", change: func(asset *domain.VideoAsset) { asset.Renditions = []domain.Rendition{domain.Rendition360p} }},
		{name: "already evicted", change: func(asset *domain.VideoAsset) { now := asset.CreatedAt; asset.EvictedAt = &now }},
		{name: "zero creation time", change: func(asset *domain.VideoAsset) { asset.CreatedAt = time.Time{} }},
		{name: "different update time", change: func(asset *domain.VideoAsset) { asset.UpdatedAt = asset.UpdatedAt.Add(time.Microsecond) }},
		{name: "expiry not after creation", change: func(asset *domain.VideoAsset) { asset.ExpiresAt = asset.CreatedAt }},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			asset := validNewAsset()
			test.change(&asset)
			if _, err := domain.NewAsset(asset); !errors.Is(err, domain.ErrInvalidRequest) {
				t.Fatalf("got %v, want %v", err, domain.ErrInvalidRequest)
			}
		})
	}
}

func validNewAsset() domain.VideoAsset {
	now := time.Date(2026, 9, 6, 12, 0, 0, 0, time.UTC)
	return domain.VideoAsset{
		ID:            assetID,
		LessonID:      lessonID,
		Generation:    1,
		State:         domain.Created,
		DeclaredBytes: domain.MinimumDeclaredBytes,
		OriginalName:  "lesson.mp4",
		ExpiresAt:     now.Add(24 * time.Hour),
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

func sameWriters(got, want []domain.TransitionWriter) bool {
	if len(got) != len(want) {
		return false
	}
	for index := range got {
		if got[index] != want[index] {
			return false
		}
	}
	return true
}
