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

## Журнал
- specify: spec.md написано (Проблема/Сценарії/AC1–AC7/Out of scope/5 відкритих).
- GATE 1: сфокусований decision-раунд (не повний grilling, flagged); 3 рішення + 2 дефолти. → plan.

