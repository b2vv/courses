//go:build integration

package postgres

import (
	"errors"
	"strings"
	"testing"
	"time"

	"courses/internal/media/domain"
)

// Multi-subtitles MS1 (docs/multi-subtitles/plan-ms1.md): migration 00030, the
// track-language CHECK and the case-insensitive per-asset uniqueness.

const (
	languageShapeVersion  = 30
	languageShapePrevious = 29
)

func languageAsset(n int) string { return "84000000-0000-4000-8000-0000000000" + twoDigits(n) }
func languageTrack(n int) string { return "a4000000-0000-4000-8000-0000000000" + twoDigits(n) }

func (h *mediaHarness) insertTranscript(t *testing.T, lessonID, language string, primary bool) {
	t.Helper()
	if _, err := h.pool.Exec(h.ctx, `INSERT INTO transcripts(lesson_id,language,is_primary) VALUES($1::text::uuid,$2,$3)`,
		lessonID, language, primary); err != nil {
		t.Fatal(err)
	}
}

func (h *mediaHarness) trackLanguage(t *testing.T, trackID string) string {
	t.Helper()
	var language string
	if err := h.pool.QueryRow(h.ctx, `SELECT language FROM audio_tracks WHERE id=$1::text::uuid`, trackID).Scan(&language); err != nil {
		t.Fatal(err)
	}
	return language
}

// atLanguageShapePrevious steps the schema to just before 00030 on reset data and
// restores the head schema (on empty data) when the subtest ends.
func (h *mediaHarness) atLanguageShapePrevious(t *testing.T) {
	t.Helper()
	provider := h.migrationProvider(t)
	h.reset(t)
	if _, err := provider.DownTo(h.ctx, languageShapePrevious); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		h.reset(t)
		if _, err := provider.Up(h.ctx); err != nil {
			t.Errorf("restore head schema: %v", err)
		}
	})
}

func (h *mediaHarness) upToLanguageShape(t *testing.T) error {
	t.Helper()
	_, err := h.migrationProvider(t).UpTo(h.ctx, languageShapeVersion)
	return err
}

// M-R4: the preflight refuses on the ORIGINAL rows — an invalid language, an `und-`
// the relabel would otherwise heal, and a case-only pair the relabel would otherwise
// split — and the refusal leaves every row exactly as it was.
func (h *mediaHarness) testLanguageShapeMigrationRefuses(t *testing.T) {
	embedded := string(domain.EmbeddedTrack)
	uploaded := string(domain.UploadedTrack)
	for _, test := range []struct {
		name    string
		message string
		tracks  [][2]string // language, kind
	}{
		{name: "a language no transcript can use", message: "carry a language no transcript can use", tracks: [][2]string{{"e", embedded}}},
		{name: "und- that the relabel would heal", message: "carry a language no transcript can use", tracks: [][2]string{{"und-", embedded}}},
		{name: "und and UND that the relabel would split", message: "differ only by case", tracks: [][2]string{{"und", embedded}, {"UND", uploaded}}},
	} {
		t.Run(test.name, func(t *testing.T) {
			h.atLanguageShapePrevious(t)
			h.insertAsset(t, languageAsset(1), mediaLessonOne, 1, domain.Ready, mediaNow.Add(time.Hour), nil)
			h.insertTranscript(t, mediaLessonOne, "uk", true)
			for index, track := range test.tracks {
				h.insertTrack(t, languageTrack(index+1), languageAsset(1), track[1], track[0], "Track", index == 0, domain.Ready, mediaNow.Add(time.Duration(index)*time.Minute))
			}
			err := h.upToLanguageShape(t)
			if err == nil || !strings.Contains(err.Error(), test.message) {
				t.Fatalf("migration err = %v, want a refusal containing %q", err, test.message)
			}
			if version := h.schemaVersion(t, h.migrationProvider(t)); version != languageShapePrevious {
				t.Fatalf("schema version after refusal = %d, want %d", version, languageShapePrevious)
			}
			for index, track := range test.tracks {
				if got := h.trackLanguage(t, languageTrack(index+1)); got != track[0] {
					t.Fatalf("track %d language after refusal = %q, want %q unchanged", index+1, got, track[0])
				}
			}
		})
	}
}

// M-R5 + M-R6: the legacy relabel elects exactly one unknown-language track per
// Ready, non-evicted asset whose lesson has a primary transcript and no track in
// that language yet; down keeps the relabelled languages and the old index shape.
func (h *mediaHarness) testLanguageShapeMigrationRelabels(t *testing.T) {
	h.atLanguageShapePrevious(t)
	embedded := string(domain.EmbeddedTrack)
	uploaded := string(domain.UploadedTrack)
	evictedAt := mediaNow.Add(-time.Minute)

	// Lesson one: primary uk. Lesson two: primary en-us.
	h.insertTranscript(t, mediaLessonOne, "uk", true)
	h.insertTranscript(t, mediaLessonOne, "de", false)
	h.insertTranscript(t, mediaLessonTwo, "en-us", true)

	// 1: und default → uk.
	h.insertAsset(t, languageAsset(1), mediaLessonOne, 1, domain.Ready, mediaNow.Add(time.Hour), nil)
	h.insertTrack(t, languageTrack(1), languageAsset(1), embedded, "und", "Original", true, domain.Ready, mediaNow)
	// 2: und (default, uploaded) + und-x-original (embedded): the default wins, the other stays.
	h.insertAsset(t, languageAsset(2), mediaLessonOne, 2, domain.Ready, mediaNow.Add(time.Hour), nil)
	h.insertTrack(t, languageTrack(2), languageAsset(2), embedded, "und-x-original", "Original", false, domain.Ready, mediaNow)
	h.insertTrack(t, languageTrack(3), languageAsset(2), uploaded, "und", "Dub", true, domain.Ready, mediaNow.Add(time.Minute))
	// 3: a uk track already exists → nothing changes.
	h.insertAsset(t, languageAsset(3), mediaLessonOne, 3, domain.Ready, mediaNow.Add(time.Hour), nil)
	h.insertTrack(t, languageTrack(4), languageAsset(3), embedded, "und", "Original", true, domain.Ready, mediaNow)
	h.insertTrack(t, languageTrack(5), languageAsset(3), uploaded, "UK", "Ukr", false, domain.Ready, mediaNow.Add(time.Minute))
	// 4: evicted → nothing changes.
	h.insertAsset(t, languageAsset(4), mediaLessonOne, 4, domain.Ready, mediaNow.Add(time.Hour), &evictedAt)
	h.insertTrack(t, languageTrack(6), languageAsset(4), embedded, "und", "Original", true, domain.Ready, mediaNow)
	// 5: awaiting audio (not Ready) → nothing changes.
	h.insertAsset(t, languageAsset(5), mediaLessonOne, 5, domain.AwaitingAudio, mediaNow.Add(time.Hour), nil)
	h.insertTrack(t, languageTrack(7), languageAsset(5), uploaded, "und", "Dub", false, domain.Created, mediaNow)
	// 6: lesson two, primary en-us → the und track becomes en-us.
	h.insertAsset(t, languageAsset(6), mediaLessonTwo, 1, domain.Ready, mediaNow.Add(time.Hour), nil)
	h.insertTrack(t, languageTrack(8), languageAsset(6), embedded, "UND", "Original", true, domain.Ready, mediaNow)
	// 7: two non-default und tracks: embedded beats uploaded.
	h.insertAsset(t, languageAsset(7), mediaLessonTwo, 2, domain.Ready, mediaNow.Add(time.Hour), nil)
	h.insertTrack(t, languageTrack(9), languageAsset(7), string(domain.UploadedTrack), "pl", "Pol", true, domain.Ready, mediaNow)
	h.insertTrack(t, languageTrack(10), languageAsset(7), uploaded, "und", "Dub", false, domain.Ready, mediaNow.Add(time.Minute))
	h.insertTrack(t, languageTrack(11), languageAsset(7), embedded, "und-x-original", "Original", false, domain.Ready, mediaNow.Add(2*time.Minute))
	// 8: a failed und track is never elected.
	h.insertAsset(t, languageAsset(8), mediaLessonTwo, 3, domain.Ready, mediaNow.Add(time.Hour), nil)
	h.insertTrack(t, languageTrack(12), languageAsset(8), embedded, "fr", "Original", true, domain.Ready, mediaNow)
	h.insertTrack(t, languageTrack(13), languageAsset(8), uploaded, "und", "Broken", false, domain.Failed, mediaNow.Add(time.Minute))

	if err := h.upToLanguageShape(t); err != nil {
		t.Fatalf("migration on clean data: %v", err)
	}
	want := map[int]string{1: "uk", 2: "und-x-original", 3: "uk", 4: "und", 5: "UK", 6: "und", 7: "und", 8: "en-us", 9: "pl", 10: "und", 11: "en-us", 12: "fr", 13: "und"}
	for track, language := range want {
		if got := h.trackLanguage(t, languageTrack(track)); got != language {
			t.Errorf("track %d language = %q, want %q", track, got, language)
		}
	}

	// M-R6: down restores the case-sensitive index and drops the CHECK, keeping the languages.
	provider := h.migrationProvider(t)
	if _, err := provider.DownTo(h.ctx, languageShapePrevious); err != nil {
		t.Fatalf("down 00030: %v", err)
	}
	if got := h.trackLanguage(t, languageTrack(1)); got != "uk" {
		t.Fatalf("down restored track 1 language to %q, want the relabelled uk kept", got)
	}
	h.insertTrack(t, languageTrack(20), languageAsset(1), uploaded, "UK", "Case", false, domain.Ready, mediaNow.Add(time.Hour))
	h.insertTrack(t, languageTrack(21), languageAsset(1), uploaded, "e", "Short", false, domain.Ready, mediaNow.Add(2*time.Hour))
}

// M-R2 + M-R3: the CHECK refuses a bad language on every write path, and the
// per-asset uniqueness ignores case while a failed row still does not count.
func (h *mediaHarness) testLanguageShapeGuards(t *testing.T) {
	h.reset(t)
	h.insertAsset(t, mediaAssetOne, mediaLessonOne, 1, domain.Ready, mediaNow.Add(time.Hour), nil)
	h.insertTrack(t, mediaTrackOne, mediaAssetOne, string(domain.EmbeddedTrack), "en", "Original", true, domain.Ready, mediaNow)
	for _, language := range []string{"e", "1x", "und-", "a--b", "uk_UA", strings.Repeat("x", 35)} {
		if _, err := h.pool.Exec(h.ctx, `INSERT INTO audio_tracks(id,video_asset_id,kind,language,label,is_default,state,source_leaf,declared_bytes,created_at,updated_at)
			VALUES($1::text::uuid,$2::text::uuid,'uploaded',$3,'Bad',false,'created',$4,1024,$5,$5)`,
			mediaTrackThree, mediaAssetOne, language, domain.AudioLeafPrefix+mediaTrackThree, mediaNow); err == nil {
			t.Fatalf("INSERT with language %q accepted", language)
		}
		if _, err := h.pool.Exec(h.ctx, `UPDATE audio_tracks SET language=$2 WHERE id=$1::text::uuid`, mediaTrackOne, language); err == nil {
			t.Fatalf("UPDATE to language %q accepted", language)
		}
	}

	h.insertTrack(t, mediaTrackTwo, mediaAssetOne, string(domain.UploadedTrack), "de", "Ger", false, domain.Ready, mediaNow.Add(time.Minute))
	if _, err := h.repository.UpdateAudioTrackMetadata(h.ctx, ownerActor(), mediaAssetOne, mediaTrackTwo, "EN", "Clash", nil, mediaNow.Add(time.Hour), noMasterRewrite); !errors.Is(err, domain.ErrDuplicateLanguage) {
		t.Fatalf("rename de→EN beside en: err = %v, want duplicate_language", err)
	}
	upload := func(id, language string, state domain.State) (domain.AudioTrack, error) {
		track := domain.AudioTrack{
			ID: id, VideoAssetID: mediaAssetOne, Kind: domain.UploadedTrack, Language: language, Label: "Dub",
			State: state, SourceLeaf: domain.AudioLeafPrefix + id, DeclaredBytes: 1024, OriginalName: "dub.m4a",
			ExpiresAt: mediaNow.Add(time.Hour), CreatedAt: mediaNow, UpdatedAt: mediaNow,
		}
		return h.repository.CreateAudioTrack(h.ctx, track)
	}
	if _, err := upload(mediaTrackThree, "EN", domain.Created); !errors.Is(err, domain.ErrDuplicateLanguage) {
		t.Fatalf("create EN beside en: err = %v, want duplicate_language", err)
	}
	// A failed track does not hold its language.
	if _, err := h.pool.Exec(h.ctx, `UPDATE audio_tracks SET state='failed', is_default=false WHERE id=$1::text::uuid`, mediaTrackTwo); err != nil {
		t.Fatal(err)
	}
	if _, err := h.pool.Exec(h.ctx, `UPDATE audio_tracks SET language='EN' WHERE id=$1::text::uuid`, mediaTrackTwo); err != nil {
		t.Fatalf("failed track renamed to EN beside en refused: %v", err)
	}
}
