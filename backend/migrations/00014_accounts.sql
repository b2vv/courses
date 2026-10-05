-- +goose Up
SET LOCAL lock_timeout = '5s';

ALTER TABLE users
    ADD COLUMN email_verified boolean NOT NULL DEFAULT false,
    ADD COLUMN totp_secret text NULL,
    ADD COLUMN totp_enabled boolean NOT NULL DEFAULT false;

CREATE TABLE account_tokens (
    id uuid PRIMARY KEY,
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    purpose text NOT NULL CHECK (purpose IN ('email_verify', 'password_reset', 'email_change')),
    token_digest bytea NOT NULL CHECK (octet_length(token_digest) = 32),
    new_email text NULL,
    expires_at timestamptz NOT NULL,
    consumed_at timestamptz NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    CHECK (expires_at > created_at),
    CHECK (
        (purpose = 'email_change' AND new_email IS NOT NULL)
        OR (purpose <> 'email_change' AND new_email IS NULL)
    ),
    UNIQUE (token_digest)
);
CREATE INDEX account_tokens_by_user_purpose ON account_tokens (user_id, purpose);

CREATE TABLE recovery_codes (
    id uuid PRIMARY KEY,
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    code_hash text NOT NULL,
    consumed_at timestamptz NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX recovery_codes_by_user ON recovery_codes (user_id);

-- +goose Down
DROP TABLE recovery_codes;
DROP TABLE account_tokens;
ALTER TABLE users
    DROP COLUMN totp_enabled,
    DROP COLUMN totp_secret,
    DROP COLUMN email_verified;
