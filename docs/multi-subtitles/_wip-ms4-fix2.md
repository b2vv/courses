NEXT: Read mandatory agent-entry rules and both wave-2 reviews; inspect the three requested fixes and add regressions before implementation. Append progress after every step; no commits, staging, branch changes, or image rebuilds.
FACT: Read current shared agent-entry rules before skills or further work. The first NEXT entry was communicated before the first tool and appended to this file immediately after the mandatory entry-rule read.
NEXT: Resolve repository/local routes, read briefing and wave-2 requirements, then inspect implementation and test seams.
FACT: Verified repository root /Users/strelia/projects/courses and HEAD 3da8fb0f. Read AGENTS.local.md; Courses Docker must use the Kiri wrapper. Existing unrelated dirty/untracked files are preserved.
DECISION: Apply receiving-code-review verification and minimal-change ponytail guidance. Scope remains the three owner-listed fixes; browser checks/image rebuilding stay with the owner session.
NEXT: Read bounded review/contracts and exact implementation/test files; avoid large combined output after initial briefing output truncated.
FACT: Both wave-2 reviews read in full: failed uncertain append loses attempted language; same-language background refetch is confirmed; archive-only teardown leaks counted fixtures. M25/M26 remain pending owner work.
DECISION: Use canonical managed-kit skill bodies (skills/pt/ponytail and skills/arch/receiving-code-review) per AGENTS.local.md, which supersedes historical adapter pin paths. No delegation required.
NEXT: Trace APPENDING/settlement callers, route query/panel identity, and fixture FK cleanup before writing regressions.
FACT: M5 currently treats any in-flight query as loading; owner decision narrows that to a new language with no data. Existing audio-tracks e2e deletes grant_audit, grants, courses via portable docker compose exec.
NEXT: Read exact coverage and route tests without combined-output truncation; inspect migration foreign keys before choosing teardown SQL.
FACT: Coverage appending currently stores no language; uncertain spans inherit only replaced labels. The sole production caller is hls-media BUFFER_APPENDING, which can read hls.audioTracks[frag.level]. buffered currently assigns the confirmed language and computes settleUntil.
NEXT: Add a failing de→failed-en→de coverage/adapter regression, keeping attempted and confirmed languages distinct; inspect route test setup and schema deletion dependencies.
DECISION: Small fix route (no backend/API/architecture change), sole writer is this session. Use an explicit attemptedLanguage at APPENDING while keeping FRAG_BUFFERED language as clean-completion evidence. Possible languages of uncertain spans union that attempted language with inherited replacement history. Independent review and browser gate are owner follow-up, not claimed here.
NEXT: Resolve actual route component tests and remaining FK definitions, then record READY and add red tests.
FACT: Migration inspection: video_assets.lesson_id uses ON DELETE SET NULL (00005), so teardown must explicitly delete these assets BEFORE courses; audio_tracks cascade from video_assets (00025). transcripts/segments cascade from lessons (00007); lesson_progress cascades from lessons (00006). course_grants and grant_audit have non-cascading FKs (00011). No orders are created by subtitles.spec.ts.
NEXT: Reuse the existing route fixture and exercise real Query refetch state; add a runnable teardown regression without launching browser checks.
FACT: Actual route is frontend/src/routes/learn.$.tsx:226 (isPending || isFetching). Existing route-parser test is not the component harness; locating the MS4 route component test before editing. Branch is ms4-player-captions as requested.
NEXT: Write tests first, run focused regressions to capture real failures, then make the minimal fixes.
FACT: Found existing MS4 route harness in frontend/src/routes/route-guards.test.tsx:1930, using real Query options and repository calls. The task is below the spec-flow threshold; exact acceptance and verification are owner-provided, no missing contract.
DECISION: Pre-code READY for scoped implementation: (1) regression and possible-language union, (2) real in-flight same-language Query regression plus isPending-only/M5 decision, (3) run-scoped SQL teardown with zero-residual assertion. No new dependency, no generated code, no backend change. Browser run and full ordered gate remain NOT_RUN here by owner instruction.
NEXT: Add coverage and HLS wiring regressions first; run them red.
FACT: Added discriminating coverage and HLS event-wiring tests for clean de [0,6) → en APPENDING/APPENDED/ERROR at 2.5 → clean de at 2.6, requiring pending through 4.09 and confirmation at 4.1. Focused red run launched before implementation.
NEXT: Capture failures, then record attempted language and union uncertain possibilities.
FACT: Focused pre-fix run exited 1: 66 passed / 2 failed. Both new tests received confirmed de at 2.6 instead of pending, proving the reported defect.
FACT: Implemented attemptedLanguage captured at APPENDING via captionLanguageOf(hls.audioTracks[frag.level].lang). Uncertain spans contribute replaced history plus attemptedLanguage; completed-clean spans contribute only their clean label. Existing settle logic now sees failed en when returning to de.
NEXT: Confirm focused green result, then add the same-language background-refetch route regression before changing isPending.
FACT: Post-fix coverage + HLS focused run exited 0: 68/68 passed, no skips. This proves state/event wiring, not real decoder behavior; the measured 1.5-second guard remains the existing documented policy.
NEXT: Add route regression with a held same-language de refetch and verify both track identity and panel search/follow state survive.
FACT: Added route regression using real QueryClient.refetchQueries with existing de data and a held second repository read. Assertions cover visible transcript, same caption track node, retained search node/text during and after refetch, and follow-paused state after clearing search. Pre-fix focused run launched.
NEXT: Capture the red route result; change only isPending and record the dated M5 owner decision.
NEXT: Check route regression result and Kiri/Postgres availability. For cleanup, add an executable database regression that invokes the same teardown SQL on a run-scoped fixture without starting any browser or rebuilding images; check residual counts explicitly.
FACT: First route red run exited 1 but hit an incidental same-language refetch when Play re-enabled the query, before the intended explicit refetch. Corrected harness to let startup reads settle, then hold only the explicit refetch; rerunning red before product change.
FACT: Kiri Compose services are healthy. The wrapper sets DOCKER_HOST, so bare context show printed default; explicit docker --context kiri context show is used to verify the named context as required. No global context changed.
NEXT: Capture discriminating route failure and current historical fixture count; implement M5 decision and database cleanup regression.
FACT: Baseline DB query succeeded on Kiri: 12 courses have slug LIKE e2e-subtitles%. These are retained fixtures, not a zero-cleanup result.
FACT: Route harness still observed a transient startup loading panel; changed confirmation to async act and wait for Query fetchStatus idle before capturing panel/search nodes. Product remains unchanged for the red run.
NEXT: Finish route red→green, then test deletion against actual PostgreSQL rows (including media, transcript and progress).
DECISION: Cleanup regression will use a small e2e-only teardown helper and standalone Node/assert script against real PostgreSQL; this runs no Playwright/browser checks. It will seed run-scoped rows, exercise the teardown used by subtitles.spec.ts, and assert zero remaining course/grant/audit/media/audio/transcript/progress rows, plus partial-setup cleanup.
NEXT: Implement helper with the current archive-only behavior, prove regression red, then replace it with ordered transactional SQL deletion.
FACT: The route test's early callback could fire before Play asynchronously created the media adapter, leaving selection pending. Added an explicit wait for media callbacks; this corrects the test setup rather than weakening its refetch assertions. Output captured at /tmp/ms4-fix2-route-red.log for bounded failure inspection.
NEXT: Inspect the true refetch failure and proceed with product fix once setup is proven.
FACT: Added e2e teardown helper initially preserving archive-only behavior and a standalone PostgreSQL regression for partial/full setup. The test always deletes its own UUID-scoped rows in finally, including on red runs. It checks 10 residual row counts, including non-cascading media and dependent transcript/progress/audio rows.
NEXT: Capture database regression red, replace archive-only teardown with FK-ordered deletion, register IDs immediately, and rerun green.
FACT: Discriminating route red exited 1 at the intended held background refetch: Caption de disappeared (expected truthy, got null). Startup and search/follow setup completed. Applied isPending-only; added one dated Courses 1 (2026-11-18) decision line to plan-ms4.md M5, explicitly preserving same-language data/state and rejecting keepPreviousData.
NEXT: Check route green, capture full database fixture red (complete setup first), then implement ordered cleanup.
FACT: Archive-only database regression for partial setup exited 1: residual counts 1|0|0|0|0|0|0|0|0|0, proving course retention; finally removed its own fixture. Full-setup archive-only regression is now running before deletion is implemented.
NEXT: Wire cleanup to subtitles.spec.ts and delete audit→grants→assets→course transactionally; keep schema-driven cascade cleanup for remaining tables.
FACT: Route regression green exited 0 (1 passed / 69 intentionally filtered). Database regression red exited 1 for full setup with all ten counts = 1; its finally cleaned all test-owned rows.
FACT: Implemented transactional SQL cleanup audit→grants→video_assets→courses, with FK comments. subtitles.spec.ts now registers each course ID before updatedAt validation/further setup and calls this cleanup in afterAll; request disposal runs in finally. Database regression also exercises repeated/idempotent teardown.
NEXT: Capture database green, check baseline fixture inventory, then run the exact owner-authorized frontend verification sequence.
FACT: Database regression ran against Kiri using inherited Compose environment and actual PostgreSQL, no browser/images. Checking its completion and retained fixture inventory before cleanup verification.
NEXT: Review final scoped diff and run format → format:check → typecheck → lint → full src Rstest; record each real exit.
FACT: Database cleanup regression exited 0: complete and partial setup each left 0|0|0|0|0|0|0|0|0|0 rows; repeated deletion succeeded. All 12 existing e2e-subtitles courses were confirmed archived.
DECISION: Remove only these retained subtitles fixtures with the same new cleanup helper, checking no non-archived subtitles course or order is present and verifying captured media IDs disappear. This satisfies requested zero-residual verification without browser runs.
NEXT: Capture retained-cleanup result and npm run format result, then proceed in the exact verification order.
FACT: npm run format exited 0 (oxfmt, 416 files).
NEXT: Finish npm run format:check, then npm run typecheck.
FACT: npm run format:check exited 0 (all matched files formatted).
FACT: Retained-fixture cleanup exited 0: deleted 12 archived e2e-subtitles courses and their 36 linked video assets; all captured video IDs absent afterward. Final SELECT count(*) FROM courses WHERE slug LIKE 'e2e-subtitles%' returned 0. No non-archived subtitles fixtures/orders were present; no browser run or image rebuild occurred.
NEXT: Capture npm run typecheck, then run npm run lint and full src Rstest.
NEXT: Inspect formatted final diff/helper/regression and phase report while typecheck completes; preserve existing unrelated changes. Required verification remains sequential, lint only after typecheck result.
FACT: npm run typecheck exited 0 (tsc --noEmit). Formatted diff inspected: only requested frontend behavior, tests/cleanup helper, and M5 decision changed in task scope; no backend source or generated output edited.
NEXT: Capture npm run lint and run npx rstest run src --testTimeout 30000.
FACT: npm run lint exited 0 (oxlint --deny-warnings; architecture boundaries checked 244 files).
NEXT: Full owner-requested src Rstest running; output at /tmp/ms4-fix2-rstest-src.log. After completion, record exact counts/exit, final DB count, phase report and unchanged HEAD/no staging status.
FACT: format, format:check, typecheck and lint all exited 0 in requested order. Full src Rstest is the remaining authorized frontend gate. No image rebuild, Playwright run, staging or commit has occurred.
NEXT: Read bounded full-suite summary; inspect fixture writes and final whitespace/status before closeout.
FACT: npx rstest run src --testTimeout 30000 exited 0: 1435/1435 tests passed across 114 files, 0 failed/skipped/todo (11.751 s). Full log: /tmp/ms4-fix2-rstest-src.log.
FACT: git diff --check exited 0; index diff is empty; HEAD remains 3da8fb0f. Reviewed fixture writes: one course, three lesson video uploads, uploaded audio, exact transcripts, grant and progress; no orders or additional non-cascading course dependencies are created.
NEXT: Append phase report with real command results, limits and owner handoff; final DB count query must still be zero.
FACT: Final Kiri query after full src tests again returned 0 for SELECT count(*) FROM courses WHERE slug LIKE 'e2e-subtitles%'. Appended scoped remediation/results/limitations to docs/multi-subtitles/report.md.
DECISION: Task implementation and authorized frontend/standalone-DB verification are complete. Browser/e2e behavior, post-browser fixture count, frontend image rebuild, M25/M26 and full ordered gate remain owner-session work; no independent review PASS is inferred from these checks.
NEXT: Final read-only diff/status check, then hand off concise summary. No commits, staging, branch switch, checkout/reset/stash/clean/rm or image rebuilds.
FACT: Final diff whitespace check exited 0; staged file list empty; HEAD 3da8fb0f. Task files are unstaged modifications plus new cleanup helper/regression and this append-only working file. No agents launched (0 launched / 0 returned).
NEXT: Owner session: rebuild frontend and run browser checks; query e2e-subtitles* residual count after the actual browser run. This session stops at the requested handoff boundary.
