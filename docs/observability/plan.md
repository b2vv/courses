# Observability (B-1) — plan (HOW)

Builds on `spec.md` (GATE 1 passed 2026-10-20). Stack per owner: Grafana + Loki + promtail, with the
existing Prometheus. GATE 1 decisions: client errors → backend ingest endpoint; promtail scrapes
container stdout; one "platform overview" dashboard; `monitoring` profile stays opt-in; release id
from a build/version env.

## Reuse-first

| Concern | Reuse / New | Why |
|---|---|---|
| Client→API HTTP call (deadline, cancellation, failure classification) | **reuse** `frontend/src/adapters/api/transport.ts` | the client-log adapter posts through it, not a raw fetch |
| Per-request headers / SSR cookie forwarding | **reuse** `frontend/src/adapters/api/request-headers.ts` | consistent transport |
| Request-scoped wiring | **reuse** `frontend/src/app/composition/` | the logger port is wired here, not imported by presentation |
| HTTP server, error envelope, request id, WriteJSON | **reuse** `backend/platform/httpserver/` | the ingest handler returns the standard envelope |
| Structured JSON logger (request id, no secrets, no query strings) | **reuse** `backend/platform/observability/` | the ingest handler emits through it; Loki then scrapes stdout |
| Prometheus server + `prometheus.yml` + `--profile monitoring` | **reuse** `docker-compose.yml`, `infrastructure/monitoring/prometheus.yml` | add Loki/Grafana/promtail beside it under the same profile |
| Per-endpoint rate limit | **reuse** the nginx zone pattern (the existing `login` zone) | a new `clientlog` nginx zone; no cross-module import of identity's Redis limiter (ADR 013) |
| Redis token-bucket limiter (`identity/adapters/redis`) | **not reused** | lives inside the identity module; ADR 013 forbids importing it; nginx zone is the right layer for a transport throttle |

New code:
- **`backend/internal/clientlog/`** — a small module. `domain`: a `ClientEvent` value (level, message,
  route, release, coarse UA, optional error name/stack) with validation + a **§17 redaction** step
  that drops/forbids any field matching the never-log set and caps sizes. `ports`: an `EventSink`
  (emits one structured log line). `adapters/http`: `POST /api/v1/client-logs` decoding one event
  (reuse `jsonbody.ValidateUniqueKeys`, `DisallowUnknownFields`, body cap), validating, redacting,
  emitting via the sink → stdout → Loki. No persistence, no DB.
- **Frontend** `frontend/src/ports/client-log.ts` (port), `frontend/src/adapters/observability/client-log-adapter.ts`
  (posts via `transport.ts`), a React **error boundary** reporting through the port, and a global
  `window.onerror` / `unhandledrejection` bridge installed in the composition root (not in a component).
- **SSR**: in `frontend/src/server.ts`, emit a structured JSON line (source=`ssr`) on a route-lane
  500 / failed upstream, distinct from browser (source=`client`) and Go API.
- **Infra** under `--profile monitoring`: `loki`, `promtail` (scrapes `/var/lib/docker/containers`
  with a source label per container), `grafana` (provisioned datasources Prometheus+Loki + one
  dashboard JSON). Configs in `infrastructure/monitoring/`.

## Architecture / dependency direction

- `clientlog` module obeys the standard layering: domain (stdlib + contracts only, holds validation +
  redaction), ports (EventSink), adapters/http (transport → domain → sink). The sink adapter wraps
  `platform/observability`. Wired in the api composition root. No module imports another module.
- Frontend: presentation/routes never call `fetch`; the error boundary and global bridge depend on
  the `client-log` **port**; the adapter (which uses `transport.ts`) is bound in composition. This
  keeps the machine architecture check (`frontend/scripts/check-architecture.mjs`) green.
- Monitoring services are infra, not app code; they bind only to the docker network.

## Side-effects

- A new public route `POST /api/v1/client-logs` (unauthenticated by design — pre-auth client errors
  must report). Mitigations: nginx `clientlog` rate-limit zone; strict body cap + `DisallowUnknownFields`;
  §17 redaction; the endpoint only emits a log line, touches no DB, returns `204`.
- New Compose services increase resource use **only when the profile is on**; default `up` (the gate
  path) does not start them, so `make check` is unaffected (AC7).
- promtail reads the Docker container log directory (read-only mount); no app change.
- Client error volume could be noisy: the adapter batches/caps client reports and drops beyond a
  per-session ceiling.

## План міграції

- No schema migration (no DB tables). Purely additive: a new module, new frontend files, new infra
  services, one nginx zone.
- Grafana/Loki data live in named volumes created on first `--profile monitoring` up; absence is
  harmless (default path never starts them).
- Rollout order (slices, each independently shippable): S1 backend `clientlog` module + endpoint; S2
  frontend port+adapter+error boundary+global bridge; S3 SSR source tagging; S4 Loki+promtail; S5
  Grafana provisioning + dashboard; S6 nginx zone + Compose profile wiring + end-to-end proof.

## Rollback

- Each slice reverts independently. The endpoint + frontend adapter revert by removing the module +
  files + composition wiring (no persisted state to unwind). Infra reverts by removing the profile
  services + volumes. nginx zone reverts with its block. No data migration to reverse.

## Конституція (manifest)

- AGENTS.md architecture: domain stdlib-only, adapters inward, no cross-module imports, no `fetch` in
  presentation — all honoured above.
- §17 never-log set enforced in `clientlog` domain redaction (AC2), tested.
- Metrics/monitoring stay inside the docker network until TLS/S01 (AC6).
- No secrets in logs; URLs without query strings — the ingest path forbids both.

## Acceptance scenarios

Delegated to `acceptance-spec`; the table below maps each spec AC to an observable check and the
slice that closes it. (To be expanded by the acceptance-spec pass before tasks.)

| AC | Observable check | Slice |
|---|---|---|
| AC1 | thrown error / unhandled rejection / boundary catch → one event with route+release+UA via the port (unit on adapter; e2e triggers a client error and asserts one `/client-logs` POST) | S2 |
| AC2 | an event built from a context carrying a §17 field does not transport it (domain redaction unit) | S1 |
| AC3 | an SSR route-lane 500 emits a `source=ssr` structured line (SSR unit / integration) | S3 |
| AC4 | API+worker+SSR stdout queryable in Loki by source label (Compose proof) | S4,S6 |
| AC5 | Grafana boots with Prometheus+Loki datasources + overview dashboard from repo provisioning (Compose proof) | S5 |
| AC6 | monitoring services only under `monitoring` profile, docker-network-only (compose inspection + proof) | S6 |
| AC7 | ordered `make check` green with the frontend logger wired; default path does not start monitoring infra | all |

## Open (to plan.md, not spec)

- Exact nginx `clientlog` zone rate (reuse login zone's shape; pick a modest req/s).
- Client report batching ceiling per session.
- Grafana anonymous-viewer vs admin login for the in-network UI (default: anonymous read within
  docker network, since no external exposure until S01).
