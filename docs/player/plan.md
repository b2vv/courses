# Phase player-pro — plan

Owner-initiated, outside roadmap §24, 2026-10-23. Goal: raise `LessonPlayer` to the interaction
quality of YouTube / Apple TV / Hulu. Scope chosen by owner 2026-10-23: **A + B + C**, with a single
YouTube-style **gear (settings) menu** consolidating speed / quality / captions.

Baseline (already shipped, PR #63 `826b94b`): play/pause, scrubber + times, volume/mute, speed
select, quality select (HLS), PiP, fullscreen, captions `<track>`, keyboard (Space/k, ←/→ ±5s, f, m),
and the full session/heartbeat/recovery machinery. **None of that session logic may regress** — it is
the load-bearing part of the player and is covered by `recovery`/`playback`/`media` e2e projects.

## Reference behaviours (researched 2026-10-23)
- **YouTube**: auto-hide controls on idle while playing; buffered range behind the playhead; hover
  scrub thumbnail + time; keyboard J/L ±10s, 0–9 → %, Home/End, ↑/↓ volume, `,`/`.` frame step,
  `<`/`>` speed ±0.25, `c` captions, `f` fullscreen; single gear menu for speed/quality/captions.
- **Apple TV (AVKit)**: single tap shows/hides the transport; edge-click ±10s with an on-screen skip
  indicator; controls shown/hidden together; "Up Next" at the end.
- **Hulu**: 10s instant-replay; seek-preview; consolidated settings menu; adaptive bitrate; "Up Next"
  group / autoplay when a title ends.

## Slices

### A1 — Core viewing polish + gear menu (frontend only) → PR #1 (with A2)
- Auto-hiding control bar + cursor while playing (idle ~2.5s), reappear on pointer move / tap / focus;
  always visible while paused or when a control has focus (keyboard users never lose the bar).
- Buffered-range indicator painted behind the scrubber playhead (from `video.buffered`).
- Loading / stall spinner on `waiting`/`seeking`, hidden on `playing`/`canplay`.
- Expanded keyboard on the stage: J/L ±10s, Home/End, 0–9 jump to %, ↑/↓ volume ±5%, `,`/`.` frame
  step when paused, `<`/`>` speed ±0.25, `c` captions toggle (existing Space/k/←/→/f/m kept).
- Gestures: double-click frame → fullscreen; double-tap left/right third → ±10s with a skip indicator;
  single click still toggles play (disambiguate dbl vs single without delaying the first tap badly).
- Hover/drag time tooltip above the scrubber showing the time under the cursor.
- CC (captions) toggle control; Replay affordance when ended.
- Persist volume, muted, rate, captions-on, autoplay-next, quality-pref across lessons via
  `localStorage` (wrapped in try/catch; per-viewer convenience only).
- **Gear menu**: one settings button opens a menu (React Aria MenuTrigger/Popover) with Speed,
  Quality (HLS only), Captions submenus/sections. Replaces the two inline `<Select>`s. Keyboard- and
  SR-navigable; must not regress the 200%-zoom / phone layout the B24a fix made green.

### A2 — Up Next autoplay (frontend + route wiring) → PR #1 (with A1)
- Compute the next playable lesson from the course outline in `learn.$.tsx`; pass `nextLesson`
  (title/id) + `onPlayNext` callback into `LessonPlayer`.
- End-of-video "Next lesson" card with a countdown (respecting the autoplay-next pref) + Cancel +
  "Play now"; honours reduced-motion and the autoplay-off preference.
- No autoplay when there is no next lesson, or when the lesson failed / is locked.

### C1 — Storyboard generation (media-worker + backend) → PR #2
- Generate a sprite sheet (tiled thumbnails at a fixed cadence, e.g. 1 frame / N seconds, scaled) and a
  WebVTT storyboard mapping time ranges → sprite rectangles, during transcode.
- Write into the same protected media tree (0750/0640) so nginx serves it by group; persist the asset
  reference in the media/transcode schema; expose via the playback media path.

### C2 — Storyboard hover preview (frontend) → PR #2
- On scrubber hover/drag, read the storyboard VTT and render the thumbnail at the cursor time next to
  the time tooltip from A1.

## Verification & process
- Branch `feat/player-pro` off `main` @826b94b; branch+PR, owner self-merges A-batch after green gate.
- Authoritative gate = ordered `make check` / `make e2e-local` (CI is down account-wide; local is the
  gate). Player changes touch shared playback → **full** e2e set each slice, never a targeted run.
- Unit (rstest) for pure helpers + `LessonPlayer.test.tsx`; e2e `accessibility.spec` + `playback`
  projects for behaviour & a11y (200% zoom, keyboard, phone layout).
- `review_mode = independent_subagent` (fresh reviewer agent for the gate review; no second session
  assigned). Record writer/reviewer + next action in the Obsidian session note.
- Report: `docs/player/report.md` (created at close).

## Risks / watch
- The gear menu must preserve the B24a 200%/phone a11y win — re-run `accessibility.spec` B24a.
- Auto-hide must never hide the bar from keyboard/SR users or while paused.
- Session/heartbeat refs and the `onPlaybackChannel` contract (transcript seek) must not change shape.
- `video.buffered` / `waiting` events are flaky under HLS.js vs native — guard both `mediaKind`s.
