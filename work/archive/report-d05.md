2026-09-29 — Початок виконання brief-d05. Спершу фіксую контекст і вимоги, далі виконую роботу зрізами з перевірками; коміт не створюю.
ФАКТ: Бриф D05 вимагає лише новий `frontend/e2e/discussion.spec.ts`, не чіпати інші e2e-файли, не змінювати продукт без доказу дефекту; sandbox e2e не запускає через Docker/bind.
РІШЕННЯ: Працюю за шаблонами `transcript.spec.ts`, `learn-access.spec.ts`, `accessibility.spec.ts` і використовую існуючі seed/auth helper-и; результати недоступного локального e2e позначатиму саме як не виконані.
ДАЛІ: Прочитати пов’язані e2e, API/фронтові селектори та seed-дані, потім скласти run-scoped сценарії K19–K23.
ФАКТ: Додано лише `frontend/e2e/discussion.spec.ts`: run-scoped published/draft курси, реальний upload двосекундного fixture, K19 UI-публікація/seek/reply, K20 edit/remove для автора/власника/admin, K21 UI+API access denial, K22 keyboard/44px/axe light+dark/200% zoom; cleanup архівує published і видаляє draft.
РІШЕННЯ: Для K19 використано `instructor` як другого наявного акаунта (другого student у seed немає); курс належить instructor, що одночасно дозволяє перевірити власницьку модерацію K20.
ДАЛІ: Запустити форматування, typecheck, lint і Playwright `--list`; виправити лише помилки нового тесту.
ФАКТ: У K20 корінь перед owner-remove тепер має відповідь, тому перевіряється саме placeholder видаленого кореня; cleanup реєструє курс одразу після створення, архівує опубліковані й видаляє невидані навіть при ранньому падінні.
ФАКТ: Локатори панелі враховують `uk-UA` (`Discussion|Обговорення`). K23 перевіряє `panel.getByRole('status')` під час навмисно затриманого завантаження, а не шукає глобальний статус за позицією.
ФАКТ: Повторні гейти: `npm run format:check` — PASS; `npm run typecheck` — PASS; `npm run lint` + architecture (118 файлів) — PASS; `npx playwright test e2e/discussion.spec.ts --list` — PASS, 9 тестів у 4 файлах із setup/teardown; `git diff --check` — PASS.
ФАКТ: Повний Playwright не запускався: за умовою brief D05 sandbox не стартує Docker/bind. Тому K19–K23 не оголошені пройденими; вони лише зареєстровані та статично перевірені.
РІШЕННЯ: Не чіпав інші e2e-файли, продукт і попередні PNG-зміни; коміт не створював.
ДАЛІ: Передача на стендовий запуск Courses: `npx playwright test e2e/discussion.spec.ts` і повний `make check`; після цього заповнити `docs/discussion/report.md` у D06.
2026-09-29 — Courses підтвердив стендовий результат: K21 PASS; K19/K20 timeout і K22 FAIL через ASCII-апостроф у локаторах української кнопки при UA locale. Виправляю регекси на `[’']` і повторюю локальні гейти; стендовий e2e повторно запускає Courses.
ФАКТ: У `frontend/e2e/discussion.spec.ts` виправлено всі три українські локатори з апострофом: `Прив[’']язати до поточного моменту` і `Прив[’']язано до`. Додатковий пошук кириличних слів із ASCII-апострофом збігів не знайшов.
ФАКТ: Після виправлення PASS: `npm run format:check`, `npm run typecheck`, `npm run lint` + architecture (118 файлів), `npx playwright test e2e/discussion.spec.ts --list` (9 тестів у 4 файлах із setup/teardown).
ДАЛІ: Courses повторно запускає стендовий `discussion.spec.ts`; коміт не створювати.
ФАКТ: Product fix у `DiscussionPanel.tsx`: timestamp без `onSeek` тепер рендериться як доступний `<time>` з тим самим `MM:SS`, `dateTime=PT…S` і `aria-label`; із `onSeek` лишається кнопкою.
ФАКТ: Додано unit-сценарій «timestamp text, not button»; `DiscussionPanel.test.tsx` PASS 13/13.
ФАКТ: K19 тепер асертить `00:01` як текст до Play, натискає `Відтворити`, потім клікає seek-кнопку. K20 редагує через стабільну `.discussion-edit-form`, після save перелоковує оновлену картку.
ФАКТ: Після змін PASS: `format:check`, `typecheck`, `lint` + architecture (118 файлів), `npx playwright test e2e/discussion.spec.ts --list` (9 тестів); стендовий e2e не запускався.
РІШЕННЯ: Змінено лише DiscussionPanel, його unit-тест, discussion e2e і цей звіт; коміт не створював.
2026-09-29 — Courses підтвердив K22 PASS після апостроф-фіксу; K19/K20 мають окремі причини: DiscussionPanel ховає мітку без onSeek, а K20 втрачає картку після переходу в textarea. Виправляю продукт і тести мінімальними змінами; unit/e2e запускає Courses.
ФАКТ: Фінальний стан після цього слайсу: unit 13/13 і всі запитані локальні гейти PASS; `git diff --check` PASS; стендовий e2e лишається за Courses; коміт не створений.
