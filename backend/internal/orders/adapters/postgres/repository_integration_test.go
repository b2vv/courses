//go:build integration

package postgres_test

import (
	"bytes"
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	ordershttp "courses/internal/orders/adapters/http"
	orderspostgres "courses/internal/orders/adapters/postgres"
	"courses/internal/orders/application"
	"courses/internal/orders/domain"
	"courses/internal/orders/ports"
	"courses/platform/httpserver"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	postgrescontainer "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	integrationUserID   = "40000000-0000-4000-8000-000000000001"
	integrationAdminID  = "40000000-0000-4000-8000-000000000003"
	integrationCourseID = "50000000-0000-4000-8000-000000000001"
	integrationFreeID   = "50000000-0000-4000-8000-000000000002"
)

func TestOrdersRepositoryAgainstPostgreSQL(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	container, err := postgrescontainer.Run(ctx, "postgres:17.5", postgrescontainer.WithDatabase("courses_test"), postgrescontainer.WithUsername("courses"), postgrescontainer.WithPassword("ephemeral-test-only"), testcontainers.WithWaitStrategy(wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(3*time.Minute)))
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
	defer pool.Close()
	seedOrders(t, ctx, pool)

	repository := orderspostgres.NewRepository(pool)
	ids := []string{
		"60000000-0000-4000-8000-000000000001", "70000000-0000-4000-8000-000000000001",
		"70000000-0000-4000-8000-000000000002", "70000000-0000-4000-8000-000000000003",
		"70000000-0000-4000-8000-000000000004", "70000000-0000-4000-8000-000000000005",
		"60000000-0000-4000-8000-000000000002", "70000000-0000-4000-8000-000000000006",
		"70000000-0000-4000-8000-000000000007", "70000000-0000-4000-8000-000000000008",
	}
	newID := func() string { id := ids[0]; ids = ids[1:]; return id }
	now := time.Date(2026, 9, 30, 9, 0, 0, 0, time.UTC)
	service := application.New(repository, repository, repository, func() time.Time { return now }, newID)

	order, err := service.StartPurchase(ctx, application.Actor{UserID: integrationUserID}, integrationCourseID, "purchase-1")
	if err != nil {
		t.Fatal(err)
	}
	repeated, err := service.StartPurchase(ctx, application.Actor{UserID: integrationUserID}, integrationCourseID, "purchase-1")
	if err != nil || repeated.ID() != order.ID() {
		t.Fatalf("idempotent repeat=%q error=%v", repeated.ID(), err)
	}
	// The service refuses an order for another course coming back from Create (unit:
	// "same key raced for another course is refused"). That guard rests on this
	// repository behaviour: a unique (user, key) conflict hands back the row that
	// won, even when the losing insert was for a different course.
	raced, err := domain.NewOrder("60000000-0000-4000-8000-000000000077", integrationUserID, integrationFreeID, order.Amount(), "purchase-1", now)
	if err != nil {
		t.Fatal(err)
	}
	winner, err := repository.Create(ctx, raced)
	if err != nil {
		t.Fatalf("conflicting create: %v", err)
	}
	if winner.ID() != order.ID() || winner.CourseID() != integrationCourseID {
		t.Fatalf("conflict returned id=%s course=%s, want the existing order %s for %s", winner.ID(), winner.CourseID(), order.ID(), integrationCourseID)
	}
	var sameKey int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM orders WHERE user_id=$1 AND idempotency_key='purchase-1'`, integrationUserID).Scan(&sameKey); err != nil || sameKey != 1 {
		t.Fatalf("orders under the key=%d err=%v, want exactly the winner", sameKey, err)
	}
	event := domain.PaymentEvent{Provider: "provider", EventID: "paid-1", OrderID: order.ID(), Type: domain.PaymentSucceeded, Amount: order.Amount(), ReceivedAt: now.Add(time.Minute)}
	if err := service.ApplyPaymentEvent(ctx, event); err != nil {
		t.Fatal(err)
	}
	if err := service.ApplyPaymentEvent(ctx, event); err != nil {
		t.Fatalf("duplicate event: %v", err)
	}
	status, err := service.PurchaseStatus(ctx, application.Actor{UserID: integrationUserID}, order.ID())
	if err != nil || status.Status() != domain.OrderPaid {
		t.Fatalf("status=%q error=%v", status.Status(), err)
	}
	var grants, events, audits int
	if err := pool.QueryRow(ctx, `SELECT (SELECT count(*) FROM course_grants WHERE order_id=$1),(SELECT count(*) FROM payment_events WHERE event_id='paid-1'),(SELECT count(*) FROM grant_audit WHERE payment_event_id IS NOT NULL)`, order.ID()).Scan(&grants, &events, &audits); err != nil {
		t.Fatal(err)
	}
	if grants != 1 || events != 1 || audits != 1 {
		t.Fatalf("purchase grants=%d events=%d audits=%d", grants, events, audits)
	}

	unknown := event
	unknown.EventID = "unknown"
	unknown.OrderID = "60000000-0000-4000-8000-000000000099"
	if err := service.ApplyPaymentEvent(ctx, unknown); err != domain.ErrPaymentEventUnmatched {
		t.Fatalf("unknown event error=%v", err)
	}
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM payment_events WHERE event_id='unknown'`).Scan(&events); err != nil || events != 0 {
		t.Fatalf("unknown event rows=%d error=%v", events, err)
	}
	mismatchOrder, err := service.StartPurchase(ctx, application.Actor{UserID: integrationAdminID}, integrationCourseID, "purchase-mismatch")
	if err != nil {
		t.Fatal(err)
	}
	wrongMoney, err := domain.NewMoney("70000", "UAH")
	if err != nil {
		t.Fatal(err)
	}
	mismatch := domain.PaymentEvent{Provider: "provider", EventID: "mismatch", OrderID: mismatchOrder.ID(), Type: domain.PaymentSucceeded, Amount: wrongMoney, ReceivedAt: now.Add(2 * time.Minute)}
	if err := service.ApplyPaymentEvent(ctx, mismatch); err != domain.ErrPaymentEventRejected {
		t.Fatalf("mismatch error=%v", err)
	}
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM payment_events WHERE event_id='mismatch'`).Scan(&events); err != nil || events != 1 {
		t.Fatalf("mismatch rows=%d error=%v", events, err)
	}

	t.Run("failed payment is terminal and idempotent", func(t *testing.T) {
		failedIDs := []string{"70000000-0000-0000-0000-000000000010", "70000000-0000-0000-0000-000000000011", "70000000-0000-0000-0000-000000000012"}
		failedID := func() string {
			id := failedIDs[0]
			failedIDs = failedIDs[1:]
			return id
		}
		failedService := application.New(repository, repository, repository, func() time.Time { return now }, failedID)
		var failedUser string
		if err := pool.QueryRow(ctx, `INSERT INTO users(id,email,display_name,role,password_hash) VALUES(gen_random_uuid(),'failed@example.test','Failed','student','$argon2id$v=19$m=65536,t=3,p=2$c3R1ZGVudC1kZW1vLXNhbHQ$5S1MhXX3wgDMnIKuygFRfUNVvTtRktPECxV3ZUBg0is') RETURNING id::text`).Scan(&failedUser); err != nil {
			t.Fatal(err)
		}
		failedOrder, err := domain.NewOrder("60000000-0000-0000-0000-000000000010", failedUser, integrationCourseID, order.Amount(), "failed-payment", now)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := repository.Create(ctx, failedOrder); err != nil {
			t.Fatal(err)
		}
		failedEvent := domain.PaymentEvent{Provider: "wayforpay", EventID: "failed-1", OrderID: failedOrder.ID(), Type: domain.PaymentFailed, Amount: failedOrder.Amount(), ReceivedAt: now.Add(3 * time.Minute)}
		if err := failedService.ApplyPaymentEvent(ctx, failedEvent); err != nil {
			t.Fatal(err)
		}
		if err := failedService.ApplyPaymentEvent(ctx, failedEvent); err != nil {
			t.Fatalf("failed replay: %v", err)
		}
		status, err := repository.Get(ctx, failedUser, failedOrder.ID())
		if err != nil || status.Status() != domain.OrderFailed {
			t.Fatalf("status=%q error=%v", status.Status(), err)
		}
		var failedEvents, grants int
		if err := pool.QueryRow(ctx, `SELECT count(*) FROM payment_events WHERE provider='wayforpay' AND event_id='failed-1'`).Scan(&failedEvents); err != nil {
			t.Fatal(err)
		}
		if err := pool.QueryRow(ctx, `SELECT count(*) FROM course_grants WHERE order_id=$1`, failedOrder.ID()).Scan(&grants); err != nil {
			t.Fatal(err)
		}
		if failedEvents != 1 || grants != 0 {
			t.Fatalf("failed events=%d grants=%d", failedEvents, grants)
		}
		paidAfterFailed := failedEvent
		paidAfterFailed.EventID = "paid-after-failed"
		paidAfterFailed.Type = domain.PaymentSucceeded
		if err := failedService.ApplyPaymentEvent(ctx, paidAfterFailed); err != domain.ErrPaymentEventRejected {
			t.Fatalf("paid after failed error=%v", err)
		}
		if err := pool.QueryRow(ctx, `SELECT count(*) FROM course_grants WHERE order_id=$1`, failedOrder.ID()).Scan(&grants); err != nil {
			t.Fatal(err)
		}
		if grants != 0 {
			t.Fatalf("paid after failed granted access: %d", grants)
		}

		var beforeEventRows, afterEventRows int
		if err := pool.QueryRow(ctx, `SELECT count(*) FROM payment_events`).Scan(&beforeEventRows); err != nil {
			t.Fatal(err)
		}
		beforeOrders, beforeGrants := 0, grants
		if err := pool.QueryRow(ctx, `SELECT count(*) FROM orders WHERE id=$1`, failedOrder.ID()).Scan(&beforeOrders); err != nil {
			t.Fatal(err)
		}
		invalidHandler := ordershttp.NewHandler(failedService, invalidCallbackVerifier{})
		server := httpserver.NewHandler(httpserver.Options{Service: "test", Register: func(mux *http.ServeMux) {
			invalidHandler.Register(mux, func(next http.Handler) http.Handler { return next }, func(next http.Handler) http.Handler { return next })
		}})
		request := httptest.NewRequest(http.MethodPost, "/api/v1/payments/wayforpay/callback", bytes.NewReader([]byte(`{"tampered":true}`)))
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()
		server.ServeHTTP(response, request)
		if response.Code != http.StatusUnauthorized {
			t.Fatalf("bad signature status=%d body=%s", response.Code, response.Body.String())
		}
		if err := pool.QueryRow(ctx, `SELECT count(*) FROM payment_events WHERE provider='wayforpay' AND event_id='failed-1'`).Scan(&failedEvents); err != nil {
			t.Fatal(err)
		}
		if err := pool.QueryRow(ctx, `SELECT count(*) FROM payment_events`).Scan(&afterEventRows); err != nil {
			t.Fatal(err)
		}
		if failedEvents != 1 || afterEventRows != beforeEventRows || beforeOrders != 1 || beforeGrants != 0 {
			t.Fatalf("bad signature changed rows events=%d total_before=%d total_after=%d orders=%d grants=%d", failedEvents, beforeEventRows, afterEventRows, beforeOrders, beforeGrants)
		}
	})

	grant, err := service.GrantCourse(ctx, application.Actor{UserID: integrationAdminID, Administrator: true}, integrationAdminID, integrationCourseID, "manual access")
	if err != nil || grant.Source() != domain.GrantAdmin {
		t.Fatalf("admin grant=%+v error=%v", grant, err)
	}

	t.Run("paid event settles when active admin grant already exists", func(t *testing.T) {
		var userID string
		if err := pool.QueryRow(ctx, `INSERT INTO users(id,email,display_name,role,password_hash) VALUES
(gen_random_uuid(),'paid-with-admin-grant@example.test','Paid with admin grant','student','$argon2id$v=19$m=65536,t=3,p=2$c3R1ZGVudC1kZW1vLXNhbHQ$5S1MhXX3wgDMnIKuygFRfUNVvTtRktPECxV3ZUBg0is') RETURNING id::text`).Scan(&userID); err != nil {
			t.Fatal(err)
		}
		order, err := service.StartPurchase(ctx, application.Actor{UserID: userID}, integrationCourseID, "purchase-with-admin-grant")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := service.GrantCourse(ctx, application.Actor{UserID: integrationAdminID, Administrator: true}, userID, integrationCourseID, "issued while payment was pending"); err != nil {
			t.Fatal(err)
		}
		event := domain.PaymentEvent{Provider: "provider", EventID: "paid-with-admin-grant", OrderID: order.ID(), Type: domain.PaymentSucceeded, Amount: order.Amount(), ReceivedAt: now.Add(3 * time.Minute)}
		if err := service.ApplyPaymentEvent(ctx, event); err != nil {
			t.Fatalf("ApplyPaymentEvent with existing active grant: %v", err)
		}
		status, err := service.PurchaseStatus(ctx, application.Actor{UserID: userID}, order.ID())
		if err != nil || status.Status() != domain.OrderPaid {
			t.Fatalf("status=%q error=%v", status.Status(), err)
		}
		var activeGrants, events, paymentAudits int
		if err := pool.QueryRow(ctx, `SELECT
  (SELECT count(*) FROM course_grants WHERE user_id=$1::text::uuid AND course_id=$2::text::uuid AND state='active'),
  (SELECT count(*) FROM payment_events WHERE provider='provider' AND event_id='paid-with-admin-grant'),
  (SELECT count(*) FROM grant_audit AS audit JOIN payment_events AS event ON event.id=audit.payment_event_id WHERE event.provider='provider' AND event.event_id='paid-with-admin-grant')`, userID, integrationCourseID).Scan(&activeGrants, &events, &paymentAudits); err != nil {
			t.Fatal(err)
		}
		if activeGrants != 1 || events != 1 || paymentAudits != 0 {
			t.Fatalf("active grants=%d events=%d payment audits=%d", activeGrants, events, paymentAudits)
		}
	})

	var freeUpdated time.Time
	if err := pool.QueryRow(ctx, `SELECT updated_at FROM courses WHERE id=$1`, integrationFreeID).Scan(&freeUpdated); err != nil {
		t.Fatal(err)
	}
	if err := service.GrandfatherOnPriceSet(ctx, integrationFreeID, freeUpdated, 70000, "UAH", freeUpdated.Add(time.Microsecond)); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM course_grants WHERE course_id=$1 AND source='grandfather'`, integrationFreeID).Scan(&grants); err != nil || grants != 1 {
		t.Fatalf("grandfather grants=%d error=%v", grants, err)
	}
}

type invalidCallbackVerifier struct{}

func (invalidCallbackVerifier) VerifyCallback([]byte, time.Time) (ports.Callback, error) {
	return ports.Callback{}, domain.ErrPaymentEventRejected
}
func (invalidCallbackVerifier) Acknowledge(string, time.Time) []byte { return nil }

func seedOrders(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()
	_, err := pool.Exec(ctx, `INSERT INTO users(id,email,display_name,role,password_hash) VALUES
('40000000-0000-4000-8000-000000000001','student@example.test','Student','student','$argon2id$v=19$m=65536,t=3,p=2$c3R1ZGVudC1kZW1vLXNhbHQ$5S1MhXX3wgDMnIKuygFRfUNVvTtRktPECxV3ZUBg0is'),
('40000000-0000-4000-8000-000000000003','admin@example.test','Admin','admin','$argon2id$v=19$m=65536,t=3,p=2$c3R1ZGVudC1kZW1vLXNhbHQ$5S1MhXX3wgDMnIKuygFRfUNVvTtRktPECxV3ZUBg0is');
INSERT INTO courses(id,slug,title,summary,description,status,instructor_id,published_at,price_minor,currency,created_at,updated_at) VALUES
('50000000-0000-4000-8000-000000000001','paid','Paid','Paid summary','Paid description','published','40000000-0000-4000-8000-000000000003','2026-09-30T08:00:00Z',50000,'UAH','2026-09-30T08:00:00Z','2026-09-30T08:00:00Z'),
('50000000-0000-4000-8000-000000000002','free','Free','Free summary','Free description','published','40000000-0000-4000-8000-000000000003','2026-09-30T08:00:00Z',NULL,NULL,'2026-09-30T08:00:00Z','2026-09-30T08:00:00Z');
INSERT INTO course_modules(id,course_id,title,position) VALUES('51000000-0000-4000-8000-000000000001','50000000-0000-4000-8000-000000000002','Module',0);
INSERT INTO course_lessons(id,module_id,title,description,position) VALUES('52000000-0000-4000-8000-000000000001','51000000-0000-4000-8000-000000000001','Lesson','',0);
INSERT INTO lesson_progress(user_id,lesson_id,position_seconds,client_seq,updated_at) VALUES('40000000-0000-4000-8000-000000000001','52000000-0000-4000-8000-000000000001',1,1,'2026-09-30T08:30:00Z')`)
	if err != nil {
		t.Fatal(err)
	}
}
