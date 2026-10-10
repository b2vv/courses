# Lens: AI-slop — docs/playback/plan.md (rev 3)

ДАЛІ: grep цитат у плані, звірка почалась

ФАКТ: sql/queries/media.sql:41-45 (правильний шлях backend/sql/queries/media.sql) — рядки 41-45 = FROM courses LEFT JOIN course_modules ... LEFT JOIN course_lessons ... LEFT JOIN video_assets. Сходиться: прямий джойн без порту, прецедент існує.
ФАКТ: identity/adapters/http/config.go — ConfigFromEnvironment (рядок 18) і NewConfig (рядок 26) існують, SessionCookieName = "courseforge_session" на рядку 11. Сходиться.
ФАКТ: identity/adapters/http/handler.go:123 = func Authenticate; :142 = func RequireSameOrigin. Сходиться.
ФАКТ: identity/adapters/http/handler.go:125 викликає identityHeaders(w); :245-247 визначає identityHeaders з Cache-Control: no-store. Сходиться.

ФАКТ: admin-video-queries.ts:69 = refetchInterval: (query) => videoRefetchInterval(...); videoPollMs = 3_000 на рядку 12. Сходиться.
ФАКТ: frontend/src/routes/learn.tsx:24-29 — компонент LearnRoute рендерить EmptyState з Link на /courses (каталог), реальних курсів студента не показує. Відповідає твердженню плану (R24, рядок 522): "маршруту немає; /learn веде в каталог". Сходиться.

ФАКТ: backend/openapi/openapi.yaml:619-624 — ProgramLesson: additionalProperties:false, required:[title], тільки властивість title. Без id. Сходиться.
ФАКТ: infrastructure/nginx/nginx.conf:45-46 = `location = /media { return 403; }` і `location ^~ /media/ { return 403; }`. Чинний стан — безумовний 403, відповідає передумові плану (R13/C2 переписують саме це). Сходиться.
ФАКТ: docker-compose.yml:141 = `- media-data:/media:ro` (том лише для читання). Сходиться.
ФАКТ: docker-compose.yml:152-153 = `read_only: true` / `tmpfs: [/tmp, /var/cache/nginx]` для nginx-сервісу. Сходиться.

ФАКТ: 00003_authoring.sql:35 = `CREATE TABLE course_lessons (` — таблиці "lessons" немає, лише course_lessons. Сходиться.
ФАКТ: 00005_media.sql:6 = `lesson_id uuid REFERENCES course_lessons(id) ON DELETE SET NULL,` — посилається на course_lessons. Сходиться.
ФАКТ: 00005_media.sql:1-2 — коментар "Down drops every media link and is test-only for disposable databases. Retained environments recover forward..." — відповідає твердженню "Down is test-only". Сходиться.
ФАКТ: course/adapters/postgres/authoring.go:237 = виклик deleteMissingAuthoringLessons; SQL (рядок 36-42) видаляє рядки course_lessons, яких нема в поточному наборі збереження курсу — при КОЖНОМУ збереженні курсу. Підтверджує ризик "FK без ON DELETE дав би порушення". Сходиться.

ФАКТ: media/ffmpeg/transcoder.go:198-202 (фактично умова на 199-200) = `if err != nil || duration < 0 || math.IsNaN(duration) || math.IsInf(duration, 0)` — перевіряє <0/NaN/Inf, але НЕ <=0, тобто нуль справді приймається. Сходиться (незначний зсув нумерації рядків, суть правильна).
ФАКТ: media/domain/source.go:40-46 = func AcceptSource — перевіряє Width/Height/videoAccepted/audioAccepted, тривалість (DurationSeconds) не перевіряє взагалі. Сходиться.
ФАКТ: course/domain/authoring.go:23 = `MaximumLessons = 120` (використовується на рядку 184 `if lessonCount > MaximumLessons`). Сходиться.

ФАКТ: infrastructure/tests/integration.sh:26-34 = docker run монтує лише `-v "$task_root/backend:/work"` (плюс docker socket, go-кеші) — nginx справді відсутній у цьому контейнері. Сходиться.
ФАКТ: Makefile:54 = `check: format-check lint typecheck test integration check-generated build up seed-identities seed smoke e2e` — smoke дійсно стоїть перед e2e. Сходиться.
ФАКТ: Makefile:52 = тіло check-generated (`git diff --exit-code -- ... generated ...`), сама ціль оголошена на рядку 51. Незначний зсув нумерації (плюс-мінус рядок), суть (check-generated звіряє згенерований код) правильна. Сходиться.

ФАКТ: backend/platform/architecture/boundaries_test.go існує (1326 байт). Сходиться.
ФАКТ: frontend/scripts/check-architecture.mjs існує (2857 байт). Сходиться.
ФАКТ: frontend/e2e/accessibility.spec.ts існує (7332 байт). Сходиться.
ФАКТ: frontend/rstest.config.ts — обидва тестові проєкти ('adapter-node', 'dom') НЕ використовують jsdom; 'dom'-проєкт має testEnvironment: 'happy-dom'. package.json підключає лише happy-dom (^20.13.2), jsdom — не пряма залежність (лише транзитивний peerDep у package-lock, не використовується). Але план (рядок 486, B25) пише "**e2e accessibility.spec.ts** (jsdom цього не спостерігає)" — реальний рушій unit-тестів **happy-dom**, не jsdom. Це єдине посилання на jsdom у плані і воно неправильне.
ФАКТ: docs/adr/004-local-media-storage.md рядок 4 містить дослівно "API and SSR get no media mount" — цитата плану (рядок 407) точна.
ФАКТ: work/PRODUCT-BRIEF.md §2 п.13 = "Відеобайти не повинні проксуватися через Go API або frontend SSR server."; п.14 = "Nginx напряму віддає HLS playlists і segments після перевірки доступу." Відповідає твердженню плану (рядок 25). Сходиться.
ФАКТ: PRODUCT-BRIEF §21 Phase 6 = signed access; Nginx validation; hls.js; progress; session limits. §24.14 = кодек H.264/AAC зараз, AV1 за тригером (не в Phase 6, лише проба). §24.17 = телеметрія AV1-декодера будується в Phase 6 через navigator.mediaCapabilities/MediaSource.isTypeSupported — план (D-7, рядки 322-347) реалізує саме це, без додаткового каналу. §24.18 = мультимовне аудіо НЕ в Phase 6 — план про другу мову/доріжки нічого не додає (grep підтверджує відсутність scope creep). План НЕ цитує ці підпункти дослівно (жодного "§21"/"§24.1x" рядка немає в тексті плану), але сама поведінка плану їм не суперечить. Сходиться по суті, хоч і без явного покликання.

## Підсумок

Перевірено 24 цитати/групи цитат. 23 сходяться дослівно (з точністю до ±1 рядка нумерації
там, де ціль — тіло/сусідній рядок функції). Одна розбіжність:

[slop] Плутанина jsdom/happy-dom у B25
де:      docs/playback/plan.md:486 (рядок таблиці приймальних сценаріїв B25)
чому:    план пише "e2e accessibility.spec.ts (jsdom цього не спостерігає)", але
         frontend/rstest.config.ts (обидва проєкти) і frontend/package.json підключають
         happy-dom, не jsdom — jsdom у репо взагалі не пряма залежність. Назва рушія
         неправильна; сам аргумент (потрібен e2e, бо unit-DOM не рендерить реальний layout/
         zoom/contrast) лишається вірним і для happy-dom, тож висновок рядка не міняється.
що зробити: замінити "jsdom" на "happy-dom" в рядку B25.
вага:    non-blocking
