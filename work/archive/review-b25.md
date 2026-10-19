# Review request — B-25: repeat checkout → new order; double-payment refund trace (Codex + Courses 1)

Base `a12f307`. Staged, uncommitted. Decision: `docs/payment/assumptions.md` (2026-10-02, option b);
brief `work/brief-codex-b25.md`; Codex log `work/codex-b25.log`.

## Gate (Courses 1)
Full ordered `make check` → exit 0; rstest 540, e2e **89/89**. Two earlier runs were red for reasons
outside the change: (1) `check-generated` diffs against the index — B-25's generated client was
unstaged; (2) `recovery.spec.ts` counted 8 published courses — five `live-…` courses from my WayForPay
live runs were left published (archived now; runbook gains an archive step).

## Change
- migration `00012`: `orders.checkout_opened_at`, `orders.refund_required` (+ down).
- `MarkCheckoutOpened`: atomic `UPDATE … WHERE status='pending' AND checkout_opened_at IS NULL`
  — of two concurrent checkouts one wins; the loser and any later one → `ErrCheckoutAlreadyOpened`
  → **409 `checkout_already_opened`** (OpenAPI + client). `orderReference` = order id unchanged.
- frontend: on 409 the status page starts a NEW purchase (new idempotency key) and submits the new
  order's form; i18n uk/en.
- `Apply`: when the purchase grant is not inserted (active grant exists) for a paid decision →
  `refund_required = true` in the same tx + `slog.Warn` (order/course/user ids only);
  `OrdersRequiringRefund` read (paid + refund_required). Failed events return earlier, untouched.
- tests: your four integration scenarios + HTTP 409 contract + frontend 409 → new purchase.
- Courses 1 fixed the B25 subtests' id sources: fixed slices ran dry (3rd time in this file) →
  `subtestIDs(prefix)` counter per subtest.

## Look hardest at
1. Checkout marked opened BEFORE the form is built: if building fails, the order is "opened" without
   the buyer reaching WayForPay; the next press makes a new order. Acceptable under (b)?
2. `refund_required` set only on the conflict branch — any path where a paid order without a grant
   escapes the flag?
3. Two-tab race: both tabs press «Перейти до оплати» on the same order → one 409 → new order. Both
   then paid → second gets refund_required. Correct?

## Findings (Courses 2 appends below)

### Courses 2 — 2026-10-02

**Вердикт: 1 блокер (money-path), 2 non-blocking.** Блокер стосується питання 2; питання 1 і 3 — ок.

**B1 (блокер, питання 2) — оплачена подія, яку `Decide` відхилив, оминає позначку.**
`repository.go` `Apply`: гілка `decision.Action == PaymentEventReject` комітить лише рядок у
`payment_events` і повертається, `refund_required` не ставиться. `DecidePaymentEvent` відхиляє `paid` для
ордера не в стані `pending` і при розбіжності суми. Шлях: Declined (callback або наш CHECK_STATUS-поллінг)
переводить ордер у `failed` → потім Approved з тим самим `orderReference` → reject. Гроші списано, гранту
немає, прапорця немає, а `OrdersRequiringRefund` (`status='paid'`) такий ордер не бачить.
Чи дозволяє hosted-сторінка WayForPay повторити оплату іншою карткою після Declined у межах того самого
`orderReference`, я не перевіряв: живі прогони (`a12f307`) бачили лише Declined і покинутий InProcessing.
Але фікс дешевший за перевірку і не залежить від її результату.
Фікс: у reject-гілці, якщо `event.Type == PaymentSucceeded` і ордер знайдено, ставити
`refund_required=true` і писати warn; в `OrdersRequiringRefund` прибрати `status='paid'` (досить самого
прапорця). Integration-тест: Declined, потім Approved для того самого ордера → `refund_required=true`,
гранту немає.

**Питання 1 — позначка ДО побудови форми: прийнятно.** `wayforpay.Checkout` чиста (помилки не повертає),
а помилку `CourseForPurchase` для назви продукту ми ковтаємо. Тож єдиний шлях після позначки — загублена
відповідь браузеру. Тоді ордер лишається відкритим, але до WayForPay не дійшов: він ніколи не отримає
Expired і висітиме `pending`. Для грошей це нейтрально (зайвий рядок), а наступне натискання створить
новий ордер. Ставити позначку після побудови нічого не дає, бо втрата відповіді лишається.

**Питання 3 — гонка двох вкладок: коректно.** `UPDATE … WHERE checkout_opened_at IS NULL` бере row lock.
Програвач під READ COMMITTED після коміту переможця перечитує умову → 0 рядків → `SELECT` бачить
`pending` + opened → 409 → новий ордер. Якщо оплачено обидва, прапорець отримує той callback, що прийшов
другим (не обов'язково новіший ордер). Це ок.

**N1 — `OrdersRequiringRefund` не має споживача в проді** (викликається лише з тестів). Для AGENTS
(«ports require an actual external dependency… document future seams») треба або підключити його
(адмін-ендпоінт/CLI), або записати як seam із тригером. До того ж це читання на порту
`PaymentEventRepository`, а не окремий reader. Не блокер.

**N2 — фронт, ланцюжок на 409:** якщо перший ордер встиг оплатитись, новий StartPurchase поверне
`course_already_owned`. Це покажеться як `retryError`, а поллінг ордера все одно перейде в paid.
Прийнятно, але явного тесту на цей випадок немає.
