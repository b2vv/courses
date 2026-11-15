package whisper

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"os/exec"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"courses/internal/transcript/domain"
	"courses/internal/transcript/ports"
)

const (
	processWaitDelay = 5 * time.Second

	// stderrLimit bounds the buffer that collects engine diagnostics. faster-whisper logs model
	// loading, progress and warnings to stderr; collecting without a limit would let a runaway
	// process fill memory. 64 KiB matches the OS pipe buffer so one blocking write cannot
	// wedge the child even without a concurrent reader.
	stderrLimit = 64 * 1024
)

// faster-whisper 1.2.1 tokenizer._LANGUAGE_CODES, read in the asr-worker image with:
// /opt/whisper/bin/python3 -c "from faster_whisper.tokenizer import _LANGUAGE_CODES"
// Refresh this list when upgrading the engine.
const languageCodes = "af am ar as az ba be bg bn bo br bs ca cs cy da de el en es et eu fa fi fo fr gl gu ha haw he hi hr ht hu hy id is it ja jw ka kk km kn ko la lb ln lo lt lv mg mi mk ml mn mr ms mt my ne nl nn no oc pa pl ps pt ro ru sa sd si sk sl sn so sq sr su sv sw ta te tg th tk tl tr tt uk ur uz vi yi yo yue zh"

// Engine implements ports.Transcriber by running a Python subprocess that invokes
// faster-whisper. Each call spawns a fresh process — the model is NOT kept resident between
// tasks (plan P-4): at Concurrency 1 the ~10 s reload is negligible against minutes of work,
// and a resident model would hold 5 GiB between tasks.
type Engine struct {
	python string // path to the Python interpreter (e.g. "python3")
	script string // path to the transcription script
	model  string // path to the model weights directory
	logger *slog.Logger
}

func NewEngine(python, script, model string, logger *slog.Logger) *Engine {
	return &Engine{
		python: python,
		script: script,
		model:  model,
		logger: logger,
	}
}

// scriptOutput is the JSON contract with the Python script. It is private to the adapter: the
// port speaks domain types, and the script is free to evolve its wire format as long as this
// struct follows.
type scriptOutput struct {
	Language      string          `json:"language"`
	SpeechSeconds float64         `json:"speech_seconds"`
	Segments      []scriptSegment `json:"segments"`
	ModelLoadSec  *float64        `json:"model_load_sec,omitempty"`
	TranscribeSec *float64        `json:"transcribe_sec,omitempty"`
}

type scriptSegment struct {
	ID         string   `json:"id"`
	StartMs    int64    `json:"start_ms"`
	EndMs      int64    `json:"end_ms"`
	Text       string   `json:"text"`
	Confidence *float64 `json:"confidence,omitempty"`
}

func (engine *Engine) Transcribe(ctx context.Context, audioPath string, trackLanguage domain.Language, multilingual bool) (ports.Recognition, error) {
	arguments := []string{engine.script, "--model", engine.model, "--audio", audioPath}
	if multilingual {
		arguments = append(arguments, "--multilingual")
	} else {
		primary := strings.SplitN(strings.ToLower(string(trackLanguage)), "-", 2)[0]
		switch {
		case primary == "und": // The und family deliberately requests auto-detection.
		case strings.Contains(" "+languageCodes+" ", " "+primary+" "):
			arguments = append(arguments, "--language", primary)
		default:
			engine.logger.Warn("hint_unsupported", "tag", string(trackLanguage))
		}
	}

	command := exec.CommandContext(ctx, engine.python, arguments...)
	command.WaitDelay = processWaitDelay

	var stdout bytes.Buffer
	command.Stdout = &stdout

	stderr := &boundedWriter{limit: stderrLimit}
	command.Stderr = stderr

	if err := command.Start(); err != nil {
		return ports.Recognition{}, fmt.Errorf("start recognition engine: %w", err)
	}

	err := command.Wait()

	// Log stderr diagnostics regardless of outcome. The text is engine internals (model loading
	// times, CUDA warnings, VAD progress) and must NOT contain recognised speech — the Python
	// script enforces that by writing segments only to stdout.
	if stderrText := stderr.String(); stderrText != "" {
		engine.logger.Info("recognition engine stderr",
			slog.String("audio", audioPath),
			slog.String("stderr", stderrText),
		)
	}

	if err != nil {
		if contextErr := ctx.Err(); contextErr != nil {
			return ports.Recognition{}, contextErr
		}
		return ports.Recognition{}, fmt.Errorf("recognition engine exited with error: %w", err)
	}

	return parseOutput(stdout.Bytes())
}

func parseOutput(raw []byte) (ports.Recognition, error) {
	var output scriptOutput
	if err := json.Unmarshal(raw, &output); err != nil {
		return ports.Recognition{}, fmt.Errorf("parse recognition output: %w", err)
	}

	if err := validateOutput(output); err != nil {
		return ports.Recognition{}, err
	}

	var language domain.Language
	if output.Language != "" {
		normalised, err := domain.NewLanguage(output.Language)
		if err != nil {
			return ports.Recognition{}, fmt.Errorf("recognition reported invalid language %q: %w", output.Language, err)
		}
		language = normalised
	}

	segments := make([]domain.Segment, len(output.Segments))
	for index, raw := range output.Segments {
		segments[index] = domain.Segment{
			ID:         raw.ID,
			StartMs:    raw.StartMs,
			EndMs:      raw.EndMs,
			Text:       raw.Text,
			Confidence: raw.Confidence,
		}
	}

	return ports.Recognition{
		Language:      language,
		Segments:      segments,
		SpeechSeconds: output.SpeechSeconds,
	}, nil
}

func validateOutput(output scriptOutput) error {
	if output.SpeechSeconds < 0 || math.IsNaN(output.SpeechSeconds) || math.IsInf(output.SpeechSeconds, 0) {
		return fmt.Errorf("recognition reported invalid speech_seconds: %f", output.SpeechSeconds)
	}
	for index, segment := range output.Segments {
		if !utf8.ValidString(segment.Text) {
			return fmt.Errorf("segment %d text is not valid UTF-8", index)
		}
		if segment.ID == "" {
			return fmt.Errorf("segment %d has no id", index)
		}
	}
	return nil
}

// boundedWriter collects up to `limit` bytes and silently drops the rest. It is safe for
// concurrent use because exec.Cmd may write to Stderr from the pipe-copying goroutine while
// the main goroutine reads it after Wait returns — but in practice the write side is done by
// the time Wait returns, so the mutex is a safety net, not a hot path.
type boundedWriter struct {
	limit int
	mu    sync.Mutex
	buf   bytes.Buffer
}

func (w *boundedWriter) Write(p []byte) (int, error) {
	n := len(p)
	w.mu.Lock()
	defer w.mu.Unlock()
	remaining := w.limit - w.buf.Len()
	if remaining <= 0 {
		return n, nil
	}
	if len(p) > remaining {
		p = p[:remaining]
	}
	w.buf.Write(p)
	return n, nil
}

func (w *boundedWriter) String() string {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.buf.String()
}

// Compile-time check.
var _ ports.Transcriber = (*Engine)(nil)

// ErrEngineUnavailable is returned when the engine binary cannot be found or started.
var ErrEngineUnavailable = errors.New("recognition engine unavailable")
