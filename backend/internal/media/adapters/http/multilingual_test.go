package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http/httptest"
	"testing"
	"time"

	identitydomain "courses/internal/identity/domain"
	"courses/internal/media/domain"
	"courses/internal/media/ports"
	"courses/internal/openapispec"
	"courses/internal/platform/openapigen"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/legacy"
)

type countingMaster struct{ calls int }

func (m *countingMaster) RegenerateMaster(context.Context, string, []ports.AudioRendition) error {
	m.calls++
	return nil
}
func validateM2R12(t *testing.T, method, path, body string, rec *httptest.ResponseRecorder) {
	t.Helper()
	doc, err := openapigen.GenerateFromSource(openapispec.AllOpenAPIContributions(), "../../../../openapi/openapi.overlay.yaml", "../../../..")
	if err != nil {
		t.Fatal(err)
	}
	router, err := legacy.NewRouter(doc)
	if err != nil {
		t.Fatal(err)
	}
	r := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	if body != "" {
		r.Header.Set("Content-Type", "application/json")
	}
	route, params, err := router.FindRoute(r)
	if err != nil {
		t.Fatal(err)
	}
	input := &openapi3filter.RequestValidationInput{Request: r, PathParams: params, Route: route, Options: &openapi3filter.Options{AuthenticationFunc: func(context.Context, *openapi3filter.AuthenticationInput) error { return nil }}}
	if rec.Code < 400 {
		if err := openapi3filter.ValidateRequest(context.Background(), input); err != nil {
			t.Fatal(err)
		}
	}
	if err := openapi3filter.ValidateResponse(context.Background(), &openapi3filter.ResponseValidationInput{RequestValidationInput: input, Status: rec.Code, Header: rec.Header(), Body: io.NopCloser(bytes.NewReader(rec.Body.Bytes()))}); err != nil {
		t.Fatal(err)
	}
}
func TestM2R12MintFlagAndDefaultHTTP(t *testing.T) {
	path := "/api/v1/admin/videos/" + mediaVideoID + "/audio-tracks"
	for _, def := range []bool{false, true} {
		for _, field := range []string{"", `,"multilingual":true`, `,"multilingual":false`} {
			t.Run(fmt.Sprintf("default=%v/field=%s", def, field), func(t *testing.T) {
				repo := audioTrackRepository()
				repo.mintAudioTrack = func(context.Context, domain.Actor, string, int64, string, string, string, time.Time, time.Time) (domain.AudioTrack, error) {
					return domain.AudioTrack{ID: mediaTrackID, ExpiresAt: mediaNow.Add(time.Hour)}, nil
				}
				identity := &mediaIdentity{}
				server := mediaAPIWithAudio(t, repo, &mediaQueue{}, identity, def, &countingMaster{})
				body := `{"declaredBytes":12,"originalName":"dub.aac","language":"uk","label":"Dub"` + field + `}`
				rec := serveMedia(t, server, identity, mediaUser(mediaAdminID, identitydomain.Admin), true, "http://localhost:8080", "POST", path, []byte(body), "application/json", true)
				want := def
				if field != "" {
					want = field == `,"multilingual":true`
				}
				if rec.Code != 201 || repo.mintedMultilingual != want {
					t.Fatalf("status=%d multilingual=%v want=%v body=%s", rec.Code, repo.mintedMultilingual, want, rec.Body.String())
				}
				validateM2R12(t, "POST", path, body, rec)
			})
		}
	}
}
func TestM2R12PatchAndGetHTTP(t *testing.T) {
	path := "/api/v1/admin/videos/" + mediaVideoID + "/audio-tracks"
	patchPath := path + "/" + mediaTrackID
	track := domain.AudioTrack{ID: mediaTrackID, VideoAssetID: mediaVideoID, Kind: domain.EmbeddedTrack, Language: "en", Label: "Original", IsDefault: true, State: domain.Ready, SourceLeaf: domain.EmbeddedAudioLeaf, UpdatedAt: mediaNow}
	repo := audioTrackRepository()
	// The fake models the locked repository callback protocol; the integration case below
	// independently verifies the production repository's decision to skip a rewrite.
	repo.updateMetadata = func(_ context.Context, _ domain.Actor, _, _, language, label string, mode *bool, now time.Time, rewrite func([]domain.AudioTrack) error) (domain.AudioTrack, error) {
		changed := track.Language != language || track.Label != label
		track.Language, track.Label, track.UpdatedAt = language, label, now
		if mode != nil {
			track.Multilingual = *mode
		}
		if changed {
			if err := rewrite([]domain.AudioTrack{track}); err != nil {
				return domain.AudioTrack{}, err
			}
		}
		return track, nil
	}
	repo.listTracks = func(context.Context, domain.Actor, string) ([]domain.AudioTrack, error) {
		return []domain.AudioTrack{track}, nil
	}
	master := &countingMaster{}
	identity := &mediaIdentity{}
	server := mediaAPIWithAudio(t, repo, &mediaQueue{}, identity, true, master)
	for _, tc := range []struct {
		name, body string
		calls      int
		mode       bool
	}{
		{"flag only", `{"language":"en","label":"Original","multilingual":true}`, 0, true},
		{"language", `{"language":"uk","label":"Original"}`, 1, true},
		{"no-op", `{"language":"uk","label":"Original"}`, 1, true},
		{"label", `{"language":"uk","label":"Updated"}`, 2, true},
		{"explicit false", `{"language":"uk","label":"Updated","multilingual":false}`, 2, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rec := serveMedia(t, server, identity, mediaUser(mediaAdminID, identitydomain.Admin), true, "http://localhost:8080", "PATCH", patchPath, []byte(tc.body), "application/json", true)
			if rec.Code != 200 || master.calls != tc.calls || track.Multilingual != tc.mode {
				t.Fatalf("status=%d calls=%d mode=%v body=%s", rec.Code, master.calls, track.Multilingual, rec.Body.String())
			}
			var response struct {
				Multilingual *bool `json:"multilingual"`
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
				t.Fatal(err)
			}
			if response.Multilingual == nil || *response.Multilingual != tc.mode {
				t.Fatalf("PATCH flag=%s", rec.Body.String())
			}
			validateM2R12(t, "PATCH", patchPath, tc.body, rec)
		})
	}
	rec := serveMedia(t, server, identity, mediaUser(mediaAdminID, identitydomain.Admin), true, "", "GET", path, nil, "", true)
	var body struct {
		Items []struct {
			Multilingual *bool `json:"multilingual"`
		}
		Default *bool `json:"multilingualDefault"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if rec.Code != 200 || body.Default == nil || !*body.Default || len(body.Items) != 1 || body.Items[0].Multilingual == nil || *body.Items[0].Multilingual {
		t.Fatalf("GET=%s", rec.Body.String())
	}
	validateM2R12(t, "GET", path, "", rec)
}
func TestM2R12StrictOptionalFlagHTTP(t *testing.T) {
	path := "/api/v1/admin/videos/" + mediaVideoID + "/audio-tracks"
	for _, method := range []string{"POST", "PATCH"} {
		for _, field := range []string{`null`, `"true"`, `1`, `true,"multilingual":false`} {
			t.Run(method+"/"+field, func(t *testing.T) {
				body := `{"language":"uk","label":"Dub","multilingual":` + field + `}`
				target := path
				if method == "POST" {
					body = `{"declaredBytes":12,"originalName":"dub.aac","language":"uk","label":"Dub","multilingual":` + field + `}`
				} else {
					target += "/" + mediaTrackID
				}
				identity := &mediaIdentity{}
				rec := serveMedia(t, mediaAPIHandler(t, audioTrackRepository(), &mediaQueue{}, identity), identity, mediaUser(mediaAdminID, identitydomain.Admin), true, "http://localhost:8080", method, target, []byte(body), "application/json", false)
				if rec.Code != 400 {
					t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
				}
				assertMediaError(t, rec, "invalid_request")
			})
		}
	}
}
