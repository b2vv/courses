package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"courses/internal/contracts/access"
	"courses/internal/course/application"
	"courses/internal/course/domain"
	"courses/internal/course/ports"
	"courses/platform/httpserver"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/legacy"
)

const documentLessonID = "20000000-0000-4000-8000-000000000001"
const documentOwnerID = "40000000-0000-4000-8000-000000000002"
const documentAdminPath = "/api/v1/admin/lessons/" + documentLessonID + "/document"
const documentLearnPath = "/api/v1/learn/lessons/" + documentLessonID + "/document"
const documentToken = "2026-10-12T10:00:00.123456Z"
const validDocumentJSON = `{"schemaVersion":1,"locale":"en","sections":[{"id":"v","type":"video"}]}`

type documentRepositoryFake struct {
	draft, published                   domain.LessonDocument
	found                              bool
	updated, expected                  time.Time
	saveCalls, publishCalls, readCalls int
	lessonID, locale                   string
	failure                            error
}

var _ ports.LessonDocumentRepository = (*documentRepositoryFake)(nil)

func (f *documentRepositoryFake) SaveDraft(_ context.Context, id, locale string, body domain.LessonDocument, expected time.Time) (time.Time, error) {
	f.saveCalls++
	f.lessonID = id
	f.locale = locale
	f.expected = expected
	if f.failure != nil {
		return time.Time{}, f.failure
	}
	f.draft = body
	f.found = true
	f.updated = f.updated.Add(time.Microsecond)
	return f.updated, nil
}
func (f *documentRepositoryFake) PublishDocument(_ context.Context, id, locale string, expected time.Time) error {
	f.publishCalls++
	f.lessonID = id
	f.locale = locale
	f.expected = expected
	if f.failure != nil {
		return f.failure
	}
	f.published = f.draft
	f.updated = f.updated.Add(time.Microsecond)
	return nil
}
func (f *documentRepositoryFake) LoadDraft(_ context.Context, id, locale string) (domain.LessonDocument, time.Time, bool, error) {
	f.readCalls++
	f.lessonID = id
	f.locale = locale
	return f.draft, f.updated, f.found, f.failure
}
func (f *documentRepositoryFake) LoadPublished(_ context.Context, id, locale string) (domain.LessonDocument, bool, error) {
	f.readCalls++
	f.lessonID = id
	f.locale = locale
	return f.published, f.found, f.failure
}

type documentAccessFake struct {
	facts            domain.LessonDocumentAccess
	calls            int
	userID, lessonID string
	failure          error
}

var _ ports.LessonDocumentAccessReader = (*documentAccessFake)(nil)

func (f *documentAccessFake) LoadLessonAccess(_ context.Context, userID, lessonID string) (domain.LessonDocumentAccess, error) {
	f.calls++
	f.userID = userID
	f.lessonID = lessonID
	return f.facts, f.failure
}
func documentFixture() (*documentRepositoryFake, *documentAccessFake, httpserver.Actor) {
	var body domain.LessonDocument
	if err := json.Unmarshal([]byte(validDocumentJSON), &body); err != nil {
		panic(err)
	}
	updated, _ := parseAdminTime(documentToken)
	return &documentRepositoryFake{draft: body, published: body, found: true, updated: updated}, &documentAccessFake{facts: domain.LessonDocumentAccess{OwnerID: documentOwnerID, Status: domain.Published, Paid: true}}, httpserver.Actor{UserID: documentOwnerID, Role: "instructor", EmailVerified: true}
}
func documentMux(repo *documentRepositoryFake, reader *documentAccessFake, actor httpserver.Actor, origin func(http.Handler) http.Handler) http.Handler {
	authenticate := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if actor.UserID != "" {
				r = r.WithContext(httpserver.WithActor(r.Context(), actor))
			}
			next.ServeHTTP(w, r)
		})
	}
	if origin == nil {
		origin = func(h http.Handler) http.Handler { return h }
	}
	h := NewDocumentHandler(application.NewDocuments(repo, reader))
	return httpserver.NewHandler(httpserver.Options{Service: "api", Register: func(mux *http.ServeMux) {
		h.Register(mux, authenticate)
		h.RegisterAdmin(mux, authenticate, origin)
	}})
}
func documentRequest(t *testing.T, mux http.Handler, method, path, body string, status int, code string) *httptest.ResponseRecorder {
	t.Helper()
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)
	if w.Code != status {
		t.Fatalf("status=%d want=%d body=%s", w.Code, status, w.Body.String())
	}
	if w.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("missing no-store")
	}
	if code != "" {
		var envelope struct {
			Error struct {
				Code string `json:"code"`
			} `json:"error"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &envelope); err != nil {
			t.Fatal(err)
		}
		if envelope.Error.Code != code {
			t.Fatalf("body=%s want code=%s", w.Body.String(), code)
		}
	}
	return w
}
func TestDocumentInvalidDraftNeverPersists(t *testing.T) {
	for _, section := range []string{`{"id":"bad","type":"iframe"}`, `{"id":"v","type":"video","content":{"blocks":[]}}`, `{"id":"row","type":"row","columns":[{"id":"nested","type":"row","columns":[]}]}`} {
		t.Run(section, func(t *testing.T) {
			repo, reader, actor := documentFixture()
			body := `{"document":{"schemaVersion":1,"locale":"uk","sections":[` + section + `]}}`
			documentRequest(t, documentMux(repo, reader, actor, nil), "PUT", documentAdminPath, body, 400, "invalid_lesson_document")
			if repo.saveCalls != 0 {
				t.Fatal("invalid document persisted")
			}
		})
	}
}
func TestDocumentNonEditor(t *testing.T) {
	for _, method := range []string{"GET", "PUT", "POST"} {
		t.Run(method, func(t *testing.T) {
			repo, reader, actor := documentFixture()
			reader.facts.OwnerID = "40000000-0000-4000-8000-000000000099"
			path, body := documentAdminPath, `{"document":`+validDocumentJSON+`}`
			if method == "POST" {
				path += "/publish"
				body = `{"expectedUpdatedAt":"` + documentToken + `"}`
			}
			documentRequest(t, documentMux(repo, reader, actor, nil), method, path, body, 403, "forbidden")
			if repo.readCalls+repo.saveCalls+repo.publishCalls != 0 {
				t.Fatal("non-editor reached repository")
			}
		})
	}
}
func TestDocumentPublishedAccess(t *testing.T) {
	for _, tc := range []struct {
		name                             string
		verified, member, preview, found bool
		status                           int
		code                             string
		accessCalls                      int
	}{
		{"paid nonmember", true, false, false, true, 403, "course_purchase_required", 1},
		{"active member", true, true, false, true, 200, "", 1},
		{"email first", false, true, false, true, 403, "email_not_verified", 0},
		{"absent publication", true, true, false, false, 200, "", 1},
		{"verified preview", true, false, true, true, 200, "", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo, reader, actor := documentFixture()
			actor.UserID = "40000000-0000-4000-8000-000000000003"
			actor.Role = "student"
			actor.EmailVerified = tc.verified
			repo.found = tc.found
			reader.facts.Preview = tc.preview
			if tc.member {
				reader.facts.Membership = access.MembershipActive
				reader.facts.Entitled = true
			}
			repo.draft = domain.LessonDocument{SchemaVersion: 1, Locale: "en", Sections: []domain.Section{}}
			w := documentRequest(t, documentMux(repo, reader, actor, nil), "GET", documentLearnPath+"?locale=en", "", tc.status, tc.code)
			if reader.calls != tc.accessCalls {
				t.Fatalf("access calls=%d", reader.calls)
			}
			if tc.status != 200 {
				if repo.readCalls != 0 {
					t.Fatal("denied request read content")
				}
				return
			}
			var doc domain.LessonDocument
			if err := json.Unmarshal(w.Body.Bytes(), &doc); err != nil {
				t.Fatal(err)
			}
			if err := domain.ValidateLessonDocument(doc); err != nil {
				t.Fatal(err)
			}
			if doc.Locale != "en" || doc.SchemaVersion != domain.CurrentLessonDocumentSchemaVersion {
				t.Fatalf("document=%+v", doc)
			}
			if !tc.found {
				if len(doc.Sections) != 0 || !strings.Contains(w.Body.String(), `"sections":[]`) {
					t.Fatal(w.Body.String())
				}
			} else if !reflect.DeepEqual(doc, repo.published) {
				t.Fatal("published response leaked draft or changed document")
			}
			if reader.userID != actor.UserID || reader.lessonID != documentLessonID {
				t.Fatal("incorrect access identity")
			}
		})
	}
}
func TestDocumentConflicts(t *testing.T) {
	for _, method := range []string{"PUT", "POST"} {
		t.Run(method, func(t *testing.T) {
			repo, reader, actor := documentFixture()
			repo.failure = fmt.Errorf("stale: %w", domain.ErrDocumentConflict)
			path, body := documentAdminPath, `{"expectedUpdatedAt":"`+documentToken+`","document":`+validDocumentJSON+`}`
			if method == "POST" {
				path += "/publish?locale=en"
				body = `{"expectedUpdatedAt":"` + documentToken + `"}`
			}
			documentRequest(t, documentMux(repo, reader, actor, nil), method, path, body, 409, "document_conflict")
			expected, _ := parseAdminTime(documentToken)
			if !repo.expected.Equal(expected) || repo.locale != "en" {
				t.Fatal("token/locale not forwarded")
			}
		})
	}
}
func TestDocumentDraftRoundTrip(t *testing.T) {
	repo, reader, actor := documentFixture()
	repo.found = false
	mux := documentMux(repo, reader, actor, nil)
	w := documentRequest(t, mux, "GET", documentAdminPath, "", 200, "")
	var draft lessonDocumentDraftResponse
	if err := json.Unmarshal(w.Body.Bytes(), &draft); err != nil {
		t.Fatal(err)
	}
	if draft.UpdatedAt != nil || draft.Document.Locale != "uk" || len(draft.Document.Sections) != 0 {
		t.Fatal(w.Body.String())
	}
	for _, token := range []string{"", `,"expectedUpdatedAt":null`, `,"expectedUpdatedAt":"` + documentToken + `"`} {
		w = documentRequest(t, mux, "PUT", documentAdminPath, `{"document":`+validDocumentJSON+token+`}`, 200, "")
		var saved lessonDocumentUpdatedResponse
		if err := json.Unmarshal(w.Body.Bytes(), &saved); err != nil {
			t.Fatal(err)
		}
		if token == "" || strings.Contains(token, "null") {
			if !repo.expected.IsZero() {
				t.Fatal("first write token not zero")
			}
		}
		w = documentRequest(t, mux, "GET", documentAdminPath+"?locale=en", "", 200, "")
		if err := json.Unmarshal(w.Body.Bytes(), &draft); err != nil {
			t.Fatal(err)
		}
		if draft.UpdatedAt == nil || *draft.UpdatedAt != saved.UpdatedAt || !reflect.DeepEqual(draft.Document, repo.draft) {
			t.Fatal("draft round trip failed")
		}
		w = documentRequest(t, mux, "PUT", documentAdminPath, `{"document":`+validDocumentJSON+`,"expectedUpdatedAt":"`+*draft.UpdatedAt+`"}`, 200, "")
		expected, _ := parseAdminTime(*draft.UpdatedAt)
		if !repo.expected.Equal(expected) {
			t.Fatal("GET token not passed to PUT")
		}
		if err := json.Unmarshal(w.Body.Bytes(), &saved); err != nil {
			t.Fatal(err)
		}
		w = documentRequest(t, mux, "POST", documentAdminPath+"/publish?locale=en", `{"expectedUpdatedAt":"`+saved.UpdatedAt+`"}`, 200, "")
		if w.Body.Len() != 0 {
			t.Fatal("publish body must be empty")
		}
		expected, _ = parseAdminTime(saved.UpdatedAt)
		if !repo.expected.Equal(expected) || repo.lessonID != documentLessonID || repo.locale != "en" {
			t.Fatal("publish token/identity lost")
		}
	}
}
func TestDocumentTransportRejections(t *testing.T) {
	for _, tc := range []struct{ name, method, path, body, code string }{
		{"unknown query", "GET", documentAdminPath + "?other=x", "", "invalid_query"},
		{"unknown locale", "GET", documentLearnPath + "?locale=fr", "", "invalid_query"},
		{"empty locale", "GET", documentAdminPath + "?locale=", "", "invalid_query"},
		{"duplicate locale", "GET", documentLearnPath + "?locale=uk&locale=en", "", "invalid_query"},
		{"malformed query", "GET", documentLearnPath + "?locale=%zz", "", "invalid_query"},
		{"PUT query", "PUT", documentAdminPath + "?locale=en", `{}`, "invalid_query"},
		{"publish query", "POST", documentAdminPath + "/publish?other=x", `{}`, "invalid_query"},
		{"unknown wrapper", "PUT", documentAdminPath, `{"other":0,"document":` + validDocumentJSON + `}`, "invalid_request"},
		{"unknown nested field", "PUT", documentAdminPath, `{"document":{"schemaVersion":1,"locale":"uk","sections":[{"id":"v","type":"video","other":1}]}}`, "invalid_request"},
		{"duplicate nested field", "PUT", documentAdminPath, `{"document":{"schemaVersion":1,"locale":"uk","locale":"en","sections":[]}}`, "invalid_request"},
		{"trailing JSON", "PUT", documentAdminPath, `{"document":` + validDocumentJSON + `} {}`, "invalid_request"},
		{"null wrapper", "PUT", documentAdminPath, `null`, "invalid_request"},
		{"bad timestamp", "PUT", documentAdminPath, `{"expectedUpdatedAt":"2026-10-12T10:00:00Z","document":` + validDocumentJSON + `}`, "invalid_request"},
		{"invalid document locale", "PUT", documentAdminPath, `{"document":{"schemaVersion":1,"locale":"fr","sections":[]}}`, "invalid_lesson_document"},
		{"publish missing token", "POST", documentAdminPath + "/publish", `{}`, "invalid_request"},
		{"publish null token", "POST", documentAdminPath + "/publish", `{"expectedUpdatedAt":null}`, "invalid_request"},
		{"publish unknown field", "POST", documentAdminPath + "/publish", `{"expectedUpdatedAt":"` + documentToken + `","other":1}`, "invalid_request"},
		{"publish trailing bytes", "POST", documentAdminPath + "/publish", `{"expectedUpdatedAt":"` + documentToken + `"} x`, "invalid_request"},
		{"publish too large", "POST", documentAdminPath + "/publish", strings.Repeat(" ", 1025) + `{}`, "invalid_request"},
		{"PUT too large", "PUT", documentAdminPath, strings.Repeat(" ", maximumDocumentWriteBody) + `{}`, "invalid_request"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo, reader, actor := documentFixture()
			documentRequest(t, documentMux(repo, reader, actor, nil), tc.method, tc.path, tc.body, 400, tc.code)
			if repo.saveCalls+repo.publishCalls+repo.readCalls != 0 {
				t.Fatal("rejected request reached persistence")
			}
		})
	}
}
func TestDocumentWriteLimitExceedsAdminLimit(t *testing.T) {
	repo, reader, actor := documentFixture()
	body := strings.Repeat(" ", maximumAdminWriteBody) + `{"document":` + validDocumentJSON + `}`
	documentRequest(t, documentMux(repo, reader, actor, nil), "PUT", documentAdminPath, body, 200, "")
	if repo.saveCalls != 1 {
		t.Fatal("large valid body not saved")
	}
}
func TestDocumentErrorMapping(t *testing.T) {
	for _, tc := range []struct {
		err    error
		status int
		code   string
	}{
		{domain.ErrInvalidLessonDocument, 400, "invalid_lesson_document"}, {domain.ErrInvalidRequest, 400, "invalid_request"}, {domain.ErrInvalidQuery, 400, "invalid_query"},
		{domain.ErrEmailNotVerified, 403, "email_not_verified"}, {domain.ErrPurchaseRequired, 403, "course_purchase_required"}, {domain.ErrForbidden, 403, "forbidden"}, {domain.ErrCourseForbidden, 403, "course_forbidden"},
		{domain.ErrLessonNotFound, 404, "lesson_not_found"}, {domain.ErrDocumentConflict, 409, "document_conflict"}, {domain.ErrDependencyUnavailable, 503, "dependency_unavailable"}, {errors.New("unexpected"), 500, "internal_error"},
	} {
		t.Run(tc.code, func(t *testing.T) {
			repo, reader, actor := documentFixture()
			reader.failure = fmt.Errorf("wrapped: %w", tc.err)
			documentRequest(t, documentMux(repo, reader, actor, nil), "GET", documentAdminPath, "", tc.status, tc.code)
		})
	}
}
func TestLearningOutlineHasNoDocumentBody(t *testing.T) {
	// Bodies belong exclusively to /learn/lessons/{lessonId}/document, never the outline.
	fields := reflect.TypeFor[learningLessonDTO]()
	for i := 0; i < fields.NumField(); i++ {
		field := fields.Field(i)
		name := strings.Split(field.Tag.Get("json"), ",")[0]
		switch name {
		case "document", "sections", "body":
			t.Fatalf("outline leaks %s", name)
		}
		if field.Anonymous {
			t.Fatal("inspect embedded outline fields before extending this guard")
		}
	}
}

func TestDocumentHTTPMatchesOpenAPI(t *testing.T) {
	spec, err := openapi3.NewLoader().LoadFromFile("../../../../openapi/openapi.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if err := spec.Validate(context.Background()); err != nil {
		t.Fatal(err)
	}
	router, err := legacy.NewRouter(spec)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ method, path, body string }{
		{"GET", documentAdminPath + "?locale=en", ""},
		{"PUT", documentAdminPath, `{"document":` + validDocumentJSON + `,"expectedUpdatedAt":null}`},
		{"POST", documentAdminPath + "/publish", `{"expectedUpdatedAt":"` + documentToken + `"}`},
		{"GET", documentLearnPath + "?locale=en", ""},
	} {
		t.Run(tc.method+tc.path, func(t *testing.T) {
			repo, reader, actor := documentFixture()
			r := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			r.Header.Set("Content-Type", "application/json")
			route, params, err := router.FindRoute(r)
			if err != nil {
				t.Fatal(err)
			}
			input := &openapi3filter.RequestValidationInput{Request: r, PathParams: params, Route: route}
			if err := openapi3filter.ValidateRequest(context.Background(), input); err != nil {
				t.Fatal(err)
			}
			w := httptest.NewRecorder()
			documentMux(repo, reader, actor, nil).ServeHTTP(w, r)
			if w.Code != 200 {
				t.Fatal(w.Body.String())
			}
			output := &openapi3filter.ResponseValidationInput{RequestValidationInput: input, Status: w.Code, Header: w.Header()}
			output.SetBodyBytes(w.Body.Bytes())
			if err := openapi3filter.ValidateResponse(context.Background(), output); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestDocumentAuthenticationAndOriginWiring(t *testing.T) {
	for _, tc := range []struct{ method, path, body string }{
		{"GET", documentAdminPath, ""}, {"GET", documentLearnPath, ""},
		{"PUT", documentAdminPath, `{"document":` + validDocumentJSON + `}`},
		{"POST", documentAdminPath + "/publish", `{"expectedUpdatedAt":"` + documentToken + `"}`},
	} {
		t.Run(tc.method+tc.path, func(t *testing.T) {
			repo, reader, actor := documentFixture()
			documentRequest(t, documentMux(repo, reader, httpserver.Actor{}, nil), tc.method, tc.path, tc.body, 401, "authentication_required")
			if reader.calls != 0 {
				t.Fatal("anonymous request reached access reader")
			}
			originCalls := 0
			origin := func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { originCalls++; w.WriteHeader(http.StatusForbidden) })
			}
			handler := documentMux(repo, reader, actor, origin)
			r := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			r.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, r)
			if tc.method == "GET" {
				if originCalls != 0 || w.Code != 200 {
					t.Fatal("read unexpectedly used origin middleware")
				}
			} else if originCalls != 1 || w.Code != 403 || repo.saveCalls+repo.publishCalls != 0 {
				t.Fatal("mutation bypassed origin middleware")
			}
		})
	}
}

func TestDocumentRecursiveSchemaRoundTrip(t *testing.T) {
	spec, err := openapi3.NewLoader().LoadFromFile("../../../../openapi/openapi.yaml")
	if err != nil {
		t.Fatal(err)
	}
	schema := spec.Components.Schemas["LessonDocument"].Value
	for _, body := range []string{
		`{"schemaVersion":1,"locale":"en","sections":[{"id":"row","type":"row","columns":[{"id":"text","type":"text","content":{"blocks":[{"type":"heading","level":2,"spans":[{"text":"Title","marks":["bold","italic"],"href":"https://example.com"}]},{"type":"list","ordered":true,"items":[[{"text":"Item"}]]}]}},{"id":"image","type":"image","imageId":"image-1","alt":"Image"}]},{"id":"task","type":"task","content":{"blocks":[]}},{"id":"video","type":"video","content":null}]}`,
		`{"schemaVersion":1,"locale":"uk","sections":null}`,
		`{"schemaVersion":1,"locale":"uk","sections":[{"id":"text","type":"text","content":{"blocks":null}}]}`,
	} {
		t.Run(body, func(t *testing.T) {
			var input interface{}
			if err := json.Unmarshal([]byte(body), &input); err != nil {
				t.Fatal(err)
			}
			if err := schema.VisitJSON(input); err != nil {
				t.Fatal(err)
			}
			repo, reader, actor := documentFixture()
			mux := documentMux(repo, reader, actor, nil)
			documentRequest(t, mux, "PUT", documentAdminPath, `{"document":`+body+`}`, 200, "")
			w := documentRequest(t, mux, "GET", documentAdminPath, "", 200, "")
			var draft lessonDocumentDraftResponse
			if err := json.Unmarshal(w.Body.Bytes(), &draft); err != nil {
				t.Fatal(err)
			}
			raw, err := json.Marshal(draft.Document)
			if err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal(raw, &input); err != nil {
				t.Fatal(err)
			}
			if err := schema.VisitJSON(input); err != nil {
				t.Fatal(err)
			}
		})
	}
}
