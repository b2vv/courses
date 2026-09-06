# Media upload and transcoding — Phase 5 plan

Edition 3.4, 2026-09-06. Edition 3.4 is a **focused amendment made after implementation
began**: M01–M03 landed green, then M04 stopped without writing code and reported six
contract-shape gaps (`work/report-m04.md`). The report is right and the fault is this plan's —
§2 listed port **method names**, which cannot express the inside-one-transaction ordering D-7,
D-8, D-10 and D-15 require. §2.1 now gives exact Go signatures with three policy callbacks and
states where each runs; D-8 gains the container sniff as a named domain function; D-10 gains
the inward permanent-failure vocabulary; and one task, **M03a (1 h)**, revises the M03 surface
before M04 reruns. §16 records it. Nothing else was restructured and no row changed owner.
Edition 3.3 followed **GATE 2 round 6, which returned GREEN**: N-17 was closed by execution
inside `courses-media-worker-1` on ffmpeg 5.1.9 (B26's presences and absences match the muxer
attribute for attribute), and the tightened named CHECK was verified against live PostgreSQL
17.11 — all four empty shapes plus `{360p,NULL}` and `{360p,""}` refused as
`23514 video_assets_ready_has_renditions`, on the UPDATE path as well as the INSERT. Edition
3.3 folds that round's four non-blocking findings (§15) and changes nothing else. Edition 3.2
closed the one blocking finding of GATE 2 round 5, which
was the first **critique** pass on editions 3/3.1: **every ffmpeg fact this plan called
"verified" was executed against the host's 7.1.1, while the worker image ships Debian
bookworm's 5.1.9** (`backend/Dockerfile:12,29-31`;
`docker exec courses-media-worker-1 ffmpeg -version`). The mechanisms survive that — the
`min()` filter, `force_divisible_by=2`, the init naming and the segment tags all behave the
same on 5.1.9 — but the playlist attributes do not, so §0, D-8, D-11 and B26 are re-stated
against the binary that actually runs, and §5.d adds a smoke assertion that prints it. Edition
3.1 folded one executed defect from a late round-2 **defense** review: the
`ready ⇒ at least one rendition` CHECK **enforced nothing** — `array_length` returns `NULL` on
an empty array and a `NULL` CHECK passes. Edition 3 superseded
edition 2, which GATE 2 round 2 returned **RED**
(`docs/media/gates.md`): seven of the eight edition-1 blocking findings were confirmed closed,
**BLK-1's crash/restart half was not**, and three new blocking findings were added — three of
the four established by execution against ffmpeg 7.1.1, `asynq@v0.26.0` and
`docker-compose.yml`. §13 records the disposition of all four blocking and twelve non-blocking
round-2 findings; §12 keeps the round-1 disposition. Edition 2 superseded edition 1 (GATE 2
round 1: 8 blocking / 16 non-blocking, plan-defense rows 2 and 4 open); those closures stand
where §12 names them.

**The two defense rows, restated with their real closures.** Row 2 (dependency failure) is
closed in D-15, whose disk-gate placement moved in this edition because the API has no media
mount. Row 4 (the sub-360p source) was closed in edition 2 **on a false claim** — the filter
it named upscales 320×240 to 480×360, executed and disproved at round 2 — and is closed here on
the corrected `min()` expression, with B30 producing the evidence from ffmpeg rather than from
prose.

Sources, read in order and re-opened against the code: `AGENTS.md`; `.claude/standards.md`
(gate order, placement, reuse index, branches); `docs/media/spec.md` (GATE 1 decisions
Q1–Q8, **binding**); `work/PRODUCT-BRIEF.md` §2 rules 12/13/14, §10 (Media module and the
six video states), §11 (media architecture, the file layout, the profiles, the twelve-step
worker sequence, "no arbitrary user arguments to FFmpeg"), §16, §17, §24.11–24.14;
`docs/media/research-codecs.md` (37 sources; the Apple HLS Authoring conformance table and
the AV1 trigger — D-11 and D-16);
`docs/admin/plan.md` (4a) and `docs/admin/plan-4b.md` (4b) for the seam, the copy-not-import
precedent, the isolation-level trap and the canary-renumber pattern;
`work/CTO-RESEARCH.md` §6; the code named inline throughout. **Where a source and the code
disagreed, the code won**; §0 lists where that happened.

Mode: **ponytail-full** — the simplest thing that works, existing seams before anything new,
every deviation from the brief's stack or from a GATE 1 decision named with the trigger that
would revisit it.

**Scope fence.** Phase 5 produces bytes and nobody may read them. No playback, no signed
URLs, no `auth_request`/`secure_link`, no hls.js, no player, no progress, no transcripts, no
captions, no orders, no S3/MinIO adapter, no DRM, no ASR, no Python. `/media/*` stays
**403 at the edge for everyone** (`infrastructure/nginx/nginx.conf:44-45`), and this plan
changes neither of those two lines.

---

## 0. Verified facts

| Claim | Verified at | Verdict |
|---|---|---|
| The worker's mux is empty; no producer exists anywhere | `backend/cmd/media-worker/main.go:50-54` | confirmed |
| ffmpeg is already in the worker image | `backend/Dockerfile:29-33` | confirmed |
| `media-data` is writable in the worker, read-only in nginx | `docker-compose.yml:88`, `:117` | confirmed |
| `/media` and `/media/` answer 403 | `nginx.conf:44-45`; `smoke.sh` asserts it | confirmed; **Phase 5 keeps it** |
| Next migration number is 00005 | `backend/migrations/` holds 00001–00004 | confirmed |
| The failing-migration canary collides | `backend/internal/course/adapters/postgres/repository_integration_test.go:190` — subtest **"failed migration rolls back its complete transaction"**, `fstest.MapFS` at `:195`, canary `"00005_failure.sql"` at `:200` | confirmed |
| Lessons have no media column | `00003_authoring.sql` | confirmed |
| Course *writes* use `pool.Begin`; only the *reads* are `RepeatableRead` | `authoring.go:76,124` (reads), `:147,186,274,317` (writes) | briefing imprecise; media follows the writes — D-9 |
| The kit has `FileUpload` and `ProgressBar` | `controls.tsx:165-178`, `data.tsx:7-23`, `index.ts:10` | exist, **no consumer outside `ui.test.tsx`** — §5.g |
| `LockEligibleAuthoringOwner` is the ownership check | `courses.sql:105-111` is `FOR SHARE` and answers "may this user *own* a course" | briefing wrong: the *may this actor edit this course* check is `Actor.CanManage` over `lockAuthoringCourse` (`courses.sql:89`, `FOR UPDATE`; `authoring.go:193-199`). Media reuses the **rule** — D-3 |
| `.gitignore` ignored `docs/media/` and the whole new module | `.gitignore:14` was `media/`, unanchored | **found in edition 1, fixed and committed in `5ef6c1f`** before this edition: the pattern is now `/media/`, `git ls-files docs/media` lists `plan.md` and `spec.md`, and `git check-ignore` matches none of the Phase 5 paths. Nothing left to do; recorded because the finding shaped M01 in edition 1 |

Three facts the sources did not state, found in the code and load-bearing below:

- **`infrastructure/tests/smoke.sh` asserts the worker has exactly zero handlers.**
  Registering the first handler turns `make smoke` red until that assertion is inverted — §5.d.
- **A 4a `PUT` deletes lessons the payload omits** (`authoring.go:237`,
  `deleteMissingAuthoringLessons`). Anything hanging off `course_lessons.id` is deleted by an
  ordinary course save — D-4 exists because of that line.
- **Asynq re-delivers the same payload while the previous goroutine may still be running**
  (`asynq@v0.26.0/processor.go:245`: lease expiry cancels the context and calls
  `handleFailedMessage(…, ErrLeaseExpired)`). D-10's claim token exists because of that line.

Facts established by the GATE 2 critics through execution or by reading the source, adopted
here as evidence. **Every ffmpeg fact names the binary it was produced on and where that binary
runs**, because rounds 2–4 attached a "verified" voice to the host's ffmpeg while the worker
image ships a different major version — the defect N-17 caught, and the reason D-11 now targets
the deployment binary rather than the developer's.

**The deployment binary.** `backend/Dockerfile:12` is `FROM debian:bookworm-slim AS runtime`
and `:29-31` builds the worker stage on it with `apt-get install ffmpeg`, so
`docker exec courses-media-worker-1 ffmpeg -version` reports **5.1.9-0+deb12u1**. A developer
host may have anything; the one measured during these gates was Homebrew's **7.1.1**. Phase 5
targets 5.1.9 (D-11).

| Fact | Binary / where run | Holds on the other? |
|---|---|---|
| Init segments are `init_N.mp4` per variant in `var_stream_map` order, audio group first (`audio/init_0.mp4`, `360p/init_1.mp4`, `720p/init_2.mp4`) — D-8 | **5.1.9, in `courses-media-worker-1`**; also 7.1.1 on the host | yes, identical |
| `scale=…:force_original_aspect_ratio=decrease` on a **1920×1040** source gives `640x347` and libx264 aborts with *"height not divisible by 2"*; `force_divisible_by=2` fixes it — D-11 | **7.1.1, host** (the original run); re-run **in the container on 5.1.9**: `640x346`, `1280x692`, `1920x1040`, exit 0 | yes |
| The same option **does not cap at the source** — 320×240 encodes at `480x360`; `w='min(640,iw)':h='min(360,ih)'` keeps `320x240` — D-11 | **7.1.1, host** (the disproof); confirmed **in the container on 5.1.9**: `RESOLUTION=320x240`, thumbnail `320,240` | yes |
| No `FRAME-RATE` anywhere; `EXT-X-INDEPENDENT-SEGMENTS` in the video variant playlists, not the master — D-11, B26 | **5.1.9, in the container**; also 7.1.1 | yes, identical |
| **`AVERAGE-BANDWIDTH` is not written at all**, and `BANDWIDTH` is nominal — `542300` for the 360p rung on a 320×240, a 640×360 and a 640×346 source alike; no `CHANNELS` on the audio `EXT-X-MEDIA` — D-11, B26 | **5.1.9, in the container** | **no** — 7.1.1 writes a measured `AVERAGE-BANDWIDTH` and `CHANNELS`. This is the difference that made B26 red, and it is why the plan targets one binary explicitly |
| The 1080p rung without `-level` probes as High level 40 (`avc1.640028`); the lower rungs `avc1.640029` — D-11 | **5.1.9, in the container** | consistent |
| Under `cpus: 2.0`, a 60 s 1280×720 source through the two-rung argv takes **9 s**; a 2 s clip renders in well under a second — R35, §5.f | **5.1.9, in the container** | n/a |

**Source facts** (read, not executed):

- **Asynq's shutdown does not cancel the handler first.** `Server.Shutdown`
  (`server.go:724-751`) waits `ShutdownTimeout` with the context live, requeues the message to
  *pending* (`processor.go:236-239, 260-272`), and only then cancels (`processor.go:207-210`);
  `srv.wg.Wait()` excludes the handler goroutine, so `main` exits underneath it. Compose gives
  every service `stop_grace_period: 40s` (`docker-compose.yml:7-10`), comfortably more than the
  10 s `ShutdownTimeout`, so that requeue really completes before SIGKILL — D-7, D-10.
- **Retry backoff is `n⁴ + 15 + rand(30)·(n+1)` seconds** (`server.go:401-405`), so retries 1–3
  wait **15–44 s, 16–74 s, 31–118 s** — D-10.
- **A lease-expiry recovery consumes one of the three retries**: `recoverer.go:96-102` calls
  `rdb.Retry`, which increments `Retried` — D-10.
- **Asynq keeps a task id after the task ends**: `internal/rdb/rdb.go:98-108` refuses an enqueue
  whose task key exists, `:382-392` keeps the key for the 24 h retention after completion, and
  `:861-885` keeps it for the archive's lifetime — D-7 C6, D-10.
- **The `api` service has no `media-data` mount** (`docker-compose.yml:72-73`: `read_only`,
  `tmpfs: [/tmp]`), so a `statfs` there has nothing to stat — D-15.
- **PostgreSQL 17**, §6's DDL run verbatim in a scratch schema: the `ready`/`failed`/claim-token
  CHECKs bite, `2147483648` is accepted and `2147483649` refused, a second pending row raises
  `23505 video_assets_pending_per_lesson`, and an evicted row does not block — but `'{NULL}'`
  and `'{""}'` satisfy `cardinality(…) >= 1`, auto-generated constraint names are
  `video_assets_check…N`, and an `updated_at` one microsecond behind `DEFAULT now()` raises
  `23514` — §6.

---

## 1. Shape of the slice

One new backend module, one new executable, one new queue handler, three JSON operations, a
three-verb tus surface, one migration, one Nginx location, one Compose service, one editor
control.

**The byte path never touches the Go API or the SSR server** (§2 rule 13, §24.11). The
browser talks tus to a dedicated `upload` executable through Nginx; that executable writes
onto the media volume and enqueues; the worker transcodes and publishes; the API sees only
metadata.

```
browser ──tus /uploads/{videoId}──▶ nginx ──▶ upload (:8082) ──write──▶ media-data ──read──▶ media-worker
   │                                              │                                             │
   └──JSON /api/v1/admin/──▶ nginx ──▶ api        │                                             │
                                       └───────── PostgreSQL: video_assets ────────────────────┘
                                                        │
                          nginx ──/media/* → 403 (unchanged in Phase 5) ──▶ media-data:ro
```

### The JSON surface (in `openapi.yaml`, Orval-generated)

| Method | Path | Who | Answer |
|---|---|---|---|
| POST | `/api/v1/admin/lessons/{lessonId}/video` | admin any lesson; instructor only lessons in courses they own | body `{declaredBytes: integer 1…2147483648, originalName: string 1…255}` — **no `contentType`**, because the first-chunk sniff is the authoritative container check (D-8) and a client-declared MIME is the least trustworthy signal. Answers `201` `{videoId, uploadUrl, chunkBytes, maxBytes, expiresAt}`; a `video_assets` row in state `created` |
| GET | `/api/v1/admin/courses/{courseId}/videos` | same rule, scoped by the course | `200` `{items: [{lessonId, videoId, state, generation, originalName, receivedBytes, declaredBytes, updatedAt, failureReason?}]}` — one row per **non-evicted asset** |
| POST | `/api/v1/admin/videos/{videoId}/retry` | same rule | `200` with the asset in `queued` and a task enqueued; `409 video_not_retryable` unless the state is `failed`, `queued`, or `processing` older than the claim age (D-10) — and in every case only when the source actually landed |

Three operations across three path templates.

### The tus surface (not in `openapi.yaml` — D-2c)

`HEAD /uploads/{videoId}`, `PATCH /uploads/{videoId}`, `DELETE /uploads/{videoId}` — tus
1.0.0 **core** plus the `expiration` extension. **No `creation` extension and no `OPTIONS`**:
the JSON mint *is* the creation, so the client already has its URL and its length, and
`OPTIONS` exists for generic-client discovery we do not need (BLK-3).

### The video state machine

Exactly the six states of `work/PRODUCT-BRIEF.md` §10. No seventh state is invented;
eviction is a timestamp, the 4b idiom (D-6).

```
created ──(first PATCH)──▶ uploading ──(completion)──▶ queued ──(claim)──▶ processing
                                                        ▲  ▲                  │
                              (retry) ──────────────────┘  └──(requeue)───────┤
                                 │                                            │
                              failed ◀───────────(final attempt)──────────────┤
                                                                              ▼
                                                                            ready
```

| # | Transition | Written by | Guard |
|---|---|---|---|
| T1 | → `created` | **API**, `POST …/video` | ownership over the locked course row; no other non-terminal asset for the lesson (D-15 places the free-space gate at the first PATCH, not here) |
| T2 | `created` → `uploading` | **upload service**, first `PATCH` | valid, unexpired ticket; the container sniff passes (D-8) |
| T3 | `uploading` → `queued` | **upload service**, on completion | `received_bytes == declared_bytes`; the source renamed into place; **committed before the enqueue** (D-10) |
| T4 | `queued` → `processing` | **worker**, on claim | guarded `UPDATE … WHERE state='queued'`, writing a fresh `claim_token` |
| T5 | `processing` → `ready` | **worker**, after atomic publish | guarded on `(state='processing', claim_token=$token)`; eviction applied in the same transaction |
| T6 | `processing` → `queued` | **worker**, on a retryable failure or on cancellation | guarded on `(state='processing', claim_token=$token)`; `claim_token` cleared |
| T7 | `processing` → `queued` | **worker recovery pass**, at start | **every** `processing` row, at any age — this process is the only claimant, so none of them is alive (D-7) |
| T8 | `created`/`uploading`/`processing` → `failed` | **worker** (final attempt or malformed source) or **upload service** (failed sniff) | guarded on `(state, claim_token)`; a reason is mandatory by CHECK. `created → failed` is the sniff refusal on the **first** PATCH, when the row has not yet reached `uploading` |
| T9 | `failed` \| `queued` \| stale `processing` → `queued` | **API**, `POST …/retry` | the source landed (`received_bytes = declared_bytes`); state is `failed`, `queued`, or `processing` with `updated_at` older than the claim age; clears `claim_token`, increments `attempt`, enqueues |

T6, T7 and T9 are the answer to BLK-1, and each covers a case the others cannot:

- **T6** — the attempt that *can* write: it returns the row to `queued` before returning a
  retryable error, so the next Asynq delivery claims cleanly.
- **T7** — the attempt that *could not* write, because the process died or was restarted.
  It is threshold-free (D-7), so a restart always frees the row.
- **T9** — the attempt that could not write while the **process kept running** (the D-15
  "PostgreSQL down on the final attempt" path): no restart is coming, so the operator's retry
  is the exit, permitted once the claim is older than the task timeout.

**No separate lease column exists**: `claim_token` is the lease *identity* and `updated_at`
its *timestamp*, and every write a claim makes is guarded on both, so a re-delivered attempt
can never publish over a row a later claim owns.

---

## 2. Placement (Розміщення)

Per `.claude/standards.md` §Розміщення. Every new file is named; extended files are marked
`(extend)`. Sibling `*_test.go` / `*.test.ts(x)` accompany every new module.

### Backend — the new `internal/media` module

| File | Contents |
|---|---|
| `backend/migrations/00005_media.sql` | `video_assets`, its named CHECKs, the three indexes, `ON DELETE SET NULL` to `course_lessons`; down migration (§6) |
| `backend/internal/media/domain/asset.go` | `VideoAsset`, `State`, `ClaimToken`, `Generation`, `StorageKey` and its four kinds, the transition matrix of §1, ticket bounds, `NewAsset` invariants |
| `backend/internal/media/domain/actor.go` | `Actor{UserID string; Administrator bool}`, `CanEditLesson(ownerID string) error` — capability, not role vocabulary (D-3) |
| `backend/internal/media/domain/source.go` | `SourceProbe`, the accepted codec set (Q2), `AcceptSource`, `SelectLadder` (always ≥1 rung, never upscaling — D-11), **`AcceptContainer` + `ContainerSniffBytes`** (the first-chunk container rule, D-8), **`FailureReason.Permanent()`** (D-10) |
| `backend/internal/media/domain/errors.go` | the business errors of §8 |
| `backend/internal/media/domain/*_test.go` | table tests: transitions, ladder (landscape, portrait, sub-360p), source acceptance, ticket bounds |
| `backend/internal/media/ports/repository.go` | `AssetRepository` and its three policy types — exact signatures in §2.1. **Ownership is inside `Mint`/`Retry`/`ListCourseVideos`, not a second port** (D-1) |
| `backend/internal/media/ports/storage.go` | `MediaStorage` — opaque `StorageKey` in, no filesystem path crosses the boundary (ADR 004); `Free` for the disk gates, whose only callers are the upload service's first `PATCH` and the worker's claim — never the API (D-15). Signatures in §2.1 |
| `backend/internal/media/ports/transcoder.go` | `Transcoder` — `Probe`, `Render`, `Thumbnail` |
| `backend/internal/media/ports/queue.go` | `TranscodingQueue` — `Enqueue(ctx, assetID string, attempt int) error` |
| `backend/internal/media/application/upload.go` | `MintTicket`, `AppendChunk` (which supplies both completion policies and enqueues only after the terminal commit), `CancelUpload`, `ListCourseVideos`, `Retry` |
| `backend/internal/media/application/transcode.go` | `Process` — the twelve steps of brief §11 over the four ports, returning a `TranscodeOutcome` (§2.1) so the queue adapter can decide the task's fate without the application knowing what a queue is; `Sweep`; `Recover` |
| `backend/internal/media/application/*_test.go` | fake-port tests: every transition, retry exhaustion, malformed source, no-op on `ready`/not-found, duplicate enqueue, enqueue-after-commit failure, cancellation, disk exhaustion, dependency failure |
| `backend/internal/media/adapters/postgres/repository.go` | the mint transaction (lesson→module→course join with `FOR SHARE`, `CanEditLesson`, generation assignment, insert), the guarded state UPDATEs on `(state, claim_token)`, the list and retry reads |
| `backend/internal/media/adapters/postgres/publish.go` | the publish-and-evict transaction, the recovery pass, the sweep candidate queries |
| `backend/internal/media/adapters/postgres/*_integration_test.go` | Testcontainers: two generations, third-evicts-oldest, generation monotonicity, orphan reaping, expired-ticket takeover, ownership refusal, claim-token ABA under recovery |
| `backend/internal/media/adapters/filesystem/storage.go` | key→path resolution inside an owned root, `os.Root` opens, traversal refusal, truncate-and-append at an offset with `fsync`, atomic `rename` publish, recursive purge with ENOENT-is-success, `Free` via `statfs` |
| `backend/internal/media/adapters/filesystem/storage_integration_test.go` | tempdir, no database: traversal, symlink, truncate-then-append, publish atomicity |
| `backend/internal/media/adapters/ffmpeg/transcoder.go` | argv construction from the fixed ladder, ffprobe JSON parsing, `exec.CommandContext` with `WaitDelay` |
| `backend/internal/media/adapters/ffmpeg/transcoder_test.go` | exact-argv assertions including `force_divisible_by=2`; cancellation proven against `/bin/sleep` |
| `backend/internal/media/adapters/asynq/queue.go` | the producer: `asynq.TaskID`, `MaxRetry`, `Timeout`, `Retention`; `ErrTaskIDConflict` treated as success |
| `backend/internal/media/adapters/asynq/handler.go` | the consumer: payload decode, `GetRetryCount`/`GetMaxRetry`, `SkipRetry` for malformed sources, duration and failure metrics |
| `backend/internal/media/adapters/http/handler.go` | the three JSON operations, DTOs, bounds, error mapping, `RegisterAdmin(mux, authenticate, requireSameOrigin)` — the 4a shape verbatim |
| `backend/internal/media/adapters/http/handler_test.go` | contract tests against `openapi.yaml`; the refusal matrix |
| `backend/internal/media/adapters/tus/server.go` | tus 1.0.0 core + `expiration` — `HEAD`, `PATCH`, `DELETE` only |
| `backend/internal/media/adapters/tus/server_test.go` | offset truth, resume, wrong-offset conflict, concurrent PATCH, unknown/expired ticket, over-length chunk, DELETE |
| `backend/sql/queries/media.sql` | the media queries; regenerates `internal/media/adapters/postgres/generated` |

`ports/ownership.go` and `adapters/postgres/ownership.go` **do not exist** — edition 1's
mistake, closed by BLK-2 (D-1).

### 2.1 Port contracts — exact signatures

Added after M04 stopped on contract shape (`work/report-m04.md`, §16). Method *names* alone
could not express the atomicity D-7, D-8, D-10 and D-15 require: three sequences must run
**inside one transaction, between the row lock and the guarded write**, and a port whose
methods are separate calls cannot order them. The fix is the 4b policy-callback idiom
(`identity/ports/administration.go`: `StatusPolicy`, `DeletePolicy`) — the adapter opens the
transaction, takes the lock, hands the locked facts to a closure the application supplied, and
commits or rolls back on what the closure returns.

```go
// ChunkPolicy runs inside AppendChunk's transaction: after the row is locked
// and Upload-Offset is verified equal to received_bytes, and before the
// guarded advance is written. It performs D-2d steps 3-4 over MediaStorage —
// scratch repair, append, fsync — and, when the locked asset is still in
// `created`, the D-15 free-space gate (Free < 3 x DeclaredBytes) and the D-8
// container sniff over the first ContainerSniffBytes bytes. It returns the new
// received-byte count. ErrUnsupportedSource is the single terminal outcome:
// AppendChunk commits the locked `created` row as `failed`, reason
// `unsupported_source`, with received_bytes still zero, then returns that typed
// error. Every other error aborts the transaction and leaves the row exactly as
// it was locked.
type ChunkPolicy func(domain.VideoAsset) (int64, error)

// CompletePolicy runs inside AppendChunk's transaction only when ChunkPolicy
// returns declared_bytes: after the row is locked and the terminal offset is
// known, and before T3 is written. It renames the scratch key onto
// source/{id}/original (D-2d). It must be idempotent — succeeding when the
// source key already holds declared_bytes — which is what lets B31's repair
// path terminate. An error aborts, leaving the row at its previously committed
// offset and state.
type CompletePolicy func(domain.VideoAsset) error

// PublishPolicy runs inside Publish's transaction: after the row is locked and
// (state = 'processing', claim_token = the caller's) is verified, and before
// the guarded UPDATE. Only then may it remove a stale hls/{id} destination and
// rename the working directory in (D-8). A goroutine whose claim was lost
// never reaches this closure, so it touches no bytes at all. An error aborts
// and publishes nothing.
type PublishPolicy func(domain.VideoAsset) error

type AssetRepository interface {
    // Reads.
    Load(ctx context.Context, id string) (domain.VideoAsset, error)          // ErrVideoNotFound when absent or evicted
    ListCourseVideos(ctx context.Context, actor domain.Actor, courseID string) ([]domain.VideoAsset, error)

    // Ticket lifecycle. Mint runs the D-1 sequence in one transaction:
    // FOR SHARE OF courses -> CanEditLesson -> MAX(generation)+1 -> INSERT.
    Mint(ctx context.Context, actor domain.Actor, lessonID string,
        declaredBytes int64, originalName string, expiresAt, now time.Time) (domain.VideoAsset, error)
    AppendChunk(ctx context.Context, id string, offset int64, now time.Time,
        chunkPolicy ChunkPolicy, completePolicy CompletePolicy) (domain.VideoAsset, error)
    Cancel(ctx context.Context, id string, now time.Time) error              // tus DELETE -> evicted_at

    // Job lifecycle. Claim reports disposition explicitly: on a zero-row
    // claim it returns the re-read row and false, so D-10 step 2 branches on
    // one round trip rather than inferring from fields.
    Claim(ctx context.Context, id string, token domain.ClaimToken, now time.Time) (domain.VideoAsset, bool, error)
    Publish(ctx context.Context, id string, token domain.ClaimToken, renditions []domain.Rendition,
        now time.Time, policy PublishPolicy) (published domain.VideoAsset, evicted *domain.VideoAsset, err error)
    Requeue(ctx context.Context, id string, token domain.ClaimToken, now time.Time) error
    Fail(ctx context.Context, id string, from domain.State, token domain.ClaimToken,
        reason domain.FailureReason, now time.Time) error
    Retry(ctx context.Context, actor domain.Actor, id string, claimAge time.Duration, now time.Time) (domain.VideoAsset, error)

    // Sweep and recovery (D-7). SweepCandidates deletes nothing; the caller
    // purges the bytes and only then calls DeleteSwept, so a purge failure
    // leaves a row the next pass retries. Recover applies C5 and C6.
    SweepCandidates(ctx context.Context, now time.Time, limit int) ([]domain.VideoAsset, error)
    DeleteSwept(ctx context.Context, id string) error
    Recover(ctx context.Context, now time.Time, limit int) ([]domain.VideoAsset, error)
}

type MediaStorage interface {
    Free(ctx context.Context) (int64, error)
    Size(ctx context.Context, key domain.StorageKey) (int64, error)
    Head(ctx context.Context, key domain.StorageKey, n int) ([]byte, error)  // the sniff's first bytes
    Append(ctx context.Context, key domain.StorageKey, offset int64, body io.Reader) (int64, error)
    Truncate(ctx context.Context, key domain.StorageKey, size int64) error
    Move(ctx context.Context, from, to domain.StorageKey) error
    Publish(ctx context.Context, work, destination domain.StorageKey) error
    Purge(ctx context.Context, key domain.StorageKey) error
}

// Transcoder and TranscodingQueue are unchanged from M03.

// Job carries what the handler knows without the application importing asynq:
// the adapter fills RetryCount and MaxRetry from asynq.GetRetryCount and
// asynq.GetMaxRetry, so D-10 step 12's "final permitted attempt" is decidable
// inward.
type Job struct {
    AssetID    string
    Attempt    int
    RetryCount int
    MaxRetry   int
}
```

**Where each callback runs, restated as a rule an implementer can check:** every policy is
invoked by the *adapter*, inside `pool.Begin`, after the row's `SELECT … FOR UPDATE` and after
the guard that method exists to enforce, and before that method's `UPDATE`. None of them may be
called by the application directly, and none of them opens its own transaction. The application
supplies the closure and never sees a `pgx.Tx` — which is what keeps `ports` free of a database
type (AGENTS.md: ports reference domain only).

**The application never learns that a queue exists.** `Process` returns a value the adapter
reads:

```go
// TranscodeOutcome is the job's business result. err is reserved for
// dependency faults, which are always retryable.
type TranscodeOutcome struct {
    Failed    bool
    Reason    domain.FailureReason // set when Failed
    Permanent bool                 // == Reason.Permanent(); Q8's never-retry
}

func (t *Transcode) Process(ctx context.Context, job ports.Job) (TranscodeOutcome, error)
```

The Asynq adapter (`adapters/asynq/handler.go`, M12) is the only place that maps it:

| `Process` returns | Handler returns | Asynq does |
|---|---|---|
| `err != nil` | that error | retries with backoff, archives after `MaxRetry` |
| `outcome.Permanent` | `fmt.Errorf("%s: %w", outcome.Reason, asynq.SkipRetry)` | archives immediately, never retries (R15) |
| `outcome.Failed` without `Permanent` | the reason as an error | retries |
| zero outcome | `nil` | marks done |

`domain.FailureReason.Permanent()` is the single inward source of that verdict:
`UnsupportedSource` is permanent, and `ProbeFailed`, `RenderFailed`, `PublishFailed` and
`StorageUnavailable` are not. Adding a permanent reason later is a one-line domain change with a
table test, and no adapter edit.

### Backend — the third executable and shared plumbing

| File | Contents |
|---|---|
| `backend/cmd/upload/main.go` | the third executable: `runtime.Open` → media repository + filesystem storage + asynq queue → `tus.Server` under `/uploads/` on `httpserver.NewHandler`, `:8082`; readiness includes a media-root write probe (D-15) |
| `backend/cmd/media-worker/main.go` (extend) | replaces the empty `asynq.NewServeMux()` with the media handler; adds `ErrorHandler`; runs `Recover` then `Sweep` at start; `/status` reports the handler; readiness gains the same write probe |
| `backend/cmd/api/main.go` (extend) | constructs the media application over the **repository and queue ports only — no `MediaStorage`** — and calls `mediaHandler.RegisterAdmin(mux, identityHandler.Authenticate, identityHandler.RequireSameOrigin)`. The API composition root never wires a storage adapter, which is what keeps "the API touches no bytes and needs no media mount" true in code rather than by convention |
| `backend/platform/observability/metrics.go` (extend) | `courses_transcoding_job_duration_seconds{outcome}` and `courses_transcoding_failures_total{reason}`, **with every label value pre-initialised at registration** (BLK-6) |
| `backend/openapi/openapi.yaml` (extend) | 3 path templates / 3 operations, the media schemas, six new `error.code` values |
| `backend/internal/course/adapters/postgres/repository_integration_test.go` (extend) | add the real `00005_media.sql` to the `fstest.MapFS` (`:195`) and renumber the canary to `00006_failure.sql` (`:200`) — §5.b |
| `backend/Dockerfile` (extend) | build `/out/upload`; new `FROM runtime AS upload` stage (**no ffmpeg**) |

### Frontend

| File | Contents |
|---|---|
| `frontend/src/domain/media/video.ts` | `VideoState`, `LessonVideo`, `VideoError`, the ticket type, the accept list |
| `frontend/src/ports/LessonVideoRepository.ts` | `mint`, `listByCourse`, `retry`, **`cancelUpload`**, and the resumable transfer with progress and cancellation |
| `frontend/src/application/admin/lesson-video.ts` | use cases over the port, returning `Result` |
| `frontend/src/adapters/api/lesson-video-schema.ts` | zod schemas |
| `frontend/src/adapters/api/lesson-video-repository.ts` | Orval calls through `createApiTransport`; business-code mapping |
| `frontend/src/adapters/api/tus-client.ts` | native `fetch`: `HEAD` for the offset, sequential `PATCH`, resume, `DELETE` to cancel, per-chunk deadline, progress (D-2) |
| `frontend/src/app/composition/admin-video-queries.ts` | query options, the bounded poll predicate (D-12), the upload and cancel mutations, `invalidateAfterVideoWrite` |
| `frontend/src/app/composition/admin.ts` (extend) | returns `{...courses, users, videos}` — the 4b pattern, no new router-context key |
| `frontend/src/presentation/admin/LessonVideoField.tsx` | per-lesson: `FileUpload`, `ProgressBar`, the state badge, the failure reason, retry, **the interrupted-upload affordance (resume or cancel)**, refusal and degraded states |
| `frontend/src/presentation/admin/CourseEditorView.tsx` (extend) | mounts `LessonVideoField` inside the existing `LessonFields` fieldset (`:231`) |
| `frontend/src/app/providers/locales/uk.ts`, `en.ts` (extend) | the media vocabulary; keys added only |
| `frontend/e2e/media.setup.ts` | generates the test clip **into `/tmp` inside the worker container** and copies it to `playwright/.fixtures/` (§5.f) |
| `frontend/playwright.config.ts` (extend) | the `setup` project's `testMatch` becomes `['**/auth.setup.ts', '**/media.setup.ts']` |
| `frontend/e2e/media.spec.ts` | the browser journey against Compose |

### Infrastructure and docs

`docker-compose.yml`, `infrastructure/nginx/nginx.conf`, `infrastructure/tests/smoke.sh`,
`README.md`, `docs/adr/008-tus-ingest.md` (new), `docs/adr/009-fmp4-segments.md` (new,
supersedes the segment clause of ADR 005), `docs/media/report.md` (new), `work/tasks.md`
(orchestrator-owned).

**Unchanged on purpose:** `.claude/standards.md` §Машинні гейти (no new gate step),
`Makefile`, `.github/workflows/ci.yml`, Prometheus, `.env.example` (no new secret — D-2b is
why), `backend/sql/seed.sql`, `seed-identities.sql`, `.gitignore` (already fixed in
`5ef6c1f`), and the two `/media` lines in `nginx.conf`.

---

## 3. Reuse-first

| Concern | Reused | Why nothing new |
|---|---|---|
| Server authorization (same-origin → authenticate → role → ownership) | `identityHandler.Authenticate`, `RequireSameOrigin`, `httpserver.Actor`, wired in `cmd/api/main.go:54` | the 4a seam (D-1 there); media is its third consumer and registers exactly as `admin_handler.go:100-110` does |
| The ownership rule and its lock shape | 4a's `lockAuthoringCourse` (`FOR UPDATE`) + `Actor.CanManage`, restated as a locked read + a capability predicate **inside one repository method** | the rule is the product's; the code is not importable without a cross-module domain dependency 4a refused |
| The guarded single-statement transition | 4a D-3f (`UPDATE … WHERE id = $1 AND status = $from`), `authoring.go:55-73` | read-then-act is a race under READ COMMITTED |
| The microsecond clock discipline | `nextAuthoringTimestamp` (`authoring.go:436`), the I03 lesson | `timestamptz` carries no nanoseconds |
| Asynq server, queue option, logger, health, inspector, queue-depth gauge, graceful shutdown | `media-worker/main.go:38-68`, `platform/redis/{client,queue}.go` | proven; Phase 5 fills the mux foundation left empty |
| Executable bootstrap (pool, Redis, metrics, tracing, readiness, `HTTP_ADDR`) | `platform/runtime.Open` + `httpserver.NewHandler`/`Run` | the third executable is wiring, not a new pattern |
| HTTP envelope, request id, JSON and error writers | `platform/httpserver` | already emits the contract's error envelope |
| Contract-test harness (kin-openapi) | `internal/*/adapters/http/handler_test.go` | new paths plug into the existing validator |
| Repository harness (Testcontainers + goose provider) | `course/…/repository_integration_test.go`, `identity/…/schema_integration_test.go` | already boots isolated PostgreSQL and applies migrations |
| The canary-renumber pattern | 4a `fd90fc9`, 4b §5.b | the third time; the procedure is known |
| UI kit: `FileUpload`, `ProgressBar`, `Button`, `ErrorState`, `EmptyState` | `presentation/ui/` | built in foundation for exactly this consumer; §5.g |
| Frontend admin idioms | `admin-course-repository.ts`, `admin-queries.ts`, `admin.ts`, `createApiTransport`, MSW, RHF | the video field is a child of the existing editor, not a new slice |
| e2e shared `storageState` | 4a `auth.setup.ts` (D-7 there) | the media spec adds **zero logins**; the ≤7-per-run budget is untouched |

**What is genuinely new, and why the existing thing did not fit:**

- **`internal/media` as a module** — brief §10 names its terms; the module does not exist.
  Putting video state on the course aggregate would drag a filesystem and a subprocess into
  the catalog module.
- **The `upload` executable** — §2 rule 13 and §24.11 forbid video bytes through the Go API
  and the SSR server; §2 rule 12 gives the shape. Nothing existing terminates a byte stream.
- **Four ports** — PostgreSQL, the filesystem, the ffmpeg subprocesses, Redis/Asynq: each a
  real external dependency, so none is a speculative seam (AGENTS.md).
- **The tus client and its transport** — `createApiTransport` has a 5 s deadline and a
  JSON-shaped `execute` (`transport.ts:26,46-51`); a chunked binary `PATCH` needs a per-chunk
  deadline, an offset protocol and progress. It reuses the transport's *failure vocabulary*
  so the domain error union is unchanged.
- **Two metrics** — §17 requires transcoding duration and failures;
  `platform/observability/metrics.go:15-18` has neither.
- **Two ADRs** — ADR 005 chose `.ts`; GATE 1 Q3 chose fMP4/CMAF `.m4s`, and
  `.claude/standards.md` puts `docs/adr/*.md` in the Мірник, so a contradicting change needs
  a new ADR.

---

## 4. Decisions

### D-1 The module boundary: media reads course rows inside its own transaction

The media module needs one fact the course module owns: **who owns the course this lesson
belongs to**, and it needs that fact *inside the same transaction as the insert*, or the lock
guards nothing.

Edition 1 put it behind a second port (`ports.LessonOwnership`). That was wrong: two ports
have no shared transaction handle, and passing a `pgx.Tx` through a port would break the port
rule (ports reference domain only). 4a solved the identical problem inside **one repository
method** — `Repository.Save` does `lockAuthoringCourse` → `actor.CanManage` → writes, all on
one `tx` (`authoring.go:186-199`).

**Decision: the join lives in the media repository.** `AssetRepository.Mint(ctx, actor,
lessonID, request)` opens one `pool.Begin` transaction and, in this order:

1. `SELECT courses.id, courses.instructor_id FROM course_lessons JOIN course_modules ON …
   JOIN courses ON … WHERE course_lessons.id = $1 FOR SHARE OF courses` — no row →
   `lesson_not_found`.
2. `domain.Actor.CanEditLesson(instructorID)` over the **locked** row — refusal →
   `course_forbidden`.
3. `SELECT COALESCE(MAX(generation), 0) + 1 FROM video_assets WHERE lesson_id = $1` (over
   **all** rows, evicted included — D-5).
4. `INSERT INTO video_assets …` — a `23505` → `upload_in_progress`, **mapped by
   `ConstraintName` and matching either unique index**, `video_assets_pending_per_lesson` or
   `video_assets_generation` (the 4a `slugTaken` shape, `authoring.go:426-429`). Two concurrent
   minters both hold `FOR SHARE` (which is compatible with itself), both compute `MAX+1`, and
   both reach the INSERT; the loser blocks on the winner's uncommitted index entries and errors
   on whichever index PostgreSQL checks first — necessarily the generation index when every
   existing row for the lesson is already `ready`. Keying the mapper on the pending index alone
   would turn an ordinary concurrent mint into `500 internal_error`.

`Retry` and `ListCourseVideos` run the same join (without the lock for the read).

**Lock order, and why it is deadlock-free.** A concurrent 4a `Save` takes `FOR UPDATE` on the
same `courses` row *first* (`lockAuthoringCourse`, `courses.sql:89`) and only then touches
lessons. The mint takes `FOR SHARE` on `courses` *first* and only then inserts. Both
transactions acquire the course row before anything else, so they serialise rather than
deadlock. If the 4a save wins and deletes the lesson, the mint's join finds no row and
answers `lesson_not_found`; if the mint wins, the save's lesson delete cascades
`lesson_id → NULL` and the sweeper reclaims the asset (D-4).

The copied surface is one two-column `SELECT` — *copied, not imported*, the precedent 4b set
for the escaped-`LIKE` replacer (`plan-4b.md` NB2-3). Trigger to revisit: a **third** module
needing the same join.

### D-2 The tus service: shape, authorization and the write protocol

GATE 1 Q1 is binding: resumable tus, terminated in a separate executable (or tusd), the Go
API never receiving video bytes.

**D-2a — our own tus core, not tusd.** tusd would be zero Go code but authorizes through
`pre-create` and `post-finish` HTTP hooks we would have to implement on our API anyway, plus
a new image and tusd's `.info` sidecar layout our storage adapter would then have to
understand. Implementing **tus 1.0.0 core plus `expiration`** — `HEAD` → `Upload-Offset`,
`PATCH` `application/offset+octet-stream`, `DELETE` terminate, `Tus-Resumable: 1.0.0` on
every exchange — is bounded, well-specified HTTP with an obvious test surface, and it keeps
the executable ours: same `runtime.Open`, same `httpserver` envelope, same readiness and
logging rules. **The `creation`, `concatenation`, `checksum` and `creation-defer-length`
extensions are not implemented, and neither is `OPTIONS`** — the JSON mint is the creation,
so a second creation point would only re-open the question of who creates and with what
length (BLK-3). Trigger to swap to tusd: parallel chunks, checksum verification, or an S3
backing store.

**D-2b — authorization is a single-use ticket row.** The brief posed "a signed ticket or a
callback"; both were rejected for a third answer.

- A **signed** ticket (HMAC) needs a shared secret in `.env` and Compose, a new crypto
  surface, and its own expiry/replay semantics — a secret for a problem the database already
  solves.
- A **callback** needs an internal HTTP contract outside `openapi.yaml` (so outside the drift
  gate), a hop per upload, and it puts the caller's *cookie* into the upload service's hands
  — precisely the identity logic Q1 says it must not have.
- **Chosen: the `video_assets` row is the capability.** `POST /api/v1/admin/lessons/{id}/video`
  is an ordinary cookie-authenticated admin operation running the full 4a chain (D-3); on
  success it mints a row with a random v4 UUID id, the declared length and `expires_at`, and
  returns `/uploads/{videoId}`. **The upload service never reads a cookie, never learns a
  role, never knows what a session is.** It validates existence, state and expiry.

*The ticket id is the public asset id.* `GET …/videos` returns `videoId` to every admin and
owning instructor of that course, so a co-editor can append to or cancel another editor's
in-flight upload. That is the same authorization scope that could have minted it, so there is
no escalation — but the property that protects the upload is the **ownership rule at mint**,
not the id's unguessability. Stated so it is a decision rather than an assumption.

*Deviation from Q1's wording, named.* Q1 and §24.11 say the upload service "notifies the API
on completion". It does not: it writes `video_assets` and enqueues directly, through the
shared module's ports, which §2 rule 12 explicitly permits ("окремий executable, але
використовує спільні backend modules і ports"). The API still sees only metadata and still
never receives a byte, so the rule the wording protects holds; but this makes the upload
service a second writer of media state and a second queue producer, and that is a decision,
not a reading. Trigger to revert to a literal notification: a **second** ingest terminator
(a mobile SDK, a third-party importer), at which point an internal completion endpoint
becomes the single place the state machine advances.

**D-2c — the tus surface is deliberately absent from `openapi.yaml`.** tus is a
header-driven binary protocol; describing it would make Orval generate a client we do not use
and put binary-body operations under the drift gate for no benefit. It is specified in
ADR 008 and exercised by `adapters/tus/server_test.go` and `e2e/media.spec.ts`.

**D-2d — the PATCH write protocol (BLK-4).** Spec criterion 1 ("the source is never lost on a
partial or interrupted upload") depends on stating exactly whose offset is the truth.

> **`video_assets.received_bytes` is the truth. The file is reconciled to it, never the
> other way round.**

Each `PATCH`, in this order:

1. Open a `pool.Begin` transaction and `SELECT … FOR UPDATE` the asset row. Missing, evicted,
   expired, or in a state other than `created`/`uploading` → `404`, no write.
2. Compare the request's `Upload-Offset` with `received_bytes`. Unequal → `409`, **no write**.
3. Ask storage for the scratch file's length. **Longer than `received_bytes`** — a crash after
   the append and before the commit — **truncate it back to `received_bytes`**. **Missing
   entirely, with `source/{id}/original` present at `declared_bytes`** — a crash between the
   completion rename and its commit (D-2d's terminal branch) — **skip the append and run the
   terminal UPDATE to `queued`**: the bytes are all there, one directory further on, and
   without this branch every later `PATCH` would answer `503` forever and the client's only
   exit would be `DELETE` plus a fresh 2 GiB upload, which is exactly the "source is never
   lost" promise of criterion 1. Missing with no source, or shorter than the column, is a real
   storage fault → `503`.
4. Append the body through `http.MaxBytesReader(w, r.Body, chunkBytes)`, then `fsync`.
   **The cap is enforced here, not only at the edge**: nginx checks `client_max_body_size`
   against `Content-Length`, and with `proxy_request_buffering off` a
   `Transfer-Encoding: chunked` request carries none, so it would stream straight past the
   edge limit. Our own client always sends a length, so this guards the case a hand-rolled or
   hostile client creates rather than the happy path. Overflow answers `413`, the file is
   truncated back, and the offset does not move. A write failure is handled the same way and
   answers `507` (out of space) or `503` (anything else), so the client resumes from the same
   place.
5. One guarded statement:
   `UPDATE video_assets SET received_bytes = $new, state = 'uploading', updated_at = $clock
    WHERE id = $1 AND received_bytes = $offset AND state IN ('created','uploading')
      AND expires_at > $clock AND evicted_at IS NULL`.
   Zero rows → another writer moved first or the ticket died: truncate back to the stored
   value and answer `409`. Commit.

Consequences, each with an acceptance row: two concurrent `PATCH`es at the same offset
serialise on the row lock, so exactly one gets `204` and the other `409` with the file length
still equal to `received_bytes` (R29); a crash mid-chunk is repaired by step 3 on the next
attempt (B03); a wrong declared offset never writes a byte (R11).

Completion is the same transaction's terminal branch: when `received_bytes` reaches
`declared_bytes`, the source is renamed from `temporary/{videoId}-upload` into
`source/{videoId}/original` **before** the commit that writes `queued`, and the enqueue
happens **after** the commit (D-10, BLK-7).

`DELETE /uploads/{videoId}` sets `evicted_at` on a `created`/`uploading` row and answers
`204`; a following `PATCH` answers `404`, and a new mint for that lesson succeeds because the
pending index counts only non-evicted rows (R30). This is what makes an abandoned upload
recoverable without waiting 24 hours (D-13).

### D-3 Authorization: a capability, not a third copy of the role vocabulary

4a D-1 point 4 recorded the trigger: *"when a third module needs roles, promote them to a
shared kernel rather than duplicating twice more."* Phase 5 is the third module.

**Decision: the trigger does not fire, because media never names a role.** The domain models
the capability, matching §24.10's additive-capability model:

```go
type Actor struct { UserID string; Administrator bool }
func (a Actor) CanEditLesson(ownerID string) error   // admin ⇒ ok; else UserID == ownerID
```

The single string comparison `httpserver.Actor.Role == "admin"` happens once, in the media
**HTTP adapter**, where transport is translated. No new `student`/`instructor`/`admin`
constant set is created, so the duplication the trigger guards against does not occur, and no
refactor is imposed on 4a/4b's proven code.

**Order of checks on every media mutation**, identical to 4a NB-8 and 4b D-1: same-origin
first (an anonymous cross-site `POST` answers `403 origin_forbidden`, not `401`), then
authenticated, then role (`instructor` or `admin`, else `403 forbidden`), then the lesson or
video exists (`404`), then ownership (`403 course_forbidden`), then the domain rule. Every
mutation acceptance row states a valid `Origin`.

### D-4 Where the lesson↔video link lives, and why the FK is `SET NULL`

Q6/§24.13 puts **two generations per lesson**, so a column on `course_lessons` cannot hold
the link. It is a table, `video_assets`, which also keeps every media column out of the course
schema and out of 4a's aggregate.

The referential action is the sharp part: `authoring.go:237` deletes every lesson a 4a `PUT`
omits, and the editor's delete button does exactly that (`CourseEditorView.tsx:275`).

| Option | Consequence |
|---|---|
| `ON DELETE RESTRICT` | a 4a `PUT` removing a lesson with video fails with `23503` — a Phase-5 feature breaking a Phase-4 operation. Rejected |
| `ON DELETE CASCADE` | the row vanishes and **the bytes stay on disk forever**, unreferenced. Rejected |
| **`ON DELETE SET NULL`** | the asset survives as an explicit orphan; the sweeper finds `lesson_id IS NULL`, deletes the bytes, then the row. **Chosen** |

Recorded meaning: **`lesson_id IS NULL` means the lesson was deleted and this asset is garbage
awaiting the sweeper**; every read filters it out, and the handler skips the transcode for one
(D-10). No trigger, no cascade, no leak.

### D-5 Generations: assignment, the two-ready rule, and the eviction point

- **`generation = COALESCE(MAX(generation), 0) + 1` over *all* rows of the lesson, evicted
  ones included**, computed inside the mint transaction and serialised by the pending index,
  with `UNIQUE (lesson_id, generation)` in the schema. Counting only live rows would restart
  at 1 after a full eviction and make "the highest generation" ambiguous. NULL `lesson_id`
  rows do not collide under a unique index, so orphans are unaffected.
- **At most one non-terminal asset per lesson**, enforced structurally:
  `CREATE UNIQUE INDEX video_assets_pending_per_lesson ON video_assets (lesson_id)
   WHERE lesson_id IS NOT NULL AND evicted_at IS NULL
     AND state IN ('created','uploading','queued','processing')`.
  A second mint is `409 upload_in_progress`, mapped from `23505` — never from a pre-check two
  concurrent minters would both pass (the 4a `slug_taken` lesson). **Takeover:** if the
  blocking row is `created`/`uploading` and its `expires_at` has passed, the mint stamps
  `evicted_at` on it in the same transaction and proceeds. A `queued` or `processing` row is
  **not** taken over — a transcode in flight is genuine work, and `expires_at` governs only
  the upload window (D-15).
- **At most two `ready` generations per lesson.** An index cannot count to two, so it is a
  transactional rule at the only place a third can appear — publish. The publish transaction
  locks the lesson's **non-evicted** assets `FOR UPDATE`, flips the new one to `ready`, and if
  more than two `ready` rows now exist, stamps `evicted_at` on the **lowest generation** among
  them. Non-evicted is the count, and the lowest generation is the victim: both stated so the
  implementer does not choose.
- **Bytes are deleted after the commit, never inside it** — a `RemoveAll` cannot be rolled
  back. The window between commit and unlink is covered by D-7. A crash there costs disk,
  never correctness, and never an asset reading `ready` with no bytes.

### D-6 Eviction is a timestamp, not a seventh state

Brief §10 fixes the vocabulary at six states and this plan does not extend it. Eviction is
`evicted_at timestamptz` — the 4b `deleted_at`/`disabled_at` idiom, projection in the domain,
column as truth. An evicted asset keeps whatever state it reached and is invisible to every
read. Three producers: the tus `DELETE`, the third-generation eviction, and the mint that
takes over an expired ticket. One consumer: the sweeper.

### D-7 The sweeper and the recovery pass

One function with two candidate families, bounded to 50 rows per invocation.

**Reclaim (delete bytes, then the row).** ENOENT from any unlink is **success**, not an
error — a previous pass may have died between the bytes and the row.

| Class | Predicate |
|---|---|
| C1 | `evicted_at IS NOT NULL` |
| C2 | `lesson_id IS NULL` (orphaned by a lesson delete, D-4) |
| C3 | `state IN ('created','uploading') AND expires_at < $now` (abandoned upload window) |
| C4 | `state = 'failed' AND updated_at < $now - 30 days` (a failed asset keeps its source ≤ 2 GiB until an operator retries or supersedes it; 30 days is the ceiling) |

Keys removed per row: `source/{id}`, `hls/{id}`, `thumbnails/{id}.jpg` and the glob
`temporary/{id}-*` — which covers both the upload scratch key `temporary/{id}-upload` and
every render working directory `temporary/{id}-{claimToken}`, so no scratch path escapes the
sweep.

**Recover (re-queue, delete nothing).**

| Class | Predicate | Action |
|---|---|---|
| C5 | `state = 'processing'` — **no age predicate at all** | `state = 'queued'`, `claim_token = NULL` (T7) |
| C6 | `state = 'queued' AND updated_at < $now - 10 min` | **increment `attempt`**, then enqueue under the new `assetId:attempt` id (the enqueue-after-commit window, BLK-7) |

**C6 must bump `attempt`, or it enqueues nothing.** Asynq keeps a task id after the task ends
— `enqueueCmd` refuses any id whose key exists (`internal/rdb/rdb.go:98-108`), completion keeps
the key for its 24 h retention (`:382-392`), and archival keeps it for the archive's lifetime
(`:861-885`). Two real paths leave a `queued` row beside an archived task of the same id: the
30-minute `Timeout` firing on the final attempt (asynq archives on `ctx.Done()` without waiting
for the handler, `processor.go:243-245`, while the handler's `Requeue` writes `queued`), and
three lease recoveries, each of which increments `Retried` via `rdb.Retry`
(`recoverer.go:96-102`). In both, a re-enqueue under the *same* id returns
`ErrTaskIDConflict`, which the adapter maps to success (D-10) — so edition 3.1's C6 would have
reported recovery while producing no delivery at all, at every restart. Incrementing `attempt`
first, exactly as T9 does, gives it a fresh id and a real task.

**What that costs, named rather than solved: C6 may enqueue a duplicate.** The predicate cannot
distinguish "committed, enqueue failed" (the BLK-7 window it exists for) from "enqueued fine,
but the worker was down or busy for ten minutes" — a backlog of `queued` rows whose tasks are
still *pending* in Redis. For those, C6 adds `assetId:n+1` beside a pending `assetId:n`. The
duplicate is a no-op by construction: with `Concurrency: 1` it can only be delivered after the
first one finished, and step 1/2's exhaustive nil list absorbs it — `ready` or `failed` → `nil`,
a legitimately requeued row → an ordinary extra attempt. Edition 3.1's same-id form had the
mirror-image trade (perfect on a backlog, silent on an archived task), and a delivery that
happens twice is strictly better than one that never happens. **Trigger to spend a second Redis
call on it:** if duplicates become noisy in the queue-depth gauge, C6 consults
`Inspector.GetTaskInfo("media", assetId:attempt)` and skips rows whose task is pending,
scheduled or retrying.

**C5 is threshold-free, and that rests on a stated assumption: there is exactly one worker
process.** `media-worker` runs `Concurrency: 1` with a single replica
(`media-worker/main.go:45`, `docker-compose.yml`), so at *this* process's start no other
process can be holding a claim — every `processing` row is by definition abandoned, whatever
its age. (An age threshold here was the BLK-1 hole; §13 records it.) **Trigger to reintroduce
an age predicate: a second worker replica or `Concurrency > 1`** — at which point "abandoned" stops
being inferable from process start and needs a real lease heartbeat.

**Ordering guarantee.** `Recover` runs to completion **before** `server.Start`, so no
delivery can race the pass that is freeing the rows. The worker's readiness stays false until
both have run.

**Call sites.** C1–C4 run at worker start **and** as step 0 of every job. C5–C6 run **at
worker start only**: with one claimant, a running worker is already draining its own queue.
Between restarts, the operator's retry is the exit for a stranded row — which is why `Retry`
accepts `queued` and stale `processing` (D-10, T9). Named rather than solved with a scheduler.

Not chosen: an `asynq.Scheduler` periodic task — a second scheduling concept for work that
has nothing to do when nothing is uploading. Trigger: media that must be reclaimed on a
deadline with no uploads happening.

A row-less file (bytes on the volume with no `video_assets` row) is **not** reachable by this
design, and no pass looks for one. The two ways one could appear are closed at the source: the
e2e fixture is generated into `/tmp` inside the worker container rather than into
`/media/temporary` (§5.f), and every scratch key is derived from an asset id.

### D-8 The storage layout and the atomic publish

Exactly brief §11, with the unshaped directories given a shape and the **verified** fMP4
naming recorded:

```
/media/
├── source/{videoId}/original
├── hls/{videoId}/
│   ├── master.m3u8
│   ├── audio/{index.m3u8, init_N.mp4, segment_00000.m4s, …}   ← one shared AAC rendition
│   ├── 360p/{index.m3u8, init_N.mp4, segment_00000.m4s, …}
│   ├── 720p/{…}                               ← only if the source supports it
│   └── 1080p/{…}                              ← only if the source supports it
├── thumbnails/{videoId}.jpg
└── temporary/
    ├── {videoId}-upload                       ← the tus scratch file
    └── {videoId}-{claimToken}/                ← one render working directory per claim
```

**Audio is one rendition shared by every video variant**, not a copy per rung — Apple HLS
Authoring 9.5, "SHOULD deliver video and audio as separate streams"
(`research-codecs.md` §2.2 [S7]). Besides conformance it removes N−1 copies of the same
audio from the volume.

**The init segment is `init_N.mp4`, not `init.mp4`** — ffmpeg appends the variant index to
`-hls_fmp4_init_filename` for multi-rendition output, verified **inside
`courses-media-worker-1` on ffmpeg 5.1.9, the binary that ships** (`audio/init_0.mp4`,
`360p/init_1.mp4`, `720p/init_2.mp4`), and identically on a host 7.1.1. **The
index follows `var_stream_map` order, so adding the audio group shifts it** — which is
exactly why no acceptance row and no Phase 6 consumer may hard-code the number: the rows
below assert *one* `init_*.mp4` per variant directory, and ADR 009 records that a consumer
must read the variant playlist rather than construct the name.

`MediaStorage` takes an opaque `StorageKey{Kind, ID, Leaf}` — never a path (ADR 004). `Kind`
is one of four constants, `ID` must parse as a UUID, `Leaf` is drawn from a fixed set. **A
client-supplied filename never reaches the filesystem**: it is a display column. Opens use
`os.Root` so a symlink planted in the volume cannot escape the root; `..` cannot appear
because no key component is free text.

**Publication is atomic because the destination is new**, not because of a swap: `{videoId}`
is minted per upload, so `hls/{videoId}` never exists before its one publish, and a
same-filesystem `rename(2)` of the working directory onto it is atomic. Two branches, stated
rather than assumed:

- destination exists **and** the asset is `ready` → the job is a no-op (D-10);
- destination exists **and** the asset is not `ready` → a previous attempt published and died
  before its commit; the adapter removes the stale directory and re-publishes.

**The claim is verified under the row lock before anything under `hls/` is touched.** The
publish transaction opens, takes `SELECT … FOR UPDATE` on the asset and checks
`state='processing' AND claim_token=$token` **first**; only then may it remove a stale
destination and rename. A goroutine whose claim was lost therefore touches no bytes at all: it
releases, removes its own working directory and returns. Without that ordering there is a real
if narrow double fault — a stale goroutine A, re-queued and superseded by claim B, reaches
publish while B is between its rename and its commit, removes B's directory, renames its own
in, then loses the guarded UPDATE; if A dies between the `RemoveAll` and its rename, B commits
`ready` over nothing. The window is milliseconds and needs two failures on a single-worker
deployment, but "publication is atomic" is a headline claim, so the ordering is stated rather
than left to the implementer.

The thumbnail is renamed into `thumbnails/` after the playlist directory lands.

**Sniff on the first chunk (Q2), as a named domain rule.** The container check runs in the
**first `PATCH`**, on the twelve bytes already in hand, and it is a business rule, so it lives
in the domain rather than in the upload adapter:

```go
// ContainerSniffBytes is how many leading bytes AcceptContainer needs: an
// ISO-BMFF box header is 8 bytes and its major brand occupies bytes 8-11.
const ContainerSniffBytes = 12

// AcceptContainer applies the Q2 container half to the head of a source.
// ErrUnsupportedSource for anything it does not recognise.
func AcceptContainer(head []byte) error
```

It accepts exactly two shapes. **ISO-BMFF**: bytes 4–7 are `ftyp` and the major brand at bytes
8–11 is one of `isom`, `iso2`, `iso4`, `iso5`, `iso6`, `mp41`, `mp42`, `avc1`, `mmp4`, `dash`,
`M4V ` (MP4 family) or `qt  ` (QuickTime `.mov`). **Matroska/WebM**: the EBML magic
`1A 45 DF A3` at offset 0. A head shorter than `ContainerSniffBytes` is a refusal, not a pass.

The rule is deliberately **permissive and cheap**, not authoritative: the EBML magic covers
`.webm` as well as `.mkv`, and a VP9 WebM will be refused later by `AcceptSource` in the worker.
That division is Q2 as amended (`spec.md:67`) — the container at the edge, the codecs where
ffprobe lives — and it is why the sniff never tries to parse a DocType out of bytes it does not
have. **M03a owns the rule and R43; M11 owns its use in the first `PATCH` and B18.** A failure consumes the ticket — `state = 'failed'`, reason `unsupported_source`,
`400 unsupported_media` — writes nothing under `source/`, and **never enqueues**. Edition 1
sniffed at completion, which refused a renamed text file only after up to 2 GiB had been
transferred; the same bytes are available in the first request.

**The split of Q2 is settled, not proposed.** The *codec* half (H.264/HEVC video, AAC/MP3
audio) cannot be checked at the edge: ffprobe lives in the worker image, and putting ffmpeg
into the byte-receiving service would double its image and give it subprocess execution. The
codec check is therefore the worker's first step and produces `failed` with `SkipRetry`, never
a retry — so a wrongly-coded file becomes a `failed` badge shortly after upload rather than a
`400` at mint. Edition 2 raised this as a proposal; **the gate owner accepted it on 2026-09-06
and spec Q2 is amended accordingly (`docs/media/spec.md:67`)**. It is a settled decision here,
not an open question.

**A failed sniff is `created → failed` (T8).** The first `PATCH` finds the row in `created`,
so the refusal writes `failed` from `created`, not from `uploading` — the matrix and
`domain/asset.go`'s transition table both list it, or B18 would fail against a table
implemented literally from §1.

### D-9 Transactions and isolation

**Every media write uses `pool.Begin` (READ COMMITTED)**, matching identity's `Rotate` and the
course module's *writes* (`authoring.go:147,186,274,317`). It does **not** copy
`authoring.go:76,124`'s `BeginTx(RepeatableRead, ReadOnly)`: those are the consistent *reads*,
and 4b already recorded what copying them into a write path costs — a `40001` that
`persistenceError` maps to `503`, turning a business outcome into a dependency outage
(`plan-4b.md` D-2). Every state flip is a single guarded `UPDATE`, and every flip a claim
makes carries `AND claim_token = $token` (D-10). Zero rows affected is disambiguated by a
re-read **inside the same transaction**, the 4a D-3f shape. Every write stamps `updated_at`
with the application clock **truncated to microseconds** (the I03 lesson).

### D-10 The job payload, the claim token, idempotency and retry

**Payload:** `{"assetId": "<uuid>", "attempt": <int>}` — nothing else. No paths, no filenames,
no actor. Enqueued as `asynq.TaskID(assetId + ":" + attempt)`, `asynq.Queue("media")`,
`asynq.MaxRetry(3)`, `asynq.Timeout(30 * time.Minute)`, `asynq.Retention(24 * time.Hour)`.

**The claim token closes the re-delivery race (BLK-1).** `asynq@v0.26.0/processor.go:245`
re-delivers the same payload on lease expiry while the previous goroutine may still be
running, so "guard on `state = 'processing'`" is an ABA: the recovery pass or a later claim
can move the row back through `queued` to `processing`, and the stale goroutine's publish
would find a row in the expected state and overwrite a newer claim's work. Therefore:

- `Claim` is `UPDATE … SET state='processing', claim_token = $fresh, updated_at = $clock
  WHERE id = $1 AND state = 'queued' AND evicted_at IS NULL` and returns the token.
- **Every subsequent write by that goroutine — publish, requeue, fail — carries
  `AND state='processing' AND claim_token = $token`.** Zero rows means the claim was lost:
  the goroutine removes its working directory and returns `nil`. Not an error, because there
  is nothing to retry.
- **The working directory is `temporary/{videoId}-{claimToken}`**, so two attempts can never
  share a byte of scratch — and, belt and braces, the handler `RemoveAll`s its own directory
  before rendering into it. Edition 1 keyed it by `attempt`, which Asynq's automatic retries
  do not increment, so every retry shared one directory with a possibly-live predecessor.

**The handler's sequence** (brief §11's twelve steps, with the guards made explicit):

0. Sweep classes C1–C4 (D-7).
1. Load the asset. **The nil list is exhaustive**, so no state can fall through it: not found
   → `nil` (a concurrent sweep removed it — NB-10); `ready`, `failed`, `evicted_at IS NOT NULL`
   → `nil`, the work is finished or abandoned; `lesson_id IS NULL` → `nil` (an orphan's publish
   is work the sweeper would undo); `created` or `uploading` → `nil` (the source never landed,
   so a task naming this asset is a duplicate from C6 — N-31). Only `queued` and `processing`
   continue.
2. `Claim` (guarded on `state='queued'`). Zero rows: **re-read the row and branch on why**,
   over the same exhaustive list. `ready`, `failed`, evicted, absent, `created`, `uploading` →
   `nil`, the work is done, abandoned, or never arrived. **Still `processing` → return a
   retryable error, never `nil`**, because `nil` makes Asynq mark the task done and strands a
   row nobody owns (§13, BLK-1). The retryable error keeps the task alive across Asynq's
   backoff — `n⁴ + 15 + rand(30)·(n+1)` seconds, so retries 1–3 wait **15–44 s, 16–74 s and
   31–118 s** (`server.go:401-405`) — which is ample for the lease-expiry case where the old
   goroutine writes T6 within seconds, and harmless after a restart because T7 has already
   freed the row before `server.Start` (D-7). **Each lease recovery also consumes one of the
   three retries** (`recoverer.go:96-102` calls `rdb.Retry`, which increments `Retried`), so a
   repeatedly-crashing job reaches `failed` rather than looping forever.
3. Free-space precondition (D-15). Insufficient → requeue (T6) and return a retryable error
   with reason `storage_unavailable`.
4. `Probe`. Refusal → `Fail` with `unsupported_source`, return an error wrapping
   `asynq.SkipRetry` (verified at `processor.go:329,343`) so it is archived, never retried.
5–9. Working directory, render, playlists, thumbnail (D-11).
10. Publish and evict in one transaction (D-5, D-8): lock the row, **verify the claim token
    before touching `hls/`**, then remove any stale destination, rename, and commit.
11. Remove the working directory.
12. On a retryable failure: `Requeue` (T6) **before** returning the error, so the row is
    `queued` and the next delivery can claim it. On the **final permitted attempt**
    (`asynq.GetRetryCount` ≥ `asynq.GetMaxRetry`, verified at `context.go:25,33`): `Fail` with
    a reason from the closed set, then return the error. On context cancellation: `Requeue`,
    remove the working directory, return the cancellation — **except when the cancellation is
    the task timeout on the final permitted attempt, which writes `Fail` with
    `render_failed`.** The branch keys on the **cause**,
    `errors.Is(ctx.Err(), context.DeadlineExceeded)`, not on cancellation as such: a source
    that cannot finish inside the 30-minute `asynq.Timeout` would otherwise rest `queued`
    forever with its task already archived, a stall dressed as a pending job. A plain
    `context.Canceled` — which is what a graceful restart delivers, after asynq has already
    requeued the task to *pending* — keeps `Requeue`. Without that distinction a restart that
    happened to land on a final-attempt render would write `failed` for a source nothing was
    wrong with, and D-15's promise that this design defers a terminal state rather than
    guessing one would be false in exactly one corner.

**Retry (Q8).** `POST /admin/videos/{id}/retry` clears `claim_token`, increments `attempt`,
sets `queued` and enqueues under the new task id. It accepts three states, each for a stated
reason, and refuses everything else with `409 video_not_retryable`:

- **`failed`** — the ordinary operator retry.
- **`queued`** — the commit-then-enqueue window (BLK-7): an asset can legitimately be `queued`
  with no task behind it, and refusing would leave the operator no exit but SQL.
- **`processing` whose `updated_at` is older than the claim age** (`taskTimeout` + 10 min =
  40 min) — the D-15 "PostgreSQL down when the terminal state must be written" path, where the
  worker *keeps running*, so T7 never fires and no restart is coming. Taking the claim from a
  genuinely live render is impossible in practice (`asynq.Timeout(30 min)` bounds it) and
  harmless in principle: that goroutine's publish is guarded on `claim_token`, finds zero rows,
  cleans up its working directory and returns.

**In every case retry also requires that the source actually landed**
(`received_bytes = declared_bytes`). A `failed` row written by the upload service's sniff
refusal has `received_bytes = 0` and nothing under `source/{id}/original`; re-queueing it would
buy a probe against a missing file and a second `failed` with a different reason. That row's
exit is a fresh upload, not a retry.

**Duplicate enqueue.** `asynq.ErrTaskIDConflict` (verified at `client.go:228,416`) is mapped
to **success** by the queue adapter — the task the caller wanted already exists. This is what
makes the retry-a-`queued`-asset path idempotent.

**Cancellation (§11 step 12), described as Asynq actually behaves** — not as "shutdown cancels
the handler's context", which is what edition 2 assumed (§13, BLK-1). `Server.Shutdown`
(`server.go:724-751`) calls `processor.shutdown()` (`processor.go:139-149`), which waits
`ShutdownTimeout` — **10 s** as configured at `media-worker/main.go:45` — with the handler's
context **still live**; on expiry the worker goroutine requeues the message to *pending*
(`processor.go:236-239, 260-272`) and returns, and only the deferred `cancel()`
(`processor.go:207-210`) then cancels the handler's context. `srv.wg.Wait()` does not include
the handler goroutine, so `main` proceeds to `dependencies.Close()` and exits. Consequences,
taken rather than wished away:

- **T6 is best-effort on shutdown.** A render longer than 10 s will not get its `Requeue`
  written, because by the time the context is cancelled the pool is closing. The plan does not
  raise `ShutdownTimeout` to paper over this: a shutdown that waits out a 30-minute transcode
  is worse than one that abandons it.
- **Correctness therefore rests on T7, not on T6** — which is exactly why C5 is threshold-free
  (D-7) and why step 2 returns a retryable error instead of `nil`. The requeued-to-pending task
  is delivered on restart *after* `Recover` has already set the row `queued`, so it claims
  normally. Compose makes the requeue itself reliable: `stop_grace_period: 40s`
  (`docker-compose.yml:7-10`) is comfortably longer than the 10 s `ShutdownTimeout`, so asynq
  finishes requeueing before Docker's SIGKILL. That precondition was previously inherited by
  luck; it is stated here because lowering either number would break the path silently.
- When the context *is* cancelled in time, `exec.CommandContext`'s default `Cancel` sends
  SIGKILL and `WaitDelay` bounds the wait. No `Setpgid` and no process-group signalling:
  ffmpeg is single-process (threads, not children), and `SysProcAttr` is Unix-only, so it would
  need a build tag or a stub to keep `make test` compiling everywhere.
- R18 (cancel mid-job → `queued`) exercises this against a fake, where the window is
  guaranteed; **R35 exercises the real server** by restarting the worker mid-render in Compose.

### D-11 The ffmpeg invocation — no user-controlled arguments, and no odd dimensions

Brief §11: *"Не передавай довільні аргументи користувача в FFmpeg command."* Made structural:
**the argv is built from three sources and no others.**

1. **Compile-time constants** — every flag and every profile value, from the fixed ladder
   table below.
2. **Paths derived from the asset UUID and the claim token** by the storage adapter — never
   from a client string. The original filename is display metadata in a column and reaches
   neither argv nor the filesystem.
3. **Integers from ffprobe**, used *only to select a subset of the ladder* by comparison —
   never interpolated into a string. The frame rate is deliberately **not** interpolated
   either: the 2-second IDR cadence is expressed as a frame-rate-independent constant
   expression (below) rather than as a computed GOP length, so rule 1 keeps covering it.

**The ladder, anchored to the Apple HLS Authoring Specification.** The numbers are not
invented; each is cited to `docs/media/research-codecs.md`, which collected them from Apple's
published table [S7] over 37 sources.

| Rung | Resolution | Bitrate ceiling | Source |
|---|---|---|---|
| `360p` | 640×360 | 365 kbit/s | Apple 1.25 ladder table, `research-codecs.md` §1/§2.2 [S7] |
| `720p` | 1280×720 | 3000 kbit/s (Apple gives 3000 and 4500) | same |
| `1080p` | 1920×1080 | 6000 kbit/s (Apple gives 6000 and 7800) | same |

Encoding is **capped CRF, not fixed CBR**: `-crf 21 -maxrate <ceiling> -bufsize <2×ceiling>`
per rung. That is the research's own recommendation (`research-codecs.md` §1): Netflix's 2015
per-title work showed a fixed ladder wastes bits on simple content and starves complex
content, and capped CRF buys most of that effect with no convex-hull search. **The ceiling is
a rate-control input, not a number this plan asserts against a playlist** — see the
conformance paragraph below: the target binary writes no measured bandwidth at all, and where
one is measured it lands slightly *above* the ceiling because VBV caps the elementary stream
while the fMP4 container and the 2 s IDR cadence add overhead. Edition 3.1 claimed the
opposite and compared against it; that comparison was red on both binaries.

**Codec constraints, all from Apple's MUST/SHOULD list** (`research-codecs.md` §2.2 [S7]):

- **H.264 High Profile, 8-bit 4:2:0** — Apple 1.4 ("SHOULD use High Profile"). In argv:
  `-profile:v high -pix_fmt yuv420p`.
- **`-level` is set on the 360p and 720p rungs only, never on 1080p.** Apple 1.3a asks that
  *some* variants be ≤ High 4.1, not all, and level 4.1 caps 1080p at about 30 fps: a 1080p60
  screen recording would be labelled 4.1 while exceeding it, which is worse for a strict
  decoder than carrying no level at all. libx264 warns and encodes anyway, so the defect would
  be silent. Dropping one flag on one rung is cheaper than a frame-rate cap filter, and the two
  lower rungs satisfy 1.3a on their own.
- **`avc1`, not `avc3`** — Apple 1.10 wants parameter sets in the sample description, which is
  what the fMP4 muxer writes by default with a global header; stated so a later `-flags
  +global_header` removal is recognised as a regression.
- **AAC-LC, stereo** — Apple 2.3. In argv: `-c:a aac -profile:a aac_low -ac 2 -b:a 128k`.
  (Apple 2.4 warns off HE-AAC above 64 kbit/s, which is why plain AAC-LC at 128k rather than
  HE-AAC.)
- **6-second target segments, IDR every 2 seconds, `EXT-X-INDEPENDENT-SEGMENTS`** — Apple 7.5,
  1.13, 7.4, 9.11/9.12. In argv: `-hls_time 6`, `-hls_flags independent_segments`, and
  `-force_key_frames:v:N expr:gte(t,n_forced*2)` per video stream. The expression is a
  constant string that holds at any frame rate, which is why no probed fps enters the argv.
- **Video and audio as separate streams** — Apple 9.5, expressed through an `agroup` in
  `-var_stream_map` so one AAC rendition serves every video variant (D-8).

**The target binary is the one the image ships: ffmpeg 5.1.9 from `debian:bookworm-slim`.**
This is a decision, not an inheritance, and it is the correction N-17 forced. Editions 2–3.1
measured a host 7.1.1 and wrote its output into this plan; the container runs 5.1.9, and the
two differ in exactly the place B26 asserts. The alternative — a `debian:trixie-slim` worker
stage for ffmpeg 7.1.x — was weighed and **declined**:

- **Nothing in Phase 5 reads a playlist.** The edge refuses `/media/*`, no player exists, and
  §11's own rule (`duration_seconds` dropped, no playlist writer built) is "no consumer in this
  phase → not built". Bumping a base image to improve an attribute nobody reads is that rule
  broken for our own convenience.
- **It would split the image.** `Dockerfile:12` is one `runtime` stage shared by `api`,
  `migrate` and `worker`; a trixie worker means two base images, two apt states and two sets of
  `curl`/CA versions to keep patched, for one metadata field.
- **The bytes are not in question.** On 5.1.9 the ladder, the `min()` filter,
  `force_divisible_by=2`, fMP4, the init naming and the codec strings (`avc1.640029` /
  `avc1.640028`) are all as designed and were executed in the container. Only playlist metadata
  differs.

**Trigger to move to trixie (or to add the line-append writer): Phase 6's player needing
measured per-variant bandwidths for ABR.** That is a real consumer, it arrives with the phase
that can test it, and it owns the change — not Phase 5.

**What 5.1.9 writes, executed in `courses-media-worker-1`:**

| Attribute | Apple | ffmpeg 5.1.9 in the worker image |
|---|---|---|
| `CODECS`, `RESOLUTION` | 9.1, 9.2 MUST | present on every `EXT-X-STREAM-INF` |
| `BANDWIDTH` | 9.13 (peak) | present, but **nominal** — derived from the rate-control flags, `542300` for the 360p rung on a 320×240, a 640×360 and a 640×346 source alike |
| `AVERAGE-BANDWIDTH` | 9.14 MUST | **not written at all** (7.1.1 does write it, measured) |
| `FRAME-RATE` | 9.15 MUST | **not written anywhere** (7.1.1 the same) |
| `EXT-X-MEDIA:TYPE=AUDIO` group | 9.5 | present, referenced by every variant; **no `CHANNELS`** (7.1.1 writes it) |
| `EXT-X-INDEPENDENT-SEGMENTS` | 9.11/9.12 | in each **video variant** playlist only — not the master, and **not `audio/index.m3u8`**: 5.1.9 writes it solely into playlists carrying a video stream. Valid per RFC 8216 §4.3.5.1 |

**Decision: take what the muxer produces and write no playlist of our own** — unchanged from
edition 3, and the reason is stronger now, because a writer would have to synthesise two
attributes rather than one.

**Three recorded Apple non-conformances, accepted together:** `FRAME-RATE` absent (9.15),
`AVERAGE-BANDWIDTH` absent (9.14), and `BANDWIDTH` nominal rather than a measured peak
(9.13, and 1.26/1.27's "declared within 10 %" cannot hold for a number that is the same for
every source at a rung). All three close by the same trigger and the same remedy: when Phase 6
needs them, either the image moves to trixie — which supplies `AVERAGE-BANDWIDTH` and
`CHANNELS` for free — or a *line-append* over ffmpeg's finished master inserts the values the
domain already holds, never a rewrite.

**B26 asserts exactly this, absences included**, so a base-image bump or an ffmpeg upgrade that
starts emitting either attribute turns the row red and fires the trigger by itself. **B33 makes
the binary itself an asserted fact** in `make smoke`, so the evidence base can never again drift
from the deployment without a gate going red.

**Deliberately not done: Apple 9.10** ("for HD content SHOULD provide at least two variants at
the highest resolution"). Three rungs at one variant each is the brief's §11 ladder, and a
second 1080p variant doubles the most expensive rendition for a self-hosted volume with no CDN
bill. Trigger to add it: measured sessions that oscillate between 720p and 1080p once Phase 6
has player telemetry.

**`SelectLadder` — the rule, including the two cases edition 1 got wrong.**

- Rungs whose height exceeds the source's height are omitted, so nothing is upscaled.
- **A source below the lowest rung still produces exactly one rendition, and the filter must
  be written to make that true.** This is the open defense row 4, and edition 2 closed it on a
  claim GATE 2 round 2 **executed and disproved**: with
  `scale=w=640:h=360:force_original_aspect_ratio=decrease` a 320×240 source encodes at
  **480×360** — the option fits the target *box* to the source's aspect ratio, it never caps at
  the source's own size. Edition 2 asserted the opposite in a "verified" voice, and R21's
  exact-argv test would have pinned the wrong expression into the codebase.

  **The corrected expression, verified on the host's 7.1.1 at GATE 2 round 2 and re-verified
  inside the worker container on ffmpeg 5.1.9 at round 5, is**
  `scale=w='min(640,iw)':h='min(360,ih)':force_original_aspect_ratio=decrease:force_divisible_by=2`.
  With it, 320×240 stays `320x240`, while 1920×1040 (`640x346`) and the 1080×1920 portrait
  case (`202x360`, `406x720`, `608x1080`) are unchanged from the runs that already passed.

  With the filter corrected, the decision itself stands, and it was taken over two
  alternatives: refuse a sub-360p source as `unsupported_source`, or relax the
  `state='ready' ⇒ cardinality(renditions) >= 1` CHECK to permit an empty publish. Instead
  **`360p` is always selected** and the `min()` guard keeps the encode at the source's own
  size, so "never upscale" and "always at least one rendition" are both true and the CHECK
  stays load-bearing rather than being weakened to fit a case. The rung name is a ladder label,
  not a claim about pixel height: Phase 6 reads the master playlist's `RESOLUTION`, which
  ffmpeg writes from the real dimensions. Refusing a 320×240 screen capture — a legitimate
  lesson recording — would be hostile, and relaxing the CHECK would let a future empty publish
  advertise `ready` with nothing to play. **B17** covers the selection, **B21** the argv,
  and **B30** executes a 320×240 clip end to end so the claim can never again be asserted
  without evidence.
- **Selection is by height, and portrait sources are accepted as they are**: a 1080×1920 clip
  selects all three rungs and is scaled to 202×360, 406×720 and 608×1080 by the `decrease`
  rule. Legal, and covered by an acceptance row rather than left to discovery.

**Every scale expression carries `force_divisible_by=2`.** Verified: a 1920×1040 source
scaled with `w=640:h=360:force_original_aspect_ratio=decrease` yields `640x347` and libx264
aborts with *"height not divisible by 2"* — a wrong result for a common class of legitimate
inputs (letterboxed exports, 4:3 screen recordings, phone video). The corrected expression is
`scale=w='min(640,iw)':h='min(360,ih)':force_original_aspect_ratio=decrease:force_divisible_by=2`
— see the `min()` correction above — and R21's exact-argv assertion pins it.

Probe: `ffprobe -v error -show_streams -show_format -print_format json -- <source>`, parsed
into a strict struct; unknown fields ignored, missing ones a refusal rather than a zero.

Render, one invocation producing every rendition and the master playlist:

```
ffmpeg -nostdin -hide_banner -loglevel error -y -i <source>
  -filter_complex "[0:v]split=N[a][b]…;[a]scale=w='min(640,iw)':h='min(360,ih)':force_original_aspect_ratio=decrease:force_divisible_by=2[v0];…"
  -map "[v0]" -c:v:0 libx264 -profile:v:0 high -level:v:0 4.1 -pix_fmt:v:0 yuv420p
              -crf:v:0 21 -maxrate:v:0 365k -bufsize:v:0 730k -preset veryfast
              -force_key_frames:v:0 "expr:gte(t,n_forced*2)"
  …one -map/-c:v group per selected rendition, with that rung's ceiling…
  -map a:0 -c:a:0 aac -profile:a:0 aac_low -b:a:0 128k -ac 2
  -f hls -hls_time 6 -hls_playlist_type vod -hls_flags independent_segments
  -hls_segment_type fmp4 -hls_fmp4_init_filename init.mp4
  -hls_segment_filename <work>/%v/segment_%05d.m4s
  -master_pl_name master.m3u8
  -var_stream_map "a:0,agroup:aac,name:audio,default:yes v:0,agroup:aac,name:360p v:1,agroup:aac,name:720p …"
  <work>/%v/index.m3u8
```

One audio output, N video outputs, one shared `agroup` — Apple 9.5. `-hls_fmp4_init_filename
init.mp4` is the *template*; ffmpeg writes `init_N.mp4` per variant in `var_stream_map` order
(D-8). `-hls_segment_type fmp4` is Q3/§24.12, which is why ADR 009 is mandatory. `libx264`
and the native `aac` encoder are built into the Debian ffmpeg already in the worker image, and
every option this argv uses — `min()` in `scale`, `force_divisible_by`, `hls_segment_type
fmp4`, `var_stream_map` with `agroup` — exists in 5.1.9 and was executed there, so **no image
change is required for the bytes**. What an image change would buy is playlist metadata only,
and that is deferred with its trigger above.

Thumbnail, a second short invocation: `-ss <1 s, or 0 if the source is shorter> -i <source>
-frames:v 1 -vf scale='min(640,iw)':-2 -f image2 <work>/thumbnail.jpg` — the same `min()`
guard, so a 320-wide source yields a 320-wide thumbnail rather than an upscaled one.

The argv builder is a pure function returning `[]string`, and its exact output is asserted —
the only honest way to prove "no user argument reaches ffmpeg" without running ffmpeg, and
the instrument that would have caught the odd-dimension defect.

### D-12 Status delivery: a bounded poll on a separate read

Q5 is polling, no websockets and no SSE. The read is **not** folded into the 4a course tree:
that response mirrors the `PUT` body, whose schema is `additionalProperties: false` and
deliberately excludes lifecycle fields (4a D-3a), and folding media in would couple the media
module to the course aggregate and invalidate the editor's tree on every tick.

`GET /api/v1/admin/courses/{courseId}/videos` is its own query key with
`videoPollMs = 3_000`. **The poll predicate is bounded by state, not by "non-terminal"**:

```
refetchInterval = (rows) =>
  rows.some(r => r.state === 'queued' || r.state === 'processing') || transferInFlightInThisTab
    ? videoPollMs
    : false
```

A `created` or `uploading` row does **not** keep the poll alive: without a transfer running in
this tab nothing will change it, and edition 1's "any non-terminal row" predicate would have
polled every open editor of that course every three seconds for 24 hours after one abandoned
upload. Query's default `refetchIntervalInBackground: false` stops it in a hidden tab.

**The two stale thresholds, named once so the client and the server agree.** A retry control
must not be offered where the server would refuse it. `queued` becomes retryable in the UI
after **10 minutes** — C6's constant, exported to the client as a documented value — because a
single worker legitimately busy with a previous 30-minute render should not invite a retry that
does nothing; `processing` becomes retryable after **40 minutes**, T9's claim age, below which
the server answers `409 video_not_retryable`. Edition 3.1 said "older than N minutes" in three
places and never gave N, which is how the two sides come to disagree on the day.

**The per-lesson badge, when several rows exist.** A lesson can hold two `ready`, one
in-flight and any number of `failed` rows. The field shows, in order: the non-terminal row if
there is one; else the highest-generation `ready` row; else the highest-generation `failed`
row. So a failure stays visible until it is retried or superseded, and a successful replace
immediately supersedes an older failure.

### D-13 The upload control: unsaved lessons, and interrupted uploads

**Unsaved lessons.** 4a mints lesson ids on the client (`CourseEditorView.tsx:288`) and they
exist in the database only after the tree `PUT` lands, so a mint against one would answer
`404 lesson_not_found` — a refusal caused by the UI. `LessonVideoField` therefore renders the
upload control only for a lesson id present in the last **loaded** tree; a newly added lesson
gets a short hint instead. Trigger to revisit: if the two-step flow annoys users, the mint
could accept a lesson id the same request creates — a course-module change, not a media one.

**Interrupted uploads (BLK-5).** A browser cannot resume a file the user has not re-selected,
and the tus client's state is in memory, so a reload or a closed tab leaves a `created` or
`uploading` row with no transfer behind it. Edition 1 left that lesson locked until
`expires_at`, 24 hours later, with no cancel anywhere in the port or the UI. Instead:

- the frontend port gains **`cancelUpload(videoId)`** = the tus `DELETE` already in the server
  surface (D-2d);
- for a `created`/`uploading` row with no transfer in this tab, the field renders
  *"upload interrupted — choose the file again to resume, or cancel"*, where **choose again**
  resumes through `HEAD` → `PATCH` from the reported offset (the point of tus) and **cancel**
  evicts the row so a fresh upload can start immediately;
- the poll does not run for such a row (D-12).

### D-14 Observability (§17)

- `courses_transcoding_job_duration_seconds{outcome}` — histogram, outcomes
  `ready|failed|cancelled`, exponential buckets from 1 s to ~34 min. `prometheus.DefBuckets`
  tops out at 10 s and would put every real transcode in `+Inf`.
- `courses_transcoding_failures_total{reason}` — counter, reasons `unsupported_source`,
  `probe_failed`, `render_failed`, `publish_failed`, `storage_unavailable`. Bounded
  cardinality: a closed set of constants, never an error string.
- **Every label value is pre-initialised with `WithLabelValues` at registration** (BLK-6), so
  both metric families and all their series exist at zero before any job runs. Without that,
  a labelled metric has no series until first observation, and `make smoke` — which runs
  **before** `make e2e` in the gate order — could not assert anything about them.
- `courses_queue_depth{queue,state}` already exists and already watches `media`
  (`platform/redis/queue.go:31-56`).
- HTTP duration and status for the upload service come free from `httpserver.NewHandler`.
- **Logging.** No filesystem path, no ticket id, no original filename, no query string. A
  media log line carries the asset id, the state and a reason constant. `httpserver` logs the
  mux `route`, not the request path (`server.go:124`), so a ticket id in a URL never reaches
  the access log; `QueueLogger` already suppresses Asynq's free-form arguments.

### D-15 Dependency failure: Redis, PostgreSQL, and a volume that is full or read-only

The open defense row 2, answered per dependency rather than in general.

**Redis unreachable at `Enqueue`.** The order is fixed as **commit, then enqueue**. So a
failed enqueue leaves a committed `queued` asset with no task: the use case returns
`503 dependency_unavailable` to the caller, the row stays `queued`, and there are two exits —
the operator's retry (which accepts `queued`, D-10) and the worker-start recovery class C6
(D-7). **C6 is an automatic exit only for this window**, and only because it now bumps
`attempt` before enqueueing: in this path no task of that id was ever created, so even the
edition-3.1 form would have worked here, while on the archived-task paths it silently would
not. The reverse order would be worse: a task pointing at an `uploading` row the handler
skips, with the same stranding and no state to retry from. `ErrTaskIDConflict` is a *success*
and is not this case.

**PostgreSQL down when the terminal state must be written.** "The handler writes `failed`
before returning the error" is circular if the database is the thing that is down, so:
nothing is written, the handler returns the error, and Asynq retries (it needs only Redis).
If the database is still down on the final attempt, the row keeps whatever it last had —
`processing` — and Asynq archives the task. **The asset is not lost and no state is guessed:**
when PostgreSQL returns, the next worker start runs recovery class C5, finds the stale
`processing` row and sets it `queued`, and it is transcoded again. The terminal state is
*deferred*, never invented, which is what spec criterion 2 actually asks for. If PostgreSQL is
down at `Claim`, nothing is claimed and the retry is a no-op; if it is down at worker start,
`runtime.Open` fails and the container exits, and Compose's `restart: unless-stopped` retries.

**The media volume full or read-only.** Three sites assume a write succeeds — the tus append,
the ffmpeg render, and the rename publish — and edition 1 mentioned none of them.

- **`MediaStorage.Free(ctx) (int64, error)`** (one `statfs`) gates two places, **both inside
  processes that actually have the volume mounted**: the **first `PATCH`** in the upload
  service refuses with `507 Insufficient Storage` when free space is below
  `3 × declared_bytes` (source plus renditions plus headroom), before a single byte is
  written — the row then rests in `created`, unchanged and cancellable, not `failed`, because
  a full disk is an operator condition rather than a bad upload; and the **claim** in the
  worker refuses to start a render when free space is below `3 × sourceBytes`, requeueing (T6)
  and returning a retryable error with reason `storage_unavailable`.
  The gate cannot live at the mint: that runs in `api`, which has no `media-data` mount
  (`docker-compose.yml:72-73`), and mounting the volume there to answer a disk question would
  widen the blast radius of the process that handles sessions (§13, N-2). The upload service
  already holds the volume *and* `declared_bytes`, so the check costs nothing there.
  **The claim-time gate is kept deliberately**, though the render already maps `ENOSPC` to a
  retryable `storage_unavailable`: one `statfs` is cheaper than starting a render that will
  occupy the single worker for up to thirty minutes before failing on a condition that was
  knowable at the start. That is its whole justification, and it is the reason it is not cut.
  ADR 004 asks for exactly this ("disk quotas … gate uploads"); D-5's two-generation cap
  remains the growth bound, so no quota subsystem is introduced.
- **A write that fails anyway** (a racing writer, or `EROFS`): the tus `PATCH` truncates the
  scratch file back to `received_bytes` and answers `507 Insufficient Storage` for `ENOSPC` or
  `503` otherwise — the offset does not move, so the client resumes later and **the source is
  not lost**, which is spec criterion 1. The render maps the same errors to a **retryable**
  `storage_unavailable`, not to a permanent failure, and removes its working directory so the
  failed attempt does not itself consume space. A rename onto the same filesystem copies no
  data, so it fails only on `EROFS`, treated identically.
- **Readiness.** Both the worker's and the upload service's `Ready` gain a media-root write
  probe (create and unlink a file under `temporary/`), so a full or read-only volume shows as
  unhealthy in `/readyz`, the Compose healthcheck and `make smoke`. **It does not stop
  traffic**: nginx does not health-check its upstreams, so requests keep arriving and the
  per-request answer is the `507`/`503` above. Stated rather than overclaimed.

**`expires_at` does not apply to a `queued` or `processing` asset.** The ticket governs the
*upload window* only. A transcode that runs past the ticket's expiry is not evicted and is not
taken over; only `created`/`uploading` rows are (D-5, D-7 class C3). Edition 1's "non-terminal
asset whose ticket expired" would have deleted a working transcode's source out from under it.

### D-16 One codec today, AV1 deferred on evidence rather than by default

Recorded so a later reader sees the ladder was chosen, not inherited. `work/PRODUCT-BRIEF.md`
§24.14 (2026-09-06, from `docs/media/research-codecs.md`, 37 sources) settles three things
this plan implements rather than re-decides:

- **H.264/AAC is the Phase 5 ladder**, and every large service keeps H.264 as its base:
  Netflix "primarily relied on H.264/AVC" with AVC still its most-watched codec, YouTube takes
  H.264 as the upload format, Bitmovin's developer report puts H.264 in production for 84–86 %
  of respondents, and Apple 1.12 asks for H.264 "for backward compatibility"
  (`research-codecs.md` §1, §2.1–2.3).
- **HEVC is deliberately excluded** — its advantage is Safari-only, and Disney+ lost 4K/HDR in
  eleven EU countries in August 2026 to a HEVC patent injunction (`research-codecs.md` §1).
  Nothing in this plan encodes, muxes or advertises HEVC.
- **AV1 is added only as an extra rung beside H.264, never instead of it, and only when all
  three §24.14 conditions hold at once**: (1) egress actually costs money — while nginx serves
  from a local disk with no CDN, a ~30 % bitrate saving saves nothing; (2) Phase 6 player
  telemetry shows a stable ≥ 50 % of sessions with an AV1 decoder; (3) the worker can afford
  a second SVT-AV1 pass without breaking publish latency.

**The fMP4/CMAF choice of Q3/§24.12 is what keeps that addition remux-free**: Apple requires
fMP4 for both HEVC and AV1 (1.5, 1.39), so a future AV1 rung joins the same master playlist as
another `CODECS="av01.0…"` variant set over the same container, and none of the bytes this
phase writes need re-packaging. That is the concrete reason Phase 5 pays the fMP4 cost now
instead of shipping MPEG-TS, and it is why ADR 009 supersedes ADR 005's `.ts` clause rather
than merely differing from it.

---

## 5. Side effects (Сайд-ефекти)

### a. A third executable changes the build, the image and the topology

`backend/Dockerfile` builds `/out/upload` in the existing `build` stage and adds
`FROM runtime AS upload` — **without ffmpeg**, so the byte-receiving service cannot execute a
codec. `docker-compose.yml` gains an `upload` service: `media-data:/media` writable,
`HTTP_ADDR=:8082`, `read_only: true` with `tmpfs: [/tmp]`, `mem_limit: 512m`,
`pids_limit: 64`, a `/readyz` healthcheck, `depends_on` on `migrate` completed and `redis`
healthy — the `media-worker` block copied and trimmed. `nginx.depends_on` gains
`upload: {condition: service_healthy}`.

**The volume-initialisation trap.** A named Docker volume is populated from the image of
whichever container mounts it first, and only if that image has content at the mount point.
Today only the worker image pre-creates `/media/{source,hls,thumbnails,temporary}`
(`Dockerfile:32`); nginx mounts `:ro` and has nothing there. With three mounters the order is
not ours, so **both binaries `MkdirAll` their roots at startup** and the Dockerfile's `mkdir`
stays only as belt. The containers are `read_only: true`, but `/media` is a volume mount and
remains writable.

**The `api` service gets no media mount, and that is load-bearing rather than incidental.**
`docker-compose.yml:72-73` leaves `api` `read_only` with `tmpfs: [/tmp]` and no `media-data`,
and Phase 5 keeps it that way: §1's rule is that the byte path never touches the API, and D-15
moves the free-space gate to the upload service precisely so no disk question forces a mount
into the process that handles sessions. The API composition root wires no storage adapter at
all (§2), so this cannot regress silently — there is nothing for a `statfs` to be called on.

**No new secret and no new environment variable beyond `HTTP_ADDR`** — a consequence of D-2b.
`.env.example` is unchanged.

### b. The migration collides with the course test canary — the third time

The course `repository_integration_test.go` subtest **"failed migration rolls back its
complete transaction"** (`:190`) builds an `fstest.MapFS` (`:195`) of the four real migrations
plus a synthetic **`00005_failure.sql`** (`:200`). The real `00005_media.sql` collides, so M02
adds it to that MapFS and **renumbers the canary to `00006_failure.sql`** — the pattern 4a
used for `00003` (`fd90fc9`) and 4b for `00004`. The identity `schema_integration_test.go`
reads via `os.DirFS` (`:57-58`) and picks up `00005` automatically.

### c. Nginx gains one location and one zone; `/media` is untouched

```nginx
limit_req_zone $binary_remote_addr zone=upload:10m rate=30r/s;   # new, beside api/web/login

location ^~ /uploads/ {
  limit_req zone=upload burst=60 nodelay;
  client_max_body_size 16m;          # one tus chunk (8 MiB) with headroom
  proxy_request_buffering off;       # stream to upstream; see below
  proxy_read_timeout 120s;
  proxy_send_timeout 120s;
  set $upload http://upload:8082;
  proxy_pass $upload;
  proxy_set_header Host $host;
  proxy_set_header X-Request-ID $request_id;
  proxy_set_header X-Forwarded-Proto $scheme;
  proxy_set_header X-Forwarded-For $remote_addr;
  proxy_set_header traceparent "";
  proxy_set_header tracestate "";
}
```

- **`proxy_request_buffering off` is mandatory.** The nginx container is `read_only: true`
  with `tmpfs: [/tmp, /var/cache/nginx]` and `client_body_temp_path /tmp/client_temp`
  (`nginx.conf:8`), so buffered request bodies land in **RAM**.
- **The global `client_max_body_size 1m` (`nginx.conf:20`) stays**, overridden only inside
  this location. Nginx enforces it from `Content-Length` before proxying, even with buffering
  off — **but a `Transfer-Encoding: chunked` request has no `Content-Length` to check and
  streams past it**, which is why the tus server carries its own `MaxBytesReader` cap
  (D-2d step 4) rather than trusting the edge. Q2's 2 GiB is met by **chunking**: 8 MiB × 256 requests, so no single request
  is large. This is why the file limit and the body limit are different numbers, and why 4a's
  and 4b's "`client_max_body_size` untouched" claims stay true everywhere else.
- **A dedicated zone, not `api`.** 256 chunk `PATCH`es through `api`'s `10r/s burst=20` would
  delay a legitimate upload behind its own rate limit and interleave badly with the editor's
  other calls.
- **`^~ /uploads/` cannot shadow `^~ /media/`**; both `/media` lines keep returning 403 and a
  smoke assertion continues to prove it.

### d. `make smoke` breaks the moment the first handler is registered

`smoke.sh` asserts the worker exposes **zero** handlers — written to prove the foundation was
not pretending to transcode. Phase 5 inverts it: the worker must report exactly
`["media:transcode"]` and `stage: "media"`. Four assertions are added in the same edit:
`HEAD /uploads/{unknown-uuid}` answers `404` (not a proxy error), a 32 MiB `PATCH` body is
refused by nginx with `413`, `/media/example/master.m3u8` still answers `403`, and — the one
that exists because of N-17 — **the worker's `ffmpeg -version` is printed and its major line
asserted** (B33), so every gate run records which binary produced the media and a base-image
bump cannot silently invalidate D-11's table and B26.

### e. `make check` gains no step, but `make integration` gains no ffmpeg

The gate order is unchanged, so `.claude/standards.md` is untouched. But the integration
container is `golang:1.27.1-bookworm` mounting only `backend/`
(`infrastructure/tests/integration.sh`) and **has no ffmpeg**:

- **No Go test may depend on ffmpeg being present.** A conditional skip is the "skipped
  reported as passed" trap AGENTS.md forbids, so there is none. What is *pure* (argv
  construction, ffprobe JSON parsing, ladder selection) is unit-tested exactly; what is
  *processual* (subprocess lifetime and cancellation) is proven against `/bin/sleep`.
- **The real ffmpeg run is proven in Compose**, where the worker image has ffmpeg. This is a
  genuine evidence gap between `make integration` and `make e2e`, named rather than papered
  over. Trigger to close it: a transcode regression e2e can detect but not localise.
- **No Redis Testcontainer is introduced.** The queue adapter's one real behaviour
  (`ErrTaskIDConflict`) is proven at the application layer with a fake port returning that
  sentinel; the real Asynq round trip is proven in Compose.

### f. e2e needs a video, and video bytes may not be committed

AGENTS.md forbids committing media bytes, so the fixture is generated by the container that
already has ffmpeg and copied out — **into `/tmp`, not into `/media/temporary`**, so it never
becomes a row-less file on the media volume (D-7):

**Four clips are generated, one per shape the acceptance rows need** — three short ones from
one command, and the long R35 fixture from the same command plus a rate-control flag:

```
for size in 640x360 1280x720 320x240; do
  docker compose exec -T media-worker ffmpeg -nostdin -loglevel error -f lavfi \
    -i testsrc2=size=$size:rate=15 -f lavfi -i sine=frequency=440 -t 2 \
    -c:v libx264 -preset veryfast -c:a aac -y /tmp/e2e-$size.mp4
  docker compose exec -T media-worker cat /tmp/e2e-$size.mp4 > playwright/.fixtures/e2e-$size.mp4
done

# R35's fixture: five minutes, and -crf 32 so it is ~8 MB rather than ~77 MB.
docker compose exec -T media-worker ffmpeg -nostdin -loglevel error -f lavfi \
  -i testsrc2=size=1280x720:rate=15 -f lavfi -i sine=frequency=440 -t 300 \
  -c:v libx264 -preset veryfast -crf 32 -c:a aac -y /tmp/e2e-1280x720-5min.mp4
docker compose exec -T media-worker cat /tmp/e2e-1280x720-5min.mp4 > playwright/.fixtures/e2e-1280x720-5min.mp4
```

**`-crf 32` on the long clip only.** Measured with the bare command: `testsrc2` is high-detail
synthetic content and libx264's default CRF 23 spends about 2 Mbit/s on it, so a 5-minute
source is **77 MB** — which the browser then uploads in ten 8 MiB chunks on every run. The
render time is set by frame count, not by source bitrate, so **R35's window is unchanged** by
the flag; only the transfer and the copy-out shrink.

`playwright/.fixtures/` joins `playwright/.auth/` in `.gitignore`.

| Clip | Length | Exists for |
|---|---|---|
| 640×360 | 2 s | the main journey (P04, P05, P08) and the one-rung ladder (B05) |
| 1280×720 | 2 s | the two-rung case, the only way to observe a **shared** audio group (B27) |
| 320×240 | 2 s | the sub-360p case executed end to end, so D-11's corrected `min()` filter is proven by output rather than asserted (B30) |
| 1280×720 | **5 min** | **R35 alone.** Measured in the container: the 2 s clips render in well under a second, and Playwright polls status every 3 s (D-12), so a restart aimed at "mid-render" would usually land after `ready` and pass without exercising the shutdown path it exists to prove — the only end-to-end evidence for the BLK-1 closure. A 5-minute 1280×720 source renders in **39 s** under `cpus: 2.0` (measured directly at GATE 2 round 6; the estimate from the 60 s-source rate was ~45 s), a comfortable window to observe `processing` and restart into. About **8 MB** with `-crf 32`, and 93 MB of HLS output on the volume |

The three short clips cost the run almost nothing. The long one costs **roughly two to two and
a half minutes** end to end — 22 s to generate, the copy-out, the upload, 39 s of render
interrupted, asynq's 10 s drain, the container restart and healthcheck, then a full 39 s render
— which is the price of the restart row being real rather than a race. Stated as measured
rather than estimated, because edition 3.2 guessed "roughly a minute" and "a few MB" and was
wrong about both.

The spec creates its own course and lesson in a per-run slug namespace (the 4a §5.9 idiom), so
`make e2e` is re-runnable on its own. It uses the shared `storageState` and performs **zero**
logins.

### g. Two kit components get their first consumer

`FileUpload` (`controls.tsx:165-178`) and `ProgressBar` (`data.tsx:7-23`) were built in
foundation from the brief's §6 list and have **no consumer outside `ui.test.tsx`**. Two gaps
are expected:

- `FileUpload` is a `FileTrigger` wrapping a `Button` — no drag-and-drop, no busy state. The
  field passes `acceptedFileTypes` (a `FileTriggerProps` member) and disables the button while
  a transfer is in flight.
- `ProgressBar` renders `percentage ?? 35` as the fill width (`data.tsx:17`), a reasonable
  indeterminate placeholder — so the field must pass a real `value` during transfer and switch
  to an indeterminate presentation during `processing`, where no percentage exists.

If either needs a prop, the change lands in the kit with every existing caller updated in the
same task — the 4a A08 rule. The kit stays i18n-free: labels arrive as props through `t()`.

### h. The editor form grows a non-form child

`CourseEditorView` is one React Hook Form over the whole tree. The video field is **not** an
RHF field: it has no value in the `PUT` body and its state lives in a separate query. It
mounts inside the existing lesson `fieldset` (`:231`) as a sibling of the RHF inputs, with its
own mutations. Two accepted consequences: an upload does not mark the form dirty, and saving
the tree does not cancel a transfer in flight. Removing a lesson and saving *will* orphan its
asset — D-4's `SET NULL` path — and the field warns before the remove button acts on a lesson
that has video.

### i. Deleting a course or a module now costs bytes

`ON DELETE CASCADE` from `courses` → `course_modules` → `course_lessons` already existed; with
D-4 the assets underneath become orphans and the sweeper reclaims them. A 4a draft delete is
therefore no longer purely a database operation — it schedules disk work, bounded by D-5's two
generations per lesson and the sweeper's 50-row cap.

### j. Unchanged on purpose

`/media` and `/media/` at the edge; `client_max_body_size 1m` outside `/uploads/`; the
`login`, `api` and `web` zones; the gate order and `.claude/standards.md`; Prometheus;
`.env.example`; `seed.sql` and `seed-identities.sql`; `.gitignore` (already fixed in
`5ef6c1f`); the course and identity modules' behaviour (media adds one read of their tables
and changes none of their code); SSR (admin routes are `ssr: false` and the SSR container
carries no `Origin`, so it cannot call a media mutation).

---

## 6. Migration plan (План міграції)

`backend/migrations/00005_media.sql`, goose, applied by `make up` / `make migrate` before any
seed.

**Up:**

1. ```sql
   CREATE TABLE video_assets (
     id uuid PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'::uuid),
     lesson_id uuid REFERENCES course_lessons(id) ON DELETE SET NULL,
     generation integer NOT NULL CHECK (generation >= 1),
     state text NOT NULL CHECK (state IN ('created','uploading','queued','processing','ready','failed')),
     claim_token uuid,
     declared_bytes bigint NOT NULL CHECK (declared_bytes BETWEEN 1 AND 2147483648),
     received_bytes bigint NOT NULL DEFAULT 0,
     original_name text NOT NULL CHECK (char_length(original_name) BETWEEN 1 AND 255),
     attempt integer NOT NULL DEFAULT 0 CHECK (attempt >= 0),
     failure_reason text CHECK (failure_reason IS NULL OR char_length(failure_reason) BETWEEN 1 AND 200),
     renditions text[] NOT NULL DEFAULT '{}',
     expires_at timestamptz NOT NULL,
     evicted_at timestamptz,
     created_at timestamptz NOT NULL,
     updated_at timestamptz NOT NULL,
     CONSTRAINT video_assets_received_within_declared CHECK (received_bytes >= 0 AND received_bytes <= declared_bytes),
     CONSTRAINT video_assets_updated_not_before_created CHECK (updated_at >= created_at),
     CONSTRAINT video_assets_evicted_not_before_created CHECK (evicted_at IS NULL OR evicted_at >= created_at),
     CONSTRAINT video_assets_ready_has_renditions CHECK (
       state <> 'ready' OR (
         cardinality(renditions) >= 1
         AND array_position(renditions, NULL) IS NULL
         AND '' <> ALL (renditions)
       )
     ),
     CONSTRAINT video_assets_failed_has_reason CHECK (state <> 'failed' OR failure_reason IS NOT NULL),
     CONSTRAINT video_assets_claim_only_while_processing CHECK (claim_token IS NULL OR state = 'processing')
   );
   ```
   Three CHECKs are load-bearing, in the 4b `deleted_implies_disabled` spirit: **a `ready`
   asset must name at least one rendition** (so "ready" cannot be advertised for an empty
   publish — and D-11's always-one-rung rule is what makes it satisfiable for a sub-360p
   source); **a `failed` asset must carry a reason** (so spec S3's "shows why in operator
   terms" cannot be silently unmet); and **a claim token exists only while `processing`**, so
   a stale token cannot survive a terminal transition and be replayed. `2147483648` is exactly
   Q2's 2 GiB.

   **`cardinality`, not `array_length` — and the element checks beside it.** Editions 2–3 wrote
   `array_length(renditions, 1) >= 1`, which enforced nothing: `array_length('{}', 1)` is
   `NULL`, `NULL >= 1` is `NULL`, and **a CHECK evaluating to `NULL` is satisfied, not
   violated**. `cardinality('{}')` is `0`, so the row is refused; the column is `NOT NULL`, so
   `cardinality` never receives `NULL` from the other side. Round 5 then executed the fixed form
   and found it counts *elements, not renditions*: `'{NULL}'` and `'{""}'` both satisfy it, and
   an array holding one NULL is as much "nothing to play" as an empty one. Hence
   `array_position(renditions, NULL) IS NULL` and `'' <> ALL (renditions)` beside the count.
   Nothing in the application can produce either — rung labels come from a constant table — but
   this constraint's whole job is to be the thing that still holds when the application is
   wrong. **B32 inserts all four shapes** rather than trusting the reading. (§13 records how a
   constraint asserted in prose survived two editions and a gate round.)

   **Every table-level CHECK is named.** PostgreSQL generates `video_assets_check`,
   `_check1` … `_check5` — ordinals that shift when a CHECK is added above them, so a test
   keying on `video_assets_check3` breaks silently and a test that does not key on the name
   proves only "some CHECK fired". All five get explicit handles, including
   `video_assets_received_within_declared`: written inline on the column it *looks* like a
   column CHECK, but referencing `declared_bytes` promotes it to table level, where PostgreSQL
   would have called it plain `video_assets_check` — the one auto-name round 6 found. Column
   CHECKs keep their generated names (`video_assets_generation_check` and so on), which are
   stable by column rather than by ordinal and need no handle. B32 asserts
   `video_assets_ready_has_renditions` by name, the same discipline D-1 already applies to
   `23505` by `ConstraintName`.

   **`created_at` and `updated_at` carry no `DEFAULT now()`.** Executed: an INSERT that writes
   `updated_at` from an application clock read *before* `pool.Begin` — the natural shape when a
   clock is injected into a use case — while leaving `created_at` to `DEFAULT now()` (the
   transaction start) fails `updated_at >= created_at` with a `23514`. 4a avoids this by writing
   both columns from one argument and never relying on the default (`courses.sql:112-130`,
   `authoring.go:16,23`), and the media mint does the same. Dropping the defaults turns a
   forgotten column into a deterministic `23502` not-null violation on the first test that runs
   the query, instead of a `23514` that appears only when the clock and the transaction race —
   an error that always fires beats one that usually does not. It is **not** a compile-time
   error, as edition 3.2 claimed: sqlc does not verify `NOT NULL` coverage of an `INSERT`.

   **`content_type` is gone.** It was written at mint and read by nothing: the first-chunk sniff
   is the authoritative container check (D-8), no Phase 5 response returns it, and the plan
   already removed `duration_seconds` under the rule "a column written for a future phase is a
   speculative column". `original_name` stays and is now actually display metadata — the
   course-videos read returns it (§1), so an editor with two generations can tell which file is
   which.

   **Audit of every other CHECK in this migration for the same class of defect.** Two forms of
   silent pass exist: a CHECK on a *nullable column* with no `IS NULL OR` guard, and a
   *function that returns `NULL` for a legal non-`NULL` input*, which is what bit here. Result:
   `array_length` was the only instance, and no other constraint is exposed.

   | CHECK | Why it cannot evaluate to `NULL` |
   |---|---|
   | `id <> '000…'::uuid` | `id` is the primary key, therefore `NOT NULL` |
   | `generation >= 1`, `attempt >= 0`, `declared_bytes BETWEEN …` | `NOT NULL` columns, plain comparisons |
   | `received_bytes >= 0 AND received_bytes <= declared_bytes` (`video_assets_received_within_declared`) | both operands `NOT NULL` |
   | `state IN (…)` | `state` is `NOT NULL`; `IN` over a literal list of non-`NULL` values |
   | `char_length(original_name) BETWEEN 1 AND 255` | `NOT NULL` column, and `char_length('')` is `0`, not `NULL`, so the empty string is refused rather than admitted |
   | `failure_reason IS NULL OR char_length(failure_reason) BETWEEN 1 AND 200` | nullable, but explicitly guarded — the guard is true when the column is `NULL`, so the function is never reached |
   | `evicted_at IS NULL OR evicted_at >= created_at` | same guard shape |
   | `updated_at >= created_at` | both `NOT NULL` |
   | `state <> 'failed' OR failure_reason IS NOT NULL` | `IS NOT NULL` yields `true`/`false`, never `NULL` |
   | `claim_token IS NULL OR state = 'processing'` | `IS NULL` yields `true`/`false`; `state` is `NOT NULL` |

   The same audit was run over the migrations already in the repository:
   `grep -rn 'array_length\|cardinality' backend/migrations backend/sql` returns nothing, so
   00001–00004 carry no instance of this defect and nothing outside Phase 5 needs correcting.
   `duration_seconds` is **not** in this table: nothing in Phase 5 reads it, and a column
   written for a future phase is a speculative column. Phase 6 adds it with its consumer.
2. `CREATE UNIQUE INDEX video_assets_pending_per_lesson ON video_assets (lesson_id)
   WHERE lesson_id IS NOT NULL AND evicted_at IS NULL
     AND state IN ('created','uploading','queued','processing');` — D-5's one-in-flight rule.
3. `CREATE UNIQUE INDEX video_assets_generation ON video_assets (lesson_id, generation);` —
   D-5's monotonic numbering; NULL `lesson_id` rows do not collide.
4. `CREATE INDEX video_assets_by_lesson ON video_assets (lesson_id, generation DESC)
   WHERE evicted_at IS NULL;` — the course-videos read and the eviction lock.
**No index serves the sweeper, deliberately.** Edition 2 added three more — a partial
reclaimable index, an `expires_at` index and `(state, updated_at)` — to make the six sweep and
recovery predicates cheap. They are dropped. The table holds two live generations per lesson
plus a bounded tail of `failed` and evicted rows: hundreds of rows for a real catalogue, not
millions, and the sweeper takes 50 of them in no particular order once per job. Three indexes
to avoid a scan of a few hundred rows is cost without a benefit, and each one is another object
the down migration and the canary test carry. **The sweep and recovery predicates scan;
trigger to index them: `video_assets` passing roughly 10⁵ rows.** The two unique indexes stay
because they are correctness, and `video_assets_by_lesson` stays because it serves the hot
read.

**No backfill and no data risk.** The table is new and empty on every database, so there is no
analogue of 00003's ownerless-courses hazard. `course_lessons` gains no column, so 4a's
queries and generated models are unchanged and an API rollback needs no schema change (§7).

**Down:** `DROP TABLE video_assets;` — the indexes go with it. Destructive of media links by
construction, so it is exercised only in ephemeral Testcontainers databases, never
operationally. It **cannot fail** on reused data the way 00004's down can, because it re-adds
no constraint.

**Bytes are not migrated and not deleted by the migration.** A `goose down 00005` leaves
`/media` exactly as it was; the volume is emptied only by `docker compose down -v`.

**Clean bring-up (unchanged):** `make up` → `make seed-identities` → `make seed` →
`make smoke` → `make e2e`.

---

## 7. Rollback

- **Application rollback is the supported path.** Redeploy the previous API, frontend and
  worker images, stop the `upload` service, and **leave migration 00005 in place**.
  `video_assets` is a new table nothing before Phase 5 touches; `course_lessons` is unchanged,
  so every 4a query and generated model still match.
- **In-flight jobs.** Tasks already in Redis outlive the worker image. A rolled-back worker has
  an **empty mux again**, so a `media:transcode` task is an unknown type: Asynq fails it
  visibly and archives it — the behaviour foundation deliberately chose
  (`media-worker/main.go:50-51`). Nothing pretends to transcode and nothing is marked `ready`.
  Assets rest in `queued` or `processing`; rolling forward re-queues them through recovery
  classes C5/C6 at worker start (D-7) — C6 bumping `attempt` so its enqueue is not refused by
  the retained task id — which is the same mechanism that handles a crash. To
  drain instead, delete the `media` queue's tasks with the Asynq inspector before rolling back.
- **Already-written media.** Nothing reads `/media` in Phase 4 or Phase 5 — the edge refuses it
  and no binary serves it — so published bytes are inert. They are also **unreachable by the
  sweeper**, which lives in the rolled-back worker, so a long rollback accumulates orphaned
  directories. A disk cost, not a correctness or disclosure one, reclaimed the moment the
  worker rolls forward. Recorded rather than mitigated: a cleanup path outside the worker would
  be a fourth place that may delete media.
- **In-flight uploads.** Stopping the `upload` service abandons every transfer at its current
  offset. The rows stay `created`/`uploading` and the sources stay in `temporary/`; on roll
  forward a client can resume from `HEAD`, and anything past `expires_at` is reclaimed by
  sweeper class C3.
- **Do not `goose down 00005`** to recover from an application defect: it drops every
  lesson↔video link while leaving the bytes on disk, converting a bad state into unattributable
  garbage. A forward migration fixes a schema mistake; the down migration exists for tests.
- **Frontend rollback alone is inert as a security measure** (the 4a stance): removing the
  upload control grants nothing; the mint endpoint, the ticket state machine and the edge's
  `/media` refusal stand on their own.
- **Nginx rollback must accompany the upload service.** If the container is stopped while the
  `/uploads/` location remains, nginx answers `502` for a path the UI no longer offers. The
  location and the service revert together, as 4a's four gate-order files do.

---

## 8. Business error vocabulary

Server-enforced, mapped once in the media handler, contract-proven. New values in the OpenAPI
`error.code` enum:

| Code | HTTP | Meaning | Where decided |
|---|---|---|---|
| `lesson_not_found` | 404 | the lesson id is unknown past the role gate | media repository, inside the mint transaction |
| `upload_in_progress` | 409 | the lesson already has a non-terminal, unexpired asset (D-5) | the partial unique index, mapped from `23505` |
| `unsupported_media` | 400 | the first chunk's container is outside Q2's set | the upload service's sniff (D-8) |
| `upload_too_large` | 400 | the declared length at mint exceeds 2 GiB (a `PATCH` past the declared length is the tus surface's `413`, not this code) | domain bound |
| `video_not_found` | 404 | the video id is unknown or evicted, for status and retry | media repository |
| `video_not_retryable` | 409 | retry was asked for an asset outside `failed` / `queued` / `processing` older than the claim age, **or** for one whose source never landed (`received_bytes < declared_bytes`) | domain transition matrix (D-10) |

Reused unchanged: `authentication_required` (401), `forbidden` (403, wrong role),
`course_forbidden` (403, non-owning instructor — the same meaning the course module gives it),
`origin_forbidden` (403), `invalid_request` (400), `dependency_unavailable` (503),
`internal_error` (500). **The disk gate is not in this vocabulary**: it lives in the upload
service and answers tus's `507` (D-15), so no JSON operation refuses on free space.

A business refusal is never a 500 (spec criterion 3). **No error message contains a filesystem
path, a storage key, a ticket, an ffmpeg command line or an ffmpeg stderr dump**;
`failure_reason` comes from D-14's closed set and raw stderr is neither stored nor logged.

**The tus surface answers tus's own vocabulary**, not this one: `204` for an accepted `PATCH`
or `DELETE`; `404` for an unknown, evicted, expired or completed ticket; `409` for a wrong
`Upload-Offset` or a lost guarded update; `413` for a chunk past the declared length; `412` for
a missing or wrong `Tus-Resumable`; `507` for `ENOSPC` and `503` for any other storage fault
(D-15). The tus endpoints are **not** same-origin gated: the ticket is the capability, and a
cross-origin `PATCH` bearing a valid ticket is its holder exercising it. Stated so the absence
is a decision, not an omission.

---

## 9. Acceptance scenarios

Per `acceptance-spec`: three groups (all present), one observable result per row, one evidence
level per row, exactly one owning task, and every owner able to produce its evidence **at the
time that task lands**. Mutation rows assume a valid `Origin` unless the row is about
`Origin`. **Manual rows: none.** M01 owns no row — its evidence is the two ADR documents
themselves (the 4b T02 precedent); every other task owns at least one.

### Happy path

| # | Scenario | Precondition | Action | Observable result | Evidence | Owner |
|---|---|---|---|---|---|---|
| P01 | An admin mints an upload ticket | admin session; a saved lesson exists | `POST /api/v1/admin/lessons/{id}/video` with a 5 MiB mp4's metadata | `201` carrying `videoId`, `uploadUrl` under `/uploads/`, `chunkBytes`, `maxBytes` and `expiresAt`; a following status read shows `created` | contract `handler_test.go` | M09 |
| P02 | An owning instructor mints | instructor session; a lesson in a course they own | the same `POST` | `201` with the same shape | contract | M09 |
| P03 | HEAD → PATCH lands the source and queues the job | Testcontainers PostgreSQL; a minted ticket; a fake queue port | `HEAD /uploads/{videoId}`, then `PATCH` the whole body | `HEAD` reports offset 0; after the `PATCH` the file exists at the `source/{videoId}/original` key with the declared byte count, the asset reads `queued`, and the fake queue recorded exactly one enqueue for `{assetId, attempt: 0}` | integration `tus/server_test.go` + repository | M11 |
| P04 | The worker transcodes a real clip | Compose up; the generated 640×360 fixture uploaded | wait for the terminal state | the asset reads `ready`; `hls/{videoId}/master.m3u8`, `hls/{videoId}/360p/index.m3u8`, exactly one `hls/{videoId}/360p/init_*.mp4`, at least one `.m4s` segment there, the same three under `hls/{videoId}/audio/`, and `thumbnails/{videoId}.jpg` all exist in the worker container | e2e `media.spec.ts` | M16 |
| P05 | The editor advances to ready without a reload | Compose up; an upload just completed | keep the editor open | the lesson's state label moves through `processing` to `ready` with no navigation and no manual refresh | e2e | M16 |
| P06 | Replacing a video keeps exactly two generations | Testcontainers; a lesson with one `ready` asset | publish a second asset for the same lesson | both read `ready`, `evicted_at` is null on both, and the higher `generation` is the current one | integration | M05 |
| P07 | The course videos read answers one row per asset | admin session; a course with two lessons, one carrying two `ready` assets | `GET /api/v1/admin/courses/{id}/videos` | `200` with exactly two items for that lesson, each carrying `lessonId`, `videoId`, `state` and `generation`, and no storage key or path member | contract | M09 |
| P08 | The full browser journey | Compose up; the fixture generated; admin `storageState` | sign in as admin, open a course editor, choose the clip on a saved lesson, wait | the progress bar advances, the lesson shows `processing` and then `ready`, and a reload still shows `ready` | e2e | M16 |
| P09 | The editor renders the control and the progress | a loaded tree with one saved lesson | choose a file in the component test | the `FileUpload` control is present for that lesson and a determinate `ProgressBar` renders while the mocked transfer runs | component `LessonVideoField.test.tsx` | M15 |
| P10 | The worker registers the media handler | Compose up | `GET /status` on the worker | `handlers` is exactly `["media:transcode"]` and `stage` is `"media"` | `make smoke` | M12 |
| P11 | Polling runs only while work is in flight | MSW answering `processing`, then `ready`, and separately a lone `created` row | mount the editor and let two intervals pass in each case | requests continue while a row is `queued`/`processing` and **stop** at `ready`; a lone `created` row with no transfer in this tab produces **no** polling at all | component with MSW | M14 |

### Boundaries

| # | Scenario | Precondition | Action | Observable result | Evidence | Owner |
|---|---|---|---|---|---|---|
| B01 | The size ceiling is exactly 2 GiB | admin session | mint with `2147483648`, then with `2147483649` | `201`, then `400 upload_too_large`; no asset for the second | contract | M09 |
| B02 | A chunk beyond the declared length is refused | a ticket for 1 MiB, 1 MiB already received | `PATCH` one more byte | `413`, and a following `HEAD` reports the unchanged offset | integration | M10 |
| B03 | A crash mid-chunk is repaired, and the upload resumes | a ticket whose scratch file is longer than `received_bytes` (an appended chunk whose commit never landed) | `HEAD`, then `PATCH` from the reported offset | `HEAD` reports the **column's** value; the `PATCH` truncates the file back to it before appending; the completed file equals the source byte-for-byte | integration | M10 |
| B04 | A zero-length upload is refused at mint | admin session | mint with `declaredBytes: 0` | `400 invalid_request`; no asset row exists | contract | M09 |
| B05 | The ladder never upscales, end to end | Compose up; the 640×360 fixture | transcode it | `master.m3u8` lists exactly one video variant, named `360p`, at `RESOLUTION=640x360`; no `720p` or `1080p` directory exists | e2e | M16 |
| B06 | A third generation evicts the oldest | Testcontainers; a lesson with two `ready` assets | publish a third | exactly two assets have `evicted_at IS NULL`, the evicted one is the lowest generation among the `ready` rows, and after the sweeper runs its directories and its row are gone | integration | M06 |
| B07 | Migration over a populated database | 5 demo courses, 3 identity fixtures, modules and lessons present | apply `00005_media.sql` | `video_assets` exists with its indexes, every pre-existing row is intact, and `course_lessons` is structurally unchanged | integration `repository_integration_test.go` | M02 |
| B08 | The failing-migration canary still rolls back | migration 00005 committed | run the course integration suite | the **"failed migration rolls back its complete transaction"** subtest passes with the real `00005_media.sql` in its `fstest.MapFS` and the canary renumbered to `00006_failure.sql` | `make integration` | M02 |
| B09 | Contract and client stay in step | all OpenAPI and query changes committed | `make check-generated` | sqlc and Orval output matches the committed files; exit 0 | `make check-generated` | M14 |
| B10 | The ordered gate passes | clean checkout | `make check` | every step exits 0 with the new surface, including a worker that actually processed a job | `make check` | M17 |
| B11 | The storage adapter cannot be walked out of | a tempdir root containing a symlink to `/etc` | resolve keys with `..`, an absolute path, a non-UUID id, and a leaf through the symlink | each is refused before any file is opened, and no file outside the root is read or written | integration `storage_integration_test.go` | M07 |
| B12 | Publication is all-or-nothing, and purge is idempotent | a tempdir root; a fully built working directory | publish; then publish an interrupted build; then purge the same key twice | after success `hls/{id}/master.m3u8` exists; after the interrupted attempt **no** `hls/{id}` directory exists at all; the second purge succeeds on ENOENT rather than erroring | integration | M07 |
| B13 | A non-video source is refused at probe | recorded ffprobe JSON for a text file, for a video with no audio stream, and for a valid H.264/AAC clip | parse and accept | the first two are refused as `unsupported_source`; the third is accepted | unit `transcoder_test.go` | M08 |
| B14 | A deleted lesson leaves a reapable orphan, not a leak | Testcontainers; a lesson with a `ready` asset | delete the lesson row (what a 4a `PUT` does) | the asset survives with `lesson_id IS NULL`, is absent from the course videos read, and the sweeper then deletes its bytes and its row | integration | M06 |
| B15 | A course with no video reads empty | admin session; a course whose lessons have no assets | `GET …/videos` | `200` with `items: []`, and the editor renders no video state for any lesson | contract | M09 |
| B16 | An expired upload window does not lock a lesson, but a running transcode does | Testcontainers; case (a) a `created` asset past `expires_at`; case (b) a `processing` asset past `expires_at` | mint again in each case | (a) `201`, the stale asset carries `evicted_at`, one non-terminal asset remains; (b) `409 upload_in_progress` and the transcoding asset is untouched | integration | M05 |
| B17 | The ladder is chosen by the source, never by the client, and always yields a rung | none | select renditions for 640×360, 1280×720, 1920×1080, 1080×1920 (portrait) and 320×240 (below the lowest rung) probes | `[360p]`, `[360p,720p]`, `[360p,720p,1080p]`, `[360p,720p,1080p]` and `[360p]` respectively; no case yields an empty set and no case selects a rung taller than the source's height | unit `source_test.go` | M03 |
| B18 | The container is sniffed on the first chunk | a ticket whose first chunk is a renamed text file | `PATCH` the first chunk | `400 unsupported_media`, the asset reads `failed` with a reason, **no** enqueue was recorded, and no file exists under `source/` | integration | M11 |
| B19 | The transcoding metrics and all their series exist before any job | Compose up, immediately after start | scrape the worker's `/metrics` | `courses_transcoding_job_duration_seconds` exists with a series for each of `ready`, `failed`, `cancelled`, and `courses_transcoding_failures_total` with a series for each of the five reason constants, all at zero | `make smoke` | M12 |
| B20 | The upload control is operable without a mouse | the editor rendered with one saved lesson | tab to the control and activate it | the file trigger is focus-visible, at least 44px, reachable by keyboard, the progress bar exposes an accessible value, and axe reports no violations | component with axe | M15 |
| B21 | Odd-dimension, portrait and sub-rung sources produce even, never-enlarged argv | probes for 1920×1040, 1080×1920 and 320×240 | build the render argv | every scale expression is the `w='min(640,iw)':h='min(360,ih)':force_original_aspect_ratio=decrease:force_divisible_by=2` form — carrying both the `min()` caps and `force_divisible_by=2` — for every rung, and the thumbnail filter carries `scale='min(640,iw)':-2` | unit `transcoder_test.go` | M08 |
| B22 | Generation numbering is monotonic across evictions | Testcontainers; a lesson whose only asset has been evicted | mint again | the new asset's `generation` is the evicted one's plus one, not 1; the unique `(lesson_id, generation)` index holds | integration | M05 |
| B23 | A full volume refuses the first byte rather than half-writing | a tempdir-backed storage adapter whose `Free` reports below `3 × declared_bytes`; a minted ticket | `PATCH` the first chunk | `507 Insufficient Storage`; no scratch file exists, `received_bytes` is still 0, and the asset still reads `created` — cancellable, not `failed` | integration | M11 |
| B24 | The recovery pass re-queues a stale claim, and the stale claim cannot publish | Testcontainers; an asset `processing` with `updated_at` older than the threshold and claim token A | run recovery, claim again (token B), then attempt a publish with token A | recovery sets `queued` and clears the token; the second claim succeeds with B; the publish with A affects zero rows and is reported as a lost claim, not as success | integration | M06 |
| B25 | A completed job is observed in the histogram | Compose up; P04 has completed | scrape the worker's `/metrics` | `courses_transcoding_job_duration_seconds_count{outcome="ready"}` is at least 1 | e2e | M16 |
| B26 | The playlists carry what the target binary really produces, and the three gaps stay visible | Compose up; the 2 s 1280×720 and the 320×240 clips transcoded | read `master.m3u8` and each variant playlist for both | every `EXT-X-STREAM-INF` carries `CODECS`, `RESOLUTION` and `BANDWIDTH`; one `EXT-X-MEDIA:TYPE=AUDIO` group is referenced by every variant; `EXT-X-INDEPENDENT-SEGMENTS` is present in every **video variant** playlist and absent from **both** the master **and `audio/index.m3u8`** — 5.1.9 writes the tag only into playlists carrying a video stream, so an implementer who generalises this to "every variant playlist" gets a red row on the audio rendition; **`AVERAGE-BANDWIDTH` and `FRAME-RATE` are absent everywhere**; and the 360p `BANDWIDTH` is byte-identical between the two clips, proving it is nominal rather than measured — the three recorded Apple non-conformances (9.14, 9.15, 9.13/1.26) asserted so an ffmpeg or base-image change that starts emitting them turns the row red and fires D-11's trigger | e2e | M16 |
| B33 | The worker's ffmpeg is the binary the plan targets | Compose up | `docker compose exec media-worker ffmpeg -version` from `smoke.sh` | the printed version is on the **5.1.x** line, and the value is echoed into the smoke output so every gate run records which binary produced the media; a major-version change fails the row rather than silently invalidating D-11 and B26 | `make smoke` | M12 |
| B27 | Audio is one shared rendition, not a copy per rung | Compose up; the **1280×720** fixture of §5.f, which selects two rungs | transcode it | `master.m3u8` carries a single `EXT-X-MEDIA:TYPE=AUDIO` group that both `EXT-X-STREAM-INF` entries reference, exactly one `audio/` directory exists, and no video variant directory contains an audio-only segment | e2e | M16 |
| B32 | The `ready`-implies-renditions constraint actually bites, on every shape of nothing | a migrated database | `INSERT` `state = 'ready'` with `renditions` at its `DEFAULT '{}'`, then `'{}'`, then `'{NULL}'`, then `'{""}'`, then `'{360p}'` | the first four are refused with `23514` naming **`video_assets_ready_has_renditions`** (an explicit constraint name, not an ordinal `video_assets_check3`), and no row is created; the fifth is accepted | integration `repository_integration_test.go` | M02 |
| B29 | Start-up recovery reclaims a claim of any age, including one whose terminal write never landed | Testcontainers; a `processing` row with a claim token whose `updated_at` is **seconds** old, left that way by a terminal write that failed | run the recovery pass | before it, the row is still `processing` — not `failed`, not `ready`, so no state was invented; after it, the row reads `queued` with a null `claim_token` and is claimable again, and a second `processing` row hours old is reclaimed identically | integration | M06 |
| B30 | A sub-360p source is encoded at its own size, not enlarged | Compose up; the **320×240** fixture of §5.f | transcode it | `master.m3u8` lists one variant at `RESOLUTION=320x240`, and the encoded segments probe as 320×240 — the claim edition 2 asserted and GATE 2 round 2 disproved is now produced by ffmpeg itself | e2e | M16 |
| B31 | A crash between the completion rename and its commit does not lose the source | a ticket whose source is fully written to `source/{id}/original` while the row still reads `uploading` with `received_bytes < declared_bytes` and no scratch file | `PATCH` the final chunk again | `204`; the asset reads `queued` with `received_bytes = declared_bytes`, no byte is appended anywhere, and no `503` is returned | integration | M11 |
| B34 | A crash after the final append and fsync but before its commit is repaired | a ticket whose scratch file holds the complete source while the row still reads `uploading` with `received_bytes < declared_bytes` and no source file | `PATCH` the final chunk again | `204`; scratch is first truncated to `received_bytes`, the retried final chunk lands byte-for-byte at `source/{id}/original`, the asset reads `queued` with `received_bytes = declared_bytes`, and exactly one enqueue is recorded | integration | M11 |
| B35 | An unsupported first chunk consumes the ticket before a waiting valid PATCH can advance it | Testcontainers PostgreSQL; a minted ticket; invalid PATCH A held inside its `ChunkPolicy` after taking the row lock; valid PATCH B waiting on that lock | release A, then observe both responses and the durable state | A answers `400 unsupported_media`; B does not advance the row; the asset reads `failed` with `received_bytes = 0` and reason `unsupported_source`; no source exists and no enqueue was recorded | integration `tus/server_test.go` + repository | M11 |
| B28 | The render argv pins the Apple codec constraints | a 1920×1080 probe | build the render argv | it carries `-profile:v:N high` and `-pix_fmt:v:N yuv420p` on every rung, `-level:v:N 4.1` on the 360p and 720p rungs and **no `-level` on the 1080p rung**, plus `-profile:a:0 aac_low`, `-ac 2`, `-hls_time 6`, `-hls_flags independent_segments`, a `force_key_frames` expression of `expr:gte(t,n_forced*2)` per video stream, and the ladder ceilings 365k/3000k/6000k — and it interpolates no probed frame rate | unit `transcoder_test.go` | M08 |

### Refusals

| # | Scenario | Precondition | Action | Observable result | Evidence | Owner |
|---|---|---|---|---|---|---|
| R01 | Anonymous is refused everywhere | no cookie, valid `Origin` | mint, status read, retry | each answers `401 authentication_required` and no repository method is called | contract | M09 |
| R02 | A student is refused everywhere | student session | the same three | each answers `403 forbidden` | contract | M09 |
| R03 | A non-owning instructor is refused | instructor B; a lesson in a course owned by A | the same three | each answers `403 course_forbidden`, and no asset is created | contract | M09 |
| R04 | Cross-site is refused first | a valid admin cookie, a foreign `Origin` | mint and retry | each answers `403 origin_forbidden` before the cookie is read, with no state change | contract | M09 |
| R05 | An unknown lesson is not found | admin session, an unused uuid | mint | `404 lesson_not_found` | contract | M09 |
| R06 | A malformed id is rejected before lookup | admin session | `POST /api/v1/admin/lessons/not-a-uuid/video` | `400 invalid_request` and no repository call | contract | M09 |
| R07 | A second upload cannot start while one is running | a lesson with an unexpired `uploading` asset | mint again | `409 upload_in_progress`, mapped from `23505`; still exactly one non-terminal asset | contract | M09 |
| R08 | A failed or stranded asset can be retried | admin session; one asset in `failed` and one in `queued` | `POST /videos/{id}/retry` on each | `200` for both; each reads `queued` with `attempt` incremented, and one enqueue was recorded per call | contract | M09 |
| R09 | A ready or freshly-claimed asset cannot be retried | admin session; assets in `ready`, `uploading`, and `processing` **claimed seconds ago** (a stale `processing` row is retryable by T9 — R37) | the same call on each | `409 video_not_retryable` each time; the assets are untouched and no enqueue happens | contract | M09 |
| R10 | An unknown, evicted or expired ticket writes nothing | the tus server; three such ticket ids | `PATCH` each with a valid chunk | each answers `404`; no file is created and no offset is recorded | integration | M10 |
| R11 | A wrong offset is a conflict, not a corruption | a ticket at offset 1 MiB | `PATCH` declaring offset 0 | `409`; the stored file is still exactly 1 MiB and `HEAD` still reports 1 MiB | integration | M10 |
| R12 | An oversized chunk is refused at the edge | Compose up | `PATCH /uploads/{any-uuid}` with a 32 MiB body | nginx answers `413`, and the upload service's request counter for that route did not move | `make smoke` | M13 |
| R13 | `/media/*` stays refused for everyone | Compose up | request `/media`, `/media/`, `/media/x/master.m3u8` anonymously and with an admin cookie | every one answers `403`, before and after a successful transcode | `make smoke` | M13 |
| R14 | A transient failure on the final attempt rests failed | fakes: a transcoder failing transiently, retry count at the maximum | run the job | the asset reads `failed` with an operator-readable reason from the closed set, the reason contains no path and no stderr, and the returned error does not wrap `SkipRetry` | application fake | M04 |
| R15 | A malformed source is never retried | fakes: a probe rejecting the source | run the job | the asset reads `failed` immediately, and the handler returns an error wrapping `asynq.SkipRetry` | application fake | M04 |
| R16 | A retried job cannot corrupt a published asset, and a swept one is not an error | fakes: an asset already `ready`; then an asset the repository reports as absent | run the job in each case | both return success having called neither the transcoder nor publish; the recorded renditions and the storage are untouched | application fake | M04 |
| R17 | A duplicate enqueue is not a failure | a fake queue returning `asynq.ErrTaskIDConflict` | complete an upload, then retry a `queued` asset | both use cases succeed and the asset still reads `queued` | application fake | M04 |
| R18 | Cancellation leaves the asset claimable, not stranded | fakes: a transcoder blocking until the context is cancelled | cancel mid-job | the working directory is removed, no `hls/{id}` key exists, the asset reads **`queued`** with a null claim token, and a later claim succeeds | application fake | M04 |
| R19 | A dependency failure degrades safely | a repository port returning dependency failure | mint and status read | `503 dependency_unavailable`; no SQL, credential, storage key or path appears in the result | application fake | M04 |
| R20 | Cancelling the transcoder ends the process | `/bin/sleep` standing in for ffmpeg | cancel the context | the process is gone within `WaitDelay` and the call returns a cancellation distinct from a failure | unit `transcoder_test.go` | M08 |
| R21 | No user-controlled string reaches ffmpeg | an asset whose `original_name` is `"; rm -rf / --preset x.mp4"` | build the render and thumbnail argv | the produced argv is compared element-by-element against the expected fixed list, which carries `force_divisible_by=2`; the original name appears in neither, and every path element is the asset's uuid or its claim token | unit | M08 |
| R22 | Media errors disclose no filesystem detail | a storage failure forced through the port | mint and status read | the response body's `message` contains no `/media`, no key, no filename and no ffmpeg text | contract | M09 |
| R23 | Business codes reach the frontend domain intact | MSW answering `409 upload_in_progress`, `400 unsupported_media`, `403 course_forbidden`, `500` | call the matching use cases | the first three surface as their named domain outcomes and the last as `dependency_unavailable`; none collapses into a generic failure | adapter with MSW | M14 |
| R24 | An unsaved lesson offers no upload | the editor with a newly added, unsaved lesson | inspect that lesson | no upload control is rendered, a localized hint explains that the course must be saved first, and no request is made | component | M15 |
| R25 | An interrupted upload offers resume or cancel, not a dead lesson | the editor loaded with an `uploading` row and no transfer in this tab | inspect that lesson, then press cancel | the field shows the interrupted state with a resume affordance and a cancel control; cancel issues the tus `DELETE`, and after it the mint control is available again | component | M15 |
| R26 | The client guard is not the mechanism | a student signed in, in the browser | call the mint endpoint directly through `page.request` | `403 forbidden`, though the SPA never renders the control | e2e | M16 |
| R27 | An unsupported codec set is refused by the domain | probes for HEVC+AAC, H.264+Opus, and video with no audio | evaluate acceptance | HEVC+AAC is accepted (Q2); H.264+Opus and audio-less are refused as `unsupported_source` | unit | M03 |
| R28 | Ownership is resolved inside the minting transaction | Testcontainers; instructor B and a course owned by A | mint as B, concurrently with a raw `UPDATE courses SET instructor_id = B` in a second transaction — 4a/4b expose no reassign operation, so the test issues SQL | the mint either refuses B against owner A or admits B against owner B — never a mixed read; the `FOR SHARE` on `courses` serialises the two, and no deadlock occurs against a concurrent 4a save | integration | M05 |
| R29 | Two concurrent PATCHes at one offset do not corrupt the file | a ticket at offset N | issue two identical `PATCH`es concurrently | exactly one answers `204`; the other answers `409`; the stored file's length equals `received_bytes` and equals N plus one chunk | integration | M10 |
| R30 | Cancel frees the lesson immediately | a `created` asset with an unexpired ticket | `DELETE /uploads/{videoId}`, then `PATCH` it, then mint again for the lesson | `204`, then `404`, then `201` — the cancelled row carries `evicted_at` and no longer blocks the pending index | integration | M10 |
| R31 | An enqueue failure after the commit is recoverable | fakes: a queue port returning a dependency failure | complete an upload | the use case returns `dependency_unavailable`, the asset reads `queued`, and a subsequent retry enqueues it successfully | application fake | M04 |
| R32 | A transient failure on a non-final attempt returns the asset to queued | fakes: a transcoder failing transiently, retry count below the maximum | run the job | the asset reads `queued` with a null claim token **before** the handler returns its error, and a following claim succeeds | application fake | M04 |
| R33 | A render that runs out of disk is retryable, not permanent | fakes: a storage port reporting free space below the render threshold, then a render returning `ENOSPC` | run the job in each case | both leave the asset `queued`, remove any working directory, count a `storage_unavailable` failure, and return an error that does **not** wrap `SkipRetry` | application fake | M04 |
| R35 | A worker restarted mid-render leaves nothing stranded | Compose up; the **5-minute** 1280×720 fixture of §5.f uploaded, and the status read **observed to be `processing`** before the action | `docker compose restart media-worker`, then wait | the test records that `processing` was observed first, so the restart genuinely lands mid-render; after it the asset is claimed again and reaches `ready`, never resting in `processing` — the only end-to-end evidence for the BLK-1 closure | e2e | M16 |
| R43 | The container sniff accepts the Q2 set and nothing else | none | evaluate `AcceptContainer` over an `ftyp isom` head, `ftyp qt  `, the EBML magic, an `ftyp` with an unlisted brand, a plain-text head, and a head shorter than `ContainerSniffBytes` | the first three are accepted; the last three are refused as `ErrUnsupportedSource` — a short head is a refusal, not a pass | unit `source_test.go` | M03a |
| R42 | C6 produces a real delivery against a retained task id, and its duplicate is inert | fakes: a `queued` asset older than 10 minutes whose `assetId:attempt` task is already archived, and separately one whose task is still pending | run the worker-start recovery, then deliver both tasks | C6 increments `attempt` first and enqueues under the **new** id in both cases, so the queue records one accepted enqueue each and never reports success on a conflict that produced no task; delivering the duplicate for the second asset returns `nil` through step 1's exhaustive list rather than starting a second render | application fake | M04 |
| R36 | A delivery for an unclaimable row is retried, not consumed | fakes: an asset that is `processing` and not claimable | run the job | the handler returns a **retryable** error and does not return `nil`, so the task is neither marked done nor archived; for a `ready`, evicted or absent asset it still returns `nil` | application fake | M04 |
| R37 | Retry frees a stale claim but never a live one | admin session; one `processing` asset older than the claim age and one claimed seconds ago | `POST /videos/{id}/retry` on each | the stale one answers `200`, reads `queued` with a null `claim_token` and is enqueued; the fresh one answers `409 video_not_retryable` and keeps its token | contract | M09 |
| R38 | Two concurrent mints refuse cleanly whichever index fires | Testcontainers; a lesson whose existing assets are all `ready` | issue two mints for that lesson concurrently | exactly one answers `201`; the other answers `409 upload_in_progress` — never `500` — whether the `23505` names the pending index or `video_assets_generation` | integration | M05 |
| R39 | A sniff-failed asset is not retryable | admin session; a `failed` asset with `received_bytes = 0` and no source | `POST /videos/{id}/retry` | `409 video_not_retryable`; no task is enqueued and the state is unchanged | contract | M09 |
| R40 | A chunked body cannot stream past the chunk cap | a valid ticket | `PATCH` a 9 MiB body with `Transfer-Encoding: chunked` and no `Content-Length` | the tus server's `MaxBytesReader` answers `413`, the scratch file is truncated back, and `HEAD` reports the unchanged offset — the case nginx's `Content-Length` check cannot see | integration | M10 |
| R41 | A stranded row offers a retry instead of polling forever, at the threshold the server honours | the editor with a `queued` row 11 minutes old, a `queued` row seconds old, a `processing` row 41 minutes old and a `processing` row seconds old | inspect each | the 11-minute `queued` and the 41-minute `processing` render an enabled retry control; the two fresh rows render progress with no retry control — the client's thresholds are C6's 10 minutes and T9's 40 minutes, so no rendered control can answer `409` | component | M15 |

Group counts: **11 happy, 35 boundary, 42 refusal = 88 rows.** Manual rows: 0. Every task
except M01 owns at least one row; no row has two owners.

**What round 5 proposed cutting, and what was cut.** The complexity lens named roughly half an
hour and four items. Two are cut: **R34 is folded into B29** (both proved "a `processing` row
becomes `queued` after recovery"; R34's extra setup — killing the pool mid-test — bought an
observable B29 already had, so B29 now carries the "no state was invented" assertion), and
**`content_type` is dropped** (§6). Two are kept, with reasons rather than silence: the
**claim-time disk gate** stays because one `statfs` is cheaper than a doomed thirty-minute
render occupying the single worker (D-15 states exactly that as its justification), and
**B09/B10 stay** because they land at different times — B09 (`make check-generated`) is owned by
M14, the last generating task, so drift is caught when it is introduced, while B10 is the
phase's exit criterion; the repo has carried an equivalent pair since 4a, and folding B09 into a
final evidence column would move the detection to the end. The thumbnail also stays: brief §11
names it in the layout, so unlike `duration_seconds` it is a specified output rather than a
column written for a later phase.

---

## 10. Task cut

Cut along **executable evidence and environment boundaries** (the I06/4a/4b lesson), never
along file or layer labels. Each task is 0.5–2 h, names the environment its evidence needs,
and regenerates its own sqlc/Orval paths (4a NB-9) so trunk is never red at `check-generated`
between tasks.

| ID | Task | Est. | Depends on | Environment / evidence | Owns |
|---|---|---:|---|---|---|
| M01 | ADR 008 (tus ingest: the three-verb surface, the third executable, the ticket capability, the Q1 deviation) and ADR 009 (fMP4/CMAF `.m4s` superseding the segment clause of ADR 005, and the verified `init_N.mp4` naming) | 0.5h | GATE 2 green | the two documents; `.claude/standards.md` Мірник consistency | — (evidence: the ADRs) |
| M02 | Migration `00005_media.sql` with its named CHECKs and three indexes; the course canary fix (real `00005` into the `fstest.MapFS` at `:195`, canary renumbered to `00006_failure.sql` at `:200`) | 1.5h | M01 | Testcontainers PostgreSQL (`make integration`) | B07, B08, B32 |
| M03 | Media domain: `VideoAsset`, the six states and the T1–T9 matrix, `ClaimToken`, `StorageKey`, `Actor.CanEditLesson`, ticket bounds, `SourceProbe`/`AcceptSource`/`SelectLadder`, the error set; and the four port interfaces | 1.5h | — | Go unit tests, stdlib only (`make test`) | B17, R27 |
| M03a | **Revise the M03 domain and port surface to §2.1**: `AcceptContainer` + `ContainerSniffBytes` and `FailureReason.Permanent()` in the domain, with table tests; `ChunkPolicy`/`CompletePolicy`/`PublishPolicy`, `Load`, the two-value `Claim`, `AppendChunk` replacing `BeginUpload`+`Advance`, `DeleteSwept`, `Head` on `MediaStorage`, and `ports.Job`. No adapter exists yet, so nothing downstream is broken by the change | 1h | M03 | Go unit tests, stdlib only (`make test`) | R43 |
| M04 | Application use cases over fakes: mint/append-and-complete/cancel/list/retry, `Process` returning a `TranscodeOutcome`, `Sweep` and `Recover` — retry exhaustion, malformed source, no-op on ready/absent, duplicate enqueue, enqueue-after-commit failure, cancellation, disk exhaustion, dependency failure | 2h | M03a | Go unit tests with fake ports (`make test`) | R14–R19, R31, R32, R33, R36, R42 |
| M05 | Media repository writes: the mint transaction (join with `FOR SHARE OF courses`, `CanEditLesson`, generation assignment, insert, `23505` mapping), the guarded state UPDATEs on `(state, claim_token)`, the list and retry reads | 2h | M02, M04 | Testcontainers PostgreSQL (`make integration`) | P06, B16, B22, R28, R38 |
| M06 | Publish, eviction, recovery and sweep queries: the publish-and-evict transaction, recovery classes C5/C6, reclaim classes C1–C4 | 1.75h | M05 | Testcontainers PostgreSQL (`make integration`) | B06, B14, B24, B29 |
| M07 | Filesystem `MediaStorage`: key→path resolution in an owned root, `os.Root` opens, traversal and symlink refusal, truncate-and-append with `fsync`, atomic publish, purge with ENOENT-is-success, `Free` | 1.5h | M03 | Go integration against a tempdir, no database (`make integration`) | B11, B12 |
| M08 | ffmpeg `Transcoder`: the argv builders — the Apple conformance constants (High/L4.1/yuv420p, AAC-LC stereo, `-hls_time 6`, `independent_segments`, the 2 s `force_key_frames` expression, the audio `agroup`, the 365k/3000k/6000k ceilings) and `force_divisible_by=2`; ffprobe JSON parsing; `exec.CommandContext` with `WaitDelay` | 1.75h | M03 | Go unit tests; process behaviour against `/bin/sleep` — **no ffmpeg needed** (`make test`) | B13, B21, B28, R20, R21 |
| M09 | The API surface: the three operations, DTOs and bounds, the actor mapping, error mapping, `RegisterAdmin` on the 4a seam, `cmd/api` wiring, the OpenAPI paths/schemas/codes and generated output | 2h | M05 | Go contract tests against `openapi.yaml` (`make test`) | P01, P02, P07, B01, B04, B15, R01–R09, R22, R37, R39 |
| M10 | The tus protocol server: `HEAD`/`PATCH`/`DELETE`, the D-2d write protocol (row lock, offset equality, truncate-to-column repair, fsync, guarded advance), the `MaxBytesReader` chunk cap, termination | 2h | M07, M09 | Testcontainers PostgreSQL (`make integration`) | B02, B03, R10, R11, R29, R30, R40 |
| M11 | Completion and the third executable: the atomic first-chunk sniff refusal, the first-`PATCH` disk gate, the rename into `source/`, commit-then-enqueue and both crash-repair branches; `cmd/upload/main.go` with its readiness write probe; the Dockerfile `upload` stage | 1.75h | M10 | Testcontainers PostgreSQL + a tempdir storage adapter + fake queue (`make integration`); the image stage by `make build`, which is in the gate | P03, B18, B23, B31, B34, B35 |
| M12 | Worker wiring: the media handler on the mux, the payload, the claim guards and the step-2 retryable branch, the `ErrorHandler`, `Recover`-then-`Sweep` before `server.Start`, the two metrics with pre-initialised label values, `/status`; the `smoke.sh` handler assertion inverted and the `ffmpeg -version` assertion added | 1.5h | M06, M08, M11 | Compose (`make smoke`) | P10, B19, B33 |
| M13 | Edge and topology: the `upload` Compose service and its limits, nginx's `upload` zone and `/uploads/` location, `nginx.depends_on`, the remaining `smoke.sh` assertions, README | 1h | M12 | Compose (`make smoke`) | R12, R13 |
| M14 | Frontend adapter and application: domain types, the port including `cancelUpload`, zod schemas, the Orval-backed repository with business-code mapping, the hand-written tus client and its transport, `admin.ts` extension, the bounded poll predicate; regenerates Orval | 1.75h | M09 | Rstest with MSW (`make test`), `make check-generated` | P11, B09, R23 |
| M15 | `LessonVideoField` and its mount inside `CourseEditorView`: the file control, progress, state badges by the D-12 rule, failure reason and retry (including the stale-`queued` retry affordance), the unsaved-lesson hint, the interrupted-upload resume/cancel affordance, degraded states, uk/en, accessibility | 2h | M14 | Rstest component tests including axe | P09, B20, R24, R25, R41 |
| M16 | The browser journey: `media.setup.ts` generating the four clips into `/tmp` (three short, one 5-minute), `media.spec.ts` (upload → processing → ready, the one-rung and sub-360p ladders, the playlist assertions against the target binary, the shared audio rendition, the histogram scrape, the restart into an observed `processing`, the direct-API refusal), the playwright `setup` `testMatch` change | 1.75h | M13, M15 | Playwright against Compose (`make e2e`) | P04, P05, P08, B05, B25, B26, B27, B30, R26, R35 |
| M17 | Ordered gate and report: full `make check` from a clean checkout; `docs/media/report.md` with measured results, limitations and the next task; `AGENTS.md`'s current-delivery line moved to Phase 5 | 1h | M01–M16 | `make check`, twelve steps | B10 |

**Total: 28.25 h** — the arithmetic of the column above: five tasks at 2 h (M04, M05, M09,
M10, M15), five at 1.75 h (M06, M08, M11, M14, M16), four at 1.5 h (M02, M03, M07, M12), three
at 1 h (M03a, M13, M17) and one at 0.5 h (M01). **Eighteen tasks** — M03a was added after M04
stopped on contract shape (§16); it is the only task inserted after implementation began.

**Which tasks need Docker:** M02, M05, M06, M07, M10, M11 need Testcontainers
(`make integration`); M12, M13 need Compose (`make smoke`); M16 needs Compose (`make e2e`);
M17 needs everything. M01, M03, M03a, M04, M08, M09, M14, M15 need no Docker.

**Which tasks need a real ffmpeg run:** only M16, and only inside the worker container. **None**
of M03–M15 executes ffmpeg, by construction (§5.e).

**How a worker job is proven, and where the line is.**

- In `make integration`: everything about the job **except the subprocess** — the state machine
  and its guards (M04, fakes), the mint and the guarded transitions (M05, real PostgreSQL),
  publish/evict/recover (M06), the atomic rename and traversal refusals (M07, real filesystem),
  the tus offsets and the chunk cap (M10), the completion path and the disk gate (M11).
- In `make test`: the exact ffmpeg argv including `force_divisible_by=2` (M08, R21, B21) and
  process termination on cancellation (M08, R20).
- In `make smoke`: that the handler is registered (M12, P10) and that both metric families and
  all their label values exist at zero *before* any job has run (M12, B19) — the check that
  would have caught "the worker starts but does nothing".
- In `make e2e`: the four things only a real ffmpeg and a real container can show — a real clip
  becoming a real master playlist with real segments and a thumbnail (M16, P04/P05/P08), what
  the muxer actually writes into the playlists (B05, B26, B27, B30), the histogram observing a
  completed job (B25), and **a worker restarted mid-render leaving nothing stranded** (R35),
  which is the only place Asynq's real shutdown sequence can be exercised at all.

**Sequencing notes.**

- **M01 first** because `.claude/standards.md` makes ADRs a Мірник document: writing code that
  contradicts ADR 005's segment choice before ADR 009 supersedes it is a standards violation,
  not a style one.
- **M02 early** because every later Go task runs against the migrated schema, and the canary
  renumber must land with the migration or `make integration` is red between tasks.
- **M03a sits between M03 and M04** because M04 stopped against M03's surface rather than
  writing around it (§16). It is a contract revision with no adapter behind it yet, so it costs
  an hour and breaks nothing; had M04 improvised the missing shapes, M05–M07 would have
  inherited them.
- **M04 split from M05/M06 by evidence environment, not by layer**: M04 is everything provable
  with fakes in `make test`; M05 and M06 are what needs real PostgreSQL, split again because
  the mint path and the publish/recovery path are two distinct bodies of SQL and one 3.75 h
  task would breach the ceiling. M14/M15 split the same way (MSW versus a rendered component).
- **M09 before M10 and M14** — the upload service consumes the ticket the API mints, and Orval
  cannot regenerate against a document that does not describe the operations.
- **M07 before M10** because the tus server writes through `MediaStorage`, not `os.OpenFile`.
- **M10 split from M11** on the same evidence-environment rule the rest of the cut follows, and
  because edition 2's single upload task had grown to eleven rows at the two-hour ceiling. M10
  is the protocol — offsets, concurrency, repair, the chunk cap — provable with a database and
  a scratch directory; M11 is what happens *after* the last byte — sniff, disk gate, rename,
  enqueue — plus the executable and its image stage.
- **M12 before M13** so the handler exists before the edge routes to it; both before **M16**,
  the only task that can observe the whole path.
- **M17 last**, and it is the task that moves `AGENTS.md`'s current-delivery line — the phase is
  not "current" until its gate is green.

---

## 11. Riskiest decisions — pointer list

1. The upload ticket is a database row, not a signed token and not a callback (D-2b).
2. tus core is implemented here rather than running tusd, with no `creation` extension — the
   JSON mint is the only creation point (D-2a).
3. `received_bytes` is the offset truth; the file is truncated to it, and every advance is one
   guarded UPDATE under a row lock (D-2d).
4. Media models a capability, not a role, so 4a's "third module" trigger does not fire (D-3).
5. The ownership join lives inside the mint transaction in the repository — no second port —
   and takes the `courses` row first, as 4a does, so the two cannot deadlock (D-1).
6. The lesson↔video FK is `ON DELETE SET NULL` and swept, not `CASCADE` (D-4).
7. `claim_token` makes a re-delivered Asynq attempt provably unable to publish over a newer
   claim, and the claim is verified under the row lock before any byte under `hls/` is touched
   (D-8, D-10 step 10).
7a. Recovery has three legs because Asynq's real shutdown gives the handler no reliable window:
   T6 when the attempt can write, a **threshold-free** T7 at worker start (safe only because
   there is exactly one worker — trigger: a second replica), and T9 for the process that keeps
   running with a dead claim. Step 2 returns a retryable error rather than `nil`, so a delivery
   is never consumed for a row nobody owns (D-7, D-10).
8. Commit-then-enqueue, with `Retry` accepting `queued` and recovery class C6 — which bumps
   `attempt`, because asynq holds a task id for the archive's lifetime — as the two exits
   from a stranded asset (D-10, D-15).
9. Q1's "notifies the API" is deviated from — the upload service writes state and enqueues
   directly through shared ports (D-2b); Q2's container/codec split is **settled**, accepted by
   the gate owner on 2026-09-06 and amended into `spec.md:67` (D-8).
10. Every scale expression is the `min(target, source)` form plus `force_divisible_by=2`, so a
    sub-360p source yields exactly one rendition **at its own size** — the point edition 2 got
    wrong, and the reason B30 now proves it with ffmpeg rather than prose (D-11).
10a. **Phase 5 targets the ffmpeg the image ships (5.1.9 from `debian:bookworm-slim`)**, not a
    developer's; a `trixie-slim` worker stage is declined with Phase 6's player as its trigger,
    and B33 asserts the binary in `make smoke` so the evidence base cannot drift from the
    deployment again (D-11, §14 N-17).
11. The ladder and codec constraints come from the Apple HLS Authoring Specification via
    `research-codecs.md` (High/8-bit 4:2:0 with `-level` on the two lower rungs only, AAC-LC
    stereo, 6 s segments with 2 s IDR, separate audio, 365/3000/6000 kbit/s under capped CRF).
    **Three Apple non-conformances are accepted and asserted**: `FRAME-RATE` (9.15),
    `AVERAGE-BANDWIDTH` (9.14) and nominal `BANDWIDTH` (9.13/1.26) are all absent or nominal on
    the target binary, no Phase 5 consumer reads a playlist, and B26 pins them so the trigger
    fires by itself; the remedy is the image bump or a line-append, never a rewrite (D-11).
12. AV1 and HEVC are excluded on evidence, with §24.14's three-condition trigger recorded and
    fMP4 chosen precisely so a future AV1 rung needs no remux (D-16).
13. A full or read-only volume is gated at the **first PATCH** and at claim — never at the
    mint, because the API has no media mount — answered per-request with `507`/`503`, and
    surfaced through readiness, though nginx keeps routing to an unready upstream (D-15).
14. No Go test may depend on ffmpeg; the real transcode is proven only in Compose (§5.e).
15. A rollback leaves in-flight jobs unhandled and published bytes unswept until roll-forward
    (§7).

---

## 12. Disposition of the 16 non-blocking findings

| # | Finding | Disposition | Where |
|---|---|---|---|
| NB-1 | §0/§5.j/§11-0/M01/B21 describe a defect `5ef6c1f` already fixed | **Addressed** — §0 records it as found-and-fixed; §5.j is gone; §11 item 0 removed; M01 is the two ADRs and owns no row; B21 is now the odd-dimension argv row | §0, §2, §10 |
| NB-2 | The Q1 "notifies the API" deviation is unnamed | **Addressed** — named as a deviation with its rationale and its revert trigger, and listed in §11 | D-2b, §11.9 |
| NB-3 | The Q2 deviation needs the gate owner's signature; the sniff belongs on the first chunk | **Addressed, now fully** — the sniff runs on the first chunk, and the gate owner accepted the container/codec split on 2026-09-06, amended into `spec.md:67`; the plan states it as settled | D-8, §11.9 |
| NB-4 | `Setpgid` + process-group kill is machinery ffmpeg does not need | **Addressed** — dropped; `exec.CommandContext`'s default `Cancel` plus `WaitDelay`; R20 reworded | D-10, R20 |
| NB-5 | `duration_seconds` has no Phase 5 reader | **Addressed** — removed from 00005; Phase 6 adds it with its consumer | §6 step 1 |
| NB-6 | The `content_type` CHECK refuses legitimate mkv/mov | **Superseded by §14 N-26** — the column was first relaxed to nullable display metadata, then dropped entirely once round 5 observed that nothing reads it; the first-chunk sniff is the only container check | §6 step 1, D-8, §14 N-26 |
| NB-7 | Several rows per lesson: the badge rule is undefined and `failed` assets are never swept | **Addressed** — publish counts non-evicted rows and evicts the lowest `ready` generation; the badge order is stated; sweeper class C4 reclaims `failed` after 30 days; P07's title is "one row per asset" | D-5, D-12, D-7, P07 |
| NB-8 | `generation` assignment is not defined | **Addressed** — `MAX+1` over all rows including evicted, inside the mint transaction, with `UNIQUE (lesson_id, generation)`; B22 is the row | D-5, §6 step 3, B22 |
| NB-9 | Sweeper idempotency and row-less orphans | **Addressed** — ENOENT-is-success stated; the e2e clip is generated into `/tmp` in the worker container; the upload scratch key is named so the `temporary/{id}-*` glob covers it | D-7, §5.f, D-8 |
| NB-10 | The step-0 sweep can delete the asset the handler was enqueued for | **Addressed** — not-found returns `nil`, and a `lesson_id IS NULL` asset is skipped at claim | D-10 step 1, R16 |
| NB-11 | R12 asks `make smoke` for a valid ticket it cannot mint | **Addressed** — R12 is trimmed to the edge-only 413 with no ticket; the offset half moved to B02 | R12, B02 |
| NB-12 | `video_assets_sweepable` indexes a column the sweeper does not order by, and misses a class | **Superseded by N-15** — the sweeper indexes this finding reshaped were then dropped entirely; the predicates scan, with the trigger "past ~10⁵ rows" | §6, §13 N-15 |
| NB-13 | The ticket is also the public asset id | **Addressed** — stated as a decision: the id is visible to everyone who could have minted it, and a co-editor cancelling or appending is in scope | D-2b |
| NB-14 | Ladder selection by height alone on portrait sources | **Addressed** — the rule is stated as "by height, portrait accepted as is"; B17 gains portrait and sub-360p probes | D-11, B17 |
| NB-15 | Filler and stale rhetoric | **Addressed** — "five years late" and the D-2b self-congratulation are cut; §11 is a bare pointer list | §5.g, D-2b, §11 |
| NB-16 | M09 owned the Dockerfile stage with evidence that cannot build an image | **Addressed** — the stage stays with the executable (now M10) and its evidence column names `make build`, which is in the gate | M10 |

Retained from round 1: **none**. NB-3's signature half closed when the gate owner accepted the
split on 2026-09-06.

---

## 13. Disposition of the GATE 2 round 2 findings

Round 2 confirmed seven of the eight edition-1 blocking findings closed, left **BLK-1's
crash/restart half open**, and added three blocking and twelve non-blocking findings. Three of
the four blocking items were established by **execution** — ffmpeg on four generated sources
(on the host's 7.1.1, which round 5 showed was the wrong binary; §14 N-17), the asynq shutdown and lease paths read in `processor.go`/`server.go`, and the
`api` service's mounts in `docker-compose.yml` — so they are corrected, not argued.

### Blocking

| # | Finding | Disposition | Where |
|---|---|---|---|
| BLK-1 residual | A graceful restart or crash strands the row in `processing`: T6 never runs, the redelivered task is consumed as `nil`, and C5's 40-minute threshold hides the row | **Closed** — C5 is threshold-free at worker start on a stated single-worker assumption and runs before `server.Start`; step 2 returns a retryable error instead of `nil`; T9 lets the operator retry a stale claim for the case where the process keeps running; D-10 now describes Asynq's real shutdown order and accepts T6 as best-effort | §1 T6–T9, D-7 C5, D-10 steps 2 and 12 + "Cancellation" + "Retry", B29, R35, R36, R37 |
| N-1 | The sub-360p no-upscale claim is false — the filter encodes 320×240 at 480×360, and R21 would have pinned it | **Closed** — every scale expression becomes the `min(640,iw)`/`min(360,ih)` form (and the thumbnail with it); the D-11 bullet is rewritten to record the disproof and the corrected, re-verified behaviour; defense row 4's closure now rests on the filter that actually holds | D-11 (ladder bullet, argv, thumbnail), B21, B30 |
| N-2 | The mint-time disk gate runs in the API, which has no `media-data` mount | **Closed** — the gate moves to the first `PATCH` in the upload service, which holds both the volume and `declared_bytes`, answering `507` before a byte is written; the claim-time gate stays; the API is stated to wire no storage adapter at all, so the mount question cannot return | D-15, §1 T1, §2 (`ports/storage.go`, `cmd/api/main.go`), §5.a, §8, B23 |
| N-3 | B26 is red on the real muxer: no `FRAME-RATE`, and `EXT-X-INDEPENDENT-SEGMENTS` sits in the variant playlists | **Reopened at round 5 and superseded by §14 N-17** — the *decision* (take what the muxer produces, build no playlist writer) stands, but the table and the row under it described the host's ffmpeg rather than the image's, so the closure is §14 N-17's, not this one's | D-11 "Master playlist conformance", B26, §11.11, §14 N-17 |
| D-1 (defense, edition 3.1) | The `ready ⇒ renditions ≥ 1` CHECK does not enforce anything: executed against PostgreSQL 17, an `INSERT` with `state='ready'` and the default `'{}'` succeeds, because `array_length('{}',1)` is `NULL` and a `NULL` CHECK is satisfied | **Closed** — the constraint becomes `cardinality(renditions) >= 1`, §6 states the empty-array/`NULL` trap as the reason for the form, every other CHECK in the migration is audited against both shapes of the defect (one nullable-column guard, one `NULL`-returning function) and found clean, and the repository's existing migrations are grepped and clear | §6 step 1, B32 |

### Non-blocking

| # | Finding | Disposition | Where |
|---|---|---|---|
| N-4 | A stale claim can still remove a newer claim's published directory | **Addressed** — the publish transaction verifies `claim_token` under the row lock before touching `hls/` | D-8, D-10 step 10 |
| N-5 | Two concurrent mints may fail on `video_assets_generation`, not the pending index | **Addressed** — `23505` is mapped by `ConstraintName` and both index names map to `upload_in_progress` | D-1 step 4, R38 |
| N-6 | `created → failed` is missing from the matrix | **Addressed** — T8 covers `created`/`uploading`/`processing`, with the sniff case named | §1 T8, D-8 |
| N-7 | Retrying a sniff-failed asset enqueues a job for a source that was never written | **Addressed** — retry requires `received_bytes = declared_bytes` in every accepted state | D-10 "Retry", §1 T9, R39 |
| N-8 | A crash between the completion rename and its commit gives a permanent `503` | **Addressed** — PATCH step 3 gains the idempotent branch: scratch missing but source present at full length → skip the append, run the terminal UPDATE | D-2d step 3, B31 |
| N-9 | A chunked PATCH streams past nginx's `Content-Length` cap | **Addressed** — the tus server wraps the body in `http.MaxBytesReader` and answers `413` itself | D-2d step 4, §5.c, R40 |
| N-10 | A stranded `queued` row polls forever with no retry in the UI | **Addressed** — the field renders the retry control for `failed` and for `queued`/`processing` older than the stale threshold | D-12/§2 `LessonVideoField`, M15, R41 |
| N-11 | `-level 4.1` on a 1080p60 source declares a level the stream exceeds | **Addressed** — `-level` is set on the 360p and 720p rungs only; Apple 1.3a asks for *some* variants, not all | D-11 codec constraints, B28 |
| N-12 | B27 needs a two-rung source; §5.f generates one clip | **Addressed** — three clips are generated (640×360, 1280×720, 320×240), each tied to the rows that need it | §5.f, B27, B30 |
| N-13 | Stale "awaiting acceptance" wording on a settled Q2 | **Addressed** — stated as accepted by the gate owner, `spec.md:67`, in all three places | D-8, §11.9, §12 NB-3 |
| N-14 | The hours parenthetical double-counts in words | **Addressed** — "five at 2 h, five at 1.75 h, four at 1.5 h, two at 1 h, one at 0.5 h", recomputed for seventeen tasks | §10 total |
| N-15 | Three sweeper indexes on a table bounded to hundreds of rows | **Addressed** — the three scan indexes are dropped; the two unique indexes and the hot read index stay, with the trigger "past ~10⁵ rows" | §6 |
| N-16 | Two disk gates plus a playlist writer where one gate and no writer would do | **Addressed** — folded into N-2 and N-3: one gate at the first PATCH plus the claim gate, and no writer at all | D-15, D-11 |

### Late defense findings confirmed already closed in edition 3

The same review raised two further items that edition 3 had already closed for other reasons.
Re-checked against the text rather than assumed:

| Finding | Already closed at | Confirmation |
|---|---|---|
| The tus completion renames the source into place before the commit, with no reconciliation branch if the process dies in between — the ticket is then permanently `503` | **D-2d step 3** (N-8) and **B31** | The step-3 branch is stated as "scratch missing **and** `source/{id}/original` present at `declared_bytes` → skip the append, run the terminal UPDATE to `queued`", and B31's precondition is exactly that state with its expected result `204` and `queued`. Covered; no edit needed |
| Nginx's `client_max_body_size` is checked against `Content-Length`, so a chunked request streams past the edge cap | **D-2d step 4** (N-9), **§5.c** and **R40** | Step 4 wraps the body in `http.MaxBytesReader(w, r.Body, chunkBytes)` and answers `413` itself; §5.c says in as many words that a chunked request "has no `Content-Length` to check and streams past it"; R40 sends a 9 MiB chunked body with no `Content-Length` and expects `413` with the offset unmoved. Covered; no edit needed |

Retained from round 2: **none**.

---

## 14. Disposition of the GATE 2 round 5 findings

Round 5 was the first **critique** pass on editions 3/3.1 (rounds 3 and 4 were defense). It
confirmed BLK-1, N-1 and N-2 closed — N-1 now on the container's binary as well — and
**reopened N-3 as N-17**.

### Blocking

| # | Finding | Disposition | Where |
|---|---|---|---|
| N-17 | Every ffmpeg fact was executed on the host's 7.1.1 while the worker image ships bookworm's 5.1.9; there `AVERAGE-BANDWIDTH` is absent entirely and `BANDWIDTH` is nominal, so B26 is red on the deployment binary — and on 7.1.1 the 360p ceiling comparison is red too (370-378 k against a 365 k ceiling) | **Closed by targeting one binary and asserting what it writes.** Phase 5 targets **5.1.9 from `debian:bookworm-slim`**; a `trixie-slim` worker stage was weighed and declined with its trigger (Phase 6's player needing measured bandwidths), because no Phase 5 consumer reads a playlist and it would split a base image shared by three executables for one metadata field. The recorded non-conformance widens from `FRAME-RATE` alone to **`FRAME-RATE` (9.15), `AVERAGE-BANDWIDTH` (9.14) and nominal `BANDWIDTH` (9.13/1.26)**; the ceiling comparison is deleted; every fact names its binary and where it ran; B33 asserts the binary in `make smoke` | §0 facts table, D-8, D-11 "target binary" + conformance table, B26, B33, §5.d |

### Non-blocking

| # | Finding | Disposition | Where |
|---|---|---|---|
| N-18 | C6 re-enqueues under a task id asynq still holds, so it is a mapped-to-success no-op on two paths | **Addressed** — C6 increments `attempt` first and enqueues under the new id; D-15 and §7 scope C6 to the BLK-7 window; step 12 writes `Fail(render_failed)` on a final-attempt cancellation instead of `queued` forever | D-7 C6, D-10 step 12, D-15, §7, R42 |
| N-19 | `updated_at >= created_at` plus `DEFAULT now()` rejects a clock read before `Begin` | **Addressed** — both defaults dropped, so the mint must write both columns from one application-clock value as 4a does; the omission becomes a query error rather than a runtime `23514` | §6 step 1 |
| N-20 | The fixed CHECK counts elements, so `{NULL}` and `{""}` pass | **Addressed** — the constraint gains `array_position(renditions, NULL) IS NULL` and `'' <> ALL (renditions)`; B32 inserts all four empty shapes | §6 step 1, B32 |
| N-21 | B32 has no stable constraint name to assert | **Addressed** — the three load-bearing CHECKs are explicitly named; B32 asserts `video_assets_ready_has_renditions` | §6 step 1, B32 |
| N-22 | R35 cannot restart "mid-render" against 2-second clips | **Addressed** — §5.f generates a fourth, 5-minute 1280x720 clip for R35 alone (~45 s of render), and R35 must observe `processing` before restarting | §5.f, R35 |
| N-23 | M09's task text still says "the disk gate" | **Addressed** — removed; M11 carries it | §10 M09 |
| N-24 | The UI's stale threshold is never numbered, and the server has two | **Addressed** — `queued` 10 min (C6's constant), `processing` 40 min (T9's claim age), stated once and used by R41 | D-12, R41 |
| N-25 | Stale counts and meanings (index counts, NB-12, §8 rows, backoff numbers, R09) | **Addressed** — "three indexes" in §2 and M02; NB-12 marked superseded by N-15; `upload_too_large` and `video_not_retryable` reworded; backoff stated as 15-44 / 16-74 / 31-118 s; R09's `processing` asset is "claimed seconds ago" | §2, §6, §8, §12, D-10, R09 |
| N-26 | Two columns nothing reads, and an unspecified mint body | **Addressed** — `content_type` dropped; `original_name` kept and now returned by the course-videos read, so it is genuinely display metadata; the mint body is specified once as `{declaredBytes, originalName}` | §1, §6 |
| N-27 | Four rows and one gate could go | **Addressed in part; two declined on the record** — R34 folded into B29 and `content_type` dropped; the **claim-time gate is kept** (one `statfs` beats a doomed 30-minute render occupying the single worker — D-15 now states that as its justification) and **B09/B10 are kept** (they land at different times: B09 is owned by M14, the last generating task, so drift is caught where it is introduced, while B10 is the phase's exit criterion). The thumbnail stays because brief §11 specifies it as an output, unlike `duration_seconds` | §9 preamble, D-15, B29 |
| N-28 | Decision sections narrate their own edition history | **Addressed in part** — the longest passages in D-7, D-10, D-11, D-15 and §6 are cut to one sentence plus a §13/§14 pointer; the reasons stay, the archaeology moves | D-7, D-10, D-11, D-15, §6 |
| N-29 | R28 names a reassign operation the product lacks | **Addressed** — the row now issues a raw `UPDATE courses SET instructor_id` in a second transaction | R28 |

Retained from round 5: **none**, with two proposals declined on the record (N-27).

---

## 15. Disposition of the GATE 2 round 6 findings

Round 6 was the closing verification and returned **GREEN**: N-17 closed by execution in the
worker container, the tightened CHECK verified against PostgreSQL 17.11, **no new blocking**.
Its four non-blocking findings are folded here.

| # | Finding | Disposition | Where |
|---|---|---|---|
| N-30 | Step 12's "final attempt ⇒ `Fail(render_failed)`" keys on any cancellation, so a graceful restart during a final-attempt render could write `failed` for a healthy source; and steps 1/2 did not enumerate a `failed` row | **Addressed** — the `Fail` branch keys on the **cause**, `errors.Is(ctx.Err(), context.DeadlineExceeded)`; a plain `context.Canceled` keeps `Requeue`. Steps 1 and 2 now carry an **exhaustive** nil list (`ready`, `failed`, evicted, absent, `lesson_id IS NULL`, `created`, `uploading`), so only `queued` and `processing` continue and no state falls through | D-10 steps 1, 2, 12 |
| N-31 | C6's `attempt` bump enqueues a duplicate task per stale `queued` row when a restart finds a backlog whose tasks are still pending | **Named, not solved** — D-7 states the trade (3.1's same-id form was perfect on a backlog and silent on an archived task; 3.2's inverts it, and a delivery that happens twice beats one that never happens), and explains why the duplicate is inert: `Concurrency: 1` plus step 1/2's exhaustive list absorb it. Trigger to spend a second Redis call: duplicates becoming visible in the queue-depth gauge. R42 asserts the duplicate is a no-op | D-7 C6, R42 |
| N-32 | The 5-minute fixture is 77 MB, not "a few MB", and R35 costs 2–2.5 min, not "roughly a minute" | **Addressed** — `-crf 32` on the long clip only (render time is set by frame count, so R35's window is unchanged; the fixture drops to ~8 MB), and the measured numbers replace the estimates: 39 s of render, 93 MB of HLS output, ~2–2.5 min for the row | §5.f |
| N-33 | Four wording residues: "compile-time error" for a runtime `23502`; "every table-level CHECK is named" while one was auto-named; stale §12 NB-6; §13's N-3 row without its §14 pointer | **Addressed** — the defaults paragraph now says "a deterministic `23502` … not a compile-time error, sqlc does not verify `NOT NULL` coverage"; the promoted two-column CHECK is named `video_assets_received_within_declared`, making the sentence true; NB-6 marked superseded by §14 N-26; the N-3 row marked reopened at round 5 and superseded by §14 N-17 | §6 step 1, §12 NB-6, §13 N-3 |

**One precision carried into M16 at the verifier's request**, because it is the kind of detail
an implementer smooths away: `audio/index.m3u8` carries **no** `EXT-X-INDEPENDENT-SEGMENTS`
either — 5.1.9 writes the tag only into playlists with a video stream. B26's "every **video
variant** playlist" is exact, and generalising it to "every variant playlist" turns the row red
on the audio rendition. Stated in B26 and in D-11's conformance table.

Retained from round 6: **none**.

---

## 16. Amendment after M04 — port contracts made expressible

M01–M03 landed green. **M04 stopped without writing code** and reported six contract-shape gaps
(`work/report-m04.md`); the report is correct, and the fault is this plan's, not the
implementer's. §2's port table listed **method names**, M03 implemented exactly those names, and
names carry no signature: three of the sequences D-7, D-8, D-10 and D-15 require must run
*inside one transaction between the row lock and the guarded write*, and separate port calls
cannot order them however they are sequenced.

| Gap | Closed by | Where |
|---|---|---|
| 1. The first-`PATCH` disk gate cannot leave the row in `created`: nothing reads `declared_bytes`, and `BeginUpload` writes T2 with no policy hook | `AppendChunk(…, ChunkPolicy)` replaces `BeginUpload`+`Advance`; the policy sees the locked asset (so it has `DeclaredBytes`), runs `Free`, the sniff and the append, and an error aborts before T2 | §2.1, D-15 |
| 2. Sweep cannot purge bytes before deleting the row | `SweepCandidates` deletes nothing; the new `DeleteSwept(id)` is called only after a successful purge, so a purge failure leaves a retryable row | §2.1, D-7 |
| 3. Publish cannot validate `(processing, claim_token)` before touching `hls/` | `Publish(…, PublishPolicy)`; the adapter locks, verifies the token, and only then invokes the closure that removes a stale destination and renames | §2.1, D-8, D-10 step 10 |
| 4. No load before claim, and `Claim` reports no disposition | `Load(id)` added; `Claim` returns `(asset, claimed bool, error)`, giving step 2 its re-read in the same round trip | §2.1, D-10 steps 1–2 |
| 5. The container sniff has no domain rule | `AcceptContainer(head []byte) error` + `ContainerSniffBytes`, with the accepted `ftyp` brands and the EBML magic named | D-8, R43 |
| 6. No inward permanent-failure vocabulary; application may not import asynq | `FailureReason.Permanent()` in the domain, `TranscodeOutcome` from `Process`, and a four-row mapping table in the Asynq adapter — the only place `asynq.SkipRetry` appears | §2.1, D-10 |

**What this cost:** one new task, **M03a (1 h)**, placed between M03 and M04, owning **R43**.
Total 27.25 h → **28.25 h**, 17 tasks → **18**, 85 rows → **86**. No existing row changed owner,
no decision was reversed, and no schema or slice shape moved — every gap was a missing
*signature* for behaviour the plan already specified in prose.

**The lesson, recorded because it is cheap to repeat.** §2's placement table is written to answer
"what files exist", and for twelve editions that was enough. It is not enough for a port whose
methods must compose into an atomic sequence: there, the contract *is* the signature, and a table
of names reads as complete while under-specifying exactly the part the decisions depend on. 4b
had already solved this once — `StatusPolicy`/`DeletePolicy` in
`identity/ports/administration.go` — and this plan cited that file as a reuse precedent without
carrying its shape across. The general rule for later phases: **when a decision says "inside the
same transaction", the port table owes a signature, not a name.**

## §17 — Amendment, 2026-09-06 (gate owner, during M08)

M08 stopped before implementing the ffmpeg adapter and reported a contract gap, correctly.
D-11's thumbnail branch seeks to one second, or to zero when the source is shorter than
that — but `domain.SourceProbe` carried no duration and `ports.Transcoder.Thumbnail`
received no probe, so the adapter could not decide the argument. Always emitting `-ss 1`
would break the sub-second branch; a private second probe inside the adapter would be a
workaround.

Repair applied (forced, single sensible shape — a repair, not a design fork):

- `domain.SourceProbe` gains `DurationSeconds float64`, documented as existing for the
  thumbnail seek.
- `ports.Transcoder.Thumbnail` gains a `domain.SourceProbe` parameter.
- `application.Transcode.Process` threads the probe it already holds into the call.

Note for later phases: `duration_seconds` was dropped in an earlier gate round as a field
with no consumer. That was true of the *asset row* and false of the *probe* — a decision in
this same document needed it. Removing a field for want of a consumer is right; the check
should be "no consumer **in this document**", not "no consumer in the schema".

A second item was a briefing error, not a plan defect: the M08 brief said cancellation
"kills the process group", while D-10 records the opposite — `exec.CommandContext`'s
default direct-process SIGKILL plus `WaitDelay`, explicitly **no `Setpgid`, no
process-group signalling**. The brief was corrected; D-10 stands.

## §18 — Amendments, 2026-09-06 (gate owner, during M09)

M09 opened the first `## Discussion` under the two-channel protocol
(`work/codex-working-model.md`). All three items were accepted.

**1. R22's precondition was impossible as written.** It asked for "a storage failure through
the port" while calling mint and the course-videos read, but neither operation touches
`MediaStorage` (`application/upload.go` — `MintTicket` calls only `repository.Mint`;
`ListCourseVideos` only `repository.ListCourseVideos`), and D-15 forbids storage in the API
outright. Following it literally would have wired a storage adapter into `api` and shipped
the very boundary violation the plan refuses. R22's precondition is amended to: **a
dependency failure carrying filesystem-looking private detail, forced through the
repository/application seam** — the row's real subject was always that the fixed 503 body
discloses no path, filename or ffmpeg text, not which port produced the fault.

**2. The API had to pass `nil` storage, which made D-15 conventional rather than
structural.** `NewUpload(repository, storage, queue, now)` demands a storage slot, and the
byte methods dereference it; the API is safe today only because it happens to call just
three methods. Accepted correction: **split construction so the API cannot receive
byte-handling methods at all** — the ticket/list/retry operations get their own type over
repository and queue, and the chunk/complete operations keep storage. D-15 then holds
because the type system says so, not because a caller is careful.

**3. `adapters/asynq/queue.go` had no owning task.** The plan named the producer and its
D-10 options but assigned the file to no slice, while M09's retry endpoint cannot work
without it. **M09 owns the producer**; M12 owns the consumer wiring, unchanged.

## §19 — Amendments, 2026-09-06 (gate owner, during M10)

M10's `## Discussion` raised three items; all accepted. The second is the most serious
defect this plan has carried.

**1. D-2d step 3, as written, could never start an upload.** The sentence said a missing
scratch file with no completed source is a `503`. But a freshly minted `created` row has no
scratch file *by design* — the first `PATCH` is what creates it — so following the sentence
literally would have refused **every upload at its first byte**, and the feature would not
have worked at all. Amended: **a missing scratch file at `received_bytes = 0` is initialised
at zero** (the storage adapter's safe zero-length initialiser); a missing or short scratch
file at a positive truth-column offset remains a `503`. The implementation follows the
amended rule, and an application test pins the short-file refusal at a positive offset.

**2. The scratch file could get ahead of the truth column (B02).** A chunk beyond the
declared length was appended and fsynced before the repository refused it; the transaction
rolled the column back, nothing rolled the file back, and the next request began with a file
longer than the column — breaking the "column is truth" invariant the whole protocol rests
on. Amended: the declared length is a **second cap inside the transaction policy**, and the
file is truncated to the locked column before the refusal returns.

**3. A failed compensating truncate was discarded.** After a failed `Append`, `Head` or
container check, the error from the compensating `Truncate` was ignored, so a failed rollback
could leave uncommitted bytes while the handler reported an ordinary resumable refusal. The
error is now surfaced.

Method note: none of the three was a missing contract, and all three could have been
silently worked around with code that runs. They surfaced because the brief gives Codex a
discussion channel as well as a stop.

## §20 — Amendment, 2026-09-06 (gate owner, during M11)

M11 stopped and reported that the published port contract contradicts D-2d, and it is right.
**The defect originates in §2.1**, the amendment written after M04's stop: it split the write
into `AppendChunk(..., ChunkPolicy)` and `Complete(..., CompletePolicy)` — two transactions —
while D-2d describes completion as the **terminal branch of the same transaction**, renaming
scratch into `source/` before the single commit that writes `queued`.

**The window this opens.** `AppendChunk` commits `received_bytes = declared_bytes` while the
row is still `uploading`. A crash before the separate `Complete` transaction leaves an asset
that no one can repair: `HEAD` reports the declared offset, so a conforming tus client
considers the upload finished and sends no further `PATCH`; there is no source and no job,
and the row is eventually swept as an expired upload. B31 covers the mirror window — source
renamed while the column is still short — so nothing covered this one.

**Amendment.** `AppendChunk` takes **both** policies. The repository invokes `ChunkPolicy` as
before and, only on the terminal branch where the new offset reaches `declared_bytes`,
invokes `CompletePolicy` (the rename into `source/`) before the single guarded UPDATE that
commits `queued`. The application enqueues after that method returns, so commit-then-enqueue
still holds and D-15's Redis-down exit is unchanged. `Complete` is removed unless a recovery
caller needs it; if kept, its only caller must be named.

**A new acceptance row** covers a crash after the final append and fsync but before the
database commit, alongside B31's mirror case.

Method note: this is the second contract defect in the same section, and both were caught by
the implementer rather than by six rounds of gates. The gates checked whether the described
mechanisms were correct; they did not check whether the port signatures could express them
(M04) or whether the signatures contradicted the mechanism they served (M11). A port table is
a contract, and a contract that disagrees with the decision it implements is a defect even
when every method in it is individually sound.

## §21 — Amendment, 2026-09-06 (gate owner, during M11)

M11 raised a second contract defect, again in §2.1, and again it is accepted.

**The conflict.** §2.1 states a blanket rule — every `ChunkPolicy` error aborts the
transaction and leaves the row exactly as locked. D-8 requires the opposite for exactly one
error: a failed first-chunk container sniff must **consume the ticket atomically**,
`created → failed` with reason `unsupported_source`, answering `400` with no enqueue (B18).

**The race the conflict opens.** Because the policy error rolls back and releases the row
lock, the failure is written by a *separate* transaction. An invalid first `PATCH` A can
release its lock, a valid `PATCH` B can advance the still-`created` asset, and A's guarded
`Fail` then loses. The invalid chunk does not consume the ticket and can answer with a
conflict or dependency error instead of the required `400`. B18's sequential case passes
throughout, so nothing existing catches it.

**Amendment.** `ErrUnsupportedSource` is `ChunkPolicy`'s one **terminal** outcome rather than
an ordinary abort: still holding the row lock, `AppendChunk` writes `state='failed'`,
`failure_reason='unsupported_source'`, keeps `received_bytes = 0` and commits, then returns
the typed error for the application to map. Every other policy error aborts unchanged, and
the application drops its separate `Fail` call. A new acceptance row holds the invalid first
`PATCH` inside the policy until a valid one is waiting, and proves the invalid one answers
`400`, the valid one cannot advance the row, the final state is `failed`, and neither source
nor job exists.

**Pattern, recorded so the next phase avoids it.** This is the third defect in §2.1, and all
three share one cause: **§2.1 wrote blanket rules where the decisions it serves are
specific.** "Every policy error aborts" reads as a clean invariant and silently contradicts
D-8; "AppendChunk then Complete" reads as a clean decomposition and silently contradicts
D-2d's single transaction; and the original method list carried no signatures at all. A
contract section must be checked *against every decision it implements*, not merely for
internal consistency — and none of the six gate rounds did that, because they were reading
the mechanisms, not the seam between mechanism and signature.
