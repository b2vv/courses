NEXT: Read the shared entry rules, verify the repository root, and load the logical-gaps contract.
FACT: The shared entry rules are loaded. Repository root and HEAD match the requested review target.
NEXT: Read local repository routing, briefing, and critique contract.
FACT: The single-lens contract is loaded and the pinned commit is checked out on main.
DECISION: Follow the owner’s read-only scope and assess only logical prerequisites, ordering, and guarantees.
NEXT: Locate the revision 11 plan and audit anchors and read the previous findings.
FACT: Revision 10 had three blocking findings covering the bare A3 route, the A6 paused-read promise, and the ACCT-3 browser session fixture.
NEXT: Read the entire pinned §18 and compare it against the audit and revision delta.
FACT: Revision 11 adds an explicit bare-URL trigger, scopes the post-result offline wait to A3, and requires a real invalid-session resend for ACCT-3.
NEXT: Read the operative state tables and the owner-approved audit findings side by side.
FACT: ACCT-2 now tests the bare guest route. Л6 supplies the A6 offline proofs missing from the ACCT-6 row.
NEXT: Confirm the identity route, query, cookie, and browser fixture behavior in code.
FACT: The existing token routes skip the mutation for absent tokens, and the registered fixture exposes the unverified banner before verification.
NEXT: Trace required identity and the signed-in resend state while checking the remaining audit repairs against the UI and backend.
FACT: Settings already fetches the shared session query. A real invalid cookie takes the resend handler through authentication before the mail service.
NEXT: Check the A6 success timing and the A3 provisional versus resolved state against concrete dependencies.
FACT: The A6 route currently returns the refresh promise from onSuccess, and the settings guard reads the same query key.
NEXT: Verify whether the new offline proof distinguishes the promised A6 timing, and finish cross-checking the backend and fixture prerequisites.
FACT: TanStack Query awaits onSuccess before marking mutation success, but invalidateQueries does not wait for a paused refetch. The new offline A6 case can therefore pass with the existing online delay.
NEXT: Verify whether the ACCT-3 proof can discriminate a document navigation, and settle the bare A3 wording against the first-match table.
DECISION: All three revision 10 findings close at their original scope. The new A6 wording adds a broader success-timing promise that needs a discriminating online proof.
NEXT: Confirm the remaining UI and persisted-token anchors, then record the findings and verdict.
FACT: Reviewed pinned main commit 4eae6978, §18, the revision diff, the owner-approved audit #1–#16 and taste decision, and the route, UI, CSS, session, backend, and e2e anchors. Static review only. No make check, Docker, or e2e execution. No other lens working file was opened. No tracked changes were made.

DECISION: Revision 10 bare A3 blocker closed for the guest resend flow. V2 now starts the session read on mount without a verification POST, and ACCT-2 proves null leads to the guest form. The parenthetical S4 label still risks obscuring the signed-in verified branch, as noted below.
DECISION: Revision 10 A6 offline blocker closed. V3 and Л6 explicitly allow E2 with its action after 204 while the refresh is paused. Л6 specifies both the pre-request E0 proof and the paused post-204 proof. The stronger revision 11 claim about timing with an active refresh is a separate gap.
DECISION: Revision 10 ACCT-3 fixture blocker closed. Clone the registered context before verifying its email, let the banner render, then replace the single cookie in that clone at its original domain and path. The real resend handler checks the cookie before calling the mail service. The replacement is also what the fresh login document will send to /me, whose 401 maps to null. Anchors: accounts.spec.ts:260–269, handler.go:346–377 and :558–585, application/identity.go:136–142, session-repository.ts:270–281, LoginView.tsx:35–43.

[logical-gaps] A6's offline proof passes even if the settings action still waits online
where: §18 revision 11 summary, V3 success timing, Л6, and ACCT-6 proof (plan.md:594, :661, :736, :771).
why: V3 promises E2 and its settings action immediately after 204 while the session refresh runs. The current route returns refresh from onSuccess. TanStack Query waits for that callback before marking the mutation successful, except that invalidateQueries resolves immediately for a paused refetch. Keeping the existing callback therefore passes the new offline proof while a slow online /me still holds the page in E1 after confirmation succeeded. Anchors: email.change.confirm.tsx:18–31, identity-queries.ts:35–37, query-core mutation.ts:238–274 and queryClient.ts:320–343.
action: In V3, say to start the existing refresh without returning or awaiting its promise from onSuccess. In ACCT-6, keep the online post-204 /me request unresolved and assert that E2 and its settings action are already visible before releasing it. Retain the separate Л6 cases.
weight: blocking

[logical-gaps] The bare A3 paragraph names a final state before the session read
where: §18 V2 first-match table and the bare-token trigger (plan.md:633, :640–645).
why: The bare-token paragraph calls its local link-failure result S4, while the first-match table includes missing tokens in S3 for a signed-in, verified viewer. The guest proof cannot expose an early return through S4. The table and the later action rule still establish the intended final selection.
action: Call the missing-token result a local link failure. Apply S3 or S4 after the session read resolves, and include a bare URL with a verified session in the route test matrix.
weight: non-blocking

NEXT: Return the two findings and the single-lens verdict. Lens verdict: 1 blocking finding.
