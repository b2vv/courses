# D5 + TYPE — evidence and interface-review (diff mode)

| | |
|---|---|
| Date | 2026-11-22 |
| Branch | `design-pass-d5-type` from `main` @ `ab29bace` (plan §12, owner decision §9 2026-11-22) |
| Stack | Kiri, `http://localhost:8080`. The frontend is rebuilt from the branch after every edit, and every capture comes from the latest image. |
| Data | Seed plus the admin fixture account (`frontend/e2e/auth.ts`). Read-only: only the trend metric switch was clicked. |
| Captures | `/admin/analytics` at phone 390, tablet portrait 768, tablet landscape 1024, desktop 1440, 390 with 200% text, and 320. TYPE: the S1 probe on 27 surfaces × 6 viewports. Captures are in the session scratchpad. |

## What changed

- **TYPE.** At 390 px:
  - `h1` is 30 px (was 38.4);
  - the home hero is 33 px (was 40);
  - `h2` is about 21 px (was 25.6);
  - the admin `h1` is 24 px (was 28.8).

  At 1440 px the sizes match the old ones within 2 px. At 768 px the `h1` is 41 px (was 38.4).
- **D5.** The page now reads top to bottom:
  - four headline figures;
  - the trend;
  - «Склад платформи», six cards built from `SegmentedBar`: a thin bar over a list of parts with figures and shares;
  - «Залученість»;
  - a short note on deferred metrics.

  Two decisions inside the cards:
  - The «Акаунти» card splits live accounts (active plus blocked), the same figure as «Усього акаунтів». Deleted accounts, verified email and 2FA appear under the bar.
  - On a single-instructor platform the «Ролі» card shows no total, bar or shares, because the instructor count is withheld (admin-stats 9a).

## Checks

| Check | Result | Evidence |
|---|---|---|
| Horizontal overflow, D5 × 6 viewports | PASS | `ui-measurements` fit `:root` x: 1440/390/390-200% PASS; capture script `ov=0` on all 6 |
| Overflow, TYPE × 27 surfaces × 6 | PASS | S1 probe: 0 |
| Words broken by the narrow column, 27 surfaces | 75 → **57** | S1 probe, sum of phone + 200% phone + 320 + desktop |
| Contrast, text | PASS: axe `color-contrast` 0 violations (wcag2a/aa/21aa/22aa) on desktop, phone, phone 200%. The lowest new ratio is 6.13:1 (`.ui-split-name`, `.ui-split-share`, meter labels, notes). | axe in real Chromium. `ui-measurements` `contrast` → NOT_TESTED (`non-opaque-or-non-srgb`: the text sits on transparent wrappers), so no hand-computed numbers. |
| Contrast, selected segment (gradient) | NOT_TESTED | axe `incomplete`, a pre-existing primitive |
| Colour of the bar segments | separation PASS (CVD ΔE ≥ 10.9, normal ≥ 18) for every card palette | dataviz `validate_palette.js --mode dark --surface #111317 --pairs all`. Lightness band FAIL for the status tokens and the de-emphasis grey; their meaning is fixed by the project, and every part also has a text label. |
| Trend switch, keyboard | PASS: arrows move focus (outline solid), Space selects and the chart changes title | key presses in Chromium at 390 |
| Trend switch, pointer and size | PASS: 172×44 on phone, 67–118×44 on desktop; a click on «Дохід» → chart «Дохід» | bounds + click |
| Unit | 286/286 (admin + presentation/ui) | `rstest` |
| Full ordered `make check`, e2e `admin-analytics.spec.ts` | NOT_TESTED | waiting for the owner's go-ahead |

## Lenses

**anti-slop §3 pre-flight on the result:**
- Theme and palette: the admin dark-only theme with its signal tokens. Two new greys are `--color-viz-neutral`/`muted`; students use the existing `--color-viz-purple`.
- Decoration: none. Every bar shows a figure from the list.
- Copy: reread in full. «stats-ендпоінта» is gone.
- Placeholders and fabricated data: none.
- Motion: none.

**ux-review, in this pass's diff:** three findings, all fixed and rechecked.

1. **Major · [copy] Consistent terminology / predictability.**
   - Before: «Усього акаунтів 7» and the «Акаунти» card showing 8 (deleted counted in) on one screen.
   - Fix: the card's whole is the same 7, and deleted accounts sit under the bar.
   - Test: `splits live accounts to the same total as the headline tile…`.
2. **Minor · [law] Common Region.** Verified email and 2FA sat under «Залученість», although they describe account security. They moved to the «Акаунти» card.
3. **Minor · dataviz status reserved / [consistency] Distinguishable roles.**
   - Before: students used the status blue that means «Чернетки» on the neighbouring card.
   - Fix: they now use the data-viz hue `series` (`--color-viz-purple`).

Fixed during implementation:
- the meter label overflowed at 390 with 200% text (19 px);
- the «Дохід» option was hidden behind the scroll;
- the chart was 60 px tall on a phone;
- verified/2FA were counted against active accounts only, while the backend counts all non-deleted ones (`snapshot.go:75-76`).

**Left as taste:**
- On desktop the grid rows have equal height, so «Ролі» keeps empty space below its two rows.
- «Залученість» is a small card: average progress «—» and completions 0 on the seed.

**Pre-existing, outside the diff:**
- the selected trend segment (gradient) cannot be measured by axe;
- the sidebar ends at the viewport height in a full-page capture;
- «Виміряно на 0 активах».

## Owner edits 1–5 (owner 2026-11-22: «Do all 5 items»), commit `526f8841` + gated hyphenation

Recheck on image `f8b142eb` (the container was confirmed running this image; the earlier `up` had failed to unpack on a 98% full Kiri disk, and the owner allowed a `kiri-build` cache prune of 21 GB, leaving 22 GB free).

1. The split cards align to the top: «Ролі» stays compact.
2. «Середній прогрес» and «Завершені курси» move into the headline tiles (six tiles on desktop, three rows of two on a phone). The «Залученість» section is removed.
3. The trend plot on analytics is `clamp(9rem, 14vw, 12rem)`, about 200 px at 1440 instead of about 240.
4. Media: «— немає виміряних відео» instead of «— Виміряно на 0 активах». With measured videos the text reads «виміряно на N відео».
5. C3 hero: padding is `clamp(16px, 5vw, 64px)` instead of a rem clamp. The title column at 390 with 200% text is 287 px (was 230). Syllable hyphenation for the title switches on only through `@container (max-width: 12em)`, about 8em of width at 200% against 19em at 100%:
   - at 200%, «фун-дамент»;
   - at 100%, no hyphens (`hyphens: manual` measured).

   This is not a React Aria tab, so the 2026-11-18 `@container` gotcha does not apply.

Results:

| Check | Result |
|---|---|
| Overflow, D5 × 6 viewports and 27 surfaces × 6 | 0 |
| axe contrast, desktop / phone / phone 200% | 0 violations |
| Broken words, 27 surfaces | 57 → **51**; the C3 title at 200% now breaks with a hyphen at a syllable (the probe still counts it) |
| Unit | 286/286 |

## Gate and review

- **Codex, round 1** (`d7e40fff`): **BLOCK**, two findings:
  - P2: `ul` with `list-style:none` loses list semantics in WebKit;
  - P3: no test pinned the withheld roles card.

  Codex also corrected plan D5-3: the denominator is active + blocked.

  Fixed in `a55326e6`: `role="list"` with a lint suppression and its reason, plus a test that fails if a total, bar or shares appear on the roles card.
- **Codex, round 2:** **PASS**, no new findings. Safari/VoiceOver was not checked: there is no such environment.
- **Ordered `make check` on `a55326e6`**, Kiri, `COMPOSE_FILE=docker-compose.yml:<kiri>/local-runtime.override.json`, 10:58–11:40 UTC:
  - result: **GATE_EXIT=0**;
  - frontend **1502/1502**;
  - Go **130 `ok` / 0 `FAIL`**;
  - e2e **168 passed** (24.2 min);
  - the frontend container ran image `8be1c3d1…`, built in this run at 11:16 UTC.

  The first run on `d7e40fff` was stopped at the `integration` step on purpose, to restart on the commit with the review fixes.

## Not checked

- Non-empty orders and media: the seed has zeros. The unit tests cover the bar with figures and the empty whole.
- Real browser zoom.
- Ordered `make check` with the full e2e, and the Codex review.
