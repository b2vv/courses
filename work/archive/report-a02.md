# A02 report — authoring domain, port and application use cases

## Changed

- `backend/internal/course/domain/authoring.go` — added the course-side `Actor` and mirrored
  role vocabulary, authoring course/tree types, the exact D-5 business errors used by this
  slice, ownership and role policy, the lifecycle matrix, draft-only deletion, pre-publication
  slug mutation, authoring field bounds, contiguous positions, aggregate counts and one
  course-wide node-id uniqueness check.
- `backend/internal/course/domain/authoring_test.go` — table tests for admin/instructor/student
  policy, instructor-owned publish/archive/restore, reassignment refusal, every legal and
  representative illegal transition, delete and slug rules, field/count/position/text bounds,
  nil ids and duplicates within and across node kinds.
- `backend/internal/course/ports/repository.go` — added `AuthoringRepository` beside the
  unchanged `CourseRepository`, with only the six methods called by A02's use cases.
- `backend/internal/course/application/authoring.go` — added bounded `List`, `Load`, `Create`,
  `Save`, `Transition` and `Delete` use cases. Locally decidable validation runs before a port
  call; ownership, lifecycle and slug rules that need the stored course run after `Load`.
- `backend/internal/course/application/authoring_test.go` — fake-port evidence for all six use
  cases, pre-port refusals, stored-course policies, deadline propagation, timestamp/token
  forwarding, R24 dependency degradation, R25 cancellation and R39 zero-call duplicate-id
  refusal.

## Decisions within A02's freedom

- `CourseDraft` is the create input, so create cannot accept lifecycle or node members and the
  repository owns server-minted course identity and initial `draft` state. Create makes one
  repository call and has no retry/idempotency mechanism; A03 maps the database slug collision
  that guards replay.
- `CourseTree` carries internal zero-based positions even though transport DTOs will not. A04/A05
  derive them from array indexes; domain validation then ensures every adapter sees contiguous
  positions.
- Save validates only editable course fields and nodes before its first port call, then overlays
  them on the loaded aggregate. The stored status, publication and creation timestamps therefore
  survive even though D-3a excludes those members from the PUT body.
- Module and lesson ids share one validation set. Reusing an id within either kind or across the
  two kinds returns `invalid_request` before `Load`, which is R39's invalid-tree outcome without
  inventing a wire error outside D-5.
- List returns a narrow `AuthoringCourse` row rather than loading whole trees. The existing
  normalized query supplies the 1–50 limit and offset rules; authoring adds the optional status
  filter and fixes ordering to recent.
- Owner UUID shape is checked in the domain. Whether an admin-selected account exists, is active
  and has an eligible role remains an A03 transactional repository check, because only that
  dependency has the data needed to decide it without a race.
- The application passes `expectedUpdatedAt` to save, transition and delete unchanged. New write
  timestamps are UTC and truncated to microseconds; if a clock reading does not exceed the
  expected token, it advances by one microsecond so an accepted write cannot retain its old
  token.
- If the caller's context is cancelled, the use case returns that cancellation/deadline instead
  of relabelling it as a dependency or business failure. An internal three-second budget keeps
  the repository's explicit classification, matching the catalog idiom; other repository
  business and dependency errors also pass through unchanged.

## Verified

- `gofmt -w` on all five changed Go files — exit 0. A local formatter was available; this also
  parsed every edited Go file.
- `gofmt -d` on the same files — exit 0 with no output.
- Static import check — domain uses only `strings`, `time` and `unicode/utf8`; application uses
  only `context`, `time`, course domain and course ports; no identity or external-package import
  exists.
- Static surface check — the original two-method `CourseRepository` is unchanged;
  `AuthoringRepository` contains exactly `List`, `Load`, `Create`, `Save`, `Transition` and
  `Delete`.
- Careful read-through against plan D-3a/D-3b/D-3c/D-3e/D-3f/D-3g, D-4a/D-4b and D-5, migration
  00003, and acceptance rows R24/R25/R39.

## Not run

- `make format-check` — not run: its Go half requires Docker, which the brief says is unavailable.
  The permitted local `gofmt` check was run instead.
- `make lint` — not run: Go vet runs through Docker, and the brief forbids compiling or running Go.
- `make typecheck` — not run: the brief forbids compiling or running Go in this environment.
- `make test` — not run: the brief explicitly says Go cannot be compiled or run here. Therefore
  R24, R25 and R39 are implemented as fake tests but are not represented as passing.
- `make integration` — not run: Docker is unavailable; A02 also adds no integration surface.
- `make check-generated` — not run: A02 changes no generated inputs, Docker is unavailable, and
  its drift recipe touches Git, which the user explicitly prohibited.
- `make build`, `make up`, `make seed-identities`, `make seed`, `make smoke`, `make e2e` and the
  aggregate `make check` — not run: Docker is unavailable and these exceed A02's fakes-only
  environment.
- No Git command was run. No commit was created.

No filesystem write was denied.

## Limitation and next task

Compilation and the fake tests require the reviewing session's Docker-enabled Go gate. A03 is
the next planned slice: implement the PostgreSQL/sqlc adapter behind this port and verify the
transactional concurrency, node-membership and rollback rules against isolated PostgreSQL.

## Reviewer verification (measured in the reviewing session)

`make format-check` pass; `make lint` pass (go vet + 48-file architecture check — no
identity import from course, boundaries suite green); `make test` pass with race —
course domain and application suites including the new authoring tables run green;
`make integration` pass (course postgres suite 4.7s, unaffected as expected);
`make typecheck` pass (no TS surface changed). R24/R25/R39 fake-port evidence now
measured, not just written. Provenance: implemented with Codex (gpt-5.6-sol); reviewed
and verified here.
