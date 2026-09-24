# Report A13 / e2e ASR

Дата: 2026-09-24

## Зроблено

Додано `frontend/e2e/asr.spec.ts` з одним `describe` на збереженому admin state і run-scoped API seed/cleanup:

- C13: silent fixture завершується станом `no_speech`, без `Draft ready`.
- C03: перевіряється спостережений проміжний `pending`/`running`, потім `ready`.
- C11: перевіряються рівно `Auto`, `English`, `Українська`; у запиті `auto` немає `language`, у запиті `English` є `language: "en"`. Тіло звужується з `unknown`, без `as { ... }`.
- C06: після другого реального upload UI показує застарілість draft.
- C17: незбережений текст лишається після завершення ASR; після явного load draft у редактор потрапляє непорожній текст чернетки.
- C18: незбережений завантажений draft після regeneration і reload пропонується як stale recovery.

Відео завантажуються через UI `input[type=file]`; `page.route`-моків немає. Курс і початковий транскрипт створюються через API, а чернетки курсів видаляються в `afterAll`.

## Перевірки

- `npm run format:check` — пройшов.
- `npm run lint` — пройшов; architecture boundaries перевірені.
- `npm run typecheck` — пройшов.
- `npx playwright test e2e/asr.spec.ts --list` — пройшов: 6 ASR-тестів знайдено.
- Сам e2e — **не запускався**: за умовами брифу цей sandbox не має Docker; запуск проти Compose виконує Courses 2.

## Discussion

Гілка названої причини після падіння навмисно не додавалась: вона покрита unit/integration/HTTP-контрактами, як визначено брифом.

## Обмеження

Git не читався для зміни стану і не змінювався; `work/tasks.md`, `docs/asr/report.md`, product source, backend, OpenAPI та generated-файли не чіпались.
