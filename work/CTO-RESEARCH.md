# CTO research — брифінг репо `courses`

| | |
|---|---|
| Дата | 2026-09-07 |
| Базовий коміт | `1eb1390` |
| Всього комітів у репо | 126 (`git rev-list --count HEAD`) |
| Гейт свіжості цього прогону | **34 коміти** від `0c46f10`, 128 файлів, +19442 рядки — уся Phase 5 (медіа) плюс закритий беклог-пункт B02. Це не оновлення шапки: брифінг перезібрано **начисто** по всіх шести розділах. Попередня редакція описувала репо без жодного рядка медіа. |
| Тез із міткою `⚠️ неперевірено` | **2** (позначені в §4 і §6) |

Брифінг — кеш первинних джерел, не істина. Розходиться з джерелом — править джерело,
а кеш переписується під нього.

---

## 1. Вердикт на один екран

**Продукт.** Self-hosted платформа продажу онлайн-курсів (вихідне ТЗ —
`work/PRODUCT-BRIEF.md`). Працює: публічний каталог із пошуком, повна сесійна
автентифікація, адміністрування курсів/модулів/уроків із publish/archive, адміністрування
акаунтів, **і завантаження відео на урок із транскодуванням у HLS-сходинки**. Постачання —
Docker Compose на одному хості, Nginx на `127.0.0.1`; деплой-таргета в репозиторії немає
(`docker-compose.yml`, `docs/media/report.md` §Post-deploy).

**Поточний стан.** Phases 1–5 закриті. Останній повний ордер-гейт: `make check` exit 0 —
334 фронт-тести у двох rstest-проєктах, 38 Go-пакетів, e2e 35/35
(`docs/media/report.md` §2, перезаміряно на `1eb1390`). Черга задач порожня
(`work/tasks.md`). Роадмап Phase 6–8 — `work/PRODUCT-BRIEF.md` §21 з уточненнями §24
(п.1–17). Беклог: `work/backlog.md` — B-1 (обсервабельність) відкритий, B-2 закритий
2026-09-07.

**Живий P0.** Специфікація **Phase 6** (плеєр + підписаний доступ до медіа) через
`spec-flow`. Два рішення для неї вже ухвалені й записані: §24.16 — друга мова приходить
окремим аудіо-вивантаженням із мовним тегом, перша доріжка витягується з вихідного відео;
§24.17 — Phase 6 ставить пробу можливостей AV1-декодера, щоб умова (2) §24.14 стала
вимірюваною.

### Архітектурний факт №1 — медіа-байти вже існують, але **шва авторизації до них немає взагалі**, і найочевидніше рішення заборонене

Це те, що найлегше зрозуміти неправильно: «Phase 5 зробила відео, отже плеєр — це майже
готове». Насправді Phase 5 виробляє байти і **не дає нікому права їх читати**.

| Що саме | Джерело |
|---|---|
| `/media` і `/media/` віддають **403 усім, беззастережно** — це не помилка конфігурації, а рішення фази | `infrastructure/nginx/nginx.conf:45-46` |
| Nginx монтує том із медіа **тільки на читання** і є єдиним, хто взагалі бачить байти | `docker-compose.yml:141` (`media-data:/media:ro`) |
| Go ніколи не віддає медіа-байти, і це правило, а не збіг | `AGENTS.md` §Architecture: «Nginx serves media bytes; Go and SSR never proxy them» |
| Отже Phase 6 винаходить шов авторизації **на edge** (підписаний URL / `X-Accel-Redirect` / `auth_request`), а не ендпоїнт у Go, що стрімить файли | наслідок трьох рядків вище |

Тобто робота Phase 6 — не «додати плеєр до готового бекенда», а **спроєктувати відсутній
механізм доступу** так, щоб він не порушив правило «байти не течуть крізь Go».
`docs/media/report.md` §6 називає це першим обмеженням фази: «No playback… no signed access».

### Архітектурний факт №2 — SSR ходить в API повз Nginx

Сторінкові маршрути рендеряться в SSR-контейнері, який б'є в API напряму:
`API_BASE_URL: http://api:8080` (`docker-compose.yml:126`), і серверна гілка `apiOrigin()`
віддає саме це (`frontend/src/adapters/api/api-origin.ts:4`). Це вже коштувало двох
дефектів (неметрований SSR-трафік; логін/логаут неможливі з SSR, бо серверний виклик не
несе `Origin`). **Правило, яке прямо стосується Phase 6:** для будь-якої edge-політики
окремо питати «а SSR цей шлях перетинає?» — підписаний URL, який видає edge, SSR не побачить.

### Архітектурний факт №3 — виконуваних файлів чотири, і байти не проходять крізь Go API

`backend/cmd/`: `api`, `media-worker`, `upload`, `migrate`. Чанки завантаження йдуть у
**окремий сервіс `upload:8082`**, куди Nginx маршрутизує `/uploads/` напряму
(`infrastructure/nginx/nginx.conf:49-63`, `docker-compose.yml:99`), і Go API їх не бачить.
Композиція API навмисно **не має адаптера сховища** (`docs/media/plan.md`, шапка задач).

---

## 2. Продукт

| Питання | Відповідь | Джерело |
|---|---|---|
| Що це | Платформа продажу й доставки онлайн-курсів, self-hosted | `work/PRODUCT-BRIEF.md` |
| Для кого | Студент (каталог, навчання), інструктор (авторство — окрема фаза після Phase 6), адмін | `work/PRODUCT-BRIEF.md` §24.3 |
| Що працює зараз | каталог із пошуком; логін/сесія/logout; адмін CRUD курсів, модулів, уроків; publish/archive; адміністрування акаунтів; вивантаження відео на урок і транскодування в HLS | `docs/admin/report.md`, `docs/identity/report.md`, `docs/media/report.md` §1 |
| Чого **немає** | програвання відео, підписаний доступ, прогрес студента, оплати, коментарі, транскрипт, робоче місце інструктора | `AGENTS.md` §Scope; `docs/media/report.md` §6 |
| Формат постачання | Docker Compose, один хост, Nginx на `127.0.0.1`; TLS немає | `docker-compose.yml`, `work/tasks.md` §Pre-deploy hardening (S01/S02 чекають TLS) |
| Масштаб даних | 5 міграцій; каталог ≤50 на сторінку; стеля відео 2 GiB, чанк 8 MiB | `backend/migrations/`, `backend/internal/media/domain/asset.go:12`, `frontend/src/domain/media/video.ts:72` |
| Роадмап | Phase 6 плеєр+прогрес, 7 транскрипт+бесіди, 8 замовлення й оплата | `work/PRODUCT-BRIEF.md` §21 + §24 |

**Рішення, які вже ухвалені за Phase 6 і які не переухвалюють:** §24.14 (AV1 відкладено за
трьома умовами), §24.15+§24.16 (мультимовне аудіо — окремі аудіо-вивантаження з мовним
тегом), §24.17 (проба можливостей AV1 будується в Phase 6), §24.3 (робоче місце інструктора
— окрема фаза **після** Phase 6).

## 3. Кодова база й seams

**Форма.** Модульний моноліт на Go + окремі виконувані файли; фронт — TanStack Start/SSR.
Домен без фреймворків; адаптери дивляться всередину; згенеровані sqlc/Orval моделі живуть в
адаптерах (`AGENTS.md` §Architecture; напрямок залежностей перевіряється машинно —
`backend/platform/architecture/boundaries_test.go`, `frontend/scripts/check-architecture.mjs`).

**Backend-модулі:** `course`, `identity`, `media` — кожен із `{domain,application,ports,adapters}`
(`backend/internal/`). Спільна інфраструктура — `backend/platform/{architecture,config,httpserver,observability,postgres,redis,runtime}`.

**Медіа-модуль (новий у Phase 5):**

| Шар | Вміст | Джерело |
|---|---|---|
| domain | `VideoAsset`, шість станів `created→uploading→queued→processing→ready\|failed`, `ClaimToken`, `Generation`, п'ять `FailureReason`, `StorageKey{Kind,ID,Leaf}` із чотирма видами (`source`, `hls`, `thumbnails`, `temporary`) | `backend/internal/media/domain/asset.go:16-39,150-161` |
| ports | `queue`, `repository`, `storage`, `transcoder` — чотири | `backend/internal/media/ports/` |
| adapters | `asynq` (черга), `ffmpeg` (транскодер), `filesystem` (сховище на `os.Root`), `http` (API-поверхня), `postgres`, `tus` (протокол-сервер) | `backend/internal/media/adapters/` |

**Ключові seams і що їх перетинає:**

| Межа | Що перетинає | Джерело |
|---|---|---|
| Браузер → `upload:8082` | чанки tus, повз Go API і повз SSR | `infrastructure/nginx/nginx.conf:49-63` |
| `upload` → Postgres | завершення вивантаження **в одній транзакції** з постановкою задачі | `docs/media/report.md` §1, plan §20/§21 |
| API → Redis/asynq | продюсер задач транскодування | `backend/internal/media/adapters/asynq/` |
| `media-worker` → ffmpeg | рендер сходинок; **воркер одноконкурентний** | `backend/cmd/media-worker/main.go:70` (`Concurrency: 1`) |
| worker → сховище | атомарна публікація з витісненням старої генерації | `backend/internal/media/adapters/filesystem/storage.go:230` (`Publish`) |
| Nginx → том медіа | **тільки читання, і зараз 403 для всіх** — шов Phase 6 | `docker-compose.yml:141`, `nginx.conf:45-46` |
| SSR → API | напряму, повз Nginx (факт №2) | `docker-compose.yml:126` |

**Frontend.** `frontend/src/{domain,application,ports,adapters,presentation,app,routes}`.
Адаптер один — `adapters/api`. П'ять портів, серед них `LessonVideoRepository.ts`.
tus-клієнт написаний вручну (`adapters/api/tus-client.ts`); згенерований Orval-клієнт —
`adapters/api/generated/` (руками не правиться, дрейф ловить `make check-generated`).

**Тестові середовища фронту розділені (B02, 2026-09-07):** `frontend/rstest.config.ts`
декларує два проєкти — `adapter-node` (`node`, `src/adapters/api/**`) і `dom` (`happy-dom`,
решта). Причина: happy-dom застосовує CORS-safelist до **response**-заголовків, тому
протокол, що живе в заголовках, не можна було довести в адаптерному шарі. Тепер можна
(`work/backlog.md` B-2 §Outcome, `docs/media/plan.md` §26).

## 4. Інфра й доставка

| Питання | Відповідь | Джерело |
|---|---|---|
| Сервіси Compose | `postgres`, `redis`, `migrate`, `api`, `media-worker`, `upload`, `frontend`, `nginx`, `prometheus` (профіль) | `docker-compose.yml` |
| Томи | `postgres-data`, `redis-data`, `media-data`, `prometheus-data` | `docker-compose.yml:164-167` |
| Порядок гейтів | `format-check → lint → typecheck → test → integration → check-generated → build → up → seed-identities → seed → smoke → e2e` | `Makefile:54`, `.claude/standards.md` |
| CI | той самий порядок у `.github/workflows/ci.yml`; артефакт — playwright-траси | `.github/workflows/ci.yml:12-39` |
| Зелене зараз | увесь ланцюг, exit 0 на `1eb1390` | перезаміряно цим прогоном |
| Rate limits на edge | зони `api` 10r/s, `web` 50r/s, `login` 10r/m, `upload` 30r/s | `infrastructure/nginx/nginx.conf:21-30` |
| Тіло запиту | `/uploads/` — `client_max_body_size 16m`, буферизація запиту вимкнена | `nginx.conf:50-52` |
| Деплой | **немає таргета в репозиторії** | `docs/media/report.md` §Post-deploy |

**Пастки прогону, зафіксовані емпірично:** `make e2e` двічі поспіль ловить 429 — скидається
`docker compose restart nginx` плюс пересів ідентичностей; медіа-подорож живе в окремому
Playwright-проєкті, бо рестартує воркер і насичує два CPU. ⚠️ **неперевірено цим прогоном:**
чи 429 відтворюється й після Phase 5 — цього разу e2e біг один раз (джерело твердження —
хендоф Phase 5, не власний замір).

## 5. Залежності

**Go** (`backend/go.mod`, Go 1.27.1): `pgx/v5` 5.10.0, `goose/v3` 3.28.0, `asynq` 0.26.0,
`go-redis/v9` 9.22.0, `kin-openapi` 0.149.0, `prometheus/client_golang` 1.24.1,
OpenTelemetry 1.46.0, `testcontainers-go` 0.44.0, `golang.org/x/crypto` 0.55.0.

**Frontend** (`frontend/package.json`, Node ≥24 <25): React 19.2.7, TanStack Start 1.168.49
/ Router 1.170.32 / Query 5.102.8, `react-aria-components` 1.21.0, Rsbuild 2.2.2,
`@rstest/core` 0.11.11, `msw` 2.15.0, `happy-dom` 20.13.2, `orval` 8.27.0, `zod` 4.5.4,
`tailwindcss` 4.3.3, `react-hook-form` ^7.87.0, Playwright 1.62.1, `oxlint` 1.81.0 /
`oxfmt` 0.66.0.

**Заблоковане на чому:**

| Що | Чекає на | Джерело |
|---|---|---|
| S01 (CSP, HSTS, `Secure`), S02 (`__Host-` кука) | TLS-термінацію, якої немає | `work/tasks.md` §Pre-deploy hardening |
| AV1 як додаткова сходинка | три умови §24.14; умову (2) робить вимірюваною §24.17 | `work/PRODUCT-BRIEF.md` §24.14, §24.17 |
| Друга мова аудіо | Phase 6+ разом із плеєром; спосіб доставки вже вирішено | `work/PRODUCT-BRIEF.md` §24.16 |
| Другий воркер / `Concurrency > 1` | справжній lease-heartbeat | `docs/media/report.md` §6, D-7 |

**Версія ffmpeg — не деталь, а вимога до доказів.** Воркер везе **ffmpeg 5.1.9**, хост —
7.1.1, і поведінка мукcера між ними різна (у 5.1.9 майстер-плейлист не має
`AVERAGE-BANDWIDTH`). Будь-яке твердження про плейлисти мусить називати бінар; `B33` друкує
версію воркера в кожен smoke-прогін (`docs/media/report.md` §3).

## 6. Ризики

| Ризик | Стан | Джерело |
|---|---|---|
| **Програвання не існує** — `/media/*` 403 для всіх, нема плеєра, підписаного доступу, прогресу | за задумом; це і є Phase 6 | `docs/media/report.md` §6, `nginx.conf:45-46` |
| **Обсервабельності немає** — фронтового логера нема взагалі, логи живуть як stdout контейнерів і гинуть при рестарті, Prometheus пише серії, які ніхто не читає | відкритий беклог **B-1**, не заплановано | `work/backlog.md` B-1 |
| **Воркер одноконкурентний, і pass відновлення це припускає** — другий репліка або `Concurrency > 1` потребує lease-heartbeat | тригер записано в D-7 | `backend/cmd/media-worker/main.go:70`, `docs/media/report.md` §6 |
| **TLS немає**, тому дві безпекові задачі стоять | S01, S02 `planned` | `work/tasks.md` |
| **Одна аудіо-доріжка** | спосіб доставки другої вирішено (§24.16), реалізація — Phase 6+ | `work/PRODUCT-BRIEF.md` §24.16 |
| **Метрики не покидають docker-мережу** без TLS і авторизації | навмисно | `work/backlog.md` B-1 §Constraints |

**Свідомий борг у коді практично відсутній:** `TODO`/`FIXME`/`t.Skip(`/`.skip(` — **0**
входжень у `backend/`, `frontend/src/`, `frontend/e2e/`; маркерів `ponytail:` — **1**
(`frontend/rstest.config.ts:12`, пояснює спільний rspack-аліас двох тестових проєктів).

**Найкрихкіше місце процесу, а не коду.** Шість раундів GATE 2 у Phase 5 — включно з
рецензентами, що ганяли справжній ffmpeg і живий PostgreSQL — не побачили, що
опублікований протокол вивантаження **не міг прийняти перший байт**, що таблиця портів
суперечила рішенню, яке реалізовувала, і що CHECK-обмеження не обмежувало нічого
(`array_length('{}',1)` — NULL, а NULL CHECK **виконано**). Усі три знайшов виконавець, у
момент, коли щось уперше мусило справді запуститись. Висновок, який варто нести в Phase 6:
гейти читають механізми, але не читають шов між механізмом і підписом, що його обслуговує,
і не питають, чи система здатна стартувати з порожнього стану
(`docs/media/report.md` §3).

⚠️ **неперевірено:** чи testcontainers-гонка ryuk-reaper, яка цього прогону дала два падіння
`make integration` з семи запусків (далі п'ять чистих поспіль), відтворюється поза
контенцією з `docker compose build`. Схоже, що це середовищний флак, а не дефект коду —
зміна B02 суто фронтова й Go-тестів не торкається, — але окремого заміру під навантаженням
не робилось.
