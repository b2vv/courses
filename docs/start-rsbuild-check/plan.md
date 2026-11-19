# M1 — TanStack Start + Rsbuild compatibility check before an upgrade (2026-11-19)

Source: owner request 2026-11-14 (`work/AGENDA.md` §Owner-initiated, quoted from `cbb645a0`); AGENDA next move M1
(rebuild `fe61852a`). Owner go 2026-11-19 ("Go"). Branch `m1-start-rsbuild-check` from `main` @ `ddee61f0`.
Plan critique: ACCEPTED round 5 (Codex, `_wip-plan-critique.md`), 2026-11-19.
Size: Full (dependency change across the SSR/build path; shared runtime). The upgrade itself is **not** authorised by
this plan: it lands only if every check below passes, and then behind its own ordered `make check` (owner consent).

## 1. Facts (measured 2026-11-19, `npm view` + `frontend/node_modules`)

| Package | Installed (pinned) | Latest | Peers / engines of latest |
|---|---|---|---|
| `@tanstack/react-start` | 1.169.3 | 1.169.7 | `@rsbuild/core ^2.0.5` (optional), `vite >=8` (optional), react ≥18; node ≥22.12 |
| `@tanstack/react-router` | 1.170.45 | 1.170.46 | react ≥18; node ≥20.19. Start 1.169.7 depends on exactly 1.170.46 |
| `@tanstack/react-router-ssr-query` | 1.167.3 | 1.167.3 | `@tanstack/react-router >=1.170.33`, react-query ≥5.102 |
| `@rsbuild/core` | 2.2.2 | 2.3.3 | `core-js >=3` (peer); node `^20.19 \|\| >=22.12` |
| `@rsbuild/plugin-react` | 2.1.0 | 2.1.1 | `@rsbuild/core ^2.0.0` |
| `@rsbuild/plugin-tailwindcss` | 2.0.3 | 2.0.3 | `@rsbuild/core ^2.0.0` |

- Node: `engines` `>=24 <25` (`frontend/package.json:73`), image `node:24-bookworm-slim` (`frontend/Dockerfile:1`). All
  candidate engines admit Node 24. Router is pinned exactly by Start, so Router and Start move together.
- Start is wired as an Rsbuild plugin (`frontend/rsbuild.config.ts:4,12`, `@tanstack/react-start/plugin/rsbuild`);
  server entry `frontend/src/server.ts` re-exports `@tanstack/react-start/server-entry`; `frontend/server.mjs` serves it
  through srvx.
- **Version-pinned behaviour:** `frontend/server.mjs:30–31` records that Start 1.169.3 wraps route components in
  Suspense, so route render errors resolve as HTTP 200 with the error deferred to the client. A Start upgrade may change
  this; it is checked (C4), not assumed.
- nginx: page route `location /` has `proxy_buffering off` (`infrastructure/nginx/nginx.conf:287`); gzip lives only in
  `location ^~ /api/` (`nginx.conf:241–254`, `gzip_proxied any`, `gzip_vary on`, `gzip_types application/json text/vtt`).
  `infrastructure/tests/smoke.sh:213–236` already asserts gzip on `/api/` JSON with `Vary`, identity untouched, and **no**
  `Content-Encoding` on the page route.
- Runtime: Kiri engine is `aarch64`; it runs amd64 containers under emulation (`postgres:17.5` amd64 reports `uname -m`
  = `x86_64`, 2026-11-19). The buildx builder `kiri-build` lists only `linux/arm64`; whether it executes amd64 `RUN`
  steps is **unknown** until tried. No amd64 `node:24-bookworm-slim` image is present locally: the amd64 check needs a
  download (owner permission required).

## 2. Candidate

Exactly the "Latest" column as measured 2026-11-19 (frozen): Start 1.169.7, Router 1.170.46, Rsbuild core 2.3.3, plugin-react 2.1.1 (ssr-query and
plugin-tailwindcss unchanged). Exact pins stay exact. No other dependency is bumped; `package-lock.json` changes only by
what `npm install <exact pins>` resolves, reviewed in the diff.

## 3. Checks (each separate, each with its own evidence; baseline first) — r3 after critique rounds 1–2

Each runtime check C1–C4 runs **with the same instrument** on the current pins (baseline, `main` code) and on the
candidate. A check passes only when the candidate PASSES; equivalence with a failing or unproven baseline is not a PASS
(it is reported as a baseline fact and leaves the check NOT_TESTED/BLOCKED). Product fixture: seeded published course
`go-backend-foundations` (`backend/sql/seed.sql:5`), read-only. No data is created or changed.

**Probe image (diagnostic, never committed).** Product SSR routes cannot show progressive streaming: catalog/detail
loaders await their data before render, and the learn routes are `ssr: false` (round-2 R2-B1). So C2 and C4 use a
throwaway probe build of each tree (baseline and candidate) from one uncommitted patch, applied byte-identically to both
(patch sha256 recorded in the report). The patch adds two file routes (no leading `_`; URLs verified in the generated
route tree before use) and nothing else:
- `frontend/src/routes/m1-stream.tsx` → `/m1-stream?run=<id>`. Its **loader** validates `run` and returns
  `{ late: fetchLate(run) }` **without awaiting it** — one request-scoped promise, created once per request by the
  loader, whose value is the complete text body of `GET http://m1-delay:9000/hold?run=<id>` (headers and body are held
  by the service until release). The component renders `<h1 data-m1-shell="1">M1 shell</h1>` outside the boundary and,
  inside `<Suspense>`, Router `<Await promise={late}>` rendering `<p data-m1-late={value} />` (value as a single
  attribute, so React text separators cannot split the marker). Router's deferred-promise streaming carries the resolved
  value to the client; the route is only ever loaded by a direct server request (no client navigation to it), so the
  delay service is reached server-side only.
- `frontend/src/routes/m1-throw.tsx` → `/m1-throw`: its component throws `new Error('m1-render-probe')` unconditionally
  (server and client), so the server path and the visible route error component are both exercised.
- The delay service `m1-delay` (task-owned Node script + local uncommitted Compose override, on the app network only):
  `GET /hold?run=` registers the request and sends nothing until released, then `200 text/plain` body `late-<run>`;
  `GET /state?run=` → `{count, held, released}` immediately; `POST /release?run=` resolves every held request for that
  run and returns `{released: n}`. State is per run id; each variant uses a fresh id. **Terminal latch (R4-B2):** `release` marks the run terminal
  even before any hold; a hold arriving on a terminal run returns `200 late-<run>` at once (counted, so a late
  duplicate is still visible); repeat release is idempotent and returns the same `{released}` ack; `count` is
  monotonic. On service shutdown every pending hold is answered `503`. `fetchLate` uses the loader's `abortController`
  signal and a 10 s `AbortSignal.timeout`, so the SSR fetch is bounded either way.
Probe images are built per the disk policy below and removed afterwards; the patch, the delay script and the probe
script are kept under `docs/start-rsbuild-check/probe/` as evidence (not wired into the app or the gate).

- **C0 dependency fit.** Frozen candidate (§2 exact versions). Record `npm --version`, sha256 of
  `package.json`/`package-lock.json` before/after. `npm install` of the four exact pins, then `npm ci` from the new lock,
  **then** `npm ls --all` exits 0 and `npm ls @tanstack/react-router --parseable` shows **one physical path** shared by
  Start and ssr-query. Every changed lock entry is listed in the report with its reason; unexplained drift is rejected.
  Then `format:check`, `typecheck`, `lint`, unit tests and `npm run build` exit 0 from that tree.
- **C1 SPA navigation (product pages).** Playwright, chromium, through nginx, anonymous. Main-frame `document` and
  `pageerror` listeners registered **before** `goto('/courses')`. Assert catalog heading and the
  `go-backend-foundations` card; set `window.__m1 = 1`; click its link; assert URL `/courses/go-backend-foundations` and
  the course heading from the seed; `page.goBack()`; assert URL `/courses` and the catalog heading. After **each**
  transition `window.__m1 === 1`. Evidence: exactly one main-frame document request, zero page errors.
- **C2 progressive SSR through nginx (probe route).** Node probe (`net` socket, raw HTTP/1.1) to nginx `:8080`
  `GET /m1-stream?run=<id>`; `data/end/error` listeners attached before the request is written; bytes accumulated across
  chunk boundaries (chunked framing decoded; gzip variant decoded incrementally). Concurrently it polls
  `GET /state?run=<id>` every 50 ms. Protocol: t0 = first state with `held == 1` (deadline 3 s after dispatch, else
  NOT_TESTED); release is sent **unconditionally at t0 + 1 s**. PASS requires: (a) status 200 `text/html`;
  (b) `data-m1-shell` received **before** the release is sent; (c) `data-m1-late` absent before release; (d) after
  release, `data-m1-late="late-<id>"` within 3 s and body completion within 5 s; (e) final `count == 1` (no duplicate
  hold). A missing/duplicate hold, early end, service or socket error, absent late marker or truncated body fails
  closed. Every control call (`/state`, `/release`, final `/state`) has a 2 s timeout and a validated status/body; the
  whole run is capped at 15 s. Nested `finally`: stop polling/timers, send release (bounded; its failure is recorded),
  and destroy the page socket unconditionally even if release fails. Variants: identity and `Accept-Encoding: gzip`, each a fresh run.
  **Instrument control (R3-N1):** on the baseline probe image only, a diagnostic mutation of `server.mjs` that
  materializes the body (`new Response(await response.arrayBuffer(), response)`) before returning it must make **only (b)** fail
  with the same probe on a fresh run — one acknowledged hold/release, no shell before release, then 200 HTML with the
  shell, the exact late marker and a complete body (R4-N1). A crash/timeout/404 makes the control invalid, and C2
  NOT_TESTED. The control lives in its own image tag with its own delta (sha256 recorded); positive runs use the
  unmutated probe images. `proxy_buffering on` is run once as an observation, not as the
  discriminator. If the normal baseline fails (b), that is a baseline finding and C2 cannot PASS for the candidate either.
- **C3 compression scope (narrowed claim).** nginx layer only: existing `smoke.sh` gzip block (lines 226–236) on
  `/api/` catalog JSON above `gzip_min_length` — `Content-Encoding: gzip` + `Vary: Accept-Encoding`, identity uncompressed
  and decoding to the same JSON; page route (`/courses`, `/courses/go-backend-foundations`, and the C2 gzip variant)
  never has `Content-Encoding`; the raw `/courses` and detail responses are 200 `text/html`, complete, and carry the
  catalog / seeded course heading in the initial SSR bytes on both trees (R3-N2); and the C2 gzip variant keeps the before-release ordering. srvx static-asset behaviour
  is recorded as baseline only. VTT/media classes are NOT_TESTED (narrowed claim).
- **C4 render-error contract (`server.mjs:30–31`, probe route).** `GET /m1-throw` (the generated URL, verified) through nginx on both probe images,
  plus the unpatched `/courses` as control. Record: HTTP status and body completion; the server log line for that
  request (`ssr/request` vs `ssr/render_failed`) **and** proof the sentinel threw on the server (`m1-render-probe` in a
  fresh, bounded frontend container log interval taken before the browser visit or the serialized error in the HTML); in a browser, the route error component is visible.
  Baseline and candidate must match; any difference blocks the upgrade until `server.mjs` comment/logging and dependent
  tests are reconciled. Loader/API failures and the `entry.fetch` 503 path are out of scope.
- **C5 Linux amd64.** One `docker build --platform linux/amd64` of the real `frontend/Dockerfile` production target
  from the candidate tree (its stages run `npm ci`, `npm run build`, `npm ci --omit=dev`); stage logs kept, plus in the
  dependencies stage `npm --version`, lock sha256 and `npm ls @tanstack/react-router --parseable` (R2-N1). In the final
  image, installing nothing: `node -p "process.version+' '+process.platform+' '+process.arch"` = `v24.* linux x64`;
  `/healthz` 200 and `/courses` 200 with the catalog heading. Record the image ID; a local uncommitted Compose override
  runs **that exact image** (`platform: linux/amd64`, `pull_policy: never`, no build) as the frontend; verify the running
  container's image ID and architecture, then repeat C1 and C3 against it, and C2 against an amd64 build of the
  candidate probe image. If the builder cannot execute amd64 steps, C5 is BLOCKED with the exact error.
- **Disk (R2-N2).** Before **every** task-owned image build (probe images, amd64): record builder/VM free space; stop
  with BLOCKED below 10 GB free (a policy threshold, not a capacity guarantee); record after-build growth. No prune, no
  global context change; only task-owned tags/overrides/services are removed afterwards.

## 4. Decision rule

- All of C0–C5 PASS → the candidate pins are committed on this branch, the ordered `make check` runs (owner consent),
  and a PR is opened. Merge needs owner approval.
- Any FAIL → no upgrade; the report names the failing check, its evidence and the mechanism if found; pins stay.
- Any BLOCKED/NOT_TESTED → no upgrade claim; the report says what is missing and the owner decides.

## 5. Risks and limits

- Kiri disk: builder had 26.9 GB free before the R6 gate (2026-11-19 21:26). An amd64 image set adds a few GB plus
  cache; measure before C5, no prune without owner permission.
- Emulated amd64 is slow; C5 timings are not performance evidence, only behaviour.
- Patch/minor bumps: the release notes between the versions are read for Start/Router/Rsbuild changes touching SSR
  streaming, server entry, Suspense or the Rsbuild plugin; findings go into the report.
- Not in scope: other dependency upgrades, Node 25, Vite, production deployment.

## 6. Checkpoint (Full)
- Required questions: (1) owner permission to download the amd64 `node:24-bookworm-slim` image and amd64 npm packages
  for C5; (2) the ordered `make check` after C0–C5 (separate consent per run).
- Review: plan critique by an independent Codex run before any runtime check; independent Codex review of the diff
  before the gate. review_mode=independent_subagent (Codex). Writer Courses 1 (probes/diff via Codex or inline).
- Evidence lives in `docs/start-rsbuild-check/report.md` (derived records only; no raw Playwright traces).
