# CTO research — брифінг репо `courses`

| | |
|---|---|
| Дата | 2026-09-05 |
| Базовий коміт | `dc4aac4` |
| Всього комітів у репо | 75 (`git rev-list --count HEAD`) |
| Гейт свіжості цього прогону | 17 комітів від `c8d2324` — уся Phase 4a (admin course CRUD, авторизаційний шов, публічна програма). Оновлено архітектурний факт №1 (серверна авторизація тепер ІСНУЄ), §2 (нові seams authoring), §5 (react-hook-form додано). Повне перезбирання відкладене; наступний значущий план — запустити `cto-research` начисто. |

Брифінг — кеш первинних джерел, не істина. Розходиться з джерелом — править джерело.

---

## 1. Вердикт на один екран

**Продукт.** Self-hosted платформа продажу онлайн-курсів (вихідне ТЗ: `work/PRODUCT-BRIEF.md`). Працює: публічний каталог із пошуком, повна сесійна автентифікація — логін, SSR-хедер із користувачем, logout, гарди `/learn` і `/admin` за роллю, три локальні фікстури. Постачання — Docker Compose на одному хості, Nginx на `127.0.0.1`; деплой-таргета немає. Масштаб даних: 2 міграції, каталог ≤50 на сторінку.

**Поточний стан.** Phases 1–3 закриті: `make check` — exit 0 по дванадцяти кроках, 175 фронт-тестів, e2e 20/20 (`docs/identity/report.md` §I06 п.4). Черга задач порожня (`work/tasks.md`). Роадмап Phase 4–8 — `work/PRODUCT-BRIEF.md` §21 з уточненнями §24.

**Живий P0.** Специфікація Phase 4 (admin CRUD курсів/модулів/уроків + publish/archive workflow) через spec-flow. Зафіксовані вимоги до спеки з першого дня: `instructor_id` у першій міграції, handler-level авторизація з першим ендпоїнтом, RHF/TanStack Table в'їжджають зі споживачами (`work/PRODUCT-BRIEF.md` §24, `docs/identity/report.md` §I06 п.9).

**Архітектурний факт №1 — серверна авторизація тепер існує (Phase 4a), клієнтські гарди лишаються лише навігаційними.**
Станом до 4a серверної перевірки ролі не було ніде; 4a це виправила. Шов: `httpserver.Actor` (`backend/platform/httpserver/actor.go`), експортовані `Authenticate` і `RequireSameOrigin` з identity-хендлера, і `RegisterAdmin` у `internal/course/adapters/http/admin_handler.go` перевіряє роль+власність на кожній admin-операції (матриця відмов пінована контрактними тестами). `Role.CanAdministerAccounts()` (admin-only) — точка розширення, яку 4b переюзує. Клієнтські гарди (`ssr:false` на `/learn`,`/admin`) — досі лише навігаційна зручність, не механізм; сервер вирішує (`docs/admin/report.md` §6).

**Архітектурний факт №2 — SSR ходить в API повз Nginx.** Сторінкові маршрути рендеряться в SSR-контейнері, який б'є в API напряму: `API_BASE_URL: http://api:8080` (`docker-compose.yml:102`), `apiOrigin()` на сервері віддає саме це (`frontend/src/adapters/api/api-origin.ts:4`). Наслідки вже коштували двох дефектів (неметрований SSR-трафік — аудит MEDIUM-1; логін/логаут фізично неможливі з SSR, бо серверний виклик не несе `Origin`, який `login`/`logout` вимагають рівно один). Правило: для будь-якої edge-політики окремо питати «а SSR цей шлях перетинає?».

---

## 2. Продукт

| Питання | Відповідь | Джерело |
|---|---|---|
| Що це | Self-hosted платформа продажу курсів із власним відеосховищем, HLS, транскриптами | `work/PRODUCT-BRIEF.md` §1 |
| Для кого | Відвідувач (каталог); student / instructor / admin — автентифіковані актори | `docs/architecture/context.md:3`, `backend/internal/identity/domain/user.go` |
| Роадмап | Phase 4 адмінка → 5 upload → 6 playback+progress → 7 транскрипт(+гілки бесід) → 8 замовлення; робоче місце інструктора — окрема пізня фаза | `work/PRODUCT-BRIEF.md` §21, §24 |
| Формат постачання | Docker Compose; хост-порт публікує тільки Nginx на `127.0.0.1` | `docker-compose.yml:114` |
| Що працює зараз | Каталог + сторінка курсу, uk/en; логін → returnTo → хедер із ім'ям → logout; `/learn`, `/admin` — чесні порожні shell-и за роллю; `make seed-identities` — три фікстури | `docs/identity/report.md` §I06 п.1, `README.md:17-23` |
| Чого немає | Реєстрація, скидання пароля, MFA, OAuth, admin CRUD, enrollment, playback, upload, прогрес, коментарі, оплата | `docs/identity/spec.md` §Out of scope, `AGENTS.md:6` |
| Масштаб даних | 2 міграції; каталог limit ≤50, offset ≤10000 | `backend/migrations/`, `backend/openapi/openapi.yaml` §Pagination |
| Мова | Контент український; i18next локалізує системний UI (uk/en) | `docs/foundation/report.md` |

---

## 3. Кодова база й seams

**Розкладка.** Один Go-модуль `courses` із `cmd/{api,media-worker,migrate}`; бізнес-модулі `internal/{course,identity}`, кожен `domain/application/ports/adapters`; інфраструктура `platform/{architecture,config,httpserver,observability,postgres,redis,runtime}`. Фронтенд: `src/{domain,application,ports,adapters,presentation,app,routes}` — identity-зріз присутній на всіх шарах (`frontend/src/domain/identity`, `application/identity`, `app/composition/identity*.ts`, `presentation/identity`).

**Межі та що їх перетинає:**

| Seam | Що перетинає | Джерело |
|---|---|---|
| Браузер ↔ Nginx | HTTP; Nginx — єдиний публічний вхід | `infrastructure/nginx/nginx.conf:37-84` |
| Nginx ↔ API | тільки `/api/*`; edge чистить `traceparent`/`tracestate`; логін має власну строгу зону | `nginx.conf:29,67-68` |
| **SSR ↔ API** | **прямий HTTP по внутрішній мережі, повз Nginx**, з явним форвардингом куки | `docker-compose.yml:102`, `frontend/src/adapters/api/request-headers.ts` |
| Корінь роутера ↔ identity | `beforeLoad: resolveRootIdentity` — один SSR-резолв на запит, дегідрація в кеш `['session']`; куки немає → `null` без мережі; `timeoutMs: 1000` на SSR-гілці | `frontend/src/routes/__root.tsx:25`, `app/composition/root-identity.ts` |
| Захищені маршрути ↔ сесія | один спільний `requireIdentity` (`fetchQuery` + роль через `notFound()`); маршрути `ssr: false`; layout підписаний на `['session']` і редіректить, коли сесія зникає | `frontend/src/routes/learn.tsx:11-18`, `app/composition/require-identity.ts` |
| API ↔ worker | тільки черга Redis/Asynq; спільного доменного коду немає | `backend/cmd/media-worker/main.go` |
| domain ↔ решта (Go) | домен імпортує лише stdlib; application — лише domain/ports; перевіряється тестом | `backend/platform/architecture/boundaries_test.go` |
| domain ↔ решта (TS) | файли `domain/` без імпортів; `application/`, `ports/` — лише inward; `fetch` у `routes/`/`presentation/` заборонений статично | `frontend/scripts/check-architecture.mjs:44-63` |
| Generated ↔ домен | sqlc/Orval живуть в адаптерах, мапляться явно; drift — червоний гейт | `Makefile:46,49` |
| Транспорт ↔ домен (фронт) | один `createApiTransport` → transport outcome; кожен домен мапить у власний union помилок | `frontend/src/adapters/api/transport.ts` |
| Фікстури ↔ БД | чистий SQL-сід `ON CONFLICT` reconcile-до-файлу, окрема ціль `make seed-identities`; Go його не виконує | `backend/sql/seed-identities.sql`, `Makefile:24` |
| Nginx ↔ медіа | Nginx віддає байти; Go і SSR не проксюють; зараз `/media/*` → 403 | `AGENTS.md:19`, `nginx.conf:44-45` |

**Request-scope.** QueryClient та i18next — на кожен SSR-запит (`frontend/src/router.tsx:11-14`); глобальних мутабельних клієнтів немає. Хедер — prop-driven, з `refetchOnMount: false` на identity-запиті — інакше кожен маунт бив у `/me` і G07 був червоний (`work/report-i06d.md`).

---

## 4. Інфра й доставка

| Питання | Відповідь | Джерело |
|---|---|---|
| CI | Один workflow `ci.yml`, тригер `[push, pull_request]`, timeout 30 хв | `.github/workflows/ci.yml:2,7` |
| Порядок гейтів | `check`: format-check → lint → typecheck → test → integration → check-generated → build → up → seed → seed-identities → smoke → e2e; CI повторює той самий порядок | `Makefile:54`, `.github/workflows/ci.yml:14-39` |
| Виміряний стан | Повний `make check` exit 0; 175 фронт-тестів / усі Go-пакети / e2e 20/20 | `docs/identity/report.md` §I06 п.4 |
| Деплой | Кроку деплою немає — CI лише перевіряє | `.github/workflows/ci.yml` (відсутність job) |
| Локальний запуск | `make up` (копіює `.env`, білдить, мігрує, чекає health, релоадить Nginx), `make seed`, `make seed-identities` | `Makefile:8-12,22-25`, `README.md:11-15` |
| Go-тулчейн | Пінований контейнер `golang:1.27.1-bookworm`; Go-тест-контейнери монтують **лише `backend/`** — репо-рутові перевірки живуть в `integration.sh` | `Makefile:2-3`, `work/report-i06f.md` |
| Node | 24 для нативних фронт-інструментів | `.github/workflows/ci.yml:11` |
| Rate limits | `api` 10r/s burst 20 · `web` 50r/s burst 100 · `login` 10r/m burst 5, ключ `$binary_remote_addr` | `nginx.conf:21-29` |
| Моніторинг | Опційний профіль Prometheus; метрики внутрішні | `docker-compose.yml` §prometheus |

**Пастки вимірювання, всі зафіксовані на практиці:**

- `make e2e` двічі поспіль → 429: e2e логіниться з хоста в зону `login`; другий прогін застає гарячий bucket. З чистого стека проходить; `docker compose restart nginx` скидає лімітер. Справжній фікс — спільний `storageState`, тригер: коли e2e виросте (`work/report-i06g.md`).
- `smoke` (30 логінів) перед `e2e` — проходить: smoke б'є з docker-мережі, Playwright — з хоста; адреси різні, buckets різні (`docs/identity/report.md` §I06 п.6).
- Правка `frontend/server.mjs` вимагає `docker compose up -d --build frontend` — інакше e2e міряє старий контейнер (`work/report-i06g.md`).
- `nginx.conf` — bind-mount; правки невидимі без релоаду; `make up` тепер релоадить (`Makefile:12`).
- `/healthz` віддає сам Nginx — марний як сигнал готовності фронтенду (`nginx.conf:42`).
- `check` свідомо білдить двічі (`build`, потім `up`) — наслідок порядку маніфесту (`docs/identity/report.md` §I06 п.6).

---

## 5. Залежності

Дельта `eada97b..HEAD` не чіпала `go.mod` і `package.json` — розділ звірено і збережено.

**Backend** (`backend/go.mod`): Go 1.27.1 · pgx/v5 5.10.0 · goose/v3 3.28.0 · asynq 0.26.0 · go-redis/v9 9.22.0 · kin-openapi 0.149.0 · prometheus client 1.24.1 · OTel 1.46.0 · `golang.org/x/crypto` 0.55.0 (Argon2id) · testcontainers-go 0.44.0.

**Frontend** (`frontend/package.json`): React 19.2.7 · TanStack Router 1.170.32 / Query 5.102.8 / Start 1.168.49 · react-aria-components 1.21.0 · i18next 26.4.1 · zod 4.5.4 · tailwindcss 4.3.3 · srvx 1.0.1. Інструменти: Rsbuild 2.2.2, Rstest 0.11.11, Playwright 1.62.1, MSW 2.15.0, Orval 8.27.0, oxlint 1.81.0 / oxfmt 0.66.0, TypeScript 6.0.3.

**Що на чому заблоковане:**

| Блокування | Джерело |
|---|---|
| Orval-клієнт ← `backend/openapi/openapi.yaml`; зміна контракту зобов'язує `make generate`, інакше drift-гейт | `frontend/orval.config.ts:5`, `Makefile:46,49` |
| sqlc ← `backend/sql/queries`, живе під адаптерами | `Makefile:46` |
| Заборонені: Rust, ORM, service locator, Vite, Next, Vitest, Jest, ESLint, Axios, Redux Toolkit, Radix, shadcn | `AGENTS.md:12,24` |
| Зі стек-списку ТЗ ще не встановлені (чекають споживача): RHF і TanStack Table/Virtual → Phase 4, Zustand → 5–6, hls.js → 6, TipTap → 7, ECharts → прогрес/аналітика | `work/PRODUCT-BRIEF.md` §3, `frontend/package.json` |

---

## 6. Ризики

**Відкриті знахідки аудиту** (`docs/security/audit-2026-09-03.md`; 0 critical/high; повторне рев'ю гілки I06 — 0 HIGH/MEDIUM, один sub-threshold logout-фікс уже змерджений `1292b3a`):

| ID | Суть | Стан |
|---|---|---|
| LOW-5 | Немає CSP, Permissions-Policy, HSTS | `S01`, за TLS (`work/tasks.md` §Pre-deploy) |
| LOW-6 | Куки без `Secure`; `__Host-` неможливий на http | `S01`→`S02`, за TLS (`work/tasks.md:37`) |

**Свідомий борг із власниками й тригерами:**

| Що | Тригер повернення | Джерело |
|---|---|---|
| `presentation/ui/data.tsx` (`Table`, `Pagination`) хардкодить українську | перший рендер цих компонентів англійською — імовірно Phase 4 | `work/report-i06d.md` |
| `rstest.config.ts` глобально аліасить `@tanstack/react-start/server$` | тест, якому потрібне щось поза `getCookie` | `frontend/rstest.config.ts:10` (`ponytail:`) |
| `RootIdentity.unavailable` продукується, але жоден споживач не відрізняє його від `anonymous` | еволюція хедера/гардів — дати сенс або згорнути | `work/report-i06a.md` |
| Надлишкова рання перевірка в `safeReturnTo` — свідомий пояс на trust boundary | — | `work/report-i06c.md` |
| Worker стартує, задач не обробляє | Phase 5 | `README.md:59`, `backend/cmd/media-worker/main.go` |
| Індетермінований commit → 503 без куки, свідомо | — | `docs/identity/spec.md` §Security contract |

**Де найкрихкіше:**

| Місце | Чому | Джерело |
|---|---|---|
| Серверна авторизація | Не існує; Phase 4 будує її з нуля і не сміє успадкувати з клієнтських гардів | `docs/identity/report.md` §I06 п.7 |
| Розріз задач | I06: чотири з восьми задач переносили роботу через свої межі; різати треба по виконуваних доказах і межах середовищ, не по файлових/UI-ярликах | `docs/identity/report.md` §I06 п.8 |
| Точність часу application↔сховище | `time.Now` (нс) не переживав round-trip `timestamptz`; логін був зламаний при зелених тестах | `docs/identity/report.md` I03 п.6 |
| MSW | Не віддає заголовки запиту хендлеру — форвардинг куки доводиться на білдері заголовків і наживо | `docs/identity/report.md` I05 п.8 |
| Вартість логіну | Argon2id `m=64MiB,t=3,p=2`; `VerifyDummy` платить ту саму ціну — тому окрема жорстка зона `login` | `backend/internal/identity/adapters/argon2id/verifier.go:14-16`, `nginx.conf:29` |
| Контракт URL сторінки | `validateSearch`, що емітить дефолт → роутер переписує адресу → 307 (`load-server.js:530`); дефект був у `/login` і `/courses` | `docs/identity/report.md` §Catalog URL contract |
| Статуси SSR | Лейн маршруту віддає захардкоджений 500 на будь-який throw; досяжні лише 200 / 404 (`notFound()`) / 3xx (`redirect()`) / 500 | `docs/identity/report.md` §Catalog URL contract |
| Перед публічним bind | TLS, керовані секрети, бекапи, квоти сховища | `README.md:61` |

**⚠️ Неперевірені тези: 0.**
