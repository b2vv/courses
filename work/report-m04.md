# Report M04 — media application use cases over fakes

## Outcome

M04 is implemented over the M03a domain and port surface. The first M04 attempt stopped
without application code because six required contracts were missing. Plan §2.1 and M03a
repaired that stop with the three transaction policy callbacks, `Load`, explicit claim
disposition, `DeleteSwept`, `MediaStorage.Head`, `ports.Job`, the container rule and
`TranscodeOutcome`.

## Changed

- `backend/internal/media/application/upload.go` adds `MintTicket`, `AppendChunk`,
  `CompleteUpload`, `CancelUpload`, `ListCourseVideos` and `Retry`. Calls are bounded,
  timestamps are UTC/microsecond-normalized, mint validates before the repository, the first
  chunk gates free space and sniffs its container inside `ChunkPolicy`, completion moves the
  source inside `CompletePolicy`, and completion/retry commit before enqueue.
- `backend/internal/media/application/transcode.go` extends the existing outcome with the
  `Transcode` use cases: the exhaustive load/claim dispositions, a fresh claim token, disk
  gate, probe and codec refusal, ladder/render/thumbnail, guarded publication, retry
  exhaustion, dependency pass-through, cause-sensitive cancellation, `Sweep` and `Recover`.
  Canceled work gets a bounded detached finalization context so cleanup and the guarded
  requeue/fail can run after the render context is canceled.
- `backend/internal/media/application/upload_test.go` and `transcode_test.go` add fake-only
  coverage with recorded call counts for T1–T9 application paths, invalid tickets before any
  repository call, first/later chunks, disk and container refusals, commit-before-enqueue,
  duplicate enqueue success at the queue-port boundary, enqueue-after-commit recovery,
  exhaustive no-ops, retry exhaustion, permanent malformed sources, lost claims, both
  cancellation causes, disk exhaustion, dependency faults, publish failure, sweep ordering,
  C5 and C6 recovery, and an inert duplicate delivery.

No adapter, SQL, HTTP, tus server, frontend, dependency, or generated file was added or
changed. `work/tasks.md` and `docs/media/report.md` were not edited.

## Read-through verification

- The six gaps in the stopped report are now represented directly by the exact §2.1 port
  signatures; no additional port or domain contract was needed.
- `Process` follows D-10's order: sweep; exhaustive load; explicit claim disposition; free
  space; probe/refusal; isolated work keys; render and thumbnail; claim-guarded publish;
  cleanup; requeue or fail before returning a retryable outcome.
- `ready`, `failed`, evicted, absent, orphaned, `created` and `uploading` assets return the
  zero outcome before claim/transcode/publish. An unclaimable `processing` row returns a
  retryable error rather than being consumed.
- Plain `context.Canceled` requeues even on the final attempt. A final-attempt
  `context.DeadlineExceeded` writes `render_failed`; non-final cancellation/deadline paths
  requeue first.
- Storage exhaustion maps only to the closed `storage_unavailable` failure reason; raw
  paths/stderr are never placed in an outcome. Repository, storage and queue dependency
  errors pass through rather than being relabelled as business refusals.
- Sweep purges source, HLS, thumbnail and temporary key families before `DeleteSwept`.
  Recovery enqueues only the rows returned by the repository, using the repository-bumped
  attempt, so C5 can only free claims while C6 produces a fresh task id.
- Production application imports only the standard library plus
  `courses/internal/media/{domain,ports}`. There is no Asynq reference under the media
  application package.

## Verification performed

- `gofmt -w` ran over the four changed/new application files: exit 0.
- `gofmt -d` ran over the same files: exit 0 with no output.
- A direct compile-oriented read-through checked imports, interface method signatures,
  fake implementations, state mutation order, policy callback placement, failure outcomes,
  cleanup order and call-count assertions.
- No filesystem write was denied, so there is no denied-write error text to report.
- No Git command was run.

## Checks not run — denied by the brief

None of these checks is represented as passed:

- `make format-check` — not run; its Go half uses Docker. Local `gofmt` is the formatting
  evidence.
- `make lint` — not run; its Go-vet half compiles through Docker.
- `make typecheck` — not run; Go compilation is denied.
- `make test` — not run; Go compilation and test execution are denied, including the new
  fake-port tests and architecture boundary test.
- `make integration` — not run; Docker/Testcontainers are denied and adapters are outside
  M04.
- `make check-generated` — not run; it invokes Git and writes generated output, both denied.
- `make build` — not run; it requires Docker and compilation.
- `make up`, `make migrate`, `make seed`, `make seed-identities` and `make smoke` — not run;
  they require Docker Compose and are outside this application-only slice.
- `make e2e` — not run; it requires Docker Compose and M04 owns no browser surface.
- Aggregate `make check` — not run because it includes the denied checks above.

## Limitation and next task

The new Go code and fake tests have formatting and read-through evidence only; compilation and
test success are deliberately not claimed. The next planned slice is M05, the PostgreSQL
repository writes that implement these application-facing contracts.

## Reviewer verification (measured in the reviewing session)

One review fix: `transcode.go` declared a function named `processDisposition` colliding
with the type of the same name, so the package did not build (`redeclared in this block`,
plus two conversion errors at the call sites). The function is renamed `dispositionOf`.
This is exactly the class the sandbox cannot catch — Codex could not compile, said so, and
claimed only formatting evidence.

After the fix: `make test` pass — `courses/internal/media/application` green (22 test
functions across upload and transcode), domain green; `make integration` pass including
`courses/platform/architecture`, so the application layer's domain+ports-only rule is
machine-checked, not asserted; format-check, lint and typecheck pass. Provenance:
implemented with Codex (gpt-5.6-sol); reviewed, corrected and verified here.
