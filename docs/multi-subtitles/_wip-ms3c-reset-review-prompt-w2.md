# Re-review: package `ms3c-reset-w2` (Makefile `e2e-reset`)

You are an independent reviewer: read-only, no edits, no Docker commands. Your w1 verdict was BLOCK with two Majors
(`docs/multi-subtitles/_wip-ms3c-reset-review-codex.md`). The new diff against HEAD `7b20d2c` is
`docs/multi-subtitles/_wip-ms3c-reset-w2.diff`, and the hashes are in `_wip-ms3c-reset-w2.sha`. Return SHIP or BLOCK with ranked
findings, each with file:line and a failure scenario. Verify specifically that both w1 Majors are closed, and look for new defects.

## Author's changes for the w1 findings

1. **Lookup failures:**
   - one `docker volume ls --format '{{.Name}} <project label> <volume label>'` call runs per lookup, and `|| exit 1` is applied to
     its own assignment;
   - selection is done with awk on that output;
   - there is no `inspect … 2>/dev/null` any more.
   - The resolver failing → exit 1. A volume key missing from the resolved config → exit 1.
2. **Project and volume names:** taken from `docker compose -f docker-compose.yml -f docker-compose.e2e.yml config --format json`,
   which is the same file set `down` uses, parsed with `node`. Node is already required by the gate (`npm run …`). The fields are
   top-level `name` and `volumes.<key>.name`. The `basename` fallback is gone.

## Author's proof (Kiri, probe `COMPOSE_PROJECT_NAME=e2ereset-probe`)

1. Unlabeled + labeled + absent → removed / removed / absent, exit 0.
2. A rerun gives absent ×3, exit 0.
3. A `docker` wrapper on PATH failing only `volume ls` → "e2e-reset: cannot list volumes", exit 2.
4. From a directory named `courses-review` with `COMPOSE_PROJECT_NAME` unset, `config --no-interpolate` resolves project `courses`
   and volume `courses_postgres-data`; with `COMPOSE_PROJECT_NAME=other` it resolves `other_postgres-data`. Without `.env` the
   interpolating `config` fails, so the recipe would exit 1. The `courses_*` volumes stayed untouched (7).
5. The held-volume refusal path is unchanged from w1 (proof: "volume is in use", exit 2).

## Check in particular

- Quoting of the Go templates `{{.Label "…"}}` inside a make variable expanded into `$$( … )`.
- awk field handling when labels are empty (unlabeled volume lines are "name  ").
- `node -e` inside single quotes in a make recipe: any `$` or quote hazards?
- The interaction with `COMPOSE_FILE` set in the environment (Kiri sets it): does `config -f … -f …` ignore it the same way `down`
  does?
- Comment accuracy.
