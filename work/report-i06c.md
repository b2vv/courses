# Report — I06c

Protected `returnTo` validation.

## What was implemented

- Narrowed the existing import-free `safeReturnTo` function to accept only `/learn` and
  `/admin` paths, including descendants.
- Split the target at the first `?`, validated only the path, and rejoined a non-empty search
  string unchanged.
- Removed one trailing slash before the protected-prefix check and returned the normalized path;
  an empty search marker is also omitted.
- Rejected every literal `.` or `..` path segment while retaining the existing rejection of
  absolute, protocol-relative, backslash, encoded-slash-prefix, and control-character forms.
- Updated the login-search parser test so its empty-string fallback still removes unusable
  targets instead of emitting a `returnTo` key.

No call-site signature changed. `frontend/src/routes/login.tsx` still uses the default `/learn`
fallback, while `frontend/src/app/router/login-search.ts` still passes `''` so invalid URL input
is omitted from the parsed search object.

## Acceptance rows covered

| Row | Test evidence |
|---|---|
| G15a | `keeps a deep protected path with search` proves `/learn/go/intro?lesson=3` is preserved exactly. |
| G15b | `keeps an administration path with a trailing slash` proves `/admin/` becomes `/admin`; `keeps a protected path with empty search` proves `/learn?` becomes `/learn`. |
| G16 | The `refuses %s as a return target` table contains all eleven specified hostile inputs and maps each to `/learn`. The existing raw control-character test remains. |

The focused command
`npm test -- src/domain/identity/identity.test.ts src/app/router/login-search.test.ts` passed:
36 tests in 2 files, 0 failures, 0 skipped. The same command was first run before the production
change and failed in the expected eight assertions, recording the red side of the TDD cycle.

## Verification

The implementer's sandbox denied the Docker socket and a localhost port bind, so it could run
only the frontend halves. Every row below was re-run unsandboxed from the reviewing session.
The original attempt stopped inside `make format-check`, so the
remaining targets were then invoked individually in the same order without changing any target.

| Step | Result |
|---|---|
| `make format-check` | pass |
| `make lint` | pass (oxlint, architecture check on 47 files, gofmt, go vet) |
| `make typecheck` | pass |
| `make test` | pass — frontend 158 tests / 18 files / 0 failures; 10 Go packages ok |
| `make integration` | pass — 12 Go packages including Testcontainers Postgres |

Supplemental `git diff --check` passed, and the changed domain file remains import-free.
`make e2e` was not run because the brief assigns its browser rows to I06g. `make build` was not
requested and was not run.

## Review

The required two-axis review used `a6be7b3` as the fixed point. The standards review found no
documented-rule violations or code smells. The specification review found no missing, partial,
extra, or incorrectly implemented behavior.

## Deviations and limitations

There is no implementation deviation from the fourth-edition plan or the brief.

A focused commit could not be created: Git failed to create
`/Users/strelia/projects/courses/.git/index.lock` with `Operation not permitted` because this
workspace exposes `.git` read-only. The supplied untracked `work/brief-i06c.md` was preserved
unchanged. The reviewing session must rerun the unmodified blocked gates with Docker and local
listener access, then commit the implementation and this report without the supplied brief.


## Review note from the verifying session

`safeReturnTo` now carries two gates that overlap completely. The early check on
`value.slice(1, 4)` — rejecting a second `/`, a `\`, or `%2f` — is **fully subsumed** by the new
prefix test: any value whose second character is one of those cannot begin with `/learn` or
`/admin`, so the regex rejects it anyway. Verified case by case against the whole G16 table.

It is kept deliberately. Ponytail does not simplify away input validation at a trust boundary,
and removing a redundant check in a redirect validator trades three dead lines for the chance
that someone later loosens the regex and silently reopens the protocol-relative hole. Recorded
here so the redundancy is a known, deliberate one rather than something a future reader mistakes
for load-bearing: **the prefix regex is the real gate.**

The `/admin//` input returns `/admin/` — a single trailing slash is stripped, a doubled one
leaves the path ending in a slash. Same origin, harmless, and no acceptance row asks otherwise;
noted because "normalized" in the doc comment is slightly stronger than what the code does.
