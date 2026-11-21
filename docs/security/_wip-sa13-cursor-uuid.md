# WIP — Courses 2 незалежне рев'ю: PR #83 SA13/D-F2 cursor UUID validation

## Handoff — МАРШРУТ ЗМІНЕНО (2026-10-29)

- **СТАТУС: я (Courses 2) stand-down для цього PR.** Courses13 повідомив, що власник авторизував рев'ю через `independent_subagent` (субагент-рев'ювер у сесії Courses13), бо моя сесія на обов'язковому handoff. Я не заперечив (режим задокументований, зріз малий). **Свіжу Courses 2-сесію під PR #83 відкривати НЕ треба — інакше дубль-рев'ю.**
- **Transfer state з мого боку:** `acknowledged → stand-down`. Reviewer-of-record = субагент Courses13. Вердикт дасть він.
- **Не верифіковано мною:** owner-авторизацію subagent-маршруту я сам не перевіряв — покладаюсь на заяву Courses13. Якщо власник ЦІЄЇ сесії скаже інакше → можу відкрити свіжу Courses 2-сесію (`review_mode=existing_session`) і виконати рев'ю сам за планом нижче.
- **Авторизація (якщо знадобиться):** стала — власник авторизував цю сесію як Courses 2 для всіх зрізів per-slice flow («approve»).
- **Чому handoff цієї сесії:** транскрипт ~4 МБ + стиснення. Draft ID: `6b9e11394d09c5ee350cead93cc140b2-1-45b80def4097`.

## Пакет

- PR **#83**, гілка `sa13-cursor-uuid`, commit `e73c255`, base `main @ 3f646d9`.
- Writer: Courses13 [c49267].
- **Спільне дерево — READ-ONLY:** `git show e73c255`, `git diff main...sa13-cursor-uuid`. НІКОЛИ checkout/reset/rm/clean на спільному дереві.

## Суть зміни

4 discussion-курсор-декодери не валідували, що `id` курсора — UUID:
- `decodeCursor` — repository.go
- `decodeQueueCursor` — moderation.go
- `decodeModerationQueueCursor` — review.go
- `decodeParticipantReportCursor` — participant_report.go

Крафчений курсор (валідний stamp, не-UUID id) доходив до `::uuid`-касту → Postgres `22P02` → `persistenceError` → `ErrDependencyUnavailable` (HTTP **503**). Має бути **400** (`ErrInvalidCursor`).

Фікс: single-source helper `cursorUUIDValid` (через `pgtype.UUID.Scan`, дзеркалить уже захищені orders-декодери); усі 4 відхиляють не-UUID як `ErrInvalidCursor`.

Доказ автора: integration-підтест `TestDiscussionPostgres/SA13-A` — RED до фіксу (усі 5 call-site = `dependency_unavailable`), GREEN після. gofmt/go vet чисто, `go test -race ./...` ok, discussion integration ok. check-generated/build/e2e автор scoped out «за впливом» (без зміни handler-сигнатур/маршрутів/форм/залежностей; happy-path курсори байт-ідентичні).

## План рев'ю (точки уваги)

1. **`pgtype.UUID.Scan` відхиляє ВСІ не-UUID форми**, що раніше давали 22P02 — не лише одну. Перевір порожній рядок, занадто короткий/довгий, валідний hex без дефісів (pgtype приймає? — звір із тим, що `::uuid`-каст у PG приймає/відхиляє), сміття. Якщо pgtype приймає форму, яку PG-каст відхилив би (або навпаки) — лишається 503-щілина або хибне 400 на валідному.
2. **Happy-path `encode*`-курсори лишились валідними** під `cursorUUIDValid` — звір формат, що `encode*` emit'ить, проти того, що `cursorUUIDValid` приймає (round-trip). Не зламати легітимну пагінацію.
3. **Немає незахищеного decode-шляху поза цими 4.** Grep усі decode*cursor / ::uuid-касти у discussion + суміжних модулях; orders — вже захищені (підтверди, не додано регрес). Будь-який інший курсор-декодер з id→::uuid без валідації = finding.
4. **Неваковість SA13-A:** підтест справді RED без фіксу (не tautology); 5 call-site реально покриті; 400/`ErrInvalidCursor` мапиться на HTTP 400 на межі (звір проти реального handler-рендера кодів помилок — пастка «sentinels vs real handler»).
5. **Стандарт shared/persisted-state:** зміна в декодерах пагінації читається багатьма; автор scoped out e2e «за впливом». Оціни, чи це обґрунтовано (чистий decode-guard, happy-path байт-ідентичний) чи треба повний e2e. Якщо сигнатури/маршрути/форми не чіпані — targeted integration достатньо; інакше ескалюй.

## Процедура

1. `it2 session archive-clippings --session=$TERM_SESSION_ID` (один раз перед новими findings).
2. `git diff main...sa13-cursor-uuid` + `git show e73c255` — прочитати повний diff.
3. Прогнати цільові тести read-only (не мутуючи дерево): discussion integration + SA13-A, за потреби `-count=1`.
4. Findings, що потребують дії → `it2 session add-clipping …`. Позитивні — без clipping.
5. Вердикт PASS/FAIL з доказом → `SendMessage` піру Courses13 [c49267]. Оновити transfer state тут → `verdict_received`.
6. Короткий звіт користувачу українською.

## Пам'ять (релевантні ноти)
- `[[shared-working-tree-with-peer]]` — read-only через git show.
- `[[check-sentinels-against-real-handler]]` — межові коди звіряй проти реального handler-виводу.
- `[[integration-tag-skipped-by-go-test-and-e2e]]` — зміна злитої поведінки гейтиться повним integration, не лише таргетним.
- `[[verify-what-a-count-counts]]`, `[[assert-only-what-compiled]]`.
