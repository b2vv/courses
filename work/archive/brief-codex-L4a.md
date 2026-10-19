# Codex brief — L4a: content gate → CanConsume (behavior change)

Репо `courses`, гілка `feat/enrollment`. План §4. L1–L3 у дереві (L3: `access.CanConsume`/`CanCommunicate` +
`Course.Membership`; усі читачі несуть membership + явний Entitled=state='active'). Роль: реалізуєш L4a; коміт/
гейти — Courses 1. **Межа L4a: ЛИШЕ контент-гейт** (playback ×2 + transcript). Комунікація (discussion) — L4b,
НЕ чіпай.

РОБОЧИЙ ФАЙЛ: /Users/strelia/projects/courses/work/_wip-L4a-codex.md — append; перший рядок `ДАЛІ:`;
близько ліміту → RESUME. Фінал = переказ.

## Поведінкова зміна (ухвалено GATE 1: G2/En03)
Контент уроку (відео-байти, сесія плеєра, транскрипт) тепер вимагає ЧЛЕНСТВА: `CanConsume` (membership
active|blocked, або preview/owner/admin) замість view-only. Free-контент більше НЕ відкритий не-члену (лише
опис+аутлайн=метадані лишаються видимими через CanView, які НЕ чіпаємо).

## Критичний нюанс: email треба РОЗЧЕПИТИ з view
Сьогодні гейт: `if !CanBeViewedBy {purchase/forbidden}; if !CanStreamBy {email}`. `CanStreamBy = CanView &&
EmailVerified`. Якщо лишити CanStreamBy як email-гейт після CanConsume — **blocked платний член** має
CanConsume=true, але CanView=false (Entitled(blocked)=false) → CanStreamBy=false → хибний ErrEmailNotVerified.
Тому email-перевірку робимо ЧИСТОЮ: `if !actor.EmailVerified { ErrEmailNotVerified }`.

## Зміни
### 1. Доменні обгортки (нові, НЕ CanBeViewedBy — AST!)
- playback `domain/authorization.go`: додай `func (course VideoCourse) CanConsumeBy(actor Actor) bool { return
  access.CanConsume(access.Subject{UserID: actor.UserID, Administrator: actor.Administrator}, access.Course{
  OwnerID: course.OwnerID, Published: course.Status==Published, Paid: course.Paid, Entitled: course.Entitled,
  Membership: course.Membership, Preview: course.Preview}) }`. (email не потрібен CanConsume.)
- transcript `domain/access.go`: аналогічний `CanConsumeBy` на його Course.
- `CanBeViewedBy` лишається рівно 4 і кожен `return access.CanView(...)` — НЕ чіпати.

### 2. Контент-гейти (3 сайти) — замінити view-check на consume + чистий email
Патерн (однаковий для трьох):
```
if !course.CanConsumeBy(actor) {
    if course.PurchaseRequiredBy(actor) { return ...ErrPurchaseRequired }   // платний не-член
    return ...ErrForbidden                                                  // free не-член / revoked
}
if !actor.EmailVerified { return ...ErrEmailNotVerified }
```
- playback `application/authorize.go` (~:31 CanBeViewedBy, ~:37 CanStreamBy) — media-bytes auth_request, НАЙКРИТИЧНІШЕ.
- playback `application/sessions.go` (~:74, ~:80).
- transcript `application/read.go` (~:60, ~:66).
Зберігай наявні типи помилок (ErrPurchaseRequired/ErrForbidden/ErrEmailNotVerified) — нових кодів для контенту
НЕ вводь (frontend-CTA «Записатися» бере з MembershipState-ендпойнта, L2b).

### 3. Прибрати мертве
Після заміни модульні `CanStreamBy`-обгортки (playback/transcript) більше ніхто не кличе → видали їх. Контрактну
`access.CanStream` + її юніт НЕ чіпай (легітимний предикат; лишається). Якщо компілятор покаже ще когось, хто
кликав CanStreamBy — повідом у робочому файлі, не видаляй наосліп.

## Інтеграційні/юніт тести (ОБОВ'ЯЗКОВО — denial мусить виконатись)
Доменні юніти CanConsumeBy (playback/transcript): matrix membership×роль×preview×published (дзеркалить L3a).
Application/integration для КОЖНОГО з 3 гейтів:
- active член verified → доступ (nil).
- blocked член verified → **доступ** (blocked споживає контент) — ключовий кейс, якого стара CanStreamBy зламала б.
- revoked член verified → ErrForbidden (free) / ErrPurchaseRequired (paid) — контент відрізано.
- не-член free verified → ErrForbidden (free-контент тепер закрито — поведінкова зміна).
- не-член paid verified → ErrPurchaseRequired.
- preview-урок не-члена verified → доступ; preview неверифікованого → ErrEmailNotVerified.
- member unverified → ErrEmailNotVerified.
- owner/admin → доступ без членства.
(Ці denial-кейси — доказ; зелений e2e з лише active членами нічого не довів би.)

## Правила
- Онови наявні тести гейтів, що очікували старої (view-only) поведінки — free-не-член тепер forbidden.
- generated/sqlc не чіпай. НЕ комість. gofmt; build; `go test ./internal/playback/... ./internal/transcript/...
  ./internal/contracts/...` + `platform/architecture` зелені. Інтеграцію/e2e — Courses 1.

## Критерій готовності
- 3 гейти на CanConsumeBy + чистий email; CanConsumeBy-обгортки додано; мертві CanStreamBy прибрано; CanBeViewedBy=4
  недоторкані; denial-тести (blocked=доступ, revoked/free-не-член=відмова) зелені на юніт/інтеграції. Робочий файл:
  які сайти змінено, що видалено, перелік denial-кейсів.
