# admin-editor-ux — plan (r3, 2026-11-18)

r2 answers plan critique r1 (BLOCKED, 3 blocking / 3 non-blocking): `_wip-plan-critique.md`. Delta in §7.

Owner: Courses 1. Trigger: owner 2026-11-15 ("виглядає все всрато") → audit
`docs/ux-review/admin-course-editor-2026-11-15/report.md` (B1, M1–M6, m1–m4); fix deferred by the owner
until MS2–MS4 were done; started by the owner 2026-11-18 ("Берися за UX-фікс адмін-редактора").
Branch `admin-editor-ux` from `main` @ `10436cdf`. Outside §24 (owner-initiated, like admin-restyle).

## 0. Re-verified on `main` @ 10436cdf (2026-11-18, built image of the same tree, 1024×768, course `a27c419b`)
- B1 holds: with Lessons selected the inert Overview panel is `display:block` at y=483 (h 674); the Lessons
  panel starts at y=1181; inert Access panel at y=4225. Page 4755 px for one lesson. Method:
  `getBoundingClientRect().top+scrollY` of each `.ui-tab-panel` after clicking tab 2.
- The `<form>` and the price block already use `hidden` and are `display:none` on Access (measured) — only
  the three force-mounted `Aria.TabPanel`s leak. `.lesson-language-panel[inert]{display:none}` (MS3) already
  hides inactive language tabs inside a lesson, so the fix pattern is established.
- M2 holds (inputs `Line N start, ms` / `end, ms` text, textarea, text `Remove line N`).
- M4 holds (text buttons `Move lesson 1 in module 1 up` disabled at edges; `Delete lesson/module` remove with
  no confirmation, `CourseEditorView.tsx:409/498`).
- M6 holds: 0 anchors containing `/lessons/` in the editor.
- m1 holds (`Add a material` disabled, form always open). m3: not re-measured, `__root.tsx:57` NotFound is the
  single not-found component for client and admin.
- MS3 added per-language lesson tabs (`LessonLanguageTabs`): the audit's "audio / ASR / transcript" now
  live inside one language tab — M1/M5 apply to that structure, not to the 2026-11-15 one.

## 1. Scope and non-goals
In: B1, M1–M6, m1–m3; m4 only as far as the existing `MediaVideo` contract carries the data
(`originalName` is present; duration — the implementer checks the schema; no backend change in this phase).
Out: backend/API changes, the `/teach` editor (separate product boundary — it reuses only the lesson document
editor), the lesson document editor itself, light theme (admin is dark-only), the client side.

## 2. Decisions (dated, author Courses 1, 2026-11-18; owner may overrule before the slice starts)
- D1 B1 fix = CSS `[inert]{display:none}` on the course-editor tab panels, keeping `shouldForceMount` (one
  `<form>`, values survive tab switches — ddd8c00/PR #39). A dedicated class `course-editor-panel` on ALL
  THREE course panels (Overview and Lessons sit inside `<form>`, Access is a direct child of Tabs, so a
  child combinator misses two of them — critique #5) and `.course-editor-panel[inert]{display:none}`; not the
  global primitive (overlays.tsx tabs do not force-mount; gotcha 2026-10-14 "admin specifics leak into global
  primitives"); language panels keep their own rule. Check: class + `inert` per selection (unit); visibility of the selected and
  the two inactive panels for each of the three selections (e2e — happy-dom does not apply main.css) + DP; `focusField` already flushSyncs the tab before `setFocus` — a NEW unit and
  e2e case for an invalid field on the opposite tab (the existing admin.spec validation case covers Overview
  only). Visibility of the UI path in e2e is asserted explicitly (§4 U1/U2 migration), not inferred from a
  green full run: `setInputFiles` has no visibility guard (Playwright coreBundle.js:20133–20168).
- D2 Delete of a lesson/module = confirmation dialog only when the item carries server state (a saved lesson
  id; for a module, any saved lesson). Unsaved new rows remove at once. Rationale: removal is persisted only on
  Save, but a saved lesson with video/transcript is lost on Save without a second look; Undo would need a
  form-history mechanism the editor lacks. Mechanism to check: the dialog appears for a saved row and not for a
  fresh one (unit).
- D3 Reorder = icon buttons (up/down) with the existing long text as the accessible name; the impossible
  direction is not rendered (no disabled dimmed control). Keyboard order and ≥44 px target kept. Focus
  (critique #6): after a move, focus stays on the pressed control in its new position if it still exists,
  otherwise on the opposite move control of the same row; after a confirmed delete, focus goes to the next
  row's title field, else the previous row's, else the "Add lesson/module" button.
- D4 Lesson rows collapse (M1). The row header is ALWAYS visible and keeps the lesson's form fields — title
  input and the "open before purchase" checkbox (they are the course-form fields; error-summary links and
  purchase.spec keep reaching them without expanding) — plus a video-state badge, a transcript yes/no badge
  (from the existing `onTranscriptKnown` callback, no new query; three states — unknown until the callback
  fires (no badge), yes, no; a known value is kept on a failed refresh; no track/language count — critique #4 showed
  "tracks" and "language tabs" differ, and neither is needed to find a lesson), the "Edit lesson content" link
  (D8) and an expand toggle (`aria-expanded`, `aria-controls`). Collapsible body = video, language tabs,
  reorder/delete actions. Body stays MOUNTED when collapsed (`hidden`, not unmounted): `LessonVideoField`
  owns upload progress state (LessonVideoField.tsx:149–209). Default: collapsed when the course has >1 lesson;
  a single lesson and a newly added lesson start expanded. Expanded state is per lesson id, survives tab
  switches, resets on reload.
- D5 One lesson-section header component (M5): `h4` + one-line status, used by video, language tabs
  (audio/ASR/transcript inside), materials. Headings keep a valid outline (h2 program → h3 lesson → h4 section).
- D6 Compact transcript line (M2): one row per line — start/end inputs, text inline, icon remove with
  accessible name `Remove line N`; line materials picker only when the course has ≥1 material. The inputs
  STAY raw milliseconds (stored, restored and edited exactly as today); next to each input a read-only
  `mm:ss.s` readout formatted by a pure function (`formatTranscriptTime`) in the transcript application layer,
  shown only when the field parses. Decision 2026-11-18, Courses 1, closes "edit in mm:ss.s": mechanism —
  local recovery keeps raw-ms text including half-typed input (recovery.ts:14–27, TranscriptEditor.test.tsx
  :250–260) in IndexedDB v1 with no format discriminator (transcript-history-indexed-database.ts:5–25); a
  format switch needs a versioned reader and a display buffer (critique #1), which costs more than the
  density gain it buys. Check (unit): a restored raw-ms snapshot keeps its time TEXT unchanged and submits the
  same integer ms values, including sub-100 ms precision; half-typed text stays recoverable. Generated
  segment ids / JSON bytes are not compared (TranscriptEditor.tsx:303–311 already mints new ids).
- D7 Sticky save bar (M3): inside the form, `position: sticky; bottom: 0`, shows "Unsaved changes" when
  `isDirty`, and the existing Save button. NO Discard — decision 2026-11-18, Courses 1, mechanism (critique
  #2): `form.reset` regenerates RHF `formKey`s (react-hook-form index.esm.mjs:1188–1194, 3356), which remounts
  every lesson subtree keyed by them (CourseEditorView.tsx:334–347, 449–462) and drops in-flight upload
  progress and unsaved transcript lines; reload remains the way to abandon edits. Immediate-apply areas
  (transcript save, materials, audio upload, ASR, price, access) get a short "Saved immediately" hint in their
  section header (D5 status line). Bar is hidden on Access (form hidden). `scroll-padding-bottom` = bar
  height so a focused field is never under it (WCAG 2.4.11).
- D8 M6: "Edit lesson content" link on each SAVED lesson row to
  `/admin/courses/$courseId/lessons/$lessonId` (route exists). Unsaved lessons show no link.
- D9 m1: materials form behind "Add a material" toggle; m2: eyebrow only in the page header; m3: admin 404
  links to `/admin` when the path is under `/admin` (root NotFound reads the location).

## 3. Rows (each proven by the named evidence; FE = rstest unit/route, E2E = Playwright on Compose, DP =
design-pass browser evidence)
| Row | Behaviour | Proof |
|---|---|---|
| Ux01 | Only the selected course-editor tab panel is displayed; inactive panels `display:none` | FE (class + inert) + E2E (visibility of all three panels per selection — happy-dom does not apply main.css, U1 review 2026-11-18) + DP 4 viewports |
| Ux02 | Error-summary link to a field on the hidden tab (Overview↔Lessons, both directions) switches and focuses it | FE + E2E (new case) |
| Ux03 | Admin 404 links back to `/admin` | FE |
| Ux04 | Lesson rows collapse/expand; header keeps title + preview inputs, video-state and transcript badges | FE + DP |
| Ux05 | Collapsed content stays mounted: form values and an in-flight upload survive collapse | FE |
| Ux06 | Lesson sections share one header (h4 + status); heading outline valid | FE (roles) + DP |
| Ux07 | Transcript line compact row; `mm:ss.s` readout from `formatTranscriptTime` (pure), absent for unparsable input | FE (pure + UI) |
| Ux08 | Raw-ms unchanged: restored time text identical, submitted integer ms identical (sub-100 ms too); half-typed input recovered | FE |
| Ux09 | Reorder icon buttons with full accessible name; impossible direction absent; focus per D3 | FE (activeElement) |
| Ux10 | Delete of a saved lesson/module asks for confirmation; unsaved removes at once; cancel keeps it; focus per D3 | FE + E2E |
| Ux11 | Sticky save bar: dirty indicator + Save; focused last field not obscured at 200 %/320 | FE + DP |
| Ux12 | Immediate-apply sections say so | FE |
| Ux13 | "Edit lesson content" link on saved lessons only, correct href | FE + E2E |
| Ux14 | Materials add form behind a toggle | FE |
| Ux15 | Eyebrows only in the page header | DP |
| Ux16 | Video card shows `originalName` (MediaVideo carries filename, no duration — openapi.yaml:2381–2437) | FE |
| Ux17 | Design-pass: 390/768/1024/1440 at 100 %, 1440+390 at 200 %, 320; keyboard; contrast of new text; no new overflow | DP |
| Ux18 | Full ordered `make check` green with the full e2e set | gate |

## 4. Slices (each: Codex write run → Codex independent review → fixes → owner session browser check)
- U1 — B1 + m3 (Ux01–03). e2e migration: discussion.spec.ts:79–101 uploads from the default Overview tab →
  add a real Lessons tab click and a visible Upload-button assertion before `setInputFiles`; in every
  upload helper that polls the API only (transcript.spec :286–310, enrollment.spec :186–225, playback.spec
  :276–325) add, before `setInputFiles`, an assertion that the target lesson's active video control is
  visible — "Завантажити відео" OR "Замінити відео" (LessonVideoField.tsx:299–301; B40 replaces a ready
  video, playback.spec :1561–1576 — keep its fresh-asset-id polling). New admin.spec case: invalid
  lesson title while Overview is shown → summary link → Lessons tab + focus.
- U2 — collapsible lesson rows + section header (Ux04–06) + M6 link (Ux13). e2e migration (critique R1
  inventory, `_wip-plan-critique.md`): one IDEMPOTENT helper `openLesson(editor, module, lesson)` — reads the
  scoped toggle's `aria-expanded`, clicks ONLY when it is `false`, then always asserts `aria-expanded=true`
  and the body visible (a single lesson starts expanded — playback.spec :695–718; audio-tracks.spec calls it
  repeatedly on one row); the caller keeps selecting the Lessons tab itself; called before body interactions in
  media.spec (:147–189 + after reload :404–408), audio-tracks.spec (:89–128, both lessons), subtitles.spec
  (:591–625, each loop target), playback.spec (:276–325, each target). purchase.spec and admin.spec need no
  expansion (fields stay in the header — D4); single-lesson specs start expanded. No `force`, no
  `includeHidden`, no API substitute for a UI step.
- U3 — compact transcript editor (Ux07–08) in the transcript module.
- U4 — reorder/delete actions (Ux09–10) + materials toggle + eyebrows + video name (Ux14–16).
- U5 — sticky save bar (Ux11–12).
- U6 — design-pass evidence (Ux17) + report + ordered `make check` (Ux18) + PR.

## 5. Risks
- R1 Existing e2e that touch elements in an inactive panel or a collapsed lesson start failing (Playwright
  refuses hidden elements). Mitigation: grep `frontend/e2e` for editor selectors before U1/U2; every slice
  that changes visibility runs the full e2e set, not targeted (AGENTS lessons).
- R2 Collapsing unmounts upload state → lost upload. Mitigated by D4 (hidden, not unmounted) + Ux05.
- R3 Transcript timing loss. Mitigated by D6 (display-only readout, raw-ms storage and editing unchanged) + Ux08.
- R4 Sticky bar overlaps focused fields at 200 % / 320 (WCAG 2.4.11 Focus Not Obscured). DP must check
  focus visibility of the last field under the bar; `scroll-padding-bottom` on the scroller.

## 6. Verification
Per slice: `cd frontend && npm run format && npm run format:check && npm run typecheck && npm run lint &&
npx rstest run src`; owner session rebuilds `frontend` on Kiri and runs a browser look. U1 and U2 change
visibility: each runs the FULL e2e set (Kiri, `courses-kiri` wrapper) before review closes; other slices run
the affected specs, and the final ordered `make check` runs everything.
Final: ordered `make check` (owner consent per run), design-pass, report `docs/admin-editor-ux/report.md`.

## 7. Delta r1 → r2 (critique r1 findings)
| # | Finding | Disposition |
|---|---|---|
| 1 | mm:ss.s editing breaks raw-ms recovery | ACCEPTED → D6 keeps raw ms, adds a read-only readout; Ux07/08 rewritten |
| 2 | Discard via reset remounts in-flight editors | ACCEPTED → D7 drops Discard (dated decision, mechanism); Ux11 rewritten |
| 3 | Full e2e does not prove visibility | ACCEPTED → U1/U2 migration lists, explicit visible assertions, full e2e per visibility slice (§4, §6) |
| 4 | Track vs language-tab count | ACCEPTED → no count; video + transcript badges from existing callbacks (D4) |
| 5 | Child-combinator selector misses two panels | ACCEPTED → dedicated `course-editor-panel` class (D1) |
| 6 | Focus after reorder/delete | ACCEPTED → D3 focus rules; Ux09/10 assert activeElement |

## 8. Delta r2 → r3 (critique r2 findings)
| # | Finding | Disposition |
|---|---|---|
| R2-1 | `openLesson` must be idempotent | ACCEPTED → U2: read `aria-expanded`, click only when false, always assert expanded + visible |
| R2-2 | Unknown transcript knowledge ≠ "No" | ACCEPTED → D4 three states |
| R2-3 | "byte-identical" overclaims | ACCEPTED → D6/Ux08: time text + integer ms equality only; R3 aligned |
| R2-4 | Upload OR Replace control | ACCEPTED → U1 assertion names both; B40 kept |

## 9. Checkpoint
- stage: plan — ACCEPTED r3 (sha256 039fb8c3c169b98df22df14b28fefbc2ed834fb69213618d397ea9ed4894a495, numbering
  the file before this §9 block), 2026-11-18. Critic: Codex session `01a1ee6a-fa74-7e31-b85f-b61d4664701e`,
  four lenses, prompt `ai-factory-kit/prompts/plan-critic.md` v1 sha256 514e4c37…dcab6. History: r1 BLOCKED
  (3/3), r2 BLOCKED (1/3), r3 ACCEPTED (0 open). Raw: `_wip-plan-critique.md`. Target: `admin-editor-ux` @
  10436cdf.
- review_mode=independent_subagent (Codex read-only review per slice); writer Courses 1 via Codex write run.
- Next: U1 READY (no open question); U2–U5 READY by the same verdict; U6 needs owner consent for `make check`.
