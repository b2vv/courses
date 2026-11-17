NEXT: Read repository instructions, accepted MS-D8/MS-D6 plan, MS3b checkpoint, and current frontend implementation before editing.
FACT: Initial repository/briefing/workflow and ponytail skill reads completed; combined output was truncated, so authoritative task sections will be read separately.
NEXT: Inspect scoped plan/checkpoint, workflow requirements, git status, and affected frontend files.
FACT: Shared WIP exists in report.md and unrelated docs/config; no affected frontend file is dirty. MS3b checkpoint explicitly splits run 1 (M3-R4 + baseLanguage) from run 2.
NEXT: Read exact authoritative sections and affected code; retain all unrelated WIP.
FACT: MS-D8 requires fresh option filtering and PATCH→PUT in admin UI; MS-D6 requires restoring the base-language/token pair. Existing editor already follows fixed language props while dirty.
NEXT: Inspect tabs orchestration, recovery adapter validation, translations, and test harness in smaller reads.
FACT: Tabs already receive video.audio.actions; no CourseEditorView/route wiring change needed. IndexedDB currently returns typed objects without runtime read validation.
DECISION: 2026-11-16 Codex: use empty string as unknown restored base-language sentinel (invalid real language tag), preserving the single language-inequality token branch. Verify with M3-R13 g and recapture.
DECISION: 2026-11-16 Codex: pause pending recovery capture during Save/dialog and keep Save focusable with aria-disabled + an invocation guard; verify failure store spies and dialog focus tests.
NEXT: Implement confirm orchestration, optional snapshot baseLanguage validation/capture/restore, translations, then unit coverage.
FACT: Root /Users/strelia/projects/courses, branch ms2-3-asr-per-track, HEAD 9ad04e3a5bb7066729f0ebbfd5cabe1847d92c51 confirmed; run-1 resume READY under accepted checkpoint. Verification is author frontend checks; independent review/browser/full e2e remain Courses 1/MS3c.
NEXT: Apply scoped implementation.
FACT: Optional baseLanguage is now captured/restored; legacy snapshots restore empty-string owner. Recovery can suspend timers/captures and Save retains focus.
NEXT: Implement und dialog and validated IndexedDB reads.
FACT: und orchestration and translations implemented. The first adapter-edit command used frontend cwd with root-relative paths and failed before editing; typecheck started separately.
NEXT: Correct the adapter edit from repository root and inspect typecheck.
FACT: Initial typecheck EXIT=0. IndexedDB read now validates optional baseLanguage and preserves legacy snapshots without a version/schema bump.
NEXT: Rewrite interim und test and cover M3-R4 branches plus M3-R13 capture/restore/legacy behavior.
FACT: Added row-labelled M3-R4 dialog/failure/retry/focus tests and M3-R13 e/f/g restore/recapture tests; old R13b fixture now explicitly names its token owner.
NEXT: Inspect targeted results and add IndexedDB optional-field validation tests.
FACT: Targeted run EXIT=1: 55/59 passed. Four new test harness failures: RAC Select accessible name includes selected/placeholder text, and Rstest uses rs rather than rst. No behavior failure observed yet.
NEXT: Fix harness names/API, test adapter reads via existing happy-dom IndexedDB support or a narrow mock, then rerun.
FACT: Select test names corrected; checking the installed test API showed the exported helper is rstest (not rs). All M3-R13 e/f/g integration tests passed in the targeted run.
NEXT: Correct helper import and add adapter coverage without introducing a dependency.
FACT: Post-adapter typecheck EXIT=2: exactOptionalPropertyTypes rejects Zod optional fields spread as explicit undefined. Fix requires explicit optional-field mapping at storage boundary.
NEXT: Map validated optional fields explicitly and inspect targeted run.
FACT: Targeted suite EXIT=0, 64/64 tests across 3 files; dialog focus entry/Cancel return/Escape return passed. Corrected exact-optional mapping at the adapter boundary.
NEXT: Run requested format→typecheck→lint→full unit gate, then inspect final scoped diff and record evidence.
FACT: Requested npm run format EXIT=0, 398 files processed; formatter finished in 189 ms.
NEXT: Requested npm run typecheck.
FACT: Requested typecheck EXIT=2: the new adapter test afterEach returned rstest utility instead of void; production code typechecked.
FACT: Diff inspection found a production-path gap: LessonTranscriptField returns null while the new-language transcript query loads, which would unmount dirty lines after PATCH/refetch. Existing fixture always reported isLoading=false.
DECISION: 2026-11-16 Codex: preserve an already-mounted editor through later transcript-language loads; retain initial loading behavior. Verify M3-R4 with an explicit new-language loading interval.
NEXT: Fix test cleanup typing and the loading unmount, strengthen retry test, then restart requested ordered frontend gate.
FACT: Mounted editor now survives new-language loading; M3-R4 retry fixture exercises loading=true then false and checks the exact textarea DOM node.
FACT: Restarted npm run format EXIT=0, 398 files, 134 ms.
NEXT: Restarted npm run typecheck.
FACT: Restarted npm run typecheck EXIT=0.
NEXT: Requested npm run lint.
FACT: Requested npm run lint EXIT=0; architecture boundaries checked for 239 files, no warnings.
NEXT: Requested full npm test.
FACT: Full frontend unit suite is running after format/typecheck/lint. No generated, e2e, or backend file appears in the frontend diff.
NEXT: Inspect final unit results; record changed files and handoff limits in the append-only run log and phase report.
FACT: Final requested ordered frontend gate passed: npm run format EXIT=0 (398 files); npm run typecheck EXIT=0; npm run lint EXIT=0 (239 architecture files); npm test EXIT=0 (105 files, 1064 passed, 0 failed/skipped/todo; 13.880 s). git diff --check -- frontend EXIT=0.
FACT: Run 1 implements M3-R4 at unit level: fresh filtering/preselection/required selection, named heading and hint, no-options Cancel, focus entry/Cancel/Escape return, PATCH→tokenless PUT, failure errors/no PUT/recovery untouched, fresh ids, same mounted editor through renamed-language query loading, ordinary fresh-read retry after PUT failure. Interim und-refusal test rewritten.
FACT: M3-R13 unit evidence covers capture baseLanguage, restore/recapture (en,T1) while saving es absent/present, exact target token after confirmation, legacy unknown token owner for absent/present target, optional-field IndexedDB validation and version 1 compatibility. Adapter request events are mocked: these tests prove parsing, not real browser persistence.
FACT: Changed frontend files (11):
- frontend/src/modules/admin/ui/LessonLanguageTabs.tsx
- frontend/src/modules/admin/ui/LessonLanguageTabs.test.tsx
- frontend/src/modules/transcript/application/recovery.ts
- frontend/src/modules/transcript/ui/use-transcript-recovery.ts
- frontend/src/modules/transcript/ui/TranscriptEditor.tsx
- frontend/src/modules/transcript/ui/TranscriptEditor.test.tsx
- frontend/src/modules/transcript/ui/LessonTranscriptField.tsx
- frontend/src/modules/transcript/adapters/transcript-history-indexed-database.ts
- frontend/src/modules/transcript/adapters/transcript-history-indexed-database.test.ts (new)
- frontend/src/app/providers/locales/en.ts
- frontend/src/app/providers/locales/uk.ts
FACT: Documentation: this append-only log plus an appended run-1 evidence section in docs/multi-subtitles/report.md. Existing shared WIP preserved. No commit, branch change, generated/e2e/backend edits, or new dependency.
NEXT: Run 1 implementation and requested tests complete; no missing run-1 test identified. Courses 1: independent review/browser verification. Run 2: detached snapshots + flush on unmount (M3-R6/M3-R12); MS3c: browser/e2e rows and final full ordered make check. These checks were NOT_RUN here; no phase-wide closure or production claim.
FACT: Final documentation/frontend diff whitespace check EXIT=0; final status confirms exactly the listed 11 frontend files plus the run log and appended phase report. Unrelated pre-existing WIP remains outside this delivery.
NEXT: Handoff ready for Courses 1; no further run-1 edits pending.
