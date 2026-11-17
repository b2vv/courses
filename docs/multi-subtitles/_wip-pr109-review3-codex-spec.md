**The mutation findings are closed. No new code or evidence finding remains; the final ordered gate is still pending.**

Reviewed `3e7de7e..c2f4310`: one commit, ten test files plus twelve evidence documents. **Production code, dependencies and CI configuration are unchanged.** No edits, Git writes, Docker commands or tests were performed during this review.

**1. Mutation closure**

I accept the [gaps-e2e package](/Users/strelia/projects/courses/docs/multi-subtitles/_wip-pr109-gaps-e2e-review-codex.md): **25/25 resolved — 6 KILLED-by-existing and 19 CLOSED**.

I inspected the raw named-failure records for all 25 rows. For all eleven backend rows, I additionally matched baseline PASS, mutant FAIL, the recorded assertion and mutation-patch hash. Nine delivered test-file hashes match the journals exactly; the tenth differs only by the reviewed X-019 cleanup.

The independently traced sample includes:

| Mutation | Verified discriminator |
|---|---|
| H38 | Removing GET-drafts authentication fails `draft_inventory/valid_cookie`: `authentication calls=0; want 1`. |
| H39 | Removing regeneration authentication fails the corresponding `regenerate/valid_cookie` assertion. |
| H40 | Removing transcript-inventory authentication fails `transcript_inventory/valid_cookie`. |
| X-019 | Baseline passes; mutant fails **named F02**: `async retirement must contain the rejected cleanup`. Both Save results and continued editing are asserted before that failure. Global F01 is not counted. |
| B055, existing kill | The cancellation-signal test also requires `multilingual=true`; forcing `false` fails its repository-call assertion. |
| B156, existing kill | Returning empty fresh tracks prevents Save; the existing guarded-Save test fails to observe `Transcript saved.` |
| A14 | Dropping the selected attempt marker fails the exact timestamp assertion. |
| Q05 | Real task handler → recognition application → Claim detects the missing track ID. |
| S68 | Real PostgreSQL execution detects missing duration: `nil` instead of `123.75` seconds. |

The [authentication fixture](/Users/strelia/projects/courses/backend/internal/transcript/adapters/http/route_auth_test.go:36) uses real route registration and identity middleware, valid/missing/invalid cookies, and no pre-injected Actor. Only session lookup is substituted.

The [X-019 cleanup](/Users/strelia/projects/courses/frontend/src/modules/transcript/ui/TranscriptEditor.test.tsx:1827) moves listener registration and mounting inside `try/finally`, with optional unmount. Its stimuli and discriminating assertions remain unchanged, so the inspected mutation evidence remains applicable. The prior cleanup finding is closed.

CI selects these detectors through `make test` or `make integration`; they no longer depend solely on CI-excluded `asr.spec.ts`. This establishes selection, not a claim of a new live CI run.

Accounting:

**500 KILLED + 208 CLOSED + 25 newly resolved + 38 accepted EQUIVALENT = 771/771.**

That is **733 demonstrated kills/closures and 38 equivalents**, with **0 E2E_ONLY and 0 INCONCLUSIVE remaining**. Previous unaffected evidence and equivalent dispositions remain accepted.

**2. Scope and decisions**

No new behavior defect, scope creep or weakened assertion was found. The six removed test lines are preserved through expanded assertions.

Existing accepted ADRs **001, 002, 003, 006, 013 and 014** remain compatible with the unchanged production implementation. ADR 006’s historical implementation/Python wording predates this delta. No new architectural decision or owner decision is introduced.

The eight coverage questions are accounted for: gate pending; plan adherence supported; no new rule contradiction; ADR checks above; affected evidence references inspected; process history remains in task documents; tests provide attributable discrimination; existing test seams are reused.

**3. Remaining gate**

The final ordered **`make check` at `c2f4310` has not run**, with owner consent pending. It must include the required full integration and full e2e, including ASR acceptance.

This is the remaining delivery gate, **not a code finding**. Package-level execution does not establish final integrated acceptance.

**4. Updated acceptance-row table**

All rows remain implemented. PASS denotes accepted evidence at the specified seam; PARTIAL retains an unexecuted browser portion.

| Row | Evidence / remaining portion | Status |
|---|---|---|
| M2-R1 | Populated migration, rollback, keys and cascade | PASS |
| M2-R2 | Pending eligibility, ordering and identities | PASS |
| M2-R3 | Claim identity and refusal cases | PASS |
| M2-R4 | Old-asset claims and completion isolation | PASS |
| M2-R5 | Distinct per-track task IDs | PASS |
| M2-R6 | Language hints and claim-snapshot forwarding | PASS |
| M2-R7 | Existing persistence proof plus executed A14 discriminator | **PASS ← PARTIAL** |
| M2-R8 | Recovery/re-Claim token isolation | PASS |
| M2-R9 | Track deletion preserves sibling/published data | PASS |
| M2-R10 | Existing HTTP/generated-contract proof plus serializer and H38–H40 discrimination | **PASS ← PARTIAL** |
| M2-R11 | Legacy task terminal refusal | PASS |
| M2-R12 | Python boundary, configuration and persistence | PASS |
| M3-R1 | Component proof accepted; integrated fallback regression pending | PARTIAL |
| M3-R2 | Fixed-language Save proof accepted; browser network journey pending | PARTIAL |
| M3-R3 | Language/stale/mode guards | PASS |
| M3-R4 | PATCH→PUT/error cases accepted; browser reload/restore pending | PARTIAL |
| M3-R5 | Per-track regeneration, pending markers and polling | PASS |
| M3-R6 | Recovery and X-019 accepted; browser persistence pending | PARTIAL |
| M3-R7 | Component, keyboard and visual evidence | PASS |
| M3-R8 | Two-audio recognition and persisted-language journey | **NOT_TESTED** |
| M3-R9 | Updated integrated ASR journeys | **NOT_TESTED** |
| M3-R10 | Refusal/retry discrimination accepted; browser PATCH/503 cases pending | PARTIAL |
| M3-R11 | IDs/rename/token proof accepted; browser sequences pending | PARTIAL |
| M3-R12 | Attachment/flush proof accepted; video-replacement journey pending | PARTIAL |
| M3-R13 | Token/conflict proof accepted; two-editor browser conflicts pending | PARTIAL |
| M3-R14 | Metadata forwarding and visual proof accepted; upload→ASR chain pending | PARTIAL |
| M3-R15 | Guard/invalidation proof accepted; integrated mode-switch journey pending | PARTIAL |

Counts: **15 PASS, 10 PARTIAL, 2 NOT_TESTED, 0 MISSING**. The remaining browser portions belong to the pending final gate; mutation closure does not promote them to PASS.

**0 remaining code/evidence findings; 0 nonblocking findings; 1 pending ordered gate.**

**VERDICT (spec, pr109-final-3 @ c2f4310): SHIP-PENDING-GATE**