# T04 report — PostgreSQL identity administration adapter

## Files

- `backend/internal/identity/adapters/postgres/administration.go` — implements the
  `AdministrationRepository` with escaped/paged directory reads and READ COMMITTED
  transactions opened only through `pool.Begin`; status, delete, and reset run with
  the required user/admin locks and update-before-session-revoke order.
- `backend/sql/queries/identity.sql` — adds `ListAdminUsers`, its total query, the
  unlocked target read, `LockAdmins`, `LockUserForStatus`, the single nullable
  `SetUserDisabledAt`, soft delete, password update, and per-user session revoke.
- `backend/internal/identity/adapters/postgres/administration_integration_test.go` —
  adds Testcontainers evidence for P03–P06, B03, B05, B11, B12, R09, R10, R12,
  R16, R18, and R22, plus directory escaping/order/paging and mapper coverage. This
  harness intentionally has no migration `DownTo`, so P05's reused-email pair never
  reaches the destructive 00004 Down.
- `backend/internal/course/adapters/postgres/authoring_integration_test.go` — adds
  R23 through the real identity administration adapter and existing authoring
  repository: blocking the owner makes an otherwise valid admin save return
  `invalid_owner` without changing the course.
- `backend/internal/identity/adapters/argon2id/verifier.go` and
  `verifier_test.go` — add the consumer-side `Hash` operation with a random 16-byte
  salt and round-trip/prefix evidence.
- `backend/internal/identity/adapters/postgres/repository.go` — renames the existing
  consumer-side interface to `Passwords` and adds `Hash`; no new port was created.
- `backend/internal/identity/adapters/postgres/schema_integration_test.go` — updates
  the existing password fake to satisfy the widened consumer-side interface.
- `work/report-t04.md` — this report.

No domain, application, HTTP/OpenAPI, frontend, migration, or generated file was
changed. No file under a `generated/` directory was written.

## Transactional decisions

- Every adapter operation uses `pool.Begin`, preserving PostgreSQL READ COMMITTED.
  No `BeginTx` or `RepeatableRead` appears in the new adapter.
- Admin block first reads the immutable role, then `LockAdmins` locks every
  non-deleted admin in deterministic id order, including the target. The active
  count excludes the target and is skipped for a same-state block. A waiting race
  loser therefore re-reads the committed admin set and returns `last_admin`.
- Non-admin block and every restore use `LockUserForStatus`. Delete uses the same
  row lock before its policy. Each policy callback runs after the required lock and
  before a write.
- Block writes `disabled_at` and `updated_at` through the one nullable-parameter
  query, then revokes sessions. Restore passes SQL NULL and does not revoke.
  Same-state transitions commit without changing timestamps or sessions.
- Delete sets `deleted_at`, `disabled_at = COALESCE(disabled_at, deleted_at)`, and
  `updated_at` together, then revokes sessions. A concurrent/double delete sees the
  row disappear from the locking query and returns `user_not_found`.
- Reset hashes before opening the transaction; after that it runs the row-locking
  password UPDATE before session revocation. Hash failure maps to
  `dependency_unavailable` with no database write.
- The directory duplicates the existing one-line `strings.NewReplacer` escape
  idiom and SQL `ESCAPE '!'`, maps both status timestamps through `domain.NewUser`,
  orders by email/id, and uses a separate count query so a page past the end still
  returns its total.

## Written evidence

- P03/P04/B11/B12: block removes a live session and bars credential verification;
  repeated block preserves timestamps and a sentinel session; restore permits
  credentials and sessions again.
- P05/R12: delete retains a disabled+deleted row, revokes its session, permits an
  active row to reuse its email, rejects the deleted user's old password, and makes
  status/delete/reset return `user_not_found`.
- P06/B05: reset stores the required Argon2id encoding, the new password verifies,
  the old password fails, and the prior session is gone; reset also succeeds for a
  blocked user without clearing the block.
- B03/R09/R10: ordinary admin block succeeds; a constructed single-active-admin
  state refuses; two concurrent admin blocks produce one success and one
  `last_admin`, leaving exactly one active admin.
- R16/R18: a fake hashing failure leaves the credential unchanged; cancellation
  while the user row is locked leaves both account status and session intact.
- R22: lock-queued login rotation against both block and delete leaves no live
  session; concurrent delete produces one success and one `user_not_found`.
- R23: blocking an instructor through this adapter freezes their existing course at
  the authoring seam.

## Verification performed

- Ran `gofmt -w` on all touched Go files, then `gofmt -d`; the final diff was empty.
- Read through the SQL argument names/types against existing sqlc symbol conventions
  and the course adapter's escaped-LIKE idiom.
- Read through every transaction against `Rotate`: `pool.Begin`, user-row lock before
  session operations, rollback on every early return, commit last.
- Searched the new identity adapter/tests for `BeginTx` and `RepeatableRead`: no
  matches. Searched the adapter for transaction starts: all four are `pool.Begin`.
- Read through the implementation and evidence against every T04-owned acceptance
  row and the Phase 4b NOT-list. Direct Standards review found no repository-rule
  breach; direct Spec review found no missing owned behavior or scope creep.

## Denials and checks not run

The brief explicitly denies Docker, sqlc, Go compilation, and Testcontainers. It
also forbids every Git operation. None was run or worked around, and no denied check
is represented as passed:

- `make format-check` — not run; its Go half requires denied Docker/compilation.
- `make lint` — not run; its Go vet half requires denied Docker/compilation.
- `make typecheck` — not run; Go compilation is denied.
- `make test` — not run; Go compilation is denied. The Argon2id test is written
  evidence awaiting the reviewer environment.
- `make integration` — not run; Docker and Testcontainers are denied. All T04
  acceptance tests remain written evidence.
- `make generate` / sqlc — not run; sqlc is denied and generated-directory writes
  belong to the reviewer.
- `make check-generated` — not run; it requires sqlc-generated output and invokes
  Git, both forbidden here.
- `make build`, `make up`, `make seed-identities`, `make seed`, `make smoke`,
  `make e2e`, and aggregate `make check` — not run; they require denied compilation
  or Docker/Compose and include work outside T04.
- Formal `code-review` skill — not run because it requires Git fixed-point,
  log, and diff commands plus a Git-resolved comparison. The brief denies Git; a
  permitted direct standards/spec read-through was used, not a workaround.
- Commit — not created because the brief explicitly says never touch Git, overriding
  the implementation skill's usual commit step.

No write was denied, so the brief's exact-text stop condition did not trigger. No Git
command was run.

## Limitation and next task

The repository's sqlc output must be regenerated and all written tests compiled/run in
the Docker-enabled reviewing session. T05 is next: HTTP/OpenAPI routes, admin gate,
DTO bounds, error mapping, and contract evidence over this adapter.

## Reviewer verification (measured in the reviewing session)

`make generate` (LockAdmins/LockUserForStatus/DeleteIdentitySessionsForUser/set-password
+ argon2id Hash); all four admin transactions use pool.Begin (READ COMMITTED), no
RepeatableRead. `make test` (argon2id Hash + fake-verifier suites green), `make
integration` 12 packages ok — the administration suite covers every owned row: block
revokes sessions + bars login (P03/R22), restore (P04), soft delete + email reuse
terminal (P05/R12), reset revokes sessions (P06/B05), R09 last admin refused
sequentially, R10 concurrent blocks leave one admin active (409 not 503), R16 hasher
failure writes nothing, R18 cancellation rollback, R22 login-vs-block/delete races, and
R23 (blocked owner freezes their course) asserted in the authoring suite. format-check,
lint, typecheck, check-generated pass. Provenance: implemented with Codex (gpt-5.6-sol);
reviewed and verified here.
