# S0.1b-3a — error-code parser (§16.1) — WORK FILE

Дописуй ПІСЛЯ КОЖНОГО КРОКУ (append, не rewrite). Перший запис — рядок `ДАЛІ:` з тим, що
збираєшся перевіряти, ДО першого інструменту. Далі ФАКТ / РІШЕННЯ / ДАЛІ / ПИТАННЯ.
Фінал = стислий переказ. Base: branch `s0.0-spike`, HEAD наданий у brief.

## Журнал
ДАЛІ: Прочитати briefing, план §16.1 повністю, §16.8 Q2a, §16.9 і код; звірити контракт до змін реалізації. За неоднозначності STOP. Scope лише error-code enum.
ФАКТ: Початковий намір записано; briefing прочитано, це історичний кеш; поточний контракт звіряю з планом і кодом.
ФАКТ: HEAD 0149d9c; наявні чужі зміни docs/security/_wip-s0.1b2-request.md і work/tasks.md — зберігаю. Trailers попередніх комітів прочитано. Review: ручна передача writer Courses 13 → Courses 2; незалежний review pending.
ФАКТ: §16.1 повністю, §16.8 Q2a і §16.9 прочитано; parser має покрити 4 форми, overlay enum видаляється лише разом із Go-джерелом і fidelity.
ДАЛІ: Звірити повну множину кодів із hand enum та реальні WriteError call sites до мутацій.
ФАКТ: Підтверджено typed const у восьми названих модулях і errors.New у discussion. Реальні call sites містять також code (discussion) й override (identity); читаю їх визначення. Шлях openapi лежить не в root — уточнюю inventory.
ФАКТ: OpenAPI у backend/openapi. Discussion code := err.Error(); identity override звужено guard-ом origin_forbidden. Native identity підтверджено host; exact Courses/codex policy row знайдено.
ФАКТ: Pre-code BLOCKED, branch s0.0-spike, HEAD 0149d9c. Read-only Python census (glob backend/internal/*/domain/errors.go; regex для Error = literal та errors.New literal; звірка з ErrorResponse у hand YAML) дав 99 унікальних доменних кодів; 20 відсутні в backend/openapi/openapi.yaml:3409:
asr_draft_superseded, course_not_purchasable, grant_invalid, image_not_found, image_too_large, image_too_many_pixels, invalid_image, invalid_progress, invalid_session, invalid_state_transition, invalid_user, learner_query_invalid, money_invalid, order_invalid, order_status_transition_invalid, payment_event_unmatched, roster_cursor_invalid, two_factor_required, unsupported_image, unsupported_source.
ФАКТ: Це не лише внутрішні коди: backend/internal/image/adapters/http/handler.go:112–118 реально передає unsupported_image, image_too_large, image_too_many_pixels, invalid_image у WriteError. Hand ErrorResponse їх не містить. Fidelity backend/internal/course/adapters/http/openapi_contract_test.go:30–36,72–79 порівнює повне ref-closure, отже додавання цих enum members змінить порівнювану поверхню.
РІШЕННЯ: 2026-10-27, Codex, pre-code STOP за явною інструкцією власника. Вимоги «кожен доменний код у emitted enum», «hand YAML незмінний» і «exact fidelity зелена» несумісні на HEAD. Перевірка: наведений census + hand enum :3409 + image handler :112–118. Не обмежую enum перетином із hand, не послаблюю fidelity і не змінюю API.
ФАКТ: Реалізацію, overlay і generated artifact не змінено; gofmt/vet/build/test НЕ ЗАПУСКАЛИСЯ, статус NOT_TESTED. Комітів/merge/push немає. Чужі WIP і work/tasks.md не змінював.
ПИТАННЯ: Яку вимогу уточнюємо: окремо дозволяємо виправити hand enum перед цим зрізом, чи переглядаємо вимогу повного доменного enum? Рекомендація — спершу узгодити окремий контрактний зріз; просте виключення внутрішніх кодів не усуває чотири реально емітовані image-коди.
ДАЛІ: Отримати рішення writer Courses 13; після узгодження повернутися до parser/coverage/overlay removal/regen й ordered gates. Незалежний review Courses 2 лишається pending.
ФАКТ: Native checkpoint спроба завершилась exit 2: mktemp у затвердженому receipt parent → Operation not permitted; dispatcher → BLOCKED required handoff. Identity актуальна, policy row знайдено; durable checkpoint не створено, робочий файл збережено.

## WRITER VERIFY OF CODEX STOP (Courses 13, 2026-10-27)
ПІДТВЕРДЖЕНО власною перевіркою: image/adapters/http/handler.go:110-118 (writeUploadError) емітує 4 коди (unsupported_image/image_too_large/image_too_many_pixels/invalid_image) через httpserver.WriteError; hand ErrorResponse enum openapi.yaml:3409 їх НЕ містить. Знахідка реальна — контракт неповний відносно емітованих клієнту кодів. Codex STOP коректний.
Конфлікт: «повний доменний enum з Go» + «hand openapi.yaml незмінний» + «exact fidelity зелена» несумісні на HEAD.
Ескальовано власнику (writer не ухвалює: правка source-of-truth + переінтерпретація §16.1). Codex нічого не комітив; код/overlay/generated не змінені; гейти NOT_RUN. S0.1b-3a BLOCKED до рішення власника.
ДАЛІ: Продовжити той самий S0.1b-3a за рішенням власника §16.11: emitted-set, дозволена правка тільки hand error enum. Спершу перечитати рішення та звірити source-universe membership із call sites до коду.
ФАКТ: §16.11 прочитано повністю; hand enum correction дозволена власником. Однак вимога every resolved code ∈ */domain/errors.go source universe має окрему суперечність: course коди оголошені в domain/course.go, discussion частково в domain/comment.go, а method_not_allowed є transport literal. Перевіряю точний список прямих literal-викликів, які не належать заданому glob.
ФАКТ: Read-only перевірка прямих bare-literal WriteError (не повний RESOLVE census) знайшла 19 кодів поза universe заданого glob: course_not_deletable, course_not_found, document_conflict, insufficient_storage, invalid_callback, invalid_lesson_document, invalid_owner, invalid_query, invalid_slug, invalid_status_transition, lesson_limit, method_not_allowed, playback_forbidden, precondition_failed, request_too_large, slug_immutable, slug_taken, unsupported_media_type, upload_offset_conflict.
ФАКТ: Конкретні контрприклади membership-вимоги: course/adapters/http/handler.go:184 емітить invalid_query, визначення в course/domain/course.go:16 (не errors.go); platform/httpserver/server.go:171 емітить method_not_allowed як transport literal, цього коду немає в жодному */domain/errors.go. Discussion/domain/comment.go:16–20 містить ще п’ять errors.New поза заданим glob. Це окрема проблема source-universe, не попередня проблема hand enum, яку §16.11 вже вирішив.
РІШЕННЯ: 2026-10-27, Codex — pre-code STOP на суперечності §16.11 membership із реальними call sites. Не створювати штучні доменні константи для transport-кодів і не оголошувати literal автоматично членом вузького universe без уточнення контракту. Реалізацію, hand та overlay не змінено; gates NOT_TESTED; commit/merge/push не виконано.
ПИТАННЯ: Чи уточнюємо source universe як об’єднання error-декларацій у всіх production Go-файлах відповідних domain/application пакетів і доведених transport literals/guard values з WriteError-шляхів (із provenance), з обов’язковим включенням */domain/errors.go? Тоді кожен resolved code перевіряється проти цього розширеного universe; невідомі expression/flow форми лишаються fail-closed. Альтернатива — окреме погоджене перенесення декларацій, яке саме по собі не розв’язує transport literals без додавання доменних сутностей.
ДАЛІ: Чекаю уточнення source-universe контракту від Courses 13; emitted set та hand additions/removals ще НЕ обчислені повним resolver-ом, hand правок немає. Попередня блокада native checkpoint файловими permissions лишається незнятою; current session 01a17ecd-6da2-76c3-8e95-a9493899cd35 підтверджена host.
ДАЛІ: Продовжити за §16.12; domain declarations у всіх production domain Go-файлах + transport allowlist із provenance, без невикористаних записів. Механізм вирішую автономно, семантика emitted-set незмінна.
ФАКТ: §16.12 прочитано. Small route: parser → tests → census → hand enum only + overlay removal → regen → ordered gates. Review передає Courses 13 до Courses 2. Динамічний discussion err.Error() потребує окремого явного flow-доказу; не підмінюю його всіма доменними кодами.
ДАЛІ: Дочитати discussion error flow перед вибором механізму err.Error(); Generate викликається лише в generator/course тестах, тож source input додам явно, без runtime I/O у модулі course.
РІШЕННЯ: 2026-10-27, Codex. Source parser використовує go/parser AST, import provenance errors.New, typed Error const; всі production domain .go під internal/*/domain/**, без тестів. Невідома Err-декларація fail-closed, sort/status/BlockType не обробляються. Call-site scanner має відокремити source facts від emission facts; hand не редагується до завершення resolver coverage.
ФАКТ: При трасуванні dynamic err.Error() виявлено persistenceError(err): discussion/adapters/postgres/repository.go:367–372 повертає ОРИГІНАЛЬНИЙ err для errors.Is(context.Canceled/DeadlineExceeded), не стабільну доменну помилку. Перевірю до реалізації, чи це відкриває нефіксовані wire code значення; це може бути runtime-контракт, а не parser-механізм.
ДАЛІ: Перевірити cancellation production-шлях без БД/мережі: pgxpool із заздалегідь скасованим context → реальний discussion Repository.LessonFacts; потім application/HTTP мапінг. Не підміняти код помилки fake-ом.
ФАКТ: persistenceError не має cancellation-тестів у discussion. Public GET /courses/{slug}/reviews дозволяє пройти справжній handler→application→pgx repository без auth fixture. Підготую тимчасовий probe з забороненим DialFunc, щоб довести шлях без зовнішнього I/O; probe не є gate зрізу.
ФАКТ: PublicReviews повертає repository error без нормалізації, writeReviewError ремапить тільки domain.ErrNotFound, writeError серіалізує err.Error(). Перевіряю, чи cancellation — лише два стабільні значення, чи відкритий wrapped-error текст.
ДАЛІ: Діагностичний probe (не gate): реальні pgxpool→discussion repository→application→HTTP; DialFunc не відкриває socket, очікує власний реальний connect deadline драйвера й повертає ctx.Err(). Це контрольований transport seam, не live DB доказ; перевіряє production форматування/пропуск wrapped error. Дві синтетичні database names перевірять, чи code залежить від конфігурації.
ФАКТ: Команда cd backend && go run ./cmd/s013a-error-probe запущена; виконується, результат ще не отримано.
ФАКТ: Діагностична команда `cd backend && go run ./cmd/s013a-error-probe`, EXIT=0. Реальний вивід:
```text
database=probe_a status=400 body={"error":{"code":"failed to connect to `user=probe database=probe_a`: 127.0.0.1:1 (127.0.0.1): dial error: timeout: context deadline exceeded","message":"Request is not valid","requestId":""}}
database=probe_b status=400 body={"error":{"code":"failed to connect to `user=probe database=probe_b`: 127.0.0.1:1 (127.0.0.1): dial error: timeout: context deadline exceeded","message":"Request is not valid","requestId":""}}
```
ФАКТ: Це НЕ live DB/e2e proof: контрольований DialFunc, без socket. Але pgxpool формує реальний ConnectError, production repository пропускає його за errors.Is(context.DeadlineExceeded), production service і HTTP серіалізують повний текст. Немає fake repository/service або підставленого error.code. requestId порожній, бо probe використовує module mux без platform middleware; це не причина dynamic code.
ФАКТ: Шлях: discussion/adapters/postgres/review.go:34–40 → repository.go:367–372 (повертає wrapped err без зміни) → application/application.go:399–403 → adapters/http/handler.go:334–336 → :681–686 → :712,:755 → platform/httpserver/server.go:55–56. Pgx v5.10.0 pgconn/errors.go ConnectError.Error() включає Config.User/Database; errors.Is через Unwrap зберігає deadline-причину. Source universe/allowlist не може перелічити конфігураційно залежні рядки.
РІШЕННЯ: 2026-10-27, Codex — STOP #3 до parser/hand мутацій. Це реальна суперечність скінченного emitted-set enum з dynamic raw-error emission, а не нова назва файлу чи форма AST. Не додаю довільні database/error strings в allowlist і не підмінюю .Error() всіма domain codes. Runtime-нормалізація не була дозволена початковим scope.
ПИТАННЯ: Дозволити окремий runtime-fix у discussion HTTP boundary: невідомі/інфраструктурні помилки (зокрема wrapped cancellation/deadline) → стабільний dependency_unavailable (503), зі збереженням поточних domain/application кодів? Або визначити інший стабільний код/статус. Після цього повернутися до S0.1b-3a. Рекомендую boundary-нормалізацію, щоб також закрити нові непередбачені error strings у майбутньому.
ФАКТ: Ordered gates не запускались (NOT_TESTED); probe — окрема діагностика. Hand/overlay/generated/runtime не змінено. Нижче повний тимчасовий probe; файл cmd після збереження прибираю. Для відтворення: зберегти цей блок у backend/cmd/s013a-error-probe/main.go, виконати зазначену команду, прибрати probe.
```go
package main

import (
 "context"
 "fmt"
 "net"
 "net/http"
 "net/http/httptest"
 "time"
 discussionhttp "courses/internal/discussion/adapters/http"
 "courses/internal/discussion/adapters/postgres"
 "courses/internal/discussion/application"
 "github.com/jackc/pgx/v5/pgxpool"
)

func main() {
 for _, database := range []string{"probe_a", "probe_b"} {
  config, err := pgxpool.ParseConfig("postgres://probe@127.0.0.1:1/"+database+"?sslmode=disable")
  if err != nil { panic(err) }
  config.MaxConns = 1
  config.ConnConfig.ConnectTimeout = time.Nanosecond
  // No socket is opened: wait for the driver's own real connect deadline.
  config.ConnConfig.DialFunc = func(ctx context.Context, _, _ string) (net.Conn, error) {
   <-ctx.Done()
   return nil, ctx.Err()
  }
  pool, err := pgxpool.NewWithConfig(context.Background(), config)
  if err != nil { panic(err) }
  mux := http.NewServeMux()
  identity := func(h http.Handler) http.Handler { return h }
  discussionhttp.NewHandler(application.New(postgres.NewRepository(pool),time.Now)).Register(mux,identity,identity)
  rec := httptest.NewRecorder()
  mux.ServeHTTP(rec,httptest.NewRequest(http.MethodGet,"/api/v1/courses/probe/reviews",nil))
  fmt.Printf("database=%s status=%d body=%s", database,rec.Code,rec.Body.String())
  pool.Close()
 }
}
```
ДАЛІ: Передати Courses 13 concrete runtime finding; чекати рішення щодо нормалізації. S0.1b-3a не завершено, independent review pending. Поточний session id: 01a17ecd-6da2-76c3-8e95-a9493899cd35.
ФАКТ: Native checkpoint через policy-approved dispatcher успішний, EXIT=0: {"status":"OK","result":{"checkpoint_id":"5c982175f6c3a95ad10ac5c8017327af-1-f1da3f090fe8"}}. Попередню файлову блокаду checkpoint знято escalation; explicit handoff ID збережено.

## WRITER VERIFY OF CODEX STOP #3 (Courses 13, 2026-10-27)
ПІДТВЕРДЖЕНО: discussion/adapters/http/handler.go:711-712 `writeError` робить `code := err.Error()` і передає в httpserver.WriteError. statusFor() default → 400. Невідома/інфраструктурна помилка (wrapped pgx ConnectError, user/db у рядку) → клієнту error.code = сирий рядок, HTTP 400. Реальний security-витік + невірний статус + unbounded enum. Probe Codex (probe_a/probe_b → різні коди) достовірний.
Патерн: ЄДИНИЙ production `code := err.Error()` — discussion:712. Інші модулі мають власні writeError-обгортки (course/admin/document/learning, adminstats, identity, orders, media, image, playback) — НЕ аудитовані на той самий default-raw-err клас. Дефект уже зафіксований у S0.0 spike (course/adapters/http/s0_spike_errcode_test.go, B1 FINDING).
Ескальовано власнику: runtime/security behavior change поза scope 3a. Codex нічого не комітив (STOP #3); hand/overlay/generated/runtime не змінені; гейти NOT_RUN. Codex session 01a17ecd-6da2-76c3-8e95-a9493899cd35. S0.1b-3a BLOCKED.
ДАЛІ: Відновити S0.1b-3a після S0.1b-sec; звірити HEAD d12b2c7, §16.11/16.12 та нові call sites. Старий census не використовувати. make e2e заборонено власником для цього прогону.
ФАКТ: HEAD d12b2c7 підтверджено; discussion тепер finite switch code literals + writeErrorCode(code), identity має guarded override. Чужі WIP/plan/tasks зберігаю. READY: контракт узгоджений, runtime-змін не потрібно.
РІШЕННЯ: 2026-10-27, Codex — bounded AST resolver: literals, import-resolved domain symbols, string cast, finite Error() values; для forwarding-параметрів обхід усіх production callers, для local variable — single definition або exhaustive switch. Guard identity override == literal звужує значення на call site. Невідома/неоднозначна dataflow форма → error, не union усіх декларацій. Перевірка негативними fixtures (unknown expression, assignment, caller, unbounded Error()).
ФАКТ: Додано errors.go: AST domain source + fail-closed call-site resolver з finite locals, switches, wrapper callers і guarded overrides. Це ще не перевірено, hand untouched.
ФАКТ: Додано fixtures для чотирьох форм, wrapper/switch/guard, негативні unknown/cycle/escape/assignment cases, provenance і production census. Development test запущено; це ще не ordered gate. Попередній regex-inventory лише допоміжний (знайшов стороннє media:transcode), source truth визначає AST.
ФАКТ: Development tests спершу FAIL: cycle guard плутав AST nodes зі спільною Pos (CallExpr та receiver), плюс redundant allowlist comment_not_found уже має orders domain provenance. Виправлено node-identity key і вилучено redundant запис.
ФАКТ: Додано GenerateFromSource: census обов’язковий перед ErrorDetail generation, enum вставляється у reflected Code schema. Pure Generate лишається для unit fixtures; ErrorDetail без source census відхиляється. Course test callers перейшли на явний backendRoot; runtime course handler не змінено.
ДАЛІ: Отримати повторний development census, посилити adversarial fixtures перед hand diff. Ordered gates ще не виконувались.
ФАКТ: Development census PASS: 166 WriteError sites, 116 domain codes, 109 emitted codes. Це попередній результат перед adversarial hardening; hand ще не редагував.
ФАКТ: Додано Error() receiver context (finite sentinel через variable/forwarder), відхилення raw error object як code і address escape; Error underlying type перевіряється як defined string. Overlay не може повторно задавати error-code enum. Конкретні flow обмеження fail-closed, не silent omissions.
ДАЛІ: Перевірити повторний census після hardening; якщо PASS, записати exact emitted set і hand diff перед enum edit.
ФАКТ: Повторний development run EXIT=0: усі форми/негативні кейси PASS, census стабільно 166 sites / 116 domain / 109 emitted. Додаю незалежний AST count WriteError calls та тест unused domain code (source universe НЕ розширює enum), unknown/computed declaration fail-closed.
ДАЛІ: Чекаю development coverage/source tests перед фіксацією audit census.
ФАКТ: AUDIT ДО РЕДАГУВАННЯ HAND (HEAD d12b2c7; source files runtime unchanged). Команда development proof: `cd backend && go test -count=1 -run 'TestErrorCode|TestErrorSourceDiscovery|TestTransportProvenance|TestTreeErrorCensus' -v ./internal/platform/openapigen`, EXIT=0. Незалежний AST count = resolved census = 166 WriteError calls; source domain universe 116 unique; emitted 109 unique.
ФАКТ: EXACT EMITTED SET (sorted):
admin_immutable, admin_not_learner, already_running, authentication_required, cannot_target_self, checkout_already_opened, comment_body_invalid, comment_cursor_invalid, comment_forbidden, comment_not_found, comment_parent_invalid, comment_position_invalid, comment_removed, comment_thread_full, course_already_owned, course_forbidden, course_modified, course_not_deletable, course_not_enrollable, course_not_found, course_purchase_required, dependency_unavailable, document_conflict, email_not_verified, email_taken, enrollment_blocked, enrollment_forbidden, enrollment_required, forbidden, grant_forbidden, grant_self_forbidden, grant_transition_invalid, image_too_large, image_too_many_pixels, insufficient_storage, internal_error, invalid_account_token, invalid_callback, invalid_credentials, invalid_display_name, invalid_email, invalid_image, invalid_lesson_document, invalid_owner, invalid_query, invalid_range, invalid_request, invalid_slug, invalid_status_transition, invalid_totp, last_admin, learner_course_forbidden, lesson_limit, lesson_not_found, membership_conflict, method_not_allowed, not_found, order_not_found, origin_forbidden, password_unchanged, payment_event_rejected, playback_forbidden, playback_session_expired, playback_taken_over, precondition_failed, range_too_wide, rate_limited, recovery_code_invalid, report_already_open, report_forbidden, report_invalid, report_own_comment, report_own_person, report_own_review, request_too_large, review_already_exists, review_invalid, review_moderation_forbidden, review_not_available, review_not_found, review_not_started, review_owner_forbidden, review_purchase_required, review_removed, slug_immutable, slug_taken, token_used, transcript_document_exists, transcript_document_limit, transcript_document_referenced, transcript_exists, transcript_modified, transcript_segment_limit, transcript_text_limit, twofactor_already_enabled, twofactor_not_enabled, unknown_bucket, unknown_metric, unsupported_image, unsupported_media, unsupported_media_type, upload_in_progress, upload_offset_conflict, upload_too_large, user_not_found, video_not_found, video_not_ready, video_not_retryable, weak_password
ФАКТ: HAND ДОДАВАННЯ (emitted \ hand), 9: image_too_large, image_too_many_pixels, insufficient_storage, invalid_image, precondition_failed, request_too_large, unsupported_image, unsupported_media_type, upload_offset_conflict
ФАКТ: HAND ВИЛУЧЕННЯ (hand \ emitted), 0: немає
ФАКТ: internal_error і dependency_unavailable лишаються в emitted-set. Hand видалень немає; чотири image-коди й п’ять tus transport-кодів додаються.
ФАКТ: Transport allowlist із provenance:
package openapigen

// TransportErrorCodes documents codes without a domain declaration. Values name
// the actual WriteError emission file, including the terminal site of a wrapper.
// ScanErrorCodes rejects missing, stale, redundant, and unused entries.
func TransportErrorCodes() map[string]string {
	return map[string]string{
		"method_not_allowed":     "platform/httpserver/server.go",
		"precondition_failed":    "internal/media/adapters/tus/server.go",
		"request_too_large":      "internal/media/adapters/tus/server.go",
		"unsupported_media_type": "internal/media/adapters/tus/server.go",
		"upload_offset_conflict": "internal/media/adapters/tus/server.go",
		"insufficient_storage":   "internal/media/adapters/tus/server.go",
		"invalid_callback":       "internal/orders/adapters/http/handler.go",
		"playback_forbidden":     "internal/playback/adapters/http/handler.go",
		"comment_forbidden":      "internal/discussion/adapters/http/handler.go",
		"comment_removed":        "internal/discussion/adapters/http/handler.go",
	}
}

ДАЛІ: Змінити тільки hand ErrorResponse.error.code.enum на sorted emitted set; видалити overlay code.enum; regenerate parallel artifact і перевірити fidelity.
ФАКТ: Після audit-запису виправлено hand тільки ErrorResponse.error.code.enum (+9, -0, sorted). Overlay code.enum видалено; $ref inlining і requestId validation збережено. Регенерація запускається окремо від ordered gates.
ФАКТ: Regen command `cd backend && UPDATE_OPENAPI_GENERATED=1 go test -count=1 -run TestCatalogOpenAPIFidelity -v ./internal/course/adapters/http` EXIT=1 setup failed: Go build cache `/Users/strelia/Library/Caches/go-build/...` operation not permitted. Це sandbox/cache блокада, не результат fidelity. Повторю із дозволом на Go cache.
ФАКТ: Regen повторно запущено з cache-дозволом; очікую реальний результат.
ФАКТ: Regen повтор EXIT=0. Реальний вивід:
```text
=== RUN   TestCatalogOpenAPIFidelity
    openapi_contract_test.go:68: catalog and status request + complete referenced component closures equal hand spec; two generations byte-identical
--- PASS: TestCatalogOpenAPIFidelity (0.23s)
PASS
ok  courses/internal/course/adapters/http 1.583s
```
ДАЛІ: Додати негативну non-vacuous enum fidelity перевірку і overlay rejection/source-required fixtures; прибрати непотрібний function index; потім фінальні ordered gates.
ФАКТ: Додано non-vacuous fidelity mutation (видалення enum member мусить дати diff), source-required і overlay-enum rejection tests. Hand git diff: змінено лише один enum рядок; generated diff — порядок enum +9 members. ДАЛІ: ordered gofmt → vet → build → tests, реальний повний вивід append у цей файл.

ДАЛІ: ORDERED GATE, cwd=backend: `gofmt -l internal/platform/openapigen/errors.go internal/platform/openapigen/errors_test.go internal/platform/openapigen/error_transport.go internal/platform/openapigen/generate.go internal/platform/openapigen/overlay.go internal/course/adapters/http/openapi_contract_test.go internal/course/adapters/http/openapi_generated_request_test.go internal/course/adapters/http/openapi_generated_response_test.go`
ФАКТ: Реальний вивід:
```text
```
EXIT=0

ДАЛІ: ORDERED GATE, cwd=backend: `go vet ./internal/platform/openapigen ./internal/course/adapters/http`
ФАКТ: Реальний вивід:
```text
```
EXIT=0

ДАЛІ: ORDERED GATE, cwd=backend: `go build ./...`
ФАКТ: Реальний вивід:
```text
```
EXIT=0

ДАЛІ: ORDERED GATE, cwd=backend: `go test -count=1 -v ./internal/platform/openapigen ./internal/course/adapters/http`
ФАКТ: Реальний вивід:
```text
=== RUN   TestErrorCodeForms
=== RUN   TestErrorCodeForms/literal
=== RUN   TestErrorCodeForms/cast
=== RUN   TestErrorCodeForms/direct_typed
=== RUN   TestErrorCodeForms/finite_Error_variable
=== RUN   TestErrorCodeForms/sentinel_variable
=== RUN   TestErrorCodeForms/error_forwarding
=== RUN   TestErrorCodeForms/sentinel_method
=== RUN   TestErrorCodeForms/wrapper
=== RUN   TestErrorCodeForms/switch
=== RUN   TestErrorCodeForms/guard
--- PASS: TestErrorCodeForms (0.02s)
    --- PASS: TestErrorCodeForms/literal (0.00s)
    --- PASS: TestErrorCodeForms/cast (0.00s)
    --- PASS: TestErrorCodeForms/direct_typed (0.00s)
    --- PASS: TestErrorCodeForms/finite_Error_variable (0.00s)
    --- PASS: TestErrorCodeForms/sentinel_variable (0.00s)
    --- PASS: TestErrorCodeForms/error_forwarding (0.00s)
    --- PASS: TestErrorCodeForms/sentinel_method (0.00s)
    --- PASS: TestErrorCodeForms/wrapper (0.00s)
    --- PASS: TestErrorCodeForms/switch (0.00s)
    --- PASS: TestErrorCodeForms/guard (0.00s)
=== RUN   TestErrorCodeFailClosed
=== RUN   TestErrorCodeFailClosed/unknown_function
=== RUN   TestErrorCodeFailClosed/typo
=== RUN   TestErrorCodeFailClosed/missing_symbol
=== RUN   TestErrorCodeFailClosed/unbounded_Error
=== RUN   TestErrorCodeFailClosed/unknown_wrapper_caller
=== RUN   TestErrorCodeFailClosed/escaped_writer
=== RUN   TestErrorCodeFailClosed/escaped_wrapper
=== RUN   TestErrorCodeFailClosed/nonexhaustive_switch
=== RUN   TestErrorCodeFailClosed/overwrite
=== RUN   TestErrorCodeFailClosed/conditional_assignment
=== RUN   TestErrorCodeFailClosed/fallthrough
=== RUN   TestErrorCodeFailClosed/guard_mutation
=== RUN   TestErrorCodeFailClosed/address_escape
=== RUN   TestErrorCodeFailClosed/raw_error_argument
=== RUN   TestErrorCodeFailClosed/cycle
--- PASS: TestErrorCodeFailClosed (0.02s)
    --- PASS: TestErrorCodeFailClosed/unknown_function (0.00s)
    --- PASS: TestErrorCodeFailClosed/typo (0.00s)
    --- PASS: TestErrorCodeFailClosed/missing_symbol (0.00s)
    --- PASS: TestErrorCodeFailClosed/unbounded_Error (0.00s)
    --- PASS: TestErrorCodeFailClosed/unknown_wrapper_caller (0.00s)
    --- PASS: TestErrorCodeFailClosed/escaped_writer (0.00s)
    --- PASS: TestErrorCodeFailClosed/escaped_wrapper (0.00s)
    --- PASS: TestErrorCodeFailClosed/nonexhaustive_switch (0.00s)
    --- PASS: TestErrorCodeFailClosed/overwrite (0.00s)
    --- PASS: TestErrorCodeFailClosed/conditional_assignment (0.00s)
    --- PASS: TestErrorCodeFailClosed/fallthrough (0.00s)
    --- PASS: TestErrorCodeFailClosed/guard_mutation (0.00s)
    --- PASS: TestErrorCodeFailClosed/address_escape (0.00s)
    --- PASS: TestErrorCodeFailClosed/raw_error_argument (0.00s)
    --- PASS: TestErrorCodeFailClosed/cycle (0.00s)
=== RUN   TestTransportProvenance
--- PASS: TestTransportProvenance (0.00s)
=== RUN   TestTreeErrorCensus
    errors_test.go:163: WriteError sites=166 domain codes=116 emitted=109
        admin_immutable, admin_not_learner, already_running, authentication_required, cannot_target_self, checkout_already_opened, comment_body_invalid, comment_cursor_invalid, comment_forbidden, comment_not_found, comment_parent_invalid, comment_position_invalid, comment_removed, comment_thread_full, course_already_owned, course_forbidden, course_modified, course_not_deletable, course_not_enrollable, course_not_found, course_purchase_required, dependency_unavailable, document_conflict, email_not_verified, email_taken, enrollment_blocked, enrollment_forbidden, enrollment_required, forbidden, grant_forbidden, grant_self_forbidden, grant_transition_invalid, image_too_large, image_too_many_pixels, insufficient_storage, internal_error, invalid_account_token, invalid_callback, invalid_credentials, invalid_display_name, invalid_email, invalid_image, invalid_lesson_document, invalid_owner, invalid_query, invalid_range, invalid_request, invalid_slug, invalid_status_transition, invalid_totp, last_admin, learner_course_forbidden, lesson_limit, lesson_not_found, membership_conflict, method_not_allowed, not_found, order_not_found, origin_forbidden, password_unchanged, payment_event_rejected, playback_forbidden, playback_session_expired, playback_taken_over, precondition_failed, range_too_wide, rate_limited, recovery_code_invalid, report_already_open, report_forbidden, report_invalid, report_own_comment, report_own_person, report_own_review, request_too_large, review_already_exists, review_invalid, review_moderation_forbidden, review_not_available, review_not_found, review_not_started, review_owner_forbidden, review_purchase_required, review_removed, slug_immutable, slug_taken, token_used, transcript_document_exists, transcript_document_limit, transcript_document_referenced, transcript_exists, transcript_modified, transcript_segment_limit, transcript_text_limit, twofactor_already_enabled, twofactor_not_enabled, unknown_bucket, unknown_metric, unsupported_image, unsupported_media, unsupported_media_type, upload_in_progress, upload_offset_conflict, upload_too_large, user_not_found, video_not_found, video_not_ready, video_not_retryable, weak_password
--- PASS: TestTreeErrorCensus (0.12s)
=== RUN   TestErrorSourceDiscovery
=== RUN   TestErrorSourceDiscovery/unused_domain_code
=== RUN   TestErrorSourceDiscovery/unknown_initializer
=== RUN   TestErrorSourceDiscovery/computed_constant
--- PASS: TestErrorSourceDiscovery (0.00s)
    --- PASS: TestErrorSourceDiscovery/unused_domain_code (0.00s)
    --- PASS: TestErrorSourceDiscovery/unknown_initializer (0.00s)
    --- PASS: TestErrorSourceDiscovery/computed_constant (0.00s)
=== RUN   TestErrorSchemaRequiresSource
--- PASS: TestErrorSchemaRequiresSource (0.00s)
=== RUN   TestOverlayCannotSupplyErrorEnum
--- PASS: TestOverlayCannotSupplyErrorEnum (0.00s)
=== RUN   TestGenerateTwiceIsByteIdentical
--- PASS: TestGenerateTwiceIsByteIdentical (0.00s)
=== RUN   TestFailClosed
=== RUN   TestFailClosed/marshaler
=== RUN   TestFailClosed/pointer_marshaler
=== RUN   TestFailClosed/enum
=== RUN   TestFailClosed/slice_decision
=== RUN   TestFailClosed/orphan_schema
=== RUN   TestFailClosed/orphan_field
=== RUN   TestFailClosed/orphan_path
=== RUN   TestFailClosed/unknown_keyword
=== RUN   TestFailClosed/malformed
=== RUN   TestFailClosed/empty
=== RUN   TestFailClosed/multiple_documents
=== RUN   TestFailClosed/duplicate_YAML_key
--- PASS: TestFailClosed (0.01s)
    --- PASS: TestFailClosed/marshaler (0.00s)
    --- PASS: TestFailClosed/pointer_marshaler (0.00s)
    --- PASS: TestFailClosed/enum (0.00s)
    --- PASS: TestFailClosed/slice_decision (0.00s)
    --- PASS: TestFailClosed/orphan_schema (0.00s)
    --- PASS: TestFailClosed/orphan_field (0.00s)
    --- PASS: TestFailClosed/orphan_path (0.00s)
    --- PASS: TestFailClosed/unknown_keyword (0.00s)
    --- PASS: TestFailClosed/malformed (0.00s)
    --- PASS: TestFailClosed/empty (0.00s)
    --- PASS: TestFailClosed/multiple_documents (0.00s)
    --- PASS: TestFailClosed/duplicate_YAML_key (0.00s)
=== RUN   TestMissingOverlay
--- PASS: TestMissingOverlay (0.00s)
=== RUN   TestNonInjectiveNameMap
--- PASS: TestNonInjectiveNameMap (0.00s)
=== RUN   TestRegisteredEnumAndOverride
--- PASS: TestRegisteredEnumAndOverride (0.00s)
=== RUN   TestNullableSliceDecision
--- PASS: TestNullableSliceDecision (0.00s)
=== RUN   TestStructMarshalerOverrideIsScalarComponent
--- PASS: TestStructMarshalerOverrideIsScalarComponent (0.00s)
=== RUN   TestUntaggedJSONFieldsAndIgnoredFields
--- PASS: TestUntaggedJSONFieldsAndIgnoredFields (0.00s)
=== RUN   TestParametersFromGoWithMetadata
--- PASS: TestParametersFromGoWithMetadata (0.00s)
=== RUN   TestParameterMetadataFailsClosed
=== RUN   TestParameterMetadataFailsClosed/orphan
=== RUN   TestParameterMetadataFailsClosed/array_replacement
=== RUN   TestParameterMetadataFailsClosed/delete_parameters
=== RUN   TestParameterMetadataFailsClosed/delete_parameter
=== RUN   TestParameterMetadataFailsClosed/name
=== RUN   TestParameterMetadataFailsClosed/in
=== RUN   TestParameterMetadataFailsClosed/required
=== RUN   TestParameterMetadataFailsClosed/type
=== RUN   TestParameterMetadataFailsClosed/enum
=== RUN   TestParameterMetadataFailsClosed/ref
=== RUN   TestParameterMetadataFailsClosed/unknown_validation
=== RUN   TestParameterMetadataFailsClosed/no_Go_parameters
--- PASS: TestParameterMetadataFailsClosed (0.01s)
    --- PASS: TestParameterMetadataFailsClosed/orphan (0.00s)
    --- PASS: TestParameterMetadataFailsClosed/array_replacement (0.00s)
    --- PASS: TestParameterMetadataFailsClosed/delete_parameters (0.00s)
    --- PASS: TestParameterMetadataFailsClosed/delete_parameter (0.00s)
    --- PASS: TestParameterMetadataFailsClosed/name (0.00s)
    --- PASS: TestParameterMetadataFailsClosed/in (0.00s)
    --- PASS: TestParameterMetadataFailsClosed/required (0.00s)
    --- PASS: TestParameterMetadataFailsClosed/type (0.00s)
    --- PASS: TestParameterMetadataFailsClosed/enum (0.00s)
    --- PASS: TestParameterMetadataFailsClosed/ref (0.00s)
    --- PASS: TestParameterMetadataFailsClosed/unknown_validation (0.00s)
    --- PASS: TestParameterMetadataFailsClosed/no_Go_parameters (0.00s)
=== RUN   TestInvalidGoParameters
--- PASS: TestInvalidGoParameters (0.00s)
=== RUN   TestParameterMetadataDoesNotInterceptResponseField
--- PASS: TestParameterMetadataDoesNotInterceptResponseField (0.00s)
=== RUN   TestRequestPointerDecisions
=== RUN   TestRequestPointerDecisions/missing_decision_Required
=== RUN   TestRequestPointerDecisions/missing_decision_Optional
=== RUN   TestRequestPointerDecisions/missing_decision_Child
--- PASS: TestRequestPointerDecisions (0.00s)
    --- PASS: TestRequestPointerDecisions/missing_decision_Required (0.00s)
    --- PASS: TestRequestPointerDecisions/missing_decision_Optional (0.00s)
    --- PASS: TestRequestPointerDecisions/missing_decision_Child (0.00s)
=== RUN   TestSharedRequestResponseCannotBypassDecisions
--- PASS: TestSharedRequestResponseCannotBypassDecisions (0.00s)
=== RUN   TestScalarFieldReference
=== RUN   TestScalarFieldReference/matching_scalar
=== RUN   TestScalarFieldReference/wrong_scalar
=== RUN   TestScalarFieldReference/nullable_target
--- PASS: TestScalarFieldReference (0.00s)
    --- PASS: TestScalarFieldReference/matching_scalar (0.00s)
    --- PASS: TestScalarFieldReference/wrong_scalar (0.00s)
    --- PASS: TestScalarFieldReference/nullable_target (0.00s)
=== RUN   TestParameterFormatMetadata
--- PASS: TestParameterFormatMetadata (0.00s)
PASS
ok  	courses/internal/platform/openapigen	1.603s
=== RUN   TestDocumentInvalidDraftNeverPersists
=== RUN   TestDocumentInvalidDraftNeverPersists/{"id":"bad","type":"iframe"}
=== RUN   TestDocumentInvalidDraftNeverPersists/{"id":"v","type":"video","content":{"blocks":[]}}
=== RUN   TestDocumentInvalidDraftNeverPersists/{"id":"row","type":"row","columns":[{"id":"nested","type":"row","columns":[]}]}
--- PASS: TestDocumentInvalidDraftNeverPersists (0.00s)
    --- PASS: TestDocumentInvalidDraftNeverPersists/{"id":"bad","type":"iframe"} (0.00s)
    --- PASS: TestDocumentInvalidDraftNeverPersists/{"id":"v","type":"video","content":{"blocks":[]}} (0.00s)
    --- PASS: TestDocumentInvalidDraftNeverPersists/{"id":"row","type":"row","columns":[{"id":"nested","type":"row","columns":[]}]} (0.00s)
=== RUN   TestDocumentNonEditor
=== RUN   TestDocumentNonEditor/GET
=== RUN   TestDocumentNonEditor/PUT
=== RUN   TestDocumentNonEditor/POST
--- PASS: TestDocumentNonEditor (0.00s)
    --- PASS: TestDocumentNonEditor/GET (0.00s)
    --- PASS: TestDocumentNonEditor/PUT (0.00s)
    --- PASS: TestDocumentNonEditor/POST (0.00s)
=== RUN   TestDocumentPublishedAccess
=== RUN   TestDocumentPublishedAccess/paid_nonmember
=== RUN   TestDocumentPublishedAccess/active_member
=== RUN   TestDocumentPublishedAccess/email_first
=== RUN   TestDocumentPublishedAccess/absent_publication
=== RUN   TestDocumentPublishedAccess/verified_preview
--- PASS: TestDocumentPublishedAccess (0.00s)
    --- PASS: TestDocumentPublishedAccess/paid_nonmember (0.00s)
    --- PASS: TestDocumentPublishedAccess/active_member (0.00s)
    --- PASS: TestDocumentPublishedAccess/email_first (0.00s)
    --- PASS: TestDocumentPublishedAccess/absent_publication (0.00s)
    --- PASS: TestDocumentPublishedAccess/verified_preview (0.00s)
=== RUN   TestDocumentConflicts
=== RUN   TestDocumentConflicts/PUT
=== RUN   TestDocumentConflicts/POST
--- PASS: TestDocumentConflicts (0.00s)
    --- PASS: TestDocumentConflicts/PUT (0.00s)
    --- PASS: TestDocumentConflicts/POST (0.00s)
=== RUN   TestDocumentDraftRoundTrip
--- PASS: TestDocumentDraftRoundTrip (0.00s)
=== RUN   TestDocumentTransportRejections
=== RUN   TestDocumentTransportRejections/unknown_query
=== RUN   TestDocumentTransportRejections/unknown_locale
=== RUN   TestDocumentTransportRejections/empty_locale
=== RUN   TestDocumentTransportRejections/duplicate_locale
=== RUN   TestDocumentTransportRejections/malformed_query
=== RUN   TestDocumentTransportRejections/PUT_query
=== RUN   TestDocumentTransportRejections/publish_query
=== RUN   TestDocumentTransportRejections/unknown_wrapper
=== RUN   TestDocumentTransportRejections/unknown_nested_field
=== RUN   TestDocumentTransportRejections/duplicate_nested_field
=== RUN   TestDocumentTransportRejections/trailing_JSON
=== RUN   TestDocumentTransportRejections/null_wrapper
=== RUN   TestDocumentTransportRejections/bad_timestamp
=== RUN   TestDocumentTransportRejections/invalid_document_locale
=== RUN   TestDocumentTransportRejections/publish_missing_token
=== RUN   TestDocumentTransportRejections/publish_null_token
=== RUN   TestDocumentTransportRejections/publish_unknown_field
=== RUN   TestDocumentTransportRejections/publish_trailing_bytes
=== RUN   TestDocumentTransportRejections/publish_too_large
=== RUN   TestDocumentTransportRejections/PUT_too_large
--- PASS: TestDocumentTransportRejections (0.02s)
    --- PASS: TestDocumentTransportRejections/unknown_query (0.00s)
    --- PASS: TestDocumentTransportRejections/unknown_locale (0.00s)
    --- PASS: TestDocumentTransportRejections/empty_locale (0.00s)
    --- PASS: TestDocumentTransportRejections/duplicate_locale (0.00s)
    --- PASS: TestDocumentTransportRejections/malformed_query (0.00s)
    --- PASS: TestDocumentTransportRejections/PUT_query (0.00s)
    --- PASS: TestDocumentTransportRejections/publish_query (0.00s)
    --- PASS: TestDocumentTransportRejections/unknown_wrapper (0.00s)
    --- PASS: TestDocumentTransportRejections/unknown_nested_field (0.00s)
    --- PASS: TestDocumentTransportRejections/duplicate_nested_field (0.00s)
    --- PASS: TestDocumentTransportRejections/trailing_JSON (0.00s)
    --- PASS: TestDocumentTransportRejections/null_wrapper (0.00s)
    --- PASS: TestDocumentTransportRejections/bad_timestamp (0.00s)
    --- PASS: TestDocumentTransportRejections/invalid_document_locale (0.00s)
    --- PASS: TestDocumentTransportRejections/publish_missing_token (0.00s)
    --- PASS: TestDocumentTransportRejections/publish_null_token (0.00s)
    --- PASS: TestDocumentTransportRejections/publish_unknown_field (0.00s)
    --- PASS: TestDocumentTransportRejections/publish_trailing_bytes (0.00s)
    --- PASS: TestDocumentTransportRejections/publish_too_large (0.00s)
    --- PASS: TestDocumentTransportRejections/PUT_too_large (0.01s)
=== RUN   TestDocumentWriteLimitExceedsAdminLimit
--- PASS: TestDocumentWriteLimitExceedsAdminLimit (0.00s)
=== RUN   TestDocumentErrorMapping
=== RUN   TestDocumentErrorMapping/invalid_lesson_document
=== RUN   TestDocumentErrorMapping/invalid_request
=== RUN   TestDocumentErrorMapping/invalid_query
=== RUN   TestDocumentErrorMapping/email_not_verified
=== RUN   TestDocumentErrorMapping/course_purchase_required
=== RUN   TestDocumentErrorMapping/forbidden
=== RUN   TestDocumentErrorMapping/course_forbidden
=== RUN   TestDocumentErrorMapping/lesson_not_found
=== RUN   TestDocumentErrorMapping/document_conflict
=== RUN   TestDocumentErrorMapping/dependency_unavailable
=== RUN   TestDocumentErrorMapping/internal_error
--- PASS: TestDocumentErrorMapping (0.00s)
    --- PASS: TestDocumentErrorMapping/invalid_lesson_document (0.00s)
    --- PASS: TestDocumentErrorMapping/invalid_request (0.00s)
    --- PASS: TestDocumentErrorMapping/invalid_query (0.00s)
    --- PASS: TestDocumentErrorMapping/email_not_verified (0.00s)
    --- PASS: TestDocumentErrorMapping/course_purchase_required (0.00s)
    --- PASS: TestDocumentErrorMapping/forbidden (0.00s)
    --- PASS: TestDocumentErrorMapping/course_forbidden (0.00s)
    --- PASS: TestDocumentErrorMapping/lesson_not_found (0.00s)
    --- PASS: TestDocumentErrorMapping/document_conflict (0.00s)
    --- PASS: TestDocumentErrorMapping/dependency_unavailable (0.00s)
    --- PASS: TestDocumentErrorMapping/internal_error (0.00s)
=== RUN   TestLearningOutlineHasNoDocumentBody
--- PASS: TestLearningOutlineHasNoDocumentBody (0.00s)
=== RUN   TestDocumentHTTPMatchesOpenAPI
=== RUN   TestDocumentHTTPMatchesOpenAPI/GET/api/v1/admin/lessons/20000000-0000-4000-8000-000000000001/document?locale=en
=== RUN   TestDocumentHTTPMatchesOpenAPI/PUT/api/v1/admin/lessons/20000000-0000-4000-8000-000000000001/document
=== RUN   TestDocumentHTTPMatchesOpenAPI/POST/api/v1/admin/lessons/20000000-0000-4000-8000-000000000001/document/publish
=== RUN   TestDocumentHTTPMatchesOpenAPI/GET/api/v1/learn/lessons/20000000-0000-4000-8000-000000000001/document?locale=en
--- PASS: TestDocumentHTTPMatchesOpenAPI (0.07s)
    --- PASS: TestDocumentHTTPMatchesOpenAPI/GET/api/v1/admin/lessons/20000000-0000-4000-8000-000000000001/document?locale=en (0.00s)
    --- PASS: TestDocumentHTTPMatchesOpenAPI/PUT/api/v1/admin/lessons/20000000-0000-4000-8000-000000000001/document (0.00s)
    --- PASS: TestDocumentHTTPMatchesOpenAPI/POST/api/v1/admin/lessons/20000000-0000-4000-8000-000000000001/document/publish (0.00s)
    --- PASS: TestDocumentHTTPMatchesOpenAPI/GET/api/v1/learn/lessons/20000000-0000-4000-8000-000000000001/document?locale=en (0.00s)
=== RUN   TestDocumentAuthenticationAndOriginWiring
=== RUN   TestDocumentAuthenticationAndOriginWiring/GET/api/v1/admin/lessons/20000000-0000-4000-8000-000000000001/document
=== RUN   TestDocumentAuthenticationAndOriginWiring/GET/api/v1/learn/lessons/20000000-0000-4000-8000-000000000001/document
=== RUN   TestDocumentAuthenticationAndOriginWiring/PUT/api/v1/admin/lessons/20000000-0000-4000-8000-000000000001/document
=== RUN   TestDocumentAuthenticationAndOriginWiring/POST/api/v1/admin/lessons/20000000-0000-4000-8000-000000000001/document/publish
--- PASS: TestDocumentAuthenticationAndOriginWiring (0.00s)
    --- PASS: TestDocumentAuthenticationAndOriginWiring/GET/api/v1/admin/lessons/20000000-0000-4000-8000-000000000001/document (0.00s)
    --- PASS: TestDocumentAuthenticationAndOriginWiring/GET/api/v1/learn/lessons/20000000-0000-4000-8000-000000000001/document (0.00s)
    --- PASS: TestDocumentAuthenticationAndOriginWiring/PUT/api/v1/admin/lessons/20000000-0000-4000-8000-000000000001/document (0.00s)
    --- PASS: TestDocumentAuthenticationAndOriginWiring/POST/api/v1/admin/lessons/20000000-0000-4000-8000-000000000001/document/publish (0.00s)
=== RUN   TestDocumentRecursiveSchemaRoundTrip
=== RUN   TestDocumentRecursiveSchemaRoundTrip/{"schemaVersion":1,"locale":"en","sections":[{"id":"row","type":"row","columns":[{"id":"text","type":"text","content":{"blocks":[{"type":"heading","level":2,"spans":[{"text":"Title","marks":["bold","italic"],"href":"https://example.com"}]},{"type":"list","ordered":true,"items":[[{"text":"Item"}]]}]}},{"id":"image","type":"image","imageId":"image-1","contentType":"image/png","alt":"Image"}]},{"id":"task","type":"task","content":{"blocks":[]}},{"id":"video","type":"video","content":null}]}
=== RUN   TestDocumentRecursiveSchemaRoundTrip/{"schemaVersion":1,"locale":"uk","sections":null}
=== RUN   TestDocumentRecursiveSchemaRoundTrip/{"schemaVersion":1,"locale":"uk","sections":[{"id":"text","type":"text","content":{"blocks":null}}]}
--- PASS: TestDocumentRecursiveSchemaRoundTrip (0.03s)
    --- PASS: TestDocumentRecursiveSchemaRoundTrip/{"schemaVersion":1,"locale":"en","sections":[{"id":"row","type":"row","columns":[{"id":"text","type":"text","content":{"blocks":[{"type":"heading","level":2,"spans":[{"text":"Title","marks":["bold","italic"],"href":"https://example.com"}]},{"type":"list","ordered":true,"items":[[{"text":"Item"}]]}]}},{"id":"image","type":"image","imageId":"image-1","contentType":"image/png","alt":"Image"}]},{"id":"task","type":"task","content":{"blocks":[]}},{"id":"video","type":"video","content":null}]} (0.00s)
    --- PASS: TestDocumentRecursiveSchemaRoundTrip/{"schemaVersion":1,"locale":"uk","sections":null} (0.00s)
    --- PASS: TestDocumentRecursiveSchemaRoundTrip/{"schemaVersion":1,"locale":"uk","sections":[{"id":"text","type":"text","content":{"blocks":null}}]} (0.00s)
=== RUN   TestErrorBoundaries
=== RUN   TestErrorBoundaries/writeDocumentError/unknown
=== RUN   TestErrorBoundaries/writeLearningError/unknown
=== RUN   TestErrorBoundaries/writeError/unknown
=== RUN   TestErrorBoundaries/writeAdminError/unknown
=== RUN   TestErrorBoundaries/writeError/invalid_query
=== RUN   TestErrorBoundaries/writeError/invalid_slug
=== RUN   TestErrorBoundaries/writeError/course_not_found
=== RUN   TestErrorBoundaries/writeError/dependency_unavailable
=== RUN   TestErrorBoundaries/writeDocumentError/invalid_lesson_document
=== RUN   TestErrorBoundaries/writeDocumentError/invalid_request
=== RUN   TestErrorBoundaries/writeDocumentError/invalid_query
=== RUN   TestErrorBoundaries/writeDocumentError/email_not_verified
=== RUN   TestErrorBoundaries/writeDocumentError/course_purchase_required
=== RUN   TestErrorBoundaries/writeDocumentError/forbidden
=== RUN   TestErrorBoundaries/writeDocumentError/course_forbidden
=== RUN   TestErrorBoundaries/writeDocumentError/lesson_not_found
=== RUN   TestErrorBoundaries/writeDocumentError/document_conflict
=== RUN   TestErrorBoundaries/writeDocumentError/dependency_unavailable
=== RUN   TestErrorBoundaries/writeAdminError/invalid_request
=== RUN   TestErrorBoundaries/writeAdminError/forbidden
=== RUN   TestErrorBoundaries/writeAdminError/course_forbidden
=== RUN   TestErrorBoundaries/writeAdminError/course_not_found
=== RUN   TestErrorBoundaries/writeAdminError/invalid_owner
=== RUN   TestErrorBoundaries/writeAdminError/slug_taken
=== RUN   TestErrorBoundaries/writeAdminError/slug_immutable
=== RUN   TestErrorBoundaries/writeAdminError/course_modified
=== RUN   TestErrorBoundaries/writeAdminError/invalid_status_transition
=== RUN   TestErrorBoundaries/writeAdminError/course_not_deletable
=== RUN   TestErrorBoundaries/writeAdminError/lesson_limit
=== RUN   TestErrorBoundaries/writeAdminError/dependency_unavailable
=== RUN   TestErrorBoundaries/writeLearningError/course_purchase_required
=== RUN   TestErrorBoundaries/writeLearningError/course_forbidden
=== RUN   TestErrorBoundaries/writeLearningError/course_forbidden#01
--- PASS: TestErrorBoundaries (0.00s)
    --- PASS: TestErrorBoundaries/writeDocumentError/unknown (0.00s)
    --- PASS: TestErrorBoundaries/writeLearningError/unknown (0.00s)
    --- PASS: TestErrorBoundaries/writeError/unknown (0.00s)
    --- PASS: TestErrorBoundaries/writeAdminError/unknown (0.00s)
    --- PASS: TestErrorBoundaries/writeError/invalid_query (0.00s)
    --- PASS: TestErrorBoundaries/writeError/invalid_slug (0.00s)
    --- PASS: TestErrorBoundaries/writeError/course_not_found (0.00s)
    --- PASS: TestErrorBoundaries/writeError/dependency_unavailable (0.00s)
    --- PASS: TestErrorBoundaries/writeDocumentError/invalid_lesson_document (0.00s)
    --- PASS: TestErrorBoundaries/writeDocumentError/invalid_request (0.00s)
    --- PASS: TestErrorBoundaries/writeDocumentError/invalid_query (0.00s)
    --- PASS: TestErrorBoundaries/writeDocumentError/email_not_verified (0.00s)
    --- PASS: TestErrorBoundaries/writeDocumentError/course_purchase_required (0.00s)
    --- PASS: TestErrorBoundaries/writeDocumentError/forbidden (0.00s)
    --- PASS: TestErrorBoundaries/writeDocumentError/course_forbidden (0.00s)
    --- PASS: TestErrorBoundaries/writeDocumentError/lesson_not_found (0.00s)
    --- PASS: TestErrorBoundaries/writeDocumentError/document_conflict (0.00s)
    --- PASS: TestErrorBoundaries/writeDocumentError/dependency_unavailable (0.00s)
    --- PASS: TestErrorBoundaries/writeAdminError/invalid_request (0.00s)
    --- PASS: TestErrorBoundaries/writeAdminError/forbidden (0.00s)
    --- PASS: TestErrorBoundaries/writeAdminError/course_forbidden (0.00s)
    --- PASS: TestErrorBoundaries/writeAdminError/course_not_found (0.00s)
    --- PASS: TestErrorBoundaries/writeAdminError/invalid_owner (0.00s)
    --- PASS: TestErrorBoundaries/writeAdminError/slug_taken (0.00s)
    --- PASS: TestErrorBoundaries/writeAdminError/slug_immutable (0.00s)
    --- PASS: TestErrorBoundaries/writeAdminError/course_modified (0.00s)
    --- PASS: TestErrorBoundaries/writeAdminError/invalid_status_transition (0.00s)
    --- PASS: TestErrorBoundaries/writeAdminError/course_not_deletable (0.00s)
    --- PASS: TestErrorBoundaries/writeAdminError/lesson_limit (0.00s)
    --- PASS: TestErrorBoundaries/writeAdminError/dependency_unavailable (0.00s)
    --- PASS: TestErrorBoundaries/writeLearningError/course_purchase_required (0.00s)
    --- PASS: TestErrorBoundaries/writeLearningError/course_forbidden (0.00s)
    --- PASS: TestErrorBoundaries/writeLearningError/course_forbidden#01 (0.00s)
=== RUN   TestCatalogOpenAPIFidelity
    openapi_contract_test.go:55: error enum ref-closure mutation detected; 109 emitted codes compared
    openapi_contract_test.go:84: catalog and status request + complete referenced component closures equal hand spec; two generations byte-identical
--- PASS: TestCatalogOpenAPIFidelity (0.23s)
=== RUN   TestAdminReadHTTPMatchesOpenAPI
=== RUN   TestAdminReadHTTPMatchesOpenAPI/admin_lists_every_course
=== RUN   TestAdminReadHTTPMatchesOpenAPI/admin_list_search_parameters
=== RUN   TestAdminReadHTTPMatchesOpenAPI/instructor_lists_own_courses
=== RUN   TestAdminReadHTTPMatchesOpenAPI/admin_reads_a_tree
=== RUN   TestAdminReadHTTPMatchesOpenAPI/owner_reads_a_tree
=== RUN   TestAdminReadHTTPMatchesOpenAPI/P07_admin_lists_owner_candidates
=== RUN   TestAdminReadHTTPMatchesOpenAPI/R01_anonymous_list
=== RUN   TestAdminReadHTTPMatchesOpenAPI/R01_anonymous_tree
=== RUN   TestAdminReadHTTPMatchesOpenAPI/R01_anonymous_instructors
=== RUN   TestAdminReadHTTPMatchesOpenAPI/R02_student_list
=== RUN   TestAdminReadHTTPMatchesOpenAPI/student_role_is_refused_before_query_validation
=== RUN   TestAdminReadHTTPMatchesOpenAPI/R02_student_tree
=== RUN   TestAdminReadHTTPMatchesOpenAPI/R02_student_instructors
=== RUN   TestAdminReadHTTPMatchesOpenAPI/R03_disabled_account
=== RUN   TestAdminReadHTTPMatchesOpenAPI/R04_foreign_instructor
=== RUN   TestAdminReadHTTPMatchesOpenAPI/R05_instructor_owner_candidates
=== RUN   TestAdminReadHTTPMatchesOpenAPI/R06_unknown_id
=== RUN   TestAdminReadHTTPMatchesOpenAPI/R07_malformed_id
=== RUN   TestAdminReadHTTPMatchesOpenAPI/invalid_list_status
=== RUN   TestAdminReadHTTPMatchesOpenAPI/unknown_list_parameter
=== RUN   TestAdminReadHTTPMatchesOpenAPI/repeated_list_parameter
--- PASS: TestAdminReadHTTPMatchesOpenAPI (0.06s)
    --- PASS: TestAdminReadHTTPMatchesOpenAPI/admin_lists_every_course (0.00s)
    --- PASS: TestAdminReadHTTPMatchesOpenAPI/admin_list_search_parameters (0.00s)
    --- PASS: TestAdminReadHTTPMatchesOpenAPI/instructor_lists_own_courses (0.00s)
    --- PASS: TestAdminReadHTTPMatchesOpenAPI/admin_reads_a_tree (0.00s)
    --- PASS: TestAdminReadHTTPMatchesOpenAPI/owner_reads_a_tree (0.00s)
    --- PASS: TestAdminReadHTTPMatchesOpenAPI/P07_admin_lists_owner_candidates (0.00s)
    --- PASS: TestAdminReadHTTPMatchesOpenAPI/R01_anonymous_list (0.00s)
    --- PASS: TestAdminReadHTTPMatchesOpenAPI/R01_anonymous_tree (0.00s)
    --- PASS: TestAdminReadHTTPMatchesOpenAPI/R01_anonymous_instructors (0.00s)
    --- PASS: TestAdminReadHTTPMatchesOpenAPI/R02_student_list (0.00s)
    --- PASS: TestAdminReadHTTPMatchesOpenAPI/student_role_is_refused_before_query_validation (0.00s)
    --- PASS: TestAdminReadHTTPMatchesOpenAPI/R02_student_tree (0.00s)
    --- PASS: TestAdminReadHTTPMatchesOpenAPI/R02_student_instructors (0.00s)
    --- PASS: TestAdminReadHTTPMatchesOpenAPI/R03_disabled_account (0.00s)
    --- PASS: TestAdminReadHTTPMatchesOpenAPI/R04_foreign_instructor (0.00s)
    --- PASS: TestAdminReadHTTPMatchesOpenAPI/R05_instructor_owner_candidates (0.00s)
    --- PASS: TestAdminReadHTTPMatchesOpenAPI/R06_unknown_id (0.00s)
    --- PASS: TestAdminReadHTTPMatchesOpenAPI/R07_malformed_id (0.00s)
    --- PASS: TestAdminReadHTTPMatchesOpenAPI/invalid_list_status (0.00s)
    --- PASS: TestAdminReadHTTPMatchesOpenAPI/unknown_list_parameter (0.00s)
    --- PASS: TestAdminReadHTTPMatchesOpenAPI/repeated_list_parameter (0.00s)
=== RUN   TestAdminCreateContract
=== RUN   TestAdminCreateContract/P02_P03_R29_admin_creates_one_invisible_draft_and_replay_is_refused
=== RUN   TestAdminCreateContract/P04_instructor_owns_a_create_that_omits_instructorId
=== RUN   TestAdminCreateContract/P05_admin_chooses_an_owner_and_the_owner_lists_the_course
--- PASS: TestAdminCreateContract (0.24s)
    --- PASS: TestAdminCreateContract/P02_P03_R29_admin_creates_one_invisible_draft_and_replay_is_refused (0.12s)
    --- PASS: TestAdminCreateContract/P04_instructor_owns_a_create_that_omits_instructorId (0.06s)
    --- PASS: TestAdminCreateContract/P05_admin_chooses_an_owner_and_the_owner_lists_the_course (0.06s)
=== RUN   TestAdminCourseListEnrichmentContracts
=== RUN   TestAdminCourseListEnrichmentContracts/Cv01_list_row_carries_lessonCount,_studentCount,_price_and_the_status_snapshot
=== RUN   TestAdminCourseListEnrichmentContracts/Cv02_a_free_course_serialises_price_as_null
=== RUN   TestAdminCourseListEnrichmentContracts/Cv03_search_is_forwarded_to_the_query
=== RUN   TestAdminCourseListEnrichmentContracts/Cv04_an_empty_search_parameter_is_rejected
=== RUN   TestAdminCourseListEnrichmentContracts/Cv06_sort_is_forwarded_and_defaults_to_recent
=== RUN   TestAdminCourseListEnrichmentContracts/Cv07_an_empty_or_unknown_sort_is_rejected_before_the_repository
--- PASS: TestAdminCourseListEnrichmentContracts (0.25s)
    --- PASS: TestAdminCourseListEnrichmentContracts/Cv01_list_row_carries_lessonCount,_studentCount,_price_and_the_status_snapshot (0.06s)
    --- PASS: TestAdminCourseListEnrichmentContracts/Cv02_a_free_course_serialises_price_as_null (0.06s)
    --- PASS: TestAdminCourseListEnrichmentContracts/Cv03_search_is_forwarded_to_the_query (0.06s)
    --- PASS: TestAdminCourseListEnrichmentContracts/Cv04_an_empty_search_parameter_is_rejected (0.00s)
    --- PASS: TestAdminCourseListEnrichmentContracts/Cv06_sort_is_forwarded_and_defaults_to_recent (0.06s)
    --- PASS: TestAdminCourseListEnrichmentContracts/Cv07_an_empty_or_unknown_sort_is_rejected_before_the_repository (0.00s)
=== RUN   TestAdminSaveHappyAndBoundaryContracts
=== RUN   TestAdminSaveHappyAndBoundaryContracts/P06_admin_reassigns_the_owner_and_both_lists_move
=== RUN   TestAdminSaveHappyAndBoundaryContracts/B04_a_never-published_draft_can_be_renamed
=== RUN   TestAdminSaveHappyAndBoundaryContracts/B05_maximal_Cyrillic_tree_stays_below_the_cap_and_saves
=== RUN   TestAdminSaveHappyAndBoundaryContracts/an_empty_tree_saves_as_the_complete_replacement
=== RUN   TestAdminSaveHappyAndBoundaryContracts/B06_invalid_tree_ceilings_leave_the_tree_unchanged
=== RUN   TestAdminSaveHappyAndBoundaryContracts/B06_invalid_tree_ceilings_leave_the_tree_unchanged/41_modules
=== RUN   TestAdminSaveHappyAndBoundaryContracts/B06_invalid_tree_ceilings_leave_the_tree_unchanged/121_lessons
=== RUN   TestAdminSaveHappyAndBoundaryContracts/B07_a_body_over_512_KiB_is_refused_with_a_bounded_response
=== RUN   TestAdminSaveHappyAndBoundaryContracts/B08_summary_accepts_500_characters_and_rejects_501
=== RUN   TestAdminSaveHappyAndBoundaryContracts/course_and_node_field_bounds_surface_as_invalid_request
=== RUN   TestAdminSaveHappyAndBoundaryContracts/course_and_node_field_bounds_surface_as_invalid_request/empty_title
=== RUN   TestAdminSaveHappyAndBoundaryContracts/course_and_node_field_bounds_surface_as_invalid_request/201_character_title
=== RUN   TestAdminSaveHappyAndBoundaryContracts/course_and_node_field_bounds_surface_as_invalid_request/20001_character_description
=== RUN   TestAdminSaveHappyAndBoundaryContracts/course_and_node_field_bounds_surface_as_invalid_request/121_character_slug
=== RUN   TestAdminSaveHappyAndBoundaryContracts/course_and_node_field_bounds_surface_as_invalid_request/201_character_module_title
=== RUN   TestAdminSaveHappyAndBoundaryContracts/course_and_node_field_bounds_surface_as_invalid_request/201_character_lesson_title
=== RUN   TestAdminSaveHappyAndBoundaryContracts/B-3_the_course-wide_ceiling_is_refused_by_name_where_the_schema_cannot_see_it
=== RUN   TestAdminSaveHappyAndBoundaryContracts/B-3_the_course-wide_ceiling_is_refused_by_name_where_the_schema_cannot_see_it/121_across_two_modules_passes_the_schema_and_is_refused_as_lesson_limit
=== RUN   TestAdminSaveHappyAndBoundaryContracts/B-3_the_course-wide_ceiling_is_refused_by_name_where_the_schema_cannot_see_it/121_in_one_module_is_stopped_by_the_per-module_maxItems
=== RUN   TestAdminSaveHappyAndBoundaryContracts/B12_read_token_round-trips_and_a_lesson-only_save_retires_it
--- PASS: TestAdminSaveHappyAndBoundaryContracts (1.17s)
    --- PASS: TestAdminSaveHappyAndBoundaryContracts/P06_admin_reassigns_the_owner_and_both_lists_move (0.06s)
    --- PASS: TestAdminSaveHappyAndBoundaryContracts/B04_a_never-published_draft_can_be_renamed (0.06s)
    --- PASS: TestAdminSaveHappyAndBoundaryContracts/B05_maximal_Cyrillic_tree_stays_below_the_cap_and_saves (0.07s)
    --- PASS: TestAdminSaveHappyAndBoundaryContracts/an_empty_tree_saves_as_the_complete_replacement (0.06s)
    --- PASS: TestAdminSaveHappyAndBoundaryContracts/B06_invalid_tree_ceilings_leave_the_tree_unchanged (0.12s)
        --- PASS: TestAdminSaveHappyAndBoundaryContracts/B06_invalid_tree_ceilings_leave_the_tree_unchanged/41_modules (0.06s)
        --- PASS: TestAdminSaveHappyAndBoundaryContracts/B06_invalid_tree_ceilings_leave_the_tree_unchanged/121_lessons (0.06s)
    --- PASS: TestAdminSaveHappyAndBoundaryContracts/B07_a_body_over_512_KiB_is_refused_with_a_bounded_response (0.06s)
    --- PASS: TestAdminSaveHappyAndBoundaryContracts/B08_summary_accepts_500_characters_and_rejects_501 (0.13s)
    --- PASS: TestAdminSaveHappyAndBoundaryContracts/course_and_node_field_bounds_surface_as_invalid_request (0.37s)
        --- PASS: TestAdminSaveHappyAndBoundaryContracts/course_and_node_field_bounds_surface_as_invalid_request/empty_title (0.06s)
        --- PASS: TestAdminSaveHappyAndBoundaryContracts/course_and_node_field_bounds_surface_as_invalid_request/201_character_title (0.06s)
        --- PASS: TestAdminSaveHappyAndBoundaryContracts/course_and_node_field_bounds_surface_as_invalid_request/20001_character_description (0.06s)
        --- PASS: TestAdminSaveHappyAndBoundaryContracts/course_and_node_field_bounds_surface_as_invalid_request/121_character_slug (0.06s)
        --- PASS: TestAdminSaveHappyAndBoundaryContracts/course_and_node_field_bounds_surface_as_invalid_request/201_character_module_title (0.06s)
        --- PASS: TestAdminSaveHappyAndBoundaryContracts/course_and_node_field_bounds_surface_as_invalid_request/201_character_lesson_title (0.06s)
    --- PASS: TestAdminSaveHappyAndBoundaryContracts/B-3_the_course-wide_ceiling_is_refused_by_name_where_the_schema_cannot_see_it (0.18s)
        --- PASS: TestAdminSaveHappyAndBoundaryContracts/B-3_the_course-wide_ceiling_is_refused_by_name_where_the_schema_cannot_see_it/121_across_two_modules_passes_the_schema_and_is_refused_as_lesson_limit (0.06s)
        --- PASS: TestAdminSaveHappyAndBoundaryContracts/B-3_the_course-wide_ceiling_is_refused_by_name_where_the_schema_cannot_see_it/121_in_one_module_is_stopped_by_the_per-module_maxItems (0.11s)
    --- PASS: TestAdminSaveHappyAndBoundaryContracts/B12_read_token_round-trips_and_a_lesson-only_save_retires_it (0.06s)
=== RUN   TestAdminSaveRefusalContracts
=== RUN   TestAdminSaveRefusalContracts/R12_instructor_cannot_reassign_ownership
=== RUN   TestAdminSaveRefusalContracts/R13_ineligible_owners_are_refused_without_changing_owner
=== RUN   TestAdminSaveRefusalContracts/R13_ineligible_owners_are_refused_without_changing_owner/student
=== RUN   TestAdminSaveRefusalContracts/R13_ineligible_owners_are_refused_without_changing_owner/disabled_instructor
=== RUN   TestAdminSaveRefusalContracts/R13_ineligible_owners_are_refused_without_changing_owner/unknown_user
=== RUN   TestAdminSaveRefusalContracts/R20_published_and_archived_slugs_are_immutable
=== RUN   TestAdminSaveRefusalContracts/R20_published_and_archived_slugs_are_immutable/published
=== RUN   TestAdminSaveRefusalContracts/R20_published_and_archived_slugs_are_immutable/archived
=== RUN   TestAdminSaveRefusalContracts/R21_rename_onto_a_taken_slug_is_refused
=== RUN   TestAdminSaveRefusalContracts/R22_lifecycle_and_position_members_are_rejected_before_repository_access
=== RUN   TestAdminSaveRefusalContracts/R22_lifecycle_and_position_members_are_rejected_before_repository_access/status
=== RUN   TestAdminSaveRefusalContracts/R22_lifecycle_and_position_members_are_rejected_before_repository_access/publishedAt
=== RUN   TestAdminSaveRefusalContracts/R22_lifecycle_and_position_members_are_rejected_before_repository_access/position
--- PASS: TestAdminSaveRefusalContracts (0.63s)
    --- PASS: TestAdminSaveRefusalContracts/R12_instructor_cannot_reassign_ownership (0.06s)
    --- PASS: TestAdminSaveRefusalContracts/R13_ineligible_owners_are_refused_without_changing_owner (0.21s)
        --- PASS: TestAdminSaveRefusalContracts/R13_ineligible_owners_are_refused_without_changing_owner/student (0.08s)
        --- PASS: TestAdminSaveRefusalContracts/R13_ineligible_owners_are_refused_without_changing_owner/disabled_instructor (0.07s)
        --- PASS: TestAdminSaveRefusalContracts/R13_ineligible_owners_are_refused_without_changing_owner/unknown_user (0.06s)
    --- PASS: TestAdminSaveRefusalContracts/R20_published_and_archived_slugs_are_immutable (0.12s)
        --- PASS: TestAdminSaveRefusalContracts/R20_published_and_archived_slugs_are_immutable/published (0.06s)
        --- PASS: TestAdminSaveRefusalContracts/R20_published_and_archived_slugs_are_immutable/archived (0.06s)
    --- PASS: TestAdminSaveRefusalContracts/R21_rename_onto_a_taken_slug_is_refused (0.06s)
    --- PASS: TestAdminSaveRefusalContracts/R22_lifecycle_and_position_members_are_rejected_before_repository_access (0.19s)
        --- PASS: TestAdminSaveRefusalContracts/R22_lifecycle_and_position_members_are_rejected_before_repository_access/status (0.06s)
        --- PASS: TestAdminSaveRefusalContracts/R22_lifecycle_and_position_members_are_rejected_before_repository_access/publishedAt (0.06s)
        --- PASS: TestAdminSaveRefusalContracts/R22_lifecycle_and_position_members_are_rejected_before_repository_access/position (0.06s)
=== RUN   TestAdminLifecycleHappyAndBoundaryContracts
=== RUN   TestAdminLifecycleHappyAndBoundaryContracts/P11_P12_P13_publish_archive_and_restore_update_public_visibility
=== RUN   TestAdminLifecycleHappyAndBoundaryContracts/P14_instructor_publishes_their_own_course
=== RUN   TestAdminLifecycleHappyAndBoundaryContracts/B01_an_empty_course_may_be_published
=== RUN   TestAdminLifecycleHappyAndBoundaryContracts/a_draft_course_can_be_deleted
--- PASS: TestAdminLifecycleHappyAndBoundaryContracts (0.38s)
    --- PASS: TestAdminLifecycleHappyAndBoundaryContracts/P11_P12_P13_publish_archive_and_restore_update_public_visibility (0.20s)
    --- PASS: TestAdminLifecycleHappyAndBoundaryContracts/P14_instructor_publishes_their_own_course (0.06s)
    --- PASS: TestAdminLifecycleHappyAndBoundaryContracts/B01_an_empty_course_may_be_published (0.06s)
    --- PASS: TestAdminLifecycleHappyAndBoundaryContracts/a_draft_course_can_be_deleted (0.06s)
=== RUN   TestAdminMutationRefusalMatrix
=== RUN   TestAdminMutationRefusalMatrix/R08_anonymous_is_refused_on_every_mutation
=== RUN   TestAdminMutationRefusalMatrix/R08_anonymous_is_refused_on_every_mutation/create
=== RUN   TestAdminMutationRefusalMatrix/R08_anonymous_is_refused_on_every_mutation/save
=== RUN   TestAdminMutationRefusalMatrix/R08_anonymous_is_refused_on_every_mutation/status
=== RUN   TestAdminMutationRefusalMatrix/R08_anonymous_is_refused_on_every_mutation/delete
=== RUN   TestAdminMutationRefusalMatrix/R09_a_student_is_refused_on_every_mutation
=== RUN   TestAdminMutationRefusalMatrix/R09_a_student_is_refused_on_every_mutation/create
=== RUN   TestAdminMutationRefusalMatrix/R09_a_student_is_refused_on_every_mutation/save
=== RUN   TestAdminMutationRefusalMatrix/R09_a_student_is_refused_on_every_mutation/status
=== RUN   TestAdminMutationRefusalMatrix/R09_a_student_is_refused_on_every_mutation/delete
=== RUN   TestAdminMutationRefusalMatrix/R10_cross-site_mutation_is_refused_before_authentication
=== RUN   TestAdminMutationRefusalMatrix/R10_cross-site_mutation_is_refused_before_authentication/create
=== RUN   TestAdminMutationRefusalMatrix/R10_cross-site_mutation_is_refused_before_authentication/save
=== RUN   TestAdminMutationRefusalMatrix/R10_cross-site_mutation_is_refused_before_authentication/status
=== RUN   TestAdminMutationRefusalMatrix/R10_cross-site_mutation_is_refused_before_authentication/delete
=== RUN   TestAdminMutationRefusalMatrix/R11_a_foreign_instructor_cannot_save_transition_or_delete
=== RUN   TestAdminMutationRefusalMatrix/R11_a_foreign_instructor_cannot_save_transition_or_delete/save
=== RUN   TestAdminMutationRefusalMatrix/R11_a_foreign_instructor_cannot_save_transition_or_delete/status
=== RUN   TestAdminMutationRefusalMatrix/R11_a_foreign_instructor_cannot_save_transition_or_delete/delete
--- PASS: TestAdminMutationRefusalMatrix (0.89s)
    --- PASS: TestAdminMutationRefusalMatrix/R08_anonymous_is_refused_on_every_mutation (0.26s)
        --- PASS: TestAdminMutationRefusalMatrix/R08_anonymous_is_refused_on_every_mutation/create (0.07s)
        --- PASS: TestAdminMutationRefusalMatrix/R08_anonymous_is_refused_on_every_mutation/save (0.07s)
        --- PASS: TestAdminMutationRefusalMatrix/R08_anonymous_is_refused_on_every_mutation/status (0.06s)
        --- PASS: TestAdminMutationRefusalMatrix/R08_anonymous_is_refused_on_every_mutation/delete (0.06s)
    --- PASS: TestAdminMutationRefusalMatrix/R09_a_student_is_refused_on_every_mutation (0.24s)
        --- PASS: TestAdminMutationRefusalMatrix/R09_a_student_is_refused_on_every_mutation/create (0.06s)
        --- PASS: TestAdminMutationRefusalMatrix/R09_a_student_is_refused_on_every_mutation/save (0.06s)
        --- PASS: TestAdminMutationRefusalMatrix/R09_a_student_is_refused_on_every_mutation/status (0.06s)
        --- PASS: TestAdminMutationRefusalMatrix/R09_a_student_is_refused_on_every_mutation/delete (0.06s)
    --- PASS: TestAdminMutationRefusalMatrix/R10_cross-site_mutation_is_refused_before_authentication (0.23s)
        --- PASS: TestAdminMutationRefusalMatrix/R10_cross-site_mutation_is_refused_before_authentication/create (0.06s)
        --- PASS: TestAdminMutationRefusalMatrix/R10_cross-site_mutation_is_refused_before_authentication/save (0.06s)
        --- PASS: TestAdminMutationRefusalMatrix/R10_cross-site_mutation_is_refused_before_authentication/status (0.06s)
        --- PASS: TestAdminMutationRefusalMatrix/R10_cross-site_mutation_is_refused_before_authentication/delete (0.06s)
    --- PASS: TestAdminMutationRefusalMatrix/R11_a_foreign_instructor_cannot_save_transition_or_delete (0.17s)
        --- PASS: TestAdminMutationRefusalMatrix/R11_a_foreign_instructor_cannot_save_transition_or_delete/save (0.06s)
        --- PASS: TestAdminMutationRefusalMatrix/R11_a_foreign_instructor_cannot_save_transition_or_delete/status (0.06s)
        --- PASS: TestAdminMutationRefusalMatrix/R11_a_foreign_instructor_cannot_save_transition_or_delete/delete (0.06s)
=== RUN   TestAdminLifecycleRefusalContracts
=== RUN   TestAdminLifecycleRefusalContracts/R14_an_illegal_transition_is_a_named_conflict_and_leaves_status_unchanged
=== RUN   TestAdminLifecycleRefusalContracts/R15_published_and_archived_courses_cannot_be_deleted
=== RUN   TestAdminLifecycleRefusalContracts/R15_published_and_archived_courses_cannot_be_deleted/published
=== RUN   TestAdminLifecycleRefusalContracts/R15_published_and_archived_courses_cannot_be_deleted/archived
=== RUN   TestAdminLifecycleRefusalContracts/R23_a_transition_retires_the_previous_tree_token
--- PASS: TestAdminLifecycleRefusalContracts (0.30s)
    --- PASS: TestAdminLifecycleRefusalContracts/R14_an_illegal_transition_is_a_named_conflict_and_leaves_status_unchanged (0.06s)
    --- PASS: TestAdminLifecycleRefusalContracts/R15_published_and_archived_courses_cannot_be_deleted (0.12s)
        --- PASS: TestAdminLifecycleRefusalContracts/R15_published_and_archived_courses_cannot_be_deleted/published (0.06s)
        --- PASS: TestAdminLifecycleRefusalContracts/R15_published_and_archived_courses_cannot_be_deleted/archived (0.06s)
    --- PASS: TestAdminLifecycleRefusalContracts/R23_a_transition_retires_the_previous_tree_token (0.12s)
=== RUN   TestAdminWriteStrictJSON
=== RUN   TestAdminWriteStrictJSON/malformed
=== RUN   TestAdminWriteStrictJSON/duplicate
=== RUN   TestAdminWriteStrictJSON/unknown
=== RUN   TestAdminWriteStrictJSON/trailing
=== RUN   TestAdminWriteStrictJSON/expectedUpdatedAt_requires_six_fractional_digits
=== RUN   TestAdminWriteStrictJSON/query_parameters_are_rejected
=== RUN   TestAdminWriteStrictJSON/non-JSON_content_type_is_rejected
--- PASS: TestAdminWriteStrictJSON (0.35s)
    --- PASS: TestAdminWriteStrictJSON/malformed (0.06s)
    --- PASS: TestAdminWriteStrictJSON/duplicate (0.06s)
    --- PASS: TestAdminWriteStrictJSON/unknown (0.06s)
    --- PASS: TestAdminWriteStrictJSON/trailing (0.06s)
    --- PASS: TestAdminWriteStrictJSON/expectedUpdatedAt_requires_six_fractional_digits (0.06s)
    --- PASS: TestAdminWriteStrictJSON/query_parameters_are_rejected (0.06s)
    --- PASS: TestAdminWriteStrictJSON/non-JSON_content_type_is_rejected (0.00s)
=== RUN   TestAdminReadWrongMethod
--- PASS: TestAdminReadWrongMethod (0.00s)
=== RUN   TestContractCeilingsEqualTheDomainConstants
=== PAUSE TestContractCeilingsEqualTheDomainConstants
=== RUN   TestCatalogHTTPMatchesOpenAPI
=== RUN   TestCatalogHTTPMatchesOpenAPI/default_list
=== RUN   TestCatalogHTTPMatchesOpenAPI/boundaries
=== RUN   TestCatalogHTTPMatchesOpenAPI/empty_list
=== RUN   TestCatalogHTTPMatchesOpenAPI/detail
=== RUN   TestCatalogHTTPMatchesOpenAPI/missing_detail
=== RUN   TestCatalogHTTPMatchesOpenAPI/draft_hidden
=== RUN   TestCatalogHTTPMatchesOpenAPI/archived_hidden
=== RUN   TestCatalogHTTPMatchesOpenAPI/bad_limit
=== RUN   TestCatalogHTTPMatchesOpenAPI/repeated_parameter
=== RUN   TestCatalogHTTPMatchesOpenAPI/unknown_parameter
=== RUN   TestCatalogHTTPMatchesOpenAPI/decimal
=== RUN   TestCatalogHTTPMatchesOpenAPI/negative
=== RUN   TestCatalogHTTPMatchesOpenAPI/empty_numeric
=== RUN   TestCatalogHTTPMatchesOpenAPI/huge_numeric
=== RUN   TestCatalogHTTPMatchesOpenAPI/plus_prefix
=== RUN   TestCatalogHTTPMatchesOpenAPI/max_limit
=== RUN   TestCatalogHTTPMatchesOpenAPI/max_offset
=== RUN   TestCatalogHTTPMatchesOpenAPI/long_Unicode
=== RUN   TestCatalogHTTPMatchesOpenAPI/unknown_sort
=== RUN   TestCatalogHTTPMatchesOpenAPI/empty_sort
=== RUN   TestCatalogHTTPMatchesOpenAPI/bad_encoding
=== RUN   TestCatalogHTTPMatchesOpenAPI/NUL_input
=== RUN   TestCatalogHTTPMatchesOpenAPI/invalid_slug
=== RUN   TestCatalogHTTPMatchesOpenAPI/detail_query_rejected
=== RUN   TestCatalogHTTPMatchesOpenAPI/dependency
=== RUN   TestCatalogHTTPMatchesOpenAPI/unexpected
--- PASS: TestCatalogHTTPMatchesOpenAPI (0.06s)
    --- PASS: TestCatalogHTTPMatchesOpenAPI/default_list (0.00s)
    --- PASS: TestCatalogHTTPMatchesOpenAPI/boundaries (0.00s)
    --- PASS: TestCatalogHTTPMatchesOpenAPI/empty_list (0.00s)
    --- PASS: TestCatalogHTTPMatchesOpenAPI/detail (0.00s)
    --- PASS: TestCatalogHTTPMatchesOpenAPI/missing_detail (0.00s)
    --- PASS: TestCatalogHTTPMatchesOpenAPI/draft_hidden (0.00s)
    --- PASS: TestCatalogHTTPMatchesOpenAPI/archived_hidden (0.00s)
    --- PASS: TestCatalogHTTPMatchesOpenAPI/bad_limit (0.00s)
    --- PASS: TestCatalogHTTPMatchesOpenAPI/repeated_parameter (0.00s)
    --- PASS: TestCatalogHTTPMatchesOpenAPI/unknown_parameter (0.00s)
    --- PASS: TestCatalogHTTPMatchesOpenAPI/decimal (0.00s)
    --- PASS: TestCatalogHTTPMatchesOpenAPI/negative (0.00s)
    --- PASS: TestCatalogHTTPMatchesOpenAPI/empty_numeric (0.00s)
    --- PASS: TestCatalogHTTPMatchesOpenAPI/huge_numeric (0.00s)
    --- PASS: TestCatalogHTTPMatchesOpenAPI/plus_prefix (0.00s)
    --- PASS: TestCatalogHTTPMatchesOpenAPI/max_limit (0.00s)
    --- PASS: TestCatalogHTTPMatchesOpenAPI/max_offset (0.00s)
    --- PASS: TestCatalogHTTPMatchesOpenAPI/long_Unicode (0.00s)
    --- PASS: TestCatalogHTTPMatchesOpenAPI/unknown_sort (0.00s)
    --- PASS: TestCatalogHTTPMatchesOpenAPI/empty_sort (0.00s)
    --- PASS: TestCatalogHTTPMatchesOpenAPI/bad_encoding (0.00s)
    --- PASS: TestCatalogHTTPMatchesOpenAPI/NUL_input (0.00s)
    --- PASS: TestCatalogHTTPMatchesOpenAPI/invalid_slug (0.00s)
    --- PASS: TestCatalogHTTPMatchesOpenAPI/detail_query_rejected (0.00s)
    --- PASS: TestCatalogHTTPMatchesOpenAPI/dependency (0.00s)
    --- PASS: TestCatalogHTTPMatchesOpenAPI/unexpected (0.00s)
=== RUN   TestListPreservesDefaultsAndRequestedPagination
--- PASS: TestListPreservesDefaultsAndRequestedPagination (0.00s)
=== RUN   TestPublishedCourseDetailExposesOnlyOrderedProgramTitles
--- PASS: TestPublishedCourseDetailExposesOnlyOrderedProgramTitles (0.00s)
=== RUN   TestUnpublishedCourseProgramsAreNotExposed
--- PASS: TestUnpublishedCourseProgramsAreNotExposed (0.00s)
=== RUN   TestLearningHTTPMatchesOpenAPIAndConcreteProgressContract
--- PASS: TestLearningHTTPMatchesOpenAPIAndConcreteProgressContract (0.06s)
=== RUN   TestLearningHTTPBoundariesAndDenials
=== RUN   TestLearningHTTPBoundariesAndDenials/authentication_required
=== RUN   TestLearningHTTPBoundariesAndDenials/invalid_slug_before_repository
=== RUN   TestLearningHTTPBoundariesAndDenials/course_not_found
=== RUN   TestLearningHTTPBoundariesAndDenials/R17_unavailable_course
=== RUN   TestLearningHTTPBoundariesAndDenials/dependency_unavailable
=== RUN   TestLearningHTTPBoundariesAndDenials/query_rejected
--- PASS: TestLearningHTTPBoundariesAndDenials (0.29s)
    --- PASS: TestLearningHTTPBoundariesAndDenials/authentication_required (0.00s)
    --- PASS: TestLearningHTTPBoundariesAndDenials/invalid_slug_before_repository (0.06s)
    --- PASS: TestLearningHTTPBoundariesAndDenials/course_not_found (0.06s)
    --- PASS: TestLearningHTTPBoundariesAndDenials/R17_unavailable_course (0.06s)
    --- PASS: TestLearningHTTPBoundariesAndDenials/dependency_unavailable (0.06s)
    --- PASS: TestLearningHTTPBoundariesAndDenials/query_rejected (0.06s)
=== RUN   TestPay19LearningOutlineLocksPaidLessonsInsteadOfRefusing
--- PASS: TestPay19LearningOutlineLocksPaidLessonsInsteadOfRefusing (0.06s)
=== RUN   TestR09LearningForeignDraftDeniedAndPublishedAllowed
=== RUN   TestR09LearningForeignDraftDeniedAndPublishedAllowed/stranger_draft
=== RUN   TestR09LearningForeignDraftDeniedAndPublishedAllowed/stranger_published
--- PASS: TestR09LearningForeignDraftDeniedAndPublishedAllowed (0.12s)
    --- PASS: TestR09LearningForeignDraftDeniedAndPublishedAllowed/stranger_draft (0.06s)
    --- PASS: TestR09LearningForeignDraftDeniedAndPublishedAllowed/stranger_published (0.06s)
=== RUN   TestB01LearningCourseWithoutLessonsUsesEmptyArraysAndNullContinue
--- PASS: TestB01LearningCourseWithoutLessonsUsesEmptyArraysAndNullContinue (0.00s)
=== RUN   TestR21AnonymousPublicCourseProgramStillExposesNoLessonIDs
--- PASS: TestR21AnonymousPublicCourseProgramStillExposesNoLessonIDs (0.00s)
=== RUN   TestAdminStatusRequestMatchesGeneratedOpenAPI
=== RUN   TestAdminStatusRequestMatchesGeneratedOpenAPI/accepted_publish
=== RUN   TestAdminStatusRequestMatchesGeneratedOpenAPI/null_status
=== RUN   TestAdminStatusRequestMatchesGeneratedOpenAPI/null_expectedUpdatedAt
=== RUN   TestAdminStatusRequestMatchesGeneratedOpenAPI/missing_status
=== RUN   TestAdminStatusRequestMatchesGeneratedOpenAPI/missing_expectedUpdatedAt
--- PASS: TestAdminStatusRequestMatchesGeneratedOpenAPI (0.08s)
    --- PASS: TestAdminStatusRequestMatchesGeneratedOpenAPI/accepted_publish (0.00s)
    --- PASS: TestAdminStatusRequestMatchesGeneratedOpenAPI/null_status (0.00s)
    --- PASS: TestAdminStatusRequestMatchesGeneratedOpenAPI/null_expectedUpdatedAt (0.00s)
    --- PASS: TestAdminStatusRequestMatchesGeneratedOpenAPI/missing_status (0.00s)
    --- PASS: TestAdminStatusRequestMatchesGeneratedOpenAPI/missing_expectedUpdatedAt (0.00s)
=== RUN   TestCatalogHandlersMatchGeneratedOpenAPI
=== RUN   TestCatalogHandlersMatchGeneratedOpenAPI//api/v1/courses
=== RUN   TestCatalogHandlersMatchGeneratedOpenAPI//api/v1/courses?q=empty
=== RUN   TestCatalogHandlersMatchGeneratedOpenAPI//api/v1/courses/go-foundations
=== RUN   TestCatalogHandlersMatchGeneratedOpenAPI//api/v1/courses/missing
=== RUN   TestCatalogHandlersMatchGeneratedOpenAPI//api/v1/courses?limit=0
--- PASS: TestCatalogHandlersMatchGeneratedOpenAPI (0.08s)
    --- PASS: TestCatalogHandlersMatchGeneratedOpenAPI//api/v1/courses (0.00s)
    --- PASS: TestCatalogHandlersMatchGeneratedOpenAPI//api/v1/courses?q=empty (0.00s)
    --- PASS: TestCatalogHandlersMatchGeneratedOpenAPI//api/v1/courses/go-foundations (0.00s)
    --- PASS: TestCatalogHandlersMatchGeneratedOpenAPI//api/v1/courses/missing (0.00s)
    --- PASS: TestCatalogHandlersMatchGeneratedOpenAPI//api/v1/courses?limit=0 (0.00s)
=== CONT  TestContractCeilingsEqualTheDomainConstants
=== RUN   TestContractCeilingsEqualTheDomainConstants/AdminModule.lessons
=== PAUSE TestContractCeilingsEqualTheDomainConstants/AdminModule.lessons
=== RUN   TestContractCeilingsEqualTheDomainConstants/LearningModule.lessons
=== PAUSE TestContractCeilingsEqualTheDomainConstants/LearningModule.lessons
=== RUN   TestContractCeilingsEqualTheDomainConstants/ProgramModule.lessons
=== PAUSE TestContractCeilingsEqualTheDomainConstants/ProgramModule.lessons
=== RUN   TestContractCeilingsEqualTheDomainConstants/AdminCourseTree.modules
=== PAUSE TestContractCeilingsEqualTheDomainConstants/AdminCourseTree.modules
=== RUN   TestContractCeilingsEqualTheDomainConstants/AdminCourseSaveRequest.modules
=== PAUSE TestContractCeilingsEqualTheDomainConstants/AdminCourseSaveRequest.modules
=== RUN   TestContractCeilingsEqualTheDomainConstants/LearningCourse.modules
=== PAUSE TestContractCeilingsEqualTheDomainConstants/LearningCourse.modules
=== RUN   TestContractCeilingsEqualTheDomainConstants/CourseDetail.program
=== PAUSE TestContractCeilingsEqualTheDomainConstants/CourseDetail.program
=== RUN   TestContractCeilingsEqualTheDomainConstants/LessonTranscript.segments
=== PAUSE TestContractCeilingsEqualTheDomainConstants/LessonTranscript.segments
=== RUN   TestContractCeilingsEqualTheDomainConstants/TranscriptSegment.documentIds
=== PAUSE TestContractCeilingsEqualTheDomainConstants/TranscriptSegment.documentIds
=== RUN   TestContractCeilingsEqualTheDomainConstants/LessonTranscript.documents
=== PAUSE TestContractCeilingsEqualTheDomainConstants/LessonTranscript.documents
=== RUN   TestContractCeilingsEqualTheDomainConstants/TranscriptSegment.text
=== PAUSE TestContractCeilingsEqualTheDomainConstants/TranscriptSegment.text
=== RUN   TestContractCeilingsEqualTheDomainConstants/the_course-wide_lesson_total_is_owned_by_the_domain_alone
=== PAUSE TestContractCeilingsEqualTheDomainConstants/the_course-wide_lesson_total_is_owned_by_the_domain_alone
=== CONT  TestContractCeilingsEqualTheDomainConstants/AdminModule.lessons
=== CONT  TestContractCeilingsEqualTheDomainConstants/CourseDetail.program
=== CONT  TestContractCeilingsEqualTheDomainConstants/AdminCourseTree.modules
=== CONT  TestContractCeilingsEqualTheDomainConstants/LessonTranscript.documents
=== CONT  TestContractCeilingsEqualTheDomainConstants/LearningCourse.modules
=== CONT  TestContractCeilingsEqualTheDomainConstants/the_course-wide_lesson_total_is_owned_by_the_domain_alone
=== CONT  TestContractCeilingsEqualTheDomainConstants/TranscriptSegment.text
=== CONT  TestContractCeilingsEqualTheDomainConstants/AdminCourseSaveRequest.modules
=== CONT  TestContractCeilingsEqualTheDomainConstants/ProgramModule.lessons
=== CONT  TestContractCeilingsEqualTheDomainConstants/TranscriptSegment.documentIds
=== CONT  TestContractCeilingsEqualTheDomainConstants/LearningModule.lessons
=== CONT  TestContractCeilingsEqualTheDomainConstants/LessonTranscript.segments
--- PASS: TestContractCeilingsEqualTheDomainConstants (0.03s)
    --- PASS: TestContractCeilingsEqualTheDomainConstants/AdminModule.lessons (0.00s)
    --- PASS: TestContractCeilingsEqualTheDomainConstants/LessonTranscript.documents (0.00s)
    --- PASS: TestContractCeilingsEqualTheDomainConstants/CourseDetail.program (0.00s)
    --- PASS: TestContractCeilingsEqualTheDomainConstants/AdminCourseTree.modules (0.00s)
    --- PASS: TestContractCeilingsEqualTheDomainConstants/LearningCourse.modules (0.00s)
    --- PASS: TestContractCeilingsEqualTheDomainConstants/AdminCourseSaveRequest.modules (0.00s)
    --- PASS: TestContractCeilingsEqualTheDomainConstants/ProgramModule.lessons (0.00s)
    --- PASS: TestContractCeilingsEqualTheDomainConstants/TranscriptSegment.documentIds (0.00s)
    --- PASS: TestContractCeilingsEqualTheDomainConstants/LearningModule.lessons (0.00s)
    --- PASS: TestContractCeilingsEqualTheDomainConstants/LessonTranscript.segments (0.00s)
    --- PASS: TestContractCeilingsEqualTheDomainConstants/TranscriptSegment.text (0.00s)
    --- PASS: TestContractCeilingsEqualTheDomainConstants/the_course-wide_lesson_total_is_owned_by_the_domain_alone (0.00s)
PASS
ok  	courses/internal/course/adapters/http	7.356s
```
EXIT=0
ФАКТ: git diff --check EXIT=0; self-review підтвердив нуль business-module imports у генераторі (лише shared platform/httpserver). Побачено косметичне import grouping; вирівняю після поточного gate, не мутую source під час перевірки.
ФАКТ: Перший повний ordered gate завершився EXIT=0 на кожному кроці; повний stdout збережений вище. gofmt stdout порожній, vet/build stdout порожній, affected tests PASS. Self-review до фінального коміту.
ФАКТ: Self-review завершено: import groups вирівняно; прибрано unused symbol.site. Додано fail-closed guard на зміну canonical domain Error.Error(): method мусить бути identity string conversion, інакше старий literal не можна вважати wire code. Новий mutation fixture перевіряє цей випадок. Це в межах error-code source семантики; інших const-set не чіпає.
РІШЕННЯ: 2026-10-27, Codex — bounded grammar limitation: single-definition locals, exhaustive switches і closed unexported wrapper callers підтримані; довільні обчислення, open error params, assignment ambiguity/address escape, wrapper/function escape і cycles fail-closed з діагностикою. Це не універсальний Go interpreter. Перевірка: errors_test.go negative fixtures; поточні 166 calls повністю покрито.
ДАЛІ: Повторити ordered gates на фінальних bytes після self-review delta, потім атомарний коміт parser+hand+overlay+generated+evidence.

ДАЛІ: FINAL ORDERED GATE, cwd=backend: `gofmt -l internal/platform/openapigen/errors.go internal/platform/openapigen/errors_test.go internal/platform/openapigen/error_transport.go internal/platform/openapigen/generate.go internal/platform/openapigen/overlay.go internal/course/adapters/http/openapi_contract_test.go internal/course/adapters/http/openapi_generated_request_test.go internal/course/adapters/http/openapi_generated_response_test.go`
ФАКТ: Реальний вивід:
```text
```
EXIT=0

ДАЛІ: FINAL ORDERED GATE, cwd=backend: `go vet ./internal/platform/openapigen ./internal/course/adapters/http`
ФАКТ: Реальний вивід:
```text
```
EXIT=0

ДАЛІ: FINAL ORDERED GATE, cwd=backend: `go build ./...`
ФАКТ: Реальний вивід:
```text
```
EXIT=0

ДАЛІ: FINAL ORDERED GATE, cwd=backend: `go test -count=1 -v ./internal/platform/openapigen ./internal/course/adapters/http`
ФАКТ: Реальний вивід:
```text
=== RUN   TestErrorCodeForms
=== RUN   TestErrorCodeForms/literal
=== RUN   TestErrorCodeForms/cast
=== RUN   TestErrorCodeForms/direct_typed
=== RUN   TestErrorCodeForms/finite_Error_variable
=== RUN   TestErrorCodeForms/sentinel_variable
=== RUN   TestErrorCodeForms/error_forwarding
=== RUN   TestErrorCodeForms/sentinel_method
=== RUN   TestErrorCodeForms/wrapper
=== RUN   TestErrorCodeForms/switch
=== RUN   TestErrorCodeForms/guard
--- PASS: TestErrorCodeForms (0.02s)
    --- PASS: TestErrorCodeForms/literal (0.00s)
    --- PASS: TestErrorCodeForms/cast (0.00s)
    --- PASS: TestErrorCodeForms/direct_typed (0.00s)
    --- PASS: TestErrorCodeForms/finite_Error_variable (0.00s)
    --- PASS: TestErrorCodeForms/sentinel_variable (0.00s)
    --- PASS: TestErrorCodeForms/error_forwarding (0.00s)
    --- PASS: TestErrorCodeForms/sentinel_method (0.00s)
    --- PASS: TestErrorCodeForms/wrapper (0.00s)
    --- PASS: TestErrorCodeForms/switch (0.00s)
    --- PASS: TestErrorCodeForms/guard (0.00s)
=== RUN   TestErrorCodeFailClosed
=== RUN   TestErrorCodeFailClosed/unknown_function
=== RUN   TestErrorCodeFailClosed/typo
=== RUN   TestErrorCodeFailClosed/missing_symbol
=== RUN   TestErrorCodeFailClosed/unbounded_Error
=== RUN   TestErrorCodeFailClosed/unknown_wrapper_caller
=== RUN   TestErrorCodeFailClosed/escaped_writer
=== RUN   TestErrorCodeFailClosed/escaped_wrapper
=== RUN   TestErrorCodeFailClosed/nonexhaustive_switch
=== RUN   TestErrorCodeFailClosed/overwrite
=== RUN   TestErrorCodeFailClosed/conditional_assignment
=== RUN   TestErrorCodeFailClosed/fallthrough
=== RUN   TestErrorCodeFailClosed/guard_mutation
=== RUN   TestErrorCodeFailClosed/address_escape
=== RUN   TestErrorCodeFailClosed/raw_error_argument
=== RUN   TestErrorCodeFailClosed/cycle
--- PASS: TestErrorCodeFailClosed (0.02s)
    --- PASS: TestErrorCodeFailClosed/unknown_function (0.00s)
    --- PASS: TestErrorCodeFailClosed/typo (0.00s)
    --- PASS: TestErrorCodeFailClosed/missing_symbol (0.00s)
    --- PASS: TestErrorCodeFailClosed/unbounded_Error (0.00s)
    --- PASS: TestErrorCodeFailClosed/unknown_wrapper_caller (0.00s)
    --- PASS: TestErrorCodeFailClosed/escaped_writer (0.00s)
    --- PASS: TestErrorCodeFailClosed/escaped_wrapper (0.00s)
    --- PASS: TestErrorCodeFailClosed/nonexhaustive_switch (0.00s)
    --- PASS: TestErrorCodeFailClosed/overwrite (0.00s)
    --- PASS: TestErrorCodeFailClosed/conditional_assignment (0.00s)
    --- PASS: TestErrorCodeFailClosed/fallthrough (0.00s)
    --- PASS: TestErrorCodeFailClosed/guard_mutation (0.00s)
    --- PASS: TestErrorCodeFailClosed/address_escape (0.00s)
    --- PASS: TestErrorCodeFailClosed/raw_error_argument (0.00s)
    --- PASS: TestErrorCodeFailClosed/cycle (0.00s)
=== RUN   TestTransportProvenance
--- PASS: TestTransportProvenance (0.00s)
=== RUN   TestTreeErrorCensus
    errors_test.go:164: WriteError sites=166 domain codes=116 emitted=109
        admin_immutable, admin_not_learner, already_running, authentication_required, cannot_target_self, checkout_already_opened, comment_body_invalid, comment_cursor_invalid, comment_forbidden, comment_not_found, comment_parent_invalid, comment_position_invalid, comment_removed, comment_thread_full, course_already_owned, course_forbidden, course_modified, course_not_deletable, course_not_enrollable, course_not_found, course_purchase_required, dependency_unavailable, document_conflict, email_not_verified, email_taken, enrollment_blocked, enrollment_forbidden, enrollment_required, forbidden, grant_forbidden, grant_self_forbidden, grant_transition_invalid, image_too_large, image_too_many_pixels, insufficient_storage, internal_error, invalid_account_token, invalid_callback, invalid_credentials, invalid_display_name, invalid_email, invalid_image, invalid_lesson_document, invalid_owner, invalid_query, invalid_range, invalid_request, invalid_slug, invalid_status_transition, invalid_totp, last_admin, learner_course_forbidden, lesson_limit, lesson_not_found, membership_conflict, method_not_allowed, not_found, order_not_found, origin_forbidden, password_unchanged, payment_event_rejected, playback_forbidden, playback_session_expired, playback_taken_over, precondition_failed, range_too_wide, rate_limited, recovery_code_invalid, report_already_open, report_forbidden, report_invalid, report_own_comment, report_own_person, report_own_review, request_too_large, review_already_exists, review_invalid, review_moderation_forbidden, review_not_available, review_not_found, review_not_started, review_owner_forbidden, review_purchase_required, review_removed, slug_immutable, slug_taken, token_used, transcript_document_exists, transcript_document_limit, transcript_document_referenced, transcript_exists, transcript_modified, transcript_segment_limit, transcript_text_limit, twofactor_already_enabled, twofactor_not_enabled, unknown_bucket, unknown_metric, unsupported_image, unsupported_media, unsupported_media_type, upload_in_progress, upload_offset_conflict, upload_too_large, user_not_found, video_not_found, video_not_ready, video_not_retryable, weak_password
--- PASS: TestTreeErrorCensus (0.12s)
=== RUN   TestErrorSourceDiscovery
=== RUN   TestErrorSourceDiscovery/unused_domain_code
=== RUN   TestErrorSourceDiscovery/unknown_initializer
=== RUN   TestErrorSourceDiscovery/computed_constant
--- PASS: TestErrorSourceDiscovery (0.00s)
    --- PASS: TestErrorSourceDiscovery/unused_domain_code (0.00s)
    --- PASS: TestErrorSourceDiscovery/unknown_initializer (0.00s)
    --- PASS: TestErrorSourceDiscovery/computed_constant (0.00s)
=== RUN   TestErrorSchemaRequiresSource
--- PASS: TestErrorSchemaRequiresSource (0.00s)
=== RUN   TestOverlayCannotSupplyErrorEnum
--- PASS: TestOverlayCannotSupplyErrorEnum (0.00s)
=== RUN   TestChangedErrorMethodFailsClosed
--- PASS: TestChangedErrorMethodFailsClosed (0.00s)
=== RUN   TestGenerateTwiceIsByteIdentical
--- PASS: TestGenerateTwiceIsByteIdentical (0.00s)
=== RUN   TestFailClosed
=== RUN   TestFailClosed/marshaler
=== RUN   TestFailClosed/pointer_marshaler
=== RUN   TestFailClosed/enum
=== RUN   TestFailClosed/slice_decision
=== RUN   TestFailClosed/orphan_schema
=== RUN   TestFailClosed/orphan_field
=== RUN   TestFailClosed/orphan_path
=== RUN   TestFailClosed/unknown_keyword
=== RUN   TestFailClosed/malformed
=== RUN   TestFailClosed/empty
=== RUN   TestFailClosed/multiple_documents
=== RUN   TestFailClosed/duplicate_YAML_key
--- PASS: TestFailClosed (0.01s)
    --- PASS: TestFailClosed/marshaler (0.00s)
    --- PASS: TestFailClosed/pointer_marshaler (0.00s)
    --- PASS: TestFailClosed/enum (0.00s)
    --- PASS: TestFailClosed/slice_decision (0.00s)
    --- PASS: TestFailClosed/orphan_schema (0.00s)
    --- PASS: TestFailClosed/orphan_field (0.00s)
    --- PASS: TestFailClosed/orphan_path (0.00s)
    --- PASS: TestFailClosed/unknown_keyword (0.00s)
    --- PASS: TestFailClosed/malformed (0.00s)
    --- PASS: TestFailClosed/empty (0.00s)
    --- PASS: TestFailClosed/multiple_documents (0.00s)
    --- PASS: TestFailClosed/duplicate_YAML_key (0.00s)
=== RUN   TestMissingOverlay
--- PASS: TestMissingOverlay (0.00s)
=== RUN   TestNonInjectiveNameMap
--- PASS: TestNonInjectiveNameMap (0.00s)
=== RUN   TestRegisteredEnumAndOverride
--- PASS: TestRegisteredEnumAndOverride (0.00s)
=== RUN   TestNullableSliceDecision
--- PASS: TestNullableSliceDecision (0.00s)
=== RUN   TestStructMarshalerOverrideIsScalarComponent
--- PASS: TestStructMarshalerOverrideIsScalarComponent (0.00s)
=== RUN   TestUntaggedJSONFieldsAndIgnoredFields
--- PASS: TestUntaggedJSONFieldsAndIgnoredFields (0.00s)
=== RUN   TestParametersFromGoWithMetadata
--- PASS: TestParametersFromGoWithMetadata (0.00s)
=== RUN   TestParameterMetadataFailsClosed
=== RUN   TestParameterMetadataFailsClosed/orphan
=== RUN   TestParameterMetadataFailsClosed/array_replacement
=== RUN   TestParameterMetadataFailsClosed/delete_parameters
=== RUN   TestParameterMetadataFailsClosed/delete_parameter
=== RUN   TestParameterMetadataFailsClosed/name
=== RUN   TestParameterMetadataFailsClosed/in
=== RUN   TestParameterMetadataFailsClosed/required
=== RUN   TestParameterMetadataFailsClosed/type
=== RUN   TestParameterMetadataFailsClosed/enum
=== RUN   TestParameterMetadataFailsClosed/ref
=== RUN   TestParameterMetadataFailsClosed/unknown_validation
=== RUN   TestParameterMetadataFailsClosed/no_Go_parameters
--- PASS: TestParameterMetadataFailsClosed (0.01s)
    --- PASS: TestParameterMetadataFailsClosed/orphan (0.00s)
    --- PASS: TestParameterMetadataFailsClosed/array_replacement (0.00s)
    --- PASS: TestParameterMetadataFailsClosed/delete_parameters (0.00s)
    --- PASS: TestParameterMetadataFailsClosed/delete_parameter (0.00s)
    --- PASS: TestParameterMetadataFailsClosed/name (0.00s)
    --- PASS: TestParameterMetadataFailsClosed/in (0.00s)
    --- PASS: TestParameterMetadataFailsClosed/required (0.00s)
    --- PASS: TestParameterMetadataFailsClosed/type (0.00s)
    --- PASS: TestParameterMetadataFailsClosed/enum (0.00s)
    --- PASS: TestParameterMetadataFailsClosed/ref (0.00s)
    --- PASS: TestParameterMetadataFailsClosed/unknown_validation (0.00s)
    --- PASS: TestParameterMetadataFailsClosed/no_Go_parameters (0.00s)
=== RUN   TestInvalidGoParameters
--- PASS: TestInvalidGoParameters (0.00s)
=== RUN   TestParameterMetadataDoesNotInterceptResponseField
--- PASS: TestParameterMetadataDoesNotInterceptResponseField (0.00s)
=== RUN   TestRequestPointerDecisions
=== RUN   TestRequestPointerDecisions/missing_decision_Child
=== RUN   TestRequestPointerDecisions/missing_decision_Required
=== RUN   TestRequestPointerDecisions/missing_decision_Optional
--- PASS: TestRequestPointerDecisions (0.00s)
    --- PASS: TestRequestPointerDecisions/missing_decision_Child (0.00s)
    --- PASS: TestRequestPointerDecisions/missing_decision_Required (0.00s)
    --- PASS: TestRequestPointerDecisions/missing_decision_Optional (0.00s)
=== RUN   TestSharedRequestResponseCannotBypassDecisions
--- PASS: TestSharedRequestResponseCannotBypassDecisions (0.00s)
=== RUN   TestScalarFieldReference
=== RUN   TestScalarFieldReference/matching_scalar
=== RUN   TestScalarFieldReference/wrong_scalar
=== RUN   TestScalarFieldReference/nullable_target
--- PASS: TestScalarFieldReference (0.00s)
    --- PASS: TestScalarFieldReference/matching_scalar (0.00s)
    --- PASS: TestScalarFieldReference/wrong_scalar (0.00s)
    --- PASS: TestScalarFieldReference/nullable_target (0.00s)
=== RUN   TestParameterFormatMetadata
--- PASS: TestParameterFormatMetadata (0.00s)
PASS
ok  	courses/internal/platform/openapigen	1.552s
=== RUN   TestDocumentInvalidDraftNeverPersists
=== RUN   TestDocumentInvalidDraftNeverPersists/{"id":"bad","type":"iframe"}
=== RUN   TestDocumentInvalidDraftNeverPersists/{"id":"v","type":"video","content":{"blocks":[]}}
=== RUN   TestDocumentInvalidDraftNeverPersists/{"id":"row","type":"row","columns":[{"id":"nested","type":"row","columns":[]}]}
--- PASS: TestDocumentInvalidDraftNeverPersists (0.00s)
    --- PASS: TestDocumentInvalidDraftNeverPersists/{"id":"bad","type":"iframe"} (0.00s)
    --- PASS: TestDocumentInvalidDraftNeverPersists/{"id":"v","type":"video","content":{"blocks":[]}} (0.00s)
    --- PASS: TestDocumentInvalidDraftNeverPersists/{"id":"row","type":"row","columns":[{"id":"nested","type":"row","columns":[]}]} (0.00s)
=== RUN   TestDocumentNonEditor
=== RUN   TestDocumentNonEditor/GET
=== RUN   TestDocumentNonEditor/PUT
=== RUN   TestDocumentNonEditor/POST
--- PASS: TestDocumentNonEditor (0.00s)
    --- PASS: TestDocumentNonEditor/GET (0.00s)
    --- PASS: TestDocumentNonEditor/PUT (0.00s)
    --- PASS: TestDocumentNonEditor/POST (0.00s)
=== RUN   TestDocumentPublishedAccess
=== RUN   TestDocumentPublishedAccess/paid_nonmember
=== RUN   TestDocumentPublishedAccess/active_member
=== RUN   TestDocumentPublishedAccess/email_first
=== RUN   TestDocumentPublishedAccess/absent_publication
=== RUN   TestDocumentPublishedAccess/verified_preview
--- PASS: TestDocumentPublishedAccess (0.00s)
    --- PASS: TestDocumentPublishedAccess/paid_nonmember (0.00s)
    --- PASS: TestDocumentPublishedAccess/active_member (0.00s)
    --- PASS: TestDocumentPublishedAccess/email_first (0.00s)
    --- PASS: TestDocumentPublishedAccess/absent_publication (0.00s)
    --- PASS: TestDocumentPublishedAccess/verified_preview (0.00s)
=== RUN   TestDocumentConflicts
=== RUN   TestDocumentConflicts/PUT
=== RUN   TestDocumentConflicts/POST
--- PASS: TestDocumentConflicts (0.00s)
    --- PASS: TestDocumentConflicts/PUT (0.00s)
    --- PASS: TestDocumentConflicts/POST (0.00s)
=== RUN   TestDocumentDraftRoundTrip
--- PASS: TestDocumentDraftRoundTrip (0.00s)
=== RUN   TestDocumentTransportRejections
=== RUN   TestDocumentTransportRejections/unknown_query
=== RUN   TestDocumentTransportRejections/unknown_locale
=== RUN   TestDocumentTransportRejections/empty_locale
=== RUN   TestDocumentTransportRejections/duplicate_locale
=== RUN   TestDocumentTransportRejections/malformed_query
=== RUN   TestDocumentTransportRejections/PUT_query
=== RUN   TestDocumentTransportRejections/publish_query
=== RUN   TestDocumentTransportRejections/unknown_wrapper
=== RUN   TestDocumentTransportRejections/unknown_nested_field
=== RUN   TestDocumentTransportRejections/duplicate_nested_field
=== RUN   TestDocumentTransportRejections/trailing_JSON
=== RUN   TestDocumentTransportRejections/null_wrapper
=== RUN   TestDocumentTransportRejections/bad_timestamp
=== RUN   TestDocumentTransportRejections/invalid_document_locale
=== RUN   TestDocumentTransportRejections/publish_missing_token
=== RUN   TestDocumentTransportRejections/publish_null_token
=== RUN   TestDocumentTransportRejections/publish_unknown_field
=== RUN   TestDocumentTransportRejections/publish_trailing_bytes
=== RUN   TestDocumentTransportRejections/publish_too_large
=== RUN   TestDocumentTransportRejections/PUT_too_large
--- PASS: TestDocumentTransportRejections (0.02s)
    --- PASS: TestDocumentTransportRejections/unknown_query (0.00s)
    --- PASS: TestDocumentTransportRejections/unknown_locale (0.00s)
    --- PASS: TestDocumentTransportRejections/empty_locale (0.00s)
    --- PASS: TestDocumentTransportRejections/duplicate_locale (0.00s)
    --- PASS: TestDocumentTransportRejections/malformed_query (0.00s)
    --- PASS: TestDocumentTransportRejections/PUT_query (0.00s)
    --- PASS: TestDocumentTransportRejections/publish_query (0.00s)
    --- PASS: TestDocumentTransportRejections/unknown_wrapper (0.00s)
    --- PASS: TestDocumentTransportRejections/unknown_nested_field (0.00s)
    --- PASS: TestDocumentTransportRejections/duplicate_nested_field (0.00s)
    --- PASS: TestDocumentTransportRejections/trailing_JSON (0.00s)
    --- PASS: TestDocumentTransportRejections/null_wrapper (0.00s)
    --- PASS: TestDocumentTransportRejections/bad_timestamp (0.00s)
    --- PASS: TestDocumentTransportRejections/invalid_document_locale (0.00s)
    --- PASS: TestDocumentTransportRejections/publish_missing_token (0.00s)
    --- PASS: TestDocumentTransportRejections/publish_null_token (0.00s)
    --- PASS: TestDocumentTransportRejections/publish_unknown_field (0.00s)
    --- PASS: TestDocumentTransportRejections/publish_trailing_bytes (0.00s)
    --- PASS: TestDocumentTransportRejections/publish_too_large (0.00s)
    --- PASS: TestDocumentTransportRejections/PUT_too_large (0.01s)
=== RUN   TestDocumentWriteLimitExceedsAdminLimit
--- PASS: TestDocumentWriteLimitExceedsAdminLimit (0.00s)
=== RUN   TestDocumentErrorMapping
=== RUN   TestDocumentErrorMapping/invalid_lesson_document
=== RUN   TestDocumentErrorMapping/invalid_request
=== RUN   TestDocumentErrorMapping/invalid_query
=== RUN   TestDocumentErrorMapping/email_not_verified
=== RUN   TestDocumentErrorMapping/course_purchase_required
=== RUN   TestDocumentErrorMapping/forbidden
=== RUN   TestDocumentErrorMapping/course_forbidden
=== RUN   TestDocumentErrorMapping/lesson_not_found
=== RUN   TestDocumentErrorMapping/document_conflict
=== RUN   TestDocumentErrorMapping/dependency_unavailable
=== RUN   TestDocumentErrorMapping/internal_error
--- PASS: TestDocumentErrorMapping (0.00s)
    --- PASS: TestDocumentErrorMapping/invalid_lesson_document (0.00s)
    --- PASS: TestDocumentErrorMapping/invalid_request (0.00s)
    --- PASS: TestDocumentErrorMapping/invalid_query (0.00s)
    --- PASS: TestDocumentErrorMapping/email_not_verified (0.00s)
    --- PASS: TestDocumentErrorMapping/course_purchase_required (0.00s)
    --- PASS: TestDocumentErrorMapping/forbidden (0.00s)
    --- PASS: TestDocumentErrorMapping/course_forbidden (0.00s)
    --- PASS: TestDocumentErrorMapping/lesson_not_found (0.00s)
    --- PASS: TestDocumentErrorMapping/document_conflict (0.00s)
    --- PASS: TestDocumentErrorMapping/dependency_unavailable (0.00s)
    --- PASS: TestDocumentErrorMapping/internal_error (0.00s)
=== RUN   TestLearningOutlineHasNoDocumentBody
--- PASS: TestLearningOutlineHasNoDocumentBody (0.00s)
=== RUN   TestDocumentHTTPMatchesOpenAPI
=== RUN   TestDocumentHTTPMatchesOpenAPI/GET/api/v1/admin/lessons/20000000-0000-4000-8000-000000000001/document?locale=en
=== RUN   TestDocumentHTTPMatchesOpenAPI/PUT/api/v1/admin/lessons/20000000-0000-4000-8000-000000000001/document
=== RUN   TestDocumentHTTPMatchesOpenAPI/POST/api/v1/admin/lessons/20000000-0000-4000-8000-000000000001/document/publish
=== RUN   TestDocumentHTTPMatchesOpenAPI/GET/api/v1/learn/lessons/20000000-0000-4000-8000-000000000001/document?locale=en
--- PASS: TestDocumentHTTPMatchesOpenAPI (0.07s)
    --- PASS: TestDocumentHTTPMatchesOpenAPI/GET/api/v1/admin/lessons/20000000-0000-4000-8000-000000000001/document?locale=en (0.00s)
    --- PASS: TestDocumentHTTPMatchesOpenAPI/PUT/api/v1/admin/lessons/20000000-0000-4000-8000-000000000001/document (0.00s)
    --- PASS: TestDocumentHTTPMatchesOpenAPI/POST/api/v1/admin/lessons/20000000-0000-4000-8000-000000000001/document/publish (0.00s)
    --- PASS: TestDocumentHTTPMatchesOpenAPI/GET/api/v1/learn/lessons/20000000-0000-4000-8000-000000000001/document?locale=en (0.00s)
=== RUN   TestDocumentAuthenticationAndOriginWiring
=== RUN   TestDocumentAuthenticationAndOriginWiring/GET/api/v1/admin/lessons/20000000-0000-4000-8000-000000000001/document
=== RUN   TestDocumentAuthenticationAndOriginWiring/GET/api/v1/learn/lessons/20000000-0000-4000-8000-000000000001/document
=== RUN   TestDocumentAuthenticationAndOriginWiring/PUT/api/v1/admin/lessons/20000000-0000-4000-8000-000000000001/document
=== RUN   TestDocumentAuthenticationAndOriginWiring/POST/api/v1/admin/lessons/20000000-0000-4000-8000-000000000001/document/publish
--- PASS: TestDocumentAuthenticationAndOriginWiring (0.00s)
    --- PASS: TestDocumentAuthenticationAndOriginWiring/GET/api/v1/admin/lessons/20000000-0000-4000-8000-000000000001/document (0.00s)
    --- PASS: TestDocumentAuthenticationAndOriginWiring/GET/api/v1/learn/lessons/20000000-0000-4000-8000-000000000001/document (0.00s)
    --- PASS: TestDocumentAuthenticationAndOriginWiring/PUT/api/v1/admin/lessons/20000000-0000-4000-8000-000000000001/document (0.00s)
    --- PASS: TestDocumentAuthenticationAndOriginWiring/POST/api/v1/admin/lessons/20000000-0000-4000-8000-000000000001/document/publish (0.00s)
=== RUN   TestDocumentRecursiveSchemaRoundTrip
=== RUN   TestDocumentRecursiveSchemaRoundTrip/{"schemaVersion":1,"locale":"en","sections":[{"id":"row","type":"row","columns":[{"id":"text","type":"text","content":{"blocks":[{"type":"heading","level":2,"spans":[{"text":"Title","marks":["bold","italic"],"href":"https://example.com"}]},{"type":"list","ordered":true,"items":[[{"text":"Item"}]]}]}},{"id":"image","type":"image","imageId":"image-1","contentType":"image/png","alt":"Image"}]},{"id":"task","type":"task","content":{"blocks":[]}},{"id":"video","type":"video","content":null}]}
=== RUN   TestDocumentRecursiveSchemaRoundTrip/{"schemaVersion":1,"locale":"uk","sections":null}
=== RUN   TestDocumentRecursiveSchemaRoundTrip/{"schemaVersion":1,"locale":"uk","sections":[{"id":"text","type":"text","content":{"blocks":null}}]}
--- PASS: TestDocumentRecursiveSchemaRoundTrip (0.04s)
    --- PASS: TestDocumentRecursiveSchemaRoundTrip/{"schemaVersion":1,"locale":"en","sections":[{"id":"row","type":"row","columns":[{"id":"text","type":"text","content":{"blocks":[{"type":"heading","level":2,"spans":[{"text":"Title","marks":["bold","italic"],"href":"https://example.com"}]},{"type":"list","ordered":true,"items":[[{"text":"Item"}]]}]}},{"id":"image","type":"image","imageId":"image-1","contentType":"image/png","alt":"Image"}]},{"id":"task","type":"task","content":{"blocks":[]}},{"id":"video","type":"video","content":null}]} (0.00s)
    --- PASS: TestDocumentRecursiveSchemaRoundTrip/{"schemaVersion":1,"locale":"uk","sections":null} (0.00s)
    --- PASS: TestDocumentRecursiveSchemaRoundTrip/{"schemaVersion":1,"locale":"uk","sections":[{"id":"text","type":"text","content":{"blocks":null}}]} (0.00s)
=== RUN   TestErrorBoundaries
=== RUN   TestErrorBoundaries/writeAdminError/unknown
=== RUN   TestErrorBoundaries/writeDocumentError/unknown
=== RUN   TestErrorBoundaries/writeLearningError/unknown
=== RUN   TestErrorBoundaries/writeError/unknown
=== RUN   TestErrorBoundaries/writeError/invalid_query
=== RUN   TestErrorBoundaries/writeError/invalid_slug
=== RUN   TestErrorBoundaries/writeError/course_not_found
=== RUN   TestErrorBoundaries/writeError/dependency_unavailable
=== RUN   TestErrorBoundaries/writeDocumentError/invalid_lesson_document
=== RUN   TestErrorBoundaries/writeDocumentError/invalid_request
=== RUN   TestErrorBoundaries/writeDocumentError/invalid_query
=== RUN   TestErrorBoundaries/writeDocumentError/email_not_verified
=== RUN   TestErrorBoundaries/writeDocumentError/course_purchase_required
=== RUN   TestErrorBoundaries/writeDocumentError/forbidden
=== RUN   TestErrorBoundaries/writeDocumentError/course_forbidden
=== RUN   TestErrorBoundaries/writeDocumentError/lesson_not_found
=== RUN   TestErrorBoundaries/writeDocumentError/document_conflict
=== RUN   TestErrorBoundaries/writeDocumentError/dependency_unavailable
=== RUN   TestErrorBoundaries/writeAdminError/invalid_request
=== RUN   TestErrorBoundaries/writeAdminError/forbidden
=== RUN   TestErrorBoundaries/writeAdminError/course_forbidden
=== RUN   TestErrorBoundaries/writeAdminError/course_not_found
=== RUN   TestErrorBoundaries/writeAdminError/invalid_owner
=== RUN   TestErrorBoundaries/writeAdminError/slug_taken
=== RUN   TestErrorBoundaries/writeAdminError/slug_immutable
=== RUN   TestErrorBoundaries/writeAdminError/course_modified
=== RUN   TestErrorBoundaries/writeAdminError/invalid_status_transition
=== RUN   TestErrorBoundaries/writeAdminError/course_not_deletable
=== RUN   TestErrorBoundaries/writeAdminError/lesson_limit
=== RUN   TestErrorBoundaries/writeAdminError/dependency_unavailable
=== RUN   TestErrorBoundaries/writeLearningError/course_purchase_required
=== RUN   TestErrorBoundaries/writeLearningError/course_forbidden
=== RUN   TestErrorBoundaries/writeLearningError/course_forbidden#01
--- PASS: TestErrorBoundaries (0.00s)
    --- PASS: TestErrorBoundaries/writeAdminError/unknown (0.00s)
    --- PASS: TestErrorBoundaries/writeDocumentError/unknown (0.00s)
    --- PASS: TestErrorBoundaries/writeLearningError/unknown (0.00s)
    --- PASS: TestErrorBoundaries/writeError/unknown (0.00s)
    --- PASS: TestErrorBoundaries/writeError/invalid_query (0.00s)
    --- PASS: TestErrorBoundaries/writeError/invalid_slug (0.00s)
    --- PASS: TestErrorBoundaries/writeError/course_not_found (0.00s)
    --- PASS: TestErrorBoundaries/writeError/dependency_unavailable (0.00s)
    --- PASS: TestErrorBoundaries/writeDocumentError/invalid_lesson_document (0.00s)
    --- PASS: TestErrorBoundaries/writeDocumentError/invalid_request (0.00s)
    --- PASS: TestErrorBoundaries/writeDocumentError/invalid_query (0.00s)
    --- PASS: TestErrorBoundaries/writeDocumentError/email_not_verified (0.00s)
    --- PASS: TestErrorBoundaries/writeDocumentError/course_purchase_required (0.00s)
    --- PASS: TestErrorBoundaries/writeDocumentError/forbidden (0.00s)
    --- PASS: TestErrorBoundaries/writeDocumentError/course_forbidden (0.00s)
    --- PASS: TestErrorBoundaries/writeDocumentError/lesson_not_found (0.00s)
    --- PASS: TestErrorBoundaries/writeDocumentError/document_conflict (0.00s)
    --- PASS: TestErrorBoundaries/writeDocumentError/dependency_unavailable (0.00s)
    --- PASS: TestErrorBoundaries/writeAdminError/invalid_request (0.00s)
    --- PASS: TestErrorBoundaries/writeAdminError/forbidden (0.00s)
    --- PASS: TestErrorBoundaries/writeAdminError/course_forbidden (0.00s)
    --- PASS: TestErrorBoundaries/writeAdminError/course_not_found (0.00s)
    --- PASS: TestErrorBoundaries/writeAdminError/invalid_owner (0.00s)
    --- PASS: TestErrorBoundaries/writeAdminError/slug_taken (0.00s)
    --- PASS: TestErrorBoundaries/writeAdminError/slug_immutable (0.00s)
    --- PASS: TestErrorBoundaries/writeAdminError/course_modified (0.00s)
    --- PASS: TestErrorBoundaries/writeAdminError/invalid_status_transition (0.00s)
    --- PASS: TestErrorBoundaries/writeAdminError/course_not_deletable (0.00s)
    --- PASS: TestErrorBoundaries/writeAdminError/lesson_limit (0.00s)
    --- PASS: TestErrorBoundaries/writeAdminError/dependency_unavailable (0.00s)
    --- PASS: TestErrorBoundaries/writeLearningError/course_purchase_required (0.00s)
    --- PASS: TestErrorBoundaries/writeLearningError/course_forbidden (0.00s)
    --- PASS: TestErrorBoundaries/writeLearningError/course_forbidden#01 (0.00s)
=== RUN   TestCatalogOpenAPIFidelity
    openapi_contract_test.go:55: error enum ref-closure mutation detected; 109 emitted codes compared
    openapi_contract_test.go:84: catalog and status request + complete referenced component closures equal hand spec; two generations byte-identical
--- PASS: TestCatalogOpenAPIFidelity (0.23s)
=== RUN   TestAdminReadHTTPMatchesOpenAPI
=== RUN   TestAdminReadHTTPMatchesOpenAPI/admin_lists_every_course
=== RUN   TestAdminReadHTTPMatchesOpenAPI/admin_list_search_parameters
=== RUN   TestAdminReadHTTPMatchesOpenAPI/instructor_lists_own_courses
=== RUN   TestAdminReadHTTPMatchesOpenAPI/admin_reads_a_tree
=== RUN   TestAdminReadHTTPMatchesOpenAPI/owner_reads_a_tree
=== RUN   TestAdminReadHTTPMatchesOpenAPI/P07_admin_lists_owner_candidates
=== RUN   TestAdminReadHTTPMatchesOpenAPI/R01_anonymous_list
=== RUN   TestAdminReadHTTPMatchesOpenAPI/R01_anonymous_tree
=== RUN   TestAdminReadHTTPMatchesOpenAPI/R01_anonymous_instructors
=== RUN   TestAdminReadHTTPMatchesOpenAPI/R02_student_list
=== RUN   TestAdminReadHTTPMatchesOpenAPI/student_role_is_refused_before_query_validation
=== RUN   TestAdminReadHTTPMatchesOpenAPI/R02_student_tree
=== RUN   TestAdminReadHTTPMatchesOpenAPI/R02_student_instructors
=== RUN   TestAdminReadHTTPMatchesOpenAPI/R03_disabled_account
=== RUN   TestAdminReadHTTPMatchesOpenAPI/R04_foreign_instructor
=== RUN   TestAdminReadHTTPMatchesOpenAPI/R05_instructor_owner_candidates
=== RUN   TestAdminReadHTTPMatchesOpenAPI/R06_unknown_id
=== RUN   TestAdminReadHTTPMatchesOpenAPI/R07_malformed_id
=== RUN   TestAdminReadHTTPMatchesOpenAPI/invalid_list_status
=== RUN   TestAdminReadHTTPMatchesOpenAPI/unknown_list_parameter
=== RUN   TestAdminReadHTTPMatchesOpenAPI/repeated_list_parameter
--- PASS: TestAdminReadHTTPMatchesOpenAPI (0.06s)
    --- PASS: TestAdminReadHTTPMatchesOpenAPI/admin_lists_every_course (0.00s)
    --- PASS: TestAdminReadHTTPMatchesOpenAPI/admin_list_search_parameters (0.00s)
    --- PASS: TestAdminReadHTTPMatchesOpenAPI/instructor_lists_own_courses (0.00s)
    --- PASS: TestAdminReadHTTPMatchesOpenAPI/admin_reads_a_tree (0.00s)
    --- PASS: TestAdminReadHTTPMatchesOpenAPI/owner_reads_a_tree (0.00s)
    --- PASS: TestAdminReadHTTPMatchesOpenAPI/P07_admin_lists_owner_candidates (0.00s)
    --- PASS: TestAdminReadHTTPMatchesOpenAPI/R01_anonymous_list (0.00s)
    --- PASS: TestAdminReadHTTPMatchesOpenAPI/R01_anonymous_tree (0.00s)
    --- PASS: TestAdminReadHTTPMatchesOpenAPI/R01_anonymous_instructors (0.00s)
    --- PASS: TestAdminReadHTTPMatchesOpenAPI/R02_student_list (0.00s)
    --- PASS: TestAdminReadHTTPMatchesOpenAPI/student_role_is_refused_before_query_validation (0.00s)
    --- PASS: TestAdminReadHTTPMatchesOpenAPI/R02_student_tree (0.00s)
    --- PASS: TestAdminReadHTTPMatchesOpenAPI/R02_student_instructors (0.00s)
    --- PASS: TestAdminReadHTTPMatchesOpenAPI/R03_disabled_account (0.00s)
    --- PASS: TestAdminReadHTTPMatchesOpenAPI/R04_foreign_instructor (0.00s)
    --- PASS: TestAdminReadHTTPMatchesOpenAPI/R05_instructor_owner_candidates (0.00s)
    --- PASS: TestAdminReadHTTPMatchesOpenAPI/R06_unknown_id (0.00s)
    --- PASS: TestAdminReadHTTPMatchesOpenAPI/R07_malformed_id (0.00s)
    --- PASS: TestAdminReadHTTPMatchesOpenAPI/invalid_list_status (0.00s)
    --- PASS: TestAdminReadHTTPMatchesOpenAPI/unknown_list_parameter (0.00s)
    --- PASS: TestAdminReadHTTPMatchesOpenAPI/repeated_list_parameter (0.00s)
=== RUN   TestAdminCreateContract
=== RUN   TestAdminCreateContract/P02_P03_R29_admin_creates_one_invisible_draft_and_replay_is_refused
=== RUN   TestAdminCreateContract/P04_instructor_owns_a_create_that_omits_instructorId
=== RUN   TestAdminCreateContract/P05_admin_chooses_an_owner_and_the_owner_lists_the_course
--- PASS: TestAdminCreateContract (0.24s)
    --- PASS: TestAdminCreateContract/P02_P03_R29_admin_creates_one_invisible_draft_and_replay_is_refused (0.12s)
    --- PASS: TestAdminCreateContract/P04_instructor_owns_a_create_that_omits_instructorId (0.06s)
    --- PASS: TestAdminCreateContract/P05_admin_chooses_an_owner_and_the_owner_lists_the_course (0.06s)
=== RUN   TestAdminCourseListEnrichmentContracts
=== RUN   TestAdminCourseListEnrichmentContracts/Cv01_list_row_carries_lessonCount,_studentCount,_price_and_the_status_snapshot
=== RUN   TestAdminCourseListEnrichmentContracts/Cv02_a_free_course_serialises_price_as_null
=== RUN   TestAdminCourseListEnrichmentContracts/Cv03_search_is_forwarded_to_the_query
=== RUN   TestAdminCourseListEnrichmentContracts/Cv04_an_empty_search_parameter_is_rejected
=== RUN   TestAdminCourseListEnrichmentContracts/Cv06_sort_is_forwarded_and_defaults_to_recent
=== RUN   TestAdminCourseListEnrichmentContracts/Cv07_an_empty_or_unknown_sort_is_rejected_before_the_repository
--- PASS: TestAdminCourseListEnrichmentContracts (0.24s)
    --- PASS: TestAdminCourseListEnrichmentContracts/Cv01_list_row_carries_lessonCount,_studentCount,_price_and_the_status_snapshot (0.06s)
    --- PASS: TestAdminCourseListEnrichmentContracts/Cv02_a_free_course_serialises_price_as_null (0.06s)
    --- PASS: TestAdminCourseListEnrichmentContracts/Cv03_search_is_forwarded_to_the_query (0.06s)
    --- PASS: TestAdminCourseListEnrichmentContracts/Cv04_an_empty_search_parameter_is_rejected (0.00s)
    --- PASS: TestAdminCourseListEnrichmentContracts/Cv06_sort_is_forwarded_and_defaults_to_recent (0.06s)
    --- PASS: TestAdminCourseListEnrichmentContracts/Cv07_an_empty_or_unknown_sort_is_rejected_before_the_repository (0.00s)
=== RUN   TestAdminSaveHappyAndBoundaryContracts
=== RUN   TestAdminSaveHappyAndBoundaryContracts/P06_admin_reassigns_the_owner_and_both_lists_move
=== RUN   TestAdminSaveHappyAndBoundaryContracts/B04_a_never-published_draft_can_be_renamed
=== RUN   TestAdminSaveHappyAndBoundaryContracts/B05_maximal_Cyrillic_tree_stays_below_the_cap_and_saves
=== RUN   TestAdminSaveHappyAndBoundaryContracts/an_empty_tree_saves_as_the_complete_replacement
=== RUN   TestAdminSaveHappyAndBoundaryContracts/B06_invalid_tree_ceilings_leave_the_tree_unchanged
=== RUN   TestAdminSaveHappyAndBoundaryContracts/B06_invalid_tree_ceilings_leave_the_tree_unchanged/41_modules
=== RUN   TestAdminSaveHappyAndBoundaryContracts/B06_invalid_tree_ceilings_leave_the_tree_unchanged/121_lessons
=== RUN   TestAdminSaveHappyAndBoundaryContracts/B07_a_body_over_512_KiB_is_refused_with_a_bounded_response
=== RUN   TestAdminSaveHappyAndBoundaryContracts/B08_summary_accepts_500_characters_and_rejects_501
=== RUN   TestAdminSaveHappyAndBoundaryContracts/course_and_node_field_bounds_surface_as_invalid_request
=== RUN   TestAdminSaveHappyAndBoundaryContracts/course_and_node_field_bounds_surface_as_invalid_request/empty_title
=== RUN   TestAdminSaveHappyAndBoundaryContracts/course_and_node_field_bounds_surface_as_invalid_request/201_character_title
=== RUN   TestAdminSaveHappyAndBoundaryContracts/course_and_node_field_bounds_surface_as_invalid_request/20001_character_description
=== RUN   TestAdminSaveHappyAndBoundaryContracts/course_and_node_field_bounds_surface_as_invalid_request/121_character_slug
=== RUN   TestAdminSaveHappyAndBoundaryContracts/course_and_node_field_bounds_surface_as_invalid_request/201_character_module_title
=== RUN   TestAdminSaveHappyAndBoundaryContracts/course_and_node_field_bounds_surface_as_invalid_request/201_character_lesson_title
=== RUN   TestAdminSaveHappyAndBoundaryContracts/B-3_the_course-wide_ceiling_is_refused_by_name_where_the_schema_cannot_see_it
=== RUN   TestAdminSaveHappyAndBoundaryContracts/B-3_the_course-wide_ceiling_is_refused_by_name_where_the_schema_cannot_see_it/121_across_two_modules_passes_the_schema_and_is_refused_as_lesson_limit
=== RUN   TestAdminSaveHappyAndBoundaryContracts/B-3_the_course-wide_ceiling_is_refused_by_name_where_the_schema_cannot_see_it/121_in_one_module_is_stopped_by_the_per-module_maxItems
=== RUN   TestAdminSaveHappyAndBoundaryContracts/B12_read_token_round-trips_and_a_lesson-only_save_retires_it
--- PASS: TestAdminSaveHappyAndBoundaryContracts (1.12s)
    --- PASS: TestAdminSaveHappyAndBoundaryContracts/P06_admin_reassigns_the_owner_and_both_lists_move (0.06s)
    --- PASS: TestAdminSaveHappyAndBoundaryContracts/B04_a_never-published_draft_can_be_renamed (0.06s)
    --- PASS: TestAdminSaveHappyAndBoundaryContracts/B05_maximal_Cyrillic_tree_stays_below_the_cap_and_saves (0.07s)
    --- PASS: TestAdminSaveHappyAndBoundaryContracts/an_empty_tree_saves_as_the_complete_replacement (0.06s)
    --- PASS: TestAdminSaveHappyAndBoundaryContracts/B06_invalid_tree_ceilings_leave_the_tree_unchanged (0.12s)
        --- PASS: TestAdminSaveHappyAndBoundaryContracts/B06_invalid_tree_ceilings_leave_the_tree_unchanged/41_modules (0.06s)
        --- PASS: TestAdminSaveHappyAndBoundaryContracts/B06_invalid_tree_ceilings_leave_the_tree_unchanged/121_lessons (0.06s)
    --- PASS: TestAdminSaveHappyAndBoundaryContracts/B07_a_body_over_512_KiB_is_refused_with_a_bounded_response (0.06s)
    --- PASS: TestAdminSaveHappyAndBoundaryContracts/B08_summary_accepts_500_characters_and_rejects_501 (0.12s)
    --- PASS: TestAdminSaveHappyAndBoundaryContracts/course_and_node_field_bounds_surface_as_invalid_request (0.35s)
        --- PASS: TestAdminSaveHappyAndBoundaryContracts/course_and_node_field_bounds_surface_as_invalid_request/empty_title (0.06s)
        --- PASS: TestAdminSaveHappyAndBoundaryContracts/course_and_node_field_bounds_surface_as_invalid_request/201_character_title (0.06s)
        --- PASS: TestAdminSaveHappyAndBoundaryContracts/course_and_node_field_bounds_surface_as_invalid_request/20001_character_description (0.06s)
        --- PASS: TestAdminSaveHappyAndBoundaryContracts/course_and_node_field_bounds_surface_as_invalid_request/121_character_slug (0.06s)
        --- PASS: TestAdminSaveHappyAndBoundaryContracts/course_and_node_field_bounds_surface_as_invalid_request/201_character_module_title (0.06s)
        --- PASS: TestAdminSaveHappyAndBoundaryContracts/course_and_node_field_bounds_surface_as_invalid_request/201_character_lesson_title (0.06s)
    --- PASS: TestAdminSaveHappyAndBoundaryContracts/B-3_the_course-wide_ceiling_is_refused_by_name_where_the_schema_cannot_see_it (0.16s)
        --- PASS: TestAdminSaveHappyAndBoundaryContracts/B-3_the_course-wide_ceiling_is_refused_by_name_where_the_schema_cannot_see_it/121_across_two_modules_passes_the_schema_and_is_refused_as_lesson_limit (0.06s)
        --- PASS: TestAdminSaveHappyAndBoundaryContracts/B-3_the_course-wide_ceiling_is_refused_by_name_where_the_schema_cannot_see_it/121_in_one_module_is_stopped_by_the_per-module_maxItems (0.10s)
    --- PASS: TestAdminSaveHappyAndBoundaryContracts/B12_read_token_round-trips_and_a_lesson-only_save_retires_it (0.06s)
=== RUN   TestAdminSaveRefusalContracts
=== RUN   TestAdminSaveRefusalContracts/R12_instructor_cannot_reassign_ownership
=== RUN   TestAdminSaveRefusalContracts/R13_ineligible_owners_are_refused_without_changing_owner
=== RUN   TestAdminSaveRefusalContracts/R13_ineligible_owners_are_refused_without_changing_owner/student
=== RUN   TestAdminSaveRefusalContracts/R13_ineligible_owners_are_refused_without_changing_owner/disabled_instructor
=== RUN   TestAdminSaveRefusalContracts/R13_ineligible_owners_are_refused_without_changing_owner/unknown_user
=== RUN   TestAdminSaveRefusalContracts/R20_published_and_archived_slugs_are_immutable
=== RUN   TestAdminSaveRefusalContracts/R20_published_and_archived_slugs_are_immutable/published
=== RUN   TestAdminSaveRefusalContracts/R20_published_and_archived_slugs_are_immutable/archived
=== RUN   TestAdminSaveRefusalContracts/R21_rename_onto_a_taken_slug_is_refused
=== RUN   TestAdminSaveRefusalContracts/R22_lifecycle_and_position_members_are_rejected_before_repository_access
=== RUN   TestAdminSaveRefusalContracts/R22_lifecycle_and_position_members_are_rejected_before_repository_access/status
=== RUN   TestAdminSaveRefusalContracts/R22_lifecycle_and_position_members_are_rejected_before_repository_access/publishedAt
=== RUN   TestAdminSaveRefusalContracts/R22_lifecycle_and_position_members_are_rejected_before_repository_access/position
--- PASS: TestAdminSaveRefusalContracts (0.57s)
    --- PASS: TestAdminSaveRefusalContracts/R12_instructor_cannot_reassign_ownership (0.06s)
    --- PASS: TestAdminSaveRefusalContracts/R13_ineligible_owners_are_refused_without_changing_owner (0.17s)
        --- PASS: TestAdminSaveRefusalContracts/R13_ineligible_owners_are_refused_without_changing_owner/student (0.06s)
        --- PASS: TestAdminSaveRefusalContracts/R13_ineligible_owners_are_refused_without_changing_owner/disabled_instructor (0.06s)
        --- PASS: TestAdminSaveRefusalContracts/R13_ineligible_owners_are_refused_without_changing_owner/unknown_user (0.06s)
    --- PASS: TestAdminSaveRefusalContracts/R20_published_and_archived_slugs_are_immutable (0.11s)
        --- PASS: TestAdminSaveRefusalContracts/R20_published_and_archived_slugs_are_immutable/published (0.06s)
        --- PASS: TestAdminSaveRefusalContracts/R20_published_and_archived_slugs_are_immutable/archived (0.05s)
    --- PASS: TestAdminSaveRefusalContracts/R21_rename_onto_a_taken_slug_is_refused (0.06s)
    --- PASS: TestAdminSaveRefusalContracts/R22_lifecycle_and_position_members_are_rejected_before_repository_access (0.17s)
        --- PASS: TestAdminSaveRefusalContracts/R22_lifecycle_and_position_members_are_rejected_before_repository_access/status (0.06s)
        --- PASS: TestAdminSaveRefusalContracts/R22_lifecycle_and_position_members_are_rejected_before_repository_access/publishedAt (0.06s)
        --- PASS: TestAdminSaveRefusalContracts/R22_lifecycle_and_position_members_are_rejected_before_repository_access/position (0.06s)
=== RUN   TestAdminLifecycleHappyAndBoundaryContracts
=== RUN   TestAdminLifecycleHappyAndBoundaryContracts/P11_P12_P13_publish_archive_and_restore_update_public_visibility
=== RUN   TestAdminLifecycleHappyAndBoundaryContracts/P14_instructor_publishes_their_own_course
=== RUN   TestAdminLifecycleHappyAndBoundaryContracts/B01_an_empty_course_may_be_published
=== RUN   TestAdminLifecycleHappyAndBoundaryContracts/a_draft_course_can_be_deleted
--- PASS: TestAdminLifecycleHappyAndBoundaryContracts (0.36s)
    --- PASS: TestAdminLifecycleHappyAndBoundaryContracts/P11_P12_P13_publish_archive_and_restore_update_public_visibility (0.18s)
    --- PASS: TestAdminLifecycleHappyAndBoundaryContracts/P14_instructor_publishes_their_own_course (0.06s)
    --- PASS: TestAdminLifecycleHappyAndBoundaryContracts/B01_an_empty_course_may_be_published (0.06s)
    --- PASS: TestAdminLifecycleHappyAndBoundaryContracts/a_draft_course_can_be_deleted (0.06s)
=== RUN   TestAdminMutationRefusalMatrix
=== RUN   TestAdminMutationRefusalMatrix/R08_anonymous_is_refused_on_every_mutation
=== RUN   TestAdminMutationRefusalMatrix/R08_anonymous_is_refused_on_every_mutation/create
=== RUN   TestAdminMutationRefusalMatrix/R08_anonymous_is_refused_on_every_mutation/save
=== RUN   TestAdminMutationRefusalMatrix/R08_anonymous_is_refused_on_every_mutation/status
=== RUN   TestAdminMutationRefusalMatrix/R08_anonymous_is_refused_on_every_mutation/delete
=== RUN   TestAdminMutationRefusalMatrix/R09_a_student_is_refused_on_every_mutation
=== RUN   TestAdminMutationRefusalMatrix/R09_a_student_is_refused_on_every_mutation/create
=== RUN   TestAdminMutationRefusalMatrix/R09_a_student_is_refused_on_every_mutation/save
=== RUN   TestAdminMutationRefusalMatrix/R09_a_student_is_refused_on_every_mutation/status
=== RUN   TestAdminMutationRefusalMatrix/R09_a_student_is_refused_on_every_mutation/delete
=== RUN   TestAdminMutationRefusalMatrix/R10_cross-site_mutation_is_refused_before_authentication
=== RUN   TestAdminMutationRefusalMatrix/R10_cross-site_mutation_is_refused_before_authentication/create
=== RUN   TestAdminMutationRefusalMatrix/R10_cross-site_mutation_is_refused_before_authentication/save
=== RUN   TestAdminMutationRefusalMatrix/R10_cross-site_mutation_is_refused_before_authentication/status
=== RUN   TestAdminMutationRefusalMatrix/R10_cross-site_mutation_is_refused_before_authentication/delete
=== RUN   TestAdminMutationRefusalMatrix/R11_a_foreign_instructor_cannot_save_transition_or_delete
=== RUN   TestAdminMutationRefusalMatrix/R11_a_foreign_instructor_cannot_save_transition_or_delete/save
=== RUN   TestAdminMutationRefusalMatrix/R11_a_foreign_instructor_cannot_save_transition_or_delete/status
=== RUN   TestAdminMutationRefusalMatrix/R11_a_foreign_instructor_cannot_save_transition_or_delete/delete
--- PASS: TestAdminMutationRefusalMatrix (0.91s)
    --- PASS: TestAdminMutationRefusalMatrix/R08_anonymous_is_refused_on_every_mutation (0.23s)
        --- PASS: TestAdminMutationRefusalMatrix/R08_anonymous_is_refused_on_every_mutation/create (0.06s)
        --- PASS: TestAdminMutationRefusalMatrix/R08_anonymous_is_refused_on_every_mutation/save (0.06s)
        --- PASS: TestAdminMutationRefusalMatrix/R08_anonymous_is_refused_on_every_mutation/status (0.06s)
        --- PASS: TestAdminMutationRefusalMatrix/R08_anonymous_is_refused_on_every_mutation/delete (0.06s)
    --- PASS: TestAdminMutationRefusalMatrix/R09_a_student_is_refused_on_every_mutation (0.25s)
        --- PASS: TestAdminMutationRefusalMatrix/R09_a_student_is_refused_on_every_mutation/create (0.06s)
        --- PASS: TestAdminMutationRefusalMatrix/R09_a_student_is_refused_on_every_mutation/save (0.06s)
        --- PASS: TestAdminMutationRefusalMatrix/R09_a_student_is_refused_on_every_mutation/status (0.06s)
        --- PASS: TestAdminMutationRefusalMatrix/R09_a_student_is_refused_on_every_mutation/delete (0.06s)
    --- PASS: TestAdminMutationRefusalMatrix/R10_cross-site_mutation_is_refused_before_authentication (0.25s)
        --- PASS: TestAdminMutationRefusalMatrix/R10_cross-site_mutation_is_refused_before_authentication/create (0.06s)
        --- PASS: TestAdminMutationRefusalMatrix/R10_cross-site_mutation_is_refused_before_authentication/save (0.08s)
        --- PASS: TestAdminMutationRefusalMatrix/R10_cross-site_mutation_is_refused_before_authentication/status (0.06s)
        --- PASS: TestAdminMutationRefusalMatrix/R10_cross-site_mutation_is_refused_before_authentication/delete (0.06s)
    --- PASS: TestAdminMutationRefusalMatrix/R11_a_foreign_instructor_cannot_save_transition_or_delete (0.18s)
        --- PASS: TestAdminMutationRefusalMatrix/R11_a_foreign_instructor_cannot_save_transition_or_delete/save (0.06s)
        --- PASS: TestAdminMutationRefusalMatrix/R11_a_foreign_instructor_cannot_save_transition_or_delete/status (0.06s)
        --- PASS: TestAdminMutationRefusalMatrix/R11_a_foreign_instructor_cannot_save_transition_or_delete/delete (0.06s)
=== RUN   TestAdminLifecycleRefusalContracts
=== RUN   TestAdminLifecycleRefusalContracts/R14_an_illegal_transition_is_a_named_conflict_and_leaves_status_unchanged
=== RUN   TestAdminLifecycleRefusalContracts/R15_published_and_archived_courses_cannot_be_deleted
=== RUN   TestAdminLifecycleRefusalContracts/R15_published_and_archived_courses_cannot_be_deleted/published
=== RUN   TestAdminLifecycleRefusalContracts/R15_published_and_archived_courses_cannot_be_deleted/archived
=== RUN   TestAdminLifecycleRefusalContracts/R23_a_transition_retires_the_previous_tree_token
--- PASS: TestAdminLifecycleRefusalContracts (0.29s)
    --- PASS: TestAdminLifecycleRefusalContracts/R14_an_illegal_transition_is_a_named_conflict_and_leaves_status_unchanged (0.06s)
    --- PASS: TestAdminLifecycleRefusalContracts/R15_published_and_archived_courses_cannot_be_deleted (0.12s)
        --- PASS: TestAdminLifecycleRefusalContracts/R15_published_and_archived_courses_cannot_be_deleted/published (0.06s)
        --- PASS: TestAdminLifecycleRefusalContracts/R15_published_and_archived_courses_cannot_be_deleted/archived (0.06s)
    --- PASS: TestAdminLifecycleRefusalContracts/R23_a_transition_retires_the_previous_tree_token (0.12s)
=== RUN   TestAdminWriteStrictJSON
=== RUN   TestAdminWriteStrictJSON/malformed
=== RUN   TestAdminWriteStrictJSON/duplicate
=== RUN   TestAdminWriteStrictJSON/unknown
=== RUN   TestAdminWriteStrictJSON/trailing
=== RUN   TestAdminWriteStrictJSON/expectedUpdatedAt_requires_six_fractional_digits
=== RUN   TestAdminWriteStrictJSON/query_parameters_are_rejected
=== RUN   TestAdminWriteStrictJSON/non-JSON_content_type_is_rejected
--- PASS: TestAdminWriteStrictJSON (0.36s)
    --- PASS: TestAdminWriteStrictJSON/malformed (0.06s)
    --- PASS: TestAdminWriteStrictJSON/duplicate (0.06s)
    --- PASS: TestAdminWriteStrictJSON/unknown (0.06s)
    --- PASS: TestAdminWriteStrictJSON/trailing (0.06s)
    --- PASS: TestAdminWriteStrictJSON/expectedUpdatedAt_requires_six_fractional_digits (0.06s)
    --- PASS: TestAdminWriteStrictJSON/query_parameters_are_rejected (0.06s)
    --- PASS: TestAdminWriteStrictJSON/non-JSON_content_type_is_rejected (0.00s)
=== RUN   TestAdminReadWrongMethod
--- PASS: TestAdminReadWrongMethod (0.00s)
=== RUN   TestContractCeilingsEqualTheDomainConstants
=== PAUSE TestContractCeilingsEqualTheDomainConstants
=== RUN   TestCatalogHTTPMatchesOpenAPI
=== RUN   TestCatalogHTTPMatchesOpenAPI/default_list
=== RUN   TestCatalogHTTPMatchesOpenAPI/boundaries
=== RUN   TestCatalogHTTPMatchesOpenAPI/empty_list
=== RUN   TestCatalogHTTPMatchesOpenAPI/detail
=== RUN   TestCatalogHTTPMatchesOpenAPI/missing_detail
=== RUN   TestCatalogHTTPMatchesOpenAPI/draft_hidden
=== RUN   TestCatalogHTTPMatchesOpenAPI/archived_hidden
=== RUN   TestCatalogHTTPMatchesOpenAPI/bad_limit
=== RUN   TestCatalogHTTPMatchesOpenAPI/repeated_parameter
=== RUN   TestCatalogHTTPMatchesOpenAPI/unknown_parameter
=== RUN   TestCatalogHTTPMatchesOpenAPI/decimal
=== RUN   TestCatalogHTTPMatchesOpenAPI/negative
=== RUN   TestCatalogHTTPMatchesOpenAPI/empty_numeric
=== RUN   TestCatalogHTTPMatchesOpenAPI/huge_numeric
=== RUN   TestCatalogHTTPMatchesOpenAPI/plus_prefix
=== RUN   TestCatalogHTTPMatchesOpenAPI/max_limit
=== RUN   TestCatalogHTTPMatchesOpenAPI/max_offset
=== RUN   TestCatalogHTTPMatchesOpenAPI/long_Unicode
=== RUN   TestCatalogHTTPMatchesOpenAPI/unknown_sort
=== RUN   TestCatalogHTTPMatchesOpenAPI/empty_sort
=== RUN   TestCatalogHTTPMatchesOpenAPI/bad_encoding
=== RUN   TestCatalogHTTPMatchesOpenAPI/NUL_input
=== RUN   TestCatalogHTTPMatchesOpenAPI/invalid_slug
=== RUN   TestCatalogHTTPMatchesOpenAPI/detail_query_rejected
=== RUN   TestCatalogHTTPMatchesOpenAPI/dependency
=== RUN   TestCatalogHTTPMatchesOpenAPI/unexpected
--- PASS: TestCatalogHTTPMatchesOpenAPI (0.06s)
    --- PASS: TestCatalogHTTPMatchesOpenAPI/default_list (0.00s)
    --- PASS: TestCatalogHTTPMatchesOpenAPI/boundaries (0.00s)
    --- PASS: TestCatalogHTTPMatchesOpenAPI/empty_list (0.00s)
    --- PASS: TestCatalogHTTPMatchesOpenAPI/detail (0.00s)
    --- PASS: TestCatalogHTTPMatchesOpenAPI/missing_detail (0.00s)
    --- PASS: TestCatalogHTTPMatchesOpenAPI/draft_hidden (0.00s)
    --- PASS: TestCatalogHTTPMatchesOpenAPI/archived_hidden (0.00s)
    --- PASS: TestCatalogHTTPMatchesOpenAPI/bad_limit (0.00s)
    --- PASS: TestCatalogHTTPMatchesOpenAPI/repeated_parameter (0.00s)
    --- PASS: TestCatalogHTTPMatchesOpenAPI/unknown_parameter (0.00s)
    --- PASS: TestCatalogHTTPMatchesOpenAPI/decimal (0.00s)
    --- PASS: TestCatalogHTTPMatchesOpenAPI/negative (0.00s)
    --- PASS: TestCatalogHTTPMatchesOpenAPI/empty_numeric (0.00s)
    --- PASS: TestCatalogHTTPMatchesOpenAPI/huge_numeric (0.00s)
    --- PASS: TestCatalogHTTPMatchesOpenAPI/plus_prefix (0.00s)
    --- PASS: TestCatalogHTTPMatchesOpenAPI/max_limit (0.00s)
    --- PASS: TestCatalogHTTPMatchesOpenAPI/max_offset (0.00s)
    --- PASS: TestCatalogHTTPMatchesOpenAPI/long_Unicode (0.00s)
    --- PASS: TestCatalogHTTPMatchesOpenAPI/unknown_sort (0.00s)
    --- PASS: TestCatalogHTTPMatchesOpenAPI/empty_sort (0.00s)
    --- PASS: TestCatalogHTTPMatchesOpenAPI/bad_encoding (0.00s)
    --- PASS: TestCatalogHTTPMatchesOpenAPI/NUL_input (0.00s)
    --- PASS: TestCatalogHTTPMatchesOpenAPI/invalid_slug (0.00s)
    --- PASS: TestCatalogHTTPMatchesOpenAPI/detail_query_rejected (0.00s)
    --- PASS: TestCatalogHTTPMatchesOpenAPI/dependency (0.00s)
    --- PASS: TestCatalogHTTPMatchesOpenAPI/unexpected (0.00s)
=== RUN   TestListPreservesDefaultsAndRequestedPagination
--- PASS: TestListPreservesDefaultsAndRequestedPagination (0.00s)
=== RUN   TestPublishedCourseDetailExposesOnlyOrderedProgramTitles
--- PASS: TestPublishedCourseDetailExposesOnlyOrderedProgramTitles (0.00s)
=== RUN   TestUnpublishedCourseProgramsAreNotExposed
--- PASS: TestUnpublishedCourseProgramsAreNotExposed (0.00s)
=== RUN   TestLearningHTTPMatchesOpenAPIAndConcreteProgressContract
--- PASS: TestLearningHTTPMatchesOpenAPIAndConcreteProgressContract (0.06s)
=== RUN   TestLearningHTTPBoundariesAndDenials
=== RUN   TestLearningHTTPBoundariesAndDenials/authentication_required
=== RUN   TestLearningHTTPBoundariesAndDenials/invalid_slug_before_repository
=== RUN   TestLearningHTTPBoundariesAndDenials/course_not_found
=== RUN   TestLearningHTTPBoundariesAndDenials/R17_unavailable_course
=== RUN   TestLearningHTTPBoundariesAndDenials/dependency_unavailable
=== RUN   TestLearningHTTPBoundariesAndDenials/query_rejected
--- PASS: TestLearningHTTPBoundariesAndDenials (0.28s)
    --- PASS: TestLearningHTTPBoundariesAndDenials/authentication_required (0.00s)
    --- PASS: TestLearningHTTPBoundariesAndDenials/invalid_slug_before_repository (0.06s)
    --- PASS: TestLearningHTTPBoundariesAndDenials/course_not_found (0.06s)
    --- PASS: TestLearningHTTPBoundariesAndDenials/R17_unavailable_course (0.06s)
    --- PASS: TestLearningHTTPBoundariesAndDenials/dependency_unavailable (0.06s)
    --- PASS: TestLearningHTTPBoundariesAndDenials/query_rejected (0.06s)
=== RUN   TestPay19LearningOutlineLocksPaidLessonsInsteadOfRefusing
--- PASS: TestPay19LearningOutlineLocksPaidLessonsInsteadOfRefusing (0.06s)
=== RUN   TestR09LearningForeignDraftDeniedAndPublishedAllowed
=== RUN   TestR09LearningForeignDraftDeniedAndPublishedAllowed/stranger_draft
=== RUN   TestR09LearningForeignDraftDeniedAndPublishedAllowed/stranger_published
--- PASS: TestR09LearningForeignDraftDeniedAndPublishedAllowed (0.12s)
    --- PASS: TestR09LearningForeignDraftDeniedAndPublishedAllowed/stranger_draft (0.06s)
    --- PASS: TestR09LearningForeignDraftDeniedAndPublishedAllowed/stranger_published (0.06s)
=== RUN   TestB01LearningCourseWithoutLessonsUsesEmptyArraysAndNullContinue
--- PASS: TestB01LearningCourseWithoutLessonsUsesEmptyArraysAndNullContinue (0.00s)
=== RUN   TestR21AnonymousPublicCourseProgramStillExposesNoLessonIDs
--- PASS: TestR21AnonymousPublicCourseProgramStillExposesNoLessonIDs (0.00s)
=== RUN   TestAdminStatusRequestMatchesGeneratedOpenAPI
=== RUN   TestAdminStatusRequestMatchesGeneratedOpenAPI/accepted_publish
=== RUN   TestAdminStatusRequestMatchesGeneratedOpenAPI/null_status
=== RUN   TestAdminStatusRequestMatchesGeneratedOpenAPI/null_expectedUpdatedAt
=== RUN   TestAdminStatusRequestMatchesGeneratedOpenAPI/missing_status
=== RUN   TestAdminStatusRequestMatchesGeneratedOpenAPI/missing_expectedUpdatedAt
--- PASS: TestAdminStatusRequestMatchesGeneratedOpenAPI (0.09s)
    --- PASS: TestAdminStatusRequestMatchesGeneratedOpenAPI/accepted_publish (0.00s)
    --- PASS: TestAdminStatusRequestMatchesGeneratedOpenAPI/null_status (0.00s)
    --- PASS: TestAdminStatusRequestMatchesGeneratedOpenAPI/null_expectedUpdatedAt (0.00s)
    --- PASS: TestAdminStatusRequestMatchesGeneratedOpenAPI/missing_status (0.00s)
    --- PASS: TestAdminStatusRequestMatchesGeneratedOpenAPI/missing_expectedUpdatedAt (0.00s)
=== RUN   TestCatalogHandlersMatchGeneratedOpenAPI
=== RUN   TestCatalogHandlersMatchGeneratedOpenAPI//api/v1/courses
=== RUN   TestCatalogHandlersMatchGeneratedOpenAPI//api/v1/courses?q=empty
=== RUN   TestCatalogHandlersMatchGeneratedOpenAPI//api/v1/courses/go-foundations
=== RUN   TestCatalogHandlersMatchGeneratedOpenAPI//api/v1/courses/missing
=== RUN   TestCatalogHandlersMatchGeneratedOpenAPI//api/v1/courses?limit=0
--- PASS: TestCatalogHandlersMatchGeneratedOpenAPI (0.08s)
    --- PASS: TestCatalogHandlersMatchGeneratedOpenAPI//api/v1/courses (0.00s)
    --- PASS: TestCatalogHandlersMatchGeneratedOpenAPI//api/v1/courses?q=empty (0.00s)
    --- PASS: TestCatalogHandlersMatchGeneratedOpenAPI//api/v1/courses/go-foundations (0.00s)
    --- PASS: TestCatalogHandlersMatchGeneratedOpenAPI//api/v1/courses/missing (0.00s)
    --- PASS: TestCatalogHandlersMatchGeneratedOpenAPI//api/v1/courses?limit=0 (0.00s)
=== CONT  TestContractCeilingsEqualTheDomainConstants
=== RUN   TestContractCeilingsEqualTheDomainConstants/AdminModule.lessons
=== PAUSE TestContractCeilingsEqualTheDomainConstants/AdminModule.lessons
=== RUN   TestContractCeilingsEqualTheDomainConstants/LearningModule.lessons
=== PAUSE TestContractCeilingsEqualTheDomainConstants/LearningModule.lessons
=== RUN   TestContractCeilingsEqualTheDomainConstants/ProgramModule.lessons
=== PAUSE TestContractCeilingsEqualTheDomainConstants/ProgramModule.lessons
=== RUN   TestContractCeilingsEqualTheDomainConstants/AdminCourseTree.modules
=== PAUSE TestContractCeilingsEqualTheDomainConstants/AdminCourseTree.modules
=== RUN   TestContractCeilingsEqualTheDomainConstants/AdminCourseSaveRequest.modules
=== PAUSE TestContractCeilingsEqualTheDomainConstants/AdminCourseSaveRequest.modules
=== RUN   TestContractCeilingsEqualTheDomainConstants/LearningCourse.modules
=== PAUSE TestContractCeilingsEqualTheDomainConstants/LearningCourse.modules
=== RUN   TestContractCeilingsEqualTheDomainConstants/CourseDetail.program
=== PAUSE TestContractCeilingsEqualTheDomainConstants/CourseDetail.program
=== RUN   TestContractCeilingsEqualTheDomainConstants/LessonTranscript.segments
=== PAUSE TestContractCeilingsEqualTheDomainConstants/LessonTranscript.segments
=== RUN   TestContractCeilingsEqualTheDomainConstants/TranscriptSegment.documentIds
=== PAUSE TestContractCeilingsEqualTheDomainConstants/TranscriptSegment.documentIds
=== RUN   TestContractCeilingsEqualTheDomainConstants/LessonTranscript.documents
=== PAUSE TestContractCeilingsEqualTheDomainConstants/LessonTranscript.documents
=== RUN   TestContractCeilingsEqualTheDomainConstants/TranscriptSegment.text
=== PAUSE TestContractCeilingsEqualTheDomainConstants/TranscriptSegment.text
=== RUN   TestContractCeilingsEqualTheDomainConstants/the_course-wide_lesson_total_is_owned_by_the_domain_alone
=== PAUSE TestContractCeilingsEqualTheDomainConstants/the_course-wide_lesson_total_is_owned_by_the_domain_alone
=== CONT  TestContractCeilingsEqualTheDomainConstants/AdminModule.lessons
=== CONT  TestContractCeilingsEqualTheDomainConstants/CourseDetail.program
=== CONT  TestContractCeilingsEqualTheDomainConstants/AdminCourseTree.modules
=== CONT  TestContractCeilingsEqualTheDomainConstants/LessonTranscript.segments
=== CONT  TestContractCeilingsEqualTheDomainConstants/LessonTranscript.documents
=== CONT  TestContractCeilingsEqualTheDomainConstants/TranscriptSegment.documentIds
=== CONT  TestContractCeilingsEqualTheDomainConstants/the_course-wide_lesson_total_is_owned_by_the_domain_alone
=== CONT  TestContractCeilingsEqualTheDomainConstants/TranscriptSegment.text
=== CONT  TestContractCeilingsEqualTheDomainConstants/ProgramModule.lessons
=== CONT  TestContractCeilingsEqualTheDomainConstants/LearningCourse.modules
=== CONT  TestContractCeilingsEqualTheDomainConstants/LearningModule.lessons
=== CONT  TestContractCeilingsEqualTheDomainConstants/AdminCourseSaveRequest.modules
--- PASS: TestContractCeilingsEqualTheDomainConstants (0.03s)
    --- PASS: TestContractCeilingsEqualTheDomainConstants/AdminModule.lessons (0.00s)
    --- PASS: TestContractCeilingsEqualTheDomainConstants/AdminCourseTree.modules (0.00s)
    --- PASS: TestContractCeilingsEqualTheDomainConstants/LessonTranscript.segments (0.00s)
    --- PASS: TestContractCeilingsEqualTheDomainConstants/CourseDetail.program (0.00s)
    --- PASS: TestContractCeilingsEqualTheDomainConstants/LessonTranscript.documents (0.00s)
    --- PASS: TestContractCeilingsEqualTheDomainConstants/TranscriptSegment.documentIds (0.00s)
    --- PASS: TestContractCeilingsEqualTheDomainConstants/ProgramModule.lessons (0.00s)
    --- PASS: TestContractCeilingsEqualTheDomainConstants/LearningCourse.modules (0.00s)
    --- PASS: TestContractCeilingsEqualTheDomainConstants/LearningModule.lessons (0.00s)
    --- PASS: TestContractCeilingsEqualTheDomainConstants/AdminCourseSaveRequest.modules (0.00s)
    --- PASS: TestContractCeilingsEqualTheDomainConstants/the_course-wide_lesson_total_is_owned_by_the_domain_alone (0.00s)
    --- PASS: TestContractCeilingsEqualTheDomainConstants/TranscriptSegment.text (0.00s)
PASS
ok  	courses/internal/course/adapters/http	7.064s
```
EXIT=0

ФАКТ: FINAL ordered gate PASS: gofmt -l (stdout порожній), go vet affected, go build ./..., go test -count=1 -v affected — кожен EXIT=0; точні команди й повний stdout вище. Final source snapshot після Error-method guard. Hand scope byte-check PASS: поза ErrorResponse.enum файл byte-identical HEAD. git diff --check EXIT=0.
ФАКТ: Coverage = 166/166 WriteError calls; domain universe 116; emitted enum 109; transport allowlist 10, unused 0. Hand +9/-0; internal_error і dependency_unavailable збережено. Fidelity ref-closure mutation detects missing enum member; дві генерації byte-identical і committed artifact перевіряється тестом.
ФАКТ: make e2e НЕ запускав за дорученням власника; cross-module/e2e перевірка належить Courses 13. Runtime handlers/domain files не змінено; sort/status/BlockType поза scope. Незалежний per-slice review Courses 2 pending, це author-check PASS.
ФАКТ: Acceptance plan snapshot SHA256 df4345dcc46b4f8d26634696d4bd544959b865cc4ebd3b3d22c74d9840275319; writer-owned план/WIP/tasks не включаю у коміт.
ДАЛІ: Атомарний commit на s0.0-spike (base d12b2c7): Go parser/tests + hand enum + overlay removal + generated + цей append-only log. Після commit передати owner Courses 13 для перевірки і Courses 2 review; без merge/push.
