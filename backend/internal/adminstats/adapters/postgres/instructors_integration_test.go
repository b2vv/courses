//go:build integration

package postgres_test

import (
	"context"
	"testing"
	"time"

	adminstatspostgres "courses/internal/adminstats/adapters/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TestAdminStatsInstructorsLeaderboard proves the per-instructor leaderboard
// ranks by paid revenue (9b) and keeps an instructor with no sales at 0 rather
// than dropping them (9c). It reads through Snapshot so the loadInstructors
// wiring is exercised too.
func TestAdminStatsInstructorsLeaderboard(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	pool := newTestPool(t, ctx)
	now := time.Now().UTC()

	// Three instructors, each owning one published course.
	insertUser(t, ctx, pool, "ins-a", "instructor", false, false, true, false, now)
	insertUser(t, ctx, pool, "ins-b", "instructor", false, false, true, false, now)
	insertUser(t, ctx, pool, "ins-c", "instructor", false, false, true, false, now)
	// Two students who enrol and buy.
	insertUser(t, ctx, pool, "buy-1", "student", false, false, true, false, now)
	insertUser(t, ctx, pool, "buy-2", "student", false, false, true, false, now)

	insertCourse(t, ctx, pool, "ca", "ins-a", now)
	insertCourse(t, ctx, pool, "cb", "ins-b", now)
	insertCourse(t, ctx, pool, "cc", "ins-c", now)

	// Active enrolments: A has 2 (ca), B has 1 (cb), C has 0.
	insertActiveGrant(t, ctx, pool, "buy-1", "ca", now)
	insertActiveGrant(t, ctx, pool, "buy-2", "ca", now)
	insertActiveGrant(t, ctx, pool, "buy-1", "cb", now)

	// Paid revenue: A = 6000+4000 = 10000, B = 3000, C = 0.
	insertPaidOrder(t, ctx, pool, "buy-1", "ca", 6000, "a1", now)
	insertPaidOrder(t, ctx, pool, "buy-2", "ca", 4000, "a2", now)
	insertPaidOrder(t, ctx, pool, "buy-1", "cb", 3000, "b1", now)

	snapshot, err := adminstatspostgres.NewRepository(pool).Snapshot(ctx)
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	got := snapshot.Instructors
	if len(got) != 3 {
		t.Fatalf("leaderboard has %d rows, want 3 (no-sales instructor must not drop): %+v", len(got), got)
	}

	// 9b: ordered by revenue DESC — A, then B, then C.
	if got[0].InstructorID != userUUID("ins-a") || got[0].RevenueMinorUAH != 10000 || got[0].ActiveEnrollments != 2 || got[0].PublishedCourses != 1 {
		t.Fatalf("rank 1 = %+v, want ins-a revenue10000 enroll2 published1", got[0])
	}
	if got[1].InstructorID != userUUID("ins-b") || got[1].RevenueMinorUAH != 3000 || got[1].ActiveEnrollments != 1 || got[1].PublishedCourses != 1 {
		t.Fatalf("rank 2 = %+v, want ins-b revenue3000 enroll1 published1", got[1])
	}
	// 9c: the no-sales instructor is present at 0, ranked last.
	if got[2].InstructorID != userUUID("ins-c") || got[2].RevenueMinorUAH != 0 || got[2].ActiveEnrollments != 0 || got[2].PublishedCourses != 1 {
		t.Fatalf("rank 3 = %+v, want ins-c revenue0 enroll0 published1", got[2])
	}
	if got[2].DisplayName == "" {
		t.Fatalf("rank 3 display name empty: %+v", got[2])
	}
}

func insertCourse(t *testing.T, ctx context.Context, pool *pgxpool.Pool, slug, instructorSlug string, now time.Time) {
	t.Helper()
	exec(t, ctx, pool, `INSERT INTO courses(id,slug,title,summary,description,status,created_at,updated_at,published_at,instructor_id)
		VALUES ($1,$2,'Course','s','d','published',$3,$3,$3,$4)`,
		courseUUID(slug), slug, now, userUUID(instructorSlug))
}

func insertActiveGrant(t *testing.T, ctx context.Context, pool *pgxpool.Pool, buyerSlug, courseSlug string, now time.Time) {
	t.Helper()
	exec(t, ctx, pool, `INSERT INTO course_grants(id,user_id,course_id,source,order_id,state,granted_at,granted_by)
		VALUES (gen_random_uuid(),$1,$2,'admin',NULL,'active',$3,NULL)`,
		userUUID(buyerSlug), courseUUID(courseSlug), now)
}

func insertPaidOrder(t *testing.T, ctx context.Context, pool *pgxpool.Pool, buyerSlug, courseSlug string, amountMinor int64, key string, now time.Time) {
	t.Helper()
	exec(t, ctx, pool, `INSERT INTO orders(id,user_id,course_id,amount_minor,currency,status,idempotency_key,created_at,paid_at)
		VALUES (gen_random_uuid(),$1,$2,$3,'UAH','paid',$4,$5,$5)`,
		userUUID(buyerSlug), courseUUID(courseSlug), amountMinor, key, now)
}
