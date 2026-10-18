import { expect, test } from '@rstest/core';
import {
  QueryClient,
  QueryClientProvider,
  useInfiniteQuery,
  useMutation,
} from '@tanstack/react-query';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import type { DiscussionComment } from '../../modules/discussion';
import type { DiscussionRepository } from '../../modules/discussion';
import { createDiscussionQueries } from './discussion-queries';

const lessonId = '30000000-0000-4000-8000-000000000001';

function comment(id: string, body: string): DiscussionComment {
  return {
    id,
    author: { displayName: 'Ada' },
    body,
    positionMs: null,
    createdAt: '2026-09-28T10:00:00Z',
    editedAt: null,
    removed: false,
    canEdit: true,
    canRemove: true,
  };
}

test('refetching loaded discussion pages after publish has no duplicates and shows the new root first', async () => {
  const first = comment('40000000-0000-4000-8000-000000000001', 'first');
  const second = comment('40000000-0000-4000-8000-000000000002', 'second');
  const fresh = comment('40000000-0000-4000-8000-000000000003', 'new');
  let published = false;
  const repository: DiscussionRepository = {
    list: async (_lesson, cursor) => {
      if (cursor === undefined) {
        return {
          ok: true,
          value: {
            threads: [{ root: published ? fresh : first, replies: [] }],
            nextCursor: 'cursor-2',
          },
        };
      }
      return {
        ok: true,
        value: {
          threads: [{ root: published ? first : second, replies: [] }],
          nextCursor: published ? 'cursor-3' : null,
        },
      };
    },
    publish: async () => {
      published = true;
      return { ok: true, value: fresh };
    },
    update: async () => ({ ok: true, value: first }),
    remove: async () => ({ ok: true, value: undefined }),
  };
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  const queries = createDiscussionQueries(repository);

  function Probe() {
    const discussion = useInfiniteQuery(queries.lesson(lessonId));
    const publish = useMutation(queries.publish(client));
    const ids =
      discussion.data?.pages.flatMap((page) => page.threads.map((thread) => thread.root.id)) ?? [];
    return (
      <>
        <output data-testid="ids">{ids.join(',')}</output>
        {discussion.hasNextPage ? (
          <button type="button" onClick={() => void discussion.fetchNextPage()}>
            next
          </button>
        ) : null}
        <button type="button" onClick={() => publish.mutate({ lessonId, input: { body: 'new' } })}>
          publish
        </button>
      </>
    );
  }

  render(
    <QueryClientProvider client={client}>
      <Probe />
    </QueryClientProvider>,
  );
  await waitFor(() => expect(screen.getByTestId('ids').textContent).toContain(first.id));
  await userEvent.click(screen.getByRole('button', { name: 'next' }));
  await waitFor(() => expect(screen.getByTestId('ids').textContent).toContain(second.id));

  await userEvent.click(screen.getByRole('button', { name: 'publish' }));
  await waitFor(() =>
    expect(screen.getByTestId('ids').textContent?.startsWith(fresh.id)).toBe(true),
  );
  const ids = screen.getByTestId('ids').textContent?.split(',').filter(Boolean) ?? [];
  expect(new Set(ids).size).toBe(ids.length);
});
