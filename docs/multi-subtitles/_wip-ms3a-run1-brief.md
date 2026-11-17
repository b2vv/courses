# MS3a — run 1 brief for Codex: frontend data layer, per-track draft queries, "Multiple languages" switch

Repository: current directory, branch `ms2-3-asr-per-track` @ `a985543` (MS2 backend committed; the generated client
already has `getLessonDrafts`, `regenerateLessonDraft(lessonId, trackId)`, `getLessonTranscripts`, media `multilingual`).
Shared working tree with other sessions: do NOT run git checkout/reset/stash/clean/rm, do not commit, do not touch
generated files, e2e, or backend. Courses 1 runs the browser, gates, review and commits.

WORKING FILE: docs/multi-subtitles/_wip-ms3a-run1.md
Append AFTER EVERY STEP (append, not rewrite). First entry before the first tool: a `NEXT:` line.
Then FACT / DECISION / NEXT / QUESTION. Final answer = concise summary of the file.

## Read first
- `AGENTS.md` (TypeScript: strict, validate external data, no fetch in presentation/routes, Query for server state;
  React Aria; >=44px controls, visible focus, semantic tokens).
- `docs/multi-subtitles/plan-ms2-3.md` (ACCEPTED): **MS-D5** (DTOs), **MS-D7** guard inputs, **MS-D10** (frontend: switch,
  cache), rows **M3-R5**, **M3-R14**, **M3-R15**; owner decisions **O7**, **O8** (§0).
- `docs/multi-subtitles/report.md` → "MS3a … checkpoint" (what MS3a leaves to MS3b) and the MS2c "temporary bridge" note.
- The bridge you are replacing: `frontend/src/modules/transcript/adapters/draft-repository.ts` (takes `drafts[0]`).

## Scope of this run (data, queries, Audio tracks switch). The language tabs UI is run 2 — do not build it here.
1. **Drafts per track (transcript module).** Domain/port/adapter return the whole list: `listDrafts(lessonId)` →
   `LessonDraft[]` (each with `trackId`, `trackLanguage`, `trackMultilingual`, `state`, `language`, `failure`, `stale`,
   `updatedAt`, `assetId`, `generation`, `segments`), strict validation of every item as today; `regenerate(lessonId,
   trackId)`. Remove the `drafts[0]` bridge.
2. **Inventory (transcript module).** Adapter + port for `GET …/lessons/{id}/transcripts` → `{language, updatedAt, stale,
   isPrimary}[]`, strict zod schema, explicit error codes as the neighbours.
3. **Queries (`app/composition/admin-draft-queries.ts`).**
   - One query per lesson for the drafts list; selectors/helpers give "the draft of track T".
   - **Pending marker per track** (key `lessonId:trackId`): two tracks never share a marker (M3-R5). Keep the existing
     15-minute window and the "marker lives with the factory, not the query cache" reasoning (gotcha: an unobserved
     `setQueryData` entry is GC'd after 5 min).
   - **Remove the stale-waiting poll** (`awaitingReplacement`, "a stale draft is a pending recognition"): under O7 the
     server does not re-recognise a stale draft; it waits for Regenerate. Polling continues only for a requested
     regeneration or a server `running`/claimed state. Unit test: stale ready draft → no refetch interval.
   - Inventory query per lesson (no polling); invalidated after a successful transcript save (find the save mutation).
4. **Cache (MS-D10, M3-R15).** A successful audio-track PATCH (language, label or `multilingual`) invalidates both the
   audio-track list AND the lesson drafts list of the affected lesson (`admin-audio-queries.ts` today invalidates only
   tracks). Also after delete / set-default if those change which tracks exist (check what they do today).
   Export a **pure guard** in the transcript module (no React): `draftLoadable(draft, track)` = `!draft.stale &&
   draft.trackMultilingual === track.multilingual && (draft.trackLanguage === lower(track.language) ||
   isUndetermined(draft.trackLanguage))` — the `und` exception covers language only, never the mode. Use the existing
   `und`-family helper if one exists (search `und`); unit-test every branch. Run 2 wires it into the UI.
5. **"Multiple languages" switch (M3-R14, `modules/media/ui/LessonAudioTracksField.tsx`).**
   - New-track form: a switch next to the language picker, initial value = list `multilingualDefault`; mint sends
     `multilingual`.
   - Each existing track row: a switch showing `track.multilingual`; changing it sends PATCH with the row's current
     `language` and `label` plus `multilingual` (the body stays complete, MS-D10). Pending/error handling as the row's
     other actions; while pending use `isPending`/`aria-disabled`, NOT `isDisabled` on a focused control (gotcha: RAC
     `isDisabled` drops focus to `<body>`).
   - Use the project's existing switch/checkbox primitive (search `ui/` for `Switch`/`Checkbox`); do not add a library.
     Accessible name "Multiple languages" via i18n (add keys in every locale file the field already uses; copy must be
     short, e.g. uk «Кілька мов», en "Multiple languages"), a one-line description that it makes recognition detect
     language per passage, >=44px hit target, keyboard (Space toggles), visible focus.
6. **Compile-through.** Whatever still uses the bridge (`LessonDraftField`, `CourseEditorView`, route wiring) keeps
   compiling and showing the **default track's** draft (the first item; server order is default first), with
   Regenerate on that track — run 2 replaces this with tabs. Mark the spot with `// MS3a run 2: tabs`.

## Tests (rstest; update the existing ones you break)
- adapter: list parse; unknown key in an item → schema failure; inventory parse/reject.
- queries: two tracks → two markers; stale → no poll; regenerate of T polls until T is claimed, not when U changes.
- audio: PATCH success invalidates drafts of that lesson; switch initial value from `multilingualDefault`; switch change
  sends complete PATCH body; mint sends `multilingual`.
- guard: every branch incl. `und` + mode mismatch → not loadable.

## Done = report
`cd frontend && npm run format && npm run typecheck && npm run lint && npm test` — exit codes and test counts; changed
files. If the plan is ambiguous or would cause a bug, write QUESTION and argue instead of implementing around it.
