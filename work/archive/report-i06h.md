# Report — I06h

Phase 3 completion-gate manifest and closing report.

## What changed

- Extended the existing `check` target to the exact manifest order:
  `format-check lint typecheck test integration check-generated build up seed seed-identities
  smoke e2e`.
- Appended the I06 closing section to `docs/identity/report.md`: product outcome, acceptance
  disposition, an explicitly unmeasured ordered-gate table, limitations, the `smoke`/`e2e`
  interaction, task-cut evidence and the next phase.
- Left `Vary: Cookie` untouched. It already landed with its browser proof in I06g.
- Left the guards, header, `safeReturnTo`, fixtures, specs and `work/tasks.md` untouched.

`check` is no longer a Compose-free local target. The original seven `AGENTS.md` targets retain
their relative order, and the manifest extends them with generated drift, stack startup, both
seeds, smoke and browser checks. `build` also runs twice in effect because `up` already builds;
that is the recorded manifest and CI behavior, accepted without another mechanism.

## Acceptance rows owned

| Row | Status |
|---|---|
| G38 | Awaiting the reviewing session's `smoke` step inside the complete ordered run |
| G41 | The target contains the exact sequence; its clean-checkout `make check` execution and measured journal await the reviewing session |

No unobserved result or test count was written into the phase report.

## Rate-limit interaction to settle

The login zone is `10r/m` with `burst=5 nodelay`. `smoke` sends thirty concurrent login attempts,
which can leave one source's leaky bucket at its ceiling; it recovers roughly one slot every six
seconds. The limiter key is `$binary_remote_addr`: smoke reaches Nginx from the `frontend`
container, while Playwright reaches `localhost:8080` from the host. Standard Compose networking
presents different source addresses, so the prediction is that smoke's hot bucket does **not**
carry into `e2e`. This has not been measured in the required order. If an environment collapses
both paths to one source and E2E receives a legitimate 429, the fix belongs in the ordering or the
specs' sign-in count, never in the security limit.

## Verification performed here

| Check | Observed result |
|---|---|
| `make -n check` | pass — emits all twelve targets in manifest order, including `seed-identities` before `smoke` |
| `make format-check` | blocked — frontend `oxfmt --check` passed on 84 files; Docker socket denied the Go formatting half |
| `make lint` | blocked — frontend oxlint and the 48-file architecture check passed; Docker socket denied `go vet` |
| `make typecheck` | pass — `tsc --noEmit` |
| `make test` | blocked before test collection — Rstest could not bind `::1:3000` (`EPERM`); the Go half was not reached |
| `git diff --check` | pass |

The requested combined command stopped at the denied Docker half of `make format-check`; the
remaining targets were then invoked individually in the same order. No target was edited to work
around the sandbox. `integration`, `check-generated`, `build`, `up`, `seed`, `seed-identities`,
`smoke`, `e2e` and the complete `make check` were not run here.

## Left for the reviewing session

1. Run `make check` from a clean checkout with Docker, Compose and Chromium available.
2. Record every observed gate result and count in the awaiting table in
   `docs/identity/report.md`; do not infer a skipped result.
3. If `smoke` leaves the login bucket hot and `e2e` receives 429, correct the ordering or reduce
   repeated spec sign-ins without weakening the login limit, then rerun the complete sequence.
4. Update `work/tasks.md` only after G38 and G41 pass.

The required focused commit was attempted with only `Makefile` and
`docs/identity/report.md`, but Git could not create `.git/index.lock` because repository metadata
is read-only in this workspace. No file was staged or committed. The supplied untracked
`work/brief-i06h.md` remains unchanged and must not be included in the task commit.
