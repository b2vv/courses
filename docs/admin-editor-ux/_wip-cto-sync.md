NEXT: Перевірити правила входу, дозволені файли та розділи 2–5 навички cto-research; підтвердити репозиторій і базу freshness gate.
FACT: Прочитано shared agent-entry; поточний запит — окремий delta re-sync, без продовження старої задачі. Дозволено змінювати лише briefing та цей append-only файл.
NEXT: Зіставити локальні правила, cto-research §2–5 та поточний briefing з main@06c1c220.
FACT: Репозиторій <repo>, main@06c1c22084d7438bef9eee190f20c9bda59dd605 підтверджено. Прочитано cto-research §2–5 та AGENTS.local.md; локальні сторонні зміни залишаються недоторканими.
FACT: AGENTS.local.md фіксує Kiri з 2026-11-16 і прямо забороняє трактувати cutover як повний test pass; machine-local paths — лише pointers.
NEXT: Прочитати briefing частинами, agenda/task pointers та adapter; побудувати список 18 first-parent комітів і змінених первинних джерел.
FACT: Прочитано наявну структуру briefing і task/adapter routes. Кеш містить застарілі посилання у §7 (enrollment/ratings як майбутнє) та дубльовані історичні статуси ризиків; перевірятиму лише зачеплені фазами твердження.
NEXT: Підтвердити commit count і first-parent gate; звірити phase reports та чинну AGENDA без запуску gates.
FACT: git rev-list --count HEAD = 1005. Freshness gate 5fbaaff..06c1c220 має 18 first-parent комітів: 16 PR squash (#96–#111) та 2 docs-коміти (27f3f03a, cbb645a0). Додано міграції 00025–00031; змінені media/transcript/course і FE-модулі, Makefile/Compose.
DECISION: Шапка називатиме 18 first-parent комітів, а не 18 squash: git log підтверджує два окремі docs-коміти. Це уточнення класифікації, усі 18 входять у gate.
NEXT: Витягти виміряні gate-результати та named limitations із трьох phase reports; зіставити з чинною AGENDA.
FACT: AGENDA має owner-initiated чергу перевірки TanStack Start + Rsbuild (2026-11-14, після multi-subtitles), тоді як її старий footer каже «черга порожня»; датований конкретний запис є джерелом поточного pending, не дозволом його виконати.
DECISION: Виправити кешовані «останнє teach-editor», відсутність learning-overview/author-block і дубль факту №1: AGENTS.md Scope, docs/learning/report.md та commits #96–#111 підтверджують доставку й чинну OpenAPI-first нумерацію.
NEXT: Зібрати завершальні limitations S7/MS4/admin-editor-ux і відокремити їх від уже закритих проміжних S3/MS1 обмежень.
FACT: multi-audio §S7: gate 932 FE / 121 Go ok / 138 e2e; native без перемикача, 1258 мс до audible switch у Chromium. admin-editor-ux §2: фінальна спроба gate на 8f7923ad GREEN (1482 FE, 130 Go ok, 168 e2e); дві попередні спроби зупинили інфра-причини, вони не є PASS.
FACT: admin-editor-ux §5 зберігає D7 без Discard, raw-ms D6, відсутність duration у MediaVideo, 7 NOT_TESTED translucent contrast observations, слабшу unit-bound перевірку observer, R5 fixture leak і дисковий ліміт Kiri.
NEXT: Прочитати MS4 завершення та phase-plan рішення; перевірити code seams, міграції, OpenAPI генератор і dependency diff.
FACT: MS1 report зберігає історичні «MS2+MS3/MS4 не почато», але фінальні MS3d/MS4 записи й merged f50f6ec3/10436cdf їх закривають; 771/771 mutation constructs resolved — не 208 відкритих GAPs.
DECISION: У §6 не переносити як живі проміжні S3 uploaded-path defects і S7 single-transcript limit: вони мають closing commits #98/#100 та #109/#110. Успадковані runtime/proof limitations зберегти окремо.
NEXT: Взяти точні MS4 gate/limitations рядки, після чого перевірити поточний production path і dependency lockfile.
FACT: MS4 report фінальний gate 5b4569a3 (2026-11-18): GATE_EXIT=0, 1435 FE / 130 Go ok / 166 e2e, жодного відкритого MS4 row. Ліміти: 1.5 с без captions після cross-language replacement, native лише unit, orphan приховано, Ready без audio лишається loading; +5 px transcript overflow при 390/200%.
FACT: PR #109 report final: 771/771 mutation constructs resolved; S5 product search/sort race та B18 offline-test event виправлені, reaper-name retry розширений; gate 03d2440 GREEN, 1322 FE / 130 Go ok / 151 e2e.
NEXT: Перевірити diff lockfiles/Compose/Makefile; прочитати остаточні phase-plan limits та код OpenAPI/audio/transcript seams.
FACT: Прочитано multi-audio plan D1–D7 та multi-subtitles plan D1/§6, plan-ms4 §6. Нові фазові плани містять початкові snapshots; поточний стан визначають код і фінальні report записи.
FACT: Compose додає one-shot media-masters після migrate; media-worker чекає його успіху. Makefile e2e-reset резолвить імена томів із Compose config та шукає і за name, і за labels (cutover Kiri втратив labels); це код, не новий runtime прогін.
NEXT: Повторити stat із quoted glob (попередній glob backend/requirements* не розгорнувся), підтвердити dependency delta та точні lockfile lines.
FACT: git diff 5fbaaff..HEAD по frontend/package.json, package-lock.json, backend/go.mod/go.sum порожній: нових/оновлених залежностей у цій дельті немає. hls.js лишається 1.7.2 (lock:7512).
DECISION: Виправити вже застарілі версії у кеші §5: Router 1.170.45 (lock:4444), Rstest 0.12.4 (package.json:50), додати Start 1.169.3 / router-core 1.171.38 / zod 4.6.7; не називати їх bump-ами цього gate.
NEXT: Перевірити архітектурний факт №1 за генератором/Makefile та API-формами; вибірково звірити per-track persistence і межі модулів у коді.
FACT: Архітектурний факт №1 підтверджено backend/cmd/openapi-gen/main.go:28–37, internal/openapispec/contributions.go:19–21 та Makefile:177–184: Go contributions + overlay → OpenAPI → Orval; check-generated порівнює робоче дерево з HEAD (git diff), не staged index.
DECISION: Виправити кешований «check-generated порівнює з індексом» (Makefile:182) та дев’ять модулів/п’ять executable у §3: collector явно містить 10 модулів, cmd/openapi-gen є шостим entry point.
NEXT: Звірити всі 00025–00031 migration mechanisms та per-track draft task identity; перевірити frontend audible-buffer state і admin UX implementation anchors.
FACT: Прочитано 00025/26/27/29/31 та preflight/backfill 00028/30: audio_tracks, awaiting_audio, upload metadata, source_leaf (історично), перехід PK drafts → audio_track_id, case-insensitive language index та legacy relabel.
DECISION: Виправлення власного попереднього запису: git diff без --cached/commit порівнює working tree саме з index, тому стара примітка «check-generated порівнює з індексом» правильна. Її не змінювати на HEAD; Makefile:182 не задає commit.
NEXT: Завершити звірку learning DTO/audio-buffer/frontend authoring; зафіксувати джерела для поточних gates і решти dependency rows перед редагуванням.
FACT: Код learning DTO містить audioTracks {language,isDefault}; course SQL сам читає audio_tracks (ADR013), без імпорту media. transcript queue taskID = asset:generation:track:marker; 00031 PK per-track. Whisper 1.2.1 отримує trackLanguage/multilingual, und → autodetect.
FACT: hls-media.ts використовує nextLevel/nextAudioTrack, audio SourceBuffer coverage для confirmed/pending; learn route відкидає чужий mount/generation. Admin editor зберігає mounted panels/lesson bodies, приховує inactive, компактний transcript використовує ResizeObserver.
NEXT: Зібрати точні локальні runtime/CI/style anchors, перевірити relevant dirty state і написати лише зачеплені секції briefing.
FACT: AGENTS.md:5/8 фіксує roadmap complete, production deferred, delivered learning/admin/teach та admin-only revoke. Нові фази підтверджені merge hashes; contract-файл уже має pointer на briefing (AGENTS.md:4).
FACT: У delta немає змін .github/workflows або lockfiles; CI-select лишається verify per PR/main + tag/manual e2e без ASR. Dirty state має adapter/package cache зміни; header відзначатиме local-only AGENTS.local Kiri pointer, не вдаватиме чистий checkout.
NEXT: Зіставити успадковані ризики з backlog/playback/transcript і доступом, щоб не закрити чужі обмеження без evidence; потім внести delta briefing.
FACT: Додаткові успадковані ризики кешу вже закриті до бази: backlog B-18/B-19/B-21 має CLOSED 2026-10-20; B-1 observability — 0b69d04. Їх треба прибрати з живого переліку, зберігши closing evidence.
FACT: multi-audio §S-backfill D-B6 не відновлює embedded оригінал для вже Ready uploaded-ассета (БД не відрізняє його від video-only); §S5-L2b приховує audio UI під queued/processing та failed reason не входить у DTO. Це named limitations, не закриті агрегатним gate.
NEXT: Уточнити closing hashes старих backlog entries і lock/runtime lines; у §4 явно розрізнити workflow configuration, історичний локальний PASS та неперевірений поточний CI/runtime.
DECISION: Список живих ризиків прибирає вже закриті до бази B-18/B-19/B-21/B-22/B-1 з посиланнями на backlog outcomes/commits; закриття B-21 primary не плутати з новою MS4 exact-language policy. Single-locale документа уроку лишається, multi-subtitles стосується transcript, не lesson_documents.locale.
FACT: Реальний multilingual e2e існує: frontend/e2e/asr.spec.ts M3-R14 mint multilingual=true → ready draft; фінальний PR109 gate охопив кожний asr.spec row. Раннє MS2b «real model not verified» не переноситься як чинний висновок.
NEXT: Внести briefing (шапка + §1–§6; у §7 лише виправлення зачеплених phase pointers), зберегти попередні gate rows і документувати всі зміни як cache corrections.
FACT: Closing hashes локалізовані: B-18 d6ec304d; B-19 801391e6; B-21 3b08f365; B-22 прийнята limitation def0db3e; B-1 0b69d04. Lock містить Storybook 10.6.3 (range ^10.6.1), Rstest 0.12.4.
DECISION: §1 буде актуальним коротким verdict замість набору старих phase snapshots; §2/§3 збережуть доменні межі та отримають multi-audio/MS4/UX rows. §6 зберігає phase limitations і нез’ясовані механізми, але не дублює давно закриті ризики.
NEXT: Записати цільовий delta update, потім виконати read-only перевірку diff, cited paths/lines, структури та кількості unverified; жодних builds/tests/Docker.
FACT: Оновлено лише work/CTO-RESEARCH.md: новий gate 2026-11-18/main@06c1c220/1005, §1 current verdict і OpenAPI fact #1; продукт/межі audio+subtitles+UX; infra/migrations/Kiri/recorded gates; dependency table без вигаданого bump; named limitations та closing hashes; §7 актуальні pointers.
DECISION: Попередні header gate rows збережено. У новому тексті 3 явно позначені непідтверджені тези (R23, identity failure, NAT); поточний runtime/CI SUCCESS не заявляється.
NEXT: Read-only QA документа: перевірити diff, відсутність випадкових замін, існування cited paths/lines, незмінність історичних header rows та дозволений write scope.
FACT: Запис завершився без помилок; це редагування документа, не виконання продуктового коду чи gates.
NEXT: Прочитати QA результати й виправити тільки помилки посилань/тексту, якщо знайдуться; підсумувати обсяг і недоведені факти в кінці цього файлу.
FACT: Документ QA: 7 секцій у потрібному порядку, 3 unverified labels із required wording, усі 5 історичних freshness rows збережені дослівно, cited line numbers не виходять за файли. Три ADR shorthand pointers потребують повних filename; це номерні references, джерела локалізуються у docs/adr.
NEXT: Уточнити ADR filenames, прибрати порожні рядки всередині двох доповнених таблиць і застарілі cross-reference «§1 факт №9»; звірити залишені історичні source pointers.
FACT: QA виправлення: повні ADR 006/012/013 filenames знайдено; прибрано розриви таблиць і мертве «§1 факт №9». Оновлено payment-era «без змін» product snapshot на виміряні phase pointers.
DECISION: Старий опис grant UNIQUE лишити лише історією 00011 та явно вказати 00015 user/course/state (джерело migration 00015 і enrollment report); прибрати вже неактуальне «перед спекою рейтингів». Сторонні джерела не редагуються.
NEXT: Перевірити повні source filenames і rendered Markdown table continuity; виконати остаточний read-only scope/label/history check.
FACT: Остаточний QA виправив ADR full paths та старий ratings-next pointer. git diff --check для briefing пройшов; структура 1–7, історія gate rows і 3 unverified labels перевіряються read-only, без продуктового test/build/runtime.
NEXT: Переглянути фінальні результати, дописати підсумок цього файлу (count, touched sections, невстановлене) та завершити без staging/commit.
FACT: QA фінал: 0 відсутніх explicit cited paths, 0 out-of-range lines, 0 розірваних таблиць; усі 7 секцій та 5 попередніх gate rows на місці. Staged diff порожній; tracked dirty список дорівнює початковим 6 стороннім файлам + briefing.
NEXT: Замінити два посилання на попередній cache прямими первинними pointers (identity integration/log, security commits), дописати остаточний count/межі перевірки.
FACT: Identity primary anchor підтверджено administration_integration_test.go:45–48 (власний postgrescontainer через containers.Start); integration.sh:245–272 зберігає integration.log і справжній exit. Для цих тверджень старий briefing не потрібен.
NEXT: Дописати прямі primary citations і завершальний запис; count unverified залишиться 3, відкритих source-conflict питань немає.
FACT: Завершено delta re-sync на main@06c1c220: 18 first-parent commits від 5fbaaff, 1005 total. Primary citations уточнено; SA13 TOTP closure = c22c9d41 (#90), identity механізм лишається невстановленим. Read-only QA фінального документа пройшов.
FACT: Змінено лише два дозволені файли: work/CTO-RESEARCH.md і цей append-only record. Сторонні WIP збережено; staging порожній; commit/branch switch/builds/tests/Docker не виконувались; agents launched/returned = 0/0.
DECISION: Re-sync завершено; новий продуктовий/infra scope не розпочинався. Відкритого source-conflict QUESTION немає.
FACT: Підсумкова кількість ⚠️ unverified у briefing: 3 — R23 heartbeat 403, identity integration failure, limit_req media під NAT; усі мають wording «it appears that…» та джерело/межу доказу.
FACT: Зачеплені секції: шапка (2026-11-18/base/count/new gate/history), §1 verdict і architectural fact #1, §2 product, §3 seams, §4 infra/migrations/Kiri/recorded gates, §5 dependencies, §6 risks/limitations/closures, §7 лише актуальні reading pointers.
FACT: Не вдалося підтвердити новим вимірюванням (поза дозволеним scope): поточний live runtime health, CI SUCCESS на HEAD, machine-local Kiri files/ємність, реальний NAT load та причини двох старих флейків. Залишені named phase limits: native Safari live-proof, 1.5 s caption gap, per-track ASR cost, migration preflight/backfill residuals, F3/MS4 visual debt і admin-editor-ux §5. Gates описані тільки як вимірювання phase reports, не як виконані в цьому re-sync.
