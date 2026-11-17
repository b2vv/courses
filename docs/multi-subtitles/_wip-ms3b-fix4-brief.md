# MS3b — fix run 4 brief for Codex (review w4 BLOCK)

Repository: current directory, branch `ms2-3-asr-per-track` @ `9ad04e3` + the uncommitted MS3b runs 1, 2, fix runs 1–3 (logs
`docs/multi-subtitles/_wip-ms3b-{run1,run2,fix,fix2,fix3}.md`). Shared working tree: do NOT run git checkout/reset/stash/clean/rm, do not
commit, do not touch generated files, e2e, or backend.

WORKING FILE: docs/multi-subtitles/_wip-ms3b-fix4.md
Append AFTER EVERY STEP (append, not rewrite). First entry before the first tool: a `NEXT:` line.
Then FACT / DECISION / NEXT / QUESTION. Final answer = concise summary of the file.

Review: `docs/multi-subtitles/_wip-ms3b-review-codex-w4.md`. Each fix gets a test that FAILS without it (revert-check, record both).

## 1. Blocker — compare-delete is not atomic (`application/recovery.ts` ~243)
Read and clear are two store operations; another window's capture between them is deleted. Add a port method to
`TranscriptHistoryStore`, e.g. `clearIfVersion(key, capturedAt): Promise<boolean>` (or a better-named equivalent), implemented in the
IndexedDB adapter as ONE `readwrite` transaction (get → compare → delete inside the same transaction) and in the in-memory adapter
synchronously. Use it for every "delete the accepted source" path; keep plain `clear` only where the author explicitly discards or the
editor's own key is finalized. Tests: in-memory interleaving (newer write before the conditional clear → kept); the IndexedDB adapter
test checks the get/delete happen in one transaction (the existing adapter test style with mocked requests is fine; state what it proves
and what it does not).

## 2. Major — cleanup ownership is lost on reload (`TranscriptEditor.tsx` ~111, ~497)
Restore detached A → capture writes A's lines under the tab's own key → reload → restore own offer → Save → own key cleared, the detached
source A stays and is offered again on every visit. Fix: the snapshot carries the accepted source (`acceptedSource?: { key; capturedAt }`,
optional, validated by the IndexedDB adapter like `baseLanguage`) — capture writes it while the editor's lines still come from that
source; restore of the own snapshot re-establishes it; a successful Save retires that exact source via the atomic method from 1. If the
author's lines were replaced since (another Restore, Load draft, Discard), the field is dropped as today's in-memory rule does.
Test: the review's full scenario through Save and two further mounts → A no longer offered; and the variant where the source got a newer
version C in between → C kept and still offered.

## Done = report
`cd frontend && npm run format && npm run typecheck && npm run lint && npm test` — exit codes and counts; changed files; revert-check per fix.
