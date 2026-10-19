# Review request — O07 gap tests: Pay17, Pay26 + grandfather idempotency (Codex + Courses 1)

Base `98499db`. Uncommitted. Context: writing the phase report row by row (the Phase 8 lesson —
count rows, don't trust a range) exposed two rows the plan claims at `integration` with NO test.
Codex wrote the tests; Courses 1 fixed four fixture/isolation defects the gate then caught.

## Rows closed
- **Pay17** — price change after purchase. New integration subtest: buy at 50000 → paid + grant;
  raise `courses.price_minor` to 70000; assert the active grant still exists AND the order row's
  amount_minor is still 50000 (the snapshot holds). Was NOT FOUND before (only a domain snapshot
  unit).
- **Pay26** — grandfather over N students + repeat no-duplicate. New subtest: two progress students
  on a free course → GrandfatherOnPriceSet → both get an active `source=grandfather` grant (count=2);
  a second GrandfatherOnPriceSet with the same price → still 2, no error. The old test seeded one
  student and never repeated.

## CODE CHANGE — please scrutinize (money/grant path)
`repository.go` `GrandfatherOnPriceSet`: the first UPDATE guards on
`updated_at=$expected AND price_minor IS NULL`. On 0 rows it used to always return
`ErrCourseModified`. Codex made a repeat idempotent: on 0 rows it now reads the current price and,
**if it already equals the requested price/currency**, commits as a no-op (no new grants);
a different price still returns `ErrCourseModified`.
- Why: the Pay26 repeat calls the repository op directly and must not error.
- The thing to weigh: this widens "success" — a call with a STALE expectedUpdatedAt now succeeds
  (no-op) when the price already matches, instead of ErrCourseModified. In production the authoring
  first-price path only calls this when `stored.Price == nil`, so the repeat path isn't reached
  there; the change is reachable only by a direct repeat. Alternative not taken: keep the code and
  have the test drive the repeat through the authoring service. Your call on whether the code change
  is the right seam or the test should change instead.

## Four defects Courses 1 found and fixed (gate-caught, not in Codex's sandbox)
1. new fixtures inserted users with an invalid `password_hash` → used the repo's Argon2id fixture.
2. Pay17 subtest had 2 ids but StartPurchase+ApplyPaymentEvent consume 3 → added a third.
3. pay26 id collided with the new pay17 id → moved.
4. the main body's audit count was global (`payment_event_id IS NOT NULL`); the new Pay17 paid
   event leaked into it → scoped the count to this order's event (the "globally-counted" rule).

## Gate
`make integration` exit 0 (Pay17, Pay26, and the whole orders package green).

## Findings (Courses 2 appends below)

---
## Findings — Courses 2 (claude-opus-4-8, 2026-10-02)

### CODE CHANGE GrandfatherOnPriceSet no-op — БЕЗПЕЧНО, правильний шов
RowsAffected()==0 означає: (a) курс не існує, (b) updated_at розійшовся, (c) price_minor уже не NULL.
No-op гілка — ЛИШЕ коли поточна ціна/валюта == запитаній; розбіжність → ErrCourseModified (optimistic
lock тримається для реальної зміни).
Безпека money/grant:
- No-op commit-ить ДО grant-INSERT → гранти не вставляє. Якщо ціна вже == запитаній, її виставив ПЕРШИЙ
  виклик, який ТАКОЖ програнтив: null→value — ЄДИНИЙ шлях через GrandfatherOnPriceSet (ordinary save
  міняє лише наявну ціну, ніколи null→value; підтверджено authoring.go). Тож no-op коректно не ре-грантить:
  пропущеного гранту нема, подвійного нема (+ON CONFLICT backstop). 
- Stale expectedUpdatedAt no-op-ить лише коли кінцевий стан = наміру; інакше ErrCourseModified. Конкурентний
  identical-price set → другий no-op (перший програнтив). Інваріант грошей цілий.
Чому це правильний шов (а не «тест через authoring»): authoring кличе grandfather ЛИШЕ при stored.Price==nil;
після першого set ціна non-null → другий authoring-save НЕ грандфазерить. Повтор GrandfatherOnPriceSet
через authoring НЕДОСЯЖНИЙ. Тож прямий репо-тест + ідемпотентність op — єдиний спосіб довести Pay26
«повтор без дублю». У проді op не повторюється (authoring-gate); зміна = defense-in-depth для прямих
retry/дублів, безпечна.
[nb-мінор] Неіснуючий курс: SELECT→ErrNoRows→persistenceError (ErrDependencyUnavailable) замість старого
ErrCourseModified. Недосяжно в проді (grandfather кличеться по завантаженому курсу). Error-код зсунувся — ок.
[nb] Додати 1 рядок коментаря на no-op гілці: «repeat з тією ж ціною = no-op (перший уже set+grant;
прямі caller-и/retry не мусять падати)» — щоб майбутній читач не думав, що прод це повторює.

### Pay17 / Pay26 тести — доводять те, що треба
Pay17: buy 50000→paid+grant; raise→70000; grant активний І order.amount_minor=50000 (snapshot тримається). ✓
Pay26: 2 progress-студенти→grandfather→count=2 active source=grandfather; repeat same price→still 2, no error. ✓
Обидва були обіцяні плdéfаном на integration, тесту не було (урок Phase 8 — рахуй рядки). Закрито.

### 4 виправлені дефекти — патерн правильний
Головне #4: audit-count був глобальний (payment_event_id IS NOT NULL) → новий Pay17 paid-event протікав →
проскоуплено за подією ЦЬОГО замовлення. Це рівно правило «globally-counted → скоупь вибірку» (моя NB з
GATE2, повторюваний клас). Решта (Argon2id-фікстура, 2→3 id, id-колізія) — isolation-гігієна. Гейт exit 0.

## ПІДСУМОК: 0 blocking
Зміна GrandfatherOnPriceSet безпечна (no-op лише коли стан=намір, грант не губиться/не дублюється,
optimistic lock тримається для реальної зміни) і правильний шов (альтернатива недосяжна через authoring-gate).
Pay17/Pay26 закривають дві діри рядкової перевірки. nb: 1 рядок коментаря на no-op. Комітити O07-тести можна;
report.md по рядках + PR далі (Pay25-live лишається відкритим до тунельного прогону — правильно).
