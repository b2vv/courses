**BLOCKED — D1–D7 phase decision surface. Three open substantive blockers.** O1–O4 are retained as authorized.

`stage: plan`; critic: Codex, this independent session.  
Target: `/Users/strelia/projects/courses`, branch `multi-subtitles`, HEAD `5cbb44bb6d0f9888750f008e50e4cca1f2093c1b`.  
Plan: untracked [plan.md](/Users/strelia/projects/courses/docs/multi-subtitles/plan.md), SHA-256 `5b4dd912a8125821f95b0736755f8db9329c1dfb73fd89ea3be9345fcdb56b55`, unchanged at completion. Relevant tracked product sources were clean; `backend/upload` was untracked and was not executed.

Prompt: version 1, `prompts/plan-critic.md` under the configured Factory cache; SHA-256 `514e4c376c2551afe76e390c5779867ecf7f04ad2a552c97c2b4f713c43dcab6`. Package metadata declares `0.8.0-alpha.1+native-handoff.1`, with its release commit unverified.

1. **Blocker, open — D1’s primary fallback can show a dub under the original.**  
   Location: [D1:55](/Users/strelia/projects/courses/docs/multi-subtitles/plan.md:55), [D5:103](/Users/strelia/projects/courses/docs/multi-subtitles/plan.md:103).

   Counterexample traced: embedded `und`/`und-x-original`, Ukrainian primary → make English dub default → publish its draft, moving primary to `en` → restore embedded default. S6 leaves published transcripts untouched, so D1 now selects English beneath Ukrainian audio.

   The exception also covers newly uploaded unknown-language tracks: S5 explicitly offers `und`, without restricting it to embedded tracks ([audio-track.ts:80](/Users/strelia/projects/courses/frontend/src/modules/media/domain/audio-track.ts:80)). Thus “default `und`” does not establish the claimed legacy relationship.

   **Required plan change:** define how unknown audio retains an established transcript association without borrowing a primary moved by another track. Preserve S6’s boundary; changing default alone must not rewrite transcripts.

2. **Blocker, open — D6’s language tab does not identify its draft or publication actions.**  
   Location: [D6:110](/Users/strelia/projects/courses/docs/multi-subtitles/plan.md:110).

   Counterexamples traced:

   - `en` and `EN` coexist under the case-sensitive index ([00026:6](/Users/strelia/projects/courses/backend/migrations/00026_audio_track_language_unique.sql:6)). D4 creates two drafts; D6 merges them into one tab promising “its own draft.” Which track does Regenerate/Publish address?
   - Embedded `und` produces a detected `uk` draft under D5. The `und` tab reads `?language=und`, while publication writes `uk`; that publication then appears as an orphan language.
   - An orphan transcript has no track draft to regenerate.
   - “Current asset” is ambiguous during replacement: the authoring surface can show a newer processing/awaiting-audio asset while playback retains the Ready asset ([LessonVideoField.tsx:48](/Users/strelia/projects/courses/frontend/src/modules/media/ui/LessonVideoField.tsx:48)).

   **Required plan change:** specify asset selection, language-to-track cardinality, unknown-language tab behavior and orphan actions. Also assign a slice to expose the published-language inventory: the existing read returns only one transcript ([repository.go:62](/Users/strelia/projects/courses/backend/internal/transcript/adapters/postgres/repository.go:62)).

3. **Blocker, open — D3/D7’s initial default transcript can contradict remembered audio.**  
   Location: [D3:77](/Users/strelia/projects/courses/docs/multi-subtitles/plan.md:77), [D7:116](/Users/strelia/projects/courses/docs/multi-subtitles/plan.md:116).

   Counterexample traced: lesson default `uk`, remembered preference `en`, both transcripts present. D3 initially supplies Ukrainian. S7 selects English synchronously during manifest processing, before requesting the default playlist ([hls-media.ts:55](/Users/strelia/projects/courses/frontend/src/modules/playback/adapters/hls-media.ts:55)). Until selection notification and the English query complete, the initial transcript is wrong for the selected audio.

   **Required plan change:** define unresolved-selection/loading behavior and prohibit retaining another language’s captions or panel during query transitions. Include remembered-start selection and rapid switching in acceptance. This requires no media recreation.

4. **Nonblocking, actionable — D2 is sound, but its impact inventory needs correction.**  
   Location: [D2:65](/Users/strelia/projects/courses/docs/multi-subtitles/plan.md:65).

   Counterexamples checked: `e`, `1x`, repeated hyphens and a 35-character single subtag pass the old audio rule but fail transcript grammar. A real frontend test explicitly accepts `'x'.repeat(35)` ([audio-track.test.ts:38](/Users/strelia/projects/courses/frontend/src/modules/media/domain/audio-track.test.ts:38)). All seven S5 offered languages pass D2; `en-US`, `und` and `und-x-original` also pass.

   Refusing nonconforming persisted rows is a reasonable pre-launch failure mode: it prevents silent semantic rewriting. Keep the planned count, identifiable offending rows and explicit remediation before retry. Actual database contents were **not inspected**.

5. **Nonblocking, actionable — D3 changes the editor’s read contract before MS4.**  
   Location: [slice order:125](/Users/strelia/projects/courses/docs/multi-subtitles/plan.md:125).

   Counterexample traced: default `en`, existing transcript only `uk`. The current admin editor reads without language ([admin route:348](/Users/strelia/projects/courses/frontend/src/routes/admin.courses.$courseId.tsx:348)); D3 returns absence. `LessonTranscriptField` then opens a new `uk` transcript without a token ([field:82](/Users/strelia/projects/courses/frontend/src/modules/transcript/ui/LessonTranscriptField.tsx:82)); saving conflicts with the existing row.

   Assign interim editor compatibility to MS1 or explicitly couple delivery. The learning route is another caller; **there is no transcript SSR loader here**: `/learn/$` declares `ssr: false` ([route:34](/Users/strelia/projects/courses/frontend/src/routes/learn.$.tsx:34)). Discussion uses its own facts rather than this transcript read.

   Reading audio facts in transcript SQL fits the existing S6 practice ([repository.go:30](/Users/strelia/projects/courses/backend/internal/transcript/adapters/postgres/repository.go:30)) and ADR 013; no cross-module package import or speculative port is needed.

6. **Nonblocking, actionable — MS3 must explicitly carry track identity beyond the database key.**  
   Location: [D4:84](/Users/strelia/projects/courses/docs/multi-subtitles/plan.md:84), [MS3:137](/Users/strelia/projects/courses/docs/multi-subtitles/plan.md:137).

   Counterexample traced: two tracks without drafts currently generate the same queue identity, `asset:generation:none` ([queue.go:61](/Users/strelia/projects/courses/backend/internal/transcript/adapters/asynq/queue.go:61)). Rekeying SQL alone therefore suppresses one job.

   Carry these requirements into the slice contract:

   - Track identity in queue deduplication, completion/recovery tokens, query keys and pending markers. The current frontend markers are lesson-scoped ([admin-draft-queries.ts:103](/Users/strelia/projects/courses/frontend/src/app/composition/admin-draft-queries.ts:103)).
   - C09’s observable replacement guarantee: old and new assets now have different draft rows, so the existing UPSERT generation guard no longer establishes it ([draft_integration_test.go:338](/Users/strelia/projects/courses/backend/internal/transcript/adapters/postgres/draft_integration_test.go:338)).
   - Replace default-leaf staleness for per-track drafts; otherwise every nondefault draft reads stale ([draft.go:178](/Users/strelia/projects/courses/backend/internal/transcript/domain/draft.go:178)).
   - Define language renaming during recognition/publication: metadata can change without changing track ID or leaf ([audiotrack.go:274](/Users/strelia/projects/courses/backend/internal/media/adapters/postgres/audiotrack.go:274)).

   D4’s atomic four-fact claim is appropriate for C21. Cascade eviction remains consistent with `DeleteSweptMediaAsset` ([generated query:14](/Users/strelia/projects/courses/backend/internal/media/adapters/postgres/generated/media.sql.go:14)). The per-lesson pending index belongs to **video assets**, not ASR drafts ([00005:34](/Users/strelia/projects/courses/backend/migrations/00005_media.sql:34)); it does not forbid multiple track drafts.

   Primary promotion is implementable, but MS3 needs a common serialization/order rule with manual PUT and a concurrent-writer test. “Same transaction” alone does not specify that rule; manual PUT derives primary through `NOT EXISTS` ([write.go:59](/Users/strelia/projects/courses/backend/internal/transcript/adapters/postgres/write.go:59)).

7. **Nonblocking, actionable — correct factual claims and preserve recovery/UI behavior.**

   - **§2 backfill:** it normally creates `und`; `und-x-original` is the collision fallback ([00028:49](/Users/strelia/projects/courses/backend/migrations/00028_audio_track_embedded_backfill.sql:49)).
   - **§2/§6 ASR cost:** 38 minutes/RTF 1.56 describes whisper.cpp. §6.10 estimates **about 19 minutes per hour** for faster-whisper on an unloaded six-CPU worker, from measured CPU consumption ([spec:552](/Users/strelia/projects/courses/docs/asr/spec.md:552)). “38 minutes CPU” is also the wrong unit.
   - **D5 hinting:** first-subtag plus supported-code lookup is implementable; unsupported → auto-detect is specified. The pinned dependency is faster-whisper 1.2.1, but its actual language-list implementation was unavailable in the inspected sources. Verify that list in MS3 rather than assuming grammar validity implies engine support. Clarify precedence between D5’s forced hint and §6’s retained manual override.
   - **D7:** media creation currently excludes captions from its dependencies ([LessonPlayer.tsx:368](/Users/strelia/projects/courses/frontend/src/modules/playback/ui/LessonPlayer.tsx:368)); preserve that, the caption preference effect at line 595, and native default behavior ([hls-media.ts:113](/Users/strelia/projects/courses/frontend/src/modules/playback/adapters/hls-media.ts:113)).
   - **MS4 recovery:** specify migration/recovery of existing `author:lesson` snapshots, not just a new key ([recovery.ts:177](/Users/strelia/projects/courses/frontend/src/modules/transcript/application/recovery.ts:177)).
   - **MS2 panel:** choose whether search/follow state resets on language change; it currently persists in component state ([TranscriptPanel.tsx:22](/Users/strelia/projects/courses/frontend/src/modules/transcript/ui/TranscriptPanel.tsx:22)). Heartbeats and discussion timestamps remain lesson-position based; no language-dependent change was found necessary.

The remaining §2 schema/read/write/editor claims match the inspected sources. No previous verdict for this target plan was supplied or found; there are no prior findings to mark resolved. Existing S6/S7 evidence is retained only for its recorded snapshot.

No writes, tests, network probes or graph refresh ran. This record could not be saved because the workspace is read-only. The four-lens gate is also **NOT_RUN: 0 launched / 0 returned**; the selected [skill](/Users/strelia/.agents/skills/plan-critique/SKILL.md) requires “a separate subagent with clean context,” while this request preserves the existing reviewer route without spawning. This independent review does not claim completion of that stronger gate.