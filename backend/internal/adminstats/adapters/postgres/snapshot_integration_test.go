//go:build integration

package postgres_test

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	adminstatspostgres "courses/internal/adminstats/adapters/postgres"
	containers "courses/internal/testsupport/containers"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	postgrescontainer "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// validArgon2Hash satisfies the users.password_hash CHECK ('$argon2id$' prefix,
// 60..512 chars) without being a real credential.
const validArgon2Hash = "$argon2id$v=19$m=65536,t=3,p=4$c29tZXNhbHR2YWx1ZQ$c29tZWRlcml2ZWRrZXl2YWx1ZXRoYXRpc2xvbmdlbm91Z2g"

// newTestPool starts a disposable Postgres, runs every migration and returns a
// pool wired to it; the container and pool are torn down via t.Cleanup.
func newTestPool(t *testing.T, ctx context.Context) *pgxpool.Pool {
	t.Helper()
	container, err := containers.Start(ctx, func(ctx context.Context) (*postgrescontainer.PostgresContainer, error) {
		return postgrescontainer.Run(ctx, "postgres:17.5", postgrescontainer.WithDatabase("courses_test"), postgrescontainer.WithUsername("courses"), postgrescontainer.WithPassword("ephemeral-test-only"), testcontainers.WithWaitStrategy(wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(3*time.Minute)))
	})
	if err != nil {
		t.Fatalf("PostgreSQL integration requires Docker: %v", err)
	}
	t.Cleanup(func() { _ = container.Terminate(context.Background()) })
	connection, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("pgx", connection)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	provider, err := goose.NewProvider(goose.DialectPostgres, db, os.DirFS(filepath.Join("..", "..", "..", "..", "migrations")), goose.WithDisableGlobalRegistry(true))
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
	t.Cleanup(pool.Close)
	return pool
}

func TestAdminStatsSnapshotAgainstPostgreSQL(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	pool := newTestPool(t, ctx)

	repository := adminstatspostgres.NewRepository(pool)

	t.Run("EmptyPlatformIsAllZeros", func(t *testing.T) {
		snapshot, err := repository.Snapshot(ctx)
		if err != nil {
			t.Fatalf("snapshot on empty platform: %v", err)
		}
		u := snapshot.Users
		if u.Active != 0 || u.Blocked != 0 || u.Deleted != 0 || u.Verified != 0 || u.TwoFactor != 0 ||
			u.ByRole.Student != 0 || u.ByRole.Instructor != 0 || u.ByRole.Admin != 0 {
			t.Fatalf("users not zero-filled: %+v", u)
		}
		if snapshot.Courses.Draft != 0 || snapshot.Courses.Published != 0 || snapshot.Courses.Archived != 0 {
			t.Fatalf("courses not zero: %+v", snapshot.Courses)
		}
		if snapshot.Grants.Active != 0 || snapshot.Grants.Blocked != 0 || snapshot.Grants.Revoked != 0 {
			t.Fatalf("grants not zero: %+v", snapshot.Grants)
		}
		if snapshot.Orders.Pending != 0 || snapshot.Orders.Paid != 0 || snapshot.Orders.Failed != 0 || snapshot.Orders.RevenueMinorUAH != 0 {
			t.Fatalf("orders not zero: %+v", snapshot.Orders)
		}
		if snapshot.Reports.Open != 0 || snapshot.Reports.Resolved != 0 {
			t.Fatalf("reports not zero: %+v", snapshot.Reports)
		}
		if snapshot.Media.InProgress != 0 || snapshot.Media.Ready != 0 || snapshot.Media.Failed != 0 {
			t.Fatalf("media counts not zero: %+v", snapshot.Media)
		}
		if snapshot.Media.TotalDurationSeconds != 0 || snapshot.Media.WithDurationCount != 0 {
			t.Fatalf("media duration aggregates not zero: %+v", snapshot.Media)
		}
		// AVG over zero rows is NULL, surfaced as nil — not a misleading 0.
		if snapshot.Media.AvgDurationSeconds != nil {
			t.Fatalf("avg duration should be nil on empty platform, got %v", *snapshot.Media.AvgDurationSeconds)
		}
	})

	seed(t, ctx, pool)

	t.Run("CountsSeededData", func(t *testing.T) {
		snapshot, err := repository.Snapshot(ctx)
		if err != nil {
			t.Fatalf("snapshot after seed: %v", err)
		}
		// Users: 2 active, 1 blocked, 1 deleted (deleted excluded from roles).
		if snapshot.Users.Active != 2 {
			t.Fatalf("active = %d, want 2", snapshot.Users.Active)
		}
		if snapshot.Users.Blocked != 1 {
			t.Fatalf("blocked = %d, want 1", snapshot.Users.Blocked)
		}
		if snapshot.Users.Deleted != 1 {
			t.Fatalf("deleted = %d, want 1", snapshot.Users.Deleted)
		}
		// Non-deleted by role: 1 student, 1 instructor, 1 admin.
		if snapshot.Users.ByRole.Student != 1 || snapshot.Users.ByRole.Instructor != 1 || snapshot.Users.ByRole.Admin != 1 {
			t.Fatalf("byRole = %+v, want 1/1/1", snapshot.Users.ByRole)
		}
		if snapshot.Users.Verified != 2 {
			t.Fatalf("verified = %d, want 2", snapshot.Users.Verified)
		}
		if snapshot.Users.TwoFactor != 1 {
			t.Fatalf("twoFactor = %d, want 1", snapshot.Users.TwoFactor)
		}
		// A3 consistency: active+blocked = non-deleted users = the identity
		// list total (deleted excluded). Checked directly against the table.
		var nonDeleted int
		if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE deleted_at IS NULL`).Scan(&nonDeleted); err != nil {
			t.Fatal(err)
		}
		if snapshot.Users.Active+snapshot.Users.Blocked != nonDeleted {
			t.Fatalf("active+blocked = %d, identity non-deleted = %d", snapshot.Users.Active+snapshot.Users.Blocked, nonDeleted)
		}

		if snapshot.Courses.Draft != 1 || snapshot.Courses.Published != 1 || snapshot.Courses.Archived != 1 {
			t.Fatalf("courses = %+v, want 1/1/1", snapshot.Courses)
		}
		if snapshot.Grants.Active != 1 || snapshot.Grants.Blocked != 1 || snapshot.Grants.Revoked != 1 {
			t.Fatalf("grants = %+v, want 1/1/1", snapshot.Grants)
		}
		if snapshot.Orders.Pending != 1 || snapshot.Orders.Paid != 2 || snapshot.Orders.Failed != 1 {
			t.Fatalf("orders counts = %+v, want pending1/paid2/failed1", snapshot.Orders)
		}
		// Revenue sums the two paid orders only (5000 + 7000), not pending/failed.
		if snapshot.Orders.RevenueMinorUAH != 12000 {
			t.Fatalf("revenue = %d, want 12000", snapshot.Orders.RevenueMinorUAH)
		}
		if snapshot.Reports.Open != 1 || snapshot.Reports.Resolved != 1 {
			t.Fatalf("reports = %+v, want 1/1", snapshot.Reports)
		}
		// Media: 1 in-progress (processing), 2 ready, 1 failed.
		if snapshot.Media.InProgress != 1 || snapshot.Media.Ready != 2 || snapshot.Media.Failed != 1 {
			t.Fatalf("media counts = %+v, want inProgress1/ready2/failed1", snapshot.Media)
		}
		// Only the two ready assets carry a duration (100 + 200).
		if snapshot.Media.WithDurationCount != 2 {
			t.Fatalf("withDuration = %d, want 2", snapshot.Media.WithDurationCount)
		}
		if snapshot.Media.TotalDurationSeconds != 300 {
			t.Fatalf("total duration = %v, want 300", snapshot.Media.TotalDurationSeconds)
		}
		if snapshot.Media.AvgDurationSeconds == nil || *snapshot.Media.AvgDurationSeconds != 150 {
			t.Fatalf("avg duration = %v, want 150 over non-null only", snapshot.Media.AvgDurationSeconds)
		}
	})
}

func seed(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()
	now := time.Now().UTC()

	// Users: active student (verified), active instructor (verified, 2FA),
	// blocked admin, deleted student.
	insertUser(t, ctx, pool, "u-active-student", "student", false, false, true, false, now)
	insertUser(t, ctx, pool, "u-active-instr", "instructor", false, false, true, true, now)
	insertUser(t, ctx, pool, "u-blocked-admin", "admin", false, true, false, false, now)
	insertUser(t, ctx, pool, "u-deleted", "student", true, false, false, false, now)

	instructorID := userUUID("u-active-instr")

	exec(t, ctx, pool, `INSERT INTO courses(id,slug,title,summary,description,status,created_at,updated_at,published_at,instructor_id)
		VALUES ($1,'c-draft','Draft course','s','d','draft',$4,$4,NULL,$5),
		       ($2,'c-pub','Published course','s','d','published',$4,$4,$4,$5),
		       ($3,'c-arch','Archived course','s','d','archived',$4,$4,$4,$5)`,
		courseUUID("c-draft"), courseUUID("c-pub"), courseUUID("c-arch"), now, instructorID)

	studentID := userUUID("u-active-student")
	// Grants: one per state, each a distinct (user,course) pair.
	exec(t, ctx, pool, `INSERT INTO course_grants(id,user_id,course_id,source,order_id,state,granted_at,granted_by)
		VALUES (gen_random_uuid(),$1,$2,'admin',NULL,'active',$4,NULL),
		       (gen_random_uuid(),$1,$3,'admin',NULL,'blocked',$4,NULL),
		       (gen_random_uuid(),$1,$5,'admin',NULL,'revoked',$4,NULL)`,
		studentID, courseUUID("c-pub"), courseUUID("c-arch"), now, courseUUID("c-draft"))

	// Orders: two paid (5000 + 7000), one pending, one failed.
	exec(t, ctx, pool, `INSERT INTO orders(id,user_id,course_id,amount_minor,currency,status,idempotency_key,created_at,paid_at)
		VALUES (gen_random_uuid(),$1,$2,5000,'UAH','paid','k1',$3,$3),
		       (gen_random_uuid(),$1,$2,7000,'UAH','paid','k2',$3,$3),
		       (gen_random_uuid(),$1,$2,3000,'UAH','pending','k3',$3,NULL),
		       (gen_random_uuid(),$1,$2,9000,'UAH','failed','k4',$3,NULL)`,
		studentID, courseUUID("c-pub"), now)

	// Reports: one open, one resolved (distinct pairs to respect the unique index).
	exec(t, ctx, pool, `INSERT INTO participant_reports(id,course_id,reporter_id,target_id,kind,reason,state,created_at)
		VALUES (gen_random_uuid(),$1,$2,$3,'on_student','bad','open',$4),
		       (gen_random_uuid(),$1,$3,$2,'on_student','bad','resolved',$4)`,
		courseUUID("c-pub"), studentID, instructorID, now)

	// Video assets: 2 ready with durations (100,200), 1 processing (null
	// duration), 1 failed (null duration). lesson_id left NULL (nullable FK).
	exec(t, ctx, pool, `INSERT INTO video_assets(id,lesson_id,generation,state,declared_bytes,received_bytes,original_name,renditions,duration_seconds,failure_reason,expires_at,created_at,updated_at)
		VALUES (gen_random_uuid(),NULL,1,'ready',1000,1000,'a.mp4','{"360p"}',100,NULL,$1,$1,$1),
		       (gen_random_uuid(),NULL,1,'ready',1000,1000,'b.mp4','{"360p"}',200,NULL,$1,$1,$1),
		       (gen_random_uuid(),NULL,1,'processing',1000,0,'c.mp4','{}',NULL,NULL,$1,$1,$1),
		       (gen_random_uuid(),NULL,1,'failed',1000,0,'d.mp4','{}',NULL,'transcode failed',$1,$1,$1)`,
		now)
}

func insertUser(t *testing.T, ctx context.Context, pool *pgxpool.Pool, slug, role string, deleted, blocked, verified, twoFactor bool, now time.Time) {
	t.Helper()
	var deletedAt, disabledAt *time.Time
	if blocked || deleted {
		// A deleted account must also be disabled (users_deleted_implies_disabled).
		d := now
		disabledAt = &d
	}
	if deleted {
		d := now
		deletedAt = &d
	}
	exec(t, ctx, pool, `INSERT INTO users(id,email,display_name,role,password_hash,created_at,updated_at,disabled_at,deleted_at,email_verified,totp_enabled)
		VALUES ($1,$2,$3,$4,$5,$6,$6,$7,$8,$9,$10)`,
		userUUID(slug), slug+"@example.test", slug, role, validArgon2Hash, now, disabledAt, deletedAt, verified, twoFactor)
}

func exec(t *testing.T, ctx context.Context, pool *pgxpool.Pool, sqlText string, args ...any) {
	t.Helper()
	if _, err := pool.Exec(ctx, sqlText, args...); err != nil {
		t.Fatalf("seed exec failed: %v", err)
	}
}

// userUUID / courseUUID resolve a readable slug to a fixed, valid UUID so the
// seed reads like its fixtures. Unknown slugs fail the test loudly.
var fixtureUUIDs = map[string]string{
	"u-active-student": "10000000-0000-4000-8000-000000000001",
	"u-active-instr":   "10000000-0000-4000-8000-000000000002",
	"u-blocked-admin":  "10000000-0000-4000-8000-000000000003",
	"u-deleted":        "10000000-0000-4000-8000-000000000004",
	"c-draft":          "50000000-0000-4000-8000-000000000001",
	"c-pub":            "50000000-0000-4000-8000-000000000002",
	"c-arch":           "50000000-0000-4000-8000-000000000003",
	// Leaderboard fixtures (instructors_integration_test.go).
	"ins-a": "10000000-0000-4000-8000-0000000000a1",
	"ins-b": "10000000-0000-4000-8000-0000000000a2",
	"ins-c": "10000000-0000-4000-8000-0000000000a3",
	"buy-1": "10000000-0000-4000-8000-0000000000b1",
	"buy-2": "10000000-0000-4000-8000-0000000000b2",
	"ca":    "50000000-0000-4000-8000-0000000000c1",
	"cb":    "50000000-0000-4000-8000-0000000000c2",
	"cc":    "50000000-0000-4000-8000-0000000000c3",
	// Timeseries fixtures (timeseries_integration_test.go).
	"ts-a":     "10000000-0000-4000-8000-0000000000d1",
	"ts-b":     "10000000-0000-4000-8000-0000000000d2",
	"ts-c":     "10000000-0000-4000-8000-0000000000d3",
	"ts-d":     "10000000-0000-4000-8000-0000000000d4",
	"ts-pre":   "10000000-0000-4000-8000-0000000000d5",
	"ts-post":  "10000000-0000-4000-8000-0000000000d6",
	"ts-instr": "10000000-0000-4000-8000-0000000000d7",
	"ts-buyer": "10000000-0000-4000-8000-0000000000d8",
	"ts-crs":   "50000000-0000-4000-8000-0000000000e1",
}

func userUUID(slug string) string   { return fixtureUUIDs[slug] }
func courseUUID(slug string) string { return fixtureUUIDs[slug] }
