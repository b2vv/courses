
# ЛІНЗА «Надлишкова складність» — docs/playback/plan.md ред.4

ДАЛІ: читаю plan.md ред.4 повністю, потім spec.md §5 (G1–G14), spike-edge.*, infrastructure/nginx/nginx.conf, історію критик. Шукаю: сесія відтворення (2 ключі Redis), §3 D-2 конфіг edge, D-8 heartbeat+прогрес, D-5 reason×6, §8 82 рядки, reuse наявних механізмів.

ФАКТ: plan.md ред.4 прочитано (522 рядки). Структура: §1 розміщення, §2 reuse-first, §3 D-1..D-8, §4 сайд-ефекти, §5 міграція, §6 rollback, §7 конституція, §8 приймальні (P01-P14, B01-B38, R01-R30 = 14+38+30 = 82), §9 ризики, §10 задачі.
ФАКТ: §8 рівні: рахую розподіл smoke/e2e/integration/unit нижче.
ДАЛІ: читаю spec.md §5 (G1-G14), infrastructure/nginx/nginx.conf, spike-edge.md/.conf.

ФАКТ (#2, велике): infrastructure/nginx/nginx.conf:8-12 ВЖЕ містить усі пʼять *_temp_path у /tmp
(client_body, proxy, fastcgi, uwsgi, scgi). План D-2 подає їх як нову знахідку спайку
(«БЕЗ ЦИХ ТРЬОХ nginx НЕ СТАРТУЄ», «Продакшн живий лише тому, що кеша в ньому немає»).
Спайк-конфіг (spike-edge.conf) — окремий мінімальний файл, у ньому їх спершу не було; у репо є.
ФАКТ: nginx.conf:36 вже має `resolver 127.0.0.11 valid=5s` і патерн `set $api http://api:8080` (66,77,88) — план це переюзує коректно.
ФАКТ: nginx.conf:40-42 три security-заголовки стоять на рівні server. Повторення їх у location
D-2 НЕ надлишкове: nginx add_header у location скасовує успадковані. Тут план правий.
ДАЛІ: перевіряю logout/сесію в identity (чи authorize уже відсікає розлогіненого), platform/redis, smoke.sh/Makefile, рахую рівні §8.

ФАКТ (#1): жодного читача по `playback:session:{id}` без userId немає. authorize має куку→userId,
heartbeat має куку→userId, logout має userId. Отже ключ, адресований id, не має шляху доступу,
який не починався б із userId. Один ключ `playback:owner:{userId}` → {playbackId, videoId, lessonId}
задовольняє всі три умови D-3 п.2.
ФАКТ (#1, дірка від двох ключів): D-4 heartbeat робить EXPIRE ЛИШЕ над owner-ключем
(«Lua EVAL: if GET==ARGV[1] then EXPIRE»). TTL `playback:session:{id}` ніхто не поновлює →
на 180-й секунді умова D-3 п.2 «session існує» падає → 403 посеред перегляду → P08 червоний.
ФАКТ (#1, lessonId виправданий): у тілі D-8 heartbeat lessonId НЕМА, отже сесія — єдине джерело
уроку для запису прогресу. Поле заслужене.
ФАКТ (logout): identity/application/identity.go:90 Logout → sessions.Revoke — серверне відкликання.
Отже після logout умова D-3 п.1 («кука дійсна») уже падає. Хук D-4 `logout` нічого не закриває:
вікно 180 с, яким він виправданий, не існує; єдине справжнє вікно — 30 с кеша edge, і DEL його не
скорочує (це і сам план визнає в §9).
ФАКТ (#2): nginx.conf:8-12 — усі 5 *_temp_path уже в /tmp. nginx.conf:36 — resolver уже є.
nginx.conf:40-42 — три security-заголовки на рівні server; повторення в location обовʼязкове
(add_header у location скасовує успадковані) → тут план правий, не роздування.
ФАКТ (#5): §8 = 82 рядки. Рівні: smoke 28, e2e 24, integration 18, unit 12, contract 1.
З 24 e2e щонайменше 10 вимагають двох контекстів (P05,P13,P14,B15,B16,B31,B33,R14,R25,R26);
4 чекають реального 30-с вікна кеша (R04,B15,R14 + P08 на TTL).
ФАКТ (#6): frontend `refetchInterval` (admin-video-queries.ts:69, videoPollMs=3_000) — це поллінг
GET-запиту за настінним часом. B09 вимагає лічильник за ЧАСОМ ВІДТВОРЕННЯ, а D-8 — POST.
Рядок §2 «refetchInterval | каденція heartbeat» веде виконавця не туди.
ФАКТ (#кеш): proxy_cache_key "$pb|$vid" не містить особи. Спека 2.5 (URL надіслано другові)
і R06 вимагають 403 для чужого; кеш віддасть 200 до 30 с. §4 і §9 стверджують протилежне.
ФАКТ (#конфіг): P08 ставить ownershipTTL=10 с у тестовому профілі, а каденція heartbeat (A23) — 30 с.
10 < 30 → володіння спливе між биттями і P08 впаде. Конфіг мусить нести ДВА значення з інваріантом.
РІШЕННЯ: 10 знахідок, 5 blocking. Пишу фінальну відповідь.
