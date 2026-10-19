# admin11 recon — збір фактів про /admin-сторінки

ДАЛІ: прочитати 5 route-файлів у frontend/src/routes/, знайти в'юшки, Query-хуки, Orval-клієнти; далі — беки-хендлери й маршрути з "/admin".

## ФАКТ: route files + views прочитано
- /admin → routes/admin.index.tsx:21-67 → CourseListView (presentation/admin/CourseListView.tsx)
- /admin/users → routes/admin.users.index.tsx:14 → UserListView (presentation/admin/UserListView.tsx)
- /admin/learners → routes/admin.learners.tsx:16 → LearnerOverview (presentation/teach/LearnerOverview.tsx)
- /admin/participant-reports → routes/admin.participant-reports.tsx:6 → ParticipantReportQueue (presentation/admin/ParticipantReportQueue.tsx)
- /admin/courses/$courseId → routes/admin.courses.$courseId.tsx:54 → CourseEditorView (presentation/admin/CourseEditorView.tsx)
- /admin layout → routes/admin.tsx (auth gate only, Outlet). NB: Outlet не має видимого AdminShell у admin.tsx — перевірити __root.tsx + AdminShell.tsx

ДАЛІ: AdminShell.tsx (navItems), __root.tsx; domain типи (AdminCourse, AdminUser, LearnerRow, ParticipantReportQueueItem); композиційні query-файли (admin.ts, admin-queries, teach-queries, enrollment-queries); бек-маршрути /admin.

## 6. AdminShell navItems (routes/__root.tsx:145-157; shell у presentation/admin/AdminShell.tsx)
Рендериться В __root.tsx (НЕ в admin.tsx), бо /admin і /teach — ssr:false client-only; shell у корені дає хром у server HTML. Сайдбар = aside.admin-sidebar з brand "</> CourseForge", nav-лінки, лінк "to client". Topbar: title, user.displayName, LanguageSwitcher, SignOut. Жодних лічильників/badge у навігації.

Admin-варіант (onTeachPath=false):
- /admin — navCourses (exact)
- /admin/participant-reports — navParticipantReports
- /admin/learners — navLearners
- /admin/users — adminUsers.title (ЛИШЕ якщо canAdministerAccounts(viewer))
Teach-варіант (onTeachPath=true):
- /teach — navModeration (exact)
- /teach/learners — navLearners

## ФАКТ: DTO-поля (domain)
AdminCourse (course.ts:109): id, slug, title, status('draft'|'published'|'archived'), instructorId, updatedAt. Page: items[] + pagination{limit,offset,total}.
AdminUser (identity/administration.ts:21): id, email, displayName, role('student'|'instructor'|'admin'), status('active'|'blocked'|'deleted'), updatedAt. Page: items[]+pagination{limit,offset,total}.
LearnerRow (teach/teach.ts:82): userId, displayName, lastActivityAt(string|null), courses[]. LearnerCourse: courseId, slug, title, state('active'|'blocked'), progressPercent(number 0-100), enrolledAt. Page: entries[] + nextCursor(string|null) [cursor-пагінація].
ParticipantReportQueueItem (enrollment/enrollment.ts:35): id, courseId, reporterId, targetId, kind('on_lecturer'|'on_student'), reason, state('open'|'resolved'), createdAt, reporterName(|null), targetName(|null), courseTitle. Page: items[] + nextCursor(string|null).
CourseTree (course.ts:163): = CourseTreeEdit{slug,title,summary,description,instructorId,modules[],price?} + price(CoursePrice|null), id, status, createdAt, updatedAt, publishedAt(string|null). AuthoringModule{id,title,lessons[]}; AuthoringLesson{id,title,description,isPreview}. CoursePrice{amount,currency? — звірити}.
Video rows (CourseEditorVideo): admin.videos.list(course.id) → звірити поля у admin-video-queries.

## ПОСЕКЦІЙНО A/B/C/D

### 1. /admin (course list) — routes/admin.index.tsx
A. section.admin-course-list → heading (eyebrow+h1 + admin-list-count "N курсів" + createAction кнопка) → Tabs (all/draft/published/archived) → Table[title(link),slug(code),status(пігулка data-status),owner(instructorId code),updated(time)] + Pagination (offset-based).
B. useSuspenseQuery admin.list(adminCourseQuery) → AdminCoursePage{items:AdminCourse[], pagination{limit,offset,total}}. AdminCourse: id,slug,title,status,instructorId,updatedAt. + admin.instructors() → InstructorOption{id,displayName,role}. Мутації: create.
C. GET /api/v1/admin/courses (course/adapters/http/admin_handler.go:128) → adminCoursePageDTO{items[adminCourseDTO],pagination}. POST .../courses(create). GET /api/v1/admin/instructors.
D. Агрегат: ЛИШЕ pagination.total (к-сть курсів за фільтром статусу). Status дає розподіл draft/published/archived ТІЛЬКИ як окремі запити (кожен таб = окремий total). Часових рядів нема. Для hero-метрики: total курсів; для бар-чарту "курси за статусом" — треба 3-4 окремі list-запити (по 1 на статус), агрегату в одній відповіді НЕМА.

### 2. /admin/users — routes/admin.users.index.tsx (гейт canAdministerAccounts)
A. section.admin-user-list → heading(eyebrow+h1 + admin-list-count "N") → UserSearch(form) → message(role status/alert) → Tabs(all/active/blocked/deleted) → Table[email,name,role,status(пігулка data-status),actions(block/restore,delete,reset-password dialog)] + Pagination.
B. useSuspenseQuery admin.users.list(AdminUserQuery{q,limit,offset,status}) → AdminUserPage{items:AdminUser[],pagination{limit,offset,total}}. AdminUser: id,email,displayName,role('student'|'instructor'|'admin'),status('active'|'blocked'|'deleted'),updatedAt. Мутації: setStatus,delete,resetPassword.
C. GET /api/v1/admin/users (identity/adapters/http/admin_handler.go, реєстрація handler.go:124) → adminUserPageDTO{items[adminUserDTO],pagination{limit,offset,total:int64}}. PUT .../users/{id}/status, DELETE .../users/{id}, PUT .../users/{id}/password, POST .../users/{id}/2fa/disable.
D. Агрегат: pagination.total (к-сть юзерів за фільтром). Розподіл за status (active/blocked/deleted) і за role — ЛИШЕ окремими запитами, в одній відповіді агрегату НЕМА. Часових рядів нема.

### 3. /admin/learners — routes/admin.learners.tsx → LearnerOverview (presentation/teach/)
A. section.teach-section → h1 → search-фільтри(q input, Select state all/active/blocked, submit, clearCourse) → ul.learners-list, кожен learner = article{h2 displayName, lastActivity time|neverActive, ul.learners-courses: per-course Link + state-badge + ProgressBar(progressPercent 0-100)} → "more" кнопка (infinite).
B. useInfiniteQuery teach.learners('admin', filters) → LearnerPage{entries:LearnerRow[], nextCursor:string|null}. LearnerRow: userId, displayName, lastActivityAt(string|null), courses:LearnerCourse[]. LearnerCourse: courseId,slug,title,state('active'|'blocked'),progressPercent(number),enrolledAt. Filters: courseId?,state?,q?.
C. GET /api/v1/admin/learners (orders/adapters/http/handler.go:54, handler.adminLearners) → learnerPageDTO{entries[learnerRowDTO{userId,displayName,lastActivityAt*,courses[learnerCourseDTO{courseId,slug,title,state,progressPercent:int,enrolledAt}]}], nextCursor*}. (teach-варіант: GET /api/v1/teach/learners:59).
D. Агрегат: progressPercent на кожен (learner×course) — ГОТОВИЙ 0-100 для метра/бар-чарту/розподілу прогресу. lastActivityAt → можна зібрати timeline/"active vs never". АЛЕ: cursor-пагінація, total НЕМА — повний підрахунок вимагає прокрутити всі сторінки. Готового агрегату (середній прогрес, к-сть активних) у відповіді НЕМА — лише рядки з progressPercent.

### 4. /admin/participant-reports — routes/admin.participant-reports.tsx → ParticipantReportQueue
A. section.admin-user-list → heading(eyebrow+h1) → loading/error/actionError → EmptyState | ul(teach-section): per-report article{h2 kind, reporter, target, course, reason, state, createdAt time, actions: Resolve, Revoke(dialog, лише on_student)} → "more"(infinite).
B. useInfiniteQuery enrollment.participantReports() → ParticipantReportQueuePage{items:ParticipantReportQueueItem[], nextCursor:string|null}. Item: id,courseId,reporterId,targetId,kind('on_lecturer'|'on_student'),reason,state('open'|'resolved'),createdAt,reporterName(|null),targetName(|null),courseTitle. Мутації: resolveReport, revoke.
C. GET /api/v1/admin/participant-reports (discussion/adapters/http/handler.go, test:372) → page{items,nextCursor}. POST .../participant-reports/{reportId}/resolve (handler.go:49). POST /api/v1/admin/courses/{courseId}/participants/{userId}/revoke (orders handler.go:53).
D. Агрегат: cursor-пагінація, total НЕМА. kind + state дають категорії для розподілу (on_lecturer/on_student × open/resolved) АЛЕ лише порахувавши рядки клієнтом. createdAt → timeline. Готового лічильника/рейту НЕМА — лише список рядків.

### 5. /admin/courses/$courseId — routes/admin.courses.$courseId.tsx → CourseEditorView
A. Редактор курсу (не дашборд): форма course tree (slug/title/summary/description/instructor select), price set, status transition (draft→published→archived), delete, модулі+уроки (drag), per-lesson video upload/resume/retry/cancel (LessonVideoField), транскрипти (per-lesson, documents library), ASR drafts (regenerate), admin-only: CourseGrantPanel (видати доступ юзеру за пошуком).
B. useSuspenseQuery admin.detail(courseId) → CourseTree{slug,title,summary,description,instructorId,modules[AuthoringModule{id,title,lessons[AuthoringLesson{id,title,description,isPreview}]}],price(CoursePrice|null),id,status,createdAt,updatedAt,publishedAt(|null)}. useQuery admin.videos.list(courseId) → videoDTO[]: lessonId,videoId,state,generation,originalName,receivedBytes,declaredBytes,updatedAt,failureReason. transcripts.documents, drafts per lesson, instructors(). Мутації: save,setPrice,transition,delete,videos.*,grants.grant,drafts.regenerate.
C. GET /api/v1/admin/courses/{id} (course/admin_handler.go:129) → adminCourseTreeDTO. PUT .../{id}(save), DELETE .../{id}, PUT .../{id}/status. priceDTO{minor:int64,currency}. GET /api/v1/admin/courses/{courseId}/videos (media handler.go:72) → courseVideosDTO{items[videoDTO]}. POST /api/v1/admin/lessons/{lessonId}/video, POST /api/v1/admin/videos/{videoId}/retry. Транскрипти: /api/v1/admin/courses/{courseId}/documents (GET/POST), DELETE .../documents/{documentId}, GET/PUT /api/v1/admin/lessons/{lessonId}/transcript, GET .../draft, POST .../draft/regenerate, POST .../draft/publish. Grants: POST /api/v1/admin/courses/{courseId}/grants.
D. Агрегат: на рівні курсу — к-сть модулів/уроків (можна порахувати з tree), video state на урок (processing/ready/failed = розподіл для health-таблиці), receivedBytes/declaredBytes (= progress% аплоаду/транскоду, inline-метр), price. Готового числа НЕМА — рахується з tree+videos клієнтом. Це edit-екран, не метрики.

## 7. ВСІ admin HTTP-ендпоінти (Go, реєстрація)
course (admin_handler.go:128-132): GET/POST /api/v1/admin/courses; GET/PUT/DELETE /api/v1/admin/courses/{id}; PUT /api/v1/admin/courses/{id}/status; GET /api/v1/admin/instructors (3106 у generated).
identity (handler.go:124-128): GET /api/v1/admin/users; DELETE .../users/{id}; PUT .../users/{id}/status; PUT .../users/{id}/password; POST .../users/{id}/2fa/disable.
orders (handler.go:52-54): POST /api/v1/admin/courses/{courseId}/grants; POST /api/v1/admin/courses/{courseId}/participants/{userId}/revoke; GET /api/v1/admin/learners.
discussion (handler.go:49): GET /api/v1/admin/participant-reports; POST .../participant-reports/{reportId}/resolve.
media (handler.go:72): POST /api/v1/admin/lessons/{lessonId}/video; GET /api/v1/admin/courses/{courseId}/videos; POST /api/v1/admin/videos/{videoId}/retry.
transcript (admin_handler.go:37-41): GET/POST /api/v1/admin/courses/{courseId}/documents; DELETE .../documents/{documentId}; GET/PUT /api/v1/admin/lessons/{lessonId}/transcript; GET /api/v1/admin/lessons/{lessonId}/draft; POST .../draft/regenerate; POST .../draft/publish.
NB: teach-дзеркала (/api/v1/teach/learners, /roster, /participants/.../block|unblock, /comments, /reviews) — для /teach surface, не /admin.

## ВИСНОВОК по агрегатах (D, зведено)
- НЕМАЄ жодного stats/metrics/dashboard/summary/analytics ендпоінта на беку (перевірено grep по всьому backend).
- Готові числа в одній відповіді: pagination.total (курси, юзери — offset-based); progressPercent per learner×course (0-100, готовий метр); video receivedBytes/declaredBytes (готовий метр аплоаду).
- Розподіли (за статусом курсу, роллю/статусом юзера, kind/state репорту, станом enrollment) — існують як ПОЛЯ, але агрегуються ЛИШЕ клієнтом: або окремими list-запитами на фільтр (courses/users мають total на фільтр), або прокруткою cursor-сторінок (learners, participant-reports — total взагалі НЕМА).
- Часові ряди: є лише createdAt/updatedAt/lastActivityAt/enrolledAt на рядках; готового timeline/букетів НЕМА.
- Отже hero-метрика "N курсів / N юзерів" реальна (один total). Будь-який бар-чарт за категорією = або N окремих запитів (courses/users), або клієнт-агрегація неповної (cursor) вибірки (learners/reports). Донат-summary прогресу learners можливий лише по завантажених сторінках.
