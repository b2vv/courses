ДАЛІ: A08a реалізовано; перед переданням лишився лише зовнішній прогін Rstest поза sandbox із дозволеним localhost bind.

2026-10-05 — старт A08a: прочитано AGENTS.md, CTO briefing та brief-codex-a08a.md; підтверджено гілку feat/accounts-selfservice. Робочий файл був відсутній, створено. Існуючі незв’язані зміни (скриншоти та інші WIP) не чіпаю.

2026-10-05 — додано A08a: presentation для register/verify/forgot/reset, file-routes з Router search-param token, remember-me у login, i18n uk/en і view-тести. `npm run build` штатно регенерував лише TanStack `routeTree.gen.ts`, щоб маршрути реально зареєструвалися; Orval і API-generated файли не чіпалися.

2026-10-05 — `npm run typecheck` і `npm run lint` пройшли. `npm run test` (також точковий запуск двох нових view-test файлів) не запускає тести в цьому sandbox: Rstest завершується на `listen EPERM ::1:3000` під час старту. `npm run build` пройшов; його наявне попередження про `route-guards.test.tsx` не є зміною A08a.

2026-10-05 — review diff: поправлено client-side form validation через `form.checkValidity()` до submit. Scope лишився A08a; settings, verification banner поза success-notice реєстрації та play-gate не додавались. Коміт навмисно не створено за брифом.
