**BLOCK — 1 Blocker**

1. **Saving a restored snapshot can delete another tab’s newer unsaved lines.** [TranscriptEditor.tsx:464](/Users/strelia/projects/courses/frontend/src/modules/transcript/ui/TranscriptEditor.tsx:464)

   Reproduced with production components:

   - An orphan `lang:uk` contains snapshot A.
   - Rename a track to `uk`; restore detached A into it.
   - Rename the track back to `en`; the orphan tab reappears.
   - In the orphan, restore, edit to C, and blur. Its key now holds C.
   - Save the track, confirming replacement.

   **Result: 1 PUT containing A in `en`; the orphan key containing C is deleted.** After remount, the orphan shows server text with no recovery offer. C was neither saved nor discarded.

   `acceptedDetached` remembers only the source key. Save later clears whatever currently occupies that key, even after it becomes another mounted editor’s own key.

   **Suggested fix:** relinquish deletion ownership when the source becomes active again, and condition deletion on the stored snapshot still being the accepted version. Add the complete rename→restore→rename-back→edit-orphan→save-track→remount regression. Existing tests cover attachment and detached cleanup separately, so they miss this failure.

The original consecutive-Restore and clean-rename fixes passed independent component checks. Initial-read locks released after null/rejection; pending offers refused edits and Save; subsequent edits survived remount. Checked request counts: Cancel **0 PATCH/0 PUT**, PATCH rejection **1/0**, success **1/1**, PUT failure plus retry **1/2**. Foreign and legacy base tokens were omitted correctly. The retained PATCH-failure capture decision produced no additional concrete defect.

Checks used production components/hooks with in-memory storage. Full gates, real-browser Escape focus, and real IndexedDB persistence were **not run**. The working log could not be appended because the filesystem is read-only; no files were changed.