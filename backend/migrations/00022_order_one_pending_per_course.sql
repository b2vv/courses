-- +goose Up
-- SA05/M2 double-charge: a buyer may hold at most one unpaid pending order per course. Without this,
-- two Idempotency-Keys open two pending orders on the same (user, course); both can be paid and the
-- second only collides on course_grants after the charge, landing in the manual refund queue. The
-- partial unique index refuses the second pending order at the source (mirrors
-- course_grants_one_active_per_user_course). Paid and failed orders carry no pending row, so a retry
-- after a failed payment is unaffected.
SET LOCAL lock_timeout = '5s';

CREATE UNIQUE INDEX orders_one_pending_per_user_course
    ON orders (user_id, course_id)
    WHERE status = 'pending';

-- +goose Down
DROP INDEX orders_one_pending_per_user_course;
