# Observability (B-1) — plan (HOW)

Builds on `spec.md` (GATE 1 passed 2026-10-20). Stack per owner: Grafana + Loki + promtail, with the
existing Prometheus. GATE 1 decisions: client errors → backend ingest endpoint; promtail scrapes
container stdout; one "platform overview" dashboard; `monitoring` profile stays opt-in; release id
from a build/version env.

**GATE 2 revision (2026-10-20):** this edition closes the 8 blocking findings (B1–B8) and the tight
defense gaps from the first GATE 2 (`_wip-gate2-review.md`). Changes are marked `[G2:Bn]`.

## Reuse-first

| Concern | Reuse / New | Why |
|---|---|---|
| Client→API HTTP call (deadline, cancellation, failure classification) | **reuse** `frontend/src/adapters/api/transport.ts` | the client-log adapter posts through it, not a raw fetch |
| Per-request headers / SSR cookie forwarding | **reuse** `frontend/src/adapters/api/request-headers.ts` | consistent transport |
| Request-scoped wiring | **reuse** `frontend/src/app/composition/` | the logger port is wired here, not imported by presentation |
| HTTP server, error envelope, request id, WriteJSON | **reuse** `backend/platform/httpserver/` | the ingest handler returns the standard envelope |
| Structured JSON logger (request id, no secrets, no query strings) | **reuse** `backend/platform/observability/` | the ingest handler emits through it directly; Loki then scrapes stdout |
| Prometheus server + `prometheus.yml` + `--profile monitoring` | **reuse** `docker-compose.yml`, `infrastructure/monitoring/prometheus.yml` | add Loki/Grafana/promtail beside it under the same profile |
| Per-endpoint rate limit | **reuse** the nginx zone pattern (the existing `login` zone + its `*.e2e.conf` variant) | a new `clientlog` nginx zone; no cross-module import of identity's Redis limiter (ADR 013) |
| Redis token-bucket limiter (`identity/adapters/redis`) | **not reused** | lives inside the identity module; ADR 013 forbids importing it; nginx zone is the right layer for a transport throttle |

New code:
- **`backend/internal/clientlog/`** — a small module. `domain`: a `ClientEvent` value (level, message,
  route, release, coarse UA, optional error name/stack) with validation, **size caps**, and a
  **§17 redaction that works at two levels** `[G2:B5]`: (a) forbid any *field* in the never-log set,
  and (b) *value-scrub* `message`, `error.stack` and any url-like substring for secret patterns
  (`password=`, `Bearer `, `token=`, cookie/JWT-shaped strings) before the event is emitted. A
  sentinel fills empty `release`→`"unknown"` / `route`→`"unknown"` so the earliest boot error is not
  dropped `[G2:NB]`. `adapters/http`: `POST /api/v1/client-logs` decoding one event (reuse
  `jsonbody.ValidateUniqueKeys`, `DisallowUnknownFields`, body cap), validating, redacting, emitting
  **directly through `platform/observability`'s logger** — no `EventSink` port, because a port with a
  single stdout implementation and no external dependency is speculative (AGENTS.md: ports require an
  actual external dependency) `[G2:NB]`. Invalid body → `400` standard envelope (no secrets); valid →
  `204` `[G2:NB]`. No persistence, no DB.
- **Frontend** `frontend/src/ports/client-log.ts` (port), `frontend/src/adapters/observability/client-log-adapter.ts`
  (posts via `transport.ts`). The adapter **never reports its own transport failure** — a failed POST
  is swallowed, never re-captured `[G2:B6]`. Capture channels are **cleanly separated, with no dedup**
  `[G2:B7,N4,N5 — reworked round 2]`: the React **error boundary** (React 19 `onCaughtError`) owns
  render errors; a `window.onerror` + `unhandledrejection` bridge owns everything else. In the prod
  build e2e runs against, StrictMode is a no-op and React 19 routes boundary-caught errors to
  `onCaughtError`, not `window.onerror`, so the channels do **not** overlap — separation alone gives
  one event per failure. The round-1 signature+TTL dedup is **removed**: unnecessary (no overlap in
  prod) and harmful (a non-`Error` rejection — `reject("x")`, `undefined`, an object — has no
  message/stack, so distinct failures collapse to one signature and lose N−1 events, breaking AC1 the
  other way). **No batching in v1** `[G2:B8]`: one POST per event; a per-session ceiling counter caps a
  storm. Installed in the composition root, not a component. A client-side pre-send scrub mirrors the
  backend never-log set as defence in depth `[G2:NB AC2]`. **Boot-window limitation** `[G2:NB]`: an
  error before the bridge installs cannot be captured — documented, not hidden.
- **Release identifier** `[G2:B2,N3 — reworked round 2]`: the **client** bundle gets a build-time
  constant via Rsbuild `source.define` (`__APP_RELEASE__`) from `APP_RELEASE` (git short SHA at build).
  The **SSR** `server.mjs` is **copied verbatim by the Dockerfile, not bundled**, so the define never
  reaches it — `server.mjs` reads `process.env.APP_RELEASE` **at runtime** instead (Compose passes the
  env to the frontend service). Dev/gate default `"dev"` on both paths.
- **SSR source** `[G2:B1,N1,N2 — reworked round 2]`: the real SSR process is **`frontend/server.mjs`**
  (Dockerfile prod `CMD ["node","server.mjs"]`, `package.json` `start`), NOT `src/server.ts` (a bare
  re-export). `server.mjs` **already** wraps `entry.fetch` in try/catch and **already** writes
  structured JSON to stdout (`http_request`+status on success; `render_failed`→503 on a shell/non-
  render throw). So S3 does **not** add a parallel wrap — it **enriches the existing lines**: add
  `source:"ssr"` + `release` + `route` to both the success and the `render_failed` lines. No new
  `ssr-logger.ts`.
  **AC3 scope correction** `[G2:N2]`: `server.mjs`'s own comment records that Start 1.169.3 wraps route
  components in Suspense, so **route-render errors resolve as HTTP 200 with the error deferred to the
  client** (caught by the S2 client logger), not a 5xx/throw. AC3 therefore covers only the
  shell-throw/503 class the catch actually sees; route-render→200 is a **documented limitation**
  (owned by the client channel), not a hand-picked throw dressed as full SSR coverage.
- **Infra** under `--profile monitoring`: `loki`, `promtail`, `grafana`. promtail uses a **pipeline
  stage that parses the JSON `source` field into a Loki label** `[G2:B4]` (so `client`, `ssr`, `api`,
  `worker` separate even though `client` lines leave via the `api` container's stdout), not a per-
  container label. `grafana` provisions datasources (Prometheus+Loki) + one dashboard JSON. Configs in
  `infrastructure/monitoring/`.
- **nginx** `[G2:NB]`: a `clientlog` rate-limit zone (modest req/s) in the prod conf **and** a relaxed
  `clientlog-req.e2e.conf` variant mounted in `docker-compose.e2e.yml`, mirroring the `login` zone
  pattern, so the strict zone does not 429 client POSTs during e2e.

## Architecture / dependency direction

- `clientlog` module obeys the standard layering: domain (stdlib + contracts only, holds validation +
  redaction + scrubbing), adapters/http (transport → domain → logger). The handler depends on the
  `platform/observability` logger directly; no new port. Wired in the api composition root. No module
  imports another module.
- Frontend: presentation/routes never call `fetch`; the error boundary and global bridge depend on
  the `client-log` **port**; the adapter (which uses `transport.ts`) is bound in composition. Keeps
  `frontend/scripts/check-architecture.mjs` green.
- The SSR logger is a leaf TS adapter writing to stdout; it imports no app module and no Go code.
- Monitoring services are infra, not app code; they bind only to the docker network.

## Side-effects

- A new public route `POST /api/v1/client-logs` (**unauthenticated by design** — pre-auth client
  errors must report). Mitigations: nginx `clientlog` rate-limit zone; strict body cap +
  `DisallowUnknownFields`; two-level §17 redaction; invalid→400 / valid→204; only emits a log line,
  touches no DB.
- **Client→ingest degradation** `[G2:defense Q2]`: if `/client-logs` returns 429/5xx, the adapter
  swallows it (no retry, no self-report, no loop — B6). Logs the browser never ships are simply lost;
  nothing breaks.
- **Loki-down degradation** `[G2:NB]`: application logs always go to container stdout regardless of
  Loki; Loki only indexes what is already on stdout. If Loki is down, logging is unaffected; the AC4
  probe polls Loki readiness/indexing (event-driven, not `sleep`).
- New Compose services increase resource use **only when the profile is on**; default `up` (the gate
  path) does not start them, so `make check` is unaffected (AC7, verified: `make up` runs
  `docker compose up` with no `--profile monitoring`).
- promtail reads the Docker container log dir (read-only mount); no app change.

## План міграції

- No schema migration (no DB tables). Additive: a new backend module, new frontend files, a Rsbuild
  define, a wrapped `server.ts`, new infra services, one nginx zone (+ e2e variant).
- Grafana/Loki data live in named volumes created on first `--profile monitoring` up; absence is
  harmless (default path never starts them).
- Rollout order (slices, each independently shippable): **S1** backend `clientlog` module + endpoint +
  release define plumbing; **S2** frontend port+adapter+error boundary+global bridge+client scrub
  (channel separation, no dedup); **S3** enrich `server.mjs` existing catch (source/release/route);
  **S4** Loki + promtail (source-label pipeline);
  **S5** Grafana provisioning + overview dashboard; **S6** nginx zone (+e2e variant) + Compose profile
  wiring + the profile-on end-to-end proof.

## Rollback

- `[G2:NB]` Slices do **not** revert blindly-independently: reverting S1 (endpoint) while S2 (client
  adapter) is live would 404 every client POST — but because the adapter **swallows** transport
  failures (B6), that 404 is a silent drop, not a storm. So the safe revert order is S2→S1, and even
  out of order nothing breaks louder than lost logs.
- Infra reverts by removing the profile services + volumes. The nginx zone reverts with its block. No
  data migration to reverse (no persisted state).

## Конституція (manifest)

- AGENTS.md architecture: domain stdlib-only, adapters inward, no cross-module imports, no `fetch` in
  presentation — honoured. No speculative port (EventSink dropped).
- §17 never-log set enforced in `clientlog` domain at field + value level (AC2), tested with a nested
  and an in-text secret `[G2:B5]`.
- Metrics/monitoring stay inside the docker network until TLS/S01 (AC6).
- No secrets in logs; URLs without query strings — the ingest path forbids both.

## Acceptance scenarios

Delegated to `acceptance-spec`; the table maps each spec AC to an observable check and the slice that
closes it.

| AC | Observable check | Slice |
|---|---|---|
| AC1 | render error / unhandled rejection / boundary catch → exactly **one** event with route+release+UA via the port, guaranteed by channel separation (no dedup); unit on adapter + each channel incl. a non-`Error` rejection; e2e triggers a client error and asserts exactly one `/client-logs` POST **that returns 204** `[G2:N6]` (a swallowed 429 must not read as a pass), awaiting the network response | S2 |
| AC2 | an event with a §17 field is dropped **and** a secret embedded in `message`/`stack`/url is scrubbed (domain unit, nested + in-text cases); client-side pre-send scrub unit | S1,S2 |
| AC3 | a shell/non-render throw in `server.mjs`'s existing catch emits a `source=ssr` `render_failed` line carrying release+route (unit/integration against `server.mjs`); route-render→200 errors are out of this channel by design (client-owned), stated as a limitation `[G2:N2]` | S3 |
| AC4 | API+worker+SSR+client stdout queryable in Loki **by `source` label** (profile-on Compose proof; LogQL asserts each source) | S4,S5,S6 |
| AC5 | Grafana boots with Prometheus+Loki datasources + overview dashboard from repo provisioning (profile-on Compose proof via Grafana API) | S5 |
| AC6 | monitoring services only under `monitoring` profile, docker-network-only (compose inspection + `docker ps`/network proof) | S6 |
| AC7 | ordered `make check` green with the frontend logger wired; default path does not start monitoring infra (verified) | all |

**AC4/AC5/AC6 proof path** `[G2:B3]`: these cannot be proven by the default gate (`make up` has no
`--profile monitoring`). A dedicated script `infrastructure/tests/monitoring-stack.sh` (mirroring
`image-authorize.sh`) brings the profile up, asserts each `source` via a LogQL query, checks Grafana's
datasource+dashboard via its HTTP API, and asserts network-only via `docker ps`/network inspect. It
runs outside `make check`; its result is recorded in `report.md` as the production-path proof.

## Open (plan-level, resolved here)

- nginx `clientlog` zone: prod is modest (mirror `login`, e.g. 10 r/s burst 20). The **e2e variant**
  `clientlog-req.e2e.conf` must be sized for the whole suite on one docker IP (`$binary_remote_addr`
  is shared) so a legitimate AC1 target POST is never 429'd `[G2:N6]`; exact burst set in S6.
- Per-session client report ceiling: a small fixed cap (e.g. 50/session) — storm guard only, no batching.
- Grafana access: anonymous read within the docker network (no external exposure until S01).
