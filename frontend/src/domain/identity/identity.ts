export type Role = 'student' | 'instructor' | 'admin';
export type AuthenticatedUser = {
  readonly id: string;
  readonly email: string;
  readonly displayName: string;
  readonly role: Role;
};
export type SessionErrorCode =
  | 'invalid_credentials'
  | 'invalid_request'
  | 'origin_forbidden'
  | 'authentication_required'
  | 'dependency_unavailable'
  | 'invalid_response'
  | 'request_timeout'
  | 'request_cancelled';
export type SessionError = { readonly code: SessionErrorCode };
export type SessionResult<T> =
  | { readonly ok: true; readonly value: T }
  | { readonly ok: false; readonly error: SessionError };
export type Credentials = { readonly email: string; readonly password: string };
export type CredentialsInput = { readonly email: string; readonly password: string };

const roles: readonly Role[] = ['student', 'instructor', 'admin'];

export function validRole(value: string): value is Role {
  return (roles as readonly string[]).includes(value);
}

/**
 * Every role may learn; only instructors and administrators reach the
 * administration shell. This mirrors the server capability policy, which stays
 * the authority: the client copy only decides what to render.
 */
export function canLearn(user: AuthenticatedUser): boolean {
  return validRole(user.role);
}

export function canAccessAdministration(user: AuthenticatedUser): boolean {
  return user.role === 'instructor' || user.role === 'admin';
}

/**
 * A syntactically impossible email fails as `invalid_credentials` rather than as
 * its own code, so the shape of the refusal never reveals whether an address
 * exists.
 */
export function makeCredentials(input: CredentialsInput): SessionResult<Credentials> {
  const email = input.email.trim().toLowerCase();
  if (
    email.length < 3 ||
    email.length > 254 ||
    !/^[^@\s]+@[^@\s]+$/.test(email) ||
    input.password.length < 1 ||
    input.password.length > 1024
  )
    return { ok: false, error: { code: 'invalid_credentials' } };
  return { ok: true, value: { email, password: input.password } };
}

/**
 * A return target is only ever a normalized protected path inside this
 * application. Anything else falls back to the default.
 */
export function safeReturnTo(value: string | undefined, fallback = '/learn'): string {
  if (value === undefined || value === '') return fallback;
  if (!value.startsWith('/')) return fallback;
  const tail = value.slice(1, 4).toLowerCase();
  if (tail.startsWith('/') || tail.startsWith('\\') || tail.startsWith('%2f')) return fallback;
  // A control character in a redirect target is never legitimate and can split
  // a header; checking code points avoids embedding them in a pattern.
  for (const character of value) {
    const code = character.codePointAt(0) ?? 0;
    if (code < 0x20 || code === 0x7f) return fallback;
  }
  const searchIndex = value.indexOf('?');
  let path = searchIndex === -1 ? value : value.slice(0, searchIndex);
  const search = searchIndex === -1 ? '' : value.slice(searchIndex);
  if (path.endsWith('/')) path = path.slice(0, -1);
  if (!/^\/(learn|admin)(\/|$)/.test(path)) return fallback;
  if (path.split('/').some((segment) => segment === '.' || segment === '..')) return fallback;
  return path + (search === '?' ? '' : search);
}
