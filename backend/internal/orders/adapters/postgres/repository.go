package postgres

import (
	"context"
	"encoding/base64"
	"errors"
	"log/slog"
	"strconv"
	"strings"
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

func (repository *Repository) CourseForEnrollment(ctx context.Context, courseID string) (ports.EnrollmentFacts, error) {
	var facts ports.EnrollmentFacts
	err := repository.pool.QueryRow(ctx, `SELECT status, (price_minor IS NULL) AS free, instructor_id::text FROM courses WHERE id=$1::text::uuid`, courseID).Scan(&facts.Status, &facts.Free, &facts.InstructorID)
	if errors.Is(err, pgx.ErrNoRows) {
		return ports.EnrollmentFacts{}, domain.ErrCourseNotEnrollable
	}
	if err != nil {
		return ports.EnrollmentFacts{}, persistenceError(err)
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

func (repository *Repository) Membership(ctx context.Context, userID, courseID string) (*domain.Grant, error) {
	var id, grantUserID, grantCourseID, source, state string
	var orderID, grantedBy pgtype.Text
	var grantedAt time.Time
	err := repository.pool.QueryRow(ctx, `SELECT id::text,user_id::text,course_id::text,source,order_id::text,state,granted_at,granted_by::text FROM course_grants WHERE user_id=$1::text::uuid AND course_id=$2::text::uuid LIMIT 1`, userID, courseID).Scan(&id, &grantUserID, &grantCourseID, &source, &orderID, &state, &grantedAt, &grantedBy)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, persistenceError(err)
	}
	grant, err := domain.NewGrant(id, grantUserID, grantCourseID, domain.GrantSource(source), textPointer(orderID), domain.GrantState(state), grantedAt, textPointer(grantedBy))
	if err != nil {
		return nil, domain.ErrDependencyUnavailable
	}
	return &grant, nil
}

// IsAdministrator reports whether userID names a live administrator. The grants adapter already
// reads the users table (Roster joins display_name); this reads the same table's role so the
// service can refuse making an administrator a learner (G8). A missing user is not an admin.
func (repository *Repository) IsAdministrator(ctx context.Context, userID string) (bool, error) {
	var role string
	err := repository.pool.QueryRow(ctx, `SELECT role FROM users WHERE id=$1::text::uuid AND deleted_at IS NULL LIMIT 1`, userID).Scan(&role)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, persistenceError(err)
	}
	return role == "admin", nil
}

// AuthorOf resolves a comment or review to (author, course). A comment reaches its course through
// its lesson's module; a review carries course_id directly. Reading the discussion tables here is a
// SQL read, not a module import (ADR 013). An unknown id yields the kind's not-found error.
func (repository *Repository) AuthorOf(ctx context.Context, kind, id string) (string, string, error) {
	var query string
	notFound := domain.ErrCommentNotFound
	switch kind {
	case "comment":
		query = `SELECT c.author_id::text, m.course_id::text FROM discussion_comments c JOIN course_lessons l ON l.id = c.lesson_id JOIN course_modules m ON m.id = l.module_id WHERE c.id = $1::text::uuid AND c.removed_at IS NULL`
	case "review":
		notFound = domain.ErrReviewNotFound
		query = `SELECT author_id::text, course_id::text FROM discussion_reviews WHERE id = $1::text::uuid AND removed_at IS NULL`
	default:
		return "", "", domain.ErrCommentNotFound
	}
	var userID, courseID string
	err := repository.pool.QueryRow(ctx, query, id).Scan(&userID, &courseID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", "", notFound
	}
	if err != nil {
		return "", "", persistenceError(err)
	}
	return userID, courseID, nil
}

func (repository *Repository) MyMemberships(ctx context.Context, userID string) ([]ports.MembershipCard, error) {
	rows, err := repository.pool.Query(ctx, `SELECT g.course_id::text, c.slug, c.title, g.state, g.granted_at,
  `+progressExpr+` AS progress_percent
FROM course_grants AS g
JOIN courses AS c ON c.id = g.course_id`+progressJoins+`
WHERE g.user_id = $1::text::uuid AND g.state IN ('active', 'blocked')
ORDER BY g.granted_at DESC
LIMIT 200`, userID)
	if err != nil {
		return nil, persistenceError(err)
	}
	defer rows.Close()
	cards := make([]ports.MembershipCard, 0)
	for rows.Next() {
		card, scanErr := scanMembershipCard(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		cards = append(cards, card)
	}
	if err := rows.Err(); err != nil {
		return nil, persistenceError(err)
	}
	return cards, nil
}

func (repository *Repository) Roster(ctx context.Context, courseID, cursor string, limit int) (ports.RosterPage, error) {
	stamp, userID, err := decodeRosterCursor(cursor)
	if err != nil {
		return ports.RosterPage{}, err
	}
	rows, err := repository.pool.Query(ctx, `SELECT g.user_id::text, u.display_name, g.state, g.granted_at,
  `+progressExpr+` AS progress_percent
FROM course_grants AS g
LEFT JOIN users AS u ON u.id = g.user_id AND u.deleted_at IS NULL`+progressJoins+`
WHERE g.course_id = $1::text::uuid AND g.state IN ('active', 'blocked')
  AND ($2::timestamptz IS NULL OR g.granted_at < $2 OR (g.granted_at = $2 AND g.user_id > $3::text::uuid))
ORDER BY g.granted_at DESC, g.user_id ASC
LIMIT $4`, courseID, stamp, userID, limit+1)
	if err != nil {
		return ports.RosterPage{}, persistenceError(err)
	}
	defer rows.Close()
	entries := make([]ports.RosterEntry, 0, limit+1)
	for rows.Next() {
		entry, scanErr := scanRosterEntry(rows)
		if scanErr != nil {
			return ports.RosterPage{}, scanErr
		}
		entries = append(entries, entry)
	}
	if err := rows.Err(); err != nil {
		return ports.RosterPage{}, persistenceError(err)
	}
	page := ports.RosterPage{Entries: entries}
	if len(entries) > limit {
		last := entries[limit-1]
		page.NextCursor = encodeRosterCursor(last.EnrolledAt, last.UserID)
		page.Entries = entries[:limit]
	}
	return page, nil
}

// progressJoins and progressExpr are the single source of the per-member progress percent. Roster,
// MyMemberships and Learners all read it through these so Lo02 cannot drift from the roster: the
// two LATERALs count completed and total lessons of g.course_id for g.user_id, and the expression
// rounds completed/total to a percent. They assume the grant alias is g.
const progressJoins = `
CROSS JOIN LATERAL (
  SELECT count(*)::int AS completed
  FROM course_modules AS m
  JOIN course_lessons AS l ON l.module_id = m.id
  JOIN lesson_progress AS p ON p.lesson_id = l.id
  WHERE m.course_id = g.course_id AND p.user_id = g.user_id AND p.completed_at IS NOT NULL
) AS completed
CROSS JOIN LATERAL (
  SELECT count(*)::int AS total
  FROM course_modules AS m
  JOIN course_lessons AS l ON l.module_id = m.id
  WHERE m.course_id = g.course_id
) AS totals`

const progressExpr = `CASE WHEN totals.total = 0 THEN 0 ELSE (completed.completed * 100 + totals.total / 2) / totals.total END`

// Learners returns the learner overview: one row per user who has at least one membership visible to
// the viewer, each row carrying that user's visible courses (state, progress, enrolment) and last
// activity. The page is read in two steps — a keyset page of users, then their course detail — which
// reads more plainly than one aggregate query and keeps the shared progress fragment intact.
func (repository *Repository) Learners(ctx context.Context, filter ports.LearnerFilter, cursor string, limit int) (ports.LearnerPage, error) {
	cursorName, cursorUserID, err := decodeLearnerCursor(cursor)
	if err != nil {
		return ports.LearnerPage{}, err
	}
	admin := filter.Administrator
	viewer := nullableText(filter.ViewerID)
	courseID := nullableText(filter.CourseID)
	state := nullableText(filter.State)
	q := learnerQueryPattern(filter.Query)
	rows, err := repository.pool.Query(ctx, `SELECT user_id::text, display_name FROM (
  SELECT DISTINCT g.user_id AS user_id, coalesce(u.display_name, '') AS display_name
  FROM course_grants AS g
  JOIN courses AS c ON c.id = g.course_id
  LEFT JOIN users AS u ON u.id = g.user_id AND u.deleted_at IS NULL
  WHERE g.state IN ('active', 'blocked')
    AND ($1::boolean OR c.instructor_id = $2::text::uuid)
    AND ($3::text IS NULL OR g.course_id = $3::text::uuid)
    AND ($4::text IS NULL OR g.state = $4)
    AND ($5::text IS NULL OR u.display_name ILIKE $5)
) AS p
WHERE ($6::text IS NULL OR (lower(p.display_name), p.user_id) > ($6, $7::text::uuid))
ORDER BY lower(p.display_name) ASC, p.user_id ASC
LIMIT $8`, admin, viewer, courseID, state, q, cursorName, cursorUserID, limit+1)
	if err != nil {
		return ports.LearnerPage{}, persistenceError(err)
	}
	type person struct {
		userID      string
		displayName string
	}
	people := make([]person, 0, limit+1)
	for rows.Next() {
		var pers person
		if scanErr := rows.Scan(&pers.userID, &pers.displayName); scanErr != nil {
			rows.Close()
			return ports.LearnerPage{}, persistenceError(scanErr)
		}
		people = append(people, pers)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return ports.LearnerPage{}, persistenceError(err)
	}
	page := ports.LearnerPage{Entries: make([]ports.LearnerRow, 0, len(people))}
	hasMore := len(people) > limit
	if hasMore {
		people = people[:limit]
	}
	if len(people) == 0 {
		return page, nil
	}
	order := make(map[string]int, len(people))
	ids := make([]string, len(people))
	for index, pers := range people {
		order[pers.userID] = index
		ids[index] = pers.userID
		page.Entries = append(page.Entries, ports.LearnerRow{UserID: pers.userID, DisplayName: pers.displayName})
	}
	detail, err := repository.pool.Query(ctx, `SELECT g.user_id::text, g.course_id::text, c.slug, c.title, g.state, g.granted_at,
  `+progressExpr+` AS progress_percent, activity.last_activity
FROM course_grants AS g
JOIN courses AS c ON c.id = g.course_id`+progressJoins+`
LEFT JOIN LATERAL (
  SELECT max(p.updated_at) AS last_activity
  FROM course_modules AS m
  JOIN course_lessons AS l ON l.module_id = m.id
  JOIN lesson_progress AS p ON p.lesson_id = l.id
  WHERE m.course_id = g.course_id AND p.user_id = g.user_id
) AS activity ON true
WHERE g.user_id = ANY($1::uuid[]) AND g.state IN ('active', 'blocked')
  AND ($2::boolean OR c.instructor_id = $3::text::uuid)
  AND ($4::text IS NULL OR g.course_id = $4::text::uuid)
  AND ($5::text IS NULL OR g.state = $5)
ORDER BY g.user_id, lower(c.title) ASC, g.course_id ASC`, ids, admin, viewer, courseID, state)
	if err != nil {
		return ports.LearnerPage{}, persistenceError(err)
	}
	defer detail.Close()
	for detail.Next() {
		var userID string
		var course ports.LearnerCourse
		var state string
		var lastActivity pgtype.Timestamptz
		if scanErr := detail.Scan(&userID, &course.CourseID, &course.Slug, &course.Title, &state, &course.EnrolledAt, &course.ProgressPercent, &lastActivity); scanErr != nil {
			return ports.LearnerPage{}, persistenceError(scanErr)
		}
		switch domain.GrantState(state) {
		case domain.GrantActive, domain.GrantBlocked:
			course.State = domain.GrantState(state)
		default:
			return ports.LearnerPage{}, domain.ErrDependencyUnavailable
		}
		index, ok := order[userID]
		if !ok {
			continue
		}
		entry := &page.Entries[index]
		entry.Courses = append(entry.Courses, course)
		if lastActivity.Valid && (entry.LastActivityAt == nil || lastActivity.Time.After(*entry.LastActivityAt)) {
			stamp := lastActivity.Time
			entry.LastActivityAt = &stamp
		}
	}
	if err := detail.Err(); err != nil {
		return ports.LearnerPage{}, persistenceError(err)
	}
	if hasMore {
		last := page.Entries[len(page.Entries)-1]
		cursor := encodeLearnerCursor(last.DisplayName, last.UserID)
		page.NextCursor = &cursor
	}
	return page, nil
}

// learnerQueryPattern turns a name search into an ILIKE pattern with the wildcards escaped, or NULL
// (no filter) for an empty query, so a literal % or _ in the search matches itself.
func learnerQueryPattern(query string) pgtype.Text {
	if query == "" {
		return pgtype.Text{}
	}
	escaped := strings.NewReplacer("\\", "\\\\", "%", "\\%", "_", "\\_").Replace(query)
	return pgtype.Text{String: "%" + escaped + "%", Valid: true}
}

func nullableText(value string) pgtype.Text {
	if value == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: value, Valid: true}
}

func encodeLearnerCursor(displayName, userID string) string {
	// userID is a UUID and holds no '|', so splitting on the first '|' recovers both parts even when
	// the display name itself contains '|'.
	return base64.RawURLEncoding.EncodeToString([]byte(userID + "|" + displayName))
}

func decodeLearnerCursor(value string) (pgtype.Text, pgtype.Text, error) {
	if value == "" {
		return pgtype.Text{}, pgtype.Text{}, nil
	}
	decoded, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return pgtype.Text{}, pgtype.Text{}, domain.ErrInvalidCursor
	}
	parts := strings.SplitN(string(decoded), "|", 2)
	if len(parts) != 2 {
		return pgtype.Text{}, pgtype.Text{}, domain.ErrInvalidCursor
	}
	var id pgtype.UUID
	if err := id.Scan(parts[0]); err != nil {
		return pgtype.Text{}, pgtype.Text{}, domain.ErrInvalidCursor
	}
	// The keyset compares lower(display_name); the cursor stores the raw name, so lower it here.
	return pgtype.Text{String: strings.ToLower(parts[1]), Valid: true}, pgtype.Text{String: parts[0], Valid: true}, nil
}

type rosterRow interface {
	Scan(...any) error
}

func scanRosterEntry(row rosterRow) (ports.RosterEntry, error) {
	var entry ports.RosterEntry
	var state string
	// LEFT JOIN with deleted_at IS NULL: a since-deleted but still-enrolled member keeps the
	// roster row but surfaces a NULL name, which maps to "" so the UI can anonymize it the way
	// the review list does, rather than leaking the deleted account's display name.
	var displayName pgtype.Text
	if err := row.Scan(&entry.UserID, &displayName, &state, &entry.EnrolledAt, &entry.ProgressPercent); err != nil {
		return ports.RosterEntry{}, persistenceError(err)
	}
	entry.DisplayName = displayName.String
	switch domain.GrantState(state) {
	case domain.GrantActive, domain.GrantBlocked:
		entry.State = domain.GrantState(state)
		return entry, nil
	default:
		return ports.RosterEntry{}, domain.ErrDependencyUnavailable
	}
}

func scanMembershipCard(row rosterRow) (ports.MembershipCard, error) {
	var card ports.MembershipCard
	var state string
	if err := row.Scan(&card.CourseID, &card.Slug, &card.Title, &state, &card.EnrolledAt, &card.ProgressPercent); err != nil {
		return ports.MembershipCard{}, persistenceError(err)
	}
	switch domain.GrantState(state) {
	case domain.GrantActive, domain.GrantBlocked:
		card.State = domain.GrantState(state)
		return card, nil
	default:
		return ports.MembershipCard{}, domain.ErrDependencyUnavailable
	}
}

func encodeRosterCursor(stamp time.Time, userID string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(stamp.UTC().Format(time.RFC3339Nano) + "|" + userID))
}

func decodeRosterCursor(value string) (*time.Time, *string, error) {
	if value == "" {
		return nil, nil, nil
	}
	decoded, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return nil, nil, domain.ErrInvalidCursor
	}
	parts := strings.Split(string(decoded), "|")
	if len(parts) != 2 || parts[1] == "" {
		return nil, nil, domain.ErrInvalidCursor
	}
	stamp, err := time.Parse(time.RFC3339Nano, parts[0])
	if err != nil {
		return nil, nil, domain.ErrInvalidCursor
	}
	var id pgtype.UUID
	if err := id.Scan(parts[1]); err != nil {
		return nil, nil, domain.ErrInvalidCursor
	}
	return &stamp, &parts[1], nil
}

func (repository *Repository) CreateMembership(ctx context.Context, grant domain.Grant, action, reason string, actorID *string) (domain.Grant, error) {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return domain.Grant{}, persistenceError(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	_, err = tx.Exec(ctx, `INSERT INTO course_grants(id,user_id,course_id,source,order_id,state,granted_at,granted_by) VALUES($1::text::uuid,$2::text::uuid,$3::text::uuid,$4,$5::text::uuid,$6,$7,$8::text::uuid)`,
		grant.ID(), grant.UserID(), grant.CourseID(), string(grant.Source()), grant.OrderID(), string(grant.State()), grant.GrantedAt(), grant.GrantedBy())
	if isUnique(err, "course_grants_one_per_user_course") {
		return domain.Grant{}, domain.ErrCourseAlreadyOwned
	}
	if err != nil {
		return domain.Grant{}, persistenceError(err)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO grant_audit(grant_id,actor_id,action,reason,created_at) VALUES($1::text::uuid,$2::text::uuid,$3,$4,$5)`, grant.ID(), actorID, action, reason, grant.GrantedAt()); err != nil {
		return domain.Grant{}, persistenceError(err)
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.Grant{}, persistenceError(err)
	}
	return grant, nil
}

func (repository *Repository) UpdateMembership(ctx context.Context, grant domain.Grant, expected domain.GrantState, action, reason string, actorID *string) (domain.Grant, error) {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return domain.Grant{}, persistenceError(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	// Compare-and-set on the prior state: the (user,course) row was read in a separate
	// transaction, so a concurrent transition may have moved it since. The state predicate
	// turns a lost update into a visible conflict instead of a blind overwrite.
	result, err := tx.Exec(ctx, `UPDATE course_grants SET source=$2,order_id=$3::text::uuid,state=$4,granted_by=$5::text::uuid WHERE id=$1::text::uuid AND state=$6`, grant.ID(), string(grant.Source()), grant.OrderID(), string(grant.State()), grant.GrantedBy(), string(expected))
	if err != nil {
		return domain.Grant{}, persistenceError(err)
	}
	if result.RowsAffected() != 1 {
		return domain.Grant{}, domain.ErrMembershipConflict
	}
	if _, err := tx.Exec(ctx, `INSERT INTO grant_audit(grant_id,actor_id,action,reason,created_at) VALUES($1::text::uuid,$2::text::uuid,$3,$4,$5)`, grant.ID(), actorID, action, reason, grant.GrantedAt()); err != nil {
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
  ON CONFLICT (user_id,course_id) DO NOTHING
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
			if _, err := tx.Exec(ctx, `UPDATE orders SET refund_required=true,refund_reason='settlement_rejected' WHERE id=$1::text::uuid`, order.ID()); err != nil {
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
	var settledGrantID string
	var inserted bool
	err = tx.QueryRow(ctx, `INSERT INTO course_grants(id,user_id,course_id,source,order_id,state,granted_at) VALUES($1::text::uuid,$2::text::uuid,$3::text::uuid,'purchase',$4::text::uuid,'active',$5)
ON CONFLICT (user_id, course_id) DO UPDATE
SET source='purchase', order_id=EXCLUDED.order_id
WHERE course_grants.source='self' AND course_grants.state='active'
RETURNING id::text, (xmax = 0) AS inserted`, grantID, decision.GrantUserID, decision.GrantCourseID, decision.GrantOrderID, event.ReceivedAt).Scan(&settledGrantID, &inserted)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return persistenceError(err)
	}
	if err == nil {
		// xmax = 0 is the freshly inserted tuple; a non-zero xmax is the self→purchase upgrade
		// of an existing row. This avoids assuming the generated id's text form matches id::text.
		action := "grant"
		if !inserted {
			action = "upgrade"
		}
		if _, err := tx.Exec(ctx, `INSERT INTO grant_audit(grant_id,action,payment_event_id,created_at) VALUES($1::text::uuid,$2,$3::text::uuid,$4)`, settledGrantID, action, eventDatabaseID, event.ReceivedAt); err != nil {
			return persistenceError(err)
		}
	} else {
		if _, updateErr := tx.Exec(ctx, `UPDATE orders SET refund_required=true,refund_reason='grant_conflict' WHERE id=$1::text::uuid`, order.ID()); updateErr != nil {
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
