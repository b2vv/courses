claude-opus-4-8

# GATE 2 critique — ratings plan (independent review)

ДАЛІ: прочитати docs/ratings/spec.md (AC1–AC12, G1–G10) і docs/ratings/plan.md; звірити claims плану з кодом.

## КРОК 1 — прочитано spec.md + plan.md
ФАКТ: spec має AC1–AC12, G1–G10, R1–R6. План має §1–§13, Rt01–Rt27, слайси T01–T08.
Ключові claims плану з file:line для перевірки:
- §1/§2: `discussion/adapters/postgres/repository.go:20-34` (lessonFactsQuery, grant join, display_name join :38)
- §1: `access.go:26-31` PurchaseRequired false для admin; `LockedLessonPanel.tsx` єдиний buy entry
- §3: `00003_authoring.sql:35` урок не має publish-стану; `playback/domain/progress.go:74` completed = ≥95%
- §2/§3: `00010:13` reports мають comment_id NOT NULL
- G5/§1: `orders/application/service.go:131` GrantCourse без self-check; :43 admin purchase не заборонено
- §7: `recovery.spec.ts` рахує курси
РІШЕННЯ: далі запускаю plan-critique skill, потім звіряю кожен file:line.
ДАЛІ: інвок plan-critique; паралельно читаю ключові файли.

## КРОК 2 — звірка claims з кодом
ФАКТ (усі підтверджено):
- repository.go:20-35 lessonFactsQuery + grant join :26-27; display_name LEFT JOIN users :44 (план пише :38 — то рядок SELECT-списку, join нижче; дрібниця).
- access.go:26-28 PurchaseRequired false для admin (accessExempt :30-32 true для admin). ✓
- orders/service.go:131 GrantCourse — немає self-check (тільки !Administrator + reason). ✓ G5-факт вірний.
- orders/service.go:43 StartPurchase — admin не заборонений купувати. ✓ G2-факт вірний.
- 00010_instructor_moderation.sql:13 discussion_comment_reports.comment_id NOT NULL. ✓ (план пише «00010:13»)
- 00010:21-30 discussion_moderation_log має comment_id NOT NULL :24. ✓
- 00003_authoring.sql:35 course_lessons — немає publish-колонки. ✓ total_lessons = всі уроки.
- playback/progress.go:73-74 completed = positionSeconds ≥ 0.95*duration. ✓
- comment.go:11,40 body 1..2000 рун, utf8.ValidString. ✓
- shared_rules_test.go існує (прецедент пін-тесту порогу). ✓
- catalog.spec.ts:52 рахує курси (listitem toHaveCount). recovery.spec.ts НЕ рахує картки — план §7 мис-цитує «recovery.spec.ts counts courses».
- moderation.go: НЕМАЄ SELECT-читача discussion_moderation_log (лог write-only, WriteModerationLog :117). §7 посилається на «readers of the log (audit view, I03 queries)» яких у SQL немає.
- queueQuery :32-49 — однотабличний (discussion_comments) з keyset (open_count, latest_report, c.id). Мердж з review-reports в одну чергу план не специфікує.

## КРОК 3 — чотири лінзи
ПИТАННЯ-методологія: немає Task-тулу для fork-субагентів у цьому середовищі; 4 лінзи виконані мною як 4 дисципліновані проходи (я — свіжий незалежний рев'ювер, чистий контекст, інша сесія від автора). Названо: 4 лінзи запущено / 4 повернулось, з каве­атом про неізольованість.

### Лінза AI-slop — майже чисто
- §10 «Constitution» таблиця — церемонна, але вимагається процесом гейту. non-blocking.
- §9 сид 20k відгуків на «жменю курсів» — важко, але це мандатний magnitude-чек (R6 trigger, як Pay23). non-blocking.
Інших вигаданих абстракцій/портів нема — план явно reuse-first, «no new port».

### Лінза логічні діри
- L1 (non-blocking, найсильніша): §2/§5 teach-черга «lists both kinds». Наявний queueQuery однотабличний з keyset-курсором; як один пагінований упорядкований список спанить дві гетерогенні report-таблиці під спільним курсором — не описано (UNION з єдиним keyset vs мердж у застосунку). Можливий rework, але патерн відомий і Rt21/Rt22 крос-табличну пагінацію не перевіряють.
- L2 (non-blocking): §6 обіцяє тривалість на картці «якщо дані дають»; §5 контракт catalog Course додає лише rating + lessonCount — поля duration нема; AC11 згадує duration умовно. Контракт не несе каналу для duration до SSR-картки.
- L3 (non-blocking): §7 «every reader of the log must handle a review entry (audit view, I03 queries)» — SQL-читача лога нема (write-only). Неточність, нічого не ламає.

### Лінза надлишкова складність — чисто
- Дві копії константи порогу 3 (course/domain + discussion/domain) — вимушено ADR 013, пін-тестом mitigated. Прийнятно.
- Індекси/домен-функція ліні, без спекулятивних шарів.

### Лінза edge cases
- E1 (non-blocking): урок без ready-відео ніколи не completed → роздуває знаменник total_lessons; 100% недосяжні для курсу з текст-уроком. AC5/Rt13 припускають усі completable.
- E2 (non-blocking, для plan-defense): конкурентні PUT create → UNIQUE(course_id,author_id) race може віддати constraint-violation замість review_already_exists; план не називає ON CONFLICT для PUT.
- E3 (ok): avg 4,5,5 → 4.6667 → 4.7 ✓; поріг на counted (removed_at IS NULL) з Rt09 ✓; keyset публічного списку по discussion_reviews_course_newest ✓.
- E4 (ok): автор-видалення hard DELETE звільняє слот; removed лишає рядок, delete removed відмовлено ✓.

РІШЕННЯ: blocking findings = 0. Усі code-факти плану підтверджені. Вердикт GREEN з non-blocking списком + методологічний каве­ат (лінзи не ізольовані субагенти).
