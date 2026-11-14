`stage: plan` · Critic: Codex `/root`, selected independent context · S7, first verdict.

Plan: [plan-s7.md](/Users/strelia/projects/courses/docs/multi-audio/plan-s7.md), SHA-256 `909e37f2f00da8e100f2e47a5dd5aeb19aa2e0bf618d15b442625b0242400bd5`.

Snapshot: `/Users/strelia/projects/courses`, branch `multi-audio-s7`, HEAD `3f0e1c84e97eedd6aa2537410803cc14165e5afe`. Plan is untracked; inspected tracked product sources and parent plan have no differences against HEAD. Other dirty documentation/configuration and historical WIP are excluded.

Prompt: **Plan critic — version 1**, candidate `17a9527421e8fac3e43a16c287c404118699833c`, `prompts/plan-critic.md`; verified SHA-256 `514e4c376c2551afe76e390c5779867ecf7f04ad2a552c97c2b4f713c43dcab6`. Package reports `0.8.0-alpha.1+native-handoff.1`, unreleased; package commit remains unverified.

Counterexamples below were **traced through source, not executed**.

1. **BLOCKER — OPEN, actionable: `{lang}` does not uniquely identify the requested track.**

   [Plan:29](/Users/strelia/projects/courses/docs/multi-audio/plan-s7.md:29) assumes DB uniqueness makes language a safe selector for [the proposed calls:43](/Users/strelia/projects/courses/docs/multi-audio/plan-s7.md:43). However, [hls.js `languagesMatch`:388](/Users/strelia/projects/courses/frontend/node_modules/hls.js/src/utils/rendition-helper.ts:388) accepts either string being a prefix of the other.

   **Counterexample:** current track `und-x-original`, user selects `und`. [setAudioOption:269](/Users/strelia/projects/courses/frontend/node_modules/hls.js/src/controller/audio-track-controller.ts:269) considers the current track a match and returns without switching. At construction, first-match search can likewise select `und` for a remembered `und-x-original`. This pair is real: [backfill:42](/Users/strelia/projects/courses/backend/migrations/00028_audio_track_embedded_backfill.sql:42) deliberately permits it. `en`/`en-us` creates the same problem, including selecting a regional track when the remembered exact language is absent.

   **Plan edit:** specify exact manifest-language resolution and an unambiguous hls selection option for both initial selection and later switches. Preserve default fallback when the exact language is absent. Add these counterexamples to verification.

2. **BLOCKER — OPEN, actionable: preference validation contradicts the accepted backend contract.**

   [Plan:51](/Users/strelia/projects/courses/docs/multi-audio/plan-s7.md:51) attributes its lowercase BCP-like regex to a DB CHECK. The actual [schema:13](/Users/strelia/projects/courses/backend/migrations/00025_multi_audio.sql:13) checks only length 1–35; [domain validation:267](/Users/strelia/projects/courses/backend/internal/media/domain/audiotrack.go:267) accepts ASCII letters of either case, digits and hyphens. [HTTP rename:264](/Users/strelia/projects/courses/backend/internal/media/adapters/http/handler.go:264) passes the value through without normalization.

   **Counterexample:** backend-valid `en-US` is selectable from the manifest but becomes `null` after preferences are read again, breaking R1.

   **Plan edit:** use validation compatible with existing accepted language values, preserving their identity. Tightening backend language policy would be a separate contract change outside S7.

3. **NONBLOCKING — OPEN, actionable: recovery and optimistic-selection behavior need explicit mechanisms and checks.**

   [Plan:57](/Users/strelia/projects/courses/docs/multi-audio/plan-s7.md:57) promises rollback after refusal, but the proposed method has no result, and hls does not emit a completion event for every invocation.

   **Counterexamples:** selecting the current option returns immediately; selecting an unavailable option can return `null` ([controller:269](/Users/strelia/projects/courses/frontend/node_modules/hls.js/src/controller/audio-track-controller.ts:269)). Recreating media after a user choice can also restore stale `initialPrefs` unless construction reads the current remembered preference ([player:183](/Users/strelia/projects/courses/frontend/src/modules/playback/ui/LessonPlayer.tsx:183)).

   **Plan edit:** distinguish confirmed selection, pending selection and remembered preference; define refusal/no-op handling. Cover same-instance reload, new session URL and instance recreation without making each audio choice recreate media.

4. **NONBLOCKING — OPEN, actionable: e2e setup, playback window and cleanup are underspecified.**

   [Plan:76](/Users/strelia/projects/courses/docs/multi-audio/plan-s7.md:76) names no overall timeout or cleanup. The short embedded-audio fixture exists, but lasts **two seconds** ([setup:67](/Users/strelia/projects/courses/frontend/e2e/media.setup.ts:67)); a user journey can reach the switch after playback ends. Existing upload helpers allow **240 seconds** ([ASR helper:149](/Users/strelia/projects/courses/frontend/e2e/asr.spec.ts:149)).

   **Counterexample:** opening settings after the two-second clip ends cannot prove uninterrupted switching through increasing `currentTime`.

   **Plan edit:** name the fixture, overall budget, controlled playback position and sufficient remaining duration; register request listeners before actions and clean up the published course. Require successful dub responses and confirmed playback, not requests alone.

5. **NONBLOCKING — OPEN, actionable: accessibility verification and the switching limitation need tightening.**

   [Plan:61](/Users/strelia/projects/courses/docs/multi-audio/plan-s7.md:61) relies on React Aria but lists no keyboard or 200% text-zoom check for the enlarged menu. Existing [RadioGroup:113](/Users/strelia/projects/courses/frontend/src/presentation/ui/controls.tsx:113) supplies suitable primitives; it does not prove the resulting panel fits.

   **Counterexample:** many tracks with long labels can overflow a small zoomed viewport. Add a browser check covering reachability, arrows, visible focus, Escape and focus return.

   Also, [plan:97](/Users/strelia/projects/courses/docs/multi-audio/plan-s7.md:97) should describe deferred forward-buffer replacement rather than promise a segment-sized switching bound: [audio-stream-controller:931](/Users/strelia/projects/courses/frontend/node_modules/hls.js/src/controller/audio-stream-controller.ts:931) still flushes future audio, and network delay is unbounded.

The six specific questions resolve as follows:

1. **Switching/persistence/first request:** `setAudioOption` avoids the immediate full flush, but performs a deferred forward flush. It writes configuration ([controller:264](/Users/strelia/projects/courses/frontend/node_modules/hls.js/src/controller/audio-track-controller.ts:264)), which [loadSource:537](/Users/strelia/projects/courses/frontend/node_modules/hls.js/src/hls.ts:537) preserves. Construction preference is resolved before initial playlist loading ([controller:166](/Users/strelia/projects/courses/frontend/node_modules/hls.js/src/controller/audio-track-controller.ts:166)), so an unambiguous `en` dub can be the first request. **Counterexample:** prefix collisions invalidate that general guarantee. `audioMatchPredicate` additionally checks codec/channels when supplied ([helper:407](/Users/strelia/projects/courses/frontend/node_modules/hls.js/src/utils/rendition-helper.ts:407)); surrounding matching also checks supplied identity attributes and associated language.

2. **Initial/reload events:** successful initial and reload playback can emit `AUDIO_TRACK_SWITCHED`; [fragment-change detection:792](/Users/strelia/projects/courses/frontend/node_modules/hls.js/src/controller/audio-stream-controller.ts:792) calls [completion:1114](/Users/strelia/projects/courses/frontend/node_modules/hls.js/src/controller/audio-stream-controller.ts:1114). **Counterexample:** paused playback or failed loading need not reach that point. Treat it as playback confirmation, not a guaranteed response to a selection call.

3. **Recovery:** same-instance media retry ([player:327](/Users/strelia/projects/courses/frontend/src/modules/playback/ui/LessonPlayer.tsx:327)) and session remint ([player:399](/Users/strelia/projects/courses/frontend/src/modules/playback/ui/LessonPlayer.tsx:399)) retain config preference. New lessons remount through [route key:285](/Users/strelia/projects/courses/frontend/src/routes/learn.$.tsx:285). **Counterexample:** same-lesson instance recreation using mount-time preferences loses a later selection; finding 3 covers this gap.

4. **Language identity:** exact language uniqueness holds among all non-failed rows, including in-flight uploads ([index:6](/Users/strelia/projects/courses/backend/migrations/00026_audio_track_language_unique.sql:6)). **Counterexample:** uniqueness does not prevent hls prefix matching; finding 1 applies.

5. **E2e/ASR:** the chromium project includes the new filename and depends on fixture setup ([config:49](/Users/strelia/projects/courses/frontend/playwright.config.ts:49)). Real upload→transcode→nginx→hls playback is feasible. One worker prevents simultaneous spec execution, but **counterexample:** background recognition survives the test that created it. New Ready lessons qualify automatically ([pending query:82](/Users/strelia/projects/courses/backend/internal/transcript/adapters/postgres/draft.go:82)); queue pause/resume is global ([ASR:37](/Users/strelia/projects/courses/frontend/e2e/asr.spec.ts:37)). Cleanup and avoidance of new global queue controls matter.

6. **Other surfaces:** SSR/storage guards already exist ([preferences:76](/Users/strelia/projects/courses/frontend/src/modules/playback/application/preferences.ts:76)). Labels render as React text, while backend rejects manifest-breaking quotes/control characters ([validation:282](/Users/strelia/projects/courses/backend/internal/media/domain/audiotrack.go:282)); **counterexample:** `<b>Dub</b>` should remain literal text. Absent-language fallback is undermined by finding 1. Heartbeat remains independent ([player:385](/Users/strelia/projects/courses/frontend/src/modules/playback/ui/LessonPlayer.tsx:385)); instance churn on every choice would unnecessarily disturb that lifecycle.

R1/R2 remain accepted owner decisions. No previous S7 verdict was found; parent/backfill/S6 evidence provides context without transferring acceptance. Reuse of the port, preferences and existing RadioGroup introduces no additional AI-slop or complexity finding.

**0 additional reviewers launched / 0 returned**, as instructed; no formal independent four-agent gate is claimed. Tests, network probes, graph refresh and writes were not performed. This record is returned here because the filesystem is read-only.

**BLOCKED — S7 at the exact plan and snapshot above. Two open product blockers: findings 1 and 2.**