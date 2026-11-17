# MS2c — run 2 brief for Codex: frontend compile-through after `make generate`

Repository: current directory, branch `ms2-3-asr-per-track` @ `6d11e76` + uncommitted MS2c backend (run 1,
`docs/multi-subtitles/_wip-ms2c-impl.md`) + Courses 1's `make generate` (generated client updated).
Shared working tree: no git checkout/reset/stash/clean/rm, no commits; edit only frontend files named
below and their tests.

WORKING FILE: docs/multi-subtitles/_wip-ms2c-fe.md
Append AFTER EVERY STEP (append, not rewrite). First entry before the first tool: a `NEXT:` line.
Then FACT / DECISION / NEXT / QUESTION. Final answer = concise summary of the file.

Read `AGENTS.md` (TypeScript rules: strict, validate external data, no fetch in UI) and
`docs/multi-subtitles/plan-ms2-3.md` MS-D5, MS-D10. This run is a TEMPORARY bridge: the language tabs
(MS3a) replace it. Keep the change minimal and keep current UI behaviour for the default track.

## 1. Draft adapter (`frontend/src/modules/transcript/adapters/draft-repository.ts`, port, domain)
`npx tsc --noEmit` now fails there (getLessonDraft/regenerateLessonDraft removed). Bridge:
- read = `getLessonDrafts`; the server orders the list default track first; take `drafts[0]` (none →
  the existing "no draft" result). Validate the WHOLE new response strictly (every item, new keys
  `trackId`, `trackLanguage`, `trackMultilingual`) — an unknown key or bad item is a schema failure as
  today, not silently dropped.
- regenerate = `regenerateLessonDraft(lessonId, trackId)` with the trackId of the draft shown; no body.
  The domain/port carry `trackId` as needed. If there is no draft there is nothing to regenerate (O7:
  tracks without a draft are recognised automatically) — keep the button's current behaviour for that
  case only if it compiles naturally; otherwise disable the action and note it as a DECISION.
- Map the new error code `track_not_found` (404) into the draft error union with a sensible existing
  message bucket; do not invent new UI copy beyond one short string if a message map requires it.
- `LessonDraftField.tsx` drops the language Select of Regenerate (MS-D4: Regenerate takes no language);
  remove its now-dead state and tests.

## 2. Audio-track schema (`frontend/src/modules/media/adapters/audio-track-schema.ts` + domain type)
The schemas are `.strict()`: the backend now returns `multilingual` on every track and
`multilingualDefault` on the list, so EVERY list/rename response would fail parsing at runtime while
typecheck stays green. Add `multilingual: z.boolean()` to the track and `multilingualDefault: z.boolean()`
to the list, carry them in the media domain types and the list result (the UI switch is MS3; nothing
renders them yet). Grep for every other strict schema/hand validator that parses a response whose DTO
changed in run 1 (audio tracks list, rename/PATCH response, mint response, drafts, anything else in
`git diff backend/openapi/openapi.yaml`) and fix each; list them in the working file.

## 3. Tests
Update/adjust unit tests of the touched adapters, fakes and `admin-draft-queries` so `npm test` passes;
add: drafts list → first item used; strict rejection of an unknown key in a drafts item; audio list with
`multilingual`/`multilingualDefault` parses; list missing `multilingual` is rejected.

## Done = report
`cd frontend && npm run format && npm run typecheck && npm run lint && npm test` results; changed files.
Do not touch generated files or e2e. If the bridge would need more than a minimal change, QUESTION.
