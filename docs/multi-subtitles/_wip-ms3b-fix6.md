NEXT: Read the repository briefing, workflow instructions, current MS3b plan/review and recovery implementation; inventory every snapshot deletion path before editing.
FACT: Read the repository briefing first; reviewed workflow, task context, review w6 and the relevant implementation/review skill instructions. User scope is frontend recovery + tests only, without commits, backend, e2e or generated edits.
NEXT: Read the MS3b plan, local instructions and transcript recovery code/tests; verify the review against current files and record the deletion inventory.
FACT: Read current recovery hooks, editor, application and store contract; searched all transcript deletion sites and reviewed repository standards. Review findings are being traced through the actual capture/finalize/offer flow.
DECISION: 2026-11-16, Codex (owner-directed): successful Save retires only the version this editor last wrote; a newer foreign capture survives and remains available next visit. Verification: same/newer version path matrix and remount regression.
NEXT: Inspect adapter implementations and existing tests, then implement version ownership across the inventoried paths and add failing regressions.
FACT: Confirmed branch ms2-3-asr-per-track at 9ad04e3a5bb7066729f0ebbfd5cabe1847d92c51 with extensive pre-existing WIP; no nested AGENTS files. Two unconditional own-key deletions remain in clean capture and finalize; detached Discard and accepted-source retirement already use clearIfVersion.
NEXT: Add per-key read/write ownership to the hook and compare-delete own Discard; preserve newer offers. Inspect focus continuation and deferred-store test fixtures.
FACT: Own Discard currently calls finalize; finalize and clean capture call unconditional clear. Detached Discard compares its offered version and refreshes on mismatch; accepted-source Save cleanup compares its accepted version atomically. Existing hook tests encode unconditional no-version deletion and must be updated to the owner rule.
DECISION: Resume/small route, Codex is the sole writer; existing owner-managed review session w6 is the input and final review delivery goes to the owner, with no reviewer spawn/send. Acceptance: all ordinary deletions use a known read/offered/written version; unknown version never deletes; mismatch is retained/offered; delayed recovery actions preserve moved focus.
NEXT: Implement scoped hook ownership and regressions; retain unconditional clear only for the explicit cross-author viewer sweep.
FACT: Correction to the preceding trace note: own Discard calls dismiss (hides/remembers the offer), not finalize. Its later clean blur is the destructive path; this run will compare-delete the displayed own version at Discard and refresh a mismatch.
FACT: Deletion inventory before edits (frontend/src/modules/transcript/ prefix): ui/use-transcript-recovery.ts:119 clean capture and :200 Save finalize use unconditional clear; ui/TranscriptEditor.tsx:262 detached Discard uses offered capturedAt; application/recovery.ts:266 accepted-source retirement uses accepted capturedAt after identity/content/ownership checks; application/recovery.ts:185 forgetOtherAuthors intentionally clears foreign-user keys. Adapters: memory.ts:19/:23 implement clear/clearIfVersion; indexed-database.ts:124/:125 implement the same, with get+delete in one readwrite transaction. Own Discard ui/TranscriptEditor.tsx:389 currently dismisses without deleting.
DECISION: All ordinary cleanup uses clearIfVersion with the instance-owned read/offered/written version. Cross-author sweep remains unconditional because removing another user's local data is its purpose, regardless of version. Save awaits its last initiated successful write and retires that exact version; no own write means no Save deletion. Unknown-version clean capture performs no deletion. Scope is a small frontend repair, not a new subsystem/API; pinned checkpoint read via exact-file fallback; plan critique/defense N/A for this owner-specified repair. Pre-code READY: no unresolved contract or writer dependency; browser/e2e outside authorized run scope.
NEXT: Add the major regressions before changing production logic, run them red, then implement and run them green.
FACT: Added the version matrix (blur/idle/unmount/own Discard/Save × same/newer), unknown-version regression and production two-window Discard → blur → remount regression. Baseline targeted test exit: 1; raw log /tmp/ms3b-fix6-major-red.log (production implementation unchanged).
NEXT: Implement the major fix, then verify the new tests and update assertions that previously required unconditional deletion.
FACT: Sandbox blocked Rstest's local listener (::1:3000), so reran the authorized targeted checks with sandbox escalation. Pre-fix targeted exit: 1; /tmp/ms3b-fix6-major-red.log.
NEXT: Apply version ownership once baseline failures are confirmed.
FACT: Pre-fix targeted tests failed on the production two-window scenario, all newer-version clean paths, own Discard same/newer, Save newer, and unknown-version cleanup (8 failures; raw log retained). Implemented per-key/store ownership; clean and own Discard compare-delete the seen version and refresh mismatches; Save awaits and compares its written version only. Updated obsolete hook expectations that demanded unknown-version deletion.
NEXT: Run the targeted suite green; inspect any existing Save tests that bypass real blur/capture and make the Save path explicitly capture the saved lines when needed.
FACT: Major targeted post-fix test exit: 1; /tmp/ms3b-fix6-major-green.log.
NEXT: Resolve any regressions, complete the remaining path-matrix rows, then address deferred focus.
FACT: Inspected major post-fix results and retained raw failure evidence; additional work is limited to the frontend recovery paths and tests already in scope.
NEXT: Ensure successful Save owns a capture of the submitted lines even when invoked without a browser blur, and add detached/accepted retirement matrix rows.
FACT: Major post-fix raw failure details reviewed; new ownership paths pass, while four existing Save tests expose that synthetic click skips the normal pre-Save blur capture. The Save action must explicitly flush submitted lines, so finalization has a written version independent of input modality.
NEXT: Add explicit pre-Save capture and deferred focus regressions; keep the newer-window Save test writing C after the editor's own capture.
FACT: Correction: post-fix failures were three Save tests lacking a prior own write and one old Discard test requiring snapshot retention. Unconditionally recapturing on every Save would overwrite a newer foreign capture, so that approach is rejected.
DECISION: Save flushes only a pending author change before capture suspension; it does not recapture already-flushed lines. Thus a foreign C after this editor's last write survives. Same-version own Discard now retires the version by explicit user request; old retention-only assertion will verify deletion plus no remount offer instead.
NEXT: Implement pending-change flush through the existing capture suspension API and add deferred focus tests.
FACT: Save now flushes only pending author changes before suspension, preserving newer captures after an earlier flush. Added deferred detached Discard focus tests for same/newer versions; they move focus to an outside control while clearIfVersion is pending.
NEXT: Run the focus tests against the unchanged focus implementation to record their failing baseline.
FACT: Deferred-focus regression baseline exit: 1; raw log /tmp/ms3b-fix6-focus-red.log. No focus fix applied yet.
NEXT: Retain the initiating element and guard focus at completion.
FACT: Focus regression baseline fails for both same/newer deferred Discard completions. Replaced the boolean focus flag with the initiating control; both synchronous Restore/Discard and async detached continuation recheck activeElement, moving focus only from that control or body.
NEXT: Run focused and broad transcript suites; add remaining explicit inventory matrix coverage and perform actual revert-checks against saved pre-run production bytes.
FACT: Broad transcript suite exit: 0; raw log /tmp/ms3b-fix6-transcript.log.
NEXT: Address any failures and finish the deletion-path matrix and revert evidence.
FACT: Inspected current hook logic and existing same/newer accepted-source/atomic-delete matrix tests as part of the final class audit; adapter atomicity is already covered by its transaction test.
NEXT: Complete explicit matrix coverage for viewer sweep and Save after another window's post-flush capture; verify no production ordinary clear remains.
FACT: Broad transcript suite passed 242/242 in 11 files (exit 0). Added explicit component Save matrices (own-key and accepted-source, same/newer written during deferred Save) with remount offers and request counts; added the same/newer viewer-sweep exception matrix. Detached Discard matrix remains fix5 same/newer in production components, plus both deferred focus variants in this run.
NEXT: Run the complete scoped matrix, then revert each production fix temporarily and record red/restore-green. Existing shared WIP stays preserved in /tmp pre-run copies; no git undo commands are used.
FACT: Full scoped deletion/focus matrix exit: 0; raw log /tmp/ms3b-fix6-matrix.log.
NEXT: Run temporary-production-revert checks, restoring exact fixed bytes afterward.
FACT: Pre-revert fixed snapshot SHA-256:
- frontend/src/modules/transcript/ui/use-transcript-recovery.ts: bfff177de5cdf45296268aef2844a353f1f1de62d2547b8bc575e8ca60336940
- frontend/src/modules/transcript/ui/TranscriptEditor.tsx: 73b07492285d8eba56e40a80923754bd31c39bf7d7b6f2be55c3b2b8bc6006dc
- frontend/src/modules/transcript/ui/use-transcript-recovery.test.tsx: b1c2f67f353b1ab389af205a1dad0edbe1969c24212e86edd8bae399d72e2150
- frontend/src/modules/transcript/ui/TranscriptEditor.test.tsx: dd6cd847333f9ee21e6677b8282bf3575a93ad3e19d8b3a75ecf39077152a7dc
- frontend/src/modules/transcript/application/recovery.test.ts: f46c1b37a38a950306e08051bc952e6c61d760cbd9e5d6eb39e0f28ba825d246
NEXT: Revert only each fix implementation to the preserved pre-run bytes, execute its regressions, restore in finally, then run the ordered frontend checks.
FACT: Actual major revert-check: replaced only use-transcript-recovery.ts with preserved pre-run production bytes, retained regression tests; npm test -- src/modules/transcript/ui/use-transcript-recovery.test.tsx src/modules/transcript/ui/TranscriptEditor.test.tsx --testNamePattern "MS3b fix6" exited 1; counts {'testFiles': 2, 'failedFiles': 2, 'tests': 86, 'failedTests': 9, 'passedTests': 9, 'skippedTests': 68, 'todoTests': 0}; raw log /tmp/ms3b-fix6-major-revert.log.
FACT: Restored exact fixed bytes for use-transcript-recovery.ts in finally after the major revert-check.
NEXT: Run the restored implementation green and complete the ordered frontend verification.
FACT: Actual focus revert-check: replaced only TranscriptEditor.tsx with preserved pre-run production bytes, retained regression tests; npm test -- src/modules/transcript/ui/TranscriptEditor.test.tsx --testNamePattern "MS3b fix6 deferred" exited 1; counts {'testFiles': 1, 'failedFiles': 1, 'tests': 67, 'failedTests': 2, 'passedTests': 0, 'skippedTests': 65, 'todoTests': 0}; raw log /tmp/ms3b-fix6-focus-revert.log.
FACT: Restored exact fixed bytes for TranscriptEditor.tsx in finally after the focus revert-check.
NEXT: Run the restored implementation green and complete the ordered frontend verification.
FACT: The full transcript matrix passed 248/248 in 11 files before the actual revert-checks. Audited all non-test transcript deletion sites again: only forgetOtherAuthors calls unconditional clear; ordinary hook cleanup, detached Discard and accepted-source retirement call clearIfVersion.
NEXT: Preserve generated/e2e/backend fingerprints, run the exact requested format → typecheck → lint → test sequence, then record final inventory line references and changed files.
FACT: Revert-check results: Major exit 1 (9 failed, 9 passed selected regressions), focus exit 1 (2 failed); fixed production bytes restored in finally. Retained synchronous store-error containment after refactoring capture and extended its existing test to thrown/rejected failures. Fingerprinted backend/e2e/generated files and all frontend inputs before formatting; formatter already excludes generated routes/API files.
NEXT: Execute npm run format, npm run typecheck, npm run lint, npm test in that exact order; record individual exits and test counts.
FACT: Ordered check 1: cd frontend && npm run format — exit 0; raw log /tmp/ms3b-fix6-format.log.
NEXT: Run typecheck, then lint, then the full frontend test set.
FACT: Ordered check 2: cd frontend && npm run typecheck — exit 2; raw log /tmp/ms3b-fix6-typecheck.log.
NEXT: Run lint, then the full frontend tests.
FACT: Ordered check 3: cd frontend && npm run lint — exit 0; raw log /tmp/ms3b-fix6-lint.log.
NEXT: Run the full frontend tests with the local listener enabled.
FACT: Fingerprint check protected: changed files [].
FACT: Fingerprint check frontend formatter delta: changed files ['frontend/src/modules/transcript/ui/use-transcript-recovery.ts', 'frontend/src/modules/transcript/ui/TranscriptEditor.test.tsx', 'frontend/src/modules/transcript/ui/use-transcript-recovery.test.tsx'].
NEXT: Run full frontend tests after completed typecheck/lint and inspect final scoped diff.
FACT: First ordered typecheck exited 2: the new matrix inferred a narrower local fixture type than TranscriptSnapshot. Lint exited 0 (architecture: 240 files). Corrected only the fixture annotation, with no assertion/cast; protected files remain byte-identical, formatter changed only this run's three frontend files.
NEXT: Restart the requested ordered frontend checks on the corrected final snapshot.
FACT: Final ordered check 1: npm run format — exit 0; 400 files processed; /tmp/ms3b-fix6-format.log.
NEXT: Final typecheck → lint → full tests.
FACT: Re-read the final ownership/focus implementation after formatting and checked completed ordered-check outputs before launching the final frontend suite.
NEXT: Execute the full frontend test set and append final inventory, evidence and scope hashes.
FACT: Final ordered check 2: npm run typecheck — exit 0; /tmp/ms3b-fix6-typecheck.log.
NEXT: Final lint → full tests.
FACT: Final ordered check 3: npm run lint — exit 0; /tmp/ms3b-fix6-lint.log.
NEXT: Final full frontend tests.
FACT: Final deletion inventory and matrix test line references gathered after formatting. No new interfaces or dependencies; existing conditional-delete adapters are reused unchanged.
NEXT: Record the final full test counts and hand off the final frontend snapshot for owner-managed review.
FACT: Final ordered check 4: cd frontend && npm test — exit 1; raw log /tmp/ms3b-fix6-test.log (full frontend suite, no test filter).
NEXT: Record counts and final snapshot inventory; verify protected-file fingerprints and working-tree scope.
FACT: Final deletion inventory (all paths; prefix frontend/src/modules/transcript/):

| Path / final file:line | Version rule | Same/newer proof |
| --- | --- | --- |
| Clean capture: ui/use-transcript-recovery.ts:153–154 → retire :130 (blur, idle, unmount) | Compare-delete last instance read/offered/written version; unknown → no delete; mismatch read → newer offer. | ui/use-transcript-recovery.test.tsx:164, paths blur/idle/unmount × same/newer; unknown :208; production two-window blur/remount ui/TranscriptEditor.test.tsx:1007. |
| Own Discard: ui/TranscriptEditor.tsx recovery Discard → ui/use-transcript-recovery.ts:229–233 → retire :130 | Compare-delete own offered/read version, refresh mismatch and remove dismissal for newer offer. | Hook matrix :164 own Discard × same/newer; production :1007 through blur/remount. |
| Save finalization: ui/TranscriptEditor.tsx:556 → ui/use-transcript-recovery.ts:236–246 → retire :130 | Only this editor's last written capturedAt, awaiting successful write; no write → no deletion. A newer foreign capture survives and is offered next visit. Pending author edits flush at capture suspension :249; previously flushed lines are not rewritten. | Hook matrix :164 Save × same/newer; component ui/TranscriptEditor.test.tsx:1225 own Save × same/newer, C written while Save pending then remount; exactly one save. |
| Detached Discard: ui/TranscriptEditor.tsx:265 | Atomic compare-delete offered capturedAt; mismatch read/refreshed offer, no handling of newer version. | Existing component matrix ui/TranscriptEditor.test.tsx:951 same/newer; deferred matrix :1187 also verifies same/newer store result and focus. |
| Accepted-source retirement: application/recovery.ts:247–266, invoked after accepted Save ui/TranscriptEditor.tsx | Accepted capturedAt plus current ownership/content guards; atomic compare-delete protects intervening writes. | Existing application matrix application/recovery.test.ts:288 same/newer plus owned/read-error/content cases, race :341; component :1225 accepted-source Save × same/newer + revisit. |
| Viewer sweep: application/recovery.ts:179–185 forgetOtherAuthors | Explicit exception: unconditional clear of foreign-author keys (or all keys after confirmed sign-out), because identity cleanup deliberately removes other users' data regardless of version. Own-author keys stay. | application/recovery.test.ts:366 same/newer both deleted for foreign, own retained; existing null-viewer coverage. |
| Adapter primitives: adapters/transcript-history-memory.ts:19/23; adapters/transcript-history-indexed-database.ts:124/125 | clear is used only by viewer sweep; clearIfVersion is the ordinary primitive. IndexedDB get+delete execute in one readwrite transaction; memory compares before delete. No adapter changes this run. | Existing atomic adapter transaction tests; application/recovery.test.ts:341 same/newer direct primitive assertions; full frontend suite includes them. |

FACT: Audited all frontend callers as well as module-local deletion sites; no additional snapshot-removal path found. Final author diff review confirms no unconditional ordinary cleanup remains. Protected backend/e2e/generated fingerprints are unchanged.
NEXT: Await full frontend test completion, append counts and final changed-file hashes; deliver this record to the owner for independent review.
FACT: Final protected-file fingerprint comparison: 769 files checked; changed=[]. No backend/e2e/generated edits.
FACT: Final changed files for this run only (pre-existing WIP elsewhere preserved):
- frontend/src/modules/transcript/ui/use-transcript-recovery.ts — SHA-256 dab5e43bf442d6a6686765581f7c9f7f9bf8d859216aa8cb3d6d14a8da1ad176
- frontend/src/modules/transcript/ui/TranscriptEditor.tsx — SHA-256 73b07492285d8eba56e40a80923754bd31c39bf7d7b6f2be55c3b2b8bc6006dc
- frontend/src/modules/transcript/ui/use-transcript-recovery.test.tsx — SHA-256 e6255740cdcd3d513ccee92d41c6ce7496a8b19d6f800357a8bf34931a9ac521
- frontend/src/modules/transcript/ui/TranscriptEditor.test.tsx — SHA-256 7a64eb10d4e2b8f4d0af0d7e1ed064b7b6d9a8e17d22a0156d57b316cdf8de76
- frontend/src/modules/transcript/application/recovery.test.ts — SHA-256 f46c1b37a38a950306e08051bc952e6c61d760cbd9e5d6eb39e0f28ba825d246
- docs/multi-subtitles/_wip-ms3b-fix6.md — append-only working report.
NEXT: Extract final full-suite counts, state verification boundaries and handoff.
FACT: Full frontend test exit 1: 1182 passed / 2 failed / 1184 total in 106 files, no skips. Both failures are in LessonLanguageTabs recovery integration tests: one still observes raw clear instead of versioned cleanup, the other detects an extra failed-Save capture after the new pending flush. The scoped transcript suite alone did not reveal them.
NEXT: Inspect both failures, update the obsolete clear spy to conditional deletion and remove duplicate capture in the failed-Save path while preserving failure recovery.
FACT: Full-suite failures resolved surgically: LessonLanguageTabs test now asserts no raw clear and exact written capturedAt on clearIfVersion; removed the redundant failed-Save recapture, since pending lines are already captured before suspension. The existing full integration test is a failing-before proof for this duplicate-write regression (2 writes instead of 1), and will remain unchanged.
DECISION: Failed Save keeps the pre-Save capture; it does not create a second capture after the asynchronous response, which could overwrite intervening work. Verification: existing failed-Save integration test checks one write, retained detached source and no extra unmount write.
NEXT: Rerun scoped recovery integration, then the exact final ordered frontend sequence on the updated snapshot.
FACT: Integrated transcript + LessonLanguageTabs recovery check exit: 1; /tmp/ms3b-fix6-integrated.log.
NEXT: Final ordered format → typecheck → lint → full tests.
FACT: The only subsequent production delta after the two actual revert-checks is removal of failed-Save duplicate capture; version ownership and the tested focus guard are unchanged. Their reverted-red evidence remains applicable to those exact mechanisms; restored-green proof is the final full suite.
NEXT: Run all four requested checks in order with separate final logs, stopping if a check fails.
FACT: Final snapshot ordered check 1: cd frontend && npm run format — exit 0; counts=N/A (static/format check); raw log /tmp/ms3b-fix6-final-format.log.
NEXT: Continue the remaining ordered checks.
FACT: Integrated rerun remains red (306 passed / 2 failed / 308); inspecting the changed failures rather than treating the earlier two as the only coupling. Final ordered run is already executing and will report its actual result.
NEXT: Diagnose the remaining exact capture-count assertions, then make the smallest correction and rerun the final gates if needed.
FACT: Final snapshot ordered check 2: cd frontend && npm run typecheck — exit 0; counts=N/A (static/format check); raw log /tmp/ms3b-fix6-final-typecheck.log.
NEXT: Continue the remaining ordered checks.
FACT: Final snapshot ordered check 3: cd frontend && npm run lint — exit 0; counts=N/A (static/format check); raw log /tmp/ms3b-fix6-final-lint.log.
NEXT: Continue the remaining ordered checks.
FACT: Remaining failures are PATCH-rejection tests that spy after a prior blur capture and require another post-failure write. Removing redundant failure capture preserves that existing snapshot but changes this assertion; inspect whether the requirement is semantic retention or an intentional retry before changing the test.
NEXT: Retain safe retry semantics if needed without duplicating a pre-Save pending capture or overwriting a newer version.
FACT: Final snapshot ordered check 4: cd frontend && npm test — exit 1; counts={'testFiles': 106, 'failedFiles': 1, 'tests': 1184, 'failedTests': 2, 'passedTests': 1182, 'skippedTests': 0, 'todoTests': 0}; raw log /tmp/ms3b-fix6-final-test.log.
NEXT: Complete final inventory/scope verification and handoff.
FACT: The two remaining PATCH tests had intentionally spied after a completed blur capture. Updated their expectation from a redundant post-response write to zero writes PLUS exact byte-for-byte snapshot retention, keeping their zero PUT/clear and displayed-lines checks. This is an equivalent/stronger retention proof under the new pre-Save flush timing, not a dropped business assertion.
FACT: Added a deferred failed-Save regression: A is captured before Save, another window writes C, server failure returns; C and displayed A both survive with zero post-response writes. This would fail with the removed failure recapture.
NEXT: Revert-check this newly discovered failed-Save overwrite regression, restore, then run the entire ordered check set once more.
FACT: Failed-Save recapture revert-check (restore only removed else captureNow branch) exit 1; regression must preserve foreign C; raw log /tmp/ms3b-fix6-failed-save-revert.log.
FACT: Restored exact fixed editor bytes in finally after failed-Save revert-check.
NEXT: Run format → typecheck → lint → full npm test on the restored final snapshot.
FACT: Verified final snapshot ordered check 1: cd frontend && npm run format — exit 0; counts=N/A (static/format check); raw log /tmp/ms3b-fix6-verified-format.log.
NEXT: Continue the remaining ordered checks.
FACT: Verified final snapshot ordered check 2: cd frontend && npm run typecheck — exit 0; counts=N/A (static/format check); raw log /tmp/ms3b-fix6-verified-typecheck.log.
NEXT: Continue the remaining ordered checks.
FACT: Verified final snapshot ordered check 3: cd frontend && npm run lint — exit 0; counts=N/A (static/format check); raw log /tmp/ms3b-fix6-verified-lint.log.
NEXT: Continue the remaining ordered checks.
FACT: Post-format protected fingerprint check: 769 backend/e2e/generated files, changed=[]. git diff --check exits 0.
NEXT: Finish final test result and final hashes; no additional product edits are planned.
FACT: Verified final snapshot ordered check 4: cd frontend && npm test — exit 0; counts={'testFiles': 106, 'failedFiles': 0, 'tests': 1185, 'failedTests': 0, 'passedTests': 1185, 'skippedTests': 0, 'todoTests': 0}; raw log /tmp/ms3b-fix6-verified-test.log.
NEXT: Complete final inventory/scope verification and handoff.
FACT: Final changed-file manifest (supersedes earlier pre-integration hashes):
- frontend/src/modules/transcript/ui/use-transcript-recovery.ts — SHA-256 dab5e43bf442d6a6686765581f7c9f7f9bf8d859216aa8cb3d6d14a8da1ad176
- frontend/src/modules/transcript/ui/TranscriptEditor.tsx — SHA-256 4ce32f241b38978d71869d1bd05027aea968bc7ee820ac3510b664e891f6dfe3
- frontend/src/modules/transcript/ui/use-transcript-recovery.test.tsx — SHA-256 e6255740cdcd3d513ccee92d41c6ce7496a8b19d6f800357a8bf34931a9ac521
- frontend/src/modules/transcript/ui/TranscriptEditor.test.tsx — SHA-256 2b9d6bf9f500a1d6d0c81904d62ec48fe4dc9f9222d9f192491d7b6c2f382402
- frontend/src/modules/transcript/application/recovery.test.ts — SHA-256 f46c1b37a38a950306e08051bc952e6c61d760cbd9e5d6eb39e0f28ba825d246
- frontend/src/modules/admin/ui/LessonLanguageTabs.test.tsx — SHA-256 09bd61c4caba2df89b726c6d906bbe933bc6ca505030349310498918cedbf628
- docs/multi-subtitles/_wip-ms3b-fix6.md — append-only task report.
FACT: Final code changes comprise versioned own-key cleanup, pending-author-edit capture before Save, no post-failure recapture, and initiating-control/body focus guard. All other transcript deletion implementations are unchanged and audited in the inventory above.
NEXT: Record verified final counts and deliver the working file as the review handoff.
FACT: Final report — both w6 findings fixed; final ordered npm run format / npm run typecheck / npm run lint / npm test all exit 0. Format processed 400 files; lint architecture checked 240 files; full tests 1185 passed in 106 files, 0 failed/skipped/todo.
FACT: Revert-check summary (actual temporary production changes, always restored):
- major-revert: exit 1 (expected), counts={'testFiles': 2, 'failedFiles': 2, 'tests': 86, 'failedTests': 9, 'passedTests': 9, 'skippedTests': 68, 'todoTests': 0}; /tmp/ms3b-fix6-major-revert.log. Restored implementation green in final full suite.
- focus-revert: exit 1 (expected), counts={'testFiles': 1, 'failedFiles': 1, 'tests': 67, 'failedTests': 2, 'passedTests': 0, 'skippedTests': 65, 'todoTests': 0}; /tmp/ms3b-fix6-focus-revert.log. Restored implementation green in final full suite.
- failed-save-revert: exit 1 (expected), counts={'testFiles': 1, 'failedFiles': 1, 'tests': 68, 'failedTests': 1, 'passedTests': 0, 'skippedTests': 67, 'todoTests': 0}; /tmp/ms3b-fix6-failed-save-revert.log. Restored implementation green in final full suite.
FACT: Verification boundaries — production React components/hooks plus memory history under happy-dom; existing IndexedDB adapter transaction tests included in full suite. Real-browser focus, real persistent IndexedDB and e2e/backend gates were not run in this authorized frontend-only scope; no such PASS is claimed. No commits, git checkout/reset/stash/clean/rm, backend/e2e/generated edits, or dependency changes.
FACT: Working tree remains on ms2-3-asr-per-track at 9ad04e3a5bb7066729f0ebbfd5cabe1847d92c51; prior WIP preserved. Final run changed six frontend files plus this append-only record. Final inventory above covers every deletion caller and the viewer-sweep exception; Save preserves newer foreign captures for the next visit.
NEXT: Owner-managed independent review of the final snapshot using this working file; no further implementation is pending for the requested fix run. Independent final review is pending, not represented by author checks.
