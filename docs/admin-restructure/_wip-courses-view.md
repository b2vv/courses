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

## Статус: виконую S1 (бекенд).


