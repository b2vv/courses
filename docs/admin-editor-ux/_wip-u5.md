NEXT: Read the current shared entry rules, repository routes, U5 contract, and existing save/status behavior before changing implementation.
FACT: Read the shared agent entry rules; no delegation is needed for this scoped implementation.
NEXT: Verify repository routes and read the accepted U5 plan plus current task record.
FACT: Verified repository root <repo>; read the briefing and local AI Factory adapter/routes. The previous combined read was output-truncated, so the task contract and local route are being read separately.
NEXT: Confirm branch/HEAD, accepted checkpoint, exact U5 acceptance, and relevant skills before editing product files.
FACT: Branch admin-editor-ux and HEAD 10c3c6ea4c145a8fcd75340f8cea9bb607256bf2 match the owner snapshot. Existing unrelated dirty/untracked files are preserved. Accepted r3 §9 marks U5 READY; U1–U4 are committed predecessors.
DECISION: Scope is the existing U5 slice (Ux11/Ux12, D5a/D7/R4), with no new contract or architecture. Reuse existing RHF dirty state, save handler, translations, and lesson status header. Owner retains browser and independent review; no e2e/Docker/commits/staging/branch changes in this run.
NEXT: Trace course save, immediate-apply panels, scrolling shell, and existing tests/status selectors; then write failing component tests.
FACT: Read repository standards and canonical Ponytail/design-pass guidance. This is a small, frontend-only accepted slice. Browser/design proof is explicitly assigned to the owner, so this run will report implementation and component checks separately.
FACT: Searched all frontend/e2e status-role selectors: editor assertions in admin.spec.ts filter by specific text; subtitles/discussion status assertions are panel-scoped. No unscoped editor status count was found.
NEXT: Read the course form and immediate-apply components plus CSS scroller rules, then pin the minimal implementation and test seams.
FACT: Existing lesson-language header already uses administration.savedImmediately; retain it without duplication. Price uses a separate form and onSetPrice; its own Save persists immediately without the course Save. Course form visibility already follows tab === access.
NEXT: Inspect grant/material mutation handlers and the real page scroller; choose a height-aware focus inset rather than assuming a fixed bar height at enlarged text.
FACT: Materials actions.add/remove and access onGrant are independent persistence actions; price has a separate Save and onSetPrice. All three qualify for the existing Saved immediately hint; existing lesson headers are unchanged.
DECISION: 2026-11-18, Codex (U5 implementation): keep the save-bar status region empty when clean. This avoids claiming that independent panel drafts or failed operations are saved; dirty state says Unsaved changes. Check: locale-parametrized clean→dirty→clean/save tests scoped to the bar.
DECISION: Use a ResizeObserver on the bar to update a scoped CSS variable on the document scroller, since wrapping/200% text changes its actual height. Restore the previous property on cleanup; Access gets zero inset. Unit proves the hook/observer wiring only; R4 physical visibility remains owner browser proof.
NEXT: Add failing Ux11/Ux12 tests and run them before product changes.
FACT: Existing tokens define control minimum 2.75rem (44px at baseline), surface and border colors; existing editor uses the document/page scroller rather than an independently scrolling form. No new dependency or theme is needed.
NEXT: Add locale-parametrized dirty/save/Access assertions and heading-local immediate-save assertions; run only the new U5 tests for the required red proof.
FACT: Added U5 tests before implementation: English/Ukrainian clean→dirty→submit feedback, single polite bar status, tab persistence/Access hiding, no Discard, CSS sticky/scroller hook, and per-h2 immediate-save hints for materials/price/access.
NEXT: Run the new U5 tests and record their actual pre-implementation failures.
FACT: First red-test command was issued from repository root (not frontend), exit 1; inspect output before attributing this to assertions.
NEXT: Run the U5 red proof from the frontend directory.
FACT: Frontend red command: npx rstest run src/modules/admin/ui/CourseEditorView.test.tsx --testNamePattern 'U5 save feedback' --testTimeout 30000; exit 1. Detailed output: /tmp/courses-u5-red.log.
NEXT: Resolve only fixture mistakes if any; otherwise implement the missing U5 behavior proven by these failures.
FACT: The frontend U5 red run has been polled; its terminal result is recorded above when the command completes. The earlier root command failed in npm cache access before tests and is not assertion evidence.
NEXT: Use the completed assertion failures to guide the product edit.
FACT: Red run completed: 5 U5 tests failed (2 missing bar statuses, 1 missing sticky/scroller CSS, 2 missing header hints); exit 1. No implementation was present during the red proof.
FACT: Added the sticky token-colored bar, one empty-when-clean polite status, en/uk dirty text, ResizeObserver height→root CSS inset with cleanup, narrow stacked layout, and existing immediate-save text directly under course-materials/price/access h2. Existing Save behavior and lesson headers are retained; no Discard.
NEXT: Run the scoped U5 tests and fix genuine failures, then execute the exact requested frontend verification sequence.
FACT: Scoped U5 post-implementation command exited 1; detailed output /tmp/courses-u5-green.log.
NEXT: Inspect the scoped result and proceed to the ordered frontend checks after any necessary corrections.
FACT: Polled the scoped U5 post-implementation test run; its exit and output are retained in this record and /tmp/courses-u5-green.log.
NEXT: Correct any locale/fixture issue before broad verification; inspect the final diff for scope and scroller behavior.
FACT: Post-implementation scoped run: 4 passed, 1 failed. The remaining failure was my Ukrainian test fixture heading (Надати instead of the existing Видати), not product behavior; corrected the fixture to existing locale copy.
NEXT: Run the requested format → format:check → typecheck → lint commands in order, recording each exit separately; then run the full src unit suite.
FACT: Requested verification npm run format: exit 0 (PASS); log /tmp/courses-u5-check-1.log.
NEXT: Continue the requested ordered frontend verification; investigate any failure before claiming completion.
FACT: Requested verification npm run format:check: exit 0 (PASS); log /tmp/courses-u5-check-2.log.
NEXT: Continue the requested ordered frontend verification; investigate any failure before claiming completion.
FACT: Requested verification npm run typecheck: exit 2 (FAIL); log /tmp/courses-u5-check-3.log.
NEXT: Continue the requested ordered frontend verification; investigate any failure before claiming completion.
FACT: Requested verification npm run lint: exit 1 (FAIL); log /tmp/courses-u5-check-4.log.
NEXT: Continue the requested ordered frontend verification; investigate any failure before claiming completion.
FACT: Polled the ordered format/static-check run; each completed command records its own exit above. Inspected diff statistics to ensure formatter changes stay within U5 files.
NEXT: Run npx rstest run src --testTimeout 30000 after the static checks finish successfully.
FACT: Initial requested static sequence: format PASS (0), format:check PASS (0), typecheck FAIL (2: unsupported exact option in my getByRole test), lint FAIL (1: prefer output over explicit status role). Corrected the test option and changed the bar feedback to semantic output (implicit role=status), retaining explicit polite/atomic live semantics.
NEXT: Rerun the full requested sequence on corrected bytes; retain the initial failures as history.
FACT: Corrected-snapshot verification npm run format: exit 0 (PASS); log /tmp/courses-u5-final-1.log.
NEXT: Continue verification, then inspect scoped diff and record remaining owner checks.
FACT: Corrected-snapshot verification npm run format:check: exit 0 (PASS); log /tmp/courses-u5-final-2.log.
NEXT: Continue verification, then inspect scoped diff and record remaining owner checks.
FACT: Corrected-snapshot verification npm run typecheck: exit 0 (PASS); log /tmp/courses-u5-final-3.log.
NEXT: Continue verification, then inspect scoped diff and record remaining owner checks.
FACT: Corrected-snapshot verification npm run lint: exit 0 (PASS); log /tmp/courses-u5-final-4.log.
NEXT: Continue verification, then inspect scoped diff and record remaining owner checks.
FACT: Reviewed the scoped U5 diff: no Save handler, persistence contract, lesson-section header, e2e, or backend changes. The only new live region is the output inside .admin-editor-save; hints are plain paragraphs. Root inset is scoped by visible form and observer cleanup restores its prior custom property.
NEXT: Await the corrected ordered checks; if green, record exact unit totals and owner browser/review handoff without claiming R4 browser PASS.
FACT: Corrected snapshot format, format:check, typecheck, and lint have each passed (exit 0). Full src unit run is in progress; no Docker or e2e command has been run.
NEXT: Wait for the full suite result, then record acceptance limits and final scoped fingerprints.
FACT: Corrected-snapshot verification npx rstest run src --testTimeout 30000: exit 1 (FAIL); log /tmp/courses-u5-final-5.log.
NEXT: Continue verification, then inspect scoped diff and record remaining owner checks.
FACT: Polled the full src suite after the four green static checks; no test verdict is assumed until the process returns.
NEXT: Finish the full-suite evidence and leave the 200%/320px focus check to the owner session as requested.
FACT: First corrected full-suite command exited 1. It surfaced an existing unit assertion that queried any status on the whole editor and expected none; the new clean bar output correctly remains a status region. e2e selectors were scoped, but this unit selector was not.
NEXT: Inspect all full-suite failures and scope the affected assertion to the specific transcript notice it is intended to test, preserving its behavior requirement.
FACT: Full-suite result was 1477 passed / 2 failed. Both failures are C19 transcript-notice tests using global status queries. They must select the course-status section, matching AGENTS Decisions bullet 2, rather than treating a new independent status as a transcript notice.
NEXT: Scope both C19 queries to their existing Course status region, then rerun all requested checks on the final test snapshot.
FACT: Scoped both existing C19 status assertions to the Course status region, preserving both presence and absence requirements for the transcript notice. Product code was unchanged after the four static-check passes.
NEXT: Execute the requested full sequence once more because test files changed; final acceptance will cite only this snapshot's result.
FACT: Final-snapshot verification npm run format: exit 0 (PASS); log /tmp/courses-u5-verified-1.log.
NEXT: Complete final evidence and owner handoff; investigate any remaining failure.
FACT: Final-snapshot verification npm run format:check: exit 0 (PASS); log /tmp/courses-u5-verified-2.log.
NEXT: Complete final evidence and owner handoff; investigate any remaining failure.
FACT: Final-snapshot verification npm run typecheck: exit 0 (PASS); log /tmp/courses-u5-verified-3.log.
NEXT: Complete final evidence and owner handoff; investigate any remaining failure.
FACT: Final-snapshot verification npm run lint: exit 0 (PASS); log /tmp/courses-u5-verified-4.log.
NEXT: Complete final evidence and owner handoff; investigate any remaining failure.
FACT: git diff --check passed (exit 0). Reviewed the C19 scoping change against its original notice requirements. No product changes followed the previous successful static checks; only the test selectors changed.
NEXT: Bind final full-suite outcome to scoped hashes and hand off owner R4 browser validation and independent review.
FACT: Final-snapshot verification npx rstest run src --testTimeout 30000: exit 0 (PASS); log /tmp/courses-u5-verified-5.log.
NEXT: Complete final evidence and owner handoff; investigate any remaining failure.
FACT: Final scoped snapshot at admin-editor-ux @ 10c3c6ea4c145a8fcd75340f8cea9bb607256bf2 (SHA-256):
FACT: frontend/src/modules/admin/ui/CourseEditorView.tsx — 38f269c49ff3bee9541d60a344bf08dce1c32712bc03fb694065278246431ce6
FACT: frontend/src/modules/admin/ui/CourseEditorView.test.tsx — 8eb17bfbef0bf67948db509ec4ce133abcad7524f448b171fc07fe7f65d162a1
FACT: frontend/src/app/styles/main.css — f5e9a31769794de799fc04d1a01831f47e5bc98919090e9f1136a9d51a42667a
FACT: frontend/src/modules/admin/ui/DocumentLibrary.tsx — c2c5b2fd2ef768b94ca252bdb3b8608899906ad6a87188d912b8fdb087d8f6ca
FACT: frontend/src/modules/admin/ui/CoursePriceField.tsx — 1d7eb047d213daa66d1e5235f8eb7b09c75d20598a27810972bbca96ec7fd879
FACT: frontend/src/modules/admin/ui/CourseGrantPanel.tsx — 52570a40a9d26cb3dcab44109a6637fab53bac344389ecafa035ae51d3987583
FACT: frontend/src/app/providers/locales/en.ts — db122ca5d7a75b8b43ebc8add3972c3f85f1e4df1c92d78ef9ddd9178c505849
FACT: frontend/src/app/providers/locales/uk.ts — d360d1e645649c8ee1e0172adc251d0d82dd4b66270512f72029a757aa9341af
NEXT: Wait for the final full-src unit verdict; no further product edits are planned.
FACT: Final requested sequence PASS, each exit 0: npm run format; npm run format:check; npm run typecheck; npm run lint; npx rstest run src --testTimeout 30000. Full suite: 1479 passed, 0 failed, 0 skipped. Five added U5 cases are included. Final run logs: /tmp/courses-u5-verified-{1,2,3,4,5}.log.
FACT: Ux11 component/CSS-hook evidence PASS: English/Ukrainian dirty text, clean blank output, exactly one polite bar status, existing Save submits, Lessons retains dirty state, Access hides the form/bar, sticky and document-inset CSS hooks. Ux12 component evidence PASS: materials/price/access each show one hint immediately below h2; lesson-section hints reused unchanged.
FACT: Final frontend scope is eight files (two locale files, main.css, CourseEditorView.tsx/test, DocumentLibrary.tsx, CoursePriceField.tsx, CourseGrantPanel.tsx). Existing unrelated WIP is preserved. No staged files, commits, branch changes, Docker, or e2e runs were made; HEAD remains 10c3c6ea4c145a8fcd75340f8cea9bb607256bf2.
DECISION: Implementation and requested frontend verification are complete; Ux11 R4 browser proof and independent slice review are NOT_TESTED in this writer run, not claimed PASS. No phase closure is claimed.
NEXT: Owner/reviewer: inspect this exact dirty snapshot; at 200% and 320px, keyboard-focus the last fields on Overview and Lessons, verify the measured root inset follows wrapped bar height and no focused field/outline is obscured, then verify Access has no visible bar. Run e2e only in the separately authorized owner window.
NEXT: Verify the independent U5 review, current snapshot, and observer behavior; add failing size-change/cleanup tests before fixing SPEC-1. Preserve the no-e2e/Docker/commit/staging constraints.
FACT: Re-read shared entry and local routes. The first canonical receiving-code-review skill path was absent; locating its actual installed path before applying it. Current task remains U5 review fixes only.
NEXT: Verify SPEC-1 against observer and CSS, read the receiving-review guidance, and add a controlled observer regression test.
FACT: Verified SPEC-1 in source: update() writes only the height variable; there is no active-field overlap check, and the clean status output is empty with no reserved height. Branch/HEAD still match 10c3c6ea; independent review is read and its two requested fixes are concrete and compatible with D7/R4.
DECISION: Accept SPEC-1 and STD-1 for this scoped follow-up. Reserve at least one line in the output while clean, retaining blank clean feedback, and correct focus only when a resized bar overlaps the focused form field (excluding controls inside the bar). Owner browser proof remains required.
NEXT: Write deterministic observer tests for changed inset, overlapping versus clear/outside/bar focus, and disconnect/inline-property restoration on unmount; add a CSS reservation hook assertion. Run these before implementation.
FACT: Loaded canonical receiving-code-review guidance at skills/arch/receiving-code-review/SKILL.md. This is a local correction under the accepted U5 checkpoint, with no contract or scope change; browser zoom/layout proof is still not represented by unit stubs.
NEXT: Use the existing Rstest mocking patterns to drive a real effect observer with deterministic rectangles, keeping mocked geometry separate from browser evidence.
FACT: Added two regression tests before product edits: reserved clean-line CSS, and controlled ResizeObserver callbacks checking changed height before scroll, no overlap/boundary/outside/bar-control behavior, disconnect and prior root-property restoration. The test restores its global observer and mocks in finally.
NEXT: Run SPEC-1 tests against the unchanged implementation for red proof.
FACT: SPEC-1 pre-fix test command exited 1; raw log /tmp/courses-u5-spec1-red.log. Product files remain unchanged during this run.
NEXT: Inspect red failures, then apply the reserved line and overlap correction.
FACT: Polled the SPEC-1 pre-fix regression run; use its recorded exit and named failures, not the earlier aggregate PASS, as evidence for this delta.
NEXT: Implement only after confirming both failures discriminate the reported defects.
FACT: Red proof confirmed exit 1 with 2/2 failures: missing reserved-line CSS and zero overlap-triggered scroll calls. No fixture/type/runtime error caused those failures.
FACT: Fixed SPEC-1: output reserves 1lh while clean; observer writes the current bar height before checking focused-field overlap and using scrollIntoView({block:'nearest'}). Zero-height/hidden bar, outside-form focus, and bar controls do not trigger scrolling; existing disconnect/property restoration remains.
NEXT: Run the SPEC-1 regressions, then the full requested format/static/unit sequence.
FACT: SPEC-1 post-fix test command exited 0; raw log /tmp/courses-u5-spec1-green.log.
NEXT: Resolve any regression-test failure, then run the exact ordered frontend verification on final bytes.
FACT: SPEC-1 verification npm run format: exit 0 (PASS); log /tmp/courses-u5-spec1-verified-1.log.
NEXT: Finish ordered verification; investigate any failure before reporting this delta ready.
FACT: SPEC-1 verification npm run format:check: exit 0 (PASS); log /tmp/courses-u5-spec1-verified-2.log.
NEXT: Finish ordered verification; investigate any failure before reporting this delta ready.
FACT: SPEC-1 verification npm run typecheck: exit 0 (PASS); log /tmp/courses-u5-spec1-verified-3.log.
NEXT: Finish ordered verification; investigate any failure before reporting this delta ready.
FACT: SPEC-1 verification npm run lint: exit 0 (PASS); log /tmp/courses-u5-spec1-verified-4.log.
NEXT: Finish ordered verification; investigate any failure before reporting this delta ready.
FACT: SPEC-1 targeted regressions passed (2 passed, 54 filtered out; exit 0). Reviewed the updated callback and test after formatting; git diff --check passed (exit 0). Tests cover write-before-scroll ordering, non-overlap/boundary, outside form, bar-control exclusion, disconnect and prior property restoration.
NEXT: Wait for full-src suite and retain browser R4 and independent re-review as separate pending proofs.
FACT: SPEC-1 verification npx rstest run src --testTimeout 30000: exit 0 (PASS); log /tmp/courses-u5-spec1-verified-5.log.
NEXT: Finish ordered verification; investigate any failure before reporting this delta ready.
FACT: Corrected SPEC-1 snapshot passed format, format:check, typecheck and lint (each exit 0); the full-src unit result is being awaited separately. No checks were skipped or relabelled.
NEXT: Record the final full-suite result, updated scoped hashes and the review/browser handoff.
FACT: SPEC-1 final verification: npm run format PASS (0); npm run format:check PASS (0); npm run typecheck PASS (0); npm run lint PASS (0); npx rstest run src --testTimeout 30000 PASS (0): 116 files, 1481 passed, 0 failed, 0 skipped. Logs /tmp/courses-u5-spec1-verified-{1,2,3,4,5}.log. git diff --check PASS (0).
FACT: SPEC-1 final SHA-256 frontend/src/modules/admin/ui/CourseEditorView.tsx — 79ed41eacc791653072f5d91edc900ff7f155b6148dcd52329358267affc74da
FACT: SPEC-1 final SHA-256 frontend/src/modules/admin/ui/CourseEditorView.test.tsx — a064bce39b2d54cf988dca97e0b4156b7b0ad21eabb3004da41d817ab1e3ca6e
FACT: SPEC-1 final SHA-256 frontend/src/app/styles/main.css — 12b129fa2b286c842a34b3d51db84a2e2245efa98890972ed9af193fe9143c3c
FACT: Review-fix delta touches only the observer callback, reserved output-line CSS and regression tests. All other U5 files and unrelated WIP remain preserved. Branch admin-editor-ux / HEAD 10c3c6ea4c145a8fcd75340f8cea9bb607256bf2 unchanged; no staged files, commit, Docker or e2e run.
DECISION: SPEC-1 and STD-1 requested changes are implemented with red→green tests; this author run does not replace the independent review verdict or physical focus proof. The old review BLOCK remains its historical verdict until re-review.
NEXT: Independent reviewer to inspect this three-file correction; owner to validate focused-field visibility through first edit, wrapped/200% bar resizing and 320px layout. Browser evidence remains NOT_TESTED in this run.
