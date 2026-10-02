//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"courses/internal/contracts/access/accesstest"
	"courses/internal/discussion/application"
	"courses/internal/discussion/domain"
	"courses/internal/discussion/ports"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	postgrescontainer "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	discussionOwner     = "30000000-0000-4000-8000-0000000000a1"
	discussionStudent   = "30000000-0000-4000-8000-0000000000a2"
	discussionReporter  = "30000000-0000-4000-8000-0000000000a3"
	discussionReporter2 = "30000000-0000-4000-8000-0000000000a4"
	moderationCourse    = "10000000-0000-4000-8000-0000000000d2"
	moderationModule    = "20000000-0000-4000-8000-0000000000e2"
	moderationLessonA   = "40000000-0000-4000-8000-0000000000b4"
	moderationLessonB   = "40000000-0000-4000-8000-0000000000b5"
	discussionLesson    = "40000000-0000-4000-8000-0000000000b1"
	positiveLesson      = "40000000-0000-4000-8000-0000000000b2"
	cascadeLesson       = "40000000-0000-4000-8000-0000000000b3"
	discussionAsset     = "50000000-0000-4000-8000-0000000000c1"
	positiveAsset       = "50000000-0000-4000-8000-0000000000c3"
)

func TestDiscussionPostgres(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	container, err := postgrescontainer.Run(ctx, "postgres:17.5",
		postgrescontainer.WithDatabase("discussion_test"),
		postgrescontainer.WithUsername("courses"),
		postgrescontainer.WithPassword("ephemeral-test-only"),
		testcontainers.WithWaitStrategy(wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(3*time.Minute)),
	)
	if err != nil {
		t.Fatalf("PostgreSQL integration requires Docker: %v", err)
	}
	t.Cleanup(func() { _ = container.Terminate(context.Background()) })
	connection, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	databaseConn, err := sql.Open("pgx", connection)
	if err != nil {
		t.Fatal(err)
	}
	defer databaseConn.Close()
	migrations := filepath.Join("..", "..", "..", "..", "migrations")
	provider, err := goose.NewProvider(goose.DialectPostgres, databaseConn, os.DirFS(migrations), goose.WithDisableGlobalRegistry(true))
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
	seedDiscussion(t, pool)
	repository := NewRepository(pool)

	t.Run("I02 course-wide moderation repository", func(t *testing.T) {
		ctx := context.Background()
		t.Run("ListByCourse spans lessons and paginates all comments", func(t *testing.T) {
			rootA := insertModerationRoot(t, repository, moderationLessonA, discussionStudent, "course root")
			rootB := insertModerationRoot(t, repository, moderationLessonB, discussionStudent, "other lesson root")
			reply := insertModerationReply(t, repository, moderationLessonA, rootA.Comment.ID, discussionReporter, "course reply")
			base := time.Now().UTC().Add(-time.Minute)
			for index, id := range []string{rootA.Comment.ID, rootB.Comment.ID, reply.Comment.ID} {
				if _, err := pool.Exec(ctx, `UPDATE discussion_comments SET created_at=$2 WHERE id=$1::uuid`, id, base.Add(time.Duration(index)*time.Second)); err != nil {
					t.Fatal(err)
				}
			}
			page, err := repository.ListByCourse(ctx, moderationCourse, "", 2)
			if err != nil || len(page.Comments) != 2 || page.NextCursor == "" {
				t.Fatalf("page=%+v err=%v", page, err)
			}
			if page.Comments[0].Comment.ID != reply.Comment.ID || page.Comments[1].Comment.ID != rootB.Comment.ID {
				t.Fatalf("first page=%+v", page.Comments)
			}
			if page.Comments[0].LessonTitle != "A" || page.Comments[1].LessonTitle != "B" {
				t.Fatalf("lesson titles=%q,%q want A,B", page.Comments[0].LessonTitle, page.Comments[1].LessonTitle)
			}
			page, err = repository.ListByCourse(ctx, moderationCourse, page.NextCursor, 2)
			if err != nil || len(page.Comments) != 1 || page.Comments[0].Comment.ID != rootA.Comment.ID {
				t.Fatalf("second page=%+v err=%v", page.Comments, err)
			}
			if page.Comments[0].LessonTitle != "A" {
				t.Fatalf("lesson title=%q want A", page.Comments[0].LessonTitle)
			}
		})

		t.Run("InsertReport is idempotent", func(t *testing.T) {
			comment := insertModerationRoot(t, repository, moderationLessonA, discussionStudent, "duplicate report")
			report := domain.Report{CommentID: comment.Comment.ID, ReporterID: discussionReporter}
			if err := repository.InsertReport(ctx, report, time.Now().UTC()); err != nil {
				t.Fatal(err)
			}
			if err := repository.InsertReport(ctx, report, time.Now().UTC().Add(time.Second)); err != nil {
				t.Fatal(err)
			}
			var count int
			if err := pool.QueryRow(ctx, `SELECT count(*) FROM discussion_comment_reports WHERE comment_id=$1::uuid`, comment.Comment.ID).Scan(&count); err != nil {
				t.Fatal(err)
			}
			if count != 1 {
				t.Fatalf("reports=%d want 1", count)
			}
		})

		t.Run("W10 reporting a removed comment is a no-op", func(t *testing.T) {
			comment := insertModerationRoot(t, repository, moderationLessonA, discussionStudent, "removed before report")
			at := time.Now().UTC()
			if err := repository.Remove(ctx, comment.Comment.ID, discussionOwner, at); err != nil {
				t.Fatal(err)
			}
			var before int
			if err := pool.QueryRow(ctx, `SELECT count(*) FROM discussion_comment_reports WHERE comment_id=$1::uuid`, comment.Comment.ID).Scan(&before); err != nil {
				t.Fatal(err)
			}
			service := application.New(repository, func() time.Time { return at.Add(time.Second) })
			if err := service.ReportComment(ctx, domain.Actor{UserID: discussionReporter}, comment.Comment.ID); err != nil {
				t.Fatal(err)
			}
			var after int
			if err := pool.QueryRow(ctx, `SELECT count(*) FROM discussion_comment_reports WHERE comment_id=$1::uuid`, comment.Comment.ID).Scan(&after); err != nil {
				t.Fatal(err)
			}
			if after != before {
				t.Fatalf("reports before=%d after=%d", before, after)
			}
			queue, err := repository.ListQueue(ctx, moderationCourse, "", 50)
			if err != nil {
				t.Fatal(err)
			}
			for _, item := range queue.Items {
				if item.Comment.ID == comment.Comment.ID {
					t.Fatalf("removed comment is queued: %+v", item)
				}
			}
		})

		t.Run("queue ordering dismiss removal and reactivation", func(t *testing.T) {
			// Isolate the report-driven queue from earlier subtests that share
			// moderationCourse (e.g. the idempotent-report comment stays open).
			if _, err := pool.Exec(ctx, `DELETE FROM discussion_comment_reports`); err != nil {
				t.Fatal(err)
			}
			most := insertModerationRoot(t, repository, moderationLessonA, discussionStudent, "most reports")
			one := insertModerationRoot(t, repository, moderationLessonB, discussionStudent, "one report")
			removed := insertModerationRoot(t, repository, moderationLessonA, discussionStudent, "removed report")
			at := time.Now().UTC()
			for _, reporter := range []string{discussionReporter, discussionReporter2} {
				if err := repository.InsertReport(ctx, domain.Report{CommentID: most.Comment.ID, ReporterID: reporter}, at); err != nil {
					t.Fatal(err)
				}
			}
			if err := repository.InsertReport(ctx, domain.Report{CommentID: one.Comment.ID, ReporterID: discussionReporter}, at.Add(time.Second)); err != nil {
				t.Fatal(err)
			}
			if err := repository.InsertReport(ctx, domain.Report{CommentID: removed.Comment.ID, ReporterID: discussionReporter}, at); err != nil {
				t.Fatal(err)
			}
			if _, err := pool.Exec(ctx, `UPDATE discussion_comments SET body=NULL, removed_at=$2 WHERE id=$1::uuid`, removed.Comment.ID, at); err != nil {
				t.Fatal(err)
			}
			queue, err := repository.ListQueue(ctx, moderationCourse, "", 50)
			if err != nil || len(queue.Items) != 2 || queue.Items[0].Comment.ID != most.Comment.ID || queue.Items[0].OpenReports != 2 || queue.Items[1].Comment.ID != one.Comment.ID {
				t.Fatalf("queue=%+v err=%v", queue, err)
			}
			if err := repository.DismissReports(ctx, most.Comment.ID, discussionOwner, at.Add(2*time.Second)); err != nil {
				t.Fatal(err)
			}
			queue, err = repository.ListQueue(ctx, moderationCourse, "", 50)
			if err != nil || len(queue.Items) != 1 || queue.Items[0].Comment.ID != one.Comment.ID {
				t.Fatalf("after dismiss queue=%+v err=%v", queue, err)
			}
			if err := repository.InsertReport(ctx, domain.Report{CommentID: most.Comment.ID, ReporterID: discussionStudent}, at.Add(3*time.Second)); err != nil {
				t.Fatal(err)
			}
			queue, err = repository.ListQueue(ctx, moderationCourse, "", 50)
			if err != nil || len(queue.Items) != 2 || queue.Items[0].Comment.ID != most.Comment.ID {
				t.Fatalf("reactivated queue=%+v err=%v", queue, err)
			}
		})

		t.Run("W26 remove and dismiss converge in either order", func(t *testing.T) {
			cases := []struct {
				name        string
				removeFirst bool
			}{
				{name: "remove then dismiss", removeFirst: true},
				{name: "dismiss then remove"},
			}
			for index, test := range cases {
				t.Run(test.name, func(t *testing.T) {
					comment := insertModerationRoot(t, repository, moderationLessonA, discussionStudent, "remove dismiss order")
					at := time.Now().UTC().Add(time.Duration(index) * time.Second)
					if err := repository.InsertReport(ctx, domain.Report{CommentID: comment.Comment.ID, ReporterID: discussionReporter}, at); err != nil {
						t.Fatal(err)
					}
					before, err := repository.ListQueue(ctx, moderationCourse, "", 50)
					if err != nil {
						t.Fatal(err)
					}
					queued := false
					for _, item := range before.Items {
						if item.Comment.ID == comment.Comment.ID {
							queued = true
							break
						}
					}
					if !queued {
						t.Fatalf("reported comment is absent before %s", test.name)
					}
					remove := func() error { return repository.Remove(ctx, comment.Comment.ID, discussionOwner, at.Add(time.Second)) }
					dismiss := func() error {
						return repository.DismissReports(ctx, comment.Comment.ID, discussionOwner, at.Add(2*time.Second))
					}
					if test.removeFirst {
						if err := remove(); err != nil {
							t.Fatal(err)
						}
						if err := dismiss(); err != nil {
							t.Fatal(err)
						}
					} else {
						if err := dismiss(); err != nil {
							t.Fatal(err)
						}
						if err := remove(); err != nil {
							t.Fatal(err)
						}
					}
					var removed bool
					if err := pool.QueryRow(ctx, `SELECT removed_at IS NOT NULL FROM discussion_comments WHERE id=$1::uuid`, comment.Comment.ID).Scan(&removed); err != nil {
						t.Fatal(err)
					}
					var openReports int
					if err := pool.QueryRow(ctx, `SELECT count(*) FROM discussion_comment_reports WHERE comment_id=$1::uuid AND dismissed_at IS NULL`, comment.Comment.ID).Scan(&openReports); err != nil {
						t.Fatal(err)
					}
					if !removed || openReports != 0 {
						t.Fatalf("removed=%v openReports=%d", removed, openReports)
					}
					queue, err := repository.ListQueue(ctx, moderationCourse, "", 50)
					if err != nil {
						t.Fatal(err)
					}
					for _, item := range queue.Items {
						if item.Comment.ID == comment.Comment.ID {
							t.Fatalf("comment is queued after %s: %+v", test.name, item)
						}
					}
				})
			}
		})
	})

	t.Run("K07 SQL facts cover access truth table", func(t *testing.T) {
		for index, truth := range accesstest.TruthTable {
			var courseID, moduleID, lessonID string
			publishedAt := "NULL"
			if truth.Status == "published" {
				publishedAt = "now()"
			}
			var price, currency any
			if truth.Paid {
				price, currency = int64(50000), "UAH"
			}
			if err := pool.QueryRow(ctx, `INSERT INTO courses(id,slug,title,summary,description,status,instructor_id,published_at,price_minor,currency) VALUES(gen_random_uuid(),$1,'Access','S','D',$2,$3,`+publishedAt+`,$4,$5) RETURNING id::text`, "access-"+string(rune('a'+index)), truth.Status, discussionOwner, price, currency).Scan(&courseID); err != nil {
				t.Fatal(err)
			}
			if err := pool.QueryRow(ctx, `INSERT INTO course_modules(id,course_id,title,position) VALUES(gen_random_uuid(),$1,'M',0) RETURNING id::text`, courseID).Scan(&moduleID); err != nil {
				t.Fatal(err)
			}
			if err := pool.QueryRow(ctx, `INSERT INTO course_lessons(id,module_id,title,description,position,is_preview) VALUES(gen_random_uuid(),$1,'L','',0,$2) RETURNING id::text`, moduleID, truth.Preview).Scan(&lessonID); err != nil {
				t.Fatal(err)
			}
			actor := domain.Actor{UserID: discussionStudent, Administrator: truth.Role == accesstest.Administrator}
			if truth.Role == accesstest.Owner {
				actor.UserID = discussionOwner
			}
			if truth.Granted {
				if _, err := pool.Exec(ctx, `INSERT INTO course_grants(user_id,course_id,source,state,granted_by) VALUES($1,$2,'admin','active',$3)`, actor.UserID, courseID, discussionOwner); err != nil {
					t.Fatal(err)
				}
			}
			facts, err := repository.LessonFacts(ctx, actor.UserID, lessonID)
			if err != nil {
				t.Fatal(err)
			}
			if got := facts.Course.CanBeViewedBy(actor); got != truth.Want {
				t.Fatalf("%s: access=%v want=%v", truth.Name, got, truth.Want)
			}
			var commentID string
			if err := pool.QueryRow(ctx, `INSERT INTO discussion_comments(lesson_id,author_id,body) VALUES($1,$2,'Access row') RETURNING id::text`, lessonID, discussionOwner).Scan(&commentID); err != nil {
				t.Fatal(err)
			}
			record, err := repository.Get(ctx, actor.UserID, commentID)
			if err != nil {
				t.Fatal(err)
			}
			if got := record.Course.CanBeViewedBy(actor); got != truth.Want {
				t.Fatalf("%s comment facts: access=%v want=%v", truth.Name, got, truth.Want)
			}
		}
	})
	t.Run("K08 constraints and keyset indexes", func(t *testing.T) {
		var lessonFK, authorFK, parentFK string
		rows, err := pool.Query(ctx, `SELECT conname FROM pg_constraint WHERE conrelid='discussion_comments'::regclass AND contype='f' ORDER BY conname`)
		if err != nil {
			t.Fatal(err)
		}
		for rows.Next() {
			var name string
			if err := rows.Scan(&name); err != nil {
				t.Fatal(err)
			}
			switch {
			case strings.Contains(name, "lesson"):
				lessonFK = name
			case strings.Contains(name, "author"):
				authorFK = name
			case strings.Contains(name, "parent"):
				parentFK = name
			}
		}
		rows.Close()
		if lessonFK == "" || authorFK == "" || parentFK == "" {
			t.Fatalf("FKs lesson=%q author=%q parent=%q", lessonFK, authorFK, parentFK)
		}
		var rootDef, replyDef string
		if err := pool.QueryRow(ctx, `SELECT indexdef FROM pg_indexes WHERE indexname='discussion_comments_roots_keyset'`).Scan(&rootDef); err != nil {
			t.Fatal(err)
		}
		if err := pool.QueryRow(ctx, `SELECT indexdef FROM pg_indexes WHERE indexname='discussion_comments_replies_keyset'`).Scan(&replyDef); err != nil {
			t.Fatal(err)
		}
		for _, part := range []string{"lesson_id", "created_at DESC", "id DESC", "WHERE (parent_id IS NULL)"} {
			if !strings.Contains(rootDef, part) {
				t.Errorf("root index missing %q: %s", part, rootDef)
			}
		}
		for _, part := range []string{"parent_id", "created_at", "id"} {
			if !strings.Contains(replyDef, part) {
				t.Errorf("reply index missing %q: %s", part, replyDef)
			}
		}
		if _, err := pool.Exec(ctx, `INSERT INTO discussion_comments(lesson_id,author_id,body) VALUES($1,$2,'')`, discussionLesson, discussionStudent); err == nil {
			t.Fatal("empty body passed the database check")
		}
		if _, err := pool.Exec(ctx, `INSERT INTO discussion_comments(lesson_id,author_id,body,position_ms) VALUES($1,$2,'x',-1)`, discussionLesson, discussionStudent); err == nil {
			t.Fatal("negative position passed the database check")
		}
		for name, statement := range map[string]string{
			"live comment with NULL body":         `INSERT INTO discussion_comments(lesson_id,author_id,body,removed_at) VALUES($1,$2,NULL,NULL)`,
			"removed comment with non-empty body": `INSERT INTO discussion_comments(lesson_id,author_id,body,removed_at) VALUES($1,$2,'still here',now())`,
		} {
			t.Run(name, func(t *testing.T) {
				_, err := pool.Exec(ctx, statement, discussionLesson, discussionStudent)
				if err == nil {
					t.Fatal("invalid body/removal state passed the database check")
				}
				var postgresError *pgconn.PgError
				if !errors.As(err, &postgresError) || postgresError.Code != "23514" {
					t.Fatalf("error=%v want CHECK violation SQLSTATE 23514", err)
				}
			})
		}
		var cascadeRoot string
		if err := pool.QueryRow(ctx, `INSERT INTO discussion_comments(lesson_id,author_id,body) VALUES($1,$2,'root') RETURNING id::text`, cascadeLesson, discussionStudent).Scan(&cascadeRoot); err != nil {
			t.Fatal(err)
		}
		if _, err := pool.Exec(ctx, `INSERT INTO discussion_comments(lesson_id,author_id,parent_id,body) VALUES($1,$2,$3,'reply')`, cascadeLesson, discussionStudent, cascadeRoot); err != nil {
			t.Fatal(err)
		}
		if _, err := pool.Exec(ctx, `DELETE FROM course_lessons WHERE id=$1`, cascadeLesson); err != nil {
			t.Fatal(err)
		}
		var remaining int
		if err := pool.QueryRow(ctx, `SELECT count(*) FROM discussion_comments WHERE lesson_id=$1`, cascadeLesson).Scan(&remaining); err != nil {
			t.Fatal(err)
		}
		if remaining != 0 {
			t.Fatalf("lesson cascade left %d discussion rows", remaining)
		}
	})
	t.Run("K09 roots cursor replies and removed shape", func(t *testing.T) {
		first := insertRoot(t, pool, repository, "first")
		second := insertRoot(t, pool, repository, "second")
		third := insertRoot(t, pool, repository, "third")
		base := time.Now().UTC()
		for index, id := range []string{first.Comment.ID, second.Comment.ID, third.Comment.ID} {
			if _, err := pool.Exec(ctx, `UPDATE discussion_comments SET created_at=$2 WHERE id=$1::text::uuid`, id, base.Add(time.Duration(index)*time.Second)); err != nil {
				t.Fatal(err)
			}
		}
		insertReply(t, repository, first.Comment.ID, "reply")
		if err := repository.Remove(ctx, second.Comment.ID, discussionStudent, time.Now()); err != nil {
			t.Fatal(err)
		}
		page, err := repository.List(ctx, discussionLesson, "", 1)
		if err != nil {
			t.Fatal(err)
		}
		if len(page.Threads) != 1 || page.Threads[0].Root.Comment.ID != third.Comment.ID {
			t.Fatalf("page=%+v", page)
		}
		if page.NextCursor == "" {
			t.Fatal("expected next cursor")
		}
		page, err = repository.List(ctx, discussionLesson, page.NextCursor, 1)
		if err != nil {
			t.Fatal(err)
		}
		if len(page.Threads) != 1 || page.Threads[0].Root.Comment.ID != first.Comment.ID || len(page.Threads[0].Replies) != 1 {
			t.Fatalf("cursor page=%+v", page)
		}
		if err := repository.Remove(ctx, first.Comment.ID, discussionStudent, time.Now()); err != nil {
			t.Fatal(err)
		}
		page, err = repository.List(ctx, discussionLesson, "", 50)
		if err != nil {
			t.Fatal(err)
		}
		if len(page.Threads) != 2 {
			t.Fatalf("removed root should remain with reply: %+v", page)
		}
		var foundRemoved bool
		for _, thread := range page.Threads {
			if thread.Root.Comment.ID == first.Comment.ID {
				foundRemoved = true
				if thread.Root.Comment.Body != "" || thread.Root.Comment.RemovedAt == nil || thread.Replies[0].Comment.Body != "reply" {
					t.Fatalf("removed thread=%+v", thread)
				}
			}
		}
		if !foundRemoved {
			t.Fatal("removed root with reply missing")
		}
	})
	t.Run("K09a filtered selector and serialized cap", func(t *testing.T) {
		positiveFacts, err := repository.LessonFacts(ctx, discussionStudent, positiveLesson)
		if err != nil {
			t.Fatal(err)
		}
		if positiveFacts.AssetID != positiveAsset || positiveFacts.Generation != 1 || positiveFacts.DurationSecs == nil || *positiveFacts.DurationSecs != 45 {
			t.Fatalf("positive selected asset=%s generation=%d duration=%v", positiveFacts.AssetID, positiveFacts.Generation, positiveFacts.DurationSecs)
		}
		facts, err := repository.LessonFacts(ctx, discussionStudent, discussionLesson)
		if err != nil {
			t.Fatal(err)
		}
		if facts.AssetID != "" || facts.Generation != 0 {
			t.Fatalf("selected asset=%s generation=%d", facts.AssetID, facts.Generation)
		}
		t.Run("two connections prove the root lock closes the race", func(t *testing.T) {
			root := insertRoot(t, pool, repository, "locked-cap-root")
			for i := 0; i < domain.MaximumReplyCount-1; i++ {
				if _, err := pool.Exec(ctx, `INSERT INTO discussion_comments(lesson_id,author_id,parent_id,body) VALUES($1,$2,$3,$4)`, discussionLesson, discussionStudent, root.Comment.ID, "reply"); err != nil {
					t.Fatal(err)
				}
			}

			lockedConnection, err := pool.Acquire(ctx)
			if err != nil {
				t.Fatal(err)
			}
			defer lockedConnection.Release()
			transaction, err := lockedConnection.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = transaction.Rollback(context.Background()) }()
			var lockedID string
			if err := transaction.QueryRow(ctx, `SELECT id::text FROM discussion_comments WHERE id=$1::text::uuid AND parent_id IS NULL FOR UPDATE`, root.Comment.ID).Scan(&lockedID); err != nil {
				t.Fatal(err)
			}

			result := make(chan error, 1)
			goroutineContext, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			go func() {
				_, err := repository.InsertReply(goroutineContext, domain.Comment{LessonID: discussionLesson, AuthorID: discussionStudent, ParentID: &root.Comment.ID, Body: "concurrent reply"})
				result <- err
			}()

			select {
			case err := <-result:
				t.Fatalf("InsertReply returned before tx1 commit: %v", err)
			case <-time.After(300 * time.Millisecond):
			}
			if _, err := transaction.Exec(ctx, `INSERT INTO discussion_comments(lesson_id,author_id,parent_id,body) VALUES($1,$2,$3,$4)`, discussionLesson, discussionStudent, root.Comment.ID, "tx1 reply"); err != nil {
				t.Fatal(err)
			}
			if err := transaction.Commit(ctx); err != nil {
				t.Fatal(err)
			}
			select {
			case err := <-result:
				if !errors.Is(err, domain.ErrThreadFull) {
					t.Fatalf("InsertReply error=%v want=%v", err, domain.ErrThreadFull)
				}
			case <-time.After(5 * time.Second):
				t.Fatal("InsertReply did not return after tx1 commit")
			}
		})
		root := insertRoot(t, pool, repository, "cap-root")
		for i := 0; i < domain.MaximumReplyCount; i++ {
			_, err := pool.Exec(ctx, `INSERT INTO discussion_comments(lesson_id,author_id,parent_id,body) VALUES($1,$2,$3,$4)`, discussionLesson, discussionStudent, root.Comment.ID, "reply")
			if err != nil {
				t.Fatal(err)
			}
		}
		if _, err := repository.InsertReply(ctx, domain.Comment{LessonID: discussionLesson, AuthorID: discussionStudent, ParentID: &root.Comment.ID, Body: "overflow"}); err != domain.ErrThreadFull {
			t.Fatalf("cap error=%v", err)
		}
	})
}

func seedDiscussion(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()
	statements := []string{
		`INSERT INTO users(id,email,display_name,role,password_hash) VALUES ('` + discussionOwner + `','owner@discussion.test','Owner','instructor','$argon2id$v=19$m=65536,t=3,p=2$c3R1ZGVudC1kZW1vLXNsdA$5S1MhXX3wgDMnIKuygFRfUNVvTtRktPECxV3ZUBg0is'),('` + discussionStudent + `','student@discussion.test','Student','student','$argon2id$v=19$m=65536,t=3,p=2$c3R1ZGVudC1kZW1vLXNsdA$5S1MhXX3wgDMnIKuygFRfUNVvTtRktPECxV3ZUBg0is'),('` + discussionReporter + `','reporter@discussion.test','Reporter','student','$argon2id$v=19$m=65536,t=3,p=2$c3R1ZGVudC1kZW1vLXNsdA$5S1MhXX3wgDMnIKuygFRfUNVvTtRktPECxV3ZUBg0is'),('` + discussionReporter2 + `','reporter2@discussion.test','Reporter 2','student','$argon2id$v=19$m=65536,t=3,p=2$c3R1ZGVudC1kZW1vLXNsdA$5S1MhXX3wgDMnIKuygFRfUNVvTtRktPECxV3ZUBg0is')`,
		`INSERT INTO courses(id,slug,title,summary,description,status,instructor_id,published_at) VALUES ('10000000-0000-4000-8000-0000000000d1','discussion','D','S','D','published','` + discussionOwner + `',now())`,
		`INSERT INTO course_modules(id,course_id,title,position) VALUES ('20000000-0000-4000-8000-0000000000e1','10000000-0000-4000-8000-0000000000d1','M',0)`,
		`INSERT INTO courses(id,slug,title,summary,description,status,instructor_id,published_at) VALUES ('` + moderationCourse + `','moderation','Moderation','S','D','published','` + discussionOwner + `',now())`,
		`INSERT INTO course_modules(id,course_id,title,position) VALUES ('` + moderationModule + `','` + moderationCourse + `','M',0)`,
		`INSERT INTO course_lessons(id,module_id,title,description,position) VALUES ('` + discussionLesson + `','20000000-0000-4000-8000-0000000000e1','L','',0),('` + positiveLesson + `','20000000-0000-4000-8000-0000000000e1','Positive','',1),('` + cascadeLesson + `','20000000-0000-4000-8000-0000000000e1','Cascade','',2)`,
		`INSERT INTO course_lessons(id,module_id,title,description,position) VALUES ('` + moderationLessonA + `','` + moderationModule + `','A','',0),('` + moderationLessonB + `','` + moderationModule + `','B','',1)`,
		`INSERT INTO video_assets(id,lesson_id,generation,state,declared_bytes,original_name,renditions,expires_at,created_at,updated_at,duration_seconds) VALUES ('` + discussionAsset + `','` + discussionLesson + `',1,'ready',10,'a.mp4','{720p}',now()+interval '1 day',now(),now(),120),('50000000-0000-4000-8000-0000000000c2','` + discussionLesson + `',2,'processing',10,'b.mp4','{}',now()+interval '1 day',now(),now(),NULL),('` + positiveAsset + `','` + positiveLesson + `',1,'ready',10,'positive-a.mp4','{720p}',now()+interval '1 day',now(),now(),45),('50000000-0000-4000-8000-0000000000c4','` + positiveLesson + `',2,'ready',10,'positive-b.mp4','{720p}',now()+interval '1 day',now(),now(),90),('50000000-0000-4000-8000-0000000000c5','` + positiveLesson + `',3,'processing',10,'positive-c.mp4','{}',now()+interval '1 day',now(),now(),NULL)`,
	}
	for _, statement := range statements {
		if _, err := pool.Exec(ctx, statement); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := pool.Exec(ctx, `UPDATE video_assets SET evicted_at=now() WHERE id = ANY($1::uuid[])`, []string{discussionAsset, "50000000-0000-4000-8000-0000000000c4"}); err != nil {
		t.Fatal(err)
	}
}

func insertModerationRoot(t *testing.T, repository *Repository, lessonID, authorID, body string) ports.CommentRecord {
	t.Helper()
	result, err := repository.InsertRoot(context.Background(), domain.Comment{LessonID: lessonID, AuthorID: authorID, Body: body})
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func insertModerationReply(t *testing.T, repository *Repository, lessonID, parentID, authorID, body string) ports.CommentRecord {
	t.Helper()
	result, err := repository.InsertReply(context.Background(), domain.Comment{LessonID: lessonID, AuthorID: authorID, ParentID: &parentID, Body: body})
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func insertRoot(t *testing.T, pool *pgxpool.Pool, repository *Repository, body string) ports.CommentRecord {
	t.Helper()
	result, err := repository.InsertRoot(context.Background(), domain.Comment{LessonID: discussionLesson, AuthorID: discussionStudent, Body: body})
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func insertReply(t *testing.T, repository *Repository, parentID, body string) ports.CommentRecord {
	t.Helper()
	result, err := repository.InsertReply(context.Background(), domain.Comment{LessonID: discussionLesson, AuthorID: discussionStudent, ParentID: &parentID, Body: body})
	if err != nil {
		t.Fatal(err)
	}
	return result
}
