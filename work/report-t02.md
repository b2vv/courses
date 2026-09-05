# T02 report — identity administration domain

## Files

- `backend/internal/identity/domain/user.go` — adds `DeletedAt`, mirrors both
  migration timestamp bounds and deleted-implies-disabled in `NewUser`, and makes
  `Active()` require both status timestamps to be nil.
- `backend/internal/identity/domain/administration.go` — adds the
  active/blocked/deleted `AccountStatus` projection, the identity-local `Actor`,
  `CanBlock`/`CanDelete`, and the 1–1024-byte password helper.
- `backend/internal/identity/domain/errors.go` — adds the four Phase 4b business
  errors: `cannot_target_self`, `admin_immutable`, `last_admin`, and
  `user_not_found`.
- `backend/internal/identity/domain/user_test.go` — covers valid soft deletion,
  both deleted timestamp bounds, and deleted-implies-disabled.
- `backend/internal/identity/domain/administration_test.go` — covers status and
  `Active()` projection, every administration policy rule and precedence, the
  zero-versus-one other-admin boundary, idempotent blocking of an already-blocked
  admin, and password byte boundaries.
- `backend/internal/identity/domain/session_test.go` — pins the four new error
  strings beside the existing identity error vocabulary.
- `work/report-t02.md` — this report.

No file under a `generated/` directory was written.

## Decisions

- The migration is the code truth: `DeletedAt` must be between `CreatedAt` and
  `UpdatedAt`, inclusively, and requires `DisabledAt`. This is stricter than the
  plan table's abbreviated lower-bound wording and exactly mirrors migration
  `00004`.
- `AccountStatus()` checks `DeletedAt` first, so a valid deleted-and-disabled row
  projects as `deleted`, never `blocked`.
- The identity-local `Actor` contains only `UserID`. The handler owns the admin-role
  gate; these pure policies need only caller identity and locked target facts.
- `CanDelete` checks self-target before admin immutability, preserving the required
  `cannot_target_self` result for an admin deleting themselves.
- `CanBlock` applies `last_admin` only to an active admin. An already-blocked admin
  is a same-state no-op and must not produce a spurious refusal.
- `ValidPassword` uses Go byte length, accepting any 1–1024 bytes including
  whitespace and rejecting 0 or more than 1024 bytes, matching login and B04.

## Verified

- Ran local `gofmt` on all six touched Go files: exit 0.
- Ran `gofmt -d` on all six touched Go files: exit 0 with no diff.
- Read through the implementation and tests against plan §2, §4 D-2/D-3, §8,
  NB2-9, `user.go`, and migration `00004`.
- Read through all non-generated backend `User` composite literals: they use keyed
  fields, so adding `DeletedAt` does not break positional construction.
- Read through production imports under `backend/internal/identity/domain`: stdlib
  only; no identity-outside import was added.

## Checks not run

The brief denies Docker and all Go compilation/testing and permits only local
`gofmt` plus read-through. None of the ordered gates is represented as passed:

- `make format-check` — not run; its Go half requires Docker.
- `make lint` — not run; its Go vet half requires Docker/Go compilation.
- `make typecheck` — not run under the brief's environment limit.
- `make test` — not run; Go unit tests require Docker/Go compilation. The T02 unit
  tests remain written evidence awaiting the reviewing environment.
- `make integration` — not run; Docker/Testcontainers are denied and T02 adds no
  integration surface.
- `make check-generated` — not run; it invokes Git, which this task forbids, and
  T02 does not modify generator inputs or generated output.
- `make build` — not run; Docker/compilation are denied.
- `make up`, `make seed-identities`, `make seed`, `make smoke`, `make e2e` — not
  run; they require the Compose environment and are outside this pure-domain task.
- `make check` — not run because it includes every denied step above.
- Formal `code-review` skill — not run; it requires Git fixed-point resolution,
  log, and diff commands, while the brief forbids touching Git. No workaround was
  attempted.

No write was denied, so the brief's exact-text stop condition did not trigger.

## Reviewer verification (measured in the reviewing session)

`make format-check`, `make lint`, `make typecheck`, `make test` (identity domain suite
green with the new CanBlock/CanDelete/AccountStatus/DeletedAt tables), `make integration`
pass. `Active()` now checks both DisabledAt and DeletedAt; the three refusal codes exist.
Provenance: implemented with Codex (gpt-5.6-sol); reviewed and verified here.
