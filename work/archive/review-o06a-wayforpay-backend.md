# Review request — O06a: WayForPay backend (Codex; Courses 1 gated)

Base `9e239a6`. Uncommitted. Brief: work/brief-codex-o06a-wayforpay-backend.md; Codex log:
work/codex-o06a.log (+ work/codex-o06a-exec*.log).

## Gates (Courses 1, outside Codex's sandbox)
`make generate format-check lint typecheck test integration` → exit 0 (real Docker Postgres).
Not run: e2e (O06c). Generated headers changed because the API description no longer says
purchases are not implemented.

## Rounds
1. Codex implemented port, adapter, routes, failed event, status check, config, OpenAPI.
2. Courses 1 found: Pay21b defect (unmatched callback was answered `accept` → WayForPay stops
   retrying → payment lost) and the brief's HTTP/Postgres contract tests missing. Fixed: unmatched
   → 503, no acknowledgement; tests added.
3–4. Integration failures found only in Courses 1's Docker run: fixture user with an invalid
   password_hash; then the new subtest drew ids from the test's shared ten-id slice and starved a
   pre-existing subtest (AGENTS "globally counted" rule) — now subtest-local ids.

## Wiki vector
The wiki's printed purchase hash (`b9593278…`) does not match its own input/key; Courses 1
recomputed independently (Python hmac) → `4941ea0c…`, which the test pins with the wiki's input.
`b9593278…` is the merchantSignature from the wiki's P2P example — a documentation copy error.
Final proof of signing = the test merchant accepting our form (O06c).

## Secrets
`.env.example`: `WAYFORPAY_MERCHANT_ACCOUNT=test_merch_n1`, `WAYFORPAY_SECRET_KEY=` (empty),
public base URL. No key in code (grep for the published test key: none).

## Your criteria 1–7 (as corrected) — please check against the code
Especially: signature verified before any DB call (handler: VerifyCallback → 401 before
ApplyPaymentEvent); duplicate → signed accept, no second grant; unmatched → 503 (retry);
CHECK_STATUS through the same ApplyPaymentEvent; failed event terminal; returnUrl applies nothing.
Also look at: checkout `returnUrl` = `/orders/<id>` without `?course=` (O06b can add it); the
status-check bound on GET.

## Findings (Courses 2 appends below)

---
## Findings — Courses 2 (claude-opus-4-8, 2026-10-01)

ДАЛІ: критерії 1–7. Пріоритет: (1) підпис ПЕРЕД будь-яким DB-дотиком (bad→401, 0 рядків);
(2) duplicate→signed accept, без 2-го гранту; (3) HMAC_MD5 на векторах; (4) ключі лише env/не в логах;
(5) CHECK_STATUS через той самий ApplyPaymentEvent; (6) failed термінальний; (7) returnUrl нічого не
застосовує. + Pay21b (unmatched→503 без ack). Читаю handler callback-flow + wayforpay adapter підпис.

## BLOCKING [B-O06a] — data race на statusAt: fatal panic під конкурентним пол-інгом статусу
де: service.go:26 (statusAt map), :87 (read), :91 (write) — у PurchaseStatus; Service будується
РАЗ у main.go:80 (синглтон, спільний на всі HTTP-запити); жодного mutex (grep: нема sync.).
чому: net/http обслуговує кожен запит у своїй горутині. GET /api/v1/learn/orders/{orderId}
читає й пише service.statusAt БЕЗ синхронізації. Frontend (зріз 2) ПОЛЛИТЬ цей GET кожні 2с під
час pending. Два пол-інги одночасно (два юзери в «оплата в обробці», або одна людина в двох
вкладках, або overlap поллера) → concurrent map read+write → Go runtime `fatal error: concurrent
map read and map write` → КРАХ УСЬОГО API-процесу (не лише запиту). Go-мапи не безпечні навіть для
РІЗНИХ ключів. Троттл (3с) узагалі має сенс лише на синглтоні — тобто синглтон і задуманий.
чому гейти не зловили: integration/unit послідовні; -race ловить лише гонки, що РЕАЛЬНО сталися в
прогоні; конкурентного PurchaseStatus тест не робить. Латентно. Порушує й AGENTS «No global mutable
clients/stores».
що зробити: обгорнути read-check-write у sync.Mutex (або RWMutex); або sync.Map; плюс (вторинне,
non-blocking) statusAt росте безмежно (ніколи не евіктить) — memory leak, прибрати разом (напр.
евікція за віком під тим самим локом).
вага: BLOCKING — fatal-паніка під штатним конкурентним навантаженням (поллінг — гарячий шлях).

## Критерії 1–7 + Pay21b — усі ЧИСТІ
1. Підпис ДО DB: VerifyCallback (чистий adapter) → bad→401 ПЕРЕД ApplyPaymentEvent → 0 рядків;
   validCallbackSignature правильний порядок полів WayForPay; constantTimeHexEqual (анти-timing). ✓
2. Duplicate → signed accept без 2-го гранту: Apply duplicate-check no-op → falls through до Acknowledge. ✓
3. HMAC_MD5: hmac(md5, secret) `;`-join; вектор 4941ea0c незалежно переобчислено (wiki-хеш — doc-помилка
   P2P-прикладу); фінальний доказ — O06c (тест-мерчант прийме форму). ✓
4. Ключі лише env (FromEnvironment); жодного секрету/підпису/payload у логах (generic-помилки);
   .env.example secret порожній, ключа в коді нема. ✓
5. CHECK_STATUS через ТОЙ САМИЙ ApplyPaymentEvent: PurchaseStatus→provider.OrderStatus→ApplyPaymentEvent. ✓
6. Failed термінальний: Decide(failed)→Apply, repository.go оновлено → order='failed'+provider_ref+commit,
   подія записана, гранту нема. Declined-потім-Approved: order не pending→terminal reject (без гранту). ✓
7. returnUrl нічого не застосовує: checkout будує returnUrl=/orders/<id> (без ?course, O06b додасть),
   гранту не видає; статус-GET може запустити ПІДПИСАНИЙ CHECK_STATUS (легітимно). ✓
Pay21b: unmatched→503 без ack (WayForPay ретраїть) — коментар пояснює; виправлення peer на місці. ✓

## ПІДСУМОК: 1 BLOCKING (statusAt race), money/safety-критерії чисті
Підпис, idempotency, failed, CHECK_STATUS-переюз, секрети, A5 — усе коректно й обережно. Єдиний
блокер — незахищена спільна мапа statusAt на синглтоні під поллінг-шляхом: fatal-паніка. Fix: mutex
(+ евікція). Вектор підпису фінально доведе O06c (тест-мерчант).

## B-O06a fix (Codex round 5; verified by Courses 1)
`statusMu sync.Mutex` guards eviction + check-and-set in one critical section; entries older than
the 3 s window are evicted under the same lock; the provider call stays outside the lock.
Test `TestPurchaseStatusThrottleIsSafeUnderConcurrentPolling` (32 goroutines). Courses 1 ran it
under `-race`: green; mutation (locks removed) → `WARNING: DATA RACE`, FAIL; restored.
Gates after the fix: `make format-check lint typecheck test integration` exit 0.

## RE-CHECK B-O06a фіксу (round 5) — ЗАКРИТО
Звірив service.go:
- statusMu sync.Mutex (:28). Критична секція :90-102 — Lock → евікція (delete старших за 3с у
  range, що Go дозволяє) → read statusAt[orderID] → throttled гілка Unlock+return (:98) → інакше
  set statusAt[orderID]=now → Unlock (:102). ВСІ read+write statusAt лише між Lock/Unlock (grep:
  91/93/96/101 — усі в секції). Обидва unlock-шляхи покриті (нема missing unlock/deadlock).
- provider.OrderStatus (:103) — ПІСЛЯ Unlock: мережевий виклик поза локом (нема утримання лока
  через I/O). ApplyPaymentEvent не чіпає statusAt.
- Евікція за віком під тим самим локом → мапа більше не росте безмежно (вторинний memory-leak знято).
- now() взято раз (:89) для евікції й throttle — консистентно.
Верифікація peer: 32-горутинний тест зелений під -race; мутація (зняти локи)→DATA RACE FAIL; відновлено
— правильна мутаційна дисципліна (тест падає без фіксу, зеленіє з ним). Гейти exit 0.
ВЕРДИКТ: B-O06a закрито. O06a commit-ready — усі 7 критеріїв + Pay21b чисті, гонка усунена.
