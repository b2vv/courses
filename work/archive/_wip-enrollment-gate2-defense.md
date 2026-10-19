claude-opus-4-8

# GATE 2 plan-defense — фаза enrollment (незалежний рецензент)

ДАЛІ: прочитати spec.md + plan.md, звірити claims з кодом, побудувати таблицю загроз, вердикт.

## ФАКТИ (звірка плану з кодом)

- Шляхи плану (`course/adapters/postgres/learning.go:46`, `playback/.../playback.go:29`) насправді
  під `backend/internal/...`. Рядки збігаються: learning.go:45-46 `LEFT JOIN course_grants ... state='active'`;
  playback.go:28-29 те саме. CanView (access.go:22-24) пускає free published без Entitled — підтверджено.
- `contracts/access/access.go`: Subject/Course/CanView/CanStream/PurchaseRequired/accessExempt — як у плані.
  `CanBeViewedBy`=`CanView` у transcript/domain/access.go:39-41 (AST/TruthTable, B1) — план його НЕ чіпає. ✔
- `orders/domain/grant.go`: source {purchase,admin,grandfather}, state {active,revoked}; NewGrant інваріант
  `(source==purchase)⇔orderID` (:42-43). `CanStartOrder` (:68-74) рефузить лише за наявного **active** гранту.
- Міграція 00011_payment.sql:40-42 — **partial unique index `course_grants_one_active_per_user_course`
  ON (user_id,course_id) WHERE state='active'**. Тобто унікальність ТІЛЬКИ на active; revoked/blocked слоту
  не займають. CHECK source IN (purchase,admin,grandfather); state IN (active,revoked) (:32,:34).
- Міграції: найвища 00014_accounts → **00015 — коректний наступний номер**. ✔ course_grants чіпає лише 00011.
- `lesson_progress` — ЄДИНА таблиця прогресу (watched=completed_at, progress=position_seconds); окремої
  «watched marks» таблиці нема → backfill з одного джерела покриває watched+progress. Але `ensureProgress`
  (progress.go:28-31) створює рядок лише в ProgressWriter.Save (на playback-апдейт), не на простий перегляд.
- Discussion moderation queue (`moderation.go:16-34`) — **instructor-scoped** (join instructor_id).
  Person-reports — НОВА сутність (grep порожній), адмін-черга = net-new адмін-surface у discussion.

## ТАБЛИЦЯ ЗАГРОЗ (5 питань plan-defense)

| # | Питання | Закрито рядком плану | Вердикт |
|---|---|---|---|
| 1 | кривий інпут | `plan §8 T2 "23505-safety"`, `§5 CHECK kind/state` частково; **enroll на archived/unpublished — жодного рядка** (En09 вимагає, предикат не названо); self-report (reporter==target) не адресовано | **дірка** |
| 2 | відмова залежності | — (реюз ErrDependencyUnavailable мається на увазі, але жодного рядка) | **дірка (слабка)** |
| 3 | гонка / повторний виклик | `§8 T2 "23505-safety"` закриває лише ДВА active (partial index); **re-enroll після revoke та re-enroll/re-purchase під blocked створюють новий active поряд із revoked/blocked → обхід revoke та mute — НЕ закрито жодним рядком** | **дірка (блокуюча H1)** |
| 4 | порожні / граничні | `§7 En08 пагінація` названо; zero-lessons `% пройденого` (ділення 0/0), порожній ростер — не названо | **дірка (слабка)** |
| 5 | rollback | `§5 DOWN` існує, але сам план пише «self/blocked рядки → **?**» — звуження CHECK впаде на backfill-рядках self/blocked; backfill — незворотна дата-мутація, яку revert гілки не відкотить | **дірка (блокуюча H2)** |

## БРИФ-СПЕЦ ПЕРЕВІРКИ

- **Backfill/00015**: покриває paid/admin/grandfather (вони вже active-grant, доступ зберігають), owner/admin
  (exempt), free-з-прогресом (self-insert). НЕ покриває **free БЕЗ progress-рядка** (відкривали опис/
  дивились без Save) → після міграції мусять явно записатись. Узгоджено з G2 (явний запис), але план НЕ
  фіксує це рішенням/out-of-scope — мовчазна зміна поведінки. e2e/seed free → план §5 обіцяє self-членство
  сідованим студентам (клас E1). Часткова дірка (не блокуюча, якщо дописати рядок-рішення).
- **Гонки**: подвійний active enroll — закрито partial unique index. self+purchase на один курс (ціна
  free→paid): purchase-insert впаде на unique-active (23505) при наявному self-active → оплата взята, grant
  нема — **план мовчить** (слабка дірка). block під час publish — TOCTOU на читанні стану (слабка).
- **«Володіння»**: архів стану не міняє (§4) ✔; revoke платного = оверрайд адміна над володінням — узгоджено з
  §24 п.9 (spec G5 «запис зберігається для сліду») ✔; розблок → active ✔. АЛЕ revoke оборотний самозаписом (H1).
- **revoke vs refund**: spec Out-of-scope:80 «`revoke` ≠ refund», план не чіпає order/payment_events — ✔ лише доступ.
- **Повний e2e на contracts/access**: `§0 рядки 13-14` + `§5 "повний e2e обов'язковий"` — ✔ закрито явно.
- **Person-reports**: чергу бачить лише адмін (G4/G6, spec En07) ✔; заблокований може скаржитись (gate-список
  §2 person-report НЕ містить) ✔; **дублі/гонки скарг — нема unique/дедуп на participant_reports** (слабка).
  Нотатка: адмін-surface у discussion net-new (наявна черга instructor-scoped).
- **G→En покриття**: G1-G9,G11 мають рядки (En01-En11). **G10 (привілейований юзер записується звичайним
  учасником на ЧУЖИЙ курс, адитивно, без зміни ролі) — НЕМАЄ acceptance-рядка**; En03 покриває лише
  exempt-ПЕРЕГЛЯД owner/admin, не enroll-as-participant half. → блокуюча дірка H3 (названий G без рядка).

## ВЕРДИКТ: 🔴 ЧЕРВОНИЙ

Блокуючі (3): H1 (рядок #3 — обхід revoke/block самозаписом), H2 (рядок #5 — DOWN + незворотний backfill),
H3 (G10 без acceptance-рядка). Слабкі: #1 enroll-on-archived/self-report, #2 dependency, #4 граничні,
self+purchase конфлікт, free-без-прогресу рішення, дедуп скарг.
