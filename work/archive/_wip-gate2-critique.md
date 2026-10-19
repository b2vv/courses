claude-opus-4-8

ДАЛІ: читаю plan.md → spec.md → _wip-facts-codex.md, потім за потреби код (LessonDocumentEditor.tsx, LessonDocumentRichText.tsx, controls.tsx, lesson-builder.spec.ts). Лінза GATE 2 CRITIQUE через skill plan-critique.

ФАКТ (прочитано): plan.md, spec.md, facts; код усіх 4 файлів; FileTrigger.mjs; grep status/FileUpload.

ФАКТ (верифікації коду):
- FileUpload справді спільний із LessonVideoField.tsx:242,279 (grep) — blast radius плану §1 правдивий.
- FileTrigger.mjs:24-50: label НЕ тече в прихований input; input дістає aria-label лише якщо передати явно (filterDOMProps global:true форвардить aria-*). Кнопка (children) має текст label. => getByLabel('Завантажити зображення') зараз резолвить input тільки тому, що <label> обгортає і <span>текст, і <input> (LessonDocumentEditor.tsx:462-473). Після заміни на FileUpload треба ЯВНИЙ aria-label на input, інакше e2e:123 .setInputFiles червоний. План step-2 (usage) його НЕ показує; step-3 лише «перевірити… якщо треба — передати». Ризик реальний, але названий.
- getByRole('status') singular ужито масово (admin-users, admin, discussion e2e; ui.test.tsx:362/422 та ін.). lesson-builder.spec НЕ має. Але нові unit-тести редактора (plan test b/d) оперують «role=status» — S2 Toast + S4 live-region обидва role=status → bare getByRole('status') у редакторі неоднозначний.

ФАКТ (ключова знахідка — trace no-text-loss, S2 step 3):
LessonDocumentEditor.tsx:186 onSave.onSuccess = `setEdited({ ...draft, token })`, де `draft` (112) — стейл-замикання з МОМЕНТУ виклику onSave. Якщо автор редагує під час pending save, `edited` оновлюється до [A''], але onSuccess перезаписує його стейл-draft [A'] + новий token → правка [A''] ГУБИТЬСЯ. Те саме onPublish:202-204. => план хибно каже «вже коректно, лише додати тест». A-BH1/сценарій2 «повільне збереження під час правки не губить текст» НЕ виконується. Треба фікс (функційний setEdited(prev=>...)), не тест. BLOCKING.

ФАКТ (reorder-межа): spec A-BH3 (spec.md:92) «позиція/межа переміщення оголошується»; plan S4.3 лишає «межа … оголошується окремим текстом АБО не оголошується» (нерозв'язане АБО) і acceptance-таблиця плану (190) тихо звузила до «позиція оголошена». Незафіксоване звуження spec-критерію без датованого рішення (AGENTS §Decisions). BLOCKING-light.

РІШЕННЯ: 4 лінзи plan-critique прогнано в межах цього свіжого рев'ю-агента; усі 4 дали знахідки. 2 blocking (no-text-loss хибна премиса; reorder-межа ambiguity) + 7 non-blocking. Вердикт CHANGES-NEEDED.
