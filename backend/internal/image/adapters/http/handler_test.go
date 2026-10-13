package http_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"image/png"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	imagehttp "courses/internal/image/adapters/http"
	"courses/internal/image/domain"
	"courses/platform/httpserver"
)

type fakeUploader struct {
	image       domain.Image
	err         error
	calls       int
	actor       domain.Actor
	lessonID    string
	contentType string
	width       int
	height      int
	data        []byte
}

func (service *fakeUploader) Upload(_ context.Context, actor domain.Actor, lessonID, contentType string, width, height int, data []byte) (domain.Image, error) {
	service.calls++
	service.actor, service.lessonID, service.contentType = actor, lessonID, contentType
	service.width, service.height, service.data = width, height, data
	return service.image, service.err
}

type fakeAuthorizer struct {
	err     error
	calls   int
	actor   domain.Actor
	imageID string
}

func (service *fakeAuthorizer) Authorize(_ context.Context, actor domain.Actor, imageID string) error {
	service.calls++
	service.actor, service.imageID = actor, imageID
	return service.err
}

func passThrough(next http.Handler) http.Handler { return next }

func rejectedAuthentication(status int) func(http.Handler) http.Handler {
	return func(http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(status)
		})
	}
}

func authenticated(actor httpserver.Actor) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r.WithContext(httpserver.WithActor(r.Context(), actor)))
		})
	}
}

func uploadRequest(t *testing.T, data []byte, missingFile bool) *http.Request {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if !missingFile {
		file, err := writer.CreateFormFile("file", "x.png")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := file.Write(data); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodPost, "/api/v1/admin/lessons/lesson-1/images", &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	return request
}

func TestUpload(t *testing.T) {
	t.Parallel()
	var fixture bytes.Buffer
	if err := png.Encode(&fixture, image.NewRGBA(image.Rect(0, 0, 1, 1))); err != nil {
		t.Fatal(err)
	}
	actor := httpserver.Actor{UserID: "owner-1", Role: "lecturer", EmailVerified: true}
	admin := httpserver.Actor{UserID: "admin-1", Role: "admin"}
	tests := []struct {
		name         string
		actor        httpserver.Actor
		data         []byte
		failure      error
		missingFile  bool
		missingActor bool
		wantStatus   int
		wantCode     string
		wantCalls    int
	}{
		{name: "owner success", actor: actor, data: fixture.Bytes(), wantStatus: 201, wantCalls: 1},
		{name: "admin actor forwarded", actor: admin, data: fixture.Bytes(), wantStatus: 201, wantCalls: 1},
		{name: "forbidden", actor: actor, data: fixture.Bytes(), failure: domain.ErrForbidden, wantStatus: 403, wantCode: "forbidden", wantCalls: 1},
		{name: "lesson not found", actor: actor, data: fixture.Bytes(), failure: domain.ErrLessonNotFound, wantStatus: 404, wantCode: "lesson_not_found", wantCalls: 1},
		{name: "unsupported", actor: actor, data: []byte("GIF89a............"), wantStatus: 415, wantCode: "unsupported_image"},
		{name: "undecodable PNG", actor: actor, data: []byte("\x89PNG\r\n\x1a\ngarbage"), wantStatus: 400, wantCode: "invalid_image"},
		{name: "missing file", actor: actor, missingFile: true, wantStatus: 400, wantCode: "invalid_request"},
		{name: "missing actor", missingActor: true, data: fixture.Bytes(), wantStatus: 403, wantCode: "forbidden"},
		{name: "request too large", actor: actor, data: make([]byte, domain.MaximumImageBytes+(512<<10)), wantStatus: 413, wantCode: "image_too_large"},
		{name: "domain unsupported", actor: actor, data: fixture.Bytes(), failure: domain.ErrUnsupportedImage, wantStatus: 415, wantCode: "unsupported_image", wantCalls: 1},
		{name: "domain too large", actor: actor, data: fixture.Bytes(), failure: domain.ErrImageTooLarge, wantStatus: 413, wantCode: "image_too_large", wantCalls: 1},
		{name: "domain too many pixels", actor: actor, data: fixture.Bytes(), failure: domain.ErrImageTooManyPixels, wantStatus: 400, wantCode: "image_too_many_pixels", wantCalls: 1},
		{name: "domain invalid image", actor: actor, data: fixture.Bytes(), failure: domain.ErrInvalidImage, wantStatus: 400, wantCode: "invalid_image", wantCalls: 1},
		{name: "wrapped forbidden", actor: actor, data: fixture.Bytes(), failure: fmt.Errorf("upload: %w", domain.ErrForbidden), wantStatus: 403, wantCode: "forbidden", wantCalls: 1},
		{name: "dependency failure", actor: actor, data: fixture.Bytes(), failure: errors.New("unavailable"), wantStatus: 503, wantCode: "dependency_unavailable", wantCalls: 1},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			service := &fakeUploader{image: domain.Image{ID: "img-1", ContentType: "image/png", Width: 1, Height: 1}, err: test.failure}
			adapter := imagehttp.NewHandler(service, &fakeAuthorizer{})
			authenticate := authenticated(test.actor)
			if test.missingActor {
				authenticate = passThrough
			}
			handler := httpserver.NewHandler(httpserver.Options{Service: "api", Register: func(mux *http.ServeMux) {
				adapter.RegisterAdmin(mux, authenticate, passThrough)
				adapter.RegisterInternal(mux, authenticate)
			}})
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, uploadRequest(t, test.data, test.missingFile))
			if response.Code != test.wantStatus || service.calls != test.wantCalls {
				t.Fatalf("status=%d calls=%d body=%q, want %d and %d", response.Code, service.calls, response.Body.String(), test.wantStatus, test.wantCalls)
			}
			if test.wantCode != "" && !strings.Contains(response.Body.String(), `"code":"`+test.wantCode+`"`) {
				t.Fatalf("body=%q, want code %q", response.Body.String(), test.wantCode)
			}
			if test.wantStatus == http.StatusCreated && strings.TrimSpace(response.Body.String()) != `{"id":"img-1","contentType":"image/png","width":1,"height":1}` {
				t.Fatalf("body=%q", response.Body.String())
			}
			if test.wantCalls == 1 {
				wantActor := domain.Actor{UserID: test.actor.UserID, Administrator: test.actor.Role == "admin", EmailVerified: test.actor.EmailVerified}
				if service.actor != wantActor || service.lessonID != "lesson-1" || service.contentType != "image/png" || service.width != 1 || service.height != 1 || !bytes.Equal(service.data, test.data) {
					t.Fatalf("unexpected upload arguments: actor=%+v lesson=%q type=%q dimensions=%dx%d bytes=%d", service.actor, service.lessonID, service.contentType, service.width, service.height, len(service.data))
				}
			}
		})
	}
}

func TestAuthorize(t *testing.T) {
	t.Parallel()
	actor := httpserver.Actor{UserID: "student-1", Role: "student", EmailVerified: true}
	tests := []struct {
		name         string
		authenticate func(http.Handler) http.Handler
		failure      error
		wantStatus   int
		wantCode     string
		wantCalls    int
	}{
		{name: "allowed", authenticate: authenticated(actor), wantStatus: 204, wantCalls: 1},
		{name: "email not verified", authenticate: authenticated(actor), failure: domain.ErrEmailNotVerified, wantStatus: 403, wantCode: "email_not_verified", wantCalls: 1},
		{name: "forbidden", authenticate: authenticated(actor), failure: domain.ErrForbidden, wantStatus: 403, wantCode: "forbidden", wantCalls: 1},
		{name: "image not found collapses to 403", authenticate: authenticated(actor), failure: domain.ErrImageNotFound, wantStatus: 403, wantCode: "forbidden", wantCalls: 1},
		{name: "wrapped forbidden", authenticate: authenticated(actor), failure: fmt.Errorf("authorize: %w", domain.ErrForbidden), wantStatus: 403, wantCode: "forbidden", wantCalls: 1},
		{name: "dependency failure", authenticate: authenticated(actor), failure: errors.New("unavailable"), wantStatus: 503, wantCode: "dependency_unavailable", wantCalls: 1},
		{name: "authentication denied", authenticate: rejectedAuthentication(401), wantStatus: 403},
		{name: "authentication unavailable", authenticate: rejectedAuthentication(503), wantStatus: 503},
		{name: "missing actor", authenticate: passThrough, wantStatus: 403, wantCode: "forbidden"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			service := &fakeAuthorizer{err: test.failure}
			adapter := imagehttp.NewHandler(&fakeUploader{}, service)
			handler := httpserver.NewHandler(httpserver.Options{Service: "api", Register: func(mux *http.ServeMux) {
				adapter.RegisterAdmin(mux, test.authenticate, passThrough)
				adapter.RegisterInternal(mux, test.authenticate)
			}})
			request := httptest.NewRequest(http.MethodGet, "/internal/image/authorize", nil)
			request.Header.Set("X-Image-Id", "img-1")
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			if response.Code != test.wantStatus || service.calls != test.wantCalls {
				t.Fatalf("status=%d calls=%d, want %d and %d", response.Code, service.calls, test.wantStatus, test.wantCalls)
			}
			if test.wantCode != "" && !strings.Contains(response.Body.String(), `"code":"`+test.wantCode+`"`) {
				t.Fatalf("body=%q, want code %q", response.Body.String(), test.wantCode)
			}
			if test.wantStatus == http.StatusNoContent && response.Body.Len() != 0 {
				t.Fatalf("body=%q, want empty", response.Body.String())
			}
			if test.wantCalls == 1 && (service.actor != (domain.Actor{UserID: actor.UserID, EmailVerified: true}) || service.imageID != "img-1") {
				t.Fatalf("actor=%+v image=%q", service.actor, service.imageID)
			}
			if (test.wantCalls == 1 || test.name == "missing actor") && response.Header().Get("Cache-Control") != "no-store" {
				t.Fatalf("Cache-Control=%q", response.Header().Get("Cache-Control"))
			}
		})
	}
}
