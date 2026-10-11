-- +goose Up
SET LOCAL lock_timeout = '5s';

-- Refund visibility: give orders.refund_required a reason and a resolution record.
-- refund_reason is written at the same two sites that already set refund_required
-- (a rejected paid settlement, or a grant conflict at settle time). Historical rows
-- flagged before this migration keep NULL — the admin screen renders them "unknown".
ALTER TABLE orders
    ADD COLUMN refund_reason text NULL
        CHECK (refund_reason IS NULL OR refund_reason IN ('settlement_rejected', 'grant_conflict')),
    ADD COLUMN refund_resolved_at timestamptz NULL,
    ADD COLUMN refund_resolved_by uuid NULL REFERENCES users(id);

-- +goose Down
ALTER TABLE orders
    DROP COLUMN refund_resolved_by,
    DROP COLUMN refund_resolved_at,
    DROP COLUMN refund_reason;
