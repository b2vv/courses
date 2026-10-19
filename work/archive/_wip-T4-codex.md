ДАЛІ: Прочитати джерела T4/S4/ib3, перевірити поточну реалізацію та внести лише T4; потім format і tsc без запуску тестів та без коміту.

ФАКТ: Гілка feat/editor-polish, tracked-змін на старті немає; T4/S4 та ib3 вимагають явних цілей фокусу й повторного оголошення межі. Наявні untracked-файли збережено.
ДАЛІ: Звірити DOM FileUpload/редактора, додати pendingFocus і стабільний live-region з keyed-вмістом.

РІШЕННЯ: 2026-10-15, Codex — для відкладеного Tiptap textbox ефект очікує childList через MutationObserver, з disconnect після фокусу/cleanup; перевірка — add text/task у наступному тестовому прогоні.
ФАКТ: Додано pendingFocus, data-section-id, якір палітри, status із keyed-вмістом, uk/en і scoped sr-only CSS; e2e delete-асерт додано після незмінених reorder/axe перевірок.
ДАЛІ: Перевірити diff, запустити format і tsc.

ФАКТ: Diff обмежений п’ятьма файлами T4; наявні e2e-асерти збережені. npm run format — exit 0 (341 файлів перевірено/відформатовано).
ДАЛІ: Обов’язковий npx tsc --noEmit та фінальна перевірка diff.

ФАКТ: npx tsc --noEmit — exit 0. git diff --check — exit 0. Форматування не внесло змін поза п’ятьма frontend-файлами T4.
ФАКТ: Unit-тестів не додано, unit/e2e не запускалися; комітів не створено. Перевірка браузерного фокусу й реального озвучення ще не виконана.
ДАЛІ: Окремий тестовий прогін T4 (unit a–e, повторне оголошення ib3, повний e2e); стан реалізації й обмеження зафіксувати у report.md.

ФАКТ: Створено docs/editor-polish/report.md з результатами лише цього реалізаційного прогону T4, обмеженнями та наступним тестовим кроком.

ДАЛІ: Тестовий прогін T4 — додати unit a–e та ib3 без змін production-коду, виконати format, tsc, повний npm run test. Якщо потрібне виправлення production — зупинитися з ПИТАННЯ.

ФАКТ: Прочитано наявний harness (happy-dom + реальний Tiptap), T4 та production-фокус. Додано 10 тестів: add усіх 5 типів, delete середньої/першої/єдиної, reorder усередину, повтор межі ib3. Фокус перевіряється через activeElement після waitFor, без моків focus.
РІШЕННЯ: ib3 перевіряється заміною DOM-вузла оголошення при стабільному status і незмінному тексті; реальне озвучення скрінрідера цим не доводиться.
ДАЛІ: format → tsc → npm run test; production-код не змінено.

ФАКТ: npm run format — exit 0.
ДАЛІ: Перевірка TypeScript.

ФАКТ: npx tsc --noEmit — exit 0.
ДАЛІ: Повний npm run test.

ФАКТ: npm run test у sandbox не стартував: listen EPERM ::1:3000 (Rstest server), жоден тест не виконався. Це обмеження запуску, не дефект production.
ДАЛІ: Повторити ту саму команду з дозволом локального listener поза sandbox.

ФАКТ: Повний npm run test виконався поза sandbox: 749 тестів, 748 passed, 1 failed, 0 skipped (84 файли). Єдина помилка — новий «adding image focuses its designated button», LessonDocumentEditor.test.tsx:147. Решта 9 нових тестів, включно з text/task activeElement, delete та ib3, пройшли без послаблення асертів.
ФАКТ: ADD image фокусує приховану службову кнопку aria-label="DropZone", а не видиму «Завантажити зображення»: селектор production `.lesson-editor-image button` обирає першу кнопку DropZone. Це дефект адресації, не обмеження happy-dom; обидві кнопки присутні в DOM, activeElement точно дорівнює неправильній. Лог: /tmp/courses-t4-unit.log. Format, tsc та git diff --check — exit 0.
ПИТАННЯ: Чи дозволяєте окремим кроком виправити production-селектор ADD image, щоб він адресував видиму кнопку FileUpload, оминаючи приховану кнопку DropZone, та повторити перевірки? Зупинено за прямою інструкцією користувача. Production-код не змінено в цьому прогоні; червоний тест збережено, коміту немає.
