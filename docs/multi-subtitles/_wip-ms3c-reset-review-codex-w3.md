**SHIP — w2 Minor closed; no regression found in the w3 delta.** Makefile SHA-256 matches `dcba12dd…c081d`.

- [Lines 129–132](/Users/strelia/projects/courses/Makefile:129) check Compose’s exit status before parsing. Valid JSON followed by exit 32 stops execution before volume lookup/removal.
- `printf '%s' "$$config"` safely preserves the JSON: fixed format, quoted expansion, no splitting, globbing or re-evaluation.
- In-memory `/bin/sh` probes passed for normal JSON, malformed JSON, producer failure and **8,388,825-byte JSON**. `printf` is a shell builtin, avoiding an external argument-size limit.

Only resolver lines changed since w2. No files edited or Docker commands run; the author’s live Kiri proof was not independently repeated.