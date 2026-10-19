ДАЛІ: Прочитати briefing, T1 у tasks.md, S1 і §7 plan.md та факти коду; перевірити гілку і поточні зміни перед реалізацією.
ФАКТ: Гілка feat/editor-polish, tracked-змін перед роботою немає; наявні untracked-файли зберігаю. Загальний вивід джерел обрізано, читаю цільові секції окремо.
ФАКТ: T1/S1 підтверджують опційний drop, MIME jpeg/png/webp та явний aria-label input. Поточний FileUpload просто передає props у FileTrigger; вихідних src RAC у node_modules немає, перевіряю dist.
ФАКТ: §7 impl-ноти стосуються інших тасків; їх не реалізую. Знайдено dist/private/FileTrigger.mjs для перевірки передачі aria-label.
ФАКТ: Реалізовано T1 у п’яти файлах: опційний DropZone, перший MIME-прийнятний файл, ImageField через FileUpload, scoped CSS і uk/en dropHint. Alt та preview збережено.
ФАКТ: Встановлений RAC FileTrigger передає rest через filterDOMProps(global:true) безпосередньо на input; aria-label із ImageField проходить через ...props. Без onDrop повертається попередній FileTrigger+Button.
ДАЛІ: npm run format, потім npx tsc --noEmit; перевірити diff на межі T1.
ФАКТ: Перший запуск format помилково з кореня завершився ENOENT package.json; повторюю у frontend/.
ФАКТ: npm run format у frontend/ пройшов (exit 0, 341 файл).
ФАКТ: npx tsc --noEmit у frontend/ пройшов (exit 0). Тестів не писав і не запускав; e2e-доказ aria-label та відео-регресії належить окремому прогону.
ФАКТ: Фінальний diff — лише 5 дозволених T1 frontend-файлів (53 додано, 16 прибрано); git diff --check зелений. Форматування сторонніх tracked-змін не внесло. Reader/backend/контракти/домен і тести не змінені. Коміт не створено.
ДАЛІ: Передати ведучому T1 impl для окремого прогону тестів і повного гейту; T1 acceptance ще не оголошено закритим.
ДАЛІ: Прогін тестів T1 — прочитати поточні ui.test.tsx і FileUpload, додати лише unit-тести; production-код не змінювати. Потім npm run format і npm run test.
ФАКТ: У робочому дереві вже є зміни ведучого у plan/e2e/production; зберігаю їх. ui.test.tsx використовує rstest, testing-library та прямі expect без jest-dom.
РІШЕННЯ: Враховано уточнення користувача: тестую accessible name видимої Button, не aria-label input; production-код не змінюю.
ФАКТ: Додано 4 unit-тести: PNG drop із тотожністю File, dropLabel і hidden input; відео без drop-зони; accessible name кнопки; PDF ігнорується. Використано fireEvent.drop через реальний RAC, без mock production-компонента.
ДАЛІ: Форматування та повний npm run test.
ФАКТ: npm run format пройшов (exit 0).
ФАКТ: npm run test не стартував через sandbox listen EPERM ::1:3000, це не падіння тесту. Повторюю той самий повний набір із дозволом на локальний listener.
ФАКТ: Повторний повний npm run test завершився exit 0; лог /tmp/courses-T1-unit.log: No test failures reported, all tests passed. git diff --check зелений. Production-файли не редагував; коміт не створював.
ФАКТ: Підсумок повного набору: 732 passedTests, 0 failedTests, 0 failedFiles; усі 4 нові тести пройшли.
ДАЛІ: Передати ведучому тестовий diff ui.test.tsx; production-виправлень тести не потребують.
