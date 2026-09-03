# Courses

Self-hosted online-course platform foundation. Current scope: a public catalog vertical slice, with infrastructure for later identity, administration, private HLS and transcripts. No payment provider or automatic transcription is connected.

Stack: React 19, strict TypeScript, TanStack Start/Router/Query, Rsbuild/Rspack, Tailwind 4, React Aria; Go 1.27.1 modular monolith, pgx/sqlc/Goose, PostgreSQL 17, Redis/Asynq, FFmpeg worker image and Nginx. Rstest, Testing Library, MSW, Playwright, Oxlint/Oxfmt and Go/Testcontainers supply checks.

## Start

Prerequisites: Docker with Compose v2+, 4 GB available RAM; Node 24 for native frontend tools. Go runs in a pinned container. Local placeholders in `.env.example` are development-only.

```sh
make up
make seed
```

`make up` copies `.env.example` if needed, builds, applies migrations and waits for health. Sample data is explicit and repeatable. `make down` preserves volumes.

URLs: [home](http://localhost:8080), [catalog](http://localhost:8080/courses), [API](http://localhost:8080/api/v1/courses), [health](http://localhost:8080/healthz). Change HTTP_PORT in `.env` when needed. Base Compose exposes only Nginx on loopback. `make dev` enables frontend source watching and loopback database/API ports.

## Development checks

```sh
cd frontend
npm ci
npm exec playwright install chromium
cd ..
make format-check lint typecheck test
make integration
make build
make up
make seed
make e2e
```

`make migrate` applies Goose migrations. `make generate` regenerates sqlc and Orval; generated code belongs to adapters, never domain. `make format` formats code. CI uses format → lint → typecheck → unit → integration → build → E2E. See [testing](docs/development/testing.md) and the factual [report](docs/foundation/report.md).

## Repository

```text
frontend/          React application, API adapter, UI kit, browser tests
backend/           Go executables, business modules, platform adapters, SQL/OpenAPI
infrastructure/    Nginx, monitoring and smoke tests
docs/              Architecture, ADRs, setup, foundation plan and evidence
work/              Task states and completion pointers
```

## Foundation limitations

Identity/login, admin CRUD, upload/transcoding, signed playback, synchronized transcript and orders are future phases. Locked routes do not simulate authentication. `/media` is deliberately forbidden; the worker starts but does not process transcoding tasks. Domain vocabulary and planned storage/transcription seams are documented before implementation. Advanced UI/player/editor dependencies are installed only with real consumers.

No DRM is claimed. Before public deployment add TLS, managed secrets, backups/restores, storage quotas and the remaining authorization gates. Optional monitoring starts with `docker compose --profile monitoring up -d`; metrics stay internal. Follow [local setup](docs/development/local-setup.md), [troubleshooting](docs/development/troubleshooting.md), [architecture](docs/architecture/context.md) and [ADR decisions](docs/adr/001-modular-monolith.md).
