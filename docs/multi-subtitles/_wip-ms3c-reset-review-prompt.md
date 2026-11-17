# Review request: package `ms3c-reset` (Makefile `e2e-reset`)

You are an independent reviewer: read-only, no file edits, no test runs that touch the shared Docker stand. Return a verdict of
SHIP or BLOCK with findings ranked Blocker / Major / Minor, each with file:line and a failure scenario.

## What changed

The only change is in `Makefile`, the `e2e-reset` target and its comment. The diff is `docs/multi-subtitles/_wip-ms3c-reset.diff`,
taken against HEAD `7b20d2c`. The original target came from commit `8780815` (K13); read that commit's message for the intent.

## Problem being fixed

On the Kiri engine, the Courses volumes `courses_postgres-data`, `courses_redis-data` and `courses_media-data` carry **no labels**;
they were copied there in a cut-over. The old recipe selected volumes only by the compose labels `project` and `volume`. So it
removed nothing, its survival check used the same blind filter, and it printed "removed postgres-data redis-data media-data"
anyway. Three `make check` runs started on accumulated data, which caused catalog/admin e2e failures (148 courses).

## Intended behaviour (owner-approved slice)

- Select each volume by compose label OR by the name compose itself resolves, `<project>_<key>`. Verify that there is no `name:`
  override in `docker-compose.yml`, `docker-compose.e2e.yml` or `/Users/strelia/project-data/courses-kiri/local-runtime.override.json`.
- Remove what was found; fail if anything survived; report per volume "removed <names>" or "absent".
- Absent is not an error (fresh machine or CI).
- The explicit list and the kept prometheus-data stay as K13 set them.

## Author's proof (Kiri engine, probe project `COMPOSE_PROJECT_NAME=e2ereset-probe`)

- One unlabeled volume, one labeled volume, one absent volume → `removed`, `removed`, `absent`, exit 0. A second run gives `absent`
  ×3. The `courses_*` volumes were untouched.
- A volume held by a container → `volume is in use`, `make: *** [e2e-reset] Error 1`.

## Check in particular

- Does the name-based lookup risk deleting a volume that is not this project's, for example with a project name containing
  `_`, an empty project, or `COMPOSE_PROJECT_NAME` set to another project?
- Does `docker volume inspect --format '{{.Name}}'` behave as intended inside make? Is `$$`-escaping correct throughout? `sh`
  semantics apply: function definition, `sort -u`, word splitting of `$$ids`.
- What happens when `docker volume ls` or `inspect` fails, for example when the daemon is down? Can that turn into a false
  "absent" plus exit 0?
- Is the comment accurate, and is anything stale left in it?
- Did any CI or other caller depend on the old final message `removed …`? Search `.github`, `docs` and `infrastructure`.
