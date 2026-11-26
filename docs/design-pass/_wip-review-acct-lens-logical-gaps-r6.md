NEXT: Read the AI Factory entry rules, then check the prior logical-gap findings against revision 6.
FACT: Read the AI Factory entry rules.
NEXT: Verify the repository root and read local routing, briefing, and the lens contract.
FACT: HEAD is cf5c5477. Read the local route and logical-gaps contract.
DECISION: Run the one requested lens with the working file as its only write.
NEXT: Read the revision 6 plan, audit decisions, and historical findings.
FACT: Verified main at the requested commit and located §18.
NEXT: Compare every historical finding with the written plan and audit.
FACT: Read the owner’s fix-all decision and the revision history.
NEXT: Read operative plan clauses and prior findings in bounded excerpts.
FACT: Revision 6 names all four revision-5 issues in its operative clauses. Earlier issues still need to be traced through the proof rows and implementation.
NEXT: Read the exact revision diff, verification matrix, and route implementations.
FACT: Proof rows cover the prior label check, held /me focus checks, in-flight token reuse, and the 403 no-retry actions.
NEXT: Trace session fetching, invalidation, and backend outcome ordering.
FACT: Settings uses fetchQuery with staleTime 0. The verification route already refreshes the shared session on success. Confirm and disable consume a factor before updating the enabled flag.
NEXT: Check the query library behavior, adapter error mappings, views, CSS, and e2e fixtures.
FACT: In the installed TanStack Query core, cancelRefetch only cancels an in-flight request if cached data exists. Revision 6 specifically relies on it for A6.
NEXT: Verify password recovery prerequisites, A3 success refresh, and query-core cancellation semantics.
FACT: The backend validates current/new password inequality before checking credentials. V10’s new instruction leaves the person typing the same value into both fields unless they infer an extra step.
FACT: A3 success currently awaits identityDependencies.refresh, while the revision says S2 needs no read and derives its action from the root cache.
NEXT: Resolve whether refresh is retained elsewhere and finish the historical closure ledger.
FACT: Reviewed the named routes, identity views, shared controls and CSS, session application/domain/adapter, backend handler and account/2FA use cases, accounts e2e, and the installed query-core behavior. The owner approved fixes for all 16 audit findings and the A6 label.

DECISION: Revision 1 (8/8), in original order. Closed by (1) conditional V1 acknowledgment and separate Mailpit proof, (2) S0 and the stale-cookie 401 path, (3) ordered S3/S4 rows and message-ID/guest-first proof, (4) terminal E3 and S4/E4 recovery after a used token, (5) V5a success state and held-read focus proof, (6) V11 direct grid-child and recovery controls with separate measurements, (7) V6 text-control scope, (8) ACCT-20 view proof.
DECISION: Revision 2 (6/6), in original order. Closed by (1) conditional copy that avoids delivery/order promises, (2) S0 with an observed screen-owned session read, (3) S3 scoped to the current account and S5 for ambiguous failures, (4) explicit replacement of setup with an enabled card and recovery notice, (5) test-controlled /me release, (6) E5 → E4 → settings proof and safe return-to. The revision-6 implementation prerequisite for a fresh A6 read is a new finding below.
DECISION: Revision 3 (4/4), in original order. Closed by (1) fresh tokenless entry and a stale-null recovery test, (2) S5 retry and S3 advice to request a new link after sign-out, (3) in-flight-only retention and actual responses on completed-attempt re-entry, (4) mounted controls while checking and explicit manual-recheck focus.
DECISION: Revision 4 (3/3), in original order. Closed by (1) a common S3 rule for initial attempt and retry, (2) a focus destination after the manual recheck, (3) session cancellation before publishing null and common 409 reconciliation for all four 2FA actions.
DECISION: Revision 5 (4/4), in original order. Closed by (1) explicit cancelQueries before null plus held-read proof, (2) the enrollment 409 branch and focus proof, (3) new-instance tests for both spent and usable tokens, (4) distinct S5/E5 navigation and no repeat on origin_forbidden.

[logical-gaps] A6 invalidation can reuse a pre-result session read
where: §18 V3 and ACCT-6/13 (plan.md:648, 717, 734)
why: The installed query-core query.ts:401–415 only cancels on cancelRefetch when state.data is already defined. An initial /me without cached data can be pending when A6 settles, for example after a failed initial identity read followed by a recovery fetch. The specified invalidation joins that older promise. The settings guard also uses fetchQuery on the same key (require-identity.ts:11–20), so it can join the older read and render the old address. The current proof can pass with a seeded cache while missing this path.
action: Require and await cancellation of the exact session query before starting the post-result fetch, and order publication of the settings action so its guard can only join the new read. Extend ACCT-6/13 with a held pre-result /me and no cached data.
weight: blocking

[logical-gaps] A3 success leaves the existing session refresh unspecified
where: §18 V2 S2 and ACCT-4 (plan.md:630, 636, 708, 749)
why: The existing A3 success callback awaits refresh (verify-email.tsx:17–20). V2 says S2 needs no session read, and its proof stops at the CTA. If that rule drops the refresh, the root query can keep emailVerified=false. The banner observes that query (__root.tsx:176, 389–394) and the lesson gate uses the root identity snapshot (root-identity.ts:26–30, learn.$.tsx:182, 205). An immediate client-side transition can carry stale verification state. The existing accounts spec verifies the banner and gate after full page navigations (accounts.spec.ts:283–290).
action: Keep the post-success shared session refresh explicit and specify its ordering relative to the learning CTA. Verify the signed-in S2 link through a router transition and assert the banner and lesson gate clear without a page reload.
weight: blocking

[logical-gaps] The password recovery instruction leads to password_unchanged
where: §18 V10 and ACCT-9/14 (plan.md:689, 737, 740)
why: The new instruction says to enter the intended new password as the current password after a retry rejects the old one. The new-password field still contains that same value. The next submission fails before credentials are checked (accounts.go:195–209 and account_policy.go:84–91). Also, the ambiguous response could have occurred after saving the password but before revoking other sessions (accounts.go:219–222).
action: Tell the user how to check whether the intended password now works, for example by signing in with it in a separate session. Explain how to proceed if another password change is needed, with a different new password. Check the recovery sequence after a simulated committed update with a lost response. Update the existing email-request success assertion in accounts.spec.ts:337 for V10’s new copy.
weight: blocking

FACT: Static review at main cf5c5477. No make check, Docker, or e2e commands ran. Only this working file was written.
DECISION: Single requested lens returned, 1/1. Prior findings closed individually. The A6 session-refresh and password-recovery additions introduce two blockers, and the A3 success handoff exposes one remaining regression risk.
lens verdict (3 blocking findings)
