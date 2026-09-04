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

## Відкриті питання

1. **What does `/learn` show a signed-in student in Phase 3?** There is no enrollment and
   no lesson. An empty state that says so is honest; anything richer would claim a
   capability that does not exist.
2. **How is a role refusal presented?** Sending a signed-in student to the login screen
   would be wrong and confusing. A dedicated "not permitted" screen is the alternative.
3. **Where does sign-out live?** A plain control in the header, or behind a menu opened
   from the user's name.
4. **Are the demo credentials written into the README?** They are local-only placeholders,
   like the existing database passwords in `.env.example`, but they authenticate a role
   that will later be able to change content.
5. **Does the header render during SSR for anonymous visitors too?** A04 demands no
   identity flash for a signed-in visitor; the anonymous case has the same requirement in
   reverse.
