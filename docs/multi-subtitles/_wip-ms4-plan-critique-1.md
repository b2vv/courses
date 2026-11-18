1. **BLOCKER — SWITCHED alone cannot complete a cancelled switch back to the audible track.** Location: [plan-ms4.md:52](/Users/strelia/projects/courses/docs/multi-subtitles/plan-ms4.md:52), E4 at line 131. The production adapter uses `nextAudioTrack` (`frontend/src/modules/playback/adapters/hls-media.ts:98–105`). In hls.js, this calls `setAudioTrack(newId, false)` (`frontend/node_modules/hls.js/src/controller/audio-track-controller.ts:254–257`); the audio controller replaces `switchingTrack` (`audio-stream-controller.ts:485–506`), but emits SWITCHED only when the playing fragment changes **audio level** (`:792–807`).

   Counterexample: `en` is audible; select `de`, then select `en` before any `de` fragment becomes current. Subsequent fragments remain level `en`, so the condition for completing the switch never becomes true. M1 has already cleared captions; M3 can remain loading indefinitely.

   **Required plan change:** define cancellation back to the confirmed audible track, preserving D7’s confirmation requirement. Make E4 deterministically return before `de` becomes current, rather than allowing two completed switches.

2. **BLOCKER — M4’s claimed existing render boundary does not exist.** Location: [plan-ms4.md:92](/Users/strelia/projects/courses/docs/multi-subtitles/plan-ms4.md:92). `frontend/src/routes/learn.$.tsx:276–299` mounts `LessonPlayer` for every non-email-gated open lesson. The child decides whether to render media: `LessonPlayer.tsx:370` requires `videoState === 'ready'`, while `:853–855` renders empty/error states for other states. Using the route’s existing render condition therefore leaves non-ready lessons waiting for a callback that cannot arrive.

   Email gating is also not a successful primary-read case: `backend/internal/transcript/application/read.go:60–61` refuses unverified readers. Locked lessons bypass `OpenLesson` entirely (`learn.$.tsx:82–90`).

   **Required plan change:** specify an explicit matrix for locked, email-gated, non-ready and Ready lessons, including query enablement and panel state. Preserve the locked short-circuit; do not describe an email-gated transcript read as an existing working primary-transcript surface.

3. **BLOCKER — Native “unavailable” and “pending” are indistinguishable, and track removal retains stale confirmation.** Location: [plan-ms4.md:66](/Users/strelia/projects/courses/docs/multi-subtitles/plan-ms4.md:66), lines 73–85. M2 sends nothing when the API, enabled track or language is missing; M3 interprets the resulting `null` as loading. Thus native HLS without `audioTracks` never reaches D7’s required “none” state.

   Counterexample: native metadata initially confirms `en`; a `removetrack` leaves no enabled track, or the enabled track has an empty language. “Send nothing” retains the previous `en` confirmation and its transcript. The existing native adapter has no invalidation channel (`frontend/src/modules/playback/adapters/hls-media.ts:113–140`; callback contract in `ports/PlaybackMedia.ts:20–26`).

   **Required plan change:** distinguish pending confirmation from unavailable selection, and explicitly invalidate previously confirmed language on native removal/change and source reload. Test valid→missing and valid→empty transitions, not only initial absence.

4. **BLOCKER — Route confirmation needs lesson/session identity, not just a language string.** Location: [plan-ms4.md:72](/Users/strelia/projects/courses/docs/multi-subtitles/plan-ms4.md:72). `OpenLesson` is not keyed by lesson (`frontend/src/routes/learn.$.tsx:90`); only its player is keyed (`:282–285`). Route-owned caption language can therefore survive navigation.

   Counterexample: lesson A confirms `en`; navigate to lesson B, whose remembered selection will be `de`. Before B’s new player reports/reset effects run, the route can enable `learn.lesson(B, 'en')`, or render cached B/`en` data. Resetting after `load` is too late to prevent that render. The real reload paths also reuse the adapter (`LessonPlayer.tsx:295`, `:359`); hls.js clears selection and fragment state on manifest loading (`audio-track-controller.ts:70–76`, `base-stream-controller.ts:375–384`).

   **Required plan change:** bind confirmation to the current lesson and load generation, synchronously rejecting obsolete confirmation. Add navigation and same-instance retry/remint acceptance cases that observe requests and rendered text throughout the transition.

5. **BLOCKER — PATCH alone cannot repair the existing WebVTT journey: its ordering deadlocks.** Location: [plan-ms4.md:140](/Users/strelia/projects/courses/docs/multi-subtitles/plan-ms4.md:140). `frontend/e2e/transcript.spec.ts:342–344` opens the lesson and waits for transcript text; it clicks Play only at `:364`. Production starts the session on that click (`LessonPlayer.tsx:542–550`), then calls media `load` (`:295`). Before it, there is no manifest selection and consequently no SWITCHED.

   hls.js does support initial alternate-track confirmation: initial selection follows `AUDIO_TRACKS_UPDATED` (`audio-track-controller.ts:182–199`), and the first buffered fragment can complete the switch (`audio-stream-controller.ts:792–807`; startup handling in `base-stream-controller.ts:2692–2718`). It cannot do so before `loadSource`.

   **Required plan change:** update this journey to assert the initial loading/no-caption state, click Play, then await confirmation and transcript/cues. Explicitly cover the single-track initial case promised by M1.

6. **BLOCKER — The browser evidence can pass while captions are hidden or wrong during transitions.** Location: [plan-ms4.md:106](/Users/strelia/projects/courses/docs/multi-subtitles/plan-ms4.md:106), E1–E7 and M06. Re-keying plus the added language dependency addresses replacement, but the current effect controls `video.textTracks[0].mode` (`frontend/src/modules/playback/ui/LessonPlayer.tsx:595–601`). The existing browser anchor forcibly sets that mode to `showing` (`frontend/e2e/transcript.spec.ts:349–360`), bypassing the production preference effect.

   E2 checks the final panel; E3 observes panel text; neither proves D7’s absence of captions before confirmation or during the request. Parsed `cues` also do not prove that the replacement track is showing. Search/follow reset, explicitly required by parent D7, has no acceptance assertion.

   **Required plan change:** use the real caption switch and assert replacement-track language, mode and cues without assigning mode in the test. Observe both captions and panel across SWITCHING and a deliberately held real transcript request. Cover remembered on/off through “none” and back, plus search/follow reset. Define how E7 observes the real SWITCHED boundary—the Hls instance is currently private inside `hls-media.ts:51`.

7. **BLOCKER — The promised existing panel error state is absent.** Location: [plan-ms4.md:87](/Users/strelia/projects/courses/docs/multi-subtitles/plan-ms4.md:87), M6 at line 115. `frontend/src/modules/transcript/ui/TranscriptPanel.tsx:12–17` accepts only transcript, position and seek; `:40–46` implements an empty state, not an error state. The route currently discards query error/status (`frontend/src/routes/learn.$.tsx:180`). M6’s proposed union also omits `error`.

   Counterexample: a language-specific GET fails; no stated panel variant can render the table’s promised error result.

   **Required plan change:** add the error variant, query-error mapping, copy and recovery behavior explicitly; verify failure leaves captions absent while playback continues.

8. **NONBLOCKING — Close the zero-track hypothesis with the existing deletion guard.** Location: [plan-ms4.md:99](/Users/strelia/projects/courses/docs/multi-subtitles/plan-ms4.md:99). The last Ready track cannot be deleted through the production API: `backend/internal/media/adapters/postgres/audiotrack.go:218–227` authorizes/locks and calls `domain.ValidateTrackDeletion`; `domain/audiotrack.go:208–224` refuses when a Ready asset has no other Ready track. This is already exercised in `adapters/postgres/audiotrack_integration_test.go:423`.

   The assertion that transitions are *only* `Processing→AwaitingAudio→Ready` is also inaccurate: `domain/asset.go:153–154` allows `Processing→Ready`.

   **Required plan change:** replace the open hypothesis with the verified invariant and references. Treat an unexpected empty manifest as unavailable/error, not an ordinary reachable asset state requiring primary fallback.

9. **NONBLOCKING — Make T0 a fixture/impact inventory, and make the fixture recipe executable.** Location: [plan-ms4.md:124](/Users/strelia/projects/courses/docs/multi-subtitles/plan-ms4.md:124), lines 140–144. The proposed fixtures and CI route are feasible: `media.setup.ts:76–79` generates the named files, and `playwright.config.ts:49–60` selects the new spec while excluding only `asr.spec.ts` under `E2E_SKIP_ASR`.

   However, PATCH renames an **existing** track and requires both `language` and `label` (`backend/internal/media/adapters/http/handler.go:129`, `:367–370`); E5 needs a third uploaded/Ready track first. T0’s narrow searches also cannot establish indirect dependencies. Current dispositions differ: the seed has no lesson/transcript rows (`backend/sql/seed.sql:4–39`); accessibility uses a mocked Ready lesson without loading media (`accessibility.spec.ts:294–340`, `:359–370`); learning creates lessons without video (`learning.spec.ts:95–104`); playback creates separate media fixtures (`playback.spec.ts:56–85`).

   **Required plan change:** record explicit dispositions for these sources, specify third-track upload/readiness and the full PATCH body, and preserve existing non-caption assertions rather than mechanically PATCHing every match.

10. **NONBLOCKING — M06 overstates the preference-write restriction.** Location: [plan-ms4.md:110](/Users/strelia/projects/courses/docs/multi-subtitles/plan-ms4.md:110), acceptance line 155. Selecting audio updates `audioLanguage` (`frontend/src/modules/playback/ui/LessonPlayer.tsx:1238–1240`), which triggers the existing whole-preferences write (`:589–591`). Serialization includes `captionsOn` (`application/preferences.ts:78–85`), even when its value is unchanged.

    **Required plan change:** assert that “none” neither changes nor overwrites the remembered **value** of `captionsOn`. Do not require absence of storage writes, which would conflict with remembering the audio choice.

VERDICT: BLOCKED