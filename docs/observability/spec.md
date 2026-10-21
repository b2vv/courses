# Observability (B-1) — specification

Phase source: `work/backlog.md` B-1. Owner decision 2026-10-20: build the **full stack** now
(frontend logger + Loki + Grafana + dashboards), despite the product being undeployed — infra is
built but not wired to prod. This file is intent only (WHAT/WHY); stack choices named here are the
owner's pre-decision, concrete HOW lands in `plan.md`.

## Проблема

Three operator-facing blind spots, all verified in code 2026-10-20:

1. **Browser-side failures are invisible.** There is no frontend logger, no error-boundary
   reporting, no client error transport. A thrown render error, an unhandled rejection or a failed
   client action leaves no trace the operator can see.
2. **SSR failures blur into the browser or the API.** The SSR Node process is a distinct source
   (a 500 in a route lane, a failed upstream API call) but its failures are not separated.
3. **Logs are ephemeral and metrics are unread.** Backend structured JSON logs exist but live only
   as container stdout, lost on restart; nothing ships or indexes them. Prometheus stores series
   with no UI reading them.

What already exists (do not rebuild): backend structured JSON logs (`platform/observability`, with
request id / trace id, no secrets, URLs without query strings), Prometheus metrics + the two media
metrics, a Prometheus server under `--profile monitoring`, partial OpenTelemetry tracing.

## Користувацькі сценарії (оператор)

1. A browser error (thrown render error, unhandled promise rejection, or an error boundary catch)
   is captured and reaches a place the operator can read, carrying **context** (route path, release
   identifier, coarse user-agent) and **never** any §17 secret.
2. An SSR-process failure is logged as its own source, distinguishable from both the browser and
   the Go API.
3. The operator opens one UI (Grafana) and sees key platform metrics (request rate / latency /
   error ratio, plus the two media metrics) and can search recent logs across API, worker and SSR
   in the same place, filtered by source and time.
4. Everything above stays reachable only inside the docker network; nothing monitoring-related is
   exposed publicly until TLS + auth (S01) land.

## Acceptance criteria

- **AC1** A client-side thrown error, an unhandled rejection, and a React error-boundary catch each
  produce one captured log event carrying route + release + coarse UA, transported through an
  **adapter behind a port** (no `fetch` in presentation/routes, per AGENTS.md).
- **AC2** The capture path redacts / never includes the brief §17 set: passwords, session cookies,
  playback tokens, payment secrets, full private transcript documents. A test asserts a payload
  built from a context containing such a field does not carry it.
- **AC3** SSR process errors are emitted as structured logs tagged with a distinct source label.
- **AC4** Container logs for API, worker and SSR ship to Loki and are queryable in Grafana by source
  label and time range (proven against Compose, not by config inspection).
- **AC5** Grafana starts with provisioned datasources (Prometheus + Loki) and at least one
  dashboard showing the key metrics and a logs panel; provisioning is in the repo (not hand-clicked).
- **AC6** All monitoring services run only under the opt-in `monitoring` Compose profile and bind
  only to the docker network (no host port publishing that reaches outside), consistent with the
  existing Prometheus server.
- **AC7** The machine gates stay green with the frontend logger wired: `make check` ordered, e2e
  included. Monitoring infra (Loki/Grafana/promtail) is not started by the default gate path.

## Out of scope

- Enabling monitoring in production or any prod wiring (product is undeployed, owner 2026-10-03).
- Alerting / paging / on-call routing.
- Completing OpenTelemetry end-to-end tracing (SSR→API→worker spans) — separate, stays partial.
- A hosted Sentry SaaS or any external log sink — the client adapter is Sentry-*shape* but ships to
  our own network, never an external DSN (keeps §17 data in-network).
- TLS/auth for the monitoring UIs (that is S01; until then, docker-network-only).

## Відкриті питання (GATE 1)

1. **Куди йдуть клієнтські помилки?** Пропозиція: мінімальний backend ingest-ендпоінт
   (`POST /api/v1/client-logs`), rate-limited, що емітить структурований лог (далі Loki його
   забирає). Альтернативи: Sentry SaaS DSN (зовнішнє — суперечить in-network + §17 ризик) або
   прямий push у Loki з браузера (оголює Loki). Lean: backend ingest-ендпоінт.
2. **Як логи потрапляють у Loki?** promtail, що скрейпить stdout контейнерів (стандарт, без
   per-container driver), проти Docker loki logging driver проти Grafana Alloy. Lean: promtail.
3. **Обсяг дашборду v1:** один «platform overview» (request rate/latency/error-ratio з Prometheus
   + 2 media-метрики + logs-панель Loki) — достатньо для першої ітерації?
4. **Release-ідентифікатор для групування помилок:** звідки (build hash / version env)?
5. **Чи лишається профіль `monitoring` opt-in** (`--profile monitoring`) — пропозиція: так, бо не
   задеплоєно.
