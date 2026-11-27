NEXT: Read the shared entry rules and verify the repository root.
FACT: Verified repository root and main branch. The shared entry rules allow the requested review to use its assigned working file.
NEXT: Read local rules, briefing and the named plan-defense contract.
FACT: HEAD is 4eae6978 on main. Local rules leave repository sources authoritative. The contract requires five line-backed answers.
DECISION: Record the defense table in the working file under the owner’s read-only instruction.
NEXT: Read §18, the revision 10 defense, and the source audit.
FACT: The audit defines 16 fixes plus A6 wording. Revision 10 ended with one dependency-failure gap.
NEXT: Compare the exact revision delta and read the token, session and limitation clauses with line numbers.
FACT: Revision 11 explicitly starts the no-token A3 read at mount, narrows post-result offline waiting to A3, and calls for invalidating the real banner cookie.
NEXT: Trace A3/A6, session guards, and the resend endpoints through source.
FACT: The backend chooses the cookie before the fallback email, and the frontend maps /me 401 to null. Existing A6 waits on refresh in the mutation, so V3 requires moving the completion boundary to the POST response.
NEXT: Check the exact no-token path, fresh-document login, test sequencing, and session transport.
FACT: The token parser drops empty and non-string values. The existing e2e spec registers a signed-in, unverified account before consuming its link. Its session can be copied and replaced in a separate context at that point.
NEXT: Read UI, CSS, 2FA, and transport seams for unanswered rows or audit fixes the revised plan cannot deliver.
FACT: ACCT-3 now exercises the real 401 and fresh-document login. V3 and Л6 agree that E2 appears on the 204 while its session refresh proceeds separately.
QUESTION: Does naming S4 as the immediate no-token result disable the S3 branch for verified sessions, or does V2 define an interim state and then classify it?
NEXT: Resolve the first-match reading against the parsing and session rules, then check the A6 guard and rollback.
DECISION: V2 makes S4 the immediate no-token result before the local read. The first-match rule then gives a verified signed-in reader S3 and a guest S4. ACCT-2 pins the newly specified guest path.
FACT: The A6 settings guard itself calls fetchQuery; the current session query has staleTime zero. The rollback lists the external effects a frontend revert leaves in place.
NEXT: Complete the five-row defense and record both revision-10 closures.
FACT: Reviewed revision 11 at 4eae6978 on main, the A3–A7 audit, the named production and fixture sources, and the two revision-10 findings. The reviewed tracked paths match the pinned revision. Static review only.
FACT: Revision-10 finding 1 closed. V3 at plan.md:661 and Л6 at :736 reserve the post-result pause UI for A3 and require E2 with its settings action immediately after 204. The implementer must stop awaiting refresh in the A6 success callback (email.change.confirm.tsx:23). The plan explicitly specifies this behavior and its offline check. require-identity.ts:11–27 explains the separate read on entry to settings.
FACT: Revision-10 finding 2 closed. ACCT-3 at plan.md:769 copies the session before verification (accounts.spec.ts:261–283) and replaces its cookie in a separate context after the banner appears. The real resend authenticates the presented cookie (handler.go:357–367). A full document load reads it in a new request (identity.ts:10–19, root-identity.ts:21–32), maps the resulting /me 401 to null (session-repository.ts:270–281), and renders the login form (login.tsx:44–50, LoginView.tsx:35–43,97).
FACT: Revision-11 no-token path has its own mount-time read with no verification POST at plan.md:644 and guest route proof at :768. Empty and non-string search tokens take this path through token-search.ts:4–8. Once the read returns, the first-match rule at :633 and :640–641 distinguishes S3 from S4.
DECISION: All five rows are closed within the owner-approved scope. No new blocking gap found.

| # | Fixed question | Plan lines and answer | Verdict |
|---|---|---|---|
| 1 | Malformed input | plan.md:627–629, 640–644, 658, 688–694. Reject invalid guest email locally, give malformed links an actionable failure, and localize field errors. | Closed |
| 2 | Dependency failure | plan.md:624–627, 637, 642–647, 654–661, 681–684, 710, 736, 769. Preserve known results, distinguish offline pauses, and provide recovery for uncertain writes or expired sessions. | Closed |
| 3 | Race and repeated invocation | plan.md:631, 664, 685, 731–735. Guard repeated actions before render, bind token attempts to the mounted instance, and name the accepted session and cross-tab boundaries. | Closed |
| 4 | Empty and boundary data | plan.md:629, 640–648, 658, 688–694, 723, 760, 768, 779. Cover absent tokens and sessions, required and length-limited fields, and the full 2FA secret at narrow widths. | Closed |
| 5 | Rollback | plan.md:791. Revert the frontend commits. Sent mail and existing API effects persist. | Closed |

DECISION: Verdict — defended, 0 gaps.
