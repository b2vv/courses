package ffmpeg

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"courses/internal/media/domain"
	"courses/internal/media/ports"
)

func readGolden(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestRewriteMasterEmbeddedBasePreservesVideoVariants(t *testing.T) {
	t.Parallel()
	// The ffmpeg master already carries the audio group on its STREAM-INF lines;
	// only the EXT-X-MEDIA line is replaced, with the embedded track's own
	// label/language and the "audio/" URI. The measured BANDWIDTH stays verbatim.
	got, err := rewriteMasterText(readGolden(t, "master_with_embedded_audio.m3u8"), []ports.AudioRendition{
		{TrackID: "33333333-3333-4333-8333-333333333333", Name: "Оригінал", Language: "uk", Default: true, Embedded: true},
	})
	if err != nil {
		t.Fatal(err)
	}
	want := "#EXTM3U\n" +
		"#EXT-X-VERSION:7\n" +
		`#EXT-X-MEDIA:TYPE=AUDIO,GROUP-ID="group_aac",NAME="Оригінал",LANGUAGE="uk",DEFAULT=YES,AUTOSELECT=YES,CHANNELS="2",URI="audio/index.m3u8"` + "\n" +
		`#EXT-X-STREAM-INF:BANDWIDTH=377864,AVERAGE-BANDWIDTH=377864,RESOLUTION=640x360,CODECS="avc1.640029,mp4a.40.2",AUDIO="group_aac"` + "\n" +
		"360p/index.m3u8\n" +
		"\n" +
		`#EXT-X-STREAM-INF:BANDWIDTH=1979748,AVERAGE-BANDWIDTH=1979748,RESOLUTION=1280x720,CODECS="avc1.640029,mp4a.40.2",AUDIO="group_aac"` + "\n" +
		"720p/index.m3u8\n"
	if got != want {
		t.Fatalf("mismatch\n got:\n%s\nwant:\n%s", got, want)
	}
}

func TestRewriteMasterVideoOnlyBaseUpgradesStreamInf(t *testing.T) {
	t.Parallel()
	// A video-only base master has no audio group: each STREAM-INF must gain
	// AUDIO="group_aac" and an mp4a codec, and one EXT-X-MEDIA per uploaded track
	// points at its own audio-<trackID>/ playlist.
	got, err := rewriteMasterText(readGolden(t, "master_video_only.m3u8"), []ports.AudioRendition{
		{TrackID: "44444444-4444-4444-8444-444444444444", Name: "Українська", Language: "uk", Default: true},
		{TrackID: "55555555-5555-4555-8555-555555555555", Name: "English", Language: "en"},
	})
	if err != nil {
		t.Fatal(err)
	}
	want := "#EXTM3U\n" +
		"#EXT-X-VERSION:7\n" +
		`#EXT-X-MEDIA:TYPE=AUDIO,GROUP-ID="group_aac",NAME="Українська",LANGUAGE="uk",DEFAULT=YES,AUTOSELECT=YES,CHANNELS="2",URI="audio-44444444-4444-4444-8444-444444444444/index.m3u8"` + "\n" +
		`#EXT-X-MEDIA:TYPE=AUDIO,GROUP-ID="group_aac",NAME="English",LANGUAGE="en",DEFAULT=NO,AUTOSELECT=YES,CHANNELS="2",URI="audio-55555555-5555-4555-8555-555555555555/index.m3u8"` + "\n" +
		`#EXT-X-STREAM-INF:BANDWIDTH=377864,AVERAGE-BANDWIDTH=377864,RESOLUTION=640x360,CODECS="avc1.640029,mp4a.40.2",AUDIO="group_aac"` + "\n" +
		"360p/index.m3u8\n" +
		"\n" +
		`#EXT-X-STREAM-INF:BANDWIDTH=1979748,AVERAGE-BANDWIDTH=1979748,RESOLUTION=1280x720,CODECS="avc1.640029,mp4a.40.2",AUDIO="group_aac"` + "\n" +
		"720p/index.m3u8\n"
	if got != want {
		t.Fatalf("mismatch\n got:\n%s\nwant:\n%s", got, want)
	}
}

func TestRewriteMasterIsIdempotent(t *testing.T) {
	t.Parallel()
	// Feeding the regenerated master back through the same track set must produce
	// byte-identical output (D7 idempotence → safe job retry).
	audios := []ports.AudioRendition{
		{TrackID: "44444444-4444-4444-8444-444444444444", Name: "Українська", Language: "uk", Default: true},
		{TrackID: "55555555-5555-4555-8555-555555555555", Name: "English", Language: "en"},
	}
	once, err := rewriteMasterText(readGolden(t, "master_video_only.m3u8"), audios)
	if err != nil {
		t.Fatal(err)
	}
	twice, err := rewriteMasterText(once, audios)
	if err != nil {
		t.Fatal(err)
	}
	if once != twice {
		t.Fatalf("not idempotent\n once:\n%s\ntwice:\n%s", once, twice)
	}
}

func TestRegenerateMasterWritesAtomicallyWithMode(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	id := "11111111-1111-4111-8111-111111111111"
	hlsDir := filepath.Join(root, "hls", id)
	if err := os.MkdirAll(hlsDir, 0o750); err != nil {
		t.Fatal(err)
	}
	masterPath := filepath.Join(hlsDir, "master.m3u8")
	if err := os.WriteFile(masterPath, []byte(readGolden(t, "master_with_embedded_audio.m3u8")), 0o640); err != nil {
		t.Fatal(err)
	}
	transcoder := New(root)
	err := transcoder.RegenerateMaster(context.Background(), id, []ports.AudioRendition{
		{TrackID: "33333333-3333-4333-8333-333333333333", Name: "Оригінал", Language: "uk", Default: true, Embedded: true},
	})
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(masterPath)
	if err != nil {
		t.Fatal(err)
	}
	if want := `NAME="Оригінал"`; !strings.Contains(string(data), want) {
		t.Fatalf("master does not carry %s:\n%s", want, data)
	}
	info, err := os.Stat(masterPath)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != fileMode {
		t.Fatalf("master mode=%o want %o", info.Mode().Perm(), fileMode)
	}
	entries, err := os.ReadDir(hlsDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("stray files left beside master: %v", entries)
	}
}

func TestRegenerateMasterRejectsBadAssetID(t *testing.T) {
	t.Parallel()
	transcoder := New(t.TempDir())
	if err := transcoder.RegenerateMaster(context.Background(), "not-a-uuid", []ports.AudioRendition{{TrackID: domain.EmbeddedAudioLeaf, Name: "n", Language: "uk", Embedded: true}}); err == nil {
		t.Fatal("bad asset id accepted")
	}
}

func TestRewriteMasterRejectsEmptyAudioAndMissingVariants(t *testing.T) {
	t.Parallel()
	if _, err := rewriteMasterText(readGolden(t, "master_video_only.m3u8"), nil); err == nil {
		t.Fatal("empty audio set accepted")
	}
	if _, err := rewriteMasterText("#EXTM3U\n#EXT-X-VERSION:7\n", []ports.AudioRendition{{TrackID: "x", Name: "n", Language: "uk", Embedded: true}}); err == nil {
		t.Fatal("master without a video variant accepted")
	}
}
