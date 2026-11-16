**BLOCK — 1 Major, 1 Minor**

1. **Major — own-offer Discard still permits deletion of newer unsaved lines.** [use-transcript-recovery.ts:119](/Users/strelia/projects/courses/frontend/src/modules/transcript/ui/use-transcript-recovery.ts:119)

   Normal two-window sequence:

   - B captures unsaved A.
   - A opens the same lesson and receives an own-key offer for A.
   - B edits to C, blurs, and leaves.
   - A discards its displayed offer A, then tabs away from the automatically focused text field.

   **Reproduced with two production editor instances:** Discard initially preserves C; the subsequent clean blur unconditionally clears its key. **Zero PUTs; C is lost.** No narrow timing window is required.

   **Fix:** retain version ownership through clean capture; an unchanged editor must not clear a newer stored version. On mismatch, preserve and offer it. Extend [the own-Discard test](/Users/strelia/projects/courses/frontend/src/modules/transcript/ui/TranscriptEditor.test.tsx:1007) through blur and remount—it currently stops before the deletion.

2. **Minor — asynchronous Discard completion steals focus back.** [TranscriptEditor.tsx:195](/Users/strelia/projects/courses/frontend/src/modules/transcript/ui/TranscriptEditor.tsx:195), [TranscriptEditor.tsx:274](/Users/strelia/projects/courses/frontend/src/modules/transcript/ui/TranscriptEditor.tsx:274)

   Discard a detached offer while storage is pending, then focus another control. When deletion completes, the effect focuses the next offer or first text field, overriding the author’s new focus.

   **Reproduced with deferred storage.** No data lost.

   **Fix:** restore focus only if it still belongs to the initiating recovery control or disappeared with that control.

The named Round 5 detached-Discard and second-detachment ancestry scenarios passed independent component checks, including preservation of newer C. Request counts matched Cancel **0/0**, PATCH rejection **1/0**, success **1/1**, failed PUT plus retry **1/2**.

Checks used production components with memory storage. Full gates, real-browser focus and IndexedDB persistence were **not run**. Read-only permissions prevented appending the working log; no files changed.