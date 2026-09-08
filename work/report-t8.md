ДАЛІ: прочитати інструкції implement, репозиторний брифінг і work/brief-t8.md повністю, потім визначити точний обсяг таску T8.
ФАКТ: скіл implement вимагає реалізувати бриф, застосувати TDD де доречно, регулярно запускати вузькі перевірки й typecheck, наприкінці виконати повну перевірку, code review та зробити коміт.
РІШЕННЯ: дотримуватися цього циклу, не виходячи за буквальний обсяг brief-t8.
ДАЛІ: прочитати work/CTO-RESEARCH.md і work/brief-t8.md повністю.
ФАКТ: work/brief-t8.md прочитано повністю. T8 обмежений frontend-навчальною поверхнею: зміст курсу біля плеєра, стани уроків, вхід через continueLessonId, докази P02/P03/B19/B26/R24/R36; backend, edge, міграції, читальна модель, T9/T10, generated-файли та залежності поза обсягом.
ФАКТ: середовище за брифом не має Docker, тому integration/build/up/e2e/check-generated не запускаються; доступні frontend format:check, lint, typecheck і unit-тести.
РІШЕННЯ: P02 доводити справжнім e2e-ланцюгом від heartbeat до повторного відкриття змісту; B19 обмежити твердженням про рендер 120 уроків; B26 порівнювати нормалізовані базові ключі й окремо мовні plural-форми та непорожність; R24 довести точним набором шляхів; R36 — окремим поведінковим тестом.
ДАЛІ: прочитати обовʼязкові розділи playback plan/spec, звіти T6/T7, work/tasks.md та еталон CourseEditorView; перевірити, чи контракт D-6 справді містить усе для змісту без другого запиту.
ФАКТ: усі названі секції й рядки знайдені. work/tasks.md позначає T8 `in progress`; план прямо відносить P02/P03/B19/B26/R24/R36 до цієї навчальної поверхні, а §12 містить пост-T7 D-9/O1/F-1, які треба врахувати без розширення T8.
ДАЛІ: прочитати точний зміст D-6, §8 із пʼятьма питаннями та потрібними рядками, §12, spec §2.2/навчальну поверхню, а також релевантні частини звітів T6/T7 і поточну frontend-реалізацію.
ФАКТ: D-6 підтверджує один read `GET /api/v1/learn/courses/{slug}` з повним деревом, video state, position, completed, clientSeq і continueLessonId; відсутній прогрес матеріалізується нулями/false. Другий запит для T8 не потрібен.
ФАКТ: D-7/D-8 і §12 підтверджують, що T7 уже володіє плеєром/відновленням/heartbeat та має відомі D-9/O1/F-1; T8 не повинен змінювати ці механізми. Пʼять питань вимагають доказу на production-origin артефакті та однореченнєвої умови почервоніння.
РІШЕННЯ: контракту вистачає; каналу «Стоп» не потрібно. Продовжити читання вузькими діапазонами, бо обʼєднаний вивід був обрізаний і не можна вдавати, що пропущене прочитано.
ДАЛІ: дочитати без обрізання P02/P03/B19/B26/R24/R36, A18–A21, §12 і Phase 6 task block; потім повністю прочитати report-t6/report-t7.
ФАКТ: названі рядки й A18–A21 дочитані без обрізання. P02 вимагає 2×2 уроки та точну пару «перший completed, другий current» після production heartbeat; P03 — course CTA до URL зі slug+lesson id; B19 — лише renderer boundary; B26 — base-key parity + locale-specific plurals + non-empty; R24 — exact 22-route set; R36 — authenticated `/learn` redirect to catalog.
ФАКТ: §12 не додає роботи T8: D-9 описує свідому нерівність native HLS, O1 лишається недоведеним до T9, F-1 уже виправлено в T7 review. Phase 6 tracker підтверджує вузький ownership T8.
РІШЕННЯ: не торкатися плеєрного recovery/heartbeat коду, крім використання наявного e2e production flow для P02.
ДАЛІ: повністю прочитати work/report-t6.md та work/report-t7.md, звертаючи увагу на фактичні frontend seams, тестові пастки й незакриті обмеження.
ФАКТ: прочитано report-t6 рядки 1–260. T6 ввів learning endpoint як окремий authenticated read, зберіг public catalog DTO незмінним, матеріалізує zero progress, агрегує video states і визначає continueLessonId за D-6; route composition уже існує.
ФАКТ: T6 зафіксував дві прийняті Discussion: детермінований continueLessonId незалежно від video readiness і learning projection, що віддає старий ready asset перед новішим pending. Для T8 це означає: UI має показувати всі уроки, а CTA довіряє server continueLessonId без клієнтського перевизначення.
ДАЛІ: дочитати report-t6 261–EOF, особливо generation delta, proof matrix, review findings і reviewer verification.
ФАКТ: report-t6 прочитано до EOF. Рецензентський повний `make check` був чистий (334 frontend tests, e2e 35/35), generated client матеріалізовано; R17 виправлено через справжній write path. T8 може спиратися на завершений T6 контракт, не повторюючи його proofs.
ФАКТ: T6 review урок релевантний P02: read-only/саморобний артефакт був хибнопозитивом для write claim; тому completed-позначку T8 не можна доводити component fixture.
ДАЛІ: прочитати report-t7 повністю чотирма діапазонами; витягти фактичний маршрут, компонент, e2e fixtures/heartbeat machinery і тестові відмови, які не можна повторити.
ФАКТ: report-t7 1–250 прочитано. T7 обрав `/learn/$slug/$lessonId`, додав request-scoped playback repository, `LessonPlayer`, hls/native adapter, isolated Playwright `playback` project і реальну media fixture; learning data вже входить у route. P02 може розширити цей serial e2e без нового backend/API seam.
ФАКТ: T7 мав недостовірні/заблоковані ранні прогони (Rstest bind EPERM, test fixture/lifecycle помилки); для T8 треба відрізняти product assertion від runner failure й не зараховувати незапущене.
ДАЛІ: прочитати report-t7 251–500, зокрема review findings і остаточну структуру route/e2e.
ФАКТ: report-t7 251–500 прочитано. Остаточний route preload використовує один request-scoped learning query і перевіряє membership lesson id; presentation не fetch-ить. Route shell лишається generated-safe `/learn/$`.
ФАКТ: T7 e2e створює власний published course, використовує реальний 5-minute media upload, admin/student storage states і serial project. Review кілька разів посилював докази до конкретних segment/heartbeat observations; T8 P02 слід вбудувати в той самий production-origin setup.
ДАЛІ: прочитати report-t7 501–750, включно з фінальними тестами, review та gates.
ФАКТ: report-t7 501–750 прочитано. Остаточна T7 реалізація має route/parser/query tests, component suite, e2e media journey й accessibility; вона також зберігає попередню `/learn` empty-state поведінку для identity contract.
ФАКТ: T7 runner failures показали, що unit bind EPERM може бути intermittent; валідний повний `npm test` усе ж запускався й викривав справжні regression. Для T8 кожну відмову класифікуватиму за точним output, не за припущенням.
ДАЛІ: прочитати report-t7 751–EOF, включно з reviewer-run хронологією та остаточним станом O1/документованих Discussions.
ФАКТ: report-t7 прочитано до EOF. Рецензент підтвердив T7 `make test` 364/364, format/lint/typecheck/check-generated/build і e2e 42/42 після тестових/інфраструктурних виправлень; T7 прийнято. O1 лишається відкритим і не належить T8.
ФАКТ: критична пастка для P02 — `uploadReady` колись чекав ready на уроці, а не на конкретному videoId, і зеленів на старій генерації. T8 не повинен додавати подібний самоздійснюваний wait; progress доказ має читатися з UI після справжнього heartbeat.
ДАЛІ: прочитати CourseEditorView.tsx повністю як React Aria взірець, потім інвентаризувати поточні learn/catalog/routes/locales/tests/e2e файли та git worktree без зміни git-стану.
ФАКТ: CourseEditorView.tsx 1–360 прочитано. Взірець використовує React Aria primitives, семантичні section/fieldset/legend, локалізовані accessible names і стабільні domain ids як keys; жодної presentation-side data fetching.
ДАЛІ: дочитати CourseEditorView.tsx 361–EOF, потім перейти до поточної learn/catalog реалізації.
ФАКТ: CourseEditorView.tsx прочитано до EOF. Взірець підтверджує composition через action props, локальні pending/status/alert стани, семантичну ієрархію heading/section і React Aria Link/Button; T8 наслідуватиме ці seams без form-level complexity.
ДАЛІ: зняти inventory актуальних learn/catalog route/component/test/style/locale/e2e файлів і чисто прочитати робочий diff/status, щоб не перезаписати чужі зміни.
ФАКТ: worktree до T8 уже має чужу модифікацію `frontend/e2e/media.setup.ts`; її не торкатимусь і не включатиму до роботи. `work/report-t8.md` — новий файл цього таску.
ФАКТ: наявні seams: `LearningCourse` уже містить continueLessonId і впорядковані modules/lessons; `/learn/$` рендерить LessonPlayer; course detail route існує окремо; playback e2e уже відкриває реальний lesson URL. Generated-файли не потребують змін.
ДАЛІ: прочитати повністю domain playback, learn route/query/composition, course detail view/route, locales, UI Link, relevant styles/tests і playback e2e; після цього сформувати RED-тести T8.
ФАКТ: domain/learn route/query/composition і course detail route прочитані. Lesson route вже має рівно один learning query й знаходить current lesson з того самого response; T8 зміст можна передати тому самому presentation-компоненту без fetch.
ПИТАННЯ: public course page не має lesson ids/continueLessonId. Найменший спосіб не заводити паралельний другий read на сторінці — CTA на `/learn/{slug}`, де existing protected learn route одним T6 read отримує continueLessonId і redirect-ить на `/learn/{slug}/{lessonId}`. Це узгоджується з P03 та зберігає public catalog model.
РІШЕННЯ: перевірити цей single-read redirect seam у наявному parser/router API; якщо він коректний, реалізувати саме його, а не змішувати public і authenticated data у CourseView.
ДАЛІ: прочитати CourseView/tests, learn-route parser/tests, route test conventions, locales й playback e2e повністю.
ФАКТ: CourseView зараз чесно показує «навчання незабаром» і посилання назад у каталог; T8 має замінити availability CTA на локалізоване «Почати навчання» до `/learn/{slug}`. Public program навмисно не містить lesson links/ids.
ФАКТ: `parseLessonRoute` зараз приймає лише slug+UUID; slug-only вважається fallback. Щоб CTA реально вів за server continueLessonId, потрібен окремий явний parser результат для course-entry, не послаблення lesson parser.
РІШЕННЯ: додати discriminated route parsing: `{kind:'course', slug}` для одного сегмента та `{kind:'lesson', slug, lessonId}` для двох. `beforeLoad` для course-entry виконує той самий learning query й router redirect за continueLessonId; lesson-entry лишається без змін.
ДАЛІ: перевірити TanStack redirect conventions у репо, UI Link API, локалі та e2e setup/heartbeat точними читаннями.
ФАКТ: UI `Link` — React Aria link із ≥існуючим semantic styling; locale objects — незалежні `as const`, нині en 274/uk 278 keys з очікуваними українськими plural extras. T8 має додати copy симетрично й окремий автоматичний B26 traversal test.
ФАКТ: у course locale досі лишився pre-playback «coming soon» текст; буквальний T8 вимагає замінити його CTA copy, не заявляючи payments/ownership.
ДАЛІ: прочитати TanStack redirect API usage/type dependency, повний playback.spec.ts і relevant styles; потім сформувати test-first slices для learning surface, route entry, CTA, B26 і R24.
ФАКТ: playback.spec.ts 1–260 прочитано. Setup має лише один модуль/урок; P02 потребує розширити fixture до 2×2. `openAndPlay` уже проходить production mint і чекає зростання currentTime, але для completed=true треба відправити heartbeat на ≥95% duration і потім відкрити другий урок.
РІШЕННЯ: P02 test має не читати БД і не інʼєктити completed; він використає справжній UI seek + реальний heartbeat response, reload/відкриття course-entry, після чого асертить рівно 4 lesson links, completed marker на першому й current state на другому.
ДАЛІ: дочитати playback.spec.ts 261–EOF і перевірити реальну тривалість/heartbeat helpers та послідовність незворотної archive.
ФАКТ: playback.spec.ts прочитано до EOF. Student block завершується незворотною архівацією, тому P02/P03 мають стояти до R04; fixture cleanup ізольований. Heartbeat profile 2s/6s робить completion proof дешевим.
ФАКТ: існує потенційна brief-неточність: P02 передумова каже 2×2 й «другий підсвічений» після завершення першого, але `continueLessonId` підсвітить саме другий лише якщо інші три уроки справді існують у program order. Це узгоджено — fixture треба розширити, не підміняти response.
ДАЛІ: прочитати routeTree export/route tests/locales tests/styles і вирішити точну R36 поведінку за планом та попередніми identity contracts.
ФАКТ: generated route tree має 11 route nodes (і 22 explicit `path`/`fullPath` declarations), не 22 distinct routes. R24 proof має порівнювати exact runtime full-path set із 11 очікуваними routes; це сильніше й не плутає повторені declarations із маршрутами.
ФАКТ: чинні identity unit/e2e очікують `/learn` empty shell, але T8 R36 навмисно змінює продуктову поведінку на каталог. Ці попередні assertions треба оновити до redirect, а invalid nested `/learn/go/intro` лишити shell fallback для G06.
РІШЕННЯ: R36 реалізувати як authenticated redirect `/learn` → `/courses`, із окремим browser test; anonymous `/learn` і далі спершу йде на login через parent guard.
ДАЛІ: прочитати повний i18n test, query cache helper/defaults і runtime route-tree traversal API; потім написати RED unit tests.
ФАКТ: frontend-design прочитано повністю. Артефакт — system; аудиторія — студент; одна робота — швидко зорієнтуватися в програмі й перейти/повернутися до потрібного уроку. Закриті осі: palette/type/radius/spacing з `tokens.css`, React Aria inventory з `presentation/ui`, light/dark через existing CSS variables, parity — playback A19/A20 та B19.
РІШЕННЯ: композиція — двоколонкова «плеєр + curriculum rail», що стає однією колонкою на вузькому екрані. Signature — повторюваний номер модуля/уроку та тонка status-мітка в кожному рядку; без нових кольорів, шрифтів, тіней чи анімацій.
ДАЛІ: через вимогу frontend-design прочитати WRITING.md перед новим UI copy; потім дописати RED-тести.
ФАКТ: WRITING.md прочитано повністю. Новий copy буде дієвим і послідовним: «Почати навчання», «Зміст курсу», «Переглянуто», «Поточний урок»; без технічних термінів і без обіцянок оплат/списку власних курсів.
ДАЛІ: додати RED-тести пʼятьма вузькими seams: CourseLearningView (order/states/B19), course CTA, slug-entry redirect/P03, B26 locale traversal, R24 exact generated route entries; окремо оновити очікування R36.
ФАКТ: додано перший RED-пакет: renderer order/states + B19 120 rows, course-page CTA href і course-only parser для P03. Production-компонента та parser ще немає; existing fixtures не підміняють P02.
ДАЛІ: запустити лише три touched test files і підтвердити, що RED спричинений відсутніми T8 production exports/copy, а не тестовою інфраструктурою.
ФАКТ: перша targeted команда не запустила Rstest: CLI не підтримує прапорець `--run`. Це помилка команди, не RED і не product failure.
ДАЛІ: перевірити `rstest --help` і повторити точну підтримувану file-filter команду.
ФАКТ: Rstest 0.11.11 приймає positional filters після `rstest run`; npm script уже містить `rstest run`, тому правильна форма — `npm test -- <filters>`.
ДАЛІ: повторити три targeted tests правильно.
ФАКТ: RED підтверджено коректним прогоном: 17 тестів, 15 pass, 2 assertion failures + 1 missing-module build failure. Причини рівно T8: немає CourseLearningView, parseCourseRoute і нового course CTA/copy; runner працює.
ДАЛІ: додати B26 і R24 acceptance proofs та R36 route expectations до RED/characterization пакета перед production GREEN.
ФАКТ: додано B26 recursive proof по всьому locale tree: base-key parity після всіх шести plural suffixes, `Intl.PluralRules` для мовних форм, жодних порожніх значень. Це замінює ручне перерахування як доказ «усі».
ФАКТ: додано R24 exact 22-entry `(path, fullPath)` proof із runtime generated tree та змінено G03 на майбутню R36 поведінку `/learn` → `/courses`; test catalog fake тепер віддає реальний empty page замість unrelated dependency failure.
ДАЛІ: запустити i18n + route-guards tests; очікувати B26/R24 green characterization і R36 red на чинному production.
ФАКТ: targeted прогін виконав 35 тестів: B26 пройшов; R36 очікувано RED для трьох ролей (`/learn` лишився `/learn`). R24 упав через неправильне очікування тесту: runtime TanStack `route.path` нормалізований без початкового `/`, тоді як `fullPath` абсолютний; generated source literals і runtime values не тотожні.
РІШЕННЯ: R24 порівнюватиме точні 22 runtime entries, включно з нормалізованими relative `path` та absolute `fullPath`; це все одно червоніє від будь-якого нового route node і стоїть на generated tree.
ДАЛІ: виправити лише R24 expected list і повторити його окремо; R36 лишити RED до production change.
ФАКТ: R24 exact generated-tree test GREEN окремо: 1 pass, 30 filtered/skipped. Додавання будь-якого route node змінить дві entries (`path`, `fullPath`) і почервонить exact list.
ДАЛІ: реалізувати мінімальний GREEN: CourseLearningView + CSS, course CTA/copy, course-only parser/redirect, same-query curriculum wiring і R36 parent redirect; не змінювати backend/generated/heartbeat.
ФАКТ: мінімальний T8 production slice додано: CourseLearningView рендерить ordered module/lesson links і visible completed/current states; lesson route обгортає existing player тим самим course object; course CTA веде на slug-only learn entry; route redirect використовує server continueLessonId; `/learn` redirect-ить каталог; en/uk copy й responsive rail CSS додані.
РІШЕННЯ: нових API calls немає: slug-entry та redirected lesson-entry використовують один query key і 60s cache; lesson screen отримує curriculum із того самого response.
ДАЛІ: перевірити нові CSS token names проти `tokens.css`, форматувати slice, потім typecheck до тестів, щоб знайти route redirect typing і strict TS проблеми.
ФАКТ: первинний CSS audit знайшов чотири неіснуючі token names (`space-5/10`, `surface-hover`, `primary-soft`). Це був би silent styling defect; замінено на чинні `space-6/8` та `color-background`, не додаючи нової палітри.
ДАЛІ: форматувати всі T8 source/test files і запустити frontend typecheck.
ФАКТ: oxfmt завершився успішно на 13 T8 frontend files.
ДАЛІ: запустити `npm run typecheck`; виправляти лише strict typing/route контрактні помилки.
ФАКТ: `npm run typecheck` PASS, exit 0. TanStack internal redirect до `/learn/$` і всі strict TypeScript mappings валідні без assertions/any.
ДАЛІ: запустити focused GREEN tests для CourseLearningView, CourseView, learn parser, i18n, route guards; перевірити фактичні accessibility names і route behavior.
ФАКТ: focused GREEN PASS: 54/54 у пʼяти файлах. Доведено renderer order/states, B19=120 rows, course CTA, parser, B26, R24 і R36 для трьох ролей; попередні route-guard assertions у цьому наборі не регресували.
ДАЛІ: розширити справжню playback e2e fixture до 2×2 й додати P03/P02 production-chain tests; оновити старі `/learn` browser expectations під окремий R36 test.
ФАКТ: e2e fixture тепер має 2 модулі × 2 уроки; upload target однозначно бере першу duplicate-named group. Додано P03 course CTA з exact URL і рівно одним learning read; P02 чекає 204 heartbeat із lessonId першого та position≥285, потім відкриває slug-entry й перевіряє 4 rows, first completed, second current. R36 має окремий named e2e.
ДАЛІ: форматувати e2e files, typecheck, потім статично audit serial ordering і всі залишкові `/learn` expectations.
ФАКТ: e2e files format PASS; повторний `npm run typecheck` PASS. Нові Playwright request-body guards проходять strict TS без unsafe external-data assertions.
ДАЛІ: audit усіх `/learn` browser/unit expectations, e2e serial order і query invalidation behavior; виправити лише stale T8 expectations або виявлені proof defects.
ФАКТ: audit підтвердив правильний serial order P03 → P02 → playback rows → R04 archive; invalid nested learn paths досі показують legacy shell за G06. Єдина stale browser expectation була Back після нового R36 redirect: history тепер повертає каталог, а не колишній `/learn` shell.
РІШЕННЯ: security assertion оновлено без послаблення: Back лишається на каталозі, локальний session уже очищений, імʼя відсутнє; anonymous protected `/learn` окремо й далі перевіряється login redirect.
ДАЛІ: форматувати route-guards e2e, запустити typecheck і повний `npm test`; потім inspect diff/architecture and improve proof gaps.
ФАКТ: route-guards e2e format PASS; `npm run typecheck` PASS; повний `npm test` PASS — 44 files, 369/369 tests, 0 skipped/todo, exit 0.
ДАЛІ: статично переглянути весь T8 diff без чужого `media.setup.ts`, перевірити architecture/no-fetch/no-second-query, стилі, R24/B26 strength і e2e proof на хибнопозитиви.
ФАКТ: статичний audit підтвердив zero fetch у presentation/routes і один learning query key. Знайдено дві proof/layout можливості: R24 string multiset не ловив би перестановку path між route nodes; третя grid-колонка state могла стискати довгі lesson titles при zoom.
РІШЕННЯ: R24 посилено до exact 11-row `{id,path,fullPath}` mapping (22 path fields), а lesson state перенесено під title у двоколонковому row. Поведінка й palette незмінні.
ДАЛІ: форматувати ці правки, повторити R24/CourseLearningView targeted tests і typecheck.
ФАКТ: post-audit targeted suite PASS 33/33; typecheck PASS. R24 mapping і renderer boundary лишаються зеленими після посилення.
ФАКТ: code-review skill прочитано повністю. Через відсутній user-supplied ref і uncommitted slice baseline є початковий HEAD + явний T8 file set; чужий `frontend/e2e/media.setup.ts` виключено. Два read-only review agents запущені паралельно для Standards і Spec, як прямо вимагає skill.
ДАЛІ: поки review триває, локально перевірити e2e proof chronology, CSS accessibility, route query count semantics і `git diff --check`; потім виправити blocking findings.
ФАКТ: `git diff --check` PASS; `npm run lint` PASS з zero warnings; architecture checker PASS на 90 files. Presentation/routes не містять fetch, generated/backend/dependencies не змінені.
ФАКТ: screenshot не знято: Docker/Compose за brief недоступний, а навчальний route потребує authenticated API + real learning response; unit DOM не є правдивим layout proof. Візуальну browser-перевірку не видаю за виконану.
ДАЛІ: отримати обидва review verdicts; до них локально перевірити всі T8 files/status і сформувати пʼять відповідей для кожного acceptance row.
ФАКТ: T8 working set містить 13 modified і 2 new frontend files плюс append-only report; старі locale keys більше ніде не споживаються. Початкова чужа модифікація `media.setup.ts` більше не присутня в status, але T8 її не редагував.
ФАКТ: перше очікування review завершилось без verdict; обидві осі ще працюють, результат не припускаю.
ДАЛІ: сформувати чернетку acceptance proof matrix із пʼятьма відповідями, потім звірити її з review findings перед append.
ФАКТ: друге очікування тривалістю 120 с також завершилось без verdict; review не зараховано як PASS.
ДАЛІ: перевірити статус review-агентів і паралельно скласти точні reviewer commands/незапущені gates.
ФАКТ: Попередня append-спроба не змінила файл через застарілий anchor; tail підтвердив, що append-only журнал цілий.
ФАКТ: Під час роботи `work/brief-t8.md` зовнішньо розширився зі 154 до 235 рядків; поточну версію прочитано повністю. Вона додала обов’язкову перевірку першого малого heartbeat і точного поля learning-read `position`, окремий R37 для clamp/NULL duration та SQL-зріз після першого реального T8-прогону.
РІШЕННЯ: Поточна 235-рядкова версія брифу є джерелом істини; попередню реалізацію не вважати завершеною, доки нові вимоги й зауваження рев’ю не закриті.
ФАКТ: Standards-review не знайшов порушень стандартів; залишив два smells — дубль нормалізації plural-ключів і невикористаний `data-completed`. Spec-review знайшов три суттєві прогалини: seek 299 може завершити медіа до heartbeat, P02 не доводить перший малий heartbeat → exact `position`, а P03 на першому уроці може дати хибнопозитивний результат.
ДАЛІ: Прочитати чинний опис R37 у плані й пов’язані фрагменти, потім виправити тести та дрібні smells без зміни cadence/event implementation.
ФАКТ: Чинний R37 вимагає окремий від P02 урок із ready-відео 300 с, heartbeat `positionSeconds=100000` і learning-read `position=300`; 100000 у відповіді означав би NULL `duration_seconds`, тобто окрему несправність. Планова таблиця T8 ще не перелічує R37, але §8 і чинний бриф прямо включають його в цей slice.
РІШЕННЯ: R37 реалізувати на першому уроці другого модуля з власним uploaded asset; P02 лишити на першому уроці першого модуля. Це ізолює completion side effect і не змінює production cadence/event code.
ДАЛІ: Переглянути поточний playback e2e та test helpers, щоб вбудувати exact-position sanity і R37 у справжній API/UI ланцюг без unsafe casts.
ФАКТ: Playback fixture зараз завантажує ready asset лише для першого уроку; P03 виконується до прогресу, P02 слухає тільки completion heartbeat і seek-ить на 299. Helper `uploadReady` жорстко прив’язаний до першого lesson id. B26 має дві regex-нормалізації plural suffix, а `CourseLearningView` має невикористаний `data-completed`.
РІШЕННЯ: Узагальнити upload helper для конкретного lesson/group, додати learning-response parser для exact `position`, переставити P03 після P02 та прибрати обидва smells. Спочатку перевірити fixture/runtime межі й точну форму learning response.
ДАЛІ: Переглянути fixture inventory, OpenAPI learning schema та нумеровану ділянку e2e навколо setup, щоб вибрати детермінований другий ready-урок і не перевищити hook timeout.
ФАКТ: Єдина 300-секундна fixture — `e2e-1280x720-5min.mp4` (28 MiB); менші fixtures не відповідають R37. Setup має один 420-секундний hook і зараз транскодує один asset; другий 300-секундний asset потребуватиме збільшеного hook timeout. Learning domain/adaptor свідомо мапить wire-поле `position`, тож e2e parser має читати саме його.
РІШЕННЯ: Два 300-секундні uploads виконати послідовно з явним 720-секундним beforeAll timeout; це дорожче, але буквально доводить known-duration clamp на окремому уроці. R37 heartbeat зробити прямим API-запитом після mint, щоб браузерний range input не обрізав 100000 до 300 до сервера.
ДАЛІ: Звірити точні play/heartbeat request schemas і same-origin headers, потім внести один цілісний patch у e2e та два cleanup patches.
ФАКТ: OpenAPI підтвердив play body `{lessonId}`, heartbeat body `{playbackId, lessonId, positionSeconds, clientSeq, reason}`, session response містить seed `clientSeq`, а learning response — `position`. Domain clamp виконується до completion check. Це дозволяє R37 mint → heartbeat(seq+1, 100000) → GET learning без UI-range втручання.
ДАЛІ: Перечитати брифову Discussion-підказку про «один запит» проти sanity-read, зафіксувати трактування до зміни тесту.
ФАКТ: Бриф забороняє другий product call лише якщо змісту бракує даних; D-6 вже віддає `position`. Додатковий `page.request.get` у e2e є тестовим спостереженням persistence contract, а не другим запитом компонента; production route і далі використовує один cached learning query.
РІШЕННЯ: Exact-position sanity читати тестовим API client усередині того самого P02 сценарію; production код і query topology не змінювати.

## Discussion

ФАКТ: У `docs/playback/plan.md` є дві актуальні неузгодженості: §8 P03 досі формулює результат як «перший урок», тоді як D-6 і чинний бриф вимагають server-selected `continueLessonId`; §11 T8 перелічує лише P02/P03/B19/B26/R24, хоча §8 і бриф уже покладають R37 на T8. Це може спрямувати наступного виконавця до хибнопозитивного P03 та загубити R37 у phase accounting.
РІШЕННЯ: Для коду й доказів застосувати конкретніші чинні вимоги D-6/brief: P03 перевіряє non-first `continueLessonId` після реального P02 progress; R37 входить у T8. План не редагувати, бо бриф забороняє docs/playback зміни; винести неузгодженість рецензенту тут.
ДАЛІ: Внести e2e patch: general upload, session seed, safe wire parsers, first-heartbeat sanity, seek 290, isolated R37 і non-first P03; потім cleanup plural/data attribute.
ФАКТ: Patch внесено. P02 тепер ловить перший реальний малий heartbeat, читає wire `position` через ручне звуження `unknown`, порівнює точне значення, далі seek-ить 290 і чекає completion heartbeat. R37 має другий 300-секундний ready asset на третьому уроці та прямий mint → heartbeat(100000, seed+1) → learning `position=300`. P03 виконується після прогресу й очікує non-first secondLessonId при одному UI learning GET. Plural regex об’єднано, `data-completed` вилучено.
ДАЛІ: Форматувати змінені файли, запустити focused unit/typecheck і статично перевірити e2e ordering/request tracking на нові хибнопозитиви.
ФАКТ: `npm run format -- e2e/playback.spec.ts src/app/providers/i18n.test.tsx src/presentation/learn/CourseLearningView.tsx` PASS, exit 0 (formatter фактично перевірив/обробив 162 frontend files).
ДАЛІ: Запустити focused unit tests для B26/B19/R24/R36 та `npm run typecheck`; browser e2e не запускати без Docker за брифом.
ФАКТ: Focused unit suite PASS — 4 files, 45/45 tests, 0 skipped/todo, exit 0. B26, B19, exact R24 mapping, R36 role matrix і learn parser зелені після правок.
ДАЛІ: Запустити strict TypeScript check, який також компілює Playwright helpers і нові unknown guards.
ФАКТ: `npm run typecheck` PASS, exit 0; нові e2e parsers, session `clientSeq` і direct API payload strict-типізовані.
ДАЛІ: Переглянути formatted e2e diff і git status; перевірити, що formatter не зачепив позаскопові файли та що P02→R37→P03 serial state дає non-first continue без взаємного забруднення доказів.
ФАКТ: Status містить лише T8 frontend files і report; formatter не додав позаскопових змін. Serial state ізольований: P02 завершує lesson 1; R37 завершує окремий lesson 3; lesson 2 лишається першим незавершеним, тому P03 мусить одержати саме non-first lesson 2. R37 використовує session seed+1, отже не залежить від cadence counter.
ФАКТ: Статичний diff підтвердив ручне звуження wire `position` без generated type/assertion і product route без додаткового fetch. Browser/DB поведінка ще не виміряна через відсутній Docker.
ДАЛІ: Передати виправлений delta обом review agents для повторної Standards/Spec перевірки; паралельно підготувати acceptance proof matrix та точні reviewer commands.
ФАКТ: Spec-review agent отримав follow-up на актуальний 235-рядковий бриф і виправлений P02/R37/P03 delta; verdict ще не отримано.
ДАЛІ: Передати той самий delta Standards-review agent із фокусом на strict TS, architecture і test determinism.
ФАКТ: Standards-review agent отримав follow-up на виправлений delta; verdict ще не отримано.
ДАЛІ: Звірити решту чинного брифу для обов’язкового формату acceptance table, локальних/незапущених gates і reviewer commands.
ФАКТ: Бриф буквально вимагає локально лише `format:check`, `lint`, `typecheck`, `npm test`; Docker gates `integration`, `build`, `up`, `e2e`, `check-generated` треба назвати незапущеними. Acceptance-звіт мусить відповісти на п’ять питань для кожного оголошеного рядка; B19 не можна видавати за API-доказ.
ДАЛІ: Додати acceptance proof matrix для P02, P03, B19, B26, R24, R36, R37 із чесним статусом: unit докази виміряні, e2e докази реалізовані, але локально не запущені.

## Рядок §8 → чим доводиться

| Рядок | Доказ / локальний статус | 1. Що саме доводить | 2. Той самий production path | 3. Чи можливий зелений не з тієї причини | 4. Саме цей рядок | 5. Що мусить зламатись, щоб тест почервонів |
|---|---|---|---|---|---|---|
| P02 | `frontend/e2e/playback.spec.ts`, тест P02; реалізовано, Docker-e2e не запущено | Реальний UI play дає малий heartbeat, persistence повертає exact `position`, seek 290 дає completion, повторний вхід показує 4 rows, completed first і current second; це весь ланцюг, не props | Так: run-scoped course створений admin API, asset пройшов UI upload/transcode/publish, студент використовує production play/heartbeat/learning endpoints | Свіжий namespace, exact first small position і окремий ≥285 heartbeat не дають наперед засіяти completed prop/рядок; окремий R37 не торкається lesson 1 | Так, цей тест один закриває P02; B19 лише renderer capacity, P03 лише entry | Якщо heartbeat не збережеться/не завершить lesson 1 або learning UI переплутає completed/current/order, P02 почервоніє |
| P03 | `frontend/e2e/playback.spec.ts`, тест P03; реалізовано, Docker-e2e не запущено | CTA з course detail використовує server `continueLessonId`, формує exact slug+lesson URL і робить один UI learning GET | Так: той самий published course і реальний P02 progress роблять lesson 2 першим незавершеним | Очікування non-first lesson 2 не дозволяє hardcoded-first реалізації позеленіти; request counter ловить другий UI read | Так, P03 не покладається на R36 чи outline assertions | Якщо CTA ігноруватиме `continueLessonId`, втратить slug/id або заведе другий UI read, P03 почервоніє |
| B19 | `CourseLearningView.test.tsx`, B19; PASS у focused suite | Renderer приймає й повністю рендерить 120 lessons без падіння; не доводить, що API приймає 120 | Ні: fixture навмисно створена на renderer boundary, повз API; межу твердження зафіксовано | Так, API міг би відхилити 120 при зеленому тесті, але це поза оголошеним renderer-only твердженням | Так, лише B19 renderer capacity | Якщо view обрізатиме список, падатиме або втратить lesson rows, count 120 не зійдеться |
| B26 | `i18n.test.tsx`, recursive B26; PASS у focused suite | Повні locale trees мають однакові base keys, усі locale-specific plural categories й жодного порожнього value | Так: тест обходить безпосередньо production `en.ts`/`uk.ts`, не копію ключів | Наївний false red усунуто base-key normalization; false green від порожнього value чи пропущеної language category має окремі assertions | Так, один recursive тест закриває повноту B26 | Якщо зникне base key/обов’язкова plural form або value стане порожнім, B26 почервоніє |
| R24 | `route-guards.test.tsx`, exact generated mapping; PASS у focused suite | Runtime generated route tree дорівнює повному expected набору 11 `{id,path,fullPath}` rows / 22 path fields і не має my-courses route | Так: читається справжній згенерований route tree, який монтує застосунок | Exact row mapping ловить не лише додавання, а й перестановку path між nodes; `check-generated` окремо стереже drift, але локально не запущений | Так, це structural R24; redirect `/learn` винесений у R36 | Якщо додати `/learn/my-courses`, змінити/переставити route path або node mapping, R24 почервоніє |
| R36 | `route-guards.test.tsx` role matrix PASS; `e2e/route-guards.spec.ts` browser proof не запущено | Signed-in `/learn` переходить у public catalog; unit перевіряє student/instructor/admin, e2e — student observable behavior | Unit монтує production router з test dependencies; e2e використовує реальний student storage state і browser route | Exact catalog URL/content не дозволяє видати відсутність child route за redirect behavior | Так, окремо від structural R24 | Якщо parent `/learn` перестане редиректити signed-in actor у `/courses`, R36 почервоніє |
| R37 | `frontend/e2e/playback.spec.ts`, тест R37; реалізовано, Docker-e2e не запущено | Heartbeat із 100000 на окремому known-300s lesson приймається 204, а learning model повертає exact `position=300`, відрізняючи clamp від NULL duration | Так: окремий asset проходить admin UI upload, worker transcode і publish; student mint/heartbeat/read ідуть production HTTP endpoints із session seed+1 | Окремий lesson не засіває P02; exact sent 100000 + returned 300 не дозволяє позеленіти при NULL duration або client-seq rejection | Так, R37 має власний asset, progress і assertions | Якщо duration не персиститься, server не клемпить/не записує або learning read повертає не 300, R37 почервоніє |

ДАЛІ: Звірити Playwright project names і Makefile targets, потім додати точні reviewer commands та перелік незапущених gates.
ФАКТ: Повторний Spec-review повернув PASS: виправлений P02/R37/P03 delta відповідає чинному брифу без blocking/non-blocking findings.
ДАЛІ: Дочекатися Standards verdict після звірки project/target names; обидва результати внести до фінального review status.
ФАКТ: Playwright project `playback` залежить від `media` і `playback-profile`, а `chromium` — від shared `setup`; тому надійна reviewer-команда — повний `make e2e`/`make check`, не targeted `--grep`, який ризикує відфільтрувати setup dependencies. `make check` має порядок format-check → lint → typecheck → test → integration → check-generated → build → up → seeds → smoke → e2e.
ДАЛІ: Додати незапущені gates, точні host/Docker/SQL reviewer commands і пояснення, що SQL-вимір не отримано локально.

## Перевірка та команди рецензента

ФАКТ: Через відсутній Docker локально НЕ запускались `make integration`, `make check-generated`, `make build`, `make up`, `make smoke`, `make e2e` і повний `make check`. Відповідно P02/P03/R36/R37 browser verdict, worker transcode та SQL-зріз не видаються за PASS.
ФАКТ: SQL `SELECT state, duration_seconds IS NULL, count(*) FROM video_assets GROUP BY 1,2` локально не виконувався; результат невідомий, бо першого реального T8 Docker-прогону не було.

Точні host-команди, які виконуються тут:

```sh
cd frontend
npm run format:check
npm run lint
npm run typecheck
npm test
```

Точна повна команда для рецензента з Docker (із кореня репо):

```sh
make check
```

Точна команда SQL після першого реального прогону:

```sh
docker compose exec -T postgres sh -c 'psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c "SELECT state, duration_seconds IS NULL, count(*) FROM video_assets GROUP BY 1,2"'
```

Якщо рецензент запускає кроки окремо, порядок і команди з кореня репо:

```sh
make format-check
make lint
make typecheck
make test
make integration
make check-generated
make build
make up
make seed-identities
make seed
make smoke
make e2e
```

ДАЛІ: Виконати фінальні дозволені gates строго в порядку `format:check` → `lint` → `typecheck` → `npm test`, append-нути кожен результат окремо, потім завершити review/status audit.
ФАКТ: Фінальний `npm run format:check` PASS — 162 files, exit 0.
ДАЛІ: Запустити `npm run lint` другим дозволеним gate.
ФАКТ: Фінальний `npm run lint` PASS — zero warnings; architecture boundaries checked 90 files; exit 0.
ДАЛІ: Запустити `npm run typecheck` третім дозволеним gate.
ФАКТ: Фінальний `npm run typecheck` PASS, exit 0.
ДАЛІ: Запустити повний `npm test` четвертим і останнім дозволеним gate.
ФАКТ: Фінальний `npm test` PASS — 44 files, 369/369 tests, 0 failed/skipped/todo, exit 0.
ДАЛІ: Отримати Standards re-review verdict, виконати фінальний read-only `git diff --check`/scope audit і дописати підсумок без запуску заборонених Docker gates.
ФАКТ: Spec re-review завершений PASS; Standards re-review ще працює, тому фінальний review verdict не припускаю.
ДАЛІ: Поки Standards-review триває, виконати `git diff --check` і scope/stat audit без зміни файлів.
ФАКТ: Фінальний `git diff --check` PASS. Scope audit: 13 modified + 2 new frontend source/test files і append-only report; жодних backend/edge/docs-playback/tasks/generated/dependency змін, жодного fetch у learn routes/presentation. Tracked diff: +412/−50; untracked CourseLearningView files не входять у цей stat.
ДАЛІ: Дочекатися Standards-review; якщо PASS — зафіксувати остаточний стан, якщо finding — виправити й повторити відповідний gate.
ФАКТ: Standards re-review: blocking findings немає; strict TS, architecture/no-fetch, request-scoped Query і scope PASS. Два non-blocking test-design findings: окремий P03 залежить від P02 serial state й не працює focused; `uploadReady` приймає незалежні lesson id та positional index, чия розбіжність дасть неправильний upload + довгий timeout.
РІШЕННЯ: Обидва виправити. Об’єднати P02→P03 в один named scenario: після real completion повторний вхід зробити саме через course CTA, одночасно довівши outline і server-selected non-first URL/one UI read. Upload target зв’язати в один descriptor і локатор звузити через module group, щоб id та UI coordinate не розходилися непомітно.
ДАЛІ: Переглянути доступні назви module/lesson fieldsets в admin editor і внести мінімальний e2e-only patch.
ФАКТ: Admin editor має вкладені fieldset/legend з унікальною парою `Модуль N` → `Урок N`; Playwright може звузити lesson locator усередині module group. Це прибирає глобальний positional index.
РІШЕННЯ: `UploadTarget` міститиме id + module/lesson numbers одним аргументом, а locator будуватиметься вкладено. P02/P03 стануть одним тестом із обома ids у назві, тому focused grep будь-якого рядка виконає повний prerequisite/scenario.
ДАЛІ: Застосувати patch і оновити два рядки proof matrix, що посилаються на спільний сценарій.
ФАКТ: E2e patch застосовано: P02 і P03 тепер один focused-safe named scenario; повторний вхід після completion відбувається через course CTA, а один UI learning GET одночасно дає non-first URL і outline. Upload helper приймає один `UploadTarget` і звужує `Модуль N` → `Урок N` перед очікуванням id-specific mint.
ФАКТ: Попередня технічна спроба оновити вже записані P02/P03 рядки таблиці суперечила append-only правилу; їх повернуто дослівно, а актуальну поправку дописано нижче як новий запис. Наявний журнал не переписується далі.

### Поправка до proof matrix після Standards-review

| Рядок | Актуалізований доказ | Поправка до п’яти відповідей |
|---|---|---|
| P02 | Спільний named тест `P02 P03` | Q1/Q4: real heartbeat/exact position/completed-current assertions закривають P02 всередині одного сценарію; Q5: поломка persistence/completion/order червонить увесь спільний тест. Решта відповідей незмінна. |
| P03 | Той самий спільний named тест `P02 P03` | Q2: prerequisite P02 створюється всередині цього самого тесту, тому focused grep P03 не залежить від іншого test case; Q3: non-first lesson 2 і count=1 відсікають hardcoded-first/second-read false green; Q4: exact URL/read assertions належать P03; Q5: ігнор `continueLessonId`, втрата slug/id або другий UI read червонять спільний тест. |

ДАЛІ: Форматувати e2e, повторити typecheck і повний unit suite; потім попросити короткий повторний Standards verdict на два виправлення.
ФАКТ: `npx oxfmt e2e/playback.spec.ts` PASS — 1 file, exit 0.
ДАЛІ: Повторити `npm run typecheck` після final e2e helper/scenario refactor.
ФАКТ: Post-review `npm run typecheck` PASS, exit 0.
ДАЛІ: Повторити повний `npm test`, хоча production unit files не змінювались, щоб фінальний журнал мав зелений state після останнього patch.
ФАКТ: Post-review повний `npm test` PASS — 44 files, 369/369, 0 failed/skipped/todo, exit 0.
ДАЛІ: Попросити Standards-review підтвердити лише два виправлені non-blocking findings; після verdict повторити format-check/lint і фінальний diff check.
ФАКТ: Standards-review agent отримав точковий follow-up на об’єднаний scenario й nested upload locator; verdict ще не отримано.
РІШЕННЯ: Оскільки після попередньої повної послідовності був e2e patch, остаточно повторити всі чотири host gates з початку у правильному порядку, а не покладатися лише на post-patch typecheck/test.
ДАЛІ: Запустити фінальний `format:check`.
ФАКТ: Остаточний `npm run format:check` PASS — 162 files, exit 0.
ФАКТ: Точковий Standards re-review повернув PASS; обидва попередні non-blocking findings закриті, залишкових findings немає.
ДАЛІ: Запустити фінальний `lint` другим gate.
ФАКТ: Остаточний `npm run lint` PASS — zero warnings, architecture 90 files, exit 0.
ДАЛІ: Запустити фінальний `typecheck` третім gate.
ФАКТ: Остаточний `npm run typecheck` PASS, exit 0.
ДАЛІ: Запустити фінальний повний `npm test` четвертим gate.
ФАКТ: Остаточний `npm test` PASS — 44 files, 369/369 tests, 0 failed/skipped/todo, exit 0. Дозволена послідовність format-check → lint → typecheck → test повністю зелена після останньої зміни.
ДАЛІ: Виконати останній diff/scope/report integrity audit і дописати фінальний підсумок; жодних Docker або git mutation команд не запускати.
ФАКТ: Останній audit PASS: `git diff --check` exit 0; перший рядок report досі починається з `ДАЛІ:`; status має лише T8 frontend files і цей report; forbidden backend/edge/docs-playback/tasks/generated/dependency paths відсутні.
ФАКТ: Standards final review PASS. Spec agent отримав останню точкову перевірку після об’єднання P02/P03; verdict ще очікується.
ДАЛІ: Отримати останній Spec verdict і, якщо він PASS, дописати завершальний стан та reviewer-only Docker next step.
ФАКТ: Останній Spec re-review PASS; Standards і Spec axes обидві завершені без blocking або non-blocking findings.
ФАКТ: T8 реалізовано в дозволеному frontend scope: curriculum surface з ordered modules/lessons, completed/current states і responsive rail; course CTA та server-selected continue route з одним UI learning read; `/learn` → catalog; recursive locale completeness; exact generated-route mapping; renderer boundary 120; production-chain P02/P03 і isolated R37 browser proofs.
ФАКТ: Остаточні host gates PASS у правильному порядку: format-check 162 files; lint zero warnings + architecture 90 files; typecheck; unit 44 files, 369/369, 0 skipped/todo. `git diff --check` PASS.
ФАКТ: Через відсутній Docker не виміряні integration/check-generated/build/up/smoke/e2e, два 300-секундні transcodes, browser assertions і SQL duration distribution; точні команди наведені вище, жоден із цих пунктів не названо PASS.
РІШЕННЯ: Таск готовий до reviewer Docker-прогону. Git mutation/commit не виконувались за прямою забороною брифу. Наступний крок рецензента — `make check`, потім наведений SQL-зріз; за результатом оновити phase report поза цим T8 scope.

---

## Рецензія (сесія-рецензент) і виправлення

**Вердикт умовний:** без браузерного прогону P02/P03/R36/R37 лишаються **реалізованими, а не
доведеними**.

### БЛОКУЄ — P02 не виключав реалізації, що позначає переглянутими ВСІ уроки

Асерти на зміст були: count 4; `nth(0)` містить «Переглянуто»; `nth(0)` без `aria-current`;
`nth(1)` має `aria-current`; `nth(1)` містить «Поточний урок». **Жоден не стверджував, що уроки
2–4 позначки не мають.** Реалізація, яка малює «Переглянуто» на кожному рядку, проходила всі
пʼять — а саме таку картину дала б помилка «взяли `completed` курсу замість `completed` уроку».

**Відповідь на пʼяте питання в матриці вище хибна.** Там написано «якщо learning UI переплутає
completed/current/order, P02 почервоніє». При змішуванні **саме в цей бік** — не почервоніє.

**Виправлено:** додано `not.toContainText('Переглянуто')` для `nth(1)`, `nth(2)`, `nth(3)`.
Порядок serial це дозволяє: P02 стоїть **перед** R37, тож на момент асерту урок 3 ще не
завершений.

### Червоне не з тієї причини — точна рівність позиції проти каденції 2 с

`expect(learningLessonPosition(...)).toBe(firstPosition)` читало модель **одразу** після першого
heartbeat, а наступний іде через 2 с. Якби він встиг до читання — позиція зросла б і тест упав
би **на правильній поведінці**. Load на цій машині сьогодні тримався 20–47, тож вікно не
безпечне.

**Виправлено без послаблення строгості:** відео ставиться на паузу перед читанням. Каденція
гаситься **в джерелі** — ефект heartbeat закритий гейтом `playing` (`LessonPlayer.tsx:251`), а
його cleanup знімає таймер. Точна рівність лишилась.

**Наслідок, який довелось урахувати:** `seekTo` heartbeat **не шле** (`LessonPlayer.tsx:305-310`
лише рухає елемент і локальний стан), тож після паузи завершальний heartbeat не прийшов би
сам. Поновлення відтворення несе нову позицію, бо натискання play робить `heartbeatNow()` **до**
`video.play()` (`:349`). Порядок тепер: пауза → читання → seek 290 → play → завершальний
heartbeat.

### Знято заміром, не міркуванням

| Що перевірялось | Результат |
|---|---|
| B26 — чи нормалізація суфіксів не зʼїла зайвого | колізій «зрізаний суфікс збігся з простим ключем» — **нуль**; суфікси лише на двох справжніх множинах |
| R37 `toBe(300)` — чи не крихке проти реальної тривалості | ffprobe **самого воркера** (5.1.9, не хостового 7.x): `"duration": "300.000000"` |
| R24 — чи дерево справжнє, а не зібране `addChildren` | `import { routeTree } from '../routeTree.gen'` — справжнє |
| R36 — чи зміна `/learn` не прорвала захист | `requireIdentity` викликається **до** редиректа; анонім іде на логін |
| `clientSeq` у R37 | структурно безпечний: R37 — **єдиний** писач на урок 3 і пише рівно раз |
| Ізоляція P02/R37/P03 | тримається; якщо порядок поламають, P03 **почервоніє**, а не позеленіє |

### Лишається неперевіреним

`expect(learningReads).toBe(1)` — припущення про те, **де** відбувається запит. Якщо навігацію
обслужить SSR, лічильник дасть 0. Це не зелене з неправильної причини, а неперевірене число:
**найімовірніший кандидат на падіння в першому прогоні, і його падіння не є дефектом продукту.**

---

## Гейт: результат і його точна межа

### Зелений прогін

**`make check` — EXIT=0**, повний ланцюг `format-check → lint → typecheck → test → integration →
check-generated → build → up → seed-identities → seed → smoke → e2e`, **e2e 47/47**, нуль
пропущених. Дерево — `8cec836`.

Прикметно: прогін пройшов із **брудного** стеку, без `make down`. Це одночасно довело дві правки:
блок T8 працює, і `smoke` більше не вимагає щойно піднятого воркера (B-10).

### Що змінилось після зеленого прогону і як перевірено

Після нього лягли три коміти: `805560f` (виніс двох heredoc у константи модульного рівня),
`93fc3ea` (ціль `e2e-local` + B-12) і документація.

**Виніс перевірено точково, а не повним гейтом:** обидві shell-команди **побайтово ідентичні**
тим, що були в зеленому `8cec836`, термінатори в колонці 0 (рядки 255, 263).

```
зелений коміт 8cec836: 2 команди
HEAD:                  2 команди
побайтово ідентичні:   True
```

**Межа цього твердження названа явно:** доведено, що **виконуваний текст не змінився**. Не
доведено браузерним прогоном — див. нижче.

### Чому повного повторного гейту немає

Хост насичений чужими навантаженнями: `load averages 53,49 39,55 33,45` на **10** ядрах
(elasticsearch, cassiopeia-*, пʼять контейнерів cargotrans-*). Під ним `rstest` дає таймаути
5000 мс у `CourseEditorView.test.tsx` — та сама B-4. Той самий `npm test` на тихішій машині:
**369/369**.

**Це не видається за пройдене.** Повторний `make check` на дереві `93fc3ea` **не отримано**, і
його треба зробити, коли хост звільниться.

## Не зроблено — не плутати зі зробленим

**O1** (з T7) лишається відкритим. Заміряно, неостаточно: перекарбування **щоразу** заставало
нуль запитів у польоті. Потрібен прогін, який навмисно тримає запит у польоті на момент
перекарбування. Нуль спостережень — це **недоведено**, а не спростовано.

**B-6** — увесь `frontend/e2e/` не звірений із контрактом; ловив би описки в тестах і **не** ловить
розходження контракту з сервером.

---

## ⚠️ B-7 — єдиний за весь день дефект ПРОДУКТУ

Виноситься окремо навмисно. З дванадцяти пунктів беклогу, заведених за цей захід, **одинадцять
про інструмент** — і лише цей бачить користувач. У спільному переліку його відсортують як «ще
одна тестова знахідка».

`LessonPlayer` не має **жодного** слухача паузи (`onPause`, `addEventListener` — нуль збігів).
`playing` рухається лише через кнопку, зміну уроку й `onEnded`. А heartbeat **поновлює
володіння** — і робить це **перед** збереженням прогресу (`sessions.go:53`, підтверджено
`TestHeartbeatRenewsBeforeSavingProgress`).

**Наслідок для користувача:** вкладка, зупинена **поза кнопкою** — медіахаб Chrome, PiP,
медіаклавіші — продовжує слати heartbeat зі старою позицією. `ApplyProgress` цю позицію відкине
як таку, що не зросла, **але володіння вже поновлено**. Тобто сесія на паузі **утримує
акаунт-вайд слот і витісняє інший пристрій користувача**, нічого не відтворюючи.

**Не заміряно в браузері**, які саме шляхи ставлять паузу тут: `<video>` не має `controls`, тож
рідного UI в елемента немає. Це висновок із коду.

**Лікується одним рядком** — `onPause={() => setPlaying(false)}`. **Місце — T10**, де живуть
володіння й витіснення.
