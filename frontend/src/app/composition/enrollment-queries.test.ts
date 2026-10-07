import { expect, test } from '@rstest/core';
import { QueryClient } from '@tanstack/react-query';
import {
  createEnrollmentQueries,
  enrollmentMembershipQueryKey,
  EnrollmentQueryError,
} from './enrollment-queries';

test('membership conflict invalidates the exact membership state for a fresh CTA read', async () => {
  const client = new QueryClient();
  const queries = createEnrollmentQueries({
    membership: async () => ({ ok: true, value: { member: false, state: null } }),
    enroll: async () => ({ ok: false, error: { code: 'membership_conflict' } }),
    myMemberships: async () => ({ ok: true, value: [] }),
  });
  const courseId = '10000000-0000-4000-8000-000000000001';
  const key = enrollmentMembershipQueryKey(courseId);
  client.setQueryData(key, { member: false, state: null });
  await queries
    .enroll(client)
    .onError?.(new EnrollmentQueryError('membership_conflict'), courseId, undefined, {
      client,
      meta: undefined,
    });
  expect(client.getQueryState(key)?.isInvalidated).toBe(true);
});
