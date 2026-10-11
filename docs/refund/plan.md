# Refund handling — plan

Реалізація рішень G1–G8 (`spec.md` §Рішення GATE 1). ЩО/ЧОМУ — у `spec.md`; тут ЯК.

## Розміщення

Backend (усе в наявному модулі `orders` — payment живе там, окремого `payment` нема):
- `backend/internal/orders/domain/` — новий тип події `PaymentRefunded`, дія `PaymentEventRevoke`,
  розширення `DecidePaymentEvent`; `refund_reason` — не домен-стан, а колонка персистенції.
- `backend/internal/orders/application/service.go` — `RevokeForRefund(actor, orderID)`,
  `RefundQueue(actor)` (читання списку).
- `backend/internal/orders/ports/repository.go` — `RefundQueue`, `ResolveRefund`,
  розширення `OrderRepository`/`PaymentEventRepository` як треба.
- `backend/internal/orders/adapters/postgres/repository.go` — гілка рефанду в `Apply`;
  читання черги; `ResolveRefund`; `refund_reason` у двох наявних точках `refund_required=true`.
- `backend/internal/orders/adapters/wayforpay/wayforpay.go` — мапінг `Refunded`/`Voided`.
- `backend/internal/orders/adapters/http/handler.go` — 3 нові admin-роути (список, resolve,
  revoke-for-refund) у `RegisterAdmin`; вебхук-роут **не чіпаємо** (рефанд тече наявним
  `callbackHandler`).
- `backend/migrations/00017_refund.sql` — колонки на `orders`.
- `backend/openapi/openapi.yaml` — 3 нові admin-ендпоінти → Orval.

Frontend:
- `frontend/src/presentation/admin/RefundQueue.tsx` — екран (дзеркало `ParticipantReportQueue`).
- `frontend/src/routes/admin.refunds.tsx` — роут (дзеркало `admin.participant-reports.tsx`).
- `frontend/src/app/composition/enrollment-queries.ts` (+`enrollment.ts`) — `refundQueue`,
  `resolveRefund`, `revokeForRefund` (той самий файл, що тримає admin participant-reports +
  revoke; рефанд — суміжна адмін-черга).
- `frontend/src/presentation/admin/AdminShell.tsx` — пункт навігації.

## Reuse-first

**Переюзуємо (існуюче підходить точно):**
| Що | Звідки | Чому підходить |
|---|---|---|
| Перехід гранта (revoke) + audit + compare-and-set — **лише ручний шлях** | `service.transitionMembership` + `UpdateMembership` | Одна-дія-один-audit (Lo17); ручний revoke дзеркалить `BlockAuthor` — order→(user,course)→той самий перехід. Webhook-шлях — НЕ це (див. "Нове") |
| Інбокс-дедуп події | `payment_events` UNIQUE(provider,event_id) + `Apply` | AC5 ідемпотентність безкоштовно; рефанд — ще один тип у наявному інбоксі |
| Доменна диспетчеризація події | `DecidePaymentEvent` | Уже приймає `(event, order, alreadyApplied)` → рішення; додаємо гілку, не новий механізм |
| Вебхук-транспорт (verify→apply→ack) | `callbackHandler` + `VerifyCallback` | `callback.Event != nil → ApplyPaymentEvent` вже є; рефанд наповнює `Event`, роут не змінюється |
| Admin-черга UI (список + resolve + revoke, infinite-query, keyset) | `ParticipantReportQueue` + `admin.participant-reports.tsx` + `enrollment-queries` | Та сама форма: адмін-список з рядковими діями |
| Admin-authz | `administrator` middleware у `RegisterAdmin` | AC7 — той самий middleware |
| Resolve-семантика | `ResolveParticipantReport` (`UPDATE ... WHERE state='open'`, no-op) | G1 — дзеркалимо ідемпотентний no-op |

**Нове (і чому наявне не підійшло):**
- `PaymentRefunded` тип + `PaymentEventRevoke`/`PaymentEventIgnore` дії — наявні `paid`/`failed`
  закривають settle, не reversal. `DecidePaymentEvent` отримує рефанд-гілку **перед** наявним
  amount-guard (payment_event.go:44, що вимагає `amount>0` і `amount==order.amount`), бо all-or-nothing
  (G9) свідомо НЕ звіряє суму — частковий `Refunded`/нульовий `Voided` не має падати в reject.
- **Webhook-revoke SQL у `Apply` — нова, НЕ reuse `UpdateMembership`.** Інлайн у транзакції `Apply`:
  `UPDATE course_grants SET state='revoked' WHERE order_id=$1 AND state IN ('active','blocked')`
  (scoping за order_id, G10) + `INSERT grant_audit(action='revoke', payment_event_id, actor_id=NULL,
  reason='wayforpay_refund')`. Інший shape, ніж `UpdateMembership` (нема actor, є payment_event_id,
  scoping за order_id не за id). Прецедент інлайн-аудиту в `Apply` є (видача гранта, repository.go:650).
  **grant_audit вставляти ЛИШЕ коли `UPDATE ... RowsAffected()==1`** (як `UpdateMembership`
  repository.go:521) — інакше повторний `Voided` на вже-`revoked` гранті (A19) напише зайвий
  revoke-audit на нуль змінених рядків.
- `orders.refund_reason` колонка — сьогодні причина лягає лише в неструктурований `slog.Warn`;
  екран без неї не рятує від ручного SQL (G3).
- `RefundQueue`/`ResolveRefund`/`RevokeForRefund` service+port — немає читача `refund_required` і
  немає дії над ним; це той самий геп, що фіча закриває. `RefundQueue` читає `orders` LEFT JOIN
  `payment_events` (type='paid') для `charged_amount_minor`/`received_at` (G12).
- `orders.refund_resolved_at`/`refund_resolved_by` колонки — наявний `refund_required bool` не несе "розглянуто".

## Сайд-ефекти

- **Стан гранта:** вебхук-рефанд та ручний revoke переводять `course_grants.state` → `revoked`.
  Це змінює `CanConsume`/`CanCommunicate` (контракт access) — покупець втрачає доступ до медіа-байтів
  (nginx `auth_request`), плеєра, транскрипту, коментарів. Саме намір фічі, але зачіпає **всі**
  модулі-читачі членства → верифікація повним e2e (не таргетним), за gotcha
  `integration-tag-skipped-by-go-test-and-e2e`.
- **Інбокс:** новий рядок `payment_events` на кожен прийнятий рефанд — зростання таблиці, звичне.
- **Audit:** новий `grant_audit` рядок на кожен revoke (системний/ручний).
- **`course_grants_one_per_user_course`:** revoke не вставляє рядок, лише `UPDATE state` — unique
  не зачіпається (глобально-лічена сутність: правило "хто додає — прибирає" не активується, бо ми
  не додаємо грант).

## План міграції (00017_refund.sql)

Up:
```sql
ALTER TABLE orders
    ADD COLUMN refund_reason text NULL
        CHECK (refund_reason IS NULL OR refund_reason IN ('settlement_rejected','grant_conflict')),
    ADD COLUMN refund_resolved_at timestamptz NULL,
    ADD COLUMN refund_resolved_by uuid NULL REFERENCES users(id);
```
- Усі три NULL-able, без backfill: наявні `refund_required=true` рядки (якщо є на тест-даних)
  лишаються з `refund_reason=NULL` — екран покаже їх із причиною "невідомо" (історичні до фічі).
  Нових даних на проді нема (тест-режим), тож backfill не потрібен.
- `refund_required` колонка вже існує (додана в **00012**_payment_checkout_refunds) — не чіпаємо.

Down: `DROP COLUMN refund_reason, refund_resolved_at, refund_resolved_by`.

**Перевірка migration-preservation** (gotcha `seed-sql-must-load-on-downgraded-schema`): seed не
вставляє в нові колонки; DownTo нижче 00017 має лишати `orders` валідним. Пінуємо тестом.

## Rollback

- Фіча за прапорцем не ховається (pre-launch, ріж сміливо — `no-prod-deploy-test-mode`). Відкат —
  revert коміту + `goose down` до 00016.
- Вебхук-гілка рефанду до відкату: якщо вже прийняті рефанд-події в `payment_events` — після
  відкату `VerifyCallback` знову поверне `Event: nil` на `Refunded`/`Voided` (стара поведінка),
  наявні рядки подій лишаються нешкідливими записами інбоксу. Гранти, вже revoked рефандом, не
  "воскресають" — відкат коду не чіпає даних; за потреби admin re-grant (`GrantCourse`, наявний).

## Конституція (AGENTS.md)

- **Domain чистий:** `PaymentRefunded`/`PaymentEventRevoke`/`DecidePaymentEvent` — лише stdlib +
  `internal/contracts`. `refund_reason` НЕ домен-стан (персистенція), у домен не протікає. ✔
- **Application → domain+ports:** `RevokeForRefund`/`RefundQueue` кличуть порти, не адаптери. ✔
- **Межі модулів:** усе в `orders`; жодного імпорту іншого модуля. Access-контракт
  (`CanConsume`/`CanCommunicate`) уже читає `course_grants.state` — ми лише міняємо стан, контракт
  не чіпаємо. ✔
- **Nginx медіа:** не чіпаємо; revoke природно закриває `auth_request` (стан гранта). ✔
- **Generated models в адаптерах:** sqlc/Orval мапінг явний; `check-generated` зловить дрейф
  (gotcha — стейджити ПЕРЕД гейтом). ✔
- **Explicit error codes:** `grant_transition_invalid` (конкурентний revoke), `order_not_found`,
  `refund_already_resolved`? — ні, resolve ідемпотентний no-op (G1), помилки не треба. ✔
- **No secrets in logs; URLs without query:** warning-логи AC6 несуть лише id, не PII/суми картки. ✔

## Приймальні сценарії

Делеговано `acceptance-spec`. Групи: happy path (A1–A4), межі (A5–A9), відмови (A10–A15).

| # | Сценарій | Передумова | Дія | Очікуваний спостережуваний результат | Чим перевіряється | AC |
|---|---|---|---|---|---|---|
| A1 | Список розбіжностей | 2 замовлення з `refund_required=true`, не resolved: одне `settlement_rejected`, одне `grant_conflict` | Адмін відкриває `/admin/refunds` | У списку рівно 2 рядки; кожен показує курс, користувача, `order_amount_minor`, `charged_amount_minor`, `received_at` і причину (`settlement_rejected`/`grant_conflict`) | e2e `refund.spec.ts`; HTTP contract `handler_test.go` | AC1 |
| A2 | Ручний revoke з екрана | Замовлення в черзі, у покупця грант `active` | Адмін тисне "Відкликати доступ" на рядку | Грант стає `revoked`; покупець втрачає доступ до медіа/плеєра; рядок **лишається** в черзі (revoke і resolve незалежні — черга фільтрує за `refund_required AND NOT resolved`), зникає лише після окремого "Позначити розглянутим" (A3) | e2e `refund.spec.ts`; integration `repository_integration_test.go`; app `service_test.go` | AC2 |
| A3 | Позначити розглянутим | Замовлення в черзі (revoke не потрібен) | Адмін тисне "Позначити розглянутим" | `refund_resolved_at`/`refund_resolved_by` виставлені; рядок зникає зі списку "потребує уваги" | e2e `refund.spec.ts`; integration `repository_integration_test.go` | AC3 |
| A4 | Вебхук-рефанд відкликає доступ | `paid` замовлення з грантом `active`; приходить WayForPay callback `transactionStatus=Refunded` (валідна сигнатура) | (системна, не користувач) callback POST | Відповідь `200` + accept-ACK; грант стає `revoked`; `payment_events` має рядок типу `refunded`; `grant_audit` рядок `action='revoke'`, `actor_id=NULL`, `reason='wayforpay_refund'`, `payment_event_id` заповнений | integration `repository_integration_test.go`; unit domain `payment_event_test.go`; HTTP `handler_test.go` | AC4 |
| A5 | Порожня черга | Жодного `refund_required=true` не-resolved замовлення | Адмін відкриває `/admin/refunds` | 0 рядків і текст порожнього стану ("Немає замовлень, що потребують уваги") | e2e `refund.spec.ts` | AC1 |
| A6 | Рефанд відкликає `blocked` грант | `paid` замовлення, грант `blocked` (адмін призупинив раніше) | callback `Voided` | Грант стає `revoked` (а не лишається `blocked`); audit-рядок як A4 | integration `repository_integration_test.go`; unit domain `payment_event_test.go` | AC5 |
| A7 | `Voided` трактується як `Refunded` | `paid` замовлення, грант `active`; callback `transactionStatus=Voided` | callback POST | Той самий результат, що A4 (обидва → `PaymentRefunded`) | unit `wayforpay` adapter test; unit domain `payment_event_test.go` | AC4 |
| A8 | Повторний вебхук (ретрай WayForPay) | Рефанд уже прийнято (грант `revoked`, подія в інбоксі); той самий callback приходить вдруге (однаковий `event_id`) | callback POST повторно | Відповідь `200` + ACK; жодного нового рядка в `payment_events`/`grant_audit`; стан гранта незмінний `revoked` | integration `repository_integration_test.go` | AC5 |
| A9 | Resolve повторно (ідемпотентний no-op) | Замовлення вже resolved | Адмін (або ретрай) тисне "Позначити розглянутим" ще раз | Відповідь успішна; `refund_resolved_at`/`by` **не** перезаписані; жодної помилки | integration `repository_integration_test.go`; HTTP `handler_test.go` | AC3 |
| A10 | Рефанд на не-`paid` замовлення | Замовлення `failed`; приходить callback `Refunded` | callback POST | Відповідь `200` + ACK; стан замовлення/гранта **незмінний**; **без** `refund_required`; у логах structured warning (order_id, course_id, user_id, transactionStatus) | integration `repository_integration_test.go`; unit domain `payment_event_test.go` | AC6 |
| A11 | Рефанд без гранта (replay на revoked) | `paid` замовлення, грант уже `revoked` (адмін встиг); callback `Refunded` | callback POST | Відповідь `200` + ACK; 0 рядків оновлено (no-op, не помилка); стан гранта лишається `revoked` | integration `repository_integration_test.go` | AC5 |
| A12 | Конкурентний revoke | Замовлення в черзі, грант `active`; інший адмін revoke-нув між читанням і дією | Адмін тисне "Відкликати" на застарілому стані | Помилка `grant_transition_invalid` (409); UI показує, що стан змінився; грант не чіпається вдруге | HTTP contract `handler_test.go`; app `service_test.go` | AC2 |
| A13 | Неадмін на екрані/дії | Автентикований не-адмін | GET `/admin/refunds` або POST revoke/resolve | `403` з кодом помилки; жодної зміни стану | HTTP contract `handler_test.go` | AC7 |
| A14 | Revoke на замовлення без живого гранта | Замовлення в черзі, гранта нема взагалі (конфлікт був на вставці, грант так і не створено) | Адмін тисне "Відкликати" | Кнопки revoke нема (UI); API повертає `grant_invalid`/`grant_transition_invalid` при прямому виклику | HTTP contract `handler_test.go`; app `service_test.go` | AC2 |
| A15 | Migration-preservation | Схема на 00017; seed завантажено | `goose down` до 00016, потім `up` | `orders` валідний на обох кроках; seed проходить без 42P01/CHECK-порушень на нових колонках | integration migration-preservation test | — (міграція) |
| A16 | Невалідна сигнатура рефанду | `paid` замовлення, грант `active`; callback `Refunded` із **кривою** сигнатурою | callback POST | Відповідь `401` `invalid_callback`; грант лишається `active`; жодного рядка в `payment_events`/`grant_audit` | HTTP contract `handler_test.go`; unit `wayforpay` adapter test | AC8 |
| A17 | Частковий рефанд теж відкликає | `paid` замовлення (сума 50000), грант `active`; callback `Refunded` із сумою 10000 (частковий) | callback POST | Грант стає `revoked` (сума не звіряється — all-or-nothing G9); не падає reject | integration `repository_integration_test.go`; unit domain `payment_event_test.go` | AC4, AC9 |
| A18 | Рефанд не чіпає admin re-grant (order_id scoping) | Покупка revoked адміном, потім admin re-grant (новий грант `active`, source=admin, order_id=NULL); callback `Refunded` на старе purchase-замовлення | callback POST | 0 рядків оновлено (живий грант має order_id=NULL ≠ цей order); admin-грант лишається `active`; ACK 200 | integration `repository_integration_test.go` | AC5, G10 |
| A19 | Refunded потім Voided (різні статуси) | `paid` замовлення, грант `active`; приходять обидва callback — спершу `Refunded`, тоді `Voided` (різні event_id) | два callback POST | Перший → `revoked`; другий → no-op на вже-`revoked` (0 рядків, не помилка); два рядки в `payment_events`, один `revoke` в `grant_audit` | integration `repository_integration_test.go` | AC5 |
| A20 | Resolve на нефлагованому замовленні | Замовлення без `refund_required` (ніколи не флаговане) | Пряма POST resolve на цей orderID | Ідемпотентний no-op (0 рядків), не помилка (дзеркалить ParticipantReport) | HTTP contract `handler_test.go`; integration `repository_integration_test.go` | AC3 |

Ручних рядків: **0** (усі сценарії покриті автоматизованим рівнем).
