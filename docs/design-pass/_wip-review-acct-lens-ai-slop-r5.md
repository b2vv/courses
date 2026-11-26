NEXT: Read shared entry rules and establish the review baseline.
FACT: Read AI Factory entry rules.
NEXT: Resolve repository root and read scoped rules and critique contract.
FACT: Repository root and HEAD match 6a8d1a3e. Read local rules and the plan-critique contract.
DECISION: Review only the AI-slop lens, with no delegation or runtime checks.
NEXT: Read §18, audit evidence and the four prior lens records.
FACT: Read each earlier AI-slop review and the applicable briefing and task pointers.
NEXT: Read the pinned plan, revision 5 delta, owner answers and untruncated revision 4 dispositions.
FACT: Main is at the requested revision. Owner approved all sixteen audit items and the A6 label.
NEXT: Read the final plan text and proof rows in full.
FACT: Read the complete status, focus, acceptance and verification portions of §18. The A3 403 recovery now prescribes a full reload; A6 still offers a same-token retry.
NEXT: Compare origin, session and 2FA recovery claims with handlers and route wiring.
FACT: The Origin middleware rejects before token handling and checks the same values on every request. The settings guard reacts to cached null. The existing admin route can disable 2FA.
NEXT: Verify shared UI and session facts, administrator recovery, and the existing browser fixture.
FACT: The shared resend UI has three real callers; V6 excludes choice controls. Mailpit and the focus assertions extend existing browser journeys. A 409 exposes a 2FA state mismatch, with no information about which request caused it.
NEXT: Verify exact line references, locale support and the remaining recovery claims before classifying findings.
FACT: The accounts browser spec already exercises administrator reset and confirms a subsequent sign-in. The plan itself records a stable origin-refusal example on the wrong host.
NEXT: Confirm the supported B-side remedies and rule out unnecessary findings against the A→B→A and admin recovery additions.
FACT: config.go:61–64 compares the presented origin to configured publicOrigin. Plan §4 records the wrong-host example. The account locale already has a plain refusal and cause-neutral messages for the two 409 codes.
DECISION: The in-flight map has two real callers and a specific A→B→A proof. Administrator reset is an existing recovery operation exercised in accounts.spec.ts:521–533. Neither adds hypothetical product work.
NEXT: Record previous finding dispositions and return the two non-blocking findings.

## Earlier findings

- Rev 1: delivery promise closed by the conditional 204 acknowledgment. Guest cases on signed-in surfaces closed by ACCT-1–3. Validation scope closed by the explicit TextField/TextArea boundary.
- Rev 2: 204 delivery and invalidation claims closed for the documented replacement path. E3/E4 now direct users to check the current address. ACCT-8 controls response release instead of waiting a fixed interval.
- Rev 3: S3 directs the holder of a rejected link for another address to request a new letter. V5a distinguishes code, session, request and uncertain-write outcomes, and offers both code types plus the existing admin reset path. V1 scopes invalidation to an attempt for a found, unverified account before issuance. The Facts section correctly describes A6's code-specific copy.
- Rev 4: expired-session redirect closed by the exact authentication_required cache write and the cached-user test. The same-origin retry was accepted in rev 4 and remains in E5. Rev 5 adds reload to S5 without showing how it can change the rejected origin.

[ai-slop] Reloading the verification page repeats the origin refusal
where: Plan §18 V2 S5 (line 625), V3 E5 (line 637), ACCT-5/6 (lines 724–725).
why: A full reload of the same URL still sends the token from the same origin. The middleware checks the Origin header against the configured public origin before token handling (handler.go:514–523, config.go:61–64). The plan itself records a wrong-host refusal (§4, line 75). ACCT-5 adds a reload path and a test that proves the reload alone. E5 retains the same-origin retry.
action: Show the refusal and keep applicable navigation for both routes. Keep retry for failures it can recover from. Add a 403 recovery action only if the plan can name a supported route that changes what the origin check sees.
weight: non-blocking

[ai-slop] The 409 copy invents another window
where: Plan §18 V5a (line 664), ACCT-9 (line 728).
why: The server reports whether 2FA is already enabled or currently disabled (two_factor.go:84–90, 258–265). For example, an administrator can disable it through the recovery route cited in V5a (two_factor.go:244–255). The instruction to show and test «Стан 2FA змінився в іншому вікні.» attributes a cause that neither the 409 nor /me establishes.
action: Keep the fresh state read and focus target. Show and test the confirmed state, using the existing cause-neutral 2FA copy where it fits.
weight: non-blocking

FACT: One requested lens, one verdict returned. Verified by source inspection at 6a8d1a3e. No make check, Docker or e2e ran. Only this working file was edited.

Lens verdict: no blocking findings.
