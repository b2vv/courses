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

Phases 1–3 завершено. Phase 3 закрито слайсом I06: `make check` проходить у порядку маніфесту, exit 0. Phase 4a завершено: `make check` exit 0 (12 кроків, 243 фронт-тести, e2e 27/27), звіт `docs/admin/report.md`. Phase 4b у роботі (нижче); Phase 5 завершено: `make check` exit 0 (330 фронт-тестів, 38 Go-пакетів, e2e 35/35), звіт `docs/media/report.md`. Phase 6 завершено 2026-09-14 за приймальними рядками: **102/102**, упорядкований `make check` exit 0 (380 фронт-тестів, 48 Go-пакетів, e2e 65/65), звіт `docs/playback/report.md`. Phase 7a (transcript editor) завершено 2026-09-18 (`work/handoff-7a.md`). Phase 8 (ASR) закрито 2026-09-24 за приймальними рядками: **22/22** (A01–A14), упорядкований `make check` `CHECK_EXIT=0` (482 фронт-тести, 34 Go-пакети, e2e 74/74); звіт `docs/asr/report.md`.

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
| T13 | Впорядкований гейт і звіт фази | 1h | усі | **done** — `make check` EXIT=0, e2e **60/60**; але гейт фази **не пройдено**: доказ мають **89 зі 102** рядків §8 | `docs/playback/report.md`. План у §8 заявляв 94 рядки, таблиці містять 102 — `plan.md:511` потребує датованої поправки. Тринадцять рядків без доказу: B07, B08, B20, B21, B21a, B22, B23, B28, B31, B36, R23, R30, R34 |
| T14 | Приймальні рядки фази: B23, B36 (smoke-частина), R30 | — | T13 | **done** — упорядкований `make check` EXIT=0, e2e **60/60**, 379 юніт, 48 Go-пакетів | B23 стверджує пару 429/403, тобто ліміт відсікає надлишок, а не все. R30 — хостова перевірка nginx (allowlist `log_format safe`; кожен `access_log` = `safe` або `off`, хоч один `safe`) плюс Go AST; у Go-контейнері гейта корінь репо не видно, тому nginx-половина на хості. B36 — лише 503 і `Retry-After`; «жодного байта медіа» перенесено в T16. Обмеження: сканер не читає `include` |
| B07 | **Продуктова правка:** читання обмежує позицію тривалістю поточного `ready`-ассета | — | T13 | **done** — `819dcac`; `courses/internal/course/adapters/postgres` **ok**, хвіст гейта `seed-identities → seed → smoke → e2e` **60/60** на тому самому дереві | Не брак доказу, а дефект: запис обрізав (`playback/domain/progress.go:63`), читання ні (`learning.go:33`). Обмеження в домені `course`, тривалість — вхід, що не зберігається, тож контракт не зачеплено. Обидві межі беруть **лише `ready`**-ассет, як `loadLessonDuration` |
| T15 | Приймальні рядки фази: B08, B31, R34 | — | T14, B07 | **done** — упорядкований `make check` EXIT=0, e2e **61/61**, 380 юніт, 48 Go-пакетів; B31 виявився дефектом продукту (медіа на розмонтованому `<video>`), виправлено callback-ref | B07 закрито окремим слайсом (`819dcac`); його тести лежать у тому самому `learning_integration_test.go`, куди йде B08 |
| T16 | Рядки, перенесені на e2e: B20, B21, B21a, B22, B28, R23; клаузула B36 «жодного байта медіа» | — | T15 | **done** — упорядкований `make check` двома частинами, GATE_EXIT=0, e2e **65/65**, 380 юніт, 48 Go-пакетів; фаза **102/102** | поправка плану `8cc1d72`: рівень змінено, вимоги ні |

## Phase 8 — ASR (speech recognition)

Джерело scope: [spec](../docs/asr/spec.md), [plan](../docs/asr/plan.md).
Звіт: [`docs/asr/report.md`](../docs/asr/report.md).

Дві паралельні сесії: Courses 1 (product slice: A05→A06→A07→A09→A10) і Courses 2
(infrastructure: A04, A08, integration tests, A11). A01–A03 зроблено спільно 2026-09-21.

| ID | Назва | Est | Depends | Status | Result |
|---|---|---:|---|---|---|
| A01 | Замір порога `no_speech` | 1h | — | done | механізм замінено (Silero VAD), §4.5 |
| A02 | Звірка рівнів доказу C01…C20 | 0.5h | — | done | spec.md §12.1 |
| A03 | Міграція 00008, домен, порти | 1.5h | A02 | done | D8-15, C21, C22 |
| A04 | MonitorQueue + media-worker | 0.5h | — | done (Courses 2) | format-check, lint зелені |
| A05 | Адаптери рушія і AudioSource | 2h | A03 | done | whisper + media adapters |
| A06 | application.Transcribe | 2h | A03, A05 | done | Process + EnqueuePending + Recover |
| A07 | Recover + таймаут | 0.75h | A06 | done | `da83a4d`, `ff4b01e` |
| A08 | asr-worker, compose, /readyz | 1.5h | A04, A06, A07 | done (Courses 2) | `def2585` |
| A09 | HTTP: стан чернетки, перегенерація | 1h | A06 | done | `ba6a096` |
| A10 | Фронтенд: стан, селект, load-draft | 2h | A09 | done | `2f38064`; 464 фронт-тестів |
| A11 | Publish draft, token conflict | 1h | A09 | done (Courses 2) | `5892174` |
| A12 | Приймальні рядки, гейт, звіт | 1.5h | усі | done | `docs/asr/report.md`; 17/22, 5 на e2e |
| A13 | 5 e2e-рядків (C03, C06 e2e-частина, C11 e2e-частина, C13, C17) проти живого `asr-worker` + тест C18 | 2h | A12 | done | 2026-09-24; виправлено 10 дефектів продукту (`docs/asr/report.md` §4); `make check` `CHECK_EXIT=0` (480 фронт, 33 Go-пакети, e2e 74/74); приймання **19/22**, відкриті C08, C16, C19 |
| A14 | Закрити C08 (integration), C16 (червона гілка `/readyz`), C19 (повідомлення автору + e2e) | 1h | A13 | done | 2026-09-24; `8ecda78` `221b9d3` `3cc8754`; `make check` `CHECK_EXIT=0` (482 фронт, 34 Go-пакети, e2e 74/74); **Phase 8 — 22/22** |

## Phase 9 — готовність до оплати

Джерело scope: [spec](../docs/readiness/spec.md), [plan](../docs/readiness/plan.md), ADR 013.
Рішення користувача 2026-09-24 (`work/PRODUCT-BRIEF.md` §24 п.21): оплата — остання фаза, перед нею
закладаємо механізми й guard-и. Ролі: Courses 1 — план, Codex, гейти, коміти; Courses 2 — рев'ю.

| ID | Назва | Est | Depends | Status | Result |
|---|---|---:|---|---|---|
| P01 | Контракт `access`, межі, unit-таблиця | 1h | — | done | `2780207`; R01–R03 |
| P02 | Три модулі на контракт, мапінг-тест, AST-тест | 1.5h | P01 | done | R04, R05; захист від порожнього суб'єкта; `make check` `CHECK_EXIT=0` (482, 35 пакетів, e2e 74/74) |
| P03 | Таблиця через SQL кожного модуля | 2h | P02 | done | R06–R08, R10: 4 адаптери × 9 випадків на справжньому Postgres; одна таблиця `accesstest` |
| P04 | Чотири точки — один сценарій відмови | 1.5h | P02 | done | `090b912`; R09: 4 контракт-тести RUN+PASS; рецензент перевірив прод-шлях |
| P05 | Фронтенд: «немає доступу» як названий стан + e2e | 1.5h | — | done | `17a5d85`; R11: route-guards R11+G09 + e2e learn-access; прод-ребілд + ширший e2e 71 passed |
| P06 | Шви, R12, гейт, звіт | 1h | P01–P05 | done | R12: grep §4 чисто + architecture-рядок «фікстура не в проді»; звіт `docs/readiness/report.md` R01–R12 по одному; упорядкований `make check` |

## Phase 7b — гілки бесід

Джерело scope: [spec](../docs/discussion/spec.md), [plan](../docs/discussion/plan.md); §24 п.4 з
поправкою 2026-09-28. Оплата лишається останньою (§24 п.21). Ролі: Courses — план, брифи, гейти,
коміти; код — Codex; рев'ю — Courses 2.

| ID | Назва | Est | Depends | Status | Result |
|---|---|---:|---|---|---|
| D01 | Домен `discussion` + `expectedMethods` 3→4 + `viaDiscussion` і гілка межі мітки в крос-тестах | 2h | — | done | `2b8c72e`; K01–K06, K05a; make check exit 0, e2e 75/75 |
| D02 | Міграція `00009` + postgres-адаптер, атомарний cap, гейт-тест селектора, борг селектора в беклог | 3h | D01 | done | K07–K09a; make check exit 0, e2e 75/75; міграція 9 на стенді |
| D03 | Application + HTTP + OpenAPI/Orval + підключення в `cmd/api` | 2h | D02 | done | K10–K15; 33 contract-підтести, assertContract на кожній відповіді; make check exit 0, e2e 75/75 |
| D04 | Фронтенд: адаптер на Orval, панель «Обговорення», мітка → `onSeek` | 2.5h | D03 | done | K16–K18; 3 дефекти продукту знайдено й закрито в слайсі; make check exit 0, e2e 75/75 (B39 → B-23) |
| D05 | e2e: мітка й перемотка, відповідь, модерація, відмова, a11y | 2h | D04 | done | K19–K23; мітку до Play виправлено (дефект продукту); make check exit 0, e2e 79/79 |
| D06 | Упорядкований `make check`, звіт K01–K23 по одному | 1h | D01–D05 | done | `docs/discussion/report.md`; K01–K23 кожен вказує на свій тест; make check exit 0, e2e 79/79 (B39 ✓), backend 501/0; B-23 рядок D06; Phase 7b закрито |

## Робоче місце інструктора

Джерело scope: [spec](../docs/instructor/spec.md) (GATE 1), [plan](../docs/instructor/plan.md)
(GATE 2 зелений); §24 п.3. Розширюємо модуль `discussion` (рішення плану, підтверджене кодом).
Ролі: Courses 1 — план, гейти, коміти; код — Codex; рев'ю — Courses 2. Гілка `feat/instructor-workspace`.
Кожен таск закриває названі рядки приймальної таблиці W01–W26 (`plan.md` §Приймальні сценарії).

| ID | Назва | Est | Depends | Status | Result / закриває |
|---|---|---:|---|---|---|
| I01 | Домен `discussion`: `Report` (правила: не власний, доступ до курсу), course-level guard модерації (owner∨admin), дискримінатор `actor≠author` для журналу, `removed_by` у Comment; unit-таблиці | 2.5h | — | done | `46a94a2`; `NewReport`/`CanBeModeratedBy`/`IsModeration`/`RemovedBy`; W12/W13/W15/W16/W18; go build+`go test -race` домену зелені (docker) |
| I02 | Міграція `00010` (removed_by+CHECK, reports+dismissed_at+UNIQUE+CASCADE, moderation_log, keyset+partial індекси) + postgres `ModerationRepository`: `CourseFacts`, `ListByCourse`, `InsertReport` ON CONFLICT, `DismissReports`, `ListQueue` | 3.5h | I01 | done | `3fb51ee`; адитивно (Remove/log→I03); 2 гейт-фікси (SQL $2::int, ізоляція тесту); integration 7b+W01/W02/W04/W09/W10/W11/W23/W26 зелено |
| I03 | Application (list-by-course, report, queue, dismiss, remove-with-audit) + HTTP (`RegisterTeach` під `/teach`, report під `/learn`, owner-скоуп, коди) + OpenAPI teach/report + Orval + `assertContract` + architecture-тест «без крос-модульного імпорту ownership» | 3h | I02 | done | `3b7b74c`; build+vet+unit+contract+architecture(discussion→course)+check-generated+integration(7b+I02+I03)+typecheck зелено; W03/W14–W20/W22/W24/W25 |
| I04 | Фронтенд: route `teach.*` (guard+loader за зразком admin), огляд + черга, кнопка «Поскаржитись» на уроці, Orval-адаптер, query boundary, стани, a11y | 3h | I03 | done | `f3b7302`; typecheck+lint+architecture(125)+rstest 505/505; гейт-фікс R24 route-tree; W01/W03/W04/W07/W08/W21(частково) |
| I05 | e2e `instructor.spec.ts`: W01–W26 спостережувані журнеї + a11y (W21, селектори в межах сторінки — правило «глобально лічене») | 2.5h | I04 | done | `1de3384`; 8 e2e (W01/W03-W08/W14/W16/W21); e2e викрив 2 дефекти I04 (W01 lesson-title/момент, W05 remove-інвалідація) — полагоджено; повний make check exit 0, e2e 87/87, rstest 505 |
| I06 | Упорядкований `make check`; звіт `docs/instructor/report.md` — W01–W26 по одному на свій тест; `work/tasks.md`; AGENDA/брифінг; PR у main | 1.5h | I01–I05 | done | звіт W01–W26 кожен на свій тест; незалежна звірка знайшла 8 прогалин доказів — закрито gap-тестами; повний make check exit 0, e2e 87/87, rstest 505, integration зелено; **фазу закрито** |

## Оплата (остання фаза роадмапу)

Джерело scope: [spec](../docs/payment/spec.md) (GATE 1), [plan](../docs/payment/plan.md) (GATE 2 зелений,
Courses 2 на opus-4-8); §24 п.21/п.22, ADR 013→014, шви `docs/readiness/spec.md` §5. Розширюємо `access`
(факти Paid/Entitled/Preview) + новий модуль `orders`. Ролі: Courses 1 — план/гейти/коміти; код — Codex (sol);
рев'ю — Courses 2. Гілка `feat/payment`. Кожен таск закриває названі рядки Pay01–Pay27 + Pay04b/Pay21b.

| ID | Назва | Est | Depends | Status | Result / закриває |
|---|---|---:|---|---|---|
| O01 | ADR 014; `access.Course`+=Paid/Entitled/Preview; `CanView`+`PurchaseRequired` (спільний accessExempt); property-тест; таблиця; 4 доменні типи | 2.5h | — | done | `40a724f`; build 57pkg + unit контракту/арх/4 доменів зелені; Pay01/05/15/27; поведінково нейтрально |
| O02 | Міграція 00011 (courses price/currency, lessons is_preview, orders, course_grants, payment_events, grant_audit) + домен `orders`: Money(UAH), Order, Grant, ApplyPaymentEvent-правило; unit | 3.5h | O01 | done | `8430f0e`; build+orders unit зелені; 00011 застосовується; Pay02/03/18 + домен Pay04/16/21; O01/O03 зчеплення (integration red до O03) |
| O03 | 6 запитів джойнять course_grants + читають price/preview; userID через 4 порти; integration-таблиця через SQL у 4 модулях; замір гарячого шляху (EXPLAIN) | 3h | O02 | done | `34745ca`; integration 4 модулі зелені (закрито зчеплення O01/O03); Pay23 partial index; гейт-фікси Pay23-сіда; Pay01(SQL)/11/14/23 |
| O04 | `orders` application+HTTP: старт покупки, статус, адмін-видача, ApplyPaymentEvent (unknown-order не термінальна), grandfather на price-set; PurchaseRequired-код на 4 шарах; каталог price/paid; авторинг ціни/прев'ю; OpenAPI+Orval; architecture-тест | 4h | O03 | done | `a90dfa2`; build+unit+check-generated+integration(orders+4 модулі) зелені; гейт зловив продукт-дефект (grandfather $2::timestamptz); Pay04/06/07/08/10/12/13/16/17/19/20/21/21b/22/26 |
| O05 | Фронт: ціна/«платний» у каталозі; «Купити»→старт; «оплата в обробці» з опитуванням; «потрібна покупка» + прев'ю-позначки; адмін-видача; авторинг ціни/прев'ю; a11y | 3h | O04 | done | `16371d2` (повний зміст платного курсу, замки пер-урок; дефект Pay11), `94b6db6` (UI покупки студента, сторінка замовлення; дефект: ціна не читалась), `81dfa14` (ціна/прев'ю в редакторі, адмін-видача Pay10), `9e239a6` («Шукаємо…»); гейт 532/532 + e2e 88/88 (перший прогін червоний на R23 — відкрита підозра); рев'ю Courses 2 — 0 blocking. Pay07/09/10/11/13(UI) |
| O06 | **WayForPay** (рішення власника 2026-10-01, `provider-research.md` → Decision; доказ Pay25 — `assumptions.md` 2026-10-01). Порт+перший адаптер: checkout(redirect), callback(підпис→ApplyPaymentEvent), CHECK_STATUS, повернення→статус; ключі лише env. Зрізи: **O06a** бекенд — done `aeff91c` (Codex, 5 раундів; рев'ю: гонка statusAt знайдена й виправлена; гейт без e2e); **O06b** фронт-редирект на WayForPay і повернення — planned; **O06c** sandbox-e2e через CHECK_STATUS + один записаний тунельний прогін callback — planned | 3h | O04 | done | O06a `aeff91c`, O06b `3b7213e`, O06c-auto `98499db`; Pay24; Pay25 офлайн-стаб (Pay25-live тунель — відкрито, власник) |
| O07 | Упорядкований make check (exit 0, rstest 539/e2e 89); звіт Pay01–Pay27+Pay04b/21b по рядках — `docs/payment/report.md`; gap-тести Pay17/Pay26 `2f430d3`; PR відкрито | 1.5h | O01–O06 | done | **фазу закрито 2026-10-03**: Pay25-live доведено живою Approved-оплатою 1 грн (callback `POST→200`, ордер `4396fba5…`→`paid`, застосовано рівно раз — `docs/payment/report.md` §6, `203befb`); локальний `make check` exit 0 (rstest 541, e2e 89); PR #3 змерджено в `main` (`df5b3802`). Тест-магазин `site_ex` власник видаляє — ротація ключа не потрібна (assumptions 2026-10-03) |

## Рейтинги й відгуки

Джерело scope: [spec](../docs/ratings/spec.md) (GATE 1 2026-10-03, AC1–AC12, G1–G10),
[plan](../docs/ratings/plan.md) (GATE 2 зелений 2026-10-03: критика 0 blocking, захист зелений після
повторної перевірки, конституція — поріг у `contracts/rating`). Відгуки в модулі `discussion`
(переюз скарг/журналу/черги), агрегат каталогу — SQL модуля `course`, G5 — `orders`. Ролі: Courses 1 —
план/гейти/коміти; код — Codex; рев'ю — Courses 2. Гілка `feat/ratings`. Тест-режим, без прод-сервера.
Кожен таск закриває названі рядки приймання Rt01–Rt29.

| ID | Назва | Est | Depends | Status | Result / закриває |
|---|---|---:|---|---|---|
| T01 | `contracts/rating.Summary` (поріг 3, округлення) + домен `discussion`: `CanReview` (порядок відмов §4), формат 20–2000/1–5, захоплення прогресу; unit-таблиці | 2h | — | in_progress | Rt10, Rt11, Rt15, Rt18; unit-частини Rt06–Rt08 |
| T02 | `orders.GrantCourse`: заборона самогранту (`grant_self_forbidden`), unit + contract + integration | 1h | — | in_progress | Rt24 |
| T03 | Міграція 00013 (reviews, review_reports, журнал: comment_id nullable + review_id + CHECK) + репозиторій discussion (review facts, CRUD з ON CONFLICT, звіти, черга UNION ALL, зняття+журнал однією транзакцією); integration | 3.5h | T01 | planned | Rt02, Rt09, Rt12, Rt13, Rt16, Rt17, Rt20, Rt22, Rt25–Rt27 |
| T04 | Каталог: агрегат `count`/`sum(stars)` + `lessonCount` у `ListPublishedCourses`/`GetPublishedCourse`; `canPurchase` у навчальному курсі; R6-замір (§9) у звіт | 2.5h | T01, T03 | planned | Rt07, Rt09; §9 |
| T05 | Application + HTTP + OpenAPI + Orval (§5): публічний список, власний відгук, PUT/DELETE, скарга, черга з `kind`, зняття адміном; contract-тести | 3h | T02–T04 | planned | Rt01–Rt04, Rt10, Rt11, Rt14–Rt19, Rt21, Rt23–Rt25 |
| T06 | Фронт: картка (★/«Новий курс»/уроки), сторінка курсу (зведення, список, «Показати ще», скарга), форма на навчальній сторінці, кнопка «Купити курс» для адміна, черга інструктора; стани помилок; a11y | 3.5h | T05 | planned | UI-частини Rt01–Rt09, Rt14, Rt21, Rt22, Rt26; Rt28, Rt29 |
| T07 | e2e `ratings.spec.ts` + axe; курси, опубліковані тестом, архівуються | 2.5h | T06 | planned | e2e-стовпець Rt01, Rt03–Rt09, Rt14, Rt21, Rt22, Rt26 |
| T08 | Упорядкований `make check`; звіт `docs/ratings/report.md` Rt01–Rt29 по рядках + R6-замір + обмеження; PR | 1.5h | T01–T07 | planned | усі |
