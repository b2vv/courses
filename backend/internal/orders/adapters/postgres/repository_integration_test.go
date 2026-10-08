//go:build integration

package postgres_test

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
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
	containers "courses/internal/testsupport/containers"
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
	service := application.New(repository, repository, repository, func() time.Time { return now }, newID, nil)

	t.Run("B25 repeat checkout uses a new order and preserves independent outcomes", func(t *testing.T) {
		const courseID = integrationCourseID
		for _, user := range []struct{ id, email string }{
			{"40000000-0000-4000-8000-000000000020", "b25-declined@example.test"},
			{"40000000-0000-4000-8000-000000000021", "b25-approved@example.test"},
		} {
			_, err := pool.Exec(ctx, `INSERT INTO users(id,email,display_name,role,password_hash) VALUES($1,$2,'B25','student','$argon2id$v=19$m=65536,t=3,p=2$c3R1ZGVudC1kZW1vLXNhbHQ$5S1MhXX3wgDMnIKuygFRfUNVvTtRktPECxV3ZUBg0is')`, user.id, user.email)
			if err != nil {
				t.Fatal(err)
			}
		}
		checkoutProvider := &integrationPaymentProvider{}
		declinedID := subtestIDs("61")
		declinedService := application.New(repository, repository, repository, func() time.Time { return now }, declinedID, checkoutProvider)
		old, err := declinedService.StartPurchase(ctx, application.Actor{UserID: "40000000-0000-4000-8000-000000000020", EmailVerified: true}, courseID, "b25-old")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := declinedService.Checkout(ctx, application.Actor{UserID: old.UserID(), EmailVerified: true}, old.ID(), ports.Buyer{}, ports.CheckoutURLs{}); err != nil {
			t.Fatal(err)
		}
		if _, err := declinedService.Checkout(ctx, application.Actor{UserID: old.UserID(), EmailVerified: true}, old.ID(), ports.Buyer{}, ports.CheckoutURLs{}); err != domain.ErrCheckoutAlreadyOpened {
			t.Fatalf("second checkout error=%v", err)
		}
		newOrder, err := declinedService.StartPurchase(ctx, application.Actor{UserID: old.UserID(), EmailVerified: true}, courseID, "b25-new")
		if err != nil || newOrder.ID() == old.ID() {
			t.Fatalf("new order=%q old=%q error=%v", newOrder.ID(), old.ID(), err)
		}
		if err := declinedService.ApplyPaymentEvent(ctx, domain.PaymentEvent{Provider: "provider", EventID: "b25-declined", OrderID: old.ID(), Type: domain.PaymentFailed, Amount: old.Amount(), ReceivedAt: now.Add(time.Minute)}); err != nil {
			t.Fatal(err)
		}
		status, err := repository.Get(ctx, old.UserID(), newOrder.ID())
		if err != nil || status.Status() != domain.OrderPending {
			t.Fatalf("new order after late decline status=%q error=%v", status.Status(), err)
		}

		approvedID := subtestIDs("62")
		approvedService := application.New(repository, repository, repository, func() time.Time { return now }, approvedID, checkoutProvider)
		approvedOld, err := approvedService.StartPurchase(ctx, application.Actor{UserID: "40000000-0000-4000-8000-000000000021", EmailVerified: true}, courseID, "b25-approved-old")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := approvedService.StartPurchase(ctx, application.Actor{UserID: approvedOld.UserID(), EmailVerified: true}, courseID, "b25-approved-new"); err != nil {
			t.Fatal(err)
		}
		if err := approvedService.ApplyPaymentEvent(ctx, domain.PaymentEvent{Provider: "provider", EventID: "b25-approved-declined", OrderID: approvedOld.ID(), Type: domain.PaymentFailed, Amount: approvedOld.Amount(), ReceivedAt: now.Add(90 * time.Second)}); err != nil {
			t.Fatalf("decline old order error=%v", err)
		}
		if err := approvedService.ApplyPaymentEvent(ctx, domain.PaymentEvent{Provider: "provider", EventID: "b25-approved", OrderID: approvedOld.ID(), Type: domain.PaymentSucceeded, Amount: approvedOld.Amount(), ReceivedAt: now.Add(2 * time.Minute)}); err != domain.ErrPaymentEventRejected {
			t.Fatalf("paid after failed error=%v", err)
		}
		var refundRequired bool
		if err := pool.QueryRow(ctx, `SELECT refund_required FROM orders WHERE id=$1`, approvedOld.ID()).Scan(&refundRequired); err != nil {
			t.Fatal(err)
		}
		if !refundRequired {
			t.Fatal("rejected paid event did not mark failed order for refund")
		}
		var grants int
		if err := pool.QueryRow(ctx, `SELECT count(*) FROM course_grants WHERE order_id=$1`, approvedOld.ID()).Scan(&grants); err != nil {
			t.Fatal(err)
		}
		if grants != 0 {
			t.Fatalf("rejected paid event granted access: %d", grants)
		}
		var pending int
		if err := pool.QueryRow(ctx, `SELECT count(*) FROM orders WHERE user_id=$1 AND idempotency_key='b25-approved-new' AND status='pending'`, approvedOld.UserID()).Scan(&pending); err != nil || pending != 1 {
			t.Fatalf("new pending orders=%d error=%v", pending, err)
		}
	})

	t.Run("B25 duplicate payment marks the paid order for refund", func(t *testing.T) {
		const userID = "40000000-0000-4000-8000-000000000022"
		if _, err := pool.Exec(ctx, `INSERT INTO users(id,email,display_name,role,password_hash) VALUES($1,'b25-refund@example.test','B25 refund','student','$argon2id$v=19$m=65536,t=3,p=2$c3R1ZGVudC1kZW1vLXNhbHQ$5S1MhXX3wgDMnIKuygFRfUNVvTtRktPECxV3ZUBg0is')`, userID); err != nil {
			t.Fatal(err)
		}
		newID := subtestIDs("63")
		svc := application.New(repository, repository, repository, func() time.Time { return now }, newID, nil)
		first, err := svc.StartPurchase(ctx, application.Actor{UserID: userID, EmailVerified: true}, integrationCourseID, "b25-refund-one")
		if err != nil {
			t.Fatal(err)
		}
		second, err := svc.StartPurchase(ctx, application.Actor{UserID: userID, EmailVerified: true}, integrationCourseID, "b25-refund-two")
		if err != nil {
			t.Fatal(err)
		}
		for _, item := range []struct {
			id    string
			order domain.Order
		}{{"b25-refund-first", first}, {"b25-refund-second", second}} {
			if err := svc.ApplyPaymentEvent(ctx, domain.PaymentEvent{Provider: "provider", EventID: item.id, OrderID: item.order.ID(), Type: domain.PaymentSucceeded, Amount: item.order.Amount(), ReceivedAt: now.Add(3 * time.Minute)}); err != nil {
				t.Fatal(err)
			}
		}
		var grants, refunds int
		if err := pool.QueryRow(ctx, `SELECT (SELECT count(*) FROM course_grants WHERE user_id=$1 AND course_id=$2 AND state='active'), (SELECT count(*) FROM orders WHERE user_id=$1 AND course_id=$2 AND status='paid' AND refund_required)`, userID, integrationCourseID).Scan(&grants, &refunds); err != nil {
			t.Fatal(err)
		}
		if grants != 1 || refunds != 1 {
			t.Fatalf("active grants=%d refund orders=%d", grants, refunds)
		}
		var secondRefundRequired bool
		if err := pool.QueryRow(ctx, `SELECT refund_required FROM orders WHERE id=$1`, second.ID()).Scan(&secondRefundRequired); err != nil {
			t.Fatal(err)
		}
		if !secondRefundRequired {
			t.Fatalf("second paid order %s was not marked for refund", second.ID())
		}
	})

	t.Run("B1 foreign paid amount marks pending order for refund", func(t *testing.T) {
		const userID = "40000000-0000-4000-0000-000000000023"
		if _, err := pool.Exec(ctx, `INSERT INTO users(id,email,display_name,role,password_hash) VALUES($1,'b25-mismatch@example.test','B25 mismatch','student','$argon2id$v=19$m=65536,t=3,p=2$c3R1ZGVudC1kZW1vLXNhbHQ$5S1MhXX3wgDMnIKuygFRfUNVvTtRktPECxV3ZUBg0is')`, userID); err != nil {
			t.Fatal(err)
		}
		ids := subtestIDs("64")
		svc := application.New(repository, repository, repository, func() time.Time { return now }, ids, nil)
		order, err := svc.StartPurchase(ctx, application.Actor{UserID: userID, EmailVerified: true}, integrationCourseID, "b25-mismatch")
		if err != nil {
			t.Fatal(err)
		}
		wrongAmount, err := domain.NewMoney("70000", "UAH")
		if err != nil {
			t.Fatal(err)
		}
		if err := svc.ApplyPaymentEvent(ctx, domain.PaymentEvent{Provider: "provider", EventID: "b25-mismatch", OrderID: order.ID(), Type: domain.PaymentSucceeded, Amount: wrongAmount, ReceivedAt: now.Add(4 * time.Minute)}); err != domain.ErrPaymentEventRejected {
			t.Fatalf("foreign amount error=%v", err)
		}
		var status string
		var refundRequired bool
		if err := pool.QueryRow(ctx, `SELECT status,refund_required FROM orders WHERE id=$1`, order.ID()).Scan(&status, &refundRequired); err != nil {
			t.Fatal(err)
		}
		if status != string(domain.OrderPending) || !refundRequired {
			t.Fatalf("status=%s refund_required=%v, want pending/true", status, refundRequired)
		}
		var grants int
		if err := pool.QueryRow(ctx, `SELECT count(*) FROM course_grants WHERE order_id=$1`, order.ID()).Scan(&grants); err != nil {
			t.Fatal(err)
		}
		if grants != 0 {
			t.Fatalf("foreign amount granted access: %d", grants)
		}
	})

	t.Run("Pay17 price change preserves access and order snapshot", func(t *testing.T) {
		const pay17UserID = "40000000-0000-4000-8000-000000000010"
		const pay17CourseID = "50000000-0000-4000-8000-000000000010"
		if _, err := pool.Exec(ctx, `INSERT INTO users(id,email,display_name,role,password_hash) VALUES($1,'pay17@example.test','Pay17','student','$argon2id$v=19$m=65536,t=3,p=2$c3R1ZGVudC1kZW1vLXNhbHQ$5S1MhXX3wgDMnIKuygFRfUNVvTtRktPECxV3ZUBg0is')`, pay17UserID); err != nil {
			t.Fatal(err)
		}
		if _, err := pool.Exec(ctx, `INSERT INTO courses(id,slug,title,summary,description,status,instructor_id,published_at,price_minor,currency,created_at,updated_at)
VALUES($1,'pay17','Pay17','Summary','Description','published',$2,$3,50000,'UAH',$3,$3)`, pay17CourseID, integrationAdminID, now); err != nil {
			t.Fatal(err)
		}
		pay17IDs := []string{"60000000-0000-4000-8000-000000000010", "60000000-0000-4000-8000-000000000011", "60000000-0000-4000-8000-000000000012"}
		pay17ID := func() string {
			id := pay17IDs[0]
			pay17IDs = pay17IDs[1:]
			return id
		}
		pay17Service := application.New(repository, repository, repository, func() time.Time { return now }, pay17ID, nil)
		pay17Order, err := pay17Service.StartPurchase(ctx, application.Actor{UserID: pay17UserID, EmailVerified: true}, pay17CourseID, "pay17-purchase")
		if err != nil {
			t.Fatal(err)
		}
		if err := pay17Service.ApplyPaymentEvent(ctx, domain.PaymentEvent{Provider: "provider", EventID: "pay17-paid", OrderID: pay17Order.ID(), Type: domain.PaymentSucceeded, Amount: pay17Order.Amount(), ReceivedAt: now.Add(time.Minute)}); err != nil {
			t.Fatal(err)
		}
		if _, err := pool.Exec(ctx, `UPDATE courses SET price_minor=70000,currency='UAH',updated_at=$2 WHERE id=$1`, pay17CourseID, now.Add(2*time.Minute)); err != nil {
			t.Fatal(err)
		}
		var activeGrants int
		if err := pool.QueryRow(ctx, `SELECT count(*) FROM course_grants WHERE user_id=$1::text::uuid AND course_id=$2::text::uuid AND state='active'`, pay17UserID, pay17CourseID).Scan(&activeGrants); err != nil || activeGrants != 1 {
			t.Fatalf("active grants=%d error=%v", activeGrants, err)
		}
		var amount int64
		if err := pool.QueryRow(ctx, `SELECT amount_minor FROM orders WHERE id=$1`, pay17Order.ID()).Scan(&amount); err != nil {
			t.Fatal(err)
		}
		if amount != 50000 {
			t.Fatalf("order amount_minor=%d, want 50000 after price changed to 70000", amount)
		}
	})

	t.Run("Pay26 grandfather covers two students and repeats idempotently", func(t *testing.T) {
		const pay26CourseID = "50000000-0000-4000-8000-000000000011"
		const pay26ModuleID = "51000000-0000-4000-8000-000000000011"
		const pay26LessonID = "52000000-0000-4000-8000-000000000011"
		const pay26StudentOne = "40000000-0000-4000-8000-000000000011"
		const pay26StudentTwo = "40000000-0000-4000-8000-000000000012"
		if _, err := pool.Exec(ctx, `INSERT INTO users(id,email,display_name,role,password_hash) VALUES
($1,'pay26-one@example.test','Pay26 one','student','$argon2id$v=19$m=65536,t=3,p=2$c3R1ZGVudC1kZW1vLXNhbHQ$5S1MhXX3wgDMnIKuygFRfUNVvTtRktPECxV3ZUBg0is'),($2,'pay26-two@example.test','Pay26 two','student','$argon2id$v=19$m=65536,t=3,p=2$c3R1ZGVudC1kZW1vLXNhbHQ$5S1MhXX3wgDMnIKuygFRfUNVvTtRktPECxV3ZUBg0is')`, pay26StudentOne, pay26StudentTwo); err != nil {
			t.Fatal(err)
		}
		for _, stmt := range []struct {
			sql  string
			args []any
		}{
			{`INSERT INTO courses(id,slug,title,summary,description,status,instructor_id,published_at,price_minor,currency,created_at,updated_at) VALUES($1,'pay26','Pay26','Summary','Description','published',$2,$3,NULL,NULL,$3,$3)`, []any{pay26CourseID, integrationAdminID, now}},
			{`INSERT INTO course_modules(id,course_id,title,position) VALUES($1,$2,'Module',0)`, []any{pay26ModuleID, pay26CourseID}},
			{`INSERT INTO course_lessons(id,module_id,title,description,position) VALUES($1,$2,'Lesson','',0)`, []any{pay26LessonID, pay26ModuleID}},
			{`INSERT INTO lesson_progress(user_id,lesson_id,position_seconds,client_seq,updated_at) VALUES($1,$3,1,1,$4),($2,$3,2,1,$4)`, []any{pay26StudentOne, pay26StudentTwo, pay26LessonID, now}},
		} {
			if _, err := pool.Exec(ctx, stmt.sql, stmt.args...); err != nil {
				t.Fatal(err)
			}
		}
		pay26IDs := []string{"60000000-0000-4000-8000-000000000013"}
		pay26ID := func() string {
			id := pay26IDs[0]
			pay26IDs = pay26IDs[1:]
			return id
		}
		pay26Service := application.New(repository, repository, repository, func() time.Time { return now }, pay26ID, nil)
		updatedAt := now.Add(time.Microsecond)
		if err := pay26Service.GrandfatherOnPriceSet(ctx, pay26CourseID, now, 70000, "UAH", updatedAt); err != nil {
			t.Fatal(err)
		}
		var grants int
		if err := pool.QueryRow(ctx, `SELECT count(*) FROM course_grants WHERE course_id=$1 AND source='grandfather' AND state='active'`, pay26CourseID).Scan(&grants); err != nil {
			t.Fatal(err)
		}
		if grants != 2 {
			t.Fatalf("first price-set grandfather grants=%d, want 2", grants)
		}
		if err := pay26Service.GrandfatherOnPriceSet(ctx, pay26CourseID, now, 70000, "UAH", updatedAt); err != nil {
			t.Fatalf("repeat price-set: %v", err)
		}
		if err := pool.QueryRow(ctx, `SELECT count(*) FROM course_grants WHERE course_id=$1 AND source='grandfather' AND state='active'`, pay26CourseID).Scan(&grants); err != nil {
			t.Fatal(err)
		}
		if grants != 2 {
			t.Fatalf("repeat price-set grandfather grants=%d, want 2", grants)
		}
	})

	order, err := service.StartPurchase(ctx, application.Actor{UserID: integrationUserID, EmailVerified: true}, integrationCourseID, "purchase-1")
	if err != nil {
		t.Fatal(err)
	}
	repeated, err := service.StartPurchase(ctx, application.Actor{UserID: integrationUserID, EmailVerified: true}, integrationCourseID, "purchase-1")
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
	status, err := service.PurchaseStatus(ctx, application.Actor{UserID: integrationUserID, EmailVerified: true}, order.ID())
	if err != nil || status.Status() != domain.OrderPaid {
		t.Fatalf("status=%q error=%v", status.Status(), err)
	}
	var grants, events, audits int
	if err := pool.QueryRow(ctx, `SELECT (SELECT count(*) FROM course_grants WHERE order_id=$1),(SELECT count(*) FROM payment_events WHERE event_id='paid-1'),(SELECT count(*) FROM grant_audit WHERE payment_event_id=(SELECT id FROM payment_events WHERE event_id='paid-1'))`, order.ID()).Scan(&grants, &events, &audits); err != nil {
		t.Fatal(err)
	}
	if grants != 1 || events != 1 || audits != 1 {
		t.Fatalf("purchase grants=%d events=%d audits=%d", grants, events, audits)
	}
	// Pay04b keeps the refund trace honest: a repeat of the same paid event is a no-op before
	// DecidePaymentEvent only because EventID = orderReference:status. If that format ever changes,
	// a repeated callback for a paid order would reach the reject branch and raise a false refund flag.
	var refundRequired bool
	if err := pool.QueryRow(ctx, `SELECT refund_required FROM orders WHERE id=$1::text::uuid`, order.ID()).Scan(&refundRequired); err != nil {
		t.Fatal(err)
	}
	if refundRequired {
		t.Fatal("a repeated paid event raised refund_required on a correctly paid order")
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
	mismatchOrder, err := service.StartPurchase(ctx, application.Actor{UserID: integrationAdminID, EmailVerified: true}, integrationCourseID, "purchase-mismatch")
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
		failedService := application.New(repository, repository, repository, func() time.Time { return now }, failedID, nil)
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
		invalidHandler := ordershttp.NewHandler(failedService, invalidCallbackVerifier{}, "")
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

	t.Run("administrator cannot grant themselves or create a row", func(t *testing.T) {
		selfGrantIDs := subtestIDs("72")
		selfGrantService := application.New(repository, repository, repository, func() time.Time { return now }, selfGrantIDs, nil)
		var before, after int
		if err := pool.QueryRow(ctx, `SELECT count(*) FROM course_grants WHERE user_id=$1::text::uuid AND course_id=$2::text::uuid AND source='admin'`, integrationAdminID, integrationCourseID).Scan(&before); err != nil {
			t.Fatal(err)
		}
		if _, err := selfGrantService.GrantCourse(ctx, application.Actor{UserID: integrationAdminID, Administrator: true, EmailVerified: true}, integrationAdminID, integrationCourseID, "manual access"); err != domain.ErrGrantSelfForbidden {
			t.Fatalf("self grant error=%v", err)
		}
		if err := pool.QueryRow(ctx, `SELECT count(*) FROM course_grants WHERE user_id=$1::text::uuid AND course_id=$2::text::uuid AND source='admin'`, integrationAdminID, integrationCourseID).Scan(&after); err != nil {
			t.Fatal(err)
		}
		if after != before {
			t.Fatalf("self grant rows before=%d after=%d", before, after)
		}
	})

	t.Run("Lo18 an administrator recipient is refused and no grant row is written", func(t *testing.T) {
		var secondAdmin string
		if err := pool.QueryRow(ctx, `INSERT INTO users(id,email,display_name,role,password_hash) VALUES
(gen_random_uuid(),'lo18-second-admin@example.test','Lo18 second admin','admin','$argon2id$v=19$m=65536,t=3,p=2$c3R1ZGVudC1kZW1vLXNhbHQ$5S1MhXX3wgDMnIKuygFRfUNVvTtRktPECxV3ZUBg0is') RETURNING id::text`).Scan(&secondAdmin); err != nil {
			t.Fatal(err)
		}
		if isAdmin, err := repository.IsAdministrator(ctx, secondAdmin); err != nil || !isAdmin {
			t.Fatalf("IsAdministrator(second admin)=%v err=%v", isAdmin, err)
		}
		if nonAdmin, err := repository.IsAdministrator(ctx, integrationUserID); err != nil || nonAdmin {
			t.Fatalf("IsAdministrator(student)=%v err=%v", nonAdmin, err)
		}
		lo18Service := application.New(repository, repository, repository, func() time.Time { return now }, subtestIDs("73"), nil)
		if _, err := lo18Service.GrantCourse(ctx, application.Actor{UserID: integrationAdminID, Administrator: true, EmailVerified: true}, secondAdmin, integrationCourseID, "support decision"); err != domain.ErrAdminNotLearner {
			t.Fatalf("grant to admin error=%v want=%v", err, domain.ErrAdminNotLearner)
		}
		var rows int
		if err := pool.QueryRow(ctx, `SELECT count(*) FROM course_grants WHERE user_id=$1::text::uuid AND course_id=$2::text::uuid`, secondAdmin, integrationCourseID).Scan(&rows); err != nil {
			t.Fatal(err)
		}
		if rows != 0 {
			t.Fatalf("grant to admin wrote %d rows, want 0", rows)
		}
	})

	t.Run("paid event settles when active admin grant already exists", func(t *testing.T) {
		var userID string
		if err := pool.QueryRow(ctx, `INSERT INTO users(id,email,display_name,role,password_hash) VALUES
(gen_random_uuid(),'paid-with-admin-grant@example.test','Paid with admin grant','student','$argon2id$v=19$m=65536,t=3,p=2$c3R1ZGVudC1kZW1vLXNhbHQ$5S1MhXX3wgDMnIKuygFRfUNVvTtRktPECxV3ZUBg0is') RETURNING id::text`).Scan(&userID); err != nil {
			t.Fatal(err)
		}
		order, err := service.StartPurchase(ctx, application.Actor{UserID: userID, EmailVerified: true}, integrationCourseID, "purchase-with-admin-grant")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := service.GrantCourse(ctx, application.Actor{UserID: integrationAdminID, Administrator: true, EmailVerified: true}, userID, integrationCourseID, "issued while payment was pending"); err != nil {
			t.Fatal(err)
		}
		event := domain.PaymentEvent{Provider: "provider", EventID: "paid-with-admin-grant", OrderID: order.ID(), Type: domain.PaymentSucceeded, Amount: order.Amount(), ReceivedAt: now.Add(3 * time.Minute)}
		if err := service.ApplyPaymentEvent(ctx, event); err != nil {
			t.Fatalf("ApplyPaymentEvent with existing active grant: %v", err)
		}
		status, err := service.PurchaseStatus(ctx, application.Actor{UserID: userID, EmailVerified: true}, order.ID())
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

	t.Run("En02 paid event upgrades an active self membership in place", func(t *testing.T) {
		var userID string
		if err := pool.QueryRow(ctx, `INSERT INTO users(id,email,display_name,role,password_hash) VALUES(gen_random_uuid(),'self-upgrade@example.test','Self upgrade','student','$argon2id$v=19$m=65536,t=3,p=2$c3R1ZGVudC1kZW1vLXNhbHQ$5S1MhXX3wgDMnIKuygFRfUNVvTtRktPECxV3ZUBg0is') RETURNING id::text`).Scan(&userID); err != nil {
			t.Fatal(err)
		}
		var beforeID string
		if err := pool.QueryRow(ctx, `INSERT INTO course_grants(user_id,course_id,source,state,granted_at) VALUES($1::text::uuid,$2::text::uuid,'self','active',$3) RETURNING id::text`, userID, integrationCourseID, now).Scan(&beforeID); err != nil {
			t.Fatal(err)
		}
		ids := subtestIDs("81")
		svc := application.New(repository, repository, repository, func() time.Time { return now }, ids, nil)
		order, err := svc.StartPurchase(ctx, application.Actor{UserID: userID, EmailVerified: true}, integrationCourseID, "self-upgrade")
		if err != nil {
			t.Fatal(err)
		}
		if err := svc.ApplyPaymentEvent(ctx, domain.PaymentEvent{Provider: "provider", EventID: "self-upgrade-paid", OrderID: order.ID(), Type: domain.PaymentSucceeded, Amount: order.Amount(), ReceivedAt: now.Add(4 * time.Minute)}); err != nil {
			t.Fatal(err)
		}
		var afterID, source, state, orderID, action string
		var rows int
		if err := pool.QueryRow(ctx, `SELECT id::text,source,state,order_id::text,(SELECT action FROM grant_audit WHERE grant_id=course_grants.id ORDER BY created_at DESC LIMIT 1),(SELECT count(*) FROM course_grants WHERE user_id=$1::text::uuid AND course_id=$2::text::uuid) FROM course_grants WHERE user_id=$1::text::uuid AND course_id=$2::text::uuid`, userID, integrationCourseID).Scan(&afterID, &source, &state, &orderID, &action, &rows); err != nil {
			t.Fatal(err)
		}
		if afterID != beforeID || source != "purchase" || state != "active" || orderID != order.ID() || action != "upgrade" || rows != 1 {
			t.Fatalf("id=%q source=%q state=%q order=%q action=%q rows=%d", afterID, source, state, orderID, action, rows)
		}
	})

	t.Run("blocked revoked and purchase memberships require refund without mutation", func(t *testing.T) {
		for _, test := range []struct {
			name   string
			source string
			state  string
			prefix string
		}{
			{name: "blocked", source: "admin", state: "blocked", prefix: "82"},
			{name: "revoked", source: "admin", state: "revoked", prefix: "83"},
			{name: "purchase active", source: "purchase", state: "active", prefix: "84"},
		} {
			t.Run(test.name, func(t *testing.T) {
				var userID string
				if err := pool.QueryRow(ctx, `INSERT INTO users(id,email,display_name,role,password_hash) VALUES(gen_random_uuid(),$1,'Refund case','student','$argon2id$v=19$m=65536,t=3,p=2$c3R1ZGVudC1kZW1vLXNhbHQ$5S1MhXX3wgDMnIKuygFRfUNVvTtRktPECxV3ZUBg0is') RETURNING id::text`, "refund-"+test.state+"-"+test.source+"@example.test").Scan(&userID); err != nil {
					t.Fatal(err)
				}
				ids := subtestIDs(test.prefix)
				svc := application.New(repository, repository, repository, func() time.Time { return now }, ids, nil)
				order, err := svc.StartPurchase(ctx, application.Actor{UserID: userID, EmailVerified: true}, integrationCourseID, "refund-"+test.state+"-"+test.source)
				if err != nil {
					t.Fatal(err)
				}
				var grantID string
				if test.source == "purchase" {
					err = pool.QueryRow(ctx, `INSERT INTO course_grants(user_id,course_id,source,order_id,state,granted_at) VALUES($1::text::uuid,$2::text::uuid,'purchase',$3::text::uuid,'active',$4) RETURNING id::text`, userID, integrationCourseID, order.ID(), now).Scan(&grantID)
				} else {
					err = pool.QueryRow(ctx, `INSERT INTO course_grants(user_id,course_id,source,state,granted_at) VALUES($1::text::uuid,$2::text::uuid,$3,$4,$5) RETURNING id::text`, userID, integrationCourseID, test.source, test.state, now).Scan(&grantID)
				}
				if err != nil {
					t.Fatal(err)
				}
				if err := svc.ApplyPaymentEvent(ctx, domain.PaymentEvent{Provider: "provider", EventID: "refund-" + test.state + "-" + test.source, OrderID: order.ID(), Type: domain.PaymentSucceeded, Amount: order.Amount(), ReceivedAt: now.Add(5 * time.Minute)}); err != nil {
					t.Fatal(err)
				}
				var source, state, gotGrantID string
				var refund bool
				if err := pool.QueryRow(ctx, `SELECT source,state,id::text,(SELECT refund_required FROM orders WHERE id=$3::text::uuid) FROM course_grants WHERE user_id=$1::text::uuid AND course_id=$2::text::uuid`, userID, integrationCourseID, order.ID()).Scan(&source, &state, &gotGrantID, &refund); err != nil {
					t.Fatal(err)
				}
				if source != test.source || state != test.state || gotGrantID != grantID || !refund {
					t.Fatalf("source=%q state=%q id=%q refund=%v", source, state, gotGrantID, refund)
				}
			})
		}
	})

	t.Run("L2b enrollment membership writes", func(t *testing.T) {
		const enrolledUser = "40000000-0000-4000-8000-000000000090"
		const otherUser = "40000000-0000-4000-8000-000000000091"
		const draftCourse = "50000000-0000-4000-8000-000000000090"
		const archivedCourse = "50000000-0000-4000-8000-000000000091"
		const otherFreeCourse = "50000000-0000-4000-8000-000000000092"
		if _, err := pool.Exec(ctx, `INSERT INTO users(id,email,display_name,role,password_hash) VALUES
($1,'enrollment@example.test','Enrollment','student','$argon2id$v=19$m=65536,t=3,p=2$c3R1ZGVudC1kZW1vLXNhbHQ$5S1MhXX3wgDMnIKuygFRfUNVvTtRktPECxV3ZUBg0is'),
($2,'enrollment-other@example.test','Enrollment other','student','$argon2id$v=19$m=65536,t=3,p=2$c3R1ZGVudC1kZW1vLXNhbHQ$5S1MhXX3wgDMnIKuygFRfUNVvTtRktPECxV3ZUBg0is')`, enrolledUser, otherUser); err != nil {
			t.Fatal(err)
		}
		if _, err := pool.Exec(ctx, `INSERT INTO courses(id,slug,title,summary,description,status,instructor_id,published_at,created_at,updated_at) VALUES
($1,'enrollment-draft','Draft','Summary','Description','draft',$2,NULL,$3,$3),
($4,'enrollment-archived','Archived','Summary','Description','archived',$2,$3,$3,$3),
($5,'enrollment-other-free','Other free','Summary','Description','published',$6,$3,$3,$3)`, draftCourse, integrationAdminID, now, archivedCourse, otherFreeCourse, otherUser); err != nil {
			t.Fatal(err)
		}
		svc := application.New(repository, repository, repository, func() time.Time { return now }, subtestIDs("90"), nil)
		student := application.Actor{UserID: enrolledUser, EmailVerified: true}
		admin := application.Actor{UserID: integrationAdminID, Administrator: true, EmailVerified: true}
		owner := application.Actor{UserID: integrationAdminID, EmailVerified: true}
		if enrolled, err := svc.Enroll(ctx, application.Actor{UserID: otherUser, EmailVerified: true}, integrationFreeID); err != nil || enrolled.Source() != domain.GrantSelf {
			t.Fatalf("En14 instructor enrolls in another free course: grant=%#v error=%v", enrolled, err)
		}
		if enrolled, err := svc.Enroll(ctx, admin, otherFreeCourse); err != nil || enrolled.Source() != domain.GrantSelf {
			t.Fatalf("En14 admin enrolls in another free course: grant=%#v error=%v", enrolled, err)
		}

		grant, err := svc.Enroll(ctx, student, integrationFreeID)
		if err != nil || grant.State() != domain.GrantActive || grant.Source() != domain.GrantSelf {
			t.Fatalf("En01 enroll grant=%#v error=%v", grant, err)
		}
		if _, err := svc.Enroll(ctx, student, integrationFreeID); err != nil {
			t.Fatalf("En15 repeat enroll: %v", err)
		}
		var rows int
		var action string
		if err := pool.QueryRow(ctx, `SELECT (SELECT count(*) FROM course_grants WHERE user_id=$1::text::uuid AND course_id=$2::text::uuid),(SELECT action FROM grant_audit WHERE grant_id=$3::text::uuid ORDER BY created_at DESC LIMIT 1)`, enrolledUser, integrationFreeID, grant.ID()).Scan(&rows, &action); err != nil || rows != 1 || action != "enroll" {
			t.Fatalf("En01/En15 rows=%d action=%q error=%v", rows, action, err)
		}
		if _, err := svc.Block(ctx, owner, enrolledUser, integrationFreeID); err != nil {
			t.Fatalf("owner block: %v", err)
		}
		if _, err := svc.Block(ctx, application.Actor{UserID: otherUser}, enrolledUser, integrationFreeID); err != domain.ErrEnrollmentForbidden {
			t.Fatalf("stranger block error=%v", err)
		}
		if unblocked, err := svc.Unblock(ctx, admin, enrolledUser, integrationFreeID); err != nil || unblocked.State() != domain.GrantActive {
			t.Fatalf("admin unblock grant=%#v error=%v", unblocked, err)
		}
		if revoked, err := svc.Revoke(ctx, admin, enrolledUser, integrationFreeID); err != nil || revoked.State() != domain.GrantRevoked {
			t.Fatalf("admin revoke grant=%#v error=%v", revoked, err)
		}
		if _, err := svc.Revoke(ctx, admin, enrolledUser, integrationFreeID); err != nil {
			t.Fatalf("re-revoke should be no-op: %v", err)
		}
		if restored, err := svc.Enroll(ctx, student, integrationFreeID); err != nil || restored.State() != domain.GrantActive || restored.Source() != domain.GrantSelf {
			t.Fatalf("free re-enroll grant=%#v error=%v", restored, err)
		}
		if _, err := svc.Enroll(ctx, student, integrationCourseID); err != domain.ErrCourseNotEnrollable {
			t.Fatalf("paid enroll error=%v", err)
		}
		for _, blockedCourse := range []string{draftCourse, archivedCourse} {
			if _, err := svc.Enroll(ctx, student, blockedCourse); err != domain.ErrCourseNotEnrollable {
				t.Fatalf("course %s enroll error=%v", blockedCourse, err)
			}
		}
		if _, err := svc.GrantCourse(ctx, admin, enrolledUser, integrationFreeID, "manual restore"); err != domain.ErrCourseAlreadyOwned {
			t.Fatalf("active admin grant error=%v", err)
		}
		if _, err := svc.Revoke(ctx, admin, enrolledUser, integrationFreeID); err != nil {
			t.Fatal(err)
		}
		if restored, err := svc.GrantCourse(ctx, admin, enrolledUser, integrationFreeID, "manual restore"); err != nil || restored.State() != domain.GrantActive || restored.Source() != domain.GrantAdmin {
			t.Fatalf("admin re-grant grant=%#v error=%v", restored, err)
		}
		if err := pool.QueryRow(ctx, `SELECT count(*) FROM course_grants WHERE user_id=$1::text::uuid AND course_id=$2::text::uuid`, enrolledUser, integrationFreeID).Scan(&rows); err != nil || rows != 1 {
			t.Fatalf("single membership rows=%d error=%v", rows, err)
		}
	})

	t.Run("L2b compare-and-set rejects a stale concurrent transition", func(t *testing.T) {
		const casUser = "40000000-0000-4000-8000-000000000095"
		if _, err := pool.Exec(ctx, `INSERT INTO users(id,email,display_name,role,password_hash) VALUES($1,'cas@example.test','CAS','student','$argon2id$v=19$m=65536,t=3,p=2$c3R1ZGVudC1kZW1vLXNhbHQ$5S1MhXX3wgDMnIKuygFRfUNVvTtRktPECxV3ZUBg0is')`, casUser); err != nil {
			t.Fatal(err)
		}
		if _, err := pool.Exec(ctx, `INSERT INTO course_grants(id,user_id,course_id,source,state,granted_at) VALUES(gen_random_uuid(),$1::text::uuid,$2::text::uuid,'self','active',$3)`, casUser, integrationFreeID, now); err != nil {
			t.Fatal(err)
		}
		// Both transitions read the same active snapshot (the lost-update window): one commits,
		// the other must be rejected by the state predicate rather than blindly overwrite.
		snapshot, err := repository.Membership(ctx, casUser, integrationFreeID)
		if err != nil || snapshot == nil || snapshot.State() != domain.GrantActive {
			t.Fatalf("snapshot=%#v error=%v", snapshot, err)
		}
		admin := integrationAdminID
		blocked, err := snapshot.Block()
		if err != nil {
			t.Fatal(err)
		}
		if _, err := repository.UpdateMembership(ctx, blocked, domain.GrantActive, "block", "", &admin); err != nil {
			t.Fatalf("winning block: %v", err)
		}
		revoked, err := snapshot.Revoke()
		if err != nil {
			t.Fatal(err)
		}
		if _, err := repository.UpdateMembership(ctx, revoked, domain.GrantActive, "revoke", "", &admin); !errors.Is(err, domain.ErrMembershipConflict) {
			t.Fatalf("stale revoke error=%v want=%v", err, domain.ErrMembershipConflict)
		}
		var state string
		var audits int
		if err := pool.QueryRow(ctx, `SELECT state,(SELECT count(*) FROM grant_audit WHERE grant_id=course_grants.id) FROM course_grants WHERE user_id=$1::text::uuid AND course_id=$2::text::uuid`, casUser, integrationFreeID).Scan(&state, &audits); err != nil {
			t.Fatal(err)
		}
		if state != "blocked" || audits != 1 {
			t.Fatalf("after stale write state=%q audits=%d (want blocked/1: no lost update, no spurious revoke audit)", state, audits)
		}
	})

	t.Run("En08 roster returns current members and their progress with keyset pagination", func(t *testing.T) {
		const (
			rosterCourse  = "50000000-0000-4000-8000-000000000101"
			zeroLessons   = "50000000-0000-4000-8000-000000000102"
			emptyCourse   = "50000000-0000-4000-8000-000000000103"
			moduleID      = "51000000-0000-4000-8000-000000000101"
			lessonOneID   = "52000000-0000-4000-8000-000000000101"
			lessonTwoID   = "52000000-0000-4000-8000-000000000102"
			lessonThreeID = "52000000-0000-4000-8000-000000000103"
			fullUser      = "40000000-0000-4000-8000-000000000101"
			blockedUser   = "40000000-0000-4000-8000-000000000102"
			noProgress    = "40000000-0000-4000-8000-000000000103"
			revokedUser   = "40000000-0000-4000-8000-000000000104"
		)
		if _, err := pool.Exec(ctx, `WITH inserted_users AS (
INSERT INTO users(id,email,display_name,role,password_hash) VALUES
($1,'roster-full@example.test','Full progress','student','$argon2id$v=19$m=65536,t=3,p=2$c3R1ZGVudC1kZW1vLXNhbHQ$5S1MhXX3wgDMnIKuygFRfUNVvTtRktPECxV3ZUBg0is'),
($2,'roster-blocked@example.test','Blocked progress','student','$argon2id$v=19$m=65536,t=3,p=2$c3R1ZGVudC1kZW1vLXNhbHQ$5S1MhXX3wgDMnIKuygFRfUNVvTtRktPECxV3ZUBg0is'),
($3,'roster-none@example.test','No progress','student','$argon2id$v=19$m=65536,t=3,p=2$c3R1ZGVudC1kZW1vLXNhbHQ$5S1MhXX3wgDMnIKuygFRfUNVvTtRktPECxV3ZUBg0is'),
($4,'roster-revoked@example.test','Revoked','student','$argon2id$v=19$m=65536,t=3,p=2$c3R1ZGVudC1kZW1vLXNhbHQ$5S1MhXX3wgDMnIKuygFRfUNVvTtRktPECxV3ZUBg0is')
), inserted_courses AS (
INSERT INTO courses(id,slug,title,summary,description,status,instructor_id,published_at,created_at,updated_at) VALUES
($5,'roster','Roster','Summary','Description','published',$6,$7,$7,$7),
($8,'roster-zero-lessons','Roster zero lessons','Summary','Description','published',$6,$7,$7,$7),
($12,'roster-empty','Roster empty','Summary','Description','published',$6,$7,$7,$7)
), inserted_modules AS (
INSERT INTO course_modules(id,course_id,title,position) VALUES($9,$5,'Roster module',0)
), inserted_lessons AS (
INSERT INTO course_lessons(id,module_id,title,description,position) VALUES($10,$9,'One','',0),($11,$9,'Two','',1),($13,$9,'Three','',2)
), inserted_grants AS (
INSERT INTO course_grants(user_id,course_id,source,state,granted_at) VALUES
($1,$5,'self','active',$7 + interval '3 minutes'),
($2,$5,'self','blocked',$7 + interval '2 minutes'),
($3,$5,'self','active',$7 + interval '1 minute'),
($4,$5,'self','revoked',$7 + interval '4 minutes'),
($1,$8,'self','active',$7)
)
INSERT INTO lesson_progress(user_id,lesson_id,position_seconds,client_seq,updated_at,completed_at) VALUES
($1,$10,1,1,$7,$7),($1,$11,1,1,$7,$7),($1,$13,1,1,$7,$7),($2,$10,1,1,$7,$7),($2,$11,1,1,$7,$7)`, fullUser, blockedUser, noProgress, revokedUser, rosterCourse, integrationAdminID, now, zeroLessons, moduleID, lessonOneID, lessonTwoID, emptyCourse, lessonThreeID); err != nil {
			t.Fatal(err)
		}
		owner := application.Actor{UserID: integrationAdminID}
		first, err := service.Roster(ctx, owner, rosterCourse, "", 1)
		if err != nil || len(first.Entries) != 1 || first.Entries[0].UserID != fullUser || first.Entries[0].DisplayName != "Full progress" || first.Entries[0].State != domain.GrantActive || first.Entries[0].ProgressPercent != 100 || !first.Entries[0].EnrolledAt.Equal(now.Add(3*time.Minute)) || first.NextCursor == "" {
			t.Fatalf("first roster page=%#v error=%v", first, err)
		}
		second, err := service.Roster(ctx, owner, rosterCourse, first.NextCursor, 1)
		if err != nil || len(second.Entries) != 1 || second.Entries[0].UserID != blockedUser || second.Entries[0].State != domain.GrantBlocked || second.Entries[0].ProgressPercent != 67 || second.NextCursor == "" {
			t.Fatalf("second roster page=%#v error=%v", second, err)
		}
		last, err := service.Roster(ctx, application.Actor{UserID: "admin", Administrator: true}, rosterCourse, second.NextCursor, 10)
		if err != nil || len(last.Entries) != 1 || last.Entries[0].UserID != noProgress || last.Entries[0].ProgressPercent != 0 || last.NextCursor != "" {
			t.Fatalf("last roster page=%#v error=%v", last, err)
		}
		if _, err := service.Roster(ctx, owner, rosterCourse, "not-a-cursor", 1); !errors.Is(err, domain.ErrInvalidCursor) {
			t.Fatalf("invalid cursor error=%v", err)
		}
		if _, err := service.Roster(ctx, application.Actor{UserID: noProgress}, rosterCourse, "", 1); !errors.Is(err, domain.ErrEnrollmentForbidden) {
			t.Fatalf("stranger roster error=%v", err)
		}
		zero, err := service.Roster(ctx, owner, zeroLessons, "", 1)
		if err != nil || len(zero.Entries) != 1 || zero.Entries[0].ProgressPercent != 0 {
			t.Fatalf("zero-lesson roster page=%#v error=%v", zero, err)
		}
		empty, err := service.Roster(ctx, owner, emptyCourse, "", 1)
		if err != nil || len(empty.Entries) != 0 || empty.NextCursor != "" {
			t.Fatalf("empty roster page=%#v error=%v", empty, err)
		}
		// A since-deleted but still-enrolled member stays on the roster with an anonymized (empty)
		// name, mirroring how the review list drops a deleted author's display_name.
		if _, err := pool.Exec(ctx, `UPDATE users SET disabled_at=updated_at, deleted_at=updated_at WHERE id=$1::text::uuid`, noProgress); err != nil {
			t.Fatal(err)
		}
		anonymized, err := service.Roster(ctx, owner, rosterCourse, "", 10)
		if err != nil {
			t.Fatalf("roster after soft-delete error=%v", err)
		}
		var sawDeleted bool
		for _, e := range anonymized.Entries {
			if e.UserID == noProgress {
				sawDeleted = true
				if e.DisplayName != "" {
					t.Fatalf("deleted member leaked display_name=%q", e.DisplayName)
				}
			}
		}
		if !sawDeleted {
			t.Fatal("deleted-but-enrolled member dropped from roster; expected anonymized row")
		}
	})

	t.Run("En01 my memberships returns own active and blocked course cards", func(t *testing.T) {
		const (
			memberUser    = "40000000-0000-4000-8000-000000000151"
			otherUser     = "40000000-0000-4000-8000-000000000152"
			activeCourse  = "50000000-0000-4000-8000-000000000151"
			blockedCourse = "50000000-0000-4000-8000-000000000152"
			revokedCourse = "50000000-0000-4000-8000-000000000153"
			moduleID      = "51000000-0000-4000-8000-000000000151"
			lessonOneID   = "52000000-0000-4000-8000-000000000151"
			lessonTwoID   = "52000000-0000-4000-8000-000000000152"
		)
		if _, err := pool.Exec(ctx, `WITH inserted_users AS (
INSERT INTO users(id,email,display_name,role,password_hash) VALUES
($1,'my-memberships@example.test','Member','student','$argon2id$v=19$m=65536,t=3,p=2$c3R1ZGVudC1kZW1vLXNhbHQ$5S1MhXX3wgDMnIKuygFRfUNVvTtRktPECxV3ZUBg0is'),
($2,'other-memberships@example.test','Other','student','$argon2id$v=19$m=65536,t=3,p=2$c3R1ZGVudC1kZW1vLXNhbHQ$5S1MhXX3wgDMnIKuygFRfUNVvTtRktPECxV3ZUBg0is')
), inserted_courses AS (
INSERT INTO courses(id,slug,title,summary,description,status,instructor_id,published_at,created_at,updated_at) VALUES
($3,'my-active','My active','Summary','Description','archived',$8,$9,$9,$9),
($4,'my-blocked','My blocked','Summary','Description','draft',$8,NULL,$9,$9),
($5,'my-revoked','My revoked','Summary','Description','published',$8,$9,$9,$9)
), inserted_modules AS (
INSERT INTO course_modules(id,course_id,title,position) VALUES($6,$3,'Module',0)
), inserted_lessons AS (
INSERT INTO course_lessons(id,module_id,title,description,position) VALUES($7,$6,'One','',0),($10,$6,'Two','',1)
), inserted_grants AS (
INSERT INTO course_grants(user_id,course_id,source,state,granted_at) VALUES
($1,$3,'self','active',$9 + interval '2 minutes'),
($1,$4,'self','blocked',$9 + interval '3 minutes'),
($1,$5,'self','revoked',$9 + interval '4 minutes'),
($2,$3,'self','active',$9 + interval '5 minutes')
)
INSERT INTO lesson_progress(user_id,lesson_id,position_seconds,client_seq,updated_at,completed_at)
VALUES($1,$7,1,1,$9,$9)`, memberUser, otherUser, activeCourse, blockedCourse, revokedCourse, moduleID, lessonOneID, integrationAdminID, now, lessonTwoID); err != nil {
			t.Fatal(err)
		}
		cards, err := service.MyMemberships(ctx, application.Actor{UserID: memberUser})
		if err != nil {
			t.Fatal(err)
		}
		if len(cards) != 2 {
			t.Fatalf("cards=%#v", cards)
		}
		if cards[0].CourseID != blockedCourse || cards[0].Slug != "my-blocked" || cards[0].Title != "My blocked" || cards[0].State != domain.GrantBlocked || cards[0].ProgressPercent != 0 || !cards[0].EnrolledAt.Equal(now.Add(3*time.Minute)) {
			t.Fatalf("first card=%#v", cards[0])
		}
		if cards[1].CourseID != activeCourse || cards[1].Slug != "my-active" || cards[1].Title != "My active" || cards[1].State != domain.GrantActive || cards[1].ProgressPercent != 50 || !cards[1].EnrolledAt.Equal(now.Add(2*time.Minute)) {
			t.Fatalf("second card=%#v", cards[1])
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

type integrationPaymentProvider struct{}

func (*integrationPaymentProvider) Checkout(context.Context, domain.Order, ports.Buyer, ports.CheckoutURLs) (ports.CheckoutForm, error) {
	return ports.CheckoutForm{Action: "https://secure.wayforpay.com/pay"}, nil
}

func (*integrationPaymentProvider) OrderStatus(context.Context, string) (ports.PaymentStatus, error) {
	return ports.PaymentStatus{}, nil
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

// subtestIDs gives a subtest its own unbounded id source: a fixed slice runs dry the moment the code
// under test calls newID once more (three subtests in this file have panicked that way). The prefix
// keeps each subtest's ids apart from every other fixture.
func subtestIDs(prefix string) func() string {
	next := 0
	return func() string {
		next++
		return fmt.Sprintf("%s000000-0000-4000-8000-%012d", prefix, next)
	}
}
