# Codex brief — O06a: WayForPay backend (provider port, checkout, callback, status check)

Repository /Users/strelia/projects/courses, branch `feat/payment`. **Read first, they bind:**
`AGENTS.md`, `docs/payment/plan.md` (A4, A5, Side-effects, Pay21/21b/24/25),
`docs/payment/assumptions.md` (last entry: "How O06 proves a full purchase"),
`docs/payment/provider-research.md` (Decision). Code: `backend/internal/orders/**`.
Working log: append each step to `work/codex-o06a.log` as you go (first line before the first
tool: what you will check next).

Do **not** commit. Do not touch the frontend (O06b). Do not run the e2e suite.

## Facts already in the code (verify, don't assume)
- `orders` application: `StartPurchase`, `PurchaseStatus`, `ApplyPaymentEvent(event)`
  (idempotent on `(provider, event_id)`, one transaction: inbox → order paid → grant → audit),
  `DecidePaymentEvent` in `domain/payment_event.go` (apply / no_op / retry_unmatched / reject).
- Only `PaymentSucceeded = "paid"` exists. The frontend already shows a `failed` order and offers
  "try again" under a new idempotency key.
- No provider port exists yet (A4: the port appears with its first adapter — now).

## What to build
1. **Port** (`orders/ports`): a payment provider with two operations the application needs:
   - `Checkout(order, buyer, urls) → CheckoutForm` — what the browser must POST to the provider
     (action URL + ordered field list). Redirect model: card data never reaches us.
   - `OrderStatus(orderID) → provider status` — signed server-to-server check.
   Plus a pure verifier for the callback (adapter-side), producing `domain.PaymentEvent`.
2. **Adapter `orders/adapters/wayforpay`** (standard library only; `crypto/hmac` + `crypto/md5`):
   - Purchase form for `https://secure.wayforpay.com/pay`: `merchantAccount`,
     `merchantDomainName`, `merchantSignature` (HMAC_MD5 with SecretKey over
     `merchantAccount;merchantDomainName;orderReference;orderDate;amount;currency;productName…;productCount…;productPrice…`
     — **take the exact field order from the WayForPay wiki**, wiki.wayforpay.com/en/view/852102,
     and pin it with the wiki's own example as a test vector), `orderReference` = our order id,
     `orderDate`, `amount` (from the order's snapshot, minor → "1890.00" style per wiki),
     `currency` UAH, one product (course title, count 1, price), `returnUrl`, `serviceUrl`,
     `language`.
   - **Callback** (`serviceUrl`): parse JSON; verify `merchantSignature` = HMAC_MD5 over
     `merchantAccount;orderReference;amount;currency;authCode;cardPan;transactionStatus;reasonCode`
     (verify the exact list on the wiki "serviceUrl" page, 608996852 / 852102); constant-time
     compare. Map `Approved` → paid event; `Declined`/`Expired` → a **failed** event (see 4);
     anything else (`InProcessing`, `WaitingAuthComplete`, `Pending`, …) → acknowledge, no change.
     Event id: provider-unique and stable across WayForPay's retries (e.g.
     `orderReference + ":" + transactionStatus` — argue a better one if the wiki offers it).
   - Callback **response**: `{"orderReference","status":"accept","time","signature"}` with
     signature HMAC_MD5 over `orderReference;status;time`. Without a valid answer WayForPay
     retries for days — answer `accept` also for duplicates and for statuses we ignore.
   - **CHECK_STATUS**: POST `https://api.wayforpay.com/api` `transactionType=CHECK_STATUS`,
     signature over `merchantAccount;orderReference`; verify the response signature; map as
     the callback. HTTP client with a timeout; no retries inside.
3. **HTTP** (`orders/adapters/http`), wired in the composition root:
   - `POST /api/v1/learn/orders/{orderId}/checkout` (authenticated buyer, same-origin) → the
     `CheckoutForm` for the buyer's own **pending** order (404 otherwise).
   - `POST /api/v1/payments/wayforpay/callback` — **unauthenticated**, no session, no same-origin
     check, body size limit; only the signature authorises. Invalid signature → 401, **nothing
     written** (Pay24). Unknown order → per Pay21b (retry, not a terminal reject).
   - `GET /api/v1/learn/orders/{orderId}` (exists): when the order is still `pending`, run the
     status check first (bounded, e.g. once per few seconds per order) and apply what it returns
     through `ApplyPaymentEvent`. A failure of the check must not fail the status read.
   - OpenAPI for all of it; `make generate` must stay clean (`check-generated` needs staging).
4. **Failed payments:** add a `failed` event type in the domain: pending → failed, no grant,
   idempotent like paid; a failed order never becomes paid by a later event without a new order
   (state why, or argue otherwise). Pay21 (foreign amount → reject) holds for paid events.
5. **Config, secrets:** `WAYFORPAY_MERCHANT_ACCOUNT`, `WAYFORPAY_SECRET_KEY`,
   `WAYFORPAY_MERCHANT_DOMAIN`, public base URL for return/service URLs — **env only**, never in
   logs (log order id and status, never the key or full payload; URLs without query strings).
   Missing config → the checkout endpoint answers 503 `dependency_unavailable`, the app still
   starts. The shared test merchant (`test_merch_n1` and its published key, wiki 852472) may be
   named in `.env.example` with a comment that it is WayForPay's public test merchant — nowhere
   else, never in code.
6. nginx: the callback route reaches the API under the existing `zone=api` limit (check
   `infrastructure/`); do not open anything else.

## Tests (write them first — TDD)
- adapter unit: purchase-form signature against the **wiki's published example** (vector);
  callback verify: valid → event; tampered amount / wrong key → refused; response signature
  vector; CHECK_STATUS request signature + response verification; status mapping table.
- domain: failed event transitions (pending→failed, replay no-op, paid stays paid).
- application: status check applies Approved once; a check error leaves the read working.
- HTTP contract: callback 401 on bad signature with zero rows written (integration, real
  Postgres); valid → applied; replay → no-op + `accept`; checkout 404 for another user's order,
  503 without config. OpenAPI contract tests like the existing ones.
- Run in the Go container (see AGENTS.md / Makefile): `make format-check lint typecheck test
  integration` — report each result. Do not run `make e2e`.

## Channels
If a contract you need is missing, stop and say so. If this brief is wrong in a way that would
cause a bug (field order, event id, status mapping), argue it in your report with the wiki
source instead of silently diverging. Final answer = a short retelling of your log + changed files.
