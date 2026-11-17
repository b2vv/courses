NEXT: Read repository rules, accepted MS-D9 / M3-R6 / M3-R12, run 1 state, and recovery code; implement run 2 within frontend only, without commits or generated files.
FACT: Read repository briefing and startup sources; run 2 scope is frontend recovery plus this append-only log. Detailed recovery/plan inspection follows.
NEXT: Inspect accepted MS-D9, existing hook/editor/tab lifecycle and tests before editing.
FACT: Confirmed run-1 dirty frontend files and shared unrelated WIP; preserve both. Plan MS-D9 explicitly requires detached read-only placement and unmount flush.
DECISION: 2026-11-16 Codex: run-2 author verification follows the requested frontend gate; independent review/browser/full e2e remains Courses 1/MS3c, no subagent or external send authorized.
NEXT: Trace editor Save/recovery and tab/single-editor wiring, then record implementation checkpoint.
FACT: Save suspension cancels pending capture; failure currently only re-enables capture. Editor also arms capture on initial mount, requiring a no-change guard.
DECISION: 2026-11-16 Codex: small frontend slice READY; reuse existing recovery hook/store and public transcript boundary. Add pure placement plus one lesson-level reader, pass assigned offers through LessonTranscriptField. Verify flush, placement, lifecycle, failures with row-labelled rstest tests.
NEXT: Implement pure placement and lesson reader, flush lifecycle, then editor offer handling.
DECISION: 2026-11-16 Codex: detached offers are mount-local queues with accepted foreign keys retained until successful Save; Discard clears only its foreign key. New captures always use the editor own key. Language comparison is case-insensitive; track-first ordering is defensive because construction excludes matching track+orphan languages. Check via pure placement and editor lifecycle tests.
NEXT: Apply implementation and targeted tests.
FACT: Added pure exact-prefix placement, transcript-public lesson reader, tabs/single-editor wiring, and pending-capture flush. Reads stay best-effort and never migrate keys.
NEXT: Complete editor own-first queue, accepted-key finalization and failure recapture, then typecheck and tests.
FACT: Editor now gates detached offers behind the loaded own offer, restores the snapshot base pair, retains accepted keys until Save, and explicitly clears discarded keys. Failed/cancelled saves capture current lines immediately after re-enabling capture.
NEXT: Inspect initial typecheck, add row-labelled pure/hook/editor/admin integration tests.
FACT: Initial production typecheck EXIT=0. Added pure M3-R6/M3-R12c coverage and hook-level flush/isolation/best-effort tests.
NEXT: Add integration coverage for queues, remount, single→tabs/orphan attachment, Save token rules and failure capture; run targeted suite.
FACT: Added integration tests for own-first/newest queues, explicit discard, accepted-key Save/remount, exact-prefix and single-reader behavior, orphan attachment, single→tabs flush, cross-language token rules, and failed-save recapture. Updated run-1 PATCH failure assertions to permit one protective capture and still forbid clear/finalize.
NEXT: Inspect targeted results, fix concrete failures, then run the requested ordered frontend gate.
FACT: Targeted rstest startup EXIT=1 due to sandbox local port bind EPERM (::1:3000), not a test failure. Retrying the same requested tests with sandbox escalation.
NEXT: Inspect targeted tests and updated typecheck; fix actual failures.
FACT: Targeted tests and typecheck launched with the new implementation; inspecting results and scoped diff.
NEXT: Resolve any test/type failures and check remaining coverage against the brief.
FACT: Targeted tests EXIT=1, 95/96 passed. Sole failure was the unknown-language fixture passing undefined to a defaulted helper parameter (it produced en); corrected the fixture. Updated typecheck EXIT=0.
FACT: Added Ukrainian unknown-language copy/refused-Discard continuation tests and editor no-edit unmount check. Orphan attachment and single→tabs production component paths already pass in rstest.
NEXT: Run requested format→typecheck→lint→full tests in order.
FACT: Requested npm run format EXIT=0 (400 files, 204 ms). Requested typecheck started after formatting.
NEXT: Finish typecheck, then lint, then full npm test.
FACT: Requested npm run typecheck EXIT=0.
NEXT: Requested lint, then full tests.
FACT: Requested lint EXIT=1: two synchronous effect-reset setState warnings (deny-warnings). No architecture or behavior failure.
NEXT: Replace effect resets with scoped async read results, restart format→typecheck→lint→tests.
FACT: First lint-fix edit used root-relative paths from frontend cwd and failed before any edit; corrected from repository root. Read results now carry their scope, so stale results are hidden without synchronous effect resets. Restarted format EXIT=0.
NEXT: Restarted typecheck→lint→full tests.
FACT: Scoped async read identity preserves own-offer priority and prevents stale inventory being shown after lesson/tab-set changes; no lint suppression added.
NEXT: Finish ordered checks and inspect final changed-file scope.
FACT: Restarted npm run typecheck EXIT=0.
NEXT: Finish restarted lint then full requested npm test.
FACT: Restarted npm run lint EXIT=0 (240 architecture files). Full requested npm test started with the local-port sandbox permission.
NEXT: Inspect full test counts and final scope; report remaining MS3c/browser limits.
FACT: Final path review includes single-editor save token ownership, not only tabbed saves; checking its shared fallback before final handoff.
NEXT: Inspect whether single-editor fallback enforces MS-D6 when detached snapshots have unknown/different base language.
FACT: Final path review found the single-editor LessonTranscriptField fallback ignored baseLanguage; a detached snapshot could therefore send an en token while saving es. Fixed the fallback to pass a token only for its owning language, with two M3-R6 integration cases (foreign/unknown owner).
DECISION: 2026-11-16 Codex: enforce MS-D6 at the single-editor save adapter as well; tokenless PUT preserves the existing conflict refusal if a target exists. Verify via single-editor detached restore tests.
FACT: Post-fix npm run format EXIT=0.
NEXT: Restart typecheck→lint→full npm test on this final implementation.
FACT: Previous full npm test EXIT=0: 106 files, 1090 passed, 0 failed/skipped/todo, 16.701 s; this precedes the final single-editor token fix and is not claimed as final-delta evidence.
NEXT: Finish final ordered frontend checks (now 1092 expected tests).
FACT: Final-delta npm run typecheck EXIT=0.
NEXT: Finish final-delta lint, then full npm test.
FACT: Final-delta npm run lint EXIT=0, 240 architecture files, no warnings. Scoped whitespace check EXIT=0. Final full npm test running.
NEXT: Record final test count, changed files, proof limits and MS3c handoff in this log and append phase evidence.
FACT: Final status remains scoped: run-2 changes only in frontend recovery/admin wiring/locales/tests plus this log (phase evidence append pending); run-1 adapter WIP and unrelated shared WIP retained.
NEXT: Await full suite result, append precise final evidence and changed-file list.
FACT: Confirmed branch ms2-3-asr-per-track at unchanged HEAD 9ad04e3a5bb7066729f0ebbfd5cabe1847d92c51; no commit or branch mutation.
NEXT: Append final test evidence and handoff.
FACT: FINAL requested ordered frontend gate: npm run format EXIT=0 (400 files, 146 ms); npm run typecheck EXIT=0; npm run lint EXIT=0 (240 architecture files); npm test EXIT=0 (106 files, 1092 passed, 0 failed/skipped/todo, 12.806 s). 28 additional tests versus run 1. Scoped diff whitespace check EXIT=0.
FACT: M3-R6 unit/component coverage: isolated tab keys, exact lesson prefix, legacy language/unknown placement, gone-track language/default placement, own-first/newest-first single-offer queue, explicit Discard and next offer, accepted-key successful Save + own finalization, Restore/remount re-offer, one lesson keys read, single-editor detached offers, Ukrainian/English copy, unique ids, tokenless/confirmed-target Save for a different base language, viewer sweep for track/lang forms. Refused writes/clears degrade without breaking the editor.
FACT: M3-R12c component path: dirty lang:uk orphan unmounts when track is renamed uk, flushes pending lines, and offers them in that track. Single-editor→ready tabs path also flushes and offers the legacy snapshot. Hook tests cover latest lines+baseline, exactly one write, no pending change→zero writes/clears, equal baseline→clear, no repeat after idle capture, and rejection swallowing.
FACT: Failed/cancelled Save now captures the current lines after capture resumes; failure never finalizes or clears accepted detached keys. Pending-capture failure test proves one write and no stale unmount write. Run-1 PATCH error tests now permit the protective write while continuing to forbid clear/finalize. Single-editor fallback also enforces token ownership (foreign/unknown base→tokenless PUT).
FACT: Run-2 changed frontend files (13; run-1 changes in these files retained):
- frontend/src/app/providers/locales/en.ts SHA256=4bf8a6e75d7318a5bcf239fde1a700c5e595c8632207aca495e688dc449414dc
- frontend/src/app/providers/locales/uk.ts SHA256=e70feda0059c6bb56f90f6934006fc0d5c7c844666abbb21881d01022a415c5f
- frontend/src/modules/admin/ui/LessonLanguageTabs.tsx SHA256=dfa27c7d8b6aa16ab357a018b0d1e8a32a79060afdbefcd30fddc488ac2e5ca6
- frontend/src/modules/admin/ui/LessonLanguageTabs.test.tsx SHA256=9e2afe3b13bf790e08b9e56a2101ff01dec52e7e213768deddfa5bf8f5717acd
- frontend/src/modules/transcript/application/recovery.ts SHA256=a4df36b1c7afba4a8f35b71178410675c86f0d10cf6bb1b1784dd20c2d603b16
- frontend/src/modules/transcript/application/recovery.test.ts SHA256=afb32396b5bb5096540bc7aef3d43979bc9e7aabf70b38ba4860f8db387f0c57
- frontend/src/modules/transcript/index.ts SHA256=5e2d614cc1f0f7d0a88d7122c788638f1a41e35ab6c23bd9de7135dad7b8e509
- frontend/src/modules/transcript/ui/LessonTranscriptField.tsx SHA256=27791ab53acd06c836ac0fc1447860d23f57eba9ee27425d015ed135486564d6
- frontend/src/modules/transcript/ui/TranscriptEditor.tsx SHA256=2050c16641db124b256aae2e2198f8c6943b4dbf68e23809a2b797aaf3b6033c
- frontend/src/modules/transcript/ui/TranscriptEditor.test.tsx SHA256=e83f8296b8a66128067ee4de70c289294a17b3e1c9e1623a95bf624b433abe98
- frontend/src/modules/transcript/ui/use-transcript-recovery.ts SHA256=010c781dc744aa795169d8d2a6b505e05296e394d8d337096945c32ebb114c61
- frontend/src/modules/transcript/ui/use-transcript-recovery.test.tsx SHA256=cfd9d3c9d0687ab4800bdec57c3d8bb742028251414409d1dc71fbe73089f623
- frontend/src/modules/transcript/ui/use-detached-transcript-recovery.ts SHA256=94604d06817628f9426bf520233bec4428a4fb65f14fa2b2a1d7b304dfe85e6a
FACT: Final frontend snapshot fingerprint SHA256=d9611f3167f0749a75091b94b71d9dad9de3ffb38aa5308a6f6308a2d281ed7f (ordered SHA256 + space + path lines above, joined by newline). Root /Users/strelia/projects/courses; branch ms2-3-asr-per-track; HEAD 9ad04e3a5bb7066729f0ebbfd5cabe1847d92c51. Author verification only; independent review has no run-2 verdict yet.
FACT: Documentation delivery is this append-only working file plus appended run-2 evidence in docs/multi-subtitles/report.md. No commit, checkout/reset/stash/clean/rm, dependency addition, generated/e2e/backend edit. Shared unrelated WIP and run-1 adapter/test WIP remain intact.
FACT: Limits: rstest exercises production React component/hook paths with in-memory history; real IndexedDB page teardown, browser focus/layout and persistence are NOT_TESTED here. Store operations remain best-effort; teardown cannot guarantee completion. Full ordered make check and browser/e2e are NOT_RUN under this run scope. No phase closure claimed.
NEXT: Run-2 implementation and requested unit tests complete; no missing in-scope test identified. Courses 1: independent review of the final dirty snapshot. MS3c: browser/e2e M3-R6 and M3-R12 a/b plus the final ordered make check. No further run-2 implementation pending.
FACT: Final evidence append and documentation/frontend whitespace verification EXIT=0. Run-2 handoff ready; final response summarizes this file.
NEXT: Courses 1 / MS3c review and browser/full-gate ownership as recorded above.
