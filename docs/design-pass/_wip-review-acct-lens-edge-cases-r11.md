NEXT: Verify repository root and read local rules, briefing, and the edge-cases contract.
FACT: Repository root verified.
FACT: Read the local routes, task pointers, and plan-critique contract.
DECISION: Apply only the edge-cases lens to §18.
NEXT: Pin revision 11 and read §18, revision 10 findings, and the audit.
FACT: Pinned §18 to 4eae6978 and read the revision 10 record and audit pointers.
NEXT: Trace account request outcomes, route transitions, and shared control behavior against the amended acceptance rows.
FACT: Read operative V1–V10, Л1–Л6, acceptance rows, the revision diff, and audit decisions.
NEXT: Verify the three prior findings against route, cookie, and refresh code.
FACT: Scoped sources match the pinned snapshot. Traced route hooks and located the real resend/session-cookie handlers and account fixtures.
NEXT: Check whether the new ACCT-3 cookie change reaches resend and survives any SameSite or identity branch.
FACT: Invalid-present cookies take the real resend 401 branch, and currentUser maps 401 to null. Registration already yields a copyable unverified browser session.
NEXT: Trace 2FA and shared controls for concrete uncovered boundaries, then verify the offline A6 contract and proof.
FACT: V3 now requires a detached A6 refresh after 204. Current onSuccess returns the refresh promise, so implementing V3 must stop returning that promise.
FACT: V2 no-token read and ACCT-3 invalid-cookie fixture address the other revision 10 blockers.
NEXT: Cross-check concrete dependency failures and boundary inputs against the acceptance claims introduced in revision 11.
FACT: Verified token consume/invalidate ordering, 2FA conflicts and partial writes, shared FieldError, and the protected settings guard.
NEXT: Decide whether the revised paused-refresh proof distinguishes immediate success from the current awaited callback.
DECISION: All three revision 10 blockers are closed in revision 11. V2 supplies a tokenless session classification, ACCT-3 replaces the cookie, and V3/Л6 agree on immediate A6 success.
FACT: Query Core awaits the mutation success callback, but substitutes a resolved promise for paused refetches. A paused-refetch test alone cannot detect an awaited refresh during online latency.
NEXT: Record the non-blocking proof gap and finalize the lens verdict.

[edge-cases] A6 offline proof misses a slow online session refresh
where: §18 V3 (661), Л6 (736), ACCT-6 proof (771).
why: The current A6 onSuccess returns the invalidation promise (email.change.confirm.tsx:18–24). Query Core awaits that promise before publishing mutation success (mutation.ts:249,274), but invalidation resolves immediately for a paused refetch (queryClient.ts:337–343). The proposed offline test can therefore pass with the current callback while a slow online /me still delays E2 and its settings link.
action: Add an ACCT-6 route case that holds an online /me pending after 204, asserts E2 and its settings link before releasing it, and then releases it. Keep the separate offline case.
weight: non-blocking

FACT: One assigned lens completed against 4eae6978. Static review only. No make check, Docker, or E2E ran. Only this working file was edited.
lens verdict (no blocking findings)
