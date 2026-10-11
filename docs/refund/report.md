# Refund handling — report

Дельта до [plan.md](plan.md). Гілка `feat/refund-handling`, фаза закрита (фінальний ordered
`make check` EXIT=0, e2e **115 passed**). Закриває `docs/payment/report.md` §6 + backlog B-26.

## 1. Що зроблено

Усе — як у plan.md §Розміщення/§Reuse; розгорнуто лише відхилення від плану.

- **R1** `3e05fb7` — міграція 00017 (`refund_reason` CHECK, `refund_resolved_at/by`); заповнення
  `refund_reason` у двох наявних точках `Apply`. Як у плані.
- **R2** `77f4f61` — домен `PaymentRefunded` + дії `PaymentEventRevoke`/`PaymentEventIgnore`,
  рефанд-гілка перед amount-guard. Як у плані (G9/G11).
- **R3** `3d4c40b` — wayforpay `Refunded`/`Voided`→`PaymentRefunded`. Як у плані.
- **R4** `4b04c63` — рефанд-гілка в `Apply` (revoke за `order_id`, audit лише RowsAffected==1,
  Ignore→warn). Як у плані (G5/G7/G10).
- **R5+R6** `a16e0c7` — **відхилення від плану:** план розділяв R5 (application) і R6 (repository),
  але зміна інтерфейсу `OrderRepository` їх злютувала (реальний репозиторій мусить задовольняти
  інтерфейс, щоб пакет компілювався) → один коміт. Логіка — як у плані.
- **R7** `b64d2a1` — 3 admin-роути + OpenAPI/Orval. **Відхилення:** resolve повертає `204` (план не
  фіксував код; 204 бо нема тіла).
- **R8** `775b5c1` — frontend екран/роут/nav/i18n + e2e. **Відхилення:** route вимагав регену
  `routeTree.gen.ts` (поза build — через `@tanstack/router-generator`) і оновлення R24 route-list
  тесту — у плані не передбачено.
- **review-followup** `9124286` — закрив знахідки code-review: **A14** (рядок несе membership-стан,
  revoke-кнопка лише для живого гранта — план/spec цього вимагали, перша реалізація показувала
  кнопку завжди); AC6-лог `transaction_status`; 409 `grant_transition_invalid`; loader `prefetchQuery`.

## 2. Протестовані сценарії (приймальна таблиця plan.md A1–A20)

| # | Очікуване (скорочено) | Фактичний результат | Вердикт |
|---|---|---|---|
| A1 | Список refund_required із причиною+сумами | HTTP contract + e2e refund.spec (grant_conflict, charged з paid-події, membershipState) | збіглося |
| A2 | Ручний revoke через `RevokeForRefund` | app unit + HTTP contract (200 revoked) + e2e (діалог) | збіглося |
| A3 | Resolve ідемпотентний, рядок зникає | integration Rf1 + HTTP 204 + e2e (рядок зникає) | збіглося |
| A4 | Вебхук revoke за order_id + system-audit | integration Rf4 (actor NULL/reason/payment_event_id) | збіглося |
| A5 | Порожня черга → empty-state | **не перевірено e2e** (shared DB не гарантує порожню чергу); покрито рендером компонента (`items.length===0`) | частково |
| A6 | Рефанд відкликає active+blocked | integration Rf6 (Voided на blocked→revoked) | збіглося |
| A7 | Voided == Refunded | domain unit + wayforpay unit + integration | збіглося |
| A8 | Повтор того самого статусу — дедуп | integration Rf8 (1 подія, 1 audit) | збіглося |
| A9 | Повторний resolve — no-op, не перезапис | integration Rf1 (resolved_at не рухається) | збіглося |
| A10 | Рефанд на не-paid → warn+ACK, без refund_required | integration Rf10 | збіглося |
| A11 | Рефанд на вже-revoked → no-op | integration Rf11 (у Rf-наборі) | збіглося |
| A12 | Конкурентний revoke → 409 | HTTP contract (membership_conflict) | збіглося |
| A13 | Неадмін → 403 | HTTP contract (student→403 усі три) | збіглося |
| A14 | Нема живого гранта → нема revoke-кнопки | DTO membershipState + e2e (після revoke кнопка зникає) | збіглося (після review-fix) |
| A15 | Migration-preservation | course migration-preservation (повний DownTo→Up+seed) | збіглося |
| A16 | Невалідна сигнатура → 401 | wayforpay unit (tamper/wrong-key) | збіглося |
| A17 | Частковий рефанд теж відкликає | domain unit + integration Rf17 (10000 на 50000) | збіглося |
| A18 | Admin re-grant під NULL order_id не чіпається | integration Rf18 | збіглося |
| A19 | Refunded потім Voided → 1 revoke, 2 події | integration Rf19 | збіглося |
| A20 | Resolve на нефлагованому → no-op | integration Rf20 + HTTP | збіглося |

«Не перевірено» рядків: **0** повних; A5 — частково (компонент, не e2e), причина названа.

## 3. Неочевидне

- **Фіксований id-слайс пакетного `service` вичерпується рефандом** — рефанд-подія з'їдає ще один
  `newID`; рефанд-виклики через пакетний `service` дали panic index-out-of-range
  (`repository_integration_test.go` — виправлено на svc з unbounded `subtestIDs`).
- **`DecidePaymentEvent` guard reject'ив би рефанд** — новий тип падав у amount-guard
  (`payment_event.go:44`: `Type∉{paid,failed}`, `amount≤0`), а `handler.go` ковтає
  `ErrPaymentEventRejected` → підписаний рефанд тихо лишив би грант active. Закрито гілкою перед
  guard + окремими діями (не reject). Знайдено на GATE 2 (defense Q1).
- **`refund_required` додано в 00012, не 00011** — spec-драфт помилявся; звірено з
  `backend/migrations/00012_payment_checkout_refunds.sql:6`.
- **Додавання роуту вимагає регену `routeTree.gen.ts`** поза Docker-build (плагін `tanstackStart`);
  зроблено через `@tanstack/router-generator` `Generator.run()`.

## 4. Скріни

Не застосовно для бекенд-слайсів (схема/домен/HTTP). Фронтенд-екран `/admin/refunds` перевірено
e2e-прогоном (`refund.spec.ts`) з axe-аудитом (0 порушень), не скріншотами до/після — новий екран,
«до» не існує.

## 5. Проходка по user flow

Прогнано e2e (`frontend/e2e/refund.spec.ts`), наскрізь:
1. Студент стартує покупку курсу (pending-замовлення).
2. Адмін видає цьому студенту доступ, поки замовлення pending.
3. Приходить підписаний `Approved` callback → settle конфліктує з admin-грантом → замовлення
   позначається `refund_required` (`grant_conflict`), доступ не дублюється.
4. Адмін відкриває `/admin/refunds` → бачить рядок із причиною «Доступ уже був», сумами, часом;
   axe чисто.
5. Адмін тисне «Відкликати доступ» → підтверджує в діалозі → грант стає `revoked`; черга
   перезапитується, і рядок більше не пропонує revoke (лишається тільки «Позначити розглянутим»).
6. Адмін тисне «Позначити розглянутим» → рядок зникає з черги.

## 6. Ризики й обмеження

- **Refund-before-paid reordering** (spec §Named limitations): якщо `Refunded`/`Voided` прийде
  раніше за `Approved` (WayForPay переставив доставку — рідко), не-paid-гілка зробить warn+no-op, а
  пізніший `Approved` видасть доступ попри повернення коштів. Реконсиляції нема — поза scope.
- **A5 empty-state** перевірено лише рендером компонента, не e2e (спільна БД не гарантує порожню
  чергу під час прогону).
- **Ми не ініціюємо рефанд** — фіча лише реагує на вже зроблений у кабінеті WayForPay рефанд і дає
  видимість розбіжностей; повернення коштів покупцю через WayForPay API — окрема інтеграція, поза
  scope (spec §Out of scope).
- **Відображення курсу/юзера в черзі — сирі id** (не назви): екран адмінський, діагностичний;
  людські назви можна додати пізніше (окремий join), борг названо тут.

## 7. Post-deploy

(Порожньо — заповнюється `post-deploy-check` після деплою. Деплою нема — тест-режим, Р1.)
