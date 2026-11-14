Reviewed `origin/main...HEAD`: `ef48c15` → `6aef653`. Read-only review; no files changed. Counterexamples below were traced through source and tests; tests and ordered gates were **NOT RUN**.

**1. Claim SQL**

Opened `backend/internal/transcript/adapters/postgres/draft.go:113`, `:214`, `:253`; `s6_integration_test.go:69`.

Tried first insert, delayed generation 1 after generation 2, same-generation automatic/requested reclaim, NULL legacy leaf, default disappearing between statements, and old Complete/Recover tokens after replacement.

Result: first insert works independently of the conflict predicate; older generations cannot overwrite newer ones, even when requested. Automatic claims reject an already-current leaf; requested claims retain manual regeneration. NULL correctly means `original`. A vanished default produces no insertion and `Claimed:false`. Complete and Recover retain token protection. The CTE/`ON CONFLICT` syntax concern was discarded after checking PostgreSQL syntax.

**2. PendingRecognitions and queue identity**

Opened `draft.go:67`, `backend/internal/transcript/adapters/asynq/queue.go:32`, `backend/internal/transcript/application/recognize.go:159`, `backend/cmd/asr-worker/main.go:30`.

Traced running A → default B → queued replacement → A completes → another replacement receives the completion marker.

Result: multiple replacement tasks can exist, but the first B claim records B and subsequent automatic claims are refused. Running work already hearing the current default is excluded. Completion changes the marker, allowing a same-generation rerun despite retained task IDs. No-default assets are intentionally excluded; migration 00028 and transactional media invariants supply defaults for Ready assets. Current-leaf failures do not loop: manual regeneration remains required, as accepted in D-S6-2.

**3. Extraction safety and leaf compatibility**

Opened `backend/internal/transcript/adapters/media/audio.go:42`, `backend/internal/transcript/domain/draft.go:70`, `domain/transcript.go:248`, migrations `00029_asr_draft_source_leaf.sql:7` and `00025_multi_audio.sql:10`, `:25`; `audio_test.go:144`.

Tried traversal, empty leaf, uppercase UUID, nil UUID, and extensionless uploaded M4A input.

Result: unsafe leaves fail before filesystem access. Extraction uses the claimed leaf; ffmpeg probes the container, and its existing arguments support audio-only input. One harmless mismatch exists: 00029 accepts a nil-UUID leaf while domain validation rejects it. Migration 00025 forbids nil track IDs and binds uploaded leaves to their IDs, so production tracks cannot produce that value.

**4. Existing callers and changed tests**

Grepped the whole backend for Extract/Claim/Stale definitions and calls, plus transcript references to `original`. Opened `application/recognize.go:60`, `:69`, `:223`; `application/draft.go:84`; `application/read.go:74`; `draft_integration_test.go:190`, `:364`, `:509`.

Result: the sole production extraction caller forwards the claimed leaf. Both transcript Claim callers pass the appropriate requested/automatic flag. Draft staleness receives the default leaf; published-transcript staleness deliberately remains generation-only under the owner’s decision.

The reclaim tests’ changes to `requested=true` preserve the retained manual behavior. They are complemented by the automatic-duplicate rejection test at `s6_integration_test.go:95`; their assertions were not simply weakened. Added default-track fixtures reflect the existing Ready-asset invariant.

**5. Frontend invalidation, polling and copy**

Opened `frontend/src/app/composition/admin-draft-queries.ts:21`, `:42`, `:109`; `admin-audio-queries.ts:90`, `:143`; `admin.ts:31`; `frontend/src/router.tsx:34`; locales `en.ts:876` / `uk.ts:885`.

Tried prefix invalidation against pending-marker keys, simultaneous manual regeneration, two lessons sharing a factory, separate SSR requests, and remounting after query-cache eviction.

Result: `['admin','draft']` does not match `['admin','draft-pending',…]`; manual pending polling remains intact. The map is lesson-keyed and allocated inside request-scoped dependency creation, with no global SSR store. Copy covers another audio track and automatic preparation. **The timeout does reset after cache eviction**, detailed below.

**6. e2e production path, timing and isolation**

Opened `frontend/e2e/asr.spec.ts:37`, `:130`, `:332`, `:410`, `:450`; `media.setup.ts:65`, `:139`; `frontend/playwright.config.ts:4`, `:36`.

The first S6 test uploads genuinely video-only bytes plus speech audio through the UI, waits for the real ASR result through the API, and checks nonempty segments. Reading after reload respects the accepted first-draft-delivery narrowing.

The second exercises automatic replacement in the mounted editor, but its stale assertion races Claim and its final assertion checks token advancement rather than distinctive dubbed content. Locators are adequate for these isolated one-lesson fixtures, with audio controls scoped to their region. The 240/300-second waits accommodate the 60-second sweep and short recognition fixtures, without guaranteeing arbitrary backlog completion.

Repository configuration uses one worker and serial tests; the earlier shared-queue pause resumes in `finally`. No supported parallel pause/resume interference was found.

**Findings — ranked**

Blocker: none. Major: none.

- **Minor — [admin-draft-queries.ts:155](/Users/strelia/projects/courses/frontend/src/app/composition/admin-draft-queries.ts:155): cache eviction resets the stale-polling window.** After unmount and cache GC, a fresh observer has undefined data. This becomes `false`, deleting the retained timestamp at lines 47–49; the stale response starts another 15-minute window. Clear the timestamp only after an actual non-stale response. Add a GC/remount case.

- **Minor — [asr.spec.ts:473](/Users/strelia/projects/courses/frontend/e2e/asr.spec.ts:473): transient-banner assertion can flake.** A sweep immediately after set-default can claim B before the invalidated GET returns, making `stale=false` before the browser observes it. Pause recognition while observing stale, then resume in `finally`, following C06’s existing pattern.

- **Minor — [asr.spec.ts:477](/Users/strelia/projects/courses/frontend/e2e/asr.spec.ts:477): replacement content is not distinguished.** A changed token proves replacement occurred, but this test alone cannot detect recognition of the embedded speech instead of the dub. Assert distinctive dubbed content from the existing fixture.

**VERDICT: SHIP**

Source-review verdict with the minor issues above. Required runtime gates remain unverified by this review.