# SA14 — 2FA enrolment/confirmation race

| | |
|---|---|
| Date | 2026-11-28 |
| Source | `docs/design-pass/plan.md` §18, limitation Л5 («Серверна гонка … окрема задача»). Owner: «Fix 2FA» (2026-11-28) |
| Root | `main` @ `58622b51`; branch `fix/sa14-2fa-race` from `main`; one writer (Courses 1). PR #124 (design-pass ACCT, frontend only) is open separately and touches no backend file |
| Status | see «Readiness» |

## Problem

Each 2FA write reads the state, decides, then writes without re-checking the state it read.

- **R1 (Л5).** `EnrollTwoFactor` (`backend/internal/identity/application/two_factor.go:58-82`) reads `Enabled == false` and then calls `SetTotpSecret`. `SetUserTotpSecret` (`backend/sql/queries/identity.sql:13-19`) sets the new secret and `totp_enabled = false` with no condition. Suppose tab B confirms 2FA between tab A's read and A's write. A's enrolment then silently turns 2FA off and replaces the secret. B's recovery codes stay in the table. The account is left unprotected while its owner believes 2FA is on.
- **R2 (found while reading R1).** `ConfirmTwoFactor` (`two_factor.go:84-111`) verifies the code against the secret it read (S1) and then calls `EnableTotp`. `EnableUserTotp` (`identity.sql:21-27`) enables whatever secret is stored now. Suppose tab B enrols between A's read and A's enable, storing S2. A's confirmation then enables 2FA with S2, which A never saw. The owner is locked out: no code from their authenticator matches.
- **R3 (consequence).** Two confirmations that both pass verification (two different steps of the same secret) both reach `EnableTotp` and `InsertCodes`. Each returns a different set of recovery codes, but only the last set stays valid in practice.

## Scope and acceptance

| # | Observable behaviour | Proof |
|---|---|---|
| A1 | Enrolment never disables 2FA that is already on. If 2FA is enabled when the write lands, the request answers `409 twofactor_already_enabled` and the stored secret, `totp_enabled` and the recovery codes are unchanged | integration (PostgreSQL): enable, then enrol through the repository guard → refused, state unchanged; application unit: store reports the guard failed → `ErrTwoFactorAlreadyEnabled` |
| A2 | Confirmation enables 2FA only for the secret its code was checked against. If the secret changed in between, the request answers `422 invalid_totp` (the code does not belong to the current secret) and 2FA stays off. If 2FA was enabled in between, it answers `409 twofactor_already_enabled` | integration: read state, replace the secret, enable with the old version → refused; enable with the current version → succeeds. Unit: both branches map to their errors |
| A3 | Of two confirmations racing on the same secret, exactly one enables 2FA and returns recovery codes; the other answers `409` and inserts no codes | integration: two enables with the same version → one true, one false. Unit: second confirm → `ErrTwoFactorAlreadyEnabled`, `InsertCodes` not called |
| A4 | Everything else is unchanged: the normal enrol → confirm → login → disable → re-enrol journey, admin disable, regeneration | existing unit tests, `two_factor_integration_test.go`, e2e `accounts.spec.ts` Acc26–Acc31 |

Non-goals: the frontend (it already maps 409 to «Стан 2FA міг змінитися…» with a reload, and 401 to an error in the form; ACCT V5a); `DisableTwoFactor` and `RegenerateRecoveryCodes` races (they act on a proven second factor; no finding); the TOTP step counter (SA13-B, unchanged); the HTTP contract (no new code or status).

## Design

Make both writes conditional on the state the decision was made on. That is optimistic concurrency in SQL, with no locks and no new transaction.

- `SetUserTotpSecret` adds `AND totp_enabled = false`. Port: `SetTotpSecret(ctx, userID, secret) (bool, error)`; `false` = the account has 2FA enabled now. The adapter tells «no row» apart from «guard failed» with one extra read, done only when the update touches 0 rows: missing user → `ErrUserNotFound`, as today.
- `ports.TwoFactor` gains `Version string`: the stored ciphertext as read, opaque to the application, empty when no secret. `EnableUserTotp` adds `AND totp_secret = $version AND totp_enabled = false`. Port: `EnableTotp(ctx, userID, version) (bool, error)`.
- Application: `EnrollTwoFactor` maps `false` to `ErrTwoFactorAlreadyEnabled`. `ConfirmTwoFactor` passes `state.Version`. On `false` it reads the state once more: enabled → `ErrTwoFactorAlreadyEnabled`; otherwise → `ErrInvalidTOTP`. Recovery codes are inserted only after a `true`.
- The ciphertext is a valid version: AES-GCM with a fresh nonce (SA13-I) makes every enrolment's stored value different, even for an equal secret.

Alternatives considered: `SELECT … FOR UPDATE` in one transaction (needs a transaction port through the application; larger for the same guarantee); a separate version column (needs a migration; the ciphertext already changes on every write).

## Proportional grilling

- *Can the ciphertext comparison fail spuriously?* Only if the row was rewritten between the read and the write, which is exactly the race. Key rotation re-encrypts? Not implemented (SA13-I is forward-only, no rewrite job), so no false refusals. Checked: `grep -rn EncryptSecret backend/internal` shows calls only in `SetTotpSecret`.
- *Does the extra read on 0 rows leak anything?* It runs after an authenticated request for the caller's own account; it returns only the existing error codes.
- *Mapping a changed secret to `invalid_totp`.* Author's decision, matching Л5's described user experience («у вкладці з перезаписаним секретом підтвердження дає «Код автентифікатора неправильний»»). The alternative 409 would also be acceptable to the frontend; `invalid_totp` keeps the form and «Скасувати», which is the right next step for a stale enrolment.
- Owner decisions: only «Fix 2FA». Everything else here is the author's inference and goes to the independent review.

## Scoped impact

- Callers of the two port methods: `two_factor.go` only (`grep -rn "SetTotpSecret\|EnableTotp" backend/internal`). `GetTwoFactor` callers (`two_factor.go:59,85,151,259`, `identity.go:81`) only read `Secret`/`Enabled`; the new field is additive.
- SQL: `backend/sql/queries/identity.sql` → `make generate` (sqlc) → `backend/internal/identity/adapters/postgres/generated`. `check-generated` gates drift.
- Tests: `application/two_factor_test.go` (fake store), `adapters/postgres/two_factor_integration_test.go`, e2e `frontend/e2e/accounts.spec.ts` (journey unchanged).
- No migration, no OpenAPI change, no frontend change.

## Verification plan

`go test ./internal/identity/...`; `make integration` in full (persisted auth state; memory: integration tag is skipped by `go test ./...`); ordered `make check` before merge (shared persisted state → full e2e). Independent Codex review of the plan before code and of the code before merge (auth, SA01/SA08 precedent).

## Planning review

`review_mode=independent_subagent` (Codex CLI); writer Courses 1; reviewer Codex. Plan review: pending.

## Readiness

BLOCKED on the independent plan review (auth, persisted state).
