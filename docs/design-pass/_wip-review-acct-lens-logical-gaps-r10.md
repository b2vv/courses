NEXT: Read the AI Factory entry rules, then verify the checkout and load the logical-gaps contract.
FACT: Read the shared entry rules. The owner designated this file as the only write target.
NEXT: Verify the repository root and revision, then read the local rules and lens contract.
FACT: HEAD, main and revision 10 resolve to 7bd73c9f. The local rules and skill are loaded.
DECISION: Run the assigned logical-gaps lens only and use the supplied audit and prior same-lens files.
NEXT: Read the pinned section, prior findings, and audit owner answers.
FACT: Prior findings cover the resend 401 route, ACCT-9 scope, A3 query outcomes and pauses, neutral A6 copy, and mail/secret proof.
NEXT: Read the operative revision 10 rules and identify the changes from revision 9.
FACT: Revision 10 introduces a locally tracked A3 session read and explicit paused-token states for A3 and A6.
NEXT: Cross-check acceptance, proofs, and the audit against the concrete route and session boundaries.
FACT: The r9 observer timestamp is gone. The former success fallback and paused token mutation now have explicit states. ACCT-3 now follows the link in its browser proof.
NEXT: Check whether the new no-token, background-read, and offline paths have the inputs and lifecycle they rely on.
FACT: The normal bare A3 URL performs no token mutation. A6 still returns the existing refresh helper from onSuccess. A resend-only mock does not invalidate the browser session.
NEXT: Verify these discrepancies against the query implementation, the login guard, and the browser fixture.
FACT: The full-document login reads server-side identity, so the ACCT-3 browser fixture must actually make the next document anonymous. The current account fixture already supports logout through its request context.
NEXT: Confirm the exact query and backend branches, and finish the audit-to-plan trace.
FACT: A3 uses staleTime 0 and fetchQuery awaits the fetch and propagates its error. A6 refresh uses active-query invalidation, whose promise resolves immediately for a paused query. The V3 proof omits the newly added E0 branch.
DECISION: The accepted Л1–Л6 cases stay in force. Report only a missing no-token trigger, the A6 offline contract mismatch, and the browser fixture that cannot prove the resend recovery as written.
NEXT: Verify the precise code anchors and write the prior-finding dispositions and final findings.
FACT: Checked the pinned plan and requested route, UI, styles, session, backend, and e2e source against audit #1–#16 and the taste decision. The audit owner approved them all.
DECISION: Revision 8 dispositions. The resend 401 behavior is closed by full-document navigation, with a distinct fixture prerequisite below. ACCT-9 aligns with V5a. Offline A3 result reading is closed by awaiting fetchQuery, with A6 separately affected by the newly broadened Л6 promise. A6 English copy is neutral. The mail and secret proofs now check the actual 410/token_used, the full field value, and both scroll dimensions.
DECISION: Revision 9 dispositions. A failed read after 204 preserves S2 and offers full-document sign-in. The millisecond observer race is gone because the route handles the fetchQuery result directly. S0 covers paused token verification. ACCT-3 adds the real click and form assertion, but the proposed mocked fixture still cannot establish anonymous identity on reload.

[logical-gaps] The tokenless verify page has no trigger for its required session read
where: §18 V2 S3/S4 and session-read rule, V4a, ACCT-2/3 (plan.md:638–646, 662, 741–742, 782).
why: V2 selects the S3/S4 actions only from the local read, and starts the read after a token attempt returns. On /verify-email without a token, the route never starts the mutation. This is the ordinary guest resend entry in ACCT-2 and the stale-cookie entry in ACCT-3. Merely rendering the S4 heading does not establish the local result needed to reveal its form. Sources: verify-email.tsx:25–38 and token-search.ts:4–8.
action: Name a missing or discarded token as an immediate local link-failure outcome that triggers the same one-time session read on mount without a verification POST. Include bare /verify-email with current() resolving null in the route proof.
weight: blocking

[logical-gaps] A6 can show the settings action while its session refresh is paused
where: §18 V3 E0–E2 and refresh rule, Л6, ACCT-6 proof (plan.md:652–659, 734, 769).
why: Л6 now promises waiting text and no action when the post-result session read pauses on A6. V3 keeps the current onSuccess refresh. That helper invalidates active queries, and TanStack Query resolves the helper promise immediately when the refetch is paused. After a 204 the mutation can therefore settle and render the E2 settings action while the read remains paused. The ACCT-6 route proof also skips E0 and the paused post-result read. Sources: email.change.confirm.tsx:18–24, identity-queries.ts:35–37 and installed queryClient.ts:320–343.
action: Keep the accepted offline limit and align V3, Л6, and ACCT-6. Either scope A6 waiting text and no action to E0 and document the existing E2 behavior, or define an awaited A6 read and its explicit post-result paused state. Exercise the agreed behavior before the token request and after 204.
weight: blocking

[logical-gaps] The ACCT-3 browser fixture leaves the session valid on the new document
where: §18 V1, ACCT-3 browser proof (plan.md:624, 742, 767).
why: Mocking only the resend POST to return 401 leaves the signed-in browser's cookie and server session valid. A full document navigation then resolves a signed-in viewer again and LoginView shows Continue instead of the asserted password field. The proposed case would fail with the intended implementation. Sources: identity.ts:10–19, root-identity.ts:21–30, LoginView.tsx:35–43, and the registered signed-in fixture in accounts.spec.ts:248–269.
action: In an isolated context seeded from the existing registered session, render the signed-in banner, replace the session cookie with an invalid value without reloading, and use the real resend 401. Then follow the action and assert the credential field. Preserve the original account flow and zero added logins.
weight: blocking

FACT: Static review only. No make check, Docker, or e2e run. No other lens's working file was opened. The designated working file was the only file edited.
NEXT: Return the single-lens verdict. Lens verdict: 3 blocking findings.
