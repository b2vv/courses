`stage: plan` · Critic: Codex, selected independent session, round 3. Scope: D1–D8 phase decisions and slice compatibility; O1–O6 treated as authorized.

1. **Blocker — OPEN: O5 publication is not bound to the language the author confirmed.**

   Location: [plan.md:137](/Users/strelia/projects/courses/docs/multi-subtitles/plan.md:137), [plan.md:143](/Users/strelia/projects/courses/docs/multi-subtitles/plan.md:143).

   **Counterexample retried:** draft A was claimed under `und`, detected `en`; author confirms `uk` and PATCH succeeds. Another author PATCHes the track to `de`. Publication of A passes both current guards: its draft token still matches, and snapshot `und`→current definite `de` is permitted. It publishes under `de`, although the author confirmed `uk`.

   The same snapshot remains exempt after `und`→`uk`→`de`; the exception does not distinguish initial confirmation from subsequent definite-to-definite renaming. Conversely, a snapshot originally captured as `en-us` correctly rejects current `en-gb`: that counterexample is resolved.

   **Required change:** carry the confirmed target language in publication and compare it atomically with current metadata. Define when the initial unknown-language exception is consumed or bound to that confirmation. Merely comparing against detected language would contradict O5, which authorizes changing the suggestion.

   The ordinary, uncontended `und`→author-selected `uk` path **is consistent** with O5 even when detection returned `en`.

2. **Blocker — OPEN: locking only the draft does not establish D5’s publication consistency.**

   Location: [plan.md:145](/Users/strelia/projects/courses/docs/multi-subtitles/plan.md:145).

   **Counterexample retried:** publication locks draft A and reads track `en` and generation 1. A concurrent PATCH commits `de`, or a replacement becomes Ready, before the transcript write commits. Reading these facts within one transaction does not itself prevent either change.

   Existing metadata mutation locks the asset and updates the track ([audiotrack.go:271](/Users/strelia/projects/courses/backend/internal/media/adapters/postgres/audiotrack.go:271)); video publication locks asset rows ([publish.go:43](/Users/strelia/projects/courses/backend/internal/media/adapters/postgres/publish.go:43)). Neither coordinates with the proposed draft-only lock.

   Stamping **draft generation 1** fixes the previous false-freshness defect: it will not falsely stamp generation 2. However, the plan still lacks a serialization rule establishing its language and latest-Ready checks through publication.

   **Required change:** specify the shared locking/isolation protocol and lock order with rename, deletion and Ready promotion, including revalidation after waiting. Add planned concurrency checks for those writers.

   Regenerate replacing A with B is now correctly refused by `draftToken`; concurrent transcript writes retain the expected-transcript-token guard. Lost-response tokens provide a credible mechanism, but MS2 must specify their atomic persistence with publication and behavior after regeneration or a later manual transcript edit.

3. **Blocker — OPEN: MS2’s compatibility alias cannot supply the new reviewed-attempt contract.**

   Location: [plan.md:220](/Users/strelia/projects/courses/docs/multi-subtitles/plan.md:220).

   **Counterexample retried:** the old client reads default-track draft A. Default changes, or Regenerate produces B. Its lesson-scoped publish request contains only `expectedUpdatedAt`, as the existing transport shows ([draft_handler.go:116](/Users/strelia/projects/courses/backend/internal/transcript/adapters/http/draft_handler.go:116)).

   Resolving “default track” at publish time can target another track. Resolving the current draft token server-side can publish B instead of reviewed A. Requiring the new fields breaks the promised unchanged frontend. There is also no old-client O5 confirmation step for an unknown-language track.

   **Required change:** choose an explicit intermediate delivery contract: introduce the necessary client binding in MS2, coordinate delivery, or document a narrowly authorized temporary legacy contract and its limitations. A route alias alone does not resolve this.

4. **Nonblocking — ACTIONABLE for MS1: D8 needs explicit collision election and regeneration completion rules.**

   Location: [plan.md:197](/Users/strelia/projects/courses/docs/multi-subtitles/plan.md:197).

   **Counterexample tried:** a Ready asset contains both `und` and `und-x-original`, with primary `uk` and no existing `uk` track. Both initially satisfy the migration predicate; both cannot become `uk`. Specify deterministic election and record the other as skipped. Also exclude evicted assets from the regeneration set.

   The architectural choices are sound:

   - Cross-table SQL reads fit the repository’s migration precedent: [00015:32](/Users/strelia/projects/courses/backend/migrations/00015_enrollment.sql:32) reads activity across modules; [00028:15](/Users/strelia/projects/courses/backend/migrations/00028_audio_track_embedded_backfill.sql:15) coordinates asset/track writers. No module-package import is required.
   - A command mode wired in `cmd/media-worker` fits existing composition roots. It should terminate without starting the worker’s recovery, sweep, queue and HTTP lifecycle ([main.go:60](/Users/strelia/projects/courses/backend/cmd/media-worker/main.go:60)).
   - Reusing the **repository’s serialized regeneration**, not just the filesystem writer, makes retries safe against concurrent metadata changes ([audiotrack.go:297](/Users/strelia/projects/courses/backend/internal/media/adapters/postgres/audiotrack.go:297)). The writer already performs an atomic, repeatable rewrite.

   A stale `und` manifest cannot select another language under D1. During MS1–MS3, the old player still uses its existing transcript read. Nevertheless, MS1 should make command failure visible and require successful regeneration before MS4 activation; otherwise O6’s caption-preservation result remains incomplete. Detailed command syntax and retry evidence reasonably belong in `plan-ms1.md`.

Explicit round-2 dispositions:

| R2 finding | Disposition | Counterexample retried against r3 |
|---|---|---|
| **1 — language snapshot/staleness** | **Partially resolved** | Unchanged `en-US` versus detected `en` now works; `en-US`→`en-GB` becomes stale. A snapshot captured as `und` remains indefinitely exempt from later definite-language changes: finding 1 above. |
| **2 — publication consistency/retry** | **Partially resolved** | A→B regeneration is bound by `draftToken`; stamping uses A’s generation; lost-response markers are specified. Metadata/replacement serialization and O5 confirmation binding remain open: findings 1–2. |
| **3 — unauthorized unknown-language policy** | **Resolved** | Existing `und` + Ukrainian primary is addressed by authorized O6/D8; new unknown-language publication requires authorized O5 confirmation. The explicit suppression of `und` transcripts is now an owner-authorized policy. |
| **4 — normalized language cardinality** | **Resolved at phase level** | `en`/`EN` collisions are diagnosed; failed duplicates do not create Ready tabs. D7 now normalizes remembered preferences. |
| **5 — historical recovery** | **Resolved at phase level** | A snapshot lacking language is no longer assigned today’s primary. MS3 explicitly owns ambiguity, destination collisions, repeat migration and O5 key movement. Its algorithm remains subject to slice critique. |
| **6 — player details and delivery order** | **Partially resolved** | Outward player notification, native initial/change/removal handling and the switching-time measurement are explicit. Timing choice is legitimately deferred to MS4. MS2 compatibility remains blocked by finding 3. |

**Verdict: BLOCKED — D1–D8 phase plan, with three open substantive blockers.** The blockers affect D4/D5 and MS2 delivery; they do not reject O1–O6 or D8’s architecture.

Binding and evidence:

- Plan: untracked `docs/multi-subtitles/plan.md`, **r3**, SHA-256 `86c12e09f8cbf604625ee3c95caf8b2250fb199cbc9eda52362e7a4f5cdd2dd4`, unchanged on final reread.
- Target: `/Users/strelia/projects/courses`, branch `multi-subtitles`, HEAD `5cbb44bb6d0f9888750f008e50e4cca1f2093c1b`. Tracked backend/frontend sources are clean. Untracked `backend/upload` was excluded and not executed; unrelated documentation/configuration WIP was excluded.
- Raw prior reviews read: r1 SHA-256 `fbff4e427ae41732a8411c05b9a44387980b325925f45617f0a12e38e615fb87`; r2 `201eaf8b9ade2542d3e63b7d1125b62e5421f4fd295f54b444ea53473bb37c41`.
- Prompt: `prompts/plan-critic.md`, **version 1**, SHA-256 `514e4c376c2551afe76e390c5779867ecf7f04ad2a552c97c2b4f713c43dcab6`, read from candidate release directory `/Users/strelia/project-data/ai-factory-workspace/runtime-data/ai-factory-rollout/candidates/17a9527421e8fac3e43a16c287c404118699833c`. Release authenticity was not independently verified.
- Inspected plan, both raw reviews, repository briefing/workflow/standards, ADR 013, transcript persistence/publication/HTTP, media metadata/publication/master generation and composition root. File evidence was used; no graph freshness claim was made. Historical S6/S7 results remain evidence only for their recorded snapshots.
- No tests, product writes, network probes or graph refresh ran. Listed future gates remain **NOT_RUN** and require their separately permitted execution window.

The stronger four-lens gate remains **NOT_RUN: 0 launched / 0 returned**. The [skill](/Users/strelia/.agents/skills/plan-critique/SKILL.md) requires separate clean-context agents; this request preserves the existing reviewer without spawning. Its referenced `reference-anti-patterns.md` was unavailable.

This review could not be saved or linked from the plan because the workspace is read-only.