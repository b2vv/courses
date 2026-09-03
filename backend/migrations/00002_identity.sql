-- +goose Up
CREATE TABLE users (
    id uuid PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'::uuid),
    email text NOT NULL UNIQUE CHECK (
        octet_length(email) BETWEEN 3 AND 254
        AND email = lower(btrim(email))
        AND email ~ '^[^@]+@[^@]+$'
        AND email !~ '[[:space:][:cntrl:]]'
    ),
    display_name text NOT NULL CHECK (
        char_length(display_name) BETWEEN 1 AND 120
        AND display_name = btrim(display_name)
        AND display_name !~ '[[:cntrl:]]'
    ),
    role text NOT NULL CHECK (role IN ('student', 'instructor', 'admin')),
    password_hash text NOT NULL CHECK (
        char_length(password_hash) BETWEEN 60 AND 512
        AND left(password_hash, 10) = '$argon2id$'
    ),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    disabled_at timestamptz,
    CHECK (updated_at >= created_at),
    CHECK (disabled_at IS NULL OR disabled_at BETWEEN created_at AND updated_at)
);

CREATE TABLE user_sessions (
    id uuid PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'::uuid),
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_digest bytea NOT NULL UNIQUE CHECK (octet_length(token_digest) = 32),
    created_at timestamptz NOT NULL,
    expires_at timestamptz NOT NULL,
    CHECK (expires_at > created_at)
);

CREATE INDEX user_sessions_by_user ON user_sessions (user_id, created_at, id);
CREATE INDEX user_sessions_by_expiry ON user_sessions (expires_at);

-- +goose Down
DROP TABLE user_sessions;
DROP TABLE users;
