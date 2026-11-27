NEXT: Read the shared entry rules, the plan-critique contract and repository routes.
FACT: Shared entry rules read. Owner authorizes only the ai-slop lens and this working file.
NEXT: Read the repository briefing and current task route, then extract revision 8 and the source audit.
FACT: Repository root and main HEAD verified at 34212f73. Local rules, repository rules and contract consulted.
DECISION: Apply owner exclusions Л1–Л6 as written. Use the fixed finding format and leave the plan read-only.
NEXT: Read §18 and audit #1–#16 with the briefing and task context.
FACT: The on-disk plan and audit match HEAD. The review material includes the owner audit answers, scope table and acceptance obligations.
NEXT: Trace the named frontend, backend and E2E paths against the proposed steps.
FACT: Audit confirms resend outcomes and paths, form and focus changes, recovery-code dismissal and secret visibility are owner-backed work.
NEXT: Verify claimed mechanisms and separate user-backed steps from redundant or speculative requirements.
FACT: Mapped §18 decisions, acceptance and browser evidence to the named source files.
NEXT: Check the strongest filler candidates against exact code and audit lines.
FACT: Existing confirmation views already apply aria-busy. Existing 2FA forms expose distinct operations with real shared-action pressure and recovery-code copy state is page-scoped.
NEXT: Confirm backend outcomes and retire any unsupported finding candidates.
FACT: Current A3/A6 invalidate the session only after a successful write. A6 navigation has a fixed destination and account settings already own the identity guard. A3 V2 has some CTA branches that do need session identity.
DECISION: Judge unnecessary A6 post-failure reads independently of the accepted stale/offline outcomes.
NEXT: Write weighted findings and verify the working file.
FACT: A6 settings navigation already does a guarded fetch; the known email_taken result returns before UpdateEmail. A3 S2 and S3 each repeat their heading in their body.
DECISION: Report these as non-blocking. V1 shares actual behavior across three places, V5 addresses observed lost focus and V6 fixes an existing shared rendering problem.
NEXT: Append the fixed-format findings and verdict.
DECISION: The following findings are actionable plan edits of non-blocking weight. None challenges the owner’s acceptance of Л1–Л6.

[ai-slop] A6 refreshes the session after definitive rejections
where: §18 V3, lines 654–658
why: The blanket post-result refresh adds a session invalidation and potentially a /me request even for a missing token or email_taken. The latter returns before UpdateEmail (accounts.go:245–265). These states have a fixed settings CTA, and the settings guard already reads the session and handles guest sign-in (settings.account.tsx:10–12, require-identity.ts:11–27).
action: Keep the existing A6 refresh on success. Render failure navigation from the token response and let account settings read the address through its guard.
weight: non-blocking

[ai-slop] A3 repeats the confirmation heading in the body
where: §18 V2, S2–S3, lines 641–642
why: S2 says the email was confirmed twice before showing the next action. S3 repeats the same status and adds only the useful distinction that it refers to the signed-in account.
action: Drop the S2 body. Put the signed-in-account distinction into the S3 heading or use its body to clarify that the link itself failed. Keep the CTAs and S2’s neutral wording about ownership.
weight: non-blocking

[ai-slop] Superseded review history repeats the current checkpoint
where: §18 lines 542–599 and 801–815
why: Earlier decisions and gate totals reappear in multiple places. A reader encounters abandoned requirements for session refresh and token tracking alongside the current slice, and a future review has to keep the duplicated status in sync.
action: Keep the revision 8 scope decision and one current review checkpoint in §18. Replace the repeated earlier summaries with one pointer to the existing history.
weight: non-blocking

FACT: One requested lens reviewed and returned. Static review of the specified files only. No make check, Docker or E2E run. No other lens’s working file opened.
DECISION: Lens verdict (no blocking findings).
FACT: Verified the appended finding format, weights and verdict in the working file.
