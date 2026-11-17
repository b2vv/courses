**BLOCK**

1. **Blocker — saving B deletes unsaved snapshot A.** [TranscriptEditor.tsx:427](/Users/strelia/projects/courses/frontend/src/modules/transcript/ui/TranscriptEditor.tsx:427)  
   Restore detached A, then Restore detached B, then Save. Reproduced: **one save call containing B; both source keys deleted**. A was neither saved nor explicitly discarded. Track which snapshot the saved lines resolve; retain superseded snapshots. Add a consecutive-Restore→Save test—the existing queue test discards the next offer instead.

2. **Major — clean rename erases an unresolved recovery snapshot on unmount.** [TranscriptEditor.tsx:146](/Users/strelia/projects/courses/frontend/src/modules/transcript/ui/TranscriptEditor.tsx:146), [use-transcript-recovery.ts:136](/Users/strelia/projects/courses/frontend/src/modules/transcript/ui/use-transcript-recovery.ts:136)  
   Open an own-snapshot offer without restoring it, rename the track, then leave. Server-driven `setLines` arms capture; unmount compares those clean lines with the new baseline and clears the key. Reproduced: **zero saves, no Restore/Discard, snapshot deleted**. Distinguish author edits from baseline updates and protect unresolved offers. Test rename→unmount with a pending offer.

3. **Minor — Cancel reports a replacement that never happened.** [LessonLanguageTabs.tsx:302](/Users/strelia/projects/courses/frontend/src/modules/admin/ui/LessonLanguageTabs.tsx:302)  
   Cancelling the `und` language dialog returns `replacement_cancelled`, displaying “Replacement cancelled” / “Заміну скасовано.” Expected counts remain **0 PATCH, 0 PUT**. Use language-dialog cancellation copy or close without an error; assert the resulting message in Cancel/Escape tests.

Both data-loss cases were reproduced with production React components and an in-memory store, without writing files. Full tests, browser checks and real IndexedDB teardown were **not run**.

The working log could not be appended because this session’s filesystem is read-only.