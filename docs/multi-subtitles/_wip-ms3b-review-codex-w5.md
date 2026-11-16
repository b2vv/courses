**BLOCK — 1 Major, 1 Minor**

1. **Major — Discard deletes newer edits from another window.** [TranscriptEditor.tsx:241](/Users/strelia/projects/courses/frontend/src/modules/transcript/ui/TranscriptEditor.tsx:241)

   Reachable in four steps with two windows:
   1. Window B captures unsaved text A in an old track.
   2. Window A replaces the video and displays detached offer A.
   3. B edits to C, blurs, then leaves.
   4. A discards its displayed offer A.

   **Reproduced:** unconditional `clear(key)` deletes C; zero transcript saves occurred. All newer unsaved lines are lost. No narrow timing window is required.

   **Fix:** condition Discard on the offered version using `clearIfVersion`. On mismatch, refresh the offer instead of marking the whole key handled. Add this two-editor regression.

2. **Minor — cleanup ownership is lost when the recovered copy itself becomes detached.** [TranscriptEditor.tsx:240](/Users/strelia/projects/courses/frontend/src/modules/transcript/ui/TranscriptEditor.tsx:240)

   Six-step scenario: type legacy A → first video becomes Ready → restore A into B → replace video, capturing B → restore detached B into C → Save and revisit.

   **Reproduced:** one save containing A; B deleted, original A retained and offered again. This repeats every visit until separately resolved. No lines were lost. Detached Restore retains B but ignores B’s `acceptedSource`.

   **Fix:** preserve versioned cleanup ancestry through detachment; retire only the versions represented by the saved work. Extend the reload regression through a second track replacement.

The original fix4 scenarios passed component checks: unchanged A is retired after reload→Restore→Save; newer C survives and remains offered. IndexedDB conditional deletion uses one readwrite transaction.

Verification used production components/helpers with memory storage. Full gates and real-browser IndexedDB were not run. No files changed; read-only permissions prevented appending the working log.