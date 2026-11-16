# MS3b — fix run 3 brief for Codex (review w3 BLOCK + one browser finding)

Repository: current directory, branch `ms2-3-asr-per-track` @ `9ad04e3` + the uncommitted MS3b runs 1, 2, fix runs 1–2 (logs
`docs/multi-subtitles/_wip-ms3b-{run1,run2,fix,fix2}.md`). Shared working tree: do NOT run git checkout/reset/stash/clean/rm, do not
commit, do not touch generated files, e2e, or backend.

WORKING FILE: docs/multi-subtitles/_wip-ms3b-fix3.md
Append AFTER EVERY STEP (append, not rewrite). First entry before the first tool: a `NEXT:` line.
Then FACT / DECISION / NEXT / QUESTION. Final answer = concise summary of the file.

Each fix gets a test that FAILS without it (revert-check, record both). Keep everything else as it is.

## 1. Blocker (review w3, `docs/multi-subtitles/_wip-ms3b-review-codex-w3.md`)
Saving a restored detached snapshot deletes another editor's newer unsaved lines. `acceptedDetached` remembers only the source KEY;
a successful Save clears whatever is under that key now, even after the key became a mounted editor's OWN key again (orphan tab
reappeared after a rename back) and holds that editor's newer capture C.
Fix (MS-D9: a detached key is deleted only after a Save of the tab that accepted IT, i.e. those exact lines):
- remember the accepted snapshot's identity (key + `capturedAt`, or key + a content check — choose and justify), and on successful Save
  delete the key only if the store still holds that same version (read-compare-delete; best-effort: a failed read = do not delete);
- if the source key is the own key of any currently mounted tab, relinquish: never delete it from another tab.
Test: the full scenario from the review — orphan `lang:uk` holds A → rename a track to `uk` → restore detached A in the track tab →
rename back to `en` (orphan reappears) → in the orphan restore, edit to C, blur → Save the track (confirm replacement) → remount → the
orphan offers C. Plus a unit test of the compare-delete (same version → deleted; newer version → kept; read fails → kept).

## 2. Focus after Restore / Discard (Courses 1, real Chromium on Kiri, 2026-11-16)
The recovery offer's Restore and Discard buttons disappear with the offer, so focus falls to `<body>` (seen live after keyboard Discard).
Since fix run 2 the author MUST resolve the offer before editing, so this is on the main path. After Restore or Discard of the tab's own
offer or a detached offer, move focus to a stable target inside the same editor: the next offer's first button if another offer follows,
otherwise the first line's text field if there are lines, otherwise "Add a line". Use a ref + `focus()` after the state update (effect),
not autofocus on mount. Test: keyboard Discard / Restore → `document.activeElement` is the expected element (both own and detached, and
the "next detached offer follows" case).

## Done = report
`cd frontend && npm run format && npm run typecheck && npm run lint && npm test` — exit codes and counts; changed files; revert-check per fix.
