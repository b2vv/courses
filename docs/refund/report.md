# Refund handling — report

Дельта до [plan.md](plan.md). Фаза в роботі на гілці `feat/refund-handling`.

## 1. Верифікація по слайсах

- **R1 (міграція 00017 + `refund_reason`):** `make integration` EXIT=0 (повний прогін у
  golang-контейнері). `orders/adapters/postgres` ok (два наявні тести розширено: rejected-paid →
  `refund_reason='settlement_rejected'`, duplicate-payment → `'grant_conflict'`);
  `course/adapters/postgres` ok (migration-preservation — повний DownTo→Up цикл покриває 00017,
  seed не торкається нових колонок). Build/vet/gofmt чисто. sqlc-дрейфу нема (orders — raw SQL,
  не в query-файлах). Закриває A15.

## 2. Обмеження (named)

- Refund-before-paid reordering (spec §Named limitations): реконсиляції нема, поза scope.

## 3. Наступний таск

- R2 — домен: `PaymentRefunded` тип + `DecidePaymentEvent` рефанд-гілка перед amount-guard.
