`stage: plan` · Critic: Codex `/root`, selected independent context · Round 2.

Plan: `docs/multi-audio/plan-s7.md`, SHA-256 `dcc3e482337db97381349e922f64c3dc6c0ab8478b3558bde010d61fc905c31d`.

Snapshot: `/Users/strelia/projects/courses`, branch `multi-audio-s7`, HEAD `3f0e1c84e97eedd6aa2537410803cc14165e5afe`. Plan and round-1 artifact are untracked. Inspected tracked frontend, media backend, migrations and parent plan have no differences against HEAD. Other WIP is excluded.

Prompt: **Plan critic — version 1**, `prompts/plan-critic.md` under candidate `17a9527421e8fac3e43a16c287c404118699833c`; verified SHA-256 `514e4c376c2551afe76e390c5779867ecf7f04ad2a552c97c2b4f713c43dcab6`. Package declares `0.8.0-alpha.1+native-handoff.1`, unreleased; its commit remains unverified.

All counterexamples below are source traces, not executed tests.

1. **BLOCKER — OPEN: the e2e can pass without the dub playing.**

   [Plan:108–110](/Users/strelia/projects/courses/docs/multi-audio/plan-s7.md:108) requires dub downloads, selected radio and three seconds of advancing `currentTime`. It claims this proves audibility at [line 79](/Users/strelia/projects/courses/docs/multi-audio/plan-s7.md:79).

   **Counterexample tried:** the dub playlist and segment download successfully, selection changes, but the dub never replaces buffered original audio. The 30-second original can keep playing through the three-second assertion. Every listed observation passes.

   hls.js itself distinguishes receiving/buffering a track from switching playback: [audio-stream-controller:772](/Users/strelia/projects/courses/frontend/node_modules/hls.js/src/controller/audio-stream-controller.ts:772), [fragment-change detection:792](/Users/strelia/projects/courses/frontend/node_modules/hls.js/src/controller/audio-stream-controller.ts:792), and [completion:1114](/Users/strelia/projects/courses/frontend/node_modules/hls.js/src/controller/audio-stream-controller.ts:1114).

   **Required plan change:** specify a playback observation tied to the requested track—such as verified switch completion associated with that track, or detecting the 880-Hz output. Keep network and uninterrupted-progress assertions. Label click→first-segment timing as download latency; it does not measure audible-switch latency.

2. **NONBLOCKING — OPEN: “every `setAudioTrack` emits SWITCHING” is false.**

   [Plan:53](/Users/strelia/projects/courses/docs/multi-audio/plan-s7.md:53) overstates the guarantee.

   **Counterexample tried:** reselect the current, already-loaded VOD track. [Controller:339](/Users/strelia/projects/courses/frontend/node_modules/hls.js/src/controller/audio-track-controller.ts:339) returns before emitting. The adapter’s proposed no-op at [plan:60](/Users/strelia/projects/courses/docs/multi-audio/plan-s7.md:60) likewise emits nothing.

   Initial selection and an actual user switch emit synchronously at controller line 347. The proposed UI remains workable because no-op selection already has the correct radio state. Correct the wording and explicitly test no-op success without requiring a callback.

3. **NONBLOCKING — OPEN: name the route for creating `en-US`.**

   [Plan:102–105](/Users/strelia/projects/courses/docs/multi-audio/plan-s7.md:102) references ASR helpers but requires a regional language unavailable in the authoring picker.

   **Counterexample tried:** reuse [ASR `addAudioTrack`:153](/Users/strelia/projects/courses/frontend/e2e/asr.spec.ts:153) with `en-US`. Its option lookup cannot succeed: [offered languages:80](/Users/strelia/projects/courses/frontend/src/modules/media/domain/audio-track.ts:80) exclude it.

   This is feasible within scope: specify authenticated API mint/upload, or upload `en` through UI and PATCH its language using the existing [endpoint:124](/Users/strelia/projects/courses/backend/internal/media/adapters/http/handler.go:124). No product picker expansion is needed.

Round-1 dispositions:

| Finding | Round-2 disposition and counterexample checked |
|---|---|
| **B1** | **Resolved.** Exact index selection at plan:57–60 bypasses prefix matching. Traced `und`/`und-x-original`, reversed order and absent exact language; default fallback remains intact. |
| **B2** | **Resolved.** Plan:67–68 matches [backend validation:267](/Users/strelia/projects/courses/backend/internal/media/domain/audiotrack.go:267). `en-US` now survives parsing without normalization. |
| **N3** | **Resolved mechanism; wording correction remains finding 2.** Boolean refusal, retained `wanted`, current preference ref and no instance recreation address unavailable selection, reload and stale mount-time preferences. |
| **N4** | **Partially resolved; remaining proof gap is finding 1.** Thirty-second fixtures, controlled position, timeout, listeners and cleanup address the previous two-second scenario. Actual dub playback still lacks a distinguishing assertion. |
| **N5** | **Resolved as a verification plan.** Plan:111–116 now names keyboard, focus-return and zoom checks; §6 describes deferred flushing and network dependence. These checks remain unexecuted. |

The requested ordering checks support the revised mechanism:

- **Startup override:** hls constructs its internal listeners before the adapter ([hls.ts:304](/Users/strelia/projects/courses/frontend/node_modules/hls.js/src/hls.ts:304)). The internal `AUDIO_TRACKS_UPDATED` listener initializes stream levels without selecting a default ([audio-stream-controller:476](/Users/strelia/projects/courses/frontend/node_modules/hls.js/src/controller/audio-stream-controller.ts:476)). The adapter can then synchronously set `hls.audioTrack`; returning to [audio-track-controller:195–199](/Users/strelia/projects/courses/frontend/node_modules/hls.js/src/controller/audio-track-controller.ts:195), `selectedTrackId` prevents default selection/loading.
- **Manifest and level paths:** `MANIFEST_PARSED` stores audio tracks; autostart follows the manifest-processing listeners. Ordinary quality changes retain the same `group_aac`, so they do not re-enter group selection. Reload clears group/selection state and reapplies `wanted`. No competing default-selection path was found for the repo-generated manifests.
- **ASR claims:** API course/lesson seeding exists at ASR:69; DELETE cleanup expects 204 at ASR:250. Current Playwright collection sorts filenames; `asr.spec.ts` precedes `audio-tracks.spec.ts`, with one worker. This orders tests, not background jobs.
- **Fixture feasibility:** two additional small encodes fit the existing ffmpeg/container mechanism. Setup has a 720-second budget and 300-second compose-call limits. Source inspection supports feasibility; it cannot establish runtime headroom. Copy/export names must also be added.

R1/R2, frontend-only scope, manifest-based track discovery and existing port/preferences/RadioGroup reuse remain appropriate. No additional abstraction or complexity finding emerged.

**0 additional reviewers launched / 0 returned.** This selected single-critic review does not establish the separately required independent four-lens gate; that gate remains unverified. No tests, network probes, graph refresh or writes occurred. The record is returned here because the filesystem is read-only.

**BLOCKED — exact S7 round-2 plan and snapshot above. One open substantive blocker: finding 1.** B1/B2 are closed; the revised playback proof requires another affected-scope critique after correction.