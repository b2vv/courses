# Report M03a — revised media domain and port seam

## Changed

- `backend/internal/media/domain/source.go` adds `ContainerSniffBytes = 12` and
  `AcceptContainer(head []byte) error`. The rule accepts the twelve listed ISO-BMFF major
  brands when bytes 4–7 are `ftyp`, or Matroska/WebM EBML magic at offset zero; every other
  shape and every head shorter than twelve bytes returns `ErrUnsupportedSource`.
- `backend/internal/media/domain/asset.go` adds `FailureReason.Permanent() bool`.
  `UnsupportedSource` is permanent; the other four closed failure reasons are retryable.
- `backend/internal/media/domain/source_test.go` and `asset_test.go` add table coverage for
  the complete accepted brand set, EBML, an unlisted brand, plain text, an eleven-byte head,
  and all five permanent-failure verdicts. This owns R43.
- `backend/internal/media/ports/repository.go` adds `ChunkPolicy`, `CompletePolicy`, and
  `PublishPolicy`, then adopts the exact §2.1 repository signatures: `Load`, `AppendChunk`
  replacing `BeginUpload` and `Advance`, policy-bearing `Complete` and `Publish`, explicit
  two-value `Claim`, and `DeleteSwept`.
- `backend/internal/media/ports/storage.go` adds the exact `Head(ctx, key, n)` signature.
- `backend/internal/media/ports/queue.go` adds `Job` with `AssetID`, `Attempt`, `RetryCount`,
  and `MaxRetry`.
- `backend/internal/media/application/transcode.go` places the required
  `TranscodeOutcome{Failed, Reason, Permanent}` value in the application package, where §2.1
  places it. It defines no M04 use case or behavior.

No adapter, SQL, HTTP, queue handler, application use case, dependency, or generated file was
added or changed. `work/tasks.md` was not edited.

## Contract read-through

- The three policies receive only a locked `domain.VideoAsset`; no database or framework type
  crosses the port. Their invocation order remains an adapter responsibility exactly as §2.1
  states: inside the transaction, after lock and guard, before the guarded update.
- `Publish` preserves both the policy argument and named evicted return; `Claim` reports the
  claimed disposition explicitly; sweep candidates remain separate from row deletion.
- Domain production files use only the standard library. Ports use only the standard library
  and media domain. The outcome value imports only media domain.
- No Asynq import is present in media domain, ports, or application; `error` remains reserved
  by the documented `Process` contract for retryable dependency faults.
- The removed `BeginUpload` and `Advance` names have no remaining media code reference.

## Verification performed

- `gofmt -w` ran over all eight changed/new Go files: exit 0.
- `gofmt -d` ran over the same files: exit 0 with no output.
- Direct standards and specification read-through matched every changed signature to plan
  §2.1 and both domain tables to D-8, D-10, and R43.
- The required two-axis review was adapted to the brief's Git denial: two independent,
  read-only reviewers inspected the known changed-file set without Git. Standards: zero
  findings. Specification: zero findings. Neither reviewer edited files, accessed a
  `generated/` directory, or ran Docker, Go compilation, or tests.

## Checks not run — denied by the brief

The brief states: “No Docker: no Go compile or test.” None of these checks is represented as
passed:

- `make format-check` — not run; its Go half uses Docker. Local `gofmt` is the only formatting
  evidence.
- `make lint` — not run; its Go-vet half compiles through Docker.
- `make typecheck` — not run; Go compilation is denied.
- `make test` — not run; Go compilation and test execution are denied, including the new
  domain unit tables and the architecture boundary test.
- `make integration` — not run; Docker/Testcontainers are denied and M03a adds no adapter.
- `make check-generated` — not run; it invokes Git and writes generated output, both denied.
- `make build` — not run; it requires Docker and compilation.
- `make up` — not run; it requires Docker Compose.
- `make seed-identities` and `make seed` — not run; they require the Compose database.
- `make smoke` — not run; it requires Compose.
- `make e2e` — not run; it requires Compose.
- `make migrate` and aggregate `make check` — not run because they include or depend on the
  denied environments above.

The code-review workflow's Git fixed-point, log, and diff commands were also denied by the
brief. No Git command was run and no workaround was attempted. No filesystem write was denied,
so there is no denied-write error text to report and the mandatory stop condition did not fire.

## Limitation and next task

The Go surface and tests have formatting and read-through evidence only; compilation and test
success are deliberately not claimed. The next planned slice is M04, the media application use
cases over the amended contracts.

## Reviewer verification (measured in the reviewing session)

`make test` pass — the media domain suite green with R43's table (accepted brands, EBML,
unlisted brand, plain text, short head). format-check and lint pass. Verified against §2.1
by inspection: the three policy types, `Load`, two-value `Claim`, `AppendChunk` (absorbing
`BeginUpload` and `Advance`), `Publish` with its policy and evicted return, `DeleteSwept`,
`MediaStorage.Head`, `ports.Job`, and `TranscodeOutcome` in `application/transcode.go`
where §2 places it. **Zero `asynq` references anywhere under `internal/media`** — the whole
point of the inward permanent-failure vocabulary, so it is checked rather than assumed.

One process note: Codex also wrote `docs/media/report.md`, the phase report M17 owns; it
was removed. Per-task evidence belongs in `work/report-*.md`, and the phase report is
written once at the end as a delta to the plan (the 4a/4b pattern). Provenance: implemented
with Codex (gpt-5.6-sol); reviewed and verified here.
