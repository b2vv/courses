# Delta review: package `ms3c-reset-w3` (Makefile `e2e-reset`)

You are an independent reviewer: read-only, no edits, no Docker commands. Your w2 verdict was SHIP with one Minor: the compose
`config` exit status was masked by node (`docs/multi-subtitles/_wip-ms3c-reset-review-codex-w2.md`). The w2 Makefile had sha256
`ee9f995a…`; w3 has `dcba12dd67a6b34d9e5b796dc27ccb238dfa05b8903f418101d429e72c7c081d`. The full diff against HEAD `7b20d2c` is
`git diff -- Makefile`; the only change since w2 is the resolver lines (`Makefile:129-132`).

The change: `config=$$($(E2E_RESET_COMPOSE) config --format json) || { …exit 1; }`, then
`resolved=$$(printf '%s' "$$config" | node -e '…') || { …exit 1; }`.

## Author's proof (Kiri, probe project)

- Mixed case: removed / absent / absent, exit 0.
- A `docker` wrapper that runs the real `config` and then exits 32 → "e2e-reset: cannot resolve the compose configuration", exit 2,
  and the probe volume was **not** removed. A later normal run removed it.
- The 7 `courses_*` volumes were untouched.

Return SHIP or BLOCK. Confirm that the Minor is closed and that the delta introduces no regression: `printf '%s'` on large JSON, and
`$$config` quoting.
