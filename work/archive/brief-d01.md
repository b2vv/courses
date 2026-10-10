# Бриф D01 — домен `discussion` і три крос-модульні тести

РОБОЧИЙ ФАЙЛ: work/report-d01.md
Дописуй ПІСЛЯ КОЖНОГО КРОКУ (append, не rewrite). Перший запис — до першого інструменту:
рядок `ДАЛІ:` з тим, що збираєшся перевіряти. Далі — ФАКТ / РІШЕННЯ / ДАЛІ / ПИТАННЯ.
Фінальна відповідь = стислий переказ файлу.

## Контекст

Phase 7b — гілки бесід. Спека: `docs/discussion/spec.md` (D1–D3, D7, рядки K01–K06, K05a).
План: `docs/discussion/plan.md` (§1 D01, §2 перший пункт, §3 X2, X4). Прочитай обидва перед кодом.
Цей слайс — **лише домен і архітектурні тести**. Жодної міграції, SQL, HTTP, фронту.

## Що зробити

1. **Пакет `backend/internal/discussion/domain`** — лише stdlib і `courses/internal/contracts/access`
   (інші модулі імпортувати не можна; це пінить `platform/architecture/boundaries_test.go`).
   Шаблон мапінгу — `backend/internal/transcript/domain/access.go`: власні `Actor`, `Status`, `Course`.
   - `Course.CanBeViewedBy(actor Actor) bool` — тіло рівно один
     `return access.CanView(access.Subject{…}, access.Course{…})`, як у transcript (це пінить AST-тест).
   - `Comment` з полями за спекою D2 (`ID`, `LessonID`, `AuthorID`, `ParentID`, `Body`,
     `PositionMs`, `CreatedAt`, `EditedAt`, `RemovedAt` — nullable там, де спека каже null).
   - Тіло: trim, 1..2000 **символів Unicode** (руни, не байти) → інакше `ErrBodyInvalid`.
   - Мітка: лише в кореня, необов'язкова. Якщо задана — потрібна тривалість ready-ассета
     (`durationSeconds *float64`, nil = невідома) і `0 ≤ positionMs ≤ тривалість`, **включно**.
     Відповідь із міткою, відсутня тривалість, від'ємна чи понад тривалість → `ErrPositionInvalid`.
     ⚠️ Одиниці: тривалість у секундах `float64`, мітка — ціле мс. Порівнюй без зрізання
     тривалості до цілого (`float64(positionMs) <= durationSeconds*1000`): мітка рівно на кінці відео
     мусить проходити.
   - Відповідь: батько — корінь того ж уроку, не прибраний → інакше `ErrParentInvalid`.
     Cap: функція, що приймає кількість **не прибраних** відповідей кореня і повертає
     `ErrThreadFull`, якщо вона вже 200 (сама атомарність — у D02, не тут; константу 200 тримає домен).
   - `Comment.CanBeEditedBy(actor) bool` — лише автор (порожній `UserID` — ніколи).
   - `Comment.CanBeRemovedBy(actor Actor, course Course) bool` — автор ∨ адмін ∨ власник курсу
     (порожній `UserID` не збігається з порожнім власником — див. захист у `access.CanView`).
   - Помилки — іменовані `var Err… = errors.New(…)`, щоб D03 мапив їх у коди спеки D3.
2. **Unit-тести домену** (K01–K04, K06): межі тіла (0, 1, 2000, 2001 символ, пробіли, багатобайтні
   символи); мітка (корінь/відповідь, nil тривалість, 0, рівно кінець, кінець+1 мс, дробова тривалість);
   батько (невідомий/інший урок/не корінь/прибраний); cap (199 → ок, 200 → `ErrThreadFull`);
   `CanBeViewedBy` над **усією** `accesstest.TruthTable` (як `transcript`); правила зміни й прибирання
   на ролях автор / сторонній студент / адмін / власник курсу / анонім.
3. **Три архітектурні тести — у цьому ж слайсі** (спека D7, план §4):
   - `backend/platform/architecture/access_mapping_ast_test.go:19` — `expectedMethods` 3 → 4.
   - `backend/platform/architecture/shared_rules_test.go` `TestTheAccessPredicateAgreesAcrossModules` —
     додай гілку `viaDiscussion` поряд із `viaCourse`/`viaPlayback`/`viaTranscript` (рядки 37–54), у
     умову й у текст `t.Fatalf`.
   - Той самий файл, `TestThePositionBoundAgreesAcrossModules` — гілка `discussion` (план X2): для
     кейсів із **відомою** тривалістю мітка `positionMs = позиція × 1000` приймається тоді й лише тоді,
     коли кламп `course` лишає позицію незмінною (`viaCourse == test.position`); для кейсу з `nil`
     тривалістю — відхиляється. Тест має падати, якщо межі розійдуться хоч на одному кейсі.

## Гейти в пісочниці

`gofmt -l backend` (порожньо); `GOCACHE=/private/tmp/courses-go-cache go vet ./...` і
`go test ./internal/discussion/... ./platform/architecture/...` (з того ж каталогу `backend`).
Контроль, що тести не вакуумні: тимчасово зламай мапінг (`Published: true`) — `shared_rules_test`
мусить почервоніти; поверни. Результат контролю — у звіт. Docker/integration/e2e не ганяй.

## Межі

Не комітити. Не чіпай інші модулі, крім двох тестових файлів у `platform/architecture`. Якщо
спека чи план двозначні — стоп, ПИТАННЯ у звіт, не вгадуй. Якщо знайдеш неточність у спеці, що
призведе до бага, — аргументуй у звіті.
