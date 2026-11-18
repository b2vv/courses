1. **BLOCKER — R1-1 OPEN: cancellation restores captions without proving which audio will play.**  
   **Location:** `docs/multi-subtitles/plan-ms4.md:73–79`, `:165–166`, `:219–220`; `frontend/node_modules/hls.js/src/controller/audio-stream-controller.ts:485–508`, `:792–807`, `:1114–1118`; `frontend/node_modules/hls.js/src/controller/base-stream-controller.ts:2573–2620`.

   Counterexample: `en` is audible; `de` has buffered ahead; the viewer returns to `en` before `de` becomes current. M1 immediately confirms `en`, but `nextAudioTrack` does not guarantee removal of every buffered `de` sample. If reloading `en` is delayed, `de` can play under restored `en` captions. Moreover, hls.js completes the **latest `switchingTrack`** when the playing fragment changes level; it does not check that the fragment’s level equals that track’s id. Thus SWITCHED can also misidentify an intermediate buffered track during rapid switches.

   E4’s held-`de` recipe prevents the buffered-`de` condition it supposedly measures. `data-caption-selection` merely repeats M1’s conclusion; it cannot establish which fragment became current. The existing decoded-audio observer is available at `frontend/e2e/audio-tracks.spec.ts:164–201`.

   **Required plan change:** define confirmation/cancellation that preserves D7/O4 through already-buffered intermediate audio. Add separate forced cases for cancellation before buffering and after buffering, with independent observation of actual audio. Recording a wrong-language interval as a limitation does not satisfy the accepted contract.

2. **NONBLOCKING — R1-2 CLOSED: the render boundary is now explicit.**  
   **Location:** `plan-ms4.md:100–108`; `frontend/src/routes/learn.$.tsx:82–90`, `:276–299`; `frontend/src/modules/playback/ui/LessonPlayer.tsx:370`, `:853–855`; `backend/internal/transcript/application/read.go:60–61`.

   M4 correctly distinguishes locked, email-gated, non-ready and Ready lessons. It no longer waits for an impossible media callback on a non-ready lesson. **Required plan change:** none to this boundary; finding 11 addresses its conflicting query rule.

3. **NONBLOCKING — R1-3 CLOSED: native invalidation has explicit states.**  
   **Location:** `plan-ms4.md:84–88`, `:197`; `frontend/src/modules/playback/adapters/hls-media.ts:113–140`; `frontend/src/modules/playback/ports/PlaybackMedia.ts:20–26`.

   The proposed callbacks now distinguish pending from unavailable and invalidate previous confirmation on missing tracks/languages. The specified initial and valid→missing/empty tests cover the original counterexamples. **Required plan change:** none.

4. **BLOCKER — R1-4 OPEN: a load counter does not identify a mounted player.**  
   **Location:** `plan-ms4.md:111–114`, `:176–177`, `:200`; `frontend/src/routes/learn.$.tsx:90`, `:282–285`; `frontend/src/modules/playback/ui/LessonPlayer.tsx:295`, `:359`.

   Counterexample: A confirms `{A, generation:1, en}`; navigate to B without pressing Play; return to A. `OpenLesson` survives, while A’s player remounts. No new `load()` has occurred to invalidate the old record. Its lesson id and “last known” generation still match, so the route can accept old confirmation and render a track **before Play**. A per-player counter also repeats after remounting.

   For retry/remint, “last known generation” has no defined authoritative update mechanism: a new pending generation must be adopted, while a late old event must be ignored. Falling back to prediction on every mismatch is unsafe once playback has started.

   **Required plan change:** define mounted-player identity, synchronous invalidation on remount, authoritative generation advancement, and stale-event rejection without reverting an active load to prediction. Add A→B-without-Play→A and same-instance retry/remint acceptance cases; E10 alone does not cover these.

5. **NONBLOCKING — R1-5 CLOSED under O7.**  
   **Location:** `plan-ms4.md:10–17`, `:90–98`, `:172–173`, `:182`; `frontend/src/modules/playback/ui/LessonPlayer.tsx:548–550`, `:295`; `frontend/e2e/transcript.spec.ts:342–364`.

   O7 explicitly preserves transcript reading before Play, and M10 moves the caption anchor after Play. Initial confirmation is supported by hls.js: synchronous remembered selection precedes default selection (`audio-track-controller.ts:195–199`), and initial buffered-fragment detection can complete the switch (`base-stream-controller.ts:2692–2718`; `audio-stream-controller.ts:792–807`). Source reload resets both selection and fragment state (`audio-track-controller.ts:70–76`; `base-stream-controller.ts:375–384`).

   **Required plan change:** none for the former deadlock; finding 12 concerns remembered-start evidence.

6. **NONBLOCKING — R1-6 CLOSED, with one remaining evidence refinement.**  
   **Location:** `plan-ms4.md:129–138`, `:156–162`, `:182`, `:204`; `frontend/src/modules/playback/ui/LessonPlayer.tsx:595–601`, `:992–997`; `frontend/e2e/transcript.spec.ts:349–360`.

   Re-keying plus a ref to the current `<track>` addresses replacement and avoids `textTracks[0]`. Tests now use the real captions control and must not assign mode. Held transcript requests cover the request-in-flight interval. The attribute is a valid observer of the application’s reported state, subject to finding 1.

   **Required plan change:** explicitly assert **follow reset**, alongside search reset, in M09. The actual state is separate (`frontend/src/modules/transcript/ui/TranscriptPanel.tsx:22–24`, `:72–75`, `:89–91`).

7. **NONBLOCKING — R1-7 CLOSED: error handling is now implementable.**  
   **Location:** `plan-ms4.md:125`, `:145–150`, `:174–175`; `frontend/src/modules/transcript/ui/TranscriptPanel.tsx:12–17`, `:40–46`; `frontend/src/routes/learn.$.tsx:180`.

   The revision adds the missing error variant, copy, retry action and playback-continuation test. **Required plan change:** none.

8. **NONBLOCKING — R1-8 CLOSED: the zero-track hypothesis was corrected.**  
   **Location:** `plan-ms4.md:53–56`; `backend/internal/media/domain/audiotrack.go:208–224`; `backend/internal/media/adapters/postgres/audiotrack.go:218–227`; `backend/internal/media/adapters/postgres/audiotrack_integration_test.go:421–423`.

   Production deletion refuses the last Ready track of a Ready asset. `Processing→Ready` is also correctly acknowledged (`backend/internal/media/domain/asset.go:153–154`). **Required plan change:** none to this invariant; finding 14 concerns unexpected empty-manifest handling.

9. **NONBLOCKING — R1-9 CLOSED: fixtures, PATCH and inventory are concrete.**  
   **Location:** `plan-ms4.md:152–157`, `:179–190`.

   The named fixtures exist (`frontend/e2e/media.setup.ts:76–79`); PATCH exists and requires both fields (`backend/internal/media/adapters/http/handler.go:129`, `:367–370`). The third Ready track is now specified. The new spec is CI-selected (`frontend/playwright.config.ts:49–60`). Inventory dispositions match `backend/sql/seed.sql:4–39`, `frontend/e2e/playback.spec.ts:56–85`, `accessibility.spec.ts:294–370` and `learning.spec.ts:95–104`.

   **Required plan change:** none to this recipe/inventory; findings 1 and 12 concern the new journeys’ proof.

10. **NONBLOCKING — R1-10 CLOSED; case handling is also covered.**  
    **Location:** `plan-ms4.md:129–143`; `frontend/src/modules/playback/ui/LessonPlayer.tsx:589–591`, `:887–891`, `:1238–1240`; `frontend/src/modules/playback/application/preferences.ts:78–85`.

    The plan correctly preserves the **value** of `captionsOn` while allowing whole-preference writes. M7 explicitly includes selection and menu lookup, retaining manifest spelling for storage. Case-insensitive matching is safe for Ready tracks because `backend/migrations/00030_audio_track_language_shape.sql:88–89` enforces uniqueness on `lower(language)` for non-failed tracks.

    **Required plan change:** none.

11. **BLOCKER — M5 disables the primary-read branch promised by M4.**  
    **Location:** `plan-ms4.md:105`, `:126`; `frontend/src/app/composition/transcript-queries.ts:14–15`, `:30–36`; `frontend/src/routes/learn.$.tsx:180`.

    M4 requires a language-less primary read for non-ready lessons. M5 says query enablement requires a language. Applied literally, a lesson without video never fetches its existing transcript.

    **Required plan change:** define distinct query modes: disabled, primary read for eligible non-ready lessons, and exact-language read for predicted/confirmed tracks. Include an assertion that the primary request actually occurs.

12. **BLOCKER — E3 can pass without exercising remembered playback selection.**  
    **Location:** `plan-ms4.md:163–164`, `:199`, `:205`; `frontend/src/modules/playback/adapters/hls-media.ts:55–61`; `frontend/src/modules/playback/ui/LessonPlayer.tsx:548–550`.

    E3 specifies the predicted `de` panel before Play and absence of `en` requests/text. It never explicitly presses Play or requires `confirmed:de` and showing `de` captions. An implementation with a correct predictor but broken remembered hls selection can pass this row without loading a manifest.

    **Required plan change:** extend E3 through Play, initial non-default confirmation, and current-track language/mode/cues, retaining whole-transition request and DOM observation.

13. **NONBLOCKING — O7’s asset-selection rationale incorrectly rejects a reusable selector.**  
    **Location:** `plan-ms4.md:57–59`, `:91–96`; `backend/internal/course/adapters/postgres/learning.go:55–63`; `backend/internal/playback/adapters/postgres/playback.go:17–23`.

    For a Ready lesson, `current_video` already prefers non-evicted Ready assets and orders them by descending generation. Its selected asset therefore matches playback. Its additional behavior matters when no Ready asset exists.

    **Required plan change:** correct the factual claim. Permit extending that lateral selection with its asset id and Ready-track metadata. Add old-Ready/new-processing and evicted-newer cases to M03; retain the actor-dependent locked-output check.

14. **NONBLOCKING — the empty-track unit event does not reproduce hls.js startup behavior.**  
    **Location:** `plan-ms4.md:80`, `:196`; `frontend/node_modules/hls.js/src/controller/audio-track-controller.ts:160–162`.

    When there were and are no tracks, hls.js deliberately emits no `AUDIO_TRACKS_UPDATED`. A fake empty event can pass M01 while a genuinely empty initial manifest leaves the adapter pending indefinitely. This is outside the valid Ready-asset invariant, so it does not block ordinary playback.

    **Required plan change:** either define production handling for this unexpected initial state, or narrow M01’s claim to empty-list events actually emitted after a previously populated list.

VERDICT: BLOCKED