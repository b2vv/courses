# MS3b — run 1 brief for Codex: `und` confirm dialog (PATCH → PUT) and `baseLanguage` in the snapshot

Repository: current directory, branch `ms2-3-asr-per-track` @ `9ad04e3` (MS3a committed: language tabs, per-track drafts,
`draftLoadable`, the "Multiple languages" switch; read `docs/multi-subtitles/report.md` → "MS3a" for what exists). Shared
working tree with other sessions: do NOT run git checkout/reset/stash/clean/rm, do not commit, do not touch generated files,
e2e, or backend.

WORKING FILE: docs/multi-subtitles/_wip-ms3b-run1.md
Append AFTER EVERY STEP (append, not rewrite). First entry before the first tool: a `NEXT:` line.
Then FACT / DECISION / NEXT / QUESTION. Final answer = concise summary of the file.

## Read first
- `AGENTS.md` (strict TS, validate external data, no fetch in presentation/routes, React Aria, >=44px, visible focus).
- `docs/multi-subtitles/plan-ms2-3.md` (ACCEPTED) — authoritative: **MS-D8** (all of it), the token/base part of **MS-D6**
  ("Токен — оптимістичний замок…": `baseLanguage` in the snapshot), rows **M3-R4** and **M3-R13** (branches e, f, g at unit level).
- `docs/multi-subtitles/report.md` → "MS3b … checkpoint".
- Code: `frontend/src/modules/admin/ui/LessonLanguageTabs.tsx` (today `isUndetermined(language)` → `'track_language_required'`,
  the MS3a interim rule you replace), `modules/transcript/application/recovery.ts`, `ui/use-transcript-recovery.ts`,
  `ui/TranscriptEditor.tsx`, `modules/media/ui/LessonAudioTracksField.tsx` (`CourseEditorAudio.actions.rename(track, language,
  label, multilingual)` — the existing PATCH), `audioTrackLanguages` (S5 list).

## Scope
1. **`und` tab Save = confirm dialog (MS-D8).** In a track tab whose current track language is in the `und` family, Save opens
   a React Aria modal dialog (same `Aria.ModalOverlay` pattern as the replace dialog in the same file; focus enters the dialog,
   Escape/Cancel closes and returns focus to Save):
   - Options: the S5 languages minus `und`, minus languages already on the asset's tracks (case-insensitive), minus languages that
     already have a transcript in the inventory. Use FRESH reads (the same direct `readTracks` / `readInventory` the save path
     already does), not the Query cache. A failed read → the existing `track_verification_failed` message, no dialog, 0 PATCH/PUT.
   - Preselected: this track's draft `language` (whisper-detected) if it is among the options; otherwise nothing selected and the
     confirm button refuses until a language is chosen (say so, do not disable a focused button — RAC `isDisabled` drops focus).
   - Hint text in the dialog: uk «Щоб прив'язати наявний транскрипт, перейменуйте доріжку в Audio tracks», en "To attach an existing
     transcript, rename the track in Audio tracks".
   - No options left → the dialog says so and offers only Cancel.
   - Confirm: (1) `audio.actions.rename(track, chosen, track.label, track.multilingual)` (PATCH; body stays complete);
     (2) on success, PUT the transcript in `chosen` WITHOUT a token, new `crypto.randomUUID()` ids as every PUT.
   - PATCH fails (`duplicate_language` 409 or any error): a named error message in the editor, PUT not sent, lines kept, recovery
     keys untouched (no write/clear/finalize).
   - PATCH ok + PUT fails: the tab identity stays `track:<id>` (no remount), the tab's language becomes `chosen` once the tracks
     query refetches, lines stay in React state; the next Save is the ordinary PUT path of that language (with MS-D6 re-read), not
     the dialog again. Make sure the editor's save language follows the new track language without a remount (MS-D6 "Зміна мови T").
   - Cancel → 0 PATCH, 0 PUT, lines kept.
   - The orchestration (dialog result → PATCH → PUT) lives in admin UI with injected functions, not in presentation fetching.
   - Pass whatever `LessonLanguageTabs` needs from `CourseEditorView`/route wiring (the audio actions are already available there
     as `video.audio.actions`).
2. **`baseLanguage` in the snapshot (MS-D6, M3-R13).**
   - `TranscriptSnapshot` and `TranscriptBaseline` gain optional `baseLanguage`. Capture writes `baseLanguage` (the editor base
     language) together with the existing `expectedUpdatedAt`; `language` stays the save language at capture time.
   - Restore sets the editor base to the pair (`snapshot.baseLanguage`, `snapshot.expectedUpdatedAt`) — today restore only sets the
     token (`TranscriptEditor.tsx` ~233–244). A re-capture after restore writes the same pair.
   - A snapshot WITHOUT `baseLanguage` (all existing ones): owner of the token unknown → the save takes the "save language ≠ base
     language" branch: no base token; L absent → PUT without token; L present → the replace confirmation with L's token. Implement
     this by making the restored base language a value that never equals a real language (document how), not by special-casing
     in several places.
   - IndexedDB adapter: no schema/version change; the field is optional on read. Check the adapter's read validation accepts and
     returns it (and still accepts snapshots without it).

## Tests (rstest + Testing Library; update what you break; name tests with the row id)
- M3-R4: preselection from draft language / none when not an option; options exclude `und`, track languages (case-insensitive),
  inventory languages; read failure → no dialog, 0 PATCH, 0 PUT; Cancel → 0 PATCH, 0 PUT; success → exactly PATCH then PUT (order,
  PUT language = chosen, no token, fresh ids); `duplicate_language` → error shown, 0 PUT, recovery store untouched (no write/clear
  calls after the failure); PATCH ok + PUT fails → same tab mounted (no remount: a typed line survives), next Save after the tracks
  refetch with the new language = ordinary PUT, no dialog; no options → only Cancel. Rewrite the MS3a test "`und` tab Save → 0 PUT".
- M3-R13 e/f/g (unit): capture writes `baseLanguage`; restore of (en, T1) into a tab now `es`, `es` absent → PUT `es` without token;
  `es` present → confirmation with `es` token; snapshot without `baseLanguage` → no base token even if save language equals the
  snapshot language.
- Dialog a11y: accessible name (heading), focus moves into the dialog and returns to Save on Cancel.

## Done = report
`cd frontend && npm run format && npm run typecheck && npm run lint && npm test` — exit codes and counts; changed files.
If the plan is ambiguous or would cause a bug, write QUESTION and argue instead of implementing around it. If the slice is too big
for one run, finish the implementation with passing existing tests first, list the missing tests as NEXT, and stop.
