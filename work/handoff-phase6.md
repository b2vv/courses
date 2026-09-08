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
  work/brief-<task>.md and execute it exactly..." > <scratchpad>/codex-<task>.log 2>&1 &
  disown`, then a persistent `Monitor` on `pgrep -f "brief-<task>"`. The harness kills its
  own foreground commands; detached survives.
- Resume for an answer or a correction: `codex exec ... resume --last "<answer>"`.
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
