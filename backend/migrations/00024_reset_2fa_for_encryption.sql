-- +goose Up
-- SA13-I / I-F6 changes the at-rest format of two-factor secrets: totp_secret becomes AES-256-GCM
-- ciphertext and recovery codes become keyed HMACs. Reads are fail-closed (a value that does not
-- decrypt / was hashed differently is rejected, never treated as plaintext), so any row written under
-- the old format (plaintext secret, sha256 recovery hash) would break that account's 2FA login after
-- deploy. The product is pre-launch with no production data (seed sets totp_secret = NULL), so this is
-- a one-time clean slate: it is a no-op on a freshly migrated+seeded database and only clears stale
-- 2FA state from a persistent dev/CI database. Affected accounts simply re-enrol 2FA; no secret is
-- recoverable from the old format anyway. Forward-only by owner decision (2026-10-30), resting on the
-- pre-launch "cut freely" decision (2026-10-08).
SET LOCAL lock_timeout = '5s';

UPDATE users
SET totp_secret = NULL,
    totp_enabled = false,
    totp_last_counter = 0
WHERE totp_secret IS NOT NULL
   OR totp_enabled
   OR totp_last_counter <> 0;

DELETE FROM recovery_codes;

-- +goose Down
-- Irreversible: the reset discards 2FA enrolment state that cannot be reconstructed. The down step is
-- intentionally empty so a downgrade leaves the (already-cleared) state as-is.
SELECT 1;
