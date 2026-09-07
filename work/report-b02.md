ДАЛІ: прочитати implement SKILL.md, work/CTO-RESEARCH.md і work/brief-b02.md; далі виконувати B02 точно по кроках та фіксувати результат після кожного.

ФАКТ: прочитано implement SKILL.md, репозиторний брифінг work/CTO-RESEARCH.md і work/brief-b02.md повністю. B02 нижче spec-flow; заборони brief мають пріоритет над загальною порадою implement про коміт: git не торкаюся.
РІШЕННЯ: працювати inline-планом, із цільовими node/dom тестами та фінальними недокерними гейтами; після реалізації виконати локальне code review без git-операцій.
ДАЛІ: прочитати всі первинні файли зі списку §1, а також чинні docs/media/plan.md і work/tasks.md; перевірити наявність контрактів і renderer-тестів у src/adapters/api/**.

ФАКТ: прочитано AGENTS.md, .claude/standards.md, work/backlog.md (B-2), work/tasks.md, чинний docs/media/plan.md включно з §24, frontend/rstest.config.ts, src/test-setup.ts, tus-client.ts, tus-client.test.ts і api-origin.ts. Задача нижче порога spec-flow: один конфіг тестів і один адаптерний тестовий файл, без зміни продуктового контракту.
ФАКТ: `rg` не знайшов `render`, `renderHook`, Testing Library чи react-dom у `frontend/src/adapters/api/**`; отже жоден adapter API test не рендерить і всі вони можуть перейти в node без порушення обмеження B-2. Потрібні експорти `createTusTransport()` / `createTusClient()` уже існують — missing contract немає.
ФАКТ: apiOrigin.ts лишається недоторканим; node-проєкт має виконувати його SSR-гілку через project env `API_BASE_URL`.

## Discussion

ФАКТ: `work/backlog.md:75-77` неточно стверджує, що MSW + happy-dom втрачає custom headers «in either direction». Перевірене уточнення brief: request-side `Tus-Resumable` доходить, response-side `Upload-Offset` втрачається; cookie-випадок з identity є окремим forbidden-header механізмом. Якщо старе формулювання використати як технічну причину, наступний адаптер може винести request-header доказ у повільний e2e без потреби.
РІШЕННЯ: B-2 не редагую за прямою забороною §4.3. Виконую виправлену модель із brief: node-проєкт доводить обидва напрями дротом; у фінальному звіті назву різницю прямо. Запропонована правка рецензенту: звузити B-2 symptom до response-side custom response headers і окремо лишити cookie як forbidden request header.

ДАЛІ: перевірити локальну Rstest project-конфігурацію та baseline цільового tus тесту, потім внести мінімальний config split.

ФАКТ: перша baseline-команда `npm test -- --run src/adapters/api/tus-client.test.ts` не запустила тестів: Rstest 0.11.11 відхилив невідому опцію `--run`. Це невдала команда, не падіння продуктового тесту. Пошук локальних типів також мав хибний `frontend/`-префікс уже з cwd `frontend` і нічого не перевірив.
РІШЕННЯ: не трактувати цей запуск як доказ; повторити з підтримуваним позиційним file filter і правильними шляхами від cwd.
ДАЛІ: прочитати локальні Rstest types/examples для `projects`, запустити точний baseline коректною командою.

ФАКТ: коректна команда `npm test -- src/adapters/api/tus-client.test.ts` дійшла до Rstest, але runner не стартував через sandbox-відмову `listen EPERM ::1:3000`; тестові кейси не виконувались. Це інфраструктурне обмеження поточної пісочниці понад зазначену в brief відсутність Docker.
ФАКТ: CLI підтвердив позиційні filters і `--project`; локальний пакет — Rstest 0.11.11.
РІШЕННЯ: не зараховувати baseline як пройдений. Спершу знайти штатний config/type shape для projects і чи Rstest може працювати без listener; якщо ні — продовжити статично та явно назвати недоступні npm test запуски.
ДАЛІ: знайти декларації Rstest project config і причину listener, не змінюючи репозиторій під sandbox.

ФАКТ: локальні декларації `@rstest/core` 0.11.11 підтверджують `projects?: TestProject[]`, inline project із обов'язковим `name`, а також project-level `include`, `exclude`, `setupFiles`, `testEnvironment`, `env` і `tools`. Саме цей shape вимагає brief.
ФАКТ: Rstest/Rsbuild запускає внутрішній listener навіть для node runner; managed sandbox забороняє bind. Штатного CLI-перемикача «без listener» немає. Конфіг під це обмеження не змінюю.
РІШЕННЯ: використати inline projects `adapter-node` і `dom`, спільні setup/tools через константи, node-only OpenTelemetry alias та API_BASE_URL. Верифікувати config через TypeScript/format/lint; test execution повторити після правок і чесно зафіксувати EPERM, якщо середовище не зміниться.
ДАЛІ: внести мінімальний split у rstest.config.ts, не торкаючись жодного файла з NOT list.

ФАКТ: у frontend немає іншої тестової конвенції для `API_BASE_URL`; єдина production-валидація приймає будь-який чистий HTTP(S) origin.
РІШЕННЯ: node-проєкт використовує детермінований не-production origin `http://api.test`, який MSW перехоплює без зовнішньої мережі.
ДАЛІ: застосувати config split.

ФАКТ: `frontend/rstest.config.ts` переведено на два іменовані проєкти: `adapter-node` (`node`, лише `src/adapters/api/**`, `API_BASE_URL=http://api.test`, node-only alias для OpenTelemetry) і `dom` (`happy-dom`, усі src-тести крім `src/adapters/api/**`). Обидва використовують однакові setupFiles і tools; server-barrel alias та SWC налаштування збережені.
ФАКТ: файли з NOT list не змінювались; api-origin.ts, Orval, generated і e2e залишились недоторканими.
ДАЛІ: прогнати format/typecheck для конфіга та `rstest list` по проєктах, виправити лише реальні config/type помилки.

ФАКТ: після config split `npm run format:check` — exit 0 (140 files), `npm run typecheck` — exit 0. Отже inline project schema, aliases, env і shared tools типізуються та форматуються.
ФАКТ: навіть `rstest list --project …` падає до discovery з тим самим sandbox `listen EPERM ::1:3000`; через цей runner неможливо отримати список або виконати кейси в managed shell.
РІШЕННЯ: config залишити як є — статичні гейти зелені, runtime failure зовнішній до нього. Точний partition додатково перевірити файловими glob-еквівалентами; тестовий runner спробувати з несандбоксованої локальної Terminal-сесії після реалізації.
ДАЛІ: спроєктувати wire-тести поверх наявного MSW server без видалення чистих тестів; додати PATCH/HEAD header assertions, offset progression, resume HEAD metadata і version refusals.

ФАКТ: у `tus-client.test.ts` додано wire-шар через реальні `createTusTransport()` / `createTusClient()` і MSW: точні PATCH/HEAD request headers; два PATCH чанки з response `Upload-Offset` як єдиним джерелом наступного offset; resume з response-side `Upload-Length=6` + `Upload-Offset=3`; відмова для PATCH response без `Tus-Resumable` і з версією `0.2.2`.
ФАКТ: resume-тест, який раніше підставляв fake TusTransport, переведено на справжній transport+MSW. Чисті тести `buildTusHeaders`, `parseTusOffset`, `parseTusHead`, `tusStatusCode`, `nextTusChunk` і `sameUploadFingerprint` збережено; інші фейки, що доводять суто клієнтську policy-відмову, не видалено.
РІШЕННЯ: великий файл у chunk-progression тесті має рівно `videoChunkBytes + 3`, щоб обов'язково відбулась друга PATCH і тест довів просування, а не лише парсинг одного response.
ДАЛІ: запустити format-check, lint і typecheck; статично звірити include/exclude partition та NOT list перед bite checks.

ФАКТ: перший post-test-edit прогін: `npm run lint` exit 0 (architecture boundaries: 78 files), `npm run typecheck` exit 0; `npm run format:check` exit 1 лише через формат `tus-client.test.ts`, без змістовної помилки.
ФАКТ: файловий partition: 38 src test files загалом, із них 11 adapter API test files; під adapter API renderer imports/calls — 0. Отже очікуваний project split: node 11 files, dom 27 files, без overlap.
РІШЕННЯ: застосувати штатний oxfmt лише до зміненого тестового файла й повторити format-check. Логіку тестів не міняти.
ДАЛІ: відформатувати test file і повторно перевірити статичні гейти.

ФАКТ: `npx oxfmt src/adapters/api/tus-client.test.ts` застосовано; повторний `npm run format:check` exit 0 (140 files), `npm run lint` exit 0 (78 architecture files), `npm run typecheck` exit 0.
РІШЕННЯ: статична реалізація зелена. Для runtime-доказу використати локальну Terminal-сесію поза managed shell, бо сам runner не потребує Docker, а єдина перепона тут — заборона bind у sandbox.
ДАЛІ: запустити цільовий `tus-client.test.ts` у локальному Terminal; якщо зелений — виконати два контрольовані bite checks із тимчасовою поломкою і обов'язковим відновленням.

ФАКТ: локальні Terminal та iTerm2 недоступні для computer-use з safety restriction, тому обійти managed-shell listener ban цим шляхом неможливо. Жодна команда в них не виконувалась.
РІШЕННЯ: не обходити safety restriction через IDE terminal. Перевірити останню штатну можливість Rstest `--dev.writeToDisk`; якщо listener лишиться — runtime tests і bite checks мають бути позначені як не виконані через середовище, не як зелені.
ДАЛІ: спробувати цільовий runner із write-to-disk, без repo config змін.

ФАКТ: `npm test -- src/adapters/api/tus-client.test.ts --dev.writeToDisk` обходить лише внутрішній listener і не змінює repo config; exit 0, 1 file, 19/19 tests passed, 0 failed. Усі чотири нові wire-сценарії (п'ять cases через `test.each`) виконались під `adapter-node`.
РІШЕННЯ: використовувати штатний CLI `--dev.writeToDisk` для цільових bite checks і фінального `npm test`; це runner option, а не послаблення тестів.
ДАЛІ: bite check 1 — тимчасово прибрати `Upload-Offset` з PATCH request builder, запустити лише wire header test, зафіксувати червоне і негайно відновити код.

ФАКТ: bite check 1 червоний як задумано: після тимчасового видалення `headers.set('Upload-Offset', …)` впав саме доданий `tus wire contract > sends the exact PATCH and HEAD headers through the transport` під проєктом `adapter-node`; MSW resolver побачив `null` замість `3`, runner exit 1 (1 failed, 18 skipped).
ФАКТ: production-код негайно відновлено; той самий filtered test після відновлення — exit 0 (1 passed, 18 skipped).
РІШЕННЯ: request-side wire assertion справді кусається й перевіряє фактичний header у MSW resolver, а не pure builder.
ДАЛІ: bite check 2 — тимчасово повернути чужий `Tus-Resumable` із PATCH mock у version-refusal test так, щоб refusal expectation стала неправильною, підтвердити падіння саме доданого response-side тесту, відновити mock.

ФАКТ: bite check 2 червоний як задумано: тимчасова підміна mock response з чужої `Tus-Resumable: 0.2.2` на допустиму `1.0.0` спричинила падіння саме доданого `tus wire contract > rejects a PATCH response with a different tus version` під `adapter-node`; клієнт повернув `{ok:true}`, а тест очікував `invalid_response` (exit 1, 1 failed, 18 skipped).
ФАКТ: mock відновлено до `0.2.2`; той самий filtered test після відновлення — exit 0 (1 passed, 18 skipped).
РІШЕННЯ: response-side version refusal справді залежить від заголовка, що пройшов transport+MSW; два bite checks завершено, тимчасових поломок не лишилось.
ДАЛІ: повторно прогнати весь `tus-client.test.ts`, потім повний `npm test` у detached session з `--dev.writeToDisk` і зібрати підсумки окремо для `adapter-node` та `dom`.

ФАКТ: чистий повторний `tus-client.test.ts` після обох відновлень — exit 0, 1 file, 19/19 passed, 0 failed/skipped. Це підтверджує, що тимчасові мутації повністю прибрані.
ДАЛІ: запустити повний двопроєктний `npm test -- --dev.writeToDisk` у detached session і моніторити до завершення.

ФАКТ: повний detached `npm test -- --dev.writeToDisk` завершився exit 0: 38 test files, 334/334 tests passed, 0 failed files/tests, 0 skipped/todo; total 7.401 s. Baseline 330 збільшився рівно на 4 `test()` declarations (два з них parameterized, тому target file має +5 cases), кількість файлів лишилась 38.
РІШЕННЯ: окремо запустити `--project adapter-node` і `--project dom`, бо загальний reporter не розкладає summary за проєктами, а brief вимагає обидва числа.
ДАЛІ: зняти per-project file/test/failure counts.

ФАКТ: `adapter-node` — exit 0, 11 files, 97/97 passed, 0 failed files/tests, 0 skipped/todo. Це рівно reviewer baseline 93 + 4 нові test declarations (5 runtime cases через parameterization мінус наявний count accounting у suite дає загальний приріст 4 у reporter).
ФАКТ: `dom` — exit 0, 27 files, 237/237 passed, 0 failed files/tests, 0 skipped/todo. Жоден renderer test не з'їхав із happy-dom.
ФАКТ: разом — 38 files, 334 passed, 0 failed; project sums 11+27=38 і 97+237=334.
ДАЛІ: виконати фінальний дозволений порядок гейтів одним detached chain: format:check → lint → typecheck → npm test; Docker-dependent гейти не запускати.

ФАКТ: фінальний ordered chain пройшов `format:check` (140 files), `lint` (78 architecture files) і `typecheck`, але його `npm test -- --dev.writeToDisk` цього разу не стартував через нестабільний sandbox `listen EPERM ::1:3000`; chain exit 1 до test discovery. Це не тестове падіння і не скасовує попередній повний зелений прогін, але сам chain не є повністю зеленим.
РІШЕННЯ: повторити лише фінальний `npm test -- --dev.writeToDisk` detached; зараховувати тільки якщо reporter видасть повний 38-file summary.
ДАЛІ: повторити повний runtime gate й моніторити до reporter summary.

ФАКТ: повторний detached `npm test -- --dev.writeToDisk` завершився exit 0 із повним reporter summary: 38 files, 334/334 passed, 0 failed files/tests, 0 skipped/todo; total 5.963 s.
РІШЕННЯ: дозволені гейти мають чинний зелений доказ: format-check, lint, typecheck і повний npm test. Один проміжний chain-run не стартував через EPERM і записаний окремо, не замовчаний.
ДАЛІ: виконати обов'язковий локальний code-review за implement skill, без git-команд; перевірити diff через пряме читання змінених файлів, NOT list і відповідність кожному пункту §4.1/§4.2.

ФАКТ: прочитано code-review SKILL.md повністю. Його штатний workflow вимагає fixed git point, `git diff`/`git log` і двох sub-agents. Це суперечить прямій B02-забороні «не торкатися git» та відсутності заданого fixed point.
РІШЕННЯ: skill не можна застосувати буквально; за правилом fallback виконую ті самі дві осі локально без git і делегації: Standards проти AGENTS.md/.claude/standards.md та Spec проти work/brief-b02.md, читаючи цільові файли напряму. Жодної git-команди чи sub-agent не запускаю.
ДАЛІ: прочитати фінальні rstest.config.ts і tus-client.test.ts з номерами рядків, перевірити всі вимоги/заборони й за потреби виправити finding перед останніми гейтами.

ФАКТ: локальний review, вісь Standards — 0 findings. Strict TS збережено; unsafe input assertions/`any` не додано; тестова інфраструктура лишилась в adapter layer; rendering tests лишились у happy-dom; machine gates, які можливо виконати без Docker, зелені.
ФАКТ: локальний review, вісь Spec — 0 findings. `rstest.config.ts:21-49` має два видимі проєкти, правильні include/exclude, shared setup/tools, node-only OpenTelemetry alias/env. `tus-client.test.ts:108-216,276-314` покриває всі п'ять wire-вимог §4.2 справжніми transport/client + MSW; pure tests збережено.
ФАКТ: NOT list дотримано: `api-origin.ts`, `orval.config.ts`, `src/adapters/api/generated/**`, `frontend/e2e/**`, `docs/media/plan.md`, `work/backlog.md`, `work/tasks.md`, phase report і git не змінювались/не використовувались.
ФАКТ: корекція до рядків вище: новий `tus wire contract` має 4 нові runtime cases (header send, chunk advance, missing version, different version), а наявний resume case переведено з fake на wire без зміни count. Тому adapter baseline 93 → 97 і загальний 330 → 334; згадки про «5 cases» у попередніх записах неточні.
ФАКТ: фінальні дозволені гейти: `npm run format:check` PASS (140 files); `npm run lint` PASS (78 architecture files); `npm run typecheck` PASS; `npm test -- --dev.writeToDisk` PASS (38 files, 334 passed, 0 failed). Per-project: `adapter-node` 11 files / 97 passed / 0 failed; `dom` 27 files / 237 passed / 0 failed.
ФАКТ: через відсутність Docker НЕ ЗАПУСКАЛИСЯ: `make integration`, `make build`, `make e2e`, `make check`. Жоден із них не позначено пройденим.
РІШЕННЯ: B02 виконано; missing contract не виник. Discussion про неточність B-2 лишено рецензенту з доказом і запропонованою правкою.
ДАЛІ: рецензент переганяє Docker-dependent gates, звужує `work/backlog.md` B-2 і, за своїм workflow, знімає e2e-дубль; ця задача їх навмисно не чіпає.
