//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"courses/internal/contracts/access/accesstest"
	"courses/internal/playback/domain"
	containers "courses/internal/testsupport/containers"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	postgrescontainer "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestVideoAccessReaderAgainstPostgreSQL(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	container, err := containers.Start(ctx, func(ctx context.Context) (*postgrescontainer.PostgresContainer, error) {
		return postgrescontainer.Run(ctx, "postgres:17.5",
			postgrescontainer.WithDatabase("playback_test"),
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

	const (
		ownerID    = "49000000-0000-4000-8000-000000000001"
		courseID   = "19000000-0000-4000-8000-000000000001"
		moduleID   = "69000000-0000-4000-8000-000000000001"
		lessonID   = "79000000-0000-4000-8000-000000000001"
		videoID    = "89000000-0000-4000-8000-000000000001"
		newVideoID = "89000000-0000-4000-8000-000000000002"
		viewerID   = "49000000-0000-4000-8000-000000000002"
	)
	fixtures := []struct {
		query string
		args  []any
	}{
		{`INSERT INTO users(id,email,display_name,role,password_hash)
		 VALUES($1,'owner@playback.test','Playback owner','instructor','$argon2id$v=19$m=65536,t=3,p=2$AAAAAAAAAAAAAAAAAAAAAA$AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA')`, []any{ownerID}},
		{`INSERT INTO users(id,email,display_name,role,password_hash)
		 VALUES($1,'viewer@playback.test','Playback viewer','student','$argon2id$v=19$m=65536,t=3,p=2$AAAAAAAAAAAAAAAAAAAAAA$AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA')`, []any{viewerID}},
		{`INSERT INTO courses(id,instructor_id,slug,title,summary,description,status,published_at)
		 VALUES($1,$2,'playback-test','Playback test','Playback test summary','Playback test description','published',now())`, []any{courseID, ownerID}},
		{`INSERT INTO course_modules(id,course_id,title,position) VALUES($1,$2,'Module',0)`, []any{moduleID, courseID}},
		{`INSERT INTO course_lessons(id,module_id,title,description,position) VALUES($1,$2,'Lesson','',0)`, []any{lessonID, moduleID}},
		{`INSERT INTO video_assets(id,lesson_id,generation,state,declared_bytes,received_bytes,original_name,renditions,expires_at,created_at,updated_at)
		 VALUES($1,$2,1,'ready',1,1,'test.mp4',ARRAY['360p'],now()+interval '1 day',now(),now())`, []any{videoID, lessonID}},
	}
	for _, fixture := range fixtures {
		if _, err := pool.Exec(ctx, fixture.query, fixture.args...); err != nil {
			t.Fatal(err)
		}
	}
	metric := &integrationCounter{}
	reader := NewVideoAccessReader(pool, metric)
	course, err := reader.LoadVideoCourse(ctx, viewerID, videoID)
	if err != nil {
		t.Fatal(err)
	}
	if course != (domain.VideoCourse{OwnerID: ownerID, Status: domain.Published}) {
		t.Fatalf("course=%+v", course)
	}
	playbackReader := NewLessonPlaybackReader(pool)

	t.Run("R07 SQL facts cover the access truth table", func(t *testing.T) {
		tableMetric := &integrationCounter{}
		tableReader := NewVideoAccessReader(pool, tableMetric)
		for index, test := range accesstest.TruthTable {
			t.Run(test.Name, func(t *testing.T) {
				var courseID, moduleID, lessonID, videoID string
				slug := fmt.Sprintf("access-table-%d", index)
				publishedAt := "NULL"
				if test.Status == "published" {
					publishedAt = "now()"
				}
				var price, currency any
				if test.Paid {
					price, currency = int64(50000), "UAH"
				}
				if err := pool.QueryRow(ctx, `INSERT INTO courses(id,instructor_id,slug,title,summary,description,status,published_at,price_minor,currency)
					VALUES(gen_random_uuid(),$1,$2,'Access table','Summary','Description',$3,`+publishedAt+`,$4,$5) RETURNING id::text`, ownerID, slug, test.Status, price, currency).Scan(&courseID); err != nil {
					t.Fatal(err)
				}
				if err := pool.QueryRow(ctx, `INSERT INTO course_modules(id,course_id,title,position) VALUES(gen_random_uuid(),$1,'Module',0) RETURNING id::text`, courseID).Scan(&moduleID); err != nil {
					t.Fatal(err)
				}
				if err := pool.QueryRow(ctx, `INSERT INTO course_lessons(id,module_id,title,description,position,is_preview) VALUES(gen_random_uuid(),$1,'Lesson','',0,$2) RETURNING id::text`, moduleID, test.Preview).Scan(&lessonID); err != nil {
					t.Fatal(err)
				}
				if err := pool.QueryRow(ctx, `INSERT INTO video_assets(id,lesson_id,generation,state,declared_bytes,received_bytes,original_name,renditions,expires_at,created_at,updated_at)
					VALUES(gen_random_uuid(),$1,1,'ready',1,1,'access.mp4',ARRAY['360p'],now()+interval '1 day',now(),now()) RETURNING id::text`, lessonID).Scan(&videoID); err != nil {
					t.Fatal(err)
				}

				actorID := viewerID
				administrator := false
				if test.Role == accesstest.Owner {
					actorID = ownerID
				} else if test.Role == accesstest.Administrator {
					administrator = true
				}
				if test.Granted {
					if _, err := pool.Exec(ctx, `INSERT INTO course_grants(user_id,course_id,source,state,granted_by) VALUES($1,$2,'admin','active',$3)`, actorID, courseID, ownerID); err != nil {
						t.Fatal(err)
					}
				}
				course, err := tableReader.LoadVideoCourse(ctx, actorID, videoID)
				if err != nil {
					t.Fatal(err)
				}
				if got := course.CanBeViewedBy(domain.Actor{UserID: actorID, Administrator: administrator}); got != test.Want {
					t.Fatalf("auth_request facts=%+v actor=%+v CanBeViewedBy()=%v, want %v", course, actorID, got, test.Want)
				}
				playback, err := playbackReader.LoadLessonPlayback(ctx, actorID, lessonID)
				if err != nil {
					t.Fatal(err)
				}
				if got := playback.Course.CanBeViewedBy(domain.Actor{UserID: actorID, Administrator: administrator}); got != test.Want {
					t.Fatalf("session facts=%+v actor=%+v CanBeViewedBy()=%v, want %v", playback.Course, actorID, got, test.Want)
				}
			})
		}
	})
	t.Run("Pay23 active grant lookup uses the partial index", func(t *testing.T) {
		var measuredCourseID, measuredModuleID, measuredLessonID, measuredVideoID string
		if err := pool.QueryRow(ctx, `INSERT INTO courses(id,instructor_id,slug,title,summary,description,status,published_at,price_minor,currency)
			VALUES(gen_random_uuid(),$1,'pay23-access','Pay23','Summary','Description','published',now(),50000,'UAH') RETURNING id::text`, ownerID).Scan(&measuredCourseID); err != nil {
			t.Fatal(err)
		}
		if err := pool.QueryRow(ctx, `INSERT INTO course_modules(id,course_id,title,position) VALUES(gen_random_uuid(),$1,'Module',0) RETURNING id::text`, measuredCourseID).Scan(&measuredModuleID); err != nil {
			t.Fatal(err)
		}
		if err := pool.QueryRow(ctx, `INSERT INTO course_lessons(id,module_id,title,description,position) VALUES(gen_random_uuid(),$1,'Lesson','',0) RETURNING id::text`, measuredModuleID).Scan(&measuredLessonID); err != nil {
			t.Fatal(err)
		}
		if err := pool.QueryRow(ctx, `INSERT INTO video_assets(id,lesson_id,generation,state,declared_bytes,received_bytes,original_name,renditions,expires_at,created_at,updated_at)
			VALUES(gen_random_uuid(),$1,1,'ready',1,1,'pay23.mp4',ARRAY['360p'],now()+interval '1 day',now(),now()) RETURNING id::text`, measuredLessonID).Scan(&measuredVideoID); err != nil {
			t.Fatal(err)
		}
		if _, err := pool.Exec(ctx, `INSERT INTO course_grants(user_id,course_id,source,state,granted_by) VALUES($1,$2,'admin','active',$3)`, viewerID, measuredCourseID, ownerID); err != nil {
			t.Fatal(err)
		}
		// A tiny fixture legitimately favours a sequential scan. Populate enough unrelated active
		// grants for the planner to make the same selective index-probe choice expected in production.
		if _, err := pool.Exec(ctx, `WITH filler_users AS (
			INSERT INTO users(id,email,display_name,role,password_hash)
			SELECT gen_random_uuid(), format('pay23-%s@example.test', n), 'Pay23 filler', 'student', '$argon2id$v=19$m=65536,t=3,p=2$c3R1ZGVudC1kZW1vLXNsdA$5S1MhXX3wgDMnIKuygFRfUNVvTtRktPECxV3ZUBg0is'
			FROM generate_series(1,1000) AS n
			RETURNING id
		)
		INSERT INTO course_grants(user_id,course_id,source,state,granted_by)
		SELECT id,$1,'admin','active',$2 FROM filler_users`, measuredCourseID, ownerID); err != nil {
			t.Fatal(err)
		}
		if _, err := pool.Exec(ctx, `ANALYZE course_grants`); err != nil {
			t.Fatal(err)
		}
		rows, err := pool.Query(ctx, `EXPLAIN (ANALYZE, BUFFERS) `+loadVideoCourse, viewerID, measuredVideoID)
		if err != nil {
			t.Fatal(err)
		}
		defer rows.Close()
		var plan strings.Builder
		for rows.Next() {
			var line string
			if err := rows.Scan(&line); err != nil {
				t.Fatal(err)
			}
			plan.WriteString(line)
			plan.WriteByte('\n')
		}
		if err := rows.Err(); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(plan.String(), "course_grants_one_per_user_course") {
			t.Fatalf("loadVideoCourse plan does not use the per-user-course membership index:\n%s", plan.String())
		}
	})
	playback, err := playbackReader.LoadLessonPlayback(ctx, viewerID, lessonID)
	if err != nil || playback.VideoID != videoID || playback.ClientSequence != 0 || playback.Course != course {
		t.Fatalf("first playback=%+v err=%v", playback, err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO lesson_progress(user_id,lesson_id,position_seconds,client_seq,updated_at)
		VALUES($1,$2,12,12,now())`, viewerID, lessonID); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO video_assets(id,lesson_id,generation,state,declared_bytes,received_bytes,original_name,renditions,expires_at,created_at,updated_at)
		VALUES($1,$2,2,'ready',1,1,'new.mp4',ARRAY['360p'],now()+interval '1 day',now(),now())`, newVideoID, lessonID); err != nil {
		t.Fatal(err)
	}
	playback, err = playbackReader.LoadLessonPlayback(ctx, viewerID, lessonID)
	if err != nil || playback.VideoID != newVideoID || playback.ClientSequence != 12 {
		t.Fatalf("current playback=%+v err=%v", playback, err)
	}
	if _, err := pool.Exec(ctx, `UPDATE video_assets SET evicted_at=now(),updated_at=now() WHERE id=$1`, videoID); err != nil {
		t.Fatal(err)
	}
	if _, err := reader.LoadVideoCourse(ctx, viewerID, videoID); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("evicted video: %v", err)
	}
	if metric.value != 2 {
		t.Fatalf("metric=%d, want 2", metric.value)
	}
	if _, err := pool.Exec(ctx, `UPDATE video_assets SET evicted_at=now(),updated_at=now() WHERE id=$1`, newVideoID); err != nil {
		t.Fatal(err)
	}
	if _, err := playbackReader.LoadLessonPlayback(ctx, viewerID, lessonID); !errors.Is(err, domain.ErrVideoNotReady) {
		t.Fatalf("no ready video: %v", err)
	}
	if _, err := pool.Exec(ctx, `DELETE FROM course_lessons WHERE id=$1`, lessonID); err != nil {
		t.Fatal(err)
	}
	if _, err := playbackReader.LoadLessonPlayback(ctx, viewerID, lessonID); !errors.Is(err, domain.ErrLessonNotFound) {
		t.Fatalf("missing lesson: %v", err)
	}
}

type integrationCounter struct{ value int }

func (metric *integrationCounter) Inc() { metric.value++ }
