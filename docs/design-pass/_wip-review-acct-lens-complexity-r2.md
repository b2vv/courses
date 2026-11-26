NEXT: Read the AI Factory entry rules and verify the repository root.
FACT: Read the AI Factory entry rules and requested the Git root.
NEXT: Read the repository briefing, local instructions and lens contract.
FACT: Verified the repository root and read the briefing, local rules, and plan-critique contract.
NEXT: Read the pinned plan section, owner evidence, and prior complexity findings.
FACT: Read §18 at 8d171f63, the owner evidence, the shared anti-pattern reference, and the revision-1 complexity findings.
NEXT: Trace the frontend and backend implementation seams cited by revision 2.
FACT: Traced the route-owned mutations, current-session query, shared form help, resend surfaces, and CSS selectors. Read the full revision-1 finding list.
NEXT: Verify the view, service and browser-test details against each proposed reuse path.
FACT: Confirmed existing useQuery/current and refresh primitives, server mutation boundaries, per-view state, and the current Mailpit/e2e fixture and assertions.
NEXT: Compare the revision delta against the four earlier findings and examine session refresh, CSS reuse and the shared resend coordination in detail.
FACT: V2 explicitly fetches the same session on error and again on retry. V5a patches an observed query whose query function consumes the abort signal. The 2FA view renders codes, setup, and management as exclusive branches.
DECISION: Revision-1 resend, localized validation, and group error findings are closed. The CSS item is partly closed: V11/V12 name existing treatments, while V13 still needs an explicit instruction to extend the A5 spacing declaration.
NEXT: Check whether an in-flight session read can overwrite the proposed 2FA patch and finalize actionable findings.
FACT: The account view exposes regeneration only outside the recovery-code branch. A component mounted only inside that branch inherits its lifetime, so keying by the codes is redundant for the specified interactions.
NEXT: Record the outstanding CSS item and the revision-2 findings, then return the single-lens verdict.
FACT: TanStack Query source confirms fetchQuery calls query.fetch without cancelRefetch; an in-flight query returns its existing promise. staleTime: 0 is already provided by current().
DECISION: The V2 ordering needs one scoped cancellation before a read that is meant to observe the token result.
NEXT: Append findings and verdict.
[complexity] Make the S3 session read start after the token attempt
where: docs/design-pass/plan.md:595 (V2), :680 (ACCT-5).
why: The shared current() options already set staleTime: 0 (identity-queries.ts:25–34). That setting leaves pending work in place: QueryClient.fetchQuery calls query.fetch without cancellation (queryClient.ts:414–420), and the query returns its existing promise (query.ts:400–415). The root observes the same session key (__root.tsx:176). If a session refetch captured emailVerified=false before verification committed and is still in flight when the verify request fails, V2 can read that old response and show S5 for an already verified signed-in user. The retry precheck can likewise join a read that predates the click and allow another token mutation.
action: Reuse the existing session key and current() options. For the V2 checks that require a new read, await cancellation of any in-flight read for that exact key, then fetch current() and branch on the result. Replace the redundant staleTime override and add an ACCT-5 case with an older session response held in flight.
weight: blocking

[complexity] Use the recovery-code branch lifetime for copy feedback
where: docs/design-pass/plan.md:644 (V14), :688 (ACCT-18).
why: The view renders the codes, the enroll form, and the enabled controls as exclusive branches (AccountSettingsView.tsx:131–177, :208–221). The next generation is only reachable after Done dismisses the current code panel. Explicitly keying by the set adds identity handling for a replacement that the specified UI cannot perform while the panel is mounted.
action: Move clipboard feedback into the child mounted only for the recovery-code branch and let Done unmount it. Keep the delayed completion and new-set assertions.
weight: non-blocking

[complexity] Extend the existing account CSS declarations
where: docs/design-pass/plan.md:638–642 (V11–V13). Revision-1 finding remains partly open.
why: V11 and V12 name the same treatments already implemented at main.css:514 and :465–472, while V13 still describes a separate spacing rule for an action already covered by the A5 declaration at :524–528. The general .notice form rule at :509–513 also covers login, register, forgot and reset, so a broad direct-form override would change those screens and enlarge verification.
action: Name the existing width and container selectors to extend. Extend the A5 spacing declaration narrowly to cover A3/A6 actions and the resend form, preserving the ordinary AUTH form margin.
weight: non-blocking

FACT: The revision-1 resend outcome matrix, localized validation scope, and 2FA error ownership are closed at plan.md:677, :625, and :627. CSS reuse remains partly open as recorded above.
FACT: Source-only review. No make check, Docker, or e2e commands ran. Only this working file was edited.
DECISION: 1 assigned / 1 returned.
Lens verdict: 1 blocking finding.
