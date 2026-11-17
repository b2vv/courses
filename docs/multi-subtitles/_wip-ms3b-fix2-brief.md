# MS3b — fix run 2 brief for Codex (review w2 BLOCK)

Repository: current directory, branch `ms2-3-asr-per-track` @ `9ad04e3` + the uncommitted MS3b runs 1, 2 and fix run 1 (logs
`docs/multi-subtitles/_wip-ms3b-run1.md`, `_wip-ms3b-run2.md`, `_wip-ms3b-fix.md`). Shared working tree: do NOT run git
checkout/reset/stash/clean/rm, do not commit, do not touch generated files, e2e, or backend.

WORKING FILE: docs/multi-subtitles/_wip-ms3b-fix2.md
Append AFTER EVERY STEP (append, not rewrite). First entry before the first tool: a `NEXT:` line.
Then FACT / DECISION / NEXT / QUESTION. Final answer = concise summary of the file.

Review: `docs/multi-subtitles/_wip-ms3b-review-codex-w2.md`. Each fix gets a test that FAILS without it (revert-check, record both).

## Decision (2026-11-16, Courses 1) — replaces fix-run-1 "unresolved own offer blocks capture"
Fix run 1 made an unresolved own offer silently refuse capture. That loses author edits made while the offer is pending (w2
Blocker 1) and does not cover a write before the initial read (w2 Blocker 2). Rule instead — **the editor is not editable while its
own recovery state is unknown or unresolved**:
- Until the tab's initial history read has resolved, and while the tab's OWN offer (`offer` / `offer-stale`) is shown, the line fields
  are read-only and the line actions (Add, Remove, Save, Load draft into this editor) refuse, with one short line next to the offer:
  uk «Спершу відновіть або відкиньте незбережені зміни» / en "Restore or discard the unsaved changes first". Use `isReadOnly` /
  `aria-disabled` + a guard in the handler — NOT `isDisabled` on a control that may be focused (RAC drops focus to `<body>`). Fields stay
  focusable and readable (keyboard, screen reader).
- Restore → lines applied, editing enabled. Discard → editing enabled, dismissal as today.
- With no author edits possible during that window, capture needs no special "unresolved offer" refusal any more: remove the
  fix-run-1 write/clear suppression that existed only for this, keep what is still needed: server/baseline updates never arm capture;
  no clear before the initial read resolved; a clean Save while an own offer is pending cannot happen (Save refuses).
- Detached offers (other keys) do NOT lock the editor: capture writes under the tab's own key, never theirs.
- A failed read (store throws) = no snapshot = editing enabled at once (best-effort store must never lock the editor permanently).
- If the read is slow, the lock lasts only until it resolves; do not add a timeout.

## Decision (2026-11-16, Courses 1) — w2 Major 3 is NOT changed
MS-D8 "ключі відновлення не чіпаються" on a PATCH failure means: no clear, no finalize, no move of lines to another key. The ordinary
capture of the lines still on screen under the tab's OWN key is the same write a blur makes; without it the failed save leaves the
lines out of the store (proved in run 2). Keep the current behaviour and its test; rename the test so its name states this rule.

## Tests (name with the row id / "MS3b fix2")
- Pending own offer: typing/Add/Remove/Save refused, the hint shown, fields focusable; Restore enables editing; Discard enables
  editing; then edit → blur → unmount → remount → the NEW lines are offered (w2 Blocker 1 scenario survives).
- Deferred initial read (store.read resolves later): editing refused until it resolves; then (a) a stored snapshot → offer shown, still
  locked; (b) none → editable; no write happens before the read resolves (w2 Blocker 2).
- store.read rejects → editable at once.
- Detached offer pending, no own offer → editable; capture writes only the own key.
- Existing fix-run-1 tests that asserted the suppression are rewritten to the new rule, not deleted silently (list them).

## Done = report
`cd frontend && npm run format && npm run typecheck && npm run lint && npm test` — exit codes and counts; changed files; the
revert-check per fix.
