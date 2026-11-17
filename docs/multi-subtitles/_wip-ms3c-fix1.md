NEXT: Read the repository briefing, task instructions and w1 review; then inspect the existing MS3c e2e delta without changing product code.
FACT: Created this append-only journal before repository inspection and read the CTO briefing. Current task is explicitly e2e-only, with no commits or product edits.
NEXT: Read local workflow, phase plan, review and prior run evidence; identify runtime and test assertions.
FACT: Branch is ms2-3-asr-per-track; baseline includes the existing asr.spec.ts and media-dialog product fix plus unrelated WIP. Reviewed w1 and prior evidence pointers; no other writer work will be altered.
DECISION: Resume/small fix route. Scope = asr.spec.ts plus this journal; user owns runtime authorization (courses-kiri targeted/full-file runs). Independent w2 delivery remains Courses 1, using the existing review route; no new reviewer spawned.
NEXT: Inspect exact helpers/tests and current runtime, then apply four review fixes.
FACT: w1 findings confirmed in source locations: self-derived PUT expectations, non-discriminating replacement setup, rename/automatic-ASR race, and teardown loop without exception isolation.
NEXT: Read bounded source ranges and exact runtime wrapper; independent expected texts will come from per-track draft GETs or typed constants.
FACT: Exact source read complete for target helpers/cases. Runtime wrapper path is /Users/strelia/project-data/courses-kiri/courses-kiri. Prior full file = 26 passed including setup/teardown, but w1 invalidates the listed proof rows.
NEXT: Check pinned workflow inputs and record READY, then implement independent text assertions and discriminating replacement variants.
FACT: Applied explicit-file receiving-code-review skill from the pinned candidate; all findings verified before edits. Bounded phase acceptance and standards inspected; this is a small resume under existing MS3c acceptance, not a new feature.
DECISION: READY (2026-11-17, Codex). Existing e2e helper reuse; no new dependencies. R12 snapshot will be a dirty language rename over a saved original transcript, so saving into a replacement language must preserve the original. Mutation runs will deliberately corrupt expected values only and be reverted before final verification.
NEXT: Implement Major 1 first and record its mutation red/green evidence, then complete placement/race/cleanup fixes.
FACT: Relevant pinned workflow files and manifest match package metadata. HEAD/full input SHA256 confirmed: 7b20d2c9b06f8457a274d18f0a19388f5bf906c0 / asr 92e0695aed62dfc2f81993ba98110d3e4f39115af860dea74938ba50cd1de812; media product TSX/test unchanged at w1 hashes.
DECISION: R12b uses saved es → dirty rename de; replacement has French default plus German non-default (matching), or English non-default (no de match). Both Save to the destination and preserve es. Asset IDs will come from media inventory, track default status from the audio inventory.
NEXT: Apply Major 1 independent assertions, then run expectation mutations for R8 and each R11b/c/d subcase.
FACT: Major 1 implemented: R8 PUT/GET expected texts come from each real draft; R11 check receives independently assembled expected arrays from typed/restored constants plus unchanged draft tail. Prior-language checks use those arrays.
NEXT: Run format/typecheck/lint and mutation R8 with swapped draft expectation.
FACT: Minor cleanup now catches request/status/parse exceptions per course, records ID+error and asserts the accumulated failures after attempting every course. Major 3 now finishes initial recognition, renames en, explicitly regenerates and only then checks trackLanguage=en before switching.
NEXT: Replace R12b with saved-original/dirty-rename cases and media identity evidence.
FACT: First local checks (after Major 1): npm run format exit 0; npm run typecheck exit 0; npm run lint exit 0 (240 architecture files). Later edits still need final checks.
NEXT: Implement discriminating R12b assertions, then start requested mutation runs.
FACT: Major 2 implemented: es original saved and reread, dirty de snapshot, replacement fr default + de/en non-default; asserts offer present/absent in distinct tabs, Restore → Save 200 → GET intended lines and exact es preservation. Old/new Ready asset IDs are mandatory media reads and attached with track/default inventory.
NEXT: Format/typecheck/lint, then mutate R8 expectation for a deliberate assertion failure.
FACT: All-fix local checks: format 0, typecheck 0, lint 0; diff whitespace check clean and no new frontend product changes.
DECISION: Mutation R8-red temporarily compares the und→en PUT against the other track draft (dubTexts). This must fail at the new independent assertion, then will be reverted.
NEXT: Run only M3-R8 with ms3c-fix1-mutation-r8-red output.
FACT: R8 mutation runtime started with the required courses-kiri prefix and isolated output; log /tmp/ms3c-fix1-mutation-r8-red.log. Source snapshot intentionally contains only this expected-value mutation.
NEXT: Inspect assertion failure and cleanup, revert the single expectation, then mutate R11b/c/d one at a time.
FACT: R8 mutation setup authenticated all roles. Static check confirms isDefault is the real audio-inventory field; no synthetic media/draft state is introduced.
NEXT: Await real recognition and the deliberately wrong-content assertion.
FACT: Runtime still active; no ASR result yet. Mutation snapshot and shared runtime remain unchanged.
NEXT: Collect the R8 red reporter line before reverting.
FACT: Setup reporter passed media fixture generation (29.1s). R8 is now using the actual speech video and dub bytes through normal uploads and recognition.
NEXT: Wait for R8 wrong-track expectation failure.
FACT: R8 real recognition is still active; no reporter outcome inferred from timing.
NEXT: Collect the intentional assertion failure, revert and proceed to R11 mutation b.
FACT: Continued waiting for the active R8 run; no source/runtime mutation during execution.
NEXT: Read exact outcome and preserve reporter evidence.
FACT: Waited for real R8 recognition; inspected the reporter log before any revert.
NEXT: Revert only after the red assertion is confirmed.
FACT: R8-red exit 1, intentional mismatch confirmed at englishPut texts vs dubTexts; actual Welcome text differs from expected dubbed text. Reporter:
```text
  ✘  5 [chromium] › e2e/asr.spec.ts:672:3 › ASR acceptance with a stored admin session › M3-R8 R2 R4 success and R6 two-track drafts, saves and isolated reload recovery (59.9s)
```
NEXT: Revert R8 expectation, run R8-green together with R11b-red.
FACT: Reverted the R8 swap. R11b-red now deliberately expects the original draft rather than the typed Polish variant; only the check argument changed.
NEXT: Run M3-R8|M3-R11; R8 should pass and R11b should fail at independent expected text equality.
FACT: R8-green/R11b-red runtime active; original R8 expectation restored and the sole intentional mutation is the R11b expected-text argument.
NEXT: Collect both reporter outcomes and revert R11b after its intended mismatch.
FACT: Setup passed; real R8 recognition/saving still in progress. No source edits during this run.
NEXT: Await R8 positive and R11b mutation assertion results.
FACT: Observed ongoing targeted runtime; no outcome claimed before reporter completion.
NEXT: Capture R8 and R11b reporter lines.
FACT: R8-green reporter PASS (1.0m), proving real per-draft PUT and GET equality plus existing recovery sequence. Awaiting R11b deliberate mismatch.
NEXT: Record red failure details and revert b before c mutation.
FACT: Combined run exit 1 as intended: R8-green passes; R11b-red fails at check(expectedTexts), expected pre-edit Welcome vs actual typed Polish. Reporter:
```text
  ✓  5 [chromium] › e2e/asr.spec.ts:672:3 › ASR acceptance with a stored admin session › M3-R8 R2 R4 success and R6 two-track drafts, saves and isolated reload recovery (1.0m)
  ✘  6 [chromium] › e2e/asr.spec.ts:972:3 › ASR acceptance with a stored admin session › M3-R11 a-d one track saves fresh ids through fr pl en and restored es without changing prior languages (27.5s)
```
NEXT: Revert R11b and mutate R11c to expect the Polish pre-edit text.
FACT: R11b reverted; R11c-red now uses polishTexts instead of replacementTexts in the replacement-save check. Product and entered text are unchanged.
NEXT: Run M3-R11 only and confirm c fails after b passes.
FACT: R11c-red active under the required wrapper; source contains only the intentional c expectation mutation.
NEXT: Verify the exact independent assertion failure, then restore c and mutate d.
FACT: Setup/ASR runtime observed for R11c mutation; awaiting completed reporter, no pass inferred.
NEXT: Collect R11c-red assertion and revert.
FACT: Inspected R11c mutation runtime outcome/log without changing the running test snapshot.
NEXT: Record mismatch and revert c before d.
FACT: R11c-red exit 1 as intended at check(expectedTexts): expected Polish pre-edit text; received typed English replacement. Reporter:
```text
  ✘  5 [chromium] › e2e/asr.spec.ts:972:3 › ASR acceptance with a stored admin session › M3-R11 a-d one track saves fresh ids through fr pl en and restored es without changing prior languages (37.8s)
```
NEXT: Revert c and expect pre-restore replacement text for d mutation.
FACT: R11c reverted; R11d-red deliberately compares the restored es PUT against pre-restore replacementTexts. Entered/restored constants are unchanged.
NEXT: Run M3-R11 only; d must fail after a–c pass.
FACT: R11d mutation runtime active; checked journal contains verbatim R8/R11b/R11c reporter evidence. No mutation has affected product code.
NEXT: Complete d-red, revert, then run all fixed tests positively.
FACT: Inspected R11d runtime reporter; source remains fixed except the deliberate d expectation argument.
NEXT: Record failure and restore final expectedTexts.
FACT: Waited for R11d real recognition and sequential save/restore; no runtime manipulation.
NEXT: Capture d-red before reverting.
FACT: R11d-red exit 1 as intended at check(expectedTexts): expected pre-restore English replacement; received restored Spanish variant. Reporter:
```text
  ✘  5 [chromium] › e2e/asr.spec.ts:972:3 › ASR acceptance with a stored admin session › M3-R11 a-d one track saves fresh ids through fr pl en and restored es without changing prior languages (45.7s)
```
NEXT: Revert d, run fixed R11/R12b/R15 positively.
FACT: All four intentional expectation mutations reverted. Correct source uses embeddedTexts/dubTexts and polishTexts/replacementTexts/restoredTexts.
NEXT: Positive targeted run for M3-R11|M3-R12b|M3-R15, isolated ms3c-fix1-targeted output.
FACT: Mutation command inventory (cwd frontend; prefix /Users/strelia/project-data/courses-kiri/courses-kiri; stdout/stderr logs at /tmp/<output-basename>.log):
- `npx playwright test e2e/asr.spec.ts --project=chromium -g 'M3-R8' --output=test-results/ms3c-fix1-mutation-r8-red` → exit 1, intended assertion mismatch.
- `npx playwright test e2e/asr.spec.ts --project=chromium -g 'M3-R8|M3-R11' --output=test-results/ms3c-fix1-r8-green-r11b-red` → exit 1, R8 green and b intended mismatch.
- `npx playwright test e2e/asr.spec.ts --project=chromium -g 'M3-R11' --output=test-results/ms3c-fix1-mutation-r11c-red` → exit 1, c intended mismatch.
- `npx playwright test e2e/asr.spec.ts --project=chromium -g 'M3-R11' --output=test-results/ms3c-fix1-mutation-r11d-red` → exit 1, d intended mismatch.
- Active positive: `npx playwright test e2e/asr.spec.ts --project=chromium -g 'M3-R11|M3-R12b|M3-R15' --output=test-results/ms3c-fix1-targeted`.
NEXT: Capture positive reporter lines; stop/question any proven product defect, then final checks/full-file once.
FACT: Positive targeted runtime started; all deliberate mutations have been removed.
NEXT: Verify R11 positive then both replacement-placement branches.
FACT: Setup passed for positive targeted run; waiting on real recognition and sequential checks.
NEXT: Inspect R11/R12 reporter outcomes.
FACT: R11 positive reporter PASS (41.6s); all independent typed/restored expectations, fresh IDs, and prior-language GET preservation checks passed. This is the green counterpart for mutation b, c and d.
NEXT: Verify matching non-default and no-match default placement through Save/readback.
FACT: Both R12b positive variants failed quickly (6.5s/6.7s), before reporter failure detail was printed; R15 independent cases continue. No product conclusion yet.
NEXT: Inspect failure context and request inventory immediately; STOP/question if a product defect is proven.
FACT: Corrected artifact lookup to Playwright hashed test directories; reading error context and trace error events to classify R12 failures.
NEXT: Identify exact assertion before any harness change.
FACT: R12b failures are harness timing: upload() sees the still-Ready old asset immediately after replacement mint; both fail at oldAssetId == newAssetId before any placement/restore assertion. Requests show replacement POST 201 followed by an immediate inventory read. This does not prove a product defect.
DECISION: Reinstate the observable promotion wait (selected tab becomes Unknown from old de) before reading the new asset identity; assert media ID changed afterward. No product edit.
NEXT: Let independent R15 finish, then apply this e2e-only wait and re-run R12b.
FACT: R15 independent runtime continues after R12 timing classification.
NEXT: Capture en/und switch results before editing the test snapshot.
FACT: R15 en PASS (1.1m): initial recognition finished, rename then explicit regeneration yielded en before mode toggle; author text remained through both recognition changes.
NEXT: Finish und case; record targeted run including expected harness failures, then re-run corrected R12b.
FACT: Positive targeted first attempt exit 1: 2 R12 timing failures, 8 passed (3.4m). R11 and both R15 variants passed. Complete reporter:
```text
  ✓   1 [setup] › e2e/auth.setup.ts:10:3 › authenticate as student (501ms)
  ✓   2 [setup] › e2e/auth.setup.ts:10:3 › authenticate as instructor (136ms)
  ✓   3 [setup] › e2e/auth.setup.ts:10:3 › authenticate as admin (107ms)
  ✓   4 [setup] › e2e/media.setup.ts:52:1 › generate the eleven media fixtures with the worker ffmpeg (32.7s)
  ✓   5 [chromium] › e2e/asr.spec.ts:972:3 › ASR acceptance with a stored admin session › M3-R11 a-d one track saves fresh ids through fr pl en and restored es without changing prior languages (41.6s)
  ✘   6 [chromium] › e2e/asr.spec.ts:1108:5 › ASR acceptance with a stored admin session › M3-R12b replacement offers detached lines in matching non-default track tab and saves without changing the original (6.5s)
  ✘   7 [chromium] › e2e/asr.spec.ts:1108:5 › ASR acceptance with a stored admin session › M3-R12b replacement offers detached lines in default fallback track tab and saves without changing the original (6.7s)
  ✓   8 [chromium] › e2e/asr.spec.ts:1261:5 › ASR acceptance with a stored admin session › M3-R15 en draft switch invalidates Load until Regenerate without replacing editor lines (1.1m)
  ✓   9 [chromium] › e2e/asr.spec.ts:1261:5 › ASR acceptance with a stored admin session › M3-R15 und draft switch invalidates Load until Regenerate without replacing editor lines (45.5s)
  ✓  10 [login-budget] › e2e/auth.teardown.ts:4:1 › B19 keeps host-originated logins within the nginx budget (4ms)
```
NEXT: Restore the R12 replacement-tab promotion wait and rerun only R12b.
FACT: R12b now waits for the selected tab to change from old de to replacement und (240s bounded observable wait), then requires a new media-inventory asset ID.
NEXT: Run both R12b variants with ms3c-fix1-placement output.
FACT: Corrected R12b targeted runtime active; R11/R15 evidence remains from the unchanged respective cases.
NEXT: Confirm replacement promotion, distinct placement, Save/readback and original preservation in both variants.
FACT: Inspected corrected placement run after fixture setup; awaiting final results.
NEXT: Record R12b success or inspect exact failure before proceeding.
FACT: Corrected placement attempt exit 1 (2 failed, 5 passed, 55.4s); inspecting exact failures before classifying them.
NEXT: Distinguish test inventory assumptions from product defects; no product edits.
FACT: R12b rerun reached real promotion but my inventory helper wrongly assumed old Ready assets disappear: media inventory correctly retains both old and new Ready rows (2). Trace shows new-track UI using the new video ID. Failure is the helper length assertion, not a product defect.
DECISION: Bind media identity to the minted upload videoId and verify that exact ID Ready in the inventory, rather than selecting an arbitrary Ready row or relying on ASR.
NEXT: Read the existing mint schema, scope both media reads by exact uploaded ID, then rerun placement.
FACT: upload() now returns the validated mint videoId; R12b requires each exact old/new ID in the media inventory with state Ready and distinct IDs. The inventory helper no longer assumes only one historical Ready row.
NEXT: Re-run only R12b under ms3c-fix1-placement-v2 output.
FACT: Checked patch placement: upload-ticket parsing is in the shared upload helper, old/new ID bindings and scoped inventory checks are in R12b.
NEXT: Observe placement-v2 runtime and classify any failure immediately.
FACT: Placement-v2 setup/runtime observed; asset identity proof now comes from real mint and inventory data.
NEXT: Collect both R12b results.
FACT: Inspected placement-v2 reporter after setup; no source edits while tests run.
NEXT: Confirm both placement/save/preservation outcomes before final checks.
FACT: Matching R12b placement-v2 failed after 19.3s; inspecting exact error/context before deciding whether this is product behavior or a test fixture issue.
NEXT: Classify this failure immediately while independent fallback runs.
FACT: Placement-v2 now passes media identity/promotion and track setup but no Restore offer is visible in either expected destination. This may be a product issue or snapshot capture timing; reading recovery ownership/capture flow before any rerun.
NEXT: Inspect actual snapshot version/language evidence and source contract; STOP the affected tests if product defect is established.
FACT: Source mechanism explains a fixture ambiguity: opening rename blurs/captures while language is still es; subsequent blur() on an unfocused textbox fires no event. A detached es snapshot correctly belongs in the es orphan, so destination absence alone is not a proven product defect.
DECISION: Explicitly focus/blur after rename and verify the real IndexedDB snapshot has language de and the typed text before replacement. This establishes the requested dirty-rename/capture precondition without synthetic storage.
NEXT: Add that snapshot precondition and re-run R12b; stop/question any product failure after the precondition is proven.
FACT: Added real post-rename focus/blur capture plus read-only IndexedDB precondition language=de, baseLanguage=es, lines=[typed text]; source snapshot attached before replacement.
NEXT: Run R12b under ms3c-fix1-placement-v3, preserving all previous failures.
FACT: R12b v3 runtime active with a verified-language snapshot precondition and exact uploaded media IDs.
NEXT: Observe precondition then placement/save assertions; stop affected row on a proven product defect.
FACT: Inspected R12b-v3 reporter after setup; source snapshot unchanged during execution.
NEXT: Record exact placement-v3 outcome.
FACT: ms3c-fix1-placement exit 1: 2 failed, 5 passed (55.4s), erroneous one-Ready-row inventory assumption. Command: courses-kiri npx playwright test e2e/asr.spec.ts --project=chromium -g M3-R12b --output=test-results/ms3c-fix1-placement (cwd frontend). Reporter:
```text
  ✓  1 [setup] › e2e/auth.setup.ts:10:3 › authenticate as student (509ms)
  ✓  2 [setup] › e2e/auth.setup.ts:10:3 › authenticate as instructor (113ms)
  ✓  3 [setup] › e2e/auth.setup.ts:10:3 › authenticate as admin (105ms)
  ✓  4 [setup] › e2e/media.setup.ts:52:1 › generate the eleven media fixtures with the worker ffmpeg (29.6s)
  ✘  5 [chromium] › e2e/asr.spec.ts:1108:5 › ASR acceptance with a stored admin session › M3-R12b replacement offers detached lines in matching non-default track tab and saves without changing the original (10.4s)
  ✘  6 [chromium] › e2e/asr.spec.ts:1108:5 › ASR acceptance with a stored admin session › M3-R12b replacement offers detached lines in default fallback track tab and saves without changing the original (12.2s)
  ✓  7 [login-budget] › e2e/auth.teardown.ts:4:1 › B19 keeps host-originated logins within the nginx budget (6ms)
```
FACT: ms3c-fix1-placement-v2 exit 1: 2 failed, 5 passed; missing post-rename capture precondition. Command: courses-kiri npx playwright test e2e/asr.spec.ts --project=chromium -g M3-R12b --output=test-results/ms3c-fix1-placement-v2 (cwd frontend). Reporter:
```text
  ✓  1 [setup] › e2e/auth.setup.ts:10:3 › authenticate as student (476ms)
  ✓  2 [setup] › e2e/auth.setup.ts:10:3 › authenticate as instructor (135ms)
  ✓  3 [setup] › e2e/auth.setup.ts:10:3 › authenticate as admin (94ms)
  ✓  4 [setup] › e2e/media.setup.ts:52:1 › generate the eleven media fixtures with the worker ffmpeg (31.9s)
  ✘  5 [chromium] › e2e/asr.spec.ts:1111:5 › ASR acceptance with a stored admin session › M3-R12b replacement offers detached lines in matching non-default track tab and saves without changing the original (19.3s)
  ✘  6 [chromium] › e2e/asr.spec.ts:1111:5 › ASR acceptance with a stored admin session › M3-R12b replacement offers detached lines in default fallback track tab and saves without changing the original (18.9s)
  ✓  7 [login-budget] › e2e/auth.teardown.ts:4:1 › B19 keeps host-originated logins within the nginx budget (5ms)
```
FACT: ms3c-fix1-placement-v3 exit 0: 7 passed (59.7s). Command: courses-kiri npx playwright test e2e/asr.spec.ts --project=chromium -g M3-R12b --output=test-results/ms3c-fix1-placement-v3 (cwd frontend). Reporter:
```text
  ✓  1 [setup] › e2e/auth.setup.ts:10:3 › authenticate as student (1.2s)
  ✓  2 [setup] › e2e/auth.setup.ts:10:3 › authenticate as instructor (290ms)
  ✓  3 [setup] › e2e/auth.setup.ts:10:3 › authenticate as admin (179ms)
  ✓  4 [setup] › e2e/media.setup.ts:52:1 › generate the eleven media fixtures with the worker ffmpeg (29.3s)
  ✓  5 [chromium] › e2e/asr.spec.ts:1111:5 › ASR acceptance with a stored admin session › M3-R12b replacement offers detached lines in matching non-default track tab and saves without changing the original (13.9s)
  ✓  6 [chromium] › e2e/asr.spec.ts:1111:5 › ASR acceptance with a stored admin session › M3-R12b replacement offers detached lines in default fallback track tab and saves without changing the original (13.3s)
  ✓  7 [login-budget] › e2e/auth.teardown.ts:4:1 › B19 keeps host-originated logins within the nginx budget (4ms)
```
FACT: Final R12b precondition proves actual IndexedDB snapshot de/base es; both placement variants now Save 200 and GET restored lines while exact es transcript stays unchanged. Earlier absence was an invalid capture precondition, not a proven product defect.
NEXT: Bind R8 embedded expectation to media trackId explicitly, run final checks, then one full asr.spec.ts.
FACT: R8 now identifies the embedded track from real media inventory by trackId and selects both independent draft expectations by exact trackId; asserts und and en-gb languages.
NEXT: Final authorized frontend format/typecheck/lint chain, preserving exact exit codes.
FACT: Final frontend checks: cd frontend && npm run format && npm run typecheck && npm run lint → each exit 0 (lint architecture 240 files). git diff --check exit 0. Product TSX/test remain at the input w1 hashes; frontend changes for this run are confined to asr.spec.ts.
NEXT: Run the whole e2e/asr.spec.ts once without -g, --project=chromium, output ms3c-fix1-full; record all reporter lines and summary.
FACT: Full-file run launched once without -g, with courses-kiri prefix and ms3c-fix1-full artifacts. Final spec SHA256 at launch: a3682a9296c465b62cfed54ed206f6048f8a5df1301b5eb7e890f8ef15e3f33f.
NEXT: Observe every full-file reporter result; keep source snapshot unchanged.
FACT: Full-file setup passed (auth roles and fixture generation 27.8s); ASR cases now run on the unchanged final snapshot.
NEXT: Collect initial ASR acceptance outcomes.
FACT: Full run C13 and C03 passed (29.6s, 24.9s): silent no-speech and real recognition progress/Ready both retained.
NEXT: Verify recognition-language and replacement regression cases.
FACT: Full-file run still active in real recognition; no failures or source changes observed.
NEXT: Collect C11 and following regression results.
FACT: Full run C11 passed (35.2s), including rename guard and empty Regenerate body. Run remains active with no failures.
NEXT: Continue C06/C17/C18 and per-track cases.
FACT: Full run C06 and C17 passed (39.0s, 25.1s). Existing replacement/mounted-editor behavior remains green on final source.
NEXT: Collect stale-recovery and per-track results.
FACT: Explicit mutation red/reverted-green reporter pairs. The R11 a–d positive sequence is the same green counterpart for each individually reverted b/c/d expectation mutation.
R8 (red then reverted green):
```text
  ✘  5 [chromium] › e2e/asr.spec.ts:672:3 › ASR acceptance with a stored admin session › M3-R8 R2 R4 success and R6 two-track drafts, saves and isolated reload recovery (59.9s)
  ✓  5 [chromium] › e2e/asr.spec.ts:672:3 › ASR acceptance with a stored admin session › M3-R8 R2 R4 success and R6 two-track drafts, saves and isolated reload recovery (1.0m)
```
R11b (red then reverted green):
```text
  ✘  6 [chromium] › e2e/asr.spec.ts:972:3 › ASR acceptance with a stored admin session › M3-R11 a-d one track saves fresh ids through fr pl en and restored es without changing prior languages (27.5s)
  ✓   5 [chromium] › e2e/asr.spec.ts:972:3 › ASR acceptance with a stored admin session › M3-R11 a-d one track saves fresh ids through fr pl en and restored es without changing prior languages (41.6s)
```
R11c (red then reverted green):
```text
  ✘  5 [chromium] › e2e/asr.spec.ts:972:3 › ASR acceptance with a stored admin session › M3-R11 a-d one track saves fresh ids through fr pl en and restored es without changing prior languages (37.8s)
  ✓   5 [chromium] › e2e/asr.spec.ts:972:3 › ASR acceptance with a stored admin session › M3-R11 a-d one track saves fresh ids through fr pl en and restored es without changing prior languages (41.6s)
```
R11d (red then reverted green):
```text
  ✘  5 [chromium] › e2e/asr.spec.ts:972:3 › ASR acceptance with a stored admin session › M3-R11 a-d one track saves fresh ids through fr pl en and restored es without changing prior languages (45.7s)
  ✓   5 [chromium] › e2e/asr.spec.ts:972:3 › ASR acceptance with a stored admin session › M3-R11 a-d one track saves fresh ids through fr pl en and restored es without changing prior languages (41.6s)
```
NEXT: Finish the single full-file run and copy every reporter line plus final summary.
FACT: Continued full run on final snapshot; explicit mutation red/green pairs now recorded for all four requested checks.
NEXT: Observe remaining cases and final cleanup.
FACT: Full run C18 and uploaded-default S6 passed (57.2s, 21.6s). No failure artifact observed.
NEXT: Finish two-track stale/regenerate case then final R8 journey.
FACT: Mid-run read-only snapshot check: HEAD remains 7b20d2c9b06f8457a274d18f0a19388f5bf906c0 and asr SHA256 remains a3682a9296c465b62cfed54ed206f6048f8a5df1301b5eb7e890f8ef15e3f33f. Final run is pinned to unchanged source.
NEXT: After runtime completion, verify run-created course IDs are gone using authenticated read-only GETs.
FACT: Full-file runtime observed without source changes.
NEXT: Capture remaining per-track outcomes.
FACT: Full run two-track S6 PASS (1.1m), including per-track isolation and rename guard. Final R8 with exact trackId draft selection is active.
NEXT: Verify R8 and PUT-failure recovery on the final snapshot.
FACT: Final R8 PASS (57.1s): und/en-gb independent drafts matched their corresponding PUTs and GETs, and two-tab recovery completed.
NEXT: Collect R4 failure/retry, multilingual upload, and R10 outcomes.
FACT: Full run M3-R4 PASS (26.4s), preserving reload/restore/retry after injected PUT failure. Remaining tests continue without failures.
NEXT: Observe R14/R10/R11 and placement regressions.
FACT: R12b placement-v3 media identity evidence from retained requests.json (the tests validated each ID from upload ticket against the exact media-inventory row before using it):
- asr-ASR-acceptance-with-a--51c7c-thout-changing-the-original-chromium: old/new=["6dc20a58-4bfe-4f7b-9272-e4a9f7f0261c", "6e4b2765-ac4b-45e7-9259-22b77961d69b"], writes=[{"language": "de", "texts": ["Знімок тієї самої мови"], "status": 200}].
- asr-ASR-acceptance-with-a--84d54-thout-changing-the-original-chromium: old/new=["6c2182fa-8dd6-4d9b-bb7a-12f6c4824b07", "59785a3c-ea09-4f0f-86da-3ca8ca4d3553"], writes=[{"language": "fr", "texts": ["Знімок у default-вкладці"], "status": 200}].
NEXT: Keep waiting for the unchanged full-file run, then capture its exact reporter and verify cleanup.
FACT: Full runtime continues with no source changes; retained targeted R12 media IDs and PUT language/text/status now copied to the journal.
NEXT: Collect upcoming full-run reporter lines.
FACT: Full run R14 and foreign-rename R10 passed (46.7s, 30.6s), including zero transcript PUT on refused Save.
NEXT: Complete GET-503 retry and independent R11 expectations.
FACT: Full run GET-503 R10 PASS (20.0s), retaining lines with zero PUT then saving on retry.
NEXT: Collect R11 and R12 final-snapshot outcomes.
FACT: Final R11 PASS (34.2s), R12a PASS (4.8s), matching non-default R12b PASS (13.6s). Full run has no failures so far.
NEXT: Verify fallback placement, both conflict branches and mode switches.
FACT: Final fallback R12b PASS (13.8s) and R13 open/restored conflict branches PASS (6.0s/7.0s). Both replacement Save/readback/preservation variants passed in the full run.
NEXT: Finish en/und switches and per-course cleanup before reporting final outcome.
FACT: Full-file mode-switch recognition still active; no failures observed and final spec unchanged.
NEXT: Capture en/und results and teardown summary.
FACT: Continued observing the final full-file run without edits; mode-switch cases are the remaining recognition work.
NEXT: Read final R15 reporter lines and cleanup outcome.
FACT: Final R15 en PASS (1.6m), completing the non-racing setup and retaining author lines through mode change. und remains active.
NEXT: Finish und and teardown; then verify all run-created courses absent by GET.
FACT: Required full-file run exited 0. Exact command (cwd frontend): /Users/strelia/project-data/courses-kiri/courses-kiri npx playwright test e2e/asr.spec.ts --project=chromium --output=test-results/ms3c-fix1-full > /tmp/ms3c-fix1-full.log 2>&1. Ran once without -g. All reporter lines and summary:
```text
  ✓   1 [setup] › e2e/auth.setup.ts:10:3 › authenticate as student (480ms)
  ✓   2 [setup] › e2e/auth.setup.ts:10:3 › authenticate as instructor (144ms)
  ✓   3 [setup] › e2e/auth.setup.ts:10:3 › authenticate as admin (107ms)
  ✓   4 [setup] › e2e/media.setup.ts:52:1 › generate the eleven media fixtures with the worker ffmpeg (27.8s)
  ✓   5 [chromium] › e2e/asr.spec.ts:446:3 › ASR acceptance with a stored admin session › C13 shows no speech for a silent video (29.6s)
  ✓   6 [chromium] › e2e/asr.spec.ts:457:3 › ASR acceptance with a stored admin session › C03 observes recognition in progress and then a ready draft (24.9s)
  ✓   7 [chromium] › e2e/asr.spec.ts:476:3 › ASR acceptance with a stored admin session › C11 the recognition hint comes from the track language and Regenerate sends no body (35.2s)
  ✓   8 [chromium] › e2e/asr.spec.ts:499:3 › ASR acceptance with a stored admin session › C06 after replacing the video the old draft is not offered (39.0s)
  ✓   9 [chromium] › e2e/asr.spec.ts:528:3 › ASR acceptance with a stored admin session › C17 keeps editor text mounted and offers the recognized draft (25.1s)
  ✓  10 [chromium] › e2e/asr.spec.ts:561:3 › ASR acceptance with a stored admin session › C18 offers stale recovery after an unsaved draft is regenerated (57.2s)
  ✓  11 [chromium] › e2e/asr.spec.ts:592:3 › ASR acceptance with a stored admin session › S6 a video-only lesson is recognised from its uploaded default track (21.6s)
  ✓  12 [chromium] › e2e/asr.spec.ts:634:3 › ASR acceptance with a stored admin session › S6 every Ready track has its own draft; renaming makes its draft stale until Regenerate (1.1m)
  ✓  13 [chromium] › e2e/asr.spec.ts:675:3 › ASR acceptance with a stored admin session › M3-R8 R2 R4 success and R6 two-track drafts, saves and isolated reload recovery (57.1s)
  ✓  14 [chromium] › e2e/asr.spec.ts:806:3 › ASR acceptance with a stored admin session › M3-R4 PATCH succeeds and PUT fails, reload restores the track key and Save retries (26.4s)
  ✓  15 [chromium] › e2e/asr.spec.ts:874:3 › ASR acceptance with a stored admin session › M3-R14 a track minted with the multilingual switch gets a ready draft (46.7s)
  ✓  16 [chromium] › e2e/asr.spec.ts:891:3 › ASR acceptance with a stored admin session › M3-R10 a foreign track rename after Load refuses Save without a transcript PUT (30.6s)
  ✓  17 [chromium] › e2e/asr.spec.ts:936:3 › ASR acceptance with a stored admin session › M3-R10 GET tracks 503 keeps lines and sends no PUT; retry saves (20.0s)
  ✓  18 [chromium] › e2e/asr.spec.ts:983:3 › ASR acceptance with a stored admin session › M3-R11 a-d one track saves fresh ids through fr pl en and restored es without changing prior languages (34.2s)
  ✓  19 [chromium] › e2e/asr.spec.ts:1065:3 › ASR acceptance with a stored admin session › M3-R12a typing during first processing is offered on Ready without blur or reload (4.8s)
  ✓  20 [chromium] › e2e/asr.spec.ts:1119:5 › ASR acceptance with a stored admin session › M3-R12b replacement offers detached lines in matching non-default track tab and saves without changing the original (13.6s)
  ✓  21 [chromium] › e2e/asr.spec.ts:1119:5 › ASR acceptance with a stored admin session › M3-R12b replacement offers detached lines in default fallback track tab and saves without changing the original (13.8s)
  ✓  22 [chromium] › e2e/asr.spec.ts:1249:5 › ASR acceptance with a stored admin session › M3-R13 a open T1 editor refuses Save after foreign T2 (6.0s)
  ✓  23 [chromium] › e2e/asr.spec.ts:1249:5 › ASR acceptance with a stored admin session › M3-R13 b restored T1 snapshot refuses Save after foreign T2 (7.0s)
  ✓  24 [chromium] › e2e/asr.spec.ts:1315:5 › ASR acceptance with a stored admin session › M3-R15 en draft switch invalidates Load until Regenerate without replacing editor lines (1.6m)
  ✓  25 [chromium] › e2e/asr.spec.ts:1315:5 › ASR acceptance with a stored admin session › M3-R15 und draft switch invalidates Load until Regenerate without replacing editor lines (52.0s)
  ✓  26 [login-budget] › e2e/auth.teardown.ts:4:1 › B19 keeps host-originated logins within the nginx budget (3ms)

  Slow test file: [chromium] › e2e/asr.spec.ts (11.8m)
  26 passed (12.3m)
```
FACT: Final full-run R12b identities and writes, bound by the tests to real upload tickets and exact media-inventory rows:
- asr-ASR-acceptance-with-a--51c7c-thout-changing-the-original-chromium: old/new=["5cec49c4-5682-434d-8440-f22e1b45def5", "56b1ac2a-3c5a-4e44-9dc8-36a13c9d26bc"], writes=[{"language": "de", "texts": ["Знімок тієї самої мови"], "status": 200}].
- asr-ASR-acceptance-with-a--84d54-thout-changing-the-original-chromium: old/new=["c0befb77-313e-4666-b8ee-5c6c26a676b6", "c501acce-2e2d-4556-870d-bb91bb049ceb"], writes=[{"language": "fr", "texts": ["Знімок у default-вкладці"], "status": 200}].
NEXT: Verify cleanup by exact run-created IDs, final changed files and source SHA256; hand off to Courses 1 for w2.
FACT: Full-file result recorded verbatim: 26 passed (12.3m), exit 0; slow ASR file 11.8m. Read-only cleanup audit started over exact IDs from this run output directories only.
NEXT: Record cleanup results and final file hashes; no further code changes.
FACT: Cleanup audit exit 0: 37 exact run-created course IDs across 9 output directories all return GET /admin/courses/{id} → 404. No deletes or repairs in this audit; all removal came from afterAll. Raw evidence /tmp/ms3c-fix1-cleanup.json. IDs/statuses:
```json
[
  {
    "id": "56b7f3c4-1c4b-42ed-b1ba-1b9bc0827c2d",
    "status": 404
  },
  {
    "id": "b09e521a-388a-4835-ac7a-78b8ccda09b8",
    "status": 404
  },
  {
    "id": "7abbc6c7-0cf6-401e-ac7f-491686e824df",
    "status": 404
  },
  {
    "id": "6446aad0-6ef8-4c8f-aeb0-c09c3abcee4b",
    "status": 404
  },
  {
    "id": "b3285f19-1fba-4225-99fa-a4b886785fb2",
    "status": 404
  },
  {
    "id": "4b1f3457-b192-46ad-94f4-2eb96e8c6624",
    "status": 404
  },
  {
    "id": "44176d14-69f5-4936-8831-b120eaf3d1b8",
    "status": 404
  },
  {
    "id": "04cdc14e-cde1-4b33-a164-b7ed0d9a09b8",
    "status": 404
  },
  {
    "id": "3e89028d-e460-4396-90f3-026aca9afe64",
    "status": 404
  },
  {
    "id": "2fcb30d4-34c2-4c48-b00d-0a0c7d89e58d",
    "status": 404
  },
  {
    "id": "31d782d6-fd99-4dc3-a587-0ca350990d09",
    "status": 404
  },
  {
    "id": "e9d10067-c541-4872-a0be-775888f43a63",
    "status": 404
  },
  {
    "id": "687c1cfa-ced8-433e-a014-984caf39f042",
    "status": 404
  },
  {
    "id": "c27072f7-4548-4ee0-b659-765a1f560ef3",
    "status": 404
  },
  {
    "id": "2db20ddc-e394-4df9-81a5-2736bb5db238",
    "status": 404
  },
  {
    "id": "1ff83b8c-d010-4f74-8968-bcaf1a3f6149",
    "status": 404
  },
  {
    "id": "3f342e3a-3091-4b32-be54-986a6c066de9",
    "status": 404
  },
  {
    "id": "095c2fa8-b100-460c-9131-6104bf26b528",
    "status": 404
  },
  {
    "id": "3005dc49-d54b-4fd7-8f8e-f39a060aa051",
    "status": 404
  },
  {
    "id": "c717ac04-d42d-4c03-92df-7dd9615c502f",
    "status": 404
  },
  {
    "id": "29a0abcf-6345-424e-8c24-d4d3ae9f8192",
    "status": 404
  },
  {
    "id": "48fbd14e-179d-4a01-918c-4de2b0c31feb",
    "status": 404
  },
  {
    "id": "460cba80-e22d-4a27-8b65-6b50fe076084",
    "status": 404
  },
  {
    "id": "b48f155c-d276-4b2e-9e46-08aaf8e8bb98",
    "status": 404
  },
  {
    "id": "2762f7f1-2925-4a2f-b3d1-0b25879d7e8c",
    "status": 404
  },
  {
    "id": "5dc046a3-449a-489b-b495-73f34589f566",
    "status": 404
  },
  {
    "id": "7ac13c69-bce0-45c2-9ee2-f5765e2e6817",
    "status": 404
  },
  {
    "id": "9a899bc5-b49b-4a98-8690-b764d5725044",
    "status": 404
  },
  {
    "id": "76aa9796-cd04-454d-bbc5-50993d04e73a",
    "status": 404
  },
  {
    "id": "65ecdec3-c93f-459e-821a-f340d3d75fd9",
    "status": 404
  },
  {
    "id": "1ca845c9-4771-40fc-b160-ea752d947dce",
    "status": 404
  },
  {
    "id": "5238b6d1-2900-4021-b933-4acb0d63a964",
    "status": 404
  },
  {
    "id": "c9fcbb60-e7fc-4150-8bce-417a4575f4ad",
    "status": 404
  },
  {
    "id": "85f3d810-bfa8-4d37-bc51-4da79d9f3297",
    "status": 404
  },
  {
    "id": "b3534bae-ed09-476d-9fc0-962fb3e11528",
    "status": 404
  },
  {
    "id": "cd271b24-b5cf-414b-9c9c-485357c21c33",
    "status": 404
  },
  {
    "id": "879ba52c-80de-46af-8dee-40b08d2335dc",
    "status": 404
  }
]
```
FACT: Done for this fix run: 3 Major + 1 Minor addressed in frontend/e2e/asr.spec.ts; every mutation red/reverted-green pair and all full-file reporter lines copied above. Final frontend format/typecheck/lint all 0; full-file 26 passed (12.3m), exit 0.
FACT: Changed files attributable to this run: frontend/e2e/asr.spec.ts and docs/multi-subtitles/_wip-ms3c-fix1.md only. SHA-256 frontend/e2e/asr.spec.ts = a3682a9296c465b62cfed54ed206f6048f8a5df1301b5eb7e890f8ef15e3f33f. No commit, checkout/reset/stash/clean/rm, product/backend/generated edit or runtime rebuild/reset performed.
FACT: Limitations: cleanup exception continuation is implemented but no artificial transport failure was injected; normal cleanup audited as above. Whole-repository ordered make check and independent w2 review are NOT_RUN by this fix run and remain with Courses 1. No proven product defect or open QUESTION arose.
NEXT: Courses 1 reviews this exact final asr hash as w2, then owns the separately agreed full ordered gate/commit/PR.
FACT: Final audit complete: git diff --check exit 0; branch/HEAD and pre-existing dirty-file inventory preserved; spec SHA256 still a3682a9296c465b62cfed54ed206f6048f8a5df1301b5eb7e890f8ef15e3f33f. Existing media product TSX/test hashes exactly match the w1 input.
NEXT: Hand off journal and final e2e snapshot for w2; this authorized fix run is complete.
