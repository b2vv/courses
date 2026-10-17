//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"courses/internal/contracts/access"
	"courses/internal/contracts/access/accesstest"
	"courses/internal/discussion/application"
	"courses/internal/discussion/domain"
	"courses/internal/discussion/ports"
	containers "courses/internal/testsupport/containers"
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
	container, err := containers.Start(ctx, func(ctx context.Context) (*postgrescontainer.PostgresContainer, error) {
		return postgrescontainer.Run(ctx, "postgres:17.5",
			postgrescontainer.WithDatabase("discussion_test"),
			postgrescontainer.WithUsername("courses"),
			postgrescontainer.WithPassword("ephemeral-test-only"),
			testcontainers.WithWaitStrategy(wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(3*time.Minute)),
		)
	})
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

	t.Run("L5 participant reports use isolated course membership and admin queue", func(t *testing.T) {
		// This fixture intentionally has no lesson and does not reuse the globally counted d1/b1 fixtures.
		// Person reports only need a course and membership, so each identity/course is generated here.
		createUser := func(t *testing.T, name string) string {
			t.Helper()
			var id string
			if err := pool.QueryRow(ctx, `INSERT INTO users(id,email,display_name,role,password_hash)
				VALUES(gen_random_uuid(),concat('l5-',gen_random_uuid()::text,'@discussion.test'),$1,'student','$argon2id$v=19$m=65536,t=3,p=2$c3R1ZGVudC1kZW1vLXNsdA$5S1MhXX3wgDMnIKuygFRfUNVvTtRktPECxV3ZUBg0is') RETURNING id::text`, name).Scan(&id); err != nil {
				t.Fatal(err)
			}
			return id
		}
		owner, active, blocked, none, admin := createUser(t, "L5 owner"), createUser(t, "L5 active"), createUser(t, "L5 blocked"), createUser(t, "L5 none"), createUser(t, "L5 admin")
		var course string
		if err := pool.QueryRow(ctx, `INSERT INTO courses(id,slug,title,summary,description,status,instructor_id,published_at)
			VALUES(gen_random_uuid(),concat('l5-',gen_random_uuid()::text),'L5 isolated reports','Summary','Description','published',$1::uuid,now()) RETURNING id::text`, owner).Scan(&course); err != nil {
			t.Fatal(err)
		}
		for _, membership := range []struct{ user, state string }{{active, "active"}, {blocked, "blocked"}} {
			if _, err := pool.Exec(ctx, `INSERT INTO course_grants(user_id,course_id,source,state,granted_by) VALUES($1::uuid,$2::uuid,'admin',$3,$4::uuid)`, membership.user, course, membership.state, owner); err != nil {
				t.Fatal(err)
			}
		}
		service := application.New(repository, time.Now)
		member := application.Actor{UserID: active, EmailVerified: true}
		if report, err := service.ReportLecturer(ctx, member, course, "lecturer report"); err != nil || report.Kind != domain.ReportOnLecturer || report.TargetID != owner {
			t.Fatalf("lecturer report=%+v err=%v", report, err)
		}
		if _, err := service.ReportLecturer(ctx, member, course, "duplicate"); !errors.Is(err, domain.ErrReportDuplicate) {
			t.Fatalf("duplicate=%v", err)
		}
		if _, err := service.ReportLecturer(ctx, application.Actor{UserID: blocked, EmailVerified: true}, course, "blocked is participant"); err != nil {
			t.Fatalf("blocked=%v", err)
		}
		if _, err := service.ReportLecturer(ctx, application.Actor{UserID: none, EmailVerified: true}, course, "none"); !errors.Is(err, domain.ErrReportForbidden) {
			t.Fatalf("none=%v", err)
		}
		if report, err := service.ReportStudent(ctx, application.Actor{UserID: owner, EmailVerified: true}, course, active, "student report"); err != nil || report.Kind != domain.ReportOnStudent {
			t.Fatalf("student report=%+v err=%v", report, err)
		}
		if _, err := service.ReportStudent(ctx, application.Actor{UserID: owner, EmailVerified: true}, course, none, "not participant"); !errors.Is(err, domain.ErrReportForbidden) {
			t.Fatalf("non-member target=%v", err)
		}
		if _, err := service.ReportQueue(ctx, application.Actor{UserID: admin}, "", "open"); !errors.Is(err, domain.ErrReportForbidden) {
			t.Fatalf("non-admin queue=%v", err)
		}
		if err := service.ResolveReport(ctx, application.Actor{UserID: admin}, "00000000-0000-4000-8000-000000000001"); !errors.Is(err, domain.ErrReportForbidden) {
			t.Fatalf("non-admin resolve=%v", err)
		}
		queue, err := service.ReportQueue(ctx, application.Actor{UserID: admin, Administrator: true}, "", "open")
		if err != nil || len(queue.Items) != 3 {
			t.Fatalf("queue=%+v err=%v", queue, err)
		}
		var lecturerID string
		for _, item := range queue.Items {
			if item.Report.Kind == domain.ReportOnLecturer && item.Report.ReporterID == active {
				lecturerID = item.Report.ID
			}
			if item.ReporterName == nil || item.TargetName == nil {
				t.Fatalf("names absent: %+v", item)
			}
		}
		if lecturerID == "" {
			t.Fatal("active lecturer report absent")
		}
		if err := service.ResolveReport(ctx, application.Actor{UserID: admin, Administrator: true}, lecturerID); err != nil {
			t.Fatal(err)
		}
		if err := service.ResolveReport(ctx, application.Actor{UserID: admin, Administrator: true}, lecturerID); err != nil {
			t.Fatalf("idempotent resolve=%v", err)
		}
		queue, err = service.ReportQueue(ctx, application.Actor{UserID: admin, Administrator: true}, "", "open")
		if err != nil || len(queue.Items) != 2 {
			t.Fatalf("resolved queue=%+v err=%v", queue, err)
		}
		// The resolved filter surfaces exactly the one just resolved; "all" (empty filter)
		// spans both states, so the resolved report rejoins the two still-open ones.
		resolved, err := service.ReportQueue(ctx, application.Actor{UserID: admin, Administrator: true}, "", "resolved")
		if err != nil || len(resolved.Items) != 1 || resolved.Items[0].Report.ID != lecturerID {
			t.Fatalf("resolved-only queue=%+v err=%v", resolved, err)
		}
		all, err := service.ReportQueue(ctx, application.Actor{UserID: admin, Administrator: true}, "", "")
		if err != nil || len(all.Items) != 3 {
			t.Fatalf("all-states queue=%+v err=%v", all, err)
		}
	})

	t.Run("L4b communication gate uses seeded membership states through the application service", func(t *testing.T) {
		service := application.New(repository, time.Now)
		// Isolated course/lesson so this subtest's comments do not land on the shared discussion
		// lesson that K09 counts roots on (the globally-counted-state rule).
		var l4bCourse, l4bModule, l4bLesson string
		if err := pool.QueryRow(ctx, `INSERT INTO courses(id,slug,title,summary,description,status,instructor_id,published_at) VALUES(gen_random_uuid(), concat('l4b-', gen_random_uuid()::text), 'L4b gate', 'Summary', 'Description', 'published', $1::uuid, now()) RETURNING id::text`, discussionOwner).Scan(&l4bCourse); err != nil {
			t.Fatal(err)
		}
		if err := pool.QueryRow(ctx, `INSERT INTO course_modules(id,course_id,title,position) VALUES(gen_random_uuid(), $1::uuid, 'Module', 0) RETURNING id::text`, l4bCourse).Scan(&l4bModule); err != nil {
			t.Fatal(err)
		}
		if err := pool.QueryRow(ctx, `INSERT INTO course_lessons(id,module_id,title,position) VALUES(gen_random_uuid(), $1::uuid, 'Lesson', 0) RETURNING id::text`, l4bModule).Scan(&l4bLesson); err != nil {
			t.Fatal(err)
		}
		createMember := func(t *testing.T, state string) string {
			t.Helper()
			var userID string
			if err := pool.QueryRow(ctx, `INSERT INTO users(id,email,display_name,role,password_hash)
				VALUES(gen_random_uuid(), concat('l4b-', gen_random_uuid()::text, '@discussion.test'), 'L4b member', 'student', '$argon2id$v=19$m=65536,t=3,p=2$c3R1ZGVudC1kZW1vLXNsdA$5S1MhXX3wgDMnIKuygFRfUNVvTtRktPECxV3ZUBg0is')
				RETURNING id::text`).Scan(&userID); err != nil {
				t.Fatal(err)
			}
			if _, err := pool.Exec(ctx, `INSERT INTO course_grants(user_id,course_id,source,state,granted_by)
				VALUES($1::uuid,$2::uuid,'admin',$3,$4::uuid)`, userID, l4bCourse, state, discussionOwner); err != nil {
				t.Fatal(err)
			}
			return userID
		}
		active := createMember(t, "active")
		blocked := createMember(t, "blocked")
		revoked := createMember(t, "revoked")

		if _, err := service.Publish(ctx, application.Actor{UserID: active, EmailVerified: true}, application.PublishRequest{LessonID: l4bLesson, Body: "active member comment"}); err != nil {
			t.Fatalf("active publish: %v", err)
		}
		if _, err := service.Publish(ctx, application.Actor{UserID: blocked, EmailVerified: true}, application.PublishRequest{LessonID: l4bLesson, Body: "blocked member comment"}); !errors.Is(err, domain.ErrEnrollmentBlocked) {
			t.Fatalf("blocked publish: %v", err)
		}
		if _, err := service.Publish(ctx, application.Actor{UserID: revoked, EmailVerified: true}, application.PublishRequest{LessonID: l4bLesson, Body: "revoked member comment"}); !errors.Is(err, domain.ErrEnrollmentRequired) {
			t.Fatalf("revoked publish: %v", err)
		}
		if _, err := service.List(ctx, application.Actor{UserID: blocked}, l4bLesson, ""); err != nil {
			t.Fatalf("blocked list: %v", err)
		}
		if _, err := service.List(ctx, application.Actor{UserID: revoked}, l4bLesson, ""); !errors.Is(err, domain.ErrForbidden) {
			t.Fatalf("revoked list: %v", err)
		}
		// Reading the in-course review list is consumption too: a blocked member reads it, a
		// non-member (revoked) is refused — consistent with comment List.
		if _, err := service.LearningReviews(ctx, application.Actor{UserID: blocked}, l4bCourse, ""); err != nil {
			t.Fatalf("blocked review list: %v", err)
		}
		if _, err := service.LearningReviews(ctx, application.Actor{UserID: revoked}, l4bCourse, ""); !errors.Is(err, domain.ErrForbidden) {
			t.Fatalf("revoked review list: %v", err)
		}
	})

	t.Run("L3b lesson and comment facts read every membership state", func(t *testing.T) {
		var membershipStudentID, membershipCourseID, membershipModuleID, membershipLessonID, membershipCommentID string
		if err := pool.QueryRow(ctx, `INSERT INTO users(id,email,display_name,role,password_hash)
			VALUES(gen_random_uuid(),'membership-student@discussion.test','Membership student','student','$argon2id$v=19$m=65536,t=3,p=2$c3R1ZGVudC1kZW1vLXNsdA$5S1MhXX3wgDMnIKuygFRfUNVvTtRktPECxV3ZUBg0is')
			RETURNING id::text`).Scan(&membershipStudentID); err != nil {
			t.Fatal(err)
		}
		if err := pool.QueryRow(ctx, `INSERT INTO courses(id,slug,title,summary,description,status,instructor_id,published_at)
			VALUES(gen_random_uuid(),'membership-discussion','Membership discussion','Summary','Description','published',$1,now()) RETURNING id::text`, discussionOwner).Scan(&membershipCourseID); err != nil {
			t.Fatal(err)
		}
		if err := pool.QueryRow(ctx, `INSERT INTO course_modules(id,course_id,title,position) VALUES(gen_random_uuid(),$1,'Module',0) RETURNING id::text`, membershipCourseID).Scan(&membershipModuleID); err != nil {
			t.Fatal(err)
		}
		if err := pool.QueryRow(ctx, `INSERT INTO course_lessons(id,module_id,title,position) VALUES(gen_random_uuid(),$1,'Lesson',0) RETURNING id::text`, membershipModuleID).Scan(&membershipLessonID); err != nil {
			t.Fatal(err)
		}
		if err := pool.QueryRow(ctx, `INSERT INTO discussion_comments(lesson_id,author_id,body) VALUES($1,$2,'membership facts') RETURNING id::text`, membershipLessonID, discussionOwner).Scan(&membershipCommentID); err != nil {
			t.Fatal(err)
		}
		if _, err := pool.Exec(ctx, `INSERT INTO course_grants(user_id,course_id,source,state,granted_by) VALUES($1,$2,'admin','active',$3)`, membershipStudentID, membershipCourseID, discussionOwner); err != nil {
			t.Fatal(err)
		}
		for _, want := range []access.Membership{access.MembershipActive, access.MembershipBlocked, access.MembershipRevoked} {
			if want != access.MembershipActive {
				if _, err := pool.Exec(ctx, `UPDATE course_grants SET state=$3 WHERE user_id=$1 AND course_id=$2`, membershipStudentID, membershipCourseID, string(want)); err != nil {
					t.Fatal(err)
				}
			}
			lesson, err := repository.LessonFacts(ctx, membershipStudentID, membershipLessonID)
			if err != nil || lesson.Course.Membership != want || lesson.Course.Entitled != (want == access.MembershipActive) {
				t.Fatalf("lesson=%+v err=%v", lesson, err)
			}
			comment, err := repository.Get(ctx, membershipStudentID, membershipCommentID)
			if err != nil || comment.Course.Membership != want || comment.Course.Entitled != (want == access.MembershipActive) {
				t.Fatalf("comment=%+v err=%v", comment, err)
			}
		}
	})

	t.Run("ratings repository", func(t *testing.T) {
		ctx := context.Background()
		fx := seedRatings(t, pool)
		create := func(t *testing.T, id, author string, stars int) ports.ReviewRecord {
			t.Helper()
			record, err := repository.CreateReview(ctx, ports.ReviewRecord{ID: id, CourseID: fx.course, AuthorID: author, Stars: stars, Body: "a sufficiently long review body", CompletedLessons: 2, TotalLessons: 4, CreatedAt: time.Now().UTC()})
			if err != nil {
				t.Fatal(err)
			}
			return record
		}
		t.Run("Rt02 facts for buyer grant and progress", func(t *testing.T) {
			facts, err := repository.ReviewFacts(ctx, fx.course, fx.buyer)
			if err != nil || !facts.Facts.Entitled || !facts.Facts.HasProgress || !facts.Facts.Paid {
				t.Fatalf("facts=%+v err=%v", facts, err)
			}
		})
		t.Run("L3b review facts read every membership state", func(t *testing.T) {
			if _, err := pool.Exec(ctx, `INSERT INTO course_grants(user_id,course_id,source,state,granted_by) VALUES($1,$2,'admin','active',$3)`, fx.preview, fx.course, discussionOwner); err != nil {
				t.Fatal(err)
			}
			for _, want := range []access.Membership{access.MembershipActive, access.MembershipBlocked, access.MembershipRevoked} {
				if want != access.MembershipActive {
					if _, err := pool.Exec(ctx, `UPDATE course_grants SET state=$3 WHERE user_id=$1 AND course_id=$2`, fx.preview, fx.course, string(want)); err != nil {
						t.Fatal(err)
					}
				}
				facts, err := repository.ReviewFacts(ctx, fx.course, fx.preview)
				if err != nil || facts.Facts.Membership != want || facts.Facts.Entitled != (want == access.MembershipActive) {
					t.Fatalf("facts=%+v err=%v", facts, err)
				}
			}
		})
		t.Run("Rt16 preview only refused", func(t *testing.T) {
			facts, err := repository.ReviewFacts(ctx, fx.course, fx.preview)
			if err != nil || facts.Facts.Entitled || !facts.Facts.HasProgress || !facts.Facts.Paid {
				t.Fatalf("facts=%+v err=%v", facts, err)
			}
			if err := domain.CanReview(facts.Facts, domain.Actor{UserID: fx.preview, EmailVerified: true}); !errors.Is(err, domain.ErrReviewPurchaseRequired) {
				t.Fatalf("CanReview=%v", err)
			}
		})
		t.Run("Rt17 admin role is not entitlement", func(t *testing.T) {
			facts, err := repository.ReviewFacts(ctx, fx.course, fx.admin)
			if err != nil {
				t.Fatal(err)
			}
			if err := domain.CanReview(facts.Facts, domain.Actor{UserID: fx.admin, Administrator: true, EmailVerified: true}); !errors.Is(err, domain.ErrReviewPurchaseRequired) {
				t.Fatalf("CanReview=%v", err)
			}
		})
		t.Run("Rt12 frozen progress and Rt13 edit recaptures", func(t *testing.T) {
			review := create(t, fx.reviewA, fx.buyer, 4)
			if _, err := pool.Exec(ctx, `INSERT INTO course_lessons(id,module_id,title,position) VALUES($1::uuid,$2::uuid,'five',4),($3::uuid,$2::uuid,'six',5)`, fx.lesson5, fx.module, fx.lesson6); err != nil {
				t.Fatal(err)
			}
			for _, lesson := range []string{fx.lesson3, fx.lesson4, fx.lesson5, fx.lesson6} {
				if _, err := pool.Exec(ctx, `INSERT INTO lesson_progress(user_id,lesson_id,position_seconds,completed_at,client_seq,updated_at) VALUES($1::uuid,$2::uuid,100,now(),1,now()) ON CONFLICT(user_id,lesson_id) DO UPDATE SET completed_at=now()`, fx.buyer, lesson); err != nil {
					t.Fatal(err)
				}
			}
			stored, err := repository.GetOwnReview(ctx, fx.course, fx.buyer)
			if err != nil || stored.CompletedLessons != 2 || stored.TotalLessons != 4 {
				t.Fatalf("stored=%+v err=%v", stored, err)
			}
			facts, err := repository.ReviewFacts(ctx, fx.course, fx.buyer)
			if err != nil {
				t.Fatal(err)
			}
			content, _ := domain.NewReviewContent(5, "an updated sufficiently long review")
			progress, _ := domain.NewReviewProgress(facts.CompletedLessons, facts.TotalLessons)
			updated, err := repository.UpdateReview(ctx, fx.course, fx.buyer, content, progress, time.Now().UTC())
			if err != nil || updated.CompletedLessons != 6 || updated.TotalLessons != 6 || updated.EditedAt == nil || updated.ID != review.ID {
				t.Fatalf("updated=%+v err=%v", updated, err)
			}
		})
		t.Run("Rt20 revocation preserves review", func(t *testing.T) {
			if _, err := pool.Exec(ctx, `UPDATE course_grants SET state='revoked' WHERE course_id=$1::uuid AND user_id=$2::uuid`, fx.course, fx.buyer); err != nil {
				t.Fatal(err)
			}
			page, err := repository.ListReviews(ctx, fx.course, "", 50)
			if err != nil || len(page.Items) != 1 {
				t.Fatalf("page=%+v err=%v", page, err)
			}
			count, _, err := repository.ReviewSummary(ctx, fx.course)
			if err != nil || count != 1 {
				t.Fatalf("count=%d err=%v", count, err)
			}
			facts, err := repository.ReviewFacts(ctx, fx.course, fx.buyer)
			if err != nil || facts.Facts.Entitled {
				t.Fatalf("facts=%+v err=%v", facts, err)
			}
		})
		t.Run("Rt09 removed review is omitted from summary and list", func(t *testing.T) {
			second := create(t, "75000000-0000-4000-8000-000000000002", fx.preview, 2)
			third := create(t, "75000000-0000-4000-8000-000000000003", fx.admin, 5)
			if err := repository.RemoveReview(ctx, third.ID, discussionOwner, "admin", nil, time.Now().UTC()); err != nil {
				t.Fatal(err)
			}
			count, sum, err := repository.ReviewSummary(ctx, fx.course)
			if err != nil || count != 2 || sum != 7 {
				t.Fatalf("count=%d sum=%d err=%v", count, sum, err)
			}
			page, err := repository.ListReviews(ctx, fx.course, "", 50)
			if err != nil || len(page.Items) != 2 {
				t.Fatalf("page=%+v err=%v", page, err)
			}
			for _, item := range page.Items {
				if item.ID == third.ID {
					t.Fatalf("removed review listed")
				}
			}
			_ = second
		})
		t.Run("Rt22 removal logs atomically", func(t *testing.T) {
			review := create(t, "75000000-0000-4000-8000-000000000004", fx.grantee, 3)
			if err := repository.RemoveReview(ctx, review.ID, discussionOwner, "admin", nil, time.Now().UTC()); err != nil {
				t.Fatal(err)
			}
			var reviewID, actor string
			if err := pool.QueryRow(ctx, `SELECT review_id::text,actor_id::text FROM discussion_moderation_log WHERE review_id=$1::uuid`, review.ID).Scan(&reviewID, &actor); err != nil || reviewID != review.ID || actor != discussionOwner {
				t.Fatalf("log review=%q actor=%q err=%v", reviewID, actor, err)
			}
			blocked := create(t, "75000000-0000-4000-8000-000000000005", fx.reporter, 1)
			if _, err := pool.Exec(ctx, `ALTER TABLE discussion_moderation_log ADD CONSTRAINT test_force_fail CHECK (reason IS DISTINCT FROM 'force-fail') NOT VALID`); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				_, _ = pool.Exec(context.Background(), `ALTER TABLE discussion_moderation_log DROP CONSTRAINT test_force_fail`)
			})
			reason := "force-fail"
			if err := repository.RemoveReview(ctx, blocked.ID, discussionOwner, "admin", &reason, time.Now().UTC()); err == nil {
				t.Fatal("expected forced log failure")
			}
			var removed bool
			var logs int
			if err := pool.QueryRow(ctx, `SELECT removed_at IS NOT NULL FROM discussion_reviews WHERE id=$1::uuid`, blocked.ID).Scan(&removed); err != nil {
				t.Fatal(err)
			}
			if err := pool.QueryRow(ctx, `SELECT count(*) FROM discussion_moderation_log WHERE review_id=$1::uuid`, blocked.ID).Scan(&logs); err != nil {
				t.Fatal(err)
			}
			if removed || logs != 0 {
				t.Fatalf("removed=%v logs=%d", removed, logs)
			}
		})
		t.Run("concurrent create and removed/delete semantics", func(t *testing.T) {
			base := ports.ReviewRecord{ID: "75000000-0000-4000-8000-000000000006", CourseID: fx.course, AuthorID: fx.concurrent, Stars: 4, Body: "a sufficiently long review body", CompletedLessons: 0, TotalLessons: 0, CreatedAt: time.Now().UTC()}
			results := make(chan error, 2)
			for i := 0; i < 2; i++ {
				go func() { _, err := repository.CreateReview(context.Background(), base); results <- err }()
			}
			var successes, duplicates int
			for i := 0; i < 2; i++ {
				err := <-results
				if err == nil {
					successes++
				} else if errors.Is(err, domain.ErrReviewAlreadyExists) {
					duplicates++
				} else {
					t.Fatalf("create=%v", err)
				}
			}
			if successes != 1 || duplicates != 1 {
				t.Fatalf("success=%d duplicate=%d", successes, duplicates)
			}
			if err := repository.RemoveReview(ctx, base.ID, discussionOwner, "admin", nil, time.Now().UTC()); err != nil {
				t.Fatal(err)
			}
			if _, err := repository.CreateReview(ctx, base); !errors.Is(err, domain.ErrReviewRemoved) {
				t.Fatalf("create removed=%v", err)
			}
			if err := repository.DeleteReview(ctx, fx.course, base.AuthorID); !errors.Is(err, domain.ErrReviewRemoved) {
				t.Fatalf("delete removed=%v", err)
			}
			// A hard delete frees its unique slot.
			rewrite := ports.ReviewRecord{ID: "75000000-0000-4000-8000-000000000008", CourseID: fx.course, AuthorID: fx.preview, Stars: 4, Body: "a sufficiently long review body", CreatedAt: time.Now().UTC()}
			if err := repository.DeleteReview(ctx, fx.course, fx.preview); err != nil {
				t.Fatal(err)
			}
			if _, err := repository.CreateReview(ctx, rewrite); err != nil {
				t.Fatal(err)
			}
		})
		t.Run("ListModerationQueue mixes kinds and keyset pages", func(t *testing.T) {
			comment, err := repository.InsertRoot(ctx, domain.Comment{LessonID: "74000000-0000-4000-8000-000000000001", AuthorID: fx.concurrent, Body: "queue comment"})
			if err != nil {
				t.Fatal(err)
			}
			if err := repository.InsertReport(ctx, domain.Report{CommentID: comment.Comment.ID, ReporterID: fx.admin}, time.Now().UTC().Add(time.Second)); err != nil {
				t.Fatal(err)
			}
			// The queue must hold one open report of each kind; this subtest files its own review
			// report instead of relying on a later subtest. Review ...0005 (fx.reporter's, from the
			// atomic-removal subtest) is asserted there to stay not removed; Rt25 counts reviewA only.
			if err := repository.InsertReviewReport(ctx, "75000000-0000-4000-8000-000000000005", fx.buyer, time.Now().UTC()); err != nil {
				t.Fatal(err)
			}
			queue, err := repository.ListModerationQueue(ctx, fx.course, "", 1)
			if err != nil || len(queue.Items) != 1 || queue.NextCursor == "" {
				t.Fatalf("queue=%+v err=%v", queue, err)
			}
			second, err := repository.ListModerationQueue(ctx, fx.course, queue.NextCursor, 1)
			if err != nil || len(second.Items) != 1 {
				t.Fatalf("second=%+v err=%v", second, err)
			}
			if queue.Items[0].Kind == second.Items[0].Kind || (queue.Items[0].Kind != "comment" && queue.Items[0].Kind != "review") || (second.Items[0].Kind != "comment" && second.Items[0].Kind != "review") {
				t.Fatalf("kinds=%q,%q", queue.Items[0].Kind, second.Items[0].Kind)
			}
			// LR3: authorBlocked reflects a real blocked membership. The comment author has none yet;
			// after a blocked grant the same queued comment reports authorBlocked.
			both, err := repository.ListModerationQueue(ctx, fx.course, "", 10)
			if err != nil {
				t.Fatal(err)
			}
			for _, item := range both.Items {
				if item.AuthorBlocked {
					t.Fatalf("unexpected authorBlocked before any block: %+v", item)
				}
			}
			if _, err := pool.Exec(ctx, `INSERT INTO course_grants(user_id,course_id,source,state,granted_at) VALUES($1::uuid,$2::uuid,'self','blocked',now()) ON CONFLICT (user_id,course_id) DO UPDATE SET state='blocked'`, fx.concurrent, fx.course); err != nil {
				t.Fatal(err)
			}
			reblocked, err := repository.ListModerationQueue(ctx, fx.course, "", 10)
			if err != nil {
				t.Fatal(err)
			}
			var sawComment bool
			for _, item := range reblocked.Items {
				if item.Kind == "comment" {
					sawComment = true
					if !item.AuthorBlocked {
						t.Fatalf("comment author should be blocked: %+v", item)
					}
				}
			}
			if !sawComment {
				t.Fatal("comment item missing after block")
			}
		})
		t.Run("Rt25 own report refused and duplicate idempotent", func(t *testing.T) {
			if err := repository.InsertReviewReport(ctx, fx.reviewA, fx.buyer, time.Now().UTC()); !errors.Is(err, domain.ErrReportOwnReview) {
				t.Fatalf("own report=%v", err)
			}
			if err := repository.InsertReviewReport(ctx, fx.reviewA, fx.reporter, time.Now().UTC()); err != nil {
				t.Fatal(err)
			}
			if err := repository.InsertReviewReport(ctx, fx.reviewA, fx.reporter, time.Now().UTC()); err != nil {
				t.Fatal(err)
			}
			var count int
			if err := pool.QueryRow(ctx, `SELECT count(*) FROM discussion_review_reports WHERE review_id=$1::uuid`, fx.reviewA).Scan(&count); err != nil || count != 1 {
				t.Fatalf("count=%d err=%v", count, err)
			}
		})
		t.Run("Rt26 archive changes review reads and facts", func(t *testing.T) {
			// Compare before and after: the course's review count depends on earlier subtests.
			before, err := repository.ListReviews(ctx, fx.course, "", 50)
			if err != nil {
				t.Fatal(err)
			}
			beforeCount, beforeSum, err := repository.ReviewSummary(ctx, fx.course)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := pool.Exec(ctx, `UPDATE courses SET status='archived' WHERE id=$1::uuid`, fx.course); err != nil {
				t.Fatal(err)
			}
			page, err := repository.ListReviews(ctx, fx.course, "", 50)
			if err != nil || len(page.Items) != len(before.Items) || len(page.Items) == 0 {
				t.Fatalf("before=%d after=%d err=%v", len(before.Items), len(page.Items), err)
			}
			afterCount, afterSum, err := repository.ReviewSummary(ctx, fx.course)
			if err != nil || afterCount != beforeCount || afterSum != beforeSum {
				t.Fatalf("summary before=(%d,%d) after=(%d,%d) err=%v", beforeCount, beforeSum, afterCount, afterSum, err)
			}
			facts, err := repository.ReviewFacts(ctx, fx.course, fx.grantee)
			if err != nil || !facts.Facts.Entitled {
				t.Fatalf("facts=%+v err=%v", facts, err)
			}
		})
		t.Run("Rt27 no derived rating storage anywhere in the schema", func(t *testing.T) {
			// Word-boundary match: course_lessons.is_preview contains "review" and is not rating storage.
			// R6: ratings are computed at read time. Any summary table, view, matview, counter column or trigger
			// that stores derived review data fails here — not only the two names the migration happens to use.
			queries := []struct {
				name  string
				sql   string
				allow int
			}{
				{"relations named review/rating (tables, views, matviews)", `SELECT count(*) FROM pg_class c JOIN pg_namespace n ON n.oid=c.relnamespace WHERE n.nspname='public' AND c.relkind IN ('r','v','m','p') AND c.relname ~* '(^|_)(reviews?|ratings?)(_|$)' AND c.relname NOT IN ('discussion_reviews','discussion_review_reports')`, 0},
				{"columns named review/rating outside the review tables", `SELECT count(*) FROM information_schema.columns WHERE table_schema='public' AND table_name NOT IN ('discussion_reviews','discussion_review_reports') AND column_name ~* '(^|_)(reviews?|ratings?)(_|$)' AND NOT (table_name='discussion_moderation_log' AND column_name='review_id')`, 0},
				{"triggers whose function touches reviews", `SELECT count(*) FROM pg_trigger t JOIN pg_proc p ON p.oid=t.tgfoid WHERE NOT t.tgisinternal AND (p.prosrc ILIKE '%review%' OR p.prosrc ILIKE '%rating%')`, 0},
				{"triggers on the review tables", `SELECT count(*) FROM pg_trigger WHERE NOT tgisinternal AND tgrelid IN ('discussion_reviews'::regclass,'discussion_review_reports'::regclass)`, 0},
				{"views or matviews depending on discussion_reviews", `SELECT count(*) FROM pg_depend d JOIN pg_rewrite r ON r.oid=d.objid WHERE d.refobjid='discussion_reviews'::regclass`, 0},
			}
			for _, q := range queries {
				var got int
				if err := pool.QueryRow(ctx, q.sql).Scan(&got); err != nil {
					t.Fatalf("%s: %v", q.name, err)
				}
				if got != q.allow {
					t.Fatalf("%s: found %d, want %d", q.name, got, q.allow)
				}
			}
		})

		t.Run("moderation log rejects both targets and neither target", func(t *testing.T) {
			// T03 review №1: the one-target CHECK (num_nonnulls = 1). Both statements must fail; nothing is left behind.
			for name, statement := range map[string]string{
				"both":    `INSERT INTO discussion_moderation_log(course_id,comment_id,review_id,actor_id,actor_role,action) VALUES ($1::uuid,gen_random_uuid(),$2::uuid,$3::uuid,'admin','remove')`,
				"neither": `INSERT INTO discussion_moderation_log(course_id,comment_id,review_id,actor_id,actor_role,action) VALUES ($1::uuid,NULL,NULL,$2::uuid,'admin','remove')`,
			} {
				args := []any{fx.course, fx.reviewA, fx.admin}
				if name == "neither" {
					args = []any{fx.course, fx.admin}
				}
				_, err := pool.Exec(ctx, statement, args...)
				var pgErr *pgconn.PgError
				if !errors.As(err, &pgErr) || pgErr.Code != "23514" || pgErr.ConstraintName != "discussion_moderation_log_one_target" {
					t.Fatalf("%s target: want CHECK violation on discussion_moderation_log_one_target, got %v", name, err)
				}
			}
		})
		t.Run("ListReviews pages newest first with a keyset cursor", func(t *testing.T) {
			all, err := repository.ListReviews(ctx, fx.course, "", 50)
			if err != nil || len(all.Items) < 3 {
				t.Fatalf("need at least 3 visible reviews in the fixture course: %d err=%v", len(all.Items), err)
			}
			first, err := repository.ListReviews(ctx, fx.course, "", 2)
			if err != nil || len(first.Items) != 2 || first.NextCursor == "" {
				t.Fatalf("first page=%d cursor=%q err=%v", len(first.Items), first.NextCursor, err)
			}
			rest, err := repository.ListReviews(ctx, fx.course, first.NextCursor, 50)
			if err != nil || len(first.Items)+len(rest.Items) != len(all.Items) || rest.NextCursor != "" {
				t.Fatalf("rest=%d total=%d cursor=%q err=%v", len(rest.Items), len(all.Items), rest.NextCursor, err)
			}
			paged := append(append([]ports.ReviewRecord{}, first.Items...), rest.Items...)
			for i := range paged {
				if paged[i].ID != all.Items[i].ID {
					t.Fatalf("position %d: paged %s, unpaged %s", i, paged[i].ID, all.Items[i].ID)
				}
				if i > 0 && paged[i].CreatedAt.After(paged[i-1].CreatedAt) {
					t.Fatalf("not newest first at %d", i)
				}
			}
		})
	})

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
			// Reporting is communication now, so the reporter must be an active member of the
			// course to reach the no-op path (enroll the fixture, not loosen the gate).
			if _, err := pool.Exec(ctx, `INSERT INTO course_grants(user_id,course_id,source,state,granted_by) VALUES($1::uuid,$2::uuid,'admin','active',$3::uuid) ON CONFLICT (user_id,course_id) DO UPDATE SET state='active'`, discussionReporter, moderationCourse, discussionOwner); err != nil {
				t.Fatal(err)
			}
			service := application.New(repository, func() time.Time { return at.Add(time.Second) })
			if err := service.ReportComment(ctx, domain.Actor{UserID: discussionReporter, EmailVerified: true}, comment.Comment.ID); err != nil {
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
			if err := pool.QueryRow(ctx, `INSERT INTO course_lessons(id,module_id,title,position,is_preview) VALUES(gen_random_uuid(),$1,'L',0,$2) RETURNING id::text`, moduleID, truth.Preview).Scan(&lessonID); err != nil {
				t.Fatal(err)
			}
			actor := domain.Actor{UserID: discussionStudent, Administrator: truth.Role == accesstest.Administrator, EmailVerified: true}
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
		`INSERT INTO course_lessons(id,module_id,title,position) VALUES ('` + discussionLesson + `','20000000-0000-4000-8000-0000000000e1','L',0),('` + positiveLesson + `','20000000-0000-4000-8000-0000000000e1','Positive',1),('` + cascadeLesson + `','20000000-0000-4000-8000-0000000000e1','Cascade',2)`,
		`INSERT INTO course_lessons(id,module_id,title,position) VALUES ('` + moderationLessonA + `','` + moderationModule + `','A',0),('` + moderationLessonB + `','` + moderationModule + `','B',1)`,
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

type ratingsFixture struct {
	course, module                                       string
	buyer, preview, admin, grantee, reporter, concurrent string
	lesson3, lesson4, lesson5, lesson6                   string
	reviewA                                              string
}

func seedRatings(t *testing.T, pool *pgxpool.Pool) ratingsFixture {
	t.Helper()
	fx := ratingsFixture{
		course: "71000000-0000-4000-8000-000000000001", module: "72000000-0000-4000-8000-000000000001",
		buyer: "73000000-0000-4000-8000-000000000001", preview: "73000000-0000-4000-8000-000000000002", admin: "73000000-0000-4000-8000-000000000003", grantee: "73000000-0000-4000-8000-000000000004", reporter: "73000000-0000-4000-8000-000000000005", concurrent: "73000000-0000-4000-8000-000000000006",
		lesson3: "74000000-0000-4000-8000-000000000003", lesson4: "74000000-0000-4000-8000-000000000004", lesson5: "74000000-0000-4000-8000-000000000005", lesson6: "74000000-0000-4000-8000-000000000006", reviewA: "75000000-0000-4000-8000-000000000001",
	}
	ctx := context.Background()
	hash := "$argon2id$v=19$m=65536,t=3,p=2$c3R1ZGVudC1kZW1vLXNsdA$5S1MhXX3wgDMnIKuygFRfUNVvTtRktPECxV3ZUBg0is"
	for i, user := range []struct{ id, role string }{{fx.buyer, "student"}, {fx.preview, "student"}, {fx.admin, "admin"}, {fx.grantee, "student"}, {fx.reporter, "student"}, {fx.concurrent, "student"}} {
		if _, err := pool.Exec(ctx, `INSERT INTO users(id,email,display_name,role,password_hash) VALUES($1::uuid,$2,$3,$4,$5)`, user.id, "rating"+strconv.Itoa(i)+"@test.local", "Rating", user.role, hash); err != nil {
			t.Fatal(err)
		}
	}
	statements := []struct {
		sql  string
		args []any
	}{
		{`INSERT INTO courses(id,slug,title,summary,description,status,instructor_id,published_at,price_minor,currency) VALUES($1::uuid,'ratings-fixture','ratings','s','d','published',$2::uuid,now(),100,'UAH')`, []any{fx.course, discussionOwner}},
		{`INSERT INTO course_modules(id,course_id,title,position) VALUES($1::uuid,$2::uuid,'ratings',0)`, []any{fx.module, fx.course}},
		{`INSERT INTO course_lessons(id,module_id,title,position,is_preview) VALUES('74000000-0000-4000-8000-000000000001',$1::uuid,'one',0,false),('74000000-0000-4000-8000-000000000002',$1::uuid,'two',1,true),($2::uuid,$1::uuid,'three',2,false),($3::uuid,$1::uuid,'four',3,false)`, []any{fx.module, fx.lesson3, fx.lesson4}},
		{`INSERT INTO course_grants(user_id,course_id,source,state,granted_at) VALUES($1::uuid,$3::uuid,'admin','active',now()),($2::uuid,$3::uuid,'admin','active',now())`, []any{fx.buyer, fx.grantee, fx.course}},
	}
	for _, statement := range statements {
		if _, err := pool.Exec(ctx, statement.sql, statement.args...); err != nil {
			t.Fatal(err)
		}
	}
	for _, row := range []struct {
		user, lesson string
		complete     bool
	}{{fx.buyer, "74000000-0000-4000-8000-000000000001", true}, {fx.buyer, "74000000-0000-4000-8000-000000000002", true}, {fx.preview, "74000000-0000-4000-8000-000000000002", false}, {fx.admin, "74000000-0000-4000-8000-000000000001", false}} {
		var completed any
		if row.complete {
			completed = time.Now().UTC()
		}
		if _, err := pool.Exec(ctx, `INSERT INTO lesson_progress(user_id,lesson_id,position_seconds,completed_at,client_seq,updated_at) VALUES($1::uuid,$2::uuid,10,$3,1,now())`, row.user, row.lesson, completed); err != nil {
			t.Fatal(err)
		}
	}
	return fx
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
