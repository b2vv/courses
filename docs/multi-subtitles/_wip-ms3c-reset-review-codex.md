**BLOCK — two Major findings.**

1. **Major — lookup failures become successful “absent” results.** [Makefile:127](/Users/strelia/projects/courses/Makefile:127)  
   `find_volumes` returns `sort -u`’s status, masking failures from both Docker commands. The assignments also lack failure checks. If `compose down` succeeds and volume lookups subsequently fail, the target prints `absent` three times and exits 0 while volumes survive. A shell-only probe reproduced this. A daemon already down would instead fail the initial `compose down`.  
   **Required fix:** propagate lookup errors and distinguish a missing volume from an unavailable lookup.

2. **Major — the new destructive lookup uses a project name that can differ from Compose’s.** [Makefile:128](/Users/strelia/projects/courses/Makefile:128), using the fallback at [Makefile:125](/Users/strelia/projects/courses/Makefile:125)  
   `docker-compose.yml:1` specifies `name: courses`, which takes precedence over the directory basename. In a checkout named `courses-review`, with `COMPOSE_PROJECT_NAME` unset, Compose stops `courses`, but this recipe inspects and removes `courses-review_*`. Unused volumes belonging to another stopped project can be deleted, while the actual Courses volumes survive. The inherited project-name guess now drives deletion by name. [Docker’s precedence rules](https://docs.docker.com/compose/how-tos/project-name/) confirm the mismatch.  
   **Required fix:** obtain the resolved project or volume names from the same Compose configuration.

Other requested checks:

- No **volume-level** `name:` overrides exist in the three specified files.
- `{{.Name}}`, `$$` escaping and the function syntax are correct. Deduplication and splitting `$$ids` work for valid Docker volume names. [Inspect formatting documentation](https://docs.docker.com/reference/cli/docker/volume/inspect/)
- Underscores introduce no ambiguity. An explicit different `COMPOSE_PROJECT_NAME` consistently targets that project. Empty/unset works in the current `courses` directory; finding 2 covers the mismatch elsewhere.
- No caller parsing the old final message was found in `.github`, `docs` or `infrastructure`.
- The explicit list and retained monitoring volumes are unchanged. The comment’s resolved-name and survival guarantees need correction alongside these fixes.

No files were edited or Docker commands run. Verification used static inspection and in-memory shell stubs.