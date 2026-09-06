# M13 ship report — upload edge and topology

## 1. What was done

- Added the `upload` Compose service using the existing upload image stage, PostgreSQL and
  Redis runtime inputs, `HTTP_ADDR=:8082`, writable `media-data`, a `/readyz` health check,
  `read_only: true`, `/tmp` tmpfs, `mem_limit: 512m`, and `pids_limit: 64`
  (`docker-compose.yml:99-122`).
- Made Nginx wait for the upload service to become healthy (`docker-compose.yml:142-146`).
- Added the dedicated `upload` request zone and `/uploads/` proxy. The location streams
  request bodies with buffering disabled, caps a request at 16 MiB, uses 120-second proxy
  timeouts, and clears inbound trace headers (`infrastructure/nginx/nginx.conf:30,49-63`).
- Kept the global 1 MiB body limit and both `/media` refusal locations unchanged
  (`infrastructure/nginx/nginx.conf:20,45-46`).
- Extended smoke coverage with upload readiness, an unknown-ticket `HEAD` through Nginx,
  the 32 MiB edge refusal, and a before/after upload-service counter comparison proving the
  oversized request did not reach Go (`infrastructure/tests/smoke.sh:3-36`).
- Added the R13 path matrix for `/media`, `/media/`, and a nested playlist path, anonymously
  and with a real seeded admin session cookie (`infrastructure/tests/smoke.sh:37-50`).
- Documented the upload/worker/shared-volume topology and retained the explicit no-playback
  boundary (`README.md:5,29-32,65`).

## 2. Acceptance rows

Docker is denied in this environment, so the assertions exist but are not represented as
passes.

| Row | Planned observable result | Actual result | Verdict |
|---|---|---|---|
| R12 | A 32 MiB PATCH is rejected by Nginx with 413 and the upload route counter does not move | Status and counter-delta assertions implemented; not executed | not verified |
| R13 | All three `/media` paths return 403 anonymously and with an admin cookie, before and after a successful transcode | Six pre-transcode variants implemented; not executed. The post-transcode half is not feasible in M13; see Discussion | partial, not verified |

Unverified acceptance rows: **2**.

## 3. Non-obvious findings

- R12 must send a body with a known length. The smoke request supplies an exact 32 MiB
  `Content-Length`; a chunked transfer would intentionally bypass Nginx's size pre-check and
  rely on the upload service's separate 8 MiB reader cap (`docs/media/plan.md:1452-1456`).
- R12 compares the sum of all status-labelled `PATCH /uploads/{videoId}` counters before and
  after the request. This proves the edge rejected the body without depending on the route's
  prior traffic or an assumed status label.
- The admin-cookie half of R13 obtains a real session from the seeded admin fixture rather
  than sending a syntactically cookie-shaped but unauthenticated value.

## Discussion

### R13 assigns post-transcode evidence to a task that cannot produce it

Evidence: R13 requires the six path/auth combinations to answer 403 both before and after a
successful transcode, assigns the row to M13, and names `make smoke` as its evidence
(`docs/media/plan.md:1884`). M13 itself is a Compose/smoke task (`docs/media/plan.md:1956`),
while the first real successful-transcode journey is owned by M16 and E2E
(`docs/media/plan.md:1819,1959`). Section 5.d also describes only a media-path smoke assertion,
not transcode setup (`docs/media/plan.md:1465-1474`).

Runtime consequence: a future existence-dependent Nginx rule could refuse missing media yet
expose a file after publication, while M13's pre-transcode smoke still passes.

Proposed correction: keep M13's six pre-transcode variants, repeat the same six requests in
M16 after its asset reaches `ready`, and assign R13 evidence to M13 plus M16 (or narrow M13's
owned half explicitly). Meanwhile M13 implements the complete pre-transcode matrix and does
not pretend the post-transcode half ran.

### The repository scope instruction is stale while M13 must document the topology

Evidence: `AGENTS.md:5-6` still calls Phase 4 current delivery and forbids claims of working
uploads. The task explicitly requires the README's new upload topology
(`work/brief-m13.md:24-29`), the current task tracker places Phase 5 at M13
(`work/tasks.md:120-134`), and the upload executable/image already exist. The plan assigns the
`AGENTS.md` current-delivery update to M17 (`docs/media/plan.md:1960`).

Runtime consequence: an agent treating the stale scope line as current could remove or
duplicate the upload topology, or incorrectly keep documenting it as entirely future work.

Proposed correction: perform the planned scope-line update in M17 after the full ordered
gate. Meanwhile the README states only the concrete routing and mount topology, explicitly
keeps the browser upload UI future, and makes no playback claim. `AGENTS.md` was not edited.

## 4. Verification

- `make -n up smoke build` completed with exit 0. It confirms that `make up` builds/waits and
  then reloads Nginx, and that smoke/build resolve to the expected denied commands; it is not
  runtime evidence.
- `sh -n infrastructure/tests/smoke.sh` completed with exit 0.
- The embedded Node module extracted from `smoke.sh` passed `node --check --input-type=module`.
- Read-through verified the Compose service against the plan's exact limits/dependencies,
  the Nginx block against §5.c, the unchanged `/media` refusals/global body limit, and every
  smoke assertion against §5.d plus R12/R13.

## 5. Environment denials and limitations

- **Docker is denied.** No Docker command was attempted or worked around. Compose config,
  service startup, health checks, Nginx config loading, build, smoke, and E2E were not run.
- The bind-mounted `nginx.conf` edit is invisible to an already-running Nginx until reload.
  `make up` contains that reload (`Makefile:8-12`), but it was only inspected by `make -n`.
- The ordered checks not run are: `make format-check`, `make lint`, `make typecheck`,
  `make test`, `make integration`, `make check-generated`, `make build`, `make up`,
  `make seed-identities`, `make seed`, `make smoke`, and `make e2e`. Consequently
  `make check` was not run. None is represented as passed.
- `make check-generated` is additionally denied because it invokes Git and writes generated
  output. No substitute drift check was attempted.
- No Git command was run. `work/tasks.md` and `docs/media/report.md` were not edited. No file
  inside a `generated/` directory was written.

## 6. Code review

### Standards

No hard implementation violation remains. The review noted that the README topology and
stale `AGENTS.md` scope wording conflict; that is recorded in Discussion. Repeated proxy
headers are a possible Fowler duplication smell, but they follow the plan's exact explicit
Nginx block and are too small to justify a new include under ponytail-full.

### Specification

Compose, Nginx, README, the R12 edge/counter proof, and the pre-transcode R13 matrix match the
brief. The sole partial item is R13's post-successful-transcode proof; the contradictory
ownership/evidence is recorded in Discussion.

Summary: Standards — 1 documentation conflict, no implementation defect; Specification —
1 partial evidence item (R13 post-transcode), no other finding.

## Post-deploy

Pending. M13 was not built, started, or deployed in this environment.

## Reviewer verification (measured in the reviewing session)

Both Discussion items accepted, recorded as plan **§23**.

R13's evidence is **split**: M13 owns the six pre-transcode variants under `make smoke`, and
M16 repeats the same six after its asset reaches `ready` under `make e2e`. The half M13 could
not run is the one that matters — an existence-dependent edge rule could refuse missing media
and still expose a *published* file, and a pre-transcode smoke would pass throughout. Same
shape as the B19 smoke-runs-before-e2e correction from GATE 2.

`AGENTS.md` was **stale for the whole phase** and is corrected now rather than at M17: it
named Phase 4 as current delivery and forbade claiming working uploads while this task
documents upload topology in the README. Parking that until the last task would have left the
repository's binding document wrong for the entire phase.

Gates: `make build`, `make up`, seeds, and **`make smoke` exit 0** —
"PASS readiness, upload proxy/body cap, private media/metrics, worker handler/metrics and
API, login and SSR rate limits", with the worker's `ffmpeg 5.1.9` printed. format-check,
lint, typecheck and test pass. Seven healthy containers including `courses-upload-1`; the
edge keeps `/media` and `/media/` at 403 while `/uploads/` proxies to the upload service with
`proxy_request_buffering off` and its own rate zone.

Provenance: implemented with Codex (gpt-5.6-sol); reviewed and verified here.
