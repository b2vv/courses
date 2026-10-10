
## Лінза «надлишкова складність» — старт
ДАЛІ: читаю docs/playback/plan.md, spec.md, _wip-phase6.md, AGENTS.md, .claude/standards.md; далі перевіряю по коду §2 плану (реальні файли/рядки), platform/redis, потребу в 3 портах, дублювання модуля playback з media/course, hls.js.

ФАКТ: план прочитано (324 рядки). Ключове: §1 новий модуль `playback`; §2 таблиця reuse з конкретними номерами рядків; §3 D-1..D-7; §7 три порти; §8 P01-P13/B01-B19/R01-R23 (55 рядків).
ДАЛІ: перевіряю по коду — platform/redis, identity handler.go:123/:142, transport.ts, admin-video-queries.ts:69, ffmpeg/transcoder.go:150-152, nginx.conf:45-46, openapi.yaml:619-624.

ФАКТ: всі номери рядків §2 підтверджені по коду — handler.go:123 Authenticate, :142 RequireSameOrigin, transcoder.go:150-152 (master_pl_name/var_stream_map/index.m3u8), nginx.conf:45-46 (return 403), openapi.yaml:619-624 (ProgramLesson {title}), admin-video-queries.ts:69 refetchInterval, transport.ts createRequestControl. §2 цитати чесні.
ФАКТ: platform/redis існує: client.go Open()->*go-redis client + QueueOption(asynq); queue.go логер+MonitorQueue. runtime/dependencies.go уже кладе Redis у Dependencies, і cmd/api/main.go:37 його відкриває. Тобто api ВЖЕ має go-redis клієнт — «нове» тут лише тонкий адаптер SET/GET.
ФАКТ: прецедент крос-модульного читання — media НЕ має порту до course: sql/queries/media.sql:41-45 сам джойнить courses/course_modules/course_lessons через власний AssetRepository. Отже «порт читання курсу» суперечить прецеденту.
ДАЛІ: перевіряю чи identity.Authenticate дає session id (для гіпотези «playback-кука зайва»), і чи є frontend/src/domain|ports як усталений патерн.

ФАКТ: identity/domain/session.go:5 `Session{ID,UserID,...}` існує; але `Identity.Authenticate` (application/identity.go:71) повертає лише `domain.User` — `SessionStore.Resolve` теж. Отже per-device id доступний або мінімальною зміною identity, або як sha256(сесійного токена) у самому playback — без змін identity взагалі.
ФАКТ: _wip-phase6.md:41 (обґрунтування G7): «особу несе сесійна кука ... A4 і сценарій 2.4 стають безкоштовними (кука вже ідентифікує)». D-1 вводить ДРУГУ куку з власним 3-хв TTL і окремим ендпоїнтом продовження — тобто повертає ту саму вартість, яку G7 оголосив нульовою.
ФАКТ: frontend/package.json не містить hls — залежність справді нова. AGENTS.md список заборон (Vite/Next/Vitest/Jest/ESLint/Axios/RTK/Radix/shadcn) hls.js не містить.
ФАКТ: frontend/src/ports — плоскі файли (*Repository.ts), не теки; frontend/src/domain — теки на модуль.
РІШЕННЯ: 8 знахідок, 2 blocking (друга кука; розщеплене володіння читальною моделлю).
