# Lens: Логічні діри (r2)

ДАЛІ: прочитати plan.md, spec.md (зроблено), далі звірити цитовані файли: server.ts, rsbuild.config.ts, transport.ts, docker-compose.e2e.yml, Makefile(e2e), image-authorize.sh.

ФАКТ: server-entry default = `ServerEntry = {fetch: RequestHandler}` (node_modules .../default-entry/esm/server.d.ts). B1 shape claim correct.
ФАКТ: REAL SSR process = frontend/server.mjs (`start`=`node server.mjs`, Dockerfile prod CMD). It imports `entry from ./dist/server/index.js` (= bundled src/server.ts) and ALREADY wraps `entry.fetch` in try/catch, ALREADY emits `render_failed` JSON on throw + `http_request` JSON with status on success.
ФАКТ: server.mjs comment: "Start 1.169.3 wraps route components in Suspense: route render errors resolve as HTTP 200 with the error deferred to the client." → route lane errors are NOT 5xx and NOT a throw at entry.fetch.
ФАКТ: Dockerfile prod COPYs server.mjs verbatim (not bundled) + dist/. So Rsbuild source.define does NOT transform server.mjs. define only reaches bundled src/* → dist.
ФАКТ: image-authorize.sh / playback-authorize.sh referenced only in docs; NOT in Makefile/smoke.sh/integration.sh/.github → never gated. `up` = `docker compose up` no --profile (AC7 verified).
ФАКТ: client.tsx StrictMode only; e2e/gate runs prod build → no StrictMode double-invoke at gate.
РІШЕННЯ: 3 blocking (SSR seam ignores server.mjs; route errors=200 break AC3; release-source incoherent), 3 non-blocking.
