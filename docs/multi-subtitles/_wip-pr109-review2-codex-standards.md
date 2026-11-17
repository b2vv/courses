**BLOCK.** Previous finding 1 is closed. Previous finding 2 is substantially addressed, but required behavioral evidence remains incomplete.

Reviewed `38c7915...ad3612c`, with whole-PR spot-checks where the delta interacts with existing code. The checkout is ahead of that revision; relevant production/test files match the pinned snapshot. No edits, tests, Docker operations or git writes were performed.

1. **Major — BLOCKER — UNRESOLVED: 24 mutations lack executed discrimination evidence.**  
   [_wip-pr109-mutation-dispositions.md:39](/Users/strelia/projects/courses/docs/multi-subtitles/_wip-pr109-mutation-dispositions.md:39), [ci.yml:115](/Users/strelia/projects/courses/.github/workflows/ci.yml:115).

   All cited e2e scenarios reach the relevant constructs, as detailed below. However, none was executed under its mutation. An unmutated final gate cannot establish mutation detection. Moreover, CI excludes `asr.spec.ts`; remaining behaviors whose sole detector is that suite need deterministic CI-selected coverage under `behavior-coverage.md`.

   Some new plain tests already cover former `E2E_ONLY` behavior—for example, application rename forwarding—but those mutations need re-execution against the final tests. Reuse those tests, add missing discriminators at the agreed seams, and record attributable failures. **Verification: source, CI selection, mutation diffs and journals inspected. Confidence: high. Status: NOT_TESTED.**

2. **Major — BLOCKER — UNRESOLVED: X-019 remains inconclusive.**  
   [use-transcript-recovery.ts:140](/Users/strelia/projects/courses/frontend/src/modules/transcript/ui/use-transcript-recovery.ts:140), [_wip-pr109-mutation-dispositions.md:21](/Users/strelia/projects/courses/docs/multi-subtitles/_wip-pr109-mutation-dispositions.md:21).

   The raw whole-suite record reports **1,191 passed tests, zero failed tests, one failed file**, with empty `fullName` and `unhandledRejection: retirement refusal`. This demonstrates an escaped rejection under mutation, but supplies no named failing assertion satisfying the supplied contract. It is neither equivalent nor an accepted kill.

   Add a named asynchronous-retirement test that observes rejection containment and retained editing/save behavior, then execute baseline and mutation. The new detached-inventory rejection test provides an existing pattern. **Verification: exact mutation, production catch, existing refusal test and raw failure inspected. Confidence: high. Status: INCONCLUSIVE.**

**Previous findings and test delta**

Capture ownership now uses `crypto.randomUUID()` revisions throughout the port, both stores, hook ownership, accepted ancestry and editor cleanup. Equal-timestamp tests exercise competing captures and preservation of the replacement; supplied mutation evidence includes a named failure when revision uniqueness is removed. **Previous finding 1: CLOSED.**

I reconciled **74 + 60 + 53 + 21 = 208** distinct closure records and spot-checked raw failures against production paths and final assertions. The additions provide useful validation, ordering, concurrency, serialization and lifecycle coverage. Database concurrency tests observe actual lock contention; reset tests isolate a fake Docker executable; teardown addresses the previously reported QueryClient retention. No additional blocking test-quality, placement, leak or cost issue was found. **Previous finding 2: PARTIALLY CLOSED**, subject to findings above.

**EQUIVALENT decisions — each accepted on the supported production path**

| Part / ID | Decision | Independently checked reason |
|---|---|---|
| fe-a A-008 | ACCEPT | Null state still fails downstream mapping. |
| fe-a R-017 | ACCEPT | Missing snapshot access throws inside the catch; no deletion occurs. |
| fe-a H-014 | ACCEPT | Current props and ownership remain aligned before supported capture events. |
| fe-a L-006 | ACCEPT | The disabled button prevents invocation without a track. |
| fe-a E-038 | ACCEPT | Production orphan identity fixes its language; the extra warning condition is redundant. |
| fe-a E-043 | ACCEPT | Fixed editors derive save language from props; changed internal language is unused. |
| fe-a E-044 | ACCEPT | Same fixed-language rule for ASR loading. |
| fe-a E-061 | ACCEPT | Successful-save state makes the editor clean; render synchronization updates edit language. |
| fe-a X-006 | ACCEPT | Validation excludes null before the state switch. |
| fe-a X-008 | ACCEPT | Validated states all have defined mappings. |
| fe-a X-011 | ACCEPT | Undefined-version access is caught without storage effects. |
| fe-a X-015 | ACCEPT | Missing-store access is caught; returned offers remain empty. |
| fe-a X-016 | ACCEPT | Production handlers invoke detached handling only with a selected detached offer. |
| fe-b B013 | ACCEPT | No production descendant draft query keys exist. |
| fe-b B015 | ACCEPT | The later `...previous` overwrites the initializer; absent previous still produces `und`. |
| fe-b B016 | ACCEPT | Same spread behavior; absent previous still produces `false`. |
| fe-b B062 | ACCEPT | React Aria read-only state blocks switch changes while busy. |
| fe-b B068 | ACCEPT | Same independent read-only protection for row switches. |
| fe-b B081 | ACCEPT | Inventory parsing admits lowercase language tags only. |
| fe-b B167 | ACCEPT | Orphan identifiers receive the same validated lowercase tags. |
| fe-b B175 | ACCEPT | Detected language is already lowercase or absent. |
| fe-b B176 | ACCEPT | Unavailable keys render no selected item; both forms fail the options-membership guard. |
| fe-b B178 | ACCEPT | Fresh inventory language is already normalized. |
| fe-b B194 | ACCEPT | No production descendant transcript-inventory keys exist. |
| be-a R07 | ACCEPT | Completion ignores track-language copies; Claim persisted the snapshot. |
| be-a R08 | ACCEPT | Completion ignores track-mode copies; Claim persisted the snapshot. |
| be-a R09 | ACCEPT | Completion ignores source leaf; reads join the track. |
| be-a R13 | ACCEPT | Constructor does not validate leaf; extraction already consumed it and completion ignores it. |
| be-a R15 | ACCEPT | Failed completion does not consume copied track language. |
| be-a R16 | ACCEPT | Failed completion does not consume copied track mode. |
| be-a R17 | ACCEPT | Failed completion does not consume copied source leaf. |
| be-a W09 | ACCEPT | Production Engine invocation never supplies both language and multilingual flags. |
| be-a S05 | ACCEPT | Latest-asset identity plus the persisted lesson/asset pairing implies lesson scope. |
| be-a S28 | ACCEPT | The latest-asset subquery independently enforces the requested lesson. |
| be-a S64 | ACCEPT | Recovery’s only production consumer discards source leaf. |
| be-a G05 | ACCEPT | The remaining anchored grammar already requires lowercase ASCII. |
| be-a G13 | ACCEPT | The primary key independently enforces non-null track identity. |
| be-b B51 | ACCEPT | Master regeneration never reads the multilingual constructor default. |

**E2E_ONLY reachability — all 24 citations accepted as static traces**

“Reaches” below does **not** mean mutation-verified.

| IDs | Cited scenario | Reachability judgment |
|---|---|---|
| B007, B055, B147 | M3-R15 | Switch → route → composition → application → persisted mode; checked state and stale-draft assertions discriminate. |
| B145, B184, B185 | M3-R8 | Actual CourseEditor prop chain must produce language tabs. |
| B146 | M3-R14 | Upload wiring reaches multilingual minting, tab labeling and stored draft mode. |
| B148 | M3-R8 | Real audio rows are required for track tabs and ASR controls. |
| B152 | M3-R8 | Distinct source texts are loaded and saved through their respective track drafts. |
| B153 | C11 / M3-R8 | Regeneration reaches the API with the selected track ID. |
| B155 | M3-R10 | Failed fresh track read must produce the named verification error. |
| B156 | M3-R8 | Successful guarded Save requires the fresh track inventory. |
| B192 | M3-R8 | Inventory must load for the real lesson before tabs appear. |
| A14 | C11 | Regeneration after completion requires a new marker; retained task identity otherwise conflicts. |
| R06 | C13 | No-speech completion must retain track identity to persist its terminal state. |
| H11 | C03 | Serialized state must reach the visible ready-state assertion. |
| H14 | C11 | Valid asset identity is required by draft parsing before regeneration. |
| H16 | C13 | Empty segments must serialize as an array for no-speech rendering. |
| H19 | C11 | Segment IDs must survive serialization and validation. |
| Q05 | C03 | Task handler must forward track identity for Claim to succeed. |
| S68 | C11 | Repository duration reaches regeneration’s timeout prerequisite. |
| H38 | C03 | GET-drafts authentication must establish Actor context. |
| H39 | C11 | POST regeneration authentication must establish Actor context. |
| H40 | C03 | Transcript-inventory authentication must succeed before the editor renders. |

**Standards and final-review coverage**

Architecture and TypeScript boundaries remain intact. The revision fix reuses existing seams; the Ukrainian copy names the actual «Аудіодоріжки» section. No additional Fowler-baseline finding warrants a change.

Existing ADRs **001, 002, 003, 013 and 014** remain compatible; ADR **006** retains pre-existing stale implementation prose. The delta introduces no contradiction requiring a new ADR under the repository threshold. Revision identity and destination decisions have dated author/mechanism records in the fix journal. Process evidence stays in phase artifacts.

The added tests provide independent behavioral oracles without weakening existing assertions. The production delta remains narrow. The **ordered final `make check` with full e2e at `ad3612c` is NOT_TESTED** and remains the owner’s required delivery gate; earlier aggregate evidence does not cover the changed persistence behavior.

Counts: **Critical 0 · Major 2 · Minor 0** — both Majors **BLOCKER / UNRESOLVED**.  
**EQUIVALENT accepted 38/38.**

**VERDICT (Standards, pr109-final-2 @ ad3612c): BLOCK**