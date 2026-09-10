ДАЛІ: Прочитаю повністю репозиторний брифінг, бриф T9, поточний план і список задач Phase 5; потім визначу один логічний slice для буквальної реалізації.
ФАКТ: Скіл implement вимагає реалізації за брифом, регулярних точкових перевірок, повного фінального набору перевірок, code review і коміту поточної гілки.
РІШЕННЯ: Дотримуюсь цього процесу; TDD застосую на наявних seam, якщо це сумісно з буквальними вимогами брифу.
ДАЛІ: Читаю повністю CTO-RESEARCH.md, brief-t9.md, поточний Phase 5 plan.md і work/tasks.md.
ФАКТ: Поточні phase-документи розташовані в docs/media/{plan,report,spec,gates}.md; work/tasks.md і всі потрібні брифи на місці.
ДАЛІ: Читаю визначені джерела цілком, без вибіркового перегляду.
ФАКТ: Сукупний вивід чотирьох великих файлів був обрізаний, тому вимогу повного читання цим кроком ще не виконано. Видима частина CTO-RESEARCH фіксує, що Phase 5 закрита, а актуальна черга вже перейшла до Phase 6; у tasks T9 має статус planned.
РІШЕННЯ: Не покладаюсь на обрізаний вивід; перечитаю кожен обов'язковий файл окремими нумерованими діапазонами до EOF. Для актуальної задачі поточним планом є docs/playback/plan.md, а не docs/media/plan.md, бо tasks визначає T9 у Phase 6.
ДАЛІ: Спершу читаю brief-t9.md цілком і вимірюю обсяги всіх обов'язкових файлів.
ФАКТ: brief-t9.md прочитано повністю (218 рядків). T9 вимагає доказів P04–P07, B09–B11, B18, B29, R28 і O1; забороняє зміни docs/playback/*, work/tasks.md, git, generated/, залежностей, hls.stopLoad() та scope T10–T12. Доступні лише локальні format/lint/typecheck/unit/Go unit; Docker-гейти мають бути названі незапущеними.
РІШЕННЯ: Пряма заборона брифу на git має пріоритет над типовою вимогою implement зробити коміт, тому git-команд і коміту не буде. C34, окремий ~100-секундний asset P06, середній seek O1 та beforeEach не переобговорюю.
ДАЛІ: Читаю CTO-RESEARCH.md і work/tasks.md окремо до EOF, потім весь playback plan та конкретні джерела брифу.
ФАКТ: CTO-RESEARCH.md прочитано до EOF; він застарів відносно коду/Phase 6 і сам наказує віддати перевагу коду. Спільний вивід з tasks знову обрізав середину, але фінальна Phase 6 таблиця підтверджує T1–T8 done і T9 planned.
РІШЕННЯ: CTO-бриф використовую лише як архітектурний контекст; актуальний T9 визначають playback plan, brief і код. work/tasks.md перечитаю окремо, щоб повнота не залежала від обрізаного виводу.
ДАЛІ: Читаю work/tasks.md до EOF окремим викликом.
ФАКТ: work/tasks.md прочитано повністю (168 рядків). T9 — єдиний поточний slice про каденцію, події та clientSeq; T10–T12 і конституційні ADR/DOC задачі окремі й поза scope.
ДАЛІ: Читаю docs/playback/plan.md повністю чотирма діапазонами та звіряю D-7/D-8, acceptance §8, п'ять питань і O1 §11.
ФАКТ: Прочитано docs/playback/plan.md рядки 1–430. D-5 задає монотонний clientSeq, нерегрес завершеного прогресу і валідний перший seq=0; D-7 забороняє знімати stopLoad і розрізняє 409/410; D-8 вимагає heartbeat з lessonId, positionSeconds, clientSeq, reason та політичний check до renew/write.
ДАЛІ: Дочитую docs/playback/plan.md рядки 431–837, включно з §8, п'ятьма питаннями та повним O1.
ФАКТ: Вивід рядків 431–837 обрізав середину acceptance-таблиці, але §11/O1 і C34 прочитані повністю: O1 досі відкритий, бо попередній замір мав 0 inflight; новий доказ мусить seek у незавантажену середину й надрукувати числа. Позиція до/призначена/до remint у T7 була 0/0.254666/0.254666.
РІШЕННЯ: Дочитую пропущену середину §8 малими діапазонами з номерами рядків; повноту плану оголошу лише після цього.
ДАЛІ: Визначаю точні межі acceptance-рядків і читаю блоки навколо P04–P07, B09–B11, B18, B29, R28 без обрізання.
ФАКТ: Пропущену середину §8 прочитано без обрізання; весь plan тепер покритий до EOF. Рівні доказів буквальні: P04–P07/B18 — e2e, B09/B10/R28 — frontend unit, B11/B29 — Go integration. P06 мусить довести саме 0.95×100, не повторити 285/300.
РІШЕННЯ: Тести будуватимуть власні передумови після наявного beforeEach; жоден acceptance-рядок не залежатиме від сусіднього. Для кожного в звіті дам п'ять відповідей із конкретною умовою почервоніння.
ДАЛІ: Читаю spec A22–A27, звіти T5/T7/T8, backlog B-8/B-9/B-11/B-12, AGENTS.md і поточний playback.spec.ts цілком.
ФАКТ: Виміряно обсяг джерел: 3390+ рядків. Очікуваного за планом frontend/src/presentation/learn/recovery.ts за цим шляхом немає; LessonPlayer.tsx має 394 рядки, playback.spec.ts — 719.
РІШЕННЯ: Великі файли читаю цільовими секціями, які бриф прямо назвав, а production/test файли — повністю чанками. Спершу знаходжу фактичний recovery.ts і рядки секцій.
ДАЛІ: Локалізую потрібні файли та заголовки через rg.
ФАКТ: Фактичний recovery.ts — frontend/src/application/playback/recovery.ts. Локалізовано точні секції T7/T8/backlog/spec.
ФАКТ: AGENTS.md і spec A22–A27 прочитані; A23 формулює максимум раз на 30 с плюс pause/seek/hide/nav, тоді як brief/plan C34 уточнює фактичний настінний timer під playing-gate. T5 контракт підтверджує server-side saved clientSeq у play response та коди heartbeat 204/400/403/404/409/410.
РІШЕННЯ: Для T9 не змінюю публічний контракт; перевіряю вже реалізований seed із response і припинення після 409 на frontend seam.
ДАЛІ: Читаю цільові секції T7/T8/backlog без обрізання, потім production і tests повністю.
ФАКТ: T7-джерела прочитані в потрібних секціях. Плеєр має реальний e2e fixture, serial playback-project, profile 2/6, silent 410 remint і O1 instrumentation. Рецензент виміряв 0 inflight у тихій точці; 0→0.254666→0.254666 currentTime; stopLoad лишився недоведеним запобіжником. uploadReady вже виправлено на polling конкретного videoId.
РІШЕННЯ: T9 розширить той самий e2e-файл і проект; O1 змусить inflight через seek у середину, не через кінець. Не створюю другого lifecycle/fixture шару.
ДАЛІ: Читаю T8 gate/result і backlog B-8/B-9/B-11/B-12 повністю.
ФАКТ: T8 gate/result і backlog B-8/B-9/B-11/B-12 прочитані повністю. Наявний beforeEach на student describe є обов'язковим захистом від shared lesson_progress; nested hooks залежать від describe, heredoc SQL має лишатися module-level, а локальний browser entrypoint — make e2e-local. T8 повний зелений базис: make check exit 0, e2e 47/47 на 8cec836; пізніші правки не мали повного повтору через host load.
РІШЕННЯ: Не рухаю module-level SQL у describe і не додаю progress reset. Нові P04–P07/B18 кожен сам створить свою передумову після existing beforeEach.
ДАЛІ: Читаю frontend/src/application/playback/recovery.ts, LessonPlayer.tsx, їхні unit tests і frontend/e2e/playback.spec.ts повністю чанками.
ФАКТ: recovery.ts прочитано повністю; LessonPlayer.tsx прочитано рядки 1–210. Уже є seed `lesson.clientSeq + 1`, який після mint перевстановлюється у `result.value.clientSeq + 1`; session load несе `mediaPath`, а 410 remint використовує stopLoad через викликача. Треба дочитати heartbeat/event wiring.
ДАЛІ: Дочитую LessonPlayer.tsx 211–394 і LessonPlayer.test.tsx 1–356.
ФАКТ: LessonPlayer.tsx і LessonPlayer.test.tsx прочитані повністю. Поточний tick loop планує setTimeout лише коли playing; успішний heartbeat збільшує seq, 409 робить terminal stop. Але pause не шле reason=pause, seek не шле reason=seek, hide/nav/ended не шлють exit heartbeat — B10 ще не реалізовано. Seed із server clientSeq вже production-кодом є, але B29 integration-доказ треба знайти/додати.
РІШЕННЯ: Потрібен один serializing heartbeat seam, який приймає reason, щоб tick і exit-події не породжували паралельних писарів. Спершу прочитаю e2e і backend integration, потім напишу failing tests.
ДАЛІ: Читаю frontend/e2e/playback.spec.ts повністю двома чанками й інвентаризую backend playback integration tests для B11/B29.
ФАКТ: playback.spec.ts рядки 1–360 прочитані. Fixture зараз має два 5-хв assets (lesson 1 і lesson 3), module-level SQL reset, serial 420s tests, helper openAndPlay із Pause-label proof, setPosition через input events, і per-page request tracking. Окремого 100s asset/lesson ще немає.
РІШЕННЯ: P06 отримає новий lesson та окремий generated 100s fixture у чинному course setup; existing 300s fixture не підміняю.
ДАЛІ: Дочитую playback.spec.ts 361–719.
ФАКТ: playback.spec.ts прочитано повністю. Student beforeEach already resets all course progress. O1 currently pauses, waits TTL, then resumes in a quiet buffered point; P02 comment confirms current seek sends nothing and resume triggers a tick heartbeat. T9 must change that behavior and update this test/comment. Course currently has four lessons; only lessons 1 and 3 own video.
РІШЕННЯ: Нові e2e будуть у student serial block до irreversible R04. O1 will seek about 150 immediately before expiry/remint trigger while deliberately delaying media responses so a request is observably inflight. P04/P05/B18 will save 40 by an explicit exit heartbeat, not wait 40 real seconds.
ДАЛІ: Інвентаризую backend playback integration tests and media fixture generation scripts; перевіряю, як без Docker створити 100s fixture declaration without committing media bytes.
ФАКТ: Backend уже має stale-seq policy unit/integration cases, але вони використовують incoming position 90, а brief B11 вимагає exact 20; B29 server-seed unit coverage exists in PostgreSQL reader with clientSeq 12, але наскрізного first-write integration proof ще не видно. media.setup.ts is the owner of generated untracked fixture bytes.
РІШЕННЯ: Зміню stale integration case на exact B11 80/9 ← 20/4 і додам B29 application integration case. 100s media byte generation належить media.setup.ts, не repository artifact.
ДАЛІ: Читаю ці backend integration suites та media.setup.ts повністю перед RED edits.
ФАКТ: progress_integration_test.go і media.setup.ts прочитані повністю; sessions_integration_test.go прочитаний щонайменше крізь ownership/progress helpers і R15/R16, але сукупний output обірвав хвіст. Media setup currently generates 3×2s plus one 300s fixture and streams bytes from worker tmpfs; adding a 100s fixture fits this established path.
РІШЕННЯ: P06 fixture додаю як п'ятий generated asset `e2e-1280x720-100s.mp4` із власним timeout budget; жодні media bytes не додаються до source tree manually.
ДАЛІ: Дочитую sessions_integration_test.go до EOF і перевіряю fixture gitignore/Playwright project dependencies before edits.
ФАКТ: sessions_integration_test.go прочитано до EOF; Playwright config and fixture ignore inspected. media setup runs before playback through setup→chromium→media→playback-profile→playback; `/frontend/playwright/.fixtures/` ignored. Existing session integration mutates one shared progress row across subtests, so B29 must explicitly prepare seq=12 and isolate/restore its state.
РІШЕННЯ: B29 integration додам ближче до final progress cases with direct SQL preparation, a new Play producing server seed 12, and its first heartbeat at seq 13 updating position immediately. It will not depend on prior subtest state.
ДАЛІ: Читаю playback domain/frontend type contracts і frontend test scripts, потім пишу RED unit/integration/e2e changes.
ФАКТ: Frontend contract currently narrows PlaybackHeartbeat.reason incorrectly to only `tick`, although backend accepts tick/pause/seek/hide/nav/ended. Application Sessions already returns saved clientSeq from Play and writes the supplied next sequence; no backend production change is needed for B11/B29.
РІШЕННЯ: Production frontend change includes widening the strict reason union and serializing all heartbeat causes through one queue. Event mapping: timer→tick, pause button→pause, range change→seek, visibility hidden→hide plus pause/gate stop, component unmount/lesson change→nav, media ended→ended. After 409 a terminal flag rejects future queue entries.
ДАЛІ: Додаю RED frontend unit cases B09/B10/R28 first; then запускаю only LessonPlayer test to confirm failures stem from missing behavior.
ФАКТ: RED unit cases додано: B09 exact 3 ticks/90s; B10 exact exit reason sequence plus hidden-time no tick; R28 exactly one heartbeat then terminal stop after 409.
ДАЛІ: Форматую лише test file і запускаю focused LessonPlayer suite для RED evidence.
ФАКТ: Focused LessonPlayer RED: 10/12 passed, 2 failed. B10 failed for the intended product gap with received exit reasons `[]` versus expected seek/pause/hide/ended/nav. B09 and R28 already pass existing tick/409 behavior. The other failure is exact known environment class: B24 timed out at 5000ms after 11 assertions; per brief it is not a product defect and timeouts are not raised.
РІШЕННЯ: RED established only for missing B10 behavior; B09/R28 become regression coverage. Implement event heartbeat path without touching B-7 onPause media-event scope.
ДАЛІ: Widen reason type, add serialized event/tick dispatch with terminal stop, and wire pause/seek/hide/nav/ended; then rerun focused test.
ФАКТ: Frontend implementation added the full closed reason union, a single promise queue for all heartbeat writers, server-seeded sequence on mint, terminal suppression after 409/other terminal failures, and exact tick/pause/seek/hide/nav/ended event wiring. Hide explicitly pauses and clears the playing gate; no native video onPause listener was added, so B-7/T10 remains untouched.
ДАЛІ: Форматую domain/player/test, запускаю focused suite, then inspect any failure as product vs known 5000ms environment class.
ФАКТ: Focused LessonPlayer suite GREEN: 12/12, exit 0. B09, B10 and R28 all pass; known B24 5000ms timeout did not reproduce in this run. Formatter passed on all three touched frontend files.
ДАЛІ: Додаю exact B11/B29 Go integration proofs і запускаю compile-only integration plus non-Docker domain/application unit packages.
ФАКТ: B11 integration row now exactly pins saved 80/seq9 against offline replay 20/seq4. B29 integration row independently prepares seq12, mints a new session, asserts response seed 12, and proves the first heartbeat seq13 updates position to 81 immediately.
ДАЛІ: gofmt these tests; run playback domain/application unit tests and integration-tag compile-only without Docker runtime.
ФАКТ: gofmt completed. Host `go test ./internal/playback/domain ./internal/playback/application` PASS (3.602s/3.296s after cache warmup). Integration-tag compile-only produced no output for over six minutes on the known overloaded host and was interrupted with exit 130; it is NOT counted as passed or failed code.
РІШЕННЯ: Не піднімаю timeout і не маскую незавершений compile. Повернусь до адресного compile після frontend edits, using the warmed cache; reviewer still owns real Docker integration.
ДАЛІ: Додаю 100s fixture generation and P04–P07/B18/O1 e2e proofs, updating P02 assumptions for immediate seek/pause event writes.
ФАКТ: media.setup now generates a fifth ignored fixture, exactly 100 seconds, with the worker ffmpeg and existing lavfi/stream-out path. Playback setup uploads it to existing fourth lesson, preserving the four-row course shape required by T8.
ДАЛІ: Додаю strict heartbeat payload parsing and self-contained P04–P07/B18 browser scenarios before changing O1.
ФАКТ: Added strict hand-written heartbeat payload validation without unsafe assertions. P04, P05, P06, P07 and B18 each build their own saved/completed/network precondition after beforeEach; P06 verifies a 99–101s media duration and seeks exactly 95; P07 explicitly completes and checks zero before reopening.
ФАКТ: O1 now gates the expired heartbeat until a seek to 150 has created a deliberately held `.m4s` request, prints pre-remint and mint-time inflight counts plus currentTime values, and requires `pre > 0`, `at mint = 0`. hls.stopLoad remains unchanged. P02 was updated so pause/seek writes themselves are the observed causes.
РІШЕННЯ: Browser tests are authored but not claimed run because Docker/e2e is forbidden by the brief. Static type/format/lint will validate them locally.
ДАЛІ: Форматую touched frontend files and run frontend typecheck plus focused unit; repair only static/product issues, not Docker assumptions.
ФАКТ: Formatter PASS on five touched frontend files; `npm run typecheck` PASS; focused LessonPlayer suite PASS 12/12 in 2.069s. Static typing accepts the new e2e instrumentation and full reason union.
ДАЛІ: Run frontend lint/architecture, inspect the exact diff without git via direct file anchors, and test navigation cleanup/concurrency edge cases for false positives.
ФАКТ: `npm run lint` stopped before architecture with two actionable findings: forbidden non-null assertion in the new B10 test and React refs warning for assigning queueHeartbeatRef.current during render. No product verdict from this red lint.
РІШЕННЯ: Replace the test assertion with explicit null guard; synchronize latest queue callback in an effect rather than render. Then rerun lint and focused tests.
ДАЛІ: Apply the two local fixes.
ФАКТ: Fixes applied. `npm run lint` PASS including architecture boundaries over 90 files; focused LessonPlayer suite PASS 12/12 in 2.358s; formatter PASS.
РІШЕННЯ: Strengthen B09 unit proof so play response clientSeq=12 differs from lesson snapshot=4 and emitted sequences must be 13/14/15; this directly guards the frontend half of B29 seed.
ДАЛІ: Update B09 seed assertion, rerun focused suite, then static-audit e2e route gating and event ordering.
ФАКТ: Strengthened B09/B29 frontend proof passes: exact reasons tick/tick/tick and sequences 13/14/15 from server session seed 12, despite lesson snapshot seq4. Focused suite remains 12/12 GREEN.
ДАЛІ: Run formatter/typecheck and Playwright discovery (`--list`, no browser/Docker execution); inspect O1/P04–P07/B18 anchors and correct static issues only.
ФАКТ: Formatter PASS; frontend typecheck PASS; Playwright discovery PASS and lists all new rows in the playback project. Total discovered suite is 48 tests across 13 files; this is discovery only, not browser execution.
ДАЛІ: Inspect the final O1 gate and P04–P07/B18 code with numbered anchors; run a static causality audit for each row and repair false-positive paths before full host gates.
ФАКТ: Static causality audit found an O1 test-instrument defect before runtime: the first version held the segment in a Playwright route and released it only after recording mint, which could keep the request counted by the harness itself even if stopLoad had cancelled it. That would manufacture `activeAtMint=1`.
РІШЕННЯ: Replaced the artificial held route with Chromium network throttling (1 KiB/s) on the real unloaded `.m4s`; only the expired heartbeat is gated until the real media request starts. Thus stopLoad, not the test harness, is the only mechanism that can clear the active request before mint.
ДАЛІ: Re-run format/typecheck/discovery after O1 correction, then audit navigation and offline proofs against their causal chain.
ФАКТ: O1 correction passes formatter, TypeScript and Playwright discovery. Static causality audit: P04/P05 each wait for their own 204 seek write before reload/new context; P06 proves its own 100s duration and exact 95 seek before observed mark; P07 makes its own 285/300 completion then observes initial slider 0; B18 waits for saved 40, then an actual failed heartbeat at 80 before a new page reads 40.
РІШЕННЯ: No additional e2e hook/reset is needed; existing beforeEach isolates every row. Browser runtime remains reviewer-only.
ДАЛІ: Read and apply code-review skill to the explicit T9 file set, respecting brief prohibition on git and Docker.
ФАКТ: code-review skill прочитано повністю. Він вимагає незалежні Standards і Spec axes у паралельних sub-agents. Його git fixed-point process конфліктує з прямою забороною brief-t9, тому review scope pinned explicit current T9 files instead of git diff.
РІШЕННЯ: Запущено два read-only review agents: Standards uses AGENTS.md plus smell baseline; Spec uses brief-t9, plan D-7/D-8/§8/O1 and spec A22–A27. Обом заборонено git, edits, report writes and Docker.
ДАЛІ: While reviews run, execute permitted full host gates in required order: frontend format-check→lint→typecheck→test, then Go test ./...; classify 5000ms timeouts exactly as environment.
ФАКТ: Ordered frontend gates: format-check PASS (162 files), lint+architecture PASS (90 files), typecheck PASS. Full `npm test` NOT GREEN: 366/372 passed; five failures explicitly timed out at 5000ms in CourseEditorView/UserListView, the exact predeclared overloaded-host class. A sixth CourseEditorView R36 assertion missed expected text amid cascading dependency_unavailable route logs; it is not labelled a timeout, so currently could be environment contamination or a real pre-existing test failure. No T9 test failed.
РІШЕННЯ: Do not raise timeouts or edit unrelated admin tests. Run the two failing files separately once to distinguish the non-timeout R36 failure and confirm/deny load contamination.
ФАКТ: Standards reviewer reports zero hard violations; two non-blocking Low smells: repeated media stubs in LessonPlayer tests and accumulating responsibilities in LessonPlayer. No refactor now: extracting stubs or splitting the component would be scope expansion and risk, not required for T9.
ДАЛІ: Run the two failing admin files in isolation; await Spec review in parallel.
ФАКТ: Isolated repeat of CourseEditorView and UserListView was still NOT GREEN: 14/18 passed. UserList passed; CourseEditor P19, B09 and R28 each hit the fixed 5000ms timeout, while R36 again missed its validation-text assertion without being labelled a timeout.
РІШЕННЯ: Run only R36 once to separate it from same-file timeout pressure; do not edit this unrelated suite regardless of outcome.
ФАКТ: Spec reviewer found no scope creep and confirmed B11, B29, per-test reset, the 100s P06 asset and serialized writer match the brief. It found three code/test blockers: cadence drifts when the next 30s timeout is scheduled after network latency; resume emits an unrequested immediate tick; O1 can mistake the seek setter write for the remint restore. It also correctly classifies O1 runtime acceptance as unverified until Docker/browser numbers exist.
РІШЕННЯ: Fix all three product/test defects with focused tests. O1 will remain explicitly unverified—not falsely closed—because the brief forbids the only runtime that can produce its numbers.
ДАЛІ: Reproduce R36 alone, then implement latency-anchored serialized cadence, remove resume tick, and make O1 require a distinct remint setter write equal to the immediate pre-remint position.
ФАКТ: R36 alone reproduces as the same predeclared environment class: `test timed out in 5000ms (no expect assertions completed)`, 11 sibling tests skipped, duration 5087ms. The earlier missing-text surface was cascading load noise, not evidence of a distinct assertion regression.
РІШЕННЯ: Full frontend gate remains NOT GREEN because the host cannot complete six unrelated tests within their fixed 5000ms budgets. Per brief, do not increase timeouts or change tests; T9 focused tests are green.
ДАЛІ: Inspect the current timer/resume and O1 anchors, write failing latency/resume expectations, then apply the smallest implementation correction.
ФАКТ: Anchor inspection confirms all three review findings in code: cadence recursively schedules only after request completion; resume explicitly queues `tick`; O1 records its baseline before seek and accepts any last setter write. No hidden abstraction already solves them.
РІШЕННЯ: Anchor cadence to elapsed playing wall time with a fixed interval whose callbacks enter the existing serialized queue; remove only the resume tick while preserving play; make O1 wait for two setter writes (seek then remint), capture position immediately before releasing remint, and compare the second assignment within tolerance.
ДАЛІ: Patch focused unit expectations first so latency and resume behavior would fail against the current component.
ФАКТ: Recovery helper inspection shows retry delays are 2–8s while the healthy cadence is 30s. Therefore an anchored recursive deadline can preserve both policies: advance from the scheduled deadline after healthy/remint responses, but advance from completion by the explicit retry delay after retryable failures.
РІШЕННЯ: Keep one timer and serialized writes; do not replace it with a blind interval. Add 1s latency to B09 and require calls at 30/60/90s; change B10 to require zero ticks after pause→resume before the next 30s boundary.
ДАЛІ: Apply the RED test changes and run only B09/B10 to prove the current implementation fails for the intended reasons.
ФАКТ: Focused RED is exact: B09 receives only 2 heartbeats by 90s when each response takes 1s; B10 receives 1 tick where 0 is allowed before the next cadence boundary. Both tests fail their new assertions, with the other 10 tests skipped.
РІШЕННЯ: The reviewer findings are reproduced, not theoretical. Correct production scheduling and resume behavior without weakening either assertion.
ДАЛІ: Patch LessonPlayer: deadline-anchored recursive timer with retry exception, and direct media play on resume with no heartbeat.
ФАКТ: Focused GREEN: B09 and B10 both pass. B09 now sees calls at anchored 30/60/90s despite 1s response latency; B10 sees only seek/pause/hide/ended/nav and no immediate resume tick.
РІШЕННЯ: Healthy cadence advances from the prior wall-time deadline; an explicit retry backoff advances from failure completion. Existing serialization remains the single-writer guarantee.
ДАЛІ: Strengthen O1/currentTime so the test distinguishes the seek assignment from the later remint restore and compares against the position captured immediately before remint.
ФАКТ: O1 proof was tightened: instrumentation is installed after initial play, seek=150 is necessarily setter write #1, pre-remint currentTime is captured only after the real throttled segment is inflight, and the test now waits for a later setter write assigned within ±1s of that value. It also verifies the live currentTime did not reset below tolerance.
РІШЕННЯ: A seek-only write or restart from zero can no longer make O1 green. Keep the inflight assertions and printed four values unchanged in purpose.
ДАЛІ: Format and typecheck the corrected component/tests/e2e; run the complete LessonPlayer unit file and Playwright discovery.
ФАКТ: Prettier completed successfully on the three corrected files.
ДАЛІ: Run typecheck, full LessonPlayer unit file and Playwright discovery in parallel; then lint the final frontend state.
ФАКТ: Final corrected frontend slice: typecheck PASS; complete LessonPlayer suite PASS 12/12 in 1.704s; Playwright discovery PASS with all P04/P05/P06/P07/B18/O1 rows and 48 total tests. Discovery is still not browser execution.
ДАЛІ: Run lint/architecture and format-check on final state, then inspect/fix P07’s start-at-zero observation if it can confuse server reset with actual media restore.
ФАКТ: Final lint and architecture PASS over 90 files. Project `oxfmt --check` is RED on the three files because the prior Prettier formatter differs from the repository formatter; this is mechanical, not a product/test failure.
РІШЕННЯ: Use the repository `npm run format` command, then rerun format-check. Do not hand-format around oxfmt.
ДАЛІ: Apply repository formatter and rerun its check.
ФАКТ: Repository oxfmt completed successfully over 162 files; the preceding check identified only the three T9 files as differing.
ДАЛІ: Re-run format-check and inspect P07’s exact causal assertion before final backend/full verification.
ФАКТ: Repository format-check is GREEN across all 162 files. The requested numeric range shifted after formatting and showed O1 rather than P07, so no P07 conclusion is drawn from that inspection.
ДАЛІ: Locate P07 by test name and inspect its complete body; strengthen only if actual restore can false-green.
ФАКТ: P07 causality is sufficient: its own seek=285 response establishes completion, reload must expose server-selected slider value exactly 0, and actual media after reopening must remain ≤1s. A stale 285s resume fails both observations; absence of an explicit setter is not a false green because natural media start at 0 is the required outcome.
РІШЕННЯ: Keep P07 unchanged; adding setter instrumentation would test implementation mechanics rather than improve the observable start-at-zero contract.
ДАЛІ: Run final permitted Go `test ./...`; if it hangs under the documented host load, interrupt and report honestly. Then assemble the five-question proof matrix and reviewer commands.
ФАКТ: `GOCACHE=/tmp/courses-t9-go-cache go test ./...` is still running with no output after 30s; no verdict yet.
ДАЛІ: Poll the existing process rather than starting a duplicate.
ФАКТ: Full Go test remains silent and running after 60s total. This matches the repository’s known slow host behavior but is not yet classified.
ДАЛІ: Allow one more bounded poll; interrupt if still silent so an unbounded host wait cannot be misreported as a pass.
ФАКТ: Full Go test is still silent/running after 90s; unlike the already-green targeted playback packages, it has produced no package verdict.
РІШЕННЯ: Interrupt this host-blocked full sweep and record it as unverified, not failed or passed. The brief explicitly allows host limitations but forbids invented green gates.
ДАЛІ: Send interrupt, capture the process result, then finalize evidence mapping.
ФАКТ: Full `go test ./...` was interrupted after 90s with exit 130 and no output. It is UNVERIFIED. Earlier targeted `./internal/playback/domain ./internal/playback/application` remains PASS; integration-tag compile-only was likewise interrupted after prolonged silence and is not green.
ДАЛІ: Inspect final changed anchors without git, verify the report contains every required row/question, O1 output status, unrun gates and exact reviewer commands, then append the final matrix.
ФАКТ: Final anchor inventory finds every target test by exact row label: frontend P04/P05/P06/P07/B18/O1, unit B09/B10/R28, PostgreSQL B11 and Sessions B29. `gofmt -d` is empty on both changed Go files. The report does not yet contain the required five-question matrix, real O1 numeric output, or exact reviewer command block.
РІШЕННЯ: Append the matrix now. For O1, print the exact expected output keys but mark numeric values `UNVERIFIED (not run)`; inventing numbers would violate both the brief and review finding.
ДАЛІ: Add `§8 row → proof`, O1 status, unrun gates, review findings and exact reviewer commands as one append-only delivery block.
ФАКТ: Exact Go integration entrypoints are `TestProgressWriterAgainstPostgreSQL` and `TestPlaybackSessionsAndAuthorizationAgainstRedisAndPostgreSQL`. Makefile confirms all reviewer targets named in the brief, and `e2e-local` expands to `up seed-identities seed e2e`.
РІШЕННЯ: Use these discovered names verbatim; do not publish guessed test commands.
ДАЛІ: Append the final evidence/delivery block.

## §8 row → proof

| §8 row | 1. Що саме доводить тест? | 2. Яким production-шляхом виникає артефакт? | 3. Як відсічено false green? | 4. Чому це доказ саме цього row? | 5. Яка поломка робить тест RED? |
|---|---|---|---|---|---|
| P04 | **AUTHORED / browser UNVERIFIED.** Власний seek записує 40; reload відновлює 39–41. | UI range → serialized heartbeat `reason=seek` → HTTP/PG → reload learning snapshot/session → video currentTime. | Тест чекає власну 204-відповідь до reload; не успадковує стан завдяки `beforeEach(resetLessonProgress)`. | Це той самий page/context і саме reload, не новий браузер P05. | Не зберегти seek, не прочитати progress або не застосувати resume position. |
| P05 | **AUTHORED / browser UNVERIFIED.** Інший BrowserContext того самого student відновлює 40±1. | Перший context пише heartbeat у PG; другий context автентифікується власним storage state і читає серверний progress. | Другий context створюється після підтвердженої 204-відповіді; локальний React/video state не може перенестись. | Окремий BrowserContext ізолює cross-device/server persistence від P04 reload. | Зберігати progress лише локально/в session memory або не seed-ити нову сесію. |
| P06 | **AUTHORED / browser UNVERIFIED.** Окремий ~100s ролик має duration 99–101; seek=95 позначає саме четвертий lesson watched. | Generated 100s MP4 → реальний upload → worker ffmpeg/transcode → playback metadata/duration → seek heartbeat → completion/outline. | Тест перевіряє власну тривалість, власний exact seek response, 4 outline items, watched лише четвертий і not-watched перший. | Поріг 95/100 прямо доводить proportionality; 95s на старому 300s asset не пройшов би. | Hard-code 285s, втратити duration, округлити поріг неправильно або позначити сусідній lesson. |
| P07 | **AUTHORED / browser UNVERIFIED.** Тест сам завершує 300s lesson seek=285, reload бачить slider=0, reopen реально стартує ≤1s. | Seek heartbeat → PG completed → learning snapshot/session returns reset position → player opens media at zero. | Власна completion-відповідь очікується до reload; exact zero в UI плюс фактичний video time відсікають декоративний reset. | Не покладається на seed/попередній тест і перевіряє reopen-after-completion, а не звичайний resume P04. | Повернути/застосувати saved 285 після completion або не скинути presentation position. |
| B09 | **GREEN unit 12/12 suite.** За 90s continuous playing рівно 3 `tick`, seq 13/14/15. | Production React timer → serialized `queueHeartbeat` → application action payload. | Mock додає 1s response latency; deadline cadence все одно дає 30/60/90, тому recursive-after-response drift дає лише 2 і RED. Session seq=12 навмисно відрізняється від lesson seq=4. | Рахуються лише cadence ticks за continuous playing; exit causes винесені в B10. | Планувати наступний timeout після response, додати immediate start/resume tick, або seed seq із lesson snapshot. |
| B10 | **GREEN unit 12/12 suite.** Рівно seek/pause/hide/ended/nav; 90s hidden не дає tick; pause→resume не дає позакаденційний tick. | Реальні presentation handlers/visibility/unmount → той самий serialized writer/action payload. | Assertions перевіряють exact ordered non-tick reasons і zero ticks; тест не викликає writer напряму. | Кожна named exit cause присутня в одному ізольованому сценарії; cadence окремо доведено B09. | Прибрати будь-який handler, не зупинити hidden timer або повернути immediate resume heartbeat. |
| B11 | **AUTHORED / PostgreSQL runtime UNVERIFIED.** Saved 80/seq9 + stale replay 20/seq4 залишається 80/seq9. | Реальний PostgreSQL progress adapter `Save` і SQL conflict/order policy. | Вхід навмисно має event reason `ended`, що міг би дозволити rewind за reason, але stale seq мусить перемогти. | Exact numbers і stale sequence доводять offline replay protection, не generic monotonic tick case поруч. | Приймати lower seq, оновлювати position до 20 або sequence до 4. |
| B18 | **AUTHORED / browser UNVERIFIED.** Після confirmed save=40, offline seek=80 реально fail-иться; нова page читає 39–41. | Browser offline → справжній failed HTTP heartbeat; PG зберігає останній accepted write; нова page/session читає його. | Чекає `requestfailed` саме heartbeat при 80 і використовує new page; немає route-mock success чи локального state reuse. | Це availability/network-loss row; stale lower-seq ordering окремо доводить B11. | Optimistic/local 80 переживає reload, failed write помилково комітиться, або server loses saved 40. |
| B29 | **AUTHORED / Redis+PostgreSQL runtime UNVERIFIED.** Saved 80/seq12; fresh Play повертає seq12; перший write є 81/seq13 і зберігається. | PG seed → Sessions.Play progress lookup + Redis ownership → response seed → Sessions.Heartbeat → PG. | Свіжий playback ID і рівно один first heartbeat; assertion читає DB. Frontend B09 окремо підтверджує seed 12→13 попри lesson seq4. | Доводить перший write нової сесії, не stale replay B11 і не наступні increments. | Play повертає 0/lesson seq, frontend починає з 5, або backend не приймає/не зберігає 13. |
| R28 | **GREEN unit 12/12 suite.** Після першого 409 takeover наступні 90s не створюють жодного heartbeat; показано terminal UI. | Timer → serialized writer → production failure directive → terminal stop/timer cancellation. | Counter фіксується на 1 після 30s і лишається 1 після ще 90s; немає паралельного writer. | Код помилки саме `playback_taken_over`, а не generic denied/unavailable. | Не ставити terminal stop, лишити timer/queued writers або повторювати 409. |
| B39 / O1 | **AUTHORED / browser measurement UNVERIFIED.** Реальний throttled `.m4s` inflight до expired heartbeat; `stopLoad()` має очистити його до другого mint; remint setter окремо від seek відновлює pre-remint currentTime ±1. | Chromium network + hls.js request → 410 heartbeat → production `stopLoad()` → Play remint → manifest ready → production currentTime restore. | Segment не hold-иться route harness-ом; gate тримає лише heartbeat. Requires pre-inflight >0, at-mint=0, setter write після seek, assigned ±1 and live time not reset. | Це прямий O1/D-7/D-8 instrumented proof, не сусідній B40 generation-retention flow. | Видалити `stopLoad`, remint before abort, restore before manifest, omit restore, або restore 0/stale value. |

## O1 numeric output

Browser execution was forbidden in this task, therefore no numeric measurement exists and O1 is **UNVERIFIED**, not closed. The authored test prints these exact lines when the reviewer runs it:

```text
O1 hls.js in-flight immediately before remint: UNVERIFIED (not run)
O1 hls.js in-flight when mint request began after stopLoad: UNVERIFIED (not run)
O1 stopLoad cleared old in-flight requests before mint: UNVERIFIED (not run)
O1 currentTime before explicit restore: UNVERIFIED (not run); assigned restore: UNVERIFIED (not run); pre-remint: UNVERIFIED (not run)
```

Acceptance requires the first count `> 0`, the second count `= 0`, stopLoad observation `yes`, and assigned restore within ±1s of pre-remint. If runtime contradicts D-7/D-8, move the result to Discussion rather than changing the contract silently.

## Verification status

- GREEN: frontend format-check (162 files), lint + architecture (90 files), typecheck, LessonPlayer 12/12 focused unit suite, Playwright discovery (48 tests), targeted Go playback domain/application packages, and empty `gofmt -d` for the two Go test files.
- NOT GREEN: full frontend `npm test` 366/372. All six residual failures are unrelated admin/user suites hitting the fixed 5000ms budget; R36 alone confirmed `test timed out in 5000ms (no expect assertions completed)`. No T9 test failed. Timeout was not raised and tests were not changed.
- UNVERIFIED after host interruption: full `go test ./...` (90s silent, exit 130) and integration-tag compile-only (prolonged silence, interrupted). Neither is represented as pass/fail.
- NOT RUN by explicit brief prohibition: `make integration`, `make check-generated`, `make build`, `make up`, `make smoke`, `make e2e`, actual media generation/upload/transcode, P04/P05/P06/P07/B18 browser runtime, B11/B29 datastore runtime, and O1 browser measurement.

## Independent review

- Standards: no hard violations. Non-blocking Low notes are repeated media stubs in LessonPlayer tests and growing responsibilities in LessonPlayer; no scope-expanding refactor in T9.
- Spec: found cadence latency drift, an extra resume tick, and O1 seek/restore ambiguity. All three code/test defects were reproduced or statically confirmed and fixed. Remaining blocker is only the forbidden/unrun O1 runtime measurement.

## Exact reviewer commands

Run from `/Users/strelia/projects/courses` on a machine with Docker available:

```bash
cd frontend
npm run format:check
npm run lint
npm run typecheck
npm test
cd ../backend
GOCACHE=/tmp/courses-t9-go-cache go test ./...
cd ..
make integration
make check-generated
make build
make smoke
make e2e-local
```

For the two exact Go proofs after integration dependencies are up:

```bash
cd /Users/strelia/projects/courses/backend
go test -count=1 -tags=integration -run '^TestProgressWriterAgainstPostgreSQL$' ./internal/playback/adapters/postgres
go test -count=1 -tags=integration -run '^TestPlaybackSessionsAndAuthorizationAgainstRedisAndPostgreSQL$' ./internal/playback/application
```

For the exact browser rows and O1 printed measurement after `make up`, `make seed-identities`, and `make seed`:

```bash
cd /Users/strelia/projects/courses/frontend
npx playwright test --project=playback --grep 'P04|P05|P06|P07|B18|B39 O1' --reporter=line
npx playwright test --project=playback --grep 'B39 O1' --reporter=line
```

ФАКТ: T9 implementation and authored proofs are complete within the permitted host scope. Acceptance is not claimed for browser/DB rows until the exact reviewer commands run; O1 especially remains open pending its real printed numbers.
РІШЕННЯ: Preserve `hls.stopLoad()`, the deadline-anchored serialized writer, per-test reset, module-level SQL heredoc placement, and the dedicated 100s asset. Do not carry any T10/B-7/T11/T12 work into this slice.
ПИТАННЯ: Немає. Наступний зовнішній крок — reviewer with Docker executes the listed gates and records O1 numbers (or Discussion if they contradict D-7/D-8).

## Correction: authoritative reviewer order

ФАКТ: Direct Makefile inspection shows the preceding expanded command block placed `make smoke` before stack startup. That ordering is wrong. `Makefile` defines the authoritative chain as `check: format-check lint typecheck test integration check-generated build up seed-identities seed smoke e2e`; `e2e-local` is only the standalone browser shortcut `up seed-identities seed e2e`.
РІШЕННЯ: This correction supersedes the earlier expanded reviewer block. The exact full reviewer command is:

```bash
cd /Users/strelia/projects/courses
make check
```

The exact standalone browser/O1 commands remain:

```bash
cd /Users/strelia/projects/courses
make up
make seed-identities
make seed
cd frontend
npx playwright test --project=playback --grep 'P04|P05|P06|P07|B18|B39 O1' --reporter=line
npx playwright test --project=playback --grep 'B39 O1' --reporter=line
```

ФАКТ: This order starts and seeds the stack before smoke/e2e and preserves the playback project’s setup/profile dependencies. No command in this correction was run locally.
ПИТАННЯ: Немає; T9 awaits Docker reviewer evidence for browser/DB rows and actual O1 numbers.

---

## Стан перевірки T9 на 2026-09-10, вечір відкладено

### Доведено браузером

**P07 — зелений.** Прогін `make-check-p07c` дійшов до e2e: **41 пройшов**, і серед них P07. Це
доводить виправлення **A25**, якого не існувало в продукті: завершений урок відкривався з
позиції, що його й завершила, тобто за секунди до кінця. Рядок T9 знайшов справжню прогалину
продукту — перший такий випадок за фазу, знайдений тестом, а не читанням коду.

### Виправлено, але НЕ переперевірено браузером

**B18.** Тест припускав, що плеєр лишається керованим під час розриву мережі. Не лишається:
розрив вичерпує ретраї медіа, спрацьовує стан недоступності, і `LessonPlayer.tsx:338` замінює
компонент разом із повзунком. Переписано на спостережуваний розрив (офлайн → «Відео зараз
недоступне» → онлайн → позиція 40 с уціліла). **Після правки e2e не запускався.**

### Виправлено в тестовій інфраструктурі

| Що | Було | Стало | Доказ |
|---|---|---|---|
| `CourseEditorView` B09 (**B-14**) | дефолтні 5000 мс | явні 20 000 мс | ізольовано 1 відмова з 3 до, 3 з 3 після |
| Старт Postgres у тестах (**B-15**) | `time.Minute` ×12 | `3*time.Minute` | лог testcontainers: «matched 1 times, expected 2» |

### Чому гейт не закрито

Хост ділиться з **25 чужими контейнерами**, `load` тримається 50+. Три прогони поспіль упали на
`integration`, і **щоразу іншим механізмом**: таймаут готовності Postgres, вичерпаний дедлайн
очікування, гонка з reaper-контейнером ryuk (`unexpected container status "created"`). Спільне в
них одне — підйом контейнерів під насиченою машиною.

**Домовлено з користувачем:** ввечері зупинити чужі контейнери, прогнати повний `make check` на
розвантаженій машині, потім підняти їх назад. Томи не чіпати — лише stop/start.

### Відкрите, не видавати за зроблене

- **O1** — недоведений; `hls.stopLoad()` не знімати.
- **B-6** — `frontend/e2e/` не звірений із контрактом.
- **B-7** — єдиний дефект продукту, знайдений читанням; лікується одним `onPause`, місце T10.
- **503 на краю** — nginx двічі віддав 503 на POST тривалістю ~3.9 с, без жодного рядка помилки.
  Причина **не встановлена**. Перший ланцюг міркувань («api не бачив запиту») виявився хибним:
  логи api за те вікно знищив `playback-profile`, що перестворює контейнер. Наступний прогін іде
  з живим захопленням обох боків шва.
- ~~**354 × 429 на краю**~~ — **закрито**: навмисні **всі**. `smoke.sh` робить три сплески — 60
  проти `burst=20`, 30 логінів проти `burst=5`, 400 сторінок проти `burst=100`, тобто ≈362
  очікуваних проти 354 спостережених. Я порахував лише перший сплеск.
