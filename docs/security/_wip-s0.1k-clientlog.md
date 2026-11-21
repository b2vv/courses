# S0.1k — clientlog contributor (10th)

Writer: Courses 1 (via Codex). Reviewer: Courses 2 (existing_session).
Branch: `s0.0-spike`. Base HEAD at brief time: `35f886f`.

ДАЛІ: запустити свіжий `codex exec` за брифом (scratchpad s0.1k), далі host-verify
(TestOpenAPIArtifact UPDATE→plain, gofmt/vet/build + повний `go test ./...`), коміт, Courses 2.

## Scope / acceptance

Останній непокритий модуль перед S0.2 cutover. Один публічний роут:
`POST /api/v1/client-logs` (unauthenticated, 204/400). Після цього генератор
покриває 79/85 (лишаються 6 course-роутів — окремий зріз).

Fidelity-ціль — FROZEN `backend/openapi/openapi.yaml`:
- path block: operationId `ingestClientLog`, requestBody.required=true (auto),
  204 (bodyless, §16.18), 400 → `$ref BadRequest`.
- `ClientLogRequest`: object, `additionalProperties:false` (auto), `required:[message]`,
  7 string props, level `enum:[error,warn,info]`, maxLength per поле (2000/512/128/256/256/8000).

## Рішення / факти (до коду)

- ФАКТ (generate.go:544-546): `required` = поля без `omitempty`. Отже `message` без
  omitempty (required), решта 6 — з omitempty (optional). `additionalProperties:false`
  авто (543).
- ФАКТ: `clientLogRequest.Level` зараз `string` → enum не чіпляється. Named-type підхід
  (прецедент media state/failureReason): `Level domain.Level`. Decode незмінний (underlying
  string), `NewClientEvent(Input{Level: string(request.Level)})` — behaviour-preserving.
- ФАКТ: enum single-source — додати `var Levels = []Level{LevelError,LevelWarn,LevelInfo}`,
  `normalizeLevel` через `slices.Contains(Levels,...)` (порожнє→error лишається), census
  `TestLevelConstSetCensus` (declared==set==emitted, +guard vacuum). Прецедент S0.1b-3b/S0.1e-1.
- РІШЕННЯ (hand `required:[message]`): СЕРВЕР message НЕ валідує як обов'язкове (порожнє
  допускається, scrub+truncate). Тобто hand `required:[message]` — документаційний вибір,
  НЕ серверно-примусовий. Політика S0 (§16.11): hand openapi.yaml НЕ чіпаємо до S0.2, окрім
  реального consumer-facing баґа. message-required таким баґом НЕ є (клієнт завжди шле message).
  → ЗБЕРІГАЄМО frozen контракт: message required, решта optional. Нуль змін hand-spec.
  Якщо Codex/Courses 2 вважає інакше — NARROW STOP, рішення власника (§16.16).
- ФАКТ: maxLength — у overlay (прошарок validation/prose, прецедент). Значення збігаються з
  domain unexported caps; дублювання мінімальне, НЕ експортуємо (scope).

## Журнал

- (до коду) Інваріанти зібрано з git+коду; брифа написано. Чекаю Codex.
- Codex run1 (session 01a18237, bg ba9jc6ge0): коректний NARROW STOP — бриф-інваріант №3
  дозволяв лише *тестові* нові файли, а пункт B потребує не-тестового `openapi_contract.go`.
  0 змін, 0 гейтів. Розширив scope (resume-бриф), resume (bg bcmk8pgt0). Приклад дисциплінованого
  stop на неточність контракту — [[codex-may-open-discussion]].

ДАЛІ: реалізувати A–F у розширеному власником scope; frozen spec і engine не змінювати.
ФАКТ: pre-code READY, small (паралельний артефакт без зміни API); root <repo>, branch s0.0-spike, HEAD 35f886fa531273bc0b8d96f7d1e3468279ae9d72. Scoped code clean; сторонній docs WIP збережено. Pin 17a9527 verify_release exit 0. Impact: event.normalizeLevel → handler.ingest; Register → contribution → AllOpenAPIContributions → artifact. File fallback достатній для одного маршруту.
РІШЕННЯ: 2026-10-28, Codex за брифом власника: окрема планова критика N/A для механічного small-зрізу за існуючим S0 планом; reviewer Courses 2 existing_session, передача через власника Courses 1, незалежний review ще NOT_TESTED. Перевірки — лише дозволені брифом, full suite на хості власника.
ФАКТ: A/B DTO — Levels і normalizeLevel через slices.Contains; empty→error збережено; optional tags та domain.Level у DTO, явне перетворення до Input string; decode/emit без змін.
ФАКТ: B — створено дозволений openapi_contract.go: один POST, Empty 204, shared BadRequest 400, enum із domain.Levels.
ФАКТ: C — overlay містить точні operationId/summary/description/204-description з frozen та шість maxLength; shape/enum/required і 400 не дубльовано.
ФАКТ: D — clientlog додано 10-м contributor у leaf collector.
ФАКТ: F — додано HandleFunc AST-census fail-closed, fidelity з ref-closure ClientLogRequest/ErrorResponse + property-removal probe і double-generation, domain_test census declared constants == ordered Levels == contribution enum (non-empty guard).
ФАКТ: gofmt -w touched Go files exit 0. Перший UPDATE-тест exit 1 до компіляції: sandbox відмовив у доступі до ~/Library/Caches/go-build. Запис цього факту з backend cwd не відбувся (FileNotFoundError); тут записано в правильний файл.
ДАЛІ: UPDATE запущено з GOCACHE=/tmp/courses-s01k-go-build, без зміни permissions.
ФАКТ: UPDATE із tmp-кешем exit 1: orphan overlay key через відступ нового path (2 замість існуючих 4). Виправлено лише відступ overlay; це помилка вставки, не обмеження engine.
ФАКТ: UPDATE_OPENAPI_GENERATED=1 go test -run TestOpenAPIArtifact ./internal/openapispec із tmp GOCACHE exit 0; artifact регенеровано. AI Factory checkpoint exit 0, ID 5835573fff55af781634e99ca4861edb-1-1471e7559b34.
ФАКТ: go test -run TestOpenAPIArtifact ./internal/openapispec — exit 0; evidence /tmp/courses-s01k-gates.log (GOCACHE=/tmp/courses-s01k-go-build).
ФАКТ: gofmt -l . — exit 0; evidence /tmp/courses-s01k-gates.log (GOCACHE=/tmp/courses-s01k-go-build). gofmt output empty.
ФАКТ: go vet ./... — exit 0; evidence /tmp/courses-s01k-gates.log (GOCACHE=/tmp/courses-s01k-go-build).
ФАКТ: self-review: generated diff лише ClientLogRequest + /api/v1/client-logs; frozen openapi.yaml, generate.go, overlay.go byte-identical до HEAD. git diff --check exit 0. Scoped SHA-256 snapshot: /tmp/courses-s01k-snapshot.json. Чинні domain/handler tests перевіряють normalize, malformed/unknown/duplicate body, 204 та emitted/scrubbed log.
ФАКТ: go build ./... — exit 0; evidence /tmp/courses-s01k-gates.log (GOCACHE=/tmp/courses-s01k-go-build).
ФАКТ: go test ./internal/clientlog/... ./internal/openapispec/... ./internal/platform/openapigen/... — exit 0; evidence /tmp/courses-s01k-gates.log (GOCACHE=/tmp/courses-s01k-go-build).
ФАКТ: усі дозволені фінальні гейти PASS: UPDATE→plain, gofmt clean, vet, build, targeted tests (4 packages). Snapshot hashes після гейтів незмінні. Fidelity + property-removal probe + two-generation identity зелені; empty→error/validation/logging збережено.
РІШЕННЯ: обмеження з брифа збережено: parallel artifact не підключено до runtime; message required — frozen documentation choice; maxLength у overlay не є автоматичним drift-proof доменних caps. Full go test ./..., integration/e2e тут NOT_RUN за дорученням. Комітів/push немає.
ДАЛІ: Courses 1 — host full suite і передача exact scoped snapshot Courses 2 (existing_session); незалежний review NOT_TESTED. Evidence: /tmp/courses-s01k-gates.log, /tmp/courses-s01k-snapshot.json; HEAD 35f886fa531273bc0b8d96f7d1e3468279ae9d72 + scoped WIP.
ФАКТ: фінальний AI Factory checkpoint exit 0, ID 5835573fff55af781634e99ca4861edb-2-66a528304445.

## Host-verify (Courses 1, НЕ довіряю Codex exit 0)

- ФАКТ: frozen `openapi.yaml` + `generate.go` + `overlay.go` byte-identical до HEAD (`git diff --stat HEAD` порожній).
- ФАКТ: `go test -run TestOpenAPIArtifact ./internal/openapispec` plain (без UPDATE) exit 0 — checked-in артефакт == регенерація.
- ФАКТ: `gofmt -l .` порожньо; `go vet ./...` 0; `go build ./...` 0.
- ФАКТ: повний `go test ./...` на ХОСТІ exit 0, 0 FAIL; clientlog(http+domain)/openapispec/openapigen зелені. Log scratchpad/s0.1k-hostgate.log.
- ФАКТ: generated diff = лише ClientLogRequest (required:[message], level enum [error,warn,info], 6×maxLength, additionalProperties:false) + /api/v1/client-logs (operationId/requestBody.required/204 bodyless/400 BadRequest $ref) — збігається з frozen hand.
- ФАКТ: тести non-vacuous — census 3-стороння рівність; coverage HandleFunc AST fail-closed; fidelity з message-removal probe + two-gen byte-identical.
- РІШЕННЯ (e2e NOT_RUN): паралельний артефакт не wired у runtime; handler change (omitempty+named Level+string()) behaviour-preserving, покрито clientlog unit/contract. Прецедент playback S0.1f.
- ФАКТ: закомічено `03d42c8` (10 файлів explicit-paths; _wip-журнали НЕ чіпав). Покриття 79/85.
- transfer: `sent`→`acknowledged` Courses 2 [e6a45d] (msg 4b3ba762). Але e6a45d оверсайз/загейчена;
  власник указав передати рев'ю іншій сесії. RE-ROUTED → свіжа **courses-96 [557618]** (msg 6d2aa2aa,
  повний пакет), e6a45d попрошено стати на паузу (msg 53d953d6). Чекаю ACK+verdict courses-96 проти
  03d42c8 (ще не підтверджено, що це рев'ювер — спитав у першому меседжі). Не мерджу.
- transfer=`acknowledged`: courses-96 [557618] ПІДТВЕРДЖЕНО власником як Courses 2 рев'ювер цього
  снапшота (review_mode=independent_subagent→ця сесія, read-only git show/diff). ACK≠PASS; вердикт
  наступним. Reviewer для S0.1k = courses-96, не e6a45d. Чекаю PASS/findings проти 03d42c8.
- transfer=`verdict_received`: **PASS** (courses-96 [557618], read-only + локальний прогін, дерево не
  чіпав). Перевірено: коректність (omitempty=marshal-only, decode-only struct → 0 рантайм; slices.Contains
  семантично = switch), fidelity (closure == frozen, probe non-vacuous, two-gen), artifact freshness,
  coverage AST, enum 3-way census, scope (frozen+движок не торкнуті). Обидва мої рішення схвалено.
  1 non-blocking (НЕ умова мерджу): maxLength caps триплікуються domain/hand/overlay без overlay↔domain
  census (level-enum для контрасту single-sourced) — кандидат S0.2; зараз копії збігаються, споживача нема.
- S0.1k ЗАКРИТО на гілці. У main НЕ мерджу (S0.2 cutover, owner-PAUSE). Коміт 03d42c8 на s0.0-spike.
