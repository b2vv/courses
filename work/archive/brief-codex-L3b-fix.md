# Codex brief — L3b-fix: isolate the new membership-state test fixtures

Репо `courses`, гілка `feat/enrollment`. Твій L3b (конверсія 7 читачів) — реалізація КОРЕКТНА, проблема лише в
ТЕСТ-ФІКСТУРАХ: нові membership-підтести мутують СПІЛЬНІ фікстури, які читають сусідні тести → два падіння на
повному `make integration`.

РОБОЧИЙ ФАЙЛ: /Users/strelia/projects/courses/work/_wip-L3b-codex.md — append; перший рядок `ДАЛІ:`.

## Падіння (обидва — забруднення спільного стану, НЕ баг реалізації)
1. **discussion** `adapters/postgres/repository_integration_test.go` — твій підтест
   «L3b lesson and comment facts read every membership state» вставляє коментар `'membership facts'` у СПІЛЬНИЙ
   `discussionLesson` (course d1) і лишає grant на `discussionStudent`. Пізніший `K09_roots_cursor_replies_and_removed_shape`
   рахує корені на тому ж уроці → бачить зайвий корінь → FAIL (repository_integration_test.go:798).
2. **playback** `adapters/postgres/access_integration_test.go` — твій підтест
   «L3b sessions and media bytes read every membership state» ганяє grant `viewerID`/`courseID` через
   active→blocked→**revoked** і лишає його revoked. Пізніше той самий `viewerID`/`courseID` читається й очікує
   доступ → бачить Membership:revoked, Entitled:false → FAIL (access_integration_test.go:246).

## Правило (AGENTS): хто додає інстанс глобально-ліченого — ізолює його
Перероби ОБИДВА нові membership-підтести так, щоб вони працювали на ВЛАСНИХ, виділених фікстурах, що їх НЕ читає
жоден інший тест:
- Створи в кожному підтесті окремий course (+ потрібний ланцюг: module/lesson, а для playback — video_asset;
  для discussion — свій lesson+comment) і окремого user-а; grant/коментар став саме туди.
- НЕ чіпай `viewerID`/`courseID`/`videoID`/`lessonID` (playback) і `discussionLesson`/`discussionStudent`/course d1
  (discussion), якими користуються сусідні підтести.
- Переконайся, що твої assert-и (active→Entitled=true/Membership=active; blocked/revoked→Entitled=false,
  Membership відповідний) лишаються — просто на ізольованих фікстурах.
- (Якщо ізоляція для якогось читача надто важка — як запасний варіант прибери за собою в кінці підтесту:
  DELETE створених course_grants/коментарів, щоб спільний стан повернувся. Але ізоляція надійніша за cleanup,
  бо не залежить від порядку сусідів. Обирай ізоляцію, якщо можеш.)

Інші читачі (course learning, transcript) пройшли — подивись, як там зроблено ізоляцію, і наслідуй, якщо зручно.

## Правила
- Чіпай ЛИШЕ тест-файли (фікстури підтестів). Реалізацію читачів (SQL/мапінг) НЕ міняй — вона правильна.
- generated/sqlc не чіпай. НЕ комість. gofmt чистий; `go build ./...` і доменні юніти зелені.
- Повний Docker-integration проганяє Courses 1.

## Критерій готовності
- Обидва нові підтести на ізольованих фікстурах; K09 і playback «first playback» більше не бачать чужого стану.
- У робочому файлі: що саме ізольовано (які нові course/lesson/user/video id), підтвердження що спільні
  `viewerID`/`discussionLesson`/d1 не мутуються.
