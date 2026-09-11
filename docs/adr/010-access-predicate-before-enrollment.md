# ADR 010 — The playback access predicate until enrollment exists
Status: accepted, 2026-09-11. Supersedes ADR 005's "Go checks session and enrollment" clause for
Phase 6 only; everything else in ADR 005 stands.

Authorize playback for **any authenticated user who can see the course**, and nothing more. There
is no enrollment table, no purchase, and no entitlement in this system yet; Phase 8 introduces
them. ADR 005 was written assuming enrollment would already exist by the time bytes were served,
and it does not, so the predicate has to say what it actually checks.

Concretely, `authorize` answers 204 when the caller has a valid session, the lesson exists, and
the course is visible to that caller. Visibility is the course's own rule, not a new one: a
published course is visible to everyone signed in; an archived or draft course is visible only to
its instructor and to administrators. That last part is not a convenience - row R34 exists because
an instructor watching their own archived course is a real case, and a predicate phrased as
"published only" would have refused them.

The predicate is one function with one implementation, shared between the media authorization
endpoint and the learning read model, because two copies of an access rule diverge and the
divergence is invisible until someone sees bytes they should not. Rows R18 and R19 pin that they
answer alike.

**What this deliberately does not do.** It does not gate on payment, enrollment, or any per-user
grant, and it must not be read as a decision that such gating is unnecessary. Phase 8 replaces the
body of this predicate; the seam - one function, called from both places, with the courses that
call it unchanged - is what this ADR fixes in place so that replacement is a body swap rather than
a redesign.

**Why it is safe to ship without entitlement.** Nothing in the catalogue is paid yet. Every
published course is intended to be watchable by anyone with an account, so the predicate is not
loosened from some stricter intent - it states the intent the product currently has. When that
changes, the row that changes with it is this one.

Evidence: `docs/playback/spec.md` G1 and A1-A5, `docs/playback/plan.md` §3 D-1, rows R18, R19,
R34 and R08 in §8.
