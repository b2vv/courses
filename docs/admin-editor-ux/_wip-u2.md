NEXT: Read the mandatory entry rules, verify repository and branch state, and inspect the U2 contract before implementation.
FACT: Read the shared entry rules; the owner-specified U2 working file is the active task record. No delegation planned.
FACT: Verified repository root <repo>, branch admin-editor-ux; existing unrelated dirty/untracked files will be preserved. Initial bulk documentation output was truncated; switching to bounded required sections.
FACT: HEAD is exactly 973a7789. Read accepted r3 D4/D5/D8, Ux04–06/Ux13, U2 migration and R1/R2. Scope is frontend-only, below the full pipeline threshold; pre-code READY on the accepted plan. Owner session owns independent review, browser and full e2e; this writer will not launch agents, Docker, e2e, or git mutations.
NEXT: Inspect existing lesson state, heading structure, translations, and exact R1 selectors; use the narrowest existing-state callbacks.
FACT: R1 inventory confirms four multi-lesson drivers need idempotent expansion, plus media reload. Header title/preview controls must remain accessible without expansion.
FACT: LessonFields already keeps generated lesson ids and formKey React keys; transcriptKnowledge is maintained at CourseEditorView. Video state is derived inside LessonVideoField from rows plus local upload progress; no header query is needed.
DECISION: Use a shared presentation LessonSectionHeader (h4/status) so media never imports admin. Keep course materials h2 per owner clarification. Store expansion by lesson id in ModuleFields, initialized once from the saved course; explicitly expand newly added ids. Add one video-status callback derived from the existing displayed asset and progress.
NEXT: Add Ux04/05/06/13 regression tests before product edits and run them to record the expected red baseline.
FACT: Added nine regression cases covering defaults/new rows, aria/hidden/tab persistence, actual pending upload identity/progress, unknown→yes/no→failed-refresh knowledge, localized saved-only hrefs, and h2/h3/h4 shared section headings.
NEXT: Inspect baseline regression results, then implement the accepted collapse and shared header behavior.
FACT: All new U2 regression cases fail against the unchanged implementation (red baseline logged at /tmp/courses-u2-red.log); heading test fixture label needs alignment with the existing localized program name before final verification.
DECISION: Keep existing fieldset/group names for all browser selectors; add the lesson h3 within its legend. Shared section header is reused by video and the language wrapper; course-level materials remain untouched.
FACT: Implemented id-keyed expansion, hidden mounted bodies, unchanged header fields, localized saved-only content links, three-state transcript badges and video status callback; added shared h4/status header and wrapping token CSS. ASR lesson subsection heading changed h3→h4 to keep the lesson outline consistent.
NEXT: Run U2 tests, correct test fixtures or implementation failures, then migrate the four e2e drivers.
FACT: New tests had a red baseline of 9 failed / 0 passed; implementation is now under targeted verification. CSS uses wrapping flex/grid, min-inline-size:0, full-width caps and existing button focus/44px rules; browser/zoom evidence remains owner-owned.
FACT: Targeted U2 regression run passed 9/9, exit 0. Media has a direct audio-management entry plus a reload that also need expansion. Audio helpers must call expansion repeatedly on the same row to preserve idempotence coverage.
NEXT: Implement shared openLesson helper returning the scoped row; add a cleanup-safe two-lesson admin journey asserting initial collapse and content-editor navigation.
FACT: Added one shared idempotent openLesson helper; migrated media upload/direct audio/reload, both audio lessons and repeated addTrack calls, every subtitles loop target, and playback uploadReady including single-lesson/replacement callers. Added cleanup-safe admin e2e initial-collapse + real content-editor link navigation assertions. E2E was not run.
FACT: npm run format:check exit 0.
FACT: npm run typecheck exit 2 (details /tmp/courses-u2-typecheck.log).
FACT: npm run format exit 0; npm run format:check exit 0. The initial verification-note append used a root-relative path from frontend and failed; subsequent command-result appends used the correct ../docs path, and this entry records the missed note.
NEXT: Resolve any typecheck findings, run lint and all src tests, preserving existing assertions when adapting unit interactions to collapsed bodies.
FACT: First typecheck failed only on a new Testing Library getByRole exact option (unsupported there); removed it (string names are exact by default). Adapted existing draft-routing and reorder/delete unit journeys to expand their real target rows; retained all original assertions. No scoped frontend AGENTS files found.
FACT: npm run format rerun exit 0.
FACT: npm run format:check rerun exit 0.
FACT: Static migration review preserves upload/ready/API assertions and adds visible controls in audio/subtitles. openLesson scopes exact module/lesson groups, clicks only aria-expanded=false, then asserts true and controlled body visibility. No force/includeHidden/API replacements added.
FACT: npm run typecheck rerun exit 0.
FACT: npm run lint exit 0 (details /tmp/courses-u2-lint.log).
NEXT: Inspect full-suite failures and final diff for state/heading regressions; adjust only confirmed issues, then repeat the required ordered checks on the final bytes.
FACT: Ordered format, format:check, typecheck and lint all exit 0 after the test-type correction; lint also checked 245 architecture-boundary files. Full src unit suite is running.
FACT: npx rstest run src --testTimeout 30000 exit 0 (details /tmp/courses-u2-units.log).
FACT: Found a real fallback knowledge risk: LessonTranscriptField reports false whenever loading ends with undefined data, including a failed refresh without language inventory. The inventory path already guards undefined. Investigating the hook contract before narrowing the fallback callback.
DECISION: Unknown means no successful transcript data, including the no-inventory fallback. The hook returns an empty LessonTranscript with language:null for known absence; undefined must never overwrite a known yes/no. Add a fallback regression before fixing that callback guard.
FACT: First complete src suite passed 1454 tests, 0 failed/0 skipped, exit 0. Found callback hardening still needed for failed no-inventory refresh; final ordered verification will follow that small correction.
FACT: Added two no-inventory fallback tests for initial unknown, successful yes/no, then undefined refresh data; verifying red before changing the callback.
FACT: Fallback test insertion script did not match formatter-wrapped source and made no write; attempted name-filter run selected 0 tests (35 skipped), so it is not a red/green result. Correcting the insertion anchor.
FACT: Fallback regression cases are now present in the test file; running only those two against the unchanged callback.
FACT: Both fallback knowledge regressions failed before the guard change (2 failed, exit 1). LessonTranscriptField now reports only successful defined data: language:null means known no; undefined leaves prior knowledge intact. No hook/API signature changes.
NEXT: Run final ordered requested checks on this snapshot. Browser/390px/200% text, full e2e and independent review remain owner-session work.
FACT: Final npm run format exit 0.
FACT: Final npm run format:check exit 0.
FACT: Final npm run typecheck exit 0.
FACT: Final npm run lint exit 0.
FACT: Final scoped diff whitespace check passed. Unrelated initial changes are preserved. E2E navigation fixture uses a draft course and deletes it with the updated concurrency token; existing scenario assertions remain in place.
FACT: Final npx rstest run src --testTimeout 30000 exit 0; full results /tmp/courses-u2-units-final.log.
FACT: Final format, format:check, typecheck and lint each exit 0. Fallback fix is inside the existing transcript-knowledge contract, and no dependency/query was added. Independent review NOT_RUN here; owner-selected review route is unchanged.
FACT: Final ordered verification passed: npm run format EXIT 0; npm run format:check EXIT 0; npm run typecheck EXIT 0; npm run lint EXIT 0 (245 architecture files); npx rstest run src --testTimeout 30000 EXIT 0 (1456 passed, 0 failed, 0 skipped). Original new U2 baseline: 9 failed; added fallback baseline: 2 failed. git diff --check EXIT 0.
FACT: U2 implementation + unit tests + planned e2e source migration are ready for owner review. No commits, staging, branch changes, Docker or e2e executions; no new dependencies. Current branch/base remain admin-editor-ux @ 973a7789 with uncommitted U2 changes.
NEXT: Owner session runs full e2e and browser checks (390px and 200% text included), then independent review of the final U2 snapshot. Ux04/06 browser evidence and Ux13 navigation e2e remain NOT_RUN, not PASS.
NEXT: Read independent review T1 and dated D5a, verify current branch/WIP, and add failing Label-in-Name tests before fixing toggle labels. Owner full e2e is running; no Docker/e2e commands.
FACT: Reverified root, branch admin-editor-ux and HEAD 973a7789; read receiving-code-review skill from canonical kit. T1 is the single authorized code fix; D5a must remain a documented scope decision.
FACT: Read review T1 and D5a (2026-11-18, Courses 1): materials retain course-level h2; their status line belongs to U5. T1 confirmed in source; structural openLesson requires no change, while admin.spec has one old Ukrainian name selector. Added two localized regression cases checking the actual visible action and exact accessible name in both states before product edits.
FACT: T1 pre-fix targeted test run exit 1; raw output /tmp/courses-u2-t1-red.log.
NEXT: After confirming both T1 regression failures, replace the fixed toggle-name key with state-specific localized names and migrate every old name-based selector.
FACT: T1 regression baseline: 2 failed / 0 passed, exit 1. Toggle now chooses expandLessonLabel/collapseLessonLabel by the same expanded state as visible text, with action-first English and Ukrainian context. Updated old unit queries and admin.spec collapsed-toggle selector; structural openLesson is unchanged. No materials code changed (D5a).
NEXT: Run requested ordered format, format:check, typecheck, lint and full src rstest; no runtime/e2e interaction.
FACT: T1 npm run format exit 0.
FACT: T1 npm run format:check exit 0.
FACT: Source inventory has no remaining old toggleLesson/Toggle lesson/old Ukrainian names. git diff --check passed. New unit cases assert both visible text and exact accessible action-plus-context names in both locales/states, using the same DOM toggle through clicks.
FACT: T1 npm run typecheck exit 0.
FACT: T1 npm run lint exit 0.
FACT: T1 ordered format, format:check, typecheck and lint all exit 0; lint includes 245 architecture files. Full src unit result is being checked separately; no Docker/e2e execution, staging, commits or branch changes.
FACT: T1 npx rstest run src --testTimeout 30000 exit 0; raw results /tmp/courses-u2-t1-units.log.
FACT: Final whitespace check still passes. D5a unchanged; no material/header implementation changes in this review-fix slice.
FACT: T1 final full src suite passed: 115 files, 1458 tests passed, 0 failed, 0 skipped; npx rstest run src --testTimeout 30000 EXIT 0. Both localized action/name regression cases now pass after their pre-fix failures. All five requested commands EXIT 0; raw outputs /tmp/courses-u2-t1-{format,format-check,typecheck,lint,units}.log.
DECISION: T1 author fix is ready for independent delta review. S1 is closed by owner decision D5a, with no materials code changes. No independent re-review or owner in-progress full-e2e outcome is claimed here.
NEXT: Owner consumes this correction in the running verification/review workflow; attach browser/full-e2e and reviewer evidence to the corrected snapshot. This session did not execute Docker/e2e, stage, commit or switch branches.
FACT: T1 final snapshot SHA256 3ae5fea8726e630b539362d86633c218f73c6ad66f85dc6b580dca0316b0c7be  frontend/src/modules/admin/ui/CourseEditorView.tsx
FACT: T1 final snapshot SHA256 e544319221e96149e0f159fe03684205a3202bc14050d9e92c3dea0589c18636  frontend/src/modules/admin/ui/CourseEditorView.test.tsx
FACT: T1 final snapshot SHA256 440e9058bb023e982fb3139dc063319ba1e31e50a32751a376330969fccba67d  frontend/src/app/providers/locales/en.ts
FACT: T1 final snapshot SHA256 aff3e50c02c23ad8b843abf67b6d0426c3927275eb08f424e2469f2993cbc27e  frontend/src/app/providers/locales/uk.ts
FACT: T1 final snapshot SHA256 78b91bd4068e6c2990aa44c7cae707d4bb3046154db31c393bf5a75e94f9b3d3  frontend/e2e/admin.spec.ts
