SHELL := /bin/sh
GO_IMAGE := golang:1.27.1-bookworm
GO_RUN := docker run --rm -v "$(CURDIR)/backend:/work" -w /work -v courses-gomod:/go/pkg/mod -v courses-gobuild:/root/.cache/go-build $(GO_IMAGE)

.PHONY: env up down dev logs migrate seed seed-identities format format-check lint typecheck test integration build e2e e2e-local smoke generate check-generated check
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
e2e:
	cd frontend && npm run e2e
# Run this, not bare `e2e`, for a standalone browser run. `e2e` is the low-level target the
# `check` chain composes; on its own it neither raises the stack nor restores the disposable
# fixtures, so it fails against an api that is still starting (503) or a `delete-me` user a
# previous run consumed - and because chromium gates media and playback, one consumed fixture
# stops the playback block from running at all. Both rules were written down and both were
# then bypassed within an hour, because the cheap path was the wrong one. This makes the
# cheap path correct by construction instead of by memory.
e2e-local: up seed-identities seed e2e
smoke:
	sh infrastructure/tests/smoke.sh
generate:
	$(GO_RUN) go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.30.0 generate
	cd frontend && npm run generate
check-generated: generate
	git diff --exit-code -- backend/internal/course/adapters/postgres/generated backend/internal/identity/adapters/postgres/generated frontend/src/adapters/api/generated
	@test -z "$$(git ls-files --others --exclude-standard -- backend/internal/course/adapters/postgres/generated backend/internal/identity/adapters/postgres/generated frontend/src/adapters/api/generated)"
check: format-check lint typecheck test integration check-generated build up seed-identities seed smoke e2e
