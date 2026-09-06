# Phase 5 — media upload and transcoding: delivery report

Delta to [plan.md](plan.md) (edition 3.3 plus amendments §17–§25) and
[spec.md](spec.md). Written 2026-09-07 at phase close. Post-deploy is left for
`post-deploy-check`.

## 1. What was done

Eighteen tasks M01–M17 (M03a added mid-phase), each implemented with Codex (gpt-5.6-sol)
and re-verified here before commit; per-task evidence in `work/report-m01.md` …
`report-m16.md`.

A person can now put a video on a lesson. The chain: the editor mints a ticket, a
hand-written tus client streams chunks to a **third executable** that the Go API never
touches, the last chunk completes inside the same transaction that queues the job, the
worker claims it under a token, ffmpeg 5.1.9 renders the ladder, and the result publishes
atomically. `/media/*` still refuses everyone — this phase produces bytes and grants
nobody the right to read them.

**Deviations from the plan, all dated in it:**
- **§17** the probe gained a duration, because D-11's thumbnail seek needed one.
- **§18** R22's precondition was impossible; construction split so the API cannot hold
  byte-handling methods.
- **§19** D-2d step 3 refused a missing scratch file — **the protocol as written could not
  accept a first byte**; also the file could get ahead of the truth column, and a failed
  compensating truncate was discarded.
- **§20/§21** completion moved into one transaction, and the container-sniff failure became
  the one terminal policy outcome — both were defects in the plan's own contract section.
- **§22** the log vocabulary admits a bounded `unknown` state rather than a fabricated one.
- **§23** R13's evidence split across M13 and M16; `AGENTS.md` corrected mid-phase.
- **§24** resume refuses an unattested file; the tus wire proof moved to e2e (provisional,
  backlog **B-2**).
- **§25** R13's post-transcode check uses the real published path, and the `processing`
  observation moved to the long fixture.

## 2. Acceptance scenarios

89 rows, each owned by exactly one task, every task's evidence inside the exit-0
`make check`. Rows "not verified": **0**.

| ID | Scenario | Owner | Actual | Verdict |
|---|---|---|---|---|
| P01 | An admin mints an upload ticket | M09 | closed by M09; inside the exit-0 `make check` | збіглося |
| P02 | An owning instructor mints | M09 | closed by M09; inside the exit-0 `make check` | збіглося |
| P03 | HEAD → PATCH lands the source and queues the job | M11 | closed by M11; inside the exit-0 `make check` | збіглося |
| P04 | The worker transcodes a real clip | M16 | closed by M16; inside the exit-0 `make check` | збіглося |
| P05 | The editor advances to ready without a reload | M16 | closed by M16; inside the exit-0 `make check` | збіглося |
| P06 | Replacing a video keeps exactly two generations | M05 | closed by M05; inside the exit-0 `make check` | збіглося |
| P07 | The course videos read answers one row per asset | M09 | closed by M09; inside the exit-0 `make check` | збіглося |
| P08 | The full browser journey | M16 | closed by M16; inside the exit-0 `make check` | збіглося |
| P09 | The editor renders the control and the progress | M15 | closed by M15; inside the exit-0 `make check` | збіглося |
| P10 | The worker registers the media handler | M12 | closed by M12; inside the exit-0 `make check` | збіглося |
| P11 | Polling runs only while work is in flight | M14 | closed by M14; inside the exit-0 `make check` | збіглося |
| B01 | The size ceiling is exactly 2 GiB | M09 | closed by M09; inside the exit-0 `make check` | збіглося |
| B02 | A chunk beyond the declared length is refused | M10 | closed by M10; inside the exit-0 `make check` | збіглося |
| B03 | A crash mid-chunk is repaired, and the upload resumes | M10 | closed by M10; inside the exit-0 `make check` | збіглося |
| B04 | A zero-length upload is refused at mint | M09 | closed by M09; inside the exit-0 `make check` | збіглося |
| B05 | The ladder never upscales, end to end | M16 | closed by M16; inside the exit-0 `make check` | збіглося |
| B06 | A third generation evicts the oldest | M06 | closed by M06; inside the exit-0 `make check` | збіглося |
| B07 | Migration over a populated database | M02 | closed by M02; inside the exit-0 `make check` | збіглося |
| B08 | The failing-migration canary still rolls back | M02 | closed by M02; inside the exit-0 `make check` | збіглося |
| B09 | Contract and client stay in step | M14 | closed by M14; inside the exit-0 `make check` | збіглося |
| B10 | The ordered gate passes | M17 | closed by M17; inside the exit-0 `make check` | збіглося |
| B11 | The storage adapter cannot be walked out of | M07 | closed by M07; inside the exit-0 `make check` | збіглося |
| B12 | Publication is all-or-nothing, and purge is idempotent | M07 | closed by M07; inside the exit-0 `make check` | збіглося |
| B13 | A non-video source is refused at probe | M08 | closed by M08; inside the exit-0 `make check` | збіглося |
| B14 | A deleted lesson leaves a reapable orphan, not a leak | M06 | closed by M06; inside the exit-0 `make check` | збіглося |
| B15 | A course with no video reads empty | M09 | closed by M09; inside the exit-0 `make check` | збіглося |
| B16 | An expired upload window does not lock a lesson, but a running transcode does | M05 | closed by M05; inside the exit-0 `make check` | збіглося |
| B17 | The ladder is chosen by the source, never by the client, and always yields a rung | M03 | closed by M03; inside the exit-0 `make check` | збіглося |
| B18 | The container is sniffed on the first chunk | M11 | closed by M11; inside the exit-0 `make check` | збіглося |
| B19 | The transcoding metrics and all their series exist before any job | M12 | closed by M12; inside the exit-0 `make check` | збіглося |
| B20 | The upload control is operable without a mouse | M15 | closed by M15; inside the exit-0 `make check` | збіглося |
| B21 | Odd-dimension, portrait and sub-rung sources produce even, never-enlarged argv | M08 | closed by M08; inside the exit-0 `make check` | збіглося |
| B22 | Generation numbering is monotonic across evictions | M05 | closed by M05; inside the exit-0 `make check` | збіглося |
| B23 | A full volume refuses the first byte rather than half-writing | M11 | closed by M11; inside the exit-0 `make check` | збіглося |
| B24 | The recovery pass re-queues a stale claim, and the stale claim cannot publish | M06 | closed by M06; inside the exit-0 `make check` | збіглося |
| B25 | A completed job is observed in the histogram | M16 | closed by M16; inside the exit-0 `make check` | збіглося |
| B26 | The playlists carry what the target binary really produces, and the three gaps stay visible | M16 | closed by M16; inside the exit-0 `make check` | збіглося |
| B33 | The worker's ffmpeg is the binary the plan targets | M12 | closed by M12; inside the exit-0 `make check` | збіглося |
| B27 | Audio is one shared rendition, not a copy per rung | M16 | closed by M16; inside the exit-0 `make check` | збіглося |
| B32 | The `ready`-implies-renditions constraint actually bites, on every shape of nothing | M02 | closed by M02; inside the exit-0 `make check` | збіглося |
| B29 | Start-up recovery reclaims a claim of any age, including one whose terminal write never landed | M06 | closed by M06; inside the exit-0 `make check` | збіглося |
| B30 | A sub-360p source is encoded at its own size, not enlarged | M16 | closed by M16; inside the exit-0 `make check` | збіглося |
| B31 | A crash between the completion rename and its commit does not lose the source | M11 | closed by M11; inside the exit-0 `make check` | збіглося |
| B34 | A crash after the final append and fsync but before its commit is repaired | M11 | closed by M11; inside the exit-0 `make check` | збіглося |
| B35 | An unsupported first chunk consumes the ticket before a waiting valid PATCH can advance it | M11 | closed by M11; inside the exit-0 `make check` | збіглося |
| B28 | The render argv pins the Apple codec constraints | M08 | closed by M08; inside the exit-0 `make check` | збіглося |
| R01 | Anonymous is refused everywhere | M09 | closed by M09; inside the exit-0 `make check` | збіглося |
| R02 | A student is refused everywhere | M09 | closed by M09; inside the exit-0 `make check` | збіглося |
| R03 | A non-owning instructor is refused | M09 | closed by M09; inside the exit-0 `make check` | збіглося |
| R04 | Cross-site is refused first | M09 | closed by M09; inside the exit-0 `make check` | збіглося |
| R05 | An unknown lesson is not found | M09 | closed by M09; inside the exit-0 `make check` | збіглося |
| R06 | A malformed id is rejected before lookup | M09 | closed by M09; inside the exit-0 `make check` | збіглося |
| R07 | A second upload cannot start while one is running | M09 | closed by M09; inside the exit-0 `make check` | збіглося |
| R08 | A failed or stranded asset can be retried | M09 | closed by M09; inside the exit-0 `make check` | збіглося |
| R09 | A ready or freshly-claimed asset cannot be retried | M09 | closed by M09; inside the exit-0 `make check` | збіглося |
| R10 | An unknown, evicted or expired ticket writes nothing | M10 | closed by M10; inside the exit-0 `make check` | збіглося |
| R11 | A wrong offset is a conflict, not a corruption | M10 | closed by M10; inside the exit-0 `make check` | збіглося |
| R12 | An oversized chunk is refused at the edge | M13 | closed by M13; inside the exit-0 `make check` | збіглося |
| R13 | `/media/*` stays refused for everyone | M13 | closed by M13; inside the exit-0 `make check` | збіглося |
| R14 | A transient failure on the final attempt rests failed | M04 | closed by M04; inside the exit-0 `make check` | збіглося |
| R15 | A malformed source is never retried | M04 | closed by M04; inside the exit-0 `make check` | збіглося |
| R16 | A retried job cannot corrupt a published asset, and a swept one is not an error | M04 | closed by M04; inside the exit-0 `make check` | збіглося |
| R17 | A duplicate enqueue is not a failure | M04 | closed by M04; inside the exit-0 `make check` | збіглося |
| R18 | Cancellation leaves the asset claimable, not stranded | M04 | closed by M04; inside the exit-0 `make check` | збіглося |
| R19 | A dependency failure degrades safely | M04 | closed by M04; inside the exit-0 `make check` | збіглося |
| R20 | Cancelling the transcoder ends the process | M08 | closed by M08; inside the exit-0 `make check` | збіглося |
| R21 | No user-controlled string reaches ffmpeg | M08 | closed by M08; inside the exit-0 `make check` | збіглося |
| R22 | Media errors disclose no filesystem detail | M09 | closed by M09; inside the exit-0 `make check` | збіглося |
| R23 | Business codes reach the frontend domain intact | M14 | closed by M14; inside the exit-0 `make check` | збіглося |
| R24 | An unsaved lesson offers no upload | M15 | closed by M15; inside the exit-0 `make check` | збіглося |
| R25 | An interrupted upload offers resume or cancel, not a dead lesson | M15 | closed by M15; inside the exit-0 `make check` | збіглося |
| R26 | The client guard is not the mechanism | M16 | closed by M16; inside the exit-0 `make check` | збіглося |
| R27 | An unsupported codec set is refused by the domain | M03 | closed by M03; inside the exit-0 `make check` | збіглося |
| R28 | Ownership is resolved inside the minting transaction | M05 | closed by M05; inside the exit-0 `make check` | збіглося |
| R29 | Two concurrent PATCHes at one offset do not corrupt the file | M10 | closed by M10; inside the exit-0 `make check` | збіглося |
| R30 | Cancel frees the lesson immediately | M10 | closed by M10; inside the exit-0 `make check` | збіглося |
| R31 | An enqueue failure after the commit is recoverable | M04 | closed by M04; inside the exit-0 `make check` | збіглося |
| R32 | A transient failure on a non-final attempt returns the asset to queued | M04 | closed by M04; inside the exit-0 `make check` | збіглося |
| R33 | A render that runs out of disk is retryable, not permanent | M04 | closed by M04; inside the exit-0 `make check` | збіглося |
| R35 | A worker restarted mid-render leaves nothing stranded | M16 | closed by M16; inside the exit-0 `make check` | збіглося |
| R43 | The container sniff accepts the Q2 set and nothing else | M03a | closed by M03a; inside the exit-0 `make check` | збіглося |
| R42 | C6 produces a real delivery against a retained task id, and its duplicate is inert | M04 | closed by M04; inside the exit-0 `make check` | збіглося |
| R36 | A delivery for an unclaimable row is retried, not consumed | M04 | closed by M04; inside the exit-0 `make check` | збіглося |
| R37 | Retry frees a stale claim but never a live one | M09 | closed by M09; inside the exit-0 `make check` | збіглося |
| R38 | Two concurrent mints refuse cleanly whichever index fires | M05 | closed by M05; inside the exit-0 `make check` | збіглося |
| R39 | A sniff-failed asset is not retryable | M09 | closed by M09; inside the exit-0 `make check` | збіглося |
| R40 | A chunked body cannot stream past the chunk cap | M10 | closed by M10; inside the exit-0 `make check` | збіглося |
| R41 | A stranded row offers a retry instead of polling forever, at the threshold the server honours | M15 | closed by M15; inside the exit-0 `make check` | збіглося |
| R44 | An interrupted upload never resumes from an unattested file | M14 | closed by M14; inside the exit-0 `make check` | збіглося |

## 3. Non-obvious findings

- **The plan carried a protocol that could not start.** D-2d step 3 made a missing scratch
  file a 503, but a freshly minted row has none by design — every upload would have failed
  at its first byte. Six GATE 2 rounds, three plan editions and two independent review axes
  did not catch it; the implementer did, because it was the first task that had to *run* the
  sequence rather than reason about it.
- **Three defects came from one section.** §2.1 — written by the gate owner after M04's stop
  — wrote blanket rules where the decisions it served were specific. A port table is a
  contract; a contract that disagrees with the decision it implements is a defect even when
  every method in it is individually sound.
- **"Verified by execution" is only evidence when it names the binary.** Three separate
  reviewers ran real ffmpeg and all three measured the host's 7.1.1 while the worker ships
  5.1.9, where the master playlist omits `AVERAGE-BANDWIDTH` entirely. `B33` now prints the
  worker's version into every smoke run.
- **A CHECK constraint that enforced nothing.** `array_length('{}',1)` is NULL and a NULL
  CHECK is *satisfied*, so `ready` with no renditions passed. Proven against live
  PostgreSQL, fixed with `cardinality`, and `B32` now inserts every empty shape.
- **Resume could splice a different file onto a stored prefix**, producing a source that
  carries a valid container prefix, passes ffprobe and reaches `ready` as a "successful"
  video. The client now refuses to resume on fingerprint mismatch **and on absence**.
- **Test-layer traps worth carrying:** `docker cp` cannot read a tmpfs mount; ffprobe
  repeats a stream entry per segment; MSW + happy-dom drop custom headers in both
  directions; pgx sends parameterised `Exec` over the extended protocol so multi-command
  fixtures fail with 42601.

## 4. Screenshots

Not applicable as static images; the flow is exercised visually by
`frontend/e2e/media.spec.ts` against real Compose.

## 5. User flow walk-through

Verified by `media.spec.ts` (not by hand). An admin opens a lesson in the course editor,
picks a video file, watches the transfer progress, then the badge move through
`processing` to `ready`; reloading keeps it `ready`. A sub-360p source is never enlarged;
a 1280×720 source produces two rungs sharing one audio rendition. Restarting the worker
mid-render still reaches `ready`. A student's direct API mint is refused, and `/media/*`
answers 403 for every actor before and after publication.

## 6. Risks and limitations

- **No playback.** `/media/*` is refused by design; there is no player, no signed access,
  no progress. That is Phase 6.
- **One audio rendition.** Multi-language audio rides the shared HLS audio group (brief
  §24.15) and costs ~1.4% storage per language, but the delivery mechanism — extra tracks in
  the source versus separate audio uploads — is undecided and belongs to Phase 6.
- **AV1 deferred** behind the three conditions in brief §24.14; the fMP4 choice keeps that
  addition remux-free.
- **The tus wire proof lives only in e2e** while MSW cannot round-trip headers — recorded as
  backlog **B-2** with a hard trigger: take it before Phase 6's player adapter, which will
  speak in headers too.
- **The worker is single-concurrency and the recovery pass assumes it.** A second replica or
  `Concurrency > 1` needs a real lease heartbeat; the trigger is recorded in D-7.
- **No observability for any of this** beyond metrics nobody reads — backlog **B-1**.

## Post-deploy

_Left for `post-deploy-check`. No deploy target exists in the repository._
