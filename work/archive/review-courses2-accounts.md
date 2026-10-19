# Code review (Courses 2) — accounts phase (retroactive, merged PR #7 / d263045)

Reviewer: Courses 2 (Opus 4.8). Read-only; nothing committed. Skills applied: `code-review` and
`pattern-architecture-review`. Verified against the v0.44-era code at `d263045`, reading each cited file.

## Scope of this review (stated so the verdict is not overclaimed)

The diff is 166 files / ~9.6k lines. I reviewed the **security-critical surface the handoff named**,
file by file, against the code — not every line of all 166 files. Specifically covered: the
email-verified seam and the `CanStream` content gate; TOTP; the Redis pending-2FA token; recovery
codes; account-token digests; the Mailer (CRLF, logging); the login/2FA rate limiter; and the CSP.
Not covered here: full input-validation of every HTTP handler, the frontend account screens end to
end, the accounts SQL migrations column by column, and session rotation/remember-me internals. Those
were behind the phase's own green gate and e2e (101 passed); this is a targeted security read, not a
line-by-line of the whole phase.

## Verdict

**APPROVE (retroactive) — 0 blocking, 5 non-blocking.** The security core is well-built: constant-time
TOTP, atomic one-time recovery-code consume, digest-at-rest tokens, a header-injection-safe Mailer that
logs no secrets, and an enforcing nonce CSP. The non-blocking items are one inaccurate claim in the
handoff and four conscious trade-offs worth recording as limitations.

## Verified correct (evidence)

- **Email-verified gate.** `access.CanStream = CanView && EmailVerified` (`contracts/access/access.go:27`)
  is a pure contract predicate. `httpserver.Actor.EmailVerified` is populated from the DB user in the
  auth middleware (`identity/adapters/http/handler.go:501`), not hardcoded. The gate is actually
  enforced on every content path: playback authorize (`playback/application/authorize.go:37`), playback
  sessions (`sessions.go:80`, the nginx `auth_request` answer), and transcript read
  (`transcript/application/read.go:66`) — all thread `EmailVerified`. Email verification consumes a
  token bound to its own user (`accounts.go:102`), so it cannot verify an arbitrary account.
- **TOTP (RFC 6238).** `subtle.ConstantTimeCompare` on the code, ±1 period window, SHA-1/6-digit/30s,
  160-bit secret, canonical unpadded Base32 (`domain/two_factor.go:36-50`). No timing oracle.
- **Pending 2FA token.** `GetDel` makes `Take` an atomic one-time read (`redis/pending_two_factor.go:44`);
  TTL on `Put`; the id is 256 random bits (`application/two_factor.go:255-261`); completion is
  rate-limited on the id and, on a failed proof only, re-stored for a bounded retry; a successful proof
  is never restored (`two_factor.go:120-154`). It holds only `UserID`+`RememberMe`, no auth material.
- **Recovery codes.** Atomic one-time consume in a single statement —
  `UPDATE recovery_codes SET consumed_at=now() WHERE user_id=$1 AND code_hash=$2 AND consumed_at IS NULL
  RETURNING id` (`generated/identity.sql.go:41-48`) — so concurrent reuse yields exactly one success.
  Stored as SHA-256 hex digests, never plaintext (`two_factor.go:247`, `repository.go:398-409`).
- **Account tokens.** 256-bit raw emailed once; only the SHA-256 digest is persisted
  (`account_tokens.go:19-27`). Verify decodes the same base64url and hashes the same bytes
  (`accounts.go:295-300`), so issue and verify digests are consistent; consume is by digest.
- **Mailer.** `encodeMessage` rejects a Subject with `\r\n` and validates both addresses with
  `net/mail` plus `parsed.Address == value` (no display-name, no CRLF, no multi-address)
  (`platform/mail/sender.go:108,157-166`) — header injection is closed. Neither `sender.go` nor the
  identity `mailer.go` logs the token or the reset link (no log calls at all in the mailer).
- **CSP.** Enforcing by default; `Content-Security-Policy-Report-Only` only when `CSP_REPORT_ONLY=true`
  for e2e (`frontend/server.mjs:6-39`). `script-src 'self' 'nonce-…'` with a fresh 256-bit per-request
  nonce and **no** script `unsafe-inline`; `object-src 'none'`, `base-uri 'self'`, `frame-ancestors
  'none'`, `connect-src 'self'`. The nonce is threaded to TanStack Start via `ssr: { nonce }`
  (`router.tsx:54`), the framework's stamping hook.

## Pattern / architecture (clean)

- Ports are well-segregated: `TwoFactorStore`, `RecoveryCodeStore`, `PendingTwoFactorStore`,
  `PasswordVerifier` (the last "deliberately separate from `PasswordHasher`" so a re-auth caller cannot
  obtain hashes — textbook ISP/least-privilege, `ports/two_factor.go`).
- The 2FA domain is pure policy (crypto + `Validate*` functions); all I/O is in adapters. `TokenMaterial`
  separates `Raw` from `Digest`. No framework/DB leakage into domain, no GoF misuse, no over-engineering.

## Non-blocking

1. **Handoff claim vs code: the pending 2FA token is NOT bound to IP.** The handoff lists "pending-token
   Redis (одноразовість, прив'язка user+IP)". `PendingTwoFactor` holds only `UserID`+`RememberMe`,
   `StartLogin(userID, rememberMe)` takes no IP, and `CompleteTwoFactor` compares no IP. IP is used for
   the login rate-limit key, not for pending-token binding. Risk is low — the id is 256-bit and
   completion still needs the TOTP/recovery proof — so this is defense-in-depth that is absent, not a
   hole. Either add the binding (capture IP in `StartLogin`, store it, compare on `Take`) or correct the
   report/plan so the claim matches the code. (Flagging per "don't mirror a peer's framing without
   checking it against the code.")
2. **Rate limiter fails open, and the same limiter guards 2FA completion and recovery codes.**
   `RateLimiter.Allow` returns `true` on any Redis error (`redis/rate_limiter.go:35-47`), logging
   `"rate limiter unavailable; allowing request"`. For login this is a common availability trade-off,
   but the same limiter also gates `CompleteTwoFactor` and recovery-code attempts, so during a Redis
   outage the only brute-force protection on a 6-digit TOTP (10^6) and a 40-bit recovery code is gone.
   Record it as a named limitation in §5; consider failing **closed** specifically on the 2FA-completion
   key (the password step already passed, so the availability cost is small).
3. **`style-src 'unsafe-inline'`.** Scripts are nonce-locked, but inline styles are allowed, which leaves
   a style-injection surface (far milder than script). Common trade-off; note it, and move styles to
   nonce/hash if practical.
4. **Report-only e2e cannot catch an enforcing-CSP break.** Because e2e runs report-only, an inline
   script emitted without the nonce would pass e2e yet break production under enforcing CSP. The wiring
   via `ssr: { nonce }` is correct per the framework contract, but this class of break is invisible to
   the suite — confirm once, manually, that the real app loads under enforcing CSP with zero console
   violations (if the phase already did this, cite it in the report).
5. **No CSP `report-uri`/`report-to`.** Report-only mode surfaces violations only to the browser
   console (fine for e2e, but no telemetry). Optional.

## Note on review method

Three verifier subagents launched earlier this session died on the model safeguard
(`reasoning_extraction`); this review was done inline. The approval is scoped as stated above — a
targeted security read of the named surface, not a full audit of 9.6k lines.
