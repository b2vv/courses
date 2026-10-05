import { describe, expect, it } from '@rstest/core';
import { createSession } from './session';
import type { AuthenticatedUser, Credentials, SessionResult } from '../../domain/identity/identity';
import type { SessionRepository } from '../../ports/SessionRepository';

const signedIn: AuthenticatedUser = {
  id: '40000000-0000-4000-8000-000000000001',
  email: 'student@courseforge.local',
  displayName: 'Студентка',
  emailVerified: true,
  role: 'student',
};

function repository(overrides: Partial<SessionRepository> = {}) {
  const calls: { logIn: Credentials[]; currentUser: number; logOut: number } = {
    logIn: [],
    currentUser: 0,
    logOut: 0,
  };
  const base: SessionRepository = {
    logIn(credentials) {
      calls.logIn.push(credentials);
      return Promise.resolve({
        ok: true,
        value: signedIn,
      } satisfies SessionResult<AuthenticatedUser>);
    },
    currentUser() {
      calls.currentUser += 1;
      return Promise.resolve({ ok: true, value: signedIn });
    },
    logOut() {
      calls.logOut += 1;
      return Promise.resolve({ ok: true, value: null });
    },
    register: () => Promise.resolve({ ok: true, value: signedIn }),
    verifyEmail: () => Promise.resolve({ ok: true, value: null }),
    resendVerification: () => Promise.resolve({ ok: true, value: null }),
    forgotPassword: () => Promise.resolve({ ok: true, value: null }),
    resetPassword: () => Promise.resolve({ ok: true, value: null }),
    changePassword: () => Promise.resolve({ ok: true, value: null }),
    requestEmailChange: () => Promise.resolve({ ok: true, value: null }),
    confirmEmailChange: () => Promise.resolve({ ok: true, value: null }),
  };
  return { calls, port: { ...base, ...overrides } };
}

describe('session application', () => {
  it('normalizes credentials before they reach the port', async () => {
    const { calls, port } = repository();
    await createSession(port).logIn({ email: '  A@B.CO ', password: 'secret' });
    expect(calls.logIn).toEqual([{ email: 'a@b.co', password: 'secret' }]);
  });

  it('refuses an impossible email without calling the port at all', async () => {
    const { calls, port } = repository();
    const result = await createSession(port).logIn({ email: 'nope', password: 'secret' });
    expect(result.ok).toBe(false);
    if (!result.ok) expect(result.error.code).toBe('invalid_credentials');
    expect(calls.logIn).toEqual([]);
  });

  it('passes a caller signal through to the port', async () => {
    const seen: (AbortSignal | undefined)[] = [];
    const { port } = repository({
      currentUser(signal) {
        seen.push(signal);
        return Promise.resolve({ ok: true, value: null });
      },
    });
    const controller = new AbortController();
    await createSession(port).current(controller.signal);
    expect(seen).toEqual([controller.signal]);
  });

  it('reports an anonymous viewer as success carrying null', async () => {
    const { port } = repository({ currentUser: () => Promise.resolve({ ok: true, value: null }) });
    const result = await createSession(port).current();
    expect(result).toEqual({ ok: true, value: null });
  });

  it('propagates a transport failure without reinterpreting it', async () => {
    const { port } = repository({
      logIn: () => Promise.resolve({ ok: false, error: { code: 'dependency_unavailable' } }),
    });
    const result = await createSession(port).logIn({ email: 'a@b.co', password: 'secret' });
    expect(result.ok).toBe(false);
    if (!result.ok) expect(result.error.code).toBe('dependency_unavailable');
  });

  it('exposes self-service commands through the application seam', async () => {
    const seen: string[] = [];
    const { port } = repository({
      register: () => {
        seen.push('register');
        return Promise.resolve({ ok: true, value: signedIn });
      },
      verifyEmail: () => {
        seen.push('verify');
        return Promise.resolve({ ok: true, value: null });
      },
      confirmEmailChange: () => {
        seen.push('confirm-email');
        return Promise.resolve({ ok: true, value: null });
      },
    });
    const session = createSession(port);
    await session.register({
      email: 'new@courseforge.local',
      password: 'password1',
      displayName: 'New',
    });
    await session.verify({ token: 'verification-token' });
    await session.confirmEmailChange({ token: 'email-token' });
    expect(seen).toEqual(['register', 'verify', 'confirm-email']);
  });
});
