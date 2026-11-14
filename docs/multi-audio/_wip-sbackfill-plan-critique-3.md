`stage: plan` · Critic: Codex `/root`, selected independent context · **S-backfill, ROUND 3**

Plan: `docs/multi-audio/plan-sbackfill.md`, revision 2026-11-14, SHA-256:
`b6b2295860761e725b0e11e9042aaa7d1c5329bc863eadefd9c430c92245a96c`.

Snapshot: `/Users/strelia/projects/courses`, HEAD `229286971f2d9b1539cbd4ed60dafdfc6b2140d6`. Plan and raw critiques 1–2 are untracked. Inspected tracked media sources, migrations, frontend media module, migration runner, parent plan/report, Makefile and Compose are unchanged against HEAD. Dirty package metadata SHA-256: `a74e0f035af81f5f21cb63d0b06e669cbc0bac1c6f60b2fd804af3b24d07dbd4`. Plan hash and HEAD rechecked at completion.

Prompt: **Plan critic — version 1**, candidate root `/Users/strelia/project-data/ai-factory-workspace/runtime-data/ai-factory-rollout/candidates/17a9527421e8fac3e43a16c287c404118699833c`, `prompts/plan-critic.md`, SHA-256 `514e4c376c2551afe76e390c5779867ecf7f04ad2a552c97c2b4f713c43dcab6`. Package `0.8.0-alpha.1+native-handoff.1`, unreleased; package commit remains unverified.

**Prior findings**

1. **F4 — RESOLVED, blocking closed.** [Plan:39](/Users/strelia/projects/courses/docs/multi-audio/plan-sbackfill.md:39) replaces unavailable cancellation with `und-x-original`. It satisfies the [server language rule:267](/Users/strelia/projects/courses/backend/internal/media/domain/audiotrack.go:267), survives [HLS LANGUAGE rendering:102](/Users/strelia/projects/courses/backend/internal/media/adapters/ffmpeg/master.go:102), and is displayed by the [frontend fallback:70](/Users/strelia/projects/courses/frontend/src/modules/media/ui/LessonAudioTracksField.tsx:70). The [unique index:6](/Users/strelia/projects/courses/backend/migrations/00026_audio_track_language_unique.sql:6) excludes failed rows and includes queued/processing. H/H2 now distinguish successful fallback from atomic refusal when both tags are occupied.

2. **F5 — RESOLVED, non-blocking closed.** [Plan:121](/Users/strelia/projects/courses/docs/multi-audio/plan-sbackfill.md:121) explicitly tests D-B3 rejection, unchanged migration version and rollback of a separate candidate’s insertion. This tests the guard itself.

3. **F6 — RESOLVED, non-blocking closed.** [Plan:49](/Users/strelia/projects/courses/docs/multi-audio/plan-sbackfill.md:49) unambiguously places `SET LOCAL lock_timeout` before `LOCK` and preserves transactional migration/version updates.

4. **Round-1 B1 — NARROWED; B2 and N1 remain RESOLVED.** The predicate covers unfinished/failed uploads; [promote:382](/Users/strelia/projects/courses/backend/internal/media/adapters/postgres/audiotrack.go:382) preserves an existing default. D-B6 explicitly excludes already-Ready uploaded cases and prohibits reporting them as repaired. Table locking still covers current mutation paths; promote/delete write before rewriting within their transactions. Down→up remains correctly described as reversibility.

**New finding**

5. **F7 — NON-BLOCKING, OPEN · logical gap: operational diagnostics are overstated.** [Plan:47](/Users/strelia/projects/courses/docs/multi-audio/plan-sbackfill.md:47) promises an exception containing the asset ID, and D-B3 promises a named failure. However, the production [migration runner:67](/Users/strelia/projects/courses/backend/cmd/migrate/main.go:67) replaces the goose error with a generic message. Direct-provider integration tests can inspect the underlying error; `make migrate` cannot expose that detail through its current application log.

   **Impact:** refusal and rollback remain correct, but diagnosis requires extra work. **Plan action:** distinguish SQL/provider diagnostics from operator-visible output and document a concrete read-only diagnostic query or route. No runner change is required by this finding.

Four lenses applied: AI-slop—clean, one bounded migration; unnecessary complexity—clean, existing goose/schema mechanisms reused; logical gaps—F7; edge cases—conflict, rollback, retry, concurrency and exclusion paths reviewed without another blocker.

Raw rounds 1–2, parent D3–D7/§5/§8/§10, migrations 00025–00027, publish/mint/promote/delete/metadata paths, historical audio requirement, master writer, frontend fallback and goose transaction handling were inspected. This materially re-critiques the fallback and test-strategy changes; retained findings are explicitly rebound to the SHA above.

**Limitations:** four readings occurred in the selected context; **0 launched / 0 returned** additional reviewers, as instructed. A formal independent four-agent gate is not claimed. No files, tests, network probes or graph refresh were performed. Planned verification remains unexecuted and requires its separately authorized execution window.

**Open findings: 0 blocking, 1 non-blocking. Acceptance covers only the declared S-backfill scope, including D-B6’s exclusion.**

VERDICT: ACCEPTED