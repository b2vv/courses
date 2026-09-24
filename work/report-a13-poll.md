# Report A13 / poll

## Зроблено

- Додано per-lesson marker у TanStack Query cache для задачі, прийнятої після успішного
  `regenerate` (202).
- `already_running` (409) також позначає задачу як таку, що виконується, інвалідовує лише
  query відповідного уроку й запускає polling.
- Відповідь сервера `state: null` під час активного marker відображається як `pending`, тому
  автор бачить `Queued for recognition` і polling не вимикається до `ready`, `no_speech` або
  `failed`.
- Додано названу стелю `draftTaskTimeoutMs` — 300 секунд із коментарем про відповідність
  мінімальному backend task timeout. Після неї marker очищається, synthetic pending замінюється
  на звичайний `null`, polling зупиняється.
- Інші уроки мають окремі Query keys і не впливають один на одного.

Змінені файли:

- `frontend/src/app/composition/admin-draft-queries.ts`
- `frontend/src/app/composition/admin-draft-queries.test.ts`
- `frontend/src/routes/admin.courses.$courseId.tsx`

## Перевірки

- `npm run format:check` — пройшов.
- `npm run lint` — пройшов; architecture boundaries — пройдені.
- `npm run typecheck` — пройшов.
- `npm test` — пройшов: 475/475 тестів, 54 test files.
- Нові boundary-тести — 3/3: 202 → pending → ready, 409 `already_running` ізоляція уроків,
  timeout і повернення до `null`.
- Backend, e2e та ordered gate не запускалися: вони поза scope цього frontend-брифу та не
  доступні в обмеженнях робочої сесії.

## Правка після рев’ю

Стелю очікування client-side marker збільшено з 5 до 15 хвилин. Вона обмежує лише очікування
задачі в черзі до `Claim`; після `Claim` сервер повертає `running`, і polling продовжується за
станом сервера. 15 хвилин потрібні через `Concurrency 1` у `asr-worker`: урок може чекати за
іншим довгим розпізнаванням. Стеля захищає від задачі, втраченої, наприклад, під час очищення
Redis.

Після цієї правки формат frontend-файлів перевірено. Повторний targeted test у sandbox не
запустився через `listen EPERM: operation not permitted ::1:3000`; попередній повний прогін
до цієї механічної зміни був зеленим (475/475).

## Відома межа

Автопостановка через `EnqueuePending` не створює client-side marker. Якщо сторінка вже відкрита,
автоматично поставлена задача стане видимою лише після reload або refetch on focus; цей fix
покриває постановку, ініційовану автором.

Git не змінювався; `work/tasks.md` і phase report не редагувалися.
