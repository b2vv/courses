model: claude-opus-4-8

ДАЛІ: читаю docs/accounts/plan.md (§4c, §4d, Acc25–Acc31, T11), docs/accounts/spec.md (AC14), потім звіряю CSP-шлях у frontend/server.mjs, routes/__root.tsx, nginx.conf, identity код.

ФАКТ (читання): spec AC14 (spec.md:64-66) і plan §4d/§4c/§6 звірені з кодом.
- B1 surface: server.mjs:18-19 відмовляє /api і /media (404) → HTML-документ віддає лише frontend SSR; nginx.conf:116-152 шле /api/ на Go (JSON), 153-166 / на frontend:3000. Отже скрипто-несуча поверхня = TanStack <Scripts/> у __root.tsx:85, рендерена Node SSR. Go = лише JSON. Заяву про Go-middleware як CSP-поверхню знято коректно.
- Nonce-хук реальний: TanStack Scripts.js читає router.options.ssr?.nonce і ставить nonce на КОЖЕН script (inline serverBufferedScript/ScriptOnce + manifest). ScriptOnce.js теж. Отже механізм ssr.nonce вбудований, не фантазія. Threading із server.mjs у getRouter() — не показаний (getRouter() без аргументів, request-scope не продемонстрований), але fallback (report-only+замір) задокументований. B1 surface = ЗАКРИТО.
- N1 (§4c:119-123): адмін-disable-2FA → структурований лог-івент; audit-таблиця названа поза скоупом (механізм є). ЗАКРИТО.
- N2 (§4c:116-118, §5:153): атомарний consume recovery_codes UPDATE…WHERE consumed_at IS NULL RETURNING id; регенерація видаляє всі попередні. ЗАКРИТО.
- N3 (Acc31:210 + §4c:118): регенерація робить старі недійсними. ЗАКРИТО.
- N4 (§4c:110-113, Acc30:209): pending одноразовий, видаляється при успіху/вичерпанні, несе remember-me. ЗАКРИТО.
- N5 (§4c:114-115): 5 спроб на pending. ЗАКРИТО.

ФАКТ (нові діри — CSP-коректність на тепер-правильній поверхні):
- NB1 (style-src): запропонований заголовок (§4d:134) default-src 'self' без style-src, і §6:166 заявляє абсолют «CSP без unsafe-inline». Але код має inline style-атрибути: presentation/ui/data.tsx:17 і :42 (ProgressBar/Meter width), learn/ReviewBlock.tsx:137, storefront/CourseView.tsx:230, плюс React Aria оверлеї (overlays.tsx, LessonPlayer) ставлять inline-позиціонування у рантаймі. style-атрибути підпадають під style-src-attr→style-src→default-src 'self' → блокуються й репортяться у консоль. Acc25 «e2e-консоль без CSP-порушень» впаде; LessonPlayer/ProgressBar активно ганяються e2e. «Без unsafe-inline» як абсолют спростовний: для script-src досяжний (nonce), для style-src — ні без style-src 'unsafe-inline'. План цього не узгодив.
- NB2 (nginx подвійний CSP): §4d:137 «add_header … always, успадковується в location /». location / (nginx.conf:153) не має власних add_header → успадкує server-level CSP. server.mjs уже ставить свій nonce-CSP на цю ж SSR-відповідь (проксовану через location /). nginx add_header ДОДАЄ другий Content-Security-Policy; два CSP-заголовки застосовуються за AND. Базовий nginx script-src (без per-request nonce — план сам каже nginx nonce вставити не може) заблокує nonce'd inline hydration-скрипти (Scripts.js serverBufferedScript/ScriptOnce — inline). Наслідок: зламана гідрація + порушення. Фікс: не вішати nginx-CSP на location /, лише на api/media/error-шляхи.

РІШЕННЯ: B1 ЗАКРИТО; N1–N5 ЗАКРИТО; але знайдено 2 нові blocking CSP-дефекти (NB1 style-src, NB2 nginx подвійний заголовок), кожен сам по собі робить Acc25 червоним як написано. Вердикт: ЧЕРВОНИЙ.
