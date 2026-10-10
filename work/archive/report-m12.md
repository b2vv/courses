# M12 ship report — worker handler and startup wiring

## 1. What was done

- Added the `media:transcode` Asynq consumer. It decodes and validates the D-10 payload,
  supplies retry metadata to `Transcode.Process`, and implements the four adapter outcomes:
  dependency error unchanged, permanent outcome wrapped with `asynq.SkipRetry`, non-permanent
  failed outcome returned as its reason, and zero outcome returned as success
  (`backend/internal/media/adapters/asynq/handler.go:36-70`).
- Added bounded structured Asynq error logging. Invalid payload identifiers and raw dependency
  errors never reach logs (`handler.go:73-101`).
- Wired the PostgreSQL repository, filesystem storage, ffmpeg transcoder and queue into the
  worker. Startup completes `Recover` and then `Sweep` before `server.Start`; the existing
  inspector, queue health and graceful shutdown remain (`backend/cmd/media-worker/main.go:49-94`).
- Worker readiness now includes the media-root write probe. `/status` reports stage `media`
  and exactly one handler, `media:transcode` (`main.go:98-110`).
- Added the duration histogram with 1–2048 second exponential buckets and the bounded failure
  counter. All three outcome series and five failure-reason series are pre-initialised at zero
  (`backend/platform/observability/metrics.go:19-39`).
- Inverted the worker-status smoke assertion, added all B19 zero-series assertions, and added
  a printed, failing assertion for the worker's ffmpeg 5.1.x line
  (`infrastructure/tests/smoke.sh:3-31`).

## 2. Acceptance rows

The environment denies Docker, compilation and smoke execution. The assertions exist but are
not represented as passes.

| Row | Planned observable result | Actual result | Verdict |
|---|---|---|---|
| P10 | `/status` returns `stage: media` and exactly `media:transcode` | Worker response and smoke assertion implemented; not executed | not verified |
| B19 | All three duration and five failure series exist at zero before a job | Registration, unit assertion and smoke assertion implemented; not executed | not verified |
| B33 | Worker prints and asserts ffmpeg 5.1.x | Smoke command prints the first version line and rejects another line; not executed | not verified |

Unverified acceptance rows: **3**.

## 3. Non-obvious findings

- An Asynq payload is an external queue boundary. A non-empty but malformed asset id would
  otherwise reach PostgreSQL's UUID cast and be retried as a dependency failure, so the adapter
  validates the same lowercase, non-zero UUID shape used by the existing media adapters.
- The `ErrorHandler` never logs the returned error text: dependency errors can contain storage
  paths or process detail. It maps only to bounded reason constants.

## Discussion

### D-14 asks for a state that the published outcome contract cannot always supply

Evidence: D-14 requires every media log line to carry the asset state
(`docs/media/plan.md:1289-1292`), but `TranscodeOutcome` carries only `Failed`, `Reason` and
`Permanent` (`backend/internal/media/application/transcode.go:21-27`). On dependency faults,
`dependencyFailure` returns the original error whether `Requeue` succeeded or finalisation
failed (`transcode.go:210-219`). Therefore the adapter cannot know whether the stored row is
`queued` or still `processing`; logging either as fact would be wrong, and querying PostgreSQL
from the error logger can fail for the same dependency outage.

Proposed correction: amend D-14's logging vocabulary to allow bounded `state="unknown"` only
for invalid payloads and dependency/finalisation faults; retain exact `queued`/`failed` for
business outcomes and cancellation. Meanwhile M12 does exactly that (`handler.go:75-99`) and
does not fabricate a database state.

Runtime consequence if left implicit: an operator could read a fabricated `failed` state and
look for a retry control while the row is actually `queued`, or assume it is safely requeued
while it remains `processing` after a failed terminal write.

## 4. Screenshots

Not applicable. M12 changes backend worker behavior and internal smoke observability only.

## 5. User-flow walkthrough

Not executed because Docker and compilation are denied. The implemented flow is: worker opens
dependencies and `/media`, recovers abandoned claims, sweeps reclaimable assets, registers
`media:transcode`, then starts Asynq. A delivery is decoded into `{assetId, attempt}` plus
Asynq retry counts; the existing application claim guards process it; the adapter records a
bounded duration/failure outcome and returns the exact retry or archive signal.

## 6. Risks and limitations

- `gofmt` completed with exit 0 on all changed Go files; `gofmt -l` returned no paths.
- Docker is denied. No Docker, Testcontainers or Compose command was attempted or worked around.
- Compilation is denied. The handler and metric unit tests were not run.
- Smoke and build are denied. The P10, B19 and B33 assertions were not executed.
- The complete ordered checks were not run: `make format-check`, `make lint`,
  `make typecheck`, `make test`, `make integration`, `make check-generated`, `make build`,
  `make up`, `make seed-identities`, `make seed`, `make smoke`, and `make e2e`.
  Consequently `make check` was also not run. None is represented as passed.
- No compose service, nginx, frontend, domain, application, generated-file, or dependency
  change was made.
- No Git command was run. `work/tasks.md` and `docs/media/report.md` were not edited. No file
  under a `generated/` directory was written.

## 7. Code review

- **Standards:** pass after correction; no remaining hard standards issue or actionable smell.
- **Specification:** pass after correction; no remaining concrete M12 implementation bug.
- The review initially found unsafe invalid-payload logging, malformed UUID retries, and a
  fabricated dependency-path state. All three were corrected before this report; the contract
  limitation behind the last item is recorded in Discussion above.

## Post-deploy

Pending; M12 was not built, started or deployed in this environment.

## Reviewer verification (measured in the reviewing session)

The Discussion item is accepted and recorded as plan **§22**: `state="unknown"` is admitted,
bounded to invalid payloads and dependency/finalisation faults, because on a dependency
outage the adapter genuinely cannot tell `queued` from `processing` and querying the
database from the error path can fail for the same reason. A fabricated state would send an
operator hunting for a retry control on a row that is actually queued.

Gates: `make test` pass (`media/adapters/asynq` green), format-check, lint, typecheck pass;
`make build`, `make up`, seeds and **`make smoke` exit 0**. Verified in the running system:
the worker's mux is no longer empty — `mux.Handle(TaskTypeTranscode, ...)` with `Recover`
then `Sweep` completing **before** `server.Start` — and `asynq.SkipRetry` appears in exactly
two files, the adapter and its test, nowhere else, which is what the inward permanent-failure
vocabulary was introduced for.

**B33 is live: smoke now prints `ffmpeg version 5.1.9-0+deb12u1`** — the binary that actually
runs, not the host's 7.1.1. That guard exists because three separate gate reviewers verified
by execution against the wrong binary; from now on every gate run records which one produced
the media.

The queue has a consumer for the first time in the repository's history. Provenance:
implemented with Codex (gpt-5.6-sol); reviewed and verified here.
