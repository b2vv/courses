# Admin Restyle — Brief (whole admin, shared primitives)

Design brief for an AI design tool. Produce a **coherent redesign of the whole admin
surface of a courses platform**, built on **shared primitives** (shell → data-table →
form/card), not 8 disconnected screen redesigns. The screens then inherit the primitives.

Paste this whole file into the design tool. Tokens, class names and the pain list are
real — copied from the codebase and a live review.

---

## 0. How to use this brief

- The output is a **reference prototype** we will **translate into our code** on **our
  tokens** — so use our CSS custom properties (`var(--…)`) and our class names, never raw
  hex/px or a new framework.
- Fix the **shared primitives once**; show each of the 8 screens using them.
- The admin surface is **dark-only**, `data-surface="admin"`. Brand is **blue** on admin
  (`#5273f2`), with the platform's **lime** accent available for highlights. Do **not**
  invent a new indigo/palette.

## 1. Hard constraints (do not violate)

1. **Stack:** React + **React Aria Components** + plain CSS with custom properties.
   **No Tailwind, Radix, shadcn, CSS-in-JS, Bootstrap.** Global CSS keyed by semantic classes.
2. **Semantic tokens only** — every color/space/radius/shadow is a `var(--token)`. If a value
   has no token, name a new semantic token; never inline a literal. **Do NOT ship a parallel
   token layer** (no `--type-*/--size-*/--weight-*/--leading-*/--tracking-*` families, no
   `--size-zero`, no `--color-transparent`) — our system names only `--color-*/--space-*/
   --radius-*/--shadow-*/--control-min-size/--card-*`.
3. **Use OUR values, not a remix:** font weights **650/700/750** (not 500/600/700); primary
   button uses our `--btn-primary-bg/-highlight/-glow` gloss (not a `color-mix` CTA).
4. **Accessibility (CI-gated):** controls ≥44px (`--control-min-size`); visible focus ring
   (`outline:3px solid var(--color-focus); outline-offset:3px`); **no horizontal scroll at
   200% text zoom / reflow**; `prefers-reduced-motion`; `forced-colors`; WCAG AA contrast
   (text 4.5:1, non-text 3:1). **Never** put `isDisabled`/native `disabled` on a focused
   nav/reorder control (drops focus to `<body>`) — keep enabled + no-op instead.
   **Contrast landmine:** solid `#5273f2` (brand blue) = 4.12:1 on white → it does NOT carry
   normal-size text (use it only as an accent/border, or deepen the fill). White text on the
   `--btn-primary-bg` gradient must be large/bold (the 3:1 path) and **verified by hand** — our
   automated check (axe) measures flat fills but **skips gradients**, so a sub-AA gradient would
   pass CI silently.
   **Icon controls need names:** any icon-only control carries an `aria-label` + a visible tooltip;
   never a bare glyph with no accessible name.
5. **Logical CSS props** throughout (`inline-size`, `block-size`, `margin-block`,
   `padding-inline`, `border-inline-start`). Bilingual (uk/en), RTL-ready.
6. **Do NOT ship your own app shell** (sidebar/topbar/breadcrumbs/brand) as production — we
   already have one; redesign OURS (below), don't replace it with a prototype chassis.
7. **Icons:** no icon set today; propose inline SVG, monochrome `currentColor`, inside ≥44px controls.

## 2. Design tokens — admin surface (effective values, paste-ready)

```
/* grounds & ink (cards are NOT lighter than the page — set off by border+shadow) */
--color-background:#0c0d12  --color-surface:#111317  --color-elevated:#17181d
--color-text:#e6e8ec  --color-muted:#8b95a3  --color-faint:#828b99
--color-border:#656e7b (control-grade, ≥3:1)  --color-hairline:#20232b (decorative)
/* brand (blue) + focus + lime accent */
--color-primary:#5273f2  --color-primary-hover:#6986ff  --color-on-primary:#fff  --color-focus:#7b93ff
--color-accent:#b9ff4a (lime, --color-on-accent:#18200d)  /* platform accent, use sparingly */
--btn-primary-bg:linear-gradient(180deg,#6986ff,#4c6bf0)  + --btn-primary-highlight (inset top) + --btn-primary-glow
/* signal */ --color-success:#4ed6a8 --color-warning:#fcb646 --color-danger:#e5695f --color-info:#6986ff
/* type */ --font-body:'Inter Variable' --font-display:'Space Grotesk Variable' --font-utility:ui-monospace
   prose 1.0625rem/1.65, micro-label --font-utility .6875rem 700 ls .08em uppercase --color-faint, weights 650/700/750
/* space */ --space-1..8: .25/.5/.75/1/1.25/1.5/2rem   /* radii */ control≈.625 panel/card .75 pill 999
/* shadow */ --shadow-card, --shadow-floating   --control-min-size:2.75rem  --card-padding:var(--space-4) (admin dense)
```
Primitives already in the system to REUSE (don't reinvent): `.ui-button(--primary/secondary/danger)`,
`.ui-card(--floating)`, `.ui-field/.ui-input/.ui-textarea`, `.ui-pill(--neutral/success/danger/info)`,
`.ui-table/.ui-table-scroll/.ui-pagination`, `.ui-segmented/.ui-segment`, `.ui-stat/.ui-chart`,
`.ui-icon-tile`, `.ui-state/.ui-skeleton`, `.ui-modal/.ui-dialog`, `.ui-toast`, `.ui-meter`, `.ui-progress`.

## 3. Shared primitives to redesign (fix once → screens inherit)

### 3a. Shell (`.admin-shell` = rail + sidebar + canvas)
- **Collapse the double nav to ONE.** Today a 3.5rem icon-rail (`.admin-rail`, decorative,
  `aria-hidden`) sits beside a 15.5rem labeled sidebar that mirrors it → reads as two navs.
  Keep one clear nav (the labeled sidebar); drop/absorb the rail.
- **Narrow (≤52rem): collapse to a drawer / top-bar that does NOT eat content height.** Today
  the sidebar becomes a full-width block ABOVE content, pushing it down (must scroll the whole
  nav to reach content). Fix with a toggle drawer or compact top-bar.
- Keep the good parts: the eyebrow micro-label (PLATFORM STATISTICS / ACCOUNT ADMINISTRATION /
  SAFETY / PAYMENTS) + large H1, and the right header (Administration · user · UA/EN · Sign out).

### 3b. Data-table (shared pattern — the biggest offender)
- **Elevation + row tracking:** the table/its container reads as a `.ui-card`; zebra or hairline
  rows + hover; human-readable first column (not a UUID); right-align numbers (tabular/mono).
- **Sticky row actions (owner requirement):** the inline-end **action column is sticky**
  (`position:sticky; inset-inline-end:0`) with a surface background + hairline, so actions stay
  visible when the table scrolls horizontally. Today actions are a fixed 11.5rem column that
  scrolls out of view.
- **Compact actions:** replace 3 stacked full buttons per row (Block/Delete/Reset) with compact
  icon buttons or an overflow menu → shorter rows. **Each compact icon action carries an
  `aria-label` + visible tooltip** — never icon-only without an accessible name (consistent
  identification across rows/screens).
- **Narrow strategy:** horizontal scroll WITH the sticky identifier + action columns, OR stacked
  cards — never mid-word wrap (today a UUID stacks char-by-char at ~500px).
- **Filters/sort:** visible active-filter / sort state; clear empty/zero state (why empty, + action).

### 3c. Form / card
- **Elevation everywhere:** admin containers are flat (thin border, no shadow) because `.ui-card`
  (which has border+shadow) is unused on admin → tiles/sections merge into one grey field. Unify
  admin cards/panels/tiles onto the `.ui-card` treatment (background + border + shadow).
- **Status = `.ui-pill` badge** (success/danger/info/neutral), not bespoke `.admin-*-status` text.
- **Form patterns:** label above field; placeholder ≠ label; inline validation on blur (not per
  keystroke); error text next to field + error summary; preserve input on error; destructive
  (delete/revoke/refund) needs confirm.
- **Interaction states:** every control defines rest / hover / active / focus / disabled / loading,
  visually distinct (today: underline-hack pressed states, no loading feedback, flat disabled).

## 4. The 8 screens (pattern → what to fix)

1. **admin.index** — courses TABLE + create dialog. Apply 3b. (Owner: this + users are the table test.)
2. **admin.analytics** — DASHBOARD of stat-tile grids (Platform / People / Content). Fix: tiles use
   `.ui-card`/`.ui-stat` elevation (stop merging); group headers separate the 3 groups; fix the odd
   letter-spacing on money ("UAH 7,560"); `.ui-chart` trend; `.ui-segmented` range.
3. **admin.courses.$courseId** — course EDITOR (form-tree: details/price/transition/delete, video,
   drafts, grants). Apply 3c (card grouping, form patterns, destructive confirm).
4. **lesson editor** — already decided: adopt the lesson-studio visual language translated to our
   tokens (see our editor-redesign decision). Make the admin restyle visually consistent with it.
5. **admin.learners** — filter card-LIST (name, last activity, enrollment link + status, % progress).
   Fix: status "Active" butts against the link (missing gap); rows merge (no cards/dividers) → card
   rows; hairline progress → real `.ui-progress`/`.ui-meter` track.
6. **admin.users.index** — TABLE (Email/Role/Status/Updated/Actions), role badge, status pill,
   per-row Block/Delete/Reset. Cleanest today → the REFERENCE; just apply compact+sticky actions (3b).
7. **admin.participant-reports** — safety QUEUE (currently a bare "Label: value" text dump per report,
   actions flung far right, "Open" as plain text). Fix: each report = a `.ui-card` panel with
   structure, actions attached, status as a `.ui-pill` badge.
8. **admin.refunds** — payments QUEUE (empty now). Fix: structured empty state (icon + framing), and
   design the populated queue (rows + resolve/revoke + status).

## 5. What to AVOID (prototype pitfalls we hit on the editor)

Parallel token families; `--size-zero`/`--color-transparent`; `color-mix` CTA remix; `isDisabled`
on nav/reorder buttons; your own app shell; localStorage/mock plumbing presented as the real thing;
font weights 500/600/700. All of these get rejected in translation.

## 6. What I need back

1. **Shared-primitive specs** (CSS on our tokens/classes): shell, data-table, form/card — with the
   sticky-action and narrow-table behavior shown.
2. **Each of the 8 screens** rendered with those primitives (desktop + narrow ~375px).
3. **Inline-SVG icon set** (monochrome `currentColor`): nav items, row actions (edit/block/delete/
   reset/resolve/revoke), status, section types. Each icon used as a control gets an `aria-label`
   + tooltip — never icon-only without an accessible name.
4. **New tokens flagged** (semantic name + value) wherever a value is missing.
5. Short **rationale** mapping each change to the pain list in §3–§4.

## 7. Acceptance (how we'll judge it)

The prototype is judged against our **ux-review** rules (Nielsen heuristics, Laws of UX, WCAG 2.2
UX criteria, forms/tables/states/nav/copy/mobile/cognitive) — every fix must survive an observable
check — plus an intake checklist: one nav, card elevation, sticky row-actions + narrow-table
strategy, participant-reports as a panel, statuses as pills, interaction states, instructive empty
states, 200%-zoom no h-scroll, learners-list spacing (no status butting the link), **accessible
names on all icon controls**, **brand-blue contrast** (no normal-size text on solid `#5273f2`;
gradient text large/bold and hand-verified because axe skips gradients), and **no parallel tokens /
isDisabled / own shell**.
