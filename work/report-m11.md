# M11 ship report — completion, disk gate, and upload executable

## 1. What was done

M11 now matches `docs/media/plan.md` §10, including the accepted §20 and §21 amendments.

- Completion is one `AppendChunk` transaction: `ChunkPolicy`, terminal `CompletePolicy`, and
  the guarded `queued` update execute under one row lock; enqueue happens only after the
  method returns. The unused `Complete` seam was removed.
- `ErrUnsupportedSource` is the one terminal `ChunkPolicy` outcome. PostgreSQL commits the
  locked `created` row as `failed` with reason `unsupported_source` and zero received bytes,
  then returns the typed error; the application maps it to `ErrUnsupportedMedia` without a
  second `Fail` transaction (`repository.go:183-200`, `upload.go:137-141`).
- The upload application performs the first-PATCH free-space gate, first-chunk container
  sniff, scratch-ahead repair, completed-source repair, terminal rename, and
  commit-then-enqueue (`upload.go:61-148`). The completed-source repair at offset zero runs
  before the disk gate because it writes no bytes.
- Added the upload composition root with dependency health plus a media-root write probe
  (`cmd/upload/main.go:35-63`), and added the no-ffmpeg Docker `upload` stage
  (`backend/Dockerfile:30-36`).
- Added the readiness probe itself and cleanup coverage (`storage.go:51-68`,
  `storage_integration_test.go:199-211`).
- Added integration coverage for P03, B18, B23, B31, B34, and the new B35. B35 waits until
  PostgreSQL reports the competing PATCH blocked on the real row lock before releasing the
  invalid policy (`tus/server_test.go:211-275`, `tus/server_test.go:558-585`).
- Updated §2.1's exact contract, added B34 and B35, assigned both to M11, and updated the
  acceptance totals to 11 happy + 35 boundary + 42 refusal = 88.

## 2. Tested scenarios

The brief denies compilation, Docker, and Testcontainers, so all six executable acceptance
rows remain unverified in this environment. Their tests and assertions were reviewed
statically; that is not represented as a pass.

| Row | Planned observable result | Actual result | Verdict |
|---|---|---|---|
| P03 | HEAD reports zero; terminal PATCH lands source, queues the row, and enqueues once | Integration case exists at `tus/server_test.go:296`; not executed | not verified |
| B18 | Invalid first container returns 400, commits failed/zero, no source or job | Integration case exists at `tus/server_test.go:192`; not executed | not verified |
| B23 | First PATCH below the 3× disk gate returns 507 without writing or advancing | Integration case exists at `tus/server_test.go:277`; not executed | not verified |
| B31 | Full source plus short row repairs to queued without another append | Integration case exists at `tus/server_test.go:146`; not executed | not verified |
| B34 | Full scratch plus short row truncates, replays, lands exact source, and enqueues once | Integration case exists at `tus/server_test.go:170`; not executed | not verified |
| B35 | Invalid PATCH commits failure while a valid PATCH waits; valid cannot advance; no source or job | Deterministic lock-wait case exists at `tus/server_test.go:211`; not executed | not verified |

Unverified acceptance rows: **6**.

## 3. Non-obvious findings

- The original §2.1 split completion into two transactions, contradicting D-2d and leaving a
  full-offset/no-source crash state. This was accepted as §20 (`docs/media/plan.md:2318`) and
  is covered by B34 (`docs/media/plan.md:1864`).
- The original §2.1 also made every policy error roll back, contradicting D-8's atomic ticket
  consumption. This was accepted as §21 (`docs/media/plan.md:2350`) and is implemented in the
  repository's terminal error branch (`repository.go:183-200`).
- The offset-zero completed-source repair must precede the first-write disk gate: a renamed
  source left by a pre-commit crash needs no new disk space or bytes (`upload.go:63-80`).

## Discussion

The two reported §2.1 conflicts are resolved by §§20–21. A complete read-through of the
touched policy contract against D-2d, D-8, D-15, B18, B31, B34, and B35 found no additional
conflict. Every non-`ErrUnsupportedSource` policy error still returns before any database
write and is rolled back; `CompletePolicy` errors do the same. The terminal unsupported
branch alone commits `failed`, and the application no longer opens a second failure
transaction.

## 4. Screenshots

Not applicable because M11 changes backend upload behavior, filesystem readiness, and an
image build stage; it has no user-interface change.

## 5. User-flow walkthrough

Not executed because the brief denies compilation and Compose. The implemented flow is:
the upload client reads the ticket offset, sends its first PATCH, receives 507 before any
write when capacity is insufficient or 400 after an atomically committed container refusal,
then continues valid chunks from the database offset. The terminal PATCH renames the source,
commits `queued`, and only then enqueues the job. A retry repairs either supported crash
shape before completing.

## 6. Risks and limitations

- `gofmt` completed with exit 0 on every changed Go file; `gofmt -l` returned no paths.
- Docker is denied. No Docker command, Testcontainers run, Compose run, or image build was
  attempted or worked around.
- Go compilation and test execution are denied. Focused application, filesystem,
  PostgreSQL-repository, and tus integration tests were not run.
- `make format-check`, `make lint`, `make typecheck`, `make test`, `make integration`,
  `make build`, and `make e2e` were not run and are not represented as passed.
- The code-review skill's parallel review workers were unavailable: the runtime agent
  inventory exposed only the primary agent. Both standards and specification axes were
  instead completed as read-through reviews; neither found an outstanding issue.
- No Compose/nginx, worker-handler, frontend, or API-storage change was made. The upload
  image contains no ffmpeg.
- No git command was run. `work/tasks.md` and `docs/media/report.md` were not edited. No file
  under a `generated/` directory was written.

## Post-deploy

Pending; M11 was not built or deployed in this environment.

## Reviewer verification (measured in the reviewing session)

Both Discussion items accepted and recorded as plan **§20** and **§21**, and both defects
were mine — §2.1's contract section, written after M04's stop, contradicted the decisions it
serves twice over: it split completion into two transactions where D-2d describes one, and
it declared a blanket "every policy error aborts" where D-8 requires the container-sniff
failure to consume the ticket atomically. Each left a window nothing else covered: an asset
stranded `uploading` with a full offset that no conforming client would ever advance, and an
invalid first chunk that could lose its guarded failure to a concurrent valid one.

Gates after the fixes: `make test` pass across all media packages; `make integration` exit 0
— `media/adapters/postgres` 41.9s and `media/adapters/tus` 31.7s, the longer runtimes being
the new deterministic race cases (B34 for the crash between final append and commit, B35 for
the concurrent first PATCH); `make build` exit 0 including the new `upload` image stage;
format-check, lint, typecheck and check-generated pass. `backend/cmd/` now holds four
executables — the third, `upload`, exists with its readiness write probe.

Provenance: implemented with Codex (gpt-5.6-sol); reviewed and verified here.
