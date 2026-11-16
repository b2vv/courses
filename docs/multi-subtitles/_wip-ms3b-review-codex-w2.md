**BLOCK**

1. **Blocker — pending recovery offer loses new author edits.** [use-transcript-recovery.ts:117](/Users/strelia/projects/courses/frontend/src/modules/transcript/ui/use-transcript-recovery.ts:117)  
   Leave the own-snapshot offer pending → type new lines → blur → unmount. Reproduced with the production editor: **0 saves, 0 snapshot writes; only the old recovery text survives**. Blur cancels the timer before the guard refuses capture. Discarding the offer afterward does not re-arm capture, so immediately leaving still loses the edits.  
   **Fix:** preserve both versions, or require resolving the offer before editing. Retain and flush pending author changes after resolution. The hook test at line 102 currently expects suppressed writes without checking whether the new edits survive remount.

2. **Blocker — capture can overwrite an unread recovery snapshot.** [use-transcript-recovery.ts:117](/Users/strelia/projects/courses/frontend/src/modules/transcript/ui/use-transcript-recovery.ts:117)  
   Delay the initial history read → edit → blur. `unresolvedOffer` is still false, and `hasRead` guards only clears. Reproduced: **1 write while `isLoaded=false`**, replacing prior unsaved lines before Restore/Discard was offered.  
   **Fix:** protect writes until the initial read resolves, while retaining new author edits for later capture. Add a deferred-read test that verifies both versions survive.

3. **Major — failed PATCH changes recovery storage against MS-D8.** [TranscriptEditor.tsx:444](/Users/strelia/projects/courses/frontend/src/modules/transcript/ui/TranscriptEditor.tsx:444)  
   Confirm the `und` dialog and return `duplicate_language`. Reproduced: **1 PATCH, 0 PUT, 1 recovery write after confirmation**. The accepted contract requires recovery keys to remain untouched on PATCH failure. [LessonLanguageTabs.test.tsx:572](/Users/strelia/projects/courses/frontend/src/modules/admin/ui/LessonLanguageTabs.test.tsx:572) explicitly expects the prohibited write.  
   **Fix:** distinguish PATCH rejection from PUT failure and skip failure-triggered capture for PATCH rejection; assert zero writes, clears and finalization.

The three original round-one scenarios passed independent component checks. The four additional browser fixes were inspected in source/tests; actual-page Escape focus remains **not independently verified**.

Inline checks used production components/hooks with in-memory storage. Full tests, browser checks and real IndexedDB persistence were not run. The sandbox rejected appending the working log; no files were changed.