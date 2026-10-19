# V-track діагностика — view 3 `#course` (CourseEditorView)

ДАЛІ: прочитати еталон `docs/admin-restructure/_wip-courses-view.md` (view 2) як зразок структури/тону, потім proto `#course`, prod `CourseEditorView.tsx`, спільні примітиви layout.tsx/data.tsx, мапінг беку.

## ФАКТ: proto `#course` композиція (src.jsx case'course' line 29 + CourseForm l38 + VideoForm l39)
PageHeader: eyebrow «COURSE WORKSPACE» + title = **selectedCourse.name** (не статична назва) + description («Відредагуйте деталі та підготуйте курс до публікації»). actions = status-Pill(selectedCourse.status) + secondary «Перегляд»(eye)→navigate('lesson')+setPreview.
Під хедером: `.ui-segmented.course-tabs` з 4 табами: **Overview / Lessons / Video settings / Access**.

- **Overview** → `<CourseForm>` (l38): `.course-form-layout` = ліворуч `form.ui-card.course-details-form`, праворуч `.course-side-panels`.
  - Форма: card-heading «Details»+icon course; error-summary(role=alert); поля в порядку: **name** (Course title, text), **description** (textarea), **price** (· UAH, number); form-footer = save-status + primary submit «Save changes».
  - Side panels (3 картки): (1) **Status** card-heading+status-Pill, опис, primary «Publish». (2) **course-mini-stats**: Lessons / Students / Updated (3 числа). (3) **danger-panel**: eyebrow «DANGER ZONE» + «Delete course» + опис + danger-button.
- **Lessons** → `.ui-card`: card-heading «Lessons»+Pill «MODULE 01»; **хардкод 4 рядки** `.course-lesson-row` (lesson-counter 01.., name + статус Draft/Published, secondary «Edit»→navigate lesson). Один модуль, без drag-handle тут (drag лише в lesson-editor).
- **Video settings** → `<VideoForm>` (l39): `.ui-card.video-settings-form`: card-heading+icon video; Field «Video URL» + hint «metadata only, no streaming backend»; primary «Save changes».
- **Access** → `.ui-card`: card-heading «Access grants»+secondary «Add grant»(plus); список `.grant-row` (email+date, Active-Pill, Revoke icon destructive); Empty(icon users) якщо порожньо.

Proto CSS-класи цього view: admin-page-header, admin-eyebrow, admin-page-actions, ui-segmented, course-tabs, ui-segment, course-form-layout, course-details-form, card-heading, ui-error-summary, form-footer, course-side-panels, course-mini-stats, danger-panel, course-lesson-row, lesson-counter, video-settings-form, grant-row, ui-icon-tile, ui-card, ui-pill, ui-field.

ДАЛІ: grep цих класів у prod CSS; прочитати prod CourseEditorView.tsx + shared layout/data/index.

## ФАКТ: prod `CourseEditorView.tsx` композиція (983 рядки)
**НЕ табований** — один довгий скрол-форм. Структура:
- `<article.admin-editor>`
- Back-Link `/admin` «До курсів».
- `header.admin-editor-heading`: eyebrow «Редактор» + **h1 = course.title** + status-pill (`ui-pill--{warning|success|neutral}`, data-status). **Без description, без Preview-кнопки в хедері.**
- message-region (role status|alert) — результат збереження/помилки (errorMessage мапить 16 AuthoringErrorCode).
- `form.admin-editor-form`:
  - `FormErrorSummary` (role=alert, збирає всі невалідні поля, focus on submit) — як proto error-summary, але багатший (лінки setFocus + динамічні module/lesson поля).
  - `section.admin-editor-details.ui-card`: section-heading (eyebrow «Деталі»+h2 «Деталі курсу»); `admin-editor-grid` поля в порядку: **title, slug** (disabled коли published, hint slugLocked), **summary** (multiline), **description** (multiline), **OwnerField** (admin→Select інструктора / інакше readonly). Порядок полів ≠ proto (proto: name/description/price; prod: title/slug/summary/description/owner, **без price тут**).
  - `DocumentLibrary` (коли transcripts) — бібліотека документів курсу.
  - `ModuleFields` = `section.admin-program.ui-card`: section-heading (eyebrow «Структура»+h2 «Програма»)+moduleCount; `admin-modules` → per-module `fieldset.admin-module` (legend, module-title field, `LessonFields`, row-actions move↑/↓/delete); «Додати модуль» (max 40).
    - `LessonFields` = `div.admin-lessons` → per-lesson `fieldset.admin-lesson` (legend, `admin-editor-grid`: lesson-title + PreviewField Checkbox isPreview; `LessonVideoField` коли video; `LessonDraftField`/`LessonTranscriptField` коли saved; row-actions move↑/↓/delete); «Додати урок» (max 120).
  - `div.admin-editor-save`: submit (Зберегти/Збереження).
- `section.admin-lifecycle.ui-card`: eyebrow «Життєвий цикл»+h2 «Статус курсу»; transcriptsPending-notice (output); row-actions: transition-button (Publish/Archive/Restore за status) + Delete (draft-only, через Dialog-confirm).
- `CoursePriceField` (коли onSetPrice) — окрема картка ціни.
- `{adminTools}` — admin-only слот (напр. access/enrollment), рендериться в кінці.

Також у файлі: `CourseCreateDialog` (Dialog «Новий курс»: title/slug/summary/description/owner + conflict-handling slug_taken) і `CourseEditorPending` (Skeleton 7 рядків — loading-стан).

**React Aria примітиви (через `../ui`):** Button, Checkbox, Dialog, Link, Select, Skeleton, TextArea, TextField. Форма — react-hook-form (useForm/useController/useFieldArray/useWatch), не власний стан.
**Стани:** loading=`CourseEditorPending` Skeleton; error=message role=alert (мапінг кодів); pending=isSubmitting+lifecyclePending (кнопки disabled); empty — немає (курс завжди існує; порожні модулі/уроки = просто порожній масив).

## ФАКТ: ПРОТО МІЛКИЙ, PROD ГЛИБОКИЙ — інверсія view 2
Це НЕ той випадок, що view 2 (там proto показував більше за prod). Тут **proto `#course` = спрощений демо-прототип**, prod = реальний глибокий редактор:
| аспект | proto | prod |
|---|---|---|
| каркас | таби (Overview/Lessons/Video/Access) | один скрол, стек ui-card секцій |
| модулі/уроки | read-only список 4 хардкод-рядків (1 модуль), Edit→навігація | **інлайн CRUD**: field-array модулів+уроків, add/move/delete, max 40/120 |
| відео | один «Video URL» форм-філд (metadata only) | **per-lesson** LessonVideoField (реальний upload/transcode pipeline) |
| транскрипт/драфт | немає | per-lesson LessonTranscriptField + LessonDraftField (ASR) |
| ціна | поле в Overview-формі | окрема CoursePriceField-картка (set/remove) |
| доступи | fake grants-список (session-only) | adminTools-слот (реальний enrollment) |
| owner | немає | OwnerField (admin змінює інструктора) |
| lifecycle | Publish-кнопка в Status-side-panel | повний transition publish/archive/restore + delete |
**Висновок-напрям:** V-track тут = накласти proto-**layout-мову** (PageHeader з eyebrow+description, card-heading-патерн, danger-zone, mini-stats, можливо табовий каркас) на prod, НЕ регресувати глибокий редактор до мілкого proto. Усе, що prod має понад proto, лишається (owner-рішення «full-fidelity з беком» для view2 тут читається навпаки — нема чого дотягувати з беку, бек БАГАТШИЙ за proto).

ДАЛІ: мапінг беку — course.ts типи (CourseTree/AdminCourse), backend/internal/course handler GET single, courses.sql; звірити кожен proto-елемент (mini-stats students, lessons-count, grants) з доменом.

## ФАКТ: мапінг беку (editor payload)
`GetAuthoringCourse` (`courses.sql:120`) — payload редактора: id, slug, title, summary, description, status, instructor_id, created_at, updated_at, published_at, price_minor, currency + `ListAuthoringModules` (id/title/position) + `ListAuthoringLessons` (id/module_id/title/position/is_preview). Домен `CourseTree` = {slug,title,summary,description,instructorId,modules[{id,title,lessons[{id,title,isPreview}]}],price,id,status,createdAt,updatedAt,publishedAt}.
**НЕ містить studentCount** — активний-enrollment count рахується ЛИШЕ у list-запиті `ListAuthoringCourses` (через `course_grants` state='active'); `studentCount`/`lessonCount` живуть на `AdminCourse` (list-item), не на `CourseTree`.
Access-grants (proto Access-tab) = реальний концепт, **вже дротований**: `admin.courses.$courseId.tsx:251` передає `adminTools={<CourseGrantTool courseId=… />}` (enrollment/course_grants).

## Таблиця V-розбіжностей (layout/структура, НЕ палітра)
| # | proto | prod зараз | дія |
|---|---|---|---|
| V-E1 | PageHeader: eyebrow «COURSE WORKSPACE» + title + **description** + actions (status-pill + secondary «Перегляд») | `header.admin-editor-heading`: eyebrow «Редактор» + title + status-pill; **без description, без Preview, не спільний PageHeader** | застосувати спільний `ui/PageHeader`; back-link+status перенести в actions-кластер; Preview-дія — див. B-E6 |
| V-E2 | табовий каркас `.ui-segmented.course-tabs` (Overview/Lessons/Video/Access) | один скрол, стек `ui-card`-секцій | **ОБ'ЄКТ РІШЕННЯ ВЛАСНИКА**: proto-4-таби НЕ покривають prod (video/transcript per-lesson, price, lifecycle, owner). Або (a) лишити скрол + proto-візуал секцій, або (b) згрупувати у таби/якорі БЕЗ втрати можливостей |
| V-E3 | `card-heading` (h2 + icon/pill) у кожній картці | `admin-section-heading` (eyebrow+h2+count) | узгодити вигляд заголовків карток (візуал) |
| V-E4 | окрема **danger-panel** картка (eyebrow «DANGER ZONE» + опис + danger-button) | Delete = danger-кнопка в `admin-lifecycle` row-actions | винести Delete у виражену danger-zone картку |
| V-E5 | **course-mini-stats** (Lessons / Students / Updated) | немає зведення чисел | додати mini-stats; Lessons=derive, Updated=є, Students=B-E1 |
| V-E6 | 2-колонковий `.course-form-layout` (форма ліворуч, side-panels праворуч) | один стовпець | розглянути 2-колонковий layout форма+бічні панелі (структура) |
| V-E7 | Lessons-список: модуль→урок, `lesson-counter` 01.. + Edit→навігація | інлайн field-array CRUD (глибше) | лишити CRUD; накласти proto-візуал (counter, рядок-картка) на module/lesson fieldset — косметика |
| V-E8 | status-Pill + Publish у Status-side-панелі | status-pill у heading + transition у lifecycle-картці | узгодити розміщення/вигляд (lifecycle-картка ≈ proto Status-панель) |

## Таблиця чесних меж беку
| межа | чому відсутнє/стан | рішення V |
|---|---|---|
| B-E1 Students number у mini-stats | `GetAuthoringCourse` НЕ віддає student-count; рахується лише в `ListAuthoringCourses` (course_grants active). **Концепт Є.** | (a) додати active-count subquery у GetAuthoringCourse (малий бек, концепт існує) АБО (b) пропустити Students; не вигадувати — owner-вибір |
| B-E2 Lessons number | не окреме поле, але **derivable** з `CourseTree.modules` | рахувати на фронті з дерева (code-only) |
| B-E3 Updated | `CourseTree.updatedAt` Є | показати напряму |
| B-E4 Access grants (proto Access-tab) | концепт Є, **вже дротований** adminTools=`<CourseGrantTool>` | лишити реальний tool; proto fake-grants НЕ відтворювати |
| B-E5 Video (proto один «Video URL» філд) | prod моделює **per-lesson** (LessonVideoField, реальний upload/transcode) — глибше й інакше | лишити per-lesson; proto single-field = спрощення прототипу |
| B-E6 Preview-дія в хедері | публічна сторінка лише для published | Preview-лінк лише коли published (патерн row-action view2), інакше ховати |
| B-E7 price у Overview-формі (proto) | prod має окрему `CoursePriceField`; перша ціна йде окремо від edit **за дизайном беку** (`course.ts:171-175`) | лишити окрему картку ціни; НЕ зливати в details-форму |

## ВИСНОВОК
**Інверсія view 2: бек-робота майже НЕ потрібна — прод-редактор БАГАТШИЙ за proto.** Єдиний потенційний бек-пункт — B-E1 (student-count у editor-payload), і той опційний (owner-вибір: дотягнути малим subquery чи пропустити). Решта — **чистий layout/візуал**: спільний PageHeader (V-E1), danger-zone (V-E4), mini-stats з derive (V-E5), card-heading-патерн (V-E3), можливо 2-колонковий/табовий каркас (V-E2/V-E6).
**ГОЛОВНИЙ РИЗИК:** механічне «привести до proto» = зрегресувати глибокий редактор (інлайн module/lesson CRUD, per-lesson video+transcript+draft, окрема ціна, owner-picker, повний lifecycle) до мілкого 4-таб демо. **Робити НЕ МОЖНА.** Proto тут — стильовий референс, не функціональна ціль.
**Потрібні рішення власника:** (1) V-E2 каркас — скрол vs таби/секції (не ламаючи per-lesson глибину); (2) B-E1 — student-count: дотягнути чи пропустити; (3) V-E6 — 2-колонковий layout чи лишити стек.

## РІШЕННЯ ВЛАСНИКА 2026-10-17 (закривають альтернативи V-E2/B-E1/V-E6)
Автор: власник (через AskUserQuestion, сесія Courses 1). Перевірка: цей запис + PR view 3.
1. **V-E2 → таби з реальним глибоким вмістом.** `ui-segmented` каркас (Overview/Lessons/Video/Access) як у proto, АЛЕ кожен таб тримає справжній prod-вміст (per-lesson video/transcript/draft, owner-picker, lifecycle, окрема ціна) — НЕ proto-заглушки. Форма react-hook-form не має розмонтовуватись при зміні табу (зберегти field-array стан). Мапінг proto→prod табів фіналізувати у плані (video-settings + price + lifecycle + owner не вкладаються в 4 proto-таби дослівно).
2. **B-E1 → дотягнути student-count** малим active-count subquery у `GetAuthoringCourse` (концепт course_grants active уже є; без міграції).
3. **V-E6 → 2-колонковий Overview** (форма ліворуч, side-panels Status/mini-stats/danger-zone праворуч; стек на вузькому).
Решта меж — як у таблиці: B-E4 grants реальні (не fake), B-E5 video per-lesson, B-E7 price окремою карткою, B-E6 Preview лише published.

## ЖУРНАЛ
- S1 (бек, B-E1) здано `bbe740b`: student-count у GetAuthoringCourse → CourseTree.StudentCount → OpenAPI/Zod/FE-тип. integration Load 2 active+1 revoked→students=2. Гейти: typecheck/FE772/integration/oxfmt/oxlint/vet зелені. Гоча: не плутати форматер — репо = oxfmt, НЕ prettier (npx prettier масово переписав лапки, відкотив).
- ДАЛІ: S2 — спільний ui/PageHeader + ui-segmented таби на CourseEditorView, react-hook-form не розмонтовувати при зміні табу.
- S2+S3 (frontend layout) здано `c297017`: PageHeader + SegmentedControl таби (Overview/Lessons/Access, inactive=hidden, форма не розмонтовується), 2-колонка Overview, mini-stats, danger-zone, card-heading, error-summary flushSync-reveal, порт proto-CSS. FE 772 зелені (6 тестів оновлено на перемикання табів).
- Рішення: таб = локальний view-стан (useState), НЕ URL-param — не драйвить фетч, flushSync-фокус зелений; ціна: reload скидає на Overview. Дата 2026-10-17.
- Крос-зріз (`8019b64`): e2e tab-aware. Playwright getByRole виключає hidden → specs, що дістають Lessons-вміст через getByRole (asr/media/playback/enrollment/transcript), клікають Lessons-таб після goto І після reload; discussion через CSS-локатор — без змін. De-dup status-pill (було двічі → strict-mode конфлікт admin.spec 'В архіві').
- ДАЛІ: повний `make check` (фон, лог scratchpad/makecheck.log) — авторитетний e2e-гейт. Потім браузерна design-pass (side-by-side з proto `#course`) + PR.
