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

	"courses/internal/contracts/access"
	coursepostgres "courses/internal/course/adapters/postgres"
	"courses/internal/course/domain"
	containers "courses/internal/testsupport/containers"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	postgrescontainer "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestLessonDocumentRepositoryAgainstPostgreSQL(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	container, err := containers.Start(ctx, func(ctx context.Context) (*postgrescontainer.PostgresContainer, error) {
		return postgrescontainer.Run(ctx, "postgres:17.5",
			postgrescontainer.WithDatabase("document_test"),
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
	repository := coursepostgres.NewRepository(pool)
	// This suite owns an isolated database; container termination removes all fixtures.
	var ownerID, viewerID, courseID, moduleID, lessonID string
	for _, user := range []struct {
		role string
		id   *string
	}{{"instructor", &ownerID}, {"student", &viewerID}} {
		if err := pool.QueryRow(ctx, `INSERT INTO users(id,email,display_name,role,password_hash)
VALUES(gen_random_uuid(),$1,'Document user',$2,'$argon2id$v=19$m=65536,t=3,p=2$AAAAAAAAAAAAAAAAAAAAAA$AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA') RETURNING id::text`, user.role+"@document.test", user.role).Scan(user.id); err != nil {
			t.Fatal(err)
		}
	}
	if err := pool.QueryRow(ctx, `INSERT INTO courses(id,instructor_id,slug,title,summary,description,status,published_at,price_minor,currency)
VALUES(gen_random_uuid(),$1,'document-course','Document course','Summary','Description','published',now(),100,'UAH') RETURNING id::text`, ownerID).Scan(&courseID); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx, `INSERT INTO course_modules(id,course_id,title,position) VALUES(gen_random_uuid(),$1,'Module',0) RETURNING id::text`, courseID).Scan(&moduleID); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx, `INSERT INTO course_lessons(id,module_id,title,description,position) VALUES(gen_random_uuid(),$1,'Lesson','',0) RETURNING id::text`, moduleID).Scan(&lessonID); err != nil {
		t.Fatal(err)
	}

	t.Run("draft publication and optimistic locks", func(t *testing.T) {
		empty := domain.LessonDocument{SchemaVersion: 1, Locale: "uk", Sections: []domain.Section{}}
		first := domain.LessonDocument{SchemaVersion: 1, Locale: "uk", Sections: []domain.Section{{ID: "video", Type: domain.SectionVideo}}}
		second := domain.LessonDocument{SchemaVersion: 1, Locale: "uk", Sections: []domain.Section{{ID: "text", Type: domain.SectionText, Content: &domain.RichText{Blocks: []domain.Block{{Type: domain.BlockParagraph, Spans: []domain.Span{{Text: "Український текст <literal>", Marks: []domain.Mark{domain.MarkBold}}}}}}}}}
		assertDraft := func(want domain.LessonDocument, token time.Time) time.Time {
			t.Helper()
			got, updatedAt, found, err := repository.LoadDraft(ctx, lessonID, "uk")
			if err != nil || !found || !reflect.DeepEqual(got, want) || updatedAt.IsZero() || !token.IsZero() && !updatedAt.Equal(token) {
				t.Fatalf("draft = %+v, %v, %v, %v; want %+v at %v", got, updatedAt, found, err, want, token)
			}
			return updatedAt
		}
		assertPublished := func(want domain.LessonDocument) {
			t.Helper()
			got, found, err := repository.LoadPublished(ctx, lessonID, "uk")
			if err != nil || !found || !reflect.DeepEqual(got, want) {
				t.Fatalf("published = %+v, %v, %v; want %+v", got, found, err, want)
			}
			if err := domain.ValidateLessonDocument(got); err != nil {
				t.Fatal(err)
			}
		}
		if _, at, found, err := repository.LoadDraft(ctx, lessonID, "uk"); err != nil || found || !at.IsZero() {
			t.Fatalf("missing draft: %v, %v, %v", at, found, err)
		}
		if _, found, err := repository.LoadPublished(ctx, lessonID, "uk"); err != nil || found {
			t.Fatalf("missing published: %v, %v", found, err)
		}
		if err := repository.PublishDocument(ctx, lessonID, "uk", time.Time{}); !errors.Is(err, domain.ErrDocumentConflict) {
			t.Fatalf("missing publish: %v", err)
		}
		initial, err := repository.SaveDraft(ctx, lessonID, "uk", first, time.Time{})
		if err != nil {
			t.Fatal(err)
		}
		assertDraft(first, initial)
		assertPublished(empty)
		next, err := repository.SaveDraft(ctx, lessonID, "uk", second, initial)
		if err != nil {
			t.Fatal(err)
		}
		if next.Equal(initial) {
			t.Fatal("save did not advance token")
		}
		assertDraft(second, next)
		assertPublished(empty)
		for _, stale := range []time.Time{time.Time{}, initial} {
			if _, err := repository.SaveDraft(ctx, lessonID, "uk", first, stale); !errors.Is(err, domain.ErrDocumentConflict) {
				t.Fatalf("stale save: %v", err)
			}
			if err := repository.PublishDocument(ctx, lessonID, "uk", stale); !errors.Is(err, domain.ErrDocumentConflict) {
				t.Fatalf("stale publish: %v", err)
			}
		}
		assertDraft(second, next)
		assertPublished(empty)
		if err := repository.PublishDocument(ctx, lessonID, "uk", next); err != nil {
			t.Fatal(err)
		}
		publishedAt := assertDraft(second, time.Time{})
		if publishedAt.Equal(next) {
			t.Fatal("publish did not advance token")
		}
		assertPublished(second)
		// Subsequent edits stay private; a stale publication cannot expose them.
		editedAt, err := repository.SaveDraft(ctx, lessonID, "uk", first, publishedAt)
		if err != nil {
			t.Fatal(err)
		}
		if err := repository.PublishDocument(ctx, lessonID, "uk", publishedAt); !errors.Is(err, domain.ErrDocumentConflict) {
			t.Fatalf("stale publish after edit: %v", err)
		}
		assertDraft(first, editedAt)
		assertPublished(second)
		english := domain.LessonDocument{SchemaVersion: 1, Locale: "en", Sections: []domain.Section{}}
		if _, err := repository.SaveDraft(ctx, lessonID, "en", english, time.Time{}); err != nil {
			t.Fatal(err)
		}
		got, _, found, err := repository.LoadDraft(ctx, lessonID, "en")
		if err != nil || !found || !reflect.DeepEqual(got, english) {
			t.Fatalf("English draft: %+v, %v, %v", got, found, err)
		}
		assertDraft(first, editedAt)
		assertPublished(second)
	})

	t.Run("per-viewer lesson access facts", func(t *testing.T) {
		owner := domain.Actor{UserID: ownerID, Role: domain.ActorInstructor}
		viewer := domain.Actor{UserID: viewerID, Role: domain.ActorStudent}
		facts, err := repository.LoadLessonAccess(ctx, ownerID, lessonID)
		if err != nil || facts.OwnerID != ownerID || !facts.CanBeEditedBy(owner) {
			t.Fatalf("owner: %+v, %v", facts, err)
		}
		checkViewer := func(wantMembership access.Membership, entitled, preview, consume bool) {
			t.Helper()
			facts, err := repository.LoadLessonAccess(ctx, viewerID, lessonID)
			if err != nil || facts.OwnerID != ownerID || facts.Status != domain.Published || !facts.Paid || facts.Entitled != entitled || facts.Membership != wantMembership || facts.Preview != preview || facts.CanConsumeBy(viewer) != consume || facts.CanBeEditedBy(viewer) {
				t.Fatalf("viewer: %+v, %v; membership=%s entitled=%v preview=%v consume=%v", facts, err, wantMembership, entitled, preview, consume)
			}
		}
		checkViewer(access.MembershipNone, false, false, false)
		if _, err := pool.Exec(ctx, `INSERT INTO course_grants(user_id,course_id,source,state,granted_by) VALUES($1,$2,'admin','active',$3)`, viewerID, courseID, ownerID); err != nil {
			t.Fatal(err)
		}
		checkViewer(access.MembershipActive, true, false, true)
		for _, state := range []access.Membership{access.MembershipBlocked, access.MembershipRevoked} {
			if _, err := pool.Exec(ctx, `UPDATE course_grants SET state=$3 WHERE user_id=$1 AND course_id=$2`, viewerID, courseID, string(state)); err != nil {
				t.Fatal(err)
			}
			checkViewer(state, false, false, state == access.MembershipBlocked)
		}
		if _, err := pool.Exec(ctx, `DELETE FROM course_grants WHERE user_id=$1 AND course_id=$2`, viewerID, courseID); err != nil {
			t.Fatal(err)
		}
		if _, err := pool.Exec(ctx, `UPDATE course_lessons SET is_preview=true WHERE id=$1`, lessonID); err != nil {
			t.Fatal(err)
		}
		checkViewer(access.MembershipNone, false, true, true)
		if _, err := pool.Exec(ctx, `UPDATE courses SET status='draft', published_at=NULL, price_minor=NULL, currency=NULL WHERE id=$1`, courseID); err != nil {
			t.Fatal(err)
		}
		facts, err = repository.LoadLessonAccess(ctx, viewerID, lessonID)
		if err != nil || facts.Status != domain.Draft || facts.Paid || facts.CanConsumeBy(viewer) {
			t.Fatalf("free draft preview: %+v, %v", facts, err)
		}
		if _, err := repository.LoadLessonAccess(ctx, viewerID, "ffffffff-ffff-4fff-8fff-ffffffffffff"); !errors.Is(err, domain.ErrLessonNotFound) {
			t.Fatalf("missing lesson: %v", err)
		}
	})
}
