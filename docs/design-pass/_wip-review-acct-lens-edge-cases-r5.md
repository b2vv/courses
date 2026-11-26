NEXT: Read the shared agent-entry rules before reviewing.
FACT: Read the shared agent-entry rules.
NEXT: Verify the repository root and read its local rules and the requested lens contract.
FACT: Checked the repository root, revision, local routing and skill contract.
DECISION: Review the edge-cases lens only, without delegation or runtime checks.
NEXT: Read §18, the owner audit, and the four prior edge-cases findings.
FACT: Read section 18, the owner audit, the four prior working files and the shared anti-pattern reference.
NEXT: Compare revision 5 with the implementation and existing account journeys.
FACT: Indexed each historical finding and the current decision and acceptance rows by line.
NEXT: Trace the link, session and 2FA state transitions against their source paths.
FACT: Isolated V5–V14 and the ACCT verification rows, including historical disposition summaries.
NEXT: Inspect exact route and backend error branches, freshness and concurrency.
FACT: Traced token consumption, resend invalidation, 2FA writes, and the route/session bindings.
NEXT: Check which 2FA commands can overlap, how session data is refreshed, and whether the acceptance fixtures expose failures.
FACT: Mapped 409s back to all four 2FA handlers and checked the session query, current settings view, controls, CSS, and e2e anchors.
NEXT: Resolve the remaining historical dispositions and check whether revision 5 introduces additional state or recovery gaps.
FACT: The first targeted comparison failed in the tool wrapper before execution.
NEXT: Retry the same read-only comparison.
FACT: Compared revision 5’s actual diff with existing response mappings and browser journeys.
NEXT: Validate the candidate findings and record the final per-revision dispositions and verdict.
FACT: A second wrapper typo prevented the targeted read from running.
NEXT: Read the session cancellation precedent and confirm source anchors.
FACT: The shared settings guard joins pending reads, and the existing sign-out path cancels a pending session query before replacing its cache.
FACT: The backend rejects both link routes before writing on origin failure and returns already-enabled for enroll as well as confirm.
NEXT: Append the historical dispositions, findings, and lens verdict.
FACT: Correction to the preceding session note. The existing sign-out path sets null after navigation; it does not cancel /me. V2 itself supplies the cancel-before-read precedent, and identity-queries.ts passes the query signal to current().
DECISION: Revision 1 dispositions, in original order. (1) The unconditional delivery claim is closed by neutral 204 feedback; revision 5's new newest-mail claim has a separate concurrency gap below. (2) Retained invalid cookie closed by V1/ACCT-3. (3) Coexisting resend controls closed by V1's dispatch-time guard and ACCT-1. (4) Consumed-token recovery is specified by S3/S4 and E3–E5; the A6 address check has a freshness gap below. (5) Cross-token result leakage closed by the keyed screen and ACCT-7. (6) Done with a delayed refresh closed by V5a's immediate cache update and ACCT-8. (7) Stale copied feedback closed by the child lifecycle and ACCT-18. (8) Enable-button stretch closed by V11/ACCT-15. (9) Picking the original Mailpit message closed by excluding its ID and covering both links. (10) Shared validation scope closed by V6's TextField/TextArea boundary.
DECISION: Revision 2 dispositions, in original order. (1) Delivery guarantee closed by neutral V1 acknowledgement, subject to the new ordering issue below. (2) Pre-render duplicate resend closed by the synchronous guard. (3) Stale or failed A3 session read closed by V2's cancellation and S0/ACCT-5. (4) Pending-token reopen closed at the client-request boundary by V4a's join and ACCT-7. (5) Failed 2FA reconciliation read closed by lock and recheck in V5a. (6) Lost regeneration response closed by the old-code warning, fresh proof, and admin fallback. (7) Edits during save closed by V10 read-only fields. (8) Secret clipping covered by V15 content sizing and ACCT-19's visible-full-value check.
DECISION: Revision 3 dispositions, in original order. (1) Stale token selection closed by keyed instances. (2) GC resubmission closed as a retention issue by a documented settled-reopen policy; the policy still equates completion with token consumption below. (3) Lost 2FA focus closed by retaining the form and specifying the return target. (4) Definitive refusal has initial states in both tables and covered refusal recovery on A3; V3's explicit retry text still permits a same-document 403 loop, but a settings exit remains available.
DECISION: Revision 4 dispositions, in original order. (1) A→B→A while the client request is pending closed by V4a and ACCT-7. (2) S5 verified-user shortcut closed by sending the token again and requiring a definitive link error for S3. (3) A direct 401 now replaces the signed-in cache and is tested, but an older in-flight /me can overwrite the replacement. (4) The reported confirm, disable, and regenerate conflicts are covered; enroll emits the same conflict and is missing. (5) Manual retry focus closed by V5a and ACCT-9 for both unchanged exits.
DECISION: Findings.

[edge-cases] Stale session data can overwrite the settings 401
where: §18 V5a and ACCT-9, lines 665 and 728.
why: setQueryData(null) leaves a pending session request active. A /me that read the user before session revocation can deliver its signed-in response after the settings write returns authentication_required. It can replace null before useRedirectWhenSessionEnds observes it, or feed the login loader the same stale response. ACCT-9 only seeds a cached user. identity-queries.ts:25–34 and require-identity.ts:11–25,30–51 show the shared key and both readers.
action: Specify cancellation of the exact pending session query before publishing null, or an equivalent ordering guarantee. Hold a pre-revocation signed-in /me response across the 401 and verify the login redirect with the settings return path and no restoration from the held response.
weight: blocking

[edge-cases] Enroll can return the same already-enabled conflict as confirm
where: §18 V5a and ACCT-9, lines 664 and 728.
why: An off-state tab can press “Увімкнути 2FA” after another tab enables it. EnrollTwoFactor checks Enabled before issuing a secret (two_factor.go:58–65), and handler.go:201–210,851–854 returns the same 409 that V5a covers only for confirm. The stale tab has no planned transition out of the off card and repeating enroll returns the same conflict.
action: Include the enroll 409 in the reconciliation table. Specify the focus target while the off-card action stays mounted and after fresh /me changes the card, and add the stale-off-card route case to ACCT-9.
weight: blocking

[edge-cases] A6 recovery can show the email from a read started before the change
where: §18 V3/V9 and ACCT-6, lines 636–639, 678 and 725.
why: An E5 lost response can follow a committed email change. Its retry returns token_used and E4, so neither outcome runs the confirmation route's existing onSuccess refresh (email.change.confirm.tsx:18–24). Settings' requireIdentity uses fetchQuery on the shared key and joins an in-flight /me started before the commit (require-identity.ts:11–19), then its observer has refetchOnMount disabled (settings.account.tsx:17–20). The user can see the old address and request an unnecessary new change.
action: Require a session snapshot started after the A6 outcome when following E4/E5 to settings. Hold a pre-change /me with the old address across the lost-response/reopen sequence and verify that only the fresh address informs recovery.
weight: blocking

[edge-cases] A completed link request may have left the token unused
where: §18 V4a and ACCT-7, lines 643, 702 and 726.
why: The documented reopen rule treats every completed POST as consumed and says the server returns token_used. Both routes can complete with origin_forbidden before reaching tokenAction (handler.go:115–121,514–523). Reopening that same link after a transient refusal can produce 204, which the current acceptance wording treats as impossible.
action: Apply the state table to the actual response on every new mount. Scope the token_used expectation to a token the server consumed or invalidated, and add a settled pre-write refusal followed by a successful reopen.
weight: blocking

[edge-cases] Newest mail does not establish which verification token is valid
where: §18 V1, lines 607 and 614.
why: accounts.go:133–140 invalidates, issues, and sends in separate steps. With separate tabs or devices, request A can issue, B can invalidate A and send, and A can send last. The newest delivered message can therefore contain the invalid link. If both invalidations run before either issue, both issued links can survive. The local mutual exclusion and the sequential Mailpit test do not establish the unconditional old-link claim across deliveries.
action: Keep the neutral acknowledgement and tell users what to do if the link they receive fails. Scope the invalidation claim to the sequential journey proved by ACCT-2 and avoid promising validity or invalidity from inbox order.
weight: blocking

FACT: Static review of revision 5 and current source. No Make, Docker, or e2e runs. The working file is the only edit.
DECISION: lens verdict (5 blocking findings).
