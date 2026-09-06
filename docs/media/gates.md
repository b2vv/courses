# Phase 5 — gates

## GATE 2 — plan-critique (fable, fresh context)

Target: `docs/media/plan.md` edition 1 (2026-09-06). Read against `docs/media/spec.md`
(GATE 1 Q1–Q8), `work/PRODUCT-BRIEF.md` §2/§10/§11/§17/§24.11–13, `docs/admin/plan.md`,
`docs/admin/plan-4b.md`, `.claude/standards.md`, `AGENTS.md`, and the code the plan cites.
Every citation below was opened; two claims were exercised with a real ffmpeg 7.1.1.

Lenses: **4 run / 4 returned** — all four run inline by this critic, none delegated.
No lens returned empty.

Verification notes that shape the findings:

- `git log -- .gitignore` shows `5ef6c1f fix: anchor the media ignore pattern and commit the
  Phase 5 spec and plan`; `.gitignore:14` is already `/media/`; `git ls-files docs/media`
  lists `plan.md` and `spec.md`; `git check-ignore -v` on the four paths the plan names
  matches nothing. The plan's §0 "verified with `git check-ignore`" is no longer true.
- The plan's D-11 render argv, run with two renditions on a 1280×720 lavfi clip: the master
  lands at `<work>/master.m3u8` (D-8 holds), but the init segments are
  `360p/init_0.mp4` and `720p/init_1.mp4`, not `init.mp4`.
- The same scale filter on a 1920×1040 source: libx264 aborts with
  `height not divisible by 2 (640x347)`.
- `asynq@v0.26.0/processor.go:245`: on lease expiry the processor cancels the context and
  calls `handleFailedMessage(…, ErrLeaseExpired)` — the task is retried with the **same**
  payload while the previous goroutine may still be running.
- `platform/httpserver/server.go:124` logs `route` (the mux pattern), not the path, so the
  ticket id does not reach Go's access log; D-14's logging claim holds.
- `courses.sql:89` is `FOR UPDATE` (`lockAuthoringCourse`), `:111` is `FOR SHARE`
  (`LockEligibleAuthoringOwner`), `authoring.go:237` deletes omitted lessons — all as cited.
- Line citations checked and correct: `media-worker/main.go:50-54`, `Dockerfile:29-33`,
  `docker-compose.yml:88,117`, `nginx.conf:8,20,44-45`, `smoke.sh` handler assertion,
  `repository_integration_test.go:190,195,200`, `controls.tsx:165-178`, `data.tsx:17`,
  `CourseEditorView.tsx:231,275,288`, `transport.ts:26`, `admin_handler.go:100-110`,
  `authoring.go:76,124,147,186`.

### Blocking

```
[logic ×2 (logic, edge)] The retry and crash-recovery model rests on a lease the schema does not have and a transition the matrix does not have
де:      §1 state matrix; D-10 bullets 2, 4, 6 ("a processing row whose lease has expired", "leaves the asset in processing with an expired lease"); §6 CREATE TABLE
чому:    video_assets has no lease/claim column (id, lesson_id, generation, state, declared_bytes, received_bytes, content_type, original_name, attempt, failure_reason, duration_seconds, renditions, expires_at, evicted_at, created_at, updated_at). The claim is "UPDATE … WHERE state = 'queued'". After a transient failure, a shutdown cancellation or a crash, the row is `processing`; Asynq redelivers the same payload (processor.go:245, same `attempt`), the claim finds no `queued` row, and every retry is a no-op — the asset never reaches `failed` from "a real outcome" (spec criterion 2) and the partial index keeps the lesson locked until the 24 h ticket takeover. Second half: "working directories are per attempt" is false for Asynq's own retries — `attempt` is only incremented by the manual endpoint, so all automatic retries share `temporary/{id}-{attempt}`, including the lease-expired case where the old goroutine is still writing there.
що зробити: Add the transition `processing → queued` (worker, guarded UPDATE, written before returning a retryable error and on cancellation) to the §1 matrix, and make the handler's claim accept only `queued`. For a hard crash (no chance to write), add to D-7 the candidate "`processing` with `updated_at` older than the task timeout → set `queued`" at worker start. Key the working directory as `temporary/{id}-{attempt}-{retryCount}` (from `asynq.GetRetryCount`) or drop the "never share a byte" claim and state that the handler `RemoveAll`s its directory before rendering. Add an M04 row: "a transient failure on a non-final attempt leaves the asset `queued`, and the next run claims it".
вага:    blocking
```

```
[complexity + logic] `ports.LessonOwnership` cannot share the minting transaction with `AssetRepository`
де:      §2 (`ports/ownership.go`, `adapters/postgres/ownership.go`); D-1 option 3; D-3 last paragraph ("takes FOR SHARE … inside the minting transaction")
чому:    The FOR SHARE on `courses` and the INSERT into `video_assets` must be one transaction or the lock guards nothing. Two ports with no shared transaction handle cannot do that, and passing a `pgx.Tx` through a port breaks the port rule (AGENTS.md: ports reference domain only). 4a solved the identical problem inside one repository method (`Repository.Save`: `lockAuthoringCourse` then `actor.CanManage` then writes). M05 will discover this and re-fold the port; M03/M04 will have built fakes for a port that then disappears.
що зробити: Delete `ports/ownership.go` and `adapters/postgres/ownership.go`. `AssetRepository.Mint(ctx, actor, lessonID, request)` runs the lesson→module→course join with FOR SHARE, applies `domain.Actor.CanEditLesson(ownerID)` over the locked row, and inserts, in one `pool.Begin` transaction; `Retry` and `ListCourseVideos` do the same read. Rewrite D-1 option 3 as "the join lives in the media repository", keep the copied-SQL justification, and move R28 to the repository method. Also state the lock order (FOR SHARE on `courses` before the INSERT) — a concurrent 4a `Save` holds FOR UPDATE on the same row and then deletes lessons; taking the course lock first is what makes the two serialise instead of deadlock.
вага:    blocking
```

```
[logic + complexity] Two creation points: the API mint and tus `POST /uploads/`
де:      §1 "The tus surface"; D-2a ("`POST` create"); D-2b (`uploadUrl` is `/uploads/{videoId}`, "the row is the capability"); P03 ("`POST /uploads/`, then `PATCH`"); §5.d smoke ("`/uploads/` without a ticket answers `404`")
чому:    D-2b says the row minted by the API is the upload and its URL is known before any tus call. Then what does `POST /uploads/` create, with what `Upload-Length`, and authorised by what — it has no cookie and no ticket in the path? The `creation` extension exists for clients that have nothing yet; ours has everything. As written, M09 must invent the answer (a second declared length, a metadata header naming the ticket, a second transition the matrix lacks) and the state machine says `created → uploading` happens on "first PATCH", not on POST.
що зробити: Drop the `creation` extension (and `OPTIONS`, which exists for generic-client discovery): the tus surface becomes `HEAD`, `PATCH`, `DELETE` on `/uploads/{videoId}`, tus 1.0.0 core + `expiration`. The API mint *is* the creation. Rewrite P03 as "HEAD → PATCH", the smoke row as "`HEAD /uploads/{unknown-uuid}` answers `404`", and the client as "HEAD for the offset, then sequential PATCH". ADR 008 records that the creation step is the JSON mint.
вага:    blocking
```

```
[edge] The PATCH write protocol is unspecified: whose offset is the truth, and what two concurrent PATCHes do
де:      D-2b; §2 `filesystem/storage.go` ("append-at-offset"); `tus/server_test.go` row list; B03, R11
чому:    Spec criterion 1 ("the source is never lost on a partial or interrupted upload") depends on three cases the plan does not state: (a) the process dies after appending a chunk and before updating `received_bytes` — file length > column; the next HEAD reports which? (b) two PATCHes with the same `Upload-Offset` on one ticket (a retrying client whose first request was not really lost) — both append, the file is corrupt, `received_bytes` is wrong by a chunk; (c) the append succeeds and the UPDATE fails. R11 covers only a *wrong* declared offset.
що зробити: State the protocol in D-2b: the column is the truth; a PATCH first takes the row lock and checks `Upload-Offset = received_bytes` (else 409, no write); if the file is longer than the column it is truncated to the column before writing; the append is written and fsynced, then one guarded `UPDATE … SET received_bytes = $new WHERE id = $1 AND received_bytes = $offset AND state IN ('created','uploading') AND expires_at > $now AND evicted_at IS NULL`; zero rows → 409 and the file is truncated back. Add an M09 row: two concurrent PATCHes at the same offset → exactly one 204, and the file length equals `received_bytes`.
вага:    blocking
```

```
[edge + logic] A reload or closed tab during an upload locks the lesson for 24 h with no way out, and polling never stops on it
де:      D-2b (24 h TTL "a resumable upload must survive a closed laptop"); D-5 (409 while a non-terminal unexpired asset exists); D-12 (poll while any row is non-terminal); §2 `ports/LessonVideoRepository.ts` (mint, list, retry, transfer — no cancel); R25 ("the lesson does not display a video state")
чому:    A browser cannot resume a file the user has not re-selected; the tus client's state is in memory. After a reload the list returns the `uploading` row, a new mint answers `409 upload_in_progress` until `expires_at`, the frontend port has no cancel, and the field has no cancel control — so the lesson is dead for a day. R25's "no video state" contradicts the list read, which returns the row. And a `created` row (mint succeeded, transfer never started) keeps `refetchInterval` alive for 24 h in every open editor of that course.
що зробити: Frontend port gains `cancelUpload(videoId)` = tus DELETE (already in the server surface). `LessonVideoField` renders, for a `created`/`uploading` row with no transfer in this tab, "upload interrupted — choose the file again to resume, or cancel", where "choose again" resumes via HEAD → PATCH from the reported offset. D-12's predicate polls only for `queued`/`processing` (plus a transfer in flight in this tab). Reword R25; add an M13 row for the interrupted-row affordance and an M09 row for tus DELETE → `evicted_at` set, following PATCH → 404.
вага:    blocking
```

```
[logic] B19 cannot be observed by its owner at the time it lands
де:      §9 B19 (precondition "one job completed", evidence `make smoke`, owner M10); §10 gate order
чому:    The gate order is `… → smoke → e2e`, and the only thing that uploads a clip is `e2e` (M14). At `make smoke` time no job has run; a labelled Prometheus histogram has no `outcome="ready"` series until it is observed, so the assertion fails on a clean `make check`. Exactly the "task owns a row it cannot observe" trap the acceptance discipline forbids.
що зробити: Pre-initialise every label value at registration (`WithLabelValues` for each outcome and reason constant), so B19 at smoke asserts the metric families and all label values exist at zero. Move "carries an `outcome="ready"` observation after a job" to an M14 e2e row (scrape the worker after P04) or into B10.
вага:    blocking
```

```
[edge] A dependency failure between the `queued` commit and the enqueue leaves an asset nobody can retry
де:      §1 matrix (`uploading → queued` on completion); D-10 (`ErrTaskIDConflict` = success; retry accepts only `failed`); §8 `video_not_retryable`
чому:    `CompleteUpload` must commit the state and enqueue in Redis — two dependencies, no atomicity. Commit-then-enqueue with Redis down yields a `queued` row with no task; enqueue-then-commit with Postgres failing yields a task for an `uploading` row the handler will skip. In both orders the asset is stranded, and `POST …/retry` answers `409 video_not_retryable` because the state is not `failed`. The only exit is the 24 h ticket takeover or manual SQL.
що зробити: Fix the order as commit-then-enqueue (D-10) and let `Retry` accept `queued` as well as `failed` — a re-enqueue under the same task id is idempotent by the plan's own `ErrTaskIDConflict` rule. Add the worker-start sweep candidate "`queued` older than N minutes → re-enqueue". Matrix row becomes `failed | queued → queued` (retry). Add an M04 row: "enqueue fails after commit → the use case returns `dependency_unavailable`, the asset reads `queued`, and a later retry enqueues it".
вага:    blocking
```

```
[edge] D-11's scale filter produces odd dimensions and libx264 refuses them — verified
де:      D-11 render argv (`scale=w=640:h=360:force_original_aspect_ratio=decrease`); R21 expected argv; B17
чому:    Any source whose aspect ratio is not exactly 16:9 (1920×1040 letterboxed exports, 4:3 screen recordings, portrait phone video) scales to an odd height or width. Reproduced with ffmpeg 7.1.1 on a 1920×1040 clip: `height not divisible by 2 (640x347)`, encoder open fails, the job ends `render_failed`. That is a wrong result for a common class of legitimate inputs, and the exact-argv test in R21 would pin the defect.
що зробити: Append `:force_divisible_by=2` to every rendition's scale expression in D-11 and in R21's expected list. Add to B17 or a new M03/M07 row: a 1920×1040 and a 1080×1920 probe select the ladder and the argv carries `force_divisible_by=2`. Record in D-8/ADR 009 the observed init-segment naming for multi-rendition output (`init_0.mp4`, `init_1.mp4`, per-variant index) so Phase 6 does not assume `init.mp4` — P04 holds only because the e2e clip has a single rendition.
вага:    blocking
```

### Non-blocking

```
[AI-slop] §0, §5.j, §11 item 0, M01 and B21 describe as pending a defect that commit 5ef6c1f already fixed
де:      §0 second block bullet 1 ("verified with git check-ignore"); §5.j; §11 item 0; M01; B21
чому:    `.gitignore:14` is `/media/` today, both docs are tracked, and check-ignore matches none of the four paths. A reader who trusts "verified" will re-verify and find the plan wrong on its first factual claim; `!docs/media/` is unnecessary with an anchored pattern.
що зробити: Rewrite the §0 bullet as "found and fixed in 5ef6c1f before this edition"; §5.j shrinks to two lines; M01 becomes "ADR 008 + ADR 009" (0.5 h stands); B21 becomes a one-line regression check on M01 or is removed; §11 item 0 is removed.
вага:    non-blocking
```

```
[logic] Q1/§24.11 say the upload service "notifies the API on completion"; the plan has it write `video_assets` and Redis directly — an unnamed deviation
де:      D-2b; spec Q1; brief §24.11 ("сповіщає API про завершення")
чому:    The plan names one GATE 1 deviation (Q2's codec half) and not this one. The design is defensible (§2.12 permits shared modules; the API still sees only metadata), but it makes the upload service a second writer of media state and a second queue producer, and that should be a recorded decision with its trigger, not a silent reading of "notify".
що зробити: Add one paragraph to D-2b naming the deviation from Q1's wording and why (no internal HTTP contract, no cookie in the upload service); list it in §11.
вага:    non-blocking
```

```
[logic] The Q2 deviation (codec check deferred to the worker) needs the gate owner's explicit acceptance, and the sniff should run on the first chunk, not at completion
де:      D-8 "Sniff-before-queue"; §11 item 7
чому:    GATE 1 decisions are binding; the plan may propose, not decide. What the user experiences changes: a wrongly-coded file becomes a `failed` badge minutes later instead of a `400` now. Separately, sniffing at completion means a renamed text file is refused only after 2 GiB have been transferred — the same twelve bytes are available in the first PATCH.
що зробити: Mark D-8's deviation "accepted at GATE 2 by <owner>" or revert it. Move the container sniff to the first PATCH (refuse → `failed` + `unsupported_media`, ticket consumed, nothing under `source/`), keep the completion path as the enqueue point; B18 becomes "the first chunk of a renamed text file".
вага:    non-blocking (accepted deviation requires a signature)
```

```
[complexity] `Setpgid` + process-group kill + a child-spawning R20 is machinery for a problem ffmpeg does not have
де:      §2 `ffmpeg/transcoder.go`; D-10 "Cancellation"; R20
чому:    ffmpeg is single-process (threads, not children); `exec.CommandContext` with the default `cmd.Cancel` (SIGKILL) and `WaitDelay` already ends it. `SysProcAttr{Setpgid}` is Unix-only and needs a build tag or a stub to keep `make test` compiling everywhere.
що зробити: Drop `Setpgid`; R20 becomes "the process is gone within `WaitDelay` and the call returns a cancellation distinct from a failure", still against `/bin/sleep`.
вага:    non-blocking
```

```
[complexity] `duration_seconds` has no Phase 5 reader
де:      §6 schema
чому:    Nothing in Phase 5 reads it; it is a Phase 6 column written early. `renditions` is different — the `ready ⇒ renditions ≥ 1` CHECK makes it load-bearing.
що зробити: Remove `duration_seconds` from 00005 (or keep it and name the Phase 6 consumer in D-8 so it is not a speculative column).
вага:    non-blocking
```

```
[edge] The `content_type` CHECK on the client-declared MIME refuses legitimate files
де:      §6 (`content_type IN ('video/mp4','video/quicktime','video/x-matroska')`); §8 `unsupported_media` "at mint"
чому:    Browsers report `''` or `application/octet-stream` for `.mkv` and sometimes `.mov`; the declared type is the least trustworthy signal and the sniff is the authoritative one. As written a valid mkv is refused before a byte is sent.
що зробити: Make `content_type` nullable display metadata with a length CHECK only; the mint refuses nothing on MIME; the sniff (first chunk, see above) decides. B01/B04 unchanged.
вага:    non-blocking
```

```
[edge] Several rows per lesson: the per-lesson badge rule is undefined, and `failed` assets are never swept
де:      §1 GET row ("one row per non-evicted asset"); P07 title ("one row per lesson"); D-5 ("if three now exist" — three non-evicted, or three ready?); D-7 candidates
чому:    A lesson can hold two `ready`, one in-flight and any number of `failed` rows. `failed` with a lesson is none of D-7's candidate classes, so its source bytes (≤ 2 GiB) linger until a later publish happens to evict it as "oldest". The field must pick one badge from several rows.
що зробити: D-5: publish counts *non-evicted* rows and evicts the lowest generation beyond two — say so. D-12/§5.g: the badge shows the non-terminal row if any, else the highest-generation row; a `failed` row is shown until retried or superseded. Fix P07's title to "one row per asset". Optionally add `failed` older than N days to the sweeper.
вага:    non-blocking
```

```
[edge] `generation` assignment is not defined
де:      D-4/D-5; §6 (`generation >= 1`, no uniqueness)
чому:    "Highest generation that is ready" needs generations that only grow per lesson. If computed over non-evicted rows, a lesson whose assets were all evicted restarts at 1 and later ordering is ambiguous.
що зробити: State: `generation = COALESCE(MAX(generation), 0) + 1` over *all* rows of the lesson (evicted included), computed inside the mint transaction, serialised by the partial index. Add a UNIQUE `(lesson_id, generation)` or accept and say why not.
вага:    non-blocking
```

```
[edge] Sweeper idempotency and row-less orphans
де:      D-7; §5.f (`/media/temporary/e2e-source.mp4`)
чому:    Delete-bytes-then-row can crash in between; the next pass must treat ENOENT as done. A row-driven sweeper never finds files without rows: the e2e fixture written into the volume, or a PATCH that recreates a temp file after a concurrent DELETE+sweep removed the row.
що зробити: State ENOENT-is-success in D-7. Generate the e2e clip into `/tmp` inside the worker container, not `/media/temporary`. Name the upload's scratch key (`temporary/{id}-upload`) so the sweeper's `temporary/{id}-*` glob is known to cover it.
вага:    non-blocking
```

```
[edge] The handler's own step-0 sweep can delete the asset it was enqueued for
де:      D-7 (sweep as step 0); D-4; D-10 bullet 2
чому:    A lesson deleted while `uploading` sets `lesson_id` NULL; the completion still enqueues; the job's sweep removes the row; the load then finds nothing. "ready or evicted → nil" does not cover "not found", which as an error would retry three times.
що зробити: D-10: "not found → return nil". Also check `lesson_id IS NULL` at claim to skip the transcode (a publish for an orphan is wasted work the sweeper undoes).
вага:    non-blocking
```

```
[logic] R12 asks `make smoke` for a valid ticket it cannot mint
де:      R12 (precondition "a valid ticket"; owner M11; evidence smoke)
чому:    smoke.sh has no session, and it already spends 30 login attempts on the login-zone test, so a real login there would meet the 429 it just proved. The 413 half needs no ticket: nginx refuses before routing.
що зробити: Trim R12 to "`PATCH /uploads/{any-uuid}` with a 32 MiB body → nginx 413; the upload service's request counter for that route did not move". Move "HEAD reports the unchanged offset" to M09's B02.
вага:    non-blocking
```

```
[complexity] `video_assets_sweepable` indexes a column the sweeper does not order by and misses one candidate class
де:      §6 index 4; D-7 candidates
чому:    The sweeper takes up to 50 rows in any order; `(updated_at)` buys nothing. The third class (non-terminal, expired > 24 h) is not in the predicate and cannot be (no `now()` in a partial index), so it scans anyway.
що зробити: Index on the predicate alone (`… WHERE evicted_at IS NOT NULL OR lesson_id IS NULL`, any cheap key such as `id`) and add `(expires_at) WHERE evicted_at IS NULL AND state IN ('created','uploading','queued','processing')` for the expiry classes, or state that the table is small enough to scan.
вага:    non-blocking
```

```
[edge] The ticket is also the public asset id, so every co-editor of the course can drive another editor's in-flight upload
де:      D-2b ("unguessable v4 UUID"); §1 GET row returns `videoId`
чому:    The bearer capability is returned by the list read to every admin and owning instructor. Same authorisation scope, so no escalation — but "unguessable" is not the property that protects it; the ownership rule at mint is. Should be a stated decision.
що зробити: One sentence in D-2b: the id is visible to everyone who could have minted it; a co-editor cancelling or appending to an in-flight upload is in scope. (Verified: Go logs `route`, not path, so D-14's "no ticket id logged" holds.)
вага:    non-blocking
```

```
[edge] Ladder selection by height alone on portrait sources
де:      D-11 point 3; B17
чому:    A 1080×1920 phone clip selects all three renditions and scales to 202×360, 405×720, 607×1080 — legal once `force_divisible_by=2` lands, but B17 covers only landscape.
що зробити: Add a portrait probe to B17 and state the rule as "by the shorter side" or "by height, portrait accepted as is".
вага:    non-blocking
```

```
[AI-slop] Filler and stale rhetoric
де:      §5.g title ("five years late" — foundation is weeks old); D-2b ("Accepted and recorded so a later reviewer finds the reasoning rather than the symptom"); §11 (restates §4 without adding a pointer the sections lack)
чому:    Reads as ceremony; a wrong number in a heading invites distrust of the right numbers.
що зробити: Cut the three phrases; keep §11 as a bare pointer list.
вага:    non-blocking
```

```
[logic] M09 owns the Dockerfile `upload` stage with Testcontainers evidence that cannot build an image
де:      §10 M09 ("the Dockerfile `upload` stage"; environment Testcontainers + fake queue)
чому:    The stage is first exercised by M11's `make smoke`. Not wrong, but the row/evidence discipline says every owner produces its evidence when the task lands.
що зробити: Move the Dockerfile stage and `cmd/upload/main.go` wiring to M11, or state that M09's stage is proven by `make build` (which is in the gate) and add that to its evidence column.
вага:    non-blocking
```

### Answers to the seven points of special scrutiny

1. **Upload-ticket authorisation without identity logic.** Sound in principle: the row *is* the
   capability, the upload service checks existence/state/expiry only, no cookie, no role. Two
   defects sit beside it: the `creation` extension re-opens the question of who creates
   (blocking 3), and the PATCH protocol leaves single-use and concurrent-use to the
   implementer (blocking 4). Replay: consumption is the `queued` transition, which the guarded
   UPDATE makes single-use once blocking 4's `received_bytes = $offset` predicate is stated.
   Expiry: `expires_at` is checked per request; airtight once it is inside the same UPDATE.
2. **Atomic publish and the crash window.** `rename(2)` onto a fresh `hls/{id}` is atomic and
   the stated "destination exists, asset not ready" branch is right. No bytes are lost: the
   source is never touched by publish, the stale directory is removed before re-rendering,
   and evicted bytes are reaped by the sweeper after commit. Bytes are stranded only for the
   duration of a rollback (§7, named). The gap is not the rename but the *claim after the
   crash* (blocking 1): the row is `processing`, and nothing in the schema lets the next run
   take it.
3. **Partial unique index vs transactional two-ready rule.** The concurrency argument is sound:
   the index makes at most one non-terminal asset per lesson, therefore at most one
   `processing`, therefore publishes for one lesson never overlap and the FOR UPDATE at
   publish is uncontended by design. The expired-ticket takeover is safe under READ
   COMMITTED (the second minter's UPDATE re-evaluates and sees 0 rows, then hits `23505`).
   The 4a `slug_taken` analogy holds. What is undefined is *which* count the publish uses
   (non-blocking) and how `generation` is assigned (non-blocking).
4. **`ON DELETE SET NULL` plus sweeper.** Orphans are reachable: the FK action fires inside
   4a's own `Save` transaction, the `lesson_id IS NULL` predicate is indexed, and the sweeper
   runs at worker start and per job. Two honest limits: no uploads → no sweeps (named in
   D-7's trigger), and a row-less file is never found (non-blocking). The lock order between
   the mint's FOR SHARE on `courses` and 4a's FOR UPDATE + lesson DELETE is deadlock-free
   *only if* the course lock precedes the INSERT — stated as part of blocking 2.
5. **Nginx.** A 2 GiB upload works: 256 × 8 MiB `PATCH`es, each under the 16 MiB location
   cap; nginx enforces `client_max_body_size` from `Content-Length` before proxying even
   with `proxy_request_buffering off`; `proxy_read_timeout 120s` covers an 8 MiB chunk on
   any plausible link; `read_only` + tmpfs is exactly why buffering must be off. The
   `/media` lines are untouched and `^~ /uploads/` cannot shadow `^~ /media/`. R12's "valid
   ticket" precondition is the only defect (non-blocking).
6. **GATE 1 Q2 deviation.** "Refused before a job is queued" is satisfied for the container
   and not for the codec; the plan says so. Two remarks: the acceptance must be the gate
   owner's, not the plan's, and the sniff belongs on the first chunk rather than after 2 GiB
   (non-blocking). The plan does not name its second deviation from Q1's wording
   ("notifies the API") — non-blocking, must be named.
7. **Evidence gap.** The gap between `make integration` and `make e2e` is named honestly and
   the split (pure argv/probe parsing in unit, subprocess lifetime against `/bin/sleep`,
   real ffmpeg only in Compose) is the right one. One row is dishonest by construction:
   B19 at `make smoke` before any job has run (blocking 6). The exact-argv test is the
   right instrument — and it would have pinned the odd-dimension defect (blocking 8) into
   the codebase, which is why that one is blocking rather than cosmetic.

### Accounting

- Lenses: **4 run / 4 returned** (AI-slop, logic holes, excess complexity, edge cases — all
  inline, none delegated). Findings per lens: AI-slop 3 (one collapsed into the stale-§0
  finding), logic 9, complexity 6, edge 12; cross-lens duplicates collapsed with the
  lens count kept in the tag.
- **Blocking: 8 open** — none applied, because this critic was instructed not to edit
  `plan.md`; each carries the exact section/row to change for edition 2.
- **Non-blocking: 16.**
- Verdict: **RED** — edition 2 required; the retry/lease model (1), the ownership port (2)
  and the double creation point (3) each cause rework in M03–M05/M09 if implemented as written.

## GATE 2 round 2 — plan-critique verification (fable, fresh context)

Target: `docs/media/plan.md` **edition 2** (2026-09-06, 1471 lines). Read against
`docs/media/spec.md` (Q1–Q8, Q2 as amended by the gate owner on 2026-09-06 — spec.md:67),
`docs/media/research-codecs.md`, `backend/internal/course/adapters/postgres/authoring.go`,
`backend/sql/queries/courses.sql`, `backend/cmd/media-worker/main.go`, `docker-compose.yml`,
and `asynq@v0.26.0` (`processor.go`, `recoverer.go`, `server.go`, `context.go`, `client.go`).
Every location the plan names for a closed finding was opened. The D-11 argv was executed
verbatim with ffmpeg 7.1.1 on four generated sources; asynq's shutdown and lease paths were
read, not assumed.

Lenses: **4 run / 4 returned** — AI-slop, logic holes, excess complexity, edge cases, all run
inline by this critic, none delegated. No lens returned empty. Note: gates.md holds no
plan-defense section; the "two open defense rows" are taken from the plan's own preamble
(D-15 = dependency failure, D-11/B17 = the sub-360p source) and verified as such.

### Facts established this round

- **ffmpeg 7.1.1, the plan's exact D-11 argv (`split` → `scale=w=W:h=H:force_original_aspect_ratio=decrease:force_divisible_by=2`, per-stream High/4.1/yuv420p/CRF 21 + maxrate, shared `agroup:aac`, fMP4, `-hls_fmp4_init_filename init.mp4`):**
  - 1920×1040 → `640x346`, `1280x694`, `1920x1040`, exit 0; the encoded streams probe as `High`, level `41`. **BLK-8's fix works.**
  - 1080×1920 portrait → `202x360`, `406x720`, `608x1080` — the exact numbers D-11 states.
  - **320×240 → `RESOLUTION=480x360`.** The plan's D-11 claim that a sub-360p source "is *not* enlarged — it is encoded at 320×240" is **false**: `force_original_aspect_ratio=decrease` fits the *target box* to the source's aspect, it never caps at the source. With `scale=w='min(640,iw)':h='min(360,ih)':force_original_aspect_ratio=decrease:force_divisible_by=2` the same source encodes at `320x240`.
  - Init segments follow `var_stream_map` order with the audio group first: `audio/init_0.mp4`, `360p/init_1.mp4`, `720p/init_2.mp4`, `1080p/init_3.mp4` — exactly what D-8 predicts; P04's "exactly one `init_*.mp4` per variant directory" holds.
  - The master carries `#EXT-X-MEDIA:TYPE=AUDIO,GROUP-ID="group_aac",…,URI="audio/index.m3u8"`, one `EXT-X-STREAM-INF` per rung with `BANDWIDTH`, `AVERAGE-BANDWIDTH`, `RESOLUTION`, `CODECS="avc1.640029,mp4a.40.2"`, `AUDIO="group_aac"`. `AVERAGE-BANDWIDTH` is present and below the ceiling (720p rung: 2 898 586 < 3 000 000). B27's shared-audio shape holds.
  - **The master carries no `FRAME-RATE` and no `EXT-X-INDEPENDENT-SEGMENTS`.** `EXT-X-INDEPENDENT-SEGMENTS` is written into each *video variant* playlist (`720p/index.m3u8`), not the master, and not the audio playlist. `FRAME-RATE` appears nowhere. B26 as written is red on the real muxer.
- **asynq@v0.26.0.** `Server.Shutdown` (`server.go:724-751`) calls `processor.shutdown()` (`processor.go:139-149`), which waits `ShutdownTimeout` (**10 s**, `media-worker/main.go:45`) with the handler's context **still live**; on expiry it closes `abort`, the worker goroutine `requeue`s the message to *pending* (`processor.go:236-239, 260-272`) and returns, and only the deferred `cancel()` (`processor.go:207-210`) then cancels the handler's context. `srv.wg.Wait()` does not include the handler goroutine; `main` then runs `dependencies.Close()` and exits. Lease expiry (`processor.go:244-246`) cancels the context and calls `handleFailedMessage`, whose `retry` is a no-op on an invalid lease (`processor.go:351-355`); re-delivery comes from the recoverer (`recoverer.go:85-103`, cutoff 30 s). `DefaultRetryDelayFunc` (`server.go:401-405`) yields ~16 s / ~31 s / ~96 s for retries 1–3.
- **docker-compose.yml.** `media-data` is mounted only in `media-worker` (`:88`, rw) and `nginx` (`:117`, ro). The `api` service (`:72-73`) is `read_only` with `tmpfs: [/tmp]` and **no media volume**. Worker `Concurrency: 1`, one replica.
- **authoring.go / courses.sql.** `Save` (`authoring.go:185-271`): `lockAuthoringCourse` (`courses.sql:75-89`, `FOR UPDATE` on `courses`) at `:193` → `LockEligibleAuthoringOwner` (`courses.sql:105-111`, `FOR SHARE` on users) at `:206` → module/lesson upserts `:216,:227` → `deleteMissingAuthoringLessons` `:237` → `updateAuthoringCourse` `:244`. As the plan cites.
- **spec.md:67** records Q2 as amended by the gate owner: container sniff at upload, codec check in the worker — **approved**. The plan still reads it as "proposed, awaiting the gate owner's acceptance" (D-8, §11.9, §12 NB-3).

### Part 1 — the eight edition-1 blocking findings

| # | Edition-1 finding | Closed at | Verdict |
|---|---|---|---|
| BLK-1 | lease/claim model; shared working dir | §1 T4–T8, D-10, D-7 C5/C6, §6 `claim_token` + CHECK, B24, R18, R32, R34 | **NOT CLOSED** — DB-level ABA closed; crash/restart recovery not (below) |
| BLK-2 | ownership port cannot share the mint transaction | D-1, §2 (`ports/ownership.go` gone), R28, M05 | **closed** — lock order verified deadlock-free (below) |
| BLK-3 | two creation points | §1 tus surface, D-2a, P03, §5.d, §2 `tus-client.ts`, M01/ADR 008 | **closed** |
| BLK-4 | PATCH protocol unspecified | D-2d steps 1–5, R29, B03, R11 | **closed** — airtight for the two cases asked (below); four residuals, all non-blocking |
| BLK-5 | interrupted upload locks the lesson for 24 h | D-13, §2 `cancelUpload`, D-12 predicate, R25, R30, P11 | **closed** |
| BLK-6 | B19 unobservable at smoke | D-14 pre-initialised label values, B19, B25 | **closed** |
| BLK-7 | enqueue failure strands a `queued` asset | D-15 commit-then-enqueue, T9, D-10 retry accepts `queued`, C6, R31, R17 | **closed** |
| BLK-8 | odd dimensions; `init_N.mp4` | D-11 `force_divisible_by=2`, R21, B21, D-8, ADR 009 | **closed and re-verified** (1920×1040 and portrait run clean) — but the same D-11 paragraph introduces a new false claim, N-1 |

**BLK-1 substance.** What edition 2 gets right: `Claim` writes a fresh token; publish/requeue/fail
are guarded on `(state='processing', claim_token=$token)`; the CHECK `claim_token IS NULL OR
state='processing'` prevents a token surviving a terminal write; `temporary/{id}-{claimToken}`
means two attempts never share scratch; T6 before returning a retryable error makes the
Asynq retry (same task id) claim from `queued`. B24 pins the ABA at the database. That half is
sound. What is not closed is the path BLK-1 was about — the attempt that could not write:

- **Graceful restart strands the asset deterministically.** `docker compose restart
  media-worker` mid-transcode: asynq keeps the context live for 10 s, ffmpeg keeps running,
  then the task is requeued to *pending* and the context is cancelled with the process
  already leaving (`server.go:751` → `main` exits, pool closed). T6 is never written; the row
  rests `processing` with a token. On restart the pending task is delivered within
  milliseconds; D-10 step 2 finds `processing`, returns `nil`, and asynq marks the task
  **done**. Recovery class C5 ran at start but requires `updated_at < now − (30 min + 10
  min)`, and the row is minutes old, so it did nothing. Operator retry refuses `processing`
  (R09). Nothing will ever touch the row again unless the worker is restarted *a second time,
  ≥ 40 minutes later*. The same sequence follows a hard crash (recoverer re-delivers ~30–60 s
  after restart, still inside the 40-minute blind spot) and the D-15 "PostgreSQL down on the
  final attempt" path (retries exhaust in ~2.5 min, the row is < 40 min old at the next
  start). Spec criterion 2 — every terminal state "reached from a real worker outcome" — is
  met only by an outcome that never arrives; edition 1's stranding has moved from `queued` to
  `processing`.
- **D-10's "worker shutdown cancels the handler's context" is not what asynq does**; it
  waits, then requeues, then cancels. R18 (cancel mid-job → `queued`) tests a fake; the real
  server never gives the handler that window.

```
[logic ×2 (logic, edge)] The recovery pass cannot see the row a restart just orphaned, and the redelivered task is consumed as success
де:      D-7 C5 ("updated_at older than taskTimeout + 10 min", "at worker start only"); D-10 step 2 ("Zero rows → nil (someone else has it)"); D-10 "Cancellation"; D-15 "PostgreSQL down when the terminal state must be written"
чому:    see BLK-1 substance above — every graceful restart or crash during a transcode leaves a `processing` row nobody owns, the next delivery returns nil and is archived as done, and C5's 40-minute threshold hides the row from the only pass that could free it. With `Concurrency: 1` and one replica (`main.go:45`, compose) there is no "someone else" at process start.
що зробити: (1) State the single-worker assumption and make C5 threshold-free at start: "every `processing` row at worker start is dead, because this process is the only claimant; set `queued`, clear the token" — trigger to reintroduce the age threshold: a second worker replica. Guarantee the order: `Recover` completes before `server.Start`. (2) In D-10 step 2, a `processing` row that is not `queued` returns a **retryable error**, not nil, so a delivery is never consumed while the row is owned by nobody; with the ~16–96 s backoff the lease-expiry case (old goroutine writes T6 within seconds) then claims on the next delivery. (3) Reword D-10 "Cancellation" to asynq's real order (wait `ShutdownTimeout`, requeue to pending, cancel) and raise `ShutdownTimeout` or accept that T6 is best-effort on shutdown — the correctness then rests on (1). Add an R row: "worker restarted mid-render → after restart the asset reads `queued` and is claimed again" (integration or e2e).
вага:    blocking
```

**BLK-2 substance — the lock order is deadlock-free, verified.** Mint: `FOR SHARE OF courses`
(join) → `MAX(generation)` → `INSERT video_assets` (which takes `FOR KEY SHARE` on the
referenced `course_lessons` row). Save: `FOR UPDATE` on `courses` first (`:193`), then lesson
upserts (`FOR NO KEY UPDATE`, compatible with `KEY SHARE`), then the lesson delete, then the
course update. Both transactions take the `courses` row before touching anything the other
locks, so they serialise; the FK `SET NULL` inside Save waits only on video_assets row locks
held by publish/PATCH transactions, neither of which takes a course lock, so no cycle runs
through `video_assets → courses`. `Transition`/`Delete` lock the same course row first.
Residual (non-blocking, N-6): two concurrent mints both hold `FOR SHARE` (compatible), both
reach `INSERT`, and the loser's `23505` may name `video_assets_generation` rather than the
pending index — D-1 step 4 maps "the pending index" only.

**BLK-4 substance.** Concurrent same-offset PATCH: both take `FOR UPDATE`; the second sees
`received_bytes` advanced and answers `409` at step 2 without writing — R29 holds. Crash
mid-chunk: any crash between step 4's fsync and the commit leaves the file longer; step 3
truncates it to the column — B03 holds; a crash *during* the append leaves a short partial
that is also longer than the column, same repair. Airtight for what was asked. Four residuals
are recorded under Part 2 (N-7, N-8, N-9, N-10), all non-blocking.

**D-15 substance.** Redis at enqueue: commit-then-enqueue, C6 and retry-accepts-`queued` —
closed (R31, R17). `expires_at` governs only the upload window: D-5's takeover and C3 name
`created`/`uploading` only, PATCH step 1 checks expiry, B16(b) pins that a `processing` row
past expiry is untouched — closed. PostgreSQL down: the deferral to C5 is the right idea but
inherits the BLK-1 threshold gap above; "down at `Claim` → the retry is a no-op" is imprecise
(it is an error, so asynq retries — the plan's own preferred outcome). Volume full/read-only:
the per-request `507`/`503` and the readiness probe are sound; **the mint-time gate is not
implementable where the plan places it** (N-2 below).

### Part 2 — findings on edition 2

```
[edge + logic + AI-slop] D-11's sub-360p rule upscales — the "verified" no-upscale claim is false, and R21 would pin the wrong argv
де:      D-11 "SelectLadder — the rule" bullet 2 ("a 320×240 source is *not* enlarged — it is encoded at 320×240"); §2 `domain/source.go` ("never upscaling"); §11 item 10; R21/B21 expected argv; the closure of defense row 4
чому:    Executed: the plan's scale expression on a 320×240 source produces `RESOLUTION=480x360` (verified, ffmpeg 7.1.1). `force_original_aspect_ratio=decrease` fits the 640×360 box to the source aspect; it never stops at the source. So the defense-row-4 closure ("never upscale" and "always one rendition" both true) rests on a wrong reading of the filter, and the exact-argv test the plan is proud of would lock it in. The sentence sits in the same "Verified:" voice as the genuinely verified odd-dimension fact — filler dressed as evidence.
що зробити: Every rung's expression becomes `scale=w='min(640,iw)':h='min(360,ih)':force_original_aspect_ratio=decrease:force_divisible_by=2` (verified: 320×240 stays `320x240`; 1920×1040 and portrait unchanged). Update R21/B21's expected list; add to B21 a 320×240 probe whose computed target is 320×240; rewrite the D-11 bullet as "verified at GATE 2 round 2" with the `min()` form. Thumbnail: `scale='min(640,iw)':-2` for the same reason.
вага:    blocking
```

```
[logic] The mint-time disk gate runs in a process that has no disk
де:      D-15 "`MediaStorage.Free(ctx)` … gates … the **mint**"; §2 `cmd/api/main.go` ("constructs the media application"); §5.a (lists the `upload` service's mount only); B23 (fake storage port, contract test)
чому:    The mint is `POST /api/v1/admin/lessons/{id}/video` in the **API** executable. `docker-compose.yml:72-73` gives `api` a tmpfs `/tmp` and no `media-data` mount; `statfs` has nothing to stat. Nothing in §2/§5 adds the mount, and §1 says the byte path never touches the API. B23 proves the gate only against a fake, so this surfaces at M09/M12 as a Compose change nobody scheduled — or the gate is silently dropped.
що зробити: Choose in the plan. Simplest (and reuse-first): drop the mint-time gate; the **first PATCH** in the upload service — which has the volume and knows `declared_bytes` — refuses with `507` when free space < 3 × declared, before any byte is written (the row then reads `created`, cancellable, not `failed`). Keep the claim-time gate. B23 moves to M10 as an integration row against a tempdir with a fake `Free`. Alternative: mount `media-data:ro` into `api` and say so in §5.a — a topology change the plan must own explicitly.
вага:    blocking
```

```
[logic] B26 is red on the real muxer: no FRAME-RATE, and EXT-X-INDEPENDENT-SEGMENTS lives in the variant playlists
де:      D-11 "Master playlist conformance" ("expected to emit all of them … this plan does not assume it"); B26; M15; the "contingency … regenerates master.m3u8"
чому:    Verified: ffmpeg 7.1.1 writes `CODECS`, `RESOLUTION`, `AVERAGE-BANDWIDTH` and the audio group into the master, writes `EXT-X-INDEPENDENT-SEGMENTS` into each *video variant* playlist (valid per RFC 8216 §4.3.5.1, absent from the master), and writes **no `FRAME-RATE` anywhere**. Apple 9.15 makes `FRAME-RATE` a MUST (`research-codecs.md:81`). So the contingency is not a contingency, it is the path — and it has no owner, no estimate and no row: M15 would go red and the writer (or the decision to drop the attribute) would be improvised in the last task.
що зробити: Decide now, one of: (a) ponytail: relax B26 to "`EXT-X-INDEPENDENT-SEGMENTS` in the master **or** every video variant playlist; `CODECS`, `RESOLUTION`, `AVERAGE-BANDWIDTH` on every variant" and record `FRAME-RATE` as an accepted Apple non-conformance with the trigger "a Phase 6 player that needs it" — no writer, no new code; or (b) own the writer: a fixed-format master writer over the rung table + probed frame rate lands in M08 (+0.75 h) with a unit row, and B26 keeps `FRAME-RATE`. Either way delete the sentence "expected to emit all of them".
вага:    blocking
```

```
[edge] A stale claim can still remove a newer claim's published directory
де:      D-8 "destination exists and the asset is not ready → the adapter removes the stale directory and re-publishes"; D-10 step 10 ("Publish and evict in one transaction")
чому:    The DB guard is airtight; the filesystem branch is not under it. Stale goroutine A (lease expired, recovery re-queued, B now claimed) finishes its render and reaches publish while B is between its rename and its commit: A's rename fails (non-empty destination), A reads the row (`processing`, not ready), removes B's directory and renames its own in, then A's guarded UPDATE affects zero rows and A returns nil. B commits `ready` over A's bytes — same source, same argv, so not corrupt — but a crash of A between its `RemoveAll` and its rename leaves `ready` with no bytes. Double fault, millisecond window, single-worker deployment makes it rarer still; stated because "publication is atomic" is a headline claim.
що зробити: Order step 10 explicitly: lock the row `FOR UPDATE` **and** verify `claim_token = $token` first; only then remove a stale destination and rename; the UPDATE and commit follow in the same transaction. A goroutine whose claim is lost then touches nothing under `hls/`. One sentence in D-8 and D-10.
вага:    non-blocking
```

```
[edge] Two concurrent mints fail on whichever unique index Postgres checks first
де:      D-1 step 4 ("a `23505` on the pending index → `upload_in_progress`"); D-5 bullet 1; §6 indexes 2 and 3
чому:    `FOR SHARE` is compatible with `FOR SHARE`, so two minters for one lesson both compute `MAX+1`, both `INSERT`; the loser blocks on the winner's uncommitted index entries and errors on the first index Postgres checks — which may be `video_assets_generation`, not `video_assets_pending_per_lesson` (and *must* be, when the existing rows are all `ready`). A mapper keyed on "the pending index" returns `500 internal_error` for a plain concurrent mint.
що зробити: Map `23505` by `ConstraintName` (the 4a `slugTaken` shape, `authoring.go:426-429`) and map **both** index names to `upload_in_progress`; add the second name to R07 or a new M05 row with two concurrent mints.
вага:    non-blocking
```

```
[logic] The sniff failure is `created → failed`, a transition the matrix does not have
де:      §1 T2 (`created → uploading` on first PATCH, guard "sniff passes"); T8 (`processing`/`uploading → failed`); D-8 "Sniff on the first chunk"; B18
чому:    The first PATCH finds the row in `created`; a failed sniff writes `failed` from `created`, not from `uploading`. M03's transition table, implemented literally from §1, refuses it and B18 fails at M10.
що зробити: T8 becomes `created`/`uploading`/`processing → failed`; one word in the matrix and in `domain/asset.go`'s table test.
вага:    non-blocking
```

```
[edge] Retrying a sniff-failed asset enqueues a job for a source that was never written
де:      D-10 "Retry" (accepts any `failed`); D-8 (sniff failure "writes nothing under `source/`"); R08
чому:    A `failed`/`unsupported_source` row from the upload service has `received_bytes = 0` and no `source/{id}/original`. Retry sets `queued`, the worker probes a missing file, fails again with a different reason — a round trip that can only end in `failed`.
що зробити: `Retry` requires `received_bytes = declared_bytes` (the source landed); otherwise `409 video_not_retryable`. One clause in the domain rule and one R row.
вага:    non-blocking
```

```
[edge] A crash between the completion rename and its commit kills the ticket with a permanent 503
де:      D-2d "Completion is the same transaction's terminal branch: … renamed … **before** the commit"; step 3 ("shorter is impossible … → `503`")
чому:    Rename lands, process dies, commit never happens: the row is `uploading` with `received_bytes = declared − lastChunk` and the scratch file is gone. Every later PATCH hits step 3 with a missing scratch → `503` forever; the client's only exit is `DELETE` and a fresh 2 GiB upload. Tiny window, but criterion 1 says the source is never lost, and the bytes are sitting in `source/`.
що зробити: Make step 3 idempotent for this case: "scratch missing **and** `source/{id}/original` exists with `declared_bytes` bytes → skip the append, run the terminal UPDATE to `queued`". One line in D-2d and an M10 row.
вага:    non-blocking
```

```
[edge] Nginx's body cap is Content-Length-based; a chunked PATCH streams past it
де:      §5.c ("nginx enforces it from `Content-Length` before proxying even with buffering off"); §8 tus vocabulary (`413` only "for a chunk past the declared length"); B02, R12
чому:    With `proxy_request_buffering off`, a `Transfer-Encoding: chunked` request has no `Content-Length` for nginx to check; the body streams to the upload service. The browser client always sends a length, so the happy path is unaffected — but the upload service is the last line and the plan gives it no chunk cap of its own.
що зробити: The tus server wraps the body in `http.MaxBytesReader(w, r.Body, chunkBytes)` and answers `413` on overflow, independent of nginx; one integration row with a 9 MiB chunked body.
вага:    non-blocking
```

```
[edge] A stranded `queued` row polls forever and offers no retry in the UI
де:      D-12 predicate (`queued` keeps polling); §2 `LessonVideoField.tsx` ("the failure reason, retry"); D-7 C6 ("the operator's retry button is the path")
чому:    The API accepts a `queued` retry (R08) and the plan calls it the operator's exit, but the field's retry control is described beside the failure reason only. A `queued`-with-no-task row keeps every open editor polling every 3 s with nothing to press.
що зробити: The field renders the retry control for `failed` and for `queued`/`processing` rows older than N minutes (`updatedAt` is in the read); add it to R25 or a new M14 row.
вага:    non-blocking
```

```
[edge] `-level:v 4.1` on a 1080p60 source declares a level the stream exceeds
де:      D-11 "H.264 High Profile, Level ≤ 4.1"; B28
чому:    Level 4.1 caps 1080p at ~30 fps (245 760 MB/s). libx264 warns and encodes anyway, so a 60 fps screen recording yields a stream labelled 4.1 that violates 4.1, which is worse for a strict decoder than an honest 4.2. Apple 1.3a asks that *some* variants be ≤ 4.1, not all.
що зробити: Either drop `-level` on the 1080p rung (the 360p/720p rungs satisfy 1.3a) or add a constant `fps=fps=min(30,source_fps)`-style cap — the former is one flag fewer. Note the choice in D-11.
вага:    non-blocking
```

```
[logic] B27 needs a two-rung source; §5.f generates one clip
де:      B27 precondition ("a source that selects two rungs"); §5.f (only the 640×360 clip); M15
чому:    The only fixture selects `360p` alone, so the shared-audio row has nothing to run against; M15 would add a second fixture on the day.
що зробити: §5.f generates a second clip at 1280×720 (same command, `size=1280x720`) and `media.setup.ts` copies both; B27 names it.
вага:    non-blocking
```

```
[AI-slop] Stale and garbled prose that a reader will trip on
де:      D-8 last paragraph, §11 item 9, §12 NB-3 ("proposed, awaiting the gate owner's acceptance"); D-11 bullet 2 ("decided rather than left to collide with the schema" around an unverified claim — see N-1); §10 total ("four tasks at 2 h … plus M04/M05/M09/M10 at 2 h")
чому:    spec.md:67 records the Q2 amendment as the gate owner's decision on 2026-09-06 — the plan reads as unsettled on a settled point. The hours parenthetical double-counts in words (five tasks are at 2 h: M04, M05, M09, M10, M14); the 25.25 h total is arithmetically right, which makes the wrong sentence beside it more distracting, not less.
що зробити: D-8/§11/§12: "accepted by the gate owner (spec Q2, amended 2026-09-06)". §10: "five at 2 h, three at 1.75 h, four at 1.5 h, two at 1 h, one at 0.5 h".
вага:    non-blocking
```

```
[complexity] Three sweeper indexes on a table bounded to a few rows per lesson
де:      §6 indexes 5, 6 and `video_assets_by_state`
чому:    Two live generations per lesson plus a bounded tail of `failed`/evicted rows is hundreds of rows, not millions; the sweeper takes 50 rows in no order once per job. Three partial/composite indexes exist to make six predicates cheap that a sequential scan already makes cheap, and each is one more thing the down migration and the canary test carry. The edition-1 critic offered "state the table is small enough to scan" as the alternative; the plan took the heavier one without saying why.
що зробити: Keep the two unique indexes (correctness) and `video_assets_by_lesson` (the hot read); drop 5, 6 and `by_state`, and write one sentence: "the sweep and recovery predicates scan; trigger to index them: a table over ~10⁵ rows".
вага:    non-blocking
```

```
[complexity] Two disk gates and a playlist writer where one gate and no writer would do
де:      D-15 (`Free` at mint **and** at claim); D-11 "Contingency … regenerates `master.m3u8`"
чому:    The mint-time gate is what forces the API-mount question (N-2); the first-PATCH gate covers the same case with the volume already in hand, so one gate site is both simpler and correct. The master writer is a second producer of the playlist a consumer must trust — if N-3 resolves as (a), it never exists; if (b), it should be scoped to *adding* `FRAME-RATE` to ffmpeg's master rather than regenerating the whole file, so `AVERAGE-BANDWIDTH` (which ffmpeg measures from the real segments) is not replaced by a ceiling.
що зробити: Fold into the N-2 and N-3 edits: one gate at first PATCH plus the claim-time gate; writer only if `FRAME-RATE` is kept, and then as a line-append over the measured master, not a rewrite.
вага:    non-blocking
```

### Four-lens accounting

- **AI-slop** — 2 findings (the false "verified" no-upscale claim, folded into N-1; the stale Q2 wording and the garbled hours). Otherwise clean: §12's disposition table, the reuse table and the placement table are all load-bearing, and the edition-1 filler ("five years late", the D-2b self-congratulation) is gone as claimed.
- **Logic holes** — 5 (N-1 recovery timing/shutdown order; N-2 the gate with no disk; N-3 B26 red on the real muxer; the `created → failed` transition; B27's missing fixture). BLK-2's lock order and BLK-7's ordering were checked and hold.
- **Excess complexity** — 2 (three sweeper indexes; two gates plus a playlist writer). Checked and accepted as earning their keep: the four ports (each a real dependency), the third executable (Q1 binding), the claim token (the cheapest thing that closes the ABA), `attempt` (still the only salt for a manual retry's task id), the `expiration` extension, and the six-class sweeper (bounded, one function).
- **Edge cases** — 9 (graceful restart, stale-claim destination race, the second unique index, sniff-failed retry, rename-before-commit crash, chunked body, stranded `queued` in the UI, level 4.1 at 60 fps, plus the executed 320×240 case). Checked and clean: two concurrent PATCHes, crash mid-chunk, the expired-ticket takeover against a live PATCH (its step-5 `expires_at > $clock` guard loses correctly), the sweeper unlinking an open scratch file, DELETE racing PATCH (row lock), `lesson_id IS NULL` at every read and at claim, `ErrTaskIDConflict` on retry-a-`queued`, ENOENT in purge.

Cross-lens duplicates collapsed with the lens count kept in the tag.

### Verdict

- Edition-1 blocking: **7 of 8 closed** (BLK-2, 3, 4, 5, 6, 7, 8). **BLK-1 not closed** — the
  ABA half is closed, the crash/restart half is not, and it is deterministic on a graceful
  restart, not a race.
- **New blocking: 3** — N-1 (the sub-360p rule upscales; exact-argv test would pin it),
  N-2 (the mint-time disk gate has no disk), N-3 (B26 is red on ffmpeg 7.1.1; the contingency
  is the path and has no owner). Each is a paragraph-sized edit; none changes the shape of the
  slice, the task cut, or the schema beyond one index decision.
- Non-blocking: **11**.
- Q2: settled by the gate owner; the plan's three "awaiting acceptance" phrases are stale
  wording only.
- **RED** — edition 3 required for the four blocking edits (BLK-1 residual, N-1, N-2, N-3);
  everything else in edition 2 stands.

## GATE 2 round 3 — plan-defense (sonnet, fresh context)

Target: `docs/media/plan.md` **edition 3** (2026-09-06, 1732 lines). Fresh five-question
plan-defense per `plan-defense/SKILL.md`, weighted per instruction toward byte handling and
process spawning: hostile input, dependency failure, race/replay, empty/boundary, rollback.
Every location the plan names below was opened and read; every ffmpeg claim was re-executed
independently (not trusted from the plan's own "verified" prose, given edition 2's disproved
sub-360p claim); the load-bearing code citations for D-1/D-7/D-10/D-15 were re-opened against
the actual repository and the vendored `asynq@v0.26.0` source, not assumed from prior rounds.

### Independent verification this round

- **ffmpeg 7.1.1** (same version the plan cites), executed fresh, not reused from earlier
  rounds' output:
  - `scale=w='min(640,iw)':h='min(360,ih)':force_original_aspect_ratio=decrease:force_divisible_by=2`
    on a generated 320×240 source: **stays `320x240`** (single-frame filter test **and** a full
    real `libx264`/`aac`/fMP4-HLS render with the plan's exact per-rung flags — `master.m3u8`
    reads `RESOLUTION=320x240`, `360p/init_1.mp4` and `audio/init_0.mp4` exist, one `.m4s`
    segment each, exit 0). Edition 2's disproved claim does **not** recur; edition 3's
    corrected `min()` expression is genuinely correct, not merely re-asserted.
  - Same filter on 1920×1040: `640x346`. On 1080×1920 portrait: `202x360`. Both match the
    plan's D-11 table exactly.
  - Thumbnail filter `scale='min(640,iw)':-2` on 320×240: stays `320x240` (no upscale).
  - A 2-second source (shorter than the 6 s `-hls_time` target) still produces a valid one-segment
    HLS output — the `min()` render above used a 2 s clip and exited 0 with one `.m4s` per
    variant, matching P04's "at least one `.m4s` segment" acceptance criterion, not merely an
    assumption about ffmpeg's behaviour on a short source.
- **`docker-compose.yml`**: `api` (`:60-73`) is `read_only: true`, `tmpfs: [/tmp]`, no
  `media-data` mount — confirmed, this is what makes D-15's gate-placement decision (N-2's
  fix) correct rather than assumed. `media-worker` (`:74-...`) mounts `media-data:/media`; no
  `deploy.replicas` anywhere in the file, consistent with (but not enforcing) the plan's stated
  single-worker assumption.
- **`backend/cmd/media-worker/main.go:45`**: `Concurrency: 1, Queues: {"media": 1},
  ShutdownTimeout: 10 * time.Second` — matches D-7's threshold-free C5 rationale and D-10's
  "10 s" shutdown-window claim exactly.
- **`asynq@v0.26.0`**, read directly (not taken on the plan's word): `server.go:724-751`
  (`Shutdown` → `processor.shutdown()` before `wg.Wait()`); `processor.go:139-149`
  (`shutdown()` calls `stop()` then `time.AfterFunc(shutdownTimeout, close(abort))` and blocks
  on the semaphore — the handler's context is untouched during the wait); `processor.go:207-210`
  (`defer cancel()` registered before the `select`); `processor.go:236-239`
  (`<-p.abort` branch: `requeue(lease, msg)` then `return` — the deferred `cancel()` fires only
  after the return). This confirms, line for line, the plan's "waits, then requeues to pending,
  only then cancels" sequence in D-10 "Cancellation" and the T6/T7 split it rests on.
  `server.go:401-405` (`DefaultRetryDelayFunc`, exponential-ish backoff, consistent with the
  plan's "~16 s/~31 s/~96 s" claim as order-of-magnitude, not exact). `client.go:225-228`
  (`ErrTaskIDConflict`) and `processor.go:328-329` (`SkipRetry`) both exist exactly as cited.
- **`backend/sql/queries/courses.sql:89,111`** and **`authoring.go:185-237`**: `lockAuthoringCourse`
  (`FOR UPDATE`) at `:193`, `actor.CanManage` at `:197`, lesson upserts, `deleteMissingAuthoringLessons`
  at `:237` — the exact sequence D-1 relies on for its deadlock-free lock-order argument.
- **`.gitignore:14`** is `/media/`; `git check-ignore -v` on `docs/media/{plan,gates,spec}.md`
  matches nothing — §0's "found and fixed in `5ef6c1f`" claim holds today.
- `backend/migrations/` holds only `00001`–`00004`; no `internal/media` Go code exists yet —
  edition 3 is still pre-implementation, as it claims to be.

### The five questions

| # | Question | Closed by (plan.md location) | Verdict |
|---|---|---|---|
| 1 | Hostile input — malformed video (`D-8` first-chunk sniff + `D-10` step 4 `Probe`/T8 codec check, R15/B13); hostile filename (`D-11` "paths derived from the asset UUID and the claim token … never from a client string … reaches neither argv nor the filesystem", proven by injection in R21); forged/replayed ticket (`D-2b` — ownership-at-mint, not id-unguessability, is the protecting property; `D-2d` step 1 — missing/evicted/expired/wrong-state → `404`, no write; R10); overlapping chunk offsets (`D-2d` step 2 — `Upload-Offset` must equal `received_bytes` exactly or `409`, no write, covering both "behind" and "ahead" replays; R29/R11); no user-controlled ffmpeg argv (`D-11` "the argv is built from three sources and no others"; R21's element-by-element assertion with a shell-metacharacter filename, independently checked against `D-11`'s render argv — no string interpolation point exists for `original_name`) | **closed** |
| 2 | Dependency failure — Redis at enqueue (`D-15` ¶1, commit-then-enqueue, `503`, retry-accepts-`queued`, recovery C6); PostgreSQL mid-transcode (`D-15` ¶2, nothing written, Asynq retries, row rests `processing` on exhaustion, C5 requeues on next start — **re-verified against real `asynq` source above**); media volume full/read-only (`D-15` ¶3, `Free()` gate at first `PATCH` and at claim, per-write `507`/`503` fallback, readiness write-probe); ffmpeg killed (folds into the generic "retryable failure" path of `D-10` step 12 / R14 / R33 — `exec.CommandContext`'s `Wait()` returning a signal-killed error is not distinguished from any other render failure, which is sufficient but not named as its own scenario); upload service down while the API is up (named concretely under `§7` Rollback rather than under `D-15`: "nginx answers `502` for a path the UI no longer offers"; rows rest `created`/`uploading` and resume via `HEAD` on recovery — same mechanism, different heading, but a client-observable answer exists) | **closed**, two residuals (below) |
| 3 | Race and replay — two uploads for one lesson (`D-5` pending-per-lesson index, `D-1` step 4, R07/R38); Asynq re-delivery against an already-`ready` asset (`D-10` step 1 "`ready`/evicted → `nil`", R16/R36); publish racing eviction (`D-8` "the claim is verified under the row lock … before anything under `hls/` is touched", with the stale-claim double-fault scenario explicitly narrated and closed by that ordering; `D-5`'s two-generation eviction happens inside the same publish transaction, not as a separate race); lesson deleted mid-transcode (`D-4` `SET NULL`, `D-10` step 1 "`lesson_id IS NULL` → `nil`", B14); worker restart mid-render (`D-7` C5 threshold-free + `D-10` step 2 retryable-not-`nil` + R35 e2e — **the asynq shutdown sequence this rests on was re-read from source above and matches line for line**) | **closed** |
| 4 | Empty and boundary data — zero-byte upload (B04 + schema `CHECK (declared_bytes BETWEEN 1 AND 2147483648)`); source shorter than one segment (§5.f's ~2 s fixtures against a 6 s `hls_time` target + P04's "at least one `.m4s` segment", **independently re-executed above with a real 2 s clip, exit 0, one segment**); source below the lowest rendition (`D-11` `SelectLadder`/`min()` expression, B17/B21/B30 — **this is the row edition 2 closed on a false claim; independently re-executed above at a fresh terminal with both a scale-only test and a full real HLS render, and the corrected filter genuinely holds: 320×240 stays 320×240, not 480×360**); exactly 2 GiB (B01, `CHECK` bound at `2147483648`, and the chunk math `8 MiB × 256 = 2147483648` checks out exactly); third upload evicting the oldest (`D-5` "at most two `ready` generations… evicts the lowest generation", B06) | **closed, and independently re-verified by execution rather than trusted** |
| 5 | Rollback — supported path named (`§7`: redeploy previous images, leave migration 00005 in place, no `course_lessons` coupling); in-flight jobs (unknown task type archived by the empty-mux behaviour foundation already chose, nothing falsely marked `ready`); already-written media named as inert-but-unswept during rollback, an explicit disk-only cost; in-flight uploads (abandoned at their offset, resumable via `HEAD` on roll-forward, reclaimed by sweeper class C3 past `expires_at`); the one explicit irreversible-if-misused action named and forbidden ("do **not** `goose down 00005`", with the reason); frontend-alone rollback stated as not a security boundary; nginx/upload-service coupling stated (`502` if mismatched) | **closed** |

### Residuals (do not reopen a row, recorded for the record)

```
[edge] "down at Claim → the retry is a no-op" mischaracterises a live connection failure as an ABA no-op
де:      D-15 ¶2, the sentence "if PostgreSQL is down at `Claim`, nothing is claimed and the
         retry is a no-op"
чому:    This exact phrase was carried over unchanged from edition 2; GATE 2 round 2's D-15
         substance narrative already called it imprecise ("it is an error, so asynq retries —
         the plan's own preferred outcome"), but it was never turned into a numbered finding
         and edition 3 did not fix the sentence. A `Claim` UPDATE failing because PostgreSQL is
         unreachable returns a connection error, not zero affected rows — the handler should
         treat it as a retryable error (Asynq backs off and redelivers), not as the "someone
         else has it" ABA case D-10 step 2 describes elsewhere. The mechanism the plan actually
         wants (Asynq retries, C5 mops up on the next restart) is right; only the sentence
         describing *why* is wrong, and a reader implementing D-15 literally could code a silent
         swallow instead of a returned error.
що зробити: One clause: "…nothing is claimed; the call returns an error, the handler returns it
         unwrapped, and Asynq retries with backoff — this is a dependency error, not the
         zero-rows ABA case of step 2."
вага:    non-blocking
```

```
[edge] "ffmpeg killed" (OOM, external SIGKILL outside our own cancellation) has no named acceptance row distinct from a generic transient failure
де:      D-10 step 12 ("On a retryable failure…"); R14/R33 ("a transcoder failing transiently…
         retry count at the maximum" / "a render returning `ENOSPC`")
чому:    The plan is explicit and well-verified about the process death *we* cause (D-10
         "Cancellation", against real asynq shutdown). A process death caused by something else
         — the OOM killer, an operator's `docker kill -9` on the child, a host restart — reaches
         the same code path (`exec.CommandContext.Wait()` returns a signal-killed error) but is
         never named as its own row the way cancellation, disk exhaustion and malformed sources
         each are. It is covered by construction (any non-`SkipRetry` render error is a
         retryable failure under R14's fake), just not by a scenario a reader can point to.
що зробити: One R-row fake: a transcoder returning a signal-killed-style error → same T6/Fail
         behaviour as R14, so the coverage is stated rather than inferred.
вага:    non-blocking
```

```
[complexity] §12's NB-12 disposition contradicts §6 and §13's N-15 — a stale row in the plan's own bookkeeping
де:      §12 NB-12 ("the reclaim index is on the predicate with an `id` key; a separate
         `expires_at` partial index serves class C3; `(state, updated_at)` serves C4–C6");
         §6 step "No index serves the sweeper, deliberately"; §13 N-15 ("the three scan indexes
         are dropped")
чому:    Not one of the five defense questions, but worth recording: §12 is edition 2's
         disposition table, carried forward unedited, and it describes three sweeper indexes
         that §6 and §13 (both edition-3 sections) explicitly say were removed. A reader who
         trusts §12 over §6 would add indexes the schema in §6 does not have. This is a
         plan-critique-shaped finding (internal contradiction), not a defense hole — none of
         the five questions depends on which indexes exist — so it is not scored above, but it
         sits exactly where a future implementer would get confused.
що зробити: Rewrite §12 NB-12's disposition to match §6/§13: "the reclaim/recovery predicates
         scan; no dedicated index; trigger to add one: ~10⁵ rows."
вага:    non-blocking (bookkeeping only; does not touch any of the five questions)
```

### Accounting

- Five questions run: hostile input, dependency failure, race/replay, empty/boundary, rollback.
- **Holes: 0.** Every row is answered by name in the plan text, not by this examiner's own
  reasoning; every ffmpeg-behaviour claim load-bearing to a row (D-11's `min()` fix, the
  short-source case) was re-executed fresh in this session, independently of the plan's own
  "verified" prose and independently of the prior rounds' output — given that edition 2's
  parallel claim in the same location was false, this was the one place a rubber-stamp would
  have been negligent. It held.
- Three residuals recorded, all non-blocking, none reopening a row: one stale sentence in D-15
  that describes a real dependency failure as the wrong kind of no-op, one unnamed-but-covered
  scenario (external process kill), and one stale cross-reference in §12 versus §6/§13.
- **Verdict: GREEN.**

## GATE 2 round 2 — plan-defense (sonnet, fresh context)

Target: `docs/media/plan.md` **edition 2** (2026-09-06, 1471 lines), as instructed. Note on
timing: the working-tree file kept growing while this review was in progress (another process
was folding a concurrent plan-critique round into a new edition) and had reached "Edition 3"
by the time this section was written. This review is against the edition-2 text captured in
full at the start of the session — line numbers below are edition 2's; they no longer match
the file on disk. Read against `docs/media/spec.md`, `work/PRODUCT-BRIEF.md`, the plan's own
cited code, and the plan-defense skill's five fixed questions, weighted for byte-handling and
process-spawning (hostile input incl. no user-controlled ffmpeg argv; dependency failure;
race/replay incl. asynq re-delivery; empty/boundary; rollback), scrutinising rows 2 and 4
hardest per instruction.

**Executed, not read** (spot-checks of the author's claims): the literal D-11 scale expression
was run against a locally-generated 320×240 `testsrc2` clip with ffmpeg 5.1.9 (Homebrew, ARM64
— a different build than the round-1/round-2-critique's 7.1.1, same libx264/scale filter
semantics); the literal `CHECK (state <> 'ready' OR array_length(renditions, 1) >= 1)` from
§6 was run against a live PostgreSQL 17 (`courses-postgres-1`, the repo's own Compose
database) with an `INSERT` carrying the column's own default value. Code citations
(`asynq@v0.26.0/processor.go:245`, `client.go:228,416`, `context.go:25,33`; `courses.sql:89`;
`authoring.go:237`; `httpserver/server.go:124`; `docker-compose.yml:88,117,49-69`;
`nginx.conf:20,44-45`; `.gitignore:14`) were opened and match the plan's text exactly.

**Cross-check note.** `docs/media/gates.md`'s own "GATE 2 round 2 — plan-critique
verification (fable)" section, appended concurrently by another reviewer against the same
edition-2 text, independently found and executable-verified two of the four holes below
(N-1, the 320×240 upscale; N-2, the mint-time disk gate; and BLK-1's residual, the
graceful-restart stranding) from a different angle (logic/edge lenses, not the five fixed
defense questions). That is agreement from an independent method, not this review copying it
— the ffmpeg run below was executed separately, with a different ffmpeg build, before this
section was written. The CHECK-constraint defect under row 4 does not appear anywhere in that
section and is this review's own find.

### The five-question table

| # | Question | Closed by (edition 2) | Verdict |
|---|---|---|---|
| 1 | hostile/malformed input, no user-controlled ffmpeg argv | D-11 point 2 + R21 (injected `original_name` proven absent from argv, element-by-element); D-8 first-chunk sniff + B18 (renamed text file → `failed`, nothing under `source/`); B13 (probe refusal); B01/B04 (2 GiB and zero-length bounds); `content_type` relaxed to non-refusing metadata (§6) | **closed**, one minor residual noted below |
| 2 | dependency failure (Redis, PostgreSQL, disk) | D-15 (commit-then-enqueue + C6 + retry-accepts-`queued`; PostgreSQL-down deferred to C5; `MediaStorage.Free` gates at mint and claim; write-probe readiness) | **open** — two defects, both below |
| 3 | race / replay, incl. Asynq re-delivery | D-10 `claim_token` (T4–T7), B24 (claim-token ABA), R29 (concurrent PATCH), R17 (`ErrTaskIDConflict`) | **open** — one defect below |
| 4 | empty and boundary data | D-11 `SelectLadder` (360p floor rung, never-upscale) + §6 CHECK (`ready ⇒ renditions ≥ 1`) + B17 | **open** — two defects, both executed, below |
| 5 | rollback | §7 (application rollback, in-flight jobs, already-written media, in-flight uploads, explicit "do not `goose down`", frontend-alone-is-inert, nginx-must-accompany-service) | **closed** |

### Row 2 — dependency failure — still open

**(a) The mint-time disk gate runs in a process the plan never gives a disk.** D-15
(line 879, `**MediaStorage.Free(ctx) (int64, error)** (one statfs) gates two places: the
mint refuses … and the claim refuses …`) requires `MediaStorage.Free()` — a `statfs` on the
media root — to run inside the **API** process at `T1` (line 124: `ownership … no other
non-terminal asset … free space (D-15)`). Confirmed independently: `docker-compose.yml`'s
`api` service (lines 49–69) has no `media-data` volume anywhere, unlike `media-worker`
(`:88`) and `nginx` (`:117`); §5.a and §2's `cmd/api/main.go` row (edition 2) name no mount
and no `MediaStorage` wiring for the API. As specified, either the mint's disk gate can never
succeed (the API container has nothing to `statfs`, so every mint would either error or the
gate silently never fires), or an implementer "fixes" it by mounting `media-data` into `api`,
which contradicts §1's own claim that "the byte path never touches the Go API" and §3's
listed reason the API needs no storage port. This is not a rare race: it fires on the first
real mint in the topology the plan itself defines.
що зробити: move the free-space precondition off the mint and onto the upload service's
first `PATCH` (T2), which already holds the mount and already knows `declared_bytes`; refuse
with `507` before the first byte is written, leaving the row `created` (cancellable). Keep the
claim-time gate as specified. State this explicitly in D-15 and in the `cmd/api/main.go` /
`cmd/upload/main.go` placement rows.

**(b) A crash between the completion rename and its commit stalls the ticket forever, not
just briefly.** D-2d (line 391, `the source is renamed … **before** the commit that writes
queued`) performs a non-transactional filesystem `rename` ahead of the DB commit that is
supposed to accompany it, inside "the same transaction's terminal branch." D-2d's step 3
(line 372) only names two cases for the scratch file's length — "longer" (truncate) and
"shorter" (impossible, `503`) — and has no branch for "**absent**, because a previous attempt
already renamed it away and then failed to commit." If PostgreSQL becomes unreachable (or the
process dies) exactly between the rename and the commit ACK, the row is left `uploading` with
`received_bytes < declared_bytes` while the file already sits at
`source/{videoId}/original`; every subsequent `PATCH` hits the "missing scratch" case D-2d
never wrote and answers `503` forever — the client's only exit is `DELETE` and re-uploading up
to 2 GiB that had already fully arrived, which is exactly the outcome spec criterion 1 ("the
source is never lost") is meant to prevent. The identical crash shape is handled explicitly
for the worker's publish path (D-8, line 564: "destination exists and the asset is not
`ready` → … removes the stale directory and re-publishes") but not here — asymmetric
treatment of the same primitive.
що зробити: give step 3 a third branch: scratch missing **and** `source/{id}/original`
already exists at exactly `declared_bytes` → skip the append, run the terminal `UPDATE` to
`queued` (or `queued`-and-enqueue if not yet done) instead of faulting.

### Row 3 — race / replay, incl. Asynq re-delivery — still open

**A graceful worker restart mid-transcode stray a `processing` row that the very next
delivery consumes as silent success.** Read `hibiken/asynq@v0.26.0` myself
(`processor.go:127-149` `shutdown()`, `:220-253` the worker goroutine): `Server.Shutdown`
starts a `shutdownTimeout` timer (10 s, `media-worker/main.go`'s configured value) with the
handler's `ctx` **still live**; on expiry `p.abort` closes, the goroutine's `select` takes the
`<-p.abort` branch, calls `p.requeue(lease, msg)` (pushing the message back to *pending*, i.e.
immediately redeliverable) and returns — and only *then*, in the goroutine's deferred cleanup,
does `cancel()` fire and actually cancel the handler's `ctx`. So on an ordinary
`docker compose restart media-worker` (or any deploy) during a render: the message is already
back in the queue before the render's context is even told to stop, D-10's step-12
"on context cancellation: Requeue" never gets a chance to run (the process is exiting), and the
row is left `processing` with a live `claim_token`. The restarted worker (or a second replica)
picks up the redelivered message within milliseconds; D-10 step 2 (`Claim`, guarded on
`state='queued'`) finds the row already `processing`, returns `nil`, and Asynq marks that
delivery **done** — not retried. Recovery class C5 (`updated_at` older than
`taskTimeout(30 min) + 10 min`) does not fire because the row is seconds old. `POST
…/retry` refuses `processing` (R09). Nothing touches the row again until a **second** restart
strictly after the 40-minute threshold, or a manual `UPDATE`. This is deterministic on the most
routine operational event a worker has (a deploy), not a rare interleaving, and it directly
defeats T6/T7/claim_token — the exact mechanism D-10 built to answer this row.
що зробити: (1) make the recovery pass threshold-free at worker start under the plan's own
`Concurrency: 1`, single-replica assumption — "every `processing` row at start is dead, this
process is the only claimant" — with the age threshold reintroduced only as a named trigger
for a second replica; (2) in D-10 step 2, a `processing` row found where a `queued` one was
expected must return a **retryable error**, never `nil`, so Asynq's own backoff (~16/31/96 s)
gives the still-finishing goroutine (or the next restart) a chance to write `queued` before the
delivery is discarded as done; (3) correct D-10's "worker shutdown cancels the handler's
context" to the sequence verified above (wait, requeue-to-pending, *then* cancel), and add an
acceptance row that restarts the worker mid-render and checks the asset reads `queued` and is
reclaimed, rather than asserting only the fake-port cancellation path (R18).

### Row 4 — empty and boundary — still open, both executed

**(a) The declared "never upscale a sub-360p source" fix does the opposite — executed and
confirmed.** D-11 (line 730, "`320×240` source is *not* enlarged — it is encoded at
`320×240`") is the plan's own closure of this row. Ran it: generated a 320×240 `testsrc2`
clip locally and applied the plan's literal expression,
`scale=w=640:h=360:force_original_aspect_ratio=decrease:force_divisible_by=2` (ffmpeg 5.1.9,
libx264) — output measured **480×360**, not 320×240. `force_original_aspect_ratio=decrease`
fits the *640×360 box* to the source's aspect ratio; it does not cap the result at the
source's own size, so a source smaller than the box is scaled *up* to the box's diagonal —
exactly the upscaling D-5/D-11 say must never happen. Re-ran with
`scale=w='min(640,iw)':h='min(360,ih)':force_original_aspect_ratio=decrease:force_divisible_by=2`
and got **320×240**, confirming the fix. None of edition 2's own acceptance rows would have
caught this: B17 (line 1281) asserts only which *named rungs* are selected (`[360p]` for a
320×240 probe), not the output's actual pixel dimensions; B21/R21 (lines 1285, 1318) assert
the argv text and `force_divisible_by=2`'s presence, not what ffmpeg does with that text. The
defect would have shipped through the entire acceptance suite silently.
що зробити: use the `min()` form in every rung's scale expression and in the thumbnail
filter; add an end-to-end row that executes a 320×240 clip and asserts the produced
`RESOLUTION` equals `320x240`, not just that `360p` was the selected label.

**(b) The schema CHECK that is supposed to make "`ready` ⇒ ≥ 1 rendition" structural does not
enforce it — executed and confirmed.** §6 (line 1117):
```sql
CHECK (state <> 'ready' OR array_length(renditions, 1) >= 1)
```
with `renditions text[] NOT NULL DEFAULT '{}'`. The plan calls this CHECK "load-bearing" and
the reason the row-4 closure needs no CHECK-weakening (line 1124, "so 'ready' cannot be
advertised for an empty publish"). In PostgreSQL, `array_length` of an **empty** array
returns `NULL`, not `0`; a `CHECK` that evaluates to `NULL` is **satisfied**, not violated.
Verified directly against the repo's own `courses-postgres-1` (Postgres 17):
```
CREATE TABLE t (state text, renditions text[] NOT NULL DEFAULT '{}',
  CHECK (state <> 'ready' OR array_length(renditions,1) >= 1));
INSERT INTO t (state) VALUES ('ready');   -- succeeds
SELECT state, renditions FROM t;          -- ready | {}
```
The insert succeeds. So a `ready` row with zero renditions — the exact empty/boundary case
row 4 is about — is **not** blocked at the database, contrary to the plan's explicit claim.
Nothing in §9's acceptance rows inserts this case to test the CHECK, so the gap is silent.
що зробити: `CHECK (state <> 'ready' OR cardinality(renditions) >= 1)` —
`cardinality('{}'::text[])` is `0` (verified against the same database), so the comparison is
never `NULL`. (`COALESCE(array_length(renditions,1), 0) >= 1` also works; `cardinality` is
shorter.)

### Row 1 residual (non-blocking, does not reopen the row)

Nginx's `client_max_body_size` (§5.c) is enforced from `Content-Length`; with
`proxy_request_buffering off` a client sending `Transfer-Encoding: chunked` has no
`Content-Length` for nginx to check against, so an oversized chunk streams straight to the
upload service. The browser's own tus client always declares a length, so the happy path is
unaffected, but the plan names no application-level chunk cap as a backstop for a
non-browser client. Cheap fix: `http.MaxBytesReader(w, r.Body, chunkBytes)` in the tus
`PATCH` handler, independent of nginx.

### Accounting

- Row 1 (hostile input): **closed** — one non-blocking residual (chunked-encoding bypass of
  the nginx cap).
- Row 2 (dependency failure): **open** — the mint-time disk gate has no disk to check, and
  the tus-completion rename-before-commit path has no reconciliation branch for its own crash
  window.
- Row 3 (race/replay incl. Asynq re-delivery): **open** — a graceful worker restart
  mid-transcode stalls a `processing` row past a redelivery that Asynq consumes as done,
  recoverable only after a second restart ≥ 40 minutes later or manual SQL.
- Row 4 (empty/boundary): **open** — the sub-360p "never upscale" claim is false as written
  (verified, upscales to 480×360) and the schema CHECK meant to make "ready ⇒ ≥ 1 rendition"
  structural does not fire on the column's own default value (verified against live
  PostgreSQL).
- Row 5 (rollback): **closed** — no counter-evidence found.

**Verdict: RED.** Four holes across three rows (2, 3, 4); two of the four (the 320×240
upscale and the mint-time disk gate) are independently corroborated by the concurrent
plan-critique round's execution against a different ffmpeg build, and two (the CHECK
constraint and the graceful-restart stranding's exact mechanism) are this review's own.
Edition 2 does not close defense rows 2 and 4 as claimed, and reopens row 3, which edition 1
had not raised as a defense gap at all.

## GATE 2 round 4 — plan-defense, edition 3.1 (sonnet, fresh context)

**Note on collision.** This session was assigned the heading "GATE 2 round 3 — plan-defense
(sonnet, fresh context)", which another concurrently-running reviewer had already claimed by
the time this section was written (it now sits above, targeting **edition 3, 1732 lines**,
verdict GREEN). This section is numbered round 4 instead of overwriting it, and reviews the
file as it stands now: **edition 3.1, 1787 lines**. The two sections disagree on one material
point, resolved below.

**The disagreement.** The existing round-3 section's row 4 verdict ("empty and boundary…
closed, and independently re-verified by execution rather than trusted") verified the ffmpeg
`min()` scale fix by execution but did not execute the schema `CHECK` that the same row leans
on. Independently, in this session, against the repository's own live `courses-postgres-1`
container (Postgres 17, `docker exec -i courses-postgres-1 psql -U courses -d courses`):

```sql
CREATE TABLE t_check_probe (state text NOT NULL, renditions text[] NOT NULL DEFAULT '{}',
  CHECK (state <> 'ready' OR array_length(renditions,1) >= 1));
INSERT INTO t_check_probe (state) VALUES ('ready');   -- INSERT 0 1 — succeeds
```

Edition 3's own §6 (the 1732-line text this session first read, before it changed under it)
carried exactly this `array_length` form. `array_length('{}',1)` is `NULL`; a `CHECK` that
evaluates to `NULL` is satisfied, not violated — so a `ready` asset with zero renditions,
exactly the empty/boundary case row 4 is about, was **not** blocked at the database, contrary
to §6's own "load-bearing" claim. This is not a new finding by this session: it is what the
file's own `## GATE 2 round 2 — plan-defense (sonnet, fresh context)` section below (targeting
edition 2, run out of numeric order) had already found and executable-verified. Between that
section being written and this one, the plan author folded the fix into **edition 3.1** — the
file changed under this review mid-session, which is why the line count moved from 1732 to
1787 while this session was working. Re-run just now against the same live container with
edition 3.1's corrected form:

```sql
CREATE TABLE t_check_probe2 (state text NOT NULL, renditions text[] NOT NULL DEFAULT '{}',
  CHECK (state <> 'ready' OR cardinality(renditions) >= 1));
INSERT INTO t_check_probe2 (state) VALUES ('ready');
-- ERROR: new row for relation "t_check_probe2" violates check constraint "t_check_probe2_check"
-- DETAIL: Failing row contains (ready, {}).
```

`cardinality('{}')` is `0`, so the fix genuinely holds. **Conclusion: the existing round-3
GREEN verdict was correct for what it checked but incomplete on row 4 — it missed executing
the one non-ffmpeg claim the row also depends on. Edition 3 (1732 lines) should not have been
called fully closed on row 4; edition 3.1 (this review's target) is.** §6 now states the
`NULL`-vs-`0` trap in prose, audits every other CHECK in the migration for the same class of
defect (a table of eleven, all clean), greps `00001`–`00004` for the same pattern (none), and
B32 inserts against both the broken and fixed shape rather than asserting either in prose.

### Independent verification this round

- **ffmpeg 7.1.1**, executed fresh (not reused from either prior round's output), including two
  cases neither prior round tried:
  - `min()`-corrected filter re-confirmed at 320×240 (`RESOLUTION=320x240`, exit 0), 1920×1040
    (`640x346`), 1080×1920 portrait (`202x360`/`406x720`/`608x1080`) — all match D-11's table.
  - **Extreme-aspect stress test, not in either prior round**: 3840×480 → `640x80`; 200×1920 →
    `38x360`; 3840×1080 → `640x180`; 100×1280 → `28x360`. All exit 0, all even dimensions, no
    libx264 "not divisible by 2" abort, no upscale. The independent per-dimension `min()` does
    not diverge from the source aspect ratio or produce a degenerate (zero or negative)
    dimension even at extreme ratios; the only consequence is a very narrow output for a very
    narrow source, which is a quality question `D-11`'s "portrait accepted as is" decision
    already owns, not a new crash class.
  - A 0.5 s source (shorter than both the 2 s IDR interval and the 6 s segment target) through
    the full multi-stream HLS command: exit 0, one `.m4s` segment, `EXTINF:0.533333`, valid
    `#EXT-X-ENDLIST`. No special-casing exists in the plan for a source shorter than one
    segment, and none is needed — ffmpeg's segmenter is duration-agnostic and just emits a
    shorter final segment. This closes "source shorter than one segment" from row 4 by
    execution rather than by the plan naming it (the plan itself never discusses this case by
    name; it is covered because nothing needed to be built to handle it).
- **PostgreSQL 17** (the repo's own `courses-postgres-1`): the `array_length`/`cardinality`
  probe above; scratch tables dropped after.
- **`docker-compose.yml`**: `api` (`:49-73`) confirmed `read_only: true`, `tmpfs: [/tmp]`, no
  `media-data` volume; `media-worker` (`:74-98`) confirmed `media-data:/media`; no
  `deploy.replicas` anywhere, consistent with the stated single-worker assumption.
- **`backend/cmd/media-worker/main.go:44-45`**: `Concurrency: 1, Queues: {"media": 1},
  ShutdownTimeout: 10 * time.Second` — read directly, matches D-7/D-10 exactly.
- **`hibiken/asynq@v0.26.0`** (`/Users/strelia/go/pkg/mod/github.com/hibiken/asynq@v0.26.0`),
  read directly: `processor.go:138-149` (`shutdown()`: `stop()`, then
  `time.AfterFunc(shutdownTimeout, close(abort))`, blocks on the semaphore); `processor.go:236-`
  `239` (`<-p.abort` branch: `requeue(lease, msg)` then `return`, with `cancel()` firing only in
  the deferred cleanup that runs after); `server.go:724-751` (`Shutdown` calls
  `processor.shutdown()`, which itself blocks until every worker goroutine has released its
  semaphore slot, **before** `srv.wg.Wait()` — the task-executing goroutine spawned inside
  `exec()` is never registered on `srv.wg` at all, so it can still be running when `Shutdown()`
  returns and `main` proceeds to close dependencies). This confirms D-10's "waits, then
  requeues to pending, only then cancels" line for line, including the detail that the
  handler's own goroutine is not joined by anything.
- **`backend/sql/queries/courses.sql:75-111`** and **`authoring.go:185-237`**: `lockAuthoringCourse`
  (`FOR UPDATE` on `courses`) precedes `LockEligibleAuthoringOwner` (`FOR SHARE`, on `users`,
  not `courses` — the briefing-is-wrong note in §0 holds) precedes the lesson upserts and
  `deleteMissingAuthoringLessons` — the exact order D-1's deadlock-free argument needs.
- **`nginx.conf:8,20,44-45`**, **`.gitignore:14`**: unchanged from prior rounds' citations,
  spot-checked and correct.

### The five questions, against edition 3.1

| # | Question | Closed by (plan.md location) | Verdict |
|---|---|---|---|
| 1 | Hostile input | `D-8` first-chunk sniff + `D-10` step 4 `Probe`/T8 (malformed video); `D-11` "the argv is built from three sources and no others" + R21's injected-filename assertion (hostile filename — no interpolation point exists for `original_name`, confirmed by re-reading D-11's three-source rule against the render argv); `D-2b` (ownership-at-mint is the protecting property, not id-secrecy) + `D-2d` step 1 (missing/evicted/expired/wrong-state ticket → `404`, no write) for forged/replayed tickets; `D-2d` step 2 (`Upload-Offset` must equal `received_bytes` exactly, else `409` with no write — covers overlapping/behind/ahead offsets alike) | **closed** |
| 2 | Dependency failure | Redis at enqueue (`D-15` ¶1, commit-then-enqueue, C6, retry-accepts-`queued`); PostgreSQL mid-transcode (`D-15` ¶2, deferred terminal state, C5 sweep on the next start — mechanism re-verified against real `asynq` source above); volume full/read-only (`D-15` ¶3, `Free()` at first `PATCH` and at claim, per-write `507`/`503`, readiness write-probe); ffmpeg killed and upload-service-down fold into the generic retryable-failure path (`D-10` step 12, `D-2d`'s crash-repair branch) without being named as their own scenarios | **closed, three residuals below (none reopen the row)** |
| 3 | Race and replay | two uploads per lesson (`D-5` pending index, R07/R38); Asynq re-delivery against `ready` (`D-10` step 1, R16/R36); publish racing eviction (`D-8`'s claim-verified-before-any-`hls/`-write ordering, with the stale-claim double fault explicitly narrated); lesson deleted mid-transcode (`D-4` `SET NULL`, `D-10` step 1, B14); worker restart mid-render (`D-7` C5 threshold-free + `D-10` step 2 retryable-not-`nil` + R35 — re-verified against real `asynq` shutdown order above, not merely re-asserted) | **closed** |
| 4 | Empty and boundary data | zero-byte upload (B04, schema `CHECK`); source shorter than one segment (no named location, but executed fresh above — ffmpeg needs no special-casing and produces a valid one-segment asset); source below the lowest rendition (`D-11` `min()` form, B17/B21/B30 — re-executed fresh, including the extreme-aspect stress cases above that neither prior round tried); exactly 2 GiB (B01, `CHECK (declared_bytes BETWEEN 1 AND 2147483648)`); third upload evicting the oldest (`D-5`, B06); **the `ready ⇒ ≥1 rendition` CHECK, the row's own load-bearing claim, re-executed against live PostgreSQL 17 in this session and confirmed genuinely fixed by the `cardinality()` form (§6, B32)** | **closed in edition 3.1** — was not closed in edition 3 (1732 lines); see disagreement above |
| 5 | Rollback | `§7`: supported path named; in-flight jobs archived visibly, nothing falsely `ready`; already-written media named as an explicit disk-only cost, not a correctness one; in-flight uploads resumable via `HEAD` on roll-forward; the one truly irreversible action named and forbidden (`goose down 00005`, with the reason); frontend-alone rollback stated as not a security boundary; nginx/upload-service coupling stated | **closed** |

### Residuals (do not reopen a row)

```
[edge] Two dependency-failure scenarios named in this review's brief have no dedicated D-15 paragraph
де:      D-15 (titled "Redis, PostgreSQL, and a volume that is full or read-only" — the upload
         service's own availability is not one of the three); D-10 step 12; D-2d step 3; §7
         "In-flight uploads"
чому:    "ffmpeg killed" (an external SIGKILL/OOM, not our own cancellation) reaches the same
         code path as any other render failure (`exec.CommandContext.Wait()` returns a
         signal-killed error, handled generically as a retryable failure under D-10 step 12/
         R14/R33) but is never named as its own row the way cancellation and disk exhaustion
         are. "Upload service down" (the container crashes or is OOM-killed mid-request) is
         answered in substance — D-2d step 3's crash-repair branch is exactly what reconciles
         the filesystem and the row after any process death mid-`PATCH`, regardless of cause,
         and §7's "In-flight uploads" states the row stays `created`/`uploading` and resumes
         via `HEAD` — but that answer sits under Rollback, not under the section titled to
         answer "dependency failure", so a reader searching D-15 specifically for "what if the
         upload service itself dies" finds nothing there.
що зробити: One sentence in D-15: "the upload service's own crash is not a special case — D-2d's
         per-`PATCH` transaction and step-3 repair branch already make every write idempotent
         under an arbitrary process death, and Compose's `restart: unless-stopped` (§5.a)
         brings it back; the client's next `HEAD` finds the true offset." One R-row fake: a
         transcoder returning a signal-killed-style error, asserting the same T6/Fail behaviour
         R14 already proves for other causes.
вага:    non-blocking
```

```
[edge] D-15's "down at Claim → the retry is a no-op" sentence is still imprecise in edition 3.1
де:      D-15 ¶2, the sentence "if PostgreSQL is down at `Claim`, nothing is claimed and the
         retry is a no-op"
чому:    Unchanged since edition 2. The file's own `## GATE 2 round 2 — plan-defense` section
         below already named this exact sentence as imprecise ("it is an error, so asynq
         retries — the plan's own preferred outcome"), and edition 3.1's author fixed the
         `array_length`/`cardinality` defect from that same section but left this sentence
         alone. A `Claim` `UPDATE` failing on a dead connection returns a connection error, not
         zero affected rows; the handler should return it (retryable), letting Asynq's own
         backoff redeliver — not treat it as the "someone else has it" zero-rows case D-10 step
         2 describes for a real `processing` row. The mechanism the plan wants is right; the
         sentence describing why is wrong, and an implementer reading D-15 literally could code
         a silent swallow instead of a returned error.
що зробити: "…nothing is claimed; the call returns an error, the handler returns it unwrapped,
         and Asynq retries with backoff — a dependency error, not the zero-rows case of step 2."
вага:    non-blocking
```

```
[complexity] Two more stale cross-references, on top of the one the round-3 section already recorded
де:      §12 NB-12 ("the reclaim index is on the predicate with an `id` key; a separate
         `expires_at` partial index serves class C3; `(state, updated_at)` serves C4–C6 | §6
         steps 5–6"), unchanged in edition 3.1, still contradicting §6's actual four numbered
         items and its own "No index serves the sweeper, deliberately" sentence — this one the
         round-3 section already flagged, still present, not re-litigated here; **new in this
         review**: §10's M02 row reads "Migration `00005_media.sql` with its CHECKs and **six
         indexes**", but §6 defines exactly three (`video_assets_pending_per_lesson`,
         `video_assets_generation`, `video_assets_by_lesson`) plus the implicit primary-key
         index — "six" is edition 2's count before N-15 (round 2) dropped three sweeper
         indexes, and M02's task description was never updated to match
чому:    Neither miscount changes what gets built — §6 is the section with the actual `CREATE
         INDEX` statements and it is correct — but a task-cut row is exactly the text an
         implementer reads to know what M02 must produce, and "six" sends them looking for
         three indexes that do not exist in this edition.
що зробити: M02's row: "…with its CHECKs and its three indexes…". Fold this into the NB-12
         rewrite the round-3 section already proposed for §12, so both stale counts are fixed
         in one pass.
вага:    non-blocking
```

### Accounting

- Five questions run against **edition 3.1** (1787 lines), the file as it stands at the time
  of writing. Holes: **0**. Every row is closed at a named location except the "source shorter
  than one segment" sub-case of row 4, which this review closes by fresh execution rather than
  by a plan citation (recorded as such, not hidden).
- This round's material contribution: (a) caught that the existing round-3 GREEN verdict, dated
  against edition-3's 1732-line text, had not executed the schema `CHECK` its own row 4 depends
  on, and confirmed independently — against the same live PostgreSQL the plan's own fix cites —
  that the defect was real in edition 3 and is genuinely fixed in edition 3.1; (b) an
  extreme-aspect-ratio ffmpeg stress test (3840×480, 200×1920, 3840×1080, 100×1280) neither
  prior round ran, which held; (c) two new non-blocking bookkeeping residuals (M02's stale
  "six indexes", D-15's stale "no-op" sentence) beyond the one the round-3 section already
  recorded (§12/NB-12 vs §6/§13).
- **Verdict: GREEN** for edition 3.1. Edition 3 (1732 lines) should be considered to have
  carried an open row 4 until this fix landed; nothing on disk today (edition 3.1) is blocking.

## GATE 2 round 5 — plan-critique verification, edition 3.1 (fable, fresh context)

Target: `docs/media/plan.md` **edition 3.1** (2026-09-06, 1787 lines). First **critique** pass
on editions 3/3.1 (rounds 3 and 4 were defense passes). Lenses: **4 run / 4 returned** —
AI-slop, logic holes, excess complexity, edge cases, all run inline by this critic, none
delegated. No lens returned empty. Method: every claim below was checked by execution or by
opening the cited source, not by reading prose. Scratch scripts live outside the repo; the
commands are quoted inline where the result is load-bearing.

### Facts established this round (execution and source)

- **The worker container runs `ffmpeg 5.1.9-0+deb12u1`, not 7.1.1.** `backend/Dockerfile:12`
  is `FROM debian:bookworm-slim AS runtime`, `:29-31` is `FROM runtime AS worker` +
  `apt-get install ffmpeg`; `docker exec courses-media-worker-1 ffmpeg -version` → 5.1.9.
  Every "verified against ffmpeg 7.1.1" statement in the plan (§0, D-8, D-11, §13) and in
  rounds 2 and 4 was executed on the host's Homebrew binary. Nobody ran the argv in the image
  that will run it.
- **The plan's D-11 argv, executed inside `courses-media-worker-1` (5.1.9)**, sources
  generated by the §5.f lavfi command into `/tmp`:
  - 320×240 → `RESOLUTION=320x240`, exit 0 (N-1 holds on the deployment binary).
  - 1280×720 → `640x360` + `1280x720`; 1920×1040 → `640x346`, `1280x692`, `1920x1040`,
    exit 0 (`force_divisible_by=2` and the `min()` form both exist in 5.1's `scale`).
  - Init naming `audio/init_0.mp4`, `360p/init_1.mp4`, `720p/init_2.mp4` — same as 7.1.1.
  - `EXT-X-INDEPENDENT-SEGMENTS` only in the video variant playlists; `FRAME-RATE` nowhere —
    same as 7.1.1.
  - **`AVERAGE-BANDWIDTH` is not written at all.** Each `EXT-X-STREAM-INF` carries only
    `BANDWIDTH`, and it is a *nominal* number derived from the rate-control flags:
    `542300` for the 360p rung on a 320×240, a 640×360 and a 640×346 source alike;
    `3440800` for 720p; `6740800` for 1080p. Nothing is measured from the segments. No
    `CHANNELS` attribute on the audio `EXT-X-MEDIA` either (7.1.1 writes it).
  - The 1080p rung without `-level` probes as High level 40 (`CODECS="avc1.640028"`); the
    lower rungs `avc1.640029`.
  - Timing under `cpus: 2.0`: a 60 s 1280×720 source through the two-rung argv takes **9 s**;
    the §5.f 2 s clips take well under a second. Thumbnail argv on 320×240 → `320,240`.
- **The same argv on the host's ffmpeg 7.1.1** (for the record, since that is what rounds 2/4
  measured): `AVERAGE-BANDWIDTH` present and measured — and for the **360p rung it is
  `377860` on the 2 s 1280×720 clip and `370072` on a 60 s clip, both above the 365 000
  ceiling** B26 compares against. The figure is video-only (the `360p/` directory's bytes
  over 60 s = 370 217 bit/s; the audio group averages 92 kbit/s and is not included). libx264's
  VBV caps the elementary stream; fMP4 container overhead plus the 2 s IDR cadence pushes the
  muxer's average ~1.4 % over the cap on a saturating source. Round 2 checked only the 720p
  rung (2 898 586 < 3 000 000) and generalised.
- **PostgreSQL 17 (`courses-postgres-1`)**, the §6 DDL run verbatim in a scratch schema and
  dropped after: `ready` + `DEFAULT '{}'` → `23514`; `ready` + explicit `'{}'` → `23514`;
  `ready` + `'{360p}'` → accepted; **`ready` + `'{NULL}'` → accepted; `ready` + `'{""}'` →
  accepted** (cardinality counts elements, not validity); `failed` without reason → `23514`;
  `claim_token` on a `queued` row → `23514`; `received_bytes > declared_bytes` → `23514`;
  `content_type = ''` → `23514`; `declared_bytes = 2147483648` accepted, `2147483649`
  refused; a second pending row for the lesson → `23505 video_assets_pending_per_lesson`;
  an evicted `queued` row does not block; `ON DELETE SET NULL` leaves orphans. **The
  auto-generated constraint names are `video_assets_check`, `video_assets_check1` …
  `video_assets_check5`** — the `ready ⇒ renditions` CHECK is `video_assets_check3`, an
  ordinal that moves if a CHECK is added above it. **`updated_at` one microsecond behind
  `created_at` (app clock read before the transaction, `created_at` left to `DEFAULT now()`)
  → `23514 video_assets_check1`.** `grep 'array_length\|cardinality' backend/migrations
  backend/sql` → nothing, as §6 states.
- **`asynq@v0.26.0`**, read: `processor.go:196-249` (the `select` on `abort` → `requeue` and
  return; on `lease.Done()` → `cancel()` + `handleFailedMessage(ErrLeaseExpired)`; on
  `ctx.Done()` → `handleFailedMessage` **without waiting for the handler goroutine**);
  `:138-149` (`shutdown` arms `abort` after `shutdownTimeout`, blocks on the semaphore);
  `:343` (`Retried >= Retry || SkipRetry` → archive); `server.go:724-751` (`Shutdown` order,
  `wg.Wait` excludes the handler goroutine); `server.go:401-405` (`n⁴ + 15 + rand(30)·(n+1)`
  seconds, so retries 1–3 wait **15–44 s, 16–74 s, 31–118 s** — the plan's "~16/~31/~96 s"
  are the minima for n = 1..3, one off; the conclusion it draws still holds);
  `recoverer.go:96-102` (lease-expired tasks `Retried >= Retry` → archive, else retry — and
  `rdb.Retry` increments `Retried`, so every crash recovery consumes one of the three);
  `context.go:25,33`, `client.go:228,416` as cited. **`internal/rdb/rdb.go:98-108`
  (`enqueueCmd`: `EXISTS <task key>` → conflict), `:382-392` (`markAsCompleteCmd` keeps the
  task key with the retention TTL), `:861-885` (`archiveCmd` keeps the task key; only
  trimming by age/size deletes it).** A task id stays taken for 24 h after completion and for
  the archive's lifetime (90 days by default) after archival.
- **`docker-compose.yml`**: `api` `:72-73` `read_only`, `tmpfs: [/tmp]`, no `media-data`;
  `media-worker` `:88` `media-data:/media`, `:89` `cpus: 2.0`; no `deploy.replicas`;
  `x-runtime` `:7-10` gives every service `restart: unless-stopped`, `init: true` and
  **`stop_grace_period: 40s`** — larger than asynq's 10 s `ShutdownTimeout`, so the graceful
  requeue-to-pending path the plan describes really completes before Docker's SIGKILL (the
  plan does not say this; it should, since it is a precondition it currently gets for free).
- **`backend/cmd/media-worker/main.go:44-45,52`**: `Concurrency: 1`, `Queues: {media: 1}`,
  `ShutdownTimeout: 10 s`, `server.Start(asynq.NewServeMux())` — `Start`, not `Run`, so
  "`Recover` completes before `server.Start`" is a one-line placement. **`backend/cmd/api/
  main.go:40-55`** wires course and identity only; nothing media, nothing filesystem — the
  plan's "no storage adapter in the API root" is a statement of intent about M09, and §2's
  row for `cmd/api/main.go` says it. But see N-23: M09's own task text still says "the disk
  gate".
- `backend/sql/queries/courses.sql:112-130` and `authoring.go:16,23`: every 4a insert writes
  `created_at` **and** `updated_at` from one application-clock argument (`$5, $5` / `$7, $7`);
  the `DEFAULT now()` is never relied on. Relevant to N-19.
- `courses.sql` has no `UPDATE courses SET instructor_id`; no reassign operation exists in
  4a/4b (relevant to R28's wording). `CourseEditorView.tsx:231` is the lesson `<fieldset>`,
  `:275` the remove button, `:288` `newLessonId()` — as cited. `smoke.sh:12` asserts zero
  handlers. `spec.md:67` carries the Q2 amendment. `nginx.conf:8,20,44-45`, `queue.go:37-51`,
  `httpserver/server.go:124`, `.claude/standards.md:12` (ADRs in the Мірник) — as cited.
  `infrastructure/tests/integration.sh:27-32` mounts only `backend/` into
  `golang:1.27.1-bookworm` — no ffmpeg there, as §5.e says.

### Part 1 — the four round-2 blocking closures, by location and substance

| # | Closure claimed at | Verdict |
|---|---|---|
| BLK-1 residual | §1 T6/T7/T9, D-7 C5 (threshold-free, before `server.Start`, single-worker assumption), D-10 step 2 (retryable error, not `nil`) and "Cancellation" (asynq's real order), B29, R34–R37 | **Closed.** The asynq sequence in D-10 matches `processor.go`/`server.go` line for line; `stop_grace_period: 40s` makes it real under Compose; `Start` (not `Run`) makes the ordering trivial; with one replica and `Concurrency: 1` every `processing` row at start is dead. Residual: the plan's own C6 leg has a silent no-op path (N-18, non-blocking) |
| N-1 | D-11 ladder bullet + argv + thumbnail (`min()` form), B21, B30, §5.f 320×240 clip | **Closed**, and now verified on the **container's** ffmpeg 5.1.9 as well as 7.1.1: `RESOLUTION=320x240`, thumbnail `320,240` |
| N-2 | D-15 (first-PATCH gate + claim gate), §1 T1, §2 `ports/storage.go` + `cmd/api/main.go`, §5.a, §8, B23 (M11) | **Closed** in substance — the gate sits in the two processes that mount the volume. One stale phrase contradicts it: M09's task text "the disk gate" (N-23) |
| N-3 | D-11 "Master playlist conformance" (decision: no writer, `FRAME-RATE` accepted non-conformance), B26, §11.11 | **Not closed — reopened as N-17.** The *decision* (no playlist writer) is sound and stands. The *row* and the *table* it rests on describe ffmpeg 7.1.1; the deployment runs 5.1.9, which writes no `AVERAGE-BANDWIDTH` at all, so B26 is red on the real muxer exactly as round 2 said — for a different attribute. And on 7.1.1 the ceiling comparison is red too (360p: 370–378 k > 365 k) |

**The §6 CHECK audit** (edition 3.1's table of eleven): spot-checked the six riskiest against
live PostgreSQL 17 — all six bite as the table says. The audit is correct as far as it goes;
two things it does not cover are recorded below (N-19: a CHECK that interacts with the clock
idiom; N-20: the fixed CHECK counts elements, not values) and one bookkeeping gap (N-21: the
constraint has no name to assert).

### Part 2 — findings on edition 3.1

```
[logic ×2 (logic, edge) + AI-slop] The plan's ffmpeg evidence is against a binary the deployment does not run; B26 is red in the worker image, and its ceiling comparison is red even on the binary it was measured on
де:      §0 "Facts established by the GATE 2 critics executing … ffmpeg 7.1.1"; D-8 "verified … against ffmpeg 7.1.1"; D-11 "Master playlist conformance — what the muxer actually emits, measured" (the table), "the `AVERAGE-BANDWIDTH` ffmpeg writes from the real segments matches reality, which is what Apple 1.26/1.27 asks for", "`libx264` and the native `aac` encoder are built into the Debian ffmpeg already in the worker image, so no image change is needed"; B26; §13 N-1/N-3 dispositions; `backend/Dockerfile:12,29-31`
чому:    Executed inside `courses-media-worker-1` (ffmpeg 5.1.9, Debian bookworm): the master carries `BANDWIDTH`, `RESOLUTION`, `CODECS`, `AUDIO` and **no `AVERAGE-BANDWIDTH`**; `BANDWIDTH` is nominal (`542300` for every 360p-rung source regardless of content). B26's "every `EXT-X-STREAM-INF` carries … `AVERAGE-BANDWIDTH` … below that rung's ceiling" fails on the first M16 run, and D-11's Apple 1.26/1.27 sentence is false for the binary that ships: the declared `BANDWIDTH` is not within 10 % of anything, it is ~2.5× the real 320×240 stream. Separately, on 7.1.1 itself the 360p rung's measured `AVERAGE-BANDWIDTH` is 370–378 kbit/s against a 365 kbit/s ceiling (VBV caps the elementary stream; fMP4 overhead lands above it on a saturating source), so the comparison is red on both binaries. This is the round-2 N-1 pattern one level up: a "verified" voice attached to an execution that was never run where it matters — and rounds 2, 3 and 4 all inherited it. The §5.f fixtures are generated *in the container*, so M16 would be the first time anyone saw 5.1.9's output.
що зробити: Decide, in the plan, one of: (a) **pin to what ships** — keep `bookworm-slim`, re-write D-11's table for 5.1.9 (no `AVERAGE-BANDWIDTH`, nominal `BANDWIDTH`, no `CHANNELS`), extend the recorded Apple non-conformance from "9.15 `FRAME-RATE`" to "9.14 `AVERAGE-BANDWIDTH` absent and 1.26/1.27 `BANDWIDTH` nominal", and rewrite B26 to assert presence of `CODECS`/`RESOLUTION`/`BANDWIDTH`, `EXT-X-INDEPENDENT-SEGMENTS` in every video variant playlist, and the **absence** of both `FRAME-RATE` and `AVERAGE-BANDWIDTH` so a base-image bump fires the trigger by itself; or (b) **own the image change** — `FROM debian:trixie-slim AS worker` (ffmpeg 7.1.x) named in §5.a as a topology change with its own smoke assertion on `ffmpeg -version`, and B26 then asserts `AVERAGE-BANDWIDTH` present and **within Apple's own +10 % of the ceiling** (or presence only), never "below". Under either: delete "no image change is needed", replace every "verified against ffmpeg 7.1.1" with the binary it was verified on, and add one `make smoke` row that prints the worker's `ffmpeg -version` so the evidence names its binary from now on. Paragraph-sized; no change to the slice, the task cut or the schema.
вага:    blocking
```

```
[logic] Recovery class C6 re-enqueues under a task id that Redis still holds, so its "re-enqueue" is a mapped-to-success no-op on two real paths
де:      D-7 C6 ("re-enqueue; no state change"); D-10 payload (`TaskID(assetId:attempt)`), "Duplicate enqueue" (`ErrTaskIDConflict` → success), step 12 ("On context cancellation: `Requeue` … return the cancellation" — silent on the final attempt); D-15 ¶1 ("two exits — the operator's retry … and … C6"); §7 "rolling forward re-queues them through recovery classes C5/C6"
чому:    `enqueueCmd` refuses any id whose task key exists; `archiveCmd` and `markAsCompleteCmd` keep the key (archive: until trimmed; complete: 24 h retention). Two paths leave a `queued` row beside an archived task of the same id: (1) the 30-minute `Timeout` fires on the **final** attempt — asynq archives on `ctx.Done()` without waiting for the handler (`processor.go:243-245`), the handler's `Requeue` then writes `queued`; (2) three crash/lease recoveries — each `rdb.Retry` increments `Retried` — after which the recoverer archives while T7 has already set `queued`. In both, C6's enqueue under `assetId:attempt` returns `ErrTaskIDConflict`, the adapter maps it to success, and the row stays `queued` at every restart. The operator exit (T9 bumps `attempt`, R41 shows the control) still exists, so the result is a manual recovery where the plan promises an automatic one — not a wrong state.
що зробити: C6 increments `attempt` in its own UPDATE and enqueues under the new id — the T9 shape, one clause — and D-15/§7 stop citing C6 as an automatic exit for anything but the BLK-7 window until it does. Step 12: state that cancellation on the final permitted attempt writes `Fail` (`render_failed`), not `Requeue`, so a source that cannot finish in 30 minutes reads `failed` with a reason instead of `queued` forever. One R row: "C6 against an archived task of the same id still produces a delivery".
вага:    non-blocking
```

```
[edge] `updated_at >= created_at` plus `DEFAULT now()` refuses the mint whenever the application clock is read before `pool.Begin`
де:      §6 step 1 (`created_at … DEFAULT now()`, `updated_at … DEFAULT now()`, `CHECK (updated_at >= created_at)`); D-9 ("Every write stamps `updated_at` with the application clock truncated to microseconds"); D-1 mint step 4 (the INSERT, columns unnamed)
чому:    Executed: an INSERT with `updated_at` one microsecond behind the transaction's `now()` fails `video_assets_check1`. `now()` is the transaction start; a clock captured before `Begin` (the natural shape of `time.Now` injected into a use case) is earlier. 4a avoids this by writing both columns from one argument (`courses.sql:127-128`, `authoring.go:16,23`) and never using the default. The plan cites 4a's clock discipline but leaves the DDL default in place and never says which column the mint writes.
що зробити: One sentence in D-9 or §6: "the mint writes `created_at` and `updated_at` from the same application-clock value, as `CreateAuthoringCourse` does; the `DEFAULT now()` on both is never relied on". Optionally drop the two defaults so the omission is a compile-time error rather than a runtime 23514.
вага:    non-blocking
```

```
[edge] The fixed CHECK counts elements, not renditions: `{NULL}` and `{""}` satisfy `cardinality(renditions) >= 1`
де:      §6 step 1 (`CHECK (state <> 'ready' OR cardinality(renditions) >= 1)`), the "load-bearing" paragraph; B32
чому:    Executed against PostgreSQL 17: `ready` with `renditions = '{NULL}'` and with `'{""}'` are both accepted. The application writes rung labels from a constant table, so nothing today produces either — but the paragraph says the CHECK is what stops "`ready` with nothing to play", and an array of one NULL is nothing to play.
що зробити: `CHECK (state <> 'ready' OR (cardinality(renditions) >= 1 AND array_position(renditions, NULL) IS NULL AND '' <> ALL (renditions)))`, or state in §6 that element validity is the domain's job and the CHECK guards emptiness only. B32 gains the `{NULL}` case either way.
вага:    non-blocking
```

```
[logic] B32 asserts a `23514` "naming the constraint", but the constraint has no name — PostgreSQL calls it `video_assets_check3`
де:      §6 step 1 (all six table-level CHECKs unnamed); B32 ("refused with a `23514` check violation naming the constraint")
чому:    Executed: the generated names are `video_assets_check`, `_check1` … `_check5`, an ordinal by position. A test that keys on `video_assets_check3` breaks when a CHECK is inserted above it; a test that does not key on the name proves only "some CHECK fired". D-1 already maps `23505` by `ConstraintName`; the three load-bearing CHECKs deserve the same handle.
що зробити: `CONSTRAINT video_assets_ready_has_renditions CHECK (…)`, `video_assets_failed_has_reason`, `video_assets_claim_only_while_processing`; B32 names the first.
вага:    non-blocking
```

```
[logic] R35 restarts the worker "mid-render" against fixtures that render in under a second
де:      R35 (precondition "a transcode in flight"; action `docker compose restart media-worker`); §5.f (three 2-second clips); M16
чому:    Measured in the container: the 2 s 1280×720 clip transcodes in well under one second; a 60 s clip takes 9 s. Playwright polls the status read every 3 s (D-12), so with the §5.f clips the "in flight" window closes before the test can observe `processing`, and the row is a race that usually observes `ready` on the first poll — passing without exercising the shutdown path it exists to prove. R35 is the only end-to-end evidence for the BLK-1 closure.
що зробити: §5.f generates a fourth clip, ≥ 5 minutes at 1280×720 (≈ 45 s of render under `cpus: 2.0`; a few MB on disk), used by R35 alone; R35 waits for `processing` before issuing the restart and asserts `processing` was observed. Alternatively assert through the worker's log that the claim happened twice.
вага:    non-blocking
```

```
[logic] M09's task text still puts "the disk gate" in the API — the one place the plan says it cannot be
де:      §10 M09 ("The API surface: … the actor mapping, **the disk gate**, error mapping …"); versus §2 `cmd/api/main.go` ("no `MediaStorage`"), D-15, §5.a ("load-bearing rather than incidental"), M11 (owns B23)
чому:    The task row is what an implementer reads to know what M09 must produce. Left as is, M09 wires a `MediaStorage` into the API root to have something to call `Free` on — the exact regression §5.a says "cannot regress silently".
що зробити: Delete the two words from M09; M11 already carries "the first-`PATCH` disk gate".
вага:    non-blocking
```

```
[logic] The UI's "stale threshold" is never given a number, and the server has two
де:      §2 `LessonVideoField.tsx`; M15 ("the stale-`queued` retry affordance"); R41 ("older than the stale threshold"); D-10 "Retry" (`processing` retryable at 40 min; `queued` at any age); D-7 C6 (10 min)
чому:    A control rendered for a `processing` row at any threshold under 40 minutes answers `409 video_not_retryable` when pressed; a `queued` row is retryable at any age server-side but should not offer retry while the single worker is legitimately busy with the previous upload (a queue of three 30-minute renders is healthy for 90 minutes). The implementer picks a number on the day and the two sides disagree.
що зробити: Name them: `queued` → retry after 10 min (C6's constant, exported to the client as a documented value); `processing` → after 40 min (T9's claim age). R41 uses both.
вага:    non-blocking
```

```
[AI-slop] Stale numbers and meanings a reader will trip on
де:      §2 migration row ("the four indexes"); M02 ("six indexes"); §6 (three plus the primary key); §12 NB-12 (still describes the dropped sweeper indexes as present, "§6 steps 5–6"); §8 `video_not_retryable` ("neither `failed` nor `queued`" — T9 also admits stale `processing` and refuses every state when the source has not landed); §8 `upload_too_large` ("or a `PATCH` would push past it" — that case is tus `413`, per the same section's last paragraph); D-10 backoff "~16 s / ~31 s / ~96 s" (minima for n = 1..3; retries 1–3 are n = 0..2: 15–44 s, 16–74 s, 31–118 s); R09 (its `processing` asset must be *fresh*, or T9 admits it)
чому:    Round 4 flagged two of these; edition 3.1 fixed neither. None changes what gets built, but the task row and the error vocabulary are the two texts most likely to be implemented literally.
що зробити: One pass: "three indexes" in §2 and M02; NB-12's disposition rewritten to "superseded by N-15"; §8 rows reworded; D-10's three numbers; R09's precondition gains "claimed seconds ago".
вага:    non-blocking
```

```
[complexity] Two columns nothing reads, and a request body nobody specifies
де:      §6 step 1 (`content_type`, `original_name`); §1 JSON surface (`GET …/videos` items carry neither); P01 ("a 5 MiB mp4's metadata" — the mint request fields are never listed anywhere in the plan); D-8 ("it is a display column")
чому:    The plan removed `duration_seconds` because "a column written for a future phase is a speculative column". `content_type` is written at mint and read by nothing — the sniff decides, and the plan says so. `original_name` is called display metadata, but no read in Phase 5 returns it, so it is displayed nowhere. Meanwhile the mint's request DTO — the one JSON body a client must send — has no field list, so `openapi.yaml` at M09 invents it.
що зробити: Drop `content_type`. Either add `originalName` to the list item (then it is display metadata) or drop it too. Write the mint body once: `{declaredBytes, originalName?}`. Saves one CHECK, one column, one zod field; costs nothing.
вага:    non-blocking
```

```
[complexity] Four rows and one gate that could go without losing evidence
де:      R34 versus B29 (both: a `processing` row → recovery → `queued`; R34 adds "force the terminal write to fail", which in Testcontainers means killing the pool mid-test for an observable B29 already has); D-15/D-10 step 3 claim-time `Free` gate + the first half of R33 (the render's `ENOSPC` is already mapped to retryable `storage_unavailable` — the pre-check only saves a doomed render's CPU); B09 and B10 (the gate asserting the gate — `make check-generated` and `make check` are already the gate order, not scenarios); the thumbnail (brief §11's layout names it, so it stays — but note the plan applies "no consumer in Phase 5 → not built" to `duration_seconds` and to the playlist writer and not to a file nothing reads until Phase 6; say why, one sentence)
чому:    Not padding — 84 rows for a slice with a third executable, a subprocess and a state machine is defensible, and the 17-task cut is along real evidence boundaries. But R34 is a second, harder-to-set-up proof of B29; the claim-time gate is a second site for one condition the render path already handles; B09/B10 restate §Машинні гейти. Roughly 0.5 h and four rows.
що зробити: Fold R34 into B29; keep the first-PATCH gate, drop the claim-time gate and R33's first case (or keep it and say it exists to avoid a 30-minute doomed render — that is a reason); keep B09/B10 if the acceptance-spec skill requires a gate row, otherwise fold into M17's evidence column.
вага:    non-blocking
```

```
[AI-slop] The decision sections narrate their own edition history at length, duplicating §12/§13
де:      D-7 (the "Edition 2 wrote … which was the correctness hole GATE 2 round 2 measured" paragraph); D-10 step 2 and "Cancellation" ("Edition 2 returned `nil` here, and GATE 2 round 2 measured what that costs …"); D-11 (three paragraphs on what edition 2 asserted and round 2 disproved); D-15 ("Edition 2 put the first gate at the mint, which cannot work …"); §6 step 1 (the `array_length` narrative); the preamble
чому:    An implementer needs the rule and the reason; the archaeology is in §13 already and in this file. About 150 lines of the 1787 are changelog prose inside decisions, and it is where the "verified" voice that N-17 catches lives — the more a section argues it was verified, the less anyone re-runs it.
що зробити: One sentence per section ("the age threshold was removed at GATE 2 round 2; §13 BLK-1"), the rest lives in §13. No behaviour change.
вага:    non-blocking
```

```
[edge] R28's action names an operation the product does not have
де:      R28 ("concurrently with an admin reassigning the course to B"); `backend/sql/queries/courses.sql` (no `UPDATE courses SET instructor_id`)
чому:    4a/4b have no reassign; the test must issue raw SQL, which is fine for an integration row but should be said, because "as an admin" implies an API path someone will look for.
що зробити: "…concurrently with a raw `UPDATE courses SET instructor_id = B` in a second transaction".
вага:    non-blocking
```

### Excess complexity — said plainly

17 tasks / 27.25 h / 84 rows is **not** ceremony for this slice. The cost drivers are binding
decisions (Q1's third executable; Q3's fMP4; the tus protocol; a subprocess that cannot run in
`make integration`), and the cut follows evidence environments rather than layers, which is
the right knife. What can go: the claim-time disk gate, R34, B09/B10 as rows, `content_type`
(and possibly `original_name`), and ~150 lines of in-section history. Roughly half an hour and
four rows. Everything else checked — four ports, claim token, `attempt`, the `expiration`
extension, the six-class sweeper, the three fixtures, two ADRs — earns its place.

### Four-lens accounting

- **AI-slop** — 3: the "verified against ffmpeg 7.1.1" voice attached to a binary that does not
  ship (folded into N-17), the stale numbers/meanings, the in-section history. Otherwise clean:
  §12/§13 tables, the placement table and the reuse table are load-bearing.
- **Logic holes** — 6: N-17 (the deployment binary), C6's same-id no-op, B32's unnamed
  constraint, R35's fixture, M09's stale gate, the unnumbered UI threshold. Checked and
  holding: the asynq shutdown order and its Compose precondition, `Start`-before-`Recover`
  placement, the single-worker inference, the lock order (unchanged since round 2), the API
  root wiring nothing, `stop_grace_period` ≥ `ShutdownTimeout`.
- **Excess complexity** — 2: the two unread columns; the four foldable rows plus one gate.
- **Edge cases** — 4: the 360p ceiling on a saturating source (folded into N-17), the clock
  idiom under the `updated_at` CHECK, `{NULL}`/`{""}` renditions, R28's nonexistent operation.
  Checked and clean by execution: 2 GiB boundary, empty array both shapes, the pending index
  with an evicted row, `SET NULL` orphans, odd-height 1920×1040 on 5.1.9, the sub-1-s thumbnail
  seek; by source: crash recovery consuming a retry (correct, just unstated).

Cross-lens duplicates collapsed with the lens count kept in the tag.

### Verdict

- Round-2 blocking: **BLK-1 residual closed; N-1 closed (now also on the container's ffmpeg);
  N-2 closed (one stale phrase in M09); N-3 not closed** — the decision stands, the row and
  the table under it are red on the deployment binary, reopened as N-17.
- **New blocking: 1** — N-17, the ffmpeg evidence base is the host's 7.1.1 while the image
  ships Debian's 5.1.9; B26 is red in the worker container (`AVERAGE-BANDWIDTH` absent) and its
  ceiling comparison is red on both binaries. A paragraph-sized decision (pin to 5.1.9 and
  assert what it writes, or own a `trixie-slim` worker stage), no change to the slice, the
  cut or the schema — but it invalidates the "verified" chain three rounds relied on, and the
  first time anyone would have seen it is M16.
- Non-blocking: **11** (N-18 … N-28 in the order above).
- Bookkeeping from round 4 not yet applied in edition 3.1: NB-12/§6, M02 "six indexes",
  D-15's "no-op" sentence — all folded into the findings above.
- **RED** — edition 3.2 required for N-17; everything else in edition 3.1 stands, and the
  edition-3/3.1 mechanisms (threshold-free C5, retryable step 2, first-PATCH gate, `min()`
  filter, `cardinality`) are each confirmed by execution or source this round.

## GATE 2 round 6 — closing verification, edition 3.2 (fable, fresh context)

Target: `docs/media/plan.md` **edition 3.2** (2026-09-06, 1952 lines). Closing round, scoped
to three things by the caller: (1) N-17's closure by execution **inside
`courses-media-worker-1`**, (2) the tightened `ready ⇒ renditions` CHECK against the repo's
live PostgreSQL, (3) a sweep of the edits edition 3.2 introduced and nothing else. Mechanisms
confirmed in rounds 2–5 (threshold-free C5, retryable step 2, first-`PATCH` gate, `min()`
filter, asynq's shutdown order, `stop_grace_period` ≥ `ShutdownTimeout`) were not re-litigated.
Lenses run inline only where they bear on those three items; nothing delegated. Scratch
scripts live outside the repo (`b26.sh`, `r35.sh`, `check.sql`), piped into the containers
with `docker exec -i … sh -s` / `psql`; every number below is from those runs.

### 1. N-17 — executed in the worker container, ffmpeg 5.1.9-0+deb12u1

`docker exec courses-media-worker-1 ffmpeg -version` → `5.1.9-0+deb12u1`. The two §5.f clips
(`testsrc2=size=1280x720:rate=15` and `320x240`, 2 s, `sine=440`) were generated in the
container into `/tmp/g6/`, then pushed through the D-11 argv **verbatim** — two rungs for the
720p source (`split=2`, `min(640,iw)/min(360,ih)` and `min(1280,iw)/min(720,ih)` with
`force_divisible_by=2`, `-level:v:N 4.1`, `-crf 21`, the 365k/730k and 3000k/6000k
rate-control pairs, `force_key_frames expr:gte(t,n_forced*2)`, `-hls_time 6`, `independent_segments`,
`fmp4`, `init.mp4` template, `var_stream_map` with `agroup:aac`), one rung for the 240p source.
Both exit 0. B26's assertions, one by one, against the files:

| B26 asserts | 1280×720 source | 320×240 source |
|---|---|---|
| every `EXT-X-STREAM-INF` carries `CODECS`, `RESOLUTION`, `BANDWIDTH` | 2 of 2 (`BANDWIDTH=542300,RESOLUTION=640x360,CODECS="avc1.640029,mp4a.40.2"`; `3440800,1280x720`) | 1 of 1 (`542300,RESOLUTION=320x240`) |
| one `EXT-X-MEDIA:TYPE=AUDIO` group, referenced by every variant | 1 line, `GROUP-ID="group_aac"`, `URI="audio/index.m3u8"`; both variants `AUDIO="group_aac"` | same |
| `EXT-X-INDEPENDENT-SEGMENTS` in every **video variant** playlist | `360p/index.m3u8` 1, `720p/index.m3u8` 1 | `360p/index.m3u8` 1 |
| … and absent from the master | 0 | 0 |
| `AVERAGE-BANDWIDTH` absent everywhere | 0 across all `.m3u8` | 0 |
| `FRAME-RATE` absent everywhere | 0 | 0 |
| 360p `BANDWIDTH` byte-identical between the two clips | `542300` | `542300` |

Also observed, consistent with D-11's table: no `CHANNELS` on the `EXT-X-MEDIA` line; init
naming `audio/init_0.mp4`, `360p/init_1.mp4`, `720p/init_2.mp4`; one `.m4s` per variant;
`EXT-X-VERSION:7`. **One precision worth recording for M16: the `audio/index.m3u8` playlist
carries no `EXT-X-INDEPENDENT-SEGMENTS` either** (5.1.9 writes the tag only into playlists
with a video stream). B26 says "every **video variant** playlist", which is exactly right; an
implementer who generalises it to "every variant playlist" gets a red row on the audio
rendition. The nominal `BANDWIDTH` is arithmetic, not luck: hlsenc sums the video stream's
CPB `max_bitrate` (from `-maxrate`) and the audio `bit_rate`, then adds 10 % —
`(365 000 + 128 000) × 1.1 = 542 300`, `(3 000 000 + 128 000) × 1.1 = 3 440 800` — so
"byte-identical across sources" is a property of the flags, and B26's identity assertion is
the right one.

**The R35 fixture, as §5.f specifies it** (`-t 300` at `1280x720:rate=15`, same command as
the short clips): generation **22 s**; the two-rung render **39 s** (§5.f says "about 45 s";
holds); output 93 MB under `/tmp`. The master playlist for it is attribute-for-attribute
identical to the 2 s clip's, `542300` / `3440800` again. See N-32 for the fixture's size.

**N-17 verdict: closed.** B26 asserts what 5.1.9 actually writes — presences and absences —
and the identity assertion is the correct replacement for the deleted ceiling comparison. §0's
table names 5.1.9 / the container on every row where that is what ran, and the two rows whose
original run was on the host's 7.1.1 say so and record the container re-run. `Dockerfile:12`
and `:29-31` are cited correctly (`FROM debian:bookworm-slim AS runtime`; `FROM runtime AS
worker` / `USER root` / `apt-get install ffmpeg`). The three recorded non-conformances use the
numbering `research-codecs.md:81` gives (9.13/9.14 `BANDWIDTH` peak and `AVERAGE-BANDWIDTH`
MUST, 9.15 `FRAME-RATE` MUST, 1.26/1.27 declared-within-10 %), and D-11's table, §0's row,
B26 and §14's N-17 row agree with each other and with the files.

### 2. The tightened CHECK — executed against `courses-postgres-1`, PostgreSQL 17.11

§6 step 1's `CREATE TABLE` run **verbatim** in a scratch schema (`gate6`, `search_path =
gate6, public` so the `course_lessons` FK resolves), probed, then dropped. Every refusal below
is the `CONSTRAINT NAME:` line psql printed under `\set VERBOSITY verbose`.

| Insert, `state = 'ready'` unless noted | Result |
|---|---|
| `renditions` omitted (`DEFAULT '{}'`) | `23514 video_assets_ready_has_renditions`, no row |
| `'{}'` explicit | `23514 video_assets_ready_has_renditions` |
| `'{NULL}'` | `23514 video_assets_ready_has_renditions` |
| `'{""}'` | `23514 video_assets_ready_has_renditions` |
| `'{360p,NULL}'` | `23514 video_assets_ready_has_renditions` |
| `'{360p,""}'` | `23514 video_assets_ready_has_renditions` |
| `'{360p}'` | **accepted** |
| `state = 'queued'`, `'{}'` | accepted (the guard is on `ready` only) |
| `UPDATE … SET state = 'ready'` on that `queued` row | `23514 video_assets_ready_has_renditions` — the constraint bites on the publish path too, not only at insert |
| `state = 'failed'`, no reason | `23514 video_assets_failed_has_reason` |
| `state = 'queued'` with a `claim_token` | `23514 video_assets_claim_only_while_processing` |
| `updated_at` 1 µs behind `created_at` | `23514 video_assets_updated_not_before_created` |
| `generation = 0` (a column-level CHECK) | `23514 video_assets_generation_check` — PostgreSQL's own `<table>_<column>_check` name, stable by column, not ordinal |

`pg_constraint` for the table: `video_assets_ready_has_renditions`,
`video_assets_failed_has_reason`, `video_assets_claim_only_while_processing`,
`video_assets_updated_not_before_created`, `video_assets_evicted_not_before_created` all
present by name; the column CHECKs are `video_assets_{id,generation,state,declared_bytes,
original_name,attempt,failure_reason}_check`; and **one constraint is still auto-named
`video_assets_check`** — the `received_bytes >= 0 AND received_bytes <= declared_bytes`
CHECK, which PostgreSQL promotes to table level because it references a second column. Nothing
keys on it (N-33).

**`DEFAULT now()` dropped:** omitting `created_at` → **`23502` not-null violation** on
`created_at`; omitting `updated_at` → `23502` on `updated_at`. Both columns written from one
clock read a minute *before* the transaction (the 4a idiom) → **accepted**. So the failure
mode N-19 described (a runtime `23514` that depends on the clock/transaction race) is gone and
replaced by a deterministic `23502` on the first test that omits the column — a query error,
as §14's N-19 row says. (§6's own prose calls it "a compile-time error in the query"; it is
not — sqlc does not check `NOT NULL` coverage of an `INSERT` — see N-33.)

**CHECK verdict: verified.** All four empty shapes, plus the two mixed ones, are refused under
the named constraint; a valid single rendition is accepted; the three load-bearing names are
what a violation reports; B32's assertion text matches the executed behaviour exactly.

### 3. Sweep of the edition-3.2 edits

Confirmed clean, by source or execution:

- **B33 / §5.d** — `ffmpeg -version` in the container prints `ffmpeg version 5.1.9-0+deb12u1`
  on its first line; `smoke.sh` runs on the host and already uses `docker compose exec -T`, so
  the assertion is placeable exactly where §5.d puts it; a `5.1.x` line check passes a
  bookworm security bump and fails a trixie one, which is the intended sensitivity.
- **R35's action from a Playwright test** — `make e2e` is `cd frontend && npm run e2e` on the
  host, and `frontend/e2e/recovery.spec.ts` already shells out to docker, so `docker compose
  restart media-worker` from `media.spec.ts` has an in-repo precedent. The observe-`processing`
  precondition is satisfiable: 39 s of render against a 3 s poll.
- **R34 → B29** — B29's precondition and result carry R34's "no state was invented" clause;
  M06 owns B29 and nothing references R34 any more.
- **`content_type` removed** — §1's mint body is `{declaredBytes, originalName}` with "no
  `contentType`" stated; §6 has no column and no CHECK; the only remaining mentions are the
  historical disposition rows (§12 NB-6, §14 N-26/N-27) and §9's cut note.
- **§0 labelling** — every ffmpeg row names binary and location; the `Dockerfile` line cites
  are right; the "a 60 s source takes 9 s" row extrapolates correctly to the 39 s measured.
- **D-15 / §7 scoping of C6** to the BLK-7 window is stated in both places; R42 is consistent
  with the D-7 row.

New findings, all on text edition 3.2 introduced:

```
[edge] Step 12's "final attempt ⇒ Fail(render_failed)" keys on any cancellation, not on the Timeout it was written for
де:      D-10 step 12 ("On context cancellation: `Requeue` … — except on the final permitted attempt, where cancellation writes `Fail` with `render_failed`"); D-10 "Cancellation" (shutdown requeues to *pending*, then cancels); §14 N-18
чому:    The rationale is `asynq.Timeout` — on the final attempt asynq archives, so `queued` would be forever. But the same `ctx.Done()` fires on graceful shutdown, where asynq has just requeued the task to pending and the retry count is unchanged. If the handler's write lands (the plan says it usually will not, pool closing — but the rule is for the case where it does), a restart that happens to hit a final-attempt render writes `failed`/`render_failed` for a source nothing was wrong with, and the pending task is delivered at restart against a `failed` row — a state neither step 1's nil list nor step 2's re-read enumerates. The operator exit exists (T9 accepts `failed` with the source landed), so nothing is lost; it is a guessed terminal state, which D-15 says the design does not do.
що зробити: One clause: the `Fail` branch applies when `errors.Is(ctx.Err(), context.DeadlineExceeded)` (the Timeout); `context.Canceled` keeps `Requeue`. And step 2's re-read gains "`failed`, `created`, `uploading` → `nil`" beside "ready/evicted/absent → nil", which N-31 needs too.
вага:    non-blocking
```

```
[logic] C6's `attempt` bump produces a duplicate task on every restart that finds a backlog
де:      D-7 C6 (predicate `queued AND updated_at < now − 10 min`; action "increment `attempt`, then enqueue"); D-15 ("C6 is an automatic exit only for this window"); R42
чому:    The predicate cannot tell "committed, enqueue failed" (BLK-7) from "enqueued fine, the worker was down or busy for ten minutes" — a backlog of `queued` rows whose tasks are still *pending* in Redis. Edition 3.1's same-id enqueue handled the backlog perfectly (conflict → success, no duplicate) and the archived path badly; 3.2 inverts that: the archived path now delivers, and the backlog path enqueues a second task per row (`assetId:n+1` beside the pending `assetId:n`). Benign under the claim guard with `Concurrency: 1` — the duplicate arrives after the first delivery finished and finds `ready` (nil) or a requeued row (a legitimate extra attempt) — but it is a change in behaviour the plan does not name, and the `failed` case of the duplicate is the unenumerated branch above.
що зробити: State it in D-7 ("a C6 enqueue may duplicate a pending task; the duplicate is a no-op through step 1/2") and add that assertion to R42's result, or, if a duplicate is unwanted, have C6 consult `Inspector.GetTaskInfo("media", assetId:attempt)` and skip rows whose task is pending/scheduled/retry — one call, but a second Redis dependency in the recovery pass; the plan's "named rather than solved" stance is fine here as long as it is named.
вага:    non-blocking
```

```
[AI-slop] §5.f's 5-minute fixture is "a few MB on disk" — it is 77 MB, and the row costs two to three minutes, not "roughly a minute"
де:      §5.f fourth-clip row ("A few MB on disk, generated like the others"; "the long one adds roughly a minute to `make e2e`"); M16
чому:    Executed with §5.f's own command and `-t 300`: `e2e-1280x720-5min.mp4` is **77 477 349 bytes** — `testsrc2` is high-detail synthetic content and the fixture command sets no rate control, so libx264's default CRF 23 spends ~2 Mbit/s on it. Nothing breaks (2 GiB cap, 8 MiB chunks → ten `PATCH`es, 93 MB of HLS on the volume), but the wall-clock is 22 s to generate + copy-out + a ~77 MB browser upload + 39 s render interrupted + the 10 s asynq drain + container restart and healthcheck + a full 39 s render ≈ 2–2.5 min, and the fixture is copied out of the container on every setup.
що зробити: Add `-crf 32` (or `-b:v 600k -maxrate 600k -bufsize 1200k`) to the fixture command for the long clip only — the render time is set by frame count, not by source bitrate, so the R35 window is unchanged — and say "roughly two minutes". One flag.
вага:    non-blocking
```

```
[AI-slop] Four wording residues in the 3.2 text
де:      §6 ("Dropping the defaults makes the omission a compile-time error in the query"); §6 ("Every table-level CHECK is named"); §12 NB-6 ("Addressed — nullable display metadata with a length CHECK only"); §13 N-3 row ("Closed by deciding …" with no pointer)
чому:    Executed: the omission is a runtime `23502`, deterministic and caught by the first test — a query error, as §14 N-19 correctly says; sqlc does not verify `NOT NULL` coverage. `received_bytes … <= declared_bytes` is a table-level constraint by promotion and is still `video_assets_check` (harmless: nothing keys on it, and it is the only one). NB-6 describes a column that no longer exists. §13's N-3 row reads as closed on its own terms; §14 N-17 is the actual closure.
що зробити: "a deterministic `23502` on the first test rather than a clock-dependent `23514`"; "every load-bearing CHECK is named" or name `video_assets_received_within_declared`; NB-6 → "superseded — column dropped, §14 N-26"; N-3 row → "superseded by §14 N-17".
вага:    non-blocking
```

### Accounting

- **AI-slop** — 2 (the fixture size/cost, the four residues). The "verified" voice is now
  attached to the binary that ships; every §0 ffmpeg row was re-executed here and holds.
- **Logic holes** — 1 (C6 duplicates on a backlog). The N-17, N-19, N-20, N-21 closures are
  each confirmed by execution.
- **Excess complexity** — 0 new; the 3.2 cuts (R34, `content_type`) are clean and the two
  declines (claim-time gate, B09/B10) carry their reasons.
- **Edge cases** — 1 (the cancellation cause on the final attempt). Checked and clean by
  execution: all six empty/mixed rendition shapes on insert and on update, the no-default
  clock idiom, the audio playlist's lack of the independent-segments tag (B26's wording is
  exact), the 320×240 rung on 5.1.9, and the 5-minute render window.

### Verdict

- **N-17: closed.** Executed inside `courses-media-worker-1` on ffmpeg 5.1.9: B26 asserts
  precisely what the muxer writes — `CODECS`/`RESOLUTION`/`BANDWIDTH` present, one audio
  group, `EXT-X-INDEPENDENT-SEGMENTS` in the video variant playlists and not the master,
  `AVERAGE-BANDWIDTH` and `FRAME-RATE` absent, 360p `BANDWIDTH` `542300` on both a 1280×720
  and a 320×240 source. No mismatch.
- **CHECK: verified** against PostgreSQL 17.11 — `{}` defaulted, `{}` explicit, `{NULL}`,
  `{""}` (and `{360p,NULL}`, `{360p,""}`, and the UPDATE path) all refused as `23514
  video_assets_ready_has_renditions`; `{360p}` accepted; dropping `DEFAULT now()` turns an
  omitted clock into `23502`, not `23514`.
- **New blocking: none.** Non-blocking: **4** (N-30 … N-33 above) — a cancellation-cause
  clause in step 12, a named behaviour for C6 on a backlog, one flag on the long fixture, and
  four wording residues. None changes the slice, the cut or the schema.
- **GREEN** — edition 3.2 closes N-17 and the round-5 non-blocking set; Phase 5 may proceed
  to task breakdown, with N-30 … N-33 folded into the first implementation commit that
  touches D-10, D-7 and §5.f rather than into another edition.


## GATE 2 — constitution (inline, orchestrator, per .claude/standards.md §Конституція)

Checked plan edition 3.2/3.3 against AGENTS.md and the product brief's architecture rules.

| Rule | Verdict |
|---|---|
| Domain: standard language only, no framework/DB/transport/external packages | PASS — `internal/media/domain` holds asset/actor/source/errors with table tests; `Actor{UserID, Administrator}` is a capability, not a re-copied role vocabulary (D-3) |
| Application depends on domain and ports only | PASS — `application/{upload,transcode}.go` over four ports, fake-port tests |
| Ports require an actual external dependency | PASS — `MediaStorage`, `Transcoder`, `TranscodingQueue`, `AssetRepository` all have real adapters; the speculative `LessonOwnership` port was DELETED at GATE 2 (D-1) rather than kept unused |
| Adapters point inward; no filesystem path crosses the port | PASS — opaque `StorageKey`, `os.Root`-scoped resolution (ADR 004) |
| Brief §2 rule 12 (worker is its own executable sharing modules/ports) | PASS — and extended: the tus upload service is a third executable on the same shape |
| Brief §2 rule 13 (video bytes never proxied through Go API or SSR) | PASS — the API composition root wires no storage adapter at all and `api` keeps no media mount; this is stated as load-bearing and was verified in code |
| Brief §2 rule 14 / media data plane | PASS — `/media/*` keeps returning 403 for everyone in Phase 5; no playback is claimed |
| Brief §11 "no arbitrary user arguments to FFmpeg" | PASS — argv from a fixed ladder plus probed dimensions only; the original filename never reaches argv or a response body (D-11, R21) |
| No microservices without real need (rule 10) | PASS — the third executable exists to keep bytes out of the API, a named rule, not taste; tusd declined with a recorded trigger |
| Explicit business error codes; no secrets in logs | PASS — §8 vocabulary; the ticket id is not logged (verified at httpserver/server.go:124) |
| No new forbidden dependency | PASS — tus core implemented in-repo, no tusd image, no new frontend dependency |
| Machine gates and placement per manifest | PASS — migration 00005 with the canary renumber, `make check` order unchanged, ADRs 008/009 owned by M01 |

Violations: 0. **Constitution: PASS.**

## GATE 2 final verdict (Phase 5)

Six rounds. Round 1 critique RED (8 blocking). Round 2 critique RED (BLK-1 residual plus
three new, two proven by executing ffmpeg). Rounds 3-4 defense: row 4 had been closed on a
FALSE sub-360p claim, and the `array_length` CHECK was proven against live PostgreSQL to
enforce nothing. Round 5 critique RED with the deepest finding of the phase — every prior
"verified by execution" had run against the host's ffmpeg 7.1.1 while the worker image
ships 5.1.9, and the master-playlist row was red on the binary that actually runs. Round 6
closing verification GREEN: N-17 closed by execution inside the worker container, the named
CHECK verified against live PostgreSQL 17.11, no new blocking. Constitution PASS.
**GATE 2 GREEN on plan edition 3.3.** Verdicts by fresh-context fable critics and sonnet
examiners; constitution by the orchestrator; plan author: delegated opus.

Method note for later phases: three separate reviewers verified by execution and all three
measured the wrong binary. "Verified by execution" is only evidence when the report names
WHICH binary and WHERE it ran — every fact table in this plan now carries that label, and
B33 prints the worker's ffmpeg version into every smoke run so the evidence base cannot
drift from the deployment silently again.
