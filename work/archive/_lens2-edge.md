
# Лінза 2 — edge cases (ред.2, повторний прогін)

ДАЛІ: читаю docs/playback/plan.md (ред.2), spec.md (A1–A37, §5 G1–G14), потім infrastructure/tests/smoke.sh, integration.sh, Makefile — щоб перевірити, чи smoke справді спостерігає те, що йому приписано в §8.

## Крок 1 — матеріал й машинна звірка покриття

ФАКТ (скрипт над §8): рядків рівно 66 (14 P + 28 B + 24 R). Збігається з брифом.
ФАКТ: **критерії без жодного рядка — A7, A34, A37.** A28 (дірка ред.1) закрито рядком P14;
A34 лишилась незакритою вдруге. Нові дірки проти ред.1: A7 (підроблений/змінений доказ;
відмова не розкриває, що не зійшлось) і A37 (доказ не потрапляє в логи, зокрема в URL).
ФАКТ: рядки без критерію A: B12, B13, B20, B21 (→ D-2 / ADR 005), R10 (→ D-2), R21 (→ D-6).
Вони не «зайві» — вони прив'язані до рішень плану, але тоді покриття A-осі рахується лише по
60 рядках.

ФАКТ (smoke.sh): весь корпус — один `docker compose exec -T frontend node --input-type=module`
heredoc; fetch на `http://nginx:8080` та напряму `http://upload:8082/metrics`,
`http://media-worker:8081/metrics`. Отже механізми, які smoke МАЄ: куки (заголовок `Cookie`
вручну), кілька акаунтів (логін фікстур), 429 бурстом (рядки 68–79), `Range` (звичайний fetch),
підрахунок звернень до Go (`courses_http_requests_total{method,route,status}` є —
`backend/platform/observability/metrics.go:26`, роут з `mux.Handler(r)`,
`backend/platform/httpserver/server.go:92`).
ФАКТ: механізмів, яких у smoke НЕМАЄ: (а) зупинка контейнерів — це host-рівень
(`docker compose stop`), а тіло тесту виконується всередині контейнера; (б) метрики **обсягу
тіл відповідей** — у `observability/metrics.go` є лише `Requests`, `Duration`,
`DatabaseDuration`, `TranscodingJobDuration`, `TranscodingFailures`; лічильника байтів немає;
(в) читання логів api (smoke ніде не читає логи).
ФАКТ (Makefile:54): `check: … build up seed-identities seed smoke e2e` — **smoke йде ДО e2e**.
ФАКТ: єдине джерело реального HLS у репо — `frontend/playwright/.fixtures/*.mp4`, які
завантажує e2e (`frontend/e2e/media.setup.ts`). `backend/sql/seed.sql` вставляє лише courses.
Отже на момент smoke том `media-data` порожній.

ДАЛІ: формулюю знахідки; перевіряю кеш-ключ проти R06, TTL куки проти каденції heartbeat,
`Cache-Control: no-store` проти `proxy_cache`.

## Крок 2 — перевірки швів (факти з коду)

ФАКТ (`backend/internal/identity/adapters/http/handler.go:245`): у репо є конвенція
`identityHeaders` → `Cache-Control: no-store` на автентифікованих відповідях. Nginx
`proxy_cache` за замовчуванням **не кешує** відповідь із `Cache-Control: no-store`
(і з `Set-Cookie`). D-2 не має ні `proxy_ignore_headers Cache-Control Set-Cookie`, ні
вимоги «authorize не ставить цих заголовків».
ФАКТ (D-2): `proxy_cache_key "$cookie_playback|$media_video"` — сесійної куки в ключі немає.
ФАКТ (D-3): дозвіл вимагає **обох** кук; отже кеш обслуговує рішення, ухвалене для іншої пари.
ФАКТ (D-1 + D-8 + P08): TTL куки 180 с (10 с у тестовому профілі), каденція heartbeat 30 с;
перевидання куки на heartbeat ніде не сказано (лише R15 каже «нової куки **не** видано» при 409).
ФАКТ (`observability/metrics.go`): лічильника байтів відповідей немає.
ФАКТ (`docker-compose.yml`): усі сервіси `restart: unless-stopped`, тож `docker compose stop`
з хоста працює; але тіло smoke виконується всередині `frontend`.

## Крок 3 — знахідки

E1 blocking  — A7, A34, A37 без жодного рядка (A28 ред.1 закрито P14, A34 — ні).
E2 blocking  — кеш-ключ без сесійної куки: теплий кеш робить доказ bearer-здатністю (A5/R06).
E3 blocking  — TTL доказу проти каденції heartbeat; перевидання куки не назване (P08 не зелений).
E4 blocking  — на момент smoke на томі немає жодного HLS-байта (`make check`: smoke до e2e).
E5 blocking  — R11/R12/R13 потребують host-оркестрації зупинок і відновлення; інакше труять e2e.
E6 blocking  — `no-store`/`Set-Cookie` від Go тихо вимикає proxy_cache → зникає вікно A32.
E7 blocking  — R22 «позиція > тривалості → 400» проти D-5 (обрізання на читанні) і B06 (NULL).
E8 non-block — «кеш порожній/обійдено» без механізму (ключ уже прогріто B13/B20/B21).
E9 non-block — B13/B21 не називають спостережуване (яка серія метрики, який роут-лейбл).
E10 blocking — R23 «обсяг тіл відповідей api» — такої метрики в репо немає.
E11 non-block — дві вкладки одного браузера: ідемпотентний `play` дає два одночасні перегляди.
E12 non-block — 403 не кешується: невалідна кука б'є в Go на кожен сегмент.
E13 non-block — теплий 204 переживає відмову Redis/PG/Go — §4 і §9 кажуть різне, рядка немає.
E14 non-block — B28/R-рядки smoke потребують стану БД (`ready` + опублікований курс), якого seed не робить.
E15 non-block — «Почати навчання» на курсі без уроків (P03 проти B01) не покрито.

ЧИСТО: порожня кука `playback` (map `$playback_absent` + bypass/no_cache) — механізм є й
коректний; кука є, а власника в Redis немає → 403 (покрито R05); гонка heartbeat/`play`
атомарна через Lua `EVAL` (D-4) і покрита R15/R16; подвійний `play` — B17.
