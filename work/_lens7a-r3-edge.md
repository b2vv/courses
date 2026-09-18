ДАЛІ: читаю docs/transcript/plan.md (ред. 6) — блок C01–C22, D-1a, D-2, D-5, X14, §7a; потім spec.md; потім infrastructure/tests/smoke.sh та integration.sh, nginx conf, код Phase 6 плеєра.

ФАКТ: план ред.6 прочитано цілком. Нові куски: C01–C22, D-1a, D-2 (таблиця «число → сторож» + нерівність), D-5 (duration===0), X14 (nginx), §7a (2 поправки).
ДАЛІ: spec.md, потім infrastructure/tests/smoke.sh + integration.sh + nginx.conf — з'ясувати, чи smoke реально ходить крізь nginx і чи integration бачить конфіг.
ФАКТ: smoke.sh реально ходить крізь nginx (`http://nginx:8080`) з контейнера frontend через Node fetch. Метрики читає прямо з сервісів (`http://upload:8082/metrics`, `http://media-worker:8081/metrics`), бо крізь nginx `/metrics` -> 404.
ФАКТ: integration.sh робить хостовий awk-скан `infrastructure/nginx/nginx.conf` ДО docker; але awk **скидає стан на кожній `{`/`}`** (`reset_statement()`), тобто не має жодного поняття про блок/location. Перевіряє лише `log_format safe` і `access_log`.
ФАКТ: nginx.conf:20 `client_max_body_size 1m` у http{}; :92 `16m` у `^~ /uploads/`; `^~ /api/` (116-126) не перевизначає. `= /api/v1/auth/login` (105) — окремий location.
ФАКТ: smoke.sh вичерпує зону `login` (10r/m) 30 запитами і зону `api` (10r/s burst 20) 60 запитами в першому ж блоці; адмінська кука пишеться в /tmp/courses-smoke-media-cookie.
ДАЛІ: spec.md (A24, A21f, G1–G6), чи є в уроків мова, LessonPlayer duration/seekTo/key, чи є /metrics в api.

ФАКТ: у backend/ НЕМАЄ жодного входження слова "language" — ні в міграціях, ні в домені. `course_lessons` (00003_authoring.sql:35-44) не має мови. D-1a «віддає мову УРОКУ» спирається на поле, якого не існує; §4 (00007) додає language лише в `transcripts`; жодна задача X1–X14 не додає його урокові.
ФАКТ: `<LessonPlayer>` у learn.$.tsx:82 БЕЗ `key`. `duration` = useState(0) (LessonPlayer.tsx:94), скидається лише в onReady. `pendingTimeRef`/`positionRef` = useRef(resumeFrom) — ініціалізуються ОДИН раз за монтування. Ефект медіа має deps [createMedia, handleMediaError, lesson.videoState, videoElement] — без lesson.id.
ФАКТ: seekTo клампить Math.max(0, Math.min(Math.max(1, duration), next)) — підтверджує D-5.
ФАКТ: Makefile:83 `check: ... seed smoke e2e`. `backend/sql/seed.sql` (20 рядків) створює ЛИШЕ 5 курсів — ні модулів, ні уроків, ні video_assets. Отже на момент `smoke` немає ані уроку, ані готового HLS.
ФАКТ: smoke.sh усі медіа-фетчі очікують 403; жодного 200 на медіа-байти в smoke не існує.
ФАКТ: api сервить GET /metrics (platform/httpserver/server.go:78) з `courses_http_requests_total{method,route,status}` — лічильник для C02 будується.
ФАКТ: прецедент мапінгу 23505 за ім'ям констрейнта є (authoring.go:428, repository.go:239, media/repository.go:502).
РІШЕННЯ: 8 знахідок, 5 blocking.

## ЗНАХІДКИ (фінал)
E1 blocking — рівність `MaximumTranscriptBody == client_max_body_size(/api/)` робить доменний код мертвим (D-2, C01/C02).
E2 blocking — C05/C04/C01/C02 на рівні `smoke`: seed.sql не дає ні уроку, ні готового HLS; C05 зелений вакуумно на 403.
E3 blocking — D-1a спирається на «мову уроку», якої в схемі не існує; імпорт VTT/SRT не має джерела мови.
E4 blocking — `duration` ніколи не скидається: C10 (після витіснення) не задовольняється правилом D-5; зміна уроку дає застале вікно.
E5 blocking — `pendingTimeRef = useRef(resumeFrom)` ініціалізується раз: `?t=` при in-app навігації або не застосується, або перетече в сусідній урок (X5, P06/B12/B15).
E6 non-blocking — одиниця `maxLength`: символи vs байти в нерівності D-2 (UTF-8 + JSON-escape до 6×).
E7 non-blocking — awk-скан integration.sh блок-сліпий (reset_statement на `{`/`}`) + не парсить суфікси `m`/`k`: C03 як описано не будується.
E8 non-blocking — C07 сформульований диз'юнкцією («не діє АБО веде в правильну секунду») — не може почервоніти; C09 склеює два механізми.
E9 non-blocking — C15/C20 не мають названого коду для unique-violation при відсутньому `updatedAt`-токені; C22 не покриває author-supplied document url.
E10 non-blocking — `gzip on` у http{} накриває і стрімлений SSR (`location /` з `proxy_buffering off`); C06 у браузері недосяжний (Accept-Encoding — forbidden header).
