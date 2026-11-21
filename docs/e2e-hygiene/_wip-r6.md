NEXT: Read the shared entry rules and R6 contract, inspect the cleanup patterns, then update only the requested specs and this working file.
FACT: Read the installed shared agent entry rules before starting repository work.
FACT: Checked repository identity, branch, status, local routes, briefing, R6 plan, helper, and reference pattern; inspection output will guide the next step.
NEXT: Resolve any scoped instructions and inventory course creation and archival cleanup in all e2e specs.
FACT: Verified root <repo> and branch r6-e2e-archived-cleanup. Existing unrelated changes are preserved.
FACT: R6 explicitly requires immediate registration, allSettled/AggregateError teardown, retained archive-behaviour calls, and untouched seeds. The existing helper covers course-related FK cleanup; no contract defect identified.
DECISION: Reuse the helper and lesson-builder hook verbatim; keep changes scoped to the ten requested specs unless the all-spec creation inventory reveals an uncovered case to report.
NEXT: Read each target spec's creation and teardown paths and classify archive calls before editing.
FACT: All-spec endpoint search found creation paths in the ten targets plus admin UI creation, asr, audio-tracks, learn-access, media, playback, lesson-builder, teach-editor, refund, and subtitles. These additional paths need cleanup classification before final reporting.
FACT: Frontend scripts are oxfmt, tsc --noEmit, oxlint and the architecture checker; no e2e/Docker command is required for the authorized checks.
NEXT: Inspect target hooks and response parsing in bounded file groups, then implement immediate registration and teardown.
FACT: Ratings currently archives courses then deletes run-scoped reviewers; enrollment archives published fixtures. Both register too late (after publishing). Instructor/discussion/learning creation and teardown paths inspected next.
DECISION: Course deletion must finish before ratings reviewer deletion; collect deletion failures without skipping the existing user cleanup.
NEXT: Inspect remaining target specs and distinguish admin behaviour deletion from cleanup deletion.
FACT: Instructor/discussion/learning maintain draft/published arrays only for teardown; replace those bookkeeping paths with a single Set registered at create parsing. Their publishing, grants, comments, and assertions stay intact.
NEXT: Check admin UI-created courses and transcript's separate editor-draft path; preserve user teardown in accounts.
FACT: Transcript has three creation sites: seed (published transcript), inline video course (published), and seedDraftCourse (editor draft). Register all three, including the previously untracked editor draft.
FACT: Accounts afterAll archives only its completed course then deletes run-scoped users. Immediate registration also covers a partially created course when setup fails before assigning its module variable.
DECISION: Keep per-course create/publish response parsing and all test assertions; remove only obsolete cleanup bookkeeping and archive cleanup requests.
NEXT: Complete admin classification, then apply the ten-file cleanup change.
FACT: Admin creates via createAdminDraft, E2's response capture, and P20/P21 UI forms. P20 and P21 archive and assert archived UI state, so both archive actions stay. Admin's existing draft API cleanup stays (helper deletion is idempotent for absent rows).
FACT: Ratings Rt26/Rt30 also updates its published-token array and checks registration during republish. Preserve that existing test-body behaviour while adding the independent immediate cleanup Set.
DECISION: Small writer scope is READY: test fixture lifecycle only, no product/API/assertion changes; writer is this session, independent review remains for the requesting owner/session, not claimed here. Authorized checks are the four requested npm scripts only.
NEXT: Add immediate Set registration and allSettled teardown; capture P20/P21 create responses before their UI clicks.
FACT: The reference registers raw parsed body.id before token validation. Follow that exact order, so updatedAt validation cannot strand a valid created id.
FACT: HEAD/base is fe61852aa4d446ce155f32603e4721d9fc5d9a03. No nested AGENTS files were found under the edited scope.
NEXT: Apply response-body registration before token validation in each creation path.
FACT: Updated ten target specs: imported deleteE2ECourse, added module Sets, registered raw create ids before status/token validation, and replaced archive-only afterAll hooks with allSettled deletion and AggregateError.
FACT: P20/P21 UI creates now capture POST responses for immediate registration. E2 UI create and shared admin draft creation also register. Existing admin draft deletion and archived-state behaviours remain.
FACT: Ratings and accounts delete run-scoped users only after course deletion attempts settle, collecting user-cleanup failures into the same AggregateError. No helper or product code changed.
NEXT: Review the diff for accidental behaviour/assertion changes and complete the all-spec inventory before running the four authorized frontend checks.
FACT: Initial diff review confirms the four direct-response admin paths (shared helper + E2 + P20 + P21), all three transcript paths, and the other eight create helpers/tests are registered. Preserved ratings' republish token-array check and all archive-state assertions.
NEXT: Finish static coverage/unchanged-assertion checks, then run npm format → format:check → typecheck → lint in frontend.
FACT: All-spec creation inventory (read-only; outside the ten requested files left unchanged):
FACT: lesson-builder, teach-editor and refund already use immediate Sets and deleteE2ECourse/allSettled teardown (R5); subtitles has a Set and shared-helper deletion in afterAll; audio-tracks stores its created fixture before structure save and deletes with the shared helper.
FACT: playback creates primary and cross-course fixtures, stores each after create parsing, and deletes both via its existing SQL afterAll (audit/grants/assets/course). Its later archive action exercises behaviour and is outside R6.
FACT: asr tracks created drafts in its courses array before structure save and deletes them by API with fresh version reads and collected failures in afterAll; no normal-path leftover identified statically.
FACT: learn-access and media both delete their draft via API in afterAll, but assign the cleanup variable only after structure save. A failed save/status/token parse can leave the initial draft; these pre-existing failure-window gaps are reported and left outside R6's named scope. audio-tracks/asr/playback also validate token/status before storing, so malformed responses can precede registration; no live failure/leftover count is claimed.
FACT: Admin's student-denied POST expects 403 and creates no course; its later admin control creation is covered by createAdminDraft. Search found no further literal admin-course create endpoint sites beyond this inventory and the ten targets (including admin UI forms).
FACT: Existing course-cleanup.ts was not modified: no execution has demonstrated an uncovered FK; all deletion efficacy and before/after database counts remain NOT_RUN per the owner's prohibition on e2e/Docker.
NEXT: Run the authorized frontend formatting and static checks in order, recording each exit.
FACT: frontend/: npm run format exited 0 (oxfmt completed).
NEXT: Run npm run format:check.
FACT: frontend/: npm run format:check exited 0; all 420 checked files are formatted.
FACT: Inventory correction: playback's existing SQL removes grant_audit, course_grants and courses, relying on its cascade/SET NULL behaviour; it does not explicitly delete video_assets. Its retained media-row/byte behaviour is outside this slice, and no live residue was measured here.
NEXT: Run npm run typecheck.
FACT: frontend/: npm run typecheck exited 0 (tsc --noEmit).
NEXT: Run npm run lint, then verify final file scope and assertion preservation.
FACT: frontend/: npm run lint exited 0 (oxlint: 0 warnings, 0 errors; architecture check passed).
NEXT: Perform final read-only diff/assertion coverage checks and record handoff limitations.
FACT: Final static check compares TypeScript AST-printed expect chains against HEAD, excluding replaced afterAll hooks: all 527 assertions are unchanged across the ten specs. Creation registration counts are ratings/enrollment/instructor/discussion/learning/purchase-wayforpay/purchase/accounts 1 each, admin 4, transcript 3 (15 paths total). Each spec has one allSettled/deleteE2ECourse/AggregateError hook.
FACT: git diff --check -- frontend/e2e exited 0. Final tracked diff scope adds only the ten target specs to the pre-existing unrelated changes; course-cleanup.ts and product files are unchanged. Formatter changed no files outside frontend/e2e.
FACT: All four required frontend commands exited 0 in order: npm run format; npm run format:check; npm run typecheck; npm run lint. No e2e, Docker, or git commit was run. No subagents launched (0 launched / 0 returned).
DECISION: Writer implementation is ready for the owner's independent review and separately authorized live R6 proofs. H5/H6 have static implementation evidence only; H7/H8 (two-run counts and fresh-database ordered gate) remain NOT_RUN. No independent review PASS is claimed.
NEXT: Owner/reviewer examines the ten-spec diff; an authorized runtime session can then run the contract's e2e/count proofs. The additional draft failure-window findings above remain outside this writer slice.
FACT: Assertion-count correction: the final AST comparison output is 530 unchanged assertions, not 527; all ten comparisons passed. The 15 creation-path count is correct.
FACT: Final review snapshot is branch r6-e2e-archived-cleanup at HEAD fe61852aa4d446ce155f32603e4721d9fc5d9a03 plus the following SHA-256 working-tree files (all four npm checks cover these bytes):
FACT: SHA-256 docs/e2e-hygiene/plan.md: 0fcaf85e581e4732c68391d876b7f1765f83c85687ded44730324a3d1d106eaa
FACT: SHA-256 frontend/e2e/ratings.spec.ts: e0c58bc7cc12936307d6e035c085dcb8be1771e06aa8dcd6b826ef1d4379be70
FACT: SHA-256 frontend/e2e/enrollment.spec.ts: 42478b5f7a8c51b655a5f886accb6078f4a8ac9fa5ec2b4f6ae58d675777147d
FACT: SHA-256 frontend/e2e/instructor.spec.ts: 28190684085efce34a1d0b0649f9dfb03acb4a12d6c02bb5c07ac04488d6c1d1
FACT: SHA-256 frontend/e2e/discussion.spec.ts: 22bad134ac2d1ddefff5a66d2e5e61f50f4b2d2c2d57483258907366941ab600
FACT: SHA-256 frontend/e2e/learning.spec.ts: a8d313ddb4e3370e4276275e4f69f51f1a863dc4c93c79c45cc9f8afa8122c2f
FACT: SHA-256 frontend/e2e/purchase-wayforpay.spec.ts: ad6c4e47775db12601442f2b2683b67ade473f6f001dfbb19d82673af83540fe
FACT: SHA-256 frontend/e2e/admin.spec.ts: f58d803b89555c37705f77462d51852dbb6d58dd7e8736a58f6f71b1afae279a
FACT: SHA-256 frontend/e2e/transcript.spec.ts: ab9b9cb75c353b67942746c6016609faa435ee1f4311af34775113c138e48e13
FACT: SHA-256 frontend/e2e/purchase.spec.ts: 91c063b9333825da1a92703ddca1cd4e37fda296ef68a27986abbc32ca0c71e1
FACT: SHA-256 frontend/e2e/accounts.spec.ts: dc435ccea7e2ad76e676d6015e880077f2c15251fe2fa568bc38eb99ade189ac
FACT: SHA-256 frontend/e2e/course-cleanup.ts: 51d6be59d03c2dd5e4cee347d302f7d50005bab625540cd3b07b2a5441e081e2
NEXT: Hand off this writer result for independent review; runtime proof remains explicitly unrun.

## Fix F1/F2

NEXT: Read the R6 review, restore status-before-body ordering at every F1 creation site, cover F2 unexpected success, then run the four authorized frontend checks.
FACT: Verified repository root and branch r6-e2e-archived-cleanup; read shared entry rules, AGENTS.local.md, briefing, and existing writer progress. Existing entries remain intact.
FACT: Read R6 H5/H6 and both review findings; verified all thirteen F1 status-after-registration sites and the unregistered student negative-create request for F2 against current source.
DECISION: Small review fix, no product/architecture change. Restore each existing status assertion verbatim before JSON parsing, then retain id validation and immediate Set registration. For F2, only status 201 enters success-body validation/registration before the unchanged 403 assertions. No new abstraction or tests outside the authorized four scripts.
FACT: Skills read from the managed kit: skills/arch/receiving-code-review/SKILL.md and skills/pt/ponytail/SKILL.md. Independent review source is docs/e2e-hygiene/_wip-r6-review.md; this writer does not claim independent delta approval or launch reviewers.
NEXT: Capture pre-fix file bytes, apply only F1/F2, and verify the exact delta before formatting.
FACT: F1 fixed at all thirteen listed sites: the identical pre-existing 201 assertion now runs before JSON parsing; id validation and cleanupCourseIDs.add immediately follow, ahead of token/structure/publish work. Inspected the delta against pre-fix copies; only assertion movement is present for F1.
FACT: F2 adds a status === 201 branch before the existing student negative-create assertions, parsing unknown, narrowing body.id to string, and immediately registering it. Both existing expected-403/error-code assertions remain verbatim.
NEXT: Run frontend/npm run format, format:check, typecheck, and lint in the requested order; then check the final delta and file scope.
FACT: frontend/: npm run format exited 0 (420 files).
NEXT: Run npm run format:check.
FACT: frontend/: npm run format:check exited 0; all 420 files are formatted.
NEXT: Run npm run typecheck.
FACT: frontend/: npm run typecheck exited 0 (tsc --noEmit).
NEXT: Run npm run lint.
FACT: frontend/: npm run lint exited 0 (oxlint --deny-warnings and architecture checker; 246 files).
NEXT: Verify all existing assertion ASTs, all thirteen status/parse/register sequences, the exact F2 branch, and allowed-file hashes against the pre-fix snapshot.
FACT: Final read-only verification passed: 513 standalone assertion statements are AST-identical to the pre-fix snapshot; all thirteen F1 sequences have status assertion -> parse -> validate id -> register, and F2 registers only status-201 responses before the unchanged 403/error-code assertions.
FACT: Pre-fix content-hash comparison confirms this fix changed only the ten frontend/e2e/*.spec.ts files and this working file. Formatter changed no out-of-scope files. git diff --check -- frontend/e2e exited 0.
FACT: Required commands completed in order from frontend/, all exit 0: npm run format; npm run format:check; npm run typecheck; npm run lint. No e2e, Docker, or git commit ran. No subagents launched (0 launched / 0 returned).
DECISION: F1/F2 writer fixes are complete. Author static checks passed; independent delta review and live H7/H8 proofs are not claimed.
NEXT: Hand the scoped F1/F2 delta back for review; runtime verification remains outside this authorized writer task.

## Fix F3 (Courses 1, inline)
FACT: admin.spec.ts student negative-create: unexpected-201 body parse is `.catch(() => undefined)`, id registered only when valid, nothing thrown — the 403 assertion stays the primary failure. format/format:check/typecheck/lint exit 0.
