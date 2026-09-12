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

Phases 1–3 завершено. Phase 3 закрито слайсом I06: `make check` проходить у порядку маніфесту, exit 0. Phase 4a завершено: `make check` exit 0 (12 кроків, 243 фронт-тести, e2e 27/27), звіт `docs/admin/report.md`. Phase 4b у роботі (нижче); Phase 5 завершено: `make check` exit 0 (330 фронт-тестів, 38 Go-пакетів, e2e 35/35), звіт `docs/media/report.md`. Phases 6–8 не починалися.

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
| A02 | Доменні правила authoring, порти й use cases на фейках: політика, переходи, slug-мутабельність, інваріанти дерева, дублікати id, дедлайни, скасування | 1.75h | A01 | done | `work/report-a02.md`; R24, R25, R39; ordered gates through `integration` pass |
| A03 | PostgreSQL authoring-адаптер і sqlc: читання дерева, транзакційний запис із гардами D-3b/D-3c, guarded delete/transition, owner-scoped list, мапінг `23505` | 2h | A02 | done | `work/report-a03.md`; P08–P10, B03, B10, B11, R16–R19, R26; ordered gates through `check-generated` pass |
| A04 | Авторизаційний шов і три захищені read: `httpserver.Actor`, експорт `Authenticate`/`RequireSameOrigin`, `GET` list/one/instructors, OpenAPI+generate | 2h | A03 | done | `work/report-a04.md`; P07, R01–R07; ordered gates through `check-generated` pass |
| A05 | Create і tree-write: `POST /admin/courses`, `PUT /admin/courses/{id}`, межі розміру/полів, slug, власник, replay, OpenAPI+generate | 2h | A04 | done | `work/report-a05.md`; P02–P06, B04–B08, B12, R12, R13, R20–R22, R29; ordered gates through `check-generated` pass |
| A06 | Життєвий цикл: `POST /status`, `DELETE`, мутаційна половина матриці відмов, OpenAPI+generate | 1.5h | A05 | done | `work/report-a06.md`; P11–P14, B01, R08–R11, R14, R15, R23; ordered gates through `check-generated` pass |
| A07 | Публічна програма: read у репозиторії, схема `CourseDetail`, фронтовий домен/схема/`CourseView`, drift-гейт | 1.5h | A06 | done | `work/report-a07.md`; P15, B17; ordered gates through `check-generated` pass |
| A08 | i18n-борг кіта й розкол локалей: label-props у `Table`/`Pagination`/`Skeleton`, оновлення всіх викликів, `locales/{uk,en}.ts` | 1h | — | done | `work/report-a08.md`; P17; one review return (ErrorState default), kit-wide Cyrillic = 0; gates pass |
| A09 | Фронтовий admin-адаптер і application: branded ids, типи дерева, порт, use cases, zod, репозиторій, композиція, інвалідація | 1.5h | A06, A08 | done | `work/report-a09.md`; R27; gates pass |
| A10 | Admin-список і перебудова shell: `admin.tsx` → layout з `Outlet`, `admin.index.tsx`, некидаючий парсинг search, таби статусів, пагінація, стани, uk/en | 2h | A09 | done | `work/report-a10.md`; P16, P18, B02, R34; gates pass |
| A11 | Редактор курсу з React Hook Form: діалог створення, редагування/reorder дерева, delete, статуси, власник, конфлікт і відмови, a11y | 2h | A10 | done | `work/report-a11.md`; P19, B09, R28, R30–R32, R35 (amended), R36; review round: 6 test fixes; gates pass |
| A12 | Деградація і in-flight-гарди обох екранів: 503 на первинному read → `ErrorState` з робочим retry; гард подвійного Save | 0.75h | A11 | done | `work/report-a12.md`; R37, R38, R40; gates pass |
| A13 | Спільні браузерні сесії: `auth.setup.ts`, `setup`-проєкт, конверсія `route-guards.spec.ts`, вимір логінів за прогін | 1h | A12 | done | `work/report-a13.md`; B19 (budget ≤7, amended); two review rounds; e2e 24/24 twice, total=7/consec=3 |
| A14 | Браузерна подорож адмінки: `admin.spec.ts` (admin+instructor journeys, direct-API refusal), slug-неймспейс на прогін | 1.5h | A13 | done | `work/report-a14.md`; P20, P21, R33; e2e 27/27 twice, budget total=7 |
| A15 | Впорядкований гейт і звіт: повний `make check` з чистого чекаута, `docs/admin/report.md` | 1h | A01–A14 | done | `docs/admin/report.md`; B18; full `make check` exit 0 (12 steps, 243 frontend tests, e2e 27/27) |

## Phase 4b — Identity administration

Джерело scope: [spec-4b](../docs/admin/spec-4b.md) (GATE 1 закрито). План 2.1, GATE 2 зелений ([plan-4b](../docs/admin/plan-4b.md), [gates-4b](../docs/admin/gates-4b.md)). Виконавець коду — Codex; кожен гейт переганяється в сесії-рев'юері перед комітом. Стрижень: `deleted ⊂ disabled` (soft-delete пише і `deleted_at`, і `disabled_at`; інваріант у CHECK).

| ID | Назва | Estimate | Depends on | Status | Owns / archive pointer |
|---|---|---:|---|---|---|
| T01 | Міграція 00004 (`deleted_at`, два CHECK, партковий unique, drop email_key, down); `GetIdentityCredential` +`deleted_at IS NULL`; фікс канарки 00004→00005; сід-фікстури +disposables | 2h | GATE 2 green | done | `work/report-t01.md`; B06, B07, B08; disposables moved to d-range; gates through check-generated pass |
| T02 | Домен: `DeletedAt`, `AccountStatus`, `CanBlock`/`CanDelete` (self/admin-immutable/last-admin), межі пароля | 1h | — | done | `work/report-t02.md`; domain unit tests; gates pass |
| T03 | Application use cases на фейках: `ListUsers`, `SetStatus`, `Delete`, `ResetPassword`; дедлайни, скасування, dependency failure | 1.25h | T02 | done | `work/report-t03.md`; R15, R17; gates pass |
| T04 | PostgreSQL-адаптер + sqlc (READ COMMITTED): escaped list, `LockAdmins`/`LockUserForStatus`, last-admin count + session revoke, `SetUserDisabledAt` (block+restore), soft-delete+revoke, set-password+revoke, mapper, гонки | 2h | T01, T03 | done | `work/report-t04.md`; P03–P06, B03, B05, B11, B12, R09, R10, R12, R16, R18, R22, R23; gates through check-generated pass |
| T05 | HTTP: чотири маршрути на шві 4a, admin-гейт, DTO/межі, error mapping, OpenAPI+generate | 2h | T04 | done | `work/report-t05.md`; P01, P02, B01, B04, R01–R08, R11, R13, R14; gates through check-generated pass |
| T06 | Фронтовий admin-users адаптер/application: ids, zod, репозиторій із мапінгом кодів, use cases, композиція, інвалідація; Orval | 1.5h | T05 | done | `work/report-t06.md`; B09, R19; gates pass |
| T07 | Маршрут `/admin/users` і в'юхи: список, таби, per-row дії, `ResetPasswordDialog`, стани, гард `canAdministerAccounts`, uk/en, a11y | 2h | T06 | done | `work/report-t07.md`; P07, B02, R21, R24; review: button-name + removed a runner-hanging redundant test; gates pass |
| T08 | Браузерна подорож `admin-users.spec.ts` на disposables (спільний storageState, 0 логінів); direct-API відмова; розширення route-guards | 1h | T07 | done | `work/report-t08.md`; P08, R20; e2e 30/30 twice, budget total=7 |
| T09 | Впорядкований гейт і звіт: повний `make check` з чистого чекаута; секція звіту 4b | 1h | T01–T08 | done | `docs/admin/report-4b.md`; B10; full `make check` exit 0 (293 frontend tests, e2e 30/30) |

## Phase 5 — Media upload and transcoding

Джерело scope: [spec](../docs/media/spec.md) (GATE 1 закрито, Q2 амендовано 2026-09-06). План редакції 3.3, GATE 2 зелений після **шести раундів** ([plan](../docs/media/plan.md), [gates](../docs/media/gates.md)); дослідження кодеків — [research-codecs.md](../docs/media/research-codecs.md). Виконавець коду — Codex; кожен гейт переганяється тут перед комітом.

**Три речі, які цей план коштував гейтам і які не можна загубити при виконанні:** ffmpeg воркера — **5.1.9** (не хостовий 7.1.1), і будь-який доказ мусить називати бінар і місце запуску; `/media/*` лишається **403** на edge протягом усієї фази; API-композиція **не має адаптера сховища** — байти не проходять крізь Go API.

| ID | Назва | Estimate | Depends on | Status | Owns |
|---|---|---:|---|---|---|
| M01 | ADR 008 (tus-інгест) і ADR 009 (fMP4/CMAF, `init_N.mp4`) | 0.5h | GATE 2 green | done | `work/report-m01.md`; ADR 008 + ADR 009 (supersedes 005 segment clause) |
| M02 | Міграція `00005_media.sql` з іменованими CHECK і трьома індексами; фікс канарки 00005→00006 | 1.5h | M01 | done | `work/report-m02.md`; B07, B08, B32; integration green against real PostgreSQL |
| M03 | Домен медіа: `VideoAsset`, шість станів і матриця T1–T9, `ClaimToken`, `StorageKey`, `Actor.CanEditLesson`, `SelectLadder`; чотири порти | 1.5h | — | done | `work/report-m03.md`; B17, R27; domain suite + architecture boundary test green |
| M03a | Ревізія доменно-портового шва під §2.1: `AcceptContainer`+`ContainerSniffBytes`, `FailureReason.Permanent()`, три policy-типи, `Load`, дворезультатний `Claim`, `AppendChunk` (замість `BeginUpload`+`Advance`), `DeleteSwept`, `Head`, `ports.Job` | 1h | M03 | done | `work/report-m03a.md`; R43; domain suite green, zero asynq in domain/ports |
| M04 | Use cases на фейках: mint/append/complete/cancel/list/retry, `Process` із гардами claim-token, `Sweep`, `Recover` | 2h | M03a | done | `work/report-m04.md`; R14–R19, R31–R33, R36, R42; 22 fake-port tests green; one build fix at review |
| M05 | Репозиторій: mint-транзакція (`FOR SHARE OF courses`, `CanEditLesson`, generation, `23505`), guarded UPDATE на `(state, claim_token)` | 2h | M02, M04 | done | `work/report-m05.md`; P06, B16, B22, R28, R38; integration exit 0; one harness fix at review |
| M06 | Publish/eviction/recovery/sweep: транзакція публікації з витісненням, класи C5/C6 і C1–C4 | 1.75h | M05 | done | `work/report-m06.md`; B06, B14, B24, B29; integration exit 0; Requeue/Fail added at review (briefing gap) |
| M07 | Filesystem `MediaStorage`: `os.Root`, відмова traversal/symlink, truncate-and-append з `fsync`, атомарний publish, `Free` | 1.5h | M03 | done | `work/report-m07.md`; B11, B12; tempdir suite green (no Docker needed) |
| M08 | ffmpeg `Transcoder`: argv із Apple-константами й `force_divisible_by=2`, парсинг ffprobe, `WaitDelay` — **без потреби в ffmpeg** | 1.75h | M03 | done | `work/report-m08.md`; B13, B21, B28, R20, R21; suite needs no ffmpeg (zero LookPath/Skip) |
| M09 | API-поверхня: три операції на шві 4a, DTO/межі, error mapping, OpenAPI+generate | 2h | M05 | done | `work/report-m09.md`; P01, P02, P07, B01, B04, B15, R01–R09, R22 (amended §18), R37, R39; +asynq producer |
| M10 | tus-сервер: `HEAD`/`PATCH`/`DELETE`, протокол запису D-2d, `MaxBytesReader` | 2h | M07, M09 | done | `work/report-m10.md`; B02, B03, R10, R11, R29, R30, R40; 3 discussion items accepted (plan §19) |
| M11 | Завершення й третій виконуваний файл: sniff першого чанка, дисковий gate, rename у `source/`, commit-then-enqueue; `cmd/upload`; стадія Dockerfile | 1.75h | M10 | done | `work/report-m11.md`; P03, B18, B23, B31 + B34, B35 (new, §20/§21); build exit 0 with the upload stage |
| M12 | Обв'язка воркера: хендлер на mux, гарди claim, `Recover`+`Sweep` до `server.Start`, метрики, `/status`; smoke з `ffmpeg -version` | 1.5h | M06, M08, M11 | done | `work/report-m12.md`; P10, B19, B33; smoke exit 0 printing ffmpeg 5.1.9 |
| M13 | Edge і топологія: сервіс `upload` у Compose, зона й локація `/uploads/` у nginx, README | 1h | M12 | done | `work/report-m13.md`; R12, R13 (pre-transcode half; post half to M16 per §23); smoke exit 0 |
| M14 | Фронтовий адаптер: типи, порт із `cancelUpload`, zod, репозиторій, tus-клієнт, поллінг | 1.75h | M09 | done | `work/report-m14.md`; P11, B09, R23 + R44 (fingerprint refusals, §24); 319 frontend tests green |
| M15 | `LessonVideoField` у `CourseEditorView`: контрол файлу, прогрес, бейджі станів, retry, resume/cancel, uk/en, a11y | 2h | M14 | done | `work/report-m15.md`; P09, B20, R24, R25, R41; 330 frontend tests green with axe |
| M16 | Браузерна подорож: `media.setup.ts` (чотири кліпи), `media.spec.ts` — upload→processing→ready, сходинки, плейлист проти цільового бінаря, рестарт у спостережений `processing` | 1.75h | M13, M15 | done | `work/report-m16.md`; P04, P05, P08, B05, B25–B27, B30, R26, R35, R13 post-half; e2e 35/35 twice |
| M17 | Впорядкований гейт і звіт: повний `make check`, `docs/media/report.md` | 1h | M01–M16 | done | `docs/media/report.md`; B10; full `make check` exit 0 (330 frontend tests, 38 Go packages, e2e 35/35) |

## Беклог, узятий за тригером

| ID | Назва | Estimate | Depends on | Status | Result / archive pointer |
|---|---|---:|---|---|---|
| B02 | B-2: адаптерні тести під `node`, tus доводиться на дроті в адаптерному шарі | 1h | Phase 5 closed | done | бриф `work/brief-b02.md`; звіт `work/report-b02.md`; беклог `work/backlog.md` B-2 §Outcome; амендмент `docs/media/plan.md` §26 |

## Phase 6 — playback

Джерело scope: [spec](../docs/playback/spec.md) (GATE 1 пройдено, 14 рішень G1–G14).
План: [plan.md](../docs/playback/plan.md) **редакція 6**; GATE 2 закрито на ній із названим
списком відкритого (§11). Спайк edge: [spike-edge.md](../docs/playback/spike-edge.md).

**Три речі, які не можна загубити:** відеобайти не йдуть крізь Go і SSR; недоступність
перевірки — відмова, ніколи не пропуск; витіснення кусається на сервері, клієнтський стоп —
лише зручність.

| ID | Назва | Estimate | Depends on | Status | Result / archive pointer |
|---|---|---:|---|---|---|
| T1 | Спайк до справжнього Go: `authorize` з Redis і Postgres, замір під навантаженням | 1.5h | GATE 2 closed | done | закриває **O2**; O1 потребує браузера й переїхав у T7 |
| T2 | Міграція 00006 + `duration_seconds` у воркері, канарка, `lock_timeout` | 1.5h | — | done | |
| T3 | Домен і порти `playback`; Lua з трьома значеннями | 2h | T2 | done | |
| T4 | **Довести** seam наскрізь: `Take → authorize`, edge-докази B12–B14, R29 | 1h | T1, T3 | done | ⟵ T1/D2: T1 уже поставив seam |
| T5 | `play` і heartbeat: контракти, 409/410, прогрес, policy-відмови | 2h | T3, T4 | done | |
| T6 | Читальна модель навчання + спільний тест предиката (R18/R19 закриті T5) | 1.5h | T2 | done | |
| T7 | **Віха:** плеєр на мінімальному маршруті, таблиця відновлення | 2h | T5, T6 | done | після неї існує працюючий продукт |
| T8 | Навчальна поверхня: зміст, позначки, «продовжити» (P02, P03, B19, B26, R24, **R36**) | 2h | T7 | **done** — умову знято: гейти T9 (52/52) і T10 (59/59) ганяли ті самі рядки на HEAD | бриф ffe0567; три рядки посилено рецензентом до запуску — P02 вимагає справжній ланцюг heartbeat→`lesson_progress`, B26 зрізає суфікси множини, R24 доводить відсутність набором маршрутів |
| T9 | Прогрес у плеєрі: каденція, події, засів `clientSeq`, **закриття O1** | 1.5h | T7 | **done** — `make check` EXIT=0, e2e **52/52** на розвантаженій машині | O1 **закрито** заміром: 2 запити в польоті, `stopLoad` їх знімає, `currentTime` не переживає `loadSource`. Знайшов і виправив дефект D-7 — медійний 403 більше не термінальний (R38) |
| T10 | Ліміт і витіснення наскрізь по курсах + **B-7** | 1.5h | T7 | **done** — `make check` EXIT=0, e2e **59/59** | закриває **B-7** і новий рядок **R39**: пауза поза кнопкою зупиняє поновлення володіння **і** пише прогрес |
| T11 | A11y, локалізація, теми, 200% зуму | 1.5h | T8 | **done** — `make check` EXIT=0, e2e **60/60** (було 1 failed, 30 did not run) | закриває **B24a/B25/R35**. Дросель R35 виводиться з `BANDWIDTH` у master, а не пінеться константою — скопійовані з O1 8192 біт/с лежали в 66 разів нижче за найнижчу сходинку, тож знижуватись не було куди. R35 довів **крок** ABR (abandon-and-retry на тому самому індексі сегмента), не осідання. В `incomplete` свідомо лишається рівно один вузол `color-contrast`/`nonBmp` на шевроні `Select`: axe відмовляється судити гліф, це не дефект контрасту. B25 міряє контраст над **порожнім** `<video>` (`readyState=0`), і ця межа тепер ассерт, а не коментар |
| T12 | Проба AV1 у полі мінту | 0.5h | T5 | **done** — `make check` EXIT=0, e2e **60/60**, три засіяні серії зняті з живого `/metrics` | закриває **C19**. Лічильник на **отриманні**, не на успіху: відмовлений мінт теж рахується, бо A17 питає про браузер, а не про сесію. `decodePlaybackObject` розділено на required/optional — механізму опційного поля в ньому не було |
| ADR-1 | ADR 010 — предикат доступу; **плюс 011 (володіння й витіснення) і 012 (що означає 403)** | 0.5h | — | **done** | C2 знято. 011 і 012 заведені понад план: рішення коштували двох днів і в беклозі не шукаються |
| DOC-1 | `AGENTS.md`: скоуп і «media access is denied» | 0.25h | T4 | **done** | C3 знято. Хибних тверджень було **три**, не два: третім був рядок «do not claim working playback… student progress», який після T7–T10 наказував недозаявляти вже виміряне |
| T13 | Впорядкований гейт і звіт фази | 1h | усі | planned | `docs/playback/report.md` |
