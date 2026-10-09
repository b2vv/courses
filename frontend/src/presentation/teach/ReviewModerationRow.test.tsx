import { describe, expect, it, rs } from '@rstest/core';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import type { ModerationQueueItem } from '../../domain/teach/teach';
import { ReviewModerationRow } from './ReviewModerationRow';

const item = {
  kind: 'review' as const,
  comment: null,
  openReports: 2,
  authorBlocked: false,
  review: {
    id: '75000000-0000-4000-8000-000000000001',
    authorDisplayName: 'Студентка',
    stars: 2,
    body: 'Лекції зачитані з папірця, практики немає.',
    completedLessons: 1,
    totalLessons: 4,
    createdAt: '2026-10-01T09:00:00Z',
    editedAt: null,
  },
} satisfies ModerationQueueItem;

function renderRow(
  admin: boolean,
  overrides: Partial<Pick<ModerationQueueItem, 'authorBlocked'>> = {},
  handlers: {
    onDismiss?: () => void;
    onRemove?: () => void;
    onToggleAuthorBlock?: () => void;
  } = {},
) {
  const onDismiss = handlers.onDismiss ?? rs.fn();
  const onRemove = handlers.onRemove ?? rs.fn();
  const onToggleAuthorBlock = handlers.onToggleAuthorBlock ?? rs.fn();
  render(
    <ul>
      <ReviewModerationRow
        item={{ ...item, ...overrides }}
        busy={false}
        admin={admin}
        authorPending={false}
        onDismiss={onDismiss}
        onRemove={onRemove}
        onToggleAuthorBlock={onToggleAuthorBlock}
      />
    </ul>,
  );
  return { onDismiss, onRemove, onToggleAuthorBlock };
}

describe('teach queue review row', () => {
  it('shows the author, stars for assistive tech, the text and the report count', () => {
    renderRow(false);
    expect(screen.getByText(/Студентка/)).toBeTruthy();
    expect(screen.getByText('★★').getAttribute('aria-hidden')).toBe('true');
    expect(screen.getByText('2 з 5')).toBeTruthy();
    expect(screen.getByText('Лекції зачитані з папірця, практики немає.')).toBeTruthy();
  });

  it('offers removal only to an admin (G9); the owner can only dismiss (Rt21)', async () => {
    const owner = renderRow(false);
    expect(screen.queryByRole('button', { name: 'Зняти' })).toBeNull();
    await userEvent.click(screen.getByRole('button', { name: /Відхилити/ }));
    expect(owner.onDismiss).toHaveBeenCalledTimes(1);
  });

  it('lets an admin remove the review', async () => {
    const admin = renderRow(true);
    await userEvent.click(screen.getByRole('button', { name: 'Зняти' }));
    expect(admin.onRemove).toHaveBeenCalledTimes(1);
  });

  it('blocks the review author from the queue (Lo13)', async () => {
    const owner = renderRow(false, { authorBlocked: false });
    expect(screen.queryByRole('button', { name: 'Розблокувати автора' })).toBeNull();
    await userEvent.click(screen.getByRole('button', { name: 'Заблокувати автора' }));
    expect(owner.onToggleAuthorBlock).toHaveBeenCalledTimes(1);
  });

  it('offers unblock when the author is already blocked', () => {
    renderRow(false, { authorBlocked: true });
    expect(screen.getByRole('button', { name: 'Розблокувати автора' })).toBeTruthy();
    expect(screen.queryByRole('button', { name: 'Заблокувати автора' })).toBeNull();
  });
});
