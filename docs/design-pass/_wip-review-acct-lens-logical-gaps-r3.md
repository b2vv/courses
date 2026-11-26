NEXT: Read the shared entry rules and the plan-critique contract, then review only the logical-gaps lens.
FACT: Verified repository root and read shared entry rules, local routing, briefing and review contract.
DECISION: Review only logical gaps in the requested revision. Use static source inspection.
NEXT: Read the pinned plan section, owner evidence and this lens’s two earlier reviews.
FACT: efc0bcf9 is the current main commit. Owner approved all 16 audit items and the title naming issue.
NEXT: Compare each earlier finding against the numbered revision 3 decisions and proof rows.
FACT: Revision 3 closes the earlier 204, session-unavailable, 2FA partial-enable transition, focus-check timing and A6 recovery-path issues on paper. S3 wording now scopes the account; its control flow still needs checking.
NEXT: Trace route guards, session query options, token mutations and 2FA transitions in source.
FACT: The session query uses staleTime: 0, so cancel then fetch can establish a fresh read. Token actions use the token account independently of the viewer. Account settings uses fetchQuery at entry and redirects a null session.
NEXT: Verify the lifetime of the token attempt cache, how sign-out navigates, and whether the proposed 2FA focus targets survive reconciliation.
FACT: Current sign-out navigates within the router and clears only the session query. TanStack mutation entries have finite default lifetime. V5a replaces the focused controls during reconciliation even when the subsequent session read restores the same form.
NEXT: Check the pinned source lines and reconcile all earlier findings against the remaining plan gaps.
FACT: Verified a five-minute client default for unobserved settled mutation records and that the router sets no mutation-specific retention.
NEXT: Check how root identity can bypass a fresh read, then record findings and the revision-by-revision dispositions.
DECISION: Revision 1 dispositions, in file order. (1) Closed by V1 conditional 204 copy and ACCT-2 Mailpit proof. (2) The stale-cookie 401 path is closed. The unavailable-session path still needs a current identity observation, as described below. (3) Closed by ordered V2 rows and the message-ID and guest-context steps. (4) Closed as a recovery design by S4, E3 and E4 plus the E5-to-settings proof. Clarify that S4 after token_used is the unverified/anonymous branch and S3 still takes precedence when the current account is verified. (5) Closed for successful transitions by V5a and ACCT-8. The new reconciliation focus gap is separate. (6) Closed by V11 and ACCT-15. (7) Closed by the explicit V6 TextField/TextArea scope. (8) Closed by the ACCT-20 verification row.

DECISION: Revision 2 dispositions, in file order. (1) Closed by conditional 204 copy. (2) S0 exists and has a recovery action. Its observation prerequisite remains open. (3) S3 no longer claims which account the token belongs to. The stated recovery still conflicts with V4a on client-side reopening. (4) Closed for the partially applied enable branch by explicit cleanup, rendering and focus, with corresponding ACCT-9 assertions. (5) Closed by the test-controlled /me release. (6) Closed by the E5 → E4 → settings proof and documented returnTo behavior, backed by requireIdentity. The guest return itself has no separate browser assertion.

[logical-gaps] S0 needs a current session outcome
where: §18 V2 S0/S4 and ACCT-3 (plan.md:598–609, 701).
why: The root resolver uses ensureQueryData (root-identity.ts:26–32). It can return previously cached null or user data even after a background /me read fails. The root observer disables refetch on mount (__root.tsx:176), and V2 explicitly requires freshness only after an attempt. Entering A3 with no token can therefore pick guest S4 from cached null without detecting the unreadable session. A valid cookie can then cause resend to ignore the address the user typed (handler.go:357–373). ACCT-3 injects unavailable directly and misses this path.
action: Name the current session result A3 must establish before exposing S4, including entry without a token and handling of pending/error status with previously cached data. Verify a cached anonymous snapshot followed by a failed /me with a valid cookie, then recover to the correct branch before allowing resend.
weight: blocking

[logical-gaps] S3 recovery can reopen the same cached failure
where: §18 V2 S3/S5 and V4a (plan.md:605–609, 623–624), ACCT-5/7.
why: For an ambiguous failure on another account's still-usable link, a verified viewer gets S3 and loses S5's explicit retry. S3 advises signing out and reopening. Sign-out navigates within the same router and only clears the session query (use-sign-out.ts:21–27), so reopening by browser Back restores the cached failed attempt. The plan leaves it unspecified whether re-evaluation exposes S5 and retry after a viewer change. A read of the signed-in viewer cannot resolve the other account's token (accounts.go:95–106).
action: Explicitly define how the user reaches a usable recovery action when the same token is reopened after sign-out in the same router. Preserve the pending-attempt deduplication. Prove that a verified viewer with an ambiguous response for another account can follow the recovery through to an actionable guest state.
weight: blocking

[logical-gaps] The token cache lacks the retention required by its automatic-send rule
where: §18 V4a and ACCT-7 (plan.md:624, 680, 704).
why: The router supplies no mutation gcTime (router.tsx:35–37). Installed TanStack Query removes unobserved settled mutations after five minutes by default (query-core src/removable.ts:14–28, mutation.ts:143–161). After leaving the route and returning after collection, the absent record triggers a second automatic POST and a successful confirmation can turn into a used-link error within the same app instance.
action: Define and prove the retention horizon and the behavior on re-entry after it, consistent with the stated guarantee. Clarify that the browser proofs expecting a fresh used-link response start a fresh app instance if same-instance revisits retain the prior result.
weight: blocking

[logical-gaps] Reconciliation removes focus even when it restores the same 2FA form
where: §18 V5, V5a and ACCT-9 (plan.md:628–644, 706).
why: V5a hides the controls during its checking phase, so the focused field or submit button disappears. When confirm fails and /me returns disabled, the plan restores setup but says focus does not move. When disable fails and /me returns enabled, the plan restores the disable group with no focus target. Neither can preserve the old focus simply by rendering the form again, and the intermediate checking phase also has no target.
action: Define where focus goes while checking and after returning to the same form, including after a manual session recheck. Verify focus on confirm → error → still disabled and disable → error → still enabled.
weight: blocking

FACT: Static inspection covered the requested plan, evidence, routes, views, controls/CSS, session layers, backend actions and accounts browser spec. No make check, Docker or e2e ran. Only this working file was edited.
NEXT: None. 1 lens assigned / 1 returned.
lens verdict: 4 blocking findings
