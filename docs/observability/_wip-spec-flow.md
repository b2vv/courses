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
- **НАСТУПНА ДІЯ (нова сесія): GATE 2 раунд 3** свіжим незалежним сабагентом на плані ред.round2.
  Якщо зелений → tasks S1–S6 → tdd по зрізах → code-review → ship-report.

## Журнал
- specify: spec.md написано (Проблема/Сценарії/AC1–AC7/Out of scope/5 відкритих).
- GATE 1: сфокусований decision-раунд (не повний grilling, flagged); 3 рішення + 2 дефолти. → plan.
- plan v1 → GATE 2 раунд 1 ЧЕРВОНИЙ → plan ред.GATE2 (8 blocking + defense закрито) → GATE 2 раунд 2.

