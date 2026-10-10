# Report M10 — tus protocol server

## Outcome

M10 implements the bounded tus upload surface at `/uploads/{videoId}`: `HEAD`, `PATCH`, and
`DELETE` only. Every handled exchange carries `Tus-Resumable: 1.0.0`; `HEAD` returns the
database-owned offset, declared length, and expiry; `PATCH` requires the tus media type and
offset, caps the streaming body with `http.MaxBytesReader`, and returns the committed offset;
`DELETE` evicts the active ticket so its lesson can be minted again immediately. No cookie,
role, session, or same-origin middleware is present on the tus surface.

The existing PostgreSQL row lock and guarded advance remain the serialization point. The
application now exposes the active-ticket read needed by `HEAD`, gives chunk streaming a
120-second budget, classifies scratch length against the truth column, and makes every failed
append roll back before its refusal is returned. M10 adds no completion, sniff, disk gate,
rename, enqueue, `cmd/upload`, image/Compose/nginx, worker, or frontend work. The pre-existing
M04 `CompleteUpload` and first-chunk policies remain unchanged; the M10 server never calls
completion, and there is still no upload executable exposing the byte surface.

## Changed

- `backend/internal/media/adapters/tus/server.go` adds the three-verb handler, tus header and
  request validation, `HEAD` metadata, the 8 MiB server-side body cap, fixed safe error
  responses, and registration on the shared HTTP mux.
- `backend/internal/media/adapters/tus/server_test.go` adds one Testcontainers PostgreSQL plus
  tempdir-storage integration harness for B02, B03, R10, R11, R29, R30, and R40. It also pins
  the absence of POST/OPTIONS, the version precondition, PATCH media type, and expiration
  headers. Every fixture issues one SQL command per parameterized `Exec`.
- `backend/internal/media/application/upload.go` adds `HeadUpload`, separates the streaming
  chunk deadline from short metadata operations, classifies positive-offset scratch length,
  and repairs failed or over-declared writes before returning a refusal.
- `backend/internal/media/application/upload_test.go` pins active-only HEAD visibility,
  short-scratch refusal, rollback failure, rollback of an over-declared append, and the
  120-second chunk-operation budget.

No generated file was written. `work/tasks.md` and `docs/media/report.md` were not edited.

## Discussion

### B02 could leave the scratch file ahead of the truth column

Evidence: D-2d step 4 (`docs/media/plan.md:601-608`) and B02 require a chunk beyond the
declared length to return 413 with both the offset and file unchanged. Before M10,
`application.Upload.AppendChunk` returned `received_bytes + written` after a successful
append, while `postgres.Repository.AppendChunk` rejected a value beyond `declared_bytes`
only after the policy had returned (`backend/internal/media/adapters/postgres/repository.go:185-187`).
The transaction rolled back the column, but nothing rolled back the already-fsynced file, so
the next request started with a scratch file longer than the truth column.

Proposed correction: treat the declared length as a second byte-receiver cap inside the
transaction policy and truncate to the locked column before returning the refusal. Implemented
at `backend/internal/media/application/upload.go:93-95,117-122`, with direct fake-port evidence and
B02 integration evidence. The tus adapter maps this refusal to 413.

### D-2d's literal missing-file rule would refuse every first PATCH

Evidence: D-2d step 3 says a missing scratch file with no completed source is a 503
(`docs/media/plan.md:592-600`), but a newly minted `created` row has no scratch file by design;
the first PATCH is what creates it. Following that sentence literally would make every upload
fail before its first byte. The existing filesystem contract intentionally creates only the
legitimate zero-length scratch file when truncating to offset zero
(`backend/internal/media/adapters/filesystem/storage.go:105-116`).

Proposed correction: amend step 3 to say that missing scratch at `received_bytes = 0` is
initialized at zero; missing or shorter scratch at a positive truth-column offset remains a
503. Implemented with that distinction: offset zero uses the storage adapter's safe initializer;
positive offsets call `Size`, reject missing/short storage, and truncate only a long file. The
new application test pins the short-file refusal. This is an explicit divergence from the
literal sentence so that the protocol can start at all.

### A failed rollback was discarded

Evidence: D-2d step 4 requires the scratch file to be back at the column before returning an
overflow/write refusal (`docs/media/plan.md:601-609`). Before this correction,
`Upload.AppendChunk` ignored the error from its compensating `Truncate` after `Append`, `Head`,
or container validation failed. At runtime a failed truncate could leave uncommitted bytes
while the handler reported an ordinary resumable refusal.

Proposed correction: make rollback failure override the original refusal, so the tus adapter
returns a storage failure (503 or 507) unless reconciliation actually succeeded. Implemented
in `rollbackChunk`, with a fake-port test that forces append failure followed by rollback
failure.

### The shared three-second application deadline was shorter than the upload contract

Evidence: the existing `mediaOperationTimeout` was applied to `AppendChunk`, including the
request-body read and fsync. The plan permits an 8 MiB chunk and gives the streaming upload
route 120-second proxy send/read timeouts (`docs/media/plan.md:1431-1436`). At three seconds,
a valid chunk failed below roughly 22 Mbit/s before either planned proxy timeout; this would
break ordinary resumable uploads even though the protocol and nginx configuration were
correct.

Proposed correction: keep short metadata operations at three seconds and give only the chunk
transaction a 120-second deadline. Implemented at
`backend/internal/media/application/upload.go:15-16,56-58` and pinned in the application test.

### M11 cannot yet preserve the 120-second contract through the shared socket server

Evidence: plan placement says `cmd/upload` will use `httpserver.Run` (`docs/media/plan.md:400`),
but `backend/platform/httpserver/lifecycle.go:21-28` hardcodes `ReadTimeout: 10s` and
`WriteTimeout: 15s`. M10's handler integration harness calls the handler in-process, so those
socket deadlines are absent. At runtime, the future executable would still terminate or fail
to answer a chunk that legitimately takes longer than 10/15 seconds, despite the application
and nginx allowing 120 seconds.

Proposed correction: M11 should add a consumed, explicit timeout configuration to the shared
HTTP lifecycle (retaining today's API/worker defaults) and construct the upload executable
with upload-safe read/write limits aligned with the 120-second route. I did not add unused
server configuration in M10 and did not touch the M11-owned executable. This remains an open
M11 prerequisite; no current executable is falsely claimed to serve uploads.

### The brief's M11 NOT-list overlaps the M04 application it tells M10 to drive

Evidence: the brief says M10 must drive `application.Upload`, while its NOT-list assigns the
first-chunk disk gate and sniff to M11. Those policies already live inside
`Upload.AppendChunk` from M04 (`backend/internal/media/application/upload.go:62-69,97-112`),
so any server calling that required method necessarily reaches them on a `created` row.

Proposed correction: interpret the M10 NOT-list as “do not add or test the M11-owned policies”
and amend M11 to wire/prove the policies already implemented by M04 rather than reimplementing
them. I left those pre-existing branches unchanged, added no M11 acceptance evidence, and did
not create the executable that would expose them.

## Acceptance read-through

- **B02:** the application truncates an over-declared append to the locked
  `received_bytes`; the handler returns 413; the integration test follows with `HEAD` and
  checks the unchanged column and byte-identical scratch file.
- **B03:** `HEAD` reads the column, never the file; `PATCH` reads scratch size and truncates a
  long file to the column before appending the replacement tail; the test compares the final
  scratch bytes and column.
- **R10:** PostgreSQL rejects unknown, evicted, and expired rows before invoking the storage
  policy; the test checks 404, absent scratch files, and unchanged stored offsets.
- **R11:** the row-lock path compares the request offset before the policy; the test checks
  409, unchanged bytes, and the same subsequent HEAD offset.
- **R29:** two handler calls race at one offset against real PostgreSQL; the row lock lets one
  commit and makes the second observe an offset conflict. The test requires exactly one 204,
  one 409, and equality between scratch length, contents, and `received_bytes`.
- **R30:** `DELETE` calls the existing guarded cancel, the next PATCH gets 404, the row has
  `evicted_at`, and the existing JSON mint surface returns 201 for the same lesson.
- **R40:** the request explicitly has `Transfer-Encoding: chunked` and `ContentLength = -1`;
  the server's 8 MiB `MaxBytesReader` stops a 9 MiB body, storage rolls back, and HEAD plus SQL
  retain the prior offset.

## Verification performed

- `gofmt -w` ran over all four changed/new Go files: exit 0.
- `gofmt -d` ran after formatting: exit 0 with no output.
- Read-through checked handler-to-application method signatures, tus response headers and
  statuses, the row-lock/offset/policy/update order, rollback paths, one-command-per-`Exec`
  fixtures, safe fixed error messages, and every M10 NOT-list item.
- The required two-axis code review was adapted to the explicit M10 file set because Git is
  denied; its final disposition is recorded below.
- No filesystem write was denied, so there is no denied-write error text and the mandatory
  stop condition did not fire.
- No Git command was run and no commit was created.

## Review

### Standards

The independent standards pass found no hard documented-standard violation: dependencies
point inward, transport owns parsing/statuses, errors are fixed and safe, the tus surface has
no identity middleware, and no generated path was touched. It raised three judgement-call
smells: the broad fake repository and duplicate application timestamps predate M10 and remain;
the one-line tus error middleman was removed.

### Specification

The independent specification pass found two high write-protocol gaps and one low parser gap:
explicit positive-offset size classification, surfaced rollback failure, and digits-only tus
offset syntax. All three were corrected and covered by tests. Its M11 scope warning is recorded
in Discussion: the disk/sniff branches are pre-existing M04 work, unchanged here, and no
completion or executable was added. The fixture review confirmed one SQL command per
parameterized `Exec`.

Final review disposition: Standards — 0 hard findings, 2 retained pre-existing smell
judgements; Specification — 0 unresolved M10 implementation findings, 3 planning/briefing
Discussion items remain for the reviewer (zero-offset scratch initialization, socket timeouts,
and M04/M11 ownership wording).

## Checks not run — denied by the brief

The exact denial is: **“No Docker: no compile, no Testcontainers. gofmt + read-through.”**
None of the following is represented as passed:

- Focused `go test ./internal/media/application` — not run; compilation and tests are denied.
- Focused integration `go test -tags=integration ./internal/media/adapters/tus` — not run;
  compilation and Docker/Testcontainers are denied.
- Direct Go compilation, `go test`, `go vet`, and `go build` — not run.
- `make format-check` — not run; its Go path uses Docker. Local `gofmt` is the only formatting
  evidence.
- `make lint` — not run; its Go-vet path compiles through Docker.
- `make typecheck` — not run; Go compilation is denied.
- `make test` — not run; compilation and test execution are denied.
- `make integration` — not run; Docker/Testcontainers are denied.
- `make check-generated` — not run; it invokes Git-backed drift checks and Git is denied. M10
  changes no generator input or output.
- `make build`, `make up`, `make migrate`, `make seed-identities`, `make seed`, and
  `make smoke` — not run; they require Docker/Compose and include later-slice topology.
- `make e2e` — not run; it requires Docker/Compose and M10 owns no browser surface.
- Aggregate `make check` — not run because it contains the denied checks above.

## Limitation and next task

This slice has formatting and read-through evidence only; compilation and runtime behavior are
not claimed. A reviewer with Docker should run the focused application test, the M10
Testcontainers package, then the ordered gates. M11 owns completion, the third executable,
and the timeout-configuration correction raised above.

## Reviewer verification (measured in the reviewing session)

All three Discussion items accepted and recorded as plan **§19**. The second — D-2d step 3
refusing a missing scratch file — would have made every upload fail at its first byte, since
a freshly minted row has no scratch file by design; the plan carried a protocol that could
not start.

Two review fixes in the test harness: `insertAsset` hardcoded `generation` to 1, so R10's two
assets on one lesson collided on `video_assets_generation`; deriving it as `max+1` then broke
type inference because `$2` was used as both `uuid` and `text` (`42P08`), fixed with an
explicit cast. After both: `make integration` 20 packages ok with the tus suite green
(B02, B03, R10, R11, R29, R30, R40), format-check, lint, typecheck and check-generated pass.

Provenance: implemented with Codex (gpt-5.6-sol); reviewed, corrected and verified here.
