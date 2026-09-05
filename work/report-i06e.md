# Report — I06e

Regression coverage for the existing `/learn` and `/admin` empty states.

## What was added

- Extended `frontend/src/routes/route-guards.test.tsx` with the parameterized G03 route test for
  `student`, `instructor`, and `admin`. Every role remains on `/learn`, renders the same exact empty
  state and catalog link, receives no not-found answer, and renders no instructor-specific
  “Ваші викладацькі курси” content.
- Added the G40 component-level route test for an administrator. It requires the exact later-phase
  message and catalog link, with no additional main content, table, or list representing courses,
  students, videos, progress, or comment threads.
- Added no Playwright specs; their halves remain assigned to I06g.

## Acceptance rows

| Row | Evidence |
|---|---|
| G03 | `G03 opens the same /learn empty state for the %s role`, parameterized over all three roles |
| G40 | `G40 keeps the /admin shell empty until the later phase` |

## Already done

I06b had already replaced both former locked placeholders with `EmptyState`, a catalog link, and
Ukrainian/English copy. The administration message already says that course management arrives in
the next phase. Those routes and translations were read and left unchanged, as required.

The task was smaller than its 0.5-hour estimate because all product behavior and copy already
existed; only the two missing regression tests were needed.

## Verification

The required targets were attempted in order from `/Users/strelia/projects/courses`. The initial
combined command stopped at the first denied target; the remaining targets were then attempted
individually in the same prescribed order without modifying any target.

| Step | Result |
|---|---|
| `make format-check` | pass |
| `make lint` | pass |
| `make typecheck` | pass |
| `make test` | pass — frontend 175 tests / 19 files / 0 failures; all Go packages ok |
| `make integration` | pass — 12 Go packages, 0 failures |

Supplemental focused evidence:

- `npm test -- src/routes/route-guards.test.tsx` passed: 20 tests, 0 failures, 0 skipped.
- `git diff --check` passed after the test change.
- `make build` and `make e2e` were not run; the I06e brief does not request them and assigns the
  browser halves to I06g.

## Review

The required two-axis review used `5f0678e` as the fixed point and reviewed the uncommitted test
diff. The Standards axis reported no findings. The Specification axis reported no missing or
partial requirement, scope creep, or incorrect implementation.

## Commit and limitations

A focused commit was attempted with message
`test(identity): cover learn roles and empty admin shell`. Git could not create
`/Users/strelia/projects/courses/.git/index.lock`: `Operation not permitted`, so nothing was staged
and no commit was created. The supplied untracked `work/brief-i06e.md` remains unchanged.

The reviewing session must rerun the blocked required targets with Docker and localhost-bind
access, then create the focused commit. No product code, copy, i18n key, guard, header,
`safeReturnTo`, fixture, or `check` target was changed.


## Verified from the reviewing session

Re-run unsandboxed. Frontend tests rose from 171 to 175 — four, which is `G03` across three roles
plus `G40`.

Both tests avoid the failure mode that makes an absence assertion worthless. Rather than querying
for strings that were never going to be there, each asserts the **exact full `textContent` of
`<main>`**. Any course table, student list or invented panel added later changes that string and
fails the test, which is what `AGENTS.md:6` needs the row to guarantee. `G40` adds structural
assertions on top: no `<table>`, no `<ul>` or `<ol>`.

One decorative line: `queryByText('Ваші викладацькі курси')` in `G03` searches for a string that
exists nowhere in the codebase, so it passes trivially. It is redundant rather than misleading —
the exact-`textContent` assertion already covers it — and is left alone.

## What the estimate got wrong

The task was smaller than 0.5h because I06b had already built the screens and both locales when it
deleted `LockedArea`: leaving the routes without a component was not an option, and a component
with no text is not a screen. The cut written at edition 4 assumed screens and guards could be
separated; they could not.

Recorded because it is information about the cut, not about the work: a task defined as "the
screens" survived to execution as "two tests for screens that already existed".
