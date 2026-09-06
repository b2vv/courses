# Media upload and transcoding — Phase 5 plan

Edition 1, 2026-09-06.

Sources, read in order and re-opened against the code: `AGENTS.md`; `.claude/standards.md`
(gate order, placement, reuse index, branches); `docs/media/spec.md` (GATE 1 decisions
Q1–Q8, **binding**); `work/PRODUCT-BRIEF.md` §2 rules 12/13/14, §10 (Media module and the
six video states), §11 (media architecture, the file layout, the profiles, the twelve-step
worker sequence, "no arbitrary user arguments to FFmpeg"), §16, §17, §24.11–24.13;
`docs/admin/plan.md` (4a — the authorization seam, the task-cut discipline, the
acceptance-table discipline, the D-numbered decision style) and `docs/admin/plan-4b.md`
(4b — the copy-not-import precedent, the isolation-level trap, the canary-renumber
pattern); `work/CTO-RESEARCH.md` §6 "Стан медіа-підсистеми перед Phase 5"; the code named
inline throughout. **Where the briefing and the code disagreed, the code won**; §0 lists
the four places that happened.

Mode: **ponytail-full** — the simplest thing that works, existing seams before anything
new, every deviation from the brief's stack or from a GATE 1 decision named together with
the trigger that would revisit it.

**Scope fence.** Phase 5 produces bytes and nobody may read them. No playback, no signed
URLs, no `auth_request`/`secure_link`, no hls.js, no player, no progress, no transcripts,
no captions, no orders, no S3/MinIO adapter, no DRM, no ASR, no Python. `/media/*` stays
**403 at the edge for everyone**, exactly as it is today
(`infrastructure/nginx/nginx.conf:44-45`), and this plan changes neither of those two
lines. Playback is Phase 6.

This is the largest slice so far: a new backend module, a **third executable**, ffmpeg,
a storage layout, and infrastructure changes to Nginx and Compose.

---

## 0. Verified facts, and where the briefing was corrected

| Claim | Verified at | Verdict |
|---|---|---|
| The worker's mux is empty; no producer exists anywhere | `backend/cmd/media-worker/main.go:50-54` — `server.Start(asynq.NewServeMux())` with the comment saying so | confirmed |
| ffmpeg is already in the worker image | `backend/Dockerfile:29-33` (`target: worker` installs ffmpeg and pre-creates `/media/{source,hls,thumbnails,temporary}` owned by 10001) | confirmed |
| `media-data` is writable in the worker and read-only in nginx | `docker-compose.yml:88` (`media-data:/media`), `:117` (`media-data:/media:ro`) | confirmed |
| `/media` and `/media/` answer 403 | `nginx.conf:44-45`; `infrastructure/tests/smoke.sh` asserts `/media/example/master.m3u8 → 403` | confirmed; **Phase 5 keeps it** |
| Next migration number is 00005 | `backend/migrations/` holds `00001_courses`, `00002_identity`, `00003_authoring`, `00004_identity_admin` | confirmed |
| The synthetic failing-migration canary must be renumbered | `backend/internal/course/adapters/postgres/repository_integration_test.go:190` — subtest **"failed migration rolls back its complete transaction"**, `fstest.MapFS` at line 195, canary `"00005_failure.sql"` at **line 200** | confirmed — it now **collides with the real 00005** |
| Lessons have no media column | `backend/migrations/00003_authoring.sql` `course_lessons(id, module_id, title, description, position, created_at, updated_at)` | confirmed |
| Identity/course use `pool.Begin`; course authoring uses `BeginTx(RepeatableRead)` | `authoring.go:76,124` are `RepeatableRead` **read-only**; `:147,186,274,317` (the writes) are `pool.Begin` | **briefing imprecise** — the RepeatableRead transactions in the course module are the *reads*, not the writes. Media follows the writes: `pool.Begin`, D-9 |
| The kit has `FileUpload` | `frontend/src/presentation/ui/controls.tsx:165-178`, exported at `index.ts:10` | **exists** — a `React Aria FileTrigger` wrapping a secondary `Button`. `ProgressBar` exists too (`data.tsx:7-23`). **Neither has a single consumer outside `ui.test.tsx`**, so both are unproven in anger; §5.g |
| `LockEligibleAuthoringOwner` is the ownership check to reuse | `backend/sql/queries/courses.sql:105-111` | **briefing wrong**. That query answers "is *this owner* eligible to own a course" (role ∈ {instructor,admin}, not disabled). The *actor may edit this course* check is `domain.Actor.CanManage(instructorID)` over the row from `LockAuthoringCourse` (`authoring.go:193-199`). Media reuses **the rule**, not that query — D-3 |

Three further facts the sources did not state, found in the code and load-bearing below:

- **`.gitignore:14` is `media/`, unanchored, so git already refuses every path this phase
  creates.** Verified with `git check-ignore -v`: `docs/media/plan.md` (this file),
  `docs/media/spec.md` (the approved GATE 1 spec — it is **not committed today**),
  `backend/internal/media/**` (the whole new module) and `frontend/src/domain/media/**` all
  match it. The pattern was written for the media *data* directory and catches every
  directory named `media` at any depth. Until it is anchored, a Phase 5 implementer can
  write the module, see a clean `git status`, and commit nothing. This is the first thing
  M01 fixes — §5.j.

- **`infrastructure/tests/smoke.sh` asserts the worker has exactly zero handlers**
  (`if(!Array.isArray(status.handlers)||status.handlers.length!==0) throw`). Registering the
  first handler turns `make smoke` red until that assertion is inverted — §5.d.
- **A 4a `PUT` deletes lessons the payload omits** (`authoring.go:237`,
  `deleteMissingAuthoringLessons`). Anything hanging off `course_lessons.id` is therefore
  deleted by an ordinary course save — D-4 exists because of this line.

---

## 1. Shape of the slice

One new backend module, one new executable, one new queue handler, three JSON operations,
one tus surface, one migration, one Nginx location, one Compose service, one editor
control.

**The byte path never touches Go's API or the SSR server** (§2 rule 13, §24.11). The
browser talks tus to a dedicated `upload` executable through Nginx; that executable writes
onto the media volume and enqueues; the worker transcodes and publishes; the API only ever
sees metadata.

```
browser ──tus /uploads/──▶ nginx ──▶ upload (:8082) ──write──▶ media-data ──read──▶ media-worker
   │                                       │                                          │
   └──JSON /api/v1/admin/──▶ nginx ──▶ api │                                          │
                                        └──── PostgreSQL: video_assets ───────────────┘
                                                        │
                          nginx ──/media/* → 403 (unchanged in Phase 5) ──▶ media-data:ro
```

### The JSON surface (in `openapi.yaml`, Orval-generated)

| Method | Path | Who | Answer |
|---|---|---|---|
| POST | `/api/v1/admin/lessons/{lessonId}/video` | admin any lesson; instructor only lessons in courses they own | `201` `{videoId, uploadUrl, chunkBytes, maxBytes, expiresAt}`; a `video_assets` row in state `created` |
| GET | `/api/v1/admin/courses/{courseId}/videos` | same rule, scoped by the course | `200` `{items: [{lessonId, videoId, state, generation, updatedAt, failureReason?}]}` — one row per non-evicted asset |
| POST | `/api/v1/admin/videos/{videoId}/retry` | same rule | `200` with the asset back in `queued`; `409 video_not_retryable` unless it is `failed` |

Three operations across three path templates. **No `DELETE` on the API**: cancelling an
upload is the tus `DELETE` on the ticket (D-2), and superseding a video is just another
mint (D-5).

### The tus surface (not in `openapi.yaml` — D-2)

`POST /uploads/`, `HEAD /uploads/{videoId}`, `PATCH /uploads/{videoId}`,
`DELETE /uploads/{videoId}`, `OPTIONS /uploads/` — tus 1.0.0 **core** plus the `creation`
and `expiration` extensions, nothing else.

### The video state machine

Exactly the six states of `work/PRODUCT-BRIEF.md` §10. No seventh state is invented; the
two lifecycle facts that are not states are timestamps, the 4b idiom (D-6):

```
created ──(first PATCH)──▶ uploading ──(completion)──▶ queued ──(claim)──▶ processing
                                                          ▲                    │
                                                          └──(retry)───────────┤
                                                                   ┌───────────┴──────────┐
                                                                 ready                  failed
```

| Transition | Written by | Guard |
|---|---|---|
| → `created` | **API**, `POST …/video` | the 4a ownership rule (D-3); no other non-terminal asset for the lesson (D-5) |
| `created` → `uploading` | **upload service**, first `PATCH` | valid, unexpired, unconsumed ticket |
| `uploading` → `queued` | **upload service**, on completion | received bytes == declared length; container sniff passes (D-8); source renamed into place |
| `queued` → `processing` | **worker**, on claim | single guarded `UPDATE … WHERE state = 'queued'` |
| `processing` → `ready` | **worker**, after atomic publish | publish succeeded; eviction applied in the same transaction (D-5) |
| `processing`/`uploading` → `failed` | **worker** (transcode) or **upload service** (sniff) | terminal only after the last permitted attempt, or immediately for a malformed source |
| `failed` → `queued` | **API**, `POST …/retry` | state is exactly `failed` |

---

## 2. Placement (Розміщення)

Per `.claude/standards.md` §Розміщення. Every new file is named; extended files are marked
`(extend)`. Sibling `*_test.go` / `*.test.ts(x)` accompany every new module and are not
listed individually.

### Backend — the new `internal/media` module

| File | Contents |
|---|---|
| `backend/migrations/00005_media.sql` | `video_assets`, its CHECKs, the two partial indexes, `ON DELETE SET NULL` to `course_lessons`; down migration (§6) |
| `backend/internal/media/domain/asset.go` | `VideoAsset`, `State`, `Generation`, `StorageKey` and its four kinds, the transition matrix, ticket bounds (size, TTL, chunk), `NewAsset` invariants |
| `backend/internal/media/domain/actor.go` | `Actor{UserID string; Administrator bool}` and `CanEditLesson(ownerID string) error` — **capability, not role vocabulary** (D-3) |
| `backend/internal/media/domain/source.go` | `SourceProbe`, the accepted container/codec set (Q2), `AcceptSource`, `SelectLadder` (360/720/1080-if-supported, never upscaling) |
| `backend/internal/media/domain/errors.go` | the six business errors (§8) |
| `backend/internal/media/domain/*_test.go` | table tests for the transitions, the ladder, the source acceptance, the ticket bounds |
| `backend/internal/media/ports/repository.go` | `AssetRepository` — mint, claim, publish-with-eviction, fail, retry, list-by-course, sweep candidates |
| `backend/internal/media/ports/ownership.go` | `LessonOwnership` — `LockEditableLesson(ctx, lessonID) (courseID, ownerID string, err error)` |
| `backend/internal/media/ports/storage.go` | `MediaStorage` — opaque `StorageKey` in, no filesystem path ever crosses the boundary (ADR 004) |
| `backend/internal/media/ports/transcoder.go` | `Transcoder` — `Probe(ctx, StorageKey) (SourceProbe, error)`, `Render(ctx, RenderRequest) error`, `Thumbnail(ctx, …) error` |
| `backend/internal/media/ports/queue.go` | `TranscodingQueue` — `Enqueue(ctx, assetID string, attempt int) error` |
| `backend/internal/media/application/upload.go` | `MintTicket`, `BeginUpload`, `CompleteUpload`, `CancelUpload`, `ListCourseVideos`, `Retry` |
| `backend/internal/media/application/transcode.go` | `Process` — the twelve-step sequence of brief §11, expressed over the four ports; `Sweep` |
| `backend/internal/media/application/*_test.go` | fake-port tests: every transition, retry exhaustion, malformed source, no-op on `ready`, duplicate enqueue, cancellation, dependency failure |
| `backend/internal/media/adapters/postgres/repository.go` | the guarded state UPDATEs, the publish-and-evict transaction, `pool.Begin` (D-9) |
| `backend/internal/media/adapters/postgres/ownership.go` | the lesson→module→course→owner join with `FOR SHARE` on the course row (D-3) |
| `backend/internal/media/adapters/postgres/*_integration_test.go` | Testcontainers: two generations, third-evicts-oldest, orphan reaping, expired-ticket takeover, ownership refusal and its serialisation against a 4a owner reassignment |
| `backend/internal/media/adapters/filesystem/storage.go` | key→path resolution inside an owned root, `O_NOFOLLOW`/`Root`-scoped opens, traversal refusal, append-at-offset, atomic `rename` publish, recursive purge |
| `backend/internal/media/adapters/filesystem/storage_integration_test.go` | tempdir, no database: traversal, symlink, partial-write, publish atomicity |
| `backend/internal/media/adapters/ffmpeg/transcoder.go` | argv construction from the fixed ladder, `ffprobe` JSON parsing, `exec.CommandContext` with `Setpgid` + process-group kill |
| `backend/internal/media/adapters/ffmpeg/transcoder_test.go` | exact-argv assertions; process-group cancellation proven against `/bin/sleep` (no ffmpeg needed) |
| `backend/internal/media/adapters/asynq/queue.go` | the producer: `asynq.TaskID`, `MaxRetry`, `Timeout`, `Retention`; `ErrTaskIDConflict` treated as success |
| `backend/internal/media/adapters/asynq/handler.go` | the consumer: payload decode, `GetRetryCount`/`GetMaxRetry`, `SkipRetry` for malformed sources, duration and failure metrics |
| `backend/internal/media/adapters/http/handler.go` | the three JSON operations, DTOs, bounds, error mapping, `RegisterAdmin(mux, authenticate, requireSameOrigin)` — the 4a shape verbatim |
| `backend/internal/media/adapters/http/handler_test.go` | contract tests against `openapi.yaml`; the refusal matrix |
| `backend/internal/media/adapters/tus/server.go` | tus 1.0.0 core + creation + expiration over `net/http`; no third-party tus dependency (D-2) |
| `backend/internal/media/adapters/tus/server_test.go` | offset semantics, resume, wrong-offset conflict, unknown/expired ticket, over-length chunk |
| `backend/sql/queries/media.sql` | the media queries; regenerates `internal/media/adapters/postgres/generated` |

### Backend — the third executable and shared plumbing

| File | Contents |
|---|---|
| `backend/cmd/upload/main.go` | the third executable: `runtime.Open` → media repository + filesystem storage + asynq queue → `tus.Server` mounted under `/uploads/` on `httpserver.NewHandler`, `:8082` |
| `backend/cmd/media-worker/main.go` (extend) | replaces the empty `asynq.NewServeMux()` with the media handler; adds `ErrorHandler`; runs `Sweep` at start; `/status` reports the registered handler |
| `backend/cmd/api/main.go` (extend) | constructs the media application and calls `mediaHandler.RegisterAdmin(mux, identityHandler.Authenticate, identityHandler.RequireSameOrigin)` |
| `backend/platform/observability/metrics.go` (extend) | `courses_transcoding_job_duration_seconds{outcome}` (exponential buckets 1 s … ~34 min) and `courses_transcoding_failures_total{reason}` (§17) |
| `backend/openapi/openapi.yaml` (extend) | 3 path templates / 3 operations, the media schemas, six new `error.code` values |
| `backend/internal/course/adapters/postgres/repository_integration_test.go` (extend) | add the real `00005_media.sql` to the `fstest.MapFS` (line 195) and renumber the canary `00005_failure.sql` → `00006_failure.sql` (line 200) — §5.b |
| `backend/Dockerfile` (extend) | build `/out/upload`; new `FROM runtime AS upload` stage (**no ffmpeg**) |

### Frontend

| File | Contents |
|---|---|
| `frontend/src/domain/media/video.ts` | `VideoState`, `LessonVideo`, `VideoError`, the ticket type, the client-side accept list mirroring Q2 |
| `frontend/src/ports/LessonVideoRepository.ts` | mint, list-by-course, retry, and the resumable byte transfer as one port method with a progress callback |
| `frontend/src/application/admin/lesson-video.ts` | use cases over the port, returning `Result` |
| `frontend/src/adapters/api/lesson-video-schema.ts` | zod schemas |
| `frontend/src/adapters/api/lesson-video-repository.ts` | Orval calls through `createApiTransport`; business-code mapping (the `admin-course-repository.ts` idiom) |
| `frontend/src/adapters/api/tus-client.ts` | ~80 lines of native `fetch`: create, `HEAD` for the offset, sequential `PATCH`, resume on failure, per-chunk deadline, progress and cancellation (D-2, §5.g) |
| `frontend/src/app/composition/admin-video-queries.ts` | query options, the poll interval, the upload mutation, `invalidateAfterVideoWrite` |
| `frontend/src/app/composition/admin.ts` (extend) | `createAdminDependencies` returns `{...courses, users, videos}` — the 4b pattern, no new router-context key |
| `frontend/src/presentation/admin/LessonVideoField.tsx` | per-lesson: `FileUpload`, `ProgressBar`, state badge, failure reason, retry, refusal and degraded states |
| `frontend/src/presentation/admin/CourseEditorView.tsx` (extend) | mounts `LessonVideoField` inside the existing `LessonFields` fieldset (`CourseEditorView.tsx:231`) |
| `frontend/src/app/providers/locales/uk.ts`, `en.ts` (extend) | the media vocabulary; keys added only |
| `frontend/e2e/media.setup.ts` | generates the test clip inside the worker container and copies it to `playwright/.fixtures/` (§5.f) |
| `frontend/playwright.config.ts` (extend) | the `setup` project's `testMatch` becomes `['**/auth.setup.ts', '**/media.setup.ts']` |
| `frontend/e2e/media.spec.ts` | the browser journey against Compose |

### Infrastructure and docs

`docker-compose.yml` (the `upload` service, nginx `depends_on`), `infrastructure/nginx/nginx.conf`
(one new `location`, one new `limit_req_zone`), `infrastructure/tests/smoke.sh` (§5.d),
`README.md`, `docs/adr/008-tus-ingest.md` (new), `docs/adr/009-fmp4-segments.md` (new,
supersedes the segment choice of ADR 005), `docs/media/report.md` (new), `work/tasks.md`
(orchestrator-owned).

**Unchanged on purpose:** `.claude/standards.md` §Машинні гейти (no new gate step),
`Makefile` (no new target), `.github/workflows/ci.yml`, Prometheus, `.env.example`
(no new secret — D-2 is why), `backend/sql/seed.sql`, `backend/sql/seed-identities.sql`
(media has no fixtures; the e2e clip is generated, §5.f), and the two `/media` lines in
`nginx.conf`.

---

## 3. Reuse-first

| Concern | Reused | Why nothing new |
|---|---|---|
| Server authorization (same-origin → authenticate → role → ownership) | `identityHandler.Authenticate`, `identityHandler.RequireSameOrigin`, `httpserver.Actor`, wired as plain `func(http.Handler) http.Handler` in `cmd/api/main.go:54` | the 4a seam (`docs/admin/plan.md` D-1); media is its third consumer. The media handler registers exactly as `RegisterAdmin` does (`admin_handler.go:100-110`) |
| The ownership rule itself | `domain.Actor.CanManage`'s rule — admin any, instructor own — restated as one predicate over a capability (D-3) | the rule is the product's, not the course module's; the *code* is not importable without a cross-module domain dependency 4a explicitly refused |
| The guarded single-statement transition | 4a D-3f (`UPDATE … WHERE id = $1 AND status = $from`), `authoring.go:55-73` | read-then-act is a race under READ COMMITTED; media's state flips are the same shape |
| The microsecond clock discipline | `nextAuthoringTimestamp` (`authoring.go:436`), the I03 lesson | `timestamptz` does not carry nanoseconds; a token that changes on read is worse than none |
| Asynq server, queue option, logger, health, inspector, queue-depth gauge, graceful shutdown | `backend/cmd/media-worker/main.go:38-68`, `backend/platform/redis/{client,queue}.go` | all of it exists and is proven; Phase 5 fills the mux it deliberately left empty |
| Executable bootstrap (Postgres pool, Redis, metrics, tracing, readiness, `HTTP_ADDR`) | `platform/runtime.Open` + `httpserver.NewHandler`/`Run` | the third executable is three lines of wiring, not a new pattern |
| HTTP envelope, request id, JSON and error writers | `platform/httpserver` | `WriteJSON`/`WriteError` already emit the contract's error envelope |
| Contract-test harness (kin-openapi) | `internal/*/adapters/http/handler_test.go` | new paths plug into the existing response validator |
| Repository harness (Testcontainers + goose provider) | `course/adapters/postgres/repository_integration_test.go`, `identity/…/schema_integration_test.go` | already boots isolated PostgreSQL and applies migrations |
| The canary-renumber pattern | 4a `fd90fc9` and 4b §5.b | the third time; the procedure is known |
| UI kit: `FileUpload`, `ProgressBar`, `Button`, required-label `Table`/`Skeleton`, `ErrorState`, `EmptyState` | `frontend/src/presentation/ui/` | `FileUpload` and `ProgressBar` were built in foundation for exactly this consumer and have waited since; §5.g |
| Frontend admin idioms | `admin-course-repository.ts`, `admin-queries.ts`, `admin.ts` composition, `createApiTransport`, MSW, RHF | the video field is a child of the existing editor form, not a new slice |
| e2e shared `storageState` | 4a `auth.setup.ts` (D-7) | the media spec adds **zero logins**, so the `login` zone budget (≤7 per run) is untouched |

**What is genuinely new, and why the existing thing did not fit:**

- **`internal/media` as a module.** Brief §10 names `VideoAsset`, `Upload`,
  `TranscodingJob`, `MediaStorage` and `Transcoder`; `internal/media` does not exist
  (`work/CTO-RESEARCH.md` §6). Putting video state on the course aggregate would drag a
  filesystem and a subprocess into the catalog module.
- **The `upload` executable.** §2 rule 13 and §24.11 forbid video bytes through the Go API
  and the SSR server; §2 rule 12 gives the shape (a separate executable over shared
  modules and ports). Nothing existing terminates a byte stream.
- **Four ports.** Each has a real external dependency — PostgreSQL, the filesystem, the
  ffmpeg/ffprobe subprocesses, Redis/Asynq — so none is a speculative seam (AGENTS.md).
- **The tus client and its transport.** `createApiTransport` has a 5 s deadline and a
  JSON-shaped `execute` contract (`transport.ts:26,46-51`); a chunked binary `PATCH`
  needs a per-chunk deadline, an offset protocol and progress. It reuses the transport's
  *failure vocabulary* (`TransportFailureCode`) so the domain error union is unchanged.
- **Two metrics.** §17 requires transcoding job duration and transcoding failures;
  `platform/observability/metrics.go:15-18` has neither.
- **Two ADRs.** ADR 005 chose `.ts` segments; GATE 1 Q3 chose fMP4/CMAF `.m4s`.
  `.claude/standards.md` puts `docs/adr/*.md` in the Мірник and says a change that
  contradicts one **needs a new ADR** — so ADR 009 supersedes that clause of 005, and
  ADR 008 records the tus protocol and the placement Q1 demands.

---

## 4. Decisions

### D-1 The module boundary: media reads course rows, imports no course code

The media module needs one fact the course module owns: *who owns the course this lesson
belongs to*. Three options were weighed.

1. Import `courses/internal/course/domain` from media — permitted by
   `boundaries_test.go` (it only requires `/domain` or `/ports` in the path), but it makes
   two business domains permanently coupled, which 4a D-1 refused in the other direction.
2. An in-process port implemented by the course module and injected at the composition
   root — a second `RegisterAdmin`-style wiring, and a course-module change for a media
   feature.
3. **Chosen:** media declares `ports.LessonOwnership` and implements it in its **own**
   Postgres adapter with its own three-table join. Postgres is a real external dependency,
   so the port is legitimate; the SQL is *copied, not imported* — the precedent 4b set for
   the escaped-`LIKE` replacer (`plan-4b.md` NB2-3).

The copied surface is one `SELECT` of two columns. Trigger to revisit: a **third** module
needing the same join, or the join outgrowing "which course, which owner".

### D-2 The tus service: its shape, and how it authorizes without duplicating identity

GATE 1 Q1 is binding: resumable tus, terminated in a separate executable (or tusd), the
Go API never receiving video bytes. Two sub-decisions.

**D-2a — our own tus core, not tusd.** tusd is a maintained implementation and would be
zero Go code, but it authorizes through HTTP hooks: a `pre-create` hook and a `post-finish`
hook, both of which we would have to implement on our API anyway, plus a new image, a hook
payload contract, and tusd's own `.info` sidecar layout that our storage adapter would then
have to understand. Implementing **tus 1.0.0 core plus `creation` and `expiration`** —
`POST` create, `HEAD` → `Upload-Offset`, `PATCH` `application/offset+octet-stream`,
`DELETE` terminate, `Tus-Resumable: 1.0.0` on every exchange — is a bounded, well-specified
piece of HTTP with an obvious test surface, and it lets the executable be *ours*: same
`runtime.Open`, same `httpserver` envelope, same readiness contract, same logging rules.
**Not adopted: the `concatenation`, `checksum` and `creation-defer-length` extensions.**
Trigger to swap to tusd: parallel chunk upload, checksum verification, or an S3 backing
store — any of the three makes hand-rolling a bad trade.

**D-2b — authorization is a single-use ticket row, not a signed token and not a callback.**
The brief posed the choice as "a signed upload ticket minted by the API, or a call back to
the API". Both were rejected in favour of a third:

- A **signed** ticket (HMAC) needs a shared secret in `.env` and Compose, a new crypto
  surface, and its own expiry/replay semantics — a secret introduced for a problem the
  database already solves.
- A **callback** needs an internal HTTP contract (which is not in `openapi.yaml`, so it is
  ungoverned by the drift gate), a hop per upload, and it puts the caller's *cookie* into
  the upload service's hands — precisely the identity logic Q1 says it must not have.
- **Chosen: the `video_assets` row is the capability.** `POST /api/v1/admin/lessons/{id}/video`
  is an ordinary cookie-authenticated admin operation: it runs the 4a chain
  (same-origin → authenticate → role → ownership, D-3), and on success mints a row with a
  random UUIDv4 id, the declared length, the declared content type, and `expires_at`.
  The response's `uploadUrl` is `/uploads/{videoId}`. **The upload service never reads a
  cookie, never learns a role, never knows what a session is.** It validates one thing:
  that the addressed row exists, is `created` or `uploading`, and is unexpired. Revocation
  is a row update; expiry is a column; single use is the state machine.

The ticket is a bearer capability, and it is scoped accordingly: it grants *append bytes to
this one upload slot* and nothing else — no read, no listing, no other lesson. Its id is an
unguessable v4 UUID, it expires in **24 hours** (a resumable upload must survive a closed
laptop), and it is consumed by the completion transition. Accepted and recorded so a later
reviewer finds the reasoning rather than the symptom.

**D-2c — the tus surface is deliberately absent from `openapi.yaml`.** tus is a
header-driven binary protocol; describing it in OpenAPI would make Orval generate a client
we do not use and would put binary-body operations under the drift gate for no benefit.
The protocol is specified in ADR 008 and exercised by `adapters/tus/server_test.go` and
`e2e/media.spec.ts`. The drift gate keeps covering exactly what Orval generates.

### D-3 Authorization: a capability, not a third copy of the role vocabulary

4a D-1 point 4 duplicated three role constants across two domains and recorded the trigger:
*"when a third module needs roles, promote them to a shared kernel rather than duplicating
twice more."* Phase 5 is the third module, so the trigger is live and must be answered.

**Decision: the trigger does not fire, because media never names a role.** The media domain
models the *capability*, matching §24.10's additive-capability model taken on 2026-09-06:

```go
type Actor struct { UserID string; Administrator bool }
func (a Actor) CanEditLesson(ownerID string) error   // admin ⇒ ok; else UserID == ownerID
```

The single string comparison `httpserver.Actor.Role == "admin"` happens once, in the media
**HTTP adapter**, which is where transport is translated. No new `student`/`instructor`/
`admin` constant set is created, so the duplication the trigger guards against does not
occur, and no refactor is imposed on 4a/4b's proven code to serve a new module.

**Order of checks on every media mutation**, identical to 4a NB-8 and 4b D-1: same-origin
first (an anonymous cross-site `POST` answers `403 origin_forbidden`, not `401`), then
authenticated, then role (`instructor` or `admin`, else `403 forbidden`), then the lesson
exists (`404 lesson_not_found`), then ownership (`403 course_forbidden`), then the domain
rule. Every mutation acceptance row states a valid `Origin` in its precondition.

The ownership read takes `FOR SHARE` on the `courses` row inside the minting transaction —
the same lock shape as `LockEligibleAuthoringOwner` (`courses.sql:105-111`) — so a mint
cannot interleave with a 4a owner reassignment and authorize against a stale owner.

### D-4 Where the lesson↔video link lives, and why the FK is `SET NULL`

Q6/§24.13 puts **two generations per lesson**, so a column on `course_lessons` cannot hold
the link: one column holds one id. The link is therefore a **table**,
`video_assets(id, lesson_id, generation, state, …)`, which also keeps every media column out
of the course schema and out of 4a's aggregate.

The referential action is the sharp part. `authoring.go:237` deletes every lesson a 4a
`PUT` omits, and the editor's "delete lesson" button does exactly that
(`CourseEditorView.tsx:275`). So an ordinary course save can remove a lesson that owns
video.

| Option | Consequence |
|---|---|
| `ON DELETE RESTRICT` | a 4a `PUT` removing a lesson with video fails with `23503` — a Phase-5 feature breaking a Phase-4 operation. Rejected |
| `ON DELETE CASCADE` | the row vanishes and **the bytes stay on disk forever**, unreferenced and unfindable. Rejected |
| **`ON DELETE SET NULL`** | the asset survives its lesson as an explicit orphan; the sweeper (D-7) finds `lesson_id IS NULL`, deletes the bytes and then the row. **Chosen** |

The cost is a nullable `lesson_id`, which weakens "every asset belongs to a lesson".
Recorded meaning: **`lesson_id IS NULL` means the lesson was deleted and this asset is
garbage awaiting the sweeper**; every read filters it out. No trigger, no cascade, no leak.

### D-5 Two generations, one in-flight, and the eviction point

- **At most one non-terminal asset per lesson**, enforced structurally:
  `CREATE UNIQUE INDEX video_assets_pending_per_lesson ON video_assets (lesson_id)
   WHERE lesson_id IS NOT NULL AND evicted_at IS NULL AND state IN ('created','uploading','queued','processing')`.
  A second mint while one is in flight is `409 upload_in_progress`. This bounds temporary
  disk without a quota subsystem and removes the two-uploads-race entirely.
  **Exception:** if the in-flight asset's ticket has **expired**, the mint marks it
  `evicted_at` in the same transaction and proceeds — otherwise an abandoned upload would
  lock a lesson for 24 hours.
- **At most two `ready` generations per lesson.** This cannot be an index (an index cannot
  count to two), so it is a transactional rule at the **publish** point, which is the only
  place a third generation can appear: the worker's publish transaction locks the lesson's
  non-evicted assets `FOR UPDATE`, flips the new asset to `ready`, and if three now exist,
  stamps `evicted_at` on the **oldest by generation** — all in one transaction. The current
  video is the highest `generation` that is `ready`; the other is the one-step rollback.
- **Bytes are deleted after the commit, never inside it**, because a `rename`/`RemoveAll`
  cannot be rolled back. The window between commit and unlink is covered by D-7: the
  sweeper reaps anything with `evicted_at` set whose bytes still exist. A crash in the
  window costs disk, never correctness, and never an asset that reads `ready` with no bytes.

### D-6 Two lifecycle facts are timestamps, not states

Brief §10 fixes the video vocabulary at six states, and this plan does not extend it.
Eviction and cancellation are recorded as `evicted_at timestamptz` — the 4b
`deleted_at`/`disabled_at` idiom, where the projection lives in the domain and the column is
the truth. An evicted asset keeps whatever state it reached (`ready`, `failed`, `uploading`)
and is invisible to every read. A tus `DELETE` on an in-flight ticket sets `evicted_at`; so
does the eviction of a third generation; so does the mint that takes over an expired ticket.
One column, three producers, one sweeper.

### D-7 The sweeper

One function, two call sites, no scheduler:

1. at worker start (so a crash-recovered process cleans up before it works), and
2. as **step 0 of every transcode job**.

It deletes, for each candidate row: the `source/{id}`, `hls/{id}`, `thumbnails/{id}.jpg` and
any `temporary/{id}-*` keys, then the row. Candidates are `evicted_at IS NOT NULL`, or
`lesson_id IS NULL` (D-4), or a non-terminal asset whose ticket expired more than 24 h ago.
Bounded to 50 rows per invocation so a job never turns into a cleanup marathon.

Not chosen: a cron/Asynq periodic task (`asynq.Scheduler`) — a second scheduling concept for
a job that already runs whenever there is work, and that has nothing to clean when there is
none. Trigger to add one: media that must be reclaimed on a deadline with no uploads
happening.

### D-8 The storage layout and the atomic publish

Exactly brief §11, with the two unshaped directories given a shape:

```
/media/
├── source/{videoId}/original
├── hls/{videoId}/{master.m3u8, 360p/, 720p/, 1080p/}   ← 1080p only if the source supports it
├── thumbnails/{videoId}.jpg
└── temporary/{videoId}-{attempt}/                       ← one working directory per attempt
```

`MediaStorage` takes an opaque `StorageKey{Kind, ID, Leaf}` — never a path (ADR 004). The
filesystem adapter resolves a key inside **one owned root** and refuses anything else:
`Kind` is one of four constants, `ID` must parse as a UUID, `Leaf` is drawn from a fixed set.
A client-supplied filename **never reaches the filesystem**: it is stored in a column for
display only. Opens use Go 1.24's `os.Root` so a symlink planted in the volume cannot escape
the root; `..` cannot appear because no key component is free text.

**Publication is atomic because the destination is new**, not because of a swap: `{videoId}`
is minted per upload, so `hls/{videoId}` never exists before its one publish, and a
same-filesystem `rename(2)` of `temporary/{videoId}-{attempt}` onto it is atomic. Two
branches are stated rather than assumed:

- destination exists **and** the asset is already `ready` → the job is a no-op (D-10);
- destination exists **and** the asset is not `ready` → a previous attempt published and then
  died before its commit; the adapter removes the stale directory and re-publishes.

The thumbnail is renamed into `thumbnails/` after the playlist directory lands, so a reader
that sees the playlist can be told the thumbnail is optional; nothing in Phase 5 reads either.

**Sniff-before-queue (Q2).** The upload service checks the container at completion, before
the source is renamed into place and before anything is enqueued: the first bytes must carry
an `ftyp` brand accepted for mp4/mov, or the Matroska EBML magic. A failure is
`unsupported_media`, the asset goes straight to `failed`, and **no job is ever queued** —
which is what Q2 asks for. **Named deviation:** the *codec* half of Q2 (H.264/HEVC video,
AAC/MP3 audio) cannot be checked there, because ffprobe lives in the worker image and putting
ffmpeg into the byte-receiving service would double its image and give it subprocess
execution for no gain. The codec check is therefore the worker's first step and produces
`failed` with `SkipRetry`, never a retry. Trigger to revisit: a real user pattern of
uploading correctly-containered but wrongly-coded video often enough that the round trip
annoys — at which point a tiny probe-only sidecar, not ffmpeg in the upload image.

### D-9 Transactions and isolation

**Every media write uses `pool.Begin` (READ COMMITTED)**, matching the identity module's
`Rotate` and the course module's *writes* (`authoring.go:147,186,274,317`). It explicitly
does **not** copy `authoring.go:76,124`'s `BeginTx(RepeatableRead, ReadOnly)`: those are the
consistent *reads*, and 4b already recorded what copying them into a write path costs — a
serialization failure (`40001`) that `persistenceError` maps to `503`, turning a business
outcome into a dependency outage (`plan-4b.md` D-2). Every state flip is a single guarded
`UPDATE … WHERE id = $1 AND state = $from AND evicted_at IS NULL`; zero rows affected is
disambiguated by a re-read **inside the same transaction**, the 4a D-3f shape. Every write
stamps `updated_at` with the application clock **truncated to microseconds** (the I03
lesson), and the schema's `updated_at >= created_at` CHECK holds because the same clock value
is written to both on insert.

### D-10 The job payload, idempotency, and retry

**Payload:** `{"assetId": "<uuid>", "attempt": <int>}` — nothing else. No paths, no
filenames, no actor. Enqueued as
`asynq.TaskID(assetId + ":" + attempt)`, `asynq.Queue("media")`, `asynq.MaxRetry(3)`,
`asynq.Timeout(30 * time.Minute)`, `asynq.Retention(24 * time.Hour)`.

- **A duplicated enqueue is refused by Asynq**, not by us: `asynq.ErrTaskIDConflict`
  (verified in `asynq@v0.26.0/client.go:228,416`) is mapped to *success* by the queue
  adapter — the task the caller wanted already exists.
- **A retried job cannot corrupt a published asset.** The handler's first act after the
  sweep is to load the asset; `ready` or `evicted_at IS NOT NULL` → return `nil`
  immediately, publishing nothing and touching no bytes. Only `queued` (or a `processing`
  row whose lease has expired) is claimed, by a guarded `UPDATE`.
- **Working directories are per attempt** (`temporary/{videoId}-{attempt}`), so two attempts
  can never share a byte of scratch space even if a lease is misjudged.
- **Bounded attempts (Q8).** `MaxRetry(3)`. The handler reads `asynq.GetRetryCount` and
  `asynq.GetMaxRetry` (verified in `asynq@v0.26.0/context.go:25,33`); on the final permitted
  attempt it writes `failed` with an operator-readable reason **before** returning the error,
  so the terminal state comes from a real outcome and never from a guess (spec criterion 2).
  A malformed source returns `asynq.SkipRetry` (verified at `processor.go:329,343`) after
  writing `failed`, so it is archived and never retried.
- **Cancellation (§11 step 12).** Worker shutdown cancels the handler's context; the
  transcoder's `exec.CommandContext` runs ffmpeg in its own process group
  (`SysProcAttr{Setpgid:true}`) and `cmd.Cancel` signals the **group**, with `WaitDelay` for
  the hard kill — otherwise ffmpeg's children outlive the worker and keep writing into a
  directory nobody owns. The handler then removes its working directory and leaves the asset
  in `processing` with an expired lease, which the next claim takes. Nothing partial is ever
  published, because publish is the last step.
- **Manual retry (Q8).** `POST /admin/videos/{id}/retry` accepts only `failed`, increments
  `attempt`, sets `queued`, and enqueues under the new task id. Any other state is
  `409 video_not_retryable`.

### D-11 The exact ffmpeg invocation shape — no user-controlled arguments

Brief §11: *"Не передавай довільні аргументи користувача в FFmpeg command."* The rule is made
structural: **the argv is built from three sources and no others.**

1. **Compile-time constants** — every flag and every profile value, from a fixed ladder
   table in `domain/source.go` (360p: 640×360 @ 800k, 720p: 1280×720 @ 2500k, 1080p:
   1920×1080 @ 5000k; H.264 `main`/`high`, `-preset veryfast`, GOP 48, AAC 96–128k stereo).
2. **Paths derived from the asset UUID** by the storage adapter — never from any client
   string. The original filename is display metadata in a column and reaches neither argv
   nor the filesystem.
3. **Integers from ffprobe**, used *only to select a subset of the ladder* by comparison —
   never interpolated into a string. A source shorter than a rendition's height simply omits
   it, so nothing is ever upscaled.

Probe: `ffprobe -v error -show_streams -show_format -print_format json -- <source>`, parsed
into a strict struct with `encoding/json`; unknown fields are ignored, missing ones are a
refusal, not a zero.

Render, one invocation producing every rendition and the master playlist in one pass:

```
ffmpeg -nostdin -hide_banner -loglevel error -y -i <source>
  -filter_complex "[0:v]split=N[a][b]…;[a]scale=w=640:h=360:force_original_aspect_ratio=decrease[v0];…"
  -map "[v0]" -c:v:0 libx264 -b:v:0 800k  -preset veryfast -g 48 -keyint_min 48 -sc_threshold 0
  -map a:0  -c:a:0 aac    -b:a:0 96k -ac 2
  …one -map/-c:v/-c:a group per selected rendition…
  -f hls -hls_time 4 -hls_playlist_type vod
  -hls_segment_type fmp4 -hls_fmp4_init_filename init.mp4
  -hls_segment_filename <work>/%v/segment_%05d.m4s
  -master_pl_name master.m3u8
  -var_stream_map "v:0,a:0,name:360p v:1,a:1,name:720p …"
  <work>/%v/index.m3u8
```

Thumbnail, a second short invocation: `-ss <1s or 0 if shorter> -i <source> -frames:v 1
-vf scale=640:-2 -f image2 <work>/thumbnail.jpg`.

`-hls_segment_type fmp4` is Q3/§24.12 (`.m4s`, CMAF-shaped), which is why ADR 009 is
mandatory. `libx264` and `aac` are ffmpeg's built-ins in the Debian package already in the
worker image, so no image change is needed for the codecs.

The argv builder is a pure function returning `[]string`, and its exact output is asserted in
a unit test — the only honest way to prove "no user argument reaches ffmpeg" without running
ffmpeg.

### D-12 Status delivery: polling a separate read, not the course aggregate

Q5 is polling, no websockets and no SSE. The read is **not** folded into the 4a course tree:
that response is the mirror of the `PUT` body, whose schema is `additionalProperties: false`
and deliberately excludes lifecycle fields (4a D-3a), and folding media in would couple the
media module to the course aggregate and invalidate the editor's tree on every state tick.

`GET /api/v1/admin/courses/{courseId}/videos` is its own query key with
`videoPollMs = 3_000`. `refetchInterval` returns the interval while any row is non-terminal
and `false` otherwise, so polling **stops by construction** at `ready`/`failed`; Query's
default `refetchIntervalInBackground: false` stops it in a hidden tab. One tick is one
indexed query over one course.

### D-13 The upload control mounts only on a persisted lesson

4a mints module and lesson ids **on the client** (`newLessonId()`, `CourseEditorView.tsx:288`)
and they exist in the database only after the tree `PUT` lands. A mint against an
unsaved lesson would answer `404 lesson_not_found` — a confusing refusal caused by the UI,
not by the user.

**Decision:** `LessonVideoField` renders the upload control only for a lesson id present in
the last **loaded** tree; for a newly added lesson it renders a short hint ("save the course
to add video"). No new server concept, no optimistic lesson creation, and no request that is
predestined to fail. Trigger to revisit: if users find the two-step flow annoying, the mint
endpoint could accept a lesson id the same request creates — which is a course-module change,
not a media one.

### D-14 Observability (§17)

- `courses_transcoding_job_duration_seconds{outcome}` — histogram, outcomes
  `ready|failed|cancelled`, exponential buckets from 1 s (a 2-second clip) to ~34 min (the
  task timeout). `prometheus.DefBuckets` tops out at 10 s and would put every real transcode
  in `+Inf`.
- `courses_transcoding_failures_total{reason}` — counter, reasons `unsupported_source`,
  `probe_failed`, `render_failed`, `publish_failed`, `storage_unavailable`. Bounded
  cardinality: the labels are a closed set of constants, never an error string.
- `courses_queue_depth{queue,state}` already exists and already watches `media`
  (`platform/redis/queue.go:31-56`) — nothing to add.
- HTTP duration and status for the upload service come free from `httpserver.NewHandler`.
- **Logging.** No filesystem path, no ticket id, no original filename, no query string is
  logged; a media log line carries the asset id, the state and a reason constant. The
  existing `QueueLogger` already suppresses Asynq's free-form arguments
  (`platform/redis/queue.go:13-27`) and stays as it is.

---

## 5. Side effects (Сайд-ефекти)

### a. A third executable changes the build, the image and the topology

`backend/Dockerfile` builds `/out/upload` in the existing `build` stage and adds
`FROM runtime AS upload` — **without ffmpeg**, so the byte-receiving service cannot execute a
codec. `docker-compose.yml` gains an `upload` service: `media-data:/media` writable,
`HTTP_ADDR=:8082`, `read_only: true` with `tmpfs: [/tmp]`, `mem_limit: 512m`,
`pids_limit: 64`, a `/readyz` healthcheck, and `depends_on` on `migrate` completed and
`redis` healthy — the `media-worker` block copied and trimmed. `nginx.depends_on` gains
`upload: {condition: service_healthy}`.

**The volume-initialisation trap, named because it is silent.** A named Docker volume is
populated from the image of whichever container mounts it first, and only if that image has
content at the mount point. Today only the worker image pre-creates
`/media/{source,hls,thumbnails,temporary}` (`Dockerfile:32`); nginx mounts `:ro` and has
nothing there. With three mounters the order is not ours to choose, so **both binaries
`MkdirAll` their roots at startup** and the Dockerfile's `mkdir` stays only as belt. The
containers are `read_only: true`, but `/media` is a volume mount and remains writable.

**No new secret and no new environment variable beyond `HTTP_ADDR`** — a direct consequence
of D-2b. `.env.example` is unchanged.

### b. The migration collides with the course test canary — the third time

Next real migration is `00005`. The course
`repository_integration_test.go` subtest **"failed migration rolls back its complete
transaction"** (line 190) builds an `fstest.MapFS` (line 195) of the four real migrations
plus a synthetic **`00005_failure.sql`** (line 200). The real `00005_media.sql` collides, so
the migration task adds it to that MapFS and **renumbers the canary to
`00006_failure.sql`** — the pattern 4a used for `00003` (commit `fd90fc9`) and 4b for
`00004`. The identity `schema_integration_test.go` reads via `os.DirFS`
(`schema_integration_test.go:57-58`) and picks up `00005` automatically.

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

- **`proxy_request_buffering off` is mandatory, not a preference.** The nginx container is
  `read_only: true` with `tmpfs: [/tmp, /var/cache/nginx]` and
  `client_body_temp_path /tmp/client_temp` (`nginx.conf:8`), so buffered request bodies land
  in **RAM**. Streaming keeps memory bounded by the chunk, not by the upload.
- **The global `client_max_body_size 1m` (`nginx.conf:20`) stays**, overridden only inside
  this location. Q2's 2 GiB maximum is met by **chunking**: 8 MiB chunks × 256 requests, so
  no single request is ever large. This is why the file limit and the body limit are two
  different numbers, and why the 4a and 4b claims that `client_max_body_size` is untouched
  remain true everywhere else.
- **A dedicated zone, not `api`.** 256 chunk `PATCH`es through `api`'s
  `10r/s burst=20 nodelay` would delay a legitimate upload behind its own rate limit and
  interleave badly with the editor's other calls. 30 r/s bounds a single uploader to roughly
  240 MiB/s of request starts — far above any real link — while keeping a cheap ceiling.
- **`/media` and `/media/` keep returning 403** (`nginx.conf:44-45`). Neither line changes,
  and a smoke assertion continues to prove it.

### d. `make smoke` breaks the moment the first handler is registered

`infrastructure/tests/smoke.sh` asserts the worker exposes **zero** handlers
(`if(!Array.isArray(status.handlers)||status.handlers.length!==0) throw`) — an assertion
written to prove the foundation was not pretending to transcode. Phase 5 inverts it: the
worker must now report exactly `["media:transcode"]`. Three assertions are added in the same
edit: `/uploads/` without a ticket answers `404` (not `403` and not a proxy error),
an over-cap body answers `413`, and `/media/example/master.m3u8` still answers `403`.
The worker's `/status` `Stage` field moves from `"foundation"` to `"media"`.

### e. `make check` gains no step, but `make integration` gains no ffmpeg

The manifest gate order is unchanged, so `.claude/standards.md` §Машинні гейти is untouched.
But the integration container is `golang:1.27.1-bookworm` mounting only `backend/`
(`infrastructure/tests/integration.sh`), and **it has no ffmpeg**. Two rules follow:

- **No Go test may depend on ffmpeg being present.** A conditionally skipped test would be
  exactly the "skipped reported as passed" trap AGENTS.md forbids, so there is none. The
  transcoder adapter is split so that what is *pure* (argv construction, ffprobe JSON
  parsing, ladder selection) is unit-tested exactly, and what is *processual* (subprocess
  lifetime, process-group cancellation) is proven against `/bin/sleep`, which the image has.
- **The real ffmpeg run is proven in Compose**, where the worker image has ffmpeg: `make e2e`
  uploads a generated clip and waits for `ready`, and `make smoke` proves the handler is
  registered. This is a genuine evidence gap between `make integration` and `make e2e`, and
  it is named rather than papered over. Trigger to close it: a transcode regression that
  e2e can detect but not localise — then an ffmpeg-bearing integration image earns its cost.
- **No Redis Testcontainer is introduced.** The queue adapter's one real behaviour
  (`ErrTaskIDConflict` on a duplicate) is proven at the application layer with a fake port
  returning that sentinel; the real Asynq round trip is proven in Compose. A second
  container type for one sentinel is not worth its minutes on every run.

### f. e2e needs a video, and video bytes may not be committed

AGENTS.md forbids committing media bytes, so the fixture cannot live in the repository.
`e2e/media.setup.ts` (in the existing `setup` project, whose `testMatch` becomes a two-entry
array) generates it inside the container that already has ffmpeg and copies it out:

```
docker compose exec -T media-worker ffmpeg -nostdin -loglevel error -f lavfi \
  -i testsrc2=size=640x360:rate=15 -f lavfi -i sine=frequency=440 -t 2 \
  -c:v libx264 -preset veryfast -c:a aac -y /media/temporary/e2e-source.mp4
docker compose exec -T media-worker cat /media/temporary/e2e-source.mp4 > playwright/.fixtures/e2e-source.mp4
```

`playwright/.fixtures/` joins `playwright/.auth/` in `.gitignore`. The clip is ~640×360, so
D-11's no-upscaling rule makes the produced master carry exactly `360p` — which is what the
e2e assertion checks, and it is a real product property rather than a test artefact.

**Re-run caveat**, the generalised 4b §5.f trap: `make e2e` alone does not re-seed, and the
media spec creates its own course and lesson in a per-run slug namespace (the 4a §5.9
idiom), so it is re-runnable. It uses the shared `storageState` and performs **zero**
logins, leaving the 4a D-7 budget (≤7 per run, ≤3 consecutive) untouched.

### g. Two kit components get their first consumer, five years late

`FileUpload` (`controls.tsx:165-178`) and `ProgressBar` (`data.tsx:7-23`) were built in
foundation from the brief's §6 component list and have **no consumer anywhere outside
`ui.test.tsx`** — verified by grep across `frontend/src` and `frontend/e2e`. So they exist
but are unproven in a real screen, and two gaps are expected:

- `FileUpload` is a `FileTrigger` wrapping a `Button`; it has no drag-and-drop, no
  `acceptedFileTypes` opinion of its own and no busy state. The field passes
  `acceptedFileTypes` (a `FileTriggerProps` member) and disables the button while an upload
  is in flight — no kit change unless a prop is genuinely missing.
- `ProgressBar` renders `percentage ?? 35` as the fill width (`data.tsx:17`), which is a
  sensible indeterminate placeholder but means the field must always pass a real `value`
  during transfer and switch to a determinate-unknown presentation during `processing`,
  where no percentage exists.

If either needs a prop, the change lands in the kit with every existing caller updated in the
same task — the 4a A08 rule. The kit stays i18n-free: labels arrive as props through `t()`.

### h. The editor form grows a non-form child

`CourseEditorView` is one React Hook Form over the whole tree (`useFieldArray` per module and
per lesson). The video field is **not** an RHF field: it has no value in the `PUT` body, and
its state lives in a separate query. It mounts inside the existing lesson `fieldset`
(`CourseEditorView.tsx:231`) as a sibling of the RHF inputs, with its own mutation. Two
consequences, both accepted: an upload does **not** mark the form dirty, and saving the tree
does **not** cancel an upload in flight. Removing a lesson from the form and saving *will*
orphan its asset — which is exactly D-4's `SET NULL` path, and the field warns before the
remove button acts on a lesson that has video.

### i. Deleting a course or a module now costs bytes

`ON DELETE CASCADE` from `courses` → `course_modules` → `course_lessons` already existed
(`00003_authoring.sql`); with D-4 the assets of every lesson underneath become orphans and
the sweeper reclaims them. A 4a draft delete is therefore no longer purely a database
operation — it schedules disk work. Bounded by D-5 (at most two generations per lesson) and
by the sweeper's 50-row cap.

### j. `.gitignore` must be anchored before a single Phase 5 file is written

`.gitignore:14` is `media/` with no leading slash, which git matches against a directory of
that name **at any depth**. It therefore already ignores `docs/media/` — so
`docs/media/spec.md`, the approved GATE 1 specification this plan is bound by, is untracked
today — and it would ignore `backend/internal/media/` and `frontend/src/domain/media/` the
moment they exist. The failure mode is the dangerous kind: no error, a clean `git status`,
and a commit that silently contains none of the new module.

The fix is one character plus one line: anchor the data pattern to the repository root and
state the exception.

```gitignore
/media/
!docs/media/
```

`/media/` still ignores a stray top-level media directory, which is all the pattern was ever
for; the media bytes that matter live in the `media-data` Docker volume, never in the working
tree. **M01 makes this change and proves it**, because every task after it writes a file
under one of those paths. AGENTS.md's "never commit media bytes" is unaffected: no rule
changes, only the pattern's scope.

### k. Unchanged on purpose

`/media` and `/media/` at the edge (403, for everyone, Phase 5 claims no playback);
`client_max_body_size 1m` outside `/uploads/`; the `login`, `api` and `web` zones and their
rates; the gate order and `.claude/standards.md`; Prometheus config; `.env.example`;
`seed.sql` and `seed-identities.sql`; the course and identity modules' behaviour (media adds
one read of their tables and changes none of their code); SSR (admin routes are `ssr: false`
and the SSR container carries no `Origin`, so it cannot call a media mutation even if a route
tried — architectural fact #2).

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
     declared_bytes bigint NOT NULL CHECK (declared_bytes BETWEEN 1 AND 2147483648),
     received_bytes bigint NOT NULL DEFAULT 0 CHECK (received_bytes >= 0 AND received_bytes <= declared_bytes),
     content_type text NOT NULL CHECK (content_type IN ('video/mp4','video/quicktime','video/x-matroska')),
     original_name text NOT NULL CHECK (char_length(original_name) BETWEEN 1 AND 255),
     attempt integer NOT NULL DEFAULT 0 CHECK (attempt >= 0),
     failure_reason text CHECK (failure_reason IS NULL OR char_length(failure_reason) BETWEEN 1 AND 200),
     duration_seconds numeric(10,3) CHECK (duration_seconds IS NULL OR duration_seconds > 0),
     renditions text[] NOT NULL DEFAULT '{}',
     expires_at timestamptz NOT NULL,
     evicted_at timestamptz,
     created_at timestamptz NOT NULL DEFAULT now(),
     updated_at timestamptz NOT NULL DEFAULT now(),
     CHECK (updated_at >= created_at),
     CHECK (evicted_at IS NULL OR evicted_at >= created_at),
     CHECK (state <> 'ready' OR array_length(renditions, 1) >= 1),
     CHECK (state <> 'failed' OR failure_reason IS NOT NULL)
   );
   ```
   Two of these CHECKs are load-bearing, in the 4b `deleted_implies_disabled` spirit: **a
   `ready` asset must name at least one rendition** (so "ready" can never be advertised for an
   empty publish), and **a `failed` asset must carry a reason** (so spec S3's "shows why in
   operator terms" cannot be silently unmet). `2147483648` is exactly Q2's 2 GiB.
2. `CREATE UNIQUE INDEX video_assets_pending_per_lesson ON video_assets (lesson_id)
   WHERE lesson_id IS NOT NULL AND evicted_at IS NULL
     AND state IN ('created','uploading','queued','processing');` — D-5's one-in-flight rule.
3. `CREATE INDEX video_assets_by_lesson ON video_assets (lesson_id, generation DESC)
   WHERE evicted_at IS NULL;` — the course-videos read and the eviction lock.
4. `CREATE INDEX video_assets_sweepable ON video_assets (updated_at)
   WHERE evicted_at IS NOT NULL OR lesson_id IS NULL;` — the sweeper's candidate scan, which
   must not table-scan as assets accumulate.

**No backfill and no data risk.** The table is new and empty on every database, fresh or
developer, so there is no analogue of 00003's ownerless-courses hazard: nothing can fail on
existing rows because no existing row is touched. `course_lessons` gains no column, so 4a's
queries and its generated sqlc models are unchanged and a rollback of the API binary needs no
schema change (§7).

**Down:** `DROP TABLE video_assets;` — the three indexes go with it. It is destructive of
authored media links by construction, so it is exercised only in ephemeral Testcontainers
databases, never operationally (the 4a/4b stance). It **cannot fail** on reused data the way
00004's down can, because it re-adds no constraint.

**Bytes are not migrated and not deleted by the migration.** A `goose down 00005` leaves
`/media` exactly as it was; the volume is emptied only by `docker compose down -v`. Stated so
nobody expects the migration to be the cleanup.

**Clean bring-up (unchanged):** `make up` → `make seed-identities` → `make seed` →
`make smoke` → `make e2e`.

---

## 7. Rollback

- **Application rollback is the supported path.** Redeploy the previous API, frontend and
  worker images, stop the `upload` service, and **leave migration 00005 in place**.
  `video_assets` is a new table nothing before Phase 5 reads or writes; `course_lessons` is
  byte-for-byte unchanged, so every 4a query and its generated model still match. A Phase 4
  binary against a 00005 schema is indistinguishable from a Phase 4 binary against a 00004
  schema.
- **What a rollback does to in-flight jobs.** Tasks already in Redis outlive the worker
  image. A rolled-back worker has an **empty mux again**, so an enqueued `media:transcode`
  task is an unknown task type: Asynq fails it visibly and retries it into `archived` — the
  behaviour foundation deliberately chose (`media-worker/main.go:50-51`). Nothing pretends to
  transcode, nothing is marked `ready`, and the assets rest in `queued` forever. Rolling
  **forward** again re-enqueues them through the retry endpoint. To drain instead, delete the
  `media` queue's tasks with the Asynq inspector before rolling back; the assets then sit in
  `queued` and a later retry re-queues them.
- **What a rollback does to already-written media.** Nothing reads `/media` in Phase 4 or
  Phase 5 — the edge refuses it and no binary serves it — so published bytes are inert. They
  are also **unreachable by the sweeper**, which lives in the rolled-back worker, so a long
  rollback accumulates orphaned directories on the volume. That is a disk cost, not a
  correctness or a disclosure one, and it is reclaimed the moment the worker rolls forward.
  Recorded rather than mitigated, because a cleanup path outside the worker would be a fourth
  place that may delete media.
- **Do not `goose down 00005` to recover from an application defect.** It drops every
  lesson↔video link while leaving the bytes on disk, converting a bad state into
  unattributable garbage. A forward migration fixes a schema mistake; the down migration
  exists for tests.
- **Frontend rollback alone is inert as a security measure** (the 4a stance): removing the
  upload control grants nothing, because the mint endpoint, the ticket state machine and the
  edge's `/media` refusal stand on their own.
- **Nginx rollback must accompany the upload service.** If the `upload` container is stopped
  while the `/uploads/` location remains, nginx answers `502` for a path the UI no longer
  offers. The location and the service revert together, exactly as 4a's four gate-order files
  do.

---

## 8. Business error vocabulary

Server-enforced, mapped once in the media handler, contract-proven. New values in the
OpenAPI `error.code` enum:

| Code | HTTP | Meaning | Where decided |
|---|---|---|---|
| `lesson_not_found` | 404 | the lesson id is unknown, or names a lesson the caller may not reach past the role gate | media adapter, inside the minting transaction |
| `upload_in_progress` | 409 | the lesson already has a non-terminal, unexpired asset (D-5) | the partial unique index, mapped from `23505` — never a pre-check, which two concurrent minters would both pass (the 4a `slug_taken` lesson) |
| `unsupported_media` | 400 | the declared content type is outside Q2's set (at mint), or the container sniff failed (at completion) | domain `AcceptSource` / the upload service's sniff |
| `upload_too_large` | 400 | the declared length exceeds 2 GiB, or a `PATCH` would push the offset past it | domain bound / the tus server |
| `video_not_found` | 404 | the video id is unknown or evicted, for status and retry | media adapter |
| `video_not_retryable` | 409 | retry was asked for an asset that is not `failed` | domain transition matrix |

Reused unchanged: `authentication_required` (401), `forbidden` (403, the wrong-role
refusal), `course_forbidden` (403, the non-owning-instructor refusal — the same meaning the
course module gives it), `origin_forbidden` (403), `invalid_request` (400),
`dependency_unavailable` (503), `internal_error` (500).

A business refusal is never a 500 (spec criterion 3). **No error message ever contains a
filesystem path, a storage key, a ticket, an ffmpeg command line or an ffmpeg stderr dump**;
`failure_reason` is drawn from the closed set of D-14's reason constants, and the raw stderr
is neither stored nor logged.

The tus surface answers tus's own vocabulary, not this one: `404` for an unknown, expired or
consumed ticket; `409` for a wrong `Upload-Offset`; `413` for a chunk past the declared
length; `412` for a missing or wrong `Tus-Resumable`; `403 origin_forbidden` — no, the tus
endpoints are **not** same-origin gated, because the ticket is the capability and a
cross-origin `PATCH` with a valid ticket is the ticket holder exercising it. Stated
explicitly so the absence is a decision rather than an omission.

---

## 9. Acceptance scenarios

Per `acceptance-spec`: three groups (all present), one observable result per row, one
evidence level per row, exactly one owning task, and every owner able to produce its evidence
**at the time that task lands**. Mutation rows assume a valid `Origin` unless the row is
about `Origin`. **Manual rows: none.** Every task owns at least one row.

### Happy path

| # | Scenario | Precondition | Action | Observable result | Evidence | Owner |
|---|---|---|---|---|---|---|
| P01 | An admin mints an upload ticket | admin session; a saved lesson exists | `POST /api/v1/admin/lessons/{id}/video` with a 5 MiB mp4's metadata | `201` carrying `videoId`, `uploadUrl` under `/uploads/`, `chunkBytes`, `maxBytes` and `expiresAt`; a following status read shows the asset in `created` | contract `handler_test.go` | M08 |
| P02 | An owning instructor mints | instructor session; a lesson in a course they own | the same `POST` | `201` with the same shape | contract | M08 |
| P03 | tus create → PATCH → completion lands the source and queues the job | Testcontainers PostgreSQL; a minted ticket; a fake queue port | `POST /uploads/`, then `PATCH` the whole body | the file exists at the `source/{videoId}/original` key with the declared byte count, the asset reads `queued`, and the fake queue recorded exactly one enqueue for `{assetId, attempt: 0}` | integration `tus/server_test.go` + repository | M09 |
| P04 | The worker transcodes a real clip | Compose up; the generated 640×360 fixture uploaded | wait for the terminal state | the asset reads `ready`; `hls/{videoId}/master.m3u8`, `hls/{videoId}/360p/init.mp4`, at least one `.m4s` segment and `thumbnails/{videoId}.jpg` all exist inside the worker container | e2e `media.spec.ts` | M14 |
| P05 | The editor advances to ready without a reload | Compose up; an upload just completed | keep the editor open | the lesson's state label moves through `processing` to `ready` with no navigation and no manual refresh | e2e | M14 |
| P06 | Replacing a video keeps exactly two generations | Testcontainers; a lesson with one `ready` asset | publish a second asset for the same lesson | both assets read `ready`, `evicted_at` is null on both, and the higher `generation` is the current one | integration | M05 |
| P07 | The course videos read answers one row per lesson | admin session; a course with two lessons, one carrying video | `GET /api/v1/admin/courses/{id}/videos` | `200` with exactly one item, carrying `lessonId`, `videoId`, `state` and `generation`, and no storage key or path member | contract | M08 |
| P08 | The full browser journey | Compose up; the fixture generated; admin `storageState` | sign in as admin, open a course editor, choose the clip on a saved lesson, wait | the progress bar advances, the lesson shows `processing` and then `ready`, and a reload still shows `ready` | e2e | M14 |
| P09 | The editor renders the control and the progress | a loaded tree with one saved lesson | choose a file in the component test | the `FileUpload` control is present for that lesson and a determinate `ProgressBar` renders while the mocked transfer runs | component `LessonVideoField.test.tsx` | M13 |
| P10 | The worker registers the media handler | Compose up | `GET /status` on the worker | `handlers` is exactly `["media:transcode"]` and `stage` is `"media"` | `make smoke` | M10 |
| P11 | Polling stops at a terminal state | MSW answering `processing` then `ready` | mount the editor and let two intervals pass | requests continue while any row is non-terminal and **stop** after every row is terminal | component with MSW | M12 |

### Boundaries

| # | Scenario | Precondition | Action | Observable result | Evidence | Owner |
|---|---|---|---|---|---|---|
| B01 | The size ceiling is exactly 2 GiB | admin session | mint with `2147483648`, then with `2147483649` | `201`, then `400 upload_too_large`; no asset is created for the second | contract | M08 |
| B02 | A chunk beyond the declared length is refused | a ticket for 1 MiB, 1 MiB already received | `PATCH` one more byte | `413`, and a following `HEAD` reports the unchanged offset | integration `tus/server_test.go` | M09 |
| B03 | An interrupted upload resumes from its offset | a ticket whose transfer died after 2 of 5 chunks | `HEAD`, then `PATCH` from the reported offset | `HEAD` reports exactly the received byte count, the resumed transfer completes, and the stored file equals the source byte-for-byte | integration | M09 |
| B04 | A zero-length upload is refused at mint | admin session | mint with `declaredBytes: 0` | `400 invalid_request`; no asset row exists | contract | M08 |
| B05 | The ladder never upscales, end to end | Compose up; the 640×360 fixture | transcode it | `master.m3u8` lists exactly one variant named `360p`; no `720p` or `1080p` directory exists | e2e | M14 |
| B06 | A third generation evicts the oldest | Testcontainers; a lesson with two `ready` assets | publish a third | exactly two assets have `evicted_at IS NULL`, the evicted one is the lowest generation, and after the sweeper runs its directories and its row are gone | integration | M05 |
| B07 | Migration over a populated database | 5 demo courses, 3 identity fixtures, modules and lessons present | apply `00005_media.sql` | `video_assets` exists with its three indexes, every pre-existing row in every table is intact, and `course_lessons` is structurally unchanged | integration `repository_integration_test.go` | M02 |
| B08 | The failing-migration canary still rolls back | migration 00005 committed | run the course integration suite | the **"failed migration rolls back its complete transaction"** subtest passes with the real `00005_media.sql` in its `fstest.MapFS` and the canary renumbered to `00006_failure.sql` | `make integration` | M02 |
| B09 | Contract and client stay in step | all OpenAPI and query changes committed | `make check-generated` | sqlc and Orval output matches the committed generated files; exit 0 | `make check-generated` | M12 |
| B10 | The ordered gate passes | clean checkout | `make check` | every step exits 0 with the new surface, including a worker that actually processed a job | `make check` | M15 |
| B11 | The storage adapter cannot be walked out of | a tempdir root containing a symlink to `/etc` | resolve keys with `..`, an absolute path, a non-UUID id, and a leaf through the symlink | each is refused before any file is opened, and no file outside the root is read or written | integration `storage_integration_test.go` | M06 |
| B12 | Publication is all-or-nothing | a tempdir root; a fully built working directory | publish, then publish an interrupted build | after success `hls/{id}/master.m3u8` exists; after the interrupted attempt **no** `hls/{id}` directory exists at all — never a partial one | integration | M06 |
| B13 | A non-video source is refused at probe | a recorded ffprobe JSON for a text file and for a video with no audio stream | parse and accept | both are refused as `unsupported_source`; a valid H.264/AAC probe is accepted | unit `transcoder_test.go` | M07 |
| B14 | A deleted lesson leaves a reapable orphan, not a leak | Testcontainers; a lesson with a `ready` asset | delete the lesson row (what a 4a `PUT` does) | the asset survives with `lesson_id IS NULL`, is absent from the course videos read, and the sweeper then deletes its bytes and its row | integration | M05 |
| B15 | A course with no video reads empty | admin session; a course whose lessons have no assets | `GET …/videos` | `200` with `items: []`, and the editor renders no video state for any lesson | contract | M08 |
| B16 | An expired ticket does not lock a lesson | Testcontainers; a lesson with a `created` asset whose `expires_at` has passed | mint again | `201`; the stale asset carries `evicted_at`, and exactly one non-terminal asset exists for the lesson | integration | M05 |
| B17 | The ladder is chosen by the source, never by the client | none | select renditions for 640×360, 1280×720 and 1920×1080 probes | `[360p]`, `[360p,720p]` and `[360p,720p,1080p]` respectively; no case produces a rendition taller than the source | unit `source_test.go` | M03 |
| B18 | The container is sniffed before anything is queued | a ticket whose bytes are a renamed text file | complete the upload | `400 unsupported_media`, the asset reads `failed` with a reason, **no** enqueue was recorded, and no file exists under `source/` | integration | M09 |
| B19 | The transcoding metrics exist and are labelled | Compose up; one job completed | scrape the worker's `/metrics` | `courses_transcoding_job_duration_seconds` carries an `outcome="ready"` observation and `courses_transcoding_failures_total` is present | `make smoke` | M10 |
| B20 | The upload control is operable without a mouse | the editor rendered with one saved lesson | tab to the control and activate it | the file trigger is focus-visible, at least 44px, reachable by keyboard, the progress bar exposes an accessible value, and axe reports no violations | component with axe | M13 |
| B21 | Every Phase 5 path is actually tracked by git | the anchored `.gitignore` committed | `git check-ignore -v docs/media/plan.md backend/internal/media/domain/asset.go frontend/src/domain/media/video.ts` | none of the three matches an ignore rule (exit code 1), and a stray top-level `media/` directory still does | `git check-ignore` in the M01 report | M01 |

### Refusals

| # | Scenario | Precondition | Action | Observable result | Evidence | Owner |
|---|---|---|---|---|---|---|
| R01 | Anonymous is refused everywhere | no cookie, valid `Origin` | mint, status read, retry | each answers `401 authentication_required` and no repository method is called | contract | M08 |
| R02 | A student is refused everywhere | student session | the same three | each answers `403 forbidden` | contract | M08 |
| R03 | A non-owning instructor is refused | instructor B; a lesson in a course owned by A | mint, status read, retry | each answers `403 course_forbidden`, and no asset is created | contract | M08 |
| R04 | Cross-site is refused first | a valid admin cookie, a foreign `Origin` | mint and retry | each answers `403 origin_forbidden` before the cookie is read, with no state change | contract | M08 |
| R05 | An unknown lesson is not found | admin session, an unused uuid | mint | `404 lesson_not_found` | contract | M08 |
| R06 | A malformed id is rejected before lookup | admin session | `POST /api/v1/admin/lessons/not-a-uuid/video` | `400 invalid_request` and no repository call | contract | M08 |
| R07 | A second upload cannot start while one is running | a lesson with an unexpired `uploading` asset | mint again | `409 upload_in_progress`, mapped from the partial index's `23505`; still exactly one non-terminal asset | contract | M08 |
| R08 | A failed asset can be retried | admin session; an asset in `failed` | `POST /videos/{id}/retry` | `200`; the asset reads `queued`, `attempt` incremented, and one enqueue was recorded | contract | M08 |
| R09 | A ready asset cannot be retried | admin session; an asset in `ready` | the same call | `409 video_not_retryable`; the asset is untouched and no enqueue happens | contract | M08 |
| R10 | An unknown, consumed or expired ticket writes nothing | the tus server; three such ticket ids | `PATCH` each with a valid chunk | each answers `404`; no file is created and no offset is recorded | integration | M09 |
| R11 | A wrong offset is a conflict, not a corruption | a ticket at offset 1 MiB | `PATCH` declaring offset 0 | `409`; the stored file is still exactly 1 MiB and `HEAD` still reports 1 MiB | integration | M09 |
| R12 | An oversized chunk is refused at the edge | Compose up; a valid ticket | `PATCH` a 32 MiB body | nginx answers `413`, the request never reaches the upload service, and a following `HEAD` reports the unchanged offset | `make smoke` | M11 |
| R13 | `/media/*` stays refused for everyone | Compose up | request `/media`, `/media/`, `/media/x/master.m3u8` anonymously and with an admin cookie | every one answers `403`, before and after a successful transcode | `make smoke` | M11 |
| R14 | A transient failure retries, then rests failed | fakes: a transcoder failing with a transient error on the final permitted attempt | run the job | the asset reads `failed` with an operator-readable reason from the closed set, the reason contains no path and no stderr, and the error returned to Asynq is not `SkipRetry` on earlier attempts | application fake | M04 |
| R15 | A malformed source is never retried | fakes: a probe rejecting the source | run the job | the asset reads `failed` immediately, and the handler returns an error wrapping `asynq.SkipRetry` | application fake | M04 |
| R16 | A retried job cannot corrupt a published asset | fakes: an asset already `ready` | run the job again | it returns success having called neither the transcoder nor publish; the recorded renditions and the storage are untouched | application fake | M04 |
| R17 | A duplicate enqueue is not a failure | a fake queue returning `asynq.ErrTaskIDConflict` | complete an upload | the use case succeeds and the asset still reads `queued` — the wanted task already exists | application fake | M04 |
| R18 | Cancellation leaves nothing behind and stays re-runnable | fakes: a transcoder blocking until the context is cancelled | cancel mid-job | the working directory is removed, no `hls/{id}` key exists, the asset is not `ready` and not `failed`, and a later claim succeeds | application fake | M04 |
| R19 | A dependency failure degrades safely | a repository port returning dependency failure | mint and status read | `503 dependency_unavailable`; no SQL, credential, storage key or path appears in the result | application fake | M04 |
| R20 | Cancelling the transcoder kills its process group | `/bin/sleep` standing in for ffmpeg, spawned with a child | cancel the context | both the process and its child are gone within the wait delay, and the call returns a cancellation distinct from a failure | unit `transcoder_test.go` | M07 |
| R21 | No user-controlled string reaches ffmpeg | an asset whose `original_name` is `"; rm -rf / --preset x.mp4"` | build the render and thumbnail argv | the produced argv is compared element-by-element against the expected fixed list; the original name appears in neither, and every path element is the asset's uuid | unit | M07 |
| R22 | Media errors disclose no filesystem detail | a storage failure forced through the port | mint and status read | the response body's `message` contains no `/media`, no key, no filename and no ffmpeg text | contract | M08 |
| R23 | Business codes reach the frontend domain intact | MSW answering `409 upload_in_progress`, `400 unsupported_media`, `403 course_forbidden`, `500` | call the matching use cases | the first three surface as their named domain outcomes and the last as `dependency_unavailable`; none collapses into a generic failure | adapter with MSW | M12 |
| R24 | An unsaved lesson offers no upload | the editor with a newly added, unsaved lesson | inspect that lesson | no upload control is rendered, a localized hint explains that the course must be saved first, and no request is made | component | M13 |
| R25 | A failed transfer is retryable and claims nothing | the upload service answering `503` mid-transfer | attempt an upload | a retryable error is shown with a working retry control, and the lesson does not display a video state | component | M13 |
| R26 | The client guard is not the mechanism | a student signed in, in the browser | call the mint endpoint directly through `page.request` | `403 forbidden`, though the SPA never renders the control | e2e | M14 |
| R27 | An unsupported codec set is refused by the domain | probes for HEVC+AAC, H.264+Opus, and video with no audio | evaluate acceptance | HEVC+AAC is accepted (Q2), H.264+Opus and audio-less are refused as `unsupported_source` | unit | M03 |
| R28 | Ownership is resolved transactionally | Testcontainers; instructor B and a course owned by A | resolve the lesson's owner as B, concurrently with an admin reassigning the course to B | the resolution either refuses B against owner A or admits B against owner B — never a mixed read; the `FOR SHARE` lock serialises the two | integration `ownership.go` | M05 |

Group counts: **11 happy, 21 boundary, 28 refusal = 60 rows.** Manual rows: 0. Every task
owns at least one row; no row has two owners.

---

## 10. Task cut

Cut along **executable evidence and environment boundaries**, per `docs/identity/report.md`
§I06 п.8 and the 4a/4b practice — never along file or layer labels. Each task is 0.5–2 h,
names the environment its evidence needs, and regenerates its own sqlc/Orval paths (4a NB-9)
so trunk is never red at `check-generated` between tasks.

| ID | Task | Est. | Depends on | Environment / evidence | Owns |
|---|---|---:|---|---|---|
| M01 | Anchor `.gitignore` (`/media/` + `!docs/media/`) and commit the now-visible `docs/media/{spec,plan}.md`; ADR 008 (tus ingest: protocol, the third executable, the ticket capability) and ADR 009 (fMP4/CMAF `.m4s`, superseding the segment clause of ADR 005) | 0.5h | GATE 2 green | `git check-ignore`; documents; `.claude/standards.md` Мірник consistency | B21 |
| M02 | Migration `00005_media.sql` with its CHECKs and three indexes; the course canary fix (real `00005` into the `fstest.MapFS` at `repository_integration_test.go:195`, canary renumbered to `00006_failure.sql` at `:200`) | 1.5h | M01 | Testcontainers PostgreSQL (`make integration`) | B07, B08 |
| M03 | Media domain: `VideoAsset`, the six states and their matrix, `StorageKey` and its kinds, `Actor.CanEditLesson`, ticket bounds, `SourceProbe`/`AcceptSource`/`SelectLadder`, the error set; and the four port interfaces | 1.5h | — | Go unit tests, stdlib only (`make test`) | B17, R27 |
| M04 | Application use cases over fakes: mint/begin/complete/cancel/list/retry and the twelve-step `Process` plus `Sweep` — retry exhaustion, malformed source, no-op on `ready`, duplicate enqueue, cancellation, dependency failure, deadlines | 1.75h | M03 | Go unit tests with fake ports (`make test`) | R14, R15, R16, R17, R18, R19 |
| M05 | PostgreSQL adapter and `media.sql`: the guarded state UPDATEs on `pool.Begin`, the publish-and-evict transaction, the sweeper queries, and the `LessonOwnership` join with `FOR SHARE` | 2h | M02, M04 | Testcontainers PostgreSQL (`make integration`) | P06, B06, B14, B16, R28 |
| M06 | Filesystem `MediaStorage`: key→path resolution in an owned root, `os.Root` opens, traversal and symlink refusal, append-at-offset, atomic publish, recursive purge | 1.5h | M03 | Go integration against a tempdir, no database (`make integration`) | B11, B12 |
| M07 | ffmpeg `Transcoder`: the argv builders, ffprobe JSON parsing, `exec.CommandContext` with `Setpgid`, group cancellation and `WaitDelay` | 1.75h | M03 | Go unit tests; process behaviour proven against `/bin/sleep` — **no ffmpeg needed** (`make test`) | B13, R20, R21 |
| M08 | The API surface: the three operations, DTOs and bounds, the actor mapping, error mapping, `RegisterAdmin` on the 4a seam, `cmd/api` wiring, the OpenAPI paths/schemas/codes and generated output | 2h | M05 | Go contract tests against `openapi.yaml` (`make test`) | P01, P02, P07, B01, B04, B15, R01–R09, R22 |
| M09 | The third executable: tus core + creation + expiration, the ticket state machine, the container sniff, the atomic move into `source/`, the enqueue; `cmd/upload/main.go`; the Dockerfile `upload` stage | 2h | M06, M08 | Testcontainers PostgreSQL + fake queue port (`make integration`) | P03, B02, B03, B18, R10, R11 |
| M10 | Worker wiring: the media handler on the mux, the payload, the lease and claim, the `ErrorHandler`, the sweep at start, the two new metrics, `/status` reporting the handler; the `smoke.sh` handler assertion inverted | 1.5h | M07, M09 | Compose (`make smoke`) | P10, B19 |
| M11 | Edge and topology: the `upload` Compose service and its limits, nginx's `upload` zone and `/uploads/` location, `nginx.depends_on`, the remaining `smoke.sh` assertions, README | 1h | M10 | Compose (`make smoke`) | R12, R13 |
| M12 | Frontend adapter and application: the domain types, the port, zod schemas, the Orval-backed repository with business-code mapping, the hand-written tus client and its transport, `admin.ts` extension, the polling query options; regenerates Orval | 1.75h | M08 | Rstest with MSW (`make test`), `make check-generated` | P11, B09, R23 |
| M13 | `LessonVideoField` and its mount inside `CourseEditorView`: the file control, progress, state badges, failure reason and retry, the unsaved-lesson hint, degraded states, uk/en, accessibility | 2h | M12 | Rstest component tests including axe | P09, B20, R24, R25 |
| M14 | The browser journey: `media.setup.ts` generating the clip, `media.spec.ts` (upload → processing → ready, the produced ladder, the direct-API refusal), the playwright `setup` `testMatch` change | 1.5h | M11, M13 | Playwright against Compose (`make e2e`) | P04, P05, P08, B05, R26 |
| M15 | Ordered gate and report: full `make check` from a clean checkout; `docs/media/report.md` with measured results, limitations and the next task; `AGENTS.md`'s current-delivery line moved to Phase 5 | 1h | M01–M14 | `make check`, twelve steps | B10 |

**Total: 23.25 h** — the arithmetic of the column above (four tasks at 2 h, three at 1.75 h,
four at 1.5 h, two at 1 h, one at 0.5 h).

**Which tasks need Docker:** M02, M05, M06, M09 need Testcontainers (`make integration`);
M10, M11 need Compose (`make smoke`); M14 needs Compose (`make e2e`); M15 needs everything.
M01, M03, M04, M07, M08, M12, M13 need no Docker at all.

**Which tasks need a real ffmpeg run:** only M14, and only inside the worker container.
**None** of M03–M13 executes ffmpeg, by construction (§5.e).

**How a worker job is proven, and where the line is.**

- In `make integration`: everything about the job **except the subprocess** — the state
  machine and its guards (M04, fakes), the transactional publish and eviction (M05, real
  PostgreSQL), the atomic rename and the traversal refusals (M06, real filesystem), the tus
  offsets and the enqueue (M09).
- In `make test`: the exact ffmpeg argv (M07, R21) and the process-group kill (M07, R20,
  against `/bin/sleep`).
- In `make smoke`: that the handler is registered at all (M10, P10) and that the metrics it
  emits exist (M10, B19) — the check that would have caught "the worker starts but does
  nothing", which is precisely the state Phase 5 inherits.
- In `make e2e`: the one thing only a real ffmpeg can show — that a real clip becomes a real
  master playlist with real segments and a real thumbnail, and that the editor sees it
  (M14, P04/P05/P08/B05).

**Sequencing notes.**

- **M01 first, and it is not ceremony.** Until `.gitignore` is anchored, every file every
  later task writes is invisible to git (§5.j) — the phase would produce commits that
  contain none of its work. And `.claude/standards.md` makes ADRs a Мірник document, so
  writing code that contradicts ADR 005's segment choice before ADR 009 supersedes it would
  be a standards violation, not a style one.
- **M02 early** because every later Go task runs against the migrated schema, and because the
  canary renumber must land with the migration or `make integration` is red between tasks —
  the lesson 4a and 4b both learned.
- **M03/M04 before M05/M06/M07**, and **M04 split from M05 by evidence environment, not by
  layer**: M04 is everything provable with fakes in `make test`, M05 everything that needs
  real PostgreSQL, M06 everything that needs a real filesystem. M12/M13 split the same way
  (MSW versus a rendered component).
- **M08 before M09 and M12** — the upload service consumes the ticket the API mints, and
  Orval cannot regenerate against a document that does not yet describe the operations.
- **M06 before M09** because the tus server writes through `MediaStorage`, not through
  `os.OpenFile`.
- **M10 before M11** so the handler exists before the edge is asked to route to it; both
  before **M14**, which is the only task that can observe the whole path.
- **M15 last**, and it is the task that moves `AGENTS.md`'s current-delivery line — the phase
  is not "current" until its gate is green.

---

## 11. Riskiest decisions and findings — pointer list

Not a restatement; each points at the section that reasons it out.

0. **`.gitignore:14` (`media/`, unanchored) already ignores `docs/media/` and would ignore
   the entire new module** (§0, §5.j) — the approved spec is untracked today, and without
   M01's one-line anchor a Phase 5 implementer commits nothing while `git status` stays
   clean. Not a decision; a defect found while verifying placement, and the reason M01 runs
   before any other task.
1. **The upload ticket is a database row, not a signed token and not a callback** (§4 D-2b) —
   the load-bearing choice: it keeps identity logic entirely out of the third executable,
   introduces no shared secret, and makes revocation and expiry ordinary columns.
2. **We implement tus core ourselves rather than running tusd** (§4 D-2a) — bounded protocol
   work bought instead of a new image plus two hook contracts; the trigger to reverse it is
   parallel chunks, checksums or S3.
3. **Media models a capability, not a role, so 4a's "third module" trigger does not fire**
   (§4 D-3) — no third copy of the role vocabulary and no refactor imposed on proven code.
4. **The lesson↔video FK is `ON DELETE SET NULL`, swept, not `CASCADE`** (§4 D-4) — because
   `authoring.go:237` deletes lessons on an ordinary course save and `CASCADE` would strand
   the bytes forever.
5. **One in-flight asset per lesson is a partial unique index; two ready generations is a
   transactional rule at publish** (§4 D-5) — an index cannot count to two, and the mapping of
   `23505` to `upload_in_progress` beats a pre-check two minters would both pass.
6. **Publication is atomic because the destination id is fresh** (§4 D-8), with both
   destination-exists branches stated — not because of a swap that would not be atomic.
7. **Q2's "refuse before a job is queued" is met by a container sniff in the upload service;
   the codec half is deferred to the worker's ffprobe** (§4 D-8) — a named deviation from the
   GATE 1 wording, taken to keep ffmpeg out of the byte-receiving image.
8. **Nginx streams uploads (`proxy_request_buffering off`) with a 16 MiB per-chunk cap while
   the global 1 MiB limit stands** (§5.c) — the nginx container's `/tmp` is a tmpfs, so
   buffering would be RAM.
9. **No Go test may depend on ffmpeg; the real transcode is proven only in Compose** (§5.e) —
   an honest evidence gap named with its trigger, rather than a conditionally skipped test.
10. **A rollback leaves in-flight jobs unhandled and published bytes unswept** (§7) — the
    rolled-back worker's mux is empty again, so tasks archive visibly and orphans accumulate
    as a disk cost until the worker rolls forward.
