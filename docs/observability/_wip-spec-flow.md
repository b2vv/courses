# Observability (B-1) — жива нота spec-flow

Гілка `feat/observability-full-stack` (база main `8c77860`). Дім артефактів `docs/observability/`.
Поріг: **вище** (нова підсистема) → spec-flow увімкнено. Власник 2026-10-20: повний стек зараз.

## План проходу
specify✓ → GATE1 (grilling / shared understanding) → plan → GATE2 (plan-critique + plan-defense +
конституція) → tasks → implement (tdd, по зрізах) → verify (code-review) → ship-report.

## Рішення
- 2026-10-20 (власник): B-1 повний стек зараз (frontend logger + Loki + Grafana + дашборди),
  стек Grafana+Loki (Prometheus уже є, таргет 1 хост 4GB). Інфра будується, не вмикається в прод.

## GATE 1 — підхід
Проводжу як сфокусований decision-раунд по 5 відкритих питаннях spec (мета гейта — спільне
розуміння + підтвердження власника), а не повний theatrical `grilling`-дрил, зважаючи на лаконічний
cadence власника. Деталь названо тут задля відтворюваності (spec-flow забороняє тиху заміну).

## Відкрите (до відповіді власника)
Див. spec.md §Відкриті питання: (1) куди йдуть client-помилки [lean: backend ingest-ендпоінт];
(2) Loki ingestion [lean: promtail]; (3) обсяг дашборду v1; (4) release-ідентифікатор; (5) профіль
monitoring opt-in [lean: так].

## GATE 1 — ПРОЙДЕНО 2026-10-20 (власник підтвердив)
- Client transport: **backend ingest-ендпоінт** `POST /api/v1/client-logs` → структурований лог → Loki.
- Loki ingestion: **promtail** скрейп stdout контейнерів.
- Дашборд v1: **один «platform overview»** (Prometheus метрики + 2 media + Loki logs-панель).
- Дефолти: профіль `monitoring` opt-in = так; release-id = build hash/env.

## GATE 2 — раунд 1 ЧЕРВОНИЙ (2026-10-20), доопрацьовано
Незалежний сабагент (general-purpose) прогнав plan-critique+plan-defense → 8 blocking + 4 незакриті
defense-питання. Деталі: `_wip-gate2-review.md`. Несучі факти звірено кодом (server.ts 91B re-export;
нема release-env; make up без monitoring-профілю). План переписано (ред. GATE2), мітки `[G2:Bn]`:
- B1 SSR: wrap server-entry default export + `ssr-logger.ts` (TS→stdout); B2 `__APP_RELEASE__` Rsbuild
  define (git SHA, dev-default); B3 окремий `monitoring-stack.sh` profile-on proof (поза гейтом);
  B4 promtail JSON `source`→Loki-мітка; B5 two-level §17 (field + value-scrub message/stack/url);
  B6 adapter ковтає власні transport-помилки (негативний тест); B7 канальний поділ boundary/onerror +
  дедуп(signature+TTL)+StrictMode → один event; B8 без батчингу, один POST/подію + per-session ceiling.
- NB: drop EventSink port (пряма ін'єкція logger'а); 400/204; sentinel unknown/dev; rollback S2→S1
  безпечний через swallow; nginx `clientlog-req.e2e.conf`; Loki-down — stdout незалежний, проба poll.
- GATE 2 раунд 2 ЧЕРВОНИЙ (`_wip-gate2-round2.md`): кодова звірка показала корінь — реальний SSR =
  `frontend/server.mjs` (не `src/server.ts`), уже обгортає entry.fetch + логує JSON. 4/8 чисто закрито
  (B3/B4/B5/B8), B6 закрито; РЕОПЕН B1/B2/AC3 (SSR) + 6 нових (N1 wrap не той файл; N2 route-render→200
  не throw, AC3 звузити; N3 release через runtime-env у server.mjs; N4 dedup зайвий+шкідливий — прибрати;
  N5 TTL без величини; N6 e2e clientlog-зона не розмірена + AC1 має вимагати 204).
- План ред.round2 (мітки `[G2:N*]`): SSR → збагатити НАЯВНИЙ catch у server.mjs (source/release/route),
  AC3 звужено до shell-throw/503 + route→200 як limitation; release: client=define, SSR=process.env;
  dedup ПРИБРАНО (поділ каналів — єдина гарантія); AC1 e2e вимагає 204; nginx e2e-зона під обсяг сьюту.
- GATE 2 раунд 3 **ЗЕЛЕНИЙ** 2026-10-20 (свіжий незалежний general-purpose рецензент, `_wip-gate2-round3.md`):
  6/6 N-findings закрито механізмом на рядку + звірено проти коду (HEAD 1a14cb9); 0 нових blocking;
  defense Q1–Q5 чисто (Q3 реопен round2 — закрито ПОДІЛОМ каналів, не поверненим dedup). 4 NB-деталі
  wiring занесено в таски (stderr render_failed; однакове release client/SSR; ambient __APP_RELEASE__;
  AC1 bounded-wait). Залишкові NB → у звіт як обмеження.
- Таски S1–S7 записано у `work/tasks.md` §Phase observability (B-1).
- **НАСТУПНА ДІЯ: імплементація S1** (backend clientlog module + endpoint + release-define) за tdd,
  тоді S2…S7 по зрізах; code-review; ship-report.

## Журнал
- specify: spec.md написано (Проблема/Сценарії/AC1–AC7/Out of scope/5 відкритих).
- GATE 1: сфокусований decision-раунд (не повний grilling, flagged); 3 рішення + 2 дефолти. → plan.
- plan v1 → GATE 2 раунд 1 ЧЕРВОНИЙ → plan ред.GATE2 (8 blocking + defense закрито) → GATE 2 раунд 2.
- GATE 2 раунд 2 ЧЕРВОНИЙ → plan ред.round2 → GATE 2 раунд 3 **ЗЕЛЕНИЙ** 2026-10-20 → tasks S1–S7.
- S1 **done** 2026-10-21 (`f0ea2f3`): backend clientlog module (domain+http) + endpoint + OpenAPI/Orval
  + Rsbuild define. Рішення: define-ключ `process.env.APP_RELEASE` (знімає NB-c ambient, дає NB-b).
  Per-slice: domain 6 + http 8 тестів, gofmt/vet/architecture, go build cmd/api, FE typecheck/format/lint.
- S2 **done** 2026-10-21: `modules/observability/{ports,adapters,index}` + bridge у `client.tsx`.
  Рішення: modules-розкладка (node-glob); generated-транспорт; лише `onCaughtError`+window-bridge;
  клієнт-scrub ПРИБРАНО (сервер авторитетний, round-3 NB); ceiling 50 лишено. 8 rstest node зелені.
  e2e AC1 (204+bounded-wait, реальний браузер) відкладено в S7 (потрібен стек up).
- S3 **done** 2026-10-21 (`c8184c0`): server.mjs catch збагачено; винесено server-log.mjs (+d.mts), 3 unit.
- S4/S5/S6 **done** 2026-10-21: loki+promtail(docker_sd socket)+grafana(provision+dashboard) під monitoring
  profile, nginx clientlog-зона+e2e-variant+nested location, APP_RELEASE build-arg+runtime-env, Dockerfile.
  **monitoring-stack.sh PASS** (production-path, поза make check): AC4 source-мітки
  {api,asr-worker,client,media-worker,ssr}; AC5 datasources prometheus+loki + dashboard platform-overview;
  AC6 grafana/loki/prometheus/promtail no host port. nginx -t ok; /client-logs наскрізь 204/400;
  §17-scrub наскрізь (0 secret, 1 [redacted] у Loki client-лінії).
  Рішення: promtail через docker.sock (не log-dir mount) — portable Linux+Desktop; COMPOSE_PROFILES env
  (не --profile flag — compose v5.5.1 парсинг); клієнтський §17-scrub знято (S2).
- S7 **done** 2026-10-21: AC1 e2e `observability.spec.ts` (рівно 1 POST→204, bounded-wait) зелений;
  ordered `make check` **EXIT=0** (frontend 784 unit, e2e **123 passed**, 14.7m); звіт `report.md`.
  **ФАЗА observability (B-1) ЗАКРИТА (S1–S7).** Наступне: PR у main (гілка feat/observability-full-stack).

