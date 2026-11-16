# Independent review — MS3b (read-only)

You are an independent reviewer. Do NOT edit any file except your own working file; do not run git checkout/reset/stash/clean/rm.
WORKING FILE: docs/multi-subtitles/_wip-ms3b-review-codex-w7-log.md — append after every step (NEXT / FACT / FINDING).

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

## Round 3
Round 2 (`_wip-ms3b-review-codex-w2.md`) was BLOCK (2 Blocker, 1 Major). Fix run 2 (`_wip-ms3b-fix2-brief.md`, log `_wip-ms3b-fix2.md`)
replaces the "unresolved own offer refuses capture" rule with "the editor is read-only until the initial read resolves and while its own
offer is unresolved". Major 3 (capture after a PATCH failure) is deliberately kept — the brief records the decision and its reason;
challenge it only if it causes a concrete defect. Check: the lock can never become permanent (read failure, offer dismissed in another
mount, stale offer); focus is not lost when the lock starts/ends; Load draft / Regenerate / dialogs while locked; the new "successful-
capture dismissal renewal" mechanism (fix2 log) — is it needed and correct? Review the WHOLE MS3b diff again.

## Round 4
Round 3 (`_wip-ms3b-review-codex-w3.md`) was BLOCK (1 Blocker: Save deletes a reactivated source key holding newer lines). Fix run 3
(`_wip-ms3b-fix3-brief.md`, log `_wip-ms3b-fix3.md`) adds version-checked deletion with relinquishment, and focus after Restore/Discard.
Verify the scenario is closed, and that the compare-delete never deletes a version the author has not saved nor discarded and never leaves
a saved snapshot to be offered forever in a way that misleads (if it can, how and how often). Check the focus effect cannot steal focus
from elsewhere (other tab, dialog, another editor) on unrelated re-renders. Review the WHOLE MS3b diff again; report only real defects.

## Round 7
Rounds 1–6 were BLOCK; every finding was fixed. Round 6 found one more unversioned deletion; fix run 6 (`_wip-ms3b-fix6-brief.md`, log
`_wip-ms3b-fix6.md`) closes the CLASS: every snapshot deletion path is inventoried in the log with its rule (versioned `clearIfVersion` with
the version this editor read/offered/wrote; unknown version → no delete; viewer sweep unconditional by design), plus focus-after-async.
Also note its decisions: own Discard now deletes the same version (Phase 8 only remembered a dismissal, but a clean blur cleared it
anyway); a failed Save no longer re-captures after the response. Verify the inventory is complete (grep every store mutation) and each
rule holds. Severity discipline: Blocker/Major only for a scenario a real author reaches in normal use of one or two windows, with steps
and data lost; "offered again, nothing lost" or rare interleavings are Minor with mechanism. Review the WHOLE MS3b diff again.
