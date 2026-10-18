# A-батч — PR2: course → contracts/ (shared kernel)

## Рішення власника (2026-10-18) — задає патерн для всього A-батчу
`domain/course` — багатий спільний домен (442 р., 33 типи + 18 pure-функцій), споживаний 4+ фічами (catalog, admin, learn, storefront). check-architecture.mjs дозволяє крос-модульний доступ ЛИШЕ через `contracts/` (модуль не тягне ні чужий barrel, ні чужий domain). Власник обрав **A: `contracts/` = спільний domain-safe kernel**. Спільні домени → `contracts/<name>.ts`; модулі імпортують звідти. Закриває альтернативи «спільний немодульний domain/ kernel» і «course як повний модуль».

**Наслідок для порядку A-батчу:** домени, що є спільним словником (course; імовірно enrollment/media/identity-AuthenticatedUser), реалізуються як `contracts/`-екстракція, а не як товстий модуль. Тонкі I/O (repository/adapter) складаються у фіча-модуль-споживач, коли той мігрує.

## Що зроблено в PR2
- `git mv domain/course/course.ts(.test) → contracts/course.ts(.test)` (домен цілком pure, 0 імпортів).
- check-architecture.mjs рядок 59: flat `application/ports` тепер можуть тягти `contracts/` (domain-safe, inward — як domain).
- Репойнт 33 споживачів: `domain/course/course` → `contracts/course`.
- Репозиторій/порт (CourseRepository, course-repository adapter) ЛИШЕНО flat — складеться в catalog-модуль, коли catalog мігрує (рішення власника: «тонкий або в catalog»). admin-course-* лишається flat (→admin).

## ФАКТ
Швидкі перевірки зелені: check-architecture (211 files), tsc EXIT=0, oxlint, oxfmt --check, rstest contracts+catalog+admin 0 failures.

## Журнал
ДАЛІ: commit → ордерний make check → PR проти main.
