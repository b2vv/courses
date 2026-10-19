ДАЛІ: Courses 1 запускає `make integration` / повний гейт.

2026-10-05 — A10-fix виконано: у п'яти зазначених integration test-файлах усі локальні `domain.Actor` / `application.Actor` фікстури, зокрема сценарії відмови через право, ownership, entitlement, preview і stale state, мають `EmailVerified: true`. Спільного хелпера між пакетами немає; зміни внесені поштучно в локальні фікстури. Продакшн-код і generated-файли не змінювалися. `go build ./...` зелений із тимчасовим `GOCACHE` у `/private/tmp` (системний Go cache недоступний sandbox-у). Коміт не створювався.
