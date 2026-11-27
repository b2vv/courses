NEXT: Read the AI Factory entry rules, confirm the repository and review revision, then check the five plan-defense questions.
FACT: The installed entry rules were read.
DECISION: The named working file is the only write destination.
NEXT: Verify the repository root and read its local instructions and the defense contract.
FACT: The workspace root and local rules were verified. The contract uses five fixed questions and requires plan line references.
DECISION: Keep the table in this working file under the read-only constraint.
NEXT: Read the briefing and task pointer, pin the review revision, and locate §18 and source-audit sections.
FACT: HEAD is 7bd73c9f on main, with §18 beginning at plan.md:538.
NEXT: Read the briefing, §18, the audit, and the two prior defense records.
FACT: The task tracker points to design-pass. Revision 10 explicitly supersedes the revision-9 session-observer gate.
NEXT: Read the normative lines and prior defense findings in bounded excerpts.
FACT: Revisions 8 and 9 each had gaps in rows 2 and 3. Revision 10 adds S0/E0, replaces the observer timestamp test, and names the enroll/confirm interleaving in Л5.
NEXT: Check the operative acceptance, limits, and audit mapping before tracing the implementation.
FACT: ACCT-1 through ACCT-20 cover audit #1–#16 and the taste item. The audit used mocked write successes and is evidence of UI behavior only.
NEXT: Trace the route and session mechanisms, including what happens when fetchQuery has cached data.
FACT: The production session query sets staleTime to zero, so fetchQuery waits for a query result or an already in-flight read. One source read needs a retry after the shell loop shadowed its search path.
NEXT: Read route and transport code alongside the two-factor and token write paths.
FACT: The prior automatic-token offline gap now has S0/E0. The new local fetchQuery promise removes the observer timestamp comparison and gives S2 a single failed-read action.
FACT: SQL confirms the exact enroll/confirm interleaving now named in Л5. Same-render settings guards remain at line 683.
NEXT: Read the affected UI, shared styles, backend error mapping, and browser fixture paths.
FACT: The named UI and CSS have the seams in V5–V15. The backend and adapter distinguish malformed input, gone tokens, conflicts, and dependency errors.
QUESTION: Does unchanged A6 refresh behavior provide the new Л6 post-result pause guarantee?
NEXT: Check the pinned query code, fixture sequencing, and final gap candidates.
FACT: The unchanged A6 refresh resolves immediately when its GET is paused, although Л6 promises a post-result waiting message and no action for A6.
QUESTION: Does the new ACCT-3 browser fixture actually expire the session it uses to show the banner?
NEXT: Verify the login branch and whether the new mocking step can reach the password field.
FACT: ACCT-3’s new banner browser scenario mocks only resend. A real session remains valid across the document navigation, so LoginView takes its signed-in branch.
NEXT: Resolve the verdict and append the prior-revision closure checks, findings, and table.
FACT: At 7bd73c9f on main, the plan, audit and requested source paths match the pinned commit. Audited all 16 fixes and the taste item by static reading. No make check, Docker or e2e was run.
FACT: Revision 8 closure. Л6 admits pre-transport offline pauses, and V1:629 gives resend a shared pending state until a response. V2:642 now waits on fetchQuery instead of the invalidateQueries promise. V5a:683 supplies same-render password/email guards. Л5:733 acknowledges concurrent enroll calls can both succeed.
FACT: Revision 9 closure. S0/E0 at :635 and :652 cover the automatic-token mutation pause. V2:642–646 uses the read promise instead of a timestamp and specifies one S2 fallback after failure. Л5:733 describes the precise enroll-before-confirm-check / enroll-write-after-confirm interleaving verified in two_factor.go:58–111 and identity.sql:13–19. The race itself is within the accepted limitation.
FACT: New row 2 gap. Л6:734 promises A6 post-result waiting copy and no action while the read is paused. V3:654,659 retains the existing success refresh. email.change.confirm.tsx:23 awaits refresh; identity-queries.ts:35–37 uses invalidateQueries; queryClient.ts:328–343 resolves immediately when the refetch is paused. After a 204 received as connectivity drops before /me starts, this exposes E2 and its action without the promised message. Specify the post-result wait for A6 or limit that promise to A3, keeping E0 for A6.
FACT: New row 2 proof defect. The added ACCT-3 browser step at :767 returns a mocked resend 401 while the banner account's real session remains authenticated. A document navigation rebuilds root identity from the still-valid session (root-identity.ts:26–30); login.tsx:44–50 and LoginView.tsx:35–43 then show “Продовжити”, so the asserted password field cannot appear. Keep a signed-in cache while invalidating the actual session, or otherwise make the new document receive anonymous identity. V1:624 remains viable with an expired session.
DECISION: Count unresolved fixed questions as gap rows. The two findings belong to dependency failure. The accepted limitations are otherwise closed for this defense.

| # | Fixed question | Plan lines and answer | Verdict |
|---|---|---|---|
| 1 | Malformed input | §18:625–627, 638–640, 655–657, 686–692. Validate guest email, give unusable links a recovery path, and localize field errors. | Closed |
| 2 | Dependency failure | §18:635, 640–646, 652, 654, 659, 679–682, 734, 767. A6's retained refresh cannot deliver the post-result offline behavior promised in Л6, and ACCT-3's banner test leaves the session authenticated. | Gap |
| 3 | Race and repeated invocation | §18:629, 662, 683, 731, 733. Action-time guards and token-instance rules cover same-client repeats. Л3 and Л5 record reopen and cross-tab boundaries. | Closed |
| 4 | Empty and boundary data | §18:638–646, 656, 686–692, 702, 721, 758. Covers absent tokens or user data, empty or invalid-length fields, and the full secret at phone widths. | Closed |
| 5 | Rollback | §18:789. Revert the frontend commits. Previously sent mail, token effects and account changes persist. | Closed |

DECISION: Verdict — 1 gap, row 2.
