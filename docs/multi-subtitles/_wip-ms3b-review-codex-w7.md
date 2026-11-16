**SHIP — 1 Minor; no Blocker/Major found.**

1. **Minor — timestamp collisions defeat version ownership.** [use-transcript-recovery.ts:158](/Users/strelia/projects/courses/frontend/src/modules/transcript/ui/use-transcript-recovery.ts:158), [cleanup:130](/Users/strelia/projects/courses/frontend/src/modules/transcript/ui/use-transcript-recovery.ts:130)

   Two windows can capture different lines A and C under the same key within one millisecond. Both receive the same `capturedAt`; successful Save of A can then delete unsaved C.

   **Reproduced with production hooks and a clock forced to produce equal timestamps.** This requires a rare timing collision, so Minor under your severity rule. The deletion matrices use distinct timestamps and miss it.

   **Suggested fix:** use a unique capture identifier for atomic comparison; retain `capturedAt` for ordering. Add equal-timestamp/different-content coverage.

The deletion inventory is complete: ordinary cleanup uses `clearIfVersion`; only the intentional cross-author sweep deletes unconditionally. Independent component checks passed newer-version preservation, Discard→blur→remount, deferred focus, initial-read lock release, unmount flushing and orphan attachment.

PATCH/PUT counts verified: Cancel **0/0**, PATCH rejection **1/0**, success **1/1**, failed PUT plus retry **1/2**. Legacy tokens were omitted correctly.

Full gates, real-browser focus and persistent IndexedDB were **not run**. No files changed; read-only permissions prevented appending the working log.