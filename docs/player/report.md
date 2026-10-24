# Phase player-pro — report (A-batch: A1a, A1b, A2)

Owner-initiated, outside roadmap §24 (2026-10-23). Raises `LessonPlayer` to the interaction quality
of YouTube / Apple TV / Hulu. This report covers the frontend A-batch (PR #1); the storyboard slice
(C, media-worker + backend) is a separate PR and is **not** in this branch.

## What shipped

- **A1a — rich player chrome** (`969997b`): auto-hiding control bar + cursor while playing (idle
  2.5 s, always up while paused or focused), buffered-range bar behind the playhead, stall spinner,
  end replay, double-click gestures (edge ±10 s with a skip badge, centre fullscreen), hover/drag
  scrub-time tooltip, expanded stage keyboard (J/L ±10 s, Home/End, 0–9 %, ↑/↓ volume, `,`/`.` frame
  step, `<`/`>` speed, `c` captions), and per-viewer persisted volume/mute/rate/captions
  (`application/preferences.ts`, localStorage, guarded).
- **A1b — gear settings menu** (`e9fbd61`): one React Aria gear popover folds speed / quality (HLS) /
  captions into radio groups + a switch, replacing the two inline `<Select>`s; keyboard- and
  SR-navigable; preserves the B24a 200 %-zoom / phone layout.
- **A2 — Up Next autoplay** (`66ea493`): `nextLesson(course, lessonId)` pure domain helper (outline
  order, crossing module boundaries; the route decides playability from `locked`/`videoState`, never
  re-derived); an end-of-video card with the next lesson title, a 5 s countdown, **Play now** and
  **Cancel**/**Replay**; an `autoplayNext` preference (default on, persisted) with a gear switch shown
  only when a next lesson exists; the route keys the player by lesson so advancing (autoplay or the
  outline) remounts a clean player. Countdown is started by the ending event (no setState-in-effect),
  parks at zero, and a sibling effect fires the advance; reduced motion yields the card without the
  automatic jump; autoplay-off and Cancel both stop a pending advance.
- **P01 test fix** (`7a7ad00`): the gear quality is clicked by `label.ui-radio` (filtered by text),
  not the hidden `role=radio` input React Aria overlays — see §Limitations / lessons.

## Verification

- `make e2e-local` ran to completion in the manifest order with one rebuild of the e2e overlay:
  **124/124 e2e passed (17.8 m)**, `MAKE_EXIT=0`. This includes the full playback / recovery / media
  projects and `accessibility.spec` (B24a: gear keyboard-visible and usable at 200 % zoom and phone
  layout, real browser).
- Frontend `format` / `typecheck` / `lint` clean; **798 frontend unit tests pass** — new: 3 `nextLesson`
  domain tests, 3 A2 player tests (end-card offered + Play now advances + axe-clean; autoplay fires on
  countdown; Cancel stops it leaving Replay), plus the `autoplayNext` preference round-trip/default.
- Browser design-pass (real Chrome, localhost:8080 e2e stack, student session): the A1 chrome and the
  A1b gear popover (speed + quality radios) render cleanly; a single-lesson course correctly shows no
  autoplay switch and no Up Next card.

## Limitations / lessons

1. **The Up Next card was not screenshotted live.** No seeded or surviving e2e fixture has a course
   with two `ready` lessons (all six ready `video_assets` are the sole lesson of their course), so the
   card cannot be reached in the running stack, and fabricating a second ready lesson would prove a
   fake. The card is covered by unit tests (behaviour + axe) and by the e2e coverage of the gear and
   controls that surround it. A multi-ready-lesson playback fixture would close this; filed as the one
   visual-verification gap of this batch.
2. **A real-browser defect hid behind a green unit test.** A1b's updated P01 clicked
   `getByRole('radio').click()`; React Aria hides the native radio inside the visible label that
   overlays it, so Playwright hit the label (the input's ancestor) and refused with "intercepts
   pointer events" — a 7-minute timeout. jsdom does not model hit-testing, so the unit `user.click`
   was a false green, and A1b had been committed without a full e2e. Fixed by clicking the label, the
   convention `ratings.spec.ts` already documented. `pointer-events:none` on the decorative
   `.ui-radio-mark` does **not** help — the label itself is the interceptor.
3. **Reduced motion** suppresses the automatic advance (the card still offers Play now), decided at
   mount via `matchMedia`; a live media-query change mid-lesson is not observed (the player is
   per-lesson and remounts on navigation).
4. Storyboard hover-preview (C) is out of scope here and ships as a separate PR (media-worker sprite +
   WebVTT + `video_assets` persistence + the frontend hover thumbnail).

## Next

- Browser design-pass of the Up Next card once a two-ready-lesson fixture exists.
- C1/C2 storyboard slice as PR #2.
