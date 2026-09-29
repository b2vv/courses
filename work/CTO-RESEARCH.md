# CTO research — брифінг репо `courses`

| | |
|---|---|
| Дата | 2026-09-29 |
| Базовий коміт | `2c0d29f` (main; дерево чисте, крім двох незакомічених PNG-скриншотів — байпродукт гейта B-18) |
| Всього комітів у репо | 491 (`git rev-list --count HEAD`) |
| Гейт свіжості цього прогону | **21 коміт** від `65d6065` (попередня база): вся фаза **робоче місце інструктора** (I01–I06: `46a94a2`, `3fb51ee`, `3b7b74c`, `f3b7302`, `1de3384`, `0f90ebd`) + закривні доки + merge PR #2 (`2c0d29f`). Зачепили продукт, межі й ризики. Перезібрано: шапку, §1 (робоче місце інструктора закрите; живий P0 → оплата), §2, §3 (розширення `discussion` модерацією + поверхня `/teach`), §6 (межі інструктора; урок gap-тестів). §4 — додано результат гейта інструктора. §5 — **без змін** (`git log 65d6065..HEAD` по `package.json`/`go.mod` порожній). Історія — у git. |
| Тез із міткою `⚠️ неперевірено` | **2** (§6 — навантаження на зону `media`, без змін; §2 — юрисдикція провайдера оплати) |
| Рішення людини (історичне) | **Оплата — остання фаза** (`PRODUCT-BRIEF.md` §24 п.21). **Модель оплати ухвалена 2026-09-28** (§24 п.22, `86fec52`): разова покупка → grant на курс, один провайдер, одна валюта, без повернень у MVP. Готовність закрита 2026-09-28; Phase 7b — 2026-09-29; **робоче місце інструктора — 2026-09-29** (`docs/instructor/report.md`, W01–W26). **Перед оплатою фаз роадмапу більше немає.** |

Брифінг — кеш первинних джерел, не істина. Розходиться з джерелом — править джерело,
а кеш переписується під нього.

## 1. Вердикт на один екран

`courses` — навчальна платформа: модульний Go-моноліт (**шість** доменних модулів:
`course`, `discussion`, `identity`, `media`, `playback`, `transcript`; плюс допоміжний
`testsupport` — тестовий пакет без шарів, і крос-модульний шар `contracts/` — domain-safe
типи + чисті функції) і **п'ять** окремих виконуваних файлів (`api`, `media-worker`, `upload`,
`asr-worker`, `migrate`), SSR-фронт на TanStack Start, усе в одному Compose-стеку за nginx.

**Поточний стан.** Фази 1–6 закриті; **7a** (транскрипт) — гейт пройдено 2026-09-18, рядкове
приймання незалежно не встановлене (87 рядків без матриці X/X; `docs/transcript/report.md`).
**Phase 8 (ASR) закрита 2026-09-24 за рядками 22/22** (`docs/asr/report.md`; A13/A14 знайшли й
закрили 10 прихованих дефектів). **Phase 9 (готовність до оплати) закрита 2026-09-28 за рядками
R01–R12** (`docs/readiness/report.md`). **Phase 7b (гілки бесід) закрита 2026-09-29 за рядками
K01–K23** (`docs/discussion/report.md`): `make check` exit 0 (backend 501/0, e2e 79/79, B39 ✓).
**Робоче місце інструктора закрите 2026-09-29 за рядками W01–W26** (`docs/instructor/report.md`):
`make check` exit 0 (frontend 505, e2e 87/87) — розширення `discussion` course-wide модерацією,
репортами, журналом і поверхнею `/teach`. Незалежна звірка знайшла 8 рядків без тесту / лише з
доменним правилом — закрито gap-тестами; e2e викрив 2 дефекти (W01/W05), гейти — ще 3 (SQL,
ізоляція тесту, route-tree). Жодна фаза зараз не в роботі.

**Живий P0 — оплата (остання й ЄДИНА фаза роадмапу, що лишилась; `PRODUCT-BRIEF.md` §24 п.21).**
Робоче місце інструктора закрило передостанній рубіж — **перед оплатою фаз більше немає**. **Модель
оплати ухвалена** (§24 п.22, нижче §2): разова покупка → grant, один провайдер, одна валюта, без
повернень у MVP — відкритої людської розвилки немає. Що ще **не** існує: платіжний порт, таблиці
замовлень, enrollment/entitlement, гроші як тип, вебхук-inbox, повернення — задокументовані шви
(`docs/readiness/spec.md` §5), бо AGENTS.md забороняє порти без реальної залежності.

**Наступний хід за агендою:** цей перезбір брифінгу → `spec-flow` фази **оплати** (перше питання
спеки — провайдер і його юрисдикція). Ранжування — `work/AGENDA.md`.

**Архітектурний факт №1 — досі вірний.** Відмова 403 на сегменті нічого не вирішує про сесію
(`docs/adr/012`); рішення ухвалює heartbeat, не сегмент. Володіння — одне на акаунт, не на пристрій
і не на курс (ADR 011). Обмеження позиції тривалістю `ready`-ассета навмисно продубльоване у доменах
(`course/domain/learning.go:27`, `playback/domain/progress.go:63`), а з 7b — **третій і четвертий
читач тієї самої межі**: `discussion`-домен **відхиляє** мітку поза `0..тривалість` (не клампить), і
крос-тест `TestThePositionBoundAgreesAcrossModules` має гілку `discussion`, щоб межі не розійшлися
(B07 був дефектом дублю лише раз, `819dcac`; сама конструкція свідома).

**Архітектурний факт №2.** Розпізнавання мовлення **не** окремий доменний модуль — чернетка й
порти ASR живуть **усередині** `backend/internal/transcript`; `asr-worker` лише виконуваний файл,
що імпортує той самий модуль (`backend/cmd/asr-worker/main.go:14-17`). Причина — §24.19: ASR
закриває одну з задач фази 7a, решта поверхні спільна з транскриптом.

**Архітектурний факт №3.** Воркер ASR пише чернетку **без HTTP-актора й без токена власника**
(`recognize.go` не проходить крізь `CanBeEditedBy`, яку проходять усі HTTP-записи транскрипту).
Конкурентність ловить не роль-перевірка, а оптимістичний замок (`ErrDraftSuperseded`,
`transcript/adapters/postgres/draft.go:205`).

**Архітектурний факт №4 — оновлено 7b: правило доступу тепер у ЧОТИРЬОХ мапінгах.** «Хто може
дивитися курс» зведене в один контракт `backend/internal/contracts/access` — чиста функція
`CanView(subject, course)` над нейтральними типами (лише stdlib, жодного імпорту модуля; ADR 013).
Кожен домен маплить свої типи: `CanBeViewedBy` = рівно `return access.CanView(…)`, і це пінить
AST-тест (`platform/architecture/access_mapping_ast_test.go`, **`expectedMethods = 4`** — 7b підняв
з 3). Чотири мапінги (`course`, `playback`, `transcript`, `discussion`) згодні на всій таблиці
істинності — гілка `viaDiscussion` у `shared_rules_test.go:53`. Домену **дозволено** імпортувати
`internal/contracts/…`, але не інший модуль. Коли прийде оплата, grant — один факт: одне поле в
контракті, одна умова в `CanView`, один join у SQL кожного модуля.

## 2. Продукт

| Питання | Відповідь | Джерело |
|---|---|---|
| Що це | Платформа онлайн-курсів: публічний каталог, вхід, адмінське керування курсами, вивантаження відео, відтворення з прогресом, транскрипт із розпізнаванням мовлення, **гілки коментарів під уроком** | `work/PRODUCT-BRIEF.md` §21, `docs/discussion/report.md` §1 |
| Для кого | Студент (дивиться, читає транскрипт, **коментує момент відео й відповідає**), адміністратор/автор (курси, акаунти, правка чернетки транскрипту, **модерація коментарів курсу через робоче місце інструктора `/teach`**) | `AGENTS.md` §Scope, `docs/instructor/spec.md`, `PRODUCT-BRIEF.md` §24.3 |
| Формат постачання | Compose-стек: `api`, `frontend` (SSR), `media-worker`, `upload`, `asr-worker`, `nginx`, `postgres`, `redis` | `docker-compose.yml`, `backend/cmd/` |
| Що працює й **виміряно** | Плеєр, навчальна поверхня, зміст курсу, позначки переглянутого, прогрес, витіснення сесій, транскрипт (читання/пошук/`?t=`/WebVTT/`<track>`/редактор/бібліотека), ASR-чернетка (22/22), **гілки коментарів (Phase 7b, 79/79 e2e): корінь із міткою часу, пласкі відповіді (cap 200), перемотка кліком по мітці, модерація автором/власником/адміном**; **робоче місце інструктора (W01–W26, 87/87 e2e): огляд коментарів усього курсу, репорти студентів, черга модерації, dismiss, removed_by + журнал** | `docs/playback/report.md`, `docs/transcript/report.md`, `docs/asr/report.md`, `docs/discussion/report.md`, `docs/instructor/report.md` |
| Чого **немає** | Оплат, замовлень, enrollment, другого мовного контенту, продакшн-розгортання. (Гілки коментарів `7b` і робоче місце інструктора — **тепер є**.) | `docs/instructor/report.md` §5, `docs/playback/report.md` §6 |
| Активний P0 | **Оплата** (остання й ЄДИНА фаза, що лишилась, §24.21). Перед нею фаз роадмапу більше немає; модель ухвалена — людської розвилки немає | `PRODUCT-BRIEF.md` §24 п.21/п.22, `work/AGENDA.md` |
| Порядок фаз далі | 7a ✓ · 8 ✓ · 9 ✓ · **7b ✓** · **робоче місце інструктора ✓** → **оплата** (остання, лишилась одна). Борги принагідно: B-22, B-21, B-18, B-23 | `PRODUCT-BRIEF.md` §24; `work/AGENDA.md` |
| Discussion, реалізовано | Домен: тіло trim 1..2000, мітка лише в кореня в `0..тривалість` (інакше `comment_position_invalid`), пласка відповідь до не-прибраного кореня того ж уроку (cap 200 → `comment_thread_full`). HTTP: 4 операції (`GET/POST /learn/lessons/{}/comments`, `PATCH/DELETE /learn/comments/{}`), `sameOrigin` на записах, можливості `canEdit`/`canRemove` рахує сервер. Фронт: панель поряд із транскриптом на Orval, мітка → `onSeek`, `useInfiniteQuery` | `backend/internal/discussion/**`, `docs/discussion/report.md` §2 |
| Instructor workspace, реалізовано | Розширення `discussion` (не новий модуль): `ModerationRepository` (CourseFacts, ListByCourse, InsertReport ON CONFLICT, DismissReports, ListQueue), `removed_by`+журнал (`discussion_moderation_log`), репорти (`discussion_comment_reports`, unique `(comment,reporter)`, `dismissed_at`). HTTP: `POST /learn/comments/{}/reports` + `RegisterTeach` (`GET /teach/courses/{}/comments`, `.../moderation-queue`, `POST /teach/comments/{}/dismiss-report`); прибрання переюзує 7b `DELETE`. Фронт: `/teach` (guard instructor|admin), огляд+черга, кнопка «Поскаржитись». Черга = відкриті скарги, `removed_at IS NULL`; порядок count DESC | `backend/internal/discussion/**`, `frontend/src/routes/teach*`, `docs/instructor/report.md` §2 |
| **Модель оплати — УХВАЛЕНА** (не розвилка) | 2026-09-28 (§24 п.22, `86fec52`): разова покупка → grant на курс (не підписка), один провайдер (перший і єдиний адаптер), одна валюта, без повернень у MVP. **Відкрите лише spec-рівня:** конкретний провайдер і валюта — перші питання самої спеки оплати | `PRODUCT-BRIEF.md` §24 п.22, `docs/readiness/spec.md` §5 |
| Юрисдикція провайдера | ⚠️ **неперевірено:** чи обслуговує запропонований провайдер (Stripe звучав як приклад) юрисдикцію продавця — перше питання спеки оплати, тут не прийняте | `PRODUCT-BRIEF.md` §24 п.22 |

## 3. Кодова база й seams

Модулі бекенду (`backend/internal/`): **шість** доменних — `course`, `discussion`, `identity`,
`media`, `playback`, `transcript` — плюс `testsupport` (тестовий пакет, **без** шарів) і
крос-модульний шар `contracts/` (ADR 013): domain-safe типи + чисті функції, лише stdlib, жодного
I/O й імпорту модуля; домен **може** його імпортувати, модуль модуль — ні. Кожен доменний тримає
`domain` / `application` / `ports` / `adapters`, залежності дивляться **всередину**. П'ять
виконуваних файлів: `api`, `media-worker`, `upload`, `asr-worker`, `migrate` (`backend/cmd/`).

| Межа | Що її перетинає | Джерело |
|---|---|---|
| **nginx ↔ playback** — несучий шов Phase 6 | Байти медіа віддає nginx, Go їх ніколи не пропускає; `auth_request` на кожен плейлист/сегмент | `infrastructure/nginx/nginx.conf:55-86` |
| Права на дереві медіа | Транскодер пише `0750`/`0640`; nginx читає за групою | `AGENTS.md` §Architecture |
| api ↔ media-worker ↔ asr-worker | Три виконувані файли, спільна черга Asynq/Redis, різні `queue`-імена (`media`, `asr`), обидва в метриці `courses_queue_depth`; `media-data` монтують лише воркери, назовні байти віддає nginx | `backend/cmd/asr-worker/main.go:14-17`, `docker-compose.yml` |
| **Новий модуль `discussion` (Phase 7b)** | Повний доменний модуль: `discussion/domain` (правила тіла/мітки/відповіді/дозволів), `application`, `ports`, `adapters/{postgres,http}`. Постгрес-адаптер — ручний SQL (не sqlc), список курсором, атомарний cap відповідей. Читає факти курсу за `lessonId` своїм SQL і маплить у контракт доступу | `backend/internal/discussion/**`, `docs/discussion/report.md` |
| **Міграція `00009_discussion.sql`** | FK з `ON DELETE CASCADE` від уроку; keyset-індекси `(lesson_id, created_at DESC, id DESC)` для коренів і `(parent_id, created_at, id)` для відповідей; `body IS NOT NULL` (дефект: живий коментар із NULL-тілом проходив CHECK — знято рев'ю міграції) | `backend/migrations/00009_discussion.sql`, `docs/discussion/report.md` §3 |
| **Comment-ендпоінти в OpenAPI/Orval** | 4 операції (`listLessonComments`, `publishLessonComment`, `updateLessonComment`, `removeLessonComment`) в `openapi.yaml`; фронт читає їх через Orval, `check-generated` у ланцюгу гейта | `backend/openapi/openapi.yaml`, `frontend/orval.config.ts` |
| **Робоче місце інструктора — розширення `discussion`, не новий модуль** | `ModerationRepository` (окремий порт), `moderation.go` (ручний SQL: course-facts, ListByCourse keyset, reports ON CONFLICT, dismiss, queue), міграція `00010` (`removed_by`+CHECK, `discussion_comment_reports`, `discussion_moderation_log`), teach+report операції в OpenAPI/Orval, поверхня `/teach`. Ownership лишається читанням `courses.instructor_id` **всередині** discussion — пінить architecture-тест `discussion` не імпортує `course` (`platform/architecture/discussion_ownership_test.go`) | `backend/internal/discussion/**`, `backend/migrations/00010_instructor_moderation.sql`, `docs/instructor/report.md` |
| **Правило доступу — один контракт, ЧОТИРИ мапінги (ADR 013 + 7b)** | `contracts/access.CanView` — чиста функція; `course`/`playback`/`transcript`/`discussion` маплять свої типи через `CanBeViewedBy` (тіло = рівно `return access.CanView(…)`, AST-тест `expectedMethods=4`); усі чотири згодні на таблиці (`shared_rules_test.go:53` гілка `viaDiscussion`) | `backend/internal/contracts/access/access.go`, `platform/architecture/access_mapping_ast_test.go` |
| **Машинний автор ASR обходить `CanBeEditedBy`** | `recognize.go` (воркер-шлях) пише без акторської перевірки; HTTP-записи її проходять; конкурентність ловить оптимістичний замок | §1 факт №3 |
| **Тестова фікстура доступу не на прод-шляху** | `contracts/access/accesstest` імпортується лише `_test.go`; architecture-тест відмовляє прод-імпорт | `platform/architecture/access_fixture_isolation_test.go` |
| Клієнт ↔ сервер | Orval-моделі лишаються в адаптерах, явний мапінг у доменні сутності (в т.ч. новий `discussion` і — після B-20 — `draft-repository.ts` на `./generated/courses`) | `frontend/src/adapters/api/generated`, `AGENTS.md` §Architecture |
| **`course` ↔ `playback` ↔ `discussion`: одна межа позиції в кількох доменах** | Обмеження позиції тривалістю `ready`-ассета: `course`/`playback` кламплять, `discussion` відхиляє; крос-тест тримає межі збіжними | `learning.go:23-33`, `progress.go:60-66`, §1 факт №1 |
| **Плеєр ↔ елемент `<video>`** | Callback-ref, ефект медіа залежить від змонтованого елемента | `LessonPlayer.tsx:48-52,226-257,423` |
| Дві гілки відтворення | `Hls.isSupported()` перший; нативна гілка в CI недосяжна, справжній Safari не ганяли | `hls-media.ts:81` |
| Документи курсу — окрема бібліотека | Репліка несе `documentId`, бібліотека їде раз на курс (вкладені давали 75.8 проти 12.8 MiB) | `docs/transcript/report.md` §4 п.1 |
| Гейт «логи без URI» | AST-тест Go-половини + хостовий скан nginx-конфігу | `backend/platform/architecture/safe_request_logs_test.go` |

ADR, обов'язкові перед відповідними шарами: `006` transcript model, `010` предикат доступу,
`011` володіння й витіснення, `012` що означає 403, **`013` одне правило доступу як контракт
(перед будь-якою зміною доступу — включно з 4-м мапінгом `discussion` — і спекою оплати)**.

## 4. Інфра й доставка

| Питання | Відповідь | Джерело |
|---|---|---|
| Гейти, у порядку | `format-check`, `lint`, `typecheck`, `test`, `integration`, `check-generated`, `build`, `up`, `drain-queue`, `seed-identities`, `seed`, `smoke`, `e2e` | `Makefile:83` (`check:`) |
| CI жене те саме | Так | `.github/workflows/ci.yml` |
| Гейт робочого місця інструктора (закриття) | `make check` exit 0 на `2c0d29f`: backend 0 FAIL, frontend **505**, e2e **87/87** (11.4 хв, +8 instructor-рядків, повна регресія); включає gap-тести W10/W12/W13/W15/W17/W19/W20/W26 | `docs/instructor/report.md` §4 |
| Гейт Phase 7b (закриття) | `make check` exit 0: backend 501/0, e2e **79/79**, B39 (B-23) ✓ | `docs/discussion/report.md` §4 |
| Гейт Phase 9 / Phase 8 | R01–R12 `CHECK_EXIT=0` (483 фронт, e2e 75/75); ASR 22/22 (482 фронт, e2e 74/74) | `docs/readiness/report.md`, `docs/asr/report.md` §5 |
| Пастка `check-generated` | Порівнює з **індексом**; незастейджена перегенерація дає хибний червоний | `Makefile:59-61` |
| **Пастка docker (гейт)** | `credsStore: desktop` у `~/.docker/config.json` вішає гейт на `load metadata` (`docker-credential-desktop`). Обхід: temp `DOCKER_CONFIG` = симлінки на `~/.docker/*` + `config.json` **без** `credsStore`; порожній `{}` ламає `docker compose` (CLI-плагін у `~/.docker/cli-plugins`) | пам'ять `docker-credential-helper-hang`; гейт 7b D06 |
| `.asr-assets/` | Стенд заміру ASR (8.7 GiB), не фікстури гейту, у `.gitignore` | `docs/asr/plan.md` §6 |
| Розгортання | Немає, не заявлено | `AGENTS.md` §Scope |
| Якорі боргу в коді | 0 `TODO`/`FIXME`/`ponytail:`/безумовних скіпів; 1 умовний `t.Skipf` (гард ffmpeg, не борг) | `grep -rn 't\.Skip' backend` |
| Трекер issues | `gh issue list` порожньо, репо `b2vv/courses` | попередній прогін |
| Брудне дерево після гейта | Гейт переписує два відслідковувані скриншоти `docs/foundation/screenshots/` — B-18; не комітять | `work/backlog.md` §B-18 |

## 5. Залежності

Перечитано проти `frontend/package.json` і `backend/go.mod` — **жодна версія не змінилась** з
попередньої редакції (`git log ba5279a..HEAD` по обох файлах — порожньо). 7b нових залежностей не
додав: використав наявні pgx, Orval, TanStack Query (`useInfiniteQuery`), Playwright.

| Що | Версія | Чому важлива |
|---|---|---|
| Go | 1.27.1 | `backend/go.mod:3` |
| hls.js | 1.7.2 | ABR увімкнено, `startLevel` не задано |
| React | 19.2.7 | |
| TanStack Router / Query | 1.170.32 / 5.102.8 | Query-кеш request-scoped; `useInfiniteQuery` у панелі discussion |
| React Aria Components | 1.21.0 | джерело єдиного лишеного axe `incomplete` |
| Rsbuild / Rstest | 2.2.2 / 0.11.11 | не Vite, не Vitest |
| Playwright | 1.62.1 | `workers: 1`, `fullyParallel: false` |
| axe-core | 4.13.0 | |
| MSW | 2.15.0 | |
| faster-whisper (Python, у `asr-worker`-образі) | — | рушій ASR, підпроцес; заборона Python-рантайму стосується foundation slice (§24.19) |
| Silero VAD | — | замінив зламаний `no_speech_prob`-поріг |

Заборонене явно: Vite, Next, Vitest, Jest, ESLint, Axios, Redux Toolkit, Radix, shadcn, Rust,
ORM, service locator.

## 6. Ризики

Ризики цього прогону — межі робочого місця інструктора (`docs/instructor/report.md` §5), перенесені
в продукт, не «закриті»:
- **Блокування авторів по-курсно — відкладено до enrollment** (§24 п.8): без enrollment немає
  per-course суб'єкта блокування. Модерація фази — на рівні коментарів, не авторів.
- **Огляд навчання (студенти з прогресом) — поза фазою**: без enrollment «студент курсу»
  напівозначений; ядро фази — модерація.
- **Прибрання лишається безповоротним**; фаза додала лише слід (`removed_by` + журнал для actor≠author),
  «оживлення» немає, аудит мінімальний (reason лише на dismiss).
- **Per-user ліміту репортів/публікацій немає** — лише nginx `zone=api` за IP; ідемпотентність
  репорту за UNIQUE.
- **Урок процесу:** незалежна звірка доказів W01–W26 знайшла 8 рядків без тесту / лише з доменним
  правилом попри зелений агрегат — закрито gap-тестами. «Зелений гейт ≠ доказ рядка» справдилось знову.

Ризики, **знято раніше:** `B-20` (контракт чернетки поза OpenAPI) — закрито `e8c238d`. **Модель оплати**
перестала бути відкритою розвилкою (ухвалена, §24.22). Нижче — раніші відкриті ризики.

1. **Нативна гілка HLS недоведена (D-9).** Потрібен справжній Safari; жодна фаза не закрила
   (`docs/playback/report.md` §6, `docs/asr/report.md` §6).
2. **Стеля 5000 реплік не заміряна під навантаженням** (`docs/transcript/report.md` §6 п.2).
3. **Другий мовний контент не наповнюється** — ключ їде, вміст одна мова.
4. **GATE 2 плану `7a` не зійшовся** — 9 знахідок третього раунду перенесені, не закриті.
5. **Борг «дві паралельні 9-кейсові таблиці істинності».** `course`/`playback` домен-юніти ганяють
   `internal/testsupport/coursevisibility` `Cases()`, а `contracts/access` + адаптери + новий
   `discussion` (K04) — `accesstest.TruthTable`. Дрейф ловиться (усі проти того самого `CanView`,
   R04 = одна реалізація), але наратив «одна таблиця» цього не покриває; `transcript`-домен
   табличного тесту `CanBeViewedBy` не має (`transcript_test.go:182` — одиничний assert). Прибирання
   дубля — окрема задача (`docs/readiness/report.md` §7).
6. **Межі Phase 7b (перенесені в продукт, не закриті)** — `docs/discussion/report.md` §5:
   прибирання коментаря **безповоротне й без `removed_by`** (сказано замовнику); **per-user ліміту
   публікацій немає** — лише nginx `zone=api` 10 r/s за IP (`nginx.conf:21,117`), кілька акаунтів з
   NAT проходять; **пагінації відповідей немає** — корінь віддається повністю під cap 200 (≈400K
   символів у найгіршому `GET`); `display_name` видно всім глядачам курсу. Усе — робоче місце
   інструктора.
7. **`B-22` (беклог, 2026-09-29, нове).** Копії селектора «поточне відео» поза `discussion`:
   гейт-тест K09a пінить лише копію в `discussion`; решта тримаються коментарями. Кандидат на злиття
   в один шов (`work/backlog.md` §B-22).
8. **`B-23` (беклог, 2026-09-29, відкрите — моніторинг).** e2e `B39` (`playback.spec.ts`) раз упав
   у повному гейті; механізм не знайдено. Після цього — **шість чистих точок поспіль** у повному
   гейті (D05, D06 та раніші), але за таким N падіння не атрибутують. Спосіб перевірки при повторі —
   лише A/B **повними** гейтами (ізоляція не розрізняє гіпотез); окремий e2e пересівати перед кожним
   прогоном (`work/backlog.md` §B-23, пам'ять `e2e-isolated-rerun-needs-reseed`).
9. **Межі ASR** (`docs/asr/report.md` §7): integration C14 бере 500 мс дедлайн, не справжній
   `TaskTimeout`; автопостановку не видно без reload; C01/C07 доведено частинами.
10. **`B-18` (беклог).** Упорядкований гейт лишає дерево брудним на відслідковуваних скриншотах
    навіть без змін (`work/backlog.md` §B-18).
11. **`B-19` (беклог).** Ранг 360p упирається в `maxrate 365k` — продуктове питання драбини
    бітрейтів, не дефект (`work/backlog.md` §B-19).
12. **`B-21` (беклог).** Читання транскрипту без мови бере перший за алфавітом
    (`repository.go:40-44`); єдиний відомий шлях до другого закрито в A13, ризик латентний.
13. **Успадковані Phase 6 обмеження:** B25 виміряний над порожнім `<video>`; R35 лише крок ABR на
    синтетичному каналі; каденція володіння 2/6 проти 30/180; P12 без дротового тесту; одна
    аудіодоріжка без AV1.
14. **`S01`/`S02`** — заблоковані на TLS-термінації (`work/tasks.md:36-37`); **`B-1`** —
    спостережуваність фронту, відкладена на кінець роадмапу.
15. ⚠️ **неперевірено:** чи витримає `limit_req zone=media rate=120r/s burst=240` реальний профіль
    сегментного трафіку при кількох глядачах з одного NAT. Навантажувального заміру в репо немає.

## 7. Що читати далі

`AGENTS.md` — контракт репо. `work/AGENDA.md` — ранжування наступного ходу (перезібране слідом за
цим брифінгом): наступне — **фаза оплати** (остання). **`docs/instructor/{spec,plan,report}.md` —
робоче місце інструктора: W01–W26, розширення discussion модерацією, architecture-межа discussion→course.**
**`docs/discussion/{spec,plan,report}.md` — Phase 7b: гілки коментарів, K01–K23.** `docs/readiness/{spec,plan,report}.md`
— Phase 9: правило доступу як контракт, R01–R12, §5 — шви фази оплати. `docs/adr/013` — одне правило
доступу (перед будь-якою зміною доступу й спекою оплати); `010`/`011`/`012`. `PRODUCT-BRIEF.md`
§24 п.22 — ухвалена модель оплати. `work/backlog.md` — відкриті записи (B-18, B-19, B-21, B-22, B-23).
