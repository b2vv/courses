# ADR 011 — One playback per account, and how the slot changes hands
Status: accepted, 2026-09-11; implemented and measured in Phase 6.

An account may have **one** playback at a time, across all courses. Pressing play takes the slot;
whoever held it stops. This is a product rule about account sharing, so it is keyed by account -
not by course, which would let one person watch three courses at once, and not by address, which
would punish a household and free a laptop moved between networks.

**Play takes the slot; opening a page does not.** Opening a lesson is browsing. A student who
opens a second tab to read the outline has not started watching, and stopping their other device
for that would be indefensible. The taking action is the one the person meant as "watch this
here".

**The slot is a lease, not a lock.** Ownership lives in Redis under the account, holds the current
`playbackId`, and expires. A session abandoned without closing - a killed tab, a dead battery -
frees the slot when the lease runs out, because nothing else will ever come to release it. Row B16
pins that a ghost does not block.

**The compare-and-renew is one Lua script with three answers, not two.** This is the part most
likely to be "simplified" later, so the reason is recorded here. The script returns:

    -1  no record          the lease expired        -> 410, silently re-mint, playback continues
     0  a different owner  someone else took it     -> 409, stop, say so
     1  ours               renewed                  -> 204, carry on

Collapsing `-1` and `0` into "not ours" makes expiry indistinguishable from eviction, and those
must lead to opposite behaviour: one resumes quietly, the other stops the student and tells them
why. It is Lua rather than WATCH/MULTI because the read and the renew have to be one step; between
a separate GET and EXPIRE another device can take the slot, and we would renew a lease we no
longer hold.

**Two layers, because a client that ignores the first is not hypothetical.** The evicted player
stops on 409 at its next heartbeat. Independently, the edge asks Go about **every** segment and
caches nothing, so an evicted device gets bytes refused on its next request whether or not its
client cooperated. Row B15 proves the second layer in the only form that means anything - the same
path served 200 before the eviction and 403 after - because a bare 403 proved nothing while a
filesystem permission bug was answering 403 to everyone.

**Re-minting after eviction is forbidden on the client.** Play always takes the slot, so an evicted
device that re-mints would take its own slot back and undo the eviction. Returning to playback
after being evicted is an explicit action by the person, never an automatic recovery.

Evidence: `docs/playback/spec.md` A28-A34 and G3, `docs/playback/plan.md` §3 D-1 and D-3, rows
P13, P14, B15, B16, B32, B33, R25 and R26; `backend/internal/playback/adapters/redis/ownership_writer.go`.
