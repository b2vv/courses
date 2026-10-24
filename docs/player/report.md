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
   card cannot be reached in the running *dev* stack, and fabricating a second ready lesson would
   prove a fake. The card is covered by unit tests (behaviour + axe) and by the e2e coverage of the
   gear and controls that surround it. **Resolved in the rc.4 viewport pass** (see §Viewport coverage):
   the playback e2e seed's T7 fixture *does* have two `ready` lessons (lesson 3 → lesson 4), so the card
   was captured and viewed live at desktop, tablet-portrait and tablet-landscape.
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

## Viewport coverage (rc.4 review, 2026-10-24)

Post-merge viewport-matrix pass of the changed player surfaces (control bar, gear popover,
Up Next card) per `~/.agents/skills/design-pass/references/viewport-coverage.md`, on
main @ `bc5c4d1`. Captures made via a temporary VPMATRIX block in `playback.spec` reusing the
T7 fixture (lesson 3 `ready` → lesson 4 `ready`); both the temp spec block and the config dep
tweak were reverted, frontend tree clean. 16 screenshots in `/tmp/vp-shots`; PW 13/13 passed.

| Class (CSS px) | Control bar | Gear popover | Up Next card |
| --- | --- | --- | --- |
| Phone 390×844 | PASS | PASS | PASS (minor: eyebrow slightly clipped by short frame — see below) |
| Tablet portrait 768×1024 | PASS | PASS | PASS |
| Tablet landscape 1024×768 | PASS | PASS | PASS |
| Desktop 1440×900 | PASS | PASS | PASS |
| Reflow 320 | PASS | PASS | — |
| 200% text (`font-size:200%`) | PASS | PASS¹ | — |

All 16 screenshots were viewed. The captured `text200-s2-gear.png` opened the popover off the
zoomed frame, so it was **re-verified live** (real Chrome, localhost:8080 e2e seed, student
session, `font-size:200%` on root): the gear popover reflows to full width below the control bar,
the speed/quality radios and the autoplay toggle stay readable, and the popover scrolls so no
action is lost. ¹ → PASS.

This closes the Up Next live-capture gap in Limitation 1 for the three viewport widths above
(the T7 two-ready-lesson fixture does exist in the playback e2e seed).

**One minor, non-blocking finding:** at phone 390, the Up Next card's "НАСТУПНИЙ УРОК" eyebrow is
slightly clipped by the short video frame; the title and both buttons (Відтворити зараз / Переглянути
знову) remain fully visible and usable. **Owner decided (2026-10-24): fix it** — adjust the card's top
inset / overlay height on narrow frames so the eyebrow is not clipped. Queued for the follow-up session
(deferred off this session by the AI-Factory size guard).

## Next

- **Fix the phone-390 Up Next eyebrow clip** (owner-approved 2026-10-24): on narrow/short video
  frames the end-card overlay crops the "НАСТУПНИЙ УРОК" eyebrow — adjust the card's top inset or the
  overlay's vertical fit so the full eyebrow shows. Small frontend CSS change; re-verify at phone 390.
- C1/C2 storyboard slice as PR #2.
