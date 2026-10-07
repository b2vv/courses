package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"courses/internal/contracts/access/accesstest"
	playbackhttp "courses/internal/playback/adapters/http"
	playbackapp "courses/internal/playback/application"
	"courses/internal/playback/domain"
	"courses/platform/httpserver"
	"courses/platform/observability"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/legacy"
	dto "github.com/prometheus/client_model/go"
)

const (
	userID     = "40000000-0000-4000-8000-000000000001"
	playbackID = "11111111111111111111111111111111"
	videoID    = "50000000-0000-4000-8000-000000000001"
	lessonID   = "70000000-0000-4000-8000-000000000001"
	r09OwnerID = "40000000-0000-4000-8000-000000000002"
)

type authorizer struct {
	calls      int
	actor      domain.Actor
	playbackID string
	videoID    string
	err        error
}

type sessions struct {
	playback       domain.LessonPlayback
	playErr        error
	heartbeatErr   error
	playCalls      int
	heartbeatCalls int
	actor          domain.Actor
	lessonID       string
	playbackID     string
	update         domain.ProgressUpdate
}

type r09PlaybackReader struct{ playback domain.LessonPlayback }

func (reader *r09PlaybackReader) LoadLessonPlayback(context.Context, string, string) (domain.LessonPlayback, error) {
	return reader.playback, nil
}

type r09OwnershipWriter struct{}

func (*r09OwnershipWriter) Take(context.Context, string, domain.Ownership, time.Duration) error {
	return nil
}

func (*r09OwnershipWriter) Renew(context.Context, string, string, time.Duration) (domain.OwnershipRenewal, error) {
	return domain.OwnershipRenewed, nil
}

type r09ProgressWriter struct{}

func (*r09ProgressWriter) Save(context.Context, string, string, domain.ProgressUpdate, time.Time) error {
	return nil
}

type r09OwnershipReader struct{}

func (*r09OwnershipReader) Load(context.Context, string) (domain.Ownership, error) {
	return domain.Ownership{PlaybackID: playbackID, VideoID: videoID, LessonID: lessonID}, nil
}

type r09VideoAccessReader struct{ course domain.VideoCourse }

func (reader *r09VideoAccessReader) LoadVideoCourse(context.Context, string, string) (domain.VideoCourse, error) {
	return reader.course, nil
}

func (service *sessions) Play(_ context.Context, actor domain.Actor, lessonID, playbackID string) (domain.LessonPlayback, error) {
	service.playCalls++
	service.actor, service.lessonID, service.playbackID = actor, lessonID, playbackID
	return service.playback, service.playErr
}

func (service *sessions) Heartbeat(_ context.Context, actor domain.Actor, playbackID, lessonID string, update domain.ProgressUpdate) error {
	service.heartbeatCalls++
	service.actor, service.playbackID, service.lessonID, service.update = actor, playbackID, lessonID, update
	return service.heartbeatErr
}

func testConfig(t *testing.T) playbackhttp.Config {
	t.Helper()
	config, err := playbackhttp.NewConfig(30, 180)
	if err != nil {
		t.Fatal(err)
	}
	return config
}

func counterValue(t *testing.T, counter interface{ Write(*dto.Metric) error }) float64 {
	t.Helper()
	metric := &dto.Metric{}
	if err := counter.Write(metric); err != nil {
		t.Fatal(err)
	}
	return metric.GetCounter().GetValue()
}

func passThrough(next http.Handler) http.Handler { return next }
func ignoreAV1Probe(string)                      {}

func (service *authorizer) Authorize(_ context.Context, actor domain.Actor, playbackID, videoID string) error {
	service.calls++
	service.actor, service.playbackID, service.videoID = actor, playbackID, videoID
	return service.err
}

func rejectedAuthentication(status int) func(http.Handler) http.Handler {
	return func(http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(status)
		})
	}
}

func authenticated(actor httpserver.Actor) func(http.Handler) http.Handler {
	if actor.UserID != "" {
		actor.EmailVerified = true
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r.WithContext(httpserver.WithActor(r.Context(), actor)))
		})
	}
}

func TestAuthorizeReturnsOnlyForbiddenOrUnavailableForFailures(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		authenticate func(http.Handler) http.Handler
		failure      error
		wantStatus   int
		wantCalls    int
		wantCode     string
	}{
		{name: "identity denial is normalized", authenticate: rejectedAuthentication(http.StatusUnauthorized), wantStatus: http.StatusForbidden},
		{name: "identity dependency stays unavailable", authenticate: rejectedAuthentication(http.StatusServiceUnavailable), wantStatus: http.StatusServiceUnavailable},
		{name: "ownership or access denial", authenticate: authenticated(httpserver.Actor{UserID: userID, Role: "student"}), failure: domain.ErrForbidden, wantStatus: http.StatusForbidden, wantCalls: 1, wantCode: "playback_forbidden"},
		{name: "purchase required", authenticate: authenticated(httpserver.Actor{UserID: userID, Role: "student"}), failure: domain.ErrPurchaseRequired, wantStatus: http.StatusForbidden, wantCalls: 1, wantCode: "course_purchase_required"},
		{name: "dependency failure", authenticate: authenticated(httpserver.Actor{UserID: userID, Role: "student"}), failure: domain.ErrDependencyUnavailable, wantStatus: http.StatusServiceUnavailable, wantCalls: 1, wantCode: "dependency_unavailable"},
		{name: "unknown verification failure", authenticate: authenticated(httpserver.Actor{UserID: userID, Role: "student"}), failure: errors.New("unexpected"), wantStatus: http.StatusServiceUnavailable, wantCalls: 1, wantCode: "dependency_unavailable"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			service := &authorizer{err: test.failure}
			adapter := playbackhttp.NewHandler(service, nil, testConfig(t), ignoreAV1Probe)
			handler := httpserver.NewHandler(httpserver.Options{Service: "api", Register: func(mux *http.ServeMux) {
				adapter.Register(mux, test.authenticate, passThrough)
			}})
			request := httptest.NewRequest(http.MethodGet, "/internal/playback/authorize", nil)
			request.Header.Set("X-Playback-Id", playbackID)
			request.Header.Set("X-Video-Id", videoID)
			response := httptest.NewRecorder()

			handler.ServeHTTP(response, request)

			if response.Code != test.wantStatus || service.calls != test.wantCalls {
				t.Fatalf("status=%d calls=%d, want %d and %d", response.Code, service.calls, test.wantStatus, test.wantCalls)
			}
			if test.wantCode != "" && !strings.Contains(response.Body.String(), `"code":"`+test.wantCode+`"`) {
				t.Fatalf("body=%q, want code %q", response.Body.String(), test.wantCode)
			}
		})
	}
}

func TestAuthorizeReturnsNoContentForAnAllowedMediaRequest(t *testing.T) {
	t.Parallel()
	service := &authorizer{}
	adapter := playbackhttp.NewHandler(service, nil, testConfig(t), ignoreAV1Probe)
	handler := httpserver.NewHandler(httpserver.Options{Service: "api", Register: func(mux *http.ServeMux) {
		adapter.Register(mux, authenticated(httpserver.Actor{UserID: "40000000-0000-4000-8000-000000000001", Role: "student"}), passThrough)
	}})
	request := httptest.NewRequest(http.MethodGet, "/internal/playback/authorize", nil)
	request.Header.Set("X-Playback-Id", "11111111111111111111111111111111")
	request.Header.Set("X-Video-Id", "50000000-0000-4000-8000-000000000001")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent || response.Body.Len() != 0 {
		t.Fatalf("status=%d body=%q", response.Code, response.Body.String())
	}
	if service.actor != (domain.Actor{UserID: "40000000-0000-4000-8000-000000000001", EmailVerified: true}) || service.playbackID != "11111111111111111111111111111111" || service.videoID != "50000000-0000-4000-8000-000000000001" {
		t.Fatalf("actor=%+v playback=%q video=%q", service.actor, service.playbackID, service.videoID)
	}
}

func TestR09PlaybackSessionForeignDraftDeniedAndActiveMemberAllowed(t *testing.T) {
	for _, test := range []accesstest.Case{
		playbackAccessCase(t, accesstest.Stranger, "draft"),
		playbackAccessCase(t, accesstest.Stranger, "published"),
	} {
		t.Run(test.Name, func(t *testing.T) {
			course := domain.VideoCourse{OwnerID: r09OwnerID, Status: domain.CourseStatus(test.Status)}
			if test.Want {
				course.Membership = "active"
			}
			service := playbackapp.NewSessions(
				&r09PlaybackReader{playback: domain.LessonPlayback{
					VideoID: videoID,
					Course:  course,
				}},
				&r09OwnershipWriter{}, &r09ProgressWriter{}, time.Minute, time.Now,
			)
			adapter := playbackhttp.NewHandler(&authorizer{}, service, testConfig(t), ignoreAV1Probe)
			handler := httpserver.NewHandler(httpserver.Options{Service: "api", Register: func(mux *http.ServeMux) {
				adapter.Register(mux, authenticated(playbackAccessActor(test)), passThrough)
			}})
			request := httptest.NewRequest(http.MethodPost, "/api/v1/playback/sessions", strings.NewReader(`{"lessonId":"`+lessonID+`"}`))
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)

			if test.Want {
				if response.Code != http.StatusCreated {
					t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
				}
			} else if response.Code != http.StatusForbidden || !strings.Contains(response.Body.String(), `"code":"playback_forbidden"`) {
				t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
			}
		})
	}
}

func TestR09PlaybackAuthorizeForeignDraftDeniedAndActiveMemberAllowed(t *testing.T) {
	for _, test := range []accesstest.Case{
		playbackAccessCase(t, accesstest.Stranger, "draft"),
		playbackAccessCase(t, accesstest.Stranger, "published"),
	} {
		t.Run(test.Name, func(t *testing.T) {
			course := domain.VideoCourse{OwnerID: r09OwnerID, Status: domain.CourseStatus(test.Status)}
			if test.Want {
				course.Membership = "active"
			}
			service := playbackapp.NewAuthorizer(&r09OwnershipReader{}, &r09VideoAccessReader{course: course})
			adapter := playbackhttp.NewHandler(service, nil, testConfig(t), ignoreAV1Probe)
			handler := httpserver.NewHandler(httpserver.Options{Service: "api", Register: func(mux *http.ServeMux) {
				adapter.Register(mux, authenticated(playbackAccessActor(test)), passThrough)
			}})
			request := httptest.NewRequest(http.MethodGet, "/internal/playback/authorize", nil)
			request.Header.Set("X-Playback-Id", playbackID)
			request.Header.Set("X-Video-Id", videoID)
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)

			wantStatus := http.StatusForbidden
			if test.Want {
				wantStatus = http.StatusNoContent
			}
			if response.Code != wantStatus {
				t.Fatalf("status=%d want=%d body=%s", response.Code, wantStatus, response.Body.String())
			}
		})
	}
}

func playbackAccessCase(t *testing.T, role accesstest.Role, status string) accesstest.Case {
	t.Helper()
	for _, test := range accesstest.TruthTable {
		if test.Role == role && test.Status == status {
			return test
		}
	}
	t.Fatalf("missing access truth-table row role=%q status=%q", role, status)
	return accesstest.Case{}
}

func playbackAccessActor(test accesstest.Case) httpserver.Actor {
	actor := httpserver.Actor{UserID: userID, Role: "student"}
	switch test.Role {
	case accesstest.Owner:
		actor.UserID, actor.Role = r09OwnerID, "instructor"
	case accesstest.Administrator:
		actor.Role = "admin"
	}
	return actor
}

func TestPlayMintsServerSelectedSession(t *testing.T) {
	t.Parallel()
	service := &sessions{playback: domain.LessonPlayback{VideoID: videoID, ClientSequence: 12}}
	adapter := playbackhttp.NewHandler(&authorizer{}, service, testConfig(t), ignoreAV1Probe)
	handler := httpserver.NewHandler(httpserver.Options{Service: "api", Register: func(mux *http.ServeMux) {
		adapter.Register(mux, authenticated(httpserver.Actor{UserID: userID, Role: "student"}), passThrough)
	}})
	request := httptest.NewRequest(http.MethodPost, "/api/v1/playback/sessions", strings.NewReader(`{"lessonId":"`+lessonID+`"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusCreated || response.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("status=%d headers=%v body=%s", response.Code, response.Header(), response.Body.String())
	}
	var body struct {
		PlaybackID          string `json:"playbackId"`
		MediaPath           string `json:"mediaPath"`
		ClientSequence      int64  `json:"clientSeq"`
		HeartbeatSeconds    int    `json:"heartbeatSeconds"`
		OwnershipTTLSeconds int    `json:"ownershipTtlSeconds"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if !regexp.MustCompile(`^[0-9a-f]{32}$`).MatchString(body.PlaybackID) || service.playbackID != body.PlaybackID {
		t.Fatalf("response playback=%q service playback=%q", body.PlaybackID, service.playbackID)
	}
	wantPath := "/media/p/" + body.PlaybackID + "/hls/" + videoID + "/master.m3u8"
	if body.MediaPath != wantPath || body.ClientSequence != 12 || body.HeartbeatSeconds != 30 || body.OwnershipTTLSeconds != 180 {
		t.Fatalf("body=%+v", body)
	}
	if service.actor != (domain.Actor{UserID: userID, EmailVerified: true}) || service.lessonID != lessonID || service.playCalls != 1 {
		t.Fatalf("service=%+v", service)
	}
}

func TestP12ValidAV1ProbeIncrementsTheBoundedMetricOnce(t *testing.T) {
	t.Parallel()
	for _, result := range []string{"supported", "unsupported", "probe_failed"} {
		t.Run(result, func(t *testing.T) {
			metrics := observability.NewMetrics()
			counter := metrics.PlaybackAV1Probe.WithLabelValues(result)
			before := counterValue(t, counter)
			observed := ""
			service := &sessions{playback: domain.LessonPlayback{VideoID: videoID}}
			adapter := playbackhttp.NewHandler(&authorizer{}, service, testConfig(t), func(result string) {
				observed = result
				counter.Inc()
			})
			handler := httpserver.NewHandler(httpserver.Options{Service: "api", Register: func(mux *http.ServeMux) {
				adapter.Register(mux, authenticated(httpserver.Actor{UserID: userID, Role: "student"}), passThrough)
			}})
			request := httptest.NewRequest(http.MethodPost, "/api/v1/playback/sessions", strings.NewReader(`{"lessonId":"`+lessonID+`","av1Support":"`+result+`"}`))
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()

			handler.ServeHTTP(response, request)
			after := counterValue(t, counter)
			t.Logf("P12 measurement: av1Support=%q counter before=%g after=%g", observed, before, after)

			if response.Code != http.StatusCreated || service.playCalls != 1 || observed != result || after-before != 1 {
				t.Fatalf("status=%d calls=%d av1Support=%q counter before=%g after=%g body=%s", response.Code, service.playCalls, observed, before, after, response.Body.String())
			}
		})
	}
}

func TestP12ValidAV1ProbeIsCountedWhenMintIsForbidden(t *testing.T) {
	t.Parallel()
	metrics := observability.NewMetrics()
	counter := metrics.PlaybackAV1Probe.WithLabelValues("supported")
	before := counterValue(t, counter)
	observed := ""
	service := &sessions{playErr: domain.ErrForbidden}
	adapter := playbackhttp.NewHandler(&authorizer{}, service, testConfig(t), func(result string) {
		observed = result
		counter.Inc()
	})
	handler := httpserver.NewHandler(httpserver.Options{Service: "api", Register: func(mux *http.ServeMux) {
		adapter.Register(mux, authenticated(httpserver.Actor{UserID: userID, Role: "student"}), passThrough)
	}})
	request := httptest.NewRequest(http.MethodPost, "/api/v1/playback/sessions", strings.NewReader(`{"lessonId":"`+lessonID+`","av1Support":"supported"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)
	after := counterValue(t, counter)
	t.Logf("P12 denied-mint measurement: av1Support=%q counter before=%g after=%g", observed, before, after)

	if response.Code != http.StatusForbidden || service.playCalls != 1 || observed != "supported" || after-before != 1 {
		t.Fatalf("status=%d calls=%d av1Support=%q counter before=%g after=%g body=%s", response.Code, service.playCalls, observed, before, after, response.Body.String())
	}
}

func TestPlayRejectsAnythingButOneValidLessonID(t *testing.T) {
	t.Parallel()
	for _, body := range []string{
		`{}`,
		`{"lessonId":"not-a-uuid"}`,
		`{"lessonId":"` + lessonID + `","av1Support":""}`,
		`{"lessonId":"` + lessonID + `","av1Support":"unknown"}`,
		`{"lessonId":"` + lessonID + `","av1Support":null}`,
		`{"lessonId":"` + lessonID + `","av1Support":1}`,
		`{"lessonId":"` + lessonID + `","videoId":"` + videoID + `"}`,
		`{"lessonId":"` + lessonID + `","lessonId":"` + lessonID + `"}`,
		`{"lessonId":"` + lessonID + `"} {}`,
	} {
		service := &sessions{}
		adapter := playbackhttp.NewHandler(&authorizer{}, service, testConfig(t), ignoreAV1Probe)
		handler := httpserver.NewHandler(httpserver.Options{Service: "api", Register: func(mux *http.ServeMux) {
			adapter.Register(mux, authenticated(httpserver.Actor{UserID: userID, Role: "student"}), passThrough)
		}})
		request := httptest.NewRequest(http.MethodPost, "/api/v1/playback/sessions", strings.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != http.StatusBadRequest || service.playCalls != 0 || !strings.Contains(response.Body.String(), `"code":"invalid_request"`) {
			t.Fatalf("input=%s status=%d calls=%d body=%s", body, response.Code, service.playCalls, response.Body.String())
		}
	}
}

func TestPlayMapsPolicyFailuresWithoutConflict(t *testing.T) {
	t.Parallel()
	tests := []struct {
		err    error
		status int
		code   string
	}{
		{err: domain.ErrForbidden, status: http.StatusForbidden, code: "playback_forbidden"},
		{err: domain.ErrPurchaseRequired, status: http.StatusForbidden, code: "course_purchase_required"},
		{err: domain.ErrLessonNotFound, status: http.StatusNotFound, code: "lesson_not_found"},
		{err: domain.ErrVideoNotReady, status: http.StatusNotFound, code: "video_not_ready"},
		{err: domain.ErrDependencyUnavailable, status: http.StatusServiceUnavailable, code: "dependency_unavailable"},
		{err: domain.ErrPlaybackTakenOver, status: http.StatusInternalServerError, code: "internal_error"},
	}
	for _, test := range tests {
		service := &sessions{playErr: test.err}
		adapter := playbackhttp.NewHandler(&authorizer{}, service, testConfig(t), ignoreAV1Probe)
		handler := httpserver.NewHandler(httpserver.Options{Service: "api", Register: func(mux *http.ServeMux) {
			adapter.Register(mux, authenticated(httpserver.Actor{UserID: userID, Role: "student"}), passThrough)
		}})
		request := httptest.NewRequest(http.MethodPost, "/api/v1/playback/sessions", strings.NewReader(`{"lessonId":"`+lessonID+`"}`))
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code == http.StatusConflict || response.Code != test.status || !strings.Contains(response.Body.String(), `"code":"`+test.code+`"`) {
			t.Fatalf("err=%v status=%d body=%s", test.err, response.Code, response.Body.String())
		}
	}
}

func TestHeartbeatMapsDistinctOwnershipOutcomes(t *testing.T) {
	t.Parallel()
	tests := []struct {
		err    error
		status int
		code   string
	}{
		{status: http.StatusNoContent},
		{err: domain.ErrPlaybackTakenOver, status: http.StatusConflict, code: "playback_taken_over"},
		{err: domain.ErrPlaybackSessionExpired, status: http.StatusGone, code: "playback_session_expired"},
		{err: domain.ErrLessonNotFound, status: http.StatusNotFound, code: "lesson_not_found"},
		{err: domain.ErrDependencyUnavailable, status: http.StatusServiceUnavailable, code: "dependency_unavailable"},
	}
	for _, test := range tests {
		service := &sessions{heartbeatErr: test.err}
		adapter := playbackhttp.NewHandler(&authorizer{}, service, testConfig(t), ignoreAV1Probe)
		handler := httpserver.NewHandler(httpserver.Options{Service: "api", Register: func(mux *http.ServeMux) {
			adapter.Register(mux, authenticated(httpserver.Actor{UserID: userID, Role: "student"}), passThrough)
		}})
		body := `{"playbackId":"` + playbackID + `","lessonId":"` + lessonID + `","positionSeconds":120,"clientSeq":7,"reason":"pause"}`
		request := httptest.NewRequest(http.MethodPost, "/api/v1/playback/heartbeat", strings.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != test.status || service.heartbeatCalls != 1 {
			t.Fatalf("err=%v status=%d calls=%d body=%s", test.err, response.Code, service.heartbeatCalls, response.Body.String())
		}
		if test.code != "" && !strings.Contains(response.Body.String(), `"code":"`+test.code+`"`) {
			t.Fatalf("err=%v body=%s", test.err, response.Body.String())
		}
		if test.err == nil && (response.Body.Len() != 0 || service.update.PositionSeconds() != 120 || service.update.ClientSequence() != 7 || service.update.Reason() != domain.ProgressPause) {
			t.Fatalf("body=%q update=%+v", response.Body.String(), service.update)
		}
	}
}

func TestHeartbeatRejectsInvalidBodiesBeforeApplication(t *testing.T) {
	t.Parallel()
	validPrefix := `{"playbackId":"` + playbackID + `","lessonId":"` + lessonID + `",`
	for _, body := range []string{
		`{"playbackId":"` + playbackID + `","lessonId":"` + lessonID + `","positionSeconds":1,"clientSeq":7}`,
		validPrefix + `"positionSeconds":1,"clientSeq":7,"reason":"tick","reason":"tick"}`,
		validPrefix + `"positionSeconds":-1,"clientSeq":7,"reason":"tick"}`,
		validPrefix + `"positionSeconds":"NaN","clientSeq":7,"reason":"tick"}`,
		validPrefix + `"positionSeconds":1,"clientSeq":2147483648,"reason":"tick"}`,
		validPrefix + `"positionSeconds":1,"clientSeq":7,"reason":"other"}`,
		`{"playbackId":"bad","lessonId":"` + lessonID + `","positionSeconds":1,"clientSeq":7,"reason":"tick"}`,
		validPrefix + `"positionSeconds":1,"clientSeq":7,"reason":"tick","extra":true}`,
	} {
		service := &sessions{}
		adapter := playbackhttp.NewHandler(&authorizer{}, service, testConfig(t), ignoreAV1Probe)
		handler := httpserver.NewHandler(httpserver.Options{Service: "api", Register: func(mux *http.ServeMux) {
			adapter.Register(mux, authenticated(httpserver.Actor{UserID: userID, Role: "student"}), passThrough)
		}})
		request := httptest.NewRequest(http.MethodPost, "/api/v1/playback/heartbeat", bytes.NewBufferString(body))
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != http.StatusBadRequest || service.heartbeatCalls != 0 {
			t.Fatalf("input=%s status=%d calls=%d body=%s", body, response.Code, service.heartbeatCalls, response.Body.String())
		}
	}
}

func TestPublicPlaybackRejectsQueryParameters(t *testing.T) {
	t.Parallel()
	tests := []struct {
		path string
		body string
	}{
		{path: "/api/v1/playback/sessions?videoId=" + videoID, body: `{"lessonId":"` + lessonID + `"}`},
		{path: "/api/v1/playback/heartbeat?retry=true", body: `{"playbackId":"` + playbackID + `","lessonId":"` + lessonID + `","positionSeconds":1,"clientSeq":0,"reason":"tick"}`},
	}
	for _, test := range tests {
		service := &sessions{}
		adapter := playbackhttp.NewHandler(&authorizer{}, service, testConfig(t), ignoreAV1Probe)
		handler := httpserver.NewHandler(httpserver.Options{Service: "api", Register: func(mux *http.ServeMux) {
			adapter.Register(mux, authenticated(httpserver.Actor{UserID: userID, Role: "student"}), passThrough)
		}})
		request := httptest.NewRequest(http.MethodPost, test.path, strings.NewReader(test.body))
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != http.StatusBadRequest || service.playCalls != 0 || service.heartbeatCalls != 0 {
			t.Fatalf("path=%s status=%d play=%d heartbeat=%d", test.path, response.Code, service.playCalls, service.heartbeatCalls)
		}
	}
}

func TestPublicPlaybackHTTPMatchesOpenAPI(t *testing.T) {
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
	tests := []struct {
		name       string
		path       string
		body       string
		service    *sessions
		wantStatus int
	}{
		{name: "play", path: "/api/v1/playback/sessions", body: `{"lessonId":"` + lessonID + `"}`, service: &sessions{playback: domain.LessonPlayback{VideoID: videoID}}, wantStatus: http.StatusCreated},
		{name: "P12 play with AV1 probe", path: "/api/v1/playback/sessions", body: `{"lessonId":"` + lessonID + `","av1Support":"unsupported"}`, service: &sessions{playback: domain.LessonPlayback{VideoID: videoID}}, wantStatus: http.StatusCreated},
		{name: "Pay19 purchase required", path: "/api/v1/playback/sessions", body: `{"lessonId":"` + lessonID + `"}`, service: &sessions{playErr: domain.ErrPurchaseRequired}, wantStatus: http.StatusForbidden},
		{name: "heartbeat", path: "/api/v1/playback/heartbeat", body: `{"playbackId":"` + playbackID + `","lessonId":"` + lessonID + `","positionSeconds":1,"clientSeq":0,"reason":"tick"}`, service: &sessions{}, wantStatus: http.StatusNoContent},
		{name: "R17 heartbeat forbidden", path: "/api/v1/playback/heartbeat", body: `{"playbackId":"` + playbackID + `","lessonId":"` + lessonID + `","positionSeconds":1,"clientSeq":0,"reason":"tick"}`, service: &sessions{heartbeatErr: domain.ErrForbidden}, wantStatus: http.StatusForbidden},
		{name: "heartbeat taken over", path: "/api/v1/playback/heartbeat", body: `{"playbackId":"` + playbackID + `","lessonId":"` + lessonID + `","positionSeconds":1,"clientSeq":0,"reason":"tick"}`, service: &sessions{heartbeatErr: domain.ErrPlaybackTakenOver}, wantStatus: http.StatusConflict},
		{name: "heartbeat expired", path: "/api/v1/playback/heartbeat", body: `{"playbackId":"` + playbackID + `","lessonId":"` + lessonID + `","positionSeconds":1,"clientSeq":0,"reason":"tick"}`, service: &sessions{heartbeatErr: domain.ErrPlaybackSessionExpired}, wantStatus: http.StatusGone},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, test.path, strings.NewReader(test.body))
			request.Header.Set("Content-Type", "application/json")
			request.Header.Set("Origin", "http://localhost:8080")
			route, parameters, err := router.FindRoute(request)
			if err != nil {
				t.Fatal(err)
			}
			input := &openapi3filter.RequestValidationInput{Request: request, PathParams: parameters, Route: route}
			if err := openapi3filter.ValidateRequest(context.Background(), input); err != nil {
				t.Fatal(err)
			}
			request.Body = io.NopCloser(strings.NewReader(test.body))
			adapter := playbackhttp.NewHandler(&authorizer{}, test.service, testConfig(t), ignoreAV1Probe)
			handler := httpserver.NewHandler(httpserver.Options{Service: "api", Register: func(mux *http.ServeMux) {
				adapter.Register(mux, authenticated(httpserver.Actor{UserID: userID, Role: "student"}), passThrough)
			}})
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			if response.Code != test.wantStatus {
				t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
			}
			if err := openapi3filter.ValidateResponse(context.Background(), &openapi3filter.ResponseValidationInput{
				RequestValidationInput: input,
				Status:                 response.Code,
				Header:                 response.Header(),
				Body:                   io.NopCloser(bytes.NewReader(response.Body.Bytes())),
			}); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestPublicPlaybackMutationsRequireSameOriginBeforeAuthentication(t *testing.T) {
	t.Parallel()
	for _, path := range []string{"/api/v1/playback/sessions", "/api/v1/playback/heartbeat"} {
		var order []string
		authenticate := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, "authenticate")
				next.ServeHTTP(w, r)
			})
		}
		requireSameOrigin := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, "same-origin")
				next.ServeHTTP(w, r)
			})
		}
		adapter := playbackhttp.NewHandler(&authorizer{}, &sessions{}, testConfig(t), ignoreAV1Probe)
		handler := httpserver.NewHandler(httpserver.Options{Service: "api", Register: func(mux *http.ServeMux) {
			adapter.Register(mux, authenticate, requireSameOrigin)
		}})
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest(http.MethodPost, path, nil))
		if strings.Join(order, ",") != "same-origin,authenticate" {
			t.Fatalf("path=%s order=%v", path, order)
		}
	}
}
