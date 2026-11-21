# TE5 — frontend teach editor route + nav (WIP, Codex working file)

РОБОЧИЙ ФАЙЛ. Codex дописує ПІСЛЯ КОЖНОГО КРОКУ (append, не rewrite).
Перший запис — рядок `ДАЛІ:` до першого інструменту. Далі — ФАКТ / РІШЕННЯ / ДАЛІ / ПИТАННЯ.

## Контекст
- Гілка `feat/teach-editor`, HEAD має TE4 (`e6b241f`): `modules/document` teach `documents` у композиції.
- PR #2 = TE4 + TE5 (frontend). Це run — **імплементація TE5 (без component-тестів; тести — окремий run)**.

## Журнал
ДАЛІ: перевірити гілку, briefing, workflow та аналоги admin/teach.
ФАКТ: гілка feat/teach-editor; наявні сторонні dirty/untracked файли збережено. Прочитано briefing, adapter, план/spec; TE5 — лише frontend implementation, тести окремим run.
РІШЕННЯ: повторно використати admin editor та mapping, без backend/Orval змін і без commit/stage/push.
ДАЛІ: звірити точні API query, lesson-list та guard перед змінами.
ФАКТ: teach.tsx викликає requireIdentity + canAccessAdministration та Outlet; admin editor lazy-load і null meta вже придатні для teach. Mapping фактично у modules/admin/adapters/admin-course-repository.ts, не admin-queries.ts; це уточнення шляху, не відсутній контракт.
ДАЛІ: перевірити mapper/transport, admin lesson links і порожній draft; обрати мінімальне публічне повторне використання.
ФАКТ: HEAD e6b241f83ed7e6971bddc22600a6ac8227485cfb містить TE4 documents; GATE 2 плану PASS (§11). Existing admin CourseEditorView рендерить ModuleFields/LessonFields, але не має готових editor links чи окремого zero-lessons EmptyState; готового компонента для копіювання немає. Document editor має sections.length === 0 → EmptyState.
РІШЕННЯ: 2026-10-30, implementer — inline TE5 implementation у схваленій фазі: export existing validated course/error mappers через admin public barrel; teach composition викликає лише getTeachCourse через штатний transport, query key teach/courses/tree. Новий route копіює admin editor, список використовує Link + EmptyState і uk/en keys. Перевірка рішення: diff + typecheck + architecture lint; behavioral proof лишається наступним component/e2e run.
ФАКТ: pre-code READY у scope доручення; writers — лише цей implementer для TE5, сторонні зміни поза scope. File fallback: перевірені admin/teach routes, composition, admin-course-repository/schema, document editor, generated getTeachCourse; graph не запускався. Немає нового backend/API/стану persistence. Незалежне code review — після цього run за дорученням.
ДАЛІ: реалізувати mapping reuse, query, route та lesson navigation; потім format/typecheck/lint.
ФАКТ: додано teach detail через getTeachCourse + існуючі Zod course/error mappers; новий lazy route передає teach.documents; course view містить модулі/lesson links, loading/error/retry/empty states; додано uk/en keys. Backend, generated API та editor не змінено.
ДАЛІ: згенерувати лише TanStack route tree для реєстрації маршруту, перевірити класи/переклади, виконати три дозволені команди.
ФАКТ: TanStack Generator.run → PASS, exit 0 (лише наявне попередження про route-guards.test.tsx). Це route-tree generation, не OpenAPI/Orval. Звірено shared button class і виправлено retry на існуючий common.retry.
ДАЛІ: npm run format → npm run typecheck → npm run lint у frontend.
ФАКТ: frontend/ npm run format → PASS, exit 0 (376 files). Native checkpoint спроба → FAIL, exit 1: sandbox заборонив mktemp у зареєстрованому receipts; це не product gate, повтор через дозволену escalation.
ДАЛІ: typecheck/lint та checkpoint у тому самому native root/session/store.
ФАКТ: npm run typecheck → FAIL, exit 2: existing route-guards fixture не має нового teach.detail; LessonDocumentEditor props досі типізовані лише admin query-key prefix, тож TE4 teach queries несумісні на рівні типів. npm run lint → PASS, exit 0 (architecture 230 files). Native checkpoint → OK, exit 0, b997b10e98526e349cc48feaa37ac714-1-c36af830648c.
ДАЛІ: перевірити точний query factory contract, щоб визначити мінімальне типове підключення без зміни поведінки чи нових тестів.
РІШЕННЯ: 2026-10-30, implementer — прибрати literal-prefix generic лише у внутрішньої query factory (keyPrefix: 'admin' | 'teach'), щоб обидві обгортки відповідали одному editor props contract. Runtime keys, endpoints, invalidation та поведінка editor не змінюються; перевірка — typecheck/diff. Existing route-guards fixture отримує тільки detail stub; нові component-тести/сценарії не пишуться.
ДАЛІ: повторити format/typecheck/lint після цієї типової інтеграції.
ФАКТ: повторні npm run format → PASS exit 0; npm run typecheck → FAIL exit 2 (лише fixture: throwing queryFn inferred never); npm run lint → PASS exit 0. Editor/teach prefix incompatibility усунено.
ДАЛІ: явно типізувати fixture Promise<CourseTree>, повторити перевірки й оглянути кінцевий diff.
ФАКТ: третій format → PASS exit 0; typecheck → FAIL exit 2 і lint → FAIL exit 1 через мій дубль уже наявного CourseTree import. Дубль прибрано; залишено існуючий import та явний Promise<CourseTree>.
ДАЛІ: фінальні три перевірки на виправленому snapshot.
ФАКТ: фінальний frontend/ npm run format → PASS, exit 0; npm run typecheck → PASS, exit 0; npm run lint → PASS, exit 0, architecture boundaries 230 files.
ДАЛІ: фінальний diff/scope check, checkpoint і передача незалежному reviewer; T9/T13 behavioral evidence лишається NOT_TESTED у цьому run.
ФАКТ: фінальний diff перевірено: generated route parent = TeachRoute; teach course tree → getTeachCourse, documents → TE4 teach adapter. Нових /admin endpoint calls у teach flow немає. git diff --check без повідомлень. Сторонні .gitignore/security WIP збережені; git add/commit/push не виконувалися.
ФАКТ: T13 реалізація повторно використовує незмінений LessonDocumentEditor (порожні sections → EmptyState), новий список має EmptyState для відсутніх lessons у всіх modules. Component/browser/e2e/integration/make check НЕ запускалися; behavioral acceptance T9/T13 і незалежне review НЕ оголошуються PASS.
РІШЕННЯ: передати implementation у робочому дереві координатору. Змінено 10 frontend файлів (включно з новим route, generated route tree та лише доповненням existing test fixture). Backend/OpenAPI/Orval/dependencies без змін.
ДАЛІ: незалежне review цього diff, окремий component-test run, TE6 e2e/full gate; coordinator вирішує commit.
ФАКТ: SHA-256 фінального frontend snapshot (HEAD e6b241f):
- frontend/src/app/composition/teach.ts: b2e80ec76381fb0ecec4d3f5fb8c5a4196cdca261d0ea873917874ea49e7b98e
- frontend/src/app/providers/locales/en.ts: ea7d9280677fa52563c7232115506f5531f4c769822727f4d6de51c83649a1db
- frontend/src/app/providers/locales/uk.ts: 3eb925dd5a3d8894641a20de1cc6f586f168258279f159456f1c36a287b3efa9
- frontend/src/modules/admin/adapters/admin-course-repository.ts: e08281662a6b3c052a3153e066ef829c5be4cc4cc5a55e96c4ddb698d439f03b
- frontend/src/modules/admin/index.ts: 6253f92052d60658c2fc795ae8ef1361f0b0701d1486d3153f5aff798b7c0a51
- frontend/src/modules/document/application/admin-document-queries.ts: 75628581c6cea76b4e0abb559c2c6a0a674c6be0766343e4bc5be40945083818
- frontend/src/routeTree.gen.ts: 4a37aa7cb08134f8e25eb9ae7b9c1d3699ebc44e50c8693ccf03c3f1a054ecf8
- frontend/src/routes/route-guards.test.tsx: 4fb671df2ca18786a3a97a933c83498548a395cec74138e5ac7f0ffab1065f2b
- frontend/src/routes/teach.courses.$courseId.tsx: 70c59cde1565beba9bf8b44fe6fe91d1c9ce11d8474d4c2c1e6a7de8610195ce
- frontend/src/routes/teach.courses.$courseId_.lessons.$lessonId.tsx: 28825ee01064c6f437369a109fc0d0e4f344ebb5ba07a34643f7a861ef03ff3d
ФАКТ: фінальний native checkpoint → OK, exit 0, ID b997b10e98526e349cc48feaa37ac714-2-50495a83bd74. Поточні root/session підтверджені host identity; implementation готова до незалежного review.

## Координатор (Courses13) — після run 1
ФАКТ: звірив shared-дифи — admin-document-queries generic→union (рантайм-ключі незмінні), admin-course-repository чистий extract+rename мапперів, admin/index +2 export, teach.ts +detail(getTeachCourse). Admin-рантайм збережено.
ФАКТ: повний frontend test виявив 1 fail — census R24 (route-guards) не містив нового teach editor route (правило «хто додає глобально-лічене — прибирає»). Координатор додав запис у expected-список.
ФАКТ: impact gate GREEN — format:check=0, lint=0 (arch 230), typecheck=0, test 837/837 0 fail.
ДАЛІ (run 2, Codex resume): component-тести TE5 — T13 empty states (course 0 lessons / lesson без draft) + lesson-list рендерить lessons з links у editor route.

## Run 2 — component extraction + tests
ДАЛІ: звірити branch/WIP та CourseRoster conventions, винести loaded list і покрити observable output.
ФАКТ: feat/teach-editor; coordinator R24 зміна присутня, попередній GREEN зафіксований. Прочитано CourseRoster.tsx/test: testing-library render/screen, rstest; переклади налаштовані глобально. Pending/error залишаються в route; extracted component отримуватиме CourseTree data без Query.
ДАЛІ: перевірити test setup і наявний спосіб render TanStack Link, типи CourseTree.
ФАКТ: LearnerOverview.test.tsx у тому самому teach модулі вже має мінімальний memory router для реального TanStack Link + I18nProvider; його використаємо як link harness, не рендер production route. Rstest dom project = happy-dom + test-setup.ts.
РІШЕННЯ: Run 2 small extraction, READY: props courseId/title/modules типізовані через CourseTree; JSX переноситься без зміни output. Чотири observable cases: grouped links, прихований порожній module, усі modules без lessons, нуль modules. Admin/document/R24 та сторонні файли не редагувати.
ДАЛІ: перенести JSX, export barrel, додати component tests.
ФАКТ: extracted TeachLessonList + public export; route лише делегує loaded branch. Збережено course title, module h3/aria-labelledby, list/link classes/params і existing empty key. Додано 4 component tests з real Link href, grouping і двома T13 empty inputs; existing tests не редагувалися.
ДАЛІ: format → typecheck → lint → requested teach/R24 tests.
ФАКТ: frontend/ npm run format → PASS exit 0 (378 files); npm run typecheck → PASS exit 0; npm run lint → PASS exit 0 (architecture 231 files).
ФАКТ: npm run test -- src/modules/teach src/routes/route-guards.test.tsx → FAIL exit 1 до виконання тестів: listen EPERM ::1:3000, sandbox listener restriction.
ДАЛІ: повторити точну test command з escalation, без зміни тестів/config.
ФАКТ: повтор npm run test -- src/modules/teach src/routes/route-guards.test.tsx → PASS, exit 0: 6 files, 70/70 tests, 0 skipped; TeachLessonList 4/4 і R24 census PASS. Перша спроба EPERM була environment failure, тести тоді не виконувалися.
ФАКТ: фінальний огляд підтверджує однаковий loaded markup; pending/error/refetch лишилися в route. git diff --check → PASS (без повідомлень), branch feat/teach-editor. Run 2 змінив лише TeachLessonList.tsx/test.tsx, modules/teach/index.ts, teach course route та цей журнал. Existing tests/R24, admin/document, .gitignore/security не редаговано в run 2.
ФАКТ: обсяг T13 цього run — empty course lesson list; empty-draft document editor тут не тестувався й не змінювався за явним scope. Full make check/integration/e2e не запускалися; git add/commit/push не виконувалися.
ДАЛІ: передати coordinator/reviewer extraction + 4 passing component tests у working tree.
ФАКТ: Run 2 final SHA-256 snapshot (R24 read-only):
- frontend/src/modules/teach/ui/TeachLessonList.tsx: 7eb3d703730973479fd87daace386100102c0a8f0884d27bf026acd2f13457ac
- frontend/src/modules/teach/ui/TeachLessonList.test.tsx: 60dc1c5d395df5ffde5f405ef2d54171d5d832c1ef13c6ca594eca3710b05a6e
- frontend/src/modules/teach/index.ts: 12cae46564808eb3ac6f9f32353166c0c0b63b4b9deb84093f2d90f5cd5f7bc5
- frontend/src/routes/teach.courses.$courseId.tsx: bb5e309db9b8795887c3401d3cfcea38186e5cfdf15dd60ba0a8805fc65b47c7
- frontend/src/routes/route-guards.test.tsx: 81a3d06a62dc791e918aa514651244ece284044624fe87f14fa451908fb84614
ФАКТ: Run 2 native checkpoint → OK, exit 0, b997b10e98526e349cc48feaa37ac714-3-b21390ffb134. Готово до незалежного review.
