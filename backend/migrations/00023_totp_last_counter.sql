-- +goose Up
-- SA13-B/I-F5 TOTP replay: without a record of which 30-second step a code belongs to, a valid TOTP
-- can be replayed within its acceptance window (the current step plus the +/-1 skew the verifier
-- allows, ~90s). This column stores the highest TOTP step already consumed for the account; a
-- successful proof advances it with a conditional UPDATE (last_counter < new), so the same step --
-- hence the same code -- can never be accepted twice, and the check is atomic against a concurrent
-- replay. Steps are unix-time/30, monotonic in wall-clock, so the default 0 ("never used") is below
-- any real step and a fresh enrolment's first code is accepted.
SET LOCAL lock_timeout = '5s';

ALTER TABLE users
    ADD COLUMN totp_last_counter bigint NOT NULL DEFAULT 0;

-- +goose Down
ALTER TABLE users
    DROP COLUMN totp_last_counter;
