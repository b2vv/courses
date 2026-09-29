# Робоче місце інструктора — звіт

| | |
|---|---|
| Фаза | Робоче місце інструктора (`PRODUCT-BRIEF.md` §24 п.3) |
| Дата закриття | 2026-09-29 |
| Автор | сесія Courses 1 (план, гейти, коміти); код — Codex (`gpt-5.6-luna` I01–I04, `gpt-5.6-sol` I05); рев'ю — Courses 2 / adversarial-гейти |
| Підстава | `docs/instructor/spec.md` (GATE 1), `docs/instructor/plan.md` (GATE 2 зелений); ADR 013 |
| Роль фази | Передостання фаза роадмапу перед оплатою. Дає власнику курсу одну поверхню `/teach/courses/{id}` для модерації студентської активності по всьому курсу, замість пер-урочного прибирання 7b (§24 п.4). Розширює модуль `discussion`; ownership лишається читанням `courses.instructor_id` **всередині** discussion (architecture-тест). |

Звіт — **дельта до плану** (`plan.md`), не другий план.

## 1. Що зроблено (I01–I06)

- **I01** `46a94a2` — домен `discussion`: `Report` (`NewReport` відхиляє власний), `Course.CanBeModeratedBy` (owner∨admin), `IsModeration` (actor≠author — дискримінатор журналу), `Comment.RemovedBy`. Unit-таблиці.
- **I02** `3fb51ee` — міграція `00010` (removed_by+CHECK, `discussion_comment_reports` з `dismissed_at`/UNIQUE/CASCADE, `discussion_moderation_log`, keyset+partial індекси) + `ModerationRepository` (CourseFacts, ListByCourse, InsertReport ON CONFLICT, DismissReports, ListQueue). Integration. **Гейт зловив 2 дефекти** (нижче §3).
- **I03** `3b7b74c` — application (report, course-comments, queue, dismiss, remove-with-audit) + HTTP (`/teach` три роути + report під `/learn`) + OpenAPI/Orval + `assertContract` + architecture-тест «discussion не імпортує course».
- **I04** `f3b7302` — фронт `/teach` (guard+loader за зразком admin, огляд+черга, dismiss/remove), teach domain/port/adapter/queries, кнопка «Поскаржитись». **Гейт зловив 1 дефект** (R24 route-tree).
- **I05** `1de3384` — e2e `instructor.spec.ts` (8 журнеїв) + **2 продуктові дефекти, викриті e2e** (W01, W05, §3).
- **I06** (цей коміт) — упорядкований `make check`, цей звіт W01–W26 по одному, `work/tasks.md`, AGENDA/брифінг/AGENTS, PR у main.

## 2. W01–W26 — по одному рядку, кожен своїм доказом

Зелений `make check` — не доказ рядка; нижче кожен рядок вказує на власний тест. Рівні як у плані.

Незалежна звірка (Explore) знайшла, що частина рядків спершу не мала тесту або доводилась лише
доменним правилом; прогалини закрито gap-тестами (I06), нижче кожен рядок має свій доказ.

| Рядок | Доказ (файл:рядок, тест) | Рівень |
|---|---|---|
| **W01** — огляд усіх уроків курсу, найновіші першими, урок/автор/час/тіло | e2e `instructor.spec.ts:181`; integration `repository_integration_test.go:84` `ListByCourse spans lessons`; rstest `teach-queries.test.tsx:18` | e2e+integration+rstest |
| **W02** — курсорна пагінація, без повтору сторінки | integration `repository_integration_test.go:84` (limit=2, NextCursor, друга сторінка) | integration |
| **W03** — студент репортить; повтор не множить | contract `handler_test.go:403`; e2e `instructor.spec.ts:236`; rstest `DiscussionPanel.test.tsx:196,221` | contract+e2e+rstest |
| **W04** — черга з лічильником, порядок count DESC | integration `repository_integration_test.go:163`; e2e `instructor.spec.ts:236`; rstest `teach-queries.test.tsx:18` | integration+e2e+rstest |
| **W05** — прибрання з робочого місця → заглушка, зникає з черги | e2e `instructor.spec.ts:289`; contract `handler_test.go:307` `TestK13 204 course owner`; integration `repository_integration_test.go:163`; rstest `teach-queries.test.tsx:18` (інвалідація teach+learn) | e2e+contract+integration+rstest |
| **W06** — dismiss із причиною, коментар лишається | contract `handler_test.go:461` (204), `:467` (400 invalid reason); e2e `instructor.spec.ts:326`; integration `repository_integration_test.go:163` | contract+e2e+integration |
| **W07** — порожній огляд | e2e `instructor.spec.ts:371`; rstest `teach-queries.test.tsx:18` | e2e+rstest |
| **W08** — порожня черга | e2e `instructor.spec.ts:371`; rstest `teach-queries.test.tsx:18` | e2e+rstest |
| **W09** — дубль репорту ідемпотентний | integration `repository_integration_test.go:113` `InsertReport is idempotent`; e2e `instructor.spec.ts:236` | integration+e2e |
| **W10** — репорт на прибраний → no-op | integration `repository_integration_test.go:131` `W10 reporting a removed comment is a no-op` | integration |
| **W11** — прибрання (7b-шлях) знімає з черги | integration `repository_integration_test.go:163` (removed_at → зникає з черги) | integration |
| **W12** — самоприбрання автором **не** пише журнал | contract `handler_test.go:349` `TestW12W13ModerationLogEffects` (`:356` author, `wantLog=false`); правило unit `report_test.go:49` | contract+unit |
| **W13** — прибрання власником/адміном пише журнал (actor, action) | contract `handler_test.go:357,358` (owner/admin, `wantLog`, `wantRole`); правило unit `report_test.go:49` | contract+unit |
| **W14** — не-власник: названа відмова / 403 course_forbidden | e2e `instructor.spec.ts:398`; contract `handler_test.go:434`; rstest `route-guards.test.tsx:465` | e2e+contract+rstest |
| **W15** — власник іншого курсу відмовлено | contract `handler_test.go:506` `W15 other course owner is forbidden`; правило unit `report_test.go:29` | contract+unit |
| **W16** — адмін проходить на будь-який курс | e2e `instructor.spec.ts:426`; contract `handler_test.go:438` (queue admin 200); unit `report_test.go:29` | e2e+contract+unit |
| **W17** — репорт без доступу → 403, без запису | contract `handler_test.go:470` `W17 report without course access is forbidden without a write` | contract |
| **W18** — репорт власного → відмова, без запису | contract `handler_test.go:409` (403 report_own_comment); unit `report_test.go:8` | contract+unit |
| **W19** — модерація не власником: dismiss→course_forbidden, remove→comment_forbidden | contract `handler_test.go:556` `W19 dismiss by another course owner` + `:308` `TestK13 403 other` (remove); unit `report_test.go:29` | contract+unit |
| **W20** — недоступна залежність → 503 dependency_unavailable | contract `handler_test.go:510` `W20 unavailable teach dependency is named` | contract |
| **W21** — a11y на /teach: клавіатура/фокус/≥44px/зум/тема/axe | e2e `instructor.spec.ts:450` | e2e (a11y) |
| **W22** — кожна відповідь звірена з OpenAPI | contract `handler_test.go:157` `serve()`→`assertContract` на кожному підтесті `TestI03ModerationContracts` | contract+gate |
| **W23** — новий репорт після dismiss реанімує чергу | integration `repository_integration_test.go:163` (після dismiss новий InsertReport повертає в чергу) | integration |
| **W24** — репорт на неіснуючий → 404 comment_not_found | contract `handler_test.go:413` | contract |
| **W25** — невалідний курсор → 400 comment_cursor_invalid | contract `handler_test.go:457` | contract |
| **W26** — гонка remove+dismiss: консистентно в обох порядках | integration `repository_integration_test.go:207` `W26 remove and dismiss converge in either order` | integration |

Примітка: W22 і частина W05/W12/W13/W19 доводяться через `assertContract`/journal-fake на HTTP-рівні, як і
заявлено у спеці §6. Gap-тести W10/W12/W13/W15/W17/W19-dismiss/W20/W26 додані в I06 після незалежної
звірки — жоден не викрив продукт-дефекту (логіка була вірна, бракувало покриття).

## 3. Продуктові дефекти фази — кожен із інструментом, що його знайшов

Роль e2e не перебільшуємо; більшість дефектів зловили інші гейти.

1. **W01 — огляд не показував ні уроку, ні моменту** — знайшов **e2e на прод-фронті** (I05): course-wide DTO не мав `lessonTitle`, а рядок рендерив лише автора+тіло. Закрито наскрізно (read `course_lessons.title` через окремий `ModerationCommentRecord`, DTO/OpenAPI/Orval, UI + спільний `format-timestamp`).
2. **W05 — прибрання не оновлювало teach-чергу** — знайшло **рев'ю коду Codex sol** при написанні e2e: черга кликала `learn.discussion.remove` (інвалідує лише learn). Закрито `TeachRepository.removeComment` + teach-мутація з інвалідацією teach/moderation+comments.
3. **SQL: `$2::timestamptz` замість `$2::int` у guard курсору черги** — знайшов **integration-гейт** (I02): маскувалось у `dependency_unavailable`, локальний `go build` Codex не бачив.
4. **Ізоляція integration-тесту черги** — знайшов **integration-гейт** (I02): підтест успадкував відкриту скаргу зі спільного `moderationCourse`.
5. **R24 route-tree не враховував нові teach-роути** — знайшов **rstest** (I04): правило «хто додає глобально лічене — прибирає».

Уроки: локальний `go build`/`go test` у пісочниці Codex не замінює справжніх docker-гейтів — дефекти 3–4 пройшли б «зеленими» повз Codex. «Компілюється» ≠ «integration/e2e проходить».

## 4. Верифікація — упорядкований `make check` на HEAD гілки

`make check` до кінця на HEAD гілки (після I06 gap-тестів), `DOCKER_CONFIG` без `credsStore` (обхід хангу):
- **exit 0**; увесь ланцюг ordered. Прогін включає всі gap-тести (W10/W12/W13/W15/W17/W19/W20/W26).
- backend Go: **0 FAIL** (unit + integration, включно з discussion-integration 7b+I02+I03).
- frontend rstest: **505 passed, 0 failed**.
- e2e: **87 passed (11.7 хв), 0 failed, 0 did-not-run** — 8 нових instructor-рядків (W01/W03-W08/W14/W16/W21) + повна регресія (7b discussion, transcript, playback, admin).
- `check-generated` зелений (OpenAPI/Orval/sqlc без дрейфу після нових teach+report операцій і `lessonTitle`).
- architecture-тест: `discussion` не імпортує `course` — ownership лишається читанням таблиці.

Крос-фазова безпека: зміну спільного стану (`discussion_comments` read-шлях, спільний `format-timestamp` у Discussion/Transcript/Teach, сигнатура `Remove`) перевірено **повним** e2e-набором, не таргетним.

## 5. Обмеження (перенесені в продукт, не «закриті»)

- **Блокування авторів — відкладено до enrollment** (§24 п.8): без enrollment немає per-course суб'єкта блокування. Модерація фази — на рівні коментарів, не авторів (рішення GATE 1).
- **Огляд навчання (студенти з прогресом) — поза фазою** (рішення GATE 1): без enrollment «студент курсу» напівозначений; ядро фази — модерація.
- **Прибрання лишається безповоротним**; фаза додала лише **слід** (`removed_by` + журнал модерації для actor≠author) — «оживлення» немає, аудит мінімальний (без reason на remove, лише на dismiss).
- **Per-user ліміту публікацій/репортів немає** — лише nginx `zone=api` за IP; ідемпотентність репорту за UNIQUE(comment,reporter). Масштаб — майбутнє.
- **Модерація «моменту відео» в огляді — за наявності**: позиційна мітка вимагає ready-відео; модераційний огляд свідомо не втягує media-пайплайн (уточнення W01).

## 6. Наступне

За роадмапом (`PRODUCT-BRIEF.md` §24): **оплата — остання фаза**; модель ухвалена (§24 п.22), перше питання спеки — провайдер і його юрисдикція. Робоче місце інструктора закрите; перед оплатою відкритих фаз роадмапу більше немає. AGENDA/брифінг оновлено.
