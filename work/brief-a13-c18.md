# Brief A13 / крок 4f — C18: знімок має триматися токена ЗАВАНТАЖЕНОЇ чернетки

Частина A13 (Phase 8). Останній червоний e2e (C18, `frontend/e2e/asr.spec.ts:300`): після reload
пропозиція відновлення звичайна, а не застаріла. Діагноз — Courses 2, звірено сесією Courses 1,
2026-09-24.

## Дефект

`frontend/src/presentation/admin/TranscriptEditor.tsx:83`:
`const currentDraftUpdatedAt = suppliedDraftUpdatedAt ?? draftUpdatedAt;`
`suppliedDraftUpdatedAt` — живий токен чернетки з композиції, він змінюється під час перегенерації.
`draftUpdatedAt` — локальний стан (:79), токен чернетки, яку автор завантажив. Baseline (:84-89)
бере живий. Знімок recovery записує `current.baseline.draftUpdatedAt`
(`use-transcript-recovery.ts:106-108`). Потік C18: завантаження T1 → правка → перегенерація (живий
токен T1 → T2) → наступне захоплення знімка пише T2. Після reload baseline теж T2, тож
T2 == T2 → `offer`, а не `offer-stale`.

## Зроби

1. Поміняй пріоритет: `const currentDraftUpdatedAt = draftUpdatedAt ?? suppliedDraftUpdatedAt;`
   У сесії рядки походять із завантаженої чернетки, тож знімок тримає її токен. Після reload
   локального токена немає (`draftOffer` відсутній), тож baseline бере живий, і порівняння
   T1 проти T2 дає `offer-stale`. Ручна правка без завантаження поводиться як досі, бо локального
   токена немає.
2. Unit-тест у `TranscriptEditor.test.tsx`: завантаження чернетки з токеном T1 → правка →
   `draftUpdatedAt`-проп змінюється на T2 (перегенерація) → захоплений знімок несе T1. Якщо
   наявний тест стверджує, що перевага за живим токеном, — він стверджував хибне: онови його й
   назви це в звіті.
3. Більше нічого не чіпай.

## NOT
Бекенд, e2e, інші компоненти, git, `tasks.md`, `report.md`.

## Перевірки
`frontend`: `npm run format:check`, `npm run lint`, `npm run typecheck`. `npm test` у пісочниці не
стартує — познач чесно; прожене reviewing-сесія.

## Звіт
`work/report-a13-c18.md`
