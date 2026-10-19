ДАЛІ: TDD — додати червоні unit-тести публічного шва `orders/domain.Grant` (конструктор, переходи, self-enroll), потім реалізувати мінімум для green.

2026-10-06 — L1 стартовано. Перевірено `work/CTO-RESEARCH.md`, `docs/enrollment/plan.md` §3–5 і бриф. Робочий файл, зазначений у брифі, був відсутній, тому створено його як append-only журнал. TDD-шви: `NewGrant`, `Grant.Block`, `Grant.Unblock`, `Grant.Revoke`, `SelfEnroll`; це публічна чиста доменна межа, потрібна для T1. sqlc/generated не чіпати; коміту не робити.

2026-10-06 — RED: `GOCACHE=/private/tmp/courses-l1-gocache go test ./internal/orders/domain/...` падає на відсутніх `GrantSelf`, `GrantBlocked`, переходах і `SelfEnroll` (локальний cache потрібен через sandbox). ДАЛІ: мінімальна доменна реалізація та green.

2026-10-06 — GREEN-код додано: джерело `self`, стан `blocked`, `ErrGrantTransitionInvalid`, чисті переходи `Block`/`Unblock`/`Revoke` і `SelfEnroll`; `NewGrant` зберігає purchase↔order інваріант. ДАЛІ: прогнати domain tests, потім перевірити seed-пари й написати 00015.

2026-10-06 — Domain tests GREEN (із sandbox-local `GOCACHE`), gofmt чистий. Таблиця переходів тепер покриває всі 9 комбінацій active/blocked/revoked × block/unblock/revoke; валідаційна таблиця — self/blocked та purchase-order інваріант. ДАЛІ: перевірити міграцію і seed.

2026-10-06 — Додано `00015_enrollment.sql`: source/state CHECK розширено, partial active unique замінено на повний `course_grants_one_per_user_course`, додано `participant_reports` із self-report CHECK та partial open-dedup, backfill union `lesson_progress ∪ discussion_comments ∪ discussion_reviews` лише для published free курсів, і симетричний DOWN з delete self та blocked→revoked перед звуженням. Seed: `student@courseforge.local` (`400…001`) отримує self/active на три published free demo-курси (`100…001` Go, `…002` React, `…003` System Design); draft/archive навмисно не включені. Backfill покриває наявні E2E-патерни progress, comment і review; точних старих seed-пар активності немає. ДАЛІ: статична перевірка SQL/збірка/vet і diff-аудит.

2026-10-06 — Фінальна локальна перевірка PASS: `gofmt -d` порожній; `git diff --check`; `GOCACHE=/private/tmp/courses-l1-gocache go test ./internal/orders/domain/...`; `go vet ./internal/orders/...`; `go build ./...`. Міграцію не застосовано: у sandbox немає `goose`/`psql` (і застосування на чистій БД належить Courses 1 за брифом). sqlc/generated не запускалися й не змінені; коміту немає. ДАЛІ: Courses 1 застосовує 00015 up/down на чистій БД та продовжує T2.
