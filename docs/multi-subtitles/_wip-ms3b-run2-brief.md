# MS3b — run 2 brief for Codex: detached recovery snapshots and flush on unmount

Repository: current directory, branch `ms2-3-asr-per-track` @ `9ad04e3` + the uncommitted MS3b run 1 (the `und` confirm dialog and
`baseLanguage` in the snapshot — read `docs/multi-subtitles/_wip-ms3b-run1.md` for what exists). Shared working tree with other
sessions: do NOT run git checkout/reset/stash/clean/rm, do not commit, do not touch generated files, e2e, or backend.

WORKING FILE: docs/multi-subtitles/_wip-ms3b-run2.md
Append AFTER EVERY STEP (append, not rewrite). First entry before the first tool: a `NEXT:` line.
Then FACT / DECISION / NEXT / QUESTION. Final answer = concise summary of the file.

## Read first
- `AGENTS.md`; `docs/multi-subtitles/plan-ms2-3.md` (ACCEPTED) — authoritative: **MS-D9** (all of it: key, detached snapshots,
  unmount), rows **M3-R6** and **M3-R12** (c + flush at unit level; a/b are e2e in MS3c, but make them possible).
- Code: `modules/transcript/application/recovery.ts` (`transcriptHistoryKey`, `recoveryOffer`), `ports/TranscriptHistoryStore.ts`
  (`keys()` exists), `ui/use-transcript-recovery.ts` (unmount today only `clearTimeout`s, ~89–93), `ui/TranscriptEditor.tsx`
  (recovery offer UI), `modules/admin/ui/LessonLanguageTabs.tsx` (tab identities `track:<id>` / `lang:<lang>`, single editor
  without tabs has key `user:lesson`).

## Scope
1. **Flush on unmount.** `useTranscriptRecovery` on unmount performs the pending idle capture (with the latest lines and baseline)
   instead of dropping it. Best-effort like every store call (errors swallowed). A capture that would be "clear" (lines equal the
   baseline) still clears, as today. Do not capture on unmount when nothing changed since the last capture/mount (no pending timer).
2. **Detached snapshots — pure decision in `application/recovery.ts`.** Input: all store keys + snapshots of this user+lesson
   (prefix exactly `user:lesson` or `user:lesson:`; never match another lesson whose id merely starts with the same characters), the
   current tab set (identity, current language, isDefault). A snapshot is detached when its key is not the key of any current tab:
   legacy `user:lesson` (when tabs exist), `track:<id>` of a track that is gone, `lang:<x>` of an orphan now attached to a track.
   Placement (reading only): the tab whose current language = `snapshot.language` (case-insensitive; a track tab wins over an orphan
   tab of the same language only if both could exist — they cannot by construction, note it); none or no language → the default
   track tab; no tabs at all (single editor) → the single editor takes detached `track:`/`lang:` snapshots by the same rule (default =
   the single editor). Several detached for one tab → offered one at a time, newest `capturedAt` first. Pure and unit-tested.
3. **Offer UI.** The tab's OWN snapshot offer has priority; only when it is `none` (or after it was handled) the next detached one is
   offered. Copy: uk «Незбережені рядки з іншої вкладки (мова: X)» / «… (мова невідома)», and for legacy `user:lesson` «Незбережені
   рядки до появи вкладок (мова: X / невідома)»; en equivalents. Actions: Restore and Discard.
   - Restore → the lines go into this tab's editor; the base becomes the snapshot pair (`baseLanguage`, `expectedUpdatedAt`) exactly as
     run 1's restore does (so the token rules of MS-D6 apply: a different save language never carries the base token). From then on
     the ordinary capture writes under THIS tab's key. Remember the detached key as "accepted by this tab".
   - The detached key is deleted ONLY after a successful Save of the tab that accepted it, or on its explicit Discard. Nothing is ever
     written under a foreign key. A reload before Save offers it again (both the tab's own snapshot and the detached one may then be
     offered — own first).
   - Discard of a detached snapshot = `store.clear(detachedKey)` (best-effort), then the next detached one, if any.
   - Offers stay inside the tab panel; ids built from `lessonId + instanceId` as the rest of the editor (no duplicate ids, M3-R7).
4. **Where the reading lives.** One read of `store.keys()` + the snapshots per lesson in the tabs component (or a hook in the
   transcript module exposed through its `index.ts`), not one per editor; the decision is the pure function from 2. The single editor
   without tabs gets the same treatment. Admin must not import transcript internals other than its public `index.ts`.
5. **Viewer sweep** (`forgetOtherAuthors`) is unchanged; check it still matches the new key forms (`user:lesson:track:<id>`).

## Tests (rstest; name tests with the row id)
- M3-R6: two tabs' snapshots are isolated (writing in A never touches B's key); legacy `user:lesson` snapshot with language `uk` →
  offered in the `uk` track tab; without language → default tab, with the "language unknown" copy; gone-track snapshot → tab of its
  language / default; own offer first, detached second; two detached → newest first, one at a time; prefix of another lesson id is
  not matched; Restore + successful Save → detached key cleared AND own key finalized; Restore without Save + remount → offered again;
  Discard → key cleared, next offered.
- M3-R12 (c): dirty orphan `lang:x` snapshot, the orphan then attached by renaming a track to `x` → offered in that track tab.
- Flush: unmount with a pending idle capture → exactly one store write with the latest lines; unmount with no pending change → 0
  writes; unmount where lines equal the baseline → clear, not write.
- Restore of a detached snapshot into a tab of another language → PUT without base token (or replace confirmation) — reuse run 1's
  helpers.

## Done = report
`cd frontend && npm run format && npm run typecheck && npm run lint && npm test` — exit codes and counts; changed files.
If the plan is ambiguous or would cause a bug, write QUESTION and argue instead of implementing around it. If the slice is too big
for one run, finish the implementation with passing existing tests first, list the missing tests as NEXT, and stop.

## Also fix (Courses 1 review of run 1, 2026-11-16)
Run 1 added `recovery.setCapturing(false)` around Save (`TranscriptEditor.tsx` Save handler). It cancels the pending idle
capture and, when the save FAILS (PATCH error, PUT error, cancelled dialog, track_changed…), capture is re-enabled but nothing is
re-armed until the lines change: the unsaved lines that the failed save left on screen are not in the store if the author then
leaves (and with flush-on-unmount there is no pending capture to flush). Fix: after a failed/cancelled save, capture the current
lines (or re-arm the idle capture) so the store holds them; keep "PATCH failure → no recovery key write/clear/finalize BY THE
FAILURE PATH ITSELF" meaning only that the failure must not clear or finalize. Test: failed save with a pending capture → after the
failure the store holds the current lines (one write), and unmount afterwards flushes nothing stale.
