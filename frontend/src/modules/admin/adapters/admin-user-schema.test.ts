import { describe, expect, test } from '@rstest/core';
import { adminUserPageSchema, adminUserSchema } from './admin-user-schema';

const user = {
  id: '40000000-0000-4000-8000-000000000001',
  email: 'student@courseforge.local',
  displayName: 'Student',
  role: 'student',
  status: 'active',
  updatedAt: '2026-09-01T00:00:00.000001Z',
};

describe('admin user response schemas', () => {
  test('accepts a strict user page with plain string ids', () => {
    const parsed = adminUserPageSchema.safeParse({
      items: [user],
      pagination: { limit: 20, offset: 0, total: 1 },
    });
    expect(parsed.success).toBe(true);
    if (parsed.success) expect(parsed.data.items[0]?.id).toBe(user.id);
  });

  test.each([
    ['a malformed id', { ...user, id: 'not-a-uuid' }],
    ['a zero id', { ...user, id: '00000000-0000-0000-0000-000000000000' }],
    ['an uppercase id', { ...user, id: '40000000-0000-4000-8000-00000000000A' }],
    ['an unnormalized email', { ...user, email: ' Student@courseforge.local ' }],
    ['a blank display name', { ...user, displayName: ' ' }],
    ['an unknown role', { ...user, role: 'owner' }],
    ['an unknown status', { ...user, status: 'disabled' }],
    ['a timestamp without microseconds', { ...user, updatedAt: '2026-09-01T00:00:00Z' }],
    ['an unknown member', { ...user, disabledAt: null }],
  ])('refuses %s', (_name, malformed) => {
    expect(adminUserSchema.safeParse(malformed).success).toBe(false);
  });

  test('refuses malformed pagination and pages over the response bound', () => {
    expect(
      adminUserPageSchema.safeParse({
        items: [user],
        pagination: { limit: 0, offset: 0, total: 1 },
      }).success,
    ).toBe(false);
    expect(
      adminUserPageSchema.safeParse({
        items: Array.from({ length: 51 }, () => user),
        pagination: { limit: 50, offset: 0, total: 51 },
      }).success,
    ).toBe(false);
  });
});
