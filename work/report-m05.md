# Report M05 — media repository writes (PostgreSQL)

## Outcome

M05 implements the first media PostgreSQL adapter at READ COMMITTED through `pool.Begin`.
The adapter owns `Mint`, `Load`, `AppendChunk`, `Complete`, two-result `Claim`, `Cancel`,
`ListCourseVideos`, and `Retry`. It does not implement the M06 publish, eviction, recovery,
or sweep slice.

## Changed

- `backend/internal/media/adapters/postgres/repository.go` adds the repository and explicit
  database-to-domain mapping. `Mint` locks the owning course first with
  `FOR SHARE OF courses`, runs `CanEditLesson` over the locked owner, takes over only expired
  `created`/`uploading` rows, assigns `MAX(generation)+1` across all generations, and maps
  both `video_assets_pending_per_lesson` and `video_assets_generation` `23505` violations to
  `upload_in_progress` by constraint name.
- `AppendChunk` and `Complete` invoke their supplied policies inside `pool.Begin`, after the
  row lock and method guard and before the guarded update. A policy error rolls the database
  state back. `Claim` returns the re-read current asset with `claimed=false` when its guarded
  update loses. `Retry` locks the asset, resolves ownership through lesson→module→course,
  delegates the T9 state set to `domain.ValidateTransition`, requires a landed source, and
  admits only a stale processing claim.
- `backend/sql/queries/media.sql` adds sqlc-owned active-asset load and course-video list
  reads. The course list is a single joined read that returns the owner even when the course
  has no assets, so ownership is decided inside this repository.
- `backend/sqlc.yaml` adds the media adapter generation target. No file was written beneath
  a `generated/` directory; the reviewer must run sqlc and regenerate that target.
- `backend/internal/media/adapters/postgres/repository_integration_test.go` adds the
  Testcontainers specification for P06, B16, B22, R28, and R38, plus one lifecycle check for
  policy rollback, append, complete, two-result claim, stale retry, and cancel. R28 uses
  controlled query/row locks to prove both ownership outcomes and deterministic overlap with
  a 4a authoring save. R38 runs concurrent mints with the opposite unique index temporarily
  removed in each isolated case, forcing both named `23505` paths before restoring the index.

## Read-through verification

- Every media write in this slice calls `pool.Begin`; the new files contain no `BeginTx` or
  `RepeatableRead` and do not copy the course module's read isolation.
- The mint lock order is course first, then stale asset takeover/generation/insert. Its actor
  policy runs after the course lock and before any media write.
- The append and complete order is lock → existence/state/offset-or-length guard → policy →
  guarded update → commit. Their policies receive only `domain.VideoAsset`.
- The retry SQL repeats the landed-source, allowed-state, claim-age, and active-row guards
  after the locked checks; it clears the claim and failure reason and increments `attempt`.
- Active reads hide evicted rows. Generation assignment includes evicted rows. Only expired
  upload-window states are evicted during mint; `queued` and `processing` remain blockers.
- The new adapter/query set contains no publish, third-generation eviction, requeue/fail,
  recovery, or sweep implementation and adds no filesystem, ffmpeg, queue, HTTP, tus, or
  frontend surface.
- The required two-axis read-only review was adapted to the brief's Git denial and run over
  the explicit four-file change set. Standards: zero findings after retry-state delegation
  and shared stored-row scanning. Specification: zero findings after deterministic R28 and
  index-specific R38 coverage.

## Verification performed

- `gofmt -w` ran on the two new Go files: exit 0.
- `gofmt -d` ran on the same files: exit 0 with no output.
- Direct text checks found all eight required repository methods, `pool.Begin` on every
  write, both named unique constraints, and no forbidden isolation or M06 method.
- No filesystem write was denied, so there is no denied-write error text and the mandatory
  stop condition did not fire.
- No Git command was run. No commit was created because the brief explicitly denies Git,
  overriding the generic implementation workflow's commit step.
- `work/tasks.md` and `docs/media/report.md` were not edited. No file inside any
  `generated/` directory was written.

## Checks not run — denied by the brief

None of the following is represented as passed:

- `sqlc` / media generated-code regeneration — not run; sqlc is denied and writing generated
  output is forbidden in this session.
- Go compilation — not run; explicitly denied. The new adapter intentionally awaits reviewer
  regeneration of `internal/media/adapters/postgres/generated`.
- `make format-check` — not run; its Go half uses Docker. Local `gofmt` is the formatting
  evidence.
- `make lint` — not run; its Go-vet half compiles through Docker.
- `make typecheck` — not run; Go compilation is denied.
- `make test` — not run; Go compilation and test execution are denied.
- `make integration` — not run; Docker and Testcontainers are denied, including P06, B16,
  B22, R28, R38, and the lifecycle integration check.
- `make check-generated` — not run; it invokes Git and writes generated output, both denied.
- `make build` — not run; it requires Docker and Go compilation.
- `make up`, `make migrate`, `make seed`, `make seed-identities`, and `make smoke` — not run;
  they require Docker Compose and are outside this repository-only slice.
- `make e2e` — not run; it requires Docker Compose and M05 owns no browser surface.
- Aggregate `make check` — not run because it includes the denied checks above.

## Limitation and next task

The adapter, generated-query contract, and integration tests have formatting and read-through
evidence only; compilation and runtime success are deliberately not claimed. The next planned
slice is M06: publish/eviction/recovery/sweep.

## Reviewer verification (measured in the reviewing session)

One review fix: the test harness's `reset` sent five SQL commands in one parameterised
`Exec`, which pgx sends over the extended protocol — every subtest failed with
`cannot insert multiple commands into a prepared statement (42601)`. Split into
single-command statements, parameters preserved. Codex could not run Testcontainers and
claimed only formatting evidence, so this surfaced exactly where it should.

After the fix: `make integration` exit 0 across all fifteen packages —
`courses/internal/media/adapters/postgres` green in 4.3s with P06, B16, B22, R28, R38 and
the lifecycle subtest passing. `make generate` clean, `check-generated` green, format-check,
lint and typecheck pass. Verified by inspection: `pool.Begin` everywhere (zero
`RepeatableRead`, the documented copy trap), `FOR SHARE OF courses` taken first in the mint
transaction, and `23505` mapped by `ConstraintName` across both
`video_assets_pending_per_lesson` and `video_assets_generation`.

Environment note: two integration runs failed on Testcontainers startup
(`wait for reaper ... unexpected container status`, and `unknown message type: h` on
`host.docker.internal`) with 23 unrelated containers from other projects running. Both were
infrastructure flakiness, confirmed by a clean re-run — recorded so a later reader does not
mistake it for a code defect. Provenance: implemented with Codex (gpt-5.6-sol); reviewed,
corrected and verified here.
