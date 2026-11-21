ДАЛІ: host full-suite verification та незалежне Courses 2 review; S0.1m writer scope завершено.

# S0.1m — course module: lesson document routes (part 2/3)

Writer: Courses 1 (via Codex). Reviewer: Courses 2 (courses-96 [557618]).
Branch `s0.0-spike`. Base HEAD: `9db7d30` (after S0.1l PASS).

4 ops: GET/PUT `/api/v1/admin/lessons/{lessonId}/document`, POST `.../document/publish`,
GET `/api/v1/learn/lessons/{lessonId}/document`. Найскладніший зріз: глибоке рекурсивне дерево
LessonDocument→Section→RichText→Block→Span, nullable-масиви, bodyless 200 publish.

ДАЛІ: Codex за брифом (scratchpad s0.1m) → host-verify → коміт → Courses 2.

## Факти (до коду)
- DTO embed-ять domain-типи прямо: lessonDocumentDraftResponse{Document domain.LessonDocument,...},
  getPublished WriteJSON-ить domain.LessonDocument напряму. Генератор рефлексить дерево.
- **domain.Block МАЄ custom MarshalJSON** (H2/SA02: ordered лише для list). АЛЕ hand LessonBlock =
  лінивий union (ordered optional boolean) = base-struct reflection співпадає → ймовірно БЕЗ Override.
  Якщо generated LessonBlock ≠ hand → Override[domain.Block]. (course має TestS0SpikeOverrideCensus.)
- Enums: SectionType [video,text,image,row,task] + Mark [bold,italic] — named-type поля → Enums map +
  нові ordered sets `SectionTypes`/`Marks` + census. BlockType [paragraph,heading,list] вже у Enums.
  locale[uk,en]/schemaVersion[1]/contentType[image/jpeg,png,webp] — plain-поля → overlay enum
  (не single-sourced; named limitation, клас як maxLength S0.1k).
- content *RichText nullable → hand inline {object,nullable,allOf:[LessonRichText]}; патерн NullablePrice
  ($ref:null + allOf+nullable в overlay). sections/blocks nullable:true (NullableFields true);
  columns НЕ nullable (false); items[][] inner nullable — звірити.
- publish 200 bodyless → Response{Empty:true} (§16.18/f-2 Empty-маркер, будь-який статус).
- Усі error-компоненти вже існують (BadRequest/Unauthenticated/LearningForbidden/NotFound/InternalError/
  Unavailable/LessonDocumentConflict) — лише $ref із contribution, overlay не чіпати.

## Журнал
- (до коду) Повний target зібрано (4 path-блоки + 9 схем + domain enums). Бриф написано. Чекаю Codex.

ФАКТ (2026-10-28, Codex): root <repo>, s0.0-spike @9db7d30 підтверджено. Relevant code clean; сторонні docs/work WIP збережено. Frozen 1388–1500 та 2130–2222 прочитано; handler DTO/decode/marshal, contribution, fidelity/census і generator guard перевірено файловим fallback (graph непропорційний). Pin 17a9527 verify_release exit 0.
РІШЕННЯ: small, READY; writer Codex, reviewer Courses 2 після host verification. Нема нового API/runtime або підсистеми; acceptance — рівно 4 ops + 9-schema closure дорівнюють frozen, deterministic generation, enum census. Додаткові plan lenses N/A для механічного зрізу затвердженого S0. Питань щодо scope нема.
ФАКТ: generate.go:468 відхиляє Block json.Marshaler без Overrides незалежно від збігу struct tags. LessonDocumentConflict відсутній у поточному overlay (припущення брифу про наявність не підтвердилось).
РІШЕННЯ (Codex, 2026-10-28): explicit frozen Block override у contribution, enum із BlockTypes; перевірка — fidelity + marshal guard. Conflict metadata можна подати через дозволений Components без error overlays. Runtime bytes не змінюються.
ДАЛІ: внести bindings, enum sets, metadata й тести.
ПИТАННЯ: немає.

ФАКТ: внесено 4 bindings, enum sets/census, frozen metadata та nested fidelity probe. Save request expectedUpdatedAt отримав omitempty: лише request DTO, decode/response bytes незмінні.
РІШЕННЯ: уточнення попереднього запису — Components.Responses генератор не читає; metadata LessonDocumentConflict додано як component overlay за наявним патерном, бо у frozen він є, а в overlay його не було. На операціях error responses залишаються виключно contribution $refs. Section switch лишився для type-specific semantics; membership gate тепер SectionTypes.
ДАЛІ: regenerate, звірити firstDifference та nullable shapes.
ПИТАННЯ: немає.

ФАКТ (STOP, 2026-10-28): UPDATE_OPENAPI_GENERATED=1 go test -run TestOpenAPIArtifact ./internal/openapispec, GOCACHE=/tmp/courses-s01m-go-cache — EXIT_CODE=1 (повторний лог /tmp/courses-s01m-generation.log). Точна помилка:
```
artifact_test.go:32: reflect domain.LessonDocument: detected cycle
```
firstDifference НЕ ДОСЯГНУТО: артефакт не згенеровано. Причина — domain.Section.Columns []Section рекурсивний. generate.go:355 вмикає openapi3gen.ThrowErrorOnCycle(); kin-openapi v0.149.0 openapi3gen.go:204–208 повертає CycleError при повторному типі; schemaCustomizer викликається лише на :412 після обходу полів. Overrides застосовується саме у customize (generate.go:508), отже override Section/Document не може перехопити цей cycle. Components та NullableFields також не змінюють порядок обходу. Заміна domain shape чи DTO замість domain.LessonDocument порушила б заданий контракт; engine змінювати заборонено.
РІШЕННЯ: narrow STOP за брифом. Відкочено тільки шість власноруч змінених code/overlay/test файлів до перевіреного clean HEAD; frozen, engine та generated не змінені. Patch спроби: /tmp/courses-s01m-attempt.patch. Сторонній docs/work WIP збережено. Нових commit немає.
ФАКТ: gofmt -w touched files виконано (exit 0); plain artifact/fidelity, gofmt -l ., vet/build/targeted tests NOT_RUN через STOP. Full go test, integration, e2e NOT_RUN за scope. Жодного PASS для S0.1m не заявлено.
ФАКТ: mandatory native checkpoint спробовано у поточній codex session 01a182bc-17c0-7302-ae58-089de91574a5: sandbox exit 1 (receipts write denied), escalated invocation exit 2 «BLOCKED required handoff»; OK checkpoint не отримано. Це не auto-review rejection і не доказ успішного checkpoint.
ДАЛІ: власник має вирішити підтримку рекурсивних схем у generator окремим scope або змінити бриф; S0.1m BLOCKED до цього рішення.
ПИТАННЯ: потрібне owner-рішення щодо виявленого обмеження, не виправлення frozen schema.

ФАКТ: фінальна checkpoint-спроба після rollback завершилась exit 0, status OK; exact checkpoint ID `293fa10e009b977e0a049eb50bd305ce-1-f2b0a201c3e1`. Попередній BLOCKED не є фінальним результатом checkpoint.
ДАЛІ: owner decision щодо recursive-schema support; S0.1m залишається STOP. Backend diff порожній.

РІШЕННЯ (власник, 2026-10-28, §16.23): СХВАЛЕНО engine-фікс — generator підтримуватиме рекурсивні
NameMapped-схеми ($ref на власний компонент замість ThrowErrorOnCycle), fail-closed для un-named циклів
збережено. Вузько: лише generate.go, лише ця здатність. Resume Codex session 01a182bc (bg bodrtgvh5) з
розширеним scope + ЖОРСТКИЙ regression-guard: усі 80 наявних роутів byte-identical, усі 10 fidelity-пакетів
PASS. Engine-зміна = найвищий blast radius → Courses 2 рев'ю прискіпливо. Diff-чистота = головний критерій.
Verify engine-STOP коректний (ThrowErrorOnCycle :354 падає до customizer/Overrides :508) — [[two-plausible-causes-means-wrong-instrument]] не застосовне, причина одна й підтверджена кодом.

РІШЕННЯ §16.23 (2026-10-28, власник у поточному дорученні): дозволено вузько змінити generate.go для recursive NameMapped schemas; overlay.go/frozen незмінні. Cycles без NameMap fail-closed; жодної зміни наявного generated output. Writer Codex; branch s0.0-spike @9db7d30, backend clean, попередній docs WIP збережено. READY для розширеного scope; окремі engine-тести явно дозволені. Перевірка: direct/transitive recursion + unnamed negative, existing artifact no-regression, усі 12 зазначених package trees, vet/build/format.
ДАЛІ: відновити patch і додати cycle preflight до reflection; canonical isolated reflection використовує початковий allowlist, щоб тимчасове звуження NameMap не забороняло вкладену дозволену рекурсію.
ПИТАННЯ: немає.

ФАКТ: engine tests PASS exit 0 (/tmp/courses-s01m-engine.log): direct/transitive mapped cycles → ref, той самий тип без NameMap та частково mapped cycle → error. Перша UPDATE після відновлення patch exit 1 на YAML placement (path-блоки опинились після servers); переміщено під paths. Generated ще не переписаний.
ДАЛІ: повторити UPDATE та field fidelity.

ФАКТ: наступна UPDATE exit 1: unsupported parameter metadata …/locale/schema/enum. overlay.go дозволяє default/format, але не enum для params.
РІШЕННЯ (Codex, 2026-10-28): locale enum передано через вже наявний Param.Enum у contribution; default лишився в overlay. Frozen shape той самий, engine/overlay.go не розширюється. Перевірка — field fidelity. Plain-field locale/schemaVersion/contentType лишаються дубльованими metadata, не заявляються drift-proof від validators.
ДАЛІ: UPDATE і fidelity.

ФАКТ: UPDATE exit 0; course full-closure fidelity exit 0, nested LessonSpan.text mutation detected, дві генерації byte-identical. Generated diff = 342 additions / 0 deletions, existing lines unchanged (/tmp/courses-s01m-regression.log).
ФАКТ: після recursion fix зняття Block override дає exit 1 «unregistered json.Marshaler domain.Block» (/tmp/courses-s01m-block-guard.log). Наявний engine invariant потребує override ще до порівняння; frozen LessonBlock відтворено explicit override, enum береться із BlockTypes. MarshalJSON guard не послаблено. Override повернуто; повний field fidelity з ним PASS. Publish тепер має literal Response{Empty:true}, без nil inference.
ДАЛІ: фінальні gofmt clean → vet → build → всі 12 package trees → UPDATE/plain artifact; зчитати exit із логів.
ПИТАННЯ: немає.

ФАКТ: сильніший regression guard порівняв цілі existing YAML-блоки з HEAD: 80 paths, 102 schemas, 36 responses, 3 headers, 1 parameter — byte-identical; решта top-level metadata також незмінна. Added рівно 9 document schemas + 3 paths/4 operations + LessonDocumentConflict response. Лог /tmp/courses-s01m-regression.log, EXIT_CODE=0.
ФАКТ: фінальні format/format-check/vet/build логи /tmp/courses-s01m-{format,format-check,vet,build}.log мають EXIT_CODE=0; gofmt -l . порожній.
ДАЛІ: дочекатися всіх package trees та UPDATE/plain; фінальний scoped diff і checkpoint.

ФАКТ (фінальний writer gate, 2026-10-28): GOCACHE=/tmp/courses-s01m-go-cache. Прочитано логи: gofmt -w / gofmt -l . (порожній) / go vet ./... / go build ./... / go test усіх 12 погоджених package trees / UPDATE artifact / plain artifact — кожен EXIT_CODE=0. Логи: /tmp/courses-s01m-{format,format-check,vet,build,packages,update,plain}.log. Усі модульні fidelity пройшли в package run; course full closure + nested deletion і дві генерації також окремо PASS.
ФАКТ: frozen openapi.yaml і overlay.go не змінені; scoped backend diff рівно 9 дозволених файлів (включно з явно запитаним engine test). Generated +342/-0; 80 існуючих paths і 102 schemas byte-identical. Runtime wiring не змінено. Domain validation еквівалентна; request omitempty не впливає на decode. Жодних commit/push, full go test ./..., integration чи e2e. Сторонній WIP збережено.
РІШЕННЯ/ЛІМІТИ: §16.23 реалізовано тільки як cycle preflight з початковим NameMap + стандартний kin-openapi cycle ref. Unnamed і частково named цикли fail-closed. Plain locale/schemaVersion/contentType/timestamp metadata лишається ручною (fidelity ловить розбіжність із frozen, не з validators). Block explicit override обов’язковий через чинний json.Marshaler guard; permissive frozen schema збережена. Locale query enum у Param.Enum, бо чинний overlay schema enum для params відхиляє. Host full-suite та незалежне review NOT_RUN цим writer, це наступний крок власника.
ФАКТ: progress checkpoint OK exit 0: 293fa10e009b977e0a049eb50bd305ce-2-3e317907091f.
ДАЛІ: передати власнику готовий diff + журнал для host verification/Courses 2; S0.1n не починати.
ПИТАННЯ: немає.

Фінальні SHA-256 (base s0.0-spike @9db7d30, dirty snapshot):
- `backend/internal/course/adapters/http/const_sets_test.go` `2d43cd96a8d7558cd5c90d543fb48ad13002eb25435a4f5de5642caa6a949692`
- `backend/internal/course/adapters/http/document_handler.go` `cb378dab4f99624fe9bd4f38a9246d0d3513daded6dca967d1e66bf12c684076`
- `backend/internal/course/adapters/http/openapi_contract.go` `cfc60a9a0faa1b93c36140090c880ed7f09f00cbb05863c040a5732c61af4cd0`
- `backend/internal/course/adapters/http/openapi_contract_test.go` `e2d9084d0cc92275bed2e9f31265d1392ba685595de936bbc7b134d25f9bb2a1`
- `backend/internal/course/domain/document.go` `ed3fbd3638dd8385873a1835ee8aaaf370669e8fae2b6dec63d980e7fd5f9298`
- `backend/internal/platform/openapigen/generate.go` `3c77583526fb24688a4536ef93dab597447762b434fd44edd69e7a97548e7a2f`
- `backend/internal/platform/openapigen/generate_test.go` `7ff1731d9e61dd94ef2db024bbbe1013d044c032fff4a7fe6c50d9410e1c60c3`
- `backend/openapi/openapi.generated.yaml` `34e06a0635107b8b128351d2cd370d4230812bc6b50fae12c79e0467fdd82ffd`
- `backend/openapi/openapi.overlay.yaml` `597189800188babb6c75e325c8233111768a9b296a1f65fe0b259226011303a3`
