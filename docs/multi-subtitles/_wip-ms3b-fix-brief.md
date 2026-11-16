# MS3b — fix run brief for Codex (review w1 BLOCK + Courses 1 browser findings)

Repository: current directory, branch `ms2-3-asr-per-track` @ `9ad04e3` + the uncommitted MS3b runs 1–2 (read
`docs/multi-subtitles/_wip-ms3b-run1.md`, `_wip-ms3b-run2.md`). Shared working tree: do NOT run git checkout/reset/stash/clean/rm,
do not commit, do not touch generated files, e2e, or backend.

WORKING FILE: docs/multi-subtitles/_wip-ms3b-fix.md
Append AFTER EVERY STEP (append, not rewrite). First entry before the first tool: a `NEXT:` line.
Then FACT / DECISION / NEXT / QUESTION. Final answer = concise summary of the file.

Authoritative: `docs/multi-subtitles/plan-ms2-3.md` MS-D6, MS-D8, MS-D9; rows M3-R4, M3-R6, M3-R12, M3-R13.
Each fix gets a test that FAILS without the fix (check it: revert the fix locally, run the test, restore; record both results).

## Review w1 (`docs/multi-subtitles/_wip-ms3b-review-codex-w1.md`)
1. **Blocker — saving B deletes unsaved snapshot A** (`TranscriptEditor.tsx` ~427). Restore detached A, then Restore detached B, then
   Save → both source keys deleted, only B's lines were saved. MS-D9: a detached key is deleted only after a successful Save of the
   tab that accepted IT (its lines are what was saved) or its explicit Discard. Fix: a Restore of B replaces the editor lines, so A is
   no longer "accepted" — it must go back to the offer queue (or stay pending), not be deleted on Save. Test: Restore A → Restore B →
   Save → only B's key cleared, A offered again.
2. **Major — clean rename erases an unresolved recovery snapshot on unmount** (`TranscriptEditor.tsx` ~146,
   `use-transcript-recovery.ts` ~136). An offer is shown and not yet resolved; a server-driven base change (rename) calls `setLines`,
   which arms the idle capture; unmount flushes, `captureDirective` says "clear" and the unresolved snapshot is deleted. Fix: only
   author edits arm capture (a baseline/server update must not), and while the tab's own offer is unresolved nothing may clear its key.
   Test: offer pending → rename (clean editor) → unmount → key still present and offered on remount.
3. **Minor — Cancel of the `und` language dialog shows "Replacement cancelled"** (`LessonLanguageTabs.tsx` ~302). Close without an
   error message, or a language-dialog copy ("Мову не змінено, рядки на місці" / "Language not changed; your lines are still here").
   Assert the message in the Cancel AND Escape tests.

## Courses 1 browser findings (Kiri, real Chromium, 2026-11-16)
4. **Escape loses focus.** Open the `und` language dialog from the keyboard (focus on "Save the transcript", Enter), press Escape →
   `document.activeElement` is `<body>`. A Cancel click returns focus to Save correctly. The unit test claims Escape returns focus,
   so jsdom does not reproduce it — find the real cause (e.g. what re-renders/remounts or what receives focus between the Escape
   keydown and the overlay's focus restore; the Escape may also reach an outer handler) and fix it. Check the replace dialog too.
   If a reliable unit test is impossible, say so explicitly and describe the mechanism; Courses 1 re-checks in the browser.
5. **Wrong copy in an orphan tab.** Orphan tab `lang:uk` → Restore a legacy snapshot (no `baseLanguage`) → the editor shows "The track
   language is now uk; these lines will be saved in uk". There is no track in an orphan tab, and the language did not change; the
   message is driven by base language ≠ save language. Show the "track language changed" notice only in a track tab when the TRACK
   language actually changed under a dirty editor (MS-D6 "Зміна мови T"). After restoring a snapshot whose base differs, the
   save path already explains itself (replace confirmation / tokenless PUT); no notice needed.
6. **Minor — replace dialog shows raw values**: "A transcript in uk (updated 2026-11-16T16:11:47.927508Z) already exists". Show the
   language name (same labels as the tabs) and a localized date-time (`Intl.DateTimeFormat` with the current i18n language).
7. **Minor — language Select placeholder duplicates its label** ("Track language" twice). Placeholder like «Оберіть мову» / "Choose a
   language" (an existing key if one exists).

## Done = report
`cd frontend && npm run format && npm run typecheck && npm run lint && npm test` — exit codes and counts; changed files; for each
of 1–7 the test name and the revert-check result (or why none).
