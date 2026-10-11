# Refund handling — report

Дельта до [plan.md](plan.md). Фаза в роботі на гілці `feat/refund-handling`.

## 1. Верифікація по слайсах

- **R1 (міграція 00017 + `refund_reason`):** `make integration` EXIT=0 (повний прогін у
  golang-контейнері). `orders/adapters/postgres` ok (два наявні тести розширено: rejected-paid →
  `refund_reason='settlement_rejected'`, duplicate-payment → `'grant_conflict'`);
  `course/adapters/postgres` ok (migration-preservation — повний DownTo→Up цикл покриває 00017,
  seed не торкається нових колонок). Build/vet/gofmt чисто. sqlc-дрейфу нема (orders — raw SQL,
  не в query-файлах). Закриває A15.

- **R2 (домен `PaymentRefunded` + рефанд-гілка):** `77f4f61`. `go test ./internal/orders/domain`
  зелений — нові рефанд-кейси (частковий→revoke, void-0→revoke, non-paid→ignore, dup→no-op,
  wrong-order→reject) + збережені paid/failed/reject. Гілка стоїть перед amount-guard; envelope-
  валідація розділена. Sentinel наявного тесту "refund"→"chargeback". Домен A4/A6/A7/A10/A17.

- **R3 (wayforpay адаптер):** `3d4c40b`. `go test ./internal/orders/adapters/wayforpay` зелений —
  Refunded/Voided→PaymentRefunded (EventID orderReference:status, різні рядки інбоксу), сигнатура
  перед мапінгом (AC8), InProcessing/tamper/wrong-key збережені. Адаптер A7/A16.

- **R4 (repository `Apply` рефанд-гілка):** `4b04c63`. `make integration` EXIT=0 (повний прогін,
  orders/adapters/postgres ok). Нові кейси: Rf4 (revoke+system-audit actor_id NULL/reason/
  payment_event_id), Rf17 (частковий→revoke), Rf6 (Voided на blocked→revoked), Rf19 (Refunded→Voided:
  1 revoke, 2 події), Rf8 (інбокс-дедуп), Rf10 (non-paid→ignore, без refund_required), Rf18 (admin
  re-grant під NULL order_id не чіпається). Audit лише при RowsAffected==1 (RETURNING). Знайдено+
  виправлено під час слайсу: рефанд через пакетний `service` вичерпував фіксований 10-id слайс
  (panic index-out-of-range) → усі рефанд-виклики переведено на svc з unbounded subtestIDs.
  Закриває A4,A5,A6,A8,A10,A11,A17,A18,A19 (repository-рівень).

- **R5+R6 (application+ports+repository):** `0c755a2` (злютовані — зміна інтерфейсу OrderRepository).
  Порт `RefundQueueItem` + OrderByID/RefundQueue/ResolveRefund; сервіс RefundQueue/ResolveRefund/
  RevokeForRefund (admin-only `ErrGrantForbidden`; revoke через transitionMembership
  reason=`refund_queue:`+orderID, дзеркалить BlockAuthor). Repo: RefundQueue LEFT JOIN
  payment_events(type=paid) для charged/received (nil якщо нема), фільтр refund_required AND NOT
  resolved; ResolveRefund ідемпотентний (WHERE resolved_at IS NULL). Unit green; integration Rf1
  (queue+idempotent resolve, resolved_at не рухається на повторі) + Rf20 (no-op на нефлагованому +
  OrderByID). HTTP-фейк доповнено методами (build). A1/A2/A3/A9/A20 (service/repo).

- **R7 (HTTP + OpenAPI/Orval):** `5252f2d`. 3 admin-роути (GET refunds / POST resolve 204 / POST
  revoke 200) за `administrator` middleware; OpenAPI +3 операції + RefundQueueItem/RefundQueuePage
  (charged/received nullable, reason enum з історичним ''); Orval regen, sqlc без дрейфу. Contract-
  тести зелені проти специфікації: student→403 усі три, список повертає item, resolve→204+actor,
  revoke→200 revoked, unknown→404. A1/A2/A3/A12/A13/A14/A20 (HTTP). Resolve повертає 204 (не 200).

## 2. Обмеження (named)

- Refund-before-paid reordering (spec §Named limitations): реконсиляції нема, поза scope.

## 3. Наступний таск

- **R7 (HTTP) `b64d2a1`, R8 (frontend) `775b5c1`** закрито фінальним ordered `make check` EXIT=0:
  e2e **115 passed** (12.8m, 0 failed/0 did-not-run), `refund.spec.ts` green (A1 список+axe, A2 revoke
  через діалог, A3 resolve чистить чергу), FE unit 688. refund.spec створює `grant_conflict`
  справжнім шляхом (purchase+admin-grant-while-pending+Approved callback). A5 (empty-state) —
  покрито рендером компонента (гілка `items.length===0`), не окремим e2e-асертом: спільна БД не
  гарантує порожню чергу під час прогону.
- **Наступне:** code-review дифу гілки + ship-report + PR у main.
