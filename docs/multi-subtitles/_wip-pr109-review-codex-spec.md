**BLOCK:** detached recovery has one confirmed specification mismatch. Required discrimination and visual evidence also remain incomplete.

Reviewed `37ec4b0...38c7915` independently, read-only. Root and HEAD matched before and after review; production files had no working-tree delta. The accepted plan hash matched `e27efa6a519ba330…`. The package’s report hash matches the committed report; its newer working-tree additions were not treated as pinned evidence.

**1. Findings**

**F1 — Major, BLOCKER, UNRESOLVED. Confidence: high.**  
[recovery.ts:232](/Users/strelia/projects/courses/frontend/src/modules/transcript/application/recovery.ts:232)

Spec, [plan-ms2-3.md:406](/Users/strelia/projects/courses/docs/multi-subtitles/plan-ms2-3.md:406):

> «відірваний знімок пропонується у вкладці доріжки, чия поточна мова = `snapshot.language`»

Then:

> «якщо такої немає … то у вкладці default-доріжки»

The implementation matches **all** language tabs and falls back to `matches[0]`, including an orphan `lang:*` tab.

Concrete case: current default track `en`, published orphan transcript `uk`, detached snapshot with `language: uk`. The code offers recovery in `lang:uk`; the specification requires the default track tab because no `uk` track exists. Restoring and saving consequently targets the orphan language instead of the specified destination.

Verified by tracing `lessonLanguageTabs` → `detachedTranscriptSnapshots` → `LanguageEditor`’s fixed save language. Existing placement tests use matching track tabs or a language absent from every tab; they omit the matching-orphan-only case. This affects M3-R6 and M3-R12.

Required correction: restrict the matching destination to track tabs, retain the default fallback, and add the discriminating orphan-only fixture.

**F2 — Verification BLOCKER, UNRESOLVED / NOT_TESTED. Confidence: high about the evidence gap.**  
[report.md:405](/Users/strelia/projects/courses/docs/multi-subtitles/report.md:405), [report.md:560](/Users/strelia/projects/courses/docs/multi-subtitles/report.md:560)

Spec, [plan-ms2-3.md:508](/Users/strelia/projects/courses/docs/multi-subtitles/plan-ms2-3.md:508):

> «390/768/1024/1440 + 200% + 320»

Required proof:

> «design-pass (браузер, скриншоти, DOM-замір) + unit»

The report claims viewport, geometry and focus results, but supplies no identifiable MS3 screenshots or raw DOM measurements in the pinned package. The image files found under this phase belong to MS1/F3. The aggregate e2e log does not prove these measurements.

The component test checks unique IDs, inert panels and accessible relationships; it cannot establish viewport fit, zoom behavior or 44px targets. M3-R7 and the visual portion of M3-R14 remain NOT_TESTED. This is an evidence finding, not a claim that the rendered interface fails.

Needed: revision-bound screenshots and raw geometry/focus results, or an authorized browser verification.

**F3 — Verification BLOCKER, UNRESOLVED / NOT_TESTED. Confidence: high.**  
[_wip-pr109-review-package.md:35](/Users/strelia/projects/courses/docs/multi-subtitles/_wip-pr109-review-package.md:35)

Spec, [plan-ms2-3.md:521](/Users/strelia/projects/courses/docs/multi-subtitles/plan-ms2-3.md:521):

> «Ordered `make check` … повний `make integration` … і повний e2e на фінальному коміті»

The additional review contract explicitly requires individual behavioral mutations, including inverse mutations for removed behavior: [behavior-coverage.md:10](/Users/strelia/project-data/ai-factory-kit/skills/arch/code-review/references/behavior-coverage.md:10).

The raw gate log supports the passing baseline. It does **not** establish that each changed guard, filter, serializer and invariant is discriminated. Historical spot-revert records exist, but they do not cover the complete inventory at the final snapshot; several recorded source/test hashes also precede later changes.

No mutations were executed in this review, as instructed. Required discrimination therefore remains NOT_TESTED rather than PASS. Closing this requires a coordinated, authorized disposable-copy run with individual mutation diffs, commands, attributable failures and restored baseline bytes.

**2. Scope, decisions and ADRs**

No unapproved product scope creep found. The Makefile reset repair is outside the original feature implementation but has a recorded owner authorization and addresses the required gate. Stopping rename-form submission propagation supports the specified rename journeys.

Source inspection supports these decisions:

- **O7:** existing drafts suppress automatic recognition; rename/mode changes produce staleness; Regenerate explicitly replaces them.
- **O8 / MS-D10:** persisted mode, environment defaults, API/media-worker wiring, engine flags and metadata-only master-rewrite suppression are implemented.
- **MS-D1–D5:** obsolete publication paths are removed; track-keyed persistence, queue identities, hint derivation and inventory endpoints are present.
- **MS-D6–D8:** stable track identities, fresh save checks, fresh segment IDs, token ownership, replacement confirmation and `und` PATCH→PUT are present.
- **MS-D9:** recovery keys and unmount capture are present; destination selection violates the decision as described in F1.
- **§6 limits:** the implementation does not silently restore the deliberately removed server-side provenance guarantees or expand into MS4.

Existing ADR checks: accepted ADRs **001, 002, 003, 006, 013 and 014** were inspected against this scope. No new behavioral contradiction found: module boundaries, transcript validation, React Aria interactions and module-owned access facts remain. ADR 006’s “No Python runtime now” is stale historical wording predating this PR.

Decision material checked: per-track identity, publication through PUT, recovery identity and multilingual configuration are explicitly recorded in dated MS-D decisions. The repository’s explicit new-ADR requirement covers contradictions of existing ADRs; none was identified here.

Doc integrity: the slice plan explicitly supersedes the older phase-plan publication/recovery design, so those older descriptions were not imposed as simultaneous requirements. Detailed process history remains in task evidence. Existing repository, transport, Query and recovery seams are reused; no separate speculative subsystem was identified.

**3. Evidence and behavior coverage**

I inspected the raw attempt-7 log: `GATE_EXIT=0`, 1186 frontend tests and 151 passing browser tests are recorded. Its header names `7b20d2c` plus working-file fingerprints. Final ASR and Makefile fingerprints match; the subsequent media TSX/test fingerprints also match their recorded package values. This supports reuse of that baseline evidence, rather than claiming I reran it.

CI configuration selects frontend unit tests and real Postgres integration tests. It deliberately excludes `asr.spec` through `E2E_SKIP_ASR`; local ASR execution is therefore distinct from CI execution. No actual PR CI result was supplied or independently verified.

The outstanding mutation inventory covers:

| Changed constructs | Production locations | Acceptance/tests mapped below |
|---|---|---|
| Backfill matching/deletion, language constraints, PK/FK/cascade, leaf uniqueness, Down survivor selection | Migration `00031` | M2-R1, R9 |
| Latest Ready/non-evicted selection, ready tracks, draft absence, removed default restriction | Postgres `draft.go` | M2-R2, R4, R7 |
| Source tuple, generation/readiness, requested replacement, refusal diagnosis | Postgres `Claim` | M2-R3, R4, R7 |
| Track/token completion fence and recovery timeout/identity | Postgres `Complete`/`Recover` | M2-R8 |
| Track snapshot validation and language/mode staleness | Transcript domain/application | M2-R6, R7, R12 |
| Task identity/payload, missing-track rejection, retry option | Asynq adapters | M2-R5, R11 |
| Supported-subtag hints, `und`, unsupported logging, multilingual Python arguments | Whisper adapters | M2-R6, R12 |
| Removed routes, body refusal, authorization, errors, DTO/inventory serialization | HTTP/OpenAPI/API adapters | M2-R10 |
| Environment parsing, optional mode semantics, persistence, rewrite conditions | Media/config/composition | M2-R12, M3-R14 |
| Tab filtering/order/orphans, loading/error preservation, identities and panel visibility | Admin language tabs | M3-R1, R2, R7 |
| Fresh reads, track matching, token branches, dialogs, PATCH→PUT failures | Language tabs/editor | M3-R4, R10, R11, R13 |
| Local draft compatibility, invalidation, per-track pending/expiry; removed automatic stale replacement | Draft domain/queries/UI | M3-R3, R5, R15 |
| Snapshot parsing, key isolation, destination/order/priority, base-token restoration | Recovery application/adapters/UI | M3-R6, R12, R13 |
| Capture flushing/suspension and version-conditional cleanup | Recovery hooks/stores/editor | M3-R4, R6, R12 |

All remaining individual mutation checks are **NOT_TESTED**. The table is an inventory of inspected behavior, not mutation execution evidence.

**4. Acceptance-row table**

“Implemented?” reflects source inspection. Status includes the required discrimination proof; a green baseline alone does not earn PASS. Test names below identify actual tests, sometimes shortened to their unique prefix.

Abbreviations: **PG** = `backend/internal/transcript/adapters/postgres/`; **Tabs** = `frontend/src/modules/admin/ui/LessonLanguageTabs.test.tsx`; **ASR** = `frontend/e2e/asr.spec.ts`.

| Row | Implemented? | Named test and observable oracle | Status |
|---|---|---|---|
| M2-R1 | Yes | PG `TestDraftRecoverAgainstPostgreSQL / M2-R1 migration 00031 maps populated drafts and round trips`: mapped/deleted rows, PK, survivors and cascade | NOT_TESTED |
| M2-R2 | Yes | Same suite, `M2-R2 … pending follows each ready track regardless of default` and selector test: exact pending identities/count | NOT_TESTED |
| M2-R3 | Yes | `M2-R3 C21 rejects foreign lesson asset track non-ready and generation`; `claim records selected track leaf language and mode`: refusal/no writes and stored snapshot | NOT_TESTED |
| M2-R4 | Yes | `M2-R4 C09 slow old completion stays hidden…`, `delayed old Ready track claim is superseded`, same-track fencing: visible text/source remains correct | NOT_TESTED |
| M2-R5 | Yes | Asynq `TestTwoTracksEnqueueDistinctTaskIDs`: exact distinct task IDs | NOT_TESTED |
| M2-R6 | Yes | Whisper `TestTranscribeDerivesHintFromTrack`, `TestLanguageCodesPinnedToFasterWhisper121`; application `TestProcessPassesClaimTrackSnapshot`: argv, log and forwarded snapshot | NOT_TESTED |
| M2-R7 | Yes | PG `automatic duplicate preserves stale output until requested`; `staleness follows track language and mode`: unchanged token/text, zero pending, explicit replacement | NOT_TESTED |
| M2-R8 | Yes | PG `concurrent re-Claim between Recover and Complete yields ErrDraftSuperseded`; `tracks recover and complete independently`: stale/cross-track token refusal | NOT_TESTED |
| M2-R9 | Yes | PG `deleting one track preserves sibling draft and published language`: deleted draft absent, sibling/published snapshot unchanged | NOT_TESTED |
| M2-R10 | Yes | HTTP `TestM2R10RegenerateHTTP`, `DraftsListHTTP`, `RemovedRoutesReturn404`, `TranscriptInventoryHTTP`, `OpenAPISurface`; PG inventory tests; generated drift gate | NOT_TESTED |
| M2-R11 | Yes | Asynq `TestLegacyTaskRefusedWithoutProcessingOrRetries`: invalid payload, processor not called, `MaxRetry(0)` | NOT_TESTED |
| M2-R12 | Yes | `TestRealTranscribePythonModelBoundary`; `TestASRMultilingualDefault`; media `TestM2R12MintFlagAndDefaultHTTP`, `PatchAndGetHTTP`; PG `multilingual metadata writes skip master rewrite`: real Python kwargs and persistence/rewrite counts | NOT_TESTED |
| M3-R1 | Yes | Tabs `tab order is default, embedded, language/id, then orphan…`; `no Ready asset keeps…`; unchanged transcript browser journeys | NOT_TESTED |
| M3-R2 | Yes | Tabs `save uses lowercased tab language, fresh reads, and new ids on every PUT`; ASR `M3-R8 R2 R4 success…`: captured PUT language/body | NOT_TESTED |
| M3-R3 | Yes | Tabs `draft guard refuses … with Regenerate instruction`: Load absent for stale/language/mode mismatch | NOT_TESTED |
| M3-R4 | Yes | Tabs `M3-R4` success/failure/dialog tests; ASR `PATCH succeeds and PUT fails, reload restores…`: operation order, zero writes on refusal, retained/restored text | NOT_TESTED |
| M3-R5 | Yes | Tabs `Regenerate receives the active track id…`; query `two tracks keep separate markers…`, `a stale ready draft has no refetch interval (O7)` | NOT_TESTED |
| M3-R6 | **Partial** | Recovery `legacy language matches…`, `gone tracks go to their language/default`; Tabs own/detached cleanup tests; ASR two-tab reload journey. Matching-orphan-only destination is omitted and wrong: F1 | PARTIAL |
| M3-R7 | Appears present | Tabs `mounted editors have unique ids and inactive inert panels…`, `tab keyboard navigation…`; raw viewport/zoom/target evidence missing: F2 | NOT_TESTED |
| M3-R8 | Yes | ASR `M3-R8 R2 R4 success and R6 two-track drafts…`: distinct Welcome/dubbed text, two languages, GET contents equal respective PUTs | NOT_TESTED |
| M3-R9 | Yes | ASR C06/C11/C17/C18 and S6 successors: current-track draft, no-body Regenerate, retained editor/recovery behavior | NOT_TESTED |
| M3-R10 | Yes | Tabs `fresh save check refuses …`; ASR foreign rename and GET-tracks-503 tests: zero PUT, text retained, retry works | NOT_TESTED |
| M3-R11 | Yes | Tabs fresh-ID/rename/token tests; ASR `M3-R11 a-d…`: disjoint IDs and preserved earlier-language contents; detached variant in R12b | NOT_TESTED |
| M3-R12 | **Partial** | Hook `flush on unmount writes exactly once…`; Tabs dirty-orphan attachment; ASR R12a/R12b. Destination with a matching orphan remains wrong: F1 | PARTIAL |
| M3-R13 | Yes | Tabs R13a–g tests; editor capture test; ASR open/restored T1 versus foreign T2: exact old token, 409, foreign server text and local text preserved | NOT_TESTED |
| M3-R14 | Yes in code | Audio-switch component tests; ASR `a track minted with the multilingual switch gets a ready draft`; Python boundary covered by M2-R12. Geometry evidence remains missing | NOT_TESTED |
| M3-R15 | Yes | Tabs `flag changes invalidate an already offered draft…`; audio-query invalidation tests; ASR `en`/`und` switch tests: Load removed, mode snapshot changes after Regenerate, editor text preserved | NOT_TESTED |

Acceptance totals: **2 PARTIAL, 25 NOT_TESTED, 0 fully proven PASS** under this review contract. This does not negate the passing baseline execution.

Severity counts: **1 Major behavior blocker; 2 verification blockers; 0 non-blocking findings. All 3 remain UNRESOLVED.**

**VERDICT (Spec, pr109-final @ 38c7915): BLOCK**