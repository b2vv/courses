# T03 report — identity administration application use cases

## Files

- `backend/internal/identity/ports/administration.go` — adds the four-method
  `AdministrationRepository`: filtered/paged list, transactional status change,
  transactional soft delete, and transactional password set. Status/delete accept
  the domain-policy callback that the future PostgreSQL adapter must run over its
  locked facts.
- `backend/internal/identity/application/administration.go` — adds `Administration`
  with `ListUsers`, `SetStatus`, `Delete`, and `ResetPassword`, all using the identity
  module's three-second operation budget and UTC microsecond clock.
- `backend/internal/identity/application/administration_test.go` — adds fake-port
  coverage for all four use cases, pre-repository self refusal, policy evaluation over
  repository facts, password bounds and consumer-side hash routing, bounded deadlines,
  R15 dependency failure, and R17 caller cancellation.
- `work/report-t03.md` — this report.

No file under a `generated/` directory was written.

## Decisions

- The port has exactly the four operations T03 calls. It adds no load/count/session or
  hasher method.
- Status and delete receive narrow policy callbacks. This lets the application consume
  `CanBlock`/`CanDelete` while T04 can invoke them after its locked read and before its
  write in the same transaction. Self-block and self-delete are decidable from the
  caller and target ids, so they return before the repository.
- `SetStatus` represents only active/blocked. `ResetPassword` applies the existing
  `ValidPassword` bound, then passes plaintext to `SetPassword`; hashing remains in the
  PostgreSQL adapter's existing consumer-side password dependency.
- Repository users are revalidated with `domain.NewUser`; malformed stored rows and a
  negative total become `dependency_unavailable`. Explicit repository business and
  dependency errors otherwise pass through unchanged.
- Caller cancellation takes precedence over a repository error. The internal bounded
  context still reaches every repository method.

## Verified

- `gofmt -w` on the three changed Go files — exit 0.
- `gofmt -d` on the three changed Go files — exit 0 with no output.
- Static surface read-through — `AdministrationRepository` has exactly `List`,
  `SetStatus`, `Delete`, and `SetPassword`; the T03 files contain no `Hash`/hasher seam.
- Architecture read-through — application imports only stdlib plus identity domain and
  ports; ports imports only stdlib plus identity domain. Domain, adapters, HTTP,
  OpenAPI, PostgreSQL/sqlc, frontend, and generated files are unchanged.
- Spec read-through against brief T03, plan §2/§3/§4 D-2, R15/R17, and the existing
  identity application idiom found no remaining T03 gap.

## Checks not run

The brief explicitly denies Docker and therefore denies Go compilation/testing. No
denial was worked around, and no unrun check is represented as passed:

- `make format-check` — not run; its Go half requires Docker. Local `gofmt` was the
  permitted substitute only for formatting/parsing.
- `make lint` — not run; Go vet requires Docker/Go compilation.
- `make typecheck` — not run; the brief forbids Go compilation.
- `make test` — not run; the fake-port tests require Go compilation. R15 and R17 are
  written evidence awaiting the reviewing environment, not measured passes.
- `make integration` — not run; Docker/Testcontainers are denied and T03 adds no
  integration adapter.
- `make check-generated` — not run; its recipe touches Git, which the brief forbids,
  and T03 changes no generator input or output.
- `make build`, `make up`, `make seed-identities`, `make seed`, `make smoke`, `make e2e`,
  and aggregate `make check` — not run; they require denied compilation/Compose or lie
  outside this fake-port slice.
- Formal `code-review` skill — not run: it requires resolving a Git fixed point,
  inspecting Git log/diff, and parallel reviewers over that diff. The brief forbids
  touching Git, so no substitute Git command was attempted. A direct standards/spec
  read-through of the three known files was performed instead.

No Git command was run. No filesystem write was denied, so the brief's exact-text stop
condition did not trigger.

## Limitation and next task

Compilation and execution of the fake tests remain for a Docker-enabled reviewing
session. T04 is next: implement the PostgreSQL/sqlc adapter behind this port, including
READ COMMITTED locking, same-transaction session revocation, consumer-side password
hashing, and the specified race/rollback evidence.

## Reviewer verification (measured in the reviewing session)

format-check, lint, typecheck pass; `make test` identity application suite green
including R15 (dependency failure) and R17 (cancellation) on fakes; `make integration`
12 packages ok. The port passes the domain policy as a callback the adapter runs inside
its locked transaction — the correct seam for the last-admin race (T04). Provenance:
implemented with Codex (gpt-5.6-sol); reviewed and verified here.
