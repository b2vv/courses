# Робоче місце інструктора — план (ЯК)

| | |
|---|---|
| Фаза | Робоче місце інструктора (`PRODUCT-BRIEF.md` §24 п.3) |
| Спека | `docs/instructor/spec.md` (GATE 1 пройдено 2026-09-29) |
| Статус | **GATE 2 — знахідки закрито** (критика+захист свіжими сабагентами, 2026-09-29); до повторної звірки й тасків |
| Автор | Courses 1 |
| Гілка | `feat/instructor-workspace` |

Усі file:line у reuse-таблиці нижче звірені напряму з кодом (модуль `discussion`, reuse-index
маніфесту, frontend-патерни `routes/admin*`); проміжна звірка велася в ефемерному робочому файлі,
але джерело істини — код, і саме на нього вказують посилання.

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
  лише тому, хто має `access.CanView` на курс коментаря (інакше 403 `course_forbidden`); автор **не**
  репортить власний (403/названий код); неіснуючий `commentId` → 404 `comment_not_found`; репорт на
  вже прибраний → no-op (репорт не створюється). **Ідемпотентність без TOCTOU:** `INSERT … ON CONFLICT
  (comment_id, reporter_id) DO NOTHING`; повторний виклик того самого користувача → той самий успіх
  (201/200), кількість репортів не зростає, DB-помилка назовні не тече.
- **Teach (власник/адмін, під `/teach`):**
  - `GET /api/v1/teach/courses/{courseId}/comments` — course-wide список, курсор, найновіші першими;
    **невалідний курсор → 400** (як 7b list). Плаский список (roots+replies разом за `created_at DESC`);
    кожен рядок несе урок і `parent_id` (репліка показує посилання на корінь для контексту).
  - `GET /api/v1/teach/courses/{courseId}/moderation-queue` — коментарі з **відкритими** скаргами й
    `removed_at IS NULL`, з лічильником відкритих скарг. **Впорядкування (зафіксовано):** за лічильником
    відкритих скарг DESC, тайбрейк — час найновішої відкритої скарги DESC.
  - `POST /api/v1/teach/comments/{commentId}/dismiss-report` — відхилити **всі відкриті** скарги на
    коментар; тіло несе `reason` (обов'язкове, 1..500 рун, trim; порожнє/задовге → 400).
  - Прибрання з /teach переюзує наявний `DELETE /api/v1/learn/comments/{commentId}` (owner уже
    авторизований) — **без другого remove-ендпоінта**; `Remove` тепер пише `removed_by` + журнал.
- Усі teach-роути owner-скоуповані: `Authenticate` + перевірка `CourseFacts.OwnerID == actor` ∨ admin,
  та сама названа відмова (`course_forbidden`), що в 7b/9.

## Модель черги, dismiss і журнал (закриває GATE 2 blocking L1, L2, L3)

- **Стан скарги (L1).** `discussion_comment_reports` несе `dismissed_at timestamptz NULL` +
  `dismissed_by uuid NULL`. **Відкрита скарга** = `dismissed_at IS NULL`. Dismiss = `UPDATE … SET
  dismissed_at=now, dismissed_by=actor WHERE comment_id=$ AND dismissed_at IS NULL` (усі відкриті
  разом). **Черга** = коментарі, де `EXISTS (report: dismissed_at IS NULL)` **і** `removed_at IS NULL`.
  Наслідки, зафіксовані явно: (а) прибрання коментаря будь-яким шляхом (`removed_at`) прибирає його з
  черги без окремої дії; (б) **новий репорт після dismiss** має `dismissed_at IS NULL` → коментар
  **знову з'являється** в черзі (свіжий сигнал реанімує) — це намір, не баг.
- **Джерело `course_id` для журналу (L2).** Read коментаря, що його вживає `Remove`, тепер селектить
  `courses.id` (додається до наявного join у `lessonFacts`/`Get`, `repository.go:20-32,62-70`), щоб
  `Remove` мав `course_id` для запису журналу. Для teach-шляхів `course_id` і так у path.
- **Дискримінатор «модерація vs самоприбрання» (L3).** Запис у журнал модерації створюється **лише
  якщо `actor.UserID != comment.AuthorID`** (тобто діє не автор), незалежно від ролі. Тож власник, що
  прибирає **власний** коментар (owner==author), журналу не пише (W12); власник, що прибирає **чужий**,
  пише (W13). `removed_by` записується завжди = actor.
- **Гонка remove+dismiss (Q3 захисту).** Дії пишуть у незалежні стовпці (`removed_at` проти
  `reports.dismissed_at`); підсумковий стан черги консистентний за будь-якого порядку: прибраний
  коментар не в черзі незалежно від dismiss. Design-safe, названо явно.

## Side-effects

- **Зміна спільного персистованого стану** (`discussion_comments` +`removed_by`; нові таблиці). Зовнішня
  поведінка 7b незмінна (безповоротність, заглушка, коди) — додається лише стовпець і запис.
- Новий student-facing запис (report) — нова поверхня навантаження; захист — унікальність +
  nginx `zone=api` (per-user ліміту немає, успадкована межа 7b §5).
- **Крос-фазова зміна:** таблицю коментарів ділять 7b і ця фаза → верифікація **повним** e2e-набором,
  не таргетним (урок 7b: рев'ю дифу поділяє сліпу пляму; `AGENTS.md` Decisions).

## План міграції — `00010_instructor_moderation.sql`

- `ALTER TABLE discussion_comments ADD COLUMN removed_by uuid NULL REFERENCES users(id)`. Nullable —
  наявні рядки не ламаються; наявний CHECK `(removed_at IS NULL)=(body IS NOT NULL)` не чіпаємо. Додаємо
  інваріант **`CHECK (removed_by IS NULL OR removed_at IS NOT NULL)`** (E3): є хто прибрав — є коли.
- `CREATE TABLE discussion_comment_reports (id, comment_id FK→discussion_comments ON DELETE CASCADE,
  reporter_id FK→users, created_at, dismissed_at timestamptz NULL, dismissed_by uuid NULL REFERENCES
  users(id), UNIQUE(comment_id, reporter_id))`. **CASCADE живий, не мертвий (спростування R1):**
  `discussion_comments.lesson_id` має `ON DELETE CASCADE` (00009:10) — видалення уроку хард-видаляє
  коментарі, тож їхні репорти мусять каскадувати разом.
- `CREATE TABLE discussion_moderation_log (id, course_id, comment_id, actor_id, actor_role text,
  action text, reason text NULL, created_at)`; індекс `(course_id, created_at DESC, id DESC)`.
- Індекс для course-wide списку: keyset за курсом `(course_id, created_at DESC, id DESC)` — потребує
  `course_id` у рядку коментаря або join; форма за `EXPLAIN` у слайсі БД. Черга спирається на
  `reports(comment_id) WHERE dismissed_at IS NULL` — частковий індекс на відкриті скарги.
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
| W04 | Черга модерації | Курс, де є коментарі з відкритими скаргами | Відкриває чергу | Список коментарів із відкритими скаргами й `removed_at IS NULL`, з лічильником відкритих скарг; впорядковано за лічильником DESC, тайбрейк — час найновішої відкритої скарги DESC | integration; e2e |
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
| W20 | Недоступна залежність | БД/Redis недоступні | Будь-яка операція робочого місця | 503 з названим кодом `dependency_unavailable` (шлях `ErrDependencyUnavailable`, як 7b), не збій. `Retry-After` тут **не** обіцяємо: він є лише в nginx `@unavailable` для media (`nginx.conf:75`), не в Go-хендлерах | contract/integration |
| W23 | Новий репорт реанімує чергу | Скарги на коментар відхилено (dismiss); коментар живий | Інший студент репортить той самий коментар | Коментар знову з'являється в черзі (нова відкрита скарга) | integration |
| W24 | Репорт на неіснуючий коментар | — | Репорт із невідомим `commentId` | 404 `comment_not_found`; нічого не створено | contract |
| W25 | Невалідний курсор огляду | Власник | Запит зі зламаним курсором | 400 (як 7b list); даних не віддано | contract |
| W26 | Гонка remove+dismiss | Репортований коментар | Одночасні прибрання і dismiss | Підсумок консистентний: прибраний коментар не в черзі незалежно від порядку | integration |

**Крос-зрізові**

| # | Сценарій | Передумова | Дія | Очікуваний спостережуваний результат | Чим перевіряється |
|---|---|---|---|---|---|
| W21 | a11y робочого місця | Сторінка `/teach` із чергою й оглядом | Клавіатура/зум/тема | Клавіатурна навігація, видимий фокус, ≥44px, 200% зум, light/dark; axe без порушень | e2e (a11y) |
| W22 | Контракт | Нові teach/report-операції | Кожна відповідь | Звірена з OpenAPI (`assertContract`, як 7b K15); `check-generated` зелений | contract; gate |

**Ручних рядків: 0** — усе покривається unit/integration/contract/e2e. Групи happy/межі/відмови всі
присутні.

## GATE 2 — знахідки й закриття (свіжі сабагенти, 2026-09-29)

Вердикт виносили не автор плану: `plan-critique` і `plan-defense` окремими сабагентами з чистим
контекстом. Обидва були **червоні**; нижче — що знайшли й де закрито.

| Знахідка | Гейт | Рівень | Закрито в плані |
|---|---|---|---|
| L1 — dismiss не прибирає коментар із черги (reports без стану) | critique | blocking | «Модель черги…»: `dismissed_at`/`dismissed_by`, запит черги, реанімація новим репортом |
| Q4/A2 — впорядкування черги не зафіксоване | defense/critique | red/nb | Ендпоінти + W04: за лічильником DESC, тайбрейк найновіша скарга |
| Q3/E2 — ідемпотентність репорту (TOCTOU), гонка remove+dismiss | defense/critique | red/nb | Ендпоінти: `ON CONFLICT DO NOTHING`; «Модель черги…»: гонка design-safe |
| Q1 — reason bound, репорт на неіснуючий, невалідний курсор | defense | red | Ендпоінти: reason 1..500, 404 `comment_not_found`, 400 на курсор (W24/W25) |
| L2 — `moderation_log.course_id` нема звідки | critique | nb | «Модель черги…»: read додає `courses.id` |
| L3 — W12/W13 колізія owner==author | critique | nb | «Модель черги…»: журнал лише якщо actor≠author |
| A1 — W20 «Retry-After як наявний» хибно | critique | nb | W20 виправлено (Retry-After лише nginx media, звірено `nginx.conf:75`) |
| E3 — `removed_by` без прив'язки до `removed_at` | critique | nb | Міграція: `CHECK (removed_by IS NULL OR removed_at IS NOT NULL)` |
| R1 — reports CASCADE нібито мертвий | critique | nb | **Спростовано:** `discussion_comments.lesson_id ON DELETE CASCADE` (00009:10) → CASCADE живий; лишено з поясненням |
| E4 — плаский список відриває репліки від кореня | critique | nb | Ендпоінти: рядок несе урок + `parent_id` для контексту |

Підтверджено чистим (перевірки, що витримали): owner-guard тримається **всередині** discussion без
імпорту `course` (adapter читає `courses` напряму, прецедент 7b); міграція forward-only; `removed_by`
nullable не ламає наявний CHECK; переюз `Remove` не ламає семантику 7b; over-engineering не знайдено;
рішення «розширити discussion» обґрунтоване кодом.

## План-секція майбутнього звіту

Звіт (`ship-report`) буде дельтою до цього плану: що з reuse/нового справдилось, які приймальні
рядки закриті яким тестом, які межі перенесені далі. Другий план у звіті не пишемо.
