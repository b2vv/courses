package http_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	transcripthttp "courses/internal/transcript/adapters/http"
	"courses/internal/transcript/application"
	"courses/internal/transcript/domain"
	"courses/internal/transcript/ports"
	"courses/platform/httpserver"
)

func TestM2R10TranscriptInventoryHTTP(t *testing.T) {
	for _, tc := range []struct {
		name, user    string
		missing       bool
		count, status int
		code          string
	}{
		{"empty", ownerID, false, 0, 200, ""}, {"one", ownerID, false, 1, 200, ""}, {"two", ownerID, false, 2, 200, ""},
		{"foreign", viewerID, false, 2, 403, "course_forbidden"}, {"anonymous", "", false, 2, 401, "authentication_required"},
		{"missing", ownerID, true, 0, 404, "lesson_not_found"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := &stubRepository{result: ports.LessonTranscript{Course: domain.Course{OwnerID: ownerID}}}
			if tc.count > 0 {
				repo.inventory = append(repo.inventory, ports.TranscriptInventory{Language: "en", UpdatedAt: "2026-11-15T10:00:00Z", Stale: true})
			}
			if tc.count > 1 {
				repo.inventory = append(repo.inventory, ports.TranscriptInventory{Language: "uk", UpdatedAt: "2026-11-15T10:00:00Z", IsPrimary: true})
			}
			if tc.missing {
				repo.err = domain.ErrNotFound
			}
			handler := transcripthttp.NewAdminHandler(application.NewAuthoring(repo), nil)
			server := httpserver.NewHandler(httpserver.Options{Service: "inventory-test", Register: func(mux *http.ServeMux) {
				handler.RegisterAdmin(mux, func(next http.Handler) http.Handler { return next }, func(next http.Handler) http.Handler { return next })
			}})
			r := request(t, "/api/v1/admin/lessons/"+lessonID+"/transcripts", httpserver.Actor{UserID: tc.user, Role: "instructor"})
			rec := httptest.NewRecorder()
			server.ServeHTTP(rec, r)
			if tc.code != "" {
				assertM2R10Error(t, rec, tc.status, tc.code)
			} else {
				if rec.Code != 200 || rec.Header().Get("Cache-Control") != "no-store" {
					t.Fatalf("response=%d %s", rec.Code, rec.Body.String())
				}
				if tc.count == 0 && rec.Body.String() != "{\"transcripts\":[]}\n" {
					t.Fatalf("empty=%s", rec.Body.String())
				}
				if tc.count == 2 && rec.Body.String() != "{\"transcripts\":[{\"language\":\"en\",\"updatedAt\":\"2026-11-15T10:00:00Z\",\"stale\":true,\"isPrimary\":false},{\"language\":\"uk\",\"updatedAt\":\"2026-11-15T10:00:00Z\",\"stale\":false,\"isPrimary\":true}]}\n" {
					t.Fatalf("inventory=%s", rec.Body.String())
				}
			}
			validateMS2Response(t, r, rec)
		})
	}
}
