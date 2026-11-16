# MS3b — fix run 6 brief for Codex (review w6: 1 Major, 1 Minor) — close the CLASS, not the instance

Repository: current directory, branch `ms2-3-asr-per-track` @ `9ad04e3` + the uncommitted MS3b runs 1, 2, fix runs 1–5 (logs
`docs/multi-subtitles/_wip-ms3b-{run1,run2,fix,fix2,fix3,fix4,fix5}.md`). Shared working tree: do NOT run git
checkout/reset/stash/clean/rm, do not commit, do not touch generated files, e2e, or backend.

WORKING FILE: docs/multi-subtitles/_wip-ms3b-fix6.md
Append AFTER EVERY STEP (append, not rewrite). First entry before the first tool: a `NEXT:` line.
Then FACT / DECISION / NEXT / QUESTION. Final answer = concise summary of the file.

Review: `docs/multi-subtitles/_wip-ms3b-review-codex-w6.md`. Each fix gets a test that FAILS without it (revert-check, record both).

## 1. Major — and the whole class: no recovery deletion without version ownership
Rounds 4, 5 and 6 each found one more place where a snapshot key is deleted without checking that the stored version is the one this
editor saw or wrote (compare-delete, detached Discard, now the clean capture after own-offer Discard: `use-transcript-recovery.ts` ~119).
Fix the class:
- Inventory EVERY call path that removes a snapshot in the transcript module (`store.clear`, `clearIfVersion`, finalize after Save,
  clean-capture "clear" directive, Discard own/detached, accepted-source retirement, viewer sweep `forgetOtherAuthors`). List them in the
  working file with file:line and the rule each now follows.
- Rule: a deletion of a key that may be written by another window happens only via `clearIfVersion(key, version)` where `version` is the
  `capturedAt` this editor instance last WROTE or was OFFERED/READ for that key. If the editor knows no version (nothing was stored when it
  read, and it has not written), it does not delete. On mismatch: keep it and, where an offer can be shown, offer the newer version.
- Exceptions, each justified in the working file: `forgetOtherAuthors` (different user — the sweep is the point); finalize after a
  successful Save (decide: should a newer capture from another window made after this editor's last write survive a Save here? MS-D9 says a
  key is deleted after Save of the lines it holds — so use the version this editor last wrote; a newer foreign version survives and is
  offered next visit). State the decision.
- Tests: the review's two-window scenario through blur and remount → C kept and offered; a table-driven test over the inventoried paths
  (each path × {same version → deleted, newer version → kept}).

## 2. Minor — async Discard completion steals focus (`TranscriptEditor.tsx` ~195, ~274)
Move focus after Restore/Discard only if focus is still on the initiating recovery control or on `<body>` (the control disappeared).
Test: deferred store, Discard, focus another control before completion → focus stays there.

## Done = report
`cd frontend && npm run format && npm run typecheck && npm run lint && npm test` — exit codes and counts; changed files; revert-check per fix.
