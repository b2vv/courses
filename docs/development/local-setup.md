# Local setup

Requires Docker Engine/Desktop with Compose v2+ and at least 4 GB available RAM. Native frontend development uses Node 24; backend tooling uses the pinned Go container so a global Go install is optional.

Run `make up` from the repository root. It creates `.env` from `.env.example` if absent, builds containers and runs schema migrations. Run `make seed` explicitly for sample catalog entries. Open http://localhost:8080. `make down` preserves named volumes.

Use `make dev` for development port bindings and source watching. `.env` is local-only; replace development credentials before any nonlocal deployment. See root README and Makefile for exact commands.
