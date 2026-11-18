# UX review — admin course editor (`/admin/courses/<id>`), 2026-11-15

Mode: audit only (interface-review = anti-slop-design + ux-review), no code changed.
Trigger: owner, 2026-11-15: "виглядає все всрато".
Revision: `f3-lesson-editor-narrow` @ `0b4f5c1` (F3 CSS applied). Admin theme is dark-only.
Fixture: course `ad637737-…` (1 module, 1 lesson with video, 1 audio track, transcript lines, 0 materials), test admin.
Viewports: 390×844, 768×1024, 1024×768, 1440×900 at 100%. No 200% or 320 pass here: F3 covered them for the same screen.

## Measurements (Lessons tab selected)

| Viewport | Tab list bottom | First active content | Page height |
|---|---|---|---|
| 390 | 461 (top) | 1700 | 5458 |
| 768 | 435 | 1328 | 4740 |
| 1024 | 451 | 1181 | 4608 |
| 1440 | — | 943 | 4229 |

Method: `getBoundingClientRect().top + scrollY` of the only non-`inert` `.ui-tab-panel`, after clicking the second tab.

## Findings (ranked; one defect = one finding)

### B1 — Blocker — all three tab panels render at once; inactive ones are visible but dead
- Lens/rule: ux-review — Nielsen #1 visibility of system status, #4 consistency (tabs must show one panel); WAI-ARIA APG Tabs pattern.
- Evidence: every screenshot `*-lessons-tab-inert-details.jpg`, plus `1440-overview-tab-shows-lessons.jpg`. With Lessons selected, inert "Course details", "Course status", "At a glance" and "Course materials" sit right under the tabs. Lessons content starts 730–1240px lower (table above). Access is stacked, inert, at the bottom.
- Cause: `CourseEditorView.tsx:768/896/928`. `Aria.TabPanel shouldForceMount` keeps one `<form>` across tabs (ddd8c00, PR #39). RAC marks inactive panels `inert`/`data-inert`, but no CSS hides them.
- Source: https://www.w3.org/WAI/ARIA/apg/patterns/tabs/
- Fix: `.ui-tab-panel[inert] { display: none }`. Keep `shouldForceMount` so the form and its values stay mounted. Then verify that the error summary still moves focus to a field on a hidden tab, switching tabs first.

### M1 — Major — one lesson produces a 4–5.5k px page; no overview of the structure
- Lens/rule: ux-review — [law] Miller/chunking, [cog] progressive disclosure; Nielsen #8 aesthetic & minimalist design.
- Evidence: page height in the table above, for ONE lesson. Each lesson expands inline: video, audio tracks, speech recognition and transcript editor. `1440-lessons-audio-asr.jpg`, `1440-lessons-transcript-actions.jpg`.
- Fix: collapsible lesson rows. Collapsed: title, video status, audio/subtitle counts, transcript state. Expanded: one lesson's media.

### M2 — Major — the transcript editor is a raw data form
- Lens/rule: ux-review — forms: match the user's mental model (Nielsen #2); [law] Fitts / density.
- Evidence: `1440-lessons-transcript-actions.jpg`. Each line takes three full-width fields, timings are raw milliseconds, there is an empty "Line N materials" fieldset, and a text "Remove" button per line.
- Fix: a compact row with `mm:ss.s` start/end and text inline, an icon remove button with an accessible name, and the materials picker only when materials exist.

### M3 — Major — saving is split and far away; no dirty state
- Lens/rule: ux-review — Nielsen #1 status, #3 user control; forms: a single clear commit point.
- Evidence: `1440-lessons-bottom-save-access.jpg`. "Save changes" is only at the very bottom, below an inert Access panel. Transcript save, materials, audio upload and ASR apply immediately. Nothing says which edits are pending.
- Fix: a sticky save bar showing "Unsaved changes", with Save and Discard. Label immediate-apply sections as such ("Saved instantly").

### M4 — Major — destructive and reorder actions
- Lens/rule: ux-review — Nielsen #5 error prevention; [copy] concise labels; WCAG 1.4.3 for dimmed text.
- Evidence:
  - Delete lesson (`CourseEditorView.tsx:457`) and delete module (`:543`) remove without confirmation; only the later Save persists it.
  - Reorder buttons are long text ("Move lesson 1 in module 1 up"), repeated per item.
  - At the edges they are disabled at opacity 0.55, with no reason given.
- Fix:
  - Confirm deletes that carry content (video, transcript), or offer Undo.
  - Use icon up/down buttons and keep the long text as the accessible name.
  - Hide or omit the impossible move rather than dimming it.

### M5 — Major — weak hierarchy inside a lesson
- Lens/rule: ux-review — [law] Common Region / Proximity (owns it); anti-slop "flat sections" collapses into corroboration.
- Evidence: `1440-lessons-audio-asr.jpg`. "Speech recognition" and "Lesson transcript" are plain text, not section headings. Label styles vary between sections, and spacing between headings and content is tight.
- Fix: one lesson-section header component (h4 plus a one-line status), applied the same way to video, audio, ASR, transcript and materials.

### M6 — Major — the lesson document editor is unreachable from the course editor
- Lens/rule: ux-review — Nielsen #6 recognition rather than recall; nav: discoverability.
- Evidence: the route `admin.courses.$courseId_.lessons.$lessonId` exists, but no link from the lesson row was found.
- Fix: an "Edit lesson content" link on each lesson row.

### Minor
- m1. The course materials add form is always open, with a disabled Add button and no reason (`1440-lessons-materials-syllabus.jpg`). Fix: an "Add material" button that opens the form.
- m2. Uppercase eyebrow labels are overused; they read as template decoration (anti-slop typography). Fix: keep them for the page header only.
- m3. The admin 404 links to the client catalog (`1440-admin-404.jpg`). Fix: link to `/admin`.
- m4. The video card shows no thumbnail, duration or file name. Fix: show duration and file name (thumbnail optional).

## Proposed direction (staged; changes only after owner agreement)

1. **Quick fix (small PR):** B1, CSS plus a focus check in e2e; m3. This alone removes most of the "broken" impression.
2. **Lesson restructure (one phase, plan → critic → slices):**
   - M1 collapsible rows, M5 section header, M2 compact transcript, M4 icon actions and delete confirmation, M3 sticky save bar;
   - m1 and m4.
3. **Navigation:** M6 link to the lesson document editor.

Gate status: NOT READY. Blocker 1, Major 6, Minor 4 open, none accepted as a limitation.
NOT_TESTED in this audit: light theme (N/A, admin is dark-only); 200% and 320 (covered for this screen by F3, not re-run); contrast of dimmed buttons (opacity compositing, unsupported by the measurement helper).
