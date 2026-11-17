NEXT: Read repository briefing, MS3 table and checkpoint decisions, then inspect existing ASR tests and Ukrainian UI names before changing e2e.
FACT: Initial instructions and repository briefing/adapter read; user authorizes e2e-only work against the existing kiri stack, no product fixes or commits.
NEXT: Resolve exact MS3 checkpoint, API response shapes, tab/restore names and existing fixture helpers.
FACT: HEAD is 7b20d2c9b06f8457a274d18f0a19388f5bf906c0; relevant asr.spec.ts initially clean. Existing unrelated dirty/untracked files preserved. MS3c checkpoint is READY, dated decisions place M3-R6 here and require en-GB mint through the real upload API.
DECISION: Resume the accepted MS3c checkpoint for run-1 scope only; no new product contract or review route. Use a one-shot mint-body rewrite to en-GB while retaining UI upload and real bytes.
NEXT: Finish API/UI inspection, then replace obsolete draft helpers and add the three run-1 journeys.
FACT: UI labels come from frontend/src/app/providers/locales/uk.ts; all track editors force-mount, so selectors must use the visible selected tabpanel. API draft list carries per-track identity and trackLanguage. Existing queue pause helper invokes bare docker and must use the required wrapper if retained.
NEXT: Implement validated per-track helpers, scoped migration cases, upload rewrite, recovery key assertions and network body proofs.
FACT: Migrated obsolete draft URLs and selectors; added real two-track/success/reload, PUT-failure/retry, and multilingual journeys. Added per-test request logs/failure screenshots and scoped IndexedDB cleanup; course registration now precedes later seed steps to preserve cleanup after partial failures.
NEXT: Run format/typecheck/lint and resolve test-code issues before executing against kiri.
FACT: First format exited 0; first typecheck exited 2 (ES target lacks Array.findLast); lint not run in that chain. Corrected to reverse().find(), and verified mint returns an upload ticket while the track inventory uses items.
NEXT: Re-run the three local checks and execute all touched ASR tests against the wrapper-managed stack.
FACT: Started wrapper-prefixed full asr.spec.ts targeted-file run (all touched cases) with output test-results/ms3c-run1-initial; setup regenerates fixtures and auth states.
NEXT: Inspect reporter output; distinguish fixture/selector defects from proven product defects, preserve evidence and continue other cases.
FACT: Final local chain at this snapshot passed: npm run format exit 0, npm run typecheck exit 0, npm run lint exit 0 (architecture 240 files). Only frontend/e2e/asr.spec.ts changed under frontend.
NEXT: Await initial runtime evidence while checking remaining test-only race/cleanup assumptions.
FACT: Initial e2e exit 1 before ASR cases: setup blocked by sandbox EPERM localhost:8080 and permission denied on kiri Docker socket; 13 did not run. Evidence: frontend/test-results/ms3c-run1-initial/*setup*/trace.zip. This is runtime access, not a product failure.
NEXT: Retry the same authorized wrapper command with sandbox escalation for localhost and kiri socket.
FACT: Escalated wrapper run started successfully (session 88483), isolated output frontend/test-results/ms3c-run1-live; all 11 ASR cases are selected plus setup/teardown.
NEXT: Observe first live scenarios and retain exact reporter lines for each case.
FACT: Live setup authenticated all three fixture roles successfully; media fixture generation is still running. Checked actual IndexedDB key format: user:lesson:track:<id>. No nested AGENTS.md found in frontend or docs/multi-subtitles.
NEXT: Run through the first silent/speech cases; repair only e2e harness errors, and stop any case that proves a product defect.
FACT: Media setup passed (28.5s), real generated speech/dub fixtures are available; ASR tests now running. Snapshot used for first run: asr SHA256 b38480afc922c244804fb36965309ce20842629ef980b51dfc9f66b5808aa3f6; transcript.spec unchanged SHA256 72afbf3a3a38ab3a45b108f9bbea67e26c71f1832c131aaac9dddc1b436c29b4.
NEXT: Await production recognition results and inspect any failed assertion artifacts.
FACT: Diff whitespace check passed. Runtime evidence collection remains scoped to admin lesson/video JSON calls, without auth headers or URL query strings.
NEXT: Continue live run; report exact test outcomes, not aggregate setup success.
FACT: Reporter: ✓ 5 [chromium] › e2e/asr.spec.ts:375:3 › ASR acceptance with a stored admin session › C13 shows no speech for a silent video (27.1s)
NEXT: Observe C03 and C11 against the same live stack.
FACT: Reporter: ✓ 6 [chromium] › e2e/asr.spec.ts:386:3 › ASR acceptance with a stored admin session › C03 observes recognition in progress and then a ready draft (25.0s)
FACT: Intermediate accidental check invocation before the ES-target correction also exited 2 at the same findLast diagnostic; superseded by the successful full format/typecheck/lint chain already recorded.
NEXT: Check C11 regeneration and video replacement, followed by editor recovery.
FACT: C11 is still running after the first two passes; no failing assertion has yet been emitted.
NEXT: Read only recent ASR worker logs to distinguish recognition latency from an e2e/UI wait.
FACT: Recent worker logs show recognition completed normally (4.288 speech seconds, one segment, en). Continuing to check the UI wait against actual localized labels.
NEXT: Verify rename button accessible name before classifying the long C11 wait.
FACT: C11 wait is an e2e selector defect: helper searched /^Редагувати/ but locales/uk.ts:876 renders Змінити доріжку {{name}}. No product defect is proven.
DECISION: Interrupt this mistaken harness run, correct the localized selector, retain C13/C03 passes, and rerun the remaining cases from C11.
NEXT: Stop only the identified Playwright process and update the test helper.
FACT: Interrupted live run exit 130; C11 interrupted at the proven wrong rename selector (1.5m), 9 subsequent tests not run. Evidence: frontend/test-results/ms3c-run1-live/asr-ASR-acceptance-with-a--55437-nd-Regenerate-sends-no-body-chromium/{requests.json,failure.png,trace.zip}. Corrected selector to /^Змінити доріжку/ and bounded UI actions to 30s; recognition retains its long explicit timeout.
NEXT: Recheck local gates and rerun all ASR cases with fresh setup and a separate evidence directory.
FACT: Corrected rename helper now uses the exact localized prefix from uk.ts; course cleanup token is refreshed immediately after the course PUT. New full-file run uses frontend/test-results/ms3c-run1-v2.
NEXT: Collect v2 reporter lines and inspect the first failing assertion, if any.
FACT: After the localized selector correction, npm run format, npm run typecheck and npm run lint all exited 0 again.
NEXT: Continue v2 runtime verification; do not modify product code for test harness issues.
FACT: v2 reporter: ✓ 5 [chromium] › e2e/asr.spec.ts:375:3 › ASR acceptance with a stored admin session › C13 shows no speech for a silent video (37.2s)
NEXT: Await v2 C03/C11 results.
FACT: v2 reporter: ✓ 6 [chromium] › e2e/asr.spec.ts:386:3 › ASR acceptance with a stored admin session › C03 observes recognition in progress and then a ready draft (25.3s)
NEXT: Check C11 and C06; later scenarios remain pending, not passed.
FACT: v2 reporter: ✓ 7 [chromium] › e2e/asr.spec.ts:405:3 › ASR acceptance with a stored admin session › C11 the recognition hint comes from the track language and Regenerate sends no body (45.5s)
FACT: C11 proved no recognition-language control, und→en rename, bodyless per-track Regenerate, and the new draft snapshot trackLanguage=en.
NEXT: Observe C06 replacement and C17/C18 editor/recovery cases.
FACT: v2 reporter: ✓ 8 [chromium] › e2e/asr.spec.ts:428:3 › ASR acceptance with a stored admin session › C06 after replacing the video the old draft is not offered (53.0s)
NEXT: Verify C17/C18, then the video-only and two-Ready-track cases.
FACT: v2 reporter: ✓ 9 [chromium] › e2e/asr.spec.ts:457:3 › ASR acceptance with a stored admin session › C17 keeps editor text mounted and offers the recognized draft (25.2s)
NEXT: Verify stale recovery C18 and uploaded-track recognition.
FACT: v2 reporter: ✓ 10 [chromium] › e2e/asr.spec.ts:490:3 › ASR acceptance with a stored admin session › C18 offers stale recovery after an unsaved draft is regenerated (41.7s)
NEXT: Verify S6 video-only, S6 independent drafts/stale rename, then run-1 acceptance journeys.
FACT: v2 reporter: ✓ 11 [chromium] › e2e/asr.spec.ts:521:3 › ASR acceptance with a stored admin session › S6 a video-only lesson is recognised from its uploaded default track (54.6s)
NEXT: Observe the two-Ready-track stale/regenerate case and the new M3 journeys.
FACT: v2 reporter: ✘ 12 [chromium] › e2e/asr.spec.ts:563:3 › ASR acceptance with a stored admin session › S6 every Ready track has its own draft; renaming makes its draft stale until Regenerate (51.9s)
FACT: v2 reporter: ✘ 13 [chromium] › e2e/asr.spec.ts:604:3 › ASR acceptance with a stored admin session › M3-R8 R2 R4 success and R6 two-track drafts, saves and isolated reload recovery (8.8s)
NEXT: Inspect error contexts/request evidence immediately; leave failed tests stopped while remaining cases run.
FACT: M3-R4 failure/retry passed (26.6s). S6 rename succeeded server-side, but the selected panel switched from the dub to default und and the page displayed Зміни збережено.; inspecting course writes to determine whether rename submitted the parent form.
FACT: M3-R8 stopped on a harness assumption: raw media inventory preserves en-GB casing (contract allows mixed case); transcript tab labels and draft snapshots normalize lower case. No product defect established by that assertion.
NEXT: Record any proven rename defect as QUESTION; adjust only the casing assumption and cleanup harness after remaining tests finish.
QUESTION: Courses 1 — proven product defect in S6 rename. After selecting the English dub and PATCHing its language to fr (200), the page also PUTs /api/v1/admin/courses/ce98619a-a611-48bc-b19c-f40ceea3cd39 (200), shows “Зміни збережено.” and switches selected transcript tab to default “Невідома мова”. The same-panel stale-message assertion fails (not found); it cannot proceed to Regenerate. Test STOPPED, no product fix. Evidence: frontend/test-results/ms3c-run1-v2/asr-ASR-acceptance-with-a--19112-raft-stale-until-Regenerate-chromium/{rename-writes.json,requests.json,failure.png,trace.zip,error-context.md}. Please decide/fix the parent-form submit + tab reset. Raw trace also proves an unexpected course PUT in C11; cleanup’s stale token then got 409 course_modified.
DECISION: Fix only the test harness: compare raw media language case-insensitively, assert en-gb on the transcript tab/draft, and refresh each created course token before afterAll deletion so a real UI course write cannot prevent cleanup. Retain the failing S6 requirement; do not bypass the tab reset.
FACT: v2 reporter: ✓ 14 [chromium] › e2e/asr.spec.ts:734:3 › ASR acceptance with a stored admin session › M3-R4 PATCH succeeds and PUT fails, reload restores the track key and Save retries (26.6s)
NEXT: Let multilingual finish, repair the casing/cleanup harness, clean any own leftovers, and rerun M3-R8 only.
FACT: v2 run exit 1: 3 failed / 13 passed including setup/teardown. Reporter: ✘ 15 [chromium] › e2e/asr.spec.ts:802:3 › ASR acceptance with a stored admin session › M3-R14 a track minted with the multilingual switch gets a ready draft (35.1s). It stopped before mint: Playwright check() targets the switch native checkbox while its decorative track intercepts pointer events.
NEXT: Inspect the Switch primitive to distinguish expected label/input composition from a real broken control; use a supported real keyboard interaction if the role is a hidden-input adapter, without force or DOM mutation.
FACT: Switch is React Aria’s label/input composition with a decorative ui-switch-track; changed the e2e to focus the named switch and press Space, then assert checked (real keyboard path; no force/DOM mutation). This pointer-target failure did not prove the visible label is broken.
FACT: Test-only cleanup now GETs a fresh course token and uses soft assertions so all registered courses are attempted. Logger now includes course writes for direct evidence of the reported rename defect.
NEXT: Clean earlier run-scoped leftovers using fresh API tokens, then rerun only M3-R8 and M3-R14 with fresh setup.
FACT: Final harness-only format/typecheck/lint chain all exited 0. v3 targeted M3-R8|M3-R14 started with fresh setup; the reported product-defect S6 case is not rerun or bypassed.
FACT: Initial cleanup discovery found zero namespaces because request bodies are resource blobs; corrected discovery reads course response blobs from traces. It identified four exact own namespaces, deleted 6 leftover courses with 204, and verified all four namespace searches empty. Evidence: frontend/test-results/ms3c-run1-cleanup.json.
NEXT: Collect v3 two-track and multilingual results; preserve all previous failure evidence and questions.
FACT: v3 reporter: ✘ 5 [chromium] › e2e/asr.spec.ts:614:3 › ASR acceptance with a stored admin session › M3-R8 R2 R4 success and R6 two-track drafts, saves and isolated reload recovery (36.2s)
NEXT: Inspect this assertion and evidence while M3-R14 continues.
FACT: v3 reporter: ✓ 6 [chromium] › e2e/asr.spec.ts:812:3 › ASR acceptance with a stored admin session › M3-R14 a track minted with the multilingual switch gets a ready draft (28.7s)
FACT: M3-R8 reached two ready drafts with the expected distinct Welcome/dubbed texts, PATCH→PUT en 200, PUT en-gb 200, inventory [en,en-gb], and track renamed en. It then failed on GET /admin/lessons/{id}/transcript (405), which may be the wrong read endpoint; checking OpenAPI before classifying it.
NEXT: Correct the GET helper if it uses a write-only route; rerun the independent two-tab proof.
FACT: OpenAPI confirms GET /api/v1/learn/lessons/{lessonId}/transcript is the read endpoint; /admin/lessons/{lessonId}/transcript is PUT-only. Corrected this harness URL. v3 exit 1 (one failed ASR case, six passed including setup/teardown); M3-R14 is passed.
NEXT: Re-run only M3-R8 with the correct production GET endpoint and finish the two-tab reload assertions.
FACT: After the GET URL correction, npm run format, npm run typecheck, npm run lint all exit 0. No product or generated files changed.
NEXT: Await v4 M3-R8 result, then verify final diff, evidence inventory and cleanup.
FACT: v4 setup passed (media fixtures 27.0s); M3-R8 is running against real uploads/recognition. S6 remains stopped with the recorded QUESTION; no attempt to route around that product defect.
NEXT: Finish transcript GET equality and IndexedDB two-tab restore/discard/save assertions.
FACT: v4 reporter: ✓ 5 [chromium] › e2e/asr.spec.ts:614:3 › ASR acceptance with a stored admin session › M3-R8 R2 R4 success and R6 two-track drafts, saves and isolated reload recovery (46.5s)
FACT: v4 exit 0, 6 passed (1.3m) including setup/teardown. Proved both transcript GET languages/texts equal their PUT bodies; both track keys existed; reload isolation, Restore, Discard, re-entry and Save retirement passed.
NEXT: Verify all recorded own namespaces are empty, final diff/file hashes, and append the concise handoff with remaining S6 failure.
FACT: Final cleanup verification: all 4 prior exact namespace searches are empty; all 3 later course IDs observed in v3/v4 page logs return 404 after afterAll. Initial stale-token leftovers: 6 deleted with 204. Evidence: frontend/test-results/ms3c-run1-cleanup.json (initial 6 deletions) and frontend/test-results/ms3c-run1-cleanup-final.json (final checks). Each test clears its own course-scoped IndexedDB keys in afterEach, including failures.
FACT: Final verification commands (cwd frontend): `npm run format && npm run typecheck && npm run lint` — format 0, typecheck 0, lint 0. `git diff --check` — 0. No full ordered make check, backend/integration/build, or transcript.spec run in this run-1 scope; these are NOT_RUN, not passed.
FACT: Runtime command inventory (cwd frontend; all npx invocations prefixed with /Users/strelia/project-data/courses-kiri/courses-kiri): (1) `npx playwright test e2e/asr.spec.ts --project=chromium --output=test-results/ms3c-run1-initial` exit 1, sandbox-blocked setup, no ASR cases ran; (2) same with output ms3c-run1-live, exit 130 after interrupting wrong Ukrainian selector; (3) same with output ms3c-run1-v2, exit 1, 3 failed/13 passed including setup/teardown; (4) `npx playwright test e2e/asr.spec.ts --project=chromium -g 'M3-R8|M3-R14' --output=test-results/ms3c-run1-v3` exit 1, M3-R14 passed, M3-R8 wrong GET URL; (5) `npx playwright test e2e/asr.spec.ts --project=chromium -g 'M3-R8' --output=test-results/ms3c-run1-v4` exit 0, 6 passed (1.3m) including setup/teardown. Exact per-case reporter lines and earlier harness failures are preserved above.
FACT: Final per-test disposition (latest applicable targeted proof; durations copied from reporter): C13 PASS 37.2s (v2); C03 PASS 25.3s (v2); C11 PASS 45.5s (v2); C06 PASS 53.0s (v2); C17 PASS 25.2s (v2); C18 PASS 41.7s (v2); S6 video-only PASS 54.6s (v2); S6 every Ready track / stale rename FAIL 51.9s (v2, stopped at product defect, QUESTION above); M3-R8/R2/R4-success/R6 PASS 46.5s (v4); M3-R4 PUT-failure/reload/retry PASS 26.6s (v2); M3-R14 PASS 28.7s (v3). Total: 10 passed, 1 failed. Earlier passes retained only where later changes were logging/cleanup or helpers outside that case’s behavior; no final all-file green run is claimed.
FACT: Acceptance disposition: M3-R2, M3-R4 success+failure, M3-R6 two-tab e2e, M3-R8 and M3-R14 proven in run 1. M3-R9 migration implemented but NOT DONE because S6 rename cannot complete. M3-R1 transcript.spec stays byte-unchanged and was not run here. Other MS3 rows and final ordered gate belong to later MS3c work.
FACT: Files changed by this task: frontend/e2e/asr.spec.ts; docs/multi-subtitles/_wip-ms3c-run1.md (append-only). Evidence artifacts under frontend/test-results/ms3c-run1-* are runtime outputs. No frontend/src, backend, generated files or transcript.spec changes; no checkout/reset/stash/clean/rm or commit. Existing unrelated WIP preserved.
FACT: Final snapshot: branch ms2-3-asr-per-track, base/HEAD 7b20d2c9b06f8457a274d18f0a19388f5bf906c0. asr.spec.ts SHA256 3c091b372b304dbf15c6933b7d38664dae44f744cc4a03876e951b6f8e211090; transcript.spec.ts SHA256 72afbf3a3a38ab3a45b108f9bbea67e26c71f1832c131aaac9dddc1b436c29b4 (same as start).
NEXT: Courses 1 decides/fixes the recorded rename parent-course PUT/tab-reset defect; rerun the stopped S6 test after that fix. Run 1 is reported with this failure, not declared done. Continue the remaining MS3c rows and full ordered gate under the coordinator’s scope.
