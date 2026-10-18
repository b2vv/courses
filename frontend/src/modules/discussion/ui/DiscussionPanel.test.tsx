import { expect, test } from '@rstest/core';
import { render, screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { createI18n, I18nProvider } from '../../../app/providers/i18n';
import type { DiscussionPage } from '../domain/discussion';
import { DiscussionPanel } from './DiscussionPanel';

const page: DiscussionPage = {
  threads: [
    {
      root: {
        id: '40000000-0000-4000-8000-000000000001',
        author: { displayName: 'Ada' },
        body: 'Watch this part',
        positionMs: 12_345,
        createdAt: '2026-09-28T10:00:00Z',
        editedAt: null,
        removed: false,
        canEdit: false,
        canRemove: false,
      },
      replies: [],
    },
  ],
  nextCursor: null,
};
const rootComment = page.threads[0]?.root;
if (rootComment === undefined) throw new Error('Discussion fixture is missing its root comment');

type PanelProps = Parameters<typeof DiscussionPanel>[0];

function renderPanel(overrides: Partial<PanelProps> = {}, withSeek = true) {
  return render(
    <I18nProvider instance={createI18n('en')}>
      <DiscussionPanel
        page={page}
        isLoading={false}
        onRetry={() => undefined}
        positionMs={9_876}
        {...(withSeek ? { onSeek: () => undefined } : {})}
        onPublish={async () => undefined}
        onUpdate={async () => undefined}
        onRemove={async () => undefined}
        {...overrides}
      />
    </I18nProvider>,
  );
}

test('timestamp is a seek button and server capabilities control edit/remove actions', async () => {
  const seeks: number[] = [];
  renderPanel({ onSeek: (positionMs) => seeks.push(positionMs) });
  await userEvent.click(screen.getByRole('button', { name: '00:12' }));
  expect(seeks).toEqual([12_345]);
  expect(screen.queryByRole('button', { name: 'Edit' })).toBeNull();
  expect(screen.queryByRole('button', { name: 'Remove' })).toBeNull();
});

test('timestamp stays readable text when seeking is unavailable', () => {
  renderPanel({}, false);
  const timestamp = screen.getByText('00:12');
  expect(timestamp.tagName).toBe('TIME');
  expect(timestamp.getAttribute('datetime')).toBe('PT12.345S');
  expect(screen.queryByRole('button', { name: '00:12' })).toBeNull();
});

test('current moment button uses the player position', async () => {
  renderPanel();
  await userEvent.click(screen.getByRole('button', { name: 'Attach current moment' }));
  expect(screen.getByText('Attached at 00:09')).toBeTruthy();
});

test('K16 scopes the loading status to the discussion panel', () => {
  renderPanel({ isLoading: true });
  const panel = screen.getByRole('region', { name: 'Discussion' });
  expect(within(panel).getByRole('status').textContent).toBe('Loading discussion…');
});

test('K16 renders an empty discussion state', () => {
  renderPanel({ page: { threads: [], nextCursor: null } });
  const panel = screen.getByRole('region', { name: 'Discussion' });
  expect(within(panel).getByText('No comments yet.')).toBeTruthy();
});

test('K16 retries a discussion load error from inside the panel', async () => {
  let retries = 0;
  renderPanel({
    error: 'dependency_unavailable',
    onRetry: () => {
      retries += 1;
    },
  });
  const panel = screen.getByRole('region', { name: 'Discussion' });
  expect(within(panel).getByRole('alert').textContent).toContain(
    'The service is unavailable. Try again.',
  );
  await userEvent.click(within(panel).getByRole('button', { name: 'Try again' }));
  expect(retries).toBe(1);
});

test('K16 posts a reply with the root parent and no position', async () => {
  const published: unknown[] = [];
  renderPanel({
    onPublish: async (input) => {
      published.push(input);
    },
  });
  await userEvent.click(screen.getByRole('button', { name: 'Reply' }));
  await userEvent.type(screen.getByRole('textbox', { name: 'Reply' }), 'A reply');
  await userEvent.click(screen.getByRole('button', { name: 'Send reply' }));
  expect(published).toEqual([{ body: 'A reply', parentId: page.threads[0]?.root.id }]);
});

test('a reply card cannot start another reply thread', () => {
  const reply = {
    ...rootComment,
    id: '40000000-0000-4000-8000-000000000002',
    body: 'Nested reply',
  };
  renderPanel({
    page: { threads: [{ root: rootComment, replies: [reply] }], nextCursor: null },
  });
  const replyBody = screen.getByText('Nested reply');
  const replyCard = replyBody.closest('article');
  if (replyCard === null) throw new Error('Reply fixture is not inside a comment card');
  expect(within(replyCard).queryByRole('button', { name: 'Reply' })).toBeNull();
  expect(screen.getByRole('button', { name: 'Reply' })).toBeTruthy();
});

test('K16 renders a removed comment without text or comment actions', () => {
  const removed = {
    ...rootComment,
    body: null,
    removed: true,
    canEdit: true,
    canRemove: true,
  };
  renderPanel({ page: { threads: [{ root: removed, replies: [] }], nextCursor: null } });
  const panel = screen.getByRole('region', { name: 'Discussion' });
  const comment = within(panel).getByText('Comment removed').closest('article');
  expect(comment).toBeTruthy();
  expect(within(comment as HTMLElement).queryByRole('button', { name: 'Reply' })).toBeNull();
  expect(within(comment as HTMLElement).queryByRole('button', { name: 'Edit' })).toBeNull();
  expect(within(comment as HTMLElement).queryByRole('button', { name: 'Remove' })).toBeNull();
  expect(within(comment as HTMLElement).queryByText('Watch this part')).toBeNull();
});

test('K16 uses the former participant label when the author was deleted', () => {
  renderPanel({
    page: {
      threads: [{ root: { ...rootComment, author: { displayName: null } }, replies: [] }],
      nextCursor: null,
    },
  });
  expect(screen.getByText('Former participant')).toBeTruthy();
});

test('K16 marks an edited comment', () => {
  renderPanel({
    page: {
      threads: [
        {
          root: { ...rootComment, editedAt: '2026-09-28T11:00:00Z' },
          replies: [],
        },
      ],
      nextCursor: null,
    },
  });
  expect(screen.getByText('edited')).toBeTruthy();
});

test('K16 shows more only when the server gives a cursor and loads the next page', async () => {
  let loads = 0;
  const firstRender = renderPanel({
    page: { ...page, nextCursor: 'next-page' },
    onLoadMore: () => {
      loads += 1;
    },
  });
  await userEvent.click(screen.getByRole('button', { name: 'Show more' }));
  expect(loads).toBe(1);

  firstRender.unmount();
  renderPanel({ page });
  expect(screen.queryByRole('button', { name: 'Show more' })).toBeNull();
});

test('K16 translates mutation errors by their named code', () => {
  const panel = renderPanel({ error: 'comment_thread_full' });
  expect(
    within(panel.getByRole('region', { name: 'Discussion' })).getByRole('alert').textContent,
  ).toContain('This thread has reached its reply limit.');
});

test('W03 reports a student comment, shows success, and hides report on own or removed comments', async () => {
  let reports = 0;
  renderPanel({
    onReport: async () => {
      reports += 1;
    },
  });
  await userEvent.click(screen.getByRole('button', { name: 'Report' }));
  expect(reports).toBe(1);
  expect(screen.getByRole('button', { name: 'Reported' })).toBeTruthy();

  const own = { ...rootComment, canEdit: true };
  renderPanel({
    page: { threads: [{ root: own, replies: [] }], nextCursor: null },
    onReport: async () => undefined,
  });
  expect(screen.queryByRole('button', { name: 'Report' })).toBeNull();
  const removed = { ...rootComment, body: null, removed: true };
  renderPanel({
    page: { threads: [{ root: removed, replies: [] }], nextCursor: null },
    onReport: async () => undefined,
  });
  expect(screen.queryByRole('button', { name: 'Report' })).toBeNull();
});

test('W03 shows an error when reporting fails', async () => {
  renderPanel({
    onReport: async () => {
      throw new Error('report failed');
    },
  });
  await userEvent.click(screen.getByRole('button', { name: 'Report' }));
  expect(await screen.findByRole('alert')).toBeTruthy();
  expect(screen.getByRole('alert').textContent).toContain('Could not report this comment');
});

for (const [code, message] of [
  ['enrollment_required', 'Enroll in this course to comment or report comments.'],
  [
    'enrollment_blocked',
    'You are blocked in this course — commenting and reporting comments are unavailable.',
  ],
] as const) {
  test(`shows the ${code} comment message`, () => {
    renderPanel({ error: code });
    expect(screen.getByRole('alert').textContent).toContain(message);
  });
  test(`shows the ${code} report message without claiming success`, async () => {
    renderPanel({
      onReport: async () => {
        throw new Error(code);
      },
    });
    await userEvent.click(screen.getByRole('button', { name: 'Report' }));
    expect((await screen.findByRole('alert')).textContent).toBe(message);
    expect(screen.queryByRole('button', { name: 'Reported' })).toBeNull();
  });
}
