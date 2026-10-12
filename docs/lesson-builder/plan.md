# lesson-builder — plan (ЯК)

Реалізація наміру з [spec.md](spec.md). Рішення/розвилки — [assumptions.md](assumptions.md).
Ґрунтовано на reuse-first дослідженні коду (3 агенти, 2026-10-12); усі anchor'и — file:line.

## Розміщення (за маніфестом)

- **Бекенд — розширюємо модуль `course`** (він володіє уроками й авторингом): нові файли
  `domain/document.go`, `application/document.go`, новий порт у `ports/`,
  `adapters/postgres/document.go`, `adapters/http/document_handler.go`.
- **Бекенд — новий малий модуль `internal/image`** (`domain/`, `application/`, `ports/`,
  `adapters/{http,filesystem,postgres}`): довільні зображення уроку. НЕ розширюємо `media`
  (його домен суто відео: VideoAsset/HLS/transcode/tus — домішування забруднило б когезивний домен).
- **Фронтенд:** reader-компонент `presentation/.../LessonDocumentView.tsx`; редактор —
  окремий маршрут `routes/admin.courses.$courseId.lessons.$lessonId.tsx` +
  `presentation/admin/LessonDocumentEditor.tsx` (lazy); домен `domain/lesson-document/`;
  порт+адаптер репозиторію; query-boundary у `app/composition/`.
- **Міграція:** `backend/migrations/00018_lesson_document.sql` (goose; наступний номер — поточний
  максимум 00017_refund).
- **nginx:** нова локація `^~ /media/i/` + окрема `location` для image-upload з більшим
  `client_max_body_size` (`/api/` стоїть на 1m — `nginx.conf:182`).

## Reuse-first

### Переюзуємо
| Концерн | Джерело | Нотатка |
|---|---|---|
| Контракт доступу (outline vs content) | `backend/internal/contracts/access/access.go` `CanConsume`/`CanView` | гейт нового body-ендпоінта й image-authorize |
| Патерн per-lesson-content read + gate | transcript: `adapters/http/handler.go:18`, `domain/access.go CanConsumeBy` | дзеркалимо для document read |
| jsonb через raw pgx | `transcript/adapters/postgres/draft.go:287-311` (asr_drafts.segments) | document body теж raw pgx + json.Marshal/Unmarshal, не sqlc |
| Атомарна зміна стану одним UPDATE | course `transitionAuthoringCourse` (`postgres/authoring.go:55-66`, COALESCE) | publish = `UPDATE SET published_body = draft_body` |
| Optimistic-lock token | course `Authoring.Save` `expectedUpdatedAt` (`application/authoring.go:100`) | save draft з перевіркою updated_at |
| filesystem storage (os.Root, 0750/0640, груп-читання nginx) | `media/adapters/filesystem/storage.go:19-43` | ПАТЕРН (свій адаптер — `resolve()` media відео-центричний) |
| nginx auth_request-патерн | `nginx.conf:67-119` (`/media/p/` → `/_media_auth`) | нова `/media/i/` → `/internal/image/authorize` |
| HTTP server/error-envelope/request-id, admin-only actor, requireSameOrigin | `platform/httpserver/`, `media/adapters/http/handler.go:66-101,145-160` | upload = multipart (не JSON-ticket, не tus) |
| HTTP транспорт, SSR cookie-forward, UI-примітиви, query-boundary, DI-wiring | reuse-index маніфесту (`transport.ts`, `request-headers.ts`, `presentation/ui/`, `catalog-queries.ts`, `app/composition/`) | фронт |
| Клавіатурне впорядкування move-up/down | `CourseEditorView.tsx:380/391` (`lessons.move`) | секції впорядковуємо тим самим ПІДХОДОМ (move-команди, без DnD-ліби), але НЕ копіюючи `isDisabled` на кнопках — див. нижче B2 |
| Port-injection важкої ліби (щоб не тягнути в main-chunk) | `LessonPlayer.tsx:36/69` (createMedia як prop) | Tiptap тільки в lazy editor-піддереві |
| react-hook-form | `CourseEditorView.tsx:3` (7.87, useFieldArray) | форма редактора |

### Нове
- **Домен документа уроку** (`course/domain/document.go`, stdlib): `LessonDocument{SchemaVersion,
  Locale, Sections[]}`; Section — дискримінований union (`video`/`text`/`image`/`row`/`task`);
  **власна валідована rich-text-модель** (абзаци, h2/h3, bold/italic, ul/ol, link http(s)) —
  НЕ сирий ProseMirror/Tiptap JSON. Валідація дерева в домені: відомі типи, кардинальність
  (`video` 0..1), `row` 1–3 колонки діти лише text/image depth-1, allowlist марок, екранований
  текст, валідні URL.
- **Таблиця `lesson_documents`** (див. Міграція): PK `(lesson_id, locale)`, FK→`course_lessons`
  ON DELETE CASCADE, `draft_body jsonb`, `published_body jsonb`, `updated_at`. PK з `locale` робить
  мультимовність майбутнім INSERT-ом рядка, без зміни схеми (Q3 locale-aware).
- **Ендпоінти:** admin `PUT /api/v1/admin/lessons/{id}/document` (save draft, optimistic-lock),
  `POST .../document/publish` (атомарно); learn `GET /api/v1/learn/lessons/{id}/document` (gated
  `CanConsume`, **окремо** від outline — не в `getLearningCourse`); image
  `POST /api/v1/admin/lessons/{id}/images` (multipart), internal `GET /internal/image/authorize`.
- **Модуль `internal/image`:** домен Image (id, lessonId, contentType, w, h, byteSize, createdBy);
  multipart-upload (magic-bytes allowlist jpeg/png/webp, MaxBytesReader, піксельний бюджет проти
  decompression-bomb); filesystem-store `/media/images/<id>`; authorize через свій reader
  lesson→course → `access.CanConsume` (БЕЗ playback-lease, БЕЗ обов'язкового EmailVerified).
- **Фронтенд reader** `LessonDocumentView` — чистий компонент секції→семантичний HTML, нуль
  редакторських залежностей; слот у `CourseLearningView` поряд із `LessonPlayer`
  (`learn.$` OpenLesson ~266-277); `<img src="/media/i/<id>">`. **Відсутній рядок `lesson_documents`
  або 0 секцій → порожній стан** (не 404, не краш): урок без збереженого/не-мігрованого документа
  рендериться так само, як порожній (сценарій №4). Ендпоінт віддає порожній документ, не 404.
- **Фронтенд editor** — Tiptap (MIT: `@tiptap/core`+`@tiptap/pm`+потрібні extensions) у
  `React.lazy(() => import('./LessonDocumentEditor'))`; секції add/reorder move-up/down; image-upload
  multipart; явні «Зберегти чернетку»/«Опублікувати». **Не** вмикається в `createAdminDependencies`/
  router-контекст (інакше Tiptap впаде в спільний main-chunk, який качає й reader).
  - **B2 (AC3, фокус при reorder):** move-кнопки секцій НЕ використовують `isDisabled` (React Aria
    `isDisabled`→native `disabled`→фокус злітає на `<body>`; пастка `isdisabled-on-focused-button-
    drops-focus`, спіймана в L4a+L7b; jsdom ховає, видно лише в реальному браузері). На межах
    (перша/остання секція) кнопка лишається фокусованою, але no-op (або явне перенесення фокуса на
    переміщену секцію після дії). Сценарій №6 перевіряє це e2e-клавіатурою в реальному браузері.

### Чому існуюче не підійшло
- **media-модуль не переюзуємо для зображень:** домен суто відео (стани T1–T9, renditions, tus,
  ClaimToken); `resolve()` whitelist'ить лише source/hls/thumbnails/temporary. Спільне — том `/media`
  (інфра) і nginx-патерн, не імпорт модуля (ADR 013).
- **`/_media_auth` не переюзуємо:** додає `ownership.Load` (concurrency-lease на відео) і безумовний
  `EmailVerified` (`playback/application/authorize.go:24,31-33`) — для зображень хибно.
- **sqlc-шлях `ListAuthoringLessons` не для body:** jsonb у репо йде raw pgx (asr-прецедент); sqlc
  налаштований лише на course/identity/media.
- **jsonb-колонка на `course_lessons` — ні:** цей рядок джойниться на кожному outline/catalog/
  authoring-завантаженні; окрема таблиця тримає його легким.
- **DnD-ліба — ні у v1:** move-up/down Button-патерн уже доведений для клавіатури/a11y в репо; Aria
  `useDragAndDrop` доступний (react-aria-components 1.21), але pointer-drag — не вимога v1.

## Сайд-ефекти

- **Видалення `course_lessons.description`** — крос-cutting (повний набір; таск робить суцільний
  `grep description` і регенерує sqlc+Orval): доменне поле `Lesson.Description` (`authoring.go:65`),
  `validateModules` (`authoring.go:209`), `MaximumLessonDescription` (`authoring.go:35`), стеля в
  OpenAPI/контракті (→ **Orval regen**), `upsertAuthoringLesson` (`postgres/authoring.go:22-34`),
  **sqlc-генерований селект** (`generated/courses.sql.go:315`) + `ListAuthoringLessons`
  (`courses.sql:132-138`) → **sqlc regen**, маппери (`repository.go`/`authoring.go:443`), admin-HTTP
  DTO, authoring-форма фронта (поле опису уроку). Рішення: **видаляємо** (pre-launch «ріж сміливо»,
  власник 08.10); тіло уроку тепер лише в новому редакторі документа.
- **Політика EmailVerified на learn-document-ендпоінті (прогалина GATE 2 — рішення власника нижче):**
  транскрипт того самого уроку ВИМАГАЄ EmailVerified (`transcript/.../read.go:63`), playback теж
  (`playback/application/authorize.go:31-33`). Щоб сторінка уроку була консистентна, document-read і
  lesson-image-serve мають застосувати ту саму вимогу. Фіксується рішенням власника (див. нижче).
- **Reader учня отримує НОВИЙ артефакт:** сьогодні студенту тіло уроку не віддається взагалі
  (Description бачить лише інструктор в authoring-дереві). Отже новий `/learn/.../document` —
  додавання, не зміна наявної віддачі.
- **nginx:** нова `/media/i/` serve-локація (auth_request→`/internal/image/authorize`) + **нова**
  upload-локація для `POST /api/v1/admin/lessons/{id}/images` з `client_max_body_size 5m` (НЕ переюз
  `/uploads/` tus-16m — то resumable відео; НЕ піднімати глобальний `/api/`=1m `nginx.conf:182`).
  Чіпає `infrastructure/nginx/nginx.conf` + e2e-варіант.
- **Нова залежність Tiptap** (MIT) — виправдана (rich-text); тільки в lazy editor-піддереві.
- **Query-кеш:** нові ключі (`learn.lessonDocument`, admin document/images); publish інвалідовує
  student document query.
- **Композиційні корені:** реєстрація `internal/image` + нових course-ендпоінтів у api-роутері;
  `migrate` виконує 00018.
- **Том `/media`:** новий підтрег `/media/images/` — не конфліктує з source/hls/thumbnails/temporary
  і не ламає `purgeTemporary`/`Free` (перевірити statfs-логіку media).

## План міграції

`00018_lesson_document.sql` (goose Up), у порядку:
1. `CREATE TABLE lesson_documents (lesson_id, locale, draft_body jsonb, published_body jsonb,
   updated_at timestamptz, PRIMARY KEY (lesson_id, locale), FK lesson_id→course_lessons ON DELETE
   CASCADE, CHECK jsonb_typeof(draft_body)='object' AND pg_column_size(draft_body) < <bound>)`
   (дешевий DB-CHECK; глибока валідація — у домені, як asr 00008:52-54).
2. `CREATE TABLE lesson_images (id uuid PK, lesson_id FK→course_lessons ON DELETE CASCADE,
   content_type text, width int, height int, byte_size bigint, created_by, created_at)`.
3. **Backfill (точна форма — ЗАФІКСОВАНО, B1):** для кожного `course_lessons` з непорожнім
   `description` — INSERT рядок `(lesson_id,'uk')`, де і `draft_body`, і `published_body` =
   рівно цей документ (section id детермінований від lesson_id, текст екранований одним абзацом):
   ```json
   {"schemaVersion":1,"locale":"uk","sections":[
     {"id":"<lesson_id>-s1","type":"text",
      "content":{"blocks":[{"type":"paragraph","spans":[{"text":"<description>"}]}]}}]}
   ```
   (форма відповідає json-тегам доменних типів LB01: `Block.spans[].text`, НЕ `paragraph.text` —
   неточність ілюстрації спіймав Codex на LB01). Детермінований `id = lesson_id||'-s1'`.
   Goose-міграції тут `.sql` (не Go — усі 00001–00017 SQL, backfill 00015 теж SQL), тож backfill =
   `jsonb_build_object(...)`. Ризик B1 (SQL-форма ≠ доменна) закриває **integration-тест**: сідить
   урок з непорожнім `description` на downgraded-схемі (DownTo 00017), мігрує Up до 00018, читає
   `lesson_documents`, **розпарсює jsonb у `domain.LessonDocument` і проганяє `ValidateLessonDocument`**
   + звіряє, що текст опису збережений. Якщо SQL-форма розійдеться з доменом — тест червоний.
4. **Дроп `course_lessons.description` — НЕ в 00018, а в окремій міграції 00019 у LB06**, після
   прибирання всіх код-посилань (sqlc-селект, authoring-upsert, домен `Lesson.Description`). Порядок
   слайсів: якщо дропнути колонку в 00018, збірка червона до LB06 (код ще читає колонку). Тож 00018 =
   лише create+backfill (колонка лишається, збірка зелена); 00019 (LB06) дропає після cleanup.
   `seed.sql` уроки НЕ сідить (лише `courses`), тож дроп у 00019 його не ламає; тест сідить власний
   урок. Render-паритет мігрованого документа — e2e в LB11 (сценарій №7 має і e2e-render-рядок).

Goose Down (test-only, disposable): drop обох таблиць, `ADD COLUMN description` назад (порожній).
Прозових описів у сідах немає → backfill майже-порожній, але коректний.

## Rollback

- **Код:** фіча адитивна для учня (новий ендпоінт/компонент); reader падає → урок усе одно має
  відео/транскрипт/дискусію. Editor — новий маршрут, не вплетений у main-chunk.
- **Схема:** pre-launch, прод-БД немає → rollback = goose Down + revert коду; **опубліковані
  документи втрачаються на Down** (прийнятно pre-launch, named limitation). Відкат відновлює
  `description` порожнім.
- **nginx:** revert конфіга (локації адитивні).

## Конституція (проведено за маніфестом)

AGENTS.md архітектура+TS + reuse-index + machine-checked boundaries
(`backend/platform/architecture/boundaries_test.go`, `frontend/scripts/check-architecture.mjs`).

- Домен `course/document.go` і `image/domain` — **лише stdlib + contracts**; валідація дерева чистими
  функціями. ✔
- Application — domain+ports; adapters inward; composition roots wire. ✔
- Крос-модульні правила — лише як чисті предикати в `contracts/access` (image-модуль читає свої факти
  lesson→course і мапить у `access.Course`, не імпортує course/enrollment — ADR 013). ✔
- jsonb-моделі лишаються в адаптерах і мапляться явно в домен-сутності. ✔
- Nginx віддає байти; Go/SSR не проксить (image serve — nginx по груп-читанню, як media). ✔
- TS strict, нема `any`; React Aria; Query для server-state; Router search-params; нема
  заборонених ліб (Tiptap/Tailwind/react-hook-form/react-aria — дозволені й уже в репо). ✔
- Семантичні токени; клавіатура, видимий фокус, ≥44px, 200% zoom, світла/темна — reader і editor. ✔
- Явні бізнес-коди помилок; без секретів у логах. ✔

## Рішення власника (ЗАКРИТО 2026-10-12)

- **Гейт сторінки уроку:** тіло-документ + зображення = **`CanConsume` + EmailVerified**
  (консистентно; власник: «є спека, що лише верифіковані користувачі мають доступ до курсів»).
  Authorize обох (`/learn/lessons/{id}/document`, `/internal/image/authorize`) **дзеркалить гейт
  транскрипта** (`transcript/domain/access.go CanConsumeBy` + EmailVerified), БЕЗ playback-lease
  (lease — суто відео-конкурентність). Прев'ю-уроки видно лише акаунту з верифікованим email, не
  аноніму. Поглинає попереднє питання про публічність прев'ю-зображень.
- **Видалення `Description`** замість shim — підтверджено pre-launch-правилом.

## Приймальні сценарії

Спостережувана поведінка, написана до коду (делеговано `acceptance-spec`, 2026-10-12). Кожен таск
імплементації посилається на номери рядків, які закриває. Три групи представлені.

### Happy path

| # | Сценарій | Передумова | Дія | Очікуваний спостережуваний результат | Чим перевіряється | AC |
|---|---|---|---|---|---|---|
| 1 | Автор будує урок езотерики | Інструктор, збережений урок курсу | Додає `video`, `text`, `row`[image\|text], `task`; «Зберегти чернетку» → «Опублікувати» | Після publish студент-учасник бачить сторінку: плеєр, текст-замовляння, 2 колонки (зображення+текст), блок завдання — у цьому порядку | e2e `lesson-builder.spec` | AC2/AC8 |
| 2 | Учень читає опублікований урок | Опублікований урок з документом, учень з членством | Відкриває `/learn/<slug>/<lessonId>` | Секції рендеряться семантичним HTML (`<section>/<figure>/<img>`), відео грає наявним плеєром | e2e `lesson-builder.spec` | AC4/AC6 |
| 3 | Урок без відео | Інструктор | Будує урок лише `text`+`image`, публікує | Сторінка показує текст і зображення, слота відео немає (0 відео-секцій валідне) | e2e | AC2 |

### Межі

| # | Сценарій | Передумова | Дія | Очікуваний спостережуваний результат | Чим перевіряється | AC |
|---|---|---|---|---|---|---|
| 4 | Порожній документ | Збережений урок без секцій | Студент відкриває | Показано порожній стан (не помилка, не порожня біла пляма) | e2e | AC2 |
| 5 | `row` 1 і 3 колонки | Редактор | Створює ряд на 1, тоді 3 колонки | Обидва валідні, рендеряться responsive-пресетом; 4-та колонка недоступна | unit (домен) + e2e | AC2 |
| 6 | Клавіатурне впорядкування | Редактор з 3 секціями, 200% zoom | Tab до секції, «вгору»/«вниз», видалити — лише клавіатурою | Порядок секцій змінюється видимо; фокус лишається на переміщеній секції; контролі ≥44px | e2e keyboard + axe | AC3 |
| 7 | Міграція плоского Description | Сід урока з **непорожнім `description`** (не з порожнім) | Запуск міграції 00018 | Урок має односекційний `text`-документ (locale uk) у draft=published; документ **проходить доменну валідацію** `LessonDocument` і **рендериться reader'ом** із текстом опису; колонки `description` більше немає | integration (backfill через доменний валідатор на наповнених даних) + e2e render | AC10 |
| 8 | Locale-aware схема | Документ locale uk | INSERT рядка (lessonId,'en') | Обидві локалі співіснують без зміни схеми; reader читає за локаллю; фолбек на uk | unit (домен) + repository test | AC12 |
| 9 | Пласкість DOM | Опублікований урок-езотерика (приклад) | Виміряти DOM-вузли reader-сторінки | Зафіксоване число вузлів; магнітуда проти наївного вкладеного підходу названа у звіті | вручну (1 ручний крок, лишається у звіті) | AC11 |
| 10 | Reader без редакторського JS | Prod-білд | Grep learn-chunk + main-chunk на Tiptap/ProseMirror/toolbar | Жодного редакторського модуля в learn-шляху й main-chunk; Tiptap лише в admin editor async-chunk | bundle-grep (e2e-фаза) | AC4 |

### Відмови

| # | Сценарій | Передумова | Дія | Очікуваний спостережуваний результат | Чим перевіряється | AC |
|---|---|---|---|---|---|---|
| 11 | Невідомий тип секції | — | PUT документа з типом `iframe` | Бекенд відмовляє з явним кодом (`invalid_request`), документ не збережено | contract test | AC1 |
| 12 | Вкладеність глибша за 1 | — | PUT `row`, що містить `row` | Відмова з явним кодом; не збережено | contract test + unit домен | AC2 |
| 13 | XSS у тексті/посиланні | — | Зберегти `text` з `<script>`/`javascript:`-URL; студент відкриває | Рендер безпечний: скрипт не виконується, URL відхилено/знешкоджено; під enforcing CSP (не report-only) | unit (sanitize) + e2e під enforcing CSP | AC5 |
| 14 | Тіло не протікає в outline | Опублікований курс, підписаний користувач БЕЗ членства | GET outline (`/learn/courses/{slug}`) | Відповідь містить назви/структуру, але НЕ тіло документа закритого уроку | contract test | AC7 |
| 15 | Читання body без доступу | Платний урок, користувач без членства | GET `/learn/lessons/{id}/document` | 403 `course_purchase_required` (або еквівалент); тіло не віддано | contract test | AC7 |
| 16 | Прев'ю-урок доступний | Published+preview урок, аноним/не-учасник | GET document + зображення | Тіло й зображення віддаються (дзеркало `CanConsume`) — рішення власника підтверджено | contract test | AC7/AC9 |
| 17 | Чернетка невидима учневі | Урок із draft, що != published | Студент відкриває | Студент бачить лише `published_body`; зміни чернетки не видно до publish | e2e | AC8 |
| 18 | Конфлікт optimistic-lock | Дві вкладки редактора | Друге збереження зі старим updated_at | Відмова з явним кодом конфлікту; перше збереження не перезаписане мовчки | contract test | AC8 |
| 19 | Кривий image-upload | Інструктор | POST не-зображення / >ліміт / decompression-bomb (велетенські пікселі) | Відмова з явним кодом (`unsupported_media_type`/`payload_too_large`/`invalid_request`); файл не збережено | contract test | AC9 |
| 20 | Image serve під auth | Зображення платного уроку, користувач без членства | GET `/media/i/<id>` | nginx 403 (auth_request відмовив); байти не віддано | integration (Compose auth_request) | AC9 |
| 21 | Наявні флоу зелені | — | Ordered `make check` + e2e | playback/transcript/discussion/enrollment e2e зелені; 0 регресій | e2e (повний набір) | AC6/AC10 |

**Ручних рядків: 1** (№9, вимір DOM — лишається у звіті як ручний крок). Решта автоматизовані.

---

## tasks → `work/tasks.md`

Наступний крок після GATE 2. Кожен таск — атомарний, з явними файлами, критерієм «готово», у порядку
залежностей, з посиланням на рядки приймальної таблиці. `[P]` — незалежні. Спайк (spec Q9) —
найперший таск, до коміту схеми.
