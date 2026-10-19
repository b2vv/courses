# Research — testcontainers-go Ryuk reaper intermittent failure

Requested by Courses 1 (owner asked for a colleague). Researcher: Courses 2. Date: 2026-10-03.

**Question:** why does testcontainers-go v0.44.0 fail intermittently with
`reaper: from container "<id>": wait for reaper <id>: unexpected container status "created"`,
always at the first `postgres.Run` in a parallel package, passing on a plain rerun?

**Hypothesis to confirm/refute (Courses 1):** packages run as separate processes under one `go test`,
share a session id, and race to reuse one reaper (Ryuk) container — one sees it still `created`.

ДАЛІ: шукаю в GitHub testcontainers-go issues/PRs/changelog по тексту помилки й reaper-reuse-гонці;
чи фіксить новіша версія; офіційні рекомендації (RYUK_DISABLED, reuse, connection timeout, -p 1).

---

## Findings (append-only)

### 1. The hypothesis is essentially CONFIRMED by upstream issues (v0.44.0)

**Issue #3918 — "an abandoned reaper create leaves `reaper_<session>` in Created and blocks the
rest of the session"** (opened 2026-09-30, reporter on **v0.44.0**).
<https://github.com/testcontainers/testcontainers-go/issues/3918>
- Mechanism (quoted): `reaperSpawner.newReaper` starts `ContainerCreate` **with the caller's
  context**. Under heavy daemon load the create request **outlives the caller's deadline**; the
  client gets a context error *without the container id*, yet the daemon finishes creating it a
  moment later, leaving a container in **`created`** state.
- "The orphan is never started. For the rest of the session: its fixed name blocks creating a new
  reaper; `lookupContainer` (`All: true`) keeps returning it; `fromContainer` waits for a `Started`
  log line that never comes."
- "Every later package binary therefore fails to connect a reaper" — i.e. it **cascades across the
  parallel package binaries that share one session id** (exactly Courses 1's hypothesis). One machine
  accumulated 3 stuck reapers + 172 leaked containers over three days.
- FACT: not fixed in a release yet (links #3867/#3868/#245 as proposals).

**Issue #3867 — "reaper reuse finds an exited ryuk container and times out instead of creating a new
one"** (opened 2026-09-01, **v0.44.0**). <https://github.com/testcontainers/testcontainers-go/issues/3867>
- Mechanism: `reaperSpawner.lookupContainer` lists with `All: true` and **no running-state filter**,
  so an **exited** ryuk container is returned as reusable; `fromContainer` then waits on
  `wait.ForLog("Started")` + `wait.ForExposedPort()` against a not-running container until the default
  **60s** startup timeout expires.
- Why ryuk is gone: Ryuk exits **10s** after its last client disconnects
  (`RYUK_RECONNECTION_TIMEOUT` default); multi-package runs create idle windows where no package holds
  the connection.
- Reporter's workaround (quoted): "Our workaround is `TESTCONTAINERS_RYUK_DISABLED=true` in CI (the VM
  is ephemeral), which sidesteps the reaper entirely."

### 2. The fix: PR #3868 (OPEN, not released as of 2026-10-03)

<https://github.com/testcontainers/testcontainers-go/pull/3868> — "recreate reaper when the existing
container is not running."
- Makes `lookupContainer` check state: **reuse running**, **retry created/restarting**, **force-remove
  any other state (exited/dead/paused/removing)** and build a fresh reaper instead of waiting 60s.
- `fromContainer` also detects termination before/during the readiness waits and retries.
- STATUS (FACT, 2026-10-03): **Open, not merged**, created 2026-09-01, last force-push 2026-09-29. So
  **no released version contains the fix yet** — upgrading past v0.44.0 will NOT fix it today (to be
  re-verified below).

### Related, for context
- #2172 "Sporadically fails to create reaper in GitHub Actions"; #2477 "creating reaper failed";
  #2945 "no port to wait for"; #2757 MacOS-after-upgrade — a long tail of reaper-startup flakes.
- testcontainers-**python** #746: the same class of race (parallel starts → ryuk Conflict) in the
  sibling library — evidence this is a cross-language design issue, not Go-specific.

### 3. Version & fix status (FACT, checked 2026-10-03)

- **v0.44.0 is the latest release** (published **2026-08-07**).
  <https://github.com/testcontainers/testcontainers-go/releases/tag/v0.44.0> ·
  <https://github.com/testcontainers/testcontainers-go/releases>
- Its own notes already claim a reaper fix ("proper waiting for Ryuk startup when reusing reaper") —
  but #3867 (2026-09-01) and #3918 (2026-09-30) were opened **after** it, showing that fix is
  **incomplete**: it waits for the reaper, but can wait on a non-running / orphaned-"created"
  container.
- **No released version fixes this today.** PR #3868 is **open**. ⇒ **Upgrading will not help right
  now** — re-check #3868's merge + the next tag before planning an upgrade-based fix.

### 4. Why a plain rerun passes (mechanism, matches the symptom)

The orphan reaper has a fixed name `reaper_<sessionId>`, where the session id is derived from the
parent `go test` process. A rerun is a **new parent process ⇒ new session id ⇒ new reaper name**, so
there is no stuck-name conflict and it starts clean. The old orphan is **leaked** (not reaped) — which
is exactly the "172 leaked containers" in #3918. So "rerun passes" is expected and does **not** mean
the run was healthy; leaks accumulate on a non-ephemeral (dev) machine.

GUESS (not pinned to a source line): the exact text Courses 1 sees,
`wait for reaper <id>: unexpected container status "created"`, is the **downstream** manifestation in
a later package binary of the #3918 orphan — a reaper found stuck in `created`. #3918 itself quotes the
*upstream* symptom as `context deadline exceeded` (the create that outran its context). I could not
find the literal string `unexpected container status "created"` tied to a specific file/version in the
issues; the symptom mapping to #3918 is nonetheless unambiguous (reaper in `created`, cascades, rerun
clears). Treat the string→file link as unconfirmed.

### 5. Official Ryuk config reference (FACT)

Source: <https://golang.testcontainers.org/features/configuration/> and
<https://golang.testcontainers.org/features/garbage_collector/>. Env var takes precedence over
`.testcontainers.properties`.

| property | env var | default | purpose |
|---|---|---|---|
| ryuk.disabled | `TESTCONTAINERS_RYUK_DISABLED` | false | disable the reaper entirely |
| ryuk.container.privileged | `TESTCONTAINERS_RYUK_CONTAINER_PRIVILEGED` | false | run Ryuk privileged |
| ryuk.reconnection.timeout | `RYUK_RECONNECTION_TIMEOUT` | **10s** | how long Ryuk waits after the last client disconnects before it exits |
| ryuk.connection.timeout | `RYUK_CONNECTION_TIMEOUT` | **1m** | timeout establishing the Ryuk connection |
| ryuk.verbose | `RYUK_VERBOSE` | false | verbose reaper logging |

Deprecated/renamed: `TESTCONTAINERS_RYUK_{CONNECTION,RECONNECTION}_TIMEOUT` and
`TESTCONTAINERS_RYUK_VERBOSE` → `RYUK_CONNECTION_TIMEOUT`, `RYUK_RECONNECTION_TIMEOUT`, `RYUK_VERBOSE`.
Ryuk reaps by **container labels** set by testcontainers; it also self-cleans when idle.

### 6. Recommendations for `courses` (ranked), with which issue each addresses

Our setup (from Courses 1): one `docker run golang:1.27.1-bookworm go test -count=1 -tags=integration
./...` with the **host Docker socket mounted**, several packages starting Postgres **in parallel** —
precisely the multi-package-one-session shape in #3867/#3918.

1. **`TESTCONTAINERS_RYUK_DISABLED=true` for `make integration`** — strongest, and the #3867
   reporter's own fix. Sidesteps both the exit-during-idle reuse (#3867) and the orphan-"created"
   (#3918) because no reaper is created/looked-up at all. TRADE-OFF (FACT): Ryuk is the *crash* safety
   net; with it off, cleanup relies on testcontainers' normal `Terminate`/`t.Cleanup`, and a crashed
   run leaks containers. The CI `docker run` is ephemeral (its container exit drops its children only
   if they are its children — here they are the **host** daemon's, so they are NOT auto-dropped), so
   pair this with a label-based sweep (`docker rm -f $(docker ps -aq --filter
   label=org.testcontainers=true)`) in the script's cleanup. GUESS: this is the lowest-friction fix
   for a test-mode repo with no prod; confirm the owner is fine trading the auto-reaper for an explicit
   sweep.
2. **`go test -p 1 -tags=integration ./...`** — run package binaries **serially**. Removes the
   parallel idle-window that lets Ryuk exit between packages (#3867) and the parallel-create races
   (python #746 analogue). Keeps the reaper. COST: slower integration (loses cross-package
   parallelism). Addresses #3867; partially #3918 (fewer concurrent creates under load).
3. **Raise `RYUK_RECONNECTION_TIMEOUT`** (e.g. `60s`) — keeps the single shared reaper alive across the
   idle gaps between packages, so a later package does not find an **exited** ryuk (#3867). Does NOT
   help #3918 (the create-timeout orphan). Cheap, keeps parallelism, keeps the reaper.
4. **Start Postgres once per package in a shared `TestMain`** (where not already) — fewer reaper
   handshakes and fewer idle windows. Mitigates, does not cure.
5. **Do not rely on an upgrade yet** — v0.44.0 is latest, #3868 open (see §3).

Combination that keeps the reaper and most parallelism: **#3 + #4**. The simplest robust one for an
ephemeral CI: **#1 + a label sweep**. #2 is the one-line "make it deterministic, accept it's slower".

### 7. Addendum (2026-10-03, later): the exact code path, read from the v0.44.0 source

Source: the local module cache, `github.com/testcontainers/testcontainers-go@v0.44.0`. This replaces
the GUESS in §4 with facts.

- **Ryuk image:** `internal/config/config.go:16` → `ReaperDefaultImage = "testcontainers/ryuk:0.14.0"`.
  The version pair is current; a version incompatibility is ruled out.
- **Where the string comes from:** `wait/wait.go:55`, `checkState` → `default:` branch →
  `fmt.Errorf("unexpected container status %q", state.Status)`: anything that is not running,
  OOMKilled or exited.
- **Which path raises it:** the full text `reaper: from container "<id>": wait for reaper <id>: …`
  is built by `reuseOrCreate` → **`fromContainer`** (`reaper.go:335`, `:344`). That is the **reuse**
  path: this process did **not** create the reaper. It found one by `lookupContainer`.
- **Why a `created` container gets found:** `lookupContainer` (`reaper.go:158`) lists `All: true` by
  session-id label + name, with no state filter. Its health switch accepts `""`, so it returns a
  container that a **sibling package process** has just `ContainerCreate`d inside
  `provider.RunContainer` but not yet `ContainerStart`ed.
- **Why it fails at once, without waiting:** `fromContainer` waits with `ForAll(ForLog("Started"),
  ForExposedPort())`. `wait/log.go:170` calls `checkTarget` on the first poll. A created container
  has no logs, so `lastLen == len(b) == 0` and `checkErr != nil`. It returns `checkErr` on the
  **first poll**.
- **Why it is not retried:** `retryLocked` → `retryError` (`reaper.go:230`) retries only
  NotFound/Conflict-cleanup, name-conflict, ECONN*, timeouts and context errors. A plain `fmt.Errorf`
  falls through to `backoff.Permanent`, so the run fails on the first attempt.

**Conclusion (FACT from code):** the failure is a **transient race** between parallel package
processes that share one session id. Process B looks up during process A's create→start window.
This confirms Courses 1's original hypothesis directly. #3918, the permanent orphan, is a sibling
variant. It would cascade to *every* later package in the session. The observed pattern is "one
package fails, others pass, rerun passes", and that fits the transient window better.

**How to tell the two causes apart after a failure:** run
`docker ps -a --filter name=reaper_ --format '{{.Names}} {{.Status}}'`.
- **Up:** a sibling started the reaper ms later, so this is the transient race.
- **Created**, and it stays Created: this is the #3918 orphan.
- **Exited**, with `.State.Error` set: this is a runtime, socket or privilege problem.

### 8. Follow-up: remedies checked against the v0.44.0 source

- **A state check inside `lookupContainer` is enough.** The lookup already runs in
  `backoff.RetryWithData` with `r.backoff()` (`reaper.go:111`): 100–250 ms steps,
  `MaxElapsedTime: 20s`. Only `errReaperNotFound` is Permanent. A plain "not running yet" error is
  therefore retried for up to 20 s, which is enough for a sibling's create→start. This is what
  **PR #3868** does ("retry created/restarting containers"), so the right upstream fix already exists
  as an open PR. Supporting #3868 with our parallel-packages reproduction beats filing a duplicate.
- **A `status=running` list filter also works, but costs an extra loop.** On a name conflict,
  `newReaper` hits `createContainerFailDueToNameConflictRegex`, which `retryError` lists as
  retryable. That gives lookup → create → 409 → backoff → lookup, a worse path than the state poll.
- **A per-process `TESTCONTAINERS_SESSION_ID` works but is invasive.** The variable is read in
  `internal/config/config.go:133`. It must be set before the first `config.Read()` in every
  integration package, which means a `TestMain` in each, and it produces N reapers.
- **The current gate is correct, but one comment in it is not.** `infrastructure/tests/integration.sh:238-246`
  disables Ryuk. That is right: it is the only option that removes the race and keeps package
  parallelism. CI (`.github/workflows/ci.yml:23` → `make integration`) runs on an ephemeral runner.
  Locally, the host is a long-lived Docker Desktop, so containers leak only on hard kills. The comment
  documents a manual sweep for that. The comment's *mechanism* is slightly off: it says
  "half-created under daemon load" (the #3918 orphan). The observed failure is the transient
  create→start window in which a sibling process finds the reaper (§7). The decision stands; only
  the wording needs a fix.
- **`-p 1` vs Ryuk off.** Both remove the race. To choose between them, measure the extra
  `make integration` wall time with `-p 1`; I have no number for it.
