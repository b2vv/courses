**BLOCK. F1 and F2 are closed. F3 remains incomplete, and the required final gate is unrun.**

Reviewed `37ec4b0...ad3612c8accdafbf539de6409f96a5807a2cf860`. The checkout is at `3e7de7e`, which adds the re-review package; production review used the pinned Git objects. No edits, Git writes, Docker commands or tests were run.

**1. Findings and previous verdict**

**F1 — CLOSED.**

[recovery.ts:233](/Users/strelia/projects/courses/frontend/src/modules/transcript/application/recovery.ts:233) now selects:

1. A **track** tab matching `snapshot.language`, case-insensitively.
2. Otherwise, the default **track** tab.
3. With tabs but neither permitted destination, no offer; storage remains untouched.

This follows [plan-ms2-3.md:406](/Users/strelia/projects/courses/docs/multi-subtitles/plan-ms2-3.md:406). Orphan tabs and arbitrary non-default tracks cannot become fallback destinations.

The dated author decision in [report.md:801](/Users/strelia/projects/courses/docs/multi-subtitles/report.md:801) is **within the spec**, interpreted as the fallback case where no matching track exists. It does not require an owner-approved behavior change: the accepted plan names no other destination, and Ready assets require exactly one Ready default track ([audiotrack.go:157](/Users/strelia/projects/courses/backend/internal/media/domain/audiotrack.go:157)). The defensive empty case preserves the snapshot.

The regression test at [recovery.test.ts:409](/Users/strelia/projects/courses/frontend/src/modules/transcript/application/recovery.test.ts:409) distinguishes matching-orphan-only, matching-track, default fallback and absent-default cases. The mounted-editor test also preserves orphan ownership.

**F2 — CLOSED for the requested visual evidence.**

All 45 artifacts—42 PNGs, measurements and two harness files—match `ad3612c`. The report binds the browser run to `5ef53d1`; subsequent production frontend changes are absent, so that visual evidence remains applicable.

I inspected:

- `s1-tabs-und-draft-320.png`
- `s4-und-dialog-390-text200.png`
- `s5-audio-tracks-390-text200.png`
- `s6-rename-dialog-1440.png`

The JSON contains six states at all seven configurations: 390/768/1024/1440, 200% text at 390 and 1440, and 320. It records no duplicate IDs and document width equal to viewport width throughout. Reported small/outside controls correspond to React Aria’s clipped, untabbable HiddenSelect, corroborated against its implementation.

Keyboard records cover arrow selection, Tab into the active panel, visible outlines, dialog entry and Escape focus restoration. The switch’s Space interaction is covered separately by its component test. This supplies the previously missing evidence for M3-R7 and **the visual portion** of M3-R14. It does not prove the final integrated ASR journey.

**F3 — Major verification BLOCKER, UNRESOLVED / NOT_TESTED. Confidence: high.**

[report.md:861](/Users/strelia/projects/courses/docs/multi-subtitles/report.md:861) still identifies **24 E2E_ONLY mutations and one INCONCLUSIVE mutation**.

The [behavior-coverage contract](/Users/strelia/project-data/ai-factory-kit/skills/arch/code-review/references/behavior-coverage.md:10) requires executed discrimination, including appropriate plain coverage when the sole service test skips in CI. Static attribution to an e2e test is insufficient.

Concrete remaining examples:

- **A14:** dropping the regeneration marker at [draft.go:133](/Users/strelia/projects/courses/backend/internal/transcript/application/draft.go:133).
- **Q05:** dropping the decoded track identity before processing at [handler.go:46](/Users/strelia/projects/courses/backend/internal/transcript/adapters/asynq/handler.go:46).
- **H38–H40:** authentication middleware wiring. Current HTTP fixtures install no-op authentication and inject the actor directly ([draft_handler_test.go:80](/Users/strelia/projects/courses/backend/internal/transcript/adapters/http/draft_handler_test.go:80)).
- **X-019:** the retirement-failure mutation produces a global unhandled rejection without a named failing assertion. It remains INCONCLUSIVE.

CI explicitly excludes `asr.spec.ts`; its configured unit/integration selection does not establish coverage for these surviving paths. An **unmutated** final e2e pass would not close these mutation requirements.

Required closure: executable discriminating coverage for the remaining cases, baseline/mutant evidence with attributable assertions, and the applicable CI-selected plain tests. Preserve real integration coverage.

**F4 — Major verification BLOCKER, UNRESOLVED / NOT_TESTED. Confidence: high.**

[plan-ms2-3.md:520](/Users/strelia/projects/courses/docs/multi-subtitles/plan-ms2-3.md:520) requires the ordered gate, full integration and full e2e on the final revision. The [package:22](/Users/strelia/projects/courses/docs/multi-subtitles/_wip-pr109-review-package-2.md:22) explicitly states that this has not run on `ad3612c`.

Attempt 7’s earlier passing baseline remains useful evidence. It cannot establish the final combined recovery changes and expanded test suite.

**2. Mutation evidence and equivalent dispositions**

I recovered **500 distinct KILLED records** and **208 distinct CLOSED gap IDs**: FE-A 74, FE-B 60, BE-A 53, BE-B 21. I checked closure records, target bindings and representative mutation→production→assertion paths across serialization, lifecycle, SQL, migration, startup wiring and reset handling.

The supplied raw evidence supports those closures. Hash differences were traced to the recorded test renames and QueryClient teardown addition; these do not weaken the relevant assertions. These are reviewed execution records, not reviewer-executed tests.

I accept **all 38 EQUIVALENT dispositions**, individually identified below, at the stated production seams. None is counted as KILLED.

| IDs | Disposition and verified reason |
|---|---|
| FE-A A-008, X-006, X-008 | ACCEPTED: adjacent validation/mapping independently rejects the same invalid state. |
| FE-A R-017, X-011, X-015 | ACCEPTED: removed early guards lead to the existing caught failure, with identical observable no-op behavior. |
| FE-A H-014 | ACCEPTED: corresponding refs update synchronously; the proposed mismatch cannot interleave through current callers. |
| FE-A L-006 | ACCEPTED: the same missing identity disables the only invoking control. |
| FE-A E-038 | ACCEPTED: orphan language changes remount their keyed editor; the extra warning cannot arise in supported flows. |
| FE-A E-043, E-044, E-061 | ACCEPTED: fixed save language and clean-editor alignment make these internal assignments observationally redundant. |
| FE-A X-016 | ACCEPTED: private callers already require the captured detached value. |
| FE-B B013, B194 | ACCEPTED: current query namespaces contain no descendant keys affected by `exact`. |
| FE-B B015, B016 | ACCEPTED: the subsequent previous-snapshot spread supplies the same values. |
| FE-B B062, B068 | ACCEPTED: React Aria read-only state independently suppresses the callback. |
| FE-B B081, B167, B175, B178 | ACCEPTED: admitted inventory/draft language values are already lowercase. |
| FE-B B176 | ACCEPTED: unavailable selections render without a selected item and fail the same submission validation. |
| BE-A R07, R08, R09, R13, R15, R16, R17 | ACCEPTED: production completion does not consume these duplicate fields; claim persistence and track reads retain the authoritative facts. |
| BE-A W09 | ACCEPTED: the production engine emits mutually exclusive language/multilingual arguments. |
| BE-A S05, S28 | ACCEPTED: retained latest-asset/pairing predicates imply the removed equality under production writer invariants. |
| BE-A S64 | ACCEPTED: recovery callers do not consume this returned leaf. |
| BE-A G05 | ACCEPTED: the retained anchored lowercase ASCII regex implies the lowercase equality. |
| BE-A G13 | ACCEPTED: the added primary key independently establishes NOT NULL. |
| BE-B B51 | ACCEPTED: master regeneration never consumes the constructor’s multilingual default. |

Final accounting: **708 demonstrated kills/closures + 38 accepted equivalents + 24 E2E_ONLY + 1 INCONCLUSIVE = 771**.

The unresolved E2E_ONLY IDs are:

- FE-B: B007, B055, B145–B148, B152, B153, B155, B156, B184, B185, B192.
- BE-A: A14, R06, H11, H14, H16, H19, Q05, S68, H38–H40.

**3. Copy, scope and decision checks**

`5ef53d1` changes only two Ukrainian strings to identify the existing section as **«Аудіодоріжки»**. This preserves the instruction’s meaning and improves navigation accuracy. The enlarged dialog screenshot includes the changed copy. No scope creep found.

The remaining delta comprises the recovery corrections, verification artifacts and tests. The capture revision change addresses the previously identified ownership collision; no additional product feature was introduced.

Both ADR checks were completed:

- **Existing decisions:** accepted ADRs 001, 002, 003, 006, 013 and 014 remain behaviorally compatible with this scope. ADR 006 retains historical wording about scheduled implementation and “No Python runtime now”; that predates this delta.
- **Decision material:** per-track identity, publication, recovery placement and multilingual configuration already have dated MS-D decisions. The no-default defensive case does not alter valid-state acceptance or contradict an ADR; no new owner decision or ADR is required for it.

The eight coverage questions were considered: gate remains blocked; plan adherence is supported subject to the evidence gaps; no additional Spec-axis rule contradiction was found; ADR checks are above; affected contracts/report references were inspected; process history remains in task evidence; test value improved substantially but remains incomplete under F3; implementation continues to reuse existing repository, query, editor and recovery seams.

**4. Acceptance-row table**

**PASS** means the row’s specified seam has accepted discriminating evidence, reused where inputs match. **PARTIAL** means a required portion remains unproved. These row results do not waive the final ordered gate.

Abbreviations: **PG** = transcript PostgreSQL tests; **Tabs** = `LessonLanguageTabs.test.tsx`; **ASR** = `frontend/e2e/asr.spec.ts`. Test names are shortened to identifiable prefixes.

| Row | Implemented? | Discriminating test / observable oracle | Status |
|---|---|---|---|
| M2-R1 | Yes | PG `M2-R1 migration 00031 maps populated drafts and round trips`; schema/rollback tests check mapped/deleted rows, keys, cascade and survivors | PASS |
| M2-R2 | Yes | PG `M2-R2 … pending follows each ready track regardless of default`; eligibility/order/limit cases assert exact identities | PASS |
| M2-R3 | Yes | PG `M2-R3 C21 rejects foreign lesson asset track non-ready and generation`; claim/reset cases assert refusals, zero writes and snapshots | PASS |
| M2-R4 | Yes | PG slow-old-completion and delayed-old-Ready-claim cases preserve current visible output | PASS |
| M2-R5 | Yes | `TestTwoTracksEnqueueDistinctTaskIDs` asserts exact distinct IDs | PASS |
| M2-R6 | Yes | `TestTranscribeDerivesHintFromTrack`, pinned language codes and claim-snapshot forwarding assert argv/log/snapshot | PASS |
| M2-R7 | Yes | PG stale/automatic-duplicate/requested-replacement cases discriminate persistence; application marker mutation A14 remains open | PARTIAL |
| M2-R8 | Yes | PG Recover×re-Claim and independent-track cases reject obsolete/cross-track tokens | PASS |
| M2-R9 | Yes | PG deleting-one-track case asserts draft deletion and sibling/published preservation | PASS |
| M2-R10 | Yes | Regenerate/list/removed-route/inventory/OpenAPI tests; serializer/authentication mutations H11/H14/H16/H19/H38–H40 remain open | PARTIAL |
| M2-R11 | Yes | `TestLegacyTaskRefusedWithoutProcessingOrRetries` asserts processor not called and terminal retry policy | PASS |
| M2-R12 | Yes | Real Python-script boundary, config, media persistence/rewrite and executable-wiring tests assert flags and effects | PASS |
| M3-R1 | Yes | Tabs ordering/filtering/newest-Ready/fallback tests; final integrated transcript regression remains unrun | PARTIAL |
| M3-R2 | Yes | Tabs `save uses lowercased tab language, fresh reads, and new ids…`; ASR network journey pending | PARTIAL |
| M3-R3 | Yes | Draft-guard cases remove Load and name stale/language/mode mismatch | PASS |
| M3-R4 | Yes | Tabs PATCH→PUT, cancellation, duplicate-language and failed-PUT cases; reload/restore browser portion pending | PARTIAL |
| M3-R5 | Yes | Per-track Regenerate/pending tests and query O7 polling cases assert isolated markers and no automatic replacement | PASS |
| M3-R6 | Yes | Placement, mounted ownership, priority and cleanup tests; final browser persistence plus X-019 remain unproved | PARTIAL |
| M3-R7 | Yes | Unique-ID/inert-panel/keyboard component tests plus revision-bound screenshots, geometry and focus records | PASS |
| M3-R8 | Yes | ASR `M3-R8 R2 R4 success and R6 two-track drafts…`: independent Welcome/dubbed text, languages and persisted contents | NOT_TESTED |
| M3-R9 | Yes | Updated ASR C06/C11/C17/C18/S6 journeys | NOT_TESTED |
| M3-R10 | Yes | Tabs refusal/error cases assert zero PUT and retained lines; foreign-PATCH/503 browser journeys pending | PARTIAL |
| M3-R11 | Yes | Fresh-ID/rename/token component cases; ASR a–d and detached continuation pending | PARTIAL |
| M3-R12 | Yes | Unmount flush, first-Ready transition and dirty-orphan attachment tests; replacement-video browser journey pending | PARTIAL |
| M3-R13 | Yes | Original-token, replacement-token and conflict component cases; two-editor browser conflicts pending | PARTIAL |
| M3-R14 | Yes | Switch default/metadata/Space tests and visual evidence accepted; minted-track→ASR browser chain pending | PARTIAL |
| M3-R15 | Yes | Mode guard and invalidation tests; integrated `en`/`und` switch→Regenerate journey pending | PARTIAL |

Totals: **13 PASS, 12 PARTIAL, 2 NOT_TESTED, 0 MISSING**. **M3-R8 and M3-R9 have only browser-level acceptance proof and remain NOT_TESTED** until the final gate runs on `ad3612c`.

Severity counts: **2 Major verification blockers; 0 remaining behavior defects; 0 Minor/non-blocking findings.**

**VERDICT (Spec, pr109-final-2 @ ad3612c): BLOCK**