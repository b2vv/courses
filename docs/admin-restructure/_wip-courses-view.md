# V-tiraj view 2 — `#courses` → `/admin` (CourseListView) — діагностика

**Дата:** 2026-10-16. **Гілка:** `feat/view-courses-vfidelity`. **Метод:** §3/§6.1 (діагностика → власник → фікс → переперевірка).

## Джерела
- proto: `docs/admin-restyle/proto/src.jsx` case `'courses'` (рядок 26) + `PageHeader`/`DataTable` (16/18).
- prod: `frontend/src/presentation/admin/CourseListView.tsx`.

## Факт: admin-restyle наклав палітру, НЕ структуру
У prod CSS **0 згадок**: `admin-page-header`, `admin-eyebrow`, `admin-page-actions`, `admin-summary-strip`,
`summary-dot`, `admin-table-card`, `admin-table-identifier`, `admin-table-actions`, `table-action-group`,
`ui-pagination`, `ui-icon-tile`. Є лише `.ui-table` (примітив) і `.course-cover` (каталог). Prod-admin
використовує власні `admin-list-heading` + React Aria `Tabs`+`Table` — композиція НЕ як у proto.

## Перелік layout-розбіжностей (V — структура, не палітра)

| # | proto | prod зараз | дія |
|---|---|---|---|
| V-C1 | **PageHeader**: eyebrow «CONTENT MANAGEMENT» + h1 + **description** + primary «Create course» справа | `admin-list-heading`: eyebrow+h1 + count+createAction, **без description** | адаптувати spільний PageHeader (eyebrow+title+intro+actions) |
| V-C2 | **summary-strip**: total / published(dot) / drafts | `admin-list-count` — один рядок «N курсів» | strip-стиль; published/draft split → МЕЖА (див. нижче) |
| V-C3 | **course-cell**: icon-tile + назва + підрядок | title(link) + окрема колонка slug | icon-tile + title + підрядок (slug — те, що маємо) |
| V-C4 | **table-card** `.ui-card.admin-table-card` з sort-хедерами, Actions-колонкою, `.ui-pagination` | React Aria `Table` всередині `Tabs` | привести composition картки таблиці до proto |
| V-C5 | **row-cards ≤52rem** (td `data-label`, стек) | — | адаптив: таблиця→row-cards на вузькому (явна нотатка §4) |
| V-C6 | status-фільтр: chips у filters-барі | React Aria `Tabs` | привести до proto-вигляду (chips/segmented), зберегти a11y |

## Чесні межі — proto показує дані, яких бек admin-контракту НЕ віддає
`AdminCourse` = `{id, slug, title, status, instructorId, updatedAt}` (перевірено `domain/course/course.ts:109`).
Admin-пошук приймає лише `status`+`page` (`app/router/admin-search.ts`). Дотягнути нижче = **бекенд-робота
(нова фаза фіч), а не V (layout)** — той самий принцип, що E6 в editor-пілоті (не вигадувати дані).

| межа | чому відсутнє | рішення V |
|---|---|---|
| B-C1 students/enrollment-count колонка | немає в admin-payload | пропустити, датована межа |
| B-C2 price колонка | лише в catalog `Course`, не в `AdminCourse` | пропустити |
| B-C3 lessons-count підрядок | немає в admin-payload | підрядок = slug |
| B-C4 category + cover-color | немає в admin-payload | generic icon-tile |
| B-C5 текстовий пошук | admin-list API не має search-терму | пропустити (status-фільтр лишається) |
| B-C6 per-row delete | не дротований для admin-списку (поведінка, не layout) | пропустити; edit = перехід у редактор |
| B-C7 published/draft split у strip | треба per-status counts (3 запити / бек-агрегація) | strip показує лише total |

## Підхід (для V-tiraj на всі 8 views)
proto має **один спільний** `PageHeader`/`DataTable`. Чистий tiraj = завести спільні примітиви
(`admin-page-header`, summary-strip, proto-стиль table-card) ОДИН раз у `presentation/ui`, далі застосовувати
по views. Узгоджується з A-віссю (спільний `ui/`-кіт). Editor-пілот зробив PageHeader lesson-scoped —
на цьому зрізі винести спільний і перевести editor на нього пізніше (або лишити, щоб не роздувати цей PR).

## РІШЕННЯ ВЛАСНИКА 2026-10-16
1. **Повна fidelity з беком** (не layout-only): V-трек для кожного view МОЖЕ розширювати бекенд, щоб
   дотягнути дані, які показує proto. Для `#courses` це означає: admin-course-list віддає lesson-count,
   student/enrollment-count, price; підтримує текстовий пошук; per-row delete; published/draft лічильники
   для summary-strip. → V перетворюється на feature-фазу per view.
   - **Застереження (чесно):** `category` і `cover/icon-колір` у proto — ймовірно proto-декор без
     доменного поняття (перевіряє Explore-агент). Дотягнути з беку можна лише те, що Є концептом;
     якщо category/cover не існує в домені — це не «дані, яких бек не віддає», а «поняття, якого нема»,
     і вони лишаються generic. Підтвердити після мапінгу беку.
2. **Спільні примітиви в `presentation/ui`**: завести спільні PageHeader + DataTable (як proto має ОДИН
   спільний) у `presentation/ui`, застосувати по всіх 8 views. Editor-пілот перевести на спільний
   PageHeader пізніше (не в цьому PR, щоб не роздувати).

## Мапінг беку 2026-10-16 (Explore) — ЖОДНОЇ міграції не треба
`backend/internal/course`, sqlc-джерело `backend/sql/queries/courses.sql`.
| пункт | стан | підстава |
|---|---|---|
| lesson-count | дані є, JOIN | `course_lessons`→`course_modules`; патерн у catalog `courses.sql:13-18` |
| student-count | дані є, subquery | `course_grants(course_id,state)` state='active' (`00011`+`00015`) |
| price | дані є, в SELECT | `courses.price_minor/currency` (`00011:4-8`), nullable; catalog уже SELECT-ить |
| пошук | тільки код | `search_vector` GIN є (`00016`); admin-гілка хардкодить search="" (`authoring.go:40`), handler відкидає ключ (`admin_handler.go:301`) |
| delete | ВЖЕ ПРАЦЮЄ | `DELETE /api/v1/admin/courses/{id}` (`admin_handler.go:131`), draft-only (`ValidateDelete`) |
| published/draft counts | тільки код | новий `GROUP BY status`; `CountAuthoringCourses` рахує лише один total |
| **category/cover/icon** | **НЕМА домену** | нема в `domain/course.go` ані в схемі; proto-декор. Дотягнути нема чого → generic icon-tile |

## План слайсів view 2 (`#courses`) — full-fidelity feature-фаза
**S1 — Backend:** розширити `ListAuthoringCourses` (lesson_count LATERAL, active-enrollment count, price);
прокинути `search` (handler→application→SQL через наявний `Query.Search`/`search_vector`); новий
status-counts запит (GROUP BY). Домен `AuthoringCourse`+`AuthoringListInput`+status-counts type; адаптер;
DTO `adminCourseDTO`. Go-тести: unit + repository-integration (Postgres) + HTTP-contract. Гейт зелений.
**S2 — Frontend data:** розширити `AdminCourse`(+lessonCount/studentCount/price)+`AdminCoursePage`(+statusCounts);
admin-course-schema (zod) + мапінг; прокинути `search` у `admin-search` router + ключ `admin.list`. Тести.
**S3 — Спільні примітиви в `presentation/ui`:** `PageHeader`(eyebrow+title+description+actions),
`DataTable`(sort-хедери + Actions-колонка + icon-tile identifier + mobile-sort; поверх наявної адаптивної
`Table`-техніки ≤52rem), `SummaryStrip`, `IconTile`. Тести примітивів.
**S4 — CourseListView проти proto:** PageHeader(CONTENT MANAGEMENT+description+Create course), summary-strip
(total/published/drafts зі statusCounts), пошук+status-фільтр, DataTable(course-cell icon-tile+title+підрядок,
status-pill, students numeric, price UAH numeric, updated, row-actions edit/preview/**delete**-дротування).
i18n uk/en. Тести view.
**S5 — design-pass + гейт + PR:** side-by-side proto↔prod desktop+mobile, закрити розбіжності, axe@200%,
клавіатура, стани; повний `make check`; PR, owner merge.

**Відкрите для власника:** category/cover — доменного поняття НЕМА. «Повна fidelity» тут = або (a) generic
icon-tile без category (рекоменд., нічого не вигадуємо), або (b) завести НОВЕ продуктове поняття
course.category/cover (міграція + це вже не «показати наявне», а нова фіча — окреме рішення).

## РІШЕННЯ ВЛАСНИКА 2026-10-16 (друге): category/cover = варіант (a)
Generic icon-tile без category; НОВЕ продуктове поняття course.category/cover НЕ заводимо (нема в домені —
нема чого показувати; інакше = вигадана фіча). Закриває альтернативу (b) міграції.

## Статус: S3 ЗДАНО (коміт 003891f), далі S4 (CourseListView проти proto)
**S3 done (спільні примітиви в `presentation/ui`, жодна в'юха ще не дротована — S4):**
- `ui/layout.tsx`: **PageHeader** (eyebrow+title+lede+actions) + **SummaryStrip** (labelled `role=group`
  глобальних чисел, decorative dot). Eyebrow = наявний `.eyebrow` (admin→`--color-primary`, контраст ок),
  НЕ proto-`admin-eyebrow` з `--color-faint` (пастка faint→muted).
- `ui/data.tsx`: **DataTable** — in-card editorial таблиця: sortable-хедери, sticky identifier+actions,
  numeric-вирівнювання, ≤52rem card/stack, mobile-sort. **Sort = controlled + OPTIONAL**: передаєш
  `sort`+`onSortChange` → хедери інтерактивні; ні — plain labels (без фейк-афордансу). Таблиця САМА НЕ
  сортує рядки (бек admin-list НЕ має sort-параметра — лише limit/offset/status/search, `admin_handler.go:320`).
- CSS → admin-токени, table-стилі scoped до `.admin-table-card` (плаский `.ui-table` і його 3 споживачі
  недоторкані); кожен label/caption = `--color-muted`, не `--color-faint`.
- **IconTile НЕ заводив окремим компонентом** (рішення нижче). Stories (PageHeader/Summary/CourseTable) — showcase для S4-design-pass.
- Гейт frontend: format-check/lint(210)/typecheck/test(764)/build зелений.

### Рішення S3 (2026-10-16, Courses 1)
- **IconTile згорнуто в наявний `.ui-icon-tile`** (не новий компонент): обгортка в один елемент над уже
  стилізованим класом = bloat (ponytail/YAGNI). Generic tile (owner варіант a) = просто `.ui-icon-tile`.
  Закриває пункт плану «IconTile примітив». Якщо в'юхам знадобляться cover-варіанти — повернутись.
- **Новий `DataTable`, а не розширення `Table`**: proto-DataTable додає sort-хедери+mobile-sort+in-card
  pagination-footer+identifier/actions/numeric; впхати це в `Table` змінило б контракт 3 наявних споживачів
  (Analytics/Users/Courses). `Table` лишається для простих. Закриває «extend Table».
- **DataTable sort = controlled+optional** (не client-side всередині примітиву): бек не має sort-param;
  client-sort однієї server-сторінки з 20 рядків вводив би в оману (сортує 20 зі 100). Примітив чесний і
  reusable; рішення server-sort-vs-none — S4/owner. Закриває «примітив володіє client-sort».

### РІШЕННЯ ВЛАСНИКА 2026-10-16 (сортування): SERVER-SIDE → S4a ЗДАНО (коміт 5ab01f2)
Обрано варіант (1): server-side sort. **S4a (бекенд sort) ЗДАНО:**
- Домен: `SortStudents`/`SortPrice` + `NewAuthoringQuery` (admin-набір: recent/title/students/price);
  `NewQuery` (catalog) лишається {recent,title} — **admin-only сорти НЕ течуть у /courses** (students/price
  → 400 для каталогу). Спільний `newQuery` з allow-set.
- SQL `ListAuthoringCourses`: `sort_by` CASE — title A→Z (C collation), students/price highest-first
  (price NULLS LAST), tie-break до `updated_at DESC, id ASC` (колишній фікс-порядок = 'recent').
- handler читає `sort` (allow-list, empty→400 як search); application дефолтить порожній sort у recent
  (direct-callers не мусять його називати). openapi + regenerate sqlc&Orval.
- Фронт query-path: `AdminCourseQuery`/`AdminCourseSort`+`makeAdminCourseQuery`-валідація, `admin-search`
  router (URL/state/query)+ключ split за sort, repository шле sort лише коли non-default. DataTable reverse-
  кнопка тепер лише з `reverseSortLabel` (key-only таблиця без мертвого toggle). **В'юха ще не дротована.**
- **Напрям fixed per-column (key-only), НЕ bidirectional** (уникнути 8-гілкового ORDER BY + `Dir` у спільному
  Query). Header = «сортувати за колонкою», індикатор = натуральний напрям, без reverse-афордансу. Якщо власник
  захоче двонапрямний toggle — окремий follow-up.
- Гейт: backend format-check/lint/typecheck/test/build/check-generated/**integration** (sort proven проти
  Postgres); frontend format-check/lint/typecheck/test(768). e2e — на S5 (sort ще не досяжний з UI).

**S4b (в'юха) ЗДАНО — коміт 3648c4b.** CourseListView переписано проти proto:
- PageHeader (eyebrow «Керування вмістом», title «Курси» — НЕ «Адміністрування»; h1=курси за proto) +
  SummaryStrip (глобальний snapshot total/published·dot/drafts).
- Filter-bar: submit-форма пошуку (як users-ростер) + status SegmentedControl + лічильник.
- DataTable з реальними колонками: course (generic icon-tile + title-Link + «уроків: N»), status-pill,
  students/price (UAH, «Безкоштовно» для null) numeric, updated; sortable-хедери → server sort; **status
  колонка `sortable:false`** (її ключ не шлеться). Pagination у footer; empty-state замінює таблицю (без pager).
- `DataColumn`+`sortable?`; naturalDirection тримає key-only sort чесним (без reverse-контролу).
- Route navigate через один `go()` — зміна status/search/sort НЕ губить інші; зміна фільтра → page 1.
- i18n uk/en; ключ колонки ціни = `priceColumn` (в admin-блоці вже є `price`-об'єкт). CSS: course-link без
  underline, muted «free».
- e2e+route-guard специ оновлено під «Курси» + pager-less empty (повний make check з e2e — на S5).
- Гейт frontend: format-check/lint(210)/typecheck/test(769) зелений.
- **НЕ зроблено в S4b (→ S4c):** row-actions edit/preview/**delete**. Delete потребує НОВОГО легкого
  application-seam (наявний `authoring.delete` вимагає повного CourseTree; список має лише AdminCourse з
  id+updatedAt+status). Preview-ціль теж під питанням (public page лише для published vs editor-preview).

**S4c (row-actions) — НАСТУПНЕ:** легкий delete-шлях (AdminCourse id+updatedAt+status, draft-only як сервер)
+ confirm-dialog + edit(→editor) + preview; actions-колонка DataTable; тести. Delete лише на draft-рядках
(сервер draft-only — показувати на published = фейк-афорданс). **S5:** design-pass proto↔prod + ПОВНИЙ make check (e2e) + PR.

## Архів: S2 ЗДАНО (коміт 459c469), S3 (спільні примітиви)
**S2 done (frontend data, no view wiring — S4):** domain `AdminCourse`+lessonCount/studentCount/price,
`AdminCoursePage`+statusCounts, `AdminCourseQuery`+search (trim/bound у `makeAdminCourseQuery`); zod
`admin-course-schema` парсить три enrichment-поля (price required-nullable) + strict status-snapshot;
repository шле `search` лише коли непорожній (сервер reject ''), і відкидає відповідь, де частини
snapshot не сумуються в total АБО нефільтрований page.total ≠ snapshot.total; `admin-search` router
несе `search` через URL-контракт/state/query, list-ключ split-иться за терміном. Гейт frontend:
format-check/lint/typecheck/test (758) зелений.
**ПАСТКА (зловлено до релізу, звірено з беком):** `statusCounts` — ГЛОБАЛЬНИЙ знімок; `search`
звужує `pagination.total`, але НЕ snapshot (бек навмисно тримає повну картину —
`authoring_integration_test.go:345-348` «search did not leak into status snapshot»). Тому інваріант
`pagination.total === statusCounts.total` у repository застосовний ЛИШЕ коли `status===undefined &&
search===''`. Перша версія перевіряла лише `status===undefined` → відкинула б валідну пошукову
відповідь. Сума частин (`total === draft+published+archived`) — універсальна, лишається завжди.
**S4 застереження:** summary-strip бере total/draft/published з ГЛОБАЛЬНОГО snapshot (не з page) —
під пошуком/статус-фільтром strip тримає повну картину, це за дизайном беку.

## Архів: S1 ЗДАНО (коміт 9c8debc), S2 (фронт-дані)
**S1 done:** admin-list віддає lessonCount/studentCount/price + status-snapshot + full-text search;
жодної міграції. Доведено: unit (domain/application), HTTP-contract, **repository-integration проти
Postgres зелений** (lessonCount=3, active-only studentCount=2, price, 3/1/1/1, search ізолює title).
openapi + regenerated sqlc & Orval; frontend tsc чистий.
**Контракт для S2 (фронт):** `AdminCourse` +`lessonCount:int64`,`studentCount:int64`,`price:NullablePrice`;
`AdminCoursePage` +`statusCounts:{total,draft,published,archived}`; query-param `search` (minLen1,maxLen100,
empty→400). Orval-моделі вже згенеровано (`adminCourse.ts`, `adminCourseStatusCounts.ts`, `listAdminCoursesParams.ts`).
**S2 TODO:** frontend domain `AdminCourse`/`AdminCoursePage` типи + zod `admin-course-schema.ts` + мапінг
адаптера; прокинути `search` у `admin-search` router state + ключ `admin.list`; тести (schema/mapping/repo).


