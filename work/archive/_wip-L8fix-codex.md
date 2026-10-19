ДАЛІ:

- Крок 1: прочитано бриф, робочу модель, briefing і план enrollment та tasks (лише читання). Межі: frontend + цей WIP; перевірки тільки typecheck, lint, npm test; e2e/Docker — Courses 1.

- Крок 2: звірено En03/04/05 та SPA-навігацію з кодом; контракти потрібних операцій існують.

## Discussion

- 2026-10-07, Codex: бриф називає CourseView формою відгуку. Насправді запис виконує `presentation/learn/ReviewBlock.tsx:119`; CourseView:295 містить ReviewCard зі скаргою. Виправляю обидва фактичні UI-шляхи, тестую видимі повідомлення форми та скарги.
- Контраст походить не з текстів ParticipantReportQueue, а зі спільного `.admin-nav-link[aria-current='page']` (`main.css:1684`). Збережений En07 trace показує активний пункт «Скарги учасників». Світла палітра: #0b6b78 на 16% цієї ж фарби поверх #eef2f4 = RGB(201.68,220.4,224.16), WCAG sRGB contrast 4.3853:1. Темна: 5.7525:1. Пропозиція/реалізація: foreground `--color-text`, фон зберегти; відповідно 12.7692:1 та 11.0340:1. Метод: лінеаризація sRGB, L=.2126R+.7152G+.0722B, (Lmax+.05)/(Lmin+.05). Browser axe після зміни залишається Courses 1.

- Крок 3: виправлено e2e асерти, активний admin-nav contrast, доменні error unions, 403-мапінги коментарів/відгуків і uk/en повідомлення. ReviewBlock та CourseView показують named відмови. Далі — message-level тести та дозволені перевірки.

- Додаткова неточність: `routes/learn.$.tsx:200–208,307–309` поглинає помилку report у runDiscussionMutation, тому CommentCard після 403 показує «Скаргу надіслано». Для publish named error уже надходить prop-ом; для report виправляю передачу rejected Promise до CommentCard і його catch на named повідомлення. Це зберігає transport у адаптерах та прибирає хибний success.

- Крок 4: додано 8 message-level unit сценаріїв: обидва коди для DiscussionPanel error, comment-report rejection (без хибного success), ReviewBlock submit та CourseView report. Починаю typecheck → lint → npm test.

- Крок 5: `cd frontend && npm run typecheck` — PASS (exit 0).

- Крок 6: `cd frontend && npm run lint` — PASS (exit 0), architecture boundaries 174 files.

- Крок 7: `cd frontend && npm test` — NOT RUN / BLOCKED перед виконанням тестів (exit 1): Rstest не може відкрити `::1:3000`, `listen EPERM: operation not permitted`. Лог: `/tmp/l8fix-npm-test.log`. Sandbox approval policy=never; підвищення прав недоступне. Жодного тесту не заявляю пройденим. Повтор npm test потрібен Courses 1 у дозволеному середовищі.

- Крок 8: відформатовано лише 14 змінених source-файлів через локальний oxfmt. e2e assertions лишено точковими, без переформатування специфікацій. Повторюю typecheck/lint після форматування.

- Крок 9: після форматування typecheck + lint PASS (174 architecture files). Фінальне читання виявило зайве розширення AuthoringErrorCode від механічної заміни — прибрано, нові коди лишаються тільки в потрібному CatalogErrorCode/DiscussionErrorCode. Повторюю дозволені статичні перевірки після цього звуження.

- Крок 10 / передача: фінальні typecheck і lint PASS, exit 0; architecture 174 files. Усі три частини реалізовано. 8 нових UI-тестів написано, але npm test BLOCKED sandbox EPERM до старту тестів. E2E/Docker не запускались; повний браузерний доказ контрасту та e2e сценарії — Courses 1 після reseed. Git не викликався; backend/generated/work/tasks.md/docs report не змінювались. Наступне: Courses 1 — npm test у середовищі з дозволеним listen, потім reseed + full make e2e; перевірити Discussion-відхилення вище. Контраст наведено як розрахунок CSS, не як новий axe-pass.
