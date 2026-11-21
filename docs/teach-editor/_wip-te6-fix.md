# TE6 publish-bug fix (B1 + guard) — Codex working file

РОБОЧИЙ ФАЙЛ. Дописуй ПІСЛЯ КОЖНОГО КРОКУ (append, не rewrite).
Перший запис — рядок `ДАЛІ:` до першого інструменту. Далі — ФАКТ / РІШЕННЯ / ДАЛІ / ПИТАННЯ.
Фінальна відповідь = стислий переказ файлу.

## Задача
Виправити пре-існуючий баг спільного (admin+teach) publish-шляху frontend: сервер віддає 200 з
ПОРОЖНІМ тілом (контракт навмисний: `Empty:true`, `document_handler.go:132` голий
`WriteHeader(200)`, generated `publishAdminLessonDocumentResponse200.data: void`), а клієнтський
репозиторій `publish` парсить `lessonDocumentUpdatedSchema` на `undefined` → `invalid_request` →
редактор показує ХИБНУ «Не вдалося зберегти» ПІСЛЯ успішної публікації.

Рішення = **B1 (frontend-only) + guard**. НЕ чіпати бекенд (контракт порожнього тіла навмисний).

## Точний обсяг (змінити САМЕ це, нічого зайвого)
1. **Порт** `frontend/src/modules/document/ports/DocumentAuthoringRepository.ts:23-28`:
   `publish(...)` повертає `Promise<DocumentAuthoringResult<void>>` (БУЛО `<{ readonly updatedAt: string }>`).
   `saveDraft` лишається з `{updatedAt}` — НЕ чіпати.

2. **Repo** `frontend/src/modules/document/adapters/admin-document-repository.ts`, метод `publish`
   (рядки ~159-177): прибрати `lessonDocumentUpdatedSchema.safeParse(outcome.data)` і повернення
   `{updatedAt}`. На 200 повертати `{ ok: true, value: undefined }`. Обробку `!outcome.ok` та
   non-200 (`named(outcome, writeRefusals)`) ЛИШИТИ без змін. `saveDraft` НЕ чіпати.
   Якщо `lessonDocumentUpdatedSchema` більше ніде в файлі не вжитий — прибрати невикористаний імпорт
   (перевір: його ще використовує `saveDraft`, тому ЗАЛИШИТИ імпорт).

3. **Queries** `frontend/src/modules/document/application/admin-document-queries.ts`, `publish`
   (рядки ~86-112): `mutationFn` тепер повертає `void` (прибрати `return result.value` → просто
   `if (!result.ok) throw new DocumentAuthoringError(result.error.code);`). `onSuccess`-інвалідацію
   (document + learn) ЛИШИТИ без змін — вона вже рефетчить draft із свіжим токеном.

4. **Редактор** `frontend/src/modules/document/ui/LessonDocumentEditor.tsx`, `onPublish`
   (рядки ~340-357). GUARD (варіант b — дзеркалити onSave, НЕ скидати overlay у null):
   ```
   const onPublish = async () => {
     setError(null);
     try {
       const saved = await save.mutateAsync({ lessonId, document, expectedUpdatedAt: draft.token });
       setEdited((cur) => ({ ...(cur ?? draft), token: saved.updatedAt }));
       await publish.mutateAsync({ lessonId, expectedUpdatedAt: saved.updatedAt });
       // publish повертає 200 з порожнім тілом (контракт re-read); свіжий токен беремо з
       // авторитетного re-read, який зробила publish-onSuccess інвалідація. Overlay НЕ скидаємо —
       // зберігаємо правки, набрані під час раунд-трипу (симетрично з onSave).
       const fresh = client.getQueryData(documents.document(lessonId).queryKey);
       setEdited((cur) => (cur === null ? null : { ...cur, token: fresh?.updatedAt ?? cur.token }));
       setDirty(false);
       setStatus({ kind: 'published' });
     } catch (failure) {
       setError(
         t(
           `administration.document.errors.${failure instanceof DocumentAuthoringError ? failure.code : 'unavailable'}`,
         ),
       );
     }
   };
   ```
   ВАЖЛИВО: `documents.document(lessonId)` — це queryOptions; `.queryKey` дає ключ. `getQueryData`
   типізується як `LessonDocumentDraft | undefined`. Звір точний тип повернення (поле `updatedAt`).
   Більше в редакторі нічого не міняти (кнопки, busy-гейт лишаються як є).

5. **Тести (ВИПРАВИТИ фейкові контракти):**
   - `frontend/src/modules/document/adapters/teach-document-repository.test.ts`:
     - рядок ~36-40: мок `http.post(.../document/publish ...)` має повертати ПОРОЖНЄ 200
       (`new HttpResponse(null, { status: 200 })`), НЕ `HttpResponse.json({updatedAt})`.
     - рядок ~59: `expect(await repo.publish(lessonId, updatedAt, 'en')).toEqual({ ok: true, value: undefined })`.
     - `seen`/`request.json()` перевірки лишити.
   - `frontend/src/modules/document/ui/LessonDocumentEditor.test.tsx`:
     - fake `publish` (рядок ~36-38): повертати `{ ok: true, value: undefined }` (БУЛО `{updatedAt:'...02:00:00Z'}`).
     - тест «publish saves the current edits first...» (~255-265): асерт `expectedUpdatedAt` =
       saved-токен (01:00:00Z) ЛИШАЄТЬСЯ правильним. Якщо тест десь читав published-токен — звір і
       адаптуй під те, що тепер токен оновлюється з re-read (fake getDraft повертає конструкторний
       updatedAt). Рядок ~522 `pending.resolve({ ok: true, value: {updatedAt} })` → `value: undefined`.
   - Якщо є ще тест, що очікує publish-value з токеном — виправ під `void`. Шукай `repo.publish`/
     `publish:` у document-модулі.

## Межі
- Торкатися ЛИШЕ `frontend/src/modules/document/**` (порт/repo/queries/editor + 2 тести).
- НЕ чіпати: бекенд, openapi, generated-клієнт, `.gitignore`, `docs/security/_wip-*`,
  `frontend/e2e/teach-editor.spec.ts` (e2e — координатора; toast-асерт лишиться регресією).
- Спільне git-дерево: НЕ роби git checkout/reset/rm/commit. Коміт і гейти — координатор.
- НЕ запускай docker/e2e/make — гейти жене координатор.
- ПІСЛЯ правок прожени САМ лише статику й unit: з `frontend/`: `npm run format` (Codex часто
  пропускає — обов'язково), потім `npm run typecheck`, `npm run lint`, і цільові unit-тести
  document-модуля (`npm test -- document` або точний шлях). Доповідай exit-коди.

## Журнал
(нижче дописує Codex)

ДАЛІ: Прочитати бриф і звірити B1 + guard із поточним document-модулем; змінювати лише погоджені шість файлів та цей append-журнал.
ФАКТ: Робочий файл прочитано. Вимогу першого запису до першого інструменту побачив лише після читання; цей запис зроблено після читання, без приховування порядку. Git/docker/e2e/make не запускалися.
ФАКТ: Порт, publish repo/query та редактор відповідають опису дефекту; saveDraft використовує lessonDocumentUpdatedSchema, імпорт зберігаємо. npm format = oxfmt .; перед запуском забезпечу обмеження записів document-модулем.
ДАЛІ: Звірити решту publish-фейків, тип draft, локальні інструкції та контекст тесту з pending.
ФАКТ: Звірено onSave/onPublish та publish-фейки в document-модулі. Додаткові publish-фейки з updatedAt є в application/admin-document-queries.test.ts:21 і application/teach-document-queries.test.ts:21; пункт 5 дозволяє їх адаптацію. Продуктові файли ще не змінені.
ПИТАННЯ: STOP — суперечність точного брифу з кодом. Пункт 5 наказує в ui/LessonDocumentEditor.test.tsx:522 замінити pending.resolve({ ok: true, value: {updatedAt} }) на value: undefined. Але рядки 503–527 — тест «an alt edit made while save is pending survives save completion», pending має тип ReturnType<DocumentAuthoringRepository['saveDraft']>, а override викликається через saveDraft. Це НЕ publish. Заміна порушить незмінний за пунктом 1 контракт saveDraft та тест onSave. Потрібне уточнення координатора: залишити цей pending.resolve без змін і застосувати void лише до publish?
ФАКТ: Через вимогу зупинитися при суперечності реалізацію та npm run format/typecheck/lint/цільові unit-тести НЕ розпочато; exit-кодів цих перевірок немає (NOT_RUN). Git/docker/e2e/make не запускалися. Змінено лише append-журнал цього робочого файлу.
ДАЛІ: Після уточнення виконати B1 + guard та дозволені перевірки. Для format доступні виключення PATH з !; можна обмежити npm run format document-модулем без змін package.json.

РІШЕННЯ: 2026-10-31, власник уточнив помилку брифу: pending.resolve у тесті saveDraft (рядок 522) залишити незмінним; void лише у publish-фейку редактора. Дозволено також адаптувати publish-фейки admin-document-queries.test.ts і teach-document-queries.test.ts. STOP знято. Scope — 8 погоджених файлів document-модуля та цей append-журнал.
ДАЛІ: Повторно звірити поточні байти восьми файлів, застосувати B1 + guard-b, потім послідовно format/typecheck/lint/document unit з frontend/. Git/docker/e2e/make заборонені й не запускатимуться.

ФАКТ: Повторна звірка підтвердила актуальність B1 + guard-b; LessonDocumentDraft.updatedAt має тип string | null. Пошук локального .codex/config.toml повернув відсутність файлу (загальний exit 2 читання), продуктових суперечностей немає.
ФАКТ: Реалізовано рівно 8 погоджених файлів: publish void у порті/repo/query, guard-b читає token з query cache та зберігає overlay, 4 publish-фейки адаптовано; saveDraft і pending.resolve незмінні. Оригінали для перевірки без git: /var/folders/hr/mx2k6bvj7s57j_sqyxl819n80000gn/T/te6-b1-q29bxkid.
ДАЛІ: Запустити npm run format зі scoped виключеннями, потім typecheck, lint і document unit.

ФАКТ: З frontend/ виконано npm run format із PATH-виключеннями всього, крім 8 погоджених файлів (script oxfmt .; package.json не змінювався). Exit 0.
ДАЛІ: npm run typecheck.

ФАКТ: З frontend/ виконано npm run typecheck; exit 0.
ДАЛІ: npm run lint.

ФАКТ: З frontend/ виконано npm run lint; exit 0.
ДАЛІ: npm test -- src/modules/document.

ФАКТ: З frontend/ виконано npm test -- src/modules/document; exit 1.
ДАЛІ: Звірити остаточну різницю з локальними оригіналами без git та зафіксувати підсумок.

ФАКТ: Перший unit-запуск exit 1: Rstest не стартував через sandbox listen EPERM ::1:3000. Повторено ту саму npm test -- src/modules/document з розширеним дозволом виконання; exit 0.
ДАЛІ: Звірити diff без git та підсумувати результати.
ФАКТ: Unit-повтор PASS: 8 test files, 61/61 tests, 0 skipped, exit 0. Зокрема зелений publish-save-token сценарій і незмінний saveDraft pending-edit тест.
ФАКТ: Остаточний diff перевірено через Python difflib проти scoped оригіналів без git: лише погоджені 8 файлів, saveDraft/schema import/error handling/invalidation/busy-гейт незмінні; у LessonDocumentEditor.test.tsx змінено виключно publish-return. Форматування обробило рівно 8 файлів. Статика та unit доводять свій рівень; browser/e2e не запускалися.
ДАЛІ: Зафіксувати службовий session checkpoint та передати результат координатору; продуктові правки завершені.
ФАКТ: Службовий checkpoint за чинною native identity спробувано через зареєстрований manual.sh; exit 1 на mktemp receipts (Operation not permitted, sandbox), до runtime dispatch. Checkpoint НЕ створено, OK не заявляється; продуктові перевірки завершені незалежно від нього. Git/docker/e2e/make не запускалися.
ФАКТ: Підсумок: B1 + guard-b виконано, format 0 / typecheck 0 / lint 0 / document unit 0 (61/61 після sandbox retry; первинний запуск 1). Журнал збережено; наступні review/e2e/full gates — координатор.
