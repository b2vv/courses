# Codex brief — O06b: frontend redirect to WayForPay and return to the status page

Repository /Users/strelia/projects/courses, branch `feat/payment`. **Read first, they bind:**
`AGENTS.md`, `docs/payment/plan.md` (§Зрізи, A5, Pay07/09/11/13), `docs/payment/assumptions.md`
(the O06 proof entry). Work on the frontend; the backend O06a (`aeff91c`) is done — use its
contract, do not change its behaviour. Log: append each step to `work/codex-o06b.log` (first line
before the first tool call: a `NEXT:` line). Do **not** commit. Do not run `make e2e` (that is O06c).

## The flow today (verify in code)
- `src/routes/learn.$.tsx` `LockedLesson`: «Купити» → `start.mutate({courseId})` →
  `createPurchaseQueries.start` creates a pending order → navigates to `/orders/$orderId?course=<slug>`.
- `src/routes/orders.$orderId.tsx`: polls `learn.purchase.status(orderId)` (which, on the server,
  runs the signed CHECK_STATUS while pending); on `paid` invalidates `learningCourseQueryKey(slug)`.
- Backend already offers `POST /api/v1/learn/orders/{orderId}/checkout` →
  `CheckoutForm { action: string, fields: {name,value}[] }` (generated: `createPurchaseCheckout`,
  model `CheckoutForm`). This is the hosted WayForPay form the browser must submit.

## What O06b must do
After the order is created, the buyer must be sent to WayForPay's hosted page by **submitting the
checkout form**, then returned to the status page. Card data never touches our origin.

1. **Port + adapter + composition** for checkout, in the purchase layer (mirror `start`/`status`):
   - `PurchaseRepository.checkout(orderId, signal?) → PurchaseResult<CheckoutForm-domain>`.
   - domain type for the form (`action: string`, `fields: readonly {name,value}[]`), mapped with a
     strict zod schema in `purchase-repository.ts` (action must be an absolute `https:` URL to a
     WayForPay host — validate the host allowlist; at least one field).
   - a `createPurchaseQueries.checkout` mutation (or query) wired through `createLearnDependencies`.
2. **Submit the form as a real navigation, not fetch.** A redirect-model provider needs a top-level
   POST to its domain: build a `<form method="post" action={form.action}>` with a hidden input per
   field and submit it (programmatically), so the browser leaves our origin for `secure.wayforpay.com`.
   Do not open card entry in our page. Decide and state in your log WHERE this happens:
   - Preferred: a dedicated step so the user sees a clear "redirecting to payment" state. Options —
     (a) «Купити» → create order → go to `/orders/<id>?course=<slug>`; the status page, for a
     `pending` order with no payment yet, shows a «Перейти до оплати» button that fetches the
     checkout form and submits it; or (b) «Купити» submits straight to WayForPay after the order is
     created. Pick (a): it survives the buyer returning without paying (WayForPay return always hits
     returnUrl), lets them retry the redirect, and keeps one status page as the hub. Argue if you
     disagree.
3. **returnUrl must carry the course slug** so the status page can unlock the right course without a
   reload (Pay13). The backend builds `returnUrl = base + "/orders/" + orderId` with no `?course=`
   (`handler.go:149`, flagged in `work/review-o06a-wayforpay-backend.md`). Fix minimally: the
   checkout request carries the course slug; the handler validates it with the **same** slug regex
   already in the codebase and appends `?course=<slug>` to returnUrl only (serviceUrl unchanged,
   grant still never depends on returnUrl — A5). An absent/invalid slug → returnUrl without the
   query, never a rejected checkout. Update OpenAPI + `make generate` (check-generated needs staging).
4. **i18n uk/en** for any new copy («Перейти до оплати», a redirecting state). a11y: ≥44px control,
   visible focus, works at 200% (you can't run e2e; assert in a unit/jsdom test what you can).
5. **A5 holds:** returning from WayForPay lands on the status page; the grant comes only from the
   callback or the signed status check (both already in O06a). The return itself applies nothing.

## Tests (write first)
- adapter: checkout maps a valid form; rejects a non-WayForPay / non-https action; rejects an empty
  field list; names 404 (another user's / non-pending order) and 503 (provider not configured).
- composition: the checkout mutation unwraps the form; a start→checkout sequence keeps the order id.
- presentation (jsdom): the status page for a pending-unpaid order shows «Перейти до оплати»; invoking
  it builds a form whose action is the provider URL and whose hidden inputs equal the returned fields,
  and submits it (spy on submit — do NOT trigger a real navigation in the test).
- Run in `frontend/`: `npm run format:check && npm run lint && npm run typecheck && npm test`. Report
  each result. If a contract you need is missing, stop and say so rather than invent it.

## Channels
If this brief is wrong in a way that would cause a bug (e.g. submitting via fetch would silently
drop the redirect, or appending the slug server-side breaks the signature), argue it in your report
with the source. Final answer = a short retelling of your log + changed files + command results.
