NEXT: Read the shared entry rules, verify the repository root, and load the logical-gaps contract.
FACT: Read the shared entry rule and verified /Users/strelia/projects/courses as the repository root.
NEXT: Read the local rules, briefing, plan-critique contract, and revision pointers.
FACT: Confirmed main is d9136f15. Read the local project rules and the logical-gaps contract.
NEXT: Compare §18 with the r8 findings and the source audit.
FACT: The audit owner approved all 16 findings and the taste change. Revision 9 records six accepted limitations.
NEXT: Read the exact rules, checkpoints, and all r8 findings without truncation.
FACT: r9 adds full-document navigation for resend 401, narrows ACCT-9 recovery and redirect, reads the A3 observer, and neutralizes E2 English copy.
NEXT: Read the acceptance/proof rows and the concrete route/session wiring.
FACT: The r8 stale-cookie branch, ACCT-9 scope, English E2 copy, and mail/secret proof edits are present. The prior offline issue now depends on the actual observer lifecycle.
NEXT: Trace token errors, account fixtures, and UI/composition prerequisites.
FACT: The real mail flow invalidates earlier verification tokens before issuing the replacement, and verification maps used tokens to 410. The browser fixture has a registered session and leaves verification until after its pre-verification assertions.
NEXT: Review the exact r9 delta, query/mutation behavior, validation APIs, and navigation implementation.
FACT: The application has a working reloadDocument bridge. The verification observer uses staleTime 0, and refresh invalidates the exact active session key.
NEXT: Check A3 no-token and success recovery, 2FA pending focus, and CSS prerequisites against the source.
FACT: S2 still says unreadable session leads to sign-in, but the new observer rule overrides it with a same-URL reload. A successful verification consumes its token before that reload, so a guest can lose the success-to-sign-in path.
NEXT: Verify query error counters and refine the A3 findings and proof request.
FACT: Query records both a millisecond error timestamp and an incrementing errorUpdateCount. The V2 strict timestamp comparison can miss a failure recorded at the start timestamp and then fall through to cached identity.
NEXT: Finish the r8 dispositions and isolate the decision-level issues supported by the r9 text.
FACT: Л6 now distinguishes paused queries in the A3 observer, but S1 still treats a paused verification mutation as an in-progress check without the offline text Л6 promises.
NEXT: Check application Query defaults and the pinned button behavior before assigning weights.
FACT: Production Query disables retries. A first-attempt `/me` failure can use the same millisecond as refresh dispatch. The pinned React Aria Button keeps the real button focusable for `isPending` and suppresses its interactions.
DECISION: r8 dispositions: (1) resend 401 behavior closed by full-document navigation, with an end-to-end proof gap left below; (2) ACCT-9 closed by matching V5a's named codes, actions, and settled-session precondition; (3) the paused-refresh-promise issue closed by observing fetchStatus, with a separate paused-mutation copy gap below; (4) English E2 closed by neutral text; (5) mail and secret oracles closed by the real 410/token_used, full value, and both scroll dimensions.
DECISION: The two blocking findings concern A3 result selection. The accepted limitations remain in force. The narrower offline finding concerns only a mismatch in the proposed messaging.

[logical-gaps] A failed session read turns successful verification into a used-link screen
where: §18 V2 S2 and observer fallback, V4a, ACCT-4/5 (plan.md:634, 639–642, 657, 738–739).
why: S2 promises sign-in if the session is unreadable, but the new observer override offers only a full reload. VerifyEmail consumes the token before returning 204. A guest who reloads the same URL therefore gets 410/token_used and S4, sending them back to resend after successful verification. This loses audit #3's continuation even if the next session read works. Sources are accounts.go:95–106 and handler.go:841–842.
action: Preserve the 204 success state and its full-document sign-in fallback when the session read fails. Limit same-page reload to the failure branch that needs session-dependent recovery. Add a route case for 204 followed by `/me` failure.
weight: blocking

[logical-gaps] The observer can miss a failed session refresh and reuse cached identity
where: §18 V2 observer fallback and ACCT-5 (plan.md:639–642, 739, 763).
why: The chosen test, `errorUpdatedAt > refreshStartedAt`, can be false when a failure is recorded during the same millisecond as dispatch. Query uses `Date.now()`, retains cached data on a refetch error, and the app sets `retry: false`. The plan then selects S3 or S4 from cached identity even though ACCT-5 requires the failure action. Sources are Query `query.ts:672–685` and frontend/src/router.tsx:35–36.
action: Identify the refresh error by observer state without using wall-clock ordering. For example, snapshot `errorUpdateCount` before invalidating, then check the settled observer's error state and whether the count advanced. Cover an immediate failed refresh with cached identity in ACCT-5.
weight: blocking

[logical-gaps] A3 offline text covers a paused session read but omits paused verification
where: §18 V2 S1 and observer rules, Л6 (plan.md:633, 639–640, 729).
why: A verification mutation started offline is pending with `isPaused` before the server ever sees the token. S1 wins, and the observer rule introduced by r9 runs only after an attempt result. The user sees “Перевіряємо посилання…” while Л6 promises “Чекаємо на з'єднання…”. The existing proof covers a paused `/me` only. Source is Query `mutation.ts:209–216`.
action: Put an explicit paused verification branch before S1, or narrow Л6's A3 copy promise to a paused session read and state the copy during a queued verification. Keep the accepted pause behavior.
weight: non-blocking

[logical-gaps] ACCT-3 checks the navigation flag without proving the sign-in form opens
where: §18 ACCT-3 and its proof (plan.md:737, 762).
why: The new route assertion checks `reloadDocument` on the link, and the browser proof stops at the 401 alert in an already-guest context. Neither follows the recovery action from cached signed-in identity to credentials, which is the branch that previously failed. The existing React Aria router bridge supports the flag, so the planned behavior is implementable.
action: Add a browser assertion that clicks the alert's sign-in link from a surface with cached signed-in identity after a 401 and sees the credential form.
weight: non-blocking

FACT: Static review of §18 at d9136f15 and the requested sources. Only this working file was edited. No other lens file, make check, Docker, or e2e was used.
NEXT: Return the logical-gaps verdict. 1 requested lens / 1 returned. Verdict: 2 blocking findings.
