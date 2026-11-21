ДАЛІ: Codex за брифом (scratchpad s0.1n-brief.md) реалізує S0.1n → host-verify → коміт → Courses 2.

# S0.1n — course module: admin course CRUD (останній зріз перед S0.2 cutover)

Writer: Courses 1 (via Codex). Reviewer: Courses 2 (courses-96 [557618]).
Branch `s0.0-spike`. Base HEAD: `bb1feee` (після S0.1m PASS + вердикт).

5 ops / 2 paths, **нерекурсивний** (engine-зміни НЕ треба; якщо Codex упреться в ліміт генератора — STOP, не чіпати frozen/engine без owner-рішення, як §16.23):
- GET `/api/v1/admin/courses` → adminCoursePageDTO; params limit/offset/status/search/sort
- POST `/api/v1/admin/courses` → Request createAdminCourseRequest, 201 adminCourseTreeDTO
- GET `/api/v1/admin/courses/{id}` → adminCourseTreeDTO
- PUT `/api/v1/admin/courses/{id}` → Request saveAdminCourseRequest, 200 adminCourseTreeDTO
- DELETE `/api/v1/admin/courses/{id}` → Request deleteAdminCourseRequest, 204 bodyless (Response.Empty)

## Факти (до коду)
- Handler-DTO всі існують у admin_handler.go: adminCourseDTO(24)/adminCourseStatusCountsDTO(36)/adminCoursePageDTO(43)/adminLessonDTO(49)/adminModuleDTO(55)/adminCourseTreeDTO(61)/adminCourseWriteRequest(77)/createAdminCourseRequest(alias)/saveAdminCourseRequest(119)/saveAdminModuleRequest/saveAdminLessonRequest/deleteAdminCourseRequest(130).
- Frozen схеми: AdminCoursePage(2937)/AdminCourseTree(2967)/AdminCourseWrite(2988)/AdminCourseSaveRequest(3191)/AdminCourseDeleteRequest(3214) — реалізація мусить бути byte-faithful (імена компонентів + shape).
- Error-компоненти ВСІ існують: AdminWriteBadRequest(3594)/AdminMutationForbidden(3602)/TranscriptConflict(3528,409 reused)/InvalidRequest(3610)/Forbidden(3634)/NotFound(3649)/Unauthenticated/InternalError/Unavailable — лише $ref із contribution, overlay НЕ чіпати.
- Єдине design-рішення: `sort` [recent,title,students,price] НЕ має ordered set (query.go: PublicCatalogSorts=[recent,title] 2-subset; NewAuthoringQuery інлайнить 4). → додати `var AdminCourseSorts = []Sort{SortRecent,SortTitle,SortStudents,SortPrice}` (коментар), NewAuthoringQuery → newQuery(...,AdminCourseSorts...), param Enum: enumStrings(domain.AdminCourseSorts) + census (generator enum == set). PublicCatalogSorts лишити.
- `status` param [draft,published,archived] = inline Enum: enumStrings(domain.CourseStatuses) на Param (НЕ через Enums type-map: domain.Sort/Status мають по два allowlist, тому param-enums inline як public /courses GET).
- Снаг: adminCourseWriteRequest.Price тип optionalPriceDTO{Value,Set} з custom UnmarshalJSON (decode-only) — reflected shape {value,set} ≠ frozen nullable Price. Мапити Price-поле через FieldRef на наявний pattern NullablePrice (компонент уже в Components). Звірити з frozen AdminCourseWrite.price.
- 204 DELETE → Response{Empty:true} (наявний маркер §16.18). Headers X-Request-ID/Cache-Control(NoStore) на кожному 2xx — як у наявних роутах.
- Патерн binding/NameMap/NullableFields — openapi_contract.go (S0.1l/S0.1m зразок): NameMap нові adminCourse* DTO → frozen-імена; NullableFields на nullable масиви/вказівники звірити з frozen.

## Журнал
- (до коду) Target зібрано, бриф написано. Чекаю Codex.
ДАЛІ: перевірити briefing, правила AI Factory, frozen admin-course paths/схеми та поточні DTO і contribution.
ФАКТ: HEAD bb1feee; є чужі modified _wip-s0.1b* та untracked журнали — зберігаю. Прочитано briefing/workflow; SA10 S0.1m закрито, S0.1n явно доручено.
ДАЛІ: прочитати frozen та generator customization/tests, визначити точні bindings.
ФАКТ: frozen search (не q); Price → NullablePrice; timestamps → AdminTimestamp/NullableAdminTimestamp; Save.modules → AdminModule. AdminLesson.isPreview optional; Write instructorId/price optional, решта required. ExpectedUpdatedAt required, omitempty не потрібне.
РІШЕННЯ: small / READY за явним brief власника; лише parallel artifact та еквівалентний sort allowlist. Plan-critique N/A: немає нового API/архітектури. Review виконує Courses 2/host після передачі; не заявляю незалежний PASS. ДАЛІ: перевірити overlay і поточні fidelity/census helpers.
ФАКТ: FieldRefs у generate.go:610–618 допускає лише non-nullable scalar targets; Price через нього неможливий. Чинний overlay підтримує replacement $ref (overlay.go:53–70); використаю існуючий NullablePrice ref pattern, decode-helper Override та точкові overrides для спільних AdminModule/AdminLesson без зміни response marshalling. Engine не потрібен.
ДАЛІ: реалізувати bindings, allowlist і additive metadata.
ФАКТ: metadata для нових paths/schemas у overlay ще немає; error components існують. Додаю лише нові sibling-блоки, existing bytes зберігаються. ДАЛІ: реалізація та targeted fidelity.
ФАКТ: додано 5 bindings/2 paths, AdminCourseSorts, shared nested overrides; request-only instructorId/price отримали omitempty (decode без змін). Metadata витягнуто з frozen, лише нові siblings. ДАЛІ: форматування та fidelity, виправити лише contribution-level відмінності.
ФАКТ: fidelity розширено на обидва paths + повний ref closure; кожне поле кожної з 9 нових object schemas має mutation non-vacuity. Sort census перевіряє validator + actual generator після зміни лише domain set. ДАЛІ: прочитати targeted result.
ФАКТ: targeted3 EXIT=1, лог /tmp/courses-s01n/targeted3.log.
ФАКТ: targeted2 EXIT=1, лог /tmp/courses-s01n/targeted2.log.
ФАКТ: metadata helper перші два запуски EXIT=1 (синтаксис/API YAML), виправлено імпорт на yaml3; фінальний metadata EXIT=0. Перший compile використовує новий GOCACHE; тести ще виконуються.
ДАЛІ: дочекатися targeted, перевірити actual refs/required.
ФАКТ: targeted 1/2/3 EXIT=1. Реальна розбіжність brief: AdminWriteBadRequest відсутній у generated та overlay; newly emitted response має порожній description/без headers. ПИТАННЯ власнику: дозволити один additive metadata-блок цього error; до відповіді error overlay не змінюю. ДАЛІ: незалежно звірити решту surface та запустити vet/build.
ФАКТ: targeted4 EXIT=1; /tmp/courses-s01n/targeted4.log.
ФАКТ: targeted4 EXIT=1 лише AdminWriteBadRequest; усі component schemas збігаються. Native checkpoint default EXIT=1: sandbox заборонив mktemp у registered receipts. ДАЛІ: повторити checkpoint з escalation, виконати ordered gofmt/vet/build.
ФАКТ: gofmt -l . EXIT=0; порожній stdout перевірено в /tmp/courses-s01n/gofmt.log.
ФАКТ: native checkpoint escalation EXIT=2 HANDOFF_OUTCOME_UNCERTAIN; не повторюю сліпо. Звіряю саме manual receipts поточної session, не transcript і не latest note.
ФАКТ: go vet ./... EXIT=0; /tmp/courses-s01n/vet.log.
ФАКТ: go build ./... EXIT=0; /tmp/courses-s01n/build.log.
ФАКТ: ordered gofmt/vet/build EXIT=0. Native handoff залишається NOT_CONFIRMED (exit 2), не видаю за успішний checkpoint. ДАЛІ: дві генерації, plain artifact, additive-only diff, full Go suite.
ФАКТ: generation1 EXIT=0; /tmp/courses-s01n/generate1.log.
ФАКТ: generation2 EXIT=0; /tmp/courses-s01n/generate2.log.
ФАКТ: plain artifact EXIT=0; /tmp/courses-s01n/artifact.log.
ФАКТ: generation1/2 EXIT=0; cmp EXIT=0 (byte-identical). Additive sibling audit виконано: existing paths/schemas/responses незмінні; деталі в tool output. ДАЛІ: full Go suite та residual fidelity.
ФАКТ: виправлення попереднього запису: additive audit НЕ пройшов — helper припустив 2-space indent, отримав 0/0 paths та StopIteration; PASS не заявляється. ДАЛІ: повторити з фактичним 4-space indent.
ФАКТ: corrected additive audit EXIT=0: paths 83→85, existing siblings byte-identical; generated/overlay deletions=0; frozen/engine diff порожній. Plain artifact EXIT=0.
ФАКТ: receipt manual.apHnDHEu має worker result OK, checkpoint_id 3f5f80a28589a785cad83e9f69a8da69-1-0b508d4ce0a4, але wrapper EXIT=2 / validator_terminal FAILED. Це не успішний native checkpoint за контрактом; сліпого retry не роблю. ДАЛІ: завершити full-suite, hash manifest; error metadata чекає owner-відповіді.
ФАКТ: full go test -count=1 ./... EXIT=1; /tmp/courses-s01n/full-test.log.
РІШЕННЯ (Codex, 2026-10-28): shared AdminModule/AdminLesson overrides зберігають frozen форму при різних request-pointer/response-value DTO. Ліміт: override може приховати майбутнє DTO-поле; frozen fidelity перевіряє wire schema, не decoder policy. Перевірка: TestCatalogOpenAPIFidelity + handler tests; engine не змінювався. Native checkpoint refresh BLOCKED через uncertain wrapper outcome, точний persisted id записано вище.
ФАКТ: full suite EXIT=1 (fidelity: missing AdminWriteBadRequest metadata + parameter ordering). РІШЕННЯ (Codex, 2026-10-28): параметри порівнюються за (in,name), як adminstats/discussion fidelity; порядок array не є API-семантикою, жодне поле/enum не ігнорується. ДАЛІ: targeted recheck, full recheck після owner metadata-рішення.
ФАКТ: targeted5 EXIT=1; /tmp/courses-s01n/targeted5.log.
ФАКТ: missing error metadata блокує aggregate fidelity раніше mutation proof. Виніс field-by-field admin schema fidelity в окремий тест: baseline кожної schema == frozen, потім видалення кожного поля повинно давати diff. ДАЛІ: final ordered Go verification (aggregate очікувано BLOCKED до owner-рішення).
ПИТАННЯ (pending): дозвіл на один additive AdminWriteBadRequest metadata-блок; точне джерело backend/openapi/openapi.yaml:3594–3601. Підготовлена пропозиція /tmp/courses-s01n/AdminWriteBadRequest-proposed.yaml, у repository error overlay НЕ застосована.
ФАКТ: isolated schema fidelity + sort census EXIT=0; /tmp/courses-s01n/schema-census.log.
ФАКТ: final gofmt EXIT=0 (stdout порожній), vet EXIT=0, build EXIT=0; schema fidelity 9/9 + mutation кожного поля і emitted-sort census PASS (EXIT=0). Aggregate fidelity після нормалізації параметрів EXIT=1 лише AdminWriteBadRequest metadata.
ФАКТ: DELETE 204 без content перевірено aggregate path comparison (відмінностей paths немає). Errors на ops — component $refs; existing errors незмінні. Schema/status validation metadata перенесено з frozen, не заявляється автоматично виведеним із validators.
ФАКТ: final full go test EXIT=1; /tmp/courses-s01n/final-test.log.

ФАКТ: SHA-256 власних змінених implementation/test/artifact файлів (перед final full-suite result):
```text
10b259f3ecf2fa3dbc179594127a0f223660ae2506d819d89a494b3f16f69112  backend/internal/course/adapters/http/admin_handler.go
9e679d7871cb9d45aacf86b2a7167115956ae7938cc29d57b1886e7309e8cae9  backend/internal/course/adapters/http/const_sets_test.go
53ee26312eb7790d0075ae1389052aca82ac0f8d079ccfff4c868a39c6581a57  backend/internal/course/adapters/http/openapi_contract.go
6525350c07323483265721738f1b3c03fc7bf891dd645ab98a3b3123c770afad  backend/internal/course/adapters/http/openapi_contract_test.go
aba4b5c42c98f3ac82005a9d90915a8465fd4105aaaa644370bf47f1f4949736  backend/internal/course/domain/query.go
12788712028acb4934a73745399a819c1900a49e272dcb13ff92a0afe00a39a8  backend/openapi/openapi.generated.yaml
3d1261e497e7bd82ffb306c512e28b02e6024a16bff3cecd49cecf6a08d21212  backend/openapi/openapi.overlay.yaml
```
ДАЛІ: Courses 1 — дозволити/додати missing AdminWriteBadRequest metadata, регенерувати і повторити fidelity/full suite; host-verify та незалежне review, без мого commit. S0.1n не закрито до зеленого full suite.
ФАКТ: final full suite EXIT=1; єдиний failing test — TestCatalogOpenAPIFidelity через AdminWriteBadRequest description/headers, panic немає. Решта packages PASS/no test files. Коміт не створено. STATUS: BLOCKED на явному обмеженні brief (error overlay не чіпати), дозвіл ще не отримано. Native checkpoint NOT_CONFIRMED: wrapper EXIT=2 попри worker OK; receipt/id вище. Усі SHA-256 implementation файлів вище актуальні.
ФАКТ: виправлення попереднього запису після повного читання final log: є ДРУГИЙ FAIL — platform/httpserver TestShutdownDrainsAnInFlightResponse; попереднє твердження «єдиний failing test/решта PASS» скасовано. ДАЛІ: ізольовано повторити цей тест і записати точну причину, не приховувати full-suite EXIT=1.
ФАКТ: shutdown isolated recheck EXIT=1; /tmp/courses-s01n/shutdown-recheck.log. Full suite лишається EXIT=1.
ФАКТ: shutdown isolated EXIT=1 також через listen tcp 127.0.0.1:0 bind operation not permitted. Full suite повторено з escalation (дозвіл отримано автоматичним review), результат очікується. SHA-256 implementation файлів незмінні; journal SHA фіналізується після результату.
ФАКТ: full suite з дозволом local listener EXIT=1; /tmp/courses-s01n/host-test.log.
ФАКТ: host full suite завершено EXIT=1; єдиний FAIL тепер TestCatalogOpenAPIFidelity (AdminWriteBadRequest metadata), platform/httpserver PASS. Panic немає. Остаточний стан BLOCKED: потрібен дозволений additive error metadata-блок, після нього regen/plain/fidelity/full suite. Frozen, engine, overlay.go незмінні; generated +479/−0, overlay +12/−0; без коміту. Native checkpoint wrapper залишається непідтвердженим (EXIT=2), не PASS.
ДАЛІ: resume за авторизацією власника — звірити поточний diff і frozen/overlay response convention; додати лише AdminWriteBadRequest metadata, регенерувати та виконати ordered verification. Native checkpoint не ретраїти (відомий wrapper EXIT=2).
ФАКТ: resume hashes усіх 7 implementation/test/artifact файлів збіглися; HEAD bb1feee. AdminMutationForbidden overlay має description+headers, без content; новий AdminWriteBadRequest додано тією ж конвенцією, опис і refs точно з frozen. Delta від resume snapshot — рівно один блок; EXIT=0.
РІШЕННЯ: попередній BLOCKED на error metadata знято явною авторизацією власника. ДАЛІ: generation×2, cmp, plain artifact; після них ordered gofmt/vet/build, additive audit та full suite.
ФАКТ: resume generate1 EXIT=0; /tmp/courses-s01n/resume-generate1.log.
ФАКТ: resume generate2 EXIT=0; /tmp/courses-s01n/resume-generate2.log.
ФАКТ: resume cmp EXIT=0 (дві генерації); /tmp/courses-s01n/resume-cmp.log.
ФАКТ: resume plain EXIT=0; /tmp/courses-s01n/resume-plain.log.
ФАКТ: resume ordered format EXIT=0; /tmp/courses-s01n/resume-format.log (format перевіряє також порожній gofmt stdout).
ФАКТ: resume ordered vet EXIT=0; /tmp/courses-s01n/resume-vet.log (format перевіряє також порожній gofmt stdout).
ФАКТ: resume ordered build EXIT=0; /tmp/courses-s01n/resume-build.log (format перевіряє також порожній gofmt stdout).
ФАКТ: resume additive regression EXIT=0; /tmp/courses-s01n/resume-additive.log. ДАЛІ: full suite з local listener.
ФАКТ: resume full go test -count=1 ./... EXIT=0; /tmp/courses-s01n/resume-full.log (local listener дозволено).

ФАКТ: фінальний resume audit EXIT=0 — усі 9 логів містять EXIT=0; full suite без FAIL/panic, platform/httpserver PASS. TestCatalogOpenAPIFidelity, TestAdminSortSingleSource та TestAdminCourseSchemaFidelity (9/9 schemas, mutation non-vacuity) включені в uncached full suite й PASS.
ФАКТ: verification EXIT: generation1=0; generation2=0; cmp=0; plain artifact=0; gofmt (порожній stdout)=0; vet=0; build=0; additive audit=0; full go test -count=1 ./...=0. GOCACHE=/tmp/courses-s01n-go-cache.
ФАКТ: generated +484/−0; overlay +19/−0 проти bb1feee. Попередні 83 paths / 111 schemas / 37 responses generated байт-ідентичні; старі overlay siblings теж незмінні. Resume overlay delta — лише AdminWriteBadRequest description+headers, content генерується contribution. Frozen, overlay.go та generate.go незмінні.
РІШЕННЯ: S0.1n implementation/verification PASS, попередній metadata blocker закрито. Незалежний host-verify/review лишається Courses 1; коміт не зроблено. Native checkpoint сліпо не повторювався згідно owner instruction: відомий wrapper EXIT=2 залишається непідтвердженим, не review-відмова.
ФАКТ: оновлені SHA-256 усіх 7 власних змінених source/test/artifact файлів:
```text
10b259f3ecf2fa3dbc179594127a0f223660ae2506d819d89a494b3f16f69112  backend/internal/course/adapters/http/admin_handler.go
9e679d7871cb9d45aacf86b2a7167115956ae7938cc29d57b1886e7309e8cae9  backend/internal/course/adapters/http/const_sets_test.go
53ee26312eb7790d0075ae1389052aca82ac0f8d079ccfff4c868a39c6581a57  backend/internal/course/adapters/http/openapi_contract.go
6525350c07323483265721738f1b3c03fc7bf891dd645ab98a3b3123c770afad  backend/internal/course/adapters/http/openapi_contract_test.go
aba4b5c42c98f3ac82005a9d90915a8465fd4105aaaa644370bf47f1f4949736  backend/internal/course/domain/query.go
dacab5358881cb7e0e1a4d1646491492bf45461ed20e5616324a8722942e231a  backend/openapi/openapi.generated.yaml
e4569995a7ce205491f4b936b2eb7c30466d3a231d411e6efe41fc487ae45f20  backend/openapi/openapi.overlay.yaml
```
ФАКТ: SHA-256 журналу до цього фінального append (перші 17613 байтів): fa3a611a931168fdce4f572fb95911c7451539ad0af6041b119b15b3b00018a6. Повний self-hash неможливо включити у той самий файл; фінальний SHA журналу разом із 7 файлами — /tmp/courses-s01n/SHA256SUMS.
ДАЛІ: Courses 1 — host-verify, незалежне review та commit за власним рішенням. Питань/нових контрактних розбіжностей немає.
