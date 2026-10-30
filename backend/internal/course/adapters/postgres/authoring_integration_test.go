//go:build integration

package postgres_test

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	coursepostgres "courses/internal/course/adapters/postgres"
	"courses/internal/course/domain"
	passwords "courses/internal/identity/adapters/argon2id"
	identitycrypto "courses/internal/identity/adapters/identitycrypto"
	identitypostgres "courses/internal/identity/adapters/postgres"
	identitydomain "courses/internal/identity/domain"
	containers "courses/internal/testsupport/containers"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	postgrescontainer "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	authoringStudentID         = "40000000-0000-4000-8000-000000000001"
	authoringInstructorID      = "40000000-0000-4000-8000-000000000002"
	authoringAdminID           = "40000000-0000-4000-8000-000000000003"
	authoringOtherInstructorID = "40000000-0000-4000-8000-000000000004"
	authoringDisabledOwnerID   = "40000000-0000-4000-8000-000000000005"

	authoringModuleOne = "60000000-0000-4000-8000-000000000001"
	authoringModuleTwo = "60000000-0000-4000-8000-000000000002"
	authoringModuleTri = "60000000-0000-4000-8000-000000000003"
	authoringLessonOne = "70000000-0000-4000-8000-000000000001"
	authoringLessonTwo = "70000000-0000-4000-8000-000000000002"
	authoringLessonTri = "70000000-0000-4000-8000-000000000003"
)

type authoringHarness struct {
	ctx        context.Context
	pool       *pgxpool.Pool
	repository *coursepostgres.Repository
}

func TestAuthoringRepositoryAgainstPostgreSQL(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	container, err := containers.Start(ctx, func(ctx context.Context) (*postgrescontainer.PostgresContainer, error) {
		return postgrescontainer.Run(ctx, "postgres:17.5",
			postgrescontainer.WithDatabase("authoring_test"),
			postgrescontainer.WithUsername("courses"),
			postgrescontainer.WithPassword("ephemeral-test-only"),
			testcontainers.WithWaitStrategy(wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(3*time.Minute)),
		)
	})
	if err != nil {
		t.Fatalf("PostgreSQL integration requires Docker: %v", err)
	}
	t.Cleanup(func() {
		if err := container.Terminate(context.Background()); err != nil {
			t.Errorf("terminate test container: %v", err)
		}
	})
	connection, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	database, err := sql.Open("pgx", connection)
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	migrations := filepath.Join("..", "..", "..", "..", "migrations")
	provider, err := goose.NewProvider(goose.DialectPostgres, database, os.DirFS(migrations), goose.WithDisableGlobalRegistry(true))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := provider.Up(ctx); err != nil {
		t.Fatal(err)
	}
	pool, err := pgxpool.New(ctx, connection)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	harness := &authoringHarness{ctx: ctx, pool: pool, repository: coursepostgres.NewRepository(pool)}

	t.Run("P08 tree save upserts and deletes missing nodes", harness.testTreeSaveAndDeleteMissing)
	t.Run("P09 module ordering persists", harness.testModuleOrdering)
	t.Run("P10 lesson moves across modules without changing identity", harness.testLessonMove)
	t.Run("B03 concurrent creates map the slug conflict", harness.testConcurrentSlug)
	t.Run("B10 pagination and owner scoping", harness.testPaginationAndScope)
	t.Run("Cv05 list enrichment: lessons, students, price, status counts, search", harness.testListEnrichment)
	t.Run("B11 transition timestamps round trip to the microsecond", harness.testTimestampRoundTrip)
	t.Run("R16 delete cannot race a publish", harness.testDeletePublishRace)
	t.Run("R17 deleting a draft cascades its tree", harness.testDeleteCascade)
	t.Run("R18 stale save rolls back instead of merging", harness.testOptimisticConflict)
	t.Run("R19 foreign module and lesson ids roll back", harness.testForeignNodeIDs)
	t.Run("R26 cancellation rolls back a partial tree write", harness.testCancelledSave)
	t.Run("owner eligibility is checked in create and save transactions", harness.testOwnerEligibility)
	t.Run("R23 blocking an owner freezes their course", harness.testBlockedOwnerFreeze)
}

func (h *authoringHarness) testBlockedOwnerFreeze(t *testing.T) {
	h.reset(t)
	stored := h.createCourse(t, "blocked-owner-freeze", authoringInstructorID)
	identityKeys, err := identitycrypto.New([]byte("0123456789abcdef0123456789abcdef"))
	if err != nil {
		t.Fatal(err)
	}
	identities := identitypostgres.NewRepository(h.pool, passwords.NewVerifier(), identityKeys)
	if err := identities.SetStatus(h.ctx, authoringInstructorID, identitydomain.AccountBlocked, time.Now().UTC().Truncate(time.Microsecond), func(identitydomain.User, int) error {
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	edit := stored
	edit.Title = "Blocked owner edit"
	if _, err := h.repository.Save(h.ctx, adminActor(), edit, stored.UpdatedAt, stored.UpdatedAt.Add(time.Second)); !errors.Is(err, domain.ErrInvalidOwner) {
		t.Fatalf("save with blocked owner: %v", err)
	}
	after, err := h.repository.Load(h.ctx, stored.ID)
	if err != nil {
		t.Fatal(err)
	}
	if after.Title != stored.Title || !after.UpdatedAt.Equal(stored.UpdatedAt) {
		t.Fatalf("blocked-owner save changed course: before=%+v after=%+v", stored, after)
	}
}

func (h *authoringHarness) testTreeSaveAndDeleteMissing(t *testing.T) {
	h.reset(t)
	tree := h.createCourse(t, "tree-save", authoringInstructorID)
	tree.Modules = fullTree()
	firstToken := tree.UpdatedAt.Add(time.Second)
	saved := h.save(t, tree, firstToken)
	assertTreeContent(t, saved, fullTree())
	if !saved.UpdatedAt.Equal(firstToken) || saved.UpdatedAt.Equal(tree.UpdatedAt) {
		t.Fatalf("child write did not advance token: before=%s after=%s", tree.UpdatedAt, saved.UpdatedAt)
	}

	trimmed := saved
	trimmed.Modules = []domain.Module{{
		ID: authoringModuleOne, Title: "Module one retained", Position: 0,
		Lessons: []domain.Lesson{{ID: authoringLessonOne, Title: "Lesson one retained", Position: 0}},
	}}
	clockReading := trimmed.UpdatedAt.Add(789 * time.Nanosecond)
	saved = h.save(t, trimmed, clockReading)
	assertTreeContent(t, saved, trimmed.Modules)
	if want := trimmed.UpdatedAt.Add(time.Microsecond); !saved.UpdatedAt.Equal(want) {
		t.Fatalf("equal truncated clock did not advance token: got=%s want=%s", saved.UpdatedAt, want)
	}
	var modules, lessons int
	if err := h.pool.QueryRow(h.ctx, `SELECT count(*) FROM course_modules WHERE course_id=$1`, tree.ID).Scan(&modules); err != nil {
		t.Fatal(err)
	}
	if err := h.pool.QueryRow(h.ctx, `SELECT count(*) FROM course_lessons`).Scan(&lessons); err != nil {
		t.Fatal(err)
	}
	if modules != 1 || lessons != 1 {
		t.Fatalf("delete-missing left modules=%d lessons=%d", modules, lessons)
	}
}

func (h *authoringHarness) testModuleOrdering(t *testing.T) {
	h.reset(t)
	tree := h.createCourse(t, "module-order", authoringInstructorID)
	tree.Modules = []domain.Module{
		{ID: authoringModuleOne, Title: "First", Position: 0, Lessons: []domain.Lesson{}},
		{ID: authoringModuleTwo, Title: "Second", Position: 1, Lessons: []domain.Lesson{}},
		{ID: authoringModuleTri, Title: "Third", Position: 2, Lessons: []domain.Lesson{}},
	}
	saved := h.save(t, tree, tree.UpdatedAt.Add(time.Second))
	saved.Modules = []domain.Module{saved.Modules[2], saved.Modules[1], saved.Modules[0]}
	for position := range saved.Modules {
		saved.Modules[position].Position = position
	}
	saved = h.save(t, saved, saved.UpdatedAt.Add(time.Second))
	for index, want := range []string{authoringModuleTri, authoringModuleTwo, authoringModuleOne} {
		if saved.Modules[index].ID != want || saved.Modules[index].Position != index {
			t.Fatalf("position %d: %+v", index, saved.Modules[index])
		}
	}
}

func (h *authoringHarness) testLessonMove(t *testing.T) {
	h.reset(t)
	tree := h.createCourse(t, "lesson-move", authoringInstructorID)
	tree.Modules = []domain.Module{
		{ID: authoringModuleOne, Title: "Origin", Position: 0, Lessons: []domain.Lesson{{ID: authoringLessonOne, Title: "Moving lesson", Position: 0}}},
		{ID: authoringModuleTwo, Title: "Destination", Position: 1, Lessons: []domain.Lesson{}},
	}
	saved := h.save(t, tree, tree.UpdatedAt.Add(time.Second))
	saved.Modules[0].Lessons = []domain.Lesson{}
	saved.Modules[1].Lessons = []domain.Lesson{{ID: authoringLessonOne, Title: "Moving lesson", Position: 0}}
	saved = h.save(t, saved, saved.UpdatedAt.Add(time.Second))
	if len(saved.Modules[0].Lessons) != 0 || len(saved.Modules[1].Lessons) != 1 || saved.Modules[1].Lessons[0].ID != authoringLessonOne {
		t.Fatalf("lesson did not move intact: %+v", saved.Modules)
	}
	var moduleID string
	if err := h.pool.QueryRow(h.ctx, `SELECT module_id::text FROM course_lessons WHERE id=$1`, authoringLessonOne).Scan(&moduleID); err != nil {
		t.Fatal(err)
	}
	if moduleID != authoringModuleTwo {
		t.Fatalf("stored parent=%s", moduleID)
	}
}

func (h *authoringHarness) testConcurrentSlug(t *testing.T) {
	h.reset(t)
	start := make(chan struct{})
	type result struct {
		tree domain.CourseTree
		err  error
	}
	results := make(chan result, 2)
	for range 2 {
		go func() {
			<-start
			tree, err := h.repository.Create(h.ctx, adminActor(), draft("concurrent-slug", authoringInstructorID), authoringTime())
			results <- result{tree: tree, err: err}
		}()
	}
	close(start)
	successes, conflicts := 0, 0
	for range 2 {
		result := <-results
		switch {
		case result.err == nil && result.tree.ID != "":
			successes++
		case errors.Is(result.err, domain.ErrSlugTaken):
			conflicts++
		default:
			t.Fatalf("unexpected concurrent result: tree=%+v err=%v", result.tree, result.err)
		}
	}
	var count int
	if err := h.pool.QueryRow(h.ctx, `SELECT count(*) FROM courses WHERE slug='concurrent-slug'`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if successes != 1 || conflicts != 1 || count != 1 {
		t.Fatalf("successes=%d conflicts=%d rows=%d", successes, conflicts, count)
	}
}

func (h *authoringHarness) testPaginationAndScope(t *testing.T) {
	h.reset(t)
	_, err := h.pool.Exec(h.ctx, `
INSERT INTO courses (id, instructor_id, slug, title, summary, description, status, created_at, updated_at)
SELECT gen_random_uuid(), $1, 'owned-' || n, 'Owned ' || n, 'Summary', 'Description', 'draft',
	   $2::timestamptz, $2::timestamptz + n * interval '1 microsecond'
	FROM generate_series(1, 51) AS n`, authoringInstructorID, authoringTime())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := h.pool.Exec(h.ctx, `INSERT INTO courses (id, instructor_id, slug, title, summary, description, status, created_at, updated_at)
	VALUES (gen_random_uuid(), $1, 'foreign-row', 'Foreign', 'Summary', 'Description', 'draft', $2, $2)`, authoringOtherInstructorID, authoringTime()); err != nil {
		t.Fatal(err)
	}
	query, err := domain.NewQuery(50, 50, "", string(domain.SortRecent))
	if err != nil {
		t.Fatal(err)
	}
	page, err := h.repository.List(h.ctx, instructorActor(), query, domain.Draft)
	if err != nil {
		t.Fatal(err)
	}
	if page.Total != 51 || len(page.Items) != 1 || page.Items[0].InstructorID != authoringInstructorID {
		t.Fatalf("instructor page=%+v", page)
	}
	adminPage, err := h.repository.List(h.ctx, adminActor(), query, domain.Draft)
	if err != nil {
		t.Fatal(err)
	}
	if adminPage.Total != 52 || len(adminPage.Items) != 2 {
		t.Fatalf("admin page=%+v", adminPage)
	}
}

func (h *authoringHarness) testListEnrichment(t *testing.T) {
	h.reset(t)
	if _, err := h.pool.Exec(h.ctx, `
INSERT INTO users (id,email,display_name,role,password_hash)
SELECT ('43000000-0000-4000-8000-'||lpad(n::text,12,'0'))::uuid,
  'enrich-'||n||'@courseforge.local','Enrich '||n,
  CASE WHEN n=1 THEN 'instructor' ELSE 'student' END,
  '$argon2id$v=19$m=65536,t=3,p=2$AAAAAAAAAAAAAAAAAAAAAA$AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA'
FROM generate_series(1,3) AS n;
INSERT INTO courses (id,instructor_id,slug,title,summary,description,status,price_minor,currency,created_at,updated_at,published_at) VALUES
  ('53000000-0000-4000-8000-000000000001','43000000-0000-4000-8000-000000000001','enrich-paid','PostgreSQL deep dive','Advanced queries','Body','draft',149900,'UAH',now(),now(),NULL),
  ('53000000-0000-4000-8000-000000000002','43000000-0000-4000-8000-000000000001','enrich-free','React basics','Components','Body','published',NULL,NULL,now(),now(),now()),
  ('53000000-0000-4000-8000-000000000003','43000000-0000-4000-8000-000000000001','enrich-arch','Legacy topic','Retired','Body','archived',NULL,NULL,now(),now(),NULL);
INSERT INTO course_modules (id,course_id,title,position) VALUES
  ('63000000-0000-4000-8000-000000000001','53000000-0000-4000-8000-000000000001','Module one',0);
INSERT INTO course_lessons (id,module_id,title,position) VALUES
  ('73000000-0000-4000-8000-000000000001','63000000-0000-4000-8000-000000000001','Lesson one',0),
  ('73000000-0000-4000-8000-000000000002','63000000-0000-4000-8000-000000000001','Lesson two',1),
  ('73000000-0000-4000-8000-000000000003','63000000-0000-4000-8000-000000000001','Lesson three',2);
INSERT INTO course_grants (user_id,course_id,source,state) VALUES
  ('43000000-0000-4000-8000-000000000002','53000000-0000-4000-8000-000000000001','admin','active'),
  ('43000000-0000-4000-8000-000000000003','53000000-0000-4000-8000-000000000001','admin','active'),
  ('43000000-0000-4000-8000-000000000001','53000000-0000-4000-8000-000000000001','admin','revoked');`); err != nil {
		t.Fatal(err)
	}

	query, err := domain.NewQuery(50, 0, "", string(domain.SortRecent))
	if err != nil {
		t.Fatal(err)
	}
	page, err := h.repository.List(h.ctx, adminActor(), query, domain.Status(""))
	if err != nil {
		t.Fatal(err)
	}
	bySlug := make(map[string]domain.AuthoringCourse, len(page.Items))
	for _, course := range page.Items {
		bySlug[course.Slug] = course
	}
	paid, free := bySlug["enrich-paid"], bySlug["enrich-free"]
	if paid.LessonCount != 3 || paid.StudentCount != 2 {
		t.Fatalf("paid counts: lessons=%d students=%d (want 3/2, active grants only)", paid.LessonCount, paid.StudentCount)
	}
	if paid.Price == nil || paid.Price.Minor != 149900 || paid.Price.Currency != "UAH" {
		t.Fatalf("paid price=%+v", paid.Price)
	}
	if free.LessonCount != 0 || free.StudentCount != 0 || free.Price != nil {
		t.Fatalf("free row should be empty/unpriced: lessons=%d students=%d price=%+v", free.LessonCount, free.StudentCount, free.Price)
	}
	if page.StatusCounts != (domain.AuthoringStatusCounts{Total: 3, Draft: 1, Published: 1, Archived: 1}) {
		t.Fatalf("status counts=%+v (want 3/1/1/1)", page.StatusCounts)
	}

	// The editor read (GetAuthoringCourse) carries the same active-grant count as the list row,
	// counting active grants only — the one revoked grant on enrich-paid must not inflate it.
	paidTree, err := h.repository.Load(h.ctx, paid.ID)
	if err != nil {
		t.Fatal(err)
	}
	if paidTree.StudentCount != 2 {
		t.Fatalf("editor tree students=%d (want 2, active grants only)", paidTree.StudentCount)
	}
	freeTree, err := h.repository.Load(h.ctx, free.ID)
	if err != nil {
		t.Fatal(err)
	}
	if freeTree.StudentCount != 0 {
		t.Fatalf("free editor tree students=%d (want 0)", freeTree.StudentCount)
	}

	searched, err := domain.NewQuery(50, 0, "postgresql", string(domain.SortRecent))
	if err != nil {
		t.Fatal(err)
	}
	filtered, err := h.repository.List(h.ctx, adminActor(), searched, domain.Status(""))
	if err != nil {
		t.Fatal(err)
	}
	if len(filtered.Items) != 1 || filtered.Items[0].Slug != "enrich-paid" {
		t.Fatalf("search did not isolate the matching title: %+v", filtered.Items)
	}
	// The status snapshot ignores the search term so the summary strip keeps the full picture.
	if filtered.StatusCounts.Total != 3 {
		t.Fatalf("search leaked into status snapshot: %+v", filtered.StatusCounts)
	}

	// Each admin sort orders the full page by its natural direction.
	for _, tc := range []struct {
		sort  domain.Sort
		first string // slug expected at the top
		last  string // slug expected at the bottom (NULLS/ties last)
	}{
		{domain.SortStudents, "enrich-paid", ""},         // 2 active grants first; the two zeros trail
		{domain.SortPrice, "enrich-paid", ""},            // priced first, the two free (NULL) last
		{domain.SortTitle, "enrich-arch", "enrich-free"}, // Legacy < PostgreSQL < React under C collation
	} {
		sorted, err := domain.NewAuthoringQuery(50, 0, "", string(tc.sort))
		if err != nil {
			t.Fatal(err)
		}
		result, err := h.repository.List(h.ctx, adminActor(), sorted, domain.Status(""))
		if err != nil {
			t.Fatal(err)
		}
		if len(result.Items) != 3 || result.Items[0].Slug != tc.first {
			t.Fatalf("sort %q: first=%+v (want %s)", tc.sort, result.Items, tc.first)
		}
		if tc.last != "" && result.Items[len(result.Items)-1].Slug != tc.last {
			t.Fatalf("sort %q: last=%s (want %s)", tc.sort, result.Items[len(result.Items)-1].Slug, tc.last)
		}
	}
}

func (h *authoringHarness) testTimestampRoundTrip(t *testing.T) {
	h.reset(t)
	tree := h.createCourse(t, "timestamp-round-trip", authoringInstructorID)
	clockReading := tree.UpdatedAt.Add(789 * time.Nanosecond)
	publishedAt := tree.UpdatedAt.Add(time.Microsecond)
	transitioned, err := h.repository.Transition(h.ctx, adminActor(), tree.ID, domain.Draft, domain.Published, tree.UpdatedAt, clockReading)
	if err != nil {
		t.Fatal(err)
	}
	if transitioned.PublishedAt == nil || !transitioned.PublishedAt.Equal(publishedAt) || !transitioned.UpdatedAt.Equal(publishedAt) {
		t.Fatalf("transition timestamps=%+v", transitioned)
	}
	if transitioned.UpdatedAt.Nanosecond()%int(time.Microsecond) != 0 {
		t.Fatalf("sub-microsecond timestamp=%s", transitioned.UpdatedAt)
	}
	if transitioned.UpdatedAt.Equal(clockReading) {
		t.Fatalf("sub-microsecond clock was not truncated: %s", clockReading)
	}
	var storedPublished, storedUpdated time.Time
	if err := h.pool.QueryRow(h.ctx, `SELECT published_at, updated_at FROM courses WHERE id=$1`, tree.ID).Scan(&storedPublished, &storedUpdated); err != nil {
		t.Fatal(err)
	}
	if !storedPublished.Equal(publishedAt) || !storedUpdated.Equal(publishedAt) || storedPublished.After(storedUpdated) {
		t.Fatalf("stored published=%s updated=%s", storedPublished, storedUpdated)
	}
}

func (h *authoringHarness) testDeletePublishRace(t *testing.T) {
	h.reset(t)
	tree := h.createCourse(t, "delete-publish-race", authoringInstructorID)
	lock, err := h.pool.Begin(h.ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer lock.Rollback(context.Background())
	if _, err := lock.Exec(h.ctx, `SELECT id FROM courses WHERE id=$1 FOR UPDATE`, tree.ID); err != nil {
		t.Fatal(err)
	}
	result := make(chan error, 1)
	go func() { result <- h.repository.Delete(h.ctx, adminActor(), tree.ID, tree.UpdatedAt) }()
	h.waitForBlockedStatement(t, "DELETE FROM courses")
	publishedAt := tree.UpdatedAt.Add(time.Second)
	if _, err := lock.Exec(h.ctx, `UPDATE courses SET status='published', published_at=$2, updated_at=$2 WHERE id=$1`, tree.ID, publishedAt); err != nil {
		t.Fatal(err)
	}
	if err := lock.Commit(h.ctx); err != nil {
		t.Fatal(err)
	}
	if err := <-result; !errors.Is(err, domain.ErrCourseNotDeletable) {
		t.Fatalf("delete race result=%v", err)
	}
	loaded, err := h.repository.Load(h.ctx, tree.ID)
	if err != nil || loaded.Status != domain.Published {
		t.Fatalf("published course lost: tree=%+v err=%v", loaded, err)
	}
}

func (h *authoringHarness) testDeleteCascade(t *testing.T) {
	h.reset(t)
	tree := h.createCourse(t, "delete-cascade", authoringInstructorID)
	tree.Modules = fullTree()
	saved := h.save(t, tree, tree.UpdatedAt.Add(time.Second))
	if err := h.repository.Delete(h.ctx, adminActor(), saved.ID, saved.UpdatedAt); err != nil {
		t.Fatal(err)
	}
	if _, err := h.repository.Load(h.ctx, saved.ID); !errors.Is(err, domain.ErrCourseNotFound) {
		t.Fatalf("deleted course load=%v", err)
	}
	var modules, lessons int
	if err := h.pool.QueryRow(h.ctx, `SELECT count(*) FROM course_modules WHERE id::text = ANY($1::text[])`, []string{authoringModuleOne, authoringModuleTwo}).Scan(&modules); err != nil {
		t.Fatal(err)
	}
	if err := h.pool.QueryRow(h.ctx, `SELECT count(*) FROM course_lessons WHERE id::text = ANY($1::text[])`, []string{authoringLessonOne, authoringLessonTwo, authoringLessonTri}).Scan(&lessons); err != nil {
		t.Fatal(err)
	}
	if modules != 0 || lessons != 0 {
		t.Fatalf("cascade left modules=%d lessons=%d", modules, lessons)
	}
}

func (h *authoringHarness) testOptimisticConflict(t *testing.T) {
	h.reset(t)
	original := h.createCourse(t, "optimistic-conflict", authoringInstructorID)
	original.Modules = fullTree()
	original = h.save(t, original, original.UpdatedAt.Add(time.Second))
	first := cloneTree(original)
	first.Title = "First writer"
	first.Modules[0].Lessons[0].Title = "First writer lesson"
	first = h.save(t, first, first.UpdatedAt.Add(time.Second))
	stale := cloneTree(original)
	stale.Title = "Stale writer"
	stale.Modules = stale.Modules[1:]
	stale.Modules[0].Position = 0
	if _, err := h.repository.Save(h.ctx, adminActor(), stale, original.UpdatedAt, first.UpdatedAt.Add(time.Second)); !errors.Is(err, domain.ErrCourseModified) {
		t.Fatalf("stale save=%v", err)
	}
	loaded, err := h.repository.Load(h.ctx, original.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(loaded, first) {
		t.Fatalf("writers merged: %+v", loaded)
	}
}

func (h *authoringHarness) testForeignNodeIDs(t *testing.T) {
	h.reset(t)
	x := h.createCourse(t, "course-x", authoringInstructorID)
	x.Modules = []domain.Module{{ID: authoringModuleOne, Title: "X module", Position: 0, Lessons: []domain.Lesson{}}}
	x = h.save(t, x, x.UpdatedAt.Add(time.Second))
	y := h.createCourse(t, "course-y", authoringOtherInstructorID)
	y.Modules = []domain.Module{{ID: authoringModuleTwo, Title: "Y module", Position: 0, Lessons: []domain.Lesson{{ID: authoringLessonOne, Title: "Y lesson", Position: 0}}}}
	y = h.save(t, y, y.UpdatedAt.Add(time.Second))
	xBefore, err := h.repository.Load(h.ctx, x.ID)
	if err != nil {
		t.Fatal(err)
	}
	yBefore, err := h.repository.Load(h.ctx, y.ID)
	if err != nil {
		t.Fatal(err)
	}

	foreignModule := cloneTree(xBefore)
	foreignModule.Modules = []domain.Module{{ID: authoringModuleTwo, Title: "Stolen module", Position: 0, Lessons: []domain.Lesson{}}}
	if _, err := h.repository.Save(h.ctx, adminActor(), foreignModule, xBefore.UpdatedAt, xBefore.UpdatedAt.Add(time.Second)); !errors.Is(err, domain.ErrInvalidRequest) {
		t.Fatalf("foreign module save=%v", err)
	}
	foreignLesson := cloneTree(xBefore)
	foreignLesson.Modules[0].Lessons = []domain.Lesson{{ID: authoringLessonOne, Title: "Stolen lesson", Position: 0}}
	if _, err := h.repository.Save(h.ctx, adminActor(), foreignLesson, xBefore.UpdatedAt, xBefore.UpdatedAt.Add(time.Second)); !errors.Is(err, domain.ErrInvalidRequest) {
		t.Fatalf("foreign lesson save=%v", err)
	}
	xAfter, err := h.repository.Load(h.ctx, x.ID)
	if err != nil {
		t.Fatal(err)
	}
	yAfter, err := h.repository.Load(h.ctx, y.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(xAfter, xBefore) || !reflect.DeepEqual(yAfter, yBefore) {
		t.Fatalf("foreign node attempt changed trees:\nX before=%+v\nX after=%+v\nY before=%+v\nY after=%+v", xBefore, xAfter, yBefore, yAfter)
	}
	var moduleID string
	if err := h.pool.QueryRow(h.ctx, `SELECT module_id::text FROM course_lessons WHERE id=$1`, authoringLessonOne).Scan(&moduleID); err != nil {
		t.Fatal(err)
	}
	if moduleID != authoringModuleTwo {
		t.Fatalf("foreign lesson moved to %s", moduleID)
	}
}

func (h *authoringHarness) testCancelledSave(t *testing.T) {
	h.reset(t)
	tree := h.createCourse(t, "cancelled-save", authoringInstructorID)
	tree.Modules = []domain.Module{{ID: authoringModuleOne, Title: "Original module", Position: 0, Lessons: []domain.Lesson{{ID: authoringLessonOne, Title: "Original lesson", Position: 0}}}}
	stored := h.save(t, tree, tree.UpdatedAt.Add(time.Second))
	lock, err := h.pool.Begin(h.ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer lock.Rollback(context.Background())
	if _, err := lock.Exec(h.ctx, `SELECT id FROM course_lessons WHERE id=$1 FOR UPDATE`, authoringLessonOne); err != nil {
		t.Fatal(err)
	}
	edit := cloneTree(stored)
	edit.Modules[0].Title = "Partial module change"
	edit.Modules[0].Lessons[0].Title = "Blocked lesson change"
	request, cancel := context.WithCancel(h.ctx)
	result := make(chan error, 1)
	go func() {
		_, err := h.repository.Save(request, adminActor(), edit, stored.UpdatedAt, stored.UpdatedAt.Add(time.Second))
		result <- err
	}()
	h.waitForBlockedStatement(t, "INSERT INTO course_lessons")
	cancel()
	if err := <-result; !errors.Is(err, domain.ErrDependencyUnavailable) {
		t.Fatalf("cancelled save=%v", err)
	}
	if err := lock.Rollback(h.ctx); err != nil {
		t.Fatal(err)
	}
	after, err := h.repository.Load(h.ctx, stored.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(after, stored) {
		t.Fatalf("cancelled save persisted a partial tree:\nbefore=%+v\nafter=%+v", stored, after)
	}
}

func (h *authoringHarness) testOwnerEligibility(t *testing.T) {
	for _, ownerID := range []string{authoringStudentID, authoringDisabledOwnerID, "40000000-0000-4000-8000-000000000099"} {
		t.Run("create "+ownerID, func(t *testing.T) {
			h.reset(t)
			if _, err := h.repository.Create(h.ctx, adminActor(), draft("invalid-owner", ownerID), authoringTime()); !errors.Is(err, domain.ErrInvalidOwner) {
				t.Fatalf("owner %s: %v", ownerID, err)
			}
			var count int
			if err := h.pool.QueryRow(h.ctx, `SELECT count(*) FROM courses`).Scan(&count); err != nil || count != 0 {
				t.Fatalf("owner %s created rows=%d err=%v", ownerID, count, err)
			}
		})
		t.Run("save "+ownerID, func(t *testing.T) {
			h.reset(t)
			stored := h.createCourse(t, "invalid-reassignment", authoringInstructorID)
			edit := stored
			edit.InstructorID = ownerID
			if _, err := h.repository.Save(h.ctx, adminActor(), edit, stored.UpdatedAt, stored.UpdatedAt.Add(time.Second)); !errors.Is(err, domain.ErrInvalidOwner) {
				t.Fatalf("owner %s: %v", ownerID, err)
			}
			after, err := h.repository.Load(h.ctx, stored.ID)
			if err != nil {
				t.Fatal(err)
			}
			if after.InstructorID != authoringInstructorID || !after.UpdatedAt.Equal(stored.UpdatedAt) {
				t.Fatalf("owner %s changed course: %+v", ownerID, after)
			}
		})
	}
	t.Run("instructor cannot choose another eligible owner", func(t *testing.T) {
		h.reset(t)
		if _, err := h.repository.Create(h.ctx, instructorActor(), draft("foreign-owner", authoringOtherInstructorID), authoringTime()); !errors.Is(err, domain.ErrCourseForbidden) {
			t.Fatalf("foreign owner: %v", err)
		}
	})
	t.Run("instructor omission defaults to self", func(t *testing.T) {
		h.reset(t)
		created, err := h.repository.Create(h.ctx, instructorActor(), draft("self-owner", ""), authoringTime())
		if err != nil {
			t.Fatal(err)
		}
		if created.InstructorID != authoringInstructorID {
			t.Fatalf("owner=%s", created.InstructorID)
		}
	})
}

func (h *authoringHarness) reset(t *testing.T) {
	t.Helper()
	if _, err := h.pool.Exec(h.ctx, `TRUNCATE users CASCADE`); err != nil {
		t.Fatal(err)
	}
	seed, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "sql", "seed-identities.sql"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := h.pool.Exec(h.ctx, string(seed)); err != nil {
		t.Fatal(err)
	}
	if _, err := h.pool.Exec(h.ctx, `
INSERT INTO users (id, email, display_name, role, password_hash, created_at, updated_at, disabled_at)
SELECT $1, 'other-instructor@courseforge.local', 'Other Instructor', 'instructor', password_hash, created_at, updated_at, NULL
	FROM users WHERE id=$2`, authoringOtherInstructorID, authoringInstructorID); err != nil {
		t.Fatal(err)
	}
	if _, err := h.pool.Exec(h.ctx, `
INSERT INTO users (id, email, display_name, role, password_hash, created_at, updated_at, disabled_at)
	SELECT $1, 'disabled-instructor@courseforge.local', 'Disabled Instructor', 'instructor', password_hash, created_at, updated_at, updated_at
	FROM users WHERE id=$2`, authoringDisabledOwnerID, authoringInstructorID); err != nil {
		t.Fatal(err)
	}
}

func (h *authoringHarness) createCourse(t *testing.T, slug, ownerID string) domain.CourseTree {
	t.Helper()
	tree, err := h.repository.Create(h.ctx, adminActor(), draft(slug, ownerID), authoringTime())
	if err != nil {
		t.Fatal(err)
	}
	return tree
}

func (h *authoringHarness) save(t *testing.T, tree domain.CourseTree, updatedAt time.Time) domain.CourseTree {
	t.Helper()
	saved, err := h.repository.Save(h.ctx, adminActor(), tree, tree.UpdatedAt, updatedAt)
	if err != nil {
		t.Fatal(err)
	}
	return saved
}

func (h *authoringHarness) waitForBlockedStatement(t *testing.T, fragment string) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		var blocked bool
		err := h.pool.QueryRow(h.ctx, `
SELECT EXISTS (
  SELECT 1 FROM pg_stat_activity
  WHERE pid <> pg_backend_pid()
    AND datname = current_database()
    AND state = 'active'
    AND wait_event_type = 'Lock'
    AND position($1 in query) > 0
)`, fragment).Scan(&blocked)
		if err != nil {
			t.Fatal(err)
		}
		if blocked {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("statement containing %q did not block", fragment)
}

func draft(slug, ownerID string) domain.CourseDraft {
	return domain.CourseDraft{Slug: slug, Title: "Authoring course", Summary: "Summary", Description: "Description", InstructorID: ownerID}
}

func fullTree() []domain.Module {
	return []domain.Module{
		{ID: authoringModuleOne, Title: "Module one", Position: 0, Lessons: []domain.Lesson{
			{ID: authoringLessonOne, Title: "Lesson one", Position: 0},
			{ID: authoringLessonTwo, Title: "Lesson two", Position: 1},
		}},
		{ID: authoringModuleTwo, Title: "Module two", Position: 1, Lessons: []domain.Lesson{
			{ID: authoringLessonTri, Title: "Lesson three", Position: 0},
		}},
	}
}

func assertTreeContent(t *testing.T, tree domain.CourseTree, want []domain.Module) {
	t.Helper()
	if !reflect.DeepEqual(tree.Modules, want) {
		t.Fatalf("tree mismatch:\ngot:  %+v\nwant: %+v", tree.Modules, want)
	}
}

func cloneTree(tree domain.CourseTree) domain.CourseTree {
	clone := tree
	clone.Modules = make([]domain.Module, len(tree.Modules))
	for index, module := range tree.Modules {
		clone.Modules[index] = module
		clone.Modules[index].Lessons = append([]domain.Lesson(nil), module.Lessons...)
	}
	return clone
}

func authoringTime() time.Time {
	return time.Date(2026, 9, 5, 12, 0, 0, 123456000, time.UTC)
}

func adminActor() domain.Actor {
	return domain.Actor{UserID: authoringAdminID, Role: domain.ActorAdmin}
}

func instructorActor() domain.Actor {
	return domain.Actor{UserID: authoringInstructorID, Role: domain.ActorInstructor}
}
