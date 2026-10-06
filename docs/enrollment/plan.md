# Фаза `enrollment` — план

> ЯК. Спирається на `spec.md` (GATE 1 2026-10-06, G1–G11, En01–En14). GATE 2: 1-й прохід ЧЕРВОНИЙ
> (критика B1–B5, захист H1–H3); дописано нижче. Статус: чекає re-review.

## 0. Критичний наслідок моделі

`entitled` = `EXISTS active course_grants` рахується окремо в SQL кожного модуля; `access.CanView` пускає
**free без членства**. G1/G2/G9/En03 змінюють контракт: членство = один авторитетний запис, і **free-контент
теж вимагає членства**. Це крос-модульно → **повний e2e обов'язковий**; міграція не замикає наявних (backfill, §5).

**Місця, що рахують факт доступу (ВСІ треба змінити):**
- `course/adapters/postgres/learning.go:46` (навчальна сторінка/аутлайн);
- `playback/adapters/postgres/playback.go:29` (плеєр);
- **`playback/adapters/postgres/access.go:19` (nginx auth_request на самі відео-байти — B1, найкритичніший)**;
- `transcript/...` (читання транскрипту уроку);
- `discussion/adapters/postgres/*` (мапить `entitled` у `Course` для читання І комунікації — B3).

## 1. Факт членства — СТАН, не boolean (B3)

Сьогодні модулі маплять один `grants.id IS NOT NULL` (state='active') → `Course.Entitled`, і цей бул годує
і читання, і публікацію. Enrollment потребує розрізняти контент (active|blocked) і комунікацію (active).
**Рішення:** у `contracts/access` нести **стан членства** (enum `membership ∈ {none, active, blocked, revoked}`)
на додачу до наявного `Entitled` (Entitled лишається — його читає `CanView`, якого НЕ чіпаємо, N2/B1-урок).
Кожен модуль у своєму SQL обчислює стан членства (`course_grants` для (user,course)) і мапить у Subject/Course.

## 2. Нові чисті предикати в `contracts/access` (N1: у контракті, не в use-case)

- `CanView` / `CanBeViewedBy` — **без змін** (опис + аутлайн курсу лишаються публічними метаданими; N3). AST/
  TruthTable цілі (B1-урок accounts).
- **`CanConsume(subject, course)`** — доступ до КОНТЕНТУ уроку: `membership ∈ {active, blocked}` **або** exempt
  (owner/admin/preview-урок). playback(обидва шляхи)+transcript кличуть його; `revoked`/`none` → відмова.
- **`CanCommunicate(subject, course)`** — комунікація: `membership == active` (блок і не-член — ні).
  Доменний метод модуля кличе предикат; HTTP мапить **два різні коди**: `enrollment_required` (не член, B2) і
  `enrollment_blocked` (blocked). Порядок із `email_not_verified` (A05): спершу verified, потім членство.
Усі — чисті функції в `contracts/access`; `Entitled` лишається (адитивно, не заміна — N2).

## 3. Модель даних — ОДИН рядок на (user,course) (B4/H1)

`course_grants` (reuse, переосмислено як membership). Проблема: наявний частковий унікальний індекс
`WHERE state='active'` дозволяє 2-й active поряд із blocked/revoked → обхід мюту/revoke + дублі в JOIN.
**Рішення:** **повна унікальність `UNIQUE(user_id, course_id)`** — рівно один рядок на пару; усі операції
(enroll/purchase/block/unblock/revoke) — **переходи стану цього рядка (UPSERT)**, не нові INSERT-и.
Наслідки, що це розвʼязує: self+purchase (оплата на наявному self → upgrade source self→purchase, той самий
рядок), re-enroll після revoke (revoked→active на тому ж рядку, свідомо — для free; для paid revoke знімає
лише адмін), інваріант G1 «єдиний authoritative запис».
Payment-insert (`orders/adapters/postgres/repository.go` ~135/182/252) та `CanStartOrder` (grant.go) —
**переписати на UPSERT/transition** замість insert-на-партковий-унік.

## 4. Стан-машина й гейти

`(none)` --enroll(self)/purchase--> `active` --block--> `blocked` --unblock--> `active`;
`active|blocked` --revoke(admin)--> `revoked`; `revoked` --admin re-grant / (free) re-enroll--> `active`.
Архів курсу стан НЕ міняє (володіння). Інваріанти + unit-таблиця переходів.
- **Контент**: `CanConsume` у playback(playback.go+access.go)/transcript/course-learning — JOIN
  `state IN ('active','blocked')`.
- **Комунікація**: `CanCommunicate` у discussion Publish/ReportComment/PutReview/ReportReview — `state='active'`;
  не-член → `enrollment_required`, blocked → `enrollment_blocked` (обидва ≠ `email_not_verified`).
- Purchase (orders) членство НЕ гейтить (оплата створює членство). **`CanStartOrder` + purchase-на-наявному
  членстві (пін перед T2):** якщо є рядок у стані `active`|`blocked`|`revoked` — purchase НЕ створює 2-го,
  НЕ знімає `blocked` (мют адмін/інструктор не обходиться оплатою) і НЕ воскрешає `revoked` (адмін-revoke
  навмисний — відновлює лише адмін); гроші, що прослизнули, → наявна сітка `refund_required`. **Поведінкова
  зміна:** re-purchase після revoke більше НЕ відновлює доступ (enrollment старший за grant). Ребра — в
  unit-таблиці переходів §4.
- **Читання дискусії = `CanConsume`, не `Entitled`** (узгодженість blocked): перегляд коментарів/відгуків — це
  споживання (членство `active`|`blocked`, як відео/транскрипт); публікація — `CanCommunicate` (`active`). Інакше
  blocked платний губив би читання дискусії, зберігаючи відео (неузгодженість, знахідка re-review #3). На free
  читання теж вимагає членства (узгоджено з «free-контент вимагає членства»).

## 5. Сайд-ефекти · міграція · rollback

**Сайд-ефекти:** free-контент тепер вимагає членства (зміна поведінки); зміна JOIN у course/playback(×2)/
transcript/discussion → повний e2e; нові гейти комунікації; нова таблиця person-reports.

**Міграція `00015_enrollment.sql`:**
- UP: розширити CHECK `source`(+`self`), `state`(+`blocked`); **DROP** `course_grants_one_active_per_user_course`,
  **ADD `UNIQUE(user_id,course_id)`**; `CREATE TABLE participant_reports` (id, course_id, reporter_id, target_id,
  kind{on_lecturer,on_student}, reason, state{open,resolved}, created_at; `CHECK reporter_id<>target_id`
  (анти-self-report); `UNIQUE(course_id,reporter_id,target_id) WHERE state='open'` (дедуп).
  **Backfill (B5):** INSERT self-active членства для кожного (user,course) на **free** курсі, де user має АБО
  рядок у `lesson_progress`, АБО discussion-активність (коментар/відгук) — **union**, щоб не вибити наявних
  комунікаторів; на конфлікті `UNIQUE(user,course)` — нічого (вже є).
- Seed/фікстури: сідовані студенти → self-active на сідованих free-курсах (інакше free-e2e playback/comment
  впадуть — клас E1).
- DOWN: `DROP TABLE participant_reports`; **DELETE рядки source='self'**; `UPDATE … SET state='revoked' WHERE
  state='blocked'` (конверсія перед звуженням); DROP `UNIQUE(user,course)`, відновити партковий-active-унік;
  звузити CHECK. Backfill — forward data-міграція (revert гілки її не відкочує; названо).

**Тести, що пінять схему (оновити разом із міграцією, re-review #1):** `playback/adapters/postgres/
access_integration_test.go` жорстко перевіряє назву `course_grants_one_active_per_user_course` в EXPLAIN-плані
media-bytes — DROP індексу вимагає оновити цю перевірку; інтеграційні тести, що вставляють `course_grants`,
потребують нових станів/джерел (self/blocked). Названо явно (AGENTS: хто змінює глобально-лічене — прибирає).

**Регресія, свідомо прийнята (En13-прим.):** free-користувач **без жодної активності** (лише дивився опис/аутлайн,
не стартував урок, не коментував) членства не отримає й муситиме «Записатися», щоб відкрити контент. Узгоджено
з G2 (явний запис). Опис+аутлайн лишаються видимими (метадані, N3).

## 6. Конституція
Домен чистий (переходи, person-report policy); application — домен+порти; `contracts/access` — чисті предикати,
`CanBeViewedBy`=`CanView` не чіпаємо (AST/TruthTable), `CanConsume`/`CanCommunicate` — нові окремі (B3/N1/N2);
generated у адаптерах; **person-reports у `discussion` читає членство ВЛАСНИМ SQL (course_grants), без імпорту
orders** (ADR-013, підтверджено критикою); модуль не імпортує модуль. TS strict/zod/React Aria/a11y.

## 7. Приймальні сценарії (делеговано `acceptance-spec`; рядки En→рівень)

| En | Сценарій | Рівень |
|---|---|---|
| En01 | free «Записатися» → active(self); «ви учасник»; у «Моє навчання» | E |
| En02 | paid purchase → active(purchase), лінк order; **на наявному self → upgrade source, той самий рядок** | I/E |
| En03 | контент члену active\|blocked; неучаснику paid відмова; free не-член — контент закрито, **опис+аутлайн видно** | E |
| En04 | admin revoke → контент відрізано (вкл. **media-bytes auth_request**); запис лишається | I/E |
| En05 | blocked: контент так (вкл. media-bytes), комунікація `enrollment_blocked`; інші курси цілі | C/E |
| En06 | instructor block/unblock свій; чужий — відмова; admin будь-де + revoke | C/E |
| En07 | скарга учень→лектор і лектор→учень у адмін-черзі; blocked теж; self-report і дубль — відмова | C/E |
| En08 | overview ростер: ім'я/%/статус/дата; пагінація; без email; **0 уроків → 0%, порожній ростер — порожній стан** | E |
| En09 | архів зберігає доступ наявним; enroll на archived/unpublished — **відмова (named guard)** | I/E |
| En10 | grandfather=active-членство; preview-урок без запису | I/E |
| En11 | комунікація = active + email_verified; не-член→`enrollment_required`, blocked→`enrollment_blocked`, неverified→`email_not_verified` | C/E |
| En12 | a11y нових екранів (клавіатура/фокус/44px/zoom/uk-en/axe) | E(axe) |
| En13 | **backfill**: наявний free-користувач із прогресом АБО коментарем/відгуком лишається з доступом/комунікацією після міграції | I/E |
| En14 | **G10**: інструктор/адмін «Записується» учасником ЧУЖОГО курсу (членство active, без зміни ролі); owner/admin бачать свій/будь-який без запису | I/E |
| En15 | **один рядок (user,course)**: blocked/revoked клікає «Записатися» → НЕ створює 2-й active (перехід або відмова), без дублів у JOIN | I |

Ручних: 0.

## 8. Порядок тасків
(T1) домен grant→membership (source self, state blocked, переходи, інваріанти) + міграція 00015 (CHECK, unique(user,course), participant_reports, backfill union, seed) + unit.
(T2) orders application/adapters: enroll/block/unblock/revoke/membershipState/roster + **UPSERT-переходи** (переписати insert+CanStartOrder) + queries.
(T3) contracts/access: enum membership + `CanConsume`/`CanCommunicate` (+юніт); оновити ВСІ JOIN/мапінги — learning.go, playback.go, **playback access.go**, transcript, discussion.
(T4) гейти: контент (CanConsume) у playback(×2)/transcript; комунікація (CanCommunicate, `enrollment_required`/`enrollment_blocked`) у discussion.
(T5) person-reports у discussion: домен (self-report/дедуп інваріанти)+міграція-частина (в 00015)+application+адмін-черга (лише адмін).
(T6) HTTP усіх операцій + openapi + Orval.
(T7) фронт: enroll/«Моє навчання», інструкторський ростер+block, адмін-черга+revoke, форма скарги; стани (0%/порожньо); i18n/a11y.
(T8) e2e `enrollment.spec.ts` + axe (вкл. backfill, media-bytes-blocked, single-row, G10).
(T9) упорядкований `make check` + звіт + PR.
Кожен таск посилається на рядки En.
