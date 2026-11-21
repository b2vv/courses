# S5-L1a Independent Review

Model: claude-opus-4-8

ДАЛІ: отримати повний дифф `git diff origin/main...HEAD`, прочитати план plan-s5l1.md, зібрати список змінених файлів.

## ФАКТИ (перевірено по коду)
- Ownership: authorizeAudioRead (non-lock) для List; authorizeAudioEdit (FOR UPDATE OF assets, той самий lockOwnedMediaAsset як Mint/retry) для Set/Rename/Delete. CanEditLesson(ownerID) — той самий join. audiotrack.go:37-65.
- Partial unique index 00026: (video_asset_id, language) WHERE state<>'failed'. Name coupled у audioTrackConflict:575. 409 через ConstraintName-гілку. 23514 лишається 400. Коректно.
- D-5b regen: rewriteReadyTracks читає Ready-набір у тій самій tx (бачить uncommitted) під локом, перед commit, на всіх 3 мутаціях. Fail rewrite → abort (тест rename rollback підтверджує).
- Lock-порядок asset→tracks однаковий у promote/set/delete/update/RegenerateMaster → без deadlock, D7 single-writer тримається локом.
- ValidateTrackDeletion: last-Ready-of-Ready refuse. validLabel забороняє ", CR/LF, ctrl → m3u8-ін'єкція неможлива. RegenerateMaster — чистий Go (master.go), без ffmpeg-бінарника; atomic temp+rename.
- strict JSON decode (decodeRenameTrack): content-type, utf8, single object, no unknown/duplicate keys, обидва поля present, EOF. MaxBytesReader. requireSameOrigin на PATCH/POST/DELETE, не на GET. Коректно.
- Тести: integration ownership(4 методи)/409-PATCH/delete-regen/embedded-delete/rename-rollback зареєстровані (repository_integration_test.go:196-199). Handler AT01-AT08 (list/rename/default/delete/non-owner/strict-body/409/malformed-id). Fidelity +4 paths +3 schemas byte-identical.
- Generated MediaAudioRename вимагає обидва поля (required) — збігається з backend strict decode, без client/server mismatch.

## ДЕВІАЦІЇ/НОТИ (не блокуючі)
1. POST .../default → 204, план казав 200 MediaAudioTrack. Внутрішньо консистентно (contract+overlay+handler 204). UI мусить refetch. Low.
2. PATCH вимагає обидва language+label (full-replace), план казав {language?,label?}. Контракт консистентний (generated required). Observation. Low.
3. media root: API хардкодить "/media" літерал, worker — const mediaRoot. Обидва "/media"; латентні 2 джерела правди. Very low.
4. Реальний API→диск master-write у спільний volume доведено unit(master_test)+механізмом, НЕ integration/e2e цього зрізу (integration — fake rewrite). Прийнятно за фазовим планом (повний HLS e2e → S7). Low-med, нести як limitation.
5. Немає окремого тесту, що мутуючий audio-ендпоінт без matching Origin → 403. Middleware той самий проваджений, wiring коректний; gap лише тестовий. Low.

## РІШЕННЯ
VERDICT: PASS_WITH_NOTES. Жодної блокуючої знахідки (correctness/security/architecture/edge чисті). Зелений гейт лишається зеленим.
