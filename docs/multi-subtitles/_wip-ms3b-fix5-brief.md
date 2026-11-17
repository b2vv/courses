# MS3b — fix run 5 brief for Codex (review w5: 1 Major, 1 Minor)

Repository: current directory, branch `ms2-3-asr-per-track` @ `9ad04e3` + the uncommitted MS3b runs 1, 2, fix runs 1–4 (logs
`docs/multi-subtitles/_wip-ms3b-{run1,run2,fix,fix2,fix3,fix4}.md`). Shared working tree: do NOT run git checkout/reset/stash/clean/rm,
do not commit, do not touch generated files, e2e, or backend.

WORKING FILE: docs/multi-subtitles/_wip-ms3b-fix5.md
Append AFTER EVERY STEP (append, not rewrite). First entry before the first tool: a `NEXT:` line.
Then FACT / DECISION / NEXT / QUESTION. Final answer = concise summary of the file.

Review: `docs/multi-subtitles/_wip-ms3b-review-codex-w5.md`. Each fix gets a test that FAILS without it (revert-check, record both).

## 1. Major — Discard deletes newer edits from another window (`TranscriptEditor.tsx` ~241)
Discard of a DETACHED offer (and of the tab's own offer if it is the same issue — check) uses unconditional `clear(key)`. Use
`clearIfVersion(key, offered.capturedAt)`. On mismatch (a newer version is stored): do not delete; refresh the offer from the store so
the author sees the newer version (it is offered again), and do not mark the key as handled/dismissed. Own-offer Discard today only
remembers a dismissal (no delete) — keep that unless you find the same loss there. Test: the review's two-editor scenario → C kept and
offered; plus the plain case (same version → deleted, next offer shown, focus as fix run 3).

## 2. Minor — cleanup ancestry lost through a second detachment (`TranscriptEditor.tsx` ~240)
Implement ONLY if it stays small: `acceptedSource` becomes a bounded list of `{key, capturedAt}` (the restored snapshot itself plus its own
`acceptedSource` chain, deduplicated, capped, e.g. 8), validated by the adapter (accept the old single-object shape too); Save retires each
via `clearIfVersion`. If it needs more than that (new stores, cross-tab coordination), do NOT implement it — write a DECISION with the
mechanism and stop; Courses 1 records it as a limitation. Test if implemented: the review's six-step scenario → A retired after Save.

## Done = report
`cd frontend && npm run format && npm run typecheck && npm run lint && npm test` — exit codes and counts; changed files; revert-check per fix.
