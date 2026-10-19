# V-track діагностика — view 6 `#users` (UserListView)

**Гілка:** `feat/users-view-designpass` (від main @403ea02). **Автор:** Courses 1. **Старт:** 2026-10-17.
Еталон методу: `_wip-courses-view.md` (view 2 — сусідній list-view, ЗАКРИТИЙ), `_wip-learners-view.md` (view 5).

## ФАКТ: сусідній пройдений list-view (courses, view 2) — канон
`CourseListView` (PR #37/#38, у main) — зразок того, як прод list-view має лягати на proto:
`<PageHeader eyebrow+title+description+actions>` → `<SummaryStrip>` → `.admin-filter-bar` →
`<DataTable>` (card, sortable headers, sticky identifier/actions, numeric align, ≤52rem card-stack,
mobile-sort) з `empty`-слотом (isFiltered→Clear filters) і `footer=<Pagination>`.

## ФАКТ: proto `#users` композиція (src.jsx case 'users')
- `PageHeader` eyebrow **«ACCOUNT ADMINISTRATION»** + title «Користувачі» + **description** +
  **actions=`Pill info` «N users»**.
- `filters('Search users', ['Active','Blocked'], roleSelect)` — спільний admin filter-bar: icon-search +
  status-select + role-select(extra) + Export CSV + Clear filters. **Без sort-select** (сортування — у headers).
- `<DataTable>` (той самий компонент, що courses) columns:
  - `email` → `.person-cell` (`.admin-avatar` initials + `<strong>`name + `<small.email-text>`email).
  - `role` → `.role-badge`.
  - `status` → `Pill` (statusKind).
  - `updated` → текст.
  - actions → block / reset / delete icon-buttons.
  - `emptyAction` = «Clear filters».

## ФАКТ: prod `UserListView` композиція (429 рядків)
- `.admin-list-heading` (`.eyebrow` + `<h1>`) — **без PageHeader, без description**.
- `.admin-filter-bar`: UserSearch (form) + SegmentedControl(status) + Select(role) + **Select(sort: email/name/recent)** +
  `.admin-filter-count` (userCount). ← filter-bar уже спільний (learners D6 вирівняно ДО нього).
- `<Table>` (ПРОСТИЙ примітив, НЕ `DataTable`): плоска таблиця без card-wrap, без sticky columns,
  без mobile card-stack, без sortable headers. columns: account/role/status/updated/actions.
  - account → `.admin-user-identity` (`.admin-status-dot` колір-за-статусом + email primary + displayName secondary).
  - role → `.admin-role-tag` (data-role, кольоровий — **багатший за proto `.role-badge`**).
  - status → `ui-pill ui-pill--variant`.
  - actions → UserActions (block/delete/resetPassword icon-buttons, aria-disabled-pattern, focus-rescue).
- Окремий `<Pagination>` (не у card).
- EmptyState: лише `<p>`hint, **без Clear filters**.

## РОЗБІЖНОСТІ (layout-мова proto / сусіднього courses, якої бракує users) — ранжовано
| # | proto / courses | prod users | обсяг | ризик |
|---|---|---|---|---|
| D1 | PageHeader (eyebrow+title+**description**) | `.admin-list-heading` bare | малий (markup+i18n `lede`) | низький |
| D2 | `.person-cell` avatar(initials)+name+email | `.admin-status-dot`+email+name | середній | initials з displayName; `''`→formerAuthor (як learners Lo07); dot-статус дублюється status-колонкою → прибрати |
| **D3** | `<DataTable>` card/sticky/mobile-stack/sortable | простий `<Table>`+окремий Pagination | **середній (ядро гапу)** | DataTable sort опційний → див. РОЗВИЛКА нижче |
| D4 | empty `emptyAction`=Clear filters | EmptyState лише hint | малий (markup+route handler+i18n) | низький; route `go()` вже вміє reset |
| — | count-pill у header actions | count у filter-bar | — | ОПУСКАЮ: тримаю `.admin-filter-count` (як courses; learners теж опустив header-pill; `ui-pill--info` без стилю) |
| — | `.role-badge` | `.admin-role-tag` (data-role) | — | лишаю prod — багатший (колір за роллю) |

**Ядро гапу = D3** (таблиця не у card, без mobile-stack — головне, що власник побачить side-by-side з courses).

## РОЗВИЛКА D3 — sort interaction (єдине справжнє поведінкове рішення)
Прод має 3 серверні сорти `email|name|recent`; колонки account/role/status/updated. Чистого 1:1 немає
(`name` не має власної колонки, `role`/`status` не серверно-сортовані). `DataTable` sort — опційний/контрольований.
- **B (keep-select):** `DataTable` зі **статичними** headers (sort/onSortChange опущено) + **лишити filter-bar
  Select(sort)**. Зберігає всі 3 сорти; layout-card матчиться; headers не клікабельні (interaction, не layout).
  Мінус: filter-bar має зайвий (проти proto) sort-select; headers не сортують як у courses.
- **C (header-sort, proto-faithful):** sortable headers (account→**name**, updated→**recent**), **прибрати**
  filter-bar Select(sort), **відкинути email-sort** (email — вторинний small-текст; сорт за видимим іменем
  інтуїтивніший). Матчить proto+courses interaction. Мінус: прибирає email-sort (pre-launch cut-freely дозволяє).

**Рекомендація: C** — матчить і proto, і пройдений сусідній courses (headers сортують), а email-sort —
найменш цінний із трьох.

## РІШЕННЯ scope (2026-10-17, Courses 1, вибір власника «C» через AskUserQuestion)
D3 sort = **C (header-sort)**: account-колонка сортує за **name**, updated за **recent**; прибрати
filter-bar Select(sort); відкинути email-sort. Наслідок (для чистого стану на load): default sort
`email`→**`name`** (бо email-колонки більше нема, а `sort=email` на load лишив би жоден header не-активним
і зламав mobile-select). Матчить і proto, і пройдений courses. Закриває альтернативу B (keep-select).

## ⚠️ БЛОКЕР залежності (знайдено 2026-10-17): avatar-примітив лише в PR #41
D2 (person-cell avatar) reuse-ить `.person-cell`/`.admin-avatar` — їх увів **view 5 (learners)**, вони
**лише в гілці `feat/learners-view-designpass` (PR #41), НЕ в `main`** (`git grep` підтвердив: main.css — 0,
learners-гілка — 2). Гілка view 6 — від main → примітиву нема.
- Визначити avatar наново у view 6 → ДУБЛЬ CSS + merge-конфлікт із #41 (той самий клас).
- Merge політика = **squash** (PR #40 squash→403ea02) → стекати view 6 на непзлитий view 5 болісно
  (після squash #41 rebase намагатиметься перенакласти вже-злиті зміни → конфлікти).
- **РЕКОМЕНДАЦІЯ: спершу змерджити PR #41**, тоді view 6 від свіжого main (примітив уже там, чистий diff).
  Це й збігається з початковим handoff («чекаємо merge #41, ДАЛІ views 6/7/8»).

## РЕАЛІЗОВАНО (2026-10-17, owner «Роби мердж і далі»)
PR #41 змерджено (squash→`04f3d92`, avatar-примітив тепер у main); гілка `feat/users-view-designpass` від свіжого main.
- **D1** — `.admin-list-heading`(eyebrow+h1) → `<PageHeader eyebrow+title+description=lede>`. count-pill ОПУЩЕНО
  (тримаю `.admin-filter-count` як courses; learners теж опустив; `ui-pill--info` без стилю).
- **D2** — identity-cell `.admin-user-identity`(status-dot+email+name) → `.person-cell`(avatar `initialsOf` + name-primary + email-secondary mono).
  `displayName===''`→`ratings.formerAuthor`. Status-dot прибрано (дублював status-колонку).
- **D3** — простий `<Table>`+окремий `<Pagination>` → спільний `<DataTable>` (card/sticky-id-actions/numeric/mobile-stack),
  footer=Pagination. **sort-C:** account→`name`, updated→`recent` header-sort; role/status `sortable:false`;
  прибрано filter-bar Select(sort); **default sort email→name** (admin-search.ts:131 + route go() omit name + sortFromKey).
  Actions: UserActions тепер фрагмент (без `.admin-user-actions` wrapper) → DataTable `.table-action-group` (горизонтальні icon-кнопки, як courses).
- **D4** — EmptyState: isFiltered→Clear filters (`onClearFilters` prop + route handler); i18n emptyFiltered/emptyFilteredHint/clearFilters.
- **CSS cleanup** (users-only, орфановано міграцією): видалено `.admin-user-identity`, `.admin-status-dot`(+variants),
  base `.admin-user-actions` + його desktop/mobile override-и, dead mobile `:is(primary,updated)`/`identity` у `.admin-list-results`.
  Retune `.admin-user-primary`(name: sans, no-nowrap, wrap) + `.admin-user-secondary`(email: mono, wrap).
- **Локалі** — + lede/account/sortMobile/emptyFiltered/emptyFilteredHint/clearFilters; − sortBy/sort_email/sort_name/sort_recent (Select прибрано).
- **Тести** — admin-search.test.ts default email→name (6 кейсів + default-query тепер шле sort:name);
  UserListView.test.tsx: +onClearFilters prop, state sort→name, column label «Користувач», sort-handler Select→header-click (name+recent),
  B02 → filtered-empty + working clear-filters (DataTable ховає pagination-footer на empty), + unfiltered-empty тест.

## ВЕРИФІКАЦІЯ (проміжна)
- format(oxfmt)/lint(oxlint+arch 210)/typecheck(tsc) = 0. Targeted unit: UserListView + admin-search — всі зелені (0 failures).
- **Браузер** (Playwright host→localhost:8080, admin-login localhost щоб не origin_forbidden; фронт ребілджено):
  desktop1440 — повна fidelity (PageHeader+опис, filter-bar, DataTable card, person-cell avatar+name+email mono, role-tag, status-pill,
  **КОРИСТУВАЧ ↑ активний на load** = default name-sort+header-sort C працює, horizontal icon-actions, pagination у card).
  mobile390 — чистий card-stack, **overflow=false** (було 409>390). axe **0 violations**. 200%@390 overflow=false.
  Скрини: scratchpad users-desktop-1440.png / users-mobile-390.png.
- **ДЕФЕКТ спійманий браузером** (за зеленими unit/lint): mobile 390 overflow 19px — status `SegmentedControl` (`.ui-segmented`)
  393px не-wrap/не-shrink розпирав single-col grid (`.admin-user-list`) до max-content. `.ui-segmented` уже має `overflow-x:auto`+`max-inline-size:100%`,
  але grid auto-column ріс до max-content контенту. **Фікс:** `.admin-course-list,.admin-user-list { grid-template-columns: minmax(0,1fr) }`
  (капить колонку на container, segmented тоді скролиться своїм overflow-x). Також гартує courses (той самий латентний ризик, 0 візуальних змін).
- Ордерний `make check`: **біжить** (лог frontend/test-logs/users-gate-*.log).

## ОБМЕЖЕННЯ
- proto count-pill у header → тримаю `.admin-filter-count` (консистентно з пройденим courses, не ламаю `ui-pill--info`).
- Status SegmentedControl на <~400px скролиться горизонтально (designed `.ui-segmented` overflow-x:auto) — «Видалений» за межею без скролу. Прийнятно для 4 чипів на 390.
- email-sort відкинуто з UI (sort-C рішення власника); лишається валідним hand-typed `?sort=email` + backend-default (adminUserQuery omit).

## Наступне
- Дочекатись `make check` exit (GATE_EXIT у лозі); записати e2e-рядок; коміт + PR від main; перенести summary в Obsidian.

## ВЕРИФІКОВАНО + ЗДАНО (2026-10-17, Courses 1)
Імплементація scope «C» у робочому дереві — доведено й відкрито PR.
- **Код (коміт `a8f5e13`, 9 файлів):** `UserListView` → `DataTable` (person-cell avatar з `.admin-avatar`/`.person-cell` з #41; sortable headers account→name / updated→recent; role-tag+status-pill лишено; EmptyState Clear-filters). `admin-search.ts` default sort `email`→`name` (інакше жоден header не активний на load + ламає mobile-select). `main.css` нетто −69: прибрано застарілі `.admin-user-identity`/`.admin-status-dot`/`.admin-user-actions` (DataTable тримає actions/identity), додано `grid-template-columns: minmax(0,1fr)` на `.admin-user-list` (status-SegmentedControl не розпирає колонку). Локалі: +lede/emptyFiltered*/clearFilters/account/sortMobile, −sort-select ключі.
- **Гейт:** ордерний `make check` **exit 0, e2e 121 passed** (лог `frontend/test-logs/users-gate-204005.log`). admin-users.spec (P08) не чіпав прибраний sort-select/стару розмітку — рядки за `getByRole('row')`+email, статус-фільтр досі `role=radio`.
- **axe (окремий Playwright-прогін, не в гейті):** wcag2a/2aa/21aa **чисто** desktop + mobile-390; overflow@390×200% **false**. (admin-users e2e axe не покриває — тому окремо.)
- **Браузер (host Chrome, localhost:8080 — origin-пастка S5):** 1440 fidelity ✓; header-sort ✓ (name default без param, клік ОНОВЛЕНО → `?sort=recent`, індикатор ▼). Mobile-reflow не знято скрином (claude-in-chrome залипання в'юпорта 1440) — покрито Playwright-овим 390.
- **PR #42** проти main: https://github.com/b2vv/courses/pull/42. CI блокується біллінгом (не дефект). Далі — views 7 #reports, 8 #refunds, тоді A-батч.
