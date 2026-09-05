# A01 report — migration 00003, gate order, seed ownership

## Changed

- `backend/migrations/00003_authoring.sql` — added the nullable ownership column, deterministic eligible-owner backfill, `NOT NULL`, restricted owner foreign key, ordered owner index, module/lesson tables with database bounds and cascade ownership, ordinary ordered indexes, and the destructive test-only down migration. The header documents the exact clean-volume recovery for a courses-without-users failure.
- `backend/sql/seed.sql` — assigned all five demo courses to the instructor fixture, added the table lock, and changed the fixed-id conflict path to update every seeded field.
- `backend/internal/course/adapters/postgres/repository_integration_test.go` — made P01 and B13–B15 observable in the existing Testcontainers harness; seeded identities before the concurrent course seed; supplied `instructor_id` to the two valid direct inserts; added real migration 00003 to the failing-migration MapFS and moved its canary to 00004.
- `Makefile`, `.github/workflows/ci.yml`, `README.md`, `.claude/standards.md` — reordered identity seed before course seed. `make seed` now depends on `seed-identities`.

## Verified

- `make -n check` — exit 0. Its dry-run output orders `build` → `up` → `seed-identities` → `seed` → `smoke` → `e2e`. Recipes were printed only; none ran. The printed `check-generated` recipe contains existing git commands, but no git process was executed.
- `gofmt -w backend/internal/course/adapters/postgres/repository_integration_test.go` — exit 0, confirming the edited Go test parses and is formatted.
- Targeted `rg`/shell assertions — exit 0: exactly five seed rows use instructor `40000000-0000-4000-8000-000000000002`; the lock and fixed-id update clause exist; the old `DO NOTHING` clause is absent; the migration contains the decided backfill, `NOT NULL`, ordered indexes and no `UNIQUE`/`DEFERRABLE`; the integration test contains real 00003, synthetic 00004 and SQLSTATE `23502` coverage.
- SQL read-through against `00001_courses.sql` and `00002_identity.sql` — UUID primary-key checks, non-blank bounded titles, bounded lesson descriptions, non-negative positions, timestamp ordering checks, `RESTRICT`/`CASCADE` ownership and down order match the repository style and plan §6.

## Not run

- `make format-check` — sandbox: no Docker (`GO_RUN` Go formatting half).
- `make lint` — sandbox: no Docker (`GO_RUN` Go vet half).
- `make test` — sandbox: no Docker.
- `make integration` — sandbox: no Docker.
- `make check-generated` — sandbox: no Docker (`GO_RUN` generation); its git drift checks were also not run because the task forbids touching git.
- `make build`, `make up`, `make seed-identities`, `make seed`, `make smoke`, `make e2e` — sandbox: no Docker.
- `make check` (real execution) — sandbox: no Docker. Only `make -n check` was run.
- Any Go compilation through `GO_RUN` — sandbox: no Docker.
- `git diff --stat` — not run because the user explicitly required never touching git.

No filesystem write was denied. No git command was run and no commit was created.

## Limitation and next task

The migration evidence reuses one sequential Testcontainers database. That is the smallest harness that covers A01; isolate databases before making these migration subtests parallel. The reviewing session must run the full ordered gate before A01 is committed. Next planned slice: A02.

## Reviewer verification (measured in the reviewing session)

Ordered gates re-run with Docker available: `make format-check` pass (frontend 84 files +
gofmt clean), `make lint` pass (oxlint, 48-file architecture check, `go vet`),
`make typecheck` pass, `make test` pass (frontend suite + all Go packages with race),
`make integration` pass — `courses/internal/course/adapters/postgres` ok in 5.6s including
the new backfill/rollback/ownerless subtests; identity suite unaffected. `make -n check`
prints `seed-identities` before `seed`. Later gates (`build`→`e2e`) belong to A15 per the
plan. Provenance: implemented with Codex (gpt-5.6-sol); reviewed and verified here.
