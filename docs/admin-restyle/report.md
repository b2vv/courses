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

### Рев'ю Courses 2: 0 S1-блокерів, 1 major (pre-existing, у hot-path), 2 minor
Валідувала 3 мої питання: oxlint-disable навантажений (емпірично 5×), clip-path+явні ролі канонічні,
SR header↔cell high-confidence (live-AT — власнику, не claim). Позитивно звірила: overlays default-гілка
незмінна, hover на th[scope=row], sticky-оклюзія, icon aria-повний/tooltip-короткий, усе scoped.
- **MAJOR (виправлено в S1):** block/restore/delete мали `isDisabled={isPending}` — натиснута кнопка
  вимикалась через native `disabled` → фокус у `<body>` ([[isdisabled-on-focused-button-drops-focus]]).
  ⚠️ Перша спроба `isDisabled`→`isPending` (конвенція `teach/ReviewModerationRow.tsx:47`) **провалила
  gate2 на axe**: RAC `isPending` на ICON-ONLY кнопці рендерить `<div role="img" aria-labelledby=…>`
  pending-scaffold без доступного тексту (svg `aria-hidden`) → `role-img-alt` serious ×6 у
  `UserListView.test.tsx`. ReviewModerationRow проходить, бо там кнопка з ВИДИМИМ текстом — тому «2 символи»
  тут не спрацювали. **Фінальний фікс:** кнопки лишаються enabled (focusable); при pending →
  `aria-disabled` (conditional spread під exactOptionalPropertyTypes; не native disabled → фокус цілий;
  не RAC isPending → без role=img) + `useRef(inFlight)`-guard у `runAction` проти даблкліку. Focus збережено
  + axe чистий + подвійний сабміт неможливий.
- **MINOR 1 (виправлено):** empty-state комірка мала `data-label={label}` → на narrow `::before` дублював
  назву таблиці над emptyMessage. Прибрав `data-label` з colspan-порожньої комірки.
- **MINOR 2 (watch-item, не фікс):** можливий подвійний SR-анонс на narrow (col-header зі схованого thead
  + `::before` у NVDA browse-mode). SR-залежно; залишено на живу AT-перевірку (власнику).
- **Re-review (aria-disabled+ref), 2 необов'язкові спостереження — прийнято як є (розсуд):**
  (1) нема візуального busy-стану: `[data-disabled].ui-button` стилізовано, але підхід ставить
  `aria-disabled` → кнопка під час in-flight виглядає активною; дія суб-секундна + toast, паритет —
  опційний tiket. (2) один `inFlight` ref серіалізує дії всіх рядків (клік по іншому рядку під час
  in-flight тихо no-op); вікно крихітне, результат безпечний (без подвійного запису), для нечастих
  admin-дій прийнятно. Повний гейт заради цих косметик — непропорційно; залишено дрібними follow-up.

### Відоме обмеження (pre-existing, окремий follow-up)
Delete робить soft-delete → `UserActions` розмонтовується у `noActions`-span; `isPending` тримає фокус ЛИШЕ
під час операції, але після розмонтування фокус із зниклої кнопки падає у `<body>`. Пересув фокуса після
delete потребує ref+координації; infrequent destructive-дія, pre-existing — окремий дрібний tiket.

**Повторний gate після рев'ю-фіксів:** перша спроба (gate2) ЧЕРВОНА — `isPending` на icon-only →
axe `role-img-alt` (зловлено читанням реального `GATE_EXIT=2`, фон брехав exit 0). Після переходу на
aria-disabled+ref: **gate3 зелений — `GATE_EXIT=0`, test 720 passed (0 fail, axe чистий), e2e 118
passed (14.0m)** (лог `codex-s1-gate3.log`).

### Наступне
Merge PR #24 у main. Далі S2 (form/card, спільний, повний e2e).

## S2a — картки + статус-пілюлі (no-behavior зріз S2)

S2 розбито на **S2a** (cards + pills, лише вигляд, нуль логіки/стану) і **S2b** (форм-поведінка:
label-above, blur-валідація, error-summary з лінками, preserve-input, destructive-confirm,
interaction-states). Цей розділ — про S2a.

### Зроблено (9 файлів: 3 CSS + 6 TSX, +53/-53)
- **`tokens.css`** (+5): доданий відсутній `.ui-pill--warning` (semantic warning text, 12% wash,
  45% mixed border) — у наборі shared-пілюль поруч із success/danger/info/neutral/accent.
- **`tokens.admin.css`** (+10/-1): `[data-surface='admin'] .ui-card { background: surface }`
  (section-картки сидять на surface-заливці, не на page-ground); `.ui-pill--warning` додано до
  admin-групи «transparent border + 14% wash + яскравіший текст» (reference-look).
- **`main.css`** (-43): прибрано мертві base/admin `.admin-course-status`/`.admin-user-status`
  правила (0 читачів після свопу на `.ui-pill`), divider `border-block-start:2px` на
  details/program/lifecycle, дубльоване padding/border/radius/background у `.admin-video-field`
  (їх тепер дає `.ui-card`). Збережено layout grids/flex/gap, акценти, `admin-status-dot`,
  `admin-role-tag`, `admin-video-state`.
- **CourseListView / CourseEditorView / UserListView** (+11/+13/+11): локальні типізовані
  `Record<Status, string>` → `.ui-pill ui-pill--<variant>` замість саморобних `.admin-*-status`
  span. Мапінг: course `draft→warning`, `published→success`, `archived→neutral`;
  user `active→success`, `blocked→danger`, `deleted→neutral`. **Точний i18n-текст статусів
  (`statusLabel`/`t(...)`), `data-status` і DOM-порядок збережено** — свопнуто лише `className`.
- **CoursePriceField / CourseGrantPanel / LessonVideoField** (+1 кожен): `.ui-card` у className
  секції (program/details/lifecycle/price/grant/video), без зміни DOM чи handlers.

### Верифікація
- `npm run format`: 0 змін (Codex-файли вже відформатовані).
- **Ordered `make check` run 1:** format-check / lint / typecheck / test / integration /
  check-generated / build — **зелені**; e2e — **ЧЕРВОНИЙ** на 1 тесті (`GATE_EXIT=2`): R04
  `archive preserves an enrolled member started view` — media-сегмент дав 403 замість 200.
- **Детермінізм-перевірка (ререран e2e з reseed, run 2):** **118 passed (15.2m), `EXIT=0`**,
  R04 пройшов (32.9s). → падіння run 1 — **флак**, не S2a (див. нижче).
- **Контраст кожного pill-варіанта** (незалежний node-WCAG + звірка з **real-browser
  computed styles** через CDP, surface=admin): warning **8.54:1**, success 8.35, danger 4.99,
  neutral 15.16, info 4.93 — усі ≥4.5. Обчислені bg (color-mix 14% wash) збіглися з браузерними
  computed до каналу (warning bg `rgb(46,37,25)`). `.ui-card` admin: bg `#111317` (surface),
  border 1px, radius 12px, padding 16px, shadow — рендериться як задумано (divider→card).
- **Клієнтський warning-pill** (real-browser, обидві теми): light `#865000` на `rgb(240,234,224)`
  = **5.55:1**; dark `#f6c983` на `rgb(44,49,54)` = **8.5:1** — обидві ≥4.5.
- Grep видалених класів `.admin-course-status`/`.admin-user-status` по src+e2e+go: **0 читачів**
  перед видаленням.
- **Рендер-скриншот admin-списку в CDP-сеансі не здобуто:** SSR-Suspense-boundary не
  розрезолвлюється в цьому chrome-devtools-тулінгу (`<main>` лишається `<!--$-->…<!--/$-->`,
  хоча `/api/v1/admin/courses` дає 200 з 20 курсами). Той самий prod-бандл рендериться чисто під
  Playwright у зеленому повному e2e (118/118, axe на admin-сторінках) — тобто це артефакт тулінгу,
  не дефект застосунку (пор. browser-quirks §S0). Перевірку CSS зроблено інжектом елементів у
  живий admin/client-контекст + computed styles (виправдано: пряміше за data-render).

### Знайдене під час верифікації
- **Клієнтська зміна поза admin-скоупом (розкрито):** `MyLearningList.tsx:21` (enrollment UI, не
  admin) вживає `ui-pill ui-pill--warning` для стану `blocked`. До цього зрізу `.ui-pill--warning`
  **не існував** у `tokens.css` → blocked-пілюля падала на базовий `.ui-pill` (без варіанта).
  Новий rule дає їй бурштинове оформлення в обох клієнтських темах. Це латентний фікс (автор явно
  писав `ui-pill--warning`), контраст проходить (5.55 / 8.5), але це **не чисто-адмінська зміна** —
  тому S2a гейтиться **повним** e2e, і клієнтський pill перевірено окремо.
- **R04-флак (pre-existing, не S2a):** `playback.spec.ts:2011` залежить від активного
  `course_grants` members-гранту, встановленого за `waitForTimeout(31_000)` публічного програвання
  перед архівом (En09 ownership: архів не відкликає доступ enrolled-члена → 200). 403 у run 1 =
  грант не був активним на момент `/internal/playback/authorize` — таймінг/стан під навантаженням
  повного e2e, не CSS (diff нуль backend/routes/auth). Тест проходив у гейтах S0 і S1 (118 passed)
  і на рерані. Кандидат на event-driven фікс замість `waitForTimeout` — окремий playback-follow-up.

### Обмеження / межі зрізу
- **No-behavior інваріант тримається:** diff = CSS + class-свопи + додавання `.ui-card` до
  className. Нуль змін логіки/стану/handlers/валідації/DOM-порядку. Форм-поведінка — S2b.
- Forced-colors live — та сама межа, що S0/S1 (тулінг не емулює); S2a вводить `.ui-card` і
  `.ui-pill`, обидва вже покриті forced-colors-блоком (`CanvasText` border, `box-shadow:none`).
- Admin `danger` (4.99) та `info` (4.93) проходять, але близькі до 4.5 — pre-existing пілюлі,
  S2a лише додав `warning` до їхньої transparent-border-групи, їх не чіпав.

### Рев'ю Courses 2: 0 blocker, 1 minor (виправлено в S2a)
Courses 2 звірила всі 6 пунктів незалежно (не з опису) — збіг; перерахувала контраст сама
(admin warning 8.54, danger 4.99). Знахідка:
- **MINOR (виправлено):** `archived`/`deleted` → `.ui-pill--neutral`, але admin-група «transparent
  border + tint» (tokens.admin.css) містила лише success/warning/danger/info — **neutral туди не
  внесли**. Тож на admin neutral падав на shared-стиль: **видимий border + bg `--color-surface`
  (#111317), що зливається з surface-тлом таблиці**, тоді як кольорові pills безбордюрні-тиновані.
  До S2a archived/deleted були безбордюрні-тиновані (admin currentColor 14%) — консолідація під
  reference «no-border» look саме для neutral цю граматику втратила (S2a-введена регресія вигляду,
  не контраст: shared neutral = 15.16). **Фікс:** `.ui-pill--neutral` внесено в admin
  transparent-border-групу + окреме правило `color: var(--color-muted)` + `color-mix(muted 14%,
  background)` — muted ink навмисно (archived/deleted тихіші за статус-сигнал; text-ink 13.27
  відкинуто як надто гучний). Контраст muted `#8b95a3` на `rgb(30,32,38)` = **5.37:1** (node-WCAG,
  звірено з real-browser computed: border → transparent, bg muted-wash). `tokens.css` (клієнт)
  neutral не чіпано → admin-scoped.

**Повторний повний гейт із фіксом:** ordered `make check` — **`GATE_EXIT=0`, e2e 118 passed
(15.3m)** (лог `s2a-gate2.log`, `api-nginx-20261014T131145Z.log`); axe на admin-сторінках звірив
neutral на перебудованому стеку; R04 знову зелений (33.8s) — остаточно підтверджує флак run 1.
Перший green e2e run (до neutral-фіксу) — `api-nginx-20261014T122803Z.log` (118 passed, 15.2m).

### Наступне
Owner-authorized squash-merge PR #25 → main. Далі S2b (форм-поведінка: label-above, blur-валідація,
error-summary з лінками, preserve-input, destructive-confirm, interaction-states).

## S2b — форм-поведінка course-editor (друга половина ticket S2)

S2 розбито на S2a (cards+pills, merged PR #25) і S2b (форм-поведінка). Нарізано Codex'ом
(session `01a13abb`, split impl / e2e), рев'ю + гейти — Courses 1. Рішення перед кодом
(D1–D3) — у session-ноті «## Рішення (S2b)».

### Зроблено (8 файлів: 2 TSX + 1 shared control + 2 CSS + 2 locale + 1 e2e)
- **blur-валідація:** `useForm` на editor (CourseEditorView.tsx:572) і CourseCreateDialog (:872)
  отримали `mode: 'onBlur'` + `shouldFocusError: false` (фокусом керує error-summary, не дефолт RHF).
- **error-summary (лише client-валідація):** `FormErrorSummary` (CourseEditorView.tsx:150) —
  `role="alert"`, зʼявляється на submit-fail, перелічує інвалідні поля (top-level + вкладені
  module/lesson, через `useWatch`) як лінки, що `form.setFocus(name)` фокусують поле; фокус
  переходить у summary на fail (`tabIndex=-1`). Серверний банер (`message`/`messageKind`) незмінний (D2).
- **ref-wiring (суть фокусу):** `controls.tsx` `TextField`/`TextArea` тепер форвардять input-ref
  (React 19 ref-as-prop), `StringField` передає `field.ref` — раніше ref не був підключений, тож
  `setFocus` був no-op. Оскільки ВСІ поля (включно вкладені) йдуть через `StringField`, одна правка
  покрила всі уніформно.
- **destructive-confirm — лише course delete (D1):** кнопка видалення (CourseEditorView.tsx:788)
  обгорнута спільним `Dialog` (overlays render-prop, як Refund/Report queues); confirm викликає
  `removeCourse()` + `close()`. Module/lesson remove і transitions — інлайн, без змін.
- **interaction-states (D3):** `[data-surface=admin] .ui-button[data-pending]` (cursor wait +
  dashed border) + `[data-pressed]` inset. CSS `.ui-error-summary` — shared у tokens.css
  (прототип-грамота: 1px danger border, padding, danger ink). **Корекція (рев'ю Courses 2):**
  `[data-pending]` — generic busy-візуал для RAC `isPending`-споживачів (RefundQueue:140/151,
  LessonDocumentEditor:191), але **НЕ закриває S1-follow-up** — UserListView block/restore/delete
  сидять на `aria-disabled` (283/296), не `data-pending`, тож структурно не матчаться. S1 busy-візуал
  лишається OPEN (потребує окремого busy-сигналу на тих кнопках; глобальний `[aria-disabled]{wait}`
  хибно позначив би й постійно-вимкнені кнопки). Не видаю skipped за done (AGENTS.md).
- **i18n:** `administration.validationSummary` + `administration.confirmDeleteCourse` у uk/en;
  summary-лінки й нові тексти reuse наявних label-ключів (moduleTitle/lessonTitle/courseTitle).
- **e2e (E1/E2, admin.spec.ts):** E1 — очистити «Назва» → submit → `role=alert` + лінк фокусує
  поле (real-browser proof `setFocus`); E2 — UI-create draft → «Видалити чернетку» → confirm-dialog,
  Скасувати повертає фокус на тригер + зберігає draft, Підтвердити → DELETE 204 → список без курсу.
  Обидва самоочисні (унікальний UUID-slug, delete у finally).
- **label-above + preserve-input — вже були (верифіковано, не перебудовано):** controls.tsx
  `.ui-field` label-above (tokens.css:204); save-catch без `reset` тримає вводи на серверній помилці.

### Верифікація
- Codex: format/lint/typecheck зелені; `npm test` 723 passed (component-тести доводять реальний
  фокус input/textarea/вкладених, повторний submit, зняття застарілих summary-записів, create-форму,
  confirm/cancel/Escape + повернення фокусу). (Codex-sandbox лог згадував `EPERM ::1:3000` на bind
  порту rstest — артефакт його пісочниці, у нашому гейті тести зелені.)
- **Повний ordered `make check` (impl, чекпоінт `d761211`): `GATE_EXIT=0`, test 723, e2e 118/118
  (15.4m, `api-nginx-20261014T141826Z.log`).**
- **Повний ordered `make check` (impl + E1/E2 e2e): `GATE_EXIT=0`, e2e 120/120 (15.1m,
  `api-nginx-20261014T145522Z.log`)** — нові flow error-summary-фокус і delete-confirm (+ restore-focus
  на cancel) зелені в реальному браузері. R04 зелений обидва рани.
- Browser (CDP inject, surface=admin, перебудований стек): `.ui-error-summary` danger-border/ink,
  padding 16px, radius 10px; `[data-pending]` cursor wait + dashed. Рендер admin-SPA у CDP не йде
  (SSR-Suspense, артефакт тулінгу) — flow покрито зеленим e2e під Playwright.

### Обмеження / межі зрізу
- Спільна зміна `controls.tsx` (ref-forward у ВСІ форми застосунку) → свідомо гейтилась повним e2e
  (регрес-гард поза admin); 120/120 зелено, регресій форм немає.
- `closeLabel` confirm-діалогу (рев'ю Courses 2 MINOR 1): спочатку reuse `enrollment.report.cancel`
  (крос-namespace). Courses 2 припустила наявний `common.cancel` — його НЕ було (звірив: `common`
  мав лише `close`); тож **додав `common.cancel`** ('Скасувати'/'Cancel') у обидві локалі й перевів
  `closeLabel` на `t('common.cancel')`. Текст 'Скасувати' незмінний (e2e E2 cancel зелений).
- Forced-colors live — та сама межа S0/S1/S2a (тулінг не емулює); `.ui-error-summary`/`[data-pending]`
  токен-керовані.

### Наступне
Owner-authorized squash-merge S2b PR → main. **S2 повністю закрито (S2a+S2b).** Далі S3 (lesson
editor) за планом.

## S3 — lesson editor: візуальний редизайн (друга частина admin-restyle)

Переклад візуальної мови прототипу lesson-studio на наявний `LessonDocumentEditor` за
`docs/lesson-builder/decision-editor-redesign.md`. Нарізано Codex'ом (session `01a13abb`-лінія,
окремий run). Рішення D4 (виключено live-preview) / D5 (editor-only, `.lesson-prose` не чіпати) —
у session-ноті «## Рішення (S3)».

### Зроблено (6 файлів: 1 TSX + 1 CSS + 2 locale + 1 test + report)
- **Section card elevation:** `.lesson-editor-section` отримав `box-shadow: var(--card-shadow)`
  (admin `--card-shadow`=`--shadow-card`, та сама card-мова, що S2a). Surface/border/radius збережено.
- **Ідентичність секції:** у `.lesson-editor-section-bar` додано `SectionIcon` (декоративний inline
  SVG за типом, `aria-hidden`), caps-лейбл типу (`text-transform:uppercase`, letter-spacing, 650) і
  1-based `.lesson-editor-section-index` (muted, tabular-nums).
- **Палітра-тайли:** add-section кнопки → grid-тайли (icon + label + 1-рядковий опис), адаптивна сітка
  `minmax(min(100%, --space-tile-min), 1fr)` (guard проти 200%-overflow). **Accessible name кнопки =
  рівно тип** ('Текст'/'Зображення'/'Завдання'/'Відео'): icon і опис `aria-hidden`, видимий лише label.
- **Toolbar pressed → filled:** `.lesson-editor-toolbar .ui-button[aria-pressed='true']` з underline-
  хака на `color-mix(--color-primary 24%, --color-background)` заливку (`.ui-segment[data-selected]`-
  мова), underline прибрано. `aria-pressed`-семантика незмінна (суто візуально).
- **i18n:** `administration.document.sectionDescription.{video,text,image,row,task}` у uk/en.
- **MUST (верифіковано, не регресовано):** move/delete enabled + no-op за край (не isDisabled),
  save/publish/upload через Query з конфлікт-token, ваги 650/700 (не прото-дрейф), нема паралельних
  токенів / color-mix CTA / прото-shell / localStorage. E (reserved drag-slot) пропущено (дозволено).

### Верифікація
- Codex: format/lint/typecheck/`npm test` 726 passed (component-тести: точні імена всіх 5 palette-
  кнопок; індекс+декоративний SVG+axe після reorder; toggle `aria-pressed` тулбара).
- **Повний ordered `make check`: `GATE_EXIT=0`, e2e 120/120 (14.3m,
  `api-nginx-20261014T163327Z.log`).** lesson-builder.spec (editor + reader-prose region + axe@200%)
  зелений — редизайн не зламав ні editor-флоу, ні student-reader; R04 зелений.
- Контраст (node-WCAG + real-browser computed, surface=admin, перебудований стек): toolbar pressed
  `#e6e8ec` на `rgb(29,37,72)` = **12.10:1**; section-index muted `#8b95a3` на surface `#111317` =
  **6.13**; palette-description muted на elevated `#17181d` = **5.85**; усі ≥4.5. Browser inject
  підтвердив: section box-shadow присутній, icon-tile primary-16%-wash + primary-hover ink,
  pressed `text-decoration:none`.

### Обмеження / межі зрізу
- **D4 — live-preview (segmented «Перегляд»→реальний reader) ВИКЛЮЧЕНО:** не існує сьогодні (net-new,
  не рестайл), а reader `LessonDocumentView` навмисно не імпортований редактором (bundle-separation,
  LessonDocumentView.tsx:18-23) — жива прев'ю затягла б reader у admin-бандл. Окремий майбутній зріз
  (зважити code-split reader). S3-диф не імпортує reader, не додає segmented.
- **D5 — editor-only:** reader-спільний `.lesson-prose` блок (main.css 1421-1478) НЕ чіпано →
  student-reader не зачеплено (підтверджено grep diff + зеленим reader-асертом lesson-builder.spec).
- Forced-colors live — та сама межа S0–S2 (тулінг не емулює); нові поверхні токен-керовані,
  декор-SVG `aria-hidden`.

### Наступне
Owner-authorized squash-merge S3 PR → main. Далі S4 (analytics) за планом row 2 / tickets S4.
