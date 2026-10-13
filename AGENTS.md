# Working in courses

## Scope and delivery
- Read the repository briefing first: [work/CTO-RESEARCH.md](work/CTO-RESEARCH.md). It caches primary sources; where it disagrees with the code, the code wins and the briefing is rewritten.
- Foundation, public catalog, identity, admin course management, media upload/transcoding (Phases 1–5), playback (Phase 6), the transcript (Phase 7a), speech recognition (Phase 8), payment readiness (Phase 9) and comment threads (Phase 7b) are complete. Phase 6 closed on 2026-09-14 at 102/102 rows (`docs/playback/report.md`). Phase 7a passed its ordered gate on 2026-09-18, but its 87 rows were never counted row by row (`docs/transcript/report.md`). Phase 8 closed on 2026-09-24 at 22/22 rows behind an ordered `make check` (482 frontend tests, 34 Go packages, e2e 74/74): `docs/asr/report.md`. Its named limitations stay in that report's §7, and so do Phase 6's in its §6; neither may be claimed away. Phase 9 closed on 2026-09-28 at R01–R12 (`docs/readiness/report.md`); Phase 7b closed on 2026-09-29 at K01–K23 behind an ordered `make check` (backend 501 tests, e2e 79/79): `docs/discussion/report.md`, limitations in its §5. The instructor workspace closed on 2026-09-29 at W01–W26 behind an ordered `make check` (frontend 505, e2e 87/87), extending the discussion module with course-wide moderation, reports and an audit log: `docs/instructor/report.md`, limitations in its §5. Payment (the last roadmap phase, §24 item 21) closed on 2026-10-03: the WayForPay sandbox flow was proven live (a real 1-UAH Approved transaction, callback `POST→200`, applied exactly once), and PR #3 merged (`df5b380`); `docs/payment/report.md`, limitations in §5. The ratings-and-reviews phase closed on 2026-10-04 behind an ordered `make check` (e2e 98 passed) at Rt01–Rt30, merged as PR #5 (`14aeb3b`): `docs/ratings/report.md`, limitations in its §5. Its e2e surfaced four real defects behind the green gate (a strict schema missing rating/lessonCount, an `""` cursor, a CSS-less RadioGroup swallowing clicks, a stale list after a write), all closed. Tails B-28 (null last-page review cursor) and B-23 (stale restore snapshot) closed 2026-10-04 as PR #6 (`83c5142`). The self-service **accounts** phase closed on 2026-10-06 behind an ordered `make check` (e2e 101 passed) at Acc01–Acc31, merged as PR #7 (`d263045`): public registration, email verification, self-service password reset, change own password/email, optional TOTP 2FA, a login rate limit and remember-me, over a Mailpit email foundation and an enforcing production CSP; `docs/accounts/report.md`, limitations in its §5. Its gate surfaced real defects behind earlier green runs (reset/email-change emails pointed at non-existent routes → 404, a decorative checkbox mark swallowed clicks, cross-module integration fixtures needed verified subjects), all closed. The **enrollment** phase closed on 2026-10-07 behind an ordered `make check` (e2e 107 passed) at En01–En15, merged as PR #8 (`9c0e64f`): course membership as the one authoritative access fact — exactly one `course_grants` row per `(user, course)` whose state (`active`/`blocked`/`revoked`) is the power to consume and communicate, free content now requiring membership; self-enrol, instructor block/unblock, admin revoke, roster with progress, person reports with an admin queue, `CanConsume`/`CanCommunicate` contract predicates, the media-bytes gate proven at nginx `auth_request`, and a one-shot backfill (migration 00015) proven on populated data; `docs/enrollment/report.md`, limitations in its §6. Its review (slice by slice) and gates caught real defects before merge (a revoked-as-member UI branch, an unnamed comms-gate error, a dead revoke control, a backfill proof-gap, an admin-nav contrast). **The roadmap is complete; no phase is open.** Production deployment was deferred by the owner on 2026-10-03 (test mode, no prod server) and is a human decision, not a queued phase (`work/AGENDA.md` Р1).

  Lessons from closing phases, for any phase. A row's proof level is an assumption about where the proof lives until it is checked against the fixtures. A row filed as missing evidence can be a defect: Phase 6 found two, and Phase 8 ten, all of which had passed a green gate. A green aggregate gate is not evidence for a row; a phase is closed when its rows are. Evidence counts only when it took the production path — a fake, a directory stat or a hand-picked fixture proves the fake. A change to shared or persisted state is checked with the full e2e set, not a targeted one: review of the diff misses breaks outside the diff.
- Playback, the learning outline, watched marks, student progress, the transcript, speech-recognition drafts, comment threads, the instructor moderation workspace, WayForPay payments and orders, ratings-and-reviews, self-service accounts (registration, email verification, password reset, change password/email, TOTP 2FA), enrollment (course membership: self-enrol, block/unblock/revoke, roster, person reports, content and communication gates), the instructor/admin learner overview and author-block-from-queue (learning phase, 2026-10-09), and the admin analytics dashboard (`/admin/analytics`: platform snapshot, instructor leaderboard, engagement, metric trend) plus the users roster with role/sort filters (admin-reconstruction + admin-stats, 2026-10-11, owner-initiated outside §24) work and are measured — do not undersell them either. Payments and orders exist as a sandbox-proven flow (a real 1-UAH Approved transaction), not as a production deployment; email is caught by Mailpit in test mode, not delivered externally. What does not exist is a production deployment; it may not be claimed. Per-course author blocking and the cross-course learning overview, both deferred from enrollment (`docs/enrollment/report.md` §6), shipped in the learning phase (`docs/learning/report.md`). **Revoke stays admin-only by the owner's decision (2026-10-13): a lecturer acts through block (communication) + report-on-learner to the admin queue, and the admin decides on revoke — a deliberate boundary (`docs/enrollment/report.md` §4), not a gap.** An earlier AGENDA rebuild (`00c7e4c`) had mislabelled "lecturer-initiated revoke" as an open owner question with false citations to learning §5 / enrollment §6, where it does not appear; do not resurrect it. Paid re-enrolment after an admin revoke is refused by design (`course_already_owned`). Claim only what the phase report records as measured. The phase roadmap lives in `work/PRODUCT-BRIEF.md` §21/§24. `work/CTO-RESEARCH.md` and `work/AGENDA.md` were rebuilt on 2026-10-11 after the learning and admin phases; rebuild them when they fall behind the code.
- Read the current phase's `docs/<phase>/plan.md` and `work/tasks.md`; execute one logical slice at a time.
- Preserve existing work. Never commit secrets, `.env`, media bytes or dependency caches.
- Use small logical commits and record current-phase verification, limitations and next task in the current phase's `docs/<phase>/report.md`.

## Architecture
- Backend modular monolith with a separate media-worker executable. No Rust, ORM or service locator.
- Domain: standard language and `internal/contracts/…` only; no framework, database, transport or external packages.
- Application: domain and ports only. Ports may reference domain. Adapters point inward.
- Framework wiring belongs to composition roots. Handlers/routes translate transport and render; business policies live inward.
- Generated sqlc/Orval models stay in adapters and map explicitly to domain entities.
- Ports require an actual external dependency. Document future seams before adding unused interfaces.
- Cross-module rules live in `internal/contracts/…` as domain-safe types and pure functions only: standard library, no I/O, no imports of any module (ADR 013). A domain may import a contract; a module never imports another module's packages. Each module reads its own facts and maps them into the contract. Anything larger than a pure predicate does not belong in `contracts`.
- No global mutable clients/stores. SSR dependencies and Query caches are request scoped.
- Nginx serves media bytes; Go and SSR never proxy them. Bytes are reachable only under `/media/p/<playbackId>/`, and only through an `auth_request` subrequest the playback module answers per request; every other path under `/media/` is refused outright. The transcoder writes the tree `0750`/`0640`, so nginx reads it by group membership rather than by the tree being world-readable.

## TypeScript
- Strict TypeScript; no `any` or assertions that disguise unsafe input. Validate external data.
- TanStack Start + Rsbuild/Rspack, React Aria, Query for server state, Router search params for URL state.
- No Vite, Next, Vitest, Jest, ESLint, Axios, Redux Toolkit, Radix or shadcn.
- No fetch in presentation/routes, no server state in Zustand. Add dependencies only when used.
- Semantic design tokens; keyboard navigation, visible focus, >=44px controls, light/dark, 200% text zoom.

## Decisions

Added by the repository owner's decision, 2026-09-18, after Phase 7a, where four defects of
these shapes reached the tree and two of
them turned the full ordered gate red. Each rule costs a sentence or two, and only on decisions
that close an alternative — not on every commit.

- **A limitation that closes an alternative carries either a magnitude or a mechanism.** A
  magnitude means a number with its unit and the method that produced it, so a reader can
  repeat it; a mechanism means something a reviewer can try to build. This covers "we will not
  do X" and "X can only be done this way" alike — both foreclose. A sentence with neither is a
  hypothesis in the costume of a limitation, and it cancels no work. The cheapest refutation is
  almost always cheaper than the work being avoided: "this needs a five-minute transcode" cost
  a journey until someone ran `ls` on the fixtures and found a two-second one.
- **Whoever adds an instance of something globally counted cleans up after itself; whoever
  reads such a thing scopes its selection** instead of asserting a name at a position. Published
  courses, Redis keys, metric series, `role="status"` on a page — all are counted or looked up
  by uniqueness somewhere. A change can be locally correct and still break a relationship no
  one holds, and neither review nor attention catches that: the reviewer shares the author's
  blind spot, which is the edge of the diff. Only a gate that runs to the end does not share it.
- **A decision that closes an alternative is an artefact:** it carries a date, an author and a
  named way to check it, whether it is taken in the plan or in the middle of a slice. Written
  when it is taken, not in the retrospective — by then it has been lived rather than reviewed.
  A decision that exists only as work not done is not a decision; it was skipped, and nobody can
  open it later because there is nothing to open.

## Verification
- `make format-check`, `make lint`, `make typecheck`, `make test`, `make integration`, `make build`, `make e2e` in that order.
- Unit tests at domain/application boundaries; repository tests against isolated Postgres; contract tests at HTTP; browser tests against Compose.
- Do not represent skipped or blocked checks as passed. CI fails if Docker integration cannot run.
- Explicit business error codes; no secrets in logs, URLs logged without query strings.
