
## Лінза 3 (AI-slop) — запуск
ДАЛІ: прочитати docs/playback/plan.md (ред.3) повністю, потім spec.md §5, _wip-phase6.md, AGENTS.md, .claude/standards.md, обидва звіти критики. Далі — механічна звірка КОЖНОЇ рядкової цитати з реальним кодом.

ФАКТ: план прочитано повністю (545 рядків). Попередній список ЗАЛИШКІВ скасованої куки (перевірю точні рядки далі):
- L216 D-3 «куки `courseforge_session` і `playback`»
- L325 D-7 «(мінт куки)»
- L333 D-7 «кладеться в тіло мінту сесії»
- L355 §4 «Нова кука `Path=/` … несе непрозорий id»
- L246 D-4 «`play` без чинної куки»
- L474 B13 «ключ містить `playback`-куку»
- L478 B17 «той самий id сесії; власна кука лишається чинною»
- L473 B12 «конфіг редакції 2», L537 §10 «після зеленого GATE 2 редакції 2» — стале посилання на ред.2
ДАЛІ: механічна звірка кожної рядкової цитати з кодом.

### Звірка цитат (партія 1)
- ФАКТ `sql/queries/media.sql:41-45` — СХОДИТЬСЯ: рядки 41-45 = FROM courses + три LEFT JOIN + WHERE. Прецедент прямого джойна реальний.
- ФАКТ `identity/adapters/http/handler.go:123,142` — СХОДИТЬСЯ: 123 = `func (handler *Handler) Authenticate`, 142 = `func (handler *Handler) RequireSameOrigin`.
- ФАКТ `identity/adapters/http/handler.go:125,245-247` — СХОДИТЬСЯ: 125 = `identityHeaders(w)`, 245-247 = `func identityHeaders` зі `Cache-Control: no-store`. Отже теза «кеш не вмикався б без proxy_ignore_headers» доказана.
- ФАКТ `docker-compose.yml:152-153` — СХОДИТЬСЯ: 152 `read_only: true`, 153 `tmpfs: [/tmp, /var/cache/nginx]`.
- ФАКТ `Makefile:54` — СХОДИТЬСЯ: рядок 54 = ціль `check:`, у ній `smoke` стоїть перед `e2e`.
- ФАКТ `infrastructure/tests/integration.sh:26-34` — НЕ СХОДИТЬСЯ ЧАСТКОВО: у файлі ЛИШЕ 32 рядки; блок docker run = 26-32. Суть (монтується лише backend/, nginx відсутній) правдива.
- ФАКТ `ffmpeg/transcoder.go:198-202` — СХОДИТЬСЯ точно; `duration < 0` відкидає лише відʼємне, нуль проходить.
- ФАКТ `transcoder.go` рядок 167 — СХОДИТЬСЯ: `if probe.DurationSeconds < 1 { seek = "0" }`, справді лише тумбнейл.
- ФАКТ `media/domain/source.go:40-46` — СХОДИТЬСЯ: `AcceptSource` 40-46, тривалість не перевіряє.
- ФАКТ `00003_authoring.sql:35` — СХОДИТЬСЯ: `CREATE TABLE course_lessons`; таблиці `lessons` у міграціях немає.
- ФАКТ `course/adapters/postgres/authoring.go:237` — рядок СХОДИТЬСЯ (`tx.Exec(ctx, deleteMissingAuthoringLessons, ...)`), але ФОРМУЛЮВАННЯ ПЕРЕБІЛЬШЕНЕ: запит `DELETE ... AND NOT (lesson.id = ANY($2))` видаляє лише уроки, ВІДСУТНІ в дереві, а не «рядки при кожному збереженні». «Порушення на першому ж редагуванні курсу» — неправда: лише при видаленні уроку, що має прогрес.
- ФАКТ `admin-video-queries.ts:69` = `refetchInterval:` — СХОДИТЬСЯ; `videoPollMs = 3_000` на рядку 12 — СХОДИТЬСЯ.
