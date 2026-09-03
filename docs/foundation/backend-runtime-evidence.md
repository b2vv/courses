# F02 — backend runtime evidence

Date: 2026-09-03. Scope: Phase 1 runtime only. No Course implementation or Phase 3–8 capability is claimed.

## Implemented files/modules

- `backend/cmd/api`: signal-aware API composition root.
- `backend/cmd/media-worker`: Asynq executable, internal health/readiness/status/metrics, zero registered media handlers.
- `backend/cmd/migrate`: pgx stdlib + instance-scoped Goose provider; an empty migration directory is a successful no-op after database connectivity is checked.
- `backend/platform/httpserver`: bounded server, JSON errors, per-request generated request ID, safe route-template logging, trace propagation, panic recovery and HTTP metrics; graceful draining before dependencies close.
- `backend/platform/observability`: executable-scoped Prometheus registry, JSON slog logger, OTel tracer provider. Optional OTLP HTTP exporter uses standard endpoint environment variables; no exporter is required locally.
- `backend/platform/postgres`: bounded pgx pool and statement timeout; query spans/duration contain no SQL or arguments.
- `backend/platform/redis`: bounded Redis client and Asynq connection, queue metrics inspector and sanitized queue logging.
- `backend/platform/runtime`: explicit dependency ownership and teardown.
- `backend/Dockerfile`: `api`, `worker`, `migrate` targets, Go 1.27.1 builder, Debian runtime, UID/GID 10001; FFmpeg/ffprobe installed only in worker target.

## Runtime contract

Required: `DATABASE_URL`, `REDIS_URL` (for example `redis://redis:6379/0`). Optional: `HTTP_ADDR` (API `:8080`, worker `:8081`), `MIGRATIONS_DIR` (migrate default `migrations`), standard `OTEL_EXPORTER_OTLP_ENDPOINT` or `OTEL_EXPORTER_OTLP_TRACES_ENDPOINT` and related OTLP settings.

API and worker expose `GET /healthz`, `GET /readyz`, `GET /metrics` internally. Worker also exposes `GET /status`, returning `stage: foundation` and `handlers: []`. Readiness checks both PostgreSQL and Redis with a shared two-second deadline. Startup dependency budget is ten seconds. HTTP draining budget is ten seconds; worker queue shutdown adds up to ten seconds. Give containers at least 40 seconds stop grace to include monitoring/exporter shutdown under failures.

HTTP integration seam for Phase 2: construct a repository with `dependencies.Database` (`*pgxpool.Pool`) in `cmd/api`, then register handlers with `httpserver.Options.Register func(*http.ServeMux)`. `WriteJSON`, `WriteError` and `RequestID(context.Context)` are shared transport helpers. HTTP errors are `{error:{code,message,requestId}}`, and the response carries `X-Request-ID`. Untrusted incoming request IDs are replaced. Metrics/logs use registered route templates, never raw paths, query strings, cookies or SQL.

## Decisions

No unused domain ports or business-module placeholders were created. Internal executables own concrete runtime dependencies. No global registry/client/service locator is used. Worker consumes the `media` queue with no registered handlers; an externally enqueued unsupported task fails through Asynq. Foundation creates no jobs. Storage/transcoding interfaces arrive only with their actual use cases.

Queue metrics report observed pending/active/scheduled/retry/archived counts; absent queues are zero and inspection failures are `-1`. Real smoke caught that Asynq `GetQueueInfo` does not map missing queues to its public `ErrQueueNotFound`; listing queues before inspection resolves this without matching exception strings.

## Verification performed

The host Go is older; all Go checks used `golang:1.27.1-bookworm` with the backend bind mount and named Go module/build caches.

1. TDD at the agreed HTTP/observability/lifecycle seams: initial health test failed before implementation; standardized unknown-route/method tests failed on the previous text responses; metrics/log correlation test failed before middleware; graceful-drain test failed before lifecycle implementation. Each was followed by a passing implementation.
2. `gofmt -l cmd platform` — passed, empty output.
3. `go vet ./...` — passed.
4. `go test -race ./...` — passed. Tests cover health/readiness, safe 404/405/503 envelopes, server-generated request correlation, no raw path/query/cookie values in logs, Prometheus request counter, and in-flight response completion during shutdown.
5. `docker build --target api -t courses-api:foundation backend` — passed; builds all three static executables.
6. Equivalent Docker builds for `migrate` and `worker` targets — passed. Worker rebuilt successfully after the queue inspection correction.
7. Isolated Docker smoke using disposable PostgreSQL 17.5 tmpfs and Redis 8 containers: API readiness returned 200; unknown method returned JSON 405 with `Allow` and request ID; runtime UID was 10001.
8. Migrate container against that database returned exit 0 and `applied: 0` with the empty Phase 1 directory.
9. Pausing the disposable Redis caused API readiness to return safe JSON 503 within its two-second budget; unpausing returned readiness to 200.
10. Worker readiness/status returned 200; `/status` declared no handlers; metrics for absent `media` queue returned zero after the fix. `docker stop -t 40` completed with exit 0 and `shutdown complete` in JSON logs.

Disposable smoke containers/network were removed; repository data and unrelated containers were not touched. Images/build caches remain available for Compose build reuse.

## Remaining limits / next task / user decision

Full Compose routing/health checks belong to F05/F06 and are not marked passed by this record. No migration schema, Course repository integration, sqlc, OpenAPI, generated client, auth, upload, transcoding, playback or payments exist in F02. Transcoding duration/failure and active-playback metrics await working features rather than publishing fictitious zero-valued capabilities. OTLP exporter delivery to a collector was not integration-tested. Redis cluster/sentinel modes are not supported in this foundation.

Next smallest integration task: wire Compose API/worker/migrate health conditions, internal ports, Nginx routing and named media volume with the contracts above. No user decision required.
