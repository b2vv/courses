# Codex brief — L2a: purchase-side membership edges (orders)

Репо `courses`, гілка `feat/enrollment`. План `docs/enrollment/plan.md` §4 (purchase-на-наявному членстві).
Роль: реалізуєш L2a; коміт/генерація/гейти — Courses 1. L1 уже в дереві (коміт 35cf77b): домен має
`GrantSelf`/`GrantBlocked` + переходи `Block`/`Unblock`/`Revoke`/`SelfEnroll`, міграція 00015 дала повну
`UNIQUE(user_id,course_id)` (constraint `course_grants_one_per_user_course`), `course_grants` тепер
«authoritative membership, стан=влада доступу».

РОБОЧИЙ ФАЙЛ: /Users/strelia/projects/courses/work/_wip-L2a-codex.md
Дописуй ПІСЛЯ КОЖНОГО КРОКУ (append, не rewrite). Перший запис — рядок `ДАЛІ:` з тим, що перевіряєш.
Далі ФАКТ / РІШЕННЯ / ДАЛІ / ПИТАННЯ. Фінал = стислий переказ файлу.

## Межа L2a (НЕ виходь за неї)
ЛИШЕ purchase-side: як наявне членство впливає на старт і сетл оплати. Enroll/Block/Unblock/Revoke
як операції користувача/адміна, MembershipState і Roster — це L2b/L2c, НЕ чіпай тут.

## Контекст коду (звірено)
- `domain/grant.go`: `CanStartOrder(userID, courseID, grants []Grant) error` — зараз відхиляє лише коли
  є active-рядок (`ErrCourseAlreadyOwned`). `Grant` має `Source()`/`State()`; `GrantSelf`/`GrantActive`/
  `GrantBlocked`/`GrantRevoked` існують.
- `application/service.go:55` StartPurchase завантажує `service.grants.ActiveGrants(...)` (лише active) →
  `CanStartOrder`. Це ТРЕБА змінити: воно має бачити й blocked/revoked, і розрізняти self-active.
- `ports/repository.go` `GrantRepository`: `ActiveGrants`, `GrantWithAudit`, `GrandfatherOnPriceSet`.
- `adapters/postgres/repository.go`: `ActiveGrants` (:100, лише state='active'); settle-INSERT у
  `PaymentEventRepository.Apply` (~:252) вже `INSERT ... 'purchase' ... 'active' ON CONFLICT (user_id,course_id)
  DO NOTHING RETURNING id`; на ErrNoRows — гілка `UPDATE orders SET refund_required=true` (сітка рефанду).
- `grant_audit(action text)` — без CHECK, тож нові action-и («upgrade») валідні.
- Інваріант `NewGrant`: `(source='purchase') == (order_id IS NOT NULL)`. Апгрейд self→purchase ПОВИНЕН
  виставити order_id (інакше інваріант порушиться на читанні).

## Рішення плану §4 — ПРАВИЛА (реалізувати точно)
Наявне членство (один рядок на (user,course)) впливає так:
- **active + source=self** → purchase ДОЗВОЛЕНО: сетл апгрейдить ТОЙ САМИЙ рядок self→purchase (без 2-го).
- **active + source purchase/admin/grandfather** → `ErrCourseAlreadyOwned` (вже володіє).
- **blocked** → purchase відхилити (`ErrCourseAlreadyOwned`): оплата НЕ знімає мют.
- **revoked** → purchase відхилити (`ErrCourseAlreadyOwned`): оплата НЕ воскрешає (адмін-revoke навмисний;
  поведінкова зміна — re-purchase після revoke більше не відновлює доступ).
- **немає рядка** → дозволено (як зараз).

## Реалізація (TDD; тести вже-написаного — першими)

### 1. Домен `domain/grant.go`
- Перероби `CanStartOrder`: сигнатура на одне членство замість зрізу —
  `CanStartOrder(membership *Grant) error` (nil = немає членства → ok). Правило: ok, якщо membership==nil
  АБО (state==active && source==self); інакше `ErrCourseAlreadyOwned`.
- Table-driven unit: усі (state × source) + nil → очікуваний результат за правилами вище. Особливо:
  self-active→ok, purchase-active→owned, blocked→owned, revoked→owned, nil→ok.

### 2. Порт + репозиторій — читання членства
- `ports.GrantRepository`: додай `Membership(ctx, userID, courseID string) (domain.Grant, bool, error)`
  — єдиний рядок (будь-який стан) або found=false. (ActiveGrants лишай для інших викликів — НЕ видаляй.)
- `adapters/postgres/repository.go`: реалізуй `Membership` (SELECT того ж набору колонок, що й ActiveGrants,
  але без фільтра по state; LIMIT 1 — рядок унікальний по (user,course)). Мапінг у `domain.NewGrant`.

### 3. Wiring StartPurchase (`application/service.go`)
- Заміни `ActiveGrants` → `Membership`; побудуй `*domain.Grant` (nil, якщо !found) іклич новий `CanStartOrder`.
- Онови fakeStore у `application/service_test.go` (:341) та `adapters/http/handler_test.go` (:85): додай
  `Membership` у фейк; існуючі кейси (active grant blocks purchase; revoked-since → allowed) переклади на нову
  модель — УВАГА: старий коментар на service_test.go:61 казав «стоїть за revoked-since», а тепер revoked
  БЛОКУЄ; онови очікування відповідно до §4 (revoked → ErrCourseAlreadyOwned).

### 4. Settle UPSERT — апгрейд self-active (`adapters/postgres/repository.go`, Apply ~:252)
- Заміни settle-INSERT на UPSERT з умовним апгрейдом:
  ```
  INSERT INTO course_grants(id,user_id,course_id,source,order_id,state,granted_at)
  VALUES($1::text::uuid,$2::text::uuid,$3::text::uuid,'purchase',$4::text::uuid,'active',$5)
  ON CONFLICT (user_id, course_id) DO UPDATE
    SET source='purchase', order_id=EXCLUDED.order_id
    WHERE course_grants.source='self' AND course_grants.state='active'
  RETURNING id::text
  ```
  — вставка (нема рядка) АБО апгрейд (був self-active) повертають id → пиши grant_audit; будь-який інший
  наявний рядок (purchase/admin/grandfather-active, blocked, revoked) не оновлюється → RETURNING порожній →
  `pgx.ErrNoRows` → наявна гілка `refund_required` (НЕ чіпай її логіку). Для апгрейду audit action='upgrade'
  (не 'grant'), payment_event_id=той самий. Для вставки лишай 'grant'. Розрізни за тим, чи id новий — або
  поверни ще `xmax=0`-прапор; простіше: SELECT source перед, або порівняй повернутий id з $1 (новий=вставка).
  Обери надійний спосіб і опиши в робочому файлі.

## Інтеграційні тести (`adapters/postgres/repository_integration_test.go`)
- **En02 upgrade:** посій self-active (user,course); застосуй paid event → рядок став source='purchase',
  order_id встановлено, state='active', id ТОЙ САМИЙ (без 2-го рядка); grant_audit має 'upgrade'.
- **blocked/revoked/purchase-active + paid event:** рядок НЕ змінюється; order.refund_required=true.
- **En15 single-row:** після апгрейду `SELECT count(*) FROM course_grants WHERE user_id,course_id` = 1.
- Наявні Pay-тести (Pay26 grandfather idempotent, settle-with-active-admin-grant) мають лишитись зеленими.

## Правила
- Домен чистий (stdlib). Якщо міняєш `ports` — онови ВСІ реалізації/фейки (шукай компілятором).
- Якщо треба sqlc — НЕ запускай генерацію (ці queries сирі в repository.go, як наявні). generated не чіпай.
- НЕ комість. gofmt чистий; `go build ./...`, `go vet ./internal/orders/...`,
  `GOCACHE=/private/tmp/courses-l2a-gocache go test ./internal/orders/domain/...` зелені.
- Інтеграцію (Docker) проганяє Courses 1 — ти лише пишеш тести й кажеш, що саме перевірити.

## Критерій готовності
- Домен-юніти зелені (нова CanStartOrder-таблиця). build/vet/gofmt чисті.
- У робочому файлі: як розрізнив вставку від апгрейду в settle; які фейки оновив; чи наявні Pay-кейси лишились.
