# Observability (B-1) — report

Дельта до [plan.md](plan.md) (ред.round2, GATE 2 ЗЕЛЕНИЙ на раунді 3 — `_wip-gate2-round3.md`).
Гілка `feat/observability-full-stack` (off main `8c77860`). Закриває backlog B-1.

Статус гейта: **ordered `make check` EXIT=0** (frontend 784 unit, e2e 124 passed, 14.7m); profile-on доказ
`infrastructure/tests/monitoring-stack.sh` — **PASS** (поза `make check`).

## 1. Що зроблено

Слайси S1–S7 як у plan.md §План міграції; нижче — лише суть і відхилення.

- **S1** `f0ea2f3` — backend `internal/clientlog` (domain + adapters/http) + endpoint
  `POST /api/v1/client-logs` + release-define plumbing + OpenAPI/Orval. Domain: закритий набір
  рівнів, size caps (truncate), route/release `"unknown"` sentinel, **2-рівнева §17-редакція**
  (закрита схема + `DisallowUnknownFields` = field-рівень; `scrubSecrets` маскує
  password/token/cookie/Bearer/JWT у message/stack/route). Handler emit'ить **напряму через
  `platform/observability` logger** (без EventSink-порту), `source=client`; invalid→400 фіксований
  envelope (без echo), valid→**204**.
- **S2** `d645b39` — frontend `modules/observability` (порт + адаптер + барел) + глобальний
  bridge у `client.tsx`. Канали розділені БЕЗ dedup: `onCaughtError` = boundary-caught render;
  `window.onerror`+`unhandledrejection` = решта; `describeUnknown` дає distinct-повідомлення для
  non-Error reject; per-session ceiling 50; `ingestClientLog().catch(()=>{})` ковтає transport.
- **S3** `c8184c0` — наявний catch у `server.mjs` збагачено `source:"ssr"`+`release`+`route`;
  білдери винесено в `server-log.mjs` (+`.d.mts`) для unit.
- **S4–S6** `b2df292` — loki (single-binary) + promtail (docker_sd через сокет, `source`→мітка) +
  grafana (provision datasources+dashboard) під `monitoring` profile, усі network-only; nginx
  `clientlog`-зона + e2e-variant + nested exact-match локація; APP_RELEASE build-arg+runtime-env.
- **S7** — AC1 e2e `observability.spec.ts` (зелений); ordered `make check` EXIT=0; цей звіт.

### Відхилення від плану (датовані рішення, 2026-10-21)

1. **Release define = `process.env.APP_RELEASE`**, не `__APP_RELEASE__` (план ряд.51-52). Дзеркалить
   наявний WAYFORPAY-define, типізація через `@types/node` → **знімає NB-c** (ambient `.d.ts` не
   потрібен) і природно дає NB-b (client build-arg і SSR runtime читають одну env).
2. **Frontend-розкладка `src/modules/observability/`**, не top-level `src/ports/`/`src/adapters/
   observability/` (план ряд.37). Щоб adapter-тест матчив node-glob rstest (`src/modules/*/adapters/**`,
   реальний fetch+MSW) і потрапляв під architecture/module-boundary. Інтент плану (порт+адаптер, без
   fetch у presentation) збережено.
3. **Клієнтський §17-scrub ПРИБРАНО** (план §New code ряд.48 `[G2:NB AC2]`). Round-3 рецензент флагнув
   його як drift-liability (3-я копія never-log, TS↔Go). Сервер редагує авторитетно (S1 тестовано),
   AC2-ядро тримається сервером. Pre-launch cut-freely (власник).
4. **Транспорт — generated `ingestClientLog().catch`**, не `createApiTransport` (план §Reuse ряд.15).
   Result/span/timeout надмірні для fire-and-forget; адаптер лише ковтає.
5. **promtail через `/var/run/docker.sock` (docker_sd)**, не файловий mount `/var/lib/docker/containers`
   (план ряд.104). Сокет працює однаково на Linux-деплої і Docker Desktop, тож proof бігає так само,
   як деплой. Інтент (promtail скрейпить stdout, `source`→мітка) збережено.

## 2. Приймальна таблиця (spec AC1–AC7)

| AC | Очікуване | Фактичний результат | Вердикт |
|---|---|---|---|
| AC1 | client-помилка → один event route+release+UA через порт |  S2 rstest (adapter) + AC1 e2e (рівно 1 POST→**204**, bounded-wait) — зелений у гейті | збіглося |
| AC2 | §17-поле/секрет не потрапляє у payload | S1 domain unit (field+nested+in-text secret) + http contract (body-з-секретом→400 без echo) + наскрізний Loki-доказ (0 secret, 1 `[redacted]`) | збіглося |
| AC3 | SSR-помилка як окремий source | S3 unit (render_failed/http_request несуть `source=ssr`+release+route); route-render→200 client-owned (обмеження §5) | збіглося |
| AC4 | API+worker+SSR+client у Loki за `source` | monitoring-stack.sh: source-мітки `{api,asr-worker,client,media-worker,ssr}` queryable | збіглося |
| AC5 | Grafana datasources+dashboard з provisioning | monitoring-stack.sh: prometheus+loki datasources + `platform-overview` dashboard via Grafana API | збіглося |
| AC6 | monitoring лише під profile, network-only | monitoring-stack.sh: grafana/loki/prometheus/promtail — no host port | збіглося |
| AC7 | ordered `make check` зелений; default-up не стартує monitoring | ordered `make check` EXIT=0 (e2e 124 passed); monitoring під `profiles:[monitoring]`, `make up` без профілю їх не стартує | збіглося |

## 3. monitoring-stack.sh (profile-on, production-path, поза `make check`)

`sh infrastructure/tests/monitoring-stack.sh` → PASS: піднімає `monitoring` profile, генерує по рядку
на source, і доводить:
- **AC4** source-мітки у Loki: `{"data":["api","asr-worker","client","media-worker","nginx","other","ssr","upload"]}`.
- **AC5** Grafana API: datasources prometheus+loki; dashboard `platform-overview`.
- **AC6** `docker compose port` grafana/loki/prometheus/promtail → жодного host-порту.
Додатково вручну: `nginx -t` ok; `/client-logs` наскрізь 204 (valid) / 400 (unknown field); §17-scrub
наскрізь (у Loki client-лінії 0 `shouldscrub`, 1 `password=[redacted]`).

## 4. Незалежне code-review (verify)

Дві осі паралельними fresh-субагентами (автор не самоверифікує; робочі файли
`_wip-review-standards.md`, `_wip-review-spec.md`): **Standards — PASS** (hard-violations 0),
**Spec — PASS** (усі 7 AC, 5 відхилень обґрунтовані, blocking 0). Закрито два non-blocking пункти
перед PR: (1) `scrubSecrets` застосовано і до `UserAgent`/`Release` (не лише message/route/stack) —
глибший §17-захист від hostile-клієнта; (2) e2e-канал `unhandledrejection` доданий (AC1 тепер має
browser-доказ для 2 із 3 каналів; `onCaughtError` воронкою веде в той самий `capture`, покритий unit).

## 5. Верифікація

- Backend: `go test ./internal/clientlog/...` (domain 6 + http 8), `go vet`, architecture boundary, `go build ./cmd/api`.
- Frontend: rstest 11 (adapter 8 + ssr-log 3), `tsc`, `oxfmt --check`, `oxlint`, architecture check.
- Ordered `make check`: **EXIT=0** (frontend 784 unit, 34 Go-пакети, e2e 124 passed).

## 6. Обмеження

1. **route-render помилки → HTTP 200, не SSR-канал.** Start 1.169.3 обгортає маршрути в Suspense:
   render-помилка маршруту резолвиться як 200 з відкладанням на клієнт (її ловить клієнтський канал
   S2). AC3 покриває лише shell-throw/503-клас, який catch реально бачить. Механізм: коментар
   `server.mjs` + Start Suspense-поведінка. Не hand-picked throw.
2. **render_failed пише в STDERR, не STDOUT.** docker json-file зводить обидва потоки → promtail
   (docker_sd) читає обидва, тож source=ssr покриває і успіх (stdout), і render_failed (stderr).
   Детермінований тригер SSR shell-throw у запущеному стеку не підведений (той самий клас, що
   відкладені runtime-прогалини TSSTART-01 O3/O4) — render_failed доведено unit'ом (той самий білдер)
   + спільністю мітки з http_request.
3. **boot-вікно.** Помилка до виконання `client.tsx` (до реєстрації bridge) не ловиться. Module-скрипти
   виконуються до `load`, тож вікно вузьке, але не нульове.
4. **per-session ceiling 50** — storm-guard браузера (інший concern, ніж nginx-троттл: захищає ресурси
   клієнта). Понад 50 подій/завантаження сторінки дропаються тихо.
5. **worker source-мітки = `media-worker`/`asr-worker`** (імена логерів), не узагальнене `worker`.
   AC4 вимагав «worker» queryable — `media-worker` є у Loki.
6. **Клієнтський §17-scrub відсутній** (рішення §1.3): сервер — єдиний авторитет редакції. Якщо серверна
   редакція має прогалину, клієнтського backstop немає; натомість нема ризику TS↔Go drift.
7. **JWT-scrub best-effort** (shape-only regex): може пропустити нетиповий токен або зрідка
   over-redact JWT-подібний не-секрет (безпечний напрям для лог-рядка).
8. **Прод-деплою немає** (власник 03.10): інфра збудована, не задеплоєна; Grafana доступна лише в
   docker-мережі до S01 (TLS/auth).
