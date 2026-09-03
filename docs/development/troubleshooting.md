# Troubleshooting

- Docker unavailable: start Docker Desktop/Engine and run `docker info`. Integration checks require daemon access.
- Port 8080 busy: set `HTTP_PORT` in `.env`, restart Compose, and set `E2E_BASE_URL` for browser tests.
- API unready: inspect `docker compose logs api migrate postgres`; verify credentials agree. Do not delete volumes to repair a configuration mismatch.
- Empty catalog: fresh schema intentionally contains no sample data; `make seed` adds them.
- 403 under `/media`: expected until Phase 6 authorization is implemented.
- Catalog error state: check API health/readiness; SSR does not mask dependency failures as empty data.
- Dependency downloads blocked: allow registry access through the normal environment permission flow; do not substitute unverified generated files.
- Shutdown retains data. `docker compose down -v` destroys it and is never part of routine setup/testing.
