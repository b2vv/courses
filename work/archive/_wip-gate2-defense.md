# GATE 2 — DEFENSE lens — editor-polish plan

Model: claude-opus-4-8

ДАЛІ: читаю plan.md, spec.md, _wip-facts-codex.md; потім прожену plan-defense (5 питань) проти плану й коду.

## ФАКТИ (звірка коду)
- FileUpload (controls.tsx:273-286): Aria.FileTrigger + secondary Button; НЕ має onDrop/dropLabel зараз. Прихований input від FileTrigger не отримує aria-label автоматично → ризик e2e getByLabel. План S1.3 (62) це називає умовно.
- ImageField (LessonDocumentEditor.tsx:462-473): сирий `<label><span>uploadImage</span><input type=file accept=...></label>`. Accessible name input'а йде від тексту label. Native accept фільтрує лише клік-шлях.
- upload = single useMutation (97); uploading={upload.isPending} тече в усі ImageField (272,338,347,405). План S2.1 → uploadingKey single-string.
- КЛЮЧОВЕ: onSave.onSuccess (186) `setEdited({ ...draft, token: result.updatedAt })`; onPublish (202,204) те саме. `draft` (112) — знімок РЕНДЕРУ, у якому створено onSave; замикання стале. mutate-колбек прив'язаний до того рендеру. Якщо автор редагує під час pending save → setEdited(newSections) у новому рендері, але in-flight onSuccess робить setEdited({...stale draft, token}) → СЕКЦІЇ ВІДКОЧУЮТЬСЯ до стану до правки. Правку втрачено.
- План S2.3 (94-96) стверджує «вже коректно: onSuccess оновлює лише token, не sections» — ХИБНО: спред `...draft` тягне стале `sections`. План пропонує лише «додати тест», а не фікс.
- applyLink (RichText 59-64): невалідний → unsetLink + закриває. План S3 (110-120) фіксить: не unsetLink, errorMessage, поле відкрите, фокус у полі, видалення — окрема кнопка. Покрито.
- moveSection (144-154): no-op на краю (146). План S4.3 (149) оголошує лише успішний рух. Покрито.
- video-body = `<p>` не фокусований (322). План S4.1 (138) приклади «richtext/file-кнопка» не називають video; fallback на section-bar кнопку працює, але не описаний явно.
- removeSection focus → палітра: палітра — fieldset Button'и (277-301), стабільний якір. План S4.2 покрито.
- DropZone: Aria.DropZone не фільтрує mime; план onDrop (54) бере перший файл isFileDropItem, НЕ фільтрує тип. Wrong-mime drop → onUpload → server-reject → ErrorState. Graceful, але план не називає.

## РІШЕННЯ (DEFENSE-лінза, 5 питань)
1. Кривий інпут — ЧАСТКОВО (URL покрито S3; drop wrong-mime не названо; e2e-label ризик умовний S1.3).
2. Відмова залежності — ПОКРИТО (upload/save/publish onError→ErrorState; onSettled чистить key).
3. Гонка й повторний виклик — ДІРА (блокуюча): стале замикання save/publish відкочує правку; план хибно вважає «вже коректно». + single uploadingKey під паралельний аплоад не визначено (§7.3 лише питання).
4. Порожні/граничні — ПОКРИТО здебільшого (empty/delete-last/edge-reorder); video-add focus-fallback не названо.
5. Rollback — ПОКРИТО (commit-per-slice, revert, branch не чіпає main, no migration).

ВЕРДИКТ: CHANGES-NEEDED (блок: §3 стале-замикання no-text-loss).
