import { describe, expect, it } from '@rstest/core';
import { canAccessAdministration, canLearn, makeCredentials, safeReturnTo } from './identity';
import type { AuthenticatedUser, Role } from './identity';

function user(role: Role): AuthenticatedUser {
  return {
    id: '40000000-0000-4000-8000-000000000001',
    email: 'a@b.co',
    displayName: 'A',
    emailVerified: true,
    twoFactorEnabled: false,
    role,
  };
}

describe('identity domain', () => {
  it('normalizes an email the same way the server does', () => {
    const result = makeCredentials({ email: '  Student@CourseForge.Local  ', password: 'secret' });
    expect(result.ok).toBe(true);
    if (result.ok) expect(result.value.email).toBe('student@courseforge.local');
  });

  it('keeps the password exactly as typed', () => {
    const password = '  p@ss word\twith spaces  ';
    const result = makeCredentials({ email: 'a@b.co', password });
    expect(result.ok).toBe(true);
    if (result.ok) expect(result.value.password).toBe(password);
  });

  it('preserves an explicit remember-me decision', () => {
    const result = makeCredentials({ email: 'a@b.co', password: 'secret', rememberMe: false });
    expect(result).toEqual({
      ok: true,
      value: { email: 'a@b.co', password: 'secret', rememberMe: false },
    });
  });

  it.each([
    ['missing at sign', 'no-at-sign.example', 'secret'],
    ['two at signs', 'a@b@c.co', 'secret'],
    ['whitespace inside', 'a b@c.co', 'secret'],
    ['too short', 'a@', 'secret'],
    ['too long', `${'a'.repeat(250)}@b.co`, 'secret'],
    ['empty password', 'a@b.co', ''],
    ['oversized password', 'a@b.co', 'x'.repeat(1025)],
  ])('refuses %s as invalid_credentials rather than a distinct code', (_name, email, password) => {
    const result = makeCredentials({ email, password });
    expect(result.ok).toBe(false);
    if (!result.ok) expect(result.error.code).toBe('invalid_credentials');
  });

  it('mirrors the server capability matrix', () => {
    expect(canLearn(user('student'))).toBe(true);
    expect(canLearn(user('instructor'))).toBe(true);
    expect(canLearn(user('admin'))).toBe(true);
    expect(canAccessAdministration(user('student'))).toBe(false);
    expect(canAccessAdministration(user('instructor'))).toBe(true);
    expect(canAccessAdministration(user('admin'))).toBe(true);
  });

  it.each([
    ['an absolute URL', 'https://evil.example/'],
    ['a scheme-relative URL', '//evil.example'],
    ['a backslash form', '/\\evil.example'],
    ['an encoded slash pair', '%2f%2fevil.example'],
    ['an encoded header split', '/admin%0d%0aSet-Cookie:x'],
    ['parent path segments', '/learn/../..'],
    ['current path segments', '/admin/./x'],
    ['a recursive login target', '/login'],
    ['a public path', '/courses'],
    ['an empty value', ''],
    ['an absent value', undefined],
  ])('refuses %s as a return target', (_name, value) => {
    expect(safeReturnTo(value)).toBe('/learn');
  });

  it('refuses a return target carrying a control character', () => {
    expect(safeReturnTo('/learn\r\nSet-Cookie: a=b')).toBe('/learn');
    expect(safeReturnTo(`/learn${String.fromCodePoint(0)}`)).toBe('/learn');
  });

  it.each([
    ['a deep protected path with search', '/learn/go/intro?lesson=3', '/learn/go/intro?lesson=3'],
    ['an administration path with a trailing slash', '/admin/', '/admin'],
    ['a protected path with empty search', '/learn?', '/learn'],
  ])('keeps %s', (_name, value, expected) => {
    expect(safeReturnTo(value)).toBe(expected);
  });

  it('falls back when no target was given', () => {
    expect(safeReturnTo(undefined)).toBe('/learn');
    expect(safeReturnTo(undefined, '/courses')).toBe('/courses');
  });
});
