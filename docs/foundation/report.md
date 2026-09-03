# Foundation delivery evidence

## 1. Що зроблено

Phase 1 working-result gate passed on 2026-09-03: Go API/worker/Goose runner, Start + Rsbuild SSR shell, 20-component React Aria kit, PostgreSQL/Redis/Nginx Compose, health and graceful shutdown, CI/Makefile and architecture docs. See plan F01–F06; Phase 2 is next and is not claimed complete here.

## 2. Протестовані сценарії

| Scope | Command/evidence | Actual | Verdict |
|---|---|---|---|
| P01 SSR shell without JS and hydration/skip link | `npm exec playwright test e2e/shell.spec.ts` | 2 passed | pass |
| P02 UI controls | `npm test` | 10 Rstest tests passed | pass; browser layout coverage pending final gate |
| P03 Runtime and routing | `docker compose up -d --build --wait`; `sh infrastructure/tests/smoke.sh` | 6 healthy runtime services, migrate exited0; media403, metrics404, rate429 | pass |
| Backend runtime | backend-runtime-evidence.md | gofmt/vet/race + binaries + dependency recovery/shutdown passed | pass |
| Frontend static/build | format:check, lint, typecheck, build | passed; dist/client and dist/server/index.js emitted | pass |
| Phase 2 Course acceptance | pending | no Course implementation yet | not run |

## 3. Неочевидне

- npm11.6 produced a platform-incomplete optional-dependency lock. npm11.19 package-lock-only repaired it; Linux npm ci and production build then passed. Exact runtime versions remain unchanged.
- Unprivileged Nginx needed all temp paths under writable /tmp; default /var/cache/nginx failed at startup. Fixed in infrastructure/nginx/nginx.conf and verified.
- UI Select geometry is not available in Happy DOM. Keyboard opening/other critical interactions passed; full real-browser select/zoom checks belong to final browser gate.

## 4. Скріни

Screenshots pending catalog implementation. User supplied a CourseForge mockup URL; it currently requires ChatGPT login from available browser sessions. No visual parity claim made.

## 5. Проходка по user flow

Playwright opened homepage without JavaScript, followed catalog navigation, then verified hydration and keyboard skip-link focus in a JavaScript session. Catalog is a shell at this checkpoint.

## 6. Ризики й обмеження

Phase 2 remains. Identity, admin CRUD, media upload/transcoding/playback, transcript and orders intentionally deferred. Media denied; worker reports no handlers. Local-only HTTP and development placeholders are not public deployment configuration. No full product/DRM claims.

Post-deploy: not applicable; no public deployment performed.
