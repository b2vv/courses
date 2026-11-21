NEXT: Прочитати чинні shared/project правила та контракт U1 (D1, D9/m3, Ux01–Ux03, §4 U1, §6), перевірити гілку й робоче дерево без змін git.
FACT: Shared agent-entry прочитано; поточний запит є окремим явно дозволеним U1, делегування не планую.
NEXT: Зіставити прочитані локальні правила й план із поточними компонентами та тестами; визначити точні місця змін.
FACT: Корінь <repo>, гілка admin-editor-ux підтверджені. Є попередні сторонні tracked/untracked зміни; їх зберігаю. AGENTS.local.md прочитано; Docker/e2e та git mutations не дозволені цим завданням.
NEXT: Повторно прочитати обмеженими фрагментами контракт і правила, оскільки спільне читання великих файлів було обрізане.
FACT: План r3 прочитано повністю: D1, D9/m3, Ux01–03, §4 U1, §6 та checkpoint ACCEPTED; U1 READY. Власник обмежив цей writer frontend-командами; повний e2e і CSS/browser proof виконує owner-сесія.
DECISION: 2026-11-18, Codex writer: виконати лише U1 без нових abstractions. Review route — наявна owner-сесія Courses 1 приймає результат і організує незалежний read-only review за планом; тут author checks, без spawn або claims review PASS.
NEXT: Прочитати CourseEditorView, root NotFound, локалі, наявні unit/e2e; внести scoped зміни та виконати п'ять дозволених команд у порядку.
FACT: Standards: U1 нижче full/spec-flow порога (frontend-only); застосовую прийнятий план і targeted author gates. Кореневий shared шлях skills/ponytail/SKILL.md відсутній — шукаю канонічний grouped шлях, без install чи fallback на stale копію.
NEXT: Встановити наявні test fixtures/route mocks та upload locators перед редагуванням.
FACT: Знайдено канонічний skill skills/pt/ponytail/SKILL.md. Три CourseEditorView TabPanel мають ui-tab-panel + shouldForceMount; focusField уже перемикає tab через flushSync. Root NotFound завжди повертає /courses. Існуючий S2b summary тест не виділяє окремо обидва hidden-tab сценарії.
NEXT: Додати локальну panel class/CSS, path-scoped admin 404 та явні Ux01/Ux02/Ux03 тести; upload assertions прив'язати до lesson locator.
FACT: UploadReady playback вже scope-ить module/lesson і poll-ить minted videoId; B40 лишається незмінним. discussion uploader потребує Lessons click і lesson scope; transcript/enrollment мають Lessons click, додаю scoped visible-control guard. 404 path boundary обмежую /admin або /admin/, не /administrator.
NEXT: Внести мінімальний product diff та unit/e2e cases; тести root виконати через реальний memory-router з production notFoundComponent.
FACT: Product diff внесено: course-editor-panel на трьох force-mounted панелях + окрема inert CSS rule; NotFound читає pathname й обирає /admin з новим uk/en page.goToAdmin або старий /courses. Глобальний ui-tab-panel не змінено.
NEXT: Додати unit cases для всіх трьох selections, обох напрямків summary focus та admin/public 404.
FACT: Додано Ux01 (3 selections; усі panel mounted, class+inert, selected ancestor не hidden), Ux02 (Overview→Lessons і Lessons→Overview, summary focus/tab/input, save=0), Ux03 (реальний memory-router з production NotFound; /admin, /admin/missing, public і /administrator; обидві локалі).
NEXT: Змінити чотири upload journeys та додати admin e2e hidden lesson validation із cleanup.
NEXT: Перевірити accessible role активного video upload control; використати точний OR Upload/Replace name перед setInputFiles, без hidden/force.
FACT: У чотирьох upload paths додано scoped Upload|Replace button visibility assertion до setInputFiles; discussion тепер реально натискає Lessons, file locator scope-нуто до target lesson. Polling/існуючі assertions збережені.
NEXT: Додати admin.spec Ux02 зі створенням валідних module/lesson UI fields, invalid lesson на Overview, summary link, tab+focus та draft cleanup.
FACT: Новий admin.spec Ux02 проходить Lessons UI → додати module+lesson → Overview → invalid submit → summary link → selected Lessons + visible/focused lesson input; reload доводить, що invalid form не збережено. Draft cleanup у finally.
NEXT: Запустити рівно дозволений frontend verification sequence; реальні exits і зауваги записувати окремо після кожної команди.
FACT: npm run format запущено у frontend; очікую завершення та фіксую фактичний exit перед наступною командою.
FACT: npm run format — PASS, exit 0 (oxfmt: 416 files, 173ms); попередній запис про очікування уточнено фактичним завершенням.
NEXT: npm run format:check.
FACT: npm run format:check — PASS, exit 0; усі файли відформатовані.
NEXT: npm run typecheck.
FACT: npm run typecheck запущено; результат фіксується після завершення нижче.
NEXT: Отримати завершення tsc; lint запускати тільки після фактичного результату typecheck.
FACT: Перший npm run typecheck — FAIL, exit 2: у нових Testing Library role options помилково використано Playwright exact; Route.options.notFoundComponent типізовано optional.
FACT: Виправлено лише нові unit tests: role name string уже exact у Testing Library; explicit undefined guard звужує production component без type assertion. Ux01 також перевіряє рівно одну не-inert панель.
NEXT: Повторити ordered format → format:check → typecheck на виправленому snapshot, потім lint і rstest.
FACT: Повторний ordered format/format:check/typecheck запущено; остаточні exits нижче після завершення.
FACT: Повторні npm run format і npm run format:check — PASS, обидва exit 0 (416 files); typecheck result перевірено за session output.
FACT: Повторний npm run typecheck — PASS, exit 0.
NEXT: npm run lint.
FACT: npm run lint запущено; зберігаю фактичний output/exit після завершення.
FACT: npm run lint — PASS, exit 0; oxlint без warnings, architecture boundaries checked (244 files).
NEXT: npx rstest run src --testTimeout 30000 (весь src, без e2e/build).
FACT: Повний src rstest запущено, raw log /tmp/courses-u1-rstest.log; результат ще не заявлено.
NEXT: Дочекатися rstest та перевірити raw failures/summary, тоді прочитати фінальний scoped diff.
FACT: Фінальний scoped diff переглянуто: product зміни лише D1+m3, нові tests і upload visibility guards. Існуючі e2e assertions/polling не видалені; заборонені force/includeHidden/API-заміни UI не додані.
NEXT: Перевірити завершення повного rstest; будь-які падіння виправити в межах U1 та повторити релевантні gates.
FACT: npm run format не зачепив сторонніх frontend файлів; status обмежений 11 tracked U1 файлами і новим __root.test.tsx. Scoped git diff --check виконано; rstest усе ще очікується, raw log перевірено.
NEXT: Отримати остаточний rstest verdict, зберегти точні counts/час/exit і owner next action.
FACT: Перший npx rstest run src --testTimeout 30000 завершився exit 1; raw /tmp/courses-u1-rstest.log (11165 bytes). Scoped diff --check — PASS, exit 0. Читаю початок log, щоб розрізнити failures та runner failure, не прирівнюю stderr warnings до причини.
NEXT: Визначити причину exit 1 й виправити в дозволених межах.
FACT: Повний rstest: 1439 passed / 3 failed / 1442 tests, 115 files; failures лише нові Ux01/Ux02. Ux01 очікував role=tabpanel на inactive panel, але React Aria прибирає role у force-mounted inert панелей; перевіряю implementation та решту failures перед correction.
NEXT: У tests вибирати mounted panels за dedicated class (контракт), не доступним role inactive content; nearest panel також за class.
FACT: Ux02 обидва failures також caused by inactive panel role removal (closest role повертав null). Селектори mounted/nearest panels замінено на .course-editor-panel; жодну assertion не видалено. Ux03 усі 4 cases вже PASS у першому full run.
NEXT: Повторити повну дозволену ordered sequence на виправлених unit selectors.
FACT: Фінальний ordered run запущено; rstest raw output піде в /tmp/courses-u1-rstest-final.log. Результат кожної команди звірю за output та chain exit.
FACT: Фінальні npm run format та npm run format:check — PASS, exit 0 (416 files; 123ms / 116ms). Inactive role removal підтверджено фактичним DOM test failure; package source TSX не поставляється за перевіреним шляхом, тому source-code proof не заявляю.
NEXT: Дочитати завершення фінального run та raw rstest summary; зафіксувати snapshot для owner review.
FACT: Final writer snapshot: root <repo>; branch admin-editor-ux; base/HEAD 10436cdf2988a17887c959dd48b363e6b117e9d7, unstaged product diff + new __root.test.tsx. No commit/staging/branch changes.
FACT: SHA-256 docs/admin-editor-ux/plan.md: 2e297b62257f96d7005678d494a9e8509390566abd130b5a014244872ae9d026
FACT: SHA-256 frontend/src/modules/admin/ui/CourseEditorView.tsx: bfc4ec95e535bdb3ddedb406dc30d3c5413b580d97b1f9deb783c7e1496dec0c
FACT: SHA-256 frontend/src/modules/admin/ui/CourseEditorView.test.tsx: d60116b8752ac9c988de7fee4d68dfbf1a4067037fa60833fd70c465fe11e1de
FACT: SHA-256 frontend/src/app/styles/main.css: d83e7a53149abd40115cee6ed588088cbdcded5ed964a3cc71b47cc966c18d0e
FACT: SHA-256 frontend/src/routes/__root.tsx: d69d3414e25efb6d37162469adcb0920c151c30a80c5952a145a61128d75d8ac
FACT: SHA-256 frontend/src/routes/__root.test.tsx: 76286883c65b23ac7734366eb371f9c2b40d6f3c641a0e1ceadd983e05ea68aa
FACT: SHA-256 frontend/src/app/providers/locales/uk.ts: ca08a6e515850d80aa8a8ea0a70359ca472d801a711f52238ae37b3722f816ab
FACT: SHA-256 frontend/src/app/providers/locales/en.ts: 5ca56c12dadcf76c04ba7c1514fb8e1a9ed8169d527a1f3d62a17cc781b6edb6
FACT: SHA-256 frontend/e2e/admin.spec.ts: e63a3091504ec4c8715097547266ee41314d784c3c50507d8cd50c9edb118ab2
FACT: SHA-256 frontend/e2e/discussion.spec.ts: 6aae3bf27988c0ef1002abf051557cab26173a6e1ab6b4bc5e110c17951419fc
FACT: SHA-256 frontend/e2e/transcript.spec.ts: 872e7e39dfaeafd7a93e313419f9fe67eadde2189b8e6509b284c158bd8a1a9f
FACT: SHA-256 frontend/e2e/enrollment.spec.ts: e92348bafe8971b34ef321bbbfc2f5d618dc036279bad025be1935ce7166d9a8
FACT: SHA-256 frontend/e2e/playback.spec.ts: 4cd61f7f99031b9d1609294f9a0494f192075a0ffa80d08da100a75f20ac3594
FACT: Scoped git diff --check — PASS, exit 0. Фінальні typecheck і lint — PASS, обидва exit 0 (chain дійшов rstest; architecture 244 files).
NEXT: Отримати rstest exit та summary; передати owner browser/e2e + незалежний review без оголошення U1 gate closed.
NEXT: Звірити остаточний результат rstest з raw log перед закриттям writer handoff.
FACT: Фінальний npx rstest run src --testTimeout 30000 — PASS, exit 0: 115 files, 1442/1442 passed, 0 failed/skipped/todo; total 16274ms. Raw: /tmp/courses-u1-rstest-final.log (2026-11-18T11:47:13.174Z). Перший failing run лишається /tmp/courses-u1-rstest.log; причини та correction записані вище.
FACT: Final verification: npm run format = 0; npm run format:check = 0; npm run typecheck = 0; npm run lint = 0; npx rstest run src --testTimeout 30000 = 0, виконані в цьому порядку на фінальному snapshot. git diff --check frontend = 0.
FACT: U1 writer завершено: D1 scoped CSS + три classes з shouldForceMount; Ux01/Ux02 unit proof (обидва напрямки); D9/m3 path-scoped 404 з uk/en та Ux03 tests; чотири upload migrations і новий admin Ux02 e2e. Global ui-tab-panel та B40 fresh-asset-id polling збережені. Інших U2–U6 змін немає.
FACT: Images/build/browser/e2e не запускалися за прямим дорученням власника; CSS computed display, реальний e2e focus і full e2e тут NOT_TESTED. Незалежний review тут NOT_RUN, U1 acceptance gate не закрито. Agents launched=0, returned=0. Комітів, staging та branch mutations немає; попередній WIP збережено.
NEXT: Owner Courses 1: взяти unstaged writer snapshot (hashes вище), перебудувати frontend на Kiri у своєму вікні, довести computed display трьох панелей × три selections, виконати повний e2e та незалежний read-only review U1; лише після цих доказів закрити U1. Відкритих QUESTION або змін contract немає.
NEXT: Owner extension m3/Ux03: перевірити чинні entry/local rules, admin splat і f73b1aec; зберегти parent guard та сьогоднішню unauthenticated поведінку, додати реальний NotFound для unmatched /admin/*, route-level unit і e2e assertion; повторити Verify.
FACT: Resume root/branch/HEAD незмінні (admin-editor-ux @10436cdf), попередній U1 diff присутній; локальні Kiri/permissions збережені. Owner browser finding прийнято як нове авторизоване розширення U1, а попередній synthetic NotFound test не доводить реальний admin splat.
NEXT: Трасувати f73b1aec і actual route-tree tests, щоб новий test проходив production route chain (guard→splat→root NotFound), а не окремий extracted component.
FACT: f73b1aec пояснює splat: успадковує parent identity guard і замінює старий LockedArea; видаляти splat не можна. admin.tsx beforeLoad спершу await requireIdentity, тоді role predicate. Порожній admin.$ не дає NotFound; route-guards.test.tsx має production routeTree + повні dependency fixtures.
DECISION: 2026-11-18, Codex writer, owner extension: зберегти /admin/$ і ssr:false; child beforeLoad кидає notFound() після успішного parent guard. Перевірка — actual route-tree admin 404 + anonymous returnTo та наявний guard suite. Попередній root-only test залишає локалізацію/path-boundary proof, але не є route-chain proof.
NEXT: Додати actual-tree Ux03 test і переконатися, що він падає до product fix; додати e2e heading+link assertion у наявний admin.spec case.
FACT: receiving-code-review прочитано з канонічного skills/arch шляху. Існуючий G06 уже доводить anonymous nested redirect; додам точний /admin/nope-xyz?from=ux03 case для збереження deep-link query, і окремий admin actual-tree NotFound test.
NEXT: Новий regression test спершу на незміненому splat, потім мінімальний beforeLoad fix.
FACT: Додано actual route-tree Ux03: /admin/nope-xyz для test admin вимагає root heading, localized /admin link, admin shell і відсутність course-list query. G06 доповнено anonymous/stale-admin-cache exact-path returnTo cases. Існуючий E1 admin e2e доповнено переходом на unknown path і heading+link assertions; попередні assertions збережені.
NEXT: Запустити targeted regression Ux03 на старому порожньому splat, очікується red.
FACT: Targeted red regression запущено; raw /tmp/courses-u1-splat-red.log. Product splat досі незмінений.
NEXT: Отримати red evidence перед beforeLoad fix.
NEXT: Зіставити targeted runner exit з першим failure у raw log; виправляти лише підтверджений empty-splat дефект.
FACT: Targeted regression — RED, exit 1: новий actual-tree test не знайшов heading; DOM містить admin shell та порожній main (відтворено owner finding). Raw /tmp/courses-u1-splat-red.log; 1 failed, 72 не відібрані через testNamePattern.
FACT: Мінімальний fix внесено в admin.$.tsx: ssr:false лишився, beforeLoad throws notFound(); parent /admin guard не змінено.
NEXT: Повторити Verify: format → format:check → typecheck → lint → повний rstest src, без images/e2e.
FACT: Verify extension запущено ordered chain; rstest raw /tmp/courses-u1-splat-final.log. Фактичні exits нижче після завершення.
FACT: Extension diff перевірено: лише child notFound(), actual-tree Ux03, G06 anonymous exact-path cases та existing E1 e2e assertions. Parent guard, ssr:false і решта U1 збережені. Scoped git diff --check — PASS, exit 0.
FACT: npm run format = PASS exit 0 (174ms); npm run format:check = PASS exit 0 (128ms), 416 files. Typecheck/lint/rstest результат звіряється за session output.
NEXT: Дочекатися остаточного Verify exit і зберегти final hashes для extension review.
FACT: Extension final snapshot: admin-editor-ux @10436cdf2988a17887c959dd48b363e6b117e9d7; наступні hashes доповнюють попередній snapshot (admin.spec hash замінено новим).
FACT: SHA-256 frontend/src/routes/admin.$.tsx: 60ac51508dc446d03c74fb8f2fb21e1afacad52a4fcde5195e9e852cc0ab5e29
FACT: SHA-256 frontend/src/routes/route-guards.test.tsx: 1645fd49b2e20f9223d4930962b0f4db6cf38481048dfb595386b59b60a28928
FACT: SHA-256 frontend/e2e/admin.spec.ts: 078f52270419561d121415b736df7ab70edbe263a03f7090ae5dcedc255df2ba
FACT: Extension npm run typecheck і npm run lint — PASS, exit 0 обидва (ordered chain дійшов rstest; architecture boundaries 244 files).
NEXT: Звірити повний rstest count/exit; browser/images/e2e лишаються owner-only.
NEXT: Завершити extension handoff лише після raw rstest verdict; aggregate gate не видавати за owner browser/e2e proof.
FACT: Extension final npx rstest run src --testTimeout 30000 — PASS, exit 0: 115 files, 1445/1445 passed, 0 failed/skipped/todo; 17003ms. Raw /tmp/courses-u1-splat-final.log (2026-11-18T11:51:28.787Z). Новий actual-tree regression перейшов RED→GREEN; anonymous empty/stale-admin cache redirects пройшли.
FACT: Extension Verify у заданому порядку: npm run format = 0; npm run format:check = 0; npm run typecheck = 0; npm run lint = 0; npx rstest run src --testTimeout 30000 = 0. Scoped git diff --check = 0.
FACT: Owner finding виправлено: unmatched /admin/* тепер кидає child notFound() після parent identity guard; admin splat та ssr:false з f73b1aec збережено. Route-level unit використовує production routeTree; existing admin E1 e2e має unknown-path heading і /admin-link assertion. Комітів/staging/git mutations немає; images/browser/e2e не запускалися. Agents launched=0, returned=0.
NEXT: Owner Courses 1: перебудувати frontend і повторити /admin/nope-xyz browser proof плюс повний e2e/independent review U1 на extension snapshot; авторський Verify PASS не закриває ці owner gates. Відкритих питань немає.
NEXT: Прочитати незалежний U1 review і оновлений Ux01 contract; додати e2e visibility proof для трьох course tabs через accessible relationship, зберігши unit. Запустити лише format/format:check/typecheck/lint; не чіпати Docker/e2e через активний owner run.
FACT: U1 independent review прочитано: Standards PASS, Specification BLOCK major через відсутній automated browser display proof. Оновлений Ux01 вимагає FE class+inert + E2E. Корінь/гілка/HEAD незмінні; наявні unit і product зміни зберігаю.
DECISION: 2026-11-18, Codex writer, owner-approved review fix: доповнити existing Ux02 fixture loop-ом трьох tab selections. Панелі знайти через tab aria-controls, перевірити їх існування та toBeVisible/toBeHidden (щоб inactive без role також перевірялися). CSS class у locator не використовувати.
NEXT: Внести loop після відкриття editor перед створенням module/lesson; далі наявний Ux02 flow сам обере Lessons і лишить cleanup незмінним.
FACT: Existing Ux02 e2e доповнено Ux01 matrix: Overview/Lessons/Access selected по черзі, кожна з трьох панелей має count=1 і expected visible/hidden. Panel IDs зчитуються через aria-controls коли tab активний (React Aria inactive role відсутній); locator за exact id, не CSS class. Залишено unit class+inert та весь Ux02 flow/cleanup.
NEXT: Запустити npm run format → npm run format:check → npm run typecheck → npm run lint; e2e/rstest/build/Docker не запускати.
FACT: Чотири дозволені frontend commands запущено ordered chain; фактичні exits будуть записані після завершення.
FACT: npm run format — PASS exit 0 (185ms, 416 files); npm run format:check — PASS exit 0 (118ms, 416 files). Final test delta переглянуто; scoped git diff --check — PASS exit 0.
NEXT: Зафіксувати typecheck/lint завершення та new admin.spec hash; передати changed test owner для наступного e2e/фокусованого re-review без claims runtime PASS.
FACT: S1 correction verification: npm run format = 0; npm run format:check = 0; npm run typecheck = 0; npm run lint = 0, виконані в цьому порядку. Lint architecture boundaries checked 244 files. Scoped git diff --check = 0.
FACT: Final SHA-256 frontend/e2e/admin.spec.ts = 29bfb4125ef11f237dc90f39bfe3fcb7a924b90b13879fa36fd47cf361883f0e. Ux01 unit file hash незмінний d60116b8752ac9c988de7fee4d68dfbf1a4067037fa60833fd70c465fe11e1de. Root/branch/HEAD: <repo>, admin-editor-ux, 10436cdf2988a17887c959dd48b363e6b117e9d7.
FACT: Додано всі 9 panel/selection visibility observations через aria-controls-derived exact ID; count=1 не дозволяє toBeHidden пройти через відсутню панель. Без course-editor CSS inactive Overview/Access при іншій selection залишаться visible — ці assertions це виявляють (статичний висновок, mutation/run не виконано).
FACT: E2E, Docker, browser, build, rstest не запускалися на S1 correction; активний owner e2e run не читався й не змінювався. Runtime proof і focused review цього нового test — NOT_TESTED/PENDING. Комітів/staging/branch mutations немає; agents launched=0, returned=0.
NEXT: Owner: після завершення поточного stack run виконати доданий Ux01 matrix на новому snapshot у своєму дозволеному e2e вікні, повернути frontend/e2e/admin.spec.ts delta на focused independent review S1. Цей writer не оголошує Specification BLOCK знятим без owner/reviewer evidence.
NEXT: Прочитати owner error-context і встановлений react-aria-components Tabs source; підтвердити lifecycle inactive panel IDs, замінити stale aria-controls locators стабільним structural panel scope без CSS під тестом. Запустити тільки format/format:check/typecheck/lint; e2e/Docker заборонені.
FACT: Owner error-context прочитано: при selected Overview збережений Lessons ID відсутній; помилка саме count=1 (не visibility). Попередній запис про стабільність ID був хибним припущенням; не оголошую CSS чи product регресією. Branch/HEAD незмінні. Owner повідомив full e2e 167 passed; я це не запускав і не перевіряв самостійно.
NEXT: Знайти shipped Tabs implementation (пакет не має раніше перевіреного src/Tabs.tsx) та прочитати точні рядки inactive prop removal до редагування.
FACT: Знайдено shipped implementation dist/private/Tabs.mjs і embedded TypeScript у Tabs.mjs.map, а не неіснуючий src/Tabs.tsx. Перевіряю саме цю встановлену версію перед зміною.
NEXT: Обрати structural direct children panels, зберігаючи selected panel proof через поточний aria-controls; inactive знайти незалежно від role/id/class/inert CSS.
FACT: Tabs source підтверджено до правки: dist/private/Tabs.mjs:276 force-mounted inactive panel лишається mounted; :308 delete DOMProps.id; :309 merge tabPanelProps (id/role/aria-labelledby) виконується лише isSelected; :316–317 inactive має inert/data-inert. Це не зміна ID, а його видалення з inactive DOM.
DECISION: 2026-11-18, Codex writer: scope = батько доступного course tablist; selected panel — current aria-controls після click, не cached ID. Inactive — лише direct inert children цього scope або його direct form (дві form panels + direct Access, без nested lesson panels); require count=2 і toBeHidden на обох. Селектор не використовує course-editor-panel class чи CSS display; видалення правила лишить inert elements і зламає hidden oracle.
NEXT: Прибрати stale-ID discovery loop, зберігши 3 selections × (1 visible + 2 hidden) і Ux02 flow/cleanup/unit.
FACT: Test fix внесено: лише один loop трьох selections; selected ID читається заново в поточній selection, дві inactive panels знаходяться structural+inert locator і count=2 перед hidden assertions. Cached IDs видалено. 9 visibility observations і всі попередні Ux02 checks збережені; unit не змінено.
NEXT: npm run format → npm run format:check → npm run typecheck → npm run lint (лише ці команди, без e2e/Docker).
FACT: Дозволені чотири commands запущено ordered chain; результати фіксую після завершення.
FACT: npm run format — PASS exit 0 (138ms, 416 files); npm run format:check — PASS exit 0 (106ms, 416 files). Фінальний test diff прочитано, scoped git diff --check — PASS exit 0; unit test hash лишився попереднім.
NEXT: Отримати typecheck/lint exit і final hash; owner rerun admin.spec залишається єдиним runtime доказом цієї correction.
FACT: Stale-ID correction Verify: npm run format = 0; npm run format:check = 0; npm run typecheck = 0; npm run lint = 0, у заданому порядку. Lint architecture checked 244 files; scoped git diff --check = 0.
FACT: Final SHA-256 frontend/e2e/admin.spec.ts = 3dfa6acd305d32375d22f835f303fa46efe0e4022600e1f5a4f75e0131825e48. Прочитаний dependency source dist/private/Tabs.mjs SHA-256 = 421ad27d139e31a1af9d7fc1c30c6a179f05788b711536dd52b0186f6a23dd54; dependency bytes не змінені. Unit SHA-256 незмінний d60116b8752ac9c988de7fee4d68dfbf1a4067037fa60833fd70c465fe11e1de.
FACT: Виправлення scoped лише admin.spec Ux01 loop: current active relationship + structural direct inactive panels; stale-ID collection прибрана. CSS oracle збережено: removing scoped rule не видаляє inert nodes, count=2 лишається, hidden assertions для Overview/Lessons selections падають за видимістю (статичне reasoning, mutation не виконано).
FACT: Не запускав e2e/Docker/browser/build/rstest; не чіпав stack або owner run. Browser PASS цієї correction ще NOT_TESTED. Комітів, staging, branch mutations немає; agents launched=0, returned=0.
NEXT: Owner повторює admin.spec на final test hash 3dfa6acd…1825e48 і передає delta для focused S1 review. Попереднє припущення про стабільні inactive IDs спростовано source+owner failure; попередній writer verification не є browser доказом.
