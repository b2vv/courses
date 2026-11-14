**BLOCKED — D1–D7 phase decisions. Three substantive blockers remain.** O1–O4 remain authorized. This verdict concerns the phase plan; slice plans still require their own critique.

`stage: plan`; critic: Codex, this selected independent session, round 2.

Plan: [docs/multi-subtitles/plan.md](/Users/strelia/projects/courses/docs/multi-subtitles/plan.md), r2, SHA-256 `8d8065c0592f037fd7b672d08c3bad6c51a773a46924db810b5f92a8c1dffcbe`.

Target: `/Users/strelia/projects/courses`, branch `multi-subtitles`, HEAD `5cbb44bb6d0f9888750f008e50e4cca1f2093c1b`. Relevant tracked backend/frontend sources are clean. Plan and raw round-1 review are untracked; the latter hashes to `fbff4e427ae41732a8411c05b9a44387980b325925f45617f0a12e38e615fb87`. Unrelated documentation/configuration WIP and untracked `backend/upload` were excluded.

Prompt: version 1, `<factory>/prompts/plan-critic.md`, SHA-256 `514e4c376c2551afe76e390c5779867ecf7f04ad2a552c97c2b4f713c43dcab6`; `<factory>` is `/Users/strelia/project-data/ai-factory-workspace/runtime-data/ai-factory-rollout/candidates/17a9527421e8fac3e43a16c287c404118699833c`. Package version declares `0.8.0-alpha.1+native-handoff.1`; its release commit remains unverified.

1. **Blocker — open: D4 cannot establish D5’s rename staleness or language comparison.**

   Locations: [plan:110](/Users/strelia/projects/courses/docs/multi-subtitles/plan.md:110), [plan:119](/Users/strelia/projects/courses/docs/multi-subtitles/plan.md:119), [plan:129](/Users/strelia/projects/courses/docs/multi-subtitles/plan.md:129).

   Counterexample: recognize track `en-US`; whisper receives `en` and returns `en`. Comparing existing `draft.language` with current `captionLanguage(T) = en-us` rejects an unchanged track. Comparing only first subtags instead misses a rename from `en-US` to `en-GB`. D4’s stated stale predicate also remains false after either rename: generation and track existence did not change.

   The existing column stores **recognition output**, populated from `recognition.Language` ([recognize.go:105](/Users/strelia/projects/courses/backend/internal/transcript/application/recognize.go:105)); it does not preserve the full track tag at recognition time.

   **Required change:** distinguish recognition output from the track-language snapshot captured for the attempt. Specify its persistence, migration, capture timing and participation in stale/publish guards. Define the `und`→detected exception against that snapshot. This is a missing contract, not merely a column-naming detail.

2. **Blocker — open: D5’s two-call flow lacks a publication consistency and retry contract.**

   Location: [plan:125](/Users/strelia/projects/courses/docs/multi-subtitles/plan.md:125).

   Frontend orchestration through the two existing module APIs is architecturally acceptable. It requires neither a media-package import into transcript nor an I/O-bearing shared contract.

   Counterexample: the UI reviews draft A, PATCHes `und`→`uk`, and another Regenerate completes draft B before publish. The existing publish request identifies only the lesson and the **published transcript’s** optimistic token; publish then reads whichever ready draft exists ([draft.go:149](/Users/strelia/projects/courses/backend/internal/transcript/application/draft.go:149)). Track identity alone does not establish that the reviewed attempt is the published attempt.

   A second interleaving is rename/replacement between the proposed language check and transcript commit. Existing save stamps the Ready generation at write time ([write.go:42](/Users/strelia/projects/courses/backend/internal/transcript/adapters/postgres/write.go:42)); an old draft could consequently be stamped against the replacement unless publication coordinates those facts.

   “Repeat (2) is idempotent” is also unsupported: after a successful publication with a lost response, retrying the old transcript token returns `transcript_modified` or `transcript_exists` ([write.go:50](/Users/strelia/projects/courses/backend/internal/transcript/adapters/postgres/write.go:50)).

   **Required change:** bind publish to the reviewed track/attempt, define atomic validation with the transcript write, and specify lost-response recovery. Add planned checks for concurrent Regenerate, rename and replacement.

   Rename collision and ordinary master-regeneration failure are already handled sensibly: PATCH performs the rewrite before committing metadata, and a rewrite error aborts the transaction ([audiotrack.go:274](/Users/strelia/projects/courses/backend/internal/media/adapters/postgres/audiotrack.go:274)). These do not require a cross-module transaction.

3. **Blocker — open, owner decision: D1/D5 add a product policy beyond O1/O3.**

   Locations: [plan:61](/Users/strelia/projects/courses/docs/multi-subtitles/plan.md:61), [plan:72](/Users/strelia/projects/courses/docs/multi-subtitles/plan.md:72), [plan:125](/Users/strelia/projects/courses/docs/multi-subtitles/plan.md:125).

   Removing the primary fallback faithfully prevents O1/O3’s wrong-language result. However, those decisions do not uniquely require **all existing unknown-language tracks to lose captions**, nor require publishing text to relabel an audio track using machine detection.

   Counterexample: an existing embedded `und` track has a reviewed Ukrainian transcript. R2 hides it until metadata is changed. A manually authored transcript can also validly use `und`; D1’s “never” suppresses even that exact-language match. This is an additional exception to the stated equality rule.

   **Escalate the concrete policy to the owner:** unknown-language tracks receive no captions until labelled; publishing an ASR draft changes the track’s language to the displayed detected language, with collision refusal. Pre-launch status permits changing merged behavior directly, but does not establish that this particular product tradeoff was authorized. An alternative association/confirmation policy remains possible without restoring the unsafe primary fallback.

4. **Nonblocking — actionable: D2 is sound for participating tracks; qualify its cardinality claim.**

   Location: [plan:82](/Users/strelia/projects/courses/docs/multi-subtitles/plan.md:82).

   Counterexamples retried: mint `en` then `EN`; rename `de`→`EN`; failed `en` followed by Ready `EN`; embedded `und`; backfill fallback `und-x-original`.

   A database index on `(asset, lower(language)) WHERE state <> 'failed'` covers inserts, renames and state changes. It ensures at most one **non-failed** track per normalized language. Failed duplicates remain legal; D6’s Ready-only tabs therefore remain unambiguous.

   Master regeneration performs no language equality comparison: it renders the supplied value verbatim into `LANGUAGE` ([master.go:95](/Users/strelia/projects/courses/backend/internal/media/adapters/ffmpeg/master.go:95)). Playback matching **is case-sensitive** ([hls-media.ts:30](/Users/strelia/projects/courses/frontend/src/modules/playback/adapters/hls-media.ts:30)); retaining raw manifest values for selection and normalizing only caption lookup avoids disagreement. A case-only rename can still invalidate a remembered raw preference; document or normalize that preference in the slice.

   Historical backfill checks use case-sensitive comparisons ([00028:27](/Users/strelia/projects/courses/backend/migrations/00028_audio_track_embedded_backfill.sql:27)). The later D2 migration must diagnose resulting normalized collisions as well as malformed languages, retaining identifiable rows and explicit remediation.

5. **Nonblocking — actionable: D6 recovery migration remains genuinely unresolved.**

   Location: [plan:145](/Users/strelia/projects/courses/docs/multi-subtitles/plan.md:145).

   Counterexample: a snapshot without `language` predates a primary change. Today’s primary cannot establish which language the old editor read. Existing snapshots explicitly permit the field to be absent ([recovery.ts:30](/Users/strelia/projects/courses/frontend/src/modules/transcript/application/recovery.ts:30)).

   It is reasonable to defer the algorithm to MS3, but remove the implication that historical primary is reconstructible. MS3 must specify preservation/explicit recovery for ambiguous snapshots, destination-key collisions and repeated migration. Also cover `und`→named-language key movement after PATCH.

   Tabs, orphan actions and asset selection are otherwise explicit enough for this phase plan.

6. **Nonblocking — actionable: D7 and the slice order need narrower implementation claims.**

   Locations: [plan:150](/Users/strelia/projects/courses/docs/multi-subtitles/plan.md:150), [plan:167](/Users/strelia/projects/courses/docs/multi-subtitles/plan.md:167).

   - The lowercase render guard is correct: `en-US` selects `en-us`, matching the normalized transcript. Absence returns `language: null` and must settle into O3’s missing state.
   - `onAudioTrack` currently exists between the adapter and player; it is **not a player-to-route prop** ([LessonPlayer.tsx:46](/Users/strelia/projects/courses/frontend/src/modules/playback/ui/LessonPlayer.tsx:46)). MS4 must add that outward binding.
   - Its current source is `AUDIO_TRACK_SWITCHING`, not completed switching ([hls-media.ts:63](/Users/strelia/projects/courses/frontend/src/modules/playback/adapters/hls-media.ts:63)). Counterexample: a cached target transcript renders while buffered old audio still plays. Define the transition boundary and verify it; S7 already measured delayed audible switching.
   - Native `audioTracks` needs initial selection plus change/removal handling; API presence alone does not confirm a usable language. The stated fake-only coverage limitation remains valid.
   - Player-last prevents the intended D1 behavior change arriving before authoring support. It does **not** establish MS2/MS3 compatibility: MS2 changes draft HTTP identity while the current frontend remains lesson-scoped. Assign interim compatibility or coordinated delivery in the MS2 slice contract.

The revised C09 formulation is otherwise sound. During replacement, a newer **processing** asset must not disqualify the old latest Ready asset. Once the replacement becomes Ready, old claims are refused and old drafts are hidden. Distinct track rows prevent a late old completion from overwriting the new draft. D4 now explicitly carries track identity through TaskID, complete/recover tokens and frontend keys.

Explicit dispositions of round-1 findings:

| R1 | Disposition | Counterexample retried |
|---|---|---|
| 1 | **Resolved technically** | Dub publication, then return to embedded `und`: no primary fallback can display the dub. New product-policy blocker is finding 3 above. |
| 2 | **Resolved at phase level** | `en`/`EN`, detected `uk` from `und`, orphan transcript, newer processing asset: D2/D5/D6 now identify tracks and actions. Publication consistency remains separately blocked. |
| 3 | **Resolved for query transitions** | Remembered non-default startup and rapid switching: D7 suppresses mismatched results. Audible-switch timing remains finding 6. |
| 4 | **Resolved** | Invalid single subtags and repeated hyphens are explicitly rejected; affected frontend test is named. |
| 5 | **Resolved** | Default `en`, existing `uk` transcript: unchanged no-language read preserves the current editor contract. |
| 6 | **Partially resolved** | Two track jobs now have distinct identities; replacement semantics are explicit. Rename staleness/publication remain findings 1–2. Primary promotion was removed. |
| 7 | **Partially resolved** | Backfill/cost facts, hint precedence, caption preference and panel reset are corrected. Ambiguous recovery and native/transition details remain findings 5–6. |

Evidence retained: unchanged schema/read/write paths, ADR 013 boundaries and S6/S7’s recorded mechanisms. Their historical test results are not transferred to this plan.

**Final verdict: BLOCKED, with three open substantive blockers.** No tests, writes, network probes or graph refresh ran. Plan/review hashes were unchanged on final reread; this record could not be saved in the read-only workspace.

The stronger four-lens gate remains **NOT_RUN — 0 launched / 0 returned**: the selected [skill](/Users/strelia/.agents/skills/plan-critique/SKILL.md) requires “a separate subagent with clean context,” while this request preserves the existing reviewer without spawning. Its referenced `reference-anti-patterns.md` was also unavailable at the referenced location. This review does not claim that gate passed.