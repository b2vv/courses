# Codex brief — L3a: membership-state predicates in contracts/access (pure, unit-only)

Репо `courses`, гілка `feat/enrollment`. План `docs/enrollment/plan.md` §2. L1/L2 у дереві. Роль: реалізуєш L3a;
коміт/гейти — Courses 1. **Межа L3a: ЛИШЕ `internal/contracts/access` + його юніти.** Жодних читачів/модулів/SQL
(це L3b). Жодного гейтингу (L4). Поведінка застосунку НЕ міняється (нові предикати ніхто ще не кличе).

РОБОЧИЙ ФАЙЛ: /Users/strelia/projects/courses/work/_wip-L3a-codex.md
Append після КОЖНОГО кроку; перший рядок `ДАЛІ:`. Фінал = стислий переказ.

## Факти (звірено картою)
`backend/internal/contracts/access/access.go`:
- `Subject{UserID string; Administrator bool; EmailVerified bool}`.
- `Course{OwnerID string; Published bool; Paid bool; Entitled bool; Preview bool}`.
- `CanView = accessExempt || course.Published && (!course.Paid || course.Preview)`.
- `accessExempt = Administrator || (UserID!="" && UserID==OwnerID) || course.Entitled`.
- `CanStream = CanView && EmailVerified`; `PurchaseRequired = Published && Paid && !Preview && !accessExempt`.

**Обмеження, які НЕ можна порушити (AST + TruthTable):**
- `platform/architecture/access_mapping_ast_test.go`: РІВНО 4 методи `CanBeViewedBy`, кожен — один стейтмент
  `return access.CanView(...)`. → НОВІ предикати мусять мати ІНШІ імена (НЕ `CanBeViewedBy`).
- `shared_rules_test.go` + `accesstest/table.go` TruthTable (14 рядків, факт `Granted`→`Entitled`) тестують
  `CanBeViewedBy`/`PurchaseRequiredBy`. **НЕ чіпай `CanView`, `accessExempt`, `PurchaseRequired`, `Entitled`,
  `Course` наявні поля, TruthTable** — усе лишається ідентичним (жодного зсуву CanView/CanStream).

## Що додати (адитивно)
1. `type Membership string` + консти: `MembershipNone=""`, `MembershipActive="active"`, `MembershipBlocked="blocked"`,
   `MembershipRevoked="revoked"`. (Відповідають `orders/domain.GrantState`; none = немає рядка.)
2. Нове поле в `Course`: `Membership Membership` (адитивно; наявні поля й їхній порядок-семантику не чіпати).
3. **`func CanConsume(subject Subject, course Course) bool`** — доступ до КОНТЕНТУ уроку:
   `return subject.Administrator || (subject.UserID != "" && subject.UserID == course.OwnerID) ||
    (course.Published && course.Preview) || course.Membership == MembershipActive || course.Membership == MembershipBlocked`.
   (blocked ДАЄ контент; revoked/none — ні; preview відкриває лише коли курс Published — дзеркалить CanView,
    інакше widening draft/archived-preview; членство Published-незалежне — archived-member/G3. Для free і paid —
    гейт вмикає L4.) **B-L3a.1 (Courses 2): Preview мусить бути під Published-гардом.**
4. **`func CanCommunicate(subject Subject, course Course) bool`** — комунікація (коментар/оцінка/відповідь):
   `return subject.Administrator || (subject.UserID != "" && subject.UserID == course.OwnerID) ||
    course.Membership == MembershipActive`.
   (blocked → false; none → false. Гейт `email_verified` — ОКРЕМИЙ, у застосунку L4, НЕ тут: щоб L4 міг
    віддати окремі коди `email_not_verified` vs `enrollment_required`/`enrollment_blocked`.)

Обидва — чисті функції stdlib, без I/O. НЕ виводь `Entitled` з `Membership` у контракті (читачі виставлять
`Entitled` явно як `state='active'` у L3b; у L3a `Entitled` лишається як є).

## Тести (`access_test.go`, table-driven — першими за TDD)
Матриця для CanConsume і CanCommunicate:
- Membership ∈ {none, active, blocked, revoked} × роль {stranger, owner (UserID==OwnerID), admin} × Preview {false,true}.
- CanConsume=true якщо: admin | owner | Preview | membership active|blocked; інакше false.
- CanCommunicate=true якщо: admin | owner | membership active; інакше false (blocked=false!).
- Явні кейси-пастки: blocked → CanConsume=true, CanCommunicate=false; revoked → обидва false (крім owner/admin);
  none+preview → CanConsume=true, CanCommunicate=false; stranger active → обидва true.

## Правила
- Нічого поза `internal/contracts/access`. generated не чіпай. НЕ комість.
- `gofmt`; `go build ./...`; `go test ./internal/contracts/access/...` зелені; і переконайся, що
  `go test ./platform/architecture/...` (AST+TruthTable) ЛИШАЄТЬСЯ зеленим (ти нічого в CanBeViewedBy/CanView
  не міняв — підтвердь прогоном).

## Критерій готовності
- Нові типи/предикати + юніт-матриця зелені; AST/TruthTable/наявний access_test зелені (CanView/CanStream/Entitled
  недоторкані). У робочому файлі: що додано, підтвердження що CanBeViewedBy=4 і TruthTable не чіпано.
