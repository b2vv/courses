# Codex brief — grant user search: "searching" is not "no matches" (O05 slice 3b, nb 3)

Repository: /Users/strelia/projects/courses (branch feat/payment). Read AGENTS.md first; it binds.
Working file: append your steps to work/codex-o05-grant-loading.log as you go.

## Problem (from review, Courses 2, work/review-o05-grant-ui.md nb 3)
In the admin course editor's «Видати доступ до курсу» panel the user ComboBox runs in server
mode: options are the server's answer for the typed text, fetched after a 250 ms debounce. While
the debounce or the request is in flight the options are empty, so the list shows and a screen
reader announces «Немає відповідних варіантів» although the answer is still coming.

## Wanted
While a search is pending (typed text not yet searched, or the request in flight), the empty
list says «Шукаємо…» / "Searching…"; only a settled empty answer says «Немає відповідних
варіантів». Keyboard behaviour and everything else unchanged.

## Where
- `frontend/src/presentation/ui/controls.tsx` — shared `ComboBox` (server mode added in 3b,
  `renderEmptyState` is hard-coded there; note the client-mode text is also hard-coded Ukrainian —
  do not widen scope beyond an optional prop, e.g. `isLoading` + its label).
- `frontend/src/presentation/admin/CourseGrantPanel.tsx` — passes the state through.
- `frontend/src/routes/admin.courses.$courseId.tsx` — `CourseGrantTool` owns `typed`, the
  debounced `search` and the `useQuery`; pending = `typed.trim() !== search || users.isFetching`.
- i18n `frontend/src/app/providers/locales/{uk,en}.ts` under `administration.grant`.

## Constraints
- Layering per AGENTS.md (presentation gets state via props; no fetch in presentation).
- Strict TS, no `any`, no assertions. React Aria only.
- Do not change behaviour of client-mode ComboBox users (catalog filters).

## Done when
- unit test in `CourseGrantPanel.test.tsx` (or ui.test.tsx): pending → «Searching…» option/empty
  state, settled empty → no-matches text; existing tests green.
- `cd frontend && npm run format:check && npm run lint && npm run typecheck && npm test` green.
- Do NOT commit. Report changed files and the commands' results.

## Channels
If a contract you need is missing, stop and say so. If you think this brief is wrong in a way
that would cause a bug, argue it in your report instead of silently diverging.
