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

	coursehttp "courses/internal/course/adapters/http"
	"courses/internal/course/application"
	"courses/internal/course/domain"
	identityhttp "courses/internal/identity/adapters/http"
	identityapplication "courses/internal/identity/application"
	identitydomain "courses/internal/identity/domain"
	"courses/platform/httpserver"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/legacy"
)

const (
	adminUserID      = "40000000-0000-4000-8000-000000000003"
	instructorUserID = "40000000-0000-4000-8000-000000000002"
	otherOwnerID     = "40000000-0000-4000-8000-000000000004"
	adminCourseID    = "10000000-0000-4000-8000-000000000001"
	missingCourseID  = "10000000-0000-4000-8000-000000000099"
)

type adminRepository struct {
	listCalls, loadCalls int
	createCalls          int
	saveCalls            int
	transitionCalls      int
	deleteCalls          int
	lastActor            domain.Actor
	lastQuery            domain.Query
	lastStatus           domain.Status
	lastDraft            domain.CourseDraft
	lastSaved            domain.CourseTree
	lastExpected         time.Time
	ownerID              string
	failure              error
	courses              map[string]domain.CourseTree
	eligibleOwners       map[string]bool
	nextID               int
}

func (repository *adminRepository) List(_ context.Context, actor domain.Actor, query domain.Query, status domain.Status) (domain.AuthoringPage, error) {
	repository.listCalls++
	repository.lastActor = actor
	repository.lastQuery = query
	repository.lastStatus = status
	if repository.failure != nil {
		return domain.AuthoringPage{}, repository.failure
	}
	if repository.courses != nil {
		items := make([]domain.AuthoringCourse, 0, len(repository.courses))
		for _, course := range repository.courses {
			if actor.Role == domain.ActorInstructor && course.InstructorID != actor.UserID || status != "" && course.Status != status {
				continue
			}
			items = append(items, domain.AuthoringCourse{
				ID: course.ID, Slug: course.Slug, Title: course.Title, Status: course.Status,
				InstructorID: course.InstructorID, UpdatedAt: course.UpdatedAt,
			})
		}
		return domain.AuthoringPage{Items: items, Limit: query.Limit, Offset: query.Offset, Total: int64(len(items))}, nil
	}
	items := []domain.AuthoringCourse{adminCourseSummary(repository.owner())}
	if actor.Role == domain.ActorAdmin {
		items = append(items, adminCourseSummary(otherOwnerID))
	}
	return domain.AuthoringPage{Items: items, Limit: query.Limit, Offset: query.Offset, Total: int64(len(items))}, nil
}

func (repository *adminRepository) Load(_ context.Context, id string) (domain.CourseTree, error) {
	repository.loadCalls++
	if repository.failure != nil {
		return domain.CourseTree{}, repository.failure
	}
	if repository.courses != nil {
		tree, exists := repository.courses[id]
		if !exists {
			return domain.CourseTree{}, domain.ErrCourseNotFound
		}
		return tree, nil
	}
	if id == missingCourseID {
		return domain.CourseTree{}, domain.ErrCourseNotFound
	}
	return adminCourseTree(repository.owner()), nil
}

func (repository *adminRepository) Create(_ context.Context, _ domain.Actor, draft domain.CourseDraft, now time.Time) (domain.CourseTree, error) {
	repository.createCalls++
	repository.lastDraft = draft
	if !repository.ownerEligible(draft.InstructorID) {
		return domain.CourseTree{}, domain.ErrInvalidOwner
	}
	for _, course := range repository.courses {
		if course.Slug == draft.Slug {
			return domain.CourseTree{}, domain.ErrSlugTaken
		}
	}
	repository.nextID++
	tree := domain.CourseTree{
		ID:   fmt.Sprintf("10000000-0000-4000-8000-%012d", repository.nextID),
		Slug: draft.Slug, Title: draft.Title, Summary: draft.Summary, Description: draft.Description,
		Status: domain.Draft, InstructorID: draft.InstructorID, CreatedAt: now, UpdatedAt: now,
		Modules: []domain.Module{},
	}
	repository.courses[tree.ID] = tree
	return tree, nil
}
func (repository *adminRepository) Save(_ context.Context, _ domain.Actor, tree domain.CourseTree, expectedUpdatedAt, updatedAt time.Time) (domain.CourseTree, error) {
	repository.saveCalls++
	repository.lastSaved = tree
	repository.lastExpected = expectedUpdatedAt
	stored, exists := repository.courses[tree.ID]
	if !exists {
		return domain.CourseTree{}, domain.ErrCourseNotFound
	}
	if !repository.ownerEligible(tree.InstructorID) {
		return domain.CourseTree{}, domain.ErrInvalidOwner
	}
	if !stored.UpdatedAt.Equal(expectedUpdatedAt) {
		return domain.CourseTree{}, domain.ErrCourseModified
	}
	for id, course := range repository.courses {
		if id != tree.ID && course.Slug == tree.Slug {
			return domain.CourseTree{}, domain.ErrSlugTaken
		}
	}
	stored.Slug = tree.Slug
	stored.Title = tree.Title
	stored.Summary = tree.Summary
	stored.Description = tree.Description
	stored.InstructorID = tree.InstructorID
	stored.Modules = tree.Modules
	stored.UpdatedAt = updatedAt
	repository.courses[tree.ID] = stored
	return stored, nil
}
func (repository *adminRepository) Transition(_ context.Context, actor domain.Actor, id string, from, to domain.Status, expectedUpdatedAt, updatedAt time.Time) (domain.CourseTree, error) {
	repository.transitionCalls++
	repository.lastActor = actor
	repository.lastExpected = expectedUpdatedAt
	stored, exists := repository.courses[id]
	if !exists {
		return domain.CourseTree{}, domain.ErrCourseNotFound
	}
	if err := actor.CanManage(stored.InstructorID); err != nil {
		return domain.CourseTree{}, err
	}
	if stored.Status != from {
		return domain.CourseTree{}, domain.ErrInvalidStatusTransition
	}
	if !stored.UpdatedAt.Equal(expectedUpdatedAt) {
		return domain.CourseTree{}, domain.ErrCourseModified
	}
	stored.Status = to
	if to == domain.Published && stored.PublishedAt == nil {
		publishedAt := updatedAt
		stored.PublishedAt = &publishedAt
	}
	stored.UpdatedAt = updatedAt
	repository.courses[id] = stored
	return stored, nil
}
func (repository *adminRepository) Delete(_ context.Context, actor domain.Actor, id string, expectedUpdatedAt time.Time) error {
	repository.deleteCalls++
	repository.lastActor = actor
	repository.lastExpected = expectedUpdatedAt
	stored, exists := repository.courses[id]
	if !exists {
		return domain.ErrCourseNotFound
	}
	if err := actor.CanManage(stored.InstructorID); err != nil {
		return err
	}
	if stored.Status != domain.Draft {
		return domain.ErrCourseNotDeletable
	}
	if !stored.UpdatedAt.Equal(expectedUpdatedAt) {
		return domain.ErrCourseModified
	}
	delete(repository.courses, id)
	return nil
}

func (repository *adminRepository) owner() string {
	if repository.ownerID != "" {
		return repository.ownerID
	}
	return instructorUserID
}

func (repository *adminRepository) ownerEligible(id string) bool {
	return repository.eligibleOwners == nil || repository.eligibleOwners[id]
}

func (repository *adminRepository) ListPublished(_ context.Context, query domain.Query) (domain.Page, error) {
	items := make([]domain.Course, 0)
	for _, tree := range repository.courses {
		if tree.Status == domain.Published {
			items = append(items, publicCourse(tree))
		}
	}
	return domain.Page{Items: items, Limit: query.Limit, Offset: query.Offset, Total: int64(len(items))}, nil
}

func (repository *adminRepository) GetPublished(_ context.Context, slug string) (domain.CourseDetail, error) {
	for _, tree := range repository.courses {
		if tree.Slug == slug && tree.Status == domain.Published {
			program := make([]domain.ProgramModule, 0, len(tree.Modules))
			for _, module := range tree.Modules {
				lessons := make([]domain.ProgramLesson, 0, len(module.Lessons))
				for _, lesson := range module.Lessons {
					lessons = append(lessons, domain.ProgramLesson{Title: lesson.Title})
				}
				program = append(program, domain.ProgramModule{Title: module.Title, Lessons: lessons})
			}
			return domain.CourseDetail{Course: publicCourse(tree), Program: program}, nil
		}
	}
	return domain.CourseDetail{}, domain.ErrCourseNotFound
}

func publicCourse(tree domain.CourseTree) domain.Course {
	return domain.Course{
		ID: tree.ID, Slug: tree.Slug, Title: tree.Title, Summary: tree.Summary,
		Description: tree.Description, Status: tree.Status, CreatedAt: tree.CreatedAt,
		UpdatedAt: tree.UpdatedAt, PublishedAt: tree.PublishedAt,
	}
}

type adminIdentity struct {
	user                  identitydomain.User
	instructors           []identitydomain.User
	authenticationFailure error
	authenticateCalls     int
	instructorListCalls   int
}

func (*adminIdentity) Login(context.Context, identityapplication.LoginInput) (identityapplication.LoginResult, error) {
	return identityapplication.LoginResult{}, errors.New("unexpected login")
}
func (identity *adminIdentity) Authenticate(context.Context, string) (identitydomain.User, error) {
	identity.authenticateCalls++
	if identity.authenticationFailure != nil {
		return identitydomain.User{}, identity.authenticationFailure
	}
	return identity.user, nil
}
func (*adminIdentity) Logout(context.Context, string) error { return errors.New("unexpected logout") }
func (identity *adminIdentity) ListInstructors(context.Context) ([]identitydomain.User, error) {
	identity.instructorListCalls++
	return identity.instructors, nil
}

func TestAdminReadHTTPMatchesOpenAPI(t *testing.T) {
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

	for _, test := range []struct {
		name, path, code string
		status           int
		role             identitydomain.Role
		ownerID          string
		authenticated    bool
		authFailure      error
		wantCourseCalls  int
		wantPeopleCalls  int
		validRequest     bool
	}{
		{name: "admin lists every course", path: "/api/v1/admin/courses", status: 200, role: identitydomain.Admin, authenticated: true, wantCourseCalls: 1, validRequest: true},
		{name: "admin list search parameters", path: "/api/v1/admin/courses?limit=50&offset=10000&status=archived", status: 200, role: identitydomain.Admin, authenticated: true, wantCourseCalls: 1, validRequest: true},
		{name: "instructor lists own courses", path: "/api/v1/admin/courses", status: 200, role: identitydomain.Instructor, ownerID: instructorUserID, authenticated: true, wantCourseCalls: 1, validRequest: true},
		{name: "admin reads a tree", path: "/api/v1/admin/courses/" + adminCourseID, status: 200, role: identitydomain.Admin, authenticated: true, wantCourseCalls: 1, validRequest: true},
		{name: "owner reads a tree", path: "/api/v1/admin/courses/" + adminCourseID, status: 200, role: identitydomain.Instructor, ownerID: instructorUserID, authenticated: true, wantCourseCalls: 1, validRequest: true},
		{name: "P07 admin lists owner candidates", path: "/api/v1/admin/instructors", status: 200, role: identitydomain.Admin, authenticated: true, wantPeopleCalls: 1, validRequest: true},
		{name: "R01 anonymous list", path: "/api/v1/admin/courses", code: "authentication_required", status: 401, validRequest: true},
		{name: "R01 anonymous tree", path: "/api/v1/admin/courses/" + adminCourseID, code: "authentication_required", status: 401, validRequest: true},
		{name: "R01 anonymous instructors", path: "/api/v1/admin/instructors", code: "authentication_required", status: 401, validRequest: true},
		{name: "R02 student list", path: "/api/v1/admin/courses", code: "forbidden", status: 403, role: identitydomain.Student, authenticated: true, validRequest: true},
		{name: "student role is refused before query validation", path: "/api/v1/admin/courses?status=unknown", code: "forbidden", status: 403, role: identitydomain.Student, authenticated: true},
		{name: "R02 student tree", path: "/api/v1/admin/courses/" + adminCourseID, code: "forbidden", status: 403, role: identitydomain.Student, authenticated: true, validRequest: true},
		{name: "R02 student instructors", path: "/api/v1/admin/instructors", code: "forbidden", status: 403, role: identitydomain.Student, authenticated: true, validRequest: true},
		{name: "R03 disabled account", path: "/api/v1/admin/courses", code: "authentication_required", status: 401, role: identitydomain.Instructor, authenticated: true, authFailure: identitydomain.ErrAuthenticationRequired, validRequest: true},
		{name: "R04 foreign instructor", path: "/api/v1/admin/courses/" + adminCourseID, code: "course_forbidden", status: 403, role: identitydomain.Instructor, ownerID: otherOwnerID, authenticated: true, wantCourseCalls: 1, validRequest: true},
		{name: "R05 instructor owner candidates", path: "/api/v1/admin/instructors", code: "forbidden", status: 403, role: identitydomain.Instructor, authenticated: true, validRequest: true},
		{name: "R06 unknown id", path: "/api/v1/admin/courses/" + missingCourseID, code: "course_not_found", status: 404, role: identitydomain.Admin, authenticated: true, wantCourseCalls: 1, validRequest: true},
		{name: "R07 malformed id", path: "/api/v1/admin/courses/not-a-uuid", code: "invalid_request", status: 400, role: identitydomain.Admin, authenticated: true},
		{name: "invalid list status", path: "/api/v1/admin/courses?status=unknown", code: "invalid_request", status: 400, role: identitydomain.Admin, authenticated: true},
		{name: "unknown list parameter", path: "/api/v1/admin/courses?q=course", code: "invalid_request", status: 400, role: identitydomain.Admin, authenticated: true},
		{name: "repeated list parameter", path: "/api/v1/admin/courses?limit=1&limit=2", code: "invalid_request", status: 400, role: identitydomain.Admin, authenticated: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			repository := &adminRepository{ownerID: test.ownerID}
			identity := &adminIdentity{
				user:                  identityUser(test.role),
				instructors:           []identitydomain.User{identityUser(identitydomain.Admin), identityUser(identitydomain.Instructor)},
				authenticationFailure: test.authFailure,
			}
			handler := adminReadHandler(t, repository, identity)
			request := httptest.NewRequest(http.MethodGet, test.path, nil)
			if test.authenticated {
				request.AddCookie(&http.Cookie{Name: identityhttp.SessionCookieName, Value: strings.Repeat("A", 43)})
			}
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)

			if response.Code != test.status {
				t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
			}
			if response.Header().Get("Cache-Control") != "no-store" {
				t.Fatalf("response is cacheable: %v", response.Header())
			}
			if repository.listCalls+repository.loadCalls != test.wantCourseCalls || identity.instructorListCalls != test.wantPeopleCalls {
				t.Fatalf("calls: list=%d load=%d instructors=%d", repository.listCalls, repository.loadCalls, identity.instructorListCalls)
			}
			if test.code != "" {
				assertAdminError(t, response, test.code)
			}
			if test.name == "P07 admin lists owner candidates" {
				assertInstructorResponse(t, response)
			}
			if test.name == "instructor lists own courses" && repository.lastActor != (domain.Actor{UserID: instructorUserID, Role: domain.ActorInstructor}) {
				t.Fatalf("identity role was not mapped explicitly: %+v", repository.lastActor)
			}
			if test.name == "admin list search parameters" && (repository.lastQuery.Limit != 50 || repository.lastQuery.Offset != 10000 || repository.lastStatus != domain.Archived) {
				t.Fatalf("search parameters were not mapped: query=%+v status=%s", repository.lastQuery, repository.lastStatus)
			}

			route, parameters, err := router.FindRoute(request)
			if err != nil {
				t.Fatal(err)
			}
			input := &openapi3filter.RequestValidationInput{Request: request, PathParams: parameters, Route: route}
			if test.validRequest {
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
		})
	}
}

func TestAdminCreateContract(t *testing.T) {
	t.Run("P02 P03 R29 admin creates one invisible draft and replay is refused", func(t *testing.T) {
		repository := newAdminWriteRepository()
		identity := &adminIdentity{user: identityUser(identitydomain.Admin)}
		handler := adminReadHandler(t, repository, identity)
		body := createCourseBody("new-course", adminUserID)

		response, request := serveAdminJSON(t, handler, identity, identityUser(identitydomain.Admin), http.MethodPost, "/api/v1/admin/courses", body)
		assertAdminHTTPContract(t, request, response, true)
		if response.Code != http.StatusCreated {
			t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
		}
		created := decodeAdminTree(t, response)
		if created.ID == "" || created.Status != "draft" || created.PublishedAt != nil || len(created.Modules) != 0 {
			t.Fatalf("created=%+v", created)
		}

		response, request = serveAdminJSON(t, handler, identity, identityUser(identitydomain.Admin), http.MethodPost, "/api/v1/admin/courses", body)
		assertAdminHTTPContract(t, request, response, true)
		if response.Code != http.StatusConflict {
			t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
		}
		assertAdminError(t, response, "slug_taken")
		if repository.createCalls != 2 || countCoursesWithSlug(repository, "new-course") != 1 {
			t.Fatalf("createCalls=%d courses=%+v", repository.createCalls, repository.courses)
		}

		publicCourseResponse := httptest.NewRecorder()
		handler.ServeHTTP(publicCourseResponse, httptest.NewRequest(http.MethodGet, "/api/v1/courses/new-course", nil))
		if publicCourseResponse.Code != http.StatusNotFound {
			t.Fatalf("draft public status=%d body=%s", publicCourseResponse.Code, publicCourseResponse.Body.String())
		}
		assertAdminError(t, publicCourseResponse, "course_not_found")
		publicListResponse := httptest.NewRecorder()
		handler.ServeHTTP(publicListResponse, httptest.NewRequest(http.MethodGet, "/api/v1/courses", nil))
		var publicList struct {
			Items []json.RawMessage `json:"items"`
		}
		decodeJSON(t, publicListResponse, &publicList)
		if publicListResponse.Code != http.StatusOK || len(publicList.Items) != 0 {
			t.Fatalf("draft leaked into catalog: %s", publicListResponse.Body.String())
		}
	})

	t.Run("P04 instructor owns a create that omits instructorId", func(t *testing.T) {
		repository := newAdminWriteRepository()
		identity := &adminIdentity{}
		handler := adminReadHandler(t, repository, identity)
		response, request := serveAdminJSON(t, handler, identity, identityUser(identitydomain.Instructor), http.MethodPost, "/api/v1/admin/courses", createCourseBody("instructor-course", ""))
		assertAdminHTTPContract(t, request, response, true)
		created := decodeAdminTree(t, response)
		if response.Code != http.StatusCreated || created.InstructorID != instructorUserID || repository.lastDraft.InstructorID != instructorUserID {
			t.Fatalf("status=%d created=%+v draft=%+v", response.Code, created, repository.lastDraft)
		}
	})

	t.Run("P05 admin chooses an owner and the owner lists the course", func(t *testing.T) {
		repository := newAdminWriteRepository()
		identity := &adminIdentity{}
		handler := adminReadHandler(t, repository, identity)
		response, request := serveAdminJSON(t, handler, identity, identityUser(identitydomain.Admin), http.MethodPost, "/api/v1/admin/courses", createCourseBody("owned-course", instructorUserID))
		assertAdminHTTPContract(t, request, response, true)
		if response.Code != http.StatusCreated || decodeAdminTree(t, response).InstructorID != instructorUserID {
			t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
		}

		list := serveAdminRead(handler, identity, identityUser(identitydomain.Instructor), "/api/v1/admin/courses")
		var page struct {
			Items []struct {
				Slug string `json:"slug"`
			} `json:"items"`
		}
		decodeJSON(t, list, &page)
		if list.Code != http.StatusOK || len(page.Items) != 1 || page.Items[0].Slug != "owned-course" {
			t.Fatalf("owner list=%s", list.Body.String())
		}
	})
}

func TestAdminSaveHappyAndBoundaryContracts(t *testing.T) {
	t.Run("P06 admin reassigns the owner and both lists move", func(t *testing.T) {
		repository := seededAdminWriteRepository(domain.Draft, instructorUserID)
		identity := &adminIdentity{}
		handler := adminReadHandler(t, repository, identity)
		tree := repository.courses[adminCourseID]
		body := saveCourseBody(tree, formatTestTime(tree.UpdatedAt))
		body["course"].(map[string]any)["instructorId"] = otherOwnerID

		response, request := serveAdminJSON(t, handler, identity, identityUser(identitydomain.Admin), http.MethodPut, "/api/v1/admin/courses/"+adminCourseID, body)
		assertAdminHTTPContract(t, request, response, true)
		if response.Code != http.StatusOK || decodeAdminTree(t, response).InstructorID != otherOwnerID {
			t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
		}
		oldList := serveAdminRead(handler, identity, identityUserWithID(instructorUserID, identitydomain.Instructor), "/api/v1/admin/courses")
		newList := serveAdminRead(handler, identity, identityUserWithID(otherOwnerID, identitydomain.Instructor), "/api/v1/admin/courses")
		if adminListCount(t, oldList) != 0 || adminListCount(t, newList) != 1 {
			t.Fatalf("old=%s new=%s", oldList.Body.String(), newList.Body.String())
		}
	})

	t.Run("B04 a never-published draft can be renamed", func(t *testing.T) {
		repository := seededAdminWriteRepository(domain.Draft, instructorUserID)
		identity := &adminIdentity{}
		handler := adminReadHandler(t, repository, identity)
		tree := repository.courses[adminCourseID]
		body := saveCourseBody(tree, formatTestTime(tree.UpdatedAt))
		body["course"].(map[string]any)["slug"] = "renamed-course"
		response, request := serveAdminJSON(t, handler, identity, identityUser(identitydomain.Admin), http.MethodPut, "/api/v1/admin/courses/"+adminCourseID, body)
		assertAdminHTTPContract(t, request, response, true)
		if response.Code != http.StatusOK || decodeAdminTree(t, response).Slug != "renamed-course" {
			t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
		}
		published := repository.courses[adminCourseID]
		publishedAt := published.UpdatedAt.Add(time.Microsecond)
		published.Status, published.PublishedAt, published.UpdatedAt = domain.Published, &publishedAt, publishedAt
		repository.courses[adminCourseID] = published
		public := httptest.NewRecorder()
		handler.ServeHTTP(public, httptest.NewRequest(http.MethodGet, "/api/v1/courses/renamed-course", nil))
		if public.Code != http.StatusOK {
			t.Fatalf("renamed public URL status=%d body=%s", public.Code, public.Body.String())
		}
	})

	t.Run("B05 maximal Cyrillic tree stays below the cap and saves", func(t *testing.T) {
		repository := seededAdminWriteRepository(domain.Draft, instructorUserID)
		identity := &adminIdentity{}
		handler := adminReadHandler(t, repository, identity)
		tree := maximalAdminTree(repository.courses[adminCourseID])
		body := saveCourseBody(tree, formatTestTime(tree.UpdatedAt))
		encoded, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		if len(encoded) >= 512*1024 {
			t.Fatalf("maximal body=%d bytes", len(encoded))
		}
		response, request := serveAdminRaw(t, handler, identity, identityUser(identitydomain.Admin), http.MethodPut, "/api/v1/admin/courses/"+adminCourseID, encoded)
		assertAdminHTTPContract(t, request, response, true)
		if response.Code != http.StatusOK || len(repository.courses[adminCourseID].Modules) != 40 || !positionsAreContiguous(repository.lastSaved.Modules) {
			t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
		}
	})

	t.Run("an empty tree saves as the complete replacement", func(t *testing.T) {
		repository := seededAdminWriteRepository(domain.Draft, instructorUserID)
		identity := &adminIdentity{}
		handler := adminReadHandler(t, repository, identity)
		tree := repository.courses[adminCourseID]
		tree.Modules = []domain.Module{}
		response, request := serveAdminJSON(t, handler, identity, identityUser(identitydomain.Admin), http.MethodPut, "/api/v1/admin/courses/"+adminCourseID, saveCourseBody(tree, formatTestTime(tree.UpdatedAt)))
		assertAdminHTTPContract(t, request, response, true)
		if response.Code != http.StatusOK || len(decodeAdminTree(t, response).Modules) != 0 || repository.lastSaved.Modules == nil {
			t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
		}
	})

	// LB06 removes the lesson-description ceiling. Module and total-lesson ceilings
	// still refuse the write without changing the stored tree (B06).
	t.Run("B06 invalid tree ceilings leave the tree unchanged", func(t *testing.T) {
		cases := []struct {
			name   string
			mutate func(*domain.CourseTree)
			code   string
		}{
			{name: "41 modules", code: "invalid_request", mutate: func(tree *domain.CourseTree) {
				tree.Modules = makeModules(41, 0)
			}},
			{name: "121 lessons", code: "lesson_limit", mutate: func(tree *domain.CourseTree) {
				tree.Modules = makeModules(1, 121)
			}},
		}
		for _, test := range cases {
			t.Run(test.name, func(t *testing.T) {
				repository := seededAdminWriteRepository(domain.Draft, instructorUserID)
				before := repository.courses[adminCourseID]
				input := before
				test.mutate(&input)
				identity := &adminIdentity{}
				handler := adminReadHandler(t, repository, identity)
				response, request := serveAdminJSON(t, handler, identity, identityUser(identitydomain.Admin), http.MethodPut, "/api/v1/admin/courses/"+adminCourseID, saveCourseBody(input, formatTestTime(input.UpdatedAt)))
				assertAdminHTTPContract(t, request, response, false)
				if response.Code != http.StatusBadRequest {
					t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
				}
				assertAdminError(t, response, test.code)
				if repository.loadCalls != 0 || repository.saveCalls != 0 || !reflect.DeepEqual(repository.courses[adminCourseID], before) {
					t.Fatalf("invalid payload reached repository or changed tree: load=%d save=%d", repository.loadCalls, repository.saveCalls)
				}
			})
		}
	})

	t.Run("B07 a body over 512 KiB is refused with a bounded response", func(t *testing.T) {
		repository := seededAdminWriteRepository(domain.Draft, instructorUserID)
		identity := &adminIdentity{}
		handler := adminReadHandler(t, repository, identity)
		oversized := append([]byte(`{"expectedUpdatedAt":"2026-09-05T12:00:00.000000Z","course":{}}`), []byte(strings.Repeat(" ", 512*1024))...)
		response, request := serveAdminRaw(t, handler, identity, identityUser(identitydomain.Admin), http.MethodPut, "/api/v1/admin/courses/"+adminCourseID, oversized)
		assertAdminHTTPContract(t, request, response, false)
		if response.Code != http.StatusBadRequest || response.Body.Len() > 1024 || repository.loadCalls != 0 || repository.saveCalls != 0 {
			t.Fatalf("status=%d responseBytes=%d load=%d save=%d", response.Code, response.Body.Len(), repository.loadCalls, repository.saveCalls)
		}
		assertAdminError(t, response, "invalid_request")
	})

	t.Run("B08 summary accepts 500 characters and rejects 501", func(t *testing.T) {
		repository := seededAdminWriteRepository(domain.Draft, instructorUserID)
		identity := &adminIdentity{}
		handler := adminReadHandler(t, repository, identity)
		tree := repository.courses[adminCourseID]
		tree.Summary = strings.Repeat("Ж", 500)
		response, request := serveAdminJSON(t, handler, identity, identityUser(identitydomain.Admin), http.MethodPut, "/api/v1/admin/courses/"+adminCourseID, saveCourseBody(tree, formatTestTime(tree.UpdatedAt)))
		assertAdminHTTPContract(t, request, response, true)
		if response.Code != http.StatusOK {
			t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
		}
		accepted := repository.courses[adminCourseID]
		tooLong := accepted
		tooLong.Summary = strings.Repeat("Ж", 501)
		response, request = serveAdminJSON(t, handler, identity, identityUser(identitydomain.Admin), http.MethodPut, "/api/v1/admin/courses/"+adminCourseID, saveCourseBody(tooLong, formatTestTime(accepted.UpdatedAt)))
		assertAdminHTTPContract(t, request, response, false)
		if response.Code != http.StatusBadRequest || repository.saveCalls != 1 || !reflect.DeepEqual(repository.courses[adminCourseID], accepted) {
			t.Fatalf("status=%d save=%d body=%s", response.Code, repository.saveCalls, response.Body.String())
		}
		assertAdminError(t, response, "invalid_request")
	})

	t.Run("course and node field bounds surface as invalid_request", func(t *testing.T) {
		cases := []struct {
			name   string
			mutate func(*domain.CourseTree)
		}{
			{name: "empty title", mutate: func(tree *domain.CourseTree) { tree.Title = "" }},
			{name: "201 character title", mutate: func(tree *domain.CourseTree) { tree.Title = strings.Repeat("Ж", 201) }},
			{name: "20001 character description", mutate: func(tree *domain.CourseTree) { tree.Description = strings.Repeat("Ж", 20001) }},
			{name: "121 character slug", mutate: func(tree *domain.CourseTree) { tree.Slug = strings.Repeat("a", 121) }},
			{name: "201 character module title", mutate: func(tree *domain.CourseTree) { tree.Modules[0].Title = strings.Repeat("Ж", 201) }},
			{name: "201 character lesson title", mutate: func(tree *domain.CourseTree) { tree.Modules[0].Lessons[0].Title = strings.Repeat("Ж", 201) }},
		}
		for _, test := range cases {
			t.Run(test.name, func(t *testing.T) {
				repository := seededAdminWriteRepository(domain.Draft, instructorUserID)
				input := cloneCourseTree(repository.courses[adminCourseID])
				test.mutate(&input)
				identity := &adminIdentity{}
				handler := adminReadHandler(t, repository, identity)
				response, request := serveAdminJSON(t, handler, identity, identityUser(identitydomain.Admin), http.MethodPut, "/api/v1/admin/courses/"+adminCourseID, saveCourseBody(input, formatTestTime(input.UpdatedAt)))
				assertAdminHTTPContract(t, request, response, false)
				if response.Code != http.StatusBadRequest || repository.loadCalls != 0 || repository.saveCalls != 0 {
					t.Fatalf("status=%d load=%d save=%d body=%s", response.Code, repository.loadCalls, repository.saveCalls, response.Body.String())
				}
				assertAdminError(t, response, "invalid_request")
			})
		}
	})

	// B-3: the ceiling is a sum over sibling arrays, and JSON Schema cannot bound a sum. Both
	// halves of that are asserted here rather than described: a course that splits 121 lessons
	// across two modules is contract-valid and domain-invalid at once, while the same 121 in one
	// module never reaches the server because the per-module maxItems does catch it. That pair is
	// the whole of what the schema can and cannot own. If a later edition finds a way to express
	// the sum, the first case fails at request validation and names which half moved.
	t.Run("B-3 the course-wide ceiling is refused by name where the schema cannot see it", func(t *testing.T) {
		overCeiling := func(t *testing.T, counts ...int) (*httptest.ResponseRecorder, *http.Request, *adminRepository) {
			t.Helper()
			repository := seededAdminWriteRepository(domain.Draft, instructorUserID)
			input := cloneCourseTree(repository.courses[adminCourseID])
			input.Modules = nil
			serial := 0
			for modulePosition, count := range counts {
				module := domain.Module{ID: treeNodeID(20000 + modulePosition), Title: "Module", Position: modulePosition}
				for lessonPosition := range count {
					module.Lessons = append(module.Lessons, domain.Lesson{ID: treeNodeID(30000 + serial), Title: "Lesson", Position: lessonPosition})
					serial++
				}
				input.Modules = append(input.Modules, module)
			}
			identity := &adminIdentity{}
			handler := adminReadHandler(t, repository, identity)
			response, request := serveAdminJSON(t, handler, identity, identityUser(identitydomain.Admin), http.MethodPut, "/api/v1/admin/courses/"+adminCourseID, saveCourseBody(input, formatTestTime(input.UpdatedAt)))
			return response, request, repository
		}

		t.Run("121 across two modules passes the schema and is refused as lesson_limit", func(t *testing.T) {
			response, request, repository := overCeiling(t, 60, 61)
			assertAdminHTTPContract(t, request, response, true)
			if response.Code != http.StatusBadRequest || repository.saveCalls != 0 {
				t.Fatalf("status=%d save=%d body=%s", response.Code, repository.saveCalls, response.Body.String())
			}
			assertAdminError(t, response, "lesson_limit")
		})

		t.Run("121 in one module is stopped by the per-module maxItems", func(t *testing.T) {
			response, request, repository := overCeiling(t, domain.MaximumLessons+1)
			assertAdminRequestSchema(t, request, false)
			assertAdminHTTPContract(t, request, response, false)
			if response.Code != http.StatusBadRequest || repository.saveCalls != 0 {
				t.Fatalf("status=%d save=%d body=%s", response.Code, repository.saveCalls, response.Body.String())
			}
			assertAdminError(t, response, "lesson_limit")
		})
	})

	t.Run("B12 read token round-trips and a lesson-only save retires it", func(t *testing.T) {
		repository := seededAdminWriteRepository(domain.Draft, instructorUserID)
		identity := &adminIdentity{}
		handler := adminReadHandler(t, repository, identity)
		read := serveAdminRead(handler, identity, identityUser(identitydomain.Admin), "/api/v1/admin/courses/"+adminCourseID)
		loaded := decodeAdminTree(t, read)
		if _, err := time.Parse("2006-01-02T15:04:05.000000Z07:00", loaded.UpdatedAt); err != nil {
			t.Fatalf("updatedAt is not RFC3339 microsecond precision: %q", loaded.UpdatedAt)
		}
		tree := cloneCourseTree(repository.courses[adminCourseID])
		tree.Modules[0].Lessons[0].Title = "Changed lesson only"
		response, request := serveAdminJSON(t, handler, identity, identityUser(identitydomain.Admin), http.MethodPut, "/api/v1/admin/courses/"+adminCourseID, saveCourseBody(tree, loaded.UpdatedAt))
		assertAdminHTTPContract(t, request, response, true)
		saved := decodeAdminTree(t, response)
		if response.Code != http.StatusOK || saved.UpdatedAt == loaded.UpdatedAt || !repository.lastExpected.Equal(tree.UpdatedAt) {
			t.Fatalf("loaded=%s saved=%s expected=%s", loaded.UpdatedAt, saved.UpdatedAt, repository.lastExpected)
		}
	})
}

func TestAdminSaveRefusalContracts(t *testing.T) {
	t.Run("R12 instructor cannot reassign ownership", func(t *testing.T) {
		repository := seededAdminWriteRepository(domain.Draft, instructorUserID)
		before := repository.courses[adminCourseID]
		identity := &adminIdentity{}
		handler := adminReadHandler(t, repository, identity)
		body := saveCourseBody(before, formatTestTime(before.UpdatedAt))
		body["course"].(map[string]any)["instructorId"] = otherOwnerID
		response, request := serveAdminJSON(t, handler, identity, identityUser(identitydomain.Instructor), http.MethodPut, "/api/v1/admin/courses/"+adminCourseID, body)
		assertAdminHTTPContract(t, request, response, true)
		if response.Code != http.StatusForbidden || !reflect.DeepEqual(repository.courses[adminCourseID], before) {
			t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
		}
		assertAdminError(t, response, "course_forbidden")
	})

	t.Run("R13 ineligible owners are refused without changing owner", func(t *testing.T) {
		for _, test := range []struct {
			name, ownerID string
		}{
			{name: "student", ownerID: "40000000-0000-4000-8000-000000000001"},
			{name: "disabled instructor", ownerID: "40000000-0000-4000-8000-000000000005"},
			{name: "unknown user", ownerID: missingCourseID},
		} {
			t.Run(test.name, func(t *testing.T) {
				repository := seededAdminWriteRepository(domain.Draft, instructorUserID)
				before := repository.courses[adminCourseID]
				identity := &adminIdentity{}
				handler := adminReadHandler(t, repository, identity)
				body := saveCourseBody(before, formatTestTime(before.UpdatedAt))
				body["course"].(map[string]any)["instructorId"] = test.ownerID
				response, request := serveAdminJSON(t, handler, identity, identityUser(identitydomain.Admin), http.MethodPut, "/api/v1/admin/courses/"+adminCourseID, body)
				assertAdminHTTPContract(t, request, response, true)
				if response.Code != http.StatusBadRequest || repository.courses[adminCourseID].InstructorID != before.InstructorID {
					t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
				}
				assertAdminError(t, response, "invalid_owner")
			})
		}
	})

	t.Run("R20 published and archived slugs are immutable", func(t *testing.T) {
		for _, status := range []domain.Status{domain.Published, domain.Archived} {
			t.Run(string(status), func(t *testing.T) {
				repository := seededAdminWriteRepository(status, instructorUserID)
				before := repository.courses[adminCourseID]
				identity := &adminIdentity{}
				handler := adminReadHandler(t, repository, identity)
				body := saveCourseBody(before, formatTestTime(before.UpdatedAt))
				body["course"].(map[string]any)["slug"] = "forbidden-rename"
				response, request := serveAdminJSON(t, handler, identity, identityUser(identitydomain.Admin), http.MethodPut, "/api/v1/admin/courses/"+adminCourseID, body)
				assertAdminHTTPContract(t, request, response, true)
				if response.Code != http.StatusConflict || repository.courses[adminCourseID].Slug != before.Slug {
					t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
				}
				assertAdminError(t, response, "slug_immutable")
				if status == domain.Archived {
					restored := repository.courses[adminCourseID]
					restored.Status = domain.Published
					repository.courses[adminCourseID] = restored
				}
				public := httptest.NewRecorder()
				handler.ServeHTTP(public, httptest.NewRequest(http.MethodGet, "/api/v1/courses/"+before.Slug, nil))
				if public.Code != http.StatusOK {
					t.Fatalf("original public URL status=%d", public.Code)
				}
			})
		}
	})

	t.Run("R21 rename onto a taken slug is refused", func(t *testing.T) {
		repository := seededAdminWriteRepository(domain.Draft, instructorUserID)
		other := adminCourseTree(otherOwnerID)
		other.ID, other.Slug = missingCourseID, "taken-slug"
		repository.courses[other.ID] = other
		before := repository.courses[adminCourseID]
		identity := &adminIdentity{}
		handler := adminReadHandler(t, repository, identity)
		body := saveCourseBody(before, formatTestTime(before.UpdatedAt))
		body["course"].(map[string]any)["slug"] = other.Slug
		response, request := serveAdminJSON(t, handler, identity, identityUser(identitydomain.Admin), http.MethodPut, "/api/v1/admin/courses/"+adminCourseID, body)
		assertAdminHTTPContract(t, request, response, true)
		if response.Code != http.StatusConflict || repository.courses[adminCourseID].Slug != before.Slug {
			t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
		}
		assertAdminError(t, response, "slug_taken")
	})

	t.Run("R22 lifecycle and position members are rejected before repository access", func(t *testing.T) {
		cases := []struct {
			name   string
			mutate func(map[string]any)
		}{
			{name: "status", mutate: func(body map[string]any) { body["course"].(map[string]any)["status"] = "published" }},
			{name: "publishedAt", mutate: func(body map[string]any) { body["course"].(map[string]any)["publishedAt"] = nil }},
			{name: "position", mutate: func(body map[string]any) { body["modules"].([]map[string]any)[0]["position"] = 0 }},
		}
		for _, test := range cases {
			t.Run(test.name, func(t *testing.T) {
				repository := seededAdminWriteRepository(domain.Draft, instructorUserID)
				identity := &adminIdentity{}
				handler := adminReadHandler(t, repository, identity)
				body := saveCourseBody(repository.courses[adminCourseID], formatTestTime(repository.courses[adminCourseID].UpdatedAt))
				test.mutate(body)
				response, request := serveAdminJSON(t, handler, identity, identityUser(identitydomain.Admin), http.MethodPut, "/api/v1/admin/courses/"+adminCourseID, body)
				assertAdminHTTPContract(t, request, response, false)
				if response.Code != http.StatusBadRequest || repository.loadCalls != 0 || repository.saveCalls != 0 {
					t.Fatalf("status=%d load=%d save=%d body=%s", response.Code, repository.loadCalls, repository.saveCalls, response.Body.String())
				}
				assertAdminError(t, response, "invalid_request")
			})
		}
	})
}

func TestAdminLifecycleHappyAndBoundaryContracts(t *testing.T) {
	t.Run("P11 P12 P13 publish archive and restore update public visibility", func(t *testing.T) {
		repository := seededAdminWriteRepository(domain.Draft, instructorUserID)
		identity := &adminIdentity{}
		handler := adminReadHandler(t, repository, identity)

		draft := repository.courses[adminCourseID]
		response, request := serveAdminJSON(t, handler, identity, identityUser(identitydomain.Admin), http.MethodPost, lifecyclePath(), statusBody(domain.Published, draft.UpdatedAt))
		assertAdminHTTPContract(t, request, response, true)
		published := decodeAdminTree(t, response)
		if response.Code != http.StatusOK || published.Status != string(domain.Published) || published.PublishedAt == nil {
			t.Fatalf("publish status=%d body=%s", response.Code, response.Body.String())
		}
		firstPublishedAt := *published.PublishedAt
		assertPublicCourseStatus(t, handler, draft.Slug, http.StatusOK)

		response, request = serveAdminJSON(t, handler, identity, identityUser(identitydomain.Admin), http.MethodPost, lifecyclePath(), statusBody(domain.Archived, repository.courses[adminCourseID].UpdatedAt))
		assertAdminHTTPContract(t, request, response, true)
		archived := decodeAdminTree(t, response)
		if response.Code != http.StatusOK || archived.Status != string(domain.Archived) {
			t.Fatalf("archive status=%d body=%s", response.Code, response.Body.String())
		}
		assertPublicCourseStatus(t, handler, draft.Slug, http.StatusNotFound)
		publicList := httptest.NewRecorder()
		handler.ServeHTTP(publicList, httptest.NewRequest(http.MethodGet, "/api/v1/courses", nil))
		if publicList.Code != http.StatusOK || adminListCount(t, publicList) != 0 {
			t.Fatalf("archived course remained in catalog: %s", publicList.Body.String())
		}

		response, request = serveAdminJSON(t, handler, identity, identityUser(identitydomain.Admin), http.MethodPost, lifecyclePath(), statusBody(domain.Published, repository.courses[adminCourseID].UpdatedAt))
		assertAdminHTTPContract(t, request, response, true)
		restored := decodeAdminTree(t, response)
		if response.Code != http.StatusOK || restored.Status != string(domain.Published) || restored.PublishedAt == nil || *restored.PublishedAt != firstPublishedAt {
			t.Fatalf("restore status=%d body=%s", response.Code, response.Body.String())
		}
		assertPublicCourseStatus(t, handler, draft.Slug, http.StatusOK)
	})

	t.Run("P14 instructor publishes their own course", func(t *testing.T) {
		repository := seededAdminWriteRepository(domain.Draft, instructorUserID)
		identity := &adminIdentity{}
		handler := adminReadHandler(t, repository, identity)
		response, request := serveAdminJSON(t, handler, identity, identityUser(identitydomain.Instructor), http.MethodPost, lifecyclePath(), statusBody(domain.Published, repository.courses[adminCourseID].UpdatedAt))
		assertAdminHTTPContract(t, request, response, true)
		if response.Code != http.StatusOK || decodeAdminTree(t, response).Status != string(domain.Published) || repository.lastActor != (domain.Actor{UserID: instructorUserID, Role: domain.ActorInstructor}) {
			t.Fatalf("status=%d actor=%+v body=%s", response.Code, repository.lastActor, response.Body.String())
		}
	})

	t.Run("B01 an empty course may be published", func(t *testing.T) {
		repository := seededAdminWriteRepository(domain.Draft, instructorUserID)
		tree := repository.courses[adminCourseID]
		tree.Modules = []domain.Module{}
		repository.courses[adminCourseID] = tree
		identity := &adminIdentity{}
		handler := adminReadHandler(t, repository, identity)
		response, request := serveAdminJSON(t, handler, identity, identityUser(identitydomain.Admin), http.MethodPost, lifecyclePath(), statusBody(domain.Published, tree.UpdatedAt))
		assertAdminHTTPContract(t, request, response, true)
		if response.Code != http.StatusOK || decodeAdminTree(t, response).Status != string(domain.Published) {
			t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
		}
	})

	t.Run("a draft course can be deleted", func(t *testing.T) {
		repository := seededAdminWriteRepository(domain.Draft, instructorUserID)
		identity := &adminIdentity{}
		handler := adminReadHandler(t, repository, identity)
		response, request := serveAdminJSON(t, handler, identity, identityUser(identitydomain.Admin), http.MethodDelete, coursePath(), expectedUpdatedAtBody(repository.courses[adminCourseID].UpdatedAt))
		assertAdminHTTPContract(t, request, response, true)
		if response.Code != http.StatusNoContent || response.Body.Len() != 0 || repository.deleteCalls != 1 {
			t.Fatalf("status=%d deleteCalls=%d body=%s", response.Code, repository.deleteCalls, response.Body.String())
		}
		if _, exists := repository.courses[adminCourseID]; exists {
			t.Fatal("deleted course remains in repository")
		}
	})
}

func TestAdminMutationRefusalMatrix(t *testing.T) {
	t.Run("R08 anonymous is refused on every mutation", func(t *testing.T) {
		for _, mutation := range adminMutationCases(seededAdminWriteRepository(domain.Draft, instructorUserID)) {
			t.Run(mutation.name, func(t *testing.T) {
				repository := seededAdminWriteRepository(domain.Draft, instructorUserID)
				mutation := mutation.forRepository(repository)
				identity := &adminIdentity{}
				handler := adminReadHandler(t, repository, identity)
				response, request := serveAdminJSONAs(t, handler, identity, identitydomain.User{}, false, "http://localhost:8080", mutation.method, mutation.path, mutation.body)
				assertAdminHTTPContract(t, request, response, true)
				if response.Code != http.StatusUnauthorized || repository.writeCalls() != 0 {
					t.Fatalf("status=%d writes=%d body=%s", response.Code, repository.writeCalls(), response.Body.String())
				}
				assertAdminError(t, response, "authentication_required")
			})
		}
	})

	t.Run("R09 a student is refused on every mutation", func(t *testing.T) {
		for _, mutation := range adminMutationCases(seededAdminWriteRepository(domain.Draft, instructorUserID)) {
			t.Run(mutation.name, func(t *testing.T) {
				repository := seededAdminWriteRepository(domain.Draft, instructorUserID)
				mutation := mutation.forRepository(repository)
				identity := &adminIdentity{}
				handler := adminReadHandler(t, repository, identity)
				response, request := serveAdminJSONAs(t, handler, identity, identityUser(identitydomain.Student), true, "http://localhost:8080", mutation.method, mutation.path, mutation.body)
				assertAdminHTTPContract(t, request, response, true)
				if response.Code != http.StatusForbidden || repository.writeCalls() != 0 {
					t.Fatalf("status=%d writes=%d body=%s", response.Code, repository.writeCalls(), response.Body.String())
				}
				assertAdminError(t, response, "forbidden")
			})
		}
	})

	t.Run("R10 cross-site mutation is refused before authentication", func(t *testing.T) {
		for _, mutation := range adminMutationCases(seededAdminWriteRepository(domain.Draft, instructorUserID)) {
			t.Run(mutation.name, func(t *testing.T) {
				repository := seededAdminWriteRepository(domain.Draft, instructorUserID)
				before := cloneCourseTree(repository.courses[adminCourseID])
				mutation := mutation.forRepository(repository)
				identity := &adminIdentity{}
				handler := adminReadHandler(t, repository, identity)
				response, request := serveAdminJSONAs(t, handler, identity, identityUser(identitydomain.Student), true, "https://foreign.example", mutation.method, mutation.path, mutation.body)
				assertAdminHTTPContract(t, request, response, true)
				if response.Code != http.StatusForbidden || identity.authenticateCalls != 0 || repository.writeCalls() != 0 || !reflect.DeepEqual(repository.courses[adminCourseID], before) {
					t.Fatalf("status=%d auth=%d writes=%d body=%s", response.Code, identity.authenticateCalls, repository.writeCalls(), response.Body.String())
				}
				assertAdminError(t, response, "origin_forbidden")
			})
		}
	})

	t.Run("R11 a foreign instructor cannot save transition or delete", func(t *testing.T) {
		for _, mutation := range adminMutationCases(seededAdminWriteRepository(domain.Draft, instructorUserID))[1:] {
			t.Run(mutation.name, func(t *testing.T) {
				repository := seededAdminWriteRepository(domain.Draft, instructorUserID)
				before := cloneCourseTree(repository.courses[adminCourseID])
				mutation := mutation.forRepository(repository)
				identity := &adminIdentity{}
				handler := adminReadHandler(t, repository, identity)
				response, request := serveAdminJSONAs(t, handler, identity, identityUserWithID(otherOwnerID, identitydomain.Instructor), true, "http://localhost:8080", mutation.method, mutation.path, mutation.body)
				assertAdminHTTPContract(t, request, response, true)
				if response.Code != http.StatusForbidden || !reflect.DeepEqual(repository.courses[adminCourseID], before) {
					t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
				}
				assertAdminError(t, response, "course_forbidden")

				read := serveAdminRead(handler, identity, identityUser(identitydomain.Admin), coursePath())
				readTree := decodeAdminTree(t, read)
				if read.Code != http.StatusOK || readTree.Title != before.Title || readTree.Status != string(before.Status) || readTree.UpdatedAt != formatTestTime(before.UpdatedAt) {
					t.Fatalf("following read changed: status=%d body=%s", read.Code, read.Body.String())
				}
			})
		}
	})
}

func TestAdminLifecycleRefusalContracts(t *testing.T) {
	t.Run("R14 an illegal transition is a named conflict and leaves status unchanged", func(t *testing.T) {
		repository := seededAdminWriteRepository(domain.Published, instructorUserID)
		before := cloneCourseTree(repository.courses[adminCourseID])
		identity := &adminIdentity{}
		handler := adminReadHandler(t, repository, identity)
		response, request := serveAdminJSON(t, handler, identity, identityUser(identitydomain.Admin), http.MethodPost, lifecyclePath(), statusBody(domain.Draft, before.UpdatedAt))
		assertAdminHTTPContract(t, request, response, true)
		if response.Code != http.StatusConflict || !reflect.DeepEqual(repository.courses[adminCourseID], before) {
			t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
		}
		assertAdminError(t, response, "invalid_status_transition")
	})

	t.Run("R15 published and archived courses cannot be deleted", func(t *testing.T) {
		for _, status := range []domain.Status{domain.Published, domain.Archived} {
			t.Run(string(status), func(t *testing.T) {
				repository := seededAdminWriteRepository(status, instructorUserID)
				before := cloneCourseTree(repository.courses[adminCourseID])
				identity := &adminIdentity{}
				handler := adminReadHandler(t, repository, identity)
				response, request := serveAdminJSON(t, handler, identity, identityUser(identitydomain.Admin), http.MethodDelete, coursePath(), expectedUpdatedAtBody(before.UpdatedAt))
				assertAdminHTTPContract(t, request, response, true)
				if response.Code != http.StatusConflict || !reflect.DeepEqual(repository.courses[adminCourseID], before) {
					t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
				}
				assertAdminError(t, response, "course_not_deletable")
			})
		}
	})

	t.Run("R23 a transition retires the previous tree token", func(t *testing.T) {
		repository := seededAdminWriteRepository(domain.Draft, instructorUserID)
		identity := &adminIdentity{}
		handler := adminReadHandler(t, repository, identity)
		before := cloneCourseTree(repository.courses[adminCourseID])
		response, request := serveAdminJSON(t, handler, identity, identityUser(identitydomain.Admin), http.MethodPost, lifecyclePath(), statusBody(domain.Published, before.UpdatedAt))
		assertAdminHTTPContract(t, request, response, true)
		if response.Code != http.StatusOK {
			t.Fatalf("publish status=%d body=%s", response.Code, response.Body.String())
		}
		published := cloneCourseTree(repository.courses[adminCourseID])
		published.Title = "Stale edit"
		response, request = serveAdminJSON(t, handler, identity, identityUser(identitydomain.Admin), http.MethodPut, coursePath(), saveCourseBody(published, formatTestTime(before.UpdatedAt)))
		assertAdminHTTPContract(t, request, response, true)
		if response.Code != http.StatusConflict || repository.courses[adminCourseID].Title == "Stale edit" {
			t.Fatalf("stale save status=%d body=%s", response.Code, response.Body.String())
		}
		assertAdminError(t, response, "course_modified")
	})
}

func TestAdminWriteStrictJSON(t *testing.T) {
	repository := seededAdminWriteRepository(domain.Draft, instructorUserID)
	identity := &adminIdentity{}
	handler := adminReadHandler(t, repository, identity)
	for _, test := range []struct {
		name, body string
	}{
		{name: "malformed", body: `{"slug":`},
		{name: "duplicate", body: `{"slug":"one","slug":"two","title":"Title","summary":"Summary","description":"Description","instructorId":"40000000-0000-4000-8000-000000000003"}`},
		{name: "unknown", body: `{"slug":"one","title":"Title","summary":"Summary","description":"Description","instructorId":"40000000-0000-4000-8000-000000000003","status":"draft"}`},
		{name: "trailing", body: `{"slug":"one","title":"Title","summary":"Summary","description":"Description","instructorId":"40000000-0000-4000-8000-000000000003"} {}`},
	} {
		t.Run(test.name, func(t *testing.T) {
			response, request := serveAdminRaw(t, handler, identity, identityUser(identitydomain.Admin), http.MethodPost, "/api/v1/admin/courses", []byte(test.body))
			assertAdminHTTPContract(t, request, response, false)
			if response.Code != http.StatusBadRequest {
				t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
			}
			assertAdminError(t, response, "invalid_request")
		})
	}
	if repository.createCalls != 0 {
		t.Fatalf("invalid JSON reached repository %d times", repository.createCalls)
	}

	t.Run("expectedUpdatedAt requires six fractional digits", func(t *testing.T) {
		body := saveCourseBody(repository.courses[adminCourseID], "2026-09-05T12:00:00Z")
		response, request := serveAdminJSON(t, handler, identity, identityUser(identitydomain.Admin), http.MethodPut, "/api/v1/admin/courses/"+adminCourseID, body)
		assertAdminHTTPContract(t, request, response, false)
		if response.Code != http.StatusBadRequest || repository.loadCalls != 0 || repository.saveCalls != 0 {
			t.Fatalf("status=%d load=%d save=%d body=%s", response.Code, repository.loadCalls, repository.saveCalls, response.Body.String())
		}
		assertAdminError(t, response, "invalid_request")
	})

	t.Run("query parameters are rejected", func(t *testing.T) {
		response, request := serveAdminJSON(t, handler, identity, identityUser(identitydomain.Admin), http.MethodPost, "/api/v1/admin/courses?status=draft", createCourseBody("query-course", adminUserID))
		assertAdminHTTPContract(t, request, response, false)
		if response.Code != http.StatusBadRequest || repository.createCalls != 0 {
			t.Fatalf("status=%d create=%d body=%s", response.Code, repository.createCalls, response.Body.String())
		}
		assertAdminError(t, response, "invalid_request")
	})

	t.Run("non-JSON content type is rejected", func(t *testing.T) {
		identity.user = identityUser(identitydomain.Admin)
		body, err := json.Marshal(createCourseBody("text-course", adminUserID))
		if err != nil {
			t.Fatal(err)
		}
		request := httptest.NewRequest(http.MethodPost, "/api/v1/admin/courses", bytes.NewReader(body))
		request.Header.Set("Content-Type", "text/plain")
		request.Header.Set("Origin", "http://localhost:8080")
		request.AddCookie(&http.Cookie{Name: identityhttp.SessionCookieName, Value: strings.Repeat("A", 43)})
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != http.StatusBadRequest || repository.createCalls != 0 {
			t.Fatalf("status=%d create=%d body=%s", response.Code, repository.createCalls, response.Body.String())
		}
		assertAdminError(t, response, "invalid_request")
	})
}

type adminTreeResponse struct {
	ID           string  `json:"id"`
	Slug         string  `json:"slug"`
	Title        string  `json:"title"`
	Status       string  `json:"status"`
	InstructorID string  `json:"instructorId"`
	UpdatedAt    string  `json:"updatedAt"`
	PublishedAt  *string `json:"publishedAt"`
	Modules      []struct {
		ID      string `json:"id"`
		Title   string `json:"title"`
		Lessons []struct {
			ID    string `json:"id"`
			Title string `json:"title"`
		} `json:"lessons"`
	} `json:"modules"`
}

type adminMutationCase struct {
	name   string
	method string
	path   string
	body   any
}

func (mutation adminMutationCase) forRepository(repository *adminRepository) adminMutationCase {
	for _, current := range adminMutationCases(repository) {
		if current.name == mutation.name {
			return current
		}
	}
	panic("unknown admin mutation case")
}

func adminMutationCases(repository *adminRepository) []adminMutationCase {
	tree := repository.courses[adminCourseID]
	saveBody := saveCourseBody(tree, formatTestTime(tree.UpdatedAt))
	saveBody["course"].(map[string]any)["title"] = "Matrix update"
	return []adminMutationCase{
		{name: "create", method: http.MethodPost, path: "/api/v1/admin/courses", body: createCourseBody("matrix-course", adminUserID)},
		{name: "save", method: http.MethodPut, path: coursePath(), body: saveBody},
		{name: "status", method: http.MethodPost, path: lifecyclePath(), body: statusBody(domain.Published, tree.UpdatedAt)},
		{name: "delete", method: http.MethodDelete, path: coursePath(), body: expectedUpdatedAtBody(tree.UpdatedAt)},
	}
}

func coursePath() string {
	return "/api/v1/admin/courses/" + adminCourseID
}

func lifecyclePath() string {
	return coursePath() + "/status"
}

func statusBody(status domain.Status, expectedUpdatedAt time.Time) map[string]any {
	return map[string]any{"status": string(status), "expectedUpdatedAt": formatTestTime(expectedUpdatedAt)}
}

func expectedUpdatedAtBody(expectedUpdatedAt time.Time) map[string]any {
	return map[string]any{"expectedUpdatedAt": formatTestTime(expectedUpdatedAt)}
}

func (repository *adminRepository) writeCalls() int {
	return repository.createCalls + repository.saveCalls + repository.transitionCalls + repository.deleteCalls
}

func assertPublicCourseStatus(t *testing.T, handler http.Handler, slug string, status int) {
	t.Helper()
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/courses/"+slug, nil))
	if response.Code != status {
		t.Fatalf("public course status=%d body=%s", response.Code, response.Body.String())
	}
	if status == http.StatusNotFound {
		assertAdminError(t, response, "course_not_found")
	}
}

func newAdminWriteRepository() *adminRepository {
	return &adminRepository{
		courses: map[string]domain.CourseTree{},
		eligibleOwners: map[string]bool{
			adminUserID: true, instructorUserID: true, otherOwnerID: true,
		},
	}
}

func seededAdminWriteRepository(status domain.Status, ownerID string) *adminRepository {
	repository := newAdminWriteRepository()
	tree := adminCourseTree(ownerID)
	tree.Status = status
	if status != domain.Draft {
		publishedAt := tree.UpdatedAt
		tree.PublishedAt = &publishedAt
	}
	repository.courses[tree.ID] = tree
	return repository
}

func createCourseBody(slug, instructorID string) map[string]any {
	body := map[string]any{
		"slug": slug, "title": "New course", "summary": "A concise summary", "description": "A full description",
	}
	if instructorID != "" {
		body["instructorId"] = instructorID
	}
	return body
}

func saveCourseBody(tree domain.CourseTree, expectedUpdatedAt string) map[string]any {
	modules := make([]map[string]any, 0, len(tree.Modules))
	for _, module := range tree.Modules {
		lessons := make([]map[string]any, 0, len(module.Lessons))
		for _, lesson := range module.Lessons {
			lessons = append(lessons, map[string]any{
				"id": lesson.ID, "title": lesson.Title,
			})
		}
		modules = append(modules, map[string]any{"id": module.ID, "title": module.Title, "lessons": lessons})
	}
	return map[string]any{
		"expectedUpdatedAt": expectedUpdatedAt,
		"course": map[string]any{
			"slug": tree.Slug, "title": tree.Title, "summary": tree.Summary,
			"description": tree.Description, "instructorId": tree.InstructorID,
		},
		"modules": modules,
	}
}

func maximalAdminTree(tree domain.CourseTree) domain.CourseTree {
	tree.Title = strings.Repeat("Ж", 200)
	tree.Summary = strings.Repeat("Ж", 500)
	tree.Description = strings.Repeat("Ж", 20000)
	tree.Modules = makeModules(40, 3)
	return tree
}

func cloneCourseTree(tree domain.CourseTree) domain.CourseTree {
	clone := tree
	clone.Modules = make([]domain.Module, len(tree.Modules))
	for index, module := range tree.Modules {
		clone.Modules[index] = module
		clone.Modules[index].Lessons = append([]domain.Lesson(nil), module.Lessons...)
	}
	return clone
}

func positionsAreContiguous(modules []domain.Module) bool {
	for modulePosition, module := range modules {
		if module.Position != modulePosition {
			return false
		}
		for lessonPosition, lesson := range module.Lessons {
			if lesson.Position != lessonPosition {
				return false
			}
		}
	}
	return true
}

func makeModules(moduleCount, lessonsPerModule int) []domain.Module {
	modules := make([]domain.Module, 0, moduleCount)
	lessonNumber := 0
	for modulePosition := 0; modulePosition < moduleCount; modulePosition++ {
		lessons := make([]domain.Lesson, 0, lessonsPerModule)
		for lessonPosition := 0; lessonPosition < lessonsPerModule; lessonPosition++ {
			lessonNumber++
			lessons = append(lessons, domain.Lesson{
				ID:    fmt.Sprintf("30000000-0000-4000-8000-%012d", lessonNumber),
				Title: strings.Repeat("Ж", 200), Position: lessonPosition,
			})
		}
		modules = append(modules, domain.Module{
			ID:    fmt.Sprintf("20000000-0000-4000-8000-%012d", modulePosition+1),
			Title: strings.Repeat("Ж", 200), Position: modulePosition, Lessons: lessons,
		})
	}
	return modules
}

func serveAdminJSON(
	t *testing.T,
	handler http.Handler,
	identity *adminIdentity,
	user identitydomain.User,
	method, path string,
	body any,
) (*httptest.ResponseRecorder, *http.Request) {
	t.Helper()
	encoded, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	return serveAdminRaw(t, handler, identity, user, method, path, encoded)
}

func serveAdminJSONAs(
	t *testing.T,
	handler http.Handler,
	identity *adminIdentity,
	user identitydomain.User,
	authenticated bool,
	origin, method, path string,
	body any,
) (*httptest.ResponseRecorder, *http.Request) {
	t.Helper()
	encoded, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	return serveAdminRawAs(t, handler, identity, user, authenticated, origin, method, path, encoded)
}

func serveAdminRaw(
	t *testing.T,
	handler http.Handler,
	identity *adminIdentity,
	user identitydomain.User,
	method, path string,
	body []byte,
) (*httptest.ResponseRecorder, *http.Request) {
	t.Helper()
	return serveAdminRawAs(t, handler, identity, user, true, "http://localhost:8080", method, path, body)
}

func serveAdminRawAs(
	t *testing.T,
	handler http.Handler,
	identity *adminIdentity,
	user identitydomain.User,
	authenticated bool,
	origin, method, path string,
	body []byte,
) (*httptest.ResponseRecorder, *http.Request) {
	t.Helper()
	identity.user = user
	request := httptest.NewRequest(method, path, bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
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
	return response, request
}

func serveAdminRead(handler http.Handler, identity *adminIdentity, user identitydomain.User, path string) *httptest.ResponseRecorder {
	identity.user = user
	request := httptest.NewRequest(http.MethodGet, path, nil)
	request.AddCookie(&http.Cookie{Name: identityhttp.SessionCookieName, Value: strings.Repeat("A", 43)})
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}

func decodeAdminTree(t *testing.T, response *httptest.ResponseRecorder) adminTreeResponse {
	t.Helper()
	var tree adminTreeResponse
	decodeJSON(t, response, &tree)
	return tree
}

func decodeJSON(t *testing.T, response *httptest.ResponseRecorder, value any) {
	t.Helper()
	if err := json.Unmarshal(response.Body.Bytes(), value); err != nil {
		t.Fatalf("decode response: %v body=%s", err, response.Body.String())
	}
}

func adminListCount(t *testing.T, response *httptest.ResponseRecorder) int {
	t.Helper()
	var page struct {
		Items []json.RawMessage `json:"items"`
	}
	decodeJSON(t, response, &page)
	return len(page.Items)
}

func countCoursesWithSlug(repository *adminRepository, slug string) int {
	count := 0
	for _, course := range repository.courses {
		if course.Slug == slug {
			count++
		}
	}
	return count
}

func formatTestTime(value time.Time) string {
	return value.UTC().Truncate(time.Microsecond).Format("2006-01-02T15:04:05.000000Z07:00")
}

func assertAdminHTTPContract(t *testing.T, request *http.Request, response *httptest.ResponseRecorder, validRequest bool) {
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

func TestAdminReadWrongMethod(t *testing.T) {
	repository := &adminRepository{}
	identity := &adminIdentity{user: identityUser(identitydomain.Admin)}
	response := httptest.NewRecorder()
	adminReadHandler(t, repository, identity).ServeHTTP(response, httptest.NewRequest(http.MethodPatch, "/api/v1/admin/courses", nil))
	if response.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	assertAdminError(t, response, "method_not_allowed")
	if repository.listCalls+repository.loadCalls != 0 {
		t.Fatal("wrong method reached the repository")
	}
}

func adminReadHandler(t *testing.T, repository *adminRepository, identity *adminIdentity) http.Handler {
	t.Helper()
	config, err := identityhttp.NewConfig("test", "http://localhost:8080", false)
	if err != nil {
		t.Fatal(err)
	}
	authoring := application.NewAuthoring(repository, func() time.Time { return time.Date(2026, 9, 5, 12, 0, 0, 0, time.UTC) })
	courseHandler := coursehttp.NewHandler(application.NewCatalog(repository), authoring)
	identityHandler := identityhttp.NewHandler(identity, identity, nil, config)
	return httpserver.NewHandler(httpserver.Options{Service: "api", Register: func(mux *http.ServeMux) {
		identityHandler.Register(mux)
		courseHandler.Register(mux)
		courseHandler.RegisterAdmin(mux, identityHandler.Authenticate, identityHandler.RequireSameOrigin)
	}})
}

func identityUser(role identitydomain.Role) identitydomain.User {
	id := adminUserID
	displayName := "Адміністратор Курсу"
	email := "admin@courseforge.local"
	if role == identitydomain.Instructor {
		id, displayName, email = instructorUserID, "Викладач Курсу", "instructor@courseforge.local"
	}
	if role == identitydomain.Student {
		id, displayName, email = "40000000-0000-4000-8000-000000000001", "Студентка Курсу", "student@courseforge.local"
	}
	now := time.Date(2026, 9, 5, 12, 0, 0, 0, time.UTC)
	return identitydomain.User{ID: id, Email: email, DisplayName: displayName, Role: role, CreatedAt: now, UpdatedAt: now}
}

func identityUserWithID(id string, role identitydomain.Role) identitydomain.User {
	user := identityUser(role)
	user.ID = id
	return user
}

func adminCourseSummary(ownerID string) domain.AuthoringCourse {
	return domain.AuthoringCourse{
		ID: adminCourseID, Slug: "go-foundations", Title: "Go foundations", Status: domain.Draft,
		InstructorID: ownerID, UpdatedAt: time.Date(2026, 9, 5, 12, 0, 0, 0, time.UTC),
	}
}

func adminCourseTree(ownerID string) domain.CourseTree {
	now := time.Date(2026, 9, 5, 12, 0, 0, 0, time.UTC)
	return domain.CourseTree{
		ID: adminCourseID, Slug: "go-foundations", Title: "Go foundations", Summary: "Build a reliable API",
		Description: "Learn modules, PostgreSQL and tests.", Status: domain.Draft, InstructorID: ownerID,
		CreatedAt: now, UpdatedAt: now,
		Modules: []domain.Module{{
			ID: "20000000-0000-4000-8000-000000000001", Title: "First module", Position: 0,
			Lessons: []domain.Lesson{{ID: "30000000-0000-4000-8000-000000000001", Title: "First lesson", Position: 0}},
		}},
	}
}

// assertAdminRequestSchema states whether openapi.yaml itself accepts this request body, which
// is the half of backlog B-3 that no server behaviour can show.
func assertAdminRequestSchema(t *testing.T, request *http.Request, wantValid bool) {
	t.Helper()
	document, err := openapi3.NewLoader().LoadFromFile(filepath.Join("..", "..", "..", "..", "openapi", "openapi.yaml"))
	if err != nil {
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
	err = openapi3filter.ValidateRequest(context.Background(), &openapi3filter.RequestValidationInput{Request: request, PathParams: parameters, Route: route})
	if wantValid && err != nil {
		t.Fatalf("the contract refuses a body it should accept: %v", err)
	}
	if !wantValid && err == nil {
		t.Fatal("the contract accepts a body it should refuse")
	}
}

func treeNodeID(number int) string {
	return fmt.Sprintf("30000000-0000-4000-8000-%012d", number)
}

func assertAdminError(t *testing.T, response *httptest.ResponseRecorder, code string) {
	t.Helper()
	var envelope httpserver.ErrorResponse
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope.Error.Code != code || envelope.Error.RequestID == "" || envelope.Error.RequestID != response.Header().Get("X-Request-ID") {
		t.Fatalf("bad envelope: %+v", envelope)
	}
}

func assertInstructorResponse(t *testing.T, response *httptest.ResponseRecorder) {
	t.Helper()
	var body []struct {
		ID          string `json:"id"`
		DisplayName string `json:"displayName"`
		Role        string `json:"role"`
	}
	var members []map[string]json.RawMessage
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(response.Body.Bytes(), &members); err != nil {
		t.Fatal(err)
	}
	if len(body) != 2 || len(members) != 2 {
		t.Fatalf("instructors=%s", response.Body.String())
	}
	want := []struct {
		id, displayName, role string
	}{
		{adminUserID, "Адміністратор Курсу", "admin"},
		{instructorUserID, "Викладач Курсу", "instructor"},
	}
	for index, instructor := range members {
		if len(instructor) != 3 || instructor["id"] == nil || instructor["displayName"] == nil || instructor["role"] == nil || instructor["email"] != nil {
			t.Fatalf("owner candidate leaks fields: %s", response.Body.String())
		}
		if body[index].ID != want[index].id || body[index].DisplayName != want[index].displayName || body[index].Role != want[index].role {
			t.Fatalf("owner candidates=%s", response.Body.String())
		}
	}
}
