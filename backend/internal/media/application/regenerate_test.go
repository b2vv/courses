package application_test

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"testing"
	"time"

	"courses/internal/media/application"
	"courses/internal/media/domain"
	"courses/internal/media/ports"
)

// MS1 M-R8: one asset's failure does not stop the others, an asset that vanished
// between the listing and its rewrite is skipped rather than failed, and an
// empty Ready set leaves the master untouched (same rule as a rename).
func TestRegenerateAllMastersKeepsGoingAndClassifies(t *testing.T) {
	const (
		good    = "84000000-0000-4000-8000-000000000001"
		broken  = "84000000-0000-4000-8000-000000000002"
		gone    = "84000000-0000-4000-8000-000000000003"
		silent  = "84000000-0000-4000-8000-000000000004"
		another = "84000000-0000-4000-8000-000000000005"
	)
	ready := []domain.AudioTrack{{ID: "a4000000-0000-4000-8000-000000000001", Language: "uk", Label: "Original", IsDefault: true, State: domain.Ready, SourceLeaf: domain.EmbeddedAudioLeaf}}
	repository := &fakeRepository{
		listReadyAssetIDs: func(context.Context) ([]string, error) { return []string{good, broken, gone, silent, another}, nil },
		regenerateMaster: func(ctx context.Context, assetID string, rewrite func([]domain.AudioTrack) error) error {
			switch assetID {
			case gone:
				return domain.ErrVideoNotFound
			case silent:
				return rewrite(nil)
			default:
				return rewrite(ready)
			}
		},
	}
	var written []string
	writer := &fakeTranscoder{regenerateMaster: func(_ context.Context, assetID string, audios []ports.AudioRendition) error {
		if assetID == broken {
			return errors.New("disk full")
		}
		if len(audios) != 1 || audios[0].Language != "uk" {
			t.Errorf("asset %s rewritten with %+v", assetID, audios)
		}
		written = append(written, assetID)
		return nil
	}}
	service := application.NewAudioTracks(repository, writer, false, func() time.Time { return time.Now() })

	result, err := service.RegenerateAllMasters(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if result.Regenerated != 3 || !slices.Equal(result.Failed, []string{broken}) || !slices.Equal(result.Skipped, []string{gone}) {
		t.Fatalf("result = %+v, want 3 regenerated (good, silent, another), failed [broken], skipped [gone]", result)
	}
	if !slices.Equal(written, []string{good, another}) {
		t.Fatalf("masters written = %v, want good and another (silent has no Ready track)", written)
	}

	repository.listReadyAssetIDs = func(context.Context) ([]string, error) { return nil, errors.New("database down") }
	if _, err := service.RegenerateAllMasters(context.Background()); err == nil {
		t.Fatal("a failed listing must be an error")
	}
}

func TestMintAudioTicketCarriesMultilingualDefault(t *testing.T) {
	t.Parallel()
	for _, multilingual := range []bool{false, true} {
		t.Run(fmt.Sprint(multilingual), func(t *testing.T) {
			repository := &fakeRepository{}
			service := application.NewAudioTracks(repository, nil, multilingual, func() time.Time { return time.Date(2026, 9, 6, 12, 0, 0, 0, time.UTC) })
			track, err := service.MintAudioTicket(context.Background(), domain.Actor{}, mediaAssetID, 12, "dub.aac", "de", "Deutsch", nil)
			if err != nil {
				t.Fatal(err)
			}
			if repository.mintedMultilingual != multilingual || track.Multilingual != multilingual {
				t.Fatalf("repository=%v returned=%v want %v", repository.mintedMultilingual, track.Multilingual, multilingual)
			}
		})
	}
}
