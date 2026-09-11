ДАЛІ: прочитати повністю work/brief-t10.md, обов’язковий контекст репозиторію та інструкції implement; визначити точний обсяг таску й перевірки.
ФАКТ: бриф T10 і CTO-брифінг прочитані повністю; T10 вимагає виправити B-7 у LessonPlayer, додати наскрізні докази P13/P14/B15/B16/B32/B33/R25/R26, не торкатися docs/playback, work/tasks.md, generated, git і Docker-залежних гейтів.
РІШЕННЯ: виконувати через implement; після реалізації застосувати code-review, як вимагає implement. Коміт не робитиму, бо буквальна заборона T10 «не чіпати git» має вищий пріоритет за загальну інструкцію скіла.
ДАЛІ: прочитати всі названі у брифі секції плану/спеки, звіти T1/T5/T7/T9, backlog, work/tasks.md і поточний e2e-файл; потім зіставити вимоги з наявним кодом.
ФАКТ: обов’язкові джерела займають 5035 рядків; локалізовано точні межі D-1–D-3, §8, A28–A34/G3 та backlog B-7/B-8/B-9/B-11/B-13/B-16. План уже фіксує R26 як змішаний доказ, а B15 — як пару 200→403 на тому самому URL.
ДАЛІ: прочитати локалізовані секції плану й спеки, повні звіти T1/T5/T7/T9, потрібні backlog-секції, tasks і весь playback e2e частинами без обрізання.
ФАКТ: спільне читання перевищило ліміт виводу й було обрізане; D-1–D-3 та початок §8 прочитані, але цей результат не зараховую як повне читання решти джерел.
РІШЕННЯ: повторити читання малими іменованими діапазонами, доки кожен обов’язковий файл не буде отримано без обрізання.
ДАЛІ: дочитати §8 і spec, потім окремо звіти/backlog/tasks, після них e2e та LessonPlayer.
ФАКТ: прочитано п’ять питань §8, P13/P14, B15/B16/B32/B33 і повну специфікацію A28–A34/G3; комбінований вивід знову обрізав неназвані сусідні рядки §8, тому потрібні R25/R26 дочитаю окремо. Критерії вимагають account-wide ключ, takeover від наступного медіазапиту та зрозумілий клієнтський стоп.
ДАЛІ: прочитати точні R25/R26 та обов’язкові звіти чотирма окремими порціями.
ФАКТ: точні R25/R26 прочитані: R25 відсікає ключ лише по IP, R26 — композитний account+IP; R26 мусить мати другий реальний `$remote_addr` через `docker compose exec -T api`. Комбінований вивід звітів був обрізаний, тому повне читання звітів ще не завершено.
ДАЛІ: дочитати report-t1 і report-t5 контрольованими половинами без обрізання.
ФАКТ: report-t1 прочитано майже повністю, але 10k-token cap обрізав середину; підтверджено O2 закрито реальним заміром, edge авторизує кожен запит, а середовищні Docker-флейки там названі окремо від продукту.
РІШЕННЯ: надалі читати не більш як 120 рядків за один вивід; пропущену середину T1 та весь T5/T7/T9 дочитати такими порціями.
ДАЛІ: дочитати пропущений діапазон report-t1, потім перейти до T5.
ФАКТ: дочитано фінальну половину report-t1: O2 закрито числом `condition3_postgres_queries=0`; важливий урок доказів — активний server-minted playbackId і перехід 200→403, інакше 403 може бути хибнопозитивним.
ДАЛІ: прочитати report-t5 у трьох малих діапазонах, особливо контракти 409/410 і reviewer verification.
ФАКТ: перші 110 рядків report-t5 прочитані: `play` завжди забирає слот; heartbeat власника поновлює його до запису прогресу; 409 означає takeover, 410 — expired; smoke уже має real-mint фікстури й policy-переходи.
ДАЛІ: дочитати середину report-t5 з доступними гейтами та матрицею доказів.
ФАКТ: середину report-t5 прочитано. Найрелевантніше для T10: takeover/expired навмисно доступні лише heartbeat; integration доводить старий ID забороненим і новий дозволеним; smoke-докази мусять використовувати активний server-minted ID; Docker runtime раніше відкладався окремо.
ДАЛІ: дочитати фінал report-t5 і reviewer runtime results.
ФАКТ: report-t5 прочитано повністю. Reviewer підтвердив повний `make check` і real-mint policy flows; 409/410 контракт усталений, а переходи доступу доводяться парою позитивного й негативного результатів, не голим статусом відмови.
ДАЛІ: прочитати report-t7 частинами, зосередившись на recovery table, `openAndPlay`, `stopLoad` та e2e мінному полі.
ФАКТ: report-t7 1–220 прочитано. Плеєр мінтить лише після explicit play; 409 зупиняє, 410 робить silent remint, 403 terminal+refresh; `stopLoad()` є несучим запобіжником; пауза/TTL і heartbeat test profile уже мали окремі крихкі сценарії.
ДАЛІ: прочитати report-t7 221–450 — матрицю, O1 та перші раунди review.
ФАКТ: report-t7 221–450 прочитано. `openAndPlay` покладається на видиму мітку «Пауза» як доказ фактичного старту; pause/expiry B39, 409/410 recovery й `stopLoad` були спеціально рев’ювані. Історично B39 паузив кнопкою, бо програмна пауза ще не гасила cadence — саме це змінює T10.
ДАЛІ: прочитати report-t7 451–700 — фінальну реалізацію маршруту/e2e та результати локальних гейтів.
ФАКТ: report-t7 451–700 прочитано. E2E використовує окрему serialized playback fixture; seek helper мусить проходити через native setter; B40 рахує саме `.m4s`; unit/openAndPlay очікує мітку «Пауза». Плеєр має recovery/lifecycle складність, яку T10 не повинен розширювати.
ДАЛІ: дочитати report-t7 701–913, включно з reviewer browser findings і остаточним O1 статусом.
ФАКТ: report-t7 прочитано повністю. Reviewer підтвердив 364/364 unit і e2e після виправлень, але зафіксував, що програмна пауза в старій моделі не була джерелом `playing`; O1 тоді лишався відкритим. Цей звіт також підтверджує окремий playback project і shared-state/fixture ризики.
ДАЛІ: прочитати report-t9 трьома блоками; встановити актуальну каденцію, B18 і фінальний стан O1.
ФАКТ: report-t9 1–125 прочитано. Актуальна реалізація серіалізує heartbeat; 409 ставить terminal flag і відкидає наступні записи; pause/seek/hide/nav/ended — окремі причини; healthy cadence прив’язана до настінних дедлайнів. T9 навмисно не додавав native `onPause`, залишаючи B-7 для T10.
ДАЛІ: прочитати report-t9 126–250 — review fixes, O1 і доступні гейти.
ФАКТ: report-t9 126–250 прочитано. B18 browser proof був authored, але не прогнаний виконавцем; O1 теж був unverified на цьому етапі. Review виправив cadence drift і зайвий resume tick; B10 очікує точний набір seek/pause/hide/ended/nav.
ДАЛІ: дочитати report-t9 251–358 — authoritative reviewer прогін, B18 і закриття O1.
ФАКТ: report-t9 прочитано повністю. На розвантаженій машині `make check` був green, e2e 52/52; B18 переписаний і потім прогнаний зеленим у фінальному ланцюгу; O1 закрито числами 1 inflight до `stopLoad`, 0 на mint, restore 150. B-7 лишився відкритим саме для T10.
ДАЛІ: прочитати потрібні backlog B-7/B-8/B-9/B-11/B-13/B-16 та work/tasks.md без обрізання.
ФАКТ: backlog B-7/B-8/B-9/B-11 прочитано. B-7 прямо пояснює slot leak від зовнішньої паузи; `beforeEach(resetLessonProgress)` на рівні student describe вже є; hooks залежать від describe, а SQL heredoc-константи мусять лишитися модульними.
ДАЛІ: дочитати B-13/B-16 і весь tasks; потім перейти до актуального коду та тестів.
ФАКТ: B-13/B-16 і work/tasks.md прочитані. `workers: 1` є детермінізмом shared catalog state й не змінюється; таймаут можна підняти лише після триразового тихого заміру, що показує власний недобюджетований тест. T10 лишається planned, попередні залежності done.
ДАЛІ: прочитати актуальний LessonPlayer, його unit-тести, playback.spec.ts і Playwright config; визначити мінімальний RED для B-7 та структуру e2e доказів T10.
ФАКТ: актуальний inventory підтвердив рівно три ручні `setPlaying` для видалення (`onReady play().then`, кнопкова pause, кнопкова play().then), окремі resets лишаються; `onPause`/`onPlaying` відсутні. playback.spec має 1001 рядок, existing student `beforeEach` і module-level SQL; Playwright `workers: 1`.
ДАЛІ: прочитати LessonPlayer.tsx і unit-тести повністю; спочатку сформулювати RED, що зовнішня media pause зупиняє cadence та external playing запускає її без раннього сигналу.
ФАКТ: LessonPlayer.tsx прочитано повністю. `onReady` і кнопка виставляють playing у `play().then`; кнопкова pause робить це вручну; `onEnded`, lesson-change і terminal reset мають окремі причини лишитися. Visibility handler уже паузить відео й окремо шле hide, тому native `pause` не повинен дублювати heartbeat reason.
РІШЕННЯ: B-7 unit proof перевірятиме DOM events як джерело state: fulfilled `play()` без `playing` не запускає cadence; `playing` запускає; зовнішній `pause` її гасить. Сам `onPause` лише змінює gate і не шле додатковий pause heartbeat.
ДАЛІ: прочитати весь LessonPlayer.test.tsx і вбудувати RED у наявний fake-media harness.
ФАКТ: LessonPlayer unit suite прочитано повністю. Наявний B09/B10 harness має fake timers і media callback; B10 currently invokes button pause and expects one `pause` heartbeat. Новий B-7 proof can reuse it but must dispatch native `playing`/`pause` events and assert cadence count, without confusing state synchronization with exit-progress writes.
ДАЛІ: прочитати playback.spec.ts повністю чотирма неперекривними блоками; інвентаризувати helper/session/cookie/fixture seams для восьми T10 рядків і B-7.
ФАКТ: playback.spec.ts 1–260 прочитано. Наявні helpers валідовують зовнішні payloads вручну, `openAndPlay` чекає 201, видиму «Пауза» і зростання часу; course fixture має два відео в різних уроках одного курсу, але для P14 потрібен окремий другий курс.
ДАЛІ: прочитати playback.spec.ts 261–520 — SQL-константи, fixture lifecycle й початок student scenarios.
ФАКТ: playback.spec.ts 261–520 прочитано. Course fixture створюється production admin API/upload/transcode path, cleanup і progress reset виконуються module-level heredoc constants. Student tests уже serial та ізольовані `beforeEach`; T10 може додати helpers/другий course без нового reset.
ДАЛІ: прочитати playback.spec.ts 521–780 — cross-context, B39/O1, generation replacement і поточні error-flow assertions.
ФАКТ: playback.spec.ts 521–780 прочитано. Cross-context helper уже існує; B39 паузить кнопкою перед TTL; O1 трекає real media requests і session mint. T10 не має часто pause/resume, бо після B-7 кожна пара скидає повний cadence interval і TTL може протухнути.
ДАЛІ: дочитати playback.spec.ts 781–1001 і Playwright config; знайти безпечне місце для T10 serial scenarios до irreversible archive.
ФАКТ: playback.spec.ts і Playwright config прочитані повністю. Безпечне місце T10 — у student serial block до B40/R04; P13+B15+B32 можна довести одним причинним сценарієм, решту — незалежними tests із чинним beforeEach. R26 реалізується literal container curl без client XFF.
РІШЕННЯ: B-7 browser proof здійсненний без медіахаба: programmatic `video.pause()` породжує справжню DOM pause event; відсутність heartbeat понад каденцію й відновлення після `video.play()` доводять потрібне правило. Discussion не відкриваю.
ДАЛІ: прочитати повний TDD skill, додати RED unit regression для `playing`/`pause` DOM events і лише після точного RED змінити LessonPlayer.
ФАКТ: TDD skill, tests.md і mocking.md прочитані повністю; CONTEXT.md відсутній. Погоджений брифом seam — публічна поведінка rendered `<video>` і cadence через injected зовнішні actions/media factory, без тестування приватних функцій.
РІШЕННЯ: один vertical slice: resolved `play()` без DOM `playing` не запускає timer; DOM `playing` запускає; DOM `pause` гасить; наступний DOM `playing` планує свіжий інтервал.
ДАЛІ: додати цей один failing test у LessonPlayer.test.tsx, не змінюючи production code.
ФАКТ: RED-test `B-7 follows native playing and pause events for ownership cadence` додано через public DOM events. Він окремо відсікає два дефекти: ранній `play().then(setPlaying)` і відсутність реакції на зовнішню pause.
ДАЛІ: запустити лише LessonPlayer test file; RED зарахувати тільки якщо падає на очікуваній ранній cadence assertion.
ФАКТ: RED підтверджено точно: 12/13 tests pass, новий B-7 падає на першій умові — після resolved `play()` без DOM `playing` уже пішов один heartbeat замість нуля. Це продуктова причина, не environment timeout.
ДАЛІ: зробити мінімальний GREEN у LessonPlayer: додати `onPause`/`onPlaying`, прибрати три ручні `setPlaying`, зберегти `onEnded` і всі non-media resets.
ФАКТ: production B-7 fix внесено буквально: `<video onPause/onPlaying>` став єдиним джерелом live media state; три ручні setters видалені; `onEnded`, lesson reset і terminal stop лишені. `hls.stopLoad()` не змінено.
ДАЛІ: повторити повний LessonPlayer file; очікувані падіння тестових fake media класифікувати окремо й узгодити з реальними DOM events.
ФАКТ: GREEN attempt дав 8/13; новий B-7 пройшов, а п’ять старих tests впали, бо їхні `play()`/`pause()` stubs не породжували браузерні `playing`/`pause` events і спирались на видалені ручні setters. Це очікувана зміна тестового контракту, не новий product failure.
РІШЕННЯ: навчити наявні public-interface tests відтворювати DOM events, які реальний HTMLMediaElement дає сам; production behavior не розширювати.
ДАЛІ: додати `fireEvent.playing` після фактичного старту в cadence tests і `pause` dispatch у button fakes; повторити suite.
ФАКТ: комбінований test patch не застосувався через неточний context у B24; apply_patch зупинився атомарно, файл не змінено.
ДАЛІ: перечитати точні короткі діапазони п’яти tests і внести context-stable edits малими hunks.
ФАКТ: короткі діапазони підтвердили п’ять точок адаптації; змішані без номерів фрагменти важко прив’язати безпомилково.
ДАЛІ: отримати нумеровані діапазони навколо кожного failure й застосувати малі незалежні hunks.
ФАКТ: старі unit tests адаптовано до браузерного контракту: explicit `playing` після fake start, fake `pause()` dispatches DOM pause. Зміни торкаються лише public events, не приватного state.
ДАЛІ: перевірити розташування всіх нових event dispatches і повторити LessonPlayer suite.
ФАКТ: composite verification не стартував: у cwd `frontend` передано зайвий префікс `frontend/`, тому `rg` не знайшов файл, а `&&` не допустив запуск tests. Це не test result.
ДАЛІ: повторити inventory й suite з правильним шляхом від frontend cwd.
ФАКТ: inventory показав event dispatches, suite 11/13. B09 ще не отримав `playing` через ambiguous hunk і закономірно має 0 ticks; B24 — відомий `test timed out in 5000ms` після 11 assertions, потребує тихого триразового заміру за B-16, не негайної правки.
ДАЛІ: виправити лише B09 на точному контексті, перевірити й прибрати випадково доданий `playing` у несуміжному mint-retry test, потім запустити focused B-7/B09.
ФАКТ: B09 тепер явно dispatches `playing`; випадкову зміну mint-retry test прибрано. TDD slice знову локалізований до cadence semantics.
ДАЛІ: запустити лише B-7 і B09; після GREEN окремо тричі заміряти B24 на тихому focused прогоні.
ФАКТ: перший focused invocation не повернув жодного stdout у 30 с, а wrapper не надрукував exit/session metadata; результат непридатний і не зарахований.
ДАЛІ: перевірити наявний rstest process; якщо його немає — повторити з явним друком exit/session, якщо є — дочекатися саме його.
ФАКТ: `pgrep` повернув PID 18718, але без командного рядка; ще не можна відрізнити живий rstest від самого пошуку.
ДАЛІ: прочитати command/status PID 18718; не запускати дубль, якщо це справжній test runner.
ФАКТ: PID 18718 вже не існує; активного focused runner немає. Попередній прогін завершився без збереженого verdict.
ДАЛІ: повторити focused test із явним exit/session metadata; якщо поверне session id — poll через write_stdin.
ФАКТ: focused B-7+B09 GREEN: 2/2 selected, 11 skipped, exit 0. Новий ownership-cadence proof і existing deadline cadence сумісні.
ДАЛІ: виконати B24 ізольовано тричі без зміни 5000ms timeout, щоб класифікувати попередній timeout за B-16.
ФАКТ: B24 пройшов ізольовано 3/3 без зміни timeout (кожен selected test PASS); попередній timeout у повному file класифіковано як host contention, не власний недобюджетований test. Таймаут не піднято.
ДАЛІ: розширити e2e fixture другим published course з власним ready asset, не змінюючи workers/reset/heredoc placement; додати helpers для cross-course route і container-origin play.
ФАКТ: e2e namespace/state розширено для другого курсу; upload helper приймає explicit courseID, а openAndPlay — explicit slug. Existing callers і defaults не змінені.
ДАЛІ: додати typed helper для `docker compose exec -T api curl` POST, який повертає server-minted PlaybackSession і не підробляє X-Forwarded-For.
ФАКТ: R26 helper додано через literal `docker compose exec -T api curl http://nginx:8080/...`; він передає лише session cookie/origin/body, не встановлює XFF, валідовує 201 і session shape.
ДАЛІ: у beforeAll створити/author/upload/publish другий run-scoped course production API шляхом і додати його cleanup тим самим module-level SQL statement.
ФАКТ: second course fixture створюється admin API → course tree → real upload/transcode → publish; використовує dedicated 100s asset і видаляється тим самим module-level SQL через afterAll loop. Existing course/lessons не перепризначені.
ДАЛІ: відформатувати поточні frontend зміни й виконати typecheck до додавання сценаріїв, щоб локалізувати fixture/helper errors.
ФАКТ: formatter PASS. Typecheck завершився єдиною очікуваною помилкою TS6133: новий `playFromAPIContainer` ще не використано; інших helper/fixture errors не виявлено. Exit 2 не названо green.
ДАЛІ: додати шість independent browser tests: B-7; P13+B15+B32; P14; B33; R25; R26; B16 окремо — усі до B40/R04 і під чинним beforeEach.
ФАКТ: додано 7 browser scenarios: B-7; combined P13/B15/B32; P14; B33; R25; R26; B16. B15 робить explicit 200→403 на тому самому `.m4s`; R26 використовує container-origin play; B16 робить другий mint одразу після закриття abandoned context і читає media 200.
ДАЛІ: відформатувати три touched frontend files; потім окремо lint/typecheck/discovery, не називаючи це browser runtime.
ФАКТ: oxfmt PASS on LessonPlayer, its unit test and playback.spec.ts.
ДАЛІ: запустити frontend lint+architecture; виправляти лише errors у T10 slice.
ФАКТ: `npm run lint` PASS; oxlint zero warnings, architecture boundaries checked 90 files.
ДАЛІ: запустити frontend typecheck з bounded polling; перевірити нові context/container types.
ФАКТ: `npm run typecheck` PASS, exit 0; helper/session/cookie/fixture/test types коректні.
ДАЛІ: запустити Playwright discovery для project=playback; підтвердити всі T10 row labels без браузера/Docker.
ФАКТ: Playwright discovery PASS, 55 tests/13 files; усі T10 row labels знайдені в playback project. Це static discovery, не browser execution.
ДАЛІ: виконати обов’язковий code-review workflow на explicit T10 file set без git; спершу прочитати skill повністю, потім виправити лише підтверджені findings.
ФАКТ: code-review skill прочитано повністю; він прямо вимагає дві незалежні паралельні осі. Fixed-point/git процес неможливий через literal NOT «не чіпати git», тому scope pinned explicit files: LessonPlayer.tsx, LessonPlayer.test.tsx, playback.spec.ts; spec source — work/brief-t10.md.
ДАЛІ: запустити Standards і Spec sub-agents read-only без git/edits/report writes; паралельно локально перевірити причинність кожного нового e2e assertion.
ФАКТ: локальний causality audit підтвердив: P13/B15/B32 має exact `.m4s` 200→403; P14 використовує інший lesson/course slug; R25 чекає heartbeat boundary і рух обох відео; R26 падає саме для composite account+IP key; B16 не чекає TTL і доводить immediate takeover abandoned ownership.
ФАКТ: audit знайшов security risk у helper: session cookie передавався як process argv, тому exec failure міг надрукувати секрет. Також comment `setPlaying(true)` у openAndPlay застарів після B-7.
РІШЕННЯ: передавати Cookie header у curl через stdin (`-H @-`) за допомогою вузького spawn helper, щоб secret не був у argv/errors; оновити comment на DOM `playing` event.
ДАЛІ: внести ці дві review-hardening правки, потім повторити format/lint/typecheck/discovery.
ФАКТ: R26 curl тепер читає Cookie header зі stdin (`-H @-`); секрет відсутній у process argv та custom error. openAndPlay comment тепер правильно називає DOM `playing` як доказ старту.
ДАЛІ: відновити попередній 30s process timeout у spawn options, потім форматувати й прогнати static gates.
ФАКТ: container curl process знову має bounded 30s timeout.
ФАКТ: Spec reviewer знайшов blocking fixture defect: uploadReady переходив на explicit cross-course id, але polling videos лишався на primary `course.id`, тому second-course setup вічно бачив `absent`.
ДАЛІ: зробити selected course id єдиним локальним значенням uploadReady для navigation і polling; потім static gates.
ФАКТ: uploadReady тепер обчислює `targetCourseID` один раз і використовує його і для UI route, і для videos polling; cross-course transcode can complete observation on its own course.
ДАЛІ: format → lint/architecture → typecheck у цьому порядку для current T10 files.
ФАКТ: current static sequence PASS: oxfmt, lint/architecture (90 files), typecheck.
ФАКТ: repeated discovery lists every T10 row and total 55 tests. Exact inspection підтвердив Spec-review finding: expired-session unit test still lacked `playing`; earlier ambiguous patch had not landed there.
ДАЛІ: виправити саме expired-session test на точних lines, потім прогнати весь LessonPlayer suite, не лише B-7/B09.
ФАКТ: expired-session test тепер має captured video і dispatches DOM `playing` після explicit start, щоб 410/remint cadence перевірявся за новим media-state contract.
ДАЛІ: перевірити точне landing місце й запустити весь LessonPlayer test file.
ФАКТ: exact inspection підтвердив `fireEvent.playing` у expired-session test; повний LessonPlayer suite GREEN 13/13, 0 skipped, exit 0.
ФАКТ: Standards first pass повторив уже виправлений cross-course bug і назвав два pre-existing LessonPlayer policy smells; вони не внесені T10 й не рефакторяться поза scope. Обидва review agents отримали current fixes на фінальний re-review.
ДАЛІ: почати буквальний локальний ordered gate: `npm run format:check`.
ФАКТ: `cd frontend && npm run format:check` — PASS, exit 0; oxfmt перевірив 162 файли, форматування відповідає правилам.
ФАКТ: перша спроба append не застосувалась через неточний контекст останнього рядка; tail підтвердив, що файл не було переписано чи пошкоджено, після чого запис додано за точним контекстом.
ДАЛІ: запустити наступний локальний gate `npm run lint`.
ФАКТ: `cd frontend && npm run lint` — PASS, exit 0; oxlint не має warnings, architecture boundaries перевірені на 90 файлах.
ДАЛІ: запустити наступний локальний gate `npm run typecheck`.
ФАКТ: `cd frontend && npm run typecheck` — PASS, exit 0; `tsc --noEmit` завершився без діагностик.
ДАЛІ: запустити останній доступний frontend gate `npm test`.
ФАКТ: `cd frontend && npm test` — PASS, exit 0: 44 files, 373/373 tests passed, 0 failed, 0 skipped, 0 todo; snapshots unchanged.
ДАЛІ: визначити й запустити Go-тести без Docker, не торкаючись integration-tag або compose gates.
ФАКТ: Makefile розділяє `test` (`go test -race ./...`) та Docker-backed `integration` (`infrastructure/tests/integration.sh`); у backend є unit і `*_integration_test.go`, тому перед запуском перевіряю build tags інтеграційних файлів.
ДАЛІ: підтвердити, що `go test -race ./...` не активує Docker integration suites, а тоді виконати саме цей unit gate.
ФАКТ: усі 9 знайдених `*_integration_test.go` мають `//go:build integration`; Makefile unit gate запускає `go test -race ./...` без цього tag, отже Docker suites не ввійдуть.
ДАЛІ: виконати `go test -race ./...` у backend з task-local Go cache.
ФАКТ: `go test -race ./...` стартував і після 30 секунд ще виконується без виведених помилок (session 65637); фінального PASS/FAIL ще немає.
ДАЛІ: дочекатися завершення цієї ж test session і зафіксувати точний результат.
ФАКТ: та сама Go test session після ще 30 секунд лишається активною без output; це не зараховано як PASS і процес не перезапускався.
ДАЛІ: продовжити очікування поточної session, не довше 30 секунд за один poll.
ФАКТ: третє 30-секундне вікно завершилось без output, test session досі активна; статус залишається RUNNING, не PASS.
ДАЛІ: ще раз опитати ту саму session; якщо процес зависатиме, локалізувати пакет без зміни test contract.
ФАКТ: Go race-suite почав віддавати результати: чотири `cmd/*` пакети без test files завершились; загальна session ще активна, тому підсумкового статусу немає.
ДАЛІ: дочекатися завершення решти пакетів у цій самій session.
ФАКТ: `cd backend && GOCACHE=/tmp/courses-t10-go-cache go test -race ./...` — PASS, exit 0; усі unit-пакети завершились, integration-tag suites не запускались.
ДАЛІ: зробити фінальний статичний аудит мінного поля: рівно потрібні media-state setters, `workers: 1`, відсутність XFF, module-level SQL, точні T10 labels і відсутність змін у заборонених шляхах.
ФАКТ: статичний аудит підтвердив: `setPlaying` лишився тільки у дозволених lesson/terminal reset, native `onPause`/`onPlaying` та `onEnded`; заборонені `play().then(setPlaying)` і manual pause setter відсутні. `workers: 1` незмінний; SQL statements module-level; усі B-7/P13/P14/B15/B16/B32/B33/R25/R26 labels присутні; XFF у T10 e2e відсутній.
РІШЕННЯ: два `beforeEach` у файлі належать різним чинним describe scopes (visibility і student playback); T10 не додавав reset hook і використовує наявний `test.beforeEach(resetLessonProgress)` у student scope.
ДАЛІ: зафіксувати окремі фінальні verdicts Standards і Spec рев’ю, не змішуючи їх, та перевірити точний change scope без git-команд.
ФАКТ: Standards review verdict — PASS. Hard violations і Fowler judgement calls у T10 delta не знайдено; перевірені shared targetCourseID у uploadReady, cookie через curl stdin, native media events, module-level SQL, validated external payloads і незмінний workers contract. Успадковані LessonPlayer policy smells не є зміною T10.
ФАКТ: Spec review verdict — PASS після двох виправлень; blocking/partial вимог не лишилось. Окремо підтверджені B-7, причинні P13/B15/B32, cross-course P14, B33, R25, real-source R26, abandoned B16 і мінне поле. E2E є authored, але не measured без Docker.
ФАКТ: усі implementation writes цієї роботи були обмежені `frontend/src/presentation/learn/LessonPlayer.tsx`, `frontend/src/presentation/learn/LessonPlayer.test.tsx`, `frontend/e2e/playback.spec.ts` та append-only `work/report-t10.md`; git не викликався, `docs/playback/**`, `work/tasks.md` і generated files не редагувались.
ДАЛІ: зібрати обов’язкову таблицю `рядок §8 → доказ` із п’ятьма відповідями для кожного T10 row, позначивши authored-only Docker scenarios чесно.

## §8: рядок → причинний доказ

| Рядок | Що доводить тест | Production artifact | Як виключено false green | Чому це саме цей рядок | Яка поломка зробить тест red |
| --- | --- | --- | --- | --- | --- |
| B-7 | GREEN unit: resolved `play()` сам не запускає cadence; native `playing` запускає, native `pause` зупиняє, наступний `playing` відновлює fresh cadence. Authored E2E повторює це реальною програмною паузою поза кнопкою. | `frontend/src/presentation/learn/LessonPlayer.tsx`; `LessonPlayer.test.tsx`; `e2e/playback.spec.ts` test `B-7…` | Лічильник heartbeat перевіряється до першого `playing`, після більш ніж двох інтервалів pause і після нового повного інтервалу resume; тест не покладається на текст кнопки як єдиний oracle. | Рядок вимагає, щоб ownership renewal відповідав фактичному media state, зокрема pause не через control button. | Повернення `play().then(setPlaying)`, видалення `onPause`/`onPlaying`, renewal під час pause або reuse старого cadence зробить unit/E2E red. |
| P13 | Authored E2E: account B-device бере ownership; A-device отримує takeover, B лишається playing. | `e2e/playback.spec.ts` test `P13 B15 B32…` | Два окремі browser contexts логіняться тим самим student account; обидва реально відкривають і запускають lesson, а не підміняють storage/API state. | Це same-account, different-device displacement у межах одного lesson. | Якщо новий playback не витіснить попередній або другий player не продовжить грати, assertions стануть red. |
| P14 | Authored E2E: takeover діє account-wide, коли другий context грає lesson іншого run-scoped published course. | Другий course fixture та test `P14…` у `e2e/playback.spec.ts` | Другий course окремо створюється, upload/transcode/publish очікуються за його власним `targetCourseID`; media paths перевіряються як різні. | Рядок перевіряє глобальність ownership між курсами, не лише між lessons одного course. | Course-scoped ownership або polling не того course лишить перший player активним/fixture завислим і зробить тест red. |
| B15 | Authored E2E: той самий captured `.m4s` має 200 до takeover і 403 одразу після. | test `P13 B15 B32…` у `e2e/playback.spec.ts` | До й після використано дослівно той самий `segmentPath` і request context першого пристрою; контрольний 200 передує displacement. | Рядок вимагає denial exact same media segment, а не іншого URL чи playlist. | Якщо stale media access лишиться дозволеним після takeover, очікування 403 стане red. |
| B16 | Authored E2E: після mint і abandon без clean-close інший device того ж account негайно mint-ить новий playback та дістає media 200. | test `B16…` у `e2e/playback.spec.ts` | Перший context закривається без stop endpoint; другий mint не чекає TTL, IDs мусять різнитися, а media authorization перевіряється реально. | Це точний abandoned-session immediate-recovery case. | Lock до TTL, вимога clean-close, повторне використання ID або denied media зроблять тест red. |
| B32 | Authored E2E: після takeover перший media request уже негайно 403. | test `P13 B15 B32…` у `e2e/playback.spec.ts` | 403 запитується одразу після підтвердженого second-device play, без очікування heartbeat/TTL; перед takeover той самий request був 200. | Рядок фіксує immediate revocation boundary наступного запиту. | Delayed propagation або grace window поверне 200 й зробить тест red. |
| B33 | Authored E2E: дві tabs одного browser context displacement-ять одна одну; друга грає, перша бачить takeover. | test `B33…` у `e2e/playback.spec.ts` | Обидві pages мають спільний context/session, але різні player instances; друга проходить real play flow. | Це окрема внутрішньобраузерна вкладкова конкуренція, не multi-device contexts. | Tab-local ownership або відсутність takeover UI зробить тест red. |
| R25 | Authored E2E: два різні accounts за однією browser host/IP одночасно грають після heartbeat boundary. | test `R25…` у `e2e/playback.spec.ts` | Student та admin contexts мають різні authenticated accounts, однаковий browser source; перевіряються відсутність takeover, обидва Pause controls і просування обох currentTime. | Рядок доводить, що IP не є ownership key і не створює cross-account collision. | IP-keyed exclusivity або випадковий cross-account revoke зупинить один player/покаже takeover й зробить тест red. |
| R26 | Authored E2E: той самий account з реального іншого network source (`docker compose exec -T api` → nginx) витісняє browser playback. | helper `playFromAPIContainer` і test `R26…` у `e2e/playback.spec.ts` | Немає client-supplied XFF; запит реально виходить із API container, cookie передається curl через stdin, response/status валідовані, playback IDs різні. | Рядок вимагає same-account takeover незалежно від `$remote_addr`, із двома реальними sources. | IP-scoped ownership, недійсний container flow або відсутність displacement зробить test red. |

ФАКТ: таблиця дає всі п’ять відповідей для B-7, P13, P14, B15, B16, B32, B33, R25 і R26. B-7 має виміряний unit-доказ; усі browser rows позначені authored E2E, а не runtime PASS.
ДАЛІ: окремо зафіксувати залежні від programmatic pause місця, точні reviewer commands і список неперевірених Docker gates.
ФАКТ: аудит pause-залежностей: B18, B39/O1 і P02/P03 натискають видиму кнопку `Пауза`, тому native `pause` event тепер синхронізує UI, а explicit pause heartbeat лишається в button handler. Unit B10 fake `pause()` тепер емітить DOM `pause`; B-7 навмисно викликає `video.pause()` поза кнопкою й доводить зупинку cadence. Інших programmatic pause dependencies не знайдено.
РІШЕННЯ: `stopLoad` у B39/O1 не змінювався: фінальний T9 доказ вимірював 1 in-flight перед stopLoad і 0 на момент mint, тож видалення зруйнувало б причинність remint regression.
ДАЛІ: сформулювати точні команди для локального повтору та Docker-reviewer, а також перелічити кожен Docker gate як NOT RUN.
ФАКТ: Playwright config підтверджує, що project `playback` залежить від `playback-profile` → `media` → `chromium` → `setup` і має teardown; Makefile називає `make e2e-local` правильним standalone browser entrypoint та `make check` повним ordered gate.

## Команди для reviewer

Локально без Docker, у перевіреному порядку:

```sh
cd /Users/strelia/projects/courses/frontend
npm run format:check
npm run lint
npm run typecheck
npm test
cd /Users/strelia/projects/courses/backend
GOCACHE=/tmp/courses-t10-go-cache go test -race ./...
```

Повна канонічна перевірка на host із Docker:

```sh
cd /Users/strelia/projects/courses
make check
```

Точковий causal replay T10 після підняття й seed disposable environment:

```sh
cd /Users/strelia/projects/courses
make up
make seed-identities
make seed
cd frontend
npx playwright test --project=playback --grep 'B-7|P13|P14|B15|B16|B32|B33|R25|R26' --reporter=line
```

ФАКТ: через заявлену відсутність Docker НЕ ЗАПУСКАЛИСЯ: `make integration`, `make check-generated`, `make build`, `make up`, `make smoke`, `make e2e`, `make e2e-local`, повний `make check`; відповідно не виміряні actual transcode другого course, browser takeover/media denial і container-origin R26. Жоден із цих gates не позначено passed.
ДАЛІ: перевірити Playwright test discovery для фінального current file, після чого дописати підсумок реалізації, обмеження й наступний task у report.
ФАКТ: `npx playwright test --project=playback --list` — PASS, exit 0; Playwright discover-ить усі T10 labels B-7/P13/P14/B15/B16/B32/B33/R25/R26, загалом 55 tests у 13 files разом із dependency/teardown chain.
ДАЛІ: перечитати фінальний хвіст append-only журналу й перевірити, що він завершується чесним підсумком, обмеженнями та наступним кроком.
ФАКТ: фінальний audit append-only журналу пройдено: до цього запису файл мав 219 рядків, історія RED → GREEN, обидва review verdicts, §8 proof table, перевірені й неперевірені gates та reviewer commands присутні; попередні записи не переписувались.
РІШЕННЯ: T10 локально завершено. LessonPlayer тепер бере playing-state з native media events; unit regressions покривають pause/resume cadence; authored browser coverage причинно закриває B-7, P13, P14, B15, B16, B32, B33, R25 і R26, включно з другим course та справжнім container network source.
ФАКТ: доступні gates GREEN: frontend format-check, lint/architecture, typecheck, 373/373 unit tests; backend `go test -race ./...`; Playwright discovery 55/55 listed. Standards review PASS; Spec review PASS.
ФАКТ: Docker/runtime evidence лишається NOT RUN, тому browser/transcode/container сценарії не оголошуються passed і phase report/task tracker не змінювались.
ДАЛІ: reviewer на Docker-capable host запускає `make check`; за потреби causal replay виконується точковою командою з секції «Команди для reviewer». Лише після runtime GREEN ці E2E rows можна назвати measured.
ПИТАННЯ: немає; implementation scope і наступний verification step визначені брифом.


---

## Облік: де насправді лежить код T10

Реалізація T10 потрапила в коміт **`6110a09`**, чиє повідомлення описує лише правку документів.
Причина — `git add -A`, зроблений ведучою сесією, поки Codex іще писав. Історію не переписано
(коміт запушено); факт записано тут і як **B-17**.

## Дірка, якої Codex не міг знати

Бриф було доповнено **після** того, як Codex завершив. Тому в зданому коді:

- `onPause`/`onPlaying` додано (`LessonPlayer.tsx:411-412`) — **B-7 полагоджено**;
- `queueHeartbeat('pause')` лишився **в гілці кнопки** (`:429`).

Отже пауза з медіахаба спиняє каденцію, але **не пише прогрес**. Це рядок **R39**, заведений у
§8 разом із рішенням: прапорець «пауза наша» перед кожним власним `video.pause()`, бо наївне
перенесення виклику дало б подвійні записи з `visibilitychange` і зайвий запис із `stop()`.

---

## Гейт зелений — 2026-09-11

**`make check` EXIT=0**, повний ланцюг, **e2e 59/59**, нуль відмов.

| Рядок | Що доведено |
|---|---|
| P13, P14 | другий пристрій зупиняє перший; ліміт діє **наскрізь по курсах** |
| B15 | **200 на тому самому шляху до витіснення → 403 після** |
| B32 | витіснення миттєве, без вікна очікування |
| B33 | дві вкладки одного браузера витісняють одна одну |
| B16 | привидна сесія не блокує — слот звільняє TTL |
| R25 + R26 | ключ ліміту — **акаунт**: R25 виключає ключ по IP, R26 — композит «акаунт + IP» |
| **B-7** | пауза поза кнопкою **зупиняє поновлення володіння** |
| **R39** | та сама пауза **пише прогрес** |
| **O1** | зелений у повному ланцюгу: 2 у польоті → 0 після `stopLoad`; `0 → 150` |

**R26 зроблено точніше, ніж передбачав план.** Браузером недосяжний: nginx перезаписує
`X-Forwarded-For` своїм `$remote_addr`. Запит іде **з контейнера api через nginx** — тим самим
шляхом, що й справжній клієнт.

**Виправлено по дорозі:** B-7 лагодив володіння й ламав прогрес на тому ж шляху (запис лишався в
гілці кнопки); дві заглушки описували неможливий стан (`pause()` без зміни `paused`); B-5 друга
половина — задачі блоку playback валили `recovery` через `unhealthy` воркера.

**Відкрите:** B-6 — `frontend/e2e/` не звірений із контрактом.
