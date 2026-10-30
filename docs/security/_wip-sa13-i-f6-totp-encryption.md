# SA13-I / I-F6 — TOTP secret encryption at rest + recovery-code keyed-hash

**Status:** planning (pre-code, post plan-critique). **Branch:** TBD `sa13-totp-encryption` off `main` @ `174b7a8`.
**Review:** `review_mode=existing_session`, reviewer `courses-96 [557618]` (mandatory per task).

## Finding (audit §LOW, `docs/security/audit-2026-10-24.md:129`)
- **I-F6a — `users.totp_secret` plaintext.** 160-bit base32 TOTP secret stored as-is. DB leak →
  attacker generates valid TOTP → bypasses 2FA.
- **I-F6b — recovery codes unsalted sha256 over 40-bit space.** `recoveryCodeBytes=5` → 40 bits;
  hash = `sha256(code)`, hex-stored. On DB leak: 2⁴⁰≈1.1·10¹² hashes; commodity GPU
  (~10¹⁰ sha256/s) cracks one code in ~110 s. **Magnitude**: offline-crackable.

## Owner decisions
1. **(2026-10-30, REVISED after plan-critique + critic debate) Recovery codes → HMAC-SHA256 keyed
   hash**, not argon2id. Rationale: this phase already introduces a server-side key for TOTP-secret
   encryption; a keyed MAC makes 40-bit codes unforgeable without the key regardless of entropy, and
   **keeps the existing atomic indexed by-hash consume** (dissolves two blocking correctness issues
   and the argon2id DoS vector). The only scenario argon2id would protect recovery codes better is a
   **DB+key leak** — but there the co-located TOTP secret is already decryptable, so 2FA is already
   lost; argon2id would defend an asset already gone. Superseded decision: 2026-10-30 argon2id+salt.
2. **(2026-10-30) 2FA secret key = required in runtime** (fail-closed). Missing/invalid → api does
   not start. No plaintext fallback, no dual-format read.
3. **(2026-10-30) Forward-only** (rests on `pre-launch-cut-freely`, owner 08.10): product not
   deployed, no prod data, `seed-identities.sql` sets `totp_secret=NULL`. No backfill command / lazy
   dual-read. Migration 00024 resets any dev/CI-DB 2FA state to a clean slate under the new format.

## Design

### Key material — one master env, HKDF domain separation
- Env `IDENTITY_2FA_SECRET_KEY` (base64 of ≥32 random bytes master). Renamed from the earlier
  `TOTP_ENCRYPTION_KEY` because one master now keys **both** TOTP encryption and recovery HMAC.
- `config.Identity2FAKey(getenv) ([]byte, error)` in `backend/platform/config` (mirrors `dsn.go`):
  base64-decode, require ≥32 bytes, else error.
- **HKDF-SHA256 domain separation** (NIST SP 800-108 hygiene; never reuse one raw key across two
  primitives): derive `k_aes = HKDF(master, info="totp-secret-aes256gcm")` and
  `k_hmac = HKDF(master, info="recovery-code-hmac")`, each 32 B. `golang.org/x/crypto/hkdf` (already
  transitively present via x/crypto/argon2). **One** required env key; not two.

### Adapter package `backend/internal/identity/adapters/identitycrypto`
- `New(master []byte) (*Keys, error)` — validates master ≥32 B; HKDF-derives k_aes, k_hmac; builds
  `cipher.NewGCM` once.
- `(*Keys).EncryptSecret(plaintext string) (string, error)` — AES-256-GCM, random 12-B nonce,
  returns `base64(nonce || ciphertext||tag)`.
- `(*Keys).DecryptSecret(encoded string) (string, error)` — GCM-opens (auth-tag verified);
  tamper/wrong-key → error.
- `(*Keys).RecoveryHash(code string) string` — `hex(HMAC-SHA256(k_hmac, code))` — deterministic,
  32-byte → 64-hex (same width as the current sha256 hex, so `code_hash text` column unchanged).
- Concrete `*identitycrypto.Keys` is injected into the repository as a **concrete type, not an
  interface** (plan-critique N1: no failing-fake consumer exists; AGENTS "no unused seams"). Unit
  tests hit the package directly; integration/e2e use a real `Keys` from a fixed test master.

### A. TOTP secret (encrypt-on-write, decrypt-on-read)
- `postgres.Repository` gains a `*identitycrypto.Keys` field.
  - `SetTotpSecret`: `EncryptSecret` before write.
  - `GetTwoFactor`: `totp_secret` NULL → `Secret=""` (unchanged, **no decrypt call** — plan-critique
    verified `GetUserTwoFactor` is the sole reader); else `DecryptSecret`; decrypt error →
    `persistenceError` (never silently treat as plaintext — forward-only).
- `GetUserTwoFactor` (identity.sql) is the ONLY reader of `totp_secret`; mapResolved /
  ResolveIdentitySession read only `totp_enabled`. No hidden reader bypasses decrypt.
- **No new request-time external dependency**: HKDF/AES-GCM/HMAC are in-process (stdlib + x/crypto);
  the master key is read **once at startup** and fails closed (§D). A failure mode at request time is
  only a decrypt error on tampered/wrong-key data → `persistenceError`/503 (same envelope as any
  persistence fault), never a hang or partial auth.

### B. Recovery codes (HMAC keyed, existing atomic consume kept)
- `Passwords` interface reuse is **dropped** for recovery (that was the argon2id path). Hashing moves
  to `identitycrypto.RecoveryHash` (adapter concern, holds k_hmac).
- Port `RecoveryCodeStore` from digests to plaintext (hashing is an adapter concern):
  - `InsertCodes(ctx, userID string, codes []string) error`
  - `ConsumeCode(ctx, userID, code string) (bool, error)`
  - `DeleteCodes` unchanged.
- Application `two_factor.go`: drop `recoveryDigest`/`recoveryDigests` (+ `crypto/sha256` import).
  Update **all 4 call sites** (plan-critique N5): `InsertCodes` at `ConfirmTwoFactor` (:109) and
  `RegenerateRecoveryCodes` (:239); `ConsumeCode` at `CompleteTwoFactor` (:167) and
  `authorizeFactor` (:276) — pass plaintext `input.Code`/codes.
- Repository:
  - `InsertCodes`: `RecoveryHash(code)` per code → existing `InsertRecoveryCode` (in tx, as now).
  - `ConsumeCode`: `RecoveryHash(code)` → existing `ConsumeRecoveryCode` by-hash query **unchanged**
    (`UPDATE … code_hash=$ AND consumed_at IS NULL RETURNING id`; already maps `pgx.ErrNoRows`→false,
    so the concurrent-replay/already-consumed branch is already correct — plan-critique blocking #2/#3
    **do not arise** under HMAC).
  - SQL queries **unchanged** (`code_hash text`, hex width unchanged). No new queries, no app-loop.
- **Boundary/dup/crooked input:** empty live set (not enrolled / all consumed) → by-hash lookup
  misses → `ErrNoRows`→false → uniform `ErrInvalidTOTP`. The by-hash `:one` consume assumes no
  duplicate `code_hash`; HMAC is deterministic, so this relies on `GenerateRecoveryCodes`' existing
  plaintext dedup (distinct codes → distinct HMAC). An oversized/garbage recovery input is a single
  cheap HMAC with no match (no length limit, no error branch — unlike argon2id's `Verify(false,err)`),
  so crooked input surfaces as the same uniform failure with no oracle.

### C. Migration 00024 (forward-only 2FA reset)
- Up: `SET LOCAL lock_timeout='5s';` then `UPDATE users SET totp_secret=NULL, totp_enabled=false,
  totp_last_counter=0;` + `DELETE FROM recovery_codes;` — clean slate so no stale plaintext secret /
  sha256 hash survives the format change (fail-closed decrypt would otherwise 503 those users).
  No-op on a fresh DB (seed has no 2FA rows); real only for a persistent dev/CI DB with prior
  enrollments — kept as cheap insurance (plan-critique N3 acknowledged).
- Down: comment-only (irreversible data reset). Valid Up/Down so the migration-preservation test
  (seed after DownTo) still passes (memory `seed-sql-must-load-on-downgraded-schema`).
- Persisted-state change → **full e2e**, not targeted (memory
  `integration-tag-skipped-by-go-test-and-e2e`).
- **Rollback:** reverting the PR restores the prior behaviour (storage format is read fresh each
  request; no in-memory migration state). 00024's data reset is **not** undone by a code revert, but
  is harmless — affected users re-enrol 2FA; no corruption, no prod data. The only irreversible effect
  is that reset, and it is bounded to test-mode.

### D. Env wiring + constructor blast radius
- `postgres.NewRepository(pool, passwords, keys *identitycrypto.Keys)` — **~22 call sites** (not 2;
  plan-critique blocking #1): cmd/api/main.go:131; course/authoring_integration_test.go:112;
  identity/two_factor_integration_test.go:69; schema_integration_test.go (×8: 92/312/342/372/459/
  473/514/533) + `newRepository` (396/442/556); administration_integration_test.go:92/491. All must
  pass a real `*Keys` (identity integration tests exercise SetTotpSecret/ConsumeCode → a nil keys
  panics on decrypt-on-read). Add a test helper `testKeys()` building `identitycrypto.New(<fixed 32-B
  master>)`; use it at every identity/course integration site.
- Composition root `cmd/api/main.go`: read `config.Identity2FAKey` near other `ConfigFromEnvironment`
  calls; on error `logger.Error(...)` + `os.Exit(1)` (fail-closed). Build `identitycrypto.New`; pass
  to `NewRepository`.
- `docker-compose.yml` api env: `IDENTITY_2FA_SECRET_KEY: ${IDENTITY_2FA_SECRET_KEY:-<fixed 32-B
  base64 dev key>}` (dev/test default, like `COOKIE_SECURE:-false`). e2e/ci overlays inherit the base
  api env block (verified by Edge lens: base+overlay via COMPOSE_FILE) — confirm during S4.
- `.env.example`: document `IDENTITY_2FA_SECRET_KEY` (base64 of 32 random bytes; **prod must set a
  real secret**; dev default is not secret).

## Tests (production path, not fakes)
- Unit `identitycrypto`: encrypt→decrypt round-trip; tamper (flip a byte) → decrypt fails; wrong
  master → decrypt fails; `RecoveryHash` deterministic + differs across distinct codes; HKDF yields
  distinct k_aes≠k_hmac (no cross-use); short master rejected.
- Unit `config.Identity2FAKey`: unset→err, bad base64→err, <32 B→err, ≥32 B→ok.
- Integration (`two_factor_integration_test.go`, real Postgres via `testKeys()`): enroll→confirm →
  **assert DB `totp_secret` column ≠ plaintext secret AND `DecryptSecret` round-trips it**; TOTP login
  works; recovery login works; consumed recovery code can't be reused; wrong recovery code → false;
  `code_hash` is HMAC hex (not the sha256 of the code — assert `RecoveryHash(code)` matches the row).
- e2e Acc 2FA scenarios already exercise enroll/confirm/login/recovery → now run through encryption +
  HMAC (real path), key supplied via compose default.
- Update `application/two_factor_test.go` mock `recoveryStore` to `[]string`/`string` (plan-critique
  N4) and re-check its assertions.

## Limitations (§report; each carries a mechanism)
- **Wrong-but-valid master key not detected until first decrypt** (plan-critique N9): fail-closed
  checks key *format* (≥32 B), not identity. A different valid 32-B master passes startup but makes
  every enrolled user's `GetTwoFactor` decrypt-fail → 2FA logins 503 at runtime. Mechanism: no
  startup canary. Out of scope here (no prod data; 00024 resets dev); note for the deploy/rotation
  tail.
- **One master guards both TOTP secret and recovery codes** (accepted): a master leak compromises
  both — but a master leak already decrypts the TOTP secret, i.e. 2FA is already lost, so recovery
  codes falling adds nothing operationally.
- **Migration 00024 invalidates an in-flight 2FA login** (dev-only): a user mid-2FA after the reset
  has `Enabled=false` → must re-login with password only. Acceptable in test-mode.
- **(review N1, accepted-bar 2026-10-30) GCM uses no AAD → ciphertext not bound to user_id.** A
  DB-**write** attacker could transplant one user's `totp_secret` ciphertext onto another row; it
  would decrypt and that user's authenticator would then work for the victim. Mechanism: pass
  `userID` as GCM additional-authenticated-data in `EncryptSecret`/`DecryptSecret`. Out of the stated
  threat model (DB-**read** leak; a DB-write capability is already a larger compromise), so accepted
  for now — a cheap defense-in-depth follow-up slice, not a blocker.
- **(review N2) No version/key-id in the stored format → master-key rotation re-enrols everyone.** The
  base64(nonce‖ct‖tag) blob carries no key id, so rotating `IDENTITY_2FA_SECRET_KEY` makes every
  stored secret undecryptable (fail-closed) until re-enrolment. Mechanism for a future rotation path:
  prefix a key-id/version byte and keep the old key during a window. Acceptable pre-launch
  (forward-only, no prod data); belongs with the deploy/rotation tail.

## Gate (ordered `make check` + full e2e; impact = persisted + shared)
`make format-check` → `lint` → `typecheck` → `test` → `integration` → `build` → `e2e`.
Read `GATE_EXIT` from log (memory `background-exit-code-is-last-pipe`); no nested `&`
(`no-nested-ampersand-in-backgrounded-bash`). sqlc unchanged (no query change) → but if any generated
touched, `git add` before `check-generated` (memory `check-generated-needs-staging`).

## Pre-code checkpoint
- READY: finding cited w/ magnitude; owner forks resolved (3, #1 revised post-debate); design maps to
  concrete files; blocking #1 addressed (full site list + test helper); blocking #2/#3 dissolved by
  HMAC; tests name the production path; gate defined; forward-only grounded in dated decision.
- Risks: (1) ~22 constructor sites — enumerate + testKeys() helper. (2) e2e/ci compose must carry the
  key — verify overlay inheritance in S4. (3) decrypt-on-read must not fall back to plaintext.
  (4) HKDF info labels fixed/distinct. (5) env rename TOTP_ENCRYPTION_KEY→IDENTITY_2FA_SECRET_KEY
  (greenfield name; no prior consumer since this is the introducing phase).
- Reversible? A/B/D behavioural-yes; C irreversible data reset but no prod data.

## Slices (each = own proof; small commits)
- **S1:** `identitycrypto` pkg (HKDF + AES-GCM + RecoveryHash) + `config.Identity2FAKey` + unit tests.
- **S2:** repository TOTP encrypt/decrypt + recovery HMAC + port→plaintext + app call sites (4) +
  update app unit mock + `NewRepository(keys)` + all ~22 sites + `testKeys()` helper.
- **S3:** migration 00024 + compose/.env wiring + integration asserts (ciphertext≠plaintext, HMAC
  row).
- **S4:** full ordered `make check` + e2e; confirm e2e/ci carry the key.
Then: PR → independent review courses-96 (PASS 0 blocking) → owner merge.

## Plan-defense (gate)
| # | Питання | Закрито рядком | Вердикт |
|---|---|---|---|
| 1 | кривий інпут | §Key material (malformed key→fail-closed), §A (tampered ciphertext→503), §B (garbage recovery code→uniform no-match) | закрито |
| 2 | відмова залежності | §A (no new request-time dep; key read once at startup, fail-closed), §D (startup fail-closed) | закрито |
| 3 | гонка й повтор | §B (existing atomic `UPDATE … consumed_at IS NULL RETURNING`, one winner; ErrNoRows→false) | закрито |
| 4 | порожні/граничні/дублі | §A (NULL secret→no decrypt), §B (empty set→false; dup relies on GenerateRecoveryCodes dedup) | закрито |
| 5 | rollback | §C (revert PR behavioural; 00024 reset irreversible but harmless, test-mode only) | закрито |

**Дірок: 0 → гейт зелений.** План іде в таски (S1–S4).

## Outcome
- **MERGED** as PR #90 (squash `c22c9d4`), 2026-10-30. Ordered `make check` GATE_EXIT=0 (unit +
  integration identity 20.7s with new ciphertext/HMAC asserts + check-generated clean + build + e2e
  127 passed). Independent review `courses-96 [557618]` (existing_session, read-only via gh pr diff on
  shared tree) **PASS 0 blocking** on `c17ef35`; 2 non-blocking (N1 GCM-AAD, N2 key-version) recorded
  in §Limitations as accepted bars. **I-F6 closed → SA13 batch fully closed.**

## Journal
- 2026-10-30: plan written; ran plan-critique (4 lenses: 3 blocking + ~10 non-blocking) + an
  adversarial critic on recovery-hash choice. Owner revised decision #1 argon2id→HMAC+HKDF after the
  debate. Plan rewritten: HMAC keeps the atomic indexed consume (blocking #2/#3 dissolved); blocking
  #1 (22 sites) folded into S2 with a testKeys() helper; non-blockings folded (concrete Cipher type,
  limitations §, call-site list, app-mock update). Next: plan-defense gate, then S1.
