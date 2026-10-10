# A03 report — PostgreSQL authoring adapter and sqlc queries

## Changed

- `backend/internal/course/adapters/postgres/authoring.go` — implemented the six-method
  `AuthoringRepository`: owner-scoped, repeatable-read list; consistent tree load;
  transactional create with an eligible-owner row lock and `23505` slug mapping; full-tree
  save with scoped node upserts, delete-missing, final optimistic course update and rollback;
  and single-statement guarded transition/delete with locked in-transaction
  disambiguation.
- `backend/sql/queries/courses.sql` — added sqlc-owned list/count, course/module/lesson reads,
  locked course/owner reads and draft creation. `sqlc.yaml` already lists the file, so no new
  target was needed. Existing public projections remain independent of Phase 4 columns so the
  catalog still works after migration 00003 is rolled back.
- `backend/internal/course/adapters/postgres/repository.go` — changed the existing catalog
  mapper to accept projected scalar fields instead of the generated table model. This lets
  sqlc emit post-00003 custom row types for the nine-column public projections without making
  those queries depend on `instructor_id`; public domain behavior is unchanged.
- `backend/internal/course/adapters/postgres/authoring_integration_test.go` — added one
  Testcontainers harness with named evidence for P08, P09, P10, B03, B10, B11, R16, R17,
  R18, R19 and R26, plus delete-missing, Create/Save owner ineligibility and instructor
  reassignment coverage.

## Decisions within A03's freedom

- sqlc owns the simple bounded queries. The aggregate write stays hand-managed so its module
  and lesson `RowsAffected` checks, child ordering and final token update share one explicit
  transaction. Transition and delete likewise remain hand-written because their guarded
  statement and refusal disambiguation must be visibly adjacent.
- Migration 00003 adds a column to sqlc's `Course` table model, so the pre-existing nine-column
  catalog queries will generate query-specific row types. Their mapper now consumes the common
  scalar projection instead of adding `instructor_id` to SQL: B15 requires those queries to
  keep executing after the authoring migration is rolled back and that column no longer exists.
- Save locks the course before touching children, locks the proposed owner with `FOR SHARE`,
  upserts every module and lesson, deletes absent lessons before absent modules, then executes
  the `id + expectedUpdatedAt` course update last. A stale token therefore rolls back all
  staged child changes, including deletes.
- Module conflicts update only when the stored module belongs to the target course. Lesson
  conflicts update only when the lesson's stored parent belongs to that course; changing the
  target `module_id` therefore supports an intra-course move while refusing a foreign lesson.
  Any zero affected-row upsert returns `invalid_request` without disclosing the foreign owner.
- Create and Save hold an eligible owner row lock (`role IN ('instructor', 'admin')` and
  `disabled_at IS NULL`) until commit, so role/disable changes cannot race ownership writes.
  Create also enforces the actor's assignment policy at the repository boundary.
- All adapter-supplied write timestamps are UTC and truncated to PostgreSQL's microsecond
  precision. Save and transition advance by one microsecond when truncation would leave the
  optimistic token equal to or earlier than the current token. Publish uses the new clock only
  when `published_at` is null; archive and restore preserve the first publication time.
- Plan row B03 says the slug already exists while also requiring one of two concurrent creates
  to succeed. Those conditions cannot both hold. The test uses an initially free slug, launches
  two creates together, and proves exactly one success, one `slug_taken`, and one stored row —
  the stated concurrency property and observable result.
- The brief named
  `backend/internal/course/adapters/postgres/schema_integration_test.go` in its read-first list,
  but that path does not exist. The only matching existing harness,
  `backend/internal/identity/adapters/postgres/schema_integration_test.go`, was read after the
  course repository integration harness; no missing file was invented.

## Verified

- `gofmt -w backend/internal/course/adapters/postgres/repository.go
  backend/internal/course/adapters/postgres/authoring.go
  backend/internal/course/adapters/postgres/authoring_integration_test.go` — exit 0; the local
  formatter parsed and formatted all edited Go files.
- `gofmt -d` on those files — exit 0 with no output.
- Static acceptance scan — all eleven owned row labels are present as named subtests.
- SQL/code read-through against plan D-3b, D-3c, D-3e, D-3f, D-3g, D-4b and D-5 — confirmed
  the final guarded course update, scoped `ON CONFLICT` predicates, affected-row refusals,
  upsert-before-delete order, in-transaction locked disambiguation, `courses_slug_key`
  `23505` mapping and owner eligibility lock.
- Two-axis read-only review against repository standards and the A03 brief found no documented
  standards violation or remaining confirmed spec defect. Its initial findings led to explicit
  Create actor enforcement, checked test setup loads, locked zero-row disambiguation, wider
  race-test timing, sub-microsecond timestamp evidence and Save-side owner eligibility evidence.
  A final invariant pass also added a strict monotonic-token guard and positive instructor
  owner-defaulting evidence.
- No file inside a `generated/` directory was written. No filesystem write was denied.

## Not run

- `sqlc` and `make generate` — not run: the brief's exact environment denial is **No Docker**
  and explicitly says sqlc cannot run. Generated symbols are intentionally left for the
  reviewing session; no generated file was hand-written.
- Go compilation, targeted `go test`, and `go test -run TestAuthoringRepositoryAgainstPostgreSQL`
  — not run: the brief explicitly forbids compiling or running Go in this environment.
- `make format-check` — not run: its Go formatter executes through Docker. The permitted local
  `gofmt` checks were run instead.
- `make lint` — not run: its Go vet half executes through Docker and compiles Go.
- `make typecheck` — not run: the brief forbids Go compilation.
- `make test` — not run: the brief forbids compiling or running Go. No unit or race result is
  represented as passing.
- `make integration` — not run: Docker/Testcontainers are unavailable. P08–P10, B03, B10,
  B11 and R16–R19/R26 are written as integration evidence but are not represented as passing.
- `make check-generated` — not run: sqlc cannot run, and its drift recipe uses git, which the
  user explicitly prohibited.
- `make build`, `make up`, `make seed-identities`, `make seed`, `make smoke`, `make e2e` and the
  aggregate `make check` — not run: Docker is unavailable and these exceed A03's permitted
  environment.
- No git command was run. No commit was created.

## Limitation and next task

The existing generated package is deliberately stale until the Docker-enabled reviewing
session runs `make generate`; that session must correct or return any sqlc naming mismatch,
then run the ordered gates through `make integration`. A04 remains the next planned slice only
after A03's generated code, compilation and Testcontainers evidence are measured.

## Reviewer verification (measured in the reviewing session)

`make generate` exit 0 — the hand-written expectations of sqlc output matched the real
generation on the first run; only `generated/courses.sql.go` and `generated/models.go`
changed, as declared. `make test` pass; `make integration` pass — the new
`authoring_integration_test.go` suite green inside `course/adapters/postgres` (5.8s),
identity suites unaffected. `make format-check` and `make lint` pass. `make
check-generated` red before commit (diff against HEAD includes the pending generated
output — the gate working as designed) and green after this commit. Provenance:
implemented with Codex (gpt-5.6-sol); reviewed and verified here.
