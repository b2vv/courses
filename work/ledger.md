# Foundation execution ledger

Записи append-only за змістом: виправлення робляться новим записом із посиланням на попередній. Стан задач — [tasks.md](tasks.md); рішення — [assumptions](../docs/foundation/assumptions.md). `planned` commands не є evidence.

| Date | Task / owner | State / decision | Evidence / next action |
|---|---|---|---|
| 2026-09-03 | F01 / parent | Started repository inspection and scope framing | Initial cwd `/Users/strelia/projects/courses`; repository initially empty |
| 2026-09-03 | F01 / foundation_plan | Read spec-flow and acceptance-spec; authored assigned planning documents only | `docs/foundation/spec.md`, `plan.md`, `assumptions.md`, `work/tasks.md`, `work/ledger.md` |
| 2026-09-03 | F01 / foundation_plan | Scope limited to Phase 1 → Phase 2. User supplied explicit decisions; no additional product confirmation required | Foundation DoD includes Course slice; Phase 3–8 deferred |
| 2026-09-03 | F01 / foundation_plan | Plan self-defense has five covered rows; independent verdict pending | Fresh reviewer must run plan-critique + plan-defense before implementation; no claim of green gate |

## Pending gates (historical snapshot before execution)

- F01 independent plan-critique: pending.
- F01 independent plan-defense: pending.
- F01 manifest/constitution review: parent-owned, pending.
- F06 Phase 1 runtime result: not started.
- C05 full acceptance/CI pipeline: not started.

## Report after each stage

Record: (1) implemented behavior, (2) actual files/modules, (3) decisions with ADR links, (4) commands run, (5) pass/fail evidence, (6) remaining limits, (7) next smallest task, (8) whether a user decision is needed. Keep environment failures distinct from product failures, and both distinct from checks never run.

2026-09-03 plan-critique: logic/slop/edges returned. Fixed 4 blocking findings in plan Phase1 shell/recovery section (gate cycle, retry, migration dependency, pagination bounds). Atomic concurrent seed added. Full UI inventory and cookie forwarding retained because explicitly required by user.
2026-09-03 plan_complexity returned clean (0/0). Critique total: 4 launched / 4 returned, 4 blocking fixed; 3 nonblocking addressed or explicitly retained. plan_defense follow-up: added full query keys, active route selection, obsolete abort and reversed A/B response check.
2026-09-03 F01 gates GREEN: 4/4 critique returns, defense0holes (plan_defense). F02 delegated backend_runtime; F03 parent; F04 delegated ui_kit. All are Phase1 shell/infrastructure scope.
2026-09-03 F02–F06 GREEN: commit `5608a03`; runtime, SSR shell, React Aria kit, Compose/Nginx and Phase 1 smoke verified.
2026-09-03 C01–C02 GREEN: commit `31e8c39`; Course domain/application/PostgreSQL and OpenAPI/HTTP contract verified with race and Testcontainers tests.
2026-09-03 C03–C05 GREEN: commit `4d73072`; Orval/API adapter, request-scoped Query/SSR, localized storefront and recovery flows verified. Independent code-review findings fixed: raw query parser, pending-loader cancellation, generated drift gate, Nginx raw error logging. Final: 53 Rstest, Go race, integration, production builds, 14 Playwright and smoke PASS. Next: Phase 3 Identity specification; no user decision required.
2026-09-03 I03 GREEN: Identity HTTP transport, cookies and runtime security config. Ordered gates PASS (format-check, lint, typecheck, 53 Rstest + Go race, Testcontainers integration, generated drift, build, smoke, 14 Playwright). Live curl matrix against Nginx covers origin, body, credential, cookie, rotation and logout paths. Defect found only by live verification: `time.Now` nanosecond precision could not survive the PostgreSQL `timestamptz` round trip, so every correct login returned 503 while committing an orphan session row; fixed by normalizing the clock reading to microsecond resolution in `application.Identity.timestamp`, guarded by a red/green application test and a PostgreSQL resolution test. Next: I04 OpenAPI/Orval and Nginx auth edge; no user decision required.
