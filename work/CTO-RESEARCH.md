# CTO research — брифінг репо `courses`

| | |
|---|---|
| Дата | 2026-09-04 |
| Базовий коміт | `eada97b` |
| Всього комітів у репо | 19 (`git rev-list --count HEAD`) |
| Гейт свіжості цього прогону | 4 коміти від `fd83598`; два з них чіпають області брифінгу (фронтові API-адаптери, контракт URL каталогу), тож §3 і §6 перезібрано. Перший прогін 2026-09-04 порівнювати не мав із чим. |

Брифінг — кеш первинних джерел, не істина. Розходиться з джерелом — править джерело.

---

## 1. Вердикт на один екран

**Продукт.** Self-hosted платформа онлайн-курсів із наміром продавати курси. Зараз працює публічний каталог (перегляд і пошук опублікованих курсів) та автентифікація сесією. Постачання — Docker Compose на одному хості, тільки loopback; жодного деплой-таргета в репо немає. Масштаб даних мізерний: 2 міграції, каталог із пагінацією до 50 елементів на сторінку (`backend/openapi/openapi.yaml`, `Pagination.limit` max 50). Логіка живе в Go-модульному моноліті; фронтенд має власний домен/застосунок і не приймає бізнес-рішень.

**Поточний стан.** Phase 1 (foundation) і Phase 2 (каталог) закриті. Phase 3 Identity — I01–I05 done, I06 лишається (`work/tasks.md:25-30`). Phases 4–8 (адмінка, upload/транскодинг, підписане відтворення, транскрипти, замовлення/оплата) не починалися (`AGENTS.md:6`).

**Живий P0.** I06: хедер із користувачем і logout, клієнтські гарди `/learn` і `/admin`, демо-фікстури, фінальний ordered gate (`work/tasks.md:30`). Без нього відкриті приймальні сценарії A01, A02, A04, A05, B01 (`docs/identity/plan.md:20-26`).

**Архітектурний факт №1 — SSR ходить в API повз Nginx.**
Природне припущення «весь трафік проходить через edge, отже edge-контролі застосовуються до всього» тут **хибне**. Сторінкові маршрути рендеряться в SSR-контейнері, який б'є в API напряму по внутрішній docker-мережі: `API_BASE_URL: 'http://api:8080'` (`docker-compose.yml:102`), а `apiOrigin()` на сервері повертає саме це значення (`frontend/src/adapters/api/api-origin.ts:4`). Стрілка є на діаграмі (`docs/architecture/containers.md:8`), але наслідки з неї не витягли, і це вже коштувало двох реальних дефектів:

- Rate limit був лише на `location ^~ /api/`, тож SSR-навантаження на каталог і `ILIKE`-пошук було неметрованим — знахідка MEDIUM-1 аудиту (`docs/security/audit-2026-09-03.md`). Закрито в I04 зоною `web` (`infrastructure/nginx/nginx.conf:25,71`).
- Серверний виклик не має заголовка `Origin`, а `login`/`logout` вимагають рівно один збіжний `Origin` (`backend/internal/identity/adapters/http/handler.go:57,92`). Тому **логін і логаут фізично неможливі з SSR** і лишаються браузерними операціями (`docs/identity/report.md`, секція I04 п.7).

Практичне правило для будь-якої наступної задачі: усе, що стосується edge-політик (ліміти, заголовки, авторизація на межі), треба окремо питати «а SSR цей шлях узагалі перетинає?».

---

## 2. Продукт

| Питання | Відповідь | Джерело |
|---|---|---|
| Що це | Self-hosted вітрина курсів і майбутня навчальна платформа | `docs/architecture/context.md:3` |
| Для кого | Відвідувач (публічний каталог); student / instructor / admin як автентифіковані актори з Phase 3 | `docs/architecture/context.md:3`, `docs/identity/spec.md` §Outcome |
| Формат постачання | Docker Compose; хост-порт публікує тільки Nginx, і той на `127.0.0.1` | `docker-compose.yml:114` |
| Що працює зараз | Публічний каталог + сторінка курсу, uk/en, login/me/logout | `README.md:3`, `docs/identity/report.md` I03–I05 |
| Чого немає | Реєстрація, скидання пароля, MFA, OAuth, адмінка, enrollment, курс-CRUD, refresh-токени | `docs/identity/spec.md` §Out of scope |
| Комерція | Платіжного провайдера не підключено; замовлення — Phase 8 | `README.md:3,49` |
| Масштаб даних | 2 міграції, сторінка каталогу ≤50, offset ≤10000 | `backend/migrations/`, `backend/openapi/openapi.yaml` §Pagination |
| Мова контенту | Контент курсів український; i18next локалізує лише системний UI | `docs/foundation/report.md` |

---

## 3. Кодова база й seams

**Розкладка.** Один Go-модуль `courses` (`backend/go.mod:1`) з трьома виконуваними файлами: `cmd/api`, `cmd/media-worker`, `cmd/migrate`. Бізнес-модулі — `internal/course` та `internal/identity`, кожен із `domain / application / ports / adapters`. Інфраструктура — `platform/{architecture,config,httpserver,observability,postgres,redis,runtime}`.

**Межі та що їх перетинає:**

| Seam | Що перетинає | Джерело |
|---|---|---|
| Браузер ↔ Nginx | HTTP; Nginx — єдиний публічний вхід | `infrastructure/nginx/nginx.conf:37-84` |
| Nginx ↔ API | тільки `/api/*`; edge чистить `traceparent`/`tracestate` | `nginx.conf:67-68` |
| **SSR ↔ API** | **прямий HTTP по внутрішній мережі, повз Nginx**, з форвардингом куки запиту | `docker-compose.yml:102`, `docs/architecture/containers.md:8` |
| API ↔ worker | тільки через чергу Redis/Asynq; спільного коду домену немає | `backend/cmd/media-worker/main.go:38,44,52` |
| domain ↔ решта | Go: домен імпортує лише stdlib; application — лише domain/ports. Перевіряється тестом | `backend/platform/architecture/boundaries_test.go:13-45` |
| domain ↔ решта (фронт) | TS: файли в `domain/` не мають **жодного** імпорту; `application/` і `ports/` — лише inward | `frontend/scripts/check-architecture.mjs:44,48` |
| presentation ↔ мережа | `fetch` у `routes/` і `presentation/` заборонений статично | `frontend/scripts/check-architecture.mjs:63` |
| Generated ↔ домен | sqlc/Orval живуть тільки в адаптерах і мапляться явно | `AGENTS.md:16`, `Makefile:49` |
| Транспорт ↔ домен (фронт) | один `createApiTransport` віддає transport outcome; кожен домен мапить його у власний union помилок, вокабуляри не змішуються | `frontend/src/adapters/api/transport.ts` |
| Nginx ↔ медіа | Nginx віддає байти; Go і SSR їх не проксюють. Зараз `/media/*` → 403 | `AGENTS.md:19`, `nginx.conf:44-45` |

**Що варто знати про request-scope.** QueryClient та i18next створюються на кожен SSR-запит (`frontend/src/router.tsx:11-14`); глобальних мутабельних клієнтів немає. Кука форвардиться явно, бо на сервері немає браузерного cookie-jar (`frontend/src/adapters/api/request-headers.ts`), і обидва API-адаптери ходять через спільний транспорт (`frontend/src/adapters/api/transport.ts`).

---

## 4. Інфра й доставка

| Питання | Відповідь | Джерело |
|---|---|---|
| CI | Один workflow `ci.yml`, тригер `[push, pull_request]` | `.github/workflows/ci.yml:2` |
| Порядок гейтів | format → lint → typecheck → unit → integration → generated drift → build → up → seed+smoke → Playwright → shutdown | `.github/workflows/ci.yml:14-39` |
| Деплой | **Кроку деплою в репо немає** — CI лише перевіряє | `.github/workflows/ci.yml` (відсутність job) |
| Локальний запуск | `make up` (копіює `.env`, білдить, мігрує, чекає health, **релоадить Nginx**), `make seed` | `Makefile:8-12,22` |
| Go-тулчейн | Пінований контейнер `golang:1.27.1-bookworm`; Go локально не потрібен | `Makefile:2-3` |
| Node | 24 для нативних фронт-інструментів | `.github/workflows/ci.yml:11` |
| Rate limits на edge | `api` 10r/s burst 20 · `web` 50r/s burst 100 · `login` 10r/m burst 5 | `nginx.conf:21,25,29,49,60,71` |
| Моніторинг | Опційний профіль Prometheus; метрики не виходять за docker-мережу | `docker-compose.yml` §prometheus, `nginx.conf:46-47` |

**Дві пастки вимірювання, зафіксовані на практиці** (`docs/identity/report.md` I04 п.6, I05 п.8):

- `nginx.conf` — bind-mount, і контейнер не перестворюється сам; правки конфігу невидимі, поки Nginx не перезавантажено. `make up` тепер робить `nginx -s reload` (`Makefile:12`).
- `/healthz` віддає **сам Nginx** (`nginx.conf:42`), тому як сигнал готовності фронтенду він марний. Треба `docker compose up -d --wait <service>`.

---

## 5. Залежності

**Backend** (`backend/go.mod`): Go 1.27.1 · pgx/v5 5.10.0 · goose/v3 3.28.0 · asynq 0.26.0 · go-redis/v9 9.22.0 · kin-openapi 0.149.0 (контрактні тести) · prometheus client 1.24.1 · OTel 1.46.0 · `golang.org/x/crypto` 0.55.0 (Argon2id) · testcontainers-go 0.44.0.

**Frontend** (`frontend/package.json`): React 19.2.7 · TanStack Router 1.170.32 / Query 5.102.8 / Start 1.168.49 · react-aria-components 1.21.0 · i18next 26.4.1 · zod 4.5.4 · tailwindcss 4.3.3 · srvx 1.0.1. Інструменти: Rsbuild 2.2.2, Rstest 0.11.11, Playwright 1.62.1, MSW 2.15.0, Orval 8.27.0, oxlint 1.81.0 / oxfmt 0.66.0, TypeScript 6.0.3.

**Що на чому заблоковане:**

| Блокування | Джерело |
|---|---|
| Orval-клієнт генерується з `backend/openapi/openapi.yaml` — зміна контракту зобов'язує `make generate`, інакше падає drift-гейт | `frontend/orval.config.ts:5`, `Makefile:46,49` |
| sqlc-код генерується з `backend/sql/queries` і живе під адаптерами | `Makefile:46` |
| Заборонені явно: Rust, ORM, service locator, Vite, Next, Vitest, Jest, ESLint, Axios, Redux Toolkit, Radix, shadcn | `AGENTS.md:12,24` |
| Залежність додається тільки коли її щось реально споживає | `AGENTS.md:25` |

---

## 6. Ризики

**Відкриті знахідки аудиту** (`docs/security/audit-2026-09-03.md`; 0 critical, 0 high):

| ID | Суть | Стан |
|---|---|---|
| LOW-5 | Немає CSP, Permissions-Policy, COOP/COEP, HSTS | відкладено в задачу `S01`, за TLS (`work/tasks.md` §Pre-deploy hardening) |
| LOW-6 | Locale-кука без `Secure` | те саме, за TLS |

MEDIUM-1, MEDIUM-2, LOW-1..LOW-4 закриті (`docs/security/audit-2026-09-03.md` §Що змінилося).

**Свідомо недороблене:**

| Що | Джерело |
|---|---|
| БД не сідить жодної ідентичності — демо-креденшели належать I06 | `backend/sql/seed.sql` (жодного рядка users), `docs/identity/report.md` I05 п.9 |
| `/learn` і `/admin` — досі Phase 1 плейсхолдери, гардів немає | `frontend/src/routes/learn.tsx`, `docs/identity/report.md` I05 п.9 |
| Немає handler-level авторизації за роллю — чекає на перший захищений admin-ендпоїнт | `docs/identity/spec.md` §Scope |
| Worker стартує, але не обробляє задач транскодингу | `README.md:49`, `backend/cmd/media-worker/main.go` |
| Індетермінований commit у PostgreSQL свідомо віддається як 503 без куки | `docs/identity/spec.md` §Security contract |

**Де найкрихкіше:**

| Місце | Чому | Джерело |
|---|---|---|
| Точність часу на межі application↔сховище | `time.Now` не переживав round-trip через `timestamptz`; логін був повністю зламаний при 100% зелених тестах, бо всі тести брали рівні секунди | `docs/identity/report.md` I03 п.6 |
| Edge-конфіг | Правки `nginx.conf` невидимі без релоаду — один замір уже дав хибний результат | `docs/identity/report.md` I04 п.6 |
| MSW як інструмент | У цій зв'язці **не віддає заголовки запиту** хендлеру, тож форвардинг куки через нього не перевіряється | `docs/identity/report.md` I05 п.8 |
| Вартість логіну | Argon2id `m=64MiB, t=3, p=2`, і `VerifyDummy` платить ту саму ціну за неіснуючі акаунти — тому логін має окрему жорстку зону, а не загальну API | `nginx.conf:29`, `backend/internal/identity/adapters/argon2id/verifier.go:14-16` |
| Контракт URL сторінки | `validateSearch`, що емітить дефолт, змушує роутер переписати адресу й віддати 307: правило `next.publicHref !== canonical.publicHref` у `router-core` `load-server.js:530`. Цей дефект був і в `/login`, і в `/courses` | `docs/identity/report.md` §Catalog URL contract |
| Статуси помилок SSR | лейн маршруту віддає захардкоджений 500 на будь-який throw (`load-server.js:303`); гачка відсутня, `setResponseStatus` не діє. Досяжні лише 200, 404 через `notFound()`, 3xx через `redirect()` і 500 | `docs/identity/report.md` §Catalog URL contract |
| Перед публічним bind | Потрібні TLS, керовані секрети, бекапи, квоти сховища | `README.md:51` |

**⚠️ Неперевірені тези: 0.** Кожна теза вище має посилання на код, документ або файл конфігурації.
