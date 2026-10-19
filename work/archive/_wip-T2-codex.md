ДАЛІ: Прочитати точні вимоги T2/S2 та ib1/ib2, перевірити редактор і реалізувати T2 без тестів та коміту.
ФАКТ: Прочитано T2 Impl, S2 та §7 ib1/ib2; гілка feat/editor-polish, tracked-змін перед реалізацією немає. Редактор має стале замикання token і upload-apply; глобальний upload.isPending потрапляє в усі зображення.
РІШЕННЯ: 2026-10-15, Codex — виконати визначений T2: функційні token/sections-апдейти, upload-дельта за section.id + column.id, Toast з role=status. Перевірка цього прогону — format і tsc; тести окремим прогоном за вказівкою користувача.
ДАЛІ: Змінити редактор і uk/en, перевірити diff та запустити погоджені перевірки.
ФАКТ: Реалізовано uploadingKey з умовним очищенням, функційні token/sections-апдейти та applyImage(sectionId,columnId,imageId,contentType), що змінює лише зображення поточного стану. Додано Toast та uk/en success-повідомлення.
РІШЕННЯ: Token fallback використовує гарантовано завантажений draft (cur ?? draft); коли cur=null, draft є base, отже ib1 синхронно сидить token без залежності від refetch.
ДАЛІ: Перевірити форматування, TypeScript та фінальний diff.
ФАКТ: npm run format --prefix frontend — exit 0. Наявний Toast рендерить output (неявна role=status), помилки лишено в ErrorState; спільний UI не змінювався.
ДАЛІ: npx tsc --noEmit та перевірка меж diff.
ФАКТ: npx tsc --noEmit — exit 2: src/presentation/ui/ui.test.tsx(44,19), TS2379: Element не сумісний з HTMLElement (exactOptionalPropertyTypes). У файлах T2 діагностик немає. ui.test.tsx не змінений від HEAD; git diff --check — exit 0.
РІШЕННЯ: Не змінювати ui.test.tsx: користувач прямо обмежив цей прогін реалізацією T2 і заборонив роботу над тестами. TypeScript не позначаю зеленим.
ДАЛІ: Окремим тестовим прогоном усунути typecheck-блокер ui.test.tsx:44 та перевірити T2 unit/e2e (зокрема alert count 0). Реалізація T2 завершена, коміту немає.
ДАЛІ: Прогін тестів T2 — додати чотири unit-сценарії лише в LessonDocumentEditor.test.tsx; production-код не змінювати.
ФАКТ: Наявний setup використовує справжні createAdminDocumentQueries/useMutation поверх in-memory DocumentAuthoringRepository. Перед прогоном є зміни реалізації у редакторі та uk/en; зберігаю їх без правок.
РІШЕННЯ: Розширити setup опційними repository overrides, керувати pending save/upload через deferred promises; перевіряти DOM у секціях за alt та success-status за текстом.
ДАЛІ: Додати сценарії a–d і запустити format → tsc → повний npm run test.
ФАКТ: Додано чотири сценарії a–d з реальними Query-мутаціями та керованими repository promises; B1 змінює alt після фактичного старту save і перевіряє його після success. Upload-тести завершують pending promises; паралельний сценарій перевіряє порядок старту/завершення і очищення індикатора B після завершення B.
ДАЛІ: Форматування та перевірки.
ФАКТ: Перший виклик npm run format помилково виконано з кореня репо (ENOENT package.json); повторено з frontend/, exit 0.
ДАЛІ: TypeScript.
ФАКТ: TypeScript виявив неповну нову upload-фікстуру (бракувало width/height). Додано обов'язкові розміри 1×1 у тесті; production-код без змін.
ДАЛІ: Повторити format і tsc, потім повний unit-набір.
ФАКТ: Повторні npm run format та npx tsc --noEmit — exit 0.
ДАЛІ: Повний npm run test у frontend/.
ФАКТ: npm run test не стартував у sandbox: listen EPERM ::1:3000 (Rstest). Це обмеження відкриття локального порту, не результат тестів.
ДАЛІ: Повторити той самий повний unit-прогін із дозволом на локальний порт.
ФАКТ: Повний npm run test після дозволу локального порту — exit 0: 84 файли, 736/736 passed, 0 failed, 0 skipped, 0 todo. npm run format і npx tsc --noEmit зелені; git diff --check exit 0. Лог: /tmp/courses-T2-unit.log.
ФАКТ: Усі чотири нові сценарії пройшли. Змінено лише тестовий файл у цьому прогоні (окрім append робочого журналу); наявні production-зміни T2 збережено. Коміту немає.
РІШЕННЯ: Межа доказу d — тест доводить спостережувану поведінку справжнього useMutation: після завершення repository-запиту A індикатор B лишається, після B зникає. Це не ізольований доказ виконання guard-гілки onSettled для A: consecutive mutate callbacks керуються observer TanStack Query. Моки useMutation чи штучне примусове виконання callback не додавалися.
ДАЛІ: Цей unit-прогін завершено. Повний e2e-гейт T2 лишається окремою перевіркою; у цьому прогоні e2e не запускався.
