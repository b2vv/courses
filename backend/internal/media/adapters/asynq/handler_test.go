package asynq

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"

	"courses/internal/media/application"
	"courses/internal/media/domain"
	"courses/internal/media/ports"
	"courses/platform/observability"
	asynqlib "github.com/hibiken/asynq"
)

const handlerAssetID = "50000000-0000-4000-8000-000000000001"

type fakeTranscodeProcessor struct {
	job     ports.Job
	outcome application.TranscodeOutcome
	err     error
}

func (processor *fakeTranscodeProcessor) Process(_ context.Context, job ports.Job) (application.TranscodeOutcome, error) {
	processor.job = job
	return processor.outcome, processor.err
}

func TestHandlerMapsTranscodeOutcomes(t *testing.T) {
	dependencyErr := errors.New("dependency failed")
	for _, test := range []struct {
		name       string
		outcome    application.TranscodeOutcome
		processErr error
		wantErr    error
		wantText   string
	}{
		{name: "dependency error retries", processErr: dependencyErr, wantErr: dependencyErr},
		{name: "permanent failure skips retry", outcome: application.TranscodeOutcome{Failed: true, Reason: domain.UnsupportedSource, Permanent: true}, wantErr: asynqlib.SkipRetry},
		{name: "non-permanent failure retries", outcome: application.TranscodeOutcome{Failed: true, Reason: domain.RenderFailed}, wantText: string(domain.RenderFailed)},
		{name: "zero outcome completes"},
	} {
		t.Run(test.name, func(t *testing.T) {
			processor := &fakeTranscodeProcessor{outcome: test.outcome, err: test.processErr}
			err := NewHandler(processor.Process, observability.NewMetrics()).ProcessTask(context.Background(),
				asynqlib.NewTask(TaskTypeTranscode, []byte(`{"assetId":"50000000-0000-4000-8000-000000000001","attempt":2}`)))
			if test.wantErr != nil && !errors.Is(err, test.wantErr) {
				t.Fatalf("got %v, want error %v", err, test.wantErr)
			}
			if test.wantText != "" && (err == nil || err.Error() != test.wantText) {
				t.Fatalf("got %v, want %q", err, test.wantText)
			}
			if test.wantErr == nil && test.wantText == "" && err != nil {
				t.Fatalf("got %v, want nil", err)
			}
			if processor.job.AssetID != handlerAssetID || processor.job.Attempt != 2 {
				t.Fatalf("job=%+v", processor.job)
			}
		})
	}
}

func TestHandlerRejectsInvalidPayloadBeforeProcessing(t *testing.T) {
	for _, payload := range []string{
		`{"assetId":"50000000-0000-4000-8000-000000000001"}`,
		`{"assetId":"not-a-uuid","attempt":0}`,
	} {
		processor := &fakeTranscodeProcessor{}
		err := NewHandler(processor.Process, observability.NewMetrics()).ProcessTask(context.Background(),
			asynqlib.NewTask(TaskTypeTranscode, []byte(payload)))
		if !errors.Is(err, errInvalidTaskPayload) || processor.job.AssetID != "" || errors.Is(err, asynqlib.SkipRetry) {
			t.Fatalf("err=%v job=%+v", err, processor.job)
		}
	}
}

func TestErrorHandlerLogsOnlyBoundedJobDetails(t *testing.T) {
	var output bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&output, nil))
	task := asynqlib.NewTask(TaskTypeTranscode, []byte(`{"assetId":"50000000-0000-4000-8000-000000000001","attempt":2}`))
	NewErrorHandler(logger).HandleError(context.Background(), task, errors.New(string(domain.RenderFailed)))
	logLine := output.String()
	for _, detail := range []string{handlerAssetID, `"state":"failed"`, `"reason":"render_failed"`} {
		if !strings.Contains(logLine, detail) {
			t.Fatalf("log line %q misses %q", logLine, detail)
		}
	}
	output.Reset()
	NewErrorHandler(logger).HandleError(context.Background(), task, errors.New("/media/private source"))
	logLine = output.String()
	if strings.Contains(logLine, "private") {
		t.Fatalf("private error detail escaped: %s", logLine)
	}
	if !strings.Contains(logLine, `"reason":"dependency_unavailable"`) {
		t.Fatalf("dependency reason is not bounded: %s", logLine)
	}
	output.Reset()
	invalidTask := asynqlib.NewTask(TaskTypeTranscode, []byte(`{"assetId":"/media/private/source","attempt":0}`))
	NewErrorHandler(logger).HandleError(context.Background(), invalidTask, errInvalidTaskPayload)
	logLine = output.String()
	if strings.Contains(logLine, "/media/private") || !strings.Contains(logLine, `"assetId":""`) {
		t.Fatalf("invalid payload escaped into the log: %s", logLine)
	}
}
