# S0.1l — course module: learn course-detail route

Writer: Courses 1 (via Codex). Reviewer: Courses 2 (existing_session, свіжа courses-96).
Branch `s0.0-spike`. Base HEAD: `c5d4a3b`.

Частина 1/3 добору 6 course-роутів до S0.2 (l=learn course, m=document, n=admin CRUD).

ДАЛІ: запустити Codex за брифом (scratchpad s0.1l), далі host-verify → коміт → Courses 2.

## Scope
Додати `GET /api/v1/learn/courses/{slug}` до наявної course-contribution. Fidelity-ціль — frozen
hand: path getLearningCourse + схеми LearningCourse/LearningModule/LearningLesson.

## Факти (до коду)
- `learningCourseDTO` embed-ить `courseDTO` (promoted → inline-поля; як CourseDetail, що вже PASS),
  + Modules/ContinueLessonID/CanPurchase. Генератор уже inline-ить courseDTO-поля правильно (Course зелений).
- rating: генератор $ref-ить ratingDTO→"Rating"; overlay НУЛИТЬ `$ref: null` і inline-ить (як у Course/CourseDetail блоках).
- videoState: `learningLessonDTO.VideoState *string`, значення з `domain.LearningVideoState`
  (processing/ready/failed; VideoAbsent="" → nil). Нема ordered-set → треба single-source
  `domain.LearningVideoStates` + census; поле змінити на `*domain.LearningVideoState` щоб enum чіплявся.
- Кожна named-схема, що inline-ить Course-поля, ДУБЛЮЄ constraints у власному overlay-блоці
  (CourseDetail так робить). LearningCourse/Module/Lesson — власні overlay-блоки.
- course-модуль НЕ має route-AST-census (на відміну від media/clientlog); гейт = fidelity проти frozen
  + const-set censuses. Додати videoState census; розширити fidelity на новий path+closure.
- pointer-поле → nullable авто (publishedAt не в NullableFields, але nullable). slice → nullable
  якщо не поставити false (pageDTO.Items=false). → Modules/Lessons треба false.

## Журнал
- (до коду) Інваріанти+overlay-патерн зібрано. Бриф написано. Чекаю Codex.

ДАЛІ: реалізувати A–D, розширити fidelity/census, UPDATE→plain artifact test, потім дозволені Go-гейти.
ФАКТ: root <repo>; branch s0.0-spike; HEAD c5d4a3b4611736648b8b4671f301a97e45e69123; scoped code clean, сторонній docs WIP збережено.
РІШЕННЯ: 2026-10-28, Codex за owner brief: small slice у чинному S0; READY. Writer Courses 1/Codex; review existing Courses 2 (courses-96), delivery owner через ручну передачу; host suite не запускаю. Fidelity frozen + mutation probe і census перевірять scope.
ПИТАННЯ: немає.

ФАКТ: A–D додано: ordered video states, named pointer DTO, один learn route, три overlay-схеми; frozen/engine не змінені. Pinned release verify PASS (exit 0).
РІШЕННЯ: constraints/prose скопійовано з frozen; enum/nullable лишено Go.
ДАЛІ: fidelity mutation та bidirectional enum census, регенерація.
ПИТАННЯ: немає.

ФАКТ: fidelity охоплює learn path і transitive closure; delete-videoState probe + existing double-generation check. Census порівнює declared non-absent == set == contribution enum, guard проти empty/Absent/extra/duplicate, synthetic new-const probe.
РІШЕННЯ: розширено єдиний const_sets_test.go; нових тестових файлів немає.
ДАЛІ: gofmt, UPDATE та plain artifact, fidelity.
ПИТАННЯ: немає.

ФАКТ: gofmt touched files exit 0. Перший UPDATE test exit 1: sandbox відмовив доступ до стандартного ~/Library/Caches/go-build; запуск із GOCACHE=/tmp/courses-s01l-go-cache триває.
РІШЕННЯ: використовувати writable /tmp GOCACHE без зміни permissions чи repo config.
ДАЛІ: завершити artifact test та scoped gates.
ПИТАННЯ: немає.

ФАКТ: UPDATE_OPENAPI_GENERATED=1 go test -run TestOpenAPIArtifact ./internal/openapispec PASS exit 0 (2.375s); plain без UPDATE PASS exit 0 (9.197s), обидва з GOCACHE=/tmp/courses-s01l-go-cache. gofmt -l . clean exit 0. generated +194/-0; frozen/engine git diff --exit-code 0. Native checkpoint OK exit 0: 8e196952d7e7af76b969a51e1d9d56ce-1-c7c0d622c0b8 (перший sandbox attempt exit 1, approved retry успішний).
РІШЕННЯ: scope без відхилень.
ДАЛІ: завершити vet, build, targeted tests.
ПИТАННЯ: немає.

ФАКТ: GOCACHE=/tmp/courses-s01l-go-cache go vet ./... exit 0; go build ./... exit 0.
РІШЕННЯ: обидва гейти PASS, без послаблень.
ДАЛІ: targeted go test course/openapispec/openapigen.
ПИТАННЯ: немає.

ФАКТ: GOCACHE=/tmp/courses-s01l-go-cache go test ./internal/course/... ./internal/openapispec/... ./internal/platform/openapigen/... PASS exit 0: http 10.083s, postgres 3.337s, application 8.320s, domain 1.299s, openapispec 7.105s, openapigen 5.323s. Fidelity (learn path + closure + delete-videoState probe + two generations) і census пройдено. git diff --check exit 0.
РІШЕННЯ: S0.1l implementation завершено; зміни лише у 7 дозволених code/artifact файлах + цей append-only журнал. Frozen openapi.yaml, generate.go, overlay.go без diff; runtime не підключено. Повний go test ./..., integration/e2e/make check NOT_RUN за брифом; незалежний review NOT_TESTED, не підмінений author gates. Commit не створено.
ДАЛІ: owner host-verify → commit → existing Courses 2 (courses-96); наступні course routes поза scope.
ПИТАННЯ: немає; stop conditions не виникли.

Final scoped snapshot на HEAD c5d4a3b4611736648b8b4671f301a97e45e69123 (SHA-256):
- `backend/internal/course/domain/learning.go`: `8a61d2655f5f90f148ddde1956e562dc40a6a06375687d7175a960f55ed1b67a`
- `backend/internal/course/adapters/http/learning_handler.go`: `ad0218a466194131b66c9f8c8708ecf92dfd0acede97724772d64da5edc9a2a1`
- `backend/internal/course/adapters/http/openapi_contract.go`: `46e26600c8da2d731bfbe6b7246c44c70180c36dfdbf1efdca80492f4457352b`
- `backend/openapi/openapi.overlay.yaml`: `a5953dcf31b82cad70d5e23c02a9ec74f64c9bd294af78f7a1ec07c565ff039b`
- `backend/internal/course/adapters/http/openapi_contract_test.go`: `4bb23a1c02d8f8398f4bf41cff11d7970bff17a2788535f513b24b572d9f3284`
- `backend/internal/course/adapters/http/const_sets_test.go`: `551132b73652d891015d243dbd4bdb3ae26fe184d3b21a7b737ccf541c259cec`
- `backend/openapi/openapi.generated.yaml`: `3b9dd6c1e2cea5f5f6f240e6185bc6b20baa288bf060b0605077cf7ea9fd9620`

ФАКТ: final native checkpoint OK, exit 0: `8e196952d7e7af76b969a51e1d9d56ce-2-4abfcdfeb172`.

## Host-verify (Courses 1)
- ФАКТ: frozen openapi.yaml + generate.go + overlay.go byte-identical до HEAD (git diff --stat порожній).
- ФАКТ: TestOpenAPIArtifact plain PASS; gofmt -l clean; vet 0; build 0; повний `go test ./...` exit 0, 0 FAIL
  (course http/application/domain/postgres + openapispec + openapigen зелені). Log scratchpad/s0.1l-hostgate.log.
- ФАКТ: код-дифи = дизайн (domain set+Valid, VideoState named-type, route+NameMap+Enums+NullableFields, 3 overlay-блоки). generated +194 (LearningCourse/Module/Lesson + path).
- ФАКТ: закомічено `9db7d30` (7 файлів explicit-paths). Покриття 80/85.
- transfer=`verdict_received`: **PASS**, 0 findings (courses-96 [557618], field-by-field звірка 3 схем,
  forced -count=1 на TestOpenAPIArtifact, scope підтверджено). S0.1l ЗАКРИТО на гілці; у main не мерджу.
  maxLength↔domain clipping — системний кандидат S0.2 (не новий тут), не блокер.
- Далі: S0.1m (document) розблоковано, стартую.
