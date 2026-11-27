NEXT: Read the shared entry rules and verify the repository root.
FACT: Shared entry rules read and repository root verified.
NEXT: Read the local project rules, briefing, task index, and edge-cases contract.
FACT: The skill defines edge cases as boundaries, repeats, concurrency, dependency failures, and partial application.
DECISION: Apply only this lens to revision 8 and honor the owner-accepted limitations.
NEXT: Read the pinned section, audit, and shared anti-pattern reference.
FACT: HEAD and main both resolve to 34212f73; reviewed the revision marker and proof matrix.
NEXT: Read V1–V9, limitations, and the audit in detail.
FACT: Mapped audit #1–#16 to V1–V15 and the acceptance rows.
NEXT: Trace route and session behavior, including token validation and refresh.
FACT: The routes use the same query key and distinguish link errors via the adapter.
NEXT: Check refresh completion, server write order, and the relevant error and validation boundaries.
FACT: Session refresh is query invalidation. API errors distinguish expired authentication, wrong passwords, wrong factor codes, and uncertain writes.
NEXT: Inspect views, form behavior, CSS, and E2E fixtures for concrete edge cases.
FACT: Recovery-code copying currently returns silently when the clipboard API is absent. The 2FA test carries state from regeneration into disable.
NEXT: Check the exact fixture sequence, backend failure modes, and CSS proof boundaries.
FACT: The server distinguishes wrong credentials and unusable sessions even though both can return 401. Confirm can leave 2FA enabled after recovery-code persistence fails.
NEXT: Verify the refresh failure, A6 copy, and focus/cancel edges against the exact code and plan lines.
FACT: Existing 2FA buttons use native isDisabled during requests. A6 success copy is neutral even if the token belongs to another account.
NEXT: Confirm paused/error refresh semantics and the current browser test contract.
FACT: TanStack Query's installed queryClient.ts:298–343 swallows refresh errors by default and resolves immediately when the refetch is already paused. The session query can retain prior data in either case.
DECISION: Record the contradiction with the stated V2/Л6 behavior and the in-scope 2FA focus and cancellation cases. Retain the accepted session consistency boundaries.
NEXT: Return the following three blocking findings and one acceptance clarification.

[edge-cases] A3 refresh completion can select an action before the session outcome is known
where: §18 V2 S2–S4 and post-attempt refresh (636–646), Л6 (732), ACCT-4/5 verification (766).
why: identityDependencies.refresh calls invalidateQueries (identity-queries.ts:35–37). The installed TanStack Query resolves that promise immediately for an already-paused refetch and suppresses rejected refetches (queryClient.ts:320–343). A3 can then use the retained signed-in or anonymous data to offer a session-dependent action before connectivity returns, contrary to Л6, or select S3/S4 despite the failed-read branch in V2.
action: Specify how V2 uses the active session query's fetch and error state after requesting invalidation. Cover an already-paused refresh and a failed refetch with retained data. Keep the existing shared query and the Л1 stale-response limitation.
weight: blocking

[edge-cases] A failed 2FA write can drop keyboard focus
where: §18 V5/V5a (666–682), ACCT-8/9 (745–746, 769–770).
why: The existing 2FA triggers become natively disabled while pending (AccountSettingsView.tsx:173, 204, 218, 234). Pressing Enter can therefore blur the trigger before a 422 or 409 arrives. V5 moves focus only when replacing the content and promises focus stays put when an error leaves it mounted. The planned success-only browser assertions miss this path, so audit #6 can still fail.
action: Extend the focus-preserving pending behavior in V1 to the focused 2FA triggers and keep the shared submission guard. Add a browser assertion for a keyboard-triggered rejection that leaves the form mounted.
weight: blocking

[edge-cases] Cancelling after a failed confirmation can erase the recovery action
where: §18 V5a (678), V8 (698), Л2/Л5 (728, 731), ACCT-12 (749).
why: Confirm may enable 2FA and then fail to store recovery codes (two_factor.go:101–109). A confirm conflict can also mean 2FA is already enabled. Once either request settles, V8 permits Cancel to replace the setup group and show “2FA is not enabled.” V5a placed the uncertainty and Reload action in the group that just disappeared. The contradiction affects audit #9 even when the accepted server uncertainty is preserved.
action: Define cancellation after an uncertain or conflicting confirm so the card retains its status warning and Reload action. Keep the current immediate, request-free cancellation for ordinary unfinished setup. Add a view case for the failed-confirmation-to-cancel transition.
weight: blocking

[edge-cases] The settings acceptance row conflates distinct 401s
where: §18 V5a (681), ACCT-9 and its proof (746, 770).
why: V5a keys the redirect on authentication_required, while ACCT-9 and its test say 401. Wrong current passwords return invalid_credentials with 401, and invalid recovery codes return recovery_code_invalid with 401 (handler.go:815–818, 849–850). Following V5a yields the intended inline error while the current acceptance wording can lead to a false test failure.
action: Name authentication_required in the acceptance and proof rows. Check that both other 401 responses keep the user on the form with their own errors.
weight: non-blocking

FACT: One requested lens reviewed the pinned source. No build or runtime checks were run.
Lens verdict: 3 blocking findings.
