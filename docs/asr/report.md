# Phase 8 — ASR (speech recognition): delivery report

Delta to [plan.md](plan.md) and [spec.md](spec.md). First written 2026-09-22 at the close of
A01–A12; rewritten 2026-09-24 at the close of A13 from code, test and live-stack evidence, and
updated the same day at the close of A14.
There is no production deployment to check.

**Status: 22/22 acceptance rows verified at the level `spec.md` §12 requires. The phase is
closed by its rows** (AGENTS.md), behind a green ordered gate (§5). A13 left it at 19/22; A14
closed C08, C16 and C19 (§3).

The previous edition of this report said «17/22 verified». That was wrong twice over: it did
not count C18, and it counted C08, C10, C14, C19 and — through fakes — C02 and C04 as verified
when the code broke them or the evidence never took the production path (§4). It is kept in git
history (`54fc22c`), not here.

## 1. What was done

An author can generate a speech-recognition draft for any lesson with a `ready` video. The
frontend sends a regeneration request (optional language; `auto` never reaches the `language`
field); the API enqueues an asynq task; the `asr-worker` claims the draft row, extracts audio
with ffmpeg, runs `faster-whisper` through a Python subprocess, and completes the draft. The
author sees the draft move through queued → recognizing → ready (or `failed` with a named reason,
or `no_speech`), sees it marked stale when the lesson's video is replaced, and can load it into
the transcript editor without the editor remounting. Unsaved edits over a loaded draft survive a
reload, and are offered as stale when the draft has been regenerated meanwhile. Publishing a
draft goes through the existing `SaveLessonTranscript` with optimistic-lock conflict detection.

Before A13 none of this worked end to end in the running stack: the engine could not load its
model, and once it could, it crashed on any real speech (§4). A13 is the slice that made
recognition actually work and proved it in the browser against real speech.

This phase did not add payments, comment threads, an instructor workspace or a production
deployment.

## 2. Acceptance rows

Level = what `spec.md` §12 requires. Verdict is against that level, not against the strongest
evidence available.

| # | Row (short) | Level | Evidence | Verdict |
|---|---|---|---|---|
| C01 | Ready asset without its draft → task; second pass → no second task | integration | Postgres: selector subtest «C01 selects only ready assets missing their current draft» (`draft_integration_test.go`); Redis/asynq: repeated enqueue with the same marker → `ErrAlreadyRunning` (`queue_integration_test.go`) | verified — **in two parts**; no single test runs `EnqueuePending` through both stores |
| C02 | Draft after success carries text, language, generation | integration | unit `TestProcessHappyPath`; live: `smoke.sh` real speech through `transcribe.py --model` → ≥1 non-empty segment; e2e C03/C17 reach a ready draft with text | verified |
| C03 | Author sees in-progress until done; named reason after failure | e2e | `asr.spec.ts` C03 (queued/recognizing observed, then ready). Failure branch: unit `recognize_test.go` named failures, integration C14, unit `LessonDraftField.test.tsx:146`, draft-status contract | verified — failure branch at lower levels by decision (§6) |
| C04 | Regeneration rewrites the draft, leaves the transcript alone | integration | `queue_integration_test.go`: after a task is really processed (completed) and after it is archived, a new marker enqueues and the old one conflicts; unit `TestRegenerate*` | verified (was broken, §4) |
| C05 | Publish via `SaveLessonTranscript`, refuses a foreign token | contract | `TestC05PublishRefusesStaleToken` (`draft_handler_test.go`); real CAS: `write_integration_test.go:66` | verified |
| C06 | Draft from an old generation shown stale | unit + e2e | unit staleness tests; `asr.spec.ts` C06 (replace video → «older video» text) | verified (UI was broken, §4) |
| C07 | Killed worker leaves no `running` draft after start | integration | `TestDraftRecoverAgainstPostgreSQL` (Recover subtests) | verified at repository level; the startup wiring (`cmd/asr-worker/main.go:63`) is not under test |
| C08 | Lesson deleted during work → task ends, draft cascades away | integration | Postgres subtest «C08 lesson deletion removes the draft before completion» (claim → delete the lesson → draft row gone → `Complete` → `ErrNotFound`); unit `TestProcessNotFoundOnComplete` for «no retry» | verified — in two parts (`8ecda78`) |
| C09 | Replaced video → old generation never overwrites the new draft | integration | «C09 delayed old claim cannot overwrite the new draft», «C09 old completion after new claim is superseded», «re-claiming the same generation succeeds» | verified (was broken at the repository level, §4) |
| C10 | Second task on the same asset/generation → visible «already running» | integration | `queue_integration_test.go` (same marker → `ErrAlreadyRunning`); contract 409 `already_running`; unit `Regenerate` on a running draft | verified (was broken, §4) |
| C11 | `auto` never reaches `language`; select offers product locales only | unit + e2e | unit `TestAutoNeverBecomesALanguage`; `asr.spec.ts` C11 (three options; request body without `language` for auto, `en` for English) | verified |
| C12 | Draft segments pass the domain ceilings | unit | `TestDraftValidatesWithTheSameConstructorsAsTheTranscript`, ceiling tests, `TestProcessSegmentLimitPermanent` | verified |
| C13 | Video without speech → named result, not an empty ready draft | e2e | `asr.spec.ts` C13 (sine fixture → «No speech detected»); `smoke.sh` sine → `speech_seconds 0` | verified |
| C14 | Over-timeout task is taken down with a named reason; NULL duration refused | integration | «deadline fails running draft with timeout» (Postgres); `TestEnqueuePendingFailsNilDuration` | verified — with a controlled 500 ms deadline, not the queue-derived one (§7) |
| C15 | `asr` queue visible in `courses_queue_depth`, `media` still there | smoke | `smoke.sh` | verified |
| C16 | `/readyz` red while the model file is unavailable | integration | `TestC16ReadyzRequiresModelFile` (`cmd/asr-worker/main_test.go`): production `httpserver` handler with the real `modelReady` — no `model.bin` → 503, `model.bin` a directory → 503, a regular file → 200; smoke for the live 200 | verified (`221b9d3`) |
| C17 | Completion with the editor open does not remount it; author gets a load proposal | e2e | `asr.spec.ts` C17 | verified (wiring was missing, §4) |
| C18 | Local snapshot carries the draft token; after regeneration the offer is stale | unit + e2e | unit `recovery.test.ts`, `TranscriptEditor.test.tsx`; `asr.spec.ts` C18 | verified (mechanism was missing, §4) |
| C19 | Publication not blocked by a missing transcript, but the author sees a message | e2e | `admin.spec.ts` P20: the notice is visible and publication goes through; unit `CourseEditorView` for both cases | verified — the product half was added in A14 (`3cc8754`) |
| C20 | Image carries ffmpeg, Python with faster-whisper, weights; runs as 10001 | smoke | `smoke.sh`: ffmpeg, uid, import, `test -f model.bin`, real sine and real speech runs through the production command | verified |
| C21 | A draft cannot name another lesson's asset | integration | «C21 rejects an asset belonging to another lesson» | verified |
| C22 | Deleting a `video_assets` row takes its draft (D8-15) | integration | «C22 deletes a draft when its asset is deleted» | verified |

## 3. Rows closed in A14

A13 ended at 19/22. These three were not proven at their level, and A14 closed them:

| # | What was missing | Closed by |
|---|---|---|
| C08 | an integration test deleting a lesson under a running draft | `8ecda78` |
| C16 | any test of the red `/readyz` branch; `cmd/asr-worker` had no tests | `221b9d3` |
| C19 | the product half: the author was never told transcripts will come — A12 counted the row whole | `3cc8754` |

## 4. Defects found and fixed in A13

Every one of these passed the A12 gate green. Each was found by running the real path — the live
worker, the real stack, the browser — not by reviewing text.

| # | Defect | Found by | Fix |
|---|---|---|---|
| 1 | Engine passed `--model /opt/whisper/models`, an `HF_HOME` cache root without a flat `model.bin`: every recognition failed at model load; `/readyz` stayed green (it stat-ed the directory) | running the engine's own command in the live worker | `23fe407` |
| 2 | `transcribe.py` read `seg.avg_log_prob`; faster-whisper 1.2.1 names it `avg_logprob`: every recognition with speech crashed → C02 never worked | first real-speech run after #1 | `3b46565` (+ speech smoke, speech fixture) |
| 3 | C10: task-ID conflict swallowed into a silent 202; key was `lessonID:generation` | code audit of the evidence path | `62f3993` |
| 4 | C04: asynq keeps a finished task's key (24 h completed, up to 90 days archived), so regeneration after any finished run was a silent no-op | reading asynq v0.26.0 Lua scripts; proven by `queue_integration_test.go` | `62f3993` (marker in the task ID) |
| 5 | C14: a task past its deadline left the draft `running` until the worker restarted | code audit | `62f3993` |
| 6 | C09: `Claim` accepted any existing asset row, so a late old-generation task could overwrite a newer draft; safe today only by queue order (Concurrency 1, FIFO) | repository test forcing the order | `9b5c837` (UPSERT guard) |
| 7 | Selectors took any ready asset, not the current one: after a video replacement the two generations queued each other forever (one lesson recognised five times in the live stack); a manual regenerate could pick the old video | e2e run + live logs | `26ffe07` |
| 8 | Deleting a lesson leaves its ready asset with `lesson_id` NULL; the selector scanned NULL into a string and every `EnqueuePending` pass failed: one deleted lesson turned automatic recognition off platform-wide | live `asr-worker` log (`poll failed` every minute) | `26ffe07` |
| 9 | UI stopped polling after 202 (no draft row before `Claim` → `null` → polling off), so no draft ever appeared without a reload; a regenerated ready draft and a replaced video were not picked up either | e2e run | `43faef5` |
| 10 | C17/C18: load-draft never reached the editor; the draft token never reached the recovery snapshot; a loaded draft's language was dropped, and recovery keyed by language lost unsaved work on reload | code review of the A10 wiring; e2e | `8299e97`, `43faef5` |

## 5. Ordered gate

`make check` (`Makefile:83`), last run at the close of A14 on the tree committed as `3cc8754`;
log `work/make-check-a14.log` (not tracked). **`CHECK_EXIT=0`.**

| Step | Result |
|---|---|
| `format-check`, `lint`, `typecheck` | green |
| `test` | frontend 482/482; Go packages ok |
| `integration` | ok — 34 unique Go packages across `test` and `integration` (63 `ok` lines: the log holds two Go runs), 0 FAIL |
| `check-generated`, `build`, `up` | green; images rebuilt |
| `drain-queue` | `asynq:{media}` 0 → 0, `asr` 22 → 0 |
| `seed-identities`, `seed`, `smoke` | green — smoke includes real sine and real speech through the production engine command |
| `e2e` | **74 passed**, 0 failed (11.6 min); B19 login budget total 7, maximum consecutive 3 |

At the close of A13 the same gate ran green on `915c987` (480 frontend, 33 packages, e2e 74/74,
`work/make-check-a13.log`). Its first two runs had been red, both on tests, not the product:
1. `transcript.spec.ts:467` (7a) asserted the recovery key `author:lesson:uk`. The language
   decision (§6) changed the key to `author:lesson`, and the e2e of other phases was not run after
   that change. The test now expects the new key and asserts the stored snapshot's language, so it
   is not weaker than before.
2. `asr.spec.ts` C06 lost a race with automatic recognition: after a video replacement the worker
   re-recognises the new generation within one enqueue tick, and «stale» disappears. See §6.

The gate leaves two tracked screenshots under `docs/foundation/screenshots/` rewritten; that is
backlog B-18 and they are not committed.

## 6. Decisions

- **C03 failure branch proven below e2e** — Courses 2 (reviewer), 2026-09-24. No honest fixture
  produces a named failure in the running stack without manipulating the `media-data` volume;
  the branch is proven by unit, integration and contract tests instead (§2). Check: the listed
  tests.
- **Draft language vs transcript language** — proposed by Courses 1 after Courses 2's diagnosis,
  confirmed by the user 2026-09-24 («Підтверджую»). The recovery key is lesson + author and the
  snapshot carries language and draft token; loading a draft in another language over an existing
  transcript is refused instead of silently switching the editor and creating a second
  transcript. Closes the 7a «lesson + language» recovery key for this phase; revisit when an
  editor for several languages exists. Full text: `work/brief-a13-language.md`.
- **C09 held as an invariant, not by queue order** — Courses 2, 2026-09-24:
  `WHERE asr_drafts.generation <= EXCLUDED.generation`; a late old claim returns `Claimed: false`
  with no failure metric. Check: the C09 subtests.
- **The UI polls on an accepted job, capped at 15 minutes of queue wait** — reviewed by
  Courses 2; the cap bounds waiting before `Claim`, not recognition itself.
- **C19 notice follows decision A6** (`docs/asr/spec.md:223`) — Courses 1, 2026-09-24,
  reviewed by Courses 2. Publication is never blocked. The notice promises transcripts only
  through speech recognition, so a lesson without a video is not promised one. Check: unit
  `CourseEditorView`, e2e P20.
- **Status assertions are narrowed by their text** — Courses 1, 2026-09-24, reviewed by
  Courses 2. A second `role="status"` on the course editor broke unscoped lookups, and lookups by
  accessible name find nothing, because a status takes no name from its content. `admin.spec.ts`
  now uses `getByRole('status').filter({ hasText })`. The product gets no `aria-label`: the text
  is the message.
- **C06 holds its window by pausing the `asr` queue** — Courses 1, 2026-09-24; reviewed and
  accepted by Courses 2. A stale draft is observable only between «the new video is ready» and «the worker
  re-recognised it» (≤ one enqueue tick); the e2e sets `asynq:{asr}:paused` for the replacement
  and removes it in `finally`. Nothing is faked: the replacement is a real upload. Precedent:
  `media.setup.ts` already drains the `media` queue through redis.

## 7. Named limitations

1. **C14 deadline is controlled.** The integration test uses a 500 ms deadline, not the
   queue-derived `TaskTimeout` (minimum 300 s); the full asynq deadline path is not exercised.
2. **Automatic enqueue is not live on an open page.** The UI polls only after the author starts a
   job; a job the worker enqueues on its own shows after a reload or refetch on focus.
3. **Manual edits over an existing draft also read as stale after regeneration** — the price of
   carrying the draft token through a reload (§6).
4. **C07 startup wiring and C01 end to end** are proven only in parts (§2).
5. **C03 is not fully e2e.** Only its happy path runs in the browser. The «named reason after a
   failure» branch is proven by unit (`recognize_test.go`), integration (C14), contract (draft
   status) and frontend unit (`LessonDraftField.test.tsx:146`) tests — a dated lowering of the
   proof level (§6), because no honest fixture yields a named failure in the running stack.
6. **Native HLS in real Safari** — inherited from Phase 6, not closed here.
7. **Review did not catch a cross-phase break; only the full gate did.** Changing the stored
   recovery key in step 4g broke a 7a e2e that asserted the old key. Neither author nor reviewer
   saw it in the diff — the break was outside the diff. It surfaced only in the full `make check`.
   Any change to shared or persisted state is checked with the full e2e set, not the targeted one.
8. **The C19 notice starts conservative.** Until a lesson's lazy transcript query resolves, its
   transcript counts as unknown and the notice shows; it hides once every saved lesson is known
   to have a transcript.
9. **Two backlog items found here stay outside the phase:** B-20 (draft endpoints absent from
   `openapi.yaml`; the frontend uses a hand-written adapter) and B-21 (a read without language
   picks the first transcript alphabetically, for students too).

## 8. What's next

Phase 8 is closed at 22/22. Its limitations (§7) stay named and are not claimed away. B-20 and B-21
are recorded in `work/backlog.md` outside the phase. The next phase is not chosen: the roadmap
candidates and the open phase-number question (ASR vs orders) are in `work/AGENDA.md`.
