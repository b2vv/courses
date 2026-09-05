SHELL := /bin/sh
GO_IMAGE := golang:1.27.1-bookworm
GO_RUN := docker run --rm -v "$(CURDIR)/backend:/work" -w /work -v courses-gomod:/go/pkg/mod -v courses-gobuild:/root/.cache/go-build $(GO_IMAGE)

.PHONY: env up down dev logs migrate seed seed-identities format format-check lint typecheck test integration build e2e smoke generate check-generated check
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
seed: env
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
smoke:
	sh infrastructure/tests/smoke.sh
generate:
	$(GO_RUN) go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.30.0 generate
	cd frontend && npm run generate
check-generated: generate
	git diff --exit-code -- backend/internal/course/adapters/postgres/generated backend/internal/identity/adapters/postgres/generated frontend/src/adapters/api/generated
	@test -z "$$(git ls-files --others --exclude-standard -- backend/internal/course/adapters/postgres/generated backend/internal/identity/adapters/postgres/generated frontend/src/adapters/api/generated)"
check: format-check lint typecheck test integration build e2e
