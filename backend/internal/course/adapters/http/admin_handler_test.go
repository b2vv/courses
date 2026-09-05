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
func (*adminRepository) Transition(context.Context, domain.Actor, string, domain.Status, domain.Status, time.Time, time.Time) (domain.CourseTree, error) {
	return domain.CourseTree{}, errors.New("unexpected transition")
}
func (*adminRepository) Delete(context.Context, domain.Actor, string, time.Time) error {
	return errors.New("unexpected delete")
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

func (repository *adminRepository) GetPublished(_ context.Context, slug string) (domain.Course, error) {
	for _, tree := range repository.courses {
		if tree.Slug == slug && tree.Status == domain.Published {
			return publicCourse(tree), nil
		}
	}
	return domain.Course{}, domain.ErrCourseNotFound
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

	t.Run("B06 invalid tree ceilings leave the tree unchanged", func(t *testing.T) {
		cases := []struct {
			name   string
			mutate func(*domain.CourseTree)
		}{
			{name: "41 modules", mutate: func(tree *domain.CourseTree) {
				tree.Modules = makeModules(41, 0, 0)
			}},
			{name: "121 lessons", mutate: func(tree *domain.CourseTree) {
				tree.Modules = makeModules(1, 121, 0)
			}},
			{name: "1001 character lesson description", mutate: func(tree *domain.CourseTree) {
				tree.Modules = makeModules(1, 1, 1001)
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
				assertAdminError(t, response, "invalid_request")
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
	Status       string  `json:"status"`
	InstructorID string  `json:"instructorId"`
	UpdatedAt    string  `json:"updatedAt"`
	PublishedAt  *string `json:"publishedAt"`
	Modules      []struct {
		ID      string `json:"id"`
		Title   string `json:"title"`
		Lessons []struct {
			ID          string `json:"id"`
			Title       string `json:"title"`
			Description string `json:"description"`
		} `json:"lessons"`
	} `json:"modules"`
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
				"id": lesson.ID, "title": lesson.Title, "description": lesson.Description,
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
	tree.Modules = makeModules(40, 3, 1000)
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

func makeModules(moduleCount, lessonsPerModule, descriptionLength int) []domain.Module {
	modules := make([]domain.Module, 0, moduleCount)
	lessonNumber := 0
	for modulePosition := 0; modulePosition < moduleCount; modulePosition++ {
		lessons := make([]domain.Lesson, 0, lessonsPerModule)
		for lessonPosition := 0; lessonPosition < lessonsPerModule; lessonPosition++ {
			lessonNumber++
			description := ""
			if descriptionLength > 0 {
				description = strings.Repeat("Ж", descriptionLength)
			}
			lessons = append(lessons, domain.Lesson{
				ID:    fmt.Sprintf("30000000-0000-4000-8000-%012d", lessonNumber),
				Title: strings.Repeat("Ж", 200), Description: description, Position: lessonPosition,
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

func serveAdminRaw(
	t *testing.T,
	handler http.Handler,
	identity *adminIdentity,
	user identitydomain.User,
	method, path string,
	body []byte,
) (*httptest.ResponseRecorder, *http.Request) {
	t.Helper()
	identity.user = user
	request := httptest.NewRequest(method, path, bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Origin", "http://localhost:8080")
	request.AddCookie(&http.Cookie{Name: identityhttp.SessionCookieName, Value: strings.Repeat("A", 43)})
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
	identityHandler := identityhttp.NewHandler(identity, identity, config)
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
			Lessons: []domain.Lesson{{ID: "30000000-0000-4000-8000-000000000001", Title: "First lesson", Description: "Lesson text", Position: 0}},
		}},
	}
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
