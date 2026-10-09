# NorthWind — адмінська дизайн-система (виміряна специфікація)

| | |
|---|---|
| Дата | 2026-10-09 |
| Автор | Courses (design-pass) + власник |
| Референс | Скрінкаст власника «Northwind Labs» ops-дашборд (57 с, 2868×1320, 2026-09-30). Кадри: `docs/design/reference/admin-ref-*.{jpg,png}` |
| Метод | `design-pass` → `reference-ui-reconstruction` (виміряні факти) → `frontend-design` → `anti-slop-design` §3 |
| Механізм | **Наш**: семантичні токени `frontend/src/app/styles/tokens.css` + `tokens.admin.css` (scope `[data-surface="admin"]`). React Aria + `ui-*` класи. Беремо лише ВИГЛЯД референсу; shadcn/Radix/Tailwind-utility-міграцію не робимо. |
| Доповнює | `docs/design/portal-redesign.md` (narrative D-a…D-d + принцип двох щільностей). Ця нота — **точні значення** для D-d. Де розходяться — виграє ця нота (вона виміряна). |

Усі `#hex` і px нижче — зняті **піксельним семплінгом кадрів** (не з пам'яті). Виміри в CIE
наближені; де важить контраст — рахувати перед комітом.

---

## 0. Рішення власника (2026-10-09) — закривають альтернативи

- **D1. Primary адмінки = синій `#5273F2`, 1:1 з референсом.** Поточні `tokens.admin.css`
  використовують тіл `#20b2c5` як primary — це **розбіжність із живим NorthWind** і головний
  кандидат на «вигляд не той». Тіл лишається, але **лише** як success/healthy/load, не як бренд.
  *Закриває:* «тіл як primary».
- **D2. Адмінка dark-only** — як референс. Прибрати light-гілку admin-токенів. Це **реверс**
  попереднього рішення (2026-09-30 «admin слідує light/dark»); допустимо до запуску
  (`[[pre-launch-cut-freely]]`, власник 08.10). *Закриває:* «admin у light-темі».
- **D3. Кнопки primary — глянцевий градієнт, не плоска заливка.** Власник прямо: «якщо зробити
  просто синім — не буде 1 в 1». Деталі в §4.1. *Закриває:* «flat `background: var(--color-primary)`».
- **D4. Форма: markdown-спец (ця нота, джерело правди для коду) + візуальна сторінка-галерея
  (Artifact).** Storybook — пізніше або цією ж ітерацією; обмеження: проєкт забороняє Vite, тож
  потрібен `@storybook/builder-rspack`/`storybook-react-rsbuild`, не дефолтний Vite-білдер.
- **D5. Порядок:** спершу адмінка (D-d), клієнт обговорюємо **окремо після** (власник).

---

## 1. Палітра (dark-only, виміряно)

### 1.1 Поверхні та текст

| Роль | Reference | Токен (нове значення) | Нотатка |
|---|---|---|---|
| Фон застосунку | `#0C0D12` | `--color-background: #0C0D12` | near-black, ледь синій |
| Поверхня (сайдбар, th, menu) | `#111215` | `--color-surface: #111317` | на 1 крок світліша за фон |
| Картка | фон + бордер (`#0D0E11`) | картка = `--color-background` + `--card-border-color` | картки НЕ світліші за фон — читаються бордером/тінню |
| Піднята (hover-рядок, popover, elevated) | `#17181D` | `--color-elevated: #17181D` *(новий токен)* | hover рядків таблиці, активні меню |
| Hairline (дивайдери, бордер картки) | `#17181D`–`#242932` | `--color-hairline: #20232B` | декоративний, 1.4.11 не діє |
| Control-бордер (інпути/кнопки/селект) | — | `--color-border: #656E7B` | ≥3:1 на `#0C0D12` (WCAG 1.4.11); лишаємо як зараз |
| Текст primary | ~білий | `--color-text: #E6E8EC` | заголовки, значення |
| Текст secondary | `#8A8F98` | `--color-muted: #8B95A3` | підзаголовки, meta |
| Мікролейбл UPPERCASE | `#525358` (дуже тьмяний) | `--color-faint: #6B7280` *(новий)* | **parity-виняток**: референс дає ~3:1; ми тримаємо ≥4.5:1 → трохи світліше. Див. §8 |

### 1.2 Бренд і семантика

| Роль | Reference | Токен |
|---|---|---|
| **Primary / бренд / in-progress / canary** | `#5273F2` (бари, активна навігація, CTA, лого) | `--color-primary: #5273F2`, `--color-primary-hover: #6986FF`, `--color-on-primary: #FFFFFF` |
| Success / healthy / позитивна дельта | текст `#4ED6A8`, load-бар `#57B496` | `--color-success: #4ED6A8` |
| Warning / degraded / highlight чарту | `#FCB646` (бурштин) | `--color-warning: #FCB646` |
| Error / failed / incident | текст `#E5695F`, бар `#C65955` | `--color-danger: #E5695F` |
| Info / env-тег «prod» | блакитний-tint `#1C2443` | `--color-info: #5273F2` (= primary) |
| Вторинні дані (duration, allocation) | фіолет `#8668F3` | `--color-viz-purple: #8668F3` *(тільки data-viz)* |
| Focus-ring | — | `--color-focus: #7B93FF` (яскравіший за primary — помітність на dark) |

### 1.3 Серії чартів (порядок)

`#5273F2` (primary) → `#8668F3` (purple) → `#4ED6A8` (teal) → `#FCB646` (amber) → `#E5695F` (red).
Аномалія/виділений діапазон у бар-чарті = `#FCB646` поверх синіх барів (як «Peak» у референсі).

---

## 2. Типографіка

- **Дворівнева система** (сигнатура dev-tooling):
  - **Display / UI** — гротеск. Великі метрики («4.82M», «206 ms», «91.5%»), заголовки.
    `--font-display: 'Space Grotesk Variable'` (вже є). Збіг із референсом по характеру — **певність середня** (точний шрифт референсу не ідентифіковано; Space Grotesk достатньо близький).
  - **Body** — `'Inter Variable'` (вже є). Навігація, опис, тіло.
  - **Mono / utility** — ВСІ дані, ID, одиниці, час, версії, %: «118ms», «v4.21.0», «ip-10-0-4-19»,
    «e81c4f0», «08:00», «62%». `--font-utility` (ui-monospace стек, вже є). `font-variant-numeric: tabular-nums` обов'язково.
- **Мікролейбли секцій**: UPPERCASE, `letter-spacing: 0.08em`, `font-size: 0.6875rem` (11px),
  `font-weight: 700`, колір `--color-faint`. Приклади: `REQUEST VOLUME`, `SERVICE`, `ENVIRONMENT`,
  `RECENT`, `INFRASTRUCTURE`, `LATENCY & ERRORS`.
- **Шкала** (rem):
  | Роль | size / line | weight | font |
  |---|---|---|---|
  | Metric XL | `clamp(1.75, 4vw, 2.5)` / 1 | 750 | display, `letter-spacing: -0.04em` |
  | Page title (H1) | 1.5 / 1.2 | 650 | display |
  | Section title (H2) | 1.0625 / 1.3 | 650 | display |
  | Body | 0.9375 / 1.55 | 400 | body |
  | Data / value | 0.875 / 1.3 | 500 | mono, tabular |
  | Micro-label | 0.6875 / 1.2 | 700 | body, UPPERCASE, +0.08em |

---

## 3. Spacing, радіуси, щільність

- **Spacing-шкала** — лишаємо `--space-1..8` (0.25…2rem). Адмінка щільна: gap між секціями
  `--surface-gap: var(--space-6)` (вже так).
- **Card padding**: `--card-padding: var(--space-4)` (16px, вже так для admin).
- **Радіуси** (виміряно з референсу):
  | Елемент | Reference | Токен |
  |---|---|---|
  | Картка/панель | ~12–14px | `--radius-panel: 0.75rem` |
  | Кнопка primary | pill / ~10px | великий; `--radius-control` або pill (див. §4.1) |
  | Кнопка secondary, інпут, селект | ~8–10px | `--radius-control: calc(panel - 2px)` ≈10px (вже) |
  | Статус-пігулка | ~8px rounded-rect | `--radius-control` (НЕ full-pill) |
  | Env-тег (prod/stg) | ~6px | малий, `0.375rem` |
  | Icon-tile лого/нав | ~10px | `--radius-panel` − |
- **Бордери**: усюди 1px. Картки/дивайдери — `--color-hairline`; контроли — `--color-border`.
- **Тіні**: картки мають ледь помітну тінь підняття (`--card-shadow: var(--shadow-card)`, вже).
  Primary-кнопка — синє glow (§4.1).
- **Щільність таблиці**: рядок ~48–56px, `td` padding `var(--space-3)` (12px). Групові заголовки
  рядків компактніші.
- **Мін. розмір контролу** `--control-min-size: 2.75rem` (44px) — тримаємо (a11y), навіть якщо
  референс візуально щільніший: візуальна висота може бути меншою, тач-таргет ≥44px через padding/hit-area.

---

## 4. Компоненти (правила побудови)

### 4.1 Кнопки

- **Primary (CTA):** глянцевий синій.
  - Фон: `linear-gradient(180deg, #6283FF 0%, #4C6BF0 100%)`.
  - Верхній внутрішній хайлайт: `inset 0 1px 0 color-mix(in srgb, #FFFFFF 28%, transparent)`.
  - Зовнішнє glow: `0 2px 10px color-mix(in srgb, #5273F2 45%, transparent)`.
  - Текст білий `#FFFFFF`, 650, з опційною leading-іконкою (`+`).
  - Радіус: pill для компактних дій тулбара («New deployment»), ~10px для широких («Promote»).
    Дефолт адмінки — великий радіус, pill для toolbar-дій.
  - Hover: градієнт на крок світліший + трохи сильніше glow. Active: прибрати glow, зсув 0.
- **Secondary (neutral):** поверхня `--color-elevated` (`#17181D`) + 1px `--color-hairline`,
  текст `--color-text`, іконка `--color-muted`. Hover → `--color-surface`.
- **Destructive (secondary):** як secondary, але текст+іконка `--color-danger` («Roll back»).
  **НЕ** червона заливка. ⚠️ Поточний `.ui-button--danger` = червоний фон → для адмінки потрібен
  новий варіант `--danger-ghost` (текст-червоний-на-поверхні). Заливку-danger лишаємо для
  критичних незворотних дій у модалках, якщо треба.
- **Icon-only** (refresh, `…`): квадрат ~40px, прозорий фон, іконка muted, hover → elevated.

### 4.2 Статус-пігулки (stateful badge)

Заокруглений rect (`--radius-control`, ~8px), **тонований напівпрозорий фон + яскравіший текст**,
бордер майже непомітний або відсутній. Padding `var(--space-1) var(--space-3)`, mono-текст
0.8125rem/700.

| Стан | bg | текст |
|---|---|---|
| Healthy / Active / Approved | `color-mix(in srgb, #4ED6A8 14%, transparent)` | `#4ED6A8` |
| Degraded / Warning / Pending | `color-mix(in srgb, #FCB646 16%, transparent)` | `#FCB646` |
| Failed / Blocked / Revoked | `color-mix(in srgb, #E5695F 16%, transparent)` | `#E5695F` |
| In-progress / Rolling out / Canary | `color-mix(in srgb, #5273F2 16%, transparent)` | `#8CA6FF` |
| Draining / Neutral / Idle | `color-mix(in srgb, #8B95A3 14%, transparent)` | `#B6BdC8` |

З крапкою-індикатором (`•`) ліворуч для статусу в першій колонці таблиці.

### 4.3 Env/категорія-теги

Малий tinted-tag (менший за статус-пігулку), ~6px радіус, mono UPPERCASE 0.6875rem: `prod`
(info-tint синій), `stg` (neutral/amber-tint), `preview` (purple-tint). Суто для короткої мітки.

### 4.4 Метрика / stat

`ui-stat`: мікролейбл-контекст («REQUEST VOLUME · All services · Production»), велике display/mono
значення, delta-пігулка поряд, порівняльний підпис («vs same window yesterday»). Delta: `ui-delta`
success (зелена) / danger (червона), mono 0.75rem, з `↑/↓` або `+/−pp`.

### 4.5 Meter / load-бар

Горизонтальний трек (`ui-meter`): тонкий (`--space-2` висота), rounded-pill, fill за семантикою
(`neutral`=primary, `success`=teal, `warning`=amber, `danger`=red). У таблиці — inline, з mono-%
праворуч. Поріг: >85% → danger-колір автоматично (як CPU 91% у Fleet). Анімація fill — §6.

### 4.6 Segmented / time-range

`ui-segmented`: контейнер `--color-surface` pill, сегменти прозорі, активний = primary-заливка +
`--color-on-primary`. Time-range (1h/24h/7d/30d) і view-tabs (Requests/Latency/Errors/Saturation)
— цей самий патерн. Hover неактивного → `--color-background`.

### 4.7 Таблиця (ключове для ростера/акаунтів)

- Хедер `th`: `--color-surface`, мікролейбл-стиль (UPPERCASE faint), sticky-опційно.
- Рядок: hairline-дивайдер знизу; hover → `--color-elevated`. Перша колонка — статус-крапка +
  mono-назва (+ muted-підзаголовок типу instance/email).
- **Групові заголовки рядків** (як `us-east-1 · 18 nodes · avg 61% CPU`): компактний рядок із
  chevron-collapse ліворуч, лічильник, агрегат праворуч. Для ростера: групування за курсом/статусом.
- Значення — mono tabular, вирівняні; %/ms/версії моно.
- **Summary-панель праворуч** (як «Fleet summary»): donut + легенда + розподіл-бари. Для ростера:
  всього/активні/заблоковані + donut прогресу.
- Адаптив: `ui-table-scroll` горизонтальний скрол; на вузькому — картки (вже патерн у проєкті).

### 4.8 Фільтр-бар (над таблицею)

Пошук-інпут з іконкою + leading-іконка-кнопки-дропдауни (`Region All ▾`), активний фільтр показує
значення + `×` для скидання, лічильник результатів праворуч («48 results»), Sort-дропдаун,
icon-кнопка налаштувань. Усе в одну лінію, wrap на вузькому.

### 4.9 Dropdown / popover / menu

`ui-popover` на `--color-background` + `--color-border`, пункти — рядки; вибраний має `✓` ліворуч
і підсвічення `--color-surface`/primary-текст. Праворуч пункту — вторинне значення muted
(«Production  24 svc»).

### 4.10 Чарти (тільки адмінка)

- **Bar**: сині бари (`#5273F2`), заокруглені зверху ~2px, рівні проміжки; виділений діапазон —
  amber. Осі: faint-мікролейбли, baseline-grid дуже тонкий.
- **Sparkline**: тонка лінія в кольорі серії, без заливки, у рядку метрики.
- **Donut**: товсте кільце, трек `--color-hairline`, сегменти за семантикою; центр — велике
  display/mono значення + faint-підпис. Невеликі сегменти (fail/rolled-back) помітні окремими дугами.
- Legend: крапка кольору + лейбл + mono-значення, вирівняне праворуч.

### 4.11 Sidebar / навігація

Патерн референсу: **вузький icon-rail** (лого + 5–6 іконок-розділів) + **широкий секційний
сайдбар** (workspace-switcher угорі, пошук `⌘K`, секції з UPPERCASE-лейблами: nav / RECENT /
INFRASTRUCTURE / PROJECTS). Активний пункт = підсвічена пігулка (primary-tint bg + primary-текст +
primary-tint бордер — вже `ui-sidebar-link[aria-current]`, лише перефарбувати під blue). Лічильники
праворуч (Fleet 48), статус-крапки у списках. Для нашої адмінки rail опційний — секційний сайдбар
обов'язковий.

---

## 5. Секції та підсекції — принципи

- **Картка-секція = заголовок + лічильник + дія.** Ліворуч: назва (display) + лічильник-пігулка
  («Service health 24»). Праворуч: контекст-контрол (env-дропдроп) + `…`-меню. Тіло — таблиця/чарт/список.
- **Мікролейбл-контекст** над значенням/чартом: `ОБ'ЄКТ · scope · середовище` faint-UPPERCASE.
- **Футер-рядок картки** (meta): дрібний faint-рядок з умовами/оновленням («Last 24 hours · 30 min
  buckets · Updated 40s ago»), опційно підсвічена пігулка-пік.
- **Права колонка = супровід** (summary/lifecycle), ліва/центр = головні дані. На вузькому — стек.
- **Ієрархія ваги:** display-метрика → section-title → mono-значення → body → faint-meta. Один
  головний акцент на картку.
- **Стани:** завжди проєктувати empty / loading (`ui-skeleton`) / error (`ui-state--error`) —
  не лишати «порожньо».

---

## 6. Рух / анімації

Стримано, інформативно, **завжди під `prefers-reduced-motion`**.

| Що | Як | Тривалість / easing |
|---|---|---|
| Meter/progress fill | `transition: inline-size` (вже) | 240ms ease-out |
| Hover кнопки/рядка/нав | bg/box-shadow transition | 120–160ms ease-out |
| Popover/menu поява | opacity + translateY(4px) | 120ms ease-out |
| «Live»/rolling-out крапка | пульс opacity | 1.6s ease-in-out infinite (reduced → статична) |
| Skeleton shimmer | opacity alternate (вже) | 1.5s |
| Чарт-бари поява | опційно stagger висоти при маунті | ≤300ms, reduced → без |
| Число-метрика | **без** count-up за замовчуванням (дорого, відволікає) | — |

Правило: рух пояснює зміну стану (fill росте, рядок підсвічується), не прикрашає. Нічого, що
рухається постійно, крім live-пульсу й skeleton — і ті глушаться reduced-motion.

---

## 7. Дельта токенів (що міняємо в `tokens.admin.css`)

1. **Прибрати light-гілку** `[data-surface='admin']` (D2). Лишити один dark-набір без
   `@media (prefers-color-scheme)` / `[data-theme]` розгалуження для admin-scope → admin завжди dark.
2. **Перефарбувати** `--color-primary` тіл→синій (§1.2), додати `--color-success` (teal лишається
   тут), `--color-warning`, `--color-danger`, `--color-info`, focus (§1.2).
3. **Додати токени:** `--color-elevated`, `--color-faint`, `--color-viz-purple`, та градієнт/тінь
   primary-кнопки (`--btn-primary-bg`, `--btn-primary-glow`, `--btn-primary-highlight`).
4. **Звірити фон/поверхні** з §1.1 (`#0C0D12` / `#111317` / hairline `#20232B`).
5. **Новий варіант кнопки** `.ui-button--danger-ghost` (текст-червоний-на-поверхні, §4.1) або
   проп у React Aria Button-обгортці.
6. **Primary-кнопка** — перейти з `background: var(--color-primary)` на градієнт+glow+highlight (D3).
7. **Статус-пігулки** — підсилити tint і яскравість тексту під §4.2 (зараз 12%/45%-border → tint
   14–16% / без бордера / rounded-rect).

Усе — лише в admin-scope; client-токени не чіпати (клієнт — окремо, D5).

---

## 8. Parity-обмеження та невідомі (чесно)

- **Мікролейбли:** референс `#525358` на `#0C0D12` ≈ ~3:1 — нижче WCAG 1.4.11 для тексту.
  Ми тримаємо ≥4.5:1 → `--color-faint: #6B7280` (трохи світліше за референс). *Магнітуда:* різниця
  контрасту, рахувати перед комітом. Це свідомий відхід від 1:1 на користь a11y (проєкт-правило).
- **Шрифт display:** точний шрифт референсу не ідентифіковано з відео. Space Grotesk близький за
  характером; якщо власник упізнає конкретний — заміна тривіальна (один токен).
- **Mono:** стек `ui-monospace`; референс, імовірно, має власний web-mono. Якщо треба точність —
  додати `@fontsource` mono (один імпорт). Поки — системний.
- **Easing/тривалості** §6 — підібрані на око під «відчуття» відео (сповільнене відтворення), не
  зчитані покадрово. Уточнити після реалізації на реальному екрані.
- **Світла тема admin** більше не підтримується (D2); якщо знадобиться — виводити окремо, не зараз.
- Контент референсу (ops-дашборд) ≠ наш контент (курси/акаунти/модерація). Беремо **мову стилю**,
  не макет — мапимо компоненти на наші сутності (§4.7 приклади).

---

## 9. Далі

1. **Верифікація сторінкою-галереєю** (Artifact) — свотчі/тайп/компоненти в цих токенах, звірка
   1:1 з відео очима власника.
2. **D-d реалізація** (окремі слайси, кожен — Playwright-скрін 3 ширини + ordered `make check`):
   токени (§7) → кнопки/пігулки/сегменти → таблиця+summary (ростер/акаунти) → чарти → нав/шелл.
   Цільові екрани зі старого handoff: `/teach/learners`, `/admin/learners`,
   `presentation/teach/LearnerOverview.tsx`, `ReviewModerationRow.tsx`, inline `ModerationRow`.
3. **Storybook** (D4) — ✅ підключено 2026-10-09 під Rsbuild-білдер (без Vite):
   `storybook@10.6.1` + `storybook-react-rsbuild@3.6.0`; конфіг `frontend/.storybook/`
   (TanStack Start виключено env-guard'ом `STORYBOOK=true` у `rsbuild.config.ts`); прев'ю
   обгортає все в `data-surface="admin"`; стартові story — `src/presentation/ui/AdminKit.stories.tsx`
   (admin-only). `npm run storybook` (dev) / `npm run build-storybook`. `build-storybook` зелений.
4. **Клієнт** — окремо (D5), після закриття адмінки.
