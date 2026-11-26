NEXT: Read the shared entry rules, then verify the repository root and load the logical-gaps contract.
FACT: Shared entry rules loaded. The requested working file is the only write target.
FACT: Verified main at ab3cae05 and loaded AGENTS.local.md, the plan-critique contract, and the task pointer.
NEXT: Read §18, the audit and owner decisions, and the six prior logical-gaps reviews.
FACT: Read the active ACCT plan, audited findings and owner authorization, and prior logical-gaps dispositions.
NEXT: Trace the listed routes, views, session plumbing, backend outcomes, and regression scenarios.
FACT: Traced the scoped frontend and backend sources and the existing account e2e journeys. The revision switches to a wait-then-read helper and gives token continuations ownership of session refresh.
NEXT: Check the query semantics, verification-state consumers, and 2FA state ownership against the revised promises.
FACT: Session queries use staleTime 0. The root banner observes the shared query, the mounted lesson gate reads a route-context snapshot, the login view switches to Continue for an already signed-in session, and the existing twoFactorPending is a render-derived boolean.
NEXT: Check the promised proof paths, decide which mismatches are blocking, and record each prior finding’s disposition.
FACT: /learn is the learning list. Its render cannot prove the absence of a gate mounted only on a concrete lesson. V3 maps the tokenless route to E4, and V4a provides no POST promise for that route.
NEXT: Append the disposition ledger, fixed-format findings, and verdict.

DECISION: R1, in finding order. (1) Closed: V1 acknowledges the request and ACCT-2 proves actual delivery separately. (2) Closed: S0 and the resend 401 path distinguish session-read failure and stale cookies. (3) Closed: V2 orders the headings and the e2e journey distinguishes message IDs, guest rejection and signed-in S3. (4) The spent-token result and recovery destinations are specified in S3/S4 and E3/E4; the tokenless E4 implementation path still needs a source for its session read. (5) Still partial: V5a publishes the mutation result immediately and plans a held-read focus check, but an already-running read can overwrite it before the new read finishes. (6) Closed: V11 includes the direct grid-child Enable control and ACCT-15 measures each audited button. (7) Closed: V6 scopes the default to TextField/TextArea. (8) Closed: ACCT-20 has a view check for the A6 label.
DECISION: R2, in finding order. (1) Closed: conditional 204 copy makes no delivery or link-order promise. (2) Closed: failed screen-owned session read reaches S0 even with cached null. (3) Closed: S3 is explicitly about the viewer’s account after a terminal link error; ambiguity remains S5. (4) Closed: partial confirmation moves to the enabled card with a recovery-code notice. (5) Closed: the browser controls /me release and checks focus on both sides; the prior-read sequence still needs proof under the new V4b rule. (6) Closed for the tested path: ACCT-6 includes E5 → E4 → settings and the actual address.
DECISION: R3, in finding order. (1) Closed: the tokenless A3 check starts a screen-owned read and exercises cached-null recovery. (2) Closed: S3 tells the viewer to request a fresh link; a remount only reuses an in-flight request. (3) Closed: the declared map lifetime matches its tests for concurrent and completed attempts. (4) Closed: reconciliation keeps the form mounted and the manual recheck names a final focus target.
DECISION: R4, in finding order. (1) Closed for retry precedence: S5 sends the same token and reuses the ordered S3/S4 rules. The new V4c read policy conflicts with S5 separately. (2) Closed: after the manual recheck the restored form receives focus when the retry control leaves. (3) Closed: V5a orders null after the older read and explicitly reconciles both state conflicts for every 2FA write.
DECISION: R5, in finding order. (1) Closed by wait-before-null and the held-old-read regression proof. (2) Closed: enrollment is now in the 409 and focus rule. (3) Closed: ACCT-7 expects the actual response on remount and covers both spent and still-usable links. (4) Closed: 403 on both routes exits through navigation rather than offering the same request again.
DECISION: R6, in finding order. (1) Closed for A6 POST outcomes: wait-then-fetch with staleTime 0 handles the initial read with no data, and the held-read proof reaches settings. (2) The missing A3 success refresh is now explicit and ordered before the CTA. The proof still lacks a concrete lesson and the introduced shared continuation has an inconsistent A3 error policy. (3) The exact password_unchanged loop is removed and the old email-request assertion is listed for update. The replacement verification path needs a cookie-free session, and the planned retry assertion needs an explicit state rule.

[logical-gaps] V4c has conflicting rules for which A3 outcomes refresh identity
where: docs/design-pass/plan.md:643, 661, 744 (V2, V4c, ACCT-4/5)
why: V4c places a refresh after any A3 write completion in the shared promise continuation. V2 and ACCT-5 explicitly require S5 for origin_forbidden without a session read, and V2 says an S5 retry does not read the session before the next POST. A continuation that wraps every outcome in a read also lets an unrelated /me delay when that shared attempt settles.
action: Specify one trigger table. For A3, share the read for S2 and terminal link outcomes that need S3/S4 classification, give tokenless entry its own read, and leave S5 outcomes independent of /me. Keep post-result reads for A6 as specified. Verify that a held /me cannot delay S5 or its permitted retry.
weight: blocking

[logical-gaps] A tokenless A6 screen has no event that releases its settings action
where: docs/design-pass/plan.md:652–661, 745 (V3 E4, V4a/V4c, ACCT-6)
why: E4 includes the missing-token case and its settings CTA waits for the post-result session read. V4c places that read in the token POST promise, which is absent on a tokenless URL (email.change.confirm.tsx checks the token before mutating). The old missing-token screen is covered by the owner's fix-all decision.
action: Define entry to tokenless A6 as a local E4 result followed by its own session read with no POST. Test that its settings CTA appears after either session-read outcome and a guest gets the existing login return path.
weight: blocking

[logical-gaps] The learning handoff proof never visits the gated lesson
where: docs/design-pass/plan.md:637, 719, 744, 760 (S2, ACCT-4 proof, Letter and links step 4)
why: The CTA opens /learn, which renders the membership list (routes/learn.tsx:32–43). The gate is mounted only within a lesson and takes its decision from a route-context identity snapshot (routes/learn.$.tsx:182, 205). Gate absence on /learn is true before verification as well. The current accounts spec checks a real lesson before and after verification but uses full page loads, so it does not establish the new SPA handoff.
action: Use a member with a concrete gated lesson, observe its gate before verification, follow the S2 CTA and then enter that same lesson via router links in the same document. Assert the real lesson renders and both the global banner and the lesson gate clear. Hold the post-verification /me response to prove CTA ordering.
weight: blocking

[logical-gaps] A pre-write session read can overwrite the immediate 2FA success state
where: docs/design-pass/plan.md:660, 675, 723, 747 (V4b, V5a, ACCT-8)
why: The current card reads twoFactorEnabled from the shared query (settings.account.tsx:76). V5a establishes the immediate result with setQueryData, while V4b deliberately lets a read started before the write finish. That response can publish the old flag after the patch and before the next read, replacing the card and its focus target. Holding the new /me alone does not exercise the stale publication.
action: Specify how the mutation-owned card state remains authoritative through completion of an older read and until fresh reconciliation. Hold a read started before confirm or disable, complete the POST, release the old response, then hold the post-result response. Assert the displayed state and focus throughout.
weight: blocking

[logical-gaps] The existing 2FA pending flag cannot reject two actions before render
where: docs/design-pass/plan.md:688, 748 (one 2FA action at a time, ACCT-9)
why: twoFactorPending is computed from the four useMutation snapshots and handed to the view (settings.account.tsx:63–64, 76–85). Two handlers in the same render can both read false. The plan cites V1, whose event-time guard reads the mutation cache directly, but tells 2FA to use the existing render-derived flag.
action: Give all four handlers one event-time guard acquired synchronously before dispatch, and keep it effective during verification so the existing two-actions-before-render proof covers both intervals.
weight: blocking

[logical-gaps] A second ordinary window does not let a signed-in user test the new password
where: docs/design-pass/plan.md:700, 748 (V10, ACCT-9)
why: Opening a normal window in the same browser profile carries the existing session cookie. The password-change path preserves the current session (accounts.go:219–222), and /login renders only Continue for a signed-in viewer (routes/login.tsx:44–51, LoginView.tsx:35–42). ACCT-9 also expects the recovery hint to survive invalid_credentials on a retry, while V10 only defines it for the uncertain error; the current route exposes only the latest mutation code.
action: Name a private window or other cookie-free session explicitly. Specify that the hint survives invalid_credentials when retrying the same indeterminate change, while ordinary invalid_credentials keeps its usual message. Check the path with a simulated committed password change and a lost response.
weight: blocking

FACT: Static review of main at ab3cae05. No make check, Docker, or e2e ran. The only write was this working file.
DECISION: Requested lens returned, 1/1.
lens verdict (6 blocking findings)
