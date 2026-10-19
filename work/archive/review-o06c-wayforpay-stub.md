# Review request — O06c-auto: WayForPay stub + offline purchase e2e (Codex; Courses 1 gated)

Base `3b7213e`. Uncommitted. Brief: work/brief-codex-o06c-wayforpay-stub-e2e.md; Codex log:
work/codex-o06c.log. The live tunnel run (O06c-live, Pay25 real callback) is a separate manual
step — runbook work/runbook-o06c-tunnel.md, owner-run; not in the gate.

## Gates (Courses 1, full ordered chain incl. the new e2e)
`make format-check lint typecheck test integration build && make seed-identities seed && make e2e`
→ exit 0; e2e **89/89** (new `purchase-wayforpay.spec.ts` green, 4.3s).

## Seam (offline by construction — please confirm production is untouched)
- Backend `wayforpay.go`: `PayURL`/`StatusURL` now from `WAYFORPAY_PAY_URL`/`WAYFORPAY_STATUS_URL`,
  **defaulting to the production hosts** when unset. Only `docker-compose.e2e.yml` overrides them to
  the stub (browser POST → published `localhost:8090`; server CHECK_STATUS → `wayforpay-stub:8080`
  over the docker network).
- Frontend allowlist (`purchase-repository.ts`): `secure.wayforpay.com` always requires https; a
  second host is accepted **only** when `PUBLIC_WAYFORPAY_PAY_HOST` is non-empty AND matches AND the
  action is http. That env is threaded via the Dockerfile build ARG (default empty) and rsbuild
  `define` (`?? ''`), so a production build has it empty → the stub branch is dead. Confirm no path
  sets it in production.
- `Makefile` `e2e` target now raises `docker-compose.yml + docker-compose.e2e.yml` (api, frontend,
  wayforpay-stub, nginx) with `--wait` before `npm run e2e`.
- Stub: `frontend/e2e/wayforpay-stub/` (standard-library Go), signs with the stack's throwaway key
  `e2e-wayforpay-secret` — not the real published test-merchant key.

## The e2e asserts
Pay24 at the HTTP layer (valid signed callback vector applied; bad signature → 401, order stays
pending; replay → accept, no second grant); Pay25 shape against the stub («Купити» → «Перейти до
оплати» → stub redirects back → status page polls → signed CHECK_STATUS → same ApplyPaymentEvent →
lesson opens); the grant comes via the status check / callback, not the return (A5).

## Please look hardest at
1. Can `PUBLIC_WAYFORPAY_PAY_HOST` leak into any production build path (compose, CI, Dockerfile)?
2. Does the stub ever let the mere browser return grant access (A5 must hold — grant only on
   callback/CHECK_STATUS)?
3. The new `e2e` Makefile target vs the existing comment that `e2e` is the low-level target `check`
   composes — does raising the overlay here break any other caller of `make e2e`?

## Findings (Courses 2 appends below)

---
## Findings — Courses 2 (claude-opus-4-8, 2026-10-01)

ДАЛІ: (1) PUBLIC_WAYFORPAY_PAY_HOST не протікає в прод (allowlist→rsbuild define→Dockerfile ARG,
default порожній → stub-гілка мертва); backend PayURL/StatusURL default на прод; (2) стаб не видає
право поверненням (A5); (3) Makefile e2e не ламає інших викликачів. Трасую ланцюг env + compose.e2e + Makefile.

### F1 [#1 PUBLIC_WAYFORPAY_PAY_HOST / URL leak] — НЕ ПРОТІКАЄ (SOLID)
Ланцюг: allowlist (purchase-repository) — secure.wayforpay.com завжди https; друга-гілка лише коли
e2ePayHost!=='' && hostname===e2ePayHost && http. rsbuild define інлайнить build-time значення
(process.env...?? '') → запечено в JS, runtime не зверне. Dockerfile ARG PUBLIC_WAYFORPAY_PAY_HOST=
(порожній default). Backend wayforpay.go: PayURL/StatusURL з env, ПУСТО→default на прод-константи
(secure/api.wayforpay.com). Прод docker-compose.yml НЕ ставить ні build-arg, ні WAYFORPAY_PAY_URL/
STATUS_URL. Усі три override — ЛИШЕ в docker-compose.e2e.yml (PAY_URL=localhost:8090, STATUS_URL=
wayforpay-stub:8080, build-arg PUBLIC_WAYFORPAY_PAY_HOST=localhost). Прод-білд (target: production)
arg не передає → '' → stub-гілка мертва. Витоку немає.

### F2 [#2 стаб не видає право поверненням] — A5 ТРИМАЄТЬСЯ
Стаб /pay (browser POST): parse form → записує order у пам'ять → http.Redirect(returnURL, 303). НЕ
кличе наш бек, НЕ грантить, НЕ шле callback. Grant лише через /api (CHECK_STATUS): стаб верифікує
підпис ЗАПИТУ (merchant;orderReference), відповідає ПІДПИСАНИМ Approved; наш бек верифікує відповідь →
ApplyPaymentEvent. Тобто повернення браузера інертне; право — лише від підписаного CHECK_STATUS/callback.
A5 by construction (бек-логіка та сама, стаб лише стоїть за мережеві ендпоінти WayForPay).

### F3 [#3 Makefile e2e overlay] — функц. ок, але КОМЕНТАР застарів (nb)
check: ...build up drain-queue seed-identities seed smoke e2e. up(рядок9) піднімає базовий стек
(postgres/redis/worker/…); e2e тепер ПЕРЕпіднімає api/frontend/nginx+wayforpay-stub з overlay. Postgres
не в списку e2e-up → seed зберігається, recreate api конектиться до того ж seeded postgres. check і
e2e-local (up seed-identities seed e2e) обидва працюють (gate exit 0). Жодного іншого викликача e2e нема.
[nb] Коментар рядки 47-53 тепер НЕТОЧНИЙ: каже «e2e ... neither raises the stack», а диф додав саме
`docker compose up` в e2e. Тепер e2e ПІДНІМАЄ api/frontend/stub/nginx (+--wait знімає 503-ризик з
коментаря), але все ще НЕ сидить → e2e-local лишається standalone-входом. Оновити коментар, щоб не
вводив в оману (репо саме про точність коментарів — цей і є застереження про обхід правил).

### Нотатка — що доводить O06c-auto (не over-claim)
Стаб підписує THROWAWAY-ключем e2e-wayforpay-secret (не реальним тест-мерчант). Тож e2e доводить
 WIRING (checkout→redirect→status-poll→signed CHECK_STATUS→той самий ApplyPaymentEvent→урок; Pay24 на
HTTP: валідний вектор, bad sig→401 order pending, replay→accept без 2-го гранту) із самоузгодженим
ключем. Він НЕ доводить відповідність РЕАЛЬНОМУ WayForPay-підпису (наш adapter і стаб могли б ділити
однаковий хибний порядок полів і все одно збігтися). Реальну відповідність дають: (a) пінований
wiki-вектор 4941ea0c з O06a, (b) ЖИВИЙ тунельний прогін (O06c-live, owner-run). За AGENTS «фаза
закривається лише після O06: доказ через sandbox, не фейк» — отже фазу закриває САМЕ тунельний прогін,
а не цей авто-гейт. Поділ записано в assumptions.md (рішення власника) — коректно scoped.

## ПІДСУМОК: 0 blocking, 1 nb (застарілий Makefile-коментар)
Шов offline-by-construction: стаб-override лише в compose.e2e, прод чистий, frontend build-time-запечений,
backend default на прод; A5 тримається (стаб /pay лише редиректить, grant лише підписаний CHECK_STATUS);
Makefile функц. ок (postgres не губиться). nb: оновити коментар e2e (тепер піднімає стек). Фазу закриває
живий тунельний прогін (O06c-live), не цей авто-гейт — як і записано.
