SHELL := /bin/sh
GO_IMAGE := golang:1.27.3-bookworm
# K12: the Go caches are named volumes locally; CI points them at host directories (absolute paths)
# that actions/cache restores and saves, so a runner does not rebuild the world on every run.
GO_MOD_CACHE ?= courses-gomod
GO_BUILD_CACHE ?= courses-gobuild
export GO_MOD_CACHE GO_BUILD_CACHE
GO_RUN := docker run --rm -e GOTOOLCHAIN=local -v "$(CURDIR)/backend:/work" -w /work -v $(GO_MOD_CACHE):/go/pkg/mod -v $(GO_BUILD_CACHE):/root/.cache/go-build $(GO_IMAGE)

.PHONY: env up down dev up-stub dev-stub logs migrate seed seed-identities drain-queue format format-check lint typecheck test integration build e2e e2e-local e2e-reset tunnel-up tunnel-down smoke generate check-generated check
env:
	@test -f .env || cp .env.example .env
up: env
	docker compose up -d --build --wait --wait-timeout 180
	@# nginx.conf is a bind mount, so an unchanged nginx container keeps serving
	@# the configuration it loaded at start. Reload it or edits are invisible.
	docker compose exec -T nginx nginx -s reload
down:
	docker compose down
dev: env
	docker compose -f docker-compose.yml -f docker-compose.dev.yml up -d --build --wait --wait-timeout 180
	docker compose exec -T nginx nginx -s reload
# up-stub / dev-stub add the offline WayForPay stub so a purchase settles locally without the real
# provider (docker-compose.stub.yml). `make down` tears either down. Not an authoritative gate.
up-stub: env
	docker compose -f docker-compose.yml -f docker-compose.stub.yml up -d --build --wait --wait-timeout 180
	docker compose -f docker-compose.yml -f docker-compose.stub.yml exec -T nginx nginx -s reload
dev-stub: env
	docker compose -f docker-compose.yml -f docker-compose.dev.yml -f docker-compose.stub.yml up -d --build --wait --wait-timeout 180
	docker compose -f docker-compose.yml -f docker-compose.dev.yml -f docker-compose.stub.yml exec -T nginx nginx -s reload
logs:
	docker compose logs -f --tail=100
migrate: env
	docker compose run --rm migrate
seed: seed-identities
	docker compose exec -T postgres sh -c 'psql -v ON_ERROR_STOP=1 -U "$$POSTGRES_USER" -d "$$POSTGRES_DB"' < backend/sql/seed.sql
seed-identities: env
	docker compose exec -T postgres sh -c 'psql -v ON_ERROR_STOP=1 -U "$$POSTGRES_USER" -d "$$POSTGRES_DB"' < backend/sql/seed-identities.sql
format:
	cd frontend && npm run format
	$(GO_RUN) gofmt -w .
format-check:
	cd frontend && npm run format:check
	$(GO_RUN) sh -c 'test -z "$$(gofmt -l .)"'
lint:
	cd frontend && npm run lint
	$(GO_RUN) go vet ./...
typecheck:
	cd frontend && npm run typecheck
test:
	cd frontend && npm test
	$(GO_RUN) go test -race ./...
integration:
	sh infrastructure/tests/integration.sh
build:
	docker compose build
# K4: every browser run, green or red, leaves the whole run's api+nginx log in frontend/test-logs/
# (gitignored via *.log). The R23 heartbeat 403 (O05) was never explained because the log was gone
# by the time anyone looked: recovery.spec stops/`up`s api without the e2e overlay, which RECREATES
# the container, and a dump taken after the run sees only the new one (2 api lines for a 15-minute
# suite, measured 2026-10-07). So the log is STREAMED from before the run: `docker compose logs -f`
# writes history to disk as it happens and re-attaches to a recreated container (verified on the
# live stack: force-recreate api → follower still alive, both "server starting" lines and the
# post-recreate requests in the file). It lives outside test-results/, which Playwright wipes on
# start. The exit status is the e2e run's own, read right after the subshell and re-raised after
# the follower is stopped, so a red run stays red.
# The suite's own `docker compose` calls (recovery.spec stop/up, media.setup exec, asr.spec) run bare,
# with no `-f`. Before this export they resolved to the BASE compose file and, finding a config that
# differs from the overlay-created container, RECREATED api/nginx/frontend on the base config mid-suite:
# frontend CSP flipped from report-only to enforcing, api lost the WayForPay stub env, nginx its login
# burst override — and the container log history went with them. COMPOSE_FILE makes those bare calls
# resolve to the same two files; the recipe's own explicit `-f` calls are unaffected (flags win).
# Measured 2026-10-07: `config --hash=api` is identical for the env form and the `-f` form, and a
# stop → up against a matching container keeps its id (base vs overlay differ → recreate). CI does
# not use this recipe: ci.yml sets its own job-level COMPOSE_FILE (base + e2e + ci override, K10).
# K10: E2E_SKIP_ASR is CI's flag (ci.yml sets it inline; the CI override starts no asr-worker). The
# local run always has the worker, so the flag is dropped here rather than trusted to be unset — a
# value leaked into a shell would otherwise skip asr.spec silently and the local gate would under-count.
# The paths are ABSOLUTE on purpose: COMPOSE_FILE is resolved against the caller's cwd and disables
# compose's parent-directory discovery, so a relative value broke every spec that shells compose
# from frontend/ (accounts, ratings: `stat frontend/docker-compose.yml`) — they had been finding
# ../docker-compose.yml by that discovery. With absolute paths the project directory is the first
# file's directory for every caller, so the project is "courses" from any cwd: a frontend-cwd
# `exec postgres` and the root stack are one project, the same container (measured: same id).
e2e:
	docker compose -f docker-compose.yml -f docker-compose.e2e.yml up -d --build --wait --wait-timeout 180 api frontend wayforpay-stub nginx
	export COMPOSE_FILE="$$(pwd)/docker-compose.yml:$$(pwd)/docker-compose.e2e.yml"; \
	mkdir -p frontend/test-logs; \
	stream="frontend/test-logs/api-nginx-$$(date -u +%Y%m%dT%H%M%SZ).log"; \
	docker compose -f docker-compose.yml -f docker-compose.e2e.yml logs -f --no-color --timestamps api nginx > "$$stream" 2>&1 & \
	follower=$$!; trap 'kill $$follower 2>/dev/null' EXIT; \
	(cd frontend && env -u E2E_SKIP_ASR npm run e2e); status=$$?; \
	kill $$follower 2>/dev/null; wait $$follower 2>/dev/null; trap - EXIT; \
	echo "api/nginx log (whole run): $$stream"; \
	exit $$status
# The stream file is run-stamped so a run never clobbers the artifact a reviewer was pointed at
# (it happened once: a second run truncated the file PR #9's counts were taken from). CI keeps the
# plain name api-nginx.log — its artifact is already per-run via upload-artifact.
# Run this, not bare `e2e`, for a standalone browser run. `e2e` raises the stack with the
# e2e overlay (api, frontend, nginx and the WayForPay stub, `--wait` so chromium never meets a
# still-starting 503) but does NOT reseed, so on its own it fails against a `delete-me` user a
# previous run consumed - and because chromium gates media and playback, one consumed fixture
# stops the playback block from running at all. `e2e-local` adds the reseed. The reseed rule was
# written down and then bypassed within an hour, because the cheap path was the wrong one. This
# makes the cheap path correct by construction instead of by memory.
e2e-local: e2e-reset up seed-identities seed e2e
# K13 (owner's decision 2026-10-08): the authoritative runs — `make check` and `make e2e-local` —
# start from empty data volumes. The e2e database persisted between runs, so an interrupted run left
# published fixtures behind and count assertions became order-dependent (L8: recovery:83 had to be
# scoped to the seeded slugs; leaked e2e-ratings-* courses needed a manual `down -v`). Local dev data
# lives in the same compose project, so it is wiped too — by decision, not by accident; `make e2e`
# alone does not reset and stays the quick rerun against a live stack. prometheus-data is kept:
# metrics history is no test input. The volumes are found by compose labels, not by guessed names,
# and the target fails if any of them survived (a container still holding one makes `rm` refuse).
# The list is explicit on purpose, not "every project volume except prometheus": a future volume is
# as likely to be a cache someone wants kept (a model, a build cache) as test state, and an inverted
# rule would wipe it silently. Whoever adds a stateful volume that tests write to adds it here.
E2E_RESET_VOLUMES := postgres-data redis-data media-data
e2e-reset:
	docker compose -f docker-compose.yml -f docker-compose.e2e.yml down --remove-orphans
	@project=$${COMPOSE_PROJECT_NAME:-$$(basename "$$PWD")}; \
	for volume in $(E2E_RESET_VOLUMES); do \
	  ids=$$(docker volume ls -q --filter label=com.docker.compose.project=$$project --filter label=com.docker.compose.volume=$$volume); \
	  if [ -n "$$ids" ]; then docker volume rm $$ids > /dev/null || exit 1; fi; \
	  left=$$(docker volume ls -q --filter label=com.docker.compose.project=$$project --filter label=com.docker.compose.volume=$$volume); \
	  if [ -n "$$left" ]; then echo "e2e-reset: $$volume survived: $$left" >&2; exit 1; fi; \
	done; \
	echo "e2e-reset: project $$project, removed $(E2E_RESET_VOLUMES) (prometheus-data kept)"
# Owner-run live WayForPay proof (Pay25-live), work/archive/runbook-o06c-tunnel.md. Needs only
# WAYFORPAY_MERCHANT_ACCOUNT and WAYFORPAY_SECRET_KEY in .env; opens a public quick tunnel to this
# stack (seeded demo accounts included), so keep it short-lived and run tunnel-down after.
tunnel-up: env
	sh infrastructure/tunnel/up.sh
tunnel-down:
	sh infrastructure/tunnel/down.sh
# Backlog B-5: Redis runs with `appendonly yes` on the persistent `redis-data` volume and
# nothing drained the asynq queue between runs, so a run aborted mid-transcode left its task
# behind. The next `up` starts the worker, `Recover` takes that task, and it eats the 2 CPU /
# 2 GB budget while this run's transcode waits behind it. The e2e drains cannot cover it: they
# live inside `e2e`, and the chain is `up seed-identities seed smoke e2e`, so nothing drained
# before `smoke` ever.
#
# The counts are printed rather than assumed, because this repository has already been bitten
# by the pair: a drain that silently did not drain (a `VAR=cmd` prefix reaches only the first
# command of a pipe, so the second `redis-cli` got NOAUTH) next to a check that reported
# "empty" whatever the state - the five `asynq:{default}:*` counters, for a queue named `media`,
# which return zero on a full queue just as confidently. `export` is load-bearing, the pattern
# is scoped to the worker's own queue so that `asynq:servers` and `asynq:workers` registration
# survives, and the before/after numbers make a broken drain visible instead of green.
#
# PRECONDITION, without which this is destructive: heavy gates in this repository run one
# session at a time (agreed between sessions 2026-09-08). There is no concurrent run whose
# legitimate task this could remove. If that changes, remove this step first - it cannot tell a
# stale task from a live one.
drain-queue: env
	@docker compose exec -T redis sh -c 'export REDISCLI_AUTH="$$REDIS_PASSWORD"; 		before=$$(redis-cli --scan --pattern "asynq:{media}:*" | wc -l); 		redis-cli --scan --pattern "asynq:{media}:*" | xargs -r -n 100 redis-cli DEL > /dev/null; 		after=$$(redis-cli --scan --pattern "asynq:{media}:*" | wc -l); 		echo "drain-queue: asynq:{media}:* keys before=$$before after=$$after"; 		test "$$after" -eq 0; 		before=$$(redis-cli --scan --pattern "asynq:{asr}:*" | wc -l); 		redis-cli --scan --pattern "asynq:{asr}:*" | xargs -r -n 100 redis-cli DEL > /dev/null; 		after=$$(redis-cli --scan --pattern "asynq:{asr}:*" | wc -l); 		echo "drain-queue: asynq:{asr}:* keys before=$$before after=$$after"; 		test "$$after" -eq 0'
smoke:
	sh infrastructure/tests/smoke.sh
generate:
	$(GO_RUN) go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.30.0 generate
	$(GO_RUN) go run ./cmd/openapi-gen
	cd frontend && npm run generate
check-generated: generate
	git diff --exit-code -- backend/internal/course/adapters/postgres/generated backend/internal/identity/adapters/postgres/generated backend/openapi/openapi.yaml frontend/src/adapters/api/generated
	@test -z "$$(git ls-files --others --exclude-standard -- backend/internal/course/adapters/postgres/generated backend/internal/identity/adapters/postgres/generated backend/openapi/openapi.yaml frontend/src/adapters/api/generated)"
check: format-check lint typecheck test integration check-generated build e2e-reset up drain-queue seed-identities seed smoke e2e
