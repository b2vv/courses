# Codex brief — B-25: repeat checkout → new order; explicit double-payment trace

Repository /Users/strelia/projects/courses, branch `feat/payment`. Read first (they bind): `AGENTS.md`,
`docs/payment/assumptions.md` (last entry, 2026-10-02 — the decision), `work/backlog.md` B-25,
`docs/payment/plan.md` (A5, Pay04/04b/12/16/21/21b). Log each step to `work/codex-b25.log` (first line
before the first tool: `NEXT:`). Do **not** commit.

## Why
Live run 2026-10-02: we send the order id as WayForPay's `orderReference`; a second hosted form for the
same pending order is refused by WayForPay `(1112) Duplicate Order ID`, so a buyer who left the payment
page cannot pay that order. Decision: a repeat attempt is a **new order** (option b). Separately, a
second paid order for an already-granted course is silently swallowed (`ON CONFLICT DO NOTHING` in
`repository.go` `Apply`) — it must leave an explicit refund trace.

## Build
1. **Migration `00012`** (goose, same style/lock_timeout as 00011): `orders.checkout_opened_at
   timestamptz NULL`, `orders.refund_required boolean NOT NULL DEFAULT false`. Down migration too.
2. **Checkout once per order:** `Service.Checkout` records `checkout_opened_at` the first time (same
   transaction/guard as reading the pending order); a second checkout of an order whose checkout was
   already opened returns a new domain error → HTTP **409 `checkout_already_opened`** (OpenAPI +
   generated client). Do not change the WayForPay adapter or the `orderReference` = order id rule.
3. **Frontend** (`routes/orders.$orderId.tsx`, `OrderStatusView`, purchase adapter/queries): on 409
   `checkout_already_opened`, «Перейти до оплати» starts a **new purchase** (`start` mutation — new
   idempotency key, as «Спробувати ще раз» does), then fetches the new order's checkout form and
   submits it; navigate to the new order's status page. The adapter names the new error code. i18n
   uk/en for any new copy. Layering per AGENTS (no fetch in presentation).
4. **Double-payment trace:** in `Apply`, when the purchase grant is NOT inserted (conflict with an active
   grant) for a `paid` decision, set `orders.refund_required = true` on that order in the same
   transaction, and emit a structured warning log (order id, course id, user id — no payload, no key).
   Add a repository read `OrdersRequiringRefund(ctx)` (paid + refund_required) for future admin use;
   no UI now.

## Tests (write first)
- integration (`repository_integration_test.go`, real Postgres; give new subtests their OWN id source —
  this file has a shared id slice that starves other subtests; and scope any count to your own rows):
  (1) second checkout of an opened order → `checkout_already_opened`; StartPurchase with a new key
  makes a new order; (2) late Declined (failed event) of the old order leaves the new order pending;
  (3) Approved of the old order after the new one exists → grant issued, new order stays pending;
  (4) two orders for one course both Approved → exactly one grant; the second order `paid` with
  `refund_required = true`; `OrdersRequiringRefund` returns it.
- HTTP contract: 409 `checkout_already_opened` against OpenAPI.
- frontend: adapter names 409; status page on 409 starts a new purchase and submits the new form
  (spy, no real navigation).
- Users insert in fixtures with the existing Argon2id fixture hash (a plain string fails the check).

## Verify
`DOCKER_CONFIG=/tmp/claude-501/dockercfg make generate format-check lint typecheck test integration`.
If Docker is denied in your sandbox, say so — Courses 1 runs the gate. Never claim a gate you didn't run.

## Channels
If the decision is wrong in a way that would lose money, argue it with the code path instead of
silently diverging. Final answer: changed files, what you ran, what you couldn't.
