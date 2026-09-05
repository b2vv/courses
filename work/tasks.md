# Foundation task tracker

Джерело scope: [spec](../docs/foundation/spec.md). Детальні файли, gates і acceptance links: [plan](../docs/foundation/plan.md). Стани: `planned` → `in_progress` → `done` або `blocked`. При завершенні рядок зберігається як archive pointer з посиланням на результат/commit і evidence в ledger; задачі не видаляються.

| ID | Назва | Estimate | Depends on | Status | Result / archive pointer |
|---|---|---|---|---|---|
| F01 | Repository conventions, architecture, spec/plan gates | 1h | — | done | docs/foundation/gates.md — 4/4 critiques, defense green |
| F02 | Go API/worker runtime and health | 1.5h | F01 | done | commit `5608a03`; backend-runtime-evidence.md |
| F03 | TanStack Start/Rsbuild SSR scaffold | 2h | F01 | done | commit `5608a03`; Phase 1 report |
| F04 | React Aria UI kit and design tokens | 2h | F03 | done | commit `5608a03`; ui-evidence.md |
| F05 | Compose, Nginx, CI and infrastructure smoke | 1.5h | F02,F03 | done | commit `5608a03`; smoke PASS |
| F06 | Phase 1 working-result gate | 0.5h | F04,F05 | done | commit `5608a03`; ordered gates PASS |
| C01 | Course domain, use cases, migrations, PostgreSQL | 2h | F06 | done | commit `31e8c39`; course-backend-evidence.md |
| C02 | HTTP API and OpenAPI contract | 1.5h | C01 | done | commit `31e8c39`; 26-case HTTP matrix |
| C03 | Orval client, frontend domain/application/adapters | 1h | C02 | done | commit `4d73072`; frontend-test-evidence.md |
| C04 | SSR catalog/course and states | 2h | C03,F04 | done | commit `4d73072`; 53 Rstest + screenshots |
| C05 | Complete integration, browser and build verification | 1.5h | C04 | done | 14 Playwright + final CI gates; report.md |

Phases 1–3 завершено. Phase 3 закрито слайсом I06: `make check` проходить у порядку маніфесту, exit 0. Phase 4a у роботі (нижче); Phase 4b і Phases 5–8 не починалися.

## Phase 3 — Identity

| ID | Назва | Estimate | Depends on | Status | Result / archive pointer |
|---|---|---:|---|---|---|
| I01 | Identity users/sessions migration and domain | 1.5h | green Identity plan gates | done | `docs/identity/report.md`; migration/domain tests pass |
| I02 | Login/session use cases and adapters | 2h | I01 | done | `docs/identity/report.md`; application/Argon2id/PostgreSQL tests pass |
| I03 | HTTP middleware, cookies and typed runtime config | 2h | I02 | done | `docs/identity/report.md` I03; full gate sequence + live curl evidence; microsecond-clock defect fixed |
| I04 | OpenAPI/Orval, Compose and Nginx auth edge | 1.5h | I03 | done | `docs/identity/report.md` I04; 12-case OpenAPI contract test; audit MEDIUM-1 closed with the `web` zone and a strict `login` zone |
| I05 | Frontend session adapter and localized login | 2h | I04 | done | `docs/identity/report.md` I05; 112 frontend tests; A03/A06 covered; login redirect defect fixed |
| I06 | Client route guards and completion gate | ~11.5h, cut into I06a–I06h below | I05 | done | `docs/identity/i06/plan.md` 4th edition; three GATE 2 refusals recorded in `docs/identity/_wip-i06.md` |

## Pre-deploy hardening

| ID | Назва | Estimate | Depends on | Status | Result / archive pointer |
|---|---|---:|---|---|---|
| S01 | CSP (report-only first), Permissions-Policy, HSTS and `Secure` cookies after TLS | 1.5h | TLS termination | planned | audit LOW-5, LOW-6 |
| S02 | `__Host-` prefix on the session cookie | 0.5h | S01 (needs `Secure`) | planned | Removes the precondition for the duplicate-cookie class entirely: a `__Host-` cookie cannot be set by a sibling or parent domain. Blocked until TLS, because the prefix requires `Secure` and local development runs on http. See `docs/identity/report.md` §Security review finding |

### I06 розрізано на задачі

Оцінка 1.5h була неправдивою; реальна ≈11.5h, і план фази обмежує задачу двома годинами.

Третя редакція казала ≈12h, бо стверджувала, що у `frontend/src` **немає жодного маршрутного
тесту**. Це неправда: `app/composition/catalog-queries.test.tsx:40-68` уже будує роутер на
memory history, `CatalogView.test.tsx:172-184` рендерить `RouterProvider`, `e2e/shell.spec.ts`
перевіряє SSR-вивід, `e2e/recovery.spec.ts:46` вживає `request`-фікстуру, а
`schema_integration_test.go:30-53` — робочий харнес на Testcontainers. Бракує тестових
**файлів**, не тестових **поверхонь**.

Третя редакція також збирала всі тести в одну кінцеву задачу I06g. `docs/identity/plan.md:3`
вимагає, щоб **кожна** задача давала власний доказ верифікації, тож шість задач поспіль без
доказу — порушення правила фази, і колонка «закриває» ставала неперевірною до самого кінця.
Тести переїхали до задач, які володіють поведінкою; у I06g лишилось те, що справді потребує
Compose і фікстур, тобто не може бігти раніше за них.

Колонка називається **owns**: задача володіє рядком і несе його доказ. Усі 49 рядків
`acceptance.md` мають власника; жоден не лишився без задачі, як було в третій редакції, де
без власника висіли G10a, G10b, G12, G13, G14, G22 і G23.

| ID | Назва | Estimate | Depends on | Status | Owns / archive pointer |
|---|---|---:|---|---|---|
| I06a | Кореневий `beforeLoad`: `ensureQueryData`, перевірка куки через `getCookie` у композиції, гілка «куки немає» пише `null` у кеш, `timeoutMs: 1000` на SSR-гілці, дискримінований результат без кидків; `login.tsx` знімає власний лоадер | 2h | GATE 2 green | done | `work/report-i06a.md`; G10, G15, G22, G23; 143 frontend tests, ordered gates through `integration` pass |
| I06b | Гарди в `learn.tsx` і `admin.tsx`: один `requireIdentity` з `fetchQuery`, роль інлайном через `notFound()`, `errorComponent` на `ErrorState`, `pendingComponent`, підписка layout-компонента на `['session']` для G12; видалення `LockedArea` і ключів `locked.*` | 2h | I06a | done | `work/report-i06b.md`; G01, G02, G04, G05, G06, G08, G09, G11, G12, G13, G14, G39; 155 frontend tests, ordered gates through `integration` pass |
| I06c | Звуження `safeReturnTo`: префікс `^/(learn\|admin)(/\|$)`, відмова dot-сегментів, нормалізація кінцевого слеша, збереження `search` | 1h | — | done | `work/report-i06c.md`; G15a, G15b, G16; 158 frontend tests, ordered gates through `integration` pass |
| I06d | Шапка: prop-driven компонент (не `SessionMenu` — `Menu` у кіті зайнятий і GATE 1 меню відхилив), вихід, посилання на адмінку за роллю, **`refetchOnMount: false` обов'язково — інакше G07 червоний** (див. `work/report-i06b.md`), uk/en, доступність | 2h | I06a | done | `work/report-i06d.md`; G19, G20, G24, G25, G26, G29, G31; browser halves of G29/G30 to I06g; 171 frontend tests |
| I06e | Екрани `/learn` і `/admin` на `EmptyState`, uk/en | 0.5h | I06b | done | `work/report-i06e.md`; G03, G40 (browser halves to I06g); screens themselves landed in I06b |
| I06f | `backend/sql/seed-identities.sql` (чистий SQL, `ON CONFLICT (id)`, `disabled_at = NULL`, `LOCK TABLE`), ціль `make seed-identities`, `ci.yml`, README; інтеграційний тест на наявному харнесі | 1.5h | — | done | `work/report-i06f.md`; G32-G37; ordered gates through `integration` pass; G37 relocated to `integration.sh` |
| I06g | Браузерні докази проти Compose: переписаний `route-guards.spec.ts`, сире SSR-HTML через `page.request`, лічильник запитів після гідратації, каденція ревалідації | 1.5h | I06a–I06f | done | `work/report-i06g.md`; G07, G10a, G10b, G17, G18, G21, G21a, G22a, G27, G28 + browser halves of G03/G29/G30/G40; e2e 20/20; `Vary` moved here from I06h |
| I06h | Ціль `check` у порядку маніфесту з `seed-identities`, фінальний впорядкований гейт, запис верифікації й обмежень у `docs/identity/report.md` (`AGENTS.md:9`). **`Vary` уже зроблено в I06g** — тест і рядок, який він доводить, мали лежати разом | 1h | усі | done | `docs/identity/report.md` §I06; G38, G41; `make check` exit 0 across twelve steps |

## Phase 4a — Admin course management

Джерело scope: [spec](../docs/admin/spec.md) (GATE 1 закрито). План, приймальна таблиця на 80 рядків і розріз: [plan](../docs/admin/plan.md), редакція 2.1, GATE 2 зелений ([gates](../docs/admin/gates.md)). Виконавець коду — Codex за бріфами `work/brief-<task>.md`; кожен гейт переганяється в сесії-рев'юері перед комітом.

| ID | Назва | Estimate | Depends on | Status | Owns / archive pointer |
|---|---|---:|---|---|---|
| A01 | Міграція 00003 (`instructor_id`, `course_modules`, `course_lessons`); перепорядкування гейта в `Makefile`/`ci.yml`/`README`/`standards.md`; власник+lock+conflict у `seed.sql`; три фікси `repository_integration_test.go` | 2h | GATE 2 green | done | `work/report-a01.md`; P01, B13–B16; ordered gates through `integration` pass |
| A02 | Доменні правила authoring, порти й use cases на фейках: політика, переходи, slug-мутабельність, інваріанти дерева, дублікати id, дедлайни, скасування | 1.75h | A01 | planned | R24, R25, R39 |
| A03 | PostgreSQL authoring-адаптер і sqlc: читання дерева, транзакційний запис із гардами D-3b/D-3c, guarded delete/transition, owner-scoped list, мапінг `23505` | 2h | A02 | planned | P08–P10, B03, B10, B11, R16–R19, R26 |
| A04 | Авторизаційний шов і три захищені read: `httpserver.Actor`, експорт `Authenticate`/`RequireSameOrigin`, `GET` list/one/instructors, OpenAPI+generate | 2h | A03 | planned | P07, R01–R07 |
| A05 | Create і tree-write: `POST /admin/courses`, `PUT /admin/courses/{id}`, межі розміру/полів, slug, власник, replay, OpenAPI+generate | 2h | A04 | planned | P02–P06, B04–B08, B12, R12, R13, R20–R22, R29 |
| A06 | Життєвий цикл: `POST /status`, `DELETE`, мутаційна половина матриці відмов, OpenAPI+generate | 1.5h | A05 | planned | P11–P14, B01, R08–R11, R14, R15, R23 |
| A07 | Публічна програма: read у репозиторії, схема `CourseDetail`, фронтовий домен/схема/`CourseView`, drift-гейт | 1.5h | A06 | planned | P15, B17 |
| A08 | i18n-борг кіта й розкол локалей: label-props у `Table`/`Pagination`/`Skeleton`, оновлення всіх викликів, `locales/{uk,en}.ts` | 1h | — | planned | P17 |
| A09 | Фронтовий admin-адаптер і application: branded ids, типи дерева, порт, use cases, zod, репозиторій, композиція, інвалідація | 1.5h | A06, A08 | planned | R27 |
| A10 | Admin-список і перебудова shell: `admin.tsx` → layout з `Outlet`, `admin.index.tsx`, некидаючий парсинг search, таби статусів, пагінація, стани, uk/en | 2h | A09 | planned | P16, P18, B02, R34 |
| A11 | Редактор курсу з React Hook Form: діалог створення, редагування/reorder дерева, delete, статуси, власник, конфлікт і відмови, a11y | 2h | A10 | planned | P19, B09, R28, R30–R32, R35, R36 |
| A12 | Деградація і in-flight-гарди обох екранів: 503 на первинному read → `ErrorState` з робочим retry; гард подвійного Save | 0.75h | A11 | planned | R37, R38, R40 |
| A13 | Спільні браузерні сесії: `auth.setup.ts`, `setup`-проєкт, конверсія `route-guards.spec.ts`, вимір логінів за прогін | 1h | A12 | planned | B19 |
| A14 | Браузерна подорож адмінки: `admin.spec.ts` (admin+instructor journeys, direct-API refusal), slug-неймспейс на прогін | 1.5h | A13 | planned | P20, P21, R33 |
| A15 | Впорядкований гейт і звіт: повний `make check` з чистого чекаута, `docs/admin/report.md` | 1h | A01–A14 | planned | B18 |
