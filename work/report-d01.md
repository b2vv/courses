## 2026-09-28 — старт

- Розпочато виконання `work/brief-d01.md`.
- Коміти не створюються.
- Наступний крок: прочитати репозиторний брифінг, brief і інструкції виконання.

## 2026-09-28 — прочитано scope

- Прочитано `work/CTO-RESEARCH.md`, `work/brief-d01.md`, `docs/discussion/spec.md`, `docs/discussion/plan.md` і `work/tasks.md`.
- Scope обмежено доменом `discussion`, його unit-тестами та двома architecture-тестами; міграції, SQL, HTTP і фронтенд не змінюються.
- Далі: звірити типи/шаблони `transcript` і поточні крос-модульні тести.

## 2026-09-28 — доменний тестовий зріз

- Додано тести K01–K04 і K06 для Unicode-тіла, мітки в мс, плоских відповідей, cap та прав доступу.
- Додано доменні типи й правила: `Actor`, `Course`, `Comment`, іменовані помилки, `NewComment`, `ValidateReplyCount`, `CanBeEditedBy`, `CanBeRemovedBy`.
- Перша команда форматування була запущена з неправильним відносним шляхом і нічого не змінила; повторюю з кореня репозиторію.
- Далі: `gofmt` і `go test ./internal/discussion/...`, виправлення знайдених проблем.

## 2026-09-28 — цільові гейти

- `gofmt -l backend` — порожній результат.
- `GOCACHE=/private/tmp/courses-go-cache go vet ./...` — PASS.
- `go test ./internal/discussion/... ./platform/architecture/...` — PASS (`discussion/domain`, `platform/architecture`).
- Далі: виконати контроль невакуумності архітектурного тесту через тимчасово зламаний mapping і повернути код.

## 2026-09-28 — контроль mapping і фінальне уточнення

- Тимчасова підміна `discussion.Status` на `Published` дала очікуваний FAIL на `stranger_draft` і `stranger_archived`; це підтвердило, що `viaDiscussion` не є вакуумною гілкою. Mapping повернуто.
- Виправлено доменну помилку: `NewComment` тепер повертає `ErrParentInvalid`, якщо parent передано без `ParentID`.
- Далі: повторити `gofmt`, `vet`, цільові тести, перевірити diff і статус робочого дерева.

## 2026-09-28 — фінальна перевірка

- Попередній паралельний виклик `gofmt -l backend` з теки `backend` був некоректним шляхом; цільові unit/architecture-тести при цьому пройшли. Повторюю форматування з кореня та `vet` із `backend` окремо.
- Незалежні зміни `docs/discussion/plan.md` і PNG-скриншоти залишені без втручання; комітів не створено.

## 2026-09-28 — review і завершення

- Проведено standards/spec review від `HEAD` по робочому diff: production-домен імпортує лише stdlib і `courses/internal/contracts/access`; architecture-тести пінять усі 4 mapping-и та межу мітки; scope не вийшов за D01.
- `git diff --check` — PASS.
- Підсумок: D01 виконано повністю — домен, K01–K04/K06 unit-тести, K05 AST-лічильник, K05a `viaDiscussion`, position-bound гілка та контроль навмисного FAIL.
- Комітів немає. Незв’язані зміни, що були до роботи (`docs/discussion/plan.md`, два PNG), не змінювалися.

## 2026-09-28 — уточнення запуску

- Комбінований фінальний виклик з кореня зупинився на `go vet` через те, що корінь репо не є Go-модулем; це помилка робочої теки команди, не коду. Повторюю `vet`/тести з `backend`, а diff-перевірку — з кореня.

## 2026-09-28 — завершено

- Коректний фінальний прогін: `gofmt -l backend` — порожньо; `GOCACHE=/private/tmp/courses-go-cache go vet ./...` з `backend` — PASS; `go test ./internal/discussion/... ./platform/architecture/...` — PASS; `git diff --check` — PASS.
- Робоче дерево містить лише D01-файли/тести та попередні незв’язані зміни; комітів не створено.
