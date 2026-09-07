
## Лінза «логічні діри» — прогін від 2026-09-07

ДАЛІ: читаю docs/playback/plan.md (предмет), потім spec.md, _wip-phase6.md, і код:
nginx.conf, docker-compose.yml, identity handler, migrations, openapi.yaml,
frontend composition/identity.ts, media transcoder. Мета — шви: кука vs hls.js,
auth_request vs proxy_cache, гонки play/heartbeat, міграція vs rollback, X-Original-URI -> lesson.

ФАКТ (код): `docker-compose.yml` nginx має `tmpfs: [/tmp, /var/cache/nginx]` — /var/cache/nginx
у compose ЗАПИСУВАНИЙ. Виміряна «пастка» D-2 суперечить власному compose репо.
ФАКТ (код): таблиця уроків зветься `course_lessons` (00003_authoring.sql:35), `lessons` НЕ існує.
План §5 обіцяє FK на `lessons(id)`.
ФАКТ (код): `duration` немає ніде — ні в міграціях, ні в openapi (grep порожній). D-5/B05/B06
рахують 95% і обрізають позицію «до тривалості поточної генерації» — джерела тривалості нема.
ФАКТ (код): ffmpeg пише відносні URI (`-master_pl_name master.m3u8`, `%v/index.m3u8`,
`hls_fmp4_init_filename init.mp4`) → усе під /media/hls/<id>/ → Path=/media покриває.
ФАКТ (код): сесійна кука Path=/ SameSiteLax (handler.go:172) → летить у /media same-origin. OK.
ФАКТ: $request_uri у subrequest = URI батьківського запиту → дефолтний cache key
$scheme$proxy_host$request_uri = ПОСЕГМЕНТНИЙ і БЕЗ куки.
ДАЛІ: перевірити глобальні Cache-Control у httpserver (чи не заб'ють proxy_cache).

ФАКТ (код): httpserver не ставить глобальний Cache-Control → 204 від authorize кешується;
але Set-Cookie у відповіді authorize мовчки вимкнув би кеш (nginx не кешує Set-Cookie).
ФАКТ (код): nginx.conf свідомо ходить у сервіси через `set $api http://api:8080; proxy_pass $api;`
+ `resolver 127.0.0.11 valid=5s` (коментар «Resolve service replacements»). Знімок D-2 бере
літерал `proxy_pass http://api:8080/...` — резолв один раз на старті.
ФАКТ (код): /media у nginx = увесь том media-data: source/<id>/original, thumbnails/, temporary/,
не лише hls/. Знімок D-2 відкриває `^~ /media/`, а не `^~ /media/hls/`.
ФАКТ (код): auth_request перетворює будь-який не-2xx/401/403 від upstream на 500 → R11 (Go лежить)
дасть 500, а не 403.

РІШЕННЯ: 8 blocking + 10 non-blocking, оформлюю у фінальну відповідь.
