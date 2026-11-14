**BLOCKED — round 3, affected audible-switch proof only.** One open evidence blocker; SWITCHING correction is resolved.

`stage: plan` · Critic: Codex `/root`, selected independent context.

Plan: `docs/multi-audio/plan-s7.md`, SHA-256 `d982e9973f70f1b190f1919f96a1acbfeacc409f148993f6d08955171c50757f`. Target: branch `multi-audio-s7`, HEAD `3f0e1c84e97eedd6aa2537410803cc14165e5afe`.

Prompt: **Plan critic — version 1**, candidate `17a9527421e8fac3e43a16c287c404118699833c`, `prompts/plan-critic.md`; verified SHA-256 `514e4c376c2551afe76e390c5779867ecf7f04ad2a552c97c2b4f713c43dcab6`. Local package metadata declares `0.8.0-alpha.1+native-handoff.1`; release manifest declares `0.7.0-rc.5`. Full release integrity was not verified.

1. **BLOCKER — OPEN, missing evidence: MSE capture feasibility remains unresolved.**  
   [Plan:113](docs/multi-audio/plan-s7.md#L113) specifies `captureStream()` on the hls.js MSE video, but neither supplied evidence nor inspected local sources establishes that the target Playwright Chromium exposes its decoded audio through this route.

   **Counterexample:** video playback and hls switching work, but capture throws, contains no audio track, or supplies silence. The mandatory 440-Hz baseline correctly fails; the planned acceptance proof cannot finish. This is an unresolved platform dependency, **not a verified assertion that Chromium cannot capture MSE audio**.

   [Plan B:122](docs/multi-audio/plan-s7.md#L122) explicitly requires another critique before adoption, so it is not an accepted substitute yet. Resolve with authoritative compatibility evidence for the target browser, or a separately authorized feasibility result; otherwise complete and critique the fallback.

2. **NONBLOCKING — OPEN: specify the frequency discriminator more precisely.**  
   [Plan:114](docs/multi-audio/plan-s7.md#L114) now distinguishes persistent original audio from the dub: a genuine dominant 440-Hz baseline followed by dominant 880-Hz samples defeats round 2’s counterexample. Silence alone cannot legitimately satisfy those observations.

   **Counterexample:** accepting the largest bin without a signal floor can classify noise as a tone; accepting one frame can classify a transient as completed switching. Name frequency tolerance, minimum magnitude/prominence, sustained observation duration and the calibration rule for `N`. This measures captured decoded audio, not physical speaker output. Preserve the existing output path: do not connect the captured stream to `AudioContext.destination`.

3. **NONBLOCKING — OPEN: distinguish language code from helper argument.**  
   [Plan:107](docs/multi-audio/plan-s7.md#L107) correctly chooses supported code `en`, but [ASR helper:156](frontend/e2e/asr.spec.ts#L156) selects by accessible option **name**.

   **Counterexample:** `addAudioTrack(page, 'en', …)` cannot select the Ukrainian-labelled option. Existing [caller:441](frontend/e2e/asr.spec.ts#L441) passes `'Англійська'`; [UI:70](frontend/src/modules/media/ui/LessonAudioTracksField.tsx#L70) translates codes into labels. Record that invocation. API seed, UI upload and DELETE cleanup remain credible routes.

The SWITCHING finding is **resolved**: [plan:53](docs/multi-audio/plan-s7.md#L53) now matches controller `:339/:347`, including loaded-current-VOD early return and adapter no-op without a callback.

Compared with round 2, the frequency observation closes the *logical* “downloads while original keeps playing” gap, conditionally on a working capture route. Earlier exact-language and validation findings remain closed; no contract reversal was found. Plan B’s `AUDIO_TRACK_SWITCHED` marker would establish hls playback-switch completion, not independently measured audibility.

Both raw prior verdicts were read. During inspection, three playback files became dirty (`hls-media.ts`, `preferences.ts`, `PlaybackMedia.ts`); their implementation is excluded from this HEAD-bound planning verdict.

No tests, network probes, graph refresh or writes occurred. **0 additional reviewers launched / 0 returned**; the formal four-lens gate remains unverified. The record is returned here because the filesystem is read-only.