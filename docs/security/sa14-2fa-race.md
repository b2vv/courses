# SA14 — 2FA enrolment/confirmation race

| | |
|---|---|
| Date | 2026-11-28 (rev 2 the same day, after plan review R1 BLOCKED) |
| Source | `docs/design-pass/plan.md` §18, limitation Л5 («Серверна гонка … окрема задача»). Owner: «Fix 2FA» (2026-11-28) |
| Root | `main` @ `688caf4b`; branch `fix/sa14-2fa-race` from `main`; one writer (Courses 1). PR #124 (design-pass ACCT) is open separately; it touches no backend file but does touch `frontend/e2e/accounts.spec.ts` (see A5) |
| Status | see «Readiness» |

## Problem

Each 2FA write reads the state, decides, then writes without re-checking the state it read, and the writes that belong together commit separately.

- **R1 (Л5).** `EnrollTwoFactor` (`backend/internal/identity/application/two_factor.go:58-82`) reads `Enabled == false` and then calls `SetTotpSecret`. `SetUserTotpSecret` (`backend/sql/queries/identity.sql:13-19`) sets the new secret and `totp_enabled = false` with no condition. Suppose tab B confirms 2FA between tab A's read and A's write. A's enrolment then silently turns 2FA off and replaces the secret, while B's recovery codes stay live. The owner believes 2FA is on.
- **R2.** `ConfirmTwoFactor` (`two_factor.go:84-111`) verifies the code against the secret it read (S1) and then calls `EnableTotp`. `EnableUserTotp` (`identity.sql:21-27`) enables whatever secret is stored now. If tab B enrols S2 in between, A enables S2, which A never saw: the owner is locked out.
- **R3.** Two confirmations that both pass verification (two different steps of the same secret) both reach `EnableTotp` and `InsertCodes`. Inserts append (`repository.go:398-420`; the table has no generation column, `migrations/00014_accounts.sql:27-33`), so **both** returned sets stay valid until used or deleted, and the user holds one of them unknowingly (review R1 N1).
- **R4 (review R1 B1).** Confirmation advances the TOTP step counter (`verifyAndConsumeTOTP`, `two_factor.go:290-296`) before enabling, and the counter update matches neither the secret nor the enabled flag (`identity.sql:38-47`). A stale confirmation that is about to fail still consumes step N, so a valid code for the replacement secret at step N is then refused; a losing confirmation can likewise consume the winner's next step.
- **R5 (review R1 B2).** Enable and code issuance commit separately (`two_factor.go:101-111`), and so do disable and code deletion (`two_factor.go:221-224, 248-253`), and code regeneration's delete and insert (`two_factor.go:231-240`). A confirmation delayed after an admin reset inserts live codes for an account with 2FA off; login completion accepts recovery codes even when the factor is off (`two_factor.go:166-178`). A failure after enable leaves 2FA on with no codes delivered, and a retry answers 409. A disable whose code deletion runs after a new enrol + confirm deletes the new codes.

## Invariant

After SA14, for every committed state of a user row:

- **I1.** Recovery codes exist only while `totp_enabled = true`, and every live code was issued for the currently stored secret.
- **I2.** `totp_enabled` becomes true only for the secret whose code was verified, and only together with the recovery-code batch returned to that caller and the step that caller consumed.
- **I3.** An operation that is refused changes nothing: secret, flag, counter and recovery codes are as before.

Every writer of `totp_secret`, `totp_enabled`, `totp_last_counter` (on the confirm path) or `recovery_codes` holds the user's row lock (`SELECT … FOR UPDATE`) for the whole transition and decides on what it read under the lock. The writers are listed in «Design»; no other writer exists (`grep` in «Scoped impact»).

## Scope and acceptance

| # | Observable behaviour | Proof (each test must fail when its guard is reverted — see «Verification plan») |
|---|---|---|
| A1 | Enrolment never disables 2FA that is already on. If 2FA is enabled under the lock, the request answers `409 twofactor_already_enabled`, and secret, flag, counter and recovery codes are unchanged | integration (PostgreSQL, real repository): enable with a known code set, then `SetTotpSecret` with a new secret → refused; assert stored ciphertext, flag, counter and the same code rows. Application unit: fake reports refusal → `ErrTwoFactorAlreadyEnabled` |
| A2 | Confirmation enables 2FA only for the secret its code was checked against. If the secret changed or was cleared before the confirmation takes the lock, it answers `422 invalid_totp`, 2FA stays off, **and the counter is unchanged**, so a valid code of the current secret at the same step then succeeds | integration: read S1, enrol S2, confirm with S1's version and a step-N code → `SecretChanged`, counter unchanged, no codes; then confirm S2 at step N with S2's version → enabled, counter = N, one code batch |
| A3 | Of two confirmations that both read the pending state, exactly one enables 2FA and returns codes; the other answers `409 twofactor_already_enabled`, inserts no codes and leaves the counter at the winner's step. Holds for the same step and for distinct steps | integration, deterministic interleaving: the application service over the real repository, with a barrier wrapper that releases both `GetTwoFactor` reads only after both completed. Cases: same step, distinct steps (N, N+1 in either order). Assert one success + one `ErrTwoFactorAlreadyEnabled`; `recovery_codes` holds exactly the winner's 10 codes (hash match); counter = winner's step |
| A4 | Reset is atomic with respect to confirmation: after an admin reset or a self-disable commits, 2FA is off and no recovery code is live, whichever side of the confirmation it lands on. A confirmation whose lock comes after the reset answers `422 invalid_totp` | integration: (a) confirm after reset (read before, write after) → 422, off, 0 codes; (b) reset after a committed confirm → off, 0 codes, counter 0; (c) self-disable with the version of a replaced secret → refused, the new pending secret is untouched |
| A5 | A failed write leaves nothing behind: a failure during code issuance rolls back the enable and the counter; the response is `503 dependency_unavailable` and a retry can succeed with a fresh code | integration: repository with an injected `random` reader that fails on the k-th UUID → `ErrDependencyUnavailable`; flag off, counter unchanged, 0 codes; then a normal confirm succeeds |
| A6 | Regeneration replaces codes only for the factor that authorised it: if the factor was reset or replaced after the proof, it answers `409 twofactor_not_enabled` and inserts nothing | integration: authorise, reset, regenerate with the old version → refused, 0 codes |
| A7 | Everything else is unchanged: enrol → confirm → login → disable → re-enrol, admin disable, regeneration, the SA13-B rule that the enrolment code cannot be replayed as the first login | existing unit tests, `two_factor_integration_test.go`, e2e `accounts.spec.ts` Acc26–Acc29, Acc31; the admin-reset journey waits for the recovery codes before the reset (review R1 N4: today the re-confirm click is not awaited, so resetting a merely pending factor also passes) |

Non-goals and residuals (each with its mechanism; none is closed by SA14):

- **N-a. Writes after an account block.** Factor writes match `deleted_at IS NULL` but not `disabled_at`; an authenticated request already past session resolution can still write after a block commits. New sessions and rotation are already refused (`repository.go:131-175`). SA14 adds `disabled_at IS NULL` to its locked reads, so a blocked account's confirm/enrol/regenerate/self-disable answers `404 user_not_found`, as a deleted one does — **decision D4**.
- **N-b. Login proof across a reset.** `CompleteTwoFactor` consumes the step and then rotates the session; a reset + re-enrol between them leaves the session issued on the old factor's proof (`two_factor.go:151-185`). The pending login carries no factor version. Mechanism to close it later: carry `Version` in `PendingTwoFactor` and check it in `Rotate`. Out of scope: login is not an enrol/confirm race.
- **N-c. Enrolment after an admin reset.** A paused enrolment can store a new pending secret after a reset. It cannot turn 2FA on without a confirmation from the same user; no action.
- **N-d. Commit with a lost response.** If the confirm commit succeeds but the response is lost, a retry answers 409 and the codes were never shown. Recovery exists: the user proves the factor and regenerates (`POST /auth/2fa/recovery-codes`).
- **N-e. 404 is undeclared.** The handler maps `ErrUserNotFound` to 404 (`handler.go:823`), but the enrol and confirm declarations omit it (`openapi_contract.go:89-102`). Pre-existing; SA14 does not change the wire contract, and the generated client is not regenerated here. Recorded as a contract gap for a separate change.
- The frontend: it already maps 409 and 422. The V5a reload on 409 and «Скасувати» on a standalone 422 exist on the ACCT branch (`473520fe`, PR #124), not on `main` (review R1 N2). The V5a user experience is proved on that ref; SA14's backend behaviour does not depend on it.

## Design

**Ports** (`backend/internal/identity/ports/two_factor.go`). The adapter owns every multi-row transition; the application keeps verification and code generation.

- `TwoFactor` gains `Version string`: the stored ciphertext as read, opaque to the application, `""` when no secret. The application only forwards it; it is never logged or returned.
- `SetTotpSecret(ctx, userID, secret) (bool, error)`: `false` = 2FA is enabled now.
- `ConfirmTotp(ctx, ConfirmTotp{UserID, Version, Counter, Codes}) (ConfirmTotpOutcome, error)`, outcome ∈ `Enabled`, `AlreadyEnabled`, `SecretChanged`, `StepUsed`.
- `DisableTotp(ctx, userID, version) (bool, error)` (self-disable: refuses if the stored secret is not `version` or 2FA is off) and `ResetTotp(ctx, userID) error` (admin: unconditional). Both clear the secret, flag and counter **and delete the codes** in the same transaction.
- `ReplaceRecoveryCodes(ctx, userID, version, codes) (bool, error)`: refuses unless 2FA is on with `version`.
- `RecoveryCodeStore` keeps `ConsumeCode` only; `InsertCodes` and `DeleteCodes` are removed (no other caller).

**Adapter** (`adapters/postgres/repository.go`). One pattern for every transition: `Begin` → `SELECT totp_secret, totp_enabled, totp_last_counter FROM users WHERE id = $1 AND deleted_at IS NULL AND disabled_at IS NULL FOR UPDATE` → no row: `ErrUserNotFound` → decide → write → `Commit`. Any error before commit rolls back. The decision for confirm, in this order:

1. `totp_enabled` → `AlreadyEnabled` (409);
2. stored ciphertext ≠ `Version` (including NULL) → `SecretChanged` (422);
3. `totp_last_counter >= Counter` → `StepUsed` (422);
4. otherwise `UPDATE users SET totp_enabled = true, totp_last_counter = Counter` + insert the batch → `Enabled`.

The comparison is on the stored ciphertext as text, done in SQL (`totp_secret = $version`) or in Go on the locked read; both read the same row under the lock.

**Application** (`two_factor.go`).

- `EnrollTwoFactor`: `false` → `ErrTwoFactorAlreadyEnabled`.
- `ConfirmTwoFactor`: read state; `Enabled` → 409 (fast path, unchanged); verify the code with `domain.VerifyTOTPAt` against `state.Secret` (no counter write); wrong code → 422; generate codes (failure → 503, nothing written); `ConfirmTotp` with `state.Version`; map `AlreadyEnabled` → `ErrTwoFactorAlreadyEnabled`, `SecretChanged` and `StepUsed` → `ErrInvalidTOTP`. SA13-B is preserved: the winning step is recorded inside the enabling transaction.
- `authorizeFactor` returns the version it proved against. `DisableTwoFactor` passes it to `DisableTotp`; `false` → `ErrTwoFactorNotEnabled` (409, already declared for disable). `RegenerateRecoveryCodes` passes it to `ReplaceRecoveryCodes`; `false` → `ErrTwoFactorNotEnabled` (409, declared). `AdminDisableTwoFactor` calls `ResetTotp`.
- Login (`CompleteTwoFactor`) and `authorizeFactor`'s own step consumption are unchanged (N-b).

**Error precedence (review R1 B3).** The response is decided by the row read under the lock in the writing transaction, never by a second read. Because every writer takes the same lock, a losing confirmation waits for the winner's commit and then reads `enabled = true`: it answers 409 for the same step and for distinct steps alike. A confirmation that reaches the lock after a reset reads a NULL secret: 422. A missing, deleted or blocked user under the lock: 404 (N-e). A failure to begin, read or commit: 503 (`persistenceError`), nothing persisted except in N-d.

**Why the ciphertext is a valid version.** AES-GCM with a fresh random nonce per encryption (`identitycrypto/keys.go:72-81`) makes every enrolment's stored value different, even for an equal secret; the only encryption call is `SetTotpSecret` (`repository.go:349-350`); there is no re-encryption job (SA13-I is forward-only).

**Alternatives considered.**

- Conditional single statements without a transaction (rev 1): close R1 and R2 but not R4 or R5 — the counter, the flag and the code batch are three writes (review R1 B1/B2).
- A transaction port through the application (`WithTx`): the same guarantee, but leaks transaction handling into the application layer and every fake; the adapter already owns such multi-row transitions (`administration.go:17-60`, `repository.go:137-175`).
- `SERIALIZABLE` without row locks: needs retry loops on serialization failure in every caller; the row lock serialises the only contended row.
- A separate version column: needs a migration; the ciphertext already changes on every write.

## Decisions

| # | Decision | Date, author | Named check |
|---|---|---|---|
| D1 | All factor transitions take the user's row lock in one adapter transaction (not conditional statements, not a transaction port) | 2026-11-28, Courses 1 (author inference, to independent review) | A3–A6 integration tests; reverting any transition to the rev-1 split writes turns at least one red |
| D2 | A confirmation for a replaced or cleared secret answers `422 invalid_totp`, matching Л5's described experience («у вкладці з перезаписаним секретом підтвердження дає «Код автентифікатора неправильний»»); the losing confirmation of two answers 409 | 2026-11-28, Courses 1 | A2 and A3 assert the exact error values |
| D3 | The response is decided by the locked read in the writing transaction; there is no post-failure re-read | 2026-11-28, Courses 1 | A3 same-step case: with the decision moved before the lock, the loser answers 422 and the test fails |
| D4 | Locked reads add `disabled_at IS NULL`: a blocked account's factor writes answer 404 like a deleted account's | 2026-11-28, Courses 1 | integration: block the user, then confirm → `ErrUserNotFound`, state unchanged |
| D5 | Login proof staleness (N-b) and the undeclared 404 (N-e) stay out of scope, recorded with their mechanisms | 2026-11-28, Courses 1 | this record; reopening needs a new task row |

## Proportional grilling

- *Does the row lock deadlock with other transactions?* Every SA14 transition locks exactly one `users` row and then writes only `recovery_codes` rows of that user. Session rotation locks the same `users` row first (`repository.go:137-175`) and does not touch `recovery_codes`. Lock order is always `users` → `recovery_codes`, so no cycle. Checked against the existing locking writers listed in «Scoped impact».
- *Can the version comparison refuse spuriously?* Only if the row was rewritten between the read and the lock, which is the race. No re-encryption path exists.
- *Does dropping the counter write from confirm's pre-check reopen SA13-B?* No: the step is written in the same transaction that enables, and `StepUsed` refuses a step that is not newer. The existing SA13-B integration and e2e checks (the enrolment code cannot be the first login) stay in A7.
- *Is the `StepUsed` branch reachable?* While pending, the counter is written only by a winning confirm (which also enables) and reset to 0 by disable/reset, so in practice it is reached only by a clock-skewed duplicate. It is kept so the transaction never moves the counter backwards; a unit test covers its mapping.
- Owner decisions: only «Fix 2FA». D1–D5 are the author's and go to the independent review.

## Scoped impact

- Callers of the changed port methods: `two_factor.go` only (`grep -rn 'InsertCodes\|DeleteCodes\|\.DisableTotp\|\.EnableTotp\|SetTotpSecret' backend/internal backend/cmd`, excluding tests and `generated/`). Wiring: `backend/cmd/api/main.go:159` (`NewTwoFactor`; the same `*Repository` serves both stores).
- `GetTwoFactor` callers (`two_factor.go:59,85,151,259`, `identity.go:81`) read `Secret`/`Enabled`; `Version` is additive.
- SQL: `backend/sql/queries/identity.sql` (locked read, enable with counter, disable-and-delete, replace codes) → `make generate` (sqlc) → `adapters/postgres/generated`; `check-generated` gates drift. `EnableUserTotp` and `DisableUserTotp` change shape or are replaced.
- Tests: `application/two_factor_test.go` (fake store gains the new methods with the same semantics), `adapters/postgres/two_factor_integration_test.go` (A1–A6; it may build the application service over the real repository, as `schema_integration_test.go:25` already imports `application`), e2e `frontend/e2e/accounts.spec.ts:513-531` (A7 wait). The e2e edit lands after PR #124 is decided, on whatever `main` then is, to avoid a conflict in the same file.
- No migration, no OpenAPI change, no frontend source change.

## Rollback

No migration and no wire change: the previous binary reads and writes the same row shapes, so a revert is a code revert. A revert reopens R1–R5; recovery codes issued under SA14 stay valid under the old code. No data repair is needed in either direction.

## Verification plan

- `go test ./internal/identity/...`; `make integration` in full (persisted auth state; the integration tag is skipped by `go test ./...`); ordered `make check` before merge (shared persisted state → full e2e).
- **Mutation check (review R1 B4).** Before the code review, revert each guard locally in turn and record which test turns red: the `totp_enabled` refusal in `SetTotpSecret` (A1), the version comparison (A2, A4a), the enabled check under the lock (A3), the counter write inside the transaction (A2), the code delete inside reset (A4b), the rollback on a failed insert (A5), the version check in regenerate (A6). A guard whose revert leaves every test green is a missing test, not a passing one. Results go to this record.
- Independent Codex review of this plan before code and of the code before merge (auth, SA01/SA08 precedent).

## Planning review

`review_mode=independent_subagent` (Codex CLI); writer Courses 1; reviewer Codex.

- R1 (2026-11-28, on `688caf4b`): **BLOCKED**, four blocking findings — B1 the counter advanced by a rejected confirm (→ R4, A2, D1), B2 non-atomic issuance and reset (→ R5, A4, A5, I1–I3), B3 error precedence via a re-read (→ D3, «Error precedence»), B4 proofs that pass on broken orchestration (→ A3 interleaving, mutation check). Non-blocking N1 (R3 corrected), N2 (V5a on the ACCT ref), N3 (residuals N-a…N-d, D4), N4 (A7 evidence and wait), N5 (decisions table). Working file `docs/security/_wip-review-sa14-plan.md` (local).
- R2: pending.

## Readiness

BLOCKED on the independent plan review R2.
