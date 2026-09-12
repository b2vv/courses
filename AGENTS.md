# Working in courses

## Scope and delivery
- Read the repository briefing first: [work/CTO-RESEARCH.md](work/CTO-RESEARCH.md). It caches primary sources; where it disagrees with the code, the code wins and the briefing is rewritten.
- Foundation, public catalog, identity, admin course management and media upload/transcoding (Phases 1–5) are complete. Current delivery: playback (Phase 6). All thirteen tasks T1–T13 are done and committed behind a green ordered `make check` (e2e 60/60, 48 race-enabled Go packages), **but the phase is not closed by its own acceptance rows**: 89 of 102 rows in `docs/playback/plan.md` §8 have evidence and thirteen do not. Seven are undone work — B07, B08, B23, B31, B36, R30, R34 — provable at the level the plan assigns them. Six are not. B20, B21, B22, B28 and R23 need an allowed media segment with real bytes, and `backend/sql/seed.sql` inserts courses only, so the stack `smoke` runs against has no playable content to produce one; B21a must read a **generated** `index.m3u8`, and no integration test runs ffmpeg, so that level cannot produce one either — the unit test pins the argv that causes the property, not the playlist the row asks to read. Their level was an assumption about where the proof lives, never checked against the fixtures; `docs/playback/plan.md` §8 carries the level each of them is assigned now, and the dated amendment that set it. `docs/playback/report.md` records what each row is missing. A green aggregate gate is not evidence for a row, and the phase may not be called complete while these are open.
- Playback, the learning outline, watched marks and student progress work and are measured — do not undersell them either. What does not exist is payments, comment threads, transcripts, an instructor workspace and a production deployment; none of those may be claimed. Phase 7 and later remain roadmap work. Claim only what the phase report records as measured. The phase roadmap lives in `work/PRODUCT-BRIEF.md` §21/§24. The current move ranking is `work/AGENDA.md`, rebuilt 2026-09-12 on the briefing at `b913e81`.
- Read the current phase's `docs/<phase>/plan.md` and `work/tasks.md`; execute one logical slice at a time.
- Preserve existing work. Never commit secrets, `.env`, media bytes or dependency caches.
- Use small logical commits and record current-phase verification, limitations and next task in the current phase's `docs/<phase>/report.md`.

## Architecture
- Backend modular monolith with a separate media-worker executable. No Rust, ORM or service locator.
- Domain: standard language only; no framework, database, transport or external packages.
- Application: domain and ports only. Ports may reference domain. Adapters point inward.
- Framework wiring belongs to composition roots. Handlers/routes translate transport and render; business policies live inward.
- Generated sqlc/Orval models stay in adapters and map explicitly to domain entities.
- Ports require an actual external dependency. Document future seams before adding unused interfaces.
- No global mutable clients/stores. SSR dependencies and Query caches are request scoped.
- Nginx serves media bytes; Go and SSR never proxy them. Bytes are reachable only under `/media/p/<playbackId>/`, and only through an `auth_request` subrequest the playback module answers per request; every other path under `/media/` is refused outright. The transcoder writes the tree `0750`/`0640`, so nginx reads it by group membership rather than by the tree being world-readable.

## TypeScript
- Strict TypeScript; no `any` or assertions that disguise unsafe input. Validate external data.
- TanStack Start + Rsbuild/Rspack, React Aria, Query for server state, Router search params for URL state.
- No Vite, Next, Vitest, Jest, ESLint, Axios, Redux Toolkit, Radix or shadcn.
- No fetch in presentation/routes, no server state in Zustand. Add dependencies only when used.
- Semantic design tokens; keyboard navigation, visible focus, >=44px controls, light/dark, 200% text zoom.

## Verification
- `make format-check`, `make lint`, `make typecheck`, `make test`, `make integration`, `make build`, `make e2e` in that order.
- Unit tests at domain/application boundaries; repository tests against isolated Postgres; contract tests at HTTP; browser tests against Compose.
- Do not represent skipped or blocked checks as passed. CI fails if Docker integration cannot run.
- Explicit business error codes; no secrets in logs, URLs logged without query strings.
