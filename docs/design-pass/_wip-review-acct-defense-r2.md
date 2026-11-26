NEXT: Read shared entry rules and verify repository root.
FACT: Repository root verified on main. Shared entry rules read.
NEXT: Read local rules, briefing, task pointer and defense contract.
FACT: Requested revision 8d171f63 is HEAD on main. The defense contract requires five rows in order.
DECISION: Record the table in this working file under the owner’s read-only instruction.
NEXT: Read briefing, task index, revision 2, revision 1 findings and source audit.
FACT: Revision 1 left rows 1, 2, 3 and 5 open. The audit captured real invalid-link and guest resend responses, with successful account mutations simulated.
NEXT: Read the exact behavioral decisions and trace the named frontend routes and views.
FACT: The routes currently start link mutations from mount effects. Resend ownership is per surface. Settings reads the enabled flag from the session query and guards 2FA actions with shared pending state.
NEXT: Verify the server write order, adapter result mapping and existing data and browser proof boundaries.
FACT: The server can return 204 after mail failure. Verification consumes the token before writing. Recovery regeneration deletes the old set before inserting the replacement.
NEXT: Check precise status mapping, session refresh semantics, and repeated invocation paths before scoring.
FACT: The new keyed link state isolates displayed results. The current session query already uses staleTime: 0. A staleTime setting controls cached freshness and does not order a read after a write or prevent joining an in-flight read.
NEXT: Confirm the exact frontend copy, concurrent-request mechanics, and verification limits for the final table.
FACT: The pinned plan, audit and named code match the review checkout. Client and server validation and the existing e2e selectors have been inspected by source.
NEXT: Verify the installed Query and React Aria behavior relevant to the final race and boundary findings.
FACT: Installed TanStack Query confirms fetchQuery can reuse an in-flight promise. The token SQL admits one consumer per token, but resend invalidation and issuance are separate calls.
NEXT: Resolve the remaining scoring decisions and write the five-row result.
DECISION: Count gaps by fixed question. Revision 1 row 1 is closed by V1/V6 and row 5 by the rollback clause. Row 2's clipboard finding is closed by V14. Its 204 finding remains because the new text still promises eventual delivery. For row 3, V2/V3 now name recovery after token consumption, V4a isolates outcomes by token, and V1 defines the scope and owner of resend feedback. The same-tab guard and the promised fresh session read lack ordering guarantees.
FACT: Source review only. The plan and named sources match 8d171f63. No Make, Docker, or e2e was run.
NEXT: Return the table and verdict to the plan author.

| # | Question | Plan line at `8d171f63` and defense | Verdict |
|---|---|---|---|
| 1 | Malformed input | §18 lines 579–581, 592, 604, 625, 678. V1 blocks submission with a missing, malformed or overlength guest address and V6 defines localized field errors. A server-side 400 gets the general resend alert. A3/A6 map missing or rejected tokens to S4/E4. | Closed |
| 2 | Dependency failure | §18 lines 575–579, 593–595, 605–607, 623, 627–630, 644. V14 now defines manual copy after clipboard failure, and V2/V3 give link recovery actions. V1 still promises «новий лист надійде за кілька хвилин» after 204, even when the mailer failed after invalidating the old link (`accounts.go:133–140`, `handler.go:373–377`). State what to do when no letter arrives and keep the response neutral. V5a also needs the visible outcome when both confirmation and the follow-up /me fail, when the account may already require 2FA and codes may be unavailable. V7's generic group error leaves recovery unspecified if regeneration deleted the old codes and failed to return replacements (`two_factor.go:227–241`). | **Gap** |
| 3 | Race and repeated invocation | §18 lines 583, 595, 605, 610, 623, 677, 682. V4a isolates late results for distinct tokens and V2/V3 route sequential `token_used` to recovery. V1 explicitly excludes cross-tab resend coordination, but `mutationKey` groups mutations and `useIsMutating` supplies a rendered pending snapshot. Specify a synchronous request-time guard and a test with two triggers before the pending render. V2's `fetchQuery` with `staleTime: 0` can join an older in-flight /me (`identity-queries.ts:19–34`, installed Query core `query.ts:401–415`), so S3 may be lost after a committed verification. Specify a read started after the relevant mutation or a subsequent retry when the result is inconclusive. V4a also needs single automatic submission for one mounted token if it retains the existing idle-triggered effect under StrictMode (`verify-email.tsx:25–27`, `email.change.confirm.tsx:25–27`, `client.tsx:24–29`). | **Gap** |
| 4 | Empty and boundary data | §18 lines 581, 592, 604, 634–646, 667–670, 689. The plan covers absent tokens or session data, empty forms after success, the completed recovery-code state, and full-secret readability at the named narrow and enlarged-text viewports. The existing adapter enforces ten codes (`session-schema.ts:19–24`). | Closed |
| 5 | Rollback | §18 line 701. Revert the frontend slice. The clause explicitly preserves the effects already applied by the API, including sent mail, invalidated links, consumed tokens, account changes, and replaced recovery codes. | Closed |

**Verdict: 2 gaps (rows 2 and 3).**
