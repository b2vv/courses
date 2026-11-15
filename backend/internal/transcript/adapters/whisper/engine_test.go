package whisper

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"courses/internal/transcript/domain"
)

func TestParseOutputAcceptsValidRecognition(t *testing.T) {
	t.Parallel()
	raw := `{
		"language": "uk",
		"speech_seconds": 87.5,
		"segments": [
			{"id": "11111111-1111-4111-8111-111111111111", "start_ms": 0, "end_ms": 1500, "text": "Привіт", "confidence": 0.95},
			{"id": "22222222-2222-4222-8222-222222222222", "start_ms": 1500, "end_ms": 3000, "text": "світ"}
		]
	}`
	recognition, err := parseOutput([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	if recognition.Language != "uk" {
		t.Fatalf("language=%q want=uk", recognition.Language)
	}
	if recognition.SpeechSeconds != 87.5 {
		t.Fatalf("speech_seconds=%f want=87.5", recognition.SpeechSeconds)
	}
	if len(recognition.Segments) != 2 {
		t.Fatalf("segments=%d want=2", len(recognition.Segments))
	}
	if recognition.Segments[0].Text != "Привіт" {
		t.Fatalf("segment[0].text=%q want=Привіт", recognition.Segments[0].Text)
	}
	if recognition.Segments[1].Confidence != nil {
		t.Fatalf("segment[1].confidence=%v want=nil", recognition.Segments[1].Confidence)
	}
	if recognition.Segments[0].Confidence == nil || *recognition.Segments[0].Confidence != 0.95 {
		t.Fatal("segment[0].confidence must be 0.95")
	}
}

func TestParseOutputNormalisesLanguage(t *testing.T) {
	t.Parallel()
	raw := `{"language": " EN ", "speech_seconds": 1.0, "segments": []}`
	recognition, err := parseOutput([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	if recognition.Language != "en" {
		t.Fatalf("language=%q want=en", recognition.Language)
	}
}

func TestParseOutputAcceptsEmptyLanguageForAutoDetection(t *testing.T) {
	t.Parallel()
	// When the model detects no speech, language may be absent.
	raw := `{"language": "", "speech_seconds": 0.0, "segments": []}`
	recognition, err := parseOutput([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	if recognition.Language != "" {
		t.Fatalf("language=%q want=empty", recognition.Language)
	}
}

func TestParseOutputRefusesBrokenJSON(t *testing.T) {
	t.Parallel()
	if _, err := parseOutput([]byte(`not json`)); err == nil {
		t.Fatal("broken JSON must be refused")
	}
}

func TestParseOutputRefusesInvalidSpeechSeconds(t *testing.T) {
	t.Parallel()
	for _, raw := range []string{
		`{"language": "uk", "speech_seconds": -1.0, "segments": []}`,
		`{"language": "uk", "speech_seconds": "NaN", "segments": []}`,
	} {
		if _, err := parseOutput([]byte(raw)); err == nil {
			t.Fatalf("invalid speech_seconds must be refused: %s", raw)
		}
	}
}

func TestParseOutputRefusesMissingSegmentID(t *testing.T) {
	t.Parallel()
	raw := `{"language": "uk", "speech_seconds": 1.0, "segments": [{"id": "", "start_ms": 0, "end_ms": 100, "text": "hello"}]}`
	if _, err := parseOutput([]byte(raw)); err == nil {
		t.Fatal("a segment without an id must be refused")
	}
}

func TestParseOutputRefusesInvalidLanguage(t *testing.T) {
	t.Parallel()
	raw := `{"language": "!!!", "speech_seconds": 1.0, "segments": []}`
	if _, err := parseOutput([]byte(raw)); err == nil {
		t.Fatal("an invalid language must be refused")
	}
}

func TestTranscribeRunsScriptAndReturnsRecognition(t *testing.T) {
	t.Parallel()
	scriptPath := writeScript(t, `#!/bin/sh
cat << 'JSON'
{"language":"en","speech_seconds":2.5,"segments":[{"id":"11111111-1111-4111-8111-111111111111","start_ms":0,"end_ms":2500,"text":"Hello world"}]}
JSON
`)
	engine := NewEngine("/bin/sh", scriptPath, "/fake/model", noopLogger())
	recognition, err := engine.Transcribe(context.Background(), "/fake/audio.wav", "und", false)
	if err != nil {
		t.Fatal(err)
	}
	if recognition.Language != "en" {
		t.Fatalf("language=%q want=en", recognition.Language)
	}
	if recognition.SpeechSeconds != 2.5 {
		t.Fatalf("speech_seconds=%f want=2.5", recognition.SpeechSeconds)
	}
	if len(recognition.Segments) != 1 || recognition.Segments[0].Text != "Hello world" {
		t.Fatalf("unexpected segments: %+v", recognition.Segments)
	}
}

func TestTranscribeDerivesHintFromTrack(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		language     domain.Language
		multilingual bool
		hint         string
		unsupported  bool
	}{
		{language: "en-US", hint: "en"},
		{language: "de", hint: "de"},
		{language: "haw", hint: "haw"},
		{language: "und"},
		{language: "und-x-original"},
		{language: "tlh", unsupported: true},
		{language: "de", multilingual: true},
	} {
		t.Run(string(test.language)+"/"+fmt.Sprint(test.multilingual), func(t *testing.T) {
			scriptPath := writeScript(t, `#!/bin/sh
printf '%s\n' "$@" >&2
printf '%s\n' '{"language":"en","speech_seconds":2,"segments":[]}'
`)
			var logs bytes.Buffer
			engine := NewEngine("/bin/sh", scriptPath, "/fake/model", slog.New(slog.NewJSONHandler(&logs, nil)))
			if _, err := engine.Transcribe(context.Background(), "/fake/audio.wav", test.language, test.multilingual); err != nil {
				t.Fatal(err)
			}
			var args []string
			unsupported := false
			decoder := json.NewDecoder(&logs)
			for decoder.More() {
				var record struct {
					Message string `json:"msg"`
					Level   string `json:"level"`
					Tag     string `json:"tag"`
					Stderr  string `json:"stderr"`
				}
				if err := decoder.Decode(&record); err != nil {
					t.Fatal(err)
				}
				if record.Message == "hint_unsupported" {
					unsupported = record.Level == "WARN" && record.Tag == string(test.language)
				}
				if record.Stderr != "" {
					args = strings.Fields(record.Stderr)
				}
			}
			if unsupported != test.unsupported {
				t.Fatalf("hint_unsupported=%v want %v", unsupported, test.unsupported)
			}
			want := []string{"--model", "/fake/model", "--audio", "/fake/audio.wav"}
			if test.multilingual {
				want = append(want, "--multilingual")
			}
			if test.hint != "" {
				want = append(want, "--language", test.hint)
			}
			if !slices.Equal(args, want) {
				t.Fatalf("argv=%q want %q", args, want)
			}
		})
	}
}

func TestLanguageCodesPinnedToFasterWhisper121(t *testing.T) {
	t.Parallel()
	codes := strings.Fields(languageCodes)
	if len(codes) != 100 {
		t.Fatalf("language codes=%d want 100", len(codes))
	}
	seen := make(map[string]bool)
	for _, code := range codes {
		if seen[code] {
			t.Fatalf("duplicate code %q", code)
		}
		seen[code] = true
	}
}

// M2-R12: execute the production script, observing the actual Python model-call kwargs.
func TestRealTranscribePythonModelBoundary(t *testing.T) {
	python, err := exec.LookPath("python3")
	if err != nil {
		t.Fatalf("M2-R12 requires python3: %v", err)
	}
	stubPath, err := filepath.Abs("testdata")
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("PYTHONPATH", stubPath+string(os.PathListSeparator)+os.Getenv("PYTHONPATH"))
	t.Setenv("PYTHONDONTWRITEBYTECODE", "1")
	for _, test := range []struct {
		name         string
		language     domain.Language
		multilingual bool
		hint         *string
	}{
		{name: "multilingual", language: "de", multilingual: true},
		{name: "regional", language: "en-US", hint: new("en")},
		{name: "undetermined", language: "und"},
	} {
		t.Run(test.name, func(t *testing.T) {
			capture := filepath.Join(t.TempDir(), "kwargs.json")
			t.Setenv("WHISPER_TEST_KWARGS", capture)
			engine := NewEngine(python, "transcribe.py", "/fake/model", noopLogger())
			if _, err := engine.Transcribe(context.Background(), "/fake/audio.wav", test.language, test.multilingual); err != nil {
				t.Fatal(err)
			}
			raw, err := os.ReadFile(capture)
			if err != nil {
				t.Fatal(err)
			}
			var kwargs map[string]json.RawMessage
			if err := json.Unmarshal(raw, &kwargs); err != nil {
				t.Fatal(err)
			}
			var multilingual bool
			if err := json.Unmarshal(kwargs["multilingual"], &multilingual); err != nil {
				t.Fatal(err)
			}
			if multilingual != test.multilingual {
				t.Fatalf("multilingual=%v want %v", multilingual, test.multilingual)
			}
			wantLanguage, err := json.Marshal(test.hint)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(kwargs["language"], wantLanguage) {
				t.Fatalf("language=%s want %s", kwargs["language"], wantLanguage)
			}
		})
	}
}

func TestTranscribeReturnsContextErrorOnCancellation(t *testing.T) {
	// A long-running script is cancelled promptly.
	scriptPath := writeScript(t, `#!/bin/sh
sleep 30
`)
	engine := NewEngine("/bin/sh", scriptPath, "/fake/model", noopLogger())

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		_, err := engine.Transcribe(ctx, "/fake/audio.wav", "und", false)
		done <- err
	}()

	// Give the process time to start, then cancel.
	time.Sleep(100 * time.Millisecond)
	cancelledAt := time.Now()
	cancel()

	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("error=%v want=context.Canceled", err)
		}
		if elapsed := time.Since(cancelledAt); elapsed > processWaitDelay+time.Second {
			t.Fatalf("cancellation took %s; WaitDelay=%s", elapsed, processWaitDelay)
		}
	case <-time.After(processWaitDelay + 2*time.Second):
		t.Fatal("Transcribe did not return after cancellation")
	}
}

func TestTranscribeReportsNonZeroExit(t *testing.T) {
	t.Parallel()
	scriptPath := writeScript(t, `#!/bin/sh
echo "model not found" >&2
exit 1
`)
	engine := NewEngine("/bin/sh", scriptPath, "/fake/model", noopLogger())
	_, err := engine.Transcribe(context.Background(), "/fake/audio.wav", "und", false)
	if err == nil {
		t.Fatal("non-zero exit must be an error")
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("error must not be a context error: %v", err)
	}
}

func TestBoundedWriterDropsBeyondLimit(t *testing.T) {
	t.Parallel()
	w := &boundedWriter{limit: 10}
	n, err := w.Write([]byte("hello world — this is longer than 10 bytes"))
	if err != nil {
		t.Fatal(err)
	}
	// Write reports accepting ALL bytes so the child process is not blocked.
	if n != len("hello world — this is longer than 10 bytes") {
		t.Fatalf("n=%d want=%d", n, len("hello world — this is longer than 10 bytes"))
	}
	if got := w.String(); got != "hello worl" {
		t.Fatalf("buffer=%q want=%q", got, "hello worl")
	}
}

func TestBoundedWriterAcceptsExactlyTheLimit(t *testing.T) {
	t.Parallel()
	w := &boundedWriter{limit: 5}
	w.Write([]byte("abc"))
	w.Write([]byte("de"))
	if got := w.String(); got != "abcde" {
		t.Fatalf("buffer=%q want=%q", got, "abcde")
	}
	// One more byte should be silently dropped.
	w.Write([]byte("f"))
	if got := w.String(); got != "abcde" {
		t.Fatalf("buffer=%q want=%q after overflow", got, "abcde")
	}
}

func TestBoundedWriterReportsFullLengthOnOverflow(t *testing.T) {
	t.Parallel()
	w := &boundedWriter{limit: 3}
	// The first write exceeds the limit — the reported `n` must still be the full length,
	// otherwise the child process sees a short write and may fail or hang.
	n, err := w.Write([]byte("abcdef"))
	if err != nil || n != 6 {
		t.Fatalf("Write=%d,%v want=6,nil", n, err)
	}
	if got := w.String(); got != "abc" {
		t.Fatalf("buffer=%q want=%q", got, "abc")
	}
}

// helpers

func writeScript(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "script.sh")
	if err := os.WriteFile(path, []byte(content), 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}

func noopLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestParseOutputHandlesZeroSegments(t *testing.T) {
	t.Parallel()
	// No-speech produces zero segments and zero speech_seconds.
	raw := `{"language": "", "speech_seconds": 0.0, "segments": []}`
	recognition, err := parseOutput([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	if len(recognition.Segments) != 0 {
		t.Fatalf("segments=%d want=0", len(recognition.Segments))
	}
}

func TestParseOutputPreservesTimestampsAndText(t *testing.T) {
	t.Parallel()
	raw := `{
		"language": "uk",
		"speech_seconds": 10.0,
		"segments": [
			{"id": "11111111-1111-4111-8111-111111111111", "start_ms": 1234, "end_ms": 5678, "text": "текст із символами: 'ї', \"є\""}
		]
	}`
	recognition, err := parseOutput([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	segment := recognition.Segments[0]
	if segment.StartMs != 1234 || segment.EndMs != 5678 {
		t.Fatalf("timestamps: start=%d end=%d", segment.StartMs, segment.EndMs)
	}
	if !strings.Contains(segment.Text, "'ї'") {
		t.Fatalf("text lost Ukrainian characters: %q", segment.Text)
	}
}
