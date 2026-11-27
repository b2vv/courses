NEXT: Read AI Factory entry rules and the plan-critique contract, then verify the revision and scope.
FACT: Verified repository root and read the AI Factory entry rules.
DECISION: Keep all review output in this working file and run only the assigned lens.
NEXT: Read local rules, the briefing, and the critique contract.
FACT: Read the edge-cases criteria, local rules, and repository briefing context.
NEXT: Pin revision 9 and read its plan, the revision-8 findings, and audit scope.
FACT: Revision 9 is on main and the reviewed plan and code match the pinned revision. Revision 8 returned three blockers and one acceptance clarification.
NEXT: Map revision 9 to the audit and inspect the changes from revision 8.
FACT: Revision 9 adds observer-based A3 decisions, pending focus preservation, confirm cancellation handling, and exact 401 codes.
NEXT: Read the current decision and verification rows without the historical revision notes.
FACT: Mapped the operative V1–V15 decisions, Л1–Л6, and ACCT coverage rows.
NEXT: Check route observers, session API contracts, account form transitions, CSS, and browser fixture sequences.
FACT: The backend distinguishes the three 401 codes and confirm updates state before writing recovery codes. Settings currently holds enrollment separately from each mutation result.
NEXT: Trace client refresh and paused-mutation semantics and check the exact fixture and error boundaries.
FACT: The TanStack observer exposes query fetch status and a separate paused flag on mutations. TanStack clears the latest mutation error when a retry starts.
NEXT: Check whether V2 covers both paused phases and whether V5a preserves uncertainty across a later rejection.
FACT: A3 can pause before any token POST or after the result on the session GET. The revision specifies only the latter in V2. The 2FA setup remains mounted when its latest confirmation result changes.
NEXT: Verify later 2FA request rejection and confirm which conflicts are direct contradictions of revision 9.
FACT: A nonempty confirmation code over 128 bytes is rejected before the 2FA service, so a later 400 cannot resolve an earlier partial write.
FACT: Session queries use staleTime 0 and default focus/reconnect refresh, so the observer can return to fetching after the action has already rendered.
NEXT: Confirm the applicable default observer behavior and record dispositions and findings.
FACT: Revision-8 refresh finding — the paused and failed post-attempt /me paths now use the observer and have route proof.
FACT: Revision-8 focus finding — V5a now uses isPending, and ACCT-8 covers keyboard submission with a rejection.
FACT: Revision-8 cancel finding — V5a/V8 and ACCT-9 cover the immediate error-to-cancel path. The later-rejection transition still needs a rule.
FACT: Revision-8 acceptance clarification — ACCT-9 and its proof now name authentication_required, invalid_credentials, and recovery_code_invalid separately.
DECISION: Report three concrete regressions or remaining gaps in the new rules. Keep the accepted stale-session, offline deferral, uncertain-write, and cross-tab bounds.
NEXT: Return the fixed-format findings and lens verdict.

[edge-cases] A background session refetch can remove a live resend form
where: §18 V2 (629–642), ACCT-1/2 and ACCT-4/5 proof (735–739, 760–763).
why: V2 suppresses the action whenever the shared observer is non-idle. The session query uses staleTime: 0 and retains TanStack's focus-refetch default (identity-queries.ts:25–37, queryObserver.ts:118–123). After S4 has shown the email form, switching to an inbox and back can remove the form and its focused field during a later background GET. This can interrupt input or retry on the surface being fixed by #1/#2 even after the post-attempt read has settled.
action: Limit the action gate to the read used for the current token result or initial no-token decision. Specify that a later background pending state preserves the mounted form and focus. Add a case with the S4 field focused when a deferred background refetch starts.
weight: blocking

[edge-cases] A3 has no waiting-for-connection state for a paused token POST
where: §18 V2 S1 and post-result rules (633, 639–642), Л6 (729), ACCT-4/5 proof (763).
why: In a loaded app, going offline before the token mutation starts leaves it pending and paused before any POST (TanStack mutation.ts:208–238, 360–371). A3 currently drives its progress view from verify.isPending (verify-email.tsx:17–31), and V2 only checks a session query's pause after a token result. Following V2 therefore leaves the user at «Перевіряємо посилання…» with no request in progress, contradicting Л6's promised A3 status.
action: Give a paused token mutation precedence over the ordinary S1 progress text and show «Чекаємо на з'єднання…» without an action. Distinguish it in route proof from a session GET paused after a token result.
weight: blocking

[edge-cases] A later rejected confirmation can reopen Cancel while 2FA status is still uncertain
where: §18 V5a/V8 (674–679, 695), ACCT-9/12 and ACCT-9 proof (743, 746, 767).
why: The first confirm can enable 2FA and return 503 while saving recovery codes (two_factor.go:101–109). A second confirmation can be rejected before checking that state, for example with a nonempty code longer than 128 bytes (handler.go:215–228, 704–715). The setup remains mounted and the latest mutation error is replaced. V5a/V8 hide Cancel for an uncertain result or 409 but give no rule for retaining that condition across the later 400, which can restore Cancel and the false disabled status that revision 9 intended to avoid.
action: Keep the reload action and Cancel restriction for that setup after an uncertain or conflicting confirm until a successful confirm or a reload resolves the state. Cover uncertain confirm followed by a pre-write rejection. Preserve request-free Cancel for ordinary unfinished setup.
weight: blocking

FACT: One assigned lens completed. No make check, Docker, or E2E was run. This working file is the only edit.
lens verdict (3 blocking findings)
