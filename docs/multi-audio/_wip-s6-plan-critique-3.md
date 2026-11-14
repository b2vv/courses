`stage: plan` · Critic: Codex `/root`, selected independent context · Round 3 · Scope: **multi-audio S6**

Plan: `docs/multi-audio/plan-s6.md`, revision 2026-11-14, SHA-256:
`75492d10cfee8a8dbef288edc63fc307f4af0c61aec9e8b798b924bc4fba02c1`.

Snapshot: `/Users/strelia/projects/courses`, branch `multi-audio-s6`, HEAD `ef48c150fbfa093f0168efce02c3b6e15d1c8f3f`. Plan and both raw critiques are untracked. Inspected tracked backend/frontend sources, parent plan and AGENTS have no differences against HEAD. Unrelated dirty configuration/documentation and other untracked artifacts are excluded.

Prompt: **Plan critic — version 1**, release directory `/Users/strelia/project-data/ai-factory-workspace/runtime-data/ai-factory-rollout/candidates/17a9527421e8fac3e43a16c287c404118699833c`, path `prompts/plan-critic.md`, verified SHA-256:
`514e4c376c2551afe76e390c5779867ecf7f04ad2a552c97c2b4f713c43dcab6`.
Package metadata: `0.8.0-alpha.1+native-handoff.1`; package commit remains unverified.

**Findings and dispositions**

1. **F7 — BLOCKING · Logical gaps / edge cases · OPEN, actionable: first automatic recognition can remain invisible in the mounted editor.**

   [Plan:88](/Users/strelia/projects/courses/docs/multi-audio/plan-s6.md:88) adds draft invalidation when the audio terminal signature changes, explicitly covering the first default on a video-only asset. However, [plan:90](/Users/strelia/projects/courses/docs/multi-audio/plan-s6.md:90) extends polling only for `stale=true`.

   Reachable sequence:

   - A mounted video-only lesson has **no draft**.
   - Its first uploaded track becomes Ready/default; the asset becomes Ready in the [promotion transaction:403](/Users/strelia/projects/courses/backend/internal/media/adapters/postgres/audiotrack.go:403).
   - Terminal-signature invalidation rereads the draft **before the next ASR poll**.
   - The [no-draft response:73](/Users/strelia/projects/courses/backend/internal/transcript/adapters/http/draft_handler.go:73) has `state=null`, `stale=false`. Existing [polling:114](/Users/strelia/projects/courses/frontend/src/app/composition/admin-draft-queries.ts:114), even with the proposed stale condition, stops.
   - ASR later creates/completes the draft without changing the audio terminal signature. The [video query:88](/Users/strelia/projects/courses/frontend/src/app/composition/admin-video-queries.ts:88) also supplies no invalidation for the first observed Ready generation.

   **Impact:** automatic recognition succeeds in the backend, but the mounted editor continues showing no draft until another refetch trigger.

   **Required plan edit:** define the waiting lifecycle for an eligible lesson whose first draft does not exist yet, including termination/failure behavior. Add a deterministic mounted-query/component test and extend the video-only e2e: no draft → first default → initial reread still empty → automatic result, without reload or manual Regenerate.

2. **F5 — NARROWED: the reported terminal-draft replacement scenario is resolved; first-draft delivery remains open as F7.**

   [Plan:83](/Users/strelia/projects/courses/docs/multi-audio/plan-s6.md:83) now connects set-default/delete invalidation with polling through stale terminal responses. That closes ready A → default B → stale → running/result. Existing [LessonDraftField:121](/Users/strelia/projects/courses/frontend/src/modules/transcript/ui/LessonDraftField.tsx:121) can render the banner; loading replacement text remains explicit. Planned frontend tests cover both mutations and polling termination.

3. **F6 — RESOLVED, previously non-blocking.**

   [Plan:45](/Users/strelia/projects/courses/docs/multi-audio/plan-s6.md:45) accurately describes `MaxRetry(0)`, expired-lease archival and application recovery to `failed/timeout`. Verified against [queue:47](/Users/strelia/projects/courses/backend/internal/transcript/adapters/asynq/queue.go:47), installed asynq `recoverer.go:98`, and [Recover:191](/Users/strelia/projects/courses/backend/internal/transcript/application/recognize.go:191). The `running` exception is removed; [plan:125](/Users/strelia/projects/courses/docs/multi-audio/plan-s6.md:125) adds payload round-trip coverage, including legacy missing fields.

4. **Earlier resolutions retained.**

   **F1:** automatic/manual distinction and conditional Claim remain specified; generation monotonicity and token-guarded completion remain preserved. **F2:** backend leaf-based staleness and the recorded owner decision concerning published transcripts remain resolved; frontend delivery has the F7 qualification above. **F3:** polling cadence remains separated from queue delay. **F4:** overwrite policy remains a dated author decision rather than a false D8-15 attribution.

Four lenses were applied within this selected context: AI-slop and unnecessary complexity are **clean** because the plan reuses existing polling, extraction and composition mechanisms; logical gaps and edge cases produced F7. **0 launched / 0 returned** additional reviewers, as instructed; a formal four-agent gate is not claimed. The skill’s referenced `reference-anti-patterns.md` was unavailable at the referenced local/package locations.

Inspected both raw critiques, parent D5, backfill plan/report/migration, transcript pending/Claim/Complete/Recover/GetDraft, queue encoding/decoding, worker wiring, extraction, audio promotion/default/delete writers, frontend queries and editor components. The briefing predates this phase; current files supplied the relevant evidence.

Compared with round 2, terminal-draft delivery and retry documentation have concrete remedies. The newly exposed gap concerns **absent initial drafts**, which stale polling cannot observe. Tests, network probes, graph refresh and writes were **NOT RUN**. Planned ordered gates require a separately authorized execution window.

**Open product findings: 1 blocking, 0 non-blocking.**

VERDICT: BLOCKED