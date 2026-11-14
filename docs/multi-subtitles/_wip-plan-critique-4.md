`stage: plan` · Critic: Codex `/root`, selected independent context, round 4. Scope: D1–D8 phase decisions and delivery compatibility; O1–O6 retained as authorized.

1. **Blocker — OPEN: D5 introduces a deadlock with video publication.**

   Location: [plan.md:152](/Users/strelia/projects/courses/docs/multi-subtitles/plan.md:152).

   Counterexample:

   - Transcript publication locks generation 1, the current Ready asset, `FOR SHARE`.
   - Media publication locks processing generation 2 first through `lockMediaAsset` ([publish.go:32](/Users/strelia/projects/courses/backend/internal/media/adapters/postgres/publish.go:32)).
   - Transcript publication continues locking all lesson assets and waits for generation 2.
   - Media publication proceeds to its sibling locks and waits for generation 1 ([publish.go:43](/Users/strelia/projects/courses/backend/internal/media/adapters/postgres/publish.go:43)).

   Neither proceeds. PostgreSQL must abort a transaction. “Asset first” does not establish a common order across multiple asset rows; adding `ORDER BY generation` only to D5 would still conflict with media’s target-first acquisition.

   **Required plan change:** specify a compatible protocol across these writers, or an explicit bounded transaction-retry policy with defined outcomes. Extend the planned promotion race test to force this interleaving. Merely allowing one participant to fail does not prove the promised wait-and-revalidate behavior.

2. **Blocker — OPEN: the `und` exception still survives subsequent definite-language renames.**

   Locations: [staleness:121](/Users/strelia/projects/courses/docs/multi-subtitles/plan.md:121), [publication rule:147](/Users/strelia/projects/courses/docs/multi-subtitles/plan.md:147).

   The original competing-author counterexample is resolved: request `language=uk` refuses current track language `de`.

   **New counterexample retried with the same author:** draft A has `track_language=und`; O5 PATCH changes the track to `uk`; publication is interrupted. The author subsequently corrects the now-defined track to `de`, reopens its tab and publishes A with `language=de`.

   Both revised predicates pass:

   - current track language equals request language `de`;
   - A’s snapshot belongs to the `und` family.

   D4 also exempts A from staleness. Thus the same old attempt remains publishable after a definite→definite rename, contradicting D5’s stated stale→Regenerate behavior. Binding a request protects that request from intervening changes; it does not consume or persist the initial confirmation.

   **Required plan change:** define durable confirmation binding or an equivalent rule distinguishing the initial O5 transition from later renames. Preserve the legitimate interrupted `und`→`uk` publication retry. Add this same-author sequence to planned acceptance.

3. **Nonblocking — ACTIONABLE in MS2: specify replay precedence after metadata changes.**

   Location: [plan.md:163](/Users/strelia/projects/courses/docs/multi-subtitles/plan.md:163).

   Atomic publication receipts now provide a credible mechanism for unchanged lost-response retries. Outcomes after manual transcript editing and Regenerate are explicit.

   One remaining ambiguity: publication under `uk` succeeds, its response is lost, the track becomes `de`, then the original request is retried. The receipt rule promises success if the published transcript token remains unchanged; the mandatory language validation refuses it.

   MS2 should state which outcome takes precedence, bind receipt lookup to the publication’s language, and preserve the attempt token when recording its receipt. Test unchanged replay, manual edit, Regenerate and metadata-change replay. This detail can remain in the independently critiqued MS2 slice plan.

Explicit round-3 dispositions:

| R3 finding | Disposition | Counterexample retried against r4 |
|---|---|---|
| **1 — confirmed language** | **Partially resolved** | Original `uk` request versus concurrent `de` refuses correctly. A fresh `de` request by the same author still reuses the indefinitely exempt `und` attempt: finding 2. |
| **2 — publication consistency** | **Partially resolved** | Asset/track locks now conflict with rename, explicit deletion and Ready promotion. Target-first media publication versus D5’s multi-asset acquisition creates finding 1. Atomic receipts and generation-of-draft stamping are retained. |
| **3 — legacy client compatibility** | **Resolved at phase level** | MS2+MS3 explicitly share one PR and ordered gate; no backend-only merge or unsafe compatibility alias remains ([plan.md:241](/Users/strelia/projects/courses/docs/multi-subtitles/plan.md:241)). |
| **4 — D8 election/command** | **Resolved at phase level** | `und` plus `und-x-original` elects exactly one deterministically. Evicted assets are excluded; command failure fails the flow; successful regeneration precedes MS4 activation ([plan.md:212](/Users/strelia/projects/courses/docs/multi-subtitles/plan.md:212)). Command syntax and detailed proofs remain MS1 work. |

The writer inspection supports the locking mechanism, with narrower claims than the plan currently makes:

- **Rename and explicit track deletion:** asset `FOR UPDATE` precedes track mutation; both conflict with D5’s asset share lock.
- **Embedded materialisation and Ready promotion:** `Publish` holds asset locks before updating Ready state and inserting the embedded track. Uploaded-track promotion also locks its asset before tracks.
- **Eviction and asset deletion:** the asset `UPDATE`/`DELETE` itself conflicts with `FOR SHARE`, even without a separate explicit lock statement.
- **Track-only writers exist:** append, cancel and audio sweep do **not** acquire an asset lock. Their track mutations conflict with D5’s track share lock. Their inspected predicates concern incomplete/failed tracks, so they do not establish another bypass of a Ready track’s language or readiness.
- Historical embedded backfill uses **table locks**, not the asserted universal asset-row protocol. `CreateAudioTrack` has no production caller found. These are not additional blockers.

**Verdict: BLOCKED — D1–D8 phase plan, with two open substantive blockers affecting D4/D5 and dependent MS2+MS3.** D8’s revised architecture and joint delivery resolve their prior findings. Previously closed findings remain closed; historical test results are retained only for their recorded snapshots.

Binding:

- Plan: untracked `docs/multi-subtitles/plan.md`, **r4**, SHA-256 `640585b20d8f24e01ec78bd053f50886b7e9f4091086acd108f871f4ea992b90`, unchanged on final reread.
- Target: `/Users/strelia/projects/courses`, branch `multi-subtitles`, HEAD `5cbb44bb6d0f9888750f008e50e4cca1f2093c1b`. Tracked backend/frontend files are clean. Unrelated documentation/configuration WIP and untracked `backend/upload` were excluded.
- Raw reviews read: r1 `fbff4e427ae41732a8411c05b9a44387980b325925f45617f0a12e38e615fb87`; r2 `201eaf8b9ade2542d3e63b7d1125b62e5421f4fd295f54b444ea53473bb37c41`; r3 `61c579a0c41011810ba54a8e41c1df10a7581d211cc95493822c790103a8e0ed`.
- Prompt: `prompts/plan-critic.md`, **version 1**, SHA-256 `514e4c376c2551afe76e390c5779867ecf7f04ad2a552c97c2b4f713c43dcab6`, under candidate directory `/Users/strelia/project-data/ai-factory-workspace/runtime-data/ai-factory-rollout/candidates/17a9527421e8fac3e43a16c287c404118699833c`. Prompt and pinned skill hashes match current package metadata; release commit authenticity remains unverified.
- Sources: revised plan, all raw prior reviews, repository briefing/rules/adapter, ADR 013, media writers and callers, migration/backfill SQL, transcript claim/save/publication/HTTP paths and gate recipes. Conclusions use file evidence; concurrency schedules were reasoned, not executed.

No tests, product writes, network probes or graph refresh ran. Future gates remain **NOT_RUN**.

The stronger [four-lens skill](/Users/strelia/project-data/ai-factory-workspace/runtime-data/ai-factory-rollout/candidates/17a9527421e8fac3e43a16c287c404118699833c/skills/plan/plan-critique/SKILL.md) remains **NOT_RUN: 0 launched / 0 returned**; it requires separate clean-context agents, while this request retains the existing reviewer without spawning. This review does not claim that gate passed.

The record could not be saved or linked from the plan because the workspace is read-only.