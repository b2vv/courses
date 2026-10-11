-- name: GetIdentityCredential :one
SELECT id::text AS id, email, display_name, role, password_hash, created_at, updated_at, disabled_at
FROM users
WHERE email = sqlc.arg(email)::text
  AND deleted_at IS NULL;

-- name: GetUserTwoFactor :one
SELECT totp_secret, totp_enabled
FROM users
WHERE id = sqlc.arg(user_id)::text::uuid
  AND deleted_at IS NULL;

-- name: SetUserTotpSecret :execrows
UPDATE users
SET totp_secret = sqlc.arg(totp_secret)::text,
    totp_enabled = false,
    updated_at = now()
WHERE id = sqlc.arg(user_id)::text::uuid
  AND deleted_at IS NULL;

-- name: EnableUserTotp :execrows
UPDATE users
SET totp_enabled = true,
    updated_at = now()
WHERE id = sqlc.arg(user_id)::text::uuid
  AND totp_secret IS NOT NULL
  AND deleted_at IS NULL;

-- name: DisableUserTotp :execrows
UPDATE users
SET totp_secret = NULL,
    totp_enabled = false,
    updated_at = now()
WHERE id = sqlc.arg(user_id)::text::uuid
  AND deleted_at IS NULL;

-- name: InsertRecoveryCode :exec
INSERT INTO recovery_codes (id, user_id, code_hash)
VALUES (sqlc.arg(id)::text::uuid, sqlc.arg(user_id)::text::uuid, sqlc.arg(code_hash)::text);

-- name: ConsumeRecoveryCode :one
UPDATE recovery_codes
SET consumed_at = now()
WHERE user_id = sqlc.arg(user_id)::text::uuid
  AND code_hash = sqlc.arg(code_hash)::text
  AND consumed_at IS NULL
RETURNING id::text AS id;

-- name: DeleteRecoveryCodesForUser :execrows
DELETE FROM recovery_codes
WHERE user_id = sqlc.arg(user_id)::text::uuid;

-- name: CreateUser :one
INSERT INTO users (id, email, display_name, role, password_hash, email_verified, created_at, updated_at)
VALUES (
  sqlc.arg(user_id)::text::uuid,
  sqlc.arg(email)::text,
  sqlc.arg(display_name)::text,
  sqlc.arg(role)::text,
  sqlc.arg(password_hash)::text,
  sqlc.arg(email_verified)::boolean,
  sqlc.arg(created_at)::timestamptz,
  sqlc.arg(updated_at)::timestamptz
)
RETURNING id::text AS id, email, display_name, role, email_verified, created_at, updated_at, disabled_at;

-- name: SetUserEmailVerified :execrows
UPDATE users
SET email_verified = true,
    updated_at = now()
WHERE id = sqlc.arg(user_id)::text::uuid
  AND deleted_at IS NULL;

-- name: UpdateUserPassword :execrows
UPDATE users
SET password_hash = sqlc.arg(password_hash)::text,
    updated_at = now()
WHERE id = sqlc.arg(user_id)::text::uuid
  AND deleted_at IS NULL;

-- name: UpdateUserEmail :execrows
UPDATE users
SET email = sqlc.arg(email)::text,
    email_verified = true,
    updated_at = now()
WHERE id = sqlc.arg(user_id)::text::uuid
  AND deleted_at IS NULL;

-- name: InsertAccountToken :exec
INSERT INTO account_tokens (id, user_id, purpose, token_digest, new_email, expires_at, created_at)
VALUES (
  sqlc.arg(token_id)::text::uuid,
  sqlc.arg(user_id)::text::uuid,
  sqlc.arg(purpose)::text,
  sqlc.arg(token_digest)::bytea,
  sqlc.narg(new_email)::text,
  sqlc.arg(expires_at)::timestamptz,
  sqlc.arg(created_at)::timestamptz
);

-- name: ConsumeAccountToken :one
UPDATE account_tokens
SET consumed_at = now()
WHERE token_digest = sqlc.arg(token_digest)::bytea
  AND purpose = sqlc.arg(purpose)::text
  AND consumed_at IS NULL
  AND expires_at > now()
RETURNING user_id::text AS user_id, new_email;

-- name: InvalidateAccountTokensByPurpose :execrows
UPDATE account_tokens
SET consumed_at = now()
WHERE user_id = sqlc.arg(user_id)::text::uuid
  AND purpose = sqlc.arg(purpose)::text
  AND consumed_at IS NULL;

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
  users.email_verified,
  users.totp_enabled,
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

-- name: ListAdminUsers :many
SELECT id::text AS id, email, display_name, role, created_at, updated_at, disabled_at, deleted_at
FROM users
WHERE (
    (sqlc.arg(status)::text = '' AND deleted_at IS NULL)
    OR (sqlc.arg(status)::text = 'active' AND deleted_at IS NULL AND disabled_at IS NULL)
    OR (sqlc.arg(status)::text = 'blocked' AND deleted_at IS NULL AND disabled_at IS NOT NULL)
    OR (sqlc.arg(status)::text = 'deleted' AND deleted_at IS NOT NULL)
  )
  AND (sqlc.arg(role_filter)::text = '' OR role = sqlc.arg(role_filter)::text)
  AND (
    email ILIKE '%' || sqlc.arg(search_pattern)::text || '%' ESCAPE '!'
    OR display_name ILIKE '%' || sqlc.arg(search_pattern)::text || '%' ESCAPE '!'
  )
ORDER BY
  CASE WHEN sqlc.arg(sort_by)::text = 'name' THEN display_name COLLATE "C" END ASC,
  CASE WHEN sqlc.arg(sort_by)::text = 'recent' THEN created_at END DESC,
  email COLLATE "C", id
LIMIT sqlc.arg(page_limit)::integer OFFSET sqlc.arg(page_offset)::integer;

-- name: CountAdminUsers :one
SELECT count(*)
FROM users
WHERE (
    (sqlc.arg(status)::text = '' AND deleted_at IS NULL)
    OR (sqlc.arg(status)::text = 'active' AND deleted_at IS NULL AND disabled_at IS NULL)
    OR (sqlc.arg(status)::text = 'blocked' AND deleted_at IS NULL AND disabled_at IS NOT NULL)
    OR (sqlc.arg(status)::text = 'deleted' AND deleted_at IS NOT NULL)
  )
  AND (sqlc.arg(role_filter)::text = '' OR role = sqlc.arg(role_filter)::text)
  AND (
    email ILIKE '%' || sqlc.arg(search_pattern)::text || '%' ESCAPE '!'
    OR display_name ILIKE '%' || sqlc.arg(search_pattern)::text || '%' ESCAPE '!'
  );

-- name: GetAdministrationUser :one
SELECT id::text AS id, email, display_name, role, created_at, updated_at, disabled_at, deleted_at
FROM users
WHERE id = sqlc.arg(user_id)::text::uuid
  AND deleted_at IS NULL;

-- name: LockAdmins :many
SELECT id::text AS id, disabled_at
FROM users
WHERE role = 'admin' AND deleted_at IS NULL
ORDER BY id
FOR UPDATE;

-- name: LockUserForStatus :one
SELECT id::text AS id, disabled_at
FROM users
WHERE id = sqlc.arg(user_id)::text::uuid
  AND deleted_at IS NULL
FOR UPDATE;

-- name: SetUserDisabledAt :execrows
UPDATE users
SET disabled_at = sqlc.narg(disabled_at)::timestamptz,
    updated_at = sqlc.arg(updated_at)::timestamptz
WHERE id = sqlc.arg(user_id)::text::uuid
  AND deleted_at IS NULL;

-- name: SoftDeleteUser :execrows
UPDATE users
SET deleted_at = sqlc.arg(deleted_at)::timestamptz,
    disabled_at = COALESCE(disabled_at, sqlc.arg(deleted_at)::timestamptz),
    updated_at = sqlc.arg(deleted_at)::timestamptz
WHERE id = sqlc.arg(user_id)::text::uuid
  AND deleted_at IS NULL;

-- name: SetUserPasswordHash :execrows
UPDATE users
SET password_hash = sqlc.arg(password_hash)::text,
    updated_at = sqlc.arg(updated_at)::timestamptz
WHERE id = sqlc.arg(user_id)::text::uuid
  AND deleted_at IS NULL;

-- name: DeleteIdentitySessionsForUser :execrows
DELETE FROM user_sessions
WHERE user_id = sqlc.arg(user_id)::text::uuid;

-- name: DeleteUserSessionsExceptCurrent :execrows
DELETE FROM user_sessions
WHERE user_id = sqlc.arg(user_id)::text::uuid
  AND token_digest <> sqlc.arg(token_digest)::bytea;
