# Інвентар готовності до оплати

Аудит: Codex read-only на `fc5adf5`, 2026-09-24, за брифом сесії Courses 1 (рішення користувача — `work/PRODUCT-BRIEF.md` §24 п.21). Ключові координати (три копії предиката, `shared_rules_test.go`, ADR 010) звірила сесія Courses 1. Це вхід для специфікації фази готовності, не рішення: класи BUILD-NOW / SEAM-DOC пропонує аудит, остаточно їх визначить спека.

**Застереження звірки:** три копії предиката — не недбалість. AGENTS.md забороняє домену модуля імпортувати домен іншого модуля, тож «одне правило доступу» — архітектурне питання: як тримати одне правило без порушення меж модулів. Це головне відкрите питання фази.


Аудит: код із `HEAD fc5adf5`; `work/PRODUCT-BRIEF.md` прочитано з робочого дерева. Нічого не редагував. У дереві вже були чужі зміни: `work/PRODUCT-BRIEF.md` і два screenshots.

## Механізми

| Механізм | Що є зараз | Чого бракує для оплати | Клас | Чому | Груба ціна |
|---|---|---|---|---|---|
| Єдине правило доступу | Предикат продубльований **3 рази**: `backend/internal/course/domain/learning.go:80-82`, `backend/internal/playback/domain/authorization.go:33-35`, `backend/internal/transcript/domain/access.go:37-41`. Узгодженість тримає лише truth-table `backend/platform/architecture/shared_rules_test.go:25-92`. | Один production seam, до якого додається entitlement/payment-правило без четвертої копії. | **BUILD-NOW** | Правило використовується вже зараз у learning, transcript і playback; саме ADR 010 фіксує заміну тіла предиката. | M, 1–2 дні |
| Точки перевірки права | Learning: `course/application/learning.go:26-32`; transcript: `transcript/application/read.go:52-62`; playback session: `playback/application/sessions.go:69-77`; nginx `auth_request`: `playback/adapters/http/handler.go:139-157`. | Усі точки мають викликати одну access policy, включно з майбутнім entitlement. | **BUILD-NOW** | Це діючі product paths; пропуск одного шляху дасть доступ до transcript або media після оплати. | S–M, 0.5–1.5 дня |
| Каталог ≠ навчальний доступ | Каталог вибирає лише `status='published'`: `course/adapters/postgres/generated/courses.sql.go:34-39,138-157`; learning окремо перевіряє `CanBeViewedBy`. | Явно зберегти можливість бачити платний курс у каталозі без права навчатися. | **BUILD-NOW** | Розділення вже існує й потрібне продукту зараз; не можна зробити каталог proxy для entitlement. | S, 0.5 дня |
| Платний/безкоштовний курс і ціна | `courses` не має `price`, `currency`, `is_free` чи аналога; у коді не знайдено money-моделі. `work/PRODUCT-BRIEF.md:436-438` описує лише майбутній Order/PaymentGateway. | Рішення: free/paid, валюта, immutable price snapshot в order. | **SEAM-DOC** | Поточний продукт не споживає ціну; додавати поле без checkout — unused interface/schema. | S, 0.5 дня документації |
| Enrollment / entitlement | Таблиць `enrollments`/`entitlements` немає. ADR 010 прямо каже: “no enrollment table, no purchase, no entitlement yet”; доступ зараз обчислюється зі session + visibility. | Визначити майбутню сутність: user × course × grant/status/source/timestamps; access policy має читати її. | **SEAM-DOC** | Сутність потрібна оплаті, але зараз немає consumer; створення таблиці порушить обмеження AGENTS.md. | M, 1–2 дні документації/decision |
| Вільний доступ зараз | Будь-який authenticated user може дивитися published course: `course/domain/learning.go:80-82`; це також ADR 010. | Майбутній policy branch: free course allow, paid course entitlement required. | **BUILD-NOW** | Поточний продукт уже має цю поведінку; її треба закріпити тестами як backward-compatible базову гілку. | S, 0.5 дня |
| Відкликання доступу / refund | Реального entitlement немає. ADR 012 розрізняє `403` revoked access від session expiry; §24.9 вимагає, щоб archive **не** відкликав придбаний доступ (`work/PRODUCT-BRIEF.md:939-941`). | Окремі стани grant/revocation і правило: refund/revoke відкликає доступ, archive — ні. | **SEAM-DOC** | Refund не існує як поточний product flow; код зараз не має що відкликати. | M, 1 день документації |
| Ідемпотентність checkout/webhook | У payment/order/inbox коді не знайдено. Є лише загальна вимога про майбутній idempotency key: `work/PRODUCT-BRIEF.md:566-574`. | Provider event ID, inbox uniqueness, dedup, повторна обробка, transactional state transition. | **SEAM-DOC** | Немає вхідного webhook consumer; будувати inbox зараз означає будувати платіжну інфраструктуру без споживача. | M, 1–2 дні документації |
| Аудит змін доступу | Таблиці/події access audit не знайдено. Є timestamp/versioning для course і transcript, але не історія прав. | Хто/коли/чому grant або revoke, provider event/order reference. | **SEAM-DOC** | Потрібно для refund/support, але жодна поточна дія не змінює entitlement. | M, 1 день |
| Money як тип | `money`, `amount`, `currency` у production backend/frontend не знайдено. | Minor units або decimal policy, currency, rounding, overflow/negative guards. | **SEAM-DOC** | До появи ціни й order немає реального споживача. | S–M, 0.5–1 день |
| Course lifecycle | DB enum/invariants: `backend/migrations/00001_courses.sql:2-18`; publish/archive transitions у `course/adapters/postgres/authoring.go:58-66`; catalog виключає archive. | Зафіксувати: archive ховає з каталогу, але не скасовує entitlement; продаж дозволений лише визначеним станам. | **BUILD-NOW** для lifecycle guard; **SEAM-DOC** для sale policy | Status уже використовується; sale-specific semantics ще не мають consumer. | S–M, 0.5–1 день |
| Власник курсу | `courses.instructor_id` — `NOT NULL`, FK `ON DELETE RESTRICT`: `backend/migrations/00003_authoring.sql:5-21`; доступ owner/admin уже працює. | Для checkout достатньо стабільного `course_id`; payout policy окрема. | **BUILD-NOW** як invariant; **NOT-NEEDED** для payouts | Ownership продукт використовує зараз. Виплати інструктору не входять у поточний scope; робочого місця інструктора ще немає (§24.3). | S для перевірок; payouts — 0 зараз |
| Сесії й playback ownership | Session — opaque DB session (`backend/migrations/00002_identity.sql:27-37`); playback slot — Redis, один на account, ADR 011; access перевіряється в heartbeat і edge auth. | Не потрібен новий payment-specific session mechanism; після revoke heartbeat/auth_request мають відмовляти. | **NOT-NEEDED** | ADR 011 регулює anti-sharing, не entitlement. Payment policy має бути додана до access predicate, не до ownership lease. | 0 |
| Transcript access | `transcript/application/read.go:57-64` перевіряє access навіть при відсутньому transcript; це не друга точка входу. Frontend query: `frontend/src/routes/learn.$.tsx:82-120`. | Переконатися, що paid-access denial однаковий для transcript і player/media. | **BUILD-NOW** | Transcript уже є окремим protected read path. | S, 0.5 дня |
| Learning route / frontend gating | `/learn/$` завантажує course query у `beforeLoad`: `frontend/src/routes/learn.$.tsx:14-40`; play і heartbeat — `:63-80`; transcript — `:82-84`. | Не покладатися на frontend як на authorization; коректно обробити `course_forbidden`/revoked access. | **BUILD-NOW** | Навчальний flow використовується вже зараз; frontend має бути лише UX-шаром над backend decision. | S–M, 0.5–1 день |

### Кількість копій предиката

Зараз production-копій: **3**.

1. `course/domain/learning.go:80`
2. `playback/domain/authorization.go:33`
3. `transcript/domain/access.go:39`

`shared_rules_test.go` — не четверта копія, а тестовий guard. Проблема: тест ловить розбіжність лише після зміни; він не є єдиним runtime source of truth.

## Guard-и, які варто закласти зараз

1. **Одна access policy для всіх read/media paths.** Замінити три ручні перевірки одним узгодженим application-level seam або іншим механізмом, сумісним із модульними межами. Підтримати truth table з `shared_rules_test.go:42-62`.

2. **Не змішувати visibility і entitlement.**
   - catalog: `published`;
   - learning/transcript/playback: access decision;
   - archive не дорівнює revoke згідно з `work/PRODUCT-BRIEF.md:939-941`.

3. **Зберегти всі чотири поточні точки enforcement:**
   - course learning `course/application/learning.go:26-32`;
   - transcript `transcript/application/read.go:57-62`;
   - playback session `playback/application/sessions.go:69-77`;
   - nginx `auth_request` через `playback/adapters/http/handler.go:139-157`.

4. **Не дозволяти authenticated-only predicate пропустити анонімного користувача.** Зараз transport authentication передує predicate; schema гарантує `instructor_id NOT NULL`. Це слід закріпити тестом на кожному seam.

5. **Paid-course access має бути course-scoped, не role-scoped.** §24.10 прямо каже, що instructor/admin/student мають адитивні здатності; роль не повинна сама по собі давати entitlement.

6. **Archive guard.** Заборонити lifecycle-операції трактувати `archived` як revoke. Course status constraints: `backend/migrations/00001_courses.sql:8-15`; lifecycle update: `backend/internal/course/adapters/postgres/authoring.go:58-66`.

7. **Media і transcript повинні давати однакову denial semantics.** ADR 012: edge `403` не розкриває причину; heartbeat є джерелом розрізнення expiry/takeover/revocation.

8. **Зберегти account-scoped playback ownership.** Не прив’язувати майбутній entitlement до `playbackId`, пристрою чи IP; ADR 011.

9. **Future seam для entitlement має бути explicit.** У документації зафіксувати майбутній input policy: `actor + course + access state`, а не створювати порожній `EnrollmentRepository`.

10. **Ціна не повинна визначати доступ напряму.** Майбутній order має фіксувати ціну на момент покупки; access перевіряє grant/entitlement, а не поточне поле ціни курсу.

## Свідомо не будувати зараз

- `order`/`checkout` — остання фаза за §24.21.
- Payment gateway port або provider adapter — немає реального external dependency; AGENTS.md прямо забороняє unused ports.
- Webhook endpoint і inbox table — документувати idempotency/event-dedup seam, але не додавати код.
- `enrollments`/`entitlements` tables — немає поточного consumer; лише зафіксувати майбутню модель і правила.
- `price`, `currency`, `Money` — немає поточного paid/free UX або order.
- Refund/revocation workflow — немає purchase/grant, який можна відкликати.
- Access audit log — немає access mutation flow.
- Instructor payouts/revenue split — не в scope; власник курсу вже існує для authoring, але payout economics не визначена.
- Новий session/lease механізм — ADR 011 уже покриває playback ownership; оплата не повинна його переписувати.
- Окремий instructor workspace — §24.3 визначає його як окрему майбутню фазу, не readiness-фазу.

**Головний висновок:** readiness-фаза повинна реально закріпити єдиний поточний access seam і його enforcement points. Усе, що починається з order, money, webhook або entitlement storage, зараз є лише документованим швом.

## Вхід для спеки від рецензента (Courses 2, 2026-09-24)

Варіант «спільне ядро окремим пакетом» можливий **без** імпорту між доменами. Для цього ядро
визначає нейтральний вхід — щось на кшталт `{isAdmin, isOwner, isPublished, hasEntitlement?}`, — а
кожен модуль мапить у нього свої типи, і залежність дивиться всередину, як з Orval-моделями. Три
поточні копії різняться саме типами актора, а не логікою (`admin || owner || published`), тож
нейтральний вхід їх зводить. Спека має зважити це проти варіанта «політика в одному модулі за
портом»: там ризик, що модуль-власник політики стає хабом.
