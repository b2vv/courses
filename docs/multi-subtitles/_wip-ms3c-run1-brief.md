# MS3c run 1 brief for Codex — e2e of the MS3 rows, part 1 (migration + two-track journey)

Repository: current directory, branch `ms2-3-asr-per-track` @ `7b20d2c`. Shared working tree: do NOT run git
checkout/reset/stash/clean/rm, do not commit, do not touch generated files or backend. Product code under `frontend/src` only if an
e2e proves a defect — then STOP that test, write a QUESTION with the evidence (request log / screenshot path) and continue the others;
Courses 1 decides the fix.

WORKING FILE: docs/multi-subtitles/_wip-ms3c-run1.md
Append AFTER EVERY STEP (append, not rewrite). First entry before the first tool: a `NEXT:` line.
Then FACT / DECISION / NEXT / QUESTION. Final answer = concise summary of the file.

Read first: `docs/multi-subtitles/plan-ms2-3.md` §4 (the MS3 table: rows M3-R1..R15) and the MS3c checkpoint at the end of
`docs/multi-subtitles/report.md` (two dated decisions there bind you). UI: `frontend/src/modules/admin/ui/LessonLanguageTabs.tsx`,
`frontend/src/modules/transcript/ui/{TranscriptEditor,LessonDraftField,LessonTranscriptField}.tsx`,
`frontend/src/modules/media/ui/LessonAudioTracksField.tsx`, locales `frontend/src/**/locales/{uk,en}.ts`. The e2e runs with
`locale: 'uk-UA'` — select by role + Ukrainian accessible names the code actually renders (read them, do not guess).

## Runtime (live Compose stack on Docker context `kiri`, already built from this branch)
Prefix every docker/npx command with `/Users/strelia/project-data/courses-kiri/courses-kiri` (it sets the context and compose files;
never change the global Docker context). Targeted run:
`cd frontend && /Users/strelia/project-data/courses-kiri/courses-kiri npx playwright test e2e/asr.spec.ts --project=chromium -g '<title>'`
(the `setup` project runs first and regenerates fixtures). Recognition takes minutes; set per-test timeouts like the existing tests.
Test credentials only from `frontend/e2e/auth.ts` storage states; never print them.

## Scope of run 1
1. **M3-R9 — migrate the existing `asr.spec.ts`.** The old `/draft`, `/draft/regenerate`, `/draft/publish` routes are 404 now
   (see `backend/openapi/openapi.yaml`: `GET /api/v1/admin/lessons/{lessonId}/drafts`, `POST .../drafts/{trackId}/regenerate`,
   `GET .../transcripts`). Rewrite helpers: `readDrafts(lessonID)` (validated parse, no `as`), `draftFor(lessonID, trackId)`,
   tab helpers scoped to the selected tab panel (`tabpanel` of the selected tab), `waitForDraft(panel, state)`. Per the plan row:
   C11 → "the hint comes from the track language" (Regenerate sends no body; there is no recognition-language Select any more — verify
   in the UI); C17/C18 helpers scoped to the selected track tab, C18 without the language Select; S6 "default change" → "every Ready
   track has its own draft" and "renaming a track makes its draft stale, Regenerate re-recognizes" (O7: no automatic replacement);
   C06 → "after replacing the video the old draft is not offered". Keep C13, C03, S6 video-only, adjusted to the new URLs.
2. **M3-R8 + M3-R2 + M3-R4 (success) — the two-track journey** (one new test): lesson with `e2e-speech-640x360.mp4` (embedded `und`,
   "Welcome to the course…") + a dub track `e2e-speech-dub.m4a` with language **`en-GB`**. The S5 picker does not offer `en-GB`
   (checkpoint decision): mint it through the upload API — the cheapest production path is the UI flow with the mint POST rewritten by
   `page.route(... route.continue({ postData }))` to `language: 'en-GB'`, bytes still go through the real upload; or a direct API mint +
   upload if simpler. Record the choice as DECISION. Then exactly the plan's steps: → 2 drafts; `und` draft text contains `Welcome` and
   not `dubbed`; `en-gb` draft contains `dubbed` and not `Welcome`; `und` tab: Save → language dialog → confirm `en` → PATCH then PUT
   (assert the PUT body `language: 'en'` from the network request, M3-R2); `en-gb` tab: Load → Save → PUT `language: 'en-gb'`;
   inventory `[en, en-gb]`; track now `en`; `GET transcript?language=en` and `?language=en-gb` return that language with segment texts
   equal to the corresponding PUT bodies (a missing transcript is also 200 — compare content, not status).
3. **M3-R6 two-tab reload (checkpoint decision: here, not transcript.spec).** On a two-track fixture (reuse the M3-R8 lesson inside the
   same test if it stays readable, else its own test): type different unsaved lines in both tabs, blur, reload → each tab offers its own
   lines only (assert the IndexedDB keys `user:lesson:track:<id>` exist, via `indexedDB` in `page.evaluate`, like `transcript.spec`);
   Restore in one, Discard in the other → re-entry offers nothing for the discarded one, restored one still offered until Save.
4. **M3-R4 failure branch e2e:** `und` tab, dialog confirm → PATCH ok, PUT 503 via `page.route` → same tab, new language, lines kept;
   reload → lines offered from `track:<id>`; Restore → Save → PUT 200. (A fixture with one speech track is enough.)
5. **M3-R14:** a track minted with `multilingual=true` (switch in the new-track form) → its draft becomes ready.

Cleanup: every course created is pushed to `courses` and deleted in the existing `afterAll` (globally counted). Clear IndexedDB keys
your test planted. `transcript.spec.ts` stays unchanged (M3-R1).

## Done = report
`cd frontend && npm run format && npm run typecheck && npm run lint` exit codes; each test you touched run targeted against the stack:
title → passed/failed + duration (copy the reporter line). A test you could not make pass is NOT reported as done: name it, the
failing assertion and the evidence path. Changed files list.
