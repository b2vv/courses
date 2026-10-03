# Оплата — жива нота (assumptions / розвилки)

Фаза: оплата (§24 п.21) — остання. Гілка `feat/payment`. Роль: Courses 1 — spec/plan/gates/commits;
код — Codex (gpt-5.6-sol); рев'ю — Courses 2. Obsidian MCP відпав → нота тут (CLAUDE.local.md rule #4).

## Рамки (§24 п.22, ухвалено 2026-09-28 — не переглядаються)
- Разова покупка → безстроковий grant на курс (не підписка).
- Один провайдер (перший і єдиний адаптер). Одна валюта. Повернення — не в MVP.
- Перевірка фази (§24 п.21): entitlement = одне поле в access.CanView + один join на модуль; без полів
  терміну; рівно один адаптер провайдера; один тип валюти; повернення в «не робимо».

## Шви (docs/readiness/spec.md §5) — блюпринт
- grant → поле Entitled у фактах access + клауза `|| entitled` у CanView + join у 3 модулях (guard D4).
- Модуль `orders`: замовлення зі знімком ціни, гроші (мінорні+валюта), вебхук-inbox (унікальність за
  event id), видача права = 1 транзакція, аудит. Каталог: позначка «платний»+ціна лише для показу.

## Setup
- Гілка feat/payment, дім docs/payment/. HEAD бази: a1d3898 (main, після мержу інструктора + брифінгу).

## ДАЛІ
- GATE 1 (grilling): ПЕРШЕ питання — провайдер і юрисдикція (⚠️ §24 п.22). Далі валюта, checkout-модель,
  джерело grant поза покупкою, UX повернення до вебхука. Доменна частина (grant/orders/гроші) —
  провайдер-незалежна, спеку не блокує.

## GATE 1 — рішення (grilling)
- **Юрисдикція (користувач 2026-09-29): Україна → Європа → далі розширення.** Наслідок: провайдер за
  портом/адаптером у orders; розширення = нові адаптери потім, не зміна домену. MVP — 1 провайдер + 1
  валюта (§24 п.22). Мульти-валюта і VAT/tax для ЄС — задокументовані ШВИ на майбутнє, НЕ MVP (YAGNI).
- **Провайдер: конкретний ще не названо.** Україна-first звужує: Stripe погано онбордить укр.продавців;
  кандидати LiqPay/WayForPay/Fondy (укр. PSP) або merchant-of-record (Paddle) для ЄС-VAT. Блокує лише
  адаптер+checkout+вебхуки, НЕ провайдер-незалежний домен. Рішення — до слайса адаптера.
- ДАЛІ grilling: валюта (Україна-first → UAH?), checkout-модель (redirect типово для укр.PSP),
  джерело grant поза покупкою (адмін-видача в MVP?), UX повернення до підтвердження вебхуком.

- **Валюта: UAH** (Україна-first; тип грошей несе валюту → EUR потім = дані, не перепис).
- **Джерело grant: покупка + адмін-видача.** Grant видається (а) оплаченим замовленням source=purchase,
  (б) ручною адмін-видачею source=admin (без order_id). Обидва в аудиті. Додається адмін-ендпоінт+UI.
  У межах швів §5 (source=purchase/gift/admin). order_id у grant — nullable (адмін-шлях без замовлення).

- **UX повернення: «оплата в обробці».** Право видає ЛИШЕ вебхук (джерело істини); редірект-повернення
  НІКОЛИ не видає право (підробний). Замовлення pending → сторінка «обробляємо», доступ з'являється по
  вебхуку. Наслідок: не грантити на redirect-callback; grant тільки в транзакції вебхука.
- checkout-модель (redirect vs embedded) — типово redirect для укр. PSP; фінал за провайдером (deferred).

## GATE 1 підсумок (оплата)
Рішення: провайдер deferred (Україна-serving, 1 адаптер); валюта UAH; grant = покупка+адмін-видача;
право видає лише вебхук (UX «в обробці»); повернення/підписки/мульти-валюта/VAT — поза MVP (шви).
Провайдер-незалежний домен (grant/entitlement/orders/гроші/аудит) спеку не блокує; блокується лише
адаптер+checkout+вебхук-парсер до вибору провайдера.

## План (крок 4)
- Звірка коду: мапінгів доступу 4 (не 3, як шви §5); join на право — 6 запитів у 5 файлах, у 4 треба
  прокинути userID через порти; гарячий шлях — nginx auth_request (`loadVideoCourse`).
- ЗНАЙДЕНО дірку швів §5: `|| entitled` без факту «платний» не закриває опублікований платний курс.
- Рішення користувача: ціна на курсі (NULL = безкоштовний); власник+адмін завжди; прев'ю-уроки в MVP.
- Правило: admin ∨ owner ∨ entitled ∨ (published ∧ (¬paid ∨ preview)); ADR 014 доповнює 013.
- Зрізи O01–O07; O06 (адаптер провайдера) заблоковано на назві провайдера; фаза закривається лише після O06.
- ДАЛІ: GATE 2 (критика + захист свіжими сабагентами).

## GATE 2 — прогін 1 провалився
- 2 сабагенти (критика, захист) запущено → 0 повернулось: обидва впали на API safeguard
  (`reasoning_extraction`, Opus 5.5). Робочі файли порожні — не бігали; вердиктів немає, гейт НЕ пройдено.
- Повтор: той самий обсяг, нейтральні формулювання (цілісність/валідація замість «атака/підробка»),
  модель sonnet. Делегування ролі не замінене інлайн-ом — вердикт і далі від свіжого сабагента.
- Прогін 2 (sonnet) зупинено на прохання користувача (хоче Opus 4.8). Прогін 3 (аліас `opus`) → API
  показав `claude-opus-5-5`: аліас opus ≠ 4.8; обидва впали на тому ж safeguard (2 запущено, 0 повернулось).
- Для 4.8 створено `~/.claude/agents/gate-reviewer.md` (`model: claude-opus-4-8`); тип агента
  підхоплюється лише новою сесією.
- GATE 2 делеговано сесії **Courses 2** (рев'ю, на Opus 4.8) через SendMessage — critique+defense на
  docs/payment/plan.md; просив підтвердити модель і нейтральні формулювання. Чекаю вердикт назад у Courses [035346].

## GATE 2 (Courses 2, Opus 4.8) — ЧЕРВОНИЙ, знахідки закрито
- Модель рецензента підтверджено: claude-opus-4-8. 2 blocking + 3 non-blocking, усі — правки плану.
- B1: спільний access.PurchaseRequired, вибір коду на 4 шарах (A6). B2 (рішення користувача): grandfather (A7, Pay26).
- NB1: один предикат (A6). NB2: Pay16 за станом active будь-якого source. NB3: Pay21b unknown-order не термінальна.
- Чисте: покриття доступу повне, A2 ок, O01 нейтральний.
- ДАЛІ: Courses 2 перепровіряє закриття; якщо зелено — tasks.

## GATE 2 ЗЕЛЕНО (повторна звірка Courses 2, claude-opus-4-8)
- Обидва blocking закриті, усі NB адресовані. 3 залишкові NB (не гейтять) прийнято:
  NB1' → property-тест PurchaseRequired⟹¬CanView (O01/Pay27); NB5 → O07 рахує рядки не діапазон;
  NB4 → grandfather на playback-прогресі, названа межа (вузька підмножина transcript/discussion-only).
- GATE 2 ПРОЙДЕНО. ДАЛІ: tasks (O01–O07 → Pay-рядки), implement Codex (sol); O06 блокований на назві провайдера.

## O01/O03 зчеплення по integration (знайдено при верифікації O02)
- O01 розширив accesstest.TruthTable рядками paid/entitled/preview. Integration-guard D4 (R06/R07/R08 +
  discussion) ганяє таблицю через SQL — а SQL віддасть нові факти лише в O03. Тому між O01 і O03
  integration ЧЕРВОНИЙ на нових рядках (напр. R06 stranger_paid_published). Міграція 00011 застосовується
  (тест падає на асерті, не на setup). На feature-гілці проміжний red припустимий; O03 закриває.
- Урок: O01 «поведінково нейтральний» — лише про прод+unit; integration-guard зчеплений із SQL-слайсом.
  O03 мусить: (а) SQL читає price/grants/preview; (б) integration-сіди створюють платні курси+гранти для
  нових рядків. Повний integration зелений — критерій O03.

## 2026-09-30 — O05 prerequisites (Courses 1, reviewed by Courses 2)

- **Retry after a failed payment needs a NEW idempotency key.** `StartPurchase` replays
  any order already stored under a key — pending, paid or failed — as the same answer
  (Pay12). A client that retries a failed purchase with the old key gets the failed
  order back. The O05 UI must mint a fresh key for "Спробувати ще раз".
- **Replay of a paid key answers `already_owned`,** not the paid order: the owner check
  runs before the idempotency lookup, and paid ⟹ active grant in production (one
  transaction). Pay12 accepts that answer.
- **First price set with explicitly stated previews** no longer fails
  (`PreviewSet` presence no longer breaks the price-only comparison); a preview
  CHANGE together with the first price is still refused. The O05 pricing editor may
  send `isPreview` with the first price.

## 2026-10-01 — The learning outline of a paid course (owner decision; found as a defect)

**Decision (owner, 2026-10-01):** a signed-in viewer without access sees the **whole
structure** of a published paid course on its learning page — preview lessons open,
every other lesson locked with "Купити" — whether or not the course has a preview
lesson. Lesson content (video, media bytes, transcript, discussion) stays gated per
lesson on the server, unchanged.
**Alternative closed:** showing the outline only when the course has a preview lesson.

**Defect it replaces:** the learning course took `Preview` from the FIRST row of a
one-row-per-lesson query, i.e. the first lesson of the first module. A paid course
whose preview lesson was not first refused the whole learning page with
`course_purchase_required`, so the preview lesson was unreachable (Pay11). The
integration test inserted one-lesson courses only, where first = only. The four
content mappings (playback, media `auth_request`, transcript, discussion) each read
their own lesson's `is_preview` and were not affected.

**Shape:** the outline mapping passes the preview fact as true to `access.CanView`
(the rule stays in the contract, ADR 013/014); each lesson carries `isPreview` and a
server-computed `locked` = `access.PurchaseRequired` with that lesson's preview fact,
so the client never re-implements owner/admin/entitled.

**How to check:** a published paid course with three lessons, preview on the second:
an unentitled student gets 200 with the outline, lesson 2 `isPreview:true, locked:false`,
lessons 1 and 3 `locked:true`; the owner, an admin and an entitled student see no
lesson locked; a draft stays refused.

## 2026-10-01 — The learning course never carried its price (defect, found by the O05 slice-2 journey)

**Defect:** `GET /api/v1/learn/courses/{slug}` declares a required `price`, but the learning
query selected only `price_minor IS NOT NULL` (the paid flag), so `price` was `null` for every
course. Nothing read it until the purchase panel did; the browser journey found a disabled
«Купити курс» with no amount. **Fix:** the query reads `price_minor` and `currency` and maps
them through `domain.NewCourse` (which keeps `Paid == (Price != nil)`). **How to check:**
integration `Pay11 preview on a later lesson…` asserts the stored 50000 UAH on the outline;
e2e `purchase.spec.ts` asserts the amount on the locked lesson.

## 2026-10-01 — Purchase UI shape (O05 slice 2, Courses 1)

- A locked lesson mounts the purchase panel **instead of** the player: no playback session,
  transcript or comments request is made for it (unit route test and e2e probe both count
  them). Alternative closed: mount the player and let the server refuse — it costs three 403s
  per visit and shows error states where an offer belongs.
- The order page lives at `/orders/$orderId?course=<slug>`, not under `/learn/`: under
  `/learn/$` a course with the slug `orders` would collide with it. **How to check:** the route
  manifest test R24 lists `/orders/$orderId` next to `/learn/$`.
- A paid order invalidates the course outline; it is re-read when the course is opened again,
  so its lessons unlock without a reload. "After the event" (Pay13) cannot be proved end to
  end before O06 brings a provider event; until then the e2e stops at the pending state.
- **For slice 3 (admin pricing):** the server refuses a first price that arrives together with
  any other change (it grandfathers existing learners in its own path — `authoring.go`
  `firstPriceSet`). The admin form must save the first price as a separate request.
  **How to check:** a save with a new structure and a first price returns 400
  `invalid_request`; the same two changes in two saves return 200 twice.

## 2026-10-01 — How O06 proves a full purchase (owner decision)

**Decision (owner, 2026-10-01):** the automated sandbox journey proves Pay25 through a **server
status check**: after the buyer returns from WayForPay, our server asks WayForPay for the order's
status (signed `CHECK_STATUS` request) and, when it is approved, applies the payment through the
same `ApplyPaymentEvent` as the callback. Plus **one recorded run through a temporary tunnel**
(cloudflared) in which WayForPay's real `serviceUrl` callback reaches our endpoint, is verified
and applied; its evidence goes in the phase report. The callback signature is also tested on
vectors (Pay24).
**Why:** WayForPay cannot reach `localhost`; a tunnel in every gate run would make the gate
depend on an external tunnel and network. The status check is also a production feature: a
delayed or lost callback (WayForPay retries for up to 4 days) no longer leaves a paid buyer
waiting.
**Alternatives closed:** status check only (the main production path never exercised end to
end); a tunnel in every e2e run (flaky gate).
**Unchanged (A5):** the browser's return never grants anything — only a signature-verified
callback or a signed server-to-server status answer does.
**How to check:** the e2e completes a test-card payment on WayForPay's test page and the lesson
opens with no callback able to reach the stack; the tunnel run's log shows the callback request
(signature verified, applied once, a replay a no-op).

## 2026-10-01 — O06c split: offline stub in the gate, live tunnel run by hand (owner decision)

**Decision (owner, 2026-10-01):** O06c has two parts.
- **O06c-auto (in the gate):** the e2e drives the purchase against a **local WayForPay stub**, not
  the real provider, so the ordered gate stays offline and deterministic. The stub serves the pay
  page target and the CHECK_STATUS / callback endpoints. It proves: checkout form built and
  submitted; a signed callback on a test vector applied once (Pay24 at the HTTP layer); the pending
  status page runs CHECK_STATUS → the **same** `ApplyPaymentEvent` → the lesson unlocks (Pay25 shape,
  against the stub); a replayed callback changes nothing.
- **O06c-live (by hand, owner):** one recorded run through a cloudflared tunnel where WayForPay's
  **real** `serviceUrl` callback reaches the stack, is signature-verified and applied once. Courses 1
  writes a runbook (`work/runbook-o06c-tunnel.md`); the owner runs it with the real test-merchant
  secret key (kept only in the owner's local `.env`, never in the repo) and returns the log/screenshot;
  Courses 1 records it in the phase report as the production-path proof.
**Seam:** the WayForPay adapter's `PayURL`/`StatusURL` become env-configurable, **defaulting to the
production hosts**; only the e2e compose stack points them at the stub. The frontend action-host
allowlist stays `secure.wayforpay.com` in production and takes the stub host only from an env the
e2e sets — production is unchanged.
**Why:** a tunnel in every gate run makes the gate depend on an external tunnel, the network and a
human completing a card form; the owner cannot put the live secret key in CI. The stub exercises our
own code paths end to end; the live run proves the one thing a stub cannot — that the real provider's
callback reaches and satisfies us.
**How to check:** the gate's e2e passes with no network to WayForPay; the runbook's recorded run
shows the real callback (signature verified, applied once, replay a no-op) and the lesson open.

## 2026-10-02 — B-25: a repeat checkout starts a new order; a double payment leaves an explicit trace

**Decision (Courses 1 on Courses 2's money-path analysis, 2026-10-02; the owner was away and may
overturn it):** when a pending order's checkout was already opened, «Перейти до оплати» starts a
**new order** (new idempotency key, new `orderReference`) instead of re-submitting the old one —
option (b) of backlog B-25. **Alternative closed:** (a) a per-attempt `orderReference` on one order.
**Mechanism that closes it:** `DecidePaymentEvent` rejects any event for a non-pending order and
`Apply` fails the order on Declined/Expired; with several attempts on one order, a late Declined of
attempt 1 fails the order and the Approved of attempt 2 is then rejected — money taken, no grant.
Avoiding that needs attempt-level state (a new table, model and `Apply`); (b) isolates each attempt in
its own order for free.
**Also (applies today, independent of B-25):** two paid orders for one course (two tabs) — the second
`Apply` marks its order paid while the grant insert hits `ON CONFLICT DO NOTHING`, silently. A paid
order without a grant must carry an explicit, queryable refund trace.
**How to check:** integration — a second checkout of an opened order creates a new order with a new
reference; a late Declined of the old order leaves the new one untouched; two Approved orders → one
grant and the second order flagged for refund.

## 2026-10-02 — Pay25 Approved leg: owner's real card on the owner's test-mode shop (owner decision)

**Decision (owner, 2026-10-02):** WayForPay publishes no test card that yields `Approved` (search of
the wiki, help centre and official SDK/plugin READMEs; four cards tried live, all declined or refused).
The Approved → grant → callback leg of Pay25 is proven with the owner's **own WayForPay shop in test
mode** and the owner's **real card for 1 UAH**, paid by the owner. **Alternatives closed:** waiting for
WayForPay support; merging with the leg unproven. **Unconfirmed:** the 15-minute automatic refund in
test mode comes from a third-party guide (Weblium), not WayForPay — the owner checks the refund in the
cabinet. **How to check:** the runbook run shows `Approved` from the signed CHECK_STATUS, an active
`purchase` grant, the lesson open, and the `serviceUrl` callback answered 200.

## 2026-10-03 — Pay25-live done; SecretKey rotation deferred to after the one-off test charge (owner decision)

**Decision (owner, 2026-10-03):** the live Approved run used the shop's existing `WAYFORPAY_SECRET_KEY`
even though it had appeared in chat, rather than rotating it first. **Alternative closed:** regenerate the
key and change the merchant password before the run (the fully by-the-book order), which the owner judged
not worth the delay for a single 1-грн charge on a shop in test mode (low blast radius: a leaked test-mode
key can at most forge callbacks for that test shop; no real funds, and the shop is archived/idle after).
**Obligation it creates:** regenerate the SecretKey and change the merchant password **before any further
live use**; the password is not used by the project and is stored nowhere. **How to check:** the live
proof is recorded (callback `POST → 200`, order `4396fba5…` → `paid`, one grant/one event — report §6);
a later `git grep` shows no key in the repo, and `.env` stays gitignored.

## 2026-10-03 — Test shop `site_ex` to be deleted; key rotation dropped (owner decision, supersedes the entry above)

**Decision (owner, 2026-10-03):** the owner deletes the `site_ex` test-mode shop used for Pay25-live,
so the SecretKey that appeared in chat dies with it. **Alternative closed:** regenerating the key and
changing the merchant password (the obligation in the previous entry) — moot once the shop no longer
exists. A future live run (deployment, Р1) uses a new shop with a new key. **How to check:** in the
WayForPay cabinet the `site_ex` shop is gone; a signed request with the old key fails. The local `.env`
lines for `site_ex` then point at nothing and may be removed.
