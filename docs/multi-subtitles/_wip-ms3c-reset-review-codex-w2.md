**SHIP — both w1 Majors are closed. One non-blocking Minor remains.**

1. **Minor — Compose config’s exit status is masked by Node.** [Makefile:129](/Users/strelia/projects/courses/Makefile:129)  
   If `config` emits complete valid JSON and then exits nonzero, Node succeeds and deletion continues. An in-memory stub emitting JSON then returning 32 reproduced recipe exit 0. Ordinary configuration errors produce no valid JSON and correctly fail, so this does not reopen either Major. Capture and check `config` separately before parsing to make “resolver failure → failure” unconditional.

Both w1 fixes verified:

- **Lookup failures closed:** [Makefile:136](/Users/strelia/projects/courses/Makefile:136) and [Makefile:139](/Users/strelia/projects/courses/Makefile:139) check Docker’s status directly. Initial and post-removal lookup failures stop the recipe.
- **Project-name mismatch closed:** [Makefile:125](/Users/strelia/projects/courses/Makefile:125) supplies the same explicit files to `down` and `config`; deletion uses resolved project/volume names. The basename guess is gone.

Requested checks passed:

- Make expansion preserves the Go template’s inner double quotes.
- Empty labels leave the name match working; partial labels do not falsely satisfy the two-label match.
- The single-quoted Node program has no `$` or embedded-single-quote hazards.
- Explicit `-f` flags override `COMPOSE_FILE` for both commands. [Docker documentation](https://docs.docker.com/reference/cli/docker/compose/#set-up-environment-variables)
- Comments accurately describe volume selection, reporting and retained monitoring data, subject to the resolver-status caveat above.

Both package hashes matched before and after review; the working Makefile diff matches the supplied package. Fourteen in-memory probes passed, plus the additional fault probe above. No files were edited or Docker commands run; the author’s live Kiri proof was not independently repeated.