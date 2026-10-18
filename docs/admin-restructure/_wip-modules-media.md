# A-батч — PR5 (наступний): media — СКОУПИНГ (read-only, гейт enrollment ще йшов)

## Слід
- domain/media/video.ts (малий). Немає application/media, presentation/media.
- adapters/api: lesson-video-repository.ts(.test), lesson-video-schema.ts, tus-client.ts(.test).
- ports: LessonVideoRepository.ts, PlaybackMedia.ts (звірити: PlaybackMedia — media чи playback?).

## Споживачі domain/media
adapters(lesson-video-*/tus-client) · app/composition/admin-video-queries · **application/admin/lesson-video.ts(.test) [FLAT APP]** · ports/LessonVideoRepository · presentation/admin/LessonVideoField · routes/admin.courses.$courseId.

## Точка рішення (на час виконання)
flat `application/admin/lesson-video.ts` тягне domain/media → media має flat-application споживача (як course).
- **(A)** Якщо `lesson-video.ts` — це media's application (оркестрація аплоуду відео) → затягнути в `modules/media/application`; тоді не flat, contracts НЕ треба. Ймовірніше (tus/upload = media-інфра).
- **(B)** Якщо справді admin-специфіка → винести video-домен у `contracts/media.ts`, admin тягне звідти.
Вирішити, прочитавши lesson-video.ts. tus-client/lesson-video-schema — media adapters, у модуль.

## ВИКОНАНО (PR5)
- Option A: media = video-upload модуль. `domain/media/video`→modules/media/domain; `application/admin/lesson-video`(+test)→application; `LessonVideoRepository`→ports; `lesson-video-repository`(+test), `lesson-video-schema`(+test), `tus-client`(+test)→adapters; `LessonVideoField`(+test)→ui; barrel index.ts.
- `PlaybackMedia` + playback-адаптери лишені flat (playback пізніше).
- **КЛЮЧОВА ЗНАХІДКА (не флейк!):** перенос адаптер-тестів у modules/ вивів їх з rstest-проєкту `adapter-node` (node env) у `dom` (happy-dom) — `tus-client.test` впав (null Tus-Resumable; wire-тест чутливий до env). Латентно з PR1 (session-repository.test тихо перейшов у happy-dom, пройшов). ФІКС генеральний у rstest.config.ts: `src/modules/*/adapters/**` додано в adapter-node include ТА dom exclude. Див. пам'ять rstest-project-globs-tied-to-flat-layout.
- ФАКТ: 773 passed/0 (84 файли, без дублів); check-architecture 214; tsc, oxlint, oxfmt зелені.
ДАЛІ: commit → ордерний make check → merge.
