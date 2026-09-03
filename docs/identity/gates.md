# Identity plan gates

Date: 2026-09-03. Review was read-only and independent from the authoring pass.

Plan-defense initially found four blockers: logout persistence failure, session fixation/at-rest token assertions, rollback behavior for live sessions, and the insecure-cookie startup matrix. The plan now expires the browser cookie but returns 503 on failed revocation; rotates a presented token and asserts unique 32-byte digests; bulk-revokes sessions before rollback; and permits `Secure=false` only for explicit loopback origins. Deterministic eviction, bounded digest-collision retries and deletion-cookie scope were also added.

Second defense: **GREEN, 0 blocking gaps**. Covered attacks: malformed input (E01); dependency failure (E03/E09); races/idempotency (B03/B04/B07–B09); empty/boundary data (A03/B02/B05/B06/E04); rollback (Rollback section). Max-Age=-1 is paired with past Expires for legacy cookie deletion.

An independent critique then reopened the gate. It found contradictions between SSR acceptance and the client-only protected routes, missing ownership for runtime security configuration and login-specific Nginx limits, incomplete `/me`/logout dependency-failure contracts, inconsistent invalid-email classification, and an unjustified five-login-session cap. The revised spec separates SSR identity rendering from client guards, assigns configuration/edge work to I03–I04, defines the 400/401/503 matrix, removes the speculative login-session cap, adds cache/history and hostile-return-path coverage, and narrows backend authorization claims until Phase 4 has a real protected business handler.

Plan-critique ran four independent lenses: **4 started / 4 returned** (AI-slop, logical gaps, excess complexity and edge cases; the interrupted complexity/edge work was retried with clean-context agents). The edge lens added atomic full-transaction login rotation, exact expiry boundaries, cross-account and duplicate-cookie behavior, 401-versus-503 client handling, strict duplicate JSON/Origin handling, atomic seed repeatability and concurrent logout. The complexity lens removed the unused Phase 3 authorization wrapper and required reuse of the existing request-scoped Query/API forwarding path.

After these plan edits, open blocking findings: **0**. Plan-critique is **GREEN**.

The revised plan-defense is **GREEN with 0 blocking gaps**. Its anchors are E01 for malformed input; E03/E09/E11 for dependency failures; B03/B06–B09/E08/E12 plus the transaction and seed decisions for races/idempotency; A03/B02/B04–B05/E04/E10 for empty and boundary states; and the audited bulk session revocation in Rollback. The one non-blocking precision note was closed by pinning the 8 KiB body, 254-byte email, 1024-byte password and 120-code-point display-name limits before implementation.
