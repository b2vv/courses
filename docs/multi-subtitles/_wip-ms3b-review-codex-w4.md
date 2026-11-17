**BLOCK — 1 Blocker, 1 Major**

1. **Blocker — compare-delete can still erase newer unsaved lines.** [recovery.ts:243](/Users/strelia/projects/courses/frontend/src/modules/transcript/application/recovery.ts:243)

   The read and clear are separate storage operations. Another browser window can queue capture C after the read of accepted snapshot A but before deletion. The comparison passes against A; the subsequent clear deletes C.

   Reproduced with the production helper and a deferred read: **accepted A, newer C written, cleanup leaves `null`**. C was neither saved nor discarded. Existing tests cover changes visible before comparison, not writes between read and deletion.

   **Fix:** compare and delete atomically in one IndexedDB readwrite transaction. Add an interleaved-write regression.

2. **Major — reloading after Restore loses cleanup ownership, so saved text keeps being offered.** [TranscriptEditor.tsx:111](/Users/strelia/projects/courses/frontend/src/modules/transcript/ui/TranscriptEditor.tsx:111), [TranscriptEditor.tsx:497](/Users/strelia/projects/courses/frontend/src/modules/transcript/ui/TranscriptEditor.tsx:497)

   Restore detached A → unmount, capturing A under the tab’s own key → remount → restore the own-key offer → Save.

   Reproduced: **1 PUT containing A; own key deleted; detached source retained**. After updating the server fixture to the saved response, two subsequent mounts both offered A as “Unsaved lines.” This repeats every visit until that detached offer is separately resolved. `acceptedDetached` exists only in memory and is not recovered from the own snapshot.

   **Fix:** preserve the accepted source’s key/version through capture and remount, then retire that exact source after Save using atomic version checking. Extend the existing restore/remount test through Save and subsequent visits.

The original Round 3 scenario now preserves orphan C. Restore moved focus correctly, and an unrelated rerender preserved external focus. Request counts checked: Cancel **0 PATCH/0 PUT**, PATCH rejection **1/0**, success **1/1**; cancellation copy is corrected.

Checks used production components/helpers with in-memory storage. Full gates and real-browser IndexedDB checks were not run. No files changed; read-only permissions prevented appending the working log.