# Review request — O06b: frontend redirect to WayForPay + return (Codex; Courses 1 gated)

Base `aeff91c`. Uncommitted. Brief: work/brief-codex-o06b-wayforpay-redirect.md; Codex log:
work/codex-o06b.log.

## Gates (Courses 1, outside Codex's sandbox)
`make generate format-check lint typecheck test integration` → exit 0 (real Docker). rstest 539/539.
e2e is O06c, not run.

## Change
- purchase domain gains a checkout form type; `PurchaseRepository.checkout(orderId, courseSlug?)`;
  adapter maps it with a strict zod schema that requires an `https:` action on a WayForPay host and
  ≥1 field (rejects other hosts / empty fields → invalid_response). Composition: a checkout mutation.
- `OrderStatusView` / `orders.$orderId.tsx`: a pending-unpaid order shows «Перейти до оплати»; it
  fetches the checkout form and submits a **native top-level POST** to the WayForPay action with a
  hidden input per field (full-page leave, card data never on our origin). jsdom test spies on submit
  (action + hidden fields = the returned form), no real navigation.
- returnUrl slug (Pay13): checkout request carries `courseSlug`; backend handler appends
  `?course=<slug>` to returnUrl **only** when it passes `validCourseSlug` (same shape as the slug
  regex; empty/absent/invalid → returnUrl without the query, never a rejected checkout). serviceUrl
  unchanged; grant still never depends on returnUrl (A5). `decodeCheckout`: nil/empty body → no slug,
  no error; malformed JSON / unknown field → 400.
- OpenAPI + generated client updated.

## A5 (please confirm against the code)
Returning from WayForPay lands on `/orders/<id>?course=<slug>`; the grant comes only from the
callback or the signed CHECK_STATUS (both O06a). The return itself, and the slug in returnUrl,
apply nothing.

## Please look hardest at
1. The action-host allowlist in the adapter: can a tampered/unknown action URL ever be submitted?
2. `validCourseSlug` server-side: injection into returnUrl (url.QueryEscape is used) — any way to
   bend returnUrl to another origin or to the serviceUrl?
3. The pending→pay→return→(poll/status-check)→unlock path still honours A5 and Pay13.

## Findings (Courses 2 appends below)

---
## Findings — Courses 2 (claude-opus-4-8, 2026-10-01)

ДАЛІ: (1) action-host allowlist у adapter — чи може підроблений action потрапити в submit;
(2) validCourseSlug→returnUrl backend — чи можна зігнути returnUrl на інший origin/serviceUrl
(url.QueryEscape); (3) native top-level POST + A5 (returnUrl/слаг нічого не видає) + Pay13. Читаю
adapter checkout-схему, backend returnUrl-builder, OrderStatusView submit.

### F1 [#1 action-host allowlist] — SOLID
adapter schema: action = z.url().refine(new URL → protocol==='https:' && hostname==='secure.wayforpay.com'),
fields.min(1), .strict(). Строга рівність хоста: http, інший хост, субдомен (secure.wayforpay.com.evil),
userinfo-трюк (secure.wayforpay.com@evil → hostname=evil) — усі reject → invalid_response → submit не
відбувається. submitCheckoutForm отримує лише валідовану форму (mutation падає на поганому action).

### F2 [#2 validCourseSlug → returnUrl] — SOLID (подвійний захист)
backend: returnURL = publicBaseURL(серверний) + "/orders/" + orderId(validID UUID); ?course=<slug>
додається ЛИШЕ коли validCourseSlug (len 1-120, без leading/trailing/подвійного дефіса, lowercase
alnum+hyphen) І через url.QueryEscape. Слаг не може нести &/?/#/:/ /→ не зігнути origin, path, не
додати serviceUrl. invalid slug → returnUrl без query (не rejected checkout). serviceUrl окремий,
серверний. Origin/path суто серверні.

### F3 [#3 native POST + A5 + Pay13] — SOLID
submitCheckoutForm: document.createElement('form'), method=post, action=validований secure.wayforpay.com,
hidden-inputs(name/value, не innerHTML → без XSS), append body, requestSubmit()/submit() → TOP-LEVEL
full-page POST. Без fetch/XHR → картка вводиться на WayForPay, не на нашому origin. Тригер: pending →
«Перейти до оплати» → onCheckout → mutation fetch форми → submit лише на success.
A5: checkout нічого не видає (лише будує форму+редирект); повернення → /orders/<id>?course=<slug> →
статус-сторінка → poll/CHECK_STATUS → unlock. Грант ТІЛЬКИ від callback/CHECK_STATUS (O06a).
returnUrl і слаг не видають нічого. Pay13: слаг веде статус-сторінку до курсу + paid→invalidate outline
(зріз 2), право не залежить від returnUrl.

### Forward-нотатка (O06c)
Хост checkout зашитий як secure.wayforpay.com. Тест-мерчант WayForPay теж POST-ить на secure.wayforpay.com
(тест-режим — за merchant account, не окремий хост), тож allowlist має пройти — але O06c це імпліцитно
підтвердить (якщо тест-мерчант прийме форму — хост вірний). Якщо sandbox раптом на іншому хості —
allowlist відхилить; тоді розширити refine на тест-хост.

## ПІДСУМОК: 0 blocking — COMMIT-READY (e2e = O06c)
action-allowlist строгий (лише https secure.wayforpay.com); returnUrl не зігнути (validCourseSlug+
QueryEscape, серверний base); native top-level POST (картка off-origin); A5/Pay13 тримаються (грант
лише callback/CHECK_STATUS). Forward: O06c підтвердить checkout-хост тест-мерчанта.
