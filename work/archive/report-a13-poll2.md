# Report A13 / крок 4d — три правки після e2e 4b/4c

## Зроблено

- У `admin-draft-queries.ts` маркер перегенерації зберігає `updatedAt` чернетки з кешу.
  Старий terminal response з тим самим значенням віддається UI як синтетичний `pending` і
  продовжує polling; маркер знімається лише після нового `updatedAt`. Стеля 15 хвилин збережена.
- У `admin-draft-queries.test.ts` додано сценарій `ready/T1 → ready/T1 → running → ready/T2`
  з перевіркою зупинки polling після нової чернетки.
- У `admin-video-queries.ts` збережено snapshot останньої `ready`-генерації в `QueryClient`.
  Після переходу конкретного уроку на нову `ready`-генерацію інвалідовується лише його
  `lessonDraftQueryKey`; проміжний `processing` не інвалідовує чернетку.
- У `admin-video-queries.test.ts` перевірено `ready → processing → ready` і відсутність впливу
  на інший урок.
- У `frontend/e2e/asr.spec.ts` два селектори кнопки завантаження чернетки зроблено точними,
  щоб не матчити кнопку-відкривач із довшою назвою.

## Перевірки

- `npm run format:check` — пройшов.
- `npm run lint` — пройшов.
- `npm run typecheck` — пройшов.
- `npx playwright test e2e/asr.spec.ts --list` — пройшов, перелічено 11 тестів.
- Цільові Rstest-тести — **не запущені**: sandbox відмовив у відкритті listener-а, `listen EPERM
  ::1:3000`. Це не позначено як pass.
- Повний `npm test` не запускався з тієї самої причини; його має прогнати reviewing-сесія.
- Docker/e2e execution не запускався згідно з `work/codex-working-model.md`; `--list` перевірив
  лише discovery.

## Scope / review

Змінені лише фронтенд-файли, перелічені в брифі, їхні unit-тести та цей звіт. Бекенд,
`openapi.yaml`, draft-адаптер, `work/tasks.md`, phase report і git не змінювалися. Наявні до
початку роботи dirty-файли збережені.

Ручний standards/spec review не виявив додаткових відхилень або scope creep; git-based review
не виконувався, бо бриф прямо забороняє чіпати git.

## Правка після рев’ю

Рев’ю виявило помилку у фікстурі тесту `keeps polling and exposes pending after a successful
regenerate until ready`: третє читання моделювало `running` зі старим `updatedAt`. Виправлено
лише тестову фікстуру на `ready/T1 → ready/T1 → running/T2 → ready/T3`, де `T2` —
`2026-09-24T12:00:00.500Z`, а фінальне очікування лишається на `T3` —
`2026-09-24T12:00:01.000Z`. Це відповідає оновленню `updated_at` під час Claim.
