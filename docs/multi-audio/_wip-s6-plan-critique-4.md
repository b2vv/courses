`stage: plan` · Critic: Codex `/root`, selected independent context · **Round 4 — multi-audio S6**

Plan: [docs/multi-audio/plan-s6.md](/Users/strelia/projects/courses/docs/multi-audio/plan-s6.md), revision 2026-11-14. SHA-256:
`daf694b87387ebe15451f5f6faf0b22eea04ac650d8a2e0df9846ca13ddc79e4`

Snapshot: `/Users/strelia/projects/courses`, branch `multi-audio-s6`, HEAD `ef48c150fbfa093f0168efce02c3b6e15d1c8f3f`. Plan and raw critiques 1–3 are untracked. Inspected tracked product sources, parent plan and AGENTS are unchanged against HEAD. Unrelated dirty configuration/documentation and other untracked files are excluded.

Prompt: **Plan critic — version 1**, `prompts/plan-critic.md`, release directory:
`/Users/strelia/project-data/ai-factory-workspace/runtime-data/ai-factory-rollout/candidates/17a9527421e8fac3e43a16c287c404118699833c`

Verified prompt SHA-256:
`514e4c376c2551afe76e390c5779867ecf7f04ad2a552c97c2b4f713c43dcab6`

Selected pinned skill: `skills/plan/plan-critique/SKILL.md`; bytes match the local package pin. Package version is `0.8.0-alpha.1+native-handoff.1`; package commit remains unverified.

1. **F7 — NARROWED; no longer blocking the declared S6 scope.**

   [Plan:94](/Users/strelia/projects/courses/docs/multi-audio/plan-s6.md:94) explicitly excludes automatic delivery of the first draft into an already mounted editor. The underlying behavior is confirmed:

   - [Video query:88](/Users/strelia/projects/courses/frontend/src/app/composition/admin-video-queries.ts:88) skips invalidation for the first observed Ready generation.
   - [No-draft response:73](/Users/strelia/projects/courses/backend/internal/transcript/adapters/http/draft_handler.go:73) returns null state and false staleness.
   - [Draft polling:114](/Users/strelia/projects/courses/frontend/src/app/composition/admin-draft-queries.ts:114) stops without running state or a manual pending marker.
   - [LessonDraftField:135](/Users/strelia/projects/courses/frontend/src/modules/transcript/ui/LessonDraftField.tsx:135) exposes Generate when no draft exists; [mutation handlers:132](/Users/strelia/projects/courses/frontend/src/app/composition/admin-draft-queries.ts:132) start observation on success or `already_running`.
   - [Router:36](/Users/strelia/projects/courses/frontend/src/router.tsx:36) sets 60-second cache staleness. Installed Query defaults permit stale-query refetch when the tab becomes visible again (`queryObserver.ts:762`, `focusManager.ts:23`).

   This establishes an existing first-video observation gap, independently of S6. The dated exclusion supplies a reproducible mechanism and therefore meets [AGENTS:38](/Users/strelia/projects/courses/AGENTS.md:38). Its “magnitude” is **not a finite latency bound**: reload/refocus only reads the state available then. That qualification does not invalidate the mechanism-based exclusion.

   [E2e scope:143](/Users/strelia/projects/courses/docs/multi-audio/plan-s6.md:143) now matches the narrowed contract. F7 remains an explicit limitation, not a repaired behavior.

2. **F1–F6 — earlier dispositions retained after rechecking.**

   [Plan:33](/Users/strelia/projects/courses/docs/multi-audio/plan-s6.md:33) preserves conditional automatic Claim, manual regeneration and generation monotonicity; existing completion remains token guarded (`postgres/draft.go:205`). Leaf-based staleness and terminal-draft delivery are specified at [plan:75](/Users/strelia/projects/courses/docs/multi-audio/plan-s6.md:75). Poll cadence is separated from queue delay; overwrite attribution is corrected; recovery accurately retains `MaxRetry(0)` and manual regeneration. Published-transcript preservation remains the recorded owner narrowing.

3. **F8 — NON-BLOCKING · Edge cases · OPEN, actionable: stale polling is not time bounded.**

   [Plan:90](/Users/strelia/projects/courses/docs/multi-audio/plan-s6.md:90) calls polling “bounded” because stale drafts qualify for pending recognition. Eligibility does not guarantee progress: with the worker unavailable, a stale terminal draft can remain stale indefinitely.

   **Impact:** continued polling while the editor remains active. **Plan edit:** describe termination as conditional on worker progress and record worker-unavailable polling as a limitation, or specify a waiting bound. This does not prevent the scoped recognition behavior.

Four lenses applied in this selected context: AI-slop and unnecessary complexity are **clean**—existing polling, extraction and composition mechanisms suffice. Logical gaps found no new blocker; edge cases produced F8. **0 launched / 0 returned** additional reviewers, per the selected route; no formal four-agent gate is claimed.

Read raw rounds 1–3, parent D5, backfill migration/report, transcript callers/queue/recovery, audio default/delete/promotion writers, frontend queries/editor and installed Query defaults. Compared with round 3, the material change is F7’s explicit scope exclusion and reload-based test strategy; unchanged backend evidence is retained.

Writes, tests, network probes and graph refresh were **NOT RUN**. Planned gates remain unexecuted.

**Open findings: 0 blocking, 1 non-blocking. Acceptance applies only to the exact plan and snapshot above.**

VERDICT: ACCEPTED