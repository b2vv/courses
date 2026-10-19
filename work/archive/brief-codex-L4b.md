# Codex brief — L4b: communication gate (discussion) → CanCommunicate

Репо `courses`, гілка `feat/enrollment`. План §4. L1–L4a у дереві (L3a: `access.CanCommunicate`(admin∨owner∨active)
/`CanConsume`(admin∨owner∨(Published&&Preview)∨active|blocked); L3b: discussion domain `Course` вже несе
`Membership access.Membership`). Роль: реалізуєш L4b-backend; коміт/гейти/браузерний e2e — Courses 1.

РОБОЧИЙ ФАЙЛ: /Users/strelia/projects/courses/work/_wip-L4b-codex.md — append; перший рядок `ДАЛІ:`;
близько ліміту → RESUME. Фінал = переказ.

## Межа L4b (цей прогін)
Бекенд комунікаційного гейта в discussion: домен-обгортки+коди, application (write-paths→CanCommunicate,
read→CanConsume), HTTP-мапінг, application-unit + integration. **Браузерний e2e НЕ роби** — його зробить Courses 1.

## Семантика (ухвалено)
- **Комунікація** (писати): коментар Publish, ReportComment, PutReview, ReportReview — вимагає `active`
  членства (або owner/admin). Email-перевірка ПЕРША (вже є в усіх чотирьох).
- **Читання** lesson-коментарів (List) — `CanConsume` (active|blocked), як відео/транскрипт (blocked читає).
- **Публічні відгуки** (`PublicReviewPage`) — лишаються `CanBeViewedBy` (публічні метадані каталогу; НЕ гейтити
  членством, інакше зламаються публічні рейтинги). НЕ чіпай.
- **Порядок кодів (email-first, як на контенті L4a):** неверифікований → `email_not_verified`; потім
  blocked → `enrollment_blocked`; інакше не-член (none/revoked) → `enrollment_required`. owner/admin проходять.

## 1. Домен `discussion/domain`
- `access.go`: додай `func (c Course) CanCommunicateBy(a Actor) bool { return access.CanCommunicate(access.Subject{
  UserID:a.UserID,Administrator:a.Administrator}, access.Course{OwnerID:c.OwnerID,Published:c.Status==Published,
  Paid:c.Paid,Entitled:c.Entitled,Membership:c.Membership,Preview:c.Preview}) }`; аналогічно `CanConsumeBy`
  (→access.CanConsume). `CanBeViewedBy` НЕ чіпай (лишається для PublicReviewPage; AST не стосується discussion
  CanBeViewedBy, але залиш його тіло як є).
- `errors.go`: додай `ErrEnrollmentRequired = errors.New("enrollment_required")`,
  `ErrEnrollmentBlocked = errors.New("enrollment_blocked")`.

## 2. Application `application.go`
Допоміжник:
```
func requireCommunication(course domain.Course, actor Actor) error {
    if course.CanCommunicateBy(actor) { return nil }
    if course.Membership == access.MembershipBlocked { return domain.ErrEnrollmentBlocked }
    return domain.ErrEnrollmentRequired
}
```
(імпорт `contracts/access` у application — чи вже є? якщо ні, додай; це чистий тип, ADR-013 ок.)
Застосуй ПІСЛЯ наявного email-first у кожному write-path:
- **Publish** (~:145): прибери `CanBeViewedBy`-блок, постав `if err := requireCommunication(facts.Course, actor); err != nil { return Comment{}, err }`.
- **ReportComment** (~:248): так само на `record.Course`.
- **PutReview** (~:374): після email — побудуй `course := domain.Course{OwnerID:facts.Facts.OwnerID,Status:facts.Facts.Status,
  Paid:facts.Facts.Paid,Entitled:facts.Facts.Entitled,Membership:facts.Facts.Membership}`; `requireCommunication(course, actor)`;
  ПОТІМ наявний `domain.CanReview(facts.Facts, actor)` (owner/progress/published лишаються). Так blocked/не-член
  відсікається комунікаційним кодом ДО CanReview, а free-blocked-з-прогресом більше НЕ напише відгук (був баг).
- **ReportReview** (~:412): зараз лише email — ДОДАЙ членський гейт. Завантаж курс-факти відгука: `GetReview(reviewID)`
  дає course_id; тоді `ReviewFacts(courseID, actor.UserID)` → Membership; побудуй domain.Course; `requireCommunication`.
  (Якщо зручніше — один новий repo-метод, що віддає membership-факти курсу відгука; але двох наявних викликів досить.)
- **List** (read, ~:117): `CanBeViewedBy` → `CanConsumeBy` (лишаємо наявний forbidden/purchase-мапінг; email НЕ додаємо —
  читання read-only; blocked читає, free-не-член → ErrForbidden).
- **PublicReviewPage** (~:322): НЕ чіпати (CanBeViewedBy).

## 3. HTTP `adapters/http/handler.go`
- `statusFor`: додай `ErrEnrollmentRequired`, `ErrEnrollmentBlocked` → `http.StatusForbidden`.
- `writeError` message-switch: додай людські повідомлення для `enrollment_required`/`enrollment_blocked`.

## 4. Application-unit (fakes) — precedence явно (Courses 2)
Для Publish (мінімум) + по можливості ReportComment/PutReview/ReportReview, table-driven:
- **unverified+blocked → email_not_verified** (email виграє);
- **verified blocked → enrollment_blocked**;
- **verified non-member (none) → enrollment_required**; revoked → enrollment_required;
- verified active → ok; owner/admin → ok.
List: free-non-member→forbidden; blocked→читає; active→читає.
PutReview: active+progress→ok; blocked→enrollment_blocked (ДО CanReview); non-member→enrollment_required; owner→review_owner_forbidden (CanReview, після active).

## 5. Integration (`adapters/postgres/..._integration_test.go`)
Сідай членства прямим course_grants-insert (active/blocked/revoked) на дискусійному курсі/уроці й через
application-сервіс (справжній repo) перевір: active коментує; **blocked Publish → enrollment_blocked**; non-member
→ enrollment_required; blocked читає List (CanConsume); non-member List → forbidden. (Фікстури ізольовані — власні
user/course, не чіпай спільні; прибери за собою як решта.)

## Правила
- Онови ВСІ фейки/реалізації, яких торкнувся (компілятор). generated/sqlc не чіпай. НЕ комість.
- gofmt; build; `go test ./internal/discussion/... ./internal/contracts/...` + architecture зелені.
- Браузерний e2e — Courses 1.

## Критерій готовності
- 4 write-paths на requireCommunication (email-first збережено), List на CanConsume, PublicReviewPage цілий,
  CanReview після CanCommunicate; HTTP-коди 403; precedence-юніти + integration (blocked→enrollment_blocked,
  non-member→enrollment_required) зелені. Робочий файл: що куди, перелік precedence-кейсів, як ReportReview
  вантажить факти.
