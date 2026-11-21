NEXT: Прочитати agent-entry, prompt plan-critic і skill plan-critique, визначити обов’язкові входи та межі перевірки.
FACT: agent-entry прочитано; поточне завдання явно задане, відновлення старих задач не потрібне. Зміни дозволені тільки у цьому файлі.
NEXT: Перевірити корінь, локальні правила, briefing, prompt і чотири лінзи.
FACT: Корінь підтверджено: <repo>; AGENTS.local.md, prompt v1 і канонічний skill прочитано. Skill вимагає 4 незалежні clean-context лінзи; окремі progress-файли будуть поза репозиторієм у /private/tmp.
FACT: Вивід briefing був обрізаний; пропущені частини будуть перечитані обмеженими порціями.
NEXT: Прочитати правила делегації й план, перевірити SHA та snapshot без зміни git-стану.
FACT: SHA-256 prompt і plan.md збігаються з наданими; branch admin-editor-ux, HEAD і origin/main = 10436cdf2988a17887c959dd48b363e6b117e9d7.
FACT: Правила stream-to-file, multi-model-roles і reference-anti-patterns прочитано. Author model невідомий; reviewer — поточний Codex, окремі fork_turns=none контексти, без зовнішніх/платних API.
NEXT: Запустити 3 незалежні лінзи, потім четверту при звільненні слота; паралельно перевірити первинний аудит, код та конкретні e2e селектори R1.
FACT: Лінза AI-slop запущена у clean context; прогрес /private/tmp/courses-admin-editor-ux-ai-slop-20261118.md. Запущено 1 / повернулося 0.
FACT: Первинний аудит B1/M1–M6/m1–m4 прочитано повністю. Великий batch із work/tasks.md обрізав вивід; наступні читання будуть bounded.
NEXT: Запустити logical-gaps та complexity; продовжити bounded читання правил/коду.
FACT: Лінза logical-gaps запущена у clean context; прогрес /private/tmp/courses-admin-editor-ux-logical-gaps-20261118.md. Запущено 2 / повернулося 0.
NEXT: Запустити complexity у третьому доступному worker-слоті.
FACT: Лінза complexity запущена у clean context; прогрес /private/tmp/courses-admin-editor-ux-complexity-20261118.md. Запущено 3 / повернулося 0; edge-cases очікує звільнення слота.
NEXT: Прочитати briefing обмежено, ідентифікувати всі e2e драйвери редактора та зображення аудиту.
FACT: Пошук знайшов редакторні UI-входи у admin, media, asr, playback, transcript, audio-tracks, discussion, subtitles, enrollment, purchase; lesson-builder відкриває окремий document editor, решта багатьох збігів — API fixture writes, не UI.
NEXT: Прочитати CourseEditorView та скласти точну R1 матрицю з селекторами і передумовами видимості.
FACT: CourseEditorView formValues копіює тільки дерево курсу; FormErrorSummary делегує reveal/focus через onSelect (рядки 165–210). Immediate-apply редактори не є частиною цього reset-контракту.
NEXT: Перевірити keys/lesson identity, мутації та приховування вкладок у решті CourseEditorView.
FACT: LessonFields uses stable formKey rows and savedLessonIds; default tab overview (CourseEditorView:546). Save/reset and course.updatedAt updates reset field arrays (570–578,600–606); focusField switches only course tab (640–646).
NEXT: Звірити DOM-структуру табів, save bar, вкладені секції та transcript state.
FACT: Overview/Lessons TabPanel вкладені у form (724/852), Access є прямим child Tabs (885). Варіант D1 `.course-tabs > .ui-tab-panel[inert]` не приховає два панелі; dedicated class/descendant course-scoped selector — працездатний варіант.
NEXT: Прочитати transcript, media contract і точні e2e переходи/локатори.
FACT: AI-slop повернув explicit CLEAN (0 findings), перевірив user outcomes, D5 multi-consumer header, D6 existing timing presentation. Запущено 3 / повернулося 1.
FACT: R1 вже конкретний: media.spec має багато saved lessons і після Lessons tab не розгортає їх; purchase.spec має 2 lessons та перевірку preview checkbox. Hidden native file-input setInputFiles саме по собі не доводить видимість/поломку; видимі downstream controls/assertions треба розділити.
NEXT: Запустити четверту лінзу edge-cases та продовжити перевірку R1.
FACT: Edge-cases запущено clean context (/private/tmp/courses-admin-editor-ux-edge-cases-20261118.md); загалом 4 launched / 1 returned.
NEXT: Зафіксувати вимоги часу/precision і видимості, перевірити решту редакторних e2e.
FACT: asr.spec перед відкриттям Lessons натискає вкладку (110), single-lesson fixture; audio-tracks.spec має два lesson rows (263/268); subtitles.spec має три (580–605). Discussion helper не перемикає вкладку (80–89), працює через hidden file input і API poll — це прихована UI-взаємодія, але не гарантований Playwright visibility failure.
NEXT: Переглянути screenshots первинного аудиту та transcript/application.
FACT: Переглянуто audit screenshots audio/ASR, transcript-actions, bottom-save-access; вони підтверджують raw-ms fields, плоскі секції та рознесені commit points на старому snapshot, не є live-перевіркою 10436cdf.
NEXT: Звірити збереження raw draft strings і поточні route/contract/style межі.
FACT: TranscriptDraftLine зберігає startMs/endMs як рядки (recovery.ts:22–27); restore копіює offeredSnapshot.lines напряму (TranscriptEditor:359); add-line робить Number(endMs)+1000 (519–520). D6 має сумістити display format зі збереженим raw-ms/recovery state, а не лише API round-trip.
NEXT: Перевірити schema MediaVideo, API mapping, root NotFound і library toggle surface.
FACT: Complexity повернув explicit CLEAN (0 blocking/0 non-blocking): existing Dialog/reset/IDs/routes/visibility sufficient, time parser has no reusable equivalent. Загалом 4 launched / 2 returned.
NEXT: Уточнити R1 повністю та контракт MediaVideo; перевірити routing і scroller CSS без запусків.
FACT: MediaVideo має required originalName, без duration/thumbnail (openapi.yaml:2381–2435); filename already in LessonVideo domain. m4 можна закрити тільки filename без backend змін. Document editor route non-nested та lazy (admin.courses.$courseId_.lessons.$lessonId:8–16,32–46).
NEXT: Уточнити transcript/editor e2e default-panel дії та semantic role checks після collapse.
FACT: Усі 10 audit screenshots переглянуті. Transcript spec відкриває Lessons явно (288,557,576,595); admin lifecycle spec також переключає Lessons→Overview; жоден із цих single-lesson шляхів не потребує default expand зміни.
FACT: R1 exact false-pass приклад: discussion.spec:86–89 робить page.locator(file).first().setInputFiles без Lessons, потім API poll — full e2e може лишитися green без реальної видимої user path.
NEXT: Дочитати briefing та локальний workflow у bounded chunks, перевірити snapshot dirty scope.
FACT: git status read-only підтвердив сторонній dirty/untracked inventory; додатково backend/upload і multi-subtitles/teach-editor WIP теж поза scope. Жоден такий файл не змінювався.
FACT: Локальний adapter забороняє трактувати make check як безпечний read-only крок (reset dev volumes); план має consent per run, поточний critique не запускає жодних gates.
NEXT: Дочитати стислий архітектурний briefing і standards; перевірити точні R1 fixture counts і path scope.
FACT: main.css уже приховує language panels через inert (3629), але глобальне правило для course panels відсутнє. Collapsed hidden wrapper має врахувати author display:grid; перевірка computed display необхідна (admin-lesson:1222).
NEXT: Перевірити frontend boundaries, завершити briefing/recovery adapter читання й кількість уроків у R1 fixtures.
FACT: Logical-gaps повернув 1 blocker (D6/recovery format contract) та 2 non-blocking (D4 language-tab/track count, D1 child selector). Загалом 4 launched / 3 returned.
NEXT: Прочитати bounded результати лінз та перевірити edge-case кандидати; уточнити B1/R1 findings незалежно від green gate.
FACT: Bounded agent logs прочитано (без JSONL). LG1 підтверджено independently: DB v1/no format discriminator + string times; existing incomplete raw-ms capture contract. LG2/LG3 лишаються non-blocking, бо дані/правильна альтернатива вже існують.
NEXT: Завершити R1 selector inventory й відрізнити hard failure від green-through-hidden file input; перевірити exact-time backend rules.
FACT: R1 fixtures: media=5 lessons (media.spec:31–35/315), playback=4 across 2 modules (playback.spec:56–85), audio-tracks=2 (75–78), subtitles=3 (580–585). В усіх default collapse D4 застосовується; головний visibility blocker у media — button «Завантажити відео» (182), в audio-tracks/subtitles — [data-video-state="ready"].toBeVisible (103/605).
NEXT: Перевірити scope hashes і дочитати результати edge-cases; нічого не редагувати поза working file.
FACT: Product source diff --name-only для перевірених src/routes/styles/e2e/OpenAPI порожній; ці inputs відповідають HEAD. Managed kit release = 0.8.0-alpha.1+native-handoff.1; AGENTS.local.md supersedes historical adapter cache pin.
DECISION: R1 буде окремим numbered finding: повний e2e не доводить відсутність hidden interactions, бо file-input + API poll обходять visibility. Потрібен explicit reveal/assertion у affected helper paths, без force/includeHidden чи API заміни user journey.
NEXT: Зафіксувати fingerprint scoped evidence і остаточну R1 матрицю у working file; дочекатися четвертого висновку.
FACT: Scoped evidence SHA-256 inventory (all e2e spec files included to bind R1 search coverage):

```text
136e3b3cd372a9c11fe9142fc1d95002189a62e594c7c60b2ce2131192211085  AGENTS.md
ffc2a416ba063367bc86476b3fdcfb537c9b2b35ed107516e3a6c413b2af7311  AGENTS.local.md
a16c3508897ab7d04a17d0ee9ce1e11e466ad472d4f6df8f2b25429cab560d85  .claude/standards.md
dff75cbcd507a1ff0daaf0ec2506b0dd91448bea257d1634badf35118b24b47e  work/CTO-RESEARCH.md
f6cccc35f4baecf294e18bd38f66360e21d9f399fff6495c6e7a103aa5063ed1  docs/ux-review/admin-course-editor-2026-11-15/report.md
4e11966249f1c9339ba2e9580fdaa666234a4f8511c9ca0aeac28ee0115634e1  frontend/src/modules/admin/ui/CourseEditorView.tsx
5751be5c1083b53ea3484f30cf2e83edb06982181838098d7ad696665ee07460  frontend/src/modules/admin/ui/LessonLanguageTabs.tsx
7ec12d2a3f74cf57d70d324d979dbe79af2b3df72e5dc63ecb91696f6d13f112  frontend/src/modules/media/ui/LessonVideoField.tsx
1498b2eee1604b14fe8b5ead1f09620acfec8095277cf4cae2f1d4bece1764b2  frontend/src/modules/transcript/ui/TranscriptEditor.tsx
d51c1e993041dcc5886c6e3ea8d347646fc0544b741d7d74583245c9e63efcf8  frontend/src/modules/admin/ui/DocumentLibrary.tsx
d3138657cfcac74b89054b41b3b097ca4c0289be3943f8c808245ddc8bd5de69  frontend/src/routes/__root.tsx
7cd33bd7754ed4e0d8616e27ac2281fe1fcbbfb995c2da59c3003da2ca24fdd4  frontend/src/routes/admin.courses.$courseId.tsx
a1ec74c19f2d4328edc035cacc31651ae7da0e9588e3bb4be54dcebc790913ae  frontend/src/routes/admin.courses.$courseId_.lessons.$lessonId.tsx
a0daf70f59d9b1a3c6a6637ebaae8ff6bc9378ea851addf36c62bf521728bdf4  frontend/src/app/styles/main.css
c11124a673cce9cfe7db7525c31854692749380c0336d880fc4e25002bd002c3  frontend/src/app/styles/tokens.css
eb94e683bc15ab7e0d67a70f4454f10a2ff30704a051cf457f27f92c11475bd3  backend/openapi/openapi.yaml
0336c9681a2de1275751330ee65fcbdb5b83f13d0658954321eacc423aa90939  frontend/src/modules/transcript/application/panel-state.ts
5c723693b535718033b111502676ba85bf9d2fc5c62b05048e1fd459c84b9e69  frontend/src/modules/transcript/application/recovery.ts
d8d6ac9dbef879a49ccb632a08f1e3d5eb640db87ba2b145dffbda45675f6b36  frontend/e2e/accessibility.spec.ts
0aeba028a7c5a6eaaf7f5069ce2e8ae3182c64a4a99c1bee6517ecc993db0154  frontend/e2e/accounts.spec.ts
6e7db10690477d983aa9e5877d5845eacccc30e8a6dba1ffd815a7891c3a3152  frontend/e2e/admin-analytics.spec.ts
b6487ac30d9261ec131605735e10668b153cc572637f8ca6b7d93ea3069a73f4  frontend/e2e/admin-users.spec.ts
36fd1ce4d33fb065a665acad8ad0cab7c016edcdd89298b8ff89942642a66f33  frontend/e2e/admin.spec.ts
f74e41d53c60e706e7b9ebaabc617f1fc65b6de78a4192c79a7f37ee94798a13  frontend/e2e/asr.spec.ts
3de33ad08c2820031dd4fb74ab29c0acdbba6360d92b6a4a4a5e54b68635b839  frontend/e2e/audio-tracks.spec.ts
10921a7298eb43eb08200e1eee5125e3409f93f5bd0011e6d8cfee70f213f0d3  frontend/e2e/catalog.spec.ts
2d2e4c35675ee80c954144b68110d673db0066e8123982aca43582f7d731ad53  frontend/e2e/discussion.spec.ts
f38c26e3e81c7b8080f052eacbca0a65809ec35e51e4a1c8d97914fb048604ba  frontend/e2e/enrollment.spec.ts
ef483b1757858f4436774ef3b5e49b58a773695e961ec5564fda6749c1762a69  frontend/e2e/instructor.spec.ts
dc502441c4fbd702de09171497031c250b6c6d95cbe992c72a77d8d706e19ce1  frontend/e2e/learn-access.spec.ts
2a7984785968fdaf86ebb051d85f7a58709f6925c02e2ea6a100912d4aeb50c9  frontend/e2e/learning.spec.ts
7ae2f2e82877855a4b2dda43b1f9abe2b0021aec919c7d0794d7629b257c3baa  frontend/e2e/lesson-builder.spec.ts
a1453e5ddbea98dc135ad7bd2abe497ec127e5b5ca899e3ea231777aca3a14f6  frontend/e2e/media.spec.ts
904f5a9cb29bc3c1086d6a6b959b6f37096d791886975e72cf5fafa9ef6deb0d  frontend/e2e/observability.spec.ts
551ca35243860d49be0d49614414bedaf4a5e3cccfb3b6f9a562c31557794bba  frontend/e2e/playback.spec.ts
ac8ead38fc59050ceed47336b3ca8e373bad02ba363900d11f8226556c23bffd  frontend/e2e/purchase-wayforpay.spec.ts
d9495d9416173ed213d309d62ba64e8e3aaceeabca1ec36a005674ff7c26c540  frontend/e2e/purchase.spec.ts
cfcd844c746cc7f4c1ccc1a293decca23f60874e5785b8728b268915f4bdc104  frontend/e2e/ratings.spec.ts
e204eeb3c56a345aee0f334d565746001831b7366e4148e363d505a3cf0176ce  frontend/e2e/recovery.spec.ts
a48b1f44ef3fed27c46a5619128efa9d98cbd561ac125dbd5e25009aee0b609e  frontend/e2e/refund.spec.ts
d7b932c340267e5f0dd18f85361ad6c92352e7bf42ab10c839165cc1d88e02f8  frontend/e2e/route-guards.spec.ts
5b9f89620dc50c36b4a62824cdb3d09a231f31b5817fa6e7ce55d4f3ef6b291d  frontend/e2e/shell.spec.ts
4abe7b4a170cb39d80307b45e34e3275a5e93bcb89f503fd5f98fadcf6a557c0  frontend/e2e/subtitles.spec.ts
4c341b0a1ec727203cd2ec690da8a93fd4bd9e0b1b18c0879e705eec01b67cf6  frontend/e2e/teach-editor.spec.ts
95debc6e0e4c50eabb1fb92498804a69ee054770a3759e3bc19f2a56046bfca8  frontend/e2e/transcript.spec.ts
1c129d4110af84687a28bf8bb3df22ddf9209f93e1ac253e7bef65efc4cc45cd  docs/ux-review/admin-course-editor-2026-11-15/1024-lessons-tab-inert-details.jpg
0f580c093cfa13d7de77090a0ed248fd757319ad9064d0e3fa82ac58db989594  docs/ux-review/admin-course-editor-2026-11-15/1440-admin-404.jpg
bdd114bc720cd109077d8bb1267b73e9bc33935f9d8aa3458eac9fa185bb0f6a  docs/ux-review/admin-course-editor-2026-11-15/1440-lessons-audio-asr.jpg
dfc6b67e1a67529ba8ef3ac88f9f1c8f676f5067e4b7d0b9ca4fe73ad4042cbf  docs/ux-review/admin-course-editor-2026-11-15/1440-lessons-bottom-save-access.jpg
ce950f6653296db7d652def446bb73a938b41d95542282686b3554c344dc25a6  docs/ux-review/admin-course-editor-2026-11-15/1440-lessons-materials-syllabus.jpg
e27c8a918a0483765c42329a5e9dc60a1c63dde214d6f57b52d2940ee0b9cbfd  docs/ux-review/admin-course-editor-2026-11-15/1440-lessons-tab-top.jpg
48d6343515add3bf6fb73d77279c0310223d453e22f5b6019ffdf20aecaa2071  docs/ux-review/admin-course-editor-2026-11-15/1440-lessons-transcript-actions.jpg
173edeb4d8a11d42f8a5df42b2dc38038ed60081179f832a025b40eeb6bb833e  docs/ux-review/admin-course-editor-2026-11-15/1440-overview-tab-shows-lessons.jpg
c59387e48c8fb865359dd5a8254e5e4d35122efab748e991d765d373bf2a883f  docs/ux-review/admin-course-editor-2026-11-15/390-lessons-tab-inert-details.jpg
b4cbf0fec89dae95f87844e5c59dd00a387cedc5cd56825b500f079c9b1fbfe9  docs/ux-review/admin-course-editor-2026-11-15/768-lessons-tab-inert-details.jpg
```
FACT: Scoped fingerprint inventory додано: 55 files, inventory SHA256 40aff4aeae8986577d22e61ceaa3c3d8b0f0958ac95c97d0036684840de0410c. Current managed manifest SHA256 213e0aaa8983e33f59430821912e155981fd8433c3afaa00303468fe3be905cc (фактичний прочитаний release; не прирівнюється до історичного receipt hash у AGENTS.local).
FACT: Edge-cases повідомив завершення: D6 recovery та D7 reset/remount — blockers, disappearing reorder focus — non-blocking. Загалом 4 launched / 4 returned; потрібно звірити механізм D7 перед остаточним disposition.
NEXT: Прочитати edge-case findings bounded, перевірити RHF field-array reset механізм і додати остаточний verdict.
FACT: 4 launched / 4 final verdicts returned, no retry/missing lens. D6 corroborated by 2 lenses; D7 by edge-cases, verified against RHF generated IDs, component keys and route-owned async transfer.
NEXT: Перевірити R1 false-pass на installed Playwright source, завершити numbered findings з disposition і scoped verdict.
FACT: RHF reset broadcast regenerates module/lesson formKey, proving D7 remount mechanism. Plan SHA remains unchanged. Optional installed Playwright dom.js path absent; це не missing required input і не підстава STOP; source discovery needed before claim of library implementation.
NEXT: Знайти фактичний installed Playwright artifact; після цього записати verdict та матрицю.
FACT: Playwright встановлений як lib/coreBundle.js; попередній legacy path не застосовний.
NEXT: Підтвердити setInputFiles checks у bundle для коректного R1 severity.
FACT: Installed Playwright _setInputFiles implementation located at coreBundle.js:20133, selector entrypoint :23705.
NEXT: Прочитати реалізацію та остаточно класифікувати R1; це read-only evidence, не запуск Playwright.
FACT: Installed Playwright coreBundle.js:20133–20168 підтверджує: _setInputFiles перевіряє INPUT/multiple/directory/connected та задає файли без visible/enabled checks. Отже hidden file-input + API poll може бути green; це source-backed mechanism, не runtime-claim.
DECISION: Нижче — остаточна консолідація initial r1 critique. Plan не змінювався. Required inputs доступні; QUESTION/STOP через missing input не потрібні.

FACT: stage: plan. Critic/coordinator: Codex /root; independent clean-context reviewers: /root/ux_ai_slop, /root/ux_logical_gaps, /root/ux_complexity, /root/ux_edge_cases. Author: Courses 1; author model unknown. Reviewer binding: inherited current Codex model, no override or external API. Route: four new fork_turns=none contexts required by the explicitly selected canonical skill; delivery owner /root. Shared runtime slot: none needed, static reads only. 4 launched / 4 final verdicts returned; 0 missing, 0 retries.
FACT: Exact artifact: docs/admin-editor-ux/plan.md r1 (2026-11-18), SHA256 62ed935e8e6070c690964db2817ef7807da276004b21197c968c44a99c62555b. Root <repo>; branch admin-editor-ux; HEAD = origin/main = 10436cdf2988a17887c959dd48b363e6b117e9d7. Relevant product-source diff empty; plan/raw audit are untracked and bound by hashes above. Unrelated dirty/untracked inventory excluded, including backend/upload and multi-subtitles/teach-editor WIP.
FACT: Prompt identity: <kit>/prompts/plan-critic.md, version 1, SHA256 514e4c376c2551afe76e390c5779867ecf7f04ad2a552c97c2b4f713c43dcab6. Managed release <kit>, manifest version 0.8.0-alpha.1+native-handoff.1, actual manifest SHA256 213e0aaa8983e33f59430821912e155981fd8433c3afaa00303468fe3be905cc. Skill: <kit>/skills/plan/plan-critique/SKILL.md; stream-to-file, model independence and reference anti-patterns applied. Historical adapter pins were not substituted for the current managed route.

DECISION: VERDICT — BLOCKED for the complete r1 plan. Open findings: 3 blocking / 3 non-blocking. No finding was silently waived, no plan edits made within this read-only review. Findings 1–2 concern product-state contracts; finding 3 concerns the claimed verification discriminator. Independent slices are not invalidated beyond those dependencies.

DECISION: 1. [logical gaps + edge cases; 2 lenses] BLOCKING / actionable / OPEN — Define time-edit storage and recovery compatibility before U3.
Where: docs/admin-editor-ux/plan.md:52–56, :76–77, :92.
Evidence: frontend/src/modules/transcript/application/recovery.ts:14–27 deliberately retains arbitrary raw-ms text, including unfinished input. frontend/src/modules/transcript/adapters/transcript-history-indexed-database.ts:5–25 has DB v1 and string startMs/endMs with no time-format discriminator. TranscriptEditor.tsx:78–84 creates raw-ms baseline strings; :359 restores snapshot.lines directly; :519–520 adds a line using Number(endMs)+1000.
Impact: Replacing draft strings with mm:ss.s breaks old local drafts such as "1000", and eager normalization loses unfinished edits. Parse/format round-trip plus untouched SERVER times (Ux08) do not prove restored exact ms or incomplete local work. A fresh implementer must invent a persisted-state contract.
Plan edit: Extend D6/U3 to select the representation explicitly. Retain raw-ms history with a separate display-edit buffer where possible; if history stores formatted text, define an unambiguous backward-compatible representation/reader that also preserves partial input. Define per-field untouched precision, ASR load and Add a line behavior without changing optimistic-lock/source ownership. Add named proofs: old raw-ms snapshot restore→save, new partial formatted time capture→reload→restore, untouched sub-100ms times after restore, and Add a line after restored/edited time. Keep backend-valid equal endpoints/overlap valid (backend/internal/transcript/domain/transcript.go:89–98). A whole recovery rewrite is not required.
Verification needed: assertions discriminating those states, not only parse(format(ms)); existing TranscriptEditor.test.tsx:250–260 already treats half-typed raw startMs "15" as recoverable user data. Static review only; proposed proofs NOT_RUN.

DECISION: 2. [edge cases; 1 lens, coordinator verified] BLOCKING / actionable / OPEN — Discard must not reset independent in-flight editors.
Where: docs/admin-editor-ux/plan.md:57–60, :80–81, :94.
Evidence: CourseEditorView.tsx:334–347 and :449–462 key subtrees by generated formKey. Installed react-hook-form/dist/index.esm.mjs:3356 broadcasts reset array values; :1188–1194 regenerates keys. LessonVideoField.tsx:149–162 owns progress/busy state, :179–188 passes local setters into the async transfer, :206–209 treats missing local progress as interrupted. The transfer owner survives in routes/admin.courses.$courseId.tsx:147–157. TranscriptEditor.tsx:106 owns unsaved lines independently of the course form.
Impact: Start upload → edit course title → confirm new Discard: prescribed form.reset remounts surviving lesson editors while upload continues. Progress/busy disappears and Resume/Cancel may be offered for the active transfer. Independently edited transcript state also remounts (recovery may offer it back, which is not preserving the current screen). Discard during pending course Save has no defined result and can be overwritten by the outstanding Save.
Plan edit: Define Discard as resetting only pending course-tree changes while preserving independent editor identity/state for surviving lesson IDs, or explicitly guard it during operations whose state would be lost. Disable Discard during pending course Save. Explain scope in its confirmation. Add deferred upload + dirty title + Discard proof (progress continues, no false Resume/duplicate upload), dirty transcript retained, and course Save success/failure while Discard unavailable. Removed unsaved rows may disappear normally. Existing Save/reset behavior is pre-existing; this finding concerns the new Discard promise and its prescribed reset mechanism.
Verification needed: FE deferred-operation assertions plus a real editor E2E covering the resulting state; full E2E remains required for shared/persisted effects. NOT_RUN.

DECISION: 3. [coordinator, R1/test-strategy check] BLOCKING / actionable / OPEN — Replace “full e2e proves no hidden interaction” with explicit reveal assertions and a concrete helper migration.
Where: docs/admin-editor-ux/plan.md:35–36, :47–49, :98–100, :107–109.
Evidence: discussion.spec.ts:80–101 uploads from default Overview through a CSS-selected file input and polls API only. playback.spec.ts:286–325 switches Lessons but never expands any of its four rows; it also uploads via CSS file input and polls API. Installed Playwright coreBundle.js:20133–20168 has no visibility guard for setInputFiles. Conversely, media.spec.ts:182 explicitly requires a visible Upload button and audio-tracks.spec.ts:103 / subtitles.spec.ts:605 require a visible Ready indicator: those multi-lesson paths will require reveal steps. See full R1 inventory below.
Impact: R1 is partly a deterministic failing-test migration and partly a false-green evidence gap. Full E2E by itself cannot prove the D1 claim that no spec interacted with a hidden panel. A test can upload bytes and become green without a usable, visible editor. Per-slice §6 says affected E2E, while R1 requires FULL E2E for visibility slices; a fresh implementer needs one unambiguous cadence.
Plan edit: Add concrete U1/U2 work to switch the correct course tab and expand the target lesson through its real control before UI operations, scoped by module/lesson identity. Assert the visible Upload button/target section before setInputFiles; preserve required API checks afterward. Retain all current scenario assertions, with no force/includeHidden or API replacement for a UI journey. Use the inventory below as the migration list; re-establish reveal after reload/reset where applicable. State full E2E for every visibility-changing slice consistently in §6 (owner-authorized Kiri run); affected E2E may supplement it. Bind Ux01–02 evidence to actual tab/collapse states, not aggregate green.
Verification needed: static helper changes reviewed plus full E2E and explicit selected-panel/expanded-row assertions on the final implementation. None executed here. Blocks U1/U2 verification strategy, not authorization to repair it.

DECISION: 4. [logical gaps; 1 lens] NON-BLOCKING / actionable / OPEN — Use one count definition in D4 and Ux04.
Where: docs/admin-editor-ux/plan.md:44–45 versus :73.
Evidence: LessonLanguageTabs.tsx:34–61 projects ready audio tracks PLUS transcript-only languages; :73–100 has a no-ready-video fallback. One track plus one orphan-language transcript gives two tabs, one track.
Impact: D4 says language-tab count while Ux04 says track count; implementing one can fail the other or mislabel orphan transcripts as audio.
Plan edit/check: Choose and align the label/source; reuse the existing tab projection if counting language tabs, define fallback/loading/error display without fabricating zero. A fixture with one track + orphan transcript distinguishes the counts. No new query subsystem needed.

DECISION: 5. [logical gaps; 1 lens] NON-BLOCKING / actionable / OPEN — Remove the ineffective CSS example.
Where: docs/admin-editor-ux/plan.md:31–34.
Evidence: CourseEditorView.tsx:692 contains Tabs; :716–724/:852 put Overview/Lessons inside form; only Access at :885 is a direct child. `.course-tabs > .ui-tab-panel[inert]` misses the two panels responsible for most B1 space.
Impact: Copying the first suggested selector retains B1. This is non-blocking because D1 already offers a correct dedicated class and Ux01 should catch it.
Plan edit/check: Name a dedicated course-panel class on all three panels (or correct explicitly scoped selectors), keep language panels independent, and assert computed display for all three tab selections.

DECISION: 6. [edge cases; 1 lens] NON-BLOCKING / actionable / OPEN — Specify focus after reorder/deletion removes its trigger.
Where: docs/admin-editor-ux/plan.md:42–43, :78–79.
Evidence: CourseEditorView.tsx:386–408/:481–496 renders controls in movable rows. D3 omits impossible controls, so pressing Up on the second row removes the focused Up control when it becomes first. Confirmed deletion removes the dialog trigger as well. Existing reorder tests assert values, not activeElement (CourseEditorView.test.tsx:471–478).
Impact: Keyboard users can lose their position after a valid action.
Plan edit/check: Define focus on a surviving action/heading in the moved row; after delete choose next/previous row or Add control when empty. FE keyboard assertions for move-to-first/last and deleting the final row. This is a small behavior detail, not a request for an abstraction.

FACT: R1 — complete course-editor browser-driver inventory from every frontend/e2e/*.spec.ts. Searches: admin/courses browser goto, stored editorPath/ownEditorPath, labels/roles/locator helper calls; API fixture writes distinguished from browser drivers. Native file input is normally hidden by FileUpload, so the required visibility assertion belongs on its user-facing button/section, not the native input itself.

| Spec and concrete location | Concrete selectors | Effect of D1/D4; plan migration |
|---|---|---|
| media.spec.ts:147–189; entrypoints :365/:417/:457/:494/:567, reload :404–408 | `getByRole('group',{name:\`Урок ${index}\`})`; `getByRole('button',{name:'Завантажити відео'}).toBeVisible()`; `locator('input[type="file"]:not(.admin-audio-field input)')`; `locator('[data-video-state]')`; audio `getByRole('region',{name:'Аудіодоріжки'})`, button `/Мова доріжки/`, textbox `Назва доріжки`, default/delete buttons (:432/:444) | Already clicks Lessons. Five saved lessons default collapsed: reveal requested row before Upload/Ready/progress/audio assertions and after reload. Visible Upload assertion is a definite failure without reveal. |
| audio-tracks.spec.ts:89–128, :261–268 | nested groups `Модуль 1` → `Урок 1/2`; video file CSS input; `[data-video-state="ready"].toBeVisible()`; audio region `Аудіодоріжки`, `/Мова доріжки/`, `Назва доріжки`, audio `input[type="file"]` | Already Lessons; two saved lessons default collapsed. Reveal BOTH targeted lessons; Ready/audio assertions otherwise hidden. |
| subtitles.spec.ts:591–625 | groups `Модуль 1` → `Урок ${index+1}`; video CSS file input; `[data-video-state="ready"].toBeVisible()`; `Аудіодоріжки` region → `/Мова доріжки/` / `Назва доріжки` | Already Lessons; three saved lessons default collapsed. Expand each loop target before upload and retain visible Ready/audio checks. |
| playback.spec.ts:276–325; primary fixture :56–85 | groups `Модуль ${target.moduleNumber}` → `Урок ${target.lessonNumber}`; video CSS file input; API list/state poll | Already Lessons; four saved lessons across two modules default collapsed. If outer group stays visible, file input + API poll can pass under hidden body. Explicit expand + visible Upload/section assertion required; do not infer visibility from API Ready. Cross-course single-lesson target need not default-collapse. |
| purchase.spec.ts:116–130, :237–255 | Lessons → `getByText('Відкритий до покупки — урок 2, модуль 1').click()`; matching `getByRole('checkbox',...)`; Overview → `Ціна, грн` / `Зберегти зміни` / `Зберегти ціну`; Access → region `Видати доступ до курсу` | Two saved lessons collapse. If preview checkbox remains operative in collapsed header, preserve these assertions there; if it moves into body, expand lesson 2 first. D4 promises a preview flag, not explicitly an interactive checkbox. Overview/Access navigation is already explicit and compatible with D1. |
| discussion.spec.ts:79–101 | default Overview; `page.locator('input[type="file"]:not(.admin-audio-field input)').first().setInputFiles(shortVideo)`; API Ready poll | Missing Lessons selection now; D1 makes actual target panel hidden. CSS locator does not establish visibility and file upload may still pass. Add real Lessons click + visible upload section/button assertion; one lesson starts expanded. |
| transcript.spec.ts:286–310, :555–599 | Lessons click; groups `Модуль 1` → `Урок 1` → video CSS file input; `getByRole('textbox',{name:'Текст репліки 1'})`; buttons `Відновити зміни` / `Зберегти транскрипт` | Single-lesson fixtures; explicit Lessons selection, including reload :576/:595. No expected collapse regression; retain current visible user-path assertions. Add visible Upload assertion to API-polled upload helper for evidence. |
| asr.spec.ts:107–128, :237–277; reload/tab selections :595/:641 | `getByRole('tab',{name:'Уроки'})`; CSS video file input; `[data-video-state="ready"].toBeVisible()`; visible `getByRole('tabpanel').filter({has: heading 'Транскрипт уроку'}).last()`; `tablist 'Мови транскрипту'`; ASR/draft/save buttons and transcript textboxes | Single-lesson fixture starts expanded; already selects Lessons. No default hidden-panel/collapse failure predicted. Preserve selected-language scoping and heading/region accessible names when applying D5. |
| enrollment.spec.ts:186–225, :243–254, :505–544 | Lessons → groups `Модуль 1`/`Урок 1` → video CSS file input; API Ready poll; reload Lessons → checkbox/getByText `Відкритий до покупки — урок 1, модуль 1` | Single-lesson fixture; already selects Lessons on load/reload; no default-collapse failure. Keep groups/accessibility contract or update scoping deliberately, assert visible Upload before file input. |
| admin.spec.ts:94–108, :274–287, :305–306, :354–376 | default Overview `getByLabel('Назва')`, `Зберегти зміни`, error summary link `Назва`; Lessons `Додати модуль`, `Назва модуля 1`, `Додати урок до модуля 1`, `Назва уроку 1 у модулі 1`; Overview `Опублікувати`; reopened Overview `Архівувати` | Correct tab transitions. Newly added lesson starts expanded per D4; no existing multi-lesson collapse interaction. Existing validation case proves Overview title focus only, so Ux02 needs NEW opposite-tab + collapsed-lesson invalid-title cases; not reuse this case as sufficient proof. |

FACT: Remaining specs were searched: lesson-builder.spec.ts opens `/admin/courses/{id}/lessons/{lessonId}` directly (:106/:179/:299), the explicitly excluded lesson-document surface; teach-editor.spec.ts uses /teach. route-guards.spec.ts visits `/admin/courses` list, not editor. accounts/instructor/learn-access/learning/purchase-wayforpay/ratings/refund and other specs create courses via API or drive other surfaces; no further course-editor browser driver found. API fixture calls alone are not hidden-panel interactions. Inventory does not claim runtime outcomes.

FACT: Source-audit disposition retained: B1→Ux01–02; M1→Ux04–05; M2→Ux07–08; M3→Ux11–12; M4→Ux09–10; M5→Ux06; M6→Ux13; m1→Ux14; m2→Ux15; m3→Ux03. These are planned, not resolved by critique. m4→Ux16 is deliberately limited to originalName: actual MediaVideo schema at backend/openapi/openapi.yaml:2381–2437 contains filename and no duration/thumbnail. This is a sourced mechanism for excluding duration under the no-backend-change scope. Course materials remain course-level (DocumentLibrary.tsx:31–34, CourseEditorView.tsx:852–856); keep their h2 rather than interpreting D5 as moving them into every lesson.
FACT: Lens outcomes: AI-slop CLEAN (concrete audit-backed users/results, no invented subsystem); unnecessary-complexity CLEAN (existing Dialog/reset/IDs/routes/scoped CSS, small parser without suitable existing equivalent); logical-gaps 1 blocker + 2 non-blocking; edge-cases 2 blockers + 1 non-blocking. Shared D6 finding deduplicated into #1. Coordinator added #3 from the explicitly requested concrete R1 inspection. Scratch progress is separate per lens under /private/tmp/courses-admin-editor-ux-{ai-slop,logical-gaps,complexity,edge-cases}-20261118.md; durable consolidated evidence is here.
FACT: Scope/limitations: static file and original screenshot inspection only. No tests, builds, browser execution, network probes, git mutation, runtime state changes or graph refresh. make/check/browser/E2E outcomes are NOT_RUN, not passed. All ten audit screenshots inspected, their old snapshot is not presented as current live evidence. In the repository only this append-only working file was written.
FACT: Previous verdict/delta: no earlier plan-critic verdict supplied or applied. This is the initial exact-r1 verdict; raw audit remains evidence of the original problem, not a prior accepted plan. No old PASS is carried to a new hash. Findings concern plan decisions/contracts and are argued with sources; no required-input QUESTION remains.
NEXT: Plan owner Courses 1 revises D6/U3, D7/U5 and R1/U1–U2 verification strategy; reconcile non-blocking #4–6. Re-critique the materially changed affected scope at the new exact plan hash, retaining unchanged evidence explicitly. Current r1 remains BLOCKED with 3 open blockers. No implementation is authorized by this critique.
FACT: Final read-only integrity check: 55 scoped source/screenshot fingerprints rechecked; changed inputs = []; plan SHA256 = 62ed935e8e6070c690964db2817ef7807da276004b21197c968c44a99c62555b. Verdict appended at line 134; 3 blocking / 3 non-blocking findings; 4/4 independent lenses returned.
DECISION: Critique complete; final delivery is a concise summary of this working file. Repository writes restricted to this file; no further action taken.
NEXT: Critique r2 delta: verify current entry rules, exact plan hash/snapshot, six r1 dispositions and changed D4/D6/D7/U1-U2/§6 against cited code. Append only to this working file; write nothing else.
FACT: Current agent-entry reread. User authorized r2 delta review only; latest “write nothing else” narrows writes to this existing working file, including outside-repository scratch. No new progress files or append to old scratch files will be made.
NEXT: Verify root/local rules, prompt/skill identities and r2 text; reuse the four selected independent reviewers read-only, with coordinator logging bounded returned evidence here.
FACT: r2 SHA256 verified: 7b00873446d9008ac6966a2833ab9f780696a63834c74894e7ebde7e89417b80; prompt v1 hash unchanged; HEAD/origin/main still 10436cdf2988a17887c959dd48b363e6b117e9d7, branch admin-editor-ux.
DECISION: Apply all four lenses to material delta, using existing independent reviewers. Latest owner write constraint overrides per-agent scratch-writing ceremony: reviewers are read-only and return bounded findings to coordinator; only /root appends this file. No new reviewer route or model override.
NEXT: Dispatch r2 delta to selected AI-slop, logical-gaps and complexity reviewers; edge-cases follows when slot frees.
FACT: All 55 r1 source/audit/e2e fingerprints remain unchanged. AI-slop reviewer re-dispatched read-only for r2; r2 reviews launched 1 / returned 0.
NEXT: Dispatch logical-gaps reviewer with same pinned r2 delta.
FACT: Logical-gaps reviewer re-dispatched read-only for r2; r2 reviews launched 2 / returned 0.
NEXT: Dispatch unnecessary-complexity reviewer, then independently inspect changed semantics and R1 migration completeness.
FACT: AI-slop r2 returned CLEAN (0 findings); no files written. D4/D6/D7 reductions retain concrete user outcomes; U1/U2 helper has repeated consumers. r2 reviewers launched 3 / returned 1.
NEXT: Dispatch edge-cases reviewer and check U1 upload assertions against replacement-label paths, U2 openLesson repeated invocation and D3 collapsed-body focus.
FACT: Edge-cases reviewer re-dispatched read-only; all 4 selected independent r2 lenses launched / 1 returned.
NEXT: Inspect concrete changed-contract callers rather than accepting §7 disposition labels.
FACT: Complexity r2 returned CLEAN (0 blocking/0 non-blocking), confirming D4 callback reuse, D6 narrow readout and D7 removal avoid new mechanisms. r2 launched 4 / returned 2.
FACT: U1 replacement path needs closer check: uploadReady is reused at playback.spec.ts:1573; LessonVideoField.tsx:299–301 changes the user-visible label to replace for ready/awaiting_audio assets.
NEXT: Inspect replacement helper and recovery save semantics; distinguish real new contract gaps from harmless shorthand.
FACT: Logical-gaps r2 returned ACCEPTED for lens: prior LG1/LG2/LG3 closed; 3 new non-blocking clarifications (idempotent openLesson, unknown transcript badge, byte-identical wording). r2 launched 4 / returned 3.
FACT: Playback B40 (playback.spec.ts:1561–1576) reuses uploadReady on a lesson with an existing Ready video; visibility assertion must accept the actual Replace control, not require Upload literally. Previous i18n lookup glob had no matching path; no evidence inferred from it.
NEXT: Check actual localized labels and missing-knowledge callback behavior; complete edge-cases disposition, then consolidate r2.
FACT: Edge-cases r2 returned BLOCKED: prior recovery/Discard/focus findings closed, but openLesson unconditionally toggles already-expanded single lessons (playback.spec.ts:695–718). r2 4 launched / 4 returned, 0 missing, 0 retries; no reviewer wrote files.
DECISION: Consolidation will treat idempotence as BLOCKING, merging logical-gaps and edge-cases reports: explicit single-lesson production fixture makes this more than an optional helper refinement. Other clarified scopes remain retained; no r1 blocker is silently waived.
NEXT: Verify exact single-lesson caller and append final r2 verdict with per-finding closure and remaining scoped findings.
FACT: Final r2 evidence check confirmed playback.spec.ts:695–700 has exactly ONE cross-course lesson and :709–718 calls the same uploadReady migration target. D4 r2:60–62 explicitly starts it expanded. Therefore the literal U2 toggle prescription fails an existing caller, independently of repeated-call concerns.
DECISION: R2 delta critique consolidated below. The six r1 dispositions are judged against code, not accepted from §7 alone. No missing required input; no QUESTION/STOP needed.

FACT: stage: plan; artifact docs/admin-editor-ux/plan.md r2 (2026-11-18), SHA256 7b00873446d9008ac6966a2833ab9f780696a63834c74894e7ebde7e89417b80. Root <repo>, branch admin-editor-ux, HEAD = origin/main = 10436cdf2988a17887c959dd48b363e6b117e9d7. Prompt <kit>/prompts/plan-critic.md version 1, SHA256 514e4c376c2551afe76e390c5779867ecf7f04ad2a552c97c2b4f713c43dcab6; same canonical managed release/skill route as r1, current entry/local rules reread. All 55 r1 source/audit/e2e/screenshot fingerprints rechecked unchanged. Retain the r1 code evidence and full R1 selector inventory; only plan decisions/acceptance/migration strategy changed.
FACT: Critic /root (Codex) and same four selected independent review contexts: /root/ux_ai_slop, /root/ux_logical_gaps, /root/ux_complexity, /root/ux_edge_cases. R2: 4 reviews dispatched / 4 final verdicts returned; 0 missing / 0 retries / no new agents or model changes. Their contexts do not inherit authoring history or each other's review results. Latest owner “write nothing else” overrides scratch-file ceremony: reviewers returned read-only results, and only coordinator appended this file.

FACT: Closure of r1 numbered findings (plan-level closure, not implementation verification):

| r1 # | Disposition after checking r2 | Plan text and retained code evidence |
|---|---|---|
| 1, raw-ms recovery | CLOSED — representation-change risk removed by a recorded scope decision, not by an unproven migration. Raw-ms editing/storage remains; readout alone formats time. | r2:65–73, :97–98. recovery.ts:14–27 and IndexedDB adapter:5–25 retain string times; TranscriptEditor.tsx:359 restores them unchanged and :519–520 keeps existing Add-line arithmetic. D6 no longer needs a new stored-format reader. “Byte-identical” needs non-blocking precision below. |
| 2, Discard/reset | CLOSED — the new Discard action is removed explicitly; this does not claim existing Save/reset remounting was repaired. | r2:74–81, :101, :148. CourseEditorView.tsx:334–347/:449–462 uses generated formKeys; RHF reset regeneration evidence retained. No new reset path remains in U5. Audit M3's proposed Discard is intentionally omitted, not implemented. |
| 3, false-green visibility proof | CLOSED as originally raised — D1 disclaims aggregate proof; U1/U2 prescribe real reveal and visible assertions. The replacement helper introduces new blocker R2-1 below. | r2:39–42, :111–122, :139–142. discussion.spec.ts:79–101 needs Lessons selection; retained Playwright _setInputFiles evidence proves why explicit visibility is necessary. U1/U2 now each require full E2E before review closes. |
| 4, track/tab count conflict | CLOSED — both count variants removed; D4 and Ux04 agree. Deliberate count omission is retained as a scope tradeoff, not claimed as implementing the audit's count proposal. | r2:53–62, :94, :150. LessonLanguageTabs.tsx:34–61 proves tracks and projected language tabs differ. Existing onTranscriptKnown is a usable source; unknown-state detail is R2-2. |
| 5, wrong child selector | CLOSED — exact dedicated class on all three course panels. | r2:33–42, :151. CourseEditorView.tsx:716–724/:852 nests two panels in the form; :885 is direct child. `.course-editor-panel[inert]` covers all without changing the global/language primitive. |
| 6, focus after removed trigger | CLOSED — D3 names surviving-control and next/previous/Add fallbacks; acceptance now names activeElement. | r2:48–52, :99–100, :152. Existing movable action subtrees at CourseEditorView.tsx:386–408/:481–496 justify the fallback. D4 keeps next-row title fields visible even if their bodies are collapsed. |

DECISION: R2-1. [logical gaps + edge cases; 2 lenses] BLOCKING / actionable / OPEN — Make openLesson an idempotent reveal operation.
Where: docs/admin-editor-ux/plan.md:117–121, interacting with :60–62.
Evidence: playback.spec.ts:695–700 creates one lesson; :709–718 uses uploadReady, whose :276–325 path is explicitly migrated “each target.” A single lesson starts expanded under D4. audio-tracks.spec.ts:264–267 also performs upload then two addTrack calls on the same row, so the helper can be invoked repeatedly before body interactions.
Impact: Literal implementation of “clicks the row's expand toggle … asserts the body visible” closes the already-expanded single lesson and fails the assertion; repeated calls can close an already-open multi-lesson row. This is an existing concrete caller, not a hypothetical edge. It prevents the new migration from meeting its own full-suite gate.
Plan edit: At :117–118 specify that the helper first reads the scoped toggle's aria-expanded, clicks ONLY when false, and always asserts true/body visible. Keep the existing tab-selection responsibility explicit at the call site. Verify initially collapsed multi-lesson, initially expanded single-lesson and repeated invocation; no force/includeHidden/API bypass.
Disposition: Logical-gaps called this non-blocking; edge-cases called it blocking. Consolidated as blocking because the single-lesson playback fixture deterministically contradicts the prescribed unconditional operation and requires correction/rework. This does not reopen recovery or Discard scope.

DECISION: R2-2. [logical gaps; 1 lens, coordinator verified] NON-BLOCKING / actionable / OPEN — Do not turn unknown transcript knowledge into “No”.
Where: docs/admin-editor-ux/plan.md:55–56, :94.
Evidence: CourseEditorView.tsx:544 initializes transcriptKnowledge as {}; LessonLanguageTabs.tsx:123–125/:142–145 calls onTranscriptKnown only after inventory data exists. No initial value is established during loading or an uncached failure.
Impact: A naïve boolean fallback gives a false negative badge for a lesson whose transcript has not loaded. Existing callback reuse is valid; it simply has a third initial state.
Plan edit/check: Omit the badge or show an explicit unknown/loading state until known, preserve a known value on failed refresh, and distinguish undefined from false in a component assertion. No new query required.

DECISION: R2-3. [logical gaps; 1 lens, coordinator verified] NON-BLOCKING / actionable / OPEN — Define equality at the timing boundary rather than promising byte-identical saved snapshots.
Where: docs/admin-editor-ux/plan.md:73, :98.
Evidence: TranscriptEditor.tsx:359 restores raw strings, but :303–311 converts times with Number and creates new segment UUIDs. A saved API payload is not byte-identical to an IndexedDB snapshot even today.
Impact: Literal whole-record equality would assert a nonexistent baseline and invite unnecessary storage/ID changes. The intended no-loss timing property is supported by unchanged raw-ms editing.
Plan edit/check: Require restored raw time TEXT to be unchanged, and submitted numeric millisecond VALUES to match the original exact integers, including values finer than 100 ms; keep partial text recoverable. Do not require equal generated IDs/JSON bytes. R3 at :133 still says “format change / pure round-trip”; align it with the new display-only decision as editorial cleanup.

DECISION: R2-4. [coordinator, concrete R1 caller check] NON-BLOCKING / actionable / OPEN — Interpret the visible upload assertion as the active Upload OR Replace control.
Where: docs/admin-editor-ux/plan.md:111–114.
Evidence: playback.spec.ts:1561–1576 (B40) reuses uploadReady for an already-ready lesson. LessonVideoField.tsx:299–301 selects administration.video.replace for ready/awaiting_audio, otherwise administration.video.upload; locales/uk.ts:809–810 names these “Завантажити відео” and “Замінити відео”.
Impact: An exact Upload-only name in the migrated helper would fail a supported replacement journey. R2 does not mandate an exact label selector, so this is a clarification, not an additional blocking contract choice.
Plan edit/check: Say “visible active video-upload/replacement control” scoped to the target lesson; preserve B40's fresh-video-ID polling and replacement assertions. Do not weaken this to native-file-input existence.

FACT: Other changed decisions checked: D4's always-visible title/preview inputs match CourseEditorView.tsx:349–367 and preserve purchase.spec.ts:120–123 plus opposite-tab validation targets without needing collapse expansion. Per-ID expansion and mounted bodies preserve local upload/transcript state on tab/collapse changes; implementing display-only collapse still requires Ux05 proof. D6 readout has a concrete human-readable purpose, with no evidence requiring a stored-format switch. D7 is a dated, mechanism-backed scope reduction; existing Save behavior and sticky-bar layout risks remain in scope as stated, with no new reset mechanism inferred.
FACT: Cadence check: §6 now explicitly requires FULL E2E for both course-panel/lesson-body visibility slices U1/U2 before their reviews close, via Kiri wrapper, plus final ordered make check. The prior affected-only conflict for these slices is closed. General AGENTS requirements for shared/persisted-state effects remain authoritative for any later slice; “affected specs” does not waive them. No listed command was executed in this critique.
FACT: Four-lens results: AI-slop CLEAN (D4 existing controls, D6 useful readout, D7 useful save visibility, repeated helper consumers); unnecessary-complexity CLEAN (no new recovery/reset/query subsystem); logical-gaps 0 blockers/3 non-blocking before consolidation; edge-cases 1 blocker. Duplicate openLesson reports merged into R2-1; coordinator added actual Upload/Replace caller clarification R2-4. Author §7 dispositions were not used as independent evidence.
FACT: What is retained since r1: exact unchanged product-source/audit/screenshot fingerprints, original ranked audit, complete R1 selector inventory and source-backed state ownership. What changed: r2 explicitly narrows D4 count/D6 editing/D7 Discard scope, replaces ineffective CSS example, specifies focus and migration/cadence. This fresh material-delta review binds only r2 SHA above; it neither erases r1 BLOCKED history nor accepts a later plan hash. Backend, /teach and lesson-document-editor exclusions remain unchanged.
FACT: Limitations: static file review only; no tests, builds, browser execution, network probes, graph refresh, git mutation or product writes. All future component/E2E/DP/gate results remain NOT_RUN. Only this working file was appended; reviewers wrote no files. Required inputs are present.
NEXT: Owner Courses 1 makes openLesson idempotence explicit at r2:117–118, then obtain a bounded delta check/rebinding for that changed plan hash. R2-2–4 are non-blocking implementation/test wording clarifications. Open blockers: 1; non-blocking findings: 3.
DECISION: BLOCKED for r2.
FACT: R2 final integrity check: 55 retained fingerprints checked; changed inputs = []; current plan SHA256 7b00873446d9008ac6966a2833ab9f780696a63834c74894e7ebde7e89417b80; r2 record starts at line 225. No plan or product writes.
DECISION: Final r2 delta verdict: BLOCKED for r2 (1 open blocker; 3 non-blocking findings; 4/4 independent lens results returned).
NEXT: Bounded r3 delta check: verify current entry rules, r3 hash/source snapshot, R2-1..R2-4 dispositions and any new gap introduced by the edits. Append only this working file; no tests/builds/network/product writes.
FACT: Current entry rules reread; active task is the explicitly identified r3 delta, same working file. Prior r2 record and source fingerprints are available locally; no archive lookup required.
NEXT: Verify root/local instructions, exact artifact and prompt identity, then read r3 against recorded r2 findings.
FACT: r3 SHA256 matches 039fb8c3c169b98df22df14b28fefbc2ed834fb69213618d397ea9ed4894a495; prompt v1 hash unchanged; HEAD/origin/main10436cdf2988a17887c959dd48b363e6b117e9d7, admin-editor-ux; all55 retained input hashes unchanged.
NEXT: Reuse selected independent lenses for bounded R2-1..R2-4 closure/new-gap check. No agent writes; coordinator records bounded results here. Preserve unchanged r1/r2 evidence.
FACT: AI-slop r3 delta dispatched read-only; r3 1 dispatched / 0 returned.
NEXT: Dispatch logical-gaps to validate affected contracts against source.
FACT: Logical-gaps r3 delta dispatched read-only; r3 2 dispatched / 0 returned.
NEXT: Dispatch complexity for the narrow revised mechanisms.
FACT: AI-slop r3 returned explicit CLEAN (0 findings); corrected helper/badge/timing/button text all have concrete existing consumers/contracts. r3 3 dispatched / 1 returned.
NEXT: Dispatch edge-cases; re-read the decisive single-lesson/repeated-call, callback, serialization and replacement branches.
FACT: R3 logical-gaps returned ACCEPTED/CLEAN: R2-1..4 closed with exact cited contracts. Complexity returned CLEAN: guarded helper, three-state badge and corrected timing/button assertions add no new mechanism. r3 4 dispatched / 3 returned; no files written by reviewers.
FACT: Coordinator re-read decisive source: playback single-lesson call (695–718), repeated audio-track operations (264–267), inventory callbacks (123–125/142–145), parsed ms + new UUIDs (303–311), Upload/Replace branch (299–301), B40 replacement (1561–1577). R3 text fits all these paths.
NEXT: Consume final edge-case delta result, then append bounded r3 closure matrix and verdict; retain prior unchanged evidence.
FACT: Edge-cases r3 returned explicit CLEAN: idempotent helper handles initially expanded single lessons and repeated same-row calls; other revised boundaries fit their code contracts. R3 bounded re-review: 4 dispatched / 4 final results returned, 0 missing / 0 retries. Every lens reports zero new findings. No reviewer wrote any file.

FACT: stage: plan; bounded r2→r3 delta check; critic /root (Codex), same independent AI-slop/logical-gaps/complexity/edge-cases contexts retained. Artifact docs/admin-editor-ux/plan.md r3, SHA256 039fb8c3c169b98df22df14b28fefbc2ed834fb69213618d397ea9ed4894a495. Root <repo>; branch admin-editor-ux; HEAD = origin/main = 10436cdf2988a17887c959dd48b363e6b117e9d7. Prompt <kit>/prompts/plan-critic.md version 1, SHA256 514e4c376c2551afe76e390c5779867ecf7f04ad2a552c97c2b4f713c43dcab6; same managed canonical skill/release route. All 55 retained source/audit/e2e/screenshot fingerprints remain unchanged.

DECISION: R2 finding dispositions after checking actual r3 text, not §8 labels alone:

| Finding | r3 text | Source check and disposition |
|---|---|---|
| R2-1 (blocking), non-idempotent openLesson | plan.md:122–125 reads scoped aria-expanded, clicks ONLY false, always asserts true + visible body; caller selects Lessons | CLOSED. playback.spec.ts:695–718 is an initially expanded single-lesson target; audio-tracks.spec.ts:264–267 repeats operations on one row. Both are now covered without collapsing an open body. Guard uses actual UI state rather than a new tracker. |
| R2-2 (non-blocking), unknown transcript badge | plan.md:55–57 explicitly defines unknown=no badge, yes/no after knowledge arrives, keeps known state on failed refresh | CLOSED. CourseEditorView.tsx:544/580–583 holds optional per-lesson knowledge; LessonLanguageTabs.tsx:123–125/:142–145 only publishes available inventory. No false undefined→No conversion is prescribed. |
| R2-3 (non-blocking), byte-identical overclaim | plan.md:74–76/:101 specifies unchanged restored raw time text and submitted integer milliseconds, including sub-100ms; excludes generated IDs/JSON; risk R3 aligned at :140 | CLOSED. TranscriptEditor.tsx:359 directly restores text; :303–311 parses integer ms and mints new UUIDs. Revised acceptance distinguishes precision preservation from nonexistent whole-record equality. Raw-ms recovery/ASR/Add-line contracts remain unchanged. |
| R2-4 (non-blocking), Upload-only assertion | plan.md:117–119 allows target lesson's visible Upload OR Replace control and retains B40 fresh-asset-ID polling | CLOSED. LessonVideoField.tsx:299–301 changes the label for ready/awaiting_audio; playback.spec.ts:1561–1577 exercises replacement, and :303–325 polls the new asset's ID. Existing replacement evidence remains meaningful. |

FACT: New-gap check CLEAN across all four lenses: r3 adds only a conditional reveal, explicit unknown-state rendering, a corrected equality boundary and a state-appropriate control assertion. It introduces no new persistence format, reset path, query, cross-module responsibility or verification bypass. D4's header fields, D6 raw-ms/readout, D7 no-Discard, prior r1 closures and §6 full-E2E cadence for U1/U2 remain as accepted in the r2 delta review. Retained source evidence and full R1 selector inventory stay applicable; no unrelated scope was reopened.
FACT: Changes since previous verdict: r2 BLOCKED had 1 blocker and 3 non-blocking findings; r3 text closes all four. Open findings now: 0 blocking / 0 non-blocking. R1/r2 historical verdicts remain intact; this acceptance binds only the exact r3 hash and unchanged source snapshot above.
FACT: Limitations: static plan/code delta review only. Tests, builds, browser checks, network probes, runtime/graph operations and git mutations NOT_RUN; no implementation/product PASS implied. Only this append-only working file was written, including by the coordinator; reviewers returned bounded results without scratch writes. Required inputs present; no unanswered QUESTION.
NEXT: Plan owner may proceed to the next required planning gate / authorized slice under existing owner-consent, independent implementation-review and verification rules. This critique does not execute or authorize runtime commands, code changes, commit or deployment.
DECISION: ACCEPTED for r3.
FACT: Final r3 integrity check: 55 retained inputs, changed = []; plan SHA256 039fb8c3c169b98df22df14b28fefbc2ed834fb69213618d397ea9ed4894a495; r3 verdict record starts at line 289.
DECISION: Final bounded delta verdict: ACCEPTED for r3.
