# Review — O05-prep backend (authoring first-price-set + orders idempotency)
Reviewer: Courses 2 (claude-opus-4-8, 2026-09-30). Uncommitted `git diff HEAD -- backend/`
on feat/payment, base d5543fd. Взято з stash@{0} (WIP мертвої сесії) як основа O05.

ДАЛІ: розібрав authoring.go (first-price-set preview fix) і orders/service.go (idempotency),
питання peer (a) paid/failed replay vs Pay12, (b) created.CourseID()!=courseID досяжність+гроші.

## F1 [authoring.go first-price-set] — ЧИСТО
Save: original=pristine loaded tree (scalar-копія; Modules-slice від'єднується пізніше через
stored.Modules=tree.Modules, тож original.Modules лишається завантаженим). На firstPriceSet
вимагається DeepEqual(original, storedForCompare) — тобто контент НЕ змінено; ціна застосовується
ОКРЕМО через grandfatherer.GrandfatherOnPriceSet (не в tree-save). Фікс clearPreviewSet знімає лише
presence-флаг PreviewSet (не значення Preview) на копії storedForCompare → presence-шум не валить
рівність, але реальна зміна Preview-ЗНАЧЕННЯ досі → DeepEqual fail → ErrInvalidRequest. clearPreviewSet
робить глибоку копію (нові слайси, значення-лессони) → stored не мутується. Тести: red-before
(TestFirstPriceSetAcceptsUnchangedPreviewsSpecifiedByTheEditor) + guard
(TestFirstPriceSetStillRefusesAPreviewChange) — правило «лише ціна» ціле. Коректно.

## F2 [orders (a) idempotency replay] — ЧИСТО, за Pay12
Зміна: повтор ключа повертає наявне замовлення БУДЬ-якого статусу (було: reject non-pending).
Pay12 (plan.md:174) = «той самий Idempotency-Key → одне замовлення, та сама відповідь». Стара
поведінка (reject non-pending) ПОРУШУВАЛА це. Нова — правильна idempotency: ключ = токен дедуп-запиту;
failed лишається під тим ключем, retry = НОВИЙ ключ → нове замовлення (модель Stripe). 
NB(O05-frontend): UI ретраю після failed мусить ГЕНЕРУВАТИ новий idempotency-key, інакше отримає
той самий failed order назад. Це фронт-контракт для O05, не бек-баг.
Нюанс Pay12: для paid+granted курсу повтор ключа дає ErrCourseAlreadyOwned (CanStartOrder РАНІШЕ за
idempotency-перевірку), а не сам paid order — безпечно й зрозуміліше, але літерально «та сама відповідь»
з Pay12 трохи розходиться. Юніт «same key paid returns paid order» тестує paid-БЕЗ-grant — стан,
недосяжний у проді (paid⟹grant у тій самій Apply-транзакції). Нешкідливо, але тест-реалізм: реальна
paid-replay-відповідь = already_owned. Звір, чи Pay12 контракт-тест пінить order-відповідь чи приймає already_owned.

## F3 [orders (b) created.CourseID()!=courseID] — ЧИСТО (money-safe), але UNTESTED
Досяжне: гонка «той самий ключ, різний курс». Repo.Create при unique(user,idempotency_key)-конфлікті
повертає FindByIdempotencyKey → ІСНУЮЧЕ (перше) замовлення. Якщо перший запит був на курс X, а цей
на Y з тим самим ключем → created.CourseID()=X != courseID=Y → ErrOrderInvalid ДО будь-якого checkout
→ грошей не беруть, замовлення Y не створюється, order X цілий. Money-safe. Захисник-на-межі коректний
(не довіряй клієнту з ключем). АЛЕ: fakeStore.Create завжди повертає переданий order (не чужий), і
integration не має cross-course-replay → guard НЕ доведений тестом. Рекоменд: тест, де Create повертає
order з іншим courseID → чекати ErrOrderInvalid (+ integration: seed key K/course X, StartPurchase K/Y
→ ErrOrderInvalid). Non-blocking (guard коректний), але за дисципліною репо недоведений guard = діра.

## ПІДСУМОК: 0 blocking, 2 nb
authoring — чисто, добре покрито. orders (a) — правильна idempotency за Pay12 (+O05-frontend нота про
новий ключ на retry). orders (b) — досяжне, money-safe, але UNTESTED — раджу тест. Дрібне: paid-replay
у проді = already_owned (Pay12 літерал-нюанс + тест-реалізм paid-без-grant).
Backend-only диф; parked frontend (learn-lesson-panel-tabs.patch) поза скоупом. B39 playback flake —
модуль не чіпано, isolation+re-run зелені; шум таймінгу, не блокер, але варто вести.
