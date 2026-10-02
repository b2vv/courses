# Phase payment — delivery report

Delta to [plan.md](plan.md), [spec.md](spec.md) and [assumptions.md](assumptions.md). Written
2026-10-02 at phase close from the slice commits O01–O06 and the ordered gate. Roles this phase:
Courses 1 planned, gated and committed; Codex wrote the code; Courses 2 reviewed every slice (the
review briefs are `work/review-*.md`). There is no production deployment to check — see §6.

## 1. What was done

A signed-in student sees a paid course's whole outline, opens its preview lessons, and buys the
course: «Купити» creates a pending order, «Перейти до оплати» submits a hosted WayForPay form
(card data never touches our origin), and the order page polls until the payment settles. A grant
opens every locked lesson without a reload. Access is one rule (`internal/contracts/access`,
ADR 014) read by four modules through SQL; a grant comes only from a signature-verified WayForPay
callback or a signed server-to-server `CHECK_STATUS`, never from the browser's return (A5). An
administrator can grant access from the course editor with a reason; setting a course's first price
grandfathers students who already have progress. Orders, grants, payment events and the audit log
live in a new provider-agnostic `orders` module; the WayForPay adapter is its first provider.

## 2. Row-by-row acceptance — 29 rows (Pay01–Pay27 + Pay04b, Pay21b)

Counted as rows, not a range (the Phase 8 lesson). Each row names the test that proves it. Rows
whose proof differs from the plan's wording, or is partial, carry a note — none is claimed away.

| # | Proof | Note |
|---|---|---|
| Pay01 | unit `access_test.go:TestCanView`, `shared_rules_test.go:TestTheAccessPredicateAgreesAcrossModules`; integration R06/R07/R08/K07 (SQL truth table ×4) | — |
| Pay02 | unit `service_test.go:TestStartPurchaseSnapshotsPublishedPaidCoursePrice` (amount_minor=50000, UAH, pending) | exact 50000 asserted in the unit; the integration starts a purchase but does not re-assert the stored value |
| Pay03 | unit `money_test.go:TestNewMoneyAcceptsPositiveUAHMinorUnits` | — |
| Pay04 | integration `repository_integration_test.go:TestOrdersRepositoryAgainstPostgreSQL` (grants=1, events=1, audit=1 in one tx); domain `payment_event_test.go:TestDecidePaymentEventAppliesMatchingPendingOrder` | — |
| Pay04b | integration (second apply of same event → grants stays 1); domain `…ClassifiesRetryRejectAndDuplicate/"duplicate is no-op"` | — |
| Pay05 | unit `access_test.go:TestCanView` (archived+entitled row Want=true); integration R06/R07/R08/K07 | proven as a truth-table row, not as an archive-after-grant sequence |
| Pay06 | integration `TestOrdersRepositoryAgainstPostgreSQL` (grant_audit with event id / admin actor); unit `service_test.go:TestGrantCourseRequiresAdministratorAndWritesAudit` | audit proven by integration+unit; no separate contract test for the audit row |
| Pay07 | contract `handler_test.go:TestCatalogHTTPMatchesOpenAPI` (price+paid); e2e `purchase.spec.ts` (price shown; locked lesson never opens) | — |
| Pay08 | gate `Makefile:check-generated` + the `…MatchesOpenAPI` contract suite per module | a build/CI gate, not one test function |
| Pay09 | e2e `purchase.spec.ts` (44px, axe light/dark, keyboard, 200% text on buy/locked/order surfaces); `accessibility.spec.ts` | — |
| Pay10 | contract `handler_test.go:TestPurchaseAndGrantHTTPMatchOpenAPI/"admin grant"`; e2e `purchase.spec.ts` (grant in editor → lesson opens); unit `CourseGrantPanel.test.tsx` | — |
| Pay11 | integration `learning_integration_test.go:"Pay11 preview on a later lesson…"`; e2e `purchase.spec.ts` (preview plays with discussion; locked lesson offers purchase) | — |
| Pay12 | contract `TestPurchaseAndGrantHTTPMatchOpenAPI`; integration (repeat key → same order); unit `TestStartPurchaseIsIdempotent…` | — |
| Pay13 | e2e `purchase.spec.ts` (processing state; access after the event); FE `purchase-queries.test.ts`, `OrderStatusView.test.tsx` | the "after the event" leg is proven against the stub (O06c-auto) and the status check; the live-provider leg is Pay25-live (§6) |
| Pay14 | integration R06/R07/R08/K07 (both playback readers: R07 VideoAccessReader; `playback_test.go` for loadLessonPlayback) | the second playback reader is a unit query-shape test, not a full SQL truth-table |
| Pay15 | unit `access_test.go:TestCanView` (free rows Want=true) | no dedicated free-course e2e regression; the free path is exercised by every existing e2e (all seed courses are free) |
| Pay16 | contract `…/"owned"→409`; domain `grant_test.go:TestActiveGrantOfEverySourceBlocksAnotherOrder`; unit `…/"active grant blocks purchase"` | proven at domain/application/contract; the postgres integration does not drive a repeat StartPurchase against an active grant |
| Pay17 | integration `…/"Pay17 price change preserves access and order snapshot"` | **added this phase (O07)** — was absent; raise price to 70000 → grant holds, order stays 50000 |
| Pay18 | unit `money_test.go:TestNewMoneyRejectsInvalidAmountsAndCurrencies` (negative/overflow/NaN/EUR) | — |
| Pay19 | contract ×4 — playback `TestPublicPlaybackHTTPMatchesOpenAPI/"Pay19…"`; transcript `TestPay19PaidTranscriptRequiresPurchase`; discussion `TestK10…/TestK11…/"Pay19 paid course"`; learning `TestPay19LearningOutlineLocksPaidLessonsInsteadOfRefusing` (200 + locked:true) | — |
| Pay20 | contract `…/"student forbidden"` (403 grant_forbidden) | — |
| Pay21 | integration (foreign amount → ErrPaymentEventRejected, recorded terminal, no grant); domain `…/"amount mismatch is terminal reject"` | — |
| Pay21b | integration (unknown order → ErrPaymentEventUnmatched, 0 inbox rows); domain `…/"unknown order retries"`; HTTP `TestWayForPayCallbackHTTPContract/"unmatched retries"` (503) | — |
| Pay22 | contract `…/"Pay22 dependency failures are explicit"` (503 dependency_unavailable) | — |
| Pay23 | integration `access_integration_test.go:"Pay23 active grant lookup uses the partial index"` (EXPLAIN ANALYZE, BUFFERS → Index Scan on `course_grants_one_active_per_user_course`) | index usage asserted; the "≤2× baseline" comparison is **not** asserted — no before/after number is pinned, so that magnitude stays unproven |
| Pay24 | unit vectors `wayforpay_test.go` (purchase/callback/ack/CHECK_STATUS signatures); HTTP `TestWayForPayCallbackHTTPContract` (bad sig→401 nothing written, approved→apply, replay→accept no second grant) | the wiki's printed purchase hash is a documentation error; the reproducible value `4941ea0c…` is pinned. Real-signature conformance is Pay25-live (§6) |
| Pay25 | e2e `purchase-wayforpay.spec.ts` against a local stub (checkout→redirect→poll→signed CHECK_STATUS→same ApplyPaymentEvent→lesson opens); **live Approved run 2026-10-03 on the owner's own test-mode shop** (see §6) | offline proof complete; **live CLOSED:** a real 1-грн card payment on the owner's test-mode shop (`site_ex`/`site.ex`) returned `Approved`; the real `serviceUrl` callback arrived `POST → 200` (`2026-10-02T22:18:55Z`, req `49ef0938`), order `4396fba5…` → `paid` (100 UAH), **exactly one `paid` event and one active `purchase` grant — applied once**. Earlier same-day legs (signed purchase form accepted, signed CHECK_STATUS round trip, return redirect) also hold (§6) |
| Pay26 | integration `…/"Pay26 grandfather covers two students and repeats idempotently"` | **extended this phase (O07)** — two students → 2 grants; repeat price-set → still 2 |
| Pay27 | unit property `access_test.go:TestPurchaseRequiredImpliesCannotView`, `TestCanViewAndPurchaseRequiredUseTheSameCourseFacts` | — |

## 3. Gate

Ordered `make check` (format-check, lint, typecheck, test, integration, check-generated, build, up,
drain-queue, seed, smoke, e2e) → exit 0. rstest 539, e2e 89, 369 Go test functions. Run on
`98499db`; the O07 gap tests (`2f430d3`) add Pay17/Pay26 to the integration set, re-run green.

## 4. Decisions taken this phase (artefacts)

Provider: WayForPay now, IBAN+QR and Paddle later (`provider-research.md` → Decision, 2026-10-01).
The learning outline of a paid course shows the whole structure (owner, 2026-10-01). O06c proof
split: offline stub in the gate + one live tunnel run (owner, 2026-10-01). All in `assumptions.md`.

## 5. Two defects found and fixed

The learning outline took the preview fact from the first SQL row only (a paid course whose preview
was not first refused the whole page); and the learning course never read `price_minor`/`currency`,
so the buy panel had no amount. Both found by the slice work, fixed, and covered by tests
(`assumptions.md`, 2026-10-01).

## 6. Limitations (named, not claimed away)

- **Pay25-live — CLOSED, live Approved run 2026-10-03 by the owner.** AGENTS requires the proof to
  take the production path, not a fake; this took it. On the owner's **own WayForPay shop in test mode**
  (`merchantAccount=site_ex`, `merchantDomainName=site.ex`; SecretKey in local `.env` only), the owner
  paid the 1-грн `live-pay25-07a81323` course with a **real card** in the browser through a stable
  **ngrok** tunnel (localhost.run rotated its hostname within minutes — the first URL was dead before
  payment, so we switched to ngrok, which holds a URL for the session). Observed end to end:
  checkout `POST → 200`, then WayForPay's real `serviceUrl` callback `POST /api/v1/payments/wayforpay/callback → 200`
  (`2026-10-02T22:18:55Z`, requestId `49ef0938b982e4c73503962181ff4774`; the body is never logged by
  design). Applied-once proof, scoped to that order/course: order `4396fba5-0860-4006-970b-e1c1f33f6650`
  → `paid`, `amount_minor=100` UAH, **1 `paid` event / 1 event total**; **1 active `purchase` grant**
  for `student@courseforge.local` on `live-pay25-07a81323`; the locked lesson opened. (The global
  `purchase`-grant count is noise from prior stub/e2e runs — 189000-minor orders — which is why the
  count is scoped to this course.) The course was archived after the run (published set back to the
  seeded three) and the tunnel/ngrok torn down. **Decision (owner, 2026-10-03):** the live run used the
  existing SecretKey (it had appeared in chat) — acceptable for a one-off 1-грн test-mode charge; the
  key is to be regenerated and the merchant password changed before any further live use (the password
  is not used by the project and is stored nowhere). Runbook: `work/runbook-o06c-tunnel.md`.
- **Pay25-live — earlier same-day attempts on the shared test merchant `test_merch_n1` (record).** Run 2026-10-02 by Courses 1 (owner away; public test merchant
  `test_merch_n1`, tunnel via localhost.run because this network blocks Cloudflare's quick-tunnel API).
  Proven against the real provider: (1) WayForPay accepted our signed Purchase form and opened its
  hosted page with the right amount and product — the purchase signature is right (this settles the
  wiki-vector doubt in Pay24); (2) our signed `CHECK_STATUS` was accepted, its signed answer verified,
  and a `Declined` result applied through `ApplyPaymentEvent` (order → `failed`, no grant);
  (3) WayForPay redirected the browser to `returnUrl` with `?course=`. **Not proven on `test_merch_n1`**
  (later closed on the owner's own test-mode shop, see the CLOSED bullet above): an `Approved`
  payment → grant, and the real `serviceUrl` callback arriving. Why: the test card
  `4111 1111 1111 1111` from WayForPay's API examples is declined by the test merchant
  (`Declined`, reason 1105 "Invalid card number"), and WayForPay publishes no test-card list; and the
  free localhost.run tunnel rotated its hostname within minutes, so the callback went to a dead URL.
  Closing it needs a valid test card or the owner's own test-mode merchant, and a stable public URL
  (a named Cloudflare tunnel or an ngrok account, or the deployment). Evidence: orders `4009ffdb…`
  (Declined), `c15523d2…`/`1ab802ce…` (InProcessing, abandoned).
  **Second run, same day, by card:** `5454 5454 5454 5454` → `Declined`, 1101 "Declined To Card
  Issuer" (order `ea6492a4…` → `failed` through the signed CHECK_STATUS); `4111 1111 1111 1111` →
  `Declined`, 1105 "Invalid card number" again (order `b2f96989…`), although it is commonly cited
  as WayForPay's success card. No test card produced `Approved` on `test_merch_n1`; the Approved →
  grant leg needs a card WayForPay confirms for test mode.
  **Third try (2026-10-02):** `4222 2222 2222 2222`, suggested as WayForPay's Approved card, fails the
  Luhn checksum; WayForPay's own form marks it «Невiрна картка» before submitting (order `09fd19e0…`
  stays pending). The tables circulating for WayForPay test cards look like mixes of other providers'
  test numbers (`4000000000000002` is Stripe's 3-D Secure card); only WayForPay can name the card.
  **Fourth try:** the 13-digit `4222222222222` (Luhn-valid, also circulated as "WayForPay's Approved
  card") went through WayForPay's form and came back `Declined`, 1101 "Declined To Card Issuer"
  (order `6aa2b4f9…` → `failed`, lesson stayed locked). Summary of live outcomes on `test_merch_n1`:
  4111… → 1105; 5454… → 1101; 4222222222222 → 1101; 4222 2222 2222 2222 → refused by the form
  (not Luhn-valid). No circulated card yields `Approved`; stop guessing — ask WayForPay support.
  **Why (search of WayForPay's own material, 2026-10-02):** no WayForPay-owned source — the wiki in
  uk/ru/en, the help centre, the `github.com/wayforpay` SDK and plugin READMEs — publishes a card that
  yields `Approved` in test mode. The only concrete description found is third-party (Weblium's
  WayForPay integration guide): the test cards `4111…` and `5454…` "will be processed as
  unsuccessful. For successful payments you need to use a real card using real funds", and "if the
  store is in test mode, all payments are refunded automatically within 15 minutes". That matches every
  live outcome above. So the Approved → grant leg is provable only with a real card on a merchant in
  test mode — a step the owner must perform (an agent never enters a real card); the auto-refund is
  unofficial until WayForPay confirms it. Trace: `work/_wip-wfp-test-cards.md`.
- **Defect found by the live run — closed (B-25, 2026-10-02).** We send the order id as WayForPay's
  `orderReference`, and WayForPay refused a second hosted form for the same pending order
  (`(1112) Duplicate Order ID`), leaving that order unpayable; the stub could not show it. Fixed: a
  repeat «Перейти до оплати» on an order whose checkout was already opened gets 409
  `checkout_already_opened` and the page starts a new order (decision in `assumptions.md`,
  2026-10-02; option b, money-path reasoning there).
- **Refund trace has no screen yet.** B-25 (2026-10-02) marks `orders.refund_required = true` when money
  arrives without a grant: a paid order for a course already granted, a paid event rejected for an
  order that had already failed, or a paid event with a foreign amount (Pay21). Each also writes a
  structured warning log. There is no admin view or API for it (no production consumer, so no port);
  until one exists, find refunds with:
  ```sql
  SELECT o.id AS order_id, o.user_id, o.course_id, o.status,
         o.amount_minor AS order_amount_minor,
         e.event_id, e.amount_minor AS charged_amount_minor, e.received_at
    FROM orders o
    JOIN payment_events e ON e.order_id = o.id AND e.type = 'paid'
   WHERE o.refund_required
   ORDER BY e.received_at;
  ```
  **Refund `charged_amount_minor`, not `order_amount_minor`:** for a foreign-amount event (Pay21)
  WayForPay charged the event's amount, which differs from the order's snapshot.
- **No production deployment.** Live payments need a ФОП and a public `serviceUrl`; both are a launch
  step (`provider-research.md`). Everything above is proven locally / against the test merchant.
- **Pay23** pins index usage, not the ≤2× timing. **Pay15** free-regression and **Pay16** repeat-
  purchase lack a dedicated e2e / integration leg (covered at other levels, see §2). **Pay02**'s
  exact 50000 is a unit, not the named integration. None of these blocks the access or money
  invariants, which are proven at the levels §2 names.
- **The status throttle is per process.** `service.go` keeps the CHECK_STATUS throttle in memory, so
  with N API instances the signed status check fires up to N× as often. Correctness is unaffected —
  the grant goes through the callback / `ApplyPaymentEvent`, not the throttle — but a shared store
  (e.g. Redis) would be needed before horizontal scaling. Named, not scheduled.
- **grandfather** keys on playback progress only: a learner who used a free course solely through
  transcript or discussion, with no video progress, gets no grant on free→paid (plan NB4).

## 7. Next

Pay25-live closed (live Approved run 2026-10-03, §6). All Pay01–Pay27 + Pay04b/Pay21b rows proven.
Remaining: merge PR #3 for the payment phase; before any further live use, rotate the shop SecretKey
and change its merchant password (owner decision, §6).
