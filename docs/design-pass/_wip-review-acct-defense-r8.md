NEXT: Read the shared entry rules, then establish the repository and revision before reviewing §18.
FACT: Read the shared agent entry and checked the repository root and Git refs.
NEXT: Read the scoped reviewer contract, repository briefing, and task pointer.
FACT: HEAD and main both resolve to 34212f73. AGENTS.local.md selects the canonical AI Factory kit.
NEXT: Read the briefing, §18 and audit with line numbers.
FACT: The task pointer identifies DP2 and the A3/A6/A7 audit. The skill requires five questions and a plan-line citation for each closure.
DECISION: The user’s read-only scope places the table in this working file and the final reply. Review the fixed commit and use Л1–Л6 as accepted scope.
NEXT: Read §18 end to end, then map all 16 audit items to plan lines and current code.
FACT: Read the repository briefing, the audit’s #1–#16 and its distinction between mock and live observations, and located §18.
NEXT: Trace route, UI, session and API behavior against §18’s commitments.
FACT: §18 acceptance spans all 16 audit numbers, with dedicated route, component, or browser checks. §18 gives an explicit frontend revert and enumerates persistent effects.
NEXT: Match the operative V-rules to concrete request, validation, focus and styling mechanisms.
FACT: The current A3/A6 token effect is unkeyed, the three resend surfaces own separate mutations, and the A7 card maps A7 form and factor states as described in the audit. Current CSS and shared fields match the proposed extension points.
NEXT: Verify server write ordering, retry semantics, runtime error mapping and existing e2e fixture paths.
FACT: Read the named routes, self-service views, shared primitives and identity domain. The code distinguishes malformed link input and API rejection from network outcomes.
NEXT: Check plan coverage for offline writes and races in each action, including server token atomicity and repeated ordinary settings submissions.
FACT: The production query client leaves the default online mode for mutations. An offline write can pause before its transport starts, leaving V1 without a result and V2/V3 in pending. Л6 only covers /me.
FACT: V4a’s token guard, the conditional database consume, and V1/V5a’s action-time mutation-key guards cover their stated repeat paths. V10 describes read-only fields and success cleanup for the other two saves.
NEXT: Verify whether the plan’s race guarantees also hold for the signed-in password/email forms and the explicit Л5 conflict mechanism.
FACT: The backend issues an independent token and attempts mail for each email-change request. V10 supplies no action-time guard equivalent to V1/V5a for the email or password forms.
FACT: Л5 promises 409 across tabs, but 2FA enroll reads Enabled and writes the setup secret in separate operations; two enroll calls can both succeed. The second can replace the secret already shown in the first tab.
FACT: The existing refresh delegates to invalidateQueries, whose refetch promise completes immediately for a paused query; awaiting it does not enforce Л6’s stated timing for A3 actions.
NEXT: Resolve the five row verdicts and append the final table.
FACT: Reviewed §18 at 34212f73 against all named source areas and A3–A7 audit #1–#16. The acceptance mapping is V1 (#1–2), V2–V4 (#3–5), V5 (#6), V6 (#7), V7–V10 (#8–11), and V11–V15 (#12–16). Verification was static. No make, Docker, or e2e was run.
DECISION: Count unresolved fixed questions as gap rows. Apply Л1–Л6 as owner-approved scope. Report only the mismatches between Л5/Л6 and the stated mechanisms, along with uncovered in-scope behavior.
FACT: Row 2. V1 covers received responses and V4b covers adapter failures. The production query client sets only query defaults (frontend/src/router.tsx:35–37). The pinned TanStack mutation code marks offline writes pending before running the transport (query-core/src/mutation.ts:202–238, retryer.ts:53–56, 227–234). Starting resend while offline can leave #1 with a pending button and no result until reconnect. This is a paused write, distinct from Л6's paused /me read. For Л6 itself, the proposed existing refresh calls invalidateQueries (identity-queries.ts:35–37), and refetchQueries resolves immediately for a paused read (queryClient.ts:328–343). V2’s refresh completion alone cannot enforce the stated wait for reconnect.
FACT: Row 3. V1 and V5a guard sibling calls synchronously and V4a covers the token effect. The token store conditionally consumes each token once (generated/identity.sql.go:39–46). V10 only specifies read-only fields and success cleanup. Current password and email handlers gate on render-time pending props (AccountSettingsView.tsx:38–46), leaving two submissions before a rerender unspecified. Each accepted email-change request issues a distinct token and attempts a separate email (accounts.go:225–242, 272–284). Л5 also states that a concurrent cross-tab action yields 409, yet two 2FA enroll calls can both read disabled and save distinct secrets (two_factor.go:58–81), leaving the first tab with a replaced setup secret. The admitted cross-tab limit stands; its promised 409 is the mismatch.
DECISION: Rows 1, 4 and 5 are closed. Rows 2 and 3 remain open.

| # | Fixed question | Plan lines | Verdict |
|---|---|---|---|
| 1 | Malformed input | §18:632, 642–643, 655, 684–690. Guest email is validated, unusable links get recovery actions, and field errors are localized. | Closed |
| 2 | Dependency failure | §18:626–634, 646, 662, 732. Offline writes can pause before reaching the transport, and the existing refresh does not wait for a paused /me. | Gap |
| 3 | Race and repeated invocation | §18:634, 661, 682, 703, 731. V10 leaves pre-render double-submit unspecified, and Л5’s 409 is not guaranteed for concurrent enrollment. | Gap |
| 4 | Empty and boundary data | §18:642–643, 655, 684–688, 700, 719, 753–756. Covers absent tokens and session data, blank and length-invalid fields, and the narrow-view secret. | Closed |
| 5 | Rollback | §18:787. Revert frontend commits. Existing emails, token effects, credentials, 2FA state and recovery-code changes survive. | Closed |

DECISION: 2 gaps, rows 2 and 3.
