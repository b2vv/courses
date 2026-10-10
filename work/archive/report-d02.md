
## 2026-09-28 — старт

- Почато виконання `work/brief-d02.md`; коміти не створюються.

ДАЛІ: звірити спеки/план D02, домен discussion, міграції та шаблон transcript/postgres перед проєктуванням портів.

ФАКТ: D01 уже містить домен `discussion` і крос-модульні access-тести; план §2/§3 підтвердив ручний pgx SQL, фільтрований selector `ready + не evicted`, keyset `(created_at, id)` і cap через `FOR UPDATE`.
РІШЕННЯ: порти лишаються доменними; `LessonFacts` повертає курс і вибраний asset/duration, `List` повертає корені з усіма відповідями, а `InsertReply` серіалізує root-lock → count не прибраних → `ValidateReplyCount` → insert в одній транзакції.
ФАКТ: додано `00009_discussion.sql`, `discussion/ports`, ручний `discussion/adapters/postgres`, named `ErrInvalidCursor`, коментарі selector у дозволених модулях і B-22 у беклозі.
ДАЛІ: форматувати й прогнати unit, vet, build та integration-tag compile; окремо спробувати Docker integration і зафіксувати її фактичний стан.

ФАКТ: `gofmt -l backend` порожній; `GOCACHE=/private/tmp/courses-go-cache go vet ./...` — PASS; `go vet -tags integration ./internal/discussion/...` — PASS; `go build ./...` — PASS; `go test ./internal/discussion/...` — PASS; `git diff --check` — PASS.
ФАКТ: integration-тест компілюється під `-tags integration`, але запуск `go test -tags integration ./internal/discussion/adapters/postgres -run TestDiscussionPostgres -count=1` заблоковано середовищем: testcontainers не отримав доступ до Docker socket (`permission denied`). Це не позначаю як PASS.
РІШЕННЯ: K09a має серіалізований доказ `200 → ErrThreadFull` і selector асертить сам asset (`AssetID == ""`, `Generation == 0`) для evicted ready N + processing N+1. Детермінований двоз'єднальний wait у цій пісочниці не перевірений, тому механізм `FOR UPDATE` залишено для рев'ю, як дозволяє план §2.
ФАКТ: міграція дозволяє `body IS NULL` лише для незворотного removal, а активний body все ще має DB CHECK 1..2000 і trim; `Remove` ставить NULL + `removed_at`. Список приховує прибраний root без живих replies і зберігає root із replies без body.
ДАЛІ: фінально звірити статус робочого дерева й відсутність комітів; у відповіді передати стислий результат і blocked integration.

ФАКТ: фінальна перевірка показала лише очікувані незакомічені зміни D02 плюс попередні користувацькі зміни (`docs/discussion/plan.md`, PNG-скриншоти, `work/tasks.md`); останній commit лишився `2b8c72e`. Комітів не створено.

ФАКТ: після зовнішнього запуску Courses виправлено лише фікстуру `repository_integration_test.go`: обидва `users.password_hash` замінено на валідний Argon2id-хеш із transcript integration-тесту; production-код не змінювався.
ФАКТ: `gofmt` і `GOCACHE=/private/tmp/courses-go-cache go vet -tags integration ./internal/discussion/...` — PASS. Docker integration не запускався за запитом; його повторно запускає Courses.

ФАКТ: K09a доповнено детермінованим two-connection доказом: tx1 блокує root через `SELECT ... FOR UPDATE`, адаптерська горутина має не завершитися за 300 мс, tx1 вставляє 200-ту відповідь і commit, після чого горутина мусить повернути `ErrThreadFull` у межах 5 секунд. Серіалізований `200 → ErrThreadFull` кейс залишено.
РІШЕННЯ: тест навмисно використовує окремо acquired connection для tx1 і repository pool для горутини; це перевіряє саме міжз'єднальний row lock, а не внутрішню послідовність одного connection.
ФАКТ: після цієї правки `gofmt` і `go vet -tags integration ./internal/discussion/...` — PASS. Integration не запускав.

ФАКТ: `00009_discussion.sql` тепер має іменований `discussion_comments_body_removal_check`: `(removed_at IS NULL) = (body IS NOT NULL)`, тому живий NULL-body і removed non-empty-body відхиляються на рівні БД.
ФАКТ: K08 доповнено двома негативними вставками з перевіркою `*pgconn.PgError.Code == "23514"`; існуючі K09 перевіряють `Remove` як `body=NULL + removed_at` і список як заглушку без body.
ФАКТ: `gofmt` і `GOCACHE=/private/tmp/courses-go-cache go vet -tags integration ./internal/discussion/...` — PASS. Integration не запускав; комітів не створено.

ФАКТ: за рев’ю Courses 2 B-22 тепер перелічує всі grep-знайдені filtered-selector копії в transcript, playback, media і discussion з файлами/рядками та датою звірки; course sorting-форму відокремлено як свідомо іншу.
ФАКТ: K09a отримав позитивний selector fixture: невитіснений ready generation 1 із duration 45, витіснений ready generation 2 і processing generation 3; асерт перевіряє asset ID, generation і duration. Негативний evicted-ready + processing кейс залишено.
ФАКТ: K08 тепер створює root+reply на окремому уроці, видаляє `course_lessons` і перевіряє нуль discussion-рядків, доводячи `ON DELETE CASCADE`, а не лише наявність FK.
ФАКТ: `InsertRoot`/`InsertReply` отримують справжні `created_at` через `INSERT ... RETURNING id, created_at`; Go `time.Now()` для цього більше не використовується.
ФАКТ: `gofmt`, `GOCACHE=/private/tmp/courses-go-cache go vet ./...` і `go vet -tags integration ./internal/discussion/...` — PASS. Integration не запускав; комітів не створено.
