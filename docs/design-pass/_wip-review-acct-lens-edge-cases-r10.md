NEXT: Read the AI Factory entry rules, then use the plan-critique skill to check revision 10 for edge cases.
FACT: Read the shared entry rules. Only this progress file will be modified.
NEXT: Verify the repository root and read the local rules, briefing, and lens contract.
FACT: Verified the courses root and scoped instructions.
DECISION: Apply the edge-cases lens only. The owner’s Л1–Л6 limitations are accepted subject to the stated contradiction and audit-fix exceptions.
NEXT: Load the briefing, fixed plan, audit evidence, and prior edge-cases findings.
FACT: Read the prior edge-cases verdicts. Revision 8 had three blockers and one proof clarification. Revision 9 had three blockers.
NEXT: Read the operative revision-10 requirements and audit answers in bounded ranges.
FACT: V2 now holds a post-attempt session result locally and defines S0. The proof covers offline token mutation and later focus refetch. V5a details and the no-token route still need tracing.
NEXT: Read V4–V10 and audit #1–#16, then trace the route and session entry points.
FACT: V5a now latches confirm uncertainty through later failures, and current A6 awaits the standard invalidation in the mutation success handler.
NEXT: Verify the no-token, login, backend error, and browser-fixture paths against the acceptance promises.
FACT: Verified staleTime 0, shared-key invalidation, the root session loader, and the login branch. Full navigation re-reads a real session cookie.
NEXT: Check whether the browser 401 fixture actually invalidates the cookie and whether V3 can represent a post-success paused read.
FACT: The proposed signed-in ACCT-3 browser fixture stubs only resend, so its real session remains valid on the reload. TanStack invalidation resolves an already-paused refetch immediately.
NEXT: Check service validation, query cancellation boundaries, and unresolved cases before assigning weights.
FACT: Verified the distinct 401 codes and the 2FA partial-write path. V5/V5a, V8, and ACCT-8/9 now cover the prior focus, cancellation, and 401 findings.
FACT: Verified the no-token route skips its mutation. V2 names a local session read only after a token attempt and gives S4 actions only from that read.
NEXT: Cross-check the rest of the audit, including shared control boundaries, then finalize the findings.
FACT: The plan and named source files match the pinned commit. The audit requires all 16 fixes. The no-token path also covers empty or rejected token search values.
NEXT: Compare revision 10 with its parent to classify new findings and verify the final progress record.
DECISION: Revision 8 follow-up closed in the plan. V2 and ACCT-5 handle the paused and failed post-attempt reads. V5a and ACCT-8 preserve focus on rejection. V5a and V8 retain the reload action and hide Cancel after ambiguous confirmation. ACCT-9 names each 401 code separately.
DECISION: Revision 9 follow-up closed in the plan. The local read protects the mounted S4 form from later background refetch; S0 covers a paused token mutation; V5a and ACCT-9 retain uncertainty after a later pre-write 400.
DECISION: Three actionable revision-10 findings remain. The accepted server race and stale in-flight response bounds are outside those findings.
NEXT: Return the fixed-format findings and verdict.

[edge-cases] A3 has no session decision when the token is absent
where: §18 V2 (639–646), ACCT-2/3 and “Лист і посилання” step 2 (741–742, 782).
why: The new local read starts after a token attempt. A missing or empty token starts no mutation (verify-email.tsx:25–27, token-search.ts:4–8). Under V2's exclusive read rule, S4 has no settled session result to choose its resend action from. The no-token guest journey can therefore start without the email form that audit #2 requires.
action: Define where initial no-token S4 gets a session result, including anonymous, signed-in, and unavailable outcomes. Cover a direct tokenless entry with no verify POST and a working guest form.
weight: blocking

[edge-cases] The banner 401 browser fixture leaves the session valid
where: §18 ACCT-3 browser proof (767).
why: Stubbing only the resend response leaves the signed-in browser's valid cookie intact. Full-document navigation reloads root identity, and the login view renders “Продовжити” for that user (root-identity.ts:26–33, login.tsx:44–51, LoginView.tsx:35–42). The planned password-field assertion fails even when V1 is implemented correctly.
action: Use a dedicated browser context whose banner renders while signed in, then invalidate the session while retaining the cached banner and cookie before clicking resend. Assert the real authentication_required response and the password field after full navigation.
weight: blocking

[edge-cases] A6 cannot show the promised post-result offline state through existing refresh
where: §18 V3 (652–659), Л6 (734), ACCT-6 proof (769).
why: If the token POST succeeds just as connectivity drops, A6's onSuccess awaits identityDependencies.refresh, whose invalidateQueries resolves immediately when the session refetch is already paused (email.change.confirm.tsx:18–24, identity-queries.ts:35–37, queryClient.ts:320–343). A6 therefore reaches E2 and offers its settings action while the refresh is paused. Л6 promises a waiting message with no action after the result, and ACCT-6 tests neither offline phase.
action: Give A6 one explicit contract after 204. Scope the post-result wait in Л6 to A3 if A6 should show E2 immediately, or define A6's pause gate if it must wait. Verify E0 before the POST and the selected behavior when the session refetch pauses after 204.
weight: blocking

FACT: One assigned lens completed against 7bd73c9f. Static review only. No make check, Docker, or E2E ran. Only this working file was edited.
lens verdict (3 blocking findings)
