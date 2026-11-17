package http_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	identityhttp "courses/internal/identity/adapters/http"
	identityapplication "courses/internal/identity/application"
	identitydomain "courses/internal/identity/domain"
	transcripthttp "courses/internal/transcript/adapters/http"
	"courses/internal/transcript/application"
	"courses/internal/transcript/domain"
	"courses/internal/transcript/ports"
	"courses/platform/httpserver"
)

// Only the session lookup is a fixture; route registration and both middleware are real.
type routeIdentity struct{ calls int }

func (*routeIdentity) Login(context.Context, identityapplication.LoginInput) (identityapplication.LoginResult, error) {
	panic("unexpected login")
}
func (*routeIdentity) Logout(context.Context, string) error { panic("unexpected logout") }
func (s *routeIdentity) Authenticate(_ context.Context, token string) (identitydomain.User, error) {
	s.calls++
	if token != "owner-session" {
		return identitydomain.User{}, identitydomain.ErrAuthenticationRequired
	}
	return identitydomain.User{ID: ownerID, Role: identitydomain.Instructor, EmailVerified: true}, nil
}

func TestTranscriptRoutesUseIdentityAuthentication(t *testing.T) {
	for _, route := range []struct {
		name, method, suffix string
		status               int
	}{
		{"draft inventory", "GET", "/drafts", 200},
		{"regenerate", "POST", "/drafts/" + draftTrackID + "/regenerate", 202},
		{"transcript inventory", "GET", "/transcripts", 200},
	} {
		t.Run(route.name, func(t *testing.T) {
			for _, session := range []struct {
				name, token string
				calls       int
			}{
				{"valid cookie", "owner-session", 1}, {"missing cookie", "", 0}, {"invalid cookie", "invalid-session", 1},
			} {
				t.Run(session.name, func(t *testing.T) {
					lookup := &routeIdentity{}
					config, err := identityhttp.NewConfig("test", "http://localhost:8080", false)
					if err != nil {
						t.Fatal(err)
					}
					identity := identityhttp.NewHandler(lookup, nil, nil, config)
					duration := 60.0
					drafts := &stubDraftRepository{lessonAsset: &ports.PendingRecognition{LessonID: lessonID, AssetID: draftAssetID, Generation: 2, TrackID: draftTrackID, DurationSeconds: &duration}}
					queue := &draftQueue{}
					drafting := transcripthttp.NewDraftHandler(application.NewDrafting(drafts, &stubCourseReader{course: domain.Course{OwnerID: ownerID}}, queue, func() time.Time { return time.Date(2026, 11, 17, 12, 0, 0, 0, time.UTC) }, nil))
					repo := &stubRepository{result: ports.LessonTranscript{Course: domain.Course{OwnerID: ownerID}}}
					authoring := transcripthttp.NewAdminHandler(application.NewAuthoring(repo), nil)
					server := httpserver.NewHandler(httpserver.Options{Service: "route-auth-test", Register: func(mux *http.ServeMux) {
						drafting.RegisterAdmin(mux, identity.Authenticate, identity.RequireSameOrigin)
						authoring.RegisterAdmin(mux, identity.Authenticate, identity.RequireSameOrigin)
					}})
					req := httptest.NewRequest(route.method, "/api/v1/admin/lessons/"+lessonID+route.suffix, nil)
					req.Header.Set("Origin", "http://localhost:8080")
					if session.token != "" {
						req.AddCookie(&http.Cookie{Name: identityhttp.SessionCookieName, Value: session.token})
					}
					if _, ok := httpserver.ActorFromContext(req.Context()); ok {
						t.Fatal("request must start without Actor context")
					}
					rec := httptest.NewRecorder()
					server.ServeHTTP(rec, req)
					if lookup.calls != session.calls {
						t.Fatalf("authentication calls=%d; want %d", lookup.calls, session.calls)
					}
					wantStatus, wantJobs := 401, 0
					if session.token == "owner-session" {
						wantStatus = route.status
						if route.method == "POST" {
							wantJobs = 1
						}
					}
					if rec.Code != wantStatus {
						t.Fatalf("response=%d %s; want %d", rec.Code, rec.Body.String(), wantStatus)
					}
					if len(queue.jobs) != wantJobs {
						t.Fatalf("queued jobs=%d; want %d", len(queue.jobs), wantJobs)
					}
					if wantJobs == 1 && queue.jobs[0].TrackID != draftTrackID {
						t.Fatalf("queued track=%q; want %s", queue.jobs[0].TrackID, draftTrackID)
					}
				})
			}
		})
	}
}
