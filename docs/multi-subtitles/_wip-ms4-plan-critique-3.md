1. **BLOCKER — R2-1 OPEN: `FRAG_BUFFERED` does not prove the proposed language interval.**

   **Location:** `docs/multi-subtitles/plan-ms4.md:77–89,230`; `frontend/node_modules/hls.js/src/controller/buffer-controller.ts:1174–1193,1420–1428`.

   The normal alternate-audio event shape is correct: `id = frag.type = 'audio'`; the playlist parser assigns the track index to `frag.level` (`loader/playlist-loader.ts:543–554`, `loader/m3u8-parser.ts:360–369`), and track ids match `hls.audioTracks` indices (`controller/audio-track-controller.ts:156–158`).

   The unsupported step is treating every such event as successful replacement of the whole interval. An append can fail, mark its cycle `errored`, and advance the operation queue (`buffer-controller.ts:1015–1071`; `buffer-operation-queue.ts:119–129`). `checkAppendProgress()` returns for that error, but its caller still emits `FRAG_BUFFERED`.

   Counterexample: `en` remains buffered under the playhead; a queued `de` append fails recoverably. M1 nevertheless replaces the interval’s label with `de`, allowing `de` captions over surviving `en` audio. The production adapter ignores nonfatal errors without an HTTP status (`frontend/src/modules/playback/adapters/hls-media.ts:76–80`).

   **Required plan change:** define how successful audio coverage and replacement are established, including failed/partial appends. `FRAG_BUFFERED` alone must not establish coverage. Add browser evidence through the real adapter and MSE failure path; a fake-Hls interval test cannot prove this inference.

2. **BLOCKER — R2-4 OPEN: generation comparison crosses mount identities.**

   **Location:** `plan-ms4.md:125–132,234`; `frontend/src/routes/learn.$.tsx:90,282–285`; `frontend/src/modules/playback/ui/LessonPlayer.tsx:295,359`.

   The new token fixes the original A→B→A collision, but the stated reducer compares `event.generation` with the single retained record without first checking **that record’s mount**.

   Counterexample: A has retried and its record is `{mount:A, generation:2}`. Navigate to B, whose first load reports `{mount:B, generation:1, pending}`. The event’s mount matches the current mount, but `1 < 2`, so both its pending and subsequent confirmation are rejected. B remains on prediction after playback starts.

   **Required plan change:** accept the first event for a different current mount; compare generations only within the same mount. Add A-generation-2→B-generation-1 coverage. The mount fix is useful, but this finding remains open.

3. **BLOCKER — New: M1/E2 explicitly contradict D7’s request-in-flight state.**

   **Location:** `plan-ms4.md:85–86,184–187`; parent `docs/multi-subtitles/plan.md:238–240`; `frontend/src/modules/playback/adapters/hls-media.ts:63–65,98–104`.

   E2 deliberately holds the new-language request while requiring old `en` captions and panel text to remain visible. D7 explicitly requires no captions and a loading panel while the new-language request is in flight. O7 changes the **before-Play** behavior; it does not authorize this switching exception.

   Calling the request a prefetch does not, by itself, reconcile the two observable requirements. E2 currently proves behavior opposite to the accepted parent.

   **Required plan change:** make M1 and E2 preserve D7’s loading/absence requirement during the language transition. Do not introduce an exception through the slice’s implementation description.

4. **BLOCKER — New: E12 never triggers the retry it claims to prove.**

   **Location:** `plan-ms4.md:209–211`; `frontend/src/modules/playback/ui/LessonPlayer.tsx:349–360`; `frontend/src/modules/playback/application/recovery.ts:123–130`; `frontend/node_modules/hls.js/src/hls.ts:537–562`.

   After `confirmed:en`, installing a 503 route for the master manifest does not cause another master request. The VOD master has already loaded. The player calls `load()` again only after a production error/recovery action; E12 specifies no initiating error.

   The journey therefore stalls before exercising generation advancement, or must introduce an unstated direct call that bypasses the intended production path.

   **Required plan change:** specify a deterministic initiating failure on a real media request, then observe the resulting same-instance `loadSource` retry. Register interception before that request can finish, and assert pending plus recovery without directly invoking adapter callbacks.

5. **NONBLOCKING — R2-2 CLOSED: the lesson-state boundary is explicit.**

   **Location:** `plan-ms4.md:110–122`; `frontend/src/routes/learn.$.tsx:82–90,276–299`; `frontend/src/modules/playback/ui/LessonPlayer.tsx:370,853–855`; `backend/internal/transcript/application/read.go:60–61`.

   Locked, email-gated, non-ready and Ready lessons now have distinct query/render behavior. Non-ready lessons no longer wait for a callback from media that cannot exist.

   **Required plan change:** none.

6. **NONBLOCKING — R2-3 CLOSED: native invalidation is specified.**

   **Location:** `plan-ms4.md:94–98,231`; `frontend/src/modules/playback/adapters/hls-media.ts:113–140`; `frontend/src/modules/playback/ports/PlaybackMedia.ts:20–26`.

   Pending and unavailable are now distinct; removal, empty language and unavailable API invalidate previous confirmation. The valid→missing and valid→empty tests address the original stale-language counterexamples.

   **Required plan change:** none.

7. **NONBLOCKING — R2-5 CLOSED: the pre-Play deadlock remains resolved.**

   **Location:** `plan-ms4.md:11–18,188–190,201–202,216`; `frontend/src/modules/playback/ui/LessonPlayer.tsx:548–550`; `frontend/e2e/transcript.spec.ts:342–364`.

   O7 preserves reading before Play, while the WebVTT anchor moves after Play. hls.js supports initial alternate-audio confirmation for a single/default track and a synchronously remembered non-default: selection follows `AUDIO_TRACKS_UPDATED` (`audio-track-controller.ts:182–199`), and the first playing audio fragment can emit SWITCHED (`audio-stream-controller.ts:792–807`). Source reload resets selection and playing-fragment state (`audio-track-controller.ts:70–76`; `base-stream-controller.ts:375–384`).

   This does not make SWITCHED reliable for cancellation; finding 1 remains separate.

   **Required plan change:** none for this finding.

8. **NONBLOCKING — R2-6 CLOSED: current-track control and follow reset are covered.**

   **Location:** `plan-ms4.md:148–169,180–187,216`; `frontend/src/modules/playback/ui/LessonPlayer.tsx:595–601,992–997`; `frontend/src/modules/transcript/ui/TranscriptPanel.tsx:22–24,72–75,89–91`.

   Re-keying plus a ref to the current `<track>` addresses the existing index-zero effect. The browser anchor no longer assigns mode, and M8/E2 explicitly cover both search and follow reset.

   **Required plan change:** add “follow” to acceptance row M09 at `plan-ms4.md:238` so the summary matches the detailed requirement.

9. **NONBLOCKING — R2-7 CLOSED: error behavior is implementable.**

   **Location:** `plan-ms4.md:144,165–171,203–204`; `frontend/src/modules/transcript/ui/TranscriptPanel.tsx:12–17,40–46`; `frontend/src/routes/learn.$.tsx:180`.

   The new error variant, retry action and playback-continuation journey supply behavior absent from the current component and route.

   **Required plan change:** none.

10. **NONBLOCKING — R2-8 CLOSED: zero Ready tracks is not an ordinary production state.**

    **Location:** `plan-ms4.md:54–57`; `backend/internal/media/domain/audiotrack.go:208–224`; `backend/internal/media/adapters/postgres/audiotrack.go:218–227`; `backend/internal/media/adapters/postgres/audiotrack_integration_test.go:421–423`.

    Production deletion refuses the last Ready track of a Ready asset. The plan also correctly acknowledges direct `Processing→Ready` (`backend/internal/media/domain/asset.go:153–154`).

    **Required plan change:** none.

11. **NONBLOCKING — R2-9 CLOSED: PATCH, fixture roles and T0 dispositions are concrete.**

    **Location:** `plan-ms4.md:173–181,213–224`; `backend/internal/media/adapters/http/handler.go:129,367–370`; `frontend/e2e/media.setup.ts:76–79`.

    PATCH renames an existing track and requires both language and label; the third track now has an upload/readiness recipe. The new spec is selected by Chromium, and CI’s ASR exclusion does not exclude it (`frontend/playwright.config.ts:49–60`; `.github/workflows/ci.yml:114–115`).

    The inventory correctly distinguishes seed data without lessons (`backend/sql/seed.sql:4–39`), mocked accessibility media (`frontend/e2e/accessibility.spec.ts:294–370`), video-less learning lessons (`frontend/e2e/learning.spec.ts:95–104`) and separate playback fixtures (`frontend/e2e/playback.spec.ts:56–85`).

    **Required plan change:** for the new 1320-Hz fixture, also name the export list at `frontend/e2e/media.setup.ts:143–161`. Adding only the ffmpeg line creates bytes inside the worker, not the local upload file.

12. **NONBLOCKING — R2-10 CLOSED: case-insensitive identity and preference preservation are covered.**

    **Location:** `plan-ms4.md:149–163`; `backend/migrations/00030_audio_track_language_shape.sql:88–89`; `frontend/src/modules/playback/adapters/hls-media.ts:30–35,98–104`; `frontend/src/modules/playback/ui/LessonPlayer.tsx:887–891,1238–1240`; `frontend/src/modules/playback/application/preferences.ts:78–85`.

    The index makes case-insensitive matching unambiguous for Ready tracks. M7 includes menu lookup and selection, and preserves manifest spelling for storage. M6 correctly preserves the boolean value of `captionsOn` while allowing whole-preference writes.

    **Required plan change:** none.

13. **NONBLOCKING — R2-11 CLOSED: primary reads now have their own query mode.**

    **Location:** `plan-ms4.md:119–121`; `frontend/src/app/composition/transcript-queries.ts:14–15,30–36`; `frontend/src/routes/learn.$.tsx:180`.

    Explicit `off`, `primary` and `exact` modes, with an assertion that the primary request occurs, close the missing-video regression.

    **Required plan change:** scope the remaining “enabled only with a language” sentence at `plan-ms4.md:145` explicitly to `exact` mode.

14. **NONBLOCKING — R2-12 CLOSED: E3 now exercises remembered playback.**

    **Location:** `plan-ms4.md:188–190`; `frontend/src/modules/playback/adapters/hls-media.ts:55–61`; `frontend/src/modules/playback/ui/LessonPlayer.tsx:548–550`; `frontend/e2e/audio-tracks.spec.ts:164–201`.

    E3 now presses Play, requires `confirmed:de`, independently observes 880-Hz audio, and checks showing `de` cues. It can no longer pass on prediction alone.

    **Required plan change:** none.

15. **NONBLOCKING — R2-13 CLOSED behaviorally; contradictory selector prose remains.**

    **Location:** `plan-ms4.md:58–61,101–106,232`; `backend/internal/course/adapters/postgres/learning.go:55–63`; `backend/internal/playback/adapters/postgres/playback.go:17–23`.

    The corrected facts and M03 permit extending `current_video` and cover old-Ready/new-processing plus evicted-newer selection. That selector already chooses the playback asset whenever a Ready asset exists.

    **Required plan change:** remove M3’s remaining “not the sorting `current_video`” exclusion at line 103; it conflicts with M03’s stated implementation.

16. **NONBLOCKING — R2-14 CLOSED: the empty-manifest claim is narrowed.**

    **Location:** `plan-ms4.md:91–92,230,254`; `frontend/node_modules/hls.js/src/controller/audio-track-controller.ts:160–162`.

    The plan no longer claims an initial empty `AUDIO_TRACKS_UPDATED` event. It explicitly records indefinite pending for this out-of-invariant state.

    **Required plan change:** none.

17. **NONBLOCKING — New: flush handling and the timing limitation need precise semantics.**

    **Location:** `plan-ms4.md:79–88,252–253`; `frontend/node_modules/hls.js/src/types/events.ts:112–124`; `controller/buffer-controller.ts:1098–1153,590–611`.

    `BUFFER_FLUSHING` announces a queued operation, not completed removal; `type:null` flushes all buffers. `BUFFER_FLUSHED` uses different range fields and can report an error with a zero range. Removing intervals at both events can discard valid coverage when removal fails. Buffer reset also removes SourceBuffers without being an ordinary range flush.

    Separately, recalculation only on media events leaves a paused player’s selection stale when buffer contents change. The claimed 15–250-ms interval is conditional on normal playback and event-handler execution, not a general wall-clock guarantee. [HTML’s time-marches-on algorithm](https://html.spec.whatwg.org/multipage/media.html#time-marches-on).

    **Required plan change:** specify successful flush completion, null-type flushes, reset/detachment, and recalculation after buffer-map mutations. Describe timing as a measured observation with stated conditions.

VERDICT: BLOCKED