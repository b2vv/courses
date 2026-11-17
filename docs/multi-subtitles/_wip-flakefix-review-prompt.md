# Review: package `flakefix-w1` (gate flakes S5 and B18, on the PR #109 branch)

Independent reviewer, read-only: no edits, no git writes, no Docker. You may run `cd frontend && npm run typecheck` and, if your sandbox
allows, `npx rstest run src/routes/route-guards.test.tsx`.

- Context: the last `flakefix` section of `docs/multi-subtitles/report.md` and the attempt 8 and 9 entries above it. The owner asked to fix
  the flakes before merging PR #109.
- Delta: the uncommitted diff of `frontend/src/routes/admin.index.tsx`, `frontend/src/routes/route-guards.test.tsx` and
  `frontend/e2e/playback.spec.ts`. Frozen copy `docs/multi-subtitles/_wip-flakefix.patch` (`bd927dbd…`). Branch @ `ea1fe47`. Other dirty files are
  out of scope.
- B18 evidence: the attempt-8 Playwright trace, timeline in `docs/multi-subtitles/_wip-flakefix-b18-timeline.md` (the raw trace stays out of the repo: it holds a session cookie and media bytes;
  `0-trace.trace`). Verify the claim yourself: the last media segment request falls before `setOffline`, and there is no media request after it.

Check:
1. **S5 product fix.** Does `navigate({ search: (previous) => … })` merge over the router's latest location in TanStack Router (check the
   installed version's `buildLocation` / functional search semantics in `frontend/node_modules/@tanstack/router-core`)? Is
   `adminListState(previous)` the right normalisation? Does anything else in `go()`'s callers (`changePage`, `clearFilters`,
   `changeStatus`) change behaviour?
2. **S5 test.** Does it discriminate? The author reports that the old `admin.index.tsx` fails it with the gate's exact symptom. Is setting
   `router.options.defaultPendingMs` after creation sound, or does it test something the browser cannot do?
3. **B18.** Is the diagnosis right? Does the new break proof (`fetch(location, { cache: 'no-store' })` must fail) keep the row's claim honest?
   Could the fetch be served from a service worker or cache? Does the row still prove that the position survives the break?
4. Are there other rows with B18's shape: waiting offline for the player's unavailable state?
5. Standards: `AGENTS.md`, `.claude/standards.md`.

Finding format: `file:line`, impact, how verified, BLOCKER or non-blocking, confidence. End with `VERDICT (flakefix-w1): SHIP | BLOCK` and counts.
