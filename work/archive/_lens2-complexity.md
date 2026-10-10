
## Лінза 2 — надлишкова складність (ред.2), старт 2026-09-07
ДАЛІ: прочитати docs/playback/plan.md (ред.2) цілком; тоді spec.md §5 G1–G14; тоді перевірити §2 "Reuse-first" цитати фактично у backend/ і frontend/.

ФАКТ: план ред.2 = 459 рядків. Структура: §1 розміщення, §2 reuse-first (2 таблиці), §3 D-1..D-8,
§4 сайд-ефекти, §5 міграція, §6 rollback, §7 конституція, §8 приймальні (P01-P14, B01-B28, R01-R24 = 66),
§9 ризики, §10 задачі (порожня).
ФАКТ: нові механізми ред.2: map $uri->$media_video, map $cookie_playback->$playback_absent,
proxy_no_cache+proxy_cache_bypass, error_page 500..504 -> @media_denied, Lua EVAL, client_seq,
кука playback Path=/, TTL з platform/config, video_assets.duration_seconds.
ДАЛІ: spec.md §5 G1-G14 (що ухвалено, не чіпати); тоді фактична перевірка §2 цитат.

## Крок 2 — фактична перевірка §2 «Reuse-first»

ФАКТ (вірні цитати): `backend/platform/httpserver/actor.go:7,16` — Actor/ActorFromContext ✓.
`identity/adapters/http/handler.go:123` Authenticate, `:142` RequireSameOrigin ✓.
`backend/platform/redis/client.go:12` Open → `*redisclient.Client` (go-redis v9) — EVAL справді доступний ✓.
`frontend/src/adapters/api/transport.ts:34` createRequestControl ✓.
`admin-video-queries.ts:12 videoPollMs = 3_000`, `:69` refetchInterval ✓ (сайд-ефект «на порядок рідше» правдивий).
`sql/queries/media.sql:41-45` — LEFT JOIN courses→modules→lessons→video_assets: прецедент прямого джойна ✓.
`course/domain/authoring.go:23 MaximumLessons = 120` ✓ (B19 правильний).
`nginx.conf:40-42` три add_header на рівні server ✓ (C25 обґрунтований), `:45-46` /media 403 ✓.
`docker-compose.yml:141` media-data:/media:ro, `:152-153` read_only+tmpfs ✓.

ФАКТ (ХИБНА цитата): `backend/platform/config/` містить РІВНО `dsn.go` + `dsn_test.go` —
це складач DSN (PostgresDSN/RedisDSN), а не «типізований рантайм-конфіг». Такого механізму в репо нема.
Прецедент репо для TTL: `identity/application/identity.go:13 sessionLifetime = 12 * time.Hour` — КОНСТАНТА.
Прецедент для env-конфігу модуля: `identity/adapters/http/config.go:18 ConfigFromEnvironment`.
→ ЗНАХІДКА 1 (blocking): конфігурований TTL має єдиного замовника — тестовий рядок P08.

ФАКТ (ХИБНЕ обґрунтування §2 C16 «два браузери надсилають однакову сесійну куку»):
`identity.sql.go:41 CreateIdentitySession` вставляє НОВИЙ рядок user_sessions на кожен логін;
`deleteIdentitySession` видаляє лише пред'явлений token_digest. Отже два браузери мають РІЗНІ
токени й різні session id. Плюс `playback`-кука браузер-скоупна так само, як і сесійна:
дві вкладки одного браузера отримують ту саму куку → нульове додаткове розрізнення.
→ ЗНАХІДКА 2 (blocking).

ФАКТ: `ResolveIdentitySession` (generated:382-396) селектить із user_sessions, але НЕ повертає
user_sessions.id — додати одне поле до наявного запиту дешевше за новий носій ідентичності.

ФАКТ (duration): у backend слово duration поза observability/time.Duration трапляється лише в
`media/domain/source.go:19 SourceProbe.DurationSeconds` і `ffmpeg/transcoder.go:167,198-202`.
У схемі колонки немає — твердження плану правдиве. Але ціна обраного шляху: ALTER + зміна шляху
публікації воркера + NULL-гілка D-5 + B06 + прийнята продуктова діра §9.
→ ЗНАХІДКА 4.

ФАКТ (офлайн-черга): у `frontend/src` НЕМАЄ жодного sendBeacon / navigator.onLine / offline /
retry-черги (grep порожній; навіть tus-client не має retry). Отже сценарій B11, заради якого
заведено `client_seq`, не породжується жодним названим компонентом.
→ ЗНАХІДКА 6.

## Крок 3 — nginx-блок ред.2, хто чий замовник
ФАКТ: `proxy_cache_valid 200 204 30s` кешує ЛИШЕ успіх. 204 не може бути видане без чинної
playback-куки → запис під ключем "|<video>" (порожня кука) недосяжний. Отже
`map $cookie_playback $playback_absent` + `proxy_no_cache` + `proxy_cache_bypass` (7 рядків)
захищають неможливий стан; коментар у плані виправдовує map лише блоками, яким map потрібна — кільце.
→ ЗНАХІДКА 3.
ФАКТ: `auth_request` пропускає лише 2xx, тож при 5xx байти й так не течуть. `error_page 500..504`
+ named location існують заради формулювання R11 («саме 403»), і роблять транзієнтний збій Go
невідрізненним від постійної відмови для плеєра. → ЗНАХІДКА 5.

## Крок 4 — §8 і D-8
ФАКТ: 14 P + 28 B + 24 R = 66. Підмножини: R14 ⊂ B15; R15 ≈ R16; R01 ⊂ B14; B13 ≈ B21; R02 ⊂ R07.
R23 друга половина («сумарний обсяг тіл відповідей api не зростає») міряє властивість, гарантовану
відсутністю маршруту. → ЗНАХІДКА 8.
ФАКТ: D-8 аргументує злиття «та сама каденція 30 с», але B10 вимагає запису на паузу/перемотування/
приховування/перехід — інша каденція й fire-and-forget на unload, де 409 нікому читати. → ЗНАХІДКА 7.
ФАКТ: §1 лишає поріг 95% у playback, а обрізання позиції на читанні — у course. → ЗНАХІДКА 9.
ДАЛІ: фінальна відповідь.
