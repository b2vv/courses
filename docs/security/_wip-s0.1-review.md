claude-opus-4-8

# S0.1 independent review (fresh context)

ДАЛІ: подивлюсь git log і --stat по семи комітах, перевірю інваріанти (жоден не чіпає openapi.yaml; генератор чіпають лише *-gen зрізи).

ФАКТ (інваріанти): `git show --stat` по 7 комітах — жоден не чіпає backend/openapi/openapi.yaml (frozen). generate.go/overlay.go змінені ЛИШЕ у 4 *-gen зрізах (53ef56d, eabc754, b06e2a6, 62cbbb6). Контриб'ютор-зрізи (d80ac66, 7b97977, 234c835) чіпають generated.yaml + overlay.yaml + файли модуля + (234c835) arch-тест; generate.go не чіпають. OK.

ФАКТ (53ef56d §16.19): injectivity замінено на schema-equality через canonicalSchema (рефлексія кожного типу ізольовано, порівняння JSON). Fail-closed для differing shape ("conflicting schema"). Тести TestSharedComponentNameAllowsIdenticalSchemaOnly (обидва напрями) + TestNonInjectiveNameMap — non-vacuous. КОНЦЕРН: прибрано check-time name-registry, що раніше ловив колізії й для fallback-імен (t.Name(), не в NameMap). Нова eager-перевірка охоплює лише NameMap; emission sameJSON (generate.go:358) ловить cross-call fallback-колізії, але within-single-graph fallback-колізію двох різних типів з однаковим простим іменем рефлектор тихо схлопне й ніхто не зловить. Вузько (усі контриб'ютори зелені, артефакт byte-identical), не блокер.

ФАКТ (eabc754 §16.20): Response.OneOf []any, emit як schema.oneOf у порядку слайса (order contract-significant, тест асертить порядок refs). Взаємовиключність із Body/Empty, мінімум 2 — усі 3 fail-closed напрями протестовані з error-substring. OK.

ФАКТ (b06e2a6 §16.21): "header" додано до дозволених локацій (cookie/object досі падають — TestInvalidGoParameters). Param.Component → emit у components.parameters + $ref; differing reuse → sameJSON → "conflicting parameter". Тест non-vacuous (ref+inline+conflict). OK.

ФАКТ (62cbbb6): mergeParamMetadata — panicking `param["name"].(string)` замінено на guarded skip $ref-параметрів. Не губить легітимну метадату: orphan overlay-ім'я досі fail-closed (overlay.go:236). Тест TestParameterMetadataToleratesRefSibling. OK.

ФАКТ (234c835 §16.22 arch-test): дві exemptions. (1) collector (internal/openapispec) звільнено від "no module imports orders" — scoped через pathWithin(collectorRoot). openapispec містить лише contributions.go+artifact_test.go, у ПРОДІ його ніхто не імпортує (лише per-module fidelity-тести) → не може протекти payment-факти в бізнес-модуль. (2) orders може імпортувати courses/internal/platform/* — у проді internal/platform містить ЛИШЕ openapigen, orders імпортує саме openapigen. Обидві prefix-exemptions на практиці мінімальні; реальний payment-isolation інваріант (жоден бізнес-модуль ↔ orders) збережено. КОНЦЕРН (forward-looking, не блокер): prefix ширший за потребу — майбутній пакет під internal/platform або бізнес-логіка в openapispec не ловитимуться.

ФАКТ (поведінка в проді): d80ac66 — omitempty лише на request-DTO (не впливає на decode), named string-типи з незмінними json-тегами, валідація slices.Contains по тому ж набору; state-filter рефактор еквівалентний (open/resolved/all/default ідентичні). 7b97977 — request-DTO json-теги = case-insensitive matching, що вже був; response named string-типи маршаляться у ті самі значення; admin request-обгортки еквівалентні конверсії. Behaviour-preserving. 234c835 wayforpay — Acknowledgement експортовано, Status=AcknowledgementStatus, json-теги й значення "accept" незмінні → byte-identical JSON.

ФАКТ (coverage/fidelity тести orders): TestOpenAPIContributionCoversRegistrations — AST-census 3 Register-функцій, 22 маршрути, guarded branches (TestRegistrationCensusIncludesGuardedBranches + негативна перевірка). TestOrdersOpenAPIFidelity — порівняння проти FROZEN openapi.yaml, non-vacuous (probe-видалення ламає fidelity, зміна header location, swap). OK.

ФАКТ (гейти, прогнав сам): gofmt -l . — чисто; go vet ./... — чисто; go build ./... — чисто; go test ./... — exit 0, 0 FAIL, courses/platform/architecture ok, orders/openapispec/openapigen ok. ЗЕЛЕНО.

РІШЕННЯ: усі 7 комітів PASS або PASS-WITH-CONCERNS; блокерів нема; гілка безпечна до S0.2 merge.
