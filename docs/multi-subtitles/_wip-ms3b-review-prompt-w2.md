# Independent review — MS3b (read-only)

You are an independent reviewer. Do NOT edit any file except your own working file; do not run git checkout/reset/stash/clean/rm.
WORKING FILE: docs/multi-subtitles/_wip-ms3b-review-codex-w2-log.md — append after every step (NEXT / FACT / FINDING).

Snapshot: branch `ms2-3-asr-per-track` @ `9ad04e3` + the uncommitted frontend diff (`git diff -- frontend` and the untracked
files `frontend/src/modules/transcript/adapters/transcript-history-indexed-database.test.ts`,
`frontend/src/modules/transcript/ui/use-detached-transcript-recovery.ts`, `frontend/src/modules/transcript/ui/use-transcript-recovery.test.tsx`).
Ignore every other dirty/untracked file (other sessions' WIP).

Authoritative behaviour: `docs/multi-subtitles/plan-ms2-3.md` (ACCEPTED) — MS-D6 (token/base pair, Save re-read, new ids),
MS-D8 (`und` dialog: options, preselection, PATCH → PUT, failure branches), MS-D9 (key per tab, detached snapshots, offer rules,
delete only after Save/Discard, flush on unmount); rows M3-R4, M3-R6, M3-R12, M3-R13. Repository rules: `AGENTS.md`.
Briefs: `docs/multi-subtitles/_wip-ms3b-run{1,2}-brief.md`.

Look for real defects: lost unsaved lines (remount, unmount, failed save, query loading), a PUT or PATCH that should not happen
(count them), a wrong token (base token reaching another language; legacy snapshot keeping a token), a detached key deleted
too early or a write under a foreign key, prefix matching another lesson, duplicate DOM ids across mounted panels, focus loss
(RAC isDisabled on a focused control), misleading copy (e.g. which message the author sees after Cancel of the language dialog),
hooks rules, architecture boundaries (admin imports only modules' public index.ts), tests that would pass with the bug present.

Output: verdict SHIP or BLOCK; findings ranked Blocker/Major/Minor with file:line, failure scenario, and a suggested fix.

## Round 2
Round 1 (`docs/multi-subtitles/_wip-ms3b-review-codex-w1.md`) was BLOCK (1 Blocker, 1 Major, 1 Minor). The fix run
(`docs/multi-subtitles/_wip-ms3b-fix-brief.md`, log `_wip-ms3b-fix.md`) addresses them plus four browser findings (Escape focus, an
orphan-tab notice, replace-dialog values, a placeholder). Verify each fix really closes its scenario, and look for NEW defects the fixes
introduced (especially the "unresolved own offer blocks capture" rule: can it lose new author edits made while an offer is pending?
can a pending offer block capture forever?). Review the WHOLE MS3b diff again, not only the fix delta.
