**SHIP — `ms3c-w3` code review.** All three package hashes matched before and after review. The actual w2 → w3 diff matches the supplied diff.

**Standards — no findings.** The change uses an awaited observable UI assertion, adds no sleep, and preserves existing assertions. No architecture or ADR change is involved.

**Spec — no blockers; one non-blocking observation.**

1. **Diagnosis: correct mechanism, but “test-only” needs qualification.** Save receives the rendered tab language, not a language refreshed by `drafts`: [LessonLanguageTabs.tsx:384](/Users/strelia/projects/courses/frontend/src/modules/admin/ui/LessonLanguageTabs.tsx:384). Before rerender, that remains `pl`; the fresh track says `en`, so [the guard at line 284](/Users/strelia/projects/courses/frontend/src/modules/admin/ui/LessonLanguageTabs.tsx:284) correctly returns `track_changed` with zero PUTs. This follows MS-D6/M3-R10.

   **Non-blocking UX observation:** [LessonAudioTracksField.tsx:464](/Users/strelia/projects/courses/frontend/src/modules/media/ui/LessonAudioTracksField.tsx:464) closes the rename dialog before starting the asynchronous action. A fast author can therefore encounter the same safe refusal. To check, delay the post-PATCH tracks refresh and click transcript Save immediately after closing the dialog. Keeping rename visibly pending, or withholding transcript Save until the new language is rendered, would improve this interaction. The current contract expressly requires refusal on a language mismatch; it does not promise completion when this dialog closes.

2. **The new wait closes the identified stale-tab-language race at all 11 current helper call sites.** [asr.spec.ts:370](/Users/strelia/projects/courses/frontend/e2e/asr.spec.ts:370) derives the renamed track’s ID from the PATCH URL and targets its `-tab-track:<id>` suffix. It does not depend on that track being selected. An existing orphan `lang:en` tab cannot satisfy this locator. The label and editor’s fixed language derive from the same tab data, so waiting for the label addresses the Save-language race without requiring drafts to settle. Callers needing regenerated drafts retain their separate waits.

3. **No acceptance assertion is weakened.** The wait establishes that the UI has applied the rename before continuing R11. The independent foreign-rename R10 test uses `page.request.patch`, bypasses this helper, and retains its refusal, zero-PUT and preserved-lines checks: [asr.spec.ts:922](/Users/strelia/projects/courses/frontend/e2e/asr.spec.ts:922). R11 still requires replacement confirmation, correct persisted text, fresh segment IDs and preservation of other languages.

4. **W07/W08:** nothing in this delta directly explains the identity 503. The wait adds no HTTP requests. Broader suite-load contribution remains unproven.

Read-only inspection covered source, all helper callers, the contract, report and retained R11 request/trace evidence. I did not run tests, mutation checks, Docker or a live browser, or independently diagnose the 503. **The final-snapshot runtime checks and successful full ordered gate remain outstanding; this SHIP verdict does not close MS3c.**