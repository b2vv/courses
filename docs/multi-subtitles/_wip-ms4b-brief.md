WORKING FILE: docs/multi-subtitles/_wip-ms4b.md
Append AFTER EVERY STEP (append, not rewrite). The first entry comes before the first tool:
a `NEXT:` line stating what you intend to check. Then use the labels FACT / DECISION / NEXT / QUESTION.
Final answer = a concise summary of the file.

# MS4 run (b) — playback port, hls.js + native adapters, case-insensitive S7 memory

Repository `/Users/strelia/projects/courses`, branch `ms4-player-captions`. You are the only writer in
this working tree for the files below (the owner checked: no other session is writing). Do not commit,
do not stage, do not switch branches, never git checkout/reset/stash/clean. Uncommitted MS4a changes
(backend course, openapi, playback schema/domain) are already in the tree — keep them.

Contract: `docs/multi-subtitles/plan-ms4.md` (ACCEPTED r9). Read §2, §3 M1, M2, M7 and §4 rows M01,
M01b (reducer is NOT in this run), M02, M04 (predictedAudioLanguage), M08 in full before coding. The plan
cites hls.js 1.7.2 source lines in `frontend/node_modules/hls.js/src` — read them; the rules there are
exact (logical append attempt, error flag never cleared by APPENDED, empty attempt keeps the label,
coverage = labels ∩ audio SourceBuffer.buffered from BUFFER_CREATED, everything unknown on
BUFFER_RESET / MEDIA_DETACHING / load(), recompute after every map change and on the listed media events,
selection rule (a)(b)(c) where (c) "transcript arrived" belongs to the route — the adapter reports
(a)+(b) only).

## Scope of this run
1. `frontend/src/modules/playback/ports/PlaybackMedia.ts`: add
   `onCaptionSelection(selection: CaptionSelection)` with
   `CaptionSelection = {state:'pending'} | {state:'confirmed'; language: string} | {state:'unavailable'}`.
   `confirmed` means: latest selected track language = L AND audio under the playhead is known to be L.
   Keep `onAudioTrack` (menu value, S7) unchanged.
2. `adapters/hls-media.ts`: implement the audio map and selection reporting per M1 as a small pure,
   separately testable unit (e.g. `audio-coverage.ts` with no hls.js import: feed it events, query it with
   currentTime + buffered ranges) plus thin wiring in `createHlsMedia`. Only public hls.js events/fields.
   `load()` emits `pending` synchronously before `loadSource`. Emit only on change.
3. Native adapter per M2 (`pending` on load; `confirmed`/`unavailable` from `video.audioTracks` on
   `loadedmetadata`, `change`, `addtrack`, `removetrack`; no API → `unavailable`).
4. M7: `exactAudioTrackIndex` compares `toLowerCase()` on both sides; no prefix matching; rewrite
   `hls-media.test.ts:27-30` accordingly. Add `predictedAudioLanguage(tracks, remembered)` (M3/M4) in the
   playback application or domain layer, sharing the same matching function, so the route (run c) and the
   adapter pick the same start track: remembered (case-insensitive) if present, else default, else first,
   else null.
5. `LessonPlayer.tsx`: only wire the new callback through as an optional prop
   `onCaptionSelection?: (event: {generation: number; selection: CaptionSelection}) => void` where
   `generation` increments on every `media.load(...)` call in this mounted player. No captions/track/panel
   behaviour change in this run (that is run c). Update fakes in `LessonPlayer.test.tsx` for the new port.

## Tests (part of this run)
- Unit for the coverage unit: every M01 rule, including: two physical executions in one attempt (quota
  requeue without new APPENDING); asynchronous error then APPENDED in the same attempt → no label;
  ERROR dispatched inside APPENDED handling (nested); empty attempt keeps the previous label;
  flush-after-append removed via buffered intersection; audio shorter than video at EOS → tail unknown;
  SourceBuffer access throwing InvalidStateError → all unknown; reset/detach/load → all unknown;
  recompute while paused when the map changes; cancellation en→de→en with en under playhead → confirmed en
  without SWITCHED; buffered de under playhead while selection is en → pending.
- Unit for createHlsMedia wiring with a fake Hls emitter (load→pending first; confirmed only on change).
- Unit native adapter: initial, valid→missing, valid→empty, no API, load→pending (fake AudioTrackList).
- Unit predictedAudioLanguage + M7 table (shared cases, case-insensitive, no prefix).

## Verify before finishing
`cd frontend && npm run format && npm run format:check && npm run typecheck && npm run lint && npx rstest run src/modules/playback src/routes`.
Report each command and its real result; a skipped check is not a pass.

## Two channels
- Missing or contradictory contract → STOP, QUESTION in the working file.
- A plan statement that is inaccurate in a way that would cause a bug (verify against the hls.js source) →
  argue it with evidence (DECISION) instead of silently following it.
