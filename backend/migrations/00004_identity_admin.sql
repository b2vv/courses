-- Down drops soft-delete history and may fail after an email has been reused.
-- It is test-only for disposable databases; retained environments recover forward.
-- +goose Up
ALTER TABLE users
    ADD COLUMN deleted_at timestamptz,
    ADD CONSTRAINT users_deleted_at_valid CHECK (
        deleted_at IS NULL OR deleted_at BETWEEN created_at AND updated_at
    ),
    ADD CONSTRAINT users_deleted_implies_disabled CHECK (
        deleted_at IS NULL OR disabled_at IS NOT NULL
    );

ALTER TABLE users DROP CONSTRAINT users_email_key;

CREATE UNIQUE INDEX users_email_active_key ON users (email)
WHERE deleted_at IS NULL;

-- +goose Down
DROP INDEX users_email_active_key;

ALTER TABLE users
    DROP CONSTRAINT users_deleted_at_valid,
    DROP CONSTRAINT users_deleted_implies_disabled,
    DROP COLUMN deleted_at,
    ADD CONSTRAINT users_email_key UNIQUE (email);
