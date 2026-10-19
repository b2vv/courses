# _wip L2b (Codex) — enrollment write operations

(Codex дописує тут після кожного кроку.)

ДАЛІ: Фаза 1 — додати TDD-таблицю для `Grant.Reinstate`, потім реалізацію доменного переходу і коди помилок.

Фаза 1: додано й перевірено `Grant.Reinstate(source, grantedBy)`: тільки `revoked → active`, із заміною source/grantedBy та `order_id=nil`; інші стани повертають `grant_transition_invalid`. Додано `course_not_enrollable` і `enrollment_forbidden`. Доменні юніти зелені.

ДАЛІ: Фаза 2 — розширити CourseReader enrollment-фактами, postgres-читанням і оновити всі фейки порту.

Фаза 2: `CourseReader.CourseForEnrollment` повертає `status/free/instructor_id`; postgres читає рівно ці факти та мапить відсутній курс у `course_not_enrollable`. Оновлено фейки application і HTTP; `go test ./internal/orders/...` зелений.

ДАЛІ: Фаза 3 — замінити grant write-порт на CreateMembership/UpdateMembership і реалізувати транзакційні INSERT/UPDATE + audit у postgres.

Фаза 3: порт має `CreateMembership` і `UpdateMembership`; обидва postgres-методи пишуть membership та `grant_audit` в одній транзакції. Create мапить unique-конфлікт у `course_already_owned`; Update вимагає рівно один рядок, інакше `grant_invalid`; `ON CONFLICT` у цих методах не використано.

ДАЛІ: Фаза 4 — application use-cases і фейкові юніти: enroll/block/unblock/revoke/state та admin re-grant.

Фаза 4: додано `Enroll`, `Block`, `Unblock`, `Revoke`, `MembershipState`; усі спершу читають Membership. No-op: enroll active/blocked, block blocked, unblock active, повторний admin revoke — повертають чинний запис без write/audit. `GrantCourse`: none→Create, revoked→Reinstate(admin)+Update, active/blocked→`course_already_owned`. Application-юніти покривають eligibility, ролі й переходи.

ДАЛІ: Фаза 5 — дописати інтеграційні En01/En14/En15, re-enroll, eligibility, block/unblock/revoke та admin re-grant; Docker-прогін лишається Courses 1.

Фаза 5: додано postgres integration-сценарій `L2b enrollment membership writes`: En01, En14, En15, free revoked re-enroll, paid/draft/archived guards, block/unblock/forbidden, revoke/re-revoke та admin re-grant (усі з перевіркою єдиного рядка). `go test -tags integration ... -run '^$'` компілює сценарій; Docker не запускався за межами ролі.

Перевірки: `GOCACHE=/private/tmp/courses-l2b-gocache go build ./...`, `go vet ./internal/orders/...`, `go test ./internal/orders/...` — зелені; `git diff --check` — чисто. Перший build без `GOCACHE` не мав доступу до системного Go cache; повтор з заданим cache зелений. Generated/sqlc не чіпались, комітів немає.
