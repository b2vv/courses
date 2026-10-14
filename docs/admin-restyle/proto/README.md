# Coursespace — Admin restyle reference

React + React Aria Components + Tiptap + plain global CSS.
Eight views share the supplied admin-shell, ui-card, ui-table and form patterns.

## Run

```bash
npm ci
npm run build
python3 -m http.server 8080 --directory dist
```

Open http://localhost:8080. Navigate with the single labeled sidebar or mobile drawer.
Direct references: #courses, #analytics, #course, #lesson, #learners, #users, #reports, #refunds.

## Scope

This is a reference prototype for translation into the production codebase.
Fixtures are explicitly labeled as examples. Changes last only for the current tab session;
there is no localStorage, backend, authentication, real password reset, payment or streaming service.
The prototype uses the shell class names provided in the brief. The actual production shell DOM
was not supplied; integrate these rules into that shell instead of copying a new chassis.

## Shared primitives

- Shell: .admin-shell + .admin-sidebar + .admin-canvas; no rail. Modal drawer at ≤52rem.
- Tables: .ui-table inside .ui-card; hairline/zebra row tracking, readable identifiers,
  numeric alignment, sticky inline-end action column and sticky inline-start identity column.
  At ≤52rem, each row becomes a labeled card; a visible sort control replaces hidden headings.
- Actions: monochrome inline SVG, ≥44px, aria-label and React Aria visible tooltip.
  Pagination/reorder boundary actions use aria-disabled + no-op, never native disabled.
- Cards: background fill, hairline and shadow; no lighter card fill.
- Forms: labels above controls, validation on blur and submit, nearby errors + summary,
  retained values, confirmation for delete/revoke/refund decisions.
- States: semantic pills, filled pressed toggles, focus rings, unavailable states,
  loading indicator and session-only confirmation toasts.
- Lesson canvas: quiet type bar, document prose, grouped Tiptap toolbar, block palette,
  responsive columns, text/image/video/task sections, student preview.
- Learners: distinct enrollment/status zones, accessible progress track, last activity.
- Safety/payments: structured panels with attached actions; refunds demonstrate both populated
  and instructive empty states.

## Contrast

The original primary gradient is preserved, including on hover. White labels measure
3.25:1 at the lightest stop and 4.50:1 at the darkest stop. The labels are 1.1875rem / 750
(19px at the normal root size) so they use the large-bold AA 3:1 threshold. They must not be
reduced below the large-bold threshold during translation. Solid brand blue never carries
normal-size white text. Gradient endpoints were calculated directly; visual browser checks
of the rendered gradient remain outstanding.

## New semantic tokens

| Name | Value | Purpose |
| --- | --- | --- |
| --space-admin-sidebar | 15.5rem | Existing sidebar reference width |
| --space-admin-canvas | 76rem | Bounded desktop content measure |
| --space-dialog | 36rem | Dialog max measure |
| --space-icon | 1.25rem | Icon geometry inside 44px controls |
| --space-prose-measure | 65ch | Shared editor/reader prose measure |
| --space-prose-min | 12rem | Editable body and textarea minimum |
| --space-tile-min | 9rem | Fluid tile minimum |
| --space-border | 1px | Border geometry |
| --space-focus-ring | 3px | Required focus width |
| --space-focus-offset | 3px | Required focus offset |
| --space-callout-rule | .25rem | Task inline-start rule |
| --color-modal-backdrop | color-mix(in srgb, var(--color-background) 85%, transparent) | Overlay scrim |
| --color-forced-background | Canvas | Forced-color background |
| --color-forced-text | CanvasText | Forced-color ink |
| --color-forced-control | ButtonFace | Forced-color CTA ground |
| --color-forced-control-text | ButtonText | Forced-color CTA ink |
| --color-forced-focus | Highlight | Forced-color focus/selection |
| --color-forced-unavailable | GrayText | Forced-color unavailable ink |

No parallel --size/--type/--weight/--leading/--tracking families. Typography uses the literal
values supplied in the brief. Intrinsic CSS geometry (0, 100%, minmax, aspect ratios), resets
(none/initial), and breakpoint conditions are structural syntax rather than a second visual
scale. Custom properties cannot be used directly in media/container query conditions.
Token values for shadows are reference defaults; production should retain its canonical values.

## Desktop / narrow layouts

Desktop: one left navigation, compact right account header, eyebrow + H1, then pattern-specific
content. Courses/users use tables; analytics uses three clearly labeled stat groups; course
uses grouped details/status panels; lesson uses header/canvas/palette; learners use card rows;
reports/refunds use decision panels.

At ≤52rem: navigation becomes a modal drawer opened from a compact header; tables become labeled
cards with local actions; forms and queues collapse fluidly. Lesson columns collapse at 40rem
container width (3 → 2 below 60rem). Controls wrap instead of scrolling the document horizontally.

## Validation

```bash
node check.mjs
node interactions.mjs
```

Checks cover SSR of all eight views, prohibited token/control patterns, gradient endpoint
contrast, search/pagination, creation and retained validation values, confirmed user actions,
report filtering, refund empty/populated states, and focus-preserving lesson reorder boundaries.
Interaction tests use jsdom and do not prove browser layout correctness.
Actual 375px, 200% text zoom, tooltip positioning and forced-colors browser QA is still required.

## Files

src.jsx: shared primitives, screens and inline SVG set.
fixtures.js: explicitly named demonstration data.
dist/styles.css: reusable CSS rules and token proposals.
dist/fonts/: local Inter Variable and Space Grotesk Variable assets with licenses.
entry.jsx / build.mjs: app mount and esbuild bundling.
