NEXT: Read repository instructions, accepted MS2–3 plan, MS3a checkpoint and existing frontend implementation before editing.
FACT: Read repository briefing/workflow and ponytail skill; user scope owns frontend-only implementation and unit checks, while Courses 1 owns browser, full gates, review and commits.
NEXT: Verify accepted checkpoint and trace draft/audio contracts, queries, save invalidation and switch primitive.
FACT: Snapshot confirmed /Users/strelia/projects/courses, ms2-3-asr-per-track @ a985543ffac3a5ed9d5f9e34bee9848245325a7f; frontend initially clean. Existing unrelated dirty/untracked files preserved.
FACT: MS3a checkpoint is READY, whole-plan r9 critique accepted; this run resumes its frontend subset. Courses 1 retains independent review/browser/full gate ownership.
NEXT: Finish scoped file fallback impact tracing (no graph refresh), then implement the list/inventory contracts and per-track polling.
FACT: File fallback traced DraftRepository→adapter→createDraftQueries→admin course route→LessonDraftField; audio repository/application/query/route→LessonAudioTracksField; transcript save lives in admin course route. No existing frontend und-family helper found.
DECISION: Reuse presentation/ui Switch and existing audio rename mutation for complete language/label/multilingual PATCH; preserve multilingualDefault through audio query state. Replace stale polling with per-track regeneration markers only (O7).
NEXT: Implement list/inventory boundary and polling, then propagate multilingual and affected lesson IDs through audio writes.
FACT: Draft port/adapter now returns every validated draft with track snapshots; removed adapter first-item bridge. Added strict transcript-inventory schema/port and pure draftLoadable/isUndetermined helpers.
NEXT: Replace query implementation with list polling and per-track markers, then update consumers/tests.
FACT: Draft queries now share one list per lesson, expose draftOfTrack, add non-polling inventory, and hold regeneration markers by lessonId:trackId in the factory. Removed awaitingReplacement/staleSince.
NEXT: Propagate audio multilingual writes/default and scoped draft invalidation; preserve default-first compile-through at the route boundary.
FACT: Audio queries retain the full list/default response; mint and rename propagate multilingual. PATCH/delete/default refresh lesson drafts, resolving lesson from cached course-video rows (fallback all only for detached callers without those rows). Removed obsolete automatic-recognition comments.
FACT: Route selects the first draft with the requested run-2 tabs marker; successful transcript saves invalidate the inventory key.
NEXT: Add both switches and locale copy, then update and extend unit coverage.
FACT: Existing Switch uses React Aria, ui-choice token styles with minimum touch size and focus-visible rules; it forwards aria-disabled. The run-1 row switch will guard onChange during pending without isDisabled.
NEXT: Add i18n descriptions/names and switch interactions; inspect compile errors to update all affected typed fixtures.
FACT: Added new-form and row switches using existing RAC primitive; new value derives from multilingualDefault until edited, upload includes it, row toggles send current language/label plus mode, and pending uses aria-disabled with guarded onChange. Added en/uk names/descriptions.
FACT: Corrected stale-banner copy in both locales to request Regenerate rather than promise automatic recognition (O7).
NEXT: Update typed fixtures/old stale-poll assertions and add list/inventory, guard, marker isolation and switch tests.
FACT: Added adapter list/inventory tests (strict item/envelope rejection and named errors), pure guard branch tests including und plus mode mismatch, and complete PATCH-body assertion. Updated impacted UI draft fixtures.
NEXT: Rewrite query tests around per-track regeneration/O7 and extend audio cache/switch coverage, then run requested frontend checks.
FACT: Query tests now exercise two simultaneous track markers, sibling changes, server claimed/running polling, stale-ready no-poll, 15-minute expiry, and real unobserved 5-minute GC/remount. Updated S6 integration-at-unit-level tests to O7 instead of asserting removed stale polling.
NEXT: Add scoped audio invalidation and switch behavior tests, then format/typecheck/lint/test.
FACT: Inspected remaining audio argument assertions and route/default state wiring before final coverage additions; no backend/generated/e2e changes made.
NEXT: Add row Space-toggle/pending focus/error coverage and mint/default propagation checks, then run the requested command sequence.
FACT: Added switch default/toggle/upload tests, row Space/full-metadata/pending-focus/error test, audio PATCH/delete/default lesson-scoping checks, and mint propagation test. Updated rename/upload expectations to include mode.
NEXT: Run frontend format → typecheck → lint → npm test, recording exits/counts; fix failures without touching backend/generated/e2e.
FACT: Started npm run format (frontend); formatter excludes generated API files and routeTree.gen.ts via existing config. Awaiting exit before typecheck.
NEXT: Collect format result, then typecheck.
FACT: npm run format exit 0 (395 files). Started npm run typecheck.
NEXT: Collect typecheck errors and repair any compile-through fixtures/callers.
FACT: Collected first typecheck output and inspected query diff for polling/cache correctness and unexpected formatting scope.
NEXT: Fix compiler findings and simplify marker reconciliation if needed before continuing lint/tests.
FACT: Initial npm run typecheck exit 2: generated inventory operation is not named getLessonTranscripts as brief states; plus callback return typing, two legacy rename test calls, optional property exactness, and route-guard port fixture need adjustment.
NEXT: Use the existing generated inventory operation (no generated changes), repair those callers and tighten marker expiry during failed initial reads.
FACT: Fixed first typecheck findings using generated listLessonTranscripts; corrected typed fixtures/calls and explicit default property. Marker expiry now runs even when no successful draft read exists.
DECISION: Generated operation naming differs from brief only; reuse listLessonTranscripts for the specified endpoint, with no API or scope change.
FACT: Re-ran frontend format after fixes (exit recorded by command output).
NEXT: Re-run typecheck, then lint and tests.
FACT: Started second npm run typecheck after compile-through repairs.
NEXT: Collect result before lint.
FACT: Collected second typecheck result; proceeded to lint when typecheck passed.
NEXT: Resolve any lint finding, then execute the full frontend unit suite.
FACT: Second npm run typecheck exit 0; npm run lint exit 0 (architecture boundaries checked: 238 files). Full npm test started; log /tmp/courses-ms3a-run1-tests.log.
NEXT: Collect full test result/counts and repair any real or fixture failures.
FACT: Initial npm test exited 1 immediately; inspecting captured startup log before interpreting as test failure.
NEXT: Address startup blocker or rerun using the safe environment correction, then collect actual test results.
FACT: npm test startup was blocked by sandbox EPERM binding ::1:3000 (0 tests executed). Requested the same authorized unit-suite command with escalation for Rstest's local listener; no browser/full gate started.
NEXT: Collect escalated suite result and test counts.
FACT: Checked live unit output, whitespace diff, and the default-first route marker while full unit suite runs.
NEXT: Collect terminal suite result; investigate any failing assertions rather than widening scope.
FACT: Polled the full unit process and inspected latest output; final unit exit/counts remain pending if process is still running.
NEXT: Finish the suite, fix scoped failures, and record the final changed-file inventory/hashes for Courses 1.
FACT: Full unit suite exited 1; captured log shows inventory test parameterization passes individual items where an array is expected. Inspecting all failures/counts before fixing.
NEXT: Correct test.each array cases and any remaining scoped failures, then rerun the necessary suite.
FACT: First executed suite: 103 files, 1018 tests, 1012 passed / 6 failed / 0 skipped. Failures are confined to new tests: RAC switch uses native checked (not aria-checked), and test.each destructures inventory arrays.
NEXT: Fix those assertions/cases without altering product behavior; rerun scoped tests followed by final full suite.
FACT: Corrected native checked assertions and inventory parameter cases; npm run format exit 0 after these test-only repairs.
NEXT: Run the two failing test files, then final full check sequence.
FACT: Started targeted rerun of audio-switch and draft-adapter unit files; log /tmp/courses-ms3a-run1-targeted.log.
NEXT: Confirm repaired cases and collect any remaining RAC behavior mismatch.
FACT: Collected targeted rerun output after assertion fixes; recording result before final full-suite verification.
NEXT: Fix any remaining switch accessibility/pending assertion before final checks.
FACT: Targeted rerun: 58 tests, 55 passed / 3 failed. Inventory now passes. New-form selector must include RAC placeholder in accessible name. Row onChange fires correctly and stays pending, but aria-disabled is absent from the semantic input; inspecting installed primitive implementation to ensure pending accessibility is real.
NEXT: Correct the select query and forward pending state through supported RAC behavior without disabling/focus loss.
FACT: Installed RAC package contains compiled distribution rather than the source paths first queried; no source files changed in this inspection.
NEXT: Read the compiled switch hook/component to choose its supported pending representation.
FACT: Located package entrypoint manifests to resolve the actual compiled switch files; investigation remains frontend dependency read-only.
NEXT: Inspect the resolved implementation and finish pending accessibility repair.
FACT: Read compiled Switch export/hook paths; pending UI must retain focus and avoid RAC isDisabled. Targeted assertions are intentionally checking the semantic input rather than only the label.
NEXT: Use supported read-only input semantics and explicit input pending ARIA if the hook omits it.
FACT: Installed useToggle supports isReadOnly (aria-readonly) and preserves focus, while RAC filters aria-disabled away from the input. A small existing-primitive enhancement may be needed to expose requested pending ARIA correctly.
NEXT: Add pending support at the Switch wrapper seam, keeping the library and all call sites stable.
DECISION (2026-11-16, Codex): Installed RAC 1.21.0 filters aria-disabled from useToggle input props. Keep existing Switch; use supported isReadOnly while pending and its inputRef to put aria-disabled on the semantic input. This avoids isDisabled focus loss and avoids changing the shared primitive. Check: row Space/pending/error test asserts input ARIA, duplicate suppression and focus.
FACT: Corrected select accessible-name query; format exit 0.
NEXT: Re-run the affected unit files and verify actual pending input behavior.
FACT: Started second targeted rerun after the pending-input repair.
NEXT: Collect targeted result, then re-run format/typecheck/lint/full npm test on final bytes.
FACT: Collected second targeted result and inspected any remaining failures; final full-suite run follows once targeted behavior passes.
NEXT: Confirm final source state and execute final ordered frontend checks.
FACT: Second targeted rerun exit 0: 58/58 tests, 2 files, 0 skipped. Pending semantic-input ARIA and Space/focus/error behavior pass.
FACT: Added explicit true/false mint HTTP-body tests through the real adapter/generated client to complement query-port coverage. Final format exit 0.
NEXT: Run final typecheck → lint → full npm test; no further implementation planned unless a check finds a defect.
FACT: Final npm run typecheck started after adding mint wire-body coverage.
NEXT: Collect final typecheck exit, then lint and full unit suite.
FACT: Collected final typecheck result and proceeded to lint on success.
NEXT: Final full npm test, then handoff with explicit unrun browser/full gates.
FACT: Final npm run typecheck exit 0; final npm run lint exit 0 (238 architecture files). Final full npm test started; log /tmp/courses-ms3a-run1-tests-final.log.
NEXT: Collect final exit/counts and record exact changed paths/snapshot for Courses 1.
FACT: Final git diff --check exit 0; inspected domain guard, switch fields and route/save wiring. Scope remains frontend plus this append-only log; generated/backend/e2e have no new tracked diffs from this run.
NEXT: Capture terminal full-suite summary and final file fingerprints, then return concise handoff.
FACT: Collected terminal full-suite output; final counts and handoff recorded in the next append after confirming the summary.
NEXT: Record final verification, changed files and remaining Courses 1/run-2 ownership.

FACT: FINAL verification on 2026-11-16: frontend npm run format exit 0 (395 files); npm run typecheck exit 0; npm run lint exit 0 (238 architecture files); npm test exit 0 — 103 files, 1020/1020 tests passed, 0 failed/skipped/todo. git diff --check exit 0. Full-suite log: /tmp/courses-ms3a-run1-tests-final.log (Rstest timestamp 2026-11-16T15:50:14.729Z).
FACT: Implemented listDrafts with strict item validation and track snapshots, strict transcript inventory and non-polling lesson inventory query, successful-save inventory invalidation, draftOfTrack helper, isolated lessonId:trackId markers surviving cache GC, O7 stale-no-poll, scoped audio-write invalidation, pure draftLoadable guard, multilingual default/mint/PATCH switches and en/uk copy. Default-first route bridge is intentionally at the UI boundary, marked // MS3a run 2: tabs.
FACT: No commit, checkout/reset/stash/clean/rm, generated edits, backend edits, e2e edits, browser run, build/Compose/integration/full make gate or independent review performed. Existing unrelated WIP preserved. Browser/focus/layout in a real browser and full ordered gates are NOT_TESTED here; Courses 1 owns them.
FACT: Run-1 changed-file inventory (all frontend source/tests below plus this append-only working file). SHA-256 fingerprints apply to the final frontend bytes; unrelated pre-existing dirty files are outside this inventory.
FACT: Snapshot root=/Users/strelia/projects/courses; HEAD=a985543ffac3a5ed9d5f9e34bee9848245325a7f; frontend tracked diff SHA-256=364d7e93b06ebbea8862938ef031b94a7e1cffed339fa25af1ac866e47052c19.
FACT: `frontend/src/app/composition/admin-audio-queries.test.ts` SHA-256 `98ed3c0470d644946863cef4cd68b051d8f0458046cac7a57ff7b3f320394353`.
FACT: `frontend/src/app/composition/admin-audio-queries.ts` SHA-256 `a60316fab85c4098f1ebe1e9f7744883239ee80f385dbb0acacf283593745046`.
FACT: `frontend/src/app/composition/admin-draft-queries.test.ts` SHA-256 `7153851162c1da70f601c1909f3e67213ce68e6bb2e4cfaed6dbe4c66c2595fc`.
FACT: `frontend/src/app/composition/admin-draft-queries.ts` SHA-256 `32ee45eaba7b9b6290e10daf1cdba3a8fc3bea2f6336dfa21b047a6006b32c10`.
FACT: `frontend/src/app/composition/admin-draft-s6.test.ts` SHA-256 `1ff42df9c23478aff075a0d041ff351f0880a3e1a7391d8cf089b979e7822923`.
FACT: `frontend/src/app/providers/locales/en.ts` SHA-256 `d20512f6500072ed58cab07b763fc79b29fd05239f9aa4dd5679f7d0410985dd`.
FACT: `frontend/src/app/providers/locales/uk.ts` SHA-256 `5a25d2439cb525cbe11db81a097b73c8cfe5dbfb17c2a2810075825d2541ca94`.
FACT: `frontend/src/modules/admin/ui/CourseEditorView.test.tsx` SHA-256 `bf70bf50f3a61bc14fa3b9335b232c2a6307aee925180139f699de647d6ecc91`.
FACT: `frontend/src/modules/media/adapters/audio-track-repository.test.ts` SHA-256 `6cd7e7a7557bd2d36e3ac75f5d372cfbe87615ecd7e5a5f5272d88a3691f0785`.
FACT: `frontend/src/modules/media/adapters/audio-track-repository.ts` SHA-256 `b95ec5832eed19bebfeb8e785eee848e23ef6fd6b9fdc2895e3b0306c0d2ebda`.
FACT: `frontend/src/modules/media/application/audio-tracks.ts` SHA-256 `27d7089a8e06495a7a023ecd8a722ad21a4295f30ccb1dc6d7a703d168939108`.
FACT: `frontend/src/modules/media/ports/AudioTrackRepository.ts` SHA-256 `184aa52c5c8e225fde509274b27b9a873b3a9b4044884d82af9098dc5aee6e9b`.
FACT: `frontend/src/modules/media/ui/LessonAudioTracksField.test.tsx` SHA-256 `2eb6ed115767b1be488dab755e304734698e4a38e9966826fa2bb28bfe41d502`.
FACT: `frontend/src/modules/media/ui/LessonAudioTracksField.tsx` SHA-256 `314b1a56bb677c3676b88a045a379003ce561098373ca7cffe744c757bfe4e43`.
FACT: `frontend/src/modules/transcript/adapters/draft-repository.test.ts` SHA-256 `4f08b3e4e5dae5f8351ac8c6a4924833677a8136825e9151fd2995bc9a4801c1`.
FACT: `frontend/src/modules/transcript/adapters/draft-repository.ts` SHA-256 `0325995baa5c4aaf86bb9053a556e7656c9d4ce91bf3fb459700a5e3a3b87120`.
FACT: `frontend/src/modules/transcript/domain/draft.test.ts` SHA-256 `dc559128cfeb5b81c7e73b3a259540630768eb151331a9b3b07a3a93783b2b8c`.
FACT: `frontend/src/modules/transcript/domain/draft.ts` SHA-256 `e12abd31a2feb5dabe3e433d890f6dfec8946b95b73ef5e71150047995cdaccf`.
FACT: `frontend/src/modules/transcript/ports/DraftRepository.ts` SHA-256 `7356493c12a169757d687545429c7a143aa6b5672e89a0be8d785e415c8487e9`.
FACT: `frontend/src/modules/transcript/ui/LessonDraftField.test.tsx` SHA-256 `5a2a410280b8c6c4938805921f6945aae1a67b9e5fd6445568035847d0fd1283`.
FACT: `frontend/src/routes/admin.courses.$courseId.tsx` SHA-256 `ab0e966ee6997d130adccd33f52c3fd9f648f68c6fce9012c50d16f73ec9cd51`.
FACT: `frontend/src/routes/route-guards.test.tsx` SHA-256 `e68cbf22649a8a205b592b8f2264b9337e52b936ad1c21c05b48077959fa30f9`.
DECISION: No unresolved plan question for run 1. The inventory operation name correction and RAC pending mechanism are recorded above with their checks.
NEXT: Courses 1 can review this snapshot/run browser and full gates, then run 2 adds language tabs and wires draftLoadable. MS3b retains the checkpoint exclusions (und confirmation/PATCH→PUT, detached snapshot/flush and baseLanguage recovery). This run made no claim to close MS3a or its browser/e2e acceptance rows.
