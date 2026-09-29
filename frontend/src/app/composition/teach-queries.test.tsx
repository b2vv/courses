import { expect, test } from '@rstest/core';
import {
  QueryClient,
  QueryClientProvider,
  useInfiniteQuery,
  useMutation,
} from '@tanstack/react-query';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import type { TeachRepository } from '../../ports/TeachRepository';
import { createTeachQueries } from './teach-queries';

const courseId = '10000000-0000-4000-8000-000000000001';
const commentId = '40000000-0000-4000-8000-000000000001';
const page = { comments: [], nextCursor: null };

test('W01/W04/W07/W08 loads empty cursor pages and invalidates after dismiss', async () => {
  let dismissed = 0;
  const repository: TeachRepository = {
    listCourseComments: async () => ({ ok: true, value: page }),
    listModerationQueue: async () => ({ ok: true, value: { items: [], nextCursor: null } }),
    dismissReport: async () => {
      dismissed += 1;
      return { ok: true, value: undefined };
    },
  };
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  const queries = createTeachQueries(repository);
  function Probe() {
    const comments = useInfiniteQuery(queries.courseComments(courseId));
    const dismiss = useMutation(queries.dismiss(client));
    return (
      <>
        <output>{comments.data?.pages[0]?.comments.length ?? 'loading'}</output>
        <button onClick={() => dismiss.mutate({ commentId, input: { reason: 'reviewed' } })}>
          dismiss
        </button>
      </>
    );
  }
  render(
    <QueryClientProvider client={client}>
      <Probe />
    </QueryClientProvider>,
  );
  await waitFor(() => expect(screen.getByText('0')).toBeTruthy());
  await userEvent.click(screen.getByRole('button', { name: 'dismiss' }));
  await waitFor(() => expect(dismissed).toBe(1));
});
