**Standards review: BLOCK.** Reviewed `37ec4b0...38c7915`; HEAD remained `38c7915`. No edits, tests, Docker or mutations were run. Unrelated working-tree changes were excluded.

1. **Major — BLOCKER — UNRESOLVED: recovery cleanup can delete another window’s snapshot.**  
   [transcript-history-indexed-database.ts:138](/Users/strelia/projects/courses/frontend/src/modules/transcript/adapters/transcript-history-indexed-database.ts:138) compares only `capturedAt`, generated with `Date.now()` at [use-transcript-recovery.ts:158](/Users/strelia/projects/courses/frontend/src/modules/transcript/ui/use-transcript-recovery.ts:158). Two different captures of the same key can receive the same millisecond timestamp. After the second overwrites the first, cleanup owned by the first still matches and deletes the second.

   This violates the documented port invariant: “delete only the captured version; newer captures must survive” ([TranscriptHistoryStore.ts:16](/Users/strelia/projects/courses/frontend/src/modules/transcript/ports/TranscriptHistoryStore.ts:16)). Verified by tracing capture → ownership → conditional deletion. Existing ownership tests construct the competing version as `capturedAt + 1`, leaving collisions undiscriminated. Use a unique capture revision separate from display time and test equal timestamps with different revisions.

   **Confidence: high; execution NOT_TESTED.** Fowler baseline: **possible Primitive Obsession**—one timestamp represents both time and version identity. This is a heuristic diagnosis, not an additional finding.

2. **Major — BLOCKER — UNRESOLVED: required behavioral verification is incomplete.**  
   Representative locations: [draft.go:125](/Users/strelia/projects/courses/backend/internal/transcript/adapters/postgres/draft.go:125), [LessonLanguageTabs.tsx:284](/Users/strelia/projects/courses/frontend/src/modules/admin/ui/LessonLanguageTabs.tsx:284), [Makefile:129](/Users/strelia/projects/courses/Makefile:129).

   The supplied `behavior-coverage.md` requires individual mutation evidence, including inverse mutations for deleted restrictions. No such execution evidence was supplied; execution was explicitly unauthorized. The final-review contract therefore leaves this required check **NOT_TESTED**, not PASS. An authorized runner must establish the pinned baseline, mutate each construct separately, and record attributable assertion failures or justified equivalent-mutation dispositions.

   **Confidence: high.** This is a verification blocker, distinct from a demonstrated implementation defect.

**Behavior-coverage inventory.** These are source-inspected discriminating tests, **not mutation-verified coverage**.

| Changed constructs | Tests/evidence inspected |
|---|---|
| Migration: leaf backfill, NULL→original, failed/unmatched removal, lowercase snapshots, track PK/cascade, rollback survivor selection | `draft_integration_test.go:testDraftTrackMigration` exercises populated Up/Down/re-Up and cascade. No dedicated negative assertion located for duplicate-leaf rejection or the new snapshot-language CHECK. |
| Pending filters: latest Ready, non-evicted asset; Ready tracks; absent draft; removal of default/stale-trigger restrictions | `testPendingRecognitionsSelector`, `testPendingPerTrack`, `testAutomaticDuplicateNotReclaimed`. |
| Claim pairing, latest-asset refusal, requested-only replacement, metadata/leaf snapshot | `testClaimRejectsForeignAsset`, `testDelayedOldClaim`, `testClaimRecordsTrack`, `testAutomaticDuplicateNotReclaimed`. |
| Completion/recovery: track isolation, attempt token, timeout, deletion and supersession | `testReclaimSameGeneration`, `testRecoverRaceWithReClaim`, `testTracksRecoverIndependently`, deletion tests. |
| Domain staleness: normalized language, `und` family exception, multilingual mismatch; removal of generation/default-leaf comparison | `TestDraftStaleFollowsTrackLanguageAndMode`; old-asset hiding is exercised through repository tests. New constructor TrackID/TrackLanguage rejection lacks dedicated negative cases. |
| Application authorization, ordering, requested-track selection, running refusal, missing duration; terminal claim errors | Draft/application tests, HTTP ordering test, `TestProcessPassesClaimTrackSnapshot`, terminal-claim tests. |
| HTTP serializers and guards: empty arrays, snapshot/current distinction, named errors, body refusal, removed routes, inventory access | `draft_handler_test.go`, `inventory_test.go`, OpenAPI contract tests. |
| Multilingual defaults, strict optional booleans, persistence, conditional manifest rewrite | Config tests; `multilingual_test.go`; `testM2R12MultilingualMetadata`; embedded-default application test. |
| Whisper hint filters and Python kwargs | `TestTranscribeDerivesHintFromTrack`, `TestRealTranscribePythonModelBoundary`. The latter executes production Python against a stub model boundary. |
| Frontend wire validation and mapping | Draft/audio repository tests cover malformed envelopes/items, timestamps, booleans, queued→pending and bodyless regeneration. |
| Query reconciliation: sibling isolation, acknowledgment, timeout/GC survival, stale polling removal, invalidation | `admin-draft-queries.test.ts`, `admin-audio-queries.test.ts`. |
| Tabs/editor: Ready filtering/order, orphan exclusion, stable identity, inactive panels, fresh-track checks, replacement confirmation/token, `und` PATCH→PUT, fresh segment IDs | `LessonLanguageTabs.test.tsx`, editor tests and corresponding ASR e2e scenarios. |
| Recovery: key scoping, legacy/detached placement, load locks, capture/finalization, ancestry, ownership and conditional deletion | Recovery application/hook/editor/IndexedDB tests. Equal-timestamp collision remains uncovered, as finding 1 describes. |
| Reset: resolved names/labels, failed resolution/listing, absent/surviving volumes, truthful reporting | Raw gate proves successful removal. Failure/project-selection probes are author-reported; no committed automated regression tests located. |

**CI and gate evidence.** CI selects frontend unit tests and real PostgreSQL integration tests; those integration harnesses fail when Docker is unavailable. ASR e2e is explicitly excluded in CI through `E2E_SKIP_ASR`. Plain tests cover the changed application, adapter and editor logic, but do not prove live-model integration.

I inspected the raw local gate log: `GATE_EXIT=0`, frontend **1186 passed**, e2e **151 passed**. Its header names pre-commit `7b20d2c`; the recorded Makefile and ASR-test fingerprints match the final files. This supports the supplied gate evidence, but does not replace per-construct mutations. Actual CI results were not supplied: **NOT_TESTED**.

**Rules, ADRs and decision hygiene.** No additional non-tooling Architecture or TypeScript violation was found in the prioritized paths. Backend policies remain inward; engine-specific language support stays in the Whisper adapter; composition roots supply configuration; generated DTOs remain in adapters. Cross-module access contracts are unchanged, with modules continuing to read their own facts.

Relevant accepted ADRs **001, 002, 003, 006, 013 and 014** remain compatible with this change. Historical freshness debt—such as ADR 006’s “No Python runtime now”—predates this diff and should not be read as current implementation documentation.

The material choices—per-track persistence and retaining publication through `PUT transcript` with client preflight—are recorded with date, author, rejected alternatives and verification mechanisms in `plan-ms2-3.md`. No contradiction requiring a new ADR under `.claude/standards.md:12` was found. Broader promotion can be decided by the owner during codification; absence of another ADR is not itself a finding. Process history stays in phase artifacts; the existing editor, Query, transport and React Aria seams are reused.

Counts: **Critical 0 · Major 2 · Minor 0**. Both Majors are **BLOCKER / UNRESOLVED**; the smell is included in finding 1.

**VERDICT (Standards, pr109-final @ 38c7915): BLOCK**