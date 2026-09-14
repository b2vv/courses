# Working in courses

## Scope and delivery
- Read the repository briefing first: [work/CTO-RESEARCH.md](work/CTO-RESEARCH.md). It caches primary sources; where it disagrees with the code, the code wins and the briefing is rewritten.
- Foundation, public catalog, identity, admin course management, media upload/transcoding (Phases 1–5) and playback (Phase 6) are complete. Phase 6 closed on 2026-09-14 by its own acceptance rows, not only by its tasks: all 102 rows of `docs/playback/plan.md` §8 are verified in `docs/playback/report.md` §2 behind an ordered `make check` (e2e 65/65, 48 race-enabled Go packages). Its named limitations — native HLS unproven without real Safari, ownership measured on a compressed cadence, B25 contrast over an empty `<video>` — stay in that report's §6 and may not be claimed away. No next phase has started.

  Two lessons from closing that set apply to any phase. A row's proof level is an assumption about where the proof lives until it is checked against the fixtures: six Phase 6 rows had to move to `e2e` by a dated plan amendment because no `smoke` or integration fixture could produce what they asked for. And a row filed as missing evidence can be a defect: B07 and B31 were both product gaps, fixed before they were verified. A green aggregate gate is not evidence for a row; a phase is closed when its rows are.
- Playback, the learning outline, watched marks and student progress work and are measured — do not undersell them either. What does not exist is payments, comment threads, transcripts, an instructor workspace and a production deployment; none of those may be claimed. Phase 7 and later remain roadmap work. Claim only what the phase report records as measured. The phase roadmap lives in `work/PRODUCT-BRIEF.md` §21/§24. `work/AGENDA.md` and `work/CTO-RESEARCH.md` were rebuilt 2026-09-12 on `b913e81`, before Phase 6 closed; both still describe it as open, so rebuild them before choosing the next move from them.
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
