NEXT: Read the AI Factory entry rules and the edge-cases review contract.
FACT: Read the shared entry rules and verified the checkout root at /Users/strelia/projects/courses.
NEXT: Read scoped rules, the briefing and the named review contract.
FACT: Main is at cf5c5477. The contract limits this lens to boundary data, repetition, concurrency, dependency failure and partial application.
DECISION: Run the assigned lens only and append all review notes to this working file.
NEXT: Read the briefing, review template and numbered §18, then check earlier edge-case findings.
FACT: This is a full slice with plan critique pending. The ordered verification gate belongs to implementation.
NEXT: Read the full behavior tables and extract the previous findings for traceable closure.
FACT: Indexed the earlier edge-case findings. Revision 6 addresses the five revision-5 findings explicitly in its changelog.
NEXT: Verify the operative rules and owner decisions, then check the actual FE and Go paths.
FACT: The operative tables cover remounts, held token attempts, refused requests, cross-surface resends and explicit 2FA reconciliation. A6 prescribes invalidateQueries with cancelRefetch; I am checking whether that primitive guarantees the promised fresh read.
NEXT: Read routes, session adapters and query implementation.
FACT: Query-core Query.fetch conditions cancelRefetch on data already being defined. A6 can therefore join a pre-outcome initial /me instead of replacing it. Default invalidateQueries also filters to active queries.
NEXT: Check root observers, protected navigation, and backend account and 2FA failure ordering.
FACT: requireIdentity uses fetchQuery on the same session key, so the guard can join an old initial /me or surface the generic route error on a failed read. V5a now orders explicit cancellation before publishing the 401 result.
NEXT: Check actual handler status codes, adapter mappings, UI boundaries and the audit examples.
FACT: The audit owner approved all 16 items and the label change. The frontend preserves the relevant 401, 403 and 409 codes. RequestEmailChange issues an additional token without invalidating earlier email-change tokens. This makes late success from an old screen a real shared-state case.
NEXT: Verify late A6 outcomes, password retry semantics, exact 2FA transitions and e2e fixture paths.
FACT: V10’s new password instructions produce a concrete 422 if the user follows them and leaves the submitted new password in the New password field. ValidatePasswordChange rejects identical values before verifying credentials. V3’s shared refresh also needs a defined policy for late results from a superseded A6 screen.
NEXT: Confirm the QueryClient behavior with an isolated in-memory check and finish disposition of every prior finding.
FACT: The installed QueryClient reproduced the stale initial-read case in memory: invalidation with cancelRefetch and then the guard yielded one request and both read and cached the old address.
DECISION: Rev 1 findings in order: closed by neutral 204 copy; stale-cookie login action; dispatch-time shared resend guard; terminal link recovery (A6 freshness remains open below); token-keyed views; immediate 2FA cache update; code-set-scoped copying; standalone button selector; distinct Mailpit ID; TextField/TextArea-only fallback.
DECISION: Rev 2 findings in order: closed by conditional copy; synchronous pending check; explicit A3 cancellation and S0; pending token join; locked 2FA recheck and signed-out exit; fresh proof/admin path after ambiguous regeneration; read-only during writes; content-sizing plus full-value check.
DECISION: Rev 3 findings in order: closed by token-keyed views; documented settled reopen; retained 2FA form plus restored focus; explicit 403 states, navigation and no repeated POST from the refusal screen.
DECISION: Rev 4 findings in order: closed by A→B→A join; same S3 predicate after a real retry; exact-key cancellation before settings 401 publication; four-action 409 reconciliation; manual recheck focus restoration.
DECISION: Rev 5 findings in order: closed by explicit cancel-then-null; enroll included in the 409 rule; still open for an initial /me with no cached data; closed by actual response on settled reopen; closed by neutral 204 copy with no inbox-order inference.
NEXT: Reconcile evidence and verification requirements for the final findings.
FACT: Cross-checked §18 against the identity routes, session code, UI controls, styles, backend, owner audit and accounts.spec.ts. Used one in-memory QueryClient check. No Make, Docker or e2e runs.
DECISION: Return the following findings.

[edge-cases] A6 can reuse a pre-outcome initial session read
where: §18 V3 line 648, ACCT-6 and ACCT-13.
why: Query Core query.ts:397–415 honors cancelRefetch only when state.data is defined. If the first /me is still pending with no cache value, invalidateQueries reuses it and requireIdentity joins the same promise. In an isolated check using the installed package, just one read started and both the guard and cache received the old address.
action: Require explicit cancellation of the exact session key before a post-outcome fetch, including when there is no cached value or observer. Add an ACCT-6 case with a held initial /me and assert that only the post-outcome snapshot reaches settings.
weight: blocking

[edge-cases] A late A6 write can leave the settings address stale
where: §18 V3–V4a lines 648–652, ACCT-7 and ACCT-13.
why: V4a ignores results after the keyed screen unmounts, while V3 requires a shared session refresh after every A6 outcome. Two email-change tokens for the same account can remain valid because accounts.go:225–265 issues tokens independently. A can still commit after B settles and the viewer opens settings. Discarding A's completion leaves B's address cached. Unconditionally cancelling the shared query from A's old screen can instead reject an in-progress settings guard, which maps that rejection to dependency_unavailable.
action: Specify which owner handles shared refresh when an A6 screen unmounts, independently of the token-specific result displayed. Cover A → B → settings with A committing last, including a held settings guard. Require the guard to complete and the address to converge to the final committed value.
weight: blocking

[edge-cases] Password recovery can create an identical-password submission
where: §18 V10 line 689 and ACCT-9/ACCT-14.
why: After an ambiguous committed change, the suggested retry fails with invalid_credentials. Following the next instruction puts the previous new password into Current password while New password still contains it. ValidatePasswordChange rejects that pair as password_unchanged before checking credentials (account_policy.go:84–89, accounts.go:195–210), so the suggested step cannot establish which password is current.
action: Give separate instructions for checking whether the submitted password works and for making a further change. If the user still wants a further change, tell them to choose a different new password. Verify the full ambiguous-commit recovery sequence.
weight: blocking

DECISION: lens verdict (3 blocking findings).
