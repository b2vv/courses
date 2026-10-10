# Лінза EDGE CASES — 7a (transcript)

ДАЛІ: читаю docs/transcript/plan.md і spec.md цілком, будую перелік меж, звіряю з §6 (приймальні рядки) і рівнями доказу.

ФАКТ: план (298 рядків) і спека (388) прочитані повністю. §6 = P01–P10, B01–B22, R01–R11.
ДАЛІ: звіряю координати коду, на які план спирається: LessonPlayer.tsx:59/:421, learn.$.tsx:13/23-27,
contract_ceilings_test.go, публікація ассета в media (транзакція), наявність валідації запиту за OpenAPI
(важливо для B05 5001), видалення курсу/уроку/ассета.

ФАКТ (координати звірені):
- `LessonPlayer.tsx:59` — `const resumeFrom = lesson.completed ? 0 : lesson.position;` ✅ як цитує D-5.
- `LessonPlayer.tsx:421` — придушення `jsx-a11y/media-has-caption` ✅.
- `learn.$.tsx:13` `ssr:false` ✅; `:23-27` `throw redirect({to,params,replace})` **без `search`** ✅.
- `contract_ceilings_test.go` існує, пінить 7 рядків (lessons/modules/program) + description ✅.
- `openapi.yaml` maxItems: 200 (:507 інлайн-схема instructors), 50 (:835/:867/:893 сторінки), 120/40 ✅.
  Доменні константи є лише на Modules/Lessons/LessonDescription (+ media OriginalNameLength) — тобто
  масиви 200 і 50 **не мають доменного власника**.
- kin-openapi присутній ЛИШЕ в тестах: рантайм-валідатора запиту за OpenAPI немає.
  Прецедент `admin_handler_test.go:673` — «121 in one module» схемно невалідний, але сервер усе одно
  відмовляє доменним `lesson_limit` (`assertAdminRequestSchema(t, request, false)`).
- `admin_handler.go:19` `maximumAdminWriteBody = 512*1024`; playback/media — 8 KiB.

РІШЕННЯ: два сильні кандидати вже є — (1) стеля 5000 фізично не проходить у 512 KiB тіло, а B04 стоїть
на рівні integration, тобто не проходить продакшн-шлях; (2) `maxLength`/`maxItems` у схемі нічого не
відсікає в рантаймі, а D-2 віддає довжину `text` «схемі».
ДАЛІ: публікація ассета в media (чи є транзакція, заміна/видалення), видалення уроку через PUT дерева,
витіснення володіння в playback.

ФАКТ (Phase 6 код, ключове для §6):
- `progress.go:68-70`: `if saved.Present && positionSeconds <= saved.PositionSeconds && (!eventTriggered() || saved.CompletedAt != nil) { return saved,false }` —
  тобто **подієвий** апдейт (seek/pause/nav/hide/ended) у НЕзавершеному уроці **відкочує** збережену позицію назад.
- `LessonPlayer.tsx:404-410` `seekTo` → `queueHeartbeat('seek')` з новою позицією; `heartbeatNow` шле `video.currentTime`.
- `seekTo` кламп: `Math.max(0, Math.min(Math.max(1, duration), next))` — при duration=0 усе зводиться до ≤1 с.
- `media/adapters/postgres/publish.go`: власна `pool.Begin`, `FOR UPDATE` на всіх ассетах уроку, evict лише при readyCount>2.
- `course/adapters/postgres/authoring.go:36-42` `deleteMissingAuthoringLessons` — урок, прибраний із дерева, **фізично видаляється**.
- FK-преcедент: `00006_playback.sql:8` `REFERENCES course_lessons(id) ON DELETE CASCADE`, `00005_media.sql:6` `ON DELETE SET NULL`.
- `locale.ts` — UI-мова `uk|en` з cookie/Accept-Language, дефолт `uk`; у `openapi.yaml` слова `language` нема взагалі.
ДАЛІ: план змінився на диску (редакція 2, D-3 переписано) — перечитую цілком перед формулюванням знахідок.

ФАКТ: план на диску — редакція 2 (D-3 переписано на виведену застарілість через
`authored_against_generation` + селектор `learning.go:44`; B06/B07 переформульовані; X2 дістав
перевірку «читання не приймає limit/offset»). Перечитано цілком; знахідки нижче — проти редакції 2.
ФАКТ: селектор `learning.go:40-45` **не фільтрує стан** — за відсутності ready повертає найновіший
created/uploading/processing ассет. Це і є вхід для знахідки №5.
ФАКТ: `course` має оптимістичний замок (`ErrCourseModified`, `updateAuthoringCourse ... AND updated_at=$2`)
— прецедент для знахідки №8, у транскрипті нічого подібного не названо.

РІШЕННЯ: 9 blocking + 6 non-blocking, усі з перевіреними координатами (див. ФАКТ-блоки вище).
Повний текст знахідок — у фінальній відповіді лінзи.
