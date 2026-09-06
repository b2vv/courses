# Backlog — named, not scheduled

Items that have been decided as *worth doing* but not placed in a phase. Not an agenda:
`work/AGENDA.md` ranks the next move and is rebuilt each time. An item leaves here by
becoming a phase in `work/PRODUCT-BRIEF.md` §21 or a task in `work/tasks.md`.

---

## B-1 — Observability: frontend logging, log aggregation, dashboards

**Raised:** 2026-09-06 (user), during Phase 5.
**Shape:** end-of-roadmap hardening, after the product does something worth watching.

### What already exists (verified in code, 2026-09-06)

| Piece | State | Where |
|---|---|---|
| Backend structured JSON logs | **exists** — request id, trace id, no secrets, URLs without query strings | `backend/platform/observability/`, the rule in `AGENTS.md` |
| Prometheus metrics | **exists**, plus the two media metrics with pre-initialised labels | `platform/observability/metrics.go`, Phase 5 M12 |
| Prometheus server | **exists**, optional profile, metrics never leave the docker network | `docker-compose.yml` (`--profile monitoring`), `infrastructure/monitoring/prometheus.yml` |
| OpenTelemetry tracing | **partial** — the package exists; end-to-end spans across SSR → API → worker are not proven | `platform/observability/tracing.go` |

### What is missing — the actual work

1. **Frontend logging does not exist at all.** No logger, no error boundary reporting, no
   client-side error transport. Today a browser-side failure is invisible to the operator.
   The product brief §3 already lists a **Sentry adapter** as the intended shape, and
   `AGENTS.md` forbids `fetch` in presentation — so this is an adapter behind a port, not
   `console.error` sprinkled in components.
2. **No log aggregation.** Logs exist as container stdout and are lost on restart. Nothing
   ships or indexes them: no Loki, no Elasticsearch.
3. **No dashboards.** Prometheus stores series with no one reading them; there is no Grafana,
   and no Kibana because there is no Elasticsearch.
4. **SSR is a distinct log source** and is easy to forget: it is a Node process whose failures
   (a 500 in a route lane, a failed API call) are separate from both the browser and the Go API.

### The choice to make when this is scheduled

Two coherent stacks, and mixing them is the expensive mistake:

- **Grafana + Loki + Prometheus** — one UI for metrics and logs, Loki indexes labels rather
  than full text, cheap to run on a single host, and Prometheus is **already here**. Log
  search is weaker than Elasticsearch's.
- **Elasticsearch + Kibana** (+ Grafana for metrics, or Kibana for both) — much stronger
  full-text search and analytics; substantially heavier (JVM, memory, index lifecycle
  management) for a self-hosted single-host deployment.

**Leaning:** Grafana + Loki, because Prometheus is already running and the deployment target
is one host with 4 GB — but this is a decision to take with the phase, not now.

### Why it is not scheduled yet

The roadmap's remaining phases (6 playback, 7 transcript, 8 orders) each add things worth
watching — playback sessions, transcode failures under real load, payment webhooks. Building
dashboards before those exist means instrumenting an empty product and rebuilding them after.

**Trigger to schedule:** the first phase that runs against real users, or the first incident
where a browser-side or SSR failure cannot be explained from container stdout — whichever
comes first.

### Constraints it must respect

- `AGENTS.md`: no secrets in logs; URLs logged without query strings; no `fetch` in
  presentation or routes — the client transport is an adapter behind a port.
- The brief's §17 list of what must never be logged: passwords, session cookies, playback
  tokens, payment secrets, full private transcript documents.
- Metrics stay internal to the docker network unless TLS and auth land first (`S01`).
