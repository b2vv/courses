# I06 — route guards, signed-in header and completion gate

Slice of Phase 3. The phase specification is [../spec.md](../spec.md); this file does not
restate it, it narrows the last slice to a shared understanding of intent.

Artifacts for a slice live in `docs/<phase>/<slice>/` — an extension of the manifest's
`## Артефакти`, which names a directory per phase but not per slice.

## Проблема

Phase 3 built login, a server-side session and a login screen, and nothing in the product
uses any of it.

- `/learn` and `/admin` render the Phase 1 "locked" placeholder to everyone, identically,
  whether the visitor is anonymous, a student or an administrator. The screen tells a
  signed-in administrator to go and sign in.
- A signed-in visitor sees no evidence anywhere that they are signed in, and has no way to
  sign out other than deleting a cookie by hand.
- The database seeds no identity, so no one can exercise login without inserting a row
  manually. Every live check in I03–I05 needed a temporary user created and deleted by
  hand, which is not a repeatable check and cannot run in CI.

The consequence is that five of the seven phase acceptance scenarios (A01, A02, A04, A05,
B01) have no evidence, and Phase 4 cannot begin, because it needs an authenticated
administrator to act as.

## Користувацькі сценарії

1. **A student signs in and starts learning.** After signing in they arrive at the
   learning area rather than at a locked screen.
2. **An administrator signs in and opens the administration shell.** An instructor sees
   the same shell; the capability matrix already exists in the domain and the screen obeys
   it.
3. **A student opens the administration shell.** They are refused, and the refusal says
   they lack the permission — not that they should sign in, which they already did.
4. **An anonymous visitor opens a protected address**, perhaps from a bookmark. They are
   sent to sign in, and after signing in they land where they were going, not on a
   default page.
5. **A signed-in visitor is reminded who they are.** Any page shows their identity, and
   offers to sign out. Signing out returns them to a public page and the protected
   addresses become locked again.
6. **A session ends while the tab is open** — it expired, or it was revoked from another
   device. The next navigation to a protected address behaves exactly like the anonymous
   case rather than showing a broken screen.

## Acceptance criteria

Discharges these rows of the phase acceptance table ([../plan.md](../plan.md) §Acceptance
scenarios), which stay the authority on wording:

| Row | What it demands of this slice |
|---|---|
| A01 | Student fixture signs in; `/learn` opens |
| A02 | Admin fixture signs in; `/admin` opens |
| A04 | A valid session cookie makes the signed-in header appear in the raw SSR HTML, and hydration does not flash an anonymous state first |
| A05 | Logout returns 204, expires the cookie, and a protected address is locked again afterwards |
| B01 | A student is refused `/admin`; instructor and admin are allowed; the screen and the domain policy express the same matrix |

Added by this slice, not yet in the phase table:

| ID | Criterion |
|---|---|
| C01 | Signing in from a protected address returns the visitor to that address; an address that is not internal is ignored in favour of the default |
| C02 | The identity fixtures are idempotent: seeding twice leaves the same rows, and the credentials are documented as local-only |
| C03 | The whole ordered gate sequence passes on a clean checkout, and the browser suite covers sign-in, role refusal and sign-out against Compose |

## Out of scope

- Server-side authorization on business endpoints. The phase specification defers it until
  a protected endpoint exists to consume it, which is Phase 4. Client guards are navigation
  convenience, never the security boundary.
- Any administration functionality behind `/admin` beyond the shell that proves the role
  check. Course, module and lesson management is Phase 4.
- Registration, password reset, email verification, device management.
- `__Host-` cookie prefix (`S02`) and the browser security headers (`S01`); both wait on
  TLS.
- Real learning content behind `/learn`. There are no lessons yet, so the area shows an
  honest empty state.

## Рішення GATE 1

Taken with the user through `grilling`; the reasoning lives in
[../_wip-i06.md](../_wip-i06.md).

| # | Question | Decision |
|---|---|---|
| 1 | When does SSR resolve identity? | Only when the request carries a session cookie. No cookie means anonymous by definition, so public storefront traffic pays nothing. This also settles the anonymous header: it is what the root already renders |
| 2 | How is a role refusal presented? | As a 404. Not a security claim — the aim is to keep the administration area out of sight of casual probing, so a wrong path looks like a wrong path. `notFound()` is already this repository's answer for an unusable address |
| 3 | What does an anonymous visitor get? | Both addresses send them to login with `returnTo`. The role is checked **after** signing in: a permitted role lands on the address they asked for, a student gets 404. Revised from an earlier answer — see below |
| 4 | What does `/learn` show? | An honest empty state with a way into the catalog, built on the existing `EmptyState`. It is shown to every signed-in visitor, whatever their role — see below |
| 5 | What is in the signed-in header? | The display name as text, a sign-out control, and an administration link rendered only for the roles allowed to use it |
| 6 | How are demo identities kept out of a shared environment? | The fixtures apply only on an explicit confirmation, and the credentials are documented in the README beside it. **Mechanism amended twice during planning; the intent the user fixed is unchanged both times.** `APP_ENV` cannot carry it — the `Makefile` never loads `.env`, `docker-compose.yml:61` defaults it to `development`, and CI runs `make seed` with it unset. A `SEED_IDENTITIES` variable through a psql `\if` cannot either: `Makefile:23` keeps the whole `psql` call inside single quotes within `sh -c`, so the variable expands **in the container**, where `docker compose exec` never forwarded it — measured live, the branch is silently skipped and the command exits 0, so CI would seed nothing and report success. The mechanism is therefore not a variable at all: the fixtures live in `backend/sql/seed-identities.sql` and apply through their own `make seed-identities` target. The explicit confirmation is the target name. A target that is not invoked cannot seed anything, so the guarantee holds by construction rather than by a runtime check |

### `/learn` is about enrolment, not role

`CanLearn` is true for every valid role in the domain (`backend/internal/identity/domain/user.go:23`)
and in its frontend twin (`frontend/src/domain/identity/identity.ts:36`), and the phase plan
gates `/learn` on being *authenticated* rather than on being a student. That is deliberate:

| What | Where it lives | Gated by |
|---|---|---|
| Courses you have access to **as a learner** — bought, or granted | `/learn` | being signed in |
| Courses you **teach** | instructor workspace, its own phase | instructor or admin role |

So an administrator or an instructor who bought someone else's course sees it on `/learn`
exactly as a student does. What they do **not** see there is their own taught courses;
those are their work, not their learning, and they belong to the workspace phase.

For this slice the consequence is nil: enrolment does not exist as a table or a module, so
all three roles see the same empty state. The distinction starts to matter when enrolment
arrives, and is recorded now because it shapes that schema.

### What decision 2 and 3 do and do not mean

Reading the two together is easy to get wrong, so the full matrix is written out:

| Visitor | `/learn` | `/admin` |
|---|---|---|
| Anonymous | login with `returnTo`, returned after signing in | login with `returnTo`, then the role decides |
| Signed in, student | opens | 404 |
| Signed in, instructor or admin | opens | opens |

**Why this changed.** The first answer gave anonymous visitors a 404 on `/admin` too, so
that the address looked like it did not exist. That cost the deep link: a bookmark opened
from a signed-out state answered 404 and carried no `returnTo`.

It was withdrawn once the hiding was measured rather than assumed. `/admin` is present
verbatim in the built client bundle — `grep -r /admin frontend/dist/client` finds it — so
the route was never concealed from anyone willing to open developer tools. Hiding only ever
worked against casual address guessing, and a login screen instead of a 404 tells such a
visitor almost nothing the bundle does not already say.

What the decision protects is unchanged and is the case the user described: a **signed-in**
student probing addresses still gets 404, learning nothing about an administration area
existing.

**A typed address is a full page load, so the guard runs before identity is known.** A
guard that reads "not yet resolved" as "anonymous" would answer 404 to an administrator on
their own administration area, briefly or permanently. The guard must wait for the identity
to resolve rather than defaulting to anonymous — the same requirement as A04's no-flash
rule, seen from the other side, and the easiest thing in this slice to get wrong.

### Withdrawn derivation

An earlier draft derived **two fixtures instead of three**, on the grounds that B01's
instructor row was covered by the domain policy test. The user then stated that an
instructor is not "an administrator with the same shell": an instructor reaches the
administration area scoped to their own data — their students, their videos, the comment
threads on those videos, and their students' progress.

That makes instructor a distinct behaviour rather than a row in a matrix, so the slice
seeds **three** fixtures — student, instructor, admin — and B01 is exercised in full.

### What this slice must not imply

None of the instructor scoping can be built here, and the shell must not pretend
otherwise. Courses have no owner (`backend/migrations/00001_courses.sql` carries no
instructor, author or owner column), there is no enrollment, no video, no progress and no
comment thread anywhere in the codebase. `/admin` in I06 is the smallest shell that proves
the role check and says plainly that management arrives in Phase 4.
