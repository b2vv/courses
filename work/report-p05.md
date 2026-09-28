# P05 — звіт

## Результат

Підозра підтвердилася. Відмова доступу до чужої чернетки тепер показується як названий стан
«немає доступу», з поясненням і посиланням до каталогу; кнопки/посилання «Спробувати ще раз» у
цьому стані немає. Інші помилки дочірнього маршруту, зокрема збій перевірки сесії, як і раніше
делегуються до `ProtectedRouteFailure`.

## Перевірений ланцюжок

1. `backend/internal/course/application/learning.go:26-31` читає курс і повертає
   `domain.ErrCourseForbidden`, якщо студент не має права бачити його.
2. `backend/internal/course/adapters/http/learning_handler.go:104-107` мапить цю помилку на
   HTTP `403` з кодом `course_forbidden`.
3. `frontend/src/adapters/api/playback-repository.ts:52-59` дозволяє для `403` лише
   `course_forbidden` і зберігає цей код у доменному результаті.
4. `frontend/src/app/composition/learn-queries.ts:24-26` перетворює невдалий результат на
   `LearnQueryError`.
5. `frontend/src/routes/learn.$.tsx:20-24` або `:35-39` викликає `ensureRouteQuery`; помилка
   виникала до монтування компонента. Раніше вона піднімалася до `learn.tsx:19` і
   `ProtectedRouteFailure`, тобто студент бачив `login.failure`.

## Зміни

- `frontend/src/presentation/identity/ProtectedRouteStates.tsx:22-37` додано
  `LearnRouteFailure`: тільки `LearnQueryError` з кодом `course_forbidden` рендерить
  `LearnAccessDenied`; усі інші помилки лишаються у `ProtectedRouteFailure`.
- `frontend/src/routes/learn.$.tsx:44` підключено локальний `errorComponent` дочірнього маршруту.
- Додано тексти стану в `frontend/src/app/providers/locales/en.ts` і `uk.ts`.
- `frontend/src/routes/route-guards.test.tsx` додано unit-регресію R11 із перевіркою відсутності
  retry та наявності посилання до каталогу. Наявний тест G09 і далі доводить retryable стан
  `ProtectedRouteFailure` для недоступної сесії.
- Додано `frontend/e2e/learn-access.spec.ts`: через API адміна створюється run-scoped чернетка
  з уроком, сторінка відкривається зі збереженою студентською сесією, після тесту курс видаляється.

## Перевірки

- `cd frontend && npm run format:check` — PASS.
- `cd frontend && npm run typecheck` — PASS.
- `cd frontend && npm run lint` — PASS; architecture boundaries: 114 files.
- `cd frontend && npx playwright test e2e/learn-access.spec.ts --list` — PASS; новий R11
  присутній у списку.
- `cd frontend && npm test -- src/routes/route-guards.test.tsx` — НЕ ЗАПУСТИВСЯ: пісочниця
  заборонила bind на `::1:3000` (`listen EPERM`). Це не позначено як passed.
- Повний e2e — НЕ ЗАПУСКАВСЯ: робоча модель забороняє Codex запускати Docker/e2e у цій
  пісочниці; потрібен прогін reviewing session.

Git не змінювався. Зміни коду обмежені `frontend/`; цей звіт записано у запитаний шлях
`work/report-p05.md`.
