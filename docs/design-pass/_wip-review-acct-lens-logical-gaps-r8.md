NEXT: Read agent-entry rules, then load the logical-gaps lens contract and repository routes.
FACT: Read the shared entry rules and recorded the user-specified working file as the sole write destination.
NEXT: Verify repository root and revision, then read the scoped instructions and skill.
FACT: Repository HEAD is 34212f73; local instructions and the skill identify logical gaps as missing prerequisites, ordering, or unsupported conclusions.
DECISION: Inspect the single requested lens and honor Л1–Л6 as accepted boundaries.
NEXT: Read the complete ACCT plan and audit with line numbers.
FACT: The audit calls for all 16 fixes, including the two suspects, plus the A3/A6 title consistency note.
NEXT: Map the revision 8 rows to the promised scope and proof.
FACT: Revision 8 supersedes the earlier orchestration summaries. The review uses the operative rows from Facts onward.
NEXT: Read the operative requirements and evidence matrix in full.
FACT: Mapped V1–V15 and ACCT-1–20 to the audit. Candidate checks concern the real session refresh, mutation state, 2FA transitions, and verification proof prerequisites.
NEXT: Read the route and session implementation.
FACT: A3/A6 only refresh on success today. A7 uses the shared current query and stores enrollment and displayed recovery codes in route state. The lesson gate currently reads root route identity.
NEXT: Verify refresh behavior, existing view/CSS seams, and browser-proof fixtures.
FACT: The refresh helper invalidates the active session query. The account browser spec starts from a real registered session, but its Mailpit helper currently returns only the link. The first view/CSS read stopped after a zsh reserved-name collision.
NEXT: Complete UI and CSS reads, then trace backend response and navigation contracts.
FACT: The UI and backend reads confirm the audit’s existing state and the server’s token-first write, neutral resend, and 2FA ordering. CSS currently uses a 20px-token start margin on notice forms and a separate 16px-token margin on primary notice links.
NEXT: Test candidate logical gaps by tracing exact helper semantics and acceptance oracles in source.
FACT: The installed Query client resolves invalidateQueries immediately for a refetch already paused at dispatch; this contradicts Л6 if V2 gates only on the refresh promise. ACCT-9 also states an unconditional redirect even though V5a and Л1 allow a stale session write.
NEXT: Check the remaining audit-fix and proof prerequisites, including locale, box metrics, and focus.
FACT: The audit’s status text and focus destinations have implementation seams. The secret proof checks only vertical fit, and the mail proof currently names only a heading shared by several causes.
NEXT: Trace remaining concrete contradictions and select findings whose proposed edits preserve the narrowed scope.
DECISION: Keep the offline finding limited to the plan’s own timing contradiction and the 401 finding limited to the unconditional acceptance wording. No session coordinator is needed to resolve either.
NEXT: Confirm cookie handling and finalize the file with supported findings only.
FACT: Authentication failures leave the stored cookie untouched, so the ACCT-3 fixture is feasible. E2’s existing English success text says “Your new address,” despite both confirmation handlers reading the user from the token.
NEXT: Check the login and same-origin routes before recording the final dispositions.
FACT: The login view uses root route identity. The root uses ensureQueryData, which can return the cached viewer after a resend 401. V1 currently changes neither the cache nor the navigation mode.
NEXT: Check whether the new shared resend result can cross the planned presentation boundary.
FACT: LoginRoute passes cached signed-in identity to LoginView, which renders Continue and omits the credential form. The V1 shared result can be consumed through the identity module's public exports. Review evidence is static. No gates or runtime tests were run.
DECISION: Two blockers remain. The first affects the ordinary recovery path after a resend 401. The second is a mismatch between ACCT-9 and the behavior the plan implements and tests. Keep the accepted Л1–Л6 cases in scope as written and correct only the overclaims. The other findings refine the plan or its proof without requiring new session orchestration.

[logical-gaps] Resend 401 can send the user to a login page with no sign-in form
where: §18 V1 and ACCT-1/3, plan.md:629, 738–740, 763–765.
why: A stale cached signed-in viewer can still use the banner or lesson-gate resend. The API returns authentication_required, but clicking a router link to /login reuses the signed-in cache through resolveRootIdentity's ensureQueryData. LoginRoute and LoginView then show Continue instead of credentials. This needs no /me already in flight, so it is outside Л1. ACCT-3 starts from an already anonymous cache and misses this path. Sources are root-identity.ts:26–30, login.tsx:44–51 and LoginView.tsx:35–43.
action: Require a fresh document navigation for the resend 401 sign-in link, or otherwise ensure the login route sees the confirmed anonymous session. Retain the alert on its originating surface. Add one route proof that begins with cached signed-in state and follows the link to the credential form.
weight: blocking

[logical-gaps] ACCT-9 promises broader 2FA recovery and redirect than V5a specifies
where: §18 V5a, Л1, ACCT-9 and its proof, plan.md:676–682, 727, 746, 770.
why: ACCT-9 reads as requiring Reload after any uncertain 2FA failure. V5a limits that rule to confirm/disable and gives regeneration distinct guidance. ACCT-9 also promises that every 401 redirects, although V5a redirects only for authentication_required (recovery_code_invalid is also 401) and lets an in-flight session read overwrite null under Л1. The effect reads the rendered session value, so setting null does not guarantee the redirect if a stale read overwrites it first. Source is require-identity.ts:30–51.
action: Align ACCT-9 and its proof with V5a. Specify confirm/disable uncertainty, named conflict responses and the distinct regeneration behavior. Require the redirect for authentication_required in a settled-session fixture and scope the in-flight case to Л1. Keep recovery_code_invalid in the action group.
weight: blocking

[logical-gaps] The offline limitation promises to await a read the helper may skip waiting for
where: §18 V2 and Л6, plan.md:646, 732.
why: refresh() calls invalidateQueries(), whose refetchQueries branch resolves immediately if the refetch starts paused. Waiting on the helper alone does not delay the A3 actions until reconnection, as Л6 says. Sources are identity-queries.ts:35–37 and the installed TanStack Query queryClient.ts:321–339.
action: Describe offline A3 as using the snapshot available when best-effort refresh returns, with actions updating when the query settles. Keep Л6 accepted.
weight: non-blocking

[logical-gaps] A6 English success copy claims the current viewer's address changed
where: §18 E2 and Л4, plan.md:653, 730.
why: E2 retains the English text “Your new address is confirmed.” The A6 handler uses the token owner's identity without matching the viewer, just as A3 does. A user can therefore see their old address after following the settings CTA even when /me is fully current. Sources are en.ts:1188–1193, handler.go:321–343 and accounts.go:245–265.
action: Use neutral A6 success copy in both languages and keep the existing settings CTA.
weight: non-blocking

[logical-gaps] The mail and secret checks can pass without proving their acceptance rows
where: §18 ACCT-2 and ACCT-19 proof, plan.md:739, 756, 764, 775, 778–783.
why: The old-link heading is also shown when no token reaches the route, so it does not establish that the server refused the first token. The sole textarea measure, scrollHeight <= clientHeight, does not catch the horizontal clipping reported by audit #16 or prove that the full expected value reached the field.
action: Assert the real verification POST's 410/token_used response for the first link before checking S4. Check the full expected secret and its horizontal and vertical fit at the stated widths.
weight: non-blocking

FACT: Reviewed revision 8 at 34212f73 against the requested audit and source paths. Only this working file was edited. No other lens's working file was opened.
NEXT: Return the five findings and the two-blocker lens verdict.
