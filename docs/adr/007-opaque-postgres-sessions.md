# ADR 007: Opaque PostgreSQL sessions

Status: accepted for Identity Phase 3.

Use revocable server-side sessions backed by PostgreSQL. The browser receives a 256-bit random base64url token in an HttpOnly SameSite cookie; persistence receives only its unique SHA-256 digest. Authentication joins the active user on every request, so role/disabled-state changes apply without waiting for a signed token to expire.

This reuses the transactional database already required by the modular monolith. Redis sessions would add a second consistency/backup model; JWT access/refresh tokens would add revocation, rotation and key-management protocols that Phase 3 does not need. Login-session device limits are deferred until they have a product requirement; playback concurrency is a separate Phase 6 policy.

Sessions expire absolutely after 12 hours. Login creates the new session and revokes any valid token already presented by that browser in one committed transaction before emitting a cookie. A unique-digest collision retries the whole transaction with a fresh token at most three times. Logout deletes the digest and expires the cookie. Production rollback revokes all session rows before restoring foundation binaries so a later rollout cannot reactivate old cookies.

Consequences: each authenticated request performs a bounded indexed database lookup. PostgreSQL outages fail closed. Horizontal API replicas need no sticky routing. Future device management, rolling renewal or Redis caching requires a separate use case and must preserve revocation semantics.
