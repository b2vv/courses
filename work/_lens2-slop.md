
## Лінза 2 — AI-slop, ред.2 (прогін 2)

ДАЛІ: читаю docs/playback/plan.md (ред.2) повністю, потім spec.md §5 G1–G14, реєстр критики C1–C41, _wip-phase6.md. Мета: знайти правки з міткою ⟵ Cnn, за якими стоїть лише проза, а не механізм; фактично перевірити всі цитати/номери рядків/«заміри» ред.2 (особливо транскрипт docker compose у §3 D-2) через sed/grep/docker compose config.

ФАКТ: план прочитано (459 рядків). Список фактичних тверджень до перевірки:
media.sql:41-45; handler.go:123,142; admin-video-queries.ts:69; docker-compose.yml:152-153 (tmpfs);
транскрипт docker compose run (101 / 1777 / 775); 00003_authoring.sql:35; 00005_media.sql:6 та :1-2;
integration.sh:26-34; videoPollMs=3_000; MaximumLessons=120; boundaries_test.go; check-architecture.mjs;
accessibility.spec.ts; platform/redis EVAL; platform/config; make check-generated; smoke.sh; AGENTS.md.
ДАЛІ: перевіряю кожне через sed/grep.

ФАКТ (перевірено фактично, не з пам'яті): транскрипт §3 D-2 **справжній**. Відтворив:
`docker compose run --rm --no-deps --entrypoint sh nginx -c 'id; stat -c "%n %U:%G %a" /tmp /var/cache/nginx; ...'`
→ uid=101(nginx); `/tmp root:root 1777` + MKDIR OK; `/var/cache/nginx root:root 775` + MKDIR DENIED
(`mount` показує `mode=775` на tmpfs /var/cache/nginx). Образ — nginxinc/nginx-unprivileged:1.30.4-alpine.
docker-compose.yml:152-153 = `read_only: true` / `tmpfs: [/tmp, /var/cache/nginx]` — цитата точна.
ВИСНОВОК: цей «замір» — не вигаданий доказ. Знімаю підозру.

ФАКТ: інші рядкові цитати теж точні —
media.sql:41-45 = FROM courses + 3 LEFT JOIN (прецедент прямого джойна) ✓
handler.go:123 Authenticate, :142 RequireSameOrigin ✓
00003_authoring.sql:35 = `CREATE TABLE course_lessons` (таблиці `lessons` справді немає) ✓
00005_media.sql:6 = `lesson_id ... REFERENCES course_lessons(id)` ✓; :1-2 = «Down ... test-only» ✓
integration.sh:26-34 = `docker run ... -v "$task_root/backend:/work"` — монтує лише backend/ ✓
admin-video-queries.ts:69 = refetchInterval; videoPollMs = 3_000 (рядок 12) ✓
MaximumLessons = 120 (authoring.go:23) ✓
boundaries_test.go, check-architecture.mjs, frontend/e2e/accessibility.spec.ts, smoke.sh, make check-generated — усі існують ✓

ФАКТ (розбіжність 1): `platform/config` — це НЕ «типізований рантайм-конфіг». Пакет містить лише
dsn.go/dsn_test.go: `PostgresDSN(getenv)` та збірка DSN. Жодного механізму типізованих
рантайм-налаштувань (TTL) там немає.
ФАКТ (розбіжність 2): server-level add_header у nginx.conf:40-42 — ТРИ заголовки, не чотири.
ФАКТ (розбіжність 3): nginx.conf уже має `resolver 127.0.0.11` і патерн `set $api http://api:8080; proxy_pass $api;`
на рядках 77-78 — тобто «виправлення C22» повторює вже наявну конвенцію (це ок, але не відкриття).
ФАКТ (розбіжність 4, кандидат у знахідку): map `$playback_absent` + proxy_no_cache/proxy_cache_bypass —
мертві директиви: кешуються лише 200/204 (`proxy_cache_valid 200 204 30s`), а запит без куки
ніколи не дає 204. Тобто map існує тільки щоб живити директиви, які самі нічого не роблять,
а коментар усередині конфіга виправдовує map тим, що без неї змінна була б неоголошена — коло.

ДАЛІ: читаю spec.md §5 G1–G14 і §приймальні A1–A36, потім реєстр C1–C41, щоб перевірити,
чи кожна мітка ⟵ Cnn несе механізм, а не переказ.

ФАКТ: §8 покриває A1–A33, A35, A36 — **A34 і A37 не згадані жодним рядком** (`grep 'A34\|A37' plan.md` → порожньо).
Реєстр C13 = «A28 і A34 не покриті». Ред.2 додала P14 з міткою ⟵ C13 — закрито лише A28.
ФАКТ: `backend/platform/observability/metrics.go` — 6 колекторів: Requests(method,route,status), Duration,
DatabaseDuration, QueueDepth, TranscodingJobDuration, TranscodingFailures. **Метрики розміру відповіді немає.**
nginx `log_format safe` = requestId/method/status/duration — **без body_bytes_sent**; `error_log /dev/null`.
Отже спостережуване R23 «сумарний обсяг тіл відповідей api не зростає ⟵ C35» не існує в жодному каналі.
ФАКТ: `frontend/rstest.config.ts:43` — `testEnvironment: 'happy-dom'` (jsdom у репо немає взагалі).
axe-core 4.13.0 є, вживається в 3 unit-тестах + e2e/accessibility.spec.ts.
ФАКТ: `backend/platform/config` = PostgresDSN/RedisDSN. Типізованого рантайм-конфігу немає.
Наявний прецедент модульного конфігу — `identityhttp.ConfigFromEnvironment(os.Getenv)` (cmd/api/main.go:32), і він не названий.
ФАКТ: `Makefile:51-53` check-generated діфає лише course/identity/frontend generated — **media вже не покрито**,
`backend/sqlc.yaml` має три записи (courses, identity, media). Нового playback гейт не побачить.
ФАКТ: `Makefile:8-12` up = `docker compose up -d --build --wait --wait-timeout 180` + `docker compose exec -T nginx nginx -s reload`.
ФАКТ: nginx.conf:40-42 — три успадковані add_header, не чотири.

РІШЕННЯ: 6 blocking + 4 non-blocking. Транскрипт D-2 знято з підозри — він відтворюється байт у байт.

## Знахідки лінзи AI-slop (ред.2)

[slop] `platform/config` названо носієм TTL — пакет цього не вміє
де:      §2 «Переюзуємо без змін», рядок «`platform/config` типізований рантайм-конфіг ⟵ C19»; §3 D-1 («TTL із конфігу, дефолт 180 с»); §8 P08
чому:    `backend/platform/config` — це dsn.go/dsn_test.go: `PostgresDSN`/`RedisDSN`. Типізованого рантайм-конфігу там немає. C19 казав «P08 спирається на механізм, якого план не проєктує» — ред.2 відповіла посиланням на пакет замість проєктування: не названо ні змінної середовища, ні структури, ні місця зчитування, ні того, хто підставляє 10 с у тестовому профілі. Виконавець дійде до P08 і винайде ручку сам. Плюс C33 (ручка TTL у проді відкручує вікно доступу) не відповіла ніде.
що зробити: у §3 D-1 назвати конкретно: ім'я env (напр. `PLAYBACK_PROOF_TTL`), дефолт 180 с, місце зчитування за прецедентом `identityhttp.ConfigFromEnvironment(os.Getenv)` (`cmd/api/main.go:32`), верхню межу валідації; у §2 замінити рядок `platform/config` на цей прецедент.
вага:    blocking
не закриває: C19 (перша половина), C33

[slop] R23 спостерігає величину, якої немає в жодному каналі
де:      §8 R23 «сумарний обсяг тіл відповідей api не зростає ⟵ C35»
чому:    `platform/observability/metrics.go` має 6 колекторів; жоден не міряє розмір відповіді (`Requests` — лічильник із method/route/status). `log_format safe` у nginx.conf несе requestId/method/status/duration без `body_bytes_sent`, `error_log /dev/null`. C35 казав саме це — «читають логи, у яких нічого немає»; ред.2 замінила один неспостережуваний спостережуваний на інший і поставила мітку. Рядок зеленітиме тавтологічно або не напишеться взагалі. При цьому робочий інструмент у репо є і не названий: smoke.sh уже рахує `courses_http_requests_total` із `/metrics` (`infrastructure/tests/smoke.sh:22-27`).
що зробити: переписати R23 на лічильник за прецедентом smoke.sh: `courses_http_requests_total` з міткою `route` для медіа-маршрутів не зростає під час перегляду; або явно додати задачу «нова метрика розміру відповіді» — і тоді назвати її в §2/§7.
вага:    blocking
не закриває: C35

[slop] B24 приписує unit-рівню рівно те, що B25 визнає для нього недосяжним
де:      §8 B24 («фокус видимий; ≥44px; axe без порушень | unit + axe») проти B25 («**e2e** … jsdom цього не спостерігає ⟵ C18»)
чому:    unit-рівень у репо — `testEnvironment: 'happy-dom'` (`frontend/rstest.config.ts:43`), без layout і paint: `getBoundingClientRect` дає нулі, видимість фокус-кільця не обчислюється. C18 виправлено точково в одному рядку й не поширено на сусідній — той самий дефект лишився в §8 після абзацу «Рівні названо правильно ⟵ C12». Додатково: у репо jsdom немає взагалі (happy-dom), тобто пояснення в B25 називає не той рушій.
що зробити: розщепити B24 — клавіатурна досяжність, ARIA й axe лишаються unit; «≥44px» і «фокус видимий» переносяться в `frontend/e2e/accessibility.spec.ts`; у B25 замінити «jsdom» на «happy-dom».
вага:    blocking
частково не закриває: C18, C12

[slop] мітка ⟵ C13 на P14 закриває половину знахідки; A34 і A37 лишились без жодного рядка
де:      §8 P14 «Ліміт діє наскрізь по курсах ⟵ C13»; §8 в цілому
чому:    C13 = «A28 **і A34** не покриті». Ред.2 додала рядок на A28 і поставила мітку, ніби C13 закрито. `grep 'A34\|A37' plan.md` — порожньо: §8 посилається на A1–A33, A35, A36, тобто 35 із 37 критеріїв. A34 («ліміт по акаунту, не по IP, без fingerprint») і A37 («жоден доказ не потрапляє в логи, зокрема як частина URL») не мають ні позитивного, ні негативного спостереження — при тому що A37 напряму стосується куки `Path=/`, яка тепер летить у `/api/*`.
що зробити: додати два рядки — негативний на A34 (два браузери з одного акаунта за різними IP → витіснення все одно; різні акаунти з одного IP → не витісняють одне одного) і рядок на A37 (`docker compose logs api|nginx` під час перегляду не містить `playback=`); зняти мітку ⟵ C13 з P14 або дописати, що закрито лише A28.
вага:    blocking
не закриває: C13 (друга половина)

[slop] §7 називає машинний гейт, який структурно не дивиться на новий модуль
де:      §7, рядок «Контракт не дрейфує ⟵ C4-конституційне | `make check-generated`»
чому:    `Makefile:51-53` діфає рівно три шляхи: `backend/internal/course/.../generated`, `backend/internal/identity/.../generated`, `frontend/src/adapters/api/generated`. `backend/sqlc.yaml` має вже **три** записи — courses, identity, **media** — тобто гейт не покриває навіть Phase 5. Новий `playback` зі своїм sqlc-виходом лишиться поза ним. Секція, переписана саме заради «вказівки на машинний гейт замість самоатестації», знову рапортує зелене там, де гейт сліпий.
що зробити: у §5/§7 додати рядок задачі: новий запис у `backend/sqlc.yaml` для `sql/queries/playback.sql` **і** розширення списку шляхів у `check-generated` (заодно дописати пропущений `media`).
вага:    blocking
не закриває: C4-конституційне (як доказ)

[slop] P12 на integration-рівні не бачить того, що заявлено назвою рядка
де:      §8 P12 «Проба AV1 **доїжджає** полем мінту ⟵ C19 | плеєр стартує | play | … | integration + unit»
чому:    механізм у D-7 правильний (поле в тілі мінту), але жоден із двох названих рівнів не спостерігає доїзд: Go-integration бачить тіло, яке сам тест і склав, а happy-dom не має `MediaSource`, тож проба в unit — це наша ж заглушка. Тобто єдиний рядок на A17 зеленітиме без працюючої фічі. Це той самий клас, що C-slop-1 знайшла на R15 ред.1 («перевіряє намір, а не поведінку»).
що зробити: додати e2e-крок у `playback.spec.ts` (справжній Chromium має `MediaSource`): після play перехопити `POST /api/v1/playback/sessions` і побачити поле проби; integration лишити на «Go інкрементує метрику за даним тілом».
вага:    blocking
не закриває: C19 (друга половина, у частині рівня)

[slop] `map $playback_absent` + `proxy_no_cache`/`proxy_cache_bypass` — мертві директиви, які виправдовують самі себе
де:      §3 D-2, блок `http` (map з тририядковим коментарем) і `location = /_media_auth`
чому:    кешуються лише `200 204` (`proxy_cache_valid 200 204 30s`), а запит без куки `playback` ніколи не дає 204 — Go його відхиляє. Отже під ключем `"|$media_video"` запис не з'явиться ніколи: ні `proxy_no_cache`, ні `proxy_cache_bypass` не змінюють жодної поведінки. Коментар усередині конфіга виправдовує map тим, що без неї «наступний блок посилався б на неоголошену змінну» — тобто map існує рівно щоб живити директиви, які існують рівно щоб виправдати map. Плюс сам коментар переказує історію рев'ю («та сама помилка, за яку критика вбила редакцію 1») — це поїде в продакшн-конфіг і житиме там роками.
що зробити: викинути map і обидві директиви (403 і так не кешується); якщо хочеться запобіжника — лишити один рядок коментаря по суті: «кешуються лише 200/204, тож рішення для аноніма не зберігається». Історію редакцій тримати в плані, не в nginx.conf.
вага:    non-blocking

[slop] B12 на дві третини повторює те, що вже доводить `make up`
де:      §8 B12 «контейнер здоровий; `nginx -t` exit 0; `/tmp/nginx-auth` існує й записуваний ⟵ C10, C37»
чому:    `Makefile:8-12`: `docker compose up -d --build --wait --wait-timeout 180` валить ціль до будь-якого smoke, якщо контейнер не здоровий, а наступний рядок `docker compose exec -T nginx nginx -s reload` валиться на невалідному конфізі — тобто «здоровий» і «`nginx -t` ok» доведені раніше, ніж smoke стартує. Нове тут рівно одне: каталог кешу існує й у нього пишеться. Це та сама знахідка, яку slop-1 вже подавала на B10 ред.1; вона не позначена й не врахована.
що зробити: скоротити B12 до одного спостережуваного — `docker compose exec -T nginx sh -c 'test -w /tmp/nginx-auth'` — і назвати `make up` передумовою, а не дією.
вага:    non-blocking

[slop] мітка ⟵ C21 закриває вікно витіснення й тихо губить другу половину — 30-секундну дірку A4
де:      §8 B15, R05, R12 (усі з ⟵ C21); §9 реєстр ризиків
чому:    C21 мав дві частини: (а) рядки без вердикту — виправлено, (б) «A4 має невизнану 30-секундну дірку: протухла кука ще обслуговується кешованим 204». Ред.2 відповіла на (б) не механізмом і не визнанням, а **передумовою тесту**: R05 має «кука старша за TTL, **кеш обійдено**», R12 — «кеш порожній». Жодного важеля «обійти кеш» D-2 не дає (`proxy_cache_bypass` прив'язано лише до відсутньої куки), тож передумова не має виконавця; а сам факт, що протухлий доказ до 30 с ще працює, не потрапив ні в §4 сайд-ефекти, ні в §9. §3 називає 30 с лише як «верхню межу вікна витіснення» — половину правди.
що зробити: дописати в §3 D-2 рядок «TTL кешу 30 с — це також верхня межа, протягом якої приймається доказ, що вже сплив», внести це в §9 як прийнятий ризик, і замінити передумову R05 на виконувану («сегмент, який ще не запитували в цьому вікні»).
вага:    non-blocking
частково не закриває: C21(б)

[slop] «усі чотири повторено» — успадкованих заголовків три
де:      §3 D-2, таблиця виправлень, рядок C25
чому:    `infrastructure/nginx/nginx.conf:40-42` — три `add_header` (nosniff, Referrer-Policy, X-Frame-Options). Четвертий, `Cache-Control`, не успадковується — він новий. Формулювання «усі чотири повторено» описує неіснуючу спадковість; механізм при цьому правильний.
що зробити: переписати клітинку як «три успадковані повторено + Cache-Control доданий».
вага:    non-blocking

## Чисто (перевірено фактично, підозру знято)
- Транскрипт `docker compose run ... nginx` у §3 D-2 — **справжній**. Відтворив командою з `--no-deps`: `uid=101(nginx)`, `/tmp root:root 1777` + MKDIR OK, `/var/cache/nginx root:root 775` + MKDIR DENIED, `mount` показує `mode=775` на tmpfs. Причина названа правильно, фікс `/tmp` правильний.
- Усі рядкові цитати ред.2 існують і збігаються: media.sql:41-45, handler.go:123/142, docker-compose.yml:152-153, 00003_authoring.sql:35, 00005_media.sql:1-2/:6, integration.sh:26-34, admin-video-queries.ts:69 + videoPollMs=3_000, authoring.go MaximumLessons=120, boundaries_test.go, check-architecture.mjs, e2e/accessibility.spec.ts, AGENTS.md (обидва хибних твердження на місці).
- Механізми, за якими стоїть код, а не проза: Lua compare-and-renew (D-4), `error_page → @media_denied` (D-3), `map $uri → X-Video-Id` (D-2), FK на `course_lessons` (§5), `client_seq` (D-5), правило NULL-тривалості (D-5), знятий третій порт (§7 ⟵ C29), реєстр ризиків §9 ⟵ C40, `set $api; proxy_pass $api` ⟵ C22 (збігається з наявною конвенцією nginx.conf:77-78).
