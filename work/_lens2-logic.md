
## Лінза 2 — логічні діри (прогін 2, ред.2 плану)
ДАЛІ: читаю docs/playback/plan.md §3 (новий текст) + §5, потім nginx.conf і docker-compose.yml, щоб перевірити, чи блок D-2 реально підніметься.

ФАКТ: план ред.2 прочитано цілком (459 рядків). §3 D-1..D-8 — новий текст; §5 міграція; §8 сценарії P/B/R.
ДАЛІ: nginx.conf + docker-compose.yml, щоб перевірити монтування тому, resolver, наявні location, tmpfs.

ФАКТ: nginx.conf — `resolver 127.0.0.11` вже є (http, рядок 36) → `set $api` резолвиться. Вже існують
`location = /media {403}` і `location ^~ /media/ {403}` (45-46). server-level add_header ×3 (40-42).
ФАКТ: docker-compose: nginx має `media-data:/media:ro` (141) → `root /;` на URI `/media/hls/...` дає
`/media/hls/...` — шлях збігається. tmpfs [/tmp, /var/cache/nginx] (153), read_only true (152).
ГІПОТЕЗА-1 (перевірити): `map $uri $media_video` у сабреквесті `auth_request` — `$uri` там ==
`/_media_auth`. Працює лише тому, що `if ($media_video = "")` у батьківському location форсує
обчислення на rewrite-фазі ДО access-фази, і sr успадковує масив змінних. Тобто коректність
ключа кешу тримається на `if`, який план мотивує зовсім іншим (валідація id, R20).
ГІПОТЕЗА-2: P08 (TTL=10с у тесті) vs D-8 (heartbeat кожні 30с) — кука і ключ протухнуть на 10-й с.
ГІПОТЕЗА-3: proxy_cache вимкнеться, якщо api віддає Set-Cookie або Cache-Control на /internal/*.
ДАЛІ: міграції, httpserver middleware (Set-Cookie/Cache-Control), розкладка /media на диску.

ФАКТ (міграції): `course_lessons` (00003:35) і `users` (00002:2) існують — §5 п.1 правий, C11 закрито.
Наступний номер справді 00006. АЛЕ конвенція репо: КОЖЕН FK явно називає ON DELETE
(`ON DELETE CASCADE` 00003:25,36; `RESTRICT` 00003:6; `SET NULL` 00005:6). §5 не називає жодного.
ФАКТ (критично): `internal/course/adapters/postgres/authoring.go:37` — `deleteMissingAuthoringLessons`
виконується на КОЖНОМУ збереженні структури курсу; `deleteAuthoringCourse` (68) каскадить у
modules→lessons. FK без ON DELETE = NO ACTION → після 00006 будь-яке видалення уроку, де вже є
прогрес, падає з FK violation. Регресія Phase 4.
ФАКТ: `Cache-Control: no-store` ставлять УСІ хендлери репо (course/handler.go:61,80;
identity/handler.go:246; media/handler.go:235 `mediaHeaders`). Якщо authorize піде за конвенцією —
`proxy_cache_valid` не спрацює, бо апстрімний Cache-Control має пріоритет без `proxy_ignore_headers`.
Кеш тихо не вмикається → B13/B21 падають, «один authorize на N сегментів» неправда.
ФАКТ: сесійна кука — SameSite=Lax, Path=/ (identity/handler.go:172); Set-Cookie ставиться лише на
login/logout, не на кожному запиті → Set-Cookie кеш не ламає. Ок.
ФАКТ: MaximumLessons=120 (authoring.go:23) — B19 правий. videoPollMs=3_000 — §4 правий.
ФАКТ: spec A4 «доказ живе 2–5 хв», A14 «перегляд довший за TTL не переривається».
РІШЕННЯ: P08 (TTL=10с у тестовому профілі) × D-8 (heartbeat раз на 30с) — кука й ключ Redis
протухають на 10-й с, перший heartbeat на 30-й. P08 недосяжний. Плюс: каденція heartbeat
рахується ЧАСОМ ВІДТВОРЕННЯ (B09 ⟵C34), а TTL — настінний → пауза довша за TTL убиває сесію.
ДАЛІ: оформити знахідки.

## ЗНАХІДКИ (10: 6 blocking, 4 non-blocking)
L1 (blocking) heartbeat на часі відтворення vs TTL на настінному — пауза/P08 рвуть сесію.
L2 (blocking) FK без ON DELETE ламає видалення уроку/курсу (authoring.go:37,68).
L3 (blocking) Cache-Control: no-store від Go вимикає proxy_cache — D-2 тихо не працює.
L4 (blocking) $media_video у сабреквесті == "" без `if`-праймінгу; жоден рядок §8 не ловить
    колапс ключа кешу (тепла кука A + інше відео).
L5 (blocking) D-4 не має рядка «play із чинною кукою, що НЕ володіє» — витіснений не рестартує.
L6 (blocking) client_seq не має точки скидання/сідування — друга сесія на тому ж уроці мертва.
L7 (non-blocking) гонка play × in-flight heartbeat → 409 самому собі.
L8 (non-blocking) duration_seconds без CHECK: 0 відтворює той самий баг, що й NULL.
L9 (non-blocking) error_page 500 → 403 знищує операційний сигнал (log_format пише лише статус).
L10 (non-blocking) R11/R12/R13 «кеш порожній» недосяжне як написано: мінт куки вимагає живого api.
+ map $playback_absent і proxy_no_cache/bypass — мертві (403 і так не кешується за proxy_cache_valid).
