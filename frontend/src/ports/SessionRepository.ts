import type { AuthenticatedUser, Credentials, SessionResult } from '../domain/identity/identity';

export interface SessionRepository {
  /** Browser-only: the server requires an Origin header a server-side call cannot send. */
  logIn(credentials: Credentials, signal?: AbortSignal): Promise<SessionResult<AuthenticatedUser>>;
  /** Anonymous is a success carrying null, not an error; only transport faults fail. */
  currentUser(signal?: AbortSignal): Promise<SessionResult<AuthenticatedUser | null>>;
  /** Browser-only, and idempotent: revoking an absent session still succeeds. */
  logOut(signal?: AbortSignal): Promise<SessionResult<null>>;
}
