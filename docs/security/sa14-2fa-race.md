# SA14 — 2FA enrolment/confirmation race

| | |
|---|---|
| Date | 2026-11-28 (rev 3 the same day, after plan reviews R1 and R2, both BLOCKED) |
| Source | `docs/design-pass/plan.md` §18, limitation Л5 («Серверна гонка … окрема задача»). Owner: «Fix 2FA» (2026-11-28) |
| Root | `main` @ `dc2575b2`; branch `fix/sa14-2fa-race` from `main`; one writer (Courses 1). PR #124 (design-pass ACCT) is open separately; it touches no backend file but does touch `frontend/e2e/accounts.spec.ts` (see A8) |
| Status | see «Readiness» |

## Problem

Each 2FA write reads the state, decides, then writes without re-checking the state it read, and the writes that belong together commit separately.

- **R1 (Л5).** `EnrollTwoFactor` (`backend/internal/identity/application/two_factor.go:58-82`) reads `Enabled == false` and then calls `SetTotpSecret`. `SetUserTotpSecret` (`backend/sql/queries/identity.sql:13-19`) sets the new secret and `totp_enabled = false` with no condition. If tab B confirms 2FA between tab A's read and A's write, A's enrolment silently turns 2FA off and replaces the secret, while B's recovery codes stay in the table.
- **R2.** `ConfirmTwoFactor` (`two_factor.go:84-111`) verifies the code against the secret it read (S1) and then calls `EnableTotp`. `EnableUserTotp` (`identity.sql:21-27`) enables whatever secret is stored now. If tab B enrols S2 in between, A enables S2, which A never saw: the owner is locked out.
- **R3.** Two confirmations that both pass verification (two steps of the same secret) both reach `EnableTotp` and `InsertCodes`. Inserts append (`repository.go:398-420`; no generation column, `migrations/00014_accounts.sql:27-33`), so both returned sets stay valid until used or deleted.
- **R4.** Every proof consumes its TOTP step with an update keyed only on the user and a newer counter (`identity.sql:38-47`, called from `two_factor.go:94,160,268` via `verifyAndConsumeTOTP` at `290-296`). A proof made against S1 and delayed past a reset and a re-enrolment of S2 consumes step N on the S2 row, so a valid S2 code at N is then refused. A confirmation that is about to fail, or a proof for a disable or regeneration that then fails, still burns the step (review R1 B1, R2 B1). Recovery codes are consumed the same way, outside any check of the factor (`identity.sql:53-59`, `two_factor.go:166,275`).
- **R5.** Enable and code issuance commit separately (`two_factor.go:101-111`); so do disable and code deletion (`221-224`, `248-253`) and regeneration's delete and insert (`231-240`). A confirmation delayed past an admin reset inserts codes for an account with 2FA off, and login completion consults recovery codes even with the factor off (`two_factor.go:166-178`). A failure after enable leaves 2FA on with no codes delivered (review R1 B2).
- **R6 (existing data).** Rows can already violate the invariant below: R5's split writes may have left codes while 2FA is off, and the identity seed resets the factor but leaves the counter and the code rows (`backend/sql/seed-identities.sql:14-27`) (review R2 B2).

## Invariant

For every committed state of a user row, after SA14:

- **I1.** A recovery code is *consumable* only while `totp_enabled = true`, and every consumable code was issued by the confirmation or regeneration of the currently stored secret. Rows that exist while 2FA is off are inert and are deleted by the next confirmation.
- **I2.** `totp_enabled` becomes true only for the secret whose code was verified, together with exactly the code batch returned to that caller and the step that caller consumed.
- **I3.** For enrol, confirm, self-disable, regenerate and admin reset: an operation that is refused or fails before commit changes nothing — secret, flag, counter, recovery codes and the proof it presented stay as they were. Login completion is outside I3 (residual N-b).
- **I4.** A proof made against one secret never changes the state of another: a TOTP step is consumed only on the row whose stored secret is the one it was verified against, while that factor is enabled.

**How existing data enters the invariant (R6).** No migration is needed: I1 is stated over *consumable* codes, and SA14 makes consumption require `totp_enabled = true` (I4 predicate below), so leftover rows while off cannot be used; the next confirmation deletes every code of the user before inserting its batch, so leftover rows cannot become consumable either. The seed additionally resets `totp_last_counter = 0` and deletes the seeded users' recovery codes, so a reseed starts clean. There is no rollout overlap: one binary is replaced; there is no production deployment (`AGENTS.md`, Scope) and no second writer process.

## Scope and acceptance

Every row's proof runs against PostgreSQL through the real repository unless marked *unit*. «Unchanged» always means: the stored ciphertext, `totp_enabled`, `totp_last_counter`, and the exact set of recovery-code hashes with their `consumed_at`, read through a separate connection.

| # | Observable behaviour | Proof |
|---|---|---|
| A1 | Enrolment never disables 2FA that is already on: if 2FA is enabled under the lock, the request answers `409 twofactor_already_enabled`, state unchanged | enable with a known batch, then `SetTotpSecret` → refused, unchanged. *Unit*: the fake's first read sees pending and the write refuses → `ErrTwoFactorAlreadyEnabled` |
| A2 | Confirmation enables 2FA only for the secret its code was checked against. If the stored secret changed or was cleared and 2FA is still off under the lock, it answers `422 invalid_totp` and the counter is unchanged, so a valid code of the current secret at the same step then succeeds. If the replacement was already confirmed, it answers `409` (enabled is checked first) | application service with a barrier wrapper: confirm A reads S1 and is held; enrol S2 commits; A proceeds with a step-N code → `ErrInvalidTOTP`, unchanged; then confirm S2 at N → enabled, counter = N, one batch. Variant: S2 confirmed before A proceeds → `ErrTwoFactorAlreadyEnabled`, S2's state unchanged |
| A3 | Of two confirmations that both read the pending state, exactly one enables 2FA and returns codes; the other answers `409`, inserts nothing, consumes nothing. Holds for the same step and for distinct steps, in both commit orders | application service over the real repository, fixed clock, a barrier wrapper that returns both `GetTwoFactor` reads before either write and then releases the writers in a forced order; bounded context. Cases: same step N; steps N−1 and N with N−1 first; N first. Assert one success and one `ErrTwoFactorAlreadyEnabled`; the stored hashes equal the winner's 10 codes; counter = winner's step |
| A4 | Reset is atomic: after an admin reset or a self-disable commits, 2FA is off and no code is consumable, on either side of a confirmation; a failure inside the reset changes nothing | (a) confirm reads, reset commits, confirm proceeds → `ErrInvalidTOTP`, off, 0 codes; (b) reset after a committed confirm → off, 0 codes, counter 0; (c) a test-only `BEFORE DELETE` trigger on `recovery_codes` raising for the test user → admin reset and self-disable return an error, state unchanged (flag still on, codes intact) |
| A5 | A failed confirmation write leaves nothing behind and can be retried | `newRepository` with a `random` reader that fails on the 3rd UUID (after two inserts) → `ErrDependencyUnavailable` (mapped to 503 at `handler.go:855`); unchanged through a second connection; then a normal repository confirms with a fresh code |
| A6 | Self-disable and regeneration act only on the factor their proof was verified against, and consume that proof only if they commit | reset, enrol S2, confirm S2 (batch C2); then self-disable and regenerate with S1's version and an S1 proof → `409 twofactor_not_enabled`; S2's state and C2 unchanged, S2's counter unchanged. Regeneration with a correct proof and a failing 3rd UUID → error, the proof's step or code not consumed, old batch intact |
| A7 | A stale proof cannot consume a step or a code on a replacement factor (I4) | consume an S1 step via the login path after reset + enrol S2 → refused, S2 counter unchanged; a leftover recovery row while 2FA is off → login consumption refused |
| A8 | Everything else is unchanged: enrol → confirm → login → disable → re-enrol, admin disable, regeneration; the SA13-B rule that the enrolment code cannot be replayed as the first login | existing unit tests, `two_factor_integration_test.go`, e2e `accounts.spec.ts` Acc26–Acc29, Acc31; the admin-reset journey (`accounts.spec.ts:513-531`) waits for the recovery codes to show before the reset (review R1 N4) |

## Design

**Lock discipline.** Every transition below runs in one adapter transaction: `Begin` → `SELECT totp_secret, totp_enabled, totp_last_counter FROM users WHERE id = $1 AND deleted_at IS NULL FOR UPDATE` → no row: `ErrUserNotFound` → decide on the locked row → write → `Commit`. Any error before commit rolls back. At READ COMMITTED a second transition of the same user waits on the row lock and then reads the committed row (review R2 N1: no lock cycle with session rotation or administration, which take the same `users` row first and do not write `recovery_codes`).

**Ports** (`backend/internal/identity/ports/two_factor.go`). The adapter owns every multi-row transition; the application keeps TOTP verification and code generation, which are pure.

- `TwoFactor` gains `Version string`: the stored ciphertext as read, opaque to the application, `""` when no secret; forwarded only, never logged or returned.
- `FactorProof{Counter int64; RecoveryCode string}`: exactly one is set. The application sets `Counter` when `domain.VerifyTOTPAt` matched against the read secret, otherwise `RecoveryCode` with the presented code; the adapter hashes it.
- `SetTotpSecret(ctx, userID, secret) (bool, error)` — refuses when enabled.
- `ConfirmTotp(ctx, userID, version, counter, codes) (ConfirmOutcome, error)` — outcome `Enabled | AlreadyEnabled | SecretChanged | StepUsed`.
- `DisableTotp(ctx, userID, version, proof) (FactorOutcome, error)` and `ReplaceRecoveryCodes(ctx, userID, version, proof, codes) (FactorOutcome, error)` — outcome `Done | FactorChanged | ProofInvalid`.
- `ResetTotp(ctx, userID) error` — admin, unconditional, no proof.
- `AdvanceTotpCounter(ctx, userID, version, counter) (bool, error)` and `RecoveryCodeStore.ConsumeCode` — login only, single statements with the I4 predicates.
- `RecoveryCodeStore.InsertCodes` and `DeleteCodes` are removed (no other caller).

**Decisions inside the locked transactions.**

- *Confirm:* enabled → `AlreadyEnabled`; ciphertext ≠ version (incl. NULL) → `SecretChanged`; `last_counter >= counter` → `StepUsed`; else delete all of the user's codes, `UPDATE users SET totp_enabled = true, totp_last_counter = $counter` (one statement), insert the batch → `Enabled`.
- *Self-disable / regenerate:* not enabled or ciphertext ≠ version → `FactorChanged`; proof is a counter and `last_counter >= counter` → `ProofInvalid`; proof is a recovery code and the consuming `UPDATE recovery_codes … WHERE consumed_at IS NULL` touches 0 rows → `ProofInvalid`; else record the counter or keep the consumption, then clear the secret, flag and counter and delete all codes (disable), or delete all codes and insert the new batch (regenerate) → `Done`.
- *Admin reset:* clear the secret, flag and counter and delete all codes. No `disabled_at` filter: an administrator can reset a blocked account, as today (`admin_handler.go:145-167`).
- *Enrol:* enabled → refuse; else store the new ciphertext with `totp_enabled = false`.
- *Login (unchanged flow, narrowed statements):* `AdvanceUserTotpCounter` adds `AND totp_enabled AND totp_secret = $version`; `ConsumeRecoveryCode` adds `AND EXISTS (SELECT 1 FROM users u WHERE u.id = recovery_codes.user_id AND u.totp_enabled AND u.deleted_at IS NULL)`.

**Application** (`two_factor.go`). Confirm: read; enabled → 409 (fast path); `VerifyTOTPAt` against `state.Secret`, no write; no match → 422; generate codes (failure → 503, nothing written); `ConfirmTotp(state.Version, …)`; `AlreadyEnabled` → `ErrTwoFactorAlreadyEnabled`, `SecretChanged` / `StepUsed` → `ErrInvalidTOTP`. `authorizeFactor` becomes a pure step: it reads the state, returns `ErrTwoFactorNotEnabled` when off, and builds the `FactorProof`; it writes nothing. Disable and regenerate pass version and proof; `FactorChanged` → `ErrTwoFactorNotEnabled` (409), `ProofInvalid` → `ErrRecoveryCodeInvalid` (401), matching today's mapping for a failed proof (`two_factor.go:268-281`, `handler.go:849,853`). Regeneration generates its codes before the transaction. Login passes `state.Version` to `AdvanceTotpCounter`.

**Error precedence.** The response is decided by the row read under the lock in the writing transaction; there is no second read. A losing confirmation waits for the winner's commit and reads `enabled = true`: 409 for the same and for distinct steps. A confirmation that reaches the lock after a reset reads NULL: 422. Errors raised before the lock (malformed body 400, wrong code 422, rate limit 429, unauthenticated 401) keep their existing precedence.

**Wire contract.** SA14 adds no newly reachable status: 409 `twofactor_not_enabled` is already declared for disable and regeneration (`openapi_contract.go:113-127`); 404 for a deleted user and 503 for a persistence failure are already reachable on all four self-service operations through the existing mapping (`handler.go:823,855`). Their absence from the declarations is pre-existing (N-e). D4 of rev 2 (a `disabled_at` filter that would have made 404 newly reachable for a blocked user and changed admin reset of a blocked target) is withdrawn — decision D4.

**Why the ciphertext is a valid version.** AES-GCM with a fresh random nonce per encryption (`identitycrypto/keys.go:72-81`) makes every enrolment's stored value different, even for an equal secret; the only encryption call is `SetTotpSecret` (`repository.go:349-350`); there is no re-encryption job (SA13-I is forward-only).

**Alternatives considered.** Conditional single statements without a transaction (rev 1): close R1 and R2, not R4 or R5. A transaction port through the application: the same guarantee, but transaction handling leaks into the application and every fake; the adapter already owns such transitions (`administration.go:17-60`, `repository.go:137-175`). `SERIALIZABLE` without row locks: retry loops in every caller. A version column: a migration for what the ciphertext already provides. A cleanup migration for R6: unnecessary once consumption requires an enabled factor and confirmation replaces the batch.

## Residuals (not closed by SA14; each with its mechanism)

- **N-a. Factor writes after an account block.** Self-service transitions match `deleted_at` but not `disabled_at`, so a request already past session resolution can write after a block commits. New sessions and rotation are refused (`repository.go:131-175`). To close: add `disabled_at IS NULL` to the self-service locked read together with a declared 404 or 403.
- **N-b. Login completion.** `CompleteTwoFactor` consumes its proof and then rotates the session (`two_factor.go:151-185`). A rotation failure after the consumption burns one step or code (outside I3), and a reset + re-enrol between the proof and the rotation leaves a session issued on the old factor's proof. To close: carry `Version` in `PendingTwoFactor` and do consume + rotate in one locked transaction.
- **N-c. Enrolment after an admin reset.** A paused enrolment can store a new pending secret after a reset commits. Reset is linearised at its commit; anything after it is a new enrolment, which cannot turn 2FA on without a confirmation by the same user.
- **N-d. Ambiguous commit.** The database can commit while the client loses the acknowledgement, so a returned 503 can coexist with committed state, and the codes were never shown. A retry answers 409; recovery is a proven regeneration (`POST /auth/2fa/recovery-codes`). Injected failures before commit (A4c, A5, A6) are definite rollbacks. A binary revert cannot undo committed state.
- **N-e. Undeclared statuses.** The enrol, confirm, disable and regenerate declarations omit 404 and (except enrol) 503, and confirm omits its malformed-body 400 (`openapi_contract.go:89-127` against `handler.go:215-233,704-715,823,855`). Pre-existing; a separate contract change regenerates the client.
- **N-f. Concurrent reseed.** The seed locks `users` in `SHARE ROW EXCLUSIVE` (`seed-identities.sql:2-4`); run concurrently with a factor transaction holding a row lock, the two can deadlock (review R2 N2). Operational boundary: the seed runs only before a test run, never during one (`make seed-identities`, the gate's setup).

## Decisions

| # | Decision | Date, author | Named check |
|---|---|---|---|
| D1 | All factor transitions take the user's row lock in one adapter transaction; proof consumption for disable and regeneration moves into that transaction | 2026-11-28, Courses 1 (author inference, to independent review) | A3–A6 and the mutation check below |
| D2 | A confirmation for a replaced or cleared secret answers `422 invalid_totp` while 2FA is off (Л5: «підтвердження дає «Код автентифікатора неправильний»»); the loser of two confirmations answers 409 | 2026-11-28, Courses 1 | A2 and A3 assert the exact error values |
| D3 | The response is decided by the locked read in the writing transaction; no post-failure re-read | 2026-11-28, Courses 1 | A3 same-step case: with the decision moved before the lock, the loser answers 422 and the test fails |
| D4 | No `disabled_at` filter (rev 2's D4 withdrawn); the blocked-account window stays residual N-a | 2026-11-28, Courses 1 | integration: admin reset of a blocked account still succeeds (existing behaviour preserved) |
| D5 | No migration for existing rows (R6): consumption requires an enabled factor, confirmation replaces the batch, the seed resets the counter and codes | 2026-11-28, Courses 1 | A7 second case; the seed change is checked by `make seed-identities` then a query of the seeded users' code rows = 0 |
| D6 | Login completion (N-b) and the undeclared statuses (N-e) stay out of scope | 2026-11-28, Courses 1 | each residual names the change that would close it; reopening needs a new task row |

## Proportional grilling

- *Deadlock?* Every transition locks one `users` row, then writes only that user's `recovery_codes`. Rotation and administration lock the same row first and do not write `recovery_codes`. Lock order is always `users` → dependants. The one exception is the seed's table lock (N-f).
- *Spurious version refusals?* Only if the row was rewritten between the read and the lock — the race. No re-encryption path.
- *Does moving the counter into confirm's transaction reopen SA13-B?* No: the step is written by the same `UPDATE` that enables, and `StepUsed` refuses a step that is not newer. Mutation linkage: dropping that column from the `UPDATE` lets the enrolment step be replayed as the first login, which the existing SA13-B assertion catches.
- *Is `StepUsed` reachable?* While pending, the counter is written only by a winning confirm (which also enables) and reset to 0 by disable/reset, so only by a clock-skewed duplicate. Kept so the transaction never moves the counter backwards; *unit*-tested mapping.
- Owner decisions: only «Fix 2FA». D1–D6 are the author's and go to the independent review.

## Scoped impact

- Callers of the changed port methods: `two_factor.go` only (`grep -rn 'InsertCodes\|DeleteCodes\|\.DisableTotp\|\.EnableTotp\|SetTotpSecret\|AdvanceTotpCounter\|ConsumeCode' backend/internal backend/cmd`, excluding tests and `generated/`). Wiring: `backend/cmd/api/main.go:159` (`NewTwoFactor`; one `*Repository` serves both stores).
- `GetTwoFactor` callers (`two_factor.go:59,85,151,259`, `identity.go:81`) read `Secret`/`Enabled`; `Version` is additive.
- SQL: `backend/sql/queries/identity.sql` (locked read, enable with counter, the clear-and-delete, the code replace, the two narrowed login statements) → `make generate` (sqlc) → `adapters/postgres/generated`; `check-generated` gates drift. `backend/sql/seed-identities.sql` (counter and codes reset).
- Tests: `application/two_factor_test.go` (fake store with the same outcomes); `adapters/postgres/two_factor_integration_test.go` (A1–A7; it builds the application service over the real repository, allowed by `AGENTS.md` Architecture and already done in `schema_integration_test.go:23-25,92-101`); e2e `frontend/e2e/accounts.spec.ts:513-531` (A8 wait), landed after PR #124 is decided, on whatever `main` then is.
- No migration, no OpenAPI change, no frontend source change.

## Rollback

No migration and no wire change: the previous binary reads and writes the same row shapes, so a revert is a code revert. It reopens R1–R6; codes issued under SA14 stay valid under the old code. A revert does not undo committed state (N-d).

## Verification plan

- `go test ./internal/identity/...`; `make integration` in full; ordered `make check` before merge (shared persisted state → full e2e).
- **Mutation check (review R1 B4, R2 B4).** Before the code review, revert each guard locally in turn, run the identity integration package, and record in this file which test turns red: enrol's enabled refusal (A1); confirm's version comparison (A2); confirm's enabled check under the lock (A3); confirm's delete-before-insert (A7 second case, via a leftover row turned consumable); the counter in confirm's `UPDATE` (SA13-B in A8); the code delete in reset and self-disable, by splitting them into two transactions (A4c); the rollback on a failed insert (A5); the version predicate alone in self-disable and in regenerate (A6, against an *enabled* replacement); proof consumption moved before the transaction (A6, failing regeneration); the I4 predicates (A7). A guard whose revert leaves every test green is a missing test. Moving a write to after `Commit` within the same request has no observable window in a test without a commit hook; it is checked by reading the adapter (every write uses the transaction's `queries`, none uses `repository.pool` after `Begin`).
- Independent Codex review of this plan before code and of the code before merge (auth, SA01/SA08 precedent).

## Planning review

`review_mode=independent_subagent` (Codex CLI); writer Courses 1; reviewer Codex.

- R1 (on `688caf4b`): **BLOCKED** — B1 counter advanced by a rejected confirm, B2 non-atomic issuance and reset, B3 error precedence via a re-read, B4 proofs pass on broken orchestration; N1–N5. Working file `docs/security/_wip-review-sa14-plan.md` (local).
- R2 (on `dc2575b2`): **BLOCKED** — B1 proof writers outside the lock break I3 and taint the replacement counter (→ R4, I4, A6, A7, D1); B2 no starting-state guarantee for I1 (→ R6, D5, seed); B3 wire contract and admin target policy unresolved (→ D4 withdrawn, «Wire contract», N-a, N-e); B4 surviving mutants (→ A4c trigger, A6 enabled replacement, A3 forced orders, the mutation list). N1 (A2 409 variant), N2 (N-f), N3 (N-d), N4 (task row). R1 B1–B3 closed for confirmation; B4 not. Working file `docs/security/_wip-review-sa14-plan-r2.md` (local).
- R3: pending.

## Readiness

BLOCKED on the independent plan review R3.
