1. **BLOCKER — R5-2 OPEN: E13 distinguishes media from init, but its recovery evidence can still pass vacuously.**

   **Location:** `docs/multi-subtitles/plan-ms4.md:248–260,280`; `frontend/node_modules/hls.js/src/controller/base-stream-controller.ts:1363–1378,2231–2270`; `src/controller/buffer-controller.ts:906–912,2043–2084`.

   The init/media fix is sound. The transcoder produces fMP4 (`backend/internal/media/adapters/ffmpeg/transcoder.go:231–247`), and hls.js reads init timescales and versioned `tfdt` fields (`src/utils/mp4-tools.ts:634–646,1256–1269`). These fields are available in the specified format. [ISO-BMFF media segments require `tfdt`.](https://www.w3.org/TR/mse-byte-stream-format-isobmff/#media-segments)

   Three gaps remain in the specified witness:

   - `{seq, kind, tfdt, threw}` establishes that an append call returned, not that MSE successfully processed it. hls.js explicitly handles asynchronous error→`updateend`; `updateend` alone also cannot establish success. [MSE specifies both events on append failure.](https://w3c.github.io/media-source/#sourcebuffer-append-error)
   - Raw `tfdt / timescale` is not necessarily the element’s timeline position. hls.js supplies `-initPTS` as `timestampOffset` immediately before appending. The proposed `[tfdt, tfdt+6)` check omits that offset; six seconds is also the transcoder’s segmentation target, not a parsed fragment duration.
   - “While the playhead is in that interval before recovery” has no required observation count. Recovery can finish before the playhead enters the interval, so this assertion can pass without checking any failed-attempt state. Later 880-Hz audio and German captions do not repair that missing observation.

   **Required plan change:** record native completion/error for the identified media operation, its timestamp offset and actual interval. Require a nonempty failed-state observation before recovery, followed by successful recovery of the identified fragment. Explicitly associate the injected append with `de`; timestamps alone distinguish positions, not languages. The original init-append counterexample is closed, but E13’s complete evidence obligation remains open.

2. **NONBLOCKING — R5-1 CLOSED for the production audio buffer; narrow the explanatory claim.**

   **Location:** `docs/multi-subtitles/plan-ms4.md:96–107,280`; `frontend/node_modules/hls.js/src/controller/buffer-controller.ts:1789–1808`; `src/remux/passthrough-remuxer.ts:163–169`; `backend/internal/media/adapters/ffmpeg/transcoder.go:234–244`.

   `BUFFER_CREATED.tracks.audio.buffer` is the live SourceBuffer, not a snapshot. This application packages one selected audio stream into its dedicated audio buffer. Consequently, the video buffer’s longer tail cannot extend that audio buffer’s ranges.

   However, “`SourceBuffer.buffered` has no EOS extension” is universally false: its getter extends constituent track ranges to the greatest end **within that SourceBuffer**. With this single audio track, that extension changes nothing. [MSE SourceBuffer attributes](https://w3c.github.io/media-source/#dom-sourcebuffer-buffered)

   **Required plan change:** state the single-track qualification. Retain the unequal-tail and removed-buffer checks; neither has been executed by this review.

3. **NONBLOCKING — R5-3 CLOSED: empty and errored fragments cannot independently establish coverage.**

   **Location:** `docs/multi-subtitles/plan-ms4.md:84–95,280`; `frontend/node_modules/hls.js/src/controller/base-stream-controller.ts:2413–2460`; `src/controller/buffer-controller.ts:1182–1193,1425–1465`.

   Empty parsing can reach `FRAG_BUFFERED`, but the proposed append requirement prevents it from relabelling retained audio. Fragment-associated errors prevent the attempted interval from becoming known.

   **Required plan change:** retain these guards; resolve the event-description issue in finding 19.

4. **NONBLOCKING — R5-4 CLOSED: E12 reserves its initiating failure before playback.**

   **Location:** `docs/multi-subtitles/plan-ms4.md:240–247`; `backend/internal/media/adapters/ffmpeg/transcoder.go:181–186`; `frontend/node_modules/hls.js/src/config.ts:455`; `frontend/src/modules/playback/ui/LessonPlayer.tsx:349–360`.

   Holding the fourth main fragment before Play allows earlier fragments to establish confirmation. Its single 503 reaches the existing same-instance reload path through `hls-media.ts:76–80` and `application/recovery.ts:123–130`.

   **Required plan change:** none; retain the explicit failure when the reserved request never arrives.

5. **NONBLOCKING — R5-5 CLOSED: generation comparison is scoped to the current mount.**

   **Location:** `docs/multi-subtitles/plan-ms4.md:154–162,281`; `frontend/src/routes/learn.$.tsx:90,282–285`; `frontend/src/modules/playback/ui/LessonPlayer.tsx:295,359`.

   A first event from B’s current mount is accepted independently of A’s higher generation. Returning A→B→A changes the proposed mount token.

   **Required plan change:** none.

6. **NONBLOCKING — R5-6 CLOSED: switching and held requests suppress captions consistently with D7.**

   **Location:** `docs/multi-subtitles/plan-ms4.md:109–113,167–175,215–217`; parent `plan.md:238–241`; `frontend/src/modules/playback/adapters/hls-media.ts:63–65,98–104`.

   A different selected language makes the state pending immediately; held transcript requests cannot preserve the previous language. The plan no longer relies on “hide on SWITCHING, trust SWITCHED.”

   **Required plan change:** none.

7. **NONBLOCKING — R5-7 CLOSED: the lesson-state boundary uses existing product facts.**

   **Location:** `docs/multi-subtitles/plan-ms4.md:139–151`; `frontend/src/routes/learn.$.tsx:82–90,276–307`; `frontend/src/modules/playback/ui/LessonPlayer.tsx:368–370`; `backend/internal/transcript/application/read.go:60–61`.

   Locked lessons bypass `OpenLesson`; email-gated playback is conditional; a mounted player does not imply available media. `videoState !== 'ready'` therefore correctly preserves primary transcript reading, while Ready lessons use prediction and subsequent confirmation.

   **Required plan change:** none.

8. **NONBLOCKING — R5-8 CLOSED: native confirmation can become unavailable again.**

   **Location:** `docs/multi-subtitles/plan-ms4.md:122–126,282`; `frontend/src/modules/playback/adapters/hls-media.ts:113–140`; `frontend/src/modules/playback/ports/PlaybackMedia.ts:20–26`.

   The proposed callback extension covers initial metadata, load invalidation and valid→missing/empty transitions. Unit-only native coverage matches the parent’s explicit limit.

   **Required plan change:** none.

9. **NONBLOCKING — R5-9 CLOSED: initial selection and reload behavior support the revised boundary.**

   **Location:** `docs/multi-subtitles/plan-ms4.md:45–52,73–120`; `frontend/node_modules/hls.js/src/controller/audio-track-controller.ts:182–199,254–257`; `src/controller/audio-stream-controller.ts:792–807,1114–1118`.

   One-track/default startup and synchronous remembered non-default selection can emit initial `AUDIO_TRACK_SWITCHED` when the playing fragment first changes from null. Ordinary `nextAudioTrack` switching can emit it on a playing-level change; rapid cancellation need not. It reports the current `switchingTrack`, without proving that track matches the playing fragment.

   Reload resets track selection and playing-fragment state (`audio-track-controller.ts:70–76`; `base-stream-controller.ts:375–384`; `hls.ts:537–562`). The revised plan correctly avoids using SWITCHED as its audible-language oracle.

   **Required plan change:** none.

10. **NONBLOCKING — R5-10 CLOSED: re-keying includes current-element mode control.**

    **Location:** `docs/multi-subtitles/plan-ms4.md:178–184,195–199,266,287`; `frontend/src/modules/playback/ui/LessonPlayer.tsx:595–601,989–997`; `frontend/e2e/transcript.spec.ts:349–356`; `frontend/src/modules/transcript/ui/TranscriptPanel.tsx:22–24`.

    Re-keying alone would leave index-zero control unsafe. The specified track ref and language dependency address that, while T0 removes test-side `mode` assignment. Panel search and follow reset are explicitly covered.

    **Required plan change:** none.

11. **NONBLOCKING — R5-11 CLOSED: transcript failure has an explicit recovery surface.**

    **Location:** `docs/multi-subtitles/plan-ms4.md:174,195–201,234–235`; `frontend/src/routes/learn.$.tsx:180`; `frontend/src/modules/transcript/ui/TranscriptPanel.tsx:12–17,40–46`.

    The proposed error state and retry callback replace the current discarded query error. E9 separately requires playback to continue.

    **Required plan change:** none.

12. **NONBLOCKING — R5-12 CLOSED: removing the last Ready track is refused.**

    **Location:** `docs/multi-subtitles/plan-ms4.md:55–58,119–120,305`; `backend/internal/media/domain/audiotrack.go:208–224`; `backend/internal/media/adapters/postgres/audiotrack.go:218–227`; `adapters/postgres/publish.go:70–91`.

    Production deletion validates the locked track set and refuses removal of the last Ready track. Video publication reaches Ready with an embedded track, otherwise AwaitingAudio. Zero-track Ready startup is therefore correctly treated as an unexpected state, rather than the normal non-ready fallback.

    **Required plan change:** none.

13. **NONBLOCKING — R5-13 CLOSED: fixture, PATCH, CI and T0 routes remain concrete.**

    **Location:** `docs/multi-subtitles/plan-ms4.md:203–212,263–274`; `frontend/e2e/media.setup.ts:76–79,143–161`; `backend/internal/media/adapters/http/handler.go:129,367–370`.

    The third tone includes generation and export. The existing PATCH API renames an existing track and requires both `language` and `label`.

    CI selects the new subtitle spec and conditionally excludes ASR (`frontend/playwright.config.ts:49–60`; `.github/workflows/ci.yml:114–115`). T0 correctly identifies seed data without lessons (`backend/sql/seed.sql:4–39`), independent playback fixtures (`playback.spec.ts:56–85`), mocked accessibility media (`accessibility.spec.ts:294–370`) and video-less learning lessons (`learning.spec.ts:95–104`).

    **Required plan change:** none to that inventory; E13 remains separately blocked.

14. **NONBLOCKING — R5-14 CLOSED: case-insensitive identity covers the remaining consumers.**

    **Location:** `docs/multi-subtitles/plan-ms4.md:178–193`; `backend/migrations/00030_audio_track_language_shape.sql:88–89`; `frontend/src/modules/playback/adapters/hls-media.ts:30–35,98–104`; `ui/LessonPlayer.tsx:887–891,1238–1240`.

    The index prevents case-only duplicates among non-failed tracks. M7 includes menu selection and adapter selection, preserves full-tag matching and stores manifest spelling. M6 preserves `captionsOn` through whole-preference writes (`application/preferences.ts:78–85`).

    **Required plan change:** none.

15. **NONBLOCKING — R5-15 CLOSED: primary requests remain independently enabled.**

    **Location:** `docs/multi-subtitles/plan-ms4.md:148–150,175`; `frontend/src/app/composition/transcript-queries.ts:14–15,30–36`.

    Explicit off/primary/exact modes preserve language-less reading for non-ready lessons. The language-required condition applies only to exact mode.

    **Required plan change:** none.

16. **NONBLOCKING — R5-16 CLOSED: E3 proves remembered playback beyond prediction.**

    **Location:** `docs/multi-subtitles/plan-ms4.md:218–220`; `frontend/src/modules/playback/adapters/hls-media.ts:55–61`; `ui/LessonPlayer.tsx:548–550`; `frontend/e2e/audio-tracks.spec.ts:164–201`.

    E3 presses Play and requires German confirmation, independently measured 880-Hz audio and showing German cues. It also monitors unwanted English requests and text from navigation onward.

    **Required plan change:** none.

17. **NONBLOCKING — R5-17 CLOSED: backend prediction selects the playback asset.**

    **Location:** `docs/multi-subtitles/plan-ms4.md:128–137,283`; `backend/internal/course/adapters/postgres/learning.go:55–63`; `backend/internal/playback/adapters/postgres/playback.go:17–23`.

    `current_video` agrees with playback whenever a Ready non-evicted asset exists. M03 covers newer processing/evicted assets, Ready-track filtering and actor-dependent locked output.

    **Required plan change:** none.

18. **NONBLOCKING — R5-18 CLOSED: empty-manifest evidence remains correctly narrowed.**

    **Location:** `docs/multi-subtitles/plan-ms4.md:119–120,305`; `frontend/node_modules/hls.js/src/controller/audio-track-controller.ts:160–162`.

    Empty startup emits no `AUDIO_TRACKS_UPDATED`. The plan explicitly specifies indefinite pending for this unexpected state instead of using a fabricated empty-list event as startup evidence.

    **Required plan change:** none.

19. **NONBLOCKING — R5-19 OPEN: both inaccurate event descriptions remain unchanged.**

    **Location:** `docs/multi-subtitles/plan-ms4.md:84–95`; `frontend/node_modules/hls.js/src/controller/buffer-controller.ts:990–1007,2043–2084`; `src/controller/buffer-operation-queue.ts:84–89`.

    Quota recovery requeues the existing append operation without another `BUFFER_APPENDING`. Asynchronous SourceBuffer failure invokes `onError`, then `updateend` invokes `onComplete`, which emits `BUFFER_APPENDED`. Thus APPENDING is not emitted before every physical execution, and APPENDED is not exclusively a success event.

    **Required plan change:** distinguish logical fragment attempts from physical executions, retain attempts across quota requeue, and explicitly test error→APPENDED→FRAG_BUFFERED without reopening success. The combined error guard can remain correct; its current explanation is not.

VERDICT: BLOCKED