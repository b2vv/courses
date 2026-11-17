# MS3a review, wave 2 — independent, read-only (delta after w1)

Same rules and authority as wave 1 (`docs/multi-subtitles/_wip-ms3a-review-codex-w1-prompt.md` — read it first; read-only,
ignore unrelated dirty files). Wave 1 (`docs/multi-subtitles/_wip-ms3a-review-codex-w1.md`) gave BLOCK with two Majors in
`frontend/src/modules/admin/ui/LessonLanguageTabs.tsx`:
1. a failed background tracks/inventory refresh unmounted every editor (unsaved lines lost);
2. a track renamed/deleted while the replace-confirmation dialog was open still got `PUT language=<new>`.
Courses 1 fixed both in that file (the alert now sits above the tabs when cached lists exist; `trackMatches` is re-checked
after confirmation, keeping the confirmed token) and added tests in `LessonLanguageTabs.test.tsx` ("a track renamed while
the replace dialog is open sends zero PUT", "a failed background refresh keeps the editors…"). The first was
mutation-checked: removing the re-check makes it fail.

Task: verify both fixes are correct and complete (other paths with the same shape: any other await between the last
verification and PUT? the orphan path? the inventory token after confirmation?), and check the whole MS3a diff once more for
anything wave 1 did not report. Output as in wave 1: findings with severity, file:line, failure scenario, plan row; then
SHIP or BLOCK.
