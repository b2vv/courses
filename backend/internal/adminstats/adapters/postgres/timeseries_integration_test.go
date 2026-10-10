//go:build integration

package postgres_test

import (
	"context"
	"testing"
	"time"

	adminstatspostgres "courses/internal/adminstats/adapters/postgres"
	"courses/internal/adminstats/domain"
)

func TestAdminStatsTimeseriesAgainstPostgreSQL(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	pool := newTestPool(t, ctx)
	repository := adminstatspostgres.NewRepository(pool)

	// Fixed UTC window so day buckets never drift with the local zone.
	from := time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)
	to := from.AddDate(0, 0, 3) // [from, from+3d)

	// Signups: four accounts inside the window across three day buckets, two
	// outside (one before, one exactly at the exclusive upper bound).
	insertUser(t, ctx, pool, "ts-a", "student", false, false, true, false, from)
	insertUser(t, ctx, pool, "ts-b", "student", false, false, true, false, from.Add(24*time.Hour))
	insertUser(t, ctx, pool, "ts-c", "student", false, false, true, false, from.Add(25*time.Hour))
	insertUser(t, ctx, pool, "ts-d", "student", false, false, true, false, from.Add(48*time.Hour))
	insertUser(t, ctx, pool, "ts-pre", "student", false, false, true, false, from.Add(-24*time.Hour))
	insertUser(t, ctx, pool, "ts-post", "student", false, false, true, false, to) // == to, excluded

	// Revenue fixture: an instructor, a course, a buyer whose created_at sits
	// before the window — created far earlier so it cannot skew signups.
	insertUser(t, ctx, pool, "ts-instr", "instructor", false, false, true, false, from.AddDate(0, 0, -30))
	insertUser(t, ctx, pool, "ts-buyer", "student", false, false, true, false, from.AddDate(0, 0, -30))
	insertCourse(t, ctx, pool, "ts-crs", "ts-instr", from.AddDate(0, 0, -30))

	t.Run("SignupsSumEqualsCountInWindow", func(t *testing.T) {
		points, err := repository.Timeseries(ctx, domain.MetricSignups, domain.BucketDay, from, to)
		if err != nil {
			t.Fatalf("signups timeseries: %v", err)
		}
		var sum float64
		for i, p := range points {
			if p.Value <= 0 {
				t.Fatalf("bucket %d has non-positive value %v", i, p.Value)
			}
			if !p.BucketStart.Before(to) || p.BucketStart.Before(from) {
				t.Fatalf("bucket %d start %v outside [%v,%v)", i, p.BucketStart, from, to)
			}
			sum += p.Value
		}
		// Scenario 3: the series sums to the number of signups in the window,
		// checked against the table directly (robust to the revenue fixtures).
		var expected int
		if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE created_at >= $1 AND created_at < $2`, from, to).Scan(&expected); err != nil {
			t.Fatal(err)
		}
		if expected == 0 {
			t.Fatal("fixture seeded no in-window signups")
		}
		if sum != float64(expected) {
			t.Fatalf("signups sum = %v, COUNT in window = %d", sum, expected)
		}
	})

	t.Run("RevenueBucketsByPaidAtNotCreatedAt", func(t *testing.T) {
		// Order created two days before the window but paid one hour into it:
		// it belongs to the window by paid_at, not created_at.
		exec(t, ctx, pool, `INSERT INTO orders(id,user_id,course_id,amount_minor,currency,status,idempotency_key,created_at,paid_at)
			VALUES (gen_random_uuid(),$1,$2,7777,'UAH','paid','ts-rev',$3,$4)`,
			userUUID("ts-buyer"), courseUUID("ts-crs"), from.AddDate(0, 0, -2), from.Add(time.Hour))

		revenue, err := repository.Timeseries(ctx, domain.MetricRevenue, domain.BucketDay, from, to)
		if err != nil {
			t.Fatalf("revenue timeseries: %v", err)
		}
		if len(revenue) != 1 {
			t.Fatalf("revenue buckets = %d, want 1 (paid_at in window): %+v", len(revenue), revenue)
		}
		if revenue[0].Value != 7777 {
			t.Fatalf("revenue value = %v, want 7777", revenue[0].Value)
		}

		// Scenario 4 contrast: the same order by created_at is outside the
		// window, so the orders-count series does not see it.
		orders, err := repository.Timeseries(ctx, domain.MetricOrders, domain.BucketDay, from, to)
		if err != nil {
			t.Fatalf("orders timeseries: %v", err)
		}
		if len(orders) != 0 {
			t.Fatalf("orders buckets = %d, want 0 (created_at before window): %+v", len(orders), orders)
		}
	})

	t.Run("EmptyWindowIsNonNilEmpty", func(t *testing.T) {
		quietFrom := from.AddDate(1, 0, 0)
		quietTo := quietFrom.AddDate(0, 0, 3)
		points, err := repository.Timeseries(ctx, domain.MetricSignups, domain.BucketDay, quietFrom, quietTo)
		if err != nil {
			t.Fatalf("empty-window timeseries: %v", err)
		}
		// Scenario 6: an empty window is [] (non-nil), rendered as [] not null.
		if points == nil {
			t.Fatal("empty window returned nil slice, want non-nil empty []")
		}
		if len(points) != 0 {
			t.Fatalf("empty window returned %d points, want 0: %+v", len(points), points)
		}
	})
}
