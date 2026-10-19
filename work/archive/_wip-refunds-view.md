# V-track діагностика — view 8 `#refunds` (RefundQueue)

**Гілка:** `feat/refunds-view-designpass` (від main @bdcb3ce). **Автор:** Courses 1. **Старт:** 2026-10-18.

## ФАКТ: прод уже card-queue на спільному `.admin-queue-*` каркасі
`RefundQueue` (`admin/refunds`, admin-only, модуль **orders**) — уже `<ul.admin-queue-list>` з
`.admin-queue-card`: `.admin-queue-heading` (icon-tile refund-гліф + eyebrow orderId·receivedAt +
h2 reason + **membershipState-pill**), `.admin-queue-details.admin-refund-summary` (**4-col**:
course/user/ordered/charged — БАГАТШЕ за proto), `.admin-queue-actions` (Revoke-діалог **gated на
active|blocked** + Resolve). CSS `.admin-refund-summary/.admin-refund-amount` вже є (main.css 3087).
Порожній стан — `.admin-queue-empty` icon-tile. Тобто каркас = дзеркало reports, вже V-track-ний.

## ФАКТ: proto `#refunds` (src.jsx case 'refunds')
PageHeader eyebrow «PAYMENTS & REFUNDS» + title «Refunds» + description + actions=Pill «N pending»;
QueueFilter **All/Pending/Refunded/Declined**; `.refund-card`: queue-heading (icon-tile + eyebrow
id·date + h2 **person** + status-pill), `.refund-summary` (course/email + великий `.refund-amount`),
`.report-body` reason, queue-actions (policy-note + Revoke→Declined + Resolve→Refunded, обидва gated
на Pending). Empty + example-tools (demo-only). Icon-tile tint = **info** (спільне правило
`.queue-heading>.ui-icon-tile{color:var(--color-info)}`).

## РОЗБІЖНОСТІ
- **D1** PageHeader (eyebrow+title+**description**[+count-pill]) — прод має bare `.admin-list-heading`
  (eyebrow+h1, без опису й лічильника). Головний гап, той самий що в reports.
- **Фільтр стану** — proto має (4 вкладки), прод ні.
- icon-tile info-tint — proto має (`color:var(--color-info)`), прод default (muted).
- **Status-pill — РІЗНА ВІСЬ**: proto pill = **refund-outcome** (Pending/Refunded/Declined);
  прод pill = **membershipState** (active/blocked/revoked), що гейтить кнопку Revoke. Це не те саме.
- `.refund-summary` 2-поля (course/email + amount) vs прод `.admin-refund-summary` 4-поля
  (course/user/ordered/charged) — **прод багатший, лишаємо**.

## ⚠️ КЛЮЧОВА ЗНАХІДКА: backend-черга single-state, outcome не персиститься
`orders.RefundQueue` SQL = `WHERE o.refund_required AND o.refund_resolved_at IS NULL`.
`ResolveRefund` → `SET refund_resolved_at=now()` (випадає з черги, ідемпотентно).
`RevokeForRefund` → та сама membership-транзиція, що roster-revoke; **окрема** від resolve
(не ставить refund_resolved_at). Тобто:
1. Немає персистованого «Refunded vs Declined» — є лише `refund_required` + `refund_resolved_at`.
2. resolve і revoke **ортогональні** (не взаємовиключні термінальні стани, як у proto).
Отже proto-фільтр All/Pending/Refunded/Declined проти поточного бекенду беззмістовний
(Refunded/Declined завжди порожні, All==Pending) — **точно ситуація reports**, але з додатковим
вивертом: proto 3-way мапиться ще гірше, бо prod-модель не має 3-значного outcome.

## SCOPE — ПОТРЕБУЄ РІШЕННЯ ВЛАСНИКА (дзеркало reports-розвилки)
- **A — layout-only:** PageHeader D1 (+count-pill «N очікують») + icon info-tint; черга лишається
  single-state, membershipState-pill і 4-col summary — без змін. Без backend. Найменший.
- **B — reports-parallel 2-state filter:** backend повертає й resolved (`?state=pending|resolved|all`),
  QueueFilter Усі/Очікують/Опрацьовані + PageHeader D1. Чистий еквівалент reports-рішення C
  (`refund_resolved_at IS NULL` vs `IS NOT NULL`; збирає proto Refunded+Declined в одне «Опрацьовано»).
  Backend(orders)+frontend+ордерний гейт.
- **C — proto-faithful 3-state:** персистити refund-outcome (refunded vs declined), фільтр 4-way.
  Найбільший і семантично сумнівний (resolve/revoke ортогональні, не взаємовиключні). НЕ рекомендую.

## РІШЕННЯ scope (2026-10-18, Courses 1, вибір власника через AskUserQuestion)
**Scope C — proto-faithful 3-state**, модель **Decouple / 3 дії** (другий AskUserQuestion).
Закриває A (layout-only) і B (2-state). outcome персиститься окремо від membership-revoke.

### Модель (owner 2026-10-18, Decouple)
- Нова колонка `orders.refund_outcome text NULL CHECK (NULL OR IN ('refunded','declined'))`
  (міграція 00020). `NULL` ⟺ на розгляді (поки `refund_required AND refund_resolved_at IS NULL`).
- Pending-картка → **3 дії**: **Повернути** (outcome=refunded), **Відхилити** (outcome=declined),
  **Відкликати доступ** (існуючий `RevokeForRefund`, membership, gated на active|blocked,
  **незалежний** від outcome). Refund/Decline ставлять outcome+resolved_at+resolved_by разом →
  картка покидає pending. Revoke лише змінює грант, outcome не чіпає.
- Термінальна картка (refunded/declined, під фільтрами Повернені/Відхилені/Усі) — **без дій**
  (grammar reports/proto: pill — вся історія). Revoke-доступу теж лише поки pending.
- Status-pill = **refund-outcome** (На розгляді/Повернено/Відхилено); membershipState →
  контекст-поле в summary (досі був pill + гейтив revoke).
- Ендпоінти: **replace** `/resolve` на `/refund` + `/decline` (pre-launch cut-freely);
  `/revoke` лишається. `GET /admin/refunds?state=all|pending|refunded|declined` (absent→pending,
  інше→400).
- Idempotent: refund/decline діють лише поки `refund_resolved_at IS NULL` (0 рядків = no-op 204).

### Слайси
1. **Backend(orders):** міграція 00020; ports `RefundQueue(ctx,state)` + `SettleRefund(ctx,order,actor,outcome)`
   (замість ResolveRefund) + RefundQueueItem.Outcome; app `RefundQueue(actor,state)` +
   `RefundOrder`/`DeclineRefund`; SQL WHERE per-state + select outcome + SettleRefund UPDATE;
   http `?state` parse + `/refund`+`/decline` routes + DTO.outcome. Оновити orders-тести.
2. **OpenAPI+Orval:** GET state-enum; refund/decline paths (replace resolve); item.outcome. regen, check-generated 0.
3. **Frontend:** domain RefundStateFilter + item.outcome; ports+adapter refundQueue(state)/refundOrder/declineRefund;
   queries per-state key + family-invalidate; route sparse ?state (pending default);
   RefundQueue.tsx SegmentedControl 4-way + outcome-pill + 3 actions(pending)/none(terminal) +
   membershipState у summary; локалі uk/en.
4. **e2e:** оновити refund.spec (resolve→refund/decline); новий кейс перемикання на Повернені/Відхилені
   показує settled без дій. Повний ордерний make check (persisted-state → full e2e, не targeted).

## ВЕРИФІКОВАНО + ЗДАНО (2026-10-18, Courses 1)
Модель C (Decouple/3 дії) реалізовано, доведено, відкрито PR.
- **Backend:** міграція 00020 (`refund_outcome`); `RefundQueue(ctx,state)` WHERE `($1='' OR
  ($1='pending' AND refund_resolved_at IS NULL) OR refund_outcome=$1)`; `SettleRefund(outcome)`
  замість ResolveRefund; app RefundOrder/DeclineRefund; http `?state` (absent→pending, інше→400) +
  `/refund`+`/decline`. Orders unit+contract+integration зелені (Rf1 перевіряє pending-вихід +
  refunded-фільтр + idempotent-незмінність outcome).
- **OpenAPI+Orval:** state-enum, 2 нові paths, `outcome` nullable; regen, check-generated 0.
- **Frontend:** domain RefundOutcome/RefundStateFilter+outcome; adapter/port refundQueue(state)/
  refundOrder/declineRefund; queries per-state key [...base,state] + family-invalidate; route sparse
  ?state (pending default); RefundQueue SegmentedControl 4-way + outcome-pill (warning pending/
  success refunded/neutral declined) + membershipState у summary + 3 дії pending/none terminal;
  локалі uk/en. e2e refund.spec: revoke→refund→зникає з pending→«Повернені» показує settled без дій.
- **Гейт:** ордерний `make check` **GATE_EXIT=0, e2e 121 passed (16.4m)**. Лог
  `frontend/test-logs/refunds-view-gate-20261018T032056Z.log`. 1-й прогін — зелений без ретраїв.
- **Браузер (1440):** PageHeader D1 + count-pill «1 на розгляді»; SegmentedControl Усі/На розгляді/
  Повернені/Відхилені (role=radio); pending-картка «Доступ уже був» pill «На розгляді» + Доступ:Активний
  + 3 дії; клік «Повернені»→`?state=refunded` показує «Повернено» (success) БЕЗ дій.
- **PR #44** проти main: https://github.com/b2vv/courses/pull/44 (коміт deec70a). Далі — A-батч.

### ПАСТКА нагадування
- Після .go-правок `gofmt -w` перед гейтом (format-check крок 1; oxfmt лише frontend).
- check-generated вимагає `git add` згенерованого (інакше червоне без дрейфу).
- Dev-БД: 2 фейкові refund_required-замовлення вставлено для скріна — гейт їх скине; не комітити.
- Код фонового гейта читати з GATE_EXIT у лозі.
