# Report M07 — filesystem MediaStorage adapter

## Outcome

M07 implements the filesystem `MediaStorage` adapter over one owned `os.Root`. Opaque
`domain.StorageKey` values map internally to the `source/`, `hls/`, `thumbnails/`, and
`temporary/` layout; no filesystem path crosses the port boundary. The adapter uses only the
Go standard library and has no database.

## Changed

- `backend/internal/media/adapters/filesystem/storage.go` adds `Storage`, opens and retains an
  `os.Root`, creates the four owned top-level directories, and implements the complete existing
  `MediaStorage` port.
- Key resolution accepts only the four storage kinds, non-zero lowercase UUID asset IDs, and
  the fixed leaf shapes used by upload/render work. Every filesystem operation runs through
  `os.Root`; no clean-and-prefix string containment check exists.
- `Append` refuses a missing or short file at a non-zero offset, truncates a longer file back
  to the database-owned offset, writes from that offset, and calls `fsync`. A body, context, or
  `fsync` failure truncates and syncs back to the starting offset before returning the error.
  `Truncate` uses the same no-zero-padding reconciliation rule and creates only the legitimate
  zero-length initial scratch file.
- `Publish` renames a completed temporary directory onto a fresh HLS destination. Destination
  exists plus missing work is the already-published no-op branch; destination and work both
  exist is the stale pre-commit branch, which removes the stale directory before rename. A
  missing work and destination returns an error without creating `hls/{id}`.
- `Purge` delegates singular keys to `os.Root.RemoveAll`, where ENOENT is success. An asset-wide
  temporary purge enumerates `temporary/` inside the root and recursively removes every
  `{videoId}-*` path while preserving other assets.
- `Head` returns up to the requested byte count and therefore succeeds on a short file. `Free`
  calls `Fstatfs` on a directory descriptor opened from the owned root and returns available
  bytes.
- `backend/internal/media/adapters/filesystem/storage_integration_test.go` adds tempdir-only
  coverage for B11 and B12 plus offset repair, short-file refusal, failed-write rollback,
  `Move`, `Size`, short-file `Head`, asset-wide temporary purge, and positive free space.

No domain, port, application, PostgreSQL, ffmpeg, HTTP, tus, worker, frontend, dependency, or
generated file was changed. `work/tasks.md` and `docs/media/report.md` were not edited.

## Acceptance read-through

- **B11:** `..`, an absolute leaf, and a non-UUID ID fail key resolution before an open. A
  valid HLS key whose path is a symlink to `/etc` fails instead of statting that directory. A
  second valid source key points through a symlink to an outside tempdir; `Truncate` fails and
  the outside file remains byte-identical.
- **D-2d:** an offset greater than the current file length returns `io.ErrUnexpectedEOF`; a
  longer file is shortened to the supplied offset before append; a reader failure after a
  partial write leaves the file exactly at that offset. Successful appends sync before return.
- **D-8 / B12:** first publication moves a complete tree atomically. A work tree replaces a
  stale destination only after that destination is removed; a repeated already-published call
  is a no-op; an interrupted call with neither work nor destination errors and leaves no
  `hls/{id}` directory. Purging that absent destination twice succeeds.
- **D-15:** free capacity is calculated from `Bavail * Bsize` on the root filesystem. The
  adapter exposes the port operation only; the API was not wired to storage.
- Source and completed-upload movement is same-root `os.Root.Rename`; temporary cleanup covers
  upload scratch, render directory, and render thumbnail names for one asset.

## Review

The implementation workflow's required code review was adapted to the brief's Git denial and
run over the explicit two-file M07 set on independent Standards and Specification axes.

- Standards: no documented-standard violation remains. One low judgement-call duplication is
  retained: UUID validation mirrors the domain's private `validID`. Reusing it would require a
  forbidden domain-contract edit, while adding a UUID dependency would be larger and less
  direct than the standard-library check.
- Specification: the first pass found ambiguous symlink evidence, a missing already-published
  no-op, short-file zero-padding, and incomplete rollback after failed appends. All were fixed
  and covered by runnable tests. The final re-review reported no remaining specification
  finding.

## Verification performed

- `gofmt -w` ran on both M07 Go files: exit 0.
- `gofmt -d` ran after the final edits: exit 0 with no output.
- Read-through checked all eight `MediaStorage` signatures against
  `backend/internal/media/ports/storage.go`, all four key kinds against the plan layout, every
  write/open/remove/rename call for `os.Root` confinement, and every M07 NOT-list term.
- The explicit-file two-axis review completed after fixes: Standards had zero hard findings;
  Specification had zero findings.
- No filesystem write was denied, so there is no denied-write error text and the mandatory stop
  condition did not fire.
- No Git command was run. No commit was created because the brief explicitly denies Git,
  overriding the generic implementation workflow's commit step.
- `work/tasks.md` and `docs/media/report.md` were not edited. No file inside a `generated/`
  directory was written.

## Checks not run — denied by the brief

The brief states: “No Docker: no Go compile or test.” None of the following is represented as
passed:

- Direct `go test ./internal/media/adapters/filesystem` — not run; Go test execution is
  explicitly denied. The reviewer can run it from `backend/`; it needs only tempdirs.
- Direct Go compilation, `go test`, `go vet`, and `go build` — not run; compilation and tests
  are explicitly denied.
- `make format-check` — not run; its Go path uses Docker. Local `gofmt` is the only formatting
  evidence.
- `make lint` — not run; its Go-vet path compiles through Docker.
- `make typecheck` — not run; Go compilation is denied.
- `make test` — not run; Go compilation and test execution are denied.
- `make integration` — not run; it requires Docker, although M07's new package itself requires
  no database or Testcontainers.
- `make check-generated` — not run; it invokes Git-backed drift checking, and Git is denied.
  M07 changes no generator input or output.
- `make build` — not run; it requires Docker and Go compilation.
- `make up`, `make migrate`, `make seed-identities`, `make seed`, and `make smoke` — not run;
  they require Docker Compose and are outside this adapter-only slice.
- `make e2e` — not run; it requires Docker Compose and M07 owns no browser surface.
- Aggregate `make check` — not run because it includes the denied checks above.

## Limitation and next task

This slice has formatting and read-through evidence only; compilation and runtime test success
are deliberately not claimed. A reviewer should run the focused filesystem package first,
then the ordered gates in an environment with Go and Docker. The next planned slice is M08,
the ffmpeg `Transcoder`; this report makes no claim that it exists.

## Reviewer verification (measured in the reviewing session)

`make test` and `make integration` pass — `courses/internal/media/adapters/filesystem`
green, with `TestB11StorageCannotBeWalkedOutOf` and
`TestB12PublishIsAtomicAndPurgeIsIdempotent` naming their acceptance rows, plus
truncate-then-append, `Head` on a short file, purge coverage and `Free`. Verified by
inspection that `os.Root` is used for real: the root is opened once and **both the opens
and the renames go through it**, so traversal and symlink escapes are refused by the
runtime rather than by string inspection — the reason the plan chose it. format-check and
lint pass. These tests need only a tempdir, so unlike most of this phase they ran here at
full strength. Provenance: implemented with Codex (gpt-5.6-sol); reviewed and verified here.
