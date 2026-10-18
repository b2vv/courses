-- +goose Up
SET LOCAL lock_timeout = '5s';

-- Refund outcome: the administrator's decision on a refund-queue order, recorded
-- separately from the membership revoke (the two are decoupled — a refund can be
-- granted or declined with the grant left untouched, and access can be revoked
-- without settling the refund). NULL while the order is still pending in the queue;
-- set to 'refunded' or 'declined' at the same site that stamps refund_resolved_at.
-- Historical resolved rows (before this migration) keep NULL and read "pending"
-- under the state filter until re-settled — none exist in production (test mode).
ALTER TABLE orders
    ADD COLUMN refund_outcome text NULL
        CHECK (refund_outcome IS NULL OR refund_outcome IN ('refunded', 'declined'));

-- +goose Down
ALTER TABLE orders
    DROP COLUMN refund_outcome;
