import { describe, expect, test } from '@rstest/core';
import {
  makeAdminUserQuery,
  validAdministrationPassword,
  validAdminUserId,
} from './administration';

describe('identity administration domain', () => {
  test('normalizes and bounds a directory query', () => {
    expect(makeAdminUserQuery({ q: '  Student  ', status: 'blocked' })).toEqual({
      ok: true,
      value: { q: 'Student', status: 'blocked', limit: 20, offset: 0 },
    });
    expect(makeAdminUserQuery({ q: 'я'.repeat(101) }).ok).toBe(false);
    expect(makeAdminUserQuery({ q: '\u0000' }).ok).toBe(false);
    expect(makeAdminUserQuery({ limit: 51 }).ok).toBe(false);
    expect(makeAdminUserQuery({ offset: -1 }).ok).toBe(false);
  });

  test('uses the server UTF-8 byte bound for administration passwords', () => {
    expect(validAdministrationPassword('')).toBe(false);
    expect(validAdministrationPassword('x'.repeat(1024))).toBe(true);
    expect(validAdministrationPassword('я'.repeat(512))).toBe(true);
    expect(validAdministrationPassword('я'.repeat(513))).toBe(false);
  });

  test('accepts only non-zero lowercase UUID-shaped user ids', () => {
    expect(validAdminUserId('40000000-0000-4000-8000-000000000001')).toBe(true);
    expect(validAdminUserId('40000000-0000-4000-8000-00000000000A')).toBe(false);
    expect(validAdminUserId('00000000-0000-0000-0000-000000000000')).toBe(false);
    expect(validAdminUserId('not-a-uuid')).toBe(false);
  });
});
