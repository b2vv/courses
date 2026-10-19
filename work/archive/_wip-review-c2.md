claude-opus-4-8

ДАЛІ: прочитати spec.md / plan.md / report.md та зняти diff main...feat/editor-polish по frontend/; потім пройти 4 фокуси рев'ю.

ФАКТ (T1 controls.tsx): DropZone рендериться лише коли onDrop задано (controls.tsx:286 `if (onDrop===undefined) return trigger`). Відео (LessonVideoField) не передає onDrop → single button. ui.test "keeps video upload as a single button" підтверджує. MIME-фільтр jpeg/png/webp на drop (controls.tsx). PDF відсіяно (тест). Регресу відео нема.

ФАКТ (T2 B1): setSections тепер функційний updater (LessonDocumentEditor.tsx:164-172), applyImage — дельта-за-id для секції/колонки (173-190), token-апдейти функційні (281,300,302). Тест "an alt edit made while save is pending survives save completion" реально редагує alt під pending save через deferred promise і перевіряє виживання — справжній регрес-тест. Паралельні аплоади: onSettled `setUploadingKey(k=>k===key?null:k)` (252) — resolve першого не гасить другий (тест N5 real).

ФАКТ (T3): applyLink невалідний URL → setLinkError+focus+return, НЕ unsetLink (RichText:67-71). removeLink кнопка умовна на isActive('link') (139-142). shouldRerenderOnTransaction:true (36). linkInputRef→Aria.Input (TextField форвардить ref на input). Тести з реальним Tiptap editor + getJSON — справжні.

ФАКТ (T4): pendingFocus ref + useEffect[sections] + MutationObserver fallback для пізнього tiptap textbox (121-144). Cleanup disconnect коректний, без витоку. Селектор add-image `.lesson-editor-image .ui-button` пропускає прихований DropZone-button. `<output aria-live=polite>` + key=sequence для re-announce. moveBlocked/movedTo обидві локалі.

ФАКТ (globally-counted role=status): lesson-editor-live завжди в DOM + Toast — обидва role=status, але ЛИШЕ в LessonDocumentEditor; інші компоненти/тести з getByRole('status') не рендерять цей редактор → колізії нема. Власні тести редактора використовують getAllByRole('status')+filter (198,482). N2 закрито.

ФАКТ (шари): усі зміни в presentation/ + main.css + locales + e2e. Домен/app/контракти не торкані.

ФАКТ (report §3 4 баги): усі чотири звірені з кодом і тестами — підтверджені.

РІШЕННЯ: APPROVE. Блокуючих нема. Non-blocking: (1) uploadingKey одиничний — при 2 конкурентних аплоадах індикатор першого зникає коли стартує другий (не Set); A-BH1 "не спільний" виконано, але "per-image" трохи переоцінено. (2) spec A-GATE браузерний a11y-вимір (44px bbox/focus) ще pending (report §5/§7 чесно) — єдиний формально відкритий acceptance, згорнутий в owner-візуальний огляд перед merge.
