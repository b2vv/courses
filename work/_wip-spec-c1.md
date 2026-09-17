# WIP — Phase 7 transcript spec (Courses 1)

ДАЛІ: звірити первинні джерела під spec — PRODUCT-BRIEF §13/§21, ADR 006, docs/playback/spec.md
як форма, і реальні seams у коді (lesson, learning read model, playback mint).

## ФАКТ (звірено на 78d0117, 2026-09-17)

- Phase 6 закрита 102/102: порахував унікальні рядки §8 сам — 14 P + 49 B + 39 R = 102.
- `docs/playback/report.md` §2 пише «B01–B47 plus B21a» = 48; насправді 49 (B21a **і** B24a).
  Підсумок 102 правильний, розкладка в прозі — ні.
- `work/ledger.md` обривається на 2026-09-07 Phase 5 GREEN. Закриття Phase 6 в реєстрі немає.
- `work/AGENDA.md` і `work/CTO-RESEARCH.md` стоять на `b913e81` — **32** коміти позаду (не 25).
- `PRODUCT-BRIEF.md:665` — не фіча, а рядок зі списку §17 «Не логуй». Приватність transcript
  documents як продуктова спроможність ніде не обіцяна.
- `:686` — §18 Testing strategy → Playwright, «click-to-seek **у майбутньому slice**».
  Опора для click-to-seek — §13 `:539`, не `:686`.
- `docs/adr/006-transcript-model.md` **існує**, статус «accepted design; implementation
  scheduled for Phase 7». Не бронь — готові рішення.

## ПИТАННЯ (у список GATE 1, не вирішую сам)

- Q1. WebVTT: §21 `:855` кладе його в Phase 7; ADR 006 — «WebVTT export is a **later** adapter».
  Пряме протиріччя двох первинних джерел по скоупу фази.

## ФАКТ (додано після листа Courses 2, звірено самостійно)

- `frontend/src/presentation/learn/LessonPlayer.tsx:421` — `oxlint-disable-next-line
  jsx-a11y/media-has-caption`, причина «T7 has no caption source contract; an empty track
  would be misleading». `T7` там — таск Phase 6. Зняти можна лише `<track>`, тобто WebVTT.
- `learning.go:27` і `playback/domain/progress.go:63` — навмисний дубль межі позиції,
  коментарі в обох. Не DRY-ити.
- `docs/playback/plan.md:514` несе ту саму неповну розкладку, що й report §2. Правку взяв
  Courses 2 (обидва файли поза моїм скоупом; у `git status` вони вже M — це його робота).

## РІШЕННЯ

- Q1 (WebVTT) підсилено: без нього A28 не закривається — придушення a11y лишається на фазу.
- Жодного з Q1–Q10 не вирішував за замовника.

## ДАЛІ

- Чернетка `docs/transcript/spec.md` готова (235 рядків), закомічена. GATE 1 **не пройдено**.
- Чекаю: (1) відповіді замовника на Q1–Q10; (2) брифінг Courses 2 на прожарку.
- Перед закриттям GATE 1 — `plan-critique` + `grilling` по spec.
