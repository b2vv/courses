1. **BLOCKER — R3-1 OPEN: absence of the three listed errors does not establish successful audio coverage.**

   **Location:** `docs/multi-subtitles/plan-ms4.md:83–90,254`; `frontend/node_modules/hls.js/src/controller/base-stream-controller.ts:2413–2460`; `frontend/node_modules/hls.js/src/controller/buffer-controller.ts:1182–1193,1437–1465`.

   The specified `ERROR → FRAG_BUFFERED` ordering is correct for the append exceptions at `buffer-controller.ts:1015–1071`. However, the whitelist omits other paths that still emit `FRAG_BUFFERED`: no-media parsing falls through to `FRAG_PARSED`, and append-no-progress emits `bufferAppendNoProgress`.

   Counterexample: an `en` interval remains buffered; the selected `de` fragment yields no audio samples. The audio controller performs its switching flush only when remuxed audio exists (`audio-stream-controller.ts:930–943`). Nevertheless, `FRAG_BUFFERED(audio)` can arrive. M1 records that interval as `de`, permitting `de` captions over surviving `en`.

   **Required plan change:** define affirmative audio-coverage criteria, including empty/gap/partial and no-progress outcomes, rather than infer success from this error whitelist. Scope failure markers to an append/load attempt so a successful retry can recover. Add corresponding real-adapter evidence.

2. **BLOCKER — R3-17 OPEN: invalidating only at flush announcement allows removed intervals to become “known” again.**

   **Location:** `plan-ms4.md:91–95,254`; `frontend/node_modules/hls.js/src/controller/buffer-controller.ts:1114–1133,1146–1153,2287–2293`; `frontend/node_modules/hls.js/src/controller/buffer-operation-queue.ts:24–49`.

   Counterexample: an old audio append and its fragment blocker are queued before a flush. `BUFFER_FLUSHING` immediately clears the map, but the old append subsequently completes and emits `FRAG_BUFFERED`, restoring its label. The queued removal then executes and emits `BUFFER_FLUSHED`, which M1 ignores. Seeking into that removed interval can therefore report confirmation from stale map coverage.

   There are also removals without this announcement: quota eviction directly queues `getFlushOp` (`buffer-controller.ts:990–1006`).

   **Required plan change:** specify pending-removal ordering and successful completion reconciliation, including completion-only removals. A pre-removal append must not resurrect coverage after removal. Add this interleaving to M01. The corrected null payload, reset/detachment invalidation, paused recalculation and conditional timing description are otherwise sound (`types/events.ts:112–124`; `buffer-controller.ts:496–596`).

3. **BLOCKER — R3-4 OPEN: E12 names a real recovery trigger but does not ensure that trigger occurs.**

   **Location:** `plan-ms4.md:227–235`; `frontend/node_modules/hls.js/src/config.ts:455`; `frontend/src/modules/playback/ui/LessonPlayer.tsx:349–360`.

   A fragment 503 does reach the stated same-instance retry path through `hls-media.ts:76–80` and `application/recovery.ts:123–130`. However, E12 fails only a request **issued after** confirmation. The 30-second fixture fits the default 30-second forward-buffer target; a resumed position near the end may require only the final fragment. There need be no subsequent media request after confirmation. No 503 then occurs, and neither E12 nor dependent E14 exercises recovery.

   **Required plan change:** reserve a specific media request before Play, hold its response until initial confirmation, then return exactly one 503. Specify fresh/resumed progress deliberately. Observe the second master request and pending→confirmation through the production retry.

4. **BLOCKER — New: E13’s unspecified corruption does not prove an MSE append failure.**

   **Location:** `plan-ms4.md:232–234,254`; `backend/internal/media/adapters/ffmpeg/transcoder.go:241–247`; `frontend/node_modules/hls.js/src/demux/transmuxer.ts:473–486`; `frontend/node_modules/hls.js/src/remux/passthrough-remuxer.ts:329–335`.

   These fixtures produce fMP4. Corruption can fail probing or produce no media samples before `SourceBuffer.appendBuffer` is reached. That produces parsing/no-media behavior, not necessarily `BUFFER_APPEND_ERROR` (`demux/transmuxer-interface.ts:330–348`). A captions-absent assertion could therefore pass without exercising the append-error map rule.

   The existing adapter also exposes no observation of a nonfatal, status-less Hls error (`frontend/src/modules/playback/adapters/hls-media.ts:76–80`). The selection attribute cannot establish the failure stage.

   **Required plan change:** provide a deterministic fMP4 corruption recipe and passive evidence that it reaches the intended real MSE/Hls error path. Establish its actual recovery behavior before requiring a later successful fragment; do not substitute synthetic adapter callbacks.

5. **NONBLOCKING — R3-2 CLOSED: generation comparison is now scoped to mount identity.**

   **Location:** `plan-ms4.md:141–149,255`; `frontend/src/routes/learn.$.tsx:90,282–285`; `frontend/src/modules/playback/ui/LessonPlayer.tsx:295,359`.

   The first B event is accepted despite A’s larger generation; comparisons apply only within one mount. This closes A-generation-2→B-generation-1 logically. E14’s browser proof remains dependent on fixing E12 above.

   **Required plan change:** none beyond finding 3.

6. **NONBLOCKING — R3-3 CLOSED: switching/request-in-flight behavior now agrees with D7.**

   **Location:** `plan-ms4.md:99–100,157,202–204`; parent `docs/multi-subtitles/plan.md:238–241`; `frontend/src/modules/playback/adapters/hls-media.ts:63–65,98–104`.

   E2 now requires absent captions and a loading panel from the switch while the new-language request is held. The previous explicit retention of old-language text is removed.

   **Required plan change:** none; the map defects above remain separate.

7. **NONBLOCKING — R3-5 CLOSED: lesson-state boundaries use existing authoritative facts.**

   **Location:** `plan-ms4.md:126–138`; `frontend/src/routes/learn.$.tsx:82–90,276–299`; `frontend/src/modules/playback/ui/LessonPlayer.tsx:370,853–855`; `backend/internal/transcript/application/read.go:60–61`.

   Locked lessons bypass the surface, email-gated reads are disabled, non-ready lessons retain primary reads, and Ready lessons use prediction/selection. No impossible media callback is required for a non-ready lesson.

   **Required plan change:** none.

8. **NONBLOCKING — R3-6 CLOSED: native confirmation can be invalidated.**

   **Location:** `plan-ms4.md:109–113,256`; `frontend/src/modules/playback/adapters/hls-media.ts:113–140`; `frontend/src/modules/playback/ports/PlaybackMedia.ts:20–26`.

   Explicit pending/unavailable states and valid→missing/empty transitions replace the current adapter’s silence. Missing API or language no longer retains an earlier confirmation.

   **Required plan change:** none.

9. **NONBLOCKING — R3-7 CLOSED: initial selection and reload facts support the revised design.**

   **Location:** `plan-ms4.md:12–19,45–52,200–207,219–220,240`; `frontend/node_modules/hls.js/src/controller/audio-track-controller.ts:182–199`; `frontend/node_modules/hls.js/src/controller/audio-stream-controller.ts:792–807`.

   Single/default selection and a remembered non-default selected synchronously during `AUDIO_TRACKS_UPDATED` can produce initial SWITCHED when the playing audio fragment changes from no previous fragment. Reload resets selection and fragment state (`audio-track-controller.ts:70–76`; `base-stream-controller.ts:375–384`; `hls.ts:537–562`).

   SWITCHED still cannot reliably confirm rapid cancellation; r4 correctly stops relying on it. O7 and moving the caption anchor after Play close the original deadlock.

   **Required plan change:** none.

10. **NONBLOCKING — R3-8 CLOSED: current-track mode and follow reset are specified.**

    **Location:** `plan-ms4.md:165–170,183–186,198–199,261–263`; `frontend/src/modules/playback/ui/LessonPlayer.tsx:595–601,992–997`; `frontend/e2e/transcript.spec.ts:349–356`.

    The ref targets the replacement element instead of index zero; the effect gains the language dependency, and browser tests stop assigning mode. M09 now explicitly includes both search and follow reset, whose existing state lives at `TranscriptPanel.tsx:22–24`.

    **Required plan change:** none.

11. **NONBLOCKING — R3-9 CLOSED: transcript errors have an implementable state and recovery action.**

    **Location:** `plan-ms4.md:161,183–187,221–222`; `frontend/src/modules/transcript/ui/TranscriptPanel.tsx:12–17,40–46`; `frontend/src/routes/learn.$.tsx:180`.

    The proposed error variant, retry callback and playback-continuation test supply behavior absent from the existing component and discarded query status.

    **Required plan change:** none.

12. **NONBLOCKING — R3-10 CLOSED: zero Ready tracks is outside the ordinary asset lifecycle.**

    **Location:** `plan-ms4.md:55–58,106–107`; `backend/internal/media/domain/audiotrack.go:208–224`; `backend/internal/media/adapters/postgres/audiotrack.go:218–227`.

    Production deletion refuses the last Ready track. Video publication enters Ready with an embedded track in the same transaction, otherwise AwaitingAudio (`media/adapters/postgres/publish.go:70–91`). This supports the narrowed unexpected-empty-manifest limitation.

    **Required plan change:** none.

13. **NONBLOCKING — R3-11 CLOSED: fixtures, PATCH, CI selection and T0 dispositions are concrete.**

    **Location:** `plan-ms4.md:190–199,237–248`; `frontend/e2e/media.setup.ts:76–79,143–161`; `backend/internal/media/adapters/http/handler.go:129,367–370`.

    The third fixture now includes export, and PATCH supplies both required fields after upload/readiness. The new spec is selected by Chromium; CI skips ASR conditionally (`frontend/playwright.config.ts:49–60`; `.github/workflows/ci.yml:114–115`).

    T0 correctly distinguishes seed data without lessons (`backend/sql/seed.sql:4–39`), separate playback fixtures (`playback.spec.ts:56–85`), mocked accessibility media (`accessibility.spec.ts:294–370`) and video-less learning lessons (`learning.spec.ts:95–104`).

    **Required plan change:** none to this inventory; E12/E13 remain separate blockers.

14. **NONBLOCKING — R3-12 CLOSED: case-insensitive matching is safe and covers the remaining consumers.**

    **Location:** `plan-ms4.md:165–180`; `backend/migrations/00030_audio_track_language_shape.sql:88–89`; `frontend/src/modules/playback/adapters/hls-media.ts:30–35,98–104`; `frontend/src/modules/playback/ui/LessonPlayer.tsx:887–891,1238–1240`.

    The index prevents case-only duplicates among non-failed tracks. M7 includes menu lookup and selection, preserves manifest spelling for preference storage, and M6 preserves the `captionsOn` value through whole-preference writes (`application/preferences.ts:78–85`).

    **Required plan change:** none.

15. **NONBLOCKING — R3-13 CLOSED: primary reads are no longer disabled by the exact-language rule.**

    **Location:** `plan-ms4.md:135–137,162`; `frontend/src/app/composition/transcript-queries.ts:14–15,30–36`; `frontend/src/routes/learn.$.tsx:180`.

    Separate off/primary/exact modes preserve the language-less request, and “enabled only with language” is now explicitly scoped to exact mode.

    **Required plan change:** none.

16. **NONBLOCKING — R3-14 CLOSED: E3 exercises remembered playback rather than prediction alone.**

    **Location:** `plan-ms4.md:205–207`; `frontend/src/modules/playback/adapters/hls-media.ts:55–61`; `frontend/src/modules/playback/ui/LessonPlayer.tsx:548–550`; `frontend/e2e/audio-tracks.spec.ts:164–201`.

    E3 explicitly presses Play and requires confirmed `de`, independently observed 880-Hz audio, and showing `de` cues while checking the entire navigation for unwanted `en` requests/text.

    **Required plan change:** none.

17. **NONBLOCKING — R3-15 CLOSED: the selector contradiction is removed.**

    **Location:** `plan-ms4.md:115–124,257`; `backend/internal/course/adapters/postgres/learning.go:55–63`; `backend/internal/playback/adapters/postgres/playback.go:17–23`.

    Extending `current_video` reuses the selector that already matches playback whenever a Ready asset exists. The acceptance row covers old-Ready/new-processing, evicted-newer, Ready-track filtering and actor-dependent locked output.

    **Required plan change:** none.

18. **NONBLOCKING — R3-16 CLOSED: empty-manifest evidence is correctly narrowed.**

    **Location:** `plan-ms4.md:106–107,279`; `frontend/node_modules/hls.js/src/controller/audio-track-controller.ts:160–162`.

    The plan acknowledges that initial empty audio inventory emits no `AUDIO_TRACKS_UPDATED`, and explicitly records indefinite pending outside the Ready-asset invariant. It no longer claims a fake empty-list event reproduces startup.

    **Required plan change:** none.

VERDICT: BLOCKED