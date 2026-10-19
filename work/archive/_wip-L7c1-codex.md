ДАЛІ: Reviewing session: re-run the full gate, build, and L8 e2e; npm test was denied by sandbox port policy.

## Step 1 — briefing read

Прочитано `work/CTO-RESEARCH.md`, `work/brief-codex-L7c1.md` і `work/codex-working-model.md`. WIP-файл створено; заборони зафіксовано: без git, `work/tasks.md`, phase report і ручного редагування `generated/`; build/e2e лишаються reviewing session.

## Step 2 — plan and scope read

Прочитано `docs/enrollment/plan.md` для En07 і порядок T7. Межа підтверджена: L7c1 лише форми скарг; черга, resolve і revoke належать L7c2. Далі перевіряю наявний контракт і L7a/L7b-патерни перед реалізацією.

## Step 3 — contract and sibling-code inspection

Перевірено наявні enrollment domain/port/API adapter/query composition, CourseView, CourseRoster, їх routes, Dialog, форму grant і локалі. Порт та generated виклики для L7c1 існують, отже missing-contract stop не застосовується. Виявлено нижче неточність, що змінює runtime error mapping.

## Discussion

**Що не так.** Brief каже: «404 для цих НЕ існує — не мапь». Проте checked-in generated API contract декларує `404` для обох потрібних функцій: `frontend/src/adapters/api/generated/courses.ts:811-814` (`reportCourseLecturerResponse404`) та `:890-893` (`reportCourseParticipantResponse404`).

**Чому це спричинить баг.** Adapter за вказівкою brief не визнає 404 у `namedError`. Якщо backend поверне документований 404, клієнт деградує його через `unnamedStatusCode` замість явного, узгодженого domain error; UI покаже хибну generic-помилку або інший fallback, а не коректно обробить контракт. Це також суперечить вимозі звірити реальні статуси в generated.

**Пропоноване виправлення.** Reviewer має визначити авторитетний контракт: або (a) підтвердити, що 404 справді можливий і назвати domain error/code для нього, після чого я додам його до domain union, adapter map та UI; або (b) виправити OpenAPI/generated у reviewing session і підтвердити, що 404 неможливий, після чого я реалізую brief без 404. Не можу вручну редагувати `generated/`.

**Що зроблено тимчасово.** Зупинено до змін у коді; git, `work/tasks.md`, phase report і `generated/` не чіпав.

## Step 4 — reviewer correction accepted

Reviewer corrected the brief: both report endpoints declare and return `404 course_not_found`. I will add that named code and map all statuses declared by generated (201/400/401/403/404/409/503); no manual generated change is needed.

## Step 5 — implementation

Implemented the enrollment report port and API adapter via the existing transport; `201` validates a minimum object response and statuses `400/401/403/404/409/503` follow the generated endpoint contract (including `404→course_not_found`). Added no-invalidation report mutations. Added a reusable React Aria report dialog with a labeled 1–500-character TextArea, form validation, pending submit, local alert mapping, dismissal/focus semantics and success status. Wired it only for active/blocked course members and per roster participant, preserving block/unblock. Added Ukrainian and English report strings. No generated, backend, git, tasks, or phase-report edits.

## Step 6 — focused tests and verification

Added focused presentation tests for member-only lecturer reporting (including success status and revoked omission), roster participant reporting, and query behavior that reports do not invalidate enrollment state. `cd frontend && npm run typecheck` passed. `npm run lint` passed, including architecture boundaries (172 files). `npm test` was **not run**: rstest was denied by the sandbox while binding `::1:3000` (`listen EPERM: operation not permitted`); no pass is claimed. Build and e2e remain for the reviewing session as instructed.

## Step 7 — final static recheck

After marking the report TextArea explicitly required, reran `cd frontend && npm run typecheck && npm run lint`: both passed; architecture boundaries again reported 172 files. The earlier `npm test` denial remains the final test result.
