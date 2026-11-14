`stage: plan` · Critic: Codex `/root`, selected independent context · Scope: **multi-audio S6**

Plan: `docs/multi-audio/plan-s6.md`, revision 2026-11-14, SHA-256:
`470e327cef48d6de6d144c257e3d61c3d8ea0f6beb772a5c8a36856688672907`.

Snapshot: `/Users/strelia/projects/courses`, branch `multi-audio-s6`, HEAD `ef48c150fbfa093f0168efce02c3b6e15d1c8f3f`. Target plan is untracked; inspected tracked product sources and parent plans have no changes against HEAD. Unrelated dirty files and historical untracked WIP were excluded.

Prompt: **Plan critic — version 1**, `prompts/plan-critic.md` under `/Users/strelia/project-data/ai-factory-workspace/runtime-data/ai-factory-rollout/candidates/17a9527421e8fac3e43a16c287c404118699833c`, SHA-256 `514e4c376c2551afe76e390c5779867ecf7f04ad2a552c97c2b4f713c43dcab6`. Package metadata reports `0.8.0-alpha.1+native-handoff.1`; current package commit remains unverified.

Four lenses applied in this context; **0 launched / 0 returned** additional reviewers, as instructed. A formal four-agent gate is not claimed.

1. **BLOCKING · Logical gaps / edge cases · OPEN — polling can schedule duplicate replacements that overwrite fresh work.**

   [Plan:42](/Users/strelia/projects/courses/docs/multi-audio/plan-s6.md:42) relies on changing `drafts.updated_at` to avoid retained task-ID conflicts, but does not distinguish an obsolete automatic task from a necessary regeneration.

   Reachable sequence: recognition of A is running; default changes to B; polling enqueues replacement with A’s claim timestamp. A completes, changing the timestamp. While that replacement waits behind other jobs, another poll enqueues a second replacement with the completion timestamp. Both tasks later read B and run.

   [Claim:125](/Users/strelia/projects/courses/backend/internal/transcript/adapters/postgres/draft.go:125) unconditionally resets same-generation drafts; [Process:60](/Users/strelia/projects/courses/backend/internal/transcript/application/recognize.go:60) does not pass the task marker into Claim. Same-generation reclaim is explicitly required by the [existing test:337](/Users/strelia/projects/courses/backend/internal/transcript/adapters/postgres/draft_integration_test.go:337).

   **Impact:** redundant expensive recognition and possible loss of corrections made after the first replacement completes.

   **Required plan edit:** define how obsolete automatic deliveries are rejected while preserving intentional manual regeneration and recovery. Add a deterministic test covering polling before/after completion with both replacements queued.

2. **BLOCKING · Logical gaps · OPEN — D5’s invalidation behavior is unresolved.**

   Parent [D5:85](/Users/strelia/projects/courses/docs/multi-audio/plan.md:85) requires default changes to invalidate and rerun. The S6 plan changes pending selection, but leaves the observable stale contract unspecified.

   [GetDraft:84](/Users/strelia/projects/courses/backend/internal/transcript/application/draft.go:84) derives `stale` exclusively through [generation comparison:146](/Users/strelia/projects/courses/backend/internal/transcript/domain/draft.go:146). Consequently, after changing default, the previous ready draft still reports `stale=false` until replacement claims it—and an old running attempt can complete during that interval.

   **Required plan edit:** explicitly define invalidation for the draft and the retained published transcript, including the interval before replacement starts. Either specify the affected read/response behavior and tests, or record an authorized narrowing of D5. Preserving published text does not by itself resolve invalidation semantics.

3. **NON-BLOCKING · Logical gaps · OPEN — the 60-second rerun bound is unsupported.**

   [Plan:48](/Users/strelia/projects/courses/docs/multi-audio/plan-s6.md:48) promises rerun delay ≤60 seconds. That is the polling interval, while [worker concurrency is one:72](/Users/strelia/projects/courses/backend/cmd/asr-worker/main.go:72), and [pending selection is capped at 50:19](/Users/strelia/projects/courses/backend/internal/transcript/application/recognize.go:19).

   **Plan edit:** describe polling cadence separately from queue wait and recognition completion; do not promise a start-time bound without backlog assumptions.

4. **NON-BLOCKING · Evidence accuracy · OPEN — D8-15 does not authorize regeneration overwrite.**

   [Plan:62](/Users/strelia/projects/courses/docs/multi-audio/plan-s6.md:62) cites D8-15 for losing unpublished corrections. [D8-15:1159](/Users/strelia/projects/courses/docs/asr/spec.md:1159) concerns deleting media and cascading deletion of its draft.

   **Plan edit:** cite the applicable regeneration decision, or state this extension as its own dated decision. Existing Claim behavior supports the mechanism, but the cited owner decision concerns a different action.

AI-slop and unnecessary-complexity lenses: **clean**—one provenance column, existing polling and the existing extraction adapter fit this bounded change. No new framework or event system is needed.

Inspected parent D5/§5/§7/§8, S-backfill plan and raw critiques/reviews, migration 00028, media default/delete writers, transcript claim/completion/recovery/manual regeneration, worker/queue/extraction paths and ASR browser tests. The backfill supports removing the temporary fallback; its earlier exclusions remain intact. This is the first S6 verdict; no predecessor’s acceptance transfers to S6.

No files changed; tests, network probes and graph refresh were **NOT RUN**. Planned gates require a separately authorized execution window.

**Open findings: 2 blocking, 2 non-blocking.**

VERDICT: BLOCKED