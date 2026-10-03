//go:build integration

package postgres_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"courses/internal/contracts/access/accesstest"
	coursepostgres "courses/internal/course/adapters/postgres"
	"courses/internal/course/application"
	"courses/internal/course/domain"
	containers "courses/internal/testsupport/containers"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	postgrescontainer "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestLearningReaderAgainstPostgreSQL(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	container, err := containers.Start(ctx, func(ctx context.Context) (*postgrescontainer.PostgresContainer, error) {
		return postgrescontainer.Run(ctx, "postgres:17.5",
			postgrescontainer.WithDatabase("learning_test"),
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

	for _, fixture := range learningFixtures() {
		if _, err := pool.Exec(ctx, fixture); err != nil {
			t.Fatal(err)
		}
	}
	repository := coursepostgres.NewRepository(pool)

	t.Run("ratings and lesson count match the public catalog", func(t *testing.T) {
		// The fixture course is shared with later subtests (B27 counts its program): remove everything
		// this subtest adds once it ends.
		t.Cleanup(func() {
			for _, statement := range []string{
				`DELETE FROM discussion_reviews WHERE id IN ('81000000-0000-4000-8000-000000000011','81000000-0000-4000-8000-000000000012','81000000-0000-4000-8000-000000000013','81000000-0000-4000-8000-000000000021','81000000-0000-4000-8000-000000000022')`,
				`DELETE FROM course_lessons WHERE id='71000000-0000-4000-8000-000000000011'`,
				`DELETE FROM course_modules WHERE id='61000000-0000-4000-8000-000000000011'`,
				`DELETE FROM users WHERE id IN ('41000000-0000-4000-8000-000000000011','41000000-0000-4000-8000-000000000012','41000000-0000-4000-8000-000000000013')`,
			} {
				if _, err := pool.Exec(context.Background(), statement); err != nil {
					t.Errorf("cleanup %q: %v", statement, err)
				}
			}
		})
		// One statement per Exec: a parameterised query cannot carry several commands (SQLSTATE 42601).
		for _, fixture := range []struct {
			sql  string
			args []any
		}{
			{`INSERT INTO users(id,email,display_name,role,password_hash) VALUES
			('41000000-0000-4000-8000-000000000011','rating-one@learning.test','Rating one','student','$argon2id$v=19$m=65536,t=3,p=2$AAAAAAAAAAAAAAAAAAAAAA$AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA'),
			('41000000-0000-4000-8000-000000000012','rating-two@learning.test','Rating two','student','$argon2id$v=19$m=65536,t=3,p=2$AAAAAAAAAAAAAAAAAAAAAA$AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA'),
			('41000000-0000-4000-8000-000000000013','rating-three@learning.test','Rating three','student','$argon2id$v=19$m=65536,t=3,p=2$AAAAAAAAAAAAAAAAAAAAAA$AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA')`, nil},
			{`INSERT INTO course_modules(id,course_id,title,position) VALUES ('61000000-0000-4000-8000-000000000011',$1,'Second module',1)`, []any{learningCourseID}},
			{`INSERT INTO course_lessons(id,module_id,title,description,position) VALUES ('71000000-0000-4000-8000-000000000011','61000000-0000-4000-8000-000000000011','Second module lesson','',0)`, nil},
			{`INSERT INTO discussion_reviews(id,course_id,author_id,stars,body,completed_lessons,total_lessons,created_at) VALUES
			('81000000-0000-4000-8000-000000000011',$1,'41000000-0000-4000-8000-000000000011',4,'First learning rating is long enough.',0,5,now()),
			('81000000-0000-4000-8000-000000000012',$1,'41000000-0000-4000-8000-000000000012',5,'Second learning rating is long enough.',0,5,now()),
			('81000000-0000-4000-8000-000000000013',$1,'41000000-0000-4000-8000-000000000013',5,'Third learning rating is long enough.',0,5,now())`, []any{learningCourseID}},
		} {
			if _, err := pool.Exec(ctx, fixture.sql, fixture.args...); err != nil {
				t.Fatal(err)
			}
		}
		learning, err := repository.GetLearningCourse(ctx, learningViewerID, "learning-reader")
		if err != nil {
			t.Fatal(err)
		}
		catalog, err := repository.GetPublished(ctx, "learning-reader")
		if err != nil {
			t.Fatal(err)
		}
		if learning.LessonCount != 5 || catalog.LessonCount != 5 || learning.Rating == nil || catalog.Rating == nil || *learning.Rating != *catalog.Rating || learning.Rating.Average != 4.7 || learning.Rating.Count != 3 {
			t.Fatalf("learning=%+v catalog=%+v", learning.Course, catalog.Course)
		}
		if _, err := pool.Exec(ctx, `INSERT INTO discussion_reviews(id,course_id,author_id,stars,body,completed_lessons,total_lessons,created_at) VALUES
			('81000000-0000-4000-8000-000000000021',$1,'41000000-0000-4000-8000-000000000011',4,'First below threshold review body.',0,1,now()),
			('81000000-0000-4000-8000-000000000022',$1,'41000000-0000-4000-8000-000000000012',5,'Second below threshold review body.',0,1,now())`, learningB08Course); err != nil {
			t.Fatal(err)
		}
		learning, err = repository.GetLearningCourse(ctx, learningViewerID, "learning-b08")
		if err != nil {
			t.Fatal(err)
		}
		catalog, err = repository.GetPublished(ctx, "learning-b08")
		if err != nil {
			t.Fatal(err)
		}
		if learning.Rating != nil || catalog.Rating != nil {
			t.Fatalf("below threshold learning=%+v catalog=%+v", learning.Rating, catalog.Rating)
		}
	})

	t.Run("R06 SQL facts cover the access truth table", func(t *testing.T) {
		const tableOwner = "41000000-0000-4000-8000-000000000002"
		const tableStranger = "41000000-0000-4000-8000-000000000001"
		for index, test := range accesstest.TruthTable {
			t.Run(test.Name, func(t *testing.T) {
				slug := fmt.Sprintf("access-table-%d", index)
				var courseID, moduleID, lessonID string
				publishedAt := "NULL"
				if test.Status == "published" {
					publishedAt = "now()"
				}
				var price, currency any
				if test.Paid {
					price, currency = int64(50000), "UAH"
				}
				if err := pool.QueryRow(ctx, `INSERT INTO courses(id,instructor_id,slug,title,summary,description,status,published_at,price_minor,currency)
					VALUES(gen_random_uuid(),$1,$2,'Access table','Summary','Description',$3,`+publishedAt+`,$4,$5) RETURNING id::text`, tableOwner, slug, test.Status, price, currency).Scan(&courseID); err != nil {
					t.Fatal(err)
				}
				if err := pool.QueryRow(ctx, `INSERT INTO course_modules(id,course_id,title,position) VALUES(gen_random_uuid(),$1,'Module',0) RETURNING id::text`, courseID).Scan(&moduleID); err != nil {
					t.Fatal(err)
				}
				if err := pool.QueryRow(ctx, `INSERT INTO course_lessons(id,module_id,title,description,position,is_preview) VALUES(gen_random_uuid(),$1,'Lesson','',0,$2) RETURNING id::text`, moduleID, test.Preview).Scan(&lessonID); err != nil {
					t.Fatal(err)
				}

				actorID := tableStranger
				role := domain.ActorStudent
				if test.Role == accesstest.Owner {
					actorID, role = tableOwner, domain.ActorInstructor
				} else if test.Role == accesstest.Administrator {
					role = domain.ActorAdmin
				}
				if test.Granted {
					if _, err := pool.Exec(ctx, `INSERT INTO course_grants(user_id,course_id,source,state,granted_by) VALUES($1,$2,'admin','active',$3)`, actorID, courseID, tableOwner); err != nil {
						t.Fatal(err)
					}
				}
				course, err := repository.GetLearningCourse(ctx, actorID, slug)
				if err != nil {
					t.Fatal(err)
				}
				// Content access for the course module is "outline visible AND lesson not
				// locked" (owner decision 2026-10-01); the lesson's preview fact comes from SQL.
				viewer := domain.Actor{UserID: actorID, Role: role}
				lesson := course.Modules[0].Lessons[0]
				if lesson.IsPreview != test.Preview {
					t.Fatalf("lesson preview=%v, want %v from SQL", lesson.IsPreview, test.Preview)
				}
				if got := course.CanBeViewedBy(viewer) && !course.PurchaseRequiredBy(viewer, lesson); got != test.Want {
					t.Fatalf("adapter facts=%+v actor=%+v content visible=%v, want %v", course.Course, actorID, got, test.Want)
				}
			})
		}
	})

	t.Run("Pay11 preview on a later lesson is that lesson's fact, not the first lesson's", func(t *testing.T) {
		// The defect: the course took its preview fact from the first row (first lesson),
		// so a preview on lesson 2 refused the whole course. Three lessons, preview on 2.
		const owner = "41000000-0000-4000-8000-000000000002"
		const stranger = "41000000-0000-4000-8000-000000000001"
		var courseID, moduleID string
		if err := pool.QueryRow(ctx, `INSERT INTO courses(id,instructor_id,slug,title,summary,description,status,published_at,price_minor,currency)
			VALUES(gen_random_uuid(),$1,'preview-on-second','Preview on second','Summary','Description','published',now(),50000,'UAH') RETURNING id::text`, owner).Scan(&courseID); err != nil {
			t.Fatal(err)
		}
		if err := pool.QueryRow(ctx, `INSERT INTO course_modules(id,course_id,title,position) VALUES(gen_random_uuid(),$1,'Module',0) RETURNING id::text`, courseID).Scan(&moduleID); err != nil {
			t.Fatal(err)
		}
		for position, preview := range []bool{false, true, false} {
			if _, err := pool.Exec(ctx, `INSERT INTO course_lessons(id,module_id,title,description,position,is_preview) VALUES(gen_random_uuid(),$1,$2,'',$3,$4)`, moduleID, fmt.Sprintf("Lesson %d", position+1), position, preview); err != nil {
				t.Fatal(err)
			}
		}
		course, err := repository.GetLearningCourse(ctx, stranger, "preview-on-second")
		if err != nil {
			t.Fatal(err)
		}
		// The locked lessons ask for this amount, so the outline carries the price itself.
		if !course.Paid || course.Price == nil || course.Price.Minor != 50000 || course.Price.Currency != "UAH" {
			t.Fatalf("paid=%v price=%+v, want the stored 50000 UAH", course.Paid, course.Price)
		}
		viewer := domain.Actor{UserID: stranger, Role: domain.ActorStudent}
		if !course.CanBeViewedBy(viewer) {
			t.Fatal("the outline of a published paid course must be visible to a viewer without access")
		}
		var previews, locks []bool
		for _, lesson := range course.Modules[0].Lessons {
			previews = append(previews, lesson.IsPreview)
			locks = append(locks, course.PurchaseRequiredBy(viewer, lesson))
		}
		if fmt.Sprint(previews) != "[false true false]" || fmt.Sprint(locks) != "[true false true]" {
			t.Fatalf("previews=%v locks=%v, want [false true false] and [true false true]", previews, locks)
		}
	})

	t.Run("B01 course without lessons", func(t *testing.T) {
		course, err := repository.GetLearningCourse(ctx, learningViewerID, "empty-learning")
		if err != nil {
			t.Fatal(err)
		}
		if course.Modules == nil || len(course.Modules) != 0 || course.ContinueLessonID != "" {
			t.Fatalf("empty course=%+v", course)
		}
	})

	t.Run("B27 ordered progress and playable video projection", func(t *testing.T) {
		course, err := repository.GetLearningCourse(ctx, learningViewerID, "learning-reader")
		if err != nil {
			t.Fatal(err)
		}
		if len(course.Modules) != 1 || len(course.Modules[0].Lessons) != 4 {
			t.Fatalf("program=%+v", course.Modules)
		}
		lessons := course.Modules[0].Lessons
		if lessons[0].ID != learningLessonOne || !lessons[0].Completed || lessons[0].Position != 95 || lessons[0].ClientSequence != 3 || lessons[0].VideoState != domain.VideoReady {
			t.Fatalf("first=%+v", lessons[0])
		}
		if lessons[1].ID != learningLessonTwo || lessons[1].Position != 0 || lessons[1].Completed || lessons[1].ClientSequence != 0 || lessons[1].VideoState != domain.VideoAbsent {
			t.Fatalf("B27 second=%+v", lessons[1])
		}
		if lessons[2].VideoState != domain.VideoReady {
			t.Fatalf("older ready must beat newer processing: %+v", lessons[2])
		}
		if lessons[3].VideoState != domain.VideoFailed {
			t.Fatalf("failed projection=%+v", lessons[3])
		}
		if course.ContinueLessonID != learningLessonTwo {
			t.Fatalf("continueLessonId=%q", course.ContinueLessonID)
		}
	})

	t.Run("B07 shorter replacement clamps read without rewriting stored progress", func(t *testing.T) {
		var storedBefore float64
		if err := pool.QueryRow(ctx, `SELECT position_seconds::double precision FROM lesson_progress WHERE user_id=$1 AND lesson_id=$2`, learningViewerID, learningB07ShortLesson).Scan(&storedBefore); err != nil {
			t.Fatal(err)
		}
		if storedBefore != 2050 {
			t.Fatalf("stored position before read=%v want 2050", storedBefore)
		}

		course, err := repository.GetLearningCourse(ctx, learningViewerID, "learning-b07")
		if err != nil {
			t.Fatal(err)
		}
		if len(course.Modules) != 1 || len(course.Modules[0].Lessons) != 3 {
			t.Fatalf("B07 program=%+v", course.Modules)
		}
		lesson := course.Modules[0].Lessons[0]
		if lesson.ID != learningB07ShortLesson || lesson.Position != 480 || lesson.Completed {
			t.Fatalf("short replacement=%+v", lesson)
		}

		var storedAfter float64
		if err := pool.QueryRow(ctx, `SELECT position_seconds::double precision FROM lesson_progress WHERE user_id=$1 AND lesson_id=$2`, learningViewerID, learningB07ShortLesson).Scan(&storedAfter); err != nil {
			t.Fatal(err)
		}
		if storedAfter != 2050 {
			t.Fatalf("stored position after read=%v want 2050", storedAfter)
		}
	})

	t.Run("B08 completed mark survives video replacement", func(t *testing.T) {
		var oldGenerationEvicted bool
		var currentGeneration int
		if err := pool.QueryRow(ctx, `
			SELECT old_asset.evicted_at IS NOT NULL, current_asset.generation
			FROM video_assets AS old_asset
			JOIN video_assets AS current_asset ON current_asset.lesson_id=old_asset.lesson_id
			WHERE old_asset.lesson_id=$1 AND old_asset.generation=1
				AND current_asset.generation=2 AND current_asset.evicted_at IS NULL`, learningB08Lesson).Scan(&oldGenerationEvicted, &currentGeneration); err != nil {
			t.Fatal(err)
		}
		if !oldGenerationEvicted || currentGeneration != 2 {
			t.Fatalf("B08 replacement fixture oldEvicted=%v currentGeneration=%d", oldGenerationEvicted, currentGeneration)
		}

		course, err := repository.GetLearningCourse(ctx, learningViewerID, "learning-b08")
		if err != nil {
			t.Fatal(err)
		}
		if len(course.Modules) != 1 || len(course.Modules[0].Lessons) != 1 {
			t.Fatalf("B08 program=%+v", course.Modules)
		}
		lesson := course.Modules[0].Lessons[0]
		if lesson.ID != learningB08Lesson {
			t.Fatalf("B08 replacement lesson=%+v", lesson)
		}
		if !lesson.Completed {
			t.Fatalf("B08 replacement completed=%v want true", lesson.Completed)
		}
	})

	t.Run("non-ready asset duration does not clamp", func(t *testing.T) {
		course, err := repository.GetLearningCourse(ctx, learningViewerID, "learning-b07")
		if err != nil {
			t.Fatal(err)
		}
		if len(course.Modules) != 1 || len(course.Modules[0].Lessons) != 3 {
			t.Fatalf("B07 program=%+v", course.Modules)
		}
		lesson := course.Modules[0].Lessons[1]
		if lesson.ID != learningB07ProcessingLesson || lesson.Position != 2050 || lesson.VideoState != domain.VideoProcessing {
			t.Fatalf("processing asset=%+v", lesson)
		}
	})

	t.Run("null ready asset duration does not clamp", func(t *testing.T) {
		course, err := repository.GetLearningCourse(ctx, learningViewerID, "learning-b07")
		if err != nil {
			t.Fatal(err)
		}
		if len(course.Modules) != 1 || len(course.Modules[0].Lessons) != 3 {
			t.Fatalf("B07 program=%+v", course.Modules)
		}
		lesson := course.Modules[0].Lessons[2]
		if lesson.ID != learningB07NullDurationLesson || lesson.Position != 2050 || lesson.VideoState != domain.VideoReady {
			t.Fatalf("null duration=%+v", lesson)
		}
	})

	t.Run("R17 denied read creates no progress and R34 owner reads archived", func(t *testing.T) {
		learning := application.NewLearning(repository)
		viewer := domain.Actor{UserID: learningViewerID, Role: domain.ActorStudent}
		if _, err := learning.GetCourse(ctx, viewer, "archived-learning"); !errors.Is(err, domain.ErrCourseForbidden) {
			t.Fatalf("viewer error=%v", err)
		}
		var rows int
		if err := pool.QueryRow(ctx, `SELECT count(*) FROM lesson_progress WHERE user_id=$1 AND lesson_id=$2`, learningViewerID, learningArchivedLesson).Scan(&rows); err != nil {
			t.Fatal(err)
		}
		if rows != 0 {
			t.Fatalf("denied read created %d progress rows", rows)
		}
		owner := domain.Actor{UserID: learningOwnerID, Role: domain.ActorInstructor}
		if _, err := learning.GetCourse(ctx, owner, "archived-learning"); err != nil {
			t.Fatalf("R34 archived owner error=%v", err)
		}
	})

	if _, err := repository.GetLearningCourse(ctx, learningViewerID, "missing-learning"); !errors.Is(err, domain.ErrCourseNotFound) {
		t.Fatalf("missing course error=%v", err)
	}
}

const (
	learningViewerID              = "41000000-0000-4000-8000-000000000001"
	learningOwnerID               = "41000000-0000-4000-8000-000000000002"
	learningCourseID              = "51000000-0000-4000-8000-000000000001"
	learningEmptyCourseID         = "51000000-0000-4000-8000-000000000002"
	learningArchiveCourse         = "51000000-0000-4000-8000-000000000003"
	learningB07Course             = "51000000-0000-4000-8000-000000000004"
	learningB08Course             = "51000000-0000-4000-8000-000000000005"
	learningModuleID              = "61000000-0000-4000-8000-000000000001"
	learningArchiveModule         = "61000000-0000-4000-8000-000000000002"
	learningB07Module             = "61000000-0000-4000-8000-000000000003"
	learningB08Module             = "61000000-0000-4000-8000-000000000004"
	learningLessonOne             = "71000000-0000-4000-8000-000000000001"
	learningLessonTwo             = "71000000-0000-4000-8000-000000000002"
	learningLessonThree           = "71000000-0000-4000-8000-000000000003"
	learningLessonFour            = "71000000-0000-4000-8000-000000000004"
	learningArchivedLesson        = "71000000-0000-4000-8000-000000000005"
	learningB07ShortLesson        = "71000000-0000-4000-8000-000000000006"
	learningB07ProcessingLesson   = "71000000-0000-4000-8000-000000000007"
	learningB07NullDurationLesson = "71000000-0000-4000-8000-000000000008"
	learningB08Lesson             = "71000000-0000-4000-8000-000000000009"
)

func learningFixtures() []string {
	return []string{
		`INSERT INTO users(id,email,display_name,role,password_hash) VALUES
		('` + learningViewerID + `','viewer@learning.test','Learning viewer','student','$argon2id$v=19$m=65536,t=3,p=2$AAAAAAAAAAAAAAAAAAAAAA$AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA'),
		('` + learningOwnerID + `','owner@learning.test','Learning owner','instructor','$argon2id$v=19$m=65536,t=3,p=2$AAAAAAAAAAAAAAAAAAAAAA$AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA')`,
		`INSERT INTO courses(id,instructor_id,slug,title,summary,description,status,published_at) VALUES
			('` + learningCourseID + `','` + learningOwnerID + `','learning-reader','Learning reader','Learning summary','Learning description','published',now()),
			('` + learningEmptyCourseID + `','` + learningOwnerID + `','empty-learning','Empty learning','Empty summary','Empty description','published',now()),
			('` + learningArchiveCourse + `','` + learningOwnerID + `','archived-learning','Archived learning','Archived summary','Archived description','archived',now()),
			('` + learningB07Course + `','` + learningOwnerID + `','learning-b07','Learning B07','Learning B07 summary','Learning B07 description','published',now())`,
		`INSERT INTO course_modules(id,course_id,title,position) VALUES
			('` + learningModuleID + `','` + learningCourseID + `','Learning module',0),
			('` + learningArchiveModule + `','` + learningArchiveCourse + `','Archived module',0),
			('` + learningB07Module + `','` + learningB07Course + `','B07 module',0)`,
		`INSERT INTO course_lessons(id,module_id,title,description,position) VALUES
		('` + learningLessonOne + `','` + learningModuleID + `','Lesson one','',0),
		('` + learningLessonTwo + `','` + learningModuleID + `','Lesson two','',1),
			('` + learningLessonThree + `','` + learningModuleID + `','Lesson three','',2),
			('` + learningLessonFour + `','` + learningModuleID + `','Lesson four','',3),
			('` + learningArchivedLesson + `','` + learningArchiveModule + `','Archived lesson','',0),
			('` + learningB07ShortLesson + `','` + learningB07Module + `','Short replacement','',0),
			('` + learningB07ProcessingLesson + `','` + learningB07Module + `','Processing duration','',1),
			('` + learningB07NullDurationLesson + `','` + learningB07Module + `','Null duration','',2)`,
		`INSERT INTO lesson_progress(user_id,lesson_id,position_seconds,completed_at,client_seq,updated_at) VALUES
			('` + learningViewerID + `','` + learningLessonOne + `',95,now(),3,now()),
			('` + learningViewerID + `','` + learningLessonThree + `',100,now(),4,now()),
			('` + learningViewerID + `','` + learningB07ShortLesson + `',2050,NULL,5,now()),
			('` + learningViewerID + `','` + learningB07ProcessingLesson + `',2050,NULL,6,now()),
			('` + learningViewerID + `','` + learningB07NullDurationLesson + `',2050,NULL,7,now())`,
		`INSERT INTO video_assets(id,lesson_id,generation,state,declared_bytes,received_bytes,original_name,renditions,failure_reason,expires_at,created_at,updated_at) VALUES
		('81000000-0000-4000-8000-000000000001','` + learningLessonOne + `',1,'ready',1,1,'one.mp4',ARRAY['360p'],NULL,now()+interval '1 day',now(),now()),
		('81000000-0000-4000-8000-000000000002','` + learningLessonThree + `',1,'ready',1,1,'three-old.mp4',ARRAY['360p'],NULL,now()+interval '1 day',now(),now()),
			('81000000-0000-4000-8000-000000000003','` + learningLessonThree + `',2,'processing',1,1,'three-new.mp4','{}',NULL,now()+interval '1 day',now(),now()),
			('81000000-0000-4000-8000-000000000004','` + learningLessonFour + `',1,'failed',1,1,'four.mp4','{}','render_failed',now()+interval '1 day',now(),now())`,
		`INSERT INTO video_assets(id,lesson_id,generation,state,declared_bytes,received_bytes,original_name,renditions,failure_reason,expires_at,evicted_at,created_at,updated_at,duration_seconds) VALUES
			('82000000-0000-4000-8000-000000000001','` + learningB07ShortLesson + `',1,'ready',1,1,'long-old.mp4',ARRAY['360p'],NULL,now()+interval '1 day',now()-interval '1 hour',now()-interval '2 hours',now()-interval '1 hour',2050),
			('82000000-0000-4000-8000-000000000002','` + learningB07ShortLesson + `',2,'ready',1,1,'short-new.mp4',ARRAY['360p'],NULL,now()+interval '1 day',NULL,now(),now(),480),
			('82000000-0000-4000-8000-000000000003','` + learningB07ProcessingLesson + `',1,'processing',1,1,'processing.mp4','{}',NULL,now()+interval '1 day',NULL,now(),now(),480),
			('82000000-0000-4000-8000-000000000004','` + learningB07NullDurationLesson + `',1,'ready',1,1,'unknown.mp4',ARRAY['360p'],NULL,now()+interval '1 day',NULL,now(),now(),NULL)`,
		`INSERT INTO courses(id,instructor_id,slug,title,summary,description,status,published_at) VALUES
			('` + learningB08Course + `','` + learningOwnerID + `','learning-b08','Learning B08','Learning B08 summary','Learning B08 description','published',now())`,
		`INSERT INTO course_modules(id,course_id,title,position) VALUES
			('` + learningB08Module + `','` + learningB08Course + `','B08 module',0)`,
		`INSERT INTO course_lessons(id,module_id,title,description,position) VALUES
			('` + learningB08Lesson + `','` + learningB08Module + `','Watched replacement','',0)`,
		`INSERT INTO lesson_progress(user_id,lesson_id,position_seconds,completed_at,client_seq,updated_at) VALUES
			('` + learningViewerID + `','` + learningB08Lesson + `',95,now()-interval '2 hours',8,now()-interval '2 hours')`,
		`INSERT INTO video_assets(id,lesson_id,generation,state,declared_bytes,received_bytes,original_name,renditions,failure_reason,expires_at,evicted_at,created_at,updated_at,duration_seconds) VALUES
			('83000000-0000-4000-8000-000000000001','` + learningB08Lesson + `',1,'ready',1,1,'watched-old.mp4',ARRAY['360p'],NULL,now()+interval '1 day',now()-interval '1 hour',now()-interval '3 hours',now()-interval '1 hour',100),
			('83000000-0000-4000-8000-000000000002','` + learningB08Lesson + `',2,'ready',1,1,'replacement-new.mp4',ARRAY['360p'],NULL,now()+interval '1 day',NULL,now()-interval '1 hour',now()-interval '1 hour',100)`,
	}
}
