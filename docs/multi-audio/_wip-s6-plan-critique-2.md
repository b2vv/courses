`stage: plan` · Critic: Codex `/root`, selected independent context · Round 2 · Scope: **multi-audio S6**

Plan: `docs/multi-audio/plan-s6.md`, revision 2026-11-14, SHA-256:
`6e79f2336facc10499807b4a18af750462e9efcc4436acad1dbe6600f537ff78`.

Snapshot: `/Users/strelia/projects/courses`, branch `multi-audio-s6`, HEAD `ef48c150fbfa093f0168efce02c3b6e15d1c8f3f`. Plan and round-one critique are untracked. Inspected tracked backend/frontend sources, parent plan and AGENTS have no changes against HEAD. Unrelated dirty documentation/configuration and historical untracked WIP are excluded.

Prompt: **Plan critic — version 1**, `prompts/plan-critic.md` under `/Users/strelia/project-data/ai-factory-workspace/runtime-data/ai-factory-rollout/candidates/17a9527421e8fac3e43a16c287c404118699833c`; SHA-256:
`514e4c376c2551afe76e390c5779867ecf7f04ad2a552c97c2b4f713c43dcab6`.
Package version: `0.8.0-alpha.1+native-handoff.1`; package commit remains unverified.

Four lenses applied within this selected context: **0 launched / 0 returned** additional reviewers. A formal four-agent gate is not claimed.

**Round-one dispositions**

1. **F1 — RESOLVED for the reported sequence.** [Plan:41](/Users/strelia/projects/courses/docs/multi-audio/plan-s6.md:41) distinguishes automatic work from manual regeneration and rejects automatic claims against an already current terminal draft. This prevents the second queued replacement from erasing the first replacement’s corrections. Existing token-guarded [Complete:205](/Users/strelia/projects/courses/backend/internal/transcript/adapters/postgres/draft.go:205) remains appropriate. The existing generation monotonicity guard must remain intact.

2. **F2 — NARROWED; backend contract resolved, observable editor behavior remains open below.** [D-S6-6:71](/Users/strelia/projects/courses/docs/multi-audio/plan-s6.md:71) records the owner’s published-transcript decision and specifies leaf-based draft staleness. This supplies the missing inputs for `GetLessonDraft` and `Draft.Stale`. It does not establish how the mounted editor receives the changed value.

3. **F3 — RESOLVED.** [Plan:59](/Users/strelia/projects/courses/docs/multi-audio/plan-s6.md:59) separates polling cadence from queue delay and removes the unsupported start-time guarantee.

4. **F4 — RESOLVED.** [Plan:88](/Users/strelia/projects/courses/docs/multi-audio/plan-s6.md:88) replaces the incorrect D8-15 attribution with a dated author decision and an explicit overwrite mechanism.

**New findings**

5. **BLOCKING · Logical gaps / edge cases · OPEN, actionable — the stale banner and automatic replacement are invisible to a mounted terminal-draft editor.**

   [Plan:76](/Users/strelia/projects/courses/docs/multi-audio/plan-s6.md:76) assumes that returning `stale=true` makes the existing banner appear. However:

   - [Draft polling:114](/Users/strelia/projects/courses/frontend/src/app/composition/admin-draft-queries.ts:114) stops for terminal drafts without a pending marker.
   - [Set-default mutation:140](/Users/strelia/projects/courses/frontend/src/app/composition/admin-audio-queries.ts:140) invalidates audio and video queries.
   - [Video reread:87](/Users/strelia/projects/courses/frontend/src/app/composition/admin-video-queries.ts:87) invalidates the draft only when **generation changes**. Default changes preserve generation.

   **Impact:** with a ready draft cached, changing default can leave the old text and `stale=false` visible throughout automatic replacement. Even a single draft invalidation is insufficient: the initial stale-but-ready response still stops polling before the worker claims it.

   **Required plan edit:** specify the composition-layer invalidation and polling lifecycle for default changes, including deletion with default re-election. Add a mounted-editor scenario: ready A → default B → stale banner before Claim → replacement progress/result without reload or manual Regenerate. Preserve existing editor text and token protections.

6. **NON-BLOCKING · Edge cases / evidence accuracy · OPEN, actionable — the claimed crash retry is not the configured recovery mechanism.**

   [Plan:44](/Users/strelia/projects/courses/docs/multi-audio/plan-s6.md:44) justifies reclaiming `running` through an “asynq retry after worker failure.” The queue uses [MaxRetry(0):47](/Users/strelia/projects/courses/backend/internal/transcript/adapters/asynq/queue.go:47). Installed asynq v0.26.0 `recoverer.go:98` archives expired leases when retries are exhausted. Application [Recover:198](/Users/strelia/projects/courses/backend/internal/transcript/application/recognize.go:198) marks expired drafts `failed/timeout`; it does not enqueue them.

   **Impact:** a direct `running → Claim` integration test cannot prove crash redelivery. The duplicate guard remains valid, but its recovery explanation overstates the available path.

   **Plan edit:** describe the existing recovery/manual-regeneration behavior accurately. Add queue round-trip tests for `Requested=true`, `false`, and missing legacy fields through payload decoding. Any proposed retry-policy change needs explicit scope and critique.

AI-slop lens: **clean**—the provenance column and existing polling/extraction mechanisms serve concrete requirements. Unnecessary-complexity lens: **clean**—no additional framework, event system or unused port is proposed.

Inspected the parent plan, raw round-one findings, backfill plan/report and migrations, transcript Claim/Complete/Recover/Pending/GetDraft, domain construction/staleness, manual regeneration, queue payload decoding, worker wiring, extraction, and frontend query/editor paths. Backfill supports removing the temporary fallback; its recorded exclusions remain retained.

Compared with round one, both original backend blockers have concrete remedies; the remaining blocker is the unplanned frontend delivery of F2’s observable behavior. **Open findings: 1 blocking, 1 non-blocking.**

No files changed. Tests, network probes and graph refresh were **NOT RUN**. Planned gates remain unexecuted and require a separately authorized execution window.

VERDICT: BLOCKED