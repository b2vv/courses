# WIP — безшовне перемикання якості (seamless quality switch)

Робочий файл. Дописувати після кожного кроку (append). Writer = Courses 1 (ця сесія).

## Рішення (датовані, закривають альтернативу)

- **2026-11-07 (власник, через AskUserQuestion):** покращити перемикання відео так, щоб не було
  «ями» (стол/тиша/чорний кадр). Робота **поза §24** (роадмап закритий) — оформлюється окремою
  гілкою `feat/player-quality-seamless-switch` + PR, не прямим комітом у `main`.
- **2026-11-07 (власник):** механізм — **hls.js `nextLevel`**, НЕ ручний crossfade. Обґрунтування:
  на quality-switch у нас один `<video>`-елемент, який сам із собою не крос-фейдиться; `nextLevel`
  вже реалізує «не глушити старе, поки нове не прогрузилось» силами бібліотеки. Ручний
  fade-з-умовою-буфера лишається для **опції 1 (multi-audio, дві доріжки)** — окрема пізніша робота.
- **2026-11-07 (власник):** спочатку крок A (quality-switch), multi-audio (крок B) — окремо потім.

## Факт (причина «ями», з коду hls.js 1.7.2)

- `hls-media.ts:43` робив `hls.currentLevel = level`. Док hls.js `set currentLevel`:
  *"Set quality level index immediately. This will flush the current buffer... playback will
  interrupt at least shortly to re-buffer and re-sync."* → це і є яма (флашиться і відео, і аудіо —
  в HLS це одна комбінована rendition, тому відповідь на «check video too»: nextLevel тримає обидва).
- `set nextLevel`: *"switch the video quality asap, without interrupting playback"* — підміна на межі
  наступного фрагмента, старий буфер догравається, новий рівень вантажиться наперед. Коштом: ефект
  не миттєвий, а з наступного фрагмента (секунда-кілька).

## Зміна

- `frontend/src/modules/playback/adapters/hls-media.ts` `selectLevel`: `currentLevel` → `nextLevel`
  (`'auto'` → `-1` без змін) + коментар-пояснення. Port-контракт `selectLevel` не змінився.

## Чому немає нового юніт-тесту

- Adapter-тести ганяються в `node` (rstest.config.ts:29-34) — hls.js потребує MSE/DOM, не працює ні
  в node, ні в happy-dom. Моку-прецедента в коді немає. Тест із мок-Hls був би тавтологічним
  («рядок саме такий») — той самий fake, проти якого застерігає AGENTS.md. Port-контракт покритий
  стабами в `LessonPlayer.test.tsx` (16 тестів зелені). Різницю currentLevel↔nextLevel чесно
  доводить лише реальний браузер + e2e.

## Верифікація

- frontend format / format:check / lint (+architecture) / typecheck: ЧИСТО (2026-11-07).
- `rstest run src/modules/playback`: усі зелені, 0 fail.
- Повний `make check` (ребілд frontend + e2e): ДАЛІ — прогін bj3kof7av, очікую.
- e2e ризик-поверхня: `playback.spec.ts` P01 (manual 360p + seek) і R35 (auto 720p→360p). Гіпотеза:
  обидва зелені (nextLevel теж ставить manualLevel; seek/наступний fragment вантажить обраний рівень).
- Браузер: ручна перевірка quality-switch без спінера/ями (ДАЛІ, після зеленого гейта).

## Факт — гейт (2026-11-07)

- `make check` bj3kof7av: **GATE_EXIT=0**, e2e **129 passed, 0 failed**. P01 ✓ 1.9s (ручний 360p +
  seek), R35 ✓ 12.3s (авто 720p→360p проти реальних сегментів). `failedTests: 0`.
- Браузерна перевірка через БД-фікстуру НЕ робилась: повний e2e вже ганяє справжній браузер проти
  справжнього HLS із цією зміною (production-path). Перцептивну «відсутність ями» дає задокументована
  семантика hls.js (currentLevel «flush buffer, interrupt playback» → nextLevel «without
  interrupting»). Flaky-e2e на субсекундний стол свідомо не додано.

## Факт — делівері

- Коміт `aed808d` на `feat/player-quality-seamless-switch` (тільки hls-media.ts + цей wip).
- Незалежне рев'ю (independent_subagent, свіжий контекст, read-only): **VERDICT_RECEIVED = PASS**,
  0 дефектів, перевірено механізмом (hls.js сетери currentLevel/nextLevel однаково ставлять
  manualLevel; -1 = auto-шлях; споживач не читає level назад; P01 — єдиний load-bearing guard,
  R35 ортогональний). Одна неточність коміту (R35 як «доказ») виправлена amend'ом → коміт `2b63c97`.

## ДАЛІ

1. Push гілки + PR проти main (крок A). Крок B (multi-audio) — окремо.
