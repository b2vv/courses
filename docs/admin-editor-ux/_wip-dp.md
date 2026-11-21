NEXT: Read agent-entry and repository-local instructions, then inspect the Ux17/U6 plan and reference harness before creating evidence.
FACT: Owner scope: evidence only on admin-editor-ux at a09d960f; no product edits, git mutations, or commits. All UI defects will be recorded as findings.
FACT: Read shared entry rules and requested repository context. Initial inspection is read-only except this append-only evidence log.
NEXT: Read reference harness, local design plan and fixture lifecycle; verify branch/revision and live-stack routes.
FACT: Verified repository root <repo>, branch admin-editor-ux, HEAD a09d960f. Existing unrelated dirty/untracked files are preserved.
DECISION: Evidence-only owner override applies: use design-pass + ui-measurements, record findings without product fixes. No agents, git writes, or aggregate gates required for this harness run.
NEXT: Adapt the existing temporary Playwright harness and isolated fixture lifecycle to D1–D8 and seven viewport/text-scale combinations.
FACT: Reference harness uses isolated browser contexts, authStatePath(admin), API fixture setup, UI file upload, per-state geometry/contrast, and SQL cleanup in afterAll. Evidence will use JPEG quality 45 rather than reference PNG.
NEXT: Resolve exact editor selectors and SQL cleanup tables; inspect bounded plan excerpts because combined reference output exceeded the display limit.
FACT: Kiri Compose services are healthy at localhost:8080. Harness runner and e2e media fixture availability are being checked. No service restart/rebuild is planned.
FACT: D3 requires focus on the pressed move control after reorder, or its opposite on the same row if the pressed direction becomes unavailable. Saved-delete dialog must contain focus and restore it on Escape.
NEXT: Read current editor markup and translations, then write the isolated harness with guaranteed scoped SQL cleanup.
FACT: Verified live frontend image sha256:fceff9f88b5046bb82b71081a86d5d73a1410c1f3bf2d31bc012eafa84be84b0, matching owner prefix fceff9f8. Playwright and 30-second e2e video fixture exist.
DECISION: Small evidence-only route READY: current U6 scope/acceptance supplied by owner; sole writer; no product implementation or independent code review claimed. Existing U1–U5 decisions remain authoritative. Admin is dark-only; light-theme screenshots are outside scope.
NEXT: Capture eight named states at all requested viewports, record unsupported contrast as NOT_TESTED, and test keyboard focus using real browser key presses.
FACT: Editor DOM confirms .course-editor-panel, .admin-lesson-summary toggle, .admin-lesson-body, data-move icons, .admin-video-filename and .transcript-editor-time readouts. The guessed managed skills/ship route was absent; exact skill discovery is next (no install/fallback substitution).
NEXT: Finish material API/selector discovery and create the temporary harness outside the e2e suite.
FACT: Managed skill sources resolved at skills/design-pass and skills/ui-measurements (flat paths). Shared viewport/evidence rules supplement the owner's exact seven viewports. No new design direction is introduced; visual/UX inspection follows the current product and D1–D9.
NEXT: Write harness setup/capture/keyboard/cleanup, then execute from a temporary directory with trace/video disabled.
DECISION: Capture D1–D8 in dark theme, browser zoom 100%, text scaling via html font-size 200% only for the two named combinations (px descendants may remain unchanged). JPEG quality 45; no traces/video/session storage copied. Geometry checks use actual controls; solid-text contrast threshold 4.5:1 (conservative for large text), unsupported paint remains NOT_TESTED.
NEXT: Execute the authored harness; append each fixture/capture/keyboard/cleanup result from the harness itself.
FACT: Temporary harness directory created under /private/tmp; existing frontend/node_modules is linked for the installed Playwright runner. Reference measure() adapted for editor targets and displayed course panels.
NEXT: Complete state transitions, fixture API setup and teardown before running the harness.
FACT: Authored temporary harness with afterAll SQL cleanup and leftover assertion before execution; screenshots JPEG 45, trace/video off, one run-scoped draft fixture. Keyboard paths and sticky rect records are separate from static geometry.
NEXT: Run the standalone harness; resolve harness errors without changing product code.
FACT: First runner attempt stopped before test discovery: temporary directory lacked package.json type=module, so import.meta was rejected. No fixture created. Added only temporary ESM metadata.
NEXT: Retry standalone harness with the ESM temporary package.
FACT: Standalone runner started one worker; polling its bounded output and current working-file tail rather than loading raw traces. No trace recording is enabled.
NEXT: Inspect first completed state measurements and screenshot coverage, then adjust only harness selectors if needed.
FACT: Browser launch blocked by sandbox macOS MachPort bootstrap permission (before fixture setup). No course created. This is runner access, not a UI defect.
NEXT: Retry the same evidence-only Playwright command with sandbox escalation so installed Chromium can launch.
NEXT: Create run-scoped draft with two modules, four lessons, one Ready video, three transcript lines and one material.
FACT: Fixture 36272608-5bd4-4aca-b6a8-276d56f0184c: draft, modules 2, lessons 3+1, Ready e2e-640x360-30s.mp4, Ukrainian transcript 3 lines, material 1. Created through admin API and production UI upload path.
NEXT: Capture D1–D8 and keyboard at 390.
FACT: Captured D1 at 390×844 text ×1; screenshot design-pass/D1-390.jpg; measurement index 0.
FACT: Escalated runner launched Chromium successfully. Continuing fixture readiness/capture; no runtime permissions or service configuration changed.
NEXT: Check actual state count and any failure from the first viewport before collecting the rest.
FACT: Captured D2 at 390×844 text ×1; screenshot design-pass/D2-390.jpg; measurement index 1.
FACT: Captured D3 at 390×844 text ×1; screenshot design-pass/D3-390.jpg; measurement index 2.
FACT: Captured D4 at 390×844 text ×1; screenshot design-pass/D4-390.jpg; measurement index 3.
FACT: Captured D5 at 390×844 text ×1; screenshot design-pass/D5-390.jpg; measurement index 4.
FACT: Captured D6 at 390×844 text ×1; screenshot design-pass/D6-390.jpg; measurement index 5.
FACT: Captured D7 at 390×844 text ×1; screenshot design-pass/D7-390.jpg; measurement index 6.
FACT: Keyboard toggle and lesson-2 reorder observed at 390; detailed key/focus outcomes in measurements.json.
FACT: First fixture reached Ready through UI upload; D1/390 captured. Checking the first measurements and next selector wait before further claims.
NEXT: Inspect any harness assertion failure and preserve cleanup evidence.
FACT: D1–D7/390 measured. Preliminary horizontal-overflow failures present in D1 and D3–D6; raw rectangles retained, not yet attributed. D2/D7 geometry shows one displayed course panel.
NEXT: Inspect D3/390 screenshot and exact overflow elements, then distinguish product findings from harness state errors.
FACT: SQL cleanup grant_audit → course_grants → payment_events/orders → courses completed for 36272608-5bd4-4aca-b6a8-276d56f0184c; leftover course count 0. Cascading modules/lessons/media/transcript/documents removed by FK.
FACT: Visual inspection D3/390 exposed a harness-state gap: default language is unknown embedded audio, while Ukrainian transcript panel remained mounted but hidden. DOM line count alone did not prove visible transcript rows. Also preliminary outside failures were off-left accessibility selects (document scrollWidth 390), not right-edge overflow.
DECISION: Correct harness to select Ukrainian language and assert visible transcript rows/readouts before D3. Restrict requested overflow element check to past-right-edge (retain document scrollWidth); do not report accessibility select positioning as product overflow.
NEXT: Fix these harness-only checks and the exact summary-link wording, then rerun with full cleanup.
FACT: First UI run stopped at a harness summary locator using element tags; React Aria renders a role=link without href. afterAll SQL cleanup ran; its count remains recorded in measurements/log. No UI defect inferred from this selector failure.
FACT: Corrected to .ui-error-summary getByRole(link), Ukrainian tab selection, and :visible rows/readouts; adjusted right-edge measurement. Earlier temporary screenshots will be superseded by the complete final run, not counted as D3 proof.
NEXT: Rerun final-state harness across all seven viewport combinations.
NEXT: Create run-scoped draft with two modules, four lessons, one Ready video, three transcript lines and one material.
FACT: Corrected harness is running; checking bounded state/keyboard counts and measurement exceptions. Prior course cleanup was confirmed 0 before retry.
NEXT: Inspect final-run screenshots, especially narrow/200% transcript and sticky bar, after capture completes.
FACT: Fixture e8e36e08-1e40-44af-93a6-3d5fc7f4a2b9: draft, modules 2, lessons 3+1, Ready e2e-640x360-30s.mp4, Ukrainian transcript 3 lines, material 1. Created through admin API and production UI upload path.
NEXT: Capture D1–D8 and keyboard at 390.
FACT: Captured D1 at 390×844 text ×1; screenshot design-pass/D1-390.jpg; measurement index 0.
FACT: Captured D2 at 390×844 text ×1; screenshot design-pass/D2-390.jpg; measurement index 1.
FACT: Read current CSS for lesson summary links (min-height 44), transcript container-query layout and course panel hiding. CSS intent is not counted as browser PASS; live measurements remain authoritative.
DECISION: Visual read: existing dark admin system for course authors; closed palette/typography/components in frontend/src/app/styles/main.css and presentation/ui. Dials variance 2, motion 1, density 7 reflect this task-focused evidence pass; no aesthetic replacement proposed.
NEXT: Finish the runtime collection and inspect the captured state contact sheets plus full-size exceptions.
FACT: Monitoring the second fixture's Ready/state progression without copying request bodies or session data. No product code has changed.
NEXT: If readiness/state assertion fails, retain exact failure and complete independent states rather than treating runner success as UI proof.
FACT: SQL cleanup grant_audit → course_grants → payment_events/orders → courses completed for e8e36e08-1e40-44af-93a6-3d5fc7f4a2b9; leftover course count 0. Cascading modules/lessons/media/transcript/documents removed by FK.
FACT: Second run reached D2/390. Reading transcript load/render contract to resolve the visible-readout readiness wait precisely.
NEXT: Inspect the runner's bounded result and any transcript selector mismatch; keep cleanup guaranteed.
FACT: Second run timed out performing pointer click on the visible Ukrainian language tab at 390; it did not reach D3. Scoped cleanup completed with leftover course count 0. Keeping this as an unresolved pointer-action observation, not asserting cause.
DECISION: Use keyboard focus+Enter to select the Ukrainian tab and record the actual selected state; this remains a real supported UI path and allows independent evidence collection. Add K-move-controls measurement/screenshot for lesson-2 up/down icons.
NEXT: Run full capture again; if pointer behavior repeats independently it becomes an owner finding with state evidence.
NEXT: Create run-scoped draft with two modules, four lessons, one Ready video, three transcript lines and one material.
FACT: Fixture c1676c6d-3aab-4dac-85da-55167c70ed99: draft, modules 2, lessons 3+1, Ready e2e-640x360-30s.mp4, Ukrainian transcript 3 lines, material 1. Created through admin API and production UI upload path.
NEXT: Capture D1–D8 and keyboard at 390.
FACT: Captured D1 at 390×844 text ×1; screenshot design-pass/D1-390.jpg; measurement index 0.
FACT: Captured D2 at 390×844 text ×1; screenshot design-pass/D2-390.jpg; measurement index 1.
FACT: Third run is active; using bounded progress counts. All previously created courses have recorded SQL leftover=0.
NEXT: Verify visible Ukrainian rows and all required state coverage before exporting the evidence folder.
FACT: Third run again pauses after D2; created a read-only diagnostic browser script against the same run-scoped fixture to observe tab focus/selection and visible rows without adding data.
NEXT: Run diagnostic and compare pointer versus keyboard language-tab behavior.
FACT: Diagnostic and main harness outputs polled; collecting exact UI/runner failure rather than assuming a language-tab defect.
NEXT: Complete accessible state collection once the stall is isolated; preserve any unavailable check as NOT_TESTED.
FACT: Independent browser focused Ukrainian tab successfully, then keyboard.press(Enter) remained pending, matching the earlier pointer-action stall. API /healthz returned 200. Root cause is unconfirmed; both actions block D3 visible-transcript proof on this fixture.
DECISION: Bound future state actions with explicit timeouts, capture screenshot through CDP if renderer action stalls, and continue independent states in fresh contexts. Do not alter product code or fake transcript visibility.
NEXT: Stop only these two task-owned runner processes, verify cleanup, and make a bounded harness that retains the finding while completing remaining states.
FACT: Sandboxed process listing was denied by macOS; requested scoped process inspection through escalation. This does not change runtime configuration.
NEXT: Signal only exact task-owned process IDs and run scoped SQL cleanup if afterAll cannot finish.
FACT: Identified exact task-owned runner PIDs 15383 (harness) and 15632 (diagnostic); sent interrupt/terminate only to those, preserving services and other work.
NEXT: Verify afterAll cleanup or apply the same course-scoped SQL manually, then complete independent state capture with bounded actions.
FACT: SQL cleanup grant_audit → course_grants → payment_events/orders → courses completed for c1676c6d-3aab-4dac-85da-55167c70ed99; leftover course count 0. Cascading modules/lessons/media/transcript/documents removed by FK.
FACT: Reading transcript field integration for a precise finding mechanism; this remains read-only and does not authorize product fixes.
NEXT: Bound any renderer hang and preserve available D3-failure evidence rather than repeating unbounded actions.
FACT: Interrupted run's afterAll completed successfully; c1676c6d-3aab-4dac-85da-55167c70ed99 leftover count 0.
DECISION: Use a valid Ukrainian-labeled embedded audio track (admin API PATCH) so the three-line Ukrainian transcript is the initial language tab. This changes only fixture metadata, within owner fixture scope; the earlier language-switch stall remains an unresolved finding, never silently accepted.
NEXT: Collect all independent states first and bound D3 operations; use the default populated language path if it renders normally.
FACT: Harness now labels the embedded track Ukrainian through the existing admin API after Ready. The transcript language is unchanged; no interception, mocked response, DOM injection or product CSS modification used.
NEXT: Run the default-populated-language fixture and inspect D3 before proceeding to all viewports.
NEXT: Create run-scoped draft with two modules, four lessons, one Ready video, three transcript lines and one material.
FACT: Fixture 018f1f11-eca7-4f31-a460-a18d8a55e0b9: draft, modules 2, lessons 3+1, Ready e2e-640x360-30s.mp4, Ukrainian transcript 3 lines, material 1. Created through admin API and production UI upload path.
NEXT: Capture D1–D8 and keyboard at 390.
FACT: Captured D1 at 390×844 text ×1; screenshot design-pass/D1-390.jpg; measurement index 0.
FACT: Captured D2 at 390×844 text ×1; screenshot design-pass/D2-390.jpg; measurement index 1.
FACT: Default-populated-language harness started; monitoring progress before accepting its fixture strategy as usable evidence.
NEXT: Inspect D3 actual transcript rows/readouts if rendered, then finish responsive and keyboard measurements.
FACT: SQL cleanup grant_audit → course_grants → payment_events/orders → courses completed for 018f1f11-eca7-4f31-a460-a18d8a55e0b9; leftover course count 0. Cascading modules/lessons/media/transcript/documents removed by FK.
FACT: Ukrainian-default fixture also stalled while expanding lesson 1 after D2, so changing default metadata did not unblock the populated transcript. Interrupted only exact task-owned harness; cleanup remains mandatory.
DECISION: Final independent-state run returns to original embedded unknown-language fixture. D3 screenshot/geometry will explicitly be PARTIAL (video filename + language tabs); populated transcript/readout/remove-line checks remain NOT_TESTED and D3 requested-state reachability FAIL. Continue D4–D8 and keyboard without switching to the stalled populated tab.
NEXT: Complete all reachable state×viewport evidence; separately capture bounded language-switch failure after independent checks.
FACT: Exact-name interrupt pattern matched no process; checking whether the previous harness already ended before launching a replacement. No unrelated process was signaled.
NEXT: Confirm terminal result and leftover count, then rerun reachable states.
FACT: Previous runner had already failed on expanding the Ukrainian-default lesson (15 s pointer action timeout); it was not interrupted by the unmatched pattern. afterAll SQL cleanup result is recorded above. Reachable-state run will not claim a populated transcript PASS.
NEXT: Execute reachable D1–D8 collection at seven viewport combinations with D3 explicitly partial.
NEXT: Create run-scoped draft with two modules, four lessons, one Ready video, three transcript lines and one material.
FACT: Fixture 9eccca68-daa7-4737-a2db-191b4bbdb232: draft, modules 2, lessons 3+1, Ready e2e-640x360-30s.mp4, Ukrainian transcript 3 lines, material 1. Created through admin API and production UI upload path.
NEXT: Capture D1–D8 and keyboard at 390.
FACT: Captured D1 at 390×844 text ×1; screenshot design-pass/D1-390.jpg; measurement index 0.
FACT: Captured D2 at 390×844 text ×1; screenshot design-pass/D2-390.jpg; measurement index 1.
FACT: Captured D3 at 390×844 text ×1; screenshot design-pass/D3-390.jpg; measurement index 3.
FACT: Captured D4 at 390×844 text ×1; screenshot design-pass/D4-390.jpg; measurement index 4.
FACT: Captured D5 at 390×844 text ×1; screenshot design-pass/D5-390.jpg; measurement index 5.
FACT: Captured D6 at 390×844 text ×1; screenshot design-pass/D6-390.jpg; measurement index 6.
FACT: Captured D7 at 390×844 text ×1; screenshot design-pass/D7-390.jpg; measurement index 7.
FACT: Captured K-move-controls at 390×844 text ×1; screenshot design-pass/K-move-controls-390.jpg; measurement index 8.
FACT: Keyboard toggle and lesson-2 reorder observed at 390; detailed key/focus outcomes in measurements.json.
FACT: Dialog Escape/focus return and summary Overview→invalid lesson checked at 390; see keyboard entries.
FACT: Captured D8 at 390×844 text ×1; screenshot design-pass/D8-390.jpg; measurement index 9.
NEXT: Capture D1–D8 and keyboard at 768.
FACT: Captured D1 at 768×1024 text ×1; screenshot design-pass/D1-768.jpg; measurement index 10.
FACT: Captured D2 at 768×1024 text ×1; screenshot design-pass/D2-768.jpg; measurement index 11.
FACT: Captured D3 at 768×1024 text ×1; screenshot design-pass/D3-768.jpg; measurement index 13.
FACT: Captured D4 at 768×1024 text ×1; screenshot design-pass/D4-768.jpg; measurement index 14.
FACT: Captured D5 at 768×1024 text ×1; screenshot design-pass/D5-768.jpg; measurement index 15.
FACT: Captured D6 at 768×1024 text ×1; screenshot design-pass/D6-768.jpg; measurement index 16.
FACT: Captured D7 at 768×1024 text ×1; screenshot design-pass/D7-768.jpg; measurement index 17.
FACT: Captured K-move-controls at 768×1024 text ×1; screenshot design-pass/K-move-controls-768.jpg; measurement index 18.
FACT: Keyboard toggle and lesson-2 reorder observed at 768; detailed key/focus outcomes in measurements.json.
FACT: Dialog Escape/focus return and summary Overview→invalid lesson checked at 768; see keyboard entries.
FACT: Captured D8 at 768×1024 text ×1; screenshot design-pass/D8-768.jpg; measurement index 19.
NEXT: Capture D1–D8 and keyboard at 1024.
FACT: Reachable-state runner active. The record explicitly distinguishes requested D3 reachability from geometry measured on the partial expanded state.
NEXT: Finish keyboard/summary and remaining viewport captures; inspect supported contrast and overflow exceptions.
FACT: Captured D1 at 1024×768 text ×1; screenshot design-pass/D1-1024.jpg; measurement index 20.
FACT: Captured D2 at 1024×768 text ×1; screenshot design-pass/D2-1024.jpg; measurement index 21.
FACT: Captured D3 at 1024×768 text ×1; screenshot design-pass/D3-1024.jpg; measurement index 23.
FACT: Captured D4 at 1024×768 text ×1; screenshot design-pass/D4-1024.jpg; measurement index 24.
FACT: Captured D5 at 1024×768 text ×1; screenshot design-pass/D5-1024.jpg; measurement index 25.
FACT: Captured D6 at 1024×768 text ×1; screenshot design-pass/D6-1024.jpg; measurement index 26.
FACT: Captured D7 at 1024×768 text ×1; screenshot design-pass/D7-1024.jpg; measurement index 27.
FACT: Captured K-move-controls at 1024×768 text ×1; screenshot design-pass/K-move-controls-1024.jpg; measurement index 28.
FACT: Keyboard toggle and lesson-2 reorder observed at 1024; detailed key/focus outcomes in measurements.json.
FACT: Dialog Escape/focus return and summary Overview→invalid lesson checked at 1024; see keyboard entries.
FACT: Captured D8 at 1024×768 text ×1; screenshot design-pass/D8-1024.jpg; measurement index 29.
NEXT: Capture D1–D8 and keyboard at 1440.
FACT: Captured D1 at 1440×900 text ×1; screenshot design-pass/D1-1440.jpg; measurement index 30.
FACT: Captured D2 at 1440×900 text ×1; screenshot design-pass/D2-1440.jpg; measurement index 31.
FACT: Captured D3 at 1440×900 text ×1; screenshot design-pass/D3-1440.jpg; measurement index 33.
FACT: Captured D4 at 1440×900 text ×1; screenshot design-pass/D4-1440.jpg; measurement index 34.
FACT: Captured D5 at 1440×900 text ×1; screenshot design-pass/D5-1440.jpg; measurement index 35.
FACT: Captured D6 at 1440×900 text ×1; screenshot design-pass/D6-1440.jpg; measurement index 36.
FACT: Captured D7 at 1440×900 text ×1; screenshot design-pass/D7-1440.jpg; measurement index 37.
FACT: Captured K-move-controls at 1440×900 text ×1; screenshot design-pass/K-move-controls-1440.jpg; measurement index 38.
FACT: Keyboard toggle and lesson-2 reorder observed at 1440; detailed key/focus outcomes in measurements.json.
FACT: Dialog Escape/focus return and summary Overview→invalid lesson checked at 1440; see keyboard entries.
FACT: Captured D8 at 1440×900 text ×1; screenshot design-pass/D8-1440.jpg; measurement index 39.
NEXT: Capture D1–D8 and keyboard at 1440-text200.
FACT: First two viewport sets completed all reachable states and four keyboard scenarios; collected explicit partial-D3 availability records. Checking numeric exceptions as the remaining viewport run proceeds.
NEXT: Inspect screenshots of any measured failure and the narrow/200% layout, then create the state×check report.
FACT: Captured D1 at 1440×900 text ×2; screenshot design-pass/D1-1440-text200.jpg; measurement index 40.
FACT: Captured D2 at 1440×900 text ×2; screenshot design-pass/D2-1440-text200.jpg; measurement index 41.
FACT: Captured D3 at 1440×900 text ×2; screenshot design-pass/D3-1440-text200.jpg; measurement index 43.
FACT: Captured D4 at 1440×900 text ×2; screenshot design-pass/D4-1440-text200.jpg; measurement index 44.
FACT: Captured D5 at 1440×900 text ×2; screenshot design-pass/D5-1440-text200.jpg; measurement index 45.
FACT: Captured D6 at 1440×900 text ×2; screenshot design-pass/D6-1440-text200.jpg; measurement index 46.
FACT: Captured D7 at 1440×900 text ×2; screenshot design-pass/D7-1440-text200.jpg; measurement index 47.
FACT: Captured K-move-controls at 1440×900 text ×2; screenshot design-pass/K-move-controls-1440-text200.jpg; measurement index 48.
FACT: Keyboard toggle and lesson-2 reorder observed at 1440-text200; detailed key/focus outcomes in measurements.json.
FACT: Dialog Escape/focus return and summary Overview→invalid lesson checked at 1440-text200; see keyboard entries.
FACT: Captured D8 at 1440×900 text ×2; screenshot design-pass/D8-1440-text200.jpg; measurement index 49.
NEXT: Capture D1–D8 and keyboard at 390-text200.
FACT: Captured D1 at 390×844 text ×2; screenshot design-pass/D1-390-text200.jpg; measurement index 50.
FACT: Tabbed last form fields for D1/390-text200; focus rect bottoms versus bar top stored in measurements.json keyboard; screenshot D1-390-text200-focus.jpg.
FACT: Captured D2 at 390×844 text ×2; screenshot design-pass/D2-390-text200.jpg; measurement index 51.
FACT: Tabbed last form fields for D2/390-text200; focus rect bottoms versus bar top stored in measurements.json keyboard; screenshot D2-390-text200-focus.jpg.
FACT: Captured D3 at 390×844 text ×2; screenshot design-pass/D3-390-text200.jpg; measurement index 53.
FACT: Tabbed last form fields for D3/390-text200; focus rect bottoms versus bar top stored in measurements.json keyboard; screenshot D3-390-text200-focus.jpg.
FACT: Captured D4 at 390×844 text ×2; screenshot design-pass/D4-390-text200.jpg; measurement index 54.
FACT: Tabbed last form fields for D4/390-text200; focus rect bottoms versus bar top stored in measurements.json keyboard; screenshot D4-390-text200-focus.jpg.
FACT: Captured D5 at 390×844 text ×2; screenshot design-pass/D5-390-text200.jpg; measurement index 55.
FACT: Captured D6 at 390×844 text ×2; screenshot design-pass/D6-390-text200.jpg; measurement index 56.
FACT: Tabbed last form fields for D6/390-text200; focus rect bottoms versus bar top stored in measurements.json keyboard; screenshot D6-390-text200-focus.jpg.
FACT: Captured D7 at 390×844 text ×2; screenshot design-pass/D7-390-text200.jpg; measurement index 57.
FACT: Captured K-move-controls at 390×844 text ×2; screenshot design-pass/K-move-controls-390-text200.jpg; measurement index 58.
FACT: Keyboard toggle and lesson-2 reorder observed at 390-text200; detailed key/focus outcomes in measurements.json.
FACT: Dialog Escape/focus return and summary Overview→invalid lesson checked at 390-text200; see keyboard entries.
FACT: Captured D8 at 390×844 text ×2; screenshot design-pass/D8-390-text200.jpg; measurement index 59.
NEXT: Capture D1–D8 and keyboard at 320.
FACT: Captured D1 at 320×640 text ×1; screenshot design-pass/D1-320.jpg; measurement index 60.
FACT: Tabbed last form fields for D1/320; focus rect bottoms versus bar top stored in measurements.json keyboard; screenshot D1-320-focus.jpg.
FACT: Captured D2 at 320×640 text ×1; screenshot design-pass/D2-320.jpg; measurement index 61.
FACT: Tabbed last form fields for D2/320; focus rect bottoms versus bar top stored in measurements.json keyboard; screenshot D2-320-focus.jpg.
FACT: Captured D3 at 320×640 text ×1; screenshot design-pass/D3-320.jpg; measurement index 63.
FACT: Tabbed last form fields for D3/320; focus rect bottoms versus bar top stored in measurements.json keyboard; screenshot D3-320-focus.jpg.
FACT: Captured D4 at 320×640 text ×1; screenshot design-pass/D4-320.jpg; measurement index 64.
FACT: Tabbed last form fields for D4/320; focus rect bottoms versus bar top stored in measurements.json keyboard; screenshot D4-320-focus.jpg.
FACT: Captured D5 at 320×640 text ×1; screenshot design-pass/D5-320.jpg; measurement index 65.
FACT: Captured D6 at 320×640 text ×1; screenshot design-pass/D6-320.jpg; measurement index 66.
FACT: Tabbed last form fields for D6/320; focus rect bottoms versus bar top stored in measurements.json keyboard; screenshot D6-320-focus.jpg.
FACT: Captured D7 at 320×640 text ×1; screenshot design-pass/D7-320.jpg; measurement index 67.
FACT: Captured K-move-controls at 320×640 text ×1; screenshot design-pass/K-move-controls-320.jpg; measurement index 68.
FACT: Keyboard toggle and lesson-2 reorder observed at 320; detailed key/focus outcomes in measurements.json.
FACT: Dialog Escape/focus return and summary Overview→invalid lesson checked at 320; see keyboard entries.
FACT: Captured D8 at 320×640 text ×1; screenshot design-pass/D8-320.jpg; measurement index 69.
FACT: SQL cleanup grant_audit → course_grants → payment_events/orders → courses completed for 9eccca68-daa7-4737-a2db-191b4bbdb232; leftover course count 0. Cascading modules/lessons/media/transcript/documents removed by FK.
FACT: Prepared bounded, read-only reproduction on the final run's existing course: 5 s action/evaluation and CDP screenshot bounds, no new fixture, no intercepted responses.
NEXT: Capture the language-switch failure before final-run afterAll removes the fixture; retain only screenshot/measurement metadata.
FACT: Reachable-state harness exited 0 (1 test passed, 41.9 s). This is runner completion only: requested populated D3 remains FAIL/NOT_TESTED. All seven viewport sets were collected and cleanup was 0.
FACT: Separate reproduction raced with afterAll cleanup; its course may already be removed. Do not treat that separate attempt as UI evidence.
NEXT: Inspect completed numeric results and screenshots, then retain a bounded reproduction inside the fixture lifetime if needed.
FACT: Numeric findings: D1/1440-text200 document scrollWidth 1469 > 1440; D1/390-text200 has past-right-edge elements despite scrollWidth 390. Sticky last-field checks FAIL in D1 at 390-text200 and 320. Other measured states have no new control below 44 px and no supported contrast ratio below 4.5.
FACT: Temporary evidence currently exceeds 8 MB; checking whether JSON repetition or JPEG sizes dominate before export. No raw traces/video recorded.
NEXT: Inspect exact D1 overflow/focus rectangles and screenshots; compact redundant measurement representation without losing values and reduce JPEG capture size if necessary.
FACT: D1/390-text200 outside entries are accessibility-hidden select descendants of a clipped ancestor, not visible overflow. Updated visibility filter to exclude ancestor clip/clip-path; actual D1/1440-text200 overflow remains a finding (three side-panel cards to x=1468.92).
FACT: Sticky evidence shows summary textarea bottom 686.81 > bar top 583.44 at 390-text200 and description bottom 621.38 > bar top 533.41 at 320. Both overlap the sticky bar; owner finding, no product fix.
DECISION: Final capture quality 20 (≤50); dialog and K-move screenshots viewport-only, other D states full-page. Compact JSON preserves all recorded values. Add bounded language reproduction inside fixture lifetime, after independent states, to avoid cleanup race.
NEXT: Run the final compact evidence pass, inspect all screenshots, and report findings without claiming Ux17 green.
NEXT: Create run-scoped draft with two modules, four lessons, one Ready video, three transcript lines and one material.
FACT: Fixture 7bdd7bd9-6248-45f3-846b-94c23797ef82: draft, modules 2, lessons 3+1, Ready e2e-640x360-30s.mp4, Ukrainian transcript 3 lines, material 1. Created through admin API and production UI upload path.
NEXT: Capture D1–D8 and keyboard at 390.
FACT: Captured D1 at 390×844 text ×1; screenshot design-pass/D1-390.jpg; measurement index 0.
FACT: Captured D2 at 390×844 text ×1; screenshot design-pass/D2-390.jpg; measurement index 1.
FACT: Captured D3 at 390×844 text ×1; screenshot design-pass/D3-390.jpg; measurement index 3.
FACT: Captured D4 at 390×844 text ×1; screenshot design-pass/D4-390.jpg; measurement index 4.
FACT: Captured D5 at 390×844 text ×1; screenshot design-pass/D5-390.jpg; measurement index 5.
FACT: Captured D6 at 390×844 text ×1; screenshot design-pass/D6-390.jpg; measurement index 6.
FACT: Captured D7 at 390×844 text ×1; screenshot design-pass/D7-390.jpg; measurement index 7.
FACT: Opened full-size D1/390-text200 focus and D1/1440-text200 screenshot for visual confirmation. Numeric rectangles identify the specific obscured fields; a terminal focus screenshot alone does not prove each earlier Tab stop.
NEXT: Add per-failing-Tab screenshot capture so the overlap finding is visible at the exact failing focus state, not only in measured steps.
FACT: Captured K-move-controls at 390×844 text ×1; screenshot design-pass/K-move-controls-390.jpg; measurement index 8.
FACT: Keyboard toggle and lesson-2 reorder observed at 390; detailed key/focus outcomes in measurements.json.
FACT: Dialog Escape/focus return and summary Overview→invalid lesson checked at 390; see keyboard entries.
FACT: Captured D8 at 390×844 text ×1; screenshot design-pass/D8-390.jpg; measurement index 9.
NEXT: Capture D1–D8 and keyboard at 768.
FACT: Captured D1 at 768×1024 text ×1; screenshot design-pass/D1-768.jpg; measurement index 10.
FACT: Captured D2 at 768×1024 text ×1; screenshot design-pass/D2-768.jpg; measurement index 11.
FACT: Captured D3 at 768×1024 text ×1; screenshot design-pass/D3-768.jpg; measurement index 13.
FACT: Captured D4 at 768×1024 text ×1; screenshot design-pass/D4-768.jpg; measurement index 14.
FACT: Captured D5 at 768×1024 text ×1; screenshot design-pass/D5-768.jpg; measurement index 15.
FACT: Captured D6 at 768×1024 text ×1; screenshot design-pass/D6-768.jpg; measurement index 16.
FACT: Captured D7 at 768×1024 text ×1; screenshot design-pass/D7-768.jpg; measurement index 17.
FACT: Captured K-move-controls at 768×1024 text ×1; screenshot design-pass/K-move-controls-768.jpg; measurement index 18.
FACT: Keyboard toggle and lesson-2 reorder observed at 768; detailed key/focus outcomes in measurements.json.
FACT: Dialog Escape/focus return and summary Overview→invalid lesson checked at 768; see keyboard entries.
FACT: Captured D8 at 768×1024 text ×1; screenshot design-pass/D8-768.jpg; measurement index 19.
NEXT: Capture D1–D8 and keyboard at 1024.
FACT: Visual inspection confirms the D1/1440-text200 side panel is clipped at the right edge. Existing focus screenshot was taken at final Save focus, so exact overlap screenshot will be added separately.
NEXT: Preserve exact obscured-field screenshots and inspect compact final evidence size.
FACT: Captured D1 at 1024×768 text ×1; screenshot design-pass/D1-1024.jpg; measurement index 20.
FACT: Captured D2 at 1024×768 text ×1; screenshot design-pass/D2-1024.jpg; measurement index 21.
FACT: Captured D3 at 1024×768 text ×1; screenshot design-pass/D3-1024.jpg; measurement index 23.
FACT: Captured D4 at 1024×768 text ×1; screenshot design-pass/D4-1024.jpg; measurement index 24.
FACT: Captured D5 at 1024×768 text ×1; screenshot design-pass/D5-1024.jpg; measurement index 25.
FACT: Captured D6 at 1024×768 text ×1; screenshot design-pass/D6-1024.jpg; measurement index 26.
FACT: Captured D7 at 1024×768 text ×1; screenshot design-pass/D7-1024.jpg; measurement index 27.
FACT: Captured K-move-controls at 1024×768 text ×1; screenshot design-pass/K-move-controls-1024.jpg; measurement index 28.
FACT: Keyboard toggle and lesson-2 reorder observed at 1024; detailed key/focus outcomes in measurements.json.
FACT: Dialog Escape/focus return and summary Overview→invalid lesson checked at 1024; see keyboard entries.
FACT: Captured D8 at 1024×768 text ×1; screenshot design-pass/D8-1024.jpg; measurement index 29.
NEXT: Capture D1–D8 and keyboard at 1440.
FACT: Captured D1 at 1440×900 text ×1; screenshot design-pass/D1-1440.jpg; measurement index 30.
FACT: Captured D2 at 1440×900 text ×1; screenshot design-pass/D2-1440.jpg; measurement index 31.
FACT: Captured D3 at 1440×900 text ×1; screenshot design-pass/D3-1440.jpg; measurement index 33.
FACT: Captured D4 at 1440×900 text ×1; screenshot design-pass/D4-1440.jpg; measurement index 34.
FACT: Captured D5 at 1440×900 text ×1; screenshot design-pass/D5-1440.jpg; measurement index 35.
FACT: Captured D6 at 1440×900 text ×1; screenshot design-pass/D6-1440.jpg; measurement index 36.
FACT: Captured D7 at 1440×900 text ×1; screenshot design-pass/D7-1440.jpg; measurement index 37.
FACT: Captured K-move-controls at 1440×900 text ×1; screenshot design-pass/K-move-controls-1440.jpg; measurement index 38.
FACT: Keyboard toggle and lesson-2 reorder observed at 1440; detailed key/focus outcomes in measurements.json.
FACT: Dialog Escape/focus return and summary Overview→invalid lesson checked at 1440; see keyboard entries.
FACT: Captured D8 at 1440×900 text ×1; screenshot design-pass/D8-1440.jpg; measurement index 39.
NEXT: Capture D1–D8 and keyboard at 1440-text200.
FACT: Captured D1 at 1440×900 text ×2; screenshot design-pass/D1-1440-text200.jpg; measurement index 40.
FACT: Captured D2 at 1440×900 text ×2; screenshot design-pass/D2-1440-text200.jpg; measurement index 41.
FACT: Captured D3 at 1440×900 text ×2; screenshot design-pass/D3-1440-text200.jpg; measurement index 43.
FACT: Captured D4 at 1440×900 text ×2; screenshot design-pass/D4-1440-text200.jpg; measurement index 44.
FACT: Captured D5 at 1440×900 text ×2; screenshot design-pass/D5-1440-text200.jpg; measurement index 45.
FACT: Captured D6 at 1440×900 text ×2; screenshot design-pass/D6-1440-text200.jpg; measurement index 46.
FACT: Captured D7 at 1440×900 text ×2; screenshot design-pass/D7-1440-text200.jpg; measurement index 47.
FACT: Captured K-move-controls at 1440×900 text ×2; screenshot design-pass/K-move-controls-1440-text200.jpg; measurement index 48.
FACT: Keyboard toggle and lesson-2 reorder observed at 1440-text200; detailed key/focus outcomes in measurements.json.
FACT: Dialog Escape/focus return and summary Overview→invalid lesson checked at 1440-text200; see keyboard entries.
FACT: Captured D8 at 1440×900 text ×2; screenshot design-pass/D8-1440-text200.jpg; measurement index 49.
NEXT: Capture D1–D8 and keyboard at 390-text200.
FACT: Captured D1 at 390×844 text ×2; screenshot design-pass/D1-390-text200.jpg; measurement index 50.
FACT: Tabbed last form fields for D1/390-text200; focus rect bottoms versus bar top stored in measurements.json keyboard; screenshot D1-390-text200-focus.jpg.
FACT: Captured D2 at 390×844 text ×2; screenshot design-pass/D2-390-text200.jpg; measurement index 51.
FACT: Tabbed last form fields for D2/390-text200; focus rect bottoms versus bar top stored in measurements.json keyboard; screenshot D2-390-text200-focus.jpg.
FACT: Captured D3 at 390×844 text ×2; screenshot design-pass/D3-390-text200.jpg; measurement index 53.
FACT: Tabbed last form fields for D3/390-text200; focus rect bottoms versus bar top stored in measurements.json keyboard; screenshot D3-390-text200-focus.jpg.
FACT: Captured D4 at 390×844 text ×2; screenshot design-pass/D4-390-text200.jpg; measurement index 54.
FACT: Tabbed last form fields for D4/390-text200; focus rect bottoms versus bar top stored in measurements.json keyboard; screenshot D4-390-text200-focus.jpg.
FACT: Captured D5 at 390×844 text ×2; screenshot design-pass/D5-390-text200.jpg; measurement index 55.
FACT: Captured D6 at 390×844 text ×2; screenshot design-pass/D6-390-text200.jpg; measurement index 56.
FACT: Tabbed last form fields for D6/390-text200; focus rect bottoms versus bar top stored in measurements.json keyboard; screenshot D6-390-text200-focus.jpg.
FACT: Captured D7 at 390×844 text ×2; screenshot design-pass/D7-390-text200.jpg; measurement index 57.
FACT: Captured K-move-controls at 390×844 text ×2; screenshot design-pass/K-move-controls-390-text200.jpg; measurement index 58.
FACT: Keyboard toggle and lesson-2 reorder observed at 390-text200; detailed key/focus outcomes in measurements.json.
FACT: Dialog Escape/focus return and summary Overview→invalid lesson checked at 390-text200; see keyboard entries.
FACT: Captured D8 at 390×844 text ×2; screenshot design-pass/D8-390-text200.jpg; measurement index 59.
NEXT: Capture D1–D8 and keyboard at 320.
FACT: Captured D1 at 320×640 text ×1; screenshot design-pass/D1-320.jpg; measurement index 60.
FACT: Tabbed last form fields for D1/320; focus rect bottoms versus bar top stored in measurements.json keyboard; screenshot D1-320-focus.jpg.
FACT: Captured D2 at 320×640 text ×1; screenshot design-pass/D2-320.jpg; measurement index 61.
FACT: Tabbed last form fields for D2/320; focus rect bottoms versus bar top stored in measurements.json keyboard; screenshot D2-320-focus.jpg.
FACT: Captured D3 at 320×640 text ×1; screenshot design-pass/D3-320.jpg; measurement index 63.
FACT: Tabbed last form fields for D3/320; focus rect bottoms versus bar top stored in measurements.json keyboard; screenshot D3-320-focus.jpg.
FACT: Captured D4 at 320×640 text ×1; screenshot design-pass/D4-320.jpg; measurement index 64.
FACT: Tabbed last form fields for D4/320; focus rect bottoms versus bar top stored in measurements.json keyboard; screenshot D4-320-focus.jpg.
FACT: Captured D5 at 320×640 text ×1; screenshot design-pass/D5-320.jpg; measurement index 65.
FACT: Captured D6 at 320×640 text ×1; screenshot design-pass/D6-320.jpg; measurement index 66.
FACT: Tabbed last form fields for D6/320; focus rect bottoms versus bar top stored in measurements.json keyboard; screenshot D6-320-focus.jpg.
FACT: Captured D7 at 320×640 text ×1; screenshot design-pass/D7-320.jpg; measurement index 67.
FACT: Captured K-move-controls at 320×640 text ×1; screenshot design-pass/K-move-controls-320.jpg; measurement index 68.
FACT: Keyboard toggle and lesson-2 reorder observed at 320; detailed key/focus outcomes in measurements.json.
FACT: Dialog Escape/focus return and summary Overview→invalid lesson checked at 320; see keyboard entries.
FACT: Captured D8 at 320×640 text ×1; screenshot design-pass/D8-320.jpg; measurement index 69.
NEXT: Reproduce populated-language activation with 5-second bounds inside fixture lifetime.
FACT: Prepared supplemental read-only Tab capture for the same live fixture at 390-text200 and 320; no data mutation. Saved harness now includes screenshot-at-covered-step for reproducibility.
NEXT: Capture exact field overlap while final harness continues; keep supplemental metadata separate and linked in measurements/report.
FACT: Supplemental real-Tab sticky screenshots captured for summary/description/owner at 390-text200 and 320; metadata design-pass/sticky-findings.json.
FACT: Bounded populated-language reproduction FAIL; action/evaluation metadata language-tab-finding.json; before screenshot F1-before-language-390.jpg; after screenshot unavailable.
FACT: SQL cleanup grant_audit → course_grants → payment_events/orders → courses completed for 7bdd7bd9-6248-45f3-846b-94c23797ef82; leftover course count 0. Cascading modules/lessons/media/transcript/documents removed by FK.
FACT: Polled final capture and supplemental sticky jobs. Final evidence will include only run-matched compact screenshots and JSON, with partial D3 clearly labeled.
NEXT: Inspect final screenshot contact sheets and exact F1/F2 exception frames; export only after cleanup and size validation.
FACT: Final harness exited 0 (52.4 s), 63 state/control captures (56 D1–D8 + 7 K-move), 38 keyboard records plus bounded language finding. SQL cleanup for final fixture 7bdd7bd9-6248-45f3-846b-94c23797ef82 left 0 courses. Compact temporary evidence 6,396,064 bytes before harness copies.
FACT: Supplemental exact-Tab measurements reproduce overlap: summary at 390-text200 (bottom 686.8125, bar top 583.4375) and description at 320 (bottom 621.375, bar top 533.40625), with matching F2 screenshots.
NEXT: Inspect all D1–D8 contact sheets (all seven viewports), then full-size exception frames and numeric matrix.
FACT: Contact-sheet helper could not use PIL (not installed); no installation attempted. The preceding numeric count is valid; contact sheets were not yet produced/inspected.
NEXT: Build temporary review sheets with installed Playwright/local JPEGs, without accessing app data, then inspect them.
FACT: Opened D1–D4 review sheets covering all seven viewports (source JPEGs retained; tall shots shown top/middle/bottom). D3/D4 remain explicitly partial for transcript content.
NEXT: Inspect D5–D8 sheets and exact F1/F2 exception frames; finish the final evidence table.
FACT: Opened D5–D8 review sheets covering all seven viewports. Saved-delete dialogs remain within viewport, materials form is revealed, Access is the single displayed course panel, and admin 404 keeps admin navigation. D1–D8 visual inspection completed on captured reachable states; no populated-transcript visual PASS claimed.
NEXT: Inspect F1/F2 exact frames and summarize targets, contrast support and state coverage in the working-file matrix.
FACT: Inspected full-size F1 pre-language and F2 obscured-field frames. F2 confirms sticky bar visually covers the focused textarea in both measured cases. F1 retains the focused Ukrainian tab immediately before the bounded stall; post-action screenshot was unavailable within 5 s (not fabricated).
NEXT: Export compact evidence and harness, append state×check/keyboard tables, and verify all six run-course IDs have zero leftovers.
FACT: Final numeric scope: 574 rendered target observations, minimum 44×44 CSS px; 1064 supported contrast observations PASS, minimum 4.509848657386538:1; 16 unsupported/empty contrast observations NOT_TESTED. Horizontal text-box overflow count 0 for selected rendered new texts. Vertical scroll/client differences need classification before any clipping PASS.
NEXT: Classify vertical clipping signals and contrast exclusions; export tables conservatively (unavailable transcript checks stay NOT_TESTED).
FACT: Final Kiri SQL audit across all six created course IDs returned courses=0, course_grants=0, orders=0. Every attempt's afterAll also recorded leftover course count 0. No fixture remains.
NEXT: Copy final harness/JPEG/JSON artifacts into design-pass, append state×check matrix and ranked findings, then verify branch/code scope and folder size.

FACT: FINAL EVIDENCE REPORT — 2026-11-18, admin-editor-ux @ a09d960f; Kiri frontend fceff9f8. Evidence-only scope; no product code changes. Ux17 gate FAIL / verification incomplete. Runner completion (exit 0, 52.4 s) does not override the findings below.
FACT: Evidence folder [design-pass](design-pass/); [measurements.json](design-pass/measurements.json), [harness.spec.ts.txt](design-pass/harness.spec.ts.txt), [harness.config.ts.txt](design-pass/harness.config.ts.txt). Exact final-run harness copied; separate real-Tab frame supplement is [sticky.supplement.mjs.txt](design-pass/sticky.supplement.mjs.txt). No raw traces, cookies, auth-state or media bytes included.
FACT: Viewport key: 390=390×844; 768=768×1024; 1024=1024×768; 1440=1440×900; 1440-text200/390-text200 use html font-size 200%, browser zoom 100%; 320=320×640. Admin dark-only. CSS text scaling does not prove enlargement of px-sized descendants.
FACT: Findings (owner action; no fixes applied):

| ID | Result / scope | Measured evidence | Owner next action |
|---|---|---|---|
| F1 | FAIL — populated transcript path at 390×844 stalls on Ukrainian tab activation, reproduced by pointer and keyboard; Ukrainian-default expansion also timed out. D3 is partial at all viewports; readouts/remove-line NOT_TESTED. Root cause unconfirmed. | [pre-action screenshot](design-pass/F1-before-language-390.jpg), [bounded measurements](design-pass/language-tab-finding.json): Enter pending at 5000 ms; post-action CDP screenshot also unavailable within 5000 ms. Before state has 0 visible transcript rows, focused Ukrainian tab. | Diagnose populated-language render/activation, then rerun full D3 rows/readouts/remove-line checks at all seven viewports. |
| F2 | FAIL — sticky Save bar covers focused Overview fields during actual Tab navigation. | [390/200% summary](design-pass/F2-390-text200-0.jpg): bottom 686.8125, bar top 583.4375 (overlap 103.375 px). [320 description](design-pass/F2-320-1.jpg): bottom 621.375, bar top 533.40625 (overlap 87.96875 px). [step metadata](design-pass/sticky-findings.json). | Correct scroll/focus accommodation, rerun last-field Tab checks at both viewports. WCAG [2.4.11](https://www.w3.org/WAI/WCAG22/Understanding/focus-not-obscured-minimum.html); project Ux11 requires no covering. |
| F3 | FAIL — Overview side cards overflow at 1440×900 / 200% text. | [D1-1440-text200.jpg](design-pass/D1-1440-text200.jpg), measurements D1/1440-text200: document scrollWidth=1469 vs clientWidth=1440; status/summary/delete side cards right=1468.921875 (28.921875 px past right). | Correct Overview grid/reflow; remeasure same state. [WCAG reflow](https://www.w3.org/WAI/WCAG22/Understanding/reflow.html), project Ux17. |

FACT: State × viewport × check table. Each row links its screenshot; numeric arrays in measurements.json are selected by state + viewportName. PASS refers only to observed/rendered targets and texts. D3-partial is never a transcript PASS. NOT_TESTED means unavailable/not applicable to this state, with scope explained below.

| State | Viewport | Requested state | Right overflow | New targets ≥44 | New-text box/clipping | Course panels | Sticky last fields | Solid-text contrast | Evidence |
|---|---|---|---|---|---|---|---|---|---|
| D1 | 390 | PASS | PASS | PASS | PASS | PASS (1) | NOT_TESTED | PASS | [D1-390.jpg](design-pass/D1-390.jpg) |
| D2 | 390 | PASS | PASS | PASS | PASS | PASS (1) | NOT_TESTED | PASS | [D2-390.jpg](design-pass/D2-390.jpg) |
| D3-partial | 390 | FAIL (F1) | PASS | PASS rendered; remove-line NOT_TESTED | NOT_TESTED (readouts); PASS rendered | PASS (1) | NOT_TESTED | PASS | [D3-390.jpg](design-pass/D3-390.jpg) |
| D4 | 390 | PASS | PASS | PASS | PASS | PASS (1) | NOT_TESTED | PASS | [D4-390.jpg](design-pass/D4-390.jpg) |
| D5 | 390 | PASS | PASS | PASS | PASS | PASS (1) | NOT_TESTED | NOT_TESTED (danger); PASS solid | [D5-390.jpg](design-pass/D5-390.jpg) |
| D6 | 390 | PASS | PASS | PASS | PASS | PASS (1) | NOT_TESTED | PASS | [D6-390.jpg](design-pass/D6-390.jpg) |
| D7 | 390 | PASS | PASS | NOT_TESTED (none) | PASS | PASS (1) | NOT_TESTED | PASS | [D7-390.jpg](design-pass/D7-390.jpg) |
| D8 | 390 | PASS | PASS | NOT_TESTED (none) | NOT_TESTED | NOT_TESTED (404: 0) | NOT_TESTED | PASS | [D8-390.jpg](design-pass/D8-390.jpg) |
| D1 | 768 | PASS | PASS | PASS | PASS | PASS (1) | NOT_TESTED | PASS | [D1-768.jpg](design-pass/D1-768.jpg) |
| D2 | 768 | PASS | PASS | PASS | PASS | PASS (1) | NOT_TESTED | PASS | [D2-768.jpg](design-pass/D2-768.jpg) |
| D3-partial | 768 | FAIL (F1) | PASS | PASS rendered; remove-line NOT_TESTED | NOT_TESTED (readouts); PASS rendered | PASS (1) | NOT_TESTED | PASS | [D3-768.jpg](design-pass/D3-768.jpg) |
| D4 | 768 | PASS | PASS | PASS | PASS | PASS (1) | NOT_TESTED | PASS | [D4-768.jpg](design-pass/D4-768.jpg) |
| D5 | 768 | PASS | PASS | PASS | PASS | PASS (1) | NOT_TESTED | NOT_TESTED (danger); PASS solid | [D5-768.jpg](design-pass/D5-768.jpg) |
| D6 | 768 | PASS | PASS | PASS | PASS | PASS (1) | NOT_TESTED | PASS | [D6-768.jpg](design-pass/D6-768.jpg) |
| D7 | 768 | PASS | PASS | NOT_TESTED (none) | PASS | PASS (1) | NOT_TESTED | PASS | [D7-768.jpg](design-pass/D7-768.jpg) |
| D8 | 768 | PASS | PASS | NOT_TESTED (none) | NOT_TESTED | NOT_TESTED (404: 0) | NOT_TESTED | PASS | [D8-768.jpg](design-pass/D8-768.jpg) |
| D1 | 1024 | PASS | PASS | PASS | PASS | PASS (1) | NOT_TESTED | PASS | [D1-1024.jpg](design-pass/D1-1024.jpg) |
| D2 | 1024 | PASS | PASS | PASS | PASS | PASS (1) | NOT_TESTED | PASS | [D2-1024.jpg](design-pass/D2-1024.jpg) |
| D3-partial | 1024 | FAIL (F1) | PASS | PASS rendered; remove-line NOT_TESTED | NOT_TESTED (readouts); PASS rendered | PASS (1) | NOT_TESTED | PASS | [D3-1024.jpg](design-pass/D3-1024.jpg) |
| D4 | 1024 | PASS | PASS | PASS | PASS | PASS (1) | NOT_TESTED | PASS | [D4-1024.jpg](design-pass/D4-1024.jpg) |
| D5 | 1024 | PASS | PASS | PASS | PASS | PASS (1) | NOT_TESTED | NOT_TESTED (danger); PASS solid | [D5-1024.jpg](design-pass/D5-1024.jpg) |
| D6 | 1024 | PASS | PASS | PASS | PASS | PASS (1) | NOT_TESTED | PASS | [D6-1024.jpg](design-pass/D6-1024.jpg) |
| D7 | 1024 | PASS | PASS | NOT_TESTED (none) | PASS | PASS (1) | NOT_TESTED | PASS | [D7-1024.jpg](design-pass/D7-1024.jpg) |
| D8 | 1024 | PASS | PASS | NOT_TESTED (none) | NOT_TESTED | NOT_TESTED (404: 0) | NOT_TESTED | PASS | [D8-1024.jpg](design-pass/D8-1024.jpg) |
| D1 | 1440 | PASS | PASS | PASS | PASS | PASS (1) | NOT_TESTED | PASS | [D1-1440.jpg](design-pass/D1-1440.jpg) |
| D2 | 1440 | PASS | PASS | PASS | PASS | PASS (1) | NOT_TESTED | PASS | [D2-1440.jpg](design-pass/D2-1440.jpg) |
| D3-partial | 1440 | FAIL (F1) | PASS | PASS rendered; remove-line NOT_TESTED | NOT_TESTED (readouts); PASS rendered | PASS (1) | NOT_TESTED | PASS | [D3-1440.jpg](design-pass/D3-1440.jpg) |
| D4 | 1440 | PASS | PASS | PASS | PASS | PASS (1) | NOT_TESTED | PASS | [D4-1440.jpg](design-pass/D4-1440.jpg) |
| D5 | 1440 | PASS | PASS | PASS | PASS | PASS (1) | NOT_TESTED | NOT_TESTED (danger); PASS solid | [D5-1440.jpg](design-pass/D5-1440.jpg) |
| D6 | 1440 | PASS | PASS | PASS | PASS | PASS (1) | NOT_TESTED | PASS | [D6-1440.jpg](design-pass/D6-1440.jpg) |
| D7 | 1440 | PASS | PASS | NOT_TESTED (none) | PASS | PASS (1) | NOT_TESTED | PASS | [D7-1440.jpg](design-pass/D7-1440.jpg) |
| D8 | 1440 | PASS | PASS | NOT_TESTED (none) | NOT_TESTED | NOT_TESTED (404: 0) | NOT_TESTED | PASS | [D8-1440.jpg](design-pass/D8-1440.jpg) |
| D1 | 1440-text200 | PASS | FAIL (F3) | PASS | PASS | PASS (1) | NOT_TESTED | PASS | [D1-1440-text200.jpg](design-pass/D1-1440-text200.jpg) |
| D2 | 1440-text200 | PASS | PASS | PASS | PASS | PASS (1) | NOT_TESTED | PASS | [D2-1440-text200.jpg](design-pass/D2-1440-text200.jpg) |
| D3-partial | 1440-text200 | FAIL (F1) | PASS | PASS rendered; remove-line NOT_TESTED | NOT_TESTED (readouts); PASS rendered | PASS (1) | NOT_TESTED | PASS | [D3-1440-text200.jpg](design-pass/D3-1440-text200.jpg) |
| D4 | 1440-text200 | PASS | PASS | PASS | PASS | PASS (1) | NOT_TESTED | PASS | [D4-1440-text200.jpg](design-pass/D4-1440-text200.jpg) |
| D5 | 1440-text200 | PASS | PASS | PASS | PASS | PASS (1) | NOT_TESTED | NOT_TESTED (danger); PASS solid | [D5-1440-text200.jpg](design-pass/D5-1440-text200.jpg) |
| D6 | 1440-text200 | PASS | PASS | PASS | PASS | PASS (1) | NOT_TESTED | PASS | [D6-1440-text200.jpg](design-pass/D6-1440-text200.jpg) |
| D7 | 1440-text200 | PASS | PASS | NOT_TESTED (none) | PASS | PASS (1) | NOT_TESTED | PASS | [D7-1440-text200.jpg](design-pass/D7-1440-text200.jpg) |
| D8 | 1440-text200 | PASS | PASS | NOT_TESTED (none) | NOT_TESTED | NOT_TESTED (404: 0) | NOT_TESTED | PASS | [D8-1440-text200.jpg](design-pass/D8-1440-text200.jpg) |
| D1 | 390-text200 | PASS | PASS | PASS | PASS | PASS (1) | FAIL | PASS | [D1-390-text200.jpg](design-pass/D1-390-text200.jpg) |
| D2 | 390-text200 | PASS | PASS | PASS | PASS | PASS (1) | PASS | PASS | [D2-390-text200.jpg](design-pass/D2-390-text200.jpg) |
| D3-partial | 390-text200 | FAIL (F1) | PASS | PASS rendered; remove-line NOT_TESTED | NOT_TESTED (readouts); PASS rendered | PASS (1) | PASS | PASS | [D3-390-text200.jpg](design-pass/D3-390-text200.jpg) |
| D4 | 390-text200 | PASS | PASS | PASS | PASS | PASS (1) | PASS | PASS | [D4-390-text200.jpg](design-pass/D4-390-text200.jpg) |
| D5 | 390-text200 | PASS | PASS | PASS | PASS | PASS (1) | NOT_TESTED | NOT_TESTED (danger); PASS solid | [D5-390-text200.jpg](design-pass/D5-390-text200.jpg) |
| D6 | 390-text200 | PASS | PASS | PASS | PASS | PASS (1) | PASS | PASS | [D6-390-text200.jpg](design-pass/D6-390-text200.jpg) |
| D7 | 390-text200 | PASS | PASS | NOT_TESTED (none) | PASS | PASS (1) | NOT_TESTED | PASS | [D7-390-text200.jpg](design-pass/D7-390-text200.jpg) |
| D8 | 390-text200 | PASS | PASS | NOT_TESTED (none) | NOT_TESTED | NOT_TESTED (404: 0) | NOT_TESTED | PASS | [D8-390-text200.jpg](design-pass/D8-390-text200.jpg) |
| D1 | 320 | PASS | PASS | PASS | PASS | PASS (1) | FAIL | PASS | [D1-320.jpg](design-pass/D1-320.jpg) |
| D2 | 320 | PASS | PASS | PASS | PASS | PASS (1) | PASS | PASS | [D2-320.jpg](design-pass/D2-320.jpg) |
| D3-partial | 320 | FAIL (F1) | PASS | PASS rendered; remove-line NOT_TESTED | NOT_TESTED (readouts); PASS rendered | PASS (1) | PASS | PASS | [D3-320.jpg](design-pass/D3-320.jpg) |
| D4 | 320 | PASS | PASS | PASS | PASS | PASS (1) | PASS | PASS | [D4-320.jpg](design-pass/D4-320.jpg) |
| D5 | 320 | PASS | PASS | PASS | PASS | PASS (1) | NOT_TESTED | NOT_TESTED (danger); PASS solid | [D5-320.jpg](design-pass/D5-320.jpg) |
| D6 | 320 | PASS | PASS | PASS | PASS | PASS (1) | PASS | PASS | [D6-320.jpg](design-pass/D6-320.jpg) |
| D7 | 320 | PASS | PASS | NOT_TESTED (none) | PASS | PASS (1) | NOT_TESTED | PASS | [D7-320.jpg](design-pass/D7-320.jpg) |
| D8 | 320 | PASS | PASS | NOT_TESTED (none) | NOT_TESTED | NOT_TESTED (404: 0) | NOT_TESTED | PASS | [D8-320.jpg](design-pass/D8-320.jpg) |

FACT: Measurement interpretation: 574 target observations across D1–D8 and K-move-control captures; minimum width=44, height=44 CSS px. Icon lesson-2 up/down and module moves are in K-move-controls-<viewport>.jpg plus measurements. Remove-line is NOT_TESTED because populated transcript cannot be reached. D7/D8 render none of the selected new controls; NOT_TESTED there does not negate other measured states. Control bounds are measured; wrapper padding or an inner SVG is not substituted for the target.
FACT: All selected rendered badges, filename and dirty-bar status have no measured X/Y box overflow. Unavailable transcript readouts stay NOT_TESTED. H1 has 1–5 px scrollHeight overhang with visible overflow; this is retained in raw measurements, not a finding of clipped new text. Ancestor-clipped accessibility selects are excluded from visible overflow, not mistaken for right-edge page content. Box measurements do not certify glyph ink or all possible content lengths.
FACT: Contrast: 1064 supported observations PASS (minimum 4.509848657386538:1, threshold 4.5:1); 16 NOT_TESTED = 9 empty clean-bar outputs (no text) + 7 destructive dialog confirmations on color(srgb … / 0.14) backgrounds. No invented composited ratio. Dialog solid text passes; danger confirmation paint remains NOT_TESTED. Transcript readouts remain NOT_TESTED. [WCAG 1.4.3](https://www.w3.org/WAI/WCAG22/Understanding/contrast-minimum.html).

FACT: Keyboard table (actual browser keys; focus metadata in measurements.json).

| Check | Viewports | Result | Evidence |
|---|---|---|---|
| Tab to lesson toggle; Enter/Space flip aria-expanded and retain focus | All 7 | PASS | measurements.json keyboard check=toggle-Tab-Enter-Space; K-move-controls / keyboard-summary screenshots per viewport |
| Lesson 2 up → same row opposite down control focused because up disappears | All 7 | PASS | measurements.json keyboard check=move-up-lesson2-D3; K-move-controls / keyboard-summary screenshots per viewport |
| Saved-delete dialog entry focus inside; Escape closes and returns to trigger | All 7 | PASS | measurements.json keyboard check=delete-dialog-Escape; K-move-controls / keyboard-summary screenshots per viewport |
| Overview summary link → Lessons selected + invalid lesson title focused | All 7 | PASS | measurements.json keyboard check=summary-Overview-to-invalid-lesson; K-move-controls / keyboard-summary screenshots per viewport |
| Sticky last fields (last three visible form inputs, actual Tab through end controls) | 390-text200, 320 | FAIL D1 (F2); PASS D2/D3-partial/D4/D6, eight scoped observations | measurements.json keyboard check=sticky-last-fields + F2 exact-frame supplement |
| Populated Ukrainian tab activation | 390 | FAIL; action/post-shot bounded 5 s each | language-tab-finding.json + F1-before-language-390.jpg |

FACT: Visual/UX review uses the same screenshots: stable course panel grouping, filename, collapse controls, materials disclosure, dirty feedback, dialog and admin 404 inspected. Existing dark tokens/components preserved; no extra aesthetic changes proposed. Usability gate fails on F1–F3; no owner acceptance of these limitations is inferred. Table/dashboard/auth/performance/production review and full WCAG conformance are outside this evidence task. No before/after regression claim without a base-revision run.
FACT: All six created courses removed through scoped SQL in afterAll (grant_audit → grants → payment_events/orders → courses; cascading children). Final audit: courses=0, course_grants=0, orders=0. Aggregate make/check suites were NOT_RUN: this task authorized a standalone evidence harness only; Ux18 remains with owner session. No commits, staging or branch changes.
NEXT: Owner session diagnoses F1, fixes F2/F3 if accepted, then reruns affected UI evidence including complete D3; obtain supported contrast evidence for danger confirmation. Keep Ux17 open and run Ux18 only in its separately authorized gate window.
FACT: Exported evidence and state×check matrix. D8's initial contrast selection covered its heading; adding a read-only return-link contrast supplement avoids extending that PASS to unmeasured text.
FACT: Earlier source inspection printed the whole auth.ts, including demo fixture credential literals, instead of restricting output to authStatePath. This violated the requested no-credential-output constraint. No credential values were copied to evidence artifacts; subsequent access uses only the storage-state path. No credentials are repeated here.
NEXT: Measure admin-404 return link at all viewports, then audit final artifact scope, size, trace absence and branch/HEAD.

FACT: D8 return-link supplement completed at all seven viewports; all seven link ratios PASS (see design-pass/admin404-supplement.json and measurements.json admin404Supplement). This supplies the previously missing return-link scope without changing the D8 screenshots or original heading measurements.
FACT: Final artifact audit: folder below 8,000,000 bytes, no trace/video/HAR/media artifacts, no demo credential literals in exported text/JSON. Branch admin-editor-ux and HEAD a09d960f unchanged. Tracked dirty files remain the pre-existing unrelated set; product-code diff is empty. No staging/commit/branch operation performed.
DECISION: Evidence collection complete with findings; Ux17 remains FAIL / verification incomplete, not an accepted limitation or a green gate. Ux18 NOT_RUN.
NEXT: Hand findings and evidence to owner session for diagnosis/fixes and targeted remeasurement, especially complete populated D3 and danger-button contrast.

FACT: FINAL FIX REPORT — 2026-11-18, admin-editor-ux @ a09d960f + uncommitted F1–F3 fixes. This appended table supersedes only the three findings and affected-state results above; historical table/evidence retained. Frontend image `sha256:5cf22121aa90dab573dfa2ce1ef9a0c00f68a3cc94daefbfb7b5bbff341ba050`.

| ID | Final result | Root cause and fix | Fresh proof |
|---|---|---|---|
| F1 | PASS | The transcript native size query recurses during React Aria visibility/layout traversal in the Ukrainian course-editor composition. Material, collapse and U5 scroll are not required; one Ready lesson / one populated orphan language suffices. Replaced only this query with a ResizeObserver boolean at the same 40rem threshold; unchanged thresholds do not re-render, observers disconnect. | Backend-free real-component Rstest fails before fix at the actual Ukrainian-tab click (5s), passes after. Additional observer test bounds 100 repeated callbacks. Full populated D3 at all 7 viewports: 3 rows, six readouts, checked material, Enter-remove to 2 rows. Changed existing transcript e2e PASS. |
| F2 | PASS | HTML is the scroller and the bar inset was already set there. Native tall-textarea focus scrolling accommodates the caret rather than its full rectangle; reserve measured visible-textarea height as well as bar/outline before Tab. Retain U5 observer for wrapped bar/form/text-size changes and restore both CSS variables on unmount. | Real Tab: 390/200% summary bottom 373.8125 < bar top 583.4375; 320 description bottom 305.375 < bar top 533.40625; overlap 0 in both. Exact focused screenshots below; all 7 viewport sticky traversals PASS. |
| F3 | PASS | Overview side-panel implicit auto tracks/minimum card width could exceed its allocated parent track. Added zero-minimum side track/cards and max-inline-size:100%; text wraps. | Same 1440×900 / 200% text: document scrollWidth=1440 vs clientWidth=1440 (formerly 1469). All 14 affected D1/D3 states have zero measured right overflow. |

FACT: Final affected-state matrix. Dimensions/enlargement match the original seven-state contract: 100% browser zoom; `html.style.fontSize=200%` for text200 (not actual browser zoom and not proof for px-sized descendants). All linked screenshots were actually inspected, including transcript regions and below-fold enlargement slices.

| Viewport | D1 / right overflow | Full D3: rows, readouts, material, remove | Exact Tab / sticky last fields | Evidence |
|---|---|---|---|---|
| 390 | PASS / 0px | PASS (3→2 rows) | PASS / PASS | [D1](design-pass/fix-D1-390.jpg), [D3](design-pass/fix-D3-390.jpg) |
| 768 | PASS / 0px | PASS (3→2 rows) | PASS / PASS | [D1](design-pass/fix-D1-768.jpg), [D3](design-pass/fix-D3-768.jpg) |
| 1024 | PASS / 0px | PASS (3→2 rows) | PASS / PASS | [D1](design-pass/fix-D1-1024.jpg), [D3](design-pass/fix-D3-1024.jpg) |
| 1440 | PASS / 0px | PASS (3→2 rows) | PASS / PASS | [D1](design-pass/fix-D1-1440.jpg), [D3](design-pass/fix-D3-1440.jpg) |
| 1440-text200 | PASS / 0px | PASS (3→2 rows) | PASS / PASS | [D1](design-pass/fix-D1-1440-text200.jpg), [D3](design-pass/fix-D3-1440-text200.jpg) |
| 390-text200 | PASS / 0px | PASS (3→2 rows) | PASS / PASS | [D1](design-pass/fix-D1-390-text200.jpg), [D3](design-pass/fix-D3-390-text200.jpg) |
| 320 | PASS / 0px | PASS (3→2 rows) | PASS / PASS | [D1](design-pass/fix-D1-320.jpg), [D3](design-pass/fix-D3-320.jpg) |

FACT: [Exact 390/200% focus](design-pass/fix-F2-390-text200.jpg), [exact 320 focus](design-pass/fix-F2-320.jpg), [keyboard language activation](design-pass/fix-F1-after-keyboard-390.jpg), [compact numeric/key evidence](design-pass/fix-measurements.json), [exact harness](design-pass/fix-harness.spec.ts.txt), [verification + source hashes](design-pass/fix-verification.json), [SQL cleanup audit](design-pass/fix-cleanup.json). Full JPEG quality 10; exact F2 frames quality 25; no original evidence overwritten and no raw traces/auth state/media exported.
FACT: Selected target bounds all ≥44px; no selected new-text horizontal box overflow. Supported solid-text contrast 224/224 PASS, minimum 4.509848657386538:1. Six empty clean-bar outputs have no rendered text and remain NOT_TESTED (not a text contrast failure). Glyph-ink/full-WCAG conformance is not inferred from box checks. Original D5 translucent destructive-confirmation contrast remains NOT_TESTED; this affected-state rerun does not claim it away.
FACT: Requested ordered commands: `npm run format` PASS (0); `npm run format:check` PASS (0); `npm run typecheck` PASS (0); `npm run lint` PASS (0); `npx rstest run src --testTimeout 30000` PASS (0), 117 files / 1483 tests, 0 skipped. Earlier typecheck failed on a missing fixture isLoading flag, fixed before this full passing sequence. Changed real-video transcript e2e: 1 PASS (8.9s); final affected-state harness: 1 PASS (26.4s), no pageerrors. Full ordered `make check` / full e2e / Ux18 NOT_RUN here.
FACT: Scoped visual/UX final read: established NorthWind dark tokens, type, controls and motion preserved. No new blocker/major in F1–F3 and their affected states; focus visibility, 320 reflow, 200% text and selected target/contrast checks measured. This is a scoped correction pass, not an independent review or full interface/WCAG certification.
FACT: Final SQL audit of all 31 recorded run-created course IDs: courses=0, course_grants=0, orders=0. No commits, staging or branch changes. Scope F1–F3 complete; broader Ux17 unsupported D5 paint and separate Ux18 gate remain with the owner session.
NEXT: Owner session can consume this source/evidence snapshot and run the separately authorized Ux18 full ordered gate; do not reopen F1–F3 without new evidence.

FACT: FINAL OWNER-REVISION TABLE — 2026-11-18, Courses 1 decision. This table supersedes the nested Chromium-in-rstest proof and 117-file/1483-test count above; historical tables remain intact. Current durable proof: [owner verification](design-pass/fix-owner-verification.json).

| Finding | Result | Current proof |
|---|---|---|
| F1 renderer hang | PASS | Existing Compose real-video transcript spec now reproduces the reported 390×844 state: default embedded und track, inactive populated Ukrainian tab, two rows, one material; asserts activation, rows/text and collapse/reopen. Pre-fix native `@container transcript-editor` + markup: frontend rebuild 0, image `084a0d649211c1724cf1fcdeba50bbf2adc618cf19289784f170c644d8e5e466`, test exit **1**, performing Ukrainian-tab click times out at 5s (`transcript.spec.ts:384:21` in red run, current line 386). Exact fixed sources restored: rebuild 0, image `90763e230b2e7c01e66d78810cf3e0576f4c2acd0aad4ca372ec623aaf715a99`, same narrow case exit **0** (7.8s), including existing student WebVTT/keyboard assertions. Default 1280×720 pre-fix experiment passed; narrow viewport matters for this short fixture. No loader error counted. |
| F2 sticky bar covers focus | PASS, unchanged | Previous seven-viewport real-Tab/sticky evidence remains valid: 390/200% focused bottom 373.8125 < bar top 583.4375; 320 focused bottom 305.375 < bar top 533.40625. F2 source unchanged in this revision. |
| F3 Overview overflow | PASS, unchanged | Previous 1440×900/200% measurement remains scrollWidth=clientWidth=1440 (formerly 1469); all 14 affected D1/D3 states have zero right overflow. F3 source unchanged in this revision. |

FACT: Deleted `frontend/scripts/test-course-editor-layout.mjs`, `frontend/src/modules/admin/ui/CourseEditorView.layout.test.ts`, and `frontend/e2e/fixtures/course-editor-layout.tsx` plus its empty directory. No Chromium runner remains in Rstest. Kept controlled-observer unit: exact 40rem/rem thresholds, ≤1 commit over 100 callbacks, disconnect on unmount. Compose browser proof uses Kiri environment, explicit `COMPOSE_FILE=docker-compose.yml:docker-compose.e2e.yml`, cwd frontend, 5s actions/10s navigation, no traces. Only frontend was rebuilt.
FACT: Requested verification rerun: `npm run format` **0**; `npm run format:check` **0**; `npm run typecheck` **0**; `npm run lint` **0**; `npx rstest run src --testTimeout 30000` **0** — **116 files, 1482 tests, 0 failed, 0 skipped**. Formatter only removed an extra e2e blank line after green; exact restored product hashes still match. All 34 recorded session-created course IDs audited: courses=0, grants=0, orders=0, including all three revision runs. Evidence folder 7,934,804 bytes, no new JPEGs or raw traces. Current source hashes and both image IDs/exits are in the owner verification JSON.
FACT: Previous seven-viewport visual evidence retained: product fix restored exactly; TranscriptEditor changed only its proof-reference comment since that visual snapshot. Full ordered make check / full e2e / Ux18 were not run in this revision; earlier unsupported D5 translucent-paint contrast remains NOT_TESTED. No commits, staging, or branch change.
NEXT: Owner session consumes the updated Compose proof and runs the separately scoped full ordered gate when authorized.
