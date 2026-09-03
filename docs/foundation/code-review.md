# Two-axis implementation review

Fixed point: `aac6961361efbe476710aae05502502f9a6c563b`. Review covered tracked changes and the then-untracked Phase 2 implementation. Standards source: `AGENTS.md`, `.claude/standards.md`, architecture docs. Specification source: `docs/foundation/spec.md` and `plan.md`.

## Standards

Reviewer found one hard violation: Nginx `error_log` used the fixed raw format, which can include full URI and Referer even though access logs used a safe format. This violated the rule that logs must not contain query strings or secrets. Fixed in `infrastructure/nginx/nginx.conf`; the recovery test generates a marker in query and Referer for rate-limit and upstream-failure paths and confirms it is absent from persisted logs. High-impact baseline code smells: 0.

## Spec

Reviewer found three partial requirements:

1. Default Router parsing coerced literal search values (`2026`, `true`, quoted text) and accepted numeric exponents. Fixed with raw string parsing, duplicate preservation and validation regression tests.
2. Pending route loaders did not bind Router cancellation to Query ownership. Fixed with an observer-scoped `ensureRouteQuery`; tests cover pending navigation cancellation, shared reads and fresh retry signals.
3. CI regenerated neither sqlc nor Orval before drift comparison. Fixed with `make check-generated` in CI and verified after commit.

The reviewer also marked E03, E10 and failure→retry as pending evidence. `frontend/e2e/recovery.spec.ts` now executes all three and passed. Missing/partial findings after fixes: Standards 0; Spec 0. Phase 3–8 was correctly excluded from this review.
