# A-батч — PR1: modules/identity/ (leaf-first, foundation)

РОБОЧИЙ ФАЙЛ. Дописую після кожного кроку.

## Контекст
- V-трек завершено (8 views, PR #44 змерджено, main @ 80a167d). A-пілот `modules/document/` зроблено.
- Власник (2026-10-18): A-батч = модуль за модулем (1 PR/модуль), порядок leaf-first (за залежностями).
- Черга наступної задачі: після цього PR — `~/project-data/ai-factory-rollout/runs/036ccad-20261018T061121Z/README.md`.

## Карта залежностей (leaf-first порядок PR)
Листки (0 крос-залеж.): identity, course, discussion, enrollment, media, playback, purchase, transcript
1-залеж.: catalog→course, teach→enrollment
Останній (важкий): admin→{course,enrollment,identity,media,purchase,transcript}

Порядок: **identity (PR1)** → course → discussion → enrollment → media → playback → purchase → transcript → catalog → teach → admin.

## Правила check-architecture.mjs (виведено з читання)
- Авто-дискавер `modules/<m>/<layer>/` — реєстрація не потрібна.
- Module domain → лише `contracts/`. Module application/ports → той самий модуль або `contracts/` або пакети (НЕ flat-код, НЕ чужий barrel).
- Ззовні модуль лише через `index.ts` barrel (deep cross-module import заборонено); dynamic import() — виняток.
- flat application/ports → лише `(domain|ports|application)/` (рядок 59) — НЕ barrel, НЕ contracts (поки).

## Рішення (датовані, 2026-10-18)
- **Міграція від джерел (leaf=залежить-ні-від-чого) першою**, бо module-шар не може імпортувати flat-код (рядок 68) — доведено читанням check. Закриває альтернативу «sink-first».
- **`domain/identity/administration.ts` належить майбутньому `admin`-модулю, не identity**: усі 4 споживачі — admin-side (application/admin/users, ports/AdminUserRepository, adapters/api/admin-user-*). Переношу в `domain/admin/` (flat) ДО переносу identity, щоб identity-модуль був чистий. Закриває альтернативу «administration лишається в identity».
- **`contracts/` у PR1 НЕ створюю** (AGENTS.md: без невживаних seam). Після переносу жоден flat application/ports не споживає barrel identity. Шов `contracts/identity.ts` форсується першим модулем, що залежить від identity (admin) — тоді й заведу. Задокументований майбутній шов.

## План PR1
Переносяться В modules/identity/:
- domain/identity/identity.ts(.test) → modules/identity/domain/
- application/identity/session.ts(.test) → modules/identity/application/
- presentation/identity/* (13 файлів) → modules/identity/ui/
- ports/SessionRepository.ts → modules/identity/ports/
- adapters/api/session-repository.ts → modules/identity/adapters/
Переноситься З identity (→ flat admin):
- domain/identity/administration.ts(.test) → domain/admin/
Далі: barrel index.ts; фікс внутрішніх + зовнішніх (barrel) імпортів; typecheck → straggler loop; ордерний make check (нейтральність).

## Журнал
- git mv: identity→modules/identity/{domain,application,ui,ports,adapters}; administration→domain/admin/; session-schema(+test), session-repository.test теж у модуль.
- barrel modules/identity/index.ts (export * domain; createSession; SessionRepository type; createApiSessionRepository; 9 UI + 4 ProtectedRouteStates).
- Set A (13 файлів): administration-шлях flat→flat. Set B (30): identity-deep→barrel.
- ФАКТ: швидкі перевірки зелені — check-architecture (211 files OK), tsc --noEmit EXIT=0, lint EXIT=0, oxfmt --check OK (5 файлів доформатовано), rstest src/modules/identity 86 passed.
- Пастка: `git add -A` змів передісні чужі файли (AI-factory/editor-polish WIP) — відкотив, застейджив лише frontend/ + цей wip (64 файли).
ДАЛІ: commit → ордерний make check (нейтральність, ~16хв, скидає спільні дані) → PR.
