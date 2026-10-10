
=== ЛІНЗА edge cases — старт (сесія) ===
ДАЛІ: читаю docs/playback/plan.md §8 (55 рядків) і docs/playback/spec.md (A1–A37), будую матрицю покриття.
Далі перевіряю існування інструментів: infrastructure/tests/smoke.sh, nginx.conf, міграції, backend/internal/media/.

ФАКТ: план прочитано (§8 = 13 P + 19 B + 23 R = 55 рядків). Спека A1–A37 прочитана.
ФАКТ: `infrastructure/tests/smoke.sh` ІСНУЄ (B10 посилається на реальний файл). `frontend/e2e/` не має `playback.spec.ts` — план його створює (не сказано явно, але зрозуміло). `recovery.spec.ts` і `route-guards.spec.ts` існують.
ФАКТ: nginx.conf `log_format safe` НЕ містить URI взагалі (лише requestId/method/status/duration), `error_log /dev/null`. → R16 «URL записані без query» у nginx-логах неспостережуване (там URL немає взагалі) — рядок проходить тавтологічно.
ФАКТ (виміряно docker): у nginx-контейнері uid=101, `/var/cache/nginx` = tmpfs root:root drwxrwxr-x → NOTWRITABLE; `/tmp` = drwxrwxrwt. Пастка D-2 РЕАЛЬНА. Але compose вже монтує tmpfs `/var/cache/nginx` (docker-compose.yml:153) — план цього не згадує.
ФАКТ: `/media` том містить не лише `hls/`, а й `source/`, `thumbnails/`, `temporary/` (domain/asset.go:152-155). `location ^~ /media/ { root /; }` з D-2 віддає ВСЕ це.
ФАКТ: міграції — стелі «40 модулів × 120 уроків» у схемі НЕМАЄ. course_modules.position CHECK (position >= 0), course_lessons.position CHECK (position >= 0); жодного UNIQUE/ліміту кількості (00003_authoring.sql:27,40).
ДАЛІ: логування Go (R16/R22), метрики (B11/P12), Redis у compose (R12), проба матриці A28/A34.

ФАКТ (матриця): усі 55 рядків §8 розібрано за колонкою «Критерії». Кожен A має рядок, КРІМ **A28** і **A34** — вони не згадані ЖОДНИМ рядком.
ФАКТ: `make integration` = `infrastructure/tests/integration.sh` → `docker run ... -v backend:/work golang go test -tags=integration ./...`. Монтується ЛИШЕ `backend/`, testcontainers-Postgres. **nginx у цьому харнесі немає.** Отже 16 рядків «integration», що стверджують поведінку edge (R01–R04, R06–R10, R13, R15, R16, R22, B11, B16, B17), названим інструментом неспостережувані. Phase 5 клала edge-твердження в smoke.sh.
ФАКТ: `recovery` — це Playwright-проєкт (playwright.config.ts, `frontend/e2e/recovery.spec.ts`), не Go-integration. R11/R12 підписані «integration `recovery`».
ФАКТ: `platform/observability/metrics.go` має рівно 6 колекторів; серії проби AV1 нема, browser→Go інжесту метрик нема ніде, `/metrics` крізь nginx = 404. P12 неспостережуваний.
ФАКТ: env-змінної TTL немає ні в `.env.example`, ні в docker-compose.yml. `make check` = `up → smoke → e2e` над ОДНИМ стеком. P08 «прискорений TTL у тестовому конфізі» ні на що не спирається.
ФАКТ: логи. nginx `log_format safe` без URI, `error_log /dev/null`; Go логер (`httpserver/server.go:124`) пише method/route/status/durationMs/requestId/traceId. R16 і R22 проходять тавтологічно.
ФАКТ: `video_assets` (00005_media.sql) **НЕ має колонки тривалості**. Бекенд ніде не знає duration. A24/B05 (95% у `progress_test.go`) і D-5 («обрізається до тривалості поточної генерації на читанні») спираються на дані, яких нема, а §5 обіцяє «наявні таблиці не змінюються».
ФАКТ: §5 називає FK на `lessons(id)`; таблиця в схемі — `course_lessons` (00003_authoring.sql:36).
ФАКТ: B15 «40×120 (стеля схеми)» — стеля доменна (`authoring.go:22-23`), не схемна; і `MaximumLessons=120` рахується КУМУЛЯТИВНО по всіх модулях (`authoring.go:184`). 40×120=4800 уроків продукт створити не може.
ФАКТ: D-2 не містить `proxy_cache_key`. Дефолтний ключ = `$scheme$proxy_host$request_uri` підзапиту, а підзапит `auth_request` завжди `/_media_auth` → один ключ на всіх користувачів і всі відео.
ФАКТ: `/media` том містить `source/`, `thumbnails/`, `temporary/` крім `hls/`; `root /` у `location ^~ /media/` віддає їх усі.
РІШЕННЯ: формую 18 знахідок; blocking — ті, що ламають дані або роблять цілий клас рядків невиконуваним.
