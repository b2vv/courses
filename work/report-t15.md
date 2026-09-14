ДАЛІ: Прочитаю бриф T15, репозиторний briefing і інструкції implement, після чого перевірю поточний стан та виконаю таск буквально без будь-яких git-операцій.

ФАКТ: `work/report-t15.md` до старту не існував; створив його з обов’язковим першим рядком. Бриф T15 прочитано повністю. Він вимагає лише тести B08, B31, R34, забороняє чіпати B07/T14/T16 і Docker-гейти та прямо скасовує вимогу `implement` про commit. Сукупний вивід решти великих документів обрізався, тому дочитаю потрібні джерела окремими порціями перед змінами.

РІШЕННЯ: Git не використовуватиму взагалі. Для R34 попередньо схиляюсь до варіанта (б) — перепризначити вже транскодований архівований курс через повний PUT, якщо аналіз порядку P14/R04 підтвердить відсутність залежності від старого власника; це уникає ще одного 100-секундного транскоду.

ДАЛІ: Дочитаю briefing, §8 плану, handoff і точні фрагменти тестів/фікстур, перевірю ризик порядку для R34 та визначу конкретні тестові зміни.

ФАКТ: Контракти рядків підтверджено: B08 — integration і вимагає збереженого `completed`; B31 — e2e з трьома окремими спостереженнями після повторного play на A; R34 — e2e сегмент 200 для власника-інструктора. `learning_integration_test.go` вже містить окремі B07-фікстури та перевірки, їх не змінюватиму. Інструкторська login-фікстура має ID `40000000-0000-4000-8000-000000000002`; UI справді дає `playAgain` у стані `taken-over`.

ДАЛІ: Прочитаю повний §8 з п’ятьма питаннями, релевантні секції handoff і весь e2e-контекст P11/P13/фікстури, потім остаточно зафіксую варіант R34 та оцінку часу до реалізації.

ФАКТ: П’ять питань §8 прочитано: предмет доказу, production-походження артефакту, можливість хибно-зеленого результату, точна прив’язка до рядка, конкретна мутація для почервоніння. CTO briefing прочитано повністю; він підтверджує account-wide ownership, edge-докази лише через реальні байти й відсутність Docker. Сукупний вивід handoff/e2e знову обрізався, тож дочитаю лише точні діапазони окремо.

ДАЛІ: Локалізую заголовки handoff, helper-и e2e, точний порядок тестів та повну форму course PUT; тоді зафіксую остаточне рішення R34 до зміни коду.

ФАКТ: E2E-проєкт серійний і має спільний run-scoped курс: три вже транскодовані відео, повне дерево з двох модулів/чотирьох уроків, окремий cross-course. R04 архівує основний курс наприкінці student-блоку; P11 після цього грає його як адмін. Отже R34 природно стає після R04 і до/поруч із P11: можна змінити власника вже архівованого основного курсу, не зачіпаючи P14 (він раніше) й не створюючи нового транскоду.

ДАЛІ: Дочитаю точні hook-и й handoff-питання, перевірю формат повного PUT та спосіб автентифікувати другого не-адміна. Після цього запишу остаточне рішення R34 з числовою ціною до редагування.

РІШЕННЯ: Для R34 обираю варіант (б): після R04 адмінським `PUT` перепризначити вже архівований основний курс інструктору `40000000-0000-4000-8000-000000000002`, передавши незмінні slug і повне дерево з тими самими module/lesson ID. P14 виконується раніше, R04 уже завершив свою перевірку, а наступний P11 є адмінським і не залежить від власника. Очікувана додаткова ціна R34: приблизно 5 с e2e, з них 0 с транскодування; варіант (а) додав би щонайменше один транскод 100-секундного файла й окрему фікстуру.

ФАКТ: Контроль для R34 робитиме прямий mint тим самим student-акаунтом на тому самому `lessonID` і вимагатиме 403; це перевіряє саме недоступність архіву чужому не-адміну, а не лише неможливість повторно використати owner playback ID. Owner-прохід окремо чекає реальний `.m4s` зі статусом 200.

ДАЛІ: Додам окрему B08-фікстуру й integration-перевірку, розширю P13 до повного B31 та додам R34 з централізованим повним деревом основного курсу. B07-код і заборонені файли не змінюватиму.

ФАКТ: B08 додано окремим integration-сценарієм і окремою фікстурою `learning-b08`: generation 1 готова до перегляду, позначка `completed_at` стоїть між поколіннями в часі, generation 1 витіснена, generation 2 готова й активна. Читальна модель мусить повернути один урок з `Completed=true`. Жодного рядка B07 не переписано.

ДАЛІ: Винесу визначення основного e2e-курсу/дерева в одне джерело всередині тесту, розширю P13 доказом B31 і додам R34 після архівації.

ФАКТ: E2E-зміни внесено. Основний course payload і повне дерево тепер мають одне джерело. P13 розширено до B31: після B takeover перевіряється його сегмент 200; A натискає `Відтворити знову`, отримує нову сесію, показує `Пауза` і читає той самий media-сегмент через новий playback ID зі статусом 200; B переходить у `taken-over`, а його щойно дозволений сегмент стає 403. R34 після R04 повним PUT змінює лише власника архівованого курсу, інструктор запускає реальне HLS і отримує `.m4s` 200, а student-контроль на тому самому lesson дістає 403 при mint.

ДАЛІ: Перевірю форматування, типи й локальні контракти змінених файлів; спершу запущу лише дозволені точкові перевірки без Docker.

ФАКТ: `gofmt -d backend/internal/course/adapters/postgres/learning_integration_test.go` не показав різниці. `npm run format:check` пройшов: 164 файли, exit 0.

ДАЛІ: Запущу дозволені `npm run lint` і `npm run typecheck`; за їх результатом виправлю лише помилки цього слайса.

ФАКТ: `npm run lint` пройшов, включно з architecture boundaries для 91 файла; exit 0.

ДАЛІ: Запущу `npm run typecheck`.

ФАКТ: `npm run typecheck` пройшов; exit 0.

ДАЛІ: Запущу `npm test`, потім Go-тести без integration build tag і без Docker.

ФАКТ: `npm test` пройшов: 45 файлів, 379/379 тестів, 0 skipped/todo, exit 0, 5.395 с.

ДАЛІ: Запущу `go test -race ./...` у backend без integration-тега; B08 при цьому не виконується, бо він має build tag `integration` і потребує Docker.

ФАКТ: Перший `go test -race ./...` не завершив гейт: sandbox заборонив запис у системний Go build cache (`/Users/strelia/Library/Caches/go-build`), exit 1. Частина пакетів пройшла, але сукупний результат не зараховую.

РІШЕННЯ: Це не Docker-відмова і не помилка коду; повторю той самий тест з `GOCACHE` у дозволеному `/private/tmp`, не змінюючи репозиторій.

ДАЛІ: Повторно запущу `go test -race ./...` з доступним build cache.

ФАКТ: Повторний `go test -race ./...` з writable `GOCACHE` виконав усі доступні пакети, але сукупний exit 1 через sandbox-відмову `bind: operation not permitted` у `courses/platform/httpserver.TestShutdownDrainsAnInFlightResponse`. Решта показаних пакетів пройшла; це обмеження локального socket bind, не Docker і не продуктова помилка.

РІШЕННЯ: Сукупний Go-гейт не називатиму зеленим. Щоб відокремити відмову sandbox від решти коду, запущу всі Go-пакети без єдиного `platform/httpserver`; integration-tag як і раніше не вмикаю.

ДАЛІ: Запущу race-тести всіх Go-пакетів, крім заблокованого socket-пакета, й зафіксую точний результат.

ФАКТ: `go test -race` для всіх Go-пакетів, крім sandbox-заблокованого `courses/platform/httpserver`, пройшов; exit 0. Повний Go-гейт лишається незарахованим через неможливість bind у sandbox.

ДАЛІ: Проведу структурний review змін без git: перевірю точні змінені фрагменти, B08-мутацію, порядок R04→R34→P11 і те, що заборонені B07/T14/T16 файли не редагувалися. Для цього прочитаю інструкцію `code-review`, викликану `implement`, але не виконуватиму жодних git-команд.

ФАКТ: `code-review` зазвичай вимагає git fixed-point і два незалежні огляди. Fixed-point не задано, а §4 брифа забороняє будь-який git, тому його git-процедура непридатна буквально.

РІШЕННЯ: Збережу сутність skill без порушення брифа: два незалежні read-only огляди — Standards проти `AGENTS.md` і smell baseline, Spec проти `work/brief-t15.md` — лише для двох відомих змінених кодових файлів; журнал не оцінюється як продуктова зміна. Жодних git-команд.

ДАЛІ: Запущу обидві осі review паралельно, потім сам перевірю знахідки й виправлю лише підтверджені.

ФАКТ: Два незалежні read-only review запущено без git: Standards перевіряє правила репо й smells, Spec — буквальну відповідність §2/§3/§4/§6 брифа. Обом заборонено редагування.

ДАЛІ: Поки review триває, перевірю компіляцію integration-файлу без запуску Docker-тесту та точний локалізований label `playAgain`; це не буде зараховано як доказ B08/B31/R34.

ФАКТ: `go test -tags=integration ./internal/course/adapters/postgres -run '^$'` успішно скомпілював integration-пакет без запуску тестів; exit 0. Перевірка локалі виявила помилковий текст кнопки в новому B31: фактичний `uk` label — `Почати перегляд знову`, а не `Відтворити знову`. Тест виправлено на джерельне значення.

ДАЛІ: Дочекаюся двох review, перевірю їх знахідки, потім повторю frontend format/lint/typecheck після виправлення.

ФАКТ: Spec-review не знайшов пропущених, часткових, зайвих або неправильно реалізованих вимог; окремо підтвердив хронологію B08, усі три спостереження B31 і безпечний порядок R04→R34→P11 з повним деревом. Standards-review не знайшов порушень правил репо; одна слушна smell-знахідка — назва наявного P13-тесту не містила `B31`, хоча тепер саме його доводить.

РІШЕННЯ: Додав `B31` до назви e2e-тесту для прямої row-level трасованості. Інших змін за review не потрібно.

ДАЛІ: Повторю `npm run format:check`, `npm run lint`, `npm run typecheck` і `npm test` після останніх правок; Docker-залежні гейти не запускатиму.

ФАКТ: Після review всі чотири frontend-гейти повторно пройшли в порядку: format-check exit 0 (164 файли), lint exit 0 (91 architecture files), typecheck exit 0, unit tests 379/379 exit 0 (6.745 с).

РІШЕННЯ: Для B08 додатково зафіксував precondition самого replacement: SQL-перевірка вимагає evicted generation 1 та активну generation 2. Без неї тест міг би лишитися зеленим після випадкового спрощення фікстури до одного відео й уже не доводив би «після заміни».

ДАЛІ: Перекомпілюю integration-пакет після посилення B08, перевірю форматування Go і потім допишу фінальну таблицю доказів, мутацію, числові очікування та список недоступних гейтів.

ФАКТ: Після посилення B08 `gofmt -d` чистий, а integration-пакет з build tag компілюється без запуску тестів; exit 0. Під час відповіді на handoff-питання «що, якщо продукт зробить краще» знайшов зайве звуження R34-контролю до 403.

РІШЕННЯ: R34-контроль тепер приймає обидві коректні відмови 403 або privacy-preserving 404; 5xx і успішний mint 201 залишаються червоними. Це зберігає вимогу «інший не-адмін не отримує доступ», не прив’язуючи acceptance до способу приховування ресурсу.

ДАЛІ: Повторю frontend format/lint/typecheck після цієї однорядкової зміни, потім додам фінальну доказову таблицю й статус гейтів.

ФАКТ: Фінальний повтор після R34-контролю пройшов: `npm run format:check` exit 0 (164 файли), `npm run lint` exit 0 (91 architecture files), `npm run typecheck` exit 0.

## Підсумок доказів T15

| Рядок | 1. Що саме доводить | 2. Походження артефакту | 3. Як виключено хибно-зелений | 4. Який саме рядок закриває | 5. Що мусить зламатись для red |
|---|---|---|---|---|---|
| B08 | Після заміни відео learning read повертає `Completed=true` для раніше переглянутого уроку. | Integration-тест піднімає міграції в ізольованому PostgreSQL і читає через справжній `GetLearningCourse`; fixture напряму задає стан у production-таблицях, бо рядок перевіряє projection, а не upload/transcode path. Старе й нове покоління мають реальні поля generation/eviction/current. | Окрема SQL-precondition вимагає evicted generation 1 і активну generation 2; окремий assert вимагає `Completed=true`. Ні B07, ні інша властивість уроку не можуть дати цей зелений результат. | Лише B08; B07-сценарії та код не змінені. | Замінити `now()-interval '2 hours'` на `NULL` саме в `completed_at` рядка `learningB08Lesson`: очікуване падіння `B08 replacement completed=false want true`. Видалення replacement окремо валить fixture-precondition. |
| B31 | Після takeover B→A: A реально знову грає, слот належить A, B зупинений. | Початковий `.m4s` приходить із реального HLS, створеного наявним e2e upload/transcode path; для того самого segment suffix змінюється лише playback ID із фактичних mint-відповідей. | До retake B-сегмент явно 200. Після `playAgain` три незалежні результати: A має `Пауза`; A-сегмент 200; B має `taken-over` і його раніше дозволений сегмент 403. | Тест прямо названо `P13 B15 B31 B32`; нова retake-частина — B31, старі before/after assertions лишають P13/B15/B32. | Якщо retry не мінтить нову сесію, A не починає грати, новий A playback ID не авторизований або heartbeat B не визнає displacement, падає відповідний окремий assert. |
| R34 | Власник-інструктор, який не є адміном, грає архівований курс і отримує реальний `.m4s` 200; інший не-адмін не може mint-ити той самий урок. | Використано той самий основний курс і байти, які створив e2e upload/transcode path. Після R04 адмінський production PUT передає незмінний slug і повне спільне дерево з усіма тими самими ID, змінюючи лише `instructorId`. | Owner `.m4s` 200 сам по собі міг би означати «архів відкритий усім», тому в тому самому тесті student-контроль мусить отримати коректну відмову 403/404 на mint; 201 і 5xx падають. | Лише R34; P14 і R04 виконані до mutation, P11 після неї лишається валідним через admin policy. | Якщо owner predicate прибрати, instructor не отримає session/segment 200; якщо архів відкрити всім, student mint поверне 201 замість 403/404. |

### Чотири питання handoff

- B08: краща поведінка щодо позиції чи нового ассета не завадить — тест вимагає лише незнищену позначку. Умови: PostgreSQL 17.5 testcontainer після всіх міграцій; тут лише скомпільовано, не виконано. Джерело — repository read; fixture є контрольованою передумовою, не копією відповіді.
- B31: швидше витіснення або швидший restart не завадять, бо немає вимоги до точного часу/кількості запитів. Умови reviewer-gate: Compose, реальний nginx/media-worker HLS, стиснений heartbeat/TTL 2/6. Джерела — browser-visible label, фактичні mint-відповіді й edge-відповіді, не fixture-копія очікуваного статусу.
- R34: privacy-preserving 404 чужому користувачу не завадить; приймаються 403/404, але не 5xx. Умови reviewer-gate: той самий Compose/HLS після серійного R04. Джерела — повний course tree має одне спільне визначення для create/PUT, owner segment береться з реальної browser response.

### R34 — ціна розвилки

Обрано (б), перепризначення вже транскодованого курсу. Очікувана додаткова ціна: приблизно **5 с e2e**, **0 с транскодування**, один PUT. Варіант (а) відхилено: він додає окремий курс, upload і щонайменше один транскод 100-секундної фікстури без сильнішого доказу.

### Гейти

Запущено:

- `npm run format:check` — фінально exit 0, 164 файли;
- `npm run lint` — фінально exit 0, architecture boundaries 91 файл;
- `npm run typecheck` — фінально exit 0;
- `npm test` — exit 0, **379/379**, 45 файлів, 0 skipped/todo;
- `gofmt -d` зміненого Go-файлу — чисто;
- `go test -tags=integration ./internal/course/adapters/postgres -run '^$'` — лише компіляція, exit 0, tests not run;
- `go test -race` для всіх пакетів, крім `courses/platform/httpserver` — exit 0.

Не пройдено/не запускалось:

- повний `go test -race ./...` — **не пройдено**, sandbox забороняє `bind` у `platform/httpserver`; не видаю за green;
- `make integration`, `make build`, `make up`, seed-и, `make smoke`, `make e2e`, повний `make check` — **не запускались: Docker відсутній**;
- `make check-generated` — не запускався за §4 брифа та через його git/index-залежність;
- `npm run format` — навмисно не запускався, заборонено брифом;
- Git — не використовувався взагалі: ні status/diff, ні staging, ні commit.

### Числове очікування reviewer-gate

- B08: тут виконано **0/1**; очікую **1/1** integration-сценарій green. Контрольна мутація має дати **1/1 red** з точним текстом вище.
- B31: тут виконано **0/1**; очікую **1/1** e2e-тест green і **3/3** заявлені спостереження (A грає; A segment 200; B taken-over + segment 403).
- R34: тут виконано **0/1**; очікую **1/1** e2e-тест green, owner segment **200**, чужий mint **403 або 404**.
- За базовими 60 e2e-тестами B31 не додає нового test case, а R34 додає один: очікування всього suite — **61/61**, якщо інші паралельні слайси не змінять лічильник.

ПИТАННЯ: Немає. Для закриття B08/B31/R34 потрібен другий, Docker-доступний reviewer-gate; у цій сесії всі три лишаються очікуваннями, а не доведеними рядками.

ДАЛІ: Друге коло T15. Перевірю компонентну обв’язку `LessonPlayer`, відтворю дефект швидким frontend-тестом, виправлю життєвий цикл медіа для поточного `<video>` без послаблення B31 і прогоню лише дозволені frontend-гейти. Git не чіпаю — ні staging, ні commit; коміт робить рецензент після недоступного тут Docker-гейта.

ФАКТ: Інструкцію `implement` перечитано. Її вимога commit прямо скасована рішенням користувача 2026-09-14; вимоги TDD, регулярних точкових перевірок і завершального review лишаються чинними.

ДАЛІ: Перечитаю актуальний repository briefing, повний `LessonPlayer`, наявні компонентні тести та test doubles для media, щоб вибрати найменший життєвий механізм і спочатку написати regression-тест.

ФАКТ: `work/CTO-RESEARCH.md` прочитано повністю. Для цієї правки критичні незмінні контракти: рішення 403 ухвалює heartbeat, ownership account-wide, а e2e працює на 2/6 замість production 30/180. Docker-гейт і реальні HLS-байти лишаються відповідальністю рецензента.

ДАЛІ: Прочитаю весь `LessonPlayer.tsx`, його component tests і playback media port/fakes; окремо звірю наявні тести R38, quality, heartbeat та restore, які мають лишитися зеленими.

ФАКТ: Компонентна обв’язка існує: `LessonPlayer.test.tsx` уже інʼєктує `CreatePlaybackMedia` і керує callbacks. Дефект підтверджено в коді: `videoRef` не є реактивною залежністю; terminal failure розмонтовує `<video>`, але media-effect не очищується й не створює адаптер для нового вузла. Додавати `failure` до dependencies ризиковано для безпомилкових remint-гілок; природний seam — callback ref зі станом поточного `HTMLVideoElement`.

РІШЕННЯ: Використаю стан `videoElement` через стабільний callback ref. Ефект створення media залежатиме від конкретного DOM-вузла і cleanup знищуватиме адаптер при callback-ref `null`; `videoRef` лишиться імперативним ref для heartbeat/controls. Спочатку додам component regression-тест і зафіксую його red на поточному продукті.

ДАЛІ: Прочитаю повну test harness і port, визначу точні наявні регресійні тести для R38, quality, heartbeat та restore, потім додам лише тест без продуктової правки.

ФАКТ: Повну component harness прочитано. Наявні захисти: quality/keyboard — `B24 keeps play, seek and hls.js quality...`; restore та стабільність адаптера — `keeps one media adapter across time updates and restores...`; безпомилковий 410 remint/heartbeat — `continues heartbeats after an expired session is silently reminted`; 403-verdict R38 — `recovery.test.ts` плюс media/error сценарії; heartbeat cadence/exits — B09, B-7, B10, R28.

ФАКТ: Додано лише regression-тест `B31 recreates media on the remounted video before retrying after takeover`. Він створює першу сесію, отримує 409 heartbeat, бачить terminal state, натискає retry і вимагає: другий media adapter створено для нового DOM `<video>`, старий adapter знищено, `play()` викликано на першому та новому елементах саме в такому порядку.

ДАЛІ: Запущу тільки новий B31 component-тест на ще невиправленому продукті й зафіксую точний red.

ФАКТ: Перший red-прогін справді впав, але не на продуктовому discriminant: test timeout 5000 ms, `no expect assertions completed`. Причина — `findByRole` очікував під fake timers після вже синхронно завершеного state transition. Цей red не зараховую як доказ дефекту.

РІШЕННЯ: Заміню лише асинхронний пошук heading на синхронний `getByRole` після `act`; продукт лишається невиправленим. Потім повторю red, який мусить дійти до лічильника media adapters.

ДАЛІ: Виправлю test harness і повторно запущу один B31-тест до продуктової зміни.

ФАКТ: Test harness виправлено: після контрольованого `advanceTimersByTimeAsync` heading перевіряється синхронно. Product code досі без змін.

ДАЛІ: Повторю targeted B31 component-тест і зафіксую його продуктовий failure text.

ФАКТ: Коректний red отримано на незміненому продукті: targeted test exit 1, `expected [ <video …> ] to have a length of 2 but got 1` на `LessonPlayer.test.tsx:607`. Отже retry справді монтує новий `<video>`, але `createMedia` лишається викликаним лише один раз для старого вузла; це точно той самий дефект, що знайшов reviewer e2e.

ДАЛІ: Додам стабільний callback ref, який синхронізує імперативний `videoRef` зі станом `videoElement`; media-effect залежатиме від цього елемента, а JSX використовуватиме callback ref. `failure` до dependencies не додаватиму, щоб 403/410 remint без розмонтування не руйнував media.

ФАКТ: Продуктову правку внесено лише в `LessonPlayer.tsx`: `attachVideo` записує поточний DOM-вузол і в `videoRef`, і в реактивний `videoElement`; media-effect створює адаптер для `videoElement` і має його в dependencies; `<video>` використовує callback ref. Cleanup того самого effect викликає `destroy()` і обнуляє `mediaRef` при зміні/зникненні вузла.

РІШЕННЯ: `failure` навмисно не є залежністю. Terminal error розмонтовує вузол і тому гарантовано змінює `videoElement`; натомість remint у режимах `expired`/`missing`, де `<video>` залишається тим самим, не спричиняє зайвого destroy/recreate і зберігає R38/O1 semantics.

ДАЛІ: Запущу один B31 component-тест після правки; очікую 1/1 green із другим adapter, destroy старого та `play()` нового вузла.

ФАКТ: Перший green-кандидат лишився red, але вже звузив дефект: `createdVideos` має 2 і `destroyedVideos` містить старий вузол, проте `playedVideos` містить лише перший. Failure: `expected [<video>] to deeply equal [<video>, <video>]` на рядку 609.

ФАКТ: Причина в другому реактивному шві: retry-effect залежить лише від `[loadSession, restartAt]`. Він один раз бачить новий `restartAt`, але media для callback-ref вузла ще не створено; наступний render із новим `videoElement` його не перезапускає.

РІШЕННЯ: Додам `videoElement` у dependencies retry-effect. Це не стосується `expired`/`missing`: вони викликають `loadSession` без `restartAt`, а при звичайному житті стабільний DOM-вузол не змінюється.

ДАЛІ: Додам залежність поточного елемента до retry-effect і повторю targeted B31.

ФАКТ: Retry-effect тепер залежить від `videoElement`; після callback-ref transition він повторно перевіряє вже створений `mediaRef` і лише тоді запускає explicit retry mint/load.

ДАЛІ: Повторно запущу один B31 component-тест.

ФАКТ: Targeted B31 component-тест пройшов: **1/1**, exit 0, 369 ms test time. Він виміряв два adapters, destroy старого й `play()` саме нового `<video>` після retry.

ДАЛІ: Запущу весь `LessonPlayer.test.tsx`, щоб одразу перевірити сусідні quality, restore, 410 remint, heartbeat, 403 retry та takeover сценарії.

ФАКТ: Увесь `LessonPlayer.test.tsx` пройшов: **14/14**, exit 0, 661 ms test time. Зеленими лишилися наявні тести quality/keyboard, restore та один media adapter на стабільному вузлі, silent expired remint, bounded media/mint retry, terminal 403 retry, B09/B-7/B10 heartbeat і R28 takeover.

ДАЛІ: Запущу обов’язкові frontend-гейти строго в порядку: `npm run format:check`, `npm run lint`, `npm run typecheck`, `npm test`. `npm run format` і Docker-команди не запускатиму.

ФАКТ: Усі дозволені frontend-гейти пройшли в заданому порядку: format-check exit 0 (164 файли), lint exit 0 (91 architecture files), typecheck exit 0, unit suite **380/380** у 45 файлах, 0 skipped/todo, exit 0.

ДАЛІ: Виконаю завершальний двовісний review, як вимагає `implement`: Standards проти правил репо та Spec проти рішення другого кола. Через пряму заборону git review буде read-only по трьох відомих frontend-файлах (`LessonPlayer.tsx`, його test, незмінений щодо очікувань B31 e2e).

ФАКТ: `code-review` перечитано. Його fixed-point/git процедура несумісна з прямою забороною цієї сесії; spec source — повідомлення користувача «ДРУГЕ КОЛО T15», standards source — `AGENTS.md`. Зберігаю дві незалежні осі й smell baseline, але без git.

ДАЛІ: Запущу паралельні read-only Standards і Spec reviews; агенти не редагуватимуть файлів і не запускатимуть git.

ФАКТ: Обидві осі review запущено паралельно без git і без редагувань. Standards перевіряє лише правила/смели; Spec окремо перевіряє lifecycle ordering, безпомилкові remint-гілки, regression red→green і незмінність B31.

ДАЛІ: Дочекаюся review, перевірю кожну знахідку й за потреби внесу мінімальне виправлення з повтором релевантних гейтів.

ФАКТ: Обидва read-only review завершені. Standards: 0 порушень правил, 0 smell-знахідок; окремо підтверджено, що B31 e2e не послаблено й жодного assertion timeout не збільшено. Spec: блокерів, scope creep і неправильної lifecycle-поведінки не знайдено; підтверджено cleanup-before-setup для old→new DOM node та збереження same-adapter remint для `expired`/`missing`.

РІШЕННЯ: Код після review не змінюю. Єдина partial-знахідка стосувалась точності звіту — нижче явно називаю всі наявні regression-тести, включно з e2e B39/O1.

## Друге коло T15 — підсумок

### Механізм правки B31

Обрано callback ref + state конкретного DOM-вузла:

- стабільний `attachVideo` пише вузол у `videoRef` для імперативних heartbeat/control операцій і в `videoElement` як реактивну lifecycle-залежність;
- media-effect створює adapter саме для `videoElement`; коли terminal error розмонтовує `<video>`, callback ref дає `null`, cleanup знищує старий adapter і очищає `mediaRef`;
- retry монтує новий `<video>`, media-effect створює adapter для нього, а retry-effect, який тепер також залежить від `videoElement`, після цього мінтить і завантажує source; `onReady` викликає `play()` на новому вузлі;
- `failure` не додано в dependencies: `expired`/`missing` не розмонтовують `<video>`, тому їхній remint продовжує використовувати той самий adapter без destroy/recreate.

### Red → green швидкого тесту

- До product fix: targeted B31 exit 1 з точним discriminant `expected [ <video …> ] to have a length of 2 but got 1` на assert `createdVideos.toHaveLength(2)`; retry змонтував новий DOM-вузол, але media лишився старим.
- Після першої половини fix: adapters уже 2 і старий destroy виконано, але `play()` нового вузла ще не відбувся; це виявило відсутню залежність `videoElement` у retry-effect.
- Після повної fix: targeted B31 **1/1 green**; весь `LessonPlayer.test.tsx` **14/14 green**; повний frontend suite **380/380 green**.

### Наявні тести, що захищають сусідню поведінку

| Поведінка | Наявний захист |
|---|---|
| 403/410 remint без terminal error, R38/O1 | `application/playback/recovery.test.ts`: `maps exact media statuses and bounds consecutive retries` фіксує 403 → `await-verdict` та ignore старого prefix; component: `continues heartbeats after an expired session is silently reminted`; e2e: `B39 O1 remints after 410 without takeover and prints browser measurements` вимірює `stopLoad`, silent remint і restore. |
| Вибір якості | component `B24 keeps play, seek and hls.js quality keyboard reachable and axe-clean`; e2e `P01 P09 P10 plays real HLS, selects a rung and seeks into unloaded media`. |
| Heartbeat/cadence/exits | component `B09 sends exactly three tick heartbeats...`, `B-7 follows native playing and pause events...`, `B10 writes pause, seek, hide, nav and ended...`, `R28 stops scheduling and sending heartbeats after a 409 takeover`. |
| Відновлення позиції | component `keeps one media adapter across time updates and restores before duration is known`; component expired-remint test вище; e2e `P04 restores the position after a reload`, `P05 restores the position in another browser context`, `B39 O1...`. |

### Гейти й поточний статус рядків

- Reviewer integration: `ok courses/internal/course/adapters/postgres`; отже B08 integration-доказ тепер **green/виконаний**.
- Локально: format-check exit 0; lint exit 0; typecheck exit 0; unit **380/380**, 45 файлів, 0 skipped/todo; targeted B31 1/1; `LessonPlayer.test.tsx` 14/14.
- Не запускались локально: integration, e2e, build/Compose/smoke — **Docker відсутній**.
- B31 e2e: тут **0/1**, очікую **1/1 green** на повторному reviewer Docker-gate без збільшення 5-секундного assertion timeout; обов’язкова «Пауза» на A лишилась.
- R34 e2e: тут і в першому reviewer-прогоні **0/1**, бо серійний проєкт зупинився після B31; очікую **1/1 green** після виправлення B31.
- Git не використовувався взагалі; commit лишається рецензенту після Docker-гейта.

## Standards

0 findings. Callback-ref/state пара типобезпечна, component-scoped і має різні ролі; тестові назви та doubles чіткі. B31 зберігає `Пауза`, A segment 200, B `taken-over` + segment 403 без нового timeout.

## Spec

0 blocking findings. Lifecycle відповідає рішенню другого кола, red→green доказ дискримінує дефект, сусідні remint/quality/heartbeat/restore сценарії покриті, заборонені області та залежності не чіпались.

ФАКТ: Підсумок review — Standards 0, Spec 0; найгірших знахідок немає.

ПИТАННЯ: Немає. Для остаточного закриття B31 і запуску R34 потрібен reviewer Docker-gate; локальні результати не видаю за e2e-доказ.

ДАЛІ: Третє коло T15. Посилю лише наявний B31 component-тест: задам нетипову ненульову позицію через `currentTime` + `timeUpdate`, перевірю її на новому `<video>`, тотожність другого created element і рівно два mint-виклики; потім тимчасово зламаю restore до 0, зафіксую red, точно відновлю `LessonPlayer.tsx` і прогоню дозволені frontend-гейти. Git не чіпаю — ні staging, ні commit.

ФАКТ: `implement` перечитано. У цьому колі діють TDD, точкові тести, повний frontend suite і review; commit заборонений рішенням рецензента.

ДАЛІ: Перечитаю актуальний B31 component-тест і точний `onReady` restore-фрагмент, не оглядаючи й не змінюючи інших областей.

ФАКТ: Repository briefing перечитано; третє коло не змінює жодного продуктового чи архітектурного контракту. Єдина дозволена постійна зміна — посилення B31 component-тесту.

ДАЛІ: Прочитаю точні фрагменти тесту й `onReady`, після чого додам чотири вимагані спостереження без зміни product code.

ФАКТ: Посилено лише B31 component-тест. До takeover перший `<video>` отримує `currentTime=47` (не збігається з fixture position 12, початком 0 чи duration 100), після `timeUpdate` slider прямо показує `47`, отже компонент справді відстежив нову позицію. Після retry тест вимагає `remountedVideo.currentTime===47`, `createdVideos[1]===remountedVideo` і рівно два виклики `actions.play`; наявні destroy та `playedVideos==[firstVideo, remountedVideo]` збережено.

ДАЛІ: Запущу targeted B31-тест на прийнятому product code; очікую green до контрольної мутації.

ФАКТ: Посилений targeted B31 пройшов: **1/1**, exit 0, 428 ms test time. Ненульова позиція 47 відстежується й відновлюється на remounted element; два mint-виклики та identity assertions проходять.

ДАЛІ: Зафіксую SHA-256 `LessonPlayer.tsx` до тимчасової мутації, зміню лише `onReady` на `resumeAt=0`, запущу один B31-тест, потім негайно поверну точний блок і порівняю SHA-256.

ФАКТ: Контрольний SHA-256 product file до мутації: `835dbe5f2ccd8d8a16d5b72f4c9fe6ab4d3ea52ab8cf302c4203c43f16d0d9af`.

ДАЛІ: Тимчасово заміню обчислення `resumeAt` у `onReady` на константу 0 і запущу тільки B31 component-тест.

ФАКТ: Контрольну мутацію застосовано: `onReady` тимчасово ігнорує `pendingTimeRef` та встановлює `resumeAt=0`. Це не фінальна продуктова зміна.

ДАЛІ: Запущу targeted B31 і зафіксую точний failure позиційного assert.

ФАКТ: Контрольна мутація дала потрібний red: targeted test exit 1, точний текст `expected +0 to be 47 // Object.is equality` на `expect(remountedVideo.currentTime).toBe(resumePosition)`; diff Expected 47 / Received 0. Усі попередні непрямі докази дійшли до цього останнього assert, отже падіння специфічне до restore position.

ДАЛІ: Негайно відновлю точне production-обчислення `resumeAt` через `pendingTimeRef`, після чого звірю SHA-256 з контрольним значенням.

ФАКТ: Тимчасову мутацію прибрано; `onReady` знову обчислює bounded `resumeAt` із `pendingTimeRef.current` точно тим самим блоком, що до мутації.

ДАЛІ: Обчислю SHA-256 `LessonPlayer.tsx`; він мусить точно збігтися з `835dbe5f…d0d9af`.

ФАКТ: SHA-256 після відновлення: `835dbe5f2ccd8d8a16d5b72f4c9fe6ab4d3ea52ab8cf302c4203c43f16d0d9af` — байт-у-байт збігається з контрольним. Нульова дельта тимчасової продуктової правки підтверджена без git.

ДАЛІ: Повторю targeted B31 на відновленому продукті, потім запущу рівно format-check, lint, typecheck і повний unit suite.

ФАКТ: Після точного відновлення product file targeted B31 знову green: **1/1**, exit 0, 351 ms test time.

ДАЛІ: Запущу обов’язкові frontend-гейти строго в порядку: `npm run format:check`, `npm run lint`, `npm run typecheck`, `npm test`.

ФАКТ: Усі дозволені frontend-гейти пройшли в порядку: format-check exit 0 (164 файли), lint exit 0 (91 architecture files), typecheck exit 0, unit suite **380/380** у 45 файлах, 0 skipped/todo, exit 0.

ДАЛІ: Виконаю завершальний двовісний review за `implement`: перевірю лише постійну дельту B31-тесту проти правил репо й вимог третього кола. Git і редагування агентами заборонені.

ФАКТ: `code-review` перечитано. Його git fixed-point процедура знову непридатна через пряму заборону; дві осі працюватимуть по відомій єдиній постійній зміні в `LessonPlayer.test.tsx` і spec source з повідомлення рецензента.

ДАЛІ: Запущу паралельні read-only Standards і Spec reviews без git та без редагувань.

ФАКТ: Обидві review-осі запущено. Standards перевіряє лише test-code quality; Spec — невакуозність 47, реальне `timeUpdate`, identity/mint assertions, mutation red і нульову product delta.

ДАЛІ: Дочекаюся review, перевірю знахідки й допишу фінальний підсумок третього кола.

---

ФАКТ (рецензент, гейт 2026-09-14): `make check` → **GATE_EXIT=2** на цілі `e2e`. До неї зелене все, зокрема
`integration`: `ok courses/internal/course/adapters/postgres` — пакет B08 на PostgreSQL. e2e: **1 failed, 45 passed,
15 did not run**. Упав B31 (`playback.spec.ts:1146`, «Пауза» не видима за 5 с). **R34 не запускався** — серійний
проєкт зупинився після падіння. Не OOM: `Terminated` у лозі немає. Teardown `playback-profile` відпрацював,
api повернуто на 30/180.

ФАКТ: знімок сторінки A в мить падіння (`test-results/…retake-the-slot-playback/error-context.md`) — «Стан плеєра:
Готовий», кнопка «Відтворити», «Пауза» відсутня, заголовка витіснення немає. Повторний мінт після «Почати перегляд
знову» пройшов із **201** (`:1144`).

ФАКТ: **продуктовий дефект, перевірено за кодом.** `LessonPlayer.tsx:221-252` — ефект створення медіа захоплює
`const video = videoRef.current`; залежності `[createMedia, handleMediaError, lesson.videoState]`. У стані помилки
`:371` робить ранній `return <ErrorState/>`, і `<video>` розмонтовується, але жодна залежність не змінюється, тож
ефект не очищається, і `mediaRef.current` тримає медіа старого елемента. Повтор монтує новий `<video>`, ефект не
перезапускається, перевірка `:362` проходить, мінт забирає слот, а `onReady` викликає `video.play()` на відчепленому
елементі. «Відтворити» теж не допомагає: сесія є, тож `:461` робить `play()` на новому елементі без медіа. Мертвий
плеєр до перезавантаження, після повтору з будь-якого стану помилки.

РІШЕННЯ (рецензент): це дефект **брифа**, не реалізації. `work/brief-t15.md` §2.2 заявив «продукт уже вміє» для
B31, перевіривши лише, що `onRetry` викликає `loadSession`, і не перевіривши, що відтворення відбувається на видимому
елементі. Тест B31 написано правильно — він і знайшов дефект. Гіпотезу рецензента «режим `'initial'` не відтворює»
спростовано кодом: `onReady` відтворює завжди, але не той елемент.

ДАЛІ: рішення користувача — виправити продукт у T15, окремим слайсом чи боргом. Гейт у будь-якому разі
переганяється, бо R34 не запускався.

ФАКТ: Корекція хронології журналу. Записи третього кола вище були додані після відомого мені попереднього останнього рядка, але рецензентський блок уже існував нижче; отже вони опинилися не в справжньому EOF. Це виявив завершальний review. Старі записи не переношу й не переписую; цей коригувальний підсумок додано після фактичного останнього рядка рецензента.

РІШЕННЯ: Єдина постійна зміна третього кола — у `frontend/src/presentation/learn/LessonPlayer.test.tsx`. `LessonPlayer.tsx` після вимаганої контрольної мутації відновлений байт-у-байт; e2e, backend, T14/T16, docs, tasks і generated не змінювались.

## Третє коло T15 — фінальний підсумок у справжньому EOF

- У B31 перед takeover встановлюється `currentTime=47`, потім `timeUpdate`; slider зі значенням `47` доводить, що позицію прийняв стан компонента. 47 відрізняється від DOM default 0, fixture position 12 і duration 100.
- Після retry явно перевіряються `createdVideos[1]===remountedVideo`, `playedVideos==[firstVideo, remountedVideo]`, `actions.play` рівно 2 рази та `remountedVideo.currentTime===47`.
- Контрольна мутація `onReady → resumeAt=0` дала red: `expected +0 to be 47 // Object.is equality`, Expected 47 / Received 0. Отже позиційний assert не порожній і падає саме при втраті restore.
- До/після мутації SHA-256 `LessonPlayer.tsx` однаковий: `835dbe5f2ccd8d8a16d5b72f4c9fe6ab4d3ea52ab8cf302c4203c43f16d0d9af`. Нульова постійна дельта product file підтверджена без git.
- Після відновлення targeted B31: **1/1 green**. Фінальні гейти: format-check exit 0 (164 файли), lint exit 0 (91 architecture files), typecheck exit 0, unit suite **380/380** у 45 файлах, 0 skipped/todo.
- `npm run format` не запускався. Docker/e2e не запускались у цій сесії. Git не використовувався взагалі.

## Standards

0 findings: hard violations 0, smells 0. Тест strict-TypeScript-safe, має явні DOM guards і локальні scenario-specific spies.

## Spec

0 findings. `timeUpdate` + slider роблять ненульову передумову спостережуваною; identity, два mint-и й restore 47 перевірені окремо; mutation red дискримінує втрату позиції; продукт відновлений точно.

ФАКТ: Підсумок review — Standards 0, Spec 0; найгірших знахідок немає.

ПИТАННЯ: Немає. Третє коло готове до reviewer Docker-gate; B31/R34 тут не оголошуються e2e-доведеними.
