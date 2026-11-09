package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	identityhttp "courses/internal/identity/adapters/http"
	identityapplication "courses/internal/identity/application"
	identitydomain "courses/internal/identity/domain"
	mediahttp "courses/internal/media/adapters/http"
	mediaapplication "courses/internal/media/application"
	mediadomain "courses/internal/media/domain"
	"courses/internal/media/ports"
	"courses/platform/httpserver"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/legacy"
)

const (
	mediaAdminID      = "40000000-0000-4000-8000-000000000003"
	mediaInstructorID = "40000000-0000-4000-8000-000000000002"
	mediaOtherOwnerID = "40000000-0000-4000-8000-000000000004"
	mediaCourseID     = "10000000-0000-4000-8000-000000000001"
	mediaLessonID     = "30000000-0000-4000-8000-000000000001"
	mediaVideoID      = "50000000-0000-4000-8000-000000000001"
	mediaTrackID      = "60000000-0000-4000-8000-000000000001"
)

var mediaNow = time.Date(2026, 9, 6, 12, 0, 0, 123456000, time.UTC)

type mediaRepository struct {
	mintCalls, listCalls, retryCalls int
	mint                             func(context.Context, mediadomain.Actor, string, int64, string, time.Time, time.Time) (mediadomain.VideoAsset, error)
	list                             func(context.Context, mediadomain.Actor, string) ([]mediadomain.VideoAsset, error)
	retry                            func(context.Context, mediadomain.Actor, string, time.Duration, time.Time) (mediadomain.VideoAsset, error)
	listTracks                       func(context.Context, mediadomain.Actor, string) ([]mediadomain.AudioTrack, error)
	setDefault                       func(context.Context, mediadomain.Actor, string, string, time.Time, func([]mediadomain.AudioTrack) error) error
	updateMetadata                   func(context.Context, mediadomain.Actor, string, string, string, string, time.Time, func([]mediadomain.AudioTrack) error) (mediadomain.AudioTrack, error)
	deleteTrack                      func(context.Context, mediadomain.Actor, string, string, time.Time, func([]mediadomain.AudioTrack) error) error
}

func (repository *mediaRepository) Mint(ctx context.Context, actor mediadomain.Actor, lessonID string, declaredBytes int64, originalName string, expiresAt, now time.Time) (mediadomain.VideoAsset, error) {
	repository.mintCalls++
	return repository.mint(ctx, actor, lessonID, declaredBytes, originalName, expiresAt, now)
}

func (repository *mediaRepository) ListCourseVideos(ctx context.Context, actor mediadomain.Actor, courseID string) ([]mediadomain.VideoAsset, error) {
	repository.listCalls++
	return repository.list(ctx, actor, courseID)
}

func (repository *mediaRepository) Retry(ctx context.Context, actor mediadomain.Actor, id string, claimAge time.Duration, now time.Time) (mediadomain.VideoAsset, error) {
	repository.retryCalls++
	return repository.retry(ctx, actor, id, claimAge, now)
}

func (*mediaRepository) Load(context.Context, string) (mediadomain.VideoAsset, error) {
	panic("unexpected Load")
}

func (*mediaRepository) AppendChunk(context.Context, string, int64, time.Time, ports.ChunkPolicy, ports.CompletePolicy) (mediadomain.VideoAsset, error) {
	panic("unexpected AppendChunk")
}

func (*mediaRepository) Cancel(context.Context, string, time.Time) error { panic("unexpected Cancel") }

func (*mediaRepository) Claim(context.Context, string, mediadomain.ClaimToken, time.Time) (mediadomain.VideoAsset, bool, error) {
	panic("unexpected Claim")
}

func (*mediaRepository) Publish(context.Context, string, mediadomain.ClaimToken, []mediadomain.Rendition, *float64, time.Time, *mediadomain.AudioTrack, ports.PublishPolicy) (mediadomain.VideoAsset, *mediadomain.VideoAsset, error) {
	panic("unexpected Publish")
}

func (*mediaRepository) RegenerateMaster(context.Context, string, func([]mediadomain.AudioTrack) error) error {
	panic("unexpected RegenerateMaster")
}

func (*mediaRepository) LoadAudioTrack(context.Context, string, string) (mediadomain.AudioTrack, error) {
	panic("unexpected LoadAudioTrack")
}

func (*mediaRepository) PromoteAudioTrack(context.Context, string, string, time.Time, func([]mediadomain.AudioTrack) error) (*mediadomain.VideoAsset, error) {
	panic("unexpected PromoteAudioTrack")
}

func (*mediaRepository) FailAudioTrack(context.Context, string, string, time.Time) error {
	panic("unexpected FailAudioTrack")
}

func (repository *mediaRepository) ListAudioTracks(ctx context.Context, actor mediadomain.Actor, assetID string) ([]mediadomain.AudioTrack, error) {
	if repository.listTracks != nil {
		return repository.listTracks(ctx, actor, assetID)
	}
	panic("unexpected ListAudioTracks")
}

func (repository *mediaRepository) SetDefaultAudioTrack(ctx context.Context, actor mediadomain.Actor, assetID, trackID string, now time.Time, rewrite func([]mediadomain.AudioTrack) error) error {
	if repository.setDefault != nil {
		return repository.setDefault(ctx, actor, assetID, trackID, now, rewrite)
	}
	panic("unexpected SetDefaultAudioTrack")
}

func (repository *mediaRepository) UpdateAudioTrackMetadata(ctx context.Context, actor mediadomain.Actor, assetID, trackID, language, label string, now time.Time, rewrite func([]mediadomain.AudioTrack) error) (mediadomain.AudioTrack, error) {
	if repository.updateMetadata != nil {
		return repository.updateMetadata(ctx, actor, assetID, trackID, language, label, now, rewrite)
	}
	panic("unexpected UpdateAudioTrackMetadata")
}

func (repository *mediaRepository) DeleteAudioTrack(ctx context.Context, actor mediadomain.Actor, assetID, trackID string, now time.Time, rewrite func([]mediadomain.AudioTrack) error) error {
	if repository.deleteTrack != nil {
		return repository.deleteTrack(ctx, actor, assetID, trackID, now, rewrite)
	}
	panic("unexpected DeleteAudioTrack")
}

func (*mediaRepository) Requeue(context.Context, string, mediadomain.ClaimToken, time.Time) error {
	panic("unexpected Requeue")
}

func (*mediaRepository) Fail(context.Context, string, mediadomain.State, mediadomain.ClaimToken, mediadomain.FailureReason, time.Time) error {
	panic("unexpected Fail")
}

func (*mediaRepository) SweepCandidates(context.Context, time.Time, int) ([]mediadomain.VideoAsset, error) {
	panic("unexpected SweepCandidates")
}

func (*mediaRepository) DeleteSwept(context.Context, string) error { panic("unexpected DeleteSwept") }

func (*mediaRepository) Recover(context.Context, time.Time, int) ([]mediadomain.VideoAsset, error) {
	panic("unexpected Recover")
}

func (repository *mediaRepository) calls() int {
	return repository.mintCalls + repository.listCalls + repository.retryCalls
}

type queuedCall struct {
	assetID string
	attempt int
}

type mediaQueue struct {
	calls []queuedCall
	err   error
}

func (queue *mediaQueue) Enqueue(_ context.Context, assetID string, attempt int) error {
	queue.calls = append(queue.calls, queuedCall{assetID: assetID, attempt: attempt})
	return queue.err
}

func (queue *mediaQueue) EnqueueAudioTrack(_ context.Context, assetID, _ string, attempt int) error {
	queue.calls = append(queue.calls, queuedCall{assetID: assetID, attempt: attempt})
	return queue.err
}

type mediaMasterWriter struct{}

func (*mediaMasterWriter) RegenerateMaster(_ context.Context, _ string, _ []ports.AudioRendition) error {
	return nil
}

type mediaIdentity struct {
	user              identitydomain.User
	err               error
	authenticateCalls int
}

func (*mediaIdentity) Login(context.Context, identityapplication.LoginInput) (identityapplication.LoginResult, error) {
	return identityapplication.LoginResult{}, errors.New("unexpected login")
}

func (identity *mediaIdentity) Authenticate(context.Context, string) (identitydomain.User, error) {
	identity.authenticateCalls++
	return identity.user, identity.err
}

func (*mediaIdentity) Logout(context.Context, string) error { return errors.New("unexpected logout") }

func TestMediaMintAndListContracts(t *testing.T) {
	t.Run("P01 admin mints a metadata-only ticket and reads created", func(t *testing.T) {
		var stored []mediadomain.VideoAsset
		repository := defaultMediaRepository()
		repository.mint = func(_ context.Context, actor mediadomain.Actor, lessonID string, declaredBytes int64, originalName string, expiresAt, now time.Time) (mediadomain.VideoAsset, error) {
			if actor != (mediadomain.Actor{UserID: mediaAdminID, Administrator: true}) || lessonID != mediaLessonID || declaredBytes != 5*1024*1024 || originalName != "lesson.mp4" || !expiresAt.Equal(now.Add(24*time.Hour)) {
				t.Fatalf("mint input actor=%+v lesson=%q bytes=%d name=%q expires=%v now=%v", actor, lessonID, declaredBytes, originalName, expiresAt, now)
			}
			asset := mediaAsset(mediaVideoID, mediadomain.Created, 1)
			asset.DeclaredBytes, asset.ReceivedBytes, asset.OriginalName, asset.ExpiresAt, asset.CreatedAt, asset.UpdatedAt = declaredBytes, 0, originalName, expiresAt, now, now
			stored = append(stored, asset)
			return asset, nil
		}
		repository.list = func(_ context.Context, actor mediadomain.Actor, courseID string) ([]mediadomain.VideoAsset, error) {
			if actor != (mediadomain.Actor{UserID: mediaAdminID, Administrator: true}) || courseID != mediaCourseID {
				t.Fatalf("list input actor=%+v course=%q", actor, courseID)
			}
			return stored, nil
		}
		identity := &mediaIdentity{user: mediaUser(mediaAdminID, identitydomain.Admin)}
		handler := mediaAPIHandler(t, repository, &mediaQueue{}, identity)

		response := serveMedia(t, handler, identity, mediaUser(mediaAdminID, identitydomain.Admin), true, "http://localhost:8080", http.MethodPost, "/api/v1/admin/lessons/"+mediaLessonID+"/video", []byte(`{"declaredBytes":5242880,"originalName":"lesson.mp4"}`), "application/json", true)
		if response.Code != http.StatusCreated {
			t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
		}
		var ticket struct {
			VideoID, UploadURL, ExpiresAt string
			ChunkBytes, MaxBytes          int64
		}
		decodeMediaJSON(t, response, &ticket)
		if ticket.VideoID != mediaVideoID || ticket.UploadURL != "/uploads/"+mediaVideoID || ticket.ChunkBytes != 8*1024*1024 || ticket.MaxBytes != mediadomain.MaximumDeclaredBytes || ticket.ExpiresAt != "2026-09-07T12:00:00.123456Z" {
			t.Fatalf("ticket=%+v", ticket)
		}

		response = serveMedia(t, handler, identity, mediaUser(mediaAdminID, identitydomain.Admin), true, "", http.MethodGet, "/api/v1/admin/courses/"+mediaCourseID+"/videos", nil, "", true)
		if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"state":"created"`) {
			t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
		}
	})

	t.Run("P02 owning instructor mints", func(t *testing.T) {
		repository := defaultMediaRepository()
		repository.mint = func(_ context.Context, actor mediadomain.Actor, _ string, _ int64, _ string, _, _ time.Time) (mediadomain.VideoAsset, error) {
			if actor != (mediadomain.Actor{UserID: mediaInstructorID}) {
				t.Fatalf("actor=%+v", actor)
			}
			return mediaAsset(mediaVideoID, mediadomain.Created, 1), nil
		}
		identity := &mediaIdentity{}
		response := serveMedia(t, mediaAPIHandler(t, repository, &mediaQueue{}, identity), identity, mediaUser(mediaInstructorID, identitydomain.Instructor), true, "http://localhost:8080", http.MethodPost, "/api/v1/admin/lessons/"+mediaLessonID+"/video", []byte(`{"declaredBytes":12,"originalName":"lesson.mp4"}`), "application/json", true)
		if response.Code != http.StatusCreated || repository.mintCalls != 1 {
			t.Fatalf("status=%d calls=%d body=%s", response.Code, repository.mintCalls, response.Body.String())
		}
	})

	t.Run("P07 one DTO per asset and no storage member", func(t *testing.T) {
		repository := defaultMediaRepository()
		first := mediaAsset(mediaVideoID, mediadomain.Ready, 2)
		second := mediaAsset("50000000-0000-4000-8000-000000000002", mediadomain.Ready, 1)
		repository.list = func(context.Context, mediadomain.Actor, string) ([]mediadomain.VideoAsset, error) {
			return []mediadomain.VideoAsset{first, second}, nil
		}
		identity := &mediaIdentity{}
		response := serveMedia(t, mediaAPIHandler(t, repository, &mediaQueue{}, identity), identity, mediaUser(mediaAdminID, identitydomain.Admin), true, "", http.MethodGet, "/api/v1/admin/courses/"+mediaCourseID+"/videos", nil, "", true)
		var page struct {
			Items []map[string]json.RawMessage `json:"items"`
		}
		decodeMediaJSON(t, response, &page)
		if response.Code != http.StatusOK || len(page.Items) != 2 {
			t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
		}
		for _, item := range page.Items {
			if len(item) != 8 || item["lessonId"] == nil || item["videoId"] == nil || item["state"] == nil || item["generation"] == nil || item["originalName"] == nil || item["receivedBytes"] == nil || item["declaredBytes"] == nil || item["updatedAt"] == nil {
				t.Fatalf("unexpected media DTO members: %v", item)
			}
			for _, forbidden := range []string{"storageKey", "path", "source", "renditions", "claimToken"} {
				if item[forbidden] != nil {
					t.Fatalf("%s leaked in %v", forbidden, item)
				}
			}
		}
	})

	t.Run("B15 a course without videos returns a non-null empty list", func(t *testing.T) {
		repository := defaultMediaRepository()
		repository.list = func(context.Context, mediadomain.Actor, string) ([]mediadomain.VideoAsset, error) { return nil, nil }
		identity := &mediaIdentity{}
		response := serveMedia(t, mediaAPIHandler(t, repository, &mediaQueue{}, identity), identity, mediaUser(mediaAdminID, identitydomain.Admin), true, "", http.MethodGet, "/api/v1/admin/courses/"+mediaCourseID+"/videos", nil, "", true)
		if response.Code != http.StatusOK || response.Body.String() != "{\"items\":[]}\n" {
			t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
		}
	})
}

func TestMediaMintBoundsAndStrictBody(t *testing.T) {
	tests := []struct {
		name, path, body, contentType, code string
		status                              int
		wantCalls                           int
		validRequest                        bool
	}{
		{name: "B01 exactly 2 GiB", path: "/api/v1/admin/lessons/" + mediaLessonID + "/video", body: `{"declaredBytes":2147483648,"originalName":"lesson.mp4"}`, contentType: "application/json", status: 201, wantCalls: 1, validRequest: true},
		{name: "B01 over 2 GiB", path: "/api/v1/admin/lessons/" + mediaLessonID + "/video", body: `{"declaredBytes":2147483649,"originalName":"lesson.mp4"}`, contentType: "application/json", status: 400, code: "upload_too_large"},
		{name: "B04 zero bytes", path: "/api/v1/admin/lessons/" + mediaLessonID + "/video", body: `{"declaredBytes":0,"originalName":"lesson.mp4"}`, contentType: "application/json", status: 400, code: "invalid_request"},
		{name: "255 character original name", path: "/api/v1/admin/lessons/" + mediaLessonID + "/video", body: `{"declaredBytes":12,"originalName":"` + strings.Repeat("ї", 255) + `"}`, contentType: "application/json", status: 201, wantCalls: 1, validRequest: true},
		{name: "empty original name", path: "/api/v1/admin/lessons/" + mediaLessonID + "/video", body: `{"declaredBytes":12,"originalName":""}`, contentType: "application/json", status: 400, code: "invalid_request"},
		{name: "256 character original name", path: "/api/v1/admin/lessons/" + mediaLessonID + "/video", body: `{"declaredBytes":12,"originalName":"` + strings.Repeat("ї", 256) + `"}`, contentType: "application/json", status: 400, code: "invalid_request"},
		{name: "unknown member", path: "/api/v1/admin/lessons/" + mediaLessonID + "/video", body: `{"declaredBytes":12,"originalName":"lesson.mp4","contentType":"video/mp4"}`, contentType: "application/json", status: 400, code: "invalid_request"},
		{name: "duplicate member", path: "/api/v1/admin/lessons/" + mediaLessonID + "/video", body: `{"declaredBytes":12,"declaredBytes":13,"originalName":"lesson.mp4"}`, contentType: "application/json", status: 400, code: "invalid_request"},
		{name: "malformed body", path: "/api/v1/admin/lessons/" + mediaLessonID + "/video", body: `{`, contentType: "application/json", status: 400, code: "invalid_request"},
		{name: "missing member", path: "/api/v1/admin/lessons/" + mediaLessonID + "/video", body: `{"declaredBytes":12}`, contentType: "application/json", status: 400, code: "invalid_request"},
		{name: "null member", path: "/api/v1/admin/lessons/" + mediaLessonID + "/video", body: `{"declaredBytes":null,"originalName":"lesson.mp4"}`, contentType: "application/json", status: 400, code: "invalid_request"},
		{name: "fractional bytes", path: "/api/v1/admin/lessons/" + mediaLessonID + "/video", body: `{"declaredBytes":12.5,"originalName":"lesson.mp4"}`, contentType: "application/json", status: 400, code: "invalid_request"},
		{name: "oversized body", path: "/api/v1/admin/lessons/" + mediaLessonID + "/video", body: `{"declaredBytes":12,"originalName":"` + strings.Repeat("x", 9*1024) + `"}`, contentType: "application/json", status: 400, code: "invalid_request"},
		{name: "wrong content type", path: "/api/v1/admin/lessons/" + mediaLessonID + "/video", body: `{"declaredBytes":12,"originalName":"lesson.mp4"}`, contentType: "text/plain", status: 400, code: "invalid_request"},
		{name: "query parameter", path: "/api/v1/admin/lessons/" + mediaLessonID + "/video?x=1", body: `{"declaredBytes":12,"originalName":"lesson.mp4"}`, contentType: "application/json", status: 400, code: "invalid_request"},
		{name: "R06 malformed lesson id", path: "/api/v1/admin/lessons/not-a-uuid/video", body: `{"declaredBytes":12,"originalName":"lesson.mp4"}`, contentType: "application/json", status: 400, code: "invalid_request"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repository := defaultMediaRepository()
			identity := &mediaIdentity{}
			response := serveMedia(t, mediaAPIHandler(t, repository, &mediaQueue{}, identity), identity, mediaUser(mediaAdminID, identitydomain.Admin), true, "http://localhost:8080", http.MethodPost, test.path, []byte(test.body), test.contentType, test.validRequest)
			if response.Code != test.status || repository.mintCalls != test.wantCalls {
				t.Fatalf("status=%d calls=%d body=%s", response.Code, repository.mintCalls, response.Body.String())
			}
			if test.code != "" {
				assertMediaError(t, response, test.code)
			}
		})
	}

	t.Run("invalid UTF-8", func(t *testing.T) {
		repository := defaultMediaRepository()
		identity := &mediaIdentity{}
		body := append([]byte(`{"declaredBytes":12,"originalName":"`), 0xff)
		body = append(body, []byte(`"}`)...)
		response := serveMedia(t, mediaAPIHandler(t, repository, &mediaQueue{}, identity), identity, mediaUser(mediaAdminID, identitydomain.Admin), true, "http://localhost:8080", http.MethodPost, "/api/v1/admin/lessons/"+mediaLessonID+"/video", body, "application/json", false)
		if response.Code != http.StatusBadRequest || repository.mintCalls != 0 {
			t.Fatalf("status=%d calls=%d body=%s", response.Code, repository.mintCalls, response.Body.String())
		}
		assertMediaError(t, response, "invalid_request")
	})
}

func TestMediaAuthorizationRefusals(t *testing.T) {
	operations := []struct {
		method, path, body, origin string
	}{
		{http.MethodPost, "/api/v1/admin/lessons/" + mediaLessonID + "/video", `{"declaredBytes":12,"originalName":"lesson.mp4"}`, "http://localhost:8080"},
		{http.MethodGet, "/api/v1/admin/courses/" + mediaCourseID + "/videos", "", ""},
		{http.MethodPost, "/api/v1/admin/videos/" + mediaVideoID + "/retry", "", "http://localhost:8080"},
	}
	for _, caller := range []struct {
		name   string
		role   identitydomain.Role
		cookie bool
		status int
		code   string
	}{
		{name: "R01 anonymous", status: 401, code: "authentication_required"},
		{name: "R02 student", role: identitydomain.Student, cookie: true, status: 403, code: "forbidden"},
	} {
		for _, operation := range operations {
			t.Run(caller.name+" "+operation.method+" "+operation.path, func(t *testing.T) {
				repository := defaultMediaRepository()
				identity := &mediaIdentity{}
				response := serveMedia(t, mediaAPIHandler(t, repository, &mediaQueue{}, identity), identity, mediaUser(mediaAdminID, caller.role), caller.cookie, operation.origin, operation.method, operation.path, []byte(operation.body), contentType(operation.body), true)
				if response.Code != caller.status || repository.calls() != 0 {
					t.Fatalf("status=%d calls=%d body=%s", response.Code, repository.calls(), response.Body.String())
				}
				assertMediaError(t, response, caller.code)
			})
		}
	}

	for _, operation := range operations {
		t.Run("R03 non-owner "+operation.method+" "+operation.path, func(t *testing.T) {
			repository := defaultMediaRepository()
			repository.mint = func(_ context.Context, actor mediadomain.Actor, _ string, _ int64, _ string, _, _ time.Time) (mediadomain.VideoAsset, error) {
				return mediadomain.VideoAsset{}, actor.CanEditLesson(mediaOtherOwnerID)
			}
			repository.list = func(_ context.Context, actor mediadomain.Actor, _ string) ([]mediadomain.VideoAsset, error) {
				return nil, actor.CanEditLesson(mediaOtherOwnerID)
			}
			repository.retry = func(_ context.Context, actor mediadomain.Actor, _ string, _ time.Duration, _ time.Time) (mediadomain.VideoAsset, error) {
				return mediadomain.VideoAsset{}, actor.CanEditLesson(mediaOtherOwnerID)
			}
			identity := &mediaIdentity{}
			response := serveMedia(t, mediaAPIHandler(t, repository, &mediaQueue{}, identity), identity, mediaUser(mediaInstructorID, identitydomain.Instructor), true, operation.origin, operation.method, operation.path, []byte(operation.body), contentType(operation.body), true)
			if response.Code != http.StatusForbidden || repository.calls() != 1 {
				t.Fatalf("status=%d calls=%d body=%s", response.Code, repository.calls(), response.Body.String())
			}
			assertMediaError(t, response, "course_forbidden")
		})
	}

	for _, operation := range []struct{ path, body string }{
		{"/api/v1/admin/lessons/" + mediaLessonID + "/video", `{"declaredBytes":12,"originalName":"lesson.mp4"}`},
		{"/api/v1/admin/videos/" + mediaVideoID + "/retry", ""},
	} {
		t.Run("R04 cross-site refused before authentication "+operation.path, func(t *testing.T) {
			repository := defaultMediaRepository()
			identity := &mediaIdentity{}
			response := serveMedia(t, mediaAPIHandler(t, repository, &mediaQueue{}, identity), identity, mediaUser(mediaAdminID, identitydomain.Admin), true, "https://foreign.example", http.MethodPost, operation.path, []byte(operation.body), contentType(operation.body), true)
			if response.Code != http.StatusForbidden || identity.authenticateCalls != 0 || repository.calls() != 0 {
				t.Fatalf("status=%d auth=%d calls=%d", response.Code, identity.authenticateCalls, repository.calls())
			}
			assertMediaError(t, response, "origin_forbidden")
		})
	}
}

func TestMediaBusinessErrorsAndSafeMessages(t *testing.T) {
	for _, test := range []struct {
		name, path, code string
		failure          error
		status           int
	}{
		{name: "R05 unknown lesson", path: "/api/v1/admin/lessons/" + mediaLessonID + "/video", failure: mediadomain.ErrLessonNotFound, status: 404, code: "lesson_not_found"},
		{name: "R07 upload in progress", path: "/api/v1/admin/lessons/" + mediaLessonID + "/video", failure: mediadomain.ErrUploadInProgress, status: 409, code: "upload_in_progress"},
		{name: "unsupported media maps to its public code", path: "/api/v1/admin/lessons/" + mediaLessonID + "/video", failure: mediadomain.ErrUnsupportedMedia, status: 400, code: "unsupported_media"},
	} {
		t.Run(test.name, func(t *testing.T) {
			repository := defaultMediaRepository()
			repository.mint = func(context.Context, mediadomain.Actor, string, int64, string, time.Time, time.Time) (mediadomain.VideoAsset, error) {
				return mediadomain.VideoAsset{}, test.failure
			}
			identity := &mediaIdentity{}
			response := serveMedia(t, mediaAPIHandler(t, repository, &mediaQueue{}, identity), identity, mediaUser(mediaAdminID, identitydomain.Admin), true, "http://localhost:8080", http.MethodPost, test.path, []byte(`{"declaredBytes":12,"originalName":"lesson.mp4"}`), "application/json", true)
			if response.Code != test.status || repository.mintCalls != 1 {
				t.Fatalf("status=%d calls=%d body=%s", response.Code, repository.mintCalls, response.Body.String())
			}
			assertMediaError(t, response, test.code)
		})
	}

	t.Run("unknown video maps to video_not_found", func(t *testing.T) {
		repository := defaultMediaRepository()
		identity := &mediaIdentity{}
		response := serveMedia(t, mediaAPIHandler(t, repository, &mediaQueue{}, identity), identity, mediaUser(mediaAdminID, identitydomain.Admin), true, "http://localhost:8080", http.MethodPost, "/api/v1/admin/videos/"+mediaVideoID+"/retry", nil, "", true)
		if response.Code != http.StatusNotFound || repository.retryCalls != 1 {
			t.Fatalf("status=%d calls=%d body=%s", response.Code, repository.retryCalls, response.Body.String())
		}
		assertMediaError(t, response, "video_not_found")
	})

	for _, operation := range []struct {
		name, method, path, body, contentType string
	}{
		{name: "mint", method: http.MethodPost, path: "/api/v1/admin/lessons/" + mediaLessonID + "/video", body: `{"declaredBytes":12,"originalName":"lesson.mp4"}`, contentType: "application/json"},
		{name: "status read", method: http.MethodGet, path: "/api/v1/admin/courses/" + mediaCourseID + "/videos"},
	} {
		t.Run("R22 "+operation.name+" hides dependency detail", func(t *testing.T) {
			private := fmt.Errorf("/media/source/private.mp4 ffmpeg stderr: %w", mediadomain.ErrDependencyUnavailable)
			repository := defaultMediaRepository()
			repository.mint = func(context.Context, mediadomain.Actor, string, int64, string, time.Time, time.Time) (mediadomain.VideoAsset, error) {
				return mediadomain.VideoAsset{}, private
			}
			repository.list = func(context.Context, mediadomain.Actor, string) ([]mediadomain.VideoAsset, error) {
				return nil, private
			}
			identity := &mediaIdentity{}
			response := serveMedia(t, mediaAPIHandler(t, repository, &mediaQueue{}, identity), identity, mediaUser(mediaAdminID, identitydomain.Admin), true, "http://localhost:8080", operation.method, operation.path, []byte(operation.body), operation.contentType, true)
			if response.Code != http.StatusServiceUnavailable {
				t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
			}
			for _, secret := range []string{"/media", "private.mp4", "ffmpeg", "stderr"} {
				if strings.Contains(response.Body.String(), secret) {
					t.Fatalf("private detail %q escaped: %s", secret, response.Body.String())
				}
			}
		})
	}
}

func TestMediaRetryContracts(t *testing.T) {
	assets := map[string]mediadomain.VideoAsset{
		"50000000-0000-4000-8000-000000000001": retryAsset("50000000-0000-4000-8000-000000000001", mediadomain.Failed, mediaNow.Add(-time.Minute), ""),
		"50000000-0000-4000-8000-000000000002": retryAsset("50000000-0000-4000-8000-000000000002", mediadomain.Queued, mediaNow.Add(-time.Minute), ""),
		"50000000-0000-4000-8000-000000000003": retryAsset("50000000-0000-4000-8000-000000000003", mediadomain.Ready, mediaNow.Add(-time.Minute), ""),
		"50000000-0000-4000-8000-000000000004": retryAsset("50000000-0000-4000-8000-000000000004", mediadomain.Uploading, mediaNow.Add(-time.Minute), ""),
		"50000000-0000-4000-8000-000000000005": retryAsset("50000000-0000-4000-8000-000000000005", mediadomain.Processing, mediaNow.Add(-time.Minute), "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"),
		"50000000-0000-4000-8000-000000000006": retryAsset("50000000-0000-4000-8000-000000000006", mediadomain.Processing, mediaNow.Add(-41*time.Minute), "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"),
		"50000000-0000-4000-8000-000000000007": retryAsset("50000000-0000-4000-8000-000000000007", mediadomain.Failed, mediaNow.Add(-time.Minute), ""),
	}
	noSource := assets["50000000-0000-4000-8000-000000000007"]
	noSource.ReceivedBytes = 0
	assets[noSource.ID] = noSource
	repository := retryMediaRepository(t, assets)
	queue := &mediaQueue{}
	identity := &mediaIdentity{}
	handler := mediaAPIHandler(t, repository, queue, identity)

	for _, id := range []string{"50000000-0000-4000-8000-000000000001", "50000000-0000-4000-8000-000000000002"} {
		before := assets[id].Attempt
		queueCalls := len(queue.calls)
		response := serveMedia(t, handler, identity, mediaUser(mediaAdminID, identitydomain.Admin), true, "http://localhost:8080", http.MethodPost, "/api/v1/admin/videos/"+id+"/retry", nil, "", true)
		if response.Code != http.StatusOK || assets[id].State != mediadomain.Queued || assets[id].Attempt != before+1 || len(queue.calls) != queueCalls+1 || queue.calls[queueCalls] != (queuedCall{assetID: id, attempt: before + 1}) {
			t.Fatalf("R08 id=%s status=%d asset=%+v queue=%+v body=%s", id, response.Code, assets[id], queue.calls, response.Body.String())
		}
	}

	for _, id := range []string{"50000000-0000-4000-8000-000000000003", "50000000-0000-4000-8000-000000000004", "50000000-0000-4000-8000-000000000005"} {
		before := assets[id]
		queueCalls := len(queue.calls)
		response := serveMedia(t, handler, identity, mediaUser(mediaAdminID, identitydomain.Admin), true, "http://localhost:8080", http.MethodPost, "/api/v1/admin/videos/"+id+"/retry", nil, "", true)
		if response.Code != http.StatusConflict || !reflect.DeepEqual(assets[id], before) || len(queue.calls) != queueCalls {
			t.Fatalf("R09 id=%s status=%d asset=%+v queue=%+v", id, response.Code, assets[id], queue.calls)
		}
		assertMediaError(t, response, "video_not_retryable")
	}

	staleID := "50000000-0000-4000-8000-000000000006"
	staleAttempt := assets[staleID].Attempt
	queueCalls := len(queue.calls)
	response := serveMedia(t, handler, identity, mediaUser(mediaAdminID, identitydomain.Admin), true, "http://localhost:8080", http.MethodPost, "/api/v1/admin/videos/"+staleID+"/retry", nil, "", true)
	if response.Code != http.StatusOK || assets[staleID].State != mediadomain.Queued || assets[staleID].ClaimToken != "" || assets[staleID].Attempt != staleAttempt+1 || len(queue.calls) != queueCalls+1 || queue.calls[queueCalls] != (queuedCall{assetID: staleID, attempt: staleAttempt + 1}) {
		t.Fatalf("R37 status=%d asset=%+v queue=%+v", response.Code, assets[staleID], queue.calls)
	}
	if assets["50000000-0000-4000-8000-000000000005"].ClaimToken == "" {
		t.Fatal("R37 live claim was cleared")
	}

	noSourceID := "50000000-0000-4000-8000-000000000007"
	queueCalls = len(queue.calls)
	response = serveMedia(t, handler, identity, mediaUser(mediaAdminID, identitydomain.Admin), true, "http://localhost:8080", http.MethodPost, "/api/v1/admin/videos/"+noSourceID+"/retry", nil, "", true)
	if response.Code != http.StatusConflict || assets[noSourceID].State != mediadomain.Failed || len(queue.calls) != queueCalls {
		t.Fatalf("R39 status=%d asset=%+v queue=%+v", response.Code, assets[noSourceID], queue.calls)
	}
	assertMediaError(t, response, "video_not_retryable")
}

func defaultMediaRepository() *mediaRepository {
	return &mediaRepository{
		mint: func(_ context.Context, _ mediadomain.Actor, _ string, declaredBytes int64, originalName string, expiresAt, now time.Time) (mediadomain.VideoAsset, error) {
			asset := mediaAsset(mediaVideoID, mediadomain.Created, 1)
			asset.DeclaredBytes, asset.ReceivedBytes, asset.OriginalName, asset.ExpiresAt, asset.CreatedAt, asset.UpdatedAt = declaredBytes, 0, originalName, expiresAt, now, now
			return asset, nil
		},
		list: func(context.Context, mediadomain.Actor, string) ([]mediadomain.VideoAsset, error) { return nil, nil },
		retry: func(context.Context, mediadomain.Actor, string, time.Duration, time.Time) (mediadomain.VideoAsset, error) {
			return mediadomain.VideoAsset{}, mediadomain.ErrVideoNotFound
		},
	}
}

func retryMediaRepository(t *testing.T, assets map[string]mediadomain.VideoAsset) *mediaRepository {
	t.Helper()
	repository := defaultMediaRepository()
	repository.retry = func(_ context.Context, actor mediadomain.Actor, id string, claimAge time.Duration, now time.Time) (mediadomain.VideoAsset, error) {
		if err := actor.CanEditLesson(mediaInstructorID); err != nil {
			return mediadomain.VideoAsset{}, err
		}
		asset, ok := assets[id]
		if !ok {
			return mediadomain.VideoAsset{}, mediadomain.ErrVideoNotFound
		}
		if claimAge != 40*time.Minute {
			t.Fatalf("claimAge=%v", claimAge)
		}
		retryable := mediadomain.ValidateTransition(mediadomain.T9, asset.State, mediadomain.Queued) == nil &&
			asset.ReceivedBytes == asset.DeclaredBytes &&
			(asset.State != mediadomain.Processing || asset.UpdatedAt.Before(now.Add(-claimAge)))
		if !retryable {
			return mediadomain.VideoAsset{}, mediadomain.ErrVideoNotRetryable
		}
		asset.State, asset.ClaimToken, asset.FailureReason, asset.UpdatedAt = mediadomain.Queued, "", "", now
		asset.Attempt++
		assets[id] = asset
		return asset, nil
	}
	return repository
}

func mediaAsset(id string, state mediadomain.State, generation mediadomain.Generation) mediadomain.VideoAsset {
	return mediadomain.VideoAsset{
		ID: id, LessonID: mediaLessonID, Generation: generation, State: state,
		DeclaredBytes: 12, ReceivedBytes: 12, OriginalName: "lesson.mp4",
		ExpiresAt: mediaNow.Add(24 * time.Hour), CreatedAt: mediaNow.Add(-time.Hour), UpdatedAt: mediaNow,
	}
}

func retryAsset(id string, state mediadomain.State, updatedAt time.Time, token mediadomain.ClaimToken) mediadomain.VideoAsset {
	asset := mediaAsset(id, state, 1)
	asset.UpdatedAt, asset.ClaimToken = updatedAt, token
	if state == mediadomain.Failed {
		asset.FailureReason = mediadomain.RenderFailed
	}
	return asset
}

func mediaAPIHandler(t *testing.T, repository *mediaRepository, queue *mediaQueue, identity *mediaIdentity) http.Handler {
	t.Helper()
	config, err := identityhttp.NewConfig("test", "http://localhost:8080", false)
	if err != nil {
		t.Fatal(err)
	}
	identityHandler := identityhttp.NewHandler(identity, nil, nil, config)
	tickets := mediaapplication.NewTickets(repository, queue, func() time.Time { return mediaNow })
	audio := mediaapplication.NewAudioTracks(repository, &mediaMasterWriter{}, func() time.Time { return mediaNow })
	mediaHandler := mediahttp.NewHandler(tickets, audio)
	return httpserver.NewHandler(httpserver.Options{Service: "api", Register: func(mux *http.ServeMux) {
		mediaHandler.RegisterAdmin(mux, identityHandler.Authenticate, identityHandler.RequireSameOrigin)
	}})
}

func mediaUser(id string, role identitydomain.Role) identitydomain.User {
	return identitydomain.User{ID: id, Email: "user@courseforge.local", DisplayName: "Media user", Role: role, CreatedAt: mediaNow, UpdatedAt: mediaNow}
}

func contentType(body string) string {
	if body == "" {
		return ""
	}
	return "application/json"
}

func serveMedia(t *testing.T, handler http.Handler, identity *mediaIdentity, user identitydomain.User, authenticated bool, origin, method, path string, body []byte, mediaType string, validRequest bool) *httptest.ResponseRecorder {
	t.Helper()
	identity.user = user
	request := httptest.NewRequest(method, path, bytes.NewReader(body))
	if mediaType != "" {
		request.Header.Set("Content-Type", mediaType)
	}
	if origin != "" {
		request.Header.Set("Origin", origin)
	}
	if authenticated {
		request.AddCookie(&http.Cookie{Name: identityhttp.SessionCookieName, Value: strings.Repeat("A", 43)})
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	request.Body = io.NopCloser(bytes.NewReader(body))
	request.ContentLength = int64(len(body))
	assertMediaContract(t, request, response, validRequest)
	if response.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("response is cacheable: %v", response.Header())
	}
	return response
}

func assertMediaContract(t *testing.T, request *http.Request, response *httptest.ResponseRecorder, validRequest bool) {
	t.Helper()
	document, err := openapi3.NewLoader().LoadFromFile(filepath.Join("..", "..", "..", "..", "openapi", "openapi.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if err := document.Validate(context.Background()); err != nil {
		t.Fatal(err)
	}
	router, err := legacy.NewRouter(document)
	if err != nil {
		t.Fatal(err)
	}
	route, parameters, err := router.FindRoute(request)
	if err != nil {
		t.Fatal(err)
	}
	input := &openapi3filter.RequestValidationInput{Request: request, PathParams: parameters, Route: route}
	if validRequest {
		if err := openapi3filter.ValidateRequest(context.Background(), input); err != nil {
			t.Fatal(err)
		}
	}
	if err := openapi3filter.ValidateResponse(context.Background(), &openapi3filter.ResponseValidationInput{
		RequestValidationInput: input,
		Status:                 response.Code,
		Header:                 response.Header(),
		Body:                   io.NopCloser(bytes.NewReader(response.Body.Bytes())),
	}); err != nil {
		t.Fatal(err)
	}
}

func decodeMediaJSON(t *testing.T, response *httptest.ResponseRecorder, target any) {
	t.Helper()
	if err := json.Unmarshal(response.Body.Bytes(), target); err != nil {
		t.Fatalf("decode response: %v body=%s", err, response.Body.String())
	}
}

func assertMediaError(t *testing.T, response *httptest.ResponseRecorder, code string) {
	t.Helper()
	var envelope httpserver.ErrorResponse
	decodeMediaJSON(t, response, &envelope)
	if envelope.Error.Code != code || envelope.Error.RequestID == "" || envelope.Error.RequestID != response.Header().Get("X-Request-ID") {
		t.Fatalf("error=%+v", envelope)
	}
}

func audioTrackRepository() *mediaRepository {
	noRewrite := func([]mediadomain.AudioTrack) error { return nil }
	repo := defaultMediaRepository()
	repo.listTracks = func(_ context.Context, actor mediadomain.Actor, assetID string) ([]mediadomain.AudioTrack, error) {
		if err := actor.CanEditLesson(mediaInstructorID); err != nil {
			return nil, err
		}
		return []mediadomain.AudioTrack{
			{ID: mediaTrackID, VideoAssetID: assetID, Kind: mediadomain.EmbeddedTrack, Language: "en", Label: "Original", IsDefault: true, State: mediadomain.Ready, UpdatedAt: mediaNow},
		}, nil
	}
	repo.setDefault = func(_ context.Context, actor mediadomain.Actor, assetID, trackID string, _ time.Time, rewrite func([]mediadomain.AudioTrack) error) error {
		if err := actor.CanEditLesson(mediaInstructorID); err != nil {
			return err
		}
		_ = rewrite(nil)
		return nil
	}
	repo.updateMetadata = func(_ context.Context, actor mediadomain.Actor, assetID, trackID, language, label string, _ time.Time, rewrite func([]mediadomain.AudioTrack) error) (mediadomain.AudioTrack, error) {
		if err := actor.CanEditLesson(mediaInstructorID); err != nil {
			return mediadomain.AudioTrack{}, err
		}
		_ = rewrite(nil)
		return mediadomain.AudioTrack{ID: trackID, VideoAssetID: assetID, Kind: mediadomain.UploadedTrack, Language: language, Label: label, State: mediadomain.Ready, UpdatedAt: mediaNow}, nil
	}
	repo.deleteTrack = func(_ context.Context, actor mediadomain.Actor, assetID, trackID string, _ time.Time, rewrite func([]mediadomain.AudioTrack) error) error {
		if err := actor.CanEditLesson(mediaInstructorID); err != nil {
			return err
		}
		_ = rewrite(nil)
		return nil
	}
	_ = noRewrite
	return repo
}

func TestAudioTrackContracts(t *testing.T) {
	t.Run("AT01 admin lists audio tracks", func(t *testing.T) {
		repository := audioTrackRepository()
		identity := &mediaIdentity{}
		handler := mediaAPIHandler(t, repository, &mediaQueue{}, identity)
		response := serveMedia(t, handler, identity, mediaUser(mediaAdminID, identitydomain.Admin), true, "", http.MethodGet, "/api/v1/admin/videos/"+mediaVideoID+"/audio-tracks", nil, "", true)
		if response.Code != http.StatusOK {
			t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
		}
		var page struct {
			Items []map[string]any `json:"items"`
		}
		decodeMediaJSON(t, response, &page)
		if len(page.Items) != 1 {
			t.Fatalf("items=%d", len(page.Items))
		}
		item := page.Items[0]
		if item["trackId"] != mediaTrackID || item["kind"] != "embedded" || item["isDefault"] != true {
			t.Fatalf("item=%+v", item)
		}
	})

	t.Run("AT02 admin renames an audio track", func(t *testing.T) {
		repository := audioTrackRepository()
		identity := &mediaIdentity{}
		handler := mediaAPIHandler(t, repository, &mediaQueue{}, identity)
		response := serveMedia(t, handler, identity, mediaUser(mediaAdminID, identitydomain.Admin), true, "http://localhost:8080", http.MethodPatch, "/api/v1/admin/videos/"+mediaVideoID+"/audio-tracks/"+mediaTrackID, []byte(`{"language":"uk","label":"Українська"}`), "application/json", true)
		if response.Code != http.StatusOK {
			t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
		}
		var track map[string]any
		decodeMediaJSON(t, response, &track)
		if track["language"] != "uk" || track["label"] != "Українська" {
			t.Fatalf("track=%+v", track)
		}
	})

	t.Run("AT03 admin sets default audio track", func(t *testing.T) {
		repository := audioTrackRepository()
		identity := &mediaIdentity{}
		handler := mediaAPIHandler(t, repository, &mediaQueue{}, identity)
		response := serveMedia(t, handler, identity, mediaUser(mediaAdminID, identitydomain.Admin), true, "http://localhost:8080", http.MethodPost, "/api/v1/admin/videos/"+mediaVideoID+"/audio-tracks/"+mediaTrackID+"/default", nil, "", true)
		if response.Code != http.StatusNoContent {
			t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
		}
	})

	t.Run("AT04 admin deletes an audio track", func(t *testing.T) {
		repository := audioTrackRepository()
		identity := &mediaIdentity{}
		handler := mediaAPIHandler(t, repository, &mediaQueue{}, identity)
		response := serveMedia(t, handler, identity, mediaUser(mediaAdminID, identitydomain.Admin), true, "http://localhost:8080", http.MethodDelete, "/api/v1/admin/videos/"+mediaVideoID+"/audio-tracks/"+mediaTrackID, nil, "", true)
		if response.Code != http.StatusNoContent {
			t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
		}
	})

	t.Run("AT05 non-owner refused", func(t *testing.T) {
		repository := audioTrackRepository()
		identity := &mediaIdentity{}
		handler := mediaAPIHandler(t, repository, &mediaQueue{}, identity)
		for _, op := range []struct{ method, path, body string }{
			{http.MethodGet, "/api/v1/admin/videos/" + mediaVideoID + "/audio-tracks", ""},
			{http.MethodPatch, "/api/v1/admin/videos/" + mediaVideoID + "/audio-tracks/" + mediaTrackID, `{"language":"uk","label":"Ukr"}`},
			{http.MethodPost, "/api/v1/admin/videos/" + mediaVideoID + "/audio-tracks/" + mediaTrackID + "/default", ""},
			{http.MethodDelete, "/api/v1/admin/videos/" + mediaVideoID + "/audio-tracks/" + mediaTrackID, ""},
		} {
			response := serveMedia(t, handler, identity, mediaUser(mediaOtherOwnerID, identitydomain.Instructor), true, "http://localhost:8080", op.method, op.path, []byte(op.body), contentType(op.body), true)
			if response.Code != http.StatusForbidden {
				t.Fatalf("%s %s status=%d body=%s", op.method, op.path, response.Code, response.Body.String())
			}
			assertMediaError(t, response, "course_forbidden")
		}
	})

	t.Run("AT06 rename strict body", func(t *testing.T) {
		repository := audioTrackRepository()
		identity := &mediaIdentity{}
		handler := mediaAPIHandler(t, repository, &mediaQueue{}, identity)
		for _, test := range []struct{ name, body string }{
			{"missing label", `{"language":"uk"}`},
			{"unknown key", `{"language":"uk","label":"Ukr","extra":1}`},
			{"duplicate key", `{"language":"uk","language":"de","label":"Ukr"}`},
		} {
			t.Run(test.name, func(t *testing.T) {
				response := serveMedia(t, handler, identity, mediaUser(mediaAdminID, identitydomain.Admin), true, "http://localhost:8080", http.MethodPatch, "/api/v1/admin/videos/"+mediaVideoID+"/audio-tracks/"+mediaTrackID, []byte(test.body), "application/json", false)
				if response.Code != http.StatusBadRequest {
					t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
				}
				assertMediaError(t, response, "invalid_request")
			})
		}
	})

	t.Run("AT07 duplicate language returns 409", func(t *testing.T) {
		repository := audioTrackRepository()
		repository.updateMetadata = func(_ context.Context, _ mediadomain.Actor, _, _, _, _ string, _ time.Time, _ func([]mediadomain.AudioTrack) error) (mediadomain.AudioTrack, error) {
			return mediadomain.AudioTrack{}, mediadomain.ErrDuplicateLanguage
		}
		identity := &mediaIdentity{}
		handler := mediaAPIHandler(t, repository, &mediaQueue{}, identity)
		response := serveMedia(t, handler, identity, mediaUser(mediaAdminID, identitydomain.Admin), true, "http://localhost:8080", http.MethodPatch, "/api/v1/admin/videos/"+mediaVideoID+"/audio-tracks/"+mediaTrackID, []byte(`{"language":"en","label":"Clash"}`), "application/json", true)
		if response.Code != http.StatusConflict {
			t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
		}
		assertMediaError(t, response, "duplicate_language")
	})

	t.Run("AT08 malformed IDs refused", func(t *testing.T) {
		repository := audioTrackRepository()
		identity := &mediaIdentity{}
		handler := mediaAPIHandler(t, repository, &mediaQueue{}, identity)
		response := serveMedia(t, handler, identity, mediaUser(mediaAdminID, identitydomain.Admin), true, "", http.MethodGet, "/api/v1/admin/videos/not-a-uuid/audio-tracks", nil, "", false)
		if response.Code != http.StatusBadRequest {
			t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
		}
	})
}
