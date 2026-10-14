# Admin restyle — звіт фази

Переклад прототипу плагіна (`docs/admin-restyle/proto/`) на наш код і токени. План —
`plan.md`, тикети-зрізи — `tickets.md`. Порядок: S0 → S1 → … → S6.

---

## S0 — токени + shell

**Статус:** реалізовано, повний гейт зелений, browser-verify на реальному коді зроблено;
**на рев'ю Courses 2** перед комітом у main.

### Зроблено (6 файлів, усі у скоупі)
- `tokens.css`: +7 контейнерних токенів (`--space-admin-sidebar/-canvas/-dialog/-prose-measure/
  -prose-min/-tile-min/-icon`), суто additive. Sidebar/canvas/dialog/icon застосовано в shell;
  prose/tile оголошено для наступних зрізів.
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

### Наступне
Рев'ю Courses 2 (intake + ux-review, зокрема ручна перевірка 4 стоп і forced-colors) → адресувати
знахідки → PR у main. Далі S1 (DataTable, спільний, повний e2e).
