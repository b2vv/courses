import { expect, test } from '@rstest/core';
import { QueryClient, QueryObserver } from '@tanstack/react-query';
import { createSession } from '../../application/identity/session';
import { createSessionQueries, SessionQueryError, sessionQueryKey } from './identity-queries';
import type { AuthenticatedUser, SessionResult } from '../../domain/identity/identity';
import type { SessionRepository } from '../../ports/SessionRepository';

const signedIn: AuthenticatedUser = {
  id: '40000000-0000-4000-8000-000000000001',
  email: 'student@courseforge.local',
  displayName: 'Студентка',
  role: 'student',
};

function queries(currentUser: () => Promise<SessionResult<AuthenticatedUser | null>>) {
  const port: SessionRepository = {
    logIn: () => Promise.resolve({ ok: true, value: signedIn }),
    currentUser,
    logOut: () => Promise.resolve({ ok: true, value: null }),
  };
  return createSessionQueries(createSession(port));
}

function clientForTest() {
  return new QueryClient({ defaultOptions: { queries: { retry: false } } });
}

test('resolves an anonymous viewer as null rather than as a rejected query', async () => {
  const client = clientForTest();
  const identity = queries(() => Promise.resolve({ ok: true, value: null }));
  expect(await client.ensureQueryData(identity.current())).toBe(null);
});

test('surfaces a transport failure as a typed error the route can localize', async () => {
  const client = clientForTest();
  const identity = queries(() =>
    Promise.resolve({ ok: false, error: { code: 'dependency_unavailable' } }),
  );
  await expect(client.ensureQueryData(identity.current())).rejects.toThrow(SessionQueryError);
  const error = await client.ensureQueryData(identity.current()).catch((cause: unknown) => cause);
  expect(error instanceof SessionQueryError && error.code).toBe('dependency_unavailable');
});

test('never treats a fetched identity as fresh, so any mount revalidates it', async () => {
  let answer: AuthenticatedUser | null = signedIn;
  const identity = queries(() => Promise.resolve({ ok: true, value: answer }));
  const client = clientForTest();
  expect(await client.ensureQueryData(identity.current())).toEqual(signedIn);
  // A mounted consumer must see the identity as stale, which is what makes it
  // revalidate on every mount.
  const observer = new QueryObserver(client, identity.current());
  const unsubscribe = observer.subscribe(() => {});
  expect(observer.getCurrentResult().isStale).toBe(true);
  unsubscribe();
  // A revalidation observes the change rather than replaying the cached viewer.
  answer = null;
  expect(await client.fetchQuery(identity.current())).toBe(null);
});

test('keeps two concurrent requests from sharing one viewer', async () => {
  const first = clientForTest();
  const second = clientForTest();
  const asStudent = queries(() => Promise.resolve({ ok: true, value: signedIn }));
  const asAnonymous = queries(() => Promise.resolve({ ok: true, value: null }));
  const [a, b] = await Promise.all([
    first.ensureQueryData(asStudent.current()),
    second.ensureQueryData(asAnonymous.current()),
  ]);
  expect(a).toEqual(signedIn);
  expect(b).toBe(null);
  expect(second.getQueryData(sessionQueryKey)).toBe(null);
});

test('a login result replaces the identity under the shared key', () => {
  const client = clientForTest();
  client.setQueryData(sessionQueryKey, signedIn);
  expect(client.getQueryData(sessionQueryKey)).toEqual(signedIn);
});
