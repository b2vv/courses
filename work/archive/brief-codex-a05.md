# Codex brief — A05: гейт верифікації (CanStream + мутаційні перевірки, крос-модульно)

Репо `courses`, гілка `feat/accounts-selfservice`. A01–A04 у дереві: `httpserver.Actor.EmailVerified` уже
несе факт (A03). План `docs/accounts/plan.md` §4 (B1!). Роль: реалізуєш A05; коміт/гейти — Courses 1.

РОБОЧИЙ ФАЙЛ: /Users/strelia/projects/courses/work/_wip-a05-codex.md
Append після КОЖНОГО кроку; перший рядок `ДАЛІ:`. Фінал = стислий переказ.

## КРИТИЧНО (B1 — не зламати архітектурні тести)
- `platform/architecture/access_mapping_ast_test.go` вимагає РІВНО N методів `CanBeViewedBy`, кожен — рівно
  один `return access.CanView(...)`. **НЕ перейменовуй і НЕ міняй жоден `CanBeViewedBy`.**
- `shared_rules_test.go` TruthTable не має виміру email_verified — **не додавай його туди**.
- Гейт verified роби ОКРЕМО (новий предикат + перевірки в use-case шарі), не всередині CanBeViewedBy.

## Що зробити

1. **`internal/contracts/access/access.go`**: додай поле `EmailVerified bool` у `Subject` (саме по собі CanView
   і TruthTable не ламає — перевірено). Додай **окремий чистий предикат**
   `CanStream(subject Subject, course Course) bool { return CanView(subject, course) && subject.EmailVerified }`.
   Юніт-тест CanStream у `access_test.go`/`accesstest` окремо від TruthTable (НЕ розширюй TruthTable).

2. **Мапінг Actor→Subject у КОЖНОМУ модулі, що будує `access.Subject`** (знайди всі місця, що читають
   `httpserver.ActorFromContext` і конструюють `access.Subject`): заповни `EmailVerified` з `actor.EmailVerified`.
   Модулі: playback, transcript, discussion, course, orders (де вони маплять). course-метадані лишаються на
   `CanView` (поле просто заповнюється, поведінка не змінюється).

3. **Застосуй `CanStream` у споживанні контенту уроку** (use-case шар, НЕ domain CanBeViewedBy):
   - **playback**: там, де авторизується відтворення/mint playback — додай вимогу `CanStream` (тобто ще й
     verified) поверх наявного CanView-гейта. Неверифікований → відмова з кодом `email_not_verified`.
   - **transcript**: читання транскрипту уроку — так само `CanStream`.

4. **Мутаційні гейти `email_not_verified`** (точні методи, E2):
   - orders: `StartPurchase` (service.go:~43) і `Checkout` (~:108).
   - discussion: `Publish` (коментар), `ReportComment`, `PutReview`, `ReportReview`.
   Перевірка `subject.EmailVerified` (або `access`-хелпер) на вході use-case; код помилки `email_not_verified`
   у стилі наявних доменних/застосункових помилок модуля. **DeleteReview НЕ гейтуй** (створення вже гейтоване).

5. **Не гейтуй**: каталог, сторінку курсу (опис/outline), читання коментарів/відгуків.

## Правила
- Конституція/ADR-013: `contracts/access` — лише чисті предикати, stdlib, без імпорту модулів. Кожен модуль
  мапить свій факт у Subject із Actor (факт постачає identity через Actor — не читай таблицю users у модулях).
- Архітектурні тести (`platform/architecture/...`) МАЮТЬ лишитися зеленими — прожени їх.
- TDD/юніт де є сенс (CanStream — юніт; мутаційні гейти — unit/contract у відповідному модулі).
- НЕ комість; НЕ чіпай generated; НЕ запускай sqlc. Час вичерпується — спершу тести вже-написаного, познач залишок.

## Критерій готовності
- `go build ./...`, `go vet`, `gofmt` чисті.
- `go test ./platform/architecture/...` ЗЕЛЕНІ (B1!).
- `go test ./internal/contracts/access/... ./internal/playback/... ./internal/transcript/... ./internal/orders/... ./internal/discussion/...`
  зелені (unit; integration — Courses 1 на хості).
- CanStream має юніт-тест; кожен названий мутаційний метод відмовляє неверифікованому з `email_not_verified`.
