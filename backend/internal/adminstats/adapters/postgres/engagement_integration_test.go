//go:build integration

package postgres_test

import (
	"context"
	"testing"
	"time"

	adminstatspostgres "courses/internal/adminstats/adapters/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	engModuleA = "60000000-0000-4000-8000-0000000000a0"
	engModuleB = "60000000-0000-4000-8000-0000000000b0"
	engLesson1 = "70000000-0000-4000-8000-000000000001"
	engLesson2 = "70000000-0000-4000-8000-000000000002"
	engLesson3 = "70000000-0000-4000-8000-000000000003"
	engLesson4 = "70000000-0000-4000-8000-000000000004"
)

// TestAdminStatsEngagement proves the three engagement metrics: active learners
// over 30 days, the mean completed-lesson fraction over active grants, and the
// count of fully-completed memberships. Progress counts a lesson by its
// completed_at; the active-learner window is the lesson_progress.updated_at.
func TestAdminStatsEngagement(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	pool := newTestPool(t, ctx)
	now := time.Now().UTC()
	old := now.Add(-40 * 24 * time.Hour) // outside the 30-day active window

	insertUser(t, ctx, pool, "eng-instr", "instructor", false, false, true, false, now)
	for _, slug := range []string{"eng-s1", "eng-s2", "eng-s3", "eng-s4"} {
		insertUser(t, ctx, pool, slug, "student", false, false, true, false, now)
	}
	insertCourse(t, ctx, pool, "eng-ca", "eng-instr", now)
	insertCourse(t, ctx, pool, "eng-cb", "eng-instr", now)
	insertModule(t, ctx, pool, engModuleA, "eng-ca", now)
	insertModule(t, ctx, pool, engModuleB, "eng-cb", now)
	insertLesson(t, ctx, pool, engLesson1, engModuleA, 0, now)
	insertLesson(t, ctx, pool, engLesson2, engModuleA, 1, now)
	insertLesson(t, ctx, pool, engLesson3, engModuleB, 0, now)
	insertLesson(t, ctx, pool, engLesson4, engModuleB, 1, now)

	// Active grants: s1,s2,s4 on course A; s3 on course B.
	insertActiveGrant(t, ctx, pool, "eng-s1", "eng-ca", now)
	insertActiveGrant(t, ctx, pool, "eng-s2", "eng-ca", now)
	insertActiveGrant(t, ctx, pool, "eng-s4", "eng-ca", now)
	insertActiveGrant(t, ctx, pool, "eng-s3", "eng-cb", now)

	// s1: 1 of 2 lessons completed, recent  -> progress 0.5, active learner.
	insertProgress(t, ctx, pool, "eng-s1", engLesson1, &now, now)
	// s2: 2 of 2 completed, recent           -> progress 1.0, completion, active.
	insertProgress(t, ctx, pool, "eng-s2", engLesson1, &now, now)
	insertProgress(t, ctx, pool, "eng-s2", engLesson2, &now, now)
	// s3: touched a lesson recently, none completed -> progress 0.0, active.
	insertProgress(t, ctx, pool, "eng-s3", engLesson3, nil, now)
	// s4: 1 of 2 completed but the row is 40 days old -> progress 0.5, NOT active.
	insertProgress(t, ctx, pool, "eng-s4", engLesson1, &old, old)

	snapshot, err := adminstatspostgres.NewRepository(pool).Snapshot(ctx)
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	engagement := snapshot.Engagement

	// Active learners: s1, s2, s3 (s4's only activity is 40 days old).
	if engagement.ActiveLearners30d != 3 {
		t.Fatalf("activeLearners30d = %d, want 3", engagement.ActiveLearners30d)
	}
	// Avg progress over all four active grants: (0.5 + 1.0 + 0.0 + 0.5) / 4.
	if engagement.AvgProgress == nil || *engagement.AvgProgress != 0.5 {
		t.Fatalf("avgProgress = %v, want 0.5 over active grants", engagement.AvgProgress)
	}
	// Completions: only s2 finished every lesson of its course.
	if engagement.Completions != 1 {
		t.Fatalf("completions = %d, want 1", engagement.Completions)
	}
}

// TestAdminStatsEngagementEmpty proves the honest empty case: no active grant
// leaves the average nil (AVG over zero rows), not a misleading 0.
func TestAdminStatsEngagementEmpty(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	pool := newTestPool(t, ctx)

	snapshot, err := adminstatspostgres.NewRepository(pool).Snapshot(ctx)
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	if snapshot.Engagement.ActiveLearners30d != 0 || snapshot.Engagement.Completions != 0 {
		t.Fatalf("empty engagement counts not zero: %+v", snapshot.Engagement)
	}
	if snapshot.Engagement.AvgProgress != nil {
		t.Fatalf("avgProgress should be nil with no active grant, got %v", *snapshot.Engagement.AvgProgress)
	}
}

func insertModule(t *testing.T, ctx context.Context, pool *pgxpool.Pool, id, courseSlug string, now time.Time) {
	t.Helper()
	exec(t, ctx, pool, `INSERT INTO course_modules(id,course_id,title,position,created_at,updated_at)
		VALUES ($1,$2,'Module',0,$3,$3)`, id, courseUUID(courseSlug), now)
}

func insertLesson(t *testing.T, ctx context.Context, pool *pgxpool.Pool, id, moduleID string, position int, now time.Time) {
	t.Helper()
	exec(t, ctx, pool, `INSERT INTO course_lessons(id,module_id,title,position,created_at,updated_at)
		VALUES ($1,$2,'Lesson',$3,$4,$4)`, id, moduleID, position, now)
}

func insertProgress(t *testing.T, ctx context.Context, pool *pgxpool.Pool, userSlug, lessonID string, completedAt *time.Time, updatedAt time.Time) {
	t.Helper()
	exec(t, ctx, pool, `INSERT INTO lesson_progress(user_id,lesson_id,position_seconds,completed_at,client_seq,updated_at)
		VALUES ($1,$2,0,$3,1,$4)`, userUUID(userSlug), lessonID, completedAt, updatedAt)
}
