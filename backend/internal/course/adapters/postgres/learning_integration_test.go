//go:build integration

package postgres_test

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	coursepostgres "courses/internal/course/adapters/postgres"
	"courses/internal/course/application"
	"courses/internal/course/domain"
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
	container, err := postgrescontainer.Run(ctx, "postgres:17.5",
		postgrescontainer.WithDatabase("learning_test"),
		postgrescontainer.WithUsername("courses"),
		postgrescontainer.WithPassword("ephemeral-test-only"),
		testcontainers.WithWaitStrategy(wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(3*time.Minute)),
	)
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
	learningModuleID              = "61000000-0000-4000-8000-000000000001"
	learningArchiveModule         = "61000000-0000-4000-8000-000000000002"
	learningB07Module             = "61000000-0000-4000-8000-000000000003"
	learningLessonOne             = "71000000-0000-4000-8000-000000000001"
	learningLessonTwo             = "71000000-0000-4000-8000-000000000002"
	learningLessonThree           = "71000000-0000-4000-8000-000000000003"
	learningLessonFour            = "71000000-0000-4000-8000-000000000004"
	learningArchivedLesson        = "71000000-0000-4000-8000-000000000005"
	learningB07ShortLesson        = "71000000-0000-4000-8000-000000000006"
	learningB07ProcessingLesson   = "71000000-0000-4000-8000-000000000007"
	learningB07NullDurationLesson = "71000000-0000-4000-8000-000000000008"
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
	}
}
