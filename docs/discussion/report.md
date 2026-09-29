# Phase 7b — гілки бесід (discussion): звіт

| | |
|---|---|
| Фаза | 7b — гілки бесід (коментарі під уроком) |
| Дата закриття | 2026-09-29 |
| Автор | сесія Courses 1 (план, гейти, коміти; код — Codex); рецензент — Courses 2 |
| Підстава | `docs/discussion/spec.md`, `docs/discussion/plan.md` (GATE 1 і GATE 2 пройдено 2026-09-28, Courses 2); `work/PRODUCT-BRIEF.md` §24 п.17 |
| Роль фази | Коментарі-гілки під уроком: корінь із міткою часу, пласкі відповіді (cap 200), модерація автором/власником/адміном. Доступ — четвертим мапінгом на той самий контракт `access.CanView` (ADR 013). Не входить: пагінація відповідей, per-user ліміт, аудит `removed_by`, огляд коментарів по курсу (все — робоче місце інструктора). |

## 1. Що зроблено (D01–D06)

- **D01** `2b8c72e` — домен `discussion`: типи, правила тіла/мітки/відповіді, `CanBeViewedBy` → `access.CanView`, `CanBeEditedBy`, `CanBeRemovedBy`. У тому ж слайсі `expectedMethods` 3→4 і гілка `viaDiscussion` у крос-модульних тестах доступу та межі мітки.
- **D02** `b2790c9` — міграція `00009_discussion.sql` + postgres-адаптер (ручний SQL): факти курсу/ассета/автора, список курсором, атомарний cap відповідей, зміна, прибирання; гейт-тест селектора «поточне відео»; борг селектора записано в беклог (B-22).
- **D03** `bbeffc3` — application + HTTP: чотири операції, коди помилок, `sameOrigin`, `no-store`, можливості глядача; OpenAPI + Orval; підключення в `cmd/api/main.go`. `assertContract` на кожній відповіді (33 контракт-підтести).
- **D04** `58d4a11` — фронтенд: порт, адаптер на Orval з явним мапінгом, Query-ключі уроку, панель «Обговорення» поряд із транскриптом, мітка → `onSeek`, «поточний момент», стани, a11y; `useInfiniteQuery`.
- **D05** `02d221d` — e2e K19–K22: публікація з міткою й перемотка, відповідь, модерація автором/власником/адміном, відмова на чужій чернетці, a11y; мітка стала видимою текстом ще до Play (виправлений дефект продукту).
- **D06** (цей коміт) — упорядкований `make check` на HEAD, цей звіт K01–K23 по одному, `work/tasks.md`, оновлення AGENDA/брифінгу.

## 2. K01–K23 — по одному рядку, кожен своїм доказом

Зелений `make check` — не доказ рядка; нижче кожен рядок вказує на власний тест чи артефакт.
Рівні доказів (unit / integration / contract / gate / review) — ті, що зафіксовані у спеці §6.

| Рядок | Доказ (файл:рядок, тест) | Рівень | Стан |
|---|---|---|---|
| **K01** — тіло: trim, 1..2000, порожнє/задовге → `comment_body_invalid` | `backend/internal/discussion/domain/comment_test.go:12` `TestNewCommentValidatesBodyByUnicodeRunesAndTrims` | unit | ✔ D01 |
| **K02** — мітка лише в кореня; без `ready`-ассета або поза `0..тривалість` → `comment_position_invalid` | `.../domain/comment_test.go:37` `TestNewCommentValidatesPositionInMilliseconds` | unit | ✔ D01 |
| **K03** — відповідь лише до наявного не-прибраного кореня; `comment_parent_invalid`; 201-ша → `comment_thread_full` | `.../domain/comment_test.go:68` `TestNewCommentValidatesFlatReplyParentAndCapacity` | unit | ✔ D01 |
| **K04** — `Course.CanBeViewedBy` = `access.CanView` на всій `accesstest` | `.../domain/comment_test.go:100` `TestCanBeViewedByMatchesAccessTruthTable` | unit | ✔ D01 |
| **K05** — AST: рівно **4** методи `CanBeViewedBy`, кожен один `return access.CanView(…)` | `backend/platform/architecture/access_mapping_ast_test.go:14` `TestEveryCanBeViewedByDelegatesToAccessCanView` (`expectedMethods=4`) | unit (arch) | ✔ D01 |
| **K05a** — `TestTheAccessPredicateAgreesAcrossModules` має гілку `viaDiscussion`: чотири мапінги згодні на всій таблиці | `.../architecture/shared_rules_test.go:53` (гілка `viaDiscussion`), порівняння всіх чотирьох на `:58` | unit (arch) | ✔ D01 |
| **K06** — прибрати: автор ∨ адмін ∨ власник; змінити — лише автор; сторонній — ні | `.../domain/comment_test.go:117` `TestCommentMutationPermissions` | unit | ✔ D01 |
| **K07** — `accesstest` через SQL адаптера `discussion` (факти курсу за `lessonId`) | `backend/internal/discussion/adapters/postgres/repository_integration_test.go:75` `t.Run "K07 SQL facts cover access truth table"` | integration | ✔ D02 |
| **K08** — міграція `00009`: FK, `ON DELETE CASCADE`, обмеження колонок; keyset-індекси `(lesson_id,created_at DESC,id DESC)` і `(parent_id,created_at,id)` | `.../postgres/repository_integration_test.go:104` `t.Run "K08 constraints and keyset indexes"` | integration | ✔ D02 |
| **K09** — корені по 50 курсором найновіші; відповіді найстаріші; прибраний корінь — заглушка/зникає | `.../postgres/repository_integration_test.go:184` `t.Run "K09 roots cursor replies and removed shape"` | integration | ✔ D02 |
| **K09a** — фільтрований селектор «поточне відео» (`processing` N+1 не вибрано); дві одночасні відповіді на 199 → рівно одна проходить | `.../postgres/repository_integration_test.go:238` `t.Run "K09a filtered selector and serialized cap"` (+ вкладений `:253` «two connections prove the root lock closes the race») | integration | ✔ D02 |
| **K10** — `GET`: 200 / 401 анонім / 403 `course_forbidden` / 404 `lesson_not_found` | `backend/internal/discussion/adapters/http/handler_test.go:153` `TestK10GetComments` | contract | ✔ D03 |
| **K11** — `POST`: 201 / 400 три коди валідації / 403 `origin_forbidden` і `course_forbidden` | `.../http/handler_test.go:182` `TestK11PostComments` | contract | ✔ D03 |
| **K12** — `PATCH`: лише автор; 403 `comment_forbidden`; 409 `comment_removed`; ставить `editedAt` | `.../http/handler_test.go:217` `TestK12PatchComments` | contract | ✔ D03 |
| **K13** — `DELETE`: автор/адмін/власник 204 тіло стерто; сторонній 403; повторне 204 | `.../http/handler_test.go:268` `TestK13DeleteComments` | contract | ✔ D03 |
| **K14** — `canEdit`/`canRemove` обчислені сервером для конкретного глядача | `.../http/handler_test.go:310` `TestK14CapabilitiesAndRemovedAuthorForm` | contract | ✔ D03 |
| **K15** — чотири операції в OpenAPI, фронт на Orval, `check-generated` зелений | `backend/openapi/openapi.yaml:916/943/973/1002` (`listLessonComments`/`publishLessonComment`/`updateLessonComment`/`removeLessonComment`) + `frontend/orval.config.ts` + `Makefile` таргет `check-generated` у ланцюгу гейта | gate | ✔ D03 |
| **K16** — панель: завантаження/порожньо/помилка; відповідь/зміна/прибирання за можливостями | `frontend/src/presentation/learn/DiscussionPanel.test.tsx` серія `K16` (`:73,79,85,101,130,148,158,173,189`) | unit (front) | ✔ D04 |
| **K17** — мітка = кнопка `onSeek(positionMs)`; «поточний момент» бере позицію плеєра | `frontend/src/presentation/learn/DiscussionPanel.test.tsx:50` і `:67` | unit (front) | ✔ D04 |
| **K18** — збій обговорення не ламає урок (плеєр і транскрипт рендеряться) | `frontend/src/routes/route-guards.test.tsx:625` `K18 keeps the lesson usable when discussion loading fails` | unit (front) | ✔ D04 |
| **K19** — студент публікує з міткою; інший бачить, перемотує кліком, відповідає | `frontend/e2e/discussion.spec.ts:160` `K19 publishes a timestamped comment, seeks, and replies` | e2e | ✔ D05 |
| **K20** — автор змінює/прибирає свій; власник курсу — чужий; адмін — будь-який | `frontend/e2e/discussion.spec.ts:214` `K20 enforces author, course-owner, and admin moderation` | e2e | ✔ D05 |
| **K21** — студент на чужій чернетці: обговорення недоступне тією ж відмовою (403) | `frontend/e2e/discussion.spec.ts:276` `K21 gives the same access denial and 403 for discussion` | e2e | ✔ D05 |
| **K22** — клавіатура, фокус, ≥44px, 200% зум, light/dark; axe без порушень на панелі | `frontend/e2e/discussion.spec.ts:301` `K22 keeps discussion keyboard accessible, sized, zoom-safe, and axe-clean` | e2e (a11y) | ✔ D05 |
| **K23** — e2e вибирає `role="status"` **у межах** панелі; run-scoped дані прибрані після прогону | `frontend/e2e/discussion.spec.ts:326` (`panel`) + `:327` (`panel.getByRole('status')`) + `test.afterAll:136` (published→archived, drafts→deleted) | review | ✔ D05 |

Примітка до рівнів: **K15 (gate)** і **K23 (review)** доводяться артефактами й перевірками у складі
наявних e2e-тестів, а не окремими іменованими тест-функціями — так вони й заявлені у спеці §6.

## 3. Шість продуктових дефектів фази — кожен із інструментом, що його знайшов

Не всі знайшов e2e; роль e2e тут не перебільшуємо.

1. **Живий коментар із NULL-тілом проходив CHECK** — знайшло **рев'ю міграції** (D02): CHECK на
   довжину не спрацьовував на `NULL`; додано `body IS NOT NULL`.
2. **Однакова доступна назва двох кнопок «Post comment»** (корінь і форма відповіді) — знайшли
   **розширені unit-тести панелі** (D04): дві кнопки з однаковим accessible name.
3. **Ручна пагінація дублювала сторінку й ховала новий корінь** — знайшли **unit-тести + розбір
   механізму** (D04): курсор рахувався так, що перша сторінка поверталася двічі.
4. **«Відповісти» на відповіді, яку сервер завжди відхиляє** — знайшло **рев'ю коду Courses 2**
   (D04): UI показував reply-кнопку на пласких відповідях, хоча домен їх відхиляє (`comment_parent_invalid`).
5. **Мітка невидима до Play** — знайшов **e2e на перезібраному прод-фронті** (D05): timestamp
   рендерився лише після старту плеєра; виправлено — показ до Play.
6. **`PATCH` розкривав факт прибирання не-автору** — знайшло **рев'ю Courses 2** (мінор D03,
   закрито в D04): різниця відповідей 403 vs 409 витікала стороннім.

## 4. Верифікація — упорядкований `make check` на HEAD

`make check` = `format-check lint typecheck test integration check-generated build up drain-queue
seed-identities seed smoke e2e` — весь ланцюг до кінця. HEAD = `02d221d` (D05); D06 додає лише
документи (звіт, `tasks.md`, AGENDA), код не змінює.

Результат на HEAD `02d221d`, 2026-09-29, з тимчасовим `DOCKER_CONFIG` без `credsStore` (обхід
хангу `docker-credential-desktop`, пам'ять `docker-credential-helper-hang`):

- **exit code 0**; увесь ланцюг пройшов упорядковано до кінця.
- backend: **501 тестів, 0 впало** (`failedTests: 0`), integration включно.
- e2e: **79 passed (11.5 хв), 0 failed, 0 flaky**.
- Чотири e2e-рядки фази зелені в цьому прогоні: K19 `:160`, K20 `:214`, K21 `:276`, K22 `:301`.
- **B39** (`playback.spec.ts:964`, історично крихкий, B-23) — **✓** (тест 56, 12.4s).

Зауваження до методу: перший запуск гейта впав на `build` через мій же обхід — порожній
`DOCKER_CONFIG` сховав від docker CLI плагін `docker compose`. Правильний обхід зберігає
`~/.docker/cli-plugins` і прибирає лише `credsStore`. Зелений прогон — після виправлення.

## 5. Обмеження (переносяться у продукт, не «закриті»)

- **Прибирання безповоротне й без `removed_by`** — сказано замовнику (спека §8, X5). Помилкове
  прибирання власником курсу непростежуване; аудит модерації — робоче місце інструктора.
- **Per-user ліміту публікацій немає** — єдиний запобіжник флуду `nginx zone=api` 10 r/s, burst 20,
  **за IP** (`nginx.conf:21,117`). Кілька акаунтів з одного NAT або повільний спамер під лімітом
  проходять. Per-user ліміт і модерація масштабу — робоче місце інструктора.
- **Пагінації відповідей немає** — корінь віддається повністю, обмежено cap 200 (≈400K символів у
  найгіршому `GET`; спека §7 п.2). Механізм, не гіпотеза: cap атомарний (K09a).
- **Селектор «поточне відео» — названий борг (B-22):** K09a пінить лише копію в `discussion`; копії
  поза модулем тримаються коментарями. Форму зафіксовано; уніфікація — окремий беклог-пункт.
- **K19 «інший акаунт» = `instructor`** — другого `student` у сідах немає; поведінка «інший глядач
  курсу» доводиться акаунтом інструктора з доступом.
- **`display_name` видно всім глядачам курсу** — публічність імені ширша, ніж у Phase 3 (спека §7 п.4).

## 6. B-23 (e2e B39) — оновлення таблиці

Додано рядок про повний гейт D06 у `work/backlog.md`: **повний гейт D06, 2026-09-29 — B39 1 ✓**.
Це шоста чиста точка поспіль у повному гейті (D05, D06 та раніші); механізм і далі не знайдено, тож
запис лишається відкритим. Правило перевірки лишається: якщо B39 впаде — A/B **повними** гейтами HEAD-до-D04 проти HEAD, по 2;
ізоляція гіпотез не розрізняє (пам'ять `e2e-isolated-rerun-needs-reseed`).

## 7. Наступне

За роадмапом (`work/PRODUCT-BRIEF.md` §24): робоче місце інструктора (п.3), оплата — остання фаза.
Comment threads 7b завершено; 7a (транскрипт) і 8 (ASR) закриті раніше. AGENDA оновлено.
