import { expect, test } from '@rstest/core';
import {
  QueryClient,
  QueryClientProvider,
  useInfiniteQuery,
  useMutation,
} from '@tanstack/react-query';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import type { TeachRepository } from '../../modules/teach';
import { createTeachQueries } from './teach-queries';

const courseId = '10000000-0000-4000-8000-000000000001';
const commentId = '40000000-0000-4000-8000-000000000001';
const lessonId = '70000000-0000-4000-8000-000000000001';
const page = { comments: [], nextCursor: null };

test('W01/W04/W05/W07/W08 loads pages and invalidates teach and learn after moderation', async () => {
  let dismissed = 0;
  let removed = 0;
  let commentReads = 0;
  let queueReads = 0;
  const repository: TeachRepository = {
    listLearners: async () => ({ ok: true, value: { entries: [], nextCursor: null } }),
    listCourseComments: async () => {
      commentReads += 1;
      return { ok: true, value: page };
    },
    listModerationQueue: async () => {
      queueReads += 1;
      return { ok: true, value: { items: [], nextCursor: null } };
    },
    blockAuthor: async () => ({ ok: true, value: undefined }),
    unblockAuthor: async () => ({ ok: true, value: undefined }),
    dismissReport: async () => {
      dismissed += 1;
      return { ok: true, value: undefined };
    },
    removeComment: async () => {
      removed += 1;
      return { ok: true, value: undefined };
    },
    dismissReviewReport: async () => ({ ok: true, value: undefined }),
    removeReview: async () => ({ ok: true, value: undefined }),
  };
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  const learnKey = ['learn', 'discussion', lessonId] as const;
  client.setQueryData(learnKey, { threads: [] });
  const queries = createTeachQueries(repository);
  function Probe() {
    const comments = useInfiniteQuery(queries.courseComments(courseId));
    const queue = useInfiniteQuery(queries.moderationQueue(courseId));
    const dismiss = useMutation(queries.dismiss(client));
    const remove = useMutation(queries.remove(client));
    return (
      <>
        <output>{comments.data?.pages[0]?.comments.length ?? 'loading'}</output>
        <output>{queue.data?.pages[0]?.items.length ?? 'loading'}</output>
        <button onClick={() => dismiss.mutate({ commentId, input: { reason: 'reviewed' } })}>
          dismiss
        </button>
        <button onClick={() => remove.mutate({ lessonId, commentId })}>remove</button>
      </>
    );
  }
  render(
    <QueryClientProvider client={client}>
      <Probe />
    </QueryClientProvider>,
  );
  await waitFor(() => expect(screen.getAllByText('0')).toHaveLength(2));
  await userEvent.click(screen.getByRole('button', { name: 'remove' }));
  await waitFor(() => expect(removed).toBe(1));
  await waitFor(() => expect(commentReads).toBeGreaterThanOrEqual(2));
  await waitFor(() => expect(queueReads).toBeGreaterThanOrEqual(2));
  expect(client.getQueryState(learnKey)?.isInvalidated).toBe(true);
  await userEvent.click(screen.getByRole('button', { name: 'dismiss' }));
  await waitFor(() => expect(dismissed).toBe(1));
  await waitFor(() => expect(commentReads).toBeGreaterThanOrEqual(3));
  await waitFor(() => expect(queueReads).toBeGreaterThanOrEqual(3));
});
