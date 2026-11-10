package tus_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"courses/internal/media/adapters/tus"
	"courses/internal/media/domain"
)

const (
	audioTusVersion = "1.0.0"
	audioTusTrackID = "90000000-0000-4000-8000-0000000000a1"
)

type fakeAudioUploadService struct {
	head        func(context.Context, string) (domain.AudioTrack, error)
	appendChunk func(context.Context, string, int64, io.Reader) (domain.AudioTrack, error)
	cancel      func(context.Context, string) error
}

func (service *fakeAudioUploadService) HeadUpload(ctx context.Context, id string) (domain.AudioTrack, error) {
	return service.head(ctx, id)
}

func (service *fakeAudioUploadService) AppendChunk(ctx context.Context, id string, offset int64, body io.Reader) (domain.AudioTrack, error) {
	return service.appendChunk(ctx, id, offset, body)
}

func (service *fakeAudioUploadService) CancelUpload(ctx context.Context, id string) error {
	return service.cancel(ctx, id)
}

func serveAudioTus(service *fakeAudioUploadService, method, id, resumable, contentType, offset, body string) *httptest.ResponseRecorder {
	server := tus.NewAudioServer(service)
	mux := http.NewServeMux()
	server.Register(mux)
	request := httptest.NewRequest(method, "/uploads/audio/"+id, strings.NewReader(body))
	if resumable != "" {
		request.Header.Set("Tus-Resumable", resumable)
	}
	if contentType != "" {
		request.Header.Set("Content-Type", contentType)
	}
	if offset != "" {
		request.Header.Set("Upload-Offset", offset)
	}
	recorder := httptest.NewRecorder()
	mux.ServeHTTP(recorder, request)
	return recorder
}

func uploadingAudioTrack() domain.AudioTrack {
	return domain.AudioTrack{
		ID:            audioTusTrackID,
		Kind:          domain.UploadedTrack,
		State:         domain.Uploading,
		DeclaredBytes: 1024,
		ReceivedBytes: 256,
		ExpiresAt:     time.Date(2026, 9, 6, 12, 0, 0, 0, time.UTC),
	}
}

func TestAudioServerHeadReportsOffsetAndLength(t *testing.T) {
	service := &fakeAudioUploadService{head: func(_ context.Context, id string) (domain.AudioTrack, error) {
		if id != audioTusTrackID {
			t.Fatalf("id=%q", id)
		}
		return uploadingAudioTrack(), nil
	}}
	response := serveAudioTus(service, http.MethodHead, audioTusTrackID, audioTusVersion, "", "", "")
	if response.Code != http.StatusOK {
		t.Fatalf("status=%d", response.Code)
	}
	if response.Header().Get("Upload-Offset") != "256" || response.Header().Get("Upload-Length") != "1024" {
		t.Fatalf("headers=%v", response.Header())
	}
	if response.Header().Get("Upload-Expires") == "" {
		t.Fatal("missing Upload-Expires")
	}
}

func TestAudioServerHeadMapsNotFound(t *testing.T) {
	service := &fakeAudioUploadService{head: func(context.Context, string) (domain.AudioTrack, error) {
		return domain.AudioTrack{}, domain.ErrVideoNotFound
	}}
	response := serveAudioTus(service, http.MethodHead, audioTusTrackID, audioTusVersion, "", "", "")
	if response.Code != http.StatusNotFound {
		t.Fatalf("status=%d", response.Code)
	}
}

func TestAudioServerPatchAppendsAndReportsNewOffset(t *testing.T) {
	service := &fakeAudioUploadService{appendChunk: func(_ context.Context, id string, offset int64, body io.Reader) (domain.AudioTrack, error) {
		if id != audioTusTrackID || offset != 256 {
			t.Fatalf("id=%q offset=%d", id, offset)
		}
		data, err := io.ReadAll(body)
		if err != nil {
			t.Fatal(err)
		}
		track := uploadingAudioTrack()
		track.ReceivedBytes = 256 + int64(len(data))
		return track, nil
	}}
	response := serveAudioTus(service, http.MethodPatch, audioTusTrackID, audioTusVersion, "application/offset+octet-stream", "256", "abcd")
	if response.Code != http.StatusNoContent {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	if response.Header().Get("Upload-Offset") != "260" {
		t.Fatalf("offset header=%q", response.Header().Get("Upload-Offset"))
	}
}

func TestAudioServerPatchRejectsWrongContentType(t *testing.T) {
	service := &fakeAudioUploadService{}
	response := serveAudioTus(service, http.MethodPatch, audioTusTrackID, audioTusVersion, "application/json", "0", "x")
	if response.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("status=%d", response.Code)
	}
}

func TestAudioServerPatchRejectsBadOffset(t *testing.T) {
	service := &fakeAudioUploadService{}
	response := serveAudioTus(service, http.MethodPatch, audioTusTrackID, audioTusVersion, "application/offset+octet-stream", "-1", "x")
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status=%d", response.Code)
	}
}

func TestAudioServerDeleteCancels(t *testing.T) {
	cancelled := false
	service := &fakeAudioUploadService{cancel: func(_ context.Context, id string) error {
		if id != audioTusTrackID {
			t.Fatalf("id=%q", id)
		}
		cancelled = true
		return nil
	}}
	response := serveAudioTus(service, http.MethodDelete, audioTusTrackID, audioTusVersion, "", "", "")
	if response.Code != http.StatusNoContent || !cancelled {
		t.Fatalf("status=%d cancelled=%v", response.Code, cancelled)
	}
}

func TestAudioServerRejectsBadResumableAndMethodAndID(t *testing.T) {
	service := &fakeAudioUploadService{}
	if got := serveAudioTus(service, http.MethodHead, audioTusTrackID, "0.0.0", "", "", ""); got.Code != http.StatusPreconditionFailed {
		t.Fatalf("resumable status=%d", got.Code)
	}
	if got := serveAudioTus(service, http.MethodHead, "not-a-uuid", audioTusVersion, "", "", ""); got.Code != http.StatusNotFound {
		t.Fatalf("bad id status=%d", got.Code)
	}
	got := serveAudioTus(service, http.MethodGet, audioTusTrackID, audioTusVersion, "", "", "")
	if got.Code != http.StatusMethodNotAllowed {
		t.Fatalf("method status=%d", got.Code)
	}
	if allow := got.Header().Get("Allow"); !strings.Contains(allow, "PATCH") {
		t.Fatalf("Allow=%q", allow)
	}
}
