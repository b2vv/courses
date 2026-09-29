# Робоче місце інструктора — план (ЯК)

| | |
|---|---|
| Фаза | Робоче місце інструктора (`PRODUCT-BRIEF.md` §24 п.3) |
| Спека | `docs/instructor/spec.md` (GATE 1 пройдено 2026-09-29) |
| Статус | **DRAFT** — до GATE 2 (`plan-critique` + `plan-defense` + конституція) |
| Автор | Courses 1 |
| Гілка | `feat/instructor-workspace` |

Джерело reuse-фактів — `scratchpad/instructor-reuse.md` (звірка модуля `discussion`, reuse-index,
frontend-патернів). Усі file:line нижче — з тієї звірки.

## Архітектурне рішення: РОЗШИРИТИ модуль `discussion` (не новий модуль)

Дата 2026-09-29, автор Courses 1. Закриває альтернативу «новий модуль `moderation`».

**Чому:** reports, журнал і `removed_by` — колонки/таблиці на тому самому агрегаті
`discussion_comments` (FK на `id`). `discussion` **уже** owner-aware (`CanBeRemovedBy` враховує
`course.OwnerID`, `domain/comment.go:76-81`) і **уже** читає ownership (`courses.instructor_id`→
`OwnerID`, `adapters/postgres/repository.go:20` `lessonFactsQuery`) через `contracts/access` (ADR 013).
Новий модуль мусив би читати коментарі крос-модульно — а модуль не імпортує модуль (AGENTS.md).
Структурних перешкод розширенню нема: SQL ручний (не sqlc), домен чистий.

**Перевірка рішення:** якщо будь-який новий код у цій фазі імпортує `discussion` ззовні модуля або
читає `discussion_comments` з іншого модуля — рішення порушено.

## Reuse-first (проти reuse-index маніфесту)

**Переюзуємо (не пишемо заново):**

| Що | Де | Як застосуємо |
|---|---|---|
| Домен коментаря, стан removed | `discussion/domain/comment.go:23-33` (`RemovedAt *time.Time`) | `removed_by` лягає поряд; стан «removed» вже є |
| Прибрання, owner-aware, ідемпотентне | `comment.go:76-81` `CanBeRemovedBy`; `application/application.go:159-171` `Remove` | розширюємо `Remove`, щоб писав `removed_by` + журнал; **та сама операція** для 7b і /teach |
| Курсор-список + cap | `adapters/postgres/repository.go:98-99` | патерн keyset для нового course-wide списку й черги |
| Факти курсу за уроком | `repository.go:20,76` `lessonFactsQuery`/`LessonFacts` | аналог `CourseFacts(courseId)` — читання `instructor_id`→`OwnerID` за courseId |
| HTTP error→код, sameOrigin, actor | `adapters/http/handler.go:23-28,145,251-264` | новий `RegisterTeach` за тим самим патерном; `statusFor` розширюємо |
| Envelope, RequestID, WriteError | `platform/httpserver/server.go:44,47,53` | без змін |
| Реєстрація модуля | `cmd/api/main.go` (discussion підключено; admin через `RegisterAdmin`) | `RegisterTeach(mux, Authenticate, RequireSameOrigin)` |
| Frontend route+guard+loader | `routes/admin.tsx` (beforeLoad guard), `routes/admin.courses.$courseId.tsx` (loader+useSuspenseQuery) | `teach.tsx` guard (instructor|admin) + `teach.courses.$courseId.tsx` |
| Orval-адаптер + query boundary | `adapters/api/discussion-repository.ts`, `app/composition/discussion-queries.ts` | новий адаптер/queries за зразком |
| UI primitives, transport | `presentation/ui/`, `adapters/api/transport.ts` | без змін |
| OpenAPI стиль | `openapi.yaml` admin-операції `*Admin*` під `/api/v1/admin` | teach-операції `*Teach*` під `/api/v1/teach` |

**Нове (і чому наявне не підійшло):**

| Нове | Чому не переюз |
|---|---|
| `CourseFacts(courseId)` у ports+postgres | зараз факти лише за `lessonId`; черга/огляд ідуть за курсом |
| `ListByCourse(courseId, cursor)` + course keyset-індекс | 7b має лише per-lesson список |
| Доменна функція course-level модерації (owner/admin над курсом) | `CanBeRemovedBy` — per-comment; для стрічки курсу треба guard рівня courseId, що лишається **в discussion** (не імпорт `course`) |
| Домен-концепт `Report` + таблиця + унікальність `(comment, reporter)` | у discussion немає репортів |
| Журнал модерації (таблиця) + запис у `Remove`/dismiss | немає аудиту (7b навмисно без нього) |
| `removed_by` колонка | 7b навмисно без неї (межа §5) — рішення GATE 1 розширює |
| Поверхня `/teach/*` | немає |

## Ендпоінти

- **Студент (під `/learn`):** `POST /api/v1/learn/comments/{commentId}/reports` — репорт. Дозволено
  лише тому, хто має `access.CanView` на курс коментаря; автор не репортить власний; унікальність
  `(comment, reporter)`; репорт на прибраний — no-op.
- **Teach (власник/адмін, під `/teach`):**
  - `GET /api/v1/teach/courses/{courseId}/comments` — course-wide список, курсор, найновіші першими.
  - `GET /api/v1/teach/courses/{courseId}/moderation-queue` — репортовані **не-прибрані** з лічильником.
  - `POST /api/v1/teach/comments/{commentId}/dismiss-report` — відхилити скарги на коментар (з причиною).
  - Прибрання з /teach переюзує наявний `DELETE /api/v1/learn/comments/{commentId}` (owner уже
    авторизований) — **без другого remove-ендпоінта**; `Remove` тепер пише `removed_by` + журнал.
- Усі teach-роути owner-скоуповані: `Authenticate` + перевірка `CourseFacts.OwnerID == actor` ∨ admin,
  та сама названа відмова (`course_forbidden`), що в 7b/9.

## Side-effects

- **Зміна спільного персистованого стану** (`discussion_comments` +`removed_by`; нові таблиці). Зовнішня
  поведінка 7b незмінна (безповоротність, заглушка, коди) — додається лише стовпець і запис.
- Новий student-facing запис (report) — нова поверхня навантаження; захист — унікальність +
  nginx `zone=api` (per-user ліміту немає, успадкована межа 7b §5).
- **Крос-фазова зміна:** таблицю коментарів ділять 7b і ця фаза → верифікація **повним** e2e-набором,
  не таргетним (урок 7b: рев'ю дифу поділяє сліпу пляму; `AGENTS.md` Decisions).

## План міграції — `00010_instructor_moderation.sql`

- `ALTER TABLE discussion_comments ADD COLUMN removed_by uuid NULL REFERENCES users(id)`. Nullable —
  наявні рядки не ламаються; CHECK `(removed_at IS NULL)=(body IS NOT NULL)` не чіпаємо.
- `CREATE TABLE discussion_comment_reports (id, comment_id FK→discussion_comments ON DELETE CASCADE,
  reporter_id FK→users, created_at, UNIQUE(comment_id, reporter_id))`.
- `CREATE TABLE discussion_moderation_log (id, course_id, comment_id, actor_id, actor_role text,
  action text, reason text NULL, created_at)`; індекс `(course_id, created_at DESC, id DESC)`.
- Індекс для course-wide списку/черги: keyset за курсом. Точна форма — за `EXPLAIN` у слайсі БД.
- Forward-only, як 00005–00009.

## Rollback

- Down-міграція: `DROP TABLE discussion_moderation_log, discussion_comment_reports; ALTER TABLE
  discussion_comments DROP COLUMN removed_by`. `removed_by` nullable → відкат не втрачає обов'язкових
  даних; журнал і репорти — новий стан, відкат їх прибирає повністю.
- Код за фіче-гілкою; `main` не чіпається до merge. Якщо репорти дають навантаження — вимикається на
  рівні route/endpoint без відкату міграції.
- Незворотність прибирання незмінна: down нічого в наявних коментарях не «оживляє».

## Конституція (AGENTS.md) — проведення

- **Домен** discussion: лише stdlib + `contracts`; `Report`, course-guard, журнал-модель — чисті типи.
- **Application** — домен+порти; нові операції (report, list-by-course, queue, dismiss) через порт.
- **Адаптери** всередину; generated Orval-моделі в адаптерах, явний мапінг у домен.
- **Модуль не імпортує модуль:** ownership лишається читанням `courses.instructor_id` **всередині**
  discussion, не імпортом `course`. ← головний ризик рішення, пінити architecture-тестом.
- **TS strict**, без `any`; no fetch у presentation; server state у Query; URL-стан у search params.
- **a11y/токени/no-store** нарівні з 7b.

## Приймальні сценарії (`acceptance-spec`)

Три групи представлені. Кожен рядок — одне спостережуване твердження. Рівень доказу як у 7b
(unit/integration/contract/e2e/a11y/gate). Таски імплементації посилаються на ці номери.

**Happy path**

| # | Сценарій | Передумова | Дія | Очікуваний спостережуваний результат | Чим перевіряється |
|---|---|---|---|---|---|
| W01 | Огляд коментарів курсу | Власник; курс із коментарями у ≥2 уроках | Відкриває `/teach/courses/{id}` | Один список коментарів **усіх уроків курсу**, найновіші першими; кожен рядок має урок, `display_name`, мітку часу, тіло (або заглушку, якщо прибрано), стан прибрання | e2e `frontend/e2e/instructor.spec.ts`; integration (`ListByCourse`) |
| W02 | Курсорна пагінація | Курс із >сторінки коментарів | Гортає далі | Наступна сторінка вантажиться курсором; перша сторінка не повторюється; весь курс не тягнеться одним запитом | integration (postgres) |
| W03 | Студент репортить | Студент із доступом до курсу; живий чужий коментар | Тисне «Поскаржитись» | Репорт прийнято (201); повторний виклик тим самим користувачем не додає другого | contract `discussion/adapters/http/handler_test.go`; e2e |
| W04 | Черга модерації | Курс, де є репортовані не-прибрані коментарі | Відкриває чергу | Список репортованих не-прибраних коментарів із лічильником скарг, найбільш/найновіші за домовленою впорядкованістю | integration; e2e |
| W05 | Прибрання з робочого місця | Власник; чужий живий коментар у черзі | Прибирає | Коментар стає заглушкою (як 7b); `removed_by`=власник; запис у журнал модерації; коментар зникає з черги | integration; e2e |
| W06 | Відхилення скарги | Власник; репортований коментар у черзі | Відхиляє скаргу з причиною | Коментар лишається живим, зникає з черги; запис у журнал (дія=dismiss, причина) | integration; e2e |

**Межі**

| # | Сценарій | Передумова | Дія | Очікуваний спостережуваний результат | Чим перевіряється |
|---|---|---|---|---|---|
| W07 | Порожній огляд | Курс без коментарів | Відкриває огляд | 0 рядків і текст порожнього стану | e2e |
| W08 | Порожня черга | Курс без репортів | Відкриває чергу | 0 рядків і текст порожньої черги | e2e |
| W09 | Дубль репорту | Студент уже репортив цей коментар | Репортить удруге | Кількість репортів на коментар не зростає (унікальність `(comment,reporter)`); відповідь ідемпотентна | contract; integration |
| W10 | Репорт на прибраний | Коментар уже прибрано | Репортить | No-op: репорт не створюється, коментар у черзі не з'являється | contract; integration |
| W11 | Прибрання шляхом 7b знімає з черги | Репортований коментар; власник відкриває **сторінку уроку** (7b) | Прибирає там | Коментар зникає і з черги робочого місця (той самий `removed_at`) | integration |
| W12 | Самоприбрання автором — не модерація | Автор свого коментаря | Прибирає свій (шлях 7b) | `removed_by`=автор; запис у журнал модерації **не** створюється (це не модерація) | integration |
| W13 | Прибрання власником — модерація | Власник; чужий коментар | Прибирає | `removed_by`=власник; запис у журнал модерації створюється | integration |

**Відмови**

| # | Сценарій | Передумова | Дія | Очікуваний спостережуваний результат | Чим перевіряється |
|---|---|---|---|---|---|
| W14 | Не-власник у робочому місці | Студент | Відкриває `/teach/courses/{id}` | Та сама названа відмова, що доступ у 7b/9 (403 `course_forbidden`); даних не видно | e2e; contract |
| W15 | Власник чужого курсу | Власник курсу A | Відкриває `/teach` курсу B | 403 `course_forbidden` | contract |
| W16 | Адмін на будь-який курс | Адмін | Відкриває `/teach` будь-якого курсу | Доступ дозволено | contract; e2e |
| W17 | Репорт без доступу до курсу | Користувач без `access.CanView` на курс | Репортить коментар | Відмова (403 `course_forbidden`), репорт не створено | contract |
| W18 | Репорт власного коментаря | Автор коментаря | Репортить свій | Відмова (названий код); репорт не створено | contract |
| W19 | Модерація не власником | Студент/чужий власник | dismiss/remove через teach-ендпоінт | 403 `course_forbidden`; стан не змінено | contract |
| W20 | Недоступна залежність | БД/Redis недоступні | Будь-яка операція робочого місця | 503 з названим кодом і `Retry-After`, не збій | integration/smoke (як наявний) |

**Крос-зрізові**

| # | Сценарій | Передумова | Дія | Очікуваний спостережуваний результат | Чим перевіряється |
|---|---|---|---|---|---|
| W21 | a11y робочого місця | Сторінка `/teach` із чергою й оглядом | Клавіатура/зум/тема | Клавіатурна навігація, видимий фокус, ≥44px, 200% зум, light/dark; axe без порушень | e2e (a11y) |
| W22 | Контракт | Нові teach/report-операції | Кожна відповідь | Звірена з OpenAPI (`assertContract`, як 7b K15); `check-generated` зелений | contract; gate |

**Ручних рядків: 0** — усе покривається unit/integration/contract/e2e. Групи happy/межі/відмови всі
присутні.

## План-секція майбутнього звіту

Звіт (`ship-report`) буде дельтою до цього плану: що з reuse/нового справдилось, які приймальні
рядки закриті яким тестом, які межі перенесені далі. Другий план у звіті не пишемо.
