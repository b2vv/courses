# S0.1 Run 1 — platform/openapigen core + course catalog fidelity (WORK FILE)

РОБОЧИЙ ФАЙЛ для Codex. Дописуй ПІСЛЯ КОЖНОГО КРОКУ (append, не rewrite). Перший запис — до першого
інструменту: рядок `ДАЛІ:` з тим, що збираєшся перевіряти. Далі — ФАКТ / РІШЕННЯ / ДАЛІ / ПИТАННЯ.
Фінальна відповідь = стислий переказ цього файлу.

Base: branch `s0.0-spike` @ `3c193ab` (base main @ 8021e83). Писатимеш на цій гілці.

## Журнал

ДАЛІ: звірити branch/HEAD, план S0 та catalog DTO/spec перед кодом. Цей перший запис внесено після початкового читання, а не перед першим інструментом, як вимагала шапка; порядок журналу не видаю за дотриманий.

ФАКТ (2026-10-27, Codex): root `/Users/strelia/projects/courses`, branch `s0.0-spike`, HEAD `3c193abd79d6386cd8954536c30da7ac0318411c`. На старті єдиний untracked файл — цей work file; його шапку збережено. Інших гілок не торкався.

ФАКТ: прочитано S0 plan, catalog DTOs, відповідні paths/schemas ручної OpenAPI та обидва spike-файли. `courseDetailDTO` вбудовує `courseDTO`; catalog має три slice-поля (`programModuleDTO.Lessons`, `courseDetailDTO.Program`, `pageDTO.Items`). План §11.2 вимагає явного nil/non-nil рішення для кожного slice, тому фраза brief «none needed for catalog» неточна: у contribution потрібні явні non-null рішення. Це можна реалізувати за пріоритетним планом без розширення scope.

РІШЕННЯ: pre-code BLOCKED через суперечність нижче. Код, overlay та generated файл не створено. Не запускав gofmt/vet/build/test: усі чотири gates NOT_RUN, результатів PASS немає. Commit/merge/push не виконував. Review route за планом — наявна сесія Courses 2, передача через власника; незалежне рев'ю цього run NOT_TESTED, повідомлень іншим сесіям не надсилав.

## ПИТАННЯ

1. Чи є §16.1/§16.4 умовами завершення всього S0.1, які дозволено виконати в наступних runs, а не умовами початку Run 1? Brief прямо забороняє request-body DTOs і cross-module error-code parser у Run 1, але надає плану пріоритет. План `docs/security/s0-openapi-gen-plan.md:427` каже «must land before S0.1»; рядки 440–442 вимагають parser + coverage як «S0.1 entry»; 470–472 вимагають fidelity ≥1 request DTO; 479–488 повторюють entry conditions. Самостійно перенести ці передумови на кінець фази означало б змінити нормативний контракт. За STOP-інструкцією brief зупинився до коду. Рекомендоване уточнення: явно дозволити catalog-only Run 1, зберігши §16.1/§16.2/§16.4 як обов'язкові умови завершення S0.1 перед cutover.

ДАЛІ: після відповіді власника продовжити обов'язкове читання джерела openapi3gen та pre-code checkpoint, реалізувати вузький catalog slice, виконати чотири gates у заданому порядку й зробити дозволені logical commits. Жодних fidelity assertions не послаблено.

ФАКТ: native identity підтверджена host-подією для Codex session `01a17e51-1197-7702-a342-ed3776c781d8`; точний courses policy row знайдено. Спроба checkpoint через зареєстрований dispatcher не виконана: sandbox відмовив `mktemp` у configured receipts parent (`Operation not permitted`), dispatcher повернув `BLOCKED required handoff`. Зовнішній checkpoint не заявляється; стан збережено в цьому work file. Це filesystem sandbox denial, не auto-review rejection.

ФАКТ (продовження, 2026-10-27): власник зняв STOP; §16.7 підтверджує completion-gate staging. PRE-CODE READY: catalog responses only, чинний full S0 plan використовується без другого pipeline; writer Codex на s0.0-spike, review delivery через власника до Courses 2. Прочитано openapi3gen v0.149.0 (generator/type/field source), architecture rule, go.mod. Нових залежностей не потрібно: YAML encoder github.com/oasdiff/yaml v0.1.1 вже pinned як dependency kin-openapi. External checkpoint веде власник.

РІШЕННЯ (Codex, 2026-10-27): використати explicit NameMap для реальних імен courseDTO→Course тощо (імена Go буквально НЕ збігаються з OpenAPI). time.Time реалізує json.Marshaler, тому contribution явно реєструє date-time override. Library component emission скидає nullable на canonical component; генератор додає allOf wrapper на pointer-полі. ErrorResponse/ErrorDetail відображаються з реальних platform DTO; enum кодів тимчасово overlay-carried, parser і його drift-proof claim лишаються наступним runs. Перевірка — unit tests + catalog fidelity.

ФАКТ: перший focused run GREEN: core fail-closed tests + TestCatalogOpenAPIFidelity. Fidelity порівнює обидва paths, повне транзитивне замикання schemas/responses/headers і окремий Rating; нормалізує лише порядок required як множини, жодні поля не ігнорує. Дві генерації byte-identical; parallel artifact записаний через UPDATE_OPENAPI_GENERATED=1. Ручний openapi.yaml незмінений. Перший compile спіймав локальну alias-помилку yaml3 (package має ім'я yaml); виправлено до зеленого прогону. Це focused перевірка, не фінальний ordered gate.

РІШЕННЯ: overlay merge має null deletion; `$ref: null` розгортає згенероване посилання перед metadata merge, зберігаючи відображені поля/required. Так досягається inline Rating/ErrorDetail без ручної копії DTO структури. ErrorDetail component після inline видаляється; NullablePrice явно оголошений contribution як schema override wrapper. Unknown YAML shapes/keywords, orphan paths/components/properties, dangling refs відхиляються. Requests/non-JSON API не додаються спекулятивно у Run 1; recursion поки fail-closed через ThrowErrorOnCycle (document surface наступних runs).

ФАКТ: другий focused run GREEN після перевірки додаткових core edge cases: pointer-receiver Marshaler, scalar override для struct, untagged JSON fields/omitempty/ignored fields, nullable slice true, malformed/empty/multi-document/duplicate-key YAML, unknown keyword. Додано catalog Layer-B test: real handler bytes (list, empty list, detail, missing, bad query) валідовано проти freshly generated doc. Runtime handlers не змінені. Spike-файли залишено: error-code spike ще не замінено parser-run, тож видаляти обидва зараз не потрібно.

ДАЛІ: фінальні gates строго gofmt -l . → go vet ./... → go build ./... → go test ./internal/platform/openapigen/... ./internal/course/...; GOCACHE=/tmp/courses-s01-go-cache через sandbox. Після них logical commits без push/merge.

ФАКТ: уточнення власника вже закомічене як `8e195f4` (поточна база Run 1 замість попередньої `3c193ab`). Фінальні four gates виконано ПОСЛІДОВНО з backend/, кожен EXIT=0. Для Go задано `GOCACHE=/tmp/courses-s01-go-cache`; інших test-selection/env overrides у фінальному gate немає.

```text
$ gofmt -l .
[stdout/stderr empty; exit 0]
$ go vet ./...
[stdout/stderr empty; exit 0]
$ go build ./...
[stdout/stderr empty; exit 0]
$ go test ./internal/platform/openapigen/... ./internal/course/...
ok  	courses/internal/platform/openapigen	1.298s
ok  	courses/internal/course/adapters/http	6.397s
ok  	courses/internal/course/adapters/postgres	1.387s
?   	courses/internal/course/adapters/postgres/generated	[no test files]
ok  	courses/internal/course/application	4.101s
ok  	courses/internal/course/domain	1.836s
?   	courses/internal/course/ports	[no test files]
```

ФАКТ: determinism — `TestGenerateTwiceIsByteIdentical` і `TestCatalogOpenAPIFidelity` входять у цей gate; focused verbose run також показав обидва PASS. Artifact SHA256 `b3323f7c68b77bd38b5fde544f2863f2381483eaf080b0cc6360cb3228866934`. Ручний spec SHA256 `48910909c4fa64f43fc70b8e977a8adabbc770b18f40afad16dd5de056b2a629`; git diff для openapi.yaml/go.mod/go.sum/Makefile порожній. Генератор не імпортує жодного business module. Повний make check/e2e та окремий integration-tag run не запускались і не заявляються — заданий gate цього run виконано буквально.

РЕЗУЛЬТАТ: реалізовано core generator, module-owned catalog contribution, curated overlay, parallel generated doc, fail-closed/determinism/fidelity/real-handler-response tests. Reconciled: component pointer nullability; exact Go↔schema names; CourseDetail narrowing; inline rating vs standalone Rating; NullablePrice wrapper; inline ErrorDetail. Жодної зміни ручного spec або послаблення fidelity. Run 1 implementation/gates PASS; незалежний Courses 2 review ще PENDING (власник передає пакет). S0.1 у цілому НЕ закрито: parser/error-code source coverage, request-layer decision/request-DTO fidelity, інші routes/modules і full coverage-gate залишаються наступним runs за §16.7. Validation keywords та тимчасовий error-code enum overlay-carried і НЕ drift-proof. External checkpoint веде власник; pending STOP питань після уточнення немає.

ФАКТ: core закомічено `bc49613` — `feat(openapigen): add deterministic response generator core (SA10/S0.1)` із заданими двома trailers. Sandbox спочатку відмовив запис .git/index.lock; повтор через approved escalation завершився EXIT=0. Другий logical commit включає catalog contribution/tests, overlay/generated artifact і цей work file. Push/merge/перемикання гілок не виконуються.
