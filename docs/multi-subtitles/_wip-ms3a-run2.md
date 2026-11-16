NEXT: Read repository instructions, accepted MS3a requirements and run 1 checkpoint; implement language tabs and save guards within the authorized frontend scope.
FACT: Read repository instructions and run 1 sources; the initial combined read was truncated, so focused reads follow before implementation. No source changes yet.
DECISION: Reuse existing editor/query/React Aria paths, following ponytail; no new dependencies, backend, generated files, e2e or git mutations.
NEXT: Inspect focused accepted decisions, the composition wiring and existing editor components.
FACT: Confirmed MS-D6/7 and recovery-key contract. Run 1 keeps a first-draft bridge in the route; CourseEditorView currently mounts draft/transcript siblings within its course form.
DECISION: Resume the accepted MS3a slice (below a new full-cycle threshold); Courses 1 remains review/full-gate owner. Preserve the existing no-Ready editor and move shared tab orchestration into admin UI with injected repository reads.
NEXT: Complete component/API tracing, then add instance identity, base-state handling and tab wiring.
FACT: Root/branch/HEAD match the brief: /Users/strelia/projects/courses, ms2-3-asr-per-track @ a985543ffac3a5ed9d5f9e34bee9848245325a7f. Existing dirty frontend files are run 1; unrelated WIP is preserved.
FACT: TranscriptEditor uses sections and button actions (no nested editor form). Only en/uk editor locales exist. The route currently invalidates inventory only after saving.
DECISION: Keep optional tab services at the admin boundary; inject fresh-read functions from the route. The editor retains its base token while dirty; replacement confirmation captures the displayed inventory token. MS3b owns detached recovery, snapshot baseLanguage and und confirmation.
NEXT: Implement the scoped component changes and direct-read wiring; then typecheck and repair existing tests before adding new coverage.
FACT: Revalidated the existing READY checkpoint and caller paths; optional graph tooling is unnecessary for this bounded frontend slice. No open requirement question; author checks remain separate from Courses 1's review/full gates.
DECISION: 2026-11-16, Codex: preserve legacy fallback DOM/copy, use tab identity as editor key, and use the existing Dialog primitive for version-specific replacement confirmation. Check via component tests and the requested frontend gate.
NEXT: Apply implementation.
FACT: Implemented tab assembly/order, per-track draft selection/Regenerate and guard, fixed-language editor props, per-tab DOM/recovery identity, direct save checks and replacement dialog. Local inert-panel CSS and both save invalidations are wired.
FACT: First locale-write command used the wrong working directory and made no locale changes; corrected it here. Initial frontend format completed exit 0; typecheck is collecting integration errors.
NEXT: Inspect typecheck, finish locale formatting and repair synchronization/token behavior before unit tests.
FACT: Initial typecheck found a local variable shadow in the recovery baseline; fixed it. Tightened base synchronization so a successful save does not reset lines from unchanged old props.
NEXT: Re-run format/typecheck/lint, then existing unit suite and focused new tab/save tests.
FACT: Second format exit 0; second typecheck has no reported errors. Started lint and the full existing frontend unit suite, with logs under /tmp/courses-ms3a-run2-*.log.
NEXT: Collect actual exits; update obsolete draft-mismatch assertions and add regression tests for tab identity and fresh-save guards.
FACT: Lint and npm test both exited 1; collecting their logs before drawing conclusions. The initial typecheck error was repaired and a second process completed.
NEXT: Resolve lint findings and any unit-run startup restriction, then finish focused coverage.
FACT: Typecheck exit 0. Initial lint exit 1 (hook dependency + set-state-in-effect); initial npm test exit 1 before tests (sandbox EPERM ::1:3000).
FACT: Found and fixed a rename hazard during tracing: a new-language query could temporarily unmount the editor. Keep previous query data while loading and pass its actual base language separately from the fixed save language.
FACT: Invalidated draft offers are cleared after a track language/mode change, so the second Load affordance cannot bypass draftLoadable.
NEXT: Re-run authorized unit suite with listener access, and repair obsolete expectations/add targeted tests.
FACT: Escalated frontend unit suite started successfully (same command, listener access only). No full/e2e gate started.
NEXT: Add focused tab/save regression coverage and collect baseline unit failures.
FACT: Added tab/order/fallback/DOM, save-language/new-id, rename/delete/read-failure zero-PUT, base/token, orphan, und and draft-guard tests. Updated the explicitly removed languageMismatch test branch.
NEXT: Run the final requested command sequence, fix failures and record exact counts; remaining browser/full gates belong to Courses 1.
FACT: Focused tests are now present; collecting the in-flight baseline suite before starting a second run to avoid concurrent unit writers/listeners.
NEXT: Format and typecheck the final test inputs, then lint and rerun units sequentially.
FACT: Baseline full unit suite completed exit 1 with the removed languageMismatch expectation failing; that test is now updated. New coverage was added after that run began and is not yet claimed green.
NEXT: Complete format/typecheck/lint for final sources and execute the full frontend unit suite again.
FACT: Final-test typecheck found two fixture-only issues (missing AudioTrack.label and unsupported ByRole exact option); corrected them. Baseline counts were 1018 passed / 2 failed, so the earlier single-failure wording was incomplete; inspecting both failures.
NEXT: Inspect lint and remaining baseline failure while the new suite runs; then repeat the ordered frontend commands on repaired sources.
FACT: Baseline's second failure was the stale-draft copy expectation; updated to the mandated guard message. Lint now passes (239 architecture files).
NEXT: Collect new suite failures, add recovery-token/key and true per-track pending regressions if practical, then finish ordered verification.
FACT: New-suite result: 1035 passed / 3 failed (exit 1). Two new failures selected hidden orphan editors with getByLabelText after rename; switch tests to accessible role/name selectors, which scope to the visible editor. Third was the already-corrected stale-copy expectation.
NEXT: Add remaining isolated key/restore and pending tests, then rerun all requested frontend checks.
FACT: Correcting failure classification: F01 was hidden-label selection; F02 was the und draft flow (inspecting its exact error), F03 stale copy. No claim that these failures are all fixture-only until inspected.
NEXT: Resolve the und flow failure and complete verification.
FACT: Und flow failure was an incorrect test button name (actual copy is “Load draft into editor”); corrected the positive and negative guard selectors so they test the real affordance.
FACT: Added R13b restored-token, isolated recovery keys, actual in-flight per-track pending, keyboard navigation and already-offered draft invalidation checks.
NEXT: Finish inventory-only known-state wiring and re-run the ordered frontend gate.
FACT: onTranscriptKnown now reads inventory in the fallback too. Replacement uses controlled React Aria Modal/Dialog (the existing wrapper only supports trigger-driven opening).
DECISION: 2026-11-16, Codex: disable line editing/add/remove during an in-flight Save so successful query refresh cannot overwrite edits made after the captured PUT. Mechanism: disabled editor controls until completion; test through pending actions. Unexpected thrown save errors also restore a retryable state.
NEXT: Execute frontend format → typecheck → lint → npm test on the completed slice.
FACT: Latest format completed exit 0; collecting latest typecheck before lint. New coverage now includes recovery-key isolation and stale token restoration without adding deferred snapshot baseLanguage.
NEXT: Continue the ordered frontend gate and inspect failures only if produced.
FACT: Latest typecheck exit 0; lint exit 0. Started npm test after both passed (log /tmp/courses-ms3a-run2-tests-verified.log).
NEXT: Collect unit exit/counts and inspect scoped diff for integration gaps before handoff.
FACT: Scoped diff check has no whitespace errors; confirmed tab services flow through CourseEditorView → ModuleFields → LessonFields and the route sends ?language via transcript query options. No tab editors introduce another form.
NEXT: Collect the full unit result, then record final files/hashes and any unverified rows for Courses 1.
FACT: Full frontend unit suite completed exit 0. Collecting its counts and final file inventory; the requested ordered format/typecheck/lint/test sequence is green.
NEXT: Document acceptance coverage, unchanged limitations and final handoff without committing.
FACT: Final ordered frontend verification at a985543 + preserved run 1 + run 2 WIP: `cd frontend && npm run format` exit 0 (397 files); `npm run typecheck` exit 0; `npm run lint` exit 0 (239 architecture files); `npm test` exit 0 (104 files, 1043 passed, 0 failed/skipped/todo). Unit suite required listener access after the documented sandbox startup failure. Logs: /tmp/courses-ms3a-run2-{format,typecheck,lint}.log and /tmp/courses-ms3a-run2-tests-verified.log. `git diff --check` exit 0.
FACT: New LessonLanguageTabs tests cover M3-R1 order/orphan/fallback, M3-R2 fixed lowercase PUT language, M3-R3 guard and invalidated offer, M3-R5 Regenerate without Select and isolated pending, M3-R7 mounted unique ids/inert/accessibility/keyboard/single course form, M3-R10 direct-read refusal/retry with zero PUT and kept lines, M3-R11 fresh ids each PUT + clean/dirty rename, M3-R13 a/b/c/d token paths, per-tab recovery keys, inventory-known and und-save zero-PUT. Existing tests were updated only for the removed mismatch branch and mandated guard copy.
FACT: Run 2 changed the following 13 frontend files (SHA-256 of final bytes; run-1 changes in shared files retained):
FACT: `frontend/src/modules/admin/ui/LessonLanguageTabs.tsx` SHA-256 `653a0edc9aa9f81d3d28e91276a89188dd7f192a7cea4959cf0023b6b76abcab`.
FACT: `frontend/src/modules/admin/ui/LessonLanguageTabs.test.tsx` SHA-256 `50b47090ed28cf211b8d63ba89eb36ce5a6fd96375b18a9728013d8c30657761`.
FACT: `frontend/src/modules/admin/ui/CourseEditorView.tsx` SHA-256 `4e11966249f1c9339ba2e9580fdaa666234a4f8511c9ca0aeac28ee0115634e1`.
FACT: `frontend/src/modules/transcript/ui/LessonDraftField.tsx` SHA-256 `c5ab86ff38e890b866e6e779d977065143be2a46758d23b8cbb07c0aebdc631c`.
FACT: `frontend/src/modules/transcript/ui/LessonDraftField.test.tsx` SHA-256 `5820f0f3672fe4217aebcd3e37bad904452d8a6f1e1d843f073130370ef34d45`.
FACT: `frontend/src/modules/transcript/ui/LessonTranscriptField.tsx` SHA-256 `26a5a4ccc0d79578dd7756a8315ed51e37d26889bd64d483147e471f96d3068e`.
FACT: `frontend/src/modules/transcript/ui/TranscriptEditor.tsx` SHA-256 `dda59f149d6f1d97ad43657c78c7bc215944d5f75c271a7ac49dafc9a3289329`.
FACT: `frontend/src/modules/transcript/ui/TranscriptEditor.test.tsx` SHA-256 `781ca315a81a99e3da47a8f26ea2d6f7b942369d29a1765cb40be207c3824675`.
FACT: `frontend/src/modules/transcript/application/recovery.ts` SHA-256 `d51da1ee211c6849b63f4ccf67b8fc474a7b796582aec1d8a48cfc43c89139f3`.
FACT: `frontend/src/app/providers/locales/en.ts` SHA-256 `80df436141f5faff6a4046a22bb2c1fb49e39485d9adab4767e3a31ea6a61cf2`.
FACT: `frontend/src/app/providers/locales/uk.ts` SHA-256 `83d7b14732b0bc256b12512f185f7c92d15f4bb807d816db98c0711f7ac0a2e0`.
FACT: `frontend/src/app/styles/main.css` SHA-256 `a0daf70f59d9b1a3c6a6637ebaae8ff6bc9378ea851addf36c62bf521728bdf4`.
FACT: `frontend/src/routes/admin.courses.$courseId.tsx` SHA-256 `7cd33bd7754ed4e0d8616e27ac2281fe1fcbbfb995c2da59c3003da2ca24fdd4`.
FACT: Working record: docs/multi-subtitles/_wip-ms3a-run2.md (append-only throughout). No commits, checkout/reset/stash/clean/rm, backend, generated or e2e changes by this run. Existing unrelated WIP remains untouched.
FACT: Browser/design viewport checks, unchanged transcript.spec execution, backend/integration/build/full e2e and independent implementation review were NOT RUN by this run. Unit fallback checks preserve its known ids/roles/copy but do not claim the browser test passed.
FACT: MS3b exclusions remain: und confirmation/PATCH→PUT, detached snapshots/flush on unmount, and snapshot baseLanguage. The restored-token unit proves the same-language T1/T2 case only; renamed/legacy snapshot base ownership remains MS3b, as accepted at checkpoint.
QUESTION: None open for this run's implementation scope.
NEXT: Courses 1 reviews the final dirty snapshot, runs authorized browser/full gates and records phase-level evidence. Optional remaining unit depth: route-level assertions of repository error-result conversion and query invalidation; latest-Ready selection with a newer processing replacement; cancellation of replacement confirmation. No claim that aggregate green closes every MS3a browser acceptance row.
