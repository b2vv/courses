package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
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
	discussionLesson     = "70000000-0000-4000-8000-000000000001"
	discussionAuthor     = "40000000-0000-4000-8000-000000000003"
	discussionOwner      = "40000000-0000-4000-8000-000000000002"
	discussionViewer     = "40000000-0000-4000-8000-000000000001"
	discussionAdmin      = "40000000-0000-4000-8000-000000000004"
	discussionOtherOwner = "40000000-0000-4000-8000-000000000005"
	discussionComment    = "70000000-0000-4000-8000-0000000000aa"
	discussionReply      = "70000000-0000-4000-8000-0000000000bb"
	discussionReview     = "70000000-0000-4000-8000-0000000000cc"
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
	reports        []domain.Report
	dismissed      []string
	moderationLogs []ports.ModerationLogEntry
	reviewFacts    ports.ReviewFactsRecord
	reviews        map[string]ports.ReviewRecord
	publicCourse   string
	publicErr      error
	reviewPage     ports.ReviewPage
	reviewCount    int
	reviewSum      int
	reviewReports  map[string]bool
}

// ReviewRepository is embedded in the production repository contract. The HTTP
// fixture exercises comment routes only; these deliberately inert methods keep
// that fixture honest without inventing review behaviour in this slice.
func (fake *fakeRepository) PublishedReviewCourse(context.Context, string) (string, error) {
	if fake.publicErr != nil {
		return "", fake.publicErr
	}
	return fake.publicCourse, nil
}
func (fake *fakeRepository) GetReview(_ context.Context, id string) (ports.ReviewRecord, error) {
	review, ok := fake.reviews[id]
	if !ok {
		return ports.ReviewRecord{}, domain.ErrNotFound
	}
	return review, nil
}
func (fake *fakeRepository) ReviewFacts(context.Context, string, string) (ports.ReviewFactsRecord, error) {
	return fake.reviewFacts, nil
}
func (fake *fakeRepository) CreateReview(_ context.Context, review ports.ReviewRecord) (ports.ReviewRecord, error) {
	if fake.reviewFacts.Removed {
		return ports.ReviewRecord{}, domain.ErrReviewRemoved
	}
	if fake.reviewFacts.Exists {
		return ports.ReviewRecord{}, domain.ErrReviewAlreadyExists
	}
	fake.reviewFacts.Exists = true
	fake.reviews[review.ID] = review
	return review, nil
}
func (fake *fakeRepository) UpdateReview(_ context.Context, _ string, author string, content domain.ReviewContent, progress domain.ReviewProgress, now time.Time) (ports.ReviewRecord, error) {
	for id, review := range fake.reviews {
		if review.AuthorID == author {
			if review.RemovedAt != nil {
				return ports.ReviewRecord{}, domain.ErrReviewRemoved
			}
			review.Stars, review.Body, review.CompletedLessons, review.TotalLessons, review.EditedAt = content.Stars, content.Body, progress.Completed, progress.Total, &now
			fake.reviews[id] = review
			return review, nil
		}
	}
	return ports.ReviewRecord{}, domain.ErrNotFound
}
func (fake *fakeRepository) DeleteReview(_ context.Context, _ string, author string) error {
	for id, review := range fake.reviews {
		if review.AuthorID == author {
			if review.RemovedAt != nil {
				return domain.ErrReviewRemoved
			}
			delete(fake.reviews, id)
			return nil
		}
	}
	return domain.ErrNotFound
}
func (fake *fakeRepository) GetOwnReview(_ context.Context, _ string, author string) (*ports.ReviewRecord, error) {
	for _, review := range fake.reviews {
		if review.AuthorID == author {
			return &review, nil
		}
	}
	return nil, nil
}
func (fake *fakeRepository) ListReviews(context.Context, string, string, int) (ports.ReviewPage, error) {
	return fake.reviewPage, nil
}
func (fake *fakeRepository) ReviewSummary(context.Context, string) (int, int, error) {
	return fake.reviewCount, fake.reviewSum, nil
}
func (fake *fakeRepository) InsertReviewReport(_ context.Context, reviewID, actor string, _ time.Time) error {
	review, ok := fake.reviews[reviewID]
	if !ok {
		return domain.ErrNotFound
	}
	if review.AuthorID == actor {
		return domain.ErrReportOwnReview
	}
	fake.reviewReports[reviewID+actor] = true
	return nil
}
func (fake *fakeRepository) DismissReviewReports(context.Context, string, string, time.Time) error {
	return nil
}
func (fake *fakeRepository) RemoveReview(_ context.Context, id, actor, _ string, _ *string, now time.Time) error {
	review, ok := fake.reviews[id]
	if !ok {
		return domain.ErrNotFound
	}
	review.RemovedAt = &now
	review.RemovedBy = &actor
	fake.reviews[id] = review
	return nil
}
func (fake *fakeRepository) ListModerationQueue(context.Context, string, string, int) (ports.ModerationQueuePage, error) {
	items := make([]ports.ModerationQueueItem, len(fake.queue.Items))
	for i, item := range fake.queue.Items {
		record := item.ModerationCommentRecord
		items[i] = ports.ModerationQueueItem{Kind: "comment", ID: record.Comment.ID, DisplayName: record.DisplayName, Body: record.Comment.Body, CreatedAt: record.Comment.CreatedAt, OpenReports: item.OpenReports, LessonID: &record.Comment.LessonID, LessonTitle: &record.LessonTitle}
	}
	stars := 5
	items = append(items, ports.ModerationQueueItem{Kind: "review", ID: discussionReview, DisplayName: stringPointer("Review author"), Stars: &stars, Body: "a sufficiently long review body", CreatedAt: time.Date(2026, 10, 3, 0, 0, 0, 0, time.UTC), OpenReports: 1})
	return ports.ModerationQueuePage{Items: items, NextCursor: fake.queue.NextCursor}, nil
}

func (fake *fakeRepository) LessonFacts(context.Context, string, string) (ports.LessonFacts, error) {
	return fake.facts, fake.factsErr
}
func (fake *fakeRepository) List(context.Context, string, string, int) (ports.CommentPage, error) {
	return fake.page, fake.listErr
}
func (fake *fakeRepository) Get(_ context.Context, _ string, id string) (ports.CommentRecord, error) {
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
	record, err := fake.Get(context.Background(), "actor", id)
	if err != nil {
		return ports.CommentRecord{}, err
	}
	record.Comment.Body, record.Comment.EditedAt = body, &now
	fake.comments[id], fake.updated = record, body
	return record, nil
}
func (fake *fakeRepository) Remove(_ context.Context, id, actorID string, now time.Time) error {
	record, err := fake.Get(context.Background(), actorID, id)
	if err != nil {
		return err
	}
	record.Comment.Body, record.Comment.RemovedAt = "", &now
	record.Comment.RemovedBy = &actorID
	fake.comments[id] = record
	return nil
}

func (fake *fakeRepository) CourseFacts(context.Context, string) (ports.CourseFacts, error) {
	if fake.factsErr != nil {
		return ports.CourseFacts{}, fake.factsErr
	}
	return ports.CourseFacts{Course: fake.facts.Course, CourseID: "course"}, nil
}
func (fake *fakeRepository) ListByCourse(context.Context, string, string, int) (ports.CommentPage, error) {
	return fake.page, fake.listErr
}
func (fake *fakeRepository) InsertReport(_ context.Context, report domain.Report, _ time.Time) error {
	fake.reports = append(fake.reports, report)
	return nil
}
func (fake *fakeRepository) DismissReports(_ context.Context, commentID, _ string, _ time.Time) error {
	fake.dismissed = append(fake.dismissed, commentID)
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
		comments:     map[string]ports.CommentRecord{discussionComment: root},
		queue:        ports.QueuePage{Items: []ports.QueueItem{{ModerationCommentRecord: moderation, OpenReports: 1}}},
		publicCourse: "course", reviews: map[string]ports.ReviewRecord{}, reviewReports: map[string]bool{}, reviewFacts: ports.ReviewFactsRecord{Facts: domain.ReviewFacts{Status: domain.Published, HasProgress: true}, CompletedLessons: 1, TotalLessons: 2},
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
		{"Pay19 paid course", func() *fakeRepository { fake := newFake(); fake.facts.Course.Paid = true; return fake }, "/api/v1/learn/lessons/" + discussionLesson + "/comments", actor(discussionViewer, "student"), http.StatusForbidden, "course_purchase_required"},
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
		{"Pay19 paid course", `{"body":"new"}`, true, func(fake *fakeRepository) { fake.facts.Course.Paid = true }, http.StatusForbidden, "course_purchase_required"},
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

func TestW12W13ModerationLogEffects(t *testing.T) {
	cases := []struct {
		name     string
		actor    *httpserver.Actor
		wantLog  bool
		wantRole string
	}{
		{name: "W12 author self-removal is not moderation", actor: actor(discussionAuthor, "student")},
		{name: "W13 owner removal is moderation", actor: actor(discussionOwner, "instructor"), wantLog: true, wantRole: "instructor"},
		{name: "W13 admin removal is moderation", actor: actor(discussionAdmin, "admin"), wantLog: true, wantRole: "admin"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			fake := newFake()
			response := serve(t, fake, http.MethodDelete, "/api/v1/learn/comments/"+discussionComment, "", test.actor, true)
			if response.Code != http.StatusNoContent {
				t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
			}
			record := fake.comments[discussionComment]
			if record.Comment.RemovedBy == nil || *record.Comment.RemovedBy != test.actor.UserID {
				t.Fatalf("removedBy=%v want=%s", record.Comment.RemovedBy, test.actor.UserID)
			}
			if !test.wantLog {
				if len(fake.moderationLogs) != 0 {
					t.Fatalf("moderation logs=%+v want none", fake.moderationLogs)
				}
				return
			}
			if len(fake.moderationLogs) != 1 {
				t.Fatalf("moderation logs=%+v want one", fake.moderationLogs)
			}
			entry := fake.moderationLogs[0]
			if entry.ActorID != test.actor.UserID || entry.ActorRole != test.wantRole || entry.CommentID != discussionComment || entry.Action != "remove" || entry.Reason != nil {
				t.Fatalf("moderation log=%+v", entry)
			}
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
		fake := newFake()
		response := serve(t, fake, http.MethodPost, "/api/v1/learn/comments/"+discussionComment+"/reports", "", actor(discussionViewer, "student"), true)
		if response.Code != http.StatusCreated {
			t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
		}
		if len(fake.reports) != 1 || fake.reports[0].CommentID != discussionComment || fake.reports[0].ReporterID != discussionViewer {
			t.Fatalf("reports=%+v", fake.reports)
		}
	})
	t.Run("report own comment forbidden", func(t *testing.T) {
		fake := newFake()
		response := serve(t, fake, http.MethodPost, "/api/v1/learn/comments/"+discussionComment+"/reports", "", actor(discussionAuthor, "student"), true)
		requireError(t, response, http.StatusForbidden, "report_own_comment")
		if len(fake.reports) != 0 {
			t.Fatalf("reports=%+v want none", fake.reports)
		}
	})
	t.Run("W17 report without course access is forbidden without a write", func(t *testing.T) {
		fake := newFake()
		record := fake.comments[discussionComment]
		record.Course.Status = domain.Draft
		fake.comments[discussionComment] = record
		response := serve(t, fake, http.MethodPost, "/api/v1/learn/comments/"+discussionComment+"/reports", "", actor(discussionViewer, "student"), true)
		requireError(t, response, http.StatusForbidden, "course_forbidden")
		if len(fake.reports) != 0 {
			t.Fatalf("reports=%+v want none", fake.reports)
		}
	})
	t.Run("Pay19 report on a paid course requires purchase without a write", func(t *testing.T) {
		fake := newFake()
		record := fake.comments[discussionComment]
		record.Course.Paid = true
		fake.comments[discussionComment] = record
		response := serve(t, fake, http.MethodPost, "/api/v1/learn/comments/"+discussionComment+"/reports", "", actor(discussionViewer, "student"), true)
		requireError(t, response, http.StatusForbidden, "course_purchase_required")
		if len(fake.reports) != 0 {
			t.Fatalf("reports=%+v want none", fake.reports)
		}
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
	t.Run("W15 other course owner is forbidden", func(t *testing.T) {
		response := serve(t, newFake(), http.MethodGet, "/api/v1/teach/courses/"+discussionOwner+"/comments", "", actor(discussionOtherOwner, "instructor"), true)
		requireError(t, response, http.StatusForbidden, "course_forbidden")
	})
	t.Run("W20 unavailable teach dependency is named", func(t *testing.T) {
		fake := newFake()
		fake.factsErr = domain.ErrDependencyUnavailable
		response := serve(t, fake, http.MethodGet, "/api/v1/teach/courses/"+discussionOwner+"/comments", "", actor(discussionOwner, "instructor"), true)
		requireError(t, response, http.StatusServiceUnavailable, "dependency_unavailable")
	})
	t.Run("queue", func(t *testing.T) {
		response := serve(t, newFake(), http.MethodGet, "/api/v1/teach/courses/"+discussionOwner+"/moderation-queue", "", actor(discussionAdmin, "admin"), true)
		if response.Code != http.StatusOK {
			t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
		}
		var body struct {
			Items []struct {
				Kind    string `json:"kind"`
				Comment *struct {
					LessonTitle string `json:"lessonTitle"`
				} `json:"comment"`
				Review *struct {
					Stars int `json:"stars"`
				} `json:"review"`
			} `json:"items"`
		}
		if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		if len(body.Items) != 2 || body.Items[0].Kind != "comment" || body.Items[0].Comment == nil || body.Items[0].Comment.LessonTitle != "Lesson one" || body.Items[1].Kind != "review" || body.Items[1].Review == nil || body.Items[1].Review.Stars != 5 {
			t.Fatalf("items=%+v", body.Items)
		}
	})
	t.Run("invalid cursor", func(t *testing.T) {
		response := serve(t, newFake(), http.MethodGet, "/api/v1/teach/courses/"+discussionOwner+"/comments?cursor=%%", "", actor(discussionOwner, "instructor"), true)
		requireError(t, response, http.StatusBadRequest, "comment_cursor_invalid")
	})
	t.Run("W06 dismiss writes a moderation log with its reason", func(t *testing.T) {
		fake := newFake()
		response := serve(t, fake, http.MethodPost, "/api/v1/teach/comments/"+discussionComment+"/dismiss-report", `{"reason":"not a violation"}`, actor(discussionOwner, "instructor"), true)
		if response.Code != http.StatusNoContent {
			t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
		}
		if len(fake.dismissed) != 1 || fake.dismissed[0] != discussionComment {
			t.Fatalf("dismissed=%+v", fake.dismissed)
		}
		if len(fake.moderationLogs) != 1 {
			t.Fatalf("moderation logs=%+v want one", fake.moderationLogs)
		}
		entry := fake.moderationLogs[0]
		if entry.ActorID != discussionOwner || entry.ActorRole != "instructor" || entry.CommentID != discussionComment || entry.Action != "dismiss" || entry.Reason == nil || *entry.Reason != "not a violation" {
			t.Fatalf("moderation log=%+v", entry)
		}
	})
	t.Run("W19 dismiss by another course owner is forbidden without effects", func(t *testing.T) {
		fake := newFake()
		response := serve(t, fake, http.MethodPost, "/api/v1/teach/comments/"+discussionComment+"/dismiss-report", `{"reason":"not mine"}`, actor(discussionOtherOwner, "instructor"), true)
		requireError(t, response, http.StatusForbidden, "course_forbidden")
		if len(fake.dismissed) != 0 || len(fake.moderationLogs) != 0 {
			t.Fatalf("dismissed=%+v logs=%+v want none", fake.dismissed, fake.moderationLogs)
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

func TestT05ReviewRoutesContract(t *testing.T) {
	valid := strings.Repeat("a", 20)
	t.Run("put boundaries eligibility and update", func(t *testing.T) {
		fake := newFake()
		put := func(stars int, body string) *httptest.ResponseRecorder {
			return serve(t, fake, http.MethodPut, "/api/v1/learn/courses/"+discussionOwner+"/review", fmt.Sprintf(`{"stars":%d,"body":"%s"}`, stars, body), actor(discussionAuthor, "student"), true)
		}
		badPut := func(stars int, body string) *httptest.ResponseRecorder {
			return serve(t, fake, http.MethodPut, "/api/v1/learn/courses/"+discussionOwner+"/review", fmt.Sprintf(`{"stars":%d,"body":"%s"}`, stars, body), actor(discussionAuthor, "student"), false)
		}
		requireError(t, badPut(3, strings.Repeat("a", 19)), http.StatusBadRequest, "review_invalid")
		first := put(3, valid)
		if first.Code != http.StatusCreated {
			t.Fatalf("first=%d %s", first.Code, first.Body.String())
		}
		if put(3, strings.Repeat("a", 2000)).Code != http.StatusOK {
			t.Fatal("2000 should update")
		}
		requireError(t, badPut(3, strings.Repeat("a", 2001)), http.StatusBadRequest, "review_invalid")
		requireError(t, badPut(0, valid), http.StatusBadRequest, "review_invalid")
		requireError(t, badPut(6, valid), http.StatusBadRequest, "review_invalid")
		fake.reviewFacts.Facts.OwnerID = discussionOwner
		requireError(t, serve(t, fake, http.MethodPut, "/api/v1/learn/courses/"+discussionOwner+"/review", `{"stars":3,"body":"aaaaaaaaaaaaaaaaaaaa"}`, actor(discussionOwner, "instructor"), true), http.StatusForbidden, "review_owner_forbidden")
		fake.reviewFacts.Facts.OwnerID = ""
		fake.reviewFacts.Facts.HasProgress = false
		requireError(t, put(3, valid), http.StatusForbidden, "review_not_started")
		fake.reviewFacts.Facts.HasProgress = true
		fake.reviewFacts.Facts.Paid = true
		requireError(t, put(3, valid), http.StatusForbidden, "review_purchase_required")
	})
	t.Run("own delete report remove dismiss", func(t *testing.T) {
		fake := newFake()
		now := time.Now()
		fake.reviews[discussionReview] = ports.ReviewRecord{ID: discussionReview, CourseID: discussionOwner, AuthorID: discussionAuthor, Stars: 4, Body: valid, CreatedAt: now}
		fake.reviewFacts.Facts = domain.ReviewFacts{Status: domain.Published, HasProgress: true}
		get := serve(t, fake, http.MethodGet, "/api/v1/learn/courses/"+discussionOwner+"/review", "", actor(discussionAuthor, "student"), true)
		if get.Code != 200 {
			t.Fatal(get.Code)
		}
		requireError(t, serve(t, fake, http.MethodPost, "/api/v1/learn/reviews/"+discussionReview+"/reports", "", actor(discussionAuthor, "student"), true), 403, "report_own_review")
		if serve(t, fake, http.MethodPost, "/api/v1/learn/reviews/"+discussionReview+"/reports", "", actor(discussionViewer, "student"), true).Code != 204 {
			t.Fatal("report")
		}
		requireError(t, serve(t, fake, http.MethodPost, "/api/v1/teach/reviews/"+discussionReview+"/remove", `{"reason":"reason"}`, actor(discussionOwner, "instructor"), true), 403, "review_moderation_forbidden")
		if serve(t, fake, http.MethodPost, "/api/v1/teach/reviews/"+discussionReview+"/remove", `{"reason":"reason"}`, actor(discussionAdmin, "admin"), true).Code != 204 {
			t.Fatal("remove")
		}
		if serve(t, fake, http.MethodPost, "/api/v1/teach/reviews/"+discussionReview+"/remove", `{"reason":"reason"}`, actor(discussionAdmin, "admin"), true).Code != 204 {
			t.Fatal("reremove")
		}
		requireError(t, serve(t, fake, http.MethodDelete, "/api/v1/learn/courses/"+discussionOwner+"/review", "", actor(discussionAuthor, "student"), true), 409, "review_removed")
	})
	t.Run("public anonymous empty rating and visible number", func(t *testing.T) {
		fake := newFake()
		response := serve(t, fake, http.MethodGet, "/api/v1/courses/discussion/reviews", "", nil, true)
		if response.Code != 200 {
			t.Fatal(response.Code)
		}
		fake.reviewCount, fake.reviewSum = 3, 14
		response = serve(t, fake, http.MethodGet, "/api/v1/courses/discussion/reviews", "", nil, true)
		var body struct {
			Rating struct {
				Average float64 `json:"average"`
				Count   int     `json:"count"`
			} `json:"rating"`
		}
		if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil || body.Rating.Average != 4.7 || body.Rating.Count != 3 {
			t.Fatalf("rating=%s err=%v", response.Body.String(), err)
		}
		fake.publicErr = domain.ErrNotFound
		requireError(t, serve(t, fake, http.MethodGet, "/api/v1/courses/archived/reviews", "", nil, true), 404, "course_not_found")
	})
}

// Gaps closed by Courses 1 after the T05 contract test: anonymous writes, the public page cursor, and
// the dismiss route. Cross-origin refusal is not testable here (sameOrigin is a pass-through in this
// harness); the browser suite exercises the real middleware.
func TestT05ReviewRoutesGaps(t *testing.T) {
	valid := strings.Repeat("a", 20)
	t.Run("anonymous review writes are refused", func(t *testing.T) {
		fake := newFake()
		requireError(t, serve(t, fake, http.MethodPut, "/api/v1/learn/courses/"+discussionOwner+"/review", `{"stars":3,"body":"`+valid+`"}`, nil, true), http.StatusUnauthorized, "authentication_required")
		requireError(t, serve(t, fake, http.MethodDelete, "/api/v1/learn/courses/"+discussionOwner+"/review", "", nil, true), http.StatusUnauthorized, "authentication_required")
		requireError(t, serve(t, fake, http.MethodGet, "/api/v1/learn/courses/"+discussionOwner+"/review", "", nil, true), http.StatusUnauthorized, "authentication_required")
	})
	t.Run("public page carries twenty items and the next cursor", func(t *testing.T) {
		fake := newFake()
		items := make([]ports.ReviewRecord, 20)
		for i := range items {
			items[i] = ports.ReviewRecord{ID: fmt.Sprintf("75000000-0000-4000-8000-%012d", i+1), CourseID: discussionOwner, AuthorID: discussionAuthor, Stars: 4, Body: valid, CreatedAt: time.Date(2026, 9, 28, 10, 0, 0, 0, time.UTC).Add(-time.Duration(i) * time.Minute)}
		}
		fake.reviewPage = ports.ReviewPage{Items: items, NextCursor: "next-page"}
		response := serve(t, fake, http.MethodGet, "/api/v1/courses/discussion/reviews", "", nil, true)
		var body struct {
			Items      []json.RawMessage `json:"items"`
			NextCursor *string           `json:"nextCursor"`
		}
		if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil || response.Code != http.StatusOK || len(body.Items) != 20 || body.NextCursor == nil || *body.NextCursor != "next-page" {
			t.Fatalf("status=%d items=%d cursor=%v err=%v body=%s", response.Code, len(body.Items), body.NextCursor, err, response.Body.String())
		}
	})
	t.Run("dismiss review reports: owner allowed, unrelated student refused", func(t *testing.T) {
		fake := newFake()
		fake.reviews[discussionReview] = ports.ReviewRecord{ID: discussionReview, CourseID: discussionOwner, AuthorID: discussionAuthor, Stars: 4, Body: valid, CreatedAt: time.Now()}
		if response := serve(t, fake, http.MethodPost, "/api/v1/teach/reviews/"+discussionReview+"/dismiss-report", `{"reason":"not abusive"}`, actor(discussionOwner, "instructor"), true); response.Code != http.StatusNoContent {
			t.Fatalf("owner dismiss=%d %s", response.Code, response.Body.String())
		}
		response := serve(t, fake, http.MethodPost, "/api/v1/teach/reviews/"+discussionReview+"/dismiss-report", `{"reason":"not abusive"}`, actor(discussionViewer, "student"), true)
		if response.Code != http.StatusForbidden {
			t.Fatalf("student dismiss=%d %s", response.Code, response.Body.String())
		}
	})
}
