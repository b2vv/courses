
---
# GATE 2 RE-CHECK — editor-polish plan (2026-10-15)
Model: claude-opus-4-8

ДАЛІ: звірити, що правки плану закрили B1 (stale-closure/no-text-loss) та B2 (reorder-межа) без нової діри; перевірити N1/N4/N5/N6/N7.

ФАКТ (код звірено):
- save+publish mutations (admin-document-queries.ts) обидва invalidateQueries(document key) on success → `base` рефетчиться зі свіжим token. Це знімає занепокоєння про null-гілку token-апдейту.
- LessonDocumentEditor: draft=edited??base (112); setEdited у 186/202/204 спредить стале `draft`; setSections(131) бере draft.sections; upload apply(164) через onChange closure за index.

ФАКТ (B1):
- План S2 крок 3 (104-123) тепер явно «ФІКС коду, НЕ вже коректно». Приписує:
  (a) token-апдейти 186/202/204 → setEdited(cur=>cur?{...cur,token}:cur);
  (b) setSections → функційний updater (cur??draft).sections;
  (c) upload-apply → дельта-функційний за section.id (не стейл-масив за index), row — вкладена дельта за column.id.
- Core-сценарій A-BH1 (правка секції під pending save) ЗАКРИТО: non-null гілка зберігає cur.sections; delta-by-id зберігає alt, змінений під час аплоаду (план прямо називає, 119-120).
- Тест пишеться ПІСЛЯ фіксу і реально править під pending (122-123, test c) + паралельні аплоади (test d).

РІШЕННЯ B1: ЗАКРИТО. Механізм коректний. Синхронні add/remove/move (масив від draft) лишаються правильні — виконуються на закоміченому рендері. delta-by-id фіксить і reorder-під-аплоад, і alt-під-аплоад.

ФАКТ (non-blocking, НЕ діра, не блокує):
- token null-гілка `cur?{...cur,token}:cur` при edited===null не сидить token синхронно (на відміну від поточного {...draft,token}); покладається на async-рефетч від invalidateQueries. Самолікується; у вузькому вікні подвійний no-edit-save міг би дати recoverable conflict (ErrorState), НЕ втрату тексту. Рекомендація (опц.): форма {...(cur??base),token} зберегла б синхронний seed. Не блокує — invalidation покриває, A-BH1 про sections, не token.
- row upload apply: дельта за column.id потребує протягнути section.id+column.id у apply (зараз apply=(imageId,contentType)=>void через index-closure). План називає механізм ("вкладена дельта за column.id"), код не дає — прийнятно для altitude плану; імплементації треба явно адресувати обидва id.

ФАКТ (B2):
- План S4 крок 3 (178-184) тепер оголошує ОБА: рух t('movedTo',{pos,total}) І межу-no-op t('moveBlocked',{edge}). moveSection(144-154) справді no-op на краю — є що оголошувати.
- Acceptance A-BH3 (228) відновлено до «позиція І межа оголошені».
- Датоване рішення є: 2026-10-15 Courses 1, §Рішення GATE 2 (258-259) — no-op оголошується, тиша на межі нечитабельна для скрінрідера.
РІШЕННЯ B2: ЗАКРИТО. Мінорна нота: повторне натискання ↑ на краю дає ідентичний текст live-region → SR може не переоголосити (aria-live re-announce лише при зміні тексту). Не блокує.

ФАКТ (non-blocking N1/N4/N5/N6/N7):
- N1 (e2e aria-label на input): S1 крок 3 (62-66) тепер ОБОВ'ЯЗКОВО, верифіковано по RAC FileTrigger.mjs, usage передає aria-label явно. Відбито.
- N4 (drop mime): S1 крок 1 клієнтський MIME-фільтр + крок 4 названий серверний fallback (unsupported_image/invalid_image→ErrorState). Відбито.
- N5 (паралельні аплоади): S2 крок 1 setUploadingKey(k=>k===key?null:k) + тест (d). Відбито.
- N6 (ціль add-фокусу=редаговане поле, не кнопка тулбара): S4 крок 1 явно; video/row→перша кнопка дій секції (закриває і defense video-fallback gap). Відбито.
- N7 (спекулятивний fileHint): прибрано, S1 i18n (75) прямо «НЕ додаємо». Відбито.
- N2 (дві role=status): скоуп за name у тестах S2/S4; НЕ role=alert → e2e alert-count-0 живе. lesson-builder.spec не має getByRole('status') (критика). Відбито.

РІШЕННЯ: нових БЛОКУЮЧИХ дір немає. B1 закрито, B2 закрито. Дві мінорні не-блокуючі ноти (token null-гілка форма; row-apply id-threading; live-region повтор на краю). Вердикт: PASS.
