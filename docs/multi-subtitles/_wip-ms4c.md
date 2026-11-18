NEXT: Перевірити правила входу, стан репозиторію та контракт MS4 (§0 O7, §3 M3–M6/M8, §4 M04–M07/M09).
FACT: Прочитано shared agent-entry; далі перевіряю локальні правила та прийнятий план. Делегування не планується, git mutations заборонені.
FACT: Корінь /Users/strelia/projects/courses, гілка ms4-player-captions. Наявні сторонні dirty/untracked файли збережено; жодних git mutations. Прочитано AGENTS.local.md; уточнюю bounded reads після завеликого briefing output.
NEXT: Прочитати M4/M5 повністю та поточні route/player/panel/query API перед реалізацією.
FACT: M4 задає off/primary/exact; M5 вимагає render-time mount identity і відкидання застарілих подій; M6/M8 вимагають keyed track/panel та 5 panel states. Використовую ponytail для мінімального рішення з існуючими Query options.
FACT: Learn route поки читає primary для всіх відкритих уроків і губить query errors; TranscriptPanel має лише transcript prop. Query factory вже підтримує language у ключі та repository call. Preferences і predictedAudioLanguage готові до reuse.
NEXT: Перевірити caption callback/effect та test harness; потім додати чисті query/panel state helpers.
FACT: LessonPlayer callback зараз не має mount; captions mode читає video.textTracks[0], track має uk fallback; settings menu приймає optional captions switch. Планую явний captionsNoneForTrack prop та локальний report стан вибору, без впливу на media instance dependencies.
DECISION: 2026-11-18, Codex: helper-и розділено по inward boundaries: playback визначає query mode та mount reducer, transcript визначає panel state з query result. Перевірка: architecture lint + unit M4/M5. Не додаю shared contract для UI-only композиції.
NEXT: Реалізувати helpers, route wiring, keyed panel/track, copy та retry.
FACT: Додано playback captionQuery/reducer і transcriptPanelState; TranscriptPanel отримав явний state, loading використовує наявний common.loading, error має retry button. Далі підключаю маршрут і плеєр.
FACT: Маршрут тепер обчислює mount/query/panel у render; email gate вимикає transcript query і панель; captions надсилаються лише після події load поточного mount та лише з transcript state. LessonPlayer повертає mount, звітує selection, keyed track використовує ref; gear має none text row. Copy додано uk/en.
NEXT: Додати тести M4/M5/M7/M9 та route matrix, перевірити типи.
FACT: API зміни застосовані без нових залежностей; старі panel/player unit callers адаптовані. Запущено попередній typecheck для раннього виявлення type/API помилок; це ще не фінальний ordered gate.
FACT: Додано unit таблиці M4/M5 (включно und-family, mount A gen2→B gen1, older gen, retry без prediction) та M09 (loading/none/error/transcript і key reset пошуку+follow). Попередній typecheck завершився; результат вище інструментом.
FACT: Попередній typecheck FAIL: масова заміна JSX зачепила Partial<LessonPlayerActions> у тесті (TS1005). Виправлено конкретний token; до фінальних gates лишається route та player coverage.
FACT: Route tests додані у наявний route-guards harness: primary actual calls для null/processing/failed, locked/email off, prediction з preferences/default/first, und без request, foreign response, post-load pending/confirmed, retry exact refetch. Кожна нова fixture очищує Query cache і preferences.
FACT: M07 unit тепер перевіряє identity track DOM, current-ref mode (старий textTracks[0] лишається disabled), captionsOn true через none і назад, gear row без switch, report pending/confirmed/unavailable та media creations=1. Retry використовує існуючий Button (semantic styling/focus/control size). Запущено format + попередній typecheck.
NEXT: Перевірити typecheck result і targeted MS4 tests із timeout 30000; після виправлень виконати повний замовлений ordered frontend gate. E2E не запускати (run d).
FACT: Попередній npm run format PASS (412 files), npm run typecheck PASS (exit 0). Targeted 5 test files запущено з --testTimeout 30000; лог /tmp/ms4c-targeted.log. Жодного default-5s запуску не було.
NEXT: Під час self-review перевірити stale callback після A→B: reducer має порівнювати event з поточним mount навіть коли подію доставив старий callback; додати route navigation regression.
FACT: Перший targeted run має failures; збираю короткі failure headings замість raw verbose dumps. Email-gated fixture помилково чекала heading плеєра, який саме має бути відсутнім — виправлю очікування на email gate.
DECISION: Незалежне review цього run не виконується: поточне доручення author-only (run c), owner — користувач, exact reviewer reference pending; delivery owner Codex; explicit permission є лише на frontend commands, shared e2e slot не використовується. Self-review не є independent PASS.
FACT: Targeted-1 FAIL (124 passed, 5 failed): 4 fixture waits чекали відсутній player heading у non-ready/email; 1 switch assertion читав aria-checked замість native checked. Виправлено test expectations. Self-review також виявив реальний stale-closure ризик: старий callback міг прийняти власний mount і стерти B record.
DECISION: 2026-11-18, Codex: current mount зберігається разом з одним record та синхронізується під час render (record не скидається); stable callback reducer читає current.mount. Без ref write/reset effect; перевірка — новий route A→B→A late-event test + reducer unit.
FACT: Попередній format/test retry не стартував: npm у корені repo ENOENT package.json (exit 254); виправлено cwd на frontend і повторно запущено format та targeted-2.
NEXT: Перевірити targeted-2 result, провести focused diff review (query in-flight, mount lifecycle, i18n, track mode), далі фінальний ordered gate.
FACT: Targeted-2 PASS: 130/130 tests, 5/5 files, 0 skipped, exit 0. Focused review: in-flight query тепер включає isFetching (M5 loading навіть при background refetch); додано held-response route test і перевірку audio preference write у none зі збереженим captionsOn=true.
NEXT: Виконати послідовно npm run format → format:check → typecheck → lint → npx rstest run src --testTimeout 30000; записати кожний реальний exit/result.
FACT: Final gate 1/5: cd frontend && npm run format — PASS, exit 0 (oxfmt, 412 files).
FACT: Final gate 2/5: cd frontend && npm run format:check — PASS, exit 0 (all matched files formatted).
NEXT: Дочекатися фінального typecheck; за PASS запустити lint, потім повний rstest src (30000 ms).
FACT: Final gate 3/5: npm run typecheck — PASS, exit 0. Final diff review виявив canonical-language detail: route має передавати нижній регістр query language у track, а не raw response spelling; виправлю й повторю ordered gate після цієї дельти.
FACT: Canonical track language виправлено: route передає language з exact query (lowercase), а не response spelling; held-response route fixture повертає uppercase EN і чекає srclang=en. Final gate restart 1/5 npm run format PASS, exit 0 (412 files).
FACT: Final gate restart 2/5 npm run format:check PASS, exit 0.
NEXT: Дочекатися restart typecheck, потім lint і повний unit/route src suite.
FACT: Final gate restart 3/5 npm run typecheck PASS, exit 0. Запущено 4/5 npm run lint.
NEXT: Після lint result виконати final suite; результати run c також додати у поточний phase report, не закриваючи MS4/e2e/design acceptance.
FACT: Final gate 4/5 npm run lint FAIL, exit 1: exhaustive-deps відхиляє дослівний M5 useMemo(() => Symbol(player), [lessonId]) як unnecessary dependency.
DECISION: 2026-11-18, Codex: використовую Symbol(`player:${lessonId}`), [lessonId] — lessonId реально читається у factory, identity та A→B→A семантика незмінні, lint rule не вимикається. Перевірка: npm run lint + route mount regression. Це технічне уточнення snippet M5, не зміна контракту.
FACT: Final gate after lint fix 1/5 npm run format PASS, exit 0.
FACT: Final gate after lint fix 2/5 npm run format:check PASS, exit 0.
NEXT: Отримати typecheck result, перевірити lint з уточненим Symbol label, виконати final rstest.
FACT: Final gate after lint fix 3/5 npm run typecheck PASS, exit 0. 4/5 npm run lint запущено.
FACT: Final gate after lint fix 4/5 npm run lint PASS, exit 0; architecture boundaries checked 244 files. 5/5 npx rstest run src --testTimeout 30000 запущено, /tmp/ms4c-final-rstest.log.
FACT: Final review scope: 15 frontend files (4 нові helper/test файли), без backend/e2e mutations. Gate завершує весь src, включно наявні adapter/prediction tests; current phase report буде позначений run c only, independent review/e2e/design pending.
NEXT: Прочитати final rstest summary/exits; за відсутності failures зафіксувати точний scoped результат у report та завершити append-only журнал.
FACT: Final gate 5/5 npx rstest run src --testTimeout 30000 PASS, exit 0: 1415 passed, 113 files, 0 failed/skipped; 11.559 s. Default-timeout run не виконувався.
FACT: Поточний docs/multi-subtitles/report.md доповнено scoped MS4(c) результатом, command/exits, technical Symbol уточненням, unit/route proof та pending browser/e2e/design/review межами. Повна MS4 acceptance не заявляється.
NEXT: Перевірити фінальний diff/status і whitespace; передати користувачу стислий підсумок цього файлу без staging/commit.
FACT: git diff --check PASS (exit 0). Фінальний status: лише очікувані frontend files + phase report у scoped delta; сторонній стартовий dirty/untracked стан збережено. HEAD лишається 22973203; staging/commit/branch switch/reset/checkout/stash/clean не виконувалися. Делеговано 0 агентів, отримано 0.
FACT: Run c implementation і замовлений frontend gate завершено. Далі owner/run d: E2E M9/M10 та подальше незалежне review/design/final gate; жодних contract blockers або pending питань у цьому run.
