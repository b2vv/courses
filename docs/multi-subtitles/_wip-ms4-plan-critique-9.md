1. **NONBLOCKING — Round-8 finding 1 CLOSED: fingerprint registration precedes delivery.**

   **Location:** `docs/multi-subtitles/plan-ms4.md:267–280`; `frontend/node_modules/playwright-core/lib/coreBundle.js:59340–59347,59403`; `frontend/node_modules/hls.js/src/controller/buffer-operation-queue.ts:115–128`.

   Previously, hls.js could append `de` before asynchronous response observation registered its fingerprint. The revised route fetches the real bytes, awaits browser registration, then fulfils the intercepted request. Synchronous FNV matching preserves hls.js’s synchronous exception path.

   **Required plan change:** none.

2. **NONBLOCKING — Round-8 finding 2 remains CLOSED: recoverable exception and observable recovery.**

   **Location:** `plan-ms4.md:277–298`; `frontend/node_modules/hls.js/src/controller/error-controller.ts:477–504,635–641`; `frontend/src/modules/playback/adapters/hls-media.ts:76–80`.

   The two-argument `DOMException` supplies the recognized `InvalidStateError` name. The positively observed held retry, successful release and absence of the application’s terminal failure distinguish recovery from an unobserved or fatal failure.

   **Required plan change:** none.

3. **NONBLOCKING — Round-8 finding 3 remains CLOSED: failed payload identity is established.**

   **Location:** `plan-ms4.md:267–296`; `frontend/node_modules/hls.js/src/demux/mp4demuxer.ts:124–140`; `frontend/node_modules/hls.js/src/remux/passthrough-remuxer.ts:338–347,532–548`.

   The ordinary audio-only fMP4 path retains the `mdat` payload. Correlating its fingerprint with the identified `de` network path excludes the queued-`en` counterexample. Requiring the recovered payload’s matching fingerprint, overlapping interval and observed 880-Hz output completes the association.

   **Required plan change:** none.

4. **NONBLOCKING — Round-8 finding 4 CLOSED: observations cannot pass after the failed interval.**

   **Location:** `plan-ms4.md:286–298`; `frontend/node_modules/hls.js/src/controller/buffer-operation-queue.ts:115–139`; `frontend/node_modules/hls.js/src/controller/buffer-controller.ts:2043–2062,2168`.

   Failure observations now require three samples satisfying `start − 0.25 ≤ currentTime < end`; recovery requires at least one sample inside the recovered interval. This removes both the after-end false positive and vacuous recovery evidence. Holding the native append leaves completion pending until its eventual `updateend`.

   **Required plan change:** none.

5. **NONBLOCKING — Round-8 finding 5 CLOSED: EOS explanation is appropriately qualified.**

   **Location:** `plan-ms4.md:103–115`; `frontend/node_modules/hls.js/src/remux/passthrough-remuxer.ts:163–169`; `backend/internal/media/adapters/ffmpeg/transcoder.go:234–244`.

   The plan now explicitly limits its reasoning to the dedicated, single-audio-track SourceBuffer. It no longer claims that `SourceBuffer.buffered` universally avoids EOS extension. The video-tail counterexample remains excluded by reading the audio buffer.

   **Required plan change:** none.

6. **NONBLOCKING — Round-8 finding 6 CLOSED: APPENDED cannot erase an attempt’s error.**

   **Location:** `plan-ms4.md:84–102`; `frontend/node_modules/hls.js/src/controller/buffer-controller.ts:990–1007,1020–1071,2043–2084`; `frontend/node_modules/hls.js/src/controller/buffer-operation-queue.ts:84–89`.

   Quota recovery can execute an operation again without another APPENDING; asynchronous errors can still be followed by APPENDED. R9 distinguishes logical attempts from physical executions and retains the error flag through completion. It also explicitly requires listener-order-independent handling and nested-dispatch tests.

   **Required plan change:** none.

7. **NONBLOCKING — Round-8 finding 7 remains CLOSED: the other production boundaries remain covered.**

   **Location:** `plan-ms4.md:73–210,212–256,301–329`.

   Initial single/default/remembered selection can reach SWITCHED; cancellation need not, and SWITCHED can report the latest switching target rather than the playing fragment. M1 correctly avoids using it as confirmation (`frontend/node_modules/hls.js/src/controller/audio-track-controller.ts:182–199,254–257`; `audio-stream-controller.ts:792–807,1114–1118`). Reload resets selection and fragment state (`audio-track-controller.ts:70–76`; `base-stream-controller.ts:375–384`). Immediate suppression plus the explicit request-pending state covers parent D7.

   M4 uses `videoState`, rather than component presence, and distinguishes locked, email-gated and non-ready lessons (`frontend/src/routes/learn.$.tsx:82–90,276–307`; `LessonPlayer.tsx:370,853–855`). Ready publication creates a track transactionally, and deletion refuses the last Ready track (`backend/internal/media/adapters/postgres/publish.go:70–91`; `audiotrack.go:218–227`).

   M6 addresses replacement-track mode through its ref and language dependency; T0 removes the existing browser test’s direct mode assignment (`LessonPlayer.tsx:595–601`; `frontend/e2e/transcript.spec.ts:349–356`). M7 covers selection, menu identity and preference spelling; MS1’s index makes case-insensitive matching unambiguous (`backend/migrations/00030_audio_track_language_shape.sql:88–89`).

   Fixture generation/export, both required PATCH fields, CI selection and the named T0 sources remain concrete (`frontend/e2e/media.setup.ts:76–79,143–161`; `backend/internal/media/adapters/http/handler.go:129,367–370`; `frontend/playwright.config.ts:49–60`; `.github/workflows/ci.yml:114–115`). No additional D7 omission or blocking production-path gap was found.

   **Required plan change:** none. **Open blockers: 0.**

VERDICT: ACCEPTED