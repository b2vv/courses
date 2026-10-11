# Refund handling — spec

## Проблема

Два пов'язані гепи, названі в `docs/payment/report.md` §6 і `work/backlog.md` B-26, обидва
відкриті з моменту закриття Payment-фази (2026-10-03):

1. **`orders.refund_required` не має консумента.** Коли гроші приходять без видачі грантa —
   оплачений курс, уже виданий тим самим юзером раніше (`already-granted`), прийнятий платіж по
   замовленню, що вже провалилось, або платіж на чужу суму (Pay21) — репозиторій ставить
   `orders.refund_required = true` і пише structured warning у лог. Немає admin-екрана чи API:
   знайти такі замовлення можна лише ручним SQL-запитом (документованим у звіті фази). Жодної дії
   над записом — "я бачив, розібрався" — не існує; флаг висить вічно.
2. **B-26: WayForPay `Refunded`/`Voided` після `Approved` ігноруються.** Якщо продавець (власник)
   повертає гроші покупцю прямо в кабінеті WayForPay (не через наш API — такого шляху в нас нема),
   WayForPay шле вебхук зі `transactionStatus = Refunded` або `Voided`. Наш адаптер
   (`wayforpay.go:151`) приймає сигнатуру лише для `Approved`/`Declined`/`Expired` — для решти
   статусів callback повертає `Event: nil` і ми просто ACK'аємо без жодної зміни стану. Грант
   лишається `active`: покупець, якому повернули гроші, продовжує мати доступ до курсу.

Обидва геми стосуються грошей і доступу — саме тієї межі, яку AGENTS.md називає "access authority"
(`course_grants`, одна правдива таблиця). Мовчазний розрив між "гроші повернулись" і "доступ
скасовано" — це про довіру до продукту, не просто незручність для адміна.

## Користувацькі сценарії

**С1 — адмін бачить список замовлень, що потребують уваги.**
Адмін заходить на новий екран (за аналогією з `admin.participant-reports.tsx`) і бачить список
замовлень із `refund_required = true`: курс, користувач, сума (списана, не замовлена — Pay21-кейс),
коли прийшов платіж, причина (яка з трьох умов спрацювала). Список не порожній тільки тоді, коли
є реальна розбіжність — на звичному шляху цей екран завжди пустий.

**С2 — адмін вручну відкликає грант за розбіжним замовленням.**
З рядка списку адмін бачить активний грант (якщо він є) і натискає "Відкликати доступ" — нова
order-scoped дія `RevokeForRefund(actor, orderID)`, тонка обгортка над тим самим
`transitionMembership`, що roster admin-revoke (resolve order→(user,course)→той самий перехід,
дзеркалить `BlockAuthor`). Якщо гранта нема (replay на вже revoked user) — кнопки revoke нема,
лишається тільки "Позначити розглянутим".

**С3 — адмін позначає запис розглянутим без revoke.**
Якщо адмін вирішив, що розбіжність не потребує відкликання (наприклад, Pay21 — платіж на іншу
суму, але курс лишається оплаченим правильно) — позначає запис "resolved". Запис зникає зі списку
"потребує уваги", але не видаляється — `refund_resolved_at`/`refund_resolved_by` лишаються на
записі назавжди (аудит). Нотатка не обов'язкова (G1); повторний resolve — ідемпотентний no-op.

**С4 — WayForPay сам повертає гроші (рефанд через кабінет WayForPay) → доступ скасовується
автоматично.** Вебхук з `transactionStatus = Refunded` або `Voided` приходить на замовлення, що
раніше було `paid` і видало грант. Система: (а) записує подію в `payment_events` (інбокс, як і
`paid`/`failed` — ідемпотентно за `orderReference:transactionStatus`), (б) якщо грант **саме цього
замовлення** (scoping за `order_id`, G10) ще `active`/`blocked` — відкликає його (`GrantRevoked`,
`grantedBy = nil`, системна дія, не адмінська); суму рефанду **не звіряє** — будь-який підписаний
Refunded/Voided на paid-замовлення відкликає доступ повністю (all-or-nothing, G9), бо курс більше
не "оплачений повністю", (в) ACK'ає WayForPay як завжди. Якщо грант уже `blocked`/`revoked` (адмін встиг раніше)
— подія лише логується, без повторного переходу (ідемпотентність, не помилка).

## Acceptance criteria

- AC1: ендпоінт/читання повертає всі записи з `refund_required = true`, що ще не `resolved`;
  кожен рядок несе courseID, userID, замовлений `order_amount_minor` (з `orders.amount_minor`),
  фактично списаний `charged_amount_minor` і `received_at` (обидва з рядка `payment_events` типу
  `paid` для цього замовлення — JOIN `payment_events e ON e.order_id=o.id AND e.type='paid'`; якщо
  такого рядка нема — наприклад reject стався до settle — поля null/0), і **причину** — одне з двох
  значень, які код фактично розрізняє: `settlement_rejected` (reject на `DecidePaymentEvent` —
  stale-order і foreign-amount сьогодні не розрізняються) / `grant_conflict` (уже є живий грант
  іншого походження). Нова колонка `orders.refund_reason`, заповнювана в тих самих двох місцях,
  що вже ставлять `refund_required` — детекція не змінюється (Out of scope). Історичні рядки,
  флаговані до цієї фічі, мають `refund_reason=NULL` → екран показує "невідомо".
- AC2: дія "відкликати" доступна тільки коли поточний membership-стан — `active` або `blocked`;
  виконує той самий `Revoke`, що roster (через `transitionMembership`, order-scoped обгортка
  `RevokeForRefund`, `reason='refund_queue:'+orderID`), і той самий код помилки
  (`grant_transition_invalid`), якщо стан уже змінився паралельно (конкурентний адмін).
- AC3: дія "позначити розглянутим" ставить `resolved_at`/`resolved_by`; **ідемпотентний no-op** на
  вже-resolved (0 рядків, не помилка — дзеркалить `ResolveParticipantReport`); нотатка **не**
  обов'язкова; окремого перемикача "показати розглянуті" в UI нема (історія — прямим запитом до БД).
- AC4: `Refunded`/`Voided` вебхук на `paid`-замовлення → грант **саме цього замовлення** (scoping
  за `order_id`, G10) переходить у `revoked` без участі адміна (всередині транзакції `Apply`,
  атомарно з інбокс-вставкою); суму рефанду **не звіряє** (all-or-nothing, G9 — частковий `Refunded`
  теж відкликає); подія видна в тому ж `payment_events` інбоксі, що `paid`/`failed`; audit-рядок
  `action='revoke'`, `payment_event_id=<подія>`, `actor_id=NULL`, `reason='wayforpay_refund'`.
- AC5: рефанд відкликає грант із **будь-якого живого** стану — `active` **або** `blocked`
  (`WHERE order_id=$1 AND state IN ('active','blocked')`); no-op лише коли грант уже `revoked`, або
  його взагалі нема, або `order_id` на живому гранті вже не цей (admin re-grant дав order_id=NULL —
  той грант не чіпаємо, G10) — 0 рядків, не помилка. Повторний вебхук (WayForPay ретраїть)
  дедуплюється на рівні EventID інбоксу, як існуючий `paid`/`failed` шлях.
- AC6: `Refunded`/`Voided` на замовлення, що не `paid` (наприклад, `failed`/`pending`) — **тихий
  no-op ACK зі structured warning-логом** (order_id, course_id, user_id, transactionStatus), без
  зміни стану й без `refund_required`. Домен повертає окрему дію (`PaymentEventIgnore`), НЕ reject —
  reject для `PaymentSucceeded` ставить `refund_required`, а тут це нормальний WayForPay-side шум.
- AC7: екран-список + обидві дії (revoke, resolve) доступні лише адміну (той самий middleware, що
  `admin.participant-reports.tsx`/roster revoke); неадмін отримує `403`.
- AC8: вебхук із невалідною сигнатурою / кривим тілом / невідомим `merchantAccount` — `401`
  `invalid_callback`, **без** зміни стану (наявна поведінка `VerifyCallback`→handler:521-525;
  рефанд-гілка її не послаблює).
- AC9: підписаний `Refunded`/`Voided` на `paid`-замовлення **завжди** або відкликає грант (AC4),
  або, якщо грант уже не живий/не цей order_id, — no-op (AC5). Не існує шляху, де підписаний рефанд
  на paid-замовлення тихо ACK'ається, лишивши живий грант цього замовлення `active` (закриває діру
  захисту Q1: all-or-nothing + окрема `PaymentEventRevoke`-гілка ПЕРЕД amount-guard `DecidePaymentEvent`
  прибирає reject-через-суму для рефанду).

## Out of scope

- Автоматичний рефанд *від нас* до покупця (ініціювати повернення коштів через WayForPay API) —
  тут тільки реакція на вже зроблений рефанд і видимість розбіжностей. Платити назад гроші —
  окрема інтеграція (WayForPay refund API), не ця фіча.
- `RefundInProcessing`/`Pending`/антифрод-статуси WayForPay — не міняють стан замовлення сьогодні
  (acknowledged, no change) і лишаються так; тільки `Refunded`/`Voided` закривають B-26.
- Email-сповіщення покупця про відкликання доступу через рефанд — існуючий `Revoke` сьогодні не
  шле емейл (перевірити на кроці plan; якщо шле — цей шлях його теж отримає безкоштовно, якщо не
  шле — лишається не шле, без нової роботи).
- Групові операції (revoke кількох записів одночасно) — один рядок, одна дія, як
  participant-reports queue.
- Будь-яка зміна `Pay21`/already-granted/stale-order логіки виявлення — вона лишається точно
  такою, як є; ця фіча лише додає читання й дію над тим, що вона вже пише.

## Рішення GATE 1 (2026-10-11, grilling; підтверджено власником)

- G1 (В1 + AC3): resolve без обов'язкової нотатки, ідемпотентний no-op на вже-resolved, без
  перемикача "показати розглянуті" — дзеркалить наявний `ResolveParticipantReport`.
- G2 (AC6 + Q3): `Refunded`/`Voided` на не-`paid` замовлення — тихий no-op ACK зі structured
  warning-логом, без зміни стану, без `refund_required`.
- G3 (Q4 + AC1): нова колонка `orders.refund_reason` (enum `settlement_rejected` / `grant_conflict`),
  заповнювана в тих самих двох точках, що вже ставлять `refund_required`; детекція не змінюється.
- G4 (Q2a): `Refunded`+`Voided` → один доменний `PaymentRefunded` тип через той самий інбокс і
  `DecidePaymentEvent`, нова доменна дія `PaymentEventRevoke`.
- G5 (Q2b): перехід гранта для вебхука — всередині транзакції `Apply` (атомарно з інбокс-вставкою),
  не через application compare-and-set port.
- G6 (В2+В3 + Q2d): атрибуція лише в `grant_audit`+`payment_events` — системний revoke:
  `action='revoke'`, `payment_event_id`, `actor_id=NULL`, `reason='wayforpay_refund'`; ручний:
  `actor_id=<admin>`, `reason='refund_queue:'+orderID`. Жодного нового `GrantSource` чи домен-поля.
- G7 (AC5 + Q2c): рефанд відкликає з `active` **і** `blocked`; no-op лише на вже-`revoked`.
- G8 (Q2e + AC2): ручний revoke — тонка order-scoped `RevokeForRefund(actor, orderID)` поверх
  наявного `transitionMembership` (дзеркалить `BlockAuthor`), ті самі коди помилок.

## Рішення GATE 2 (2026-10-11, critique+defense; підтверджено власником)

- G9 (critique #1 + defense Q1): **all-or-nothing** — будь-який підписаний `Refunded`/`Voided` на
  paid-замовлення відкликає доступ повністю, сума рефанду НЕ звіряється. `DecidePaymentEvent`
  отримує окрему `PaymentRefunded`-гілку **перед** amount-guard, щоб частковий рефанд чи `Voided`
  із сумою 0 не падали в reject. Усуває "тихий ACK лишає грант active".
- G10 (critique #6): webhook-revoke scoping за **`order_id`** (`WHERE order_id=$1`), не за
  (user,course). Admin re-grant після revoke (order_id=NULL) рефанд старого purchase-замовлення не
  чіпає — свідоме рішення адміна живе.
- G11 (critique #4): не-paid рефанд → окрема доменна дія `PaymentEventIgnore` (warn+ACK, без
  `refund_required`), відмінна від reject.
- G12 (critique #2): `charged_amount_minor`/`received_at` екрана читаються JOIN'ом з
  `payment_events` (рядок типу `paid`), не з `orders` (там нема такої колонки).

## Named limitations (defense Q, critique #8)

- **Refund-before-paid reordering:** якщо `Refunded`/`Voided` прийде РАНІШЕ за `Approved` (WayForPay
  переставив доставку — рідко, зазвичай `Approved` перший), не-paid-гілка (AC6) зробить warn+no-op,
  а пізніший `Approved` видасть доступ попри повернення грошей. Реконсиляції нема — поза scope цієї
  фічі; механізм: WayForPay практично завжди шле `Approved` першим, тож вікно теоретичне. Названо
  тут, аби не вважати закритим.

## Відкриті питання

(Усі закрито на GATE 1/GATE 2 — див. рішення G1–G12 вище.)
