# admin-editor-ux — report (2026-11-18)

Owner-initiated phase outside §24. Trigger: the owner's 2026-11-15 complaint ("виглядає все всрато"), which led to the
audit `docs/ux-review/admin-course-editor-2026-11-15/report.md` (B1, M1–M6, m1–m4). The fix was deferred to after MS4. The
owner started it on 2026-11-18 ("Берися за UX-фікс адмін-редактора").
- Plan: `plan.md`, ACCEPTED at r3 after r1 BLOCKED (3 blocking findings) and r2 BLOCKED (1). Codex critic, four lenses, raw
  log `_wip-plan-critique.md`.
- Branch `admin-editor-ux` from `main` @ `10436cdf`.

## 1. Done (slice → audit findings → rows)
| Slice | Commit | Closes | Rows |
|---|---|---|---|
| U1 | `973a7789` | B1: only the selected tab panel is displayed. m3: the admin 404 leads to `/admin`; unknown `/admin/*` paths no longer render an empty page (found in the browser). | Ux01–Ux03 |
| U2 | `524fd7e6` | M1: collapsible lesson rows; the header keeps title, preview, video and transcript badges; the body stays mounted. M5: one h4 lesson-section header. M6: "Edit lesson content" link. | Ux04–Ux06, Ux13 |
| U3 | `6708f735` | M2: compact transcript line with a read-only `m:ss.s` readout. Raw-ms storage is unchanged (D6). | Ux07–Ux08 |
| U4 | `10c3c6ea` | M4: icon reorder, the impossible direction is not rendered, D3 focus rules, confirmed delete of saved rows. m1: materials form behind a toggle. m2: section eyebrows removed. m4: video file name. | Ux09–Ux10, Ux14–Ux16 |
| U5 | `a09d960f` | M3: sticky Save bar with "Unsaved changes"; "saved immediately" hints. No Discard (D7). | Ux11–Ux12 |
| U6 fixes | `ea7af7d8` | Design-pass findings F1–F3 (§3). | Ux07, Ux11, Ux17 |
| U6 evidence | `4436ac04` | Design-pass harness, measurements and screenshots in `design-pass/`. | Ux17 |

Every slice had an independent Codex review (spec + standards) that ended SHIP on both axes. The raw review logs are
`_wip-u*-review.md` and `_wip-dp-fix-review.md`.

## 2. Verification (measured)
- **Unit:** 1482 tests, 0 failed, 0 skipped (`npx rstest run src`, after the F1 proof moved to e2e).
- **Full e2e after `e2e-reset` + seed, on U1:** 167 passed.
- **Full e2e after `e2e-reset` + seed, on U2 (pre-T1 image):** 168 passed.
- **Affected specs on each later final image:**
  - U2 after T1: admin 8, audio-tracks 5, subtitles 15, media 6, playback 26.
  - U3: transcript 3, asr 21.
  - U4: admin 8, media 6.
  - U5: admin 8, purchase 1, enrollment 8, transcript 3.
  - F1: transcript.spec red on the pre-fix CSS at the Ukrainian tab click, green with the fix
    (`design-pass/fix-owner-verification.json`).
- **Design-pass (Ux17):** 8 states × 7 viewports (390/768/1024/1440 at 100 %, 1440 and 390 at 200 % text, 320).
  - Every new control is ≥ 44 × 44 CSS px (574 observations).
  - Supported solid-text contrast ≥ 4.5:1 (minimum 4.51:1).
  - One course panel is displayed in every state.
  - Zero right overflow after F3.
  - The focused field stays above the sticky bar at 390/200 % and 320 after F2.
  - Keyboard: toggle, D3 focus after move/delete, dialog Escape and focus return, and the summary link across tabs.
  - Leftover fixture courses: 0 (34 created across runs).
- **Final ordered `make check` (Ux18), attempt 1** on `d9f1d25f` (owner consent 2026-11-18, 19:41→19:55):
  format-check, lint, typecheck, test (0 failed / 0 skipped), integration (130 Go packages ok) and check-generated passed;
  `build` FAILED with `no space left on device` while exporting the `asr-worker` image (Whisper large-v3-turbo model
  layer). The Kiri VM disk (63 GB) had 13 GB free at the start and the uncached build cache grew 12.3 GB. Infrastructure,
  not code. Owner decision 2026-11-18: grow the Kiri disk, then rerun. Ux18 BLOCKED until that rerun.

## 3. Non-obvious findings
- **F1 — a defect this phase introduced and the gates missed.** U3's `@container transcript-editor` stalled Chromium
  natively when a populated transcript became visible through the force-mounted language tabs. There was no JS stack:
  `Debugger.pause` returned nothing. It was reproduced on a Ukrainian-labelled tab with one Ready lesson.
  - Unit tests (happy-dom has no size-query layout) and every e2e run stayed green; design-pass found it.
  - Fix: a measured boolean with the same 40rem threshold.
  - Courses 1 decided on 2026-11-18 that the regression proof belongs in Compose e2e, not in a Chromium-in-rstest runner
    (AGENTS verification layers; it would flake under load).
- **F2.** Native focus scrolling does not honour `scroll-padding` for the whole rectangle of a tall textarea. The inset now
  reserves the bar height plus the tallest visible textarea.
- **Review catches that unit tests did not prove:**
  - U1: the CSS is invisible to happy-dom → e2e visibility per tab.
  - U2: label-in-name of the toggle (WCAG 2.5.3).
  - U5: a growing bar covering an already-focused field.
- **Stale or invalid aria-controls ids.** React Aria drops `id` and `role` from inactive tab panels, so a locator built from
  `aria-controls` read up front breaks later (U1 e2e).

## 4. Decisions taken during the phase (dated, author Courses 1 unless noted)
- 2026-11-18 D5a: the lesson-section header covers lesson sections only; course materials keep their h2.
- 2026-11-18: F1 regression proof lives in Compose e2e; the nested browser unit runner was rejected.
- 2026-11-18, owner: e2e-reset + full e2e allowed for U1 and U2.
- 2026-11-18, owner: standing permission to prune the Kiri build cache for this phase. The cutover-snapshot images stay.
  `courses-integration-go` was removed (rebuilt by `make integration`).

## 5. Limitations and debt (named, not claimed away)
- **No Discard (D7).** A reload abandons edits. Mechanism: RHF `reset` regenerates `formKey`s and remounts in-flight editors.
- **Transcript times are still edited as raw ms (D6).** The `m:ss.s` readout is read-only. Mechanism: IndexedDB v1 recovery
  stores raw-ms text without a format discriminator.
- **Video card shows no duration.** `MediaVideo` carries no duration (openapi.yaml ~2381–2437).
- **NOT_TESTED in design-pass:**
  - Contrast of the destructive dialog confirmation on its translucent `/0.14` background (7 observations; no invented
    composited ratio).
  - The empty clean-state bar output (no text).
- **Review minor (non-blocking).** The observer unit batches 100 callbacks in one `act`, which weakens its render-loop bound.
  The real-browser proof is the e2e.
- **R5 (outside this diff).** lesson-builder, refund and teach-editor e2e leave published `e2e-*` courses behind. After
  repeated runs without a reset they push the seeded course off catalog page 1 (catalog.spec:3). Harmless after
  `e2e-reset`.
- **Kiri disk.** 63 GB at 74–99 % used during the phase. The build cache regrows about 17 GB per few builds.

## 6. Next
Ordered `make check` on the final revision (owner consent), then the PR and merge (owner approval).
