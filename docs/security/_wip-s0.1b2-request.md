# S0.1b-2 — one course request DTO + request conformance (16.4 + 16.2) — WORK FILE

Дописуй ПІСЛЯ КОЖНОГО КРОКУ (append). Перший запис — `ДАЛІ:` до першого інструменту.
ФАКТ / РІШЕННЯ / ДАЛІ / ПИТАННЯ. Фінал = стислий переказ. Base: branch `s0.0-spike` @ `cb2e9b5`.

## Журнал

### 2026-10-27 — Codex: discovery / pre-code checkpoint

ДАЛІ: прочитати контракт і Go DTO, вибрати один scalar-field request route, звірити hand spec та handler fixture; product mutation лише після усунення суперечностей. Цей запис зроблено після read-only discovery (вступну вимогу запису до першого інструменту побачено під час читання файлу).

ФАКТ: root `/Users/strelia/projects/courses`, branch `s0.0-spike`, HEAD `36fee5eb7382c91c7140ef17f28953a3e9597144`; requested base `cb2e9b5` є його безпосереднім предком. HEAD додає тільки попереднє review/disposition; не скидав його. На вході єдиний untracked файл — цей наданий work-file; оригінальний заголовок збережений. Початковий SHA-256 work-file: `db020f22ee5c4c27d66ce55796b286a3ff3f6f54d2c89c12e0177cc3fdb9e6a9`.

Scope/owner: доручення власника S0.1b-2, §16.2 + §16.4; один request route, без document/Block, error-code parser, інших модулів, hand-spec edits, merge/push. Small: генератор/контрактні тести, без зміни публічного API чи runtime behavior; inline plan достатній за `.claude/standards.md`. Незалежне per-slice review — Courses 2 за планом §16.6, NOT_RUN; planning critique/defense N/A для цього малого scope. Виконавець/writer — поточний Codex; інших writers не виявлено.

ФАКТ: `RouteBinding` не має Request; `Generate` відбиває лише responses. Existing `customize` робить pointer nullable; `NullableFields=false` не скасовує pointer-nullability (`nullable || Kind()==Pointer`). Overlay зараз не містить request structure, тож переносити/видаляти її поки немає звідки. File fallback замість graph: перевірено generate.go/overlay.go, course contribution/fidelity/response-conformance, admin DTO/handlers/fixtures і відповідні hand schemas. Ніяких graph refresh або product mutations.

РІШЕННЯ (пропозиція Codex, 2026-10-27): кандидат `POST /api/v1/admin/courses/{id}/status`, `transitionAdminCourseRequest`: лише два scalar-string поля, без вкладених документів, масивів чи custom UnmarshalJSON; є реальний handler fixture. DELETE має лише один scalar, але потребує додаткової підтримки bodyless 204, яку нинішній generator відмовляє як nil response body. Create/save додають optionalPriceDTO або дерево. Перевірити вибір можна за `admin_handler.go:125`, hand spec `openapi.yaml:936`, fixture `admin_handler_test.go`.

План після узгодження: спільна reflection-функція request/response → Request binding та scalar metadata → fidelity до hand request/ref closure → ValidateRequest над тими самими байтами, що успішно пройшли admin handler, плюс missing/null negative cases → regen + second-generation equality → чотири gates у brief-порядку → scoped commit з заданими trailers. Це план, НЕ виконаний доказ.

## ПИТАННЯ

**STOP / pre-code BLOCKED — request pointer null semantics.** §11.2 вимагає `*T → nullable:true`, brief вимагає ті самі rules і незмінну fidelity. Але `transitionAdminCourseRequest.Status` та `.ExpectedUpdatedAt` — `*string` без omitempty (`admin_handler.go:125–128`), тоді як `AdminCourseStatusRequest` (`openapi.yaml:3206`) вимагає обидва поля і не дозволяє null; `AdminTimestamp` (`:3271`) теж non-null. Handler `:268` явно відхиляє nil для обох полів. Отже pointer тут кодує missing/null detection під час decode, не дозволений null у wire contract. DELETE має ту саму суперечність (`:130`, `:296`).

Потрібне рішення власника: дозволити явне request-specific non-null narrowing за nil-check handler, лишивши required і hand spec без змін, чи змінити контракт генератора іншим узгодженим способом. Не сховав суперечність в overlay `nullable:false`, не послабив fidelity/validation assertions, не підмінив production request DTO синтетичним. §16.4 також згадує optional-pointer semantics; у transition DTO немає omitempty — окремий generator unit case може довести це правило, але його не можна назвати production-route optional-field proof.

ФАКТ: fixture доступний (`seededAdminWriteRepository`, `adminReadHandler`, `serveAdminJSON`/`serveAdminRawAs` відновлює ті самі request bytes після handler). Blocker — контракт, не fixture.

Gates: gofmt / go vet / go build / scoped go test — **NOT_RUN** (STOP до реалізації). Request fidelity / ValidateRequest / regeneration — **NOT_IMPLEMENTED / NOT_RUN**. Код і generated/hand spec не змінені; комітів, merge, push немає. ДАЛІ: відповідь власника на nullability, потім продовження цього самого slice.

## RESUME (для свіжої сесії після handoff — Codex НЕ відновлено, це наступна дія)
Codex S0.1b-2 session id: `01a17e84-4953-7a41-8ce1-b62996482d93` (branch s0.0-spike).
Рішення §16.10 ухвалено. Наступна дія: відновити Codex цією відповіддю (НЕ переробляти):
`cd /Users/strelia/projects/courses && codex exec --approve-for-me resume 01a17e84-4953-7a41-8ce1-b62996482d93 "<ТЕКСТ НИЖЧЕ>" -o <scratch>/r.md > <scratch>/r.log 2>&1 < /dev/null` (у фоні).
Текст відповіді (= §16.10 + кроки):
1. Правило `*T→nullable` — RESPONSE-only; для REQUEST nullability вказівника = явне per-field рішення
   (розширити NullableFields на request-поля), генератор fail-close на request-pointer без рішення;
   контекст req/resp генератор знає з RouteBinding.Request vs Responses. required — за hand-контрактом/omitempty.
2. transitionAdminCourseRequest (POST /admin/courses/{id}/status): обидва *string = явний non-null, required;
   status enum [draft,published,archived] → overlay (як response status); expectedUpdatedAt→AdminTimestamp.
3. Довести: fidelity на request-body + closure vs рукописний openapi.yaml (non-vacuous); ValidateRequest на
   реальних байтах, що хендлер приймає (admin-фікстури) + негативні кейси (nil status/nil expectedUpdatedAt → FAIL)
   = route two-layer (16.2 опція b).
4. 16.4 optional-pointer: transition доводить non-null required шлях; + generator UNIT-тест на genuine-optional
   (request `*T` з omitempty → nullable+optional).
5. Структура з Go; overlay лише validation/prose; видалити будь-яку overlay-request-структуру в тому ж коміті;
   fidelity зелений = доказ; regen openapi.generated.yaml (2-а генерація byte-identical).
6. Гейти по порядку (gofmt/vet/build/test), реальний вивід. Малі коміти на s0.0-spike, trailers як було.
7. Scope: лише цей request-route + request-conformance + generator request-nullability + optional-pointer unit test.
Далі: writer-верифікація гейтів → пакет Courses 2 (per-slice) з акцентом на request-nullability handling.
