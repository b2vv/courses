package postgres

import (
	"context"
	"errors"
	"log/slog"
	"strconv"
	"time"

	"courses/internal/orders/domain"
	"courses/internal/orders/ports"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct{ pool *pgxpool.Pool }

func NewRepository(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

func (repository *Repository) CourseForPurchase(ctx context.Context, courseID string) (ports.CourseFacts, error) {
	var status string
	var minor pgtype.Int8
	var currency pgtype.Text
	var title string
	err := repository.pool.QueryRow(ctx, `SELECT status, title, price_minor, currency FROM courses WHERE id=$1::text::uuid`, courseID).Scan(&status, &title, &minor, &currency)
	if errors.Is(err, pgx.ErrNoRows) {
		return ports.CourseFacts{}, domain.ErrCourseNotPurchasable
	}
	if err != nil {
		return ports.CourseFacts{}, persistenceError(err)
	}
	facts := ports.CourseFacts{Published: status == "published", Title: title}
	if minor.Valid && currency.Valid {
		money, moneyErr := domain.NewMoney(strconv.FormatInt(minor.Int64, 10), currency.String)
		if moneyErr != nil {
			return ports.CourseFacts{}, domain.ErrDependencyUnavailable
		}
		facts.Price = &money
	} else if minor.Valid || currency.Valid {
		return ports.CourseFacts{}, domain.ErrDependencyUnavailable
	}
	return facts, nil
}

func (repository *Repository) FindByIdempotencyKey(ctx context.Context, userID, key string) (domain.Order, error) {
	return scanOrder(repository.pool.QueryRow(ctx, `SELECT id::text,user_id::text,course_id::text,amount_minor,currency,status,idempotency_key,created_at,paid_at FROM orders WHERE user_id=$1::text::uuid AND idempotency_key=$2`, userID, key))
}

func (repository *Repository) Create(ctx context.Context, order domain.Order) (domain.Order, error) {
	_, err := repository.pool.Exec(ctx, `INSERT INTO orders(id,user_id,course_id,amount_minor,currency,status,idempotency_key,created_at) VALUES($1::text::uuid,$2::text::uuid,$3::text::uuid,$4,$5,$6,$7,$8)`,
		order.ID(), order.UserID(), order.CourseID(), int64(order.Amount().MinorUnits()), string(order.Amount().Currency()), string(order.Status()), order.IdempotencyKey(), order.CreatedAt())
	if err == nil {
		return order, nil
	}
	if isUnique(err, "orders_user_id_idempotency_key_key") {
		return repository.FindByIdempotencyKey(ctx, order.UserID(), order.IdempotencyKey())
	}
	return domain.Order{}, persistenceError(err)
}

func (repository *Repository) Get(ctx context.Context, userID, orderID string) (domain.Order, error) {
	return scanOrder(repository.pool.QueryRow(ctx, `SELECT id::text,user_id::text,course_id::text,amount_minor,currency,status,idempotency_key,created_at,paid_at FROM orders WHERE id=$1::text::uuid AND user_id=$2::text::uuid`, orderID, userID))
}

func (repository *Repository) MarkCheckoutOpened(ctx context.Context, userID, orderID string) (domain.Order, error) {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return domain.Order{}, persistenceError(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	order, err := scanOrder(tx.QueryRow(ctx, `UPDATE orders
SET checkout_opened_at=now()
WHERE id=$1::text::uuid AND user_id=$2::text::uuid AND status='pending' AND checkout_opened_at IS NULL
RETURNING id::text,user_id::text,course_id::text,amount_minor,currency,status,idempotency_key,created_at,paid_at`, orderID, userID))
	if err == nil {
		if err := tx.Commit(ctx); err != nil {
			return domain.Order{}, persistenceError(err)
		}
		return order, nil
	}
	if !errors.Is(err, domain.ErrOrderNotFound) {
		return domain.Order{}, err
	}
	var status string
	var opened pgtype.Timestamptz
	if err := tx.QueryRow(ctx, `SELECT status,checkout_opened_at FROM orders WHERE id=$1::text::uuid AND user_id=$2::text::uuid`, orderID, userID).Scan(&status, &opened); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Order{}, domain.ErrOrderNotFound
		}
		return domain.Order{}, persistenceError(err)
	}
	if status == string(domain.OrderPending) && opened.Valid {
		return domain.Order{}, domain.ErrCheckoutAlreadyOpened
	}
	return domain.Order{}, domain.ErrOrderNotFound
}

func (repository *Repository) ActiveGrants(ctx context.Context, userID, courseID string) ([]domain.Grant, error) {
	rows, err := repository.pool.Query(ctx, `SELECT id::text,user_id::text,course_id::text,source,order_id::text,state,granted_at,granted_by::text FROM course_grants WHERE user_id=$1::text::uuid AND course_id=$2::text::uuid AND state='active'`, userID, courseID)
	if err != nil {
		return nil, persistenceError(err)
	}
	defer rows.Close()
	var grants []domain.Grant
	for rows.Next() {
		var id, grantUserID, grantCourseID, source, state string
		var orderID, grantedBy pgtype.Text
		var grantedAt time.Time
		if err := rows.Scan(&id, &grantUserID, &grantCourseID, &source, &orderID, &state, &grantedAt, &grantedBy); err != nil {
			return nil, persistenceError(err)
		}
		grant, err := domain.NewGrant(id, grantUserID, grantCourseID, domain.GrantSource(source), textPointer(orderID), domain.GrantState(state), grantedAt, textPointer(grantedBy))
		if err != nil {
			return nil, domain.ErrDependencyUnavailable
		}
		grants = append(grants, grant)
	}
	if err := rows.Err(); err != nil {
		return nil, persistenceError(err)
	}
	return grants, nil
}

func (repository *Repository) GrantWithAudit(ctx context.Context, command ports.GrantCommand) (domain.Grant, error) {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return domain.Grant{}, persistenceError(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	grant := command.Grant
	_, err = tx.Exec(ctx, `INSERT INTO course_grants(id,user_id,course_id,source,order_id,state,granted_at,granted_by) VALUES($1::text::uuid,$2::text::uuid,$3::text::uuid,$4,NULL,$5,$6,$7::text::uuid)`,
		grant.ID(), grant.UserID(), grant.CourseID(), string(grant.Source()), string(grant.State()), grant.GrantedAt(), grant.GrantedBy())
	if isUnique(err, "course_grants_one_active_per_user_course") {
		return domain.Grant{}, domain.ErrCourseAlreadyOwned
	}
	if err != nil {
		return domain.Grant{}, persistenceError(err)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO grant_audit(grant_id,actor_id,action,reason,created_at) VALUES($1::text::uuid,$2::text::uuid,'grant',$3,$4)`, grant.ID(), grant.GrantedBy(), command.Reason, grant.GrantedAt()); err != nil {
		return domain.Grant{}, persistenceError(err)
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.Grant{}, persistenceError(err)
	}
	return grant, nil
}

func (repository *Repository) GrandfatherOnPriceSet(ctx context.Context, courseID string, expectedUpdatedAt time.Time, price domain.Money, updatedAt time.Time) error {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return persistenceError(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	result, err := tx.Exec(ctx, `UPDATE courses SET price_minor=$3,currency=$4,updated_at=$5 WHERE id=$1::text::uuid AND updated_at=$2 AND price_minor IS NULL`,
		courseID, expectedUpdatedAt, int64(price.MinorUnits()), string(price.Currency()), updatedAt)
	if err != nil {
		return persistenceError(err)
	}
	if result.RowsAffected() != 1 {
		var currentMinor pgtype.Int8
		var currentCurrency pgtype.Text
		if err := tx.QueryRow(ctx, `SELECT price_minor,currency FROM courses WHERE id=$1::text::uuid`, courseID).Scan(&currentMinor, &currentCurrency); err != nil {
			return persistenceError(err)
		}
		if !currentMinor.Valid || !currentCurrency.Valid || currentMinor.Int64 != int64(price.MinorUnits()) || currentCurrency.String != string(price.Currency()) {
			return domain.ErrCourseModified
		}
		// Repeat of the same price-set is a no-op: the first call set the price AND grandfathered,
		// so there is nothing left to grant. Production never reaches here (authoring grandfathers
		// only when the stored price is NULL); this keeps a direct repeat idempotent.
		return persistenceError(tx.Commit(ctx))
	}
	_, err = tx.Exec(ctx, `WITH inserted AS (
  INSERT INTO course_grants(user_id,course_id,source,state,granted_at)
  SELECT DISTINCT progress.user_id, modules.course_id, 'grandfather', 'active', $2::timestamptz
  FROM lesson_progress AS progress
  JOIN course_lessons AS lessons ON lessons.id=progress.lesson_id
  JOIN course_modules AS modules ON modules.id=lessons.module_id
  WHERE modules.course_id=$1::text::uuid
  ON CONFLICT (user_id,course_id) WHERE state='active' DO NOTHING
  RETURNING id, granted_at
)
INSERT INTO grant_audit(grant_id,action,reason,created_at)
SELECT id,'grant','free-to-paid grandfather',granted_at FROM inserted`, courseID, updatedAt)
	if err != nil {
		return persistenceError(err)
	}
	return persistenceError(tx.Commit(ctx))
}

func (repository *Repository) Apply(ctx context.Context, event domain.PaymentEvent, newID func() string) error {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return persistenceError(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var duplicate bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM payment_events WHERE provider=$1 AND event_id=$2)`, event.Provider, event.EventID).Scan(&duplicate); err != nil {
		return persistenceError(err)
	}
	if duplicate {
		return persistenceError(tx.Commit(ctx))
	}
	order, err := scanOrder(tx.QueryRow(ctx, `SELECT id::text,user_id::text,course_id::text,amount_minor,currency,status,idempotency_key,created_at,paid_at FROM orders WHERE id=$1::text::uuid FOR UPDATE`, event.OrderID))
	if errors.Is(err, domain.ErrOrderNotFound) {
		return domain.ErrPaymentEventUnmatched
	}
	if err != nil {
		return err
	}
	eventDatabaseID := newID()
	err = tx.QueryRow(ctx, `INSERT INTO payment_events(id,provider,event_id,order_id,type,amount_minor,received_at) VALUES($1::text::uuid,$2,$3,$4::text::uuid,$5,$6,$7)
ON CONFLICT (provider,event_id) DO NOTHING RETURNING id::text`, eventDatabaseID, event.Provider, event.EventID, event.OrderID, string(event.Type), int64(event.Amount.MinorUnits()), event.ReceivedAt).Scan(&eventDatabaseID)
	if errors.Is(err, pgx.ErrNoRows) {
		return persistenceError(tx.Commit(ctx))
	}
	if err != nil {
		return persistenceError(err)
	}
	decision, decisionErr := domain.DecidePaymentEvent(event, &order, false)
	if decision.Action == domain.PaymentEventReject {
		if event.Type == domain.PaymentSucceeded {
			if _, err := tx.Exec(ctx, `UPDATE orders SET refund_required=true WHERE id=$1::text::uuid`, order.ID()); err != nil {
				return persistenceError(err)
			}
			slog.Warn("payment event rejected", "order_id", order.ID(), "course_id", order.CourseID(), "user_id", order.UserID(), "reason", "rejected paid event")
		}
		if err := tx.Commit(ctx); err != nil {
			return persistenceError(err)
		}
		return decisionErr
	}
	if decisionErr != nil || decision.Action != domain.PaymentEventApply ||
		(decision.Order.Status() != domain.OrderFailed && !decision.IssueGrant) {
		return domain.ErrPaymentEventRejected
	}
	if decision.Order.Status() == domain.OrderFailed {
		if _, err := tx.Exec(ctx, `UPDATE orders SET status='failed',paid_at=NULL,provider=$2,provider_ref=$3 WHERE id=$1::text::uuid`, order.ID(), event.Provider, event.EventID); err != nil {
			return persistenceError(err)
		}
		return persistenceError(tx.Commit(ctx))
	}
	paidAt := decision.Order.PaidAt()
	if _, err := tx.Exec(ctx, `UPDATE orders SET status='paid',paid_at=$2,provider=$3,provider_ref=$4 WHERE id=$1::text::uuid`, order.ID(), paidAt, event.Provider, event.EventID); err != nil {
		return persistenceError(err)
	}
	grantID := newID()
	var insertedGrantID string
	err = tx.QueryRow(ctx, `INSERT INTO course_grants(id,user_id,course_id,source,order_id,state,granted_at) VALUES($1::text::uuid,$2::text::uuid,$3::text::uuid,'purchase',$4::text::uuid,'active',$5)
ON CONFLICT (user_id, course_id) WHERE state='active' DO NOTHING
RETURNING id::text`, grantID, decision.GrantUserID, decision.GrantCourseID, decision.GrantOrderID, event.ReceivedAt).Scan(&insertedGrantID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return persistenceError(err)
	}
	if err == nil {
		if _, err := tx.Exec(ctx, `INSERT INTO grant_audit(grant_id,action,payment_event_id,created_at) VALUES($1::text::uuid,'grant',$2::text::uuid,$3)`, insertedGrantID, eventDatabaseID, event.ReceivedAt); err != nil {
			return persistenceError(err)
		}
	} else {
		if _, updateErr := tx.Exec(ctx, `UPDATE orders SET refund_required=true WHERE id=$1::text::uuid`, order.ID()); updateErr != nil {
			return persistenceError(updateErr)
		}
		slog.Warn("paid order requires refund", "order_id", order.ID(), "course_id", order.CourseID(), "user_id", order.UserID())
	}
	return persistenceError(tx.Commit(ctx))
}

func scanOrder(row interface{ Scan(...any) error }) (domain.Order, error) {
	var id, userID, courseID, currency, status, key string
	var minor int64
	var createdAt time.Time
	var paidAt pgtype.Timestamptz
	if err := row.Scan(&id, &userID, &courseID, &minor, &currency, &status, &key, &createdAt, &paidAt); errors.Is(err, pgx.ErrNoRows) {
		return domain.Order{}, domain.ErrOrderNotFound
	} else if err != nil {
		return domain.Order{}, persistenceError(err)
	}
	money, err := domain.NewMoney(strconv.FormatInt(minor, 10), currency)
	if err != nil {
		return domain.Order{}, domain.ErrDependencyUnavailable
	}
	var paidAtPointer *time.Time
	if paidAt.Valid {
		paidAtPointer = &paidAt.Time
	}
	order, err := domain.RestoreOrder(id, userID, courseID, money, domain.OrderStatus(status), key, createdAt, paidAtPointer)
	if err != nil {
		return domain.Order{}, domain.ErrDependencyUnavailable
	}
	return order, nil
}

func persistenceError(err error) error {
	if err == nil || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	return domain.ErrDependencyUnavailable
}

func isUnique(err error, constraint string) bool {
	var postgresError *pgconn.PgError
	return errors.As(err, &postgresError) && postgresError.Code == "23505" && postgresError.ConstraintName == constraint
}

func textPointer(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	return &value.String
}
