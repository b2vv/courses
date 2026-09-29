package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	discussionhttp "courses/internal/discussion/adapters/http"
	"courses/internal/discussion/application"
	"courses/internal/discussion/domain"
	"courses/internal/discussion/ports"
	"courses/platform/httpserver"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/legacy"
)

const (
	discussionLesson  = "70000000-0000-4000-8000-000000000001"
	discussionAuthor  = "40000000-0000-4000-8000-000000000003"
	discussionOwner   = "40000000-0000-4000-8000-000000000002"
	discussionViewer  = "40000000-0000-4000-8000-000000000001"
	discussionAdmin   = "40000000-0000-4000-8000-000000000004"
	discussionComment = "70000000-0000-4000-8000-0000000000aa"
	discussionReply   = "70000000-0000-4000-8000-0000000000bb"
)

type fakeRepository struct {
	facts          ports.LessonFacts
	page           ports.CommentPage
	comments       map[string]ports.CommentRecord
	factsErr       error
	listErr        error
	insertErr      error
	updated        string
	queue          ports.QueuePage
	moderationLogs []ports.ModerationLogEntry
}

func (fake *fakeRepository) LessonFacts(context.Context, string) (ports.LessonFacts, error) {
	return fake.facts, fake.factsErr
}
func (fake *fakeRepository) List(context.Context, string, string, int) (ports.CommentPage, error) {
	return fake.page, fake.listErr
}
func (fake *fakeRepository) Get(_ context.Context, id string) (ports.CommentRecord, error) {
	record, ok := fake.comments[id]
	if !ok {
		return ports.CommentRecord{}, domain.ErrNotFound
	}
	return record, nil
}
func (fake *fakeRepository) InsertRoot(_ context.Context, comment domain.Comment) (ports.CommentRecord, error) {
	return fake.insert(comment, discussionComment)
}
func (fake *fakeRepository) InsertReply(_ context.Context, comment domain.Comment) (ports.CommentRecord, error) {
	if fake.insertErr != nil {
		return ports.CommentRecord{}, fake.insertErr
	}
	return fake.insert(comment, discussionReply)
}
func (fake *fakeRepository) insert(comment domain.Comment, id string) (ports.CommentRecord, error) {
	if fake.insertErr != nil {
		return ports.CommentRecord{}, fake.insertErr
	}
	comment.ID = id
	record := ports.CommentRecord{Comment: comment, Course: fake.facts.Course}
	fake.comments[id] = record
	return record, nil
}
func (fake *fakeRepository) UpdateBody(_ context.Context, id, _ string, body string, now time.Time) (ports.CommentRecord, error) {
	record, err := fake.Get(context.Background(), id)
	if err != nil {
		return ports.CommentRecord{}, err
	}
	record.Comment.Body, record.Comment.EditedAt = body, &now
	fake.comments[id], fake.updated = record, body
	return record, nil
}
func (fake *fakeRepository) Remove(_ context.Context, id, actorID string, now time.Time) error {
	record, err := fake.Get(context.Background(), id)
	if err != nil {
		return err
	}
	record.Comment.Body, record.Comment.RemovedAt = "", &now
	record.Comment.RemovedBy = &actorID
	fake.comments[id] = record
	return nil
}

func (fake *fakeRepository) CourseFacts(context.Context, string) (ports.CourseFacts, error) {
	return ports.CourseFacts{Course: fake.facts.Course, CourseID: "course"}, nil
}
func (fake *fakeRepository) ListByCourse(context.Context, string, string, int) (ports.CommentPage, error) {
	return fake.page, nil
}
func (fake *fakeRepository) InsertReport(context.Context, domain.Report, time.Time) error { return nil }
func (fake *fakeRepository) DismissReports(context.Context, string, string, time.Time) error {
	return nil
}
func (fake *fakeRepository) ListQueue(context.Context, string, string, int) (ports.QueuePage, error) {
	return fake.queue, nil
}
func (fake *fakeRepository) WriteModerationLog(_ context.Context, entry ports.ModerationLogEntry) error {
	fake.moderationLogs = append(fake.moderationLogs, entry)
	return nil
}

func newFake() *fakeRepository {
	now := time.Date(2026, 9, 28, 10, 0, 0, 0, time.UTC)
	name := "Ada"
	root := ports.CommentRecord{
		Comment:     domain.Comment{ID: discussionComment, LessonID: discussionLesson, AuthorID: discussionAuthor, Body: "hello", CreatedAt: now},
		DisplayName: &name, Course: domain.Course{OwnerID: discussionOwner, Status: domain.Published},
	}
	moderation := ports.ModerationCommentRecord{CommentRecord: root, LessonTitle: "Lesson one"}
	return &fakeRepository{
		facts: ports.LessonFacts{Course: root.Course, DurationSecs: floatPointer(100)},
		page: ports.CommentPage{
			Threads:  []ports.Thread{{Root: root}},
			Comments: []ports.ModerationCommentRecord{moderation},
		},
		comments: map[string]ports.CommentRecord{discussionComment: root},
		queue:    ports.QueuePage{Items: []ports.QueueItem{{ModerationCommentRecord: moderation, OpenReports: 1}}},
	}
}

func floatPointer(value float64) *float64 { return &value }
func stringPointer(value string) *string  { return &value }

// serve owns contract validation: every response from every D03 table case is checked against
// the real method/path/body and OpenAPI response schema. validRequest skips only request-schema
// validation for cases intentionally testing transport/domain rejection.
func serve(t *testing.T, repository ports.Repository, method, path, body string, actor *httpserver.Actor, validRequest bool) *httptest.ResponseRecorder {
	t.Helper()
	handler := discussionhttp.NewHandler(application.New(repository, func() time.Time { return time.Date(2026, 9, 28, 11, 0, 0, 0, time.UTC) }))
	server := httpserver.NewHandler(httpserver.Options{Service: "discussion-test", Register: func(mux *http.ServeMux) {
		handler.Register(mux, passThrough, passThrough)
		handler.RegisterTeach(mux, passThrough, passThrough)
	}})
	request := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	if actor != nil {
		request = request.WithContext(httpserver.WithActor(request.Context(), *actor))
	}
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)
	request.Body = io.NopCloser(bytes.NewReader([]byte(body)))
	assertContract(t, request, response, validRequest)
	return response
}

func passThrough(next http.Handler) http.Handler { return next }
func actor(id, role string) *httpserver.Actor    { return &httpserver.Actor{UserID: id, Role: role} }

func requireError(t *testing.T, response *httptest.ResponseRecorder, status int, code string) {
	t.Helper()
	if response.Code != status {
		t.Fatalf("status=%d want=%d body=%s", response.Code, status, response.Body.String())
	}
	var body struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Error.Code != code {
		t.Fatalf("code=%q want=%q body=%s", body.Error.Code, code, response.Body.String())
	}
}

func TestK10GetComments(t *testing.T) {
	cases := []struct {
		name       string
		repository func() *fakeRepository
		path       string
		actor      *httpserver.Actor
		status     int
		code       string
	}{
		{"200", newFake, "/api/v1/learn/lessons/" + discussionLesson + "/comments", actor(discussionViewer, "student"), http.StatusOK, ""},
		{"401 anonymous", newFake, "/api/v1/learn/lessons/" + discussionLesson + "/comments", nil, http.StatusUnauthorized, "authentication_required"},
		{"403 draft", func() *fakeRepository { fake := newFake(); fake.facts.Course.Status = domain.Draft; return fake }, "/api/v1/learn/lessons/" + discussionLesson + "/comments", actor(discussionViewer, "student"), http.StatusForbidden, "course_forbidden"},
		{"404 lesson", func() *fakeRepository { fake := newFake(); fake.factsErr = domain.ErrNotFound; return fake }, "/api/v1/learn/lessons/" + discussionLesson + "/comments", actor(discussionViewer, "student"), http.StatusNotFound, "lesson_not_found"},
		{"400 cursor", newFake, "/api/v1/learn/lessons/" + discussionLesson + "/comments?cursor=%%", actor(discussionViewer, "student"), http.StatusBadRequest, "comment_cursor_invalid"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			response := serve(t, test.repository(), http.MethodGet, test.path, "", test.actor, true)
			if test.code == "" {
				if response.Code != test.status {
					t.Fatalf("status=%d want=%d", response.Code, test.status)
				}
				return
			}
			requireError(t, response, test.status, test.code)
		})
	}
}

func TestK11PostComments(t *testing.T) {
	cases := []struct {
		name, body string
		valid      bool
		setup      func(*fakeRepository)
		status     int
		code       string
	}{
		{"201 root", `{"body":"new","positionMs":1000}`, true, nil, http.StatusCreated, ""},
		{"201 reply", `{"body":"reply","parentId":"` + discussionComment + `"}`, true, nil, http.StatusCreated, ""},
		{"400 body", `{"body":"   "}`, true, nil, http.StatusBadRequest, "comment_body_invalid"},
		{"400 position", `{"body":"bad","positionMs":100001}`, true, nil, http.StatusBadRequest, "comment_position_invalid"},
		{"400 parent", `{"body":"bad","parentId":"70000000-0000-4000-8000-0000000000cc"}`, true, nil, http.StatusBadRequest, "comment_parent_invalid"},
		{"409 full", `{"body":"reply","parentId":"` + discussionComment + `"}`, true, func(fake *fakeRepository) { fake.insertErr = domain.ErrThreadFull }, http.StatusConflict, "comment_thread_full"},
		{"403 course", `{"body":"new"}`, true, func(fake *fakeRepository) { fake.facts.Course.Status = domain.Draft }, http.StatusForbidden, "course_forbidden"},
		{"400 malformed JSON", `{"body":`, false, nil, http.StatusBadRequest, "invalid_request"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			fake := newFake()
			if test.setup != nil {
				test.setup(fake)
			}
			response := serve(t, fake, http.MethodPost, "/api/v1/learn/lessons/"+discussionLesson+"/comments", test.body, actor(discussionViewer, "student"), test.valid)
			if test.code == "" {
				if response.Code != test.status {
					t.Fatalf("status=%d want=%d body=%s", response.Code, test.status, response.Body.String())
				}
				return
			}
			requireError(t, response, test.status, test.code)
		})
	}
}

func TestK12PatchComments(t *testing.T) {
	cases := []struct {
		name, id string
		actor    *httpserver.Actor
		setup    func(*fakeRepository)
		status   int
		code     string
	}{
		{"200 author", discussionComment, actor(discussionAuthor, "student"), nil, http.StatusOK, ""},
		{"403 other", discussionComment, actor(discussionViewer, "student"), nil, http.StatusForbidden, "comment_forbidden"},
		{"404 unknown", "70000000-0000-4000-8000-0000000000cc", actor(discussionAuthor, "student"), nil, http.StatusNotFound, "comment_not_found"},
		{"403 other on removed", discussionComment, actor(discussionViewer, "student"), func(fake *fakeRepository) {
			now := time.Now()
			record := fake.comments[discussionComment]
			record.Comment.RemovedAt = &now
			fake.comments[discussionComment] = record
		}, http.StatusForbidden, "comment_forbidden"},
		{"409 removed", discussionComment, actor(discussionAuthor, "student"), func(fake *fakeRepository) {
			now := time.Now()
			record := fake.comments[discussionComment]
			record.Comment.RemovedAt = &now
			fake.comments[discussionComment] = record
		}, http.StatusConflict, "comment_removed"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			fake := newFake()
			if test.setup != nil {
				test.setup(fake)
			}
			response := serve(t, fake, http.MethodPatch, "/api/v1/learn/comments/"+test.id, `{"body":"edited"}`, test.actor, true)
			if test.code == "" {
				if response.Code != test.status || fake.updated != "edited" {
					t.Fatalf("status=%d updated=%q", response.Code, fake.updated)
				}
				var body struct {
					EditedAt *time.Time `json:"editedAt"`
				}
				if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
					t.Fatal(err)
				}
				if body.EditedAt == nil {
					t.Fatal("editedAt is missing")
				}
				return
			}
			requireError(t, response, test.status, test.code)
		})
	}
}

func TestK13DeleteComments(t *testing.T) {
	cases := []struct {
		name   string
		actor  *httpserver.Actor
		setup  func(*fakeRepository)
		status int
		code   string
	}{
		{"204 author", actor(discussionAuthor, "student"), nil, http.StatusNoContent, ""},
		{"204 admin", actor(discussionAdmin, "admin"), nil, http.StatusNoContent, ""},
		{"204 course owner", actor(discussionOwner, "student"), nil, http.StatusNoContent, ""},
		{"403 other", actor(discussionViewer, "student"), nil, http.StatusForbidden, "comment_forbidden"},
		{"204 repeat", actor(discussionAuthor, "student"), func(fake *fakeRepository) {
			now := time.Now()
			record := fake.comments[discussionComment]
			record.Comment.RemovedAt = &now
			fake.comments[discussionComment] = record
		}, http.StatusNoContent, ""},
		{"404 unknown", actor(discussionAuthor, "student"), nil, http.StatusNotFound, "comment_not_found"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			fake := newFake()
			if test.setup != nil {
				test.setup(fake)
			}
			id := discussionComment
			if test.name == "404 unknown" {
				id = "70000000-0000-4000-8000-0000000000cc"
			}
			response := serve(t, fake, http.MethodDelete, "/api/v1/learn/comments/"+id, "", test.actor, true)
			if test.code == "" {
				if response.Code != test.status {
					t.Fatalf("status=%d want=%d", response.Code, test.status)
				}
				return
			}
			requireError(t, response, test.status, test.code)
		})
	}
}

func TestK14CapabilitiesAndRemovedAuthorForm(t *testing.T) {
	cases := []struct {
		name               string
		actor              *httpserver.Actor
		canEdit, canRemove bool
	}{
		{"author", actor(discussionAuthor, "student"), true, true},
		{"other student", actor(discussionViewer, "student"), false, false},
		{"admin", actor(discussionAdmin, "admin"), false, true},
		{"course owner", actor(discussionOwner, "student"), false, true},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			response := serve(t, newFake(), http.MethodGet, "/api/v1/learn/lessons/"+discussionLesson+"/comments", "", test.actor, true)
			var body struct {
				Threads []struct {
					Root struct {
						CanEdit   bool `json:"canEdit"`
						CanRemove bool `json:"canRemove"`
					}
				}
			}
			if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			root := body.Threads[0].Root
			if root.CanEdit != test.canEdit || root.CanRemove != test.canRemove {
				t.Fatalf("capabilities=%+v", root)
			}
		})
	}

	t.Run("removed root with reply and deleted author", func(t *testing.T) {
		fake := newFake()
		now := time.Now()
		root := fake.comments[discussionComment]
		root.Comment.Body, root.Comment.RemovedAt, root.DisplayName = "", &now, nil
		reply := root
		reply.Comment.ID, reply.Comment.ParentID, reply.Comment.Body = discussionReply, stringPointer(discussionComment), "reply"
		fake.comments[discussionComment], fake.comments[discussionReply] = root, reply
		fake.page.Threads[0].Root, fake.page.Threads[0].Replies = root, []ports.CommentRecord{reply}
		response := serve(t, fake, http.MethodGet, "/api/v1/learn/lessons/"+discussionLesson+"/comments", "", actor(discussionViewer, "student"), true)
		var body struct {
			Threads []struct {
				Root struct {
					Author struct {
						DisplayName *string `json:"displayName"`
					}
					Body    *string `json:"body"`
					Removed bool    `json:"removed"`
				}
			}
		}
		if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		rootBody := body.Threads[0].Root
		if rootBody.Body != nil || !rootBody.Removed || rootBody.Author.DisplayName != nil {
			t.Fatalf("removed form=%+v", rootBody)
		}
	})
}

func TestI03ModerationContracts(t *testing.T) {
	t.Run("report accepted", func(t *testing.T) {
		response := serve(t, newFake(), http.MethodPost, "/api/v1/learn/comments/"+discussionComment+"/reports", "", actor(discussionViewer, "student"), true)
		if response.Code != http.StatusCreated {
			t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
		}
	})
	t.Run("report own comment forbidden", func(t *testing.T) {
		response := serve(t, newFake(), http.MethodPost, "/api/v1/learn/comments/"+discussionComment+"/reports", "", actor(discussionAuthor, "student"), true)
		requireError(t, response, http.StatusForbidden, "report_own_comment")
	})
	t.Run("report unknown comment", func(t *testing.T) {
		response := serve(t, newFake(), http.MethodPost, "/api/v1/learn/comments/70000000-0000-4000-8000-0000000000cc/reports", "", actor(discussionViewer, "student"), true)
		requireError(t, response, http.StatusNotFound, "comment_not_found")
	})
	t.Run("course comments", func(t *testing.T) {
		response := serve(t, newFake(), http.MethodGet, "/api/v1/teach/courses/"+discussionOwner+"/comments", "", actor(discussionOwner, "instructor"), true)
		if response.Code != http.StatusOK {
			t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
		}
		var body struct {
			Comments []struct {
				LessonTitle string `json:"lessonTitle"`
			} `json:"comments"`
		}
		if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		if len(body.Comments) != 1 || body.Comments[0].LessonTitle != "Lesson one" {
			t.Fatalf("comments=%+v", body.Comments)
		}
	})
	t.Run("course comments forbidden", func(t *testing.T) {
		response := serve(t, newFake(), http.MethodGet, "/api/v1/teach/courses/"+discussionOwner+"/comments", "", actor(discussionViewer, "student"), true)
		requireError(t, response, http.StatusForbidden, "course_forbidden")
	})
	t.Run("queue", func(t *testing.T) {
		response := serve(t, newFake(), http.MethodGet, "/api/v1/teach/courses/"+discussionOwner+"/moderation-queue", "", actor(discussionAdmin, "admin"), true)
		if response.Code != http.StatusOK {
			t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
		}
		var body struct {
			Items []struct {
				Comment struct {
					LessonTitle string `json:"lessonTitle"`
				} `json:"comment"`
			} `json:"items"`
		}
		if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		if len(body.Items) != 1 || body.Items[0].Comment.LessonTitle != "Lesson one" {
			t.Fatalf("items=%+v", body.Items)
		}
	})
	t.Run("invalid cursor", func(t *testing.T) {
		response := serve(t, newFake(), http.MethodGet, "/api/v1/teach/courses/"+discussionOwner+"/comments?cursor=%%", "", actor(discussionOwner, "instructor"), true)
		requireError(t, response, http.StatusBadRequest, "comment_cursor_invalid")
	})
	t.Run("dismiss", func(t *testing.T) {
		response := serve(t, newFake(), http.MethodPost, "/api/v1/teach/comments/"+discussionComment+"/dismiss-report", `{"reason":"not a violation"}`, actor(discussionOwner, "instructor"), true)
		if response.Code != http.StatusNoContent {
			t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
		}
	})
	t.Run("dismiss invalid reason", func(t *testing.T) {
		response := serve(t, newFake(), http.MethodPost, "/api/v1/teach/comments/"+discussionComment+"/dismiss-report", `{"reason":"   "}`, actor(discussionOwner, "instructor"), true)
		requireError(t, response, http.StatusBadRequest, "invalid_request")
	})
}

func assertContract(t *testing.T, request *http.Request, response *httptest.ResponseRecorder, validRequest bool) {
	t.Helper()
	document, err := openapi3.NewLoader().LoadFromFile(filepath.Join("..", "..", "..", "..", "openapi", "openapi.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	router, err := legacy.NewRouter(document)
	if err != nil {
		t.Fatal(err)
	}
	route, params, err := router.FindRoute(request)
	if err != nil {
		t.Fatal(err)
	}
	input := &openapi3filter.RequestValidationInput{Request: request, PathParams: params, Route: route}
	if validRequest {
		if err := openapi3filter.ValidateRequest(context.Background(), input); err != nil {
			t.Fatalf("request contract: %v", err)
		}
	}
	content := response.Body.Bytes()
	if response.Code == http.StatusNoContent {
		content = nil
	}
	if err := openapi3filter.ValidateResponse(context.Background(), &openapi3filter.ResponseValidationInput{RequestValidationInput: input, Status: response.Code, Header: response.Header(), Body: io.NopCloser(bytes.NewReader(content))}); err != nil {
		t.Fatalf("response contract: %v; body=%s", err, response.Body.String())
	}
}
