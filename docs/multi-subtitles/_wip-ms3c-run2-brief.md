# MS3c run 2 brief for Codex — e2e of the MS3 rows, part 2

Repository: current directory, branch `ms2-3-asr-per-track` @ `7b20d2c` + uncommitted run 1 (`frontend/e2e/asr.spec.ts`, journal
`docs/multi-subtitles/_wip-ms3c-run1.md`) + the Courses 1 fix of your run-1 QUESTION: the track edit dialog's `<form>` is portaled but
React bubbled its submit into the course editor's `<form>` → extra `PUT /admin/courses/{id}` + tab reset. Fixed by `stopPropagation`
in `frontend/src/modules/media/ui/LessonAudioTracksField.tsx` with a unit test; the frontend image is rebuilt. Shared working tree: do
NOT run git checkout/reset/stash/clean/rm, do not commit, do not touch generated files or backend. Product code only as in run 1: a
proven defect → STOP that test, QUESTION with evidence, continue the others.

WORKING FILE: docs/multi-subtitles/_wip-ms3c-run2.md
Append AFTER EVERY STEP (append, not rewrite). First entry before the first tool: a `NEXT:` line.
Then FACT / DECISION / NEXT / QUESTION. Final answer = concise summary of the file.

Read: `docs/multi-subtitles/plan-ms2-3.md` §4 rows M3-R10..R13, R15 (exact steps); the MS3c checkpoint in `report.md`; your run-1
journal (runtime commands, harness lessons: localized names from `uk.ts`, `en-GB` casing, the read route is
`GET /api/v1/learn/lessons/{id}/transcript`, fresh course token before cleanup, Switch via focus + Space). Runtime and cleanup rules
are the same as run 1 (`courses-kiri` prefix, `--project=chromium -g`, own `--output=test-results/ms3c-run2-*`).

## Scope
0. **Guard the class just fixed:** the rename helper (and every UI track rename in the spec) asserts that no `PUT /api/v1/admin/courses/`
   request was sent during the rename and that the selected transcript tab stays the same.
1. **M3-R10 e2e:** (a) Load a draft, rename the track via `page.request` PATCH (not the UI) before Save → Save refuses with the message,
   0 `PUT .../transcript`; (b) `page.route` 503 on GET audio tracks at Save → retryable error, lines kept, 0 PUT; unroute → Save → 200.
2. **M3-R11 (a)→(d)** — separate one-track lesson (embedded `und`, speech), steps sequential on one track exactly as the plan row.
   Assert each PUT body's language and new segment ids (ids differ from the previous PUT's ids), and that the prior languages'
   transcripts are unchanged (read route, compare texts). (e) is not in scope (unit + M3-R12b).
3. **M3-R12 (a, b):** (a) type lines while the first video is still processing (no blur, no reload), wait for Ready → lines offered in
   the default tab; (b) type in a track tab, upload a replacement video → after promotion the lines are offered in the tab of the same
   language, else default. Use `pauseRecognition`/existing helpers only if needed to keep a window open; resume in `finally`.
4. **M3-R13 (a, b):** (a) two editors of the same language — open `en` in the page (T1), `page.request` PUT `en` (T2), Save in the page →
   `transcript_modified` message, server has the T2 lines, page keeps its lines; (b) restored snapshot with base T1 after a foreign T2
   save → Save → `transcript_modified`.
5. **M3-R15:** ready draft → toggle "Multiple languages" of that track in the open editor → Load disabled with the explanation, no
   automatic replacement → Regenerate → new draft → Load enabled; the same for an `und` draft.
6. **Full file once:** after your tests pass targeted, run the whole `e2e/asr.spec.ts` once (`--project=chromium`, no `-g`) and copy the
   summary line + every reporter line.

## Done = report
`cd frontend && npm run format && npm run typecheck && npm run lint` exit codes; per test: title → passed/failed + duration; the
full-file summary; changed files. A test you could not make pass is NOT done: name it, the failing assertion, the evidence path.
