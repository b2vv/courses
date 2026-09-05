-- name: GetIdentityCredential :one
SELECT id::text AS id, email, display_name, role, password_hash, created_at, updated_at, disabled_at
FROM users
WHERE email = sqlc.arg(email)::text;

-- name: LockActiveIdentityUser :one
SELECT id::text AS id
FROM users
WHERE id = sqlc.arg(user_id)::text::uuid AND disabled_at IS NULL
FOR UPDATE;

-- name: DeleteIdentitySession :execrows
DELETE FROM user_sessions
WHERE token_digest = sqlc.arg(token_digest)::bytea;

-- name: DeleteExpiredIdentitySession :execrows
DELETE FROM user_sessions
WHERE token_digest = sqlc.arg(token_digest)::bytea
  AND expires_at <= sqlc.arg(now)::timestamptz;

-- name: CreateIdentitySession :one
INSERT INTO user_sessions (id, user_id, token_digest, created_at, expires_at)
VALUES (
  sqlc.arg(session_id)::text::uuid,
  sqlc.arg(user_id)::text::uuid,
  sqlc.arg(token_digest)::bytea,
  sqlc.arg(created_at)::timestamptz,
  sqlc.arg(expires_at)::timestamptz
)
RETURNING id::text AS id, user_id::text AS user_id, created_at, expires_at;

-- name: ResolveIdentitySession :one
SELECT
  users.id::text AS id,
  users.email,
  users.display_name,
  users.role,
  users.created_at,
  users.updated_at,
  users.disabled_at
FROM user_sessions
JOIN users ON users.id = user_sessions.user_id
WHERE user_sessions.token_digest = sqlc.arg(token_digest)::bytea
  AND user_sessions.created_at <= sqlc.arg(now)::timestamptz
  AND user_sessions.expires_at > sqlc.arg(now)::timestamptz
  AND users.disabled_at IS NULL;

-- name: ListInstructors :many
SELECT id::text AS id, email, display_name, role, created_at, updated_at, disabled_at
FROM users
WHERE disabled_at IS NULL
  AND role IN ('instructor', 'admin')
ORDER BY display_name COLLATE "C", id
LIMIT 200;
