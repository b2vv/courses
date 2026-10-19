# Codex brief — O06c-auto: WayForPay stub + offline purchase e2e

Repository /Users/strelia/projects/courses, branch `feat/payment`. **Read first, they bind:**
`AGENTS.md`, `docs/payment/plan.md` (Pay24, Pay25, §Зрізи), `docs/payment/assumptions.md`
(the two O06 entries, esp. "O06c split" 2026-10-01). Backend O06a (`aeff91c`) and frontend
O06b (`3b7213e`) are done — build on them, don't change their production behaviour. Log:
`work/codex-o06c.log` (first line a `NEXT:`). Do **not** commit. The gate's `make e2e` IS this
task's target — run it.

## Goal
Prove the purchase path end to end **offline**, against a local WayForPay stub, in the e2e stack.
No network to WayForPay. The live tunnel run (Pay25 with the real callback) is NOT yours — a human
runs it from a runbook; do not attempt it.

## Seam (decided; implement exactly)
1. The WayForPay adapter's `PayURL` and `StatusURL` (currently constants in
   `backend/internal/orders/adapters/wayforpay/wayforpay.go`) become **env-configurable**, defaulting
   to the current production hosts when the env var is unset. New env:
   `WAYFORPAY_PAY_URL`, `WAYFORPAY_STATUS_URL`. Production unchanged (defaults).
2. The frontend action-host allowlist (`purchase-repository.ts`, currently strict
   `secure.wayforpay.com`) keeps that host in production and **additionally** accepts a host from a
   build/runtime env the e2e sets (e.g. `PUBLIC_WAYFORPAY_PAY_HOST`), empty/unset in production.
   Production stays strict. State how you thread this env into the browser bundle for the e2e only.
3. A **stub service** for the e2e Compose stack (small Go program or a stub route set; put it where
   e2e fixtures live — check how the media/asr e2e stacks add helper services). It must:
   - serve the pay page target: accept the POST'd checkout form and immediately redirect the browser
     to the returnUrl's order page (simulating "payment done"), so the frontend return path runs;
   - answer CHECK_STATUS (POST, signed like the real api) with `Approved` for a known order, signed
     with the **test** secret key the stack uses, so the backend's signed CHECK_STATUS verifies and
     applies it;
   - expose a way for the test to POST a **real-shaped signed callback** to our
     `/api/v1/payments/wayforpay/callback` (or the test does this directly with a signed vector).
   The stub shares the test secret key with the API via env — a throwaway key for the stack, not the
   real published test-merchant key.

## The e2e (`frontend/e2e/`, a new spec; follow the existing run-scoped, self-cleaning style)
Drive it as a signed-in student against a published paid course (create via admin API as
`purchase.spec.ts` does). Assert:
- **Pay24 (callback at the HTTP layer):** a callback with a valid signature (test vector, test key)
  → applied, 200 with a valid signed acknowledgement; a bad signature → 401 and nothing changes
  (poll the order: still pending); a replay of the valid one → 200 accept, no second grant.
- **Pay25 shape (against the stub):** «Купити» → order → «Перейти до оплати» → the stub redirects
  back to `/orders/<id>?course=<slug>` → the status page polls → the backend's signed CHECK_STATUS
  (to the stub) returns Approved → **the same `ApplyPaymentEvent`** grants → the locked lesson opens
  (discussion served, no «Купити»). Assert the grant came via the status check, not the return.
- a11y pass on the new states you can reach (axe, 44px) as `purchase.spec.ts` does.

## Constraints
- A5: the browser return applies nothing; only the callback or CHECK_STATUS grants. The stub must
  not be wired so the mere return grants — prove the grant happens on the status check / callback.
- Layering, strict TS, standard-library-only Go, env-only secrets, no key in logs (AGENTS.md).
- `make generate` clean; OpenAPI if any endpoint shape changes (the callback/checkout already exist).

## Verify
Run in order with `DOCKER_CONFIG=/tmp/claude-501/dockercfg` exported:
`make format-check lint typecheck test integration` then **`make seed-identities seed && make e2e`**.
Report each result. If the e2e stack can't bring up the stub in your sandbox (Docker denied), say so
and leave the spec + stub in place; Courses 1 runs the gate. Do not represent a gate you could not
run as passed.

## Channels
If this seam is wrong in a way that would weaken production (e.g. the allowlist env leaking into a
production build, or the stub key path touching prod config), stop and argue it. Final answer =
retelling of the log + changed files + command results.
