# Review — O05 slice 1: the learning outline of a paid course + lesson locks

Reviewer: **Courses 2**. Uncommitted on `feat/payment` (base `993d28d`). Only the files
listed by `git status` EXCEPT `frontend/src/domain/purchase/` (next slice, not ready).

РОБОЧИЙ ФАЙЛ: `work/review-o05-outline.md` — append, перший рядок `ДАЛІ:`.

## Defect
`GetLearningCourse` reads one row per lesson and took the course's `Preview` fact from the
FIRST row (first lesson of the first module). A paid course whose preview lesson was not
first refused the whole learning page with `course_purchase_required` → the preview lesson
was unreachable (Pay11). The integration test only inserted one-lesson courses.
The four content mappings (playback session, media `auth_request`, transcript,
discussion) each select their own lesson and were NOT affected (`WHERE lessons.id = $…`
or via the asset's lesson).

## Owner decision (2026-10-01, `docs/payment/assumptions.md`)
A published paid course shows its WHOLE outline to a signed-in viewer without access;
preview lessons open, the rest locked; content stays gated per lesson.

## Change
- domain: `LearningCourse.Preview` removed; `CanBeViewedBy` passes `Preview: true` to
  `access.CanView` (still a one-statement delegate — AST test); new
  `PurchaseRequiredBy(actor, lesson)` with the lesson's own preview fact.
- application: `GetCourse` refuses only what `CanBeViewedBy` refuses (unpublished for
  strangers) and sets `lesson.Locked` per lesson; `ErrPurchaseRequired` no longer
  returned by the outline.
- adapter: `IsPreview` per lesson from the row.
- HTTP/OpenAPI: `LearningLesson` gains required `isPreview`, `locked`; Orval regenerated.
- tests: domain table over `accesstest.TruthTable` (outline = Want||WantPurchase, lock =
  WantPurchase); cross-module agreement test now compares "outline AND NOT locked" for the
  course; app test rewritten to the decision; Pay19 course row → 200 + locked (plan row
  amended); integration R06 checks the lesson preview read from SQL + new
  "preview on lesson 2 of 3" regression.
- frontend: strict schema + domain + mapping accept/carry `isPreview`, `locked`; fixtures;
  e2e learning mock.

## Gates
Go unit all ✓ · vet ✓ · gofmt ✓ · `make integration` ✓ · generate ✓ · frontend format/lint/
typecheck ✓ · rstest 509/509 ✓ · full e2e 87/87 ✓ (api + frontend rebuilt).

## Please look hardest at
1. Security: can a viewer without access now reach any CONTENT through the outline change?
   (The outline carries titles only; titles are public on the storefront program.)
2. The cross-module agreement test: is "outline AND NOT locked" the right equivalence to the
   other three modules' `Want`, or does it hide a real disagreement?
3. Draft/archived for strangers still refused (`ErrCourseForbidden`); archived + entitled
   still visible?

## Findings (Courses 2 appends below)

---
## Findings — Courses 2 (claude-opus-4-8, 2026-10-01)

ДАЛІ: (1) БЕЗПЕКА — чи LearningLesson DTO outline несе ВМІСТ (description/video/transcript),
а не лише title/isPreview/locked; (2) agreement-тест outline∧¬locked≡Want — незалежне обчислення
чи тавтологія; (3) draft/archived для стороннього відмова, archived+entitled видно. Читаю domain
learning.go (CanBeViewedBy Preview:true, PurchaseRequiredBy, Locked), application, і поля LearningLesson.

### F1 [#1 безпека — чи outline відкриває вміст] — ЧИСТО
Outline DTO пер-урок (learning_handler.go mapLearningCourse): id, title, videoState (статус-рядок
ready/processing/failed або null — НЕ відео), position/completed/clientSeq (ВЛАСНИЙ прогрес глядача),
isPreview, locked. ЖОДНОГО URL/playbackId/опису/транскрипту. Відкриття outline не-entitled глядачу
leak-ає лише: titles (вже публічні в storefront program), video-readiness-статус (негідж. метадані),
власний (порожній для locked) прогрес, + прапорці. Реальний вміст — за 4 НЕЗМІННИМИ пер-урок мапінгами
(session/media auth_request/transcript/discussion), кожен бере свій урок. Контенту НЕ відкрито.
Негідж. нотатка: videoState розкриває «відео готове/обробляється» для locked-уроку — метадані, не вміст.

### F2 [#2 agreement-тест outline∧¬locked≡Want] — ЧИСТО, незалежний (не тавтологія)
shared_rules_test:70 порівнює 4 модулі проти test.Want ОКРЕМО: viaCourse=CanBeViewedBy∧¬PurchaseRequiredBy
(композиція 2 предикатів), viaPlayback/viaTranscript/viaDiscussion=їхній CanBeViewedBy (single CanView),
усі з тих самих табличних входів через РІЗНІ code-path-и. Якби composition course розійшлась — мисмач
зловиться. Алгебра: outline∧¬locked (pub) = ¬Paid∨preview∨exempt = content Want. Рівність вірна.
Domain learning_test окремо пінить outline-видимість (Want||WantPurchase) і lock (WantPurchase) по
accesstest.TruthTable. Доказ тришаровий, узгоджений.

### F3 [#3 draft/archived] — ЧИСТО
accesstest.TruthTable містить draft(3)/archived(4)/published(7) → domain-табличний тест реально покриває.
CanBeViewedBy(Preview:true)=exempt∨Published: draft/archived для стороннього (¬exempt,¬Published)→false→
ErrCourseForbidden (app повертає саме його, не PurchaseRequired). archived+entitled→exempt→видно. ✓
Integration: «Pay11 preview на 2-му з 3» (stranger, 3 уроки preview=[f,t,f]) — точна репродукція дефекту,
IsPreview читається з SQL пер-урок.

### Forward-нотатка (наступний зріз, не тут)
Frontend purchase-домен (frontend/src/domain/purchase/, поза скоупом) мусить шанувати `locked`: не
смикати контент locked-уроку (все одно 403, але UX). Downstream-споживачі, що колись вважали «outline
завантажився ⇒ є доступ», тепер хибні — але бек-контент незмінно захищений 4 мапінгами.

## ПІДСУМОК: 0 blocking, 0 значущих nb — COMMIT-READY
Зміна правила доступу БЕЗПЕЧНА: outline = структура+titles(публічні)+статус+власний прогрес, вміст
гейтований пер-урок незмінними мапінгами. Agreement-тест незалежний (не тавтологія), рівність вірна.
draft/archived покрито TruthTable. Дефект відтворено в integration. AST-сумісність збережена (one-stmt
delegate). Негідж. нотатка (videoState метадані) + forward-нотатка (frontend honor locked, наступний зріз).
