claude-opus-4-8

# GATE 2 re-review — enrollment plan (2-й прохід)

ДАЛІ: прочитати оновлений plan.md (зроблено), spec.md (зроблено), критику+захист (зроблено).
Пройти КОЖНУ знахідку B1-B5/H1-H3/N1-N3 + слабкі, звірити з кодом. Шукати нові діри від DROP partial-unique + UPSERT-переписування orders.

## ФАКТИ (звірка з кодом)

- access.go: const loadVideoCourse на 11-21, JOIN grants state='active' на :19-20 — media-bytes auth_request. Підтверджено.
- contracts/access/access.go: CanView :22-24 пускає free published без Entitled; accessExempt :37 читає course.Entitled (Entitled не прибрати). CanStream :27-28.
- grant.go: source{purchase,admin,grandfather} :7-11, state{active,revoked} :15-18; NewGrant інваріант (source==purchase)==hasOrder :43; CanStartOrder :68 рефузить лише active.
- repository.go: ON CONFLICT (user_id,course_id) WHERE state='active' на :182 (grandfather) і :252 (Apply purchase); isUnique "course_grants_one_active_per_user_course" на :135 (GrantWithAudit). ← ВСІ три спираються на партковий індекс; DROP його → ON CONFLICT-інференс зламається без рерайту. План §3 :47-48 називає саме ~135/182/252. Збіг рядків підтверджено.
- orders/application/service.go: ActiveGrants :55 (фільтр state='active') → CanStartOrder :59; GrantWithAudit :153.
- 00011_payment.sql:40-42 partial unique; CHECK source(3)/state(2) :32,:34.
- Найвища міграція 00014_accounts → 00015 коректний. Підтверджено.
- discussion application.go Publish :146 email-гейт ПЕРШИЙ, потім CanBeViewedBy :153 — збігається з планом «спершу verified, потім членство». Комунікація зараз через CanBeViewedBy(Entitled) → free-не-член публікує (дірка існує, план закриває через CanCommunicate).
- Entitled-JOIN у: learning.go:46, playback.go:29, access.go:20, transcript/repository.go:39, discussion/repository.go:27/74, discussion/review.go:19. План §0+§8T3 називає всі.

## НОВА ДІРА (від DROP+UPSERT)

1. access_integration_test.go:219 ЖОРСТКО перевіряє наявність рядка "course_grants_one_active_per_user_course" у EXPLAIN-плані media-bytes. DROP індексу → тест падає. План НЕ називає цей тест серед змін. Гейт ловить (гучно, не тихо), але за рішенням AGENTS «хто додає глобально-лічене — прибирає за собою» план мав би його назвати. Також інтеграційні тести вставляють course_grants і потребуватимуть нових CHECK-станів. NON-BLOCKING.
2. Apply/purchase-UPSERT на наявному blocked/revoked(paid) рядку: ActiveGrants фільтрує state='active', тож blocked/revoked-paid юзер ПРОХОДИТЬ CanStartOrder і може стартувати order; grant-UPSERT мусить НЕ знімати mute / НЕ воскрешати revoke (§3 «paid revoke знімає лише адмін», §4 немає blocked→active-крім-unblock і revoked→active-крім-admin/free). Резолвиться у refund_required через наявну сітку (repository.go:262). Структурно безпечно (один рядок, без спурйозного active → H1 закрито структурно), АЛЕ: (а) план не фіксує, що paid re-purchase-after-revoke більше НЕ відновлює доступ (поведінкова зміна проти поточного коду); (б) новий предикат CanStartOrder (рефузити й на blocked?) не названий. Має бути в §4 unit-таблиці переходів. NON-BLOCKING (інваріант не порушено, сітка refund_required є).
3. discussion READ лишається на CanBeViewedBy(Entitled=active-only); blocked на PAID курсі втрачає читання дискусії, зберігаючи відео/транскрипт через CanConsume — можлива контент/комунікація-неузгодженість для blocked. Не було в попередніх знахідках; дрібне. NON-BLOCKING.

## duplicate-order / re-purchase НЕ зламано
orders UNIQUE(user_id, idempotency_key) (00011:25) незалежний від grant-індексу; Create() :57 ловить дубль ідемпотентності. Re-purchase (той самий ключ) повертає наявний order. Підтверджено — grant-unique зміна не чіпає цей шлях.

## ВЕРДИКТ: ЗЕЛЕНИЙ
Усі blocking (B1, B2, B3, B4/H1, B5, H2, H3/G10, N1-N3) закрито. Три нові/залишкові — NON-BLOCKING; пункт 2 має бути уточнений у §4 таблиці переходів перед імплементацією T2.
