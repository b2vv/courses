# Codex brief — O07 gap tests: Pay17 and Pay26 (integration)

Repository /Users/strelia/projects/courses, branch `feat/payment`. Read `AGENTS.md` and
`docs/payment/plan.md` (rows Pay17, Pay26). Writing the phase report row by row exposed two rows the
plan claims at `integration` level with NO integration test. Add them. Log: `work/codex-o07.log`.
Do **not** commit.

Both go in `backend/internal/orders/adapters/postgres/repository_integration_test.go` (the
`//go:build integration` file with the real Postgres), matching its existing style. Give any new
subtest its OWN id source if it draws from a shared id slice (the file already had a starvation bug
from the shared slice — see the `failedService` pattern added earlier; do the same).

## Pay17 — price change after purchase keeps access and the order's snapshot
Scenario (plan): a course bought at 500 (amount_minor=50000), then the author changes the price to
700. Expected: the buyer's access holds; the order row keeps amount_minor=50000 (the snapshot).
Write an integration subtest `t.Run("Pay17 ...")` that: seeds a paid course at 50000, StartPurchase
→ ApplyPaymentEvent(paid) so the grant exists and the order is paid at 50000; then raises the course
price (UPDATE courses SET price_minor=70000 — or via the authoring repo if that is how price-set
runs); then asserts (a) the active grant for (user,course) still exists, (b) the order row's
amount_minor is still 50000. State in your log how you change the price.

## Pay26 — grandfather covers N students and a repeat price-set does not duplicate
The existing grandfather integration seeds ONE progress student and never re-runs price-set. The
plan claims "all N get an active grant source=grandfather" and "repeat price-set no duplicate".
Extend (or add a sibling subtest): seed **two** students with progress on a free course; run
GrandfatherOnPriceSet; assert BOTH get an active `source=grandfather` grant (count=2). Then run
GrandfatherOnPriceSet **again** (same course) and assert the grant count is STILL 2 (the partial
UNIQUE index makes the repeat a no-op), and no error that would abort. State how the repeat behaves.

## Verify
Run with `DOCKER_CONFIG=/tmp/claude-501/dockercfg` exported:
`make integration` (or the orders package: a Go container `go test -tags integration ./internal/orders/...`).
If the sandbox denies Docker, say so and leave the tests in place — Courses 1 runs them. Do not
claim a gate you could not run. Report the changed lines and the result.

## Channels
If a row is actually already proven somewhere I missed, say so with the file:test instead of adding
a duplicate. If the plan's scenario can't be built (e.g. price-set path differs), argue it.
