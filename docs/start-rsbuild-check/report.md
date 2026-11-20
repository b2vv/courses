# M1 — TanStack Start + Rsbuild compatibility check — report (2026-11-20, r2)

Plan: `docs/start-rsbuild-check/plan.md` (critique ACCEPTED r5). Owner permission for the downloads (candidate npm
packages, amd64 `node:24-bookworm-slim`): 2026-11-20 ("Так, усе"). All runs on Kiri, local stack on `:8080`, in isolated
git worktrees of `12657c14` (the shared working tree and its `node_modules` were not touched). Probe sources, patches
and derived results: `docs/start-rsbuild-check/probe/` (no raw traces, no credentials; auth came from the e2e `setup`
project's storage state; absolute paths replaced by `<repo>`, `<kiri>`, `<scratch>`, `<home>`).

This is revision r2: it follows the independent Codex review r1 (findings F1–F7, working file
`_wip-report-review.md`, both axes BLOCK). Every runtime check was rerun (`probe/m1-r2.sh`) with fixed instruments and
an identity record for every build and swap; C5 was rerun as a real amd64 build of the real Dockerfile
(`probe/m1-c5-rosetta.sh`). Primary evidence: `probe/results/r2/` (r1 files in `probe/results/` are kept as history).

## Verdict

**All of C0–C5 PASS for the candidate under the plan, with C1 as amended by D1 (owner-accepted 2026-11-20).** Per plan
§4 the pins were committed on this branch (`3d8c6223`, owner consent 2026-11-20) and the ordered `make check` ran green
at `c2074e6a` (§Gate), then a PR.
Deviations re-run in r3 (owner, 2026-11-20: "розбери щоб запускалось; перезбери"): the two preliminary C5 builds were
rebuilt with per-build disk samples (13 → 13 each), and the manual amd64 build that exited 1 in r2 was rebuilt 6 times
with every build log kept — 6/6 exit 0; the r2 exit 1 did not reproduce (§C5, §Environment). Only r1's superseded
builds stay without per-build samples; no r1 result is used for a verdict.

| Check | Baseline (1.169.3 / 1.170.45 / Rsbuild 2.2.2) | Candidate (1.169.7 / 1.170.46 / Rsbuild 2.3.3) |
|---|---|---|
| C0 dependency fit | npm ci + `npm ls --all` 0 (macOS) | **PASS** (macOS); linux `npm ls --all` 1 `invalid`, named risk |
| C1 SPA navigation (D1, owner-accepted) | admin Router-Link path PASS; catalog = 3 documents | admin **PASS**; catalog = 3 documents (same) |
| C2 progressive SSR through nginx | PASS identity + gzip; control fails only (b) | **PASS** identity + gzip |
| C3 compression scope | PASS (h1/h2 parsed from body) | **PASS** |
| C4 render-error contract | 200, `http_request`, sentinel in log, error UI | identical → **PASS** |
| C5 linux/amd64 (real Dockerfile) | QEMU builder: `npm run build` segfault (139) | **PASS** via Rosetta builder; C1–C4 on that image |

## Decisions taken during the run

- **D1 (2026-11-20, Courses 1; accepted by the owner 2026-11-20, "D1 так"): C1 instrument amended.** The plan's
  catalog → detail → back path is a full document navigation **on the baseline**: three main-frame document requests,
  `window.__m1` lost after the click (`results/r2/c1-catalog-base.out`). Mechanism: catalog cards render the UI `Link`
  (`frontend/src/presentation/ui/controls.tsx:19–24`, a wrapper of `Aria.Link`, used at
  `modules/catalog/ui/CatalogView.tsx:3,93,109`), and production code wires no React Aria `RouterProvider`
  (`frontend/src/router.tsx:57` supplies only the `I18nProvider` wrapper; the only `RouterProvider` imports in
  `frontend/src` are TanStack's, in tests, e.g. `CatalogView.test.tsx:8`), so the link is a plain `<a>`. That path
  cannot test SPA navigation on either tree. C1 therefore runs the same instrument (main-frame document listener and
  `pageerror` before `goto`, `window.__m1` marker, exact `h1` per step) over the product path that does use Router
  `<Link>`: the admin shell nav (`modules/admin/ui/AdminShell.tsx:3,78`) `/admin` → `/admin/analytics` →
  `/admin/users` → back, read-only, admin storage state from the e2e `setup` project. **Coverage difference accepted
  with it:** the admin routes are `ssr: false` (`routes/admin.tsx:7`, `admin.analytics.tsx:8`), so C1 proves client
  Router navigation, not hydration of an anonymous SSR catalog page. The catalog path is kept as a recorded fact on
  both trees. Checkable: `probe/c1-admin-spa.mjs` (`:51–54` require one document and marker 1 after every step),
  `probe/c1-spa.mjs`.
- **D2 (2026-11-20, Courses 1): the candidate lock is generated in linux.** `npm install` with the host npm 11.6.2
  (macOS) rewrites even the unchanged baseline lock: it drops `@floating-ui/dom` (needed by the installed
  `@tiptap/extension-bubble-menu`/`floating-menu`) and `@emnapi/core`/`runtime`, and `npm ls --all` then fails with
  `missing` (measured: `npm install --package-lock-only` on the untouched baseline, 82 additions / 95 deletions = 177
  changed lines by `git diff --no-index --numstat`; `results/r2/supporting-records.txt` §D2). This is the known macOS-lock trap (`Domain/_gotchas.md`,
  2026-10-10 S0; also `work/tasks.md:412,456`). The candidate lock was generated with
  `npm install --package-lock-only --ignore-scripts --save-exact <pins>` inside `node:24-bookworm-slim` (npm 11.19.1)
  from the baseline lock; it changes only candidate-caused entries (inventory below).
- **D3 (2026-11-20, owner asked for a real amd64 build, then "Спочатку спробувати змусити збирач Docker працювати через
  Rosetta"): C5 runs on a task-owned buildx builder.** Mechanism of the two failed instruments, below in C5. The
  builder is a temporary `docker-container` builder on the `kiri` endpoint whose BuildKit image is the local
  `moby/buildkit:buildx-stable-1` with `/usr/bin/buildkit-qemu-*` deleted (`probe/c5-buildkit-noqemu.Dockerfile`),
  so the image carries no emulator of its own. The VM's only registered binfmt handler is `rosetta`
  (`interpreter /run/rosetta/rosetta`, flags `OCF`; `results/r2/c5-env-binfmt-mirror.txt`), so the amd64 steps ran
  through it — an inference from that handler list plus the observed `x86_64` / ELF `3e 00` results; no per-process
  interpreter trace was taken. `kiri-build`, its config and the global Docker
  context were not touched; the builder was removed afterwards. Checkable: `probe/m1-c5-rosetta.sh`,
  `results/r2/c5-rosetta.log`.

## C0 — dependency fit

- Host: npm 11.6.2, node v24.11.1. Linux lock generator: npm 11.19.1, node v24.22.3 (`node:24-bookworm-slim`, arm64).
- sha256 before: `package.json` `a09783e4…df37c8`, `package-lock.json` `9d80f637…886b35`.
  After: `package.json` `5c5bcd6f…d1fc43303`, `package-lock.json` `7dedaf02…f66831`.
- `package.json`: exactly the four pins (Router 1.170.46, Start 1.169.7, Rsbuild core 2.3.3, plugin-react 2.1.1).
- Lock inventory (computed independently by the reviewer and matching this list): 788 → 809 package entries;
  **21 added, 0 removed, 26 changed** (root pin metadata + 21 version changes + 4 metadata-only). No top-level lock
  field outside `packages` changed.
  - **Pinned:** `@tanstack/react-start` 1.169.3→1.169.7, `@tanstack/react-router` 1.170.45→1.170.46,
    `@rsbuild/core` 2.2.2→2.3.3, `@rsbuild/plugin-react` 2.1.0→2.1.1.
  - **Start's exact internal pins:** `react-start-client` 1.168.43→1.168.45, `react-start-server` 1.167.51→1.167.52,
    `react-start/node_modules/@tanstack/react-start-rsc` 0.1.64→0.1.68, `router-core` 1.171.38→1.171.39,
    `router-generator` 1.167.45→1.168.0, `router-plugin` 1.168.47→1.169.0, `router-utils` 1.162.4→1.163.0,
    `start-client-core` 1.170.38→1.170.39, `start-plugin-core` 1.171.54→1.172.0, `start-server-core`
    1.169.44→1.169.45, `start-storage-context` 1.167.40→1.167.41.
  - **Required by a changed dependency edge (raised minimum):** `unplugin` 3.4.0→3.7.0 (router-plugin), which adds
    `@rolldown/pluginutils` 1.1.1 (unplugin → pluginutils edge); `react-refresh` 0.18.0→0.19.0 (plugin-react 2.1.1
    requires `^0.19.0`); `@swc/helpers` 0.5.23→0.5.24 (Rsbuild 2.3.3 requires `^0.5.24`; 0.5.24 also drops its `tslib`
    dependency).
  - **Compatible re-resolution, not forced:** `exsolve` 1.1.3→1.2.0 (the new `^1.1.3` was already met by 1.1.3),
    `ansis` 4.4.0→4.5.0 (range stays `^4.3.1`, lock:5051), `vitefu` 1.1.3→1.1.4 (range stays `^1.1.3`, lock:5106).
    npm re-resolved these within their ranges while rewriting the affected subtree.
  - **Nested Rspack 2.3.3 (20 added entries under `@rsbuild/core/node_modules/`):** Rsbuild 2.3.3 depends on
    `@rspack/core ~2.3.3`, so npm nests `@rspack/core` and `@rspack/binding` 2.3.3, **14** `@rspack/binding-*`
    platform packages 2.3.3 (darwin-arm64, darwin-x64, linux-arm64-gnu, linux-arm64-musl, linux-ppc64-gnu,
    linux-riscv64-gnu, linux-riscv64-musl, linux-s390x-gnu, linux-x64-gnu, linux-x64-musl, wasm32-wasi,
    win32-arm64-msvc, win32-ia32-msvc, win32-x64-msvc), and the wasm fallback support `@napi-rs/wasm-runtime` 1.2.7,
    `@emnapi/core` and `@emnapi/runtime` 2.0.0-beta.3, `@emnapi/wasi-threads` 2.2.0 (lock:3158). The top-level
    `@rspack/core` stays 2.2.9 because `@rstest/core`'s nested Rsbuild 2.2.14 requires `~2.2.x`.
  - **Metadata only:** `libc` added to the four `lightningcss-linux-{arm64,x64}-{gnu,musl}` entries (written by the
    linux npm).
- `npm ci` from the new lock: exit 0 on macOS and in `node:24-bookworm-slim` (arm64 and, in C5, amd64).
- `npm ls --all`: exit 0 on macOS (the plan's C0 instrument). **Linux observation (named risk, not a C0 criterion):**
  exit 1 with one `invalid` —
  `@rspack/binding-linux-arm64-musl@2.2.9 invalid: "2.3.3" from node_modules/@rsbuild/core/node_modules/@rspack/binding`.
  Mechanism (reproduced by the reviewer with `require.resolve` on the read-only linux tree): the nested 2.3.3 musl
  binding carries `libc: ["musl"]` and is correctly skipped on glibc, so `npm ls` resolves that optional name upward to
  the hoisted 2.2.9 musl binding (installed on glibc because its lock entry has no `libc`). It is never loaded on
  glibc: Rspack's `binding.js:223–242` picks gnu vs musl by `isMusl()`, and `require.resolve` from the nested binding
  returns the nested **2.3.3 gnu** package. The baseline has a single Rspack copy and exits 0 in linux too.
- `npm ls @tanstack/react-router --parseable`: one physical path (`node_modules/@tanstack/react-router`) on macOS, in
  linux arm64, and in the C5 amd64 `dependencies` stage; Start, its client/server/rsc packages and
  `react-router-ssr-query` all resolve it (deduped).
- Gates from the candidate tree (macOS), baseline in brackets, each recorded `RC=0` in the task's gate log
  (summary preserved in `results/r2/supporting-records.txt` §C0):
  `format:check` 0 (0), `typecheck` 0 (0), `lint` 0 (0), `test` 1482 passed / 0 skipped (1482 / 0), `build` 0 (0);
  client bundle 2835.9 kB (2844.2 kB). `routeTree.gen.ts` is byte-identical after the build.
- **Named risk (not a failure):** with the candidate, the build compiler is Rspack 2.3.3 while the plugins that
  peer-resolve `@rspack/core` (plugin-react-refresh, `@tailwindcss/webpack`, react-start-rsc, unplugin,
  ts-checker-rspack-plugin) get the hoisted 2.2.9. The build, tests, SSR and the runtime checks below pass with it.

## Identity records (r2)

Each runtime image build of r2 and of the Rosetta run (the images in the table below, plus `m1-deps:amd64-rosetta` and the buildx attempt)
records free space before and after, exit code, image ID and `os/arch` (`results/r2/run.log`,
`results/r2/c5-rosetta.log`); the full build inventory, including builds with no disk sample, is in §Environment; every swap records the built ID, the running container's image
ID (`match=yes` for all 7 swaps: 5 in `run.log`, 2 in `c5-rosetta.log`; the 2 manual-image swaps planned in
`m1-r2.sh:115–120` did not run because the script stopped at the manual build exit 1), health, in-container `uname -m` + `process.arch`, `/healthz` and `/courses` (200 for
all). The frontend was swapped only through a local, uncommitted Compose override (`probe/m1-override.yml`:
`pull_policy: never`, platform per swap).

| Image | Tree | ID | os/arch | in container |
|---|---|---|---|---|
| `m1-front:base` | baseline (clean) | `dbb51755…` | linux/arm64 | aarch64 arm64 |
| `m1-front:cand` | candidate (clean) | `3fc46d96…` | linux/arm64 | aarch64 arm64 |
| `m1-probe:base` | baseline + probe patch | `ea005b05…` | linux/arm64 | aarch64 arm64 |
| `m1-probe:cand` | candidate + probe patch | `7d06b538…` | linux/arm64 | aarch64 arm64 |
| `m1-probe:control` | baseline + probe + control delta | `cabeb270…` | linux/arm64 | aarch64 arm64 |
| `m1-front:amd64-rosetta` | candidate (clean), Dockerfile | `0c91fd30…` | linux/amd64 | x86_64 x64 |
| `m1-probe:amd64-rosetta` | candidate + probe, Dockerfile | `8e3ab567…` | linux/amd64 | x86_64 x64 |

## C1 — SPA navigation

Same scripts against the clean images (r2): baseline, candidate, candidate amd64.

- Admin Router-Link path: baseline **PASS**, candidate **PASS**, candidate amd64 **PASS**. Each: one main-frame
  document request (`/admin`), zero page errors, `window.__m1 === 1` after each of the four steps, `h1` `Курси` →
  `Аналітика` → `Користувачі` → `Аналітика` (`results/r2/c1-admin-*.out`).
- Catalog path (fact, D1): all three issue three document requests (`/courses`, `/courses/go-backend-foundations`,
  `/courses`), marker lost after the first click, zero page errors (`results/r2/c1-catalog-*.out`). Unchanged by the
  upgrade.

## C2 — progressive SSR through nginx (probe route `/m1-stream`)

Probe images from one patch (`probe/m1-probe.patch`, sha256
`2b6bb0f2a8eb1c212bc89a19b2dd547a279ff25f376fd8c0358a4177d965e3b2`, identical for both trees; generated route ids
`/m1-stream` and `/m1-throw` confirmed in `routeTree.gen.ts`; both trees' `server.mjs` equal `HEAD`). Raw `net` probe
from a container on the compose network to `nginx:8080`.

**Instrument fixed after review F6** (`probe/c2-probe.mjs`): a hard 15 s cap (`CAP_MS`, `checkCap()` in every wait and
poll), a 3 s connect deadline from dispatch, schema validation of every control body (`count`/`held` integers,
`released` boolean, release ack `released` integer), fail-closed on any duplicate, invalid body, bad status or cap
overrun, and a socket destroy timer in the `finally`.

Times in ms from probe start (r2, `results/r2/c2-*.out`):

| Run | headers | shell | held (t0) | release | late | end | criteria a–e |
|---|---|---|---|---|---|---|---|
| base identity | 45 | 45 | 98 | 1098 | 1101 | 1101 | all true → PASS |
| base gzip | 30 | 30 | 34 | 1039 | 1043 | 1043 | all true → PASS |
| **cand identity** | 45 | 45 | 94 | 1097 | 1107 | 1107 | all true → **PASS** |
| **cand gzip** | 31 | 31 | 86 | 1089 | 1096 | 1097 | all true → **PASS** |
| control (base + body materialized, `probe/m1-control.patch`, delta sha256 `87042703…a34e536e`) | 1090 | 1091 | 85 | 1087 | 1091 | 1091 | only **b** false → control valid |
| **cand amd64 identity** | 69 | 69 | 97 | 1098 | 1108 | 1108 | all true → **PASS** |
| **cand amd64 gzip** | 42 | 43 | 82 | 1092 | 1108 | 1108 | all true → **PASS** |

Every run: status 200 `text/html; charset=utf-8`, `Transfer-Encoding: chunked`, no `Content-Encoding`, release ack
`{released: 1}`, final state `{count 1, held 0, released true}`, complete body. The control proves the probe detects a
buffered body (headers and shell only after release). r1 observation kept: an nginx copy with `proxy_buffering on`
still forwarded the shell before release (`results/c2-buffered.jsonl`), which is why the plan uses the `server.mjs`
control as the discriminator.

## C3 — compression scope

**Instrument fixed after review F2:** the r1 grep could pass on the document `<title>` or serialized loader data
(both routes set the course/catalog title independently: `routes/courses.index.tsx:22`, `courses.$slug.tsx:25–28`).
`probe/c3-headings.py` now parses the raw response with `HTMLParser` and collects `h1`–`h3` text outside
`head`/`script`/`style`/`template`/`title`; `probe/c3-compress.sh` asserts the exact headings.

Baseline, candidate and candidate amd64 identical (`results/r2/c3-*.out`): `/api/v1/courses?limit=20` with gzip → 200,
`gzip`, `Vary: Accept-Encoding`, 1161 bytes; identity → 200, no encoding, 2626 bytes; decoded gzip equals identity JSON.
`/courses` → 200 `text/html`, no `Content-Encoding`, complete, body headings start
`h1 Каталог курсів`, `h2 Go: фундамент надійного бекенду`; `/courses/go-backend-foundations` → 200, complete, `h1
Go: фундамент надійного бекенду`. The C2 gzip variants carried no encoding and kept the before-release order.

**srvx static-asset baseline (observation, plan C3; review F7):** `/assets/js/index.<hash>.js` through nginx with
`Accept-Encoding: gzip` → 200 `text/javascript; charset=utf-8`, `Content-Encoding: gzip` on baseline (571 337 bytes)
and candidate (567 358 bytes; same on amd64). srvx's static middleware compresses static assets by default on both
trees; not a candidate change. VTT/media: NOT_TESTED (narrowed claim).

## C4 — render-error contract (probe route `/m1-throw`)

Baseline, candidate and candidate amd64 probe images, identical (`results/r2/c4-*.out`, `c4-browser-*.out`):
`GET /m1-throw` → 200, complete document; SSR log line `{"event":"http_request","status":200,"route":"/m1-throw"}`
(no `render_failed`); the bounded log interval taken around the curl, before the browser visit, contains
`Error in renderToReadableStream: Error: m1-render-probe` (the sentinel threw on the server); the sentinel is not in
the HTML; in chromium the root `PageError` alert is visible (`h1` `Не вдалося відкрити сторінку`, one retry button),
no page errors. Control `/courses` → 200 `http_request`. The `server.mjs:30–31` comment ("route render errors resolve
as HTTP 200") still holds for 1.169.7.

## C5 — Linux amd64 (real `frontend/Dockerfile`)

**Two instruments that cannot run on Kiri (observed, both trees):**

1. `docker build --platform linux/amd64` on `kiri-build` (the plan's literal command): exit 0, image labelled
   `linux/amd64`, but the content is arm64 (`uname -m` `aarch64`, `process.arch` `arm64`, ELF `e_machine` of
   `/usr/local/bin/node` `b7 00`). Observed: BuildKit warns `InvalidBaseImagePlatform: Base image
   node:24-bookworm-slim was pulled with platform "linux/arm64", expected "linux/amd64"` and continues
   (`results/r2/build-m1-front_cand-amd64-buildx.excerpt.log`); `kiri-build` advertises `linux/arm64` only, and Kiri
   writes a `docker.io` registry mirror into its config (`results/r2/c5-env-binfmt-mirror.txt`). **Inferred, not
   isolated:** the mirror is why the wrong base was selected (the temporary builders below, without the mirror, got
   the real amd64 base); the mirror's manifest list was not inspected.
2. A temporary `docker-container` builder without the mirror and with `--platform linux/amd64` pulls the real amd64
   base, but BuildKit does not detect Rosetta and runs amd64 steps through its bundled QEMU
   (`/dev/.buildkit_qemu_emulator`): `npm ci` passes, `npm run build` dies with `Segmentation fault`, exit 139 — **for
   the baseline as well** (candidate: `results/r2/supporting-records.txt` §c5-real, `c5-build-cand-qemu.excerpt.log`;
   baseline: `c5-base-qemu.console`, `c5-build-base-qemu.excerpt.log`). Both trees failing the same way on that path
   strongly implicates the bundled-QEMU execution, not the candidate; it is not proven to be the only cause.

**Instrument used (D3) — PASS.** Builder `m1-rosetta` (BuildKit v0.35.1 from the local `buildx-stable-1` image, 0
`qemu` files; `kiri-build` is v0.34.2), platforms `linux/amd64, linux/arm64`. Sanity build first: `uname -m` `x86_64`,
`v24.22.3 x64`, ELF `3e 00`. Then `docker buildx build --platform linux/amd64 --load` of the real Dockerfile
(`results/r2/c5-rosetta.log`, build excerpts alongside):

- `dependencies` target from the clean candidate: exit 0, image `e275df2e…` linux/amd64; in it npm 11.19.1, lock
  sha256 `7dedaf029ed966cd…`, `npm ls @tanstack/react-router --parseable` → `/app/node_modules/@tanstack/react-router`
  (one path), nested `@rspack/binding-linux-x64-gnu` ELF `3e 00`.
- `production` target (stages `npm ci`, `npm run build`, `npm ci --omit=dev`) from the clean candidate: exit 0, image
  `0c91fd30…` linux/amd64, `built in 7.83s (client)`, `built in 2.05s (ssr)`; in the image, installing nothing:
  `x86_64`, `v24.22.3 linux x64`, node ELF `3e 00`.
- `production` target from the candidate probe tree: exit 0, image `8e3ab567…` linux/amd64, same content facts.
- Swap `m1-front:amd64-rosetta` (`platform: linux/amd64`): running ID = built ID, healthy, in container `x86_64 x64`,
  `/healthz` 200, `/courses` 200 with the catalog `h1` (C3 parser). **C1** admin PASS, catalog 3 documents (as on
  arm64); **C3** PASS.
- Swap `m1-probe:amd64-rosetta`: running ID = built ID, `x86_64 x64`; **C2** identity and gzip PASS (table above);
  **C4** identical to arm64.

**Superseded supplementary evidence.** r1's manual run of the Dockerfile's commands in a real amd64 container
(`probe/c5-manual.sh`, `results/c5-manual-amd64*.log`) is no longer used for any verdict. Review F3 was right that it
did not capture the build's exit code; r2 captured it with `pipefail` and one manual `npm run build` exited **1** with
its log not kept (`results/r2/run.log`, `manual-clean.log`). Repeats of the same manual build: 4/4 `BUILD_EXIT=0` for
the candidate and 4/4 for the baseline (runs 1–3 in `results/r2/repeat-summary.txt`, the 4th in
`supporting-records.txt` §repeats); the 3 Rosetta Dockerfile builds above also exit 0. The single exit 1 is
unexplained and recorded as a limitation, not attributed.
**r3 rebuild (2026-11-20, `probe/m1-manual-repro.sh`, `results/r3/manual-repro.log`).** The r2 `manual()` steps
verbatim (`pipefail`, same amd64 `node` digest, same `m1-amd/frontend` source, `/out` mount), 6 runs; the
`npm run build` logs are retained (`results/r3/manual-build-run1..6.log`), the `npm ci` step is recorded by its exit
(6/6 `exit=0` in `manual-repro.log`), its log not retained: 6/6 `npm run build exit=0`, each with both
`built in … (client)` and `(ssr)` lines (client 5.26–13.4 s), disk 13 → 13 GB every run. The r2 failing log had only the
`(ssr)` line, which suggests (inference, not isolated — the log was discarded) that the client compilation failed; no
failure recurred in 6 + 8 + 2 = 16 later `npm run build` runs (the third Rosetta target, `dependencies`, runs no build). A
plausible mechanism (the Kiri balloon squeezing guest memory, `_gotchas.md` 2026-11-17) is a hypothesis, not tested.
**Position (Courses 1, 2026-11-20, for the owner to
overrule):** it is a superseded diagnostic outside the plan's C0–C5 instruments, so it is not counted as a plan
"FAIL" (plan §4); the owner decides whether it should block.

## Release notes read (window baseline → candidate)

- Start 1.169.5 / start-plugin-core 1.171.56: the Rsbuild plugin keeps the app's `server.compress` instead of forcing
  it off (default stays off). It affects the Rsbuild dev/preview server only; production serves through srvx
  (`server.mjs`), and C2/C3 show no compression on the page route. `rsbuild.config.ts` sets no `server.compress`.
- start-plugin-core 1.172.0: custom `serverFns.generateFunctionId` ids that never worked over HTTP now fail the build.
  The app defines no server functions (`grep -rn createServerFn frontend/src` → 0). The same release (TanStack/router
  PR #8876, `results/r2/relnote-pr8876.txt`) changes the **Rsbuild production route manifest**: a route's client
  manifest entry lists only its own chunks and CSS (before, when two routes shared a module in a chunk of its own,
  each route's entry also listed the other route's chunk as a preload and its stylesheet). This is in the Rsbuild
  integration path. C1–C4 cover only the routes they name and assert no CSS; the ordered `make check` (e2e 168
  passed on the candidate image, §Gate) exercises the rest of the product routes and their styles through the browser.
- router-plugin 1.169.0 / router-generator 1.168.0: Node `^20.19 || >=22.12` (Node 24 is fine), Vite ≥6 peer (no
  Vite), `enableRouteGeneration` removed (not used), pathless-layout typing (no route groups here).
- Rsbuild 2.2.13–2.3.3: Rspack 2.3, perf work, config-merge fixes, dev/preview Host validation; nothing found in the
  SSR output path. This short list is not a claim that the whole build path is unchanged.

## Gate (ordered `make check`, 2026-11-20)

Run at `c2074e6a` (pins `3d8c6223`) on Kiri, 20:56–21:28 UTC, with Kiri `DOCKER_HOST` and
`COMPOSE_FILE=docker-compose.yml:<kiri>/local-runtime.override.json` (the `smoke` rate-limit recipe, `_gotchas.md`
2026-11-17). `GATE_EXIT=0`. Every step in order — `format-check lint typecheck test integration check-generated build
e2e-reset up drain-queue seed-identities seed smoke e2e`:

- frontend unit tests 1482/1482 passed, 0 skipped; Go `ok` lines 130, `FAIL` 0;
- the production frontend image built with the candidate (`ready built in … (client)` / `(ssr)` in the log);
- e2e **168 passed** (24.0 min), 0 failed, 0 flaky, 0 did-not-run.
- Image identity: the frontend container ran `sha256:182270a6…` after the gate, equal to the freshly built
  `courses-frontend` image ID (before: `ed6a31ec…`), so e2e did not run on the old image. VM free space 13 → 13 GB; the
  build cache was not pruned.

The gate log stays in the session scratchpad; the api/nginx stream is `frontend/test-logs/api-nginx-20261120T210405Z.log`
(not committed, as in every phase).

## Environment, disk and cleanup

- Disk (plan R2-N2, review F7). Readings are `df -BG` samples (rounded GiB) of the VM root, not a continuous minimum.
  Every task-owned build, with its record:

  | Build | before → after | record |
  |---|---|---|
  | r1: 5 arm64 images, 3 buildx amd64 attempts, manual amd64 image commits | one pair around the whole run (27 → 16) | **NOT_TESTED per build** (deviation from plan R2-N2; no r1 result is used for a verdict) |
  | r2: `m1-front:base`, `m1-front:cand`, `m1-probe:base`, `m1-probe:cand`, `m1-probe:control`, `m1-front:cand-amd64-buildx` | 16 → 16 each | `results/r2/run.log` |
  | first temporary builder (QEMU): `m1-deps:amd64-real`, `m1-front:amd64-real` (failed) | 16 → 14, 13 → 13 | `supporting-records.txt` §c5-real |
  | QEMU builder, baseline `production` (failed) | 15 → 15 | `c5-base-qemu.console` |
  | r2: local `m1-buildkit-norqemu:local` image; `m1-rosetta` sanity build (cache only) | none taken | superseded by the r3 row below |
  | r3 rebuild: `m1-buildkit-norqemu:local` (id `05d172dd…`, 0 qemu files); `m1-rosetta` (v0.35.1, `linux/amd64*, linux/arm64*`) sanity build `--no-cache`: `x86_64`, `v24.22.3 x64`, ELF `3e 00` | 13 → 13, 13 → 13 | `results/r3/c5-prelim.log`, `probe/m1-c5-prelim-disk.sh`, `probe/c5-sanity.Dockerfile` |
  | r3 manual amd64 rebuilds ×6 | 13 → 13 each | `results/r3/manual-repro.log` |
  | Rosetta: `m1-deps`, `m1-front`, `m1-probe` `:amd64-rosetta` | 15 → 14, 13 → 12, 11 → 11 | `results/r2/c5-rosetta.log` |

  No recorded pre-build sample was below the 10 GB stop line (lowest 11, before `m1-probe:amd64-rosetta`). The r1 row
  stays unrecorded per build (a disclosed deviation; superseded, used for no verdict); the two r2 preliminary C5 builds
  were rebuilt in r3 with samples.
- r3 cleanup (2026-11-20 17:17–17:20 UTC): `m1-rosetta` builder and `m1-buildkit-norqemu:local` removed
  (`results/r3/c5-prelim.log`), the re-pulled amd64 `node` digest removed; inventory in `results/r3/cleanup-inventory.txt`:
  0 `m1-*` tags, builders, buildx volumes and containers, digest absent, frontend `ed6a31ec…` healthy, `/courses` 200,
  13 GB free.
- VM free space: 27 GB before the first r1 build, 16 GB at r2 start, 13 GB after cleanup. The `kiri-build` cache grew
  from 14.5 GB to about 21 GB and was not pruned (no permission).
- Restored: the frontend runs `sha256:ed6a31ec…` (`courses-frontend`) again, healthy, `/courses` 200. `m1-delay` and the
  r1 temporary `m1-nginx-buffered` container are removed; the shared `nginx.conf` was never edited (a copy was used).
- Task-owned images, builders and volumes removed; inventory after cleanup in `results/r2/cleanup-inventory.txt`:
  0 `m1-*` tags, the amd64 `node` digest absent, 0 `m1-*` builders, volumes and containers; the local
  `node:24-bookworm-slim` tag is unchanged (arm64 `8400c2d5…`). Untagged cutover images were not touched (no image
  removal outside the task's own tags was run).
- The e2e `setup` project was run once (2026-11-20 11:31 local) to refresh the expired storage states; it writes
  `frontend/playwright/.auth/*.json` and the media fixtures, as in every gate.
- The scratch worktrees (`m1-base`, `m1-cand`, `m1-amd`, `m1-base0`, detached at `12657c14`) live in the session
  scratchpad and are removed after review.

## Limitations (named)

- C1 SPA evidence is the admin shell (`ssr: false`), not the anonymous SSR catalog: catalog cards are document
  navigation on both trees (D1, owner-accepted). Making them client-side links (React Aria `RouterProvider`) is a
  separate product change, not part of M1.
- C2/C4 run on a diagnostic route; product SSR routes await their data before render (plan §3), so "progressive SSR"
  is proven for the mechanism, not for a product page.
- C5 ran on a task-owned builder (D3), not on `kiri-build`, with BuildKit v0.35.1 instead of v0.34.2, pulling the base
  from Docker Hub rather than Kiri's mirror. On Kiri's default builder the amd64 build selected an arm64 base
  (observed; the mirror as cause is inferred, not isolated) — a Kiri runtime fact, not a candidate one. No native
  amd64 host (CI) has run it.
- One manual amd64 `npm run build` exited 1 in r2 with no log kept (probably the client compilation, inferred from the
  missing `(client)` line). It did not reproduce in 16 later `npm run build` runs (r3 6/6 with build logs kept, 8/8
  repeats, 2/2 Rosetta Dockerfile `production` builds). Cause unknown; a one-in-17 flake under emulated amd64 on Kiri
  cannot be excluded.
- The linux `npm ls --all` `invalid` (C0) is a lock-shape consequence of two Rspack copies, measured but not removed;
  whether a metadata/install-shape change could remove it was not explored.
- Emulated amd64 timings are behaviour evidence only, not performance.
