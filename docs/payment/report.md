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
| Pay25 | e2e `purchase-wayforpay.spec.ts` against a local stub (checkout→redirect→poll→signed CHECK_STATUS→same ApplyPaymentEvent→lesson opens); **live run 2026-10-02 against the real WayForPay test merchant** (see §6) | offline proof complete; live: signed purchase form accepted by WayForPay, signed CHECK_STATUS round trip applied through ApplyPaymentEvent, return redirect works. **Still OPEN:** an Approved payment → grant, and the real `serviceUrl` callback arriving (§6) |
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

- **Pay25-live is OPEN — the phase does not close until it passes.** AGENTS requires the proof to
  take the production path, not a fake. The offline stub (Pay25, O06c-auto) proves our wiring with a
  self-consistent key; it cannot prove the real WayForPay signature. The live proof is a recorded
  cloudflared tunnel run where the real `serviceUrl` callback reaches, verifies and applies once —
  runbook `work/runbook-o06c-tunnel.md`, run by the owner with the real test-merchant key (kept out
  of the repo). This row is recorded here as open and will be filled from that run.
- **Pay25-live — partial, run 2026-10-02 by Courses 1** (owner away; public test merchant
  `test_merch_n1`, tunnel via localhost.run because this network blocks Cloudflare's quick-tunnel API).
  Proven against the real provider: (1) WayForPay accepted our signed Purchase form and opened its
  hosted page with the right amount and product — the purchase signature is right (this settles the
  wiki-vector doubt in Pay24); (2) our signed `CHECK_STATUS` was accepted, its signed answer verified,
  and a `Declined` result applied through `ApplyPaymentEvent` (order → `failed`, no grant);
  (3) WayForPay redirected the browser to `returnUrl` with `?course=`. **Not proven:** an `Approved`
  payment → grant, and the real `serviceUrl` callback arriving. Why: the test card
  `4111 1111 1111 1111` from WayForPay's API examples is declined by the test merchant
  (`Declined`, reason 1105 "Invalid card number"), and WayForPay publishes no test-card list; and the
  free localhost.run tunnel rotated its hostname within minutes, so the callback went to a dead URL.
  Closing it needs a valid test card or the owner's own test-mode merchant, and a stable public URL
  (a named Cloudflare tunnel or an ngrok account, or the deployment). Evidence: orders `4009ffdb…`
  (Declined), `c15523d2…`/`1ab802ce…` (InProcessing, abandoned).
- **Defect found by the live run — a pending order cannot be re-checked-out.** We send the order id
  as WayForPay's `orderReference`; a second «Перейти до оплати» for the same pending order (buyer left
  the payment page and came back) is refused by WayForPay with `(1112) Duplicate Order ID`, so that
  order is unpayable. The stub could not show this. Backlog B-25; fix before launch.
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

Pay25-live (owner tunnel run) → fill the Pay25 row → PR for the payment phase.
