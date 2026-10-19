# Codex brief — L3b: plumb membership state into all 7 readers (NO gating)

Репо `courses`, гілка `feat/enrollment`. План §1/§5. L1/L2/L3a у дереві (L3a: `access.Membership` + поле
`Course.Membership` + `CanConsume`/`CanCommunicate`). Роль: реалізуєш L3b; коміт/гейти — Courses 1.

РОБОЧИЙ ФАЙЛ: /Users/strelia/projects/courses/work/_wip-L3b-codex.md — append після кожного кроку; перший рядок
`ДАЛІ:`; близько ліміту → `RESUME` на межі фази. Фінал = переказ.

## Межа й ІНВАРІАНТ L3b
Плюмбінг стану членства в УСІХ читачів. **ЖОДНОГО гейтингу** (гейти CanConsume/CanCommunicate — L4). Поведінка
НЕ міняється на L3b.
**ГОЛОВНИЙ ІНВАРІАНТ (Courses 2):** кожен читач мусить виводити `Entitled` ЯВНО як `state = 'active'`, НЕ як
`grants.id IS NOT NULL`. Стара еквівалентність (`id IS NOT NULL` == active) трималась ЛИШЕ на фільтрі
`AND state='active'`. Коли прибираєш фільтр (щоб прочитати всі стани), будь-який залишений `id IS NOT NULL`-проксі
почне бачити blocked/revoked як present → `Entitled` хибно true → revoked відновлює доступ. Це регрес на L3b,
НЕвидимий у e2e (сьогодні немає non-active членів). Тому: (1) конвертуй кожен читач, (2) доведи фікстурами.

## Патерн конверсії (однаковий для всіх)
Було: `LEFT JOIN course_grants grants ON grants.user_id=$ AND grants.course_id=... AND grants.state='active'`
+ `SELECT grants.id IS NOT NULL` → entitled.
Стало: `LEFT JOIN course_grants grants ON grants.user_id=$ AND grants.course_id=...` (БЕЗ state-фільтра)
+ `SELECT grants.state` (nullable). Під повною UNIQUE(user,course) — не більше одного рядка, фан-ауту немає.
Мапінг у коді:
- `membership := access.MembershipNone` якщо state NULL; інакше `access.Membership(state)` (active/blocked/revoked).
- `entitled := (state.Valid && state.String == "active")` — ЯВНО. Це єдине джерело Entitled.
Клади `membership` у нове поле domain-Course модуля (див. нижче) і `entitled` — у наявне `Entitled`.

## Доменні Course кожного модуля — додати поле membership + прокинути в access.Course
Кожен модуль має власний domain Course/VideoCourse, що в `CanBeViewedBy` будує `access.Course{...}`. Додай у
кожен domain-Course поле `Membership access.Membership` (або власний тип, що мапиться) і виставляй
`access.Course{... Membership: course.membership}` у місці, де вже будується access.Course. **`CanBeViewedBy`
лишається рівно `return access.CanView(...)`** (AST!) — ти лише додаєш поле в літерал Course, не міняєш тіло на
інший виклик. (CanConsumeBy/CanCommunicateBy-обгортки НЕ додавай — це L4.)

## 7 читачів (чеклист — тикай кожен у робочому файлі)
1. **course** `adapters/postgres/learning.go` — SELECT `grants.id IS NOT NULL` :25; JOIN state='active' :45-46;
   scan `entitled` :99; map `LearningCourse{Entitled}` :112. Домен `course/domain/learning.go` LearningCourse.
2. **playback sessions** `adapters/postgres/playback.go` — :14 bool; :28-29 JOIN; scan :51; VideoCourse{Entitled}.
   Домен `playback/domain/authorization.go` VideoCourse.
3. **playback media-bytes (НАЙКРИТИЧНІШЕ)** `adapters/postgres/access.go` — :13 bool; :19-20 JOIN; scan :42;
   VideoCourse{Entitled} :53. Той самий VideoCourse-домен.
4. **transcript** `adapters/postgres/repository.go` getLessonAccessCourse :31-40 (:33 bool, :38-39 JOIN); scan :74;
   `transcript/domain` Course{Entitled} :82.
5. **discussion lessonFacts** `adapters/postgres/repository.go` lessonFactsQuery :20-35 (:21 bool, :26-27 JOIN);
   scan :88; Course.Entitled :99. Домен `discussion/domain/access.go` Course.
6. **discussion commentQuery** same file :65-76 (:68 bool, :73-74 JOIN); scanRecordFields :274-323 (:284);
   c.Course.Entitled :320.
7. **discussion reviewFacts (EXISTS-форма!)** `adapters/postgres/review.go` reviewFactsQuery :17-25, :19 =
   `EXISTS(... g.state='active')` → `result.Facts.Entitled` :50. Конвертуй на читання state:
   замінити EXISTS на `LEFT JOIN course_grants g ON g.course_id=courses.id AND g.user_id=$2` + `SELECT g.state`;
   `Entitled = state='active'` ЯВНО; прокинь membership у `ReviewFacts` (нове поле) для L4. (ReviewFacts —
   `discussion/domain/review.go`, має власний Entitled; див. bypass нижче.)

## Bypass-споживачі Entitled (НЕ через CanView) — НЕ зламати
- `discussion/domain/review.go` `CanReview` читає `facts.Entitled` прямо. Лишається коректним, бо Entitled
  тепер = active (та сама семантика). НЕ міняй CanReview у L3b (його членськи-гейт — L4-рішення). Додай
  `ReviewFacts.Membership` для L4, але поведінку CanReview не чіпай.
- `course/domain/learning.go` `CanPurchaseBy` читає `Entitled` прямо — лишається коректним (active). Не чіпай.

## Інтеграційні фікстури (ОБОВ'ЯЗКОВО — доказ інваріанту)
Для КОЖНОГО з 5 postgres-читачів (course learning, playback playback.go, playback access.go, transcript,
discussion lessonFacts+commentQuery+reviewFacts) додай/розший інтеграційний тест: сідай користувача з членством
**blocked** і окремо **revoked** на курсі, завантаж через читач і **assert `Entitled == false`** (і, де є поле,
`Membership == blocked|revoked`); а member **active** → `Entitled == true`, `Membership == active`. Без цього
конверсія неперевірена. (Фікстура потрібна і L4.)

## Правила
- `CanView`/`accessExempt`/`PurchaseRequired`/`Entitled`-поле/TruthTable/AST — не чіпати; рівно 4 CanBeViewedBy,
  кожен `return access.CanView(...)` (лише додаєш Membership у Course-літерал, тіло методу те саме).
- sqlc НЕ генеруй (сирий SQL). generated не чіпай. НЕ комість.
- gofmt; build; `go test ./internal/...` (юніти) зелені; vet. Інтеграцію (Docker) — Courses 1 (повний набір).
- Повний e2e НЕ є доказом тут (немає non-active членів) — доказ лише non-active інтеграційні фікстури на кожному читачі.

## Критерій готовності
- Усі 7 читачів конвертовані (Entitled явно = state='active'); кожен domain-Course несе Membership і прокидає
  в access.Course; CanBeViewedBy недоторкані. Юніти зелені. У робочому файлі — чеклист 7/7 + перелік non-active
  інтеграційних assert-ів по кожному читачу.
