ДАЛІ: Прочитати T3/S3, поточний компонент, TextField та локалі; реалізувати лише D2a, потім форматування і tsc без тестів та коміту.

ФАКТ: Прочитано брифінг і work/tasks.md; джерелом поточного зрізу лишаються editor-polish T3/S3.
ДАЛІ: Звірити API TextField, стан гілки та наявні toolbar-тести перед правкою.

ФАКТ: Гілка feat/editor-polish; tracked-змін перед роботою немає. TextField підтримує ref до input та Aria.TextFieldProps; наявний тест перевіряє aria-pressed кнопки Посилання.
РІШЕННЯ: Для T3 використати TextField з type=url, isInvalid=linkError та локалізованим errorMessage; наявні toolbar-кнопки й aria-pressed зберегти.
ДАЛІ: Внести правки компонента та двох локалей.

ФАКТ: Реалізовано T3: invalid URL повертає фокус без зміни лінку/URL/відкритості поля; onChange скидає помилку; додано TextField, явне видалення та uk/en рядки. Наявні aria-pressed не змінено.
ДАЛІ: Запустити npm run format у frontend, потім npx tsc --noEmit.

ФАКТ: npm run format у frontend завершився успішно (exit 0).
ДАЛІ: Обов’язкова перевірка TypeScript.

ФАКТ: Перший tsc виявив TS2375: exactOptionalPropertyTypes не дозволяє явний undefined для errorMessage?: string. Diff обмежений трьома цільовими файлами, diff --check чистий.
РІШЕННЯ: Передавати errorMessage умовним spread лише при помилці, без зміни спільного TextField.
ДАЛІ: Виправити проп та повторити format і tsc.

ФАКТ: Умовний errorMessage виправлено; повторний npm run format — exit 0.
ДАЛІ: Повторити npx tsc --noEmit.

ФАКТ: Фінальний npx tsc --noEmit — exit 0; git diff --check — exit 0. Tracked diff містить лише LessonDocumentRichText.tsx та locales uk/en. Toolbar aria-pressed збережено за переглядом diff. Тести не писалися й не запускалися; коміту немає.
ДАЛІ: Окремий тестовий прогін T3 за tasks.md; реалізація цього прогону завершена.

ДАЛІ: Тестовий прогін T3: перевірити наявні тести й додати сценарії invalid/valid/remove без змін production-коду.

ФАКТ: RichText тестується через LessonDocumentEditor.test.tsx; середовище happy-dom, реальний Tiptap. Наявні production-зміни T3 вже у diff і залишаються недоторканими.
РІШЕННЯ: Виділення тексту задавати командою реального Tiptap (DOM editor перевіряється instanceof); дії виконувати кнопками форми, перевіряти DOM, стан документа та payload збереження. Це unit-доказ без імітації браузерного виділення мишею.
ДАЛІ: Додати три сценарії в наявний файл тестів.

ФАКТ: Додано три тести: invalid з попереднім лінком і переведенням фокусу з кнопки; invalid→valid зі скиданням помилки й збереженням href; явне видалення зі збереженням тексту. У всіх перевіряється payload saveDraft.
ДАЛІ: Форматування, обов’язковий tsc, повний npm run test.

ФАКТ: npm run format — exit 0.
ДАЛІ: npx tsc --noEmit.

ФАКТ: tsc виявив помилку лише в новому тесті: ByRoleOptions не має exact (рядковий name уже точний).
ДАЛІ: Прибрати зайвий параметр тесту, повторити format і tsc.

ФАКТ: Зайвий exact прибрано; format — exit 0.
ДАЛІ: Повторний tsc.

ФАКТ: npx tsc --noEmit — exit 0.
ДАЛІ: Повний npm run test.

ФАКТ: npm run test заблоковано sandbox до запуску тестів: listen EPERM ::1:3000. Це обмеження середовища, не результат тестів.
ДАЛІ: Повторити ту саму команду з дозволом на локальний сервер rstest.

ФАКТ: Повний npm run test завершився exit 1: 84 файли, 739 тестів, 738 passed / 1 failed / 0 skipped. Нові сценарії invalid та invalid→valid пройшли; remove падає на появі кнопки, до натискання. Лог: /tmp/courses-t3-unit.log. Форматування, tsc та git diff --check зелені.
ФАКТ: Реальний editor.isActive('link') після setTextSelection({from:1,to:5}) = true, але findByRole('button', {name:'Видалити посилання'}) не знаходить кнопку. У встановленому @tiptap/react/src/useEditor.ts:359–367 selector повертає null при shouldRerenderOnTransaction === undefined, тобто зміна selection сама не оновлює React toolbar. Компонент не задає цю опцію й не підписує toolbar на editor state. Інші два тести мають локальні state-зміни через відкриття поля, які спричиняють рендер.
ПИТАННЯ: Потрібна окрема production-правка реактивності toolbar при зміні виділення (наприклад, підписка на editor state або shouldRerenderOnTransaction), щоб кнопка видалення з’являлася при активному лінку. Чи дозволяєте її наступним прогоном? За прямою умовою власника цей прогін зупинено; production-код не змінювався, failing-тест збережено без штучного додаткового рендеру. Перевірки зняття mark і payload у сценарії remove ще не виконалися через цей блокер.
ДАЛІ: Після окремо дозволеної production-правки повторити тестовий гейт. Коміту немає. Обмеження доказу: selection встановлюється через реальну команду Tiptap у happy-dom, не мишею в браузері; зміни лінків виконуються через UI.
