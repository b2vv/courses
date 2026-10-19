# editor-polish — жива нота spec-flow (Courses 1)

РОБОЧИЙ ФАЙЛ. Дописуй по кроках; розвилки — одразу сюди (flush рано й часто).

## Поріг (названо 2026-10-15)
Структурний поріг spec-flow = нова підсистема / зміна публічного API / крос-зріз >1 шар І бекенду
І фронту. Editor-polish — **фронт-онлі** (presentation + scoped CSS + reader-типографіка + D1/D2
поведінка в presentation). Бекенд/контракти/схема не чіпаються → за буквою маніфесту **нижче** порога.
**Рішення:** проходимо spec-flow усе одно — явний owner-opt-in (власник 2026-10-15 «editor-polish через
spec-flow»; Конституція: user scope перебиває церемонію) + робота крос-FE зі спільним CSS (клас пастки
admin-restyle, що протікав 3/6 зрізів) → структура дає цінність. Це opt-in, не поріг-мандат.

## Setup
- Гілка: `feat/editor-polish` → PR у `main` (ff коли лінійна). Створюю ПІСЛЯ закриття B-18 (щоб не
  мішати дерево під фоновим B-18-гейтом). Планувальні артефакти пишу зараз (untracked), комічу на гілці.
- Дім артефактів: `docs/editor-polish/` — `spec.md`, `plan.md`, `tasks.md`, цей `_wip-spec-flow.md`.
- Роль: Courses 1 (план/код через Codex/гейти/коміти/PR), Courses 2 — рев'ю зрізів. GATE-вердикти —
  свіжим незалежним сабагентом (gate-reviewer / grilling / plan-critique / plan-defense).

## Основа-аналіз
`docs/lesson-builder/_wip-editor-design-codex.md` (Codex, 2026-10-13): критика editor+reader, ранжований
план QW1–QW6 (scoped CSS/примітиви, малий ризик) + поведінковий D1–D2 + дизайн-прохід D3–D5.

## Скоуп фази (пропозиція — підтвердити на GATE 1)
- **В скоупі:** QW1–QW6 (видимі межі секцій + зона вводу; FileUpload/TextField замість сирих input;
  prose-типографіка reader+editor + виразний task; grid колонок + підписи палітр; toolbar-групування +
  active-state; loading/error/empty) + D1–D2 (чесні save/publish/upload-статуси; URL-помилка/фокус-ладування).
- **Поза скоупом (майбутнє):** D3 (draft-preview + контекст уроку), D4 (сторінка уроку без відео —
  зачіпає оболонку/route + повний e2e), D5 (довгий документ / sticky), drag-reorder і drag-полотно
  (окрема owner-фаза), зміна доменної моделі/контрактів/санітизації, редизайн NorthWind-оболонки.

## Відомі розходження (з аналізу — не «виправляти» тут)
- `main.css` імпортує tailwindcss — не чіпати; весь новий CSS scoped + токени.
- Admin примусово dark (`tokens.admin.css`) — editor перевіряти в dark; reader — у light І dark.

## GATE 1 — спільне розуміння (ПРОЙДЕНО 2026-10-15)
Метод: grilling-інтерв'ю (1 раунд, 4 питання фронтиру), вердикт — власник підтвердив усі 4
(«Підтверджую все починай»). Рішення:
- **Q1 скоуп:** QW1–QW6 + D1–D2 в скоупі; D3–D5 + drag поза. Наслідок: «порожній player виглядає
  незавершено» (D4) цієї фази НЕ вирішується; reader-робота = лише типографіка документа, оболонку уроку
  не чіпаємо.
- **Q2 тема:** editor dark-only (світлу адмінку не відкриваємо — окреме owner-рішення); reader — обидві теми.
- **Q3 нарізка:** одна гілка `feat/editor-polish`, зрізи-коміти (QW1→…→D2), рев'ю Courses 2 на ключових
  зрізах, ОДИН фінальний PR у `main` з owner merge-сигналом (editor-polish цілісніший за admin-restyle).
- **Q4 верифікація:** host-Playwright скриншоти + computed-styles + WCAG-розрахунок контрасту + bbox-вимір
  44px (CDP не дає холістичного SSR-Suspense рендеру); + owner фінальний візуальний огляд ключових екранів
  (editor dark + reader light/dark) перед merge.

## Журнал
ДАЛІ: написати spec.md (намір: Проблема/Сценарії/Acceptance/Out-of-scope/Відкриті), тоді GATE 1 (grilling).
ФАКТ: B-18 фікс зроблено (e2e-скриншоти → `playwright/.artifacts/` ignore; закомічені PNG лишено як
doc-артефакти), повний `make check` запущено у фоні (лог `scratchpad/b18-gate.log`), чекаю GATE_EXIT.
ФАКТ: spec.md написано; GATE 1 пройдено (4/4 owner-confirmed). ДАЛІ: plan.md (ЯК) — reuse-first, side-effects/
migration/rollback, конституція, acceptance-spec таблиця; тоді GATE 2 (plan-critique + plan-defense + конституція).

## ⚠️ ПЕРЕУЗГОДЖЕННЯ СКОУПУ (2026-10-15, перед планом) — преміса spec протухла
B-18 закрито (GATE_EXIT=0, e2e 120/120, дерево чисте, скрини → .artifacts). Explore+CSS-читання+host-
Playwright скрини (cur-editor-1280.png, cur-reader-1280.png) ДОВЕЛИ: візуальний editor-polish уже
відвантажено **admin-restyle S3** (`639c7ca`, 2026-10-14, ПІСЛЯ аналізу 2026-10-13). Скрин редактора —
доглянутий: картки секцій з іконками-плитками/типом/нумерацією/↑↓×, згрупований toolbar, зони вводу з
рамками, палітра «Додати секцію» плитками, Save/Publish ієрархія. QW1/QW3/QW4/QW5/QW6 = ЗРОБЛЕНО.
**Реально відкрите (НЕ закрив S3 — він лише візуал):**
- QW2: image-upload — сирий нативний `<input type=file>` «Choose File / No file chosen» (не локалізовано,
  не через примітив FileUpload). Єдиний видимий візуальний нит у редакторі.
- D1: немає підтвердження save/publish; `upload.isPending` спільний для всіх зображень (поведінка).
- D2: невалідний URL тихо видаляє посилання + закриває поле; фокус після add/delete/reorder невизначений.
**Reader-скрин:** документ-рендер оформлений (task-картка, проза), АЛЕ найжирніше на сторінці — «ще немає
відео» (проблема D4, оболонка уроку) — D4 поза скоупом за owner-рішенням Q1.
РІШЕННЯ ПОТРІБНЕ ВІД ВЛАСНИКА: переузгодити скоуп — більшість spec (QW1/3/4/5/6) вже є. Чекаю вибору:
(A) QW2+D1+D2 [рек.], (B) лише D1+D2, (C) переглянути D4 (більше, зачіпає route/оболонку+e2e).

РІШЕННЯ ВЛАСНИКА (2026-10-15): **скоуп = QW2 + D1 + D2** (варіант A). D4 лишається поза (найбільший
learner-facing недолік, але окрема більша фаза route/playback+e2e).
СТЕР ВЛАСНИКА (2026-10-15): «в RAC є компонент DropZone». Звірено — RAC експортує `DropZone` + `FileTrigger`
(runtime-підтверджено); наш примітив `FileUpload` (controls.tsx:273-286) обгортає ЛИШЕ FileTrigger (клік).
→ QW2 plan: **розширити примітив `FileUpload`** drop-зоною (DropZone drag-drop + FileTrigger клік), тримати
спільним примітивом (НЕ інлайн DropZone у LessonDocumentEditor); обґрунтувати додавання + accept jpeg/png/webp
+ a11y (клавіатура, 44px, фокус) + локалізувати підписи (замість нативного «Choose File / No file chosen»).

## ПОРІГ (переназвано після звуження): QW2+D1+D2 — фронт-онлі, 3 фокусні зміни → НИЖЧЕ структурного
порога ще явніше. Тримаю spec-flow (owner-opt-in), але ПРОПОРЦІЙНО: plan.md (reuse/side-effects/rollback/
acceptance інлайн — таблиця мала, acceptance-spec делегацію пропускаю, назву в ноті) + GATE 2 незалежним
рев'ю (цінний для D2 focus-пасток, [[isdisabled-on-focused-button-drops-focus]]) + Codex-імплементація
(роль Courses 1) + verify (гейт + Courses 2). Будь-яке скорочення проти skill — назване тут, не мовчазне.

═══════════════════════════════════════════════════════════════════════════════
## HANDOFF ДЛЯ ПЕРЕЗАПУСКУ (2026-10-15, транскрипт 5.7МБ → нова сесія)
**Гілка:** `main` (ще не відгалужено). HEAD `039e345`. Робоче дерево: B-18-зміни НЕ закомічені.
**Зроблено цієї сесії:**
1. **B-18 ЗАКРИТО** (не закомічено): e2e-скриншоти → `playwright/.artifacts/` (ignore), закомічені PNG
   лишено як doc-артефакти. Змінені: `frontend/e2e/accessibility.spec.ts`, `frontend/e2e/recovery.spec.ts`,
   `.gitignore`. Повний `make check` = GATE_EXIT=0, e2e 120/120, дерево `docs/foundation/screenshots/` чисте.
   **НАСТУПНЕ: закомітити B-18 на гілці `feat/b18-screenshot-artifacts` (або fold у editor-polish?) → PR → merge.**
2. **editor-polish spec-flow:** маніфест прочитано; spec.md + GATE 1 (grilling, 4/4 owner-confirmed) ✓;
   **скоуп переузгоджено → QW2+D1+D2** (візуал уже зробив S3 `639c7ca`); DropZone-стер у план QW2.
**НАСТУПНА ТОЧНА ДІЯ (нова сесія):** написати `docs/editor-polish/plan.md` (ЯК) для QW2+D1+D2:
reuse-first (розширити FileUpload DropZone'ом; Toast/output для D1; TextField isInvalid+Dialog/focus для D2),
side-effects/migration/rollback, конституція, acceptance-таблиця (інлайн). Тоді GATE 2 (plan-critique +
plan-defense свіжими gate-reviewer сабагентами + конституція інлайн). Тоді tasks.md → Codex-імплементація → гейт → Courses 2.
**Артефакти:** `docs/editor-polish/{spec.md, _wip-spec-flow.md (цей), _wip-facts-codex.md}`. Скрини стану:
`scratchpad/cur-editor-1280.png`, `cur-reader-1280.png`.
**Відкриті питання власника:** нема (скоуп підтверджено). Merge editor-polish — owner-gated (один фінальний PR, Q3).
═══════════════════════════════════════════════════════════════════════════════
