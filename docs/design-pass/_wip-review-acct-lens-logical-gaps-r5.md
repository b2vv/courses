NEXT: Read the shared entry rules, then verify the repository root and review contract.
FACT: Shared entry rules read. Git confirms the Courses root.
NEXT: Read repository routes and the requested review contract.
FACT: Review target confirmed at main 6a8d1a3e. Local rules and the single-lens contract are in scope.
DECISION: Assess historical findings and newly introduced dependencies within logical gaps only.
NEXT: Read revision 5, prior findings, and the audit decisions.
FACT: Audit authorizes all 16 fixes and the naming change. Revision 5 identifies three logical gaps from revision 4.
NEXT: Extract the exact decision rules and all prior lens findings for closure checks.
FACT: The earlier lens files identify recurring dependencies in A3 state selection and 2FA reconciliation. Revision 5 adds explicit rules for each.
NEXT: Check the acceptance rows and exact revision delta against the named frontend and backend sources.
FACT: ACCT-1–20 and the Mailpit sequence are present. Revision 5 adds same-token in-flight reuse, a 401 cache write, 409 reconciliation and a manual-recheck focus target.
NEXT: Trace route mounting, session cache behavior and backend response ordering.
FACT: Account Settings observes the shared session query. The backend can return twofactor_already_enabled from both enrollment and confirmation.
NEXT: Check API error mapping, current-user refresh and which operations the written plan actually reconciles.
FACT: Frontend wire mapping already preserves the two 409 codes and distinguishes session 401 from a rejected recovery code. V5a names confirm, disable and regenerate, but omits enrollment 409.
NEXT: Check the real origin predicate, stale-query overwrite, and route result lifetime before finalizing findings.
FACT: The same-origin guard checks only the submitted Origin values and configured application origins. A same-URL reload cannot repair a stable mismatch. A6 uses the same guard.
FACT: Invalidation and token consumption occur before the subsequent write, with several pre-consumption failure paths. A completed request does not prove a spent token.
NEXT: Verify the session-cache ordering and classify prior findings one by one.
DECISION: R1 (8 items): conditional 204 acknowledgment, stale-cookie and unavailable-session recovery, heading/fixture ordering, used-token recovery, successful 2FA focus, individual button widths, localized text-field scope and the A6 label proof are covered by V1–V6/V11 and ACCT-2/3/5/6/8/10/15/20.
DECISION: R2 (6 items): neutral 204 copy, unavailable-session gating, S3 account scope, partially applied 2FA exit, controlled /me hold and the E5-to-settings sequence are covered.
DECISION: R3 (4 items): the screen-owned session read, removal of the settled-token cache trap, removal of GC-dependent retention and in-place 2FA checking are covered. The new completed-request wording still overstates the used-token outcome.
DECISION: R4 items 1–2 are addressed by unified S3 precedence and the manual recheck focus rule. Item 3 is partial: the cache update establishes null momentarily, and the 409 cases named in R4 are covered, but the new steps omit a pending session read and enrollment 409.
NEXT: Cross-check the final findings against locale, origin and route wiring.
FACT: The installed QueryClient implements setQueryData with a manual cache write and leaves an active fetch running. A successful fetch later calls setData. The redirect hook checks session.data === null in an effect.
FACT: The origin guard checks the single Origin against config.publicOrigin. Both confirmation endpoints go through it. The E5 retry text promises a same-token retry but does not explicitly inherit S5's full-reload action.
DECISION: Record four open findings from static inspection. The 401 and enrollment 409 findings are remaining parts of the revision 4 rejection finding. The completed-token and origin findings concern rules introduced or generalized in revision 5.

[logical-gaps] A pending session read can undo the expired-session redirect
where: §18 V5a and ACCT-9 (plan.md:665, 728).
why: setQueryData(sessionQueryKey, null) leaves a pending /me alive. A read that captured a valid user before revocation can resolve after the POST's authentication_required and overwrite null before useRedirectWhenSessionEnds sees it (identity-queries.ts:25–37, require-identity.ts:30–50, installed query-core queryClient.ts:209–211 and query.ts:575, 656–669). ACCT-9 covers cached identity without controlling a pending read.
action: Retire the older session read before publishing null, and verify the login redirect with a held pre-revocation /me response that is released after the POST's 401.
weight: blocking

[logical-gaps] Enrollment can report the same stale 2FA state as confirmation
where: §18 V5a and ACCT-9 (plan.md:664, 728).
why: EnrollTwoFactor checks Enabled and returns twofactor_already_enabled (two_factor.go:58–65, handler.go:201–212). With a cached disabled state after another window enabled 2FA, pressing «Увімкнути 2FA» returns 409. The plan reconciles that code only on Confirm, so the stale Enable button remains the user's offered action.
action: Include enrollment in the 409 reconciliation rule. Name the focus behavior while its button stays mounted and when the enabled card replaces it. Add the stale-disabled enrollment case to ACCT-9.
weight: blocking

[logical-gaps] A completed request does not establish a spent token
where: §18 V4a and ACCT-7 (plan.md:642–643, 702, 726).
why: V4a evicts every settled promise, then promises token_used on any completed-token revisit. A3 and A6 can return origin_forbidden before calling the handler, and a client failure can occur before the request reaches the server. The next POST can consume a still-valid token and succeed. Treating any completed attempt as a used link contradicts the response-driven tables.
action: Say that a newly mounted screen handles the actual new POST response. Restrict the token_used expectation to a previously consumed or invalid token and verify reopening after both a successful write and a failure before consumption.
weight: blocking

[logical-gaps] Reload does not supply the missing origin prerequisite
where: §18 V2 S5, V3 E5 and ACCT-5/6 (plan.md:625, 637, 724–725).
why: Both token POSTs must supply the configured application origin. The guard checks the Origin header against config.publicOrigin (handler.go:115–121, 514–523, config.go:61–63). Reloading the same URL leaves a stable origin mismatch intact, so S5 can send the user around the same loop. E5 also leaves it unclear whether its 403 retry inherits S5's full reload or reposts in place. The proof checks the reload action alone.
action: Define the same 403 recovery contract for both screens. Name any condition a reload can repair, and give a persistent 403 a recovery path that exits the loop. Verify a persistent rejection separately from a transient response.
weight: blocking

FACT: Reviewed main@6a8d1a3e, audit decisions, four historical logical-gaps files and the named code. No gates, Docker or e2e ran. Only this working file was written.
DECISION: Single requested lens returned, 1/1. Review complete.
lens verdict (4 blocking findings)
