-- +goose Up
SET LOCAL lock_timeout = '5s';

ALTER TABLE orders
    ADD COLUMN checkout_opened_at timestamptz NULL,
    ADD COLUMN refund_required boolean NOT NULL DEFAULT false;

-- +goose Down
ALTER TABLE orders
    DROP COLUMN refund_required,
    DROP COLUMN checkout_opened_at;
