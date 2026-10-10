ДАЛІ: прочитати work/brief-t16.md цілком, звірити обов’язкові репозиторні інструкції та визначити точний обсяг T16 без будь-яких git-операцій.
ФАКТ: work/report-t16.md на старті не існував; перша спроба append завершилася без змін, тому робочий файл створено з обов’язкового першого запису.
ФАКТ: прочитано скіл implement цілком; він вимагає регулярних локальних перевірок, повного фінального тестового проходу та code-review.
РІШЕННЯ: вимогу скіла про commit не виконувати, бо користувач прямо заборонив будь-які git-операції з посиланням на §4 брифа; решту процесу застосувати в межах брифа.
ДАЛІ: прочитати work/brief-t16.md цілком, а потім обов’язкові CTO-RESEARCH, план playback і tasks.
ФАКТ: work/brief-t16.md прочитано цілком. T16 змінює лише frontend/e2e/playback.spec.ts і має дати e2e-докази B20, B21, B21a, B22, B28, R23 та медіа-клаузули B36 без змін продукту, T14/T15, документації чи залежностей.
РІШЕННЯ: локально дозволені рівно npm run format:check, npm run lint, npm run typecheck і npm test; Docker-гейти та git-операції не запускати.
ФАКТ: спільний вивід CTO-RESEARCH, plan і tasks обрізався, тому їх не можна вважати прочитаними цілком.
ДАЛІ: дочитати CTO-RESEARCH, §8 плану, handoff і tasks окремими контрольованими порціями; потім оглянути точні шви e2e та інфраструктури.
ФАКТ: визначено розміри обов’язкових джерел; найбільші — plan 864 рядки, handoff 468 і playback.spec.ts 1553, тому читання буде порційним без ризику обрізання.
ДАЛІ: прочитати CTO-RESEARCH і tasks повністю, а з plan — щонайменше весь §8 з датованими поправками та «П’ять питань».
ФАКТ: CTO-RESEARCH прочитано повністю; він підтверджує несучий шов nginx→auth_request→Go, відсутність Docker у цій сесії та обов’язковість reviewer-run Compose/e2e гейта.
ФАКТ: кінець work/tasks.md із T16 прочитано, але середину сукупного виводу знову обрізано; весь файл ще не прочитаний без прогалин.
ДАЛІ: дочитати work/tasks.md двома частинами, потім знайти межі §8 plan і прочитати його весь окремо.
ФАКТ: work/tasks.md прочитано повністю; T16 залежить від завершеного T15, має статус planned і володіє рівно шістьма перенесеними e2e-рядками та незакритою клаузулою B36.
ДАЛІ: визначити точні межі §8 docs/playback/plan.md і прочитати секцію без пропусків.
ФАКТ: §8 займає рядки 509–682; прочитано його вступ, датовану поправку рівня доказу, «П’ять питань» і цільові рядки, але один сукупний вивід обрізав частину таблиці B.
ДАЛІ: повторно прочитати §8 трьома частинами, щоб виконати вимогу цілковитого читання без прогалин.
ФАКТ: повторне читання підтвердило датовані поправки, усі п’ять питань, P-таблицю та R-таблицю; ліміт виводу знову сховав лише вузьку середину B-таблиці.
ДАЛІ: прочитати B-таблицю окремим викликом, після чого перейти до handoff і кодових швів.
ФАКТ: §8 docs/playback/plan.md тепер прочитано повністю без прогалин. Цільові очікування дослівно: B20 206/1024 байти; B21 три authorize; B21a відсутність BYTERANGE; B22 точні security/cache headers; B28 edge 404 і видимий стан помилки; B36 503 без медіабайтів; R23 Δtx кожного Go/SSR контейнера менш як 1/10 розміру відео.
ДАЛІ: прочитати work/handoff-phase6.md цілком і зафіксувати рецензентські пастки для дизайну тестів.
ФАКТ: прочитано handoff 1–160. Обов’язкові лінзи: що тест реально доводить; що станеться, якщо продукт кращий за поріг; за яких умов міряли; також не плутати артефакт продакшн-шляху з копією тесту.
ДАЛІ: дочитати handoff 161–468, особливо четверте питання, порожні множини та спостерігачів.
ФАКТ: прочитано handoff 161–320. Четверте питання: «це джерело чи копія, і коли її востаннє звіряли»; для множин обов’язкова непорожність, а runtime самого гейта треба перевіряти окремо від runtime твердження.
РІШЕННЯ: кожен тест T16 починатиме з позитивної опори/непорожності, а вимірювання друкуватимуться до assertion; лічильники, контейнерні ID та байти братимуться з джерел, не з форматованих копій.
ДАЛІ: дочитати handoff 321–468.
ФАКТ: work/handoff-phase6.md прочитано повністю. Додаткові правила: твердження пишеться лише після результату; scope виміру має збігатися зі scope вимоги; не виводити статус гейта з pipeline без pipefail; передумови брифа перевіряються кодом.
ДАЛІ: оглянути playback.spec.ts та згадані інфраструктурні джерела, визначити порядок тестів, наявні хелпери, ассети й безпечні місця для B28/B36/R23.
ФАКТ: playback.spec.ts має один serial describe, глобальну фікстуру з 5-хвилинними й 100-секундними ассетами, openAndPlay, playFromAPIContainer, парсинг master у наявному тесті та P13 біля кінця takeover-блоку.
ДАЛІ: прочитати імпорти/хелпери/фікстуру й увесь тестовий порядок релевантних ділянок, потім зіставити їх з nginx, metrics, compose і transcoder.
ФАКТ: прочитано playback.spec.ts 1–220. Наявні валідатори без unsafe assertions, uploadReady повертає videoId, а всі цільові lesson/video ID є run-scoped; нові хелпери треба тримати в цьому стилі.
ДАЛІ: дочитати хелпери 221–415 і setup/teardown 416–580.
ФАКТ: прочитано playback.spec.ts 221–415. segmentCovering уже доводить status 200, валідний EXTINF і окремий URI; playFromAPIContainer показує безпечний spawn-патерн для docker compose exec; execute доступний для коротких команд.
РІШЕННЯ: додати вузькі хелпери для master/variant/segment, compose stdout, Prometheus series, container ID/tx і паузи до тиші; не дублювати мережеву логіку в кожному тесті.
ДАЛІ: прочитати setup/teardown 416–580 і ранні тести 581–850.
ФАКТ: прочитано setup/teardown 416–580. Фікстура вже транскодує 5 хв для lessonID/thirdLessonID і 100 с для fourthLessonID/crossCourseLessonID; teardown видаляє БД-курси, а тести serial з workers=1.
ДАЛІ: прочитати всі наявні тести та останні споживання кожного ассета, щоб обрати недорогий і безпечний порядок B28/R23/B36.
ФАКТ: прочитано playback.spec.ts 581–850. 100-секундний fourthLesson уже має окреме P06; master-парсинг R35 валідовує дві сходинки; видимий error-текст продукту — «Відео зараз недоступне.».
ДАЛІ: прочитати тести 851–1100, зокрема recovery і P13 takeover.
ФАКТ: прочитано playback.spec.ts 851–1100. Наявний патерн безшумної паузи: виклик video.pause(), очікування кнопки «Відтворити», потім перевірка стабільності heartbeat; O1 уже друкує виміри до assertion.
РІШЕННЯ: для B21 зупиняти video.pause(), чекати успішний pause-heartbeat, а далі poll-ити authorize counters до двох однакових послідовних знімків перед baseline; рахувати delta всіх status-серій, success окремо.
ДАЛІ: прочитати 1101–1350, щоб не перетнутися з takeover/cross-course/replacement сценаріями.
ФАКТ: прочитано playback.spec.ts 1101–1350. P13 уже витягає обслужений .m4s шлях; P14/R26 використовують cross-course asset; B40 замінює primary lesson asset, тому будь-який спільний сегмент треба брати зі свіжого session.mediaPath, не зі старого videoId.
ДАЛІ: дочитати 1351–1553 і визначити справжній кінець serial-порядку для деструктивних тестів.
ФАКТ: playback.spec.ts прочитано повністю. R04 наприкінці student-блоку архівує primary course; instructor/admin сценарії після нього використовують primary asset, а cross-course asset після P14 більше не споживається.
РІШЕННЯ: B28 обирає розвилку (б): використати crossCourseLessonID і видалити його HLS-каталог у самому кінці student-блоку, після нових B20/B21/B21a/B22, R23 і B36; пізніших тестів на цьому ассеті за порядком файлу немає. Ціна: 0 додаткових upload/transcode, лише один play/error цикл; окремий ассет коштував би ще до 240 с очікування транскоду.
РІШЕННЯ: R23 використає наявний 100-секундний cross-course asset і прискорене повне відтворення; очікувана ціна близько 7–15 с при playbackRate 16 замість 100 с реального часу, без нового транскоду.
ДАЛІ: звірити точні контракти nginx headers/503, metrics labels, compose mounts/ports, Redis recovery і HLS disk layout.
ФАКТ: джерела підтверджують: media edge додає exact Cache-Control private, no-store; X-Content-Type-Options nosniff; Referrer-Policy strict-origin-when-cross-origin; X-Frame-Options DENY; Redis failure мапиться у 503; authorize route label дослівно «GET /internal/playback/authorize»; media-worker єдиний із потрібним RW mount; HLS лежить /media/hls/<videoId>/<variant>/segment_*.m4s.
РІШЕННЯ: B36 відновлювати Redis у finally через docker compose start redis і poll того самого справжнього сегмента до статусу !=503; після цього наступний B28 все одно мінтить нову сесію.
ДАЛІ: спроєктувати точні хелпери й тести, перевірити package scripts/style conventions, потім внести єдиний кодовий patch.
ФАКТ: frontend scripts відповідають дозволеним чотирьом командам; playback-проєкт ізольований і serial. У repo вже є shell-патерни парсингу тієї ж Prometheus family, видалення /media/hls/<videoId> через media-worker і start Redis.
ДАЛІ: прочитати ці наявні патерни як джерело точних команд, не змінюючи T14-файлів.
ФАКТ: наявний shell підтверджує безпечну команду rm через media-worker, lazy-sum усіх status series для authorize та readiness polling після docker compose start redis.
ДАЛІ: звірити HTTP status успішного authorize і точний DOM error state, після чого зафіксувати остаточний дизайн перед patch.
ФАКТ: пошук підтвердив named UI state «Відео зараз недоступне.» і локалізаційне джерело; authorize success пише StatusNoContent, але треба прочитати метод повністю для мапінгу відмов.
ДАЛІ: прочитати authorize method і після цього внести patch.
ФАКТ: успішна authorize-series має status="204"; політична відмова — 403, dependency failure — 503. B21 рахуватиме приріст 204 рівно 3 і сумарний приріст усіх інших status рівно 0.
РІШЕННЯ: остаточний порядок нових тестів у student-блоці перед R04: (1) B20/B21/B21a/B22 на обслуженому cross-course сегменті; (2) R23 повністю програє той самий 100-секундний ассет і міряє raw tx_bytes; (3) B36 з finally відновлює Redis та доводить перехід 200→503→не-503; (4) B28 останнім видаляє cross-course HLS і доводить 404 плюс видимий error state. Після B28 жоден тест не використовує cross-course asset.
ДАЛІ: додати хелпери, зберегти crossCourseVideoID з uploadReady і вставити чотири serial e2e-тести.
ФАКТ: додано типізовані хелпери composeOutput, authorizeCounts, waitForAuthorizeSilence, containerNetworkSnapshot і mediaAssetBytes; beforeAll тепер зберігає videoId обох 100-секундних ассетів.
ДАЛІ: вставити чотири цільові e2e-сценарії перед R04, дотримавши порядок Redis recovery → destructive B28.
ФАКТ: вставлено чотири serial e2e-тести. B20/B21/B21a/B22 мають позитивні status/size/set опори; R23 друкує asset/api/frontend raw bytes до асертів і перевіряє container identity; B36 порівнює тіло з реальними медіабайтами та відновлює Redis у finally; B28 видаляє лише cross-course asset, перевіряє mint 201, edge 404 і текст помилки.
ДАЛІ: запустити npm run format:check; у разі стилістичних відхилень виправити лише точковим apply_patch, не npm run format.
ФАКТ: npm run format:check запущено; exit 1, єдиний файл із форматними відхиленнями — e2e/playback.spec.ts. Це очікувана локальна проблема нового patch, не пройдений гейт.
ДАЛІ: отримати formatter diff у stdout без запису та застосувати його через apply_patch.
ФАКТ: oxfmt не має diff/--stdout режиму у CLI help, але підтримує stdin-filepath; форматований результат можна безпечно отримати через stdin без запису.
ДАЛІ: прогнати лише playback.spec.ts через stdin formatter і порівняти з файлом.
ФАКТ: stdin-oxfmt показав лише чотири механічні відхилення: один regex-рядок, два рівні відступу в upload options і один перенос endsWith.
ДАЛІ: застосувати рівно ці чотири форматні зміни через apply_patch і повторити format:check.
ФАКТ: чотири formatter-відхилення виправлено точковим apply_patch; npm run format не запускався.
ДАЛІ: повторити npm run format:check.
ФАКТ: фінальний `npm run format:check` exit 0; 164 файли у правильному форматі.
ДАЛІ: запустити фінальний npm run lint.
ФАКТ: фінальний `npm run lint` exit 0; oxlint без warnings, 91 architecture-boundary файл перевірено.
ДАЛІ: запустити фінальний npm run typecheck.
ФАКТ: фінальний `npm run typecheck` exit 0.
ДАЛІ: запустити фінальний повний npm test після останнього patch.
ФАКТ: фінальний `npm test` exit 0; 45 files, 380/380 passed, 0 failed/skipped/todo, snapshots без змін, 8.261 с.
ДАЛІ: попросити standards reviewer підтвердити закриття єдиного UUID/destructive-path finding без git чи edits.
ФАКТ: standards re-review запущено лише на UUID/destructive-path finding; git та edits заборонені.
ДАЛІ: дочекатися verdict і дописати фінальний підсумок чотирьох груп змін з тим, що ловить стару версію.
ФАКТ: standards re-review підтвердив: canonical lowercase UUID guard стоїть до assetPath, відсікає `/`, `..` та інші path-control символи; залишкових actionable findings немає. Підсумок code-review: Standards — 0 відкритих findings; Spec — 0 findings.

РІШЕННЯ: підсумок другого кола — чотири групи змін і мутація старої версії, яку вони викривають.

| Група | Що змінено | Що впало б на старій/слабкій версії |
|---|---|---|
| R23 | `deliveredBytes`, N of M, terminal proof та tx guards лишено без змін. Поруч із container snapshots додано authorize before/after і друк successful/rejected deltas без асерту. Одразу після `openAndPlay`, до rate 16, знято startPosition і додано named `<=2`. | Новий start assertion червоніє, якщо стан відновлює урок на ~95 с або будь-яку позицію >2: `ended` більше не може видати хвіст за повний шлях від початку. Authorize print навмисно нічого не валить — рішення вимагало лише діагностику кількості медіазапитів за `apiDelta`. |
| B36 | Замість повної нерівності тіл взято first 64 bytes і middle 64 bytes; кожне вікно спершу має `length=64` та exact source index, потім мусить мати `indexOf=-1` у 503 body. Лог містить content-type і length здорової та 503 відповіді. | Стара `equals(false)` зеленіла б на обрізаному сегменті чи segment+prefix; нові `indexOf=-1` падають, якщо 503 містить початок або середину медіа. Source-index asserts падають на порожній/хибно прив’язаний опорі. |
| B28 | Shell спершу доводить canonical UUID, directory, `master.m3u8` і non-empty; видаляє лише вміст; потім доводить directory exists і empty. | Старе `rm -rf "$1"` падає на postcondition `test -d`; asset, якого не було, падає на precondition `test -f master.m3u8`; неповна очистка падає на empty check; traversal id падає до shell-команди. |
| B21 | Exact successful delta `===3` не змінено; додано повідомлення з порядком діагностики: спершу перевірити тишу/hls.js buffering за логом, потім duplicate auth_request. | Нового поведінкового асерту немає за рішенням. Наявний exact `toBe(3)` як і раніше падає на старій регресії `4+` або будь-якому `!=3`; тепер падіння не видає першу можливу причину за доведену. |

ФАКТ: фінальні дозволені перевірки після останнього patch: `npm run format:check` exit 0; `npm run lint` exit 0; `npm run typecheck` exit 0; `npm test` exit 0, **380/380**. `npm run format` не запускався.
ФАКТ: **Docker відсутній і не запускався**; e2e/integration/build/up/seed/smoke/check-generated/full make check не виконані. Друге коло не називає e2e доведеними до reviewer Docker-gate.
ФАКТ: змінено лише `frontend/e2e/playback.spec.ts` і append-only `work/report-t16.md` у межах дозволеного scope.
ФАКТ: git не чіпався — жодної git-команди, staging чи commit.
ПИТАННЯ: немає.
ДАЛІ: рецензенту прогнати Docker e2e, перевірити R23 authorize/traffic prints, B36 content-types/windows і B28 empty-directory state; commit робити лише після зеленого Docker-гейта.
ФАКТ: після UUID guard `npm run format:check` exit 1; єдине відхилення у playback.spec.ts, гейт не зараховано.
ДАЛІ: отримати stdin-formatter diff і застосувати лише механічний перенос через apply_patch.
ФАКТ: formatter diff містить лише перенос довгого UUID regex condition.
ДАЛІ: застосувати цей перенос і повторити format:check.
ФАКТ: formatter-перенос UUID guard застосовано точково; npm run format не запускався.
ДАЛІ: повторити npm run format:check.
ФАКТ: post-review npm run format:check exit 0; 164 файли у правильному форматі.
ДАЛІ: повторити npm run lint.
ФАКТ: post-review npm run lint exit 0; oxlint без warnings, 91 architecture-boundary файл перевірено.
ДАЛІ: повторити npm run typecheck.
ФАКТ: post-review npm run typecheck exit 0.
ДАЛІ: повторити повний дозволений npm test після остаточної зміни.
ФАКТ: post-review npm test exit 0; 45 files, 380/380 passed, 0 failed/skipped/todo, snapshots без змін, 5.656 с.
ДАЛІ: повернути обом review виправлений R23/B21 для короткої read-only перевірки закриття findings; git і edits лишаються заборонені.
ФАКТ: обидва review повторно запущені на вузьку перевірку виправлень без git і edits.
ДАЛІ: дочекатися verdict; після нього дописати остаточний перелік запущених/незапущених гейтів та підсумок без заяви про e2e pass.
ФАКТ: повторний standards review закрив findings R23 і B21; нових actionable standards issues немає. Повторний spec review підтвердив відповідність R23 поправці 2026-09-14; залишкових actionable проблем не знайдено.

РІШЕННЯ: поправка до попередньої таблиці, що має пріоритет над старим R23-рядком: знаменник R23 — **не assetBytes на диску**, а `deliveredBytes`, сума тіл фактичних **200 `.m4s`** responses у measured window. Для кожної реально використаної сходинки лог містить **N унікальних відданих шляхів із M playlist-сегментів**, окремо загальну кількість response bodies та їх суму. Terminal proof: `ended=true`, `duration>0`, `currentTime>=duration-0.25`. Червоне: немає відданих сегментів/сходинки, position не кінцева, container ID змінився, `Δapi<=0`, або `Δ×10>=deliveredBytes` для api/frontend.

ФАКТ: фінальні локальні гейти після всіх змін: `npm run format:check` — exit 0; `npm run lint` — exit 0; `npm run typecheck` — exit 0; `npm test` — exit 0, **380/380**. До виправлень format:check двічі чесно мав exit 1; кожен раз виправлено лише названі рядки через apply_patch, не запуском format.
ФАКТ: **Docker відсутній**. Не запускалися: `make integration`, `make check-generated`, `make build`, `make up`, `make seed-identities`, `make seed`, `make smoke`, `make e2e`, повний `make check`. Жоден із семи e2e-рядків не оголошено доведеним.
ФАКТ: B36 повертає Redis безумовно у `finally` через `docker compose start redis`, а потім до 30 с poll-ить той самий реальний segment URL до status **!=503**; відновлення завершується до наступного B28, який робить новий mint. У цій Docker-less сесії фактичний post-recovery status не спостерігався; e2e друкує його (`200` або policy denial на кшталт `403`) для рецензента.
ФАКТ: B21 тиша: `video.pause()` → успішний pause-heartbeat → три послідовні незмінні знімки total authorize по 500 мс; далі рахуються рівно три 204-authorize та нуль приросту всіх інших status.
ФАКТ: git не торкався: не виконано жодної git-команди, staging чи commit.
ПИТАННЯ: немає; контракт достатній, Discussion не відкривалась. Поправка R23, додана рецензентом під час роботи, виконана.
ДАЛІ: рецензенту прогнати Docker-гейт, прочитати надруковані B21/R23/B28/B36 виміри, за потреби повернути друге коло; лише після зеленого e2e вирішувати статус T16 і commit.
ДАЛІ: git не чіпати — ні staging, ні commit; прочитати implement-інструкцію заново, потім буквально внести п’ять рецензентських уточнень другого кола лише у frontend/e2e/playback.spec.ts і перевірити дозволеними npm-гейтами.
ФАКТ: implement-інструкцію прочитано цілком. Вона вимагає регулярних перевірок, фінального повного suite і code-review.
РІШЕННЯ: вимогу implement про commit не виконувати: користувач прямо заборонив git, а commit робить рецензент лише після недоступного тут Docker-гейта.
ДАЛІ: перечитати поточні хелпери та B21/R23/B36/B28 у playback.spec.ts перед точковим patch.
ФАКТ: знайдено точні межі: helper waitForAuthorizeSilence; об’єднаний B20–B22; R23; B36; B28 безпосередньо перед R04.
ДАЛІ: прочитати суцільно helper і всі чотири цільові блоки, не покладаючись на попередній переказ.
ФАКТ: цільові блоки перечитано. Поточні прогалини збігаються з рішенням рецензента: R23 не знімає authorize counters і не фіксує старт ≤2 с; B36 перевіряє лише повну нерівність тіл; B28 видаляє сам каталог; B21 exact +3 не має діагностичного повідомлення.
РІШЕННЯ: R23 додасть read-only authorize snapshots/дельти лише до логу; startPosition буде знята одразу після openAndPlay і заасерчена до playbackRate. B36 використає два непорожні 64-byte windows з опорними позиціями 0 та middle. B28 shell спершу доведе master/non-empty, видалить лише вміст і доведе directory empty. B21 exact assertion отримає порядок діагностики, не припущення про причину.
ДАЛІ: внести один точковий patch лише у playback.spec.ts.
ФАКТ: patch внесено. R23 тепер друкує successful/rejected authorize deltas за тим самим playback window і має named startPosition ≤2 до rate 16; B36 перевіряє перше та середнє 64-byte windows після self-check опорних позицій і друкує обидва content-type; B28 лишає каталог, довівши master/non-empty до очистки та directory/empty після; B21 exact +3 має двоступеневу діагностику.
ДАЛІ: запустити npm run format:check; npm run format не запускати.
ФАКТ: друге коло `npm run format:check` exit 0; 164 файли у правильному форматі. `npm run format` не запускався.
ДАЛІ: запустити npm run lint.
ФАКТ: друге коло `npm run lint` exit 0; oxlint без warnings, architecture boundaries перевірено для 91 файла.
ДАЛІ: запустити npm run typecheck.
ФАКТ: друге коло `npm run typecheck` exit 0; tsc --noEmit прийняв нові Buffer windows, shell string і лічильникові дельти.
ДАЛІ: запустити повний дозволений npm test.
ФАКТ: друге коло `npm test` exit 0; 45 files, 380/380 passed, 0 failed/skipped/todo, snapshots без змін, 5.762 с.
ФАКТ: Docker не запускався, тому нові e2e assertions у цьому середовищі не виконані й не оголошені доведеними.
ДАЛІ: виконати обов’язковий read-only code-review після змін, без git: спершу перечитати інструкцію скіла code-review.
ФАКТ: code-review skill прочитано цілком. Його штатні fixed-point git-команди неможливі через пряму заборону користувача.
РІШЕННЯ: виконати два паралельні read-only огляди поточних B21/R23/B36/B28 блоків: standards і відповідність рішенню другого кола; обом агентам прямо заборонити git та edits.
ДАЛІ: запустити два review й дочекатися обох verdict.
ФАКТ: запущено standards і spec review другого кола; обом заборонено git та файлові зміни.
ДАЛІ: дочекатися обох результатів, виправляти лише підтверджений blocking finding.
ФАКТ: spec review — 0 findings, усі вимоги другого кола виконані без scope creep. Standards review знайшов одну medium проблему: destructive B28 будує rm path із зовнішнього videoId, валідованого лише як string; execFile прибирає shell injection, але не slash/`..` traversal.
РІШЕННЯ: finding підтверджений правилом AGENTS.md «Validate external data» і destructive guardrails. Перед assetPath додати canonical lowercase UUID check; malformed API value має впасти до shell-команди.
ДАЛІ: додати вузьку валідацію B28, потім повторити чотири npm-гейти й re-review цього finding.
ФАКТ: B28 тепер відмовляється від будь-якого videoId, що не є canonical lowercase UUID, до побудови `/media/hls/<id>` і до destructive shell-команди.
ДАЛІ: повторити npm run format:check.
ФАКТ: npm run format:check повторно запущено; exit 0, 164 файли відповідають формату.
ДАЛІ: запустити npm run lint і виправити лише фактичні помилки нової зміни.
ФАКТ: npm run lint запущено; exit 0, oxlint без warnings, architecture boundaries перевірено для 91 файла.
ДАЛІ: запустити npm run typecheck; це має перевірити сигнатури promisified execFile, bigint assertions і Playwright API.
ФАКТ: npm run typecheck запущено; exit 0, tsc --noEmit прийняв усі нові типи, bigint та Playwright виклики.
ДАЛІ: перед npm test провести статичний self-review цільового diff без git: перечитати нові хелпери/тести й перевірити сім вимог, destructive order, shell quoting та можливі always-green умови.
ФАКТ: символи нової зміни знайдено; B28 стоїть безпосередньо перед R04, а R04 та наступні instructor/admin тести не використовують cross-course asset.
ДАЛІ: перечитати нові хелпери та тести суцільними діапазонами.
ФАКТ: статичний self-review пройдено: сім вимог мають окремі falsifiable assertions; усі response checks починаються зі status, усі множини — з непорожності; B21 вікно відділене паузою й стабільним counter; B28 directory absence перевіряється самою командою; shell-аргументи передають path/videoId позиційно.
РІШЕННЯ: код не пом’якшувати до >=3, header presence, порожнього playlist, docker stats чи «body length !=0»; поточні exact checks відповідають брифу.
ДАЛІ: запустити дозволений повний npm test; він не виконає нові e2e, але перевірить увесь локально доступний frontend unit suite.
ФАКТ: npm test запущено; exit 0, 45 test files, 380/380 tests passed, 0 skipped/todo, snapshots без змін, 4.871 с.
ФАКТ: нові сім e2e-рядків цим не виконані й не доведені; Docker у середовищі відсутній, тому їх перевіряє рецензент.
ДАЛІ: виконати обов’язковий code-review скіла implement у read-only режимі без git, спершу прочитавши інструкцію скіла цілком.
ФАКТ: code-review skill прочитано цілком; штатний процес вимагає git fixed-point/diff, але це прямо заборонено §4 брифа.
РІШЕННЯ: адаптувати review без git: два незалежні read-only огляди поточного playback.spec.ts — standards та spec — з прямою забороною агентам виконувати git чи редагувати файли.
ДАЛІ: запустити обидва огляди паралельно й перевірити їхні findings перед фінальними гейтами.
ФАКТ: запущено два незалежні read-only review: standards і spec; обом заборонено git та файлові зміни.
ДАЛІ: поки review працюють, перевірити цільові очікування числом і підготувати обов’язкову таблицю п’яти питань для звіту без заяви про e2e pass.
РІШЕННЯ: таблиця «рядок → чим доводиться»; колонки Q1–Q5 відповідають п’яти питанням §8 плану.

| Рядок | Q1 — що саме доводить / очікування числом | Q2 — походження артефакту | Q3 — захист від зеленого з хибної причини | Q4 — чому саме цей рядок | Q5 — що робить тест червоним |
|---|---|---|---|---|---|
| B20 | Реальний дозволений `.m4s` довший за 1024; `Range: bytes=0-1023` → **206**, тіло **1024**, `Content-Range=bytes 0-1023/N`. | Сегмент узято з фактичної 200-відповіді hls.js; asset пройшов UI upload → worker ffmpeg → publish. | Спершу full **200** і `N>1024`, тому короткий/відсутній файл не зелений. | Єдиний тут предмет — byte range contract B20. | Range перестає бути 206, повертає не 1024 або неправильну межу/total. |
| B21 | Після тиші три різні range-запити дають **3×206**; `authorize{status="204"}` Δ=**3**, усі інші status Δ=**0**. | Той самий обслужений prod-generated сегмент; метрика читається прямо з api `/metrics` у Compose network. | Серія після позитивної події мусить існувати; baseline береться після pause-heartbeat і двох однакових знімків через 500 мс; route label дослівно з mux. | Рахується ціна рівно трьох range-запитів, не загальна властивість кешу чи B13. | Будь-який запит обходить/подвоює auth, відхиляється, або фоновий плеєр не вдається вгамувати за 15 с. |
| B21a | Master має **2** variant URI; кожен playlist → **200**, має `EXTINF` (**>0 URI**), усі segment URI окремі/унікальні, `BYTERANGE` **0**. | Обидва `index.m3u8` згенеровані worker ffmpeg для опублікованого ассета. | Порожній master/playlist не проходить; `EXTINF` без наступного окремого URI не проходить. | Перевіряється саме файлове компонування плейлистів, не Range поведінка edge. | Кодер додає `single_file`/BYTERANGE, губить сходинку, EXTINF або повторно вживає URI. |
| B22 | Дозволений сегмент спершу **200**; exact headers: `private, no-store`, `nosniff`, `strict-origin-when-cross-origin`, `DENY`. | Відповідь edge на реальний опублікований сегмент, не 403-заглушка. | Status 200 стоїть перед header checks; порівнюються значення, не presence. | Це прямий контракт shared-cache/security headers B22. | Будь-яке точне значення змінене/відсутнє або сегмент не 200. |
| B28 | Для DB-ready asset каталог справді відсутній; mint → **201**, mediaPath несе його videoId; master від edge → **404**; видно текст «Відео зараз недоступне.». | Cross-course asset спершу пройшов штатний upload/transcode; файли прибрані RW media-worker з `/media/hls/<videoId>`. | Не вигаданий id: `uploadReady` зберіг videoId; `rm` одразу перевіряє `! -e`; mint 201 доводить ready-read path. | Поєднано edge 404 і названий UI error саме B28. | Файли не зникли, mint не 201, edge не 404 або UI не показує стан. |
| R23 | Наявний **100 с** asset дограє до `ended` при rate **16**; `assetBytes>0`; container IDs незмінні; `Δapi>0`; `Δapi×10<assetBytes`, `Δfrontend×10<assetBytes`. | Розмір — сума `wc -c` усіх файлів published HLS на media-worker; tx — raw kernel `tx_bytes` api/frontend. | До ratios друкуються всі числа; ID відсікає recreate; позитивний api delta відсікає нульовий/від’ємний фальш-вимір; frontend Δ=0 дозволено. | Вимірює topology A9, а не response headers чи browser transfer size. | Байти підуть через Go/SSR ≥10%, api не подасть heartbeat bytes, контейнер зміниться або lesson не завершиться. |
| B36 | Той самий сегмент: healthy **200**, `N>0`; Redis stopped → **503**; тіло не дорівнює N медіабайтам; recovery poll → status **!=503**. | Опора — байти реальної 200-відповіді prod-generated asset; failure інʼєктовано в справжній Redis Compose service. | Не невідомий playbackId і не перевірка `length!=0`; порівнюються actual bodies того самого URL. | Smoke вже володіє 503/Retry-After; тут рівно перенесена клаузула «жодного медіа». | Під outage просочаться опорні bytes/не буде 503, або finally не поверне policy path за 30 с. |

РІШЕННЯ: B28 — розвилка (б): cross-course asset знищується після всіх його споживачів; ціна 0 нових транскодів і близько одного error/retry циклу до 60 с. R23 — наявний 100-секундний fourth-lesson asset, rate 16; очікуване додавання до playback близько **7–15 с**, без нового upload/transcode.
ДАЛІ: дочекатися двох незалежних review і виправити лише підтверджені blocking findings.
ФАКТ: обидва review завершилися. Спільний blocking finding: R23 denominator all on-disk HLS files може бути завеликим; обидва посилаються на корекцію брифа 2026-09-14 про суму реально відданих 200 `.m4s`, якої не було в прочитаному на старті тексті. Standards review також просить terminal position/duration і довше вікно тиші B21.
РІШЕННЯ: за четвертим питанням не приймати переказ як джерело; спершу повторно звірити поточний work/brief-t16.md на згадані поправки й лише тоді змінювати код.
ДАЛІ: знайти в поточному брифі всі згадки denominator, N of M, terminal position і quiet window.
ФАКТ: поточний work/brief-t16.md справді має нову «Поправку 2026-09-14, R23 — знаменник» на рядках після первісного кінця; файл змінився під час роботи. Вона скасовує on-disk denominator і вимагає реально віддані 200 segment bodies, N із M по сходинці та terminal position.
РІШЕННЯ: поправка є сильнішою й актуальною; R23 треба переробити. B21 «перестане рости» не задає тривалість, але одна пара 500 мс слабка — посилити до трьох послідовних стабільних інтервалів (1.5 с) без зміни exact +3.
ДАЛІ: прочитати нову поправку брифа цілком і буквально перед patch.
ФАКТ: поправку прочитано повністю. R23 тепер збирає тіла лише 200 `.m4s` response під час measured playback, групує унікальні paths за реально зіграною сходинкою, читає M з її playlist після tx snapshot, друкує N of M, response count, суму bytes і terminal currentTime/duration до асертів.
ФАКТ: mediaAssetBytes видалено; ratios тепер мають denominator deliveredBytes. B21 тиша посилена до трьох послідовних незмінних 500-мс інтервалів у межах 15 с.
РІШЕННЯ: якщо R23 покаже N<M, «цілком» означає, що session пройшов урок від початку до terminal position/ended; це не твердження, що browser завантажив усі M сегментів при rate 16.
ДАЛІ: повторити всі чотири дозволені локальні гейти після review-fix, починаючи з format:check.
ФАКТ: post-review npm run format:check exit 1; єдиний файл — playback.spec.ts, тож потрібна лише механічна поправка нового R23.
ДАЛІ: отримати stdin-formatter diff і застосувати його через apply_patch.
ФАКТ: formatter diff має рівно два переноси рядків у R23: reduce і page.request.get.
ДАЛІ: застосувати ці два переноси та знову запустити format:check.
ФАКТ: два formatter-переноси застосовано точковим patch.
ДАЛІ: повторити npm run format:check.
