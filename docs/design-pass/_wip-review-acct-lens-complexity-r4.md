NEXT: Read the AI Factory entry rules and the complexity lens contract.
FACT: Read the shared entry rules. The owner explicitly scoped this run to one lens and one append-only working file.
NEXT: Verify the repository root and read its local rules, the phase briefing, and the lens contract.
FACT: The verified repository root is courses. HEAD and main both point to 08fc8ae7. The local rules and the kit contract were read. Current review authority is the requested snapshot and audit.
NEXT: Read §18, audit findings and owner answers, and prior complexity findings.
FACT: The prior three reviews contain four, three, and three findings respectively. The revision-4 commit touches one plan line, so closure requires examining the full §18 as well as the commit delta.
NEXT: Read §18, the audit, and changes between the reviewed revisions.
FACT: The owner approved all sixteen audit items and the terminology item. Revision 4 narrows token deduplication, keeps 2FA forms mounted during checks, and derives resend pending state from the mutation cache.
NEXT: Read the exact implementation and acceptance wording for those closures.
FACT: V4a and ACCT-7 now explicitly allow a fresh POST on remount and require recovery. V5a and ACCT-9 keep the current form during reconciliation. V1 and ACCT-1 use the mutation cache for the same-tick guard.
NEXT: Verify caller ownership and reuse paths against the frontend, backend, and browser-spec source.
FACT: Current routes already own their mutations, identity composition owns the session key and current query, and the account view renders the code panel separately. V6, V7, V11–V14 and ACCT-1 preserve the prior reuse fixes.
NEXT: Check the existing session read, the shared error codes, and the concrete cost of the revised retry path.
FACT: Origin rejection occurs before the token action, while verification and email change can fail after consumption. V2 currently routes both kinds of S5 errors through a session read and lets its retry shortcut ignore the S3 table predicate.
NEXT: Verify the client error mapping and whether the updated flow can reuse a single classification and S3 rule.
FACT: The adapter preserves 403 origin_forbidden and 429 rate_limited as typed errors (session-repository.ts:76–87). The origin middleware returns before invoking the token handler (handler.go:514–523). E5 already shows determinate refusals directly. Also verified the existing protected-settings return path and browser fixtures. The source diff between the plan's code baseline and 08fc8ae7 is empty for frontend and backend.
DECISION: Rev 1 closed 4/4. ACCT-1 verifies resend outcomes once plus caller wiring (:710), V6 scopes fallback to text controls (:658), V7 keeps error ownership with each existing mutation (:660), and V11–V13 extend existing selectors (:671–675).
DECISION: Rev 2 closed 3/3. V2 cancels the older session read before fetching the existing current query and ACCT-5 holds the older request (:618, :713). V14 uses the mounted recovery-panel lifetime (:677, :721). V13 extends the existing A5 margin declaration (:675).
DECISION: Rev 3 closed 3/3. V4a explicitly narrows the deduplication promise to the mounted instance and tests reopening (:633–634, :715). V5a keeps the same 2FA form mounted with read-only controls and a pending button (:649–653, :717). V1 derives shared pending state from the exact resend mutation key (:605, :710).
NEXT: Record the new retry-path finding and return the single-lens verdict.

[complexity] Use one S3 rule and read the session only when it can change the result
where: docs/design-pass/plan.md:614–618 (V2), :689 and :713 (ACCT-5).
why: V2 puts a fresh /me read on the path for all S5 results, including a 403 origin_forbidden where the middleware refused the POST before the token handler ran. Signed-in and guest users get the same S5 retry action, yet a failed /me can replace the known refusal with S0 and withhold Retry. V2 also adds a retry shortcut that goes to S3 for any verified signed-in user. On an uncertain result for a user who was already verified, the table correctly says S5, but pressing Retry can show S3 without any new evidence about the attempted link.
action: In V2, show known pre-write refusals as S5 immediately and let Retry repeat the token request. Keep the fresh exact-key session read for S3/S4 classification when identity changes the outcome or the available actions. Apply the same S3 predicate to the initial result and to retry reconciliation for an uncertain write. In ACCT-5, cover 403 with /me unavailable and an already-verified user's uncertain failure followed by Retry.
weight: blocking

FACT: Source-only review of main at 08fc8ae7. Only this working file was edited. No make check, Docker, or e2e commands ran. No other lens working file was read.
DECISION: 1 assigned / 1 returned.
Lens verdict: 1 blocking finding.
