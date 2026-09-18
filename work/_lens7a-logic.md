# Лінза ЛОГІЧНІ ДІРИ — Phase 7a (transcript)

ДАЛІ: прочитати docs/transcript/plan.md і docs/transcript/spec.md цілком, зафіксувати §3 (рішення), §7 (таски) і приймальні рядки з рівнями доказу. Потім перевірити кожну координату (файл:рядок), на яку план посилається.

ФАКТ (координати плану перевірені в джерелі):
- `LessonPlayer.tsx:59` = `const resumeFrom = lesson.completed ? 0 : lesson.position;` — збіг точний.
- `LessonPlayer.tsx:421` = коментар `oxlint-disable-next-line jsx-a11y/media-has-caption` — збіг точний.
- `learn.$.tsx:23-27` = `throw redirect({to,params,replace:true})` без `search`; `ssr:false` на `:13` — збіг.
- `docs/foundation/assumptions.md:21` = рядок «Catalog ordering … offset pagination може зміщуватися між окремими запитами під час публікації» — збіг.
- `docs/foundation/plan.md:49` = H04 (URL стан) — збіг.

ДАЛІ: contract_ceilings_test.go (форма піна), openapi.yaml масиви/maxItems, чи є runtime-валідація запиту за схемою (це вирішує долю B05), медіа-публікація (чи є транзакційний шов для D-3), і чи існують фікстури рівнів `lint` та `unit Go (AST)`.

ФАКТ: `kin-openapi` у репо існує ЛИШЕ в _test.go (8 файлів). У проді жодної валідації запиту за схемою немає (`grep -rln openapi backend --include=*.go | grep -v _test` → лише `course/domain/authoring.go`). Контракт-тести мають прапорець `validRequest`, тому «невалідне тіло» просто не валідується схемою (admin_handler_test.go:357-362, 1411-1416).
ФАКТ: прецедент репо — стеля живе в домені І перевіряється доменом: `authoring.go:34` + `authoring.go:198` (`validOptionalText(lesson.Description, MaximumLessonDescription)`), хоча в схемі стоїть `maxLength: 1000` (openapi.yaml:903).
РІШЕННЯ: D-2 «довжина text — робота схеми, відмова invalid_request» = крок, що не випливає: у рантаймі схема нічого не відсікає.
ФАКТ: поточний пін `contract_ceilings_test.go` перелічує 7 масивів + 1 maxLength. Масивні властивості у схемах: 742,760,797,817,834(CourseList.items=50),866(AdminUserList.items=50),892(AdminCourseList.items=50),912,931,952,1014(MediaVideoList.items — БЕЗ maxItems). Тобто «сувора форма» D-2 червона на CourseList.items/AdminCourseList.items, бо доменної константи-власника для 50 у `course/domain` не існує (`grep "Maximum.*=" → лише Modules/Lessons/LessonDescription`).

ДАЛІ: шов транзакції публікації медіа (D-3), задача, що реалізує stale_at, фікстури рівнів `lint` та `unit Go (AST)`, e2e-фікстури адмінки.

ФАКТ: план змінився на диску під час читання (D-3 переписано на «застарілість виводиться»). Перечитую поточну редакцію цілком; висновок про «нема задачі на stale_at у транзакції» ревізується.

ФАКТ (перевірено в коді, нова редакція D-3):
- селектор `course/adapters/postgres/learning.go:44` = `ORDER BY CASE WHEN assets.state='ready' THEN 0 ELSE 1 END, assets.generation DESC, assets.id` — цитата точна; крос-модульне читання `lesson_progress` на :47-48 — точне; `publish.go:24` Begin — точне; `policy(locked)` на :62 — точне; `readyCount > 2` на :85 (план каже :84); `FOR UPDATE` на 42-46 (план каже 41-45).
- `AGENTS.md:16` = «Application: domain and ports only. Ports may reference domain.» — точне.
- `work/ledger.md:47` містить «completion inside the queueing transaction»; `work/report-m11.md:16` — «commit-then-enqueue». Обидві цитати точні.

ФАКТ (предикат): `playback` НЕ викликає course-предикат — має власний `VideoCourse.CanBeViewedBy` (`playback/domain/authorization.go:33`) і власний порт `VideoAccessReader` (`playback/ports/authorization.go`). Друга копія — `course/domain/learning.go:80`. ADR 010 каже «one function with one implementation… rows R18 and R19 pin that they answer alike», але `docs/playback/plan.md:660-661` R18/R19 — про 404/400 на id, не про рівність предикатів. Піна рівності в репо немає.
ФАКТ: `50` у `CourseList.items`/`AdminCourseList.items` не має доменної константи — літерал у `course/domain/query.go:30`.
ФАКТ: `A2.3` у P05 і §2 — неіснуючий критерій; у спеці §3 про click-to-seek критерію немає взагалі (є сценарій §2.3, spec.md:68-71).
ФАКТ: критерії без жодного рядка §6: A1, A5, A19, A19a, A21, A21d, A21e-2, A21f, A24, A25, A26, A27, A32.
ФАКТ: spec.md не мінялась після `9ab4e4b`; spec.md:12 і :322 кажуть, що Q6 ЧЕКАЄ підтвердження — plan.md:9 каже «підтверджена словом».
ФАКТ: `- A…` рядків у спеці 47, шапка плану каже «35 критеріїв».
ФАКТ (фікстури, які Є): `lint` — oxlint з плагіном jsx-a11y і `--deny-warnings` (frontend/oxlint.config.ts, package.json:12); `unit Go (AST)` — `backend/platform/architecture/safe_request_logs_test.go`; e2e-перехоплення мережі — `accessibility.spec.ts:282`, `playback.spec.ts:986`; e2e-завантаження файлу — `media.spec.ts:126`; contract — kin-openapi + `validRequest`-прапорець у *_test.go; `docs/media/report.md:148` — урок про NULL CHECK; `00006_playback.sql` має `SET LOCAL lock_timeout='5s'` і test-only Down.

РІШЕННЯ: дев'ять blocking + шість non-blocking, усі з перевіреними координатами. Деталі — у фінальній відповіді.
