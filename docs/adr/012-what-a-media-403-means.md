# ADR 012 — A refused segment does not decide; the heartbeat does
Status: accepted, 2026-09-11. Amends the 403 branch of the recovery table in
`docs/playback/plan.md` §3 D-7.

The edge answers **403 identically** to an anonymous caller, an evicted session, an expired lease
and a revoked course. That is deliberate: A7 requires denials not to reveal what failed, and row
R29 pins it. The consequence took two days and a measurement to see.

**A signal that distinguishes nothing must not decide anything.** The player used to treat a
refused segment as terminal - it showed "playback stopped, the course is no longer available" and
required an explicit action to return. For an expired lease that message is simply false: the
student's access is intact, and the correct response is a silent re-mint. The player could not
tell the difference, and the plan recorded that inability as a limitation rather than as a
conclusion.

So a media 403 now **stops loading and asks the heartbeat**, which is the only place the cases are
distinguishable:

| Heartbeat says | What happened | What the player does |
|---|---|---|
| 410 | the lease expired | silent re-mint, playback continues |
| 409 | another device took the slot | stop, and say so |
| 403 | access was revoked | refuse, return is an explicit action |
| 204, session unchanged | ownership is fine | the refusal was about access after all |

The last row is load-bearing. Without it a genuinely revoked course loops: the segment is refused,
the heartbeat says the slot is ours, playback resumes, the segment is refused again. "Session
unchanged" is what separates that from a re-mint, because a re-mint also answers `continue`.

**A second rule, narrower but real:** a 403 addressed to a `playbackId` that has already been
replaced says nothing about access at all - it answers about a session that no longer exists - and
is ignored outright. Keying this on "a mint is in flight" is not enough; that flag clears the
moment the mint resolves, while a request already on the wire lands afterwards.

**How this was found, since it is not findable by reading.** The O1 measurement held a segment in
flight on purpose while the lease expired behind a pause. Nothing in ordinary playback produces
that timing, and no test had produced it in two days of trying - the row had been recorded as an
inconclusive measurement rather than as a defect it was carrying. A measurement that found nothing
and a measurement that never happened look identical from outside.

Evidence: `docs/playback/plan.md` §3 D-7 and §11 (O1), rows R38 and R29;
`frontend/src/application/playback/recovery.ts`.
