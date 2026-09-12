# Handoff — `courses`, Phase 5 close (SUPERSEDED 2026-09-08)

> **Do not act on this file as current state.** It was written at the close of Phase 5 and
> committed a day later without re-reading the repository, by which time **Phase 6 was well
> under way**: `docs/playback/` holds its spec, plan (edition 6) and a closed GATE 2, and
> `work/tasks.md` carries its task queue. Two of its instructions are already wrong — Phase 6
> is not unplanned, and **backlog B-2 has been taken** (`frontend/rstest.config.ts` now runs
> adapter tests under a `node` project beside the `happy-dom` one).
>
> Current state lives in `work/tasks.md`, `docs/playback/`, and `work/backlog.md`. What is
> still worth reading below: the working model with Codex, the traps, and the methodological
> lesson — none of which is phase-specific.

---

# Handoff — `courses`, Phase 5 closed, Phase 6 unplanned

Written 2026-09-07. Everything durable is in the repository; this file carries only what is
not, plus pointers. `main` = `cd2f20b`, pushed, tree clean, task queue empty.

## State

- **Phases 1–5 complete.** `make check` exit 0 from a clean checkout: 330 frontend tests,
  38 Go packages, e2e 35/35 (run twice, with an identity reseed and `docker compose restart
  nginx` between).
- A person can put a video on a lesson. `/media/*` still answers **403 for everyone** — that
  is deliberate and must stay until Phase 6 builds signed access.
- Four executables now: `api`, `media-worker`, `upload`, `migrate`.

## Read first, in this order

| Read | For |
|---|---|
| `AGENTS.md` | binding rules (current-delivery line is accurate as of this handoff) |
| `.claude/standards.md` | manifest: gate order, placement, reuse index |
| `work/PRODUCT-BRIEF.md` §21 + §24 | the roadmap and every standing decision (§24.1–§24.15) |
| `docs/media/report.md` | what Phase 5 shipped and what it taught |
| `docs/media/plan.md` §17–§25 | the nine dated amendments, and *why* each was needed |
| `work/backlog.md` | B-1 observability, **B-2 (take before Phase 6)** |
| `work/codex-working-model.md` | how to brief Codex — read before writing any brief |

## Do these before Phase 6 code

1. **B-2 first.** Adapter tests run under `happy-dom` although none renders anything, so MSW
   drops custom headers in both directions and the tus client's wire proof lives only in e2e.
   Phase 6's player adapter will also speak in headers (range requests, cache-control, signed
   URLs) — fixing it after that means retrofitting two adapters. The hypothesis and the cheap
   way to test it are in `work/backlog.md` B-2.
2. **Refresh the briefing** (`cto-research`) — `work/CTO-RESEARCH.md` is based on `0c46f10`
   and predates all of Phase 5's code.
3. **Two decisions belong to Phase 6, both already framed:**
   - Multi-language audio delivery (brief §24.15): extra tracks in the source
     (`-map 0:a:0 -map 0:a:1`) versus separate audio uploads attached to a ready video. The
     shared HLS audio group is already in place, so a language costs ~1.4% storage, not 100%.
   - AV1 stays deferred behind the three conditions in §24.14; fMP4 keeps it remux-free.

## The working model (not written anywhere else until now — now it is)

`work/codex-working-model.md` is the durable copy. The essentials:

- **Codex writes code; this session plans, briefs, verifies, commits.** Codex's sandbox has
  no Docker, so **every gate is re-run here before any commit, without exception.**
- Launch detached: `nohup codex exec -s workspace-write --skip-git-repo-check "Read
  work/brief-<task>.md and execute it exactly..." < /dev/null > <scratchpad>/codex-<task>.log
  2>&1 & disown`. The harness kills its own foreground commands; detached survives.
  - **`< /dev/null` is load-bearing** — measured 2026-09-12, a T12 launch stood two minutes
    doing nothing. `nohup` does not close stdin, and detached it is neither a TTY nor a closed
    pipe, so `codex` blocks on `Reading additional input from stdin...` even when the whole
    prompt came in as an argument. Invisible in a TTY, silent in the background: no log past
    that line, no working file, `git status` byte-identical to before.
  - **Watch the working file, not the pid.** `pgrep` proves a process exists, which is exactly
    what the hung launch also proved. The observable that separates "running" from "working" is
    `work/report-<task>.md` growing — the brief requires an append before the first tool call,
    so a launch that has not written that line within a minute is stuck, not thinking.
- Resume for an answer or a correction: `codex exec -s workspace-write --skip-git-repo-check
  resume <SESSION_ID> - < prompt.md`.
  - **Not `--last`.** Its help says "the most recent **recorded** session" — recorded, not
    running, and `~/.codex/session_index.jsonl` carries no `cwd`, so the pick is global by
    recency across every project on the host. With another repository's Codex recorded more
    recently, a resume lands your correction in an agent writing a different repository, and
    it looks like a successful resume. Measured 2026-09-12: two T12 resumes went to the right
    session by luck, not by design.
  - The id is the first `session id:` line of the launch log. Pass the prompt as `-` from a
    file: stdin ends by itself, so this also avoids the `< /dev/null` hang above.
  - Watch it by PID written at launch (`kill -0`), not by `pgrep -f`. With the prompt on
    stdin the brief name is not in the process arguments at all, so a name filter reports
    "finished" immediately — and a filter on the session id matches the watcher too.
  - **Also the recovery path when the run dies mid-step.** Measured 2026-09-12: a T12 run was
    killed with "stopped because the system is running low on memory" — this host carries other
    projects' stacks, and it sat at 31G used / 76M free with 14G in the compressor while two
    agents worked. Resume, never relaunch: say the work was not stopped by a decision, that the
    edits on disk are kept, and quote the run's own last `ДАЛІ:` back to it so it continues
    instead of restarting.
  - **Tell the two failures apart before choosing a recipe.** OOM: the working file is
    non-empty and the tree carries edits, cut off mid-step. The stdin hang above: zero entries
    and `git status` byte-identical to before.
  - **One Codex at a time on this host.** Two concurrent runs is what the memory ceiling
    actually bit on; the gain was never worth it, since neither can run a Docker gate anyway.
- **Two channels, and the second one is where the value was:** stop on a missing contract;
  **open a `## Discussion` for an inaccuracy that would cause a bug, even when the
  instruction could be followed as written.** Fourteen items were raised in Phase 5 and all
  fourteen were accepted. Three were defects in an amendment this session had written.
- Briefs forbid: touching git, editing `work/tasks.md` (status is the reviewer's call, after
  the gates ran), writing the phase report, writing inside `generated/`.

## Traps that cost real time in Phase 5

- **`docker cp` cannot read a tmpfs mount.** The worker's `/tmp` is tmpfs; fixtures must be
  streamed out with `docker compose exec -T <svc> cat <path>`.
- **ffprobe repeats a stream entry per playlist segment.** Assert sets, not line counts.
- **MSW + happy-dom drop custom headers in both directions** (B-2).
- **pgx sends a parameterised `Exec` over the extended protocol** — a multi-command SQL
  fixture fails with `42601`. One command per `Exec`.
- **Do not hardcode `generation` in fixtures**; derive it and cast the parameter explicitly
  or you get `42P08`.
- **`make e2e` twice in a row hits 429.** `docker compose restart nginx` resets the limiter;
  reseed identities too, because `delete-me` is terminal.
- **The media journey must stay in its own Playwright project** — it restarts the worker and
  saturates two CPUs, which starved a 4b dialog to a 5s timeout.
- **The worker ships ffmpeg 5.1.9, the host has 7.1.1.** Any claim about muxer behaviour must
  name the binary; `B33` prints the worker's version into every smoke run.
- **rstest is slow (~10 min for the full suite)** and exceeds foreground timeouts — run it
  detached and watch the process.

## The methodological lesson worth keeping

Six GATE 2 rounds — including reviewers who ran real ffmpeg and a live PostgreSQL — did not
catch that the published upload protocol **could not accept a first byte**, nor that a port
table contradicted the decision it implemented, nor that a CHECK constraint enforced nothing.
The gates read mechanisms; they did not read the seam between a mechanism and the signature
serving it, and they did not ask whether the system could start from empty. All three were
found by the implementer, at the first moment something had to actually run.

When a decision says "inside the same transaction", the port table owes a **signature**, not
a name. And a section written to repair one defect deserves the same adversarial reading as
the plan it repairs — §2.1 produced three.

## What Phase 6 added to that lesson (2026-09-12)

Nine times in one day a check turned out to prove less than it claimed, and not once did
reading the diff find it. Three questions did, and they are different questions:

- **"What does this actually prove?"** — six cases. The discriminant compared against a
  baseline taken *before* a control seek that deliberately moved it, so it could never be
  false. `axeAudit` mapped `id` and `target` but not `messageKey`, so pinning one node would
  have swallowed a real finding on the same node. `pgrep` proved a process existed, which is
  exactly what a hung launch also proved. An SSR test guarded a route where SSR was already
  off. B25 measured contrast over a `<video>` that never loaded — and the old brief's
  prediction that controls over video would raise `incomplete` was never reproducible in that
  fixture at all.
- **"What happens if the product does better than I demand?"** — two. Asserting that another
  720p fragment must be requested before the downswitch would have gone red on a player that
  dropped the level sooner. The counter varied 15 and 5 across two runs while the thing being
  claimed grew by exactly 1 in both. An over-constraint is invisible on the run where it did
  not bite, which is what makes it worse than a flaky check.
- **"Under what conditions was this measured?"** — one, and it was the one neither brief had.
  Ownership and heartbeat are proven at 2/6 because the `playback-profile` project recreates
  api there; the product ships 30/180. The invariant `h*3 <= ttl` holds in both, so the
  timings were compressed, not the logic — but "proven" without the cadence says more than was
  measured.

The same day produced the mirror image, twice, in one file. `work/tasks.md` carried `planned`
on a task already in history — a claim *weaker* than the work — and nearly carried `done` on
T13 while the phase gate it ran had failed, a claim *stronger*. The next session picks work
from that table, so both mislead, and both are caught by asking what stands behind the words.

And the count in a plan header is a claim like any other: §8 said 94 rows through four GATE 2
rounds. It has 102. Nobody recounted it until a report had to be written against it.

## A fourth question, and it catches something else (2026-09-12)

The three above catch a **weak check** — something that runs and proves less than it claims.
The day ended with a different failure that none of them would have found, because nothing was
checked at all.

- **"Is this a source or a copy, and when was it last reconciled with what it describes?"**

Three instances, and they are the same shape:

- **A count in a plan header.** §8 said 94 rows through four GATE 2 rounds. It has 102. Eight
  live rows sat outside the ledger, and carrying 94 into the phase report would have made the
  phase look better verified than it was.
- **A column in an acceptance table.** The `level` column reads as a fact about where a row's
  proof lives. For B20, B21, B22, B28 and R23 it was an assumption: they need an allowed media
  segment with real bytes, and `backend/sql/seed.sql` inserts courses only — no lesson, no
  video, no media. The level was never checked against the fixtures that would have to satisfy
  it. Anyone taking those rows plans a test and hits a decision instead.
- **The read-first chain itself.** `AGENTS.md` said T11–T13 remained after all thirteen were
  committed; it sends the reader first to a briefing built at 126 commits (the repository has
  266) whose live P0 is the delivered Phase 6 specification, and to a ranking 208 commits
  behind that recommends the delivered Phase 4. A session following the instructions exactly
  would have started work that was already in the history.

None of the three was a reasoning error, and none would survive five minutes of checking. They
survived because each was treated as the thing you check *against*. A source is only a source
until the code moves past it, and nothing announces when that happened.

One clarification, because it is tempting to file it here: the discriminant compared against
the wrong baseline belongs to **"what does this prove"**, not to this question. That baseline
was read correctly and meant what it said — the fault was that the comparison could not fail.
Widening this category to hold it would blunt it. The test for this question is narrower:
**nobody looked**.

The practical form is cheap. Anything that carries a count, a date, a commit or a "current"
gets recomputed the first time you rely on it, not the first time it is questioned — and a
document that claims to rank or summarise work says which commit it was built at, so the
distance is visible without reading it.

## Three things the four questions did not catch (2026-09-12, later)

### The third question does not protect the person who wrote it

«Under what conditions was this measured?» was written into this file at half past twelve.
An hour later the session that wrote it shipped `R30` measured on a full checkout, into a gate
that runs Go tests inside a container mounting **only `backend/`** (`Makefile:3`). From there
`../../../infrastructure/nginx/nginx.conf` does not exist, so the nginx half of the check
**fails always**, whatever the config says — the mirror of an always-true check, and its
mutation proof means nothing in the gate, where it is red without any mutation.

The repository already knew. `infrastructure/tests/integration.sh:11-13` says it outright:
"The Go tests cannot assert this — their container mounts only `backend/`, so the repository
root is invisible from inside them," which is why G37 checks root files on the host, before
`docker run`. The rule was written, delivered and unread at the moment it applied.

**The reviewer missed it too, and the shape of that miss is the lesson.** Asked whether the
moved test needed its mutations re-measured, the review verified four conditions — the file
was restored byte-for-byte, the package was unchanged, paths resolve from the package
directory rather than the file, the assertion does not depend on location — and every one of
them was true. None of them asked in **what runtime the gate runs the test**. A correct answer
inside the wrong frame reads exactly like a correct answer.

So the third question has a second half that is easy to skip: ask it of the **check**, not
only of the claim. Where does this run in the gate, and has it ever run there?

### A prohibition in a brief loses to an instruction in a skill

Both briefs forbade Codex to touch git, in §3, among the other "do not"s. In T11 Codex
resolved the conflict in favour of the brief; in T14 the `implement` skill won and it reached
`git commit`. What stopped it was not the rule but the sandbox —
`fatal: Unable to create '.git/index.lock': Operation not permitted`. It then reported
honestly and did not work around it, but the guard that held was **accidental**: another
machine, other permissions, and the finding would have arrived as a commit in the history.

Write the prohibition as the first line of §4 with its **reason** — the commit is made by the
reviewing session after a Docker gate you do not have. A reason is a consequence; a skill
does not override a consequence the way it overrides a bare rule.

### A filter that names a task also names its watcher

`pkill -f 'brief-t14'` killed the monitor watching that task, because the monitor's own script
carried the string. Anything matching on a task's name matches the things observing it.


## Four observers that trusted a shape instead of measuring it (2026-09-12)

Every one of these was a correct recipe that quietly stopped being correct when the conditions
around it changed. None failed loudly; each reported something plausible.

| Observer | Believed | Actually |
|---|---|---|
| `pgrep -f 'brief-<task>'` | the process exists ⇒ the agent is working | a launch hung on stdin proves the process exists just as well. And once the prompt moves to stdin, the brief name leaves the argv entirely, so the same filter reports "finished" at the moment of start |
| `codex exec resume --last` | the most recent session is mine | "most recent **recorded**" is global: `~/.codex/session_index.jsonl` carries no `cwd`, so another project's Codex, already finished, can receive the correction — and it reads as a successful resume |
| "the agent appends at the end" | new lines arrive at the tail | Codex appended after *its own* previous last line, so reviewer blocks stayed at the tail and a tail-watcher saw them instead of the new records |
| `tail -F` on a run log | the file I opened is the log | a new run truncates or replaces it; a stale `GATE_EXIT=2` from the previous run arrived a minute into the next one |

The common shape: the observer asserts a **property of the thing watched** — a name in argv, a
session's recency, an append position, a file identity — and that property is incidental, not
guaranteed. It holds until someone changes how the work is launched, and then the observer
keeps reporting with the same confidence.

The cheap fixes all replace an inferred property with a recorded one: a PID written at launch
and `kill -0`; an explicit session id taken from the launch log; a diff of snapshots rather
than a tail; a log path unique per run. None is cleverer than what it replaces — each just
measures instead of assuming.

This is the same failure as the fourth question one level out. There the unchecked record was a
count, a column, a pointer; here it is a recipe. A recipe is a claim about how the world works,
and it goes stale the same silent way.

