`stage: plan` · Critic: Codex `/root`, selected independent context, round 5. Scope: D1–D8 phase plan; O1–O6 retained as authorized.

1. **Blocker — OPEN: single-asset locking misses uploaded-audio promotion of a replacement video.**

   Location: [plan.md:169](/Users/strelia/projects/courses/docs/multi-subtitles/plan.md:169).

   The assertion that every Ready promotion locks the old asset is false. Video `Publish` locks siblings, but `PromoteAudioTrack` locks its **own** asset and tracks, then changes `awaiting_audio` to `ready` ([audiotrack.go:370](/Users/strelia/projects/courses/backend/internal/media/adapters/postgres/audiotrack.go:370), [audiotrack.go:419](/Users/strelia/projects/courses/backend/internal/media/adapters/postgres/audiotrack.go:419)). It touches another asset only when eviction is necessary.

   **Concrete missed-change interleaving:**

   - A, generation 1, is the lesson’s sole Ready asset. B, generation 2, is `awaiting_audio`.
   - Transcript publication locks A, its track and draft; validates that A is latest Ready.
   - B’s uploaded-track promotion locks B and its tracks, makes B Ready and commits. Ready count is two: no eviction, no lock on A.
   - Transcript publication commits A’s text after B became latest Ready.

   This violates D5’s promised serialization: promotion neither waits nor causes `draft_superseded`. Stamping generation 1 prevents false freshness, but does not establish the required latest-Ready publication guard.

   **Required plan change:** define compatible serialization covering **both** Ready-promotion paths. Include `PromoteAudioTrack(awaiting_audio→ready)` explicitly in the forced-interleaving tests, including the no-eviction case. Do not restore multi-asset acquisition without checking its order against existing writers.

No additional nonblocking finding.

Round-4 dispositions:

| R4 finding | Disposition | Counterexample retried against r5 |
|---|---|---|
| **1 — deadlock** | **Original cycle resolved; consistency partially resolved** | Publication holds only A and never requests B, removing the A→B/B→A cycle with video `Publish`. Uploaded-track promotion bypasses A altogether in the schedule above: the remaining blocker. |
| **2 — `und` snapshot** | **Resolved — rebuttal accepted** | `und` claim → PATCH `uk` → interrupted publish → PATCH `de` → publish(`de`) legitimately reuses the reviewed, unhinted recognition. Later labels did not influence that recognition; O5 permits the author to correct its language. The sequence itself establishes neither wrong text nor wrong language shown to a student. Publishing the original `uk` request after a concurrent `de` rename still refuses. A defined `en` snapshot renamed to `de` also refuses. Durable consumption of the initial confirmation is unnecessary for the stated acceptance. |
| **3 — receipt precedence** | **Resolved at phase level** | Publish under `uk`, lose response, rename to `de`, retry original `(draftToken, uk)`: receipt lookup now precedes language/superseded checks and returns success without another write when the published transcript token is unchanged. Manual editing returns `transcript_modified`; regeneration returns `draft_modified`. MS2 must preserve the reviewed attempt token when recording the receipt. |

Writer-by-writer inspection of revised D5:

| Writer | Deadlock assessment against D5 | Missed-change assessment |
|---|---|---|
| Rename | Asset-first acquisition conflicts before track mutation; no D5 cycle found. | Lock plus post-wait validation covers language changes. |
| Explicit track deletion | Asset before tracks/cascade; no D5 cycle found. | Deletion waits, or publication observes the missing track. |
| Video Ready promotion, including embedded materialisation | Publication no longer requests the target asset while holding its old asset; original cycle removed. | Sibling locks cover A before Ready change and embedded insertion. |
| Uploaded-track promotion | No D5 cycle found in the inspected acquisition path. | **Uncovered:** B can become Ready without touching A. Promotion on A itself conflicts correctly. |
| Eviction | Updating the affected asset conflicts with publication’s share lock; no D5 cycle found. | Protects eviction of A; eviction of another asset does not protect A’s latest-Ready predicate. |
| Asset deletion | Asset deletion precedes cascading track/draft deletion; no D5 cycle found. | Conflicts on A, or publication observes absence after waiting. |
| Append / cancel / audio sweep | Track-only paths have no subsequent asset acquisition creating a D5 cycle. Append’s inspected callbacks perform storage operations. | Their mutation predicates exclude Ready tracks; they do not bypass the inspected Ready-track language/existence checks. |

These are file-based concurrency analyses, not executed tests.

**Verdict: BLOCKED — D1–D8 phase plan, with one open substantive blocker affecting D5 and dependent MS2+MS3.**

Since round 4, the original deadlock is removed, the `und` rebuttal is accepted, and receipt precedence is explicit. Previously closed findings remain closed. Unchanged schema, module-boundary and coordinated-delivery evidence is retained; historical test results remain bound to their original snapshots.

Binding and limitations:

- Plan: untracked `docs/multi-subtitles/plan.md`, **r5**, SHA-256 `e697697dfae16ef739685d55870ffb4d3d7d42c925837ac54cd814045fde1fe1`, unchanged on final reread.
- Target: `/Users/strelia/projects/courses`, branch `multi-subtitles`, HEAD `5cbb44bb6d0f9888750f008e50e4cca1f2093c1b`. Tracked backend/frontend sources are clean. Untracked `backend/upload` and unrelated documentation/configuration WIP were excluded.
- Raw rounds 1–4 read; r4 SHA-256 `1435f261e217d81e35cea3651ac6f0d2f5f5ddd4d71237017924cb08f86a6a9b`. R1–R3 hashes match those recorded in r4.
- Prompt: `prompts/plan-critic.md`, **version 1**, SHA-256 `514e4c376c2551afe76e390c5779867ecf7f04ad2a552c97c2b4f713c43dcab6`, under candidate directory `/Users/strelia/project-data/ai-factory-workspace/runtime-data/ai-factory-rollout/candidates/17a9527421e8fac3e43a16c287c404118699833c`. Package declares `0.8.0-alpha.1+native-handoff.1`; release commit authenticity remains unverified.
- Inspected revised plan, raw prior findings, repository briefing/standards/adapter, ADR 013, media writers and application callers, transcript persistence/recognition, migration SQL and gate recipes. No graph claims used.
- No tests, product writes, network probes or graph refresh ran. Future gates remain **NOT_RUN**.
- Stronger four-lens gate: **NOT_RUN — 0 launched / 0 returned**, retaining the requested existing-reviewer route without spawning.

The review could not be saved or linked from the plan because the workspace is read-only.