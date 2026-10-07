import { expect, test } from '@rstest/core';
import { QueryClient } from '@tanstack/react-query';
import {
  createEnrollmentQueries,
  enrollmentMembershipQueryKey,
  enrollmentRosterQueryKey,
  EnrollmentQueryError,
} from './enrollment-queries';

test('membership conflict invalidates the exact membership state for a fresh CTA read', async () => {
  const client = new QueryClient();
  const queries = createEnrollmentQueries({
    membership: async () => ({ ok: true, value: { state: null } }),
    enroll: async () => ({ ok: false, error: { code: 'membership_conflict' } }),
    myMemberships: async () => ({ ok: true, value: [] }),
    roster: async () => ({ ok: true, value: { entries: [], nextCursor: null } }),
    block: async () => ({ ok: true, value: undefined }),
    unblock: async () => ({ ok: true, value: undefined }),
    reportLecturer: async () => ({ ok: true, value: undefined }),
    reportStudent: async () => ({ ok: true, value: undefined }),
  });
  const courseId = '10000000-0000-4000-8000-000000000001';
  const key = enrollmentMembershipQueryKey(courseId);
  client.setQueryData(key, { state: null });
  await queries
    .enroll(client)
    .onError?.(new EnrollmentQueryError('membership_conflict'), courseId, undefined, {
      client,
      meta: undefined,
    });
  expect(client.getQueryState(key)?.isInvalidated).toBe(true);
});

test('roster mutations invalidate only their course roster', async () => {
  const client = new QueryClient();
  const queries = createEnrollmentQueries({
    membership: async () => ({ ok: true, value: { state: null } }),
    enroll: async () => ({ ok: true, value: { state: 'active' } }),
    myMemberships: async () => ({ ok: true, value: [] }),
    roster: async () => ({ ok: true, value: { entries: [], nextCursor: null } }),
    block: async () => ({ ok: true, value: undefined }),
    unblock: async () => ({ ok: true, value: undefined }),
    reportLecturer: async () => ({ ok: true, value: undefined }),
    reportStudent: async () => ({ ok: true, value: undefined }),
  });
  const courseId = '10000000-0000-4000-8000-000000000001';
  const otherCourseId = '20000000-0000-4000-8000-000000000002';
  const rosterKey = enrollmentRosterQueryKey(courseId);
  const otherRosterKey = enrollmentRosterQueryKey(otherCourseId);
  client.setQueryData(rosterKey, { pages: [], pageParams: [] });
  client.setQueryData(otherRosterKey, { pages: [], pageParams: [] });
  await queries.block(client).onSuccess?.(undefined, { courseId, userId: '30000000-0000-4000-8000-000000000003' }, undefined, {
    client,
    meta: undefined,
  });
  expect(client.getQueryState(rosterKey)?.isInvalidated).toBe(true);
  expect(client.getQueryState(otherRosterKey)?.isInvalidated).toBe(false);
});

test('block/unblock refresh the roster on a membership_conflict 409 but not on other errors (B-L7b.2)', async () => {
  const client = new QueryClient();
  const queries = createEnrollmentQueries({
    membership: async () => ({ ok: true, value: { state: null } }),
    enroll: async () => ({ ok: true, value: { state: 'active' } }),
    myMemberships: async () => ({ ok: true, value: [] }),
    roster: async () => ({ ok: true, value: { entries: [], nextCursor: null } }),
    block: async () => ({ ok: true, value: undefined }),
    unblock: async () => ({ ok: true, value: undefined }),
    reportLecturer: async () => ({ ok: true, value: undefined }),
    reportStudent: async () => ({ ok: true, value: undefined }),
  });
  const courseId = '10000000-0000-4000-8000-000000000001';
  const userId = '30000000-0000-4000-8000-000000000003';
  const rosterKey = enrollmentRosterQueryKey(courseId);
  const context = { client, meta: undefined };

  client.setQueryData(rosterKey, { pages: [], pageParams: [] });
  await queries.block(client).onError?.(new EnrollmentQueryError('membership_conflict'), { courseId, userId }, undefined, context);
  expect(client.getQueryState(rosterKey)?.isInvalidated).toBe(true);

  // A non-conflict error leaves the roster untouched (no spurious refetch).
  client.setQueryData(rosterKey, { pages: [], pageParams: [] });
  await queries.unblock(client).onError?.(new EnrollmentQueryError('enrollment_forbidden'), { courseId, userId }, undefined, context);
  expect(client.getQueryState(rosterKey)?.isInvalidated).toBe(false);
});

test('report mutations do not invalidate visible enrollment state', async () => {
  const client = new QueryClient();
  const queries = createEnrollmentQueries({
    membership: async () => ({ ok: true, value: { state: null } }),
    enroll: async () => ({ ok: true, value: { state: 'active' } }),
    myMemberships: async () => ({ ok: true, value: [] }),
    roster: async () => ({ ok: true, value: { entries: [], nextCursor: null } }),
    block: async () => ({ ok: true, value: undefined }),
    unblock: async () => ({ ok: true, value: undefined }),
    reportLecturer: async () => ({ ok: true, value: undefined }),
    reportStudent: async () => ({ ok: true, value: undefined }),
  });
  const courseId = '10000000-0000-4000-8000-000000000001';
  const key = enrollmentMembershipQueryKey(courseId);
  client.setQueryData(key, { state: 'active' });
  await queries.reportLecturer().onSuccess?.(undefined, { courseId, reason: 'reason' }, undefined, {
    client,
    meta: undefined,
  });
  expect(client.getQueryState(key)?.isInvalidated).toBe(false);
});
