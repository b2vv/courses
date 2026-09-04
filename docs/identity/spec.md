# Identity Phase 3 specification

Статус: proposal before implementation. Foundation report: [../foundation/report.md](../foundation/report.md).

## Outcome

A student, instructor or administrator can sign in with email/password, keep a time-bounded server-side session in an HttpOnly cookie, read the current identity, sign out, and enter only routes allowed by the role. The Go API owns authentication policy; frontend guards improve navigation but are not the security boundary.

## Scope

- Domain: `User`, `Role` (`student`, `instructor`, `admin`), `Session`, explicit identity errors and role policy.
- PostgreSQL users/sessions migration and repositories; no ORM models in domain.
- Argon2id credential verifier behind an application port. Password hashes never leave the credential adapter. Raw session tokens never enter domain or persistence models; the session adapter returns one through an explicitly secret-bearing application result solely so the HTTP adapter can set the cookie.
- `POST /api/v1/auth/login`, `GET /api/v1/auth/me`, `POST /api/v1/auth/logout` in OpenAPI and Orval.
- Authentication middleware used by `/auth/me` and a domain role policy. Phase 4 adds the handler-level authorization wrapper when the first protected admin business endpoint provides a real caller.
- Frontend identity domain/application/port/API adapter, Query-owned current session, React Hook Form login, SSR cookie forwarding and route guards.
- Typed startup configuration owns `APP_ENV`, `PUBLIC_ORIGIN` and cookie security validation; Compose and `.env.example` provide development values.
- Local-only demo identities loaded by an explicit idempotent seed. Public fixture credentials are documentation, not deployment secrets.

## Out of scope

Registration, email verification, password reset, MFA/passkeys, OAuth, user administration, enrollment, course CRUD, API keys, refresh tokens and device management. Phase 3 does not expose protected business data yet.

## Security contract

- Login JSON is limited to 8 KiB. Email is trimmed/lowercased, must be 3–254 UTF-8 bytes, contain exactly one `@` with non-empty sides, and contain no whitespace/control characters. Password is 1–1024 bytes. A syntactically invalid, unknown or wrong email/password pair returns the same `invalid_credentials` response. Malformed JSON, missing/wrong field types, unknown or duplicate fields and values beyond those transport limits return `invalid_request`. Stored display names are trimmed and contain 1–120 Unicode code points.
- Argon2id PHC strings use fixed supported parameters and bounded decoding. Unknown users execute exactly one dummy Argon2id verification with the same configured parameters as a stored credential; tests assert that call path rather than make an unreliable wall-clock timing claim.
- Raw session tokens contain 256 bits from `crypto/rand` and use 43-character unpadded base64url encoding; PostgreSQL stores only unique 32-byte SHA-256 token digests. Session creation and revocation of the token currently presented by the browser commit in one PostgreSQL transaction before `Set-Cookie` is written. A digest collision rolls back that whole attempt and retries with a fresh transaction and token at most three times, then fails closed. An insertion failure leaves the prior session intact. A commit result can be indeterminate after a connection loss; the API returns 503 and emits no new cookie, so no unconfirmed session grants browser access. Successful login always creates a new token and revokes any valid token already presented by that browser, including a token for a different account; an attacker-chosen invalid token is never persisted. Sessions have a 12-hour absolute expiry. `expires_at > now` is active and `expires_at == now` is expired, using one injected timestamp for the operation. Login-session device limits are deferred until there is a product requirement; playback concurrency is a separate Phase 6 policy.
- Cookie: `courseforge_session`, HttpOnly, SameSite=Lax, Path=/, Max-Age bounded by session expiry; `Secure=false` is rejected unless the configured origin hostname is loopback. Deletion uses the same Path/SameSite/Secure scope, Max-Age=-1 and an Expires value in the past.
- Logout is idempotent, including simultaneous requests for the same token. Missing/expired/revoked sessions are indistinguishable when the store is available. Expired matching rows are deleted during resolution; a cleanup/store failure returns safe 503. The handler always expires the browser cookie; if persistent revocation fails it returns 503 instead of claiming that a potentially copied token was revoked.
- Login/logout accept only JSON state-changing requests from the configured same origin. Duplicate credential keys, duplicate session cookies and missing, multiple, `null` or malformed Origin values fail closed before mutation. Origin comparison parses and canonicalizes scheme, hostname and effective port; configured origins cannot contain credentials, paths beyond `/`, queries or fragments. Nginx applies a login-specific limit stricter than the normal API limit, with separate smoke evidence for both thresholds. Logs and traces never contain passwords, hashes, cookies or session tokens.
- `/auth/me` uses backend authentication middleware. The role policy is tested now; the transport authorization wrapper is deferred to the first protected admin business endpoint in Phase 4. Frontend guards never grant API access.

## HTTP contract

- Successful login: `200` user DTO plus `Set-Cookie`.
- Invalid login: `401 invalid_credentials`; malformed transport input: `400 invalid_request`; unavailable store: `503 dependency_unavailable`.
- Current user: `200` user DTO; missing, malformed, expired or revoked token with an available store: `401 authentication_required`; missing/malformed/duplicate tokens do not cause a database lookup; a valid-format token with an unavailable store or failed expiry cleanup: `503 dependency_unavailable`.
- Logout: `204` and an expired cookie whether a current token exists or not; persistent-store failure returns `503` with the cookie still expired.
- User DTO contains `id`, `email`, `displayName`, `role`; never password/session material.
- Identity responses set `Cache-Control: no-store`; identity-aware SSR responses vary on `Cookie`, while localized HTML continues to vary on the locale inputs already supported by the frontend.

## Localization

Visible login validation/errors, authenticated navigation, logout state and guard messages have explicit Ukrainian and English resources in the existing request-scoped `i18next` setup. Backend error codes/messages remain stable transport data and are mapped to localized UI descriptors. Logout removes private Query data before navigation so browser history cannot render stale identity.

## Rendering and route policy

- `/`, `/courses`, `/courses/$slug` and `/login` remain SSR routes. Their shared identity-aware header resolves `/auth/me` on the server with the incoming cookie and hydrates the request-scoped Query cache; raw HTML tests assert the signed-in shell and `Vary: Cookie`.
- `/learn` and `/admin/*` remain client-only as required by the foundation architecture. Client guard tests separately assert anonymous redirect on 401, a localized retryable error without redirect on 503, authenticated access, the not-found answer a student receives for `/admin`, instructor/admin shell access and absence of a post-hydration identity flash.
- `/learn` accepts every authenticated role. `/admin/*` accepts `instructor` and `admin`. Every visitor who is not permitted — a `student`, and an anonymous visitor who signs in as one — receives the ordinary not-found answer rather than a forbidden state, so the administration area is not advertised to casual address probing. This is a deliberate amendment made during I06; the client guard is not a security boundary either way, and the server-side boundary arrives in Phase 4.
- `returnTo` accepts only normalized internal protected paths. External URLs, protocol-relative URLs, encoded or backslash variants, control characters, recursive `/login` targets and non-protected paths fall back to `/learn`.
