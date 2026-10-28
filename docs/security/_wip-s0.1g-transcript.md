# S0.1g — transcript as 6th OpenAPI contributor (writer journal)

Working file for the Codex run. Append after every step (ДАЛІ before each tool; then ФАКТ /
РІШЕННЯ / ДАЛІ / ПИТАННЯ / STOP). Writer = Courses 13 (commits + host-verify). Reviewer = Courses 2.
Branch s0.0-spike. Do NOT commit/stage/touch git. Do NOT edit the frozen backend/openapi/openapi.yaml.

## Journal
ДАЛІ: прочитати briefing, AI Factory adapter, media/playback contracts і frozen transcript spec.
ФАКТ: 5 contributors у contributions.go; прочитано generator та зразки media/playback (cat).
ДАЛІ: звірити 8 операцій, DTO і nullable поля з frozen spec.
ФАКТ: 8 операцій / 7 шляхів у frozen spec (sed -n 1501,1778p); GET transcript без 409, draft без 500.
РІШЕННЯ: review → наявна окрема сесія Courses 13; owner/delivery owner writer; ручна передача за дорученням користувача; локально лише author host-safe checks, без git.
ДАЛІ: прочитати referenced schemas, overlay merge policy і hook configuration.
ФАКТ: 1456 рядків baseline openapi.generated.yaml (wc -l); overlay підтримує inline через $ref:null та optional requestBody.
ДАЛІ: підготувати точні DTO/NameMap і overlay metadata без frozen edits.
ФАКТ: 8 named schemas: LessonTranscript, LessonDraft, SaveTranscriptRequest, PublishLessonDraftRequest, RegenerateLessonDraftRequest, TranscriptDocumentLink, TranscriptSegment, DraftSegment (sed 2452,2612).
РІШЕННЯ: 2026-10-28 Codex; зберегти json tags/decoder behavior, optional required arrays і DraftSegment omitted pointers відобразити overlay; unnamed result/save segment schemas inline через $ref:null. Перевірка: generated vs frozen structural comparison.
ДАЛІ: додати contribution та named handler results.
ФАКТ: 8 route bindings, 8 NameMap entries, 5 new response metadata blocks, 7 path blocks додано (python3 /tmp/courses-s01g-edit.py; gofmt -w), exit 0. Frozen spec прочитано через pinned yaml3; PyYAML відсутній, залежностей не додано.
ДАЛІ: regenerate artifact і перевірити generator acceptance.
ФАКТ: initial artifact generation exit 1: Go default cache operation not permitted; retry запущено з GOCACHE=/tmp/courses-s01g-go-18473 (session 40961).
ФАКТ: 1 journal append не виконався через backend cwd; цей запис відновлює пропущений крок.
ДАЛІ: дочекатись retry; перевірити generated schemas проти frozen.
ФАКТ: current native identity підтверджена host event; exact courses/codex policy row знайдено; dispatcher checkpoint спробовано за operator mapping.
ДАЛІ: artifact generation retry й scoped gates.
ФАКТ: regenerate з GOCACHE=/tmp/courses-s01g-go-18473 PASS exit 0: ok courses/internal/openapispec 2.057s.
ФАКТ: checkpoint BLOCKED exit 1 до dispatcher: mktemp receipt parent Operation not permitted; локальний журнал збережено.
ДАЛІ: artifact test без UPDATE, gofmt/vet/build/scoped tests.
ФАКТ: artifact TestOpenAPIArtifact робить 2 генерації та bytes.Equal; окремо перевіряється checked artifact (cat artifact_test.go).
ДАЛІ: отримати gate exits та structural comparison.
ФАКТ: artifact test без UPDATE PASS: ok courses/internal/openapispec 1.546s; gofmt -l . порожній stdout; go vet ./... exit 0 (session 89515).
ФАКТ: initial comparison script мав TypeError на boolean required; виправлено нормалізацію лише array required, повторено /tmp/courses-s01g-verify.py.
ДАЛІ: go build ./... та scoped go test.
ФАКТ: GOCACHE=/tmp/courses-s01g-go-18473 go build ./...; exit 0
(empty output)
ДАЛІ: scoped tests
ФАКТ: structural comparison знайшов лише omitted required:false у serializer для optional requestBody; нормалізовано OpenAPI default false без зміни продуктового коду.
ДАЛІ: підтвердити scoped gate output та фінальний запис.
ФАКТ: GOCACHE=/tmp/courses-s01g-go-18473 go test -count=1 ./internal/transcript/... ./internal/openapispec/... ./internal/platform/openapigen/...; exit 0
ok  	courses/internal/transcript/adapters/asynq	1.473s
ok  	courses/internal/transcript/adapters/http	2.679s
ok  	courses/internal/transcript/adapters/media	2.372s
?   	courses/internal/transcript/adapters/postgres	[no test files]
ok  	courses/internal/transcript/adapters/whisper	6.901s
ok  	courses/internal/transcript/application	2.353s
ok  	courses/internal/transcript/domain	2.960s
?   	courses/internal/transcript/ports	[no test files]
ok  	courses/internal/openapispec	3.485s
ok  	courses/internal/platform/openapigen	3.893s
ДАЛІ: final artifact count and writer handoff
ФАКТ: python3 /tmp/courses-s01g-verify.py PASS: 8 named schemas + 8 complete operations equal frozen spec; normalization: required arrays, parameter order, omitted required:false.
ФАКТ: wc -l backend/openapi/openapi.generated.yaml = 2117 рядків; baseline 1456, delta +661 рядок.
ДАЛІ: scoped tests completion; final read-only git status explicitly requested by writer, без інших git operations.
ФАКТ: final git status --short (read-only, explicitly requested), exit 0:
 M backend/internal/openapispec/contributions.go
 M backend/internal/transcript/adapters/http/admin_handler.go
 M backend/internal/transcript/adapters/http/draft_handler.go
 M backend/internal/transcript/domain/draft.go
 M backend/openapi/openapi.generated.yaml
 M backend/openapi/openapi.overlay.yaml
 M docs/security/_wip-s0.1b-sec-errorboundary.md
 M docs/security/_wip-s0.1b2-request.md
 M docs/security/_wip-s0.1b3a-errorcode.md
 M docs/security/_wip-s0.1b3b-constset.md
?? backend/internal/transcript/adapters/http/openapi_contract.go
?? docs/security/_wip-s0.1g-transcript.md
ФАКТ: generated.yaml 1456 → 2117 рядків (+661), wc -l. Frozen openapi.yaml не редагувався; pre-existing s0.1b journals не редагувалися.
ФАКТ: real gate outputs: gofmt -l . — empty stdout; go vet ./... — empty stdout, exit 0; go build ./... — empty stdout, exit 0; scoped go test -count=1 ./internal/transcript/... ./internal/openapispec/... ./internal/platform/openapigen/... — exit 0, 8 ok packages, 2 packages [no test files] (повний output вище). Artifact generation UPDATE=1 — ok 2.057s; без UPDATE — ok 1.546s, exit 0; кожен run перевіряє 2 byte-identical generations та checked artifact. GOCACHE=/tmp/courses-s01g-go-18473 після default cache permission failure.
ФАКТ: 8/8 routes завершено; structural comparison 8 schemas + 8 operations PASS. Повний go test, integration/e2e не запускалися. Writer host verification pending.
ФАКТ: safe-boundary checkpoint refresh BLOCKED: mktemp у approved receipt parent знову Operation not permitted, exit 1; current host identity відома, policy row знайдена. Dispatcher не запускався, Docker не запускався; durable handoff — цей журнал.
ДАЛІ: leave for writer host-verify
