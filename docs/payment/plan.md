# Оплата — план (ЯК)

| | |
|---|---|
| Фаза | Оплата (`PRODUCT-BRIEF.md` §24 п.21) — остання |
| Спека | `docs/payment/spec.md` (GATE 1 пройдено 2026-09-29) |
| Статус | **GATE 2 ПРОЙДЕНО (зелений)** — Courses 2 на claude-opus-4-8 підтвердив закриття 2026-09-29; до тасків |
| Автор | Courses 1 |
| Гілка | `feat/payment` |

Reuse-факти звірені з кодом (контракт `access`, чотири мапінги, admin-патерни, каталог, міграції);
file:line нижче — з тієї звірки.

## Архітектурні рішення (2026-09-29, Courses 1; кожне закриває альтернативу)

**A1. Контракт доступу знає платність, право і прев'ю — ADR 014.** `access.Course` дістає факти
`Paid`, `Entitled`, `Preview`; `CanView = Administrator ∨ owner ∨ Entitled ∨ (Published ∧ (¬Paid ∨ Preview))`.
Альтернатива «лише `|| entitled`» (шви §5) закрита: з нею опублікований платний курс лишався б відкритим
через `Published`. Контракт лишається однією чистою функцією на stdlib (ADR 013); ADR 014 доповнює 013
(ADR 013:53 вимагав перегляду, коли доступ вирішуватиме четвертий модуль — умова настала з 7b).
**Перевірка:** `access_mapping_ast_test.go` — досі рівно 4 методи, кожен `return access.CanView(…)`;
`shared_rules_test.go` проганяє розширену таблицю через усі 4 доменні типи.

**A2. Новий модуль `orders` володіє правом, замовленням, грошима, inbox і аудитом.** Інші модулі **читають**
таблицю `course_grants` у своїх SQL-адаптерах (прецедент: `discussion` читає `courses` напряму), але не
імпортують пакети `orders`. Альтернатива «право у `course`» закрита: оплата — окремий життєвий цикл із
власними інваріантами (замовлення, події провайдера).
**Перевірка:** architecture-тест — жоден модуль не імпортує `internal/orders/**`, `orders` не імпортує
модулі (дзеркало `discussion_ownership_test.go`).

**A3. Ціна живе на курсі, прев'ю — на уроці (модуль `course`).** `courses.price_minor` (NULL = безкоштовний),
`course_lessons.is_preview`. `Paid` для інших модулів = `price_minor IS NOT NULL` — вони вже джойнять
`courses`. Замовлення бере **знімок** ціни (Pay02/Pay17). Альтернатива «ціни в `orders`» закрита: ціну
задає автор курсу, а доступові потрібен лише факт «платний».

**A4. Порту провайдера до першого адаптера немає (AGENTS.md).** Провайдер-незалежне ядро — application-
операція `ApplyPaymentEvent(event)`: ідемпотентна за `(provider, event_id)`, в **одній транзакції**
записує подію в inbox, переводить замовлення в `paid`, видає право й пише аудит. Порт провайдера,
checkout-сесія і вебхук-ендпоінт з перевіркою підпису з'являються разом із першим адаптером (зріз O06),
який **заблокований на назві провайдера**.

**A5. Право видають лише дві дороги:** застосована подія оплати (`source=purchase`, з `order_id`) або
адмін (`source=admin`, `order_id` NULL). Активне право одне на `(user, course)` — частковий UNIQUE.
Редірект-повернення права не видає ніколи.

**A7. free→paid grandfathering (B2, рішення користувача 2026-09-29).** Коли автор уперше ставить курсу
ціну, orders видає активний grant (`source=grandfather`, `order_id` NULL) усім, хто вже має **прогрес**
на курсі (playback watched/progress) — щоб ніхто не втратив доступ посеред навчання. Видача — одна
транзакція з `UPDATE courses SET price_minor`; ідемпотентна (частковий UNIQUE не дублює наявні права).
`source` розширюється до `purchase|admin|grandfather`; аудит фіксує підставу.

**A6. Причина відмови — один спільний предикат у контракті, вибір коду в application (B1+NB1).**
Контракт `CanView` повертає bool; поряд — чиста `access.PurchaseRequired(subject, course) bool` =
`Published ∧ Paid ∧ ¬Preview ∧ ¬(Administrator ∨ owner ∨ Entitled)`. **Усі чотири** application-шари на
відмові view кличуть цей один предикат і обирають `course_purchase_required` (403) інакше `course_forbidden`.
Один предикат — без дубля з §A1-правилом (NB1, знято ризик десинку коду); Pay19 виконується на 4 поверхнях
однаково. Сьогодні transcript/discussion віддають generic `ErrForbidden` (`discussion/application/application.go:96,121,218`,
`transcript/application/read.go`) — зріз O04 підмінює на вибір за предикатом.

## Reuse-first

| Що | Де | Як |
|---|---|---|
| Контракт доступу | `contracts/access/access.go:6-21` | розширюємо факти + правило (A1) |
| Таблиця істинності | `contracts/access/accesstest/table.go:14-32` (9 кейсів) | додаємо виміри paid/entitled/preview |
| 3 architecture-тести | `access_mapping_ast_test.go`, `shared_rules_test.go`, `access_fixture_isolation_test.go` | AST без змін; shared_rules — нові факти в 4 типах |
| 4 SQL-guard-и D4 | integration-тести course/playback/transcript/discussion | нові рядки таблиці з правом (сід grant) |
| Admin-only гейт | `identity/adapters/http/admin_handler.go:136-147` + `RegisterAdmin` (`cmd/api/main.go:112-115`) | адмін-видача права |
| Лог-таблиця | `migrations/00010_instructor_moderation.sql:21-30` | форма `grant_audit` |
| Ідемпотентність | UNIQUE-патерн `discussion_comment_reports` (00010) | inbox UNIQUE(provider, event_id) |
| `IdempotencyKey` | `openapi.yaml:1152-1156` (зарезервовано «для оплати») | старт покупки |
| Каталог | `course/adapters/http/handler.go:28-58`, `sql/queries/courses.sql:1,19`, `openapi.yaml:1503-1564` | + `price`, `paid` (лише показ) |
| Фронт | `presentation/storefront/CourseView.tsx:41-45`; `routes/admin.users.index.tsx:16-41` | кнопка «Купити»; адмін-видача |

**Нове:** модуль `orders`; факти `Paid/Entitled/Preview`; userID у 4 портах і запитах; міграція 00011;
ендпоінти старту покупки, статусу, адмін-видачі; `ApplyPaymentEvent`; поля ціни/прев'ю в авторингу;
стан «потрібна покупка» і «оплата в обробці» у фронті; ADR 014.

## Де ляже join на право — 6 запитів у 5 файлах

| # | Запит | userID | Прев'ю-урок |
|---|---|---|---|
| 1 | `course/adapters/postgres/learning.go:11` `getLearningCourse` | є (`:50`) | вибраний урок |
| 2 | `playback/adapters/postgres/access.go:11` `loadVideoCourse` — **nginx `auth_request`, гарячий шлях** | **додати** (порт `VideoAccessReader`) | урок відео |
| 3 | `playback/adapters/postgres/playback.go:12` `loadLessonPlayback` | є | урок |
| 4 | `transcript/adapters/postgres/repository.go:31` `getLessonCourse` | **додати** | урок |
| 5 | `discussion/adapters/postgres/repository.go:20` `lessonFactsQuery` | **додати** | урок |
| 6 | `discussion/adapters/postgres/repository.go:62` `commentQuery` | **додати** | урок коментаря |

## Зрізи

| ID | Що | Закриває |
|---|---|---|
| **O01** | ADR 014; `access.Course` += `Paid/Entitled/Preview`; `CanView` + `PurchaseRequired` (спільні факти); `accesstest` розширено; `access_test`, `shared_rules_test`; **property-тест `PurchaseRequired ⟹ ¬CanView` над таблицею (NB1' анти-дрейф)**; 4 доменні типи фактів. **Поведінку не міняє**, доки SQL не дає фактів | Pay01 (unit), Pay05, Pay15, Pay27 (unit) |
| **O02** | Міграція 00011 + модуль `orders`: `Money` (мінорні, UAH), `Order` (pending→paid\|failed), `Grant` (source, state), правило застосування події; unit | Pay02, Pay03, Pay18 (+ доменна частина Pay04/Pay16/Pay21) |
| **O03** | 6 запитів джойнять `course_grants` і читають `price_minor`/`is_preview`; userID через 4 порти; integration-таблиця через SQL у 4 модулях; **замір гарячого шляху** | Pay01 (SQL), Pay11 (integration), Pay14, Pay23 |
| **O04** | `orders` application+HTTP: старт покупки (IdempotencyKey, pending), статус, адмін-видача (admin-only, аудит), `ApplyPaymentEvent` (unknown-order → не термінальна відмова, Pay21b), **grandfather при першому price-set** (A7); спільний `access.PurchaseRequired` і вибір коду на **4** application-шарах (A6); каталог `price/paid`; авторинг ціни й прев'ю; OpenAPI+Orval; architecture-тест `orders` | Pay04, Pay06, Pay07, Pay08, Pay10, Pay12, Pay13, Pay16, Pay17, Pay19, Pay20, Pay21, Pay21b, Pay22, Pay26 |
| **O05** | Фронт: ціна/«платний» у каталозі; «Купити» → старт; «оплата в обробці» з опитуванням статусу; «потрібна покупка» + позначки прев'ю на уроці; адмін-видача; авторинг ціни/прев'ю | Pay07, Pay09, Pay11, Pay13 (UI) |
| **O06** | **Провайдер: WayForPay** (рішення власника 2026-10-01, `provider-research.md` → Decision; sandbox — без ФОП). Порт + перший адаптер: checkout-сесія (redirect), вебхук-ендпоінт (без сесії, перевірка підпису → `ApplyPaymentEvent`), повернення → сторінка статусу; ключі лише з env | Pay24, Pay25 |
| **O07** | e2e (непровайдерні журнеї одразу, провайдерна — після O06), упорядкований `make check`, звіт **Pay01–Pay27 + Pay04b/Pay21b по одному** (NB5: рахувати РЯДКИ, не діапазон — урок Phase 8), PR | усі |

Фаза закривається лише після O06: доказ повної покупки мусить пройти production-шлях (sandbox
провайдера), а не фейк (AGENTS.md: «фейк доводить фейк»).

## Side-effects

- **Семантика доступу змінюється в 4 модулях** → повний e2e-набір, не таргетний (урок 7b/інструктора).
- **Гарячий шлях:** `loadVideoCourse` на кожен сегмент отримує join на `course_grants`. **Механізм:**
  частковий UNIQUE-індекс `(user_id, course_id) WHERE state='active'` — один index probe. **Магнітуда
  вимірюється в O03:** `EXPLAIN (ANALYZE, BUFFERS)` запиту до й після на стенді з сідом, числа — у звіті;
  якщо план не index-only/index scan або час запиту зросте понад ×2 — стоп і перегляд.
- Сіди й e2e: усі курси зараз без ціни → безкоштовні → поведінка не міняється (Pay15).
- Новий неавтентифікований вебхук-ендпоінт (O06) — нова поверхня атаки: лише підпис провайдера, ліміт
  nginx `zone=api`, дані картки до нас не потрапляють (redirect-checkout).
- Персональні дані: inbox зберігає id події, тип, суму, посилання на замовлення — не повний payload.
- **Названа межа (NB4):** grandfather (A7) ключиться на **playback-прогрес** (`lesson_progress`, 00006).
  Механізм: хто вчився на free-курсі лише через транскрипт/обговорення **без відео-прогресу**, гранта не
  дістане й втратить доступ при free→paid. Магнітуда: підмножина «активні без жодного watched-запису» —
  вузька. Прийнято; за потреби ширшати до «прогрес ∨ коментар» — окреме рішення, не MVP.

## План міграції — `00011_payment.sql` (goose, `SET LOCAL lock_timeout = '5s'`)

- `ALTER TABLE courses ADD COLUMN price_minor bigint NULL CHECK (price_minor IS NULL OR price_minor > 0),
  ADD COLUMN currency text NULL CHECK (currency IS NULL OR currency = 'UAH'),
  CHECK ((price_minor IS NULL) = (currency IS NULL))`.
- `ALTER TABLE course_lessons ADD COLUMN is_preview boolean NOT NULL DEFAULT false`.
- `orders (id, user_id FK users, course_id FK courses, amount_minor bigint CHECK >0, currency CHECK='UAH',
  status text CHECK IN ('pending','paid','failed'), idempotency_key text, created_at, paid_at NULL,
  provider text NULL, provider_ref text NULL, UNIQUE(user_id, idempotency_key))`.
- `course_grants (id, user_id FK, course_id FK, source text CHECK IN ('purchase','admin','grandfather'), order_id NULL FK orders,
  state text CHECK IN ('active','revoked'), granted_at, granted_by NULL FK users,
  CHECK ((source='purchase') = (order_id IS NOT NULL)))` + частковий `UNIQUE (user_id, course_id) WHERE state='active'`.
  Grandfather-видача (A7) при першому price-set — `INSERT … SELECT` з тих, хто має прогрес, `ON CONFLICT DO NOTHING`.
- `payment_events (id, provider text, event_id text, order_id NULL FK, type text, amount_minor NULL,
  received_at, UNIQUE(provider, event_id))`.
- `grant_audit (id, grant_id FK, actor_id NULL, action text, reason text NULL, payment_event_id NULL, created_at)`.
- Каскади: видалення курсу/користувача — **не** каскадить на `orders`/`payment_events` (фінансовий слід);
  FK без `ON DELETE CASCADE`, звір у O02 із тим, як курси реально видаляються (м'яко чи ні).

## Rollback

- Down: DROP нових таблиць, DROP колонок `price_minor`, `currency`, `is_preview`.
- **Вимикач без відкату коду:** зняти ціни з курсів → `Paid=false` скрізь → правило зводиться до
  сьогоднішнього. Право без ціни нічого не відкриває понад наявне.
- O01 поведінково нейтральний: відкат контракту безпечний, доки SQL фактів не дає.
- Незворотне після запуску: отримані гроші й видані права — фінансовий слід; down-міграцію на проді
  з реальними замовленнями не виконують (записано як обмеження, не як крок).

## Конституція (AGENTS.md)

- Домен `orders` і `access` — stdlib + `contracts`; `Money` — доменний тип.
- Модуль не імпортує модуль: інші модулі лише читають `course_grants` SQL-ом в адаптерах; architecture-тест (A2).
- Порт лише з реальною залежністю: провайдер-порт з'являється в O06 разом з адаптером (A4).
- Generated sqlc/Orval — в адаптерах, явний мапінг.
- Явні бізнес-коди: `course_purchase_required`, `course_already_owned`, `payment_event_rejected`,
  `grant_forbidden`; секрети провайдера — лише env, не в логах; URL у логах без query.
- TS strict, React Aria, Query для серверного стану, семантичні токени, ≥44px, 200% зум.
- Рішення, що закривають альтернативи (A1–A6), датовані й мають спосіб перевірки.

## Приймальні сценарії (метод `acceptance-spec`)

**Happy path**

| # | Сценарій | Передумова | Дія | Спостережуваний результат | Чим |
|---|---|---|---|---|---|
| Pay01 | Правило доступу | таблиця ролей × статусів × paid × entitled × preview | `CanView` + 4 мапінги + SQL | кожен рядок дає вердикт правила A1 | unit `access_test`, `shared_rules_test`; integration ×4 |
| Pay02 | Знімок ціни | курс 500 UAH | старт покупки | замовлення pending з `amount_minor=50000`, `UAH` | integration |
| Pay04 | Застосування події | pending-замовлення | `ApplyPaymentEvent(paid)` | замовлення `paid`, активне право, запис аудиту, подія в inbox — одна транзакція | integration |
| Pay05 | Архів не відкликає | право на курс, курс архівовано | відкрити урок | доступ є | unit table; integration |
| Pay06 | Аудит | покупка / адмін-видача | — | у `grant_audit` рядок з актором (або подією) і дією | integration; contract |
| Pay07 | Каталог показує ціну | курс з ціною | відкрити каталог / сторінку | «платний» + ціна; уроків без права не відкрито | contract; e2e |
| Pay10 | Адмін-видача | адмін, користувач без права | видати право | активне право `source=admin`, аудит; урок відкривається | contract; e2e |
| Pay11 | Прев'ю | платний опублікований курс, урок-прев'ю | студент без права відкриває прев'ю / звичайний урок | прев'ю: плеєр+транскрипт+обговорення; звичайний: стан «потрібна покупка» з кнопкою | integration; e2e |
| Pay12 | Старт покупки ідемпотентний | — | двічі з тим самим `Idempotency-Key` | одне замовлення, та сама відповідь | contract; integration |
| Pay13 | «Оплата в обробці» | pending-замовлення | повернення на сторінку статусу | текст «обробляємо»; після події — доступ без перезавантаження | contract; e2e |
| Pay14 | Guard D4 через SQL | сід права | 6 запитів 4 модулів | вердикти збігаються з таблицею | integration ×4 (обидва читачі playback) |

**Межі**

| # | Сценарій | Передумова | Дія | Результат | Чим |
|---|---|---|---|---|---|
| Pay03 | Гроші | — | створити Money | мінорні одиниці, UAH | unit |
| Pay15 | Безкоштовний курс не змінився | курс без ціни | будь-хто відкриває | як сьогодні (published → відкрито) | unit table; e2e регресія |
| Pay16 | Повторна покупка | активне право **будь-якого source** (purchase/admin/grandfather) | старт покупки | 409 `course_already_owned`, замовлення не створено (NB2: перевіряти `course_grants` за станом active, не за source) | contract; integration |
| Pay17 | Зміна ціни після покупки | куплено за 500, ціну змінено на 700 | — | доступ є; замовлення лишається 50000 | integration |
| Pay18 | Некоректні гроші | — | від'ємна / переповнення / не UAH | відмова | unit |
| Pay23 | Гарячий шлях | сід із правами | `EXPLAIN (ANALYZE, BUFFERS)` `loadVideoCourse` до/після | index scan; час ≤ ×2 від базового; числа у звіті | integration (замір) |

**Відмови**

| # | Сценарій | Передумова | Дія | Результат | Чим |
|---|---|---|---|---|---|
| Pay19 | Без права на платний | студент без права, не-прев'ю | API уроку/сесії/транскрипту/коментарів | 403 `course_purchase_required` на всіх 4 (через спільний `access.PurchaseRequired`, A6); стан не змінено | contract ×4 |
| Pay20 | Адмін-видача не адміном | інструктор/студент | видати право | 403 `grant_forbidden` | contract |
| Pay21 | Подія з **чужою сумою** на відоме замовлення | замовлення існує, сума події ≠ сума замовлення | `ApplyPaymentEvent` | `payment_event_rejected`, права немає, подія записана як термінальна відмова | integration |
| Pay21b | Подія на **ще-не-створене** замовлення (гонка «подія до коміту», NB3) | замовлення не знайдено | `ApplyPaymentEvent` | **НЕ** термінальна відмова: подія не пишеться в inbox як reject; відповідь провайдеру спонукає ретрай (або запис pending-unmatched для звірки, коли замовлення з'явиться) — інакше гроші без доступу | integration |
| Pay26 | Grandfather при free→paid (A7) | безкоштовний курс, N студентів із прогресом | автор ставить ціну | усі N дістають активний grant `source=grandfather`; доступ не зникає; повтор price-set не дублює | integration |
| Pay27 | Анти-дрейф предикатів (NB1') | таблиця істинності | `PurchaseRequired` над кожним рядком | скрізь `PurchaseRequired ⟹ ¬CanView`; обидва предикати з однієї збірки фактів (не можна забути Preview в одному) | unit (property) |
| Pay22 | Недоступна залежність | БД недоступна | старт/статус/видача | 503 `dependency_unavailable` | contract |
| Pay04b | Дубль події | подія вже застосована | повтор | no-op, друге право не створено | integration |

**Крос-зрізові**

| # | Сценарій | Результат | Чим |
|---|---|---|---|
| Pay08 | Контракт | кожна нова відповідь звірена з OpenAPI; `check-generated` зелений | contract; gate |
| Pay09 | a11y | «Купити», «в обробці», «потрібна покупка», адмін-видача — клавіатура, фокус, ≥44px, 200%, тема, axe | e2e |

**Провайдерні (O06, заблоковано на назві провайдера)**

| # | Сценарій | Результат | Чим |
|---|---|---|---|
| Pay24 | Вебхук | валідний підпис → `ApplyPaymentEvent`; невалідний → 401, нічого не записано | contract на тестових векторах провайдера |
| Pay25 | Повна покупка | sandbox провайдера: оплата → вебхук → право → урок відкрито | e2e (sandbox) |

Ручних рядків: **0**. Три групи присутні.

## GATE 2 — знахідки й закриття (свіжий рецензент Courses 2, Opus 4.8, 2026-09-29)

Вердикт виносив не автор плану: сесія Courses 2 (підтверджено `claude-opus-4-8`), critique+defense інлайн,
нейтральні формулювання. Було **червоно**; нижче — закриття.

| Знахідка | Рівень | Закрито |
|---|---|---|
| B1 — `course_purchase_required` лише в course/player, а Pay19 вимагає ×4 (transcript/discussion віддають generic forbidden) | blocking | A6: спільний `access.PurchaseRequired`, вибір коду на 4 шарах; O04 підмінює generic |
| B2 — free→paid відбирає доступ у вже-навчених студентів | blocking | **A7 grandfather** (рішення користувача): price-set видає grant усім із прогресом; Pay26 |
| NB1 — дубль предиката платної видимості (десинк коду) | non-blocking | A6: один предикат у контракті |
| NB2 — Pay16 має ловити active grant будь-якого source | non-blocking | Pay16 уточнено (за станом active, не source) |
| NB3 — подія-до-замовлення пишеться як reject → гроші без доступу | non-blocking | Pay21b: unknown-order не термінальна відмова, ретрай/pending-unmatched |

Підтверджено чистим: покриття доступу повне (6 запитів ловлять усі студентські read+write; media лише
`/admin/*`, WebVTT не окремий ендпоінт, ASR-draft admin-only, discussion publish через запит #5); A2 vs
AGENTS.md (boundaries_test забороняє імпорт пакетів, не читання таблиць); O01 поведінково нейтральний.

**Повторна звірка (Courses 2, claude-opus-4-8, 2026-09-29): ЗЕЛЕНО на осі blocking.** 3 залишкові
non-blocking адресовані: NB1' — property-тест анти-дрейфу `PurchaseRequired ⟹ ¬CanView` (O01, Pay27);
NB5 — O07 рахує рядки Pay01–Pay27+Pay04b/21b, не діапазон; NB4 — grandfather на playback-прогресі, названа
межа (Side-effects). Нових blocking правки не внесли.

## План-секція майбутнього звіту

Звіт — дельта до цього плану: що справдилось, які Pay-рядки якими тестами, магнітуда гарячого шляху
(Pay23), перенесені межі.
