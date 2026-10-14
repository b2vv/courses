# Admin restyle — звіт фази

Переклад прототипу плагіна (`docs/admin-restyle/proto/`) на наш код і токени. План —
`plan.md`, тикети-зрізи — `tickets.md`. Порядок: S0 → S1 → … → S6.

---

## S0 — токени + shell

**Статус:** реалізовано, повний гейт зелений, browser-verify на реальному коді зроблено;
**на рев'ю Courses 2** перед комітом у main.

### Зроблено (6 файлів, усі у скоупі)
- `tokens.css`: +6 контейнерних токенів (`--space-admin-sidebar/-canvas/-dialog/-prose-measure/
  -prose-min/-tile-min`), суто additive — рівно затверджений планом набір (рішення 2). Sidebar/canvas/
  dialog застосовано в shell; prose/tile оголошено для наступних зрізів.
- `tokens.admin.css` — **рішення 1 (deepen GRADIENT, не flat):** `--btn-primary-bg` і
  `--btn-primary-bg-hover` виведено з `--color-primary` через `color-mix(…+#000)` (rest 92→86%,
  hover 94→88%), глянець D3 (highlight+glow) збережено, **без паралельних хардкоджених hex**. Той
  самий токен живить `.ui-segment[data-selected]`. +`--color-modal-backdrop`, tokenized admin dialog
  width, `@media (forced-colors: active)` для primary/segment.
- `main.css`: прибрано декоративний `.admin-rail`; admin-scoped двоколонковий shell на токенах;
  nav ≥44px; mobile drawer ≤52rem із toggle; логічні padding/inset; 3px focus; prefers-contrast/
  forced-colors.
- `AdminShell.tsx`: один reusable navigation-фрагмент (desktop sidebar АБО RAC-drawer); topbar
  icon-toggle (`aria-label`+`aria-expanded`+tooltip); `DialogTrigger/ModalOverlay/Modal/Dialog`
  (focus-containment/Escape/dismiss/restore); закриття після nav-click і при переході ≥52rem.
  Header (title · user · UA/EN · Sign out) збережено.
- `AdminNavIcon.tsx`: +`menu`-гліф для toggle.
- `__root.tsx`: `data-surface="admin"` на `<body>` (lifecycle з restore-cleanup), щоб RAC-портали
  (drawer/tooltip/dialog) успадкували admin-токени.

### Дефект, зловлений до гейту
Codex лишив `autoFocus` на close-кнопці drawer → oxlint `jsx-a11y(no-autofocus)` + `--deny-warnings`
= `make lint` FAIL (exit 1). Прибрано — RAC `Dialog` сам веде фокус у діалог (підтверджено нижче).
`$?` через `| tail` маскував це як 0 — читати реальний exit без пайпа ([[background-exit-code-is-last-pipe]]).

### Верифікація
- **Ordered `make check` (повний e2e)** — зелений локально: `GATE_EXIT=0`, **e2e 118 passed (15.2m)**,
  go-тести ok, 0 make-fail. Метод: `make -C <repo> check`, лог `codex-s0-gate.log`. CI на білінгу
  (job verify не стартує) — гейт лише локальний.
- **Рішення 1 — градієнт виміряно в браузері** на РЕЗОЛЬВНУТИХ стопах (браузер розкрив color-mix у
  `color(srgb …)`; probe-елемент під `[data-surface=admin]`, WCAG-люмінанс vs білий):
  rest **4.754 / 5.307**, hover **4.585 / 5.114**; min **4.585 ≥ 4.5** — усі стопи проходять normal-
  text AA, текст не роздуто (`.8125rem/650`), глянець збережено (дві різні стопи). Збігається з
  розрахунком Codex. Language-switcher pressed-fill = `rgb(71,99,208)` = `#4763d0` (5.30:1) ✓.
- **Shell desktop** (1280): один nav (рейл `.admin-rail` count = 0), `body[data-surface]=admin`,
  h-scroll відсутній, eyebrow+H1+правий хедер на місці.
- **Narrow 390px** (CDP-емуляція, `innerWidth=390`, mqNarrow=true): sidebar `display:none`, toggle
  видимий **46×44px**, лише один видимий nav-landmark, h-scroll відсутній. Drawer відкривається
  (`aria-expanded=true`, `role=dialog` «Admin», 6 лінків + To client), **фокус переходить у діалог**,
  h-scroll відсутній; закривається (RAC restore фокуса на toggle).
- **200% zoom:** в'юпорт 640 (=1280@200%) та text-zoom root-font 200% — h-scroll відсутній в обох.

### Обмеження / межі зрізу
- **Forced-colors — CSS реалізовано й оглянуто в дифі** (`@media (forced-colors: active)` у
  tokens.admin.css для primary/segment; prefers-contrast/forced-colors у main.css), але **live-
  емуляцію forced-colors наявний браузер-тулінг не дає** — не claim'иться як live-верифіковане;
  підтвердити OS-перемикачем high-contrast (Courses 2 / власник).
- **Таблиця на вузькому кришиться по символах — це S1** (stacked-cards <52rem), не S0; contained
  (page h-scroll відсутній). Shell-частина S0 на вузькому коректна.
- prose-measure/prose-min/tile-min токени оголошено, але споживачів у S0 нема (наступні зрізи).

### Рев'ю Courses 2 (intake + anti-slop + ux-review): 1 major, 2 minor, 0 blocker
Незалежно підтверджено: 4 стопи градієнта (rest 4.74/5.30, hover 4.60/5.13, min 4.60 ≥ 4.5 — збіг
із probe), nav-контраст (muted 6.1:1, faint 5.4:1), відсутність FOUC (`data-surface=admin` стоїть на
`.app-shell` у рендері `__root.tsx`, не лише в ефекті), skip-link → `<main id="main">`, один nav-landmark.
- **MAJOR (виправлено):** `--space-icon:1.25rem` — поза затвердженим планом набором (рішення 2 називає
  рівно 6), мисномер (геометрія гліфа ≠ spacing), і `inline-size:var(--space-icon)` тихо перебивав svg
  `width/height="16"` → nav-гліфи виросли 16→20px. Прибрано токен + usage → гліфи повернулись до
  intrinsic 16px (revert ненавмисної зміни, не нове рішення).
- **MINOR 1 (виправлено):** drawer `Dialog aria-label` розходився з toggle → приведено до
  `common.mainNavigation` (= лейбл toggle).
- **MINOR 2 (прийнято як є, розсуд):** на відкритті drawer фокус на весь `Dialog` (3px рамка навколо
  панелі) — доступно; RAC-default, без autoFocus (який валив lint). Фокус на перший лінк — опційний
  майбутній tweak, у S0 не роблю (мінімальність зрізу).

**Повторний gate після фіксів (gate2):** ordered `make check` зелений — `GATE_EXIT=0`, **e2e 118
passed (17.1m)**, 0 fail (лог `codex-s0-gate2.log`). Фікси верифіковано: `--space-icon` grep порожній,
гліфи 16px; drawer aria-label = toggle. Courses 2 звірила дерево — 0 blocker, approve.

### Наступне
S0 змержено в main (PR #23 squash `ca592c4`). Далі S1.

---

## S1 — DataTable (спільний примітив)

**Статус:** реалізовано, повний гейт зелений, browser-verify зроблено; **на рев'ю Courses 2** перед комітом.

### Рекон-корекція скоупу (Explore, `scratchpad/s1-recon.md`)
Тикет злегка переобтяжив: **courses-таблиця read-only (без дій)** — compact-icon/sticky стосуються лише
Users; **жодної числової колонки** в consumer'ах S1 → `align` відкладено на S4 (Analytics). E2e прибиває
name-bearing aria-labels Users-дій + `role=row` → compact мусить зберегти РІВНО їх.

### Зроблено (7 файлів)
- `ui/data.tsx` (спільний Table): opt-in `TableColumn.pin?:'end'` / `rowHeader?:boolean` (default-вигляд
  незмінний → Analytics не регресить); **явні ролі** `table/rowgroup/row/cell/columnheader/rowheader` +
  `data-label`/`data-pin`. caption/emptyMessage/tabIndex/Pagination збережено.
- `app/styles/main.css` + `tokens.css`: elevation (`--shadow-card`) на `.admin-list-results`; sticky
  `[data-pin='end']` (непрозорий `--color-surface` + hairline); `th[scope='row']` reset (ідентифікатор не
  uppercase); **narrow ≤52rem stacked-cards** (`display:block` + `[data-label]::before`), pin→static,
  compact actions — усе **скоупнуто до `.admin-list-results`** → Analytics (S4) і report/refund-черги (S6)
  не зачеплені.
- `admin/UserListView.tsx`: account→rowHeader, actions→pin='end'; стековані текст-кнопки → compact
  icon-`Button` у `Tooltip` (inline-SVG 16px), **aria-labels дослівні** (`adminUsers.block/restore/delete/
  resetPassword` + name), tooltip=`*Short`; danger/handlers/`noActions` збережено.
- `admin/CourseListView.tsx`: лише title→rowHeader.
- `ui/overlays.tsx` + `admin/ResetPasswordDialog.tsx` (поза явним скоупом, обґрунтовано): opt-in
  `trigger?:ReactElement` до спільного `Dialog` (default-гілка незмінна → інші споживачі безпечні), reset
  передає icon-trigger. Conditional spread під `exactOptionalPropertyTypes`.

### Дефект, зловлений до гейту
Явні ролі валили oxlint `no-interactive-element-to-noninteractive-role` (false-positive — td/tr/th не
інтерактивні). Розширив justified `oxlint-disable` на обидва правила з поясненням: ролі потрібні, бо
narrow `display:block` **губить implicit table-ролі** (інакше SR і e2e `getByRole('row')` втрачають таблицю).

### Верифікація
- **Ordered `make check` (повний e2e)** зелений: `GATE_EXIT=0`, **e2e 118 passed (16.7m)**, 0 fail
  (лог `codex-s1-gate.log`). Зокрема `admin-users.spec P08` (block/restore/**reset**/delete) ПРОЙШОВ —
  доводить, що `DialogTrigger>Tooltip>Button` reset-trigger працює і name-bearing aria-labels цілі.
- **Browser-verify на реальному prod-коді** (CDP, `localhost:8080`, admin):
  - Desktop (1280): compact icon-дії 3×`44×44` з aria-labels «Block/Delete/Reset password for <name>»;
    sticky `[data-pin=end]` (`position:sticky; inset-inline-end:0`, непрозорий `rgb(17,19,23)`); elevation-
    тінь; `role=table`, 7×`role=row`, rowheader-ідентифікатор.
  - **Narrow 390 (критично):** `tr display:block`, але **`role=row` ЗБЕРЕЖЕНО** (Users 7, Courses 20 —
    явні ролі переживають reflow); обидві таблиці → stacked labeled cards (`::before`: Email/Role/Status/
    Updated · Title/Slug/Status/Owner/Updated); дії на narrow видимі `44×44` з aria-labels, pin→static;
    h-scroll 0.
  - 200% zoom (в'юпорт 640 + text-zoom root 200%): h-scroll 0.

### Обмеження / межі
- Forced-colors live — та сама межа (тулінг не емулює); S1 НЕ вводить нової forced-colors-чутливої
  поверхні (нема gradient/flat CTA), лише токен-адаптивні таблиці. OS-перемикач — власнику.
- `align` (right-align чисел) відкладено на S4 (нема числового consumer'а в S1).
- Narrow card-семантика тримається на явних ролях + `data-label::before` (стандартний responsive-table
  патерн); SR-асоціацію header↔cell при схованому `thead` варто окремо оцінити рев'ю.

### Наступне
Рев'ю Courses 2 → адресувати → PR/merge. Далі S2 (form/card, спільний).
